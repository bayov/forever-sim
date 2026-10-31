# Findings: Forever client updates

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### client_update-20261030-2: Missile Barrage's "missiles fire every 0.5 sec" effect becomes a dummy in 70707
- severity: low until verified (the sim's model matches the unchanged tooltip; if the server now fires the base 1 s ticks in a halved channel, arcane loses about half its Missile Barrage missiles)
- affects: mage Arcane (and any build that casts Arcane Missiles on Missile Barrage)
- found: 2026-10-30, master 36a772f
- evidence: 400589 SpellEffect 1045767 (effect 1): aura 107 op 19 (tick period) bp −500 mask 2048 → aura 4 (dummy), misc 0, mask 0. Effect 0 (−50% duration) and effect 2 (−100% cost) unchanged; the tooltip still says "missiles fire every ${$400588w2/10}.1 sec". A dummy usually means a server script does the job, so probably no change in game. Sim: `sim/mage/arcane_missiles.go:98` ("the same missiles every 0.5 sec"). Ledger MAGE-U-27.
- proposed fix: none unless a combat log shows 2-3 missiles in a Missile Barrage channel instead of 5; check the next arcane log (log-audit).
- needs: verify (combat log)
- 2026-10-30: the 70707 port (client_update-20261030-1) kept the sim's five missiles in half the time.

### client_update-20261030-4: Fengus' Ferocity has only dummy effects in 70707
- severity: low (no report request uses Dire Maul world buffs)
- affects: any build that turns on Fengus' Ferocity
- found: 2026-10-30, master 36a772f
- evidence: 22817 effects 0-3 (aura 99 +200 AP, 124 +200 RAP, 13 +230 spell damage, 135 healing) all become aura 4 in 70707. Either the buff is dead or a script applies it. Sim: `sim/core/buffs.go:2244` gives 200 AP and 200 RAP. Ledger SHR-U-16.
- proposed fix: none until confirmed in game; if it is dead, drop it under Forever.
- needs: verify (in game)

### client_update-20261027-3: Amplify Curse no longer amplifies Bane of Agony in 70630
- severity: low (no preset or APL casts Amplify Curse)
- affects: warlock (Amplify Curse users only)
- found: 2026-10-27, master d3df924
- evidence: 70630 SpellEffect 695823 (18288 effect 0, aura 108 +50% all effects) EffectSpellClassMask_0 33792 → 32768: Bane of Agony (mask 1024) dropped, Curse of Weakness (32768) kept; new effect 1365363 −20 on Curse of Exhaustion (the 10-16 FC hotfix, WL-U-HF1). Bane of Agony 980-11713 gain an effect 1 aura 271 with base 0 and its own mask, nothing fills it in the client. Sim: `sim/warlock/curses.go:74` still multiplies Bane of Agony's base damage by 1.5 under Amplify Curse. Ledger WL-U-17.
- proposed fix: under Forever, don't let Amplify Curse touch Bane of Agony (or drop the option from the warlock APL actions).
- needs: orchestrator fix
- 2026-10-27: the 70630 mask reading holds, but 18288's 70630 description and Wowhead's Forever tooltip still name Bane of Agony, and every Agony rank gains a base-0 aura 271 effect (w2) that its new debuff text multiplies by, so the +50% looks moved onto Agony as a server-filled value, not removed. The sim keeps +50% (no preset casts Amplify Curse). needs: user ruling.
- 2026-10-27 (user): park it until a build takes Amplify Curse, then test in game (TASKS.md item of 2026-10-27). The sim keeps +50% of base meanwhile. needs: in-game test
- 2026-10-30 (70707): every Bane of Agony rank gains a third effect (index 2, dummy aura 4), and the tooltip's amplify term moves from $w2 to $w3, so the client still expects Amplify Curse to touch Agony through a scripted effect. 18288's mask is unchanged. Ledger WL-U-18. Still parked.

## Resolved

### client_update-20261030-3: Efficiency now discounts Strider Kick
- severity: low (mana only; Survival builds with Efficiency)
- affects: hunter Survival (any build with Efficiency and Strider Kick)
- found: 2026-10-30, master 36a772f
- evidence: 70707 19416 effect 0 EffectSpellClassMask_2 8192 → 2105344, adding bit 0x200000, which is Strider Kick's (1317257; also SoD's Flanking Strike 415320). The sim says the opposite: `sim/hunter/strider_kick.go:13` "Efficiency doesn't", and `talents.go:61` leaves it out. Ledger HUN-U-13. Not measured; 3% a point off Strider Kick's 5.81% base mana every 8 s.
- proposed fix: under Forever add `SpellCode_HunterStriderKick` to Efficiency's list in `sim/hunter/talents.go:63` and fix the comment.
- needs: worker fix
- resolved: 2026-10-31, e1fcd57b7: Efficiency's list has Strider Kick (sim/hunter/talents.go); the hunter builds re-simmed.

