import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { today, rp, tglID } from '../../utils/format'
import { Field, Stat, Empty } from '../../components/ui'

const KAT = { kas: 'Kas', pembelian_susu: 'Penjualan susu', topup_bo: 'Top up back office', investor: 'Modal investor', lainnya: 'Lainnya',
  gaji: 'Gaji pegawai', beli_pakan: 'Beli pakan', operasional: 'Operasional', bulanan_investor: 'Bulanan investor' }

const OPSI = { masuk: ['kas', 'topup_bo', 'lainnya'], keluar: ['gaji', 'operasional', 'lainnya'] }

export default function KeuanganKandang() {
  const [bulan, setBulan] = useState(today().slice(0, 7))
  const [d, setD] = useState({ rows: [], masuk: 0, keluar: 0, saldo: 0 })
  const [form, setForm] = useState({ tgl: today(), arah: 'masuk', kategori: 'kas', jumlah: '', keterangan: '' })
  const [err, setErr] = useState('')
  const load = useCallback(() => req('/kas?bulan=' + bulan).then(setD).catch(e => setErr(e.message)), [bulan])
  useEffect(() => { load() }, [load])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try { await req('/kas', 'POST', { ...form, jumlah: Number(form.jumlah) }); setForm({ ...form, jumlah: '', keterangan: '' }); load() }
    catch (e) { setErr(e.message) }
  }
  const hapus = async (id) => { try { await req(`/kas/${id}`, 'DELETE'); load() } catch (e) { setErr(e.message) } }
  return (<>
    <h1>Keuangan Kandang</h1><p className="sub">Pakan, penjualan susu, dan investor tercatat otomatis. Di sini kamu hanya mengisi kas, top up, gaji, dan operasional.</p>
    <div className="stats">
      <Stat hi label="Saldo kas kandang" value={rp(d.saldo)} />
      <Stat label="Uang masuk bulan ini" value={rp(d.masuk)} />
      <Stat label="Uang keluar bulan ini" value={rp(d.keluar)} />
      <Stat label="Selisih bulan ini" value={rp(d.masuk - d.keluar)} />
    </div>
    <div className="panel">
      <h3>Catat transaksi</h3>
      <form className="form" onSubmit={simpan}>
        <Field label="Tanggal"><input type="date" value={form.tgl} onChange={e => setForm({ ...form, tgl: e.target.value })} /></Field>
        <Field label="Arah"><select value={form.arah} onChange={e => setForm({ ...form, arah: e.target.value, kategori: OPSI[e.target.value][0] })}><option value="masuk">Uang masuk</option><option value="keluar">Uang keluar</option></select></Field>
        <Field label="Kategori"><select value={form.kategori} onChange={e => setForm({ ...form, kategori: e.target.value })}>{OPSI[form.arah].map(k => <option key={k} value={k}>{KAT[k]}</option>)}</select></Field>
        <Field label="Jumlah (Rp)"><input type="number" min="0" value={form.jumlah} onChange={e => setForm({ ...form, jumlah: e.target.value })} /></Field>
        <Field label="Keterangan"><input value={form.keterangan} onChange={e => setForm({ ...form, keterangan: e.target.value })} /></Field>
        <button>Simpan</button>
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    <div className="panel tw">
      <div className="tabs"><input type="month" value={bulan} onChange={e => setBulan(e.target.value)} /></div>
      <table>
        <thead><tr><th>Tanggal</th><th>Kategori</th><th>Keterangan</th><th>Masuk</th><th>Keluar</th><th></th></tr></thead>
        <tbody>{d.rows.map(r => <tr key={r.id}><td>{tglID(r.tgl)}</td>
          <td><span className={'pill ' + (r.arah === 'keluar' ? 'keluar' : '')}>{KAT[r.kategori] || r.kategori}</span>{r.otomatis && <span style={{ color: 'var(--muted)', fontSize: 12 }}> otomatis</span>}</td>
          <td>{r.keterangan}</td><td>{r.arah === 'masuk' ? rp(r.jumlah) : ''}</td><td>{r.arah === 'keluar' ? rp(r.jumlah) : ''}</td>
          <td style={{ textAlign: 'right' }}>{!r.otomatis && <button className="del" onClick={() => hapus(r.id)}>Hapus</button>}</td></tr>)}</tbody>
      </table>
      {!d.rows.length && <Empty text="Belum ada transaksi di bulan ini." />}
    </div>
  </>)
}
