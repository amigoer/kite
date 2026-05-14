import { getRoom, getEvents } from '../api';
import { RoomStream } from '../ws';
import { RoomIO } from '../room-io';
import { applyEvent, CommandBlock } from '../components/command-block';
import { Timeline } from '../components/timeline';
import { TerminalView } from '../components/terminal-view';
import type { BaseEvent, Room, WSMessage } from '../types';

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

  const hasTerminalEvents = () => allEvents.some((e) => e.type === 'terminal.output');
  const isInteractive = () => room?.mode === 'interactive';

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
  };

  const refreshModeBar = () => {
    if (!room) return;
    // Interactive rooms only make sense as a terminal — hide structured
    // views. Scripted rooms show all three, with Terminal becoming visible
    // only once we have terminal.output bytes to render.
    const interactive = isInteractive();
    liveBtn.style.display = interactive ? 'none' : '';
    replayBtn.style.display = interactive ? 'none' : '';
    termBtn.style.display = interactive || hasTerminalEvents() ? '' : 'none';
    // Re-label for clarity.
    termBtn.textContent = interactive ? '● Live' : 'Terminal';
  };

  // --- Header rendering ----------------------------------------------
  const updateMeta = () => {
    if (!room) return;
    const interactive = isInteractive();
    const statusActive = room.status === 'active';
    const cwd = collapsedPath(room.cwd);
    meta.innerHTML = `
      <div class="meta-line">
        <span class="status-pill ${statusActive ? 'on' : 'off'}">
          <span class="status-dot"></span>${statusActive ? 'active' : 'closed'}
        </span>
        <span class="mode-pill ${interactive ? 'interactive' : 'scripted'}">
          ${interactive ? 'interactive' : 'scripted'}
        </span>
        <code class="room-id" title="click to copy" data-id="${room.id}">${room.id}</code>
        ${room.name ? `<span class="name">${escape(room.name)}</span>` : ''}
      </div>
      <div class="meta-line dim">
        <span title="${escape(room.shell)}">${shellName(room.shell)}</span>
        <span class="sep">·</span>
        <span title="${escape(room.cwd)}">${cwd}</span>
        ${!interactive ? `<span class="sep">·</span><span>${room.command_count ?? 0} command${(room.command_count ?? 0) === 1 ? '' : 's'}</span>` : ''}
      </div>
    `;
    // Copy room id to clipboard on click.
    const idEl = meta.querySelector<HTMLElement>('.room-id');
    if (idEl) {
      idEl.addEventListener('click', () => copyToClipboard(room!.id, idEl));
    }
    // Show CLI attach hint for active interactive rooms.
    attachHint.innerHTML = '';
    if (statusActive && interactive) {
      attachHint.innerHTML = `<code>kite attach ${room.id}</code>`;
    }
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

    // Only wire interactive input when the room is live (active) AND
    // interactive. Closed rooms or scripted history views stay read-only.
    const interactive = isInteractive() && room?.status === 'active';
    const decoder = interactive ? new TextDecoder() : null;

    terminal = new TerminalView({
      onInput: interactive
        ? (data) => {
            io?.send(data);
          }
        : undefined,
      onResize: interactive
        ? (rows, cols) => {
            io?.resize(rows, cols);
          }
        : undefined,
    });
    blocksHost.append(terminal.el);
    terminal.reset();
    for (const ev of allEvents) terminal.applyEvent(ev);

    // After history replay, snap to the latest line and re-fit once layout
    // has settled (the flex container's height isn't final until the next
    // paint).
    requestAnimationFrame(() => {
      terminal?.refit();
      terminal?.scrollToBottom();
    });

    if (interactive) {
      io = new RoomIO(roomId, {
        onOpen: () => {
          ioLive = true;
          // Push our current dimensions; /stream events are now suppressed
          // (we'll render directly from /io to avoid duplicate output).
          const dim = terminal?.dimensions();
          if (dim) io?.resize(dim.rows, dim.cols);
        },
        onOutput: (bytes) => {
          if (!terminal || !decoder) return;
          terminal.writeText(decoder.decode(bytes, { stream: true }));
        },
        onClose: () => {
          ioLive = false;
        },
      });
      io.connect();
      // Focus the terminal so the user can just start typing.
      requestAnimationFrame(() => terminal?.focus());
    }
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

  const handleWS = (msg: WSMessage) => {
    if (msg.type === 'init') {
      room = msg.room;
      updateMeta();
      allEvents.length = 0;
      allEvents.push(...msg.recent_events);
      refreshModeBar();
      if (mode === 'live' && isInteractive()) {
        // Auto-promote interactive rooms to Terminal mode.
        setMode('terminal');
      } else if (mode === 'live') rebuildLive();
      else if (mode === 'terminal') rebuildTerminal();
      else timeline?.update(allEvents);
    } else if (msg.type === 'event') {
      allEvents.push(msg.event);
      refreshModeBar();
      if (mode === 'live') {
        applyEvent(allBlocks, blocksHost, msg.event);
      } else if (mode === 'terminal') {
        // When /io is live, we already render PTY bytes directly from there;
        // applying the persisted terminal.output event would double-render.
        if (!(ioLive && msg.event.type === 'terminal.output')) {
          terminal?.applyEvent(msg.event);
        }
      } else {
        timeline?.update(allEvents);
      }
    }
  };

  // --- Initial load --------------------------------------------------
  (async () => {
    try {
      room = await getRoom(roomId);
      updateMeta();
      const events = await getEvents(roomId, { limit: 100000 });
      allEvents.length = 0;
      allEvents.push(...events);
      refreshModeBar();
      if (isInteractive() || hasTerminalEvents()) {
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
