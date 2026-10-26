import { track, toast } from '../ui/store'

export const API = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

// Sesi login disimpan backend di cookie HttpOnly (tidak bisa dibaca JavaScript).
// Token CSRF dikirim di header untuk setiap request yang mengubah data.
const CSRF_KEY = 'rbf-csrf'

export function setCsrf(t) {
  try { t ? sessionStorage.setItem(CSRF_KEY, t) : sessionStorage.removeItem(CSRF_KEY) } catch { /* storage tidak tersedia */ }
}

const getCsrf = () => { try { return sessionStorage.getItem(CSRF_KEY) || '' } catch { return '' } }

const tanpaNotif = (path) => path === '/login' || path === '/logout'

// Semua request ke backend lewat fungsi ini (sesi dan error ditangani di sini)
export async function req(path, method = 'GET', body) {
  track(1) // indikator loading (tampilan saja)
  try {
    const headers = { 'Content-Type': 'application/json' }
    if (method !== 'GET') headers['X-CSRF-Token'] = getCsrf()
    const res = await fetch(API + path, {
      method,
      credentials: 'include',
      headers,
      body: body ? JSON.stringify(body) : undefined,
    })
    let data = {}
    try { data = await res.json() } catch { /* respons bukan JSON */ }
    if (res.status === 401 && !tanpaNotif(path)) { localStorage.clear(); setCsrf(''); location.reload() }
    if (!res.ok) {
      const msg = data.error || `Terjadi kesalahan (kode ${res.status})`
      if (method === 'DELETE') toast('error', msg)
      throw new Error(msg)
    }
    if (method !== 'GET' && !tanpaNotif(path)) toast('success', method === 'DELETE' ? 'Data berhasil dihapus' : 'Data berhasil disimpan')
    return data
  } finally {
    track(-1)
  }
}
