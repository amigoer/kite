import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, FolderOpen, Pencil } from "lucide-react";
import { albumApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";

interface Album {
  id: string;
  name: string;
  description: string;
  is_public: boolean;
  file_count: number;
  created_at: string;
}

export default function AlbumsPage() {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editingAlbum, setEditingAlbum] = useState<Album | null>(null);
  const [form, setForm] = useState({ name: "", description: "", is_public: false });

  const { data, isLoading } = useQuery({
    queryKey: ["albums", page],
    queryFn: () => albumApi.list({ page, size: 20 }).then((r) => r.data.data),
  });

  const createMutation = useMutation({
    mutationFn: () => albumApi.create(form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["albums"] });
      setCreateOpen(false);
      setForm({ name: "", description: "", is_public: false });
    },
  });

  const updateMutation = useMutation({
    mutationFn: () => albumApi.update(editingAlbum!.id, form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["albums"] });
      setEditOpen(false);
      setEditingAlbum(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => albumApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["albums"] }),
  });

  const openEdit = (album: Album) => {
    setEditingAlbum(album);
    setForm({ name: album.name, description: album.description, is_public: album.is_public });
    setEditOpen(true);
  };

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("albums.title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t("albums.description")}</p>
        </div>
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

      <Separator />

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-lg" />
          ))}
        </div>
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {data?.items?.map((album: Album) => (
              <div
                key={album.id}
                className="group relative flex flex-col rounded-lg border bg-card p-4 transition-shadow hover:shadow-md"
              >
                <div className="mb-3 flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <FolderOpen className="size-5 text-muted-foreground" />
                    <h3 className="font-medium">{album.name}</h3>
                  </div>
                  <div className="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                    <Button size="icon-xs" variant="ghost" onClick={() => openEdit(album)}>
                      <Pencil />
                    </Button>
                    <Button
                      size="icon-xs"
                      variant="ghost"
                      onClick={() => deleteMutation.mutate(album.id)}
                    >
                      <Trash2 className="text-destructive" />
                    </Button>
                  </div>
                </div>
                {album.description && (
                  <p className="mb-2 text-sm text-muted-foreground line-clamp-2">
                    {album.description}
                  </p>
                )}
                <div className="mt-auto flex items-center gap-2">
                  <span className="text-xs text-muted-foreground">
                    {album.file_count} {t("albums.files")}
                  </span>
                  {album.is_public && (
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
              <div className="mb-4 flex size-14 items-center justify-center rounded-full bg-muted">
                <FolderOpen className="size-6 text-muted-foreground" />
              </div>
              <p className="text-sm font-medium text-muted-foreground">{t("albums.noAlbums")}</p>
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

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={(open) => { if (!open) { setEditOpen(false); setEditingAlbum(null); } }}>
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
            <Button variant="outline" onClick={() => { setEditOpen(false); setEditingAlbum(null); }}>
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
