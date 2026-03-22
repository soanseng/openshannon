import CodeBlock from '../../components/CodeBlock'

export default function GettingStarted() {
  return (
    <>
      <h1>Getting Started</h1>
      <p>
        OpenShannon is a Go daemon that bridges Telegram to Claude Code. It lets you control
        a Claude Code agent from your phone via Telegram. Each Forum Topic maps to an isolated
        session with its own working directory and conversation context.
      </p>

      <h2>Prerequisites</h2>
      <ul>
        <li><strong>Go 1.22+</strong> &mdash; <a href="https://go.dev/dl/" target="_blank" rel="noopener noreferrer">Download Go</a></li>
        <li><strong>Claude Code CLI</strong> &mdash; installed and authenticated (<code>claude --version</code>)</li>
        <li><strong>Telegram Bot</strong> &mdash; create one via <a href="https://t.me/BotFather" target="_blank" rel="noopener noreferrer">@BotFather</a></li>
        <li><strong>Your Telegram User ID</strong> &mdash; get it from <a href="https://t.me/userinfobot" target="_blank" rel="noopener noreferrer">@userinfobot</a></li>
        <li>(Optional) <strong>Groq API key</strong> for voice note transcription</li>
        <li>(Optional) <strong>Gemini API key</strong> for image generation</li>
      </ul>

      <h2>1. Create a Telegram Bot</h2>
      <ol>
        <li>Open <a href="https://t.me/BotFather" target="_blank" rel="noopener noreferrer">@BotFather</a> in Telegram</li>
        <li>Send <code>/newbot</code> and follow the prompts</li>
        <li>Copy the bot token (looks like <code>7123456789:AAH...</code>)</li>
        <li>Send <code>/setprivacy</code>, select your bot, then choose <code>Disable</code> so the bot can read group messages</li>
        <li>(Optional) Send <code>/setcommands</code> and paste the command list from the README</li>
      </ol>

      <h2>2. Set Up a Forum Group</h2>
      <p>
        Forum-enabled groups give you the best experience. Each topic becomes an isolated Claude Code session.
      </p>
      <ol>
        <li>Create a new Telegram Group</li>
        <li>Go to Group Settings &rarr; Topics &rarr; Enable</li>
        <li>Add your bot to the group</li>
        <li>Make the bot an admin (needed for topic access)</li>
        <li>Create topics for your projects: &ldquo;infra&rdquo;, &ldquo;feedbot&rdquo;, etc.</li>
      </ol>

      <h2>3. Install</h2>
      <CodeBlock language="bash">{`git clone https://github.com/soanseng/openshannon.git
cd openshannon

# Interactive setup wizard (recommended)
bash install.sh

# Or non-interactive:
make setup`}</CodeBlock>

      <p>The install wizard guides you through:</p>
      <ol>
        <li><strong>Build</strong> &mdash; compiles the Go binary</li>
        <li><strong>Telegram</strong> &mdash; bot token + user ID setup</li>
        <li><strong>Gemini</strong> &mdash; (optional) API key for <code>/imagine</code></li>
        <li><strong>Google Services</strong> &mdash; (optional) gog CLI authentication</li>
        <li><strong>Config</strong> &mdash; writes config files with correct permissions</li>
        <li><strong>Workspace</strong> &mdash; creates <code>~/OpenShannon/</code> with CLAUDE.md and systemd service</li>
      </ol>

      <h3>Files Created</h3>
      <ul>
        <li><code>~/.config/openshannon/config.yaml</code> &mdash; bot config (mode 600)</li>
        <li><code>~/.config/openshannon/env</code> &mdash; secrets (mode 600)</li>
        <li><code>~/OpenShannon/</code> &mdash; default workspace with git</li>
        <li><code>~/OpenShannon/CLAUDE.md</code> &mdash; Claude instructions for Telegram use</li>
        <li><code>~/.config/systemd/user/openshannon.service</code> &mdash; systemd unit</li>
      </ul>

      <h2>4. Test Run</h2>
      <CodeBlock language="bash">{`# Run in foreground to verify
cd ~/infra/openshannon
make run`}</CodeBlock>
      <p>
        Open Telegram and send <code>/status</code> to your bot. You should see a status message
        with uptime and version info.
      </p>

      <h2>5. Deploy as a Service</h2>
      <CodeBlock language="bash">{`# Enable and start the systemd user service
make start

# Verify
make status
make logs`}</CodeBlock>
      <p>
        Enable lingering so the service keeps running without a login session:
      </p>
      <CodeBlock language="bash">{`loginctl enable-linger $(whoami)`}</CodeBlock>

      <h2>First Message</h2>
      <p>
        Once the bot is running, send any text message to your Telegram bot (or to a topic in
        your Forum group). The message goes straight to Claude Code as a prompt:
      </p>
      <CodeBlock language="text">{`You: help me find all TODO comments in the codebase
Bot: (processing...)
Bot: I found 12 TODO comments across 5 files...`}</CodeBlock>
    </>
  )
}
