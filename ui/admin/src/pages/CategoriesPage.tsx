import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Search, Plus, Pencil, Trash2, FolderOpen, FileText, X, Loader2 } from 'lucide-react'
import { useCategoryList, useCreateCategory, useDeleteCategory } from '@/hooks/use-categories'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search as SearchBtn } from '@/components/search'

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

/**
 * 分类管理页面
 */
export function CategoriesPage() {
  const [keyword, setKeyword] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({ name: '', slug: '', description: '' })

  const { data: categories, isLoading } = useCategoryList(keyword)
  const createMutation = useCreateCategory()
  const deleteMutation = useDeleteCategory()

  function handleNameChange(name: string) {
    setFormData((prev) => ({ ...prev, name, slug: name.toLowerCase().replace(/[\s]+/g, '-').replace(/[^a-z0-9\u4e00-\u9fa5-]/g, '') }))
  }

  function handleCreate() {
    if (!formData.name.trim()) return
    createMutation.mutate(formData, { onSuccess: () => { setFormData({ name: '', slug: '', description: '' }); setShowForm(false) } })
  }

  function handleDelete(id: string, name: string) {
    if (window.confirm(`确定删除分类「${name}」吗？此操作不可撤销。`)) deleteMutation.mutate(id)
  }

  return (
    <>
      <Header fixed>
        <SearchBtn />
        <div className='ml-auto' />
      </Header>
      <Main>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50">分类管理</h1>
          <p className="text-sm text-zinc-500 mt-1">管理博客文章的分类体系</p>
        </div>
        <Button className="bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950 shadow-none rounded-md hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={() => setShowForm(true)}>
          <Plus className="w-4 h-4 mr-1.5" /> 新建分类
        </Button>
      </div>

      {/* 新建表单 */}
      {showForm && (
        <Card className="mb-6 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-none rounded-md">
          <CardHeader className="pb-3 flex-row items-center justify-between">
            <CardTitle className="text-sm font-medium">新建分类</CardTitle>
            <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => setShowForm(false)}>
              <X className="w-4 h-4" />
            </Button>
          </CardHeader>
          <CardContent className="pt-0">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-xs font-medium text-zinc-500 mb-1.5 block">分类名称</label>
                <Input value={formData.name} onChange={(e) => handleNameChange(e.target.value)} placeholder="例如：前端开发" className="border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none rounded-md" />
              </div>
              <div>
                <label className="text-xs font-medium text-zinc-500 mb-1.5 block">Slug</label>
                <Input value={formData.slug} onChange={(e) => setFormData((prev) => ({ ...prev, slug: e.target.value }))} placeholder="例如：frontend" className="border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none rounded-md" />
              </div>
            </div>
            <div className="mt-4">
              <label className="text-xs font-medium text-zinc-500 mb-1.5 block">描述</label>
              <Input value={formData.description} onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))} placeholder="简要描述此分类的内容范围…" className="border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none rounded-md" />
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <Button variant="outline" className="shadow-none border-zinc-200 dark:border-zinc-800" onClick={() => setShowForm(false)}>取消</Button>
              <Button className="bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950 shadow-none rounded-md hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={handleCreate} disabled={!formData.name.trim() || createMutation.isPending}>
                {createMutation.isPending ? '创建中…' : '创建'}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 搜索 */}
      <div className="mb-4 max-w-sm">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
          <Input placeholder="搜索分类名称或 Slug…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className="pl-9 border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none rounded-md" />
        </div>
      </div>

      {isLoading && (
        <div className="text-center py-16">
          <Loader2 className="w-5 h-5 animate-spin text-zinc-400 mx-auto" />
        </div>
      )}

      {!isLoading && categories?.length === 0 && (
        <div className="border border-zinc-200 dark:border-zinc-800 rounded-md bg-white dark:bg-zinc-900 py-16">
          <div className="flex flex-col items-center text-center">
            <FolderOpen className="w-10 h-10 text-zinc-300 dark:text-zinc-700 mb-3" />
            <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">暂无分类</p>
            <p className="text-sm text-zinc-500 mt-1">点击上方按钮创建第一个分类</p>
            <Button variant="outline" size="sm" className="mt-3 shadow-none border-zinc-200 dark:border-zinc-800" onClick={() => setShowForm(true)}>新建分类</Button>
          </div>
        </div>
      )}

      {/* 分类卡片网格 */}
      {categories && categories.length > 0 && (
        <>
          <div className="grid grid-cols-3 gap-4">
            {categories.map((cat) => (
              <Card key={cat.id} className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-none rounded-md">
                <CardContent className="p-5">
                  <div className="flex justify-between items-start">
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium text-zinc-950 dark:text-zinc-50">{cat.name}</p>
                      <p className="text-xs text-zinc-500 mt-0.5">/{cat.slug}</p>
                    </div>
                    <div className="flex gap-0.5">
                      <Button variant="ghost" size="icon" className="w-7 h-7"><Pencil className="w-3.5 h-3.5" /></Button>
                      <Button variant="ghost" size="icon" className="w-7 h-7 text-red-500 hover:text-red-600" onClick={() => handleDelete(cat.id, cat.name)}><Trash2 className="w-3.5 h-3.5" /></Button>
                    </div>
                  </div>
                  {cat.description && <p className="text-sm text-zinc-500 mt-3">{cat.description}</p>}
                  <Separator className="my-3 bg-zinc-100 dark:bg-zinc-800" />
                  <div className="flex justify-between items-center text-xs text-zinc-500">
                    <span className="flex items-center gap-1"><FileText className="w-3 h-3" /> {cat.postCount} 篇文章</span>
                    <span>更新于 {formatDate(cat.updatedAt)}</span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
          <p className="text-xs text-zinc-500 mt-4">共 {categories.length} 个分类</p>
        </>
      )}
    </Main>
    </>
  )
}
