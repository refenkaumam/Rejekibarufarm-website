export const today = () => new Date().toLocaleDateString('sv-SE')

export const num = (n) => Number(n).toLocaleString('id-ID', { maximumFractionDigits: 2 })

export const rp = (n) => 'Rp ' + Number(n).toLocaleString('id-ID')

export const tglID = (s) => new Date(s.slice(0, 10) + 'T00:00:00').toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })

export const berat = (g) => g >= 1000 ? num(g / 1000) + ' kg' : num(g) + ' gr'

export const bln = () => today().slice(0, 7)
