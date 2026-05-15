import { wsURL } from './config';
import type { WriteHolder } from './types';

/**
 * RoomIO is the browser-side companion to `kite attach`: a duplex WS to
 * /api/v1/rooms/{id}/ws?role=write. Connecting blocks (server-side) until
 * the writeArbiter grants the claim, then the socket carries:
 *   - server → client binary frames: raw PTY bytes
 *   - server → client text frames:   JSON init / event / claim_changed /
 *                                    error (we surface a couple of these
 *                                    so the caller can show a UI hint)
 *   - client → server binary frames: stdin bytes
 *   - client → server text frames:   resize / detach JSON
 */
export interface RoomIOHandlers {
  /** Called for every binary frame from the daemon (raw PTY bytes). */
  onOutput?: (bytes: Uint8Array) => void;
  /** Fired when the connection opens (claim granted). */
  onOpen?: () => void;
  /** Fired when the write claim changes hands (including release). */
  onClaimChanged?: (holder: WriteHolder | null) => void;
  /** Fired when the daemon sends a text error frame (e.g. read-only reject). */
  onError?: (code: string, message: string) => void;
  onClose?: (ev: CloseEvent) => void;
  onSocketError?: () => void;
}

export class RoomIO {
  private ws: WebSocket | null = null;
  private closed = false;
  private pendingResize: { rows: number; cols: number } | null = null;
  private pendingInput: Uint8Array[] = [];

  constructor(private roomId: string, private handlers: RoomIOHandlers = {}) {}

  connect() {
    const url = wsURL(`/api/v1/rooms/${this.roomId}/ws?role=write`);
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
      if (typeof ev.data === 'string') {
        this.handleTextFrame(ev.data);
        return;
      }
      const bytes = new Uint8Array(ev.data as ArrayBuffer);
      this.handlers.onOutput?.(bytes);
    });
    ws.addEventListener('close', (ev) => {
      this.handlers.onClose?.(ev);
    });
    ws.addEventListener('error', () => {
      this.handlers.onSocketError?.();
    });
  }

  private handleTextFrame(data: string) {
    try {
      const msg = JSON.parse(data);
      switch (msg.type) {
        case 'claim_changed':
          this.handlers.onClaimChanged?.(msg.holder ?? null);
          break;
        case 'error':
          this.handlers.onError?.(msg.code ?? '', msg.message ?? '');
          break;
        case 'init':
          // Treat the init's current_writer as the initial claim_changed.
          this.handlers.onClaimChanged?.(msg.current_writer ?? null);
          break;
      }
    } catch {
      /* malformed frame; ignore */
    }
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
