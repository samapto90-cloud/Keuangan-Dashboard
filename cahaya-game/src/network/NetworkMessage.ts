export const CHAT_CHANNEL = {
  LOCAL: "LOCAL",
  WORLD: "WORLD",
  PARTY: "PARTY",
  GUILD: "GUILD",
  RAID: "RAID",
  WHISPER: "WHISPER",
  SYSTEM: "SYSTEM",
} as const;

export type ChatChannel = (typeof CHAT_CHANNEL)[keyof typeof CHAT_CHANNEL];

export type NetMsgType =
  | "AUTH"
  | "AUTH_OK"
  | "AUTH_FAIL"
  | "JOIN_WORLD"
  | "WELCOME"
  | "PLAYER_SPAWN"
  | "PLAYER_DESPAWN"
  | "PLAYER_JOIN"
  | "PLAYER_LEAVE"
  | "MOVE_INPUT"
  | "WORLD_SNAPSHOT"
  | "ERROR"
  | "PING"
  | "PONG"
  | "PLAYER_ATTACK"
  | "PLAYER_SKILL"
  | "PLAYER_DODGE"
  | "PLAYER_COMBO"
  | "PLAYER_RESPAWN"
  | "ATTACK_RESULT"
  | "DAMAGE_RESULT"
  | "PLAYER_HIT"
  | "ENEMY_HIT"
  | "ENEMY_DEATH"
  | "ENEMY_SPAWN"
  | "PLAYER_DEATH"
  | "PLAYER_LEVEL_UP"
  | "ACTION_REJECT"
  | "INTERACT"
  | "INTERACT_RESULT"
  | "MANUAL_SAVE"
  | "SAVE_OK"
  | "QUEST_ACCEPT"
  | "QUEST_DECLINE"
  | "QUEST_CLAIM"
  | "QUEST_UPDATED"
  | "QUEST_REWARD"
  | "QUEST_COMPLETE"
  | "COLLECT_ITEM"
  | "FOREST_UNLOCK"
  | "EDUCATION_ANSWER"
  | "EDUCATION_QUESTION"
  | "EDUCATION_FEEDBACK"
  | "EDUCATION_CORRECT"
  | "HEAL"
  | "SHOP_OPEN"
  | "FAST_TRAVEL"
  | "PICKUP_ITEM"
  | "USE_ITEM"
  | "EQUIP_ITEM"
  | "UNEQUIP_ITEM"
  | "DISCARD_ITEM"
  | "GET_INVENTORY"
  | "INVENTORY_UPDATED"
  | "ITEM_ADDED"
  | "ITEM_REMOVED"
  | "ITEM_USED"
  | "ITEM_CONSUMED"
  | "EQUIPMENT_UPDATED"
  | "PLAYER_STATS_UPDATED"
  | "GIVE_ITEM"
  | "GIVE_CURRENCY"
  | "ADD_ITEM"
  | "SET_QUANTITY"
  | "SPLIT_STACK"
  | "UPGRADE_ITEM"
  | "ENCHANT_ITEM"
  | "EXPAND_BAG"
  | "SAVE_GEAR_LOADOUT"
  | "LOAD_GEAR_LOADOUT"
  | "CLAIM_TEMP_LOOT"
  | "SALVAGE_ITEMS"
  | "TOGGLE_COSMETIC"
  | "TRACE_ITEM"
  | "SET_ITEM_STATS"
  | "SET_ITEM_LEVEL"
  | "DUPLICATE_ITEM"
  | "SET_INSTANCE"
  | "PARTY_INVITE"
  | "PARTY_ACCEPT"
  | "PARTY_DECLINE"
  | "PARTY_LEAVE"
  | "PARTY_KICK"
  | "PARTY_DISBAND"
  | "PARTY_SET_TARGET"
  | "PARTY_UPDATED"
  | "PARTY_MEMBER_JOINED"
  | "PARTY_MEMBER_LEFT"
  | "FRIEND_REQUEST"
  | "ACCEPT_FRIEND"
  | "DECLINE_FRIEND"
  | "REMOVE_FRIEND"
  | "BLOCK_PLAYER"
  | "UNBLOCK_PLAYER"
  | "INSPECT_PLAYER"
  | "INSPECT_RESULT"
  | "GET_SOCIAL"
  | "FRIEND_UPDATED"
  | "SOCIAL_NOTIFICATION"
  | "SET_PRIVACY"
  | "GET_PRIVACY"
  | "PRIVACY_UPDATED"
  | "SET_PRESENCE"
  | "LOCAL_MUTE"
  | "UNMUTE_LOCAL"
  | "REPORT_MESSAGE"
  | "PARTY_CREATE"
  | "CHAT"
  | "NEARBY_PLAYERS"
  | "DUNGEON_ENTER"
  | "DUNGEON_READY"
  | "DUNGEON_LEAVE"
  | "DUNGEON_ABANDON"
  | "DUNGEON_RETRY"
  | "DUNGEON_OFFER"
  | "DUNGEON_READY_CHECK"
  | "DUNGEON_LOADING"
  | "DUNGEON_STARTED"
  | "DUNGEON_STATE"
  | "DUNGEON_WAVE"
  | "DUNGEON_OBJECTIVE"
  | "DUNGEON_COMPLETE"
  | "DUNGEON_FAILED"
  | "DUNGEON_LEFT"
  | "GET_DUNGEONS"
  | "DUNGEON_LIST"
  | "QUEUE_JOIN"
  | "QUEUE_LEAVE"
  | "QUEUE_UPDATE"
  | "DUNGEON_JOIN"
  | "DUNGEON_FILL"
  | "DUNGEON_REVIVE"
  | "DUNGEON_VOTE"
  | "DUNGEON_VOTE_UPDATE"
  | "DUNGEON_TAUNT"
  | "DUNGEON_WIPE"
  | "PLAYER_DOWNED"
  | "PLAYER_REVIVED"
  | "BOSS_INTERRUPT"
  | "BOSS_LOCK"
  | "SKIP_DUNGEON_INTRO"
  | "GET_DUNGEON_HISTORY"
  | "RAID_EXCHANGE"
  | "SET_OBJECTIVE"
  | "GIVE_LOOT"
  | "OBJECTIVE_COMPLETE"
  | "SET_BOSS_DEAD"
  | "SPAWN_BOSS"
  | "COMPLETE_DUNGEON"
  | "SKIP_MECHANIC"
  | "DAMAGE_BOSS"
  | "GET_OPEN_WORLD"
  | "OPEN_WORLD"
  | "GET_PVP"
  | "PVP_LOBBY"
  | "PVP_QUEUE_JOIN"
  | "PVP_QUEUE_LEAVE"
  | "PVP_QUEUE_UPDATE"
  | "PVP_READY"
  | "PVP_DECLINE"
  | "PVP_READY_CHECK"
  | "PVP_LOADING"
  | "PVP_COUNTDOWN"
  | "PVP_STATE"
  | "PVP_KILL_FEED"
  | "PVP_CAPTURE"
  | "PVP_RESULT"
  | "PVP_LEADERBOARD"
  | "PVP_HISTORY"
  | "PVP_EMOTE"
  | "PVP_SPECTATE"
  | "PVP_REPORT"
  | "PVP_LEAVE"
  | "PVP_LEFT"
  | "PVP_AFK"
  | "PVP_SHOP_BUY"
  | "PVP_TRAINING"
  | "PVP_DUEL"
  | "PVP_DUEL_ACCEPT"
  | "PVP_DUEL_DECLINE"
  | "PVP_DUEL_REQUEST"
  | "GET_REPLAY"
  | "PVP_REPLAY"
  | "SET_RATING"
  | "SET_RANK"
  | "PVP_WIN"
  | "SET_DAMAGE"
  | "BOSS_SPAWN"
  | "BOSS_PHASE"
  | "BOSS_TELEGRAPH"
  | "BOSS_AOE"
  | "BOSS_ENRAGE"
  | "BOSS_DEFEATED"
  | "BOSS_RESET"
  | "LOOT_RESULT"
  | "CLAIM_LOOT"
  | "GET_CHAPTERS"
  | "CHAPTER_LIST"
  | "UNLOCK_CHAPTER"
  | "SET_BOSS_HP"
  | "SET_CHAPTER"
  | "ALLOCATE_ATTRIBUTE"
  | "RESET_ATTRIBUTES"
  | "UNLOCK_SKILL"
  | "GET_PROGRESSION"
  | "REQUEST_TRANSFORMATION"
  | "SET_TRANSFORMATION"
  | "SET_LEVEL"
  | "SET_SKILL_POINTS"
  | "PROGRESSION_STATE"
  | "SKILL_UNLOCKED"
  | "SKILL_USED"
  | "TRANSFORMATION_STARTED"
  | "TRANSFORMATION_UPDATED"
  | "TRANSFORMATION_ENDED"
  | "TRANSFORMATION_REJECTED"
  | "POWER_RATING_UPDATED"
  | "SET_COMBAT_STYLE"
  | "SAVE_BUILD"
  | "LOAD_BUILD"
  | "SWITCH_BUILD"
  | "GET_BUILDS"
  | "BUILD_LIST"
  | "BUILD_SAVED"
  | "BUILD_LOADED"
  | "SET_LOADOUT"
  | "RESET_SKILLS"
  | "PLAYER_BLOCK"
  | "PLAYER_COUNTER"
  | "PLAYER_CHARGE"
  | "SET_ENERGY"
  | "SET_COOLDOWN"
  | "UNLOCK_TRANSFORM"
  | "TRAINING_METER"
  | "GET_WORLD_JOURNAL"
  | "GET_ADVENTURE"
  | "WORLD_JOURNAL"
  | "ZONE_DISCOVERED"
  | "LANDMARK_DISCOVERED"
  | "LORE_DISCOVERED"
  | "REQUEST_MOUNT"
  | "DISMOUNT"
  | "MOUNT_UPDATED"
  | "WORLD_EVENT"
  | "JOIN_WORLD_EVENT"
  | "CLAIM_EVENT_REWARD"
  | "EVENT_REWARD"
  | "GUARDIAN_DEFEATED"
  | "UNLOCK_GUARDIAN"
  | "UNLOCK_REGION"
  | "SET_WEATHER"
  | "TELEPORT"
  | "SKIP_CINEMATIC"
  | "CINEMATIC_SKIPPED"
  | "CINEMATIC_START"
  | "CINEMATIC_DONE"
  | "SET_LANGUAGE"
  | "GET_STORY"
  | "STORY_STATE"
  | "STORY_CHOICE"
  | "CLAIM_STORY_CHAPTER"
  | "REPLAY_CINEMATIC"
  | "REPLAY_CHAPTER"
  | "START_NG_PLUS"
  | "FAST_TRAVEL_OK"
  | "WEATHER_UPDATED"
  | "SWITCH_CHANNEL"
  | "CHANNEL_LIST"
  | "WORLD_BOSS_STATE"
  | "WORLD_BOSS_ANNOUNCE"
  | "TRIGGER_WORLD_BOSS"
  | "CLAIM_WORLD_BOSS"
  | "GET_COLLECTIONS"
  | "COLLECTION_BOOK"
  | "CHAPTER_COMPLETE"
  | "ENTER_MASJID"
  | "SET_WORLD_BOSS_HP"
  | "SET_WORLD_TIME"
  | "SPAWN_TREASURE"
  | "START_WORLD_EVENT"
  | "RANDOM_ENCOUNTER"
  | "GET_MOUNTS"
  | "MOUNT_COLLECTION"
  | "FAVORITE_MOUNT"
  | "EQUIP_MOUNT"
  | "SET_MOUNT_COSMETIC"
  | "MOUNT_EMOTE"
  | "UNLOCK_MOUNT"
  | "GRANT_MOUNT"
  | "CLAIM_MOUNT"
  | "SET_MOUNT_SPEED"
  | "RACE_START"
  | "RACE_CHECKPOINT"
  | "RACE_FINISH"
  | "RACE_UPDATED"
  | "TRAVEL_EVENT"
  | "INSPECT_LANDMARK"
  | "PARTY_SET_WAYPOINT"
  | "FOLLOW_PARTY"
  | "TRAVEL_SUGGESTION"
  | "GET_ENDGAME"
  | "ENDGAME_STATE"
  | "CLAIM_DAILY"
  | "CLAIM_WEEKLY"
  | "CLAIM_CHALLENGE"
  | "CLAIM_SEASON"
  | "GET_SEASON"
  | "SET_SEASON_XP"
  | "GET_HORIZON"
  | "GET_CALENDAR"
  | "GET_LIVE_EVENT"
  | "CONTRIBUTE_EVENT"
  | "UNLOCK_ACHIEVEMENT"
  | "UNLOCK_COSMETIC"
  | "SET_ACHIEVEMENT"
  | "SET_SHOWCASE"
  | "GET_PUBLIC_PROFILE"
  | "GET_LEARNING"
  | "GET_LORE_BOOK"
  | "GET_LEADERBOARDS"
  | "ACHIEVEMENT_UNLOCKED"
  | "CHAT_MESSAGE"
  | "GUILD_CREATE"
  | "GUILD_INVITE"
  | "GUILD_ACCEPT"
  | "GUILD_DECLINE"
  | "GUILD_LEAVE"
  | "GUILD_KICK"
  | "GUILD_DISBAND"
  | "GUILD_TRANSFER"
  | "GUILD_ANNOUNCE"
  | "GUILD_UPDATED"
  | "SET_GUILD_RANK"
  | "GET_GUILD"
  | "TRADE_REQUEST"
  | "TRADE_ACCEPT"
  | "TRADE_DECLINE"
  | "TRADE_OFFER"
  | "TRADE_READY"
  | "TRADE_CONFIRM"
  | "TRADE_CANCEL"
  | "TRADE_UPDATED"
  | "SHOP_BUY"
  | "SHOP_SELL"
  | "SHOP_CATALOG"
  | "SET_COIN"
  | "SET_TITLE"
  | "SET_COSMETIC"
  | "REPORT_PLAYER"
  | "SEARCH_PLAYER"
  | "SEARCH_RESULT"
  | "PARTY_FINDER_LIST"
  | "PARTY_FINDER_CREATE"
  | "PARTY_FINDER_JOIN"
  | "PARTY_TRANSFER"
  | "PARTY_SET_ROLE"
  | "PARTY_READY"
  | "NOTIFY_LIST"
  | "GET_NOTIFIES"
  | "MUTE_PLAYER"
  | "MOD_KICK"
  | "MOD_BAN"
  | "GET_MARKET"
  | "MARKET_SEARCH"
  | "MARKET_LIST"
  | "MARKET_BUY"
  | "MARKET_CANCEL"
  | "MARKET_LISTINGS"
  | "BANK_DEPOSIT"
  | "BANK_WITHDRAW"
  | "GET_BANK"
  | "BANK_UPDATED"
  | "LOCK_ITEM"
  | "FAVORITE_ITEM"
  | "HOUSE_ENTER"
  | "HOUSE_LEAVE"
  | "HOUSE_VISIT"
  | "HOUSE_PLACE"
  | "HOUSE_REMOVE"
  | "SET_HOUSE_ACCESS"
  | "GET_HOUSE"
  | "HOUSE_STATE"
  | "GUILD_APPLY"
  | "GUILD_REVIEW"
  | "GUILD_DEPOSIT"
  | "GUILD_WITHDRAW"
  | "GET_GUILD_LOG"
  | "GUILD_LOG"
  | "GUILD_SET_EMBLEM"
  | "GUILD_SET_DESC"
  | "SOCIAL_EMOTE"
  | "EMOTE_PLAYED"
  | "GET_PLAYER_CARD"
  | "PLAYER_CARD"
  | "SET_NAME"
  | "GATHER"
  | "GATHER_RESULT"
  | "CRAFT"
  | "CRAFT_RESULT"
  | "GET_CRAFTING"
  | "CRAFTING_STATE"
  | "SET_PROFESSION"
  | "RESET_PROFESSION"
  | "FISH_START"
  | "FISH_CATCH"
  | "FISH_STATE"
  | "NPC_SHOP_OPEN"
  | "NPC_SHOP_BUY"
  | "NPC_SHOP_SELL"
  | "NPC_REPAIR"
  | "STALL_OPEN"
  | "STALL_LIST"
  | "STALL_BUY"
  | "STALL_CLOSE"
  | "CRAFT_ORDER"
  | "CRAFT_ORDER_ACCEPT"
  | "GET_WORKSHOP"
  | "ADD_GOLD"
  | "ADD_MATERIAL"
  | "SET_PRICE"
  | "SET_GOLD"
  | "GIVE_GOLD"
  | "CREATE_RECIPE"
  | "GUILD_CONTRIBUTE"
  | "GET_GOLD_LOG"
  | "CREATE_HOUSE"
  | "HOUSE_LOCK"
  | "HOUSE_RENAME"
  | "HOUSE_STYLE"
  | "HOUSE_MOVE"
  | "HOUSE_DECORATE"
  | "HOUSE_STORE"
  | "HOUSE_TAKE"
  | "GARDEN_PLANT"
  | "GARDEN_WATER"
  | "GARDEN_HARVEST"
  | "PET_CLAIM"
  | "PET_SUMMON"
  | "PET_DISMISS"
  | "PET_CARE"
  | "PET_NAME"
  | "GET_PETS"
  | "ADD_PET"
  | "GUILD_HALL_ENTER"
  | "GUILD_HALL_LEAVE"
  | "GUILD_HOST"
  | "GET_LIFE"
  | "CLAIM_DAILY_LIFE"
  | "HOUSE_VOTE"
  | "LIFE_QUIZ"
  | "CLAIM_COLLECTION"
  | "LIFE_STATE"
  | "PET_STATE";

