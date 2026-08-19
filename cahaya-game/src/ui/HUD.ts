import { DEBUG_MODE, GAME_VERSION, PLAYER_NAME, type GraphicsPreset } from "../game/GameConfig";
import type { CombatView } from "../combat/CombatSystem";
import type { GameState } from "../game/GameState";
import type { Player } from "../player/Player";
import type { CameraController } from "../game/CameraController";
import type { NetStatus } from "../network/NetworkClient";
import { Camera, Vector3 } from "three";
import type { PlayerProgress, QuestView } from "../network/NetworkMessage";
import { objectiveLine } from "../quest/QuestStore";
import {
  formatMeters,
  heightGlyph,
  heightMark,
  mainPathSamples,
  markerScale,
} from "../world/Guidance";

export interface WorldHudView {
  zone: string;
  interact: string;
  progress: PlayerProgress | null;
  coin?: number;
  crystal?: number;
  battleToken?: number;
  guardianToken?: number;
  raidToken?: number;
  eduToken?: number;
  knowledge?: number;
  clock?: string;
  clockLabel?: string;
  navX?: number;
  navZ?: number;
  journey?: {
    title?: string;
    objective?: string;
    hint?: string;
    landmark?: string;
    cardinal?: string;
    subObjective?: string;
    optional?: string;
    hintJv?: string;
    hintId?: string;
  };
  optionalX?: number;
  optionalZ?: number;
  landmarks?: Array<{ x: number; z: number; name: string; discovered?: boolean }>;
  cameraYaw?: number;
  inCombat?: boolean;
  guidancePulse?: number;
  guidanceToast?: string;
  guidanceSub?: string;
  regionBanner?: string;
  followMain?: boolean;
  occluded?: boolean;
  npcs: Array<{ x: number; z: number; marker: string }>;
  objects: Array<{ x: number; z: number; kind: string }>;
  enemies: Array<{ x: number; z: number }>;
  formLabel?: string;
  transformReady?: string;
  transEnergy?: number;
  maxTransEnergy?: number;
  ultCharge?: number;
  dps?: number;
}

export interface NetView {
  status: NetStatus;
  playerId: string;
  online: number;
  channel: string;
  pingMs: number;
  messagesPerSec: number;
  serverPos: Vector3;
  interpOffset: number;
  reconOffset: number;
  others: Array<{ x: number; z: number; self: boolean }>;
}

export class HUD {
  onAdventure: (() => void) | null = null;
  onMounts: (() => void) | null = null;
  onMap: (() => void) | null = null;
  onJournal: (() => void) | null = null;
  onEndgame: (() => void) | null = null;
  onDungeon: (() => void) | null = null;
  onCraft: (() => void) | null = null;
  onLife: (() => void) | null = null;
  onHome: (() => void) | null = null;
  onPhoto: (() => void) | null = null;
  onFollowPath: (() => void) | null = null;
  private readonly host: HTMLElement;
  private readonly hp: HTMLElement | null;
  private readonly energy: HTMLElement | null;
  private readonly stamina: HTMLElement | null;
  private readonly debug: HTMLElement | null;
  private readonly pause: HTMLElement | null;
  private readonly hint: HTMLElement | null;
  private readonly targetBox: HTMLElement | null;
  private readonly targetName: HTMLElement | null;
  private readonly targetHp: HTMLElement | null;
  private readonly targetHpText: HTMLElement | null;
  private readonly connEl: HTMLElement | null;
  private readonly onlineEl: HTMLElement | null;
  private readonly nameEl: HTMLElement | null;
  private readonly worldPanel: HTMLElement | null;
  private readonly settingsPanel: HTMLElement | null;
  private readonly questPanel: HTMLElement | null;
  private readonly mapPanel: HTMLElement | null;
  private readonly tracker: HTMLElement | null;
  private readonly interactEl: HTMLElement | null;
  private readonly toastEl: HTMLElement | null;
  private readonly mini: HTMLCanvasElement | null;
  private readonly deathEl: HTMLElement | null;
  private readonly respawnCount: HTMLElement | null;
  private debugOn = DEBUG_MODE;
  vfxOn = true;
  shakeOn = true;
  numbersOn = true;
  graphics: GraphicsPreset = "AUTO";
  language: "jv" | "id" = "jv";
  textSize: "normal" | "large" = "normal";
  subtitleOn = true;
  masterVol = 1;
  musicVol = 0.6;
  sfxVol = 0.8;
  onAccessChange: ((vfx: boolean, shake: boolean, numbers: boolean) => void) | null = null;
  onGraphicsChange: ((preset: GraphicsPreset) => void) | null = null;
  onTextSize: ((size: "normal" | "large") => void) | null = null;
  onSubtitle: ((on: boolean) => void) | null = null;
  onAudioChange: ((master: number, music: number, sfx: number) => void) | null = null;
  onLanguage: ((lang: "jv" | "id") => void) | null = null;
  onDialogueSpeed: ((speed: "slow" | "normal" | "fast") => void) | null = null;
  onCollidersDebug: ((on: boolean) => void) | null = null;
  onCrowdDebug: ((on: boolean) => void) | null = null;
  private frames = 0;
  private fps = 0;
  private fpsMs = 0;
  private trackerHidden = false;
  private toastUntil = 0;

