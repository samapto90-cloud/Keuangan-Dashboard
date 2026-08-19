export class GameLoop {
  private last = 0;
  private raf = 0;
  private running = false;
  private freeze = 0;

  constructor(private readonly onFrame: (dt: number) => void) {}

  addHitStop(seconds: number): void {
    this.freeze = Math.max(this.freeze, seconds);
  }

  start(): void {
    if (this.running) return;
    this.running = true;
    this.last = performance.now();
    const tick = (now: number) => {
      if (!this.running) return;
      const real = Math.min(0.05, (now - this.last) / 1000);
      this.last = now;
      let scale = 1;
      if (this.freeze > 0) {
        scale = 0.12;
        this.freeze = Math.max(0, this.freeze - real);
      }
      this.onFrame(real * scale);
      this.raf = requestAnimationFrame(tick);
    };
    this.raf = requestAnimationFrame(tick);
  }

  stop(): void {
    this.running = false;
    cancelAnimationFrame(this.raf);
  }
}
