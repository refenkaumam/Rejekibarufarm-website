import { useEffect, useRef, useState } from 'react'
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'
import { login } from '../../api/auth'
import { Field } from '../../components/ui'
import Cow from '../../components/illustrations/Cow'
import CowMark from '../../components/illustrations/CowMark'
import Pasture from '../../components/illustrations/Pasture'
import ThemeToggle from '../../components/ThemeToggle'
import { IconUser, IconLock, IconAlert, IconArrowRight } from '../../components/icons'
import { useTheme } from '../../ui/theme'
import { phaseOf, sapaan } from '../../ui/time'

const clamp = (v) => Math.max(-1, Math.min(1, v))

export default function Login({ onLogin }) {
  const [f, setF] = useState({ username: '', password: '' })
  const [err, setErr] = useState('')
  // state di bawah ini hanya untuk tampilan (animasi sapi & tombol)
  const [pose, setPose] = useState('idle')
  const [busy, setBusy] = useState(false)
  const [pwFocus, setPwFocus] = useState(false)
  const [look, setLook] = useState({ x: 0, y: 0 })
  const [moo, setMoo] = useState(0)
  const cowRef = useRef(null)
  const reduce = useReducedMotion()
  const theme = useTheme()
  const phase = theme === 'dark' ? 'malam' : phaseOf()

  // Mata sapi mengikuti kursor (mouse) atau posisi tap terakhir (layar sentuh)
  useEffect(() => {
    let raf = 0
    const follow = (e) => {
      if (e.type === 'pointermove' && e.pointerType !== 'mouse') return
      cancelAnimationFrame(raf)
      raf = requestAnimationFrame(() => {
        const el = cowRef.current
        if (!el) return
        const r = el.getBoundingClientRect()
        const cx = r.left + r.width / 2
        const cy = r.top + r.height * .38
        setLook({ x: clamp((e.clientX - cx) / (window.innerWidth / 2.5)), y: clamp((e.clientY - cy) / (window.innerHeight / 2.5)) })
      })
    }
    window.addEventListener('pointermove', follow)
    window.addEventListener('pointerdown', follow)
    return () => {
      cancelAnimationFrame(raf)
      window.removeEventListener('pointermove', follow)
      window.removeEventListener('pointerdown', follow)
    }
  }, [])

  useEffect(() => {
    if (!moo) return
    const t = setTimeout(() => setMoo(0), 1500)
    return () => clearTimeout(t)
  }, [moo])

  const submit = async (e) => {
    e.preventDefault()
    setBusy(true)
    try {
      const u = await login(f)
      setErr(''); setPose('jump')
      setTimeout(() => onLogin(u), reduce ? 0 : 800)
    } catch (e) {
      setErr(e.message); setPose('shake'); setBusy(false)
      setTimeout(() => setPose('idle'), 700)
    }
  }

  const cowPose = pose !== 'idle' ? pose : pwFocus ? 'eyes-closed' : 'idle'

  return (
    <div className="login-page">
      <Pasture phase={phase} />
      <ThemeToggle className="login-theme" />

      <div className="login-stage">
        <motion.div className="login-hello" initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: .3 }}>
          {sapaan()}! Klik aku untuk menyapa.
        </motion.div>
        <div ref={cowRef} className="login-cow">
          <Cow pose={cowPose} look={look} say={moo ? 'Mooo!' : null} size="100%"
            onClick={() => setMoo(n => n + 1)} label="Sapi Rejeki Baru Farm, klik untuk menyapa" />
        </div>
      </div>

      <motion.form className="login-card" onSubmit={submit}
        initial={{ opacity: 0, y: 24, scale: .98 }} animate={{ opacity: 1, y: 0, scale: 1 }} transition={{ duration: .5, ease: [.22, 1, .36, 1] }}>
        <svg className="login-spots" viewBox="0 0 400 400" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
          <path d="M306 -6c26-18 66-12 82 16 12 20-2 34 2 54 4 22-18 40-42 30-18-8-26-2-40-14-16-14-6-34-14-50-8-18 0-26 12-36z" />
          <path d="M-4 304c18-14 46-8 54 10 6 14 22 14 26 32 4 20-18 34-38 30-14-3-22 6-38-2-12-8-16-24-14-40 2-12 4-24 10-30z" />
          <path d="M330 338c10-8 26-4 30 6 4 10 14 12 10 24-4 12-20 12-30 6-10-6-18-26-10-36z" />
        </svg>
        <div className="login-brand">
          <CowMark size={52} />
          <div>
            <h2>Rejeki Baru Farm</h2>
            <p>Masuk untuk mengelola peternakan.</p>
          </div>
        </div>
        <Field label="Username">
          <span className="input-ico"><IconUser />
            <input autoComplete="username" value={f.username} onChange={e => setF({ ...f, username: e.target.value })} />
          </span>
        </Field>
        <Field label="Password">
          <span className="input-ico"><IconLock />
            <input type="password" autoComplete="current-password" value={f.password}
              onFocus={() => setPwFocus(true)} onBlur={() => setPwFocus(false)}
              onChange={e => setF({ ...f, password: e.target.value })} />
          </span>
        </Field>
        <AnimatePresence>
          {err && (
            <motion.div key={err} className="err" role="alert"
              initial={{ opacity: 0, height: 0 }} animate={{ opacity: 1, height: 'auto', x: [0, -6, 6, -3, 3, 0] }} exit={{ opacity: 0, height: 0 }}>
              <IconAlert />{err}
            </motion.div>
          )}
        </AnimatePresence>
        <motion.button className="login-btn" disabled={busy} whileHover={{ y: -2 }} whileTap={{ scale: .97 }}>
          {busy ? <span className="btn-spinner" aria-hidden="true" /> : null}
          {busy ? 'Memeriksa...' : 'Masuk'}
          {!busy && <IconArrowRight />}
        </motion.button>
      </motion.form>
    </div>
  )
}
