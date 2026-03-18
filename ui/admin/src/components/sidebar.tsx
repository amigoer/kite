import { useLocation, useNavigate } from 'react-router'
import {
  LayoutDashboard, FileText, Files, FolderOpen, Tags,
  MessageSquare, Link2, Settings, Search, LogOut,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Tooltip, TooltipContent, TooltipTrigger,
} from '@/components/ui/tooltip'
import { useSidebarStore } from '@/stores/use-sidebar-store'
import { useCurrentUser, useLogout } from '@/hooks/use-auth'
import { cn } from '@/lib/utils'

const navItems = [
  { path: '/', label: '仪表盘', icon: LayoutDashboard },
  { path: '/posts', label: '文章', icon: FileText },
  { path: '/pages', label: '页面', icon: Files },
  { path: '/categories', label: '分类', icon: FolderOpen },
  { path: '/tags', label: '标签', icon: Tags },
  { path: '/comments', label: '评论', icon: MessageSquare },
  { path: '/links', label: '友链', icon: Link2 },
  { path: '/settings', label: '设置', icon: Settings },
]

/**
 * 侧边栏 — Vercel 风格：纯白 + 极浅灰 active
 */
export function Sidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  const { isCollapsed } = useSidebarStore()
  const { data: currentUser } = useCurrentUser()
  const logoutMutation = useLogout()

  const displayName = currentUser?.user?.displayName || currentUser?.user?.username || 'Admin'
  const username = currentUser?.user?.username || ''

  return (
    <aside
      className={cn(
        'fixed left-0 top-0 h-screen flex flex-col bg-white dark:bg-zinc-950',
        'border-r border-zinc-200 dark:border-zinc-800 transition-all duration-200 z-30',
        isCollapsed ? 'w-16' : 'w-56'
      )}
    >
      {/* Logo */}
      <div className={cn('h-14 flex items-center shrink-0 border-b border-zinc-100 dark:border-zinc-800', isCollapsed ? 'px-0 justify-center' : 'px-5')}>
        <span className="text-sm font-bold tracking-tight text-zinc-900 dark:text-zinc-50 select-none">
          {isCollapsed ? 'K' : 'Kite'}
        </span>
      </div>

      {/* 搜索 */}
      {!isCollapsed && (
        <div className="px-3 py-3 shrink-0">
          <button
            className="flex items-center gap-2 w-full px-3 py-1.5 rounded-lg border border-zinc-200 dark:border-zinc-800 text-sm text-zinc-400 hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors cursor-pointer"
            onClick={() => window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true }))}
          >
            <Search className="w-3.5 h-3.5" />
            <span className="flex-1 text-left text-xs">搜索…</span>
            <kbd className="text-[10px] px-1 py-0.5 rounded border border-zinc-200 dark:border-zinc-700 text-zinc-400 font-mono leading-none">⌘K</kbd>
          </button>
        </div>
      )}

      {/* 导航 */}
      <nav className="flex-1 overflow-auto px-3 py-1 space-y-0.5">
        {navItems.map(({ path, label, icon: Icon }) => {
          const isActive = path === '/'
            ? location.pathname === '/'
            : location.pathname.startsWith(path)

          const btn = (
            <Button
              key={path}
              variant="ghost"
              className={cn(
                'w-full justify-start gap-2.5 h-auto px-3 py-1.5 text-[13px] rounded-md transition-colors',
                isCollapsed && 'justify-center px-0',
                isActive
                  ? 'bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-50 font-medium'
                  : 'text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-200 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 font-normal'
              )}
              onClick={() => navigate(path)}
            >
              <Icon className="w-4 h-4 shrink-0" />
              {!isCollapsed && label}
            </Button>
          )

          if (isCollapsed) {
            return (
              <Tooltip key={path} delayDuration={0}>
                <TooltipTrigger asChild>{btn}</TooltipTrigger>
                <TooltipContent side="right" className="text-xs">{label}</TooltipContent>
              </Tooltip>
            )
          }
          return btn
        })}
      </nav>

      {/* 底部用户 */}
      <div className="shrink-0 border-t border-zinc-100 dark:border-zinc-800 px-3 py-3">
        <div className={cn('flex items-center', isCollapsed ? 'justify-center' : 'gap-3')}>
          <div className="w-7 h-7 rounded-full bg-zinc-900 dark:bg-zinc-200 text-white dark:text-zinc-900 flex items-center justify-center text-[11px] font-semibold shrink-0 select-none">
            {displayName.charAt(0).toUpperCase()}
          </div>
          {!isCollapsed && (
            <>
              <div className="flex-1 min-w-0">
                <div className="text-[13px] font-medium text-zinc-700 dark:text-zinc-300 truncate leading-tight">{displayName}</div>
                <div className="text-[11px] text-zinc-400 dark:text-zinc-500 truncate leading-tight">{username}</div>
              </div>
              <Tooltip delayDuration={0}>
                <TooltipTrigger asChild>
                  <button className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-200 transition-colors cursor-pointer" onClick={() => logoutMutation.mutate()}>
                    <LogOut className="w-3.5 h-3.5" />
                  </button>
                </TooltipTrigger>
                <TooltipContent side="right" className="text-xs">退出登录</TooltipContent>
              </Tooltip>
            </>
          )}
        </div>
      </div>
    </aside>
  )
}
