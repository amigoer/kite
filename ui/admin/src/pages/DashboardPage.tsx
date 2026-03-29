import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import {
  FileText, FolderOpen, Tags, MessageSquare, Eye,
  PenLine, ExternalLink, Settings, Loader2, ChevronRight,
  Link2, User, Palette, Upload, FilePlus,
  TrendingUp, BarChart3, CalendarDays, ArrowUpRight,
} from 'lucide-react'
import { useDashboardStats, useRecentPosts, useRecentComments } from '@/hooks/use-dashboard'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search } from '@/components/search'

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  return formatDate(dateStr)
}

/**
 * 仪表盘 — 丰富的管理概览
 */
export function DashboardPage() {
  const navigate = useNavigate()
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: recentPosts, isLoading: postsLoading } = useRecentPosts(5)
  const { data: recentComments, isLoading: commentsLoading } = useRecentComments(5)

  // 主统计卡片
  const statsCards = [
    { label: '文章', value: stats?.postCount ?? '—', icon: FileText, color: 'text-blue-500 bg-blue-50 dark:bg-blue-950/40', path: '/posts' },
    { label: '分类', value: stats?.categoryCount ?? '—', icon: FolderOpen, color: 'text-emerald-500 bg-emerald-50 dark:bg-emerald-950/40', path: '/categories' },
    { label: '评论', value: stats?.commentPending ?? '—', icon: MessageSquare, color: 'text-amber-500 bg-amber-50 dark:bg-amber-950/40', path: '/comments' },
    { label: '标签', value: stats?.tagCount ?? '—', icon: Tags, color: 'text-violet-500 bg-violet-50 dark:bg-violet-950/40', path: '/tags' },
    { label: '总浏览', value: stats?.totalViews ?? '—', icon: Eye, color: 'text-rose-500 bg-rose-50 dark:bg-rose-950/40', path: '/posts' },
  ]

  // 快捷访问
  const quickActions = [
    { icon: PenLine, label: '创建文章', color: 'text-sky-500 bg-sky-50 dark:bg-sky-950/30', onClick: () => navigate('/posts/new') },
    { icon: FilePlus, label: '创建页面', color: 'text-orange-500 bg-orange-50 dark:bg-orange-950/30', onClick: () => navigate('/pages/new') },
    { icon: ExternalLink, label: '查看站点', color: 'text-emerald-500 bg-emerald-50 dark:bg-emerald-950/30', onClick: () => window.open('/', '_blank') },
    { icon: User, label: '个人中心', color: 'text-blue-500 bg-blue-50 dark:bg-blue-950/30', onClick: () => navigate('/profile') },
    { icon: Upload, label: '附件上传', color: 'text-teal-500 bg-teal-50 dark:bg-teal-950/30', onClick: () => navigate('/posts/new') },
    { icon: Link2, label: '友情链接', color: 'text-cyan-500 bg-cyan-50 dark:bg-cyan-950/30', onClick: () => navigate('/links') },
    { icon: Palette, label: '主题管理', color: 'text-purple-500 bg-purple-50 dark:bg-purple-950/30', onClick: () => navigate('/settings') },
    { icon: Settings, label: '系统设置', color: 'text-zinc-500 bg-zinc-100 dark:bg-zinc-800', onClick: () => navigate('/settings') },
  ]

  // 文章状态分布
  const statusData = [
    { label: '已发布', count: stats?.publishedCount ?? 0, color: 'bg-emerald-500' },
    { label: '草稿', count: stats?.draftCount ?? 0, color: 'bg-zinc-400' },
    { label: '定时中', count: stats?.scheduledCount ?? 0, color: 'bg-blue-500' },
  ]
  const totalStatus = statusData.reduce((s, d) => s + (d.count as number), 0) || 1

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ml-auto' />
      </Header>
      <Main>

      {/* 统计卡片 — 5 列 */}
      <div className="grid grid-cols-2 gap-4 md:grid-cols-5 mb-6">
        {statsCards.map((card) => {
          const Icon = card.icon
          return (
            <div
              key={card.label}
              className="flex items-center gap-4 px-5 py-4 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-all hover:shadow-sm group"
              onClick={() => navigate(card.path)}
            >
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 ${card.color} transition-transform group-hover:scale-110`}>
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

      {/* 内容分布 + 快捷访问 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">

        {/* 内容概览 */}
        <div className="rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden">
          <div className="px-5 py-4 border-b border-zinc-100 dark:border-zinc-800 flex items-center gap-2">
            <BarChart3 className="w-4 h-4 text-zinc-400" />
            <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">内容概览</h2>
          </div>
          <div className="p-5 space-y-4">
            {/* 状态分布条 */}
            <div>
              <div className="flex gap-1 h-3 rounded-full overflow-hidden bg-zinc-100 dark:bg-zinc-800">
                {statusData.map((s) => (
                  <div
                    key={s.label}
                    className={`${s.color} transition-all duration-500 rounded-full`}
                    style={{ width: `${((s.count as number) / totalStatus) * 100}%`, minWidth: (s.count as number) > 0 ? '8px' : '0' }}
                  />
                ))}
              </div>
              <div className="flex items-center gap-4 mt-3">
                {statusData.map((s) => (
                  <div key={s.label} className="flex items-center gap-1.5">
                    <div className={`w-2.5 h-2.5 rounded-full ${s.color}`} />
                    <span className="text-xs text-zinc-500">{s.label}</span>
                    <span className="text-xs font-semibold text-zinc-700 dark:text-zinc-300 tabular-nums">{s.count}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* 关键指标 */}
            <div className="grid grid-cols-2 gap-3 pt-2">
              <div className="p-3 rounded-lg bg-zinc-50 dark:bg-zinc-800/50">
                <div className="flex items-center gap-1.5 text-zinc-500 mb-1">
                  <TrendingUp className="w-3.5 h-3.5" />
                  <span className="text-[11px] font-medium">发布率</span>
                </div>
                <p className="text-lg font-bold text-zinc-900 dark:text-zinc-100 tabular-nums">
                  {stats ? `${Math.round((stats.publishedCount / Math.max(stats.postCount, 1)) * 100)}%` : '—'}
                </p>
              </div>
              <div className="p-3 rounded-lg bg-zinc-50 dark:bg-zinc-800/50">
                <div className="flex items-center gap-1.5 text-zinc-500 mb-1">
                  <Eye className="w-3.5 h-3.5" />
                  <span className="text-[11px] font-medium">篇均浏览</span>
                </div>
                <p className="text-lg font-bold text-zinc-900 dark:text-zinc-100 tabular-nums">
                  {stats && stats.postCount > 0 ? Math.round(stats.totalViews / stats.postCount) : '—'}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* 快捷访问 — 占 2 列 */}
        <div className="md:col-span-2 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden">
          <div className="px-5 py-4 border-b border-zinc-100 dark:border-zinc-800 flex items-center gap-2">
            <ArrowUpRight className="w-4 h-4 text-zinc-400" />
            <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">快捷访问</h2>
          </div>
          <div className="grid grid-cols-4 gap-px bg-zinc-100 dark:bg-zinc-800">
            {quickActions.map((action) => {
              const Icon = action.icon
              return (
                <div
                  key={action.label}
                  className="flex items-center gap-3 px-4 py-4 bg-white dark:bg-zinc-900 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/60 transition-colors group"
                  onClick={action.onClick}
                >
                  <div className={`w-9 h-9 rounded-lg flex items-center justify-center shrink-0 ${action.color} transition-transform group-hover:scale-110`}>
                    <Icon className="w-4.5 h-4.5" />
                  </div>
                  <span className="text-sm text-zinc-700 dark:text-zinc-300 flex-1">{action.label}</span>
                  <ChevronRight className="w-3.5 h-3.5 text-zinc-300 dark:text-zinc-600 group-hover:text-zinc-400 dark:group-hover:text-zinc-500 transition-colors" />
                </div>
              )
            })}
          </div>
        </div>
      </div>

      {/* 最近文章 + 最近评论 双栏 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

        {/* 最近文章 */}
        <div className="rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden flex flex-col">
          <div className="px-5 py-4 border-b border-zinc-100 dark:border-zinc-800 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CalendarDays className="w-4 h-4 text-zinc-400" />
              <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">最近文章</h2>
            </div>
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
                      className="flex items-center gap-3 px-5 py-3.5 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors cursor-pointer group"
                      onClick={() => navigate(`/posts/${post.slug || post.id}/edit`)}
                    >
                      <div className="flex-1 min-w-0">
                        <p className="text-sm text-zinc-800 dark:text-zinc-200 truncate leading-relaxed group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">{post.title}</p>
                        <div className="flex items-center gap-3 mt-1">
                          <span className="text-xs text-zinc-400">{formatDate(post.createdAt)}</span>
                          {post.category && (
                            <span className="text-xs text-zinc-400 flex items-center gap-1">
                              <FolderOpen className="w-3 h-3" /> {post.category.name}
                            </span>
                          )}
                          <span className="text-xs text-zinc-400 flex items-center gap-1">
                            <Eye className="w-3 h-3" /> {post.viewCount ?? 0}
                          </span>
                        </div>
                      </div>
                      <span className={`text-[10px] px-2 py-0.5 rounded-full font-medium shrink-0 ${
                        post.status === 'published'
                          ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-400'
                          : post.status === 'scheduled'
                          ? 'bg-blue-50 text-blue-600 dark:bg-blue-950/40 dark:text-blue-400'
                          : 'bg-zinc-100 text-zinc-500 dark:bg-zinc-800 dark:text-zinc-400'
                      }`}>
                        {post.status === 'published' ? '已发布' : post.status === 'scheduled' ? '定时中' : '草稿'}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <FileText className="w-8 h-8 text-zinc-200 dark:text-zinc-700 mb-2" />
                <p className="text-sm text-zinc-500">还没有文章</p>
                <Button variant="outline" size="sm" className="mt-3 shadow-none" onClick={() => navigate('/posts/new')}>
                  <PenLine className="w-3.5 h-3.5 mr-1.5" /> 写文章
                </Button>
              </div>
            )}
          </div>
        </div>

        {/* 最近评论 */}
        <div className="rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 overflow-hidden flex flex-col">
          <div className="px-5 py-4 border-b border-zinc-100 dark:border-zinc-800 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <MessageSquare className="w-4 h-4 text-zinc-400" />
              <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">最近评论</h2>
            </div>
            <button
              className="text-xs text-blue-500 hover:text-blue-600 cursor-pointer transition-colors"
              onClick={() => navigate('/comments')}
            >
              查看全部
            </button>
          </div>
          <div className="flex-1 min-h-0">
            {commentsLoading ? (
              <div className="flex justify-center py-12">
                <Loader2 className="w-5 h-5 animate-spin text-zinc-400" />
              </div>
            ) : recentComments && recentComments.length > 0 ? (
              <div>
                {recentComments.map((comment, idx) => (
                  <div key={comment.id}>
                    {idx > 0 && <div className="h-px bg-zinc-50 dark:bg-zinc-800 mx-4" />}
                    <div className="px-5 py-3.5 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors">
                      <div className="flex items-center gap-2 mb-1.5">
                        <div className="w-6 h-6 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center text-[10px] text-white font-bold shrink-0">
                          {comment.author?.[0]?.toUpperCase() ?? '?'}
                        </div>
                        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300 truncate">{comment.author}</span>
                        <span className="text-xs text-zinc-400 shrink-0">{timeAgo(comment.createdAt)}</span>
                        <span className={`ml-auto text-[10px] px-1.5 py-0.5 rounded-full font-medium shrink-0 ${
                          comment.status === 'approved'
                            ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-400'
                            : 'bg-amber-50 text-amber-600 dark:bg-amber-950/40 dark:text-amber-400'
                        }`}>
                          {comment.status === 'approved' ? '已通过' : '待审'}
                        </span>
                      </div>
                      <p className="text-xs text-zinc-600 dark:text-zinc-400 line-clamp-2 leading-relaxed">{comment.content}</p>
                      {comment.post && (
                        <p className="text-[11px] text-zinc-400 mt-1.5 flex items-center gap-1 truncate">
                          <FileText className="w-3 h-3 shrink-0" />
                          <span className="truncate">{comment.post.title}</span>
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <MessageSquare className="w-8 h-8 text-zinc-200 dark:text-zinc-700 mb-2" />
                <p className="text-sm text-zinc-500">暂无评论</p>
                <p className="text-xs text-zinc-400 mt-1">有读者留言时会显示在这里</p>
              </div>
            )}
          </div>
        </div>
      </div>

      </Main>
    </>
  )
}
