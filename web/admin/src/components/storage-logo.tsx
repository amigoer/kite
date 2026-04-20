import { Icon } from "@iconify/react";

import { cn } from "@/lib/utils";

type SvglRoute =
  | string
  | {
      light: string;
      dark: string;
    };

type LogoVendor =
  | "s3"
  | "aws"
  | "cloudflare"
  | "aliyun"
  | "tencent"
  | "huawei"
  | "baidu"
  | "gcp"
  | "backblaze"
  | "minio"
  | "wasabi"
  | "do"
  | "ftp"
  | "local";

interface VendorMeta {
  label: string;
  color: string;
  source:
    | {
        kind: "svgl";
        route: SvglRoute;
      }
    | {
        kind: "iconify";
        icon: string;
      };
}

const STORAGE_VENDOR_META: Record<LogoVendor, VendorMeta> = {
  s3: {
    label: "S3 Compatible",
    color: "#64748B",
    source: { kind: "iconify", icon: "hugeicons:cloud-server" },
  },
  aws: {
    label: "Amazon Web Services",
    color: "#FF9900",
    source: {
      kind: "svgl",
      route: {
        light: "https://svgl.app/library/aws_light.svg",
        dark: "https://svgl.app/library/aws_dark.svg",
      },
    },
  },
  cloudflare: {
    label: "Cloudflare",
    color: "#F38020",
    source: {
      kind: "svgl",
      route: "https://svgl.app/library/cloudflare.svg",
    },
  },
  aliyun: {
    label: "Alibaba Cloud",
    color: "#FF6A00",
    source: { kind: "iconify", icon: "simple-icons:alibabacloud" },
  },
  tencent: {
    label: "Tencent Cloud",
    color: "#00A4FF",
    source: { kind: "iconify", icon: "hugeicons:cloud-server" },
  },
  huawei: {
    label: "Huawei Cloud",
    color: "#CF0A2C",
    source: { kind: "iconify", icon: "simple-icons:huawei" },
  },
  baidu: {
    label: "Baidu Cloud",
    color: "#2932E1",
    source: { kind: "iconify", icon: "simple-icons:baidu" },
  },
  gcp: {
    label: "Google Cloud",
    color: "#4285F4",
    source: {
      kind: "svgl",
      route: "https://svgl.app/library/google-cloud.svg",
    },
  },
  backblaze: {
    label: "Backblaze",
    color: "#E93D25",
    source: { kind: "iconify", icon: "simple-icons:backblaze" },
  },
  minio: {
    label: "MinIO",
    color: "#C72E49",
    source: { kind: "iconify", icon: "simple-icons:minio" },
  },
  wasabi: {
    label: "Wasabi",
    color: "#01CD3E",
    source: { kind: "iconify", icon: "simple-icons:wasabi" },
  },
  do: {
    label: "DigitalOcean",
    color: "#0080FF",
    source: {
      kind: "svgl",
      route: "https://svgl.app/library/digitalocean.svg",
    },
  },
  ftp: {
    label: "FTP",
    color: "#64748B",
    source: { kind: "iconify", icon: "arcticons:ftp-server" },
  },
  local: {
    label: "Local",
    color: "#64748B",
    source: { kind: "iconify", icon: "lucide:hard-drive" },
  },
};

export const STORAGE_VENDOR_LABELS: Record<LogoVendor, string> = Object.fromEntries(
  Object.entries(STORAGE_VENDOR_META).map(([vendor, meta]) => [vendor, meta.label]),
) as Record<LogoVendor, string>;

function normalizeDriver(driver: string): string {
  return (driver ?? "").toLowerCase();
}

export function resolveLogoVendor(provider: string | undefined, driver: string): LogoVendor {
  const p = (provider ?? "").toLowerCase();

  if (p.includes("custom-s3")) return "s3";
  if (p.includes("aws") || p.includes("amazon")) return "aws";
  if (p.includes("cloudflare") || p.includes("r2")) return "cloudflare";
  if (p.includes("aliyun") || p.includes("alibaba") || p.includes("aliyuncs")) return "aliyun";
  if (p.includes("tencent") || p.includes("cos") || p.includes("qcloud")) return "tencent";
  if (p.includes("huawei") || p.includes("obs") || p.includes("myhuaweicloud")) return "huawei";
  if (p.includes("baidu") || p.includes("bos") || p.includes("bcebos")) return "baidu";
  if (p.includes("gcp") || p.includes("google") || p.includes("gcs")) return "gcp";
  if (p.includes("backblaze") || p.includes("b2")) return "backblaze";
  if (p.includes("minio")) return "minio";
  if (p.includes("wasabi")) return "wasabi";
  if (p.includes("digitalocean") || p.includes("spaces")) return "do";

  switch (normalizeDriver(driver)) {
    case "ftp":
      return "ftp";
    case "local":
      return "local";
    case "s3":
    case "oss":
    case "cos":
    case "obs":
    case "bos":
      return "s3";
    default:
      return "local";
  }
}

export function getStorageBrandMeta(provider: string | undefined, driver: string) {
  const vendor = resolveLogoVendor(provider, driver);
  return {
    vendor,
    ...STORAGE_VENDOR_META[vendor],
  };
}

function SvglLogo({
  route,
  label,
  className,
}: {
  route: SvglRoute;
  label: string;
  className?: string;
}) {
  if (typeof route === "string") {
    return (
      <img
        src={route}
        alt={label}
        loading="lazy"
        draggable={false}
        className={cn("size-full object-contain", className)}
      />
    );
  }

  return (
    <>
      <img
        src={route.light}
        alt={label}
        loading="lazy"
        draggable={false}
        className={cn("size-full object-contain dark:hidden", className)}
      />
      <img
        src={route.dark}
        alt={label}
        loading="lazy"
        draggable={false}
        className={cn("hidden size-full object-contain dark:block", className)}
      />
    </>
  );
}

interface StorageLogoProps {
  vendor: LogoVendor;
  size?: number;
  rounded?: string;
  className?: string;
}

export function StorageLogo({
  vendor,
  size = 40,
  rounded = "rounded-lg",
  className,
}: StorageLogoProps) {
  const meta = STORAGE_VENDOR_META[vendor] ?? STORAGE_VENDOR_META.local;

  return (
    <div
      className={cn(
        "relative flex shrink-0 items-center justify-center overflow-hidden border border-border/70 bg-linear-to-br from-background to-muted/35",
        rounded,
        className,
      )}
      style={{ width: size, height: size }}
      title={meta.label}
    >
      <div
        className="flex size-[72%] items-center justify-center"
        style={{ color: meta.color }}
      >
        {meta.source.kind === "svgl" ? (
          <SvglLogo route={meta.source.route} label={meta.label} />
        ) : (
          <Icon icon={meta.source.icon} className="size-full" />
        )}
      </div>
    </div>
  );
}

export type { LogoVendor };
