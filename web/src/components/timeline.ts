import type { BaseEvent } from '../types';

export interface TimelineOptions {
  events: BaseEvent[];
  onPositionChange: (cutoff: number) => void;
}

/**
 * Timeline is a slider that scrubs through events by their index. The host
 * page rebuilds the command list to reflect everything up to `cutoff`.
 */
export class Timeline {
  el: HTMLDivElement;
  private slider: HTMLInputElement;
  private stamp: HTMLSpanElement;
  private speed: HTMLSelectElement;
  private playBtn: HTMLButtonElement;
  private playing = false;
  private timer: number | null = null;

  constructor(private opts: TimelineOptions) {
    this.el = document.createElement('div');
    this.el.className = 'timeline';

    const row = document.createElement('div');
    row.className = 'row';

    this.playBtn = document.createElement('button');
    this.playBtn.textContent = '▶ play';
    this.playBtn.addEventListener('click', () => this.togglePlay());

    this.slider = document.createElement('input');
    this.slider.type = 'range';
    this.slider.min = '0';
    this.slider.max = String(Math.max(opts.events.length, 1));
    this.slider.value = String(opts.events.length);
    this.slider.addEventListener('input', () => this.fire());

    this.stamp = document.createElement('span');
    this.stamp.className = 'stamp';

    this.speed = document.createElement('select');
    for (const v of ['0.5', '1', '2', 'instant']) {
      const o = document.createElement('option');
      o.value = v;
      o.textContent = v + (v === 'instant' ? '' : 'x');
      this.speed.append(o);
    }
    this.speed.value = '1';

    row.append(this.playBtn, this.slider, this.stamp, this.speed);
    this.el.append(row);
    this.fire();
  }

  update(events: BaseEvent[]) {
    this.opts.events = events;
    const wasAtEnd = parseInt(this.slider.value, 10) === parseInt(this.slider.max, 10);
    this.slider.max = String(Math.max(events.length, 1));
    if (wasAtEnd) this.slider.value = this.slider.max;
    this.fire();
  }

  jumpToEnd() {
    this.slider.value = this.slider.max;
    this.fire();
  }

  private fire() {
    const idx = parseInt(this.slider.value, 10);
    const ev = this.opts.events[idx - 1];
    this.stamp.textContent = ev
      ? new Date(ev.timestamp).toLocaleTimeString()
      : `(${this.opts.events.length} events)`;
    this.opts.onPositionChange(idx);
  }

  private togglePlay() {
    if (this.playing) {
      this.playing = false;
      this.playBtn.textContent = '▶ play';
      if (this.timer) clearInterval(this.timer);
      this.timer = null;
      return;
    }
    this.playing = true;
    this.playBtn.textContent = '⏸ pause';
    const stepMs = this.stepInterval();
    this.timer = window.setInterval(() => {
      const cur = parseInt(this.slider.value, 10);
      const max = parseInt(this.slider.max, 10);
      if (cur >= max) {
        this.togglePlay();
        return;
      }
      this.slider.value = String(cur + 1);
      this.fire();
    }, stepMs);
  }

  private stepInterval(): number {
    switch (this.speed.value) {
      case 'instant': return 1;
      case '0.5': return 200;
      case '2': return 50;
      default: return 100;
    }
  }
}
