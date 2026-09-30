import { useCallback, useEffect, useState } from 'react'
import { CheckCircle2, Star, Trash2, XCircle } from 'lucide-react'
import { Dialog } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import {
  createCLIProvider,
  createHostedProvider,
  deleteProvider,
  healthCheckProvider,
  listProviders,
  setDefaultProvider,
} from '@/lib/api'
import type { HealthResult, Provider } from '@/lib/types'

interface Props {
  open: boolean
  onClose: () => void
}

export function ProviderSettingsDialog({ open, onClose }: Props) {
  const [providers, setProviders] = useState<Provider[]>([])
  const [health, setHealth] = useState<Record<number, HealthResult | 'checking'>>({})
  const [error, setError] = useState<string | null>(null)

  const [hostedName, setHostedName] = useState('')
  const [hostedModel, setHostedModel] = useState('')
  const [hostedEndpoint, setHostedEndpoint] = useState('')
  const [hostedKey, setHostedKey] = useState('')
  const [savingHosted, setSavingHosted] = useState(false)

  const [cliName, setCliName] = useState('')
  const [cliPath, setCliPath] = useState('')
  const [cliArgs, setCliArgs] = useState('')
  const [cliError, setCliError] = useState<string | null>(null)
  const [savingCli, setSavingCli] = useState(false)

  const refresh = useCallback(() => {
    listProviders()
      .then(setProviders)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load providers'))
  }, [])

  useEffect(() => {
    if (open) refresh()
  }, [open, refresh])

  if (!open) return null

  const handleAddHosted = async () => {
    if (!hostedName.trim() || !hostedEndpoint.trim()) return
    setSavingHosted(true)
    setError(null)
    try {
      await createHostedProvider({
        name: hostedName.trim(),
        model: hostedModel.trim(),
        endpoint: hostedEndpoint.trim(),
        apiKey: hostedKey,
        isDefault: providers.length === 0,
      })
      setHostedName('')
      setHostedModel('')
      setHostedEndpoint('')
      setHostedKey('')
      refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to add hosted provider')
    } finally {
      setSavingHosted(false)
    }
  }

  const handleAddCli = async () => {
    if (!cliName.trim() || !cliPath.trim()) return
    setSavingCli(true)
    setCliError(null)
    try {
      await createCLIProvider({
        name: cliName.trim(),
        cliPath: cliPath.trim(),
        cliArgs: cliArgs.trim() ? cliArgs.trim().split(/\s+/) : [],
        isDefault: providers.length === 0,
      })
      setCliName('')
      setCliPath('')
      setCliArgs('')
      refresh()
    } catch (err) {
      setCliError(err instanceof Error ? err.message : 'executable not found')
    } finally {
      setSavingCli(false)
    }
  }

  const handleHealthCheck = async (id: number) => {
    setHealth((h) => ({ ...h, [id]: 'checking' }))
    try {
      const result = await healthCheckProvider(id)
      setHealth((h) => ({ ...h, [id]: result }))
    } catch (err) {
      setHealth((h) => ({ ...h, [id]: { ok: false, message: err instanceof Error ? err.message : 'failed' } }))
    }
  }

  return (
    <Dialog open={open} onClose={onClose} title="AI Providers" className="max-w-2xl">
      <div className="flex flex-col gap-6">
        {error && <p className="text-sm text-destructive">{error}</p>}

        <div>
          <h4 className="mb-2 text-xs font-medium text-muted-foreground">Configured providers</h4>
          {providers.length === 0 && <p className="text-xs text-muted-foreground">None configured yet.</p>}
          <div className="flex flex-col gap-2">
            {providers.map((p) => {
              const h = health[p.id]
              return (
                <div key={p.id} className="flex items-center justify-between rounded-md border border-border p-2.5">
                  <div className="flex items-center gap-2">
                    {p.isDefault && <Star className="h-3.5 w-3.5 fill-primary text-primary" />}
                    <div>
                      <div className="text-sm font-medium">{p.name}</div>
                      <div className="text-[11px] text-muted-foreground">
                        {p.kind === 'hosted' ? `hosted · ${p.model || 'no model'}` : `cli · ${p.cliPath}`}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {h === 'checking' && <span className="text-[11px] text-muted-foreground">checking…</span>}
                    {h && h !== 'checking' && (
                      <span className={`flex items-center gap-1 text-[11px] ${h.ok ? 'text-emerald-600' : 'text-destructive'}`}>
                        {h.ok ? <CheckCircle2 className="h-3.5 w-3.5" /> : <XCircle className="h-3.5 w-3.5" />}
                        {h.message}
                      </span>
                    )}
                    <Button size="sm" variant="outline" onClick={() => handleHealthCheck(p.id)}>
                      Check
                    </Button>
                    {!p.isDefault && (
                      <Button size="sm" variant="outline" onClick={() => setDefaultProvider(p.id).then(refresh)}>
                        Set default
                      </Button>
                    )}
                    <button
                      type="button"
                      onClick={() => deleteProvider(p.id).then(refresh)}
                      className="rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-destructive"
                      aria-label="Delete provider"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        <div className="rounded-md border border-border p-3">
          <h4 className="mb-2 text-xs font-medium text-muted-foreground">Add hosted API provider</h4>
          <div className="grid grid-cols-2 gap-2">
            <input
              className="rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Name"
              value={hostedName}
              onChange={(e) => setHostedName(e.target.value)}
            />
            <input
              className="rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Model"
              value={hostedModel}
              onChange={(e) => setHostedModel(e.target.value)}
            />
            <input
              className="col-span-2 rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="API endpoint URL"
              value={hostedEndpoint}
              onChange={(e) => setHostedEndpoint(e.target.value)}
            />
            <input
              className="col-span-2 rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="API key"
              type="password"
              value={hostedKey}
              onChange={(e) => setHostedKey(e.target.value)}
            />
          </div>
          <div className="mt-2 flex justify-end">
            <Button size="sm" onClick={handleAddHosted} disabled={savingHosted}>
              {savingHosted ? 'Adding…' : 'Add hosted provider'}
            </Button>
          </div>
        </div>

        <div className="rounded-md border border-border p-3">
          <h4 className="mb-2 text-xs font-medium text-muted-foreground">Add CLI agent provider</h4>
          <div className="grid grid-cols-2 gap-2">
            <input
              className="rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Name"
              value={cliName}
              onChange={(e) => setCliName(e.target.value)}
            />
            <input
              className="rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Executable path (e.g. claude)"
              value={cliPath}
              onChange={(e) => setCliPath(e.target.value)}
            />
            <input
              className="col-span-2 rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Extra args (space-separated, optional)"
              value={cliArgs}
              onChange={(e) => setCliArgs(e.target.value)}
            />
          </div>
          {cliError && <p className="mt-1 text-xs text-destructive">{cliError}</p>}
          <div className="mt-2 flex justify-end">
            <Button size="sm" onClick={handleAddCli} disabled={savingCli}>
              {savingCli ? 'Adding…' : 'Add CLI provider'}
            </Button>
          </div>
        </div>
      </div>
    </Dialog>
  )
}
