# tank_mechanics pass 30 (2026-11-24, master c7e3b3e)

Why due: 0f567e196 and d757d5d make the Protection warrior's level 60 APLs cast the racial actives
(`tools/rotopt/warrior_tank.go` tankRacialLines, `ui/tank_warrior/apls/level60_*.apl.json`, the tankbench requests)
and add the report's Touch of the Grave, fears and racial-timing notes (`tools/dps_report/tankheal_text/tank.md`).
No sim code changed. Client: 1.60.1.71037 still newest.

## Racial actives on a warrior tank (client 71037 rows: `tools/audit/tank_mechanics/spellrows.py`)

| Racial | Client | Sim | Verdict | Tier |
|---|---|---|---|---|
| Blood Fury 20572 | auras 166/167/317 +10% AP, RAP, SP; 15 s; 120 s; no GCD | AP, RAP, SP, healing x1.1 for 15 s, 2 min, no GCD; weapon part off (report assumption) | OK: threat comes from the AP in Revenge, Shield Slam, HS and swings | 1 |
| Berserking 20554 (sim 26297 tag 2) | auras 319/140/65 +10%, 15 s, 180 s, no cost | flat x1.1 attack speed for 15 s, 3 min, free | OK: 10% more swings, so 10% more white Rage while up | 1 |
| Elune's Light 1259799 | aura 290 +10% crit, 15 s, 180 s | +10% melee and spell crit, 15 s, 3 min | OK | 1 |
| Eureka! 1259813 | aura 108 misc 14 -10% cost, misc 0 +10% damage, one mask; 8 s; 120 s | generated scope: Shield Slam, Revenge, HS, Cleave, Thunder Clap, Shield Bash in; Sunder out | OK: Sunder Armor isn't in the client mask; flat threat (Shield Slam's, Sunder's) untouched | 1 |
| Touch of the Grave 1260189 / drain 1260198 | 20% proc, 15 s ICD; drain "equal to 5% of your maximum Health", effect 9, no coefficient | 5% max health, magic hit roll, heals what it deals, threat x stance | OK (43.6 TPS from 20.2 DPS on Balanced); a code comment still says "up to"/1 s ICD (in -2) | 1 (tooltip) |

Casts a fight (Balanced request, 180 +- 20 s, 1000 fights): Eureka! 2.0, Blood Fury 2.0, Elune's Light 1.5,
Berserking 1.5. The APL ids (20572, 26297 tag 2, 1259799, 1259812) are the ids the sim registers.

## Passives on a tank

| Race | Client | Sim | Verdict |
|---|---|---|---|
| Human | Ruthless +1% crit, Spirit +10% | both | OK (Spirit does nothing in combat) |
| Dwarf | +1% hit, Beasts +5%; Stoneform 20594 -10% physical taken 8 s, 3 min | hit, Beasts; no Stoneform | Stoneform missing: 20261124-2 |
| Night Elf | Quickness 20582 +1% dodge | +1% dodge (healing needed 384 -> 372) | OK |
| Gnome | Expansive Mind 1259802 +5% max Rage | x1.05 max Rage | OK |
| Orc | Command (pets), Hardiness (stuns) | nothing for a warrior | OK |
| Undead | Touch of the Grave | as above | OK |
| Tauren | Endurance +5% health | x1.05 health (8889 -> 9333) | OK |
| Troll | Regeneration 20555 +20% regen, 20% in combat | not modelled | missing: 20261124-1 (high) |

Shatter Curse (Orc, 5% magic taken 8 s) and Rapid Regeneration (Troll, cancelled by damage) are left out; neither
matters to a boss tank.

## Measured (scratch patch `$AUDIT_WORK/tank_mechanics/troll_regen_stoneform.patch`, Balanced request, 6000 fights, -labeled)

| Build | Healing needed | Max threat | Balanced | Max survival |
|---|---|---|---|---|
| Troll | 384 | 3594 | 3018 | 1291 |
| Troll + in-combat Regeneration | 373 | 3602 | 3043 | 1366 |
| Dwarf | 385 | 3560 | 2983 | 1252 |
| Dwarf + Stoneform on cooldown | 384 | 3561 | 2986 | 1259 |

## Fears

Panic (19408) in tools/encounters/logs: 26 immune (Anniversary MC), 22 immune + 1 resist (SoD MC); Bellowing Roar
(18431): 1 and 11 immune. No landed fear, so nothing calibrates one; the notes' exclusion and reasons hold.

## Tests

`go test -tags=with_db ./sim/warrior/... ./sim/core/ ./tools/rotopt/`: all pass. No test covers the racial actives on
a tank or Eureka!'s warrior scope (20261124-3).

## Findings

- 20261124-1 (high): Troll Regeneration's in-combat health regen missing.
- 20261124-2 (low): Forever Stoneform not simmed; stale Touch of the Grave comment.
- 20261124-3 (low): no test for the tank racial actives or Eureka!'s warrior scope.
- 20261117-1, 20261108-1, 20261107-1: unchanged, still open.
