# Notes round 2026-11-11: every class after shaman (sim changes only)

Brief: `briefs/classes.md`. Landed on master; the report is **not** re-published (a full-run session re-sims every
build next). No combat logs are in this environment, so every value with no server row is **notes only**
(standards 37) and is listed in the class's `orchestration/round_70753/<class>.md` notes-only list with the row or log
that would confirm it.

180 s numbers below are this round's own sims: the report requests through the sim CLI, 20,000 iterations, seed 1.

## Changes and how each was confirmed

### Warlock

| # | Change | Sim | Confirmed by |
|---|---|---|---|
| WL1 | Wrack 25.2 + 1.75 a level (60.2 at 60), coefficient 0.2002 | `sim/warlock/wrack.go` | hotfix 113766 (SpellEffect 1349556) |
| WL2 | Wrack's x1.1 on periodic damage removed. While Wrack channels, Corruption, Siphon Life and Bane of Agony tick at half their period (same total duration, so more ticks). Progress to the next tick is kept and rescaled when the channel starts or ends (also when cut short); a DoT applied or refreshed mid-channel starts at the fast rate; no partial tick at expiry; Wrack's own ticks and Bane of Doom are not sped up. Bane of Agony's ramp steps every 4 ticks and stays at 150% past the 12th. | `sim/warlock/talents.go` `applyWrack`, new `Dot.SetTickPeriod` in `sim/core/dot.go`, `curses.go` | hotfix 113766 (SpellEffect 1349557, aura 4, value 100); the tick-timing model is notes only |
| WL3 | Soul Siphon counts only Corruption, Siphon Life and Bane of Agony (4% a point each, cap 36%) | `talents.go` `soulSiphonMultiplier` | notes only (live, no row) |
| WL4 | Imp Firebolt rank 1: base 6, 0.3 a level, variance 0.333 (6-8) | `sim/warlock/spell_data.go` | hotfix 113766 (SpellEffect 681913). Ranks 2-4 have no row and the post gives no numbers: client values kept, rows 7799/7800/7801 wanted. |

Tests: `TestWarlockWrack70888`, `TestWarlockWrackSpeedsDots`, `TestWarlockWrackStoppedEarly`,
`TestWarlockWrackAgonyMidChannel`, `TestWarlockSoulSiphon70888`, `TestImpFireboltRankOne`.

### Hunter

| # | Change | Sim | Confirmed by |
|---|---|---|---|
| H1 | Arcane Shot costs 11% of base mana at every rank (189.2 at 60, was 190) | `sim/hunter/arcane_shot.go` | hotfix 113892 (SpellPower 3044, 14281-14287) |
| H2 | Summon Hawk costs 11% of base mana at every rank | `sim/hunter/summon_hawk.go` | hotfix 113892 |
| H3 | Hawk melee damage x1.8 on the log-fitted swing | `summon_hawk.go` `hawkMeleeDamageHotfix20261110` | notes only. A log after Nov 10 needs hawk melee hits (before armor, by crit/glance) against the owner's RAP: about 1.8x the old fit (3.3-6.7 + 0.0179 x RAP at level 30 before Unleashed Fury), same 2.5 s swing. |

Tests: `TestForeverHotfix20261110Costs`, `TestSummonHawkHawks`.

### Mage

| # | Change | Sim | Confirmed by |
|---|---|---|---|
| M1 | Arcane Concentration: one 10% chance per Arcane Missiles channel, rolled in even shares on the cast and each missile, behind the 1 s cooldown (about 10% of channels give Clearcasting, was about 41%) | `sim/mage/talents.go` `applyArcaneConcentration` | notes only. Client 70888 SpellAuraOptions 124202 (spell 11213): ProcChance 100, ProcCategoryRecovery 1000 ms, no hotfix row, so the per-roll chance is server side. See ruling 1. |
| M2 | Flamestrike burn ticks don't consume Clearcasting | already so; test added | `TestMageFlamestrikeTicksKeepClearcasting` |

