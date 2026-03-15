import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface ThemeStore {
  /** 当前是否为暗色模式 */
  isDark: boolean
  /** 切换主题 */
  toggle: () => void
}

/**
 * 主题状态管理
 * 通过 body[theme-mode] 属性控制 Semi Design 暗色模式
 */
export const useThemeStore = create<ThemeStore>()(
  persist(
    (set, get) => ({
      isDark: false,
      toggle: () => {
        const nextDark = !get().isDark
        applyTheme(nextDark)
        set({ isDark: nextDark })
      },
    }),
    {
      name: 'kite-theme',
      onRehydrateStorage: () => (state) => {
        // 恢复时同步 DOM 属性
        if (state) applyTheme(state.isDark)
      },
    }
  )
)

/** 将主题状态同步到 DOM */
function applyTheme(isDark: boolean) {
  const body = document.body
  if (isDark) {
    body.setAttribute('theme-mode', 'dark')
  } else {
    body.removeAttribute('theme-mode')
  }
}
