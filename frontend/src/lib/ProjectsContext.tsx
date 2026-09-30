import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'
import { listProjects, openProject, createProject } from './api'
import { useSSE } from './useSSE'
import type { Project } from './types'

interface ProjectsContextValue {
  projects: Project[]
  activeProjectId: number | null
  setActiveProjectId: (id: number | null) => void
  refreshProjects: () => Promise<void>
  openProjectByPath: (path: string) => Promise<void>
  createProjectByPath: (path: string) => Promise<void>
}

const ProjectsContext = createContext<ProjectsContextValue | null>(null)

export function ProjectsProvider({ children }: { children: ReactNode }) {
  const [projects, setProjects] = useState<Project[]>([])
  const [activeProjectId, setActiveProjectId] = useState<number | null>(null)

  const refreshProjects = useCallback(async () => {
    try {
      setProjects(await listProjects())
    } catch (err) {
      console.error('failed to list projects', err)
    }
  }, [])

  useEffect(() => {
    refreshProjects()
  }, [refreshProjects])

  useSSE(refreshProjects)

  const openProjectByPath = useCallback(
    async (path: string) => {
      const project = await openProject(path)
      await refreshProjects()
      setActiveProjectId(project.id)
    },
    [refreshProjects],
  )

  const createProjectByPath = useCallback(
    async (path: string) => {
      const project = await createProject(path)
      await refreshProjects()
      setActiveProjectId(project.id)
    },
    [refreshProjects],
  )

  return (
    <ProjectsContext.Provider
      value={{ projects, activeProjectId, setActiveProjectId, refreshProjects, openProjectByPath, createProjectByPath }}
    >
      {children}
    </ProjectsContext.Provider>
  )
}

export function useProjects() {
  const ctx = useContext(ProjectsContext)
  if (!ctx) throw new Error('useProjects must be used within a ProjectsProvider')
  return ctx
}