export type MovementState = "IDLE" | "WALK" | "RUN" | "JUMP" | "FALL" | "DODGE" | "DEAD";

export interface Envelope<T = unknown> {
  type: NetMsgType;
  data?: T;
}

export interface AuthIn {
  name: string;
  token?: string;
}

export interface AuthOkOut {
  token: string;
  playerId: string;
  sessionId: string;
}

export interface JoinWorldIn {
  worldId: string;
}

export interface MoveInput {
  seq: number;
  ax: number;
  az: number;
  yaw: number;
  sprint: boolean;
  jump: boolean;
}

export interface PingIn {
  t: number;
}

export interface PongOut {
  t: number;
  st: number;
}

export interface AttackAction {
  attackType: string;
  targetId: string;
  timestamp: number;
  direction: number;
}

export interface SkillAction {
  skillId: string;
  targetId: string;
  timestamp: number;
  direction: number;
}

export interface CombatAction {
  kind: "attack" | "skill" | "dodge" | "combo" | "respawn";
}

export interface PlayerSnapshot {
  id: string;
  x: number;
  y: number;
  z: number;
  yaw: number;
  vx: number;
  vy: number;
  vz: number;
  st: MovementState | string;
  cs?: string;
  hp?: number;
  maxHp?: number;
  energy?: number;
  maxEnergy?: number;
  stamina?: number;
  level?: number;
  exp?: number;
  expToNext?: number;
  seq: number;
  formId?: string;
  transform?: string;
  mountId?: string;
  mounted?: boolean;
  zoneId?: string;
  guildTag?: string;
  title?: string;
  petId?: string;
}

