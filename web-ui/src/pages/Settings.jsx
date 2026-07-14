import { useRef, useState } from 'react'
import { Download, Upload, RotateCcw, Wifi, Check, AlertTriangle } from 'lucide-react'
import Card, { CardHeader, CardTitle, CardBody } from '@/components/ui/Card.jsx'
import Button from '@/components/ui/Button.jsx'
import Input from '@/components/ui/Input.jsx'
import Select from '@/components/ui/Select.jsx'
import { useSettingsStore } from '@/stores/useSettingsStore.js'
import { useApi } from '@/api/client.js'
import { useToast } from '@/components/ui/Toast.jsx'
import { pingHello } from '@/api/system.js'

export default function Settings() {
  const baseUrl = useSettingsStore((s) => s.baseUrl)
  const theme = useSettingsStore((s) => s.theme)
  const density = useSettingsStore((s) => s.density)
  const defaultProject = useSettingsStore((s) => s.defaultProject)
  const defaultTable = useSettingsStore((s) => s.defaultTable)
  const setBaseUrl = useSettingsStore((s) => s.setBaseUrl)
  const setTheme = useSettingsStore((s) => s.setTheme)
  const setDensity = useSettingsStore((s) => s.setDensity)
  const setDefaultProject = useSettingsStore((s) => s.setDefaultProject)
  const setDefaultTable = useSettingsStore((s) => s.setDefaultTable)
  const exportConfig = useSettingsStore((s) => s.exportConfig)
  const importConfig = useSettingsStore((s) => s.importConfig)
  const reset = useSettingsStore((s) => s.reset)

  const api = useApi()
  const toast = useToast()
  const fileRef = useRef(null)
  const [probe, setProbe] = useState({ status: 'idle', message: '' })

  async function runProbe() {
    setProbe({ status: 'running', message: '' })
    try {
      const m = await pingHello(api)
      setProbe({ status: 'ok', message: m || 'OK' })
    } catch (e) {
      setProbe({ status: 'fail', message: e.normalizedMessage || e.message })
    }
  }

  function handleExport() {
    const blob = new Blob([exportConfig()], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'mock-sandbox-console.config.json'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    toast.success('Configuration exported')
  }

  function handleImportClick() {
    fileRef.current?.click()
  }

  function handleImportChange(e) {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => {
      const result = importConfig(reader.result)
      if (result.ok) toast.success('Configuration imported')
      else toast.error('Import failed: ' + result.error)
    }
    reader.readAsText(file)
    e.target.value = ''
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-2xl font-semibold tracking-tight">
          Settings
        </h1>
        <p className="mt-1 text-sm text-ink-500">
          Configure the console's connection and appearance. Changes are saved
          to local storage.
        </p>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Backend connection</CardTitle>
          </CardHeader>
          <CardBody className="space-y-4">
            <div>
              <label className="text-xs font-medium text-ink-500">Base URL</label>
              <Input
                value={baseUrl}
                onChange={(e) => setBaseUrl(e.target.value)}
                placeholder="(empty = use Vite proxy)"
              />
              <p className="mt-1 text-xs text-ink-500">
                Leave empty during dev — the Vite proxy forwards{' '}
                <code className="font-mono">/sandbox</code> and{' '}
                <code className="font-mono">/hello</code> to your local Go
                backend at <code className="font-mono">localhost:8080</code>.
                Set a full URL (e.g. <code className="font-mono">https://api.example.com</code>)
                when targeting a remote backend (CORS required).
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="primary" onClick={runProbe}>
                <Wifi className="h-4 w-4" /> Test connection
              </Button>
              {probe.status === 'running' && (
                <span className="text-xs text-ink-500">checking…</span>
              )}
              {probe.status === 'ok' && (
                <span className="inline-flex items-center gap-1 text-xs text-brand-500">
                  <Check className="h-3.5 w-3.5" /> {probe.message}
                </span>
              )}
              {probe.status === 'fail' && (
                <span className="inline-flex items-center gap-1 text-xs text-rose-500">
                  <AlertTriangle className="h-3.5 w-3.5" /> {probe.message}
                </span>
              )}
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs font-medium text-ink-500">
                  Default project
                </label>
                <Input
                  value={defaultProject}
                  onChange={(e) => setDefaultProject(e.target.value)}
                />
              </div>
              <div>
                <label className="text-xs font-medium text-ink-500">
                  Default table
                </label>
                <Input
                  value={defaultTable}
                  onChange={(e) => setDefaultTable(e.target.value)}
                />
              </div>
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Appearance</CardTitle>
          </CardHeader>
          <CardBody className="space-y-4">
            <div>
              <label className="text-xs font-medium text-ink-500">Theme</label>
              <Select value={theme} onChange={(e) => setTheme(e.target.value)}>
                <option value="dark">Dark</option>
                <option value="light">Light</option>
              </Select>
            </div>
            <div>
              <label className="text-xs font-medium text-ink-500">
                Table density
              </label>
              <Select
                value={density}
                onChange={(e) => setDensity(e.target.value)}
              >
                <option value="comfortable">Comfortable</option>
                <option value="compact">Compact</option>
              </Select>
            </div>
          </CardBody>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Export / Import</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="flex flex-wrap items-center gap-2">
              <Button variant="outline" onClick={handleExport}>
                <Download className="h-4 w-4" /> Export config
              </Button>
              <Button variant="outline" onClick={handleImportClick}>
                <Upload className="h-4 w-4" /> Import config
              </Button>
              <Button
                variant="ghost"
                onClick={() => {
                  if (confirm('Reset all settings to defaults?')) {
                    reset()
                    toast.info('Settings reset')
                  }
                }}
              >
                <RotateCcw className="h-4 w-4" /> Reset
              </Button>
              <input
                ref={fileRef}
                type="file"
                accept="application/json"
                onChange={handleImportChange}
                className="hidden"
              />
            </div>
            <p className="mt-2 text-xs text-ink-500">
              Config is stored in your browser's localStorage. Exporting gives
              you a JSON file you can share with teammates.
            </p>
          </CardBody>
        </Card>
      </div>
    </div>
  )
}
