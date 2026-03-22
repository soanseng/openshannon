import CodeBlock from '../../components/CodeBlock'

export default function ImageGeneration() {
  return (
    <>
      <h1>Image Generation</h1>
      <p>
        OpenShannon uses Claude to enhance your prompt, then sends it to Gemini 3.1 Flash to
        generate images. The result is sent directly back to your Telegram chat.
      </p>

      <h2>How It Works</h2>
      <ol>
        <li>You send <code>/imagine a cat wearing a space helmet</code></li>
        <li>Claude enhances your prompt with more detail and artistic direction</li>
        <li>The enhanced prompt is sent to Gemini Flash for image generation</li>
        <li>The generated image is sent back to your Telegram chat</li>
      </ol>

      <h2>Setup</h2>

      <h3>1. Get a Gemini API Key</h3>
      <p>
        Get your API key from{' '}
        <a href="https://aistudio.google.com/apikey" target="_blank" rel="noopener noreferrer">
          Google AI Studio
        </a>.
      </p>

      <h3>2. Add to Environment</h3>
      <p>
        Add the key to your <code>~/.config/openshannon/env</code> file:
      </p>
      <CodeBlock language="bash">{`GEMINI_API_KEY=your_google_ai_api_key`}</CodeBlock>

      <h3>3. Restart the Service</h3>
      <CodeBlock language="bash">{`systemctl --user restart openshannon`}</CodeBlock>

      <h2>Usage</h2>
      <CodeBlock language="text">{`/imagine a cat wearing a space helmet painting the Mona Lisa
/imagine minimalist logo for a coffee shop called "Bean There"
/imagine photorealistic sunset over a cyberpunk Tokyo`}</CodeBlock>
      <p>
        Just describe what you want in natural language. Claude will handle turning your
        description into a detailed image generation prompt.
      </p>

      <h2>Model Selection for Image Generation</h2>
      <p>
        To use Gemini models for text chat as well (not just image generation), you can switch
        your session model:
      </p>
      <CodeBlock language="text">{`/model gemini       # Gemini 2.5 Flash
/model gemini-pro   # Gemini 2.5 Pro`}</CodeBlock>

      <h2>Available Models</h2>
      <div className="overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>Shortcut</th>
              <th>Model</th>
              <th>Best For</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td><code>haiku</code></td>
              <td>Claude Haiku 4.5</td>
              <td>Fast, cheap tasks</td>
            </tr>
            <tr>
              <td><code>sonnet</code></td>
              <td>Claude Sonnet 4.6</td>
              <td>Balanced performance</td>
            </tr>
            <tr>
              <td><code>opus</code></td>
              <td>Claude Opus 4.6</td>
              <td>Most capable reasoning</td>
            </tr>
            <tr>
              <td><code>gemini</code></td>
              <td>Gemini 2.5 Flash</td>
              <td>Fast multimodal</td>
            </tr>
            <tr>
              <td><code>gemini-pro</code></td>
              <td>Gemini 2.5 Pro</td>
              <td>Advanced multimodal</td>
            </tr>
          </tbody>
        </table>
      </div>
    </>
  )
}
