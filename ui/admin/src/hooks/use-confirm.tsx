import { useState, useCallback } from 'react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

interface ConfirmOptions {
  title?: string
  description: string
  confirmText?: string
  cancelText?: string
  variant?: 'default' | 'destructive'
}

interface ConfirmState extends ConfirmOptions {
  resolve: (value: boolean) => void
}

/**
 * 通用确认弹窗 Hook，替代 window.confirm
 *
 * 用法：
 * ```tsx
 * const { confirm, ConfirmDialog } = useConfirm()
 *
 * async function handleDelete() {
 *   if (await confirm({ description: '确定删除吗？' })) {
 *     // 执行删除
 *   }
 * }
 *
 * return <><ConfirmDialog />...</>
 * ```
 */
export function useConfirm() {
  const [state, setState] = useState<ConfirmState | null>(null)

  const confirm = useCallback((options: ConfirmOptions): Promise<boolean> => {
    return new Promise((resolve) => {
      setState({ ...options, resolve })
    })
  }, [])

  const handleClose = useCallback((result: boolean) => {
    state?.resolve(result)
    setState(null)
  }, [state])

  function ConfirmDialog() {
    return (
      <AlertDialog open={!!state} onOpenChange={(open) => { if (!open) handleClose(false) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{state?.title ?? '确认操作'}</AlertDialogTitle>
            <AlertDialogDescription>{state?.description}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => handleClose(false)}>
              {state?.cancelText ?? '取消'}
            </AlertDialogCancel>
            <AlertDialogAction
              variant={state?.variant ?? 'default'}
              onClick={() => handleClose(true)}
            >
              {state?.confirmText ?? '确定'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    )
  }

  return { confirm, ConfirmDialog }
}
