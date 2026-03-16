/**
 * 评论相关类型定义
 */

/** 评论状态 */
export type CommentStatus = 'approved' | 'pending' | 'spam'

/** 评论数据结构（与后端对齐） */
export interface Comment {
  id: string
  postId: string
  postTitle: string
  author: string
  email: string
  content: string
  status: CommentStatus
  ip: string
  userAgent: string
  createdAt: string
  updatedAt: string
}

/** 评论统计 */
export interface CommentStats {
  total: number
  approved: number
  pending: number
  spam: number
}
