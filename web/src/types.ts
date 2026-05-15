// Mirrors the Go types in internal/room/types.go.

export type RoomStatus = 'active' | 'closed';

export interface Room {
  id: string;
  name?: string;
  created_at: string;
  closed_at?: string | null;
  status: RoomStatus;
  cwd: string;
  shell: string;
  metadata?: Record<string, string>;
  url: string;
  command_count: number;
}

export type EventType =
  | 'room.created'
  | 'room.closed'
  | 'command.started'
  | 'command.output'
  | 'command.finished'
  | 'terminal.output'
  | 'write.claimed'
  | 'write.released';

export interface BaseEvent {
  id: number;
  room_id: string;
  timestamp: string;
  type: EventType;
  payload: unknown;
}

export interface CommandStartedPayload {
  command_id: string;
  cmd: string;
  source: string;
}

export interface CommandOutputPayload {
  command_id: string;
  stream: string;
  data: string; // base64-encoded bytes
}

export interface CommandFinishedPayload {
  command_id: string;
  exit_code: number;
  duration_ms: number;
}

export interface TerminalOutputPayload {
  data: string; // base64-encoded raw bytes
}

export interface RoomCreatedPayload {
  name?: string;
  cwd?: string;
  shell?: string;
}

export interface RoomClosedPayload {
  reason?: string;
}

export interface WriteHolder {
  id: string;
  kind: string; // "exec" | "attach" | "ws"
  label?: string;
}

export interface WriteClaimedPayload {
  holder_id: string;
  kind: string;
  label?: string;
}

export interface WriteReleasedPayload {
  holder_id: string;
}

export interface WSInitMessage {
  type: 'init';
  role: 'read' | 'write';
  room: Room;
  recent_events: BaseEvent[];
  current_writer: WriteHolder | null;
}

export interface WSEventMessage {
  type: 'event';
  event: BaseEvent;
}

export interface WSClaimChangedMessage {
  type: 'claim_changed';
  holder: WriteHolder | null;
}

export interface WSErrorMessage {
  type: 'error';
  code: string;
  message: string;
}

export type WSMessage =
  | WSInitMessage
  | WSEventMessage
  | WSClaimChangedMessage
  | WSErrorMessage;
