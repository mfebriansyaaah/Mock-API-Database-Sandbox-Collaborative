import CountUp from './CountUp.jsx'
import { cn } from '@/utils/cn.js'

export default function StatCard({ icon: Icon, label, value, hint, accent = 'brand', className }) {
  const accents = {
    brand: 'text-brand-500 bg-brand-500/10',
    sky: 'text-sky-500 bg-sky-500/10',
    amber: 'text-amber-500 bg-amber-500/10',
    rose: 'text-rose-500 bg-rose-500/10',
    violet: 'text-violet-500 bg-violet-500/10'
  }
  return (
    <div className={cn('stat-card', className)}>
      <div className="flex items-start justify-between">
        <div>
          <div className="text-xs font-medium uppercase tracking-wide text-ink-500">
            {label}
          </div>
          <div className="mt-2 font-display text-3xl font-semibold tracking-tight">
            {typeof value === 'number' ? <CountUp value={value} /> : value}
          </div>
          {hint && (
            <div className="mt-1 text-xs text-ink-500">{hint}</div>
          )}
        </div>
        {Icon && (
          <div className={cn('flex h-9 w-9 items-center justify-center rounded-lg', accents[accent])}>
            <Icon className="h-4 w-4" />
          </div>
        )}
      </div>
    </div>
  )
}