export interface PlayerSpawn {
  playerId: string;
  name: string;
  level: number;
  class: string;
  hp: number;
  maxHp: number;
  energy?: number;
  maxEnergy?: number;
  exp?: number;
  expToNext?: number;
  x: number;
  y: number;
  z: number;
  yaw: number;
  state: MovementState | string;
  combatState?: string;
  formId?: string;
  guildTag?: string;
  title?: string;
  mountId?: string;
  mounted?: boolean;
}

export interface NPCSnapshot {
  id: string;
  name: string;
  role: string;
  type: string;
  x: number;
  y: number;
  z: number;
  yaw: number;
  activity?: string;
  voiceLineId?: string;
  marker?: string;
}

export interface ObjectSnapshot {
  id: string;
  kind: string;
  x: number;
  z: number;
  text?: string;
}

export interface DialogOption {
  id: string;
  label: string;
}

export interface QuestionOut {
  id: string;
  index: number;
  total: number;
  category: string;
  prompt: string;
  choices: string[];
}

export interface RewardView {
  exp: number;
  coin: number;
  crystal?: number;
  potion?: number;
  eduToken?: number;
  perfect?: boolean;
}

export interface ShopItem {
  id: string;
  name: string;
  price: number;
}

export interface InteractResult {
  kind: string;
  targetId: string;
  title: string;
  speaker: string;
  role: string;
  text: string;
  marker?: string;
  options?: DialogOption[];
  toast?: string;
  locked?: boolean;
  shop?: ShopItem[];
  question?: QuestionOut;
  rewards?: RewardView;
  subtitle?: string;
  cinematicId?: string;
  emotion?: string;
  gesture?: string;
  voiceId?: string;
  history?: string[];
}

