import { useState } from 'react'
import { login } from '../../api/auth'
import { Field } from '../../components/ui'

export default function Login({ onLogin }) {
  const [f, setF] = useState({ username: '', password: '' })
  const [err, setErr] = useState('')
  const submit = async (e) => {
    e.preventDefault()
    try { onLogin(await login(f)) } catch (e) { setErr(e.message) }
  }
  return (
    <div className="login">
      <form onSubmit={submit}>
        <h2>Rejeki Baru Farm</h2>
        <p>Masuk untuk mengelola peternakan.</p>
        <Field label="Username"><input value={f.username} onChange={e => setF({ ...f, username: e.target.value })} /></Field>
        <Field label="Password"><input type="password" value={f.password} onChange={e => setF({ ...f, password: e.target.value })} /></Field>
        {err && <div className="err">{err}</div>}
        <button>Masuk</button>
      </form>
    </div>
  )
}
