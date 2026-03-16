/**
 * 独立页面相关类型定义
 */

/** 页面状态枚举 */
export type PageStatus = 'published' | 'draft'

/** 独立页面数据结构（与后端对齐） */
export interface Page {
  id: string
  title: string
  slug: string
  status: PageStatus
  sortOrder: number
  showInNav: boolean
  template: string
  config: string
  createdAt: string
  updatedAt: string
  publishedAt: string | null
}

/** 独立页面详情（含正文） */
export interface PageDetail extends Page {
  contentMarkdown: string
  contentHtml: string
}

/** 独立页面表单数据 */
export interface PageFormData {
  title: string
  slug: string
  contentMarkdown: string
  status: PageStatus
  sortOrder: number
  showInNav: boolean
  template: string
  config: string
}
