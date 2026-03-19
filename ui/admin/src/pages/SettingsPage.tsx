import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { Save, Check, Globe, FileText, Server, Sparkles, Loader2 } from 'lucide-react'
import { useSettings, useSaveSettings } from '@/hooks/use-settings'
import type { AllSettings } from '@/types/settings'
import { cn } from '@/lib/utils'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search as HeaderSearch } from '@/components/search'

const tabs = [
  { key: 'site', label: '站点信息', icon: Globe },
  { key: 'post', label: '文章设置', icon: FileText },
  { key: 'render', label: '渲染模式', icon: Server },
  { key: 'ai', label: 'AI 集成', icon: Sparkles },
]

const inputCls = 'border-zinc-200 dark:border-zinc-700 bg-transparent rounded-md shadow-none focus-visible:ring-1 focus-visible:ring-zinc-400'

/**
 * 系统设置 — Vercel 风格
 */
export function SettingsPage() {
  const [form, setForm] = useState<AllSettings | null>(null)
  const [saved, setSaved] = useState(false)
  const [activeTab, setActiveTab] = useState('site')
  const [aiKeywordsLoading, setAiKeywordsLoading] = useState(false)
  const [aiSuggestions, setAiSuggestions] = useState<string[]>([])

  const { data: settings, isLoading } = useSettings()
  const saveMutation = useSaveSettings()

  useEffect(() => {
    if (settings && !form) setForm(structuredClone(settings))
  }, [settings, form])

  function handleSave() {
    if (!form) return
    saveMutation.mutate(form, { onSuccess: () => { setSaved(true); setTimeout(() => setSaved(false), 2000) } })
  }

  function handleAiKeywords() {
    setAiKeywordsLoading(true)
    setAiSuggestions([])
    setTimeout(() => {
      setAiSuggestions(['博客', '技术', 'Go', 'React', '静态网站', '极简风'])
      setAiKeywordsLoading(false)
    }, 800)
  }

  function applyAiSuggestions() {
    if (!form || aiSuggestions.length === 0) return
    updateField(setForm, 'site', 'keywords', aiSuggestions.join(','))
  }

  if (isLoading || !form) {
    return <div className="flex justify-center py-16"><Loader2 className="w-4 h-4 animate-spin text-zinc-400" /></div>
  }

  return (
    <>
      <Header fixed>
        <HeaderSearch />
        <div className='ml-auto' />
      </Header>
      <Main>
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50 tracking-tight">设置</h1>
          <p className="text-sm text-zinc-500 mt-1">管理站点配置</p>
        </div>
        <Button className="bg-zinc-900 dark:bg-zinc-100 text-white dark:text-zinc-900 rounded-md shadow-sm hover:bg-zinc-800 dark:hover:bg-zinc-200 flex gap-2 items-center h-9 text-sm" onClick={handleSave} disabled={saveMutation.isPending}>
          {saved ? <><Check className="w-4 h-4" /> 已保存</> : saveMutation.isPending ? '保存中…' : <><Save className="w-4 h-4" /> 保存设置</>}
        </Button>
      </div>

      <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg shadow-sm">
        <div className="flex min-h-[520px]">
          <div className="w-52 border-r border-zinc-100 dark:border-zinc-800 py-4">
            {tabs.map((tab) => {
              const Icon = tab.icon
              return (
                <button
                  key={tab.key}
                  className={cn(
                    'w-full flex items-center gap-2.5 px-6 py-2.5 text-[13px] transition-colors cursor-pointer',
                    activeTab === tab.key
                      ? 'bg-zinc-50 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-50 font-medium'
                      : 'text-zinc-500 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 hover:text-zinc-700'
                  )}
                  onClick={() => setActiveTab(tab.key)}
                >
                  <Icon className="w-4 h-4" /> {tab.label}
                </button>
              )
            })}
          </div>

          <div className="flex-1 p-8 space-y-7">
            {activeTab === 'site' && (
              <>
                <FormRow label="站点名称"><Input value={form.site.siteName} onChange={(e) => updateField(setForm, 'site', 'siteName', e.target.value)} placeholder="Kite Blog" className={cn(inputCls, 'max-w-md')} /></FormRow>
                <FormRow label="站点 URL"><Input value={form.site.siteUrl} onChange={(e) => updateField(setForm, 'site', 'siteUrl', e.target.value)} placeholder="https://blog.example.com" className={cn(inputCls, 'max-w-md')} /></FormRow>
                <FormRow label="站点描述"><Textarea value={form.site.description} onChange={(e) => updateField(setForm, 'site', 'description', e.target.value)} placeholder="简要描述你的博客…" rows={2} className={cn(inputCls, 'max-w-md resize-none')} /></FormRow>

                <FormRow label="SEO 关键词">
                  <div>
                    <div className="flex gap-2 items-center">
                      <Input value={form.site.keywords} onChange={(e) => updateField(setForm, 'site', 'keywords', e.target.value)} placeholder="博客,技术,Go,React" className={cn(inputCls, 'max-w-[240px]')} />
                      <Button variant="outline" className="rounded-md shadow-sm border-zinc-200 dark:border-zinc-700 text-sm h-9 gap-1.5 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 shrink-0" onClick={handleAiKeywords} disabled={aiKeywordsLoading}>
                        <Sparkles className="w-3.5 h-3.5" /> {aiKeywordsLoading ? '生成中…' : 'AI 生成建议'}
                      </Button>
                    </div>
                    {aiSuggestions.length > 0 && (
                      <div className="mt-2 flex items-center gap-2">
                        <p className="text-xs text-zinc-500">AI 建议：<span className="text-zinc-700 dark:text-zinc-300">{aiSuggestions.join(', ')}</span></p>
                        <button className="text-xs text-zinc-900 dark:text-zinc-100 underline underline-offset-2 cursor-pointer hover:no-underline" onClick={applyAiSuggestions}>应用</button>
                      </div>
                    )}
                  </div>
                </FormRow>
                <FormRow label="ICP 备案号"><Input value={form.site.icp} onChange={(e) => updateField(setForm, 'site', 'icp', e.target.value)} placeholder="京ICP备XXXXXXXX号" className={cn(inputCls, 'max-w-md')} /></FormRow>
                <FormRow label="Favicon"><Input value={form.site.favicon} onChange={(e) => updateField(setForm, 'site', 'favicon', e.target.value)} placeholder="/favicon.svg" className={cn(inputCls, 'max-w-md')} /></FormRow>
                <FormRow label="页脚文本"><Input value={form.site.footer} onChange={(e) => updateField(setForm, 'site', 'footer', e.target.value)} placeholder="© 2026 Blog" className={cn(inputCls, 'max-w-md')} /></FormRow>
              </>
            )}
            {activeTab === 'post' && (
              <>
                <FormRow label="每页文章数"><Input type="number" value={String(form.post.postsPerPage)} onChange={(e) => updateField(setForm, 'post', 'postsPerPage', Number(e.target.value))} className={cn(inputCls, 'w-24')} /></FormRow>
                <FormRow label="摘要截取长度"><Input type="number" value={String(form.post.summaryLength)} onChange={(e) => updateField(setForm, 'post', 'summaryLength', Number(e.target.value))} className={cn(inputCls, 'w-24')} /></FormRow>
                <FormRow label="启用评论" description="允许读者发表评论"><Switch checked={form.post.enableComment as boolean} onCheckedChange={(v) => updateField(setForm, 'post', 'enableComment', v)} /></FormRow>
                <FormRow label="启用目录" description="自动生成目录导航"><Switch checked={form.post.enableToc as boolean} onCheckedChange={(v) => updateField(setForm, 'post', 'enableToc', v)} /></FormRow>
              </>
            )}
            {activeTab === 'render' && (
              <>
                <FormRow label="渲染模式">
                  <Select value={form.render.renderMode} onValueChange={(v) => updateField(setForm, 'render', 'renderMode', v)}>
                    <SelectTrigger className={cn(inputCls, 'max-w-md')}><SelectValue /></SelectTrigger>
                    <SelectContent><SelectItem value="classic">Classic — Go Template</SelectItem><SelectItem value="headless">Headless — JSON API</SelectItem></SelectContent>
                  </Select>
                </FormRow>
                <FormRow label="API 前缀"><Input value={form.render.apiPrefix} onChange={(e) => updateField(setForm, 'render', 'apiPrefix', e.target.value)} placeholder="/api/v1" className={cn(inputCls, 'max-w-md')} /></FormRow>
                <FormRow label="启用 CORS" description="允许跨域请求"><Switch checked={form.render.enableCors as boolean} onCheckedChange={(v) => updateField(setForm, 'render', 'enableCors', v)} /></FormRow>
              </>
            )}
            {activeTab === 'ai' && (
              <>
                <FormRow label="启用 AI" description="开启 AI 辅助写作"><Switch checked={form.ai.enabled as boolean} onCheckedChange={(v) => updateField(setForm, 'ai', 'enabled', v)} /></FormRow>
                {form.ai.enabled && (
                  <>
                    <FormRow label="API 地址" description="OpenAI 兼容接口"><div><Input value={form.ai.provider} onChange={(e) => updateField(setForm, 'ai', 'provider', e.target.value)} placeholder="https://api.deepseek.com" className={cn(inputCls, 'max-w-md')} /><p className="text-xs text-zinc-500 mt-1.5">DeepSeek / OpenAI 兼容</p></div></FormRow>
                    <FormRow label="模型"><Input value={form.ai.model} onChange={(e) => updateField(setForm, 'ai', 'model', e.target.value)} placeholder="deepseek-chat" className={cn(inputCls, 'max-w-md')} /></FormRow>
                    <FormRow label="API Key"><Input type="password" value={form.ai.apiKey} onChange={(e) => updateField(setForm, 'ai', 'apiKey', e.target.value)} placeholder="sk-xxxxxxxxxxxx" className={cn(inputCls, 'max-w-md')} /></FormRow>
                    <FormRow label="自动摘要" description="发布时自动生成"><Switch checked={form.ai.autoSummary as boolean} onCheckedChange={(v) => updateField(setForm, 'ai', 'autoSummary', v)} /></FormRow>
                    <FormRow label="自动标签" description="基于内容推荐"><Switch checked={form.ai.autoTag as boolean} onCheckedChange={(v) => updateField(setForm, 'ai', 'autoTag', v)} /></FormRow>
                  </>
                )}
              </>
            )}
          </div>
        </div>
      </div>
      </Main>
    </>
  )
}

function FormRow({ label, description, children }: { label: string; description?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-10">
      <div className="w-36 shrink-0 pt-2">
        <p className="text-sm font-medium text-zinc-700 dark:text-zinc-300">{label}</p>
        {description && <p className="text-xs text-zinc-500 mt-0.5">{description}</p>}
      </div>
      <div className="flex-1">{children}</div>
    </div>
  )
}

function updateField(setForm: React.Dispatch<React.SetStateAction<AllSettings | null>>, section: keyof AllSettings, key: string, value: unknown) {
  setForm((prev) => prev ? { ...prev, [section]: { ...prev[section], [key]: value } } : prev)
}
