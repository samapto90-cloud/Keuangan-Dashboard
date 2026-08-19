export class SkillUI {
  readonly root: HTMLElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "skill-toast";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  flash(name: string): void {
    this.root.hidden = false;
    this.root.textContent = name;
    window.setTimeout(() => {
      this.root.hidden = true;
    }, 700);
  }
}
