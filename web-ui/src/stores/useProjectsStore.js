import { create } from 'zustand'
import { storage } from '@/utils/storage.js'

const STORAGE_KEY = 'mocksbx.projects.v1'

function loadInitial() {
  return storage.get(STORAGE_KEY, {
    knownProjects: [
      { id: 'demo', createdAt: Date.now() - 86400000, lastAccessedAt: Date.now() - 600000 }
    ],
    knownTables: {
      // [projectId]: [{ name, lastAccessedAt, docCount }]
      demo: [
        { name: 'users', lastAccessedAt: Date.now() - 300000, docCount: 0 },
        { name: 'products', lastAccessedAt: Date.now() - 1200000, docCount: 0 }
      ]
    }
  })
}

export const useProjectsStore = create((set, get) => ({
  ...loadInitial(),
  addProject: (id) => {
    const list = get().knownProjects
    if (list.find((p) => p.id === id)) return
    const next = [
      ...list,
      { id, createdAt: Date.now(), lastAccessedAt: Date.now() }
    ]
    set({ knownProjects: next })
    storage.set(STORAGE_KEY, { ...get() })
  },
  removeProject: (id) => {
    const next = get().knownProjects.filter((p) => p.id !== id)
    const tables = { ...get().knownTables }
    delete tables[id]
    set({ knownProjects: next, knownTables: tables })
    storage.set(STORAGE_KEY, { ...get() })
  },
  touchProject: (id) => {
    const next = get().knownProjects.map((p) =>
      p.id === id ? { ...p, lastAccessedAt: Date.now() } : p
    )
    set({ knownProjects: next })
    storage.set(STORAGE_KEY, { ...get() })
  },
  addTable: (projectId, name) => {
    const tables = { ...get().knownTables }
    const list = tables[projectId] || []
    if (list.find((t) => t.name === name)) {
      // Touch
      tables[projectId] = list.map((t) =>
        t.name === name ? { ...t, lastAccessedAt: Date.now() } : t
      )
    } else {
      tables[projectId] = [
        ...list,
        { name, lastAccessedAt: Date.now(), docCount: 0 }
      ]
    }
    set({ knownTables: tables })
    storage.set(STORAGE_KEY, { ...get() })
  },
  setTableDocCount: (projectId, name, docCount) => {
    const tables = { ...get().knownTables }
    const list = tables[projectId] || []
    tables[projectId] = list.map((t) =>
      t.name === name ? { ...t, docCount, lastAccessedAt: Date.now() } : t
    )
    set({ knownTables: tables })
    storage.set(STORAGE_KEY, { ...get() })
  },
  removeTable: (projectId, name) => {
    const tables = { ...get().knownTables }
    tables[projectId] = (tables[projectId] || []).filter((t) => t.name !== name)
    set({ knownTables: tables })
    storage.set(STORAGE_KEY, { ...get() })
  }
}))
