import { NavLink, Link, useLocation, useNavigate } from "react-router-dom";
import {
  LayoutDashboard,
  Upload,
  FolderOpen,
  KeyRound,
  FileText,
  HardDrive,
  Users,
  Settings,
  LogOut,
  ShieldCheck,
  ArrowLeft,
  User as UserIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/use-auth";
import { useI18n } from "@/i18n";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { KiteLogo } from "@/components/kite-logo";

const userNavItems = [
  { to: "/user/dashboard", icon: LayoutDashboard, labelKey: "nav.dashboard" },
  { to: "/user/files", icon: Upload, labelKey: "nav.files" },
  { to: "/user/folders", icon: FolderOpen, labelKey: "nav.albums" },
  { to: "/user/tokens", icon: KeyRound, labelKey: "nav.tokens" },
  { to: "/user/profile", icon: UserIcon, labelKey: "profile.title" },
];

const adminNavItems = [
  { to: "/admin/dashboard", icon: LayoutDashboard, labelKey: "nav.dashboard" },
  { to: "/admin/files", icon: FileText, labelKey: "nav.adminFiles" },
  { to: "/admin/storage", icon: HardDrive, labelKey: "nav.storage" },
  { to: "/admin/users", icon: Users, labelKey: "nav.users" },
  { to: "/admin/settings", icon: Settings, labelKey: "nav.settings" },
];

interface SidebarProps {
  onClose?: () => void;
  collapsed?: boolean;
}

export function Sidebar({ onClose, collapsed = false }: SidebarProps) {
  const { user, logout } = useAuth();
  const { t } = useI18n();
  const location = useLocation();
  const navigate = useNavigate();
  const displayName = user?.nickname?.trim() || user?.username;

  const isAdminWorkspace = location.pathname.startsWith("/admin");
  const navItems = isAdminWorkspace ? adminNavItems : userNavItems;
  const homePath = isAdminWorkspace ? "/admin/dashboard" : "/user/dashboard";

  const handleWorkspaceSwitch = () => {
    navigate(isAdminWorkspace ? "/user/dashboard" : "/admin/dashboard");
    onClose?.();
  };

  return (
    <aside className={cn("flex h-full flex-col bg-background transition-[width] duration-200", collapsed ? "w-16" : "w-55")}>
      {/* Logo — h-14 + border-b aligns with desktop header */}
      <div className={cn("flex h-14 shrink-0 items-center border-b", collapsed ? "justify-center px-0" : "justify-start px-4")}>
        <Link
          to={homePath}
          onClick={onClose}
          className={cn("flex items-center transition-opacity hover:opacity-80", collapsed ? "justify-center" : "gap-2.5")}
        >
          <KiteLogo className="size-6" />
          {!collapsed && (
            <div className="flex min-w-0 items-baseline gap-1.5">
              <span className="font-semibold tracking-tight">Kite</span>
              {isAdminWorkspace && (
                <span className="inline-flex items-center gap-1 rounded bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary">
                  <ShieldCheck className="size-3" />
                  {t("nav.admin")}
                </span>
              )}
            </div>
          )}
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-3">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            onClick={onClose}
            className={({ isActive }) =>
              cn(
                "flex items-center rounded-lg py-2 text-sm font-medium transition-colors",
                collapsed ? "justify-center px-2" : "gap-3 px-3",
                isActive
                  ? "bg-accent text-accent-foreground"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
              )
            }
          >
            <item.icon className="h-4 w-4" />
            {!collapsed && t(item.labelKey)}
          </NavLink>
        ))}
      </nav>

      {/* Workspace switch — admin only. h-14 matches the footer */}
      {user?.role === "admin" && (
        <div className={cn("flex h-14 shrink-0 items-center border-t", collapsed ? "px-2" : "px-3")}>
          <Button
            variant={isAdminWorkspace ? "secondary" : "outline"}
            size="sm"
            className={cn("w-full gap-2", collapsed && "px-0")}
            onClick={handleWorkspaceSwitch}
            aria-label={isAdminWorkspace ? t("nav.backToUser") : t("nav.adminPanel")}
          >
            {isAdminWorkspace ? (
              <ArrowLeft className="size-3.5" />
            ) : (
              <ShieldCheck className="size-3.5" />
            )}
            {!collapsed && (isAdminWorkspace ? t("nav.backToUser") : t("nav.adminPanel"))}
          </Button>
        </div>
      )}

      {/* Bottom user section — h-14 matches the footer. Mobile only (desktop has the avatar dropdown in header) */}
      <div className="flex h-14 shrink-0 items-center gap-1 border-t px-2 md:hidden">
        <Link
          to="/user/profile"
          onClick={onClose}
          className="flex min-w-0 flex-1 items-center gap-2 rounded-md px-2 py-1.5 transition-colors hover:bg-accent"
          aria-label={t("profile.title")}
        >
          <Avatar className="size-7">
            {user?.avatar_url ? (
              <AvatarImage src={user.avatar_url} alt={user.username ?? ""} />
            ) : null}
            <AvatarFallback className="text-[10px] font-medium">
              {user?.username?.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium leading-none">
              {displayName}
            </p>
            <p className="mt-0.5 truncate text-[11px] text-muted-foreground">
              @{user?.username}
            </p>
          </div>
        </Link>
        <Button
          variant="ghost"
          size="icon-sm"
          className="shrink-0 rounded-full text-muted-foreground hover:text-destructive"
          onClick={logout}
          aria-label={t("auth.logout")}
        >
          <LogOut className="h-4 w-4" />
        </Button>
      </div>
    </aside>
  );
}
