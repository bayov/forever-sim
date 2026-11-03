# Round 70753: Druid

The Druid worker's ledger and report for the November 2 development notes (client 1.60.1.70753).
Branch `round-70753-druid`.

**Status: Phase A done (mechanics). Phase B (talent remap, re-optimization, audit) waits for "sources landed".**

## Sources checked (2026-11-03)

- Client tables 1.60.1.70707 and 1.60.1.70753 from wago.tools (`wago.py fetch ... --set client_update`), diffed with
  `db2diff.py` (druid section, Attributes_14 noise dropped: 33 spells).
- wago.tools hotfixes: no row for build 70753 at all; for build 70707 the 2026-10-30 push 113283 (the launch realms'
  hotfix, which 70753 now ships in its own tables) and nothing newer on any druid damage row (searched the effect rows
  of Shred 688179, Ravage 687812, Starfire 702704 and Shifting Power 1322605, and every 70707 SpellEffect hotfix row).
- Wowhead Forever tooltips (dataEnv 16): Shred 155%, Claw 110%, Ravage 350%, Starfire 364-426 at 60, Ferocious Bite
  rank 6 199-259 at 1 combo point. The old values.
- foreverchanges.pro patch notes (fetched 2026-11-03): its 70753 entry lists Poised as new; its development-notes
  section repeats Blizzard's lines (Shred 200%, etc.) without "found in the game".

**The tier question.** The notes' talent reworks are all in the 70753 client (and were in the 10-30 hotfix), but none of
the notes' damage lines is, for any class (hunter Aimed Shot 20904 bp 166 and Starfire 25298 bp 381 are 70707's). The
client isn't contradicting the notes with another value, as the Trueshot Aura hotfix did; it just doesn't carry the
server's tuning yet. The sim takes the notes for these lines (`sim/druid/tuning_70753.go`), "about 25%" and "about 40%"
read as 1.25 and 1.40. Sent to the orchestrator as a cross-class question with that recommendation.

## Notes-only values

Per the orchestrator's provisional rule (round README, a32a20024): note lines that neither client 70753 nor a hotfix
carries are modelled at the notes' value, gated under Forever, citing the notes and "client 70753 unchanged". The
2026-11-03 server hotfix has no druid lines.

| Line | Client 70753 (unchanged) | Sim under Forever | Where |
|---|---|---|---|
| Shred weapon damage | 155% | 200% | `ShredWeaponMultiplierNov2` |
| Claw weapon damage | 110% | 140% | `ClawWeaponMultiplierNov2` |
| Ravage weapon damage | 350% | 435% | `RavageWeaponMultiplierNov2` |
| Rip base and AP part | 15 + 25.5/pt, 1% AP/pt a tick | x1.25 | `nov2BleedScale`, rip.go |
| Rake base and AP part | 61 + 34/tick, 0.055 AP | x1.25 | `nov2BleedScale`, rake.go |
| Ferocious Bite base and AP part | 22829 rows | x1.25 (Energy part unchanged) | `nov2BleedScale`, ferocious_bite.go |
| Starfire base damage | 25298 / 9876 rows | x1.40 | `nov2StarfireScale`, starfire.go |
| Rend and Tear | 2% a point | 3% a point | `RendAndTearMultiplier` |
| Pounce | | n/a (not simulated) | |

## Ledger

