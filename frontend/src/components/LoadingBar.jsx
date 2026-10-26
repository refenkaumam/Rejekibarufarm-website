import { AnimatePresence, motion } from 'framer-motion'
import { useLoadingDelayed } from '../ui/store'
import WalkingCow from './illustrations/WalkingCow'

// Garis rumput tipis dengan sapi kecil berjalan selama ada request ke server
export default function LoadingBar() {
  const loading = useLoadingDelayed(300)
  return (
    <AnimatePresence>
      {loading && (
        <motion.div className="loadbar" role="progressbar" aria-label="Memuat data"
          initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}>
          <div className="loadbar-track" />
          <div className="loadbar-cow"><WalkingCow size={34} /></div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
