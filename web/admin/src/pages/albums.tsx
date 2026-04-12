import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2, FolderOpen, Pencil } from "lucide-react";
import { albumApi } from "@/lib/api";
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

export default function AlbumsPage() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);
  const [newAlbum, setNewAlbum] = useState({ name: "", description: "" });

  const { data, isLoading } = useQuery({
    queryKey: ["albums", page],
    queryFn: () => albumApi.list({ page, size: 20 }).then((r) => r.data.data),
  });

  const createMutation = useMutation({
    mutationFn: () => albumApi.create(newAlbum),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["albums"] });
      setCreateOpen(false);
      setNewAlbum({ name: "", description: "" });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => albumApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["albums"] }),
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Albums</h1>
          <p className="text-sm text-muted-foreground">
            Organize your files into albums
          </p>
        </div>
        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="size-4" />
              New Album
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Album</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Name</Label>
                <Input
                  value={newAlbum.name}
                  onChange={(e) =>
                    setNewAlbum((a) => ({ ...a, name: e.target.value }))
                  }
                  placeholder="Album name"
                />
              </div>
              <div className="space-y-2">
                <Label>Description</Label>
                <Input
                  value={newAlbum.description}
                  onChange={(e) =>
                    setNewAlbum((a) => ({ ...a, description: e.target.value }))
                  }
                  placeholder="Optional description"
                />
              </div>
            </div>
            <DialogFooter>
              <Button
                onClick={() => createMutation.mutate()}
                disabled={!newAlbum.name || createMutation.isPending}
              >
                {createMutation.isPending ? "Creating..." : "Create"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
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
            {data?.items?.map(
              (album: {
                id: string;
                name: string;
                description: string;
                is_public: boolean;
                file_count: number;
                created_at: string;
              }) => (
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
                      <Button size="icon-xs" variant="ghost">
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
                      {album.file_count} files
                    </span>
                    {album.is_public && (
                      <Badge variant="outline" className="text-[10px]">
                        Public
                      </Badge>
                    )}
                  </div>
                </div>
              )
            )}
          </div>

          {data?.items?.length === 0 && (
            <div className="flex flex-col items-center py-16 text-center">
              <FolderOpen className="mb-3 size-12 text-muted-foreground/30" />
              <p className="text-sm text-muted-foreground">No albums yet</p>
            </div>
          )}

          {data && data.total > 20 && (
            <div className="flex items-center justify-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </Button>
              <span className="text-sm text-muted-foreground">
                Page {page} of {Math.ceil(data.total / 20)}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= Math.ceil(data.total / 20)}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
