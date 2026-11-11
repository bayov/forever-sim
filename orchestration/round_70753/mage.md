# Mage: round 70753

Status: landed on master (Phase A and B), 2026-11-03.

## Final report

### What changed (source, DPS effect)

| Change | Source | Effect at 180 s |
|---|---|---|
| Clearcasting stays up on an Arcane Missiles cast on Missile Barrage (`sim/mage/talents.go`) | notes 2026-11-02, server side (client 70753 unchanged) | within noise on every preset |
| Ignite takes Flamestrike's burn crits (`sim/mage/ignite.go`) | notes 2026-11-02; client: the burn is direct damage spells 1279983 to 1279990 | none (no preset casts Flamestrike) |
| Arcane: Arcane Blast ahead of the Frostbolt filler while the mana is more than 75 a second of the fight left (new `blastFiller` knob, `tools/rotopt/mage_arcane.go`), and the talent search then took Arcane Meditation 3 and Incineration 5 for Ice Shards | our audit: Arcane ended every fight with about 5000 mana | +16.6 (+36.4 at 120, +6.3 at 300), two seed sets |
| Fire: Scorch refreshes Fire Vulnerability with under 9 s left (was 5) | rotation knob search | +0.6 (+1.4 at 120, +1.5 at 300), two seed sets |
| A2's race base stats and the Nov 2 racials (shared) | notes, client 70753 | no Mage preset moves beyond 0.5 |

Everything else in the notes' Mage section was in the sim already, from the 10-27 to 10-31 hotfixes and client 70707; the ledger below gives each line.

### Follow-up: exploratory search (2026-11-03, after landing)

Talent searches seeded from nine archetypes (Arcane + Fire with a Fireball filler, Arcane + Frost, deep Arcane, Frost + Arcane Blast with the spare-mana knobs, deep Frost, deep Fire, Fire + Arcane Blast stacks, Frostfire + an Arcane splash, and the current Arcane), and two new knobs (`blastClearcast`, Arcane Blast on Clearcasting; `blastFiller` for Frost too). Landed: Arcane with Arcane Blast on Clearcasting and Fireball as the filler, talents `050205203100311531-5551-003`, 986.1 / 962.3 / 944.5 as a Gnome (+4.5 / +8.3 / +7.5 on two seed sets; finisher pad 0.85 sec, Fireball's flight from 20 yd). Gnome stays best (962.4), Undead best Horde (960.2); the High Order override was regenerated (+4.9 / +5.2). Rejected: deep Frost (+2.5 at 120 and 180, -45 at 300, out of mana); every other seed returned to its preset or below it.

### Per build (20,000 iterations, seed 1, the report requests at 120 / 180 / 300 s with ±20 s)

| Build | Race | Talents | Old | New |
|---|---|---|---|---|
| Arcane | Gnome | `050205203100311531-5551-003` (was `050205203100211531-25-0535`) | 944.9 / 937.5 / 930.6 | 986.1 / 962.3 / 944.5 (981.3 / 953.9 / 936.8 before the follow-up) |
| Fire | Human | `0502052-55500201230133031-003` (unchanged) | 1065.4 / 1057.1 / 1040.4 | 1066.8 / 1057.9 / 1042.0 |
| Frost | High Order | `0502050030001-4-253500230111301211` (unchanged) | 944.5 / 944.7 / 937.7 | unchanged |
| Frostfire (default) | Human | `-2550030103013303-0235000311203002` (unchanged) | 1052.4 / 1064.3 / 1041.8 | unchanged |

Old Arcane and Fire at 120 and 300 s are rotopt evaluations of the old knobs at those lengths (20k, seed 1); the 180 s figures are the published ones. Frost's talent search found Incineration 5 for Arcane Meditation (+2.6 at 120 and 180 s, -6.4 at 300 s); it stays out. Frost's knob search found no mana gems: +11 at 300 s, but nothing at 120 and 180 s on the old talents, so it stays out too. Frostfire kept every talent and knob.

### Races at 180 s (each race's own rotation and gear where it has them)

| Build | Best overall | Best Horde | Best Alliance |
|---|---|---|---|
| Arcane | Gnome 953.9 | Undead 952.2 | Gnome 953.9 |
| Fire | Gnome 1058.4 (Human 1057.9, 0.4 behind: stays Human) | Undead 1057.0 | Gnome 1058.4 |
| Frost | High Order 944.7 | Troll 938.4 | High Order 944.7 |
| Frostfire | Human 1064.3 | Troll 1059.4 | Human 1064.3 |

Full tables in `ui/mage/presets.ts` (race check 2026-11-03) and the wiki. No race switch: Fire's Gnome lead is under the 0.5 a switch needs. The Arcane High Order override was regenerated with the new knobs (+5.2 over the preset's rotation at High Order on both seeds); Fire's High Order override already matched. The Dwarf can't be a mage, so there is no hit-adjusted mage gear.

