import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api-client'
import type { Post, PostQueryParams, PaginatedData, PostDetail, PostFormData } from '@/types/post'

/**
 * 获取文章列表 Hook
 */
export function usePosts(params: PostQueryParams) {
  return useQuery({
    queryKey: ['posts', params],
    queryFn: () =>
      apiGet<PaginatedData<Post>>('/admin/posts', {
        page: params.page,
        pageSize: params.pageSize,
        keyword: params.keyword,
        status: params.status === 'all' ? undefined : params.status,
        categoryId: params.categoryId,
        tagId: params.tagId,
      }),
  })
}

/**
 * 获取分类列表 Hook（用于文章筛选下拉）
 */
export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const result = await apiGet<{ items: { id: string; name: string; slug: string }[] }>('/admin/categories', { pageSize: 100 })
      return result.items.map((c) => c.name)
    },
  })
}

/**
 * 获取文章详情 Hook
 */
export function usePostDetail(id: string | undefined) {
  return useQuery<PostDetail>({
    queryKey: ['post', id],
    queryFn: () => apiGet<PostDetail>(`/admin/posts/${id}`),
    enabled: !!id,
  })
}

/**
 * 保存文章 Hook（新建 / 更新）
 */
export function useSavePost() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: PostFormData & { id?: string }) => {
      const payload = {
        ...data,
        published_at: data.publishAt ? new Date(data.publishAt).toISOString() : undefined,
      }
      if (data.id) {
        return apiPut<Post>(`/admin/posts/${data.id}`, payload)
      }
      return apiPost<Post>('/admin/posts', payload)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] })
    },
  })
}

/**
 * 删除文章 Hook
 */
export function useDeletePost() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiDelete(`/admin/posts/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] })
    },
  })
}
