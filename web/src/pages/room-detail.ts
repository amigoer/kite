import { getRoom, getEvents } from '../api';
import { RoomStream } from '../ws';
import { RoomIO } from '../room-io';
import { applyEvent, CommandBlock } from '../components/command-block';
import { Timeline } from '../components/timeline';
import { TerminalView } from '../components/terminal-view';
import type { BaseEvent, Room, WriteHolder, WSMessage } from '../types';

type Mode = 'live' | 'terminal' | 'replay';

export function renderRoomDetail(host: HTMLElement, roomId: string): () => void {
  host.innerHTML = '';
  const main = document.createElement('main');
  main.className = 'room-detail';
  host.append(main);

  // --- Header card ----------------------------------------------------
  const card = document.createElement('div');
  card.className = 'room-card';
  main.append(card);

  const meta = document.createElement('div');
  meta.className = 'room-meta';
  card.append(meta);

  const modeBar = document.createElement('div');
  modeBar.className = 'mode-bar';
  card.append(modeBar);

  // mode buttons live in a left-aligned group; back link is on the right.
  const modeGroup = document.createElement('div');
  modeGroup.className = 'mode-group';
  modeBar.append(modeGroup);

  const liveBtn = makeBtn('● Live');
  const termBtn = makeBtn('Terminal');
  const replayBtn = makeBtn('Replay');
  modeGroup.append(liveBtn, termBtn, replayBtn);

  const right = document.createElement('div');
  right.className = 'mode-right';
  modeBar.append(right);

  const writerPill = document.createElement('span');
  writerPill.className = 'writer-pill';
  right.append(writerPill);

  const takeControlBtn = makeBtn('Take control');
  takeControlBtn.className = 'take-control';
  right.append(takeControlBtn);

  const attachHint = document.createElement('span');
  attachHint.className = 'attach-hint';
  right.append(attachHint);

  const back = document.createElement('a');
  back.href = '#/rooms';
  back.textContent = '← back to rooms';
  right.append(back);

  // --- Body -----------------------------------------------------------
  const timelineHost = document.createElement('div');
  main.append(timelineHost);

  const searchHost = document.createElement('div');
  searchHost.className = 'search-row';
  main.append(searchHost);

  const blocksHost = document.createElement('div');
  blocksHost.className = 'blocks-host';
  main.append(blocksHost);

  // --- State ----------------------------------------------------------
  const allEvents: BaseEvent[] = [];
  const allBlocks = new Map<string, CommandBlock>();
  let terminal: TerminalView | null = null;
  let room: Room | null = null;
  let mode: Mode = 'live';
  let stream: RoomStream | null = null;
  let io: RoomIO | null = null;
  let timeline: Timeline | null = null;
  let searchInput: HTMLInputElement | null = null;
  let renderedCutoff = 0;
  let currentWriter: WriteHolder | null = null;
  let wantsWrite = false;

  const hasTerminalEvents = () => allEvents.some((e) => e.type === 'terminal.output');

  // --- Mode handling --------------------------------------------------
  const disposeTerminal = () => {
    if (terminal) {
      terminal.dispose();
      terminal = null;
    }
    if (io) {
      io.close();
      io = null;
    }
  };

  /** True while we're hooked up to /io and rendering live PTY bytes from it
   *  instead of from the slower /stream terminal.output events. */
  let ioLive = false;

  const setMode = (next: Mode) => {
    if (mode === 'terminal' && next !== 'terminal') disposeTerminal();
    mode = next;
    liveBtn.classList.toggle('primary', next === 'live');
    termBtn.classList.toggle('primary', next === 'terminal');
    replayBtn.classList.toggle('primary', next === 'replay');

    // Toggle a class on <main> so CSS can switch to a viewport-locked flex
    // layout in terminal mode (otherwise the page scrolls and rows can be
    // clipped mid-cell at the boundary).
    main.classList.toggle('terminal-mode', next === 'terminal');

    timelineHost.innerHTML = '';
    searchHost.innerHTML = '';
    timeline = null;
    searchInput = null;

    if (next === 'live') rebuildLive();
    else if (next === 'terminal') rebuildTerminal();
    else setupReplayUI();
    updateWriterUI();
  };

  const refreshModeBar = () => {
    if (!room) return;
    // All three views are always available; Terminal becomes the most
    // useful once there's PTY activity to render. We don't hide buttons
    // anymore — the user picks.
    liveBtn.style.display = '';
    termBtn.style.display = '';
    replayBtn.style.display = '';
    termBtn.textContent = 'Terminal';
  };

  const updateWriterUI = () => {
    if (!room) return;
    if (currentWriter) {
      const who = currentWriter.label || currentWriter.id;
      writerPill.textContent = `writer: ${who} (${currentWriter.kind})`;
      writerPill.classList.add('busy');
    } else {
      writerPill.textContent = 'idle';
      writerPill.classList.remove('busy');
    }
    // Take-control is only meaningful in Terminal mode on an active room.
    const canControl = room.status === 'active' && mode === 'terminal';
    takeControlBtn.style.display = canControl ? '' : 'none';
    if (!canControl) return;
    if (wantsWrite) {
      // We currently hold (or are queued for) the claim — offer release.
      takeControlBtn.textContent = currentWriter ? 'Release control' : 'Waiting for claim…';
      takeControlBtn.disabled = !currentWriter;
    } else {
      takeControlBtn.textContent = currentWriter ? 'Take control (queue)' : 'Take control';
      takeControlBtn.disabled = false;
    }
  };

  // --- Header rendering ----------------------------------------------
  const updateMeta = () => {
    if (!room) return;
    const statusActive = room.status === 'active';
    const cwd = collapsedPath(room.cwd);
    meta.innerHTML = `
      <div class="meta-line">
        <span class="status-pill ${statusActive ? 'on' : 'off'}">
          <span class="status-dot"></span>${statusActive ? 'active' : 'closed'}
        </span>
        <code class="room-id" title="click to copy" data-id="${room.id}">${room.id}</code>
        ${room.name ? `<span class="name">${escape(room.name)}</span>` : ''}
      </div>
      <div class="meta-line dim">
        <span title="${escape(room.shell)}">${shellName(room.shell)}</span>
        <span class="sep">·</span>
        <span title="${escape(room.cwd)}">${cwd}</span>
        <span class="sep">·</span>
        <span>${room.command_count ?? 0} command${(room.command_count ?? 0) === 1 ? '' : 's'}</span>
      </div>
    `;
    const idEl = meta.querySelector<HTMLElement>('.room-id');
    if (idEl) {
      idEl.addEventListener('click', () => copyToClipboard(room!.id, idEl));
    }
    attachHint.innerHTML = statusActive ? `<code>kite attach ${room.id}</code>` : '';
  };

  // --- Renderers ------------------------------------------------------
  const rebuildLive = () => {
    blocksHost.innerHTML = '';
    allBlocks.clear();
    for (const ev of allEvents) applyEvent(allBlocks, blocksHost, ev);
    if (!allBlocks.size) {
      blocksHost.innerHTML = `<div class="empty">no commands yet — agents talk here via <code>POST /api/v1/rooms/${roomId}/exec</code></div>`;
    }
  };

  const rebuildTerminal = () => {
    blocksHost.innerHTML = '';
    allBlocks.clear();
    disposeTerminal();
    ioLive = false;

    const writable = wantsWrite && room?.status === 'active';
    const decoder = writable ? new TextDecoder() : null;

    terminal = new TerminalView({
      onInput: writable
        ? (data) => {
            io?.send(data);
          }
        : undefined,
      onResize: writable
        ? (rows, cols) => {
            io?.resize(rows, cols);
          }
        : undefined,
    });
    blocksHost.append(terminal.el);
    terminal.reset();
    for (const ev of allEvents) terminal.applyEvent(ev);

    requestAnimationFrame(() => {
      terminal?.refit();
      terminal?.scrollToBottom();
    });

    if (writable) {
      io = new RoomIO(roomId, {
        onOpen: () => {
          ioLive = true;
          const dim = terminal?.dimensions();
          if (dim) io?.resize(dim.rows, dim.cols);
        },
        onOutput: (bytes) => {
          if (!terminal || !decoder) return;
          terminal.writeText(decoder.decode(bytes, { stream: true }));
        },
        onClaimChanged: (holder) => {
          currentWriter = holder;
          updateWriterUI();
        },
        onClose: () => {
          ioLive = false;
          wantsWrite = false;
          updateWriterUI();
        },
      });
      io.connect();
      requestAnimationFrame(() => terminal?.focus());
    }
    updateWriterUI();
  };

  const setupReplayUI = () => {
    blocksHost.innerHTML = '';
    allBlocks.clear();
    renderedCutoff = 0;

    timeline = new Timeline({
      events: allEvents,
      onPositionChange: (cutoff) => renderReplayCutoff(cutoff),
    });
    timelineHost.innerHTML = '';
    timelineHost.append(timeline.el);

    searchHost.innerHTML = '';
    searchInput = document.createElement('input');
    searchInput.placeholder = 'filter commands by text';
    searchInput.addEventListener('input', () => {
      const q = searchInput!.value.toLowerCase();
      for (const b of allBlocks.values()) {
        b.el.style.display = !q || b.state.cmd.toLowerCase().includes(q) ? '' : 'none';
      }
    });
    searchHost.append(searchInput);
  };

  const renderReplayCutoff = (cutoff: number) => {
    if (cutoff < renderedCutoff) {
      blocksHost.innerHTML = '';
      allBlocks.clear();
      for (let i = 0; i < cutoff; i++) applyEvent(allBlocks, blocksHost, allEvents[i]);
    } else {
      for (let i = renderedCutoff; i < cutoff; i++) applyEvent(allBlocks, blocksHost, allEvents[i]);
    }
    renderedCutoff = cutoff;
  };

  // --- Wiring ---------------------------------------------------------
  liveBtn.addEventListener('click', () => setMode('live'));
  termBtn.addEventListener('click', () => setMode('terminal'));
  replayBtn.addEventListener('click', () => setMode('replay'));

  takeControlBtn.addEventListener('click', () => {
    if (wantsWrite) {
      // Release: close the RoomIO, drop out of writable mode, re-render
      // terminal read-only.
      wantsWrite = false;
      rebuildTerminal();
      return;
    }
    // Acquire: switch into writable mode and let RoomIO queue for the claim.
    wantsWrite = true;
    if (mode !== 'terminal') setMode('terminal');
    else rebuildTerminal();
  });

  const handleWS = (msg: WSMessage) => {
    if (msg.type === 'init') {
      room = msg.room;
      currentWriter = msg.current_writer;
      updateMeta();
      updateWriterUI();
      allEvents.length = 0;
      allEvents.push(...msg.recent_events);
      refreshModeBar();
      if (mode === 'live') rebuildLive();
      else if (mode === 'terminal') rebuildTerminal();
      else timeline?.update(allEvents);
    } else if (msg.type === 'event') {
      allEvents.push(msg.event);
      refreshModeBar();
      if (mode === 'live') {
        applyEvent(allBlocks, blocksHost, msg.event);
      } else if (mode === 'terminal') {
        if (!(ioLive && msg.event.type === 'terminal.output')) {
          terminal?.applyEvent(msg.event);
        }
      } else {
        timeline?.update(allEvents);
      }
    } else if (msg.type === 'claim_changed') {
      currentWriter = msg.holder;
      updateWriterUI();
    } else if (msg.type === 'error') {
      console.warn('[kite] ws error', msg.code, msg.message);
    }
  };

  // --- Initial load --------------------------------------------------
  (async () => {
    try {
      room = await getRoom(roomId);
      updateMeta();
      updateWriterUI();
      const events = await getEvents(roomId, { limit: 100000 });
      allEvents.length = 0;
      allEvents.push(...events);
      refreshModeBar();
      // Default to the command-block dashboard, but jump straight to the
      // terminal if there's already raw PTY history to show.
      if (hasTerminalEvents()) {
        setMode('terminal');
      } else {
        rebuildLive();
      }
    } catch (err) {
      blocksHost.innerHTML = `<div class="error-banner">${(err as Error).message}</div>`;
      return;
    }
  })();

  stream = new RoomStream(roomId, { onMessage: handleWS });
  stream.connect();

  return () => {
    stream?.close();
    disposeTerminal();
    main.classList.remove('terminal-mode');
  };
}

// --- helpers --------------------------------------------------------

function makeBtn(text: string): HTMLButtonElement {
  const b = document.createElement('button');
  b.textContent = text;
  return b;
}

function escape(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function shellName(p: string): string {
  if (!p) return 'shell';
  const parts = p.split('/');
  return parts[parts.length - 1] || p;
}

function collapsedPath(p: string): string {
  if (!p) return '';
  const home = guessHome();
  if (home && p.startsWith(home)) return '~' + p.slice(home.length);
  return p;
}

let cachedHome: string | null = null;
function guessHome(): string | null {
  if (cachedHome !== null) return cachedHome;
  // The room's cwd often starts with /Users/foo or /home/foo. We can't ask
  // the OS from the browser, but the path itself reveals it.
  return (cachedHome = null);
}

function copyToClipboard(text: string, srcEl: HTMLElement) {
  navigator.clipboard?.writeText(text).then(
    () => {
      const before = srcEl.textContent;
      srcEl.textContent = 'copied!';
      setTimeout(() => {
        srcEl.textContent = before ?? text;
      }, 900);
    },
    () => {
      /* ignore */
    },
  );
}
