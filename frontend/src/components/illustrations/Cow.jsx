import { useId } from 'react'
import { motion, AnimatePresence } from 'framer-motion'

// Sapi orisinal Rejeki Baru Farm (SVG inline).
// pose: idle | eyes-closed | shake | jump | sleep | confused
// look: { x, y } antara -1 dan 1, arah pandangan mata
// say:  teks balon (mis. "Mooo!"), null untuk menyembunyikan

const ROOT = {
  idle: { x: 0, y: 0, scaleY: 1 },
  'eyes-closed': { x: 0, y: 0, scaleY: 1 },
  sleep: { x: 0, y: 0, scaleY: 1 },
  confused: { x: 0, y: 0, scaleY: 1 },
  shake: { x: [0, -10, 10, -8, 8, -4, 4, 0], y: 0, scaleY: 1, transition: { duration: .6, ease: 'easeInOut' } },
  jump: { x: 0, y: [0, -38, 0, -12, 0], scaleY: [1, 1.05, .92, 1.02, 1], transition: { duration: .8, ease: 'easeOut' } },
}

const clamp = (v) => Math.max(-1, Math.min(1, v || 0))

function Eye({ cx, cy, clip, open, look }) {
  if (!open) return null
  return (
    <g>
      <clipPath id={clip}><circle cx={cx} cy={cy} r="11.5" /></clipPath>
      <circle cx={cx} cy={cy} r="11.5" fill="var(--cow-eye)" stroke="var(--cow-outline)" strokeWidth="2.6" />
      <g clipPath={`url(#${clip})`}>
        <g className="cow-pupil" style={{ transform: `translate(${look.x * 4.5}px, ${look.y * 4}px)` }}>
          <circle cx={cx} cy={cy + 1} r="6.6" fill="var(--cow-pupil)" />
          <circle cx={cx + 2.3} cy={cy - 1.6} r="2.2" fill="var(--cow-eye)" />
        </g>
        <rect className="cow-lid" x={cx - 13} y={cy - 13} width="26" height="26" fill="var(--cow-white)" />
      </g>
    </g>
  )
}

function Mouth({ pose }) {
  const line = { fill: 'none', stroke: 'var(--cow-outline)', strokeWidth: 2.6, strokeLinecap: 'round' }
  if (pose === 'jump') return <path d="M106 131 Q120 148 134 131 Z" fill="var(--cow-mouth)" stroke="var(--cow-outline)" strokeWidth="2.4" strokeLinejoin="round" />
  if (pose === 'shake') return <path d="M108 138 Q120 131 132 138" {...line} />
  if (pose === 'confused') return <path d="M107 136 q4.3 -4 8.6 0 t8.6 0 t8.6 0" {...line} />
  if (pose === 'sleep') return <path d="M114 134 Q120 138 126 134" {...line} />
  return <path d="M108 133 Q120 141 132 133" {...line} />
}

