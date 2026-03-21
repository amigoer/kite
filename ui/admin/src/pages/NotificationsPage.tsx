import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Bell, CheckCheck, MessageSquare, Info, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search } from '@/components/search'
import {
  useNotifications, useUnreadCount, useMarkRead, useMarkAllRead,
} from '@/hooks/use-notifications'
import type { Notification } from '@/hooks/use-notifications'

/** 通知图标映射 */
function NotificationIcon({ type }: { type: string }) {
  if (type === 'comment') return <MessageSquare className="w-5 h-5 text-blue-500 shrink-0" />
  return <Info className="w-5 h-5 text-zinc-400 shrink-0" />
}

/** 通知类型文本 */
function typeLabel(type: string) {
  if (type === 'comment') return '评论'
  return '系统'
}

/** 相对时间 */
function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  return `${days} 天前`
}

/**
 * 通知页面
 */
export function NotificationsPage() {
  const [page, setPage] = useState(1)
  const navigate = useNavigate()
  const { data, isLoading } = useNotifications(page, 20)
  const { data: unreadCount = 0 } = useUnreadCount()
  const markRead = useMarkRead()
  const markAllRead = useMarkAllRead()

  const notifications = data?.items ?? []
  const total = data?.pagination?.total ?? 0
  const totalPages = Math.ceil(total / 20) || 1

  function handleClick(n: Notification) {
    if (!n.isRead) markRead.mutate(n.id)
    if (n.link) navigate(n.link.replace('/admin', ''))
  }

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ml-auto' />
      </Header>
      <Main>
        <div className="flex justify-between items-center mb-6">
          <div>
            <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50 tracking-tight">通知中心</h1>
            <p className="text-sm text-zinc-500 mt-1">
              共 {total} 条通知
              {unreadCount > 0 && <span className="ml-2 text-blue-500 font-medium">· {unreadCount} 条未读</span>}
            </p>
          </div>
          {unreadCount > 0 && (
            <Button
              variant="outline"
              size="sm"
              className="shadow-sm rounded-md border-zinc-200 dark:border-zinc-700"
              onClick={() => markAllRead.mutate()}
              disabled={markAllRead.isPending}
            >
              <CheckCheck className="w-4 h-4 mr-1.5" />
              全部标记已读
            </Button>
          )}
        </div>

        {/* 加载 */}
        {isLoading && (
          <div className="text-center py-16">
            <Loader2 className="w-4 h-4 animate-spin text-zinc-400 mx-auto" />
          </div>
        )}

        {/* 空状态 */}
        {!isLoading && notifications.length === 0 && (
          <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm py-16">
            <div className="flex flex-col items-center text-center">
              <Bell className="w-8 h-8 text-zinc-300 dark:text-zinc-600 mb-3" />
              <p className="text-sm text-zinc-700 dark:text-zinc-300">暂无通知</p>
              <p className="text-sm text-zinc-500 mt-1">当有新评论时会在这里收到通知</p>
            </div>
          </div>
        )}

        {/* 通知列表 */}
        {notifications.length > 0 && (
          <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm divide-y divide-zinc-100 dark:divide-zinc-800">
            {notifications.map((n) => (
              <div
                key={n.id}
                className={cn(
                  'flex gap-4 px-5 py-4 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors',
                  !n.isRead && 'bg-blue-50/40 dark:bg-blue-950/10'
                )}
                onClick={() => handleClick(n)}
              >
                <div className="mt-0.5">
                  <NotificationIcon type={n.type} />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <p className={cn('text-sm', !n.isRead ? 'font-semibold text-zinc-800 dark:text-zinc-200' : 'text-zinc-600 dark:text-zinc-400')}>
                      {n.title}
                    </p>
                    {!n.isRead && <span className="w-2 h-2 bg-blue-500 rounded-full shrink-0" />}
                  </div>
                  <p className="text-sm text-zinc-500 mt-1 line-clamp-2">{n.content}</p>
                  <div className="flex items-center gap-3 mt-2">
                    <span className="text-xs text-zinc-400">{timeAgo(n.createdAt)}</span>
                    <span className="text-xs text-zinc-300 dark:text-zinc-600">·</span>
                    <span className="text-xs text-zinc-400">{typeLabel(n.type)}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* 分页 */}
        {totalPages > 1 && (
          <div className="flex justify-center items-center gap-2 mt-6">
            <Button
              variant="outline"
              size="sm"
              className="shadow-sm rounded-md border-zinc-200 dark:border-zinc-700"
              disabled={page <= 1}
              onClick={() => setPage(p => p - 1)}
            >
              上一页
            </Button>
            <span className="text-sm text-zinc-500 px-2">
              {page} / {totalPages}
            </span>
            <Button
              variant="outline"
              size="sm"
              className="shadow-sm rounded-md border-zinc-200 dark:border-zinc-700"
              disabled={page >= totalPages}
              onClick={() => setPage(p => p + 1)}
            >
              下一页
            </Button>
          </div>
        )}
      </Main>
    </>
  )
}
