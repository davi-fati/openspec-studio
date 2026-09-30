import { TABS, type Tab } from '@/lib/types'
import { cn } from '@/lib/utils'

interface Props {
  active: Tab
  onSelect: (tab: Tab) => void
}

export function TabNav({ active, onSelect }: Props) {
  return (
    <nav className="flex gap-0.5 rounded-lg border border-border bg-background p-0.5">
      {TABS.map((tab) => (
        <button
          key={tab}
          type="button"
          onClick={() => onSelect(tab)}
          className={cn(
            'rounded-md px-3.5 py-1.5 text-[13.5px] font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground',
            tab === active && 'bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground',
          )}
        >
          {tab}
        </button>
      ))}
    </nav>
  )
}
