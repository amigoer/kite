import { useQuery } from "@tanstack/react-query";
import { Image, Video, Music, FileText, HardDrive, Users } from "lucide-react";
import { statsApi } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

function formatBytes(bytes: number) {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function DashboardPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["stats"],
    queryFn: () => statsApi.get().then((r) => r.data.data),
  });

  const cards = [
    {
      title: "Total Files",
      value: data?.total_files ?? 0,
      icon: FileText,
    },
    {
      title: "Storage Used",
      value: data ? formatBytes(data.total_size) : "0 B",
      icon: HardDrive,
    },
    { title: "Images", value: data?.images ?? 0, icon: Image },
    { title: "Videos", value: data?.videos ?? 0, icon: Video },
    { title: "Audio", value: data?.audios ?? 0, icon: Music },
    { title: "Users", value: data?.users ?? 0, icon: Users },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Overview of your media hosting
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {cards.map((card) => (
          <Card key={card.title}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {card.title}
              </CardTitle>
              <card.icon className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <Skeleton className="h-7 w-20" />
              ) : (
                <div className="text-2xl font-bold">{card.value}</div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
