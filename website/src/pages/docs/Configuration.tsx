import CodeBlock from '../../components/CodeBlock'

export default function Configuration() {
  return (
    <>
      <h1>Configuration</h1>
      <p>
        OpenShannon uses two configuration files: a YAML config for bot settings and an env
        file for secrets. Both live under <code>~/.config/openshannon/</code>.
      </p>

      <h2>config.yaml</h2>
      <p>
        The main configuration file at <code>~/.config/openshannon/config.yaml</code>:
      </p>
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

      <h2>Environment File</h2>
      <p>
        Secrets go in <code>~/.config/openshannon/env</code> (file mode 600):
      </p>
      <CodeBlock language="bash">{`TELEGRAM_BOT_TOKEN=7123456789:AAHxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Optional integrations:
GEMINI_API_KEY=your_google_ai_api_key
GROQ_API_KEY=gsk_xxxxxxxxxxxxx
GOG_KEYRING_PASSWORD=your_gog_keyring_password
GOG_ACCOUNT=your@gmail.com
NTFY_TOPIC=claude-agent
NTFY_TOKEN=tk_xxxxxxxxxxxxx`}</CodeBlock>
      <p>
        Set the correct file permissions:
      </p>
      <CodeBlock language="bash">{`chmod 600 ~/.config/openshannon/env`}</CodeBlock>

      <h2>Configuration Reference</h2>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>Setting</th>
              <th>Default</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>claude.default_timeout</code></td>
              <td><code>5m</code></td>
              <td>Max time per Claude invocation</td>
            </tr>
            <tr>
              <td><code>claude.long_task_timeout</code></td>
              <td><code>30m</code></td>
              <td>Timeout for <code>/long</code> commands</td>
            </tr>
            <tr>
              <td><code>claude.max_budget_usd</code></td>
              <td><code>10.0</code></td>
              <td>Cost cap per invocation</td>
            </tr>
            <tr>
              <td><code>safety.shell_timeout</code></td>
              <td><code>30s</code></td>
              <td>Max time for <code>/shell</code> commands</td>
            </tr>
            <tr>
              <td><code>streaming.min_interval</code></td>
              <td><code>1s</code></td>
              <td>Min time between Telegram message edits</td>
            </tr>
            <tr>
              <td><code>streaming.max_message_length</code></td>
              <td><code>4096</code></td>
              <td>Telegram message length limit</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>systemd Service</h2>
      <p>
        The install wizard creates a systemd user service at{' '}
        <code>~/.config/systemd/user/openshannon.service</code>.
      </p>
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
      <p>
        Enable lingering so the service runs even when you're not logged in:
      </p>
      <CodeBlock language="bash">{`loginctl enable-linger $(whoami)`}</CodeBlock>

      <h2>Default Workspace</h2>
      <p>
        The install wizard creates <code>~/OpenShannon/</code> as the default workspace. This
        directory contains a <code>CLAUDE.md</code> file with instructions tailored for Telegram
        interaction. New sessions that don't specify a workdir will use this path.
      </p>

      <h2>ntfy Notifications</h2>
      <p>
        OpenShannon can send push notifications via{' '}
        <a href="https://ntfy.sh" target="_blank" rel="noopener noreferrer">ntfy</a> for
        daemon events like startup, crashes, safety blocks, and long task completion. Configure
        the <code>notify</code> section in your config.yaml.
      </p>
    </>
  )
}
