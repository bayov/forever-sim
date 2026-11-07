# Healer mechanics: pass 5 (2026-11-07 UTC, master 12aa2796fc, client 1.60.1.70862)

Quiet pass. Due because the client moved 1.60.1.70790 -> 1.60.1.70862; no watched file changed on master.

## Get current

| Source | Seen | Healer content |
|---|---|---|
| wago.tools latest build | 1.60.1.70862 (new) | Bloom heal target only, below |
| Dev notes thread 2360696 | re-fetched 2026-11-07; still 10 posts, last post 10 (2026-11-04T19:35Z) | none new |
| update_ledger.md | no 70862 entry yet | none |
| master c67d9e014..12aa2796f | item fold-in (12aa2796f) and an audit-loop commit | no heal, talent or engine file |

## Client diff 70790 -> 70862 (`python3 tools/sources/wago.py diff 1.60.1.70790 1.60.1.70862 --set mechanics`)

Changed rows are boss/NPC spells (Hogger set: Ground Slam, Warstomp, Cleave, Meathook, Bloodlust), Defensive Stance,
Bear Form passive, flasks, Adventurer's Tent, Kidney Shot/Gouge levels, tooltip text (Thorns, Ravage) and items.
New rows on existing spells (SpellEffect 9, SpellPower 0, SpellAuraOptions 0, others 0) and removed rows (Study,
Impaled, Bloodlust, Warstomp, Cleave cost, Soulstone aura options) touch no healer spell.

| Spell | Client 70790 | Client 70862 | Sim | Verdict | Tier |
|---|---|---|---|---|---|
| Bloom heal 1328622 (SpellEffect 1367728) | HEAL, ImplicitTarget 1 (caster) | ImplicitTarget 21 (target ally) | heals the expiring HoT's target (`sim/druid/healing.go:313-318,735`) | match, the client now agrees | 1 |
| Riptide mana (SpellPower 304804/310863/310864) | 150/180/230 | unchanged | 147/180/231 | finding 11 stands | 1 |

Steps 3-4: nothing else in the healers' spells, talents or the engine changed, so the verdicts in
`ref/healer_mechanics/verdicts.md` stand (Bloom row and header updated).

## Tests

No new mechanic, so no new test gap. `TestDruidBloom` covers the heal on the target.

## Open findings

- healer_mechanics-20261104-11 (Riptide's mana): still open, re-checked on 70862.
- healer_mechanics-20261104-10 (untested heals, talents, engine paths): still open.

## New findings

None.
