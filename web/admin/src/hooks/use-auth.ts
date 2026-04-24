import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from 'react'
import { authApi } from '@/lib/api'
import { toast } from 'sonner'
import { translate } from '@/i18n'

export interface User {
  user_id: string
  username: string
  nickname?: string
  email?: string
  avatar_url?: string
  has_local_password?: boolean
  role: string
  password_must_change?: boolean
  storage_limit?: number
  storage_used?: number
  totp_enabled?: boolean
  created_at?: string
}

// LoginOutcome distinguishes the two login results:
//   - Password-only account → { ok: true }, session established.
//   - 2FA account → { pending2fa: true, ... }, caller must prompt
//     for a TOTP code and exchange the challenge via verifyTotp.
// Modelling this as a tagged union (instead of throwing on 2FA) keeps
// the 2FA branch out of the error channel, where it would conflict
// with real failures the UI needs to show.
export type LoginOutcome =
  | { ok: true }
  | {
      ok: false
      pending2fa: true
      challengeToken: string
      expiresAt: string
    }

export interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (username: string, password: string) => Promise<LoginOutcome>
  // verifyTotp completes a login that returned pending_2fa. Success
  // establishes the session exactly the way a password-only login
  // would, so the caller's post-login flow is identical.
  verifyTotp: (challengeToken: string, code: string) => Promise<void>
  register: (username: string, email: string, password: string) => Promise<void>
  logout: () => void
  // refreshProfile re-fetches the current user from the server and updates
  // the in-memory store. Auth tokens live in HttpOnly cookies and are
  // rotated server-side — the frontend never touches them.
  refreshProfile: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

const PROFILE_CACHE_KEY = 'user_profile'

function readCachedUser(): User | null {
  try {
    const raw = localStorage.getItem(PROFILE_CACHE_KEY)
    return raw ? (JSON.parse(raw) as User) : null
  } catch {
    return null
  }
}

function writeCachedUser(user: User | null) {
  if (user) {
    localStorage.setItem(PROFILE_CACHE_KEY, JSON.stringify(user))
  } else {
    localStorage.removeItem(PROFILE_CACHE_KEY)
  }
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

export function useAuthProvider(): AuthContextValue {
  // Auth tokens are HttpOnly cookies — JS can't see them. We use a cached
  // user profile in localStorage as our "probably logged in" signal, so we
  // can hydrate optimistically and avoid a /profile probe on every anon
  // page load. The cache is validated against the server on mount.
  const [user, setUser] = useState<User | null>(() => {
    if (typeof window === 'undefined') return null
    return readCachedUser()
  })
  const [loading, setLoading] = useState<boolean>(() => {
    if (typeof window === 'undefined') return false
    // If we have a cached profile, we're optimistic: surface the cached
    // user immediately and validate in the background (loading=false). If
    // the cache is empty, we assume anonymous and show the logged-out
    // state without a probe — callers hit /login themselves.
    return false
  })

  const fetchProfile = useCallback(async () => {
    try {
      const { data } = await authApi.profile()
      setUser(data.data)
      writeCachedUser(data.data)
    } catch {
      // Either no cookie or the refresh interceptor failed to renew — in
      // either case we're unauthenticated. The interceptor already handles
      // the redirect for pages that require auth.
      writeCachedUser(null)
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    // Revalidate on mount only if we have a cached user. A cold anonymous
    // load shouldn't hit /profile just to get a 401.
    if (readCachedUser()) {
      fetchProfile()
    }
  }, [fetchProfile])

  const login = useCallback(
    async (username: string, password: string): Promise<LoginOutcome> => {
      // The server sets HttpOnly access/refresh cookies on success.
      // When 2FA is enabled it returns { pending_2fa: true, ... } and
      // NO cookies — so calling /profile here would 401. We surface
      // the pending state to the caller instead and let them prompt
      // for the code.
      const res = await authApi.login(username, password)
      const payload = res.data.data as {
        pending_2fa?: boolean
        challenge_token?: string
        expires_at?: string
      } | null
      if (payload?.pending_2fa) {
        return {
          ok: false,
          pending2fa: true,
          challengeToken: payload.challenge_token ?? '',
          expiresAt: payload.expires_at ?? '',
        }
      }
      const profile = await authApi.profile()
      setUser(profile.data.data)
      writeCachedUser(profile.data.data)
      return { ok: true }
    },
    []
  )

  const verifyTotp = useCallback(
    async (challengeToken: string, code: string) => {
      await authApi.verifyTotp(challengeToken, code)
      const profile = await authApi.profile()
      setUser(profile.data.data)
      writeCachedUser(profile.data.data)
    },
    []
  )

  const register = useCallback(
    async (username: string, email: string, password: string) => {
      await authApi.register(username, email, password)
    },
    []
  )

  const logout = useCallback(() => {
    // Server clears the cookies; we drop the cached profile.
    authApi.logout().catch(() => {})
    writeCachedUser(null)
    setUser(null)
    toast.success(translate('auth.loggedOut'))
  }, [])

  const refreshProfile = useCallback(async () => {
    const profile = await authApi.profile()
    setUser(profile.data.data)
    writeCachedUser(profile.data.data)
  }, [])

  return {
    user,
    loading,
    login,
    verifyTotp,
    register,
    logout,
    refreshProfile,
  }
}
