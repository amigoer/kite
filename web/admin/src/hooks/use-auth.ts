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
  created_at?: string
}

export interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (username: string, password: string) => Promise<void>
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

  const login = useCallback(async (username: string, password: string) => {
    // The server sets HttpOnly access/refresh cookies in the response.
    await authApi.login(username, password)
    const profile = await authApi.profile()
    setUser(profile.data.data)
    writeCachedUser(profile.data.data)
  }, [])

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

  return { user, loading, login, register, logout, refreshProfile }
}
