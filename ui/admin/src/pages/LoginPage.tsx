import { useState } from 'react'
import { Loader2, Command, Github, LogIn } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { Label } from '@/components/ui/label'
import { PasswordInput } from '@/components/PasswordInput'
import { useLogin } from '@/hooks/use-auth'
import { toast } from 'sonner'
import dashboardDark from '@/assets/dashboard-dark.png'
import dashboardLight from '@/assets/dashboard-light.png'
import { ThemeSwitch } from '@/components/theme-switch'

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

  function handleMockSocialLogin(provider: string) {
    toast.info(`暂不支持 ${provider} 登录`)
  }

  return (
    <div className='relative container grid h-svh flex-col items-center justify-center lg:max-w-none lg:grid-cols-2 lg:px-0'>
      <ThemeSwitch className='absolute right-4 top-4 md:right-8 md:top-8 z-50 bg-background/50 backdrop-blur-sm' />
      <div className='lg:p-8'>
        <div className='mx-auto flex w-full flex-col justify-center space-y-2 py-8 sm:w-[480px] sm:p-8'>
          <div className='mb-4 flex items-center justify-center gap-2'>
            <Command className='h-6 w-6' />
            <h1 className='text-xl font-medium'>Kite Admin</h1>
          </div>
        </div>
        <div className='mx-auto flex w-full max-w-sm flex-col justify-center space-y-2'>
          <div className='grid gap-6 mt-4'>
            <form onSubmit={handleSubmit}>
              <div className='grid gap-4'>
                <div className='grid gap-2'>
                  <Label htmlFor='username'>用户名</Label>
                  <Input
                    id='username'
                    placeholder='admin'
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    autoFocus
                  />
                  {attempted && usernameEmpty && (
                    <p className='text-xs text-destructive'>请输入用户名</p>
                  )}
                </div>
                <div className='grid gap-2'>
                  <div className='flex items-center justify-between'>
                    <Label htmlFor='password'>密码</Label>
                    <a
                      href='#'
                      className='text-sm text-muted-foreground hover:text-primary'
                      onClick={(e) => { e.preventDefault(); toast.info('请联系系统管理员重置密码') }}
                    >
                      忘记密码？
                    </a>
                  </div>
                  <PasswordInput
                    id='password'
                    placeholder='********'
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                  {attempted && passwordEmpty && (
                    <p className='text-xs text-destructive'>请输入密码</p>
                  )}
                </div>
                <Button type='submit' className='w-full h-11' disabled={loginMutation.isPending}>
                  {loginMutation.isPending ? <Loader2 className='mr-2 h-4 w-4 animate-spin' /> : <LogIn className='mr-2 h-4 w-4' />}
                  登录
                </Button>
              </div>
            </form>

            <div className='relative'>
              <div className='absolute inset-0 flex items-center'>
                <span className='w-full border-t' />
              </div>
              <div className='relative flex justify-center text-xs uppercase'>
                <span className='bg-background px-2 text-muted-foreground'>
                  其他登录方式
                </span>
              </div>
            </div>

            <div className='grid grid-cols-2 gap-4'>
              <Button
                variant='outline'
                type='button'
                className='h-11'
                onClick={() => handleMockSocialLogin('GitHub')}
              >
                <Github className='mr-2 h-4 w-4' />
                GitHub
              </Button>
              <Button
                variant='outline'
                type='button'
                className='h-11'
                onClick={() => handleMockSocialLogin('Google')}
              >
                <svg
                  className='mr-2 h-4 w-4'
                  aria-hidden='true'
                  focusable='false'
                  data-prefix='fab'
                  data-icon='google'
                  role='img'
                  xmlns='http://www.w3.org/2000/svg'
                  viewBox='0 0 488 512'
                >
                  <path
                    fill='currentColor'
                    d='M488 261.8C488 403.3 391.1 504 248 504 110.8 504 0 393.2 0 256S110.8 8 248 8c66.8 0 123 24.5 166.3 64.9l-67.5 64.9C258.5 52.6 94.3 116.6 94.3 256c0 86.5 69.1 156.6 153.7 156.6 98.2 0 135-70.4 140.8-106.9H248v-85.3h236.1c2.3 12.7 3.9 24.9 3.9 41.4z'
                  ></path>
                </svg>
                Google
              </Button>
            </div>
          </div>

          <p className='px-8 text-center text-sm text-muted-foreground mt-4'>
            登录即表示您同意我们的{' '}
            <a
              href='#'
              className='underline underline-offset-4 hover:text-primary'
            >
              服务条款
            </a>{' '}
            和{' '}
            <a
              href='#'
              className='underline underline-offset-4 hover:text-primary'
            >
              隐私政策
            </a>
            .
          </p>
        </div>
      </div>

      <div
        className={cn(
          'relative h-full overflow-hidden bg-muted max-lg:hidden',
          '[&>img]:absolute [&>img]:top-[15%] [&>img]:left-20 [&>img]:h-full [&>img]:w-full [&>img]:object-cover [&>img]:object-top-left [&>img]:select-none'
        )}
      >
        <img
          src={dashboardLight}
          className='dark:hidden'
          width={1024}
          height={1151}
          alt='Kite-Admin'
        />
        <img
          src={dashboardDark}
          className='hidden dark:block'
          width={1024}
          height={1138}
          alt='Kite-Admin'
        />
      </div>
    </div>
  )
}
