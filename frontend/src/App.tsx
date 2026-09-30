import { useState } from 'react'
import { Settings } from 'lucide-react'
import { Logo } from './components/Logo'
import { TabNav } from './components/TabNav'
import { ProjectSwitcher } from './components/ProjectSwitcher'
import { ThemeToggle } from './components/ThemeToggle'
import { Button } from './components/ui/button'
import { ProviderSettingsDialog } from './components/providers/ProviderSettingsDialog'
import { Overview } from './pages/Overview'
import { Kanban } from './pages/Kanban'
import { Specs } from './pages/Specs'
import { Specflow } from './pages/Specflow'
import { ProjectsProvider, useProjects } from './lib/ProjectsContext'
import { NavigationProvider, useNavigation } from './lib/NavigationContext'
import { useTheme } from './lib/useTheme'
import { hasOverlayTitleBar, useWindowControlsInset } from './lib/desktop'
import { cn } from './lib/utils'
import { HighlightTheme } from './components/markdown/HighlightTheme'
import type { Tab } from './lib/types'

const PAGES: Record<Tab, React.ComponentType> = {
  Overview,
  Kanban,
  Specs,
  Specflow,
}

function AppShell() {
  const [theme, toggleTheme] = useTheme()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const { activeTab, setActiveTab } = useNavigation()
  const { projects, activeProjectId, setActiveProjectId, openProjectByPath, createProjectByPath } = useProjects()

  const ActivePage = PAGES[activeTab]
  const overlayTitleBar = hasOverlayTitleBar()
  const windowControlsInset = useWindowControlsInset()

  return (
    <div className="flex h-full flex-col">
      <HighlightTheme />
      {/* In the desktop app this header is the macOS title bar: it drags the
          window (controls excluded) and leaves room for the window buttons. */}
      <header
        data-tauri-drag-region={overlayTitleBar ? 'deep' : undefined}
        className={cn(
          'sticky top-0 z-10 flex h-14 flex-wrap items-center gap-6 border-b border-border bg-card px-6',
          overlayTitleBar && 'select-none',
          windowControlsInset && 'pl-[84px]',
        )}
      >
        <Logo onClick={() => setActiveTab('Overview')} />
        <TabNav active={activeTab} onSelect={setActiveTab} />
        <div className="ml-auto flex items-center gap-3">
          <ProjectSwitcher
            projects={projects}
            activeId={activeProjectId}
            onSelect={setActiveProjectId}
            onOpen={openProjectByPath}
            onCreate={createProjectByPath}
          />
          <Button variant="outline" size="icon" onClick={() => setSettingsOpen(true)} aria-label="AI provider settings">
            <Settings className="h-4 w-4" />
          </Button>
          <ThemeToggle theme={theme} onToggle={toggleTheme} />
        </div>
      </header>
      <main className="flex-1 overflow-auto px-6 py-8 sm:px-8 lg:px-10">
        <ActivePage />
      </main>
      <ProviderSettingsDialog open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </div>
  )
}

export default function App() {
  return (
    <ProjectsProvider>
      <NavigationProvider>
        <AppShell />
      </NavigationProvider>
    </ProjectsProvider>
  )
}
