// Ikon garis orisinal 24x24, stroke 1.8, mengikuti warna teks (currentColor)
function Svg({ children, size = '1em', ...p }) {
  return (
    <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" strokeWidth="1.8"
      strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" focusable="false" {...p}>
      {children}
    </svg>
  )
}

export function IconCow(p) {
  return <Svg {...p}>
    <path d="M7.2 9.6C5 10 3.3 9.2 2.6 7.6c1.7-.5 3.3-.1 4.6 1.1" />
    <path d="M16.8 9.6c2.2.4 3.9-.4 4.6-2-1.7-.5-3.3-.1-4.6 1.1" />
    <path d="M8.8 6.4 8 3.8M15.2 6.4l.8-2.6" />
    <path d="M7.2 12.8V10a4.8 4.8 0 0 1 9.6 0v2.8" />
    <rect x="6.4" y="12.6" width="11.2" height="7" rx="3.4" />
    <circle cx="10" cy="16.1" r=".7" fill="currentColor" stroke="none" />
    <circle cx="14" cy="16.1" r=".7" fill="currentColor" stroke="none" />
    <circle cx="9.7" cy="9.9" r=".8" fill="currentColor" stroke="none" />
    <circle cx="14.3" cy="9.9" r=".8" fill="currentColor" stroke="none" />
  </Svg>
}

export function IconWheat(p) {
  return <Svg {...p}>
    <path d="M12 21V6M12 6V3" />
    <path d="M12 9.5c-2.3-.2-3.6-1.8-3.7-4.1 2.3.2 3.6 1.7 3.7 4.1zM12 9.5c2.3-.2 3.6-1.8 3.7-4.1-2.3.2-3.6 1.7-3.7 4.1z" />
    <path d="M12 14c-2.3-.2-3.6-1.8-3.7-4.1 2.3.2 3.6 1.7 3.7 4.1zM12 14c2.3-.2 3.6-1.8 3.7-4.1-2.3.2-3.6 1.7-3.7 4.1z" />
    <path d="M12 18.5c-2.3-.2-3.6-1.8-3.7-4.1 2.3.2 3.6 1.7 3.7 4.1zM12 18.5c2.3-.2 3.6-1.8 3.7-4.1-2.3.2-3.6 1.7-3.7 4.1z" />
  </Svg>
}

export function IconMilk(p) {
  return <Svg {...p}>
    <path d="M9.4 3h5.2M10 3v2.6L8 8.7V19a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2V8.7l-2-3.1V3" />
    <path d="M8 12.5c1.3.8 2.7.8 4 0s2.7-.8 4 0" />
  </Svg>
}

export function IconTag(p) {
  return <Svg {...p}>
    <path d="M3.5 12.2V4.5a1 1 0 0 1 1-1h7.7l8.3 8.3a1.4 1.4 0 0 1 0 2l-6.6 6.6a1.4 1.4 0 0 1-2 0z" />
    <circle cx="8" cy="8" r="1.4" />
  </Svg>
}

export function IconSprout(p) {
  return <Svg {...p}>
    <path d="M12 20v-8M7 20h10" />
    <path d="M12 12c0-3.3-2.5-5.5-6-5.5 0 3.3 2.5 5.5 6 5.5zM12 10.5c0-3 2.2-5 5.5-5 0 3-2.2 5-5.5 5z" />
  </Svg>
}

export function IconWallet(p) {
  return <Svg {...p}>
    <path d="M4 7.5A2.5 2.5 0 0 1 6.5 5H17v3" />
    <path d="M4 7.5V17a2 2 0 0 0 2 2h13a1 1 0 0 0 1-1V9a1 1 0 0 0-1-1H6.5A2.5 2.5 0 0 1 4 5.5" />
    <circle cx="16.5" cy="13.5" r="1" fill="currentColor" stroke="none" />
  </Svg>
}

export function IconCheese(p) {
  return <Svg {...p}>
    <path d="M3 14.5 14.5 6l6.5 4.5v8a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1zM3 14.5h18" />
    <circle cx="8" cy="17" r="1" />
    <circle cx="14" cy="16.8" r="1.2" />
    <circle cx="17.6" cy="12.3" r=".7" />
  </Svg>
}

export function IconCalendar(p) {
  return <Svg {...p}>
    <rect x="4" y="5" width="16" height="15" rx="2" />
    <path d="M4 10h16M8 3v4M16 3v4M8 14h2M12 14h4M8 17h4" />
  </Svg>
}

export function IconChart(p) {
  return <Svg {...p}>
    <path d="M4 20h16" />
    <rect x="5.5" y="12" width="3" height="5" rx="1" />
    <rect x="10.5" y="7" width="3" height="10" rx="1" />
    <rect x="15.5" y="9.5" width="3" height="7.5" rx="1" />
  </Svg>
}