### Druid

| # | Change | Sim | Confirmed by |
|---|---|---|---|
| D1 | Nature's Swiftness also makes Wild Growth instant | `sim/druid/healing.go` | hotfix 113885 (SpellEffect 695239 class mask). No report build casts Nature's Swiftness. |
| D4 | Hotfix 113860's SpellEffect 704563 (spell 27851) is Idol of Ferocity's equip effect: -2 Claw and Rake cost. The hotfix only drops mask bit word1 0x400, which no druid spell carries, so no cost moves. The sim had Classic's -3; now -2. | `sim/druid/items.go` | hotfix 113860 and client 70888 rows. No build wears the idol. |

### Warrior

| # | Change | Sim | Confirmed by |
|---|---|---|---|
| WR1 | Sweeping Strikes keeps its charge on a hit with no second target | `sim/warrior/sweeping_strikes.go` | notes only; `TestForeverWarriorSweepingStrikesKeepsCharges`. No report build has a second target. |

### Items

Wowhead Forever scrape of 2026-11-11 (13,243 items, was 12,199) and client 1.60.1.70888 item facts with their hotfix
overlays (ItemSet, ItemEffect, ItemXItemEffect, SkillLineAbility, SpellItemEnchantment; ItemSparse and SpellEffect are
the build's own tables). `state.json`'s `foreverClientBuild` still says 70755.

| # | Item | Result |
|---|---|---|
| I1 | Rage of the Storm | Wowhead now carries 16 Str, 15 Sta, 102-154 at 3.6 (hotfix 113886); the shaman worker's hand entry is dropped. |
| I2 | Earthen Silk Slippers | Wowhead still old: hand entry in `forever_seen_in_game.json` from hotfix 113944 (+17 healing / +6 damage). |
| I3 | Sparklance Wand | Arcane, via a fallback in `tools/database/gen_db/main.go` (its ItemSparse row is hidden at 70888). Notes only. In the wand gear-search pool (req 58); no build picks it. |
| I4 | Searing Dagger | Proc 1291568 -> 1291570 (7 Fire) from damaging spells. Not modelled: req 13, in no build or search. |
| I5 | Gloves of Holy Might | Hotfix 113833: equip aura 1330855, 15% / 20 s on melee and harmful spells, Smite 101-123 Holy (x3 vs Undead). Not modelled: req 37, in no build. |
| I6 | Enchant Lesser Spirit (bracer, boots) | Now enchant 8203, +4 Spirit (hotfix 113844), in `tools/database/forever_enchants.go`. |
| I7 | Berdun's Trusty Rifle etc. | No db effect (quest reward BoP; ammo and Goliath's Wristguards not in the db). |
| I8 | Engineering trinkets, Flamebreak Talisman | Left to TASKS.md. Wowhead moved Adaptive Combat Assistant, Weakness Analyzer and Serenity Field (cooldowns up), none in builds. |
| I9 | Spell 1324323 | Violet Sorcerer's Vestments 5-piece, Summon Arcane Elemental, now 100% / 15 s (hotfix 113846). Level 29-30 set, not modelled. |

Side effect of the scrape: Blackhand's Breadth now 1% crit (Wowhead Forever; db had Classic's 2%). It is in the
Enhancement, rogue and warrior test gear sets, not in any report build.

### Standards

Standards 25 amended (02dee9e): since Blizzard's Nov 10 hotfix a weapon imbue and a weapon totem both work; the game
changed, not the "both totems" ruling.

## 180 s DPS, before -> after (own sims)

