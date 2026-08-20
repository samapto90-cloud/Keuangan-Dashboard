export type SubjectAccuracy = {
  correct: number;
  total: number;
  accuracy: number;
};

export type AchievementView = {
  id: string;
  name: string;
  description: string;
  icon: string;
  rewardXp: number;
  rewardCoins: number;
  unlocked: boolean;
};

export type TitleView = {
  id: string;
  name: string;
  unlocked: boolean;
};

export type PlayerProfileView = {
  playerId?: string;
  userId: string;
  username: string;
  email?: string;
  avatar: string;
  title: string;
  level: number;
  xp: number;
  xpIntoLevel: number;
  xpToNext: number;
  xpNextAt: number;
  coins: number;
  trophies?: number;
  totalMatches: number;
  wins: number;
  losses: number;
  draws: number;
  winRate: number;
  totalQuestions: number;
  correctAnswers: number;
  wrongAnswers: number;
  timeoutAnswers: number;
  accuracy: number;
  currentWinStreak: number;
  bestWinStreak: number;
  dailyStreak: number;
  subjectAccuracy: Record<string, SubjectAccuracy>;
  achievements: AchievementView[];
  titles: TitleView[];
  avatars: string[];
  rankXp?: number;
  rankWins?: number;
  rankAccuracy?: number;
  rankTier?: string;
  rankDivision?: string;
  rankRr?: number;
  rankLabel?: string;
  rankIndex?: number;
  peakTier?: string;
  peakDivision?: string;
  peakRr?: number;
  rankedWins?: number;
  seasonId?: string;
  seasonName?: string;
  myRank?: number;
};

export type MatchHistoryItem = {
  matchId: string;
  rank: number;
  position: number;
  correct: number;
  wrong: number;
  timeout: number;
  xpEarned: number;
  coinsEarned: number;
  won: boolean;
  mode?: string;
  rrBefore?: number;
  rrChange?: number;
  rrAfter?: number;
  createdAt: number;
};

export type DailyStatus = {
  claimed: boolean;
  day: number;
  coins: number;
  streak: number;
  date: string;
};

export type RewardEvent = {
  userId?: string;
  xp?: number;
  coins?: number;
  trophies?: number;
  trophyTotal?: number;
  won?: boolean;
  levelBefore?: number;
  levelAfter?: number;
  levelUp?: boolean;
  achievements?: string[];
  rankLabel?: string;
  rr?: number;
  rrDelta?: number;
  rankUp?: boolean;
  rankDown?: boolean;
};

export const AVATAR_FACE: Record<string, string> = {
  bird: "🐦",
  batik: "🧵",
  gunung: "⛰️",
  candi: "🛕",
  wayang: "🎭",
  kapal: "⛵",
};

export function fmtPct(n: number): string {
  if (!n) return "0%";
  return `${Number(n).toFixed(n % 1 === 0 ? 0 : 1)}%`;
}

export function xpBarWidth(p: PlayerProfileView): number {
  const span = (p.xpIntoLevel || 0) + (p.xpToNext || 0);
  if (span <= 0) return 100;
  return Math.max(4, Math.min(100, Math.round((p.xpIntoLevel / span) * 100)));
}
