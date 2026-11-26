# Healer mechanics: pass 31

2026-11-25, master 85691655ab (last c3bbe33eab), client 1.60.1.71091 (unchanged; external check done 2026-11-25T13:10Z, under the 72 h interval). Result: quiet.

## Why due

master: 22 watched files changed (2f42f4a44 tank batch, efeb06b95 caster rules, 295206ddf item procs, b663145ad batch procs, merges).

## What changed, read for healer reach

| Commit | File | Change | Healer reach | Verdict |
|---|---|---|---|---|
| 2f42f4a44 | sim/core/spell_result.go | Damage shields (NoPartialResist + PassiveSpell, Forever) take no flat school bonus (EV-0066); new `areaCastHit` flag | damage path only (`applyTargetModifiers` else branch); heals, absorbs, health booking untouched | no healing effect |
| 295206ddf / b663145ad | sim/shaman/item_sets_pve.go | Five Thunders 4pc now `item_sets.ForeverSpellPowerProc` at FuriousStormPPM: white swings PPM, casts (heals included, `foreverCastRolls`) at ppm x max(cast,1.5 s)/60 instead of flat 5% | no healer preset or healbench wears Five Thunders (as pass 17) | no preset effect |
| efeb06b95 | sim/shaman/talents.go, shaman.go | Maelstrom Ready! (408501) aura, Healing Wave / LHW / LB / CL take it off; Elemental Focus keep rule; Forever Flurry consumption | Maelstrom Weapon not on any Restoration shaman preset; Elemental Focus reads damaging spells only | none |
| efeb06b95 | sim/shaman/chain_lightning.go, lightning_overload.go, lightning_bolt.go, earth_shock.go, fire_totems.go, frostbrand_weapon.go | damage, overload, resist rules | none | none |
| 295206ddf | sim/paladin/sow.go | Seal of Wisdom 15 PPM on main-hand speed under Forever (was 50%) | melee only; Holy paladin presets don't melee or use it | none |
| 2f42f4a44 | sim/paladin/talents.go, soc.go, unsealed_wrath.go | Redoubt rolls on avoided swings; Vengeance skips NoProcs hits | tank / Ret only | none |
| efeb06b95 | sim/priest/devouring_plague.go, mind_flay.go, starshards.go, shadow_word_death.go | FullResistanceTicks / binary flags | damage; shadow test hps (Vampiric Embrace) shifts follow damage | none |
| 2f42f4a44 | sim/druid/bear.go, bear_attacks.go, faerie_fire.go, shifting_power.go | bear threat, Shifting Power energy threat | no Restoration druid heal path | none |

## Get current

No new client build (71091) and no Blizzard post newer than 2026-11-21 recorded at today's 13:10Z check; nothing reopened.

## Open findings (unchanged)

- healer_mechanics-20261124-17 (Mana Well MP5 under Innervate/Energized)
- healer_mechanics-20261104-11 (Riptide mana)
- healer_mechanics-20261104-10 (untested heal paths)
