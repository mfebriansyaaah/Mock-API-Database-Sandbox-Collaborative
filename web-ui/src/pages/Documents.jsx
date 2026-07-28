import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  Plus,
  Trash2,
  Edit3,
  Search,
  RefreshCcw,
  ChevronLeft,
  ChevronRight,
  FileJson,
  Link as LinkIcon
} from 'lucide-react'
import Card, { CardHeader, CardTitle, CardBody } from '@/components/ui/Card.jsx'
import Button from '@/components/ui/Button.jsx'
import Skeleton from '@/components/ui/Skeleton.jsx'
import SearchInput from '@/components/data/SearchInput.jsx'
import Drawer from '@/components/ui/Drawer.jsx'
import JsonEditor from '@/components/data/JsonEditor.jsx'
import Dialog from '@/components/ui/Dialog.jsx'
import EndpointLink from '@/components/ui/EndpointLink.jsx'
import { useApi } from '@/api/client.js'
import { useToast } from '@/components/ui/Toast.jsx'
import { useProjectsStore } from '@/stores/useProjectsStore.js'
import {
  getTableDocuments,
  createDocument,
  updateDocument,
  deleteDocument
} from '@/api/sandbox.js'
import { formatRelative, truncate } from '@/utils/format.js'
import { nanoid } from 'nanoid/non-secure'

const META_KEYS = new Set(['_createdAt', '_createdBy', '_updatedAt', '_updatedBy'])

function deriveId(doc) {
  return (
    doc.id ||
    doc.ID ||
    doc.Id ||
    doc._id ||
    doc.uuid ||
    doc.key ||
    Object.values(doc).find((v) => typeof v === 'string' && /^[0-9a-f-]{8,}/i.test(v))
  )
}

function pickValue(doc, key) {
  if (key in doc) return doc[key]
  if (key.toLowerCase() in doc) return doc[key.toLowerCase()]
  return undefined
}

