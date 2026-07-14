import { cn } from '@/utils/cn.js'

export default function Button({
  variant = 'default',
  size = 'md',
  className,
  children,
  type = 'button',
  ...rest
}) {
  const base =
    'inline-flex items-center justify-center gap-2 rounded-md font-medium transition-all duration-200 ease-out disabled:opacity-50 disabled:pointer-events-none whitespace-nowrap select-none focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-ink-950'

  const sizes = {
    sm: 'h-8 px-2.5 text-xs',
    md: 'h-9 px-3 text-sm',
    lg: 'h-11 px-4 text-base'
  }

  const variants = {
    default:
      'bg-ink-900 text-white hover:bg-ink-800 dark:bg-ink-100 dark:text-ink-900 dark:hover:bg-white',
    primary:
      'bg-brand-500 text-ink-950 hover:bg-brand-400 hover:shadow-glow active:scale-[0.99]',
    outline:
      'border border-ink-200 text-ink-700 hover:border-brand-500 hover:text-brand-600 dark:border-ink-700 dark:text-ink-200',
    ghost:
      'text-ink-600 hover:bg-ink-100 dark:text-ink-300 dark:hover:bg-ink-800',
    danger:
      'bg-rose-500/10 text-rose-500 hover:bg-rose-500/20',
    link:
      'text-brand-600 underline-offset-4 hover:underline'
  }

  return (
    <button
      type={type}
      className={cn(base, sizes[size], variants[variant], className)}
      {...rest}
    >
      {children}
    </button>
  )
}
