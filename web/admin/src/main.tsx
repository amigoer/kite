import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

if (document.title === '__KITE_ADMIN_TITLE__') {
  document.title = 'Kite 管理后台'
}

const faviconLink = document.querySelector<HTMLLinkElement>("link[rel='icon']")
if (faviconLink?.getAttribute('href') === '__KITE_FAVICON_URL__') {
  faviconLink.setAttribute('href', '/favicon.svg')
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
)

// Fade out the pre-mount splash once React has painted.
requestAnimationFrame(() => {
  requestAnimationFrame(() => {
    document.getElementById('splash')?.classList.add('ready')
  })
})
