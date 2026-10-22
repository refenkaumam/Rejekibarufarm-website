export const API = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

// Semua request ke backend lewat fungsi ini (token dan error ditangani di sini)
export async function req(path, method = 'GET', body) {
  const res = await fetch(API + path, {
    method,
    headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + (localStorage.getItem('token') || '') },
    body: body ? JSON.stringify(body) : undefined,
  })
  let data = {}
  try { data = await res.json() } catch { /* respons bukan JSON */ }
  if (res.status === 401 && path !== '/login') { localStorage.clear(); location.reload() }
  if (!res.ok) throw new Error(data.error || `Terjadi kesalahan (kode ${res.status})`)
  return data
}
