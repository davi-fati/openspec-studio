import { Archive, CheckCircle2, Lightbulb, ListTodo, Loader } from 'lucide-react'
import type { ChangeStatus } from '@/lib/types'

// Same lucide icons + colors as openspec-ui's KanbanBoard (Todo/In
// Progress/Done/Archived); Draft has no equivalent there, so it borrows
// their "Ideas" icon/color (Lightbulb, violet) as the closest fit -
// proposal-only, nothing started yet.
export const COLUMN_ICON: Record<ChangeStatus, React.ReactNode> = {
  draft: <Lightbulb className="h-4 w-4 text-violet-500" />,
  todo: <ListTodo className="h-4 w-4 text-blue-500" />,
  in_progress: <Loader className="h-4 w-4 text-amber-500" />,
  done: <CheckCircle2 className="h-4 w-4 text-emerald-500" />,
  archived: <Archive className="h-4 w-4 text-gray-500" />,
}