export default function Cow({ pose = 'idle', look, say = null, size = 220, onClick, label, className = '' }) {
  const uid = useId().replace(/[^a-zA-Z0-9_-]/g, '')
  const l = pose === 'confused' ? { x: -.4, y: -1 } : { x: clamp(look?.x), y: clamp(look?.y) }
  const eyesOpen = pose === 'idle' || pose === 'shake' || pose === 'confused'
  const sleeping = pose === 'sleep'
  const chewing = pose === 'idle' || pose === 'eyes-closed'
  const closedLine = { fill: 'none', stroke: 'var(--cow-outline)', strokeWidth: 3.2, strokeLinecap: 'round' }

  const Root = onClick ? motion.button : motion.div
  const a11y = onClick
    ? { type: 'button', onClick, 'aria-label': label || 'Sapi' }
    : label ? { role: 'img', 'aria-label': label } : { 'aria-hidden': true }

  return (
    <Root className={'cow ' + className} data-pose={pose} style={{ width: size, originY: 1 }}
      variants={ROOT} animate={pose} initial={false} {...a11y}>
      <svg viewBox="0 0 240 240" width="100%" height="100%" overflow="visible">
        <defs>
          <clipPath id={uid + 'body'}><ellipse cx="120" cy="180" rx="60" ry="38" /></clipPath>
          <clipPath id={uid + 'head'}><ellipse cx="120" cy="92" rx="54" ry="47" /></clipPath>
        </defs>

        <ellipse cx="120" cy="229" rx={sleeping ? 78 : 64} ry="7" fill="var(--cow-shadow)" />

        <g transform={sleeping ? 'translate(0 14)' : undefined}>
          {/* ekor */}
          <g className="cow-tail">
            <path d="M176 176 C 196 172, 206 152, 200 132" fill="none" stroke="var(--cow-outline)" strokeWidth="4.5" strokeLinecap="round" />
            <path d="M200 133 c -7 -2 -9 -12 -1 -16 c 7 2 9 12 1 16 z" fill="var(--cow-spot)" />
          </g>

          {/* kaki */}
          {!sleeping && <g>
            {[[68, 192, 30], [154, 192, 30], [92, 196, 31], [130, 196, 31]].map(([x, y, h]) => (
              <g key={x}>
                <rect x={x} y={y} width="18" height={h} rx="8" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="2.6" />
                <path d={`M${x + 1.3} ${y + h - 9} h15.4 v4 a5 5 0 0 1 -5 5 h-5.4 a5 5 0 0 1 -5 -5 z`} fill="var(--cow-hoof)" />
              </g>
            ))}
          </g>}

          {/* badan */}
          <g className="cow-body">
            <ellipse cx="120" cy="180" rx="60" ry="38" fill="var(--cow-white)" />
            <g clipPath={`url(#${uid}body)`} fill="var(--cow-spot)">
              <path d="M64 168 c 8 -16 34 -16 38 0 c 4 16 -14 26 -28 22 c -12 -3 -16 -12 -10 -22 z" />
              <path d="M146 190 c 6 -14 30 -14 34 0 c 2 12 -12 22 -24 18 c -8 -3 -13 -10 -10 -18 z" />
              <path d="M118 202 c 6 -6 16 -4 15 4 c -1 7 -12 9 -15 3 z" />
            </g>
            <ellipse cx="120" cy="180" rx="60" ry="38" fill="none" stroke="var(--cow-outline)" strokeWidth="3" />
          </g>

          {/* kalung & lonceng */}
          <path d="M86 138 Q120 158 154 138" fill="none" stroke="var(--cow-collar)" strokeWidth="7" strokeLinecap="round" />
          <g className="cow-bell">
            <path d="M111 164 C 111 154, 114 150, 120 150 C 126 150, 129 154, 129 164 L 131 167 H 109 Z" fill="var(--cow-bell)" stroke="var(--cow-bell-dark)" strokeWidth="2" strokeLinejoin="round" />
            <circle cx="120" cy="168.5" r="3" fill="var(--cow-bell-dark)" />
          </g>

          {/* kepala */}
          <g transform={sleeping ? 'translate(0 10) rotate(-6 120 140)' : undefined}>
            <g className="cow-head" style={{ transform: sleeping ? undefined : `translate(${l.x * 3}px, ${l.y * 2}px) rotate(${l.x * 5}deg)` }}>
              <g className="cow-ear-l">
                <ellipse cx="54" cy="80" rx="22" ry="11" transform="rotate(-18 54 80)" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="2.8" />
                <ellipse cx="57" cy="81" rx="13" ry="5.5" transform="rotate(-18 57 81)" fill="var(--cow-muzzle)" />
              </g>
              <g className="cow-ear-r">
                <ellipse cx="186" cy="80" rx="22" ry="11" transform="rotate(18 186 80)" fill="var(--cow-white)" stroke="var(--cow-outline)" strokeWidth="2.8" />
                <ellipse cx="183" cy="81" rx="13" ry="5.5" transform="rotate(18 183 81)" fill="var(--cow-muzzle)" />
              </g>
              <path d="M100 52 C 94 42, 94 32, 102 26 C 102 34, 105 42, 110 48 Z" fill="var(--cow-horn)" stroke="var(--cow-outline)" strokeWidth="2.5" strokeLinejoin="round" />
              <path d="M140 52 C 146 42, 146 32, 138 26 C 138 34, 135 42, 130 48 Z" fill="var(--cow-horn)" stroke="var(--cow-outline)" strokeWidth="2.5" strokeLinejoin="round" />

              <ellipse cx="120" cy="92" rx="54" ry="47" fill="var(--cow-white)" />
              <g clipPath={`url(#${uid}head)`} fill="var(--cow-spot)">
                <path d="M136 52 C 152 44, 172 54, 172 68 C 172 78, 160 80, 150 74 C 140 69, 130 60, 136 52 Z" />
                <path d="M72 70 C 76 60, 90 60, 90 70 C 90 78, 76 82, 72 70 Z" />
              </g>
              <ellipse cx="120" cy="92" rx="54" ry="47" fill="none" stroke="var(--cow-outline)" strokeWidth="3" />
              <path d="M106 52 C 106 42, 116 38, 120 46 C 124 38, 134 42, 134 52 Z" fill="var(--cow-spot)" />

              {/* alis */}
              {pose === 'shake' && <g {...closedLine} strokeWidth="3">
                <path d="M88 74 L108 69" /><path d="M132 69 L152 74" />
              </g>}
              {pose === 'confused' && <g {...closedLine} strokeWidth="3">
                <path d="M88 70 Q98 61 108 67" /><path d="M132 73 L152 73" />
              </g>}

              {/* mata */}
              <Eye cx={99} cy={88} clip={uid + 'el'} open={eyesOpen} look={l} />
              <Eye cx={141} cy={88} clip={uid + 'er'} open={eyesOpen} look={l} />
              {(pose === 'eyes-closed' || sleeping) && <g {...closedLine}>
                <path d="M88 89 Q99 97 110 89" /><path d="M130 89 Q141 97 152 89" />
              </g>}
              {pose === 'jump' && <g {...closedLine}>
                <path d="M88 92 Q99 79 110 92" /><path d="M130 92 Q141 79 152 92" />
              </g>}

              <circle cx="80" cy="109" r={pose === 'eyes-closed' ? 9.5 : 8} fill="var(--cow-blush)" opacity={pose === 'eyes-closed' ? .85 : .5} />
              <circle cx="160" cy="109" r={pose === 'eyes-closed' ? 9.5 : 8} fill="var(--cow-blush)" opacity={pose === 'eyes-closed' ? .85 : .5} />

              {/* moncong */}
              <ellipse cx="120" cy="124" rx="35" ry="21" fill="var(--cow-muzzle)" stroke="var(--cow-outline)" strokeWidth="2.5" />
              <ellipse cx="107" cy="118" rx="4.2" ry="5.6" fill="var(--cow-muzzle-dark)" />
              <ellipse cx="133" cy="118" rx="4.2" ry="5.6" fill="var(--cow-muzzle-dark)" />
              <g className={chewing ? 'cow-jaw chewing' : 'cow-jaw'}>
                <Mouth pose={pose} />
                {pose === 'idle' && <g fill="none" strokeLinecap="round" strokeWidth="3.2">
                  <path d="M131 134 Q143 136 155 126" stroke="var(--cow-grass)" />
                  <path d="M131 135 Q145 141 154 137" stroke="var(--cow-grass-2)" />
                  <path d="M132 134 Q140 128 146 118" stroke="var(--cow-grass-2)" />
                </g>}
              </g>
            </g>
          </g>
        </g>

        {sleeping && <g className="cow-zz" fill="var(--color-text-muted)" fontFamily="var(--font-heading)" fontWeight="600">
          <text x="166" y="62" fontSize="20">z</text>
          <text x="182" y="42" fontSize="15">z</text>
          <text x="195" y="26" fontSize="11">z</text>
        </g>}
        {pose === 'confused' && <text className="cow-q" x="176" y="44" fontSize="40" fontFamily="var(--font-heading)" fontWeight="700"
          fill="var(--color-accent)" stroke="var(--cow-outline)" strokeWidth="1.6">?</text>}
        {pose === 'jump' && <g className="cow-sparkle" fill="var(--color-accent)">
          <path d="M44 46 l3 -9 l3 9 l9 3 l-9 3 l-3 9 l-3 -9 l-9 -3 z" />
          <path d="M192 30 l2 -6 l2 6 l6 2 l-6 2 l-2 6 l-2 -6 l-6 -2 z" />
        </g>}
      </svg>

      <AnimatePresence>
        {say && (
          <motion.span key={say} className="cow-bubble" role="status"
            initial={{ opacity: 0, scale: .6, y: 8 }} animate={{ opacity: 1, scale: 1, y: 0 }} exit={{ opacity: 0, scale: .8, y: -6 }}
            transition={{ type: 'spring', stiffness: 420, damping: 22 }}>
            {say}
          </motion.span>
        )}
      </AnimatePresence>
    </Root>
  )
}
