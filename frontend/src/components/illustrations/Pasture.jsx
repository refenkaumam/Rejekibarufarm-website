import { useEffect } from 'react'
import { motion, useMotionValue, useSpring, useTransform, useReducedMotion } from 'framer-motion'
import Cow from './Cow'

// Latar padang rumput berlapis: langit, matahari/bulan, awan, bukit, lumbung, pagar, rumput.
// phase: pagi | siang | sore | malam

const W = 1440
const H = 900
const far = (x) => 505 + 32 * Math.sin(x / 260 + .4) + 14 * Math.sin(x / 97)
const mid = (x) => 628 + 24 * Math.sin(x / 300 + 1.2) + 8 * Math.sin(x / 120)
const near = (x) => 738 + 20 * Math.sin(x / 230 + 2.1)

const hill = (fn) => {
  let d = `M-80 ${H} L-80 ${fn(-80).toFixed(1)}`
  for (let x = -56; x <= W + 80; x += 24) d += ` L${x} ${fn(x).toFixed(1)}`
  return d + ` L${W + 80} ${H} Z`
}

function rng(seed) {
  return () => {
    seed |= 0; seed = (seed + 0x6D2B79F5) | 0
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const r = rng(7)
const HILLS = { far: hill(far), mid: hill(mid), near: hill(near) }

const BLADES = Array.from({ length: 110 }, (_, i) => {
  const x = -40 + r() * (W + 80), h = 28 + r() * 50, w = 5 + r() * 5, lean = (r() - .5) * 16
  return {
    d: `M${x - w / 2} ${H + 2} Q${x + lean * .3} ${H - h * .6} ${x + lean} ${H - h} Q${x + lean * .4 + w * .2} ${H - h * .5} ${x + w / 2} ${H + 2} Z`,
    c: i % 3 ? 'var(--grass-blade)' : 'var(--grass-blade-2)',
    delay: -(r() * 4).toFixed(2) + 's', dur: (3 + r() * 2.4).toFixed(2) + 's',
  }
})

const TUFTS = Array.from({ length: 46 }, () => {
  const x = r() * W, y = near(x) + 14 + r() * 120, h = 9 + r() * 9
  return { d: `M${x - 6} ${y} q2 -${h * .7} 3 -${h} M${x} ${y} q0 -${h} 1 -${h * 1.2} M${x + 6} ${y} q-2 -${h * .7} -3 -${h}`, delay: -(r() * 4).toFixed(2) + 's' }
})

const FLOWERS = Array.from({ length: 22 }, (_, i) => {
  const x = r() * W
  return { x, y: near(x) + 22 + r() * 110, c: ['var(--flower-1)', 'var(--flower-2)', 'var(--flower-3)'][i % 3] }
})

const STARS = Array.from({ length: 80 }, () => ({ x: r() * W, y: r() * 430, s: .6 + r() * 1.4, delay: -(r() * 5).toFixed(2) + 's' }))

const FIREFLIES = Array.from({ length: 12 }, () => ({ x: 80 + r() * (W - 160), y: 660 + r() * 180, delay: -(r() * 6).toFixed(2) + 's' }))

const CLOUDS = [
  { x: 90, y: 120, s: 1.15, dur: '120s', delay: '-10s' },
  { x: 520, y: 70, s: .8, dur: '150s', delay: '-60s' },
  { x: 900, y: 150, s: 1, dur: '135s', delay: '-35s' },
  { x: 1250, y: 95, s: .7, dur: '165s', delay: '-95s' },
  { x: 300, y: 230, s: .6, dur: '180s', delay: '-130s' },
]

const FENCE = Array.from({ length: 24 }, (_, i) => -20 + i * 64)

const TREES = [150, 226, 1236, 1312]

const SUN = { pagi: [560, 330], siang: [1060, 170], sore: [620, 455], malam: [980, 160] }

function Cloud({ x, y, s, dur, delay }) {
  return (
    <g className="cloud" style={{ animationDuration: dur, animationDelay: delay }}>
      <g transform={`translate(${x} ${y}) scale(${s})`}>
        <ellipse cx="0" cy="12" rx="74" ry="20" fill="var(--cloud-shade)" />
        <ellipse cx="-28" cy="0" rx="40" ry="24" fill="var(--cloud)" />
        <ellipse cx="12" cy="-14" rx="44" ry="32" fill="var(--cloud)" />
        <ellipse cx="50" cy="2" rx="40" ry="22" fill="var(--cloud)" />
      </g>
    </g>
  )
}

export default function Pasture({ phase = 'siang', className = '', sleeper }) {
  const reduce = useReducedMotion()
  const mx = useMotionValue(0)
  const my = useMotionValue(0)
  const sx = useSpring(mx, { stiffness: 60, damping: 20 })
  const sy = useSpring(my, { stiffness: 60, damping: 20 })
  const skyX = useTransform(sx, v => v * -4)
  const farX = useTransform(sx, v => v * -9)
  const midX = useTransform(sx, v => v * -16)
  const nearX = useTransform(sx, v => v * -26)
  const frontX = useTransform(sx, v => v * -36)
  const farY = useTransform(sy, v => v * -3)
  const nearY = useTransform(sy, v => v * -6)

  useEffect(() => {
    if (reduce) return
    const move = (e) => {
      if (e.pointerType && e.pointerType !== 'mouse') return
      mx.set((e.clientX / window.innerWidth - .5) * 2)
      my.set((e.clientY / window.innerHeight - .5) * 2)
    }
    window.addEventListener('pointermove', move)
    return () => window.removeEventListener('pointermove', move)
  }, [reduce, mx, my])

  const night = phase === 'malam'
  const [sunX, sunY] = SUN[phase] || SUN.siang
  const showSleeper = sleeper ?? night

  return (
    <div className={'scene ' + className} data-phase={phase} aria-hidden="true">
      <svg className="scene-svg" viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="xMidYMax slice">
        <defs>
          <linearGradient id="sky" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" style={{ stopColor: 'var(--sky-top)' }} />
            <stop offset="1" style={{ stopColor: 'var(--sky-bottom)' }} />
          </linearGradient>
          <radialGradient id="glow">
            <stop offset="0" style={{ stopColor: 'var(--glow)' }} />
            <stop offset="1" style={{ stopColor: 'var(--glow)', stopOpacity: 0 }} />
          </radialGradient>
          <mask id="moon-cut">
            <rect width={W} height={H} fill="#fff" />
            <circle cx={sunX + 18} cy={sunY - 10} r="34" fill="#000" />
          </mask>
        </defs>

        <rect x="-100" y="-100" width={W + 200} height={H + 200} fill="url(#sky)" />

        <motion.g style={{ x: skyX }}>
          {night && STARS.map((s, i) => <circle key={i} className="star" cx={s.x} cy={s.y} r={s.s} fill="var(--star)" style={{ animationDelay: s.delay }} />)}
          <circle cx={sunX} cy={sunY} r="420" fill="url(#glow)" />
          {night
            ? <circle cx={sunX} cy={sunY} r="40" fill="var(--moon)" mask="url(#moon-cut)" />
            : <g>
              <g className="sun-rays" style={{ transformOrigin: `${sunX}px ${sunY}px` }} fill="var(--sun)" opacity=".18">
                {Array.from({ length: 12 }, (_, i) => (
                  <path key={i} d={`M${sunX} ${sunY} L${sunX - 16} ${sunY - 200} L${sunX + 16} ${sunY - 200} Z`} transform={`rotate(${i * 30} ${sunX} ${sunY})`} />
                ))}
              </g>
              <circle cx={sunX} cy={sunY} r="58" fill="var(--sun)" />
              <circle cx={sunX} cy={sunY} r="44" fill="var(--sun-core)" />
            </g>}
          {CLOUDS.map((c, i) => <Cloud key={i} {...c} />)}
        </motion.g>

        <motion.g style={{ x: farX, y: farY }}>
          <path d={HILLS.far} fill="var(--hill-far)" />
          <g transform={`translate(690 ${far(720) - 52})`}>
            <path d="M0 22 L30 0 L60 22 Z" fill="var(--barn-roof)" />
            <rect x="5" y="20" width="50" height="34" fill="var(--barn)" />
            <rect x="21" y="32" width="18" height="22" fill="var(--barn-roof)" />
            <path d="M21 32 L39 54 M39 32 L21 54" stroke="var(--barn-trim)" strokeWidth="2" />
            <rect x="24" y="12" width="12" height="8" fill="var(--barn-trim)" />
          </g>
        </motion.g>

        <motion.g style={{ x: midX }}>
          <path d={HILLS.mid} fill="var(--hill-mid)" />
          {TREES.map((x, i) => (
            <g key={x} transform={`translate(${x} ${mid(x) + 4}) scale(${i % 2 ? .8 : 1})`}>
              <rect x="-4" y="-34" width="8" height="36" rx="3" fill="var(--trunk)" />
              <circle cx="0" cy="-48" r="24" fill="var(--tree)" />
              <circle cx="-14" cy="-36" r="16" fill="var(--tree-dark)" />
              <circle cx="14" cy="-38" r="17" fill="var(--tree)" />
            </g>
          ))}
        </motion.g>

        <motion.g style={{ x: nearX, y: nearY }}>
          <path d={HILLS.near} fill="var(--hill-near)" />
          <path d={HILLS.near} fill="var(--hill-shade)" transform="translate(0 46)" />
          <g>
            <path d={FENCE.map((x, i) => `${i ? 'L' : 'M'}${x} ${near(x) - 44}`).join(' ')} fill="none" stroke="var(--wood-dark)" strokeWidth="7" />
            <path d={FENCE.map((x, i) => `${i ? 'L' : 'M'}${x} ${near(x) - 22}`).join(' ')} fill="none" stroke="var(--wood-dark)" strokeWidth="7" />
            <path d={FENCE.map((x, i) => `${i ? 'L' : 'M'}${x} ${near(x) - 45}`).join(' ')} fill="none" stroke="var(--wood)" strokeWidth="4" />
            <path d={FENCE.map((x, i) => `${i ? 'L' : 'M'}${x} ${near(x) - 23}`).join(' ')} fill="none" stroke="var(--wood)" strokeWidth="4" />
            {FENCE.map(x => <rect key={x} x={x - 5} y={near(x) - 60} width="10" height="70" rx="3" fill="var(--wood)" stroke="var(--wood-dark)" strokeWidth="2" />)}
          </g>
          <g stroke="var(--grass-blade-2)" strokeWidth="2" fill="none" strokeLinecap="round">
            {TUFTS.map((t, i) => <path key={i} className="tuft" d={t.d} style={{ animationDelay: t.delay }} />)}
          </g>
          {FLOWERS.map((f, i) => (
            <g key={i}>
              <circle cx={f.x} cy={f.y} r="3.6" fill={f.c} />
              <circle cx={f.x} cy={f.y} r="1.3" fill="var(--flower-2)" />
            </g>
          ))}
        </motion.g>

        <motion.g style={{ x: frontX }}>
          {BLADES.map((b, i) => <path key={i} className="blade" d={b.d} fill={b.c} style={{ animationDelay: b.delay, animationDuration: b.dur }} />)}
          {night && FIREFLIES.map((f, i) => <circle key={i} className="firefly" cx={f.x} cy={f.y} r="3" fill="var(--firefly)" style={{ animationDelay: f.delay }} />)}
        </motion.g>
      </svg>
      {showSleeper && <div className="scene-sleeper"><Cow pose="sleep" size={150} /></div>}
    </div>
  )
}
