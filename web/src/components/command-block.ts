import type {
  BaseEvent,
  CommandStartedPayload,
  CommandOutputPayload,
  CommandFinishedPayload,
} from '../types';
import { ansiToHTML, decodeBase64 } from '../ansi';

export interface CommandState {
  id: string;
  cmd: string;
  source: string;
  startedAt: Date;
  finishedAt?: Date;
  exitCode?: number;
  durationMs?: number;
  output: string;
  collapsed: boolean;
}

/**
 * CommandBlock renders one command-event-derived dashboard card. It is
 * stateful: feed it events as they arrive and it updates the DOM in place.
 */
export class CommandBlock {
  el: HTMLDivElement;
  private bodyEl: HTMLDivElement;
  private metaEl: HTMLSpanElement;
  private toggleEl: HTMLSpanElement;

  constructor(public state: CommandState) {
    this.el = document.createElement('div');
    this.el.className = 'cmd-block running' + (state.collapsed ? ' collapsed' : '');
    this.el.dataset.id = state.id;

    const head = document.createElement('div');
    head.className = 'head';

    const prompt = document.createElement('span');
    prompt.className = 'prompt';
    prompt.textContent = '$';

    const cmd = document.createElement('span');
    cmd.className = 'cmd';
    cmd.textContent = state.cmd;

    this.metaEl = document.createElement('span');
    this.metaEl.className = 'meta';

    this.toggleEl = document.createElement('span');
    this.toggleEl.className = 'toggle';
    this.toggleEl.textContent = state.collapsed ? '▸' : '▾';

    head.append(prompt, cmd, this.metaEl, this.toggleEl);
    head.addEventListener('click', () => this.toggle());

    this.bodyEl = document.createElement('div');
    this.bodyEl.className = 'body';
    if (state.output) this.bodyEl.innerHTML = ansiToHTML(state.output);

    this.el.append(head, this.bodyEl);
    this.renderMeta();
  }

  toggle() {
    this.state.collapsed = !this.state.collapsed;
    this.el.classList.toggle('collapsed', this.state.collapsed);
    this.toggleEl.textContent = this.state.collapsed ? '▸' : '▾';
  }

  appendOutput(text: string) {
    this.state.output += text;
    this.bodyEl.innerHTML += ansiToHTML(text);
    this.bodyEl.scrollTop = this.bodyEl.scrollHeight;
  }

  finish(exitCode: number, durationMs: number, finishedAt: Date) {
    this.state.exitCode = exitCode;
    this.state.durationMs = durationMs;
    this.state.finishedAt = finishedAt;
    this.el.classList.remove('running');
    this.el.classList.add(exitCode === 0 ? 'success' : 'failed');
    this.renderMeta();
  }

  private renderMeta() {
    const s = this.state;
    if (s.exitCode === undefined) {
      this.metaEl.innerHTML = `<span class="spinner"></span> running…`;
      return;
    }
    const cls = s.exitCode === 0 ? 'exit-ok' : 'exit-fail';
    const sym = s.exitCode === 0 ? '✓' : '✗';
    this.metaEl.innerHTML =
      `<span class="${cls}">${sym} exit ${s.exitCode}</span>` +
      ` · ${s.durationMs ?? 0}ms`;
  }
}

/**
 * applyEvent merges one event into the (id -> block) map, creating new
 * blocks as needed. Order of events from the daemon is canonical (id asc),
 * so callers should just iterate.
 */
export function applyEvent(
  blocks: Map<string, CommandBlock>,
  container: HTMLElement,
  ev: BaseEvent,
) {
  if (ev.type === 'command.started') {
    const p = ev.payload as CommandStartedPayload;
    if (blocks.has(p.command_id)) return; // already rendered (recent_events + ws splice)
    const state: CommandState = {
      id: p.command_id,
      cmd: p.cmd,
      source: p.source,
      startedAt: new Date(ev.timestamp),
      output: '',
      collapsed: false,
    };
    const block = new CommandBlock(state);
    blocks.set(p.command_id, block);
    container.append(block.el);
    block.el.scrollIntoView({ block: 'end', behavior: 'smooth' });
  } else if (ev.type === 'command.output') {
    const p = ev.payload as CommandOutputPayload;
    const block = blocks.get(p.command_id);
    if (!block) return;
    block.appendOutput(decodeBase64(p.data));
  } else if (ev.type === 'command.finished') {
    const p = ev.payload as CommandFinishedPayload;
    const block = blocks.get(p.command_id);
    if (!block) return;
    block.finish(p.exit_code, p.duration_ms, new Date(ev.timestamp));
  }
}
