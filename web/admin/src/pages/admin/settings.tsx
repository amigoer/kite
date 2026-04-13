import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { settingsApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Globe, LinkIcon, UserPlus, Languages, Monitor, Sun, Moon } from "lucide-react";
import { cn } from "@/lib/utils";
import { useTheme } from "@/components/theme-provider";
import { localeLabels, type Locale } from "@/i18n";

export default function SettingsPage() {
  const { t, locale, setLocale } = useI18n();
  const { theme, setTheme } = useTheme();
  const queryClient = useQueryClient();
  const [form, setForm] = useState<Record<string, string>>({});

  const { data, isLoading } = useQuery({
    queryKey: ["settings"],
    queryFn: () => settingsApi.get().then((r) => r.data.data),
  });

  useEffect(() => {
    if (data) setForm(data);
  }, [data]);

  const mutation = useMutation({
    mutationFn: () => settingsApi.update(form),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["settings"] }),
  });

  const updateField = (key: string, value: string) =>
    setForm((prev) => ({ ...prev, [key]: value }));

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-48" />
        <Skeleton className="h-32" />
      </div>
    );
  }

  return (
    <div className="space-y-6 pb-12">
      <div className="mb-8 mt-2">
        <h1 className="text-[22px] font-bold tracking-tight text-foreground">{t("settings.title")}</h1>
        <p className="mt-2 text-[14px] text-muted-foreground">{t("settings.description")}</p>
      </div>

      <Card>
        <CardHeader className="pt-6 pb-4">
          <CardTitle className="text-base font-semibold">{t("settings.appearance")}</CardTitle>
          <p className="mt-1 text-[13px] text-muted-foreground">{t("settings.appearanceDesc")}</p>
        </CardHeader>
        <CardContent className="space-y-5 pb-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <Languages className="size-4 text-muted-foreground" />
              <span className="text-[14px] font-medium">{t("settings.language")}</span>
            </div>
            <div className="flex bg-muted p-1 rounded-lg">
               {(Object.keys(localeLabels) as Locale[]).map((l) => (
                 <button
                   key={l}
                   className={cn(
                     "px-4 py-1.5 rounded-md text-[13px] font-medium transition-all shadow-none",
                     locale === l ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"
                   )}
                   onClick={() => setLocale(l)}
                 >
                   {localeLabels[l]}
                 </button>
               ))}
            </div>
          </div>
          
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <Monitor className="size-4 text-muted-foreground" />
              <span className="text-[14px] font-medium">{t("settings.theme")}</span>
            </div>
            <div className="flex bg-muted p-1 rounded-lg">
                 <button
                   className={cn(
                     "flex items-center gap-2 px-4 py-1.5 rounded-md text-[13px] font-medium transition-all shadow-none",
                     theme === "light" ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"
                   )}
                   onClick={() => setTheme("light")}
                 >
                   <Sun className="size-3.5" />
                   {t("settings.themeLight")}
                 </button>
                 <button
                   className={cn(
                     "flex items-center gap-2 px-4 py-1.5 rounded-md text-[13px] font-medium transition-all shadow-none",
                     theme === "dark" ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"
                   )}
                   onClick={() => setTheme("dark")}
                 >
                   <Moon className="size-3.5" />
                   {t("settings.themeDark")}
                 </button>
                 <button
                   className={cn(
                     "flex items-center gap-2 px-4 py-1.5 rounded-md text-[13px] font-medium transition-all shadow-none",
                     theme === "system" ? "bg-background shadow-sm text-foreground" : "text-muted-foreground hover:text-foreground"
                   )}
                   onClick={() => setTheme("system")}
                 >
                   <Monitor className="size-3.5" />
                   {t("settings.themeSystem")}
                 </button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pt-6 pb-4">
          <CardTitle className="text-base font-semibold">{t("settings.site")}</CardTitle>
          <p className="mt-1 text-[13px] text-muted-foreground">设置本站的基础信息及对外的访问 URL。</p>
        </CardHeader>
        <CardContent className="space-y-5 pb-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <Globe className="size-4 text-muted-foreground" />
              <span className="text-[14px] font-medium">{t("settings.siteName")}</span>
            </div>
            <Input
              className="w-full sm:w-[320px]"
              value={form.site_name ?? ""}
              onChange={(e) => updateField("site_name", e.target.value)}
            />
          </div>
          
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <LinkIcon className="size-4 text-muted-foreground" />
              <span className="text-[14px] font-medium">{t("settings.siteUrl")}</span>
            </div>
            <Input
              className="w-full sm:w-[320px]"
              value={form.site_url ?? ""}
              onChange={(e) => updateField("site_url", e.target.value)}
              placeholder={t("settings.siteUrlPlaceholder")}
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pt-6 pb-4">
          <CardTitle className="text-base font-semibold">{t("settings.registration")}</CardTitle>
          <p className="mt-1 text-[13px] text-muted-foreground">配置系统的用户注册策略，关闭后仅能通过管理员手动创建账号。</p>
        </CardHeader>
        <CardContent className="pb-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <UserPlus className="size-4 text-muted-foreground" />
              <div className="flex flex-col">
                <span className="text-[14px] font-medium">{t("settings.allowRegistration")}</span>
                <span className="text-[12px] text-muted-foreground hidden sm:block">{t("settings.allowRegistrationDesc")}</span>
              </div>
            </div>
            <Button
              variant={form.allow_registration === "true" ? "default" : "secondary"}
              className={cn(
                "h-8 px-4 text-[13px]",
                form.allow_registration !== "true" && "bg-muted hover:bg-muted/80 text-foreground"
              )}
              onClick={() =>
                updateField(
                  "allow_registration",
                  form.allow_registration === "true" ? "false" : "true"
                )
              }
            >
              {form.allow_registration === "true" ? t("common.enabled") : t("common.disabled")}
            </Button>
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-start mt-8">
        <Button 
          className="px-6" 
          onClick={() => mutation.mutate()} 
          disabled={mutation.isPending}
        >
          {mutation.isPending ? t("settings.saving") : t("settings.saveSettings")}
        </Button>
      </div>
    </div>
  );
}
