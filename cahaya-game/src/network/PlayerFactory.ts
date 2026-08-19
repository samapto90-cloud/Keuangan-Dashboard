import * as THREE from "three";
import { Player } from "../player/Player";
import type { PlayerSpawn } from "./NetworkMessage";
import { RemotePlayer } from "./RemotePlayer";

export class PlayerFactory {
  constructor(private readonly scene: THREE.Scene) {}

  createLocalPlayer(): Player {
    const player = new Player();
    this.scene.add(player.group);
    return player;
  }

  createRemotePlayer(spawn: PlayerSpawn): RemotePlayer {
    const remote = new RemotePlayer(spawn);
    this.scene.add(remote.group);
    return remote;
  }

  destroyRemotePlayer(remote: RemotePlayer): void {
    remote.dispose();
  }
}
