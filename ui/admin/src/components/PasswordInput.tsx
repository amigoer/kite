import * as React from 'react'
import { Eye, EyeOff } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from './ui/button'
import { Input } from './ui/input'

type PasswordInputProps = Omit<
  React.InputHTMLAttributes<HTMLInputElement>,
  'type'
> & {
  ref?: React.Ref<HTMLInputElement>
}

/**
 * 密码输入框 — 支持点击眼睛图标切换明文/密文
 */
export function PasswordInput({
  className,
  disabled,
  ref,
  ...props
}: PasswordInputProps) {
  const [showPassword, setShowPassword] = React.useState(false)

  return (
    <div className={cn('relative w-full', className)}>
      <Input
        type={showPassword ? 'text' : 'password'}
        className='pr-9'
        ref={ref}
        disabled={disabled}
        {...props}
      />
      <Button
        type='button'
        size='icon'
        variant='ghost'
        disabled={disabled}
        className='absolute end-1 top-1/2 h-6 w-6 -translate-y-1/2 rounded-md text-muted-foreground hover:bg-transparent'
        onClick={() => setShowPassword((prev) => !prev)}
      >
        {showPassword ? <Eye size={16} /> : <EyeOff size={16} />}
      </Button>
    </div>
  )
}
