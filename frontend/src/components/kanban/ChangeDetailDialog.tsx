import { useEffect, useState } from 'react'
import { Dialog } from '@/components/ui/dialog'
import { Markdown } from '@/components/markdown/Markdown'
import { cn } from '@/lib/utils'
import { getChangeDetail } from '@/lib/api'
import type { ChangeDetail } from '@/lib/types'

interface Props {
  changeId: string | null
  onClose: () => void
}

type DetailTab = 'Proposal' | 'Specs' | 'Tasks' | 'Design'

export function ChangeDetailDialog({ changeId, onClose }: Props) {
  const [detail, setDetail] = useState<ChangeDetail | null>(null)
  const [tab, setTab] = useState<DetailTab>('Proposal')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!changeId) {
      setDetail(null)
      return
    }
    setError(null)
    getChangeDetail(changeId)
      .then((d) => {
        setDetail(d)
        setTab(d.proposalMarkdown ? 'Proposal' : d.tasksMarkdown ? 'Tasks' : d.specs && Object.keys(d.specs).length ? 'Specs' : 'Design')
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load change'))
  }, [changeId])

  const availableTabs: DetailTab[] = detail
    ? ([
        detail.proposalMarkdown ? 'Proposal' : null,
        detail.specs && Object.keys(detail.specs).length ? 'Specs' : null,
        detail.tasksMarkdown ? 'Tasks' : null,
        detail.designMarkdown ? 'Design' : null,
      ].filter(Boolean) as DetailTab[])
    : []

  return (
    <Dialog open={!!changeId} onClose={onClose} title={detail?.name ?? 'Change'} className="w-[92vw] max-w-5xl">
      {error && <p className="text-sm text-destructive">{error}</p>}
      {!detail && !error && <p className="text-sm text-muted-foreground">Loading…</p>}
      {detail && (
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span className="rounded-full bg-muted px-2 py-0.5 font-medium">{detail.status}</span>
            <span>{detail.projectName}</span>
            {detail.hasTasks && (
              <span>
                {detail.tasksDone}/{detail.tasksTotal} tasks
              </span>
            )}
          </div>
          <div className="flex gap-1 border-b border-border">
            {availableTabs.map((t) => (
              <button
                key={t}
                type="button"
                onClick={() => setTab(t)}
                className={cn(
                  'border-b-2 px-3 py-1.5 text-xs font-medium text-muted-foreground',
                  tab === t ? 'border-primary text-foreground' : 'border-transparent hover:text-foreground',
                )}
              >
                {t}
              </button>
            ))}
          </div>
          <div>
            {tab === 'Proposal' && detail.proposalMarkdown && <Markdown>{detail.proposalMarkdown}</Markdown>}
            {tab === 'Tasks' && detail.tasksMarkdown && <Markdown>{detail.tasksMarkdown}</Markdown>}
            {tab === 'Design' && detail.designMarkdown && <Markdown>{detail.designMarkdown}</Markdown>}
            {tab === 'Specs' &&
              Object.entries(detail.specs ?? {}).map(([cap, content], i) => (
                <div key={cap}>
                  {i > 0 && <hr className="my-5 border-0 border-t border-border" />}
                  <div className="mb-1 text-sm font-semibold text-foreground">{cap}</div>
                  <Markdown>{content}</Markdown>
                </div>
              ))}
          </div>
        </div>
      )}
    </Dialog>
  )
}
