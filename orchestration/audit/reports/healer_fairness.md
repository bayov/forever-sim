# Healer fairness pass 1 (2026-11-04, master c8f5858bd)

Result: findings (4 new: 1 high, 1 medium, 2 low). This was the forced first pass after H3. It began on a1b1163d2. When
the raid damage fix (healer_mechanics-20261104-1, b2c16f6) landed, both roles were re-benched on c8f5858bd, each
with a healbench built from a worktree of that commit. The attribution experiments stayed on a1b1163d2: the
tank role moved by at most 0.05 between the two.

## 1. Same terms (the 9 requests in tools/healbench/req)

Identical in all 9:

- level 60, Engineering + Alchemy
- consumables: Flask of Supreme Power, Lobster Stew, Greater Mageblood, Greater Arcane, Greater Intellect, Sages,
  Brilliant Mana Oil, Major Mana Potion, Demonic Rune, Cerebral Cortex
- raid buffs and debuffs, Kings and Improved Wisdom, armor enchants, the request's encounter and options

healbench overrides these for every healer (`tools/healbench/main.go:276-325`):

- reaction 0.3 s ±25%
- stopcast 0.5
- heal assignment
- encounter
- seed

| Difference | Who | Verdict |
|---|---|---|
| Race | Human for the priest and paladin, Skyborne Windshaper for the shaman and druid | Each class takes its best legal race; priest and paladin can't be Skyborne (`ui/core/proto_utils/utils.ts:1042-1045`). The shaman's race was never swept: Orc/Troll/Tauren/Dwarf are within 0.002 of Windshaper here. Fine. |
| Slot 16 | The priest has a synthetic wand (+21 weighted, +2.4% of budget); the others have stat-less relics | Finding 3 (carry -0.003 without it) |
| Weapon enchant | Paladin MH Mighty Intellect + OH Wisdom; the others staff Spellblasting | Weapon type; fine |
| Trinkets | Each borrows its class's DPS synthetic pair (45.5-46.0 weighted) | Equal; fine |
| Gear budget | 901-903 weighted for all four (standards 1b, 35); raw mixes differ by their Wowhead lists | Fine (healing+SP: druid 892, shaman 774, paladin 765, priest 704) |
| Mana cooldowns | Shaman Mana Tide, druid Innervate; priest and paladin none | Class kit; nobody is mana-bound at carry anyway |
| Power Infusion | An action in priest_holy.json the talents never take | Finding 3 (0 effect) |

## 2. The numbers (healbench -carry, 1,000 iterations, each file at its own length, c8f5858bd)

Tank role (in brackets: the change against the H3 table / a1b1163d2):

| Boss | Shaman | Holy priest | Disc priest | Paladin | Druid |
|---|---|---|---|---|---|
| Onyxia | 1.154 (+.013) | 1.240 (+.010) | 1.225 (+.030) | 1.097 (+.016) | 1.164 (-.001) |
| Broodlord | 0.839 | 0.891 | 0.859 | 0.754 | 0.732 |
| Chromaggus | 0.878 | 1.002 | 0.818 | 0.837 | 0.857 |
| Golemagg | 1.752 (+.036) | 1.865 (-.050) | 1.802 (-.024) | 1.647 (+.022) | 1.649 |
| Magmadar | 1.793 | 2.062 (-.027) | 1.930 | 1.637 | 1.691 |
| Ragnaros | 0.511 | 0.578 | 0.553 | 0.510 | 0.517 |
| Vaelastrasz | 0.793 (+.018) | 0.905 | 0.854 | 0.711 | 0.672 (-.016) |
| **Mean** | **1.103** | **1.220** | **1.149** | **1.028** | **1.040** |

So the tank role does move a little on c8f5858bd. In the tank bench the raid damage falls on the healer too, so
the fix touches it.

Raid role (old a1b1163d2 in brackets):

| Boss | Shaman | Holy priest | Paladin | Druid |
|---|---|---|---|---|
| Onyxia | 0.923 (0.678) | 0.947 (0.709) | 0.984 (0.702) | 0.977 (0.711) |
| Broodlord | 0.879 (0.817) | 0.925 (0.879) | 0.906 (0.838) | 0.906 (0.828) |
| Chromaggus | 1.082 (0.906) | 1.035 (0.911) | 1.081 (0.970) | 1.103 (0.953) |
| Golemagg | 2.645 (1.913) | 2.542 (2.064) | 2.699 (2.067) | 2.656 (2.021) |
| Magmadar | 1.772 (1.540) | 2.007 (1.693) | 1.919 (1.500) | 1.869 (1.605) |
| Ragnaros | 1.203 (1.171) | 1.304 (1.255) | 1.302 (1.244) | 1.287 (1.240) |
| Vaelastrasz | 1.110 | 1.139 | 1.002 | 1.092 |
| **Mean** | **1.373** (1.162) | **1.414** (1.236) | **1.413** (1.189) | **1.413** (1.207) |

- The raid role is now nearly level. Priest, paladin and druid are at 1.413-1.414, and the shaman is 3% behind,
  last on 4 of 7 files.
- No raid healer runs out of mana. Per-healer means are EHPS / overheal / EH per mana:
  - shaman 271 / 45% / 3.56
  - priest 286 / 43% / 2.63
  - paladin 249 / 21% / 5.48
  - druid 292 / 72% / 2.04
- Full tables, per-file mana, OOM and death floors are in `ref/healer_fairness/raid_c8f5858.tsv`.

