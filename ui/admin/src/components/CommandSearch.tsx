import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router'
import {
  CommandDialog, Command, CommandInput, CommandList,
  CommandEmpty, CommandGroup, CommandItem, CommandSeparator,
} from '@/components/ui/command'
import {
  LayoutDashboard, FileText, Files, FolderOpen, Tags,
  MessageSquare, Link2, Settings, Plus, Search,
} from 'lucide-react'

/** 导航项 */
const navItems = [
  { label: '仪表盘', path: '/', icon: LayoutDashboard, group: '导航' },
  { label: '文章管理', path: '/posts', icon: FileText, group: '导航' },
  { label: '页面管理', path: '/pages', icon: Files, group: '导航' },
  { label: '分类管理', path: '/categories', icon: FolderOpen, group: '导航' },
  { label: '标签管理', path: '/tags', icon: Tags, group: '导航' },
  { label: '评论管理', path: '/comments', icon: MessageSquare, group: '导航' },
  { label: '友链管理', path: '/links', icon: Link2, group: '导航' },
  { label: '系统设置', path: '/settings', icon: Settings, group: '导航' },
]

/** 快捷操作 */
const quickActions = [
  { label: '写文章', path: '/posts/new', icon: Plus, group: '操作' },
  { label: '新建页面', path: '/pages/new', icon: Plus, group: '操作' },
]

/**
 * ⌘K 搜索命令面板 — Vercel 风格
 * 居中弹窗，支持键盘导航
 */
export function CommandSearch() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()

  // 监听 ⌘K / Ctrl+K 快捷键
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setOpen((prev) => !prev)
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [])

  const handleSelect = useCallback((path: string) => {
    setOpen(false)
    navigate(path)
  }, [navigate])

  return (
    <CommandDialog
      open={open}
      onOpenChange={setOpen}
      title="搜索"
      description="搜索页面和快捷操作"
    >
      <Command className="rounded-lg border-none">
        <CommandInput placeholder="搜索页面、操作…" />
        <CommandList>
          <CommandEmpty>
            <div className="flex flex-col items-center gap-1.5 py-4">
              <Search className="w-5 h-5 text-zinc-300" />
              <span className="text-sm text-zinc-500">没有找到结果</span>
            </div>
          </CommandEmpty>

          <CommandGroup heading="快捷操作">
            {quickActions.map((item) => (
              <CommandItem
                key={item.path}
                onSelect={() => handleSelect(item.path)}
                className="gap-2.5 cursor-pointer"
              >
                <item.icon className="w-4 h-4 text-zinc-400" />
                <span>{item.label}</span>
              </CommandItem>
            ))}
          </CommandGroup>

          <CommandSeparator />

          <CommandGroup heading="页面导航">
            {navItems.map((item) => (
              <CommandItem
                key={item.path}
                onSelect={() => handleSelect(item.path)}
                className="gap-2.5 cursor-pointer"
              >
                <item.icon className="w-4 h-4 text-zinc-400" />
                <span>{item.label}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </Command>
    </CommandDialog>
  )
}
