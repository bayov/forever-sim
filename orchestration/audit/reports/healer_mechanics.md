# healer_mechanics pass 19 (2026-11-18, master c198d81e3)

Range 2f7e255af8..c198d81e3, 2 watched files (sim/priest/priest.go, sim/priest/talents.go), one commit (ac0217aed).
Client 1.60.1.70959 (wago latest, unchanged since pass 17). Blizzard posts not rechecked this pass (last seen
t/2403040 edit 2026-11-17T02:18Z).

## What changed

| Source | Change | Client / source | Sim | Verdict | Tier |
|---|---|---|---|---|---|
| sim/priest/priest.go, talents.go (ac0217aed) | PartyBuffs.dark_inspiration (given form, Spirit x(1+0.15*0.999)) dropped for a party holding a priest with Dark Inspiration, who gives the real aura (registerDarkInspiration) | 1326168: aura 80 (% stat), implicit target 20 (caster's party); same aura does not stack | Priest.AddPartyBuffs clears the flag; Party.GetPartyBuffs clones the base buffs then runs every member's AddPartyBuffs before applying, so the clear reaches the whole party; pets already excluded (buffs.go:566) | match: no double count | 1 |

Healer reach: Spirit drives healer mana regen, so this removes a Spirit double count for any healer partied with a
Dark Inspiration priest. Edge case (not a finding): a priest who takes the talent but never casts Mind Blast or a
full Mind Flay (a healing priest) would now remove the given form from its party and give nothing back. No healer
priest preset is known to take this Shadow talent; recheck if one does.

## Tests

`go test -tags=with_db ./sim/ -run DarkInspiration`: TestRaidForeverDarkInspirationGivenOnce passes
(sim/raid_dark_inspiration_test.go). The change has a test that fails if it breaks.

## Findings

- New: none.
- Resolved: none.
- Still open: 10 (test coverage), 11 (Riptide mana 147/180/231 vs client 150/180/230).
