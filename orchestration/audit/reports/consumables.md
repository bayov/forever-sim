# Consumables, buffs and debuffs: pass 33

Master 1e2f7ec5b (last audited dc37245e5). Client still 1.60.1.70450 (wago probe: nothing newer). Result: quiet, no findings.

## What moved
| File | Change | Verdict |
|---|---|---|
| tools/database/forever_enchants.go, assets/db_inputs/forever_enchants.json | Enchant stats moved out of Go into a client-read JSON (tools/sources/enchants.py), filled by gen_db | Compared all 205 enchants in assets/database/db.json before and after: identical except 2H Lesser Agility (1248510) |
| ui/enhancement_shaman/presets.ts | adds `rotationsNameTopRanks: true` | UI naming flag only; consumes, buffs, debuffs unchanged |

2H Lesser Agility went +15 to +9. The comment says the client enchant is named "Agility +15" but the spell worn applies Classic's +9. No preset or gear set uses it (user 2026-10-05: not in the beta), so no DPS effect.

## Verified
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty. Minor-armor exclusive count 2. MobTypeUnknown default. Salvation and Sunder on. Totems match the 2026-10-16 ruling (Enhancement gets Grace of Air + Flametongue).
- Class and consumes branches show as ahead only because they predate master's merges; their diffs against master are stale copies (gotcha 1), not new audited commits.

No open items.