export interface EducationFeedback {
  correct: boolean;
  explain: string;
  retry?: boolean;
  toast?: string;
  question?: QuestionOut;
}

export interface ObjectiveView {
  type: string;
  target: string;
  text: string;
  count: number;
  progress: number;
}

export interface QuestView {
  id: string;
  title: string;
  kind: string;
  state: string;
  description: string;
  npc: string;
  npcName: string;
  location: string;
  objectives: ObjectiveView[];
  rewards: RewardView;
}

export interface PlayerProgress {
  playerId: string;
  quests: QuestView[];
  flags: Record<string, boolean>;
  coin: number;
  potion: number;
  crystal: number;
  eduToken: number;
  energyPotion: number;
  forestUnlocked: boolean;
  claimed: string[];
  fastTravel: string[];
  timeOfDay: string;
  activeQuestId: string;
  chapters?: ChapterView[];
  zoneId?: string;
  weather?: string;
  knowledgePoints?: number;
  clock?: string;
  clockLabel?: string;
  regionReputation?: Record<string, number>;
  factionReputation?: Record<string, number>;
  journey?: JourneyView;
}

export interface JourneyView {
  title: string;
  objective: string;
  hint?: string;
  region?: string;
  nextRegion?: string;
  navX?: number;
  navZ?: number;
  landmark?: string;
  cardinal?: string;
  subObjective?: string;
  optional?: string;
  optionalX?: number;
  optionalZ?: number;
  hintJv?: string;
  hintId?: string;
}

export function emptyProgress(): PlayerProgress {
  return {
    playerId: "",
    quests: [],
    flags: {},
    coin: 0,
    potion: 0,
    crystal: 0,
    eduToken: 0,
    energyPotion: 0,
    forestUnlocked: false,
    claimed: [],
    fastTravel: [],
    timeOfDay: "DAY",
    activeQuestId: "",
  };
}

export interface ItemEffects {
  healPct?: number;
  energyPct?: number;
  staminaPct?: number;
  attack?: number;
  defense?: number;
  maxHp?: number;
  maxEnergy?: number;
  strength?: number;
  agility?: number;
  energyPower?: number;
  criticalChance?: number;
  movementSpeed?: number;
  range?: number;
  dodge?: number;
  energyRegen?: number;
}

export interface ItemDefView {
  id: string;
  name: string;
  description: string;
  type: string;
  slot?: string;
  rarity: string;
  stackable: boolean;
  maxStack: number;
  icon: string;
  value: number;
  levelRequirement: number;
  effects: ItemEffects;
  tradable?: boolean;
  bind?: string;
  itemLevel?: number;
  setId?: string;
  lore?: string;
}

export interface InvSlotView {
  index: number;
  item?: ItemDefView | null;
  qty: number;
  locked?: boolean;
  favorite?: boolean;
  itemInstanceId?: string;
  upgrade?: number;
  itemLevel?: number;
}

export interface EquipmentView {
  HEAD?: string;
  BODY?: string;
  LEGS?: string;
  WEAPON?: string;
  ACCESSORY_1?: string;
  ACCESSORY_2?: string;
  ACCESSORY_3?: string;
}

export interface StatsView {
  level: number;
  class: string;
  hp: number;
  maxHp: number;
  energy: number;
  maxEnergy: number;
  stamina: number;
  attack: number;
  defense: number;
  strength: number;
  agility: number;
  energyPower: number;
  vitality?: number;
  criticalChance: number;
  moveSpeed: number;
  attributePoints?: number;
  skillPoints?: number;
  powerRating?: number;
  formId?: string;
  transformState?: string;
  dodge?: number;
  energyRegen?: number;
  showCosmetic?: boolean;
}

export interface InventoryUpdated {
  playerId: string;
  inventoryVersion: number;
  changedSlots?: InvSlotView[];
  slots?: InvSlotView[];
  equipment: EquipmentView;
  stats: StatsView;
  coin: number;
  crystal: number;
  eduToken: number;
  battleToken?: number;
  guardianToken?: number;
  raidToken?: number;
  bank?: InvSlotView[];
  toast?: string;
  tempLoot?: { itemId: string; name: string; qty: number; until: number }[];
  setPieces?: number;
  bagCapacity?: number;
  showCosmetic?: boolean;
  itemHistory?: { itemInstanceId: string; playerId: string; itemId: string; action: string; at: number }[];
}

export interface PartyMemberView {
  playerId: string;
  name: string;
  level: number;
  class: string;
  hp: number;
  maxHp: number;
  energy: number;
  maxEnergy: number;
  distance: number;
  online: boolean;
  role: string;
  leader: boolean;
  ready?: boolean;
  status?: string;
}

export interface PartyView {
  partyId: string;
  leaderId: string;
  members: PartyMemberView[];
  targetId?: string;
  targetName?: string;
  targetHp?: number;
  targetMaxHp?: number;
  targetLevel?: number;
  notifyIds?: string[];
  activity?: string;
  minLevel?: number;
  requiredRole?: string;
  state?: string;
  ready?: Record<string, boolean>;
  waypointId?: string;
  waypointX?: number;
  waypointZ?: number;
}

