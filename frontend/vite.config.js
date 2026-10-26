import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    sourcemap: false, // jangan kirim source map ke produksi
    // font selalu jadi file terpisah (tidak di-inline sebagai data:) agar lolos CSP font-src 'self'
    assetsInlineLimit: (file) => (/\.(woff2?|ttf|otf)$/.test(file) ? false : undefined),
  },
})
