import { NavLink, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  FolderTree,
  Terminal,
  Activity,
  Settings,
  Hexagon
} from 'lucide-react'
import { cn } from '@/utils/cn.js'

const nav = [
  { to: '/', label: 'Overview', icon: LayoutDashboard, end: true },
  { to: '/projects', label: 'Projects', icon: FolderTree, end: false },
  { to: '/tester', label: 'REST Tester', icon: Terminal, end: false },
  { to: '/logs', label: 'Access Logs', icon: Activity, end: false },
  { to: '/settings', label: 'Settings', icon: Settings, end: false }
]

export default function Sidebar() {
  const location = useLocation()
  return (
    <aside className="hidden h-full w-60 shrink-0 flex-col border-r border-ink-200 bg-white/60 backdrop-blur md:flex dark:border-ink-800 dark:bg-ink-950/60">
      <div className="flex h-14 items-center gap-2 border-b border-ink-200 px-4 dark:border-ink-800">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand-500/10 text-brand-500">
          <Hexagon className="h-4 w-4" strokeWidth={2.5} />
        </div>
        <div>
          <div className="font-display text-sm font-semibold leading-tight tracking-tight">
            Mock Sandbox
          </div>
          <div className="text-[10px] uppercase tracking-widest text-ink-500">
            console
          </div>
        </div>
      </div>
      <nav className="flex-1 space-y-0.5 p-3">
        {nav.map((item) => {
          const Icon = item.icon
          return (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                cn(
                  'group flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-all',
                  isActive
                    ? 'bg-brand-500/10 text-brand-600 dark:text-brand-400'
                    : 'text-ink-600 hover:bg-ink-100 hover:text-ink-900 dark:text-ink-300 dark:hover:bg-ink-800/60 dark:hover:text-ink-100'
                )
              }
            >
              {({ isActive }) => (
                <>
                  <Icon
                    className={cn(
                      'h-4 w-4 shrink-0',
                      isActive ? 'text-brand-500' : 'text-ink-400 group-hover:text-ink-600'
                    )}
                  />
                  <span className="flex-1">{item.label}</span>
                  {isActive && (
                    <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
                  )}
                </>
              )}
            </NavLink>
          )
        })}
      </nav>
      <div className="border-t border-ink-200 p-3 text-xs text-ink-500 dark:border-ink-800">
        <div className="flex items-center gap-2">
          <span className="glow-dot" />
          <span className="font-mono">/sandbox/*</span>
        </div>
        <div className="mt-1 font-mono text-[10px] text-ink-400">
          {location.pathname}
        </div>
      </div>
    </aside>
  )
}