### client_update-20261030-1: Client 70707 confirms the mage talent rework the user parked on 2026-10-29 (Critical Mass 3 ranks, Incineration 5, Arcane Focus gone, Precision on all spells, Accelerant on Arcane Wrath, Ice Lance cooldown, tree moves)
- severity: high (fire −3.5% from Critical Mass 5 → 3 alone, −2.4% with Incineration 5; every mage talent string changes)
- affects: mage, all four report builds (arcane, fire, frost, frostfire)
- found: 2026-10-30, master 36a772f
- evidence: client 1.60.1.70707 (wago.tools tables, sure): TraitNodeEntry 130513 Critical Mass MaxRanks 5 → 3 (curve 83083 now 2/4/6), 130525 Incineration 3 → 5 (curve 83095 adds 8/10), TraitNode 105814 Arcane Focus removed, 29438 renamed Precision with effect 0 aura 107 → 55 (2/4/6% hit, all spells), 11069 renamed Accelerant with mask word 1 4096 → 36864 (adds Arcane Wrath), Ice Lance 1312002 CategoryRecoveryTime 0 → 8000 (rank 1 only: ranks 2-6, the sim's 1240047 included, have no cooldown row), TraitDefinitions 135332 / 135339 swap Missile Barrage and Arcane Resilience (Missile Barrage to tier 2, Arcane Resilience to tier 1), Improved Fire Blast tier 1 → tier 3, Improved Channeling moves left, Frost Blast leaves the spellbook. foreverchanges.pro "Build 1.60.1.70707" says this build "puts the server's hotfixes of October 27 to 29 into the client files: the mage trees' rework, Ice Lance's 8 sec cooldown". Ledger rows MAGE-U-20 to MAGE-U-26. TASKS.md (2026-10-29 mage item) parked it "until a source confirms them (... or a newer client build)": this is that build. Measured (fire request edited in scratch, 180 s, 20k iterations, seeds 1 / 1000003): 1089.03 / 1088.66 → Critical Mass 3: 1051.59 / 1050.76 (−37.7, −3.5%); Critical Mass 3 + Incineration 5: 1063.87 / 1062.52 (−25.7, −2.4%), before re-spending the freed points. Arcane estimate: Arcane Focus 5 (5%) gives way to Precision 3 (6% on every school), so about +1% Arcane hit and 5 free points; not measured.
- proposed fix: under Forever, in `sim/mage/talents.go` drop Arcane Focus (102), widen Elemental Precision to every school (214), keep Incineration and Critical Mass per point (184, 195) but update their comments; Accelerant on Arcane Wrath (`arcane_wrath.go:64`, as `fireball.go:76`); Ice Lance's 8 s cooldown, at least for rank 1 (`ice_lance.go:72`; rank 6 needs FC's word or the game); regenerate `ui/core/talents/trees/mage.json` from 70707 (update-sources talent trees) and re-pick the four mage presets' talents (fire: 2 points out of Critical Mass, 2 into Incineration, Arcane Focus's 4 elsewhere; arcane: Arcane Focus's 5 elsewhere).
- needs: orchestrator fix (the user's parking condition is met; tell them)
- resolved: 2026-10-30, the sim takes 70707's mage tree (regenerated client/data/tree json and proto), drops Arcane Focus, gives Precision 2% a point to every school, puts Accelerant on Arcane Wrath and an 8 sec cooldown on every Ice Lance rank (user ruling "All ranks", assumption mage_ice_lance_cd). Presets re-searched at 180 s, 20k iterations, seeds 1 / 1000003: Fire 1101.0 → 1061.0, Frostfire 1096.4 → 1051.0, Arcane 901.6 → 915.5, Frost 946.6 → 908.3 (with new Frost knobs). forever-wiki/classes/mage.md "Client build 1.60.1.70707".

### client_update-20261028-2: Launch-realm hotfix puts Trueshot Aura ranks 1-4 at 15/22/30/40, which makes rank 5's 50 look intended (ruling hunter_tsa_hm_rank assumes 100)
- severity: high (−0.9% to −1.3% on three hunter presets if the ruling flips)
- affects: hunter BM, Lone Wolf, Marksmanship (Survival −0.09%); TSA is ranged-only under Forever, so no other class
- found: 2026-10-28, master 224b8a4
- evidence: foreverchanges.pro/patch-notes, "October 28, 2026 Hotfix" (read from the launch realms' hotfix cache, not in Blizzard's notes; FC tier, not checkable from the cloud since `tools/sources/dbcache.py` needs a local client): "Trueshot Aura: gives 15, 22, 30 and 40 ranged attack power by rank instead of 30, 40, 50 and 75". Rank 5 (20906) is not listed, so it stays 50 (client since 70338). With 15/22/30/40/50 the top rank sits on the curve, which undercuts the 2026-10-15 ruling that Blizzard would lift rank 5 to 100. Sim: `sim/core/buffs.go:1748-1751` (rank 5 = 100 by ruling; ranks 1-4 30/40/50/75, only matter below 60). Measured with ranks 15/22/30/40/50 (180 s, 20k iterations, paired seeds 1 and 1000003): beast-mastery 1150.37 → 1139.50 (−10.87, −0.94%), lone-wolf 1118.87 → 1104.50 (−14.37, −1.28%), marksmanship 1137.33 → 1124.58 (−12.75, −1.12%), survival 1029.84 → 1028.89 (−0.09%). Ledger HUN-HF-02.
- proposed fix: user re-rules hunter_tsa_hm_rank (keep 100, or take 50 now that the whole curve is half); either way set ranks 1-4 to 15/22/30/40 under Forever once confirmed in game or in a local DBCache read.
- needs: user ruling (and verify: FC-tier hotfix)
- resolved: 2026-10-28, user ruling ("Adopt both"): the sim takes the launch-realm hotfix, Trueshot Aura 15 / 22 / 30 / 40 / 50 and Aspect of the Hawk 15 / 22 / 30 / 38 / 47 / 55 with rank 6 cast at 60 (the hunter rotations name 14322 now); hunter_tsa_hm_rank one value.

### client_update-20261028-1: Launch-realm hotfix halves Aspect of the Hawk ranks 1-5 (rank 5 90 → 47), so rank 6's 55 becomes the best at 60
- severity: medium (−0.66% to −0.90% on three hunter presets)
- affects: hunter BM, Lone Wolf, Marksmanship (Survival 0)
- found: 2026-10-28, master 224b8a4
- evidence: foreverchanges.pro/patch-notes, "October 28, 2026 Hotfix" (launch-realm hotfix cache, FC tier, not in Blizzard's notes, not in client 70658's tables): "Aspect of the Hawk: gives 15, 22, 30, 38 and 47 ranged attack power by rank instead of 20, 35, 50, 70 and 90". Rank 6 (14322, 55) is not listed. Sim: `sim/hunter/aspects.go:46` rap 20/35/50/70/90/55/120 and `getMaxHawkRank` caps Forever at rank 5 (assuming rank 6 is worse). Measured with rank 5's value at 55, i.e. the hunter casting rank 6 (paired seeds, 20k): beast-mastery 1150.37 → 1142.76 (−7.61, −0.66%), lone-wolf 1118.87 → 1108.81 (−10.06, −0.90%), marksmanship 1137.33 → 1128.41 (−8.92, −0.78%), survival 0. Ledger HUN-HF-01.
- proposed fix: once confirmed, under Forever set ranks 1-5 to 15/22/30/38/47 and let `getMaxHawkRank` pick the best learned rank (rank 6, 55, from level 58) instead of capping at 5; fix the comment.
- needs: verify (in game or local DBCache read), then orchestrator fix
- resolved: 2026-10-28, user ruling ("Adopt both"): the sim takes the launch-realm hotfix, Trueshot Aura 15 / 22 / 30 / 40 / 50 and Aspect of the Hawk 15 / 22 / 30 / 38 / 47 / 55 with rank 6 cast at 60 (the hunter rotations name 14322 now); hunter_tsa_hm_rank one value.

### client_update-20261027-1: Missile Barrage no longer procs from Arcane Missiles hits (70630 tooltip)
- severity: high (−7.9% on mage__arcane if real)
- affects: mage Arcane (any build with Missile Barrage that casts Arcane Missiles)
- found: 2026-10-27, master d3df924
- evidence: client 70630 Spell 400588 Description_lang drops "your Arcane Missile hits a ${$m1/4}% chance" and adds Frost Blast to the ${$m1/2}% list. Its SpellEffect, SpellAuraOptions and SpellClassOptions rows are identical in 70532 and 70630 (the proc is scripted), so the tooltip is the only client evidence; not in Blizzard's notes or on foreverchanges yet. Sim: `sim/mage/talents.go` applyMissileBarrage `OnSpellHitDealt` rolls 5% a point on every missile that lands. Measured with that handler removed (180 s, 20k iterations, paired seeds 1 and 1000003): mage__arcane 982.50 → 904.74 and 982.47 → 905.06 (−77.6, −7.9%); mage__frostfire 0. Ledger MAGE-U-15. Fairly sure the tooltip reflects a server change (the same rewrite brings in the 10-16 Frost Blast hotfix), but unconfirmed.
- proposed fix: confirm in game (Missile Barrage procs during a channel) or in a combat log, then drop the `OnSpellHitDealt` roll under Forever, re-optimize the Arcane talents and rotation (Missile Barrage may no longer be worth its points), and rerun the Arcane presets.
- needs: user ruling (take the tooltip or verify first), then orchestrator fix
- 2026-10-27: still open; the mage re-check found no dev note withdrawing it and post /6 (10-16) explicitly added it. Recommendation stays: verify in game before changing.
- resolved: 2026-10-27, user ruling ("Missile Barrage should collapse to a single value assumption, doesn't proc"): the missiles never proc it under Forever (sim/mage/talents.go), and mage_am_missile_barrage is a one-value row; patch 0004 drops mage_am_mb.

### client_update-20261027-2: Spellstone gives Shadow damage only since 70630, the sim adds Fire too
- severity: low (−0.13% to −0.28%)
- affects: warlock affliction, demonology and destruction presets and their duty-curse variants (all on Spellstone)
- found: 2026-10-27, master d3df924
- evidence: client 70630 SpellEffect 1233765/1233768/1233770 (spells 1237162/1237164/1237165, "Increase Shadow Dam", the Spellstone ranks) EffectMiscValue_0 36 (Fire | Shadow) → 32 (Shadow); bp 14/17/21 and the 1/2/2% haste effect unchanged. Sim: `sim/warlock/talents.go` applyWeaponImbue adds `stats.FirePower` and `stats.ShadowPower`. Measured without the FirePower (paired seeds, 20k): affliction 1029.38 → 1027.87 (−1.51), demonology 1060.35 → 1058.98 (−1.37), destruction 1011.91 → 1009.12 (−2.79), the duty-curse variants the same, destruction_fire (Firestone) 0. Ledger WL-U-16.
- proposed fix: drop the FirePower line from the Spellstone case under Forever, fix the comment above `firestoneRanks`, update the goldens and the warlock presets' published DPS.
- needs: orchestrator fix
- resolved: 2026-10-27, d6fcdd3 (Spellstone Shadow only under Forever; goldens and wiki updated).

### client_update-20261027-4: Touch of the Grave's drain loses its "ignore caster modifiers" attribute in 70630
- severity: low (Undead presets; the drain is modelled at its 5% max-health cap, a few DPS at most; not measured)
- affects: every Undead preset
- found: 2026-10-27, master d3df924
- evidence: 70630 SpellMisc 813545 (spell 1260198): Attributes_3 0x20000000 → 0, Attributes_9 0x1000000 → 0, Attributes_10 0x2 → 0, Attributes_13 0x4000000 → 0; Attributes_4 0x100 and Attributes_6 0x28000000 (added with them in 70425) stay. 70425's notes said the drain no longer scales with Shadow increases on target or caster, which the sim models with `SpellFlagIgnoreModifiers` (`sim/core/racials_forever.go` registerTouchOfTheGrave). Which half of the 70425 attributes carried that rule is not known, so this may change nothing. Ledger SHR-U-15.
- proposed fix: check in a combat log whether the drain on a Shadowform priest or a Shadow Mastery warlock still equals 5% of max health; if it grows, drop the caster-side ignore under Forever.
- needs: user ruling (log check)
- resolved: 2026-10-27, no sim change needed: the drain keeps Attributes_6 0x28000000 (ignore caster damage and damage-taken modifiers) and has no spell power coefficient, so it still takes no modifier; comment added in 500e3a1.

### client_update-20261027-5: Dev notes post /7 (2026-10-26) racial pass, not in client 70630
- severity: medium (high once it lands: it moves the race tables and the Human, Orc, Dwarf, Tauren and Skyborne presets; estimated, not measured)
- affects: every preset of those races, the race tables, warlock and hunter pets (Orc Command)
- found: 2026-10-27, master d3df924
- evidence: Blizzard development notes topic 2360696 post /7 (Kaivax, 2026-10-26 20:35, edited 21:09; tier: official). Human Ruthless 2% → 1% crit, Human Spirit 5% → 10%; Dwarf Earth Fury (2% crit damage) removed, Artisan's Precision +1% hit added; Tauren Endurance only 5% health (no hit or crit damage); Orc Savage (1% crit) removed, Command +5% warlock and hunter pet damage; Skyborne Wind-Blessed Haste 3% → 2%, Read Leyline 50% for 15 sec (100% while Attuned). Client 70630 has none of it (no Artisan's Precision spell; Ruthless 20597, Savage 20574, Endurance 20550 rows unchanged). Sim: `sim/core/racials_forever.go` Human 2% crit and Spirit ×1.05, Dwarf and Tauren `critDamageBonus(1.02)`, Orc 1% crit, Skyborne ×1.03 haste. Rough size: −1% crit is about −0.7 to −1% DPS, +1% hit for a Dwarf below the cap about +1%. Ledger SHR-P7-01..05.
- proposed fix: wait for the client build or hotfix that carries it and read the values there, or apply it from the notes now if the user prefers; then rerun the race tables and the affected presets.
- needs: user ruling (apply from the notes or wait for the client)
- resolved: 2026-10-27, 500e3a1, the post /7 racial pass applied from the notes (user asked for the sim to follow them); UI text in tools/sources/racials_hotfix.json.

