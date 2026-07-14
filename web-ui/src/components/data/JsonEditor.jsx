import { useEffect, useRef, useState } from 'react'
import { cn } from '@/utils/cn.js'

/**
 * A textarea-based JSON editor with simple line-number gutter and validation
 * feedback. Designed to stay light and dependency-free.
 */
export default function JsonEditor({ value, onChange, rows = 18, className }) {
  const [error, setError] = useState(null)
  const gutterRef = useRef(null)
  const areaRef = useRef(null)

  useEffect(() => {
    if (!value) {
      setError(null)
      return
    }
    try {
      JSON.parse(value)
      setError(null)
    } catch (e) {
      setError(e.message)
    }
  }, [value])

  function handleScroll() {
    if (gutterRef.current && areaRef.current) {
      gutterRef.current.scrollTop = areaRef.current.scrollTop
    }
  }

  const lineCount = Math.max(1, (value || '').split('\n').length)

  return (
    <div className={cn('relative overflow-hidden rounded-md border border-ink-200 dark:border-ink-700', className)}>
      <div className="flex">
        <div
          ref={gutterRef}
          className="select-none overflow-hidden border-r border-ink-200 bg-ink-50 px-2 py-2 text-right font-mono text-xs leading-6 text-ink-400 dark:border-ink-800 dark:bg-ink-950"
          style={{ minWidth: '2.5rem' }}
          aria-hidden
        >
          {Array.from({ length: lineCount }, (_, i) => (
            <div key={i}>{i + 1}</div>
          ))}
        </div>
        <textarea
          ref={areaRef}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onScroll={handleScroll}
          rows={rows}
          spellCheck="false"
          className="block w-full resize-none border-0 bg-white px-3 py-2 font-mono text-xs leading-6 text-ink-900 focus:outline-none focus:ring-0 dark:bg-ink-900 dark:text-ink-100"
          style={{ minHeight: `${rows * 1.5}rem` }}
        />
      </div>
      <div className="flex items-center justify-between border-t border-ink-200 bg-ink-50/50 px-3 py-1.5 text-xs dark:border-ink-800 dark:bg-ink-900/60">
        <span className="font-mono text-ink-500">JSON</span>
        {error ? (
          <span className="font-mono text-rose-500">⚠ {error}</span>
        ) : (
          <span className="font-mono text-brand-600">✓ valid</span>
        )}
      </div>
    </div>
  )
}
