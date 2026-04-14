import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useI18n } from "@/i18n";
import { formatSize, calcPercent } from "@/lib/utils";

interface FileTypeChartProps {
  totalFiles: number;
  imageCount: number;
  videoCount: number;
  audioCount: number;
  otherCount: number;
  storageUsed: number;
  isLoading: boolean;
}

export default function FileTypeChart({
  totalFiles,
  imageCount,
  videoCount,
  audioCount,
  otherCount,
  storageUsed,
  isLoading,
}: FileTypeChartProps) {
  const { t } = useI18n();

  const fileTypes = [
    { name: t("dashboard.images"), count: imageCount, color: "bg-amber-500" },
    { name: t("dashboard.videos"), count: videoCount, color: "bg-violet-500" },
    { name: t("dashboard.audio"), count: audioCount, color: "bg-blue-500" },
    { name: t("dashboard.otherFiles"), count: otherCount, color: "bg-gray-400 dark:bg-gray-500" },
  ];

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-28" />
          <Skeleton className="h-4 w-36" />
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-5 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-semibold">
          {t("dashboard.fileTypeDistribution")}
        </CardTitle>
        <CardDescription>{t("dashboard.byFileType")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div className="space-y-0">
          {fileTypes.map((ft) => (
            <div
              key={ft.name}
              className="flex items-center gap-3 border-b border-border/40 py-3 last:border-b-0 last:pb-0 first:pt-0"
            >
              <span className={`size-2 shrink-0 rounded-full ${ft.color}`} />
              <span className="min-w-[3.5rem] text-sm">{ft.name}</span>
              <div className="mx-1 h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                <div
                  className={`h-full rounded-full transition-all duration-500 ${ft.color}`}
                  style={{
                    width: `${totalFiles > 0 ? (ft.count / totalFiles) * 100 : 0}%`,
                  }}
                />
              </div>
              <span className="min-w-[2.5rem] text-right text-sm font-semibold tabular-nums">
                {ft.count.toLocaleString()}
              </span>
              <span className="min-w-[2.5rem] text-right text-xs text-muted-foreground tabular-nums">
                {calcPercent(ft.count, totalFiles)}%
              </span>
            </div>
          ))}
        </div>

        {/* Storage bar */}
        <div className="rounded-lg bg-muted/50 p-4">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">{t("dashboard.used")}</span>
            <span className="font-semibold">{formatSize(storageUsed)}</span>
          </div>
          <div className="mt-2.5 h-2 overflow-hidden rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-primary transition-all duration-500"
              style={{ width: storageUsed > 0 ? "100%" : "0%" }}
            />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
