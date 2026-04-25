import CodeBlock from '../../components/CodeBlock'
import { useLanguage } from '../../i18n'

export default function Commands() {
  const { t } = useLanguage()

  return (
    <>
      <h1>{t('docs.commands.title')}</h1>
      <p>{t('docs.commands.intro')}</p>

      <h2>{t('docs.commands.fullTable')}</h2>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>{t('docs.commands.thCommand')}</th>
              <th>{t('docs.commands.thDescription')}</th>
              <th>{t('docs.commands.thExample')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>/new [workdir]</code></td>
              <td>{t('docs.commands.newDesc')}</td>
              <td><code>/new ~/infra</code></td>
            </tr>
            <tr>
              <td><code>/resume [id]</code></td>
              <td>{t('docs.commands.resumeDesc')}</td>
              <td><code>/resume</code></td>
            </tr>
            <tr>
              <td><code>/sessions</code></td>
              <td>{t('docs.commands.sessionsDesc')}</td>
              <td><code>/sessions</code></td>
            </tr>
            <tr>
              <td><code>/clear</code></td>
              <td>{t('docs.commands.clearDesc')}</td>
              <td><code>/clear</code></td>
            </tr>
            <tr>
              <td><code>/kill [id]</code></td>
              <td>{t('docs.commands.killDesc')}</td>
              <td><code>/kill</code></td>
            </tr>
            <tr>
              <td><code>/cd &lt;path&gt;</code></td>
              <td>{t('docs.commands.cdDesc')}</td>
              <td><code>/cd ~/apps/feedbot</code></td>
            </tr>
            <tr>
              <td><code>/status</code></td>
              <td>{t('docs.commands.statusDesc')}</td>
              <td><code>/status</code></td>
            </tr>
            <tr>
              <td><code>/cancel</code></td>
              <td>{t('docs.commands.cancelDesc')}</td>
              <td><code>/cancel</code></td>
            </tr>
            <tr>
              <td><code>/shell &lt;cmd&gt;</code></td>
              <td>{t('docs.commands.shellDesc')}</td>
              <td><code>/shell git status</code></td>
            </tr>
            <tr>
              <td><code>/long &lt;prompt&gt;</code></td>
              <td>{t('docs.commands.longDesc')}</td>
              <td><code>/long refactor the entire module</code></td>
            </tr>
            <tr>
              <td><code>/model [name]</code></td>
              <td>{t('docs.commands.modelDesc')}</td>
              <td><code>/model haiku</code></td>
            </tr>
            <tr>
              <td><code>/agent [name]</code></td>
              <td>{t('docs.commands.agentDesc')}</td>
              <td><code>/agent codex</code></td>
            </tr>
            <tr>
              <td><code>/imagine &lt;desc&gt;</code></td>
              <td>{t('docs.commands.imagineDesc')}</td>
              <td><code>/imagine a cat in space</code></td>
            </tr>
            <tr>
              <td><code>/gog &lt;cmd&gt;</code></td>
              <td>{t('docs.commands.gogDesc')}</td>
              <td><code>/gog gmail search newer_than:1d</code></td>
            </tr>
            <tr>
              <td><code>/help</code></td>
              <td>{t('docs.commands.helpDesc')}</td>
              <td><code>/help</code></td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>{t('docs.commands.sessionManagement')}</h2>
      <h3>{t('docs.commands.forumTopicsSessions')}</h3>
      <p>{t('docs.commands.forumTopicsDesc')}</p>
      <CodeBlock language="text">{`Topic: "infra"        -> workdir: ~/infra
Topic: "feedbot"      -> workdir: ~/apps/feedbot
Topic: "openshannon"  -> workdir: ~/infra/openshannon`}</CodeBlock>
      <p>{t('docs.commands.firstMessageAutoCreates')}</p>

      <h3>{t('docs.commands.sessionLifecycle')}</h3>
      <ul>
        <li><code>/clear</code> &mdash; {t('docs.commands.clearExplain')}</li>
        <li><code>/kill</code> &mdash; {t('docs.commands.killExplain')}</li>
      </ul>

      <h2>{t('docs.commands.directShell')}</h2>
      <p>{t('docs.commands.directShellDesc')}</p>
      <CodeBlock language="text">{`You: /shell docker ps
Bot: CONTAINER ID  IMAGE         STATUS
     a1b2c3d4      nginx:latest  Up 2 hours`}</CodeBlock>
      <p>{t('docs.commands.shellSafety')}</p>

      <h2>{t('docs.commands.modelSwitching')}</h2>
      <p>{t('docs.commands.modelSwitchingDesc')}</p>
      <CodeBlock language="text">{`/model haiku       # Claude Haiku 4.5 (fast, cheap)
/model sonnet      # Claude Sonnet 4.6 (balanced)
/model opus        # Claude Opus 4.6 (most capable)
/model gemini      # Gemini 2.5 Flash
/model gemini-pro  # Gemini 2.5 Pro
/model default     # Reset to config default`}</CodeBlock>

      <h2>{t('docs.commands.agentSwitching')}</h2>
      <p>{t('docs.commands.agentSwitchingDesc')}</p>
<CodeBlock language="text">{`/agent claude   # Use Claude Code for this session
/agent codex    # Use Codex CLI for this session
/agent default  # Reset to Claude Code`}</CodeBlock>

      <h2>{t('docs.commands.extendedTimeout')}</h2>
      <p>{t('docs.commands.extendedTimeoutDesc')}</p>
      <CodeBlock language="text">{`/long refactor the entire authentication module`}</CodeBlock>
    </>
  )
}
