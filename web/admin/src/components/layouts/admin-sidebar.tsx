import { Link, NavLink } from "react-router-dom";
import {
  LayoutDashboard,
  HardDrive,
  Users,
  Settings,
  LogOut,
  ArrowLeft,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/use-auth";
import { useI18n } from "@/i18n";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

const adminNavItems = [
  { to: "/admin", icon: LayoutDashboard, labelKey: "nav.dashboard", end: true },
  { to: "/admin/storage", icon: HardDrive, labelKey: "nav.storage" },
  { to: "/admin/users", icon: Users, labelKey: "nav.users" },
  { to: "/admin/settings", icon: Settings, labelKey: "nav.settings" },
];

export function AdminSidebar() {
  const { user, logout } = useAuth();
  const { t } = useI18n();

  return (
    <aside className="flex h-full w-[240px] shrink-0 flex-col border-r border-sidebar-border bg-sidebar">
      {/* Brand */}
      <div className="flex h-16 items-center justify-between border-b border-sidebar-border px-5">
        <Link to="/admin" className="flex items-center gap-2.5 transition-opacity hover:opacity-80">
          <div className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
            <span className="text-sm font-bold">K</span>
          </div>
          <div className="flex flex-col">
            <span className="text-base font-bold tracking-tight">Kite</span>
            <span className="text-[10px] text-muted-foreground">{t("nav.adminPanel")}</span>
          </div>
        </Link>
      </div>

      {/* Back to user center */}
      <div className="border-b border-sidebar-border px-3 py-2">
        <Link
          to="/dashboard"
          className="flex items-center gap-2 rounded-lg px-3 py-2 text-[13px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        >
          <ArrowLeft size={14} />
          {t("nav.backToUser")}
        </Link>
      </div>

      {/* Nav */}
      <nav className="flex-1 space-y-1 overflow-y-auto p-3">
        <div className="mb-2 px-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/60">
          {t("nav.admin")}
        </div>
        {adminNavItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-[13px] font-medium transition-colors",
                isActive
                  ? "bg-primary/10 text-primary"
                  : "text-muted-foreground hover:bg-accent hover:text-foreground"
              )
            }
          >
            <item.icon size={18} />
            {t(item.labelKey)}
          </NavLink>
        ))}
      </nav>

      {/* User */}
      <div className="border-t border-sidebar-border p-3">
        <div className="flex items-center gap-3 rounded-lg px-3 py-2">
          <Avatar className="size-8 border">
            <AvatarFallback className="bg-primary/10 text-xs font-semibold text-primary">
              {user?.username?.charAt(0).toUpperCase() ?? "U"}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1 overflow-hidden">
            <p className="truncate text-sm font-medium">{user?.username}</p>
            <p className="truncate text-[11px] text-muted-foreground">{t("nav.roleAdmin")}</p>
          </div>
          <button
            onClick={logout}
            className="rounded p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
          >
            <LogOut size={14} />
          </button>
        </div>
      </div>
    </aside>
  );
}
