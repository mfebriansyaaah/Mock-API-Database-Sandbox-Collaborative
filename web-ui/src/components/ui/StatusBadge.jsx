import { cn } from '@/utils/cn.js'

export default function StatusBadge({ status, method, className }) {
  if (method) {
    const m = method.toUpperCase()
    const cls = {
      GET: 'badge-method-get',
      POST: 'badge-method-post',
      PUT: 'badge-method-put',
      PATCH: 'badge-method-patch',
      DELETE: 'badge-method-delete'
    }[m] || 'bg-ink-200 text-ink-700'
    return <span className={cn('badge font-mono text-[10px] uppercase', cls, className)}>{m}</span>
  }
  if (status == null) return null
  const s = Number(status)
  let tone = 'bg-ink-200 text-ink-700 dark:bg-ink-800 dark:text-ink-200'
  if (s >= 200 && s < 300) tone = 'bg-brand-500/15 text-brand-600 dark:text-brand-400'
  else if (s >= 300 && s < 400) tone = 'bg-sky-500/15 text-sky-500'
  else if (s >= 400 && s < 500) tone = 'bg-amber-500/15 text-amber-500'
  else if (s >= 500) tone = 'bg-rose-500/15 text-rose-500'
  return <span className={cn('badge font-mono', tone, className)}>{s}</span>
}
