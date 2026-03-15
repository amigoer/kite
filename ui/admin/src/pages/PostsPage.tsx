import { useState } from 'react'
import { useNavigate } from 'react-router'
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
import { usePosts, useCategories } from '@/hooks/use-posts'
import type { PostStatus } from '@/types/post'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card } from '@/components/ui/card'

/** 状态标签配置 */
const statusConfig: Record<PostStatus, { label: string; variant: 'default' | 'secondary' | 'outline' }> = {
  published: { label: '已发布', variant: 'default' },
  draft: { label: '草稿', variant: 'secondary' },
  archived: { label: '已归档', variant: 'outline' },
}

/** 格式化日期 */
function formatDate(dateStr: string | null) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
  })
}

/**
 * 文章管理页面
 * 使用 shadcn Table / Input / Select / Badge / Button / Card
 */
export function PostsPage() {
  const navigate = useNavigate()
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
          <h1 className="text-2xl font-semibold tracking-tight">文章管理</h1>
          <p className="mt-1 text-sm text-muted-foreground">管理你的所有博客文章</p>
        </div>
        <Button onClick={() => navigate('/posts/new')}>
          <Plus className="mr-2 h-4 w-4" />
          新建文章
        </Button>
      </div>

      {/* 搜索与筛选栏 */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <div className="relative flex-1 min-w-[240px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="搜索文章标题、摘要或标签…"
            value={keyword}
            onChange={(e) => { setKeyword(e.target.value); setPage(1) }}
            className="pl-9"
          />
        </div>

        <Select value={status} onValueChange={(v) => { setStatus(v as PostStatus | 'all'); setPage(1) }}>
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder="全部状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="published">已发布</SelectItem>
            <SelectItem value="draft">草稿</SelectItem>
            <SelectItem value="archived">已归档</SelectItem>
          </SelectContent>
        </Select>

        <Select value={category || 'all'} onValueChange={(v) => { setCategory(v === 'all' ? '' : v); setPage(1) }}>
          <SelectTrigger className="w-[130px]">
            <SelectValue placeholder="全部分类" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部分类</SelectItem>
            {categories?.map((c) => (
              <SelectItem key={c} value={c}>{c}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* 文章列表 */}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>标题</TableHead>
              <TableHead className="w-[100px]">分类</TableHead>
              <TableHead className="w-[80px]">状态</TableHead>
              <TableHead className="w-[80px] text-center">浏览</TableHead>
              <TableHead className="w-[80px] text-center">评论</TableHead>
              <TableHead className="w-[100px]">更新日期</TableHead>
              <TableHead className="w-[90px] text-center">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-16 text-muted-foreground">
                  加载中…
                </TableCell>
              </TableRow>
            )}

            {!isLoading && data?.list.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-16 text-muted-foreground">
                  没有找到匹配的文章
                </TableCell>
              </TableRow>
            )}

            {data?.list.map((post) => {
              const statusInfo = statusConfig[post.status]
              return (
                <TableRow key={post.id}>
                  <TableCell>
                    <p className="truncate font-medium">{post.title}</p>
                    <p className="mt-0.5 truncate text-xs text-muted-foreground">{post.summary}</p>
                  </TableCell>
                  <TableCell>{post.category}</TableCell>
                  <TableCell>
                    <Badge variant={statusInfo.variant}>{statusInfo.label}</Badge>
                  </TableCell>
                  <TableCell className="text-center">
                    <div className="flex items-center justify-center gap-1 text-muted-foreground">
                      <Eye className="h-3.5 w-3.5" strokeWidth={1.5} />
                      <span className="text-xs">{post.viewCount}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-center">
                    <div className="flex items-center justify-center gap-1 text-muted-foreground">
                      <MessageSquare className="h-3.5 w-3.5" strokeWidth={1.5} />
                      <span className="text-xs">{post.commentCount}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {formatDate(post.updatedAt)}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center justify-center gap-1">
                      <Button variant="ghost" size="icon" className="h-7 w-7" title="编辑" onClick={() => navigate(`/posts/${post.id}/edit`)}>
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive hover:text-destructive" title="删除">
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </Card>

      {/* 分页 */}
      {data && data.total > 0 && (
        <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
          <span>共 {data.total} 篇文章，第 {data.page}/{totalPages} 页</span>
          <div className="flex items-center gap-1">
            <Button variant="outline" size="icon" className="h-8 w-8" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((n) => (
              <Button key={n} variant={n === page ? 'default' : 'outline'} size="sm" className="h-8 w-8 p-0" onClick={() => setPage(n)}>
                {n}
              </Button>
            ))}
            <Button variant="outline" size="icon" className="h-8 w-8" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