| Note line | Tier and evidence | Sim status (before) | Fix |
|---|---|---|---|
| Lacerate is now a Feral talent, not learned by all | client 70753: TraitDefinition 134410 -> 1328617, new ranks 1328616/1328617 | n/a for the DPS specs (bear) | tank session |
| Ferocious Bite new rank 1 at 24, ranks relabeled | client 70753 (1328618, "Rank N" subtexts), hotfix 113283 | landed (10-30 hotfix: `ferocious_bite.go`) | none |
| Bear / Dire Bear Form threat 50% -> 115% | notes | n/a | tank session |
| Poised (new passive, level 32) | client 70753: 1328829 passive, buff 1328827 aura 107 misc 14 bp -10, 1 charge, DurationIndex 21, class mask Claw/Shred/Rake/Rip/FB/Ravage/Pounce/Cower; not in the notes' text beyond the line | missing (finding client_update-20261103-1) | **fixed**: `sim/druid/poised.go`, one charge at the pull at full Energy, -10 on the next of those abilities; test `TestCatPoised`. DPS +0.0 on both feral builds: the opener is a Clearcast Ravage, which is free already and uses up the charge |
| Shred 155% -> 200% | notes only (client 70753 bp 155, Wowhead 155%, no hotfix) | wrong (155%) | **fixed**: `ShredWeaponMultiplier` 2.0; `TestCatShredDamage` |
| Claw 110% -> 140% | notes only (client bp 110) | wrong | **fixed**: 1.4; `TestCatClawDamage` |
| Ravage 350% -> 435% | notes only (client bp 350) | wrong | **fixed**: 4.35; `TestCatRavageDamage` |
| Rip base and AP scaling +~25% | notes only (client 9896 bp 15 / 25.5 a point) | wrong | **fixed**: x1.25 on base, per point and the AP part; `TestCatRipTicksLive`, `TestCatRendAndTearBleeds` |
| Rake base and AP scaling +~25% | notes only (client 9904 bp 61 / 34) | wrong | **fixed**: x1.25 on the hit, the tick and the tick's 0.055 AP; `TestCatRakeTickAttackPower` |
| Ferocious Bite base and AP scaling +~25% | notes only (client 22829 unchanged) | wrong | **fixed**: x1.25 on base, range, per point and the 3% AP a point; the per-Energy damage unchanged; `TestCatFerociousBiteNov2` |
| Pounce base and AP scaling +~25% | notes only | n/a (the sim has no Pounce; Ravage is the stealth opener) | none |
| Starfire base damage +~40% | notes only (client 25298 / 9876 unchanged, Wowhead 364-426) | wrong | **fixed**: x1.40 on the base damage at the level; `TestStarfireNov2Damage` |
| Genesis 2/4/6/8/10% | hotfix 113283 (curves doubled) | landed (`applyGenesis`) | none |
| Moonkin Form aura: all crit, exclusive with LotP | client 70753 SpellEffect 701335 aura 57 -> 290; hotfix 113283 | landed (`sim/core/buffs.go`, user ruling 2026-11-01) | none |
| Heart of the Wild to Restoration, swapping with Furor | client 70753 TraitDefinition 134400/134419; hotfix | landed (values unchanged; tree positions are A1's) | A1 tree |
| Lacerate talent row 4 replacing Primal Bite; 2 new ranks | client 70753 | n/a (bear) | tank session; A1 tree |
| Predatory Discipline (row 4 col 3) | client 70753 1328613; hotfix 113283 | landed (`ferocious_bite.go`, report assumption `druid_predatory_discipline`) | none |
| Predatory Strikes to column 1 | client/A1 | tree position only | A1 tree |
| Leader of the Pack: all crit, exclusive with Moonkin Aura | client 70753 702727 aura 52 -> 290; hotfix | landed (`sim/core/buffs.go`) | none |
| Natural Instinct to column 4 | A1 | tree position only | A1 tree |
| Rend and Tear to column 1; 3/6/9/12/15% | position: client/hotfix. Value: notes only (client and hotfix trees still say 2% a point) | value wrong (2%) | **fixed**: 3% a point (`RendAndTearMultiplier`); `TestCatRendAndTear`, `TestCatRendAndTearBleeds` |
| Berserk removed | hotfix 113283 | landed (no Berserk in the sim) | none |
| Primal Bite row 7 col 2, 3 ranks, LotP prerequisite | client 70753: 407995 / 1238069 / 1238070 at 40 / 50 / 60, bp 42 / 62 / 77 | n/a for the DPS specs (bear ability) | tank session |
| Shifting Power row 7, needs Predatory Discipline; 8 s | client 70753 SpellCooldowns 16000 -> 8000; hotfix | landed (`shifting_power.go` 8 s) | none |
| Furor to Feral | client/hotfix | landed (values unchanged) | A1 tree |
| Naturalist 2/4/6/8/10% damage; Wild Growth cast time | hotfix 113283; client 70753 class mask | landed (`talents.go`); Wild Growth is a heal | none |
| Reflection / Subtlety swap rows | client 70753 TraitDefinition 134378/134381 | tree position only | A1 tree |
| Gift of the Earthmother removed; Bloom; Swiftmend own HoT; Wild Growth 3 s | client 70753 | n/a | healing session |
| Races: base stats corrected; Tauren Endurance loses hit / crit damage; Human, Orc, Dwarf, Troll, Undead, Skyborne changes | notes | A2's (shared sim/core) | A2; race sweep in Phase B |
| Talents no longer have a low-rank proc penalty | notes | A2's shared part; the druid sim has no low-rank penalty of its own (checked `sim/druid`) | none |
| Throw windup, roots DR, quitter debuff, stun | notes | n/a | none |
| Threat and healing rows | | n/a | tank / healer sessions |

## DPS effect of Phase A (180 s, 10k iterations, seed 1, the current presets, before A1 and A2)

| Build | master | Poised only | Phase A |
|---|---|---|---|
| feral-cat | 942.55 | 942.55 | 1128.99 (+19.8%) |
| feral-cat-duty-ff | 939.18 | 939.18 | 1125.56 (+19.8%) |
| balance | 891.94 | 891.94 | 951.04 (+6.6%) |
| balance-duty-ff | 858.74 | 858.74 | 914.86 (+6.5%) |
| balance-wrath | 853.11 | 853.11 | 861.60 (+1.0%) |
| balance-wrath-duty-ff | 821.01 | 821.01 | 829.21 (+1.0%) |

Goldens (`TestP1Feral`, `TestP1Balance`) regenerated for these changes; Phase B regenerates them again after A1/A2.

## Open questions

1. Answered (orchestrator, provisional pending the user): model the notes-only values, gated under Forever. Done.
2. Poised and Clearcasting: the opener's Clearcast Ravage spends Poised for nothing. Phase B's rotation search can try an
   opener without pre-pull Clearcasting fishing, or with Poised on a costed ability.
