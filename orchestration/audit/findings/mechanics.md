# Findings: Forever mechanics

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### mechanics-20261027-2: Amplify Curse no longer boosts Bane of Agony on client 70630
- severity: low (no preset takes Amplify Curse; rotopt's `amplifyCurse` knob defaults to 0)
- affects: warlock (only builds that take Amplify Curse)
- found: 2026-10-27, master d3df9247f
- evidence: tier 1, client 1.60.1.70630: 18288 e0 (aura 108 misc 8, +50%) mask 33792 -> 32768, which covers Curse of Weakness only (`AUDIT_DB=<70630> python3 tools/audit/mechanics/mask.py 18288`; on 70532 it listed Bane of Agony too). A new e1 (aura 107 misc 8, -20) covers Curse of Exhaustion again. sim/warlock/curses.go:74-77 still multiplies Bane of Agony's base damage by 1.5 and spends the aura.
- proposed fix: under Forever, Amplify Curse leaves Bane of Agony alone (drop the x1.5 and the deactivate in the Bane of Agony snapshot, or don't register the knob's cast), and update the forever-wiki talent text.
- needs: worker fix
- 2026-10-27: the 70630 mask reading holds, but 18288's 70630 description and Wowhead's Forever tooltip still name Bane of Agony, and every Agony rank gains a base-0 aura 271 effect (w2) that its new debuff text multiplies by, so the +50% looks moved onto Agony as a server-filled value, not removed. The sim keeps +50% (no preset casts Amplify Curse). needs: user ruling.
- 2026-10-27 (user): park it until a build takes Amplify Curse, then test in game (TASKS.md item of 2026-10-27). The sim keeps +50% of base meanwhile. needs: in-game test

## Resolved

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

