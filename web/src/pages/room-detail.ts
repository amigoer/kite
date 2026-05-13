import { getRoom, getEvents } from '../api';
import { RoomStream } from '../ws';
import { applyEvent, CommandBlock } from '../components/command-block';
import { Timeline } from '../components/timeline';
import { TerminalView } from '../components/terminal-view';
import type { BaseEvent, Room, WSMessage } from '../types';

export function renderRoomDetail(host: HTMLElement, roomId: string): () => void {
  host.innerHTML = '';
  const main = document.createElement('main');
  host.append(main);

  const meta = document.createElement('div');
  meta.className = 'room-meta';
  main.append(meta);

  const modeBar = document.createElement('div');
  modeBar.className = 'mode-bar';
  main.append(modeBar);

  const liveBtn = document.createElement('button');
  liveBtn.textContent = '● Live';
  liveBtn.className = 'primary';
  modeBar.append(liveBtn);

  const termBtn = document.createElement('button');
  termBtn.textContent = 'Terminal';
  modeBar.append(termBtn);

  const replayBtn = document.createElement('button');
  replayBtn.textContent = 'Replay';
  modeBar.append(replayBtn);

  const back = document.createElement('a');
  back.href = '#/rooms';
  back.textContent = '← back to rooms';
  back.style.marginLeft = 'auto';
  modeBar.append(back);

  const timelineHost = document.createElement('div');
  main.append(timelineHost);

  const searchHost = document.createElement('div');
  searchHost.className = 'search-row';
  main.append(searchHost);

  const blocksHost = document.createElement('div');
  main.append(blocksHost);

  // State
  const allEvents: BaseEvent[] = [];
  const allBlocks = new Map<string, CommandBlock>();
  const terminal = new TerminalView();
  let room: Room | null = null;
  let mode: 'live' | 'terminal' | 'replay' = 'live';
  let stream: RoomStream | null = null;
  let timeline: Timeline | null = null;
  let searchInput: HTMLInputElement | null = null;
  let renderedCutoff = 0; // for replay rendering: events 0..cutoff have been laid out

  const hasTerminalEvents = () => allEvents.some((e) => e.type === 'terminal.output');

  const setMode = (next: 'live' | 'terminal' | 'replay') => {
    mode = next;
    liveBtn.className = next === 'live' ? 'primary' : '';
    termBtn.className = next === 'terminal' ? 'primary' : '';
    replayBtn.className = next === 'replay' ? 'primary' : '';
    if (next === 'live') {
      timelineHost.innerHTML = '';
      searchHost.innerHTML = '';
      timeline = null;
      searchInput = null;
      rebuildLive();
    } else if (next === 'terminal') {
      timelineHost.innerHTML = '';
      searchHost.innerHTML = '';
      timeline = null;
      searchInput = null;
      rebuildTerminal();
    } else {
      setupReplayUI();
    }
  };

  const updateMeta = () => {
    if (!room) return;
    const dot = room.status === 'active' ? '<span class="live-dot"></span>' : '<span class="live-dot dim"></span>';
    meta.innerHTML = `
      ${dot}<span class="${room.status === 'active' ? 'live' : 'closed'}">${room.status.toUpperCase()}</span>
      <span class="label">id</span><span class="value" style="font-family:var(--mono)">${room.id}</span>
      ${room.name ? `<span class="label">name</span><span class="value">${escape(room.name)}</span>` : ''}
      <span class="label">cwd</span><span class="value">${escape(room.cwd)}</span>
      <span class="label">commands</span><span class="value">${room.command_count}</span>
    `;
  };

  const rebuildLive = () => {
    blocksHost.innerHTML = '';
    allBlocks.clear();
    for (const ev of allEvents) applyEvent(allBlocks, blocksHost, ev);
  };

  const rebuildTerminal = () => {
    blocksHost.innerHTML = '';
    allBlocks.clear();
    terminal.reset();
    blocksHost.append(terminal.el);
    for (const ev of allEvents) terminal.applyEvent(ev);
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
      // Going backwards: rebuild from scratch up to cutoff.
      blocksHost.innerHTML = '';
      allBlocks.clear();
      for (let i = 0; i < cutoff; i++) applyEvent(allBlocks, blocksHost, allEvents[i]);
    } else {
      for (let i = renderedCutoff; i < cutoff; i++) applyEvent(allBlocks, blocksHost, allEvents[i]);
    }
    renderedCutoff = cutoff;
  };

  liveBtn.addEventListener('click', () => setMode('live'));
  termBtn.addEventListener('click', () => setMode('terminal'));
  replayBtn.addEventListener('click', () => setMode('replay'));

  const handleWS = (msg: WSMessage) => {
    if (msg.type === 'init') {
      room = msg.room;
      updateMeta();
      allEvents.length = 0;
      allEvents.push(...msg.recent_events);
      if (mode === 'live') rebuildLive();
      else if (mode === 'terminal') rebuildTerminal();
      else timeline?.update(allEvents);
    } else if (msg.type === 'event') {
      allEvents.push(msg.event);
      if (mode === 'live') {
        applyEvent(allBlocks, blocksHost, msg.event);
      } else if (mode === 'terminal') {
        terminal.applyEvent(msg.event);
      } else {
        timeline?.update(allEvents);
      }
    }
  };

  // Initial load (covers the case where recent_events from WS init has been
  // truncated). Also gives the user something even if WS fails. If the room
  // has terminal.output events, default to Terminal view; otherwise show
  // command blocks.
  (async () => {
    try {
      room = await getRoom(roomId);
      updateMeta();
      const events = await getEvents(roomId, { limit: 100000 });
      allEvents.length = 0;
      allEvents.push(...events);
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

  // mode is set by the async initial-load handler based on whether the room
  // has terminal.output events; the default state of mode='live' covers
  // anything that races before the load completes.

  return () => {
    stream?.close();
  };
}

function escape(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
