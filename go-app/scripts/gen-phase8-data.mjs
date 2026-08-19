import { writeFileSync } from "fs";
import { dirname, join } from "path";
import { fileURLToPath } from "url";

const dir = join(dirname(fileURLToPath(import.meta.url)), "../mmo/data");

const guardiansSrc = [
  ["ragha", "Ragha", "Guardian of Mist", "MIST", "NORMAL"],
  ["gravon", "Gravon", "Guardian of Stone", "STONE", "ELITE"],
  ["velra", "Velra", "Guardian of Thorns", "NATURE", "ELITE"],
  ["kairoth", "Kairoth", "Guardian of Flame", "FIRE", "BOSS"],
  ["nymra", "Nymra", "Guardian of Moon", "MOON", "BOSS"],
  ["vorak", "Vorak", "Guardian of Dust", "DUST", "BOSS"],
  ["zeran", "Zeran", "Guardian of Wind", "WIND", "BOSS"],
  ["malvek", "Malvek", "Guardian of Ash", "ASH", "BOSS"],
  ["torga", "Torga", "Guardian of Iron", "IRON", "BOSS"],
  ["selka", "Selka", "Guardian of Rain", "WATER", "BOSS"],
  ["dravon", "Dravon", "Guardian of Thunder", "THUNDER", "BOSS"],
  ["morgha", "Morgha", "Guardian of Roots", "EARTH", "BOSS"],
  ["kelran", "Kelran", "Guardian of Sand", "SAND", "BOSS"],
  ["arvok", "Arvok", "Guardian of Ice", "ICE", "BOSS"],
  ["nyxara", "Nyxara", "Guardian of Shadow", "SHADOW", "BOSS"],
  ["torven", "Torven", "Guardian of Crystal", "CRYSTAL", "ANCIENT"],
  ["veyra", "Veyra", "Guardian of Storm", "STORM", "ANCIENT"],
  ["gharon", "Gharon", "Guardian of Ruins", "RUIN", "ANCIENT"],
  ["lurak", "Lurak", "Guardian of Beast", "BEAST", "ANCIENT"],
  ["sorya", "Sorya", "Guardian of Flame", "FIRE", "ANCIENT"],
  ["varok", "Varok", "Guardian of Night", "NIGHT", "ANCIENT"],
  ["zavra", "Zavra", "Guardian of Silence", "SILENCE", "ANCIENT"],
  ["orvak", "Orvak", "Guardian of Gravity", "GRAVITY", "ANCIENT"],
  ["merra", "Merra", "Guardian of Echo", "ECHO", "ANCIENT"],
  ["tharos", "Tharos", "Guardian of Sky", "SKY", "ANCIENT"],
  ["ragna", "Ragna", "Guardian of Bloodstone", "BLOODSTONE", "ANCIENT"],
  ["velkor", "Velkor", "Guardian of Time", "TIME", "ANCIENT"],
  ["aranya", "Aranya", "Guardian of Light", "LIGHT", "ANCIENT"],
  ["gorven", "Gorven", "Guardian of Abyss", "ABYSS", "ANCIENT"],
  ["sylra", "Sylra", "Guardian of Dreams", "DREAM", "ANCIENT"],
  ["korvan", "Korvan", "Guardian of Chaos", "CHAOS", "ANCIENT"],
  ["elyra", "Elyra", "Guardian of Stars", "STAR", "ANCIENT"],
  ["avaron", "Avaron", "Guardian of the Celestial Gate", "CELESTIAL", "CELESTIAL"],
];

const regions = ["village", "forest", "valley", "plains", "canyon", "temple", "ruins", "celestial"];
const arenas = ["fog", "earthquake", "nature", "fire", "moon", "dust", "wind", "ash", "iron", "rain", "thunder", "roots", "sand", "ice", "shadow", "crystal", "storm", "ruin", "beast", "fire", "night", "silence", "gravity", "echo", "sky", "bloodstone", "time", "light", "abyss", "dream", "chaos", "stars", "celestial"];

const guardians = guardiansSrc.map((row, i) => {
  const n = i + 1;
  const ch = "ch" + String(n).padStart(2, "0");
  const region = regions[Math.min(7, Math.floor(i / 5))];
  return {
    id: row[0],
    name: row[1],
    title: row[2],
    chapterId: ch,
    region,
    level: 4 + Math.floor(i * 1.4),
    element: row[3],
    difficulty: row[4],
    story: `${row[1]} menjaga jalan menuju Cahaya.`,
    bossArena: arenas[i],
    skills: n === 1 ? ["mist_strike", "ground_shock", "summon_imp", "dark_wave"] : ["ground_shock"],
    lootTable: "lt-" + row[0],
    uniqueItem: row[0] + "_sigil",
    uniqueItemName: `${row[1]}'s ${row[2].replace("Guardian of ", "")} Sigil`,
    phases: n >= 16 ? 3 : n >= 4 ? 2 : 1,
    x: ((i % 3) - 1) * 8,
    z: 28 + i * 3.6,
    intro: `Kabut cerita menyingkap ${row[1]}.`,
    encounter: "Jejaknya muncul di arena.",
    defeat: `${row[1]} tunduk. Jalan berikutnya terbuka.`,
    aftermath: "Dunia menjadi lebih terang.",
  };
});

writeFileSync(join(dir, "guardians.json"), JSON.stringify(guardians, null, 2));
console.log("wrote", guardians.length, "guardians");
