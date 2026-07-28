import { useEffect, useState } from 'react'
import { Link, Copy, Check, Code, Globe } from 'lucide-react'
import Dialog from '@/components/ui/Dialog.jsx'
import Button from '@/components/ui/Button.jsx'
import { useToast } from '@/components/ui/Toast.jsx'
import { useApi } from '@/api/client.js'
import { useSettingsStore } from '@/stores/useSettingsStore.js'

const TABS = [
  { id: 'url', label: 'URL', icon: Globe },
  { id: 'fetch', label: 'Fetch', icon: Code },
  { id: 'axios', label: 'Axios', icon: Code },
  { id: 'curl', label: 'cURL', icon: Code },
  { id: 'python', label: 'Python', icon: Code }
]

export default function EndpointLink({ open, onClose, projectId, table }) {
  const toast = useToast()
  const baseUrl = useSettingsStore((s) => s.baseUrl)
  const api = useApi()
  const [activeTab, setActiveTab] = useState('url')
  const [copied, setCopied] = useState(false)
  const [apiKey, setApiKey] = useState('')
  const [keys, setKeys] = useState([])
  const [keysLoaded, setKeysLoaded] = useState(false)

  // Lazy-load keys when dialog opens
  useEffect(() => {
    if (!open || keysLoaded) return
    let alive = true
    async function loadKeys() {
      try {
        const res = await api.get('/__api/keys', { params: { projectId } })
        const list = Array.isArray(res.data) ? res.data : []
        if (!alive) return
        setKeys(list)
        if (list.length > 0 && !apiKey) {
          setApiKey(list[0].key)
        }
      } catch {
        // silent
      } finally {
        if (alive) setKeysLoaded(true)
      }
    }
    loadKeys()
    return () => { alive = false }
  }, [open, keysLoaded, api, projectId]) // eslint-disable-line react-hooks/exhaustive-deps

  const endpointBase = baseUrl || window.location.origin
  const fullUrl = `${endpointBase}/sandbox/${projectId}/${table}`

  const safeKey = apiKey || '<YOUR_API_KEY>'

  const snippets = {
    url: fullUrl,
    fetch: [
      '// Fetch — GET all documents',
      "const response = await fetch('" + fullUrl + "', {",
      '  headers: {',
      "    'X-API-Key': '" + safeKey + "'",
      '  }',
      '});',
      'const { data, count, nextOffset } = await response.json();',
      'console.log(data);'
    ].join('\n'),

    axios: [
      '// Axios — GET all documents',
      "import axios from 'axios';",
      '',
      "const { data } = await axios.get('" + fullUrl + "', {",
      '  headers: {',
      "    'X-API-Key': '" + safeKey + "'",
      '  }',
      '});',
      'console.log(data.data);'
    ].join('\n'),

    curl: [
      '# cURL — GET all documents',
      "curl -X GET \\",
      "  '" + fullUrl + "' \\",
      "  -H 'X-API-Key: " + safeKey + "' \\",
      "  -H 'Content-Type: application/json'"
    ].join('\n'),

    python: [
      "# Python requests — GET all documents",
      "import requests",
      '',
      "response = requests.get(",
      "    '" + fullUrl + "',",
      "    headers={'X-API-Key': '" + safeKey + "'}",
      ")",
      "data = response.json()",
      "print(data['data'])"
    ].join('\n')
  }

  function copySnippet() {
    navigator.clipboard.writeText(snippets[activeTab]).then(
      () => {
        setCopied(true)
        toast.success('Copied to clipboard')
        setTimeout(() => setCopied(false), 2000)
      },
      () => toast.error('Failed to copy')
    )
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Endpoint Link"
      description={`Shareable endpoint for ${projectId}/${table}`}
      size="lg"
    >
      <div className="space-y-4">
        {/* API Key selector */}
        <div>
          <label className="block text-sm font-medium text-ink-700 dark:text-ink-300">
            API Key
          </label>
          {keys.length > 0 ? (
            <select
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              className="mt-1 w-full rounded-md border border-ink-200 bg-white px-3 py-2 font-mono text-xs dark:border-ink-700 dark:bg-ink-900"
            >
              {keys.filter((k) => k.active).map((k) => (
                <option key={k.id} value={k.key}>
                  {k.name} — {k.key.slice(0, 24)}…
                </option>
              ))}
            </select>
          ) : (
            <div className="mt-1 rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-xs text-amber-700 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-300">
              No API keys found for this project.{' '}
              <a href="/api-keys" className="underline hover:no-underline" onClick={onClose}>
                Create one first →
              </a>
            </div>
          )}
        </div>

        {/* URL preview */}
        <div className="rounded-md border border-ink-200 bg-ink-50 p-3 dark:border-ink-700 dark:bg-ink-950">
          <div className="flex items-center gap-2 text-xs text-ink-500">
            <Link className="h-3.5 w-3.5" />
            <span>Endpoint URL</span>
          </div>
          <code className="mt-1 block break-all font-mono text-sm text-ink-800 dark:text-ink-200">
            {fullUrl}
          </code>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 rounded-lg border border-ink-200 bg-ink-100 p-1 dark:border-ink-700 dark:bg-ink-800">
          {TABS.map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition ${
                  activeTab === tab.id
                    ? 'bg-white text-ink-900 shadow-sm dark:bg-ink-700 dark:text-ink-100'
                    : 'text-ink-500 hover:text-ink-700 dark:hover:text-ink-300'
                }`}
              >
                <Icon className="h-3 w-3" />
                {tab.label}
              </button>
            )
          })}
        </div>

        {/* Code snippet */}
        <div className="relative">
          <pre className="max-h-[320px] overflow-auto rounded-md border border-ink-200 bg-ink-950 p-4 font-mono text-xs leading-relaxed text-ink-100 dark:border-ink-700">
            {snippets[activeTab]}
          </pre>
          <button
            onClick={copySnippet}
            className="absolute right-2 top-2 rounded-md border border-ink-700 bg-ink-800 p-1.5 text-ink-300 transition hover:bg-ink-700 hover:text-white"
            title="Copy to clipboard"
          >
            {copied ? <Check className="h-3.5 w-3.5 text-brand-400" /> : <Copy className="h-3.5 w-3.5" />}
          </button>
        </div>

        {/* Additional endpoints */}
        <div className="text-xs text-ink-500">
          <div className="mb-1 font-medium text-ink-700 dark:text-ink-300">
            Available endpoints
          </div>
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="badge badge-method-get font-mono text-[10px]">GET</span>
              <code className="font-mono">/sandbox/{projectId}/{table}</code>
              <span className="text-ink-400">— list documents (paginated)</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="badge badge-method-get font-mono text-[10px]">GET</span>
              <code className="font-mono">/sandbox/{projectId}/{table}/{'<id>'}</code>
              <span className="text-ink-400">— get single document</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="badge badge-method-post font-mono text-[10px]">POST</span>
              <code className="font-mono">/sandbox/{projectId}/{table}</code>
              <span className="text-ink-400">— create document</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="badge badge-method-put font-mono text-[10px]">PUT</span>
              <code className="font-mono">/sandbox/{projectId}/{table}/{'<id>'}</code>
              <span className="text-ink-400">— update document</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="badge badge-method-delete font-mono text-[10px]">DEL</span>
              <code className="font-mono">/sandbox/{projectId}/{table}/{'<id>'}</code>
              <span className="text-ink-400">— delete document</span>
            </div>
          </div>
        </div>
      </div>
    </Dialog>
  )
}
