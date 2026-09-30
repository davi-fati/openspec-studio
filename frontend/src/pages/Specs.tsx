import { useCallback, useEffect, useMemo, useState } from 'react'
import { AlertTriangle, Plus } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { SpecDetailDialog } from '@/components/specs/SpecDetailDialog'
import { SpecCreateDialog } from '@/components/specs/SpecCreateDialog'
import { ReviewsPanel } from '@/components/context/ReviewsPanel'
import { DebtsPanel } from '@/components/context/DebtsPanel'
import { listSpecs } from '@/lib/api'
import { useProjects } from '@/lib/ProjectsContext'
import { useSSE } from '@/lib/useSSE'
import type { Spec } from '@/lib/types'

export function Specs() {
  const [specs, setSpecs] = useState<Spec[]>([])
  const [error, setError] = useState<string | null>(null)
  const [openSpecId, setOpenSpecId] = useState<string | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const { activeProjectId, projects } = useProjects()
  const activeProject = projects.find((p) => p.id === activeProjectId)
  const subtitle =
    activeProjectId == null
      ? `Specs across all ${projects.length} registered project(s).`
      : 'Specs for the selected project.'

  const refresh = useCallback(() => {
    listSpecs(activeProjectId)
      .then(setSpecs)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load specs'))
  }, [activeProjectId])

  useEffect(() => {
    refresh()
  }, [refresh])

  useSSE(refresh)

  const grouped = useMemo(() => {
    const byProject = new Map<string, Spec[]>()
    for (const s of specs) {
      const list = byProject.get(s.projectName) ?? []
      list.push(s)
      byProject.set(s.projectName, list)
    }
    return Array.from(byProject.entries()).sort(([a], [b]) => a.localeCompare(b))
  }, [specs])

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Specs</h1>
          <p className="mt-1 text-muted-foreground">{subtitle}</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="h-3.5 w-3.5" /> New spec
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}
      {!error && specs.length === 0 && (
        <p className="text-sm text-muted-foreground">No specs found yet.</p>
      )}

      {grouped.map(([projectName, list]) => (
        <div key={projectName}>
          <h2 className="mb-2 text-sm font-semibold text-muted-foreground">{projectName}</h2>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {list.map((s) => {
              const broken = [...s.dependsOn, ...s.dependedBy].some((d) => !d.available)
              return (
                <Card
                  key={s.id}
                  className="cursor-pointer p-4 transition-shadow hover:shadow-md"
                  onClick={() => setOpenSpecId(s.id)}
                >
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{s.capability}</span>
                    {broken && <AlertTriangle className="h-3.5 w-3.5 shrink-0 text-destructive" />}
                  </div>
                  <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{s.purpose}</p>
                  <p className="mt-2 text-[11px] text-muted-foreground">
                    {s.requirements.length} requirement{s.requirements.length === 1 ? '' : 's'}
                  </p>
                </Card>
              )
            })}
          </div>
        </div>
      ))}

      {activeProjectId != null && activeProject && (
        <div className="border-t border-border pt-6">
          <h2 className="mb-4 text-sm font-semibold text-muted-foreground">
            Project context - {activeProject.name}
          </h2>
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <div>
              <h3 className="mb-2 text-xs font-medium text-muted-foreground">Reviews</h3>
              <ReviewsPanel projectId={activeProjectId} />
            </div>
            <div>
              <h3 className="mb-2 text-xs font-medium text-muted-foreground">Tech Debt</h3>
              <DebtsPanel projectId={activeProjectId} />
            </div>
          </div>
        </div>
      )}

      <SpecDetailDialog specId={openSpecId} onClose={() => setOpenSpecId(null)} onChanged={refresh} />
      <SpecCreateDialog open={createOpen} onClose={() => setCreateOpen(false)} onCreated={refresh} />
    </div>
  )
}
