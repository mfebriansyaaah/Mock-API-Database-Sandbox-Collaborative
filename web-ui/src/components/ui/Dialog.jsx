import { useEffect } from 'react'
import { X } from 'lucide-react'
import Button from './Button.jsx'
import { cn } from '@/utils/cn.js'

export default function Dialog({
  open,
  onClose,
  title,
  description,
  children,
  size = 'md',
  onConfirm,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  destructive = false
}) {
  useEffect(() => {
    if (!open) return
    const onKey = (e) => e.key === 'Escape' && onClose?.()
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onClose])

  const sizes = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-2xl'
  }

  return (
    <div
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center px-4 transition',
        open ? 'pointer-events-auto opacity-100' : 'pointer-events-none opacity-0'
      )}
      aria-hidden={!open}
    >
      <div className="absolute inset-0 bg-ink-950/50 backdrop-blur-sm" onClick={onClose} />
      <div
        className={cn(
          'relative w-full overflow-hidden rounded-xl border border-ink-200 bg-white shadow-2xl dark:border-ink-800 dark:bg-ink-900',
          sizes[size]
        )}
        role="dialog"
        aria-modal="true"
      >
        <div className="flex items-start justify-between border-b border-ink-200 px-5 py-4 dark:border-ink-800">
          <div>
            <h2 className="font-display text-base font-semibold tracking-tight">{title}</h2>
            {description && (
              <p className="mt-1 text-xs text-ink-500">{description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            className="rounded p-1 text-ink-500 transition hover:bg-ink-100 dark:hover:bg-ink-800"
            aria-label="Close"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="px-5 py-4">{children}</div>
        {(onConfirm || cancelLabel) && (
          <div className="flex items-center justify-end gap-2 border-t border-ink-200 bg-ink-50/60 px-5 py-3 dark:border-ink-800 dark:bg-ink-900/60">
            <Button variant="ghost" onClick={onClose}>
              {cancelLabel}
            </Button>
            {onConfirm && (
              <Button variant={destructive ? 'danger' : 'primary'} onClick={onConfirm}>
                {confirmLabel}
              </Button>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
