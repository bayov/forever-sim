# Shaman talent verification (2026-10-12)

This is a check of every talent with points in each shipped Level 60 preset build, run on the branch head after the Ahn'Qiraj ranks were removed.

**Expected values.** Each expected value is the client's trait curve for that rank (TraitDefinitionEffectPoints in 1.60.1.70338). The list of spells a talent should affect is the one its SpellEffect class mask picks out.

**Evidence.** The evidence comes from simulated fights, not from reading the code:
- **One-fight logs.** These are debug logs of a single seed-1 fight from `rotopt -timeline` with `ROTOPT_FULL_LOG=1`, using the preset's mksettings JSON and APL. "Base" and "after" are the `BaseDamage` and `AfterAttackerMods` fields of a log line.
- **Breakdowns.** `rotopt -breakdown` gives casts, crits and aura uptimes over 10000 iterations at 180 sec.
- **Stat differences.** `rotopt -print-stats` with and without the talent.
- **A/B.** The DPS with the talent minus the DPS with its points removed. Both runs use 10000 iterations, 180 sec and the same seed (standard error about 0.3 to 0.5). The talent string is the only thing that changes.

**Result.** Every talent's effect showed up at its client value on the right spells. There were no bugs to fix.

## Elemental Level 60 (`5504301300103051-052-05305`), 690.4 DPS

| Talent | Points | Expected (client) | Evidence | Status |
|---|---|---|---|---|
| Convection | 5 | −10% cost on Lightning Bolt, Chain Lightning, Lava Burst and the shocks | Lightning Bolt costs 198 (220 × 0.9), Flame Shock 369 (410 × 0.9), Lava Burst 238.5 (265 × 0.9), and the rank 7 bolt 121.5 (135 × 0.9). A/B +3.1 | verified |
| Concussion | 5 | +5% on Lightning Bolt, Chain Lightning and Earth Shock | Lightning Bolt base 849.1, after 891.6 (×1.05). Overload base 893.2, after 468.9 (×0.525 = ½ × 1.05). Rank 7 bolt 797.5 → 837.4. A/B +21.6 | verified |
| Reverberation | 4 | Shock cooldown −0.8 sec | Flame Shock is only cast when its DoT runs out (12 casts in 180 sec), so the cooldown never binds. A/B +0.2. The per-rank value shows on Enhancement (5/5 gives a 5.0 sec cooldown) | verified, no effect in this rotation |
| Call of Flame | 3 | +15% on Flame Shock (direct and DoT), Lava Burst, Searing Totem attacks, Magma Totem and Fire Nova | Flame Shock 365.9 → 420.8, tick 137.4 → 158.0, Searing 56.8 → 65.3 (all ×1.15). Lava Burst 888.0 → 1225.4 (×1.15 × 1.2 with Flame Shock up). A/B +29.5 | verified |
| Elemental Focus | 1 | 10% chance of Clearcasting after a damage spell. Clearcasting makes the next Lightning Bolt, Chain Lightning, Lava Burst, Fire Nova or shock free | 9.11 procs over 90.4 eligible casts (10.1%). The next Lightning Bolt costs 0. Searing attacks neither trigger it nor use it up. A/B +2.9 | verified |
| Elemental Alacrity | 3 | −0.5 sec cast on Lightning Bolt, Chain Lightning and Lava Burst | Lightning Bolt (2.5 sec base), Lava Burst (2.5 sec base) and the rank 7 bolt all log a 2 sec cast. A/B +111.6 | verified |
| Call of Thunder | 1 | +3% crit on Lightning Bolt and Chain Lightning | Lightning Bolt crits 17.5 of 65.7 landed (26.6%), against 23.5% on the sheet. Flame Shock is 23.4%, Lava Burst 23.7%, Searing 23.5%. A/B +10.8 | verified |
| Lightning Overload | 3 | 10% chance of an overload at half damage | 6.6 overloads per 66.4 bolts (9.9%), hitting for ×0.525 as above. A/B +22.0 | verified |
| Elemental Fury | 5 | +100% crit bonus, so a spell crit does 2.0× | Lightning Bolt 1577.0 / 788.5, Flame Shock tick 347.7 / 173.8, Searing 107.8 / 53.9, and the overload 1031.6 / 515.8: all 2.0. A/B +69.1 | verified |
| Lava Burst | 1 | The spell: 2.5 sec cast, 10 sec cooldown, +20% with Flame Shock up | 11.9 casts per 180 sec, and the ×1.2 shows above. A/B +16.0 | verified |
| Thundering Strikes | 5 | +5% crit, melee and spell | SpellCrit and MeleeCrit both +5.0. A/B +27.4 | verified |
| Ancestral Knowledge | 2 | +4% Intellect | Intellect +14, Mana +210. A/B +1.4 | verified |
| Totemic Focus | 5 | −25% cost on every totem | Searing costs 127.5 (170 × 0.75), Mana Spring 75 (100 × 0.75). A/B +0.0, because mana doesn't run short in 180 sec | verified |
| Mindfulness | 3 | 50% of spirit regen while casting | Casting regen goes from 12 to 33 mana/sec. Spirit regen is 0.25 × 50 + 0.125 × 236 = 42.0, and half of that is 21. A/B +13.9 | verified |
| Tidal Focus | 5 | +5% spell and melee hit, −5% heal cost | SpellHit and MeleeHit +5.0, reaching the 16% cap. A/B +33.8 | verified |

