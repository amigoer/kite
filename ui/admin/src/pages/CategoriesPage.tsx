import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Search, Plus, Pencil, Trash2, FolderOpen, FileText, X, Loader2, CornerDownRight, ChevronRight, ChevronDown } from 'lucide-react'
import { useCategoryList, useCreateCategory, useDeleteCategory } from '@/hooks/use-categories'
import type { Category } from '@/types/category'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search as SearchBtn } from '@/components/search'
import { useConfirm } from '@/hooks/use-confirm'
import { CategoryCascader, buildCascaderTree } from '@/components/category-cascader'

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

/**
 * 递归构建多级分类树
 */
function buildTree(flat: Category[]): Category[] {
  const nodeMap = new Map<string, Category>()
  for (const cat of flat) {
    nodeMap.set(cat.id, { ...cat, children: [] })
  }

  const roots: Category[] = []
  for (const cat of flat) {
    const node = nodeMap.get(cat.id)!
    if (cat.parentId) {
      const parent = nodeMap.get(cat.parentId)
      if (parent) {
        parent.children = parent.children || []
        parent.children.push(node)
      } else {
        roots.push(node)
      }
    } else {
      roots.push(node)
    }
  }
  return roots
}

/**
 * 计算节点在树中的深度（用于缩进和标签显示）
 */
function getDepthLabel(depth: number): string {
  if (depth === 0) return '顶级'
  return `L${depth + 1}`
}

// ─── 递归渲染分类节点 ──────────────────────────────────────────────

interface CategoryNodeProps {
  cat: Category
  depth: number
  onDelete: (id: string, name: string) => void
}