  constructor(root: HTMLElement) {
    root.innerHTML = `
      <div class="hud-card" id="hud-stats">
        <div class="hud-name" id="hud-name">NAMA: ${PLAYER_NAME}</div>
        <div class="hud-level">LEVEL: <span id="hud-level">1</span></div>
        <div class="hud-form">KELAS: <span id="hud-class">WARRIOR</span></div>
        <div class="hud-form">WUJUD: <span id="hud-form">NORMAL</span></div>
        <div class="hud-form">ASCEND: <span id="hud-tready">READY</span></div>
        <div class="hud-form">EXP: <span id="hud-exp">0</span></div>
        <div class="hud-form">GOLD: <span id="hud-coin">0</span></div>
        <div class="hud-form">BT: <span id="hud-battle">0</span> · GT: <span id="hud-guardian">0</span> · RT: <span id="hud-raid">0</span> · EDU: <span id="hud-edu">0</span></div>
        <div class="hud-form">CRYSTAL: <span id="hud-crystal">0</span></div>
        <div class="hud-form" id="hud-clock">09:00 · Morning</div>
        <div class="hud-form">KP: <span id="hud-kp">0</span></div>
        <div class="hud-online" id="hud-online">ONLINE: 0</div>
        <label>HP</label>
        <div class="hud-bar hud-hp"><i id="hud-hp"></i></div>
        <label>ENERGY</label>
        <div class="hud-bar hud-energy"><i id="hud-energy"></i></div>
        <label>ASCEND</label>
        <div class="hud-bar hud-trans"><i id="hud-trans"></i></div>
        <label>ULTIMATE <span id="hud-ult-pct">0%</span></label>
        <div class="hud-bar hud-ult"><i id="hud-ult"></i></div>
        <label>STAMINA</label>
        <div class="hud-bar hud-stamina"><i id="hud-stamina"></i></div>
      </div>
      <div class="hud-conn" id="hud-conn">MENGHUBUNGKAN</div>
      <div class="hud-net" id="hud-net"></div>
      <canvas id="hud-mini" class="hud-mini" width="128" height="128"></canvas>
      <canvas id="hud-compass" class="hud-compass" width="280" height="44"></canvas>
      <div id="hud-obj-chip" class="hud-obj-chip" hidden></div>
      <div id="hud-obj-float" class="hud-obj-float" hidden></div>
      <div id="hud-region" class="hud-region" hidden></div>
      <div id="hud-guide" class="hud-guide" hidden></div>
      <div class="hud-menu">
        <button type="button" id="hud-inv-btn">INV</button>
        <button type="button" id="hud-craft-btn">CRAFT</button>
        <button type="button" id="hud-life-btn">LIFE</button>
        <button type="button" id="hud-home-btn">HOME</button>
        <button type="button" id="hud-photo-btn">PHOTO</button>
        <button type="button" id="hud-char-btn">CHAR</button>
        <button type="button" id="hud-party-btn">PARTY</button>
        <button type="button" id="hud-social-btn">SOCIAL</button>
        <button type="button" id="hud-guild-btn">GUILD</button>
        <button type="button" id="hud-adv-btn">ADVENTURE</button>
        <button type="button" id="hud-mounts-btn">MOUNTS</button>
        <button type="button" id="hud-end-btn">ENDGAME</button>
        <button type="button" id="hud-dun-btn">DUNGEON</button>
        <button type="button" id="hud-quest-btn">QUEST</button>
        <button type="button" id="hud-map-btn">MAP</button>
        <button type="button" id="hud-world-btn">WORLD</button>
        <button type="button" id="hud-settings-btn">SETTINGS</button>
      </div>
      <div id="hud-tracker" class="hud-tracker">
        <button type="button" id="hud-tracker-hide" class="hud-tracker-hide">sembunyikan</button>
        <div class="hud-tracker-kicker">PERJALANAN</div>
        <strong id="hud-tracker-title">Awal Perjalanan</strong>
        <div id="hud-tracker-obj">Temui Elder Ardan, lalu ikuti jalan ke gerbang.</div>
        <div id="hud-tracker-sub"></div>
        <div id="hud-tracker-opt"></div>
        <div id="hud-tracker-prog"></div>
        <div id="hud-tracker-dist"></div>
        <button type="button" id="hud-follow" class="hud-follow">Ikuti Jalan Utama</button>
      </div>
      <div id="hud-interact" class="hud-interact" hidden>[INTERACT] E</div>
      <div id="hud-toast" class="hud-toast" hidden></div>
      <div id="hud-world-panel" class="hud-panel" hidden></div>
      <div id="hud-quest-panel" class="hud-panel hud-journal" hidden></div>
      <div id="hud-map-panel" class="hud-panel hud-map" hidden></div>
      <div id="hud-settings-panel" class="hud-panel" hidden>
        <label class="hud-check"><input type="checkbox" id="hud-debug-toggle" ${DEBUG_MODE ? "checked" : ""}/> Debug panel</label>
        ${DEBUG_MODE ? `<label class="hud-check"><input type="checkbox" id="hud-colliders-toggle"/> Show colliders</label>
        <label class="hud-check"><input type="checkbox" id="hud-crowd-toggle"/> Show NPC radius</label>` : ""}
        <label class="hud-check"><input type="checkbox" id="hud-vfx-toggle" checked/> Boss VFX / telegraph</label>
        <label class="hud-check"><input type="checkbox" id="hud-shake-toggle" checked/> Camera shake</label>
        <label class="hud-check"><input type="checkbox" id="hud-dmg-toggle" checked/> Damage numbers</label>
        <label class="hud-check"><input type="checkbox" id="hud-sub-toggle" checked/> Subtitle / terjemahan NPC</label>
        <p class="hud-hint">Ukuran teks</p>
        <div class="hud-gfx">
          <button type="button" data-text="normal">NORMAL</button>
          <button type="button" data-text="large">BESAR</button>
        </div>
        <p class="hud-hint">Audio</p>
        <label class="hud-check">Master <input type="range" id="hud-vol-master" min="0" max="100" value="100"/></label>
        <label class="hud-check">Musik <input type="range" id="hud-vol-music" min="0" max="100" value="60"/></label>
        <label class="hud-check">SFX <input type="range" id="hud-vol-sfx" min="0" max="100" value="80"/></label>
        <p class="hud-hint">Graphics</p>
        <div class="hud-gfx">
          <button type="button" data-gfx="AUTO">AUTO</button>
          <button type="button" data-gfx="LOW">LOW</button>
          <button type="button" data-gfx="MEDIUM">MEDIUM</button>
          <button type="button" data-gfx="HIGH">HIGH</button>
          <button type="button" data-gfx="ULTRA">ULTRA</button>
        </div>
        <p class="hud-hint">Versi ${GAME_VERSION}</p>
        <p class="hud-hint">Language</p>
        <div class="hud-gfx">
          <button type="button" data-lang="jv">JAWA</button>
          <button type="button" data-lang="id">INDONESIA</button>
        </div>
        <p class="hud-hint">Dialogue</p>
        <div class="hud-gfx">
          <button type="button" data-dlg="slow">SLOW</button>
          <button type="button" data-dlg="normal">NORMAL</button>
          <button type="button" data-dlg="fast">FAST</button>
        </div>
        <p class="hud-hint">O Social · G Guild · P Party · V PvP · H Endgame · Enter Chat</p>
      </div>
      <div id="hud-target" class="hud-target" hidden>
        <div>TARGET: <span id="hud-target-name">-</span></div>
        <div class="hud-bar hud-hp"><i id="hud-target-hp"></i></div>
        <div id="hud-target-hp-text">HP 0 / 100</div>
      </div>
      <div id="hud-death" class="hud-death" hidden>
        <strong id="hud-death-title">YOU ARE DEFEATED</strong>
        <span id="hud-respawn-label">RESPAWNING...</span>
        <span id="hud-respawn-count">5</span>
      </div>
      <div id="hud-debug" class="hud-debug" hidden></div>
      <div id="hud-pause" class="hud-pause" hidden><strong>Jeda</strong><span>Tekan ESC atau tap lanjut</span></div>
      <div id="hint" class="hud-hint">WASD gerak · 1-4 skill · R ultimate · Shift block · Tab target · T transform · O social · G guild · P party · V pvp · H endgame · Enter chat · M mount · N journal · B peta · E bicara</div>
      <button type="button" id="photo-exit" class="photo-exit" hidden>EXIT PHOTO</button>
    `;
    this.host = root;
    this.hp = root.querySelector("#hud-hp");
    this.energy = root.querySelector("#hud-energy");
    this.stamina = root.querySelector("#hud-stamina");
    this.debug = root.querySelector("#hud-debug");
    this.pause = root.querySelector("#hud-pause");
    this.hint = root.querySelector("#hint");
    this.targetBox = root.querySelector("#hud-target");
    this.targetName = root.querySelector("#hud-target-name");
    this.targetHp = root.querySelector("#hud-target-hp");
    this.targetHpText = root.querySelector("#hud-target-hp-text");
    this.connEl = root.querySelector("#hud-conn");
    this.onlineEl = root.querySelector("#hud-online");
    this.nameEl = root.querySelector("#hud-name");
    this.worldPanel = root.querySelector("#hud-world-panel");
    this.settingsPanel = root.querySelector("#hud-settings-panel");
    this.questPanel = root.querySelector("#hud-quest-panel");
    this.mapPanel = root.querySelector("#hud-map-panel");
    this.tracker = root.querySelector("#hud-tracker");
    this.interactEl = root.querySelector("#hud-interact");
    this.toastEl = root.querySelector("#hud-toast");
    this.mini = root.querySelector("#hud-mini");
    this.deathEl = root.querySelector("#hud-death");
    this.respawnCount = root.querySelector("#hud-respawn-count");
    const hidePanels = (): void => {
      if (this.worldPanel) this.worldPanel.hidden = true;
      if (this.settingsPanel) this.settingsPanel.hidden = true;
      if (this.questPanel) this.questPanel.hidden = true;
      if (this.mapPanel) this.mapPanel.hidden = true;
    };
    root.querySelector("#hud-world-btn")?.addEventListener("click", () => {
      const on = this.worldPanel?.hidden !== false;
      hidePanels();
      if (this.worldPanel) this.worldPanel.hidden = !on;
    });
    root.querySelector("#hud-settings-btn")?.addEventListener("click", () => {
      const on = this.settingsPanel?.hidden !== false;
      hidePanels();
      if (this.settingsPanel) this.settingsPanel.hidden = !on;
    });
    root.querySelector("#hud-adv-btn")?.addEventListener("click", () => this.onAdventure?.());
    root.querySelector("#hud-mounts-btn")?.addEventListener("click", () => this.onMounts?.());
    root.querySelector("#hud-end-btn")?.addEventListener("click", () => this.onEndgame?.());
    root.querySelector("#hud-craft-btn")?.addEventListener("click", () => this.onCraft?.());
    root.querySelector("#hud-life-btn")?.addEventListener("click", () => this.onLife?.());
    root.querySelector("#hud-home-btn")?.addEventListener("click", () => this.onHome?.());
    root.querySelector("#hud-photo-btn")?.addEventListener("click", () => this.onPhoto?.());
    root.querySelector("#photo-exit")?.addEventListener("click", () => this.onPhoto?.());
    root.querySelector("#hud-dun-btn")?.addEventListener("click", () => this.onDungeon?.());
    root.querySelector("#hud-quest-btn")?.addEventListener("click", () => {
      if (this.onJournal) {
        this.onJournal();
        return;
      }
      const on = this.questPanel?.hidden !== false;
      hidePanels();
      if (this.questPanel) this.questPanel.hidden = !on;
    });
    root.querySelector("#hud-map-btn")?.addEventListener("click", () => {
      if (this.onMap) {
        this.onMap();
        return;
      }
      const on = this.mapPanel?.hidden !== false;
      hidePanels();
      if (this.mapPanel) this.mapPanel.hidden = !on;
    });
    root.querySelector("#hud-tracker-hide")?.addEventListener("click", () => {
      this.trackerHidden = !this.trackerHidden;
    });
    root.querySelector("#hud-follow")?.addEventListener("click", () => this.onFollowPath?.());
    root.querySelector("#hud-debug-toggle")?.addEventListener("change", (e) => {
      this.debugOn = (e.target as HTMLInputElement).checked;
    });
    root.querySelector("#hud-colliders-toggle")?.addEventListener("change", (e) => {
      this.onCollidersDebug?.((e.target as HTMLInputElement).checked);
    });
    root.querySelector("#hud-crowd-toggle")?.addEventListener("change", (e) => {
      this.onCrowdDebug?.((e.target as HTMLInputElement).checked);
    });
    const bindAccess = (id: string, apply: (on: boolean) => void): void => {
      root.querySelector(id)?.addEventListener("change", (e) => {
        apply((e.target as HTMLInputElement).checked);
        this.onAccessChange?.(this.vfxOn, this.shakeOn, this.numbersOn);
      });
    };
    bindAccess("#hud-vfx-toggle", (on) => {
      this.vfxOn = on;
    });
    bindAccess("#hud-shake-toggle", (on) => {
      this.shakeOn = on;
    });
    bindAccess("#hud-dmg-toggle", (on) => {
      this.numbersOn = on;
    });
    root.querySelectorAll("[data-gfx]").forEach((el) => {
      el.addEventListener("click", () => {
        const v = (el as HTMLElement).dataset.gfx;
        if (v === "AUTO" || v === "LOW" || v === "MEDIUM" || v === "HIGH" || v === "ULTRA") {
          this.graphics = v;
          this.onGraphicsChange?.(v);
        }
      });
    });
    root.querySelectorAll("[data-lang]").forEach((el) => {
      el.addEventListener("click", () => {
        const v = (el as HTMLElement).dataset.lang;
        if (v === "jv" || v === "id") {
          this.language = v;
          this.onLanguage?.(v);
        }
      });
    });
    root.querySelectorAll("[data-dlg]").forEach((el) => {
      el.addEventListener("click", () => {
        const v = (el as HTMLElement).dataset.dlg;
        if (v === "slow" || v === "normal" || v === "fast") this.onDialogueSpeed?.(v);
      });
    });
    root.querySelectorAll("[data-text]").forEach((el) => {
      el.addEventListener("click", () => {
        const v = (el as HTMLElement).dataset.text;
        if (v === "normal" || v === "large") {
          this.textSize = v;
          this.onTextSize?.(v);
        }
      });
    });
    root.querySelector("#hud-sub-toggle")?.addEventListener("change", (e) => {
      this.subtitleOn = (e.target as HTMLInputElement).checked;
      this.onSubtitle?.(this.subtitleOn);
    });
    const bindVol = (id: string, apply: (n: number) => void): void => {
      root.querySelector(id)?.addEventListener("input", (e) => {
        apply(Number((e.target as HTMLInputElement).value) / 100);
        this.onAudioChange?.(this.masterVol, this.musicVol, this.sfxVol);
      });
    };
    bindVol("#hud-vol-master", (n) => {
      this.masterVol = n;
    });
    bindVol("#hud-vol-music", (n) => {
      this.musicVol = n;
    });
    bindVol("#hud-vol-sfx", (n) => {
      this.sfxVol = n;
    });
  }

