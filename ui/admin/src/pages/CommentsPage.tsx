import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Search, Check, Trash2, Shield, Clock, FileText, MessageSquare, Loader2 } from 'lucide-react'
import { useComments, useCommentStats, useModerateComment } from '@/hooks/use-comments'
import type { CommentStatus } from '@/types/comment'
import { cn } from '@/lib/utils'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search as SearchBtn } from '@/components/search'

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
 * 评论管理页面
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

  const tabItems = [
    { key: 'all' as const, label: '全部', count: stats?.total ?? 0, icon: MessageSquare },
    { key: 'pending' as const, label: '待审核', count: stats?.pending ?? 0, icon: Clock },
    { key: 'approved' as const, label: '已通过', count: stats?.approved ?? 0, icon: Check },
    { key: 'spam' as const, label: '垃圾', count: stats?.spam ?? 0, icon: Shield },
  ]

  return (
    <>
      <Header fixed>
        <SearchBtn />
        <div className='ml-auto' />
      </Header>
      <Main>
      <div className="mb-6">
        <h1 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50">评论管理</h1>
        <p className="text-sm text-zinc-500 mt-1">审核和管理读者的评论</p>
      </div>

      {/* 统计 */}
      {stats && (
        <div className="grid grid-cols-4 gap-3 mb-6">
          {[
            { label: '全部评论', value: stats.total },
            { label: '待审核', value: stats.pending },
            { label: '已通过', value: stats.approved },
            { label: '垃圾评论', value: stats.spam },
          ].map((s) => (
            <Card key={s.label} className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-none rounded-md">
              <CardContent className="p-4">
                <p className="text-xs text-zinc-500">{s.label}</p>
                <p className="text-2xl font-semibold text-zinc-950 dark:text-zinc-50 mt-1">{s.value}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 筛选 */}
      <div className="flex justify-between items-center mb-4">
        <div className="flex gap-1.5">
          {tabItems.map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.key}
                className={cn(
                  'flex items-center gap-1.5 px-3 py-1 text-sm rounded-md transition-colors cursor-pointer',
                  statusFilter === tab.key
                    ? 'bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950'
                    : 'text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800'
                )}
                onClick={() => setStatusFilter(tab.key)}
              >
                <Icon className="w-3.5 h-3.5" /> {tab.label} {tab.count}
              </button>
            )
          })}
        </div>
        <div className="relative max-w-[280px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
          <Input placeholder="搜索评论内容、作者…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className="pl-9 border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none" />
        </div>
      </div>

      {isLoading && <div className="text-center py-16"><Loader2 className="w-5 h-5 animate-spin text-zinc-400 mx-auto" /></div>}

      {!isLoading && comments?.length === 0 && (
        <div className="border border-zinc-200 dark:border-zinc-800 rounded-md bg-white dark:bg-zinc-900 py-16">
          <div className="flex flex-col items-center text-center">
            <MessageSquare className="w-10 h-10 text-zinc-300 dark:text-zinc-700 mb-3" />
            <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">暂无评论</p>
            <p className="text-sm text-zinc-500 mt-1">当读者留下评论后会在这里显示</p>
          </div>
        </div>
      )}

      {/* 评论列表 */}
      {comments && comments.length > 0 && (
        <div className="border border-zinc-200 dark:border-zinc-800 rounded-md bg-white dark:bg-zinc-900">
          {comments.map((comment, index) => {
            const badge = statusBadge[comment.status]
            const isExpanded = expandedId === comment.id
            return (
              <div key={comment.id}>
                {index > 0 && <Separator className="bg-zinc-100 dark:bg-zinc-800" />}
                <div className={cn('flex gap-3 p-4', comment.status === 'spam' && 'bg-red-50/50 dark:bg-red-950/10')}>
                  {/* 头像 */}
                  <div className="w-8 h-8 rounded-md bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center text-xs font-medium text-zinc-600 dark:text-zinc-400 shrink-0">
                    {comment.author.charAt(0).toUpperCase()}
                  </div>
                  {/* 内容 */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-zinc-950 dark:text-zinc-50">{comment.author}</span>
                      <Badge variant={badge.variant} className="text-[10px] h-4">{badge.text}</Badge>
                      <span className="text-xs text-zinc-500 ml-auto flex items-center gap-1">
                        <Clock className="w-3 h-3" /> {timeAgo(comment.createdAt)}
                      </span>
                    </div>
                    <p className="text-xs text-zinc-500 flex items-center gap-1 mt-1">
                      <FileText className="w-3 h-3" /> {comment.postTitle}
                    </p>
                    <p className={cn('text-sm text-zinc-700 dark:text-zinc-300 mt-2', !isExpanded && 'line-clamp-2')}>
                      {comment.content}
                    </p>
                    {!isExpanded && comment.content.length > 120 && (
                      <button className="text-xs text-zinc-500 hover:text-zinc-700 mt-1 cursor-pointer" onClick={() => setExpandedId(comment.id)}>展开</button>
                    )}
                    {isExpanded && (
                      <>
                        <button className="text-xs text-zinc-500 hover:text-zinc-700 mt-1 cursor-pointer" onClick={() => setExpandedId(null)}>收起</button>
                        <Separator className="my-2 bg-zinc-100 dark:bg-zinc-800" />
                        <p className="text-xs text-zinc-500">邮箱：{comment.email} · IP：{comment.ip} · UA：{comment.userAgent}</p>
                      </>
                    )}
                  </div>
                  {/* 操作 */}
                  <div className="flex gap-1 items-start shrink-0">
                    {comment.status !== 'approved' && (
                      <Button variant="outline" size="sm" className="h-7 text-xs shadow-none border-zinc-200 dark:border-zinc-800" disabled={moderateMutation.isPending} onClick={() => handleModerate(comment.id, 'approve')}>
                        <Check className="w-3 h-3 mr-1" /> 通过
                      </Button>
                    )}
                    {comment.status !== 'spam' && (
                      <Button variant="outline" size="sm" className="h-7 text-xs shadow-none border-zinc-200 dark:border-zinc-800" disabled={moderateMutation.isPending} onClick={() => handleModerate(comment.id, 'spam')}>
                        <Shield className="w-3 h-3 mr-1" /> 垃圾
                      </Button>
                    )}
                    <Button variant="ghost" size="icon" className="w-7 h-7 text-red-500 hover:text-red-600" disabled={moderateMutation.isPending} onClick={() => handleModerate(comment.id, 'delete')}>
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </Main>
    </>
  )
}
