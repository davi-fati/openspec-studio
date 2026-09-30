import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import { MermaidDiagram } from './MermaidDiagram'

interface Props {
  children: string
}

/** The single shared Markdown renderer for every artifact view (proposal,
 * design, spec, tasks) - GitHub-flavored (tables, task lists, footnotes),
 * syntax-highlighted code blocks, and Mermaid diagrams for ```mermaid fences. */
export function Markdown({ children }: Props) {
  return (
    <div className="markdown-body text-sm leading-relaxed">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight]}
        components={{
          code(props) {
            const { className, children, ...rest } = props
            const match = /language-(\w+)/.exec(className || '')
            const isBlock = Boolean(match)
            if (isBlock && match?.[1] === 'mermaid') {
              return <MermaidDiagram code={String(children).trim()} />
            }
            return (
              <code className={className} {...rest}>
                {children}
              </code>
            )
          },
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  )
}
