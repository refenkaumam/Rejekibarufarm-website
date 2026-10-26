import { useEffect, useRef, useState } from 'react'
import { motion, useReducedMotion } from 'framer-motion'
import Cow from './illustrations/Cow'
import WalkingCow from './illustrations/WalkingCow'
import { useLoadingDelayed } from '../ui/store'
import {
  IconCow, IconWheat, IconMilk, IconCheese, IconWallet, IconSprout, IconArrowIn, IconArrowOut,
  IconScale, IconBarn, IconChart, IconSun, IconSunset, IconLeaf, IconChevronLeft, IconChevronRight,
} from './icons'

export const Field = ({ label, children }) => <label className="f"><span className="f-label">{label}</span>{children}</label>

// Ikon dan warna kartu statistik ditentukan dari labelnya
const STAT_RULES = [
  [/^pagi$/, IconSun, 'straw'],
  [/^sore$/, IconSunset, 'wood'],
  [/^bulan ini$/, IconMilk, 'sky'],
  [/laktasi/, IconCow, 'grass'],
  [/kering/, IconCow, 'straw'],
  [/dara/, IconCow, 'sky'],
  [/pedet/, IconCow, 'berry'],
  [/populasi|sapi/, IconCow, 'grass'],
  [/rumput/, IconWheat, 'grass'],
  [/konsentrat/, IconWheat, 'wood'],
  [/tebon/, IconWheat, 'straw'],
  [/mineral/, IconWheat, 'sky'],
  [/uang keluar/, IconArrowOut, 'brick'],
  [/biaya|dibayar|rugi/, IconArrowOut, 'brick'],
  [/uang masuk|pendapatan|hasil jual/, IconArrowIn, 'grass'],
  [/selisih/, IconScale, 'sky'],
  [/keju/, IconCheese, 'straw'],
  [/susu|liter|hari ini/, IconMilk, 'sky'],
  [/investor|modal/, IconSprout, 'straw'],
  [/kas awal|back office/, IconWallet, 'wood'],
  [/kandang/, IconBarn, 'grass'],
  [/melk/, IconMilk, 'sky'],
  [/laba/, IconChart, 'straw'],
  [/saldo|kas|total/, IconWallet, 'grass'],
]

function statMeta(label) {
  const l = String(label).toLowerCase()
  const hit = STAT_RULES.find(([re]) => re.test(l))
  return hit ? { Icon: hit[1], tone: hit[2] } : { Icon: IconLeaf, tone: 'grass' }
}

// Angka berformat id-ID (mis. "Rp 1.250.000" atau "12,5 L") dianimasikan naik
const NUM = /-?\d{1,3}(?:\.\d{3})+(?:,\d+)?|-?\d+(?:,\d+)?/

export function CountUp({ value }) {
  const reduce = useReducedMotion()
  const str = String(value)
  const m = str.match(NUM)
  const target = m ? Number(m[0].replace(/\./g, '').replace(',', '.')) : null
  const dec = m && m[0].includes(',') ? m[0].split(',')[1].length : 0
  const [shown, setShown] = useState(0)
  const last = useRef(0)

  useEffect(() => {
    if (target == null || reduce) return
    const from = last.current
    let raf, start
    const step = (t) => {
      start ??= t
      const k = Math.min(1, (t - start) / 900)
      const v = from + (target - from) * (1 - Math.pow(1 - k, 3))
      last.current = v
      setShown(v)
      if (k < 1) raf = requestAnimationFrame(step)
    }
    raf = requestAnimationFrame(step)
    return () => cancelAnimationFrame(raf)
  }, [target, reduce])

  if (target == null || reduce) return str
  const txt = str.slice(0, m.index) + shown.toLocaleString('id-ID', { minimumFractionDigits: dec, maximumFractionDigits: dec }) + str.slice(m.index + m[0].length)
  return <><span className="sr-only">{str}</span><span aria-hidden="true">{txt}</span></>
}

export function Stat({ label, value, hi }) {
  const { Icon, tone } = statMeta(label)
  return (
    <motion.div className={'stat tone-' + tone + (hi ? ' hi' : '')}
      initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} whileHover={{ y: -4 }} transition={{ duration: .35, ease: [.22, 1, .36, 1] }}>
      <span className="stat-icon"><Icon /></span>
      <span className="stat-label">{label}</span>
      <b className="stat-value"><CountUp value={value} /></b>
      <svg className="stat-spot" viewBox="0 0 80 60" aria-hidden="true">
        <path d="M18 8c12-8 34-6 40 8 6 15-6 30-22 30S6 40 6 28c0-8 4-15 12-20z" />
        <path d="M62 44c6-4 14-1 14 6s-8 9-13 6c-3-2-4-9-1-12z" />
      </svg>
    </motion.div>
  )
}

export function Empty({ text }) {
  const loading = useLoadingDelayed()
  if (loading) {
    return (
      <div className="skeleton" role="status" aria-live="polite">
        <div className="skeleton-head"><WalkingCow size={52} /><span>Sedang menggiring data...</span></div>
        {[0, 1, 2].map(i => <div key={i} className="skeleton-row" style={{ animationDelay: i * .12 + 's' }}><i /><i /><i /></div>)}
      </div>
    )
  }
  return (
    <div className="empty">
      <Cow pose="sleep" size={128} />
      <p>{text}</p>
    </div>
  )
}

export function Pager({ page, pages, setPage, total, size }) {
  if (pages <= 1) return null
  const from = (page - 1) * size + 1
  const to = Math.min(total, page * size)
  const start = Math.max(1, Math.min(page - 2, pages - 4))
  const nums = Array.from({ length: Math.min(5, pages) }, (_, i) => start + i)
  return (
    <nav className="pager" aria-label="Halaman tabel">
      <span className="pager-info">Menampilkan {from}-{to} dari {total}</span>
      <div className="pager-btns">
        <button type="button" className="ghost icon-btn" onClick={() => setPage(page - 1)} disabled={page === 1} aria-label="Halaman sebelumnya"><IconChevronLeft /></button>
        {nums.map(n => (
          <button type="button" key={n} className={n === page ? 'pager-num on' : 'ghost pager-num'} aria-current={n === page ? 'page' : undefined} onClick={() => setPage(n)}>{n}</button>
        ))}
        <button type="button" className="ghost icon-btn" onClick={() => setPage(page + 1)} disabled={page === pages} aria-label="Halaman berikutnya"><IconChevronRight /></button>
      </div>
    </nav>
  )
}
