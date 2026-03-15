import { useCallback, useState, useRef, useEffect } from 'react'
import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import TiptapImage from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { common, createLowlight } from 'lowlight'
import { Divider, Modal, Input } from '@douyinfe/semi-ui'
import TurndownService from 'turndown'
import { marked } from 'marked'
import {
  Bold, Italic, Strikethrough, Code, List, ListOrdered,
  Quote, Minus, Heading1, Heading2, Heading3,
  Link as LinkIcon, Image as ImageIcon, Undo2, Redo2, FileCode,
  CodeXml,
} from 'lucide-react'
import '@/styles/tiptap.css'

/** 代码高亮引擎 */
const lowlight = createLowlight(common)

/** HTML → Markdown 转换器 */
const turndown = new TurndownService({
  headingStyle: 'atx',        // # 风格标题
  codeBlockStyle: 'fenced',   // ``` 风格代码块
  bulletListMarker: '-',      // - 风格无序列表
  emDelimiter: '*',
  strongDelimiter: '**',
  hr: '---',
})

// 保留删除线语法
turndown.addRule('strikethrough', {
  filter: ['del', 's'],
  replacement: (content) => `~~${content}~~`,
})

interface TiptapEditorProps {
  content?: string
  onChange?: (html: string) => void
  placeholder?: string
}

/**
 * Tiptap 富文本编辑器
 * - 工具栏按钮使用原生 button + CSS tooltip
 * - 图片/链接插入使用 Semi Modal
 * - 支持 Markdown 源码模式切换
 */
