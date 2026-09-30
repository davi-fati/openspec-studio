import { CheckSquare, FileText, Layers, Loader2, Palette } from 'lucide-react'
import { Card } from '@/components/ui/card'
import type { Change } from '@/lib/types'
import { cn } from '@/lib/utils'

interface Props {
  change: Change
  showProject: boolean
  implementing?: boolean
  onClick: () => void
}

const ARTIFACTS: { key: keyof Change; label: string; Icon: typeof FileText; tone: string }[] = [
  { key: 'hasProposal', label: 'Proposal', Icon: FileText, tone: 'bg-blue-500/10 text-blue-500 dark:text-blue-400' },
  { key: 'hasDesign', label: 'Design', Icon: Palette, tone: 'bg-purple-500/10 text-purple-500 dark:text-purple-400' },
  { key: 'hasSpecs', label: 'Specs', Icon: Layers, tone: 'bg-emerald-500/10 text-emerald-500 dark:text-emerald-400' },
  { key: 'hasTasks', label: 'Tasks', Icon: CheckSquare, tone: 'bg-amber-500/10 text-amber-500 dark:text-amber-400' },
]

export function ChangeCard({ change, showProject, implementing, onClick }: Props) {
  const progress = change.tasksTotal > 0 ? (change.tasksDone / change.tasksTotal) * 100 : 0
  const isComplete = change.hasTasks && change.tasksTotal > 0 && progress === 100

  return (
    <Card
      onClick={onClick}
      className={cn(
        'cursor-pointer p-3 transition-shadow hover:shadow-md',
        isComplete && 'border-emerald-500/30 bg-emerald-500/5',
      )}
    >
      <div className="mb-2 flex items-start justify-between gap-2">
        <h3 className="flex-1 text-sm font-medium leading-snug">{change.name}</h3>
        {showProject && (
          <span className="shrink-0 rounded-md bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
            {change.projectName}
          </span>
        )}
        {implementing && (
          <span className="flex shrink-0 items-center gap-1 rounded-full bg-primary/15 px-2 py-0.5 text-[10px] font-medium text-primary">
            <Loader2 className="h-2.5 w-2.5 animate-spin" /> implementing
          </span>
        )}
      </div>

      {change.hasTasks && change.tasksTotal > 0 && (
        <div className="mb-2.5 space-y-1">
          <div className="flex items-center justify-between text-[11px]">
            <span className="font-medium text-muted-foreground">
              {change.tasksDone} of {change.tasksTotal} tasks
            </span>
            <span className={cn('font-semibold tabular-nums', isComplete ? 'text-emerald-600' : 'text-foreground')}>
              {Math.round(progress)}%
            </span>
          </div>
          <div className="h-1.5 overflow-hidden rounded-full bg-muted">
            <div
              className={cn('h-full rounded-full transition-all', isComplete ? 'bg-emerald-500' : 'bg-primary')}
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}

      <div className="flex flex-wrap items-center gap-1.5">
        {ARTIFACTS.map(({ key, label, Icon, tone }) => {
          const present = Boolean(change[key])
          return (
            <div
              key={key}
              title={present ? `${label}: written` : `${label}: not written yet`}
              className={cn(
                'flex items-center gap-1 rounded-md px-1.5 py-0.5',
                present ? tone : 'bg-muted/50 text-muted-foreground/50',
              )}
            >
              <Icon className="h-3 w-3" />
              <span className="text-[10px] font-medium">{label}</span>
            </div>
          )
        })}
      </div>
    </Card>
  )
}
