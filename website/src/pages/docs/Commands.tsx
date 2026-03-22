import CodeBlock from '../../components/CodeBlock'

export default function Commands() {
  return (
    <>
      <h1>Command Reference</h1>
      <p>
        All commands are sent as Telegram messages. Just type the command in your bot chat or
        in a Forum Topic.
      </p>

      <h2>Full Command Table</h2>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>Command</th>
              <th>Description</th>
              <th>Example</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>/new [workdir]</code></td>
              <td>Create a new Claude Code session</td>
              <td><code>/new ~/infra</code></td>
            </tr>
            <tr>
              <td><code>/resume [id]</code></td>
              <td>Resume an idle session</td>
              <td><code>/resume</code></td>
            </tr>
            <tr>
              <td><code>/sessions</code></td>
              <td>List all active sessions</td>
              <td><code>/sessions</code></td>
            </tr>
            <tr>
              <td><code>/clear</code></td>
              <td>Reset Claude context, keep workdir</td>
              <td><code>/clear</code></td>
            </tr>
            <tr>
              <td><code>/kill [id]</code></td>
              <td>Kill a session completely</td>
              <td><code>/kill</code></td>
            </tr>
            <tr>
              <td><code>/cd &lt;path&gt;</code></td>
              <td>Change the session working directory</td>
              <td><code>/cd ~/apps/feedbot</code></td>
            </tr>
            <tr>
              <td><code>/status</code></td>
              <td>Show daemon status and stats</td>
              <td><code>/status</code></td>
            </tr>
            <tr>
              <td><code>/cancel</code></td>
              <td>Cancel the currently running command</td>
              <td><code>/cancel</code></td>
            </tr>
            <tr>
              <td><code>/shell &lt;cmd&gt;</code></td>
              <td>Run a shell command directly (bypasses Claude)</td>
              <td><code>/shell git status</code></td>
            </tr>
            <tr>
              <td><code>/long &lt;prompt&gt;</code></td>
              <td>Run with extended 30-minute timeout</td>
              <td><code>/long refactor the entire module</code></td>
            </tr>
            <tr>
              <td><code>/model [name]</code></td>
              <td>Switch the model for this session</td>
              <td><code>/model haiku</code></td>
            </tr>
            <tr>
              <td><code>/imagine &lt;desc&gt;</code></td>
              <td>Generate an image via Gemini</td>
              <td><code>/imagine a cat in space</code></td>
            </tr>
            <tr>
              <td><code>/gog &lt;cmd&gt;</code></td>
              <td>Access Google services (Gmail, Calendar, etc.)</td>
              <td><code>/gog gmail search newer_than:1d</code></td>
            </tr>
            <tr>
              <td><code>/help</code></td>
              <td>Show all available commands</td>
              <td><code>/help</code></td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>Session Management</h2>
      <h3>Forum Topics = Sessions</h3>
      <p>
        In a Forum-enabled group, each topic is an isolated session with its own Claude Code
        process and working directory:
      </p>
      <CodeBlock language="text">{`Topic: "infra"        -> workdir: ~/infra
Topic: "feedbot"      -> workdir: ~/apps/feedbot
Topic: "openshannon"  -> workdir: ~/infra/openshannon`}</CodeBlock>
      <p>
        The first message in a new topic auto-creates a session. Use <code>/cd</code> to set the
        working directory.
      </p>

      <h3>Session Lifecycle</h3>
      <ul>
        <li><code>/clear</code> &mdash; Resets Claude context but keeps the workdir and topic binding</li>
        <li><code>/kill</code> &mdash; Removes everything; the topic returns to an unbound state</li>
      </ul>

      <h2>Direct Shell</h2>
      <p>
        <code>/shell</code> bypasses Claude and runs commands directly on the system:
      </p>
      <CodeBlock language="text">{`You: /shell docker ps
Bot: CONTAINER ID  IMAGE         STATUS
     a1b2c3d4      nginx:latest  Up 2 hours`}</CodeBlock>
      <p>
        Shell commands are safety-filtered (no <code>sudo</code>, <code>rm -rf</code>,{' '}
        <code>git push --force</code>, etc.) and have a 30-second timeout.
      </p>

      <h2>Model Switching</h2>
      <p>Each topic/session can use a different model:</p>
      <CodeBlock language="text">{`/model haiku       # Claude Haiku 4.5 (fast, cheap)
/model sonnet      # Claude Sonnet 4.6 (balanced)
/model opus        # Claude Opus 4.6 (most capable)
/model gemini      # Gemini 2.5 Flash
/model gemini-pro  # Gemini 2.5 Pro
/model default     # Reset to config default`}</CodeBlock>

      <h2>Extended Timeout</h2>
      <p>
        Use <code>/long</code> for tasks that need more than the default 5-minute timeout.
        It gives the command a 30-minute window:
      </p>
      <CodeBlock language="text">{`/long refactor the entire authentication module`}</CodeBlock>
    </>
  )
}
