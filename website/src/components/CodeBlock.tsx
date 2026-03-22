interface CodeBlockProps {
  children: string
  language?: string
}

export default function CodeBlock({ children, language }: CodeBlockProps) {
  return (
    <pre className="bg-code-bg text-code-text rounded-lg p-4 overflow-x-auto text-sm leading-relaxed">
      <code className={language ? `language-${language}` : ''}>
        {children}
      </code>
    </pre>
  )
}
