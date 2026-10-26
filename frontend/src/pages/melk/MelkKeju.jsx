import { useCallback, useEffect, useState } from 'react'
import { req } from '../../api/client'
import { today, rp, tglID, berat, bln } from '../../utils/format'
import { Field, Stat, Empty, Pager } from '../../components/ui'
import { usePaged } from '../../ui/usePaged'

const kosongKeju = () => ({ tgl: today(), arah: 'masuk', jumlah: '', satuan: 'kg', harga: '', pembeli: '' })

function Tabel({ judul, list, jual, hapus }) {
  const pg = usePaged(list)
  return (
    <div className="panel tw"><h3>{judul}</h3>
      <table><thead><tr><th>Tanggal</th>{jual && <th>Dijual ke</th>}<th>Jumlah</th><th>Harga</th><th></th></tr></thead>
        <tbody>{pg.rows.map(r => <tr key={r.id}><td>{tglID(r.tgl)}</td>{jual && <td>{r.pembeli}</td>}<td>{berat(r.gram)}</td><td><b>{rp(r.harga)}</b></td>
          <td style={{ textAlign: 'right' }}><button className="del" onClick={() => hapus(r.id)}>Hapus</button></td></tr>)}</tbody></table>
      <Pager {...pg} />
      {!list.length && <Empty text="Belum ada data." />}
    </div>)
}

export default function MelkKeju() {
  const [masuk, setMasuk] = useState([])
  const [keluar, setKeluar] = useState([])
  const [form, setForm] = useState(kosongKeju())
  const [err, setErr] = useState('')
  const load = useCallback(() => {
    req('/melk/keju?arah=masuk').then(setMasuk).catch(e => setErr(e.message))
    req('/melk/keju?arah=keluar').then(setKeluar).catch(() => {})
  }, [])
  useEffect(() => { load() }, [load])
  const simpan = async (e) => {
    e.preventDefault(); setErr('')
    try {
      await req('/melk/keju', 'POST', { tgl: form.tgl, arah: form.arah, gram: Number(form.jumlah) * (form.satuan === 'kg' ? 1000 : 1), harga: Number(form.harga || 0), pembeli: form.pembeli })
      setForm({ ...kosongKeju(), tgl: form.tgl, arah: form.arah }); load()
    } catch (e) { setErr(e.message) }
  }
  const hapus = async (id) => { if (confirm('Hapus data ini? Kas terkait ikut terhapus.')) { await req(`/melk/keju/${id}`, 'DELETE'); load() } }
  const bi = (l) => l.filter(x => x.tgl.slice(0, 7) === bln())
  const sum = (l, k) => l.reduce((a, x) => a + x[k], 0)
  return (<>
    <h1>Keju Melk by RBF</h1><p className="sub">Keju masuk dari kandang menjadi kas keluar Melk. Keju keluar yang dijual menjadi kas masuk Melk.</p>
    <div className="stats">
      <Stat hi label="Keju masuk bulan ini" value={berat(sum(bi(masuk), 'gram'))} />
      <Stat label="Keju keluar bulan ini" value={berat(sum(bi(keluar), 'gram'))} />
      <Stat label="Hasil jual keju" value={rp(sum(bi(keluar), 'harga'))} />
    </div>
    <div className="panel">
      <h3>Catat keju</h3>
      <form className="form" onSubmit={simpan}>
        <Field label="Tanggal"><input type="date" value={form.tgl} onChange={e => setForm({ ...form, tgl: e.target.value })} /></Field>
        <Field label="Arah"><select value={form.arah} onChange={e => setForm({ ...form, arah: e.target.value })}><option value="masuk">Masuk (dari kandang)</option><option value="keluar">Keluar (dijual)</option></select></Field>
        {form.arah === 'keluar' && <Field label="Dijual ke"><input value={form.pembeli} onChange={e => setForm({ ...form, pembeli: e.target.value })} /></Field>}
        <Field label="Jumlah"><input type="number" step="0.01" min="0" value={form.jumlah} onChange={e => setForm({ ...form, jumlah: e.target.value })} /></Field>
        <Field label="Satuan"><select value={form.satuan} onChange={e => setForm({ ...form, satuan: e.target.value })}><option>kg</option><option>gr</option></select></Field>
        <Field label="Total harga (Rp)"><input type="number" min="0" value={form.harga} onChange={e => setForm({ ...form, harga: e.target.value })} /></Field>
        <button>Simpan</button>
      </form>
      {err && <div className="err">{err}</div>}
    </div>
    <Tabel judul="Keju keluar" list={keluar} jual hapus={hapus} />
    <Tabel judul="Keju masuk dari kandang" list={masuk} hapus={hapus} />
  </>)
}
