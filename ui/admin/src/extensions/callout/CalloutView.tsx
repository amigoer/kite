import { NodeViewWrapper, NodeViewContent } from '@tiptap/react'
import type { NodeViewProps } from '@tiptap/react'
import { useState, useRef, useEffect } from 'react'
import { CALLOUT_TYPES, type CalloutType } from './callout-extension'

/**
 * Callout React NodeView
 * 紧凑 blockquote 风格，支持自定义标题
 */
export function CalloutView({ node, updateAttributes }: NodeViewProps) {
  const type = (node.attrs.type || 'info') as CalloutType
  const title = node.attrs.title || ''
  const config = CALLOUT_TYPES[type]
  const [menuOpen, setMenuOpen] = useState(false)
  const [editing, setEditing] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const IconComponent = config.icon
  /** 显示的标题：自定义标题 > 默认类型标签 */
  const displayTitle = title || config.label

  useEffect(() => {
    if (!menuOpen) return
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [menuOpen])

  // 自动聚焦编辑输入框
  useEffect(() => {
    if (editing && inputRef.current) {
      inputRef.current.focus()
      inputRef.current.select()
    }
  }, [editing])

  return (
    <NodeViewWrapper className={`callout callout-${type}`} data-callout={type}>
      {/* 顶部标题行 */}
      <div className="callout-header" contentEditable={false}>
        {/* 类型图标按钮（点击切换类型） */}
        <button
          className="callout-type-trigger"
          onClick={() => setMenuOpen(!menuOpen)}
          type="button"
          style={{ color: config.color }}
        >
          <IconComponent size={14} />
        </button>

        {/* 标题文字（点击进入编辑） */}
        {editing ? (
          <input
            ref={inputRef}
            className="callout-title-input"
            style={{ color: config.color }}
            value={title}
            placeholder={config.label}
            onChange={(e) => updateAttributes({ title: e.target.value })}
            onBlur={() => setEditing(false)}
            onKeyDown={(e) => { if (e.key === 'Enter') setEditing(false) }}
          />
        ) : (
          <span
            className="callout-title-text"
            style={{ color: config.color }}
            onDoubleClick={() => setEditing(true)}
            title="双击编辑标题"
          >
            {displayTitle}
          </span>
        )}

        {/* 类型切换下拉 */}
        {menuOpen && (
          <div className="callout-type-menu" ref={menuRef}>
            {(Object.entries(CALLOUT_TYPES) as [CalloutType, typeof config][]).map(([key, val]) => {
              const ItemIcon = val.icon
              return (
                <button
                  key={key}
                  className={`callout-type-option${key === type ? ' active' : ''}`}
                  onClick={() => { updateAttributes({ type: key }); setMenuOpen(false) }}
                  type="button"
                >
                  <ItemIcon size={14} style={{ color: val.color }} />
                  <span>{val.label}</span>
                </button>
              )
            })}
          </div>
        )}
      </div>

      {/* 内容区域 */}
      <NodeViewContent className="callout-body" />
    </NodeViewWrapper>
  )
}
