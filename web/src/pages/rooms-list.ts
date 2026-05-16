import { listRooms, createRoom } from '../api';
import type { Room } from '../types';

const EMPTY_ICON = `
<svg class="empty-icon" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
  <rect x="6" y="12" width="52" height="38" rx="6" stroke="currentColor" stroke-width="2"/>
  <path d="M14 22l6 6-6 6M24 36h14" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
</svg>`;

export function renderRoomsList(host: HTMLElement) {
  host.innerHTML = '';

  const main = document.createElement('main');
  host.append(main);

  const head = document.createElement('div');
  head.className = 'page-head';
  head.innerHTML = `
    <div class="titles">
      <h2>Rooms</h2>
      <div class="subtitle">
        Programmable shell sessions —
        or run <code>kite shell</code> for an interactive room.
      </div>
    </div>
    <div class="actions"></div>
  `;
  main.append(head);

  const newBtn = document.createElement('button');
  newBtn.className = 'primary';
  newBtn.innerHTML = `
    <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true" style="margin-right:6px;vertical-align:-2px">
      <path d="M8 3v10M3 8h10"/>
    </svg>New room`;
  newBtn.addEventListener('click', async () => {
    const name = prompt('Room name (optional):') ?? '';
    try {
      const room = await createRoom({ name });
      window.location.hash = `#/rooms/${room.id}`;
    } catch (err) {
      alert(`Failed to create room: ${(err as Error).message}`);
    }
  });
  head.querySelector('.actions')!.append(newBtn);

  const body = document.createElement('div');
  main.append(body);

  const refresh = async () => {
    try {
      const rooms = await listRooms();
      renderTable(body, rooms);
    } catch (err) {
      body.innerHTML = `<div class="error-banner">${escape((err as Error).message)}</div>`;
    }
  };

  refresh();
  const timer = window.setInterval(refresh, 5000);
  return () => clearInterval(timer);
}

function renderTable(host: HTMLElement, rooms: Room[]) {
  host.innerHTML = '';
  if (rooms.length === 0) {
    host.innerHTML = `
      <div class="rooms-card">
        <div class="empty">
          ${EMPTY_ICON}
          <h3>No rooms yet</h3>
          <p>Click <strong>New room</strong> above, or run <code>kite room create</code> in your shell.</p>
        </div>
      </div>`;
    return;
  }
  const card = document.createElement('div');
  card.className = 'rooms-card';
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
    tr.addEventListener('click', (e) => {
      // Allow native link-click behavior on the ID anchor itself.
      const target = e.target as HTMLElement;
      if (target.closest('a')) return;
      window.location.hash = `#/rooms/${r.id}`;
    });
    const statusLabel = r.status === 'active' ? 'active' : 'closed';
    tr.innerHTML = `
      <td class="id"><a href="#/rooms/${r.id}">${escape(r.id)}</a></td>
      <td>${escape(r.name ?? '')}</td>
      <td>
        <span class="status-cell status-${escape(r.status)}">
          <span class="dot"></span>${statusLabel}
        </span>
      </td>
      <td class="count">${r.command_count}</td>
      <td class="cwd" title="${escape(r.cwd)}">${escape(r.cwd)}</td>
      <td class="created" title="${escape(r.created_at)}">${formatTime(r.created_at)}</td>
    `;
    tbody.append(tr);
  }
  card.append(table);
  host.append(card);
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
