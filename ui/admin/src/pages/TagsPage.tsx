import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import {
  Tooltip, TooltipContent, TooltipTrigger,
} from '@/components/ui/tooltip'
import { Search, Plus, Trash2, Hash, List, X, FileText, Loader2 } from 'lucide-react'
import { useTagList, useCreateTag, useDeleteTag } from '@/hooks/use-tags'
import { cn } from '@/lib/utils'

/**
 * 标签管理页面 — 标签云 + 列表视图
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
    createMutation.mutate(formData, { onSuccess: () => { setFormData({ name: '', slug: '' }); setShowForm(false) } })
  }

  function handleDelete(id: string) {
    if (window.confirm('确定删除此标签吗？')) deleteMutation.mutate(id)
  }

  const maxPostCount = tags ? Math.max(...tags.map((t) => t.postCount), 1) : 1
  const totalPosts = tags ? tags.reduce((sum, t) => sum + t.postCount, 0) : 0
  const hotTags = tags ? tags.filter((t) => t.postCount >= 3).length : 0
  const emptyTags = tags ? tags.filter((t) => t.postCount === 0).length : 0

  /** 根据引用数获取标签样式 */
  function getTagClasses(postCount: number): string {
    const ratio = maxPostCount > 0 ? postCount / maxPostCount : 0
    if (ratio > 0.7) return 'text-base font-medium bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950'
    if (ratio > 0.4) return 'text-sm bg-zinc-200 dark:bg-zinc-700 text-zinc-900 dark:text-zinc-100'
    if (ratio > 0.15) return 'text-sm bg-zinc-100 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300'
    return 'text-xs bg-zinc-50 dark:bg-zinc-900 text-zinc-500'
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50">标签管理</h1>
          <p className="text-sm text-zinc-500 mt-1">管理文章标签，组织内容的细粒度分类</p>
        </div>
        <Button className="bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950 shadow-none rounded-md hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={() => setShowForm(true)}>
          <Plus className="w-4 h-4 mr-1.5" /> 新建标签
        </Button>
      </div>

      {/* 统计 */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        {[
          { label: '标签总数', value: tags?.length ?? 0 },
          { label: '总引用次数', value: totalPosts },
          { label: '热门标签', value: hotTags, sub: '≥3 篇' },
          { label: '空标签', value: emptyTags, warn: emptyTags > 0 },
        ].map((s) => (
          <Card key={s.label} className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-none rounded-md">
            <CardContent className="p-4">
              <p className="text-xs text-zinc-500">{s.label}</p>
              <div className="flex items-baseline gap-1.5 mt-1">
                <span className={cn('text-2xl font-semibold', s.warn ? 'text-amber-500' : 'text-zinc-950 dark:text-zinc-50')}>{s.value}</span>
                {s.sub && <span className="text-xs text-zinc-500">{s.sub}</span>}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 新建表单 */}
      {showForm && (
        <Card className="mb-6 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-none rounded-md">
          <CardHeader className="pb-3 flex-row items-center justify-between">
            <CardTitle className="text-sm font-medium">新建标签</CardTitle>
            <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => setShowForm(false)}><X className="w-4 h-4" /></Button>
          </CardHeader>
          <CardContent className="pt-0">
            <div className="flex gap-4 items-end">
              <div className="flex-1">
                <label className="text-xs font-medium text-zinc-500 mb-1.5 block">标签名称 *</label>
                <div className="relative">
                  <Hash className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
                  <Input value={formData.name} onChange={(e) => handleNameChange(e.target.value)} placeholder="例如：Kubernetes" className="pl-9 border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none" autoFocus />
                </div>
              </div>
              <div className="flex-1">
                <label className="text-xs font-medium text-zinc-500 mb-1.5 block">Slug</label>
                <Input value={formData.slug} onChange={(e) => setFormData((p) => ({ ...p, slug: e.target.value }))} placeholder="自动生成" className="border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none" />
              </div>
              <div className="flex gap-2">
                <Button variant="outline" className="shadow-none border-zinc-200 dark:border-zinc-800" onClick={() => setShowForm(false)}>取消</Button>
                <Button className="bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950 shadow-none hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={handleCreate} disabled={!formData.name.trim() || createMutation.isPending}>
                  {createMutation.isPending ? '创建中…' : '创建标签'}
                </Button>
              </div>
            </div>
            {formData.name && (
              <div className="mt-3">
                <span className="text-xs text-zinc-500">预览：</span>
                <Badge variant="secondary" className="ml-2"># {formData.name}</Badge>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* 搜索 + 视图切换 */}
      <div className="flex justify-between items-center mb-4">
        <div className="relative max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
          <Input placeholder="搜索标签…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className="pl-9 border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none" />
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs text-zinc-500">{tags?.length ?? 0} 个标签 · {totalPosts} 次引用</span>
          <div className="flex border border-zinc-200 dark:border-zinc-800 rounded-md overflow-hidden">
            <button className={cn('px-2.5 py-1 text-xs cursor-pointer', viewMode === 'cloud' ? 'bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950' : 'text-zinc-500 hover:bg-zinc-50 dark:hover:bg-zinc-900')} onClick={() => setViewMode('cloud')}>
              <Hash className="w-3.5 h-3.5" />
            </button>
            <button className={cn('px-2.5 py-1 text-xs cursor-pointer', viewMode === 'list' ? 'bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950' : 'text-zinc-500 hover:bg-zinc-50 dark:hover:bg-zinc-900')} onClick={() => setViewMode('list')}>
              <List className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>

      {isLoading && <div className="text-center py-16"><Loader2 className="w-5 h-5 animate-spin text-zinc-400 mx-auto" /></div>}

      {!isLoading && tags?.length === 0 && (
        <div className="border border-zinc-200 dark:border-zinc-800 rounded-md bg-white dark:bg-zinc-900 py-16">
          <div className="flex flex-col items-center text-center">
            <Hash className="w-10 h-10 text-zinc-300 dark:text-zinc-700 mb-3" />
            <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">暂无标签</p>
            <p className="text-sm text-zinc-500 mt-1">点击上方按钮创建第一个标签</p>
            <Button variant="outline" size="sm" className="mt-3 shadow-none border-zinc-200 dark:border-zinc-800" onClick={() => setShowForm(true)}>新建标签</Button>
          </div>
        </div>
      )}

      {/* 标签云 */}
      {tags && tags.length > 0 && viewMode === 'cloud' && (
        <div className="border border-zinc-200 dark:border-zinc-800 rounded-md bg-white dark:bg-zinc-900 p-6">
          <div className="flex flex-wrap gap-2 items-center justify-center">
            {tags.map((tag) => (
              <Tooltip key={tag.id} delayDuration={0}>
                <TooltipTrigger asChild>
                  <span className={cn('inline-flex items-center gap-1.5 px-3 py-1 rounded-md cursor-default transition-opacity', getTagClasses(tag.postCount), tag.postCount === 0 && 'opacity-50')}>
                    # {tag.name}
                    <span className="text-[10px] opacity-60">{tag.postCount}</span>
                  </span>
                </TooltipTrigger>
                <TooltipContent className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm">
                  {tag.name} · {tag.postCount} 篇文章 · slug: {tag.slug}
                </TooltipContent>
              </Tooltip>
            ))}
          </div>
          <Separator className="my-4 bg-zinc-100 dark:bg-zinc-800" />
          <div className="flex justify-center gap-6">
            {[
              { label: '热门', cls: 'bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950', desc: '≥ 70%' },
              { label: '活跃', cls: 'bg-zinc-200 dark:bg-zinc-700', desc: '40-70%' },
              { label: '普通', cls: 'bg-zinc-100 dark:bg-zinc-800', desc: '15-40%' },
              { label: '冷门', cls: 'bg-zinc-50 dark:bg-zinc-900 text-zinc-500', desc: '< 15%' },
            ].map((l) => (
              <div key={l.label} className="flex items-center gap-1.5">
                <span className={cn('text-[10px] px-1.5 py-0.5 rounded', l.cls)}>{l.label}</span>
                <span className="text-xs text-zinc-500">{l.desc}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 列表视图 */}
      {tags && tags.length > 0 && viewMode === 'list' && (
        <div className="border border-zinc-200 dark:border-zinc-800 rounded-md bg-white dark:bg-zinc-900">
          <Table>
            <TableHeader>
              <TableRow className="border-b border-zinc-100 dark:border-zinc-800 hover:bg-transparent">
                <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider">标签名称</TableHead>
                <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[160px]">Slug</TableHead>
                <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[100px] text-center">文章数</TableHead>
                <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[120px]">热度</TableHead>
                <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[80px] text-center">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tags.map((tag) => {
                const pct = maxPostCount > 0 ? Math.round((tag.postCount / maxPostCount) * 100) : 0
                return (
                  <TableRow key={tag.id} className="border-b border-zinc-100 dark:border-zinc-800">
                    <TableCell><Badge variant="secondary" className="text-xs"># {tag.name}</Badge></TableCell>
                    <TableCell className="text-xs text-zinc-500 font-mono">{tag.slug}</TableCell>
                    <TableCell className="text-center">
                      <span className="flex items-center justify-center gap-1 text-xs text-zinc-500"><FileText className="w-3 h-3" /> {tag.postCount}</span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="flex-1 h-1.5 bg-zinc-100 dark:bg-zinc-800 rounded-full overflow-hidden">
                          <div className="h-full rounded-full bg-zinc-400 dark:bg-zinc-500 transition-all" style={{ width: `${pct}%` }} />
                        </div>
                        <span className="text-xs text-zinc-500 w-8 text-right">{pct}%</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-center">
                      <Button variant="ghost" size="icon" className="w-7 h-7 text-red-500 hover:text-red-600" onClick={() => handleDelete(tag.id)}>
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
