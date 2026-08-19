export class Prediction {
  seq = 0;

  nextSeq(): number {
    this.seq += 1;
    return this.seq;
  }

  reset(): void {
    this.seq = 0;
  }
}
