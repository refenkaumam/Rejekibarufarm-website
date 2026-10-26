import { useEffect, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { getUser, fetchMe, clearSession } from './api/auth'
import { MENU } from './config/menu'
import Sidebar from './components/Sidebar'
import Header from './components/Header'
import { Empty } from './components/ui'
import Login from './pages/auth/Login'

const TanpaMenu = () => <Empty text="Menu untuk akun ini belum tersedia." />

export default function App() {
  const [user, setUser] = useState(getUser())
  const [page, setPage] = useState(null)
  const [navOpen, setNavOpen] = useState(false)
  useEffect(() => { if (getUser()) fetchMe().then(setUser).catch(() => {}) }, [])
  if (!user) return <Login onLogin={setUser} />

  const boleh = (k) => (user.menus || []).includes(k) || (k === 'akun' && user.role === 'super_admin')
  const menu = MENU.map(g => ({ ...g, items: g.items.filter(i => boleh(i[0])) })).filter(g => g.items.length)
  const aktif = menu.flatMap(g => g.items).find(i => i[0] === page) || menu.flatMap(g => g.items)[0]
  const Page = aktif ? aktif[2] : TanpaMenu
  const keluar = () => { clearSession(); setUser(null) }

  return (
    <div className="shell">
      <Sidebar user={user} menu={menu} aktif={aktif && aktif[0]} open={navOpen} onClose={() => setNavOpen(false)}
        onSelect={(k) => { setPage(k); setNavOpen(false) }} onLogout={keluar} />
      <AnimatePresence>
        {navOpen && <motion.div className="side-backdrop" onClick={() => setNavOpen(false)}
          initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} />}
      </AnimatePresence>
      <div className="content">
        <Header user={user} onMenu={() => setNavOpen(true)} />
        <AnimatePresence mode="wait">
          <motion.main key={aktif ? aktif[0] : 'kosong'} className="main"
            initial={{ opacity: 0, y: 14 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -8 }}
            transition={{ duration: .28, ease: [.22, 1, .36, 1] }}>
            <Page />
          </motion.main>
        </AnimatePresence>
      </div>
    </div>
  )
}
