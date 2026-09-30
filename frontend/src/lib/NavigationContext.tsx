import { createContext, useContext, useState, type ReactNode } from 'react'
import type { Tab } from './types'

interface NavigationContextValue {
  activeTab: Tab
  setActiveTab: (tab: Tab) => void
}

const NavigationContext = createContext<NavigationContextValue | null>(null)

export function NavigationProvider({ children }: { children: ReactNode }) {
  const [activeTab, setActiveTab] = useState<Tab>('Overview')
  return <NavigationContext.Provider value={{ activeTab, setActiveTab }}>{children}</NavigationContext.Provider>
}

export function useNavigation() {
  const ctx = useContext(NavigationContext)
  if (!ctx) throw new Error('useNavigation must be used within a NavigationProvider')
  return ctx
}
