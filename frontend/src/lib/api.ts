import type {
  ArtifactRef,
  BrowseResult,
  Change,
  ChangeDetail,
  DebtItemSummary,
  DebtStatus,
  Flow,
  HealthResult,
  Overview,
  Project,
  Provider,
  ReviewSummary,
  Spec,
  SpecRequirement,
} from './types'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    const err = new Error(body.error ?? `request failed: ${res.status}`) as Error & {
      status?: number
      body?: unknown
    }
    err.status = res.status
    err.body = body
    throw err
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export function listProjects(): Promise<Project[]> {
  return request<Project[]>('/projects')
}

export function openProject(path: string): Promise<Project> {
  return request<Project>('/projects', { method: 'POST', body: JSON.stringify({ path }) })
}

export function createProject(path: string): Promise<Project> {
  return request<Project>('/projects/new', { method: 'POST', body: JSON.stringify({ path }) })
}

export function browseDirectory(path?: string): Promise<BrowseResult> {
  const qs = path ? `?path=${encodeURIComponent(path)}` : ''
  return request<BrowseResult>(`/fs/browse${qs}`)
}

export function listChanges(projectId: number | null): Promise<Change[]> {
  const qs = projectId != null ? `?projectId=${projectId}` : ''
  return request<Change[]>(`/changes${qs}`)
}

export function getChangeDetail(id: string): Promise<ChangeDetail> {
  return request<ChangeDetail>(`/changes/${id}`)
}

export function listSpecs(projectId: number | null): Promise<Spec[]> {
  const qs = projectId != null ? `?projectId=${projectId}` : ''
  return request<Spec[]>(`/specs${qs}`)
}

export function getSpec(id: string): Promise<Spec> {
  return request<Spec>(`/specs/${id}`)
}

export function createSpec(input: {
  projectId: number
  capability: string
  purpose: string
  requirements: SpecRequirement[]
}): Promise<Spec> {
  return request<Spec>('/specs', { method: 'POST', body: JSON.stringify(input) })
}

export function updateSpec(
  id: string,
  input: { purpose: string; requirements: SpecRequirement[] },
): Promise<Spec> {
  return request<Spec>(`/specs/${id}`, { method: 'PATCH', body: JSON.stringify(input) })
}

export class SpecHasDependentsError extends Error {
  dependents: Spec[]
  constructor(dependents: Spec[]) {
    super('spec has dependents')
    this.dependents = dependents
  }
}

export async function deleteSpec(id: string, force = false): Promise<void> {
  const qs = force ? '?force=true' : ''
  try {
    await request<void>(`/specs/${id}${qs}`, { method: 'DELETE' })
  } catch (err) {
    const e = err as Error & { status?: number; body?: unknown }
    if (e.status === 409) {
      throw new SpecHasDependentsError((e.body as Spec[]) ?? [])
    }
    throw err
  }
}

export function addDependency(id: string, targetProjectId: number, targetCapability: string): Promise<void> {
  return request<void>(`/specs/${id}/dependencies`, {
    method: 'POST',
    body: JSON.stringify({ targetProjectId, targetCapability }),
  })
}

export function removeDependency(id: string, targetProjectId: number, targetCapability: string): Promise<void> {
  return request<void>(`/specs/${id}/dependencies`, {
    method: 'DELETE',
    body: JSON.stringify({ targetProjectId, targetCapability }),
  })
}

export function generateSpecDraft(description: string, providerId?: number): Promise<Spec> {
  return request<Spec>('/specs/generate', {
    method: 'POST',
    body: JSON.stringify({ description, providerId: providerId || undefined }),
  })
}

export function listProviders(): Promise<Provider[]> {
  return request<Provider[]>('/ai-providers')
}

export function createHostedProvider(input: {
  name: string
  model: string
  endpoint: string
  apiKey: string
  isDefault?: boolean
}): Promise<Provider> {
  return request<Provider>('/ai-providers', {
    method: 'POST',
    body: JSON.stringify({ kind: 'hosted', ...input }),
  })
}

export function createCLIProvider(input: {
  name: string
  cliPath: string
  cliArgs?: string[]
  isDefault?: boolean
}): Promise<Provider> {
  return request<Provider>('/ai-providers', {
    method: 'POST',
    body: JSON.stringify({ kind: 'cli', ...input }),
  })
}

export function deleteProvider(id: number): Promise<void> {
  return request<void>(`/ai-providers/${id}`, { method: 'DELETE' })
}

export function setDefaultProvider(id: number): Promise<void> {
  return request<void>(`/ai-providers/${id}/default`, { method: 'POST' })
}

export function healthCheckProvider(id: number): Promise<HealthResult> {
  return request<HealthResult>(`/ai-providers/${id}/health`, { method: 'POST' })
}

export function listReviews(projectId: number): Promise<ReviewSummary[]> {
  return request<ReviewSummary[]>(`/projects/${projectId}/reviews`)
}

export function createReview(
  projectId: number,
  input: { title: string; body: string; link?: ArtifactRef },
): Promise<ReviewSummary> {
  return request<ReviewSummary>(`/projects/${projectId}/reviews`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function deleteReview(projectId: number, id: string): Promise<void> {
  return request<void>(`/projects/${projectId}/reviews/${id}`, { method: 'DELETE' })
}

export function listDebts(projectId: number): Promise<DebtItemSummary[]> {
  return request<DebtItemSummary[]>(`/projects/${projectId}/debts`)
}

export function createDebt(
  projectId: number,
  input: { title: string; body: string; link?: ArtifactRef },
): Promise<DebtItemSummary> {
  return request<DebtItemSummary>(`/projects/${projectId}/debts`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function updateDebtStatus(projectId: number, id: string, status: DebtStatus): Promise<DebtItemSummary> {
  return request<DebtItemSummary>(`/projects/${projectId}/debts/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
}

export function deleteDebt(projectId: number, id: string): Promise<void> {
  return request<void>(`/projects/${projectId}/debts/${id}`, { method: 'DELETE' })
}

export function getOverview(projectId: number | null): Promise<Overview> {
  const qs = projectId != null ? `?projectId=${projectId}` : ''
  return request<Overview>(`/overview${qs}`)
}

export function listFlows(projectId: number | null): Promise<Flow[]> {
  const qs = projectId != null ? `?projectId=${projectId}` : ''
  return request<Flow[]>(`/specflow/flows${qs}`)
}

export function getFlow(id: number): Promise<Flow> {
  return request<Flow>(`/specflow/flows/${id}`)
}

export function createFlow(input: {
  projectId: number
  providerId: number
  scheduledAt: string
  changeNames: string[]
}): Promise<Flow> {
  return request<Flow>('/specflow/flows', { method: 'POST', body: JSON.stringify(input) })
}

export function reorderFlow(id: number, itemIds: number[]): Promise<void> {
  return request<void>(`/specflow/flows/${id}/reorder`, { method: 'PATCH', body: JSON.stringify({ itemIds }) })
}

export function deleteFlow(id: number): Promise<void> {
  return request<void>(`/specflow/flows/${id}`, { method: 'DELETE' })
}

export function cancelFlow(id: number): Promise<void> {
  return request<void>(`/specflow/flows/${id}/cancel`, { method: 'POST' })
}
