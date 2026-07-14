import { useEffect } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/utils/cn.js'

export default function Drawer({ open, onClose, title, description, children, side = 'right', width = '480px' }) {
  useEffect(() => {
    if (!open) return
    const onKey = (e) => e.key === 'Escape' && onClose?.()
    document.addEventListener('keydown', onKey)
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = ''
    }
  }, [open, onClose])

  return (
    <div
      className={cn(
        'fixed inset-0 z-40 transition',
        open ? 'pointer-events-auto opacity-100' : 'pointer-events-none opacity-0'
      )}
      aria-hidden={!open}
    >
      <div
        className="absolute inset-0 bg-ink-950/40 backdrop-blur-sm"
        onClick={onClose}
      />
      <div
        className={cn(
          'absolute top-0 h-full max-w-full bg-white shadow-2xl transition-transform duration-300 ease-out dark:bg-ink-900',
          side === 'right' ? 'right-0 border-l border-ink-200 dark:border-ink-800' : 'left-0 border-r border-ink-200 dark:border-ink-800',
          open
            ? 'translate-x-0'
            : side === 'right'
            ? 'translate-x-full'
            : '-translate-x-full'
        )}
        style={{ width }}
        role="dialog"
        aria-modal="true"
      >
        <div className="flex h-full flex-col">
          <div className="flex items-start justify-between border-b border-ink-200 px-5 py-4 dark:border-ink-800">
            <div>
              <h2 className="font-display text-lg font-semibold tracking-tight">{title}</h2>
              {description && (
                <p className="mt-1 text-xs text-ink-500">{description}</p>
              )}
            </div>
            <button
              onClick={onClose}
              className="rounded p-1.5 text-ink-500 transition hover:bg-ink-100 dark:hover:bg-ink-800"
              aria-label="Close"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
          <div className="flex-1 overflow-auto p-5">{children}</div>
        </div>
      </div>
    </div>
  )
}
