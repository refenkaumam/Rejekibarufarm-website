import { motion } from 'framer-motion'
import CowMark from './illustrations/CowMark'
import {
  IconCow, IconWheat, IconMilk, IconTag, IconSprout, IconWallet, IconCheese, IconCalendar,
  IconChart, IconLedger, IconUsers, IconLogout, IconClose, IconLeaf,
} from './icons'

const ICON = {
  sapi: IconCow, pakan: IconWheat, produksi: IconMilk, penjualan: IconTag, investor: IconSprout, kas: IconWallet,
  msusu: IconMilk, mkeju: IconCheese, mrekap: IconCalendar, mkas: IconWallet, bo: IconChart, bokas: IconLedger, akun: IconUsers,
}

const ROLE = { super_admin: 'Super admin', owner: 'Owner', peternakan: 'Peternakan', melk: 'Melk by RBF', back_office: 'Back office' }

export default function Sidebar({ user, menu, aktif, onSelect, onLogout, open, onClose }) {
  return (
    <aside className={'side' + (open ? ' open' : '')} aria-label="Navigasi utama">
      <div className="brand">
        <CowMark size={42} />
        <div className="brand-text">Rejeki Baru Farm<small>Pembukuan peternakan</small></div>
        <button type="button" className="icon-btn side-close" onClick={onClose} aria-label="Tutup menu"><IconClose /></button>
      </div>
      <nav className="side-nav">
        {menu.map(g => (
          <div key={g.grup} className="grp-wrap">
            <div className="grp">{g.grup}</div>
            {g.items.map(([k, label]) => {
              const Icon = ICON[k] || IconLeaf
              const on = aktif === k
              return (
                <button key={k} type="button" className={'nav' + (on ? ' on' : '')} onClick={() => onSelect(k)} aria-current={on ? 'page' : undefined}>
                  {on && <motion.span layoutId="nav-pill" className="nav-pill" transition={{ type: 'spring', stiffness: 500, damping: 38 }} />}
                  <span className="nav-ico"><Icon /></span>
                  <span className="nav-label">{label}</span>
                </button>
              )
            })}
          </div>
        ))}
      </nav>
      <svg className="side-deco" viewBox="0 0 264 70" preserveAspectRatio="none" aria-hidden="true">
        <path d="M0 40 L264 40 M0 56 L264 56" stroke="var(--side-accent)" strokeWidth="3" opacity=".35" />
        {[14, 62, 110, 158, 206, 254].map(x => <rect key={x} x={x - 4} y="28" width="8" height="42" rx="2" fill="var(--side-accent)" opacity=".35" />)}
        <path d="M0 70 C 30 58, 50 62, 70 70 C 100 56, 130 60, 160 70 C 190 58, 220 60, 264 70 Z" fill="var(--side-bg-2)" />
      </svg>
      <div className="who">
        <span className="who-avatar" aria-hidden="true">{(user.username || '?').slice(0, 1).toUpperCase()}</span>
        <span className="who-text"><b>{user.username}</b><small>{ROLE[user.role] || user.role}</small></span>
        <button type="button" className="icon-btn who-out" onClick={onLogout} aria-label="Keluar" title="Keluar"><IconLogout /></button>
      </div>
    </aside>
  )
}
