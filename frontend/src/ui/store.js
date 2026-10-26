import { useEffect, useState, useSyncExternalStore } from 'react'

// State tampilan global: jumlah request yang sedang berjalan dan daftar toast.
// Murni untuk UI (loading indicator, skeleton, notifikasi), tidak menyimpan data bisnis.
let state = { pending: 0, toasts: [] }
const listeners = new Set()

const set = (patch) => {
  state = { ...state, ...patch }
  listeners.forEach(fn => fn())
}

const subscribe = (fn) => {
  listeners.add(fn)
  return () => listeners.delete(fn)
}

export const track = (delta) => set({ pending: Math.max(0, state.pending + delta) })

let nextId = 0
export function toast(type, text) {
  const id = ++nextId
  set({ toasts: [...state.toasts, { id, type, text }] })
  setTimeout(() => dismiss(id), 3800)
}

export const dismiss = (id) => set({ toasts: state.toasts.filter(t => t.id !== id) })

export const usePending = () => useSyncExternalStore(subscribe, () => state.pending > 0)

export const useToasts = () => useSyncExternalStore(subscribe, () => state.toasts)

// true hanya kalau loading sudah berlangsung lebih dari `delay` ms, supaya tidak berkedip
export function useLoadingDelayed(delay = 220) {
  const pending = usePending()
  const [late, setLate] = useState(false)
  useEffect(() => {
    const t = setTimeout(() => setLate(pending), pending ? delay : 0)
    return () => clearTimeout(t)
  }, [pending, delay])
  return pending && late
}
