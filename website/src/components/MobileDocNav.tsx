import { Link, useLocation } from 'react-router-dom'
import { useState } from 'react'
import { useLanguage } from '../i18n'

const sectionKeys = [
  { path: '/docs/getting-started', key: 'sidebar.gettingStarted' },
  { path: '/docs/commands', key: 'sidebar.commands' },
  { path: '/docs/configuration', key: 'sidebar.configuration' },
  { path: '/docs/google-services', key: 'sidebar.googleServices' },
  { path: '/docs/image-generation', key: 'sidebar.imageGeneration' },
  { path: '/docs/security', key: 'sidebar.security' },
]

export default function MobileDocNav() {
  const location = useLocation()
  const [open, setOpen] = useState(false)
  const { t } = useLanguage()
  const current = sectionKeys.find(s => s.path === location.pathname) ?? sectionKeys[0]

  return (
    <div className="lg:hidden border-b border-card-border bg-cream">
      <div className="max-w-6xl mx-auto px-6">
        <button
          onClick={() => setOpen(!open)}
          className="w-full py-3 flex items-center justify-between text-sm font-medium text-navy"
        >
          <span>{t(current.key)}</span>
          <svg
            className={`w-4 h-4 transition-transform ${open ? 'rotate-180' : ''}`}
            fill="none" stroke="currentColor" viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        {open && (
          <div className="pb-3 flex flex-col gap-1">
            {sectionKeys.map(({ path, key }) => {
              const isActive = location.pathname === path ||
                (path === '/docs/getting-started' && location.pathname === '/docs')
              return (
                <Link
                  key={path}
                  to={path}
                  onClick={() => setOpen(false)}
                  className={`px-3 py-2 rounded-lg text-sm no-underline ${
                    isActive
                      ? 'bg-accent-light text-accent font-medium'
                      : 'text-navy-light hover:bg-cream-dark'
                  }`}
                >
                  {t(key)}
                </Link>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
