import { useQuery } from "@tanstack/react-query";
import { Image, Video, Music, FileText, Upload } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useI18n } from "@/i18n";
import { fileApi } from "@/lib/api";
import { formatSize, formatRelativeTime } from "@/lib/utils";

interface FileItem {
  id: string;
  original_name: string;
  size_bytes: number;
  mime_type: string;
  file_type: "image" | "video" | "audio" | "file";
  url: string;
  thumb_url: string | null;
  created_at: string;
}

const fileTypeConfig = {
  image: { icon: Image, bg: "bg-amber-50 dark:bg-amber-950/30", color: "text-amber-600 dark:text-amber-400" },
  video: { icon: Video, bg: "bg-violet-50 dark:bg-violet-950/30", color: "text-violet-600 dark:text-violet-400" },
  audio: { icon: Music, bg: "bg-blue-50 dark:bg-blue-950/30", color: "text-blue-600 dark:text-blue-400" },
  file: { icon: FileText, bg: "bg-gray-100 dark:bg-gray-800/50", color: "text-gray-500 dark:text-gray-400" },
};

function getExtension(mimeType: string): string {
  const map: Record<string, string> = {
    "image/png": "PNG",
    "image/jpeg": "JPEG",
    "image/webp": "WebP",
    "image/gif": "GIF",
    "image/svg+xml": "SVG",
    "video/mp4": "MP4",
    "video/webm": "WebM",
    "audio/mpeg": "MP3",
    "audio/wav": "WAV",
    "audio/ogg": "OGG",
    "application/zip": "ZIP",
    "application/pdf": "PDF",
  };
  return map[mimeType] ?? mimeType.split("/").pop()?.toUpperCase() ?? "";
}

export default function RecentFiles() {
  const { t, locale } = useI18n();

  const { data, isLoading } = useQuery({
    queryKey: ["files", "recent"],
    queryFn: () =>
      fileApi
        .list({ page: 1, size: 5, sort: "created_at", order: "desc" })
        .then((r) => r.data.data),
    staleTime: 15_000,
  });

  if (isLoading) {
    return (
      <Card className="border-0 shadow-sm rounded-2xl overflow-hidden">
        <CardHeader className="pt-5 pb-0">
          <Skeleton className="h-5 w-20" />
          <Skeleton className="h-4 w-28 mt-1" />
        </CardHeader>
        <CardContent className="pt-4 pb-5">
          <div className="space-y-0">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 py-2.5">
                <Skeleton className="h-8 w-8 rounded-lg shrink-0" />
                <div className="flex-1 space-y-1.5">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/3" />
                </div>
                <Skeleton className="h-3 w-12" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  const files: FileItem[] = data?.items ?? [];

  return (
    <Card className="border-0 shadow-sm rounded-2xl overflow-hidden">
      <CardHeader className="pt-5 pb-0">
        <CardTitle className="text-sm font-medium">
          {t("dashboard.recentUploads")}
        </CardTitle>
        <CardDescription className="text-xs">
          {t("dashboard.latestFiles")}
        </CardDescription>
      </CardHeader>
      <CardContent className="pt-4 pb-5">
        {files.length === 0 ? (
          <div className="flex flex-col items-center py-10 text-center">
            <Upload size={32} className="mb-3 text-muted-foreground/30" />
            <p className="text-sm text-muted-foreground">
              {t("dashboard.noFilesYet")}
            </p>
          </div>
        ) : (
          <div className="space-y-0">
            {files.map((file) => {
              const config =
                fileTypeConfig[file.file_type] ?? fileTypeConfig.file;
              const Icon = config.icon;
              return (
                <div
                  key={file.id}
                  className="flex items-center gap-3 border-b border-border/50 py-2.5 last:border-b-0 last:pb-0"
                >
                  <div
                    className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ${config.bg}`}
                  >
                    <Icon size={14} className={config.color} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="truncate text-[13px] text-foreground">
                      {file.original_name}
                    </p>
                    <p className="text-[11px] text-muted-foreground">
                      {formatSize(file.size_bytes)} &middot;{" "}
                      {getExtension(file.mime_type)}
                    </p>
                  </div>
                  <span className="shrink-0 text-[11px] text-muted-foreground/70">
                    {formatRelativeTime(file.created_at, locale)}
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
