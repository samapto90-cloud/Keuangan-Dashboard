import * as THREE from "three";
import { GameLoop } from "./GameLoop";
import { GameState } from "./GameState";
import { InputManager } from "./InputManager";
import { CameraController } from "./CameraController";
import type { Player } from "../player/Player";
import { PlayerController } from "../player/PlayerController";
import { PlayerAnimationController } from "../player/PlayerAnimationController";
import { World } from "../world/World";
import { GuidanceSystem } from "../world/Guidance";
import { FishingUI } from "../world/FishingUI";
import { CollisionWorld } from "../world/Collision";
import { CollisionDebug } from "../world/CollisionDebug";
import {
  VILLAGE_AABBS,
  VILLAGE_CIRCLES,
  VILLAGE_ROCKS,
  VILLAGE_TREES,
} from "../world/WorldColliders";
import { HUD } from "../ui/HUD";
import { MobileControls } from "../ui/MobileControls";
import { CombatSystem } from "../combat/CombatSystem";
import { EnemyStore } from "../combat/EnemyStore";
import type { MoveMode } from "../player/PlayerController";
import { PlayerFactory } from "../network/PlayerFactory";
import { NetworkClient } from "../network/NetworkClient";
import { RemotePlayerStore } from "../network/RemotePlayerStore";
import { WorldSync } from "../network/WorldSync";
import { PlayerSync } from "../network/PlayerSync";
import { Reconciliation } from "../network/Reconciliation";
import { QuestStore } from "../quest/QuestStore";
import { DialogueUI } from "../dialogue/DialogueUI";
import { EducationUI } from "../education/EducationUI";
import type {
  EducationFeedback,
  InspectOut,
  InventoryUpdated,
  InteractResult,
  NetMsgType,
  PartyView,
  ChatOut,
  FinderListing,
  GuildView,
  ShopCatalogOut,
  SearchHit,
  TradeView,
  CraftingState,
  FishState,
  PlayerProgress,
  QuestionOut,
  SocialNote,
  SocialState,
  StatsView,
  HouseState,
  LifeState,
  MarketState,
  PlayerCardView,
  BossAOEOut,
  BossPhaseOut,
  BossTelegraphOut,
  ChapterListOut,
  ChapterView,
  DungeonListOut,
  DungeonLoading,
  DungeonOffer,
  DungeonReadyOut,
  DungeonView,
  QueueView,
  LootResult,
  PvpLobbyOut,
  PvpQueueView,
  PvpReadyOut,
  PvpResultOut,
  PvpView,
  LBEntry,
} from "../network/NetworkMessage";
import { LoadoutStore } from "../inventory/LoadoutStore";
import { InventoryUI } from "../inventory/InventoryUI";
import { CraftingUI } from "../inventory/CraftingUI";
import { LifeUI } from "../progression/LifeUI";
import { PetFollower } from "../player/PetFollower";
import { EquipmentVisual } from "../inventory/EquipmentVisual";
import { CharacterProgressionUI } from "../progression/CharacterProgressionUI";
import { SkillTreeUI } from "../progression/SkillTreeUI";
import { TransformationWheel } from "../progression/TransformationWheel";
import { TransformationUI } from "../progression/TransformationUI";
import { TransformationFX } from "../progression/TransformationFX";
import { ProgressionStore, visualForm, type ProgressionView, type TransformView } from "../progression/ProgressionStore";
import { SkillUI } from "../progression/SkillUI";
import { BuildUI } from "../progression/BuildUI";
import { PartyUI } from "../party/PartyUI";
import { PartyInviteUI } from "../party/PartyInviteUI";
import { SocialUI } from "../social/SocialUI";
import { ChatUI } from "../social/ChatUI";
import { GuildUI } from "../social/GuildUI";
import { TradeUI } from "../social/TradeUI";
import { ShopUI } from "../social/ShopUI";
import { NotifyCenter } from "../social/NotifyCenter";
import { DungeonUI } from "../dungeon/DungeonUI";
import { PvpUI, PvpHud } from "../pvp/PvpUI";
import { PvpTarget } from "../pvp/PvpTarget";
import { BossUI } from "../dungeon/BossUI";
import { ChapterUI } from "../dungeon/ChapterUI";
import { DiscoverySystem } from "../world/DiscoverySystem";
import { MountSystem } from "../world/MountSystem";
import { MountCollection } from "../world/MountCollection";
import { FastTravel } from "../world/FastTravel";
import { WorldEventUI } from "../world/WorldEventUI";
import { WorldMap } from "../world/WorldMap";
import { AdventureJournal } from "../world/AdventureJournal";
import { CinematicUI } from "../story/CinematicUI";
import { LoreJournal } from "../world/LoreJournal";
import { MusicManager } from "../world/MusicManager";
import { EndgameUI } from "../endgame/EndgameUI";
import type { WorldJournal, WorldEventView, WorldBossView, LoreView, NotifyView, EndgameState } from "../network/NetworkMessage";
import { GRAPHICS, resolveGraphics, type GraphicsPreset } from "./GameConfig";

export class Game {
  readonly state = new GameState();
  private readonly renderer: THREE.WebGLRenderer;
  private readonly camera: THREE.PerspectiveCamera;
  private readonly world: World;
  private readonly player: Player;
  private readonly input: InputManager;
  private readonly controller: PlayerController;
  private readonly animation: PlayerAnimationController;
  private readonly cameraRig: CameraController;
  private readonly collision: CollisionWorld;
  private readonly collisionDebug: CollisionDebug;
  readonly hud: HUD;
  private readonly mobile: MobileControls;
  private readonly combat: CombatSystem;
  private readonly net = new NetworkClient();
  private readonly remotes: RemotePlayerStore;
  private readonly enemies: EnemyStore;
  private readonly worldSync: WorldSync;
  private readonly playerSync = new PlayerSync();
  private readonly reconciliation = new Reconciliation();
  private readonly quests = new QuestStore();
  private readonly dialogue: DialogueUI;
  private readonly education: EducationUI;
  private trainingDps = 0;
  private readonly loadout = new LoadoutStore();
  private readonly inventoryUI: InventoryUI;
  private readonly characterPanel: CharacterProgressionUI;
  private readonly skillTree: SkillTreeUI;
  private readonly buildUI: BuildUI;
  private readonly formWheel: TransformationWheel;
  private readonly transformHud: TransformationUI;
  private readonly skillFlash: SkillUI;
  private readonly progression = new ProgressionStore();
  private transformFx: TransformationFX | null = null;
  private transformCam = 0;
  private dialogueCam = false;
  private readonly partyUI: PartyUI;
  private readonly inviteUI: PartyInviteUI;
  private readonly socialUI: SocialUI;
  private readonly chatUI: ChatUI;
  private readonly guildUI: GuildUI;
  private readonly tradeUI: TradeUI;
  private readonly shopUI: ShopUI;
  private readonly craftingUI: CraftingUI;
  private readonly lifeUI: LifeUI;
  private readonly petFollow: PetFollower;
  private photoMode = false;
  private decorateOn = false;
  private selectedDecor = "";
  private readonly remotePets = new Map<string, PetFollower>();
  private lastPetIds = new Map<string, string>();
  private readonly fishingUI: FishingUI;
  private readonly notifyCenter: NotifyCenter;
  private readonly dungeonUI: DungeonUI;
  private readonly pvpUI: PvpUI;
  private readonly pvpHud: PvpHud;
  private readonly pvpLocks = new Map<string, PvpTarget>();
  private readonly bossUI: BossUI;
  private readonly chapterUI: ChapterUI;
  private readonly discovery: DiscoverySystem;
  private readonly mountSys: MountSystem;
  private readonly mountCol: MountCollection;
  private readonly travel: FastTravel;
  private readonly eventUI: WorldEventUI;
  private readonly worldMap: WorldMap;
  private readonly journalUI: AdventureJournal;
  private readonly cinematicUI: CinematicUI;
  private readonly loreUI: LoreJournal;
  private readonly guidance = new GuidanceSystem();
  private readonly navRay = new THREE.Raycaster();
  private readonly navFrom = new THREE.Vector3();
  private readonly navDir = new THREE.Vector3();
  private readonly endgameUI: EndgameUI;
  private readonly music = new MusicManager();
  private journal: WorldJournal | null = null;
  private worldEvent: WorldEventView | null = null;
  private worldBoss: WorldBossView | null = null;
  private chapters: ChapterView[] = [];
  private inDungeon = false;
  private inPvp = false;
  private pvpMap = "";
  private readonly equipmentVisual = new EquipmentVisual();
  private readonly loop: GameLoop;
  private readonly canvas: HTMLCanvasElement;
  private readonly hudHost: HTMLElement;
  private touchUI = false;
  private clock = "09:00";
  private clockLabel = "Morning";