## Elemental Level 60 + Mana Tide (`5504301300103051-04-053150001001`), 682.1 DPS

The 13 talents it shares with Level 60 show the same evidence on this build. Their A/B deltas:

| Talent | A/B |
|---|---|
| Convection | +5.7 |
| Concussion | +21.3 |
| Reverberation | +0.1 |
| Call of Flame | +29.2 |
| Elemental Focus | +6.2 |
| Elemental Alacrity | +109.0 |
| Call of Thunder | +10.8 |
| Lightning Overload | +22.0 |
| Elemental Fury | +65.5 |
| Lava Burst | +15.0 |
| Totemic Focus | +0.2 |
| Mindfulness | +10.5 |
| Tidal Focus | +33.7 |

All 13 are verified. The talents that differ:

| Talent | Points | Expected (client) | Evidence | Status |
|---|---|---|---|---|
| Thundering Strikes | 4 | +4% crit | SpellCrit and MeleeCrit +4.0. A/B +21.8 | verified |
| Mana Tide Totem | 1 | Rank 3: 290 mana every 3 sec for 12 sec | It goes down at 179.1 sec at 18% mana, for 45 mana (60 × 0.75 with Totemic Focus), and gives 4 × 290. A/B −1.4: a 3 min fight with a potion and a rune doesn't need the mana, and the build is shipped for the group's mana | verified |
| Water Shield | 1 | Mana when the shaman is hit | Nothing hits the shaman in the preset. A/B 0.0 | not verifiable here (inert at 0 hits) |
| Natural Grace | 1 | −5% threat | Utility. A/B 0.0 | not DPS |

## Enhancement Level 60 (`050533102-053030031005112251`), 791.6 DPS

