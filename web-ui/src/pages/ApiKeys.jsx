import { useCallback, useEffect, useState } from 'react'
import {
  Key,
  Plus,
  Trash2,
  Copy,
  Eye,
  EyeOff,
  Shield,
  ExternalLink,
  RefreshCcw
} from 'lucide-react'
import Card, { CardHeader, CardTitle, CardBody } from '@/components/ui/Card.jsx'
import Button from '@/components/ui/Button.jsx'
import Input from '@/components/ui/Input.jsx'
import Dialog from '@/components/ui/Dialog.jsx'
import Skeleton from '@/components/ui/Skeleton.jsx'
import { useApi } from '@/api/client.js'
import { useToast } from '@/components/ui/Toast.jsx'
import { useProjectsStore } from '@/stores/useProjectsStore.js'
import { useSettingsStore } from '@/stores/useSettingsStore.js'
import { formatRelative, truncate } from '@/utils/format.js'

export default function ApiKeys() {
  const api = useApi()
  const toast = useToast()
  const baseUrl = useSettingsStore((s) => s.baseUrl)
  const knownProjects = useProjectsStore((s) => s.knownProjects)

  const [keys, setKeys] = useState([])
  const [loading, setLoading] = useState(true)
  const [selectedProject, setSelectedProject] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyProject, setNewKeyProject] = useState('')
  const [pendingRevoke, setPendingRevoke] = useState(null)
  const [revealedKeys, setRevealedKeys] = useState(new Set())

  const refresh = useCallback(async () => {
    if (!selectedProject) {
      setKeys([])
      setLoading(false)
      return
    }
    setLoading(true)
    try {
      const res = await api.get('/__api/keys', { params: { projectId: selectedProject } })
      setKeys(Array.isArray(res.data) ? res.data : [])
    } catch (e) {
      toast.error('Failed to load API keys')
      setKeys([])
    } finally {
      setLoading(false)
    }
  }, [api, selectedProject, toast])

  useEffect(() => {
    refresh()
  }, [refresh])

  useEffect(() => {
    if (!selectedProject && knownProjects.length > 0) {
      setSelectedProject(knownProjects[0].id)
    }
  }, [selectedProject, knownProjects])

  async function handleCreate() {
    if (!newKeyProject) {
      toast.warning('Select a project')
      return
    }
    if (!newKeyName.trim()) {
      toast.warning('Enter a key name')
      return
    }
    try {
      const res = await api.post('/__api/keys', {
        projectId: newKeyProject,
        name: newKeyName.trim()
      })
      toast.success('API key created')
      setCreateOpen(false)
      setNewKeyName('')
      setSelectedProject(newKeyProject)
      // Show the new key temporarily
      setKeys((prev) => [res.data, ...prev])
    } catch (e) {
      toast.error('Failed to create API key')
    }
  }

  async function handleRevoke(key) {
    try {
      await api.delete(`/__api/keys/${key.projectId}/${key.id}`)
      toast.success('API key revoked')
      setPendingRevoke(null)
      refresh()
    } catch (e) {
      toast.error('Failed to revoke API key')
    }
  }

  function toggleReveal(keyId) {
    setRevealedKeys((prev) => {
      const next = new Set(prev)
      if (next.has(keyId)) next.delete(keyId)
      else next.add(keyId)
      return next
    })
  }

  function copyKey(key) {
    navigator.clipboard.writeText(key).then(
      () => toast.success('API key copied to clipboard'),
      () => toast.error('Failed to copy')
    )
  }

  function getEndpointBase() {
    return baseUrl || window.location.origin
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="font-display text-2xl font-semibold tracking-tight">API Keys</h1>
          <p className="mt-1 text-sm text-ink-500">
            Manage API keys for external access to your sandbox endpoints.
          </p>
        </div>
        <Button variant="primary" onClick={() => setCreateOpen(true)}>
          <Plus className="h-4 w-4" />
          Generate key
        </Button>
      </div>

      {/* Project selector */}
      <Card>
        <CardHeader>
          <CardTitle>Project</CardTitle>
        </CardHeader>
        <CardBody>
          <div className="flex items-center gap-3">
            <select
              value={selectedProject}
              onChange={(e) => setSelectedProject(e.target.value)}
              className="h-9 rounded-md border border-ink-200 bg-white px-3 text-sm dark:border-ink-700 dark:bg-ink-900"
            >
              <option value="">Select a project…</option>
              {knownProjects.map((p) => (
                <option key={p.id} value={p.id}>{p.id}</option>
              ))}
            </select>
            <Button variant="ghost" onClick={refresh} title="Refresh">
              <RefreshCcw className="h-4 w-4" />
            </Button>
          </div>
        </CardBody>
      </Card>

      {/* Endpoint example */}
      {selectedProject && (
        <div className="panel-muted p-4 text-xs text-ink-500">
          <div className="flex items-center gap-2 font-semibold text-ink-700 dark:text-ink-300">
            <ExternalLink className="h-3.5 w-3.5" />
            Base endpoint
          </div>
          <code className="mt-1 block font-mono text-ink-600 dark:text-ink-400">
            {getEndpointBase()}/sandbox/{'{projectId}'}/{'{table}'}
          </code>
          <p className="mt-1">
            Pass the API key as{' '}
            <code className="font-mono">X-API-Key</code> header or{' '}
            <code className="font-mono">?api_key=</code> query parameter.
          </p>
        </div>
      )}

      {/* Keys table */}
      <Card>
        <CardHeader>
          <CardTitle>Keys</CardTitle>
        </CardHeader>
        <CardBody className="p-0">
          {!selectedProject ? (
            <div className="px-5 py-12 text-center text-sm text-ink-500">
              <Shield className="mx-auto mb-2 h-8 w-8 text-ink-300 dark:text-ink-700" />
              Select a project to view its API keys.
            </div>
          ) : loading ? (
            <div className="space-y-2 p-5">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : keys.length === 0 ? (
            <div className="px-5 py-12 text-center text-sm text-ink-500">
              <Key className="mx-auto mb-2 h-8 w-8 text-ink-300 dark:text-ink-700" />
              No API keys for this project. Generate one to get started.
            </div>
          ) : (
            <table className="table-base">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Key</th>
                  <th>Scopes</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {keys.map((k) => (
                  <tr key={k.id}>
                    <td className="font-medium text-ink-900 dark:text-ink-100">{k.name}</td>
                    <td>
                      <div className="flex items-center gap-1.5 font-mono text-xs">
                        <span className="text-ink-600 dark:text-ink-400">
                          {revealedKeys.has(k.id) ? k.key : truncate(k.key, 20)}
                        </span>
                        <button
                          onClick={() => toggleReveal(k.id)}
                          className="rounded p-1 text-ink-400 transition hover:bg-ink-100 hover:text-ink-600 dark:hover:bg-ink-800"
                          title={revealedKeys.has(k.id) ? 'Hide key' : 'Reveal key'}
                        >
                          {revealedKeys.has(k.id) ? (
                            <EyeOff className="h-3 w-3" />
                          ) : (
                            <Eye className="h-3 w-3" />
                          )}
                        </button>
                        <button
                          onClick={() => copyKey(k.key)}
                          className="rounded p-1 text-ink-400 transition hover:bg-ink-100 hover:text-ink-600 dark:hover:bg-ink-800"
                          title="Copy key"
                        >
                          <Copy className="h-3 w-3" />
                        </button>
                      </div>
                    </td>
                    <td className="text-xs text-ink-500">
                      {(k.scopes || []).join(', ') || '—'}
                    </td>
                    <td>
                      <span
                        className={`badge text-[10px] ${
                          k.active
                            ? 'bg-brand-500/15 text-brand-600 dark:text-brand-400'
                            : 'bg-ink-200 text-ink-600 dark:bg-ink-700 dark:text-ink-400'
                        }`}
                      >
                        {k.active ? 'active' : 'revoked'}
                      </span>
                    </td>
                    <td className="text-xs text-ink-500">
                      {formatRelative(k.createdAt)}
                    </td>
                    <td className="text-right">
                      {k.active && (
                        <button
                          onClick={() => setPendingRevoke(k)}
                          className="rounded p-1.5 text-ink-400 transition hover:bg-rose-500/10 hover:text-rose-500"
                          title="Revoke key"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </CardBody>
      </Card>

      {/* Usage guide */}
      <div className="panel-muted p-4 text-xs text-ink-500">
        <div className="font-semibold text-ink-700 dark:text-ink-300">How to use</div>
        <p className="mt-1">Include the API key in your HTTP requests:</p>
        <pre className="mt-2 overflow-auto rounded-md border border-ink-200 bg-white p-3 font-mono text-[11px] leading-relaxed dark:border-ink-700 dark:bg-ink-900">
{`// JavaScript / fetch
const res = await fetch('${getEndpointBase()}/sandbox/${selectedProject || '{project}'}/{table}', {
  headers: { 'X-API-Key': '<YOUR_API_KEY>' }
});
const data = await res.json();

// cURL
curl -H "X-API-Key: <YOUR_API_KEY>" \\
  ${getEndpointBase()}/sandbox/${selectedProject || '{project}'}/{table}`}
        </pre>
      </div>

      {/* Create dialog */}
      <Dialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title="Generate API Key"
        description="Create a new key for external sandbox access."
        onConfirm={handleCreate}
        confirmLabel="Generate"
      >
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium">Project</label>
            <select
              value={newKeyProject}
              onChange={(e) => setNewKeyProject(e.target.value)}
              className="mt-1 w-full rounded-md border border-ink-200 bg-white px-3 py-2 text-sm dark:border-ink-700 dark:bg-ink-900"
            >
              <option value="">Select project…</option>
              {knownProjects.map((p) => (
                <option key={p.id} value={p.id}>{p.id}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium">Key name</label>
            <Input
              autoFocus
              placeholder="e.g. My React App, Frontend Staging"
              value={newKeyName}
              onChange={(e) => setNewKeyName(e.target.value)}
              className="mt-1"
            />
          </div>
        </div>
      </Dialog>

      {/* Revoke dialog */}
      <Dialog
        open={Boolean(pendingRevoke)}
        onClose={() => setPendingRevoke(null)}
        title="Revoke API key?"
        description="This key will immediately stop working for all clients."
        onConfirm={() => handleRevoke(pendingRevoke)}
        confirmLabel="Revoke"
        destructive
      >
        <p className="text-sm text-ink-600 dark:text-ink-300">
          Revoke key{' '}
          <code className="rounded bg-ink-100 px-1.5 py-0.5 font-mono text-xs dark:bg-ink-800">
            {pendingRevoke?.name}
          </code>
          ?
        </p>
      </Dialog>
    </div>
  )
}
