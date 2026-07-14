import { create } from 'zustand'
import { storage } from '@/utils/storage.js'
import { useEffect } from 'react'

const STORAGE_KEY = 'mocksbx.settings.v1'

// Default `baseUrl` is empty so axios sends requests to the page origin.
// During dev, Vite proxies `/sandbox` and `/hello` to the local Go backend.
// Users targeting a remote backend should set an absolute URL in Settings.
const defaultState = {
  baseUrl: '',
  theme: 'dark',
  density: 'comfortable',
  defaultProject: 'demo',
  defaultTable: 'users'
}

function loadInitial() {
  const persisted = storage.get(STORAGE_KEY, {})
  // Heal old defaults saved by earlier builds that pointed directly at
  // localhost:8080 (which bypasses the Vite proxy and trips CORS).
  if (
    typeof persisted.baseUrl === 'string' &&
    persisted.baseUrl === 'http://localhost:8080'
  ) {
    persisted.baseUrl = ''
  }
  return { ...defaultState, ...persisted }
}

export const useSettingsStore = create((set, get) => ({
  ...loadInitial(),
  setBaseUrl: (baseUrl) => {
    set({ baseUrl })
    storage.set(STORAGE_KEY, { ...get() })
  },
  setTheme: (theme) => {
    set({ theme })
    storage.set(STORAGE_KEY, { ...get() })
  },
  setDensity: (density) => {
    set({ density })
    storage.set(STORAGE_KEY, { ...get() })
  },
  setDefaultProject: (defaultProject) => {
    set({ defaultProject })
    storage.set(STORAGE_KEY, { ...get() })
  },
  setDefaultTable: (defaultTable) => {
    set({ defaultTable })
    storage.set(STORAGE_KEY, { ...get() })
  },
  exportConfig: () => {
    const snapshot = { ...get() }
    delete snapshot.exportConfig
    delete snapshot.importConfig
    return JSON.stringify(snapshot, null, 2)
  },
  importConfig: (raw) => {
    try {
      const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
      const next = { ...defaultState, ...parsed }
      set(next)
      storage.set(STORAGE_KEY, next)
      return { ok: true }
    } catch (e) {
      return { ok: false, error: e.message }
    }
  },
  reset: () => {
    set(defaultState)
    storage.set(STORAGE_KEY, defaultState)
  }
}))

/**
 * Apply theme to <html> root. Light = remove .dark, Dark = add .dark.
 */
export function useThemeSync() {
  const theme = useSettingsStore((s) => s.theme)
  useEffect(() => {
    const root = document.documentElement
    if (theme === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
  }, [theme])
}
