import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Search, Plus, Trash2, ExternalLink, Globe, Link2, X, Loader2, ArrowUpCircle, ArrowDownCircle } from 'lucide-react'
import { useFriendLinks, useCreateFriendLink, useDeleteFriendLink, useToggleLinkStatus } from '@/hooks/use-friend-links'
import type { LinkStatus } from '@/types/friend-link'
import { cn } from '@/lib/utils'

const statusConfig: Record<LinkStatus, { text: string }> = {
  active: { text: '正常' },
  pending: { text: '待审核' },
  down: { text: '已下线' },
}

function extractDomain(url: string): string {
  try { return new URL(url).hostname } catch { return url }
}

const inputCls = 'border-zinc-200 dark:border-zinc-700 bg-transparent rounded-md shadow-none focus-visible:ring-1 focus-visible:ring-zinc-400'

/**
 * 友链管理 — Vercel 风格
 */
export function FriendLinksPage() {
  const [keyword, setKeyword] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({ name: '', url: '', description: '' })

  const { data: links, isLoading } = useFriendLinks(keyword)
  const createMutation = useCreateFriendLink()
  const deleteMutation = useDeleteFriendLink()
  const toggleMutation = useToggleLinkStatus()

  function handleCreate() {
    if (!formData.name.trim() || !formData.url.trim()) return
    createMutation.mutate(formData, { onSuccess: () => { setFormData({ name: '', url: '', description: '' }); setShowForm(false) } })
  }
  function handleDelete(id: string, name: string) { if (window.confirm(`确定删除友链「${name}」吗？`)) deleteMutation.mutate(id) }
  function handleToggle(id: string, currentStatus: LinkStatus) { toggleMutation.mutate({ id, status: currentStatus === 'active' ? 'down' : 'active' }) }

  const activeCount = links?.filter((l) => l.status === 'active').length ?? 0

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50 tracking-tight">友链管理</h1>
          <p className="text-sm text-zinc-500 mt-1">管理博客友情链接</p>
        </div>
        <Button className="bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-md shadow-sm hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={() => setShowForm(true)}>
          <Plus className="w-4 h-4 mr-1.5" /> 新增友链
        </Button>
      </div>

      {showForm && (
        <div className="mb-6 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm p-5">
          <div className="flex justify-between items-center mb-4">
            <p className="text-sm font-medium text-zinc-700 dark:text-zinc-300">新增友链</p>
            <Button variant="ghost" size="icon" className="w-7 h-7 rounded-md" onClick={() => setShowForm(false)}><X className="w-4 h-4" /></Button>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div><label className="text-xs font-medium text-zinc-500 mb-1.5 block">名称 *</label><Input value={formData.name} onChange={(e) => setFormData((p) => ({ ...p, name: e.target.value }))} placeholder="例如：阮一峰的网络日志" className={inputCls} /></div>
            <div><label className="text-xs font-medium text-zinc-500 mb-1.5 block">链接 *</label><Input value={formData.url} onChange={(e) => setFormData((p) => ({ ...p, url: e.target.value }))} placeholder="https://example.com" className={inputCls} /></div>
          </div>
          <div className="mt-4"><label className="text-xs font-medium text-zinc-500 mb-1.5 block">描述</label><Input value={formData.description} onChange={(e) => setFormData((p) => ({ ...p, description: e.target.value }))} placeholder="一句话介绍…" className={inputCls} /></div>
          <div className="mt-4 flex justify-end gap-2">
            <Button variant="outline" className="shadow-sm rounded-md border-zinc-200 dark:border-zinc-700 bg-white" onClick={() => setShowForm(false)}>取消</Button>
            <Button className="bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-md shadow-sm hover:bg-zinc-800" onClick={handleCreate} disabled={!formData.name.trim() || !formData.url.trim() || createMutation.isPending}>
              {createMutation.isPending ? '创建中…' : '创建'}
            </Button>
          </div>
        </div>
      )}

      <div className="flex justify-between items-center mb-4">
        <div className="relative max-w-sm"><Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" /><Input placeholder="搜索友链…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className={`pl-9 ${inputCls}`} /></div>
        <span className="text-xs text-zinc-500">共 {links?.length ?? 0} 条 · {activeCount} 条在线</span>
      </div>

      {isLoading && <div className="text-center py-16"><Loader2 className="w-4 h-4 animate-spin text-zinc-400 mx-auto" /></div>}

      {!isLoading && links?.length === 0 && (
        <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm py-16">
          <div className="flex flex-col items-center text-center">
            <Link2 className="w-8 h-8 text-zinc-300 dark:text-zinc-600 mb-3" />
            <p className="text-sm text-zinc-700 dark:text-zinc-300">暂无友链</p>
            <p className="text-sm text-zinc-500 mt-1">点击上方按钮添加</p>
            <Button variant="outline" size="sm" className="mt-3 shadow-sm border-zinc-200 bg-white rounded-md" onClick={() => setShowForm(true)}>新增友链</Button>
          </div>
        </div>
      )}

      {links && links.length > 0 && (
        <div className="grid grid-cols-2 gap-4">
          {links.map((link) => {
            const st = statusConfig[link.status]
            return (
              <div key={link.id} className={cn('bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm p-5 hover:border-zinc-300 dark:hover:border-zinc-700 transition-colors', link.status === 'down' && 'opacity-60')}>
                <div className="flex justify-between items-start">
                  <div className="flex gap-3 items-center">
                    <div className="w-8 h-8 rounded-full bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center text-sm font-medium text-zinc-600 dark:text-zinc-400">{link.name.charAt(0)}</div>
                    <div>
                      <p className="text-sm font-medium text-zinc-700 dark:text-zinc-300">{link.name}</p>
                      <p className="text-xs text-zinc-500 flex items-center gap-1 mt-0.5"><Globe className="w-3 h-3" /> {extractDomain(link.url)}</p>
                    </div>
                  </div>
                  <span className="bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 border border-zinc-200 dark:border-zinc-700 rounded-full px-2.5 py-0.5 text-xs font-medium">{st.text}</span>
                </div>
                {link.description && <p className="text-sm text-zinc-500 mt-3">{link.description}</p>}
                <Separator className="my-3 bg-zinc-100 dark:bg-zinc-800" />
                <div className="flex justify-between items-center">
                  <span className="text-xs text-zinc-500">#{link.sort}</span>
                  <div className="flex gap-1">
                    <Button variant="outline" size="sm" className="h-7 text-xs shadow-sm rounded-md border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50" onClick={() => window.open(link.url, '_blank')}><ExternalLink className="w-3 h-3 mr-1" /> 访问</Button>
                    <Button variant="outline" size="sm" className="h-7 text-xs shadow-sm rounded-md border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50" disabled={toggleMutation.isPending} onClick={() => handleToggle(link.id, link.status)}>
                      {link.status === 'active' ? <><ArrowDownCircle className="w-3 h-3 mr-1" /> 下线</> : <><ArrowUpCircle className="w-3 h-3 mr-1" /> 上线</>}
                    </Button>
                    <Button variant="ghost" size="icon" className="w-7 h-7 rounded-md text-zinc-400 hover:text-red-500" disabled={deleteMutation.isPending} onClick={() => handleDelete(link.id, link.name)}><Trash2 className="w-3.5 h-3.5" /></Button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
