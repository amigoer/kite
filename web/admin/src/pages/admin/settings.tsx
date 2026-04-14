import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

import { settingsApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Globe,
  LinkIcon,
  UserPlus,
  Languages,
  Monitor,
  Sun,
  Moon,
  Upload,
  Eye,
  Check,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useTheme } from "@/components/theme-provider";
import { localeLabels, type Locale } from "@/i18n";
import { toast } from "sonner";

export default function SettingsPage() {
  const { t, locale, setLocale } = useI18n();
  const { theme, setTheme } = useTheme();
  const queryClient = useQueryClient();
  const [form, setForm] = useState<Record<string, string>>({});
  const [saved, setSaved] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["settings"],
    queryFn: () => settingsApi.get().then((r) => r.data.data),
  });

  useEffect(() => {
    if (data) setForm(data);
  }, [data]);

  const mutation = useMutation({
    mutationFn: () => settingsApi.update(form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["settings"] });
      setSaved(true);
      toast.success("全局设置已保存");
      setTimeout(() => setSaved(false), 2000);
    },
    onError: () => toast.error("请求受阻，全局设置保存失败"),
  });

  const updateField = (key: string, value: string) =>
    setForm((prev) => ({ ...prev, [key]: value }));

  const toggleField = (key: string) =>
    setForm((prev) => ({ ...prev, [key]: prev[key] === "true" ? "false" : "true" }));

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
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t("settings.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("settings.description")}</p>
      </div>

      {/* Appearance */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("settings.appearance")}</CardTitle>
          <CardDescription>{t("settings.appearanceDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-3">
              <Languages className="size-4 text-muted-foreground" />
              <span className="text-sm font-medium">{t("settings.language")}</span>
            </div>
            <div className="flex rounded-lg bg-muted p-1">
              {(Object.keys(localeLabels) as Locale[]).map((l) => (
                <button
                  key={l}
                  className={cn(
                    "rounded-md px-4 py-1.5 text-[13px] font-medium transition-all",
                    locale === l
                      ? "bg-background text-foreground shadow-sm"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                  onClick={() => setLocale(l)}
                >
                  {localeLabels[l]}
                </button>
              ))}
            </div>
          </div>

          <Separator />

          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-3">
              <Monitor className="size-4 text-muted-foreground" />
              <span className="text-sm font-medium">{t("settings.theme")}</span>
            </div>
            <div className="flex rounded-lg bg-muted p-1">
              {([
                { value: "light", icon: Sun, label: t("settings.themeLight") },
                { value: "dark", icon: Moon, label: t("settings.themeDark") },
                { value: "system", icon: Monitor, label: t("settings.themeSystem") },
              ] as const).map((item) => (
                <button
                  key={item.value}
                  className={cn(
                    "flex items-center gap-1.5 rounded-md px-3 py-1.5 text-[13px] font-medium transition-all",
                    theme === item.value
                      ? "bg-background text-foreground shadow-sm"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                  onClick={() => setTheme(item.value)}
                >
                  <item.icon className="size-3.5" />
                  {item.label}
                </button>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Site config */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("settings.site")}</CardTitle>
          <CardDescription>{t("settings.siteDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-3">
              <Globe className="size-4 text-muted-foreground" />
              <span className="text-sm font-medium">{t("settings.siteName")}</span>
            </div>
            <Input
              className="w-full sm:w-80"
              value={form.site_name ?? ""}
              onChange={(e) => updateField("site_name", e.target.value)}
            />
          </div>

          <Separator />

          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-3">
              <LinkIcon className="size-4 text-muted-foreground" />
              <span className="text-sm font-medium">{t("settings.siteUrl")}</span>
            </div>
            <Input
              className="w-full sm:w-80"
              value={form.site_url ?? ""}
              onChange={(e) => updateField("site_url", e.target.value)}
              placeholder={t("settings.siteUrlPlaceholder")}
            />
          </div>
        </CardContent>
      </Card>

      {/* Access control */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("settings.accessControl")}</CardTitle>
          <CardDescription>{t("settings.accessControlDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Registration */}
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <UserPlus className="size-4 shrink-0 text-muted-foreground" />
              <div>
                <span className="text-sm font-medium">{t("settings.allowRegistration")}</span>
                <p className="hidden text-xs text-muted-foreground sm:block">
                  {t("settings.allowRegistrationDesc")}
                </p>
              </div>
            </div>
            <Switch
              checked={form.allow_registration === "true"}
              onCheckedChange={() => toggleField("allow_registration")}
            />
          </div>

          <Separator />

          {/* Guest upload */}
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <Upload className="size-4 shrink-0 text-muted-foreground" />
              <div>
                <span className="text-sm font-medium">{t("settings.allowGuestUpload")}</span>
                <p className="hidden text-xs text-muted-foreground sm:block">
                  {t("settings.allowGuestUploadDesc")}
                </p>
              </div>
            </div>
            <Switch
              checked={form.allow_guest_upload === "true"}
              onCheckedChange={() => toggleField("allow_guest_upload")}
            />
          </div>

          <Separator />

          {/* Public gallery */}
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <Eye className="size-4 shrink-0 text-muted-foreground" />
              <div>
                <span className="text-sm font-medium">{t("settings.allowPublicGallery")}</span>
                <p className="hidden text-xs text-muted-foreground sm:block">
                  {t("settings.allowPublicGalleryDesc")}
                </p>
              </div>
            </div>
            <Switch
              checked={form.allow_public_gallery === "true"}
              onCheckedChange={() => toggleField("allow_public_gallery")}
            />
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-start">
        <Button
          className="px-8"
          onClick={() => mutation.mutate()}
          disabled={mutation.isPending}
        >
          {saved ? (
            <>
              <Check className="size-4" />
              {t("settings.saved")}
            </>
          ) : mutation.isPending ? (
            t("settings.saving")
          ) : (
            t("settings.saveSettings")
          )}
        </Button>
      </div>
    </div>
  );
}
