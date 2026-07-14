import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, RefreshCcw, Terminal, Edit3, Trash2 } from 'lucide-react'
import Card, { CardHeader, CardTitle, CardBody } from '@/components/ui/Card.jsx'
import Button from '@/components/ui/Button.jsx'
import Skeleton from '@/components/ui/Skeleton.jsx'
import StatusBadge from '@/components/ui/StatusBadge.jsx'
import { useApi } from '@/api/client.js'
import { useToast } from '@/components/ui/Toast.jsx'
import { getDocument, deleteDocument, updateDocument } from '@/api/sandbox.js'
import { formatDate, formatRelative } from '@/utils/format.js'
import Drawer from '@/components/ui/Drawer.jsx'
import JsonEditor from '@/components/data/JsonEditor.jsx'
import Dialog from '@/components/ui/Dialog.jsx'

const META_KEYS = ['_createdAt', '_createdBy', '_updatedAt', '_updatedBy']

export default function DocumentDetail() {
  const { projectId, table, docId } = useParams()
  const api = useApi()
  const toast = useToast()

  const [doc, setDoc] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [editOpen, setEditOpen] = useState(false)
  const [editBody, setEditBody] = useState('{}')
  const [confirmDelete, setConfirmDelete] = useState(false)

  async function refresh() {
    setLoading(true)
    setError(null)
    try {
      const d = await getDocument(api, projectId, table, docId)
      setDoc(d)
      const copy = { ...d }
      META_KEYS.forEach((k) => delete copy[k])
      delete copy.id
      setEditBody(JSON.stringify(copy, null, 2))
    } catch (e) {
      setError(e.normalizedMessage || e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId, table, docId])

  async function handleSave() {
    let parsed
    try {
      parsed = JSON.parse(editBody || '{}')
    } catch (e) {
      toast.error('Invalid JSON: ' + e.message)
      return
    }
    try {
      await updateDocument(api, projectId, table, docId, parsed)
      toast.success('Document updated')
      setEditOpen(false)
      refresh()
    } catch (e) {
      toast.error(e.normalizedMessage || e.message)
    }
  }

  async function handleDelete() {
    try {
      await deleteDocument(api, projectId, table, docId)
      toast.success('Document deleted')
    } catch (e) {
      toast.error(e.normalizedMessage || e.message)
    } finally {
      setConfirmDelete(false)
    }
  }

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-40 w-full" />
      </div>
    )
  }
  if (error) {
    return (
      <Card className="p-8 text-center">
        <div className="font-display text-base font-semibold text-rose-500">
          Failed to load document
        </div>
        <p className="mt-1 text-sm text-ink-500">{error}</p>
        <Button className="mt-3" variant="outline" onClick={refresh}>
          Retry
        </Button>
      </Card>
    )
  }

  const meta = META_KEYS.reduce((acc, k) => {
    if (doc?.[k] != null) acc[k] = doc[k]
    return acc
  }, {})

  const other = doc
    ? Object.fromEntries(
        Object.entries(doc).filter(
          ([k]) => !META_KEYS.includes(k) && k !== 'id' && k !== '_id'
        )
      )
    : {}

  return (
    <div className="space-y-6">
      <div>
        <Link
          to={`/projects/${projectId}/${table}`}
          className="inline-flex items-center gap-1 text-xs text-ink-500 transition hover:text-ink-900 dark:hover:text-ink-100"
        >
          <ArrowLeft className="h-3.5 w-3.5" /> Back to {table}
        </Link>
        <div className="mt-1 flex flex-wrap items-end justify-between gap-3">
          <div>
            <h1 className="font-display text-2xl font-semibold tracking-tight">
              <span className="text-ink-400">{projectId} / {table} /</span>{' '}
              <span className="font-mono text-base">{docId}</span>
            </h1>
            <p className="mt-1 text-sm text-ink-500">Document detail view.</p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" onClick={refresh}>
              <RefreshCcw className="h-4 w-4" />
            </Button>
            <Link to={`/tester?path=/sandbox/${projectId}/${table}/${docId}&method=GET`}>
              <Button variant="outline">
                <Terminal className="h-4 w-4" /> Open in tester
              </Button>
            </Link>
            <Button variant="primary" onClick={() => setEditOpen(true)}>
              <Edit3 className="h-4 w-4" /> Edit
            </Button>
            <Button variant="danger" onClick={() => setConfirmDelete(true)}>
              <Trash2 className="h-4 w-4" /> Delete
            </Button>
          </div>
        </div>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Body</CardTitle>
            <span className="font-mono text-xs text-ink-500">
              {Object.keys(other).length} field{Object.keys(other).length === 1 ? '' : 's'}
            </span>
          </CardHeader>
          <CardBody>
            {Object.keys(other).length === 0 ? (
              <p className="text-sm text-ink-500">No body fields.</p>
            ) : (
              <div className="divide-y divide-ink-100 dark:divide-ink-800">
                {Object.entries(other).map(([k, v]) => (
                  <div key={k} className="grid grid-cols-3 gap-3 py-2.5 text-sm">
                    <div className="font-mono text-xs text-ink-500">{k}</div>
                    <div className="col-span-2 break-all font-mono text-xs">
                      {typeof v === 'object' && v !== null
                        ? JSON.stringify(v, null, 2)
                        : String(v)}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Metadata</CardTitle>
          </CardHeader>
          <CardBody className="space-y-3 text-sm">
            {Object.keys(meta).length === 0 ? (
              <p className="text-ink-500">No metadata.</p>
            ) : (
              Object.entries(meta).map(([k, v]) => (
                <div key={k}>
                  <div className="font-mono text-xs text-ink-500">{k}</div>
                  <div className="mt-0.5 text-sm">
                    {k.endsWith('At') && typeof v === 'string'
                      ? formatDate(v)
                      : String(v)}
                  </div>
                </div>
              ))
            )}
            {doc?.id && (
              <div>
                <div className="font-mono text-xs text-ink-500">id</div>
                <div className="mt-0.5 break-all font-mono text-xs">{String(doc.id)}</div>
              </div>
            )}
          </CardBody>
        </Card>
      </div>

      <Drawer
        open={editOpen}
        onClose={() => setEditOpen(false)}
        title={`Edit ${docId}`}
        description="Persisted via PUT (MergeAll). Metadata is recomputed by the server."
      >
        <div className="space-y-3">
          <JsonEditor value={editBody} onChange={setEditBody} rows={20} />
          <div className="flex items-center justify-end gap-2">
            <Button variant="ghost" onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <Button variant="primary" onClick={handleSave}>
              Save changes
            </Button>
          </div>
        </div>
      </Drawer>

      <Dialog
        open={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        title="Delete document?"
        description="This action cannot be undone."
        onConfirm={handleDelete}
        confirmLabel="Delete"
        destructive
      >
        <p className="text-sm text-ink-600 dark:text-ink-300">
          Delete{' '}
          <code className="rounded bg-ink-100 px-1.5 py-0.5 font-mono text-xs dark:bg-ink-800">
            {docId}
          </code>{' '}
          from {table}?
        </p>
      </Dialog>
    </div>
  )
}