/** 单个分类节点（支持递归渲染子节点） */
function CategoryNode({ cat, depth, onDelete }: CategoryNodeProps) {
  const [expanded, setExpanded] = useState(true)
  const hasChildren = !!(cat.children?.length)

  return (
    <div>
      {/* 当前节点行 */}
      <div className="flex items-center gap-3 px-5 py-3 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors"
        style={{ paddingLeft: `${20 + depth * 24}px` }}
      >
        {/* 展开/折叠 或占位 */}
        {hasChildren ? (
          <button
            className="w-5 h-5 flex items-center justify-center text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300 cursor-pointer rounded-sm hover:bg-zinc-100 dark:hover:bg-zinc-700 transition-colors"
            onClick={() => setExpanded(!expanded)}
          >
            {expanded ? <ChevronDown className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
          </button>
        ) : (
          <span className="w-5 h-5 flex items-center justify-center">
            <CornerDownRight className="w-3 h-3 text-zinc-300 dark:text-zinc-600" />
          </span>
        )}

        {/* 图标 */}
        {depth === 0 ? (
          <div className="w-8 h-8 rounded-md bg-zinc-900 dark:bg-zinc-100 flex items-center justify-center shrink-0">
            <FolderOpen className="w-4 h-4 text-white dark:text-zinc-900" />
          </div>
        ) : (
          <div className="w-7 h-7 rounded-md bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center shrink-0">
            <FolderOpen className="w-3.5 h-3.5 text-zinc-400" />
          </div>
        )}

        {/* 名称 & slug */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <p className={`text-sm ${depth === 0 ? 'font-semibold text-zinc-950 dark:text-zinc-50' : 'text-zinc-700 dark:text-zinc-300'}`}>
              {cat.name}
            </p>
            <span className="text-[10px] px-1.5 py-0.5 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500 font-medium">
              {getDepthLabel(depth)}
            </span>
            {hasChildren && (
              <span className="text-[10px] text-zinc-400">{cat.children!.length} 个子分类</span>
            )}
          </div>
          <p className="text-xs text-zinc-400 font-mono mt-0.5">/{cat.slug}</p>
        </div>

        {/* 信息 & 操作 */}
        <span className="text-xs text-zinc-500 flex items-center gap-1"><FileText className="w-3 h-3" /> {cat.postCount} 篇</span>
        <span className="text-xs text-zinc-400 w-24 text-right">{formatDate(cat.updatedAt)}</span>
        <div className="flex gap-0.5">
          <Button variant="ghost" size="icon" className="w-7 h-7"><Pencil className="w-3.5 h-3.5" /></Button>
          <Button variant="ghost" size="icon" className="w-7 h-7 text-red-500 hover:text-red-600" onClick={() => onDelete(cat.id, cat.name)}><Trash2 className="w-3.5 h-3.5" /></Button>
        </div>
      </div>

      {/* 递归渲染子节点 */}
      {hasChildren && expanded && (
        <div className="border-l-2 border-zinc-100 dark:border-zinc-800" style={{ marginLeft: `${32 + depth * 24}px` }}>
          {cat.children!.map((child, childIdx) => (
            <div key={child.id}>
              {childIdx > 0 && <Separator className="bg-zinc-50 dark:bg-zinc-800/30" />}
              <CategoryNode cat={child} depth={depth + 1} onDelete={onDelete} />
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// ─── 主页面 ────────────────────────────────────────────────────────

/**
 * 分类管理页面 — 支持多级分类
 */
export function CategoriesPage() {
  const [keyword, setKeyword] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({ name: '', slug: '', description: '', parentId: '' })

  const { data: categories, isLoading } = useCategoryList(keyword)
  const createMutation = useCreateCategory()
  const deleteMutation = useDeleteCategory()
  const { confirm, ConfirmDialog } = useConfirm()

  // 构建树结构
  const tree = categories ? buildTree(categories) : []

  function handleNameChange(name: string) {
    setFormData((prev) => ({ ...prev, name, slug: name.toLowerCase().replace(/[\s]+/g, '-').replace(/[^a-z0-9\u4e00-\u9fa5-]/g, '') }))
  }

  function handleCreate() {
    if (!formData.name.trim()) return
    createMutation.mutate(
      {
        name: formData.name,
        slug: formData.slug,
        description: formData.description,
        parent_id: formData.parentId || undefined,
      },
      { onSuccess: () => { setFormData({ name: '', slug: '', description: '', parentId: '' }); setShowForm(false) } },
    )
  }

  async function handleDelete(id: string, name: string) {
    if (await confirm({ title: '删除分类', description: `确定删除分类「${name}」吗？此操作不可撤销。`, confirmText: '删除', variant: 'destructive' })) {
      deleteMutation.mutate(id)
    }
  }

  return (
    <>
      <ConfirmDialog />
      <Header fixed>
        <SearchBtn />
        <div className='ml-auto' />
      </Header>
      <Main>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50">分类管理</h1>
          <p className="text-sm text-zinc-500 mt-1">管理博客文章的分类体系，支持多级分类结构</p>
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
            <div className="grid grid-cols-3 gap-4">
              <div>
                <label className="text-xs font-medium text-zinc-500 mb-1.5 block">分类名称</label>
                <Input value={formData.name} onChange={(e) => handleNameChange(e.target.value)} placeholder="例如：前端开发" className="border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none rounded-md" />
              </div>
              <div>
                <label className="text-xs font-medium text-zinc-500 mb-1.5 block">Slug</label>
                <Input value={formData.slug} onChange={(e) => setFormData((prev) => ({ ...prev, slug: e.target.value }))} placeholder="例如：frontend" className="border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none rounded-md" />
              </div>
              <div>
                <label className="text-xs font-medium text-zinc-500 mb-1.5 block">父分类</label>
                <CategoryCascader
                  options={buildCascaderTree(categories || [])}
                  value={formData.parentId || null}
                  onChange={(id) => setFormData((prev) => ({ ...prev, parentId: id || '' }))}
                  placeholder="无（顶级分类）"
                  allowSelectParent
                />
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

      {/* 分类列表 — 递归树形展示 */}
      {tree.length > 0 && (
        <>
          <div className="border border-zinc-200 dark:border-zinc-800 rounded-lg bg-white dark:bg-zinc-900 overflow-hidden">
            {tree.map((root, idx) => (
              <div key={root.id}>
                {idx > 0 && <Separator className="bg-zinc-100 dark:bg-zinc-800" />}
                <CategoryNode cat={root} depth={0} onDelete={handleDelete} />
              </div>
            ))}
          </div>
          <p className="text-xs text-zinc-500 mt-4">共 {categories?.length || 0} 个分类</p>
        </>
      )}
    </Main>
    </>
  )
}
