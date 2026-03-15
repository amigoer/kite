import { useState } from 'react'
import { Search, Check, Trash2, AlertTriangle, MessageSquare, Clock, FileText, Shield, ChevronDown, ChevronUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useComments, useCommentStats, useModerateComment } from '@/hooks/use-comments'
import type { CommentStatus } from '@/types/comment'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

/** 状态徽标 */
const statusBadge: Record<CommentStatus, { text: string; variant: 'default' | 'secondary' | 'destructive' }> = {
  approved: { text: '已通过', variant: 'default' },
  pending: { text: '待审核', variant: 'secondary' },
  spam: { text: '垃圾', variant: 'destructive' },
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 60) return `${minutes} 分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

/**
 * 评论管理页面 - 使用 shadcn Tabs / Card / Badge / Button / Input / Separator
 */
export function CommentsPage() {
  const [statusFilter, setStatusFilter] = useState<CommentStatus | 'all'>('all')
  const [keyword, setKeyword] = useState('')
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const { data: comments, isLoading } = useComments({ status: statusFilter, keyword })
  const { data: stats } = useCommentStats()
  const moderateMutation = useModerateComment()

  function handleModerate(id: string, action: 'approve' | 'spam' | 'delete') {
    if (action === 'delete' && !window.confirm('确定删除此评论？此操作不可撤销。')) return
    moderateMutation.mutate({ id, action })
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">评论管理</h1>
        <p className="mt-1 text-sm text-muted-foreground">审核和管理读者的评论</p>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="mb-6 grid grid-cols-4 gap-4">
          {[
            { label: '全部评论', value: stats.total },
            { label: '待审核', value: stats.pending },
            { label: '已通过', value: stats.approved },
            { label: '垃圾评论', value: stats.spam },
          ].map((card) => (
            <Card key={card.label}>
              <CardContent className="pt-4 pb-4">
                <p className="text-xs text-muted-foreground">{card.label}</p>
                <p className="mt-1 text-2xl font-semibold">{card.value}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 筛选栏 */}
      <div className="mb-4 flex items-center justify-between gap-4">
        <Tabs value={statusFilter} onValueChange={(v) => setStatusFilter(v as CommentStatus | 'all')}>
          <TabsList>
            <TabsTrigger value="all">
              <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
              全部 {stats?.total ?? 0}
            </TabsTrigger>
            <TabsTrigger value="pending">
              <Clock className="mr-1.5 h-3.5 w-3.5" />
              待审核 {stats?.pending ?? 0}
            </TabsTrigger>
            <TabsTrigger value="approved">
              <Check className="mr-1.5 h-3.5 w-3.5" />
              已通过 {stats?.approved ?? 0}
            </TabsTrigger>
            <TabsTrigger value="spam">
              <AlertTriangle className="mr-1.5 h-3.5 w-3.5" />
              垃圾 {stats?.spam ?? 0}
            </TabsTrigger>
          </TabsList>
        </Tabs>

        <div className="relative max-w-xs flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="搜索评论内容、作者…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className="pl-9" />
        </div>
      </div>

      {isLoading && <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">加载中…</div>}

      {!isLoading && comments?.length === 0 && (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-16 text-muted-foreground">
            <MessageSquare className="mb-3 h-8 w-8" strokeWidth={1} />
            <p className="text-sm">暂无评论</p>
          </CardContent>
        </Card>
      )}

      {/* 评论列表 */}
      {comments && comments.length > 0 && (
        <Card>
          {comments.map((comment, index) => {
            const badge = statusBadge[comment.status]
            const isExpanded = expandedId === comment.id
            return (
              <div key={comment.id}>
                {index > 0 && <Separator />}
                <div className={cn('flex items-start gap-4 px-5 py-4', comment.status === 'spam' && 'bg-destructive/5')}>
                  <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-muted text-xs font-semibold">
                    {comment.author.charAt(0).toUpperCase()}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">{comment.author}</span>
                      <Badge variant={badge.variant}>{badge.text}</Badge>
                      <span className="ml-auto flex items-center gap-1 text-xs text-muted-foreground">
                        <Clock className="h-3 w-3" strokeWidth={1.5} />
                        {timeAgo(comment.createdAt)}
                      </span>
                    </div>
                    <div className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                      <FileText className="h-3 w-3" strokeWidth={1.5} />
                      {comment.postTitle}
                    </div>
                    <p className={cn('mt-2 text-sm leading-relaxed', !isExpanded && 'line-clamp-2')}>{comment.content}</p>
                    {comment.content.length > 100 && (
                      <Button variant="link" size="sm" className="h-auto p-0 text-xs text-muted-foreground" onClick={() => setExpandedId(isExpanded ? null : comment.id)}>
                        {isExpanded ? <>收起 <ChevronUp className="ml-0.5 h-3 w-3" /></> : <>展开 <ChevronDown className="ml-0.5 h-3 w-3" /></>}
                      </Button>
                    )}
                    {isExpanded && (
                      <>
                        <Separator className="my-3" />
                        <div className="flex gap-4 text-xs text-muted-foreground">
                          <span>邮箱：{comment.email}</span>
                          <span>IP：{comment.ip}</span>
                          <span>UA：{comment.userAgent}</span>
                        </div>
                      </>
                    )}
                  </div>
                  <div className="flex shrink-0 items-center gap-1">
                    {comment.status !== 'approved' && (
                      <Button variant="outline" size="sm" disabled={moderateMutation.isPending} onClick={() => handleModerate(comment.id, 'approve')}>
                        <Check className="mr-1 h-3.5 w-3.5" />通过
                      </Button>
                    )}
                    {comment.status !== 'spam' && (
                      <Button variant="outline" size="sm" disabled={moderateMutation.isPending} onClick={() => handleModerate(comment.id, 'spam')}>
                        <Shield className="mr-1 h-3.5 w-3.5" />垃圾
                      </Button>
                    )}
                    <Button variant="outline" size="sm" className="text-destructive hover:text-destructive" disabled={moderateMutation.isPending} onClick={() => handleModerate(comment.id, 'delete')}>
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
              </div>
            )
          })}
        </Card>
      )}
    </div>
  )
}
