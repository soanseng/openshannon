import CodeBlock from '../../components/CodeBlock'
import { useLanguage } from '../../i18n'

export default function Security() {
  const { t } = useLanguage()

  return (
    <>
      <h1>{t('docs.security.title')}</h1>
      <p>{t('docs.security.intro')}</p>

      <h2>{t('docs.security.dualLayer')}</h2>

      <h3>{t('docs.security.layer1')}</h3>
      <p>{t('docs.security.layer1Desc')}</p>
      <ul>
        <li><strong>{t('docs.security.dangerousCommands')}</strong> <code>rm -rf /</code>, <code>mkfs</code>, <code>dd if=</code>, <code>curl | sh</code></li>
        <li><strong>{t('docs.security.privEscalation')}</strong> <code>sudo</code>, <code>shutdown</code>, <code>reboot</code></li>
        <li><strong>{t('docs.security.destructiveGit')}</strong> <code>git push --force</code></li>
        <li><strong>{t('docs.security.protectedPathsLabel')}</strong> <code>/etc/</code>, <code>/boot/</code>, <code>~/.ssh/authorized_keys</code></li>
      </ul>
      <p>{t('docs.security.cdValidation')}</p>

      <h3>{t('docs.security.layer2')}</h3>
      <p>{t('docs.security.layer2Desc')}</p>
      <CodeBlock language="json">{`{
  "deny": [
    "Bash(sudo *)",
    "Bash(rm -rf /*)",
    "Bash(chmod 777 *)",
    "Bash(curl * | sh)",
    "Bash(wget * | sh)"
  ]
}`}</CodeBlock>

      <h2>{t('docs.security.blockedPatterns')}</h2>
      <p>{t('docs.security.blockedPatternsDesc')}</p>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>{t('docs.security.thCategory')}</th>
              <th>{t('docs.security.thExamples')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>{t('docs.security.catFilesystem')}</td>
              <td><code>rm -rf /</code>, <code>mkfs</code>, <code>dd if=</code></td>
            </tr>
            <tr>
              <td>{t('docs.security.catPrivilege')}</td>
              <td><code>sudo</code>, <code>su -</code></td>
            </tr>
            <tr>
              <td>{t('docs.security.catSystem')}</td>
              <td><code>shutdown</code>, <code>reboot</code>, <code>init</code></td>
            </tr>
            <tr>
              <td>{t('docs.security.catRemoteCode')}</td>
              <td><code>curl | sh</code>, <code>wget | bash</code></td>
            </tr>
            <tr>
              <td>{t('docs.security.catDestructiveGit')}</td>
              <td><code>git push --force</code>, <code>git reset --hard</code></td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>{t('docs.security.protectedPaths')}</h2>
      <p>{t('docs.security.protectedPathsDesc')}</p>
      <ul>
        <li><code>/etc/</code> &mdash; {t('docs.security.pathEtc')}</li>
        <li><code>/boot/</code> &mdash; {t('docs.security.pathBoot')}</li>
        <li><code>/sys/</code> &mdash; {t('docs.security.pathSys')}</li>
        <li><code>/proc/</code> &mdash; {t('docs.security.pathProc')}</li>
        <li><code>~/.ssh/authorized_keys</code> &mdash; {t('docs.security.pathSsh')}</li>
      </ul>

      <h2>{t('docs.security.shellSafety')}</h2>
      <p>{t('docs.security.shellSafetyDesc')}</p>
      <ul>
        <li>{t('docs.security.shellSafety1')}</li>
        <li>30 {t('docs.security.shellSafety2')}</li>
        <li>{t('docs.security.shellSafety3')}</li>
        <li>{t('docs.security.shellSafety4')}</li>
      </ul>

      <h2>{t('docs.security.filePermissions')}</h2>
      <p>{t('docs.security.filePermissionsDesc')}</p>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>{t('docs.security.thFile')}</th>
              <th>{t('docs.security.thMode')}</th>
              <th>{t('docs.security.thContains')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>~/.config/openshannon/config.yaml</code></td>
              <td><code>600</code></td>
              <td>{t('docs.security.fileBotConfig')}</td>
            </tr>
            <tr>
              <td><code>~/.config/openshannon/env</code></td>
              <td><code>600</code></td>
              <td>{t('docs.security.fileApiKeys')}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <h2>{t('docs.security.userAuth')}</h2>
      <p>{t('docs.security.userAuthDesc')}</p>
      <CodeBlock language="yaml">{`telegram:
  allowed_users:
    - 123456789  # Your Telegram user ID`}</CodeBlock>
    </>
  )
}
