# tank_mechanics pass 16

- master: 70ea588bb1 (from d630df2cb0)
- client: 1.60.1.70938 (unchanged; wago.tools latest)
- result: quiet

## What changed

| Commit | File | Change | Tank impact |
|---|---|---|---|
| 0cc2c3c8b | sim/druid/ferocious_bite.go | Forever: an avoided Bite refunds 80% of its 35 Energy (ATTR_EX_1_DISCOUNT_POWER_ON_MISS; wowsims/forever 0384d8f98c on a 70862 log) | none: cat only, no bear preset or tank rotation casts Ferocious Bite |
| 0cc2c3c8b | sim/druid/forms.go | Sharpened Claws adds 3% spell crit a point in the feral forms as well as melee crit | none measurable: no bear ability rolls spell crit |

## Checks

| Item | Client (70938) | Sim | Verdict | Tier |
|---|---|---|---|---|
| Sharpened Claws 16942 aura | effect 6, aura 290 A_MOD_CRIT_PCT | MeleeCrit + SpellCrit, 3% a point | match | 1 |
| Sharpened Claws forms | SpellShapeshift mask 145 (Cat, Bear, Dire Bear) | GetFormShiftStats, used by cat (forms.go:144) and bear (bear.go:98) | match | 1 |
| Bear outcomes | - | Maul/Swipe/Lacerate/Primal Bite melee special hit+crit; Faerie Fire (Feral), Demo Roar magic hit, no crit | spell crit unused | - |
| Ferocious Bite refund | rows carry DISCOUNT_POWER_ON_MISS | Refund 0.8 on Forever, IssueRefund when not landed | cat only | 1 + wowsims log |

Tests: bear_rules_test.go:768 pins Sharpened Claws' melee crit in bear form; spell crit is not pinned
there, but no tank output reads it (not a finding). Update ledger has no entries since pass 15.

## Not done

- Blizzard posts since dev_notes_seen not re-read (no new client build).

## Open findings (unchanged)

- tank_mechanics-20261108-1 (low): no test pins Warden shock threat under the Stormstrike mark.
- tank_mechanics-20261107-1 (low): client 70862 threat labels in comments and docs.
