import { useEffect, useState } from 'react'

/**
 * True when the UI runs inside the Tauri desktop shell. The webview loads the
 * same backend-served UI as a browser, so this is the only switch between
 * desktop-only integrations and their web fallbacks.
 */
export function isDesktop(): boolean {
  return typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window
}

/** Opens the native OS folder picker; resolves null when dismissed. */
export async function pickDirectory(title: string): Promise<string | null> {
  const { open } = await import('@tauri-apps/plugin-dialog')
  const selected = await open({ directory: true, multiple: false, title })
  return typeof selected === 'string' ? selected : null
}

/**
 * On macOS the desktop window uses an overlay title bar: the Studio header is
 * the title bar and the window controls sit on its left (tauri.conf.json).
 */
export function hasOverlayTitleBar(): boolean {
  return isDesktop() && navigator.userAgent.includes('Mac')
}

// Fullscreen is read from the viewport filling the screen, so the webview
// needs no window-state permission for it.
function isFullscreen(): boolean {
  return window.innerWidth >= screen.width && window.innerHeight >= screen.height
}

/**
 * Whether the header must leave room for the macOS window controls: only in
 * the desktop app, and not in fullscreen, where macOS hides them.
 */
export function useWindowControlsInset(): boolean {
  const overlay = hasOverlayTitleBar()
  const [fullscreen, setFullscreen] = useState(() => overlay && isFullscreen())
  useEffect(() => {
    if (!overlay) return
    const update = () => setFullscreen(isFullscreen())
    window.addEventListener('resize', update)
    return () => window.removeEventListener('resize', update)
  }, [overlay])
  return overlay && !fullscreen
}
