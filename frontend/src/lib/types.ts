export interface Project {
  id: number
  name: string
  path: string
  lastOpenedAt: string
  available: boolean
}

export interface BrowseEntry {
  name: string
  path: string
  hasOpenspec: boolean
  isEmpty: boolean
}

export interface BrowseResult {
  path: string
  parent?: string
  hasOpenspec: boolean
  isEmpty: boolean
  entries: BrowseEntry[]
}

export const TABS = ['Overview', 'Kanban', 'Specs', 'Specflow'] as const
export type Tab = (typeof TABS)[number]

export type ChangeStatus = 'draft' | 'todo' | 'in_progress' | 'done' | 'archived'

export const KANBAN_COLUMNS: { status: ChangeStatus; label: string }[] = [
  { status: 'draft', label: 'Draft' },
  { status: 'todo', label: 'Todo' },
  { status: 'in_progress', label: 'In Progress' },
  { status: 'done', label: 'Done' },
  { status: 'archived', label: 'Archived' },
]

export interface Change {
  id: string
  name: string
  projectId: number
  projectName: string
  status: ChangeStatus
  hasProposal: boolean
  hasDesign: boolean
  hasSpecs: boolean
  hasTasks: boolean
  tasksTotal: number
  tasksDone: number
}

export interface ChangeDetail extends Change {
  proposalMarkdown?: string
  designMarkdown?: string
  tasksMarkdown?: string
  specs?: Record<string, string>
}

export interface SpecProvenance {
  id: string
  createdAt: string
  author: string
  project: string
}

export interface SpecRequirement {
  name: string
  body: string
}

export interface SpecDependency {
  projectPath: string
  projectName: string
  capability: string
  available: boolean
}

export interface ArtifactRef {
  kind?: 'spec' | 'change' | ''
  capability?: string
  changeName?: string
}

export interface DebtItemSummary {
  id: string
  projectId: number
  title: string
  body: string
  status: DebtStatus
  link: ArtifactRef
  createdAt: string
}

export interface ReviewSummary {
  id: string
  projectId: number
  title: string
  body: string
  link: ArtifactRef
  createdAt: string
}

export interface Spec {
  id: string
  capability: string
  projectId: number
  projectName: string
  purpose: string
  requirements: SpecRequirement[]
  provenance: SpecProvenance
  dependsOn: SpecDependency[]
  dependedBy: SpecDependency[]
  linkedDebts?: DebtItemSummary[]
  linkedReviews?: ReviewSummary[]
}

export type DebtStatus = 'open' | 'in_progress' | 'resolved'

export const DEBT_STATUSES: DebtStatus[] = ['open', 'in_progress', 'resolved']

export interface ProjectOverview {
  projectId: number
  projectName: string
  available: boolean
  specCount: number
  brokenSpecCount: number
  changeCounts: Record<ChangeStatus, number>
  progressRatio: number
}

export interface RecentProject {
  projectId: number
  projectName: string
  lastOpenedAt: string
}

export interface Overview {
  projects: ProjectOverview[]
  totalOpenChanges: number
  totalInProgress: number
  recentlyUpdated: RecentProject[]
}

export type ProviderKind = 'hosted' | 'cli'

export interface Provider {
  id: number
  name: string
  kind: ProviderKind
  model?: string
  endpoint?: string
  cliPath?: string
  cliArgs?: string[]
  isDefault: boolean
  hasApiKey: boolean
  createdAt: string
}

export interface HealthResult {
  ok: boolean
  message: string
}

export type FlowStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'cancelled'
export type FlowItemStatus = 'pending' | 'running' | 'succeeded' | 'failed'

export interface FlowItem {
  id: number
  position: number
  changeName: string
  status: FlowItemStatus
  log: string
  startedAt?: string
  finishedAt?: string
}

export interface Flow {
  id: number
  projectId: number
  projectName: string
  providerId: number
  scheduledAt: string
  status: FlowStatus
  missed: boolean
  createdAt: string
  items: FlowItem[]
}
