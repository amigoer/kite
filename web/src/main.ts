import './style.css';
import { renderRoomsList } from './pages/rooms-list';
import { renderRoomDetail } from './pages/room-detail';

let dispose: (() => void) | null = null;

function renderHeader() {
  const header = document.createElement('header');
  header.className = 'top';
  header.innerHTML = `
    <h1><a href="#/rooms">kite</a></h1>
    <span class="pill">programmable shell sessions</span>
    <span class="spacer"></span>
    <span class="pill">v0.1</span>
  `;
  return header;
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
