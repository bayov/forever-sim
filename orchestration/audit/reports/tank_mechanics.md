# tank_mechanics pass 1 (2026-11-04, master f62bfc74b, client 1.60.1.70755+hf)

This was the first pass, forced by hand, after the four tank specs landed and were realigned to 70753 and the November notes. It covered every ability and talent the tank presets use, the engine and the tests. Five subagents ran it in parallel (one per spec, plus the engine). All read-only. Proposed patches are in `ref/tank_mechanics/patches/`.

**Getting current:** the newest build is 70755, unchanged from 70753 for the tank rows (ledger). Blizzard's posts were read through 2026-11-04T02:12Z by today's mechanics pass 65, so this pass took its `dev_notes_seen`. Hotfix 113542 (Vengeance, Two-Handed Specialization) is Retribution only, so no tank row is pending.

## Protection warrior
| Ability / talent | Client | Sim | Verdict | Tier |
|---|---|---|---|---|
| Defensive Stance + Defiance | 7376 −10% taken, −10% dealt, +30% threat; Defiance 12792 | 0.9 / 0.9 / 1.8×(1+Defiance), shield only | ok (80%: Nov 2 notes) | 1 + notes |
| Shield Block 2565 | 75%, 7 s, 2 charges, 10 Rage, 5 s | same | ok | 1 |
| Shield Slam 23925 | 655 ±4.57% + block value, 20 Rage, 6 s | same; threat +508 | ok; threat is an open call | 1 / none |
| Revenge 11601 | 121 base, 5 Rage, 5 s, Defensive State 5 s | 97–145 + 0.28 AP, 2.25× threat | AP ratio open (-4) | 1 / 2 |
| Sunder Armor 11597 | −450 ×5, 15 Rage, threat 206 | 206 + 0.05 AP | ok | 1 + 2 |
| Heroic Strike 11567 | +138, 15 Rage | threat +142.2 | ok | 1 + 2 |
| Thunder Clap 11581 | 103, +20% attack time, 20 Rage, 6 s | 103 + 0.07 AP, 2.5× threat | ok; threat is an open call | 1 / 2 / 3 |
| Bloodrage, Improved Bloodrage, Shield Wall | 2687 / 12301 / 871 | same | ok | 1 |
| Shield Specialization, Anticipation, Improved Revenge, Master of Defense, Improved Sunder, Improved HS, Focused Rage, Bastion, Deflection, Cruelty, Enrage, Blood Craze | talent rows and curves | same | ok | 1 |
| Unbridled Wrath 12322 | proc mask 4 (auto attacks) | also procs on Heroic Strike | wrong (-6) | 1 |

## Protection paladin
| Ability / talent | Client | Sim | Verdict | Tier |
|---|---|---|---|---|
| Righteous Fury 25780 + Improved RF 20468 | Holy +90%, all +33%; −3% a rank | 1.33 × 1.9 Holy, 0.97 a rank | ok (multiplicative stacking unlogged) | 1 |
| Holy Shield r4 1239570 | 240 mana, 30% block, 221, coef .08, 4 charges, 10 s; ×1.2 threat from tooltip only | same | ok | 1 / 3 for threat |
| Seal of Fury r7, Judgement of Fury | 20423 / 20418 / 20414 | same | ok (8.76% spread not in a table) | 1 |
| Judgement, Swift Judgement, Holy Strike r8, Consecration r5, Templar's Bulwark | 20271 / 1310994 / 10333 / 20924+1280349 / 1311015 | same | ok | 1 |
| Iron Creed, Shieldsworn, Redoubt, Shield Specialization, Toughness, Anticipation, Deflection, Precision, Conviction, Improved Judgement, Divine Strength/Intellect, Improved Seals | talent rows and curves | same | ok | 1 |
| Reckoning 20177 | proc mask 0x202A8 includes spell crits | melee/ranged crits only | wrong, low (-9) | 1 |
| Redoubt rank scaling | curve 2..10 on a missing effect 1, ProcChance 10 | flat 10% | unverified; presets take 5/5 | 1 |