  constructor(canvas: HTMLCanvasElement, hudRoot: HTMLElement) {
    this.canvas = canvas;
    this.hudHost = hudRoot;
    this.renderer = new THREE.WebGLRenderer({
      canvas,
      antialias: true,
      alpha: false,
      powerPreference: "high-performance",
    });
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, this.state.embedMode ? 1.25 : 1.75));
    this.renderer.setSize(canvas.clientWidth || 640, canvas.clientHeight || 360, false);
    this.renderer.shadowMap.enabled = true;
    this.renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    this.renderer.toneMapping = THREE.ACESFilmicToneMapping;
    this.renderer.toneMappingExposure = 1.05;

    this.camera = new THREE.PerspectiveCamera(50, 16 / 9, 0.1, 280);
    this.world = new World();
    this.player = new PlayerFactory(this.world.scene).createLocalPlayer();
    this.player.stats.name = this.net.name;

    this.input = new InputManager();
    this.cameraRig = new CameraController(
      this.camera,
      this.player,
      canvas,
      this.world.environment.props,
      this.state.embedMode,
    );
    this.collision = new CollisionWorld();
    this.collision.aabbs.push(...VILLAGE_AABBS);
    this.collision.circles.push(...VILLAGE_CIRCLES);
    this.collision.trees.push(...VILLAGE_TREES);
    this.collision.rocks.push(...VILLAGE_ROCKS);
    this.collisionDebug = new CollisionDebug(this.world.scene, this.collision);
    this.controller = new PlayerController(
      this.player,
      this.input,
      () => this.cameraRig.getYaw(),
      this.collision,
      () => this.mountSys.speed(),
    );
    this.animation = new PlayerAnimationController();
    this.hud = new HUD(hudRoot);
    this.hud.onFollowPath = () => {
      this.guidance.followMain = !this.guidance.followMain;
    };
    this.hud.onCollidersDebug = (on) => this.collisionDebug.setShowColliders(on);
    this.hud.onCrowdDebug = (on) => this.collisionDebug.setShowCrowd(on);
    this.dialogue = new DialogueUI(hudRoot);
    this.education = new EducationUI(hudRoot);
    this.inventoryUI = new InventoryUI(hudRoot, this.loadout, () => this.player.stats.name);
    this.characterPanel = new CharacterProgressionUI(hudRoot);
    this.skillTree = new SkillTreeUI(hudRoot);
    this.buildUI = new BuildUI(hudRoot);
    this.formWheel = new TransformationWheel(hudRoot);
    this.transformHud = new TransformationUI(hudRoot);
    this.skillFlash = new SkillUI(hudRoot);
    this.partyUI = new PartyUI(hudRoot, this.loadout, () => this.net.playerId);
    this.inviteUI = new PartyInviteUI(hudRoot);
    this.socialUI = new SocialUI(hudRoot, this.loadout);
    this.chatUI = new ChatUI(hudRoot);
    this.guildUI = new GuildUI(hudRoot);
    this.tradeUI = new TradeUI(hudRoot, this.loadout, () => this.net.playerId);
    this.shopUI = new ShopUI(hudRoot, this.loadout);
    this.craftingUI = new CraftingUI(hudRoot);
    this.lifeUI = new LifeUI(hudRoot);
    this.petFollow = new PetFollower(this.player.group);
    this.fishingUI = new FishingUI(hudRoot);
    this.notifyCenter = new NotifyCenter(hudRoot);
    this.dungeonUI = new DungeonUI(hudRoot);
    this.pvpUI = new PvpUI(hudRoot);
    this.pvpHud = new PvpHud(hudRoot);
    this.endgameUI = new EndgameUI(hudRoot);
    this.bossUI = new BossUI(hudRoot, this.world.scene);
    this.chapterUI = new ChapterUI(hudRoot);
    this.discovery = new DiscoverySystem(hudRoot);
    this.mountSys = new MountSystem(this.player.group, this.player.mesh);
    this.mountCol = new MountCollection(hudRoot);
    this.travel = new FastTravel(hudRoot);
    this.eventUI = new WorldEventUI(hudRoot);
    this.worldMap = new WorldMap(hudRoot);
    this.journalUI = new AdventureJournal(hudRoot);
    this.cinematicUI = new CinematicUI(hudRoot);
    this.cinematicUI.onSkip = () => this.net.sendWorld("SKIP_CINEMATIC", {});
    this.cinematicUI.onDone = () => this.net.sendWorld("CINEMATIC_DONE", {});
    this.loreUI = new LoreJournal();
    this.loreUI.attach(hudRoot);
    this.travel.onTravel = (id) => this.net.sendWorld("FAST_TRAVEL", { landmarkId: id });
    this.worldMap.onTravel = (id) => this.travel.confirm(id);
    this.hud.onMap = () => this.worldMap.toggle(this.journal, { x: this.player.position.x, z: this.player.position.z }, this.loadout.party, this.worldEvent, this.worldBoss, this.worldHud().objects);
    this.hud.onMounts = () => {
      this.net.sendWorld("GET_MOUNTS", {});
      this.mountCol.toggle(this.mountSys.mounted);
    };
    this.mountCol.onSummon = (id) => this.net.sendWorld("REQUEST_MOUNT", { mountId: id });
    this.mountCol.onDismiss = () => this.net.sendWorld("DISMOUNT", {});
    this.mountCol.onFavorite = (id) => this.net.sendWorld("FAVORITE_MOUNT", { mountId: id });
    this.mountCol.onEquip = (id) => this.net.sendWorld("EQUIP_MOUNT", { mountId: id });
    this.hud.onJournal = () => {
      this.net.sendWorld("GET_WORLD_JOURNAL", {});
      this.net.sendWorld("GET_ADVENTURE", {});
      this.net.sendWorld("GET_STORY", {});
      this.journalUI.toggle(this.quests.progress);
    };
    this.hud.onLanguage = (lang) => this.net.sendWorld("SET_LANGUAGE", { language: lang });
    this.hud.onDialogueSpeed = (speed) => this.dialogue.setSpeed(speed);
    this.bindRpg();
    this.mobile = new MobileControls(hudRoot, this.input);
    this.remotes = new RemotePlayerStore(this.world.scene);
    this.enemies = new EnemyStore(this.world.scene);
    this.combat = new CombatSystem(
      this.world.scene,
      hudRoot,
      this.player,
      this.enemies,
      this.remotes,
      this.net,
      this.input,
      this.cameraRig,
      canvas,
    );
    this.transformFx = new TransformationFX(this.world.scene);
    this.worldSync = new WorldSync(this.remotes, this.enemies);
    this.bindNetwork();
    this.loop = new GameLoop((dt) => this.update(dt));

    this.syncTouchUI();
    this.resize();
    window.addEventListener("resize", this.onResize);
    hudRoot.querySelector("#hud-pause")?.addEventListener("click", () => this.setPaused(false));
    hudRoot.querySelector("#hud-inv-btn")?.addEventListener("click", () => this.toggleInventory());
    hudRoot.querySelector("#hud-char-btn")?.addEventListener("click", () => this.toggleCharacter());
    hudRoot.querySelector("#hud-craft-btn")?.addEventListener("click", () => this.toggleCrafting());
    hudRoot.querySelector("#hud-life-btn")?.addEventListener("click", () => this.toggleLife());
    hudRoot.querySelector("#hud-home-btn")?.addEventListener("click", () => this.toggleHome());
    hudRoot.querySelector("#hud-photo-btn")?.addEventListener("click", () => this.togglePhoto());
    hudRoot.querySelector("#hud-party-btn")?.addEventListener("click", () => this.partyUI.toggle());
    hudRoot.querySelector("#hud-social-btn")?.addEventListener("click", () => this.socialUI.toggle());
    hudRoot.querySelector("#hud-guild-btn")?.addEventListener("click", () => this.guildUI.toggle());
    this.hud.onAdventure = () => {
      this.net.sendWorld("GET_CHAPTERS", {});
      this.chapterUI.toggle(this.chapters);
    };
    this.mobile.onAdventure = this.hud.onAdventure;
    this.mobile.onMounts = this.hud.onMounts;
    this.mobile.onEndgame = this.hud.onEndgame;
    this.mobile.onDungeon = this.hud.onDungeon;
    this.mobile.onCraft = () => this.toggleCrafting();
    this.mobile.onLife = () => this.toggleLife();
    this.mobile.onHome = () => this.toggleHome();
    this.mobile.onPhoto = () => this.togglePhoto();
  }

  start(): void {
    this.state.phase = "playing";
    window.localStorage.setItem("cahaya-continue", "1");
    this.applyGraphics(this.hud.graphics);
    this.loop.start();
    this.net.connect();
  }

  useSession(token: string, username: string): void {
    this.net.setSession(token, username);
  }

  dispose(): void {
    this.loop.stop();
    this.net.disconnect();
    this.remotes.dispose();
    this.enemies.dispose();
    this.input.dispose();
    this.cameraRig.dispose();
    this.mobile.dispose();
    this.combat.dispose();
    this.inventoryUI.close();
    this.characterPanel.close();
    this.skillTree.close();
    this.buildUI.close();
    this.formWheel.close();
    this.partyUI.close();
    this.socialUI.close();
    this.guildUI.close();
    this.tradeUI.close();
    this.shopUI.close();
    this.craftingUI.close();
    this.fishingUI.close();
    window.removeEventListener("resize", this.onResize);
    this.renderer.dispose();
  }

  private setPaused(on: boolean): void {
    this.state.paused = on;
    if (on) this.cameraRig.exitLock();
    document.body.classList.toggle("is-paused", on);
  }

  private update(dt: number): void {
    this.input.pollGamepad();
    if (this.input.consumeOverlayClose() && this.cinematicUI.blocking) {
      this.cinematicUI.skip();
      this.input.consumePause();
      this.input.consumeDodge();
    } else if (this.input.consumeOverlayClose() && this.closeRpg()) {
      this.input.consumePause();
    } else if (this.input.consumePause()) {
      this.setPaused(!this.state.paused);
    }
    if (this.cinematicUI.blocking || this.dialogue.blocking) {
      if (this.input.consumeDodge()) this.cinematicUI.blocking ? this.cinematicUI.advance() : this.dialogue.advance();
    }
    if (this.fishingUI.blocking) {
      this.fishingUI.tick(dt);
      if (this.input.consumeDodge()) this.fishingUI.catchNow();
    }
    if (this.input.consumeInventory()) this.toggleInventory();
    if (this.input.consumeCharacter()) this.toggleCharacter();
    if (this.input.consumeParty()) this.partyUI.toggle();
    if (this.input.consumeSocial()) this.socialUI.toggle();
    if (this.input.consumeGuild()) this.guildUI.toggle();
    if (this.input.consumeChat()) this.chatUI.focus();
    if (this.input.consumeMount()) {
      if (this.mountSys.mounted) this.net.sendWorld("DISMOUNT", {});
      else this.net.sendWorld("REQUEST_MOUNT", { mountId: this.mountSys.mountId || "wind-runner" });
    }
    if (this.input.consumeMounts()) {
      this.net.sendWorld("GET_MOUNTS", {});
      this.mountCol.toggle(this.mountSys.mounted);
    }
    if (this.input.consumeJournal()) {
      this.net.sendWorld("GET_WORLD_JOURNAL", {});
      this.journalUI.toggle(this.quests.progress);
    }
    if (this.input.consumeMap()) {
      this.net.sendWorld("GET_WORLD_JOURNAL", {});
      this.net.sendWorld("GET_OPEN_WORLD", {});
      this.worldMap.toggle(this.journal, { x: this.player.position.x, z: this.player.position.z }, this.loadout.party, this.worldEvent, this.worldBoss, this.worldHud().objects);
    }
    if (this.input.consumePvp()) {
      this.pvpUI.toggle();
      if (this.pvpUI.blocking) this.net.sendWorld("GET_PVP", {});
    }
    if (this.input.consumeEndgame()) {
      this.endgameUI.toggle();
      if (this.endgameUI.blocking) this.net.sendWorld("GET_ENDGAME", {});
    }
    const uiBlock = this.dialogue.blocking || this.education.blocking || this.rpgBlocking() || this.chatUI.focused || this.cinematicUI.blocking || this.fishingUI.blocking;
    this.combat.inputPaused = uiBlock;
    if (this.input.consumeTransform()) {
      if (uiBlock) {
        /* restricted */
      } else {
        this.net.sendWorld("REQUEST_TRANSFORMATION", { formId: this.progression.selectedForm });
      }
    }
    if (this.input.consumePotion()) {
      const slot = this.loadout.slots.find((s) => s.item?.id === "potion_heal" || s.item?.type === "CONSUMABLE");
      if (slot) this.net.sendWorld("USE_ITEM", { slot: slot.index });
    }

    if (!this.state.paused) {
      this.combat.prepare(dt);
      const move: MoveMode = uiBlock || this.photoMode || this.player.health.isDead() || this.player.combatState === "DOWNED" || this.player.combatState === "HIT" || this.player.combatState === "STUNNED"
        ? "halt"
        : this.combat.dodge.active
          ? "dash"
          : this.combat.charging
            ? "halt"
          : this.combat.attacks.attacking
            ? "halt"
            : "normal";
      const jump = this.input.peekJumpQueued();
      this.controller.update(dt, move);
      this.combat.resolve(dt);
      const stop = this.combat.consumeHitStop();
      if (stop > 0) this.loop.addHitStop(stop);
      const moving = this.player.currentSpeed > 0.35;
      const sprinting =
        move === "normal" &&
        this.input.isSprintPressed() &&
        moving &&
        this.player.isGrounded &&
        !this.player.stats.exhausted &&
        this.player.stats.stamina > 1;
      this.tickGuidance(dt, moving);
      this.animation.update(this.player, dt, moving, sprinting, this.guidance.glance);
      this.playerSync.sendLocalInput(
        this.net,
        move === "normal" ? this.input.compositeStrafe() : 0,
        move === "normal" ? this.input.compositeForward() : 0,
        this.cameraRig.getYaw(),
        move === "normal" && this.input.isSprintPressed(),
        move === "normal" && jump,
        performance.now(),
      );
      this.reconciliation.apply(this.player, this.worldSync.serverSelf, dt);
      this.petFollow.update(dt);
      this.syncRemotePets();
      this.handleDecorate();
      this.remotes.update(dt, this.camera);
      this.enemies.update(dt, this.camera, this.combat.targets.current?.id ?? "");
      this.syncLivingWorld(dt);
      this.world.zones.update(this.player.position.z);
      this.world.weather.update(dt);
      this.discovery.update();
      const zone = this.world.zones.current();
      this.music.set(this.combat.view.target && zone !== "masjid" ? "combat" : zone === "masjid" ? "calm" : zone === "forest" ? "forest" : "explore");
      const formId = this.progression.view?.formId || "normal";
      const active = this.progression.view?.transformState === "TRANSFORMING" || this.progression.view?.transformState === "TRANSFORMED";
      this.transformFx?.follow(this.player.position, formId, !!active, dt);
      if (this.transformCam > 0) {
        this.transformCam -= dt;
        if (this.transformCam <= 0) this.cameraRig.endCinematic();
      }
    }

    if (!this.dialogue.blocking && this.dialogueCam) {
      this.dialogueCam = false;
      this.world.npcs.talkingId = "";
      if (this.transformCam <= 0 && !this.cinematicUI.blocking) this.cameraRig.endCinematic();
    }

    this.cameraRig.setPresence(cameraPresence(this.combat.targets.current));
    this.cameraRig.update(dt);
    this.cinematicUI.update(dt);
    this.partyUI.tick();
    this.hud.update(this.player, this.cameraRig, this.state, dt, this.combat.view, {
      status: this.net.status,
      playerId: this.net.playerId,
      online: this.worldSync.online,
      channel: this.worldSync.channel || this.net.channel,
      pingMs: this.net.pingMs,
      messagesPerSec: this.net.messagesPerSec,
      serverPos: this.reconciliation.serverPos,
      interpOffset: this.remotes.meanOffset(),
      reconOffset: this.reconciliation.offset,
      others: this.remotes.dots(),
    }, this.worldHud());
    this.transformHud.update(this.progression);
    this.bossUI.update();
    this.renderer.render(this.world.scene, this.camera);
  }

  private bindNetwork(): void {
    this.net.onWelcome = (w) => {
      this.worldSync.applyWelcome(w, this.player);
      this.worldSync.applyVitals(this.player);
      if (w.progress) {
        this.quests.apply(w.progress);
        if (w.progress.chapters) this.chapters = w.progress.chapters;
      }
      if (w.catalog) this.loadout.setCatalog(w.catalog);
      if (w.loadout) this.applyLoadout(w.loadout);
      if (w.social) {
        this.loadout.applySocial(w.social);
        this.socialUI.render();
        this.partyUI.render();
        this.guildUI.apply(w.social.guild);
        this.notifyCenter.apply(w.social.notifies);
      }
      if (w.progression) this.applyProgression(w.progression);
      else this.net.sendWorld("GET_PROGRESSION", {});
      this.net.sendWorld("GET_WORLD_JOURNAL", {});
      this.refreshWorldActors();
    };
    this.net.onSpawn = (p) => this.worldSync.spawn(p, this.net.playerId);
    this.net.onDespawn = (id) => this.worldSync.despawn(id);
    this.net.onSnapshot = (s) => {
      this.worldSync.applySnapshot(s, this.net.playerId);
      this.worldSync.applyVitals(this.player);
      if (s.timeOfDay) this.world.environment.setPhase(s.timeOfDay);
      if (s.weather) this.world.weather.apply(this.world.scene, s.weather);
      if (s.clock) this.clock = s.clock;
      if (s.clockLabel) this.clockLabel = s.clockLabel;
      this.eventUI.apply(s.event);
      this.worldEvent = s.event ?? null;
      this.worldBoss = s.worldBoss ?? null;
      if (this.worldBoss) this.eventUI.applyBoss(this.worldBoss);
      const self = s.players.find((p) => p.id === this.net.playerId);
      if (self && (self.mounted !== this.mountSys.mounted || (self.mountId || "") !== this.mountSys.mountId)) {
        this.mountSys.apply(!!self.mounted, self.mountId || "", resolveGraphics(this.hud.graphics));
      }
      if (self) this.petFollow.apply(self.petId || this.lifeUI.view?.activePet || "", resolveGraphics(this.hud.graphics));
      this.lastPetIds.clear();
      for (const pl of s.players) this.lastPetIds.set(pl.id, pl.petId || "");
      this.cameraRig.setMounted(this.mountSys.mounted);
      const pvp = !!s.pvp?.matchId;
      const dungeon = !!s.instanceId && !pvp;
      if (pvp) {
        this.inPvp = true;
        this.inDungeon = false;
        this.pvpMap = s.pvp?.map || "";
        this.world.setInstance(this.pvpMap === "valley-of-dawn" || this.pvpMap === "dawn-arena" ? "bg" : "arena");
        this.pvpHud.apply(s.pvp, this.net.playerId);
        this.syncPvpTargets();
      } else if (dungeon !== this.inDungeon || this.inPvp) {
        this.inPvp = false;
        this.inDungeon = dungeon;
        this.world.setInstance(dungeon ? "dungeon" : "hub");
        this.pvpHud.clear();
        this.combat.pvpTargets = [];
        if (!dungeon) {
          this.dungeonUI.clear();
          this.bossUI.clear();
        }
      }
      if (s.dungeon) {
        this.dungeonUI.applyState(s.dungeon, this.player.position.x, this.player.position.z);
        this.bossUI.apply(s.dungeon);
      }
      this.refreshWorldActors();
    };
    this.net.onCombat = (type, data) => this.combat.controller.handle(type, data);
    this.net.onWorld = (type, data) => this.handleWorld(type, data);
  }

  private syncTouchUI(): void {
    this.touchUI =
      !this.state.embedMode &&
      (window.matchMedia("(pointer: coarse)").matches || window.innerWidth <= 820);
    this.mobile.setVisible(this.touchUI);
    this.hud.setMobile(this.touchUI);
    this.cameraRig.setPointerLockEnabled(!this.touchUI && !this.state.embedMode);
    document.body.classList.toggle("touch-ui", this.touchUI);
    document.body.classList.toggle("desktop-ui", !this.touchUI);
  }

  private onResize = (): void => {
    this.syncTouchUI();
    this.resize();
  };

  private resize = (): void => {
    const parent = this.canvas.parentElement;
    const w = parent?.clientWidth || window.innerWidth;
    const h = parent?.clientHeight || window.innerHeight;
    this.camera.aspect = w / Math.max(1, h);
    this.camera.updateProjectionMatrix();
    this.renderer.setSize(w, h, false);
  };

  private refreshWorldActors(): void {
    const snap = this.worldSync.lastSnapshot;
    if (!snap) return;
    this.world.npcs.sync(snap.npcs ?? [], (id) => this.quests.markerFor(id));
    const objects = [...(snap.objects ?? [])];
    for (const drop of snap.drops ?? []) {
      objects.push({ id: drop.id, kind: "drop", x: drop.x, z: drop.z, text: drop.name });
    }
    this.world.interact.sync(objects, (id) => this.quests.claimed(id), this.quests.forestUnlocked());
  }

  private syncLivingWorld(dt: number): void {
    this.world.npcs.update(dt, this.camera, this.player.position.x, this.player.position.z);
    if (this.world.environment.isHubMode()) {
      this.collision.resolvePlayerPosition(this.player, this.world.npcs.positions());
      this.collisionDebug.update(this.player.position.x, this.player.position.z, this.world.npcs.positions());
    }
    this.refreshWorldActors();
    if (this.dialogue.blocking || this.education.blocking || this.rpgBlocking() || this.cinematicUI.blocking) {
      this.input.consumeInteract();
      return;
    }
    if (this.input.consumeInteract()) {
      const target = this.nearestInteract();
      if (!target) return;
      if (target.kind === "drop") this.net.sendWorld("PICKUP_ITEM", { dropId: target.id });
      else if (target.id === "arena_master") {
        this.pvpUI.open();
        this.net.sendWorld("GET_PVP", {});
      } else if (["challenge_master", "guardian_archivist", "season_keeper", "raid_keeper", "explorer_npc", "cosmetic_merchant"].includes(target.id)) {
        this.endgameUI.open();
        this.net.sendWorld("GET_ENDGAME", {});
      } else this.net.sendWorld("INTERACT", { targetId: target.id, kind: target.kind });
    }
  }

  private tickGuidance(dt: number, moving: boolean): void {
    const j = this.quests.progress?.journey;
    const navX = j?.navX ?? 0;
    const navZ = j?.navZ ?? this.player.position.z + 12;
    const zone = this.world.zones.zoneAt(this.player.position.z);
    const npc = this.world.npcs.nearest(this.player.position.x, this.player.position.z, 14);
    this.guidance.update(dt, {
      x: this.player.position.x,
      y: this.player.position.y,
      z: this.player.position.z,
      navX,
      navZ,
      landmark: j?.landmark || "tujuan",
      zone: zone.id,
      zoneTitle: zone.titleId || zone.name,
      moving,
      combat: !!this.combat.targets.current,
      npcName: npc?.name,
    });
    this.world.path.setFollow(this.guidance.followMain, this.player.position.z);
    const dist = Math.hypot(navX - this.player.position.x, navZ - this.player.position.z);
    this.navFrom.set(this.player.position.x, this.player.position.y + 1.4, this.player.position.z);
    this.navDir.set(navX - this.navFrom.x, 0, navZ - this.navFrom.z);
    const len = this.navDir.length();
    if (len > 3) {
      this.navDir.multiplyScalar(1 / len);
      this.navRay.set(this.navFrom, this.navDir);
      this.navRay.far = Math.max(0.5, len - 1.4);
      const hits = this.navRay.intersectObjects(this.world.environment.props, true);
      this.guidance.occluded = hits.some((h) => h.object.visible && h.distance > 0.5 && h.distance < len - 1.2);
    } else {
      this.guidance.occluded = false;
    }
    if (!this.combat.targets.current && this.guidance.maybePeek(j?.landmark || zone.id, dist)) {
      this.cameraRig.peekAt(navX, 3.4, navZ, 2.8);
    }
  }

  private nearestInteract(): { id: string; kind: string } | null {
    const x = this.player.position.x;
    const z = this.player.position.z;
    const npc = this.world.npcs.nearest(x, z, 2.6);
    const obj = this.world.interact.nearest(x, z, 2.4);
    const nd = npc ? Math.hypot(npc.x - x, npc.z - z) : 99;
    const od = obj ? Math.hypot(obj.x - x, obj.z - z) : 99;
    const npcNear = !!(npc && nd <= od && nd <= 2.6);
    if (npcNear) return { id: npc.id, kind: "npc" };
    if (obj && od <= 2.4) return { id: obj.id, kind: obj.kind };
    return null;
  }

  private worldHud() {
    const near = this.nearestInteract();
    return {
      zone: this.inPvp ? pvpZoneName(this.pvpMap) : this.inDungeon ? this.dungeonUI.view?.name || "Dungeon" : this.world.zones.zoneAt(this.player.position.z).name,
      interact: near ? gatherPrompt(near.kind, this.touchUI) : "",
      progress: this.quests.progress,
      coin: this.loadout.coin,
      crystal: this.loadout.crystal,
      battleToken: this.loadout.battleToken,
      guardianToken: this.loadout.guardianToken,
      raidToken: this.loadout.raidToken,
      eduToken: this.loadout.eduToken,
      knowledge: this.quests.progress?.knowledgePoints ?? 0,
      clock: this.clock,
      clockLabel: this.clockLabel,
      navX: this.quests.progress?.journey?.navX ?? questNav(this.quests.progress)?.x,
      navZ: this.quests.progress?.journey?.navZ ?? questNav(this.quests.progress)?.z,
      journey: this.quests.progress?.journey,
      optionalX: this.quests.progress?.journey?.optionalX,
      optionalZ: this.quests.progress?.journey?.optionalZ,
      landmarks: (this.journal?.landmarks ?? []).filter((l) => l.discovered || Math.hypot(l.x - this.player.position.x, l.z - this.player.position.z) < 36),
      cameraYaw: this.cameraRig.yaw,
      inCombat: !!this.combat.targets.current,
      guidancePulse: this.guidance.pulse,
      guidanceToast: this.guidance.toastVisible() ? this.guidance.toast : "",
      guidanceSub: this.guidance.toastVisible() ? this.guidance.toastSub : "",
      regionBanner: this.guidance.bannerVisible() ? this.guidance.regionBanner : "",
      followMain: this.guidance.followMain,
      occluded: this.guidance.occluded,
      npcs: this.world.npcs.all().map((n) => ({ x: n.x, z: n.z, marker: this.quests.markerFor(n.id) })),
      objects: this.world.interact.all().map((o) => ({ x: o.x, z: o.z, kind: o.kind })),
      enemies: this.enemies.list().map((e) => ({ x: e.position.x, z: e.position.z })),
      formLabel: this.progression.formLabel(),
      transformReady: this.progression.readyLabel(),
      transEnergy: this.progression.view?.transEnergy ?? this.progression.transform?.energy ?? 0,
      maxTransEnergy: this.progression.view?.maxTransEnergy ?? this.progression.transform?.maxEnergy ?? 1,
      ultCharge: this.progression.view?.ultCharge ?? 0,
      dps: this.trainingDps,
    };
  }

  private handleWorld(type: NetMsgType, data: unknown): void {
    if (type === "ENDGAME_STATE") {
      this.endgameUI.apply(data as EndgameState);
      if (!this.endgameUI.blocking) this.endgameUI.open();
      return;
    }
    if (type === "ACHIEVEMENT_UNLOCKED") {
      this.hud.toast("Achievement unlocked");
      this.notifyCenter.push("achievement", "Achievement unlocked");
      return;
    }
    if (type === "QUEST_UPDATED") {
      const progress = data as PlayerProgress;
      this.quests.apply(progress);
      if (progress.chapters) this.chapters = progress.chapters;
      this.refreshWorldActors();
      return;
    }
    if (type === "EDUCATION_QUESTION") {
      this.dialogue.close();
      this.education.showQuestion(data as QuestionOut, (choice) => {
        this.net.sendWorld("EDUCATION_ANSWER", { questionId: (data as QuestionOut).id, choice });
      });
      return;
    }
    if (type === "EDUCATION_FEEDBACK") {
      const fb = data as EducationFeedback;
      if (fb.toast) this.hud.toast(fb.toast);
      else if (!fb.correct) this.hud.toast("Coba lagi!");
      if (fb.explain) this.hud.toast(fb.explain);
      if (fb.question) {
        this.education.showQuestion(fb.question, (choice) => {
          this.net.sendWorld("EDUCATION_ANSWER", { questionId: fb.question?.id, choice });
        });
      } else if (fb.correct) {
        this.education.close();
      }
      return;
    }
    if (type === "QUEST_REWARD" || type === "INTERACT_RESULT") {
      const res = data as InteractResult;
      if (res.question) {
        this.dialogue.close();
        this.education.showQuestion(res.question, (choice) => {
          this.net.sendWorld("EDUCATION_ANSWER", { questionId: res.question?.id, choice });
        });
        return;
      }
      if (res.rewards) {
        const parts = [`+${res.rewards.exp || 0} EXP`, `+${res.rewards.coin || 0} KOIN`];
        if (res.rewards.potion) parts.push(`Potion x${res.rewards.potion}`);
        if (res.rewards.crystal) parts.push(`Crystal x${res.rewards.crystal}`);
        if (res.rewards.eduToken) parts.push(`Education Token x${res.rewards.eduToken}`);
        if (res.rewards.perfect) parts.push("PERFECT!");
        this.hud.toast(`${res.title || "QUEST COMPLETE"}\n${parts.join(" · ")}`);
      }
      if (res.shop?.length) {
        res.text = `${res.text}\n\n${res.shop.map((s) => `• ${s.name} — ${s.price} koin`).join("\n")}`;
      }
      if (res.toast) this.hud.toast(res.toast);
      this.dialogue.open(res.title || res.speaker, res.speaker, res.role || "", res.text, res.options ?? [], (id) => this.onDialogOption(id, res), res.subtitle || "", {
        emotion: res.emotion, gesture: res.gesture, history: res.history,
      });
      if (res.kind === "npc" && res.targetId) {
        this.world.npcs.talkingId = res.targetId;
        this.cameraRig.beginCinematic("focus");
        this.dialogueCam = true;
      }
      if (res.cinematicId && !this.cinematicUI.blocking) {
        this.net.sendWorld("GET_STORY", {});
      }
      return;
    }
    if (type === "CINEMATIC_START") {
      const cin = data as { id?: string; title?: string; durationSec?: number; lines?: string[]; subtitle?: string[]; camera?: string; music?: string };
      this.cinematicUI.open(cin);
      this.cameraRig.beginCinematic(cin.camera);
      if (cin.music) this.music.set(cin.music);
      return;
    }
    if (type === "CINEMATIC_SKIPPED" || type === "STORY_STATE") {
      if (type === "CINEMATIC_SKIPPED") {
        this.cinematicUI.close();
        this.cameraRig.endCinematic();
      }
      return;
    }
    if (this.handleDungeon(type, data)) return;
    if (this.handlePvp(type, data)) return;
    if (this.handleProgression(type, data)) return;
    if (type === "ZONE_DISCOVERED") {
      const z = data as { name?: string; exp?: number; toast?: string };
      const label = (z.toast || `${(z.name || "New Area").toUpperCase()} DISCOVERED`).toUpperCase();
      this.discovery.show(z.name || "New Area", z.exp || 0, label);
      this.hud.toast(`${label}\n+${z.exp || 0} EXP`);
      return;
    }
    if (type === "LANDMARK_DISCOVERED") {
      const lm = data as { name?: string };
      this.discovery.landmark(lm.name || "Landmark");
      return;
    }
    if (type === "LORE_DISCOVERED") {
      const lore = data as LoreView;
      this.loreUI.popup(this.hudHost, lore.title, lore.text);
      return;
    }
    if (type === "WORLD_JOURNAL") {
      this.journal = data as WorldJournal;
      this.journalUI.apply(this.journal);
      this.world.landmarks.sync(this.journal.landmarks, this.player.position.z);
      return;
    }
    if (type === "OPEN_WORLD") {
      const ow = data as { markers?: WorldJournal["markers"]; nextWorldBoss?: WorldJournal["nextWorldBoss"]; regions?: WorldJournal["regions"] };
      if (this.journal) {
        if (ow.markers) this.journal.markers = ow.markers;
        if (ow.nextWorldBoss) this.journal.nextWorldBoss = ow.nextWorldBoss;
        if (ow.regions) this.journal.regions = ow.regions;
      }
      return;
    }
    if (type === "WORLD_EVENT") {
      this.worldEvent = data as WorldEventView;
      this.eventUI.apply(this.worldEvent);
      this.hud.toast(`WORLD EVENT\n${this.worldEvent.announce || this.worldEvent.name}`);
      return;
    }
    if (type === "WORLD_BOSS_ANNOUNCE" || type === "WORLD_BOSS_STATE") {
      this.worldBoss = data as WorldBossView;
      this.eventUI.applyBoss(this.worldBoss);
      this.hud.toast(this.worldBoss.announce || this.worldBoss.name);
      return;
    }
    if (type === "ENTER_MASJID") {
      this.music.set("calm");
      this.hud.toast("MASJID CAHAYA\nTempat aman. Combat dimatikan.");
      return;
    }
    if (type === "CHAPTER_COMPLETE") {
      const c = data as { title?: string; guardians?: number; guardiansTotal?: number };
      this.hud.toast(`${c.title || "CHAPTER COMPLETE"}\n33 GUARDIANS: ${c.guardians ?? 0}/${c.guardiansTotal ?? 33}\nMASJID CAHAYA: DISCOVERED`);
      return;
    }
    if (type === "COLLECTION_BOOK") {
      return;
    }
    if (type === "MOUNT_UPDATED") {
      const m = data as { mounted?: boolean; mountId?: string };
      this.mountSys.apply(!!m.mounted, m.mountId || "", resolveGraphics(this.hud.graphics));
      this.cameraRig.setMounted(!!m.mounted);
      return;
    }
    if (type === "MOUNT_COLLECTION") {
      this.mountCol.apply(data as { mounts?: Array<{ mountId: string; name: string }>; active?: string; mounted?: boolean; mountId?: string });
      if (this.mountCol.blocking) this.mountCol.open(this.mountSys.mounted);
      return;
    }
    if (type === "TRAVEL_SUGGESTION") {
      const s = data as { landmarkId?: string; text?: string };
      this.hud.toast(`FOLLOW PARTY\n${s.text || s.landmarkId || ""}`);
      return;
    }
    if (type === "TRAVEL_EVENT") {
      const e = data as { text?: string; kind?: string; action?: string; toast?: string };
      this.hud.toast(e.toast || e.text || "Travel event");
      return;
    }
    if (type === "RACE_UPDATED") {
      const r = data as { state?: string; checkpoint?: string; elapsed?: number };
      this.hud.toast(`DAWN RACE\n${r.state || r.checkpoint || ""}`);
      return;
    }
    if (type === "FAST_TRAVEL_OK") {
      this.hud.toast("FAST TRAVEL");
      return;
    }
    if (type === "RANDOM_ENCOUNTER") {
      const enc = data as { text?: string; kind?: string };
      this.hud.toast(`RANDOM ENCOUNTER\n${enc.text || enc.kind || ""}`);
      if (enc.kind === "lost-traveler") {
        this.dialogue.open("Lost Traveler", "Pengembara", "NPC", enc.text || "Le, aku kesasar. Iso tulung aku?", [
          { id: "travel-help", label: "HELP" },
          { id: "travel-ignore", label: "IGNORE" },
        ], (id) => {
          this.net.sendWorld("TRAVEL_EVENT", { kind: "lost-traveler", action: id === "travel-help" ? "help" : "ignore" });
          this.dialogue.close();
        }, "Le, aku kesasar. Bisa bantu aku?");
      }
      return;
    }
    if (type === "WEATHER_UPDATED") {
      const w = data as { weather?: string; clock?: string; label?: string; phase?: string };
      if (w.clock) this.clock = w.clock;
      if (w.label) this.clockLabel = w.label;
      if (w.phase) this.world.environment.setPhase(w.phase);
      if (w.weather) this.world.weather.apply(this.world.scene, w.weather);
      return;
    }
    if (type === "GUARDIAN_DEFEATED") {
      this.hud.toast("GUARDIAN DEFEATED");
      this.net.sendWorld("GET_WORLD_JOURNAL", {});
      return;
    }
    if (type === "EVENT_REWARD") {
      this.hud.toast("EVENT REWARD");
      return;
    }
    if (type === "ACTION_REJECT") {
      const reason = String((data as { reason?: string; action?: string })?.reason ?? "");
      this.hud.toast(rejectText(reason));
      return;
    }
    if (type === "INVENTORY_UPDATED" || type === "ITEM_ADDED" || type === "ITEM_REMOVED" || type === "ITEM_USED" || type === "ITEM_CONSUMED" || type === "EQUIPMENT_UPDATED") {
      this.applyLoadout(data as InventoryUpdated);
      if (this.shopUI.blocking) this.shopUI.render();
      if (this.tradeUI.blocking) this.tradeUI.render();
      return;
    }
    if (type === "PLAYER_STATS_UPDATED") {
      this.loadout.applyStats(data as StatsView);
      this.applyRpgStats();
      this.characterPanel.render(this.loadout, this.progression, this.player.stats.name);
      return;
    }
    if (type === "PARTY_UPDATED" || type === "PARTY_MEMBER_JOINED" || type === "PARTY_MEMBER_LEFT") {
      this.loadout.applyParty(data as PartyView);
      this.partyUI.render();
      this.socialUI.render();
      return;
    }
    if (type === "PARTY_INVITE") {
      const inv = data as { from?: string; fromId?: string };
      this.inviteUI.showParty(inv.from || "Player", inv.fromId || "");
      this.inviteUI.onAccept = () => this.net.sendWorld("PARTY_ACCEPT", {});
      this.inviteUI.onDecline = () => this.net.sendWorld("PARTY_DECLINE", {});
      return;
    }
    if (type === "FRIEND_REQUEST") {
      const inv = data as { from?: string; fromId?: string };
      this.inviteUI.showFriend(inv.from || "Player", inv.fromId || "");
      this.inviteUI.onAccept = () => this.net.sendWorld("ACCEPT_FRIEND", { targetId: this.inviteUI.fromId });
      this.inviteUI.onDecline = () => this.net.sendWorld("DECLINE_FRIEND", { targetId: this.inviteUI.fromId });
      return;
    }
    if (type === "FRIEND_UPDATED") {
      this.loadout.applySocial(data as SocialState);
      this.socialUI.render();
      this.partyUI.render();
      this.guildUI.apply((data as SocialState).guild);
      this.notifyCenter.apply((data as SocialState).notifies);
      return;
    }
    if (type === "PRIVACY_UPDATED") {
      this.loadout.social.privacy = data as SocialState["privacy"];
      if (this.socialUI.blocking) this.socialUI.render();
      return;
    }
    if (type === "INSPECT_RESULT") {
      this.socialUI.inspect.show(data as InspectOut, this.loadout);
      return;
    }
    if (type === "SOCIAL_NOTIFICATION") {
      const note = data as SocialNote;
      if (note.text) this.hud.toast(note.text);
      this.loadout.pushNote(note.text);
      this.notifyCenter.push(note.kind || "system", note.text || "");
      return;
    }
    if (type === "CHAT_MESSAGE") {
      this.chatUI.push(data as ChatOut);
      return;
    }
    if (type === "GUILD_UPDATED") {
      this.guildUI.apply(data as GuildView);
      if (this.loadout.social.guild) this.loadout.social.guild = data as GuildView;
      this.socialUI.render();
      return;
    }
    if (type === "GUILD_INVITE") {
      const inv = data as { from?: string; fromId?: string };
      this.inviteUI.showGuild(inv.from || "Player", inv.fromId || "");
      this.inviteUI.onAccept = () => this.net.sendWorld("GUILD_ACCEPT", {});
      this.inviteUI.onDecline = () => this.net.sendWorld("GUILD_DECLINE", {});
      this.notifyCenter.push("guild_invite", `${inv.from || "Player"} invited you to a guild.`);
      return;
    }
    if (type === "TRADE_REQUEST") {
      const inv = data as { from?: string; fromId?: string };
      this.inviteUI.showTrade(inv.from || "Player", inv.fromId || "");
      this.inviteUI.onAccept = () => this.net.sendWorld("TRADE_ACCEPT", {});
      this.inviteUI.onDecline = () => this.net.sendWorld("TRADE_DECLINE", {});
      this.notifyCenter.push("trade_request", `${inv.from || "Player"} wants to trade.`);
      return;
    }
    if (type === "TRADE_UPDATED") {
      this.tradeUI.apply(data as TradeView);
      this.socialUI.trade = data as TradeView;
      this.socialUI.render();
      return;
    }
    if (type === "MARKET_LISTINGS") {
      this.socialUI.market = data as MarketState;
      if (this.socialUI.blocking) this.socialUI.render();
      return;
    }
    if (type === "HOUSE_STATE") {
      this.socialUI.house = data as HouseState;
      if (this.socialUI.blocking) this.socialUI.render();
      this.input.decorateMode = this.decorateOn && !!(data as HouseState).instanceId && !(data as HouseState).left;
      return;
    }
    if (type === "LIFE_STATE" || type === "PET_STATE") {
      const d = data as LifeState & { active?: string; pets?: LifeState["pets"] };
      this.lifeUI.apply({ ...d, activePet: d.activePet || d.active });
      const petId = d.activePet || d.active || "";
      this.petFollow.apply(petId, resolveGraphics(this.hud.graphics));
      return;
    }
    if (type === "PLAYER_CARD") {
      this.socialUI.card.show(data as PlayerCardView);
      return;
    }
    if (type === "EMOTE_PLAYED") {
      const em = data as { emote?: string; playerId?: string };
      if (em.emote) this.hud.toast(`${em.playerId === this.net.playerId ? "You" : "Player"}: ${em.emote}`);
      return;
    }
    if (type === "BANK_UPDATED" || type === "GUILD_LOG") {
      if (this.socialUI.blocking) this.socialUI.render();
      return;
    }
    if (type === "SHOP_CATALOG") {
      const cat = data as ShopCatalogOut & { mode?: string };
      if (cat.mode === "material") {
        this.craftingUI.tab = "shop";
        this.craftingUI.shopId = cat.id || "mbah-karya-shop";
        this.craftingUI.open();
        this.net.sendWorld("GET_CRAFTING", {});
        this.dialogue.close();
        return;
      }
      this.shopUI.apply(cat);
      this.dialogue.close();
      return;
    }
    if (type === "CRAFTING_STATE") {
      this.craftingUI.apply(data as CraftingState);
      if (this.craftingUI.blocking) this.craftingUI.render();
      return;
    }
    if (type === "GATHER_RESULT") {
      const g = data as { anim?: string; resourceId?: string };
      this.player.combatPose = g.anim === "chop" ? "KICK" : g.anim === "cast" ? "KICK" : "PUNCH";
      if (g.anim === "mining") this.combat.audio.mine();
      else if (g.anim === "chop") this.combat.audio.woodcut();
      else if (g.anim === "cast") this.combat.audio.fish();
      else this.combat.audio.gather();
      window.setTimeout(() => {
        if (this.player.combatPose === "PUNCH" || this.player.combatPose === "KICK") this.player.combatPose = null;
      }, 280);
      this.hud.toast(`Gather ${g.resourceId || ""}`);
      this.net.sendWorld("GET_CRAFTING", {});
      return;
    }
    if (type === "CRAFT_RESULT") {
      const c = data as { result?: string; quality?: string };
      this.combat.audio.craft();
      this.hud.toast(`Craft ${c.result || ""} · ${c.quality || "NORMAL"}`);
      this.net.sendWorld("GET_INVENTORY", {});
      return;
    }
    if (type === "FISH_STATE") {
      const f = data as FishState;
      if (f.caught) {
        this.fishingUI.close();
        this.hud.toast(`Iwak: ${f.reward || "fish"}`);
        this.net.sendWorld("GET_CRAFTING", {});
        return;
      }
      this.fishingUI.start(f.targetA ?? 42, f.targetB ?? 68, f.prompt);
      this.combat.audio.fish();
      this.fishingUI.onCatch = (progress) => this.net.sendWorld("FISH_CATCH", { spotId: f.spotId, progress });
      return;
    }
    if (type === "PARTY_FINDER_LIST") {
      this.partyUI.listings = (data as { listings?: FinderListing[] }).listings ?? [];
      this.partyUI.render();
      return;
    }
    if (type === "SEARCH_RESULT") {
      this.socialUI.searchHits = (data as { results?: SearchHit[] }).results ?? [];
      this.socialUI.render();
      return;
    }
    if (type === "NOTIFY_LIST") {
      this.notifyCenter.apply((data as { notifies?: NotifyView[] }).notifies);
      return;
    }
    if (type === "ERROR") {
      const msg = String((data as { message?: string })?.message ?? "");
      if (msg) this.hud.toast(msg);
    }
  }

  private bindRpg(): void {
    this.inventoryUI.onUse = (slot) => this.net.sendWorld("USE_ITEM", { slot });
    this.inventoryUI.onEquip = (slot) => this.net.sendWorld("EQUIP_ITEM", { slot });
    this.inventoryUI.onUnequip = (equipSlot) => this.net.sendWorld("UNEQUIP_ITEM", { equipSlot });
    this.inventoryUI.onDiscard = (slot) => this.net.sendWorld("DISCARD_ITEM", { slot });
    this.inventoryUI.onLock = (slot, on) => this.net.sendWorld("LOCK_ITEM", { slot, on });
    this.inventoryUI.onFavorite = (slot, on) => this.net.sendWorld("FAVORITE_ITEM", { slot, on });
    this.inventoryUI.onSplit = (slot, qty) => this.net.sendWorld("SPLIT_STACK", { slot, qty });
    this.inventoryUI.onUpgrade = (slot) => this.net.sendWorld("UPGRADE_ITEM", { slot });
    this.inventoryUI.onExpand = () => this.net.sendWorld("EXPAND_BAG", {});
    this.inventoryUI.onLoadoutSave = (slot) => this.net.sendWorld("SAVE_GEAR_LOADOUT", { slot });
    this.inventoryUI.onLoadoutLoad = (slot) => this.net.sendWorld("LOAD_GEAR_LOADOUT", { slot });
    this.inventoryUI.onClaimTemp = () => this.net.sendWorld("CLAIM_TEMP_LOOT", {});
    this.inventoryUI.onCosmetic = () => this.net.sendWorld("TOGGLE_COSMETIC", {});
    this.characterPanel.onUnequip = (equipSlot) => this.net.sendWorld("UNEQUIP_ITEM", { equipSlot });
    this.characterPanel.onOpenTree = () => {
      this.characterPanel.close();
      this.skillTree.open(this.progression);
    };
    this.characterPanel.onOpenForms = () => {
      this.characterPanel.close();
      this.formWheel.open(this.progression);
    };
    this.characterPanel.onOpenBuilds = () => {
      this.characterPanel.close();
      this.buildUI.open(this.progression);
    };
    this.characterPanel.onSetStyle = (styleId) => this.net.sendWorld("SET_COMBAT_STYLE", { styleId });
    this.characterPanel.attributes.onAllocate = (stat) => this.net.sendWorld("ALLOCATE_ATTRIBUTE", { stat });
    this.characterPanel.attributes.onReset = () => this.net.sendWorld("RESET_ATTRIBUTES", {});
    this.skillTree.onUnlock = (nodeId, skillId) => this.net.sendWorld("UNLOCK_SKILL", { nodeId, skillId });
    this.skillTree.onReset = () => this.net.sendWorld("RESET_SKILLS", {});
    this.buildUI.onSave = (slot, name) => this.net.sendWorld("SAVE_BUILD", { slot, name });
    this.buildUI.onLoad = (slot) => this.net.sendWorld("LOAD_BUILD", { slot });
    this.formWheel.onSelect = (formId) => {
      this.progression.selectedForm = formId;
      if (formId === "normal") this.net.sendWorld("REQUEST_TRANSFORMATION", { formId: "normal" });
    };
    this.partyUI.onLeave = () => this.net.sendWorld("PARTY_LEAVE", {});
    this.partyUI.onKick = (id) => this.net.sendWorld("PARTY_KICK", { targetId: id });
    this.partyUI.onDisband = () => this.net.sendWorld("PARTY_DISBAND", {});
    this.partyUI.onTransfer = (id) => this.net.sendWorld("PARTY_TRANSFER", { targetId: id });
    this.partyUI.onRole = (role) => this.net.sendWorld("PARTY_SET_ROLE", { role });
    this.partyUI.onReady = () => this.net.sendWorld("PARTY_READY", {});
    this.partyUI.onFollow = () => this.net.sendWorld("FOLLOW_PARTY", {});
    this.partyUI.onWaypoint = (id) => this.net.sendWorld("PARTY_SET_WAYPOINT", { landmarkId: id });
    this.partyUI.onFinderCreate = (activity, role, minLevel) => this.net.sendWorld("PARTY_FINDER_CREATE", { activity, role, minLevel });
    this.partyUI.onFinderJoin = (id) => this.net.sendWorld("PARTY_FINDER_JOIN", { listingId: id });
    this.partyUI.onFinderList = () => this.net.sendWorld("PARTY_FINDER_LIST", {});
    this.partyUI.onCreate = () => this.net.sendWorld("PARTY_CREATE", {});
    this.socialUI.onInvite = (id) => this.net.sendWorld("PARTY_INVITE", { targetId: id });
    this.socialUI.onFriend = (id) => this.net.sendWorld("FRIEND_REQUEST", { targetId: id });
    this.socialUI.onInspect = (id) => this.net.sendWorld("INSPECT_PLAYER", { targetId: id });
    this.socialUI.onAcceptFriend = (id) => this.net.sendWorld("ACCEPT_FRIEND", { targetId: id });
    this.socialUI.onDeclineFriend = (id) => this.net.sendWorld("DECLINE_FRIEND", { targetId: id });
    this.socialUI.onRemoveFriend = (id) => this.net.sendWorld("REMOVE_FRIEND", { targetId: id });
    this.socialUI.onBlock = (id) => this.net.sendWorld("BLOCK_PLAYER", { targetId: id });
    this.socialUI.onUnblock = (id) => this.net.sendWorld("UNBLOCK_PLAYER", { targetId: id });
    this.socialUI.onLeaveParty = () => this.net.sendWorld("PARTY_LEAVE", {});
    this.socialUI.onRefresh = () => {
      this.net.sendWorld("GET_SOCIAL", {});
      if (this.socialUI.tab === "MARKET") this.net.sendWorld("GET_MARKET", {});
      if (this.socialUI.tab === "HOUSING") this.net.sendWorld("GET_HOUSE", {});
      if (this.socialUI.tab === "GUILD") this.net.sendWorld("GET_GUILD", {});
    };
    this.socialUI.onMessage = (_id, name) => {
      this.chatUI.channel = "WHISPER";
      this.chatUI.whisperTo = name;
      this.chatUI.focus();
      this.socialUI.close();
    };
    this.socialUI.onTrade = (id) => this.net.sendWorld("TRADE_REQUEST", { targetId: id });
    this.socialUI.onSearch = (name) => this.net.sendWorld("SEARCH_PLAYER", { name });
    this.socialUI.onReport = (id, category) => this.net.sendWorld("REPORT_PLAYER", { targetId: id, category, evidence: "ui" });
    this.socialUI.onCard = (id) => this.net.sendWorld("GET_PLAYER_CARD", { targetId: id });
    this.socialUI.onPartyReady = () => this.net.sendWorld("PARTY_READY", {});
    this.socialUI.onPartyCreate = () => this.net.sendWorld("PARTY_CREATE", {});
    this.socialUI.onGuildHall = () => this.net.sendWorld("GUILD_HALL_ENTER", {});
    this.socialUI.onPrivacy = (key, value) => this.net.sendWorld("SET_PRIVACY", { [key]: value });
    this.socialUI.onMarketBuy = (id) => this.net.sendWorld("MARKET_BUY", { listingId: id, transactionId: `mkt-${Date.now()}` });
    this.socialUI.onMarketCancel = (id) => this.net.sendWorld("MARKET_CANCEL", { listingId: id });
    this.socialUI.onMarketList = (slot) => {
      const item = this.loadout.slot(slot);
      const price = Math.max(1, item.item?.value ?? 10);
      this.net.sendWorld("MARKET_LIST", { slot, itemId: item.item?.id, qty: Math.min(item.qty, 1) || 1, price, transactionId: `lst-${Date.now()}` });
    };
    this.socialUI.onMarketRefresh = () => this.net.sendWorld("GET_MARKET", {});
    this.socialUI.onHouseEnter = () => this.net.sendWorld("HOUSE_ENTER", {});
    this.socialUI.onHouseLeave = () => this.net.sendWorld("HOUSE_LEAVE", {});
    this.socialUI.onHousePlace = (slot) => {
      const x = Math.round(this.player.position.x / 0.25) * 0.25;
      const z = Math.round(this.player.position.z / 0.25) * 0.25;
      this.net.sendWorld("HOUSE_PLACE", { slot, x, z, yaw: this.player.mesh.rotation.y, transactionId: `hp_${Date.now()}` });
    };
    this.socialUI.onHouseRemove = (id) => this.net.sendWorld("HOUSE_REMOVE", { decorId: id });
    this.socialUI.onHouseAccess = (access) => this.net.sendWorld("SET_HOUSE_ACCESS", { access });
    this.socialUI.onHouseCmd = (cmd, id) => this.handleHouseCmd(cmd, id);
    this.socialUI.onEmote = (emote) => this.net.sendWorld("SOCIAL_EMOTE", { emote });
    this.socialUI.onGuildDeposit = (slot) => this.net.sendWorld("GUILD_DEPOSIT", { slot, qty: 1 });
    this.socialUI.onGuildWithdraw = (slot) => this.net.sendWorld("GUILD_WITHDRAW", { slot, qty: 1 });
    this.socialUI.card.onInspect = (id) => this.net.sendWorld("INSPECT_PLAYER", { targetId: id });
    this.socialUI.card.onInvite = (id) => this.net.sendWorld("PARTY_INVITE", { targetId: id });
    this.socialUI.card.onFriend = (id) => this.net.sendWorld("FRIEND_REQUEST", { targetId: id });
    this.socialUI.card.onWhisper = (_id, name) => {
      this.chatUI.channel = "WHISPER";
      this.chatUI.whisperTo = name;
      this.chatUI.focus();
      this.socialUI.close();
    };
    this.socialUI.card.onTrade = (id) => this.net.sendWorld("TRADE_REQUEST", { targetId: id });
    this.socialUI.card.onReport = (id, category) => this.net.sendWorld("REPORT_PLAYER", { targetId: id, category, evidence: "card" });
    this.socialUI.card.onBlock = (id) => this.net.sendWorld("BLOCK_PLAYER", { targetId: id });
    this.chatUI.onSend = (channel, text, target) => this.net.sendWorld("CHAT", { channel, text, target });
    this.chatUI.onReport = (id, category) => this.net.sendWorld("REPORT_MESSAGE", { targetId: id, category, evidence: "chat" });
    this.chatUI.onMute = (id) => this.net.sendWorld("LOCAL_MUTE", { targetId: id });
    this.guildUI.onCreate = (name, tag) => this.net.sendWorld("GUILD_CREATE", { name, tag });
    this.guildUI.onInvite = (id) => this.net.sendWorld("GUILD_INVITE", { targetId: id });
    this.guildUI.onLeave = () => this.net.sendWorld("GUILD_LEAVE", {});
    this.guildUI.onDisband = () => this.net.sendWorld("GUILD_DISBAND", {});
    this.guildUI.onAnnounce = (text) => this.net.sendWorld("GUILD_ANNOUNCE", { text });
    this.guildUI.onRefresh = () => this.net.sendWorld("GET_GUILD", {});
    this.guildUI.onHall = () => this.net.sendWorld("GUILD_HALL_ENTER", {});
    this.guildUI.onContribute = () => this.net.sendWorld("GUILD_CONTRIBUTE", { itemId: "mat-iron-ore", qty: 1 });
    this.tradeUI.onReady = () => this.net.sendWorld("TRADE_READY", {});
    this.tradeUI.onConfirm = (tx) => this.net.sendWorld("TRADE_CONFIRM", { transactionId: tx });
    this.tradeUI.onCancel = () => this.net.sendWorld("TRADE_CANCEL", {});
    this.tradeUI.onOffer = (slots, coin) => this.net.sendWorld("TRADE_OFFER", { slots, coin });
    this.shopUI.onBuy = (shopItemId, itemId, tx) => this.net.sendWorld("SHOP_BUY", { shopItemId, itemId, transactionId: tx });
    this.shopUI.onSell = (slot, itemId, qty) => this.net.sendWorld("SHOP_SELL", { slot, itemId, qty });
    this.craftingUI.onCraft = (recipeId, stationId) => this.net.sendWorld("CRAFT", { recipeId, stationId, transactionId: `cr_${Date.now()}` });
    this.craftingUI.onProfession = (id) => this.net.sendWorld("SET_PROFESSION", { professionId: id });
    this.craftingUI.onReset = () => this.net.sendWorld("RESET_PROFESSION", {});
    this.craftingUI.onShopBuy = (shopId, itemId) => {
      this.combat.audio.buy();
      this.net.sendWorld("NPC_SHOP_BUY", { shopId, itemId, transactionId: `nb_${Date.now()}` });
    };
    this.craftingUI.onShopSell = (shopId, itemId) => {
      this.combat.audio.sell();
      this.net.sendWorld("NPC_SHOP_SELL", { shopId, itemId, qty: 1, transactionId: `ns_${Date.now()}` });
    };
    this.craftingUI.onContribute = () => this.net.sendWorld("GUILD_CONTRIBUTE", { itemId: "mat-iron-ore", qty: 1 });
    this.craftingUI.onStallList = (itemId, price) => this.net.sendWorld("STALL_LIST", { itemId, qty: 1, price });
    this.craftingUI.onStallBuy = (sellerId, itemId) => this.net.sendWorld("STALL_BUY", { sellerId, itemId, transactionId: `st_${Date.now()}` });
    this.craftingUI.onWorkshop = () => this.net.sendWorld("GET_WORKSHOP", {});
    this.hud.onCraft = () => this.toggleCrafting();
    this.hud.onLife = () => this.toggleLife();
    this.hud.onHome = () => this.toggleHome();
    this.hud.onPhoto = () => this.togglePhoto();
    this.lifeUI.onRefresh = () => this.net.sendWorld("GET_LIFE", {});
    this.lifeUI.onClaimPet = (id) => this.net.sendWorld("PET_CLAIM", { petId: id, transactionId: `pet_${Date.now()}` });
    this.lifeUI.onSummon = (id) => this.net.sendWorld("PET_SUMMON", { petId: id });
    this.lifeUI.onCare = (id, action) => this.net.sendWorld("PET_CARE", { petId: id, action });
    this.lifeUI.onDaily = (id) => this.net.sendWorld("CLAIM_DAILY_LIFE", { questId: id });
    this.lifeUI.onCollect = (id) => this.net.sendWorld("CLAIM_COLLECTION", { id, transactionId: `col_${Date.now()}` });
    this.lifeUI.onQuiz = (choice) => this.net.sendWorld("LIFE_QUIZ", { questionId: "q-kursi-4-2", choice });
    this.dungeonUI.onEnter = (id, difficulty) => this.net.sendWorld("DUNGEON_ENTER", { dungeonId: id, difficulty });
    this.dungeonUI.onQueue = (id, role, difficulty) => this.net.sendWorld("QUEUE_JOIN", { dungeonId: id, role, difficulty });
    this.dungeonUI.onQueueLeave = () => this.net.sendWorld("QUEUE_LEAVE", {});
    this.dungeonUI.onReady = (ready) => this.net.sendWorld("DUNGEON_READY", { ready });
    this.dungeonUI.onLeave = () => this.net.sendWorld("DUNGEON_LEAVE", {});
    this.dungeonUI.onRetry = () => this.net.sendWorld("DUNGEON_RETRY", {});
    this.dungeonUI.onRevive = (id) => this.net.sendWorld("DUNGEON_REVIVE", { targetId: id });
    this.dungeonUI.onVote = (vote) => this.net.sendWorld("DUNGEON_VOTE", { vote });
    this.dungeonUI.onSkip = () => this.net.sendWorld("SKIP_DUNGEON_INTRO", {});
    this.dungeonUI.rewards.onClaim = (claimId) => this.net.sendWorld("CLAIM_LOOT", { claimId });
    this.dungeonUI.onExchange = (id) => this.net.sendWorld("RAID_EXCHANGE", { shopItemId: id });
    this.pvpUI.onQueue = (mode) => this.net.sendWorld("PVP_QUEUE_JOIN", { mode });
    this.pvpUI.onLeaveQueue = () => this.net.sendWorld("PVP_QUEUE_LEAVE", {});
    this.pvpUI.onReady = (ok) => this.net.sendWorld(ok ? "PVP_READY" : "PVP_DECLINE", { ready: ok });
    this.pvpUI.onShop = (id) => this.net.sendWorld("PVP_SHOP_BUY", { shopItemId: id, transactionId: `pvp_${Date.now()}` });
    this.pvpUI.onBoard = (board) => this.net.sendWorld("PVP_LEADERBOARD", { board });
    this.pvpUI.onEmote = (emote) => this.net.sendWorld("PVP_EMOTE", { emote });
    this.pvpUI.onTraining = () => this.net.sendWorld("PVP_TRAINING", {});
    this.pvpUI.onDuel = (id) => this.net.sendWorld("PVP_DUEL", { targetId: id });
    this.pvpUI.onDuelAccept = (ok) => this.net.sendWorld(ok ? "PVP_DUEL_ACCEPT" : "PVP_DUEL_DECLINE", {});
    this.endgameUI.onClaimDaily = (id) => this.net.sendWorld("CLAIM_DAILY", { id, transactionId: `egd_${Date.now()}` });
    this.endgameUI.onClaimWeekly = (id) => this.net.sendWorld("CLAIM_WEEKLY", { id, transactionId: `egw_${Date.now()}` });
    this.endgameUI.onClaimChallenge = (id) => this.net.sendWorld("CLAIM_CHALLENGE", { id, transactionId: `egc_${Date.now()}` });
    this.endgameUI.onClaimSeason = (level) => this.net.sendWorld("CLAIM_SEASON", { level, transactionId: `egs_${Date.now()}` });
    this.hud.onEndgame = () => {
      this.endgameUI.toggle();
      if (this.endgameUI.blocking) this.net.sendWorld("GET_ENDGAME", {});
    };
    this.hud.onDungeon = () => {
      if (this.dungeonUI.overlayOpen) this.dungeonUI.hideOverlay();
      else this.net.sendWorld("GET_DUNGEONS", {});
    };
    this.hud.onAccessChange = (vfx, shake, numbers) => {
      this.bossUI.telegraph.vfxOn = vfx;
      this.combat.feedback.shakeOn = shake;
      this.combat.feedback.numbersOn = numbers;
    };
    this.hud.onGraphicsChange = (preset) => this.applyGraphics(preset);
    this.hud.onTextSize = (size) => {
      document.documentElement.dataset.text = size;
    };
    this.hud.onSubtitle = (on) => this.dialogue.setSubtitle(on);
    this.hud.onAudioChange = (master, music, sfx) => {
      this.combat.audio.master = master;
      this.combat.audio.sfx = sfx;
      this.music.setVolume(master, music);
    };
    this.chapterUI.onEnter = (id) => {
      this.chapterUI.close();
      this.net.sendWorld("DUNGEON_ENTER", { dungeonId: id });
    };
  }

  private applyGraphics(preset: GraphicsPreset): void {
    const quality = resolveGraphics(preset);
    const g = GRAPHICS[quality];
    this.renderer.shadowMap.enabled = g.shadow;
    this.renderer.toneMappingExposure = g.bloom ? 1.12 : 1.02;
    this.world.environment.sun.castShadow = g.shadow;
    this.world.environment.sun.shadow.mapSize.set(g.shadowSize, g.shadowSize);
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, quality === "LOW" ? 1 : quality === "ULTRA" ? 2 : 1.5));
    const fog = this.world.scene.fog;
    if (fog instanceof THREE.Fog) {
      fog.near = quality === "LOW" ? 18 : 28;
      fog.far = quality === "LOW" ? 70 : quality === "MEDIUM" ? 95 : 130;
    }
    this.world.environment.sun.shadow.radius = quality === "ULTRA" || quality === "HIGH" ? 2.2 : 1;
    this.world.weather.apply(this.world.scene, this.world.weather.weather);
    this.mountSys.apply(this.mountSys.mounted, this.mountSys.mountId, quality);
    this.petFollow.apply(this.petFollow.petId, quality);
    this.hud.toast(`GRAPHICS ${preset}${preset === "AUTO" ? ` → ${quality}` : ""}`);
  }

  private toggleInventory(): void {
    if (this.inventoryUI.blocking) this.inventoryUI.close();
    else {
      this.characterPanel.close();
      this.inventoryUI.open();
      this.net.sendWorld("GET_INVENTORY", {});
    }
  }

  private toggleCharacter(): void {
    if (this.characterPanel.blocking) this.characterPanel.close();
    else {
      this.inventoryUI.close();
      this.characterPanel.open(this.loadout, this.progression, this.player.stats.name);
    }
  }

  private toggleCrafting(): void {
    if (this.craftingUI.blocking) this.craftingUI.close();
    else {
      this.craftingUI.open();
      this.net.sendWorld("GET_CRAFTING", {});
    }
  }

  private toggleLife(): void {
    if (this.lifeUI.blocking) this.lifeUI.close();
    else {
      this.lifeUI.open();
      this.net.sendWorld("GET_LIFE", {});
    }
  }

  private toggleHome(): void {
    this.socialUI.open("HOUSING");
    this.net.sendWorld("GET_HOUSE", {});
  }

  private togglePhoto(): void {
    this.photoMode = !this.photoMode;
    this.hud.setPhotoMode(this.photoMode);
    this.mobile.setVisible(!this.photoMode && this.touchUI);
    if (this.photoMode) this.hud.toast("PHOTO MODE");
  }

  private handleDecorate(): void {
    if (!this.decorateOn) return;
    const items = this.socialUI.house?.items ?? [];
    if (!this.selectedDecor && items[0]) this.selectedDecor = items[0].id;
    if (this.input.consumeDecorateRotate() && this.selectedDecor) {
      const it = items.find((x) => x.id === this.selectedDecor);
      this.net.sendWorld("HOUSE_MOVE", { decorId: this.selectedDecor, x: it?.x ?? 0, z: it?.z ?? 0, yaw: (it?.yaw ?? 0) + Math.PI / 2, transactionId: `mv_${Date.now()}` });
    }
    if (this.input.consumeDecorateRemove() && this.selectedDecor) {
      this.net.sendWorld("HOUSE_REMOVE", { decorId: this.selectedDecor });
      this.selectedDecor = "";
    }
  }

  private handleHouseCmd(cmd: string, id: string): void {
    const house = this.socialUI.house;
    if (cmd === "lock") this.net.sendWorld("HOUSE_LOCK", { on: !house?.locked });
    if (cmd === "decorate") {
      this.decorateOn = !this.decorateOn;
      this.input.decorateMode = this.decorateOn;
      this.net.sendWorld("HOUSE_DECORATE", { on: this.decorateOn });
      this.hud.toast(this.decorateOn ? "DECORATE ON · R rotate · DEL remove" : "DECORATE OFF");
    }
    if (cmd === "hall") this.net.sendWorld("GUILD_HALL_ENTER", {});
    if (cmd === "plant") this.net.sendWorld("GARDEN_PLANT", { plotId: id, plantId: "dawn-berry", transactionId: `gp_${Date.now()}` });
    if (cmd === "water") this.net.sendWorld("GARDEN_WATER", { plotId: id });
    if (cmd === "harvest") this.net.sendWorld("GARDEN_HARVEST", { plotId: id, transactionId: `gh_${Date.now()}` });
    if (cmd === "style") {
      const [k, v] = id.split(":");
      this.net.sendWorld("HOUSE_STYLE", { [k]: v });
    }
    if (cmd === "vote") {
      const owner = house?.ownerId && house.ownerId !== this.net.playerId ? house.ownerId : "";
      if (owner) this.net.sendWorld("HOUSE_VOTE", { ownerId: owner, category: id, score: 5 });
    }
  }

  private syncRemotePets(): void {
    const seen = new Set<string>();
    for (const remote of this.remotes.all()) {
      let pet = this.remotePets.get(remote.id);
      if (!pet) {
        pet = new PetFollower(remote.group);
        this.remotePets.set(remote.id, pet);
      }
      seen.add(remote.id);
      pet.apply(this.lastPetIds.get(remote.id) || "", resolveGraphics(this.hud.graphics));
      pet.update(0.016);
    }
    for (const id of [...this.remotePets.keys()]) {
      if (!seen.has(id)) this.remotePets.delete(id);
    }
  }

  private rpgBlocking(): boolean {
    return this.inventoryUI.blocking || this.characterPanel.blocking || this.skillTree.blocking || this.buildUI.blocking || this.formWheel.blocking || this.partyUI.blocking || this.socialUI.overlayOpen || this.inviteUI.blocking || this.dungeonUI.blocking || this.pvpUI.blocking || this.endgameUI.blocking || this.chapterUI.blocking || this.journalUI.blocking || this.worldMap.blocking || this.travel.blocking || this.mountCol.blocking || this.guildUI.blocking || this.tradeUI.blocking || this.shopUI.blocking || this.craftingUI.blocking || this.fishingUI.blocking || this.lifeUI.blocking || this.cinematicUI.blocking;
  }

  private closeRpg(): boolean {
    if (this.cinematicUI.blocking) {
      this.cinematicUI.skip();
      return true;
    }
    if (this.dialogue.blocking) {
      this.dialogue.close();
      return true;
    }
    if (this.inviteUI.blocking) {
      this.inviteUI.close();
      return true;
    }
    if (this.socialUI.inspect.blocking) {
      this.socialUI.inspect.close();
      return true;
    }
    if (this.chatUI.focused) {
      this.chatUI.blur();
      return true;
    }
    if (this.guildUI.blocking) {
      this.guildUI.close();
      return true;
    }
    if (this.tradeUI.blocking) {
      this.tradeUI.close();
      return true;
    }
    if (this.shopUI.blocking) {
      this.shopUI.close();
      return true;
    }
    if (this.craftingUI.blocking) {
      this.craftingUI.close();
      return true;
    }
    if (this.lifeUI.blocking) {
      this.lifeUI.close();
      return true;
    }
    if (this.fishingUI.blocking) {
      this.fishingUI.close();
      return true;
    }
    if (this.inventoryUI.blocking) {
      this.inventoryUI.close();
      return true;
    }
    if (this.characterPanel.blocking) {
      this.characterPanel.close();
      return true;
    }
    if (this.skillTree.blocking) {
      this.skillTree.close();
      return true;
    }
    if (this.buildUI.blocking) {
      this.buildUI.close();
      return true;
    }
    if (this.formWheel.blocking) {
      this.formWheel.close();
      return true;
    }
    if (this.partyUI.blocking) {
      this.partyUI.close();
      return true;
    }
    if (this.socialUI.blocking) {
      this.socialUI.close();
      return true;
    }
    if (this.chapterUI.blocking) {
      this.chapterUI.close();
      return true;
    }
    if (this.journalUI.blocking) {
      this.journalUI.close();
      return true;
    }
    if (this.worldMap.blocking) {
      this.worldMap.close();
      return true;
    }
    if (this.travel.blocking) {
      this.travel.close();
      return true;
    }
    if (this.mountCol.blocking) {
      this.mountCol.close();
      return true;
    }
    if (this.dungeonUI.overlayOpen && this.dungeonUI.view?.state !== "ACTIVE" && this.dungeonUI.view?.state !== "BOSS") {
      this.dungeonUI.hideOverlay();
      return true;
    }
    if (this.pvpUI.blocking) {
      this.pvpUI.close();
      return true;
    }
    if (this.endgameUI.blocking) {
      this.endgameUI.close();
      return true;
    }
    return false;
  }

  private handleDungeon(type: NetMsgType, data: unknown): boolean {
    if (type === "DUNGEON_OFFER") {
      this.dialogue.close();
      this.dungeonUI.showOffer(data as DungeonOffer);
      return true;
    }
    if (type === "DUNGEON_LIST") {
      this.dungeonUI.showList(data as DungeonListOut);
      return true;
    }
    if (type === "QUEUE_UPDATE") {
      this.dungeonUI.showQueue(data as QueueView);
      return true;
    }
    if (type === "DUNGEON_READY_CHECK") {
      this.dungeonUI.showReady(data as DungeonReadyOut);
      return true;
    }
    if (type === "DUNGEON_LOADING") {
      this.dungeonUI.showLoading((data as DungeonLoading).name);
      return true;
    }
    if (type === "DUNGEON_STARTED" || type === "DUNGEON_STATE" || type === "DUNGEON_WAVE" || type === "DUNGEON_OBJECTIVE") {
      const view = data as DungeonView;
      this.inDungeon = true;
      this.dungeonUI.selfId = this.net.playerId;
      this.world.setDungeon(true);
      this.dungeonUI.applyState(view, this.player.position.x, this.player.position.z);
      this.bossUI.apply(view);
      return true;
    }
    if (type === "DUNGEON_COMPLETE") {
      this.dungeonUI.showComplete(data as DungeonView);
      this.bossUI.clear();
      this.hud.toast((data as DungeonView).kind === "RAID" ? "RAID CLEAR" : "DUNGEON COMPLETE");
      return true;
    }
    if (type === "DUNGEON_FAILED") {
      this.dungeonUI.showFail(data as DungeonView);
      this.bossUI.clear();
      return true;
    }
    if (type === "DUNGEON_WIPE") {
      this.dungeonUI.showWipe(data as DungeonView);
      this.hud.toast("Party wipe — checkpoint");
      return true;
    }
    if (type === "DUNGEON_LEFT") {
      this.inDungeon = false;
      this.world.setDungeon(false);
      this.dungeonUI.clear();
      this.bossUI.clear();
      return true;
    }
    if (type === "BOSS_SPAWN" || type === "BOSS_ENRAGE" || type === "BOSS_RESET") {
      this.bossUI.apply(data as DungeonView);
      this.dungeonUI.applyState(data as DungeonView, this.player.position.x, this.player.position.z);
      if (type === "BOSS_ENRAGE") this.hud.toast("ENRAGED +25%");
      return true;
    }
    if (type === "BOSS_PHASE") {
      this.hud.toast((data as BossPhaseOut).label || "PHASE");
      return true;
    }
    if (type === "BOSS_TELEGRAPH") {
      this.bossUI.onTelegraph(data as BossTelegraphOut);
      return true;
    }
    if (type === "BOSS_AOE") {
      this.bossUI.onAoe(data as BossAOEOut);
      return true;
    }
    if (type === "BOSS_INTERRUPT") {
      this.bossUI.telegraph.clear();
      this.hud.toast("Interrupted");
      return true;
    }
    if (type === "BOSS_LOCK") {
      this.hud.toast("Boss lock — join ditutup");
      return true;
    }
    if (type === "PLAYER_DOWNED") {
      this.hud.toast("DOWNED — minta revive");
      this.dungeonUI.updateRevive();
      return true;
    }
    if (type === "PLAYER_REVIVED") {
      this.hud.toast("Revived");
      this.dungeonUI.updateRevive();
      return true;
    }
    if (type === "DUNGEON_VOTE_UPDATE") {
      this.hud.toast("Vote dungeon");
      return true;
    }
    if (type === "BOSS_DEFEATED") {
      const name = String((data as { bossId?: string }).bossId ?? "Boss");
      this.hud.toast(`${name} dikalahkan.`);
      return true;
    }
    if (type === "LOOT_RESULT") {
      this.dungeonUI.rewards.show(data as LootResult);
      return true;
    }
    if (type === "CHAPTER_LIST") {
      this.chapters = (data as ChapterListOut).chapters ?? [];
      if (this.chapterUI.blocking) this.chapterUI.open(this.chapters);
      return true;
    }
    if (type === "SAVE_OK") {
      const name = String((data as { checkpoint?: string }).checkpoint || "Checkpoint");
      this.hud.toast(`Tersimpan · ${name}`);
      return true;
    }
    return false;
  }

  private handlePvp(type: NetMsgType, data: unknown): boolean {
    if (type === "PVP_LOBBY") {
      this.pvpUI.applyLobby(data as PvpLobbyOut);
      if (!this.pvpUI.blocking) this.pvpUI.open();
      return true;
    }
    if (type === "PVP_QUEUE_UPDATE") {
      this.pvpUI.showQueue(data as PvpQueueView);
      const st = String((data as { state?: string }).state || "");
      this.hud.toast(st === "MATCH_FOUND" ? "MATCH FOUND" : "Matchmaking...");
      return true;
    }
    if (type === "PVP_READY_CHECK") {
      this.pvpUI.showReady(data as PvpReadyOut);
      return true;
    }
    if (type === "PVP_LOADING") {
      this.pvpUI.close();
      const map = String((data as { map?: string }).map || "");
      this.hud.toast("Loading " + pvpZoneName(map));
      return true;
    }
    if (type === "PVP_COUNTDOWN") {
      const fight = Boolean((data as { fight?: boolean }).fight);
      this.hud.toast(fight ? "FIGHT" : "3 · 2 · 1");
      return true;
    }
    if (type === "PVP_STATE") {
      this.pvpHud.apply(data as PvpView, this.net.playerId);
      this.syncPvpTargets();
      return true;
    }
    if (type === "PVP_KILL_FEED") {
      const line = String((data as { text?: string }).text || "defeated");
      this.hud.toast(line);
      return true;
    }
    if (type === "PVP_CAPTURE") {
      this.pvpHud.apply({ ...(this.pvpHud.view || { matchId: "", mode: "", map: "dawn-arena", state: "ACTIVE", timeLeft: 0, scoreA: 0, scoreB: 0, members: [], team: 1 }), ...(data as object) } as PvpView, this.net.playerId);
      return true;
    }
    if (type === "PVP_RESULT") {
      this.pvpHud.showResult(data as PvpResultOut);
      this.inPvp = false;
      this.world.setInstance("hub");
      this.combat.pvpTargets = [];
      return true;
    }
    if (type === "PVP_LEADERBOARD") {
      const entries = ((data as { entries?: LBEntry[] }).entries || []) as LBEntry[];
      if (!this.pvpUI.lobby) this.net.sendWorld("GET_PVP", {});
      if (this.pvpUI.lobby) this.pvpUI.lobby.leaderboard = entries;
      if (this.pvpUI.blocking) this.pvpUI.render();
      return true;
    }
    if (type === "PVP_AFK") {
      const msg = String((data as { text?: string; warning?: boolean }).text || "");
      this.hud.toast(msg || ((data as { warning?: boolean }).warning ? "Anda tidak aktif. Bergerak atau melakukan aksi untuk tetap berada dalam pertandingan." : "AFK penalty"));
      return true;
    }
    if (type === "PVP_LEFT") {
      this.pvpHud.clear();
      this.inPvp = false;
      this.world.setInstance("hub");
      return true;
    }
    if (type === "PVP_SPECTATE") {
      this.hud.toast("Spectate teammate");
      return true;
    }
    if (type === "PVP_EMOTE") {
      return true;
    }
    if (type === "PVP_DUEL_REQUEST") {
      const d = data as { from?: string; level?: number };
      this.pvpUI.showDuelRequest(d.from || "Player", d.level || 1);
      return true;
    }
    if (type === "PVP_TRAINING") {
      this.hud.toast("Training Arena");
      this.pvpUI.close();
      return true;
    }
    return false;
  }

  private syncPvpTargets(): void {
    if (!this.inPvp) {
      this.combat.pvpTargets = [];
      return;
    }
    const live = this.remotes.all();
    const next: PvpTarget[] = [];
    for (const remote of live) {
      let lock = this.pvpLocks.get(remote.id);
      if (!lock) {
        lock = new PvpTarget(remote);
        this.pvpLocks.set(remote.id, lock);
      }
      lock.sync();
      next.push(lock);
    }
    this.combat.pvpTargets = next;
  }

  private handleProgression(type: NetMsgType, data: unknown): boolean {
    if (type === "PROGRESSION_STATE") {
      this.applyProgression(data as ProgressionView);
      return true;
    }
    if (type === "TRANSFORMATION_STARTED" || type === "TRANSFORMATION_UPDATED") {
      const view = data as TransformView;
      this.progression.applyTransform(view);
      this.applyFormVisual(view.formId, view.visual);
      if (type === "TRANSFORMATION_STARTED" && view.playerId === this.net.playerId) {
        this.cameraRig.beginCinematic("zoom");
        this.transformCam = 1.85;
        this.cameraRig.pulseZoom();
        this.cameraRig.pulseFov();
        this.loop.addHitStop(0.55);
        this.transformFx?.burst(this.player.position, view);
        this.combat.audio.transform();
        this.hud.toast(view.name || "ASCEND");
      }
      return true;
    }
    if (type === "TRANSFORMATION_ENDED") {
      const view = data as TransformView;
      this.progression.applyTransform({ ...view, formId: "normal", state: "COOLDOWN" });
      this.applyFormVisual("normal", "asal");
      if (view.playerId === this.net.playerId) this.hud.toast("Transformation ended");
      return true;
    }
    if (type === "TRANSFORMATION_REJECTED") {
      const reason = String((data as { reason?: string }).reason ?? "locked");
      this.hud.toast(rejectText(reason));
      return true;
    }
    if (type === "SKILL_UNLOCKED") {
      this.hud.toast("Skill unlocked");
      this.net.sendWorld("GET_PROGRESSION", {});
      return true;
    }
    if (type === "SKILL_USED") {
      const skillId = String((data as { skillId?: string }).skillId ?? "");
      this.skillFlash.flash(skillId.replaceAll("_", " "));
      return true;
    }
    if (type === "POWER_RATING_UPDATED") {
      return true;
    }
    if (type === "TRAINING_METER") {
      const dps = Number((data as { dps?: number }).dps ?? 0);
      this.trainingDps = dps;
      this.hud.toast(`DPS ${dps}`);
      return true;
    }
    if (type === "BUILD_LIST" || type === "BUILD_SAVED" || type === "BUILD_LOADED") {
      this.net.sendWorld("GET_PROGRESSION", {});
      this.hud.toast(type === "BUILD_LOADED" ? "Build loaded" : "Build saved");
      if (this.buildUI.blocking) this.buildUI.render(this.progression);
      return true;
    }
    return false;
  }

  private applyProgression(view: ProgressionView): void {
    this.progression.apply(view);
    this.applyFormVisual(view.formId, view.forms.find((f) => f.id === view.formId)?.visual);
    this.characterPanel.render(this.loadout, this.progression, this.player.stats.name);
    if (this.skillTree.blocking) this.skillTree.render(this.progression);
    if (this.buildUI.blocking) this.buildUI.render(this.progression);
    if (this.formWheel.blocking) this.formWheel.render(this.progression);
  }

  private applyFormVisual(formId?: string, visual?: string): void {
    this.player.setForm(visualForm(formId, visual));
    this.refreshEquipmentVisual();
  }

  private applyLoadout(upd: InventoryUpdated): void {
    this.loadout.applyLoadout(upd);
    this.applyRpgStats();
    this.quests.patchWallet(this.loadout.coin, this.loadout.crystal, this.loadout.eduToken);
    this.refreshEquipmentVisual();
    this.inventoryUI.render();
    this.characterPanel.render(this.loadout, this.progression, this.player.stats.name);
    for (const note of this.loadout.takeNotes()) this.hud.toast(note);
  }

  private applyRpgStats(): void {
    const s = this.loadout.stats;
    if (!s) return;
    this.player.stats.level = s.level;
    this.player.stats.health.maxHp = s.maxHp;
    this.player.stats.health.hp = s.hp;
    this.player.stats.energy = s.energy;
    this.player.stats.maxEnergy = s.maxEnergy;
    this.player.stats.stamina = s.stamina;
    this.player.stats.attack = s.attack;
    this.player.stats.defense = s.defense;
    this.player.stats.strength = s.strength;
    this.player.stats.agility = s.agility;
    this.player.stats.energyPower = s.energyPower;
    this.player.stats.criticalChance = s.criticalChance;
    this.player.className = s.class || this.player.className;
  }

  private refreshEquipmentVisual(): void {
    this.equipmentVisual.apply(this.player, this.loadout.equipment);
  }

  private onDialogOption(id: string, res: InteractResult): void {
    if (id === "close") {
      this.dialogue.close();
      return;
    }
    const [kind, questId] = id.split(":");
    if (kind === "accept" && questId) this.net.sendWorld("QUEST_ACCEPT", { questId });
    else if (kind === "decline" && questId) this.net.sendWorld("QUEST_DECLINE", { questId });
    else if (kind === "claim" && questId) this.net.sendWorld("QUEST_CLAIM", { questId });
    else if (id === "heal") this.net.sendWorld("HEAL", {});
    else if (id === "shop") this.net.sendWorld("SHOP_OPEN", {});
    else if (id === "craft-ui") {
      this.dialogue.close();
      this.toggleCrafting();
    }
    else if (id === "repair") this.net.sendWorld("NPC_REPAIR", { transactionId: `rp_${Date.now()}` });
    else if (id.startsWith("npc-shop:")) this.net.sendWorld("NPC_SHOP_OPEN", { shopId: id.slice(9) });
    else if (id === "quiz-prof") this.net.sendWorld("INTERACT", { targetId: res.targetId || "mbok_rasa", kind: "quiz-prof" });
    else if (id === "quiz-econ") this.net.sendWorld("INTERACT", { targetId: res.targetId || "pak_dagang", kind: "quiz-econ" });
    else if (id === "quiz-craft") this.net.sendWorld("INTERACT", { targetId: res.targetId || "pak_dagang", kind: "quiz-craft" });
    else if (id === "fish-ui") this.net.sendWorld("FISH_START", { spotId: "spot-village" });
    else if (id === "guild") {
      this.dialogue.close();
      this.guildUI.open();
    }
    else if (id === "finder") {
      this.dialogue.close();
      this.partyUI.open();
    }
    else if (id === "quiz-guru") this.net.sendWorld("INTERACT", { targetId: res.targetId || "mbah_guru", kind: "quiz-guru" });
    else if (id.startsWith("cin:")) this.net.sendWorld("INTERACT", { targetId: res.targetId || "mbah_karya", kind: id });
    else if (id.startsWith("choice") && (res.kind === "siluman" || res.targetId === "mbah_jagat")) this.net.sendWorld("STORY_CHOICE", { choiceId: res.kind === "siluman" ? `siluman:${res.targetId}` : res.targetId, option: id, targetId: res.targetId });
    else if (id.startsWith("choice")) this.net.sendWorld("INTERACT", { targetId: res.targetId || "mira", kind: id });
    else if (id === "quiz-apel") this.net.sendWorld("INTERACT", { targetId: "petani", kind: "quiz-apel" });
    else if (id === "quiz-latihan") this.net.sendWorld("INTERACT", { targetId: "mbah_jagat", kind: "quiz-latihan" });
    else if (id === "quiz") this.net.sendWorld("INTERACT", { targetId: "mira", kind: "quiz" });
    else if (id === "talk:forest") this.net.sendWorld("INTERACT", { targetId: res.targetId || "elder_ardan", kind: "forest" });
    else if (kind === "dungeon" && questId) this.net.sendWorld("DUNGEON_ENTER", { dungeonId: questId });
    else if (kind === "queue" && questId) this.net.sendWorld("QUEUE_JOIN", { dungeonId: questId, role: "FLEX" });
    else this.dialogue.close();
  }
}

