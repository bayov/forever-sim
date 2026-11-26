# Healer mechanics: pass 32

2026-11-26, master bfbee78fcc (last 85691655ab), client 1.60.1.71139 (was 71091). Result: findings (2 low).

## Why due

Client build 71091 -> 71139. No `sim/` file changed between 85691655ab and bfbee78fcc (gear sets, rotopt, report raws and
audit files only); `orchestration/update_ledger.md` unchanged.

## Client 71091 -> 71139 (`python3 tools/sources/wago.py diff 1.60.1.71091 1.60.1.71139 --set mechanics`)

| Change | Rows | Healer reach | Verdict |
|---|---|---|---|
| Sayge's Dark Fortunes: old 237xx effect 0 aura -> 4 (dummy); new 13047xx aura 80 -> 137 (total stat %), misc1 bitmask | 692746, 699970, 1331584-98 ... | world buff; off in raid instances; no healer preset sets SaygesFortune | none (DPS side: mechanics) |
| Fengus' Ferocity 230 -> 115, Weight of Duty 7 -> 8 Str, Reinforce, Reagent Economy, Explosive Runeshard, Enveloping Winds | | not healing | none |
| Rogue poisons / Vanish Attributes_5, world buffs Attributes_6 | SpellMisc | not healing | none |
| Holy Fire 18167 DurationIndex 21 -> 0 | SpellMisc 329246 | not a player rank (player ranks 14914, 15261-15267) | none |
| Clam foods, Nutritious Food ItemEffect, Hurricane / Prowl text, CurvePoint 295686-90 sign | | not healing | none |
| SpellPower, SpellCastTimes, SpellCooldowns, SpellLevels, Trait tables | | no change | heal ranks unchanged |

## Blizzard since 2026-11-21T00:18Z (tier 1)

Read: Forever hotfixes t/2403040 posts 15 (Nov 23) and 16 (Nov 25); blue tracker (Balance druid t/2429007 is in
mechanics-20261125-1; AV and store posts not mechanics).

| Note | Healer reach | Verdict |
|---|---|---|
| Holy: Oath of Light duration 60 min | sim 5 min (holy_talents.go:114); tank_healer APL recasts it | **finding -19** (low) |
| Genesis now raises Hurricane damage and Tranquility healing | sim and docs say "not Tranquility"; no preset APL casts Tranquility | **finding -18** (low) |
| Moonfury replaced by Dreamstate (Innervate on others) | Restoration preset has no Moonfury; self-Innervate unaffected | covered by mechanics-20261125-1 |
| Darkmoon Faire / Fengus double power; world buffs persist across zones | world buffs, off in raids | DPS side |
| Paladin Judgement refresh, Exorcism crit; Shaman Chain Lightning Clearcasting | damage | none |
| Priest Mind Vision procs, Dark Inspiration (Shadow) | not healing | none |

## Open findings re-checked

- -17 Mana Well energize: camp_buffs.go:58 unchanged, still open.
- -11 Riptide cost: sim/shaman/healing.go unchanged, still open.

## Engine and tests

No engine or healer test file changed; pass 31's engine verdicts stand.
