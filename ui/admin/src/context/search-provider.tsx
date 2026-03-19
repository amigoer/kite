import { createContext, useContext, useEffect, useState } from 'react'
import { CommandPalette } from '@/components/CommandPalette'

type SearchContextType = {
  open: boolean
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
}

const SearchContext = createContext<SearchContextType | null>(null)

type SearchProviderProps = {
  children: React.ReactNode
}

export function SearchProvider({ children }: SearchProviderProps) {
  const [open, setOpen] = useState(false)

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      // Allow user to use Cmd+K inside inputs without hijacking unless strictly needed, 
      // but CommandPalette.css/setup usually wants to hijack.
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
      }
    }
    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [])

  return (
    <SearchContext value={{ open, setOpen }}>
      {children}
      <CommandPalette />
    </SearchContext>
  )
}

export const useSearch = () => {
  const searchContext = useContext(SearchContext)
  if (!searchContext) {
    throw new Error('useSearch has to be used within SearchProvider')
  }
  return searchContext
}
