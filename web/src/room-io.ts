import { wsURL } from './config';

/**
 * RoomIO is the browser-side companion to `kite attach`: a raw binary WS
 * to /api/v1/rooms/{id}/io. Binary frames are stdin keystrokes; text
 * frames are JSON control messages (resize, etc.). The server broadcasts
 * PTY output back on the same socket as binary frames.
 */
export interface RoomIOHandlers {
  /** Called for every binary frame from the daemon (raw PTY bytes). */
  onOutput?: (bytes: Uint8Array) => void;
  onOpen?: () => void;
  onClose?: (ev: CloseEvent) => void;
  onError?: () => void;
}

export class RoomIO {
  private ws: WebSocket | null = null;
  private closed = false;
  private pendingResize: { rows: number; cols: number } | null = null;
  private pendingInput: Uint8Array[] = [];

  constructor(private roomId: string, private handlers: RoomIOHandlers = {}) {}

  connect() {
    const url = wsURL(`/api/v1/rooms/${this.roomId}/io`);
    const ws = new WebSocket(url);
    ws.binaryType = 'arraybuffer';
    this.ws = ws;

    ws.addEventListener('open', () => {
      this.handlers.onOpen?.();
      // Flush whatever the consumer queued before we were open.
      if (this.pendingResize) {
        const { rows, cols } = this.pendingResize;
        this.pendingResize = null;
        this.resize(rows, cols);
      }
      for (const buf of this.pendingInput) {
        ws.send(buf);
      }
      this.pendingInput = [];
    });
    ws.addEventListener('message', (ev) => {
      if (typeof ev.data === 'string') return; // we don't expect text from server
      const bytes = new Uint8Array(ev.data as ArrayBuffer);
      this.handlers.onOutput?.(bytes);
    });
    ws.addEventListener('close', (ev) => {
      this.handlers.onClose?.(ev);
    });
    ws.addEventListener('error', () => {
      this.handlers.onError?.();
    });
  }

  /** Send raw stdin bytes (already encoded as UTF-8). */
  send(data: string | Uint8Array) {
    const bytes = typeof data === 'string' ? new TextEncoder().encode(data) : data;
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(bytes);
    } else if (!this.closed) {
      this.pendingInput.push(bytes);
    }
  }

  /** Tell the daemon the terminal cell dimensions. */
  resize(rows: number, cols: number) {
    if (rows <= 0 || cols <= 0) return;
    const payload = JSON.stringify({ type: 'resize', rows, cols });
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(payload);
    } else if (!this.closed) {
      // Keep only the most-recent resize.
      this.pendingResize = { rows, cols };
    }
  }

  close() {
    this.closed = true;
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify({ type: 'detach' }));
      } catch {
        /* ignore */
      }
    }
    this.ws?.close();
    this.ws = null;
  }
}
