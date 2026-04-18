import { useEffect, useMemo, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, FolderOpen, Pencil, ChevronLeft, ChevronRight, FolderTree, ArrowLeft } from "lucide-react";
import { useSearchParams } from "react-router-dom";

import { albumApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";

interface FolderItem {
  id: string;
  name: string;
  description: string;
  is_public: boolean;
  file_count: number;
  folder_count: number;
  parent_id?: string;
  created_at: string;
}

interface FolderListData {
  items: FolderItem[];
  total: number;
  page: number;
  size: number;
  current_folder?: FolderItem;
  ancestors?: FolderItem[];
}

export default function AlbumsPage() {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editingFolder, setEditingFolder] = useState<FolderItem | null>(null);
  const [form, setForm] = useState({ name: "", description: "", is_public: false });
  const currentFolderId = searchParams.get("parent_id") || "";

  const { data, isLoading } = useQuery({
    queryKey: ["folders", page, currentFolderId],
    queryFn: () =>
      albumApi
        .list({ page, size: 20, ...(currentFolderId ? { parent_id: currentFolderId } : {}) })
        .then((r) => r.data.data as FolderListData),
  });

  useEffect(() => {
    setPage(1);
  }, [currentFolderId]);

  const breadcrumbItems = useMemo(() => {
    const ancestors = data?.ancestors ?? [];
    return [{ id: "root", name: t("albums.root") }, ...ancestors];
  }, [data?.ancestors, t]);

  const createMutation = useMutation({
    mutationFn: () => albumApi.create({ ...form, parent_id: currentFolderId || undefined }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["folders"] });
      setCreateOpen(false);
      setForm({ name: "", description: "", is_public: false });
      toast.success(t("albums.createSuccess"));
    },
    onError: () => toast.error(t("albums.createFailed")),
  });

  const updateMutation = useMutation({
    mutationFn: () => albumApi.update(editingFolder!.id, form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["folders"] });
      setEditOpen(false);
      setEditingFolder(null);
      toast.success(t("albums.updateSuccess"));
    },
    onError: () => toast.error(t("albums.updateFailed")),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => albumApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["folders"] });
      toast.success(t("albums.deleteSuccess"));
    },
    onError: () => toast.error(t("albums.deleteFailed")),
  });

  const openEdit = (folder: FolderItem) => {
    setEditingFolder(folder);
    setForm({ name: folder.name, description: folder.description, is_public: folder.is_public });
    setEditOpen(true);
  };

  const openFolder = (folderId: string) => {
    setSearchParams(folderId ? { parent_id: folderId } : {});
  };

  const goToAncestor = (folderId?: string) => {
    setSearchParams(folderId ? { parent_id: folderId } : {});
  };

  const goUp = () => {
    const ancestors = data?.ancestors ?? [];
    if (ancestors.length <= 1) {
      setSearchParams({});
      return;
    }
    const parent = ancestors[ancestors.length - 2];
    setSearchParams(parent ? { parent_id: parent.id } : {});
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("albums.title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t("albums.description")}</p>

          <div className="mt-3 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            {breadcrumbItems.map((item, index) => {
              const isLast = index === breadcrumbItems.length - 1;
              const folderId = item.id === "root" ? undefined : item.id;
              return (
                <div key={item.id} className="flex items-center gap-2">
                  {index > 0 && <ChevronRight className="size-3 text-muted-foreground/70" />}
                  <button
                    type="button"
                    className={isLast ? "rounded-lg bg-muted px-2 py-1 font-medium text-foreground" : "rounded-lg px-2 py-1 transition-colors hover:bg-muted hover:text-foreground"}
                    onClick={() => goToAncestor(folderId)}
                  >
                    {item.name}
                  </button>
                </div>
              );
            })}
          </div>
        </div>
        <div className="flex items-center gap-2 self-start">
          {currentFolderId && (
            <Button variant="outline" onClick={goUp}>
              <ArrowLeft className="size-4" />
              {t("albums.up")}
            </Button>
          )}
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="size-4" />
                {t("albums.newAlbum")}
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>{t("albums.createAlbum")}</DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>{t("common.name")}</Label>
                  <Input
                    value={form.name}
                    onChange={(e) => setForm((a) => ({ ...a, name: e.target.value }))}
                    placeholder={t("albums.albumName")}
                  />
                </div>
                <div className="space-y-2">
                  <Label>{t("albums.albumDescLabel")}</Label>
                  <Input
                    value={form.description}
                    onChange={(e) => setForm((a) => ({ ...a, description: e.target.value }))}
                    placeholder={t("albums.albumDesc")}
                  />
                </div>
                <div className="rounded-lg border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
                  <span className="font-medium text-foreground">{t("albums.current")}</span>
                  <span className="ml-2">{data?.current_folder?.name ?? t("albums.root")}</span>
                </div>
                <div className="flex items-center justify-between">
                  <Label>{t("albums.publicAlbum")}</Label>
                  <Button
                    type="button"
                    size="sm"
                    variant={form.is_public ? "default" : "outline"}
                    onClick={() => setForm((a) => ({ ...a, is_public: !a.is_public }))}
                  >
                    {form.is_public ? t("common.public") : t("albums.private")}
                  </Button>
                </div>
              </div>
              <DialogFooter>
                <Button
                  onClick={() => createMutation.mutate()}
                  disabled={!form.name || createMutation.isPending}
                >
                  {createMutation.isPending ? t("albums.creating") : t("common.create")}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-lg" />
          ))}
        </div>
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {data?.items?.map((folder: FolderItem) => (
              <div
                key={folder.id}
                className="group relative flex cursor-pointer flex-col rounded-xl border bg-card p-4 transition-colors hover:border-primary/30 hover:bg-accent/20"
                onClick={() => openFolder(folder.id)}
              >
                <div className="mb-3 flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <FolderTree className="size-5 text-muted-foreground" />
                    <h3 className="font-medium">{folder.name}</h3>
                  </div>
                  <div className="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                    <Button size="icon-xs" variant="ghost" onClick={(event) => { event.stopPropagation(); openEdit(folder); }}>
                      <Pencil />
                    </Button>
                    <Button
                      size="icon-xs"
                      variant="ghost"
                      onClick={(event) => { event.stopPropagation(); deleteMutation.mutate(folder.id); }}
                    >
                      <Trash2 className="text-destructive" />
                    </Button>
                  </div>
                </div>
                {folder.description && (
                  <p className="mb-2 text-sm text-muted-foreground line-clamp-2">
                    {folder.description}
                  </p>
                )}
                <div className="mt-auto flex flex-wrap items-center gap-2">
                  <span className="text-xs text-muted-foreground">
                    {folder.folder_count} {t("albums.folders")}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {folder.file_count} {t("albums.files")}
                  </span>
                  {folder.is_public && (
                    <Badge variant="outline" className="text-[10px]">
                      {t("common.public")}
                    </Badge>
                  )}
                </div>
              </div>
            ))}
          </div>

          {data?.items?.length === 0 && (
            <div className="flex flex-col items-center py-20 text-center">
              <div className="flex size-14 items-center justify-center rounded-full bg-muted">
                <FolderOpen className="size-6 text-muted-foreground" />
              </div>
              <p className="mt-4 text-sm font-medium text-muted-foreground">{t("albums.noAlbums")}</p>
            </div>
          )}

          {data && data.total > 20 && (
            <div className="flex items-center justify-center gap-2">
              <Button variant="outline" size="icon-sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                <ChevronLeft className="size-4" />
              </Button>
              <span className="min-w-15 text-center text-sm text-muted-foreground">
                {page} / {Math.ceil(data.total / 20)}
              </span>
              <Button variant="outline" size="icon-sm" disabled={page >= Math.ceil(data.total / 20)} onClick={() => setPage((p) => p + 1)}>
                <ChevronRight className="size-4" />
              </Button>
            </div>
          )}
        </>
      )}

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={(open) => { if (!open) { setEditOpen(false); setEditingFolder(null); } }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("albums.editAlbum")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>{t("common.name")}</Label>
              <Input
                value={form.name}
                onChange={(e) => setForm((a) => ({ ...a, name: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label>{t("albums.albumDescLabel")}</Label>
              <Input
                value={form.description}
                onChange={(e) => setForm((a) => ({ ...a, description: e.target.value }))}
              />
            </div>
            <div className="flex items-center justify-between">
              <Label>{t("albums.publicAlbum")}</Label>
              <Button
                type="button"
                size="sm"
                variant={form.is_public ? "default" : "outline"}
                onClick={() => setForm((a) => ({ ...a, is_public: !a.is_public }))}
              >
                {form.is_public ? t("common.public") : t("albums.private")}
              </Button>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setEditOpen(false); setEditingFolder(null); }}>
              {t("common.cancel")}
            </Button>
            <Button
              onClick={() => updateMutation.mutate()}
              disabled={!form.name || updateMutation.isPending}
            >
              {updateMutation.isPending ? t("common.loading") : t("common.save")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
