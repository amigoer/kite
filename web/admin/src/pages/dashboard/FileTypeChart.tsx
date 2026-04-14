import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useI18n } from "@/i18n";
import { calcPercent } from "@/lib/utils";

interface FileTypeChartProps {
  totalFiles: number;
  imageCount: number;
  videoCount: number;
  audioCount: number;
  otherCount: number;
  isLoading: boolean;
}

export default function FileTypeChart({
  totalFiles,
  imageCount,
  videoCount,
  audioCount,
  otherCount,
  isLoading,
}: FileTypeChartProps) {
  const { t } = useI18n();

  const fileTypes = [
    { name: t("dashboard.images"), count: imageCount, color: "#f59e0b" },
    { name: t("dashboard.videos"), count: videoCount, color: "#8b5cf6" },
    { name: t("dashboard.audio"), count: audioCount, color: "#3b82f6" },
    { name: t("dashboard.otherFiles"), count: otherCount, color: "#6b7280" },
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
    <Card className="col-span-1">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          {t("dashboard.fileTypeDistribution")}
        </CardTitle>
        <CardDescription>{t("dashboard.byFileType")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {fileTypes.map((ft) => (
          <div key={ft.name} className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center gap-2">
                <span
                  className="size-2 shrink-0 rounded-full"
                  style={{ backgroundColor: ft.color }}
                />
                <span>{ft.name}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="font-medium tabular-nums">
                  {ft.count.toLocaleString()}
                </span>
                <span className="text-xs text-muted-foreground tabular-nums">
                  {calcPercent(ft.count, totalFiles)}%
                </span>
              </div>
            </div>
            <div className="h-2 overflow-hidden rounded-full bg-secondary">
              <div
                className="h-full rounded-full transition-all duration-500"
                style={{
                  backgroundColor: ft.color,
                  width: `${totalFiles > 0 ? (ft.count / totalFiles) * 100 : 0}%`,
                }}
              />
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
