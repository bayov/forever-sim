# healer_mechanics pass 18 (2026-11-18, master 2f7e255af8)

Range 5f20dfdf4c..2f7e255af8, 5 watched files. Client 1.60.1.70959 (unchanged since pass 17). Blizzard: hotfix
thread t/2403040 has no post or edit after post 11 / post 1's 2026-11-17T02:18Z edit.

## What changed

| Source | Change | Client / source | Sim | Verdict | Tier |
|---|---|---|---|---|---|
| sim/druid/items_restoration.go, healing.go (192cb2e06) | Idol of Rejuvenation 22398: Regrowth's direct heal +5% on a target with our Rejuvenation | spell 27853: aura 4 (dummy), bp 5; tooltip carries the rule | DamageMultiplier x1.05 around the direct heal only when one of our Rejuvenation ranks is active (hasHot); restored before the HoT applies | match | 1 (value), rule from tooltip |
| sim/shaman/items.go (6a3cadbfc) | Totem of Tides 22345: Riptide mana -5% | spell 27797: aura 108 misc 14 (cost %) -5, mask word 2 0x10 = Riptide 1239242/1239243 | Cost.Factor x0.95 on every SpellCode_ShamanRiptide spell | match | 1 |
| sim/priest/vampiric_embrace.go (4e81669a9) | Vampiric Embrace 10% party, 20% caster, own Shadow damage only | Nov 16 hotfix (t/2403040/11) | vampiricEmbraceHealing(damage, isCaster) | match: finding 16 resolved | 1 |
| sim/paladin/holy_strike.go (4e81669a9) | Shieldsworn Holy Strike no longer rolls Windfury Totem | Nov 16 hotfix | SpellFlagNoWindfuryTotem when Shieldsworn | no healer reach (Prot talent, melee) | 1 |

Note (not a finding): Regrowth's ExpectedHealing (HealComm / heal-in-flight) does not add the Idol's 5%. No healer
preset wears Idol of Rejuvenation, and whether HealComm counts a dummy-aura bonus in game is unknown; recheck if a
preset takes the Idol.

## Tests

`go test -tags=with_db ./sim/ -run 'TestDruidIdolOfRejuvenation|TestRaidForeverVampiricEmbrace'` and
`./sim/shaman/enhancement/ -run TestOrcShamanTotemOfTides`: all ok. Each new mechanic has a test that fails if it breaks.

## Findings

- New: none.
- Resolved: healer_mechanics-20261117-16 (4e81669a9).
- Still open: 10 (test coverage), 11 (Riptide mana 147/180/231 vs client 150/180/230).
