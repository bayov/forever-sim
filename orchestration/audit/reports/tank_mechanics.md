# tank_mechanics pass 31 (2026-11-24, master eceb090)

Why due: the warrior Balanced and Max survival requests gained a Stoneform line (dfd015925), and client
1.60.1.71037 -> 1.60.1.71091. In range on tank code: 8d5fdd664 and 7a2311b36 (Troll Regeneration and Forever
Stoneform in `sim/core/racials_forever.go`, the fixes for pass 30's -1 and -2).

## Get current

- Client 71091 vs 71037 (`tools/sources/db2diff.py 71037 71091`, 75 spells): no Protection warrior, Protection
  paladin, bear or Warden ability, talent or racial row changed. Stoneform 20594 and Regeneration 20555 unchanged.
- Tank-adjacent rows: Judgement 20271 and Execute 5308/20658/20660/20661/20662 gain Attributes_4 0x800000 (Suppress
  Weapon Procs; Execute also loses Attributes_3 0x10000). Matches the Nov 23 hotfix "a few spells and abilities were
  causing more weapon procs than intended" (t/2403040 post 15, tier 1). No tank preset casts Execute, and the paladin
  presets judge Righteousness/Fury, already ranged (no weapon procs). No tank effect. Execute is client_update's
  open finding (client_update-20261120-1).
- 25752/25757/25945 (Judgement of Light/Wisdom/Justice, duration 10 -> 40 s): level 1 NPC copies nothing triggers; the
  sim's judgements use other ids. The 1289xxx/1331xxx druid spells (Maul, Swipe, Demoralizing Roar, Berserk) are
  NPC spells; no sim code references them.
- Blizzard: hotfix posts 14 (Nov 20) and 15 (Nov 23) read. Tank-relevant: none beyond the weapon-proc line above
  (Stormshroud 3P and Feralheart 4P are Energy; Mortal Strike spellbook fix). Dev notes thread unchanged since Nov 4.
  The ledger has no 71091 section yet (client_update's job).

## The fixes (8d5fdd664, 7a2311b36, dfd015925)

| Racial | Client 71091 | Sim | Verdict | Tier |
|---|---|---|---|---|
| Regeneration 20555 | aura 88 +20%, aura 116 20% in combat | 2 s tick of outOfCombatHealthRegen x 1.2 x 0.2 from the first damage taken; reads Spirit, not health | OK (size rests on Wrath's Spirit tables, +-35%) | 1 (rows), 3 (regen table) |
| Stoneform 20594 | aura 87 -10% misc 1, 8 s, 180 s, GCD | Physical SchoolDamageTakenMultiplier x0.9, 8 s, 3 min, GCDDefault, survival MCD | OK | 1 |

Measured (tankbench, warrior_protection_balanced with race swapped, 3000 fights, labeled seed 1, part off via
DASH_RACIALS_OFF):

| Build | Max threat | Balanced | Max survival | SafeDPS | Healing needed |
|---|---|---|---|---|---|
| Troll, regen off | 3597 | 3022 | 1294 | 1943 | 384 |
| Troll, regen on | 3606 | 3047 | 1370 | 1943 | 373 |
| Dwarf, Stoneform off | 3561 | 2984 | 1251 | 1925 | 385 |
| Dwarf, Stoneform from 20 s | 3539 | 2964 | 1239 | 1913 | 383 |

Troll matches pass 30's estimate. Stoneform casts 1.0 a fight on this request and costs a GCD (SafeDPS -12) for
-2 healing needed, so on the request's own encounter it loses at all three weights; dfd015925's sweep over the six
rotopt encounters (10,000 iterations) has +10 Balanced and +44 Max survival from 20 s. That is a timing call for
tank_rotations, not a mechanic error; noted, not a finding.

- Scoring replay: neither racial reads health (Stoneform scales damage, the regen reads Spirit), and the regen's
  heals land as replay heal events (`sim/core/tank_score.go:121`). OK.
- Paladin and Warden: their requests use autocastOtherCooldowns, and survival cooldowns never autocast in a scoring
  run (`survivalCooldownDue`, `sim/core/major_cooldown.go:152`), so a Dwarf paladin or Warden never casts
  Stoneform. New finding -4 (low).

## Tests

- Troll Regeneration: only the leveling ratio test (TestLevelingTrollRegen, 20%). Stoneform: none. Added to -3.

## Findings

- New: tank_mechanics-20261124-4 (low, worker fix): Dwarf paladin / Warden never cast Stoneform.
- Resolved: -20261124-1 (Troll Regeneration) and -20261124-2 (Stoneform, warrior part).
- Still open: -3 (racial tests, updated), 20261117-1, 20261108-1, 20261107-1.
