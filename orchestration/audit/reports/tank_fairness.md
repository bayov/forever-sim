# tank_fairness pass 34

Branch phase0-tankheal 3e9748641 (last: master 85691655a). First pass on the Phase 0 tank gear (standards 43). All 12
presets re-simmed: tankbench, 3,000 iterations, seeds 1 and 2, six scored bosses (`tools/audit/tank_fairness/bench.sh`).
Baseline refreshed: ref/tank_fairness/summary.tsv (seed 1).

## Same terms

Requests changed since the last pass only in equipment (Phase 0 sets), bear rotations (all three), Warden Max threat
and Max survival rotations, paladin and Warden Max threat talents. Consumes, buffs, debuffs, level 60, Engineering /
Blacksmithing, reaction 200 ms: unchanged and equal. Races: warrior and paladin Undead, bear Night Elf / Night Elf /
Tauren, Warden Night Elf / Night Elf / Undead (open -20261105-2, now in the gear too). Only the warrior casts Stoneform
(open -20261124-1).

## The numbers (two-seed mean, own column; synthetic from pass 33 in brackets)

| weight | order | scores |
|---|---|---|
| Max threat (1/3) | Bear, Warrior, Paladin, Warden | 5603 / 4455 / 3913 / 3400 (4537 / 4097 / 3282 / 2606) |
| Balanced (1) | Bear, Warrior, Paladin, Warden | 3696 / 3058 / 2952 / 2024 (3342 / 3251 / 2863 / 1244) |
| Max survival (3) | Warrior, Bear, Paladin, Warden | 1079 / 1047 / 780 / -1401 (1617 / 1301 / 1875 / -2277) |

TPS at Balanced: bear 2612, warrior 2137, paladin 1858, Warden 1738. Healing needed at Balanced: paladin 449, warrior
524, Warden 700, bear 711.

- The bear leads Balanced by 21% over the warrior (3% on the synthetic sets): gear. The bear's Phase 0 sets gain at Max
  threat (+647 in its notes, Manual Crowd Pummeler / Hammer of Thundering Hooves and leather Agility), the warrior's
  Balanced set loses 183 to synthetic, and the warrior list doesn't match the bench (tank_gear-20261126-1).
- Max survival reorders: the paladin falls from first (1875) to third (780). Its set has 8271 health against the
  warrior's 10269 at the same 382 defense; Flask of the Titans wins it back +120 (new -20261126-1). Everyone drops at
  Max survival because the real sets carry less armor and defense than the synthetic ones (bear and paladin notes).
- The Warden: 45% behind the bear at Balanced and far below 0 at Max survival, as on the synthetic sets (standards 40:
  a class gap), plus the illegal race (open).
- Logs: not re-compared this pass (gear only); the Anniversary main-tank mix in the boss files is unchanged.

## What each preset rewards (two-seed mean, rows = build, columns = weight)

| build | Max threat | Balanced | Max survival |
|---|---|---|---|
| bear Balanced / Max threat / Max survival | 5013 / 5603 / 3536 | 3696 / 3889 / 2914 | -254 / -1250 / 1047 |
| paladin | 3780 / 3913 / 3328 | 2952 / 2880 / 2691 | 468 / -221 / 780 |
| warrior | 4090 / 4455 / 3234 | 3058 / 2891 / 2695 | -37 / -1802 / 1079 |
| Warden | 3315 / 3400 / 2822 | 2025 / 1752 / 1767 | -1847 / -3192 / -1401 |

Each preset wins its own weight except the bear's Balanced: its Max threat build is 193 ahead at w 1 (open
-20261125-1, new evidence). The three builds differ for every spec.

## Price coverage (requested with this pass)

Every feasible entry of the tank lists and the finished healer lists (priest_holy, priest_discipline, paladin_holy,
shaman_restoration; druid_restoration pending) checked for a price or a stated reason: -20261126-2. 11 crafted picks
have no price; 5 priced items lack the list's own number; 3 picks have no source note.

## New findings

- tank_fairness-20261126-1 (medium): paladin Max survival should take Flask of the Titans on Phase 0 gear (+120).
- tank_fairness-20261126-2 (medium): price coverage gaps (needs a user ruling for crafts with unsourced plans).

---

## Master's pass, carried over from origin/master: tank_fairness pass 34: 2026-11-26, master ab3fff7007

Merged in from master's audit loop (synthetic-gear tank/healer rows; numbers there are master's baseline, not the
Phase 0 one above). Kept for the verdicts and finding notes that exist only on master.

Due: 38 watched files changed since bfbee78. They are the warrior leveling batch (565f32155, 270731faa, 143bf4f72,
8f5b59325, 1223e764e and others): warrior spells take the rank known at the character's level (`knownRank`; at 60 the
same top ranks as before: Shield Slam 23925, Revenge 11601, Sunder 11597, Thunder Clap 11581), Defensive Stance only once
known, and core rules that apply only inside a leveling session (swings from behind can't be avoided, Cannibalize,
kill hooks). Stunned targets can no longer block, dodge or parry (`sim/core/spell_outcome.go`); no tank preset is
stunned on the six scoring bosses. No change to `tools/tankbench`, its requests or `tools/encounters`.

### Method

tankbench built in the master worktree (/tmp/audit-cache/cli/ab3fff700700/tree, `--tags=with_db`), `bench.sh` on the
12 preset requests x six scoring bosses, 3,000 iterations, seed 1; `summary.py --tsv` diffed against
ref/tank_fairness/summary.tsv: all 84 rows (72 boss rows and 12 means) identical. Baseline kept.

### Ranking (unchanged from pass 33)

| weight | order | scores |
|---|---|---|
| Max threat | Bear, Warrior, Paladin, Warden | 4537 / 4097 / 3282 / 2606 |
| Balanced | Bear, Warrior, Paladin, Warden | 3342 / 3251 / 2863 / 1244 |
| Max survival | Paladin, Warrior, Bear, Warden | 1875 / 1617 / 1301 / -2277 |

TPS (Balanced): bear 2371, warrior 1987, paladin 1531, Warden 1367. Healing needed (Balanced): paladin 233, warrior 341,
bear 468, Warden 664. The explanations and the comparison with the Anniversary logs in pass 33 stand
(orchestration/fairness_tanks.md).

### What each preset rewards

Unchanged from pass 33: every preset wins its own weight except the bear's Balanced, which ties its Max threat build
at w 1 (finding -20261125-1). The three builds remain distinct for every spec.

### Same terms

Requests unchanged; the open unequal terms stand: -20261105-1 (paladin potion), -20261105-2 (Warden races),
-20261124-1 (Stoneform on paladin and Warden Dwarf rows).

### Findings

- New: none.
- Open, unchanged: -20261125-1, -20261124-1, -20261105-2, -20261105-1, -20261104-2, -20261104-5.
