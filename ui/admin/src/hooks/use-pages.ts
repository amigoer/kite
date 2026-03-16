import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api-client'
import type { Page, PageDetail, PageFormData } from '@/types/page'

/** 后端页面列表响应 */
interface PageListResponse {
  items: Page[]
  pagination: { page: number; pageSize: number; total: number }
}

/**
 * 获取独立页面列表 Hook
 */
export function usePageList(keyword?: string) {
  return useQuery<Page[]>({
    queryKey: ['pages', keyword],
    queryFn: async () => {
      const result = await apiGet<PageListResponse>('/admin/pages', {
        pageSize: 100,
        keyword,
      })
      return result.items
    },
  })
}

/**
 * 获取独立页面详情 Hook
 */
export function usePageDetail(id: string | undefined) {
  return useQuery<PageDetail>({
    queryKey: ['page', id],
    queryFn: () => apiGet<PageDetail>(`/admin/pages/${id}`),
    enabled: !!id,
  })
}

/**
 * 保存独立页面 Hook（新建 / 更新）
 */
export function useSavePage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: PageFormData & { id?: string }) => {
      if (data.id) {
        return apiPut<Page>(`/admin/pages/${data.id}`, data)
      }
      return apiPost<Page>('/admin/pages', data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pages'] })
    },
  })
}

/**
 * 删除独立页面 Hook
 */
export function useDeletePage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiDelete(`/admin/pages/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pages'] })
    },
  })
}
