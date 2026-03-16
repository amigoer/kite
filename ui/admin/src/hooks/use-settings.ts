import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiGet, apiPut } from '@/lib/api-client'
import type { AllSettings } from '@/types/settings'

/**
 * 获取设置 Hook
 */
export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: () => apiGet<AllSettings>('/admin/settings'),
  })
}

/**
 * 保存设置 Hook
 */
export function useSaveSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: AllSettings) => apiPut<AllSettings>('/admin/settings', data),
    onSuccess: (data) => {
      queryClient.setQueryData(['settings'], data)
    },
  })
}
