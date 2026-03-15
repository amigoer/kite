import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router'
import { ArrowLeft, Save, Send, Check } from 'lucide-react'
import { TiptapEditor } from '@/components/TiptapEditor'
import { usePostDetail, useSavePost, useCategories } from '@/hooks/use-posts'
import type { PostFormData } from '@/types/post'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

/**
 * 文章编辑页面
 * 支持新建（/posts/new）和编辑（/posts/:id/edit）两种模式
 */
export function PostEditorPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isEdit = !!id
  const [saved, setSaved] = useState(false)

  // 表单状态
  const [form, setForm] = useState<PostFormData>({
    title: '',
    slug: '',
    summary: '',
    content: '',
    category: '',
    tags: [],
    status: 'draft',
    coverUrl: '',
  })
  const [tagInput, setTagInput] = useState('')

  // 数据获取
  const { data: post, isLoading } = usePostDetail(id)
  const { data: categories } = useCategories()
  const saveMutation = useSavePost()

  // 编辑模式：填充表单
  useEffect(() => {
    if (post) {
      setForm({
        title: post.title,
        slug: post.slug,
        summary: post.summary,
        content: post.content,
        category: post.category,
        tags: [...post.tags],
        status: post.status,
        coverUrl: post.coverUrl,
      })
    }
  }, [post])

  /** 自动生成 slug */
  function handleTitleChange(title: string) {
    setForm((prev) => ({
      ...prev,
      title,
      slug: prev.slug || title.toLowerCase().replace(/[\s]+/g, '-').replace(/[^a-z0-9\u4e00-\u9fa5-]/g, ''),
    }))
  }

  /** 添加标签 */
  function addTag() {
    const tag = tagInput.trim()
    if (tag && !form.tags.includes(tag)) {
      setForm((prev) => ({ ...prev, tags: [...prev.tags, tag] }))
    }
    setTagInput('')
  }

  /** 移除标签 */
  function removeTag(tag: string) {
    setForm((prev) => ({ ...prev, tags: prev.tags.filter((t) => t !== tag) }))
  }

  /** 保存 */
  function handleSave(publish = false) {
    const data = { ...form, id }
    if (publish) data.status = 'published'
    saveMutation.mutate(data, {
      onSuccess: () => {
        setSaved(true)
        setTimeout(() => setSaved(false), 2000)
        if (!isEdit) navigate('/posts')
      },
    })
  }

  if (isEdit && isLoading) {
    return <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">加载中…</div>
  }

  return (
    <div>
      {/* 顶部操作栏 */}
      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => navigate('/posts')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">
              {isEdit ? '编辑文章' : '新建文章'}
            </h1>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {isEdit ? `正在编辑：${post?.title}` : '撰写新的博客文章'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={() => handleSave(false)} disabled={saveMutation.isPending}>
            {saved ? (
              <><Check className="mr-2 h-4 w-4" />已保存</>
            ) : (
              <><Save className="mr-2 h-4 w-4" />保存草稿</>
            )}
          </Button>
          <Button onClick={() => handleSave(true)} disabled={saveMutation.isPending || !form.title.trim()}>
            <Send className="mr-2 h-4 w-4" />
            发布文章
          </Button>
        </div>
      </div>

      <div className="flex gap-6">
        {/* 左侧：主编辑区 */}
        <div className="flex-1 space-y-4">
          {/* 标题输入 */}
          <Input
            value={form.title}
            onChange={(e) => handleTitleChange(e.target.value)}
            placeholder="文章标题"
            className="h-12 text-lg font-medium"
          />

          {/* 编辑器 */}
          <TiptapEditor
            content={form.content}
            onChange={(html) => setForm((prev) => ({ ...prev, content: html }))}
            placeholder="开始写作…"
          />
        </div>

        {/* 右侧：元数据面板 */}
        <div className="w-72 shrink-0 space-y-4">
          {/* Slug */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">URL Slug</CardTitle>
            </CardHeader>
            <CardContent>
              <Input
                value={form.slug}
                onChange={(e) => setForm((prev) => ({ ...prev, slug: e.target.value }))}
                placeholder="article-slug"
              />
            </CardContent>
          </Card>

          {/* 分类 */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">分类</CardTitle>
            </CardHeader>
            <CardContent>
              <Select value={form.category} onValueChange={(v) => setForm((prev) => ({ ...prev, category: v }))}>
                <SelectTrigger><SelectValue placeholder="选择分类" /></SelectTrigger>
                <SelectContent>
                  {categories?.map((c) => (
                    <SelectItem key={c} value={c}>{c}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </CardContent>
          </Card>

          {/* 标签 */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">标签</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2">
                <Input
                  value={tagInput}
                  onChange={(e) => setTagInput(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addTag())}
                  placeholder="输入标签…"
                  className="flex-1"
                />
                <Button variant="outline" size="sm" onClick={addTag} disabled={!tagInput.trim()}>
                  添加
                </Button>
              </div>
              {form.tags.length > 0 && (
                <div className="mt-3 flex flex-wrap gap-1.5">
                  {form.tags.map((tag) => (
                    <Badge key={tag} variant="secondary" className="cursor-pointer gap-1" onClick={() => removeTag(tag)}>
                      {tag} ×
                    </Badge>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* 摘要 */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">摘要</CardTitle>
            </CardHeader>
            <CardContent>
              <textarea
                value={form.summary}
                onChange={(e) => setForm((prev) => ({ ...prev, summary: e.target.value }))}
                placeholder="文章摘要…"
                rows={3}
                className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none"
              />
            </CardContent>
          </Card>

          {/* 封面图 */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">封面图</CardTitle>
            </CardHeader>
            <CardContent>
              <Input
                value={form.coverUrl}
                onChange={(e) => setForm((prev) => ({ ...prev, coverUrl: e.target.value }))}
                placeholder="https://example.com/cover.jpg"
              />
              {form.coverUrl && (
                <img src={form.coverUrl} alt="封面预览" className="mt-2 rounded-md border" />
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
