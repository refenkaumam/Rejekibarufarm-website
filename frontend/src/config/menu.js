import Sapi from '../pages/peternakan/Sapi'
import Pakan from '../pages/peternakan/Pakan'
import Produksi from '../pages/peternakan/Produksi'
import Penjualan from '../pages/peternakan/Penjualan'
import Investor from '../pages/peternakan/Investor'
import KeuanganKandang from '../pages/peternakan/KeuanganKandang'
import MelkSusu from '../pages/melk/MelkSusu'
import MelkKeju from '../pages/melk/MelkKeju'
import MelkRekap from '../pages/melk/MelkRekap'
import MelkKeuangan from '../pages/melk/MelkKeuangan'
import RekapBO from '../pages/backoffice/RekapBO'
import KeuanganBO from '../pages/backoffice/KeuanganBO'
import Akun from '../pages/pengaturan/Akun'

// [kunci menu, label, halaman] - kunci harus sama dengan backend (internal/services/access.go)
export const MENU = [
  { grup: 'Peternakan', items: [['sapi', 'Rekap Sapi', Sapi], ['pakan', 'Stok Pakan', Pakan], ['produksi', 'Produksi Susu', Produksi], ['penjualan', 'Penjualan Susu', Penjualan], ['investor', 'Investor', Investor], ['kas', 'Keuangan Kandang', KeuanganKandang]] },
  { grup: 'Melk by RBF', items: [['msusu', 'Susu', MelkSusu], ['mkeju', 'Keju', MelkKeju], ['mrekap', 'Rekap per 10 Hari', MelkRekap], ['mkas', 'Keuangan Melk', MelkKeuangan]] },
  { grup: 'Back Office', items: [['bo', 'Rekap Keseluruhan', RekapBO], ['bokas', 'Keuangan Back Office', KeuanganBO]] },
  { grup: 'Pengaturan', items: [['akun', 'Akun', Akun]] },
]