export default function Documents() {
  const { projectId, table } = useParams()
  const api = useApi()
  const toast = useToast()
  const setTableDocCount = useProjectsStore((s) => s.setTableDocCount)

  // Server-side pagination state. Backend caps limit at 100.
  const [docs, setDocs] = useState([])
  const [pageInfo, setPageInfo] = useState({ offset: 0, limit: 25, count: 0, nextOffset: 0 })
  const [hasMore, setHasMore] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [errorCode, setErrorCode] = useState(null)
  const [search, setSearch] = useState('')
  const [pageSize, setPageSize] = useState(25)
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState(new Set())
  const [editorState, setEditorState] = useState({ open: false, mode: 'create', id: null, body: '{}' })
  const [pendingDelete, setPendingDelete] = useState(null)
  const [pendingBulkDelete, setPendingBulkDelete] = useState(false)
  const [endpointOpen, setEndpointOpen] = useState(false)

  const refresh = useCallback(async () => {
    setLoading(true)
    setError(null)
    setErrorCode(null)
    try {
      const offset = (page - 1) * pageSize
      const result = await getTableDocuments(api, projectId, table, {
        limit: pageSize,
        offset
      })
      setDocs(result.data || [])
      setPageInfo({
        offset: result.offset,
        limit: result.limit,
        count: result.count,
        nextOffset: result.nextOffset
      })
      setHasMore((result.data || []).length === result.limit)
      // We only know the size of the current page; let the store know the
      // lower bound (>= current page size) to avoid an extra HEAD call.
      setTableDocCount(projectId, table, Math.max(result.count, 0))
    } catch (e) {
      setError(e.normalizedMessage || e.message)
      setErrorCode(e.response?.status || null)
      setDocs([])
    } finally {
      setLoading(false)
    }
  }, [api, projectId, table, page, pageSize, setTableDocCount])

  useEffect(() => {
    refresh()
  }, [refresh])

  // Derive column union (exclude metadata & id) from the current page
  const columns = useMemo(() => {
    const set = new Set()
    docs.forEach((d) => {
      Object.keys(d || {}).forEach((k) => {
        if (!META_KEYS.has(k)) set.add(k)
      })
    })
    const sorted = [...set].sort((a, b) => {
      const ia = /id$|^id$|_id$|uuid$/i.test(a) ? -1 : 1
      const ib = /id$|^id$|_id$|uuid$/i.test(b) ? -1 : 1
      return ia - ib || a.localeCompare(b)
    })
    return sorted.slice(0, 6)
  }, [docs])

  // Client-side search filter on the current page (server lacks a search API)
  const filtered = useMemo(() => {
    if (!search) return docs
    const q = search.toLowerCase()
    return docs.filter((d) => {
      const id = deriveId(d)
      if (id && String(id).toLowerCase().includes(q)) return true
      return Object.values(d || {}).some((v) =>
        String(v).toLowerCase().includes(q)
      )
    })
  }, [docs, search])

  const pageItems = filtered

  function openCreate() {
    setEditorState({
      open: true,
      mode: 'create',
      id: null,
      body: JSON.stringify(
        {
          name: 'New record',
          createdAt: new Date().toISOString().slice(0, 10)
        },
        null,
        2
      )
    })
  }

  function openEdit(doc) {
    const id = deriveId(doc)
    const copy = { ...doc }
    META_KEYS.forEach((k) => delete copy[k])
    delete copy.id
    delete copy._id
    setEditorState({
      open: true,
      mode: 'edit',
      id,
      body: JSON.stringify(copy, null, 2)
    })
  }

  async function handleSave() {
    let parsed
    try {
      parsed = JSON.parse(editorState.body || '{}')
    } catch (e) {
      toast.error('Invalid JSON: ' + e.message)
      return
    }
    try {
      if (editorState.mode === 'create') {
        const id = nanoid(10)
        await createDocument(api, projectId, table, parsed, id)
        toast.success(`Document created (id: ${id})`)
      } else if (editorState.mode === 'edit') {
        await updateDocument(api, projectId, table, editorState.id, parsed)
        toast.success(`Document ${editorState.id} updated`)
      }
      setEditorState((s) => ({ ...s, open: false }))
      await refresh()
    } catch (e) {
      toast.error(e.normalizedMessage || e.message)
    }
  }

  async function handleDelete(doc) {
    const id = deriveId(doc)
    try {
      await deleteDocument(api, projectId, table, id)
      toast.success(`Deleted ${id}`)
      setSelected((s) => {
        const next = new Set(s)
        next.delete(id)
        return next
      })
      await refresh()
    } catch (e) {
      toast.error(e.normalizedMessage || e.message)
    } finally {
      setPendingDelete(null)
    }
  }

  async function handleBulkDelete() {
    const ids = [...selected]
    let ok = 0
    for (const id of ids) {
      try {
        await deleteDocument(api, projectId, table, id)
        ok += 1
      } catch (e) {
        // ignore
      }
    }
    toast.success(`Deleted ${ok}/${ids.length} documents`)
    setSelected(new Set())
    await refresh()
    setPendingBulkDelete(false)
  }

  function toggleAll() {
    if (selected.size === pageItems.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(pageItems.map((d) => deriveId(d)).filter(Boolean)))
    }
  }

  function toggleOne(id) {
    setSelected((s) => {
      const next = new Set(s)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <Link
            to={`/projects/${projectId}`}
            className="inline-flex items-center gap-1 text-xs text-ink-500 transition hover:text-ink-900 dark:hover:text-ink-100"
          >
            <ArrowLeft className="h-3.5 w-3.5" /> Back to {projectId}
          </Link>
          <h1 className="mt-1 font-display text-2xl font-semibold tracking-tight">
            <span className="text-ink-400">{projectId} /</span> {table}
          </h1>
          <p className="mt-1 text-sm text-ink-500">
            Manage documents in this sandbox table.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" onClick={refresh} title="Reload">
            <RefreshCcw className="h-4 w-4" />
          </Button>
          <Button variant="outline" onClick={() => setEndpointOpen(true)} title="Get endpoint link">
            <LinkIcon className="h-4 w-4" />
            Endpoint
          </Button>
          {selected.size > 0 && (
            <Button
              variant="danger"
              onClick={() => setPendingBulkDelete(true)}
            >
              <Trash2 className="h-4 w-4" />
              Delete {selected.size}
            </Button>
          )}
          <Button variant="primary" onClick={openCreate}>
            <Plus className="h-4 w-4" />
            New document
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <CardTitle>Documents</CardTitle>
            <span className="text-xs text-ink-500">
              {loading
                ? 'Loading…'
                : `showing ${docs.length === 0 ? 0 : pageInfo.offset + 1}–${pageInfo.offset + docs.length}`}
              {hasMore && <span className="ml-1 text-ink-400">(more available)</span>}
            </span>
          </div>
          <div className="flex w-full items-center gap-2 sm:w-auto">
            <div className="w-64">
              <SearchInput
                value={search}
                onChange={(v) => {
                  setSearch(v)
                }}
                placeholder="Search any field…"
              />
            </div>
            <select
              value={pageSize}
              onChange={(e) => {
                setPageSize(Number(e.target.value))
                setPage(1)
              }}
              className="h-9 rounded-md border border-ink-200 bg-white px-2 text-xs dark:border-ink-700 dark:bg-ink-900"
            >
              {[10, 25, 50, 100].map((n) => (
                <option key={n} value={n}>
                  {n} / page
                </option>
              ))}
            </select>
          </div>
        </CardHeader>
        <CardBody className="p-0">
          {loading ? (
            <div className="space-y-2 p-5">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : error ? (
            <div className="px-5 py-12 text-center text-sm">
              <div className="font-display text-base font-semibold text-rose-500">
                {errorCode === 503
                  ? 'Backend quota exceeded'
                  : errorCode === 404
                  ? 'Table not found'
                  : 'Failed to load documents'}
              </div>
              <div className="mx-auto mt-1 max-w-md text-ink-500">{error}</div>
              {errorCode === 503 && (
                <p className="mx-auto mt-2 max-w-md text-xs text-ink-500">
                  The Firestore backend is hitting its free-tier daily quota or a
                  query-size limit. Try reducing the page size above, or retry
                  in a few minutes.
                </p>
              )}
              <Button className="mt-3" variant="outline" onClick={refresh}>
                <RefreshCcw className="h-3.5 w-3.5" />
                Retry
              </Button>
            </div>
          ) : filtered.length === 0 ? (
            <div className="px-5 py-12 text-center text-sm text-ink-500">
              <FileJson className="mx-auto mb-2 h-8 w-8 text-ink-300 dark:text-ink-700" />
              {search
                ? 'No documents match your search.'
                : 'This table is empty. Create your first document.'}
              {!search && (
                <div className="mt-3">
                  <Button variant="primary" onClick={openCreate}>
                    <Plus className="h-4 w-4" />
                    New document
                  </Button>
                </div>
              )}
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="table-base">
                  <thead>
                    <tr>
                      <th className="w-10">
                        <input
                          type="checkbox"
                          checked={
                            selected.size === pageItems.length && pageItems.length > 0
                          }
                          onChange={toggleAll}
                          className="h-3.5 w-3.5 accent-brand-500"
                        />
                      </th>
                      <th>ID</th>
                      {columns.map((c) => (
                        <th key={c}>{c}</th>
                      ))}
                      <th>Updated</th>
                      <th className="text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {pageItems.map((doc) => {
                      const id = deriveId(doc) || '—'
                      return (
                        <tr key={String(id)}>
                          <td>
                            <input
                              type="checkbox"
                              checked={selected.has(id)}
                              onChange={() => toggleOne(id)}
                              className="h-3.5 w-3.5 accent-brand-500"
                            />
                          </td>
                          <td className="max-w-[14rem]">
                            <Link
                              to={`/projects/${projectId}/${table}/${id}`}
                              className="font-mono text-xs text-brand-600 hover:underline dark:text-brand-400"
                              title={String(id)}
                            >
                              {truncate(String(id), 22)}
                            </Link>
                          </td>
                          {columns.map((c) => {
                            const v = pickValue(doc, c)
                            return (
                              <td key={c} className="max-w-[18rem]">
                                <span
                                  className="block truncate font-mono text-xs"
                                  title={typeof v === 'object' ? JSON.stringify(v) : String(v ?? '')}
                                >
                                  {typeof v === 'object' && v !== null
                                    ? JSON.stringify(v)
                                    : String(v ?? '—')}
                                </span>
                              </td>
                            )
                          })}
                          <td className="text-xs text-ink-500">
                            {formatRelative(doc._updatedAt)}
                          </td>
                          <td className="text-right">
                            <div className="inline-flex items-center gap-1">
                              <button
                                onClick={() => openEdit(doc)}
                                className="rounded p-1.5 text-ink-400 transition hover:bg-brand-500/10 hover:text-brand-500"
                                aria-label="Edit"
                              >
                                <Edit3 className="h-3.5 w-3.5" />
                              </button>
                              <button
                                onClick={() => setPendingDelete(doc)}
                                className="rounded p-1.5 text-ink-400 transition hover:bg-rose-500/10 hover:text-rose-500"
                                aria-label="Delete"
                              >
                                <Trash2 className="h-3.5 w-3.5" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
              <div className="flex items-center justify-between border-t border-ink-200 px-4 py-2.5 text-xs text-ink-500 dark:border-ink-800">
                <div>
                  Page {page}
                  {hasMore ? '+' : ''}
                </div>
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1}
                  >
                    <ChevronLeft className="h-3.5 w-3.5" /> Prev
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setPage((p) => (hasMore ? p + 1 : p))}
                    disabled={!hasMore}
                    title={hasMore ? 'Load next page' : 'No more documents'}
                  >
                    Next <ChevronRight className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardBody>
      </Card>

      <Drawer
        open={editorState.open}
        onClose={() => setEditorState((s) => ({ ...s, open: false }))}
        title={editorState.mode === 'create' ? 'New document' : `Edit ${editorState.id}`}
        description={
          editorState.mode === 'create'
            ? 'A new ID will be auto-generated. Metadata fields are added by the server.'
            : 'Changes are persisted via PUT (MergeAll).'
        }
      >
        <div className="space-y-3">
          <JsonEditor
            value={editorState.body}
            onChange={(v) => setEditorState((s) => ({ ...s, body: v }))}
            rows={20}
          />
          <div className="flex items-center justify-end gap-2">
            <Button
              variant="ghost"
              onClick={() => setEditorState((s) => ({ ...s, open: false }))}
            >
              Cancel
            </Button>
            <Button variant="primary" onClick={handleSave}>
              {editorState.mode === 'create' ? 'Create' : 'Save changes'}
            </Button>
          </div>
        </div>
      </Drawer>

      <Dialog
        open={Boolean(pendingDelete)}
        onClose={() => setPendingDelete(null)}
        title="Delete document?"
        description="This action cannot be undone."
        onConfirm={() => handleDelete(pendingDelete)}
        confirmLabel="Delete"
        destructive
      >
        <p className="text-sm text-ink-600 dark:text-ink-300">
          Delete document{' '}
          <code className="rounded bg-ink-100 px-1.5 py-0.5 font-mono text-xs dark:bg-ink-800">
            {pendingDelete && deriveId(pendingDelete)}
          </code>
          ?
        </p>
      </Dialog>

      <Dialog
        open={pendingBulkDelete}
        onClose={() => setPendingBulkDelete(false)}
        title={`Delete ${selected.size} documents?`}
        description="This action cannot be undone."
        onConfirm={handleBulkDelete}
        confirmLabel="Delete all"
        destructive
      >
        <p className="text-sm text-ink-600 dark:text-ink-300">
          You are about to delete{' '}
          <span className="font-semibold text-rose-500">{selected.size}</span> documents
          from <code className="font-mono">{table}</code>. Continue?
        </p>
      </Dialog>

      <EndpointLink
        open={endpointOpen}
        onClose={() => setEndpointOpen(false)}
        projectId={projectId}
        table={table}
      />
    </div>
  )
}
