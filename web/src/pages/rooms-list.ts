import { listRooms, createRoom } from '../api';
import type { Room } from '../types';

export function renderRoomsList(host: HTMLElement) {
  host.innerHTML = '';

  const main = document.createElement('main');
  host.append(main);

  const headerBar = document.createElement('div');
  headerBar.style.display = 'flex';
  headerBar.style.alignItems = 'center';
  headerBar.style.gap = '8px';
  headerBar.style.marginBottom = '16px';

  const title = document.createElement('h2');
  title.textContent = 'Rooms';
  title.style.margin = '0';
  title.style.flex = '1';

  const newBtn = document.createElement('button');
  newBtn.textContent = '+ New room';
  newBtn.className = 'primary';
  newBtn.addEventListener('click', async () => {
    const name = prompt('Room name (optional):') ?? '';
    try {
      const room = await createRoom({ name });
      window.location.hash = `#/rooms/${room.id}`;
    } catch (err) {
      alert(`Failed to create room: ${(err as Error).message}`);
    }
  });

  const hint = document.createElement('span');
  hint.style.color = 'var(--text-dim)';
  hint.style.fontSize = '12px';
  hint.style.marginRight = '8px';
  hint.innerHTML = `or <code style="background:var(--panel);padding:2px 6px;border-radius:4px;border:1px solid var(--border)">kite shell</code> for an interactive room`;

  headerBar.append(title, hint, newBtn);
  main.append(headerBar);

  const body = document.createElement('div');
  main.append(body);

  const refresh = async () => {
    try {
      const rooms = await listRooms();
      renderTable(body, rooms);
    } catch (err) {
      body.innerHTML = `<div class="error-banner">${(err as Error).message}</div>`;
    }
  };

  refresh();
  const timer = window.setInterval(refresh, 5000);
  return () => clearInterval(timer);
}

function renderTable(host: HTMLElement, rooms: Room[]) {
  host.innerHTML = '';
  if (rooms.length === 0) {
    host.innerHTML = `<div class="empty">No rooms yet. Click <strong>+ New room</strong> or run <code>kite room create</code>.</div>`;
    return;
  }
  const table = document.createElement('table');
  table.className = 'rooms';
  table.innerHTML = `
    <thead>
      <tr>
        <th>ID</th>
        <th>Name</th>
        <th>Status</th>
        <th>Commands</th>
        <th>Cwd</th>
        <th>Created</th>
      </tr>
    </thead>
    <tbody></tbody>
  `;
  const tbody = table.querySelector('tbody')!;
  for (const r of rooms) {
    const tr = document.createElement('tr');
    tr.style.cursor = 'pointer';
    tr.addEventListener('click', () => (window.location.hash = `#/rooms/${r.id}`));
    tr.innerHTML = `
      <td class="id"><a href="#/rooms/${r.id}">${r.id}</a></td>
      <td>${escape(r.name ?? '')}</td>
      <td class="status-${r.status}">${r.status}</td>
      <td>${r.command_count}</td>
      <td>${escape(r.cwd)}</td>
      <td>${formatTime(r.created_at)}</td>
    `;
    tbody.append(tr);
  }
  host.append(table);
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  const now = Date.now();
  const diff = now - d.getTime();
  if (diff < 60_000) return 'just now';
  if (diff < 3600_000) return `${Math.floor(diff / 60_000)} min ago`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3600_000)} h ago`;
  return d.toLocaleString();
}

function escape(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
