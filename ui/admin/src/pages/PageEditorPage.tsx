import { useState, useEffect, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { ArrowLeft, Save, Send, Check, Loader2, ExternalLink } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { TiptapEditor } from '@/components/TiptapEditor'
import { usePageDetail, useSavePage } from '@/hooks/use-pages'
import type { PageFormData } from '@/types/page'

/** 模板配置字段定义 */
interface TemplateField {
  key: string
  label: string
  description?: string
  type: 'text' | 'number' | 'switch'
  defaultValue: string | number | boolean
  placeholder?: string
}

const TEMPLATE_FIELDS: Record<string, TemplateField[]> = {
  default: [],
  github: [
    { key: 'username', label: 'GitHub 用户名', type: 'text', defaultValue: '', placeholder: '如：octocat' },
    { key: 'count', label: '展示仓库数量', description: '按 Star 数排序取前 N 个', type: 'number', defaultValue: 6 },
    { key: 'show_fork', label: '显示 Fork 仓库', description: '是否包含 Fork 的仓库', type: 'switch', defaultValue: false },
  ],
}

const PAGE_TEMPLATES = [
  { value: 'default', label: '默认模板' },
  { value: 'github', label: 'GitHub 风格' },
]

/**
 * 独立页面编辑器
 */
export function PageEditorPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isEdit = !!id
  const [saved, setSaved] = useState(false)

  const [form, setForm] = useState<PageFormData>({
    title: '', slug: '', contentMarkdown: '',
    status: 'draft', sortOrder: 10, showInNav: false,
    template: 'default', config: '',
  })

  const { data: page, isLoading } = usePageDetail(id)
  const saveMutation = useSavePage()

  const configObj = useMemo<Record<string, unknown>>(() => {
    if (!form.config) return {}
    try { return JSON.parse(form.config) } catch { return {} }
  }, [form.config])

  function updateConfigField(key: string, value: unknown) {
    const newConfig = { ...configObj, [key]: value }
    setForm((prev) => ({ ...prev, config: JSON.stringify(newConfig) }))
  }

  const currentFields = TEMPLATE_FIELDS[form.template] || []

  function handleTemplateChange(template: string) {
    const fields = TEMPLATE_FIELDS[template] || []
    const defaults: Record<string, unknown> = {}
    fields.forEach((f) => { defaults[f.key] = configObj[f.key] !== undefined ? configObj[f.key] : f.defaultValue })
    setForm((prev) => ({ ...prev, template, config: fields.length > 0 ? JSON.stringify(defaults) : '' }))
  }

  useEffect(() => {
    if (page) {
      setForm({
        title: page.title, slug: page.slug, contentMarkdown: page.contentMarkdown || '',
        status: page.status, sortOrder: page.sortOrder, showInNav: page.showInNav,
        template: page.template || 'default', config: page.config || '',
      })
    }
  }, [page])

  function handleTitleChange(title: string) {
    setForm((prev) => ({
      ...prev, title,
      slug: prev.slug || title.toLowerCase().replace(/[\s]+/g, '-').replace(/[^a-z0-9\u4e00-\u9fa5-]/g, ''),
    }))
  }

  function handleSave(publish = false) {
    const data = { ...form, id }
    if (publish) data.status = 'published'
    saveMutation.mutate(data, {
      onSuccess: (savedPage) => { 
        setSaved(true); setTimeout(() => setSaved(false), 2000)
        if (!isEdit) {
          navigate('/pages')
        } else if (savedPage?.slug && savedPage.slug !== id) {
          navigate(`/pages/${savedPage.slug}/edit`, { replace: true })
        }
      },
    })
  }

  if (isEdit && isLoading) {
    return <div className="flex justify-center py-16"><Loader2 className="w-5 h-5 animate-spin text-zinc-400" /></div>
  }

  return (
    <>
      <Header fixed>
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" className="w-8 h-8 rounded-md" onClick={() => navigate('/pages')}>
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div className="hidden sm:block">
            <h1 className="text-sm font-semibold text-zinc-950 dark:text-zinc-50 tracking-tight">{isEdit ? '编辑页面' : '新建页面'}</h1>
            <p className="text-xs text-zinc-500">{isEdit ? `正在编辑：${page?.title}` : '创建独立页面'}</p>
          </div>
        </div>
        <div className="flex gap-2 ml-auto">
          <Button variant="outline" className="shadow-none border-zinc-200 dark:border-zinc-800 rounded-md" onClick={() => handleSave(false)} disabled={saveMutation.isPending}>
            {saved ? <><Check className="w-4 h-4 sm:mr-1.5" /> <span className="hidden sm:inline">已保存</span></> : <><Save className="w-4 h-4 sm:mr-1.5" /> <span className="hidden sm:inline">保存草稿</span></>}
          </Button>
          <Button className="bg-zinc-950 dark:bg-zinc-50 text-white dark:text-zinc-950 shadow-none rounded-md hover:bg-zinc-800 dark:hover:bg-zinc-200" onClick={() => handleSave(true)} disabled={saveMutation.isPending || !form.title.trim()}>
            <Send className="w-4 h-4 sm:mr-1.5" /> <span className="hidden sm:inline">发布页面</span>
          </Button>
        </div>
      </Header>

      <Main>
      <div className="flex gap-6 relative">
        {/* 左侧编辑区 */}
        <div className="flex-1 flex flex-col gap-4">
          <Input value={form.title} onChange={(e) => handleTitleChange(e.target.value)} placeholder="页面标题" className="text-xl font-bold border-border bg-card shadow-sm rounded-lg h-14" />
          <TiptapEditor content={form.contentMarkdown} onChange={(html) => setForm((prev) => ({ ...prev, contentMarkdown: html }))} placeholder="开始编写页面内容…" />
        </div>

        {/* 右侧元数据面板 */}
        <div className="w-72 flex flex-col gap-4 shrink-0">
          <Card className="shadow-sm">
            <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-zinc-500 uppercase">URL Slug</CardTitle></CardHeader>
            <CardContent className="pt-0">
              <Input value={form.slug} onChange={(e) => setForm((prev) => ({ ...prev, slug: e.target.value }))} placeholder="page-slug" />
              <a
                href={`/pages/${form.slug || 'page-slug'}`}
                target="_blank"
                rel="noreferrer"
                className="text-xs text-muted-foreground hover:text-primary mt-1.5 inline-flex items-center gap-1 w-fit transition-colors"
              >
                访问页面：{window.location.origin}/pages/{form.slug || 'page-slug'}
                <ExternalLink className="w-3 h-3" />
              </a>
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-zinc-500 uppercase">页面模板</CardTitle></CardHeader>
            <CardContent className="pt-0">
              <Select value={form.template} onValueChange={handleTemplateChange}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent className="shadow-sm">
                  {PAGE_TEMPLATES.map((t) => <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>)}
                </SelectContent>
              </Select>
              <p className="text-xs text-zinc-500 mt-1.5">模板文件：pages/{form.template || 'default'}.html</p>
            </CardContent>
          </Card>

          {/* 模板参数 */}
          {currentFields.length > 0 && (
            <Card className="shadow-sm">
              <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-zinc-500 uppercase">模板参数</CardTitle></CardHeader>
              <CardContent className="pt-0 space-y-4">
                {currentFields.map((field) => (
                  <div key={field.key}>
                    {field.type === 'switch' ? (
                      <div className="flex justify-between items-center">
                        <div>
                          <p className="text-sm font-medium text-zinc-950 dark:text-zinc-50">{field.label}</p>
                          {field.description && <p className="text-xs text-zinc-500">{field.description}</p>}
                        </div>
                        <Switch checked={Boolean(configObj[field.key] ?? field.defaultValue)} onCheckedChange={(v) => updateConfigField(field.key, v)} />
                      </div>
                    ) : (
                      <>
                        <label className="text-sm font-medium text-zinc-950 dark:text-zinc-50 block mb-1.5">{field.label}</label>
                        <Input
                          type={field.type === 'number' ? 'number' : 'text'}
                          value={String(configObj[field.key] ?? field.defaultValue)}
                          onChange={(e) => updateConfigField(field.key, field.type === 'number' ? Number(e.target.value) : e.target.value)}
                          placeholder={field.placeholder}
                          className="bg-card"
                        />
                        {field.description && <p className="text-xs text-zinc-500 mt-1">{field.description}</p>}
                      </>
                    )}
                  </div>
                ))}
              </CardContent>
            </Card>
          )}

          <Card className="shadow-sm">
            <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-zinc-500 uppercase">页面设置</CardTitle></CardHeader>
            <CardContent className="pt-0 space-y-4">
              <div className="flex justify-between items-center">
                <div>
                  <p className="text-sm font-medium text-zinc-950 dark:text-zinc-50">显示在导航栏</p>
                  <p className="text-xs text-zinc-500">前台顶部导航显示此页面</p>
                </div>
                <Switch checked={form.showInNav} onCheckedChange={(v) => setForm((prev) => ({ ...prev, showInNav: v }))} />
              </div>
              <div>
                <label className="text-sm font-medium text-zinc-950 dark:text-zinc-50 block mb-1.5">排序优先级</label>
                <Input type="number" value={String(form.sortOrder)} onChange={(e) => setForm((prev) => ({ ...prev, sortOrder: Number(e.target.value) }))} min={0} max={999} />
                <p className="text-xs text-zinc-500 mt-1">数值越小越靠前</p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
      </Main>
    </>
  )
}
