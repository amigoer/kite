import { useState } from 'react'
import { toast } from 'sonner'
import { apiPost } from '@/lib/api-client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog'
import {
  Search, Plus, Trash2, ExternalLink, Link2, Loader2,
  ArrowUpCircle, ArrowDownCircle, Image as ImageIcon, Wand2,
} from 'lucide-react'
import {
  useFriendLinks, useCreateFriendLink, useUpdateFriendLink,
  useDeleteFriendLink, useToggleLinkStatus,
} from '@/hooks/use-friend-links'
import type { FriendLinkFormData } from '@/hooks/use-friend-links'
import type { FriendLink, LinkStatus } from '@/types/friend-link'
import { cn } from '@/lib/utils'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search as HeaderSearch } from '@/components/search'
import { useConfirm } from '@/hooks/use-confirm'

const statusConfig: Record<LinkStatus, { text: string }> = {
  active: { text: '正常' },
  pending: { text: '待审核' },
  down: { text: '已下线' },
}

function extractDomain(url: string): string {
  try { return new URL(url).hostname } catch { return url }
}

const inputCls = 'border-zinc-200 dark:border-zinc-700 bg-transparent rounded-md shadow-none focus-visible:ring-1 focus-visible:ring-zinc-400'

/** 空表单默认值 */
const emptyForm: FriendLinkFormData = { name: '', url: '', description: '', logo: '', sort: 0 }

/**
 * 友链管理
 */