### client_update-20261021-1: Champion of the Light may be back to 20/40/60% of Intellect (unannounced 10-16 hotfix)
- severity: medium (high if confirmed: it is spell power for every paladin preset that takes the talent; not measured)
- affects: paladin sub Holy, undead and twist presets (any build with Champion of the Light)
- found: 2026-10-21, master 599f9de
- evidence: foreverchanges.pro/patch-notes "Hotfix-2026-10-16" (third party, tier: unconfirmed): "Champion of the Light gives spell damage equal to 20/40/60% of your Intellect again, as before October 15 (10/20/30%)". Not in Blizzard's notes. Client 70532 (identical to 70450) can't confirm: the ranks come from trait curve 110350, and the Curve tables aren't downloaded. Sim: `sim/paladin/talents.go:49` uses `{0, .1, .2, .3}` (PAL-70425-05). Ledger row PAL-U-HF1.
- proposed fix: verify in game (the character sheet's spell power with 3/3 vs 0/3) or with the client's Curve/CurvePoint rows for curve 110350 in a hotfixed export. If confirmed, change the multipliers to `.2,.4,.6` (Forever-gated) and rerun the paladin presets.
- needs: user ruling (in-game check) then orchestrator fix
- resolved: 2026-10-23 (on master by 99bd53a; client 70630 now carries the values). confirmed in wago.tools' hotfix table. Push 112842 (2026-10-20 09:51 on 70450) sets CurvePoint 331369-331371 of curve 110350 to 20 / 40 / 60, and curve 110350 is Champion of the Light's (TraitDefinition 142634 for spell 1311084, TraitDefinitionEffectPoints 24621). No later push touches it. Applied (Forever's curve only, `sim/paladin/talents.go`): +34 to +39 at 180 sec on the twist, sub Holy and undead presets; Shock doesn't take the talent.

### client_update-20261021-2: Oct 20 mage Frost hotfix (Piercing Ice removed, Elemental Precision 3x2%), not in the client
- severity: low (needs verify; if real it moves Frost DPS by several percent, since Piercing Ice is +6% Frost damage at 3/3)
- affects: mage Frost and Frostfire presets; Missile Barrage builds (Frost Blast trigger)
- found: 2026-10-21, master 599f9de
- evidence: foreverchanges.pro "Hotfix-2026-10-20" (Piercing Ice removed; Elemental Precision 5x1% -> 3x2%; Ice Barrier +10% absorb and, by FC's wording only, +3% Frost damage) and "Hotfix-2026-10-16" (Frost Blast can trigger Missile Barrage). Not in Blizzard's notes. Client 70532 still has Piercing Ice node 130502 (def 135303, 3 ranks) and Elemental Precision entry 130507 MaxRanks 5, so it is server side or not real. Sim: `sim/mage/talents.go:170-191`. Already a master task (TASKS.md, commit 599f9de). Ledger rows MAGE-U-HF1..HF5.
- proposed fix: as in TASKS.md. Confirm through the wago hotfix table or in game, re-import the mage tree, drop Piercing Ice, make Elemental Precision 2%/rank over 3 ranks, re-optimize the Frost talents, and size it with dps_beforeafter.sh on mage__frost and mage__frostfire.
- needs: orchestrator fix (after verify)
- resolved: 2026-10-23 (on master by 99bd53a; client 70630 now carries the values). confirmed in wago.tools' hotfix table (push of 2026-10-20 09:51 on 70450; the trait, curve, Spell and SpellEffect rows are in update_ledger.md MAGE-U-HF1..4) and wowsims/forever 983ea6fcc1. Frostbite and Permafrost also swapped columns. Applied: tree, proto, client/mage.json, Elemental Precision 2%/rank, Ice Barrier with the 3% while shielded and Frigid Power, the presets and the Frost rotation. At 180 sec Frostfire 1081.9 → 1041.5, Frost 843.9 → 820.8, Fire 1067.6 → 1094.3 and Arcane 964.7 → 981.9 (the last two put the freed points in Arcane Subtlety). HF5 (Frost Blast and Missile Barrage) is untouched. The same hotfix push also has curve 110350 at 20/40/60 (CurvePoint 331369-331371), which looks like client_update-20261021-1's Champion of the Light.