export interface FriendView {
  playerId: string;
  name: string;
  level: number;
  class: string;
  online: boolean;
  lastSeen: number;
  status?: string;
  guild?: string;
  title?: string;
  avatar?: string;
  region?: string;
}

export interface NearbyView {
  playerId: string;
  name: string;
  level: number;
  class: string;
  distance: number;
  online: boolean;
}

export interface SocialState {
  party?: PartyView | null;
  friends: FriendView[];
  pending: FriendView[];
  outgoing: FriendView[];
  blocked: string[];
  nearby: NearbyView[];
  notifies?: NotifyView[];
  guild?: GuildView | null;
  wallet?: WalletView;
  toId?: string;
  privacy?: { friend?: string; party?: string; trade?: string; pm?: string };
}

export interface SocialNote {
  kind: string;
  text: string;
  fromId?: string;
  from?: string;
  toId?: string;
  priority?: string;
}

export interface InspectOut {
  playerId: string;
  name: string;
  level: number;
  class: string;
  stats: StatsView;
  equipment: EquipmentView;
  powerRating?: number;
  guild?: string;
  guildTag?: string;
  title?: string;
  rank?: string;
  season?: string;
  seasonLevel?: number;
  badge?: string;
  aura?: string;
  mount?: string;
  avatar?: string;
  status?: string;
  region?: string;
  achievementScore?: number;
}

export interface EndgameQuestView {
  id: string;
  title: string;
  kind: string;
  progress: number;
  count: number;
  claimed: boolean;
  tier?: string;
}

export interface EndgameState {
  unlocked?: boolean;
  hub?: string;
  level?: number;
  seasonName?: string;
  seasonId?: string;
  seasonLevel?: number;
  seasonXP?: number;
  seasonXPNeed?: number;
  seasonEnd?: string;
  currentReward?: { level?: number; kind?: string; id?: string };
  nextReward?: { level?: number; kind?: string; id?: string };
  history?: Array<{ seasonId?: string; name: string; level: number; rank?: string }>;
  daily?: EndgameQuestView[];
  weekly?: EndgameQuestView[];
  challenges?: EndgameQuestView[];
  achievements?: string[];
  titles?: string[];
  cosmetics?: string[];
  collection?: { aura?: number; auraTotal?: number; mount?: number; mountTotal?: number; titles?: number; titleTotal?: number };
  learning?: { answered?: number; correct?: number; accuracy?: number; streak?: number };
  lore?: Array<{ id?: string; title: string; text?: string }>;
  horizon?: { best?: number; week?: string; board?: Array<{ rank?: number; name?: string; score?: number; level?: number }> };
  calendar?: { timezone?: string; today?: unknown; week?: unknown[]; upcoming?: unknown[] };
  community?: { id?: string; name?: string; points?: number; target?: number; state?: string };
  leaderboards?: { horizon?: unknown[]; season?: unknown[]; guilds?: unknown[]; level?: unknown[]; combat?: unknown[]; bossDefeated?: unknown[]; quest?: unknown[]; crafting?: unknown[]; gathering?: unknown[] };
  fragments?: number;
  showcase?: { title?: string; badge?: string; aura?: string; mount?: string };
  serverTime?: string;
  version?: string;
  phase?: string;
  release?: string;
  phase29?: {
    mainStory?: { title?: string; objective?: string; chapters?: Array<{ index: number; title: string; state: string; locked?: boolean }>; regions?: Array<{ id: string; theme: string; purpose: string }> };
    siluman33?: Array<{ id: string; name: string; title: string; region: string; level: number; element: string; skills: string[]; weakness: string; lore: string; bossType: string }>;
    worldEvent?: { id?: string; name?: string; active?: boolean; globalProgress?: number; timer?: number };
    worldBoss?: { id?: string; name?: string; state?: string; dailyLimit?: string };
    raid?: { id?: string; name?: string; minPlayers?: number; maxPlayers?: number; difficulties?: string[]; checkpoint?: boolean; bosses?: string[] };
    seasonal?: { theme?: string; seasonName?: string; festival?: string; upcoming?: unknown };
    guildEvent?: { name?: string; weekly?: boolean; hallEvent?: string; contributions?: number };
    finalJourney?: { questId?: string; questName?: string; raidId?: string; gateLocked?: boolean; partyRequired?: number; cinematic?: string };
  };
}

export interface WalletView {
  coins: number;
  crystals: number;
  educationTokens: number;
  guildTokens: number;
  battleTokens?: number;
  guardianTokens?: number;
  raidTokens?: number;
}

export interface NotifyView {
  id: string;
  type: string;
  message: string;
  read?: boolean;
  timestamp: number;
  priority?: string;
}

export interface ChatOut {
  channel: string;
  fromId: string;
  from: string;
  text: string;
  toId?: string;
  system?: boolean;
  notifyIds?: string[];
}

export interface GuildMemberView {
  playerId: string;
  name: string;
  rank: string;
  contribution: number;
  joinedAt: number;
}

export interface GuildView {
  guildId: string;
  name: string;
  tag: string;
  leader: string;
  level: number;
  exp: number;
  members: GuildMemberView[];
  announcement: string;
  emblemId: string;
  quest?: string;
  description?: string;
  capacity?: number;
  storage?: InvSlotView[];
  logs?: Array<{ player: string; itemId: string; qty: number; action: string; at: number }>;
  applications?: string[];
  questProgress?: number;
  notifyIds?: string[];
}

export interface MarketListing {
  id: string;
  sellerId: string;
  seller: string;
  itemId: string;
  category: string;
  rarity: string;
  qty: number;
  price: number;
  level: number;
  created: number;
}

export interface MarketHistoryRow {
  id: string;
  kind: string;
  itemId: string;
  other?: string;
  qty: number;
  price: number;
  at: number;
}

export interface MarketState {
  listings?: MarketListing[];
  page?: number;
  pageSize?: number;
  total?: number;
  history?: MarketHistoryRow[];
  wallet?: WalletView;
}

export interface HouseItemView {
  id: string;
  itemId: string;
  x: number;
  y: number;
  z: number;
  yaw: number;
}

export interface HouseGuestView {
  playerId: string;
  at: number;
}

export interface GardenPlotView {
  id: string;
  plant?: string;
  state?: string;
}

export interface HouseState {
  instanceId?: string;
  ownerId?: string;
  houseId?: string;
  type?: string;
  access?: string;
  district?: string;
  items?: HouseItemView[];
  visitors?: string[];
  limit?: number;
  left?: boolean;
  name?: string;
  sign?: string;
  locationId?: string;
  layoutId?: string;
  locked?: boolean;
  guildHall?: boolean;
  wall?: string;
  floor?: string;
  roof?: string;
  light?: string;
  color?: string;
  plots?: GardenPlotView[];
  rooms?: Array<{ id: string; name: string }>;
  locations?: Array<{ id: string; name: string }>;
  guestLog?: HouseGuestView[];
  grid?: number;
}

