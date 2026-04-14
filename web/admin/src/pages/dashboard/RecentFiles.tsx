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

const fileTypeIcons: Record<string, typeof FileText> = {
  image: Image,
  video: Video,
  audio: Music,
  file: FileText,
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
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-20" />
          <Skeleton className="h-4 w-28" />
        </CardHeader>
        <CardContent>
          <div className="space-y-0">
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="flex items-center gap-3 border-b py-3 last:border-b-0"
              >
                <Skeleton className="h-4 w-4 shrink-0" />
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
    <Card className="col-span-1">
      <CardHeader>
        <CardTitle className="text-sm font-medium">
          {t("dashboard.recentUploads")}
        </CardTitle>
        <CardDescription>{t("dashboard.latestFiles")}</CardDescription>
      </CardHeader>
      <CardContent>
        {files.length === 0 ? (
          <div className="flex flex-col items-center py-8 text-center">
            <div className="flex size-12 items-center justify-center rounded-full bg-muted">
              <Upload className="size-5 text-muted-foreground" />
            </div>
            <p className="mt-3 text-sm font-medium text-muted-foreground">
              {t("dashboard.noFilesYet")}
            </p>
          </div>
        ) : (
          <div className="space-y-0">
            {files.map((file) => {
              const Icon = fileTypeIcons[file.file_type] ?? FileText;
              return (
                <div
                  key={file.id}
                  className="flex items-center gap-4 py-3 -mx-2 px-2 hover:bg-muted/50 rounded-lg transition-all duration-200 group border-b border-transparent hover:border-transparent first:pt-2 last:pb-2"
                >
                  <div className="p-2 bg-muted/40 rounded-md shrink-0">
                    <Icon className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium transition-colors">
                      {file.original_name}
                    </p>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      <span className="inline-block bg-muted px-1 rounded text-[10px] mr-1.5 font-medium">{getExtension(file.mime_type)}</span>
                      {formatSize(file.size_bytes)}
                    </p>
                  </div>
                  <span className="shrink-0 text-xs text-muted-foreground opacity-80 group-hover:opacity-100 transition-opacity">
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