## Feral tank (bear)
| Ability / talent | Client | Sim | Verdict | Tier |
|---|---|---|---|---|
| Dire Bear Form 9634/9635 | +360% armor, 600+32/lvl health, AP | ×4.6, 1240, 180 | ok | 1 |
| Form threat 21178 | ×1.5; Nov 2 notes 115% | ×2.15 | ok (notes-only) | 1 notes |
| White-hit Rage | standards 21: 11.25 a swing, ×2 on a crit | Classic damage formula | wrong (-7) | 1 notes + 2 |
| Maul r7 9881 | +128, 15 Rage | same | ok; extra threat unverified | 1 |
| Primal Bite r3 1238070 | weapon + 77, 6 s, 20 Rage; notes: threat ×2 | threat ×1 | wrong (-2) | 1 notes |
| Lacerate r5 1235827 | 15 a stack / 3 s, 15 s, 10% weapon a stack | same | ok; "high threat" unmodelled | 1 |
| Demoralizing Roar r5 9898, Enrage 5229, Frenzied Regeneration 22842, Barkskin 22812 | rows | same (Enrage's 16% armor cut is a call) | ok | 1 |
| Thick Hide, Ferocity, Shredding Attacks, Savage Fury, Natural Instinct, Sharpened Claws, Predatory Strikes, Feral Swiftness, Natural Reaction, Blood Frenzy, LotP, Heart of the Wild, Naturalist | rows and curves | same | ok | 1 |
| Faerie Fire (Feral) threat (not cast) | 70425 notes: makes threat in bear form; value server-side | none | unverified, no finding | 1 notes |

## Warden shaman
| Ability / talent | Client | Sim | Verdict | Tier |
|---|---|---|---|---|
| Rockbiter 16313 AP | 653 at 60 | same, ×Elemental Weapons | ok | 1 |
| Rockbiter threat, white | 20871 bp 72 "threat per second" | 72 × speed | ok | 1 (designer note) |
| Rockbiter threat on Stormstrike | procs on specials | 72 × speed | open (-5) | 2 |
| Spirit Weapons 16268 | −30% threat, +86% to Rockbiter | 0.7, 1.302 | ok | 1 |
| Earth Shock r7 10414 | 450 mana, 6 s; notes 517–545 | same; ×2 threat | ok; ×2 is user call (a), tier 2 (SoD) | 1 notes / 2 |
| Stormstrike, Lightning Shield r7, Stoneskin r6, Mana Spring r4 | rows; notes | same | ok | 1 |
| Anticipation, Toughness, Improved Reincarnation, Tidal Focus, Thundering Strikes, Elemental Warding, Elemental Weapons, Improved LS, Shamanistic Focus, Concussion, Reverberation, Improved Stormstrike, Mental Dexterity / Quickness | rows and curves | same | ok | 1 |

The Nov 2 notes give the other tanks more threat to make up for Salvation's removal, and the Warden nothing. The sim matches.

## Engine
| Check | Verdict |
|---|---|
| Boss white table: order, base rates vs L60, Defense, crush by level only | right (tier 3) |
| Boss crit: Forever 0.6% / `classic_crit` 5.6%, minus Defense | right; rolled crit matches the sheet in both modes |
| Boss melee ability: one roll, own crit roll, no crush | right |
| Raid damage events: melee ones roll avoidance, ticks don't | right |
| Armor A/(A+400+85L), 75% cap; damage order | right (Classic) |
| Block chance and value | right (Classic) |
| Rage and mana gain threat | right |
| Salvation off, OwnDebuffs, SafeDPS threat per damage | right (settled) |
| Sheet, boss's side, vs rolled outcomes (2–3k iterations, four presets) | match |
| Sheet, our side (boss dodge/parry from the front) vs runs | **mismatch: the runs roll from behind (-1)** |
| Parry haste | never fires (-1); counted twice in encounters (-3) |
| Scoring run reads no health | holds for the presets at 3 HPS each; latent hole (-8) |

Not flagged: no dodging, parrying or blocking while casting (spell_outcome.go:673) rests on tier 3 only, and only the Warden casts. Worth a source check.

## New findings
Two engine findings are high: -1, every tank stands behind the boss; -3, parry haste is counted twice once -1 lands. -2 is also high: Primal Bite misses its ×2 threat. -4 (Revenge's AP ratio, needs a ruling), -5 (Rockbiter on Stormstrike, needs a ruling), -6 (Unbridled Wrath on Heroic Strike) and -7 (the bear's white-hit Rage) are medium. -8 (health read in scoring runs), -9 (Reckoning on spell crits), -10 (docs) and -11 (untested mechanics) are low.
-1, -2 and -3 change the tank ranking, so the tank presets and docs need a rerun once they land.
