import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPatch, apiDelete } from '@/lib/api-client'
import type { FriendLink } from '@/types/friend-link'

/** 后端友链列表响应 */
interface FriendLinkListResponse {
  items: FriendLink[]
  pagination: { page: number; pageSize: number; total: number }
}

/**
 * 获取友链列表 Hook
 */
export function useFriendLinks(keyword?: string) {
  return useQuery({
    queryKey: ['friendLinks', keyword],
    queryFn: async () => {
      const result = await apiGet<FriendLinkListResponse>('/admin/friend-links', {
        pageSize: 100,
        keyword,
      })
      return result.items
    },
  })
}

/**
 * 创建友链 Hook
 */
export function useCreateFriendLink() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Pick<FriendLink, 'name' | 'url' | 'description'>) =>
      apiPost<FriendLink>('/admin/friend-links', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['friendLinks'] })
    },
  })
}

/**
 * 删除友链 Hook
 */
export function useDeleteFriendLink() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiDelete(`/admin/friend-links/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['friendLinks'] })
    },
  })
}

/**
 * 切换友链状态 Hook
 */
export function useToggleLinkStatus() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: { id: string; status: FriendLink['status'] }) =>
      apiPatch<FriendLink>(`/admin/friend-links/${data.id}`, { status: data.status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['friendLinks'] })
    },
  })
}
