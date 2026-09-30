import { useCallback, useEffect, useState } from 'react'
import { KanbanBoard } from '@/components/kanban/KanbanBoard'
import { listChanges, listFlows } from '@/lib/api'
import { useProjects } from '@/lib/ProjectsContext'
import { useSSE } from '@/lib/useSSE'
import { useSpecflowEvents } from '@/lib/useSpecflowEvents'
import type { Change } from '@/lib/types'

export function Kanban() {
  const { activeProjectId, projects } = useProjects()
  const [changes, setChanges] = useState<Change[]>([])
  const [runningChangeKeys, setRunningChangeKeys] = useState<Set<string>>(new Set())
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(() => {
    listChanges(activeProjectId)
      .then(setChanges)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load changes'))
  }, [activeProjectId])

  const refreshRunningFlows = useCallback(() => {
    listFlows(activeProjectId)
      .then((flows) => {
        const keys = new Set<string>()
        for (const f of flows) {
          if (f.status !== 'running') continue
          for (const item of f.items) {
            if (item.status === 'running') keys.add(`${f.projectId}:${item.changeName}`)
          }
        }
        setRunningChangeKeys(keys)
      })
      .catch(() => {})
  }, [activeProjectId])

  useEffect(() => {
    refresh()
    refreshRunningFlows()
  }, [refresh, refreshRunningFlows])

  useSSE(refresh)
  useSpecflowEvents(refreshRunningFlows)

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Kanban</h1>
        <p className="mt-1 text-muted-foreground">
          {activeProjectId == null
            ? `Changes across all ${projects.length} registered project(s)`
            : 'Changes for the selected project'}
        </p>
      </div>
      {error && <p className="text-sm text-destructive">{error}</p>}
      {!error && changes.length === 0 && (
        <p className="text-sm text-muted-foreground">No changes found yet.</p>
      )}
      <KanbanBoard changes={changes} showProject={activeProjectId == null} runningChangeKeys={runningChangeKeys} />
    </div>
  )
}
