import { Command } from 'lucide-react'

/**
 * Studio mark (⌘ on the brand gradient) plus wordmark. Same artwork as
 * public/favicon.svg and the desktop icon (src-tauri/icons/source.svg).
 */
export function Logo({ onClick }: { onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      title="Go to Overview"
      className="flex items-center gap-2 whitespace-nowrap rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <span className="grid h-7 w-7 shrink-0 place-items-center rounded-lg bg-gradient-to-b from-[#fc9e47] to-[#d84a00] text-white shadow-sm">
        <Command className="h-4 w-4" strokeWidth={2.6} />
      </span>
      <span className="text-sm font-bold tracking-tight">
        OpenSpec <span className="text-primary">Studio</span>
      </span>
    </button>
  )
}
