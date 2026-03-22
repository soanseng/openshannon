import { Link, useLocation } from 'react-router-dom'
import { useLanguage } from '../i18n'

const sectionKeys = [
  { path: '/docs/getting-started', key: 'sidebar.gettingStarted' },
  { path: '/docs/commands', key: 'sidebar.commands' },
  { path: '/docs/configuration', key: 'sidebar.configuration' },
  { path: '/docs/google-services', key: 'sidebar.googleServices' },
  { path: '/docs/image-generation', key: 'sidebar.imageGeneration' },
  { path: '/docs/security', key: 'sidebar.security' },
]

export default function DocsSidebar() {
  const location = useLocation()
  const { t } = useLanguage()

  return (
    <aside className="w-64 shrink-0 border-r border-card-border py-8 pr-6 hidden lg:block">
      <div className="sticky top-24">
        <h4 className="font-semibold text-xs uppercase tracking-wider text-navy-light/60 mb-4 px-3">
          {t('sidebar.heading')}
        </h4>
        <nav className="flex flex-col gap-1">
          {sectionKeys.map(({ path, key }) => {
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
                {t(key)}
              </Link>
            )
          })}
        </nav>
      </div>
    </aside>
  )
}
