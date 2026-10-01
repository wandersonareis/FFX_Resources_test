import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "cn"

const alertVariants = cva(
  "group/alert relative grid w-full grid-cols-[0_auto] items-start gap-y-0.5 rounded-4xl border px-3.5 py-2 text-xs has-[>svg]:grid-cols-[auto_auto] has-[>svg]:gap-x-2.5 [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current",
  {
    variants: {
      variant: {
        default: "bg-card text-card-foreground",
        info: "bg-primary/10 text-primary dark:bg-primary/20 [border-color:color-mix(in_oklab,var(--primary)_35%,transparent)]",
        warning:
          "bg-amber-500/10 text-amber-700 dark:text-amber-400 [border-color:color-mix(in_oklab,var(--amber-500)_35%,transparent)]",
        destructive:
          "bg-destructive/10 text-destructive dark:bg-destructive/20 [border-color:color-mix(in_oklab,var(--destructive)_35%,transparent)]",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

function Alert({
  className,
  variant,
  ...props
}: React.ComponentProps<"div"> & VariantProps<typeof alertVariants>) {
  return (
    <div
      role="alert"
      data-slot="alert"
      data-variant={variant}
      className={cn(alertVariants({ variant }), className)}
      {...props}
    />
  )
}

function AlertTitle({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-title"
      className={cn(
        "col-start-2 line-clamp-1 min-h-4 font-medium tracking-tight",
        className
      )}
      {...props}
    />
  )
}

function AlertDescription({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-description"
      className={cn(
        "col-start-2 grid justify-items-start gap-1 text-muted-foreground text-sm [&_p]:leading-relaxed",
        className
      )}
      {...props}
    />
  )
}

export { Alert, AlertTitle, AlertDescription, alertVariants }
