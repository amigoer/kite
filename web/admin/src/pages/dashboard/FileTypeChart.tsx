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

interface FileTypeData {
  name: string;
  count: number;
  color: string;
  barColor: string;
}

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

  const fileTypes: FileTypeData[] = [
    {
      name: t("dashboard.images"),
      count: imageCount,
      color: "bg-amber-500",
      barColor: "bg-amber-500",
    },
    {
      name: t("dashboard.videos"),
      count: videoCount,
      color: "bg-violet-500",
      barColor: "bg-violet-500",
    },
    {
      name: t("dashboard.audio"),
      count: audioCount,
      color: "bg-blue-500",
      barColor: "bg-blue-500",
    },
    {
      name: t("dashboard.otherFiles"),
      count: otherCount,
      color: "bg-gray-400",
      barColor: "bg-gray-400",
    },
  ];

  if (isLoading) {
    return (
      <Card className="border-0 shadow-sm rounded-2xl overflow-hidden">
        <CardHeader className="pt-5 pb-0">
          <Skeleton className="h-5 w-28" />
          <Skeleton className="h-4 w-36 mt-1" />
        </CardHeader>
        <CardContent className="pt-4 pb-5">
          <div className="space-y-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-5 w-full" />
            ))}
          </div>
          <Skeleton className="mt-6 h-2 w-full rounded-full" />
          <div className="mt-3 flex justify-between">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-16" />
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 shadow-sm rounded-2xl overflow-hidden">
      <CardHeader className="pt-5 pb-0">
        <CardTitle className="text-sm font-medium">
          {t("dashboard.fileTypeDistribution")}
        </CardTitle>
        <CardDescription className="text-xs">
          {t("dashboard.byFileType")}
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-4 pb-5">
        <div className="space-y-0">
          {fileTypes.map((ft) => (
            <div
              key={ft.name}
              className="flex items-center gap-2.5 border-b border-border/50 py-2.5 last:border-b-0 last:pb-0"
            >
              <span
                className={`h-2 w-2 shrink-0 rounded-full ${ft.color}`}
              />
              <span className="min-w-[52px] text-[13px] text-foreground">
                {ft.name}
              </span>
              <div className="mx-2 flex-1 overflow-hidden rounded-full bg-muted/60 h-1">
                <div
                  className={`h-full rounded-full ${ft.barColor}`}
                  style={{
                    width: `${totalFiles > 0 ? (ft.count / totalFiles) * 100 : 0}%`,
                  }}
                />
              </div>
              <span className="min-w-[40px] text-right text-[13px] font-medium text-foreground tabular-nums">
                {ft.count.toLocaleString()}
              </span>
              <span className="min-w-[40px] text-right text-[11px] text-muted-foreground tabular-nums">
                {calcPercent(ft.count, totalFiles)}%
              </span>
            </div>
          ))}
        </div>

        {/* Storage bar */}
        <div className="mt-5">
          <div className="h-2 overflow-hidden rounded-full bg-muted/60">
            <div
              className="h-full rounded-full bg-primary transition-all"
              style={{ width: storageUsed > 0 ? "100%" : "0%" }}
            />
          </div>
          <div className="mt-2.5 flex justify-between text-xs text-muted-foreground">
            <span>
              {t("dashboard.used")}{" "}
              <span className="font-medium text-foreground">
                {formatSize(storageUsed)}
              </span>
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
