import { useState } from 'react'
import { Search, Plus, FolderTree, FileText, Pencil, Trash2, X } from 'lucide-react'
import { useCategoryList, useCreateCategory, useDeleteCategory } from '@/hooks/use-categories'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

/** 格式化日期 */
function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

/**
 * 分类管理页面 - 使用 shadcn Card / Input / Button
 */
export function CategoriesPage() {
  const [keyword, setKeyword] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({ name: '', slug: '', description: '' })

  const { data: categories, isLoading } = useCategoryList(keyword)
  const createMutation = useCreateCategory()
  const deleteMutation = useDeleteCategory()

  function handleNameChange(name: string) {
    setFormData((prev) => ({
      ...prev, name,
      slug: name.toLowerCase().replace(/[\s]+/g, '-').replace(/[^a-z0-9\u4e00-\u9fa5-]/g, ''),
    }))
  }

  function handleCreate() {
    if (!formData.name.trim()) return
    createMutation.mutate(formData, {
      onSuccess: () => { setFormData({ name: '', slug: '', description: '' }); setShowForm(false) },
    })
  }

  function handleDelete(id: string, name: string) {
    if (window.confirm(`确定删除分类「${name}」吗？此操作不可撤销。`)) {
      deleteMutation.mutate(id)
    }
  }

  return (
    <div>
      {/* 标题区 */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">分类管理</h1>
          <p className="mt-1 text-sm text-muted-foreground">管理博客文章的分类体系</p>
        </div>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          新建分类
        </Button>
      </div>

      {/* 新建表单 */}
      {showForm && (
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">新建分类</CardTitle>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setShowForm(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">分类名称</label>
                <Input value={formData.name} onChange={(e) => handleNameChange(e.target.value)} placeholder="例如：前端开发" />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">Slug</label>
                <Input value={formData.slug} onChange={(e) => setFormData((prev) => ({ ...prev, slug: e.target.value }))} placeholder="例如：frontend" />
              </div>
            </div>
            <div className="mt-4">
              <label className="mb-1.5 block text-xs font-medium text-muted-foreground">描述</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
                placeholder="简要描述此分类的内容范围…"
                rows={2}
                className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
              />
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setShowForm(false)}>取消</Button>
              <Button onClick={handleCreate} disabled={!formData.name.trim() || createMutation.isPending}>
                {createMutation.isPending ? '创建中…' : '创建'}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 搜索栏 */}
      <div className="mb-4 relative max-w-sm">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input placeholder="搜索分类名称或 Slug…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className="pl-9" />
      </div>

      {/* 加载态 */}
      {isLoading && (
        <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">加载中…</div>
      )}

      {/* 空态 */}
      {!isLoading && categories?.length === 0 && (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-16 text-muted-foreground">
            <FolderTree className="mb-3 h-8 w-8" strokeWidth={1} />
            <p className="text-sm">暂无分类，点击上方按钮创建第一个分类</p>
          </CardContent>
        </Card>
      )}

      {/* 分类卡片网格 */}
      {categories && categories.length > 0 && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {categories.map((cat) => (
            <Card key={cat.id} className="group">
              <CardContent className="pt-5">
                <div className="flex items-start justify-between">
                  <div className="min-w-0 flex-1">
                    <h3 className="text-sm font-semibold">{cat.name}</h3>
                    <p className="mt-0.5 text-xs text-muted-foreground">/{cat.slug}</p>
                  </div>
                  <div className="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                    <Button variant="ghost" size="icon" className="h-7 w-7" title="编辑">
                      <Pencil className="h-3.5 w-3.5" />
                    </Button>
                    <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive hover:text-destructive" title="删除" onClick={() => handleDelete(cat.id, cat.name)}>
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
                <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{cat.description}</p>
                <Separator className="my-3" />
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <FileText className="h-3.5 w-3.5" strokeWidth={1.5} />
                    <span>{cat.postCount} 篇文章</span>
                  </div>
                  <span className="text-xs text-muted-foreground">更新于 {formatDate(cat.updatedAt)}</span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {categories && categories.length > 0 && (
        <div className="mt-4 text-sm text-muted-foreground">共 {categories.length} 个分类</div>
      )}
    </div>
  )
}
