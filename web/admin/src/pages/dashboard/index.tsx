import { useQuery } from "@tanstack/react-query";
import { FileText, HardDrive, Image, Video, TrendingUp } from "lucide-react";
import { statsApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { formatSize, calcPercent } from "@/lib/utils";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import FileTypeChart from "./FileTypeChart";
import RecentFiles from "./RecentFiles";

interface DashboardStats {
  total_files: number;
  total_size: number;
  images: number;
  videos: number;
  audios: number;
  others: number;
  users: number;
}

interface StatCardProps {
  label: string;
  value: string | number;
  subtitle: string;
  icon: React.ElementType;
  accent: string;
  isLoading: boolean;
}

function StatCard({ label, value, subtitle, icon: Icon, accent, isLoading }: StatCardProps) {
  return (
    <Card className="gap-0 py-0">
      <CardContent className="p-5">
        <div className="flex items-center justify-between">
          <span className="text-[13px] font-medium text-muted-foreground">{label}</span>
          <div className={`flex size-8 items-center justify-center rounded-lg ${accent}`}>
            <Icon size={15} />
          </div>
        </div>
        {isLoading ? (
          <Skeleton className="mt-3 h-8 w-20" />
        ) : (
          <div className="mt-3 text-[28px] font-bold leading-none tracking-tight">
            {typeof value === "number" ? value.toLocaleString() : value}
          </div>
        )}
        <p className="mt-2 text-[12px] text-muted-foreground">{subtitle}</p>
      </CardContent>
    </Card>
  );
}

export default function DashboardPage() {
  const { t, locale } = useI18n();

  const { data, isLoading } = useQuery<DashboardStats>({
    queryKey: ["dashboard", "stats"],
    queryFn: () => statsApi.get().then((r) => r.data.data),
    staleTime: 30_000,
    refetchInterval: 60_000,
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

  const today = new Date().toLocaleDateString(
    locale === "zh" ? "zh-CN" : "en-US",
    { year: "numeric", month: "long", day: "numeric" }
  );

  const statCards: Omit<StatCardProps, "isLoading">[] = [
    {
      label: t("dashboard.totalFiles"),
      value: stats.total_files,
      icon: FileText,
      accent: "bg-muted text-foreground",
      subtitle: `${stats.users} ${t("dashboard.users")}`,
    },
    {
      label: t("dashboard.storageUsed"),
      value: formatSize(stats.total_size),
      icon: HardDrive,
      accent: "bg-blue-50 text-blue-600 dark:bg-blue-950/40 dark:text-blue-400",
      subtitle: `${stats.total_files.toLocaleString()} ${t("dashboard.filesTotal")}`,
    },
    {
      label: t("dashboard.images"),
      value: stats.images,
      icon: Image,
      accent: "bg-amber-50 text-amber-600 dark:bg-amber-950/40 dark:text-amber-400",
      subtitle: `${t("dashboard.percentOfTotal")} ${calcPercent(stats.images, stats.total_files)}%`,
    },
    {
      label: t("dashboard.videos"),
      value: stats.videos,
      icon: Video,
      accent: "bg-violet-50 text-violet-600 dark:bg-violet-950/40 dark:text-violet-400",
      subtitle: `${t("dashboard.percentOfTotal")} ${calcPercent(stats.videos, stats.total_files)}%`,
    },
  ];

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("dashboard.title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {t("dashboard.description")} · {today}
          </p>
        </div>
        <div className="hidden items-center gap-1.5 text-xs text-muted-foreground sm:flex">
          <TrendingUp size={14} />
          <span>{t("dashboard.totalFiles")}: {stats.total_files.toLocaleString()}</span>
        </div>
      </div>

      <Separator />

      {/* Stats grid */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {statCards.map((card) => (
          <StatCard key={card.label} {...card} isLoading={isLoading} />
        ))}
      </div>

      {/* Bottom grid */}
      <div className="grid gap-6 lg:grid-cols-2">
        <FileTypeChart
          totalFiles={stats.total_files}
          imageCount={stats.images}
          videoCount={stats.videos}
          audioCount={stats.audios}
          otherCount={stats.others}
          storageUsed={stats.total_size}
          isLoading={isLoading}
        />
        <RecentFiles />
      </div>
    </div>
  );
}
