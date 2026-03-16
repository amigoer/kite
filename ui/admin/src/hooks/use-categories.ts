import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiDelete } from '@/lib/api-client'
import type { Category } from '@/types/category'

/** 后端分类列表响应 */
interface CategoryListResponse {
  items: Category[]
  pagination: { page: number; pageSize: number; total: number }
}

/**
 * 获取分类列表 Hook
 */
export function useCategoryList(keyword?: string) {
  return useQuery({
    queryKey: ['categoryList', keyword],
    queryFn: async () => {
      const result = await apiGet<CategoryListResponse>('/admin/categories', {
        pageSize: 100,
        keyword,
      })
      return result.items
    },
  })
}

/**
 * 创建分类 Hook
 */
export function useCreateCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Pick<Category, 'name' | 'slug' | 'description'>) =>
      apiPost<Category>('/admin/categories', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categoryList'] })
    },
  })
}

/**
 * 删除分类 Hook
 */
export function useDeleteCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiDelete(`/admin/categories/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categoryList'] })
    },
  })
}
