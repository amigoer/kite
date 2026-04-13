import { Link } from "react-router-dom";
import {
  Image,
  Video,
  Music,
  FileText,
  Shield,
  Zap,
  Globe,
  HardDrive,
  ArrowRight,
  Upload,
  Link2,
  FolderOpen,
} from "lucide-react";
import { useI18n } from "@/i18n";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

const fileTypes = [
  { icon: Image, label: "home.typeImage", color: "text-amber-600 dark:text-amber-400", bg: "bg-amber-50 dark:bg-amber-950/30" },
  { icon: Video, label: "home.typeVideo", color: "text-violet-600 dark:text-violet-400", bg: "bg-violet-50 dark:bg-violet-950/30" },
  { icon: Music, label: "home.typeAudio", color: "text-blue-600 dark:text-blue-400", bg: "bg-blue-50 dark:bg-blue-950/30" },
  { icon: FileText, label: "home.typeFile", color: "text-gray-600 dark:text-gray-400", bg: "bg-gray-100 dark:bg-gray-800/30" },
];

const features = [
  { icon: Upload, titleKey: "home.feat1Title", descKey: "home.feat1Desc" },
  { icon: Link2, titleKey: "home.feat2Title", descKey: "home.feat2Desc" },
  { icon: FolderOpen, titleKey: "home.feat3Title", descKey: "home.feat3Desc" },
  { icon: Shield, titleKey: "home.feat4Title", descKey: "home.feat4Desc" },
  { icon: Globe, titleKey: "home.feat5Title", descKey: "home.feat5Desc" },
  { icon: HardDrive, titleKey: "home.feat6Title", descKey: "home.feat6Desc" },
];

const steps = [
  { num: "1", titleKey: "home.step1Title", descKey: "home.step1Desc" },
  { num: "2", titleKey: "home.step2Title", descKey: "home.step2Desc" },
  { num: "3", titleKey: "home.step3Title", descKey: "home.step3Desc" },
];

export default function HomePage() {
  const { t } = useI18n();

  return (
    <div>
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 -z-10 bg-[radial-gradient(ellipse_60%_50%_at_50%_-20%,hsl(var(--foreground)/0.05),transparent)]" />
        <div className="mx-auto max-w-6xl px-6 pb-20 pt-24 text-center md:pt-32">
          <div className="mx-auto mb-6 inline-flex items-center gap-2 rounded-full border border-border bg-muted px-4 py-1.5 text-xs font-medium text-foreground">
            <Zap size={14} />
            {t("home.badge")}
          </div>
          <h1 className="mx-auto max-w-3xl text-4xl font-bold tracking-tight md:text-5xl lg:text-6xl">
            {t("home.heroTitle")}
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg text-muted-foreground">
            {t("home.heroDesc")}
          </p>
          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Button size="lg" asChild className="gap-2">
              <Link to="/register">
                {t("home.getStarted")}
                <ArrowRight size={16} />
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild>
              <Link to="/login">{t("home.loginNow")}</Link>
            </Button>
          </div>

          {/* File type badges */}
          <div className="mt-16 flex flex-wrap items-center justify-center gap-3">
            {fileTypes.map((ft) => (
              <div
                key={ft.label}
                className={`flex items-center gap-2 rounded-full border px-4 py-2 text-sm ${ft.bg}`}
              >
                <ft.icon size={16} className={ft.color} />
                <span className="text-foreground">{t(ft.label)}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="border-t bg-muted/30 py-20">
        <div className="mx-auto max-w-6xl px-6">
          <div className="text-center">
            <h2 className="text-2xl font-bold tracking-tight md:text-3xl">
              {t("home.featuresTitle")}
            </h2>
            <p className="mt-3 text-muted-foreground">
              {t("home.featuresDesc")}
            </p>
          </div>
          <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((feat) => (
              <Card
                key={feat.titleKey}
                className="border-0 shadow-sm transition-shadow hover:shadow-md"
              >
                <CardContent className="p-6">
                  <div className="mb-4 flex size-10 items-center justify-center rounded-lg bg-muted">
                    <feat.icon
                      size={20}
                      className="text-foreground"
                    />
                  </div>
                  <h3 className="font-semibold">{t(feat.titleKey)}</h3>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {t(feat.descKey)}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* How it works */}
      <section id="usage" className="py-20">
        <div className="mx-auto max-w-6xl px-6">
          <div className="text-center">
            <h2 className="text-2xl font-bold tracking-tight md:text-3xl">
              {t("home.stepsTitle")}
            </h2>
            <p className="mt-3 text-muted-foreground">
              {t("home.stepsDesc")}
            </p>
          </div>
          <div className="mt-12 grid gap-8 md:grid-cols-3">
            {steps.map((step) => (
              <div key={step.num} className="text-center">
                <div className="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-primary text-lg font-bold text-primary-foreground">
                  {step.num}
                </div>
                <h3 className="font-semibold">{t(step.titleKey)}</h3>
                <p className="mt-2 text-sm text-muted-foreground">
                  {t(step.descKey)}
                </p>
              </div>
            ))}
          </div>

          <div className="mt-16 text-center">
            <Button size="lg" asChild className="gap-2">
              <Link to="/register">
                {t("home.ctaButton")}
                <ArrowRight size={16} />
              </Link>
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
}
