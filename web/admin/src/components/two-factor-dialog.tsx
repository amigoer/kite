import { useEffect, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Copy, Loader2, ShieldCheck, ShieldOff } from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { toast } from 'sonner'

import { authApi } from '@/lib/api'
import { useI18n } from '@/i18n'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface EnrollmentPayload {
  secret: string
  uri: string
}

interface EnableTwoFactorDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

// Formats the raw Base32 secret in groups of 4 so the user can read
// it off the screen into an authenticator app that doesn't support
// QR scanning — matches the grouping most apps use for manual entry.
function formatSecret(secret: string): string {
  return secret.replace(/(.{4})/g, '$1 ').trim()
}

export function EnableTwoFactorDialog({
  open,
  onOpenChange,
  onSuccess,
}: EnableTwoFactorDialogProps) {
  const { t } = useI18n()
  const [step, setStep] = useState<'scan' | 'confirm'>('scan')
  const [enrollment, setEnrollment] = useState<EnrollmentPayload | null>(null)
  const [code, setCode] = useState('')
  const codeInputRef = useRef<HTMLInputElement | null>(null)

  const setupMutation = useMutation({
    mutationFn: () => authApi.setupTotp(),
    onSuccess: (res) => {
      const data = res.data.data as EnrollmentPayload
      setEnrollment(data)
    },
    onError: () => {
      toast.error(t('twoFactor.setupFailed'))
      onOpenChange(false)
    },
  })

  const enableMutation = useMutation({
    mutationFn: (c: string) => authApi.enableTotp(c),
    onSuccess: () => {
      toast.success(t('twoFactor.enableSuccess'))
      onSuccess()
      onOpenChange(false)
    },
    onError: (err: unknown) => {
      const backendMsg = (err as { response?: { data?: { message?: string } } })
        ?.response?.data?.message
      toast.error(backendMsg ?? t('twoFactor.codeInvalid'))
      setCode('')
      codeInputRef.current?.focus()
    },
  })

  // Kick off enrollment as soon as the dialog opens; the secret lives
  // on the user row in "pending" state and is promoted to active only
  // when the user confirms with a valid code. Re-opening the dialog
  // generates a fresh secret, which is what we want if the user
  // bailed halfway through last time.
  useEffect(() => {
    if (open) {
      setStep('scan')
      setCode('')
      setEnrollment(null)
      setupMutation.mutate()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const handleCopy = async () => {
    if (!enrollment) return
    try {
      await navigator.clipboard.writeText(enrollment.secret)
      toast.success(t('twoFactor.secretCopied'))
    } catch {
      toast.error(t('twoFactor.secretCopyFailed'))
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (code.length !== 6) return
    enableMutation.mutate(code)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('twoFactor.enableTitle')}</DialogTitle>
          <DialogDescription>
            {step === 'scan'
              ? t('twoFactor.enableScanDesc')
              : t('twoFactor.enableConfirmDesc')}
          </DialogDescription>
        </DialogHeader>

        {setupMutation.isPending && (
          <div className="flex h-48 items-center justify-center">
            <Loader2 className="size-5 animate-spin text-muted-foreground" />
          </div>
        )}

        {enrollment && step === 'scan' && (
          <div className="space-y-4">
            <div className="flex justify-center rounded-lg border bg-white p-4">
              <QRCodeSVG value={enrollment.uri} size={180} />
            </div>
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground">
                {t('twoFactor.manualKey')}
              </Label>
              <div className="flex items-center gap-2 rounded-md border bg-muted/40 px-3 py-2">
                <code className="flex-1 font-mono text-xs tracking-wider">
                  {formatSecret(enrollment.secret)}
                </code>
                <Button
                  type="button"
                  size="icon-xs"
                  variant="ghost"
                  onClick={handleCopy}
                >
                  <Copy className="size-3.5" />
                </Button>
              </div>
              <p className="text-[11px] text-muted-foreground">
                {t('twoFactor.manualKeyHint')}
              </p>
            </div>
            <Button
              type="button"
              className="w-full"
              onClick={() => {
                setStep('confirm')
                setTimeout(() => codeInputRef.current?.focus(), 0)
              }}
            >
              {t('twoFactor.scannedContinue')}
            </Button>
          </div>
        )}

        {enrollment && step === 'confirm' && (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="enable-totp-code">
                {t('twoFactor.enterCode')}
              </Label>
              <Input
                id="enable-totp-code"
                ref={codeInputRef}
                inputMode="numeric"
                autoComplete="one-time-code"
                maxLength={6}
                value={code}
                onChange={(e) =>
                  setCode(e.target.value.replace(/\D/g, '').slice(0, 6))
                }
                placeholder="123456"
                className="text-center text-lg tracking-[0.5em]"
                autoFocus
                required
              />
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setStep('scan')}
                disabled={enableMutation.isPending}
              >
                {t('twoFactor.back')}
              </Button>
              <Button
                type="submit"
                disabled={enableMutation.isPending || code.length !== 6}
              >
                {enableMutation.isPending && (
                  <Loader2 className="size-4 animate-spin" />
                )}
                <ShieldCheck className="size-4" />
                {t('twoFactor.confirmEnable')}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}

interface DisableTwoFactorDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export function DisableTwoFactorDialog({
  open,
  onOpenChange,
  onSuccess,
}: DisableTwoFactorDialogProps) {
  const { t } = useI18n()
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')

  const disableMutation = useMutation({
    mutationFn: () => authApi.disableTotp(password, code),
    onSuccess: () => {
      toast.success(t('twoFactor.disableSuccess'))
      onSuccess()
      onOpenChange(false)
    },
    onError: (err: unknown) => {
      const backendMsg = (err as { response?: { data?: { message?: string } } })
        ?.response?.data?.message
      toast.error(backendMsg ?? t('twoFactor.disableFailed'))
    },
  })

  // Reset fields whenever the dialog toggles open so the last attempt's
  // values don't linger for the next invocation.
  useEffect(() => {
    if (open) {
      setPassword('')
      setCode('')
    }
  }, [open])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!password || code.length !== 6) return
    disableMutation.mutate()
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('twoFactor.disableTitle')}</DialogTitle>
          <DialogDescription>{t('twoFactor.disableDesc')}</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="disable-totp-password">
              {t('profile.currentPassword')}
            </Label>
            <Input
              id="disable-totp-password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="disable-totp-code">
              {t('twoFactor.enterCode')}
            </Label>
            <Input
              id="disable-totp-code"
              inputMode="numeric"
              autoComplete="one-time-code"
              maxLength={6}
              value={code}
              onChange={(e) =>
                setCode(e.target.value.replace(/\D/g, '').slice(0, 6))
              }
              placeholder="123456"
              className="text-center text-lg tracking-[0.5em]"
              required
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={disableMutation.isPending}
            >
              {t('profile.cancel')}
            </Button>
            <Button
              type="submit"
              variant="destructive"
              disabled={
                disableMutation.isPending || !password || code.length !== 6
              }
            >
              {disableMutation.isPending && (
                <Loader2 className="size-4 animate-spin" />
              )}
              <ShieldOff className="size-4" />
              {t('twoFactor.confirmDisable')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
