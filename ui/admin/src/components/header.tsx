import { PanelLeftClose, PanelLeft, Settings } from 'lucide-react'
import { useLocation, useNavigate } from 'react-router'
import { useSidebarStore } from '@/stores/use-sidebar-store'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

/** 面包屑路径映射 */
const breadcrumbMap: Record<string, string> = {
  '/': '仪表盘',
  '/posts': '文章管理',
  '/categories': '分类管理',
  '/tags': '标签管理',
  '/comments': '评论管理',
  '/links': '友链管理',
  '/settings': '系统设置',
}

/**
 * 顶部 Header 组件
 * 使用 shadcn Button / Separator
 */
export function Header() {
  const { isCollapsed, toggle } = useSidebarStore()
  const location = useLocation()
  const navigate = useNavigate()
  const currentLabel = breadcrumbMap[location.pathname] || '未知页面'

  return (
    <header className="flex h-14 items-center justify-between border-b bg-background px-4">
      {/* 左侧 */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={toggle} aria-label={isCollapsed ? '展开侧边栏' : '折叠侧边栏'}>
          {isCollapsed ? <PanelLeft className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
        </Button>

        <Separator orientation="vertical" className="h-4" />

        <nav className="flex items-center gap-1.5 text-sm text-muted-foreground">
          <span>Kite</span>
          <span>/</span>
          <span className="font-medium text-foreground">{currentLabel}</span>
        </nav>
      </div>

      {/* 右侧 */}
      <div className="flex items-center gap-1">
        <Button variant="ghost" size="icon" onClick={() => navigate('/settings')} title="设置">
          <Settings className="h-4 w-4" />
        </Button>
      </div>
    </header>
  )
}
