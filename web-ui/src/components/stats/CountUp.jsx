import { useEffect, useRef, useState } from 'react'

/**
 * Animates a number from 0 to `value` over `duration` ms on mount and on value
 * change. Lightweight, no dependencies.
 */
export default function CountUp({ value, duration = 800, className }) {
  const [display, setDisplay] = useState(0)
  const startRef = useRef(null)
  const fromRef = useRef(0)
  const toRef = useRef(Number(value) || 0)
  const frameRef = useRef(null)

  useEffect(() => {
    fromRef.current = display
    toRef.current = Number(value) || 0
    startRef.current = null
    cancelAnimationFrame(frameRef.current)

    const step = (ts) => {
      if (startRef.current == null) startRef.current = ts
      const progress = Math.min(1, (ts - startRef.current) / duration)
      // ease-out cubic
      const eased = 1 - Math.pow(1 - progress, 3)
      const current = fromRef.current + (toRef.current - fromRef.current) * eased
      setDisplay(current)
      if (progress < 1) frameRef.current = requestAnimationFrame(step)
    }
    frameRef.current = requestAnimationFrame(step)
    return () => cancelAnimationFrame(frameRef.current)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value, duration])

  const formatted =
    Number.isInteger(toRef.current)
      ? Math.round(display).toLocaleString()
      : display.toFixed(1).toLocaleString()

  return <span className={className}>{formatted}</span>
}
