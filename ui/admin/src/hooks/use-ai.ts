import { useQueryClient, useMutation } from '@tanstack/react-query'
import { apiPost } from '@/lib/api-client'
import type { AllSettings } from '@/types/settings'

/**
 * 读取 AI 是否启用（从 settings query 缓存中获取，避免额外请求）
 */
export function useAiEnabled(): boolean {
  const queryClient = useQueryClient()
  const settings = queryClient.getQueryData<AllSettings>(['settings'])
  return settings?.ai?.enabled ?? false
}

/** AI 摘要生成请求参数 */
interface AiSummaryInput {
  content: string
  length?: number
}

/** AI 摘要生成响应 */
interface AiSummaryOutput {
  summary: string
}

/**
 * 调用后端 AI 摘要生成接口
 */
export function useAiSummary() {
  return useMutation({
    mutationFn: (input: AiSummaryInput) =>
      apiPost<AiSummaryOutput>('/admin/ai/summary', input),
  })
}

/** AI 标签推荐请求参数 */
interface AiTagsInput {
  title: string
  content: string
}

/** AI 标签推荐响应 */
interface AiTagsOutput {
  tags: string[]
}

/**
 * 调用后端 AI 标签推荐接口
 */
export function useAiTags() {
  return useMutation({
    mutationFn: (input: AiTagsInput) =>
      apiPost<AiTagsOutput>('/admin/ai/tags', input),
  })
}