### Audit (passes results, rotations, gear, consumables, mechanics, client_update)

- client_update and mechanics: the Phase A ledger below (every notes line, client 70707 against 70753, the hotfixes). Open: finding client_update-20261030-2 (Missile Barrage's tick effect still a dummy in 70753), which needs an Arcane combat log.
- rotations: Arcane's 5000 unspent mana was the finding, fixed above. The new line states its intent ("mana the fight can't otherwise spend"), and the change holds at 120, 180 and 300 s and at 180 ± 20 s (the request). Arcane's casts, hand-checked: 43.5 Arcane Blasts and 12.2 Frostbolts a fight, stacks up 86% of the time, 2.5 wand casts at the end. Optional, not taken: Arcane `finishPad` 0 (+1.5 over the three lengths) against the Frostbolt's 0.75 s flight from 20 yd (standards 31).
- results: Arcane's request reproduces 953.9 and Fire's 1057.9 (20k, seed 1); Frost and Frostfire reproduce their published 944.7 and 1064.3 exactly.
- gear: no Mage gear change in this round. `tools/audit/gear/lin2.py` fails on the current `orchestration/gear_targets.md` (a "Sta" header where it expects a number); that's a shared tool, so I left it.
- consumables: no consumable or enchant the Mage presets use changed in 70753.

### Open questions and shared changes

- Shared changes: none. Everything is in sim/mage, tools/rotopt/mage_*, ui/mage, tools/dps_report/req/mage__* and races/mage__*, the wiki and notes.md (the Ice Lance item).
- `mage_ice_lance_cd`: client 70753 gives every Ice Lance rank the cooldown and category 19, so the assumption's alternative no longer has a source. Recommendation: one value.
- The orchestrator's re-sim will move `mage__arcane` and `mage__fire` (report_hashes.txt rewritten).

## Phase A: what the sim had, what changed

Sources checked:
- Blizzard's development notes post 9 (2026-11-02), "Mage", "Races", "Bug Fixes" and the general class lines.
- Client 1.60.1.70753 against 70707 on wago.tools (`wago.py diff`, set client_update). The mage tree (TraitNode, TraitNodeEntry, TraitEdge) is unchanged from 70707: every TraitNode move and MaxRanks change in the diff is a druid, paladin or hunter talent. Mage rows that changed: Arcane Blast 1045737 15 -> 12 and 1045738 175 -> 100, Ice Lance's frozen bonus (1203129, 1238140 to 1238152) 200 -> 300, Winter's Chill 688253 5 -> 20, Arctic Reach 694455 gains Ice Lance's mask, Ice Lance ranks 2 to 6 gain the 8 sec category cooldown and category 19 (SpellCooldowns 83266, 94686 to 94689; SpellCategories 148169, 173563 to 173566), Missile Barrage 400589 a new visual only (its effect 1 is still the dummy of finding client_update-20261030-2).
- Server hotfixes: wago.tools has no hotfix rows for 70753 (`hotfixes.py fetch 1.60.1.70753`, 0 rows). foreverchanges.pro's November 3 hotfix touches only the hunter.

Tiers: H server hotfix, C client data (70753), W Wowhead, N Blizzard's notes, F foreverchanges.pro.

| Note line | Tier and evidence | Sim status | Fix |
|---|---|---|---|
| Imbue Iceknife's visual changed | N | n/a | none (a visual) |
| Frost Blast removed | N, C (out of the spellbook since 70707) | landed | the sim never cast it |
| Missile Barrage consumed at the start of the Arcane Missiles channel (was the end) | N; server side | landed | `arcane_missiles.go` always used it up in ApplyEffects, as the channel starts |
| Missile Barrage moved to row 2 | N, C (70707 tree, unchanged in 70753) | landed | none; A1 owns the tree |
| Arcane Resilience to row 1, 3 points, 75% of Intellect as armor | N, C | landed (n/a for DPS) | none |
| Clearcasting not consumed on Arcane Missiles while Missile Barrage is up | N; server side | **missing** | fixed: `talents.go` Clearcasting keeps itself when the Arcane Missiles cast used up Missile Barrage (`mage.arcaneMissilesOnBarrage`, Forever only). Test `TestMageClearcastingStaysOnMissileBarrage`. Preset DPS: within noise (Arcane 937.52 -> 937.37 +/- 0.28, 180 s, 20k, seed 1) |
| Arcane Missiles did not proc Clearcasting (fixed) | N; server side | landed | the sim's missiles always rolled Arcane Concentration (1 sec internal cooldown); nothing to do |
| Arcane Focus removed, folded into Precision | N, C (70707) | landed | none |
| Arcane Blast: +100% mana a stack, 12% of Intellect a stack, 18 sec | N, C (70753 1045737/38), H (10-30, in the sim since PR #118) | landed | none |
| Ignite not applying to Flamestrike past the first hit (fixed) | N; C: Flamestrike's burn is the area trigger's direct damage spells 1279983 to 1279990 (effect 2, no can't-crit attribute) | **missing** | fixed: `ignite.go` also ignites Flamestrike tick crits (Forever only; other periodic crits still don't). Test `TestMageIgniteOnFlamestrikeBurn`. No preset casts Flamestrike |
| Incineration 5 points, 10% | N, C (70707) | landed | none |
| Improved Fireball renamed Accelerant, also Arcane Wrath | N, C (11069 effect 0 mask) | landed | `arcane_wrath.go:64` |
| Critical Mass back to 3 points, 6% | N, C (70707) | landed | none |
| Improved Fire Blast to row 3 | N, C (70707) | landed | none |
| Piercing Ice removed | N, C/H (since the 10-20 hotfix) | landed | none |
| Ice Barrier +15% Frost damage for 1 min, +5% more while it holds | N, H (10-30), sim since 2026-11-01 | landed | `ice_barrier.go` (`FrigidPowerDamage`, `IceBarrierShieldedDamage`); the shield is a real absorb and the encounter's raid hits (2 a minute at 1500) break it, so the generic boss is modelled, not assumed. Re-checked in Phase B's Frost search |
| Frost Warding raises Ice Barrier and Frost Ward absorbs | N, H | landed | `IceBarrierAbsorbAt` (10% a point) |
| Frost Warding negates instead of reflecting | N | n/a | defensive |
| Precision 3 points, 2/4/6% hit with all spells | N, C (70707) | landed | none |
| Ice Lance 8 sec cooldown shared with Fire Blast; 300% on Frozen | N, C (70753: every rank now has the cooldown and category 19), H (10-30) | landed | the sim had it by ruling (2026-10-30, 2026-11-02); the client now agrees, so `notes.md`'s open item is settled. Question below on `mage_ice_lance_cd` |
| Winter's Chill 20% a stack to Cone of Cold | N, C (688253), H | landed | none |
| Talents lose the low-rank proc penalty, except Blizzard | N; server side | n/a | the sim never modelled the penalty (update_ledger SHR-70170-02); no preset casts a rank more than 12 levels below 60 (Frostfire's rank 2 Frostfire Bolt is learned at 50) |
| Races: Human Ruthless 1% (was 2%), The Human Spirit 10% | N, C (697219 2 -> 1, 698126 5 -> 10) | A2 | Fire and Frostfire are Human: Phase B race sweep |
| Races: Dwarf Earth Fury -> Artisan's Precision 1% hit | N, C (1268914) | A2 | Phase B: standards 29's hit-adjusted gear for the Dwarf |
| Races: Orc Savage removed, Blood Fury 2 min | N, C | A2 | Phase B race sweep |
| Races: Skyborne Wind-Blessed 2% (was 3%), Read Leyline 50% for 15 sec | N, C (1268898 3 -> 2) | A2 | Frost is High Order and every mage rotation has the Read Leyline knob: Phase B race sweep and race-tuned rotations |
| Races: Troll Berserking 15 sec, Undead Touch of the Grave 15 sec internal cooldown, Tauren Endurance health only | N, C (243729/33) | A2 | Phase B race sweep |
| All races' base stats corrected (High Order too) | N | A2 | Phase B |
| Proc roots (Frostbite) have diminishing returns | N | n/a | the boss is immune to roots |
| Moonkin Aura and Leader of the Pack crit is all crit again | N, C (701335, 702727) | n/a here | a raid buff; the mage requests take Moonkin Aura's spell crit as before |
| Threat and healing rows | | n/a | tank/healer sessions |

Other:
- Finding client_update-20261030-2 (Missile Barrage's 0.5 sec tick effect a dummy): 70753 leaves 400589 effect 1 as the dummy, so nothing new. Still needs an Arcane combat log. The notes' "consumed at the start of the channel (was at the end)" reads like a server script that runs the halved channel, which supports the sim's five missiles in half the time.

## Notes-only values

Per the round's source rule (README, a32a20024): a note line that neither client 70753 nor a hotfix carries is modelled from the notes, gated under Forever.
- Every numeric mage line is in the client (70707 or 70753) or the 10-30 hotfix, so no mage value rests on the notes alone.
- Two mechanics are server side, with no client field, and are modelled from the notes (client 70753 unchanged): Clearcasting kept on a Missile Barrage Arcane Missiles (`talents.go`), and Ignite on Flamestrike's burn crits (`ignite.go`). Both are gated on `Env.IsForever()`.
- Arcane Concentration on Arcane Missiles (Blizzard hotfixes of 2026-11-09, "now correctly splits across the cast and every bolt (was granting 10% on cast and per bolt)"; server side, no row: client 70888 SpellAuraOptions 124202 for 11213 is ProcChance 100, ProcCategoryRecovery 1000, unchanged). Modelled as the talent's chance once per channel in all, rolled in even shares on the cast and on each missile (1/6 each at five missiles), behind the 1 s ICD (`talents.go` `applyArcaneConcentration`, Forever only; `TestMageArcaneConcentrationSplitsAcrossMissiles`). A log would confirm it: about 10% of Arcane Missiles channels should leave Clearcasting (it was about 41%).
- Flamestrike's burn doesn't use up Clearcasting (same hotfixes): the sim already ends Clearcasting only on a cast, never on a periodic tick (`TestMageFlamestrikeTicksKeepClearcasting`).

## Phase A questions for the orchestrator

1. `mage_ice_lance_cd` ("Does every Ice Lance rank have the cooldown?"): client 70753 gives every rank the cooldown and category 19, so the alternative no longer has a source. Recommendation: collapse it to one value, the sim's (it is in `tools/dps_report/assumptions_def.py`, which I haven't touched).
