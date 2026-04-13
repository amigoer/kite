import { Outlet } from "react-router-dom";
import { PublicNavbar } from "./public-navbar";

export function PublicLayout() {
  return (
    <div className="flex min-h-screen flex-col">
      <PublicNavbar />
      <main className="flex-1">
        <Outlet />
      </main>
      <footer className="border-t border-border/40 py-8 text-center text-xs text-muted-foreground">
        &copy; {new Date().getFullYear()} Kite. All rights reserved.
      </footer>
    </div>
  );
}
