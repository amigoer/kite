import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPatch } from '@/lib/api-client'

/** 通知类型 */
export interface Notification {
  id: string
  type: string
  title: string
  content: string
  link: string
  isRead: boolean
  createdAt: string
}

interface NotificationListResponse {
  items: Notification[]
  pagination: { page: number; pageSize: number; total: number }
}

interface UnreadCountResponse {
  count: number
}

/**
 * 获取通知列表
 */
export function useNotifications(page = 1, pageSize = 20) {
  return useQuery({
    queryKey: ['notifications', page, pageSize],
    queryFn: async () => {
      const res = await apiGet<NotificationListResponse>('/admin/notifications', { page, pageSize })
      return res
    },
  })
}

/**
 * 获取未读通知数量（每 30s 轮询）
 */
export function useUnreadCount() {
  return useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: async () => {
      const res = await apiGet<UnreadCountResponse>('/admin/notifications/unread-count')
      return res.count
    },
    refetchInterval: 30000, // 每 30 秒轮询
  })
}

/**
 * 标记单条通知为已读
 */
export function useMarkRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiPatch(`/admin/notifications/${id}/read`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifications'] })
    },
  })
}

/**
 * 标记所有通知为已读
 */
export function useMarkAllRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => apiPatch('/admin/notifications/read-all'),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifications'] })
    },
  })
}
