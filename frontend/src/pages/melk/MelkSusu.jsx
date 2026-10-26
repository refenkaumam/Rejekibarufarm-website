import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { today, num, rp, tglID, bln } from '../../utils/format'
import { Field, Stat, Empty, Pager } from '../../components/ui'
import { usePaged } from '../../ui/usePaged'

export default function MelkSusu() {
  const [masuk, setMasuk] = useState([])
  const [keluar, setKeluar] = useState([])
  const [form, setForm] = useState({ tgl: today(), liter: '', harga_per_liter: '', pembeli: '' })
  const [err, setErr] = useState('')
  const pk = usePaged(keluar)
  const pm = usePaged(masuk)
  const load = useCallback(() => {
    req('/melk/susu?arah=masuk').then(setMasuk).catch(e => setErr(e.message))
    req('/melk/susu?arah=keluar').then(setKeluar).catch(() => {})
  }, [])
  useEffect(() => { load() }, [load])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try { await req('/melk/susu', 'POST', { ...form, liter: Number(form.liter), harga_per_liter: Number(form.harga_per_liter) }); setForm({ ...form, liter: '' }); load() }
    catch (e) { setErr(e.message) }
  }
  const hapus = async (id) => { if (confirm('Hapus data ini? Kas terkait ikut terhapus.')) { try { await req(`/melk/susu/${id}`, 'DELETE'); load() } catch (e) { setErr(e.message) } } }
  const bi = (l) => l.filter(x => x.tgl.slice(0, 7) === bln())
  const sum = (l, k) => l.reduce((a, x) => a + x[k], 0)
  return (<>
    <h1>Susu Melk by RBF</h1><p className="sub">Susu masuk tercatat otomatis dari penjualan kandang. Susu keluar yang dijual menjadi kas masuk Melk.</p>
    <div className="stats">
      <Stat hi label="Susu masuk bulan ini" value={num(sum(bi(masuk), 'liter')) + ' L'} />
      <Stat label="Biaya beli susu" value={rp(sum(bi(masuk), 'harga'))} />
      <Stat label="Susu keluar bulan ini" value={num(sum(bi(keluar), 'liter')) + ' L'} />
      <Stat label="Hasil jual susu" value={rp(sum(bi(keluar), 'harga'))} />
    </div>
    <div className="panel">
      <h3>Catat susu keluar</h3>
      <form className="form" onSubmit={simpan}>
        <Field label="Tanggal"><input type="date" value={form.tgl} onChange={e => setForm({ ...form, tgl: e.target.value })} /></Field>
        <Field label="Dijual ke"><input value={form.pembeli} onChange={e => setForm({ ...form, pembeli: e.target.value })} /></Field>
        <Field label="Liter"><input type="number" step="0.01" min="0" value={form.liter} onChange={e => setForm({ ...form, liter: e.target.value })} /></Field>
        <Field label="Harga per liter (Rp)"><input type="number" min="0" value={form.harga_per_liter} onChange={e => setForm({ ...form, harga_per_liter: e.target.value })} /></Field>
        <Field label="Total"><input readOnly value={rp(Number(form.liter || 0) * Number(form.harga_per_liter || 0))} /></Field>
        <button>Simpan</button>
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    <div className="panel tw"><h3>Susu keluar</h3>
      <table><thead><tr><th>Tanggal</th><th>Dijual ke</th><th>Liter</th><th>Total</th><th></th></tr></thead>
        <tbody>{pk.rows.map(r => <tr key={r.id}><td>{tglID(r.tgl)}</td><td>{r.pembeli}</td><td>{num(r.liter)} L</td><td><b>{rp(r.harga)}</b></td>
          <td style={{ textAlign: 'right' }}><button className="del" onClick={() => hapus(r.id)}>Hapus</button></td></tr>)}</tbody></table>
      <Pager {...pk} />
      {!keluar.length && <Empty text="Belum ada susu keluar." />}
    </div>
    <div className="panel tw"><h3>Susu masuk dari kandang</h3>
      <table><thead><tr><th>Tanggal</th><th>Liter</th><th>Total harga</th></tr></thead>
        <tbody>{pm.rows.map(r => <tr key={r.id}><td>{tglID(r.tgl)}</td><td>{num(r.liter)} L</td><td>{rp(r.harga)}</td></tr>)}</tbody></table>
      <Pager {...pm} />
      {!masuk.length && <Empty text="Belum ada susu masuk. Data muncul saat kandang menjual susu ke Melk by RBF." />}
    </div>
  </>)
}
