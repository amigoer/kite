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

function renderHeader() {
  const header = document.createElement('header');
  header.className = 'top';
  const current = getTheme();
  const next: Theme = current === 'dark' ? 'light' : 'dark';
  const icon = current === 'dark' ? '☀' : '☾';
  // When the bundle is served from a hub the URL is /d/<name>/, so we
  // surface that as a pill — and link back to the picker. Direct-daemon
  // deployments (basePath === '') hide both.
  const dname = daemonName();
  const daemonPill = dname
    ? `<a href="/" class="pill" title="back to daemon picker">⤴ ${escapeAttr(dname)}</a>`
    : '';
  header.innerHTML = `
    <h1><a href="#/rooms">kite</a></h1>
    <span class="pill">programmable shell sessions</span>
    ${daemonPill}
    <span class="spacer"></span>
    <button class="icon-btn" id="theme-toggle" title="Switch to ${next} theme" aria-label="Toggle theme">${icon}</button>
    <span class="pill">v0.1</span>
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
