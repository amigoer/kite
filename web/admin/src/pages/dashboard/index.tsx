import { useQuery } from "@tanstack/react-query";
import { FileText, HardDrive, Image, Video } from "lucide-react";
import { statsApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { cn, formatSize, calcPercent } from "@/lib/utils";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
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
  iconBg: string;
  iconColor: string;
  subtitleClass?: string;
  isLoading: boolean;
}

function StatCard({
  label,
  value,
  subtitle,
  icon: Icon,
  iconBg,
  iconColor,
  subtitleClass,
  isLoading,
}: StatCardProps) {
  return (
    <Card className="border-0 shadow-sm rounded-2xl overflow-hidden cursor-pointer transition-shadow hover:shadow-md">
      <CardHeader className="flex flex-row items-center justify-between pb-0 pt-4 px-5">
        <span className="text-xs text-muted-foreground">{label}</span>
        <div
          className={cn(
            "flex h-7 w-7 items-center justify-center rounded-lg",
            iconBg
          )}
        >
          <Icon size={14} className={iconColor} />
        </div>
      </CardHeader>
      <CardContent className="px-5 pb-4 pt-3">
        {isLoading ? (
          <>
            <Skeleton className="h-8 w-24" />
            <Skeleton className="mt-2 h-3.5 w-32" />
          </>
        ) : (
          <>
            <div className="text-2xl font-semibold tracking-tight text-foreground">
              {typeof value === "number" ? value.toLocaleString() : value}
            </div>
            <p
              className={cn(
                "mt-1.5 text-[11px] text-muted-foreground",
                subtitleClass
              )}
            >
              {subtitle}
            </p>
          </>
        )}
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
      iconBg: "bg-muted",
      iconColor: "text-foreground",
      subtitle: `${stats.users} ${t("dashboard.users")}`,
    },
    {
      label: t("dashboard.storageUsed"),
      value: formatSize(stats.total_size),
      icon: HardDrive,
      iconBg: "bg-blue-50 dark:bg-blue-950/30",
      iconColor: "text-blue-600 dark:text-blue-400",
      subtitle: `${stats.total_files.toLocaleString()} ${t("dashboard.filesTotal")}`,
    },
    {
      label: t("dashboard.images"),
      value: stats.images,
      icon: Image,
      iconBg: "bg-amber-50 dark:bg-amber-950/30",
      iconColor: "text-amber-600 dark:text-amber-400",
      subtitle: `${t("dashboard.percentOfTotal")} ${calcPercent(stats.images, stats.total_files)}%`,
    },
    {
      label: t("dashboard.videos"),
      value: stats.videos,
      icon: Video,
      iconBg: "bg-violet-50 dark:bg-violet-950/30",
      iconColor: "text-violet-600 dark:text-violet-400",
      subtitle: `${t("dashboard.percentOfTotal")} ${calcPercent(stats.videos, stats.total_files)}%`,
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-xl font-medium text-foreground">
          {t("dashboard.title")}
        </h1>
        <p className="mt-1 text-[13px] text-muted-foreground">
          {t("dashboard.description")} &middot; {today}
        </p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {statCards.map((card) => (
          <StatCard key={card.label} {...card} isLoading={isLoading} />
        ))}
      </div>

      {/* Bottom grid */}
      <div className="grid gap-4 lg:grid-cols-2">
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
