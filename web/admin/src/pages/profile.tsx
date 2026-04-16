import { useState, useEffect } from "react";
import { useMutation } from "@tanstack/react-query";
import { User as UserIcon, Mail, KeyRound, Loader2, ShieldCheck, Camera } from "lucide-react";

import { authApi, fileApi } from "@/lib/api";
import { useAuth } from "@/hooks/use-auth";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";

interface User {
  user_id: string;
  username: string;
  nickname?: string;
  email?: string;
  avatar_url?: string;
  role: string;
}

export default function ProfilePage() {
  const { user, applyTokensAndRefresh } = useAuth();
  const { t } = useI18n();

  const [profileForm, setProfileForm] = useState({
    username: user?.username ?? "",
    nickname: user?.nickname ?? "",
    email: user?.email ?? "",
    avatarUrl: user?.avatar_url ?? "",
  });
  const [passwordForm, setPasswordForm] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });

  useEffect(() => {
    if (user) {
      setProfileForm({
        username: user.username ?? "",
        nickname: user.nickname ?? "",
        email: user.email ?? "",
        avatarUrl: user.avatar_url ?? "",
      });
    }
  }, [user]);

  const avatarUploadMutation = useMutation({
    mutationFn: async (file: File) => {
      const res = await fileApi.upload(file);
      const uploaded = res.data?.data?.links?.url as string | undefined;
      if (!uploaded) {
        throw new Error("avatar upload did not return URL");
      }

      await authApi.updateProfile({
        username: profileForm.username,
        nickname: profileForm.nickname,
        email: profileForm.email,
        avatar_url: uploaded,
      });

      const tokens = {
        access_token: localStorage.getItem("access_token") ?? "",
        refresh_token: localStorage.getItem("refresh_token") ?? "",
      };
      if (tokens.access_token) {
        await applyTokensAndRefresh(tokens);
      }

      return uploaded;
    },
    onSuccess: (url) => {
      setProfileForm((prev) => ({ ...prev, avatarUrl: url }));
      toast.success(t("profile.avatarUploaded"));
    },
    onError: () => {
      toast.error(t("profile.avatarUploadFailed"));
    },
  });

  const profileMutation = useMutation({
    mutationFn: () =>
      authApi.updateProfile({
        username: profileForm.username,
        nickname: profileForm.nickname,
        email: profileForm.email,
        avatar_url: profileForm.avatarUrl || undefined,
      }),
    onSuccess: async (res) => {
      const updated: User = res.data.data;
      const tokens = {
        access_token: localStorage.getItem("access_token") ?? "",
        refresh_token: localStorage.getItem("refresh_token") ?? "",
      };
      if (tokens.access_token) {
        await applyTokensAndRefresh(tokens);
      }
      toast.success(t("profile.profileSaved"));
      setProfileForm({
        username: updated.username,
        nickname: updated.nickname ?? "",
        email: updated.email ?? "",
        avatarUrl: updated.avatar_url ?? "",
      });
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data
          ?.message ?? t("profile.profileFailed");
      toast.error(msg);
    },
  });

  const passwordMutation = useMutation({
    mutationFn: () =>
      authApi.changePassword({
        current_password: passwordForm.currentPassword,
        new_password: passwordForm.newPassword,
      }),
    onSuccess: () => {
      toast.success(t("profile.passwordSaved"));
      setPasswordForm({
        currentPassword: "",
        newPassword: "",
        confirmPassword: "",
      });
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data
          ?.message ?? t("profile.passwordFailed");
      toast.error(msg);
    },
  });

  const handleProfileSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!profileForm.username.trim() || !profileForm.email.trim()) {
      toast.error(t("profile.allFieldsRequired"));
      return;
    }
    profileMutation.mutate();
  };

  const handlePasswordSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (
      !passwordForm.currentPassword ||
      !passwordForm.newPassword ||
      !passwordForm.confirmPassword
    ) {
      toast.error(t("profile.allFieldsRequired"));
      return;
    }
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      toast.error(t("auth.passwordMismatch"));
      return;
    }
    if (passwordForm.newPassword === passwordForm.currentPassword) {
      toast.error(t("profile.newPasswordMustDiffer"));
      return;
    }
    passwordMutation.mutate();
  };

  const handleAvatarFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (!file.type.startsWith("image/")) {
      toast.error(t("profile.avatarImageOnly"));
      e.target.value = "";
      return;
    }
    avatarUploadMutation.mutate(file);
    e.target.value = "";
  };

  if (!user) return null;

  const displayName = profileForm.nickname.trim() || user.nickname?.trim() || user.username;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">
          {t("profile.title")}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("profile.description")}
        </p>
      </div>

      {/* Overview */}
      <section className="relative overflow-hidden rounded-4xl bg-[linear-gradient(135deg,rgba(242,246,237,0.95)_0%,rgba(255,255,255,1)_45%,rgba(248,241,231,0.92)_100%)] px-6 py-7 shadow-[0_20px_60px_rgba(15,23,42,0.08)] ring-1 ring-black/5 md:px-8 md:py-8">
        <div className="pointer-events-none absolute -right-12 -top-12 size-40 rounded-full bg-[radial-gradient(circle,rgba(120,170,90,0.18),rgba(120,170,90,0))]" />
        <div className="pointer-events-none absolute -bottom-16 left-1/3 size-48 rounded-full bg-[radial-gradient(circle,rgba(214,161,76,0.14),rgba(214,161,76,0))]" />
        <div className="relative flex flex-col gap-6 md:flex-row md:items-center">
          <div className="group relative shrink-0 self-start">
            <Avatar className="size-26 ring-4 ring-white/80 shadow-xl md:size-30">
              <AvatarImage src={profileForm.avatarUrl || user.avatar_url} alt={displayName} className="object-cover" />
              <AvatarFallback className="bg-stone-200 text-2xl font-semibold text-stone-700 md:text-3xl">
                {displayName.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <Label
              htmlFor="avatar-file"
              className="absolute inset-0 flex cursor-pointer flex-col items-center justify-center rounded-full bg-black/56 text-white opacity-0 transition-opacity group-hover:opacity-100"
            >
              {avatarUploadMutation.isPending ? (
                <Loader2 className="size-5 animate-spin" />
              ) : (
                <>
                  <Camera className="size-4" />
                  <span className="mt-1 text-[11px] font-medium">{t("profile.uploadAvatar")}</span>
                </>
              )}
            </Label>
            <Input
              id="avatar-file"
              type="file"
              accept="image/*"
              className="hidden"
              onChange={handleAvatarFileChange}
              disabled={avatarUploadMutation.isPending}
            />
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-3">
              <h2 className="truncate text-3xl font-semibold tracking-tight text-foreground md:text-4xl">
                {displayName}
              </h2>
              <Badge className="rounded-full px-3 py-1 text-[11px] font-medium" variant={user.role === "admin" ? "default" : "secondary"}>
                {user.role === "admin" ? t("nav.roleAdmin") : t("nav.roleUser")}
              </Badge>
            </div>
            <div className="mt-3 flex flex-col gap-2 text-sm text-muted-foreground md:flex-row md:flex-wrap md:items-center md:gap-4">
              <span className="font-medium text-foreground/75">@{user.username}</span>
              <span className="hidden md:inline text-border">/</span>
              <span className="truncate">{user.email ?? "—"}</span>
            </div>
            <p className="mt-4 max-w-2xl text-sm leading-6 text-muted-foreground">
              {avatarUploadMutation.isPending
                ? t("profile.uploadingAvatar")
                : t("profile.heroDescription")}
            </p>
          </div>
        </div>
      </section>

      <section className="rounded-4xl bg-[linear-gradient(180deg,rgba(255,255,255,0.82),rgba(249,248,244,0.86))] px-5 py-6 md:px-6">
        <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_420px] lg:gap-10">
          <div>
            <div>
              <div className="flex items-center gap-2 text-base font-semibold text-foreground">
                <UserIcon className="size-4" />
                {t("profile.basicInfo")}
              </div>
              <p className="mt-1 text-sm text-muted-foreground">{t("profile.basicInfoDesc")}</p>
            </div>

            <form
              onSubmit={handleProfileSubmit}
              className="mt-6 grid max-w-xl gap-4"
            >
              <div className="grid gap-2">
                <Label htmlFor="nickname">{t("profile.nickname")}</Label>
                <Input
                  id="nickname"
                  value={profileForm.nickname}
                  onChange={(e) =>
                    setProfileForm((p) => ({ ...p, nickname: e.target.value }))
                  }
                  maxLength={32}
                  placeholder={t("profile.nicknamePlaceholder")}
                />
                <p className="text-xs text-muted-foreground">{t("profile.nicknameHint")}</p>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="username">{t("auth.username")}</Label>
                <Input
                  id="username"
                  autoCapitalize="none"
                  autoCorrect="off"
                  value={profileForm.username}
                  onChange={(e) =>
                    setProfileForm((p) => ({ ...p, username: e.target.value }))
                  }
                  minLength={3}
                  maxLength={32}
                  required
                />
                <p className="text-xs text-muted-foreground">{t("profile.usernameHint")}</p>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="email">
                  <Mail className="me-1 inline size-3.5" />
                  {t("auth.email")}
                </Label>
                <Input
                  id="email"
                  type="email"
                  autoCapitalize="none"
                  autoCorrect="off"
                  value={profileForm.email}
                  onChange={(e) =>
                    setProfileForm((p) => ({ ...p, email: e.target.value }))
                  }
                  placeholder={t("auth.emailPlaceholder")}
                  required
                />
              </div>

              <div className="flex">
                <Button type="submit" disabled={profileMutation.isPending}>
                  {profileMutation.isPending && (
                    <Loader2 className="size-4 animate-spin" />
                  )}
                  {profileMutation.isPending
                    ? t("settings.saving")
                    : t("profile.saveProfile")}
                </Button>
              </div>
            </form>
          </div>

          <div className="relative">
            <div className="hidden lg:block absolute left-0 top-0 h-full w-px bg-linear-to-b from-transparent via-border to-transparent -translate-x-5" />
            <div>
              <div className="flex items-center gap-2 text-base font-semibold text-foreground">
                <KeyRound className="size-4" />
                {t("profile.changePassword")}
              </div>
              <p className="mt-1 text-sm text-muted-foreground">{t("profile.changePasswordDesc")}</p>
            </div>

            <form
              onSubmit={handlePasswordSubmit}
              className="mt-6 grid max-w-lg gap-4"
            >
              <div className="grid gap-2">
                <Label htmlFor="currentPassword">
                  {t("profile.currentPassword")}
                </Label>
                <Input
                  id="currentPassword"
                  type="password"
                  autoComplete="current-password"
                  value={passwordForm.currentPassword}
                  onChange={(e) =>
                    setPasswordForm((p) => ({
                      ...p,
                      currentPassword: e.target.value,
                    }))
                  }
                  required
                />
              </div>

              <Separator />

              <div className="grid gap-2">
                <Label htmlFor="newPassword">{t("profile.newPassword")}</Label>
                <Input
                  id="newPassword"
                  type="password"
                  autoComplete="new-password"
                  value={passwordForm.newPassword}
                  onChange={(e) =>
                    setPasswordForm((p) => ({
                      ...p,
                      newPassword: e.target.value,
                    }))
                  }
                  placeholder={t("auth.passwordHint")}
                  minLength={6}
                  maxLength={64}
                  required
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="confirmPassword">
                  {t("auth.confirmPassword")}
                </Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  autoComplete="new-password"
                  value={passwordForm.confirmPassword}
                  onChange={(e) =>
                    setPasswordForm((p) => ({
                      ...p,
                      confirmPassword: e.target.value,
                    }))
                  }
                  placeholder={t("auth.confirmPasswordPlaceholder")}
                  required
                />
              </div>

              <div className="flex items-center gap-3 rounded-md border border-muted bg-muted/40 p-3 text-xs text-muted-foreground">
                <ShieldCheck className="size-4 shrink-0" />
                <span>{t("profile.passwordTip")}</span>
              </div>

              <div className="flex">
                <Button type="submit" disabled={passwordMutation.isPending}>
                  {passwordMutation.isPending && (
                    <Loader2 className="size-4 animate-spin" />
                  )}
                  {passwordMutation.isPending
                    ? t("settings.saving")
                    : t("profile.savePassword")}
                </Button>
              </div>
            </form>
          </div>
        </div>
      </section>
    </div>
  );
}
