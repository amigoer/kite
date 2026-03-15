/**
 * 文章相关类型定义
 */

/** 文章状态枚举 */
export type PostStatus = 'published' | 'draft' | 'archived'

/** 文章数据结构 */
export interface Post {
  id: string
  title: string
  slug: string
  summary: string
  category: string
  tags: string[]
  status: PostStatus
  coverUrl: string
  viewCount: number
  commentCount: number
  createdAt: string
  updatedAt: string
  publishedAt: string | null
}

/** 文章列表查询参数 */
export interface PostQueryParams {
  page: number
  pageSize: number
  keyword?: string
  status?: PostStatus | 'all'
  category?: string
}

/** 分页响应 */
export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}
