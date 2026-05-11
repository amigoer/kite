import type { Room, BaseEvent } from './types';

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

async function request<T>(input: string, init?: RequestInit): Promise<T> {
  const r = await fetch(input, init);
  if (!r.ok) {
    let code = 'http_error';
    let msg = r.statusText;
    try {
      const body = await r.json();
      code = body.error?.code ?? code;
      msg = body.error?.message ?? msg;
    } catch {
      /* ignore */
    }
    throw new ApiError(r.status, code, msg);
  }
  return r.json() as Promise<T>;
}

export async function listRooms(): Promise<Room[]> {
  const r = await request<{ rooms: Room[] }>('/api/v1/rooms');
  return r.rooms ?? [];
}

export async function getRoom(id: string): Promise<Room> {
  return request<Room>(`/api/v1/rooms/${id}`);
}

export async function getEvents(id: string, opts: { afterId?: number; limit?: number } = {}): Promise<BaseEvent[]> {
  const params = new URLSearchParams();
  if (opts.afterId) params.set('after_id', String(opts.afterId));
  if (opts.limit) params.set('limit', String(opts.limit));
  const path = `/api/v1/rooms/${id}/events${params.toString() ? '?' + params : ''}`;
  const r = await request<{ events: BaseEvent[] }>(path);
  return r.events ?? [];
}

export async function createRoom(body: { name?: string; cwd?: string }): Promise<Room> {
  return request<Room>('/api/v1/rooms', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function closeRoom(id: string): Promise<void> {
  await request(`/api/v1/rooms/${id}`, { method: 'DELETE' });
}
