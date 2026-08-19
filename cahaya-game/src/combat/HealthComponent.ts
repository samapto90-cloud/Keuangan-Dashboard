export class HealthComponent {
  hp: number;

  constructor(public maxHp: number, hp = maxHp) {
    this.hp = hp;
  }

  takeDamage(amount: number): number {
    if (amount <= 0 || this.hp <= 0) return 0;
    const dealt = Math.min(this.hp, amount);
    this.hp -= dealt;
    return dealt;
  }

  heal(amount: number): void {
    if (amount <= 0 || this.hp <= 0) return;
    this.hp = Math.min(this.maxHp, this.hp + amount);
  }

  isDead(): boolean {
    return this.hp <= 0;
  }

  getHealthPercent(): number {
    return this.maxHp <= 0 ? 0 : this.hp / this.maxHp;
  }
}