export function TiptapEditor({ content = '', onChange, placeholder = '开始写作…' }: TiptapEditorProps) {
  const [imageModalVisible, setImageModalVisible] = useState(false)
  const [imageUrl, setImageUrl] = useState('')
  const [linkModalVisible, setLinkModalVisible] = useState(false)
  const [linkUrl, setLinkUrl] = useState('')
  const [sourceMode, setSourceMode] = useState(false)
  const [sourceCode, setSourceCode] = useState('')
  const sourceRef = useRef<HTMLTextAreaElement>(null)

  const editor = useEditor({
    extensions: [
      StarterKit.configure({ codeBlock: false, link: false }),
      Placeholder.configure({ placeholder }),
      TiptapImage.configure({ inline: false, allowBase64: true }),
      Link.configure({ openOnClick: false, autolink: true }),
      CodeBlockLowlight.configure({ lowlight }),
    ],
    content,
    onUpdate: ({ editor }) => { onChange?.(editor.getHTML()) },
  })

  /** 切换 Markdown 源码模式 */
  const toggleSourceMode = useCallback(() => {
    if (!editor) return
    if (!sourceMode) {
      // 进入源码模式：HTML → Markdown
      const html = editor.getHTML()
      const markdown = turndown.turndown(html)
      setSourceCode(markdown)
    } else {
      // 退出源码模式：Markdown → HTML → 写回编辑器
      const html = marked.parse(sourceCode, { async: false }) as string
      editor.commands.setContent(html, { emitUpdate: true })
      onChange?.(html)
    }
    setSourceMode(!sourceMode)
  }, [editor, sourceMode, sourceCode, onChange])

  /** 源码模式下自动聚焦 textarea */
  useEffect(() => {
    if (sourceMode && sourceRef.current) {
      sourceRef.current.focus()
    }
  }, [sourceMode])

  /** 打开链接对话框 */
  const openLinkModal = useCallback(() => {
    if (!editor) return
    const previousUrl = editor.getAttributes('link').href || ''
    setLinkUrl(previousUrl)
    setLinkModalVisible(true)
  }, [editor])

  /** 确认设置链接 */
  const confirmLink = useCallback(() => {
    if (!editor) return
    if (linkUrl === '') {
      editor.chain().focus().extendMarkRange('link').unsetLink().run()
    } else {
      editor.chain().focus().extendMarkRange('link').setLink({ href: linkUrl }).run()
    }
    setLinkModalVisible(false)
    setLinkUrl('')
  }, [editor, linkUrl])

  /** 打开图片对话框 */
  const openImageModal = useCallback(() => {
    setImageUrl('')
    setImageModalVisible(true)
  }, [])

  /** 确认插入图片 */
  const confirmImage = useCallback(() => {
    if (!editor || !imageUrl.trim()) return
    editor.chain().focus().setImage({ src: imageUrl.trim() }).run()
    setImageModalVisible(false)
    setImageUrl('')
  }, [editor, imageUrl])

  if (!editor) return null

  return (
    <div className="tiptap-wrapper">
      {/* 工具栏 */}
      <div className="tiptap-toolbar">
        {/* 源码模式下隐藏格式化按钮 */}
        {!sourceMode && (
          <>
            <ToolBtn icon={Undo2} tooltip="撤销" onClick={() => editor.chain().focus().undo().run()} disabled={!editor.can().undo()} />
            <ToolBtn icon={Redo2} tooltip="重做" onClick={() => editor.chain().focus().redo().run()} disabled={!editor.can().redo()} />

            <Divider layout="vertical" style={{ margin: '0 4px', height: 20 }} />

            <ToolBtn icon={Heading1} tooltip="标题 1" onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()} active={editor.isActive('heading', { level: 1 })} />
            <ToolBtn icon={Heading2} tooltip="标题 2" onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()} active={editor.isActive('heading', { level: 2 })} />
            <ToolBtn icon={Heading3} tooltip="标题 3" onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()} active={editor.isActive('heading', { level: 3 })} />

            <Divider layout="vertical" style={{ margin: '0 4px', height: 20 }} />

            <ToolBtn icon={Bold} tooltip="加粗" onClick={() => editor.chain().focus().toggleBold().run()} active={editor.isActive('bold')} />
            <ToolBtn icon={Italic} tooltip="斜体" onClick={() => editor.chain().focus().toggleItalic().run()} active={editor.isActive('italic')} />
            <ToolBtn icon={Strikethrough} tooltip="删除线" onClick={() => editor.chain().focus().toggleStrike().run()} active={editor.isActive('strike')} />
            <ToolBtn icon={Code} tooltip="行内代码" onClick={() => editor.chain().focus().toggleCode().run()} active={editor.isActive('code')} />

            <Divider layout="vertical" style={{ margin: '0 4px', height: 20 }} />

            <ToolBtn icon={List} tooltip="无序列表" onClick={() => editor.chain().focus().toggleBulletList().run()} active={editor.isActive('bulletList')} />
            <ToolBtn icon={ListOrdered} tooltip="有序列表" onClick={() => editor.chain().focus().toggleOrderedList().run()} active={editor.isActive('orderedList')} />
            <ToolBtn icon={Quote} tooltip="引用" onClick={() => editor.chain().focus().toggleBlockquote().run()} active={editor.isActive('blockquote')} />
            <ToolBtn icon={FileCode} tooltip="代码块" onClick={() => editor.chain().focus().toggleCodeBlock().run()} active={editor.isActive('codeBlock')} />

            <Divider layout="vertical" style={{ margin: '0 4px', height: 20 }} />

            <ToolBtn icon={LinkIcon} tooltip="链接" onClick={openLinkModal} active={editor.isActive('link')} />
            <ToolBtn icon={ImageIcon} tooltip="图片" onClick={openImageModal} />
            <ToolBtn icon={Minus} tooltip="分隔线" onClick={() => editor.chain().focus().setHorizontalRule().run()} />
          </>
        )}

        {/* 源码模式提示标签 */}
        {sourceMode && (
          <span className="tiptap-source-label">Markdown 源码模式</span>
        )}

        {/* 源码切换按钮（始终显示在最右侧） */}
        <div style={{ marginLeft: 'auto' }}>
          <ToolBtn icon={CodeXml} tooltip={sourceMode ? '退出源码' : '源码模式'} onClick={toggleSourceMode} active={sourceMode} />
        </div>
      </div>

      {/* 编辑区域：富文本 / Markdown 源码 */}
      {sourceMode ? (
        <textarea
          ref={sourceRef}
          className="tiptap-source-editor"
          value={sourceCode}
          onChange={(e) => setSourceCode(e.target.value)}
          spellCheck={false}
        />
      ) : (
        <EditorContent editor={editor} />
      )}

      {/* 链接对话框 */}
      <Modal
        title="插入链接"
        visible={linkModalVisible}
        onOk={confirmLink}
        onCancel={() => { setLinkModalVisible(false); setLinkUrl('') }}
        okText="确认"
        cancelText="取消"
        width={480}
        maskClosable={false}
      >
        <Input
          value={linkUrl}
          onChange={setLinkUrl}
          placeholder="https://example.com"
          prefix="🔗"
          size="large"
          autofocus
          onEnterPress={confirmLink}
        />
      </Modal>

      {/* 图片对话框 */}
      <Modal
        title="插入图片"
        visible={imageModalVisible}
        onOk={confirmImage}
        onCancel={() => { setImageModalVisible(false); setImageUrl('') }}
        okText="插入"
        cancelText="取消"
        width={480}
        maskClosable={false}
        okButtonProps={{ disabled: !imageUrl.trim() }}
      >
        <Input
          value={imageUrl}
          onChange={setImageUrl}
          placeholder="https://example.com/image.jpg"
          prefix="🖼️"
          size="large"
          autofocus
          onEnterPress={confirmImage}
        />
        {imageUrl.trim() && (
          <div style={{ marginTop: 12, textAlign: 'center' }}>
            <img
              src={imageUrl.trim()}
              alt="预览"
              style={{ maxWidth: '100%', maxHeight: 200, borderRadius: 4, border: '1px solid var(--semi-color-border)' }}
              onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
              onLoad={(e) => { (e.target as HTMLImageElement).style.display = 'block' }}
            />
          </div>
        )}
      </Modal>
    </div>
  )
}

/* ========== 工具栏按钮 — 纯 CSS tooltip ========== */
interface ToolBtnProps {
  icon: React.ComponentType<{ className?: string }>
  tooltip: string
  onClick: () => void
  active?: boolean
  disabled?: boolean
}

function ToolBtn({ icon: Icon, tooltip, onClick, active, disabled }: ToolBtnProps) {
  return (
    <button
      className={`tiptap-tool-btn${active ? ' active' : ''}${disabled ? ' disabled' : ''}`}
      onClick={onClick}
      disabled={disabled}
      data-tooltip={tooltip}
      type="button"
    >
      <Icon className="h-4 w-4" />
    </button>
  )
}
