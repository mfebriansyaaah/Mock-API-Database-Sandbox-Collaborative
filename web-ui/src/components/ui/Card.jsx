import { cn } from '@/utils/cn.js'

export function Card({ as: Component = 'div', className, children, ...rest }) {
  return (
    <Component
      className={cn(
        'rounded-xl border border-ink-200 bg-white shadow-sm dark:border-ink-800 dark:bg-ink-900',
        className
      )}
      {...rest}
    >
      {children}
    </Component>
  )
}

export default Card

export function CardHeader({ className, children, ...rest }) {
  return (
    <div
      className={cn('flex items-center justify-between border-b border-ink-200 px-5 py-3 dark:border-ink-800', className)}
      {...rest}
    >
      {children}
    </div>
  )
}

export function CardTitle({ className, children }) {
  return (
    <h3 className={cn('font-display text-base font-semibold tracking-tight', className)}>
      {children}
    </h3>
  )
}

export function CardBody({ className, children }) {
  return <div className={cn('p-5', className)}>{children}</div>
}
