import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import {
  FileText, FolderOpen, Tags, MessageSquare, Eye,
  PenLine, ExternalLink, Settings, Loader2, ChevronRight,
  Link2, User, Palette, Upload, FilePlus,
} from 'lucide-react'
import { useDashboardStats, useRecentPosts } from '@/hooks/use-dashboard'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search } from '@/components/search'

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

/**
 * 仪表盘 — 参考 Halo 布局风格
 * 顶部统计卡片 → 快捷访问网格 + 最近文章列表
 */
export function DashboardPage() {
  const navigate = useNavigate()
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: recentPosts, isLoading: postsLoading } = useRecentPosts(5)

  // 统计数据定义
  const statsCards = [
    { label: '文章', value: stats?.postCount ?? '—', icon: FileText, color: 'text-blue-500 bg-blue-50 dark:bg-blue-950/40', path: '/posts' },
    { label: '分类', value: stats?.categoryCount ?? '—', icon: FolderOpen, color: 'text-emerald-500 bg-emerald-50 dark:bg-emerald-950/40', path: '/categories' },
    { label: '评论', value: stats?.commentPending ?? '—', icon: MessageSquare, color: 'text-amber-500 bg-amber-50 dark:bg-amber-950/40', path: '/comments' },
    { label: '标签', value: stats?.tagCount ?? '—', icon: Tags, color: 'text-violet-500 bg-violet-50 dark:bg-violet-950/40', path: '/tags' },
  ]

  // 快捷访问定义
  const quickActions = [
    { icon: User, label: '个人中心', color: 'text-blue-500 bg-blue-50 dark:bg-blue-950/30', onClick: () => navigate('/profile') },
    { icon: ExternalLink, label: '查看站点', color: 'text-emerald-500 bg-emerald-50 dark:bg-emerald-950/30', onClick: () => window.open('/', '_blank') },
    { icon: PenLine, label: '创建文章', color: 'text-sky-500 bg-sky-50 dark:bg-sky-950/30', onClick: () => navigate('/posts/new') },
    { icon: FilePlus, label: '创建页面', color: 'text-orange-500 bg-orange-50 dark:bg-orange-950/30', onClick: () => navigate('/pages/new') },
    { icon: Upload, label: '附件上传', color: 'text-teal-500 bg-teal-50 dark:bg-teal-950/30', onClick: () => navigate('/posts/new') },
    { icon: Palette, label: '主题管理', color: 'text-purple-500 bg-purple-50 dark:bg-purple-950/30', onClick: () => navigate('/settings') },
    { icon: Link2, label: '友情链接', color: 'text-cyan-500 bg-cyan-50 dark:bg-cyan-950/30', onClick: () => navigate('/links') },
    { icon: Settings, label: '系统设置', color: 'text-zinc-500 bg-zinc-100 dark:bg-zinc-800', onClick: () => navigate('/settings') },
  ]

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ml-auto' />
      </Header>
      <Main>

      {/* 统计卡片横排 */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4 mb-6">
        {statsCards.map((card) => {
          const Icon = card.icon
          return (
            <div
              key={card.label}
              className="flex items-center gap-4 px-5 py-4 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors"
              onClick={() => navigate(card.path)}
            >
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 ${card.color}`}>
                <Icon className="w-5 h-5" />
              </div>
              <div>
                <p className="text-xs text-zinc-500 font-medium">{card.label}</p>
                {statsLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin text-zinc-400 mt-1" />
                ) : (
                  <p className="text-xl font-bold text-zinc-900 dark:text-zinc-100 mt-0.5 tabular-nums">{card.value}</p>
                )}
              </div>
            </div>
          )
        })}
      </div>

      {/* 快捷访问 + 最近文章 双栏 */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-6">

        {/* 快捷访问 — 左侧 3/5 宽度 */}
        <div className="md:col-span-3 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden">
          <div className="px-5 py-4 border-b border-zinc-100 dark:border-zinc-800">
            <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">快捷访问</h2>
          </div>
          <div className="grid grid-cols-3 gap-px bg-zinc-100 dark:bg-zinc-800">
            {quickActions.map((action) => {
              const Icon = action.icon
              return (
                <div
                  key={action.label}
                  className="flex items-center gap-3 px-4 py-4 bg-white dark:bg-zinc-900 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/60 transition-colors group"
                  onClick={action.onClick}
                >
                  <div className={`w-9 h-9 rounded-lg flex items-center justify-center shrink-0 ${action.color}`}>
                    <Icon className="w-4.5 h-4.5" />
                  </div>
                  <span className="text-sm text-zinc-700 dark:text-zinc-300 flex-1">{action.label}</span>
                  <ChevronRight className="w-3.5 h-3.5 text-zinc-300 dark:text-zinc-600 group-hover:text-zinc-400 dark:group-hover:text-zinc-500 transition-colors" />
                </div>
              )
            })}
          </div>
        </div>

        {/* 最近文章 — 右侧 2/5 宽度 */}
        <div className="md:col-span-2 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden flex flex-col">
          <div className="px-5 py-4 border-b border-zinc-100 dark:border-zinc-800 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">最近文章</h2>
            <button
              className="text-xs text-blue-500 hover:text-blue-600 cursor-pointer transition-colors"
              onClick={() => navigate('/posts')}
            >
              查看全部
            </button>
          </div>
          <div className="flex-1 min-h-0">
            {postsLoading ? (
              <div className="flex justify-center py-12">
                <Loader2 className="w-5 h-5 animate-spin text-zinc-400" />
              </div>
            ) : recentPosts && recentPosts.length > 0 ? (
              <div>
                {recentPosts.map((post, idx) => (
                  <div key={post.id}>
                    {idx > 0 && <div className="h-px bg-zinc-50 dark:bg-zinc-800 mx-4" />}
                    <div
                      className="flex items-start gap-3 px-5 py-3.5 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors cursor-pointer"
                      onClick={() => navigate(`/posts/${post.id}/edit`)}
                    >
                      <div className="flex-1 min-w-0">
                        <p className="text-sm text-zinc-800 dark:text-zinc-200 truncate leading-relaxed">{post.title}</p>
                        <p className="text-xs text-zinc-400 mt-1">{formatDate(post.createdAt)}</p>
                      </div>
                      <span className={`text-[10px] px-2 py-0.5 rounded-full font-medium shrink-0 mt-0.5 ${
                        post.status === 'published'
                          ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-400'
                          : 'bg-zinc-100 text-zinc-500 dark:bg-zinc-800 dark:text-zinc-400'
                      }`}>
                        {post.status === 'published' ? '已发布' : '草稿'}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <FileText className="w-8 h-8 text-zinc-200 dark:text-zinc-700 mb-2" />
                <p className="text-sm text-zinc-500">还没有文章</p>
                <p className="text-xs text-zinc-400 mt-1">开始写第一篇吧</p>
                <Button variant="outline" size="sm" className="mt-3 shadow-none" onClick={() => navigate('/posts/new')}>
                  <PenLine className="w-3.5 h-3.5 mr-1.5" /> 写文章
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>

      </Main>
    </>
  )
}
