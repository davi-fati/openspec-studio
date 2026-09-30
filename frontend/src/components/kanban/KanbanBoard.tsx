import { useState } from 'react'
import { KANBAN_COLUMNS, type Change } from '@/lib/types'
import { ChangeCard } from './ChangeCard'
import { ChangeDetailDialog } from './ChangeDetailDialog'
import { COLUMN_ICON } from './statusIcons'

interface Props {
  changes: Change[]
  showProject: boolean
  runningChangeKeys?: Set<string>
}

export function KanbanBoard({ changes, showProject, runningChangeKeys }: Props) {
  const [openChangeId, setOpenChangeId] = useState<string | null>(null)

  return (
    <>
      <div className="flex gap-4 pb-2">
        {KANBAN_COLUMNS.map((column) => {
          const items = changes.filter((c) => c.status === column.status)
          return (
            <div key={column.status} className="flex min-w-0 flex-1 flex-col">
              <div className="sticky top-0 z-[1] mb-3 flex items-center gap-2 border-b border-border bg-background py-2">
                {COLUMN_ICON[column.status]}
                <h3 className="text-sm font-semibold">{column.label}</h3>
                <span className="ml-auto rounded-full bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
                  {items.length}
                </span>
              </div>
              <div className="flex max-h-[calc(100dvh-16rem)] flex-col gap-2.5 overflow-y-auto pr-1">
                {items.map((c) => (
                  <ChangeCard
                    key={c.id}
                    change={c}
                    showProject={showProject}
                    implementing={runningChangeKeys?.has(`${c.projectId}:${c.name}`) ?? false}
                    onClick={() => setOpenChangeId(c.id)}
                  />
                ))}
                {items.length === 0 && (
                  <p className="py-4 text-center text-xs text-muted-foreground">No changes</p>
                )}
              </div>
            </div>
          )
        })}
      </div>
      <ChangeDetailDialog changeId={openChangeId} onClose={() => setOpenChangeId(null)} />
    </>
  )
}
