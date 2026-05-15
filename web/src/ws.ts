import type { WSMessage } from './types';
import { wsURL } from './config';

export interface StreamHandlers {
  onMessage: (msg: WSMessage) => void;
  onOpen?: () => void;
  onClose?: () => void;
}

/**
 * RoomStream subscribes to the unified /ws endpoint with role=read. The
 * daemon sends a mix of text JSON frames (init / event / claim_changed /
 * error) and binary frames (raw PTY bytes); this stream only consumes the
 * JSON channel. Use RoomIO for the byte pipe.
 */
export class RoomStream {
  private ws: WebSocket | null = null;
  private closed = false;
  private reconnectTimer: number | null = null;

  constructor(private roomId: string, private handlers: StreamHandlers) {}

  connect() {
    const url = wsURL(`/api/v1/rooms/${this.roomId}/ws?role=read`);
    const ws = new WebSocket(url);
    this.ws = ws;

    ws.addEventListener('open', () => this.handlers.onOpen?.());
    ws.addEventListener('message', (ev) => {
      // Ignore binary frames — those are the live PTY byte stream and are
      // handled by RoomIO when the user enters terminal mode.
      if (typeof ev.data !== 'string') return;
      try {
        const msg = JSON.parse(ev.data) as WSMessage;
        this.handlers.onMessage(msg);
      } catch (err) {
        console.warn('[kite] bad ws message', err);
      }
    });
    ws.addEventListener('close', () => {
      this.handlers.onClose?.();
      if (!this.closed) {
        this.reconnectTimer = window.setTimeout(() => this.connect(), 1500);
      }
    });
    ws.addEventListener('error', () => {
      ws.close();
    });
  }

  close() {
    this.closed = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    this.ws?.close();
  }
}
