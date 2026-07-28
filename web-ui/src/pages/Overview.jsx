import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  Database,
  FolderTree,
  FileJson,
  Activity,
  ArrowUpRight,
  Plus,
  Terminal,
  PlayCircle,
  Key,
  Settings
} from 'lucide-react'
import StatCard from '@/components/stats/StatCard.jsx'
import { Card, CardHeader, CardTitle, CardBody } from '@/components/ui/Card.jsx'
import StatusBadge from '@/components/ui/StatusBadge.jsx'
import Skeleton from '@/components/ui/Skeleton.jsx'
import Button from '@/components/ui/Button.jsx'
import { useApi } from '@/api/client.js'
import { useProjectsStore } from '@/stores/useProjectsStore.js'
import { formatRelative } from '@/utils/format.js'

export default function Overview() {
  const api = useApi()
  const knownProjects = useProjectsStore((s) => s.knownProjects)
  // Subscribe to a stable string signature of the tracked tables so this
  // effect doesn't re-run when only `docCount` (which we update ourselves
  // below) changes. Otherwise the effect would loop forever and exhaust
  // the browser's connection pool (ERR_INSUFFICIENT_RESOURCES).
  const tablesSignature = useProjectsStore((s) =>
    Object.entries(s.knownTables)
      .flatMap(([pid, list]) => (list || []).map((t) => `${pid}/${t.name}`))
      .sort()
      .join('|')
  )

  const [stats, setStats] = useState({ projects: 0, tables: 0, documents: 0, online: null })
  const [recent, setRecent] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let alive = true
    let requestSeq = 0
    const myReq = ++requestSeq
    async function load() {
      setLoading(true)
      const projectIds = knownProjects.map((p) => p.id)

      // Auto-sync: discover tables from backend for each project.
      // Falls back to existing localStorage data if the backend is unreachable.
      for (const pid of projectIds) {
        if (!alive || myReq !== requestSeq) return
        try {
          const res = await api.get('/__api/tables', { params: { projectId: pid } })
          if (!alive || myReq !== requestSeq) return
          const tables = Array.isArray(res.data?.tables) ? res.data.tables : []
          const store = useProjectsStore.getState()
          for (const name of tables) {
            store.addTable(pid, name)
          }
        } catch (e) {
          // Backend unreachable — keep local data as-is
        }
      }

      // Build a stable list from the signature (not the live store object).
      const tablePairs = tablesSignature
        ? tablesSignature.split('|').map((s) => {
            const [pid, ...rest] = s.split('/')
            return [pid, rest.join('/')]
          })
        : []
      const counts = { projects: projectIds.length, tables: tablePairs.length, documents: 0 }
      const recentRequests = []

      for (const [pid, tname] of tablePairs) {
        if (!alive || myReq !== requestSeq) return
        try {
          const t0 = performance.now()
          const res = await api.get(`/sandbox/${pid}/${tname}`)
          if (!alive || myReq !== requestSeq) return
          const lat = performance.now() - t0
          // Backend returns paginated envelope { data: [...], limit, offset, count }
          const list = Array.isArray(res.data?.data) ? res.data.data : Array.isArray(res.data) ? res.data : []
          counts.documents += list.length
          recentRequests.push({
            id: `${pid}/${tname}`,
            method: 'GET',
            path: `/sandbox/${pid}/${tname}`,
            status: res.status,
            latency: lat,
            timestamp: Date.now()
          })
        } catch (e) {
          // ignore offline / network error for individual table
        }
      }
      // Probe /hello for health
      try {
        const r = await api.get('/hello', { responseType: 'text', transformResponse: [(d) => d] })
        if (!alive || myReq !== requestSeq) return
        counts.online = r.status === 200
        recentRequests.push({
          id: 'hello-probe',
          method: 'GET',
          path: '/hello',
          status: r.status,
          latency: 0,
          timestamp: Date.now()
        })
      } catch (e) {
        if (alive && myReq === requestSeq) counts.online = false
      }

      if (alive && myReq === requestSeq) {
        setStats(counts)
        setRecent(recentRequests.sort((a, b) => b.timestamp - a.timestamp).slice(0, 8))
        setLoading(false)
      }
    }
    load()
    return () => {
      alive = false
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [api, knownProjects.length, tablesSignature])

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="font-display text-2xl font-semibold tracking-tight">
            Welcome back
          </h1>
          <p className="mt-1 text-sm text-ink-500">
            State of your mock sandbox at a glance.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Link to="/projects">
            <Button variant="primary">
              <Plus className="h-4 w-4" />
              New project
            </Button>
          </Link>
          <Link to="/tester">
            <Button variant="outline">
              <Terminal className="h-4 w-4" />
              Open tester
            </Button>
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard
          icon={FolderTree}
          label="Projects"
          value={loading ? '…' : stats.projects}
          hint="Tracked locally"
          accent="brand"
        />
        <StatCard
          icon={Database}
          label="Tables"
          value={loading ? '…' : stats.tables}
          hint="Across all projects"
          accent="sky"
        />
        <StatCard
          icon={FileJson}
          label="Documents"
          value={loading ? '…' : stats.documents}
          hint="Persisted in backend"
          accent="violet"
        />
        <StatCard
          icon={Activity}
          label="Backend"
          value={stats.online == null ? '…' : stats.online ? 'Online' : 'Offline'}
          hint="Probed /hello"
          accent={stats.online ? 'brand' : 'rose'}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Recent activity</CardTitle>
            <span className="text-xs text-ink-500">last {recent.length} request{recent.length !== 1 ? 's' : ''}</span>
          </CardHeader>
          <CardBody className="p-0">
            {loading ? (
              <div className="space-y-2 p-5">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} className="h-9 w-full" />
                ))}
              </div>
            ) : recent.length === 0 ? (
              <div className="px-5 py-12 text-center text-sm text-ink-500">
                <PlayCircle className="mx-auto mb-2 h-8 w-8 text-ink-300 dark:text-ink-700" />
                No activity yet. Open the REST Tester to send your first request.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="table-base">
                  <thead>
                    <tr>
                      <th>Method</th>
                      <th>Path</th>
                      <th>Status</th>
                      <th>Latency</th>
                      <th>When</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recent.map((r) => (
                      <tr key={r.id}>
                        <td>
                          <StatusBadge method={r.method} />
                        </td>
                        <td className="max-w-[20rem] truncate font-mono text-xs">{r.path}</td>
                        <td>
                          <StatusBadge status={r.status} />
                        </td>
                        <td className="font-mono text-xs text-ink-500">
                          {r.latency ? `${r.latency.toFixed(1)} ms` : '—'}
                        </td>
                        <td className="text-xs text-ink-500">
                          {formatRelative(r.timestamp)}
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
            <CardTitle>Quick actions</CardTitle>
          </CardHeader>
          <CardBody className="space-y-2">
            <QuickAction
              to="/projects"
              icon={Plus}
              label="Create new project"
              hint="Add a namespace to track"
            />
            <QuickAction
              to="/tester"
              icon={Terminal}
              label="Send a test request"
              hint="Try a custom endpoint"
            />
            <QuickAction
              to="/api-keys"
              icon={Key}
              label="API Keys"
              hint="Manage access keys"
            />
            <QuickAction
              to="/settings"
              icon={Settings}
              label="Configure backend"
              hint="Base URL & preferences"
            />
          </CardBody>
        </Card>
      </div>
    </div>
  )
}

function QuickAction({ to, icon: Icon, label, hint }) {
  return (
    <Link
      to={to}
      className="group flex items-center gap-3 rounded-lg border border-ink-200 bg-white px-3 py-2.5 transition hover:border-brand-500/60 hover:bg-brand-500/5 dark:border-ink-800 dark:bg-ink-900"
    >
      <div className="flex h-8 w-8 items-center justify-center rounded-md bg-ink-100 text-ink-600 transition group-hover:bg-brand-500/15 group-hover:text-brand-500 dark:bg-ink-800 dark:text-ink-300">
        <Icon className="h-4 w-4" />
      </div>
      <div className="flex-1">
        <div className="text-sm font-medium">{label}</div>
        <div className="text-xs text-ink-500">{hint}</div>
      </div>
      <ArrowUpRight className="h-4 w-4 text-ink-300 transition group-hover:text-brand-500" />
    </Link>
  )
}
