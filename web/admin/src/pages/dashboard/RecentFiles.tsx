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
  image: { icon: Image, accent: "bg-amber-50 text-amber-600 dark:bg-amber-950/40 dark:text-amber-400" },
  video: { icon: Video, accent: "bg-violet-50 text-violet-600 dark:bg-violet-950/40 dark:text-violet-400" },
  audio: { icon: Music, accent: "bg-blue-50 text-blue-600 dark:bg-blue-950/40 dark:text-blue-400" },
  file: { icon: FileText, accent: "bg-muted text-muted-foreground" },
};

function getExtension(mimeType: string): string {
  const map: Record<string, string> = {
    "image/png": "PNG", "image/jpeg": "JPEG", "image/webp": "WebP",
    "image/gif": "GIF", "image/svg+xml": "SVG", "video/mp4": "MP4",
    "video/webm": "WebM", "audio/mpeg": "MP3", "audio/wav": "WAV",
    "audio/ogg": "OGG", "application/zip": "ZIP", "application/pdf": "PDF",
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
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-20" />
          <Skeleton className="h-4 w-28" />
        </CardHeader>
        <CardContent>
          <div className="space-y-0">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 py-3">
                <Skeleton className="size-9 shrink-0 rounded-lg" />
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
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-semibold">
          {t("dashboard.recentUploads")}
        </CardTitle>
        <CardDescription>{t("dashboard.latestFiles")}</CardDescription>
      </CardHeader>
      <CardContent>
        {files.length === 0 ? (
          <div className="flex flex-col items-center py-10 text-center">
            <div className="mb-3 flex size-12 items-center justify-center rounded-full bg-muted">
              <Upload size={20} className="text-muted-foreground" />
            </div>
            <p className="text-sm text-muted-foreground">
              {t("dashboard.noFilesYet")}
            </p>
          </div>
        ) : (
          <div className="space-y-0">
            {files.map((file) => {
              const config = fileTypeConfig[file.file_type] ?? fileTypeConfig.file;
              const Icon = config.icon;
              return (
                <div
                  key={file.id}
                  className="flex items-center gap-3 border-b border-border/40 py-3 last:border-b-0 last:pb-0 first:pt-0"
                >
                  <div className={`flex size-9 shrink-0 items-center justify-center rounded-lg ${config.accent}`}>
                    <Icon size={15} />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{file.original_name}</p>
                    <p className="text-xs text-muted-foreground">
                      {formatSize(file.size_bytes)} · {getExtension(file.mime_type)}
                    </p>
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
  );
}
