import { useQuery } from "@tanstack/react-query";
import { FileText, HardDrive, Image, Video, Music, Upload } from "lucide-react";
import { statsApi, fileApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { formatSize, calcPercent, formatRelativeTime } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";

interface Stats {
  total_files: number;
  total_size: number;
  images: number;
  videos: number;
  audios: number;
  others: number;
}

const fileTypeConfig: Record<string, { icon: typeof FileText; accent: string }> = {
  image: { icon: Image, accent: "bg-amber-50 text-amber-600 dark:bg-amber-950/40 dark:text-amber-400" },
  video: { icon: Video, accent: "bg-violet-50 text-violet-600 dark:bg-violet-950/40 dark:text-violet-400" },
  audio: { icon: Music, accent: "bg-blue-50 text-blue-600 dark:bg-blue-950/40 dark:text-blue-400" },
  file: { icon: FileText, accent: "bg-muted text-muted-foreground" },
};

export default function UserDashboard() {
  const { t, locale } = useI18n();

  const { data: stats, isLoading: statsLoading } = useQuery<Stats>({
    queryKey: ["dashboard", "stats"],
    queryFn: () => statsApi.get().then((r) => r.data.data),
    staleTime: 30_000,
  });

  const { data: recentData, isLoading: filesLoading } = useQuery({
    queryKey: ["files", "recent"],
    queryFn: () => fileApi.list({ page: 1, size: 5, sort: "created_at", order: "desc" }).then((r) => r.data.data),
    staleTime: 15_000,
  });

  const s = stats ?? { total_files: 0, total_size: 0, images: 0, videos: 0, audios: 0, others: 0 };
  const recentFiles = recentData?.items ?? [];

  const statCards = [
    { label: t("dashboard.totalFiles"), value: s.total_files, icon: FileText, accent: "bg-muted text-foreground" },
    { label: t("dashboard.storageUsed"), value: formatSize(s.total_size), icon: HardDrive, accent: "bg-blue-50 text-blue-600 dark:bg-blue-950/40 dark:text-blue-400" },
    { label: t("dashboard.images"), value: s.images, icon: Image, accent: "bg-amber-50 text-amber-600 dark:bg-amber-950/40 dark:text-amber-400" },
    { label: t("dashboard.videos"), value: s.videos, icon: Video, accent: "bg-violet-50 text-violet-600 dark:bg-violet-950/40 dark:text-violet-400" },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t("dashboard.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("dashboard.description")}</p>
      </div>

      <Separator />

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {statCards.map((card) => (
          <Card key={card.label} className="gap-0 py-0">
            <CardContent className="p-5">
              <div className="flex items-center justify-between">
                <span className="text-[13px] font-medium text-muted-foreground">{card.label}</span>
                <div className={`flex size-8 items-center justify-center rounded-lg ${card.accent}`}>
                  <card.icon size={15} />
                </div>
              </div>
              {statsLoading ? (
                <Skeleton className="mt-3 h-8 w-20" />
              ) : (
                <div className="mt-3 text-[28px] font-bold leading-none tracking-tight">
                  {typeof card.value === "number" ? card.value.toLocaleString() : card.value}
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* File type distribution */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-semibold">{t("dashboard.fileTypeDistribution")}</CardTitle>
            <CardDescription>{t("dashboard.byFileType")}</CardDescription>
          </CardHeader>
          <CardContent>
            {statsLoading ? (
              <div className="space-y-4">
                {Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-5 w-full" />)}
              </div>
            ) : (
              <div className="space-y-0">
                {[
                  { name: t("dashboard.images"), count: s.images, color: "bg-amber-500" },
                  { name: t("dashboard.videos"), count: s.videos, color: "bg-violet-500" },
                  { name: t("dashboard.audio"), count: s.audios, color: "bg-blue-500" },
                  { name: t("dashboard.otherFiles"), count: s.others, color: "bg-gray-400 dark:bg-gray-500" },
                ].map((ft) => (
                  <div key={ft.name} className="flex items-center gap-3 border-b border-border/40 py-3 last:border-b-0 last:pb-0 first:pt-0">
                    <span className={`size-2 rounded-full ${ft.color}`} />
                    <span className="min-w-[3.5rem] text-sm">{ft.name}</span>
                    <div className="mx-1 h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                      <div
                        className={`h-full rounded-full transition-all duration-500 ${ft.color}`}
                        style={{ width: `${s.total_files > 0 ? (ft.count / s.total_files) * 100 : 0}%` }}
                      />
                    </div>
                    <span className="min-w-[2.5rem] text-right text-sm font-semibold tabular-nums">{ft.count}</span>
                    <span className="min-w-[2.5rem] text-right text-xs text-muted-foreground tabular-nums">
                      {calcPercent(ft.count, s.total_files)}%
                    </span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Recent files */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-semibold">{t("dashboard.recentUploads")}</CardTitle>
            <CardDescription>{t("dashboard.latestFiles")}</CardDescription>
          </CardHeader>
          <CardContent>
            {filesLoading ? (
              <div className="space-y-4">
                {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12 w-full" />)}
              </div>
            ) : recentFiles.length === 0 ? (
              <div className="flex flex-col items-center py-10 text-center">
                <div className="mb-3 flex size-12 items-center justify-center rounded-full bg-muted">
                  <Upload size={20} className="text-muted-foreground" />
                </div>
                <p className="text-sm text-muted-foreground">{t("dashboard.noFilesYet")}</p>
              </div>
            ) : (
              <div className="space-y-0">
                {recentFiles.map((file: { id: string; original_name: string; size_bytes: number; mime_type: string; file_type: string; created_at: string }) => {
                  const cfg = fileTypeConfig[file.file_type] ?? fileTypeConfig.file;
                  const Icon = cfg.icon;
                  return (
                    <div key={file.id} className="flex items-center gap-3 border-b border-border/40 py-3 last:border-b-0 last:pb-0 first:pt-0">
                      <div className={`flex size-9 shrink-0 items-center justify-center rounded-lg ${cfg.accent}`}>
                        <Icon size={15} />
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">{file.original_name}</p>
                        <p className="text-xs text-muted-foreground">{formatSize(file.size_bytes)}</p>
                      </div>
                      <span className="shrink-0 text-xs text-muted-foreground">
                        {formatRelativeTime(file.created_at, locale)}
                      </span>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
