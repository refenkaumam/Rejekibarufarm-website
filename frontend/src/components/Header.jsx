import ThemeToggle from './ThemeToggle'
import CowMark from './illustrations/CowMark'
import LoadingBar from './LoadingBar'
import { IconMenu, IconSun, IconSunset, IconMoon } from './icons'
import { phaseOf, sapaan, tanggalPanjang } from '../ui/time'

const PHASE_ICON = { pagi: IconSun, siang: IconSun, sore: IconSunset, malam: IconMoon }

export default function Header({ user, onMenu }) {
  const now = new Date()
  const phase = phaseOf(now)
  const Icon = PHASE_ICON[phase]
  return (
    <header className="topbar">
      <button type="button" className="ghost icon-btn menu-btn" onClick={onMenu} aria-label="Buka menu"><IconMenu /></button>
      <div className="greet" data-phase={phase}>
        <span className="greet-art" aria-hidden="true">
          <span className="greet-sky"><Icon /></span>
          <CowMark size={34} />
        </span>
        <div>
          <p className="greet-title">{sapaan(now)}, <b>{user.username}</b></p>
          <p className="greet-date">{tanggalPanjang(now)}</p>
        </div>
      </div>
      <ThemeToggle />
      <LoadingBar />
    </header>
  )
}
