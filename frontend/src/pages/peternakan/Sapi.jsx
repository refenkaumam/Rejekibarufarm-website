import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { bln } from '../../utils/format'
import { Field, Stat, Empty } from '../../components/ui'

const STATUS = ['laktasi', 'kering', 'dara', 'pedet']

const kosongSapi = { nomor: '', nama: '', tgl_lahir: '', jumlah_anak: 0, status: 'laktasi' }

export default function Sapi() {
  const [list, setList] = useState([])
  const [form, setForm] = useState(kosongSapi)
  const [editId, setEditId] = useState(null)
  const [err, setErr] = useState('')
  const load = useCallback(() => req('/sapi').then(setList).catch(e => setErr(e.message)), [])
  useEffect(() => { load() }, [load])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try {
      await req(editId ? `/sapi/${editId}` : '/sapi', editId ? 'PUT' : 'POST', { ...form, jumlah_anak: Number(form.jumlah_anak) })
      setForm(kosongSapi); setEditId(null); load()
    } catch (e) { setErr(e.message) }
  }
  const hapus = async (id) => { if (confirm('Hapus sapi ini?')) { await req(`/sapi/${id}`, 'DELETE'); load() } }
  const umur = (b) => `${Math.floor(b / 12)} th ${b % 12} bln`
  return (<>
    <h1>Rekap Sapi</h1><p className="sub">Data populasi sapi. Umur dihitung otomatis dari tanggal lahir.</p>
    <div className="stats">
      <Stat hi label="Total populasi" value={list.length} />
      {STATUS.map(s => <Stat key={s} label={s[0].toUpperCase() + s.slice(1)} value={list.filter(x => x.status === s).length} />)}
    </div>
    <div className="panel">
      <h3>{editId ? 'Edit sapi' : 'Tambah sapi'}</h3>
      <form className="form" onSubmit={simpan}>
        <Field label="Nomor sapi"><input value={form.nomor} onChange={e => setForm({ ...form, nomor: e.target.value })} /></Field>
        <Field label="Nama sapi"><input value={form.nama} onChange={e => setForm({ ...form, nama: e.target.value })} /></Field>
        <Field label="Tanggal lahir"><input type="date" value={form.tgl_lahir} onChange={e => setForm({ ...form, tgl_lahir: e.target.value })} /></Field>
        <Field label="Jumlah anak"><input type="number" min="0" value={form.jumlah_anak} onChange={e => setForm({ ...form, jumlah_anak: e.target.value })} /></Field>
        <Field label="Status"><select value={form.status} onChange={e => setForm({ ...form, status: e.target.value })}>{STATUS.map(s => <option key={s}>{s}</option>)}</select></Field>
        <button>{editId ? 'Simpan perubahan' : 'Tambah sapi'}</button>
        {editId && <button type="button" className="ghost" onClick={() => { setForm(kosongSapi); setEditId(null) }}>Batal</button>}
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    <div className="panel tw">
      <table>
        <thead><tr><th>Nomor</th><th>Nama</th><th>Umur</th><th>Anak</th><th>Status</th><th></th></tr></thead>
        <tbody>{list.map(s => (
          <tr key={s.id}><td>{s.nomor}</td><td><b>{s.nama}</b></td><td>{umur(s.umur_bulan)}</td><td>{s.jumlah_anak}</td>
            <td><span className={'pill ' + s.status}>{s.status}</span></td>
            <td style={{ textAlign: 'right' }}>
              <button className="ghost" onClick={() => { setEditId(s.id); setForm({ ...s, tgl_lahir: s.tgl_lahir.slice(0, 10) }) }}>Edit</button>{' '}
              <button className="del" onClick={() => hapus(s.id)}>Hapus</button></td></tr>))}</tbody>
      </table>
      {!list.length && <Empty text="Belum ada sapi. Tambahkan sapi pertama lewat form di atas." />}
    </div>
  </>)
}
