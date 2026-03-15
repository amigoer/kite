import { NodeViewWrapper } from '@tiptap/react'
import type { NodeViewProps } from '@tiptap/react'
import { useRef, useState, useCallback } from 'react'

/**
 * 可调整尺寸的图片 NodeView
 * 选中时显示八个方向的拖拽手柄
 */
export function ResizableImageView({ node, updateAttributes, selected }: NodeViewProps) {
  const { src, alt, title, width } = node.attrs
  const imgRef = useRef<HTMLImageElement>(null)
  const [resizing, setResizing] = useState(false)

  /** 拖拽调整尺寸 */
  const handleMouseDown = useCallback((e: React.MouseEvent, direction: string) => {
    e.preventDefault()
    e.stopPropagation()
    setResizing(true)

    const startX = e.clientX
    const startY = e.clientY
    const img = imgRef.current
    if (!img) return

    const startWidth = img.offsetWidth
    const startHeight = img.offsetHeight
    const aspectRatio = startWidth / startHeight

    const handleMouseMove = (moveEvent: MouseEvent) => {
      let newWidth = startWidth
      const deltaX = moveEvent.clientX - startX
      const deltaY = moveEvent.clientY - startY

      if (direction.includes('right')) {
        newWidth = startWidth + deltaX
      } else if (direction.includes('left')) {
        newWidth = startWidth - deltaX
      }

      // 保持宽高比（仅在对角方向时使用较大的变化量）
      if (direction.includes('top') || direction.includes('bottom')) {
        const heightBasedWidth = (startHeight + (direction.includes('bottom') ? deltaY : -deltaY)) * aspectRatio
        if (!direction.includes('left') && !direction.includes('right')) {
          newWidth = heightBasedWidth
        } else {
          newWidth = Math.max(newWidth, heightBasedWidth)
        }
      }

      // 限制最小/最大宽度
      newWidth = Math.max(50, Math.min(newWidth, 1200))
      updateAttributes({ width: Math.round(newWidth) })
    }

    const handleMouseUp = () => {
      setResizing(false)
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
  }, [updateAttributes])

  return (
    <NodeViewWrapper className="resizable-image-wrapper" data-drag-handle>
      <div className={`resizable-image${selected || resizing ? ' selected' : ''}`}>
        <img
          ref={imgRef}
          src={src}
          alt={alt || ''}
          title={title || ''}
          style={{ width: width ? `${width}px` : undefined }}
          draggable={false}
        />
        {/* 拖拽手柄（选中时才显示） */}
        {(selected || resizing) && (
          <>
            <div className="resize-handle nw" onMouseDown={(e) => handleMouseDown(e, 'top-left')} />
            <div className="resize-handle ne" onMouseDown={(e) => handleMouseDown(e, 'top-right')} />
            <div className="resize-handle sw" onMouseDown={(e) => handleMouseDown(e, 'bottom-left')} />
            <div className="resize-handle se" onMouseDown={(e) => handleMouseDown(e, 'bottom-right')} />
            <div className="resize-handle n"  onMouseDown={(e) => handleMouseDown(e, 'top')} />
            <div className="resize-handle s"  onMouseDown={(e) => handleMouseDown(e, 'bottom')} />
            <div className="resize-handle w"  onMouseDown={(e) => handleMouseDown(e, 'left')} />
            <div className="resize-handle e"  onMouseDown={(e) => handleMouseDown(e, 'right')} />
          </>
        )}
        {/* 宽度指示器 */}
        {resizing && width && (
          <div className="resize-indicator">{width}px</div>
        )}
      </div>
    </NodeViewWrapper>
  )
}
