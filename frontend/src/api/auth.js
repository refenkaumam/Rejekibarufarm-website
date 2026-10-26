import { req, setCsrf } from './client'

// Data user di localStorage hanya untuk tampilan (nama, role, menu); hak akses tetap dicek backend.
export const getUser = () => JSON.parse(localStorage.getItem('user') || 'null')

// Keluar: backend membatalkan sesi dan menghapus cookie
export const clearSession = () => {
  req('/logout', 'POST').catch(() => {})
  localStorage.clear()
  setCsrf('')
}

function simpan(d) {
  if (d.csrf) setCsrf(d.csrf)
  localStorage.removeItem('token') // sisa versi lama yang menyimpan token di localStorage
  const u = { username: d.username, role: d.role, menus: d.menus }
  localStorage.setItem('user', JSON.stringify(u))
  return u
}

export async function login(f) {
  return simpan(await req('/login', 'POST', f))
}

// Ambil ulang role, menu, dan token CSRF terbaru dari server
export async function fetchMe() {
  return simpan(await req('/me'))
}
