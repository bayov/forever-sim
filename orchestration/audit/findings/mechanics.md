# Findings: Forever mechanics

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### mechanics-20261103-1: Read Leyline is a 0.5 sec cast on client 70753; the sim casts it in 2 sec
- severity: low (no report request or preset casts it: `grep -c 1259705 tools/dps_report/req/*.json` is 0 everywhere, and the race tables show Read Leyline at 0.0 DPS; it would matter for a High Order build whose APL casts it, about 1.5 sec of casting saved every 2 min)
- affects: Skyborne High Order (shared racial; rotopt `leylineMana` knob, tools/rotopt/racials.go and hunter_ranged.go comments)
- found: 2026-11-03, master a0b306984
- evidence: tier 1, client 1.60.1.70753: SpellMisc of 1259705 CastingTimeIndex 5 -> 3 (SpellCastTimes 5 = 2000 ms, 3 = 500 ms); 70707 had 5. Dev notes post 7 only says "Read Leyline: Now castable while moving". sim/core/racials_forever.go:141 `CastTime: time.Second * 2`. Not checked whether this rode in on a 70707 hotfix (the hotfix fetch was cut short this pass); either way the newest tier 1 value is 0.5 sec. Sure: high on the client value.
- proposed fix: sim/core/racials_forever.go:141 `CastTime: time.Millisecond * 500`, update the comments at :111, tools/rotopt/racials.go:10 and tools/rotopt/hunter_ranged.go:33 ("a 0.5 sec cast"). No published number moves.
- needs: orchestrator fix

### mechanics-20261031-1: Efficiency should discount Lacerate (its class bit is in the mask)
- severity: low (mana only; no report build both takes Efficiency and casts Lacerate: Marksmanship / Lone Wolf have Efficiency 5 but no Lacerate in their APLs, Survival casts Lacerate 1299332 but has Efficiency 0). Not simmed: no published number moves.
- affects: hunter (any build with Efficiency and Lacerate)
- found: 2026-10-31, master e05fd732e
- evidence: tier 1, clients 1.60.1.70425 / 70450 / 70630 / 70707: Efficiency 19416 e0 (aura 108 misc 14, -3% a rank) EffectSpellClassMask_1 = -2146955264 = 0x80081000. Lacerate's SpellClassOptions (24118 / 24119 / 24120 / 1299332, set 9) mask word 1 = 4096 = 0x1000, inside it; `AUDIT_DB=<70707> python3 tools/audit/mechanics/mask.py 19416` lists Lacerate. The sim leaves it out: sim/hunter/talents.go:59 "Lacerate is not in it", and sim/hunter/talents_test.go:48 TestEfficiencyClassMask wants Lacerate at 100 because it reads the mask as "0x80080000" (the 0x1000 bit dropped). Resourcefulness (440529, word 1 0x80000) rightly skips Lacerate. Sure: high (four builds, two tables agree).
- proposed fix: add `SpellCode_HunterLacerate` to Efficiency's case list in sim/hunter/talents.go:63-67, fix the comment, and change the test's Lacerate row to 85 with the mask 0x80081000 (patch: $AUDIT_WORK/mechanics/efficiency_lacerate.patch, key hunk: `SpellCode_HunterSniperShot, SpellCode_HunterStriderKick, SpellCode_HunterLacerate:`).
- needs: worker fix

## Resolved

