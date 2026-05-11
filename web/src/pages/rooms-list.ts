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

  headerBar.append(title, newBtn);
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
      <td>${new Date(r.created_at).toLocaleString()}</td>
    `;
    tbody.append(tr);
  }
  host.append(table);
}

function escape(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
