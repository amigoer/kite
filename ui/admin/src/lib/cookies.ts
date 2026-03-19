/**
 * Cookie 工具函数
 * 用于主题/布局偏好持久化
 */

const DEFAULT_MAX_AGE = 60 * 60 * 24 * 7 // 7 天

/** 获取 cookie */
export function getCookie(name: string): string | undefined {
  if (typeof document === 'undefined') return undefined
  const value = `; ${document.cookie}`
  const parts = value.split(`; ${name}=`)
  if (parts.length === 2) {
    const cookieValue = parts.pop()?.split(';').shift()
    return cookieValue
  }
  return undefined
}

/** 设置 cookie */
export function setCookie(
  name: string,
  value: string,
  maxAge: number = DEFAULT_MAX_AGE
): void {
  if (typeof document === 'undefined') return
  document.cookie = `${name}=${value}; path=/; max-age=${maxAge}`
}

/** 删除 cookie */
export function removeCookie(name: string): void {
  if (typeof document === 'undefined') return
  document.cookie = `${name}=; path=/; max-age=0`
}
