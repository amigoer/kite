import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Trash2,
  Image,
  Video,
  Music,
  FileText,
  Search,
  Copy,
  Check,
  ExternalLink,
} from "lucide-react";
import { adminFileApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";

const typeIcons: Record<string, typeof FileText> = {
  image: Image,
  video: Video,
  audio: Music,
  file: FileText,
};

function formatBytes(bytes: number) {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

interface FileItem {
  id: string;
  user_id: string;
  original_name: string;
  file_type: string;
  mime_type: string;
  size_bytes: number;
  url: string;
  thumb_url?: string;
  created_at: string;
  width?: number;
  height?: number;
}

export default function AdminFilesPage() {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState("");
  const [fileType, setFileType] = useState("");
  const [detailFile, setDetailFile] = useState<FileItem | null>(null);
  const [copied, setCopied] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["admin-files", page, keyword, fileType],
    queryFn: () =>
      adminFileApi
        .list({ page, size: 20, keyword, file_type: fileType })
        .then((r) => r.data.data),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminFileApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["admin-files"] }),
  });

  const copyUrl = (text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopied(key);
    setTimeout(() => setCopied(""), 1500);
  };

  const types = [
    { value: "", labelKey: "common.all" },
    { value: "image", labelKey: "files.images" },
    { value: "video", labelKey: "files.videos" },
    { value: "audio", labelKey: "files.audio" },
    { value: "file", labelKey: "files.otherFiles" },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t("files.adminTitle")}</h1>
        <p className="text-sm text-muted-foreground">{t("files.adminDescription")}</p>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row flex-wrap items-start sm:items-center gap-3">
        <div className="relative w-full sm:flex-1 sm:max-w-xs">
          <Search className="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
          <Input
            placeholder={t("files.searchFiles")}
            value={keyword}
            onChange={(e) => { setKeyword(e.target.value); setPage(1); }}
            className="pl-9"
          />
        </div>
        <div className="flex gap-1 flex-wrap">
          {types.map((tp) => (
            <Button
              key={tp.value}
              variant={fileType === tp.value ? "default" : "outline"}
              size="sm"
              onClick={() => { setFileType(tp.value); setPage(1); }}
            >
              {t(tp.labelKey)}
            </Button>
          ))}
        </div>
      </div>

      {/* File Table */}
      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-14 rounded-lg" />
          ))}
        </div>
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden sm:block rounded-lg border">
            <div className="grid grid-cols-[2fr_1fr_auto_auto_auto] gap-4 border-b px-4 py-2 text-xs font-medium text-muted-foreground">
              <span>{t("files.fileName")}</span>
              <span>{t("files.uploader")}</span>
              <span>{t("common.type")}</span>
              <span>{t("common.size")}</span>
              <span />
            </div>
            {data?.items?.map((file: FileItem) => {
              const Icon = typeIcons[file.file_type] ?? FileText;
              return (
                <div
                  key={file.id}
                  className="grid grid-cols-[2fr_1fr_auto_auto_auto] items-center gap-4 border-b px-4 py-2.5 text-sm last:border-0 hover:bg-accent/30 cursor-pointer"
                  onClick={() => setDetailFile(file)}
                >
                  <div className="flex items-center gap-3 min-w-0">
                    {file.file_type === "image" && file.thumb_url ? (
                      <img src={file.thumb_url} className="size-8 rounded object-cover shrink-0" alt="" />
                    ) : (
                      <div className="flex size-8 items-center justify-center rounded bg-muted shrink-0">
                        <Icon className="size-4 text-muted-foreground" />
                      </div>
                    )}
                    <span className="truncate font-medium">{file.original_name}</span>
                  </div>
                  <span className="text-xs text-muted-foreground truncate">{file.user_id === "guest" ? t("files.guest") : file.user_id.slice(0, 8)}</span>
                  <Badge variant="secondary" className="text-[10px]">{file.file_type}</Badge>
                  <span className="text-xs text-muted-foreground whitespace-nowrap">{formatBytes(file.size_bytes)}</span>
                  <div className="flex gap-1">
                    <Button
                      size="icon-xs"
                      variant="ghost"
                      onClick={(e) => { e.stopPropagation(); copyUrl(file.url, file.id); }}
                    >
                      {copied === file.id ? <Check className="size-3" /> : <Copy className="size-3" />}
                    </Button>
                    <Button
                      size="icon-xs"
                      variant="ghost"
                      onClick={(e) => { e.stopPropagation(); deleteMutation.mutate(file.id); }}
                    >
                      <Trash2 className="size-3 text-destructive" />
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Mobile cards */}
          <div className="sm:hidden space-y-3">
            {data?.items?.map((file: FileItem) => {
              const Icon = typeIcons[file.file_type] ?? FileText;
              return (
                <div
                  key={file.id}
                  className="flex items-center gap-3 rounded-lg border bg-card p-3 cursor-pointer"
                  onClick={() => setDetailFile(file)}
                >
                  {file.file_type === "image" && file.thumb_url ? (
                    <img src={file.thumb_url} className="size-10 rounded-lg object-cover shrink-0" alt="" />
                  ) : (
                    <div className="flex size-10 items-center justify-center rounded-lg bg-muted shrink-0">
                      <Icon className="size-5 text-muted-foreground" />
                    </div>
                  )}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{file.original_name}</p>
                    <p className="text-xs text-muted-foreground">
                      {formatBytes(file.size_bytes)} · {file.file_type}
                    </p>
                  </div>
                  <Button
                    size="icon-xs"
                    variant="ghost"
                    onClick={(e) => { e.stopPropagation(); deleteMutation.mutate(file.id); }}
                  >
                    <Trash2 className="size-3.5 text-destructive" />
                  </Button>
                </div>
              );
            })}
          </div>

          {data?.items?.length === 0 && (
            <div className="flex flex-col items-center py-16 text-center">
              <FileText className="mb-3 size-12 text-muted-foreground/30" />
              <p className="text-sm text-muted-foreground">{t("files.noFiles")}</p>
            </div>
          )}

          {data && data.total > 20 && (
            <div className="flex items-center justify-center gap-2">
              <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                {t("common.previous")}
              </Button>
              <span className="text-sm text-muted-foreground">
                {t("common.page")} {page} {t("common.of")} {Math.ceil(data.total / 20)}
              </span>
              <Button variant="outline" size="sm" disabled={page >= Math.ceil(data.total / 20)} onClick={() => setPage((p) => p + 1)}>
                {t("common.next")}
              </Button>
            </div>
          )}
        </>
      )}

      {/* File Detail Dialog */}
      <Dialog open={!!detailFile} onOpenChange={(open) => !open && setDetailFile(null)}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle className="truncate pr-8">{detailFile?.original_name}</DialogTitle>
          </DialogHeader>
          {detailFile && (
            <div className="space-y-4">
              {detailFile.file_type === "image" && (
                <div className="rounded-lg border overflow-hidden bg-muted/30">
                  <img src={detailFile.url} alt={detailFile.original_name} className="w-full max-h-64 object-contain" />
                </div>
              )}
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-xs text-muted-foreground">{t("common.type")}</span>
                  <p className="font-medium">{detailFile.mime_type}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">{t("common.size")}</span>
                  <p className="font-medium">{formatBytes(detailFile.size_bytes)}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">{t("files.uploader")}</span>
                  <p className="font-medium">{detailFile.user_id === "guest" ? t("files.guest") : detailFile.user_id.slice(0, 8)}</p>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">{t("common.date")}</span>
                  <p className="font-medium">{new Date(detailFile.created_at).toLocaleString()}</p>
                </div>
              </div>
              <div className="space-y-2">
                <span className="text-xs font-medium text-muted-foreground">{t("files.linkFormats")}</span>
                {[
                  { label: "URL", value: detailFile.url },
                  { label: "Markdown", value: `![${detailFile.original_name}](${detailFile.url})` },
                  { label: "HTML", value: `<img src="${detailFile.url}" alt="${detailFile.original_name}">` },
                  { label: "BBCode", value: `[img]${detailFile.url}[/img]` },
                ].map((link) => (
                  <div key={link.label} className="flex items-center gap-2 bg-muted/50 rounded-md px-3 py-2">
                    <span className="text-xs text-muted-foreground w-16 shrink-0">{link.label}</span>
                    <input
                      type="text"
                      readOnly
                      value={link.value}
                      className="flex-1 bg-transparent text-xs outline-none truncate"
                      onClick={(e) => (e.target as HTMLInputElement).select()}
                    />
                    <Button size="icon-xs" variant="ghost" onClick={() => copyUrl(link.value, link.label)}>
                      {copied === link.label ? <Check className="size-3" /> : <Copy className="size-3" />}
                    </Button>
                  </div>
                ))}
              </div>
              <div className="flex gap-2 justify-end">
                <Button size="sm" variant="outline" asChild>
                  <a href={detailFile.url} target="_blank" rel="noopener noreferrer">
                    <ExternalLink className="size-3.5" />
                    {t("files.openOriginal")}
                  </a>
                </Button>
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => { deleteMutation.mutate(detailFile.id); setDetailFile(null); }}
                >
                  <Trash2 className="size-3.5" />
                  {t("common.delete")}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
