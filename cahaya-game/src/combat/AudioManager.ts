export class AudioManager {
  private ctx: AudioContext | null = null;
  master = 1;
  sfx = 0.8;

  punch(): void {
    this.blip(190, 0.07, "square", 0.04);
  }

  kick(): void {
    this.blip(110, 0.09, "sawtooth", 0.05);
  }

  hit(): void {
    this.blip(90, 0.06, "triangle", 0.05);
  }

  dodge(): void {
    this.blip(420, 0.08, "sine", 0.03);
  }

  charge(): void {
    this.blip(220, 0.2, "sine", 0.035);
  }

  perfect(): void {
    this.blip(660, 0.1, "triangle", 0.04);
  }

  heavy(): void {
    this.blip(80, 0.1, "sawtooth", 0.055);
  }

  crit(): void {
    this.blip(720, 0.08, "square", 0.04);
  }

  energy(): void {
    this.blip(540, 0.12, "sine", 0.045);
  }

  mine(): void {
    this.blip(95, 0.1, "square", 0.05);
  }

  woodcut(): void {
    this.blip(140, 0.09, "sawtooth", 0.045);
  }

  gather(): void {
    this.blip(310, 0.07, "sine", 0.03);
  }

  fish(): void {
    this.blip(380, 0.11, "triangle", 0.035);
  }

  craft(): void {
    this.blip(240, 0.12, "triangle", 0.04);
    window.setTimeout(() => this.blip(480, 0.08, "sine", 0.03), 70);
  }

  buy(): void {
    this.blip(520, 0.07, "sine", 0.035);
  }

  sell(): void {
    this.blip(260, 0.07, "triangle", 0.03);
  }

  transform(): void {
    this.blip(280, 0.18, "triangle", 0.05);
    window.setTimeout(() => this.blip(520, 0.16, "sine", 0.04), 80);
  }

  private blip(freq: number, dur: number, type: OscillatorType, gain: number): void {
    try {
      if (!this.ctx) this.ctx = new AudioContext();
      const ctx = this.ctx;
      if (ctx.state === "suspended") void ctx.resume();
      const osc = ctx.createOscillator();
      const g = ctx.createGain();
      osc.type = type;
      osc.frequency.value = freq;
      g.gain.value = gain * this.master * this.sfx;
      g.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + dur);
      osc.connect(g);
      g.connect(ctx.destination);
      osc.start();
      osc.stop(ctx.currentTime + dur);
    } catch {
      /* audio opsional */
    }
  }
}
