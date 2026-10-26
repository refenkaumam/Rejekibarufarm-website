import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { today, num, rp, tglID } from '../../utils/format'
import { Field, Stat, Empty, Pager } from '../../components/ui'
import { usePaged } from '../../ui/usePaged'

const TUJUAN = { koperasi: 'Koperasi', melk_rbf: 'Melk by RBF' }

export default function Penjualan() {
  const [list, setList] = useState([])
  const [form, setForm] = useState({ tgl: today(), tujuan: 'koperasi', liter: '', harga_per_liter: '' })
  const [err, setErr] = useState('')
  const pg = usePaged(list)
  const load = useCallback(() => req('/penjualan').then(setList).catch(e => setErr(e.message)), [])
  useEffect(() => { load() }, [load])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try {
      await req('/penjualan', 'POST', { ...form, liter: Number(form.liter), harga_per_liter: Number(form.harga_per_liter) })
      setForm({ ...form, liter: '' }); load()
    } catch (e) { setErr(e.message) }
  }
  const hapus = async (id) => { if (confirm('Hapus penjualan ini? Kas terkait ikut terhapus.')) { await req(`/penjualan/${id}`, 'DELETE'); load() } }
  const bulan = today().slice(0, 7)
  const bi = list.filter(x => x.tgl.slice(0, 7) === bulan)
  const total = Number(form.liter || 0) * Number(form.harga_per_liter || 0)
  return (<>
    <h1>Penjualan Susu</h1><p className="sub">Hasil jual susu otomatis menjadi kas masuk kandang. Penjualan ke Melk by RBF juga tercatat sebagai susu masuk di Melk.</p>
    <div className="stats">
      <Stat hi label="Pendapatan bulan ini" value={rp(bi.reduce((a, x) => a + x.total, 0))} />
      <Stat label="Liter ke koperasi" value={num(bi.filter(x => x.tujuan === 'koperasi').reduce((a, x) => a + x.liter, 0)) + ' L'} />
      <Stat label="Liter ke Melk by RBF" value={num(bi.filter(x => x.tujuan === 'melk_rbf').reduce((a, x) => a + x.liter, 0)) + ' L'} />
    </div>
    <div className="panel">
      <h3>Catat penjualan</h3>
      <form className="form" onSubmit={simpan}>
        <Field label="Tanggal"><input type="date" value={form.tgl} onChange={e => setForm({ ...form, tgl: e.target.value })} /></Field>
        <Field label="Dijual ke"><select value={form.tujuan} onChange={e => setForm({ ...form, tujuan: e.target.value })}>{Object.entries(TUJUAN).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select></Field>
        <Field label="Liter"><input type="number" step="0.01" min="0" value={form.liter} onChange={e => setForm({ ...form, liter: e.target.value })} /></Field>
        <Field label="Harga per liter (Rp)"><input type="number" min="0" value={form.harga_per_liter} onChange={e => setForm({ ...form, harga_per_liter: e.target.value })} /></Field>
        <Field label="Total"><input readOnly value={rp(total)} /></Field>
        <button>Simpan</button>
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    <div className="panel tw">
      <table>
        <thead><tr><th>Tanggal</th><th>Tujuan</th><th>Liter</th><th>Harga/L</th><th>Total</th><th></th></tr></thead>
        <tbody>{pg.rows.map(p => <tr key={p.id}><td>{tglID(p.tgl)}</td><td>{TUJUAN[p.tujuan]}</td><td>{num(p.liter)} L</td><td>{rp(p.harga_per_liter)}</td><td><b>{rp(p.total)}</b></td>
          <td style={{ textAlign: 'right' }}><button className="del" onClick={() => hapus(p.id)}>Hapus</button></td></tr>)}</tbody>
      </table>
      <Pager {...pg} />
      {!list.length && <Empty text="Belum ada penjualan susu." />}
    </div>
  </>)
}
