import CodeBlock from '../../components/CodeBlock'

export default function Security() {
  return (
    <>
      <h1>Security</h1>
      <p>
        OpenShannon implements dual-layer protection to prevent dangerous operations, whether
        initiated by the user or by Claude Code itself.
      </p>

      <h2>Dual-Layer Protection</h2>

      <h3>Layer 1: Go Daemon Blocklist</h3>
      <p>
        Before Claude ever sees a prompt, the Go daemon filters it through a blocklist. This
        catches dangerous patterns at the entry point:
      </p>
      <ul>
        <li><strong>Dangerous commands:</strong> <code>rm -rf /</code>, <code>mkfs</code>, <code>dd if=</code>, <code>curl | sh</code></li>
        <li><strong>Privilege escalation:</strong> <code>sudo</code>, <code>shutdown</code>, <code>reboot</code></li>
        <li><strong>Destructive git:</strong> <code>git push --force</code></li>
        <li><strong>Protected paths:</strong> <code>/etc/</code>, <code>/boot/</code>, <code>~/.ssh/authorized_keys</code></li>
      </ul>
      <p>
        The <code>/cd</code> command also validates paths &mdash; you cannot change the working
        directory to protected system paths.
      </p>

      <h3>Layer 2: Claude Code Deny List</h3>
      <p>
        Your existing Claude Code <code>settings.json</code> provides a second layer. This blocks
        tool executions that bypass the text filter:
      </p>
      <CodeBlock language="json">{`{
  "deny": [
    "Bash(sudo *)",
    "Bash(rm -rf /*)",
    "Bash(chmod 777 *)",
    "Bash(curl * | sh)",
    "Bash(wget * | sh)"
  ]
}`}</CodeBlock>

      <h2>Blocked Patterns</h2>
      <p>
        The Go daemon blocklist catches these categories:
      </p>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>Category</th>
              <th>Examples</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>Filesystem destruction</td>
              <td><code>rm -rf /</code>, <code>mkfs</code>, <code>dd if=</code></td>
            </tr>
            <tr>
              <td>Privilege escalation</td>
              <td><code>sudo</code>, <code>su -</code></td>
            </tr>
            <tr>
              <td>System control</td>
              <td><code>shutdown</code>, <code>reboot</code>, <code>init</code></td>
            </tr>
            <tr>
              <td>Remote code execution</td>
              <td><code>curl | sh</code>, <code>wget | bash</code></td>
            </tr>
            <tr>
              <td>Destructive git</td>
              <td><code>git push --force</code>, <code>git reset --hard</code></td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>Protected Paths</h2>
      <p>
        The daemon prevents <code>/cd</code> and file operations in sensitive directories:
      </p>
      <ul>
        <li><code>/etc/</code> &mdash; System configuration</li>
        <li><code>/boot/</code> &mdash; Boot loader files</li>
        <li><code>/sys/</code> &mdash; Kernel parameters</li>
        <li><code>/proc/</code> &mdash; Process information</li>
        <li><code>~/.ssh/authorized_keys</code> &mdash; SSH access control</li>
      </ul>

      <h2>Shell Command Safety</h2>
      <p>
        Commands run via <code>/shell</code> have additional protections:
      </p>
      <ul>
        <li>Same blocklist filtering as regular prompts</li>
        <li>30-second timeout (configurable via <code>safety.shell_timeout</code>)</li>
        <li>Output is truncated to fit Telegram message limits</li>
        <li>No interactive commands (stdin is closed)</li>
      </ul>

      <h2>File Permissions</h2>
      <p>
        The install wizard sets restrictive permissions on sensitive files:
      </p>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>File</th>
              <th>Mode</th>
              <th>Contains</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>~/.config/openshannon/config.yaml</code></td>
              <td><code>600</code></td>
              <td>Bot configuration</td>
            </tr>
            <tr>
              <td><code>~/.config/openshannon/env</code></td>
              <td><code>600</code></td>
              <td>API keys and tokens</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>User Authorization</h2>
      <p>
        Only Telegram user IDs listed in <code>config.yaml</code> under{' '}
        <code>telegram.allowed_users</code> can interact with the bot. All messages from
        unauthorized users are silently ignored.
      </p>
      <CodeBlock language="yaml">{`telegram:
  allowed_users:
    - 123456789  # Your Telegram user ID`}</CodeBlock>
    </>
  )
}
