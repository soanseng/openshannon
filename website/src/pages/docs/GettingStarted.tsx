import CodeBlock from '../../components/CodeBlock'
import { useLanguage } from '../../i18n'

export default function GettingStarted() {
  const { t } = useLanguage()

  return (
    <>
      <h1>{t('docs.gettingStarted.title')}</h1>
      <p>{t('docs.gettingStarted.intro')}</p>

      <h2>{t('docs.gettingStarted.prerequisites')}</h2>
      <ul>
        <li><strong>{t('docs.gettingStarted.prereqGo')}</strong> &mdash; <a href="https://go.dev/dl/" target="_blank" rel="noopener noreferrer">{t('docs.gettingStarted.prereqGoLink')}</a></li>
        <li><strong>{t('docs.gettingStarted.prereqClaude')}</strong> &mdash; {t('docs.gettingStarted.prereqClaudeDesc')} (<code>claude --version</code>)</li>
        <li><strong>{t('docs.gettingStarted.prereqBot')}</strong> &mdash; {t('docs.gettingStarted.prereqBotDesc')} <a href="https://t.me/BotFather" target="_blank" rel="noopener noreferrer">@BotFather</a></li>
        <li><strong>{t('docs.gettingStarted.prereqUserId')}</strong> &mdash; {t('docs.gettingStarted.prereqUserIdDesc')} <a href="https://t.me/userinfobot" target="_blank" rel="noopener noreferrer">@userinfobot</a></li>
        <li>{t('docs.gettingStarted.optional')} <strong>{t('docs.gettingStarted.prereqGroq')}</strong> {t('docs.gettingStarted.prereqGroqDesc')}</li>
        <li>{t('docs.gettingStarted.optional')} <strong>{t('docs.gettingStarted.prereqGemini')}</strong> {t('docs.gettingStarted.prereqGeminiDesc')}</li>
      </ul>

      <h2>{t('docs.gettingStarted.createBot')}</h2>
      <ol>
        <li>{t('docs.gettingStarted.createBotStep1')}</li>
        <li>{t('docs.gettingStarted.createBotStep2')}</li>
        <li>{t('docs.gettingStarted.createBotStep3')}</li>
        <li>{t('docs.gettingStarted.createBotStep4')}</li>
        <li>{t('docs.gettingStarted.createBotStep5')}</li>
      </ol>

      <h2>{t('docs.gettingStarted.setupForum')}</h2>
      <p>{t('docs.gettingStarted.setupForumDesc')}</p>
      <ol>
        <li>{t('docs.gettingStarted.setupForumStep1')}</li>
        <li>{t('docs.gettingStarted.setupForumStep2')}</li>
        <li>{t('docs.gettingStarted.setupForumStep3')}</li>
        <li>{t('docs.gettingStarted.setupForumStep4')}</li>
        <li>{t('docs.gettingStarted.setupForumStep5')}</li>
      </ol>

      <h2>{t('docs.gettingStarted.install')}</h2>
      <CodeBlock language="bash">{`git clone https://github.com/soanseng/openshannon.git
cd openshannon

# Interactive setup wizard (recommended)
bash install.sh

# Or non-interactive:
make setup`}</CodeBlock>

      <p>{t('docs.gettingStarted.installWizardIntro')}</p>
      <ol>
        <li><strong>{t('docs.gettingStarted.installStep1')}</strong> &mdash; {t('docs.gettingStarted.installStep1Desc')}</li>
        <li><strong>{t('docs.gettingStarted.installStep2')}</strong> &mdash; {t('docs.gettingStarted.installStep2Desc')}</li>
        <li><strong>{t('docs.gettingStarted.installStep3')}</strong> &mdash; {t('docs.gettingStarted.installStep3Desc')}</li>
        <li><strong>{t('docs.gettingStarted.installStep4')}</strong> &mdash; {t('docs.gettingStarted.installStep4Desc')}</li>
        <li><strong>{t('docs.gettingStarted.installStep5')}</strong> &mdash; {t('docs.gettingStarted.installStep5Desc')}</li>
        <li><strong>{t('docs.gettingStarted.installStep6')}</strong> &mdash; {t('docs.gettingStarted.installStep6Desc')}</li>
      </ol>

      <h3>{t('docs.gettingStarted.filesCreated')}</h3>
      <ul>
        <li><code>~/.config/openshannon/config.yaml</code> &mdash; {t('docs.gettingStarted.fileConfig')}</li>
        <li><code>~/.config/openshannon/env</code> &mdash; {t('docs.gettingStarted.fileEnv')}</li>
        <li><code>~/OpenShannon/</code> &mdash; {t('docs.gettingStarted.fileWorkspace')}</li>
        <li><code>~/OpenShannon/CLAUDE.md</code> &mdash; {t('docs.gettingStarted.fileClaudeMd')}</li>
        <li><code>~/.config/systemd/user/openshannon.service</code> &mdash; {t('docs.gettingStarted.fileSystemd')}</li>
      </ul>

      <h2>{t('docs.gettingStarted.testRun')}</h2>
      <CodeBlock language="bash">{`# Run in foreground to verify
cd ~/infra/openshannon
make run`}</CodeBlock>
      <p>{t('docs.gettingStarted.testRunDesc')}</p>

      <h2>{t('docs.gettingStarted.deploy')}</h2>
      <CodeBlock language="bash">{`# Enable and start the systemd user service
make start

# Verify
make status
make logs`}</CodeBlock>
      <p>{t('docs.gettingStarted.deployLinger')}</p>
      <CodeBlock language="bash">{`loginctl enable-linger $(whoami)`}</CodeBlock>

      <h2>{t('docs.gettingStarted.firstMessage')}</h2>
      <p>{t('docs.gettingStarted.firstMessageDesc')}</p>
      <CodeBlock language="text">{`You: help me find all TODO comments in the codebase
Bot: (processing...)
Bot: I found 12 TODO comments across 5 files...`}</CodeBlock>
    </>
  )
}
