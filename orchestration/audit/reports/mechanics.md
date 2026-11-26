# Mechanics pass 107 (2026-11-26)

- Master: bfbee78fc (last audited 85691655a). Client 1.60.1.71091 -> 1.60.1.71139.
- Result: findings (1 medium, 2 low).

## Repo since 85691655a

No commit touches `sim/**/*.go` or `tools/rotopt/mksettings/**`; no TASKS.md / ledger change, no forever-update merge. Only Phase 0 re-rank preset work.

## Client 71091 -> 71139 (81 spell rows, `wago.py diff --by-spell`)

| Spells | Change | Sim |
|---|---|---|
| World buffs 22888, 16609, 460939, 15366, 22817, 22818, 22820, 23735-38, 23766-69, 24425, 25101 | Attributes_6 0x800 (not in raid instances) | Matches ForeverRaidInstance; comment stale (mechanics-20261126-3) |
| Fengus' Ferocity 22817 | e2/e3 bp 230 -> 115 (spell damage, Nov 23 hotfix baked) | No spell part at all (mechanics-20261126-2) |
| Sayge's fortunes | effect 0 aura -> 4 (dummy), bp kept (23768: 10) | Server side; added to mechanics-20261124-1 |
| Rogue poisons, Reagent Economy 1262636-47 | Attributes_5 0x2, Attributes_0 / duration | Reagents, no DPS |
| Fist of Shahram 16601 | ImplicitTarget 1 -> 27 (wielder only) | Sim buffs wielder only |
| Holy Fire ranks | default SpellAuraOptions rows removed | No effect |
| Reinforce 1225497, Summon Charger / Dreadsteed, Enveloping Winds | legacy / mounts / NPC | n/a |

## Blizzard

- Dev notes t/2360696: still 8 posts, newest post 10 (2026-11-04). Nothing new.
- Hotfix post t/2403040/16 (2026-11-25 23:25Z), not yet in TASKS.md:
  - "Fixed a bug causing Libram of Law to not affect Judgement of Command." Sim misses it: mechanics-20261126-1 (JoC is 13.4% of Shock Ret damage, 1000 iterations; about +0.5% DPS).
  - "The Blackhand's Breadth's trinket can no longer be resisted." Evidence added to mechanics-20261125-2.
  - "The Blackblade of Shahram will now correctly perform its haste buff, but only to the wielder." Sim already does.
  - Balance druid items repeat the Nov 25 blue post (mechanics-20261125-1).
  - Class items for the class auditors: pet melee crit 0% fixed (hunter, warlock), Succubus/Incubus -5% crit, hawks +5% crit, Exorcism can crit (sim already), Judgement refresh below 20 s, Hurricane final hit, Genesis on Hurricane, Chain Lightning Clearcasting with 5 Maelstrom stacks.
- Blue tracker: newest 2026-11-25 23:25Z (the hotfix post above); AV adjustments (PvP, not simmed).