export interface LifeSkillView {
  id: string;
  name: string;
  xp: number;
  level: number;
  max: number;
}

export interface PetView {
  petId: string;
  name: string;
  owned: boolean;
  source?: string;
  happiness?: number;
  mood?: string;
  cosmetic?: string;
  active?: boolean;
}

export interface DailyLifeView {
  id: string;
  title: string;
  claimed?: boolean;
  ready?: boolean;
}

export interface LifeState {
  lifeLevel?: number;
  lifeXp?: number;
  farmingLevel?: number;
  plots?: number;
  utcDay?: string;
  skills?: LifeSkillView[];
  pets?: PetView[];
  activePet?: string;
  dailies?: DailyLifeView[];
  collections?: Record<string, boolean>;
  collectionClaimed?: Record<string, boolean>;
}

export interface PlayerCardView {
  playerId: string;
  name: string;
  level: number;
  class: string;
  guild?: string;
  guildTag?: string;
  title?: string;
  rank?: string;
  achievements?: number;
  cosmetic?: string;
  season?: string;
  status?: string;
}

export interface TradeSlotView {
  slot: number;
  itemId: string;
  qty: number;
}

export interface TradeOfferView {
  playerId: string;
  slots: TradeSlotView[];
  coin: number;
  ready: boolean;
  confirm: boolean;
}

export interface TradeView {
  tradeId: string;
  transactionId: string;
  a: TradeOfferView;
  b: TradeOfferView;
  state: string;
  result?: string;
  notifyIds?: string[];
}

export interface ShopItemView {
  shopItemId: string;
  itemId: string;
  name: string;
  price: number;
  currency: string;
  stock: number;
  purchaseLimit: number;
  bought: number;
}

export interface ShopCatalogOut {
  id?: string;
  name?: string;
  items?: ShopItemView[];
  wallet?: WalletView;
}

export interface FinderListing {
  id: string;
  leader: string;
  activity: string;
  level: number;
  requiredRole: string;
  players: number;
  cap: number;
}

export interface SearchHit {
  playerId: string;
  name: string;
  level: number;
  guild: string;
  status: string;
}

export interface DropSnapshot {
  id: string;
  itemId: string;
  name: string;
  qty: number;
  x: number;
  z: number;
  rarity: string;
}

export interface WorldSnapshot {
  worldId: string;
  channel: string;
  t: number;
  online: number;
  players: PlayerSnapshot[];
  npcs: NPCSnapshot[];
  enemies: EnemySnapshot[];
  objects: ObjectSnapshot[];
  drops?: DropSnapshot[];
  timeOfDay?: string;
  weather?: string;
  zoneId?: string;
  event?: WorldEventView | null;
  worldBoss?: WorldBossView | null;
  instanceId?: string;
  dungeon?: DungeonView;
  pvp?: PvpView;
  clock?: string;
  clockLabel?: string;
  worldName?: string;
}

export interface WelcomeOut {
  worldId: string;
  channel: string;
  tickRate: number;
  playerId: string;
  sessionId: string;
  self: PlayerSpawn;
  players: PlayerSpawn[];
  snapshot: WorldSnapshot;
  progress?: PlayerProgress;
  loadout?: InventoryUpdated;
  catalog?: ItemDefView[];
  social?: SocialState;
  progression?: import("../progression/ProgressionStore").ProgressionView;
}

export interface EnemySnapshot {
  id: string;
  kind: string;
  name: string;
  level: number;
  x: number;
  y: number;
  z: number;
  yaw: number;
  hp: number;
  maxHp: number;
  st: string;
  rank: string;
}

export interface EnemyState {
  id: string;
  kind: string;
  name: string;
  level: number;
  hp: number;
  maxHp: number;
  combatState: string;
}

export interface DespawnOut {
  playerId: string;
}

export interface ErrorOut {
  message: string;
}

export interface AttackResult {
  attackerId: string;
  attackType: string;
  skillId?: string;
  targetIds: string[];
  comboHits?: number;
  finisher?: boolean;
}

export interface DamageResult {
  attackerId: string;
  targetId: string;
  damage: number;
  isCritical: boolean;
  blocked?: boolean;
  hitX: number;
  hitY: number;
  hitZ: number;
  attackType: string;
  timestamp: number;
  targetHp: number;
  targetMaxHp: number;
  killed: boolean;
  kind: "enemy" | "player" | string;
}

export interface DeathOut {
  playerId: string;
  respawnAt: number;
}

export interface RespawnOut {
  playerId: string;
  x: number;
  y: number;
  z: number;
  hp: number;
}

export interface LevelUpOut {
  playerId: string;
  fromLevel?: number;
  newLevel: number;
  maxHp: number;
  attributePoints?: number;
  skillPoints?: number;
  reward: string;
}

export interface EnemyDeathOut {
  enemyId: string;
  killerId: string;
  exp: number;
}

export interface RejectOut {
  action: string;
  reason: string;
  playerId?: string;
}

