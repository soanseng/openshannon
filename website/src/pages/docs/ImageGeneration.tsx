import CodeBlock from '../../components/CodeBlock'
import { useLanguage } from '../../i18n'

export default function ImageGeneration() {
  const { t } = useLanguage()

  return (
    <>
      <h1>{t('docs.imageGeneration.title')}</h1>
      <p>{t('docs.imageGeneration.intro')}</p>

      <h2>{t('docs.imageGeneration.howItWorks')}</h2>
      <ol>
        <li>{t('docs.imageGeneration.step1')}</li>
        <li>{t('docs.imageGeneration.step2')}</li>
        <li>{t('docs.imageGeneration.step3')}</li>
        <li>{t('docs.imageGeneration.step4')}</li>
      </ol>

      <h2>{t('docs.imageGeneration.setup')}</h2>

      <h3>{t('docs.imageGeneration.getApiKey')}</h3>
      <p>
        {t('docs.imageGeneration.getApiKeyDesc')}{' '}
        <a href="https://aistudio.google.com/apikey" target="_blank" rel="noopener noreferrer">
          {t('docs.imageGeneration.getApiKeyLink')}
        </a>
        .
      </p>

      <h3>{t('docs.imageGeneration.addToEnv')}</h3>
      <p>{t('docs.imageGeneration.addToEnvDesc')}</p>
      <CodeBlock language="bash">{`GEMINI_API_KEY=your_google_ai_api_key`}</CodeBlock>

      <h3>{t('docs.imageGeneration.restart')}</h3>
      <CodeBlock language="bash">{`systemctl --user restart openshannon`}</CodeBlock>

      <h2>{t('docs.imageGeneration.usage')}</h2>
      <CodeBlock language="text">{`/imagine a cat wearing a space helmet painting the Mona Lisa
/imagine minimalist logo for a coffee shop called "Bean There"
/imagine photorealistic sunset over a cyberpunk Tokyo`}</CodeBlock>
      <p>{t('docs.imageGeneration.usageDesc')}</p>

      <h2>{t('docs.imageGeneration.modelSelection')}</h2>
      <p>{t('docs.imageGeneration.modelSelectionDesc')}</p>
      <CodeBlock language="text">{`/model gemini       # Gemini 2.5 Flash
/model gemini-pro   # Gemini 2.5 Pro`}</CodeBlock>

      <h2>{t('docs.imageGeneration.availableModels')}</h2>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>{t('docs.imageGeneration.thShortcut')}</th>
              <th>{t('docs.imageGeneration.thModel')}</th>
              <th>{t('docs.imageGeneration.thBestFor')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>haiku</code></td>
              <td>Claude Haiku 4.5</td>
              <td>{t('docs.imageGeneration.haikuBestFor')}</td>
            </tr>
            <tr>
              <td><code>sonnet</code></td>
              <td>Claude Sonnet 4.6</td>
              <td>{t('docs.imageGeneration.sonnetBestFor')}</td>
            </tr>
            <tr>
              <td><code>opus</code></td>
              <td>Claude Opus 4.6</td>
              <td>{t('docs.imageGeneration.opusBestFor')}</td>
            </tr>
            <tr>
              <td><code>gemini</code></td>
              <td>Gemini 2.5 Flash</td>
              <td>{t('docs.imageGeneration.geminiBestFor')}</td>
            </tr>
            <tr>
              <td><code>gemini-pro</code></td>
              <td>Gemini 2.5 Pro</td>
              <td>{t('docs.imageGeneration.geminiProBestFor')}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </>
  )
}
