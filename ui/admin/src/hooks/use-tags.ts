import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiDelete } from '@/lib/api-client'
import type { Tag } from '@/types/tag'

/** 后端标签列表响应 */
interface TagListResponse {
  items: Tag[]
  pagination: { page: number; pageSize: number; total: number }
}

/**
 * 获取标签列表 Hook
 */
export function useTagList(keyword?: string) {
  return useQuery({
    queryKey: ['tagList', keyword],
    queryFn: async () => {
      const result = await apiGet<TagListResponse>('/admin/tags', {
        pageSize: 100,
        keyword,
      })
      return result.items
    },
  })
}

/**
 * 创建标签 Hook
 */
export function useCreateTag() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Pick<Tag, 'name' | 'slug'>) =>
      apiPost<Tag>('/admin/tags', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tagList'] })
    },
  })
}

/**
 * 删除标签 Hook
 */
export function useDeleteTag() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiDelete(`/admin/tags/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tagList'] })
    },
  })
}
