/**
 * 友链相关类型定义
 */

/** 友链状态 */
export type LinkStatus = 'active' | 'pending' | 'down'

/** 友链数据结构（与后端对齐） */
export interface FriendLink {
  id: string
  name: string
  url: string
  logo: string
  description: string
  status: LinkStatus
  sort: number
  createdAt: string
  updatedAt: string
}
