import { Outlet } from 'react-router'
import { Sidebar } from '@/components/sidebar'
import { Header } from '@/components/header'
import { CommandSearch } from '@/components/CommandSearch'
import { useSidebarStore } from '@/stores/use-sidebar-store'

/**
 * 管理后台主布局 — Vercel 风格
 * 极浅灰底 + 纯白卡片的层级对比
 */
export function AdminLayout() {
  const { isCollapsed } = useSidebarStore()

  return (
    <div className="flex h-screen overflow-hidden bg-[#FAFAFA] dark:bg-zinc-950">
      <Sidebar />
      <div
        className="flex flex-col flex-1 min-w-0 transition-all duration-200"
        style={{ marginLeft: isCollapsed ? 64 : 224 }}
      >
        <Header />
        <main className="flex-1 overflow-auto px-8 py-6">
          <Outlet />
        </main>
      </div>
      <CommandSearch />
    </div>
  )
}
