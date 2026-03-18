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
 * 通过 html.dark class 控制暗色模式（Tailwind 标准方式）
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
        if (state) applyTheme(state.isDark)
      },
    }
  )
)

/** 将主题状态同步到 DOM */
function applyTheme(isDark: boolean) {
  const root = document.documentElement
  if (isDark) {
    root.classList.add('dark')
  } else {
    root.classList.remove('dark')
  }
}
