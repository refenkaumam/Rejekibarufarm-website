import { useState } from 'react'

// Pagination di sisi tampilan untuk daftar yang sudah diambil dari API
export function usePaged(list, size = 10) {
  const [page, setPage] = useState(1)
  const pages = Math.max(1, Math.ceil(list.length / size))
  const cur = Math.min(page, pages)
  return { rows: list.slice((cur - 1) * size, cur * size), page: cur, pages, setPage, total: list.length, size }
}
