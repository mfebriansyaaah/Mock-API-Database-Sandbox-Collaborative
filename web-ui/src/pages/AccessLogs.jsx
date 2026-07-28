import { useEffect, useState } from 'react'
import {
  Activity,
  RefreshCcw,
  ExternalLink,
  Info
} from 'lucide-react'
import Card, { CardHeader, CardTitle, CardBody } from '@/components/ui/Card.jsx'
import Button from '@/components/ui/Button.jsx'
import Skeleton from '@/components/ui/Skeleton.jsx'
import SearchInput from '@/components/data/SearchInput.jsx'
import Select from '@/components/ui/Select.jsx'
import StatusBadge from '@/components/ui/StatusBadge.jsx'
import { useTesterStore } from '@/stores/useTesterStore.js'
import { formatLatency, formatRelative, truncate } from '@/utils/format.js'

export default function AccessLogs() {
  const history = useTesterStore((s) => s.history)
  const [filterMethod, setFilterMethod] = useState('ALL')
  const [filterStatus, setFilterStatus] = useState('ALL')
  const [search, setSearch] = useState('')
  const [autoRefresh, setAutoRefresh] = useState(false)

  useEffect(() => {
    if (!autoRefresh) return
    const id = setInterval(() => {
      // Force re-render by toggling a state — but since history is a stable
      // reference we can also just rely on the user making more requests.
    }, 5000)
    return () => clearInterval(id)
  }, [autoRefresh])

  const filtered = history.filter((h) => {
    if (filterMethod !== 'ALL' && h.method !== filterMethod) return false
    if (filterStatus !== 'ALL') {
      if (filterStatus === '2xx' && !(h.status >= 200 && h.status < 300)) return false
      if (filterStatus === '4xx' && !(h.status >= 400 && h.status < 500)) return false
      if (filterStatus === '5xx' && !(h.status >= 500)) return false
    }
    if (search && !h.path.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-2xl font-semibold tracking-tight">
          Access Logs
        </h1>
        <p className="mt-1 text-sm text-ink-500">
          Local view of recent requests made from this console. The Go backend also
          streams logs to Firestore.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Local request stream</CardTitle>
          <div className="flex items-center gap-2">
            <label className="flex items-center gap-1.5 text-xs text-ink-500">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="h-3.5 w-3.5 accent-brand-500"
              />
              auto
            </label>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => {
                /* no-op, history is live */
              }}
            >
              <RefreshCcw className="h-3.5 w-3.5" />
            </Button>
          </div>
        </CardHeader>
        <CardBody>
          <div className="mb-3 flex flex-wrap items-center gap-2">
            <div className="w-64">
              <SearchInput
                value={search}
                onChange={setSearch}
                placeholder="Filter by path…"
              />
            </div>
            <span className="text-xs font-medium text-ink-500">Method:</span>
            <div className="w-32">
              <Select
                value={filterMethod}
                onChange={(e) => setFilterMethod(e.target.value)}
              >
                {['ALL', 'GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map((m) => (
                  <option key={m} value={m}>
                    {m}
                  </option>
                ))}
              </Select>
            </div>
            <span className="text-xs font-medium text-ink-500">Status:</span>
            <div className="w-32">
              <Select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
              >
                {['ALL', '2xx', '4xx', '5xx'].map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
              </Select>
            </div>
          </div>

          {filtered.length === 0 ? (
            <div className="px-5 py-12 text-center text-sm text-ink-500">
              <Activity className="mx-auto mb-2 h-8 w-8 text-ink-300 dark:text-ink-700" />
              {history.length === 0
                ? 'No requests yet. Use the REST Tester or browse Documents to populate.'
                : 'No entries match the current filter.'}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="table-base">
                <thead>
                  <tr>
                    <th>When</th>
                    <th>Method</th>
                    <th>Path</th>
                    <th>Status</th>
                    <th>Latency</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((h, i) => (
                    <tr key={`${h.timestamp}-${i}`}>
                      <td className="text-xs text-ink-500">
                        {formatRelative(h.timestamp)}
                      </td>
                      <td>
                        <StatusBadge method={h.method} />
                      </td>
                      <td className="max-w-[24rem] truncate font-mono text-xs">
                        {h.path}
                      </td>
                      <td>
                        <StatusBadge status={h.status} />
                      </td>
                      <td className="font-mono text-xs text-ink-500">
                        {formatLatency(h.latency)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardBody>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>About backend logs</CardTitle>
        </CardHeader>
        <CardBody>
          <div className="flex gap-3 text-sm text-ink-600 dark:text-ink-300">
            <Info className="mt-0.5 h-4 w-4 shrink-0 text-brand-500" />
            <p>
              The Go backend writes every request to a Firestore{' '}
              <code className="font-mono">/projects/{'{projectId}'}/logs</code>{' '}
              collection via an async, non-blocking pipeline. Logs are FIFO-capped
              per project. This console only shows requests issued from inside the
              UI; for the full audit trail, query the Firestore collection directly.
            </p>
          </div>
        </CardBody>
      </Card>
    </div>
  )
}
