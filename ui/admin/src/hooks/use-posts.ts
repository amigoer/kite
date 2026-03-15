import { useQuery } from '@tanstack/react-query'
import { mockPosts, mockCategories } from '@/mocks/posts'
import type { Post, PostQueryParams, PaginatedData } from '@/types/post'

/**
 * 模拟 API 请求延迟
 */
function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * Mock 获取文章列表
 * 模拟后端分页、搜索、筛选逻辑
 */
async function fetchPosts(params: PostQueryParams): Promise<PaginatedData<Post>> {
  await delay(300)

  let filtered = [...mockPosts]

  // 关键词搜索
  if (params.keyword) {
    const kw = params.keyword.toLowerCase()
    filtered = filtered.filter(
      (p) =>
        p.title.toLowerCase().includes(kw) ||
        p.summary.toLowerCase().includes(kw) ||
        p.tags.some((t) => t.toLowerCase().includes(kw))
    )
  }

  // 状态筛选
  if (params.status && params.status !== 'all') {
    filtered = filtered.filter((p) => p.status === params.status)
  }

  // 分类筛选
  if (params.category) {
    filtered = filtered.filter((p) => p.category === params.category)
  }

  // 按更新时间倒序
  filtered.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())

  // 分页
  const total = filtered.length
  const start = (params.page - 1) * params.pageSize
  const list = filtered.slice(start, start + params.pageSize)

  return { list, total, page: params.page, pageSize: params.pageSize }
}

/**
 * 获取文章列表 Hook
 */
export function usePosts(params: PostQueryParams) {
  return useQuery({
    queryKey: ['posts', params],
    queryFn: () => fetchPosts(params),
  })
}

/**
 * 获取分类列表 Hook
 */
export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      await delay(100)
      return mockCategories
    },
  })
}
