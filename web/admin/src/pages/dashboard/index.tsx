import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowUp,
  Files,
  FileIcon,
  HardDrive,
  Image as ImageIcon,
  Music,
  Upload,
  Users,
  Video,
  type LucideIcon,
} from "lucide-react";
import { useLocation, useNavigate } from "react-router-dom";
import { adminFileApi, adminStatsApi, fileApi, statsApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { useAuth } from "@/hooks/use-auth";
import { formatSize } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import FileTypeChart from "./FileTypeChart";
import UploadTrendChart from "./UploadTrendChart";

interface DashboardStats {
  total_files: number;
  total_size: number;
  images: number;
  videos: number;
  audios: number;
  others: number;
  users?: number;
}

interface DailyPoint {
  day: string;
  uploads: number;
  accesses: number;
  bytes_served: number;
}

interface RecentFile {
  id: string;
  original_name: string;
  file_type: string;
  size_bytes: number;
  url: string;
  thumb_url?: string;
  created_at: string;
}

function intPercent(value: number, total: number) {
  if (!total) return 0;
  return Math.round((value / total) * 100);
}

function formatRelativeTime(iso: string, locale: string) {
  const date = new Date(iso);
  const diffMs = Date.now() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return locale === "zh" ? "刚刚" : "just now";
  if (diffMin < 60) return locale === "zh" ? `${diffMin} 分钟前` : `${diffMin}m ago`;
  const diffHour = Math.floor(diffMin / 60);
  if (diffHour < 24) return locale === "zh" ? `${diffHour} 小时前` : `${diffHour}h ago`;
  const diffDay = Math.floor(diffHour / 24);
  if (diffDay < 30) return locale === "zh" ? `${diffDay} 天前` : `${diffDay}d ago`;
  return date.toLocaleDateString(locale === "zh" ? "zh-CN" : "en-US");
}

function fileTypeIcon(type: string): LucideIcon {
  switch (type) {
    case "image":
      return ImageIcon;
    case "video":
      return Video;
    case "audio":
      return Music;
    default:
      return FileIcon;
  }
}

export default function DashboardPage() {
  const { t, locale } = useI18n();
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();
  const isAdminWorkspace = location.pathname.startsWith("/admin");

  const { data, isLoading } = useQuery<DashboardStats>({
    queryKey: ["dashboard", "stats", isAdminWorkspace ? "admin" : "user"],
    queryFn: () =>
      (isAdminWorkspace ? adminStatsApi.get() : statsApi.get()).then(
        (r) => r.data.data
      ),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  const { data: daily } = useQuery<{ days: DailyPoint[] }>({
    queryKey: ["dashboard", "daily-stats", isAdminWorkspace ? "admin" : "user"],
    queryFn: () =>
      (isAdminWorkspace ? adminStatsApi.daily(7) : statsApi.daily(7)).then(
        (r) => r.data.data
      ),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  const { data: recent, isLoading: recentLoading } = useQuery<{
    items: RecentFile[];
  }>({
    queryKey: ["dashboard", "recent", isAdminWorkspace ? "admin" : "user"],
    queryFn: () =>
      (isAdminWorkspace
        ? adminFileApi.list({ page: 1, size: 5 })
        : fileApi.list({ page: 1, size: 5, only_self: true })
      ).then((r) => r.data.data),
    staleTime: 30_000,
  });

  const stats = data ?? {
    total_files: 0,
    total_size: 0,
    images: 0,
    videos: 0,
    audios: 0,
    others: 0,
    users: 0,
  };

  const recentUploads = useMemo(
    () => (daily?.days ?? []).reduce((acc, p) => acc + p.uploads, 0),
    [daily?.days]
  );

  const storageLimit = isAdminWorkspace ? 0 : user?.storage_limit ?? 0;
  const storageUsed = stats.total_size;
  const storagePct =
    storageLimit > 0 ? Math.min((storageUsed / storageLimit) * 100, 100) : 0;

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0 space-y-1">
          <h1 className="text-xl font-semibold tracking-tight sm:text-2xl">
            {t("dashboard.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("dashboard.description")}
          </p>
        </div>
        <Button
          size="sm"
          onClick={() => navigate("/user/files")}
          className="w-full sm:w-auto"
        >
          <Upload className="size-4" />
          {t("common.upload")}
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
        <StatCard label={t("dashboard.totalFiles")} icon={Files} isLoading={isLoading}>
          <div className="text-2xl font-semibold tabular-nums tracking-tight">
            {stats.total_files.toLocaleString()}
          </div>
          <div className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
            {recentUploads > 0 && (
              <ArrowUp className="size-3 shrink-0" strokeWidth={2} />
            )}
            <span className="leading-none">
              {t("dashboard.vsLastWeek")}{" "}
              {recentUploads > 0 ? `+${recentUploads}` : recentUploads}
            </span>
          </div>
        </StatCard>

        <StatCard
          label={t("dashboard.storageUsed")}
          icon={HardDrive}
          isLoading={isLoading}
        >
          <div className="flex items-center gap-3">
            <CircularProgress
              value={storagePct}
              unlimited={storageLimit <= 0}
            />
            <div className="min-w-0 flex-1">
              <StorageValue bytes={storageUsed} />
              <div className="mt-0.5 truncate text-[10px] text-muted-foreground tabular-nums">
                {storageLimit > 0
                  ? `/ ${formatSize(storageLimit)}`
                  : t("dashboard.unlimitedStorage")}
              </div>
            </div>
          </div>
        </StatCard>

        <StatCard label={t("dashboard.images")} icon={ImageIcon} isLoading={isLoading}>
          <div className="text-2xl font-semibold tabular-nums tracking-tight">
            {stats.images.toLocaleString()}
          </div>
          <div className="mt-1 text-xs text-muted-foreground">
            {t("dashboard.percentOfTotal")}{" "}
            {intPercent(stats.images, stats.total_files)}%
          </div>
        </StatCard>

        <StatCard
          label={t("dashboard.videos")}
          icon={Video}
          isLoading={isLoading}
        >
          <div className="text-2xl font-semibold tabular-nums tracking-tight">
            {stats.videos.toLocaleString()}
          </div>
          <div className="mt-1 text-xs text-muted-foreground">
            {t("dashboard.percentOfTotal")}{" "}
            {intPercent(stats.videos, stats.total_files)}%
          </div>
        </StatCard>

        {isAdminWorkspace ? (
          <StatCard
            label={t("dashboard.users")}
            icon={Users}
            isLoading={isLoading}
          >
            <div className="text-2xl font-semibold tabular-nums tracking-tight">
              {(stats.users ?? 0).toLocaleString()}
            </div>
            <div className="mt-1 text-xs text-muted-foreground">&nbsp;</div>
          </StatCard>
        ) : (
          <StatCard
            label={t("dashboard.audio")}
            icon={Music}
            isLoading={isLoading}
          >
            <div className="text-2xl font-semibold tabular-nums tracking-tight">
              {stats.audios.toLocaleString()}
            </div>
            <div className="mt-1 text-xs text-muted-foreground">
              {t("dashboard.percentOfTotal")}{" "}
              {intPercent(stats.audios, stats.total_files)}%
            </div>
          </StatCard>
        )}
      </div>

      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        <FileTypeChart
          totalFiles={stats.total_files}
          imageCount={stats.images}
          videoCount={stats.videos}
          audioCount={stats.audios}
          otherCount={stats.others}
          isLoading={isLoading}
        />
        <UploadTrendChart admin={isAdminWorkspace} />
      </div>

      <RecentUploadsSection
        items={recent?.items ?? []}
        isLoading={recentLoading}
        onViewAll={() =>
          navigate(isAdminWorkspace ? "/admin/files" : "/user/files")
        }
        locale={locale}
        viewAllLabel={t("dashboard.viewAll")}
        title={t("dashboard.recentUploads")}
        subtitle={t("dashboard.recentUploadsDesc").replace(
          "{count}",
          String(recent?.items?.length ?? 0)
        )}
        emptyLabel={t("dashboard.noRecentFiles")}
      />
    </div>
  );
}

function StatCard({
  label,
  icon: Icon,
  isLoading,
  children,
}: {
  label: string;
  icon: LucideIcon;
  isLoading: boolean;
  children: React.ReactNode;
}) {
  return (
    <Card className="gap-0 py-0 shadow-xs">
      <CardContent className="p-4">
        <div className="mb-3 flex items-center justify-between gap-2">
          <span className="truncate text-xs font-medium text-muted-foreground">
            {label}
          </span>
          <Icon
            className="size-3.5 shrink-0 text-muted-foreground/70"
            strokeWidth={1.8}
          />
        </div>
        {isLoading ? (
          <>
            <Skeleton className="mb-2 h-7 w-16" />
            <Skeleton className="h-3 w-24" />
          </>
        ) : (
          children
        )}
      </CardContent>
    </Card>
  );
}

function CircularProgress({
  value,
  unlimited,
}: {
  value: number;
  unlimited?: boolean;
}) {
  const size = 52;
  const stroke = 3;
  const r = (size - stroke) / 2;
  const c = 2 * Math.PI * r;
  const pct = unlimited ? 0 : Math.max(0, Math.min(100, value));
  const offset = c * (1 - pct / 100);
  const label = unlimited
    ? "∞"
    : pct > 0 && pct < 1
      ? "<1%"
      : `${Math.round(pct)}%`;
  return (
    <svg
      width={size}
      height={size}
      viewBox={`0 0 ${size} ${size}`}
      className="shrink-0"
    >
      <g transform={`rotate(-90 ${size / 2} ${size / 2})`}>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          stroke="hsl(var(--foreground))"
          strokeOpacity={0.1}
          strokeWidth={stroke}
        />
        {!unlimited && (
          <circle
            cx={size / 2}
            cy={size / 2}
            r={r}
            fill="none"
            stroke="hsl(var(--foreground))"
            strokeWidth={stroke}
            strokeDasharray={c}
            strokeDashoffset={offset}
            strokeLinecap="round"
          />
        )}
      </g>
      <text
        x={size / 2}
        y={size / 2}
        textAnchor="middle"
        dominantBaseline="central"
        className="fill-foreground"
        style={{
          fontSize: unlimited ? 15 : label.length > 3 ? 9 : 11,
          fontWeight: 600,
          fontVariantNumeric: "tabular-nums",
        }}
      >
        {label}
      </text>
    </svg>
  );
}

function StorageValue({ bytes }: { bytes: number }) {
  const raw = formatSize(bytes);
  const match = raw.match(/^([\d.,]+)\s*(\S+)$/);
  if (!match) {
    return (
      <div className="text-2xl font-semibold tabular-nums tracking-tight">
        {raw}
      </div>
    );
  }
  const [, value, unit] = match;
  return (
    <div className="flex items-baseline gap-1">
      <span className="text-2xl font-semibold tabular-nums tracking-tight">
        {value}
      </span>
      <span className="text-xs font-normal text-muted-foreground">{unit}</span>
    </div>
  );
}

function RecentUploadsSection({
  items,
  isLoading,
  onViewAll,
  locale,
  viewAllLabel,
  title,
  subtitle,
  emptyLabel,
}: {
  items: RecentFile[];
  isLoading: boolean;
  onViewAll: () => void;
  locale: string;
  viewAllLabel: string;
  title: string;
  subtitle: string;
  emptyLabel: string;
}) {
  return (
    <Card className="gap-0 py-0 shadow-xs">
      <CardHeader className="px-4 pt-4 pb-3 sm:px-6 sm:pt-5">
        <CardTitle className="text-sm">{title}</CardTitle>
        <CardDescription className="text-xs">{subtitle}</CardDescription>
        <CardAction>
          <Button
            variant="ghost"
            size="xs"
            onClick={onViewAll}
            className="text-muted-foreground hover:text-foreground"
          >
            {viewAllLabel}
            <span aria-hidden>→</span>
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="px-4 pb-4 sm:px-6 sm:pb-5">
        {isLoading ? (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-5">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="aspect-square w-full rounded-md" />
            ))}
          </div>
        ) : items.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            {emptyLabel}
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-5">
            {items.map((file) => {
              const Icon = fileTypeIcon(file.file_type);
              const thumb =
                file.file_type === "image" ? file.thumb_url || file.url : null;
              return (
                <div
                  key={file.id}
                  className="group flex min-w-0 flex-col gap-2"
                >
                  <div className="relative flex aspect-square items-center justify-center overflow-hidden rounded-md border border-border/60 bg-muted">
                    {thumb ? (
                      <img
                        src={thumb}
                        alt={file.original_name}
                        loading="lazy"
                        className="h-full w-full object-cover transition-transform group-hover:scale-[1.03]"
                      />
                    ) : (
                      <Icon
                        className="size-6 text-muted-foreground"
                        strokeWidth={1.6}
                      />
                    )}
                  </div>
                  <div className="min-w-0">
                    <div className="truncate text-xs text-foreground">
                      {file.original_name}
                    </div>
                    <div className="mt-0.5 truncate text-[10px] text-muted-foreground tabular-nums">
                      {formatSize(file.size_bytes)} ·{" "}
                      {formatRelativeTime(file.created_at, locale)}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
