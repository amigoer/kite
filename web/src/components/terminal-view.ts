import { ansiToHTML, decodeBase64 } from '../ansi';
import type { BaseEvent, TerminalOutputPayload } from '../types';

/**
 * TerminalView is a fixed-pitch, ANSI-coloured scrollback panel that
 * appends raw bytes from `terminal.output` events. It mimics a basic
 * terminal: bytes go straight in, ANSI SGR codes are coloured, and the
 * view auto-scrolls to the bottom unless the user has scrolled up.
 */
export class TerminalView {
  el: HTMLDivElement;
  private body: HTMLDivElement;
  private autoScroll = true;
  private mounted = false;

  constructor() {
    this.el = document.createElement('div');
    this.el.className = 'term-view';
    this.body = document.createElement('div');
    this.body.className = 'term-body';
    this.el.append(this.body);

    // If the user scrolls up, stop pinning the view to the bottom; resume
    // when they scroll back down to within a few px of the end.
    this.body.addEventListener('scroll', () => {
      const atBottom =
        this.body.scrollHeight - this.body.scrollTop - this.body.clientHeight < 8;
      this.autoScroll = atBottom;
    });
  }

  reset() {
    this.body.innerHTML = '';
    this.autoScroll = true;
    this.mounted = true;
  }

  /** Append already-decoded text. */
  appendText(text: string) {
    if (!text) return;
    this.body.insertAdjacentHTML('beforeend', ansiToHTML(text));
    if (this.autoScroll) {
      this.body.scrollTop = this.body.scrollHeight;
    }
  }

  /** Convenience: feed one terminal.output event. */
  applyEvent(ev: BaseEvent) {
    if (ev.type !== 'terminal.output') return;
    const p = ev.payload as TerminalOutputPayload;
    this.appendText(decodeBase64(p.data));
  }

  isMounted() {
    return this.mounted;
  }
}
