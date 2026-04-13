import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { UserSidebar } from "./user-sidebar";

export function UserLayout() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-sm text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <UserSidebar />
      <main className="flex-1 overflow-y-auto bg-muted/20 p-6 lg:p-8">
        <Outlet />
      </main>
    </div>
  );
}
