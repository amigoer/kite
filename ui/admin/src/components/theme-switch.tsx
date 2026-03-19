import { Moon, Sun } from 'lucide-react'
import { useTheme } from '@/context/theme-provider'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface ThemeSwitchProps {
  className?: string
}

export function ThemeSwitch({ className }: ThemeSwitchProps) {
  const { theme, setTheme } = useTheme()

  return (
    <Button
      variant='ghost'
      size='icon'
      className={cn('rounded-full', className)}
      onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
    >
      {theme === 'dark' ? <Moon className='h-5 w-5' /> : <Sun className='h-5 w-5' />}
      <span className='sr-only'>切换主题</span>
    </Button>
  )
}
