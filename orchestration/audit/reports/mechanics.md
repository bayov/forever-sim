# Mechanics pass 66 (2026-11-04, master 0facd9428)

One new finding (low): mechanics-20261104-1.

## Range

6b480b060..0facd9428 (70 watched files). Shared-engine commits (sim/core, no mksettings change):

| Commit | What | Mechanics check |
|---|---|---|
| 49f02184c, 9c20070c9 (mana.go) | Mana-gain threat under Forever no longer multiplied by the threat multiplier; NoThreat sources make none | Marked "to verify" (standards 40). Classic and wowsims/forever leave it unmultiplied; no Forever source. Tank threat only. Added to verdicts' open items. |
| 659cad198 (tank_score.go) | Tank score threat per damage 0.7 -> 1 | Salvation removed (verdicts row, post 9, tier 1). Consistent. |
| 0aed93989 (spell_result.go, unit.go) | Direct heal crit takes the school crit bonus; losing max health clamps current health | Healing only; same rule HoT ticks used. No DPS path. |
| b2c16f671 (raid_damage.go) | A fractional raid-damage share can hit nobody | Raid-damage model, not a game mechanic; the Level 60 presets' 2 hits a minute at full share are unchanged. |
| 87f28803b, 318fc7efb, 0950b2242 (health.go, major_cooldown.go, apl_values_resources.go, racials.go, apl_helpers.go) | Health costs (Bloodrage, Reckless Swings) out of TMI; tank-score replay reads no health (survival cooldowns, health APL values, Berserking at 10%) | Replay-only scoring guards. Berserking outside the replay unchanged (10% to 30% by health). |
| Class commits (shaman 113380, warrior Revenge / Unbridled Wrath, paladin, priest, druid, warden) | Class scope | Ported hotfix values; class auditors and the client update ledger. |

## Getting current

| Source | Result |
|---|---|
| Client build | **1.60.1.70790** new (probe). 23 of 34 mechanics tables differ from 70755. |
| 70790 content | Bake of hotfix pushes since 70707 (shaman, warrior, rogue, paladin, hunter, Riptide) plus the November 2 druid values, new in the client. Shared: Trueshot Aura 20/35/50/75/100 matches the sim; no racial, raid debuff, consumable or enchant change (SpellItemEnchantment: rogue poison charges only). CurvePoint (+756 rows) and Trait tables: talent-tree data (rogue/warrior/paladin hotfix trees). |
| Dev notes thread | 9 posts; post 9 last edited 2026-11-04T02:12:27Z (seen). Nothing new. |
| Blue tracker | One post since 11-03 21:41: Kaivax "World of Warcraft: Forever Launch Time" (2026-11-04T18:13Z), no mechanics. |
| Daily update | No forever-update merges; TASKS.md / update_ledger.md only via branch merges. |
| Hotfix pushes | `hotfixes.py pushes` printed nothing this pass; 70790 supersedes the pushes through 113605 seen last pass. |

## Spot checks of 70790 class deltas against the sim

| Spell | Client 70790 | Sim |
|---|---|---|
| Chain Lightning | 0.714, CastingTimeIndex 19 | 0.714, 2.5 s (hotfix 113341) — matches |
| Slam | 15 s category cooldown, Improved Slam -5000 | 15 s minus 2.5 s a point — matches |
| Earth Shock r7 | 531 | 531 — matches |
| Shred / Claw / Ravage | 200 / 140 / 435 | matches |
| Ferocious Bite r6 | 102, 181 a CP, 3.3 an Energy | 102.5, 183.75, 2.7 — finding |
| Starfire r7 | 533, 3.4 a level | 533.4, 3.36 — finding (tiny) |
| Rogue poisons | charges 250 / 240 (no damage change) | charges not modelled; no DPS |

## Conclusions

No shared mechanic moved against tier 1 data. Every shared verdict holds on 70790. The client update auditor should read 70790 in full for class rows (the notes-only gates of every class can now be checked against the client).
