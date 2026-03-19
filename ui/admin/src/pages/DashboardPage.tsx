import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import {
  FileText, FolderOpen, Tags, MessageSquare,
  PenLine, ExternalLink, Settings, Loader2,
} from 'lucide-react'
import { useDashboardStats, useRecentPosts } from '@/hooks/use-dashboard'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search } from '@/components/search'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

/**
 * 仪表盘 — Vercel 风格：单色克制、紧凑阴影、无彩色
 */
export function DashboardPage() {
  const navigate = useNavigate()
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: recentPosts, isLoading: postsLoading } = useRecentPosts(5)

  const statsCards = [
    { label: '文章', value: stats?.postCount ?? '—', icon: FileText, path: '/posts' },
    { label: '分类', value: stats?.categoryCount ?? '—', icon: FolderOpen, path: '/categories' },
    { label: '标签', value: stats?.tagCount ?? '—', icon: Tags, path: '/tags' },
    { label: '待审评论', value: stats?.commentPending ?? '—', icon: MessageSquare, path: '/comments' },
  ]

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ml-auto' />
      </Header>
      <Main>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4 mb-8">
        {statsCards.map((card) => {
          const Icon = card.icon
          return (
            <Card
              key={card.label}
              className="cursor-pointer hover:bg-muted/50 transition-colors"
              onClick={() => navigate(card.path)}
            >
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  {card.label}
                </CardTitle>
                <Icon className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                {statsLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin text-muted-foreground mt-1" />
                ) : (
                  <div className="text-2xl font-bold">{card.value}</div>
                )}
              </CardContent>
            </Card>
          )
        })}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* 快捷操作 */}
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle>快捷操作</CardTitle>
            <CardDescription>管理您的博客内容和设置</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-2">
            {[
              { icon: PenLine, label: '写文章', onClick: () => navigate('/posts/new') },
              { icon: FolderOpen, label: '管理分类', onClick: () => navigate('/categories') },
              { icon: ExternalLink, label: '预览站点', onClick: () => window.open('/', '_blank') },
              { icon: Settings, label: '系统设置', onClick: () => navigate('/settings') },
            ].map((a) => (
              <Button
                key={a.label}
                variant="outline"
                className="justify-start gap-2 h-9 w-full"
                onClick={a.onClick}
              >
                <a.icon className="w-4 h-4 text-muted-foreground" /> {a.label}
              </Button>
            ))}
          </CardContent>
        </Card>

        {/* 最近文章 */}
        <Card className="col-span-1 md:col-span-2">
          <CardHeader className="flex flex-row items-start justify-between">
            <div className="space-y-1">
              <CardTitle>最近文章</CardTitle>
              <CardDescription>您最近创建或修改的博客文章</CardDescription>
            </div>
            <Button variant="ghost" size="sm" onClick={() => navigate('/posts')} className="hidden sm:flex">
              查看全部
            </Button>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {postsLoading ? (
                <div className="flex justify-center py-6"><Loader2 className="w-4 h-4 animate-spin text-muted-foreground" /></div>
              ) : recentPosts && recentPosts.length > 0 ? (
                recentPosts.map((post) => (
                  <div
                    key={post.id}
                    className="flex items-center justify-between rounded-md p-3 hover:bg-muted/50 transition-colors cursor-pointer border border-transparent hover:border-border"
                    onClick={() => navigate(`/posts/${post.id}/edit`)}
                  >
                    <div className="min-w-0 flex-1 pr-3">
                      <p className="text-sm font-medium leading-none truncate">{post.title}</p>
                      <p className="text-xs text-muted-foreground mt-1.5 tabular-nums">{formatDate(post.createdAt)}</p>
                    </div>
                    <div className="ml-auto font-medium pl-2 shrink-0">
                      <span className="bg-secondary text-secondary-foreground rounded-full px-2.5 py-0.5 text-xs font-medium">
                        {post.status === 'published' ? '已发布' : '草稿'}
                      </span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="flex flex-col items-center py-8 text-center border rounded-md border-dashed">
                  <FileText className="w-8 h-8 text-muted-foreground mb-2 opacity-20" />
                  <p className="text-sm text-foreground">还没有文章</p>
                  <p className="text-xs text-muted-foreground mt-1">开始写第一篇吧</p>
                  <Button variant="outline" size="sm" className="mt-3" onClick={() => navigate('/posts/new')}>
                    <PenLine className="w-3.5 h-3.5 mr-1.5" /> 写文章
                  </Button>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
      </Main>
    </>
  )
}
