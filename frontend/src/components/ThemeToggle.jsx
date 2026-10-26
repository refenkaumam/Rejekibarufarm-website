import { AnimatePresence, motion } from 'framer-motion'
import { useTheme, toggleTheme } from '../ui/theme'
import { IconSun, IconMoon } from './icons'

export default function ThemeToggle({ className = '' }) {
  const theme = useTheme()
  const dark = theme === 'dark'
  return (
    <button type="button" className={'theme-toggle ' + className} onClick={toggleTheme}
      aria-label={dark ? 'Ganti ke tema siang' : 'Ganti ke tema padang rumput malam'} title={dark ? 'Tema siang' : 'Tema malam'}>
      <AnimatePresence mode="wait" initial={false}>
        <motion.span key={theme} className="theme-toggle-ico"
          initial={{ rotate: -90, scale: .4, opacity: 0 }} animate={{ rotate: 0, scale: 1, opacity: 1 }} exit={{ rotate: 90, scale: .4, opacity: 0 }}
          transition={{ duration: .25 }}>
          {dark ? <IconMoon /> : <IconSun />}
        </motion.span>
      </AnimatePresence>
    </button>
  )
}
