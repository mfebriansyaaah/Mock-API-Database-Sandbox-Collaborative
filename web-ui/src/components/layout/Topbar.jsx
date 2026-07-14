import { useEffect, useState } from 'react'
import { useApi } from '@/api/client.js'
import { useSettingsStore } from '@/stores/useSettingsStore.js'
import { Activity, Moon, Sun, Wifi, WifiOff } from 'lucide-react'

export default function Topbar() {
  const api = useApi()
  const baseUrl = useSettingsStore((s) => s.baseUrl)
  const theme = useSettingsStore((s) => s.theme)
  const setTheme = useSettingsStore((s) => s.setTheme)
  const [status, setStatus] = useState('checking')

  useEffect(() => {
    let alive = true
    setStatus('checking')
    api
      .get('/hello', { timeout: 4000, responseType: 'text', transformResponse: [(d) => d] })
      .then(() => alive && setStatus('online'))
      .catch(() => alive && setStatus('offline'))
    return () => {
      alive = false
    }
  }, [api, baseUrl])

  function toggleTheme() {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }

  return (
    <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-ink-200 bg-white/80 px-4 backdrop-blur md:px-6 dark:border-ink-800 dark:bg-ink-950/80">
      <div className="flex items-center gap-2 md:hidden">
        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-brand-500/10 text-brand-500">
          <Activity className="h-3.5 w-3.5" />
        </div>
        <span className="font-display text-sm font-semibold">Mock Sandbox</span>
      </div>
      <div className="hidden md:block" />

      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2 rounded-md border border-ink-200 bg-white px-2.5 py-1.5 text-xs dark:border-ink-700 dark:bg-ink-900">
          {status === 'online' && (
            <>
              <Wifi className="h-3.5 w-3.5 text-brand-500" />
              <span className="font-mono text-ink-700 dark:text-ink-200">online</span>
            </>
          )}
          {status === 'offline' && (
            <>
              <WifiOff className="h-3.5 w-3.5 text-rose-500" />
              <span className="font-mono text-rose-500">offline</span>
            </>
          )}
          {status === 'checking' && (
            <>
              <span className="h-2 w-2 animate-pulse-dot rounded-full bg-amber-500" />
              <span className="font-mono text-ink-500">checking…</span>
            </>
          )}
          <span className="font-mono text-ink-400">
            @ {baseUrl ? new URL(baseUrl).host : 'vite proxy'}
          </span>
        </div>
        <button
          onClick={toggleTheme}
          aria-label="Toggle theme"
          className="flex h-9 w-9 items-center justify-center rounded-md border border-ink-200 bg-white text-ink-600 transition hover:border-brand-500 hover:text-brand-600 dark:border-ink-700 dark:bg-ink-900 dark:text-ink-300"
        >
          {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
        </button>
      </div>
    </header>
  )
}
