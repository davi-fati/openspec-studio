import { useEffect } from 'react'
import lightCss from 'highlight.js/styles/github.css?inline'
import darkCss from 'highlight.js/styles/github-dark.css?inline'
import { useDocumentTheme } from '@/lib/useDocumentTheme'

/** Injects the light/dark highlight.js theme matching the current app theme
 * into a single <style> tag, so code blocks re-color on theme toggle. */
export function HighlightTheme() {
  const theme = useDocumentTheme()

  useEffect(() => {
    let tag = document.getElementById('hljs-theme') as HTMLStyleElement | null
    if (!tag) {
      tag = document.createElement('style')
      tag.id = 'hljs-theme'
      document.head.appendChild(tag)
    }
    tag.textContent = theme === 'dark' ? darkCss : lightCss
  }, [theme])

  return null
}
