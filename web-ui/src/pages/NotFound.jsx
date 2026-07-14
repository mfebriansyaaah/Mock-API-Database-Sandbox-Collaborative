import { Link } from 'react-router-dom'
import { Compass, ArrowLeft } from 'lucide-react'
import Button from '@/components/ui/Button.jsx'

export default function NotFound() {
  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center text-center">
      <div className="font-display text-7xl font-semibold text-brand-500">404</div>
      <div className="mt-2 flex items-center gap-2 text-ink-500">
        <Compass className="h-4 w-4" /> page not found
      </div>
      <p className="mt-3 max-w-sm text-sm text-ink-500">
        The path you tried to reach is not part of the sandbox console.
      </p>
      <Link to="/" className="mt-6">
        <Button variant="primary">
          <ArrowLeft className="h-4 w-4" /> Back to overview
        </Button>
      </Link>
    </div>
  )
}