function questNav(progress: { quests?: Array<{ state: string; objectives: Array<{ type: string; target: string; progress: number; count: number }> }> } | null | undefined): { x: number; z: number } | null {
  const q = progress?.quests?.find((item) => item.state === "ACTIVE");
  const obj = q?.objectives.find((o) => o.progress < o.count);
  if (!obj) return null;
  if (obj.type === "REACH") {
    const marks: Record<string, { x: number; z: number }> = {
      "village-gate": { x: 0, z: 19.6 },
      "forest-path": { x: 0, z: 34 },
      "valley-gate": { x: 0, z: 50 },
    };
    return marks[obj.target] ?? null;
  }
  if (obj.type === "VISIT" || obj.type === "DISCOVER") {
    const marks: Record<string, { x: number; z: number }> = {
      forest: { x: 0, z: 34 },
      valley: { x: 0, z: 56 },
      "edu-shrine": { x: -6, z: 33 },
      "mist-waterfall": { x: -11, z: 42 },
      village: { x: 0, z: 6 },
    };
    return marks[obj.target] ?? null;
  }
  if (obj.type === "TALK") {
    const npc: Record<string, { x: number; z: number }> = {
      mira: { x: -9.2, z: 2.2 },
      elder_ardan: { x: 0, z: 3.2 },
      raven: { x: 0, z: 19.6 },
      mbah_jagat: { x: 1.6, z: 4.2 },
      ibu_desa: { x: -1.6, z: 5.2 },
      petani: { x: 6.2, z: 6.4 },
      lio: { x: 9.4, z: 3.4 },
      mbah_karya: { x: 6.2, z: 4.6 },
      mbah_guru: { x: -8.6, z: 2.4 },
      laras: { x: 2.2, z: 6.2 },
      jaka: { x: -2.4, z: 5.6 },
      bagas: { x: 1.2, z: 8.4 },
      sari: { x: -3.6, z: 3.8 },
      wira: { x: 4.4, z: 7.2 },
      mbok_rasa: { x: -6.4, z: 4.8 },
      pak_jala: { x: -8.0, z: 7.4 },
      mbah_batu: { x: 8.0, z: 8.2 },
    };
    return npc[obj.target] ?? null;
  }
  return { x: 0, z: 20 };
}

