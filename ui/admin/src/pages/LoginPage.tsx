import { useState } from 'react'
import { Loader2, LogIn } from 'lucide-react'
import { KiteIcon } from '@/components/KiteIcon'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { PasswordInput } from '@/components/PasswordInput'
import { useLogin } from '@/hooks/use-auth'
import { toast } from 'sonner'
import { ThemeSwitch } from '@/components/theme-switch'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

function formatErrorMessage(msg: string): string {
  const map: Record<string, string> = {
    'invalid username or password': '用户名或密码错误',
    'invalid request payload': '请求格式错误，请检查输入',
    'unauthorized': '未授权，请重新登录',
  }
  return map[msg] || '登录失败，请重试'
}

/**
 * 登录页 — 完全复刻 shadcn-admin 的 sign-in-2 模板
 */
export function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [attempted, setAttempted] = useState(false)
  const loginMutation = useLogin()

  const usernameEmpty = !username.trim()
  const passwordEmpty = !password

  function handleSubmit(e?: React.FormEvent) {
    e?.preventDefault()
    setAttempted(true)
    if (usernameEmpty || passwordEmpty) return
    loginMutation.mutate({ username: username.trim(), password }, {
      onError: (err) => { toast.error('登录失败', { description: formatErrorMessage(err.message) }) },
    })
  }

  return (
    <div className='relative flex h-svh w-full items-center justify-center px-4'>
      <ThemeSwitch className='absolute right-4 top-4 md:right-8 md:top-8 z-50 bg-background/50 backdrop-blur-sm' />
      
      <div className="w-full max-w-sm flex flex-col items-center">
        <Card className="w-full shadow-md border-zinc-200 dark:border-zinc-800">
          <CardHeader className="space-y-3 text-center pt-8 pb-6">
            <div className="flex justify-center">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-zinc-950 dark:bg-zinc-100 shadow-sm">
                <KiteIcon className="h-7 w-7 text-white dark:text-zinc-950" />
              </div>
            </div>
            <CardTitle className="text-2xl font-bold tracking-tight mt-2">Kite Admin</CardTitle>
            <CardDescription className="text-sm">登录以管理您的博客内容</CardDescription>
          </CardHeader>
          <CardContent className="pb-8">
            <form onSubmit={handleSubmit} className="space-y-5">
              <div className="space-y-2 text-left">
                <Label htmlFor="username" className="text-sm font-medium">用户名</Label>
                <Input
                  id="username"
                  placeholder="admin"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="shadow-sm"
                  autoFocus
                />
                {attempted && usernameEmpty && (
                  <p className="text-xs text-destructive">请输入用户名</p>
                )}
              </div>
              <div className="space-y-2 text-left">
                <div className="flex items-center justify-between">
                  <Label htmlFor="password" className="text-sm font-medium">密码</Label>
                  <a
                    href="#"
                    className="text-xs font-medium text-muted-foreground hover:text-primary transition-colors duration-200"
                    onClick={(e) => { e.preventDefault(); toast.info('请联系系统管理员重设密码') }}
                  >
                    忘记密码？
                  </a>
                </div>
                <PasswordInput
                  id="password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="shadow-sm"
                />
                {attempted && passwordEmpty && (
                  <p className="text-xs text-destructive">请输入密码</p>
                )}
              </div>
              <Button type="submit" className="w-full h-11 text-sm font-medium shadow-sm active:scale-[0.98] transition-all" disabled={loginMutation.isPending}>
                {loginMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <LogIn className="mr-2 h-4 w-4" />}
                登录
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
      <p className='absolute bottom-4 left-0 right-0 text-center text-xs text-muted-foreground'>
        © {new Date().getFullYear()}{' '}
        <a
          href='https://github.com/amigoer/kite'
          target='_blank'
          rel='noopener noreferrer'
          className='hover:text-primary underline-offset-4 hover:underline'
        >
          Kite
        </a>{' '}
        · 轻量级博客引擎
      </p>
    </div>
  )
}
