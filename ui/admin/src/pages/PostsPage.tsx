import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import { Search, Plus, Pencil, Trash2, FileText, Loader2 } from 'lucide-react'
import { usePosts, useDeletePost } from '@/hooks/use-posts'
import { useCategoryList } from '@/hooks/use-categories'
import type { PostStatus } from '@/types/post'
import { cn } from '@/lib/utils'

/** 状态标签配置 */
const statusConfig: Record<PostStatus, { label: string; variant: 'default' | 'secondary' | 'outline' }> = {
  published: { label: '已发布', variant: 'default' },
  draft: { label: '草稿', variant: 'secondary' },
  archived: { label: '已归档', variant: 'outline' },
}

/** 状态筛选标签 */
const statusTabs = [
  { key: 'all' as const, label: '全部' },
  { key: 'published' as const, label: '已发布' },
  { key: 'draft' as const, label: '草稿' },
  { key: 'archived' as const, label: '已归档' },
]

function formatDate(dateStr: string | null) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

/**
 * 文章管理页面
 */
export function PostsPage() {
  const navigate = useNavigate()
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<PostStatus | 'all'>('all')
  const [categoryId, setCategoryId] = useState('')
  const [page, setPage] = useState(1)
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; title: string } | null>(null)
  const pageSize = 10

  const { data, isLoading } = usePosts({ page, pageSize, keyword, status, categoryId: categoryId || undefined })
  const { data: categories } = useCategoryList()
  const deleteMutation = useDeletePost()

  function handleDelete() {
    if (!deleteTarget) return
    deleteMutation.mutate(deleteTarget.id)
    setDeleteTarget(null)
  }

  const totalPages = data ? Math.ceil(data.pagination.total / pageSize) : 0

  return (
    <div>
      {/* 标题区 */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50 tracking-tight">文章管理</h1>
          <p className="text-[13px] text-zinc-500 mt-1">管理所有博客文章</p>
        </div>
        <Button className="bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 shadow-sm rounded-md hover:bg-zinc-800 dark:hover:bg-zinc-200 text-sm h-9 transition-colors" onClick={() => navigate('/posts/new')}>
          <Plus className="w-4 h-4 mr-1.5" /> 新建文章
        </Button>
      </div>

      {/* 状态筛选 */}
      <div className="flex gap-1 mb-4 p-1 bg-[#FAFAFA] dark:bg-zinc-800/40 rounded-lg w-fit">
        {statusTabs.map((tab) => (
          <button
            key={tab.key}
            className={cn(
              'px-3.5 py-1.5 text-sm rounded-md transition-all cursor-pointer',
              status === tab.key
                ? 'bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-50 shadow-sm font-medium'
                : 'text-zinc-500 dark:text-zinc-400 hover:text-zinc-700 dark:hover:text-zinc-200 font-normal'
            )}
            onClick={() => { setStatus(tab.key); setPage(1) }}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* 搜索与筛选 */}
      <div className="flex gap-3 mb-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-[15px] h-[15px] text-zinc-400" />
          <Input
            placeholder="搜索文章标题、摘要…"
            value={keyword}
            onChange={(e) => { setKeyword(e.target.value); setPage(1) }}
            className="pl-9 h-9 border-zinc-200 dark:border-zinc-700 bg-transparent shadow-none rounded-md text-sm placeholder:text-zinc-400 focus-visible:ring-1 focus-visible:ring-zinc-400 transition-shadow"
          />
        </div>
        <Select value={categoryId || 'all'} onValueChange={(v) => { setCategoryId(v === 'all' ? '' : v); setPage(1) }}>
          <SelectTrigger className="w-40 h-9 border-zinc-200 dark:border-zinc-700 shadow-none rounded-md text-sm">
            <SelectValue placeholder="分类" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部分类</SelectItem>
            {categories?.map((c) => (
              <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* 文章列表 */}
      <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow className="border-b border-zinc-100 dark:border-zinc-800 hover:bg-transparent">
              <TableHead className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">标题</TableHead>
              <TableHead className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider w-[100px]">分类</TableHead>
              <TableHead className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider w-[90px] text-center">状态</TableHead>
              <TableHead className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider w-[120px]">更新日期</TableHead>
              <TableHead className="text-[11px] font-medium text-zinc-400 uppercase tracking-wider w-[80px] text-center">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-16">
                  <Loader2 className="w-5 h-5 animate-spin text-zinc-400 mx-auto" />
                </TableCell>
              </TableRow>
            ) : data?.items && data.items.length > 0 ? (
              data.items.map((post) => {
                const cfg = statusConfig[post.status]
                return (
                  <TableRow
                    key={post.id}
                    className="border-b border-zinc-50 dark:border-zinc-800 cursor-pointer hover:bg-zinc-50/50 dark:hover:bg-zinc-800/30 transition-colors"
                    onDoubleClick={() => navigate(`/posts/${post.id}/edit`)}
                  >
                    <TableCell onClick={() => navigate(`/posts/${post.id}/edit`)}>
                      <p className="text-[13px] font-medium text-zinc-900 dark:text-zinc-100">{post.title}</p>
                      <p className="text-[11px] text-zinc-400 mt-0.5 truncate max-w-[400px] leading-relaxed">{post.summary}</p>
                    </TableCell>
                    <TableCell>
                      {post.category ? (
                        <Badge variant="outline" className="text-xs border-zinc-200 dark:border-zinc-700">{post.category.name}</Badge>
                      ) : (
                        <span className="text-xs text-zinc-400">—</span>
                      )}
                    </TableCell>
                    <TableCell className="text-center">
                      <Badge
                        variant={cfg.variant}
                        className={cn(
                          'text-[11px] font-medium rounded-full px-2 py-0.5',
                          post.status === 'published'
                            ? 'bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 border border-zinc-200 dark:border-zinc-700'
                            : ''
                        )}
                      >{cfg.label}</Badge>
                    </TableCell>
                    <TableCell className="text-[11px] text-zinc-400 tabular-nums">{formatDate(post.updatedAt)}</TableCell>
                    <TableCell>
                      <div className="flex gap-1 justify-center" onClick={(e) => e.stopPropagation()}>
                        <Tooltip delayDuration={0}>
                          <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => navigate(`/posts/${post.id}/edit`)}>
                              <Pencil className="w-3.5 h-3.5" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent side="top" className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm">编辑</TooltipContent>
                        </Tooltip>
                        <Tooltip delayDuration={0}>
                          <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon" className="w-7 h-7 text-red-500 hover:text-red-600" onClick={() => setDeleteTarget({ id: post.id, title: post.title })}>
                              <Trash2 className="w-3.5 h-3.5" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent side="top" className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm">删除</TooltipContent>
                        </Tooltip>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })
            ) : (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-16">
                  <FileText className="w-10 h-10 text-zinc-300 dark:text-zinc-700 mx-auto mb-3" />
                  <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">没有找到匹配的文章</p>
                  <p className="text-sm text-zinc-500 mt-1">尝试更改筛选条件或搜索关键词</p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* 分页 */}
      {data && data.pagination.total > 0 && (
        <div className="flex justify-between items-center mt-4">
          <p className="text-xs text-zinc-500">共 {data.pagination.total} 篇文章</p>
          {totalPages > 1 && (
            <Pagination>
              <PaginationContent>
                <PaginationItem>
                  <PaginationPrevious onClick={() => setPage(Math.max(1, page - 1))} className={page <= 1 ? 'pointer-events-none opacity-50' : 'cursor-pointer'} />
                </PaginationItem>
                {Array.from({ length: totalPages }, (_, i) => i + 1)
                  .filter(p => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
                  .map((p, i, arr) => (
                    <PaginationItem key={p}>
                      {i > 0 && arr[i - 1] !== p - 1 && <span className="px-2 text-zinc-400">…</span>}
                      <PaginationLink onClick={() => setPage(p)} isActive={p === page} className="cursor-pointer">{p}</PaginationLink>
                    </PaginationItem>
                  ))}
                <PaginationItem>
                  <PaginationNext onClick={() => setPage(Math.min(totalPages, page + 1))} className={page >= totalPages ? 'pointer-events-none opacity-50' : 'cursor-pointer'} />
                </PaginationItem>
              </PaginationContent>
            </Pagination>
          )}
        </div>
      )}

      {/* 删除确认 */}
      <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <AlertDialogContent className="rounded-lg">
          <AlertDialogHeader>
            <AlertDialogTitle className="text-zinc-950 dark:text-zinc-50">删除文章</AlertDialogTitle>
            <AlertDialogDescription className="text-zinc-500">
              {deleteTarget && `确定要删除「${deleteTarget.title}」吗？此操作不可撤销。`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel className="shadow-none border-zinc-200 dark:border-zinc-800">取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-red-500 text-white hover:bg-red-600 shadow-none">确认删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
