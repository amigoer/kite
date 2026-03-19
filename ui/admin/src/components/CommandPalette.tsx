import { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { useNavigate } from 'react-router'
import {
  LayoutDashboard, FileText, Files, FolderOpen, Tags,
  MessageSquare, Link2, Settings, Search, PenLine,
} from 'lucide-react'
import { useSearch } from '@/context/search-provider'
import '@/styles/command-palette.css'

/** 命令项类型 */
interface CommandItem {
  id: string
  title: string
  subtitle?: string
  icon: React.ReactNode
  group: string
  action: () => void
}

/**
 * 全局命令面板（⌘K）
 */
export function CommandPalette() {
  const { open, setOpen } = useSearch()
  const [query, setQuery] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)
  const navigate = useNavigate()

  const navCommands: CommandItem[] = useMemo(() => [
    { id: 'nav-dashboard', title: '仪表盘', icon: <LayoutDashboard className="w-4 h-4" />, group: '导航', action: () => navigate('/') },
    { id: 'nav-posts', title: '文章管理', icon: <FileText className="w-4 h-4" />, group: '导航', action: () => navigate('/posts') },
    { id: 'nav-pages', title: '页面管理', icon: <Files className="w-4 h-4" />, group: '导航', action: () => navigate('/pages') },
    { id: 'nav-categories', title: '分类管理', icon: <FolderOpen className="w-4 h-4" />, group: '导航', action: () => navigate('/categories') },
    { id: 'nav-tags', title: '标签管理', icon: <Tags className="w-4 h-4" />, group: '导航', action: () => navigate('/tags') },
    { id: 'nav-comments', title: '评论管理', icon: <MessageSquare className="w-4 h-4" />, group: '导航', action: () => navigate('/comments') },
    { id: 'nav-links', title: '友情链接', icon: <Link2 className="w-4 h-4" />, group: '导航', action: () => navigate('/links') },
    { id: 'nav-settings', title: '系统设置', icon: <Settings className="w-4 h-4" />, group: '导航', action: () => navigate('/settings') },
    { id: 'nav-new-post', title: '撰写新文章', subtitle: '创建一篇新的博客文章', icon: <PenLine className="w-4 h-4" />, group: '操作', action: () => navigate('/posts/new') },
    { id: 'nav-new-page', title: '创建新页面', subtitle: '创建一个新的独立页面', icon: <Files className="w-4 h-4" />, group: '操作', action: () => navigate('/pages/new') },
  ], [navigate])

  const allCommands = navCommands

  const filteredCommands = useMemo(() => {
    if (!query.trim()) return navCommands
    const q = query.toLowerCase()
    return allCommands.filter(
      cmd => cmd.title.toLowerCase().includes(q) ||
        cmd.subtitle?.toLowerCase().includes(q) ||
        cmd.group.toLowerCase().includes(q)
    )
  }, [query, allCommands, navCommands])

  const groupedCommands = useMemo(() => {
    const groups: Record<string, CommandItem[]> = {}
    for (const cmd of filteredCommands) {
      if (!groups[cmd.group]) groups[cmd.group] = []
      groups[cmd.group].push(cmd)
    }
    return groups
  }, [filteredCommands])

  const flatList = filteredCommands

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false)
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [setOpen])

  useEffect(() => {
    if (open) {
      setQuery('')
      setActiveIndex(0)
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open])

  useEffect(() => {
    const item = listRef.current?.querySelector(`[data-index="${activeIndex}"]`)
    item?.scrollIntoView({ block: 'nearest' })
  }, [activeIndex])

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') { e.preventDefault(); setActiveIndex(i => (i + 1) % flatList.length) }
    else if (e.key === 'ArrowUp') { e.preventDefault(); setActiveIndex(i => (i - 1 + flatList.length) % flatList.length) }
    else if (e.key === 'Enter') { e.preventDefault(); if (flatList[activeIndex]) { flatList[activeIndex].action(); setOpen(false) } }
  }, [flatList, activeIndex])

  const executeCommand = useCallback((cmd: CommandItem) => { cmd.action(); setOpen(false) }, [])

  if (!open) return null
  let globalIndex = 0

  return (
    <div className="cmd-overlay" onClick={() => setOpen(false)}>
      <div className="cmd-panel" onClick={e => e.stopPropagation()} onKeyDown={handleKeyDown}>
        <div className="cmd-input-wrapper">
          <Search className="cmd-input-icon w-4 h-4" />
          <input
            ref={inputRef}
            className="cmd-input"
            value={query}
            onChange={e => { setQuery(e.target.value); setActiveIndex(0) }}
            placeholder="搜索文章、页面或快速跳转…"
            autoComplete="off"
            spellCheck={false}
          />
          <kbd className="cmd-kbd">ESC</kbd>
        </div>
        <div className="cmd-list" ref={listRef}>
          {flatList.length === 0 && <div className="cmd-empty">没有找到匹配的结果</div>}
          {Object.entries(groupedCommands).map(([group, items]) => (
            <div key={group} className="cmd-group">
              <div className="cmd-group-label">{group}</div>
              {items.map(item => {
                const idx = globalIndex++
                return (
                  <div
                    key={item.id}
                    className={`cmd-item${idx === activeIndex ? ' active' : ''}`}
                    data-index={idx}
                    onClick={() => executeCommand(item)}
                    onMouseEnter={() => setActiveIndex(idx)}
                  >
                    <span className="cmd-item-icon">{item.icon}</span>
                    <div className="cmd-item-text">
                      <span className="cmd-item-title">{item.title}</span>
                      {item.subtitle && <span className="cmd-item-subtitle">{item.subtitle}</span>}
                    </div>
                  </div>
                )
              })}
            </div>
          ))}
        </div>
        <div className="cmd-footer">
          <span><kbd>↑</kbd><kbd>↓</kbd> 移动</span>
          <span><kbd>↵</kbd> 选择</span>
          <span><kbd>ESC</kbd> 关闭</span>
        </div>
      </div>
    </div>
  )
}
