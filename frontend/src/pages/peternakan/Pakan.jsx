import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { today, num, rp, tglID } from '../../utils/format'
import { Field, Stat, Empty } from '../../components/ui'

const JENIS = { rumput: 'Rumput', konsentrat: 'Konsentrat', tebon_jagung: 'Tebon jagung', mineral: 'Mineral' }

const kosongPakan = () => ({ tgl: today(), jenis: 'rumput', arah: 'masuk', jumlah: '', satuan: 'kg', harga_total: '', keterangan: '' })

export default function Pakan() {
  const [list, setList] = useState([])
  const [stok, setStok] = useState({})
  const [form, setForm] = useState(kosongPakan())
  const [err, setErr] = useState('')
  const load = useCallback(() => {
    req('/pakan').then(setList).catch(e => setErr(e.message))
    req('/pakan/stok').then(setStok).catch(() => {})
  }, [])
  useEffect(() => { load() }, [load])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try {
      await req('/pakan', 'POST', { ...form, jumlah: Number(form.jumlah), harga_total: Number(form.harga_total || 0) })
      setForm(kosongPakan()); load()
    } catch (e) { setErr(e.message) }
  }
  const hapus = async (id) => { if (confirm('Hapus catatan ini? Kas terkait ikut terhapus.')) { await req(`/pakan/${id}`, 'DELETE'); load() } }
  return (<>
    <h1>Stok Pakan</h1><p className="sub">Pakan masuk yang ada harganya otomatis tercatat sebagai kas keluar kandang.</p>
    <div className="stats">{Object.keys(JENIS).map(j => <Stat key={j} label={JENIS[j]} value={num(stok[j] || 0) + ' kg'} />)}</div>
    <div className="panel">
      <h3>Catat pakan</h3>
      <form className="form" onSubmit={simpan}>
        <Field label="Tanggal"><input type="date" value={form.tgl} onChange={e => setForm({ ...form, tgl: e.target.value })} /></Field>
        <Field label="Jenis"><select value={form.jenis} onChange={e => setForm({ ...form, jenis: e.target.value })}>{Object.entries(JENIS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select></Field>
        <Field label="Arah"><select value={form.arah} onChange={e => setForm({ ...form, arah: e.target.value })}><option value="masuk">Masuk (beli)</option><option value="keluar">Keluar (dipakai)</option></select></Field>
        <Field label="Jumlah (kg)"><input type="number" step="0.01" min="0" value={form.jumlah} onChange={e => setForm({ ...form, jumlah: e.target.value })} /></Field>
        {form.arah === 'masuk' && <Field label="Total harga (Rp)"><input type="number" min="0" value={form.harga_total} onChange={e => setForm({ ...form, harga_total: e.target.value })} /></Field>}
        <Field label="Keterangan"><input value={form.keterangan} onChange={e => setForm({ ...form, keterangan: e.target.value })} /></Field>
        <button>Simpan</button>
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    <div className="panel tw">
      <table>
        <thead><tr><th>Tanggal</th><th>Jenis</th><th>Arah</th><th>Jumlah</th><th>Harga</th><th>Keterangan</th><th></th></tr></thead>
        <tbody>{list.map(p => (
          <tr key={p.id}><td>{tglID(p.tgl)}</td><td>{JENIS[p.jenis]}</td>
            <td><span className={'pill ' + (p.arah === 'keluar' ? 'keluar' : '')}>{p.arah}</span></td>
            <td>{num(p.jumlah)} {p.satuan}</td><td>{p.harga_total ? rp(p.harga_total) : '-'}</td><td>{p.keterangan}</td>
            <td style={{ textAlign: 'right' }}><button className="del" onClick={() => hapus(p.id)}>Hapus</button></td></tr>))}</tbody>
      </table>
      {!list.length && <Empty text="Belum ada catatan pakan." />}
    </div>
  </>)
}
