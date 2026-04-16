import { useState } from "react";
import { Navigate, Link, useLocation, useNavigate } from "react-router-dom";
import { PageTransition } from "@/components/page-transition";
import { Bell, Menu, User as UserIcon, LogOut } from "lucide-react";
import { useAuth } from "@/hooks/use-auth";
import { useI18n } from "@/i18n";
import { Sidebar } from "@/components/layouts/sidebar";
import { Button } from "@/components/ui/button";
import { KiteLogo } from "@/components/kite-logo";
import { ThemeToggle } from "@/components/theme-toggle";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import {
  Sheet,
  SheetContent,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";

const SIDEBAR_COLLAPSED_KEY = "kite_sidebar_collapsed";

export default function AppLayout() {
  const { user, loading, logout } = useAuth();
  const { t } = useI18n();
  const location = useLocation();
  const navigate = useNavigate();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
    if (typeof window === "undefined") return false;
    return localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === "true";
  });
  const displayName = user?.nickname?.trim() || user?.username;

  const toggleSidebar = () => {
    setSidebarCollapsed((prev) => {
      const next = !prev;
      localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(next));
      return next;
    });
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <KiteLogo className="size-8 animate-[splash-pulse_1.4s_ease-in-out_infinite]" />
      </div>
    );
  }

  if (!user) {
    return (
      <Navigate
        to="/login"
        replace
        state={{ from: location.pathname + location.search }}
      />
    );
  }

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Desktop sidebar */}
      <div className="hidden shrink-0 border-r md:flex">
        <Sidebar collapsed={sidebarCollapsed} onToggleCollapse={toggleSidebar} />
      </div>

      {/* Right content area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Mobile header */}
        <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
          <header className="flex h-14 shrink-0 items-center justify-between border-b px-4 md:hidden">
            <div className="flex items-center gap-2">
              <SheetTrigger asChild>
                <Button variant="ghost" size="icon-sm">
                  <Menu className="h-5 w-5" />
                </Button>
              </SheetTrigger>
              <Link
                to="/dashboard"
                className="flex items-center gap-2 transition-opacity hover:opacity-80"
              >
                <KiteLogo className="size-6" />
                <span className="font-semibold tracking-tight">Kite</span>
              </Link>
            </div>

            <ThemeToggle />
          </header>
          <SheetContent side="left" className="w-55 p-0">
            <SheetTitle className="sr-only">Navigation</SheetTitle>
            <Sidebar onClose={() => setMobileOpen(false)} />
          </SheetContent>
        </Sheet>

        {/* Desktop header */}
        <header className="hidden h-14 shrink-0 border-b md:flex">
          <div className="mx-auto flex w-full max-w-screen-2xl items-center justify-end px-4 sm:px-6 lg:px-8">
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="icon-sm"
                className="text-muted-foreground hover:text-foreground"
                aria-label="Notifications"
              >
                <Bell className="size-4" />
              </Button>
              <ThemeToggle />
              <div className="mx-1 h-5 w-px bg-border" />
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="gap-2 px-2 text-xs text-muted-foreground hover:bg-accent hover:text-foreground"
                  >
                    <Avatar className="size-5">
                      <AvatarImage src={user.avatar_url} alt={user.username ?? ""} />
                      <AvatarFallback className="text-[10px] font-medium">
                        {user.username?.charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <span className="font-medium">{displayName}</span>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-44">
                  <DropdownMenuLabel className="font-normal">
                    <div className="flex flex-col">
                      <span className="text-sm font-medium leading-none">
                        {displayName}
                      </span>
                      <span className="mt-1 truncate text-[11px] text-muted-foreground">
                        @{user.username}
                      </span>
                      <span className="mt-1 truncate text-[11px] text-muted-foreground">
                        {user.email ?? ""}
                      </span>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => navigate("/profile")}>
                    <UserIcon className="size-4" />
                    {t("profile.title")}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem variant="destructive" onClick={logout}>
                    <LogOut className="size-4" />
                    {t("auth.logout")}
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </header>

        {/* Scrollable content */}
        <main className="flex-1 overflow-y-auto">
          <div className="mx-auto w-full max-w-screen-2xl px-4 py-6 sm:px-6 lg:px-8">
            <PageTransition />
          </div>
        </main>

        {/* Footer — sticky at bottom, h-14 + border-t matches sidebar bottom */}
        <footer className="flex h-14 shrink-0 border-t bg-background">
          <div className="mx-auto flex w-full max-w-screen-2xl items-center justify-between px-4 text-xs text-muted-foreground sm:px-6 lg:px-8">
            <div className="flex items-center gap-1.5">
              <KiteLogo className="size-4" />
              <span className="font-medium text-foreground/70">Kite</span>
              <span className="mx-0.5 text-border">·</span>
              <span className="hidden sm:inline">{t("footer.description")}</span>
            </div>
            <div className="flex items-center gap-3">
              <a
                href="https://github.com/amigoer/kite"
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1 transition-colors hover:text-foreground"
              >
                <svg className="size-3.5" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z" />
                </svg>
                <span className="hidden sm:inline">GitHub</span>
              </a>
              <span className="text-border">·</span>
              <span>&copy; 2026 Kite</span>
            </div>
          </div>
        </footer>
      </div>
    </div>
  );
}