| Talent | Points | Expected (client) | Evidence | Status |
|---|---|---|---|---|
| Concussion | 5 | +5% on Lightning Bolt, Chain Lightning and Earth Shock | Earth Shock base 570.8, after 599.4 (×1.05). With the Stormstrike mark 574.3 → 723.6 (×1.26 = 1.05 × 1.2), and Lightning Bolt 701.7 → 884.2 (×1.26). A/B +6.1 | verified |
| Reverberation | 5 | Shock cooldown −1.0 sec | Flame Shock at 1.50 sec, Earth Shock at 6.50 sec: the shared cooldown is 5.0 sec. A/B +18.7 | verified |
| Call of Flame | 3 | +15% on Flame Shock, Fire Nova and the fire totems (not the Flametongue Totem proc, per its class mask) | Flame Shock 307.2 → 353.3, tick 110.0 → 126.5 (×1.15). Flametongue Totem proc 46.3 → 46.3. A/B +22.8 | verified |
| Elemental Devastation | 3 | +9% melee crit for 10 sec after a direct harmful spell crit | The log shows `Gained {"MeleeCrit": 9.000} from {SpellID: 30165}`. It came from Earth Shock (3), Lightning Bolt (1) and direct Flame Shock (3) crits, never from the 8 Flame Shock tick crits, Fire Nova or Flametongue procs. Uptime 40.7%. A/B +17.8 | verified |
| Elemental Focus | 1 | 10% Clearcasting, as on Elemental | 6.81 procs per 68.0 eligible casts (Earth Shock, Flame Shock, Fire Nova, Lightning Bolt), 10.0%. A/B +11.9 | verified |
| Improved Fire Nova | 2 | Fire Nova +20%, cooldown −4 sec | Fire Nova base 553.1, after 763.3 (×1.38 = 1.15 × 1.2: the two talents multiply, per the beta, shaman_audit 5.6). Casts at 3.0, 9.5 and 15.5 sec on its 6 sec cooldown. A/B +42.1 | verified |
| Thundering Strikes | 5 | +5% crit | SpellCrit and MeleeCrit +5.0. A/B +37.6 | verified |
| Ancestral Knowledge | 3 | +6% Intellect | Intellect +17. Through Mental Dexterity and Mental Quickness that is AP +17 and SpellPower +5.1. A/B +5.8 | verified |
| Mental Dexterity | 3 | 100% of Intellect as attack power | AttackPower +300 with 300 Intellect. A/B +46.0 | verified |
| Elemental Weapons | 3 | Windfury Weapon's bonus attack power +40%, applied once | Windfury base 819.9 against 787.5 without the talent on the same seed: +32.4, which is 333 × 0.4 = 133.2 AP / 14 × 3.4 sec. A/B +5.5 | verified |
| Shamanistic Focus | 1 | −45% shock cost | Earth Shock costs 247.5 (450 × 0.55), Flame Shock 225.5 (410 × 0.55). A/B +36.9 | verified |
| Flurry | 5 | +25% attack speed for 3 swings after a melee crit, at most one charge per 0.5 sec | Swing 3.37 sec → 2.69 sec (×1.253). With Rage of the Farseer 2.59 → 2.07 sec, so the two multiply. 3 charges show in the log. Uptime 68.6%. A/B +55.2 | verified |
| Stormstrike | 1 | 8 sec cooldown. The mark gives +20% to the next Earth Shock, Lightning Bolt or Chain Lightning | Casts at 0, 8 and 17 sec. Marks were used up by Earth Shock 11 times and Lightning Bolt 9 times, never by Flame Shock or Fire Nova. ×1.2 as above. A/B +150.2 | verified |
| Spirit Weapons | 1 | −30% threat, parry | Utility. A/B 0.0 | not DPS |
| Mental Quickness | 2 | Spell power = 30% of Intellect | SpellPower +90 with 300 Intellect. A/B +24.0 | verified |
| Improved Stormstrike | 2 | Each Stormstrike has a 100% chance (50% a point) of a 15 sec buff for 50% of spirit regen while casting | 22.24 procs per 22.2 Stormstrikes. Casting regen 33.4 mana/sec = 12 + 0.5 × 42.75 (292 Spirit). A/B +21.0. The cooldown reset on dodge or parry needs something to hit the shaman | verified (the reset is inert here) |
| Maelstrom Weapon | 5 | −20% cast time and cost per stack on Lightning Bolt (−4% a point) | At 5 stacks Lightning Bolt logs `Cost = 0.000, Cast Time = 0s`. Only Lightning Bolt is affected, matching the buff's class mask. A/B +333.0 | partly verified: the proc rate, 2 PPM a point, is the sim's assumption. The client has a 100% chance with no PPM, and a dummy effect of 50 it doesn't explain |
| Rage of the Farseer | 1 | +30% attack speed for 25 sec, 3 min cooldown | Swing 3.37 sec → 2.59 sec (×1.30) during the buff. Cast at 0, fades at 25.0 sec. A/B +18.3 | verified |

## Enhancement Level 60 + Imp GW (`050533102-051032031005112251`), 788.0 DPS

It is the Level 60 build with 2 points moved from Ancestral Knowledge to Improved Ghost Wolf. The shared talents show the same evidence:

| Talent | A/B |
|---|---|
| Concussion | +6.1 |
| Reverberation | +18.7 |
| Call of Flame | +22.7 |
| Elemental Devastation | +17.9 |
| Elemental Focus | +12.2 |
| Improved Fire Nova | +41.5 |
| Thundering Strikes | +37.6 |
| Mental Dexterity | +44.4 |
| Elemental Weapons | +5.6 |
| Shamanistic Focus | +37.9 |
| Flurry | +55.4 |
| Stormstrike | +150.1 |
| Mental Quickness | +23.2 |
| Improved Stormstrike | +22.0 |
| Maelstrom Weapon | +331.3 |
| Rage of the Farseer | +18.6 |

All are verified, and Maelstrom Weapon's proc rate stays an assumption. The talents that differ:

| Talent | Points | Expected (client) | Evidence | Status |
|---|---|---|---|---|
| Ancestral Knowledge | 1 | +2% Intellect | Intellect +6. A/B +2.2 | verified |
| Improved Ghost Wolf | 2 | Ghost Wolf cast time −3 sec | Utility. A/B 0.0 | not DPS |

## Enhancement Level 60 + Water Shield (`050031-053030031005112151-052030001`), 750.8 DPS, Magma Totem rotation

