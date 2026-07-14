import { createContext, useCallback, useContext, useMemo, useState } from 'react'
import { CheckCircle2, AlertTriangle, XCircle, Info, X } from 'lucide-react'
import { cn } from '@/utils/cn.js'

const ToastContext = createContext(null)

let nextId = 1

const ICONS = {
  success: CheckCircle2,
  error: XCircle,
  warning: AlertTriangle,
  info: Info
}

const TONE = {
  success: 'border-brand-500/40 text-brand-600 dark:text-brand-400',
  error: 'border-rose-500/40 text-rose-500',
  warning: 'border-amber-500/40 text-amber-500',
  info: 'border-sky-500/40 text-sky-500'
}

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])

  const dismiss = useCallback((id) => {
    setToasts((list) => list.filter((t) => t.id !== id))
  }, [])

  const push = useCallback(
    (toast) => {
      const id = nextId++
      const t = { id, tone: 'info', duration: 4000, ...toast }
      setToasts((list) => [...list, t])
      if (t.duration > 0) {
        setTimeout(() => dismiss(id), t.duration)
      }
      return id
    },
    [dismiss]
  )

  const api = useMemo(
    () => ({
      push,
      success: (message, opts) => push({ tone: 'success', message, ...opts }),
      error: (message, opts) => push({ tone: 'error', message, duration: 6000, ...opts }),
      warning: (message, opts) => push({ tone: 'warning', message, ...opts }),
      info: (message, opts) => push({ tone: 'info', message, ...opts }),
      dismiss
    }),
    [push, dismiss]
  )

  return (
    <ToastContext.Provider value={api}>
      {children}
      <div className="pointer-events-none fixed bottom-6 right-6 z-50 flex w-full max-w-sm flex-col gap-2">
        {toasts.map((t) => {
          const Icon = ICONS[t.tone] || Info
          return (
            <div
              key={t.id}
              className={cn(
                'pointer-events-auto flex items-start gap-3 rounded-lg border bg-white/95 px-3 py-2.5 shadow-lg backdrop-blur transition animate-fade-in-up dark:bg-ink-900/95',
                TONE[t.tone]
              )}
            >
              <Icon className="mt-0.5 h-4 w-4 shrink-0" />
              <div className="flex-1 text-sm text-ink-800 dark:text-ink-100">
                {t.title && <div className="font-semibold">{t.title}</div>}
                <div className="leading-relaxed">{t.message}</div>
              </div>
              <button
                onClick={() => dismiss(t.id)}
                className="rounded p-1 text-ink-400 transition hover:bg-ink-100 hover:text-ink-700 dark:hover:bg-ink-800 dark:hover:text-ink-200"
                aria-label="Dismiss"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </div>
          )
        })}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) {
    // Safe no-op outside provider (e.g. tests)
    return { push: () => {}, success: () => {}, error: () => {}, info: () => {}, warning: () => {}, dismiss: () => {} }
  }
  return ctx
}
