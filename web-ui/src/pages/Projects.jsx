import { useState } from 'react'
import { Link } from 'react-router-dom'
import { FolderTree, Plus, Database, Trash2, ArrowUpRight, Loader2 } from 'lucide-react'
import { Card } from '@/components/ui/Card.jsx'
import Button from '@/components/ui/Button.jsx'
import Dialog from '@/components/ui/Dialog.jsx'
import Input from '@/components/ui/Input.jsx'
import { useProjectsStore } from '@/stores/useProjectsStore.js'
import { useToast } from '@/components/ui/Toast.jsx'
import { formatRelative, slugify } from '@/utils/format.js'
import { cn } from '@/utils/cn.js'
import { useApi } from '@/api/client.js'

export default function Projects() {
  const api = useApi()
  const knownProjects = useProjectsStore((s) => s.knownProjects)
  const knownTables = useProjectsStore((s) => s.knownTables)
  const addProject = useProjectsStore((s) => s.addProject)
  const removeProject = useProjectsStore((s) => s.removeProject)

  const toast = useToast()
  const [createOpen, setCreateOpen] = useState(false)
  const [newId, setNewId] = useState('')
  const [pendingDelete, setPendingDelete] = useState(null)
  const [deleting, setDeleting] = useState(false)

  function handleCreate(e) {
    e?.preventDefault?.()
    const id = slugify(newId)
    if (!id) {
      toast.warning('Please enter a valid project name')
      return
    }
    if (knownProjects.find((p) => p.id === id)) {
      toast.warning('Project already exists')
      return
    }
    addProject(id)
    toast.success(`Project "${id}" added`)
    setNewId('')
    setCreateOpen(false)
  }

  async function handleDelete() {
    if (!pendingDelete) return
    setDeleting(true)
    try {
      await api.delete(`/__api/projects/${pendingDelete}`)
      removeProject(pendingDelete)
      toast.success(`Project "${pendingDelete}" and all backend data deleted`)
      setPendingDelete(null)
    } catch (e) {
      const msg = e.normalizedMessage || e.message || 'Unknown error'
      toast.error(`Failed to delete project data: ${msg}`)
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="font-display text-2xl font-semibold tracking-tight">Projects</h1>
          <p className="mt-1 text-sm text-ink-500">
            Sandboxes are isolated by{' '}
            <code className="font-mono text-ink-700 dark:text-ink-200">projectId</code>{' '}
            in the URL path.
          </p>
        </div>
        <Button variant="primary" onClick={() => setCreateOpen(true)}>
          <Plus className="h-4 w-4" />
          New project
        </Button>
      </div>

      {knownProjects.length === 0 ? (
        <Card className="p-12 text-center">
          <FolderTree className="mx-auto mb-3 h-10 w-10 text-ink-300 dark:text-ink-700" />
          <div className="font-display text-base font-semibold">No projects yet</div>
          <div className="mt-1 text-sm text-ink-500">
            Add a project to start tracking its tables and documents.
          </div>
          <Button
            className="mt-4"
            variant="primary"
            onClick={() => setCreateOpen(true)}
          >
            <Plus className="h-4 w-4" />
            Add your first project
          </Button>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {knownProjects.map((p) => {
            const tables = knownTables[p.id] || []
            const docCount = tables.reduce((acc, t) => acc + (t.docCount || 0), 0)
            return (
              <Card
                key={p.id}
                className="group relative overflow-hidden transition hover:-translate-y-0.5 hover:border-brand-500/60 hover:shadow-glow"
              >
                <div className="pointer-events-none absolute -right-12 -top-12 h-32 w-32 rounded-full bg-brand-500/5 blur-2xl transition group-hover:bg-brand-500/15" />
                <div className="p-5">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-500/10 text-brand-500">
                        <FolderTree className="h-5 w-5" />
                      </div>
                      <div>
                        <div className="font-display text-lg font-semibold tracking-tight">
                          {p.id}
                        </div>
                        <div className="text-xs text-ink-500">
                          created {formatRelative(p.createdAt)}
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={(e) => {
                        e.preventDefault()
                        setPendingDelete(p.id)
                      }}
                      className="rounded p-1.5 text-ink-400 transition hover:bg-rose-500/10 hover:text-rose-500"
                      aria-label="Remove project"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                  <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
                    <div className="rounded-md border border-ink-200 px-3 py-2 dark:border-ink-800">
                      <div className="text-xs uppercase tracking-wide text-ink-500">
                        Tables
                      </div>
                      <div className="mt-0.5 font-display text-lg font-semibold">
                        {tables.length}
                      </div>
                    </div>
                    <div className="rounded-md border border-ink-200 px-3 py-2 dark:border-ink-800">
                      <div className="text-xs uppercase tracking-wide text-ink-500">
                        Documents
                      </div>
                      <div className="mt-0.5 font-display text-lg font-semibold">
                        {docCount}
                      </div>
                    </div>
                  </div>
                  <div className="mt-4 flex items-center justify-between text-xs text-ink-500">
                    <span>last activity {formatRelative(p.lastAccessedAt)}</span>
                    <Link
                      to={`/projects/${p.id}`}
                      className="inline-flex items-center gap-1 text-brand-500 transition hover:gap-2"
                    >
                      Open <ArrowUpRight className="h-3.5 w-3.5" />
                    </Link>
                  </div>
                </div>
              </Card>
            )
          })}
        </div>
      )}

      <div className="panel-muted p-4 text-xs text-ink-500">
        <Database className="mr-1 inline h-3.5 w-3.5" />
        Projects tracked here are synced with the backend. Deleting a project will also
        remove all its tables, documents, and API keys from the backend.{' '}
        <code className="font-mono">/sandbox/{'{projectId}'}/{'{table}'}</code>.
      </div>

      <Dialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title="Add project"
        description="Pick a URL-safe identifier for this sandbox."
        onConfirm={handleCreate}
        confirmLabel="Add project"
      >
        <form onSubmit={handleCreate} className="space-y-3">
          <label className="block text-sm font-medium">Project ID</label>
          <Input
            autoFocus
            placeholder="e.g. acme-frontend"
            value={newId}
            onChange={(e) => setNewId(e.target.value)}
          />
          <p className="text-xs text-ink-500">
            Letters, numbers, hyphens. Becomes{' '}
            <code className="font-mono">/sandbox/&lt;id&gt;/...</code>
          </p>
        </form>
      </Dialog>

      <Dialog
        open={Boolean(pendingDelete)}
        onClose={() => {
          if (!deleting) setPendingDelete(null)
        }}
        title="Delete project?"
        description="All backend data (tables, documents, API keys) will be permanently deleted. This cannot be undone."
        onConfirm={handleDelete}
        confirmLabel={deleting ? 'Deleting…' : 'Delete everything'}
        destructive
        disabled={deleting}
      >
        <div className="space-y-3">
          <p className="text-sm text-ink-600 dark:text-ink-300">
            Continue deleting{' '}
            <code className="rounded bg-ink-100 px-1.5 py-0.5 font-mono text-xs dark:bg-ink-800">
              {pendingDelete}
            </code>
            {' '}and all its data?
          </p>
          {deleting && (
            <div className="flex items-center gap-2 text-sm text-ink-500">
              <Loader2 className="h-4 w-4 animate-spin" />
              Deleting backend data…
            </div>
          )}
        </div>
      </Dialog>
    </div>
  )
}
