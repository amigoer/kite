// Minimal ANSI SGR -> HTML converter. Handles 8/16/256/24-bit color, bold,
// italic, underline. Strips other escape sequences (cursor moves, etc.).

const FG = [
  '#000000', '#cd3131', '#0dbc79', '#e5e510',
  '#2472c8', '#bc3fbc', '#11a8cd', '#e5e5e5',
  '#666666', '#f14c4c', '#23d18b', '#f5f543',
  '#3b8eea', '#d670d6', '#29b8db', '#ffffff',
];
const BG = FG;

interface SgrState {
  fg?: string;
  bg?: string;
  bold?: boolean;
  italic?: boolean;
  underline?: boolean;
  dim?: boolean;
}

function emptyState(): SgrState {
  return {};
}

function applyCodes(state: SgrState, codes: number[]) {
  for (let i = 0; i < codes.length; i++) {
    const c = codes[i];
    switch (c) {
      case 0:
        Object.assign(state, emptyState());
        Object.keys(state).forEach((k) => delete (state as any)[k]);
        break;
      case 1: state.bold = true; break;
      case 2: state.dim = true; break;
      case 3: state.italic = true; break;
      case 4: state.underline = true; break;
      case 22: state.bold = false; state.dim = false; break;
      case 23: state.italic = false; break;
      case 24: state.underline = false; break;
      case 39: state.fg = undefined; break;
      case 49: state.bg = undefined; break;
      default:
        if (c >= 30 && c <= 37) state.fg = FG[c - 30];
        else if (c >= 90 && c <= 97) state.fg = FG[c - 90 + 8];
        else if (c >= 40 && c <= 47) state.bg = BG[c - 40];
        else if (c >= 100 && c <= 107) state.bg = BG[c - 100 + 8];
        else if (c === 38 || c === 48) {
          const target = c === 38 ? 'fg' : 'bg';
          if (codes[i + 1] === 5 && codes[i + 2] !== undefined) {
            state[target] = palette256(codes[i + 2]);
            i += 2;
          } else if (codes[i + 1] === 2 && codes[i + 4] !== undefined) {
            state[target] = `rgb(${codes[i + 2]},${codes[i + 3]},${codes[i + 4]})`;
            i += 4;
          }
        }
    }
  }
}

function palette256(n: number): string {
  if (n < 16) return FG[n];
  if (n >= 232) {
    const v = 8 + (n - 232) * 10;
    return `rgb(${v},${v},${v})`;
  }
  const i = n - 16;
  const r = Math.floor(i / 36);
  const g = Math.floor((i % 36) / 6);
  const b = i % 6;
  const m = (x: number) => (x === 0 ? 0 : 55 + x * 40);
  return `rgb(${m(r)},${m(g)},${m(b)})`;
}

function escapeHTML(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

function stateToStyle(s: SgrState): string {
  const parts: string[] = [];
  if (s.fg) parts.push(`color:${s.fg}`);
  if (s.bg) parts.push(`background:${s.bg}`);
  if (s.bold) parts.push('font-weight:bold');
  if (s.italic) parts.push('font-style:italic');
  if (s.underline) parts.push('text-decoration:underline');
  if (s.dim) parts.push('opacity:0.6');
  return parts.join(';');
}

const SGR_RE = /\x1b\[([0-9;]*)m/g;
const STRIP_RE = /\x1b\[[0-9;?]*[a-zA-Z]/g;

export function ansiToHTML(input: string): string {
  // First strip non-SGR escape sequences.
  let buf = '';
  let lastIdx = 0;
  const state: SgrState = emptyState();
  let openSpan = false;

  const tokens: { kind: 'text' | 'sgr'; value: string; codes?: number[] }[] = [];
  let m: RegExpExecArray | null;
  while ((m = SGR_RE.exec(input))) {
    if (m.index > lastIdx) tokens.push({ kind: 'text', value: input.slice(lastIdx, m.index) });
    const codes = m[1] === '' ? [0] : m[1].split(';').map((x) => parseInt(x, 10) || 0);
    tokens.push({ kind: 'sgr', value: m[0], codes });
    lastIdx = m.index + m[0].length;
  }
  if (lastIdx < input.length) tokens.push({ kind: 'text', value: input.slice(lastIdx) });

  for (const t of tokens) {
    if (t.kind === 'sgr') {
      if (openSpan) {
        buf += '</span>';
        openSpan = false;
      }
      applyCodes(state, t.codes!);
      const style = stateToStyle(state);
      if (style) {
        buf += `<span style="${style}">`;
        openSpan = true;
      }
    } else {
      buf += escapeHTML(t.value.replace(STRIP_RE, ''));
    }
  }
  if (openSpan) buf += '</span>';
  return buf;
}

export function decodeBase64(s: string): string {
  const bin = atob(s);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  return new TextDecoder().decode(bytes);
}
