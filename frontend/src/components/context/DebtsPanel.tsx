import { useCallback, useEffect, useState } from 'react'
import { Trash2 } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { createDebt, deleteDebt, listDebts, updateDebtStatus } from '@/lib/api'
import { DEBT_STATUSES } from '@/lib/types'
import type { DebtItemSummary } from '@/lib/types'

const STATUS_LABEL: Record<string, string> = {
  open: 'Open',
  in_progress: 'In Progress',
  resolved: 'Resolved',
}

export function DebtsPanel({ projectId }: { projectId: number }) {
  const [debts, setDebts] = useState<DebtItemSummary[]>([])
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [capability, setCapability] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  const refresh = useCallback(() => {
    listDebts(projectId)
      .then(setDebts)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load tech debt'))
  }, [projectId])

  useEffect(() => {
    refresh()
  }, [refresh])

  const handleCreate = async () => {
    if (!title.trim()) return
    setSaving(true)
    setError(null)
    try {
      await createDebt(projectId, {
        title: title.trim(),
        body,
        link: capability.trim() ? { kind: 'spec', capability: capability.trim() } : undefined,
      })
      setTitle('')
      setBody('')
      setCapability('')
      refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to create debt item')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex flex-col gap-3">
      <Card className="p-3">
        <div className="flex flex-col gap-2">
          <input
            className="rounded-md border border-border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            placeholder="Debt title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
          <textarea
            className="rounded-md border border-border bg-background p-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            rows={2}
            placeholder="Description…"
            value={body}
            onChange={(e) => setBody(e.target.value)}
          />
          <div className="flex gap-2">
            <input
              className="flex-1 rounded-md border border-border bg-background px-2 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Link to spec capability (optional)"
              value={capability}
              onChange={(e) => setCapability(e.target.value)}
            />
            <Button size="sm" onClick={handleCreate} disabled={saving || !title.trim()}>
              {saving ? 'Adding…' : 'Add debt'}
            </Button>
          </div>
        </div>
      </Card>

      {error && <p className="text-xs text-destructive">{error}</p>}
      {debts.length === 0 && <p className="text-xs text-muted-foreground">No tech debt tracked yet.</p>}
      {debts.map((d) => (
        <Card key={d.id} className="p-3">
          <div className="flex items-start justify-between gap-2">
            <div>
              <div className="text-sm font-medium">{d.title}</div>
              {d.body && <p className="mt-1 text-xs text-muted-foreground">{d.body}</p>}
              {d.link.capability && (
                <span className="mt-1 inline-block rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
                  spec / {d.link.capability}
                </span>
              )}
            </div>
            <div className="flex shrink-0 items-center gap-2">
              <select
                className="h-7 rounded-md border border-border bg-background px-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
                value={d.status}
                onChange={(e) =>
                  updateDebtStatus(projectId, d.id, e.target.value as DebtItemSummary['status']).then(refresh)
                }
              >
                {DEBT_STATUSES.map((s) => (
                  <option key={s} value={s}>
                    {STATUS_LABEL[s]}
                  </option>
                ))}
              </select>
              <button
                type="button"
                onClick={() => deleteDebt(projectId, d.id).then(refresh)}
                className="rounded-md p-1 text-muted-foreground hover:bg-accent hover:text-destructive"
                aria-label="Delete debt item"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
        </Card>
      ))}
    </div>
  )
}
