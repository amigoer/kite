import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import {
  Tooltip, TooltipContent, TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Search, Plus, Pencil, Trash2, ExternalLink, Files, Loader2 } from 'lucide-react'
import { usePageList, useDeletePage } from '@/hooks/use-pages'
import type { Page } from '@/types/page'

function formatDate(dateStr: string | null) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

/**
 * 独立页面管理列表
 */
export function PagesPage() {
  const navigate = useNavigate()
  const [keyword, setKeyword] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<Page | null>(null)

  const { data: pages, isLoading } = usePageList(keyword)
  const deleteMutation = useDeletePage()

  const publishedCount = pages?.filter((p) => p.status === 'published').length ?? 0
  const navCount = pages?.filter((p) => p.showInNav).length ?? 0

  return (
    <div>
      {/* 标题区 */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-lg font-semibold text-zinc-950 dark:text-zinc-50">页面管理</h1>
          <p className="text-sm text-zinc-500 mt-1">管理博客的独立页面，如关于我、归档、留言板等</p>
        </div>
        <Button className="bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950 shadow-none rounded-md hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={() => navigate('/pages/new')}>
          <Plus className="w-4 h-4 mr-1.5" /> 新建页面
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-3 gap-3 mb-6">
        {[
          { label: '页面总数', value: pages?.length ?? 0 },
          { label: '已发布', value: publishedCount },
          { label: '导航栏显示', value: navCount },
        ].map((s) => (
          <Card key={s.label} className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-none rounded-md">
            <CardContent className="p-4">
              <p className="text-xs text-zinc-500">{s.label}</p>
              <p className="text-2xl font-semibold text-zinc-950 dark:text-zinc-50 mt-1">{s.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 搜索 */}
      <div className="mb-4 max-w-sm">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
          <Input
            placeholder="搜索页面标题或 Slug…"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            className="pl-9 border-zinc-200 dark:border-zinc-800 bg-transparent shadow-none rounded-md"
          />
        </div>
      </div>

      {/* 页面列表 */}
      <div className="border border-zinc-200 dark:border-zinc-800 rounded-md bg-white dark:bg-zinc-900">
        <Table>
          <TableHeader>
            <TableRow className="border-b border-zinc-100 dark:border-zinc-800 hover:bg-transparent">
              <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider">页面标题</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[90px] text-center">状态</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[80px] text-center">排序</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[90px] text-center">导航栏</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[120px]">更新日期</TableHead>
              <TableHead className="text-xs font-medium text-zinc-500 uppercase tracking-wider w-[100px] text-center">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow><TableCell colSpan={6} className="text-center py-16"><Loader2 className="w-5 h-5 animate-spin text-zinc-400 mx-auto" /></TableCell></TableRow>
            ) : pages && pages.length > 0 ? (
              pages.map((p) => (
                <TableRow
                  key={p.id}
                  className="border-b border-zinc-100 dark:border-zinc-800 cursor-pointer hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
                  onDoubleClick={() => navigate(`/pages/${p.id}/edit`)}
                >
                  <TableCell onClick={() => navigate(`/pages/${p.id}/edit`)}>
                    <p className="text-sm font-medium text-zinc-950 dark:text-zinc-50">{p.title}</p>
                    <div className="flex items-center gap-2 mt-0.5">
                      <span className="text-xs text-zinc-500">/{p.slug}</span>
                      {p.showInNav && <Badge variant="outline" className="text-[10px] h-4 border-green-300 text-green-600 dark:border-green-700 dark:text-green-400">导航</Badge>}
                    </div>
                  </TableCell>
                  <TableCell className="text-center">
                    <Badge variant={p.status === 'published' ? 'default' : 'secondary'} className="text-xs">
                      {p.status === 'published' ? '已发布' : '草稿'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-center text-xs text-zinc-500">{p.sortOrder}</TableCell>
                  <TableCell className="text-center text-xs text-zinc-500">{p.showInNav ? '是' : '否'}</TableCell>
                  <TableCell className="text-xs text-zinc-500">{formatDate(p.updatedAt)}</TableCell>
                  <TableCell>
                    <div className="flex gap-1 justify-center" onClick={(e) => e.stopPropagation()}>
                      {p.status === 'published' && (
                        <Tooltip delayDuration={0}>
                          <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => window.open(`/pages/${p.slug}`, '_blank')}>
                              <ExternalLink className="w-3.5 h-3.5" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent side="top" className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm">预览</TooltipContent>
                        </Tooltip>
                      )}
                      <Tooltip delayDuration={0}>
                        <TooltipTrigger asChild>
                          <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => navigate(`/pages/${p.id}/edit`)}>
                            <Pencil className="w-3.5 h-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent side="top" className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm">编辑</TooltipContent>
                      </Tooltip>
                      <Tooltip delayDuration={0}>
                        <TooltipTrigger asChild>
                          <Button variant="ghost" size="icon" className="w-7 h-7 text-red-500 hover:text-red-600" onClick={() => setDeleteTarget(p)}>
                            <Trash2 className="w-3.5 h-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent side="top" className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm">删除</TooltipContent>
                      </Tooltip>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-16">
                  <Files className="w-10 h-10 text-zinc-300 dark:text-zinc-700 mx-auto mb-3" />
                  <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">暂无独立页面</p>
                  <p className="text-sm text-zinc-500 mt-1">点击上方按钮创建第一个页面</p>
                  <Button variant="outline" size="sm" className="mt-3 shadow-none border-zinc-200 dark:border-zinc-800" onClick={() => navigate('/pages/new')}>
                    新建页面
                  </Button>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* 删除确认 */}
      <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <AlertDialogContent className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm rounded-md">
          <AlertDialogHeader>
            <AlertDialogTitle>删除页面</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteTarget && `确定要删除页面「${deleteTarget.title}」吗？此操作不可撤销。`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel className="shadow-none border-zinc-200 dark:border-zinc-800">取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => { if (deleteTarget) { deleteMutation.mutate(deleteTarget.id); setDeleteTarget(null) } }} className="bg-red-500 text-white hover:bg-red-600 shadow-none">确认删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
