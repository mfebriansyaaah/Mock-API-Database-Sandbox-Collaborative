import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Database, Plus, FileText, Trash2, ArrowUpRight, Link as LinkIcon } from 'lucide-react'
import Card, { CardHeader, CardTitle, CardBody } from '@/components/ui/Card.jsx'
import Button from '@/components/ui/Button.jsx'
import Input from '@/components/ui/Input.jsx'
import Dialog from '@/components/ui/Dialog.jsx'
import Skeleton from '@/components/ui/Skeleton.jsx'
import SearchInput from '@/components/data/SearchInput.jsx'
import EndpointLink from '@/components/ui/EndpointLink.jsx'
import { useToast } from '@/components/ui/Toast.jsx'
import { useApi } from '@/api/client.js'
import { useProjectsStore } from '@/stores/useProjectsStore.js'
import { formatRelative } from '@/utils/format.js'
import { slugify } from '@/utils/format.js'

const EMPTY_TABLES = []

export default function Tables() {
  const { projectId } = useParams()
  const navigate = useNavigate()
  const api = useApi()
  const toast = useToast()

  const knownTables = useProjectsStore((s) => s.knownTables[projectId] ?? EMPTY_TABLES)
  const addTable = useProjectsStore((s) => s.addTable)
  const touchProject = useProjectsStore((s) => s.touchProject)
  const setTableDocCount = useProjectsStore((s) => s.setTableDocCount)
  const removeTable = useProjectsStore((s) => s.removeTable)

  const [counts, setCounts] = useState({})
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [endpointTable, setEndpointTable] = useState(null)

  useEffect(() => {
    touchProject(projectId)
  }, [projectId, touchProject])

  useEffect(() => {
    let alive = true
    async function refresh() {
      setLoading(true)
      const next = {}
      await Promise.all(
        knownTables.map(async (t) => {
          try {
            const r = await api.get(`/sandbox/${projectId}/${t.name}/_count`)
            next[t.name] = typeof r.data?.count === 'number' ? r.data.count : 0
            setTableDocCount(projectId, t.name, next[t.name])
          } catch (e) {
            next[t.name] = 0
          }
        })
      )
      if (alive) {
        setCounts(next)
        setLoading(false)
      }
    }
    if (knownTables.length > 0) refresh()
    else setLoading(false)
    return () => {
      alive = false
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId, knownTables.length])

  function handleCreate(e) {
    e?.preventDefault?.()
    const name = slugify(newName)
    if (!name) {
      toast.warning('Enter a valid table name')
      return
    }
    addTable(projectId, name)
    toast.success(`Table "${name}" tracked`)
    setNewName('')
    setCreateOpen(false)
  }

  const filtered = knownTables.filter((t) => t.name.includes(search.toLowerCase()))

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <Link
            to="/projects"
            className="inline-flex items-center gap-1 text-xs text-ink-500 transition hover:text-ink-900 dark:hover:text-ink-100"
          >
            <ArrowLeft className="h-3.5 w-3.5" /> All projects
          </Link>
          <h1 className="mt-1 font-display text-2xl font-semibold tracking-tight">
            {projectId}
          </h1>
          <p className="mt-1 text-sm text-ink-500">
            Tables in this project are isolated paths under{' '}
            <code className="font-mono">/sandbox/{projectId}/...</code>
          </p>
        </div>
        <Button variant="primary" onClick={() => setCreateOpen(true)}>
          <Plus className="h-4 w-4" />
          Track new table
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Tables</CardTitle>
          <div className="w-64">
            <SearchInput
              value={search}
              onChange={setSearch}
              placeholder="Filter tables…"
            />
          </div>
        </CardHeader>
        <CardBody className="p-0">
          {loading ? (
            <div className="space-y-2 p-5">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : filtered.length === 0 ? (
            <div className="px-5 py-12 text-center text-sm text-ink-500">
              <Database className="mx-auto mb-2 h-8 w-8 text-ink-300 dark:text-ink-700" />
              {knownTables.length === 0
                ? 'No tables tracked yet. Add one to start exploring data.'
                : 'No tables match your filter.'}
            </div>
          ) : (
            <table className="table-base">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Documents</th>
                  <th>Last accessed</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((t) => (
                  <tr key={t.name}>
                    <td>
                      <Link
                        to={`/projects/${projectId}/${t.name}`}
                        className="inline-flex items-center gap-2 font-medium text-ink-900 hover:text-brand-500 dark:text-ink-100"
                      >
                        <FileText className="h-4 w-4 text-brand-500" />
                        {t.name}
                      </Link>
                    </td>
                    <td className="font-mono text-xs">
                      {counts[t.name] ?? t.docCount ?? '—'}
                    </td>
                    <td className="text-xs text-ink-500">
                      {formatRelative(t.lastAccessedAt)}
                    </td>
                    <td className="text-right">
                      <div className="inline-flex items-center gap-1">
                        <button
                          onClick={() => setEndpointTable(t.name)}
                          className="rounded p-1.5 text-ink-400 transition hover:bg-brand-500/10 hover:text-brand-500"
                          title="Get endpoint link"
                          aria-label="Endpoint link"
                        >
                          <LinkIcon className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => removeTable(projectId, t.name)}
                          className="rounded p-1.5 text-ink-400 transition hover:bg-rose-500/10 hover:text-rose-500"
                          aria-label="Untrack"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                        <Link
                          to={`/projects/${projectId}/${t.name}`}
                          className="inline-flex items-center gap-1 rounded-md border border-ink-200 px-2 py-1 text-xs font-medium text-ink-600 transition hover:border-brand-500 hover:text-brand-500 dark:border-ink-700 dark:text-ink-300"
                        >
                          Open <ArrowUpRight className="h-3 w-3" />
                        </Link>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </CardBody>
      </Card>

      <div className="panel-muted p-4 text-xs text-ink-500">
        Tip: The backend creates a table on the first write. Tracking it here just
        adds it to your sidebar shortcuts. To populate, open the tester or the
        documents page and POST your first record.
      </div>

      <Dialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title="Track a table"
        description="Add this table to your local tracking."
        onConfirm={handleCreate}
        confirmLabel="Add"
      >
        <form onSubmit={handleCreate} className="space-y-3">
          <label className="block text-sm font-medium">Table name</label>
          <Input
            autoFocus
            placeholder="e.g. users, products, orders"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
          />
          <p className="text-xs text-ink-500">
            Path:{' '}
            <code className="font-mono">/sandbox/{projectId}/&lt;name&gt;</code>
          </p>
        </form>
      </Dialog>

      <EndpointLink
        open={Boolean(endpointTable)}
        onClose={() => setEndpointTable(null)}
        projectId={projectId}
        table={endpointTable || ''}
      />
    </div>
  )
}
