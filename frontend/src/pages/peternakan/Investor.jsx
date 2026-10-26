import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { today, rp, tglID } from '../../utils/format'
import { Field, Stat, Empty, Pager } from '../../components/ui'
import { usePaged } from '../../ui/usePaged'

const kosongInv = { nama: '', kontak: '', jumlah_invest: '', mulai: today(), selesai: '', nomor_sapi: '' }

export default function Investor() {
  const [list, setList] = useState([])
  const [bayar, setBayar] = useState([])
  const [form, setForm] = useState(kosongInv)
  const [editId, setEditId] = useState(null)
  const [fb, setFb] = useState({ investor_id: '', tgl: today(), jumlah: '', keterangan: '' })
  const [err, setErr] = useState('')
  const pi = usePaged(list)
  const pb = usePaged(bayar)
  const load = useCallback(() => {
    req('/investor').then(setList).catch(e => setErr(e.message))
    req('/investor-bayar').then(setBayar).catch(() => {})
  }, [])
  useEffect(() => { load() }, [load])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try {
      await req(editId ? `/investor/${editId}` : '/investor', editId ? 'PUT' : 'POST', { ...form, jumlah_invest: Number(form.jumlah_invest) })
      setForm(kosongInv); setEditId(null); load()
    } catch (e) { setErr(e.message) }
  }
  const hapus = async (id) => { if (confirm('Hapus investor beserta catatan kas-nya?')) { await req(`/investor/${id}`, 'DELETE'); load() } }
  const simpanBayar = async (e) => {
    e.preventDefault(); setErr('')
    try {
      await req('/investor-bayar', 'POST', { ...fb, investor_id: Number(fb.investor_id), jumlah: Number(fb.jumlah) })
      setFb({ ...fb, jumlah: '', keterangan: '' }); load()
    } catch (e) { setErr(e.message) }
  }
  return (<>
    <h1>Investor</h1><p className="sub">Modal investor otomatis menjadi kas masuk kandang. Pembayaran bulanan menjadi kas keluar kandang.</p>
    <div className="stats">
      <Stat hi label="Jumlah investor" value={list.length} />
      <Stat label="Total modal" value={rp(list.reduce((a, x) => a + x.jumlah_invest, 0))} />
      <Stat label="Total dibayar bulanan" value={rp(list.reduce((a, x) => a + x.total_dibayar, 0))} />
    </div>
    <div className="panel">
      <h3>{editId ? 'Edit investor' : 'Tambah investor'}</h3>
      <form className="form" onSubmit={simpan}>
        <Field label="Nama"><input value={form.nama} onChange={e => setForm({ ...form, nama: e.target.value })} /></Field>
        <Field label="Kontak"><input value={form.kontak} onChange={e => setForm({ ...form, kontak: e.target.value })} /></Field>
        <Field label="Jumlah invest (Rp)"><input type="number" min="0" value={form.jumlah_invest} onChange={e => setForm({ ...form, jumlah_invest: e.target.value })} /></Field>
        <Field label="Mulai kerja sama"><input type="date" value={form.mulai} onChange={e => setForm({ ...form, mulai: e.target.value })} /></Field>
        <Field label="Selesai kerja sama"><input type="date" value={form.selesai} onChange={e => setForm({ ...form, selesai: e.target.value })} /></Field>
        <Field label="Nomor sapi (pisah koma)"><input placeholder="S01, S02" value={form.nomor_sapi} onChange={e => setForm({ ...form, nomor_sapi: e.target.value })} /></Field>
        <button>{editId ? 'Simpan perubahan' : 'Tambah investor'}</button>
        {editId && <button type="button" className="ghost" onClick={() => { setForm(kosongInv); setEditId(null) }}>Batal</button>}
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    <div className="panel tw">
      <table>
        <thead><tr><th>Nama</th><th>Modal</th><th>Periode</th><th>Nomor sapi</th><th>Sudah dibayar</th><th></th></tr></thead>
        <tbody>{pi.rows.map(i => <tr key={i.id}><td><b>{i.nama}</b><div style={{ color: 'var(--muted)', fontSize: 12 }}>{i.kontak}</div></td>
          <td>{rp(i.jumlah_invest)}</td><td>{tglID(i.mulai)} - {tglID(i.selesai)}</td><td>{i.nomor_sapi || '-'}</td><td>{rp(i.total_dibayar)}</td>
          <td style={{ textAlign: 'right' }}>
            <button className="ghost" onClick={() => { setEditId(i.id); setForm({ ...i, mulai: i.mulai.slice(0, 10), selesai: i.selesai.slice(0, 10) }) }}>Edit</button>{' '}
            <button className="del" onClick={() => hapus(i.id)}>Hapus</button></td></tr>)}</tbody>
      </table>
      <Pager {...pi} />
      {!list.length && <Empty text="Belum ada investor." />}
    </div>
    <div className="panel">
      <h3>Bayar bulanan investor</h3>
      <form className="form" onSubmit={simpanBayar}>
        <Field label="Investor"><select value={fb.investor_id} onChange={e => setFb({ ...fb, investor_id: e.target.value })}>
          <option value="">Pilih investor</option>{list.map(i => <option key={i.id} value={i.id}>{i.nama}</option>)}</select></Field>
        <Field label="Tanggal"><input type="date" value={fb.tgl} onChange={e => setFb({ ...fb, tgl: e.target.value })} /></Field>
        <Field label="Jumlah (Rp)"><input type="number" min="0" value={fb.jumlah} onChange={e => setFb({ ...fb, jumlah: e.target.value })} /></Field>
        <Field label="Keterangan"><input value={fb.keterangan} onChange={e => setFb({ ...fb, keterangan: e.target.value })} /></Field>
        <button>Catat pembayaran</button>
      </form>
    </div>
    <div className="panel tw">
      <h3>Riwayat pembayaran</h3>
      <table>
        <thead><tr><th>Tanggal</th><th>Investor</th><th>Jumlah</th><th>Keterangan</th><th></th></tr></thead>
        <tbody>{pb.rows.map(b => <tr key={b.id}><td>{tglID(b.tgl)}</td><td>{b.nama}</td><td><b>{rp(b.jumlah)}</b></td><td>{b.keterangan}</td>
          <td style={{ textAlign: 'right' }}><button className="del" onClick={async () => { await req(`/investor-bayar/${b.id}`, 'DELETE'); load() }}>Hapus</button></td></tr>)}</tbody>
      </table>
      <Pager {...pb} />
      {!bayar.length && <Empty text="Belum ada pembayaran." />}
    </div>
  </>)
}
