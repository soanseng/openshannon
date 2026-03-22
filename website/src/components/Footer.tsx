import { useLanguage } from '../i18n'

export default function Footer() {
  const { t } = useLanguage()

  return (
    <footer className="border-t border-card-border bg-cream-dark/50">
      <div className="max-w-6xl mx-auto px-6 py-8 flex flex-col items-center gap-4 text-sm text-navy-light">
        <div className="flex flex-col sm:flex-row items-center justify-between w-full gap-4">
          <div className="flex items-center gap-2">
            <img src="/shannon.jpg" alt="" className="w-5 h-5 rounded" />
            <span>OpenShannon</span>
            <span className="text-card-border mx-1">|</span>
            <span>{t('footer.license')}</span>
          </div>
          <div className="flex items-center gap-4">
            <a
              href="https://github.com/soanseng/openshannon"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-accent transition-colors"
            >
              GitHub
            </a>
            <span className="text-card-border">|</span>
            <span>openshannon.org</span>
          </div>
        </div>
        <p className="text-xs text-navy-light/60 text-center">
          {t('footer.disclaimer')}
        </p>
      </div>
    </footer>
  )
}
