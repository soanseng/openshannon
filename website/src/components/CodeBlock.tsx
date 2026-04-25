interface CodeBlockProps {
  children: string
  language?: string
}

export default function CodeBlock({ children, language }: CodeBlockProps) {
  return (
    <pre className="bg-code-bg text-code-text rounded-lg p-4 w-full max-w-full box-border overflow-x-auto whitespace-pre-wrap break-words text-sm leading-relaxed">
      <code className={language ? `language-${language} block max-w-full` : 'block max-w-full'}>
        {children}
      </code>
    </pre>
  )
}
