import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Send, Clock, Trash2, ChevronRight, Zap } from 'lucide-react'
import Card, { CardHeader, CardTitle, CardBody } from '@/components/ui/Card.jsx'
import Button from '@/components/ui/Button.jsx'
import Select from '@/components/ui/Select.jsx'
import Input from '@/components/ui/Input.jsx'
import Textarea from '@/components/ui/Textarea.jsx'
import StatusBadge from '@/components/ui/StatusBadge.jsx'
import { useApi } from '@/api/client.js'
import { useToast } from '@/components/ui/Toast.jsx'
import { useTesterStore } from '@/stores/useTesterStore.js'
import { useProjectsStore } from '@/stores/useProjectsStore.js'
import { formatLatency, formatRelative, truncate } from '@/utils/format.js'
import { cn } from '@/utils/cn.js'

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']

function defaultTemplate(projectId, table) {
  return {
    name: 'Sample record',
    role: 'engineer',
    tags: ['demo'],
    active: true
  }
}

export default function RESTTester() {
  const api = useApi()
  const toast = useToast()
  const [params] = useSearchParams()
  const knownProjects = useProjectsStore((s) => s.knownProjects)
  const knownTables = useProjectsStore((s) => s.knownTables)
  const history = useTesterStore((s) => s.history)
  const addHistory = useTesterStore((s) => s.addHistory)
  const clearHistory = useTesterStore((s) => s.clearHistory)

  const initialPath = params.get('path') || '/sandbox/demo/users'
  const initialMethod = params.get('method') || 'GET'

  const [method, setMethod] = useState(initialMethod)
  const [path, setPath] = useState(initialPath)
  const [body, setBody] = useState(
    JSON.stringify(defaultTemplate('demo', 'users'), null, 2)
  )
  const [sending, setSending] = useState(false)
  const [response, setResponse] = useState(null)

  // Autocomplete suggestions derived from project/table tracking
  const suggestions = useMemo(() => {
    const list = []
    for (const p of knownProjects) {
      list.push(`/sandbox/${p.id}`)
      for (const t of knownTables[p.id] || []) {
        list.push(`/sandbox/${p.id}/${t.name}`)
        list.push(`/sandbox/${p.id}/${t.name}/<id>`)
      }
    }
    return list
  }, [knownProjects, knownTables])

  const send = useCallback(async () => {
    setSending(true)
    setResponse(null)
    const t0 = performance.now()
    try {
      let data = body
      let isJson = true
      if (['POST', 'PUT', 'PATCH'].includes(method)) {
        try {
          if (body && body.trim().length > 0) {
            JSON.parse(body)
            isJson = true
          } else {
            data = undefined
            isJson = false
          }
        } catch (e) {
          toast.error('Body is not valid JSON: ' + e.message)
          setSending(false)
          return
        }
      } else {
        data = undefined
      }
      const config = {
        method,
        url: path,
        data,
        headers: isJson ? { 'Content-Type': 'application/json' } : {},
        transformRequest: [(d) => (typeof d === 'string' ? d : JSON.stringify(d))]
      }
      const res = await api.request(config)
      const elapsed = performance.now() - t0
      const resp = {
        status: res.status,
        statusText: res.statusText,
        headers: res.headers,
        data: res.data,
        latency: elapsed,
        timestamp: Date.now()
      }
      setResponse(resp)
      addHistory({
        method,
        path,
        status: res.status,
        latency: elapsed,
        timestamp: Date.now()
      })
      toast.success(`${method} ${path} → ${res.status}`)
    } catch (e) {
      const elapsed = performance.now() - t0
      const status = e.response?.status || 0
      const resp = {
        status,
        statusText: e.response?.statusText || 'ERROR',
        headers: e.response?.headers || {},
        data: e.response?.data || e.normalizedMessage || e.message,
        latency: elapsed,
        timestamp: Date.now(),
        error: true
      }
      setResponse(resp)
      addHistory({ method, path, status, latency: elapsed, timestamp: Date.now() })
      toast.error(`${method} ${path} → ${status || 'ERR'}`)
    } finally {
      setSending(false)
    }
  }, [api, method, path, body, addHistory, toast])

  // Keyboard shortcut: Ctrl/Cmd+Enter to send
  useEffect(() => {
    const onKey = (e) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault()
        send()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [send])

  function loadFromHistory(entry) {
    setMethod(entry.method)
    setPath(entry.path)
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-2xl font-semibold tracking-tight">
          REST Tester
        </h1>
        <p className="mt-1 text-sm text-ink-500">
          Send any request to the backend. Use{' '}
          <kbd className="rounded border border-ink-200 bg-ink-50 px-1.5 py-0.5 font-mono text-xs dark:border-ink-700 dark:bg-ink-800">
            Ctrl+Enter
          </kbd>{' '}
          to send.
        </p>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Request</CardTitle>
            <Button
              variant="primary"
              size="sm"
              onClick={send}
              disabled={sending || !path}
            >
              {sending ? (
                <>
                  <Zap className="h-3.5 w-3.5 animate-pulse" />
                  Sending…
                </>
              ) : (
                <>
                  <Send className="h-3.5 w-3.5" />
                  Send
                </>
              )}
            </Button>
          </CardHeader>
          <CardBody className="space-y-4">
            <div className="flex flex-col gap-2 sm:flex-row">
              <div className="sm:w-36">
                <Select value={method} onChange={(e) => setMethod(e.target.value)}>
                  {METHODS.map((m) => (
                    <option key={m} value={m}>
                      {m}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="flex-1">
                <Input
                  value={path}
                  onChange={(e) => setPath(e.target.value)}
                  placeholder="/sandbox/:projectId/:table/:id?"
                  spellCheck={false}
                />
                {suggestions.length > 0 && (
                  <datalist id="path-suggestions">
                    {suggestions.map((s) => (
                      <option key={s} value={s} />
                    ))}
                  </datalist>
                )}
              </div>
            </div>
            {['POST', 'PUT', 'PATCH'].includes(method) && (
              <div>
                <div className="mb-1.5 flex items-center justify-between">
                  <label className="text-xs font-medium text-ink-500">Body (JSON)</label>
                  <button
                    className="text-xs text-brand-500 hover:underline"
                    onClick={() =>
                      setBody(JSON.stringify(defaultTemplate('demo', 'users'), null, 2))
                    }
                  >
                    Use template
                  </button>
                </div>
                <Textarea
                  rows={10}
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                  spellCheck={false}
                />
              </div>
            )}
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>History</CardTitle>
            {history.length > 0 && (
              <button
                onClick={clearHistory}
                className="text-xs text-ink-500 transition hover:text-rose-500"
              >
                <Trash2 className="inline h-3 w-3" /> clear
              </button>
            )}
          </CardHeader>
          <CardBody className="p-0">
            {history.length === 0 ? (
              <div className="px-5 py-10 text-center text-xs text-ink-500">
                No requests yet.
              </div>
            ) : (
              <ul className="divide-y divide-ink-100 dark:divide-ink-800">
                {history.map((h, i) => (
                  <li
                    key={`${h.timestamp}-${i}`}
                    className="group flex cursor-pointer items-center gap-2 px-4 py-2.5 transition hover:bg-brand-500/5"
                    onClick={() => loadFromHistory(h)}
                  >
                    <StatusBadge method={h.method} />
                    <div className="flex-1 truncate font-mono text-xs">{truncate(h.path, 32)}</div>
                    <StatusBadge status={h.status} />
                    <ChevronRight className="h-3.5 w-3.5 text-ink-300 transition group-hover:text-brand-500" />
                  </li>
                ))}
              </ul>
            )}
          </CardBody>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Response</CardTitle>
          {response && (
            <div className="flex items-center gap-2 text-xs">
              <span className={cn('glow-dot', response.error && 'bg-rose-500')} />
              <span className="font-mono text-ink-500">
                {formatRelative(response.timestamp)}
              </span>
            </div>
          )}
        </CardHeader>
        <CardBody>
          {!response ? (
            <div className="flex h-32 items-center justify-center text-sm text-ink-500">
              Send a request to see the response here.
            </div>
          ) : (
            <div className="space-y-4">
              <div className="flex flex-wrap items-center gap-2 text-sm">
                <StatusBadge status={response.status} />
                <span className="text-ink-500">{response.statusText}</span>
                <span className="text-ink-400">·</span>
                <span className="font-mono text-ink-500">
                  {formatLatency(response.latency)}
                </span>
              </div>
              <div>
                <div className="mb-1 text-xs font-medium uppercase tracking-wide text-ink-500">
                  Body
                </div>
                <pre className="max-h-[420px] overflow-auto rounded-md border border-ink-200 bg-ink-50 p-3 font-mono text-xs leading-relaxed dark:border-ink-800 dark:bg-ink-950">
                  {formatBody(response.data)}
                </pre>
              </div>
              <div>
                <div className="mb-1 text-xs font-medium uppercase tracking-wide text-ink-500">
                  Headers
                </div>
                <div className="rounded-md border border-ink-200 p-3 font-mono text-xs dark:border-ink-800">
                  {Object.keys(response.headers || {}).length === 0 ? (
                    <span className="text-ink-500">—</span>
                  ) : (
                    Object.entries(response.headers).map(([k, v]) => (
                      <div key={k} className="grid grid-cols-3 gap-2">
                        <div className="text-ink-500">{k}</div>
                        <div className="col-span-2 break-all">{String(v)}</div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
          )}
        </CardBody>
      </Card>
    </div>
  )
}

function formatBody(data) {
  if (data == null) return '—'
  if (typeof data === 'string') return data
  try {
    return JSON.stringify(data, null, 2)
  } catch (e) {
    return String(data)
  }
}
