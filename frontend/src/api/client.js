export const API = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

// Sesi login disimpan backend di cookie HttpOnly (tidak bisa dibaca JavaScript).
// Token CSRF dikirim di header untuk setiap request yang mengubah data.
const CSRF_KEY = 'rbf-csrf'

export function setCsrf(t) {
  try { t ? sessionStorage.setItem(CSRF_KEY, t) : sessionStorage.removeItem(CSRF_KEY) } catch { /* storage tidak tersedia */ }
}

const getCsrf = () => { try { return sessionStorage.getItem(CSRF_KEY) || '' } catch { return '' } }

// Semua request ke backend lewat fungsi ini (sesi dan error ditangani di sini)
export async function req(path, method = 'GET', body) {
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
  if (res.status === 401 && path !== '/login' && path !== '/logout') { localStorage.clear(); setCsrf(''); location.reload() }
  if (!res.ok) throw new Error(data.error || `Terjadi kesalahan (kode ${res.status})`)
  return data
}
