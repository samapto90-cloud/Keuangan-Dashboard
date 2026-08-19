import type { Player } from "../player/Player";
import type { RemotePlayer } from "../network/RemotePlayer";

export function applyCombatPose(player: Player, attackType?: string, skillId?: string): void {
  const kind = skillId ?? attackType ?? "";
  if (kind === "dead") player.combatPose = "DEAD";
  else if (kind === "dodge") player.combatPose = "DODGE";
  else if (kind === "kick" || kind === "whirlwind_kick") player.combatPose = "KICK";
  else if (kind === "energy_bolt" || kind === "energy") player.combatPose = "ENERGY_ATTACK";
  else if (kind === "punch" || kind === "power_strike" || kind === "skill") player.combatPose = "PUNCH";
  else player.combatPose = null;
}

export function applyRemoteCombat(remote: RemotePlayer, attackType?: string, skillId?: string): void {
  remote.playCombat(attackType ?? "", skillId);
}
