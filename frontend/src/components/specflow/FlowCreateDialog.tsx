import { useEffect, useState } from 'react'
import { ArrowDown, ArrowUp, X } from 'lucide-react'
import { Dialog } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { createFlow, listChanges, listProviders } from '@/lib/api'
import { useProjects } from '@/lib/ProjectsContext'
import type { Change, Provider } from '@/lib/types'

interface Props {
  open: boolean
  onClose: () => void
  onCreated: () => void
}

export function FlowCreateDialog({ open, onClose, onCreated }: Props) {
  const { projects, activeProjectId } = useProjects()
  const available = projects.filter((p) => p.available)

  const [projectId, setProjectId] = useState<number | ''>(activeProjectId ?? available[0]?.id ?? '')
  const [changes, setChanges] = useState<Change[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [selected, setSelected] = useState<string[]>([]) // change names, ordered
  const [providerId, setProviderId] = useState<number | ''>('')
  const [scheduledAt, setScheduledAt] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open || !projectId) return
    listChanges(Number(projectId)).then(setChanges).catch(() => setChanges([]))
  }, [open, projectId])

  useEffect(() => {
    if (open) listProviders().then(setProviders).catch(() => setProviders([]))
  }, [open])

  if (!open) return null

  const eligible = changes.filter((c) => c.hasProposal && c.hasTasks && c.status !== 'archived')
  const move = (name: string, dir: -1 | 1) => {
    setSelected((s) => {
      const i = s.indexOf(name)
      const j = i + dir
      if (j < 0 || j >= s.length) return s
      const copy = [...s]
      ;[copy[i], copy[j]] = [copy[j], copy[i]]
      return copy
    })
  }

  const handleCreate = async () => {
    if (!projectId || !providerId || !scheduledAt || selected.length === 0) {
      setError('project, provider, schedule, and at least one change are required')
      return
    }
    setSaving(true)
    setError(null)
    try {
      await createFlow({
        projectId: Number(projectId),
        providerId: Number(providerId),
        scheduledAt: new Date(scheduledAt).toISOString(),
        changeNames: selected,
      })
      setSelected([])
      onCreated()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to schedule flow')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onClose={onClose} title="Schedule a Specflow">
      <div className="flex flex-col gap-4">
        {error && <p className="text-sm text-destructive">{error}</p>}

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-xs font-medium text-muted-foreground">Project</label>
            <select
              className="w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              value={projectId}
              onChange={(e) => {
                setProjectId(Number(e.target.value))
                setSelected([])
              }}
            >
              {available.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-muted-foreground">Provider</label>
            <select
              className="w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              value={providerId}
              onChange={(e) => setProviderId(Number(e.target.value))}
            >
              <option value="">Select a provider…</option>
              {providers.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">Scheduled start</label>
          <input
            type="datetime-local"
            className="w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            value={scheduledAt}
            onChange={(e) => setScheduledAt(e.target.value)}
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            Fully-planned changes available in this project
          </label>
          <div className="flex flex-col gap-1">
            {eligible.length === 0 && (
              <p className="text-xs text-muted-foreground">No fully-planned changes (proposal + tasks) found.</p>
            )}
            {eligible.map((c) => (
              <label key={c.name} className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={selected.includes(c.name)}
                  onChange={(e) =>
                    setSelected((s) => (e.target.checked ? [...s, c.name] : s.filter((n) => n !== c.name)))
                  }
                />
                {c.name}
              </label>
            ))}
          </div>
        </div>

        {selected.length > 0 && (
          <div>
            <label className="mb-1 block text-xs font-medium text-muted-foreground">Execution order</label>
            <div className="flex flex-col gap-1">
              {selected.map((name, i) => (
                <div key={name} className="flex items-center gap-2 rounded-md border border-border px-2 py-1 text-sm">
                  <span className="flex-1">
                    {i + 1}. {name}
                  </span>
                  <button type="button" onClick={() => move(name, -1)} disabled={i === 0} className="disabled:opacity-30">
                    <ArrowUp className="h-3.5 w-3.5" />
                  </button>
                  <button
                    type="button"
                    onClick={() => move(name, 1)}
                    disabled={i === selected.length - 1}
                    className="disabled:opacity-30"
                  >
                    <ArrowDown className="h-3.5 w-3.5" />
                  </button>
                  <button type="button" onClick={() => setSelected((s) => s.filter((n) => n !== name))}>
                    <X className="h-3.5 w-3.5" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleCreate} disabled={saving}>
            {saving ? 'Scheduling…' : 'Schedule flow'}
          </Button>
        </div>
      </div>
    </Dialog>
  )
}
