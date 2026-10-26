// Sapi kecil tampak samping yang berjalan (indikator loading)
export default function WalkingCow({ size = 56, className = '' }) {
  const leg = { width: 5, height: 14, rx: 2.4, fill: 'var(--cow-white)', stroke: 'var(--cow-outline)', strokeWidth: 1.6 }
  return (
    <svg className={'walk-cow ' + className} viewBox="0 0 64 44" width={size} height={size * 44 / 64} aria-hidden="true">
      <ellipse cx="32" cy="41.5" rx="22" ry="2" fill="var(--cow-shadow)" />
      <rect className="leg leg-b" x="15" y="25" {...leg} />
      <rect className="leg leg-b" x="38" y="25" {...leg} />
      <rect className="leg leg-a" x="20" y="25" {...leg} />
      <rect className="leg leg-a" x="43" y="25" {...leg} />
      <g className="walk-body">
        <path className="walk-tail" d="M11 16 C 6 16, 4 20, 5 26" fill="none" stroke="var(--cow-outline)" strokeWidth="1.8" strokeLinecap="round" />
        <rect x="10" y="10" width="38" height="20" rx="10" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="1.8" />
        <path d="M18 10.5c4 0 7 3 6 7-1 3-5 4-8 2-2-2-2-6 2-9z" fill="var(--cow-spot)" />
        <path d="M36 20c3-1 6 1 6 4s-3 5-6 4c-2-1-3-6 0-8z" fill="var(--cow-spot)" />
        <ellipse cx="47" cy="9" rx="4" ry="2" transform="rotate(-25 47 9)" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="1.4" />
        <rect x="44" y="6" width="14" height="13" rx="6" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="1.8" />
        <rect x="52" y="12" width="9" height="8" rx="4" fill="var(--cow-muzzle)" stroke="var(--cow-outline)" strokeWidth="1.4" />
        <circle cx="51" cy="10.5" r="1.3" fill="var(--cow-pupil)" />
        <path d="M49 6.5c-.5-2 0-3.5 1.5-4.3" fill="none" stroke="var(--cow-outline)" strokeWidth="1.4" strokeLinecap="round" />
      </g>
    </svg>
  )
}
