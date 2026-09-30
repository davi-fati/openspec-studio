import { useEffect, useId, useRef, useState } from 'react'
import { useDocumentTheme } from '@/lib/useDocumentTheme'

interface Props {
  code: string
}

/** Renders a Mermaid diagram, re-coloring on theme toggle and falling back to
 * an inline error (without breaking the rest of the document) on invalid syntax. */
export function MermaidDiagram({ code }: Props) {
  const id = useId().replace(/:/g, '-')
  const theme = useDocumentTheme()
  const [svg, setSvg] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const generation = useRef(0)

  useEffect(() => {
    const myGeneration = ++generation.current

    import('mermaid')
      .then(({ default: mermaid }) => {
        mermaid.initialize({ startOnLoad: false, theme: theme === 'dark' ? 'dark' : 'default' })
        return mermaid.render(`mermaid-${id}`, code)
      })
      .then(({ svg }) => {
        if (generation.current === myGeneration) {
          setSvg(svg)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (generation.current === myGeneration) {
          setError(err instanceof Error ? err.message : 'invalid mermaid syntax')
          setSvg(null)
        }
      })
  }, [code, theme, id])

  if (error) {
    return (
      <div className="my-2 rounded-md border border-destructive/40 bg-destructive/5 p-3 text-xs text-destructive">
        Mermaid diagram failed to render: {error}
      </div>
    )
  }

  if (!svg) {
    return <div className="my-2 text-xs text-muted-foreground">Rendering diagram…</div>
  }

  // eslint-disable-next-line react/no-danger -- mermaid's own SVG output, not user-facing HTML
  return <div className="my-2 flex justify-center" dangerouslySetInnerHTML={{ __html: svg }} />
}
