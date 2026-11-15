# tank_rotations pass 15

Master 25ce81de7 (2026-11-15), against pass 14's baseline (master 70ea588bb). 1,000 iterations, default seed, every
tank request on all seven boss files. Quiet: no new finding.

## What changed

Since 70ea588bb (gate: 15 watched files), the tank inputs that moved:

- **PR #130 (Nov 13 hotfixes, 9ec69044f / d4325b775)**: Windfury Totem rolls on every landed main-hand melee hit, white
  or ability, behind its 100 ms cooldown (`windfuryTotemRollsOnForever`, sim/core/buffs.go;
  windfury_totem_procs_forever.go removed). Judgement of Command rolls on melee defense again (sim/paladin/judgement.go,
  soc.go).
- **baa363615**: Two-Handed Weapon Specialization's Holy part follows the defense type (seal hits, Holy Strike, JoC,
  Consecration), not the Holy school (sim/paladin/talents.go, consecration.go). The Prot paladins are one-handed, so
  this mostly doesn't reach them.
- **f2a5dfe25**: Prot warrior Balanced and Max survival re-tuned for the totem rule: Heroic Strike from 40 Rage (was
  50), Revenge ahead of Shield Slam, Thunder Clap refreshed under 8 s (was 2 s). The commit reports +16.1 / +15.0
  (Balanced) and +24.6 / +35.4 (Max survival) at 20k, two seeds.
- rotopt tank templates gained Windfury timing knobs, all off by default: `wfGate` (warrior_tank, druid_tank) and
  `wfIcdGate` / `swingGap` (paladin_tank). The shipped APLs don't use them.
- Wild Growth tick coefficient 0.033 -> 0.066 (no tank casts it). Windfury Weapon now exposes its ICD to the APL (no
  tank request uses Windfury Weapon).

No tools/encounters, tools/tankbench/main.go or tools/tankscore change.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for 25ce81de7. Scoring mean (six bosses without Ragnaros) at the preset's weight:

| request | score old -> new | SafeDPS old -> new | healing needed old -> new | cause |
|---|---|---|---|---|
| druid_bear_max_threat | 2787.2 -> 3121.0 | 1680.7 -> 1850.8 | 634 -> 642 | totem on Lacerate, Primal Bite, Swipe |
| druid_bear_balanced | 1862.8 -> 2144.7 | 1572.5 -> 1720.0 | 462 -> 472 | same |
| druid_bear_max_survival | 261.2 -> 351.2 | 1231.3 -> 1324.3 | 279 -> 288 | same |
| paladin_protection_max_threat | 3248.8 -> 3272.8 | 1739.0 -> 1751.5 | 272 -> 273 | totem on Holy Strike, JoC melee |
| paladin_protection_balanced | 2809.3 -> 2836.5 | 1710.8 -> 1722.2 | 244 -> 241 | same |
| paladin_protection_max_survival | 1780.2 -> 1833.5 | 1629.3 -> 1645.0 | 202 -> 199 | same |
| shaman_warden_max_threat | 2562.0 -> 2645.7 | 1671.8 -> 1715.7 | 843 -> 850 | totem on Stormstrike |
| shaman_warden_balanced | 1224.2 -> 1282.8 | 1523.3 -> 1558.0 | 659 -> 664 | same |
| shaman_warden_max_survival | -2337.0 -> -2221.2 | 1464.0 -> 1499.5 | 639 -> 633 | same |
| warrior_protection_balanced | 3420.8 -> 3443.2 | 2201.2 -> 2211.8 | 345 -> 344 | f2a5dfe25 re-tune |
| warrior_protection_max_survival | 1776.0 -> 1795.2 | 1796.2 -> 1805.8 | 217 -> 217 | f2a5dfe25 re-tune |
| warrior_protection_max_threat | bit-identical | | | |
| warrior_protection (legacy) | bit-identical | | | |

Every move matches a ruled mechanic change (standards 17, the Nov 13 hotfix) or a preset re-tune with its own A/B.
The bear numbers agree with docs/tank_druid.md's own re-measure (Max threat 3,110.0, Balanced 2,120.2 at 10k). The
warrior totem rule doesn't change Max threat because warrior strikes already rolled under the Nov 11 rule. No finding.

## Open item: main-hand ability timing for Windfury (redone on the Nov 13 rule)

Pass 14 measured this on 70ea588bb, which was before PR #130, so it's redone here. The workers' re-searches on this
master already cover three of the four tanks:

- warrior (docs/tank_warrior.md, "Re-tuned on Windfury Totem rolling on every main-hand hit"): `wfGate` 1 and 2 lose
  at every weight (-0.3 to -89.9, 20k, two seeds);
- bear (docs/tank_druid.md): `wfGate` 1/2/3 lose (wfGate=2 at 3k: Balanced -2.2 ± 4.0, Max threat -0.4 ± 1.8);
- paladin (docs/tank_paladin.md, paladin_presets.sh): `wfIcdGate` -0.7 to +0.0 (10k, two seeds), `swingGap` -6.3 to +0.9.

The Warden wasn't re-searched, so this pass tested it. Stormstrike is its only main-hand strike. The helper is
`tools/audit/tank_rotations/mk_warden_wfgate.py`, and it adds one of three gates to Stormstrike:

- **after**: not in the 100 ms after a main-hand swing;
- **spread**: also not in the 100 ms before the next one;
- **icd**: only while the totem's 100 ms internal cooldown (10610) is ready.

Paired `abtest` (gate minus preset), 3,000 iterations, scoring mean at the preset's weight, seed 1 / seed 2:

| request | after | spread | icd |
|---|---|---|---|
| shaman_warden_max_threat | -4.7 ± 2.2 / -5.8 ± 2.6 | -10.2 ± 3.2 / -6.5 ± 3.2 | -1.4 ± 1.2 / +0.5 ± 1.5 |
| shaman_warden_balanced | +3.0 ± 4.6 / -1.6 ± 4.3 | -1.8 ± 5.8 / -0.7 ± 5.7 | +2.5 ± 2.3 / +3.7 ± 2.3 |
| shaman_warden_max_survival | +11.8 ± 11.0 / -17.3 ± 10.9 | +6.3 ± 14.4 / -12.9 ± 14.4 | +7.1 ± 6.0 / -1.5 ± 6.9 |

Balanced icd at 10,000 iterations comes out +0.9 ± 1.2 / +2.0 ± 1.2. That's about 0.1%, within noise. The Enhancement
try (hold Stormstrike for Windfury Weapon's cooldown) doesn't apply because all three Warden requests use Rockbiter.
Nothing passes the bar, so there's no finding. The open item is done for every tank, and the orchestrator can remove it
from passes/tank_rotations.md (this pass can't edit that doc). Pass 14 cited a `mk_wfgate.py` helper, but it isn't in
the repo.

## Rotation lines

The only rotation change is the warrior re-tune (f2a5dfe25). Its three lines do what their notes say. HS at 40 Rage
still keeps Rage for Shield Slam and Revenge, because the baseline's Balanced SafeDPS and healing needed both
improved. The worker's A/Bs (20k, two seeds) are above the bar, so this pass didn't repeat them. The other presets'
earlier A/Bs still stand.

## Open findings

- 20261104-11 (warrior Max survival armor trinket, Broodlord): user call. The re-tune kept 990617 held 35 s.
- 20261104-3 (paladin Bulwark 13 s, Broodlord): user ruling. Still 10 s.