  toast(text: string): void {
    if (!this.toastEl) return;
    this.toastEl.hidden = false;
    this.toastEl.textContent = text;
    this.toastUntil = performance.now() + 2800;
  }

  setMobile(on: boolean): void {
    if (this.hint) this.hint.hidden = on;
  }

  update(player: Player, camera: CameraController, state: GameState, dt: number, combat?: CombatView, net?: NetView, world?: WorldHudView): void {
    this.frames += 1;
    this.fpsMs += dt * 1000;
    if (this.fpsMs >= 500) {
      this.fps = Math.round(this.frames / (this.fpsMs / 1000));
      this.frames = 0;
      this.fpsMs = 0;
    }
    const s = player.stats;
    if (this.hp) this.hp.style.width = `${s.hpRatio() * 100}%`;
    if (this.energy) this.energy.style.width = `${s.energyRatio() * 100}%`;
    const trans = document.getElementById("hud-trans");
    if (trans) {
      const maxT = world?.maxTransEnergy || 1;
      trans.style.width = `${Math.min(100, ((world?.transEnergy ?? 0) / maxT) * 100)}%`;
    }
    const ult = document.getElementById("hud-ult");
    const ultPct = document.getElementById("hud-ult-pct");
    const charge = world?.ultCharge ?? 0;
    if (ult) ult.style.width = `${charge}%`;
    if (ultPct) ultPct.textContent = `${charge}%`;
    if (this.stamina) this.stamina.style.width = `${s.staminaRatio() * 100}%`;
    if (this.nameEl) this.nameEl.textContent = `NAMA: ${s.name}`;
    const level = document.getElementById("hud-level");
    if (level) level.textContent = String(s.level);
    const klass = document.getElementById("hud-class");
    if (klass) klass.textContent = player.className;
    const form = document.getElementById("hud-form");
    if (form) form.textContent = world?.formLabel || player.formName();
    const ready = document.getElementById("hud-tready");
    if (ready) ready.textContent = world?.transformReady || "READY";
    const exp = document.getElementById("hud-exp");
    if (exp) exp.textContent = `${s.exp} / ${s.expToNext}`;
    if (this.pause) this.pause.hidden = !state.paused;
    const online = net?.online ?? 0;
    if (this.onlineEl) this.onlineEl.textContent = `ONLINE: ${online}`;
    if (this.connEl) {
      const ping = Math.round(net?.pingMs ?? 0);
      let pingStatus = "OK";
      if (ping < 80) pingStatus = "LOW";
      else if (ping < 140) pingStatus = "OK";
      else pingStatus = "HIGH";
      this.connEl.textContent = `${statusLabel(net?.status ?? "offline")} · PING ${ping} ms (${pingStatus})`;
      this.connEl.dataset.status = net?.status ?? "offline";
      this.connEl.dataset.ping = pingStatus;
    }
    if (this.worldPanel && !this.worldPanel.hidden) {
      this.worldPanel.textContent = `World of Dawn\n${world?.zone || "Dawn Village"}\n${world?.clock || "09:00"} ${world?.clockLabel || "Morning"}\nworld-01 · ${net?.channel || "channel-01"}\nONLINE: ${online}`;
    }
    const clock = document.getElementById("hud-clock");
    if (clock) clock.textContent = `${world?.clock || "09:00"} · ${world?.clockLabel || "Morning"}`;
    const kp = document.getElementById("hud-kp");
    if (kp) kp.textContent = fmtMoney(world?.knowledge ?? world?.progress?.knowledgePoints ?? 0);
    const coin = document.getElementById("hud-coin");
    if (coin) coin.textContent = fmtMoney(world?.coin ?? world?.progress?.coin ?? 0);
    const crystal = document.getElementById("hud-crystal");
    if (crystal) crystal.textContent = fmtMoney(world?.crystal ?? world?.progress?.crystal ?? 0);
    const bt = document.getElementById("hud-battle");
    if (bt) bt.textContent = fmtMoney(world?.battleToken ?? 0);
    const gt = document.getElementById("hud-guardian");
    if (gt) gt.textContent = fmtMoney(world?.guardianToken ?? 0);
    const rt = document.getElementById("hud-raid");
    if (rt) rt.textContent = fmtMoney(world?.raidToken ?? 0);
    const edu = document.getElementById("hud-edu");
    if (edu) edu.textContent = fmtMoney(world?.eduToken ?? 0);
    this.renderTracker(player, world, camera);
    this.renderJournal(world);
    this.renderMap(player, world);
    if (this.interactEl) {
      this.interactEl.hidden = !world?.interact;
      if (world?.interact) this.interactEl.textContent = world.interact;
    }
    if (this.toastEl && performance.now() > this.toastUntil) this.toastEl.hidden = true;
    if (this.targetBox) {
      const t = combat?.target ?? null;
      this.targetBox.hidden = !t;
      if (t) {
        if (this.targetName) this.targetName.textContent = `${t.name}  LV.${t.level}`;
        const pct = t.health.getHealthPercent() * 100;
        if (this.targetHp) this.targetHp.style.width = `${pct}%`;
        if (this.targetHpText) this.targetHpText.textContent = `HP ${Math.ceil(t.health.hp)} / ${t.health.maxHp}`;
      }
    }
    if (this.deathEl) {
      const downed = player.combatState === "DOWNED";
      const dead = player.combatState === "DEAD" || (!downed && player.health.isDead());
      this.deathEl.hidden = !dead && !downed;
      const title = document.getElementById("hud-death-title");
      const label = document.getElementById("hud-respawn-label");
      if (downed) {
        if (title) title.textContent = "DOWNED";
        if (label) label.textContent = "Menunggu revive party...";
        if (this.respawnCount) this.respawnCount.textContent = "HOLD";
      } else if (dead) {
        if (title) title.textContent = "YOU ARE DEFEATED";
        if (label) label.textContent = "RESPAWNING...";
        if (this.respawnCount) {
          const left = combat?.respawnAt ? Math.max(0, Math.ceil((combat.respawnAt - Date.now()) / 1000)) : 0;
          this.respawnCount.textContent = left > 0 ? String(left) : "RESPAWN";
        }
      }
    }
    this.drawMini(player, net, world);
    const showDebug = this.debugOn;
    if (this.debug) {
      this.debug.hidden = !showDebug;
      if (showDebug) {
        const sp = net?.serverPos;
        this.debug.textContent =
          `Ping ${Math.round(net?.pingMs ?? 0)} ms\n` +
          `Players Online ${online}\n` +
          `Client FPS ${this.fps}\n` +
          `Network ${net?.messagesPerSec ?? 0} msg/s\n` +
          `Client Pos ${player.position.x.toFixed(2)}, ${player.position.y.toFixed(2)}, ${player.position.z.toFixed(2)}\n` +
          `Server Pos ${sp ? `${sp.x.toFixed(2)}, ${sp.y.toFixed(2)}, ${sp.z.toFixed(2)}` : "-"}\n` +
          `Interp Offset ${(net?.interpOffset ?? 0).toFixed(3)}\n` +
          `Reconcile ${(net?.reconOffset ?? 0).toFixed(3)}\n` +
          `Connection ${statusLabel(net?.status ?? "offline")}\n` +
          `${player.animState}`;
      }
    }
    void camera;
  }

