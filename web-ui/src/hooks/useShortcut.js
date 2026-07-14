import { useEffect, useState } from 'react'

/**
 * Listen to keyboard shortcut and trigger callback. Ignores when focus is in
 * inputs/textareas so that typing still works.
 */
export function useShortcut(combo, callback) {
  useEffect(() => {
    function handler(e) {
      const target = e.target
      if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)) {
        return
      }
      const keys = combo.toLowerCase().split('+')
      const want = {
        ctrl: keys.includes('ctrl') || keys.includes('cmd'),
        shift: keys.includes('shift'),
        alt: keys.includes('alt'),
        key: keys[keys.length - 1]
      }
      if (want.ctrl !== (e.ctrlKey || e.metaKey)) return
      if (want.shift !== e.shiftKey) return
      if (want.alt !== e.altKey) return
      if (e.key.toLowerCase() !== want.key) return
      e.preventDefault()
      callback(e)
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [combo, callback])
}
