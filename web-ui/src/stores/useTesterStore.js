import { create } from 'zustand'
import { storage } from '@/utils/storage.js'

const STORAGE_KEY = 'mocksbx.tester.v1'
const MAX_HISTORY = 10

function loadInitial() {
  return storage.get(STORAGE_KEY, { history: [] })
}

export const useTesterStore = create((set, get) => ({
  ...loadInitial(),
  addHistory: (entry) => {
    const next = [entry, ...get().history].slice(0, MAX_HISTORY)
    set({ history: next })
    storage.set(STORAGE_KEY, { ...get() })
  },
  clearHistory: () => {
    set({ history: [] })
    storage.set(STORAGE_KEY, { ...get() })
  }
}))