  private renderTracker(player: Player, world?: WorldHudView, camera?: CameraController): void {
    if (!this.tracker) return;
    this.tracker.hidden = this.trackerHidden;
    const q = trackedQuest(world?.progress);
    const title = document.getElementById("hud-tracker-title");
    const obj = document.getElementById("hud-tracker-obj");
    const sub = document.getElementById("hud-tracker-sub");
    const opt = document.getElementById("hud-tracker-opt");
    const prog = document.getElementById("hud-tracker-prog");
    const distEl = document.getElementById("hud-tracker-dist");
    const follow = document.getElementById("hud-follow");
    const chip = document.getElementById("hud-obj-chip");
    const floatEl = document.getElementById("hud-obj-float");
    const region = document.getElementById("hud-region");
    const guide = document.getElementById("hud-guide");
    const journey = world?.journey;
    if (journey?.objective) {
      if (title) title.textContent = journey.title || "PERJALANAN";
      if (obj) obj.textContent = journey.objective;
      if (sub) sub.textContent = journey.subObjective || "";
      if (opt) opt.textContent = journey.optional ? `○ ${journey.optional}` : "";
      if (prog) prog.textContent = journey.hint || "";
    } else if (!q) {
      if (title) title.textContent = "PERJALANAN";
      if (obj) obj.textContent = "Temui Elder Ardan, lalu ikuti jalan ke gerbang.";
      if (sub) sub.textContent = "";
      if (opt) opt.textContent = "";
      if (prog) prog.textContent = "";
      if (distEl) distEl.textContent = "";
    } else {
      if (title) title.textContent = q.title;
      const current = q.objectives.find((o) => o.progress < o.count) ?? q.objectives[0];
      if (obj) obj.textContent = q.state === "COMPLETED" ? "Lanjutkan ke tujuan berikutnya." : (current?.text ?? q.description);
      if (prog) prog.textContent = current && current.count > 1 ? `${current.progress} / ${current.count}` : "";
    }
    if (follow) follow.textContent = world?.followMain ? "Berhenti mengikuti" : "Ikuti Jalan Utama";
    if (region) {
      region.hidden = !world?.regionBanner;
      if (world?.regionBanner) region.innerHTML = `<strong>${world.regionBanner}</strong><span>${journey?.objective || ""}</span>`;
    }
    if (guide) {
      const show = !!(world?.guidanceToast);
      guide.hidden = !show;
      if (show) guide.innerHTML = `<p>${world?.guidanceToast}</p>${world?.guidanceSub ? `<small>${world.guidanceSub}</small>` : ""}`;
    }
    const navOn = world?.navX != null && world.navZ != null;
    if (!navOn) {
      if (distEl) distEl.textContent = "";
      if (chip) chip.hidden = true;
      if (floatEl) floatEl.hidden = true;
      this.drawCompass(0, false, 0, world?.cameraYaw ?? 0, world?.guidancePulse ?? 0);
      return;
    }
    const dx = (world?.navX ?? 0) - player.position.x;
    const dz = (world?.navZ ?? 0) - player.position.z;
    const dist = Math.hypot(dx, dz);
    const mark = heightMark(0 - player.position.y);
    const glyph = world?.occluded ? "▲" : heightGlyph(mark);
    const label = journey?.landmark || "Tujuan";
    if (distEl) distEl.textContent = `${label}  ${formatMeters(dist)}`;
    if (chip) {
      chip.hidden = dist < 1.1;
      chip.classList.toggle("pulse", (world?.guidancePulse ?? 0) > 0.35);
      chip.textContent = `${glyph} ${label}  ${formatMeters(dist)}`;
    }
    this.drawCompass(Math.atan2(dx, dz), dist >= 1.1, dist, world?.cameraYaw ?? 0, world?.guidancePulse ?? 0);
    this.placeFloat(floatEl, camera, player, world?.navX ?? 0, world?.navZ ?? 0, dist, glyph, world?.occluded ?? false, label);
  }

