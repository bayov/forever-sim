# Mechanics pass 73 (2026-11-10, master da272241e)

No new finding. Both open findings (mechanics-20261107-1 natural flasks, mechanics-20261104-1 druid scales) are still open: no commit touched them.

## Range
ca5e35c3f..da272241e: one sim commit.

| commit | what | verdict |
|---|---|---|
| 2e320521f | The Chief's Enforcer (9477): Farraki Enforcement 1294339 proc, 294 Physical on the melee table, 1.14 PPM, no self-proc | Client 70888 agrees on 1294339 (effect 2 bp 294, effect 3 stun); no SpellProcsPerMinute rate, so the rate rests on the Nov 10 log (69/1071 hits). Item-level, not in any preset. No finding. |

## Client 1.60.1.70862 -> 1.60.1.70888
`wago.py diff --by-spell`: 20 spells. Shared mechanics: none (no racial, raid buff/debuff, consume, enchant).
- SpellItemEnchantment, TraitDefinitionEffectPoints: unchanged. CurvePoint: only new 0->1 identity curves.
- ItemEffect: engineering trinkets 272437/272438/272439 (Damage Sigmoid, Analyzing Weaknesses, Serenity) on-use cd 90 s -> 300/360/360 s; Flamebreak Talisman 280071 cd 180 -> 300 s, Agility 50 -> 33. None is modeled in sim/ or used by a preset.
- Class rows (Frost Nova radius, Pounce attrs, Natural Instinct mask, Cultivation cd, Feral Charge attrs, Frost Trap) and NPC spells (Bloodbolt, Charge, Fireball, Poison Shock): class/NPC side, covered by TASKS.md.

## Blizzard
Dev notes thread t/2360696: still 10 posts, newest post 10 (2026-11-04). Blue tracker: newest is the Nov 9 hotfixes (t/2403040 post 5, 2026-11-10 04:23), already read. The Nov 10 server hotfix (per the daily update) is hunter Arcane Shot / Summon Hawk / Great Stamina and Nature's Swiftness: class-only.
