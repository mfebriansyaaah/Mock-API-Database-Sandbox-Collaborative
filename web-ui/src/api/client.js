import { useMemo } from 'react'
import axios from 'axios'
import { useSettingsStore } from '@/stores/useSettingsStore.js'

/**
 * Returns an axios instance using the current base URL from the settings store.
 * The instance is memoized and only recreated when the base URL changes.
 */
export function createApiClient(baseUrl) {
  const client = axios.create({
    baseURL: baseUrl || undefined,
    timeout: 15000,
    headers: { 'Content-Type': 'application/json' }
  })

  client.interceptors.response.use(
    (response) => response,
    (error) => {
      const status = error.response?.status
      const message =
        error.response?.data?.message ||
        error.response?.data ||
        error.message ||
        'Unknown error'
      // Normalize error message (backend may return plain text)
      const normalized =
        typeof message === 'string' ? message : JSON.stringify(message)
      error.normalizedMessage = `HTTP ${status || 'ERR'} · ${normalized}`
      return Promise.reject(error)
    }
  )

  return client
}

/**
 * Hook helper: get the latest client. The instance is memoized so we don't
 * re-create it on every render (which would cause StrictMode double-invokes
 * to fire two requests back-to-back).
 */
export function useApi() {
  const baseUrl = useSettingsStore((s) => s.baseUrl)
  return useMemo(() => createApiClient(baseUrl), [baseUrl])
}
