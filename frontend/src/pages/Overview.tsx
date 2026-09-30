import { useCallback, useEffect, useState } from 'react'
import { AlertTriangle, FolderOpen, FolderPlus, Loader2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { DirectoryBrowserDialog } from '@/components/DirectoryBrowserDialog'
import { getOverview, listFlows } from '@/lib/api'
import { useNavigation } from '@/lib/NavigationContext'
import { useProjects } from '@/lib/ProjectsContext'
import { useSSE } from '@/lib/useSSE'
import { useSpecflowEvents } from '@/lib/useSpecflowEvents'
import { KANBAN_COLUMNS } from '@/lib/types'
import type { Overview as OverviewData, Tab } from '@/lib/types'

export function Overview() {
  const [overview, setOverview] = useState<OverviewData | null>(null)
  const [runningByProject, setRunningByProject] = useState<Set<number>>(new Set())
  const [error, setError] = useState<string | null>(null)
  const [browserMode, setBrowserMode] = useState<'open' | 'new' | null>(null)
  const { activeProjectId, setActiveProjectId, projects, openProjectByPath, createProjectByPath } = useProjects()
  const { setActiveTab } = useNavigation()

  const refresh = useCallback(() => {
    getOverview(activeProjectId)
      .then(setOverview)
      .catch((err) => setError(err instanceof Error ? err.message : 'failed to load overview'))
  }, [activeProjectId])

  const refreshRunningFlows = useCallback(() => {
    listFlows(null)
      .then((flows) => setRunningByProject(new Set(flows.filter((f) => f.status === 'running').map((f) => f.projectId))))
      .catch(() => {})
  }, [])

  useEffect(() => {
    refresh()
    refreshRunningFlows()
  }, [refresh, refreshRunningFlows])

  useSSE(refresh)
  useSpecflowEvents(refreshRunningFlows)

  const goToProject = (projectId: number, tab: Tab) => {
    setActiveProjectId(projectId)
    setActiveTab(tab)
  }

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Overview</h1>
        <p className="mt-1 text-muted-foreground">
          {activeProjectId == null
            ? `Progress across all ${projects.length} registered project(s).`
            : 'Progress for the selected project.'}
        </p>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {!overview && !error && (
        <div className="flex items-center justify-center gap-2 py-16 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" /> Loading overview…
        </div>
      )}

      {overview && (
        <>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <Card className="p-4">
              <div className="text-xs text-muted-foreground">Open changes</div>
              <div className="mt-1 text-2xl font-bold">{overview.totalOpenChanges}</div>
            </Card>
            <Card className="p-4">
              <div className="text-xs text-muted-foreground">In progress</div>
              <div className="mt-1 text-2xl font-bold">{overview.totalInProgress}</div>
            </Card>
            <Card className="p-4">
              <div className="text-xs text-muted-foreground">Recently updated</div>
              <div className="mt-1 flex flex-col gap-1">
                {(overview.recentlyUpdated ?? []).length === 0 && (
                  <span className="text-sm text-muted-foreground">—</span>
                )}
                {(overview.recentlyUpdated ?? []).map((r) => (
                  <button
                    key={r.projectId}
                    type="button"
                    onClick={() => goToProject(r.projectId, 'Kanban')}
                    className="flex items-center justify-between gap-2 text-left text-sm hover:text-primary"
                  >
                    <span className="truncate font-medium">{r.projectName}</span>
                    <span className="shrink-0 text-[11px] text-muted-foreground">
                      {new Date(r.lastOpenedAt).toLocaleString()}
                    </span>
                  </button>
                ))}
              </div>
            </Card>
          </div>

          {(overview.projects ?? []).length === 0 && (
            <div className="flex flex-col items-center gap-3 py-10 text-center">
              <p className="text-sm text-muted-foreground">No projects registered yet.</p>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={() => setBrowserMode('open')}>
                  <FolderOpen className="h-3.5 w-3.5" /> Open
                </Button>
                <Button size="sm" onClick={() => setBrowserMode('new')}>
                  <FolderPlus className="h-3.5 w-3.5" /> New
                </Button>
              </div>
            </div>
          )}

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {(overview.projects ?? []).map((p) => (
              <Card key={p.projectId} className="p-4">
                <CardHeader className="p-0">
                  <div className="flex items-center justify-between">
                    <CardTitle>{p.projectName}</CardTitle>
                    <div className="flex items-center gap-1.5">
                      {runningByProject.has(p.projectId) && (
                        <span className="flex items-center gap-1 rounded-full bg-primary/15 px-2 py-0.5 text-[11px] font-medium text-primary">
                          <Loader2 className="h-3 w-3 animate-spin" /> specflow running
                        </span>
                      )}
                      {p.brokenSpecCount > 0 && (
                        <span className="flex items-center gap-1 rounded-full bg-destructive/10 px-2 py-0.5 text-[11px] font-medium text-destructive">
                          <AlertTriangle className="h-3 w-3" /> {p.brokenSpecCount} broken
                        </span>
                      )}
                      {!p.available && (
                        <span className="rounded-full bg-destructive/10 px-2 py-0.5 text-[11px] font-medium text-destructive">
                          unavailable
                        </span>
                      )}
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="p-0 pt-3">
                  {p.available ? (
                    <>
                      <div className="mb-2 text-xs text-muted-foreground">{p.specCount} spec(s)</div>
                      <div className="flex flex-wrap gap-1.5 text-[11px]">
                        {KANBAN_COLUMNS.map((col) => (
                          <span key={col.status} className="rounded-full bg-muted px-2 py-0.5">
                            {col.label}: {p.changeCounts[col.status] ?? 0}
                          </span>
                        ))}
                      </div>
                      <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-muted">
                        <div
                          className={`h-full rounded-full ${
                            p.changeCounts.todo === 0 && p.changeCounts.in_progress === 0
                              ? 'bg-emerald-500'
                              : 'bg-amber-500'
                          }`}
                          style={{ width: `${Math.round(p.progressRatio * 100)}%` }}
                        />
                      </div>
                      <div className="mt-3 flex items-center gap-1.5">
                        <Button variant="outline" size="sm" onClick={() => goToProject(p.projectId, 'Kanban')}>
                          Kanban
                        </Button>
                        <Button variant="outline" size="sm" onClick={() => goToProject(p.projectId, 'Specs')}>
                          Specs
                        </Button>
                        <Button variant="outline" size="sm" onClick={() => goToProject(p.projectId, 'Specflow')}>
                          Specflow
                        </Button>
                      </div>
                    </>
                  ) : (
                    <p className="text-xs text-muted-foreground">Path is currently inaccessible.</p>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        </>
      )}

      <DirectoryBrowserDialog
        open={browserMode !== null}
        mode={browserMode ?? 'open'}
        onClose={() => setBrowserMode(null)}
        onConfirm={browserMode === 'new' ? createProjectByPath : openProjectByPath}
      />
    </div>
  )
}