function pvpZoneName(map: string): string {
  if (map === "dawn-arena" || map === "valley-of-dawn") return "Dawn Arena";
  if (map === "mistwood-battlefield") return "Mistwood Battlefield";
  return "Celestial Courtyard";
}

function rejectText(reason: string): string {
  if (reason === "not_leader") return "Hanya leader party yang boleh melakukan aksi ini.";
  if (reason === "not_available") return "Aksi belum tersedia.";
  if (reason === "blocked") return "Pemain diblokir.";
  if (reason === "spam") return "Chat dibatasi sejenak.";
  if (reason === "mute") return "Kamu sedang dimute.";
  if (reason === "coin") return "Koin tidak cukup.";
  if (reason === "ready") return "Kedua pemain harus READY.";
  if (reason === "lockout") return "Weekly lockout masih aktif.";
  if (reason === "size") return "Jumlah pemain belum memenuhi.";
  if (reason === "membership") return "Kamu bukan anggota instance ini.";
  if (reason === "boss_lock") return "Boss sudah dimulai. Join ditutup.";
  if (reason === "duplicate") return "Hadiah sudah diambil.";
  if (reason === "token") return "Revive token habis.";
  if (reason === "locked") return "Transformation locked.";
  if (reason === "energy") return "Energy tidak cukup.";
  if (reason === "cooldown") return "Transformation cooldown.";
  if (reason === "dead" || reason === "stunned" || reason === "restricted" || reason === "education") return "Tidak dapat transform sekarang.";
  if (reason === "level") return "Level belum cukup.";
  if (reason === "points") return "Poin tidak cukup.";
  if (reason === "prerequisite") return "Syarat skill belum terpenuhi.";
  if (reason === "server_authoritative") return "Aksi ditolak server.";
  if (reason === "capacity") return "Inventory penuh.";
  if (reason === "drop" || reason === "item") return "Item tidak valid.";
  if (reason === "distance") return "Terlalu jauh.";
  if (reason === "full") return "Party penuh.";
  if (reason === "already_in_party") return "Pemain sudah dalam party.";
  return reason ? `Ditolak: ${reason}` : "Aksi ditolak server.";
}

function cameraPresence(target: { health?: { maxHp?: number } } | null): "explore" | "combat" | "boss" {
  if (!target) return "explore";
  if ((target.health?.maxHp ?? 0) >= 400) return "boss";
  return "combat";
}

function gatherPrompt(kind: string, touch: boolean): string {
  if (kind === "npc") return touch ? "[TALK]" : "[TALK] E";
  if (kind.startsWith("gather-")) return touch ? "[GATHER]" : "[ E ] Gather";
  if (kind === "fishing-spot") return touch ? "[FISH]" : "[ E ] Fish";
  if (kind === "forge" || kind === "workbench" || kind === "cooking-fire" || kind === "alchemy") return touch ? "[CRAFT]" : "[ E ] Craft";
  return touch ? "[INTERACT]" : "[INTERACT] E";
}
