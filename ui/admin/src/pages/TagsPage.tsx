import { useState } from 'react'
import { Search, Plus, X, Tag as TagIcon, FileText, Trash2, Hash } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTagList, useCreateTag, useDeleteTag } from '@/hooks/use-tags'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

/**
 * 标签管理页面 - 使用 shadcn Card / Input / Badge / Button / Table
 */
export function TagsPage() {
  const [keyword, setKeyword] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({ name: '', slug: '' })
  const [viewMode, setViewMode] = useState<'cloud' | 'list'>('cloud')

  const { data: tags, isLoading } = useTagList(keyword)
  const createMutation = useCreateTag()
  const deleteMutation = useDeleteTag()

  function handleNameChange(name: string) {
    setFormData({ name, slug: name.toLowerCase().replace(/[\s]+/g, '-').replace(/[^a-z0-9\u4e00-\u9fa5-]/g, '') })
  }

  function handleCreate() {
    if (!formData.name.trim()) return
    createMutation.mutate(formData, {
      onSuccess: () => { setFormData({ name: '', slug: '' }); setShowForm(false) },
    })
  }

  function handleDelete(id: string, name: string) {
    if (window.confirm(`确定删除标签「${name}」吗？`)) deleteMutation.mutate(id)
  }

  function getTagSize(postCount: number, maxCount: number): string {
    if (maxCount === 0) return 'text-sm'
    const ratio = postCount / maxCount
    if (ratio > 0.7) return 'text-xl font-semibold'
    if (ratio > 0.4) return 'text-base font-medium'
    if (ratio > 0.15) return 'text-sm'
    return 'text-xs'
  }

  const maxPostCount = tags ? Math.max(...tags.map((t) => t.postCount), 1) : 1
  const totalPosts = tags ? tags.reduce((sum, t) => sum + t.postCount, 0) : 0

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">标签管理</h1>
          <p className="mt-1 text-sm text-muted-foreground">管理文章标签，组织内容的细粒度分类</p>
        </div>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          新建标签
        </Button>
      </div>

      {/* 新建表单 */}
      {showForm && (
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">新建标签</CardTitle>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setShowForm(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex gap-4">
              <div className="flex-1">
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">标签名称</label>
                <Input value={formData.name} onChange={(e) => handleNameChange(e.target.value)} placeholder="例如：Kubernetes" />
              </div>
              <div className="flex-1">
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">Slug</label>
                <Input value={formData.slug} onChange={(e) => setFormData((p) => ({ ...p, slug: e.target.value }))} placeholder="例如：kubernetes" />
              </div>
              <div className="flex items-end gap-2">
                <Button variant="outline" onClick={() => setShowForm(false)}>取消</Button>
                <Button onClick={handleCreate} disabled={!formData.name.trim() || createMutation.isPending}>
                  {createMutation.isPending ? '创建中…' : '创建'}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 搜索 + 视图切换 */}
      <div className="mb-4 flex items-center justify-between gap-4">
        <div className="relative max-w-sm flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="搜索标签…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className="pl-9" />
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs text-muted-foreground">{tags?.length ?? 0} 个标签 · {totalPosts} 次引用</span>
          <div className="flex">
            <Button variant={viewMode === 'cloud' ? 'default' : 'outline'} size="icon" className="h-8 w-8 rounded-r-none" onClick={() => setViewMode('cloud')} title="标签云">
              <Hash className="h-3.5 w-3.5" />
            </Button>
            <Button variant={viewMode === 'list' ? 'default' : 'outline'} size="icon" className="h-8 w-8 rounded-l-none" onClick={() => setViewMode('list')} title="列表">
              <FileText className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      </div>

      {isLoading && <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">加载中…</div>}

      {!isLoading && tags?.length === 0 && (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-16 text-muted-foreground">
            <TagIcon className="mb-3 h-8 w-8" strokeWidth={1} />
            <p className="text-sm">暂无标签，点击上方按钮创建</p>
          </CardContent>
        </Card>
      )}

      {/* 标签云 */}
      {tags && tags.length > 0 && viewMode === 'cloud' && (
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-wrap gap-2.5">
              {tags.map((tag) => (
                <div key={tag.id} className="group relative">
                  <Badge variant="outline" className={cn('gap-1.5 px-3 py-1.5 cursor-default', getTagSize(tag.postCount, maxPostCount))}>
                    <Hash className="h-3 w-3 text-muted-foreground" strokeWidth={1.5} />
                    {tag.name}
                    <span className="text-xs font-normal text-muted-foreground">{tag.postCount}</span>
                  </Badge>
                  <Button
                    variant="destructive" size="icon"
                    className="absolute -right-1 -top-1 h-4 w-4 opacity-0 transition-opacity group-hover:opacity-100"
                    onClick={() => handleDelete(tag.id, tag.name)}
                    title="删除"
                  >
                    <X className="h-2.5 w-2.5" />
                  </Button>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* 列表 */}
      {tags && tags.length > 0 && viewMode === 'list' && (
        <Card>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>标签名称</TableHead>
                <TableHead className="w-[120px]">Slug</TableHead>
                <TableHead className="w-[100px] text-center">文章数</TableHead>
                <TableHead className="w-[60px] text-center">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tags.map((tag) => (
                <TableRow key={tag.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Hash className="h-3.5 w-3.5 text-muted-foreground" strokeWidth={1.5} />
                      <span className="font-medium">{tag.name}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-muted-foreground">{tag.slug}</TableCell>
                  <TableCell className="text-center">
                    <div className="flex items-center justify-center gap-1 text-muted-foreground">
                      <FileText className="h-3 w-3" strokeWidth={1.5} />
                      {tag.postCount}
                    </div>
                  </TableCell>
                  <TableCell className="text-center">
                    <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive hover:text-destructive" onClick={() => handleDelete(tag.id, tag.name)} title="删除">
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      )}
    </div>
  )
}
