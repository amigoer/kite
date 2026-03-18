import { Toaster as Sonner, type ToasterProps } from "sonner"
import { useThemeStore } from "@/stores/use-theme-store"

/**
 * Toast — Vercel 暗黑紧凑风格
 * 黑底白字、rounded-md、极小 padding
 */
const Toaster = ({ ...props }: ToasterProps) => {
  const { isDark } = useThemeStore()

  return (
    <Sonner
      theme={isDark ? "dark" : "light"}
      className="toaster group"
      position="bottom-right"
      gap={8}
      toastOptions={{
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-zinc-950 group-[.toaster]:text-white group-[.toaster]:border-zinc-800 group-[.toaster]:shadow-lg group-[.toaster]:rounded-md group-[.toaster]:px-4 group-[.toaster]:py-3",
          title: "text-sm font-medium",
          description: "text-xs text-zinc-400",
          actionButton: "text-xs font-medium text-white bg-zinc-800 px-2 py-1 rounded hover:bg-zinc-700",
          cancelButton: "text-xs text-zinc-500 hover:text-zinc-300",
        },
      }}
      {...props}
    />
  )
}

export { Toaster }
