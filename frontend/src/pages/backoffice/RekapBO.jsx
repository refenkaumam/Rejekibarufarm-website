import { useEffect, useState } from 'react'
import { req, API } from '../../api/client'
import { today, rp } from '../../utils/format'
import { Field, Stat } from '../../components/ui'

const awalBulan = () => today().slice(0, 7) + '-01'

const akhirBulan = () => { const d = new Date(); return new Date(d.getFullYear(), d.getMonth() + 1, 0).toLocaleDateString('sv-SE') }

const boleh = () => ['super_admin', 'owner'].includes((JSON.parse(localStorage.getItem('user') || '{}')).role)

export default function RekapBO() {
  const [r, setR] = useState({})
  const [dari, setDari] = useState(awalBulan())
  const [sampai, setSampai] = useState(akhirBulan())
  const [d, setD] = useState({ lr: { pendapatan: [], biaya: [], total_pendapatan: 0, total_biaya: 0, laba: 0 } })
  useEffect(() => { req('/bo/ringkasan').then(setR).catch(() => {}) }, [])
  useEffect(() => { req(`/bo/labarugi?dari=${dari}&sampai=${sampai}`).then(setD).catch(() => {}) }, [dari, sampai])
  const unduh = async () => {
    const res = await fetch(`${API}/bo/export?dari=${dari}&sampai=${sampai}`, { headers: { Authorization: 'Bearer ' + localStorage.getItem('token') } })
    const a = document.createElement('a')
    a.href = URL.createObjectURL(await res.blob()); a.download = `laporan_${dari}_${sampai}.csv`; a.click()
  }
  const lr = d.lr
  const Tabel = ({ judul, list, total }) => (
    <div className="panel tw"><h3>{judul}</h3>
      <table><tbody>{list.map(x => <tr key={x.label}><td>{x.label}</td><td style={{ textAlign: 'right' }}>{rp(x.jumlah)}</td></tr>)}
        <tr><td><b>Total {judul.toLowerCase()}</b></td><td style={{ textAlign: 'right' }}><b>{rp(total)}</b></td></tr></tbody></table>
    </div>)
  return (<>
    <h1>Rekap Keuangan Keseluruhan</h1><p className="sub">Gabungan kas kandang, Melk by RBF, dan back office.</p>
    <div className="stats">
      <Stat hi label="Total kas semua unit" value={rp(r.saldo_total || 0)} />
      <Stat label="Kas back office" value={rp(r.saldo_bo || 0)} />
      <Stat label="Kas kandang" value={rp(r.saldo_kandang || 0)} />
      <Stat label="Kas Melk by RBF" value={rp(r.saldo_melk || 0)} />
      <Stat label="Laba bulan ini" value={rp(r.laba_bulan_ini || 0)} />
    </div>
    <div className="panel">
      <h3>Laporan laba rugi</h3>
      <div className="form">
        <Field label="Dari"><input type="date" value={dari} onChange={e => setDari(e.target.value)} /></Field>
        <Field label="Sampai"><input type="date" value={sampai} onChange={e => setSampai(e.target.value)} /></Field>
        {boleh() && <button onClick={unduh}>Unduh Excel (CSV)</button>}
        <button className="ghost" onClick={() => window.print()}>Cetak / simpan PDF</button>
      </div>
      <p className="sub" style={{ margin: '12px 0 0' }}>Penjualan susu kandang ke Melk by RBF tidak dihitung dua kali. Setoran kas, top up, dan modal investor bukan pendapatan.</p>
    </div>
    <div className="stats">
      <Stat label="Total pendapatan" value={rp(lr.total_pendapatan)} />
      <Stat label="Total biaya" value={rp(lr.total_biaya)} />
      <Stat hi label={lr.laba >= 0 ? 'Laba' : 'Rugi'} value={rp(Math.abs(lr.laba))} />
    </div>
    <Tabel judul="Pendapatan" list={lr.pendapatan} total={lr.total_pendapatan} />
    <Tabel judul="Biaya" list={lr.biaya} total={lr.total_biaya} />
  </>)
}
