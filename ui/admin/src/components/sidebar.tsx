import { Link, useLocation } from 'react-router'
import {
  LayoutDashboard,
  FileText,
  FolderTree,
  Tag,
  MessageSquare,
  Link2,
  Settings,
  Wind,
  LogOut,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useSidebarStore } from '@/stores/use-sidebar-store'

/** 导航项定义 */
interface NavItem {
  label: string
  icon: LucideIcon
  path: string
}

/** 导航分组 */
interface NavGroup {
  title: string
  items: NavItem[]
}

/** 侧边栏导航分组配置 */
const navGroups: NavGroup[] = [
  {
    title: '',
    items: [
      { label: '仪表盘', icon: LayoutDashboard, path: '/' },
    ],
  },
  {
    title: '内容',
    items: [
      { label: '文章', icon: FileText, path: '/posts' },
      { label: '分类', icon: FolderTree, path: '/categories' },
      { label: '标签', icon: Tag, path: '/tags' },
      { label: '评论', icon: MessageSquare, path: '/comments' },
    ],
  },
  {
    title: '管理',
    items: [
      { label: '友链', icon: Link2, path: '/links' },
      { label: '设置', icon: Settings, path: '/settings' },
    ],
  },
]

/**
 * 侧边栏组件
 * 使用 shadcn Button / Tooltip / Separator 组件
 */
export function Sidebar() {
  const location = useLocation()
  const { isCollapsed } = useSidebarStore()

  return (
    <aside
      className={cn(
        'h-screen border-r bg-sidebar flex flex-col transition-[width] duration-200 ease-in-out',
        isCollapsed ? 'w-16' : 'w-56'
      )}
    >
      {/* Logo */}
      <div className="flex h-14 items-center px-4 gap-2.5">
        <Wind className="h-5 w-5 text-foreground shrink-0" strokeWidth={2} />
        {!isCollapsed && (
          <span className="text-base font-bold tracking-tight text-foreground">
            Kite
          </span>
        )}
      </div>

      <Separator />

      {/* 导航列表 */}
      <nav className="flex-1 overflow-y-auto px-2 py-3">
        {navGroups.map((group, gi) => (
          <div key={gi} className={cn(gi > 0 && 'mt-4')}>
            {group.title && !isCollapsed && (
              <p className="mb-1.5 px-3 text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                {group.title}
              </p>
            )}
            {group.title && isCollapsed && gi > 0 && (
              <Separator className="mx-2 mb-2" />
            )}
            <ul className="space-y-1">
              {group.items.map((item) => {
                const isActive = location.pathname === item.path
                const Icon = item.icon

                const linkContent = (
                  <Button
                    variant={isActive ? 'secondary' : 'ghost'}
                    className={cn(
                      'w-full justify-start gap-2.5 text-[13px]',
                      isCollapsed && 'justify-center px-0',
                      isActive && 'font-medium'
                    )}
                    asChild
                  >
                    <Link to={item.path}>
                      <Icon className="h-4 w-4 shrink-0" strokeWidth={1.5} />
                      {!isCollapsed && <span className="truncate">{item.label}</span>}
                    </Link>
                  </Button>
                )

                // 折叠模式下使用 Tooltip 显示标签
                if (isCollapsed) {
                  return (
                    <li key={item.path}>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          {linkContent}
                        </TooltipTrigger>
                        <TooltipContent side="right">
                          {item.label}
                        </TooltipContent>
                      </Tooltip>
                    </li>
                  )
                }

                return <li key={item.path}>{linkContent}</li>
              })}
            </ul>
          </div>
        ))}
      </nav>

      <Separator />

      {/* 底部用户信息 */}
      <div className="p-3">
        {isCollapsed ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="flex h-9 w-full items-center justify-center rounded-md bg-muted text-xs font-semibold text-foreground">
                A
              </div>
            </TooltipTrigger>
            <TooltipContent side="right">Admin · 超级管理员</TooltipContent>
          </Tooltip>
        ) : (
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-muted text-xs font-semibold text-foreground">
              A
            </div>
            <div className="min-w-0 flex-1">
              <p className="truncate text-[13px] font-medium text-foreground">Admin</p>
              <p className="truncate text-[11px] text-muted-foreground">超级管理员</p>
            </div>
            <Button variant="ghost" size="icon" className="h-7 w-7" title="退出登录">
              <LogOut className="h-3.5 w-3.5" strokeWidth={1.5} />
            </Button>
          </div>
        )}
      </div>
    </aside>
  )
}
