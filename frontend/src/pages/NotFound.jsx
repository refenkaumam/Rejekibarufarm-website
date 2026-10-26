import { motion } from 'framer-motion'
import Cow from '../components/illustrations/Cow'
import Pasture from '../components/illustrations/Pasture'
import ThemeToggle from '../components/ThemeToggle'
import { IconHome } from '../components/icons'
import { useTheme } from '../ui/theme'
import { phaseOf } from '../ui/time'

export default function NotFound() {
  const theme = useTheme()
  const phase = theme === 'dark' ? 'malam' : phaseOf()
  return (
    <div className="nf-page">
      <Pasture phase={phase} sleeper={false} />
      <ThemeToggle className="login-theme" />
      <motion.div className="nf-card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: .5 }}>
        <p className="nf-code">404</p>
        <h1>Waduh, sapinya nyasar!</h1>
        <p>Halaman yang kamu cari tidak ada di padang rumput ini. Mungkin sudah pindah kandang.</p>
        <a className="btn" href={import.meta.env.BASE_URL}><IconHome />Kembali ke beranda</a>
      </motion.div>
      <div className="nf-cow"><Cow pose="confused" size="100%" label="Sapi yang kebingungan" /></div>
    </div>
  )
}
