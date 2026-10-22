import { req } from './client'

export const getUser = () => JSON.parse(localStorage.getItem('user') || 'null')
export const clearSession = () => localStorage.clear()

function simpan(d) {
  const u = { username: d.username, role: d.role, menus: d.menus }
  localStorage.setItem('user', JSON.stringify(u))
  return u
}

export async function login(f) {
  const d = await req('/login', 'POST', f)
  localStorage.setItem('token', d.token)
  return simpan(d)
}

// Ambil ulang role dan menu terbaru dari server
export async function fetchMe() {
  return simpan(await req('/me'))
}
