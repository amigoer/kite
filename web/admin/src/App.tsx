import { BrowserRouter, Routes, Route, Navigate, Outlet } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthContext, useAuthProvider, useAuth } from "@/hooks/use-auth";
import { I18nContext, useI18nProvider } from "@/i18n";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";

// Layouts
import AppLayout from "@/components/layouts/app-layout";
import { AuthLayout } from "@/components/layout";

// Auth pages
import LoginPage from "@/pages/login";
import RegisterPage from "@/pages/register";
import FirstLoginPage from "@/pages/first-login";

// Pages
import DashboardPage from "@/pages/dashboard";
import FilesPage from "@/pages/files";
import AlbumsPage from "@/pages/albums";
import TokensPage from "@/pages/tokens";
import ProfilePage from "@/pages/profile";

// Admin pages
import AdminFilesPage from "@/pages/admin/files";
import StoragePage from "@/pages/admin/storage";
import UsersPage from "@/pages/admin/users";
import SettingsPage from "@/pages/admin/settings";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, refetchOnWindowFocus: false },
  },
});

function AdminRoute() {
  const { user } = useAuth();
  if (user?.role !== "admin") return <Navigate to="/dashboard" replace />;
  return <Outlet />;
}

// FirstLoginGate: 当用户 password_must_change=true 时，除 /first-login 外其他路径
// 一律跳转到首次配置页；反之访问 /first-login 则跳回 dashboard。
function FirstLoginGate({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  if (user?.password_must_change) {
    return <Navigate to="/first-login" replace />;
  }
  return <>{children}</>;
}

function FirstLoginOnlyGate() {
  const { user, loading } = useAuth();
  if (loading) return null;
  if (!user) return <Navigate to="/login" replace />;
  if (!user.password_must_change) return <Navigate to="/dashboard" replace />;
  return (
    <div className="flex min-h-svh items-center justify-center bg-background px-4 py-8">
      <div className="w-full max-w-md">
        <FirstLoginPage />
      </div>
    </div>
  );
}

function AppRoutes() {
  const auth = useAuthProvider();

  return (
    <AuthContext.Provider value={auth}>
      <Routes>
        {/* First-login reset (强制首次配置，独立布局) */}
        <Route path="/first-login" element={<FirstLoginOnlyGate />} />

        {/* Auth (centered card) */}
        <Route element={<AuthLayout />}>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
        </Route>

        {/* App (sidebar layout) — 受 first-login 守卫保护 */}
        <Route
          element={
            <FirstLoginGate>
              <AppLayout />
            </FirstLoginGate>
          }
        >
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/files" element={<FilesPage />} />
          <Route path="/albums" element={<AlbumsPage />} />
          <Route path="/tokens" element={<TokensPage />} />
          <Route path="/profile" element={<ProfilePage />} />

          {/* Admin pages (role guard) */}
          <Route element={<AdminRoute />}>
            <Route path="/admin" element={<Navigate to="/dashboard" replace />} />
            <Route path="/admin/files" element={<AdminFilesPage />} />
            <Route path="/admin/storage" element={<StoragePage />} />
            <Route path="/admin/users" element={<UsersPage />} />
            <Route path="/admin/settings" element={<SettingsPage />} />
          </Route>
        </Route>

        {/* Fallback */}
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </AuthContext.Provider>
  );
}

export default function App() {
  const i18n = useI18nProvider();

  return (
    <ThemeProvider>
      <I18nContext.Provider value={i18n}>
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <AppRoutes />
          </BrowserRouter>
          <Toaster
            richColors
            closeButton
            position="top-center"
            offset={16}
            toastOptions={{
              classNames: {
                toast:
                  "rounded-lg border shadow-lg backdrop-blur-md bg-background/95",
              },
            }}
          />
        </QueryClientProvider>
      </I18nContext.Provider>
    </ThemeProvider>
  );
}
