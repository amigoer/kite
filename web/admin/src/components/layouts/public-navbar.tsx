import { Link } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";

export function PublicNavbar() {
  const { user } = useAuth();
  const { t } = useI18n();

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/80 backdrop-blur-lg">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link to="/" className="flex items-center gap-2.5">
          <div className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
            <span className="text-sm font-bold">K</span>
          </div>
          <span className="text-lg font-bold tracking-tight">Kite</span>
        </Link>

        <nav className="hidden items-center gap-8 md:flex">
          <a href="#features" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
            {t("home.features")}
          </a>
          <a href="#usage" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
            {t("home.usage")}
          </a>
        </nav>

        <div className="flex items-center gap-3">
          {user ? (
            <Button asChild size="sm">
              <Link to="/dashboard">{t("home.enterApp")}</Link>
            </Button>
          ) : (
            <>
              <Button variant="ghost" size="sm" asChild>
                <Link to="/login">{t("auth.login")}</Link>
              </Button>
              <Button size="sm" asChild>
                <Link to="/register">{t("auth.register")}</Link>
              </Button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
