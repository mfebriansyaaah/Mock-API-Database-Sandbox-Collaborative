/**
 * Thin wrapper around localStorage with JSON serialization and error handling.
 */
export const storage = {
  get(key, fallback = null) {
    try {
      const raw = window.localStorage.getItem(key)
      if (raw == null) return fallback
      return JSON.parse(raw)
    } catch (e) {
      console.warn('storage.get failed', e)
      return fallback
    }
  },
  set(key, value) {
    try {
      window.localStorage.setItem(key, JSON.stringify(value))
      return true
    } catch (e) {
      console.warn('storage.set failed', e)
      return false
    }
  },
  remove(key) {
    try {
      window.localStorage.removeItem(key)
    } catch (e) {
      console.warn('storage.remove failed', e)
    }
  }
}
