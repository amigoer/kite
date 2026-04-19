import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useLocation, useNavigate } from "react-router-dom";
import {
  Activity,
  ArrowDown,
  ArrowRight,
  ArrowUp,
  Cpu,
  Eye,
  Files,
  HardDrive,
  Key,
  Plus,
  ShieldCheck,
  TrendingUp,
  Upload,
  User as UserIcon,
  Users,
} from "lucide-react";
import {
  adminFileApi,
  adminStatsApi,
  fileApi,
  statsApi,
  storageApi,
  tokenApi,
  userApi,
} from "@/lib/api";
import { useI18n, type Locale } from "@/i18n";
import { useAuth } from "@/hooks/use-auth";
import { cn, formatSize } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import { BrandIcon } from "@/components/storage-brand";
import {
  CardLinkFooter,
  Donut,
  type DonutSegment,
  FileThumb,
  FileThumbSkeleton,
  Heatmap,
  HeatmapLegend,
  HeroKPI,
  PageHero,
  StackedStorageBar,
  type StorageSegment,
  type ThumbFile,
  TrendCombo,
  type TrendPoint,
} from "./components";

/* ────────────────────────────────────────────────────────────
 * Types
 * ──────────────────────────────────────────────────────────── */
interface DashboardStats {
  total_files: number;
  total_size: number;
  images: number;
  videos: number;
  audios: number;
  others: number;
  users?: number;
}

interface DailyPoint {
  day: string;
  uploads: number;
  accesses: number;
  bytes_served: number;
}

interface TopUser {
  id: string;
  username: string;
  nickname?: string;
  avatar_url?: string;
  storage_used: number;
  storage_limit: number;
}

interface DashboardStorageBackend {
  id: string;
  name: string;
  driver: string;
  provider: string;
  capacity_limit_bytes: number;
  used_bytes: number;
  is_default: boolean;
  is_active: boolean;
}

/* ────────────────────────────────────────────────────────────
 * Helpers
 * ──────────────────────────────────────────────────────────── */

/** Chinese + English greeting based on hour of day. */
function buildGreeting(locale: Locale): { main: string; sub: string } {
  const h = new Date().getHours();
  if (locale === "zh") {
    if (h < 5) return { main: "夜深了", sub: "记得早些休息 🌙" };
    if (h < 11) return { main: "早安", sub: "新的一天从这里起飞" };
    if (h < 14) return { main: "午安", sub: "午后时光，轻盈整理" };
    if (h < 18) return { main: "下午好", sub: "收纳一下今天的灵感" };
    if (h < 22) return { main: "晚上好", sub: "今天又攒了不少好素材" };
    return { main: "夜安", sub: "最后看一眼，然后休息" };
  }
  if (h < 5) return { main: "Still up?", sub: "Get some rest soon 🌙" };
  if (h < 11) return { main: "Good morning", sub: "A fresh day, ready to fly" };
  if (h < 14) return { main: "Good afternoon", sub: "Tidy up in the soft midday" };
  if (h < 18) return { main: "Afternoon", sub: "File today's inspiration" };
  if (h < 22) return { main: "Good evening", sub: "Nice haul of material today" };
  return { main: "Wind down", sub: "One last look, then rest" };
}

/** Deterministic PRNG — used for the mocked heatmap & activity feed so the
 *  rendering is stable across renders without calling the backend. */
function rng(seed: number) {
  let s = seed;
  return () => {
    s = (s * 9301 + 49297) % 233280;
    return s / 233280;
  };
}

/** 7 rows (Mon-Sun) × 24 cols (hours) activity heatmap. Weighted toward work
 *  hours with a weekend damping. Replace with real data when the endpoint
 *  lands. */
function buildHeatmap(seed = 11): number[][] {
  const r = rng(seed);
  const grid: number[][] = [];
  for (let d = 0; d < 7; d++) {
    const row: number[] = [];
    for (let h = 0; h < 24; h++) {
      const workHours = h >= 9 && h <= 22 ? 1 : 0.2;
      const weekend = d >= 5 ? 0.7 : 1;
      const base = workHours * weekend;
      const val = Math.max(
        0,
        Math.round(base * (r() * 16 + 2) + r() * 4 - 2)
      );
      row.push(val);
    }
    grid.push(row);
  }
  return grid;
}

type ActivityWho = "system" | "you" | "other";
interface ActivityItem {
  who: ActivityWho;
  actor: string; // display name when who === "other"
  actionKey:
    | "uploaded"
    | "createdAlbum"
    | "revokedToken"
    | "downloaded"
    | "createdToken"
    | "backupDone";
  target: string;
  minutes: number;
}

function buildActivity(locale: Locale): ActivityItem[] {
  // Stable mocked feed — same across renders. Replace when an activity
  // endpoint exists.
  const zh = locale === "zh";
  return [
    {
      who: "other",
      actor: zh ? "苏伟明" : "Suweiming",
      actionKey: "uploaded",
      target: "street-food-014.jpg",
      minutes: 4,
    },
    {
      who: "other",
      actor: zh ? "Ada Liu" : "Ada Liu",
      actionKey: "createdAlbum",
      target: zh ? "Studio · Q2" : "Studio · Q2",
      minutes: 18,
    },
    {
      who: "you",
      actor: zh ? "你" : "You",
      actionKey: "revokedToken",
      target: zh ? "只读分享" : "read-only share",
      minutes: 42,
    },
    {
      who: "other",
      actor: zh ? "于杰" : "Yujie",
      actionKey: "downloaded",
      target: "retro-console-007.mp4",
      minutes: 66,
    },
    {
      who: "system",
      actor: zh ? "系统" : "System",
      actionKey: "backupDone",
      target: zh ? "S3 · 主存储" : "S3 · primary",
      minutes: 120,
    },
    {
      who: "other",
      actor: zh ? "刘宇" : "Liuyu",
      actionKey: "createdToken",
      target: zh ? "博客自动化" : "blog automation",
      minutes: 240,
    },
  ];
}