  private drawCompass(objAng: number, showObj: boolean, dist: number, yaw: number, pulse: number): void {
    const el = document.getElementById("hud-compass") as HTMLCanvasElement | null;
    if (!el) return;
    const ctx = el.getContext("2d");
    if (!ctx) return;
    const w = el.width;
    const h = el.height;
    ctx.clearRect(0, 0, w, h);
    ctx.fillStyle = "rgba(7,17,31,0.45)";
    ctx.fillRect(0, 0, w, h);
    const labels = ["N", "NE", "E", "SE", "S", "SW", "W", "NW"];
    const heading = yaw;
    ctx.font = "600 11px sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    for (let i = 0; i < 8; i++) {
      const ang = (i * Math.PI) / 4 - heading;
      const x = w / 2 + Math.sin(ang) * (w * 0.38);
      ctx.fillStyle = i === 0 ? "#fde68a" : "rgba(244,248,252,0.7)";
      ctx.fillText(labels[i], x, h / 2);
    }
    ctx.fillStyle = "#e8b86d";
    ctx.beginPath();
    ctx.moveTo(w / 2, 6);
    ctx.lineTo(w / 2 - 5, 16);
    ctx.lineTo(w / 2 + 5, 16);
    ctx.closePath();
    ctx.fill();
    if (showObj) {
      const ang = objAng - heading;
      const x = w / 2 + Math.sin(ang) * (w * 0.38);
      const r = 3 + (pulse > 0.3 ? Math.sin(performance.now() / 180) * 1.2 : 0);
      ctx.fillStyle = dist < 40 ? "#fde68a" : "rgba(253,230,138,0.75)";
      ctx.beginPath();
      ctx.arc(x, h / 2 + 10, r, 0, Math.PI * 2);
      ctx.fill();
    }
  }

  private placeFloat(
    el: HTMLElement | null,
    camera: CameraController | undefined,
    player: Player,
    navX: number,
    navZ: number,
    dist: number,
    glyph: string,
    occluded: boolean,
    label: string,
  ): void {
    if (!el || !camera || dist < 1.2) {
      if (el) el.hidden = true;
      return;
    }
    const cam = camera.camera;
    const ndc = worldToNdc(cam, navX, 1.6, navZ);
    el.hidden = false;
    const scale = markerScale(dist);
    el.style.transform = `translate(-50%, -50%) scale(${scale})`;
    el.textContent = occluded || !ndc.inFront ? `${glyph} ${formatMeters(dist)}` : `${glyph}`;
    el.title = label;
    if (!ndc.inFront || occluded) {
      const ang = Math.atan2(navX - player.position.x, navZ - player.position.z) - camera.yaw;
      el.style.left = `${50 + Math.sin(ang) * 28}%`;
      el.style.top = "22%";
      el.classList.add("edge");
    } else {
      el.style.left = `${(ndc.x * 0.5 + 0.5) * 100}%`;
      el.style.top = `${(-ndc.y * 0.5 + 0.5) * 100}%`;
      el.classList.remove("edge");
    }
  }

  private renderJournal(world?: WorldHudView): void {
    if (!this.questPanel || this.questPanel.hidden) return;
    const quests = world?.progress?.quests ?? [];
    const block = (kind: string, label: string): string => {
      const rows = quests.filter((q) => (kind === "COMPLETED" ? q.state === "CLAIMED" : q.kind === kind && q.state !== "CLAIMED"));
      if (!rows.length) return `<h4>${label}</h4><p>Tidak ada.</p>`;
      return `<h4>${label}</h4>` + rows.map((q) => journalRow(q)).join("");
    };
    this.questPanel.innerHTML =
      block("main", "JOURNEY") +
      block("side", "JEJAK SAMPING") +
      block("education", "UJIAN PERJALANAN") +
      block("COMPLETED", "COMPLETED");
  }

  private renderMap(player: Player, world?: WorldHudView): void {
    if (!this.mapPanel || this.mapPanel.hidden) return;
    const forest = world?.progress?.forestUnlocked ? "Terbuka" : "Terkunci";
    this.mapPanel.innerHTML = `
      <h4>PETA DUNIA</h4>
      <div>Village of Dawn — hub awal</div>
      <div>Whisper Forest — ${forest}</div>
      <div>Posisi: ${player.position.x.toFixed(1)}, ${player.position.z.toFixed(1)}</div>
      <div>Zona: ${world?.zone ?? "Village of Dawn"}</div>
      <div class="map-legend">Kuning: quest · Hijau: NPC · Merah: musuh · Putih: gerbang</div>`;
  }

  private drawMini(player: Player, net?: NetView, world?: WorldHudView): void {
    const c = this.mini;
    if (!c) return;
    const ctx = c.getContext("2d");
    if (!ctx) return;
    const w = c.width;
    const h = c.height;
    ctx.clearRect(0, 0, w, h);
    ctx.fillStyle = "rgba(7,17,31,0.55)";
    ctx.beginPath();
    ctx.arc(w / 2, h / 2, w / 2 - 2, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "rgba(255,255,255,0.25)";
    ctx.stroke();
    ctx.save();
    ctx.beginPath();
    ctx.arc(w / 2, h / 2, w / 2 - 2, 0, Math.PI * 2);
    ctx.clip();
    const range = world?.inCombat ? 42 : 32;
    const scale = (w - 16) / (range * 2);
    const px = player.position.x;
    const pz = player.position.z;
    const yaw = world?.cameraYaw ?? 0;
    ctx.translate(w / 2, h / 2);
    ctx.rotate(-yaw);
    const to = (x: number, z: number): { x: number; y: number } => ({ x: (x - px) * scale, y: -(z - pz) * scale });
    const plot = (x: number, z: number, color: string, r: number): void => {
      const dx = x - px;
      const dz = z - pz;
      if (dx * dx + dz * dz > range * range) return;
      const p = to(x, z);
      ctx.fillStyle = color;
      ctx.beginPath();
      ctx.arc(p.x, p.y, r, 0, Math.PI * 2);
      ctx.fill();
    };
    ctx.strokeStyle = "rgba(232,184,109,0.45)";
    ctx.lineWidth = 2;
    ctx.beginPath();
    let started = false;
    for (const s of mainPathSamples(pz - range, pz + range, 6)) {
      const p = to(s.x, s.z);
      if (!started) {
        ctx.moveTo(p.x, p.y);
        started = true;
      } else ctx.lineTo(p.x, p.y);
    }
    ctx.stroke();
    for (const lm of world?.landmarks ?? []) plot(lm.x, lm.z, "#c4b49a", 2.5);
    for (const obj of world?.objects ?? []) {
      const color = obj.kind === "gate" ? "#f4f4f5" : obj.kind === "checkpoint" ? "#fde68a" : obj.kind === "crystal" ? "#67e8f9" : obj.kind === "drop" ? "#fbbf24" : "#d6d3d1";
      plot(obj.x, obj.z, color, 2);
    }
    for (const npc of world?.npcs ?? []) plot(npc.x, npc.z, npc.marker ? "#fbbf24" : "#86efac", npc.marker ? 3 : 2);
    for (const enemy of world?.enemies ?? []) plot(enemy.x, enemy.z, "#f87171", 2);
    for (const other of net?.others ?? []) plot(other.x, other.z, "#7dd3fc", 3);
    if (world?.optionalX != null && world.optionalZ != null) plot(world.optionalX, world.optionalZ, "#86efac", 2);
    if (world?.navX != null && world?.navZ != null) plot(world.navX, world.navZ, "#fbbf24", 3.4);
    const npos = to(px, pz + Math.min(10, range * 0.42));
    ctx.fillStyle = "#fde68a";
    ctx.font = "700 9px sans-serif";
    ctx.textAlign = "center";
    ctx.fillText("N", npos.x, npos.y);
    ctx.fillStyle = "#e8b86d";
    ctx.beginPath();
    ctx.moveTo(0, -6);
    ctx.lineTo(4, 5);
    ctx.lineTo(-4, 5);
    ctx.closePath();
    ctx.fill();
    ctx.restore();
  }

  setPhotoMode(on: boolean): void {
    this.host.classList.toggle("photo-hidden", on);
    const exit = this.host.querySelector("#photo-exit") as HTMLElement | null;
    if (exit) exit.hidden = !on;
  }
}

function trackedQuest(progress: PlayerProgress | null | undefined): QuestView | null {
  return progress?.quests.find((q) => q.state === "ACTIVE" || q.state === "COMPLETED") ?? null;
}

function journalRow(q: QuestView): string {
  return `<div class="journal-row"><b>${q.title}</b><span>${q.state}</span><p>${q.description}</p><p>${objectiveLine(q)}</p><p>${q.npcName} · ${q.location}</p><p>+${q.rewards.exp} EXP · ${q.rewards.coin} koin</p></div>`;
}

function worldToNdc(camera: Camera, x: number, y: number, z: number): { x: number; y: number; inFront: boolean } {
  const v = _ndc.set(x, y, z);
  v.project(camera);
  return { x: v.x, y: v.y, inFront: v.z < 1 };
}

const _ndc = new Vector3();

function fmtMoney(n: number): string {
  return n.toLocaleString("en-US");
}

function statusLabel(status: NetStatus): string {
  if (status === "online") return "CONNECTED";
  if (status === "reconnecting" || status === "connecting") return "RECONNECTING";
  return "DISCONNECTED";
}
