import { useState } from 'react'
import { Lock, User, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useLogin } from '@/hooks/use-auth'
import { toast } from 'sonner'

function formatErrorMessage(msg: string): string {
  const map: Record<string, string> = {
    'invalid username or password': '用户名或密码错误',
    'invalid request payload': '请求格式错误，请检查输入',
    'unauthorized': '未授权，请重新登录',
  }
  return map[msg] || '登录失败，请重试'
}

/**
 * 登录页 — Vercel 风格：纯白 + border + shadow-sm
 */
export function LoginPage() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [attempted, setAttempted] = useState(false)
  const loginMutation = useLogin()

  const usernameEmpty = !username.trim()
  const passwordEmpty = !password

  function handleSubmit() {
    setAttempted(true)
    if (usernameEmpty || passwordEmpty) return
    loginMutation.mutate({ username: username.trim(), password }, {
      onError: (err) => { toast.error('登录失败', { description: formatErrorMessage(err.message) }) },
    })
  }

  const inputCls = 'border-zinc-200 dark:border-zinc-700 bg-transparent rounded-md shadow-none focus-visible:ring-1 focus-visible:ring-zinc-400'

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#FAFAFA] dark:bg-zinc-950 p-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-10">
          <div className="w-10 h-10 rounded-full bg-zinc-900 dark:bg-zinc-200 text-white dark:text-zinc-900 inline-flex items-center justify-center mb-4">
            <span className="text-lg">🪁</span>
          </div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50 tracking-tight">Kite</h1>
          <p className="text-sm text-zinc-500 mt-1">管理后台登录</p>
        </div>

        <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm p-8">
          <div className="space-y-5">
            <div>
              <label className="text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-2 block">用户名</label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
                <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="admin" className={`pl-9 h-10 ${inputCls} ${attempted && usernameEmpty ? 'border-red-400' : ''}`} onKeyDown={(e) => e.key === 'Enter' && handleSubmit()} autoFocus />
              </div>
              {attempted && usernameEmpty && <p className="text-xs text-red-500 mt-1.5">请输入用户名</p>}
            </div>
            <div>
              <label className="text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-2 block">密码</label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
                <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="••••••••" className={`pl-9 h-10 ${inputCls} ${attempted && passwordEmpty ? 'border-red-400' : ''}`} onKeyDown={(e) => e.key === 'Enter' && handleSubmit()} />
              </div>
              {attempted && passwordEmpty && <p className="text-xs text-red-500 mt-1.5">请输入密码</p>}
            </div>
            <Button className="w-full h-10 bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-md shadow-sm hover:bg-zinc-800 dark:hover:bg-zinc-200 text-sm font-medium transition-colors" onClick={handleSubmit} disabled={loginMutation.isPending}>
              {loginMutation.isPending ? '登录中…' : <><span>登录</span><ArrowRight className="w-3.5 h-3.5 ml-1.5" /></>}
            </Button>
          </div>
        </div>
        <p className="text-center text-[11px] text-zinc-400 mt-8">Kite Blog — 轻量级博客引擎</p>
      </div>
    </div>
  )
}
