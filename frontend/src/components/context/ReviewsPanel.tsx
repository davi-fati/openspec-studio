import { useCallback, useEffect, useState } from 'react'
import { Trash2 } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { createReview, deleteReview, listReviews } from '@/lib/api'
import type { ReviewSummary } from '@/lib/types'

export function ReviewsPanel({ projectId }: { projectId: number }) {
  const [reviews, setReviews] = useState<ReviewSummary[]>([])
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [capability, setCapability] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  const refresh = useCallback(() => {
    listReviews(projectId)
      .then(setReviews)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load reviews'))
  }, [projectId])

  useEffect(() => {
    refresh()
  }, [refresh])

  const handleCreate = async () => {
    if (!title.trim()) return
    setSaving(true)
    setError(null)
    try {
      await createReview(projectId, {
        title: title.trim(),
        body,
        link: capability.trim() ? { kind: 'spec', capability: capability.trim() } : undefined,
      })
      setTitle('')
      setBody('')
      setCapability('')
      refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to create review')
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
            placeholder="Review title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
          <textarea
            className="rounded-md border border-border bg-background p-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            rows={2}
            placeholder="Notes…"
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
              {saving ? 'Adding…' : 'Add review'}
            </Button>
          </div>
        </div>
      </Card>

      {error && <p className="text-xs text-destructive">{error}</p>}
      {reviews.length === 0 && <p className="text-xs text-muted-foreground">No reviews yet.</p>}
      {reviews.map((r) => (
        <Card key={r.id} className="p-3">
          <div className="flex items-start justify-between gap-2">
            <div>
              <div className="text-sm font-medium">{r.title}</div>
              {r.body && <p className="mt-1 text-xs text-muted-foreground">{r.body}</p>}
              {r.link.capability && (
                <span className="mt-1 inline-block rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
                  spec / {r.link.capability}
                </span>
              )}
            </div>
            <button
              type="button"
              onClick={() => deleteReview(projectId, r.id).then(refresh)}
              className="shrink-0 rounded-md p-1 text-muted-foreground hover:bg-accent hover:text-destructive"
              aria-label="Delete review"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          </div>
        </Card>
      ))}
    </div>
  )
}
