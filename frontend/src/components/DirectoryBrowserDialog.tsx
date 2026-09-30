import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { AlertTriangle, ChevronRight, Folder, FolderCheck, Home } from 'lucide-react'
import { Dialog } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { browseDirectory } from '@/lib/api'
import { isDesktop, pickDirectory } from '@/lib/desktop'
import type { BrowseEntry, BrowseResult } from '@/lib/types'

type Mode = 'open' | 'new'

interface Props {
  open: boolean
  mode: Mode
  onClose: () => void
  onConfirm: (path: string) => Promise<void>
}

/**
 * Folder picker for the Open/New project flows. The desktop app uses the
 * native OS dialog; a browser can never learn a real absolute path from a
 * native dialog, so web mode keeps the in-app browser. Both hand the chosen
 * path to the same onConfirm (Open/New API calls).
 */
export function DirectoryBrowserDialog(props: Props) {
  return isDesktop() ? <NativeDirectoryPicker {...props} /> : <InAppDirectoryBrowser {...props} />
}

/**
 * Shows the native dialog each time `open` turns true. The OS dialog can't
 * pre-filter folders like the in-app browser does, so a selection the
 * backend rejects (no openspec/, not empty) is reported in a small dialog.
 */
function NativeDirectoryPicker({ open, mode, onClose, onConfirm }: Props) {
  const [error, setError] = useState<string | null>(null)
  // Parents pass inline callbacks; the effect below must fire only when
  // `open` flips, so it reads the latest props through a ref.
  const latest = useRef({ mode, onClose, onConfirm })
  useLayoutEffect(() => {
    latest.current = { mode, onClose, onConfirm }
  })
  // StrictMode runs effects twice in dev; never stack two native dialogs.
  const showing = useRef(false)

  useEffect(() => {
    if (!open || showing.current) return
    showing.current = true
    const { mode, onClose, onConfirm } = latest.current
    pickDirectory(mode === 'open' ? 'Open an OpenSpec project' : 'Choose an empty folder for a new project')
      .then(async (path) => {
        if (path === null) {
          onClose()
          return
        }
        await onConfirm(path)
        onClose()
      })
      .catch((err) => setError(err instanceof Error ? err.message : String(err)))
      .finally(() => {
        showing.current = false
      })
  }, [open])

  const close = () => {
    setError(null)
    onClose()
  }

  return (
    <Dialog
      open={open && error !== null}
      onClose={close}
      title={mode === 'open' ? 'Could not open project' : 'Could not create project'}
      className="max-w-md"
    >
      <div className="flex flex-col gap-3">
        <p className="text-sm text-destructive">{error}</p>
        <div className="flex justify-end">
          <Button size="sm" variant="outline" onClick={close}>
            Close
          </Button>
        </div>
      </div>
    </Dialog>
  )
}

/**
 * In-app directory browser backed by the Go backend's own local filesystem
 * access (GET /api/fs/browse).
 */
function InAppDirectoryBrowser({ open, mode, onClose, onConfirm }: Props) {
  const [result, setResult] = useState<BrowseResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)

  useEffect(() => {
    if (!open) {
      setResult(null)
      setError(null)
      return
    }
    browseDirectory()
      .then(setResult)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to browse'))
  }, [open])

  if (!open) return null

  const navigate = (path?: string) => {
    setError(null)
    browseDirectory(path)
      .then(setResult)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to browse'))
  }

  const isEntryConfirmable = (e: BrowseEntry) => (mode === 'open' ? e.hasOpenspec : e.isEmpty)
  const isCurrentDirConfirmable = (r: BrowseResult) => (mode === 'open' ? r.hasOpenspec : r.isEmpty)

  const handleConfirm = async (path: string) => {
    setConfirming(true)
    setError(null)
    try {
      await onConfirm(path)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to confirm selection')
    } finally {
      setConfirming(false)
    }
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title={mode === 'open' ? 'Open a project' : 'New project'}
      className="max-w-xl"
    >
      <div className="flex flex-col gap-3">
        <p className="text-xs text-muted-foreground">
          {mode === 'open'
            ? 'Browse to and select a directory that already has an openspec/ structure.'
            : 'Browse to and select an empty directory - a fresh OpenSpec project will be scaffolded there.'}
        </p>

        {error && <p className="text-sm text-destructive">{error}</p>}

        {result && (
          <>
            <div className="flex items-center gap-1 overflow-x-auto rounded-md border border-border bg-muted/40 px-2 py-1.5 text-xs">
              <button
                type="button"
                onClick={() => navigate()}
                className="flex shrink-0 items-center gap-1 rounded p-1 hover:bg-accent"
                title="Home"
              >
                <Home className="h-3.5 w-3.5" />
              </button>
              <span className="truncate font-mono">{result.path}</span>
              {result.hasOpenspec && (
                <span className="ml-auto shrink-0 rounded-full bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-medium text-emerald-600">
                  OpenSpec project
                </span>
              )}
              {result.isEmpty && (
                <span className="ml-auto shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                  empty
                </span>
              )}
            </div>

            <div className="flex items-center justify-between">
              <Button variant="outline" size="sm" onClick={() => navigate(result.parent)} disabled={!result.parent}>
                Up one level
              </Button>
              <Button
                size="sm"
                onClick={() => handleConfirm(result.path)}
                disabled={confirming || !isCurrentDirConfirmable(result)}
                title={
                  mode === 'open'
                    ? 'Use this folder as the project (only enabled if it already has openspec/)'
                    : 'Scaffold a new project here (only enabled if this folder is empty)'
                }
              >
                {confirming ? 'Confirming…' : 'Use this folder'}
              </Button>
            </div>

            <div className="flex max-h-72 flex-col gap-1 overflow-y-auto">
              {result.entries.length === 0 && (
                <p className="py-6 text-center text-xs text-muted-foreground">No subdirectories here.</p>
              )}
              {result.entries.map((e) => {
                const confirmable = isEntryConfirmable(e)
                return (
                  <div
                    key={e.path}
                    className="flex items-center gap-2 rounded-md border border-border px-2 py-1.5 text-sm"
                  >
                    <button
                      type="button"
                      onClick={() => navigate(e.path)}
                      className="flex flex-1 items-center gap-2 text-left hover:text-primary"
                    >
                      {e.hasOpenspec ? (
                        <FolderCheck className="h-4 w-4 shrink-0 text-emerald-500" />
                      ) : (
                        <Folder className="h-4 w-4 shrink-0 text-muted-foreground" />
                      )}
                      <span className="truncate">{e.name}</span>
                      <ChevronRight className="ml-auto h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                    </button>
                    {e.hasOpenspec && (
                      <span className="shrink-0 rounded-full bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-medium text-emerald-600">
                        OpenSpec project
                      </span>
                    )}
                    {e.isEmpty && (
                      <span className="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                        empty
                      </span>
                    )}
                    {confirmable && (
                      <Button size="sm" className="shrink-0" onClick={() => handleConfirm(e.path)} disabled={confirming}>
                        Select
                      </Button>
                    )}
                  </div>
                )
              })}
            </div>

            {mode === 'open' && !result.hasOpenspec && (
              <div className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 p-2 text-xs">
                <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-amber-600" />
                <span>
                  This folder has no valid OpenSpec project. Browse to one flagged "OpenSpec project", or use{' '}
                  <strong>New</strong> instead to scaffold one from scratch in an empty folder.
                </span>
              </div>
            )}
          </>
        )}

        {!result && !error && <p className="text-sm text-muted-foreground">Loading…</p>}
      </div>
    </Dialog>
  )
}
