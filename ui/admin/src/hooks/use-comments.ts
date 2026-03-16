import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPatch, apiDelete } from '@/lib/api-client'
import type { Comment, CommentStatus, CommentStats } from '@/types/comment'

/** 查询参数 */
interface CommentQueryParams {
  status?: CommentStatus | 'all'
  keyword?: string
}

/** 后端评论列表响应 */
interface CommentListResponse {
  items: Comment[]
  pagination: { page: number; pageSize: number; total: number }
}

/**
 * 获取评论列表 Hook
 */
export function useComments(params: CommentQueryParams) {
  return useQuery({
    queryKey: ['comments', params],
    queryFn: async () => {
      const result = await apiGet<CommentListResponse>('/admin/comments', {
        pageSize: 100,
        status: params.status === 'all' ? undefined : params.status,
        keyword: params.keyword,
      })
      return result.items
    },
  })
}

/**
 * 评论审核状态统计
 */
export function useCommentStats() {
  return useQuery({
    queryKey: ['commentStats'],
    queryFn: () => apiGet<CommentStats>('/admin/comments/stats'),
  })
}

/**
 * 审核评论 Hook
 */
export function useModerateComment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: { id: string; action: 'approve' | 'spam' | 'delete' }) => {
      if (data.action === 'delete') {
        return apiDelete(`/admin/comments/${data.id}`)
      }
      const status = data.action === 'approve' ? 'approved' : 'spam'
      return apiPatch(`/admin/comments/${data.id}`, { status })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments'] })
      queryClient.invalidateQueries({ queryKey: ['commentStats'] })
    },
  })
}
