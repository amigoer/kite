import { useQuery } from "@tanstack/react-query";
import { FileText, HardDrive, Image, Video, Upload } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { statsApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { useAuth } from "@/hooks/use-auth";
import { formatSize, calcPercent } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
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

export default function DashboardPage() {
  const { t } = useI18n();
  const { user } = useAuth();
  const navigate = useNavigate();

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

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex flex-row items-center justify-between gap-4">
        <div className="space-y-1">
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">
            {t("dashboard.title")}
          </h1>
          <p className="text-xs sm:text-sm text-muted-foreground line-clamp-1">
            {t("dashboard.description")}
          </p>
        </div>
        <Button size="sm" className="shrink-0" onClick={() => navigate("/files")}>
          <Upload className="mr-2 h-4 w-4 hidden sm:block" />
          <Upload className="h-4 w-4 sm:hidden" />
          <span className="hidden sm:inline-block ml-2">{t("common.upload")}</span>
          <span className="sm:hidden ml-2">{t("common.upload")}</span>
        </Button>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {t("dashboard.totalFiles")}
            </CardTitle>
            <div className="flex size-9 items-center justify-center rounded-lg border bg-background shadow-xs">
              <FileText className="h-4 w-4 text-foreground/70" />
            </div>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-16" />
            ) : (
              <>
                <div className="text-2xl font-bold tracking-tight">
                  {stats.total_files.toLocaleString()}
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  {user?.role === "admin"
                    ? `${stats.users} ${t("dashboard.users")}`
                    : `${stats.total_files.toLocaleString()} ${t("dashboard.filesTotal")}`}
                </p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {t("dashboard.storageUsed")}
            </CardTitle>
            <div className="flex size-9 items-center justify-center rounded-lg border bg-background shadow-xs">
              <HardDrive className="h-4 w-4 text-foreground/70" />
            </div>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <>
                <div className="text-2xl font-bold tracking-tight">
                  {formatSize(stats.total_size)}
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  {stats.total_files.toLocaleString()} {t("dashboard.filesTotal")}
                </p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {t("dashboard.images")}
            </CardTitle>
            <div className="flex size-9 items-center justify-center rounded-lg border bg-background shadow-xs">
              <Image className="h-4 w-4 text-foreground/70" />
            </div>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-12" />
            ) : (
              <>
                <div className="text-2xl font-bold tracking-tight">
                  {stats.images.toLocaleString()}
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  {t("dashboard.percentOfTotal")}{" "}
                  {calcPercent(stats.images, stats.total_files)}%
                </p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              {t("dashboard.videos")}
            </CardTitle>
            <div className="flex size-9 items-center justify-center rounded-lg border bg-background shadow-xs">
              <Video className="h-4 w-4 text-foreground/70" />
            </div>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Skeleton className="h-8 w-12" />
            ) : (
              <>
                <div className="text-2xl font-bold tracking-tight">
                  {stats.videos.toLocaleString()}
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  {t("dashboard.percentOfTotal")}{" "}
                  {calcPercent(stats.videos, stats.total_files)}%
                </p>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Bottom grid */}
      <div className="grid gap-4 lg:grid-cols-2">
        <FileTypeChart
          totalFiles={stats.total_files}
          imageCount={stats.images}
          videoCount={stats.videos}
          audioCount={stats.audios}
          otherCount={stats.others}
          isLoading={isLoading}
        />
        <RecentFiles />
      </div>
    </div>
  );
}
