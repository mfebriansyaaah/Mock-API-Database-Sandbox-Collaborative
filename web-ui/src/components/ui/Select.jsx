import { forwardRef } from 'react'
import { cn } from '@/utils/cn.js'

const Select = forwardRef(function Select({ className, children, ...rest }, ref) {
  return (
    <select
      ref={ref}
      className={cn(
        'w-full appearance-none rounded-md border border-ink-200 bg-white bg-[length:14px] bg-[right_0.65rem_center] bg-no-repeat px-3 py-2 pr-8 text-sm text-ink-900 transition focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500 dark:border-ink-700 dark:bg-ink-900 dark:text-ink-100',
        className
      )}
      style={{
        backgroundImage:
          "url(\"data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%2371717a' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E\")"
      }}
      {...rest}
    >
      {children}
    </select>
  )
})

export default Select
