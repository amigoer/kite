import type { WSMessage } from './types';
import { wsURL } from './config';

export interface StreamHandlers {
  onMessage: (msg: WSMessage) => void;
  onOpen?: () => void;
  onClose?: () => void;
}

export class RoomStream {
  private ws: WebSocket | null = null;
  private closed = false;
  private reconnectTimer: number | null = null;

  constructor(private roomId: string, private handlers: StreamHandlers) {}

  connect() {
    const url = wsURL(`/api/v1/rooms/${this.roomId}/stream`);
    const ws = new WebSocket(url);
    this.ws = ws;

    ws.addEventListener('open', () => this.handlers.onOpen?.());
    ws.addEventListener('message', (ev) => {
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
