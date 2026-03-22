import CodeBlock from '../../components/CodeBlock'
import { useLanguage } from '../../i18n'

export default function GoogleServices() {
  const { t } = useLanguage()

  return (
    <>
      <h1>{t('docs.googleServices.title')}</h1>
      <p>
        {t('docs.googleServices.intro')}{' '}
        <a href="https://github.com/AarynSmith/gog" target="_blank" rel="noopener noreferrer">
          {t('docs.googleServices.introGog')}
        </a>
        {t('docs.googleServices.introSuffix')}
      </p>

      <h2>{t('docs.googleServices.setup')}</h2>

      <h3>{t('docs.googleServices.installGog')}</h3>
      <p>{t('docs.googleServices.installGogDesc')}</p>
      <CodeBlock language="bash">{`make setup-gog`}</CodeBlock>

      <h3>{t('docs.googleServices.authenticate')}</h3>
      <p>{t('docs.googleServices.authenticateDesc')}</p>
      <CodeBlock language="bash">{`GOG_KEYRING_PASSWORD="your_password" gog auth add your@gmail.com`}</CodeBlock>
      <p>{t('docs.googleServices.authenticateNote')}</p>

      <h3>{t('docs.googleServices.configureEnv')}</h3>
      <p>{t('docs.googleServices.configureEnvDesc')}</p>
      <CodeBlock language="bash">{`GOG_KEYRING_PASSWORD=your_password
GOG_ACCOUNT=your@gmail.com`}</CodeBlock>

      <h2>{t('docs.googleServices.gmail')}</h2>
      <CodeBlock language="text">{`/gog gmail search newer_than:1d              # Recent emails
/gog gmail search from:boss@company.com      # Search by sender
/gog gmail send --to x@y.com --subject "Hi" --body "Hello from OpenShannon"
/gog gmail labels                            # List labels`}</CodeBlock>

      <h2>{t('docs.googleServices.calendar')}</h2>
      <CodeBlock language="text">{`/gog calendar events                         # Today's events
/gog calendar events --days 7                # Next 7 days
/gog calendar create primary --title "Meeting" --start "2026-03-23 15:00" --end "2026-03-23 16:00"
/gog calendar list                           # List calendars`}</CodeBlock>

      <h2>{t('docs.googleServices.drive')}</h2>
      <CodeBlock language="text">{`/gog drive ls                                # List Drive files
/gog drive search "quarterly report"         # Search files
/gog drive info <file-id>                    # File details`}</CodeBlock>

      <h2>{t('docs.googleServices.tasks')}</h2>
      <CodeBlock language="text">{`/gog tasks lists list                        # List task lists
/gog tasks list <list-id>                    # View tasks in a list
/gog tasks create <list-id> --title "Deploy v2"`}</CodeBlock>

      <h2>{t('docs.googleServices.contacts')}</h2>
      <CodeBlock language="text">{`/gog contacts search "John"                  # Search contacts
/gog contacts list                           # List all contacts`}</CodeBlock>

      <h2>{t('docs.googleServices.help')}</h2>
      <p>{t('docs.googleServices.helpDesc')}</p>
    </>
  )
}
