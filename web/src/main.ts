import './style.css';
import { renderRoomsList } from './pages/rooms-list';
import { renderRoomDetail } from './pages/room-detail';
import { daemonName } from './config';

let dispose: (() => void) | null = null;

type Theme = 'dark' | 'light';

function getTheme(): Theme {
  const saved = (localStorage.getItem('kite.theme') || 'dark') as Theme;
  return saved === 'light' ? 'light' : 'dark';
}

function setTheme(t: Theme) {
  document.documentElement.setAttribute('data-theme', t);
  try { localStorage.setItem('kite.theme', t); } catch (_) { /* ignore */ }
}

const KITE_LOGO_SVG = `
<svg class="logo" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
  <path d="M12 2.5 4.2 10.3a1 1 0 0 0 0 1.4l7.1 7.1a1 1 0 0 0 1.4 0l7.1-7.1a1 1 0 0 0 0-1.4L12 2.5Z"
        stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"
        fill="currentColor" fill-opacity="0.15"/>
  <path d="M12 2.5v16.8M4.2 11l15.6 0" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" opacity="0.65"/>
  <path d="M12 19.5 9 22.5M12 19.5 15 22.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/>
</svg>`;

const SUN_ICON = `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/></svg>`;
const MOON_ICON = `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8Z"/></svg>`;

function renderHeader() {
  const header = document.createElement('header');
  header.className = 'top';
  const current = getTheme();
  const next: Theme = current === 'dark' ? 'light' : 'dark';
  const icon = current === 'dark' ? SUN_ICON : MOON_ICON;
  // When the bundle is served from a hub the URL is /d/<name>/, so we surface
  // that as a pill — and link back to the picker.
  const dname = daemonName();
  const daemonPill = dname
    ? `<a href="/" class="pill" title="back to daemon picker">⤴ ${escapeAttr(dname)}</a>`
    : '';
  header.innerHTML = `
    <a href="#/rooms" class="brand">
      ${KITE_LOGO_SVG}
      <span class="wordmark">kite</span>
    </a>
    <span class="pill">programmable shell sessions</span>
    ${daemonPill}
    <span class="spacer"></span>
    <button class="icon-btn" id="theme-toggle" title="Switch to ${next} theme" aria-label="Toggle theme">${icon}</button>
    <span class="version">v0.1</span>
  `;
  header.querySelector('#theme-toggle')!.addEventListener('click', () => {
    setTheme(next);
    route();
  });
  return header;
}

function escapeAttr(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function route() {
  const app = document.getElementById('app')!;
  app.innerHTML = '';
  app.append(renderHeader());

  const host = document.createElement('div');
  app.append(host);

  if (dispose) {
    dispose();
    dispose = null;
  }

  const hash = window.location.hash || '#/rooms';
  const m = hash.match(/^#\/rooms\/([^/?]+)/);
  if (m) {
    dispose = renderRoomDetail(host, m[1]);
  } else {
    dispose = renderRoomsList(host) ?? null;
  }
}

window.addEventListener('hashchange', route);
route();