### mechanics-20261027-2: Amplify Curse no longer boosts Bane of Agony on client 70630
- severity: low (no preset takes Amplify Curse; rotopt's `amplifyCurse` knob defaults to 0)
- affects: warlock (only builds that take Amplify Curse)
- found: 2026-10-27, master d3df9247f
- evidence: tier 1, client 1.60.1.70630: 18288 e0 (aura 108 misc 8, +50%) mask 33792 -> 32768, which covers Curse of Weakness only (`AUDIT_DB=<70630> python3 tools/audit/mechanics/mask.py 18288`; on 70532 it listed Bane of Agony too). A new e1 (aura 107 misc 8, -20) covers Curse of Exhaustion again. sim/warlock/curses.go:74-77 still multiplies Bane of Agony's base damage by 1.5 and spends the aura.
- proposed fix: under Forever, Amplify Curse leaves Bane of Agony alone (drop the x1.5 and the deactivate in the Bane of Agony snapshot, or don't register the knob's cast), and update the forever-wiki talent text.
- needs: none
- 2026-10-27: the 70630 mask reading holds, but 18288's 70630 description and Wowhead's Forever tooltip still name Bane of Agony, and every Agony rank gains a base-0 aura 271 effect (w2) that its new debuff text multiplies by, so the +50% looks moved onto Agony as a server-filled value, not removed. The sim keeps +50% (no preset casts Amplify Curse). needs: user ruling.
- 2026-10-27 (user): park it until a build takes Amplify Curse, then test in game (TASKS.md item of 2026-10-27). The sim keeps +50% of base meanwhile. needs: in-game test
- resolved: 2026-11-03, round 70753 (warlock): Blizzard's development notes of 2026-11-02, "Amplify Curse now correctly affects Curse of Exhaustion and Bane of Agony" (foreverchanges.pro: "Now really works on Curse of Exhaustion and Bane of Agony, as its tooltip says"). The sim keeps +50% of Bane of Agony's base damage (sim/warlock/curses.go; TestWarlockAmplifyCurseAgony), and the round's talent search tries Amplify Curse.

### mechanics-20261101-1: Shadow presets stand at 25 yd but Mind Flay reaches 20 yd (24 with Shadow Reach 3/3); the sim casts it anyway
- severity: low (no DPS effect measured: priest__shadow 948.73 DPS at 25, 24 and 20 yd, 20k iterations, same seed; the preset is just not playable as set)
- affects: Shadow priest presets priest__shadow, priest__shadow-holy (mksettings spec_shadow.go:49, from 905787afd)
- found: 2026-11-01, master dd053d24c
- evidence: client 70707 SpellMisc RangeIndex 3 (20 yd) on every player Mind Flay rank (15407, 17311-17314, 18807); the APL channels 18807. Shadow Reach 17322 is aura 108 misc 5 (range) +20%, so at most 24 yd. The engine checks Spell.MaxRange only in leveling (sim/core/spell.go:548, apl_idle.go:424), so a 25 yd caster still channels Mind Flay. Warlock presets at 25 yd are fine (no Drain Life 689 ranks, also 20 yd, in their APLs; their spells are 30 yd). Smite / Holy Fire are 30 yd. Tier 1. Sure about the ranges; RangeIndex 3 = 20 yd is the standard SpellRange row (SpellRange.csv is not in the mechanics set).
- proposed fix: in spec_shadow.go set `player.DistanceFromTarget = 20` for the Shadow specs (or 24 if Shadow Reach 3/3 is in the talents; Smite may keep 25). Same result for explosives (still out of the 10 yd reach), no DPS change. Optionally give Mind Flay MaxRange 20 so a wrong distance shows.
- needs: worker fix
- resolved: 2026-11-02, fa2d49997 (standards 33). Shadow and Shadow + Holy stand at 22 yd, Mind Flay's 20 with Shadow Reach 1/2 (Forever's Shadow Reach has 2 ranks).

