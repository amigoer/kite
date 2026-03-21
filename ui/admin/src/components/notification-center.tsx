import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Bell, CheckCheck, MessageSquare, Info } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Popover, PopoverContent, PopoverTrigger,
} from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import {
  useNotifications, useUnreadCount, useMarkRead, useMarkAllRead,
} from '@/hooks/use-notifications'
import type { Notification } from '@/hooks/use-notifications'

/** 通知图标映射 */
function NotificationIcon({ type }: { type: string }) {
  if (type === 'comment') return <MessageSquare className="w-4 h-4 text-blue-500 shrink-0" />
  return <Info className="w-4 h-4 text-zinc-400 shrink-0" />
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

export function NotificationCenter() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()
  const { data: unreadCount = 0 } = useUnreadCount()
  const { data } = useNotifications(1, 30)
  const markRead = useMarkRead()
  const markAllRead = useMarkAllRead()

  const notifications = data?.items ?? []

  function handleClick(n: Notification) {
    if (!n.isRead) markRead.mutate(n.id)
    if (n.link) {
      setOpen(false)
      navigate(n.link.replace('/admin', ''))
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative rounded-md">
          <Bell className="w-4 h-4" />
          {unreadCount > 0 && (
            <span className="absolute -top-0.5 -right-0.5 w-4 h-4 bg-red-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center">
              {unreadCount > 9 ? '9+' : unreadCount}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-80 p-0">
        {/* 头部 */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-zinc-100 dark:border-zinc-800">
          <span className="text-sm font-semibold text-zinc-800 dark:text-zinc-200">通知</span>
          {unreadCount > 0 && (
            <Button
              variant="ghost"
              size="sm"
              className="h-7 text-xs text-zinc-500 hover:text-zinc-700"
              onClick={() => markAllRead.mutate()}
              disabled={markAllRead.isPending}
            >
              <CheckCheck className="w-3.5 h-3.5 mr-1" />
              全部已读
            </Button>
          )}
        </div>
        {/* 列表 */}
        <div className="max-h-80 overflow-y-auto">
          {notifications.length === 0 ? (
            <div className="py-10 text-center">
              <Bell className="w-8 h-8 text-zinc-200 dark:text-zinc-700 mx-auto mb-2" />
              <p className="text-sm text-zinc-400">暂无通知</p>
            </div>
          ) : (
            notifications.map((n) => (
              <div
                key={n.id}
                className={cn(
                  'flex gap-3 px-4 py-3 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors border-b border-zinc-50 dark:border-zinc-800/50 last:border-0',
                  !n.isRead && 'bg-blue-50/50 dark:bg-blue-950/10'
                )}
                onClick={() => handleClick(n)}
              >
                <div className="mt-0.5">
                  <NotificationIcon type={n.type} />
                </div>
                <div className="min-w-0 flex-1">
                  <p className={cn('text-sm truncate', !n.isRead ? 'font-medium text-zinc-800 dark:text-zinc-200' : 'text-zinc-600 dark:text-zinc-400')}>
                    {n.title}
                  </p>
                  <p className="text-xs text-zinc-400 truncate mt-0.5">{n.content}</p>
                  <p className="text-xs text-zinc-400 mt-1">{timeAgo(n.createdAt)}</p>
                </div>
                {!n.isRead && (
                  <span className="w-2 h-2 bg-blue-500 rounded-full shrink-0 mt-1.5" />
                )}
              </div>
            ))
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}
