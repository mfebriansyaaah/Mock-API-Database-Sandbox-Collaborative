import { Routes, Route, Navigate } from 'react-router-dom'
import AppShell from '@/components/layout/AppShell.jsx'
import Overview from '@/pages/Overview.jsx'
import Projects from '@/pages/Projects.jsx'
import Tables from '@/pages/Tables.jsx'
import Documents from '@/pages/Documents.jsx'
import DocumentDetail from '@/pages/DocumentDetail.jsx'
import RESTTester from '@/pages/RESTTester.jsx'
import ApiKeys from '@/pages/ApiKeys.jsx'
import AccessLogs from '@/pages/AccessLogs.jsx'
import Settings from '@/pages/Settings.jsx'
import NotFound from '@/pages/NotFound.jsx'
import { ToastProvider } from '@/components/ui/Toast.jsx'
import { useThemeSync } from '@/stores/useSettingsStore.js'

function ThemeSync() {
  useThemeSync()
  return null
}

export default function App() {
  return (
    <ToastProvider>
      <ThemeSync />
      <Routes>
        <Route element={<AppShell />}>
          <Route path="/" element={<Overview />} />
          <Route path="/projects" element={<Projects />} />
          <Route path="/projects/:projectId" element={<Tables />} />
          <Route
            path="/projects/:projectId/:table"
            element={<Documents />}
          />
          <Route
            path="/projects/:projectId/:table/:docId"
            element={<DocumentDetail />}
          />
          <Route path="/tester" element={<RESTTester />} />
          <Route path="/api-keys" element={<ApiKeys />} />
          <Route path="/logs" element={<AccessLogs />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/404" element={<NotFound />} />
          <Route path="*" element={<Navigate to="/404" replace />} />
        </Route>
      </Routes>
    </ToastProvider>
  )
}
