import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Trash2, Copy, Loader2, ChevronLeft, ChevronRight, Image } from 'lucide-react'
import { apiGet, apiPost } from '@/lib/api-client'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search } from '@/components/search'
import { useConfirm } from '@/hooks/use-confirm'
import { toast } from 'sonner'

interface ImageItem {
  url: string
  filename: string
  size: number
  updatedAt: string
}

interface ImageListResponse {
  items: ImageItem[]
  total: number
  page: number
  pageSize: number
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

export function MediaPage() {
  const [page, setPage] = useState(1)
  const pageSize = 20
  const queryClient = useQueryClient()
  const { confirm, ConfirmDialog } = useConfirm()

  const { data, isLoading } = useQuery({
    queryKey: ['media', page],
    queryFn: () => apiGet<ImageListResponse>('/admin/upload/images', { page, pageSize }),
  })

  const deleteMutation = useMutation({
    mutationFn: (path: string) => apiPost('/admin/upload/images/delete', { path }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media'] })
      toast.success('删除成功')
    },
    onError: () => {
      toast.error('删除失败')
    },
  })

  async function handleDelete(filename: string) {
    if (await confirm({
      title: '删除图片',
      description: '确定删除此图片吗？已引用此图片的文章将无法显示。',
      confirmText: '删除',
      variant: 'destructive',
    })) {
      deleteMutation.mutate(filename)
    }
  }

  function handleCopy(url: string) {
    navigator.clipboard.writeText(window.location.origin + url)
    toast.success('链接已复制')
  }

  const totalPages = data ? Math.ceil(data.total / pageSize) : 1

  return (
    <>
      <ConfirmDialog />
      <Header fixed>
        <Search />
        <div className='ml-auto' />
      </Header>
      <Main>
        <div className='flex justify-between items-center mb-6'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>媒体库</h1>
            <p className='text-muted-foreground text-sm mt-1'>
              管理已上传的图片 {data ? `(${data.total} 个文件)` : ''}
            </p>
          </div>
        </div>

        {isLoading ? (
          <div className='flex items-center justify-center py-20'>
            <Loader2 className='w-6 h-6 animate-spin text-muted-foreground' />
          </div>
        ) : !data?.items?.length ? (
          <div className='text-center py-20 text-muted-foreground'>
            <Image className='w-12 h-12 mx-auto mb-4 opacity-50' />
            <p>暂无上传的图片</p>
          </div>
        ) : (
          <>
            <div className='grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4'>
              {data.items.map((img) => (
                <div
                  key={img.filename}
                  className='group border rounded-lg overflow-hidden bg-card hover:shadow-md transition-shadow'
                >
                  <div className='aspect-square bg-muted flex items-center justify-center overflow-hidden'>
                    <img
                      src={img.url}
                      alt={img.filename}
                      className='w-full h-full object-cover'
                      loading='lazy'
                    />
                  </div>
                  <div className='p-2'>
                    <p className='text-xs text-muted-foreground truncate' title={img.filename}>
                      {img.filename.split('/').pop()}
                    </p>
                    <div className='flex items-center justify-between mt-1'>
                      <span className='text-xs text-muted-foreground'>
                        {formatSize(img.size)} &middot; {formatDate(img.updatedAt)}
                      </span>
                      <div className='flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity'>
                        <Button
                          variant='ghost'
                          size='icon'
                          className='h-6 w-6'
                          onClick={() => handleCopy(img.url)}
                          title='复制链接'
                        >
                          <Copy className='w-3 h-3' />
                        </Button>
                        <Button
                          variant='ghost'
                          size='icon'
                          className='h-6 w-6 text-destructive hover:text-destructive'
                          onClick={() => handleDelete(img.filename)}
                          title='删除'
                        >
                          <Trash2 className='w-3 h-3' />
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {totalPages > 1 && (
              <div className='flex justify-center items-center gap-2 mt-6'>
                <Button
                  variant='outline'
                  size='sm'
                  disabled={page <= 1}
                  onClick={() => setPage(p => p - 1)}
                >
                  <ChevronLeft className='w-4 h-4' />
                </Button>
                <span className='text-sm text-muted-foreground'>
                  {page} / {totalPages}
                </span>
                <Button
                  variant='outline'
                  size='sm'
                  disabled={page >= totalPages}
                  onClick={() => setPage(p => p + 1)}
                >
                  <ChevronRight className='w-4 h-4' />
                </Button>
              </div>
            )}
          </>
        )}
      </Main>
    </>
  )
}
