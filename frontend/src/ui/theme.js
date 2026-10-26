import { useSyncExternalStore } from 'react'

// Tema terang ("padang rumput siang") dan gelap ("padang rumput malam")
const KEY = 'rbf-theme'
const listeners = new Set()

function initial() {
  try {
    const v = localStorage.getItem(KEY)
    if (v === 'light' || v === 'dark') return v
  } catch { /* storage tidak tersedia */ }
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

let theme = initial()

function apply() {
  document.documentElement.dataset.theme = theme
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', theme === 'dark' ? '#0E1822' : '#F6EFDF')
}
apply()

export function setTheme(t) {
  theme = t
  try { localStorage.setItem(KEY, t) } catch { /* storage tidak tersedia */ }
  apply()
  listeners.forEach(fn => fn())
}

export const toggleTheme = () => setTheme(theme === 'dark' ? 'light' : 'dark')

export const useTheme = () => useSyncExternalStore(
  (fn) => { listeners.add(fn); return () => listeners.delete(fn) },
  () => theme,
)
