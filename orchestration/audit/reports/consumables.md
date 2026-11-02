# Consumables pass 65 (master 796495d6f)

Trigger: 21 watched files moved since b647a21ff (standards.md, assumptions_def.py, mksettings spec_*.go and specs.go, seven presets.ts, the new Demonology warlock gear set).

## Verified
- Diffs of every presets.ts and spec_*.go: only DistanceFromTarget changes (standards 33 caster distances), rotation and race notes, a Seal Fate Daggers rotation preset and comments. No Consumes, RaidBuffs, Debuffs or IndividualBuffs object changed; presets.ts still matches mksettings.
- level60_phase1_synthetic_demonology.gear.json: same enchants as the shared set (1509, 1249057, 1254966, 1220624, 1219580), trinkets synthetic 997xxx.
- assumptions_def.py: no buff, oil or imbue entry changed.
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty. Counts as before (minor-armor count 3, known: Dust Cloud; ref expects 2).
- wago latest 1.60 build: 1.60.1.70707, unchanged.
- Note for the open Sapper finding: BM hunter now starts at 12 yd (still outside the Sapper's 10 yd); casters at 20 to 36 yd. No new Sapper or Land Mine use.

## Open
- consumables-20261101-1 (Sapper on ranged hunters, low) remains open.
