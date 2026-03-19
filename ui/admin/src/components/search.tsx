import { SearchIcon } from 'lucide-react'
import { useSearch } from '@/context/search-provider'
import { Button } from '@/components/ui/button'

/**
 * 搜索触发按钮 — 点击打开命令面板
 */
export function Search() {
  const { setOpen } = useSearch()
  return (
    <Button
      variant='outline'
      className='relative h-8 w-full justify-start rounded-md bg-muted/50 text-sm font-normal text-muted-foreground shadow-none sm:w-40 lg:w-56 xl:w-64'
      onClick={() => setOpen(true)}
    >
      <SearchIcon className='mr-2 h-4 w-4' />
      搜索…
      <kbd className='pointer-events-none absolute right-[0.3rem] top-[0.3rem] hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium opacity-100 sm:flex'>
        <span className='text-xs'>⌘</span>K
      </kbd>
    </Button>
  )
}