### mechanics-20261031-2: Moonkin Aura and Leader of the Pack are both 3% all crit on live 70707 (hotfix push 113283); the sim still gives the Era split
- severity: high if the two stack (+2.3 to +3.3% DPS on every preset with both), none if they don't (then the sim's 3% melee + 3% spell already equals one 3% all crit for those presets)
- affects: shared raid buffs (sim/core/buffs.go:292-303), every Level 60 preset with both buffs; the raid sim's crit-aura counting
- found: 2026-10-31, master d96adcce1
- evidence: wago.tools/hotfixes, push 113283 on 1.60.1.70707 (2026-10-31 10:59): SpellEffect 701335 (24907 Moonkin Aura) and 702727 (24932 Leader of the Pack) both effect 65 (raid area aura), aura 290 (all crit), base points 3; Spell 24907 text "Increases Critical Strike chance by $s1%." (was "spell Critical Strike"), 17007 text "...increases the Critical Strike chance of all party or raid members". The build's own table still has auras 57 / 52 (`sp.py 24907`, `sp.py 24932`). Tier 1 (server data on the current build); matches foreverchanges' October 30 hotfix (TASKS.md 2026-10-31 druid item). Reproduce: `curl -s 'https://wago.tools/hotfixes?search=24907'` (data-page props). Nothing in the client says whether two different aura-290 raid auras stack; under 70338 the sim assumed they did not, without a cited source. Measured +3% crit (bonus stats, seed 1, 20k iterations, report requests, CLI d96adcce1): Fury 1041.5 -> 1075.5 (+3.26%), Combat Swords 777.0 -> 796.3 (+2.49%), Fire 1061.0 -> 1090.3 (+2.77%), Elemental 772.7 -> 790.4 (+2.29%).
- proposed fix: under `IsForever()`, give each buff 3% MeleeCrit + 3% SpellCrit; if they stack, add both (each preset +3% crit); if not, give 3% for either (the pre-70425 code in 9016b8511's parent). Either way Balance's own Moonkin Form and Feral's own LotP become all crit (no DPS change for those specs). Update the buffs.go comment and the raid sim's aura count.
- needs: user ruling (stacking), then orchestrator fix
- resolved: 2026-11-01 (user ruling: they don't stack), f962f6c: under Forever either aura gives 3% MeleeCrit and 3% SpellCrit once (sim/core/buffs.go). Every report build with both is unchanged; druid__feral-cat 885.38 -> 885.57 and feral-cat-duty-ff 881.21 -> 881.26 (the cat's own aura now adds spell crit), Balance hash only; docs/raid_sim.md and forever-wiki/classes/druid.md follow. The raid sim's comp search needs no change (it sims through core). The DPS report is not regenerated yet.

### mechanics-20261030-2: DPS report and standards still say Windfury Totem 1.5 sec and extra attacks reset the swing, after the sim took 100 ms and no reset
- severity: medium (the numbers are right per the 2026-10-30 user rulings; the published assumption cards describe the opposite of what moved Arms +10%)
- affects: DPS report assumption cards (all melee: warrior, rogue, Ret paladin, Survival hunter, Feral, Enhancement), orchestration/standards.md 17 / 174 / 175
- found: 2026-10-30, master fb1485888
- evidence: f3017897f sets the totem's proc ICD to 100 ms under Forever (tier 1: 70707 10612 aura 42 proc 20% icd 100, verified) and keeps the swing timer on extra attacks (user ruling, notes.md to verify). fb1485888 re-simmed the report with both, but `tools/dps_report/assumptions_def.py:399-412` and `ui/dps_report/data/assumptions.json` still show `wf_totem_icd` "At most once every 1.5s, as in Classic." and `extra_attack_swing` "Yes. The next swing waits a full swing after an extra attack."; standards.md 17 ("internal cooldown is 1.5 sec ... the 100 ms alternative is closed"), 174 and 175 say the same. Effect of f3017897f alone (run_reqs, seed 1, 5000 iterations, bbbaf2e3c vs f3017897f CLIs): Arms 905.0 -> 995.4 (+10.0%), Fury 2H +6.6%, Fury +6.2%, Survival +4.3%, Shock Ret +3.3%, Mutilate +2.9%, Combat Swords +1.7%. In Arms, white swings +13% and Heroic Strike casts 2.7x, so the swing-timer ruling carries most of it. Eviscerate's 4% AP a combo point (same commit, user ruling) has no assumption card either.
- proposed fix: flip both cards (short/desc/settings: "100 ms, the client's 10612"; "No, the next swing comes when it was due", source wowsims/forever, to verify in game) and add one for Eviscerate 4% AP; rewrite standards 17 and 174/175 to the 2026-10-30 rulings; redeploy the report.
- needs: orchestrator fix
- resolved: 2026-10-31, e1fcd57b7: the `wf_totem_icd` and `extra_attack_swing` cards say 0.1s and the swing timer kept, a new `evisc_ap` card says 4% a combo point, standards 17 and 28 follow; the report is redeployed.

### mechanics-20261030-1: Client 70707 widens Tauren Endurance's 2% critical damage to every school, against post 7's removal
- severity: high if the bonus is live (estimate, not measured: about +1.0% for a crit-heavy Tauren melee build at ~35% crit and 2x crits, +0.4% for a 15% crit caster; it can move Tauren in the race tables); none if post 7 stands
- affects: Tauren (every class and build in the race tables that Tauren can be)
- found: 2026-10-30, master 36a772fd1
- evidence: tier 1, client 1.60.1.70707: 20550 e1 aura 163 bp 2 EffectMiscValue_0 0 -> 127 (all schools); e2 aura 50 bp 2 (crit healing) kept; tooltip still "critical strike bonus to damage and healing increased by $s2%" (`AUDIT_DB=<70707> python3 tools/audit/mechanics/sp.py 20550`, against 70685). The same build bakes in Blood Fury's 2 min cooldown (20572 RecoveryTime 180000 -> 120000) from post 7 / 8, so the client team edited the racials after the notes, yet kept and widened Endurance's crit effect. Tier 1 against it: dev notes post 7 (2026-10-26, edited 10-27 16:47): "Endurance no longer provides any modifiers beyond 5% increased Health (no hit chance or critical damage modifiers)." The sim follows post 7 (sim/core/racials_forever.go:83-87, health only). Server side may still strip it; the client alone can't settle it.
- proposed fix: user ruling. If the bonus is live: under Forever, Tauren get a permanent crit-damage multiplier of 1.02 on every school (aura 163, multiplying the whole crit multiplier, as other 163 effects in the sim do) in racials_forever.go's Tauren case, then re-sim the Tauren rows. Otherwise note in verdicts.md that post 7 overrides the client row.
- needs: user ruling
- resolved: 2026-10-31 (user ruling): Tauren get 5% health only, no crit damage, as post 7 says; the client row doesn't override it for now. No sim change (sim/core/racials_forever.go already health only); verdicts.md notes it.

### mechanics-20261027-3: Blood Fury, Berserking and Touch of the Grave still use the pre-post-8 values
- severity: medium (estimate, not measured: Blood Fury uptime 15/180 -> 15/120 of +10% AP/SP and Berserking 10 -> 15 sec of 10% haste each look like +0.2-0.4% DPS on Orc / Troll builds; Touch of the Grave is a drain, no DPS)
- affects: all Orc, Troll and Undead builds (shared racials)
- found: 2026-10-27, master 37a8a8ccd
- evidence: development notes post 8 and the edited post 7 (tier 1, us.forums.blizzard.com/en/wow/t/2360696/8, 2026-10-27 15:37, edited 16:46): "Blood Fury's cooldown reduced to 2 minutes (was 3 minutes)", "Berserking duration increased to 15 seconds (was 10 seconds)", "Touch of the Grave internal cooldown reduced to 15 seconds (was 25 seconds)". Server side: client 70634 is byte-identical to 70630 on all 25 mechanics-set tables (20572 RecoveryTime 180000, 1260189/1260201 ProcCategoryRecovery 25000). Master: sim/core/racials_forever.go Blood Fury `time.Minute * 3`, Touch of the Grave ICD `time.Second * 25` (line 546), sim/core/racials.go makeBerserkingCooldown aura `time.Second * 10`.
- proposed fix: PR #117 (forever-update/2026-10-27-70634) makes exactly these three changes, Berserking 15 sec only under Forever; reviewed, matches the notes. Merge it and re-sim the Orc, Troll and Undead builds.
- needs: orchestrator fix
- resolved: 2026-10-28, ae8522c14 (via PR #117 merge 297226d14); verified pass 44: Blood Fury 2 min, Berserking 15 sec under Forever only (3 min CD, flat 10%), Touch of the Grave ICD 15 sec

### mechanics-20261027-1: Blizzard's 26 Oct racial pass (dev notes post 7) is not in the sim
- severity: high (moves every class's race table; the Human, Dwarf, Orc, Tauren and Skyborne rows all change)
- affects: shared racials (sim/core/racials_forever.go), every Level 60 preset's race table, Orc warlocks and hunters (pet damage)
- found: 2026-10-27, master d3df9247f
- evidence: tier 1, development notes post 7 (us.forums.blizzard.com/en/wow/t/2360696/7, Kaivax, 2026-10-26 20:35, edited 21:09), "Racial Abilities and Traits": Human "Ruthless now increases critical chance by 1% (was 2%)", "Human Spirit now increases Spirit by 10% (was 5%)"; Dwarf "Earth Fury (2% increased critical strike damage and healing) removed", "Added: Artisan's Precision - Increases hit chance with spells and attacks by 1%"; Tauren "Endurance no longer provides any modifiers beyond 5% increased Health (no hit chance or critical damage modifiers)"; Orc "Savage removed (1% increased critical strike chance)", "Command re-added, increasing Warlock and Hunter pet damage by 5%"; Skyborne "Wind-Blessed Haste bonus reduced to 2% (was 3%)". Master: racials_forever.go:27-29 (Human 2% crit, Spirit x1.05), :36 (Dwarf critDamageBonus 1.02), :63-64 (Orc 1% crit, no Command), :73-74 (Tauren Health x1.05 and critDamageBonus 1.02), :86-88 (Skyborne 1.03 haste). Client 1.60.1.70630 (2026-10-27) still has the old rows (20597 bp 2, 1259719, 20550 e1/e2 aura 163/50, 20574 bp 1, 1259710 bp 3), so the change is server side or in a later build. Size not measured (needs a code change): by the stat weights, about -1% for Human and Skyborne, -0.5 to -1% for Dwarf crit builds (partly back through 1% hit below the cap), -1% for Orc non-pet builds, Tauren crit builds -0.5% (estimates).
- proposed fix: in racials_forever.go under Forever: Human 1% melee and spell crit, Spirit x1.10; Dwarf drop critDamageBonus(1.02), add 1% melee and spell hit (Artisan's Precision); Tauren drop critDamageBonus(1.02); Orc drop the 1% crit, add Command (x1.05 pet damage for warlock and hunter pets); Skyborne haste 1.02. Then rerun the race tables. Check the next client build for the new rows (Artisan's Precision, Command).
- needs: orchestrator fix
- resolved: 2026-10-27, 500e3a1 (racial pass from the dev notes post 7).

### mechanics-20261025-1: Spellstone gives Fire spell damage too, but a server hotfix made it Shadow only
- severity: low
- affects: warlock (Affliction, Demonology, Destruction and their duty-curse builds, all on Spellstone)
- found: 2026-10-25, master 54878fa3d
- evidence: `sim/warlock/talents.go` applyWeaponImbue gives Major Spellstone 21 FirePower and 21 ShadowPower, citing the client rows' Misc 36 (Fire | Shadow). Hotfix push 112965 on 1.60.1.70532 (wago.tools/hotfixes, first seen 2026-10-23) rewrites the damage effects of 1237162 / 1237164 / 1237165 ("Increase Shadow Dam", SpellEffect 1233765 / 1233768 / 1233770) to EffectMiscValue_0 32, Shadow only; base points 14 / 17 / 21 and the haste effect are unchanged. Tier 1 (server data on the current build). Reproduce: `python3 -I tools/audit/mechanics/hotfixes.py $AUDIT_DB 70532 | grep "Increase Shadow"`. Measured with the patch below (seed 1, 20k iterations, 180 s, paired with the shared master run): Affliction 1035.28 -> 1033.78 (-1.5), Demonology 1065.09 -> 1063.73 (-1.4), Destruction 1011.77 -> 1008.98 (-2.8), duty-curse builds the same. Fire's share is Immolate (and Conflagrate for Destruction). Firestone's rows were not hotfixed. Sure of the data; the sim value is just the client row the hotfix replaced.
- proposed fix: under Forever add only ShadowPower for Spellstone (patch `$AUDIT_WORK/mechanics/spellstone.patch`): `if !warlock.Env.IsForever() { warlock.AddStat(stats.FirePower, r.damage) }`, and fix the comment's "Fire and Shadow damage". Spellstone should still beat Firestone for the Shadow builds, since Fire is only 10-20% of their damage (an estimate, not measured).
- needs: worker fix
- 2026-10-27 (pass 41): client 1.60.1.70630 now ships it: 1237162 / 1237164 / 1237165 EffectMiscValue_0 36 -> 32 in the build's own SpellEffect table, no hotfix needed. Still open on master d3df9247f. (tools/audit/mechanics/hotfixes.py is not in the repo; the 70630 table is the reproduction now: `python3 -I tools/sources/wago.py diff 1.60.1.70532 1.60.1.70630 --by-spell | grep "Increase Shadow"`.)
- resolved: 2026-10-27, d6fcdd3 (Spellstone Shadow only under Forever).