export function wsUrl(): string {
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${window.location.host}/cahaya/ws`;
}

export function displayName(): string {
  const q = new URLSearchParams(window.location.search).get("name");
  if (q && q.trim()) return q.trim().slice(0, 18);
  return "Raka";
}

export interface DungeonOffer {
  dungeonId: string;
  name: string;
  chapterId: string;
  kind?: string;
  description?: string;
  difficulty?: string;
  recommendedLevel: number;
  minPlayers?: number;
  maxPlayers: number;
  timeLimit: number;
  rewards?: RewardView;
  status: string;
  region?: string;
  difficulties?: string[];
  lockoutResetAt?: number;
  lockoutLabel?: string;
}

export interface DungeonReadyMember {
  playerId: string;
  ready: boolean;
}

export interface DungeonReadyOut {
  dungeonId: string;
  leaderId: string;
  until: number;
  members: DungeonReadyMember[];
  cancelled?: boolean;
  fromQueue?: boolean;
}

export interface DungeonLoading {
  instanceId: string;
  dungeonId: string;
  name: string;
}

export interface BossView {
  id: string;
  name: string;
  title: string;
  level: number;
  hp: number;
  maxHp: number;
  phase: number;
  enraged: boolean;
  alive: boolean;
}

export interface DungeonMemberView {
  playerId: string;
  name: string;
  level: number;
  hp: number;
  maxHp: number;
  dead: boolean;
  downed?: boolean;
  online: boolean;
  distance: number;
  role?: string;
  reviveProgress?: number;
  reviveToken?: number;
  energy?: number;
  maxEnergy?: number;
}

export interface LootItemView {
  itemId: string;
  name: string;
  qty: number;
  rarity: string;
}

export interface DungeonView {
  instanceId: string;
  dungeonId: string;
  chapterId: string;
  kind?: string;
  name: string;
  title: string;
  state: string;
  wave: number;
  waveTotal: number;
  encounter?: number;
  enemies: number;
  objective: string;
  objectiveType: string;
  progress: number;
  count: number;
  timeLeft: number;
  rating: string;
  claimId?: string;
  chest: boolean;
  elapsed: number;
  toast?: string;
  bossLocked?: boolean;
  crystalShield?: boolean;
  puzzleStep?: number;
  wipeCount?: number;
  synergy?: boolean;
  boss?: BossView;
  members?: DungeonMemberView[];
  loot?: LootItemView[];
  votes?: Record<string, string>;
  difficulty?: string;
  mechanic?: string;
  guideHp?: number;
  eduShield?: boolean;
  lockoutLabel?: string;
  room?: string;
  checkpoint?: string;
}

export interface QueueView {
  state: string;
  dungeonId: string;
  name: string;
  role?: string;
  players?: number;
  minPlayers?: number;
  maxPlayers?: number;
  waitMs?: number;
}

export interface DungeonRunRow {
  playerId: string;
  name: string;
  dungeonId: string;
  nameDun?: string;
  kind: string;
  difficulty: string;
  rating: string;
  elapsed: number;
  deaths: number;
  wipes: number;
  guild?: string;
  party?: string;
  bosses?: number;
  at: number;
  firstClear?: boolean;
}

export interface DungeonBoardRow {
  dungeonId: string;
  name: string;
  playerId: string;
  player: string;
  guild?: string;
  party?: string;
  elapsed: number;
  difficulty: string;
  rating: string;
  at: number;
}

export interface RaidShopItem {
  id: string;
  name: string;
  cost: number;
  kind: string;
  rewardId: string;
}

export interface DungeonListOut {
  dungeons: DungeonOffer[];
  queue?: QueueView;
  history?: DungeonRunRow[];
  board?: DungeonBoardRow[];
  raidShop?: RaidShopItem[];
  raidTokens?: number;
  lockoutResetAt?: number;
  lockoutLabel?: string;
}

export interface LootResult {
  playerId: string;
  claimId: string;
  items: LootItemView[];
  exp: number;
  coin: number;
  crystal: number;
}

export interface BossTelegraphOut {
  instanceId: string;
  skill: string;
  x: number;
  z: number;
  radius: number;
  until: number;
  vfx: string;
  shape?: string;
  interruptible?: boolean;
  pulse?: boolean;
}

export interface BossAOEOut {
  instanceId: string;
  x: number;
  z: number;
  radius: number;
  damage: number;
}

export interface BossPhaseOut {
  instanceId: string;
  phase: number;
  label: string;
}

export interface ChapterView {
  id: string;
  title: string;
  region: string;
  story: string;
  bossId: string;
  bossName: string;
  requiredLevel: number;
  reward: RewardView;
  status: string;
  dungeonId: string;
}

export interface ChapterListOut {
  chapters: ChapterView[];
}

export interface WorldEventView {
  id: string;
  name: string;
  kind?: string;
  state: string;
  phase?: string;
  region: string;
  announce: string;
  announceJv?: string;
  objective?: string;
  progress?: number;
  need?: number;
  gateHp: number;
  maxGateHp: number;
  until: number;
  startsAt?: number;
  startsIn?: number;
  endsIn?: number;
  participants?: number;
  success?: boolean;
  x?: number;
  z?: number;
}

export interface GuardianView {
  id: string;
  name: string;
  title: string;
  status: string;
  chapterId: string;
  region: string;
  index?: number;
  personality?: string;
  weakness?: string;
  story?: string;
  storyName?: string;
  codexStatus?: string;
  ally?: boolean;
  miniBoss?: boolean;
}

export interface WorldBossView {
  id: string;
  name: string;
  region: string;
  state: string;
  announce: string;
  hp: number;
  maxHp: number;
  phase: number;
  phaseName?: string;
  until: number;
  players: number;
  x?: number;
  z?: number;
}

export interface RegionView {
  id: string;
  name: string;
  title?: string;
  discovered: boolean;
  unlocked: boolean;
  completion: number;
  recommendedLevel?: number;
  minimumLevel?: number;
  enemyTier?: string;
  resourceTier?: string;
}

export interface LoreView {
  id: string;
  title: string;
  text: string;
  region?: string;
  kind?: string;
  discovered?: boolean;
  personality?: string;
  mechanic?: string;
}

export interface NpcRelView {
  id: string;
  name: string;
  relationship: string;
  xp: number;
  nextReward?: string;
  memory?: boolean;
  role?: string;
  trait?: string;
}

export interface ChoiceHistView {
  id: string;
  choice: string;
  impact?: string;
}

export interface OverlayChapterView {
  index: number;
  title: string;
  state: string;
  locked?: boolean;
}

export interface EnemyLoreView {
  id: string;
  name: string;
  region: string;
  personality: string;
  mechanic: string;
  lore?: string;
  encountered: boolean;
  defeated: boolean;
  discovered: boolean;
}

export interface LandmarkView {
  id: string;
  name: string;
  region: string;
  discovered: boolean;
  x: number;
  z: number;
}

export interface WorldJournal {
  playerId: string;
  guardians: GuardianView[];
  regions: RegionView[];
  lore: LoreView[];
  landmarks: LandmarkView[];
  guardiansDefeated: number;
  guardiansTotal: number;
  regionsDiscovered: number;
  regionsTotal: number;
  mounts: string[];
  celestialGate: boolean;
  objective: string;
  tokens?: number;
  storyCompleted?: boolean;
  explorerMode?: boolean;
  achievements?: string[];
  channel?: string;
  language?: string;
  storyChapter?: string;
  storyState?: string;
  storyChapters?: Array<{ id: string; index: number; title: string; state: string; music?: string }>;
  allies?: string[];
  endingId?: string;
  ngPlus?: number;
  markers?: Array<{ id: string; kind: string; name: string; region?: string; x: number; z: number }>;
  nextWorldBoss?: { id: string; name: string; region: string; announce?: string; when?: string };
  worldMood?: string;
  npcBook?: NpcRelView[];
  choiceHistory?: ChoiceHistView[];
  overlayChapters?: OverlayChapterView[];
  enemyLore?: EnemyLoreView[];
  loreCards?: LoreView[];
}

export function spawnToSnapshot(spawn: PlayerSpawn): PlayerSnapshot {
  return {
    id: spawn.playerId,
    x: spawn.x,
    y: spawn.y,
    z: spawn.z,
    yaw: spawn.yaw,
    vx: 0,
    vy: 0,
    vz: 0,
    st: spawn.state,
    cs: spawn.combatState,
    hp: spawn.hp,
    maxHp: spawn.maxHp,
    seq: 0,
  };
}

export interface PvpModeView {
  id: string;
  name: string;
  kind: string;
  teamSize: number;
  minLevel: number;
  duration: number;
  map: string;
  enabled: boolean;
  status: string;
}

export interface PvpProfileView {
  rating: number;
  rank: string;
  rankName: string;
  division?: string;
  rankVisual?: string;
  wins: number;
  losses: number;
  winRate: number;
  placementLeft: number;
  winStreak: number;
  lossStreak: number;
  highestRank: string;
  battleToken: number;
  season: string;
  seasonId: string;
}

export interface PvpSeasonView {
  id: string;
  name: string;
  number: number;
  start: string;
  end: string;
  weeks: number;
}

export interface PvpShopView {
  shopItemId: string;
  itemId: string;
  name: string;
  kind: string;
  price: number;
  currency: string;
}

export interface PvpRewardView {
  id: string;
  rank: string;
  kind: string;
  name: string;
  unlocked: boolean;
}

export interface PvpHistoryRow {
  matchId: string;
  mode: string;
  opponent: string;
  result: string;
  ratingChange: number;
  date: number;
  duration?: number;
  kills?: number;
  deaths?: number;
  assists?: number;
  damage?: number;
  objective?: number;
}

export interface PvpSeasonHistoryRow {
  seasonId: string;
  season: string;
  highestRank: string;
  finalRating: number;
  wins: number;
  losses: number;
}

export interface PvpNearbyView {
  playerId: string;
  name: string;
  level: number;
  cosmetic?: string;
}

export interface PvpQueueView {
  state: string;
  mode?: string;
  name?: string;
  players?: number;
  need?: number;
  waitMs?: number;
  waitEstMs?: number;
  waitNote?: string;
}

export interface PvpMemberView {
  playerId: string;
  name: string;
  team: number;
  ready?: boolean;
  hp?: number;
  maxHp?: number;
  alive?: boolean;
  kills?: number;
  deaths?: number;
  assists?: number;
  damage?: number;
  pingMs?: number;
  spectateId?: string;
}

export interface PvpPointView {
  id: string;
  owner: number;
  contested: boolean;
  progressA: number;
  progressB: number;
  x: number;
  z: number;
}

export interface PvpView {
  matchId: string;
  mode: string;
  map: string;
  state: string;
  timeLeft: number;
  scoreA: number;
  scoreB: number;
  members: PvpMemberView[];
  points?: PvpPointView[];
  killFeed?: string[];
  team: number;
  instanceId?: string;
  countdown?: number;
}

export interface PvpReadyOut {
  matchId: string;
  mode: string;
  name: string;
  until: number;
  members: PvpMemberView[];
}

export interface LBEntry {
  rank: number;
  playerId: string;
  player: string;
  rankName: string;
  rating: number;
  wins: number;
  losses: number;
  winRate: number;
}

export interface PvpLobbyOut {
  modes: PvpModeView[];
  profile: PvpProfileView;
  season: PvpSeasonView;
  shop: PvpShopView[];
  rewards: PvpRewardView[];
  history: PvpHistoryRow[];
  seasonHistory?: PvpSeasonHistoryRow[];
  nearby?: PvpNearbyView[];
  queue?: PvpQueueView;
  match?: PvpView;
  training?: boolean;
  leaderboard?: LBEntry[];
}

export interface PvpResultOut {
  matchId: string;
  mode: string;
  title: string;
  victory: boolean;
  draw?: boolean;
  kills: number;
  deaths: number;
  assists: number;
  damage: number;
  objective?: number;
  duration?: number;
  mvp?: boolean;
  mvpName?: string;
  ratingChange: number;
  rating: number;
  rank: string;
  rankName: string;
  promoted: boolean;
  demoted: boolean;
  battleToken: number;
  seasonXP?: number;
  rewards?: string[];
}

export interface CraftMaterialNeed {
  ItemID?: string;
  itemId?: string;
  Qty?: number;
  qty?: number;
}

export interface CraftingRecipeView {
  id: string;
  name: string;
  profession: string;
  category: string;
  requiredLevel: number;
  craftTime: number;
  result: string;
  materials: CraftMaterialNeed[];
  owned: Record<string, number>;
  status: string;
  station: string;
}

export interface ProfessionView {
  id: string;
  name: string;
  kind: string;
  title: string;
  xp: number;
  level: number;
  next?: number;
  active: boolean;
}

export interface MaterialView {
  id: string;
  name: string;
  qty: number;
  desc?: string;
  region?: string;
  type?: string;
  rarity?: string;
}

export interface CraftingState {
  gold?: number;
  goldBalance?: number;
  knowledgeToken?: number;
  knowledgePoints?: number;
  professions?: ProfessionView[];
  recipes?: CraftingRecipeView[];
  materials?: MaterialView[];
  merchants?: Array<{ id: string; name: string; items: Array<{ itemId: string; name: string; price?: number; buyPrice?: number; sellPrice?: number; rarity?: string }> }>;
  goldHistory?: Array<{ source: string; amount: number; destination: string; timestamp?: number }>;
  guildContrib?: Array<{ playerId: string; itemId: string; qty: number; at?: number }>;
  codex?: Array<{ id: string; name: string; desc?: string; region?: string; rarity?: string }>;
  queue?: Array<{ id: string; recipeId: string; result: string; quality: string; readyAt: number }>;
  stallOpen?: boolean;
  stall?: Array<{ itemId: string; qty: number; price: number }>;
  stalls?: Array<{ playerId: string; items: Array<{ itemId: string; qty: number; price: number }> }>;
  orders?: Array<{ id: string; requester: string; crafter: string; recipe: string; reward: number; status: string }>;
  workshop?: boolean;
  festival?: boolean;
  festivalName?: string;
  economy?: { goldCreated?: number; goldRemoved?: number; tradeVolume?: number; craftVolume?: number };
  durability?: number;
  resetsLeft?: number;
  achievements?: string[];
  titles?: string[];
  toId?: string;
}

export interface FishState {
  spotId?: string;
  place?: string;
  targetA?: number;
  targetB?: number;
  prompt?: string;
  caught?: boolean;
  reward?: string;
  anim?: string;
}


