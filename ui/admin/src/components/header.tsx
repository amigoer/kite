import { useLocation, useNavigate } from 'react-router'
import {
  PanelLeft, ExternalLink, Moon, Sun, Bell, Settings, LogOut,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList,
  BreadcrumbPage, BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { useSidebarStore } from '@/stores/use-sidebar-store'
import { useThemeStore } from '@/stores/use-theme-store'
import { useCurrentUser, useLogout } from '@/hooks/use-auth'
import { useCommentStats } from '@/hooks/use-comments'

const breadcrumbMap: Record<string, string> = {
  '/': '仪表盘', '/posts': '文章管理', '/posts/new': '新建文章',
  '/categories': '分类管理', '/tags': '标签管理', '/comments': '评论管理',
  '/links': '友链管理', '/settings': '系统设置', '/pages': '页面管理',
}

/**
 * Header — Vercel 风格：极简纯白
 */
export function Header() {
  const { toggle } = useSidebarStore()
  const location = useLocation()
  const navigate = useNavigate()
  const { isDark, toggle: toggleTheme } = useThemeStore()
  const { data: currentUser } = useCurrentUser()
  const logoutMutation = useLogout()
  const { data: commentStats } = useCommentStats()
  const pendingCount = commentStats?.pending ?? 0

  const displayName = currentUser?.user?.displayName || currentUser?.user?.username || 'Admin'
  const pathSegments = location.pathname.split('/').filter(Boolean)
  const currentLabel = breadcrumbMap[location.pathname]
  const isEditPost = pathSegments[0] === 'posts' && pathSegments.length >= 3

  return (
    <header className="h-12 shrink-0 flex items-center justify-between px-6 bg-white dark:bg-zinc-950 border-b border-zinc-200 dark:border-zinc-800">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon" className="w-7 h-7 rounded-md text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-200" onClick={toggle}>
          <PanelLeft className="w-4 h-4" />
        </Button>
        <div className="w-px h-4 bg-zinc-200 dark:bg-zinc-700 mx-0.5" />
        <Breadcrumb>
          <BreadcrumbList className="text-[13px]">
            <BreadcrumbItem>
              <BreadcrumbLink className="text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-200 cursor-pointer transition-colors" onClick={() => navigate('/')}>
                Kite
              </BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            {isEditPost ? (
              <>
                <BreadcrumbItem>
                  <BreadcrumbLink className="text-zinc-500 hover:text-zinc-900 cursor-pointer transition-colors" onClick={() => navigate('/posts')}>文章管理</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem><BreadcrumbPage className="text-zinc-900 dark:text-zinc-50 font-semibold">编辑文章</BreadcrumbPage></BreadcrumbItem>
              </>
            ) : (
              <BreadcrumbItem><BreadcrumbPage className="text-zinc-900 dark:text-zinc-50 font-semibold">{currentLabel || '页面'}</BreadcrumbPage></BreadcrumbItem>
            )}
          </BreadcrumbList>
        </Breadcrumb>
      </div>
      <div className="flex items-center gap-0.5">
        <Button variant="ghost" size="icon" className="w-7 h-7 rounded-md text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-200" onClick={() => window.open('/', '_blank')}>
          <ExternalLink className="w-3.5 h-3.5" />
        </Button>
        <Button variant="ghost" size="icon" className="w-7 h-7 rounded-md text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-200" onClick={toggleTheme}>
          {isDark ? <Sun className="w-3.5 h-3.5" /> : <Moon className="w-3.5 h-3.5" />}
        </Button>
        <Button variant="ghost" size="icon" className="w-7 h-7 rounded-md text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-200 relative" onClick={() => navigate('/comments')}>
          <Bell className="w-3.5 h-3.5" />
          {pendingCount > 0 && (
            <span className="absolute -top-0.5 -right-0.5 h-[14px] min-w-[14px] px-[3px] text-[9px] rounded-full bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 flex items-center justify-center font-medium leading-none">
              {pendingCount > 99 ? '99+' : pendingCount}
            </span>
          )}
        </Button>
        <div className="w-px h-4 bg-zinc-200 dark:bg-zinc-700 mx-2" />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="flex items-center gap-2 px-1 py-1 rounded-md hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors cursor-pointer outline-none">
              <div className="w-6 h-6 rounded-full bg-zinc-900 dark:bg-zinc-200 text-white dark:text-zinc-900 flex items-center justify-center text-[10px] font-semibold">
                {displayName.charAt(0).toUpperCase()}
              </div>
              <span className="text-[13px] text-zinc-700 dark:text-zinc-300">{displayName}</span>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-44">
            <DropdownMenuItem className="cursor-pointer text-[13px]" onClick={() => navigate('/settings')}>
              <Settings className="w-3.5 h-3.5 mr-2 text-zinc-500" /> 系统设置
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="cursor-pointer text-[13px] text-red-500 focus:text-red-500" onClick={() => logoutMutation.mutate()}>
              <LogOut className="w-3.5 h-3.5 mr-2" /> 退出登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )
}
