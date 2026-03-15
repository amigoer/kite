import { useState } from 'react'
import { Search, Plus, X, ExternalLink, Trash2, Link as LinkIcon, CheckCircle, Clock, AlertTriangle, Globe } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useFriendLinks, useCreateFriendLink, useDeleteFriendLink, useToggleLinkStatus } from '@/hooks/use-friend-links'
import type { LinkStatus } from '@/types/friend-link'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

/** 状态配置 */
const statusConfig: Record<LinkStatus, { text: string; variant: 'default' | 'secondary' | 'destructive' }> = {
  active: { text: '正常', variant: 'default' },
  pending: { text: '待审核', variant: 'secondary' },
  down: { text: '已下线', variant: 'destructive' },
}

/**
 * 友链管理页面 - 使用 shadcn Card / Input / Badge / Button / Separator
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
    createMutation.mutate(formData, {
      onSuccess: () => { setFormData({ name: '', url: '', description: '' }); setShowForm(false) },
    })
  }

  function handleDelete(id: string, name: string) {
    if (window.confirm(`确定删除友链「${name}」吗？`)) deleteMutation.mutate(id)
  }

  function handleToggle(id: string, currentStatus: LinkStatus) {
    toggleMutation.mutate({ id, status: currentStatus === 'active' ? 'down' : 'active' })
  }

  function extractDomain(url: string): string {
    try { return new URL(url).hostname } catch { return url }
  }

  const activeCount = links?.filter((l) => l.status === 'active').length ?? 0

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">友链管理</h1>
          <p className="mt-1 text-sm text-muted-foreground">管理博客友情链接，互换链接、共建生态</p>
        </div>
        <Button onClick={() => setShowForm(true)}>
          <Plus className="mr-2 h-4 w-4" />
          新增友链
        </Button>
      </div>

      {/* 新增表单 */}
      {showForm && (
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">新增友链</CardTitle>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setShowForm(false)}>
                <X className="h-4 w-4" />
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">名称 *</label>
                <Input value={formData.name} onChange={(e) => setFormData((p) => ({ ...p, name: e.target.value }))} placeholder="例如：阮一峰的网络日志" />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium text-muted-foreground">链接 *</label>
                <Input type="url" value={formData.url} onChange={(e) => setFormData((p) => ({ ...p, url: e.target.value }))} placeholder="https://example.com" />
              </div>
            </div>
            <div className="mt-4">
              <label className="mb-1.5 block text-xs font-medium text-muted-foreground">描述</label>
              <Input value={formData.description} onChange={(e) => setFormData((p) => ({ ...p, description: e.target.value }))} placeholder="一句话介绍这个博客…" />
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <Button variant="outline" onClick={() => setShowForm(false)}>取消</Button>
              <Button onClick={handleCreate} disabled={!formData.name.trim() || !formData.url.trim() || createMutation.isPending}>
                {createMutation.isPending ? '创建中…' : '创建'}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 搜索 + 统计 */}
      <div className="mb-4 flex items-center justify-between gap-4">
        <div className="relative max-w-sm flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="搜索友链…" value={keyword} onChange={(e) => setKeyword(e.target.value)} className="pl-9" />
        </div>
        <span className="text-xs text-muted-foreground">共 {links?.length ?? 0} 条友链 · {activeCount} 条在线</span>
      </div>

      {isLoading && <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">加载中…</div>}

      {!isLoading && links?.length === 0 && (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-16 text-muted-foreground">
            <LinkIcon className="mb-3 h-8 w-8" strokeWidth={1} />
            <p className="text-sm">暂无友链，点击上方按钮添加</p>
          </CardContent>
        </Card>
      )}

      {/* 友链卡片 */}
      {links && links.length > 0 && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {links.map((link) => {
            const st = statusConfig[link.status]
            return (
              <Card key={link.id} className={cn('group', link.status === 'down' && 'opacity-60')}>
                <CardContent className="pt-5">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-muted text-sm font-semibold">
                        {link.name.charAt(0)}
                      </div>
                      <div>
                        <h3 className="text-sm font-semibold">{link.name}</h3>
                        <div className="mt-0.5 flex items-center gap-1 text-xs text-muted-foreground">
                          <Globe className="h-3 w-3" strokeWidth={1.5} />{extractDomain(link.url)}
                        </div>
                      </div>
                    </div>
                    <Badge variant={st.variant}>{st.text}</Badge>
                  </div>
                  {link.description && <p className="mt-3 line-clamp-2 text-sm text-muted-foreground leading-relaxed">{link.description}</p>}
                </CardContent>
                <Separator />
                <div className="flex items-center justify-between px-5 py-2.5">
                  <span className="text-xs text-muted-foreground">#{link.sortOrder}</span>
                  <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                    <Button variant="outline" size="sm" asChild>
                      <a href={link.url} target="_blank" rel="noopener noreferrer">
                        <ExternalLink className="mr-1 h-3 w-3" />访问
                      </a>
                    </Button>
                    <Button variant="outline" size="sm" disabled={toggleMutation.isPending} onClick={() => handleToggle(link.id, link.status)}>
                      {link.status === 'active' ? <><AlertTriangle className="mr-1 h-3 w-3" />下线</> : <><CheckCircle className="mr-1 h-3 w-3" />上线</>}
                    </Button>
                    <Button variant="outline" size="sm" className="text-destructive hover:text-destructive" disabled={deleteMutation.isPending} onClick={() => handleDelete(link.id, link.name)}>
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
