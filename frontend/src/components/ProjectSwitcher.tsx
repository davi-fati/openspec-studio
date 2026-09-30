import { useState } from 'react'
import { FolderOpen, FolderPlus } from 'lucide-react'
import type { Project } from '@/lib/types'
import { Button } from './ui/button'
import { DirectoryBrowserDialog } from './DirectoryBrowserDialog'

interface Props {
  projects: Project[]
  activeId: number | null
  onSelect: (id: number | null) => void
  onOpen: (path: string) => Promise<void>
  onCreate: (path: string) => Promise<void>
}

export function ProjectSwitcher({ projects, activeId, onSelect, onOpen, onCreate }: Props) {
  const [browserMode, setBrowserMode] = useState<'open' | 'new' | null>(null)

  return (
    <div className="flex items-center gap-2">
      <select
        className="h-9 rounded-md border border-border bg-card px-2.5 text-[13px] text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
        value={activeId ?? 'all'}
        onChange={(e) => onSelect(e.target.value === 'all' ? null : Number(e.target.value))}
      >
        <option value="all">All projects</option>
        {projects.map((p) => (
          <option key={p.id} value={p.id} disabled={!p.available}>
            {p.name}
            {p.available ? '' : ' (unavailable)'}
          </option>
        ))}
      </select>

      <Button
        type="button"
        size="sm"
        variant="outline"
        title="Open an existing OpenSpec project"
        onClick={() => setBrowserMode('open')}
      >
        <FolderOpen className="h-3.5 w-3.5" />
        Open
      </Button>
      <Button
        type="button"
        size="sm"
        title="Scaffold a new project in an empty directory"
        onClick={() => setBrowserMode('new')}
      >
        <FolderPlus className="h-3.5 w-3.5" />
        New
      </Button>

      <DirectoryBrowserDialog
        open={browserMode !== null}
        mode={browserMode ?? 'open'}
        onClose={() => setBrowserMode(null)}
        onConfirm={browserMode === 'new' ? onCreate : onOpen}
      />
    </div>
  )
}
