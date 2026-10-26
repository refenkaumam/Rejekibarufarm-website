// Waktu lokal -> suasana (pagi, siang, sore, malam)
export function phaseOf(d = new Date()) {
  const h = d.getHours()
  if (h >= 4 && h < 11) return 'pagi'
  if (h >= 11 && h < 15) return 'siang'
  if (h >= 15 && h < 18) return 'sore'
  return 'malam'
}

export const sapaan = (d = new Date()) => 'Selamat ' + phaseOf(d)

export const tanggalPanjang = (d = new Date()) =>
  d.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