**The spread.** No cell is 15% off its file's mean except the tank role's Holy priest:

- +11% over the field's mean (1.220 against 1.103 for the shaman)
- up to +16% on Vaelastrasz and Magmadar against the shaman, +27% against the paladin

### Why the Holy priest leads the tank role (attribution, a1b1163d2, 1,000 iterations, all 7 files)

Its lead over the shaman there was +0.135. Each ablation's mean delta to the priest:

| Change | Δ priest mean |
|---|---|
| Bench with 7 idle fillers in tank role (overlay) | **-0.085** (Magmadar -0.177) |
| Same overlay: shaman / paladin / druid | +0.003 / +0.010 / -0.004 |
| No Prayer of Mending in the rotation | -0.081 (with fillers: +0.007) |
| No Holy Binding | -0.081 (with fillers: -0.012) |
| No Inspiration | -0.066 (with fillers: -0.062) |
| No Spiritual Healing | -0.031 (with fillers: -0.037) |
| Meditation / Improved Renew / Holy Spec / Spiritual Guidance / Twin Disc / Inner Focus | -0.021 / -0.013 / -0.011 / -0.011 / -0.009 / -0.007 |
| No wand (slot 16 empty) | -0.003 |
| Inspiration on item armor only (as the shaman's Ancestral Fortitude) | -0.007 to -0.010 |

The pieces of the +0.135:

- **0.085 is the bench (finding 1).** With no one else in a tank-role raid, Prayer of Mending can only jump
  between the tank and the priest. Holy Binding's self-heal sets it off on the priest at once, so the tank takes
  about 3 PoM heals a cast instead of 1.
- **About 0.02 is the presets' "missing ≥ Expected Healing" gate.** It costs the slow-heal classes more (finding
  2).
- **About 0.01 is the armor asymmetry.** Inspiration is applied to all armor and Ancestral Fortitude to item armor
  only. This is already open as healer_mechanics-20261104-3, which says Ancestral Fortitude should be all armor.
- **The rest, about 0.02, is kit.** Most of it is Spiritual Healing (0.037 with fillers). Inspiration is worth
  0.06, but the shaman has the same buff through Ancestral Healing.

So the shaman worker's reading is partly right: the shaman's ranks, mana gates and cooldowns don't close it. Every
shaman lever tried was within ±0.01 except the gate:

- Lesser Healing Wave gates
- Nature's Swiftness thresholds
- Healing Wave 10 only
- no Mana Tide
- Thundering Strikes
- races

But most of the lead is the bench, not the class. With the bench and the gates fixed, the expected order is
priest about +0.03 to +0.05 over shaman, then druid and paladin. That needs a re-bench after the fixes.

**Mana is not the limit for any tank healer.** At its carried damage each ends with 39-89% mana, and a mana gate
of 0 changes nothing (`ref/healer_fairness/mana_at_carry_a1b1163.tsv`).

Effective stats on Magmadar (healing / crit):

- priest 1153 / 19.2%
- shaman 1109 / 23.0%
- paladin 1030 / 26.1%
- druid 1226 / 19.4%

Spiritual Guidance adds exactly 0.25 × Spirit. No stat bug.

### Against the logs

The boss files keep healers by class only as counts, about 5.1 priests, 3.0 shamans, 1.7 paladins and 1.2 druids of
11 a kill. They keep no HPS by class, so the sim's ranking can't be checked against the logs' (finding 4). The
composition fits the sim's raid role, where the classes are within 3% of each other: raids bring priests by number,
not because a priest out-heals.

## 3. What a role rewards (cross-check, 500 iterations, Onyxia / Magmadar / Vaelastrasz)

Means over the three files:

| Bench | Shaman | Holy priest | Disc priest | Paladin | Druid |
|---|---|---|---|---|---|
| Raid presets in the tank role (a1b1163d2) | 1.021 | 1.356 | - | 1.135 | 1.139 |
| Tank presets in the raid role (c8f5858bd) | 1.191 | 1.033 | 1.017 | 1.319 | 1.006 |

On their own role's presets, the means over the same three files are:

- tank role: shaman 1.238, Holy priest 1.407, paladin 1.150, druid 1.184
- raid role: shaman 1.268, Holy priest 1.364, paladin 1.302, druid 1.313

What this shows:

- **Each preset is fitted to its role.** Every healer does worse on the other role's preset, except the paladin's
  tank preset in the raid role. It beats the paladin's raid preset (1.319 against 1.302, within the 500-iteration
  screen's noise), so the healer_rotations auditor may want to look at the raid preset.
- **The H3 table is like for like.** Each column is the preset for its role, judged at that file's own length
  (the tank time for tank, the logs' median kill for raid).

## Reproduce

- Build (worktree of the audited commit):
  `protoc -I=./proto --go_out=./sim/core ./proto/*.proto && go build --tags=with_db -o healbench ./tools/healbench`
- Run, one per healer and file:
  `healbench -encounter tools/encounters/<boss>.json -role <tank|raid> -carry -iters 1000 tools/healbench/req/<req>.json`
- Filler overlay: `ref/healer_fairness/tank_fillers_overlay.diff`, run with `HB_TANK_FILLERS=1`.

## Open items for the next pass

- Re-bench the tank role after finding 1's fix and give the ranking.
- Re-check the gate (finding 2) on each preset after any re-tune.
- The Discipline priest OOMs on one tank file at ×1.00 damage (about 106 s, a1b1163d2): check after the fixes.
