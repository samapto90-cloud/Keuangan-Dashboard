export class MusicManager {
  private ctx: AudioContext | null = null;
  private osc: OscillatorNode | null = null;
  private gain: GainNode | null = null;
  private track = "";
  master = 1;
  music = 0.6;

  setVolume(master: number, music: number): void {
    this.master = master;
    this.music = music;
    if (this.gain) this.gain.gain.value = 0.012 * this.master * this.music;
  }

  set(track: string): void {
    if (track === this.track) return;
    this.track = track;
    this.start(freqFor(track));
  }

  combat(on: boolean): void {
    this.set(on ? "combat" : this.track === "combat" ? "explore" : this.track);
  }

  private start(freq: number): void {
    try {
      if (!this.ctx) this.ctx = new AudioContext();
      const ctx = this.ctx;
      if (ctx.state === "suspended") void ctx.resume();
      this.osc?.stop();
      this.osc = ctx.createOscillator();
      this.gain = ctx.createGain();
      this.osc.type = "sine";
      this.osc.frequency.value = freq;
      this.gain.gain.value = 0.012 * this.master * this.music;
      this.osc.connect(this.gain);
      this.gain.connect(ctx.destination);
      this.osc.start();
    } catch {
      /* optional */
    }
  }
}

function freqFor(track: string): number {
  if (track === "combat") return 196;
  if (track === "boss") return 147;
  if (track === "victory") return 330;
  if (track === "forest") return 220;
  if (track === "event") return 174;
  if (track === "calm") return 196;
  if (track === "dawn") return 294;
  if (track === "valley") return 196;
  if (track === "city") return 247;
  if (track === "siluman") return 185;
  if (track === "dark") return 131;
  if (track === "light") return 349;
  if (track === "war") return 156;
  if (track === "ending") return 392;
  return 262;
}