export function FriendLinksPage() {
  const [keyword, setKeyword] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [formData, setFormData] = useState<FriendLinkFormData>({ ...emptyForm })
  const [fetchingFavicon, setFetchingFavicon] = useState(false)

  const { data: links, isLoading } = useFriendLinks(keyword)
  const createMutation = useCreateFriendLink()
  const updateMutation = useUpdateFriendLink()
  const deleteMutation = useDeleteFriendLink()
  const toggleMutation = useToggleLinkStatus()
  const { confirm, ConfirmDialog } = useConfirm()

  /** 打开新建弹窗 */
  function openCreate() {
    setEditingId(null)
    setFormData({ ...emptyForm })
    setDialogOpen(true)
  }

  /** 打开编辑弹窗 */
  function openEdit(link: FriendLink) {
    setEditingId(link.id)
    setFormData({
      name: link.name,
      url: link.url,
      description: link.description,
      logo: link.logo || '',
      sort: link.sort,
      status: link.status,
    })
    setDialogOpen(true)
  }

  /** 提交 */
  function handleSubmit() {
    if (!formData.name.trim() || !formData.url.trim()) return
    const onSuccess = () => { setDialogOpen(false); setEditingId(null); setFormData({ ...emptyForm }) }

    if (editingId) {
      updateMutation.mutate({ id: editingId, ...formData }, { onSuccess })
    } else {
      createMutation.mutate(formData, { onSuccess })
    }
  }

  async function handleDelete(id: string, name: string) {
    if (await confirm({ title: '删除友链', description: `确定删除友链「${name}」吗？`, confirmText: '删除', variant: 'destructive' })) {
      deleteMutation.mutate(id)
    }
  }

  function handleToggle(id: string, currentStatus: LinkStatus) {
    toggleMutation.mutate({ id, status: currentStatus === 'active' ? 'down' : 'active' })
  }

  const isPending = createMutation.isPending || updateMutation.isPending
  const activeCount = links?.filter((l) => l.status === 'active').length ?? 0

  return (
    <>
      <ConfirmDialog />
      <Header fixed>
        <HeaderSearch />
        <div className='ml-auto' />
      </Header>
      <Main>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50 tracking-tight">友链管理</h1>
          <p className="text-sm text-zinc-500 mt-1">管理博客友情链接</p>
        </div>
        <Button className="bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-md shadow-sm hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={openCreate}>
          <Plus className="w-4 h-4 mr-1.5" /> 新增友链
        </Button>
      </div>

      {/* 搜索 + 统计 */}
      <div className="flex justify-between items-center mb-4">
        <div className="relative max-w-sm"><Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" /><Input placeholder="搜索友链…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className={`pl-9 ${inputCls}`} /></div>
        <span className="text-xs text-zinc-500">共 {links?.length ?? 0} 条 · {activeCount} 条在线</span>
      </div>

      {/* 加载 */}
      {isLoading && <div className="text-center py-16"><Loader2 className="w-4 h-4 animate-spin text-zinc-400 mx-auto" /></div>}

      {/* 空状态 */}
      {!isLoading && links?.length === 0 && (
        <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm py-16">
          <div className="flex flex-col items-center text-center">
            <Link2 className="w-8 h-8 text-zinc-300 dark:text-zinc-600 mb-3" />
            <p className="text-sm text-zinc-700 dark:text-zinc-300">暂无友链</p>
            <p className="text-sm text-zinc-500 mt-1">点击上方按钮添加</p>
            <Button variant="outline" size="sm" className="mt-3 shadow-sm border-zinc-200 bg-white rounded-md" onClick={openCreate}>新增友链</Button>
          </div>
        </div>
      )}

      {/* 友链卡片 */}
      {links && links.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {links.map((link) => {
            const st = statusConfig[link.status]
            return (
              <div
                key={link.id}
                className={cn(
                  'group relative bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg p-4 hover:border-zinc-300 dark:hover:border-zinc-700 transition-all cursor-pointer',
                  link.status === 'down' && 'opacity-50'
                )}
                onClick={() => openEdit(link)}
              >
                <div className="flex items-start gap-3">
                  {link.logo ? (
                    <img src={link.logo} alt={link.name} className="w-10 h-10 rounded-xl object-cover bg-zinc-100 dark:bg-zinc-800 shrink-0" />
                  ) : (
                    <div className="w-10 h-10 rounded-xl bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center text-base font-semibold text-zinc-400 shrink-0">{link.name.charAt(0)}</div>
                  )}
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-1.5">
                      <p className="text-sm font-semibold text-zinc-800 dark:text-zinc-200 truncate">{link.name}</p>
                      <span className={cn('w-1.5 h-1.5 rounded-full shrink-0', link.status === 'active' ? 'bg-emerald-500' : link.status === 'pending' ? 'bg-amber-500' : 'bg-zinc-300 dark:bg-zinc-600')} title={st.text} />
                    </div>
                    <p className="text-xs text-zinc-400 mt-0.5 flex items-center gap-1 truncate">
                      <ExternalLink className="w-3 h-3 shrink-0" />{extractDomain(link.url)}
                    </p>
                    {link.description && <p className="text-xs text-zinc-500 mt-1.5 line-clamp-2">{link.description}</p>}
                  </div>
                </div>
                {/* hover 操作 */}
                <div className="absolute top-2 right-2 hidden group-hover:flex gap-0.5 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-md shadow-sm p-0.5" onClick={(e) => e.stopPropagation()}>
                  <Button variant="ghost" size="icon" className="w-7 h-7 rounded text-zinc-400 hover:text-zinc-700" onClick={() => window.open(link.url, '_blank')}>
                    <ExternalLink className="w-3.5 h-3.5" />
                  </Button>
                  <Button variant="ghost" size="icon" className="w-7 h-7 rounded text-zinc-400 hover:text-zinc-700" disabled={toggleMutation.isPending} onClick={() => handleToggle(link.id, link.status)}>
                    {link.status === 'active' ? <ArrowDownCircle className="w-3.5 h-3.5" /> : <ArrowUpCircle className="w-3.5 h-3.5" />}
                  </Button>
                  <Button variant="ghost" size="icon" className="w-7 h-7 rounded text-zinc-400 hover:text-red-500" disabled={deleteMutation.isPending} onClick={() => handleDelete(link.id, link.name)}>
                    <Trash2 className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* 新建/编辑弹窗 */}
      <Dialog open={dialogOpen} onOpenChange={(open) => { if (!open) { setDialogOpen(false); setEditingId(null); setFormData({ ...emptyForm }) } }}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingId ? '编辑友链' : '新增友链'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            {/* 名称 */}
            <div>
              <label className="text-xs font-medium text-zinc-500 mb-1.5 block">名称 *</label>
              <Input value={formData.name} onChange={(e) => setFormData((p) => ({ ...p, name: e.target.value }))} placeholder="例如：阮一峰的网络日志" className={inputCls} />
            </div>
            {/* 链接 */}
            <div>
              <label className="text-xs font-medium text-zinc-500 mb-1.5 block">链接 *</label>
              <Input value={formData.url} onChange={(e) => setFormData((p) => ({ ...p, url: e.target.value }))} placeholder="https://example.com" className={inputCls} />
            </div>
            {/* Logo URL + 自动获取 */}
            <div>
              <label className="text-xs font-medium text-zinc-500 mb-1.5 block flex items-center gap-1">
                <ImageIcon className="w-3 h-3" /> Logo URL
              </label>
              <div className="flex gap-2">
                <Input value={formData.logo} onChange={(e) => setFormData((p) => ({ ...p, logo: e.target.value }))} placeholder="https://example.com/avatar.png 或留空" className={`flex-1 ${inputCls}`} />
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-9 text-xs shrink-0 shadow-sm rounded-md border-zinc-200 dark:border-zinc-700"
                  disabled={!formData.url.trim() || fetchingFavicon}
                  onClick={async () => {
                    setFetchingFavicon(true)
                    try {
                      const res = await apiPost<{ favicon: string }>('/admin/utils/favicon', { url: formData.url.trim() })
                      if (res.favicon) {
                        setFormData((p) => ({ ...p, logo: res.favicon }))
                        toast.success(`已获取到 favicon: ${res.favicon}`)
                      } else {
                        toast.warning('未找到 favicon，请手动填写 Logo URL')
                      }
                    } catch {
                      toast.warning('未找到 favicon，请手动填写 Logo URL')
                    }
                    setFetchingFavicon(false)
                  }}
                >
                  {fetchingFavicon ? <Loader2 className="w-3 h-3 mr-1 animate-spin" /> : <Wand2 className="w-3 h-3 mr-1" />}
                  {fetchingFavicon ? '获取中…' : '自动获取'}
                </Button>
              </div>
              {/* Logo 预览 */}
              {formData.logo && (
                <div className="flex items-center gap-2 mt-2 p-2 bg-zinc-50 dark:bg-zinc-800/50 rounded-md">
                  <img src={formData.logo} alt="Logo 预览" className="w-6 h-6 rounded object-cover bg-white dark:bg-zinc-700 shrink-0" onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }} />
                  <span className="text-xs text-zinc-500 truncate">{formData.logo}</span>
                </div>
              )}
            </div>
            {/* 描述 */}
            <div>
              <label className="text-xs font-medium text-zinc-500 mb-1.5 block">描述</label>
              <Input value={formData.description} onChange={(e) => setFormData((p) => ({ ...p, description: e.target.value }))} placeholder="一句话介绍…" className={inputCls} />
            </div>
            {/* 排序 */}
            <div>
              <label className="text-xs font-medium text-zinc-500 mb-1.5 block">排序（数字越小越前）</label>
              <Input type="number" value={formData.sort} onChange={(e) => setFormData((p) => ({ ...p, sort: parseInt(e.target.value) || 0 }))} className={inputCls} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" className="shadow-sm rounded-md border-zinc-200 dark:border-zinc-700" onClick={() => setDialogOpen(false)}>取消</Button>
            <Button className="bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-md shadow-sm hover:bg-zinc-800" onClick={handleSubmit} disabled={!formData.name.trim() || !formData.url.trim() || isPending}>
              {isPending ? (editingId ? '保存中…' : '创建中…') : (editingId ? '保存' : '创建')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      </Main>
    </>
  )
}
