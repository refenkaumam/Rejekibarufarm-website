import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { MotionConfig } from 'framer-motion'
import './ui/theme'
import '@fontsource-variable/fredoka'
import '@fontsource-variable/plus-jakarta-sans'
import './styles/tokens.css'
import './styles/index.css'
import './styles/illustrations.css'
import App from './App.jsx'
import NotFound from './pages/NotFound.jsx'
import Toaster from './components/Toaster.jsx'

// Aplikasi berjalan di alamat dasar; alamat lain menampilkan halaman 404
const base = import.meta.env.BASE_URL
const path = location.pathname
const dikenal = path === base || path === base + 'index.html' || path + '/' === base

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <MotionConfig reducedMotion="user">
      {dikenal ? <App /> : <NotFound />}
      <Toaster />
    </MotionConfig>
  </StrictMode>,
)
