import { useState } from 'react'
import {
  Search,
  Plus,
  Eye,
  MessageSquare,
  Pencil,
  Trash2,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/cn'
import { usePosts, useCategories } from '@/hooks/use-posts'
import type { PostStatus } from '@/types/post'

/** 状态标签配置 */
const statusConfig: Record<PostStatus, { label: string; className: string }> = {
  published: {
    label: '已发布',
    className: 'border-emerald-300 bg-emerald-50 text-emerald-700',
  },
  draft: {
    label: '草稿',
    className: 'border-amber-300 bg-amber-50 text-amber-700',
  },
  archived: {
    label: '已归档',
    className: 'border-slate-300 bg-slate-50 text-slate-500',
  },
}

/** 格式化日期 */
function formatDate(dateStr: string | null) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

/**
 * 文章管理页面
 * 包含搜索筛选、文章列表表格和分页
 */
export function PostsPage() {
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<PostStatus | 'all'>('all')
  const [category, setCategory] = useState('')
  const [page, setPage] = useState(1)
  const pageSize = 5

  const { data, isLoading } = usePosts({ page, pageSize, keyword, status, category })
  const { data: categories } = useCategories()

  const totalPages = data ? Math.ceil(data.total / pageSize) : 0

  return (
    <div>
      {/* 页面标题区 */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--kite-text-heading)]">
            文章管理
          </h1>
          <p className="mt-1 text-sm text-[var(--kite-text-muted)]">
            管理你的所有博客文章
          </p>
        </div>
        <button className="flex h-9 items-center gap-2 border border-[var(--kite-accent)] bg-[var(--kite-accent)] px-4 text-sm font-medium text-white transition-colors duration-100 hover:bg-[#333] cursor-pointer">
          <Plus className="h-4 w-4" strokeWidth={1.5} />
          新建文章
        </button>
      </div>

      {/* 搜索与筛选栏 */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        {/* 搜索框 */}
        <div className="relative flex-1 min-w-[240px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--kite-text-muted)]" strokeWidth={1.5} />
          <input
            type="text"
            placeholder="搜索文章标题、摘要或标签…"
            value={keyword}
            onChange={(e) => { setKeyword(e.target.value); setPage(1) }}
            className="h-9 w-full border border-[var(--kite-border)] bg-[var(--kite-bg)] pl-9 pr-3 text-sm text-[var(--kite-text)] outline-none transition-colors duration-100 placeholder:text-[var(--kite-text-muted)] focus:border-[var(--kite-accent)]"
          />
        </div>

        {/* 状态筛选 */}
        <select
          value={status}
          onChange={(e) => { setStatus(e.target.value as PostStatus | 'all'); setPage(1) }}
          className="h-9 border border-[var(--kite-border)] bg-[var(--kite-bg)] px-3 text-sm text-[var(--kite-text)] outline-none transition-colors duration-100 focus:border-[var(--kite-accent)] cursor-pointer"
        >
          <option value="all">全部状态</option>
          <option value="published">已发布</option>
          <option value="draft">草稿</option>
          <option value="archived">已归档</option>
        </select>

        {/* 分类筛选 */}
        <select
          value={category}
          onChange={(e) => { setCategory(e.target.value); setPage(1) }}
          className="h-9 border border-[var(--kite-border)] bg-[var(--kite-bg)] px-3 text-sm text-[var(--kite-text)] outline-none transition-colors duration-100 focus:border-[var(--kite-accent)] cursor-pointer"
        >
          <option value="">全部分类</option>
          {categories?.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </div>

      {/* 文章列表表格 */}
      <div className="border border-[var(--kite-border)] bg-[var(--kite-bg)]">
        {/* 表头 */}
        <div className="grid grid-cols-[1fr_100px_100px_80px_80px_100px_90px] gap-4 border-b border-[var(--kite-border)] px-5 py-3 text-xs font-medium uppercase tracking-wider text-[var(--kite-text-muted)]">
          <span>标题</span>
          <span>分类</span>
          <span>状态</span>
          <span className="text-center">浏览</span>
          <span className="text-center">评论</span>
          <span>更新日期</span>
          <span className="text-center">操作</span>
        </div>

        {/* 加载态 */}
        {isLoading && (
          <div className="flex items-center justify-center py-16 text-sm text-[var(--kite-text-muted)]">
            加载中…
          </div>
        )}

        {/* 空态 */}
        {!isLoading && data?.list.length === 0 && (
          <div className="flex flex-col items-center justify-center py-16 text-sm text-[var(--kite-text-muted)]">
            <p>没有找到匹配的文章</p>
          </div>
        )}

        {/* 数据行 */}
        {data?.list.map((post, index) => {
          const statusInfo = statusConfig[post.status]
          return (
            <div
              key={post.id}
              className={cn(
                'group grid grid-cols-[1fr_100px_100px_80px_80px_100px_90px] gap-4 px-5 py-3.5 text-sm transition-colors duration-100 hover:bg-[var(--kite-bg-hover)]',
                index < (data.list.length - 1) && 'border-b border-[var(--kite-border)]'
              )}
            >
              {/* 标题 + 摘要 */}
              <div className="min-w-0">
                <p className="truncate font-medium text-[var(--kite-text-heading)]">
                  {post.title}
                </p>
                <p className="mt-0.5 truncate text-xs text-[var(--kite-text-muted)]">
                  {post.summary}
                </p>
              </div>

              {/* 分类 */}
              <div className="flex items-center">
                <span className="text-[var(--kite-text)]">{post.category}</span>
              </div>

              {/* 状态标签 */}
              <div className="flex items-center">
                <span className={cn('inline-flex items-center border px-2 py-0.5 text-xs font-medium', statusInfo.className)}>
                  {statusInfo.label}
                </span>
              </div>

              {/* 浏览数 */}
              <div className="flex items-center justify-center gap-1 text-[var(--kite-text-muted)]">
                <Eye className="h-3.5 w-3.5" strokeWidth={1.5} />
                <span className="text-xs">{post.viewCount}</span>
              </div>

              {/* 评论数 */}
              <div className="flex items-center justify-center gap-1 text-[var(--kite-text-muted)]">
                <MessageSquare className="h-3.5 w-3.5" strokeWidth={1.5} />
                <span className="text-xs">{post.commentCount}</span>
              </div>

              {/* 更新日期 */}
              <div className="flex items-center text-xs text-[var(--kite-text-muted)]">
                {formatDate(post.updatedAt)}
              </div>

              {/* 操作按钮 */}
              <div className="flex items-center justify-center gap-1">
                <button
                  className="flex h-7 w-7 items-center justify-center border border-transparent text-[var(--kite-text-muted)] transition-colors duration-100 hover:border-[var(--kite-border)] hover:text-[var(--kite-text-heading)] cursor-pointer"
                  title="编辑"
                >
                  <Pencil className="h-3.5 w-3.5" strokeWidth={1.5} />
                </button>
                <button
                  className="flex h-7 w-7 items-center justify-center border border-transparent text-[var(--kite-text-muted)] transition-colors duration-100 hover:border-red-300 hover:text-red-600 cursor-pointer"
                  title="删除"
                >
                  <Trash2 className="h-3.5 w-3.5" strokeWidth={1.5} />
                </button>
              </div>
            </div>
          )
        })}
      </div>

      {/* 分页控制 */}
      {data && data.total > 0 && (
        <div className="mt-4 flex items-center justify-between text-sm text-[var(--kite-text-muted)]">
          <span>
            共 {data.total} 篇文章，第 {data.page}/{totalPages} 页
          </span>
          <div className="flex items-center gap-1">
            <button
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
              className="flex h-8 w-8 items-center justify-center border border-[var(--kite-border)] transition-colors duration-100 hover:border-[var(--kite-border-hover)] disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
            >
              <ChevronLeft className="h-4 w-4" strokeWidth={1.5} />
            </button>
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((n) => (
              <button
                key={n}
                onClick={() => setPage(n)}
                className={cn(
                  'flex h-8 w-8 items-center justify-center border text-sm transition-colors duration-100 cursor-pointer',
                  n === page
                    ? 'border-[var(--kite-accent)] bg-[var(--kite-accent)] text-white font-medium'
                    : 'border-[var(--kite-border)] hover:border-[var(--kite-border-hover)]'
                )}
              >
                {n}
              </button>
            ))}
            <button
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
              className="flex h-8 w-8 items-center justify-center border border-[var(--kite-border)] transition-colors duration-100 hover:border-[var(--kite-border-hover)] disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
            >
              <ChevronRight className="h-4 w-4" strokeWidth={1.5} />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
