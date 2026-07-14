import { NavLink, Outlet, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  FolderTree,
  Terminal,
  Activity,
  Settings
} from 'lucide-react'
import Sidebar from './Sidebar.jsx'
import Topbar from './Topbar.jsx'
import { cn } from '@/utils/cn.js'

const mobileNav = [
  { to: '/', label: 'Home', icon: LayoutDashboard, end: true },
  { to: '/projects', label: 'Projects', icon: FolderTree, end: false },
  { to: '/tester', label: 'Tester', icon: Terminal, end: false },
  { to: '/logs', label: 'Logs', icon: Activity, end: false },
  { to: '/settings', label: 'Settings', icon: Settings, end: false }
]

export default function AppShell() {
  const location = useLocation()
  return (
    <div className="app-grid flex h-screen w-screen overflow-hidden bg-ink-50 text-ink-900 dark:bg-ink-950 dark:text-ink-100">
      <Sidebar />
      <div className="flex h-full flex-1 flex-col overflow-hidden">
        <Topbar />
        <main
          key={location.pathname}
          className="flex-1 overflow-auto px-4 py-6 animate-fade-in-up md:px-8 md:py-8"
        >
          <div className="mx-auto w-full max-w-7xl">
            <Outlet />
          </div>
        </main>
        <nav className="sticky bottom-0 z-30 flex h-14 items-center justify-around border-t border-ink-200 bg-white/90 backdrop-blur md:hidden dark:border-ink-800 dark:bg-ink-950/90">
          {mobileNav.map((item) => {
            const Icon = item.icon
            return (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end}
                className={({ isActive }) =>
                  cn(
                    'flex flex-1 flex-col items-center justify-center gap-1 py-2 text-[10px] font-medium transition',
                    isActive
                      ? 'text-brand-500'
                      : 'text-ink-500 hover:text-ink-900 dark:hover:text-ink-100'
                  )
                }
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </NavLink>
            )
          })}
        </nav>
      </div>
    </div>
  )
}
