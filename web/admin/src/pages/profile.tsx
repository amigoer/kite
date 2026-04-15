import { useState, useEffect } from "react";
import { useMutation } from "@tanstack/react-query";
import { User as UserIcon, Mail, KeyRound, Loader2, ShieldCheck } from "lucide-react";

import { authApi } from "@/lib/api";
import { useAuth } from "@/hooks/use-auth";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";

interface User {
  user_id: string;
  username: string;
  email?: string;
  role: string;
}

export default function ProfilePage() {
  const { user, applyTokensAndRefresh } = useAuth();
  const { t } = useI18n();

  const [profileForm, setProfileForm] = useState({
    username: user?.username ?? "",
    email: user?.email ?? "",
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
        email: user.email ?? "",
      });
    }
  }, [user]);

  const profileMutation = useMutation({
    mutationFn: () =>
      authApi.updateProfile({
        username: profileForm.username,
        email: profileForm.email,
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
        email: updated.email ?? "",
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

  if (!user) return null;

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
      <Card>
        <CardContent className="flex items-center gap-4 py-5">
          <Avatar className="size-14">
            <AvatarFallback className="text-base font-medium">
              {user.username?.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="truncate text-base font-semibold">
                {user.username}
              </p>
              <Badge variant={user.role === "admin" ? "default" : "secondary"}>
                {user.role === "admin"
                  ? t("nav.roleAdmin")
                  : t("nav.roleUser")}
              </Badge>
            </div>
            <p className="mt-0.5 truncate text-sm text-muted-foreground">
              {user.email ?? "—"}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Basic info */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <UserIcon className="size-4" />
            {t("profile.basicInfo")}
          </CardTitle>
          <CardDescription>{t("profile.basicInfoDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={handleProfileSubmit}
            className="grid max-w-lg gap-4"
          >
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
        </CardContent>
      </Card>

      {/* Password */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <KeyRound className="size-4" />
            {t("profile.changePassword")}
          </CardTitle>
          <CardDescription>{t("profile.changePasswordDesc")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={handlePasswordSubmit}
            className="grid max-w-lg gap-4"
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
        </CardContent>
      </Card>
    </div>
  );
}