| Talent | Points | Expected (client) | Evidence | Status |
|---|---|---|---|---|
| Concussion | 5 | +5% on Lightning Bolt, Chain Lightning and Earth Shock | As on Level 60. A/B +5.3 | verified |
| Call of Flame | 3 | +15% on fire spells and totems | Magma Totem pulse 94.8 → 109.0 and Fire Nova 565.7 → 650.6 (both ×1.15; this build has no Improved Fire Nova). Flame Shock as on Level 60. A/B +23.3 | verified |
| Elemental Devastation | 1 | +3% melee crit | `Gained {"MeleeCrit": 3.000} from {SpellID: 30165}`. Uptime 36.5%. A/B +5.3 | verified |
| Thundering Strikes | 5 | +5% crit | Stat difference +5.0. A/B +35.0 | verified |
| Ancestral Knowledge | 3 | +6% Intellect | Intellect +17. A/B +5.0 | verified |
| Mental Dexterity | 3 | 100% of Intellect as attack power | AttackPower +300. A/B +44.7 | verified |
| Elemental Weapons | 3 | Windfury bonus AP +40% | As on Level 60. A/B +5.4 | verified |
| Shamanistic Focus | 1 | −45% shock cost | Flame Shock costs 225.5. A/B +23.4 | verified |
| Flurry | 5 | +25% attack speed | Uptime 64.4%. A/B +48.9 | verified |
| Stormstrike | 1 | +20% mark, 8 sec cooldown | As on Level 60. A/B +138.0 | verified |
| Spirit Weapons | 1 | −30% threat | Utility. A/B 0.0 | not DPS |
| Mental Quickness | 2 | Spell power = 30% of Intellect | SpellPower +90. A/B +22.9 | verified |
| Improved Stormstrike | 1 | 50% chance of the regen buff | 10.99 procs per 21.9 Stormstrikes (50.2%). Uptime 68.7%. A/B +9.4 | verified |
| Maelstrom Weapon | 5 | −20% per stack on Lightning Bolt | As on Level 60. A/B +304.2 | partly verified (proc rate assumed) |
| Rage of the Farseer | 1 | +30% attack speed | As on Level 60. A/B +17.2 | verified |
| Totemic Focus | 5 | −25% totem cost | Magma Totem costs 487.5 (650 × 0.75). A/B +3.8 | verified |
| Mindfulness | 2 | 33% of spirit regen while casting | Casting regen 26.1 mana/sec = 12 + 0.33 × 42.75. A/B +8.4 | verified |
| Tidal Focus | 3 | +3% hit | SpellHit and MeleeHit +3.0. Melee is already at the 9% cap from gear, so it helps the spells. A/B +9.4 | verified |
| Water Shield | 1 | Mana when hit | Up 100% of the fight from before the pull. Nothing hits the shaman in the preset. A/B 0.0 | not verifiable here (inert at 0 hits) |

## Enhancement Level 60 + Mana Tide (`02-055030031005112151-053051001001`), 729.0 DPS, Mana Tide rotation

These talents show the same evidence as on Water Shield or Level 60:

| Talent | A/B |
|---|---|
| Thundering Strikes | +33.6 |
| Mental Dexterity | +45.9 |
| Elemental Weapons | +5.3 |
| Shamanistic Focus | +7.2 |
| Flurry | +46.6 |
| Stormstrike | +134.1 |
| Mental Quickness | +22.9 |
| Improved Stormstrike 1 | +1.6 |
| Maelstrom Weapon | +289.8 |
| Rage of the Farseer | +15.9 |
| Totemic Focus | +0.6 |
| Mindfulness 3 | +3.8 |
| Tidal Focus 5 | +15.6 |

All are verified, and Maelstrom Weapon's proc rate stays an assumption. The talents that differ:

| Talent | Points | Expected (client) | Evidence | Status |
|---|---|---|---|---|
| Concussion | 2 | +2% on Lightning Bolt, Chain Lightning and Earth Shock | A/B +2.2, about 2/5 of the +5.3 to +6.1 for 5 points on the other builds | verified |
| Ancestral Knowledge | 5 | +10% Intellect | Intellect +29. A/B +7.2 | verified |
| Mana Tide Totem | 1 | 4 × 290 mana | As on Elemental. A/B +0.1 | verified |
| Water Shield | 1 | Mana when hit | Inert at 0 hits. A/B 0.0 | not verifiable here |
| Improved Reincarnation | 1 | +2% health | Health +102.8. A/B 0.0 | not DPS |

## Not verifiable, and why

- **Maelstrom Weapon's proc rate.** The client's proc entry has a 100% chance with no PPM, plus an unexplained dummy effect of 50. The sim's 2 PPM a point stays an assumption until the beta reaches level 34 (notes.md). The cast time and cost per stack are verified.
- **Elemental Focus's trigger set.**
  - The client's proc flags (0x15550) also include melee-class spells such as Stormstrike, and positive spells such as totems. The tooltip says "damage spell".
  - The sim triggers it on shaman damage spells only. That gives 10.0 to 10.1% of those casts, as measured above.
  - If Stormstrike can trigger it too, the sim gives Enhancement too few Clearcasting procs (Stormstrike is 22 casts against the 68 eligible ones). Only a beta log settles it.
- **Inert in the presets.** Water Shield, and Improved Stormstrike's cooldown reset, need the boss to hit the shaman, and the preset encounter has none.
- **Unexplained dummy effects.** Mental Dexterity effect 1 (20) and Spirit Weapons effect 2 (30) are dummy effects with no curve, and the client data doesn't say what they do. The sim doesn't model them.
