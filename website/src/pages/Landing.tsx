import { Link } from 'react-router-dom'
import Navbar from '../components/Navbar'
import Footer from '../components/Footer'
import FeatureCard from '../components/FeatureCard'
import CodeBlock from '../components/CodeBlock'
import { useLanguage } from '../i18n'

const featureKeys = [
  { icon: '\u{1F4AC}', key: 'chat' },
  { icon: '\u{1F9F5}', key: 'sessions' },
  { icon: '\u{1F504}', key: 'multiModel' },
  { icon: '\u{1F3A8}', key: 'imageGen' },
  { icon: '\u{1F4E7}', key: 'google' },
  { icon: '\u{1F6E1}\uFE0F', key: 'safety' },
]

const quickStart = `git clone https://github.com/soanseng/openshannon.git
cd openshannon && bash install.sh
make start`

export default function Landing() {
  const { t } = useLanguage()

  const features = featureKeys.map((f) => ({
    icon: f.icon,
    title: t(`features.${f.key}.title`),
    description: t(`features.${f.key}.desc`),
  }))

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />

      {/* Hero */}
      <section className="dot-pattern">
        <div className="max-w-4xl mx-auto px-6 py-24 md:py-32 text-center">
          <img
            src="/shannon.jpg"
            alt="OpenShannon"
            className="w-32 h-32 md:w-40 md:h-40 rounded-2xl mx-auto mb-8 shadow-lg"
          />
          <h1 className="text-4xl md:text-6xl font-bold text-navy tracking-tight mb-4">
            {t('hero.title')}
          </h1>
          <p className="text-xl md:text-2xl text-navy-light mb-10 max-w-2xl mx-auto leading-relaxed">
            {t('hero.tagline')}
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              to="/docs/getting-started"
              className="inline-flex items-center justify-center gap-2 bg-accent hover:bg-accent-hover text-white font-medium px-8 py-3 rounded-lg transition-colors no-underline"
            >
              {t('hero.getStarted')}
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </Link>
            <a
              href="https://github.com/soanseng/openshannon"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center justify-center gap-2 bg-white border border-card-border hover:border-accent/40 text-navy font-medium px-8 py-3 rounded-lg transition-colors no-underline"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z" />
              </svg>
              {t('hero.viewGithub')}
            </a>
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="bg-cream-dark/40 border-y border-card-border">
        <div className="max-w-6xl mx-auto px-6 py-20">
          <h2 className="text-3xl font-bold text-center text-navy mb-12">
            {t('features.heading')}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((f) => (
              <FeatureCard key={f.title} {...f} />
            ))}
          </div>
        </div>
      </section>

      {/* Quick Start */}
      <section className="max-w-6xl mx-auto px-6 py-20">
        <div className="max-w-2xl mx-auto">
          <h2 className="text-3xl font-bold text-center text-navy mb-3">
            {t('quickStart.heading')}
          </h2>
          <p className="text-center text-navy-light mb-8">
            {t('quickStart.subheading')}
          </p>
          <CodeBlock language="bash">{quickStart}</CodeBlock>
        </div>
      </section>

      {/* Screenshots placeholder */}
      <section className="bg-cream-dark/40 border-y border-card-border">
        <div className="max-w-6xl mx-auto px-6 py-20">
          <h2 className="text-3xl font-bold text-center text-navy mb-12">
            {t('screenshots.heading')}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="bg-cream-dark border border-card-border rounded-xl h-56 flex items-center justify-center"
              >
                <span className="text-navy-light/50 text-sm font-medium">
                  {t('screenshots.comingSoon')}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <Footer />
    </div>
  )
}
