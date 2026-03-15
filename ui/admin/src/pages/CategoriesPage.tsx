import { useState } from 'react'
import {
  Search,
  Plus,
  FolderTree,
  FileText,
  Pencil,
  Trash2,
  X,
} from 'lucide-react'
import { cn } from '@/lib/cn'
import { useCategoryList, useCreateCategory, useDeleteCategory } from '@/hooks/use-categories'

/** 格式化日期 */
function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

/**
 * 分类管理页面
 * 卡片网格布局 + 新建对话框 + 搜索
 */
export function CategoriesPage() {
  const [keyword, setKeyword] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({ name: '', slug: '', description: '' })

  const { data: categories, isLoading } = useCategoryList(keyword)
  const createMutation = useCreateCategory()
  const deleteMutation = useDeleteCategory()

  /** 自动生成 slug */
  function handleNameChange(name: string) {
    setFormData((prev) => ({
      ...prev,
      name,
      slug: name
        .toLowerCase()
        .replace(/[\s]+/g, '-')
        .replace(/[^a-z0-9\u4e00-\u9fa5-]/g, ''),
    }))
  }

  /** 提交新建 */
  function handleCreate() {
    if (!formData.name.trim()) return
    createMutation.mutate(formData, {
      onSuccess: () => {
        setFormData({ name: '', slug: '', description: '' })
        setShowForm(false)
      },
    })
  }

  /** 确认删除 */
  function handleDelete(id: string, name: string) {
    if (window.confirm(`确定删除分类「${name}」吗？此操作不可撤销。`)) {
      deleteMutation.mutate(id)
    }
  }

  return (
    <div>
      {/* 页面标题区 */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--kite-text-heading)]">
            分类管理
          </h1>
          <p className="mt-1 text-sm text-[var(--kite-text-muted)]">
            管理博客文章的分类体系
          </p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex h-9 items-center gap-2 border border-[var(--kite-accent)] bg-[var(--kite-accent)] px-4 text-sm font-medium text-white transition-colors duration-100 hover:bg-[#333] cursor-pointer"
        >
          <Plus className="h-4 w-4" strokeWidth={1.5} />
          新建分类
        </button>
      </div>

      {/* 新建分类表单 */}
      {showForm && (
        <div className="mb-6 border border-[var(--kite-border)] bg-[var(--kite-bg)] p-5">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-[var(--kite-text-heading)]">新建分类</h2>
            <button
              onClick={() => setShowForm(false)}
              className="flex h-7 w-7 items-center justify-center text-[var(--kite-text-muted)] hover:text-[var(--kite-text-heading)] cursor-pointer"
            >
              <X className="h-4 w-4" strokeWidth={1.5} />
            </button>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1.5 block text-xs font-medium text-[var(--kite-text-muted)]">
                分类名称
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="例如：前端开发"
                className="h-9 w-full border border-[var(--kite-border)] bg-[var(--kite-bg)] px-3 text-sm text-[var(--kite-text)] outline-none placeholder:text-[var(--kite-text-muted)] focus:border-[var(--kite-accent)]"
              />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium text-[var(--kite-text-muted)]">
                Slug
              </label>
              <input
                type="text"
                value={formData.slug}
                onChange={(e) => setFormData((prev) => ({ ...prev, slug: e.target.value }))}
                placeholder="例如：frontend"
                className="h-9 w-full border border-[var(--kite-border)] bg-[var(--kite-bg)] px-3 text-sm text-[var(--kite-text)] outline-none placeholder:text-[var(--kite-text-muted)] focus:border-[var(--kite-accent)]"
              />
            </div>
          </div>
          <div className="mt-4">
            <label className="mb-1.5 block text-xs font-medium text-[var(--kite-text-muted)]">
              描述
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
              placeholder="简要描述此分类的内容范围…"
              rows={2}
              className="w-full resize-none border border-[var(--kite-border)] bg-[var(--kite-bg)] px-3 py-2 text-sm text-[var(--kite-text)] outline-none placeholder:text-[var(--kite-text-muted)] focus:border-[var(--kite-accent)]"
            />
          </div>
          <div className="mt-4 flex justify-end gap-2">
            <button
              onClick={() => setShowForm(false)}
              className="h-9 border border-[var(--kite-border)] bg-transparent px-4 text-sm text-[var(--kite-text)] transition-colors duration-100 hover:border-[var(--kite-border-hover)] cursor-pointer"
            >
              取消
            </button>
            <button
              onClick={handleCreate}
              disabled={!formData.name.trim() || createMutation.isPending}
              className="h-9 border border-[var(--kite-accent)] bg-[var(--kite-accent)] px-4 text-sm font-medium text-white transition-colors duration-100 hover:bg-[#333] disabled:opacity-50 cursor-pointer"
            >
              {createMutation.isPending ? '创建中…' : '创建'}
            </button>
          </div>
        </div>
      )}

      {/* 搜索栏 */}
      <div className="mb-4">
        <div className="relative max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--kite-text-muted)]" strokeWidth={1.5} />
          <input
            type="text"
            placeholder="搜索分类名称或 Slug…"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            className="h-9 w-full border border-[var(--kite-border)] bg-[var(--kite-bg)] pl-9 pr-3 text-sm text-[var(--kite-text)] outline-none placeholder:text-[var(--kite-text-muted)] focus:border-[var(--kite-accent)]"
          />
        </div>
      </div>

      {/* 加载态 */}
      {isLoading && (
        <div className="flex items-center justify-center py-16 text-sm text-[var(--kite-text-muted)]">
          加载中…
        </div>
      )}

      {/* 空态 */}
      {!isLoading && categories?.length === 0 && (
        <div className="flex flex-col items-center justify-center border border-dashed border-[var(--kite-border)] py-16 text-sm text-[var(--kite-text-muted)]">
          <FolderTree className="mb-3 h-8 w-8" strokeWidth={1} />
          <p>暂无分类，点击上方按钮创建第一个分类</p>
        </div>
      )}

      {/* 分类卡片网格 */}
      {categories && categories.length > 0 && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {categories.map((cat) => (
            <div
              key={cat.id}
              className="group border border-[var(--kite-border)] bg-[var(--kite-bg)] p-5 transition-colors duration-100 hover:border-[var(--kite-border-hover)]"
            >
              {/* 标题行 */}
              <div className="flex items-start justify-between">
                <div className="min-w-0 flex-1">
                  <h3 className="text-sm font-semibold text-[var(--kite-text-heading)]">
                    {cat.name}
                  </h3>
                  <p className="mt-0.5 text-xs text-[var(--kite-text-muted)]">
                    /{cat.slug}
                  </p>
                </div>
                {/* 操作按钮 - hover 时显示 */}
                <div className="flex gap-1 opacity-0 transition-opacity duration-100 group-hover:opacity-100">
                  <button
                    className="flex h-7 w-7 items-center justify-center border border-transparent text-[var(--kite-text-muted)] transition-colors duration-100 hover:border-[var(--kite-border)] hover:text-[var(--kite-text-heading)] cursor-pointer"
                    title="编辑"
                  >
                    <Pencil className="h-3.5 w-3.5" strokeWidth={1.5} />
                  </button>
                  <button
                    onClick={() => handleDelete(cat.id, cat.name)}
                    className="flex h-7 w-7 items-center justify-center border border-transparent text-[var(--kite-text-muted)] transition-colors duration-100 hover:border-red-300 hover:text-red-600 cursor-pointer"
                    title="删除"
                  >
                    <Trash2 className="h-3.5 w-3.5" strokeWidth={1.5} />
                  </button>
                </div>
              </div>

              {/* 描述 */}
              <p className="mt-3 text-sm leading-relaxed text-[var(--kite-text)]">
                {cat.description}
              </p>

              {/* 底部统计 */}
              <div className="mt-4 flex items-center justify-between border-t border-[var(--kite-border)] pt-3">
                <div className="flex items-center gap-1.5 text-xs text-[var(--kite-text-muted)]">
                  <FileText className="h-3.5 w-3.5" strokeWidth={1.5} />
                  <span>{cat.postCount} 篇文章</span>
                </div>
                <span className="text-xs text-[var(--kite-text-muted)]">
                  更新于 {formatDate(cat.updatedAt)}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 底部统计 */}
      {categories && categories.length > 0 && (
        <div className={cn('mt-4 text-sm text-[var(--kite-text-muted)]')}>
          共 {categories.length} 个分类
        </div>
      )}
    </div>
  )
}
