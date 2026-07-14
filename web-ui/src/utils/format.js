/**
 * Lightweight formatting helpers used across the UI.
 */

export function formatDate(value) {
  if (!value) return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return String(value)
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

export function formatRelative(value) {
  if (!value) return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return String(value)
  const diff = Date.now() - d.getTime()
  const sec = Math.round(diff / 1000)
  if (sec < 60) return `${sec}s ago`
  const min = Math.round(sec / 60)
  if (min < 60) return `${min}m ago`
  const hr = Math.round(min / 60)
  if (hr < 24) return `${hr}h ago`
  const day = Math.round(hr / 24)
  if (day < 30) return `${day}d ago`
  return formatDate(d)
}

export function formatLatency(value) {
  if (value == null) return '—'
  if (typeof value === 'string') {
    // Backend may send duration in ISO-8601 form or as nanoseconds.
    if (value.endsWith('ns')) return `${(Number(value.slice(0, -2)) / 1e6).toFixed(1)} ms`
    if (value.endsWith('µs')) return `${(Number(value.slice(0, -2)) / 1e3).toFixed(1)} ms`
    if (value.endsWith('ms')) return `${Number(value.slice(0, -2)).toFixed(1)} ms`
    if (value.endsWith('s')) return `${(Number(value.slice(0, -1)) * 1000).toFixed(1)} ms`
  }
  const ms = Number(value)
  if (Number.isNaN(ms)) return String(value)
  if (ms < 1) return `${(ms * 1000).toFixed(0)} µs`
  if (ms < 1000) return `${ms.toFixed(1)} ms`
  return `${(ms / 1000).toFixed(2)} s`
}

export function formatBytes(value) {
  if (!value) return '—'
  const str = typeof value === 'string' ? value : JSON.stringify(value)
  const bytes = new Blob([str]).size
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}

export function truncate(str, max = 64) {
  if (str == null) return ''
  const s = String(str)
  return s.length > max ? s.slice(0, max - 1) + '…' : s
}

export function slugify(value) {
  return String(value)
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9-_]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 48)
}
