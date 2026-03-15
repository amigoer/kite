import { Node, mergeAttributes } from '@tiptap/core'
import { ReactNodeViewRenderer } from '@tiptap/react'
import { CalloutView } from './CalloutView'
import type { LucideIcon } from 'lucide-react'
import { Lightbulb, CircleCheck, TriangleAlert, ShieldAlert, Star } from 'lucide-react'

/** Callout 类型定义 */
export type CalloutType = 'info' | 'warning' | 'danger' | 'tip' | 'important'

/** Callout 类型配置映射 */
export const CALLOUT_TYPES: Record<CalloutType, { label: string; icon: LucideIcon; color: string }> = {
  info:      { label: '提示',   icon: Lightbulb,     color: '#3b82f6' },
  tip:       { label: '建议',   icon: CircleCheck,    color: '#22c55e' },
  warning:   { label: '警告',   icon: TriangleAlert,  color: '#f59e0b' },
  danger:    { label: '危险',   icon: ShieldAlert,    color: '#ef4444' },
  important: { label: '重要',   icon: Star,           color: '#a855f7' },
}

// 声明自定义命令类型
declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    callout: {
      setCallout: (type?: CalloutType) => ReturnType
      toggleCallout: (type?: CalloutType) => ReturnType
    }
  }
}

/**
 * 自定义 Callout 块级节点
 * 支持自定义标题，Markdown 语法: ::: info 自定义标题
 */
export const Callout = Node.create({
  name: 'callout',

  group: 'block',

  content: 'block+',

  defining: true,

  addAttributes() {
    return {
      type: {
        default: 'info' as CalloutType,
        parseHTML: (element: HTMLElement) => element.getAttribute('data-callout') || 'info',
        renderHTML: (attributes: Record<string, string>) => ({
          'data-callout': attributes.type,
          class: `callout callout-${attributes.type}`,
        }),
      },
      /** 自定义标题，为空时使用类型默认标签 */
      title: {
        default: '',
        parseHTML: (element: HTMLElement) => element.getAttribute('data-callout-title') || '',
        renderHTML: (attributes: Record<string, string>) => {
          if (!attributes.title) return {}
          return { 'data-callout-title': attributes.title }
        },
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-callout]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return ['div', mergeAttributes(HTMLAttributes), 0]
  },

  addNodeView() {
    return ReactNodeViewRenderer(CalloutView)
  },

  addCommands() {
    return {
      setCallout: (type: CalloutType = 'info') => ({ commands }) => {
        return commands.wrapIn(this.name, { type })
      },
      toggleCallout: (type: CalloutType = 'info') => ({ commands, state }) => {
        const { selection } = state
        const node = selection.$head.node(-1)
        if (node?.type.name === this.name && node.attrs.type === type) {
          return commands.lift(this.name)
        }
        if (node?.type.name === this.name) {
          return commands.updateAttributes(this.name, { type })
        }
        return commands.wrapIn(this.name, { type })
      },
    }
  },
})
