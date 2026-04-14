import * as React from "react"
import { cn } from "@/lib/utils"

interface ProgressProps extends React.ComponentProps<"div"> {
  value?: number
  indicatorClassName?: string
}

function Progress({
  className,
  value = 0,
  indicatorClassName,
  ...props
}: ProgressProps) {
  return (
    <div
      data-slot="progress"
      role="progressbar"
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={100}
      className={cn(
        "bg-secondary relative h-2 w-full overflow-hidden rounded-full",
        className
      )}
      {...props}
    >
      <div
        className={cn(
          "bg-primary h-full rounded-full transition-all duration-300 ease-in-out",
          indicatorClassName
        )}
        style={{ width: `${Math.max(0, Math.min(100, value))}%` }}
      />
    </div>
  )
}

export { Progress }
