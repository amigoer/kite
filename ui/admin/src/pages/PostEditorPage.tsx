import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import {
  Popover, PopoverContent, PopoverTrigger,
} from '@/components/ui/popover'
import { CategoryCascader, buildCascaderTree } from '@/components/category-cascader'
import { Calendar } from '@/components/ui/calendar'
import { ArrowLeft, Save, Send, Check, X, Loader2, Sparkles, Clock } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { TiptapEditor } from '@/components/TiptapEditor'
import { usePostDetail, useSavePost } from '@/hooks/use-posts'
import { useCategoryList } from '@/hooks/use-categories'
import { useTagList } from '@/hooks/use-tags'
import { ImageUploader } from '@/components/ImageUploader'
import type { PostFormData } from '@/types/post'
import { toast } from 'sonner'


/**
 * 文章编辑器 — Vercel 风格
 */
export function PostEditorPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isEdit = !!id
  const [saved, setSaved] = useState(false)
  const [aiSummaryLoading, setAiSummaryLoading] = useState(false)
  const [aiSummary, setAiSummary] = useState('')
  const [aiKeywordsLoading, setAiKeywordsLoading] = useState(false)
  const [aiKeywords, setAiKeywords] = useState<string[]>([])

  const [form, setForm] = useState<PostFormData>({
    title: '', slug: '', summary: '', contentMarkdown: '', contentHtml: '',
    categoryId: '', tagIds: [], status: 'draft', coverImage: '', password: '',
  })

  const { data: post, isLoading } = usePostDetail(id)
  const { data: categories } = useCategoryList()
  const { data: allTags } = useTagList()
  const saveMutation = useSavePost()

  useEffect(() => {
    if (post) {
      setForm({
        title: post.title, slug: post.slug, summary: post.summary,
        contentMarkdown: post.contentMarkdown || '', contentHtml: post.contentHtml || '',
        categoryId: post.categoryId || '', tagIds: post.tags?.map((t) => t.id) || [],
        status: post.status, coverImage: post.coverImage || '',
        password: (post as unknown as { password?: string })?.password || '',
      })
    }
  }, [post])

  function handleTitleChange(title: string) {
    setForm((prev) => ({ ...prev, title, slug: prev.slug || title.toLowerCase().replace(/[\s]+/g, '-').replace(/[^a-z0-9\u4e00-\u9fa5-]/g, '') }))
  }
  function removeTag(tagId: string) { setForm((prev) => ({ ...prev, tagIds: prev.tagIds.filter((id) => id !== tagId) })) }
  function getTagName(tagId: string): string { return allTags?.find((t) => t.id === tagId)?.name || tagId }
  function handleSave(publish = false, schedule = false) {
    const data = { ...form, id }
    if (schedule && form.publishAt) {
      data.status = 'scheduled'
    } else if (publish) {
      data.status = 'published'
    }
    saveMutation.mutate(data, {
      onSuccess: () => {
        setSaved(true); setTimeout(() => setSaved(false), 2000)
        if (schedule) {
          toast.success('定时发布已设定', { description: `「${form.title}」将在指定时间自动发布` })
        } else if (publish) {
          toast.success('文章已发布', { description: `「${form.title}」已成功发布` })
        } else {
          toast.success('草稿已保存')
        }
        if (!isEdit) navigate('/posts')
      },
      onError: (err) => {
        toast.error(publish ? '发布失败' : '保存失败', { description: err.message || '请稍后重试' })
      },
    })
  }
  function handleAiSummary() {
    setAiSummaryLoading(true); setAiSummary('')
    setTimeout(() => { setAiSummary('Kite 是一个轻量级 Go 博客引擎，内置 AI 写作助手、富文本编辑器和 CSR 主题，提供极致极简的写作体验…'); setAiSummaryLoading(false) }, 800)
  }
  function handleAiKeywords() {
    setAiKeywordsLoading(true); setAiKeywords([])
    setTimeout(() => { setAiKeywords(['博客', '技术', 'Go', 'React', '静态网站', '极简风']); setAiKeywordsLoading(false) }, 800)
  }

  if (isEdit && isLoading) return <div className="flex justify-center py-16"><Loader2 className="w-4 h-4 animate-spin text-zinc-400" /></div>

  return (
    <>
      <Header fixed>
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" className="w-8 h-8 rounded-md" onClick={() => navigate('/posts')}><ArrowLeft className="w-4 h-4" /></Button>
          <div className="hidden sm:block">
            <h1 className="text-sm font-semibold text-zinc-900 dark:text-zinc-50 tracking-tight">{isEdit ? '编辑文章' : '新建文章'}</h1>
            <p className="text-xs text-zinc-500">{isEdit ? `正在编辑：${post?.title}` : '撰写新的博客文章'}</p>
          </div>
        </div>
        <div className="flex gap-2 ml-auto">
          <Button variant="outline" className="shadow-sm border-zinc-200 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50" onClick={() => handleSave(false)} disabled={saveMutation.isPending}>
            {saved ? <><Check className="w-4 h-4 sm:mr-1.5" /> <span className="hidden sm:inline">已保存</span></> : <><Save className="w-4 h-4 sm:mr-1.5" /> <span className="hidden sm:inline">保存草稿</span></>}
          </Button>
          {/* 定时发布 */}
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" className="shadow-sm border-zinc-200 dark:border-zinc-700 rounded-md bg-white dark:bg-zinc-900 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50" disabled={saveMutation.isPending || !form.title.trim()}>
                <Clock className="w-4 h-4 sm:mr-1.5" /> <span className="hidden sm:inline">定时发布</span>
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="end">
              <div className="p-3 space-y-3">
                <p className="text-sm font-medium px-1">选择发布时间</p>
                <Calendar
                  mode="single"
                  selected={form.publishAt ? new Date(form.publishAt) : undefined}
                  onSelect={(date: Date | undefined) => {
                    if (!date) return
                    const prev = form.publishAt ? new Date(form.publishAt) : new Date()
                    date.setHours(prev.getHours(), prev.getMinutes())
                    setForm((p) => ({ ...p, publishAt: date.toISOString() }))
                  }}
                  disabled={(date: Date) => date < new Date(new Date().setHours(0, 0, 0, 0))}
                  initialFocus
                />
                <div className="flex items-center gap-2 px-1">
                  <Clock className="w-4 h-4 text-muted-foreground shrink-0" />
                  <Input
                    type="time"
                    value={form.publishAt ? `${String(new Date(form.publishAt).getHours()).padStart(2, '0')}:${String(new Date(form.publishAt).getMinutes()).padStart(2, '0')}` : ''}
                    onChange={(e) => {
                      const [h, m] = e.target.value.split(':').map(Number)
                      const d = form.publishAt ? new Date(form.publishAt) : new Date()
                      d.setHours(h, m)
                      setForm((p) => ({ ...p, publishAt: d.toISOString() }))
                    }}
                    className="flex-1 border-zinc-200 dark:border-zinc-700 shadow-none rounded-md text-sm h-9"
                  />
                </div>
                {form.publishAt && (
                  <p className="text-xs text-muted-foreground px-1">
                    文章将在 {new Date(form.publishAt).toLocaleString('zh-CN')} 自动发布
                  </p>
                )}
                <Button
                  className="w-full"
                  disabled={!form.publishAt || saveMutation.isPending}
                  onClick={() => handleSave(false, true)}
                >
                  <Clock className="w-4 h-4 mr-1.5" /> 设定定时发布
                </Button>
              </div>
            </PopoverContent>
          </Popover>
          <Button className="bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-md shadow-sm hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={() => handleSave(true)} disabled={saveMutation.isPending || !form.title.trim()}>
            <Send className="w-4 h-4 sm:mr-1.5" /> <span className="hidden sm:inline">发布文章</span>
          </Button>
        </div>
      </Header>

      <Main>
      <div className="flex gap-0 relative">
        <div className="flex-1 flex flex-col gap-4 pr-6">
          {/* 沉浸式标题 — 原生 input，零 border */}
          <div className="mb-2">
            <input
              type="text"
              value={form.title}
              onChange={(e) => handleTitleChange(e.target.value)}
              placeholder="输入文章标题..."
              className="w-full bg-white dark:bg-zinc-900 text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50 placeholder:text-zinc-400 dark:placeholder:text-zinc-600 border border-zinc-200 dark:border-zinc-700 rounded-lg outline-none focus:ring-2 focus:ring-zinc-900/10 dark:focus:ring-zinc-100/10 focus:border-zinc-300 dark:focus:border-zinc-600 px-4 py-3 shadow-sm transition-colors"
            />
          </div>
          <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm overflow-hidden">
            <TiptapEditor key={post?.id || 'new'} content={post ? (post.contentHtml || post.contentMarkdown || '') : ''} onChange={(html, markdown) => setForm((prev) => ({ ...prev, contentHtml: html, contentMarkdown: markdown }))} placeholder="开始写作…" />
          </div>
        </div>

        {/* 右侧属性面板：无外边框，纯白底色 + border-l 分割 */}
        <aside className="w-80 shrink-0 bg-white dark:bg-zinc-900 flex flex-col gap-6 p-6 border-l border-zinc-100 dark:border-zinc-800">

          {/* URL Slug */}
          <div className="flex flex-col gap-2 pb-6 border-b border-zinc-100 dark:border-zinc-800">
            <label className="text-xs font-semibold text-zinc-500 uppercase tracking-wider">URL Slug</label>
            <Input value={form.slug} onChange={(e) => setForm((prev) => ({ ...prev, slug: e.target.value }))} placeholder="article-slug" className="shadow-none rounded-md border-zinc-200 dark:border-zinc-700" />
          </div>

          {/* 分类 */}
          <div className="flex flex-col gap-2 pb-6 border-b border-zinc-100 dark:border-zinc-800">
            <label className="text-xs font-semibold text-zinc-500 uppercase tracking-wider">分类</label>
            <CategoryCascader
              options={buildCascaderTree(categories || [])}
              value={form.categoryId || null}
              onChange={(id) => setForm((prev) => ({ ...prev, categoryId: id || '' }))}
              allowSelectParent
            />
          </div>

          {/* 标签 */}
          <div className="flex flex-col gap-2 pb-6 border-b border-zinc-100 dark:border-zinc-800">
            <label className="text-xs font-semibold text-zinc-500 uppercase tracking-wider">标签</label>
            <Select key={form.tagIds.length} value="" onValueChange={(v) => { if (v && !form.tagIds.includes(v)) setForm((prev) => ({ ...prev, tagIds: [...prev.tagIds, v] })) }}><SelectTrigger className="shadow-none rounded-md border-zinc-200 dark:border-zinc-700"><SelectValue placeholder="选择标签…" /></SelectTrigger><SelectContent>{allTags?.filter((t) => !form.tagIds.includes(t.id)).map((t) => <SelectItem key={t.id} value={t.id}>{t.name}</SelectItem>)}</SelectContent></Select>
            {form.tagIds.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mt-1">
                {form.tagIds.map((tagId) => <Badge key={tagId} variant="secondary" className="text-xs gap-1 rounded-md bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 border border-zinc-200 dark:border-zinc-700">{getTagName(tagId)}<button className="cursor-pointer" onClick={() => removeTag(tagId)}><X className="w-3 h-3" /></button></Badge>)}
              </div>
            )}
          </div>

          {/* 摘要 + AI */}
          <div className="flex flex-col gap-2 pb-6 border-b border-zinc-100 dark:border-zinc-800">
            <div className="flex justify-between items-center">
              <label className="text-xs font-semibold text-zinc-500 uppercase tracking-wider">摘要</label>
              <button className="text-xs text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-200 flex items-center gap-1 cursor-pointer transition-colors" onClick={handleAiSummary} disabled={aiSummaryLoading}><Sparkles className="w-3 h-3" /> {aiSummaryLoading ? '…' : 'AI 生成'}</button>
            </div>
            <Textarea value={form.summary} onChange={(e) => setForm((prev) => ({ ...prev, summary: e.target.value }))} placeholder="文章摘要…" rows={3} className="shadow-none rounded-md border-zinc-200 dark:border-zinc-700 resize-none" />
            {aiSummary && (<div className="mt-1"><p className="text-xs text-zinc-500 leading-relaxed">{aiSummary}</p><button className="text-xs text-zinc-900 dark:text-zinc-100 underline underline-offset-2 cursor-pointer mt-1 hover:no-underline" onClick={() => { setForm((prev) => ({ ...prev, summary: aiSummary })); setAiSummary('') }}>应用此摘要</button></div>)}
          </div>

          {/* SEO 关键词 */}
          <div className="flex flex-col gap-2 pb-6 border-b border-zinc-100 dark:border-zinc-800">
            <label className="text-xs font-semibold text-zinc-500 uppercase tracking-wider">SEO 关键词</label>
            <div className="flex gap-2"><Input value={(form as unknown as { keywords?: string }).keywords || ''} onChange={(e) => setForm((prev) => ({ ...prev, keywords: e.target.value } as unknown as PostFormData))} placeholder="博客,技术,Go" className="flex-1 shadow-none rounded-md border-zinc-200 dark:border-zinc-700" /><Button variant="outline" size="sm" className="rounded-md shadow-none border-zinc-200 dark:border-zinc-700 text-zinc-700 dark:text-zinc-300 shrink-0 gap-1 text-xs h-9 hover:bg-zinc-50 dark:hover:bg-zinc-800" onClick={handleAiKeywords} disabled={aiKeywordsLoading}><Sparkles className="w-3 h-3" /> {aiKeywordsLoading ? '…' : 'AI'}</Button></div>
            {aiKeywords.length > 0 && <p className="text-xs text-zinc-500 mt-1">建议：<span className="text-zinc-700 dark:text-zinc-300">{aiKeywords.join(', ')}</span></p>}
          </div>

          {/* 封面图 */}
          <div className="flex flex-col gap-2 pb-6 border-b border-zinc-100 dark:border-zinc-800">
            <label className="text-xs font-semibold text-zinc-500 uppercase tracking-wider">封面图</label>
            <ImageUploader value={form.coverImage} onChange={(url) => setForm((prev) => ({ ...prev, coverImage: url }))} placeholder="上传封面图片" />
          </div>

          {/* 文章密码 */}
          <div className="flex flex-col gap-2">
            <label className="text-xs font-semibold text-zinc-500 uppercase tracking-wider">🔒 文章密码</label>
            <Input type="password" value={form.password} onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))} placeholder="留空则不加密" className="shadow-none rounded-md border-zinc-200 dark:border-zinc-700" />
            <p className="text-xs text-zinc-500">设置后文章需密码查看</p>
          </div>

        </aside>
      </div>
      </Main>
    </>
  )
}
