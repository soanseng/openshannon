import CodeBlock from '../../components/CodeBlock'
import { useLanguage } from '../../i18n'

export default function Configuration() {
  const { t } = useLanguage()

  return (
    <>
      <h1>{t('docs.configuration.title')}</h1>
      <p>{t('docs.configuration.intro')}</p>

      <h2>{t('docs.configuration.configYaml')}</h2>
      <p>{t('docs.configuration.configYamlDesc')}</p>
      <CodeBlock language="yaml">{`telegram:
  token: "\${TELEGRAM_BOT_TOKEN}"
  allowed_users:
    - YOUR_TELEGRAM_USER_ID    # Replace with your numeric ID

claude:
  default_timeout: 5m          # Max time per Claude invocation
  long_task_timeout: 30m       # Timeout for /long commands
  max_budget_usd: 10.0         # Cost cap per invocation

safety:
  shell_timeout: 30s           # Max time for /shell commands

streaming:
  min_interval: 1s             # Min time between message edits
  max_message_length: 4096     # Telegram message length limit

notify:
  enabled: true
  ntfy_server: "https://ntfy.sh"
  ntfy_topic: "claude-agent"
  events:
    - daemon_start
    - daemon_crash
    - safety_block
    - long_task_complete`}</CodeBlock>

      <h2>{t('docs.configuration.envFile')}</h2>
      <p>{t('docs.configuration.envFileDesc')}</p>
      <CodeBlock language="bash">{`TELEGRAM_BOT_TOKEN=7123456789:AAHxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Optional integrations:
GEMINI_API_KEY=your_google_ai_api_key
GROQ_API_KEY=gsk_xxxxxxxxxxxxx
GOG_KEYRING_PASSWORD=your_gog_keyring_password
GOG_ACCOUNT=your@gmail.com
NTFY_TOPIC=claude-agent
NTFY_TOKEN=tk_xxxxxxxxxxxxx`}</CodeBlock>
      <p>{t('docs.configuration.envPermissions')}</p>
      <CodeBlock language="bash">{`chmod 600 ~/.config/openshannon/env`}</CodeBlock>

      <h2>{t('docs.configuration.configRef')}</h2>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>{t('docs.configuration.thSetting')}</th>
              <th>{t('docs.configuration.thDefault')}</th>
              <th>{t('docs.configuration.thDescription')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>claude.default_timeout</code></td>
              <td><code>5m</code></td>
              <td>{t('docs.configuration.descDefaultTimeout')}</td>
            </tr>
            <tr>
              <td><code>claude.long_task_timeout</code></td>
              <td><code>30m</code></td>
              <td>{t('docs.configuration.descLongTimeout')}</td>
            </tr>
            <tr>
              <td><code>claude.max_budget_usd</code></td>
              <td><code>10.0</code></td>
              <td>{t('docs.configuration.descBudget')}</td>
            </tr>
            <tr>
              <td><code>safety.shell_timeout</code></td>
              <td><code>30s</code></td>
              <td>{t('docs.configuration.descShellTimeout')}</td>
            </tr>
            <tr>
              <td><code>streaming.min_interval</code></td>
              <td><code>1s</code></td>
              <td>{t('docs.configuration.descMinInterval')}</td>
            </tr>
            <tr>
              <td><code>streaming.max_message_length</code></td>
              <td><code>4096</code></td>
              <td>{t('docs.configuration.descMaxLength')}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>{t('docs.configuration.systemd')}</h2>
      <p>{t('docs.configuration.systemdDesc')}</p>
      <CodeBlock language="bash">{`# Start the service
make start

# Check status
make status

# View logs
make logs

# Stop the service
systemctl --user stop openshannon

# Restart
systemctl --user restart openshannon`}</CodeBlock>
      <p>{t('docs.configuration.systemdLinger')}</p>
      <CodeBlock language="bash">{`loginctl enable-linger $(whoami)`}</CodeBlock>

      <h2>{t('docs.configuration.defaultWorkspace')}</h2>
      <p>{t('docs.configuration.defaultWorkspaceDesc')}</p>

      <h2>{t('docs.configuration.ntfyNotifications')}</h2>
      <p>
        {t('docs.configuration.ntfyDescBefore')}{' '}
        <a href="https://ntfy.sh" target="_blank" rel="noopener noreferrer">ntfy</a>{' '}
        {t('docs.configuration.ntfyDescAfter')}
      </p>
    </>
  )
}
