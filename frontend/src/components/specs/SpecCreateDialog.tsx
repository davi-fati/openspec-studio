import { useEffect, useState } from 'react'
import { Sparkles } from 'lucide-react'
import { Dialog } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { createSpec, generateSpecDraft, listProviders } from '@/lib/api'
import { useProjects } from '@/lib/ProjectsContext'
import type { Provider, SpecRequirement } from '@/lib/types'

interface Props {
  open: boolean
  onClose: () => void
  onCreated: () => void
}

export function SpecCreateDialog({ open, onClose, onCreated }: Props) {
  const { projects, activeProjectId } = useProjects()
  const available = projects.filter((p) => p.available)

  const [projectId, setProjectId] = useState<number | ''>(activeProjectId ?? available[0]?.id ?? '')
  const [capability, setCapability] = useState('')
  const [purpose, setPurpose] = useState('')
  const [requirements, setRequirements] = useState<SpecRequirement[]>([])
  const [description, setDescription] = useState('')
  const [providers, setProviders] = useState<Provider[]>([])
  const [providerId, setProviderId] = useState<number | ''>('')
  const [generating, setGenerating] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (open) listProviders().then(setProviders).catch(() => setProviders([]))
  }, [open])

  const reset = () => {
    setCapability('')
    setPurpose('')
    setRequirements([])
    setDescription('')
    setError(null)
  }

  const handleGenerate = async () => {
    if (!description.trim()) return
    setGenerating(true)
    setError(null)
    try {
      const draft = await generateSpecDraft(description, providerId || undefined)
      setCapability(draft.capability)
      setPurpose(draft.purpose)
      setRequirements(draft.requirements)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to generate draft')
    } finally {
      setGenerating(false)
    }
  }

  const handleSave = async () => {
    if (!projectId || !capability.trim()) {
      setError('project and capability are required')
      return
    }
    setSaving(true)
    setError(null)
    try {
      await createSpec({ projectId, capability: capability.trim(), purpose, requirements })
      onCreated()
      reset()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to create spec')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onClose={onClose} title="New spec">
      <div className="flex flex-col gap-4">
        {error && <p className="text-sm text-destructive">{error}</p>}

        <div className="rounded-md border border-border bg-muted/30 p-3">
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            Generate from a description (optional)
          </label>
          <div className="flex gap-2">
            <input
              className="flex-1 rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="e.g. Let users export their data as CSV"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
            <Button variant="outline" size="sm" onClick={handleGenerate} disabled={generating || !description.trim()}>
              <Sparkles className="h-3.5 w-3.5" /> {generating ? 'Generating…' : 'Generate'}
            </Button>
          </div>
          {providers.length > 0 && (
            <select
              className="mt-2 h-7 rounded-md border border-border bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
              value={providerId}
              onChange={(e) => setProviderId(e.target.value ? Number(e.target.value) : '')}
            >
              <option value="">Default provider{providers.find((p) => p.isDefault) ? ` (${providers.find((p) => p.isDefault)!.name})` : ''}</option>
              {providers.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          )}
          <p className="mt-1 text-[11px] text-muted-foreground">
            Produces a draft below for you to review and edit - nothing is saved until you click Create.
          </p>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="mb-1 block text-xs font-medium text-muted-foreground">Project</label>
            <select
              className="w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              value={projectId}
              onChange={(e) => setProjectId(Number(e.target.value))}
            >
              {available.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-muted-foreground">Capability</label>
            <input
              className="w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="e.g. data-export"
              value={capability}
              onChange={(e) => setCapability(e.target.value)}
            />
          </div>
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">Purpose</label>
          <textarea
            className="w-full rounded-md border border-border bg-background p-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            rows={2}
            value={purpose}
            onChange={(e) => setPurpose(e.target.value)}
          />
        </div>

        {requirements.length > 0 && (
          <div className="flex flex-col gap-2">
            <label className="text-xs font-medium text-muted-foreground">Requirements (editable)</label>
            {requirements.map((req, i) => (
              <div key={i} className="rounded-md border border-border p-2">
                <input
                  className="mb-1.5 w-full rounded-md border border-border bg-background px-2 py-1 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-ring"
                  value={req.name}
                  onChange={(e) =>
                    setRequirements((rs) => rs.map((r, idx) => (idx === i ? { ...r, name: e.target.value } : r)))
                  }
                />
                <textarea
                  className="mono w-full rounded-md border border-border bg-background p-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                  rows={4}
                  value={req.body}
                  onChange={(e) =>
                    setRequirements((rs) => rs.map((r, idx) => (idx === i ? { ...r, body: e.target.value } : r)))
                  }
                />
              </div>
            ))}
          </div>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? 'Creating…' : 'Create spec'}
          </Button>
        </div>
      </div>
    </Dialog>
  )
}