export function IconLedger(p) {
  return <Svg {...p}>
    <path d="M5 4.5h11a3 3 0 0 1 3 3V20H8a3 3 0 0 1-3-3z" />
    <path d="M5 17a3 3 0 0 1 3-3h11M9 8h6" />
  </Svg>
}

export function IconBarn(p) {
  return <Svg {...p}>
    <path d="M3 10 12 4l9 6M5 9v11h14V9" />
    <path d="M9.5 20v-6h5v6M9.5 14l5 6M14.5 14l-5 6" />
  </Svg>
}

export function IconUsers(p) {
  return <Svg {...p}>
    <circle cx="9" cy="8.5" r="3" />
    <path d="M3.5 19c.6-3 2.8-4.8 5.5-4.8s4.9 1.8 5.5 4.8M16 5.6a3 3 0 0 1 0 5.8M17.5 14.6c1.6.6 2.7 2.1 3 4.4" />
  </Svg>
}

export function IconUser(p) {
  return <Svg {...p}>
    <circle cx="12" cy="8" r="3.5" />
    <path d="M5 20c.8-3.6 3.6-5.6 7-5.6s6.2 2 7 5.6" />
  </Svg>
}

export function IconLock(p) {
  return <Svg {...p}>
    <rect x="5" y="10.5" width="14" height="9.5" rx="2.5" />
    <path d="M8 10.5V8a4 4 0 0 1 8 0v2.5" />
    <circle cx="12" cy="15.2" r="1.2" fill="currentColor" stroke="none" />
  </Svg>
}

export function IconLogout(p) {
  return <Svg {...p}><path d="M14 4h4a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-4M10 16l-4-4 4-4M6 12h10" /></Svg>
}

export function IconSun(p) {
  return <Svg {...p}>
    <circle cx="12" cy="12" r="4" />
    <path d="M12 2.5v2M12 19.5v2M4.6 4.6 6 6M18 18l1.4 1.4M2.5 12h2M19.5 12h2M4.6 19.4 6 18M18 6l1.4-1.4" />
  </Svg>
}

export function IconSunset(p) {
  return <Svg {...p}>
    <path d="M3.5 18h17M7 18a5 5 0 0 1 10 0M12 6.5v3M5.6 10.1l1.8 1.8M18.4 10.1l-1.8 1.8M8 21h8" />
  </Svg>
}

export function IconMoon(p) {
  return <Svg {...p}><path d="M20 14.5A8 8 0 1 1 9.5 4a6.5 6.5 0 0 0 10.5 10.5z" /></Svg>
}

export function IconMenu(p) {
  return <Svg {...p}><path d="M4 7h16M4 12h16M4 17h10" /></Svg>
}

export function IconClose(p) {
  return <Svg {...p}><path d="M6 6l12 12M18 6 6 18" /></Svg>
}

export function IconArrowIn(p) {
  return <Svg {...p}><path d="M12 4v11M7 10l5 5 5-5M5 20h14" /></Svg>
}

export function IconArrowOut(p) {
  return <Svg {...p}><path d="M12 15V4M7 9l5-5 5 5M5 20h14" /></Svg>
}

export function IconScale(p) {
  return <Svg {...p}>
    <path d="M12 4v16M5 7h14M8 20h8" />
    <path d="M5 7 2.5 13a2.5 2.5 0 0 0 5 0zM19 7l-2.5 6a2.5 2.5 0 0 0 5 0z" />
  </Svg>
}

export function IconCheck(p) {
  return <Svg {...p}><path d="M5 12.5l4.5 4.5L19 7.5" /></Svg>
}

export function IconAlert(p) {
  return <Svg {...p}>
    <circle cx="12" cy="12" r="9" />
    <path d="M12 7.5V13" />
    <circle cx="12" cy="16.5" r=".9" fill="currentColor" stroke="none" />
  </Svg>
}

export function IconChevronLeft(p) {
  return <Svg {...p}><path d="M15 6l-6 6 6 6" /></Svg>
}

export function IconChevronRight(p) {
  return <Svg {...p}><path d="M9 6l6 6-6 6" /></Svg>
}

export function IconLeaf(p) {
  return <Svg {...p}><path d="M5 19c0-8 5-13 14-14-1 9-6 14-14 14zM5 19l7-7" /></Svg>
}

export function IconHome(p) {
  return <Svg {...p}><path d="M4 11 12 4l8 7M6 9.5V20h12V9.5M10 20v-5h4v5" /></Svg>
}

export function IconArrowRight(p) {
  return <Svg {...p}><path d="M5 12h14M13 6l6 6-6 6" /></Svg>
}
