import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';
import { decodeBase64 } from '../ansi';
import type { BaseEvent, TerminalOutputPayload } from '../types';

/**
 * TerminalView wraps an xterm.js terminal so we get a proper screen-style
 * renderer: backspace, cursor moves, line erase, ANSI colours, alt-screen
 * apps (vim/less/top) all render the way they do in a native terminal.
 *
 * It's display-only for now — the user can attach interactively from the
 * CLI for actual input.
 */
export class TerminalView {
  el: HTMLDivElement;
  private term: Terminal;
  private mounted = false;

  constructor() {
    this.el = document.createElement('div');
    this.el.className = 'term-view';

    // Inherit the current page theme so the terminal blends in.
    const isLight = document.documentElement.getAttribute('data-theme') === 'light';
    this.term = new Terminal({
      fontFamily: 'ui-monospace, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace',
      fontSize: 13,
      convertEol: true, // treat \n as \r\n on write — friendlier for our streams
      cursorBlink: false,
      disableStdin: true,
      scrollback: 5000,
      theme: isLight
        ? {
            background: '#fafbfc',
            foreground: '#1f2328',
            cursor: '#1f2328',
            black: '#24292f',
            red: '#cf222e',
            green: '#1a7f37',
            yellow: '#9a6700',
            blue: '#0969da',
            magenta: '#8250df',
            cyan: '#1b7c83',
            white: '#6e7781',
          }
        : {
            background: '#0d1117',
            foreground: '#c9d1d9',
            cursor: '#c9d1d9',
            black: '#484f58',
            red: '#ff7b72',
            green: '#3fb950',
            yellow: '#d29922',
            blue: '#58a6ff',
            magenta: '#bc8cff',
            cyan: '#39c5cf',
            white: '#b1bac4',
          },
    });
  }

  /** Mount the terminal into the DOM and size it. Idempotent. */
  private mount() {
    if (this.mounted) return;
    this.term.open(this.el);
    this.fitToHost();
    window.addEventListener('resize', this.onResize);
    this.mounted = true;
  }

  private onResize = () => {
    this.fitToHost();
  };

  private fitToHost() {
    // Estimate cell size from a hidden span; xterm.js exposes _core.actualCellWidth
    // but that's private. The simple math here gives us a working layout
    // without pulling in @xterm/addon-fit (which is 30KB on its own).
    const rect = this.el.getBoundingClientRect();
    const charWidth = 7.7; // approx for fontSize 13 monospace
    const charHeight = 17;
    const cols = Math.max(20, Math.floor(rect.width / charWidth));
    const rows = Math.max(8, Math.floor(rect.height / charHeight));
    this.term.resize(cols, rows);
  }

  reset() {
    this.mount();
    this.term.reset();
  }

  /** Write raw bytes (already decoded into a JS string). */
  writeText(text: string) {
    if (!text) return;
    this.term.write(text);
  }

  /** Convenience: feed one terminal.output event. */
  applyEvent(ev: BaseEvent) {
    if (ev.type !== 'terminal.output') return;
    const p = ev.payload as TerminalOutputPayload;
    this.writeText(decodeBase64(p.data));
  }

  dispose() {
    window.removeEventListener('resize', this.onResize);
    this.term.dispose();
    this.mounted = false;
  }
}
