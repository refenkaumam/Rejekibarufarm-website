import { useEffect, useState } from 'react'
import { req } from '../../api/client'
import { num, rp, berat, bln } from '../../utils/format'
import { Field, Empty, Pager } from '../../components/ui'
import { usePaged } from '../../ui/usePaged'

const PERIODE = { 1: 'Tgl 1-10', 2: 'Tgl 11-20', 3: 'Tgl 21-akhir' }

function Tabel({ judul, list, fmt }) {
  const pg = usePaged(list)
  return (
    <div className="panel tw"><h3>{judul}</h3>
      <table><thead><tr><th>Periode</th><th>Penadah</th><th>Jumlah</th><th>Total</th></tr></thead>
        <tbody>{pg.rows.map((r, i) => <tr key={i}><td>{PERIODE[r.periode]}</td><td>{r.pembeli}</td><td>{fmt(r.jumlah)}</td><td><b>{rp(r.total)}</b></td></tr>)}</tbody></table>
      <Pager {...pg} />
      {!list.length && <Empty text="Belum ada penjualan di bulan ini." />}
    </div>)
}

export default function MelkRekap() {
  const [bulan, setBulan] = useState(bln())
  const [d, setD] = useState({ susu: [], keju: [] })
  useEffect(() => { req('/melk/rekap?bulan=' + bulan).then(setD).catch(() => {}) }, [bulan])
  return (<>
    <h1>Rekap per 10 Hari</h1><p className="sub">Rekap penjualan susu dan keju per penadah, dibagi tanggal 1-10, 11-20, dan 21-akhir bulan.</p>
    <div className="panel"><Field label="Bulan"><input type="month" value={bulan} onChange={e => setBulan(e.target.value)} /></Field></div>
    <Tabel judul="Susu keluar" list={d.susu} fmt={(j) => num(j) + ' L'} />
    <Tabel judul="Keju keluar" list={d.keju} fmt={berat} />
  </>)
}
