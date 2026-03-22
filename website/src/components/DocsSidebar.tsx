import { Link, useLocation } from 'react-router-dom'

const sections = [
  { path: '/docs/getting-started', label: 'Getting Started' },
  { path: '/docs/commands', label: 'Commands' },
  { path: '/docs/configuration', label: 'Configuration' },
  { path: '/docs/google-services', label: 'Google Services' },
  { path: '/docs/image-generation', label: 'Image Generation' },
  { path: '/docs/security', label: 'Security' },
]

export default function DocsSidebar() {
  const location = useLocation()

  return (
    <aside className="w-64 shrink-0 border-r border-card-border py-8 pr-6 hidden lg:block">
      <div className="sticky top-24">
        <h4 className="font-semibold text-xs uppercase tracking-wider text-navy-light/60 mb-4 px-3">
          Documentation
        </h4>
        <nav className="flex flex-col gap-1">
          {sections.map(({ path, label }) => {
            const isActive = location.pathname === path ||
              (path === '/docs/getting-started' && location.pathname === '/docs')
            return (
              <Link
                key={path}
                to={path}
                className={`px-3 py-2 rounded-lg text-sm transition-colors no-underline ${
                  isActive
                    ? 'bg-accent-light text-accent font-medium'
                    : 'text-navy-light hover:bg-cream-dark hover:text-navy'
                }`}
              >
                {label}
              </Link>
            )
          })}
        </nav>
      </div>
    </aside>
  )
}
