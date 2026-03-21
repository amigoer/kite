import { useQuery } from '@tanstack/react-query'
import { apiGet } from '@/lib/api-client'

/** 仪表盘统计数据 */
export interface DashboardStats {
  postCount: number
  categoryCount: number
  tagCount: number
  commentPending: number
  totalViews: number
  publishedCount: number
  draftCount: number
  scheduledCount: number
}

/**
 * 获取仪表盘统计数据 Hook
 * 聚合多个 API 调用获取统计概览
 */
export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: async (): Promise<DashboardStats> => {
      // 并行请求所有统计数据
      const [posts, categories, tags, commentStats, allPosts] = await Promise.all([
        apiGet<{ pagination: { total: number } }>('/admin/posts', { pageSize: 1 }),
        apiGet<{ pagination: { total: number } }>('/admin/categories', { pageSize: 1 }),
        apiGet<{ pagination: { total: number } }>('/admin/tags', { pageSize: 1 }),
        apiGet<{ pending: number }>('/admin/comments/stats'),
        apiGet<{ items: { status: string; viewCount: number }[] }>('/admin/posts', { pageSize: 100 }),
      ])

      const items = allPosts.items || []
      const totalViews = items.reduce((sum, p) => sum + (p.viewCount || 0), 0)
      const publishedCount = items.filter(p => p.status === 'published').length
      const draftCount = items.filter(p => p.status === 'draft').length
      const scheduledCount = items.filter(p => p.status === 'scheduled').length

      return {
        postCount: posts.pagination.total,
        categoryCount: categories.pagination.total,
        tagCount: tags.pagination.total,
        commentPending: commentStats.pending,
        totalViews,
        publishedCount,
        draftCount,
        scheduledCount,
      }
    },
  })
}

/** 最近文章类型 */
export interface RecentPost {
  id: string
  title: string
  status: string
  createdAt: string
  viewCount: number
  category?: { name: string } | null
}

/**
 * 获取最近文章 Hook
 */
export function useRecentPosts(limit = 5) {
  return useQuery({
    queryKey: ['dashboard', 'recentPosts', limit],
    queryFn: async () => {
      const result = await apiGet<{ items: RecentPost[] }>('/admin/posts', { pageSize: limit })
      return result.items
    },
  })
}

/** 最近评论类型 */
export interface RecentComment {
  id: string
  author: string
  content: string
  status: string
  createdAt: string
  post?: { title: string } | null
}

/**
 * 获取最近评论 Hook
 */
export function useRecentComments(limit = 5) {
  return useQuery({
    queryKey: ['dashboard', 'recentComments', limit],
    queryFn: async () => {
      const result = await apiGet<{ items: RecentComment[] }>('/admin/comments', { pageSize: limit })
      return result.items
    },
  })
}

