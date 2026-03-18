import { useState, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { Upload, Trash2, Loader2 } from 'lucide-react'
import { apiUpload } from '@/lib/api-client'
import { cn } from '@/lib/utils'

interface ImageUploaderProps {
  value?: string
  onChange?: (url: string) => void
  placeholder?: string
}

/**
 * 图片上传组件 — 支持上传和拖拽
 */
export function ImageUploader({ value, onChange, placeholder = '上传或粘贴图片 URL' }: ImageUploaderProps) {
  const [uploading, setUploading] = useState(false)
  const [dragOver, setDragOver] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  async function handleFile(file: File) {
    if (!file.type.startsWith('image/')) return
    setUploading(true)
    try {
      const result = await apiUpload(file)
      onChange?.(result.url)
    } catch (e) {
      console.error('上传失败', e)
    } finally {
      setUploading(false)
    }
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) handleFile(file)
  }

  return (
    <div>
      {value ? (
        <div>
          <img src={value} alt="封面预览" className="w-full rounded-md border border-zinc-200 dark:border-zinc-800" />
          <div className="flex gap-1.5 mt-2">
            <Button variant="outline" size="sm" className="h-7 text-xs shadow-none border-zinc-200 dark:border-zinc-800" onClick={() => fileInputRef.current?.click()} disabled={uploading}>
              <Upload className="w-3 h-3 mr-1" /> {uploading ? '上传中…' : '更换'}
            </Button>
            <Button variant="outline" size="sm" className="h-7 text-xs shadow-none border-zinc-200 dark:border-zinc-800 text-red-500 hover:text-red-600" onClick={() => onChange?.('')}>
              <Trash2 className="w-3 h-3 mr-1" /> 移除
            </Button>
          </div>
        </div>
      ) : (
        <div
          className={cn(
            'border-2 border-dashed rounded-md p-5 text-center cursor-pointer transition-colors',
            dragOver
              ? 'border-zinc-400 dark:border-zinc-600 bg-zinc-50 dark:bg-zinc-900'
              : 'border-zinc-200 dark:border-zinc-800 bg-transparent hover:border-zinc-300 dark:hover:border-zinc-700'
          )}
          onClick={() => fileInputRef.current?.click()}
          onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleDrop}
        >
          {uploading ? (
            <Loader2 className="w-5 h-5 animate-spin text-zinc-400 mx-auto" />
          ) : (
            <>
              <Upload className="w-6 h-6 text-zinc-400 mx-auto mb-1" />
              <p className="text-xs text-zinc-500">{placeholder}</p>
              <p className="text-[10px] text-zinc-400 mt-1">支持 JPG / PNG / WebP / GIF</p>
            </>
          )}
        </div>
      )}
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={(e) => { const f = e.target.files?.[0]; if (f) handleFile(f); e.target.value = '' }}
      />
    </div>
  )
}
