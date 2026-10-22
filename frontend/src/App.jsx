import { useEffect, useState } from 'react'
import { getUser, fetchMe, clearSession } from './api/auth'
import { MENU } from './config/menu'
import Sidebar from './components/Sidebar'
import { Empty } from './components/ui'
import Login from './pages/auth/Login'

export default function App() {
  const [user, setUser] = useState(getUser())
  const [page, setPage] = useState(null)
  useEffect(() => { if (getUser()) fetchMe().then(setUser).catch(() => {}) }, [])
  if (!user) return <Login onLogin={setUser} />

  const boleh = (k) => (user.menus || []).includes(k) || (k === 'akun' && user.role === 'super_admin')
  const menu = MENU.map(g => ({ ...g, items: g.items.filter(i => boleh(i[0])) })).filter(g => g.items.length)
  const aktif = menu.flatMap(g => g.items).find(i => i[0] === page) || menu.flatMap(g => g.items)[0]
  const Page = aktif ? aktif[2] : () => <Empty text="Menu untuk akun ini belum tersedia." />
  const keluar = () => { clearSession(); setUser(null) }

  return (
    <div className="shell">
      <Sidebar user={user} menu={menu} aktif={aktif && aktif[0]} onSelect={setPage} onLogout={keluar} />
      <main className="main"><Page /></main>
    </div>
  )
}
