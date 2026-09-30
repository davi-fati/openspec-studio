import { useState } from 'react'
import { AlertTriangle, CheckCircle2, Circle, Loader2, Square, XCircle } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { cancelFlow, deleteFlow } from '@/lib/api'
import type { Flow, FlowItemStatus } from '@/lib/types'
import { cn } from '@/lib/utils'

const ITEM_ICON: Record<FlowItemStatus, React.ReactNode> = {
  pending: <Circle className="h-3.5 w-3.5 text-muted-foreground" />,
  running: <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" />,
  succeeded: <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600" />,
  failed: <XCircle className="h-3.5 w-3.5 text-destructive" />,
}

interface Props {
  flow: Flow
  showProject: boolean
  onChanged: () => void
}

export function FlowCard({ flow, showProject, onChanged }: Props) {
  const [openLogItem, setOpenLogItem] = useState<number | null>(null)

  return (
    <Card className="p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            {showProject && <span className="text-xs font-semibold text-muted-foreground">{flow.projectName}</span>}
            <span
              className={cn(
                'rounded-full px-2 py-0.5 text-[11px] font-medium',
                flow.status === 'running' && 'bg-primary/15 text-primary',
                flow.status === 'succeeded' && 'bg-emerald-500/15 text-emerald-600',
                flow.status === 'failed' && 'bg-destructive/15 text-destructive',
                flow.status === 'cancelled' && 'bg-muted text-muted-foreground',
                flow.status === 'pending' && 'bg-muted text-muted-foreground',
              )}
            >
              {flow.status}
            </span>
            {flow.missed && (
              <span className="flex items-center gap-1 rounded-full bg-destructive/15 px-2 py-0.5 text-[11px] font-medium text-destructive">
                <AlertTriangle className="h-3 w-3" /> missed
              </span>
            )}
          </div>
          <div className="mt-1 text-xs text-muted-foreground">
            Scheduled {new Date(flow.scheduledAt).toLocaleString()}
          </div>
        </div>
        <div className="flex gap-2">
          {flow.status === 'running' && (
            <Button size="sm" variant="outline" onClick={() => cancelFlow(flow.id).then(onChanged)}>
              <Square className="h-3.5 w-3.5" /> Cancel
            </Button>
          )}
          {flow.status === 'pending' && (
            <Button size="sm" variant="outline" onClick={() => deleteFlow(flow.id).then(onChanged)}>
              Remove
            </Button>
          )}
        </div>
      </div>

      <div className="mt-3 flex flex-col gap-1.5">
        {flow.items.map((item) => (
          <div key={item.id} className="rounded-md border border-border">
            <button
              type="button"
              className="flex w-full items-center gap-2 p-2 text-left text-sm"
              onClick={() => setOpenLogItem(openLogItem === item.id ? null : item.id)}
            >
              {ITEM_ICON[item.status]}
              <span className="flex-1">{item.changeName}</span>
              <span className="text-[11px] text-muted-foreground">{item.status}</span>
            </button>
            {openLogItem === item.id && (
              <pre className="mono max-h-48 overflow-auto border-t border-border bg-muted/40 p-2 text-[11px] leading-relaxed whitespace-pre-wrap">
                {item.log || 'No output yet.'}
              </pre>
            )}
          </div>
        ))}
      </div>
    </Card>
  )
}
