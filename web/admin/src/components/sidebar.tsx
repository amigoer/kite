import { useState } from "react";
import { Link, NavLink } from "react-router-dom";
import {
  LayoutDashboard,
  Upload,
  FolderOpen,
  Key,
  Settings,
  Users,
  User,
  HardDrive,
  LogOut,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/use-auth";
import { useI18n } from "@/i18n";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const navItems = [
  { to: "/", icon: LayoutDashboard, labelKey: "nav.dashboard" },
  { to: "/files", icon: Upload, labelKey: "nav.files" },
  { to: "/albums", icon: FolderOpen, labelKey: "nav.albums" },
  { to: "/tokens", icon: Key, labelKey: "nav.tokens" },
];

const adminItems = [
  { to: "/admin/storage", icon: HardDrive, labelKey: "nav.storage" },
  { to: "/admin/users", icon: Users, labelKey: "nav.users" },
  { to: "/admin/settings", icon: Settings, labelKey: "nav.settings" },
];

export function Sidebar() {
  const { user, logout } = useAuth();
  const { t } = useI18n();
  const isAdmin = user?.role === "admin";
  const [isCollapsed, setIsCollapsed] = useState(false);

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    cn(
      "group relative flex items-center rounded-md font-medium transition-all duration-200 outline-none",
      isCollapsed
        ? "justify-center px-0 py-2.5 h-[42px]"
        : "gap-3 px-3 py-2.5",
      isActive
        ? "bg-accent text-accent-foreground"
        : "text-muted-foreground hover:bg-accent hover:text-foreground"
    );

  const iconClass = (isActive: boolean) =>
    cn(
      "size-[18px] shrink-0 transition-colors",
      isActive
        ? "text-foreground"
        : "text-muted-foreground/70 group-hover:text-foreground"
    );

  return (
    <aside
      className={cn(
        "relative z-20 flex h-full flex-col border-r border-sidebar-border bg-sidebar transition-all duration-300 ease-in-out",
        isCollapsed ? "w-[72px]" : "w-[260px]"
      )}
    >
      {/* Logo */}
      <div
        className={cn(
          "flex h-[72px] items-center border-b border-transparent",
          isCollapsed ? "justify-center px-0" : "px-6"
        )}
      >
        <Link
          to="/"
          className="flex items-center gap-3 outline-none transition-opacity hover:opacity-80 overflow-hidden"
        >
          <div className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground shadow-sm">
            <span className="text-[13px] font-bold">K</span>
          </div>
          {!isCollapsed && (
            <div className="flex flex-col whitespace-nowrap">
              <span className="text-[15px] font-bold tracking-tight text-sidebar-foreground">
                Kite
              </span>
              <span className="text-[11px] font-medium text-muted-foreground">
                {t("nav.brandSub")}
              </span>
            </div>
          )}
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 overflow-y-auto overflow-x-hidden p-4 pt-4">
        <div
          className={cn(
            "mb-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/60 transition-all duration-300",
            isCollapsed ? "px-0 text-center" : "px-2"
          )}
        >
          {isCollapsed ? "·" : t("nav.general")}
        </div>
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === "/"}
            title={isCollapsed ? t(item.labelKey) : undefined}
            className={navLinkClass}
          >
            {({ isActive }) => (
              <>
                <item.icon className={iconClass(isActive)} />
                {!isCollapsed && (
                  <span className="truncate text-[14px]">
                    {t(item.labelKey)}
                  </span>
                )}
              </>
            )}
          </NavLink>
        ))}

        {isAdmin && (
          <div className="mt-8">
            <div
              className={cn(
                "mb-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/60 transition-all duration-300",
                isCollapsed ? "px-0 text-center" : "px-2"
              )}
            >
              {isCollapsed ? "·" : t("nav.admin")}
            </div>
            {adminItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                title={isCollapsed ? t(item.labelKey) : undefined}
                className={navLinkClass}
              >
                {({ isActive }) => (
                  <>
                    <item.icon className={iconClass(isActive)} />
                    {!isCollapsed && (
                      <span className="truncate text-[14px]">
                        {t(item.labelKey)}
                      </span>
                    )}
                  </>
                )}
              </NavLink>
            ))}
          </div>
        )}
      </nav>

      {/* User panel */}
      <div className="mt-auto border-t border-sidebar-border bg-sidebar p-3 sm:p-4">
        <DropdownMenu>
          <DropdownMenuTrigger
            className={cn(
              "flex w-full items-center rounded-md p-2 text-left outline-none transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring",
              isCollapsed ? "justify-center" : "gap-3"
            )}
          >
            <Avatar className="size-9 shrink-0 border border-border">
              <AvatarFallback className="bg-primary/10 text-xs font-semibold text-primary">
                {user?.username?.charAt(0).toUpperCase() ?? "U"}
              </AvatarFallback>
            </Avatar>
            {!isCollapsed && (
              <>
                <div className="flex flex-1 flex-col overflow-hidden">
                  <span className="truncate text-[14px] font-semibold text-sidebar-foreground">
                    {user?.username}
                  </span>
                  <span className="truncate text-[12px] text-muted-foreground">
                    {user?.role === "admin"
                      ? t("nav.roleAdmin")
                      : t("nav.roleUser")}
                  </span>
                </div>
                <LogOut
                  className="size-[16px] shrink-0 text-muted-foreground/50 transition-colors hover:text-destructive"
                  onClick={(e) => {
                    e.stopPropagation();
                    logout();
                  }}
                />
              </>
            )}
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align="end"
            className="w-[220px]"
            sideOffset={12}
          >
            <DropdownMenuItem className="cursor-pointer">
              <User className="mr-2 size-4" />
              {t("auth.profile")}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="cursor-pointer text-destructive focus:bg-destructive/10 focus:text-destructive"
              onClick={logout}
            >
              <LogOut className="mr-2 size-4" />
              {t("auth.logout")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Collapse toggle */}
      <button
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="absolute -right-3 top-1/2 z-50 flex size-6 -translate-y-1/2 items-center justify-center rounded-full border border-border bg-background text-muted-foreground shadow-sm transition-colors hover:text-foreground focus:outline-none"
      >
        {isCollapsed ? (
          <ChevronRight className="size-3.5" />
        ) : (
          <ChevronLeft className="size-3.5" />
        )}
      </button>
    </aside>
  );
}
