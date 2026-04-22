import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loader2, ShieldAlert } from 'lucide-react'
import { useAuth } from '@/hooks/use-auth'
import { authApi } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { KiteLogo } from '@/components/kite-logo'
import { toast } from 'sonner'

export default function FirstLoginPage() {
  const { refreshProfile } = useAuth()
  const navigate = useNavigate()
  const [form, setForm] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  })
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (form.username === 'admin') {
      toast.error('新用户名不能仍然是 admin')
      return
    }
    if (form.password === 'admin') {
      toast.error('新密码不能仍然是 admin')
      return
    }
    if (form.password !== form.confirmPassword) {
      toast.error('两次输入的密码不一致，请重新核对')
      return
    }

    setLoading(true)
    try {
      await authApi.firstLoginReset({
        new_username: form.username,
        new_email: form.email,
        new_password: form.password,
      })
      // Server set fresh HttpOnly cookies on the reset response; we
      // only need to pull in the new profile.
      await refreshProfile()
      toast.success('账号已重置，欢迎使用 Kite！')
      navigate('/user/dashboard', { replace: true })
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data
          ?.message ?? '重置失败，请稍后重试'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  const update = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((prev) => ({ ...prev, [field]: e.target.value }))

  return (
    <>
      <div className="mx-auto flex w-full flex-col justify-center space-y-2 py-8 sm:w-[480px] sm:p-8">
        <div className="mb-4 flex items-center justify-center">
          <KiteLogo className="me-2 size-6" />
          <h1 className="text-xl font-medium">Kite</h1>
        </div>
      </div>

      <div className="mx-auto flex w-full max-w-sm flex-col justify-center">
        <div className="mb-4 flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 p-3 text-amber-700 dark:text-amber-400">
          <ShieldAlert className="mt-0.5 size-4 shrink-0" />
          <div className="text-xs leading-relaxed">
            检测到默认账号 <span className="font-mono">admin / admin</span>
            ，出于安全考虑，请先设置一个新的用户名与密码再继续使用。
          </div>
        </div>

        <div className="flex flex-col space-y-1.5 text-start">
          <h2 className="text-2xl font-semibold tracking-tight">
            完成首次配置
          </h2>
          <p className="text-sm text-muted-foreground">设置管理员账号信息</p>
        </div>

        <form onSubmit={handleSubmit} className="grid gap-3 pt-2">
          <div className="grid gap-2">
            <Label htmlFor="username">新用户名</Label>
            <Input
              id="username"
              autoCapitalize="none"
              autoCorrect="off"
              value={form.username}
              onChange={update('username')}
              placeholder="请输入新用户名"
              required
              minLength={3}
              maxLength={32}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="email">邮箱</Label>
            <Input
              id="email"
              type="email"
              autoCapitalize="none"
              autoComplete="email"
              autoCorrect="off"
              value={form.email}
              onChange={update('email')}
              placeholder="name@example.com"
              required
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="password">新密码</Label>
            <Input
              id="password"
              type="password"
              autoComplete="new-password"
              value={form.password}
              onChange={update('password')}
              placeholder="至少 6 位"
              required
              minLength={6}
              maxLength={64}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="confirmPassword">确认密码</Label>
            <Input
              id="confirmPassword"
              type="password"
              autoComplete="new-password"
              value={form.confirmPassword}
              onChange={update('confirmPassword')}
              placeholder="再次输入新密码"
              required
            />
          </div>

          <Button type="submit" className="mt-2" disabled={loading}>
            {loading && <Loader2 className="size-4 animate-spin" />}
            {loading ? '保存中...' : '保存并继续'}
          </Button>
        </form>

        <p className="mt-6 text-center text-xs text-muted-foreground">
          Kite v1.0 · 首次启动向导
        </p>
      </div>
    </>
  )
}
