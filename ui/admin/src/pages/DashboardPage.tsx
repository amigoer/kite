import { useNavigate } from 'react-router'
import {
  FileText,
  FolderTree,
  Eye,
  MessageSquare,
  TrendingUp,
  Clock,
  PenSquare,
  ArrowUpRight,
  Activity,
  Server,
  Database,
  Cpu,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

/* Mock 仪表盘数据 */
const statsCards = [
  { label: '文章总数', value: '34', change: '+3', icon: FileText, trend: 'up' as const },
  { label: '分类数量', value: '6', change: '0', icon: FolderTree, trend: 'neutral' as const },
  { label: '本月访问', value: '12,847', change: '+18%', icon: Eye, trend: 'up' as const },
  { label: '待审评论', value: '7', change: '+2', icon: MessageSquare, trend: 'up' as const },
]

const weeklyTraffic = [
  { day: '周一', views: 1420 },
  { day: '周二', views: 1680 },
  { day: '周三', views: 2310 },
  { day: '周四', views: 1890 },
  { day: '周五', views: 2540 },
  { day: '周六', views: 1120 },
  { day: '周日', views: 980 },
]

const recentPosts = [
  { title: 'AI 驱动的博客引擎：DeepSeek 自动摘要集成方案', status: 'draft' as const, date: '2026-03-14', views: 0 },
  { title: 'Tailwind CSS v4：从零到一的设计系统构建', status: 'draft' as const, date: '2026-03-10', views: 0 },
  { title: 'Docker Compose 生产环境最佳实践', status: 'published' as const, date: '2026-03-02', views: 2104 },
  { title: '从 Webpack 到 Vite：大型项目迁移实录', status: 'published' as const, date: '2026-02-26', views: 1567 },
  { title: 'React 19 新特性全解析：Server Components 与 Actions', status: 'published' as const, date: '2026-02-21', views: 3456 },
]

const activityLog = [
  { action: '发布了文章', target: 'Docker Compose 生产环境最佳实践', time: '3 天前', type: 'publish' as const },
  { action: '创建了草稿', target: 'AI 驱动的博客引擎：DeepSeek 自动摘要集成方案', time: '1 天前', type: 'draft' as const },
  { action: '新增分类', target: '开源项目', time: '2 周前', type: 'category' as const },
  { action: '更新了设置', target: '站点描述与 SEO 关键词', time: '2 周前', type: 'setting' as const },
  { action: '修改了文章', target: 'GORM 高级用法：自定义类型、Hook 与性能优化', time: '3 周前', type: 'edit' as const },
  { action: '回复了评论', target: 'React 19 新特性全解析 下 #42', time: '3 周前', type: 'comment' as const },
]

const systemInfo = [
  { label: '引擎版本', value: 'Kite v0.1.0', icon: Cpu },
  { label: '运行环境', value: 'Go 1.23 / Linux', icon: Server },
  { label: '数据库', value: 'SQLite 3.45', icon: Database },
  { label: '运行时长', value: '14 天 6 小时', icon: Activity },
]

const activityIcon: Record<string, string> = {
  publish: '📤', draft: '📝', category: '📁', setting: '⚙️', edit: '✏️', comment: '💬',
}

/**
 * 仪表盘页面 - 使用 shadcn Card / Badge / Button / Separator
 */
export function DashboardPage() {
  const navigate = useNavigate()
  const maxViews = Math.max(...weeklyTraffic.map((d) => d.views))

  return (
    <div>
      {/* 页面标题 */}
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">仪表盘</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          欢迎回到 Kite 后台管理 — 以下是你的站点概览
        </p>
      </div>

      {/* 统计卡片 */}
      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {statsCards.map((card) => {
          const Icon = card.icon
          return (
            <Card key={card.label}>
              <CardContent className="pt-6">
                <div className="flex items-center gap-4">
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-md bg-muted">
                    <Icon className="h-5 w-5 text-muted-foreground" strokeWidth={1.5} />
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">{card.label}</p>
                    <p className="mt-0.5 text-2xl font-bold">{card.value}</p>
                  </div>
                </div>
                {card.change !== '0' && (
                  <>
                    <Separator className="my-3" />
                    <div className="flex items-center gap-1">
                      <TrendingUp className="h-3 w-3 text-emerald-600" strokeWidth={1.5} />
                      <span className="text-xs text-emerald-600">{card.change}</span>
                      <span className="text-xs text-muted-foreground">较上月</span>
                    </div>
                  </>
                )}
              </CardContent>
            </Card>
          )
        })}
      </div>

      {/* 中间区域：七日趋势 + 最近文章 */}
      <div className="mb-6 grid grid-cols-1 gap-4 lg:grid-cols-5">
        {/* 折线图 */}
        <Card className="flex flex-col lg:col-span-3">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">七日访问趋势</CardTitle>
              <span className="text-xs text-muted-foreground">最近 7 天</span>
            </div>
          </CardHeader>
          <CardContent className="flex-1">
            {(() => {
              const chartW = 560
              const chartH = 140
              const padX = 40
              const padY = 10
              const padBottom = 18
              const innerW = chartW - padX * 2
              const innerH = chartH - padY - padBottom
              const minV = 0
              const maxV = maxViews * 1.05

              const points = weeklyTraffic.map((d, i) => ({
                x: padX + (i / (weeklyTraffic.length - 1)) * innerW,
                y: padY + innerH - ((d.views - minV) / (maxV - minV)) * innerH,
                views: d.views,
                day: d.day,
              }))

              const linePath = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x},${p.y}`).join(' ')
              const areaPath = `${linePath} L${points[points.length - 1].x},${padY + innerH} L${points[0].x},${padY + innerH} Z`
              const yTicks = [0, Math.round(maxV * 0.33), Math.round(maxV * 0.66), Math.round(maxV)]

              return (
                <svg viewBox={`0 0 ${chartW} ${chartH}`} className="w-full flex-1" preserveAspectRatio="xMidYEnd meet">
                  {yTicks.map((tick) => {
                    const y = padY + innerH - ((tick - minV) / (maxV - minV)) * innerH
                    return (
                      <g key={tick}>
                        <line x1={padX} y1={y} x2={chartW - padX} y2={y} className="stroke-border" strokeWidth={0.5} />
                        <text x={padX - 6} y={y + 3} textAnchor="end" className="fill-muted-foreground" fontSize={9}>
                          {tick >= 1000 ? `${(tick / 1000).toFixed(1)}k` : tick}
                        </text>
                      </g>
                    )
                  })}
                  <path d={areaPath} className="fill-primary/5" />
                  <path d={linePath} fill="none" className="stroke-primary" strokeWidth={2} />
                  {points.map((p, i) => (
                    <g key={i}>
                      <text x={p.x} y={padY + innerH + 14} textAnchor="middle" className="fill-muted-foreground" fontSize={10}>
                        {p.day}
                      </text>
                      <circle cx={p.x} cy={p.y} r={3} className="fill-background stroke-primary" strokeWidth={1.5} />
                    </g>
                  ))}
                </svg>
              )
            })()}
          </CardContent>
        </Card>

        {/* 最近文章 */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">最近文章</CardTitle>
              <Button variant="ghost" size="sm" onClick={() => navigate('/posts')}>
                查看全部 <ArrowUpRight className="ml-1 h-3 w-3" />
              </Button>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            {recentPosts.map((post, i) => (
              <div key={i}>
                {i > 0 && <Separator />}
                <div className="flex items-center justify-between px-6 py-3">
                  <div className="min-w-0 flex-1 pr-3">
                    <p className="truncate text-sm font-medium">{post.title}</p>
                    <p className="mt-0.5 text-xs text-muted-foreground">{post.date}</p>
                  </div>
                  <Badge variant={post.status === 'published' ? 'default' : 'secondary'}>
                    {post.status === 'published' ? '已发布' : '草稿'}
                  </Badge>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* 底部区域 */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* 近期动态 */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-sm">近期动态</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {activityLog.map((item, i) => (
              <div key={i}>
                {i > 0 && <Separator />}
                <div className="flex items-start gap-3 px-6 py-3">
                  <span className="mt-0.5 text-sm">{activityIcon[item.type]}</span>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm">
                      <span className="font-medium">{item.action}</span>
                      {' '}
                      <span className="text-muted-foreground">「{item.target}」</span>
                    </p>
                  </div>
                  <div className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
                    <Clock className="h-3 w-3" strokeWidth={1.5} />
                    {item.time}
                  </div>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>

        {/* 右侧 */}
        <div className="space-y-4">
          {/* 快捷操作 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">快捷操作</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-2">
                <Button variant="outline" className="justify-start gap-2" onClick={() => navigate('/posts')}>
                  <PenSquare className="h-4 w-4" strokeWidth={1.5} />
                  写文章
                </Button>
                <Button variant="outline" className="justify-start gap-2" onClick={() => navigate('/categories')}>
                  <FolderTree className="h-4 w-4" strokeWidth={1.5} />
                  管理分类
                </Button>
                <Button variant="outline" className="justify-start gap-2" onClick={() => navigate('/settings')}>
                  <Eye className="h-4 w-4" strokeWidth={1.5} />
                  预览站点
                </Button>
                <Button variant="outline" className="justify-start gap-2" onClick={() => navigate('/settings')}>
                  <Activity className="h-4 w-4" strokeWidth={1.5} />
                  系统设置
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* 系统信息 */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">系统信息</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {systemInfo.map((item, i) => {
                const Icon = item.icon
                return (
                  <div key={i}>
                    {i > 0 && <Separator />}
                    <div className="flex items-center justify-between px-6 py-2.5">
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <Icon className="h-3.5 w-3.5" strokeWidth={1.5} />
                        {item.label}
                      </div>
                      <span className="text-sm font-medium">{item.value}</span>
                    </div>
                  </div>
                )
              })}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
