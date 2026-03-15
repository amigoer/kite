import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { mockPosts, mockCategories } from '@/mocks/posts'
import type { Post, PostQueryParams, PaginatedData, PostDetail, PostFormData } from '@/types/post'

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

/** Mock 文章正文内容 */
const mockContent = `<h2>引言</h2>
<p>这是一篇示例文章的正文内容。在这里你可以使用<strong>加粗</strong>、<em>斜体</em>、<code>行内代码</code>等格式。</p>
<h2>代码示例</h2>
<pre><code class="language-go">package main

import "fmt"

func main() {
    fmt.Println("Hello, Kite Blog!")
}</code></pre>
<h2>列表</h2>
<ul>
  <li>第一项</li>
  <li>第二项</li>
  <li>第三项</li>
</ul>
<blockquote><p>这是一段引用文字，用于展示引用样式的效果。</p></blockquote>
<p>更多内容可以在编辑器中继续编写…</p>`

/**
 * 获取文章详情 Hook
 */
export function usePostDetail(id: string | undefined) {
  return useQuery({
    queryKey: ['post', id],
    queryFn: async (): Promise<PostDetail> => {
      await delay(300)
      const post = mockPosts.find((p) => p.id === id)
      if (!post) throw new Error('文章不存在')
      return { ...post, content: mockContent }
    },
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
      await delay(500)
      // Mock: 返回生成的 ID
      return { id: data.id || crypto.randomUUID(), ...data }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] })
    },
  })
}
