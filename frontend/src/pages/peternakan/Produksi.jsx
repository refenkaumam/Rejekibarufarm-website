import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { today, num, tglID } from '../../utils/format'
import { Field, Stat, Empty, Pager } from '../../components/ui'
import { usePaged } from '../../ui/usePaged'

export default function Produksi() {
  const [list, setList] = useState([])
  const [rekap, setRekap] = useState([])
  const [by, setBy] = useState('bulan')
  const [form, setForm] = useState({ tgl: today(), sesi: 'pagi', liter: '' })
  const [err, setErr] = useState('')
  const pg = usePaged(list)
  const pr = usePaged(rekap, 12)
  const load = useCallback(() => req('/produksi').then(setList).catch(e => setErr(e.message)), [])
  useEffect(() => { load() }, [load])
  useEffect(() => { req('/produksi/rekap?by=' + by).then(setRekap).catch(() => {}) }, [by, list])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try {
      await req('/produksi', 'POST', { ...form, liter: Number(form.liter) })
      setForm({ ...form, liter: '', sesi: form.sesi === 'pagi' ? 'sore' : 'pagi' }); load()
    } catch (e) { setErr(e.message) }
  }
  const hariIni = list.find(x => x.tgl.slice(0, 10) === today()) || { pagi: 0, sore: 0, total: 0 }
  const bulanIni = rekap[0] && by === 'bulan' ? rekap[0] : null
  return (<>
    <h1>Produksi Susu</h1><p className="sub">Catat hasil perahan pagi dan sore. Mengisi sesi yang sama di tanggal yang sama akan menimpa angka lama.</p>
    <div className="stats">
      <Stat hi label="Total hari ini" value={num(hariIni.total) + ' L'} />
      <Stat label="Pagi" value={num(hariIni.pagi) + ' L'} />
      <Stat label="Sore" value={num(hariIni.sore) + ' L'} />
      {bulanIni && <Stat label="Bulan ini" value={num(bulanIni.total) + ' L'} />}
    </div>
    <div className="panel">
      <h3>Catat produksi</h3>
      <form className="form" onSubmit={simpan}>
        <Field label="Tanggal"><input type="date" value={form.tgl} onChange={e => setForm({ ...form, tgl: e.target.value })} /></Field>
        <Field label="Sesi"><select value={form.sesi} onChange={e => setForm({ ...form, sesi: e.target.value })}><option>pagi</option><option>sore</option></select></Field>
        <Field label="Liter"><input type="number" step="0.01" min="0" value={form.liter} onChange={e => setForm({ ...form, liter: e.target.value })} /></Field>
        <button>Simpan</button>
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    <div className="panel tw">
      <h3>Harian (31 hari terakhir)</h3>
      <table>
        <thead><tr><th>Tanggal</th><th>Pagi</th><th>Sore</th><th>Total</th></tr></thead>
        <tbody>{pg.rows.map(r => <tr key={r.tgl}><td>{tglID(r.tgl)}</td><td>{num(r.pagi)} L</td><td>{num(r.sore)} L</td><td><b>{num(r.total)} L</b></td></tr>)}</tbody>
      </table>
      <Pager {...pg} />
      {!list.length && <Empty text="Belum ada data produksi." />}
    </div>
    <div className="panel tw">
      <div className="tabs">
        <button className={by === 'bulan' ? '' : 'ghost'} onClick={() => setBy('bulan')}>Per bulan</button>
        <button className={by === 'tahun' ? '' : 'ghost'} onClick={() => setBy('tahun')}>Per tahun</button>
      </div>
      <table>
        <thead><tr><th>Periode</th><th>Total</th><th>Hari tercatat</th><th>Rata-rata per hari</th></tr></thead>
        <tbody>{pr.rows.map(r => <tr key={r.periode}><td>{r.periode}</td><td><b>{num(r.total)} L</b></td><td>{r.hari}</td><td>{num(r.total / r.hari)} L</td></tr>)}</tbody>
      </table>
      <Pager {...pr} />
    </div>
  </>)
}