function actionLabel(key: ActivityItem["actionKey"], locale: Locale): string {
  if (locale === "zh") {
    return {
      uploaded: "上传了",
      createdAlbum: "新建了相册",
      revokedToken: "撤销了 Token",
      downloaded: "下载了",
      createdToken: "创建了 Token",
      backupDone: "完成备份",
    }[key];
  }
  return {
    uploaded: "uploaded",
    createdAlbum: "created album",
    revokedToken: "revoked token",
    downloaded: "downloaded",
    createdToken: "created token",
    backupDone: "finished backup",
  }[key];
}

function relativeMinutes(mins: number, locale: Locale): string {
  if (locale === "zh") {
    if (mins < 60) return `${mins} 分钟前`;
    const h = Math.floor(mins / 60);
    if (h < 24) return `${h} 小时前`;
    return `${Math.floor(h / 24)} 天前`;
  }
  if (mins < 60) return `${mins}m ago`;
  const h = Math.floor(mins / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

/* ────────────────────────────────────────────────────────────
 * DashboardPage
 * ──────────────────────────────────────────────────────────── */
export default function DashboardPage() {
  const { t, locale } = useI18n();
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();
  const isAdminWorkspace = location.pathname.startsWith("/admin");
  const displayName = user?.nickname?.trim() || user?.username || "";

  /* ── stats ──────────────────────────────────────────────── */
  const { data: stats, isLoading: statsLoading } = useQuery<DashboardStats>({
    queryKey: ["dashboard", "stats", isAdminWorkspace ? "admin" : "user"],
    queryFn: () =>
      (isAdminWorkspace ? adminStatsApi.get() : statsApi.get()).then(
        (r) => r.data.data
      ),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  /* ── daily trend (30d for chart, 7d sum for KPI) ────────── */
  const { data: daily } = useQuery<{ days: DailyPoint[] }>({
    queryKey: ["dashboard", "daily", isAdminWorkspace ? "admin" : "user"],
    queryFn: () =>
      (isAdminWorkspace
        ? adminStatsApi.daily(30)
        : statsApi.daily(30)
      ).then((r) => r.data.data),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  /* ── recent uploads (user) / top users (admin) ───────────── */
  const { data: recent, isLoading: recentLoading } = useQuery<{
    items: ThumbFile[];
  }>({
    queryKey: ["dashboard", "recent", isAdminWorkspace ? "admin" : "user"],
    queryFn: () =>
      (isAdminWorkspace
        ? adminFileApi.list({ page: 1, size: 8 })
        : fileApi.list({ page: 1, size: 8, only_self: true })
      ).then((r) => r.data.data),
    staleTime: 30_000,
  });

  const { data: topUsersData } = useQuery<{ items: TopUser[] }>({
    queryKey: ["dashboard", "topUsers"],
    enabled: isAdminWorkspace,
    queryFn: () =>
      userApi
        .list({ page: 1, size: 50 })
        .then((r) => r.data.data),
    staleTime: 60_000,
  });

  /* ── tokens count (user KPI) ────────────────────────────── */
  const { data: tokensData } = useQuery<{ items: Array<{ id: string }> }>({
    queryKey: ["dashboard", "tokens"],
    enabled: !isAdminWorkspace,
    queryFn: () => tokenApi.list().then((r) => r.data.data),
    staleTime: 60_000,
  });

  /* ── admin storage backends ─────────────────────────────── */
  const { data: backendsData } = useQuery<DashboardStorageBackend[]>({
    queryKey: ["dashboard", "storageBackends"],
    enabled: isAdminWorkspace,
    queryFn: () => storageApi.list().then((r) => r.data.data),
    staleTime: 60_000,
  });

  /* ── derived values ─────────────────────────────────────── */
  const greeting = useMemo(() => buildGreeting(locale), [locale]);
  const heatmap = useMemo(() => buildHeatmap(), []);
  const activity = useMemo(() => buildActivity(locale), [locale]);

  const days30: TrendPoint[] = useMemo(() => {
    return (daily?.days ?? []).map((d) => ({
      day: d.day,
      uploads: d.uploads,
      accesses: d.accesses,
    }));
  }, [daily?.days]);

  const last7 = useMemo(() => days30.slice(-7), [days30]);
  const weekUploads = last7.reduce((a, b) => a + b.uploads, 0);
  const weekAccesses = last7.reduce((a, b) => a + b.accesses, 0);
  const todayUploads = days30.length ? days30[days30.length - 1].uploads : 0;

  const storageSegs: StorageSegment[] = useMemo(() => {
    const s = stats ?? {
      total_files: 0,
      total_size: 0,
      images: 0,
      videos: 0,
      audios: 0,
      others: 0,
    };
    const total = Math.max(1, s.total_files);
    // Approximate per-kind bytes by proportion of count · total_size. Not
    // exact but avoids a new endpoint. Replace with bytes-per-kind when the
    // backend exposes it.
    const byKind = (count: number) =>
      total > 0 ? Math.round((count / total) * s.total_size) : 0;
    return [
      {
        kind: "image",
        label: t("dashboard.images"),
        count: s.images,
        bytes: byKind(s.images),
        color: "hsl(var(--chart-3))",
      },
      {
        kind: "video",
        label: t("dashboard.videos"),
        count: s.videos,
        bytes: byKind(s.videos),
        color: "hsl(var(--chart-2))",
      },
      {
        kind: "audio",
        label: t("dashboard.audio"),
        count: s.audios,
        bytes: byKind(s.audios),
        color: "hsl(var(--chart-1))",
      },
      {
        kind: "other",
        label: t("dashboard.otherFiles"),
        count: s.others,
        bytes: byKind(s.others),
        color: "hsl(var(--chart-4))",
      },
    ];
  }, [stats, t]);

  const donutSegs: DonutSegment[] = storageSegs.map((s) => ({
    value: s.count,
    color: s.color,
    label: s.label,
  }));
  const totalBytes = storageSegs.reduce((a, b) => a + b.bytes, 0);
  const totalCount = storageSegs.reduce((a, b) => a + b.count, 0);

  const storageLimit = user?.storage_limit ?? 0;
  const storageUsed = stats?.total_size ?? 0;
  const isUnlimited = storageLimit < 0;
  const hasLimit = storageLimit > 0;
  const storagePct = hasLimit
    ? Math.min((storageUsed / storageLimit) * 100, 100)
    : 0;

  const topUsers = useMemo(
    () =>
      [...(topUsersData?.items ?? [])]
        .sort((a, b) => (b.storage_used ?? 0) - (a.storage_used ?? 0))
        .slice(0, 5),
    [topUsersData?.items]
  );

  /* ── hero content ───────────────────────────────────────── */
  const heroTitle = isAdminWorkspace
    ? `${greeting.main}${locale === "zh" ? "，" : ", "}${t("nav.roleAdmin")}`
    : `${greeting.main}${locale === "zh" ? "，" : ", "}${displayName}`;

  const heroSubUser =
    locale === "zh"
      ? `${greeting.sub} · 今天已同步 ${todayUploads} 个文件`
      : `${greeting.sub} · ${todayUploads} file${
          todayUploads === 1 ? "" : "s"
        } synced today`;
  const heroSubAdmin = t("dashboard.hero.adminSub");

  /* ── admin KPI derived counts ───────────────────────────── */
  const activeBackends = (backendsData ?? []).filter((b) => b.is_active).length;
  const totalBackends = (backendsData ?? []).length;
  const activeUsers = (topUsersData?.items ?? []).filter(
    (u) => (u.storage_used ?? 0) > 0,
  ).length;

  /* ── render ─────────────────────────────────────────────── */
  return (
    <div className="page-enter flex flex-col gap-5 sm:gap-6">
      {/* ═════ HERO ═════════════════════════════════════════ */}
      <PageHero>
        <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between lg:gap-8">
          <div className="flex min-w-0 items-start gap-3 sm:gap-4">
            {isAdminWorkspace ? (
              <div className="flex size-12 shrink-0 items-center justify-center rounded-full border bg-background/70 backdrop-blur sm:size-[52px]">
                <ShieldCheck className="size-5" />
              </div>
            ) : (
              <Avatar className="size-12 shrink-0 ring-2 ring-background sm:size-[52px]">
                <AvatarImage src={user?.avatar_url} alt={displayName} />
                <AvatarFallback className="bg-foreground text-sm font-medium text-background">
                  {displayName.charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
            )}
            <div className="min-w-0 space-y-1.5">
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="text-xl font-semibold leading-tight tracking-tight sm:text-[26px]">
                  {heroTitle}
                </h1>
                {isAdminWorkspace ? (
                  <span className="inline-flex items-center gap-1 rounded-md border border-violet-500/20 bg-violet-500/10 px-1.5 py-0.5 text-[10px] font-medium text-violet-700 dark:text-violet-400">
                    <ShieldCheck className="size-3" />
                    ADMIN
                  </span>
                ) : (
                  user?.username && (
                    <span className="hidden rounded-md border bg-background/60 px-2 py-0.5 text-[11px] font-medium text-muted-foreground backdrop-blur sm:inline-flex">
                      @{user.username}
                    </span>
                  )
                )}
              </div>
              <p className="text-sm text-muted-foreground">
                {isAdminWorkspace ? heroSubAdmin : heroSubUser}
              </p>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {isAdminWorkspace ? (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => navigate("/admin/users")}
                >
                  <Users className="size-3.5" />
                  {t("dashboard.hero.inviteUsers")}
                </Button>
                <Button
                  size="sm"
                  onClick={() => navigate("/admin/storage")}
                >
                  <HardDrive className="size-3.5" />
                  {t("dashboard.hero.newBackend")}
                </Button>
              </>
            ) : (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => navigate("/user/folders")}
                >
                  <Plus className="size-3.5" />
                  {t("dashboard.hero.newFolder")}
                </Button>
                <Button
                  size="sm"
                  onClick={() => navigate("/user/files")}
                >
                  <Upload className="size-3.5" />
                  {t("common.upload")}
                </Button>
              </>
            )}
          </div>
        </div>

        {/* KPI strip */}
        <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
          {isAdminWorkspace ? (
            <>
              <HeroKPI
                label={t("dashboard.kpi.totalUsers")}
                value={(stats?.users ?? 0).toLocaleString()}
                delta={t("dashboard.kpi.activeThisWeek").replace(
                  "{count}",
                  String(activeUsers),
                )}
                deltaIcon={<ArrowUp className="size-3" />}
                accent="hsl(var(--chart-2))"
              />
              <HeroKPI
                label={t("dashboard.kpi.totalResources")}
                value={(stats?.total_files ?? 0).toLocaleString()}
                delta={t("dashboard.kpi.newThisWeek").replace(
                  "{count}",
                  String(weekUploads),
                )}
                deltaIcon={<TrendingUp className="size-3" />}
                accent="hsl(var(--chart-3))"
              />
              <HeroKPI
                label={t("dashboard.kpi.storageUsed")}
                value={formatSize(stats?.total_size ?? 0)}
                delta={t("dashboard.kpi.backendsCount")
                  .replace("{total}", String(totalBackends))
                  .replace("{active}", String(activeBackends))}
                accent="hsl(var(--chart-1))"
              />
              <HeroKPI
                label={t("dashboard.kpi.weeklyAccesses")}
                value={weekAccesses.toLocaleString()}
                delta={t("dashboard.kpi.apiPlusDownloads")}
                deltaIcon={<Activity className="size-3" />}
                accent="hsl(var(--chart-5))"
              />
            </>
          ) : (
            <>
              <HeroKPI
                label={t("dashboard.totalFiles")}
                value={(stats?.total_files ?? 0).toLocaleString()}
                delta={
                  locale === "zh"
                    ? `本周 +${weekUploads}`
                    : `+${weekUploads} this week`
                }
                deltaIcon={<TrendingUp className="size-3" />}
                accent="hsl(var(--chart-3))"
              />
              <HeroKPI
                label={t("dashboard.storageUsed")}
                value={formatSize(storageUsed)}
                delta={
                  isUnlimited
                    ? t("dashboard.unlimitedStorage")
                    : hasLimit
                      ? locale === "zh"
                        ? `共 ${formatSize(storageLimit)} · ${Math.round(storagePct)}%`
                        : `of ${formatSize(storageLimit)} · ${Math.round(storagePct)}%`
                      : "—"
                }
                progress={hasLimit ? storagePct : undefined}
                accent="hsl(var(--chart-2))"
              />
              <HeroKPI
                label={t("dashboard.kpi.weeklyAccesses")}
                value={weekAccesses.toLocaleString()}
                delta={t("dashboard.kpi.resourceRefs")}
                deltaIcon={<Eye className="size-3" />}
                accent="hsl(var(--chart-1))"
              />
              <HeroKPI
                label={t("dashboard.kpi.activeTokens")}
                value={(tokensData?.items?.length ?? 0).toLocaleString()}
                delta={
                  locale === "zh"
                    ? `共 ${tokensData?.items?.length ?? 0} 个 Token`
                    : `${tokensData?.items?.length ?? 0} total`
                }
                deltaIcon={<Key className="size-3" />}
                accent="hsl(var(--chart-5))"
              />
            </>
          )}
        </div>
      </PageHero>

      {/* ═════ ROW 1 ═════════════════════════════════════════ */}
      {isAdminWorkspace ? (
        <div className="grid gap-4 lg:grid-cols-5">
          <StorageBackendsCard
            backends={backendsData ?? []}
            onManage={() => navigate("/admin/storage")}
            t={t}
          />
          <Card className="gap-4 py-5 shadow-xs lg:col-span-2">
            <CardHeader className="px-5">
              <CardTitle className="text-sm">
                {t("dashboard.composition.title")}
              </CardTitle>
              <CardDescription className="text-xs">
                {t("dashboard.composition.sub")}
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col items-center gap-4 px-5 pb-1">
              {statsLoading ? (
                <Skeleton className="size-40 rounded-full" />
              ) : (
                <>
                  <Donut
                    segments={donutSegs}
                    total={totalCount}
                    size={160}
                    stroke={18}
                    label={t("dashboard.composition.centerLabel")}
                  />
                  <div className="grid w-full grid-cols-2 gap-x-4 gap-y-1.5">
                    {storageSegs.map((s) => {
                      const pct =
                        totalCount > 0
                          ? Math.round((s.count / totalCount) * 100)
                          : 0;
                      return (
                        <div
                          key={s.kind}
                          className="flex items-center gap-2 text-[11px]"
                        >
                          <span
                            className="size-2 shrink-0 rounded-sm"
                            style={{ background: s.color }}
                          />
                          <span className="truncate text-muted-foreground">
                            {s.label}
                          </span>
                          <span className="ml-auto tabular-nums text-foreground/80">
                            {pct}%
                          </span>
                        </div>
                      );
                    })}
                  </div>
                </>
              )}
            </CardContent>
          </Card>
        </div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-5">
          <Card className="gap-4 py-5 shadow-xs lg:col-span-3">
            <CardHeader className="px-5">
              <CardTitle className="text-sm">
                {t("dashboard.storage.breakdownTitle")}
              </CardTitle>
              <CardDescription className="text-xs">
                {t("dashboard.storage.breakdownSub")}
              </CardDescription>
              <CardAction>
                <span className="inline-flex items-center gap-1 rounded-md border bg-background/50 px-2 py-0.5 text-[11px] tabular-nums text-muted-foreground">
                  <HardDrive className="size-3" />
                  {formatSize(totalBytes)}
                </span>
              </CardAction>
            </CardHeader>
            <CardContent className="space-y-5 px-5">
              {statsLoading ? (
                <>
                  <Skeleton className="h-2.5 w-full rounded-full" />
                  <div className="grid grid-cols-2 gap-x-5 gap-y-3 sm:grid-cols-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                      <div key={`sk-pct-${i}`} className="space-y-1.5">
                        <Skeleton className="h-3 w-20" />
                        <Skeleton className="h-2 w-14" />
                      </div>
                    ))}
                  </div>
                  <div className="grid grid-cols-2 gap-x-5 gap-y-4 pt-1 sm:grid-cols-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                      <div key={`sk-cnt-${i}`} className="space-y-2">
                        <Skeleton className="h-2.5 w-12" />
                        <Skeleton className="h-7 w-16" />
                        <Skeleton className="h-2 w-10" />
                      </div>
                    ))}
                  </div>
                </>
              ) : (
                <StackedStorageBar data={storageSegs} total={totalBytes} />
              )}
            </CardContent>
          </Card>

          <Card className="gap-4 py-5 shadow-xs lg:col-span-2">
            <CardHeader className="px-5">
              <CardTitle className="text-sm">
                {t("dashboard.composition.title")}
              </CardTitle>
              <CardDescription className="text-xs">
                {t("dashboard.composition.sub")}
              </CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-center px-5">
              {statsLoading ? (
                <Skeleton className="size-40 rounded-full" />
              ) : (
                <Donut
                  segments={donutSegs}
                  total={totalCount}
                  size={160}
                  stroke={18}
                  label={t("dashboard.composition.centerLabel")}
                />
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* ═════ ROW 2: Trend + Heatmap ══════════════════════ */}
      <div className="grid gap-4 lg:grid-cols-5">
        <Card className="min-w-0 gap-3 overflow-hidden py-5 shadow-xs lg:col-span-3">
          <CardHeader className="px-5">
            <CardTitle className="text-sm">
              {isAdminWorkspace
                ? t("dashboard.trend.adminTitle")
                : t("dashboard.trend.title")}
            </CardTitle>
            <CardDescription className="text-xs">
              {isAdminWorkspace
                ? t("dashboard.trend.adminSub")
                : t("dashboard.trend.sub")}
            </CardDescription>
            <CardAction>
              <div className="flex items-center gap-3 text-[11px] text-muted-foreground">
                <span className="flex items-center gap-1">
                  <span
                    className="size-2 rounded-sm"
                    style={{ background: "hsl(var(--chart-3))" }}
                  />
                  {t("dashboard.uploads")}
                </span>
                <span className="flex items-center gap-1">
                  <span
                    className="size-2 rounded-sm"
                    style={{ background: "hsl(var(--chart-2))" }}
                  />
                  {t("dashboard.accesses")}
                </span>
              </div>
            </CardAction>
          </CardHeader>
          <CardContent className="min-w-0 px-5">
            <div className="h-44 w-full sm:h-56">
              {!daily ? (
                <Skeleton className="h-full w-full" />
              ) : (
                <TrendCombo
                  data={days30}
                  height={220}
                  uploadsLabel={t("dashboard.uploads")}
                  accessesLabel={t("dashboard.accesses")}
                />
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="min-w-0 gap-3 overflow-hidden py-5 shadow-xs lg:col-span-2">
          <CardHeader className="px-5">
            <CardTitle className="text-sm">
              {isAdminWorkspace
                ? t("dashboard.heatmap.adminTitle")
                : t("dashboard.heatmap.title")}
            </CardTitle>
            <CardDescription className="text-xs">
              {isAdminWorkspace
                ? t("dashboard.heatmap.adminSub")
                : t("dashboard.heatmap.sub")}
            </CardDescription>
          </CardHeader>
          <CardContent className="min-w-0 px-5">
            <Heatmap
              grid={heatmap}
              weekdayLabels={[
                t("dashboard.weekdays.mon"),
                t("dashboard.weekdays.tue"),
                t("dashboard.weekdays.wed"),
                t("dashboard.weekdays.thu"),
                t("dashboard.weekdays.fri"),
                t("dashboard.weekdays.sat"),
                t("dashboard.weekdays.sun"),
              ]}
            />
            <HeatmapLegend
              lowLabel={t("dashboard.heatmap.low")}
              highLabel={t("dashboard.heatmap.high")}
            />
          </CardContent>
        </Card>
      </div>

      {/* ═════ ROW 3 ═════════════════════════════════════════ */}
      {isAdminWorkspace ? (
        <div className="grid gap-4 lg:grid-cols-5">
          <TopUsersCard users={topUsers} locale={locale} t={t} />
          <SystemStatusCard t={t} />
        </div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-5">
          <RecentUploadsCard
            items={recent?.items ?? []}
            isLoading={recentLoading}
            locale={locale}
            t={t}
          />

          <Card className="gap-3 py-5 shadow-xs lg:col-span-2">
            <CardHeader className="px-5">
              <CardTitle className="text-sm">
                {t("dashboard.activity.title")}
              </CardTitle>
              <CardDescription className="text-xs">
                {t("dashboard.activity.sub")}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 px-5">
              {activity.map((a, i) => (
                <div key={i} className="flex items-start gap-3">
                  <div
                    className={cn(
                      "mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full text-xs font-medium",
                      a.who === "system"
                        ? "bg-muted text-muted-foreground"
                        : a.who === "you"
                          ? "bg-foreground text-background"
                          : "bg-muted",
                    )}
                  >
                    {a.who === "system" ? (
                      <Cpu className="size-4" />
                    ) : a.who === "you" ? (
                      <UserIcon className="size-4" />
                    ) : (
                      a.actor.charAt(0).toUpperCase()
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm leading-snug">
                      <b className="font-medium">{a.actor}</b>
                      <span className="text-muted-foreground">
                        {" "}
                        {actionLabel(a.actionKey, locale)}{" "}
                      </span>
                      <span className="truncate">
                        {locale === "zh" ? "「" : "\u201c"}
                        {a.target}
                        {locale === "zh" ? "」" : "\u201d"}
                      </span>
                    </p>
                    <p className="mt-1 text-[11px] tabular-nums text-muted-foreground">
                      {relativeMinutes(a.minutes, locale)}
                    </p>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

/* ────────────────────────────────────────────────────────────
 * RecentUploadsCard — user-dashboard 3rd row, left column
 * ──────────────────────────────────────────────────────────── */
function RecentUploadsCard({
  items,
  isLoading,
  locale,
  t,
}: {
  items: ThumbFile[];
  isLoading: boolean;
  locale: Locale;
  t: (k: string) => string;
}) {
  return (
    <Card className="gap-4 py-5 shadow-xs lg:col-span-3">
      <CardHeader className="px-5">
        <CardTitle className="text-sm">
          {t("dashboard.recentUploads")}
        </CardTitle>
        <CardDescription className="text-xs">
          {t("dashboard.recentUploadsDesc").replace(
            "{count}",
            String(items.length),
          )}
        </CardDescription>
        <CardAction>
          <CardLinkFooter to="/user/files">
            {t("dashboard.viewAll")}
          </CardLinkFooter>
        </CardAction>
      </CardHeader>
      <CardContent className="px-5">
        {isLoading ? (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <FileThumbSkeleton key={i} />
            ))}
          </div>
        ) : items.length === 0 ? (
          <div className="py-12 text-center text-sm text-muted-foreground">
            {t("dashboard.noRecentFiles")}
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4">
            {items.slice(0, 8).map((f) => (
              <FileThumb key={f.id} file={f} locale={locale} />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

/* ────────────────────────────────────────────────────────────
 * TopUsersCard — admin-dashboard 3rd row, left column
 * ──────────────────────────────────────────────────────────── */
function TopUsersCard({
  users,
  locale,
  t,
}: {
  users: TopUser[];
  locale: Locale;
  t: (k: string) => string;
}) {
  return (
    <Card className="gap-3 py-5 shadow-xs lg:col-span-3">
      <CardHeader className="px-5">
        <CardTitle className="text-sm">
          {t("dashboard.topUsers.title")}
        </CardTitle>
        <CardDescription className="text-xs">
          {t("dashboard.topUsers.sub")}
        </CardDescription>
        <CardAction>
          <CardLinkFooter to="/admin/users">
            {t("dashboard.topUsers.all")}
          </CardLinkFooter>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-2.5 px-5">
        {users.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            {t("common.noData")}
          </div>
        ) : (
          users.map((u) => {
            const pct =
              u.storage_limit > 0
                ? Math.min(100, (u.storage_used / u.storage_limit) * 100)
                : 0;
            const name = u.nickname?.trim() || u.username;
            return (
              <div
                key={u.id}
                className="flex items-center gap-3 rounded-lg px-1.5 py-1.5 transition-colors hover:bg-muted/40"
              >
                <Avatar className="size-8">
                  <AvatarImage src={u.avatar_url} alt={name} />
                  <AvatarFallback className="text-[11px] font-medium">
                    {name.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="min-w-0 flex-1">
                  <div className="flex items-baseline justify-between gap-2">
                    <span className="truncate text-sm font-medium">
                      {name}{" "}
                      <span className="text-muted-foreground">
                        @{u.username}
                      </span>
                    </span>
                    <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground">
                      {formatSize(u.storage_used)}
                      {u.storage_limit > 0 && (
                        <span>
                          {" "}
                          /{" "}
                          {u.storage_limit < 0
                            ? locale === "zh"
                              ? "∞"
                              : "∞"
                            : formatSize(u.storage_limit)}
                        </span>
                      )}
                    </span>
                  </div>
                  <div className="mt-1.5 h-1 w-full overflow-hidden rounded-full bg-muted">
                    <div
                      className="h-full bg-foreground/80 transition-[width]"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>
              </div>
            );
          })
        )}
      </CardContent>
    </Card>
  );
}

/* ────────────────────────────────────────────────────────────
 * SystemStatusCard — admin dashboard Row 3, right (lg:col-span-2)
 *
 * Matches the target layout 1:1:
 *  • Two semi-circle gauges (CPU / 内存)
 *  • Bandwidth tile (上行 / 下行)
 *  • 4 metric tiles (API 延迟 / 磁盘 I/O / 活跃连接 / 错误率)
 *  • Emerald "all operational" banner
 * ──────────────────────────────────────────────────────────── */

/** Semi-circle gauge. `value` is 0..100, `display` is the string shown inside
 *  the arc (e.g. "12%" or "1.2 / 4 GB"). */
function Gauge({
  label,
  value,
  display,
  color,
}: {
  label: string;
  value: number;
  display: string;
  color: string;
}) {
  const r = 30;
  const cx = 36;
  const cy = 36;
  const circumference = Math.PI * r;
  const pct = Math.max(0, Math.min(100, value));
  const dash = (pct / 100) * circumference;
  const arcPath = `M ${cx - r} ${cy} A ${r} ${r} 0 0 1 ${cx + r} ${cy}`;
  return (
    <div className="flex flex-col items-center rounded-lg border p-2.5">
      <div className="relative" style={{ width: 72, height: 42 }}>
        <svg width="72" height="42" viewBox="0 0 72 42">
          <path
            d={arcPath}
            fill="none"
            stroke="hsl(var(--muted))"
            strokeWidth="6"
            strokeLinecap="round"
          />
          <path
            d={arcPath}
            fill="none"
            stroke={color}
            strokeWidth="6"
            strokeLinecap="round"
            strokeDasharray={`${dash} ${circumference}`}
            style={{ transition: "stroke-dasharray 400ms ease" }}
          />
        </svg>
        <div className="absolute inset-x-0 bottom-0 text-center text-[13px] font-semibold tabular-nums">
          {display}
        </div>
      </div>
      <div className="mt-1 text-[11px] text-muted-foreground">{label}</div>
    </div>
  );
}

/** Two-column bandwidth tile showing upload and download speed. */
function BandwidthTile({
  upLabel,
  downLabel,
  up,
  down,
  upPct = 0,
  downPct = 0,
}: {
  upLabel: string;
  downLabel: string;
  up: string;
  down: string;
  upPct?: number;
  downPct?: number;
}) {
  return (
    <div className="grid grid-cols-2 gap-0 overflow-hidden rounded-lg border">
      <div className="flex items-center gap-2.5 border-r p-2.5">
        <div
          className="flex size-8 shrink-0 items-center justify-center rounded-md"
          style={{
            background: "hsl(var(--chart-4) / 0.12)",
            color: "hsl(var(--chart-4))",
          }}
        >
          <ArrowUp className="size-4" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="text-[10px] uppercase tracking-wide text-muted-foreground">
            {upLabel}
          </div>
          <div className="truncate text-[13px] font-semibold tabular-nums">
            {up}
          </div>
          <div className="mt-1 h-0.5 w-full overflow-hidden rounded-full bg-muted">
            <div
              className="h-full"
              style={{
                width: `${upPct}%`,
                background: "hsl(var(--chart-4))",
              }}
            />
          </div>
        </div>
      </div>
      <div className="flex items-center gap-2.5 p-2.5">
        <div
          className="flex size-8 shrink-0 items-center justify-center rounded-md"
          style={{
            background: "hsl(var(--chart-1) / 0.12)",
            color: "hsl(var(--chart-1))",
          }}
        >
          <ArrowDown className="size-4" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="text-[10px] uppercase tracking-wide text-muted-foreground">
            {downLabel}
          </div>
          <div className="truncate text-[13px] font-semibold tabular-nums">
            {down}
          </div>
          <div className="mt-1 h-0.5 w-full overflow-hidden rounded-full bg-muted">
            <div
              className="h-full"
              style={{
                width: `${downPct}%`,
                background: "hsl(var(--chart-1))",
              }}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

/** Small metric card: colored dot + label + value. */
function MetricTile({
  label,
  value,
  dot,
}: {
  label: string;
  value: string;
  dot: string;
}) {
  return (
    <div className="rounded-lg border p-2.5">
      <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
        <span
          className="size-1.5 rounded-full"
          style={{ background: dot }}
        />
        {label}
      </div>
      <div className="mt-1 text-[13px] font-semibold tabular-nums">{value}</div>
    </div>
  );
}

function SystemStatusCard({ t }: { t: (k: string) => string }) {
  // Live metrics are not yet exposed by the API — use the same mocked values
  // as the design target until an endpoint lands. (See pages-admin.jsx.)
  const uptimeDays = 7;
  return (
    <Card className="gap-3 py-5 shadow-xs lg:col-span-2">
      <CardHeader className="px-5">
        <CardTitle className="text-sm">
          {t("dashboard.systemStatus.title")}
        </CardTitle>
        <CardDescription className="text-xs">
          {t("dashboard.systemStatus.subRefresh")}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3 px-5">
        <div className="grid grid-cols-2 gap-2.5">
          <Gauge
            label={t("dashboard.systemStatus.cpu")}
            value={12}
            display="12%"
            color="hsl(var(--chart-3))"
          />
          <Gauge
            label={t("dashboard.systemStatus.memory")}
            value={30}
            display="1.2 / 4 GB"
            color="hsl(var(--chart-2))"
          />
        </div>
        <BandwidthTile
          upLabel={t("dashboard.systemStatus.upload")}
          downLabel={t("dashboard.systemStatus.download")}
          up="38 Mbps"
          down="112 Mbps"
          upPct={19}
          downPct={22}
        />
        <div className="grid grid-cols-2 gap-2">
          <MetricTile
            label={t("dashboard.systemStatus.apiLatency")}
            value="46 ms"
            dot="hsl(var(--chart-5))"
          />
          <MetricTile
            label={t("dashboard.systemStatus.diskIO")}
            value="28 MB/s"
            dot="hsl(var(--chart-3))"
          />
          <MetricTile
            label={t("dashboard.systemStatus.activeConnections")}
            value="184"
            dot="hsl(var(--chart-2))"
          />
          <MetricTile
            label={t("dashboard.systemStatus.errorRate")}
            value="0.02%"
            dot="hsl(var(--chart-1))"
          />
        </div>
        <div className="flex items-center gap-2 rounded-lg border bg-emerald-500/5 p-2.5 text-[11px]">
          <span className="size-2 rounded-full bg-emerald-500" />
          <span className="font-medium text-emerald-700 dark:text-emerald-400">
            {t("dashboard.systemStatus.allOperational")}
          </span>
          <span className="ml-auto text-muted-foreground">
            {t("dashboard.systemStatus.uptimeDays").replace(
              "{days}",
              String(uptimeDays),
            )}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}

/* ────────────────────────────────────────────────────────────
 * StorageBackendsCard — admin dashboard Row 1, left (lg:col-span-3)
 *
 * Full-width rows: logo · name+badge · used/total · mono path · progress.
 * When there are more than 3 backends, a dashed-border footer card shows
 * stacked logos and "还有 N 个后端".
 * ──────────────────────────────────────────────────────────── */
function StorageBackendsCard({
  backends,
  onManage,
  t,
}: {
  backends: DashboardStorageBackend[];
  onManage: () => void;
  t: (k: string) => string;
}) {
  const sorted = [...backends].sort(
    (a, b) => Number(b.is_default) - Number(a.is_default),
  );
  const visible = sorted.slice(0, 3);
  const overflow = sorted.slice(3, 6);
  const remaining = Math.max(0, sorted.length - 3);

  return (
    <Card className="gap-3 py-5 shadow-xs lg:col-span-3">
      <CardHeader className="px-5">
        <CardTitle className="text-sm">
          {t("dashboard.backends.title")}
        </CardTitle>
        <CardDescription className="text-xs">
          {t("dashboard.backends.sub")}
        </CardDescription>
        <CardAction>
          <Button
            variant="ghost"
            size="xs"
            onClick={onManage}
            className="text-muted-foreground hover:text-foreground"
          >
            {t("dashboard.backends.all")}
            <ArrowRight className="size-3" />
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-3 px-5">
        {visible.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            {t("dashboard.backends.noBackends")}
          </div>
        ) : (
          visible.map((b) => {
            const hasLimit = b.capacity_limit_bytes > 0;
            const pct = hasLimit
              ? Math.min(100, (b.used_bytes / b.capacity_limit_bytes) * 100)
              : 0;
            const path = [b.driver, b.provider].filter(Boolean).join(" · ");
            return (
              <div
                key={b.id}
                className="flex items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-muted/30"
              >
                <BrandIcon
                  provider={b.provider}
                  driver={b.driver}
                  className="size-8 shrink-0 rounded-md"
                />
                <div className="min-w-0 flex-1">
                  <div className="flex items-baseline justify-between gap-3">
                    <div className="flex min-w-0 items-center gap-2">
                      <span className="truncate text-sm font-medium">
                        {b.name}
                      </span>
                      {b.is_active ? (
                        <span className="shrink-0 rounded-md border border-emerald-500/30 bg-emerald-500/10 px-1.5 py-px text-[10px] font-medium text-emerald-700 dark:text-emerald-400">
                          active
                        </span>
                      ) : (
                        <span className="shrink-0 rounded-md border bg-muted/50 px-1.5 py-px text-[10px] font-medium text-muted-foreground">
                          idle
                        </span>
                      )}
                      {b.is_default && (
                        <span className="shrink-0 rounded-md border bg-muted/50 px-1.5 py-px text-[10px] font-medium uppercase text-muted-foreground">
                          {t("dashboard.backends.defaultBadge")}
                        </span>
                      )}
                    </div>
                    <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground">
                      {formatSize(b.used_bytes)} /{" "}
                      {hasLimit ? formatSize(b.capacity_limit_bytes) : "∞"}
                    </span>
                  </div>
                  <p className="mt-0.5 truncate font-mono text-[10px] text-muted-foreground">
                    {path || b.id}
                  </p>
                  <div className="mt-2 h-1 w-full overflow-hidden rounded-full bg-muted">
                    <div
                      className={cn(
                        "h-full transition-[width]",
                        b.is_active
                          ? "bg-foreground/70"
                          : "bg-muted-foreground/40",
                      )}
                      style={{
                        width: hasLimit ? `${pct}%` : "100%",
                        opacity: hasLimit ? 1 : 0.3,
                      }}
                    />
                  </div>
                </div>
              </div>
            );
          })
        )}
        {remaining > 0 && (
          <button
            type="button"
            onClick={onManage}
            className="group flex w-full items-center justify-between rounded-lg border border-dashed p-3 text-left transition-colors hover:border-foreground/30 hover:bg-muted/30"
          >
            <div className="flex items-center gap-3">
              <div className="flex -space-x-1.5">
                {overflow.map((b) => (
                  <BrandIcon
                    key={b.id}
                    provider={b.provider}
                    driver={b.driver}
                    className="size-6 rounded-md ring-2 ring-card"
                  />
                ))}
              </div>
              <span className="text-xs font-medium text-muted-foreground">
                {t("dashboard.backends.moreCount").replace(
                  "{count}",
                  String(remaining),
                )}
              </span>
            </div>
            <span className="text-[11px] font-medium text-muted-foreground group-hover:text-foreground">
              {t("dashboard.backends.viewAllMore")}
            </span>
          </button>
        )}
      </CardContent>
    </Card>
  );
}

/* Keep a no-op export of Files icon to satisfy tooling that inspects
 * imports from the old entry; TS allows unused imports to tree-shake. */
export { Files };
