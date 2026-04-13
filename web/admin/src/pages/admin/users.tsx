import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Users, Trash2, Plus, Pencil } from "lucide-react";
import { userApi } from "@/lib/api";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface UserItem {
  id: string;
  username: string;
  email: string;
  role: string;
  storage_used: number;
  storage_limit: number;
  is_active: boolean;
}

function formatBytes(bytes: number) {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function UsersPage() {
  const { t } = useI18n();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<UserItem | null>(null);
  const [form, setForm] = useState({ username: "", email: "", password: "", role: "user", storage_limit: "10737418240", is_active: true });
  const [error, setError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["users", page],
    queryFn: () => userApi.list({ page, size: 20 }).then((r) => r.data.data),
  });

  const createMutation = useMutation({
    mutationFn: () => userApi.create({
      username: form.username,
      email: form.email,
      password: form.password,
      role: form.role,
      storage_limit: parseInt(form.storage_limit),
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      closeDialog();
    },
    onError: (err: unknown) => {
      setError((err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? "Failed");
    },
  });

  const updateMutation = useMutation({
    mutationFn: () => {
      const payload: Record<string, unknown> = {
        role: form.role,
        is_active: form.is_active,
        storage_limit: parseInt(form.storage_limit),
      };
      if (form.email) payload.email = form.email;
      if (form.password) payload.password = form.password;
      return userApi.update(editingUser!.id, payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      closeDialog();
    },
    onError: (err: unknown) => {
      setError((err as { response?: { data?: { message?: string } } })?.response?.data?.message ?? "Failed");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => userApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["users"] }),
  });

  const openCreate = () => {
    setEditingUser(null);
    setForm({ username: "", email: "", password: "", role: "user", storage_limit: "10737418240", is_active: true });
    setError("");
    setDialogOpen(true);
  };

  const openEdit = (user: UserItem) => {
    setEditingUser(user);
    setForm({
      username: user.username,
      email: user.email,
      password: "",
      role: user.role,
      storage_limit: String(user.storage_limit),
      is_active: user.is_active,
    });
    setError("");
    setDialogOpen(true);
  };

  const closeDialog = () => {
    setDialogOpen(false);
    setEditingUser(null);
    setError("");
  };

  const handleSave = () => {
    if (editingUser) {
      updateMutation.mutate();
    } else {
      createMutation.mutate();
    }
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">{t("users.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("users.description")}</p>
        </div>
        <Button onClick={openCreate}>
          <Plus className="size-4" />
          {t("users.addUser")}
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-16 rounded-lg" />
          ))}
        </div>
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden sm:block rounded-lg border">
            <div className="grid grid-cols-[1fr_1fr_auto_auto_auto] gap-4 border-b px-4 py-2 text-xs font-medium text-muted-foreground">
              <span>{t("users.username")}</span>
              <span>{t("users.email")}</span>
              <span>{t("users.role")}</span>
              <span>{t("users.storageCol")}</span>
              <span />
            </div>
            {data?.items?.map((user: UserItem) => (
              <div
                key={user.id}
                className="grid grid-cols-[1fr_1fr_auto_auto_auto] items-center gap-4 border-b px-4 py-3 text-sm last:border-0"
              >
                <div className="flex items-center gap-2">
                  <span className="font-medium">{user.username}</span>
                  {!user.is_active && (
                    <Badge variant="outline" className="text-[10px]">{t("common.disabled")}</Badge>
                  )}
                </div>
                <span className="text-muted-foreground truncate">{user.email}</span>
                <Badge
                  variant={user.role === "admin" ? "default" : "secondary"}
                  className="text-[10px]"
                >
                  {user.role}
                </Badge>
                <span className="text-xs text-muted-foreground whitespace-nowrap">
                  {formatBytes(user.storage_used)}
                  {user.storage_limit > 0 && ` / ${formatBytes(user.storage_limit)}`}
                </span>
                <div className="flex gap-1">
                  <Button size="icon-xs" variant="ghost" onClick={() => openEdit(user)}>
                    <Pencil className="size-3.5" />
                  </Button>
                  <Button
                    size="icon-xs"
                    variant="ghost"
                    onClick={() => deleteMutation.mutate(user.id)}
                  >
                    <Trash2 className="size-3.5 text-destructive" />
                  </Button>
                </div>
              </div>
            ))}
          </div>

          {/* Mobile cards */}
          <div className="sm:hidden space-y-3">
            {data?.items?.map((user: UserItem) => (
              <div key={user.id} className="rounded-lg border bg-card p-4">
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">{user.username}</span>
                      <Badge variant={user.role === "admin" ? "default" : "secondary"} className="text-[10px]">
                        {user.role}
                      </Badge>
                      {!user.is_active && <Badge variant="outline" className="text-[10px]">{t("common.disabled")}</Badge>}
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">{user.email}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {formatBytes(user.storage_used)}
                      {user.storage_limit > 0 && ` / ${formatBytes(user.storage_limit)}`}
                    </p>
                  </div>
                  <div className="flex gap-1">
                    <Button size="icon-xs" variant="ghost" onClick={() => openEdit(user)}>
                      <Pencil className="size-3.5" />
                    </Button>
                    <Button size="icon-xs" variant="ghost" onClick={() => deleteMutation.mutate(user.id)}>
                      <Trash2 className="size-3.5 text-destructive" />
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {data?.items?.length === 0 && (
            <div className="flex flex-col items-center py-16 text-center">
              <Users className="mb-3 size-12 text-muted-foreground/30" />
              <p className="text-sm text-muted-foreground">{t("users.noUsers")}</p>
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

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={(open) => !open && closeDialog()}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingUser ? t("users.editUser") : t("users.addUser")}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {error && (
              <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</div>
            )}
            <div className="space-y-2">
              <Label>{t("users.username")}</Label>
              <Input
                value={form.username}
                onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))}
                disabled={!!editingUser}
                placeholder={t("auth.chooseUsername")}
              />
            </div>
            <div className="space-y-2">
              <Label>{t("users.email")}</Label>
              <Input
                type="email"
                value={form.email}
                onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                placeholder={t("auth.emailPlaceholder")}
              />
            </div>
            <div className="space-y-2">
              <Label>
                {t("auth.password")}
                {editingUser && <span className="text-xs text-muted-foreground ml-1">({t("users.leaveBlank")})</span>}
              </Label>
              <Input
                type="password"
                value={form.password}
                onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
                placeholder={editingUser ? t("users.leaveBlank") : t("auth.passwordHint")}
              />
            </div>
            <div className="flex gap-4">
              <div className="flex-1 space-y-2">
                <Label>{t("users.role")}</Label>
                <div className="flex gap-2">
                  {["user", "admin"].map((r) => (
                    <Button
                      key={r}
                      type="button"
                      size="sm"
                      variant={form.role === r ? "default" : "outline"}
                      onClick={() => setForm((f) => ({ ...f, role: r }))}
                    >
                      {r}
                    </Button>
                  ))}
                </div>
              </div>
              {editingUser && (
                <div className="flex-1 space-y-2">
                  <Label>{t("users.status")}</Label>
                  <div className="flex gap-2">
                    <Button
                      type="button"
                      size="sm"
                      variant={form.is_active ? "default" : "outline"}
                      onClick={() => setForm((f) => ({ ...f, is_active: true }))}
                    >
                      {t("common.active")}
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant={!form.is_active ? "destructive" : "outline"}
                      onClick={() => setForm((f) => ({ ...f, is_active: false }))}
                    >
                      {t("common.disabled")}
                    </Button>
                  </div>
                </div>
              )}
            </div>
            <div className="space-y-2">
              <Label>{t("users.storageLimit")}</Label>
              <div className="flex gap-2">
                {[
                  { label: "1 GB", value: "1073741824" },
                  { label: "5 GB", value: "5368709120" },
                  { label: "10 GB", value: "10737418240" },
                  { label: "50 GB", value: "53687091200" },
                  { label: t("users.unlimited"), value: "-1" },
                ].map((opt) => (
                  <Button
                    key={opt.value}
                    type="button"
                    size="sm"
                    variant={form.storage_limit === opt.value ? "default" : "outline"}
                    onClick={() => setForm((f) => ({ ...f, storage_limit: opt.value }))}
                  >
                    {opt.label}
                  </Button>
                ))}
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={closeDialog}>{t("common.cancel")}</Button>
            <Button onClick={handleSave} disabled={isPending || (!editingUser && (!form.username || !form.email || !form.password))}>
              {isPending ? t("common.loading") : t("common.save")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
