import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'

interface ConfirmState {
  title: string
  message: string
  confirmLabel: string
  cancelLabel: string
  variant: 'default' | 'danger'
}

interface ConfirmOptions {
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'default' | 'danger'
}

const defaultState: ConfirmState = {
  title: '',
  message: '',
  confirmLabel: 'Подтвердить',
  cancelLabel: 'Отмена',
  variant: 'default',
}

export function useConfirmDialog() {
  const [state, setState] = useState<ConfirmState>(defaultState)
  const [open, setOpen] = useState(false)
  const resolverRef = useRef<((value: boolean) => void) | null>(null)

  useEffect(() => {
    if (!open) {
      return undefined
    }

    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setOpen(false)
        resolverRef.current?.(false)
        resolverRef.current = null
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => {
      document.body.style.overflow = previousOverflow
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [open])

  function close(value: boolean) {
    setOpen(false)
    resolverRef.current?.(value)
    resolverRef.current = null
  }

  function confirm(options: ConfirmOptions) {
    setState({
      title: options.title,
      message: options.message,
      confirmLabel: options.confirmLabel || 'Подтвердить',
      cancelLabel: options.cancelLabel || 'Отмена',
      variant: options.variant || 'default',
    })
    setOpen(true)

    return new Promise<boolean>((resolve) => {
      resolverRef.current = resolve
    })
  }

  const dialogContent = open ? (
    <div className="modal-backdrop" role="presentation" onClick={() => close(false)}>
      <div
        className="modal-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="modal-content">
          <h3 id="confirm-dialog-title">{state.title}</h3>
          <p>{state.message}</p>
        </div>
        <div className="button-row">
          <button type="button" className="ghost-button" onClick={() => close(false)}>
            {state.cancelLabel}
          </button>
          <button
            type="button"
            className={state.variant === 'danger' ? 'danger-button' : 'primary-button'}
            onClick={() => close(true)}
          >
            {state.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  ) : null

  const dialog = dialogContent ? createPortal(dialogContent, document.body) : null

  return { confirm, dialog }
}
