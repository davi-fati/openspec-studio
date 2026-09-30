import { useCallback, useEffect, useMemo, useState } from 'react'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { FlowCard } from '@/components/specflow/FlowCard'
import { FlowCreateDialog } from '@/components/specflow/FlowCreateDialog'
import { getFlow, listFlows } from '@/lib/api'
import { useProjects } from '@/lib/ProjectsContext'
import { useSpecflowEvents } from '@/lib/useSpecflowEvents'
import type { Flow } from '@/lib/types'

export function Specflow() {
  const { activeProjectId } = useProjects()
  const [flows, setFlows] = useState<Flow[]>([])
  const [error, setError] = useState<string | null>(null)
  const [createOpen, setCreateOpen] = useState(false)

  const refresh = useCallback(() => {
    listFlows(activeProjectId)
      .then(setFlows)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load flows'))
  }, [activeProjectId])

  useEffect(() => {
    refresh()
  }, [refresh])

  // Real-time: refresh the single affected flow on any status/log event,
  // instead of refetching the whole list on every log chunk.
  useSpecflowEvents((ev) => {
    getFlow(ev.flowId)
      .then((updated) => {
        setFlows((prev) => {
          const exists = prev.some((f) => f.id === updated.id)
          return exists ? prev.map((f) => (f.id === updated.id ? updated : f)) : [...prev, updated]
        })
      })
      .catch(() => {})
  })

  const grouped = useMemo(() => {
    const byProject = new Map<string, Flow[]>()
    for (const f of flows) {
      const list = byProject.get(f.projectName) ?? []
      list.push(f)
      byProject.set(f.projectName, list)
    }
    return Array.from(byProject.entries()).sort(([a], [b]) => a.localeCompare(b))
  }, [flows])

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Specflow</h1>
          <p className="mt-1 text-muted-foreground">Scheduled, unattended implementation runs.</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="h-3.5 w-3.5" /> Schedule flow
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}
      {!error && flows.length === 0 && <p className="text-sm text-muted-foreground">No flows scheduled yet.</p>}

      {activeProjectId == null ? (
        <div className="flex flex-col gap-6">
          {grouped.map(([projectName, list]) => (
            <div key={projectName}>
              <h2 className="mb-2 text-sm font-semibold text-muted-foreground">{projectName}</h2>
              <div className="flex flex-col gap-3">
                {list.map((f) => (
                  <FlowCard key={f.id} flow={f} showProject={false} onChanged={refresh} />
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {flows.map((f) => (
            <FlowCard key={f.id} flow={f} showProject={false} onChanged={refresh} />
          ))}
        </div>
      )}

      <FlowCreateDialog open={createOpen} onClose={() => setCreateOpen(false)} onCreated={refresh} />
    </div>
  )
}