| Build | Before | After |
|---|---|---|
| Warlock Affliction | 1034.4 | 1076.9 (956.2 with the old rotation) |
| Warlock Affliction, duty curse | 1025.1 | 1066.6 (947.9 old rotation) |
| Warlock Affliction, duty Recklessness | 1019.7 | 1057.3 (943.3 old rotation) |
| Hunter Beast Mastery | 1195.39 | 1195.40 |
| Hunter BM wolf | 1150.41 | 1150.43 |
| Hunter BM tallstrider | 1136.91 | 1136.92 |
| Hunter Marksmanship | 1253.96 | 1254.27 |
| Hunter MM wolf | 1228.76 | 1229.06 |
| Hunter MM tallstrider | 1221.11 | 1221.45 |
| Hunter Lone Wolf | 1313.76 | 1314.26 |
| Mage Arcane | 969.28 | 962.12 |
| Mage Fire | 1057.33 | 1052.82 |

Affliction at 120 / 180 / 300 s: 1004.7 / 1034.4 / 1032.2 -> 1075.5 / 1076.9 / 1064.0. `report_hashes.txt` moved
for exactly these 12 builds (0df5a14); the item pull moved none.

## Re-optimized builds

- **Affliction (all three):** rotopt knob search, then talent search (talents unchanged,
  05500021035001351-0050003201-005004). Kept at 120/180/300 s: Bane of Agony only, no Bane of Doom (+109.8); Corruption
  recast 2.5 s early (+6.5); Wrack whenever nothing above it is due, a new `wrack` knob value 4 (+2.9);
  corruptionMinTime 9 -> 3 (+2.0). Shadow Bolt is only the pre-pull cast. Wrack still earns its place (815 without it).
  Undead stays. Req json/meta, `ui/warlock/presets.ts` and the wiki updated. Shadow Bolt Affliction (1015.1),
  Destruction, Demonology and Destruction Fire can't reach Wrack and don't move.
- **Hunter BM/MM:** Arcane Shot knobs not searched (cost -0.8 mana at 60, mana floors already 0). Hawk vs the current
  BM build at 120/180/300 s: current 1197.4/1195.8/1177.3; two-hawk rule after +80% 1181.8/1179.9/1170.4; one hawk kept
  out 1188.7/1186.2/1174.3. Talents, pet (cat) and rotations unchanged.
- **Mage:** no talent or rotation change.

## Tests

Baseline (dbafa63): hunter, mage, priest/healing, priest/shadow, rogue/dps_rogue, warlock/dps and warrior/dps_warrior
failed from older expected results (the shaman round's list). This round regenerated hunter and warlock/dps, which
now pass. tools/rotopt `TestWarriorDpsHoldUnderBuff` fails on master's db too (wants Enrage about 165, gets 167.05).
Full suite after landing: see the end of this file.

## Needs a user ruling

1. **Arcane Concentration (M1).** Modelled: one 10% chance a channel, split over the cast and the missiles (Arcane
   962.12, Fire 1052.82). Other reading: only the cast's extra roll was removed and each missile keeps the full 10%,
   which is the old sim (Arcane 969.28, Fire 1057.33). A variant, 10% split over the missiles only, gives Arcane 962.07
   (same as the model). A log settles it: about 10% of channels leave Clearcasting under the model, about 41% under the
   other reading.
2. **Bane of Agony's ramp under Wrack's faster ticks.** Modelled: steps every 4 ticks, 150% past the 12th (Affliction
   1076.7 at 180 s, 10k iterations). Alternative: the ramp follows 8 s thirds of real time (1065.3, about -11).
3. **Warlock class default.** Affliction (1076.9) now beats Destruction (1062.8) at 180 s. Round 70753's precedent
   makes the highest build the default; not changed.
4. **Sparklance Wand's spell stat.** The in-game sighting (BayovCore, 2026-11-09) has NaturePower 7; Wowhead lists +7
   Arcane spell damage. The db keeps NaturePower 7. No build uses it.
5. **Edgemaster's Handguards.** Wowhead Forever lists 10 expertise rating (stored as 1% expertise), and the db also
   keeps Classic's +7 weapon skills, so it may count both. Only the warrior phase 1 test gear set wears it.
6. **Survival hunter is stale in the report:** published 1041.2 / 1014.5 / 1007.4, but the requests sim 945.2 /
   918.9 / 911.3 on master, from an earlier change. The full-run session's re-sim will publish the new numbers.
