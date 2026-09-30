import { Moon, Sun } from 'lucide-react'
import type { Theme } from '@/lib/useTheme'
import { Button } from './ui/button'

interface Props {
  theme: Theme
  onToggle: () => void
}

export function ThemeToggle({ theme, onToggle }: Props) {
  return (
    <Button variant="outline" size="icon" onClick={onToggle} aria-label="Toggle theme">
      {theme === 'dark' ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
    </Button>
  )
}
