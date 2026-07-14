import { forwardRef } from 'react'
import { cn } from '@/utils/cn.js'

const Textarea = forwardRef(function Textarea({ className, rows = 4, ...rest }, ref) {
  return (
    <textarea
      ref={ref}
      rows={rows}
      className={cn(
        'w-full rounded-md border border-ink-200 bg-white px-3 py-2 font-mono text-sm text-ink-900 placeholder-ink-400 transition focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500 dark:border-ink-700 dark:bg-ink-900 dark:text-ink-100',
        className
      )}
      {...rest}
    />
  )
})

export default Textarea
