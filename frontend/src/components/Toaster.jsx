import { AnimatePresence, motion } from 'framer-motion'
import { useToasts, dismiss } from '../ui/store'
import { IconCheck, IconAlert, IconClose } from './icons'

export default function Toaster() {
  const toasts = useToasts()
  return (
    <div className="toaster" role="status" aria-live="polite">
      <AnimatePresence initial={false}>
        {toasts.map(t => (
          <motion.div key={t.id} layout className={'toast ' + t.type}
            initial={{ opacity: 0, y: 24, scale: .92 }} animate={{ opacity: 1, y: 0, scale: 1 }} exit={{ opacity: 0, x: 40, scale: .95 }}
            transition={{ type: 'spring', stiffness: 380, damping: 28 }}>
            <span className="toast-ico">{t.type === 'error' ? <IconAlert /> : <IconCheck />}</span>
            <span className="toast-text">{t.text}</span>
            <button type="button" className="toast-x" onClick={() => dismiss(t.id)} aria-label="Tutup notifikasi"><IconClose /></button>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  )
}
