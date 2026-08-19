import * as THREE from "three";
import { Environment } from "./Environment";
import { TrainingGround } from "./TrainingGround";
import { NPCRenderer } from "./NPCRenderer";
import { WorldInteraction } from "./WorldInteraction";
import { ZoneManager } from "./ZoneManager";
import { LandmarkSystem } from "./LandmarkSystem";
import { WeatherSystem } from "./WeatherSystem";
import { WorldPath } from "./WorldPath";

export class World {
  readonly scene = new THREE.Scene();
  readonly hubGroup = new THREE.Group();
  readonly dungeonGroup = new THREE.Group();
  readonly environment: Environment;
  readonly training: TrainingGround;
  readonly npcs: NPCRenderer;
  readonly interact: WorldInteraction;
  readonly zones: ZoneManager;
  readonly landmarks: LandmarkSystem;
  readonly weather: WeatherSystem;
  readonly path: WorldPath;

  constructor() {
    this.scene.add(this.hubGroup, this.dungeonGroup);
    this.environment = new Environment(this.scene, this.hubGroup, this.dungeonGroup);
    this.environment.build();
    this.training = new TrainingGround(this.hubGroup);
    this.npcs = new NPCRenderer(this.scene);
    this.interact = new WorldInteraction(this.scene);
    this.zones = new ZoneManager(this.hubGroup);
    this.landmarks = new LandmarkSystem(this.hubGroup);
    this.path = new WorldPath(this.hubGroup);
    this.weather = new WeatherSystem();
  }

  setDungeon(on: boolean): void {
    this.environment.setDungeonMode(on);
  }

  setInstance(mode: "hub" | "dungeon" | "arena" | "bg"): void {
    this.environment.setInstanceMode(mode);
  }
}
