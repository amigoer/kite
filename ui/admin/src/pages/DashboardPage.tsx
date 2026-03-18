import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import {
  FileText, FolderOpen, Tags, MessageSquare,
  PenLine, ExternalLink, Settings, Loader2,
} from 'lucide-react'
import { useDashboardStats, useRecentPosts } from '@/hooks/use-dashboard'

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
    <div>
      <div className="mb-8">
        <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50 tracking-tight">仪表盘</h1>
        <p className="text-sm text-zinc-500 mt-1">站点概览</p>
      </div>

      {/* 统计卡片 — 纯白 + border + shadow-sm，裸图标 */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        {statsCards.map((card) => {
          const Icon = card.icon
          return (
            <div
              key={card.label}
              className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm p-5 cursor-pointer hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors"
              onClick={() => navigate(card.path)}
            >
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-xs font-medium text-zinc-500 uppercase tracking-wider">{card.label}</p>
                  {statsLoading ? (
                    <Loader2 className="w-4 h-4 animate-spin text-zinc-300 mt-3" />
                  ) : (
                    <p className="text-3xl font-bold text-zinc-900 dark:text-zinc-50 mt-1 tabular-nums tracking-tight">{card.value}</p>
                  )}
                </div>
                <Icon className="w-5 h-5 text-zinc-500" />
              </div>
            </div>
          )
        })}
      </div>

      <div className="grid grid-cols-3 gap-4">
        {/* 快捷操作 */}
        <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm p-5">
          <h2 className="text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-4 tracking-tight">快捷操作</h2>
          <div className="grid grid-cols-2 gap-2">
            {[
              { icon: PenLine, label: '写文章', onClick: () => navigate('/posts/new') },
              { icon: FolderOpen, label: '管理分类', onClick: () => navigate('/categories') },
              { icon: ExternalLink, label: '预览站点', onClick: () => window.open('/', '_blank') },
              { icon: Settings, label: '系统设置', onClick: () => navigate('/settings') },
            ].map((a) => (
              <Button
                key={a.label}
                variant="outline"
                className="justify-start gap-2 h-8 text-[13px] shadow-sm bg-white dark:bg-zinc-900 border-zinc-200 dark:border-zinc-800 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 rounded-md"
                onClick={a.onClick}
              >
                <a.icon className="w-3.5 h-3.5 text-zinc-500" /> {a.label}
              </Button>
            ))}
          </div>
        </div>

        {/* 最近文章 */}
        <div className="col-span-2 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm">
          <div className="flex items-center justify-between px-5 py-3.5 border-b border-zinc-100 dark:border-zinc-800">
            <h2 className="text-sm font-medium text-zinc-700 dark:text-zinc-300 tracking-tight">最近文章</h2>
            <button className="text-xs text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-200 transition-colors cursor-pointer" onClick={() => navigate('/posts')}>
              查看全部 →
            </button>
          </div>
          <div>
            {postsLoading ? (
              <div className="flex justify-center py-10"><Loader2 className="w-4 h-4 animate-spin text-zinc-300" /></div>
            ) : recentPosts && recentPosts.length > 0 ? (
              <div>
                {recentPosts.map((post, i) => (
                  <div
                    key={post.id}
                    className={`flex items-center justify-between px-5 py-3 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors ${i < recentPosts.length - 1 ? 'border-b border-zinc-100 dark:border-zinc-800' : ''}`}
                    onClick={() => navigate(`/posts/${post.id}/edit`)}
                  >
                    <div className="min-w-0 flex-1 pr-3">
                      <p className="text-sm text-zinc-700 dark:text-zinc-300 truncate">{post.title}</p>
                      <p className="text-xs text-zinc-500 mt-0.5 tabular-nums">{formatDate(post.createdAt)}</p>
                    </div>
                    <span className="bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 border border-zinc-200 dark:border-zinc-700 rounded-full px-2.5 py-0.5 text-xs font-medium shrink-0">
                      {post.status === 'published' ? '已发布' : '草稿'}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="flex flex-col items-center py-12 text-center">
                <FileText className="w-8 h-8 text-zinc-300 dark:text-zinc-600 mb-2" />
                <p className="text-sm text-zinc-700 dark:text-zinc-300">还没有文章</p>
                <p className="text-xs text-zinc-500 mt-1">开始写第一篇吧</p>
                <Button variant="outline" size="sm" className="mt-3 shadow-sm border-zinc-200 dark:border-zinc-800 bg-white text-sm rounded-md" onClick={() => navigate('/posts/new')}>
                  <PenLine className="w-3.5 h-3.5 mr-1.5" /> 写文章
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
