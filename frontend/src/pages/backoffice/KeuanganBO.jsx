import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { today, rp, tglID } from '../../utils/format'
import { Field, Stat, Empty } from '../../components/ui'

const KAT = { kas: 'Kas', lainnya: 'Lainnya', gaji: 'Gaji pegawai', operasional: 'Operasional', topup_kandang: 'Top up ke kandang' }

export default function KeuanganBO() {
  const [bulan, setBulan] = useState(today().slice(0, 7))
  const [d, setD] = useState({ rows: [], masuk: 0, keluar: 0, saldo: 0 })
  const [ka, setKa] = useState({ tgl: today(), jumlah: 0 })
  const [kaForm, setKaForm] = useState({ tgl: today(), jumlah: '' })
  const [form, setForm] = useState({ tgl: today(), arah: 'masuk', kategori: 'kas', jumlah: '', keterangan: '' })
  const [err, setErr] = useState('')
  const OPSI = { masuk: ['kas', 'lainnya'], keluar: ['gaji', 'operasional', 'lainnya'] }
  const load = useCallback(() => {
    req('/bo/kas?bulan=' + bulan).then(setD).catch(e => setErr(e.message))
    req('/bo/kasawal').then(setKa).catch(() => {})
  }, [bulan])
  useEffect(() => { load() }, [load])
  const simpanKa = async (e) => {
    e.preventDefault(); setErr('')
    try { await req('/bo/kasawal', 'POST', { tgl: kaForm.tgl, jumlah: Number(kaForm.jumlah) }); setKaForm({ ...kaForm, jumlah: '' }); load() } catch (e) { setErr(e.message) }
  }
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try { await req('/bo/kas', 'POST', { ...form, jumlah: Number(form.jumlah) }); setForm({ ...form, jumlah: '', keterangan: '' }); load() } catch (e) { setErr(e.message) }
  }
  return (<>
    <h1>Keuangan Back Office</h1><p className="sub">Top up ke kandang tercatat otomatis dari menu Keuangan Kandang.</p>
    <div className="stats">
      <Stat hi label="Saldo kas back office" value={rp(d.saldo)} />
      <Stat label="Kas awal" value={rp(ka.jumlah)} />
      <Stat label="Uang masuk bulan ini" value={rp(d.masuk)} />
      <Stat label="Uang keluar bulan ini" value={rp(d.keluar)} />
    </div>
    <div className="panel">
      <h3>Kas awal back office</h3>
      <form className="form" onSubmit={simpanKa}>
        <Field label="Per tanggal"><input type="date" value={kaForm.tgl} onChange={e => setKaForm({ ...kaForm, tgl: e.target.value })} /></Field>
        <Field label="Jumlah (Rp)"><input type="number" min="0" value={kaForm.jumlah} onChange={e => setKaForm({ ...kaForm, jumlah: e.target.value })} /></Field>
        <button>Simpan kas awal</button>
      </form>
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
      <table><thead><tr><th>Tanggal</th><th>Kategori</th><th>Keterangan</th><th>Masuk</th><th>Keluar</th><th></th></tr></thead>
        <tbody>{d.rows.map(r => <tr key={(r.otomatis ? 'a' : 'm') + r.id}><td>{tglID(r.tgl)}</td>
          <td><span className={'pill ' + (r.arah === 'keluar' ? 'keluar' : '')}>{KAT[r.kategori]}</span>{r.otomatis && <span style={{ color: 'var(--muted)', fontSize: 12 }}> otomatis</span>}</td>
          <td>{r.keterangan}</td><td>{r.arah === 'masuk' ? rp(r.jumlah) : ''}</td><td>{r.arah === 'keluar' ? rp(r.jumlah) : ''}</td>
          <td style={{ textAlign: 'right' }}>{!r.otomatis && <button className="del" onClick={async () => { await req(`/bo/kas/${r.id}`, 'DELETE'); load() }}>Hapus</button>}</td></tr>)}</tbody></table>
      {!d.rows.length && <Empty text="Belum ada transaksi di bulan ini." />}
    </div>
  </>)
}
