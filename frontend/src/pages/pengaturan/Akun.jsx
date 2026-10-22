import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { Field } from '../../components/ui'

const ROLES = { super_admin: 'Super admin', owner: 'Owner', peternakan: 'Peternakan', melk: 'Melk by RBF', back_office: 'Back office' }

const KATALOG = [
  ['Peternakan', [['sapi', 'Rekap Sapi'], ['pakan', 'Stok Pakan'], ['produksi', 'Produksi Susu'], ['penjualan', 'Penjualan Susu'], ['investor', 'Investor'], ['kas', 'Keuangan Kandang']]],
  ['Melk by RBF', [['msusu', 'Susu'], ['mkeju', 'Keju'], ['mrekap', 'Rekap per 10 Hari'], ['mkas', 'Keuangan Melk']]],
  ['Back Office', [['bo', 'Rekap Keseluruhan'], ['bokas', 'Keuangan Back Office']]],
]

const ALL = KATALOG.flatMap(g => g[1].map(i => i[0]))

const NAMA = Object.fromEntries(KATALOG.flatMap(g => g[1]))

const DEF = { super_admin: ALL, owner: ALL, peternakan: ALL.slice(0, 6), melk: ALL.slice(6, 10), back_office: ALL.slice(10) }

function Akses({ value, onChange }) {
  return KATALOG.map(([g, items]) => (
    <div key={g} style={{ marginBottom: 10 }}>
      <div style={{ fontWeight: 700, fontSize: 13, marginBottom: 6 }}>{g}</div>
      <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
        {items.map(([k, l]) => (
          <label key={k} style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
            <input type="checkbox" style={{ minWidth: 0 }} checked={value.includes(k)}
              onChange={e => onChange(e.target.checked ? [...value, k] : value.filter(x => x !== k))} />{l}
          </label>
        ))}
      </div>
    </div>
  ))
}

export default function Akun() {
  const [list, setList] = useState([])
  const [form, setForm] = useState({ username: '', password: '', role: 'peternakan', menus: DEF.peternakan })
  const [edit, setEdit] = useState(null)
  const [err, setErr] = useState('')
  const load = useCallback(() => req('/users').then(setList).catch(e => setErr(e.message)), [])
  useEffect(() => { load() }, [load])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try { await req('/users', 'POST', form); setForm({ ...form, username: '', password: '' }); load() } catch (e) { setErr(e.message) }
  }
  const simpanAkses = async () => {
    try { await req(`/users/${edit.id}/menus`, 'PUT', { menus: edit.menus }); setEdit(null); load() } catch (e) { setErr(e.message) }
  }
  const hapus = async (id) => { if (confirm('Hapus akun ini?')) { try { await req(`/users/${id}`, 'DELETE'); load() } catch (e) { setErr(e.message) } } }
  return (<>
    <h1>Akun</h1><p className="sub">Buat akun, pilih role, lalu centang menu yang boleh diakses akun tersebut.</p>
    <div className="panel">
      <h3>Tambah akun</h3>
      <form onSubmit={simpan}>
        <div className="form" style={{ marginBottom: 14 }}>
          <Field label="Username"><input value={form.username} onChange={e => setForm({ ...form, username: e.target.value })} /></Field>
          <Field label="Password"><input type="password" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} /></Field>
          <Field label="Role"><select value={form.role} onChange={e => setForm({ ...form, role: e.target.value, menus: DEF[e.target.value] })}>{Object.entries(ROLES).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select></Field>
        </div>
        <h3 style={{ fontSize: 13 }}>Akses menu</h3>
        <Akses value={form.menus} onChange={m => setForm({ ...form, menus: m })} />
        <button>Tambah akun</button>
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    {edit && <div className="panel">
      <h3>Akses menu: {edit.username}</h3>
      <Akses value={edit.menus} onChange={m => setEdit({ ...edit, menus: m })} />
      <button onClick={simpanAkses}>Simpan akses</button>{' '}
      <button className="ghost" onClick={() => setEdit(null)}>Batal</button>
    </div>}
    <div className="panel tw">
      <table><thead><tr><th>Username</th><th>Role</th><th>Menu</th><th></th></tr></thead>
        <tbody>{list.map(u => <tr key={u.id}><td><b>{u.username}</b></td><td><span className="pill">{ROLES[u.role]}</span></td>
          <td style={{ maxWidth: 360 }}>{u.role === 'super_admin' ? 'Semua menu' : (u.menus.length ? u.menus.map(k => NAMA[k]).join(', ') : 'Tidak ada menu')}</td>
          <td style={{ textAlign: 'right' }}>
            {u.role !== 'super_admin' && <button className="ghost" onClick={() => setEdit({ id: u.id, username: u.username, menus: u.menus })}>Atur akses</button>}{' '}
            <button className="del" onClick={() => hapus(u.id)}>Hapus</button></td></tr>)}</tbody></table>
    </div>
  </>)
}
