import { useEffect, useMemo, useState } from 'react'
import { AlertTriangle, Plus, Trash2 } from 'lucide-react'
import { Dialog } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import {
  addDependency,
  deleteSpec,
  getSpec,
  listSpecs,
  removeDependency,
  updateSpec,
  SpecHasDependentsError,
} from '@/lib/api'
import type { Spec, SpecRequirement } from '@/lib/types'

interface Props {
  specId: string | null
  onClose: () => void
  onChanged: () => void
}

export function SpecDetailDialog({ specId, onClose, onChanged }: Props) {
  const [spec, setSpec] = useState<Spec | null>(null)
  const [allSpecs, setAllSpecs] = useState<Spec[]>([])
  const [purpose, setPurpose] = useState('')
  const [requirements, setRequirements] = useState<SpecRequirement[]>([])
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [dependents, setDependents] = useState<Spec[] | null>(null)
  const [depTarget, setDepTarget] = useState('')

  useEffect(() => {
    if (!specId) {
      setSpec(null)
      setDependents(null)
      setConfirmDelete(false)
      return
    }
    setError(null)
    Promise.all([getSpec(specId), listSpecs(null)])
      .then(([s, all]) => {
        setSpec(s)
        setPurpose(s.purpose)
        setRequirements(s.requirements)
        setAllSpecs(all)
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load spec'))
  }, [specId])

  const dependencyOptions = useMemo(
    () =>
      allSpecs
        .filter((s) => s.id !== specId)
        .map((s) => ({ id: s.id, label: `${s.projectName} / ${s.capability}`, projectId: s.projectId, capability: s.capability })),
    [allSpecs, specId],
  )

  if (!specId) return null

  const handleSave = async () => {
    if (!spec) return
    setSaving(true)
    setError(null)
    try {
      const updated = await updateSpec(spec.id, { purpose, requirements })
      setSpec(updated)
      onChanged()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to save spec')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (force: boolean) => {
    if (!spec) return
    try {
      await deleteSpec(spec.id, force)
      onChanged()
      onClose()
    } catch (err) {
      if (err instanceof SpecHasDependentsError) {
        setDependents(err.dependents)
        return
      }
      setError(err instanceof Error ? err.message : 'failed to delete spec')
    }
  }

  const handleAddDependency = async () => {
    if (!spec || !depTarget) return
    const opt = dependencyOptions.find((o) => o.id === depTarget)
    if (!opt) return
    try {
      await addDependency(spec.id, opt.projectId, opt.capability)
      const refreshed = await getSpec(spec.id)
      setSpec(refreshed)
      setDepTarget('')
      onChanged()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to add dependency')
    }
  }

  const handleRemoveDependency = async (targetProjectId: number, targetCapability: string) => {
    if (!spec) return
    try {
      await removeDependency(spec.id, targetProjectId, targetCapability)
      const refreshed = await getSpec(spec.id)
      setSpec(refreshed)
      onChanged()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to remove dependency')
    }
  }

  return (
    <Dialog open={!!specId} onClose={onClose} title={spec ? spec.capability : 'Spec'} className="max-w-3xl">
      {error && <p className="mb-3 text-sm text-destructive">{error}</p>}
      {!spec && !error && <p className="text-sm text-muted-foreground">Loading…</p>}
      {spec && (
        <div className="flex flex-col gap-6">
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
            <span>
              <strong className="text-foreground">Project:</strong> {spec.provenance.project || spec.projectName}
            </span>
            <span>
              <strong className="text-foreground">Author:</strong> {spec.provenance.author || '—'}
            </span>
            <span>
              <strong className="text-foreground">Created:</strong>{' '}
              {spec.provenance.createdAt ? new Date(spec.provenance.createdAt).toLocaleString() : '—'}
            </span>
            <span>
              <strong className="text-foreground">ID:</strong> {spec.provenance.id || '—'}
            </span>
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

          <div>
            <div className="mb-2 flex items-center justify-between">
              <label className="text-xs font-medium text-muted-foreground">Requirements</label>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setRequirements((r) => [...r, { name: 'New requirement', body: 'The system SHALL ...\n\n#### Scenario: ...\n- **WHEN** ...\n- **THEN** ...' }])}
              >
                <Plus className="h-3.5 w-3.5" /> Add
              </Button>
            </div>
            <div className="flex flex-col gap-3">
              {requirements.map((req, i) => (
                <div key={i} className="rounded-md border border-border p-3">
                  <div className="mb-2 flex items-center gap-2">
                    <input
                      className="flex-1 rounded-md border border-border bg-background px-2 py-1 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-ring"
                      value={req.name}
                      onChange={(e) =>
                        setRequirements((rs) => rs.map((r, idx) => (idx === i ? { ...r, name: e.target.value } : r)))
                      }
                    />
                    <button
                      type="button"
                      className="rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-destructive"
                      onClick={() => setRequirements((rs) => rs.filter((_, idx) => idx !== i))}
                      aria-label="Remove requirement"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
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
          </div>

          <div className="flex justify-end">
            <Button onClick={handleSave} disabled={saving}>
              {saving ? 'Saving…' : 'Save changes'}
            </Button>
          </div>

          <div className="border-t border-border pt-4">
            <h4 className="mb-2 text-xs font-medium text-muted-foreground">Depends on</h4>
            <div className="flex flex-wrap gap-2">
              {spec.dependsOn.length === 0 && <span className="text-xs text-muted-foreground">None</span>}
              {spec.dependsOn.map((d) => (
                <DependencyBadge
                  key={`${d.projectPath}/${d.capability}`}
                  label={`${d.projectName || d.projectPath} / ${d.capability}`}
                  broken={!d.available}
                  onRemove={() => {
                    const targetSpec = allSpecs.find((s) => s.projectName === d.projectName && s.capability === d.capability)
                    if (targetSpec) handleRemoveDependency(targetSpec.projectId, targetSpec.capability)
                  }}
                />
              ))}
            </div>
            <div className="mt-2 flex gap-2">
              <select
                className="h-8 flex-1 rounded-md border border-border bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                value={depTarget}
                onChange={(e) => setDepTarget(e.target.value)}
              >
                <option value="">Add dependency…</option>
                {dependencyOptions.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.label}
                  </option>
                ))}
              </select>
              <Button size="sm" variant="outline" onClick={handleAddDependency} disabled={!depTarget}>
                Link
              </Button>
            </div>
          </div>

          <div>
            <h4 className="mb-2 text-xs font-medium text-muted-foreground">Depended on by</h4>
            <div className="flex flex-wrap gap-2">
              {spec.dependedBy.length === 0 && <span className="text-xs text-muted-foreground">None</span>}
              {spec.dependedBy.map((d) => (
                <DependencyBadge
                  key={`${d.projectPath}/${d.capability}`}
                  label={`${d.projectName || d.projectPath} / ${d.capability}`}
                  broken={!d.available}
                />
              ))}
            </div>
          </div>

          {((spec.linkedDebts?.length ?? 0) > 0 || (spec.linkedReviews?.length ?? 0) > 0) && (
            <div className="border-t border-border pt-4">
              <h4 className="mb-2 text-xs font-medium text-muted-foreground">Project context</h4>
              <div className="flex flex-col gap-1.5">
                {spec.linkedDebts?.map((d) => (
                  <div key={d.id} className="flex items-center gap-2 text-xs">
                    <span className="rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium">Debt</span>
                    <span>{d.title}</span>
                    <span className="text-muted-foreground">({d.status})</span>
                  </div>
                ))}
                {spec.linkedReviews?.map((r) => (
                  <div key={r.id} className="flex items-center gap-2 text-xs">
                    <span className="rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium">Review</span>
                    <span>{r.title}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="border-t border-border pt-4">
            {!confirmDelete && (
              <Button variant="outline" size="sm" onClick={() => setConfirmDelete(true)}>
                <Trash2 className="h-3.5 w-3.5" /> Delete spec
              </Button>
            )}
            {confirmDelete && (
              <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3">
                <p className="mb-2 text-sm">Delete <strong>{spec.capability}</strong>? This removes it from disk.</p>
                {dependents && dependents.length > 0 && (
                  <div className="mb-2 flex items-start gap-2 rounded-md bg-background p-2 text-xs">
                    <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-destructive" />
                    <span>
                      Depended on by: {dependents.map((d) => `${d.projectName}/${d.capability}`).join(', ')}. Deleting
                      anyway will leave those links broken.
                    </span>
                  </div>
                )}
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => setConfirmDelete(false)}>
                    Cancel
                  </Button>
                  <Button size="sm" onClick={() => handleDelete(!!dependents && dependents.length > 0)}>
                    {dependents && dependents.length > 0 ? 'Delete anyway' : 'Confirm delete'}
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </Dialog>
  )
}

function DependencyBadge({ label, broken, onRemove }: { label: string; broken: boolean; onRemove?: () => void }) {
  return (
    <span
      className={
        broken
          ? 'inline-flex items-center gap-1 rounded-full border border-destructive/40 bg-destructive/10 px-2 py-0.5 text-xs text-destructive'
          : 'inline-flex items-center gap-1 rounded-full border border-border bg-muted px-2 py-0.5 text-xs'
      }
    >
      {broken && <AlertTriangle className="h-3 w-3" />}
      {label}
      {onRemove && (
        <button type="button" onClick={onRemove} className="ml-1 opacity-60 hover:opacity-100" aria-label="Remove">
          ×
        </button>
      )}
    </span>
  )
}
