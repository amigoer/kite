import { useState, useEffect } from 'react'
import { Save, Globe, FileText, Server, Sparkles, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useSettings, useSaveSettings } from '@/hooks/use-settings'
import type { AllSettings } from '@/types/settings'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

/**
 * 系统设置页面
 * 参考风格：左侧竖排标签 | 分割线 | 右侧水平 label-value 表单
 */
export function SettingsPage() {
  const [form, setForm] = useState<AllSettings | null>(null)
  const [saved, setSaved] = useState(false)

  const { data: settings, isLoading } = useSettings()
  const saveMutation = useSaveSettings()

  useEffect(() => {
    if (settings && !form) setForm(structuredClone(settings))
  }, [settings, form])

  function handleSave() {
    if (!form) return
    saveMutation.mutate(form, {
      onSuccess: () => { setSaved(true); setTimeout(() => setSaved(false), 2000) },
    })
  }

  if (isLoading || !form) {
    return <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">加载中…</div>
  }

  return (
    <div>
      {/* 页面标题 */}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">设置</h1>
        <Button onClick={handleSave} disabled={saveMutation.isPending}>
          {saved ? (
            <><Check className="mr-2 h-4 w-4" />已保存</>
          ) : saveMutation.isPending ? '保存中…' : (
            <><Save className="mr-2 h-4 w-4" />保存设置</>
          )}
        </Button>
      </div>

      <Tabs defaultValue="site" orientation="vertical" className="gap-0">
        {/* 左侧导航 */}
        <TabsList variant="line" className="h-auto w-48 shrink-0 items-stretch p-0 gap-0">
          <TabsTrigger value="site" className="justify-start gap-2.5 px-3 py-2.5 data-[state=active]:bg-muted">
            <Globe className="h-4 w-4" />站点信息
          </TabsTrigger>
          <TabsTrigger value="post" className="justify-start gap-2.5 px-3 py-2.5 data-[state=active]:bg-muted">
            <FileText className="h-4 w-4" />文章设置
          </TabsTrigger>
          <TabsTrigger value="render" className="justify-start gap-2.5 px-3 py-2.5 data-[state=active]:bg-muted">
            <Server className="h-4 w-4" />渲染模式
          </TabsTrigger>
          <TabsTrigger value="ai" className="justify-start gap-2.5 px-3 py-2.5 data-[state=active]:bg-muted">
            <Sparkles className="h-4 w-4" />AI 集成
          </TabsTrigger>
        </TabsList>

        {/* 分割线 */}
        <Separator orientation="vertical" className="mx-0 h-auto" />

        {/* 站点信息 */}
        <TabsContent value="site" className="mt-0 flex-1 pl-8">
          <div className="space-y-6">
            <FormRow label="站点名称">
              <Input value={form.site.siteName} onChange={(e) => updateField(setForm, 'site', 'siteName', e.target.value)} placeholder="Kite Blog" className="max-w-sm" />
            </FormRow>
            <FormRow label="站点 URL">
              <Input value={form.site.siteUrl} onChange={(e) => updateField(setForm, 'site', 'siteUrl', e.target.value)} placeholder="https://blog.example.com" className="max-w-sm" />
            </FormRow>
            <FormRow label="站点描述">
              <textarea value={form.site.description} onChange={(e) => updateField(setForm, 'site', 'description', e.target.value)} placeholder="简要描述你的博客…" rows={2} className="flex max-w-sm w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-none" />
            </FormRow>
            <FormRow label="SEO 关键词">
              <Input value={form.site.keywords} onChange={(e) => updateField(setForm, 'site', 'keywords', e.target.value)} placeholder="博客,技术,Go,React" className="max-w-sm" />
            </FormRow>
            <FormRow label="ICP 备案号">
              <Input value={form.site.icp} onChange={(e) => updateField(setForm, 'site', 'icp', e.target.value)} placeholder="京ICP备XXXXXXXX号" className="max-w-sm" />
            </FormRow>
            <FormRow label="Favicon 路径">
              <Input value={form.site.favicon} onChange={(e) => updateField(setForm, 'site', 'favicon', e.target.value)} placeholder="/favicon.svg" className="max-w-sm" />
            </FormRow>
            <FormRow label="页脚文本">
              <Input value={form.site.footer} onChange={(e) => updateField(setForm, 'site', 'footer', e.target.value)} placeholder="© 2026 Blog" className="max-w-sm" />
            </FormRow>
          </div>
        </TabsContent>

        {/* 文章设置 */}
        <TabsContent value="post" className="mt-0 flex-1 pl-8">
          <div className="space-y-6">
            <FormRow label="每页文章数">
              <Input type="number" value={form.post.postsPerPage} onChange={(e) => updateField(setForm, 'post', 'postsPerPage', Number(e.target.value))} min={1} max={50} className="w-24" />
            </FormRow>
            <FormRow label="摘要截取长度">
              <Input type="number" value={form.post.summaryLength} onChange={(e) => updateField(setForm, 'post', 'summaryLength', Number(e.target.value))} min={50} max={500} className="w-24" />
            </FormRow>
            <FormRow label="启用评论" description="允许读者在文章下方发表评论">
              <ToggleSwitch checked={form.post.enableComment as boolean} onChange={(v) => updateField(setForm, 'post', 'enableComment', v)} />
            </FormRow>
            <FormRow label="启用目录" description="自动生成文章目录导航">
              <ToggleSwitch checked={form.post.enableToc as boolean} onChange={(v) => updateField(setForm, 'post', 'enableToc', v)} />
            </FormRow>
          </div>
        </TabsContent>

        {/* 渲染模式 */}
        <TabsContent value="render" className="mt-0 flex-1 pl-8">
          <div className="space-y-6">
            <FormRow label="渲染模式">
              <Select value={form.render.renderMode} onValueChange={(v) => updateField(setForm, 'render', 'renderMode', v)}>
                <SelectTrigger className="max-w-sm"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="classic">Classic — Go Template 服务端渲染</SelectItem>
                  <SelectItem value="headless">Headless — 纯 JSON API 输出</SelectItem>
                </SelectContent>
              </Select>
            </FormRow>
            <FormRow label="API 前缀">
              <Input value={form.render.apiPrefix} onChange={(e) => updateField(setForm, 'render', 'apiPrefix', e.target.value)} placeholder="/api/v1" className="max-w-sm" />
            </FormRow>
            <FormRow label="启用 CORS" description="允许跨域请求">
              <ToggleSwitch checked={form.render.enableCors as boolean} onChange={(v) => updateField(setForm, 'render', 'enableCors', v)} />
            </FormRow>
          </div>
        </TabsContent>

        {/* AI 集成 */}
        <TabsContent value="ai" className="mt-0 flex-1 pl-8">
          <div className="space-y-6">
            <FormRow label="启用 AI 功能" description="开启后提供 AI 辅助">
              <ToggleSwitch checked={form.ai.enabled as boolean} onChange={(v) => updateField(setForm, 'ai', 'enabled', v)} />
            </FormRow>
            {form.ai.enabled && (
              <>
                <FormRow label="AI 服务商">
                  <Select value={form.ai.provider} onValueChange={(v) => updateField(setForm, 'ai', 'provider', v)}>
                    <SelectTrigger className="max-w-sm"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="deepseek">DeepSeek</SelectItem>
                      <SelectItem value="openai">OpenAI</SelectItem>
                    </SelectContent>
                  </Select>
                </FormRow>
                <FormRow label="模型名称">
                  <Input value={form.ai.model} onChange={(e) => updateField(setForm, 'ai', 'model', e.target.value)} placeholder="deepseek-chat" className="max-w-sm" />
                </FormRow>
                <FormRow label="API Key">
                  <Input type="password" value={form.ai.apiKey} onChange={(e) => updateField(setForm, 'ai', 'apiKey', e.target.value)} placeholder="sk-xxxxxxxxxxxx" className="max-w-sm" />
                </FormRow>
                <FormRow label="自动生成摘要" description="发布时自动调用 AI 生成摘要">
                  <ToggleSwitch checked={form.ai.autoSummary as boolean} onChange={(v) => updateField(setForm, 'ai', 'autoSummary', v)} />
                </FormRow>
                <FormRow label="自动推荐标签" description="基于内容推荐标签">
                  <ToggleSwitch checked={form.ai.autoTag as boolean} onChange={(v) => updateField(setForm, 'ai', 'autoTag', v)} />
                </FormRow>
              </>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}

/* ========== 辅助组件 ========== */

/** 水平表单行：label 左侧对齐，input 右侧 */
function FormRow({ label, description, children }: { label: string; description?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start gap-8">
      <div className="w-36 shrink-0 pt-2">
        <p className="text-sm font-medium">{label}</p>
        {description && <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>}
      </div>
      <div className="flex-1">{children}</div>
    </div>
  )
}

/** 开关组件 */
function ToggleSwitch({ checked, onChange }: { checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className={cn(
        'mt-1 inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors',
        checked ? 'bg-primary' : 'bg-input'
      )}
    >
      <span className={cn(
        'pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform',
        checked ? 'translate-x-5' : 'translate-x-0'
      )} />
    </button>
  )
}

/** 通用表单字段更新 */
function updateField(
  setForm: React.Dispatch<React.SetStateAction<AllSettings | null>>,
  section: keyof AllSettings,
  key: string,
  value: unknown
) {
  setForm((prev) => prev ? { ...prev, [section]: { ...prev[section], [key]: value } } : prev)
}
