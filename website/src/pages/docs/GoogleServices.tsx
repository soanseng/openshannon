import CodeBlock from '../../components/CodeBlock'

export default function GoogleServices() {
  return (
    <>
      <h1>Google Services</h1>
      <p>
        OpenShannon integrates with{' '}
        <a href="https://github.com/AarynSmith/gog" target="_blank" rel="noopener noreferrer">
          gog CLI
        </a>{' '}
        to provide access to Google Workspace services directly from Telegram via the{' '}
        <code>/gog</code> command.
      </p>

      <h2>Setup</h2>

      <h3>1. Install gog CLI</h3>
      <p>
        If you used the install wizard (<code>bash install.sh</code>), gog setup is included as
        an optional step. To add it later:
      </p>
      <CodeBlock language="bash">{`make setup-gog`}</CodeBlock>

      <h3>2. Authenticate</h3>
      <p>
        Authenticate your Google account with gog:
      </p>
      <CodeBlock language="bash">{`GOG_KEYRING_PASSWORD="your_password" gog auth add your@gmail.com`}</CodeBlock>
      <p>
        This opens a browser for the OAuth flow. Grant the requested permissions.
      </p>

      <h3>3. Configure Environment</h3>
      <p>
        Add the following to your <code>~/.config/openshannon/env</code> file:
      </p>
      <CodeBlock language="bash">{`GOG_KEYRING_PASSWORD=your_password
GOG_ACCOUNT=your@gmail.com`}</CodeBlock>

      <h2>Gmail</h2>
      <CodeBlock language="text">{`/gog gmail search newer_than:1d              # Recent emails
/gog gmail search from:boss@company.com      # Search by sender
/gog gmail send --to x@y.com --subject "Hi" --body "Hello from OpenShannon"
/gog gmail labels                            # List labels`}</CodeBlock>

      <h2>Calendar</h2>
      <CodeBlock language="text">{`/gog calendar events                         # Today's events
/gog calendar events --days 7                # Next 7 days
/gog calendar create primary --title "Meeting" --start "2026-03-23 15:00" --end "2026-03-23 16:00"
/gog calendar list                           # List calendars`}</CodeBlock>

      <h2>Drive</h2>
      <CodeBlock language="text">{`/gog drive ls                                # List Drive files
/gog drive search "quarterly report"         # Search files
/gog drive info <file-id>                    # File details`}</CodeBlock>

      <h2>Tasks</h2>
      <CodeBlock language="text">{`/gog tasks lists list                        # List task lists
/gog tasks list <list-id>                    # View tasks in a list
/gog tasks create <list-id> --title "Deploy v2"`}</CodeBlock>

      <h2>Contacts</h2>
      <CodeBlock language="text">{`/gog contacts search "John"                  # Search contacts
/gog contacts list                           # List all contacts`}</CodeBlock>

      <h2>Help</h2>
      <p>
        Type <code>/gog</code> without arguments in Telegram to see the full command reference
        for all supported Google services.
      </p>
    </>
  )
}
