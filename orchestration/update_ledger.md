# Game-update ledger

Every change announced for the WoW Forever beta, from build 1.60.1.70009 (the first balance build) to 70425 (2026-10-16), checked against the client data and the sim. It also includes the reverse check: client changes between consecutive builds that no announcement mentioned. Written by the game-update auditor on 2026-10-16, against master 4f5c60833. The paladin rows also cover PR #54's head (718e65ba0), and the class rows cover each worker branch's head on 2026-10-16.

## How it was checked

- **Sources.** These are newest first. Every one was reachable except Wowhead.
  - Blizzard's "WoW Forever Beta Development Notes" thread, https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696. Its posts map to builds as follows:

    | Post | Date | Build |
    |---|---|---|
    | /6 | 2026-10-15 | 70425 |
    | /5 | 2026-10-08 | 70291 |
    | /4 | 2026-10-01 | 70170 |
    | /1 | 2026-09-24 | 70009 |

  - "Warrior Updates in Today's Beta Build", https://us.forums.blizzard.com/en/wow/t/warrior-updates-in-todays-beta-build/2369360 (2026-10-02, the 70170 build, folded into post /4).
  - The "Beta Client Update" (69977), "Auto-shoot Bug" and Known Issues posts.
  - The five class deep dives:
    - Hunter and Druid: https://worldofwarcraft.blizzard.com/en-us/news/24301515
    - Priest and Warrior: /news/24301514
    - Mage and Shaman: /news/24302097
    - Rogue and Warlock: /news/24310968
    - Paladin: /news/24302099
  - foreverchanges.pro/patch-notes, including its "Found in the game" lines.
  - **Not reachable:** Wowhead's Forever news (HTTP 403 from this environment's network policy). Its build articles retell the Blizzard posts above, so no announced change should be missing for that reason. Their tooltip diffs, however, were not used.
- **Client data.** wago.tools DB2 CSVs for every 1.60.1 beta build with a balance change: 69893 (beta open), 70009, 70170, 70291, 70334, 70338, 70374 and 70425. The tables: Spell, SpellName, SpellEffect, SpellAuraOptions, SpellCooldowns, SpellCategories, SpellCategory, SpellEquippedItems, SpellLevels, SpellMisc, SpellPower, SpellClassOptions, SpellDuration, SpellCastTimes, SpellProcsPerMinute, SpellInterrupts, SpellShapeshift, SpellTargetRestrictions, SpellAuraRestrictions, SpellReagents, SpellLabel, SkillLineAbility, TraitDefinition, TraitDefinitionEffectPoints, TraitNode, TraitNodeEntry, TraitNodeXTraitNodeEntry, TraitCond, TraitEdge, TraitTree, TraitNodeGroup, ItemSparse, ItemEffect, ItemXItemEffect, SpellItemEnchantment and ChrRaces. Each consecutive pair was diffed and grouped by class. The trait curve table was not downloaded, so a talent value changed only through its curve would not show.
- **Client-only builds.** 70334, 70338 and 70374 changed no class spell, talent or item stat that matters here. 69977, 70058, 70124, 70205, 70235 and 70245 are client-only per foreverchanges.
- **Sim status values:**
  - `landed`: the sim has it, read in the code with the value checked.
  - `missing`
  - `partial`
  - `wrong`: the sim has it with the wrong value or rule.
  - `not simulated on purpose`
  - `in flight (worker PR)`: a class worker's branch or PR has it.
  - `ambiguous (needs verify)`: a Need to Verify item.
  - `n/a`: healing, tanking, threat, PvP, leveling, UI.
- **Per-class sections.** Each was checked against master and the class worker's branch, then reviewed by me.

## Summary

| Class / system | Rows | Landed | In flight | Missing | Partial | Wrong | Not simulated on purpose | Ambiguous | n/a | Other |
|---|---|---|---|---|---|---|---|---|---|---|
| Warrior | 83 | 52 | 0 | 2 | 0 | 2 | 0 | 4 | 23 | 0 |
| Paladin | 65 | 37 | 5 | 2 | 0 | 2 | 1 | 2 | 14 | 2 superseded |
| Hunter | 82 | 44 | 2 | 7 | 1 | 1 | 3 | 1 | 22 | 1 info |
| Rogue | 68 | 38 | 0 | 0 | 2 | 2 | 1 | 3 | 22 | 0 |
| Priest | 54 | 20 | 14 | 2 | 0 | 2 | 1 | 1 | 14 | 0 |
| Shaman | 102 | 57 | 12 | 4 | 0 | 2 | 2 | 2 | 23 | 0 |
| Mage | 48 | 15 | 9 | 3 | 0 | 1 | 4 | 0 | 16 | 0 |
| Warlock | 86 | 47 | 4 | 0 | 0 | 1 | 3 | 1 | 29 | 1 covered by another row |
| Druid | 69 | 31 | 0 | 2 | 2 | 2 | 2 | 0 | 30 | 0 |
| Shared (races, mechanics, items) | 61 | 25 | 1 | 2 | 3 | 2 | 2 | 2 | 24 | 0 |
| **Total** | **718** | **366** | **47** | **24** | **8** | **17** | **19** | **16** | **217** | **4** |

The counts are as audited, before this audit's own PRs. A row that touches two scopes, such as Eureka!, Touch of the Grave or Grace of Air, appears under each class it affects.

### This audit's PRs (open, for the orchestrator to verify and merge)

| PR | Branch | Rows it closes | DPS at 180 s (20k iterations, seeds 1 and 1000003) |
|---|---|---|---|
| #61 | `claude/wow-forever-dps-presets-m9u15h-updates-druid` | DRU-70170-05 (Faerie Fire no longer resets the swing), DRU-70291-08 (Shifting Power not at full Energy), DRU-U-02 (Wolfshead Helm +5 Energy on Shifting Power) | feral-cat-duty-ff +1.9 / +2.1; the other druid presets 0 |
| #69 | `claude/wow-forever-dps-presets-m9u15h-updates-racials` | SHR-70170-01, SHR-U-06, WL-U-02, ROG-70170-04, ROG U-06 and G1 (Eureka! follows the client's cost, direct and periodic masks); SHR-70425-12, PRI-70425-20 (Touch of the Grave takes no Shadow modifiers) | Gnome warlocks −10.7 to −13.5 (−1.1 to −1.3%), destruction_fire −1.7; Undead shadow priests −3.6; other Undead presets −0.7 to −1.4 |
| #71 | `claude/wow-forever-dps-presets-m9u15h-updates-warrior` | WAR-70425-02 (Slam ignores pushback), WAR-U-01 (Bloodthrill's 5 s window), WAR-70425-04 (Weaponmaster's sword proc skips shield attacks) | Arms −0.9 to −1.2; Fury 2H +0.8 to +1.4; dual-wield Fury 0 |
| #73 | `claude/wow-forever-dps-presets-m9u15h-updates-rogue` | DD-18 (Seal Fate once per Mutilate), DD-13 (Cold Blood on both halves) | mutilate −2.0 to −2.4, mutilate-humanoid −2.1 to −2.5; others 0 |
| #77 | `claude/wow-forever-dps-presets-m9u15h-updates-items` | SHR-70425-24, SHA-70425-03 (Grace of Air no longer exclusive with Windfury Totem), MAGE-U-08 (Alcor's Sunrazor's Firebolt 94.5–115.5), SHR-U-01 Shadowstrike (no spell power) | 0 on every preset (no preset carries both totems or wears either weapon) |

### Missed by the class workers (routed to the orchestrator, not fixed here)

- **Paladin (#54).**
  - Seal of Righteousness still counts spell power twice: the 70425 notes say this is fixed, and the client set the proc's coefficient to 0. That's about −50 DPS on all 9 presets (PAL-70425-01, PAL-U-03).
  - Shield Specialization should give 4% mana, not 6% (`talents.go:218`).
  - Reckoning should store up to 4 charges and spend them on the next swing.
  - Doc hygiene: the wiki line "the sim takes all of it", and the `sor.go:45` comment.
- **Shaman.**
  - The Fire Nova nerf (70425) isn't on master or the worker branch: coefficient 0.214 → 0.143, rank 5 397–443 → 232–259 (`fire_totems.go:239-241`). About −44 DPS (−5%) on enhancement-level60 and imp-gw, about −25 on water-shield and mana-tide.
  - A Stormstrike charge is still spent by a spell that misses (`stormstrike.go:120-131`), about +2 DPS once fixed.
  - Water Shield should trigger on any spell crit (`water_shield.go:62`).
  - Flametongue Totem should use a shapeshifted druid's attack speed (`totem_weapon_buffs.go:108`).
  - The worker's Elemental presets spend 50 of 51 points and skip Elemental Mastery.
- **Mage.**
  - Fire Blast should always add a Heating Up charge, two on a crit, and Heating Up shouldn't be spent when Presence of Mind makes the Pyroblast instant (`applyHeatingUp`). About +1–4% for Fire and Frostfire.
  - None of the four mage presets has been ported to the 70425 tree:
    - The Arcane talent strings are shifted.
    - Fire Power is over its new cap of 3 (Fire 5, Frostfire 4).
    - Frost puts 5 points into Winter's Chill, which now only helps Cone of Cold.
    - Fire's meta still lists Wake of Fire.
  - notes.md items on Arcane Blast, Arcane Missiles, Missile Barrage and Ice Lance are stale.
- **Hunter.**
  - Sniper Shot is now covered by Mortal Shots and Efficiency (70425 client masks): +0.5–0.8% on the 3 MM and 3 Lone Wolf presets (`sniper_shot.go`, `talents.go:62-65, 319-323`).
  - Multi-Shot is instant in 70425 (unannounced; the sim has a 0.5 s cast).
  - Resourcefulness should trigger only on ability crits (`talents.go:397`; mana only).
  - Furious Howl isn't simulated. Its 40% cut reached the client in 70170, not 70009.
  - notes.md:114 needs a correction: the buff lasts 60 s on a 30 s cooldown (15 is its radius).
  - The `hunter.go:144-145` comment on Trueshot Aura is stale.
- **Priest.**
  - Power in Light now also counts Shadow Word: Pain on the target (`penance.go:103-112`), +0.3–1% on priest__smite.
  - Vampiric Embrace can no longer be resisted (`vampiric_embrace.go:62`, no preset).
  - The branch's meta files still say Devouring Plague has a 1 min cooldown.
- **Warlock.** Nothing missed. 70425 deleted Demonic Embrace's Spirit effect, so the Need to Verify item at notes.md:172 can be closed. The branch's metas still carry the old published DPS.

### Rulings needed

1. **Grace of Air in the melee raid.** 70425 lets Grace of Air stack with Windfury Totem, and both are raid-wide. Should the reference raid have a second shaman's Grace of Air? That's about +4–5% for rogues, +2–3% for warriors, feral and Survival, +1–2% for Ret. PR #77 also reverses the user's "for now" ruling of 2026-10-08 that they don't stack.
2. **Trueshot Aura** is raid-wide in 70425, but the BM, Lone Wolf and Survival presets don't get it. Is there a Marksmanship hunter in the reference raid? About +2–3% for those hunters.
3. **Paladin Seal of Righteousness.** PR #54 keeps the double spell power behind a `pal_sor_sp_double` assumption, but the 70425 notes and client say it's fixed. Make "Once" the expected setting?
4. **Warlock races.** After #69, Gnome may no longer be the best warlock race. Rerun the race tables.

### Report assumptions and Need to Verify items this ledger settles

For the assumptions audit (orchestration/assumptions_audit.md). Keys are from tools/dps_report/assumptions_def.py.

**Settled:**
- `warr_crit_rage`: settled at the expected setting, +100%. The client's hidden "Rule of Rage (DND)" (1322574) went from 10 to 100 base points in 70291. That matches the second 10-01 warrior pass, and the 10-08 notes give bears the same 100%. Rows WAR-70170-02 and WAR-U-05; the sim has it (`warrior.go:179`).
- `pal_sor_sp_double` (PR #54's branch, not on master): the client and the 10-15 notes point to "Once". Seal of Righteousness's proc spells 25713 and 25735-25742 have coefficient 0.1 → 0 in 70425, and the notes say it "gained almost twice as much benefit from Spell Damage than intended" and that this is fixed. Row PAL-70425-01; a ruling is still needed, so it's in the rulings list above.
- Consecration's spell power coefficient is 0.042 a tick (70425 client and notes; PR #54). Row PAL-70425-03.

**Not settled by any build from 69893 to 70425** (no client value change, or server side):
- Ice Lance's coefficient: 70425 only changed the frozen bonus, 300 → 200.
- Bane of Doom's coefficient.
- The top ranks of Smite, Frostbolt, Lightning Bolt, Wrath, Rip and Raptor Strike: only ranks 1-5 were retuned, in 70009 and 70291.
- Earth Shock rank 7: 70425 added a combat-text attribute only.
- Holy Strike's flat bonus.
- Trueshot Aura rank 5 and Hunter's Mark rank 4.
- Shred's flat bonus: only rank 1 changed, 24 → 22.
- Warlock pet scaling and demon melee: Warlock Pet Scaling 416189 changed only its health (70291) and armor (70425) effects.
- Hunter pet attack power and hit: Hunter Pet Scaling 415429 changed only its health and armor effects, and every value is 0 in the client. 70425's pet lines are covered by HUN-70425-03 and HUN-70425-06.
- The Maelstrom Weapon and Omen of Clarity rates.
- Rend's attack power: 11574 has no coefficient in any build.
- `wf_totem_offhand` and the Windfury Totem ICD: 70425 only made the totem raid-wide and players-only.

## Per-class ledgers

Each section below has five parts:

1. Announced changes.
2. Unannounced client changes (the reverse check).
3. Proposed fixes.
4. Need to Verify candidates. These are also copied to notes.md, "Game-update audit (2026-10-16)".
5. Counts.

`S/db2/<Table>.<build>.csv` and `S/diffs/<old>-<new>.md` in the sections below are the auditor's working copies. To rebuild them:

1. Run `orchestration/tools/wago_db2_fetch.sh <dir>` to download the tables.
2. Run `python3 -I orchestration/tools/db2diff.py <dir> <old> <new> <out.md>` for each pair of consecutive builds.

Row IDs: `<CLASS>-<build>-NN` for an announced line, `<CLASS>-U-NN` for an unannounced client change, `DD-NN` for a deep-dive statement. Sources: DN or P with a number is a dev-notes post, DD a deep dive, FC foreverchanges.

### Warrior: WoW Forever beta update audit (builds 70009 to 70425)

Scope: every warrior spec (sim/warrior, presets tools/dps_report/req/warrior__*.json: Arms, Arms vs Dragonkin, Fury, Fury 2H and the three Battle Shout duty presets) plus the warrior lines of the general Classes/Rage sections. Checkout origin/master 6a87c9c. Warrior has no 70425 worker branch, so nothing here is "in flight".

Presets checked against: Arms `20305213032515201-55050000001` (Orc, Improved Slam 2/2, Bloodthrill 5/5, Weaponmaster 5/5, Booming Voice 5/5), Arms Dragonkin `20305213132515201-3505-2` (Spearing Strike 1/1), Fury `30305013002-25053005112110501-2` (Undead, DWS 5/5, Raging Blows, Furious Precision 3/3, Flurry 5/5, Booming Voice 2/5), Fury 2H `30305013032-25053000152110501` (Orc). All three have raid hits (2 a minute of 1500, PR #55) and the Fury 2H and Arms rotations cast Slam.

Key finding on dating: the 1 October "second warrior pass" (post 4, the 10-02 Warrior Updates thread) is mostly not in the 70170 client. 70170 holds only the DWS hit move, Unbridled Wrath's ProcChance, Spearing Strike's stance and the Rend/Sunder tap flag. Bloodthirst 45%, Booming Voice's cost, Raging Blows' -3, Furious Precision, Gore Drinker, Lingering Rage, the tree moves and the 100% crit Rage marker all first appear in the 70291 client (server hotfix on 10-02, client data followed).

Links: P1 = https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/1 (2026-09-24, 70009); P4 = .../2360696/4 (2026-10-01/02, 70170; the 10-02 thread https://us.forums.blizzard.com/en/wow/t/warrior-updates-in-todays-beta-build/2369360 is the same build); P5 = .../2360696/5 (2026-10-08, 70291); P6 = .../2360696/6 (2026-10-15, 70425); DD = Priest & Warrior deep dive https://worldofwarcraft.blizzard.com/en-us/news/24301514 (2026-09-30); FC = foreverchanges.pro/patch-notes.

#### 1. Announced changes

##### Build 70425 (P6, 2026-10-15)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WAR-70425-01 | P6, 10-15, 70425 | Battle Shout is now raid wide | minor | SpellEffect 683198/684488/685066/688665/689577/689517/702382 (all 7 ranks): ImplicitTarget_0 20 -> 56. Value unchanged (rank 6 11551: 111 + 0.6/level to 61, so 115 at 60) | landed | Wording only (PR #51, orchestration/standards.md duties). A single-player sim is unaffected: the warrior's own shout uses `NewPartyAuraArray` (sim/warrior/shouts.go:52), the raid buff is `core.BattleShoutAura` (sim/core/buffs.go:1641, 1633 `foreverBattleShoutAP`, 3 min at :1656). |
| WAR-70425-02 | P6, 10-15, 70425 | Slam is no longer subject to spell pushback | yes | SpellMisc Attributes_6 + 0x8000 (bit 15, NO_PUSHBACK) on Slam 1464, 8820, 11604, 11605, 1240193, 462893-462897 (33554432 -> 33587200) and 1310196-1310200 (0 -> 32768) | **missing** | sim/warrior/slam.go:33 has no `PushbackReduction`, so `core.applySpellPushback` (sim/core/cast.go:115, `sim.Roll(0,1) > hcSpell.PushbackReduction`) pushes Slam back on every PR #55 raid hit (a ProcMaskSpellDamage Fire hit, ProcMaskDirect) that lands during the cast, and also re-stops the swing (`StopMeleeUntil`, cast.go:152) even with Improved Slam. |
| WAR-70425-03 | P6, 10-15, 70425 | Overpower can activate against much lower level enemies again | n/a | server side | n/a | Leveling only; the boss is higher level. |
| WAR-70425-04 | P6, 10-15, 70425 | Weaponmaster's Sword Specialization can no longer trigger from Shield Bash and Shield Slam | n/a (tank) | server side (1290261 unchanged) | missing | sim/warrior/talents.go:118-134 rolls the sword extra attack on every landed MH/OH sword spell; Shield Slam is `ProcMaskMeleeMHSpecial` (shield_slam.go:26). Only a sword-and-board Protection sim is affected; no DPS preset. Fix: skip `SpellCode_WarriorShieldSlam` in the trigger. |
| WAR-70425-05 | P6, 10-15, 70425 | Lingering Rage no longer stops health regen after combat | n/a | 1323965 EffectAura 94 -> 110, bp 0 -> -100 | n/a | Out of combat; the talent has no sim effect (proto field only). |
| WAR-70425-06 | P6, 10-15, 70425 | Improved Berserker Rage removes movement impairing effects again | n/a | 23690 tooltip only | n/a | Movement. |
| WAR-70425-07 | P6, 10-15, 70425 | Racials on warrior presets: Orc Savage 1% crit, Blood Fury 3 min; Undead Touch of the Grave 20% for warriors with a 25 s ICD; Human Ruthless, Dwarf Earth Fury, Tauren Endurance, Gnome Eureka 8 s, Troll Regeneration | yes | 1260189 (Paladin/Rogue/Warrior): ProcChance 5 -> 20, ProcCategoryRecovery 1000 -> 25000; Eureka 1259813 charges removed, DurationIndex 8 -> 31 (8 s) | landed | Shared port PR #50/#53: sim/core/racials_forever.go:58 (Savage), :122/:166 (Blood Fury 3 min), :342 and :371 (Touch of the Grave 0.20, 25 s). Detailed in the shared/racial ledger. |

##### Build 70291 (P5, 2026-10-08)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WAR-70291-01 | P5 Classes, 10-08, 70291 | Rage from damage taken uses an expected creature's health, not the player's | yes | server side | landed | PR #55: `core.ForeverDamageTakenRage` (sim/core/rage.go:88-99), used at rage.go:196: 20 x (1 - a) x damage / `NormalMobStatsAt(level).Health`, 4.51 Rage for a 1500 hit at 60. Tested in dps_warrior/forever_mechanics_test.go:146 and raid_hits_rage_test.go. Constant 20 is a community fit (notes.md Warrior item 3). |
| WAR-70291-02 | P5 Classes | Rage calc now handles crit and crushing multipliers | yes | server side | landed | rage.go:92-96 uses `result.Damage` (crit/crush included), dividing out only the resistance multiplier. |
| WAR-70291-03 | P5 Classes | Absorbs ignored for Rage | n/a | server side | n/a | The sim models no absorbs on a DPS warrior (no absorb code in sim/core). |
| WAR-70291-04 | P5 Classes (dev note) | Rage balanced for 20-40% armor by level instead of 50% | yes | server side | landed | rage.go:97 `assumedArmor := 0.2 + 0.2*(level-1)/59` (40% at 60). See U-14 for Berserker Rage, which kept the Classic formula. |
| WAR-70291-05 | P5 Warrior | Demoralizing Shout now generates threat | n/a | SpellEffect 1361983-1361987 added (Effect 63, bp 11/19/27/35/43, 0.8/level) on 1160...11556 | n/a | Threat. |
| WAR-70291-06 | P5 Arms | Deep Wounds is again required for Impale | minor | TraitEdge Deep Wounds -> Impale (tree 1117) | landed | ui/core/talents/trees/warrior.json: impale prereq (2,2). |
| WAR-70291-07 | P5 Arms | Deep Wounds no longer scales with off-hand AP; does not scale with AP at all; rolling DoT | yes | server side (412609: 4 ticks of 3 s, 12 s; 412613 unchanged) | landed | sim/warrior/deep_wounds.go:80 `ap := TernaryFloat64(forever, 0, ...)` for both hands, :94-96 adds the unticked damage of the old DoT to the new one. Tests forever_mechanics_test.go:211. See NtV-3 for caster modifiers. |
| WAR-70291-08 | P5 Protection | Last Stand grants the right health | n/a | 12976 EffectAura 230 -> 34 | n/a | Tank (talents.go:490, 30% max health). |

##### Build 70170 (P4 and the 10-02 Warrior Updates thread; client data mostly in 70291)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WAR-70170-01 | P4, 10-01, 70170 | 75% more Rage on a critical basic attack | yes | server side; superseded by -02 | landed | Superseded by WAR-70170-02. |
| WAR-70170-02 | P4 (10-02 second pass) | Crit Rage increased to 100% (was 75%) | yes | Hidden "Rule of Rage (DND)" 1322574 added in 70170 (aura 4, bp 10), bp 10 -> 100 in 70291; 1313291 renamed Dual Wield Specialization -> Rule of Rage (DND). Sim should use x2 | landed | sim/warrior/warrior.go:179 `AddCritDealtRageMultiplier(2)`, applied to crit white hits MH and OH at rage.go:180. Abilities give no Rage (HS/Cleave replace the swing). |
| WAR-70170-03 | P4 | Rend and Sunder Armor tap enemies instantly | n/a | Attributes_6 + 0x800000 (bit 23, TAPS_IMMEDIATELY) on Rend 772...11574 and Sunder 7386...11597 | n/a | Tapping. |
| WAR-70170-04 | P4 | Queueing Heroic Strike no longer raises off-hand hit | yes | server side | landed | sim/warrior/heroic_strike_cleave.go:116 `DisableDWMissPenalty = !IsForever()`. |
| WAR-70170-05 | P4 + Known Issues | DWS no longer grants its hit to both weapons (bug fix) | yes | 23584: 70170 moved the hit to eff 1 (Aura 54, bp 10); 70291 removed it | landed | Superseded by -18: DWS gives no hit at all (talents.go:203-218). |
| WAR-70170-06 | P4 Arms | Spearing Strike no longer needs a two-hander; needs Battle Stance | minor | 1310222: SpellShapeshift 58595 added (ShapeshiftMask 65536 = Battle Stance), EquippedItemSubclass 1378 -> 173555; 20 s CD, 15 Rage, eff1 40% weapon, eff2 x2 | landed | sim/warrior/spearing_strike.go:18 `RegisterSpell(BattleStance, ...)`, no weapon check, 0.4 (+0.8 vs Giant/Dragonkin) normalized MH (:47-51). Used by warrior__arms_dragonkin. |
| WAR-70170-07 | P4 Fury | DWS no longer gives 20/40/60/80/100% more off-hand Rage | yes | 23584 in 70291: eff 1 back to Aura 4 (dummy, server side), bp 5, trait curve 111755 (curve values not downloadable). Aidan Moon (X) and FC: 10-50% now | landed | talents.go:211 `AddOffHandDealtRageMultiplier(1 + 0.1*points)` (x1.5 at 5/5). See NtV-1. |
| WAR-70170-08 | P4 Fury | Unbridled Rage (Wrath) proc chance fixed (too low) | yes | 12322 SpellAuraOptions ProcChance 60 -> 100 (70170); talent curve 83299 | landed | talents.go:178 12% a point, white hits only (proc mask 4), 1 Rage (spell 12964 bp 10). |
| WAR-70170-09 | P4 Fury | Booming Voice: shouts' area +10-50% and Rage cost -5-25% | yes (duty presets) | 12321 eff 1 added (Aura 108, mv 14 cost, bp -25, curve 125261) in 70291; mask 196608 -> 983040 (all shouts) | landed | sim/warrior/shouts.go:62-63 `10 * (1 - 0.05*BoomingVoice)` (Fury 2/5 = 9 Rage, Arms 5/5 = 7.5). Area n/a. |
| WAR-70170-10 | P4 Fury | Unbridled Wrath no longer gives 2 Rage with a two-hander | yes | 12322 tooltip (70291); 12964 bp 10 (1 Rage) | landed | talents.go:179 `rageGain := 1.0` for any weapon. |
| WAR-70170-11 | P4 Fury | New Lingering Rage (row 2): Rage decays 2-10 s later out of combat | n/a | 1323964/5/6 added (70291), curve 125260 | n/a | Out of combat; proto field `lingering_rage` only. |
| WAR-70170-12 | P4 Fury | Blood Craze no longer triggers from Bloodthirst | n/a | 16487 ProcTypeMask 139960 -> 139944 (70291) | n/a | Healing, not simulated. |
| WAR-70170-13 | P4 Fury | Improved Cleave removed | yes | trait node removed (70291) | landed | Not in proto/tree; Cleave 20 - 3 with Raging Blows (heroic_strike_cleave.go:57). |
| WAR-70170-14 | P4 Fury | Boundless Rage removed | minor | 1310236 removed (70291) | landed | Not in proto; max Rage 100. |
| WAR-70170-15 | P4 Fury | New Furious Precision (row 3): off-hand hit +4/7/10% | yes | 1323963 (Aura 54, bp 10, curve 125259), replaced Boundless Rage's trait def 135484 (70291) | landed | talents.go:226 `{0,4,7,10}` on OH spells. |
| WAR-70170-16 | P4 Fury | Raging Blows: -3 Rage on Cleave and Whirlwind | yes | 1310315 eff 1 bp -20 -> -30, SpellClassMask_1 0 -> 4 (Whirlwind) (70291) | landed | whirlwind.go:29 (25 - 3), heroic_strike_cleave.go:57 (20 - 3). |
| WAR-70170-17 | P4 Fury | Whirlwind strikes with both weapons baseline | yes | server side (1680 has only a normalized MH effect; tooltip "both melee weapons" in 70291) | landed | whirlwind.go:16 off-hand spell when Forever and dual wielding, normalized OH damage x OH multiplier and DWS. |
| WAR-70170-18 | P4 Fury | DWS no longer grants off-hand hit | yes | 23584 Aura 54 effect gone (70291) | landed | talents.go:203-218 gives damage and Rage only. |
| WAR-70170-19 | P4 Fury | Enrage no longer required for Flurry; Death Wish is | minor | TraitEdge changes (70291) | landed | warrior.json: flurry prereq (4,1) deathWish. |
| WAR-70170-20 | P4 Fury | New Gore Drinker (row 6, after Enrage): next 3 melee attacks heal 0.5/1% | n/a | 1323967/8/9 added (70291) | n/a | Healing; proto field only. |
| WAR-70170-21 | P4 Fury | Improved Berserker Rage moved to row 5 | minor | TraitNode PosY 5130 -> 4530 | landed | warrior.json row 4 (0-based). |
| WAR-70170-22 | P4 Fury | Berserker Rage at level 30 (was 32) | n/a | 18499 SpellLevels 32 -> 30 (70291) | landed | berserker_rage.go:11. |
| WAR-70170-23 | P4 Fury | Bloodthirst AP ratio 45% (was 35%) | yes | 23881/23892/23893/23894 eff 1 bp 35 -> 45 (70291); eff 0 flat 30/37/43/48 | landed | bloodthirst.go:17-21 `0.45*AP + 48` at 60. |
| WAR-70170-24 | P4 Protection | Iron Will to Prot row 1; Anticipation row 2; Toughness removed; Improved Bloodrage row 1; Improved Revenge row 2; Improved Disarm row 3; Improved Shield Bash row 4 | n/a | TraitNode moves, Toughness 12299 node removed (70291) | landed | warrior.json Protection rows match. |

##### Build 70009 (P1, 2026-09-24)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WAR-70009-01 | P1, 09-24, 70009 | Sunder Armor threat values corrected | n/a | Sunder eff 1 (Effect 63) e.g. 11597 1013 -> 206, 7386 1 -> 34 | n/a | Threat. |
| WAR-70009-02 | P1 Protection | Bastion (row 5) and Focused Rage (row 6) swapped | n/a | TraitNode moves (70291 data: Focused Rage 4530, Bastion 5130) | landed | warrior.json: focusedRage row 4, bastion row 5. |
| WAR-70009-03 | P1 Arms | Bloodthrill 4/8/12/16/20% (was 2-10%) | yes | curve 109366 (values server/curve side, not downloadable; spell 1289682 bp 15 is overridden by the curve) | landed | talents.go:144 `0.04*points`. |
| WAR-70009-04 | P1 Arms | Bloodthrill only from main-hand melee attacks, including Cleave and Heroic Strike | yes | 1289682: EffectAura 4 -> 42 (proc trigger), trigger 1282733, ProcTypeMask 4 -> 20 (+ mask_1 32) | landed | talents.go:157 `ProcMaskMeleeMH` (HS/Cleave are MHSpecial, heroic_strike_cleave.go:19/65), needs Rend on target. Duration: see U-01. |
| WAR-70009-05 | P1 Arms | Slam cooldown 18 s (was 15) | yes | SpellCooldowns CategoryRecoveryTime 15000 -> 18000 on all 10 Slam rows | landed | slam.go:26. |
| WAR-70009-06 | P1 Arms | Improved Slam reduces Slam's cooldown by 1.5/3 s | yes | 70009: eff 2 (Aura 107, mv 11, bp -3000) with curve 124012 (1.5/3); **70291 removed the curve row**, so -3 s at either rank | landed | slam.go:26 3 s at 1 or 2 points (presets take 2/2). See NtV-2. |
| WAR-70009-07 | P1 Races | Gnome Eureka is 10% on Rage abilities | n/a | warrior variant 1259813 eff 0 -40 -> -10 | landed | Shared racials (PR #50); no Gnome warrior preset. |

##### Deep dive (DD, 2026-09-30; describes the design around 70009-70170)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WAR-DD-01 | DD Base | Rage from taking damage: original formula ignoring armor (as if 50%), absorbs ignored | yes | server side | n/a | Superseded by WAR-70291-01..04. |
| WAR-DD-02 | DD Base | Rage from dealing damage normalized to weapon speed; crit 75% more | yes | server side | landed | rage.go:73-79 k x speed (3.46 1H, 4.5 2H at 30, x1.5 at 60, assumptions warr_rage_k/oh); crit superseded by WAR-70170-02. |
| WAR-DD-03 | DD Base/Arms | Tactical Mastery at 14 keeps 10 Rage; Improved Tactical Mastery +3 a point to 25 | yes | 1310185 bp 10; 12295 bp 15 (curve 83313) | landed | stances.go:41-44. |
| WAR-DD-04 | DD Base | Victory Rush at 20 | n/a | 402975 tooltip only | n/a | Needs kills; not simulated on purpose. |
| WAR-DD-05 | DD Base | Recklessness, Retaliation, Shield Wall no longer share a cooldown | minor | 1719: 15 s, 30 min CD, +100% crit (aura 290), +20% damage taken | landed | recklessness.go:21-42 own timer. Header comment (:10) is a stale SoD line ("50%, 12 s, 5 min"). |
| WAR-DD-06 | DD Base | Rend gains significant AP scaling | yes | none: 11574 eff 0 bp 21, no AP coefficient, unchanged through 70425 | ambiguous (needs verify) | rend.go:12 `ForeverRendAPPerTick = 0.02` (assumption warr_rend_ap). Already in notes.md Warrior. |
| WAR-DD-07 | DD Base | Thunder Clap AP scaling, usable in Defensive, 6 s CD | n/a | server side | n/a | Tank. |
| WAR-DD-08 | DD Base | Battle Shout AP reduced, Improved Battle Shout baked in | yes | 25289 bp 139 (+0.6/level), 11551 bp 111 (+0.6 to 61), 3 min | landed | buffs.go:1633-1656: rank 6 = floor(111 + 0.6 x 8) = 115 at 60, talent ignored. |
| WAR-DD-09 | DD Base | Sunder usable over a stronger Expose Armor; threat scales with AP | minor | see U-10 | ambiguous (needs verify) | Shared debuff code (sim/core/debuffs.go:155). |
| WAR-DD-10 | DD Arms | Improved Overpower moved to row 2 | minor | TraitNode | landed | warrior.json (1,3). |
| WAR-DD-11 | DD Arms | Spearing Strike: 15 Rage, 20 s CD, double vs mounted/Dragon/Giant | minor | 1310222 eff1 40%, eff2 x2 | landed | spearing_strike.go (40% + 80%). |
| WAR-DD-12 | DD Arms | Weaponmaster: axes/polearms crit, maces/staves armor ignore, swords extra attack | yes | 1290261 bp 5/15/5 (curves 83305/100832/100831) | landed | talents.go:70-110: 1% crit, 3% armor, 1% extra attack a point (200 ms ICD, server side). |
| WAR-DD-13 | DD Arms | Improved Slam: 2 points, cuts CD, GCD and cast; no swing delay | yes | 12862 eff 0 cast -500, eff 1 GCD -500 (curves 109365/109364), eff 2 CD -3000 | landed | slam.go:16 250 ms a point off cast and GCD; :50-53 swing stopped only without the talent. |
| WAR-DD-14 | DD Fury | Booming Voice no longer changes shout duration (baseline) | yes | 25289 duration 180000 | landed | buffs.go:1656 3 min. |
| WAR-DD-15 | DD Fury | Unbridled Rage chance up, 2 Rage with a 2H | yes | superseded by WAR-70170-08/10 | n/a | Superseded. |
| WAR-DD-16 | DD Fury | Improved Cleave reduces Cleave's cost; Boundless Rage; DWS also gives OH hit; Precision (row 5) | yes | all removed by 70291 | n/a | Superseded by WAR-70170-13/14/18 and U-04. |
| WAR-DD-17 | DD Fury | Raging Blows: Whirlwind with both weapons, cheaper Cleave | yes | superseded | n/a | Superseded by WAR-70170-16/17. |
| WAR-DD-18 | DD Fury | Enrage triggers from any damage taken | yes | 12317 ProcChance 30, ProcTypeMask 139944; 12880 12 s, curve 2-10% Physical | landed | talents.go:238-271 (30%, 2% a point, 12 s), fed by raid hits since PR #55. |
| WAR-DD-19 | DD Fury | Death Wish increases damage taken instead of lowering armor | yes | 12328 eff 0 Aura 79 +20% Physical, eff 2 Aura 87 +5% taken, 30 s, 3 min | landed | talents.go:438-439. |
| WAR-DD-20 | DD Fury | Flurry total attack speed reduced by 5% | yes | 12966 15 s, 3 charges; rank curve 83289 | landed | talents.go:309 1.05..1.25; 15 s at :315. |
| WAR-DD-21 | DD Fury | Bloodthirst: no heal, 10% speed, base value plus less AP | yes | 23894 eff 0 48, eff 1 45% (after 70291), eff 2 Aura 31 speed | landed | bloodthirst.go (speed n/a). |
| WAR-DD-22 | DD Fury | Improved Demoralizing/Battle Shout baseline | minor | see DD-08 | landed | Demo Shout is a tank tool. |
| WAR-DD-23 | DD Protection | Shield Spec 5 Rage, Anticipation 2x, Improved Bloodrage all Rage, Master of Defense, Improved Revenge damage, Defiance needs shield, Improved Disarm/Shield Wall CD, Vanguard, Bastion, Focused Rage, Shield Slam block scaling | n/a (Focused Rage minor) | trait curves unchanged since 70338 | landed | talents.go:351-421, bloodrage.go:14. Shield Slam/Revenge base values stale (notes.md, tank only). |

#### 2. Unannounced client changes (reverse check)

Client-only builds: 70291 -> 70334 (only Summon Azurky), 70334 -> 70338 and 70338 -> 70374 (no spell rows): **no warrior changes, confirmed**. Talent tree 1117 is identical between 70374 and 70425.

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WAR-U-01 | DB2 69893 -> 70009 | Bloodthrill now procs Overpower's own enabler 1282733 (5 s) instead of the 6 s 1289681; 1282733 lost CumulativeAura 3 (one charge). 70425 tooltip drops "Lasts $1289681d" | yes (Arms) | 1289682 eff 0 trigger 0 -> 1282733; 1282733 DurationIndex 5000 ms; 1289681 (6000 ms) is no longer referenced | **wrong** (minor) | talents.go:149 `Duration: time.Second * 6` -> 5 s. |
| WAR-U-02 | DB2 70170 -> 70291 | Improved Slam's cooldown curve removed: -3 s flat at rank 1 and 2. FC reads "3 sec per rank" | yes | TraitDefinitionEffectPoints 25167 (eff 2, curve 124012) REMOVED; 12862 eff 2 bp -3000 | landed | slam.go:26. NtV-2. |
| WAR-U-03 | DB2 70291 (FC "found in the game") | Bloodthirst no longer needs Death Wish; Last Stand no longer needs Improved Bloodrage; Precision (1225295) removed | minor | TraitEdge 136314/124811 removed; node 105929 removed | landed | warrior.json: bloodthirst and lastStand have no prereq; no Precision in proto. |
| WAR-U-04 | DB2 70009 -> 70170 -> 70291 | DWS: 70170 dropped the off-hand Rage effect (eff 1 became hit), 70291 restored it as a dummy with curve 111755 | yes (Fury) | 23584 eff 1 Aura 4/bp 5 -> Aura 54/bp 10 -> Aura 4/bp 5; curve 111755 -> 112760 -> 111755 | landed | talents.go:211 (10% a point, Aidan Moon's 50%). NtV-1. |
| WAR-U-05 | DB2 70170/70291 | Hidden "Rule of Rage (DND)" 1322574 (bp 10 -> 100), the crit Rage marker | yes | as WAR-70170-02 | landed | warrior.go:179. |
| WAR-U-06 | DB2 70374 -> 70425 | Demoralizing Shout taps immediately (Attributes_6 bit 23) | n/a | 1160...11556 Attributes_6 0 -> 8388608 | n/a | Tapping. |
| WAR-U-07 | DB2 70374 -> 70425 | Attributes_6 bit 21 (combat text for mechanics) on ~40 warrior spells: Charge, Hamstring, Pummel, Shield Bash, Disarm, Intimidating Shout, Piercing Howl, Rend, Deep Wound(s), Concussion Blow | n/a | Attributes_6 + 2097152 | n/a | Display only. |
| WAR-U-08 | DB2 70170 -> 70291, 70425 | Sunder Armor's armor aura type 22 -> 674 (players 7386...11597 in 70291, NPC Sunders in 70425); armor values unchanged | minor | SpellEffect 680100...688831 EffectAura 22 -> 674 | ambiguous (needs verify) | Probably the new Sunder/Expose Armor coexistence (WAR-DD-09). Shared debuffs: sim/core/debuffs.go:140-155. |
| WAR-U-09 | DB2 70170 | Defensive Stance gains an effect (Aura 668) | n/a | SpellEffect 1357990 added on 71 | n/a | Tank; aura type unknown. |
| WAR-U-10 | DB2 70170 -> 70291 | Spearing Strike tooltip drops then restores "Mounted targets are dismounted"; Bloodrage tooltip drops "considered in combat" | n/a | Spell text | n/a | FC "found in the game". |
| WAR-U-11 | DB2 70291 | Berserker Rage, Bloodrage, Death Wish, Enrage class masks (Gore Drinker hookup); Iron Will curve 125265 | n/a | SpellClassOptions masks | n/a | Gore Drinker/Iron Will not simulated. |
| WAR-U-12 | DB2 69893 -> 70009 | Hamstring Rage Reduction 22778 and Improved Hamstring 24428 EffectBonusCoefficient 1 -> 0; Deep Wound 412613 Attributes_10 0 -> 2; Battle Shout 25101 (NPC/item version, 15 min, no cost) bp 400 -> 200 | n/a | as listed | n/a | No player-DPS effect (25101 is no warrior skill line ability). |
| WAR-U-13 | DB2 baseline (69893, unchanged) | Deep Wound damage spell 412613 has Attributes_6 bit 29 (IGNORE_CASTER_DAMAGE_MODIFIERS / NO_DONE_PCT_DAMAGE_MODS) | yes | 412613 Attributes_6 536870912 (69893) -> 538968064 (70425, + bit 21 only) | ambiguous (needs verify) | deep_wounds.go:44 re-applies `AttackerDamageMultiplier` on every tick ("Double dips"), and :85/:89 already fold it into the snapshot. NtV-3. |
| WAR-U-14 | Consequence of WAR-70291-01..04 (server side) | Berserker Rage's "extra rage when taking damage" still uses Classic's formula | minor (no preset casts Berserker Rage) | 18499 tooltip unchanged; amount server side | **wrong** (minor) | berserker_rage.go:33-38: a spell hit gives `RawDamage*2.5/GetRageConversion(63)` = 15.4 extra Rage for a 1500 raid hit, 3.4x the 4.5 base Forever Rage. NtV-4. |

#### 3. Proposed fixes

1. **WAR-70425-02 Slam pushback (missing).** sim/warrior/slam.go `registerSlamSpell`: add `PushbackReduction: core.TernaryFloat64(warrior.Env.IsForever(), 1, 0)` to the SpellConfig (cast.go:115 rolls `> PushbackReduction`, so 1 never pushes back, and the swing re-stop at cast.go:152 goes with it). Test in dps_warrior/forever_mechanics_test.go: a level 60 warrior with Improved Slam 2/2, raid hits of 1500 at a high rate, check every Slam's cast lasts exactly 1.0 s (`SpellMetrics TotalCastTime / Casts`) and the MH swing is never delayed. Impact: tiny, about 12 Slams a fight x 2 hits/min x 1.0 s cast = ~0.4 pushbacks per fight, well under 0.1% DPS. Presets: warrior__arms*, warrior__fury_2h* (and their duty presets); Fury 1H does not cast Slam.
2. **WAR-U-01 Bloodthrill window (wrong).** sim/warrior/talents.go:149 `time.Second * 6` -> `time.Second * 5` (1282733's 5000 ms), and comment the client source. Test: proc Bloodthrill, check Overpower is castable at 4.9 s and not at 5.1 s. Impact: near zero (Overpower is used within a GCD or two), maybe -0.1 DPS. Presets: warrior__arms, warrior__arms-duty-battle-shout, warrior__arms_dragonkin.
3. **WAR-U-14 Berserker Rage (wrong, no preset).** sim/warrior/berserker_rage.go:33-38: under Forever give the bonus as a multiple of `core.ForeverDamageTakenRage(result, level)` (Classic's Berserker Rage roughly doubled damage-taken Rage, so 1x on top is the natural choice), not the Classic conversion; keep the melee 2.0 flat or route it the same way. Test: Berserker Rage up, one 1500 spell hit at 60 gives 2 x 4.51 Rage. Impact: none on current presets (no rotation casts 18499); it would otherwise overrate Berserker Rage by ~10 Rage a minute with raid hits on in a rotation search. Presets: none (Fury/Fury 2H take Improved Berserker Rage 1/2 but never cast it).
4. **WAR-70425-04 Weaponmaster vs Shield Slam (missing, tank only).** talents.go:118 trigger: return when `spell.SpellCode == SpellCode_WarriorShieldSlam` (Shield Bash is not in the sim). No DPS preset.
5. Wording only: rage.go:243-244 says "Rage from taking damage stays Classic's" (false since PR #55); recklessness.go:10 carries a stale SoD comment. No DPS impact.

DPS-impact rows still open for verification (not code fixes yet): NtV-1 (DWS Rage, Fury), NtV-3 (Deep Wounds caster modifiers, Arms and Fury, roughly 0.3-0.6% DPS if the sim double counts), WAR-DD-06 (Rend AP, existing assumption).

#### 4. Need to Verify candidates

* **NtV-1 Dual Wield Specialization's off-hand Rage.** The 1 October notes say DWS "no longer provides a 20/40/60/80/100% increase to your Off-Hand weapon's Rage generation". The 70170 client did drop the effect, but 70291 put it back as a server-side dummy (23584 eff 1, Aura 4, bp 5) on the same trait curve (111755) it used when it gave 20-100%. Aidan Moon (X) and ForeverChanges say 10-50%, and the sim gives 10% a point (x1.5 at 5/5, talents.go:211). The curve values are not in our download, so the client cannot settle 50% vs 100%. Beta: off-hand white swings' Rage with 0 and 5 points of DWS, same weapon, Battle Stance, no other Rage talents.
* **NtV-2 Improved Slam's cooldown at one point.** 70009 gave the -3 s effect a rank curve (1.5/3 s, as the notes say); 70291 removed that curve row, so the client now takes 3 s off at either rank, and ForeverChanges reads "3 sec per rank" (6 s at 2/2). The sim uses 3 s at 1 or 2 points (slam.go:26). Every preset takes 2/2, so only a 6 s reading would change DPS (a 12 s Slam). Beta: Slam's cooldown on the action bar with Improved Slam 1/2 and 2/2.
* **NtV-3 Deep Wounds and caster damage modifiers.** Forever's Deep Wound damage spell (412613) carries SPELL_ATTR6 bit 29 (ignore caster damage modifiers) since the beta opened. The sim follows Classic's double dip: the weapon damage behind the DoT is taken with the attacker multipliers (deep_wounds.go:85/:89) and every tick applies `AttackerDamageMultiplier` again (:44), so Death Wish (+20%), Enrage (+2-10%) and the like count twice. If the server honours the attribute, the ticks should take no caster modifiers (one dip, or none if the weapon average is also raw). Beta: Deep Wounds tick sizes from the same crit with and without Death Wish up.
* **NtV-4 Berserker Rage's bonus Rage under Forever.** The tooltip still says "generating extra rage when taking damage" and the amount is server side. Since the 8 October build Forever's damage-taken Rage is about a third of Classic's at 60, but the sim's Berserker Rage bonus still uses Classic's conversion (15.4 for a 1500 hit). Beta: Rage from the same mob hit with and without Berserker Rage up.
* **NtV-5 Sunder Armor's new aura type (674).** 70291 moved every player Sunder rank's armor effect from aura 22 (mod resistance) to a new aura type 674, and 70425 did the same to NPC Sunders; values are unchanged. It probably implements "usable even if the target has a more powerful Expose Armor" (deep dive), which would let Sunder and Improved Expose Armor coexist or not. Beta: target armor with 5 Sunders and a 5-point Expose Armor applied in both orders.
* (Existing, unchanged by these builds) Rend's attack power (WAR-DD-06), normalized Rage k and its level scaling, the damage-taken Rage constant 20, Weaponmaster's sword proc ICD: already in notes.md "Warrior".

#### 5. Count

Rows by status (sections 1 and 2, 83 rows): landed 52, missing 2, wrong 2, ambiguous (needs verify) 4, n/a 23, partial 0, not simulated on purpose 0, in flight 0. Missing with DPS impact: 1 (Slam pushback); wrong with DPS impact: 1 (Bloodthrill 6 s vs 5 s), plus 1 with no preset (Berserker Rage).

### Paladin: game-update audit ledger (builds 70009 to 70425)

Scope: sim/paladin, the paladin bits of sim/core (Retribution Aura, Undead racial on the paladin presets), the nine presets `tools/dps_report/req/paladin__*.json` (all Undead, Level 60; twist = `52003-503-05225331001330311`, Shock = `55003003201121251--05225230001`, sub Holy = `550030032--0521/2225331001330321/11`, undead variant `250030012-502-05225231001330301`). Before PR #54 the presets ran Retribution Aura. The PR switches them to Concentration Aura.

"Landed" checks are against **origin/master 4f5c60833** (racials #56 merged; no `sim/paladin/*.go` change since 6a87c9c). "In flight" means **PR #54** = `origin/claude/wow-forever-dps-presets-m9u15h-paladin` at 718e65ba0 (commits 9acc23f09, 9ff9618fb, afb4e9345 Salvation, ecf6c7ac7 and 446561dff Concentration Aura). Client values come from `S/db2/*.csv` and `S/diffs/*.md`. Trait curves (the Curve table) are not in the local download.

Sources: DN1 = [dev notes post 1](https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/1) 2026-09-24, build 70009. DN4 = [post 4](https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/4) 2026-10-01, build 70170. DN5 = [post 5](https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/5) 2026-10-08, build 70291. DN6 = [post 6](https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/6) 2026-10-15, build 70425. DD = [Paladin deep dive](https://worldofwarcraft.blizzard.com/en-us/news/24302099), 2026-10-15. FC = foreverchanges.pro/patch-notes.

#### 1. Announced changes

##### Build 70425 (DN6, 2026-10-15)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| PAL-70425-01 | DN6, 70425 | Seal of Righteousness "gained almost twice as much benefit from Spell Damage than intended": fixed | **yes** (all 9 presets) | Proc spells 25713 (r8), 25735-25742: SpellEffect idx0 `EffectBonusCoefficient` 0.1 → **0**. JoR 20286 idx2 (dummy, bp 1786 ppl 47) keeps coef 0.2. The client has dropped the engine-side coefficient and the server script's share is left. **Sim should count spell power once**: 0.1 (×1.1 on a two-hander if the SoD factor stands) | **wrong (missed by worker)** | master and PR `sim/paladin/sor.go:111,127,131`: `coeff *= 1.1`, then `BonusCoefficient: coeff` **and** `baseDamage := damage + spell.BonusCoefficient*(SP...)`, which counts 2 × 0.11 of SP. PR #54 saw the 0 coefficient and chose to keep the double count (wiki "not taken", spec_assumptions `pal_sor_sp_double` still Expected = twice). The dev note settles the assumption. PR's own sims put "once" at −54 DPS for twist and −50 for Shock at 180 s |
| PAL-70425-02 | DN6, 70425 | JoR and JoF get the instant-cast coefficients. JoR's "incorrect Spell Damage coefficient" fixed | yes (JoR); JoF n/a | JoR 20286 and every rank (20187, 20280-20285): 0.5 → **0.429**. JoF 20183/20411-20414/1311650/1311655: 0.45 → 0.429 | in flight (worker PR) | PR `sor.go:94` `TernaryFloat64(IsForever, 0.429, …)`. Master `sor.go:94` 0.5. The comment at PR `sor.go:45` still says "50% on every rank" (stale). Seal of Fury and JoF are not in the sim, on purpose |
| PAL-70425-03 | DN6, 70425 (FC: listed; the worker found it first by diffing) | Consecration SP coefficient a tick 0.095 → 0.042 | **yes** | 1280345-1280349 idx1 (the first-four extra part) and 1280353 idx0: 0.095 → **0.042**. idx0 of the rank spells (the "everyone" part) has coef 0 | in flight (worker PR) | PR `consecration.go:48` `tickCoefficient = 0.042`. Master `:47` 0.095. Applying it to the whole tick equals the client for a first-four target (all sim targets are first-four). The PR moves `pal_consecration_coef` Expected to 4.2% |
| PAL-70425-04 | DN6, 70425 | Righteous Fury Holy threat 60% → 90% | n/a (threat) | 25780 idx0 bp 60 → 90 | in flight (worker PR) | PR `righteous_fury.go:16` 1.9 (master 1.6) |
| PAL-70425-05 | DN6, 70425; DD | Champion of the Light 20/40/60% → **10/20/30%** of Intellect as spell damage | **yes** (twist, sub Holy, undead) | Spell 1311084 idx0 aura 174 bp 100 is unchanged; the ranks come from trait curve 110350, which is not in the local download (the PR read it on wago). DD says "up to 30%" | in flight (worker PR) | PR `talents.go:49` `{0,.1,.2,.3}`. Master `:48` `{0,.2,.4,.6}`. Talent JSONs updated in the PR |
| PAL-70425-06 | DN6, 70425 | Salvation / Mass Salvation replace the blessings (no blessing slot); Revoke learned | n/a | 1038 renamed Salvation; 25895 renamed Mass Salvation; 1912, 1325499 (Revoke) added | in flight (worker PR) | PR afb4e9345, proto comment `proto/paladin.proto:68`. No DPS effect (coordinator routing) |
| PAL-70425-07 | DN6, 70425 | Shield Specialization mana 6% → 4% (3 s ICD) | minor (sim-modelled, no preset takes it; raid hits from #55 make blocks possible) | 1310925 idx0 bp 6 → **4**; 20150 ProcCategoryRecovery 3000 unchanged | **missing (missed by worker)** | master and PR `talents.go:218/219` `0.06*paladin.MaxMana()`, comment `:195-196` says 6% |
| PAL-70425-08 | DN6, 70425 (also DD's "we plan to fix") | Reckoning stores charges (max 4); the next auto attack fires them all | minor (no preset takes it) | 20177 idx0 aura 42 → 4, trigger 20178 removed; new 1326084 "Reckoning" (CumulativeAura 4, ProcCharges 1, ProcTypeMask 4 = on your melee swing) | **missing (missed by worker)** | master and PR `talents.go:188/189` `ExtraMHAttack(sim, 1, …)` fires at once. Should add a stack and spend them on the next MH white swing (and keep the 1.5 s ICD on blocks) |
| PAL-70425-09 | DN6, 70425 | Infusion of Light no longer gives Holy Shock 50% damage against enemies | minor | server side (no client change on Infusion spells) | landed | The sim never modelled the bug: Infusion of Light is not read (`talents.go:12-16`), and no preset takes it |
| PAL-70425-10 | DN6, 70425 | Sanctified Judgement mana return makes no threat | n/a | 1311077 Attributes_1 0 → 1024 | n/a | threat |
| PAL-70425-11 | DN6, 70425 | Seal of Fury no longer overwritten by PW:S; SoF rank trained at 26 | n/a | 20163 SpellLevels 25 → 26 | n/a | Seal of Fury not simulated (tank seal) |
| PAL-70425-12 | DN6, 70425 | Light's Vigil works with focus macros | n/a | none | n/a | UI |
| PAL-70425-13 | DN6, 70425 (Undead racial) | Touch of the Dead (Grave): 20% for Paladins/Rogues/Warriors, 25 s ICD, no Shadow scaling, no rank penalty | **yes** (all 9 presets are Undead) | 1260189 SpellAuraOptions ProcChance 5 → **20**, ProcCategoryRecovery 1000 → **25000** | landed | `sim/core/racials_forever.go:342` `procChance := 0.20`, `:371` 25 s (shared #50/#56). The drain still runs through `CalcAndDealDamage` (target Shadow-taken multipliers apply). The presets carry no Shadow debuff, so 0 effect here; this belongs to the racials ledger |
| PAL-70425-14 | DN6, 70425 | Wands need 0.5 s standing; Scrolls on GCD; other non-paladin lines | n/a | | n/a | |

##### Build 70291 (DN5, 2026-10-08)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| PAL-70291-01 | DN5, 70291 | Retribution Aura uses the caster's SP, not the struck target's | yes (master presets run Ret Aura; it only fires on raid hits) | server side | landed | `sim/core/buffs.go:731-737`: the paladin's own aura gets `ForeverRetributionAuraCoefficient` |
| PAL-70291-02 | DN5, 70291 | Retribution Aura back to Vanilla base (~33% less), SP ratio 13.3% → 6% | yes (tiny) | 7294 7→5, 10298 12→8, 10299 18→12, 10300 24→16, 10301 30→20 (idx0 aura 15). Coefficient server side | landed | `buffs.go:721` `{0,5,8,12,16,20}`, `:726` `0.06`. PR #54 moves the presets to Concentration Aura, so this row stops mattering for the paladin presets after merge |
| PAL-70291-03 | DN5, 70291 | Reckoning at most once every 1.5 s | minor | 20177 ProcCategoryRecovery 0 → 1500 | landed | `talents.go:172` 1500 ms |
| PAL-70291-04 | DN5, 70291 | Seal of Fury mana without Improved SoF fixed; ISoF and Shield Spec make no threat | n/a | 1314104, 1310925 Attributes_1 → 1024 | n/a | tank |

##### Build 70170 (DN4, 2026-10-01)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| PAL-70170-01 | DN4, 70170 | Redoubt block 6/…/30% → 4/…/20% | minor | 20128 bp 30 unchanged; curve | landed | `talents.go:126` `4.0 * Redoubt` |
| PAL-70170-02 | DN4, 70170 | Holy Shield block 20% → 30% | minor | 20925/20927/20928 idx0 bp 20 → 30 | landed | `holy_shield.go:32` `30.0` |
| PAL-70170-03 | DN4, 70170 | Champion of the Light 33/66/100% → 20/40/60%; no healing | yes | 1311084 idx1 (aura 175, healing) removed; damage ranks on the curve | landed on master, superseded by PAL-70425-05 | `talents.go:48` |

##### Build 70009 (DN1, 2026-09-24)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| PAL-70009-01 | DN1, 70009 | Retribution Aura scales with the caster's SP | yes | server side | landed | see PAL-70291-01 |
| PAL-70009-02 | DN1, 70009 | Consecration fixes (all targets, debuff applies, Z axis, no vertical pulls) | minor | 1280353 targets 6 → 87/16, DurationIndex 1 → 1476; 1280345-49 Attributes_3 | not simulated on purpose | `consecration.go:11-16`: every target is one of the first four |
| PAL-70009-03 | DN1, 70009 | Holy Strike weapon % 25/29/32/36/39/43/46/50 | **yes** | 678 25→29, 1866 30→32, 680 30→36, 2495 35→39, 5569 35→43, 10332 40→46, 10333 40→**50** (idx1 eff 31) | landed | `holy_strike.go:44-51` (r8 `weaponPercent: 0.50`, flat 93 + 3.2/lvl, coef 0.429 ×wpn%, all equal to client 10333 idx0) |
| PAL-70009-04 | DN1, 70009 | Holy Strike cooldown 12 → 10 s; Improved Holy Strike removed | **yes** | all 8 ranks CategoryRecoveryTime 12000 → 10000; TraitNode 105328 removed (entry 130022 = 1310902 Improved Holy Strike) | landed | `holy_strike.go:56` 10 s; not in the sim tree |
| PAL-70009-05 | DN1, 70009 | Holy Power also gives Holy Strike its 15% crit | **yes** | 5923 idx1 bp 10 → 15, mask_3 + 32768 | landed | `holy_strike.go:94` `3 * HolyPower`; `holy_shock.go` the same |
| PAL-70009-06 | DN1, 70009 | Vengeance on non-periodic crits only; max stacks 5 → 3 | **yes** | 20049 ProcTypeMask 332116 → 69972 (periodic bit gone); 20050 CumulativeAura 5 → 3 | landed | `talents.go:255-264`, 1%/pt a stack, 30 s, MaxStacks 3, hit callback only |
| PAL-70009-07 | DN1, 70009 | Two-Handed Weapon Spec 3/6/9 → 2/4/6% | **yes** | curve (12163/20111 store 1/2) | landed | `talents.go:233` `{1,1.02,1.04,1.06}` |
| PAL-70009-08 | DN1, 70009 | Sacred Arbiter Holy Strike +10% → +20% | **yes** | 1311087 idx0 bp 10 → 20 | landed | `holy_strike.go:60` `*= 1.2` |
| PAL-70009-09 | DN1, 70009 | Twist of Light also −20% Seal mana | **yes** | 1310735 idx0 now aura 108 misc 14 bp −20 | landed | `talents.go:94` `benediction()*80/100` |
| PAL-70009-10 | DN1, 70009 | Echo of Justice proc chance fixed | n/a | server side (1311705 ProcChance 100 unchanged) | n/a | Seal of Justice is not a DPS seal |
| PAL-70009-11 | DN1, 70009 | Righteous Fury 90 → 60% | n/a | 25780 bp 90 → 60 | superseded (PAL-70425-04) | |
| PAL-70009-12 | DN1, 70009 | FoL animation; Light's Vigil tooltip; Aura flyout | n/a | | n/a | |

##### Deep dive (DD, 2026-10-15) mechanics

| ID | Source | Statement | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| PAL-DD-01 | DD | Judgement no longer consumes the Seal | **yes** | server side | landed | `judgement.go:56-60` |
| PAL-DD-02 | DD | Consecration baseline at 20; full damage to the first 4, low to the rest | yes | 26573 lvl 20; 1280349 idx0 12/tick (all) + idx1 27/tick (first 4) = 96+216 | landed | `consecration.go:26-30` |
| PAL-DD-03 | DD | Seal twisting only through Twist of Light echoes | **yes** | 1310735 | landed | `twist_of_light.go` |
| PAL-DD-04 | DD | Improved Seals: all seals and judgements, 3 points | **yes** | 20224 aura 108 (curve 5/10/15) | landed | `talents.go:62` (see U-01 for 70425's bp 3 → 0) |
| PAL-DD-05 | DD | Reverence: up to 30% regen while casting | yes (Shock, sub Holy) | | landed | `talents.go:42` |
| PAL-DD-06 | DD | Purifying Power: −33% Exorcism / Holy Wrath cooldown | yes | | landed | `talents.go:115` `{1,.83,.67}` |
| PAL-DD-07 | DD | Consecrated Ground: +10% Holy damage to the first 4 | yes (Shock) | 1310905 bp 10 | landed | `talents.go:327` 5%/pt |
| PAL-DD-08 | DD | Divine Precision: up to 18% Holy spell hit | yes (Shock) | 1310904 aura 107 misc 16 bp 18 | landed | `talents.go` `divinePrecisionHitRating` 6%/pt |
| PAL-DD-09 | DD | Holy Shock 21-pt, 10 s cd; Divine Favor 16-pt | yes (Shock) | 20930 CategoryRecovery 10000; 25902 bp 348 v .079 coef .429 | landed | `holy_shock.go:26-30,65` (334-362, 0.429) |
| PAL-DD-10 | DD | Light's Vigil 1.5 s cast, 30 s, 6 s cd, 75% refund, next Holy Shock no cd | yes (Shock) | 1311598 bp 395 coef .429; 1311595 idx1 bp 75 | landed | `lights_vigil.go` |
| PAL-DD-11 | DD | Precision: all spells and attacks | yes | 20189 aura 54 + 55 | landed | `talents.go:18-22` |
| PAL-DD-12 | DD | Benediction 10% off all instants; Holy Conduit −40% Consecration/HW/Exorcism | yes | 20101 bp −10 misc 14 | landed | `talents.go` `benediction()`, `holyConduit()` |
| PAL-DD-13 | DD | Vindication up to +3% AP for 30 s | yes | 440668 bp 15 → 3 (70291) | landed | `talents.go:292` |
| PAL-DD-14 | DD | Sanctified Judgement up to 100% chance, 60% of the seal cost back | yes | 1311074 bp 100 / 60 | landed | `judgement.go:11-12` |
| PAL-DD-15 | DD | Sacred Arbiter: +20% Holy Strike and refreshes all judgements | yes | | landed | damage `holy_strike.go:60`; the refresh is moot (raid judgements are permanent debuffs in the presets) |
| PAL-DD-16 | DD | Vengeance 3%/stack, 30 s, 3 stacks | yes | 20050 | landed | PAL-70009-06 |
| PAL-DD-17 | DD | Instrument of Law: Hammer of Wrath −1 s cast; −20% threat without RF | yes | | landed (cast); threat n/a | `hammer_of_wrath.go:61` |
| PAL-DD-18 | DD | Improved SotC folded into the judgement; Improved Ret Aura / Sanctity Aura removed | yes | | landed | `sotc.go:34`, `retribution_aura.go` comment |
| PAL-DD-19 | DD | Protection kit (Toughness, Anticipation, ISoF, IRF, Sacred Duty, Swift Judgement, Templar's Bulwark, Iron Creed, Holy Shield prereq), LoH 20 min, resistance auras, Voice of Truth, Eye for an Eye, Pursuit of Justice, Repentance | n/a / minor | | n/a | Swift Judgement modelled (`swift_judgement.go`), no preset takes it |

#### 2. Unannounced client changes

| ID | Build | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| PAL-U-01 | 70425 | Improved Seals rank spell's base points 3 → 0 | yes (all presets, 3/3) | 20224 idx0 aura 108 bp 3 → **0**; its TraitDefinitionEffectPoints row is unchanged (curve, not downloaded) | ambiguous (needs verify) | sim `talents.go:62` 5/10/15%. Champion of the Light shows the curve overrides the spell's bp (1311084 holds 100 for 10/20/30), so likely harmless. NtV-2 |
| PAL-U-02 | 70425 | SoR tooltip formula rewritten: `${($s1/87+$<1h>)*$<mult>}` to `${($s1/25+$<2h>)*$<mult>}` on 20154, 20287-20293, 21084 | yes | new description variables `$<1h>`, `$<2h>`, `$<mult>` (SpellDescriptionVariables not in the download) | ambiguous (needs verify) | Points to a spell power term that differs between one-handers and two-handers, plus a multiplier (Improved Seals?) in the tooltip. Bears on the ×1.1 two-hander factor (`sor.go:111`) and `pal_sor_sp_speed`. NtV-1 |
| PAL-U-03 | 70425 | SoR proc coefficient 0.1 → 0 | yes | as PAL-70425-01 | wrong (missed by worker) | the same row; the client side of the announced fix |
| PAL-U-04 | 70425 | Sense Undead out of the spellbook (FC "found in the game") | n/a | 5502 SpellLevel 20 → 0 | n/a | |
| PAL-U-05 | 70425 | Attributes_6 0x200000 added to HoJ, BoP, Divine Shield/Protection, Forbearance, Repentance, Rebuke, Avenger's Shield; SoR 20293 Attributes_8 0x1000 | n/a | SpellMisc | n/a | CC / immunity flag and client-side aura points; no damage |
| PAL-U-06 | 70009 | Crusade removed from the Retribution tree (FC "not in Blizzard's notes") | yes | TraitNode 110883 removed (entry 137881 = 1311083 Crusade) | landed | not in the sim tree (`retribution/retribution_test.go:41`) |
| PAL-U-07 | 70009 | Holy Power rebuilt as a class-masked crit modifier (aura 57 → 108 misc 7) | yes | 5923 idx0 | landed | `talents.go` `holyPowerCrit` lists the masked spells |
| PAL-U-08 | 70009 | Light's Vigil "one per party" dummy effect removed; range changes | n/a | 1310911/1311590/1311595 idx2 removed | n/a | |
| PAL-U-09 | 70170 | Righteous Fury gains aura 668; Devotion Aura aura 22 → 674 (70291) | n/a | 25780 idx3; 465 etc | n/a | threat / armor |
| PAL-U-10 | 70170 | Paladin Revelation buff 1323419 (+100% crit on the next spell, class set 10) | n/a | added | n/a | Revelation weapon enchant (items, `sim/common/enchant_effects.go`); no paladin preset uses it |
| PAL-U-11 | 70170 | T1 Retribution 4P text Demons → Undead | yes (undead presets with T1) | 1301084 aura 102 bp 36 misc 32 (Undead) in every build | landed | `item_sets_pve.go:183-199` |
| PAL-U-12 | 70291 | Vindication AP 15% → 3% in the spell | yes | 440668 bp 15 → 3 | landed | DD states 3% (PAL-DD-13) |
| PAL-U-13 | 70291 | New HoJ / Repentance / Turn Undead helper auras (1324513-1324546) | n/a | added | n/a | CC |

**Client-only builds**: 70291→70334 adds only Summon Azurky (item). 70334→70338 and 70338→70374 change no spells. Their lookup tables have no paladin TraitNode, TraitNodeEntry or TraitDefinitionEffectPoints rows (checked against tree 1100). The trait diff over all builds finds only the two 70009 removals (Improved Holy Strike, Crusade). **Nothing changed for the paladin in the client-only builds.**

#### 3. Proposed fixes

1. **PAL-70425-01, Seal of Righteousness counts spell power once** (wrong; affects all 9 presets: twist and sub Holy twist into SoR, Shock and undead run it as the primary seal).
   - `sim/paladin/sor.go`, `registerSealOfRighteousness`: keep the script term in `baseDamage` (`:131`) with a local `scriptCoeff := coeff` (0.1, ×1.1 on a two-hander pending NtV-1). Set the proc spell's `BonusCoefficient` (`:127`) to 0 under Forever, so `spell.BonusCoefficient` no longer adds a second share.
   - Update the comments at `:45-48` and `:130`.
   - `tools/dps_report/assumptions_def.py` `pal_sor_sp_double`: Expected = "Once (client 70425 + dev notes)", keep "Twice" as High. Fix patch 0004's knob (`:412-418`) to match.
   - `orchestration/spec_assumptions.md`, wiki and notes.md: "settled by DN6".
   - Test: in `sim/paladin` a unit test that adding 100 spell power raises the mean SoR proc by about 11 (two-hander), not 22. Then regenerate the TestRetribution and TestProtection goldens.
   - Impact (PR #54's own sims, 180 s): twist −54 (~−5.5%), Shock −50, sub Holy about like twist. Presets: every `paladin__level_60*.json` (9); metas need the new numbers.
2. **PAL-70425-07, Shield Specialization 4%**: `talents.go:218` `0.06` → `0.04`, comment `:195-196`. Test: a block restores 4% of max mana. DPS impact 0 on the presets (none take it); it does change the protection goldens.
3. **PAL-70425-08, Reckoning charges**: `talents.go:applyReckoning`.
   - Replace the immediate `ExtraMHAttack` with a "Reckoning" aura (spell 1326084, MaxStacks 4) that gains a stack per trigger.
   - On the paladin's next MH white swing (`OnSpellHitDealt`, `ProcMaskMeleeMHAuto`), fire `ExtraMHAttack(sim, stacks, …)` and drop the aura.
   - Keep the 1.5 s ICD (70291).
   - Test: two triggers before a swing give 2 extra attacks, and the swing timer is not reset. DPS impact 0 on the presets.
4. **PR #54 hygiene**: the wiki line "The sim takes all of it" is wrong while row 1 is open. `sor.go:45` still says "50% on every rank".

#### 4. Need to Verify candidates

* **NtV-1, Seal of Righteousness after the 70425 fix (beta).** DN6 says SoR "gained almost twice as much benefit from Spell Damage than intended" and is fixed. Client 70425 zeroes the proc spells' 10% coefficient (25713 and lower ranks), while Judgement of Righteousness still carries the seal's 1786 / 47 dummy at 20%. The sim should count SP once. Still open: whether the surviving share is 10% flat, 11% on a two-hander (the Season of Discovery ×1.1), or tied to weapon speed. 70425's tooltip now has separate `$<1h>` and `$<2h>` terms plus a `$<mult>` (SpellDescriptionVariables, not downloaded), which hints at a hand-dependent SP term. To settle it, compare SoR proc hits at level 60 with a 2H and a 1H, with and without about 100 spell power (one Brilliant Wizard Oil swap). Knobs: `pal_sor_sp_double`, `pal_sor_sp_speed`, `sor.go:111`.
* **NtV-2, Improved Seals' stored value (client).** Client 70425 changed Improved Seals (20224) base points from 3 to 0 without a note. The talent's ranks come from a trait curve, which overrides the spell, as Champion of the Light shows (spell holds 100, ranks are 10/20/30). The TraitDefinitionEffectPoints row did not change. Read the 70425 curve for 20224's trait (wago Curve / CurvePoint), or check that a 3/3 paladin's SoR and JoR still hit 15% harder than 0/3. Sim keeps 5/10/15% (`talents.go:62`). If it were really 0, every preset loses about 10-13%.
* **NtV-3, Champion of the Light 10/20/30 (client curve).** PR #54 takes 10/20/30% from trait curve 110350. The local download has no Curve table to confirm it, but DN6 and the DD ("up to 30%") agree, so this is low risk. Recheck the character sheet's spell damage with 3/3.

#### 5. Count

Rows by status (sections 1 and 2, 65 rows): landed 37, in flight (worker PR) 5, wrong (missed by worker) 2 (PAL-70425-01 and its client side PAL-U-03), missing (missed by worker) 2 (Shield Spec 4%, Reckoning charges), ambiguous (needs verify) 2, not simulated on purpose 1, superseded 2, n/a 14.

### Hunter and hunter pets: game-update ledger (70009 to 70425)

Scope: sim/hunter (hunter, pet, hawk), presets tools/dps_report/req/hunter__*.json (9 ranged: beast-mastery, marksmanship, lone-wolf, each plain / duty-mark / duty-scorpid; 2 melee: survival, survival-duty-mark). Master is 6a87c9c. Worker branch origin/claude/wow-forever-dps-presets-m9u15h-hunter is master + one commit, 3d0e70e9b "Hunter: client 1.60.1.70425 talent changes" (Resourcefulness covers Strider Kick, Resourcefulness proc curve, Improved Mend Pet curve, Natural Armor aura note). Anything else 70425 below is **missed by worker**.

Preset facts used below: all presets are Night Elf except Survival (Orc); every pet preset uses a Cat (Bite + Claw); no preset brings a wolf, and no preset raid buff list has Furious Howl. Talent ranks read with ui/core/talents/trees/hunter.json order: Marksmanship has Efficiency 4, Mortal Shots 5, Sniper Shot; Lone Wolf has Efficiency 5, Mortal Shots 5, Sniper Shot; Beast Mastery has Mortal Shots 4, no Sniper Shot; Survival has Resourcefulness 1, Clever Traps 2, Survivalist's Discipline 2 and casts Immolation Trap rank 5 (14305).

Sources: DN = Blizzard dev notes thread https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696 (post /1 = 70009 2026-09-24, /4 = 70170 2026-10-01, /5 = 70291 2026-10-08, /6 = 70425 2026-10-15). DD = Hunter deep dive https://worldofwarcraft.blizzard.com/en-us/news/24301515 (2026-09-30). FC = foreverchanges.pro/patch-notes.

#### 1. Announced changes (newest first)

##### Build 70425 (DN /6, 2026-10-15)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| HUN-70425-01 | DN /6, 10-15, 70425 | Auto Shot needs 0.5 sec standing still before it shoots; still fires while casting Aimed/Sniper Shot | minor | Auto Shot (75) SpellInterrupts unchanged 70374→70425: server side | not simulated on purpose | sim/hunter/hunter.go:162-175 gives Forever's Auto Shot no aim time and lets it fire while casting. All presets stand still the whole fight (Survival at 5 yd never switches to ranged), so the stand-still rule never triggers. |
| HUN-70425-02 | DN /6 | Arrows visible when the bow is drawn | n/a | visual | n/a | |
| HUN-70425-03 | DN /6 | Pets now correctly benefit from Expertise | minor | Hunter Pet Scaling 415429 has no expertise effect in 70374 or 70425: server side | missing (zero impact today) | sim/hunter/pet.go:236-246 `makeStatInheritance` passes AttackPower (0.1 x RAP), MeleeCrit, MeleeHit only. No hunter source of stats.Expertise exists in the sim (only sim/rogue/talents.go grants it), so nothing moves. Add `stats.Expertise: ownerStats[stats.Expertise]` for completeness. |
| HUN-70425-04 | DN /6; FC "Immolation Trap: fire traps scale with attack power" | Fire traps now correctly scale with Attack Power | **yes** (Survival casts Immolation Trap 14305) | Immolation Trap effect spells 13797/14298/14299/14300/14301 and Explosive Trap Effect 13812/14314/14315: BonusCoefficientFromAP 0 and EffectBonusCoefficient 0 in both 70374 and 70425 (14301: 138 a tick x 5). Server side, coefficient unknown | missing | sim/hunter/immolation_trap.go:56-58 snapshots a flat `dotDamage/5` (690 total at rank 5); sim/hunter/explosive_trap.go:151-176 flat too. No AP term. Need to Verify (section 4). |
| HUN-70425-05 | DN /6 | 1 active Fire trap and 1 active Frost trap at once (was 1 among all) | minor | trap summon effects (Effect 104) on 13795/14302-14305 and 13813/14316/14317: EffectMiscValue_1 0 → 3 (a trap slot) | landed | sim/hunter/hunter.go:220-226 separate Fire and Frost timers; the sim has no cap on live traps, and Survival only lays Immolation Trap. |
| HUN-70425-06 | DN /6; FC | Pets take the hunter's higher crit (melee or ranged) instead of always melee | yes (all pet presets) | server side (415429 crit effects unchanged) | landed | sim/hunter/pet.go:243 inherits `stats.MeleeCrit`, which is the sim's one crit stat for melee and ranged. Ranged-only reductions are per spell (hunter.go:359-382, Human/Dwarf only), so MeleeCrit is always the higher. |
| HUN-70425-07 | DN /6 | Dismember (Crocolisk) 25% heal cut, 6 s, 15 s CD; Sonic Blast (Bat) 30% slow, 20 s, 60 s CD; Mine! (Bird of Prey) 3 s, 90 s CD | n/a | confirmed: Dismember 1264758/927/929/930/933 bp -50→-25, CD 6000→15000, DurationIndex 1→32 (6 s); Sonic Blast 1264478-82 bp -60→-30, CD 30000→60000, DurationIndex 9→18 (20 s); Mine! 1265054-58 CD 60000→90000, DurationIndex 35→27 (3 s) | n/a | Utility abilities of pets no preset uses; not simulated (sim/hunter/pet_abilities.go). |
| HUN-70425-08 | DN /6 | Summon Hawk spawns at and sticks to its target | minor | server side | landed | sim/hunter/summon_hawk.go: the hawk always attacks the hunter's target. |
| HUN-70425-09 | DN /6 | Improved Mend Pet now lowers Mend Pet's cost | n/a | trait curve (per worker) | in flight (worker PR) | 3d0e70e9b wiki note only; the sim casts no Mend Pet. |
| HUN-70425-10 | DN /6; FC | Trueshot Aura is now raid wide | yes (MM, LW presets) | 19506/20905/20906/1299346/1299348 eff 0: Effect 35 (party area aura) → 65 (raid area aura); tooltip "party or raid members" | landed (comments stale) | sim/hunter/hunter.go:144-150 sets RaidBuffs.TrueshotAura from the talent (sim/core/buffs.go:304, 1677-1686: 100 RAP at rank 5). Fix the comment at hunter.go:144-145 and notes.md:112 ("party aura under Forever; the sim only has a raid wide switch"): it is now raid wide in the client too. |
| HUN-70425-11 | DN /6 ("Sniper Shot was not benefitting from a few Marksmanship talents"); FC | Sniper Shot now benefits from the MM talents it missed | **yes** (MM x3, LW x3) | Efficiency 19416 eff 0 (row 696992) and Mortal Shots 19485 eff 0 (row 696507): EffectSpellClassMask_3 0 → 8388608, Sniper Shot's only mask bit (1310687/1310785/1310786 SpellClassMask_3 = 8388608). Barrage's mask still lacks it; Rapid Killing and Hawk Eye already had it | **missing** (missed by worker) | sim/hunter/sniper_shot.go:13 says "Mortal Shots, Barrage and Efficiency leave it out" and the config (sniper_shot.go ~line 64) sets no CritDamageBonus; talents.go:62-65 Efficiency switch omits SpellCode_HunterSniperShot; talents.go:319-323 mortalShots comment says "Not Sniper Shot". Sim should give Sniper Shot +6% crit damage a rank and 3% off a rank. |
| HUN-70425-12 | DN /6 | Strider Kick usable while disarmed / without a main-hand weapon | n/a | SpellEquippedItems row 256904 removed from 1317257 | n/a | |
| HUN-70425-13 | DN /6 | Resourcefulness now reduces Strider Kick's cost | yes (Survival x2) | 440529 eff 0 EffectSpellClassMask_2 0 → 2097152 (Strider Kick) | in flight (worker PR) | branch: sim/hunter/talents.go `applyResourcefulness` adds SpellCode_HunterStriderKick; TestResourcefulnessClassMask. |
| HUN-70425-14 | DN /6 ("only activating on melee and ranged ability critical strikes ... not activated by Traps") | Resourcefulness crit procs work again, but only from melee and ranged *abilities* | yes (Survival x2, mana only) | 440529 SpellAuraOptions row 242512 ProcTypeMask_0 328020 (melee auto + melee ability + ranged auto + ranged ability + harmful spell + periodic) → 272 (melee ability + ranged ability); TraitDefinitionEffectPoints 23260 CurveID 82780→85425 (proc chance 50/100) | **wrong** (curve: in flight; proc mask: missed by worker) | sim/hunter/talents.go:394-400 uses `ProcMask: core.ProcMaskMeleeOrRanged`, which includes white melee swings and Auto Shot. Should be `core.ProcMaskMeleeSpecial \| core.ProcMaskRangedSpecial`. Branch did not touch it. |
| HUN-70425-15 | DN /6 | Predator's Edge tooltip: only abilities' crit damage | yes (Survival) | 1310627 tooltip "melee ability critical strike damage" | landed | sim/hunter/talents.go:130-139 `isMeleeAbility` only; OH +10%/rank at :140-142. |
| HUN-70425-16 | DN /6 Interface | Auto Attack/Auto Shot switch setting | n/a | UI | n/a | |

##### Build 70291 (DN /5, 2026-10-08)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| HUN-70291-01 | DN /5 | Disengage threat reduction doubled as intended | n/a | already in client 70170: 781/14272/14273 bp -140/-280/-405 → -280/-560/-810 | n/a | threat |
| HUN-70291-02 | DN /5 | Pets no longer regain health repeatedly when low | n/a | Hunter Pet Scaling 415429 eff 0 aura 34→230, Pet Hardiness 6280-6286 aura 34→230 (health) | n/a | |
| HUN-70291-03 | DN /5 | Unleashed Fury now correctly increases pet damage | yes (BM x3) | 19616 eff 0 aura 108 → 107, bp 20, misc 3, mask incl. Summon Hawk | landed | sim/hunter/talents.go:34-35 pet x(1+0.03/rank); hawks summon_hawk.go:47. |
| HUN-70291-04 | DN /5 | Improved Concussive Shot correct chance | n/a | 19407 ProcChance 20→100 (+ new DR spell 1324511) | n/a | |
| HUN-70291-05 | DN /5 | Improved Wing Clip correct chance, less with low ranks | n/a | 19228 ProcChance 4→100 | n/a | root on a boss |
| HUN-70291-06 | DN /5 | Expose Prey lasts 10 sec (was 5) | yes (Survival x2) | 1310726 DurationIndex 28 (5 s) → 1 (10 s) | landed | sim/hunter/talents.go:418-421 sets DefensiveState 10 s; TestExposePreyWindow. |
| HUN-70291-07 | DN /5 | Entrapment diminishing returns; only on Frost Trap's first activation | n/a | 19185 DiminishType 0→1 | n/a | |
| HUN-70291-08 | DN /5 known issue | Ferocity: pet crit does not show on the sheet, visual only | yes (BM) | server side | landed | sim/hunter/talents.go:37-39 +2% pet crit a rank. (orchestration/pet_buffs.md:147 still calls it a real bug; the 10-08 note says visual only.) |

##### Build 70170 (DN /4, 2026-10-01)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| HUN-70170-01 | DN /4 | Sniper Shot 45 yd range; next 3 shots within 10 s +10 yd | n/a | 1310687/785/786 RangeIndex 114→582 (8-45 yd), new eff 1 aura 107 misc 5 bp 10, 3 charges, DurationIndex 0→1 | not simulated on purpose | range only; sniper_shot.go:12 "range bonus ... does nothing here" |
| HUN-70170-02 | DN /4 | Improved Stings / Predator's Edge icons; Deflection parry 1-5% (was 2-10%) | n/a | Deflection via trait curve | n/a | tanking/visual |
| HUN-70170-03 | DN /4 | Pets: Aggressive Mode back; pet ability ranks fixed for low-level beasts | n/a | | n/a | |

##### Build 70009 (DN /1, 2026-09-24)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| HUN-70009-01 | DN /1 | Furious Howl Attack Power bonus reduced 40% | minor (no preset has a wolf; no raid buff for it) | **Not in the 70009 client; it lands in 70170**: 24604 (L10) bp 9+1.0/lvl → 5+0.6/lvl; 24605 (L24) 56+0.03 → 34+0.02; 24603 (L40) 100+0.91 → 60+0.5; 24597 (L56) 136+0.7 → 82+0.4 (EffectRealPointsPerLevel; 83.6 at level 60 if it grows to MaxLevel 60). Aura 99 (melee AP), DurationIndex 3 = **60 s**, CD 30 s, 60 focus, radius 15 yd | missing (not simulated at all) | sim/hunter/pet_abilities.go:32-33 and 239-280 the ability is commented out (and carries Wrath's 320 AP / 20 s); pet.go:480 Wolf has no SpecialAbility; proto/common.proto:778 `RaidBuffs.furious_howl` exists but nothing in sim/core reads it. See unannounced U-01 (raid wide in 70425). Stacking with Battle Shout is already Need to Verify (notes.md:114), not re-raised. |
| HUN-70009-02 | DN /1 | Sonic Blast available to Bats; Tame Beast level cap; Pet Happiness decay fix | n/a | | n/a | |
| HUN-70009-03 | DN /1 | Talenting out of Lone Wolf no longer gives -4% damage with the pet out | yes (MM, BM presets have pets and no Lone Wolf) | Lone Wolf 1310684 SpellInterrupts row 250210 added (AuraInterruptFlags_1 16384) | landed | sim/hunter/talents.go:81-84 only `LoneWolf && pet == nil` gives x1.2; no penalty path exists. |
| HUN-70009-04 | DN /1 | Lacerating Strikes can crit | yes (Survival x2) | 1310536 Attributes_8 4097→4609 | landed | sim/hunter/mongoose_bite.go registerLaceratingStrikes ticks via `dot.OutcomeTick` (core/spell_outcome.go:31-41 crits when `dot.canCrit`, on under Forever). |
| HUN-70009-05 | DN /1 | Strider Kick also +30% move speed for 3 s | n/a | 1317257 new eff 2 aura 31 bp 30, DurationIndex 0→27 (3 s) | not simulated on purpose | strider_kick.go comment "run speed ... does nothing here" |

##### Deep dive (DD, 2026-09-30; design as of ~70170)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| HUN-DD-01 | DD | Aimed Shot baseline at 20 | yes | 19434 SpellLevels 20 | landed | aimed_shot.go: ranks from level 20, rank 6 166 + normalized weapon, 2 s cast, 6 s CD |
| HUN-DD-02 | DD | Arcane Shot no longer shares CD with Aimed; Multi-Shot shares with Aimed | yes | category 2 on Aimed+Multi; 1173 Arcane+Summon Hawk | landed | hunter.go:198-208 |
| HUN-DD-03 | DD | Aspect of the Beast grants melee AP | yes (Survival) | 1299447 rank 4 110 | landed | aspects.go: AP 50/70/90/110 |
| HUN-DD-04 | DD | Traps usable in combat, 30 s CD; Fire and Frost separate | yes (Survival) | categories 411/412 | landed | immolation_trap.go 30 s; hunter.go:220-226 |
| HUN-DD-05 | DD | Scorpid Sting: -2% target hit instead of Str/Agi | yes (duty-scorpid presets cast it) | 3043 | landed | scorpid_sting.go, core.ForeverScorpidStingAura; 20 s +15/rank Improved Stings |
| HUN-DD-06 | DD | Pets scale with the hunter's gear | yes | 415429 (coefficients server side) | landed (assumption) | pet.go:236-246, 10% RAP + crit + hit; Need to Verify already in notes.md:100 |
| HUN-DD-07 | DD | Pets no longer take player buffs | yes | players-only flag on buffs | landed | pet.go:86-88 IgnoresPlayerBuffs |
| HUN-DD-08 | DD | Families: Bite/Claw for all, Dive/Dash, one unique each (Fox, Furious Howl, Savage Rend ...) | minor | | partial | Cat Bite+Claw modelled; family uniques other than Lightning Breath / Scorpid Poison / Screech not simulated (Wolf's Furious Howl: HUN-70009-01). No preset uses them. |
| HUN-DD-09 | DD | Disengage threat doubled; Aspect of the Monkey, Pathfinding, Improved Mend Pet, Endurance Training (health/armor), Hawk Eye, Entrapment, Survivalist, Improved Wing Clip, Survival Tactics' Feign Death part | n/a | | n/a | threat, defence, movement |
| HUN-DD-10 | DD | Deadly Aspects: chance for Quick Shots (Hawk) / Quick Strikes (Beast) | yes | 1299448 | landed | aspects.go: 2%/rank, 30% speed 12 s |
| HUN-DD-11 | DD | Focused Fire: hunter and pet damage while pet out | yes (BM, MM) | 1223755 | landed | talents.go:27-32, 1%/rank |
| HUN-DD-12 | DD | Unleashed Fury lower, also Hawks; Ferocity lower, also Hawks | yes (BM) | 19616, 19598 | landed | talents.go:34-39 (3%, 2%); summon_hawk.go:47, :105 |
| HUN-DD-13 | DD | Summon Hawk (16 pt): instant damage + hawk 18 s, max 2, CD shared with Arcane Shot | yes (BM) | | landed | summon_hawk.go |
| HUN-DD-14 | DD | Intimidation: +100% crit on pet's next attack | yes (BM) | | landed | talents.go:221-273 |
| HUN-DD-15 | DD | Bestial Discipline adds mana regen while casting | yes (BM) | | landed | talents.go:46-48, 25%/rank; pet focus pet.go:129 |
| HUN-DD-16 | DD | Lethal Attacks: crit to all attacks | yes | | landed | talents.go:52-56 |
| HUN-DD-17 | DD | Improved Stings: more Serpent Sting damage, Scorpid Sting longer | yes | | landed | serpent_sting.go:27-28 (6/13/20%), scorpid_sting.go:23 |
| HUN-DD-18 | DD | Efficiency: bigger discount, also stings and melee | yes | 19416 | landed (Sniper Shot gap: HUN-70425-11) | talents.go:58-71 |
| HUN-DD-19 | DD | Careful Aim: AP from Intellect | yes (MM, LW) | 1223984 | landed | talents.go:73-79, 20%/rank melee + ranged |
| HUN-DD-20 | DD | Rapid Killing: shorter Rapid Fire CD; next Shot bonus after a kill | yes | 415407 | landed (CD) / not simulated on purpose (kill proc) | rapid_fire.go:80-88, -1 min/rank |
| HUN-DD-21 | DD | Improved Arcane Shot up to an extra 0.5 s | yes | | landed | arcane_shot.go: -0.3 s/rank (1.5 s vs Classic 1.0) |
| HUN-DD-22 | DD | Lone Wolf (11 pt): +20% with no pet | yes (LW x3) | 415370 bp 20 | landed | talents.go:81-84 |
| HUN-DD-23 | DD | Trueshot Aura moved to 21 pt, new low ranks | yes | | landed | core/buffs.go:1677-1686 (rank 5 100, assumption) |
| HUN-DD-24 | DD | Rapid Recuperation: regen while casting after Serpent Sting | yes (MM, LW, mana) | | landed | talents.go:330-355 |
| HUN-DD-25 | DD | Barrage lower, now also Aimed Shot | yes | | landed | talents.go:325-328 3/7/10% on Aimed, Multi, Volley |
| HUN-DD-26 | DD | Sniper Shot (31 pt): 4 s cast, 15 s CD, base 45 yd | yes (MM, LW) | 1310786: 295 + normalized weapon | landed | sniper_shot.go |
| HUN-DD-27 | DD | Improved Tracking: damage vs tracked type | yes (Survival) | | landed | talents.go:103-124 |
| HUN-DD-28 | DD | Savage Strikes: "increases all melee damage by 2/4%" | yes (Survival) | client 19159 tooltip: "critical strike chance of all your melee abilities by $s1%" | ambiguous (needs verify) | talents.go:126-139 follows the client (2% crit a rank on melee abilities). See section 4. |
| HUN-DD-29 | DD | Surefooted: hit on all attacks | yes | | landed | talents.go:148-150 |
| HUN-DD-30 | DD | Survival Tactics: trap hit | yes (Survival) | | landed | talents.go:152-163 |
| HUN-DD-31 | DD | Predator's Edge: melee crit damage, off-hand damage | yes (Survival) | 1310627 | landed | talents.go:126-142 |
| HUN-DD-32 | DD | Counterattack 16 pt (after parry) | n/a | | n/a | boss hits the tank; not in sim |
| HUN-DD-33 | DD | Resourcefulness: cheaper traps and melee, crit regen proc | yes (Survival) | 440529 | landed (cost) / see HUN-70425-13/14 | talents.go:360-401 |
| HUN-DD-34 | DD | Expose Prey: Hunter's Mark target attacks open Mongoose Bite | yes (Survival) | 1310532 | landed | talents.go:405-425 |
| HUN-DD-35 | DD | Survivalist's Discipline: trap/Deterrence CD | yes (Survival) | 1310496 | landed | talents.go:160-162, 20%/rank |
| HUN-DD-36 | DD | Strider Kick (21 pt): instant weapon damage + speed | yes (Survival) | 1317257: 100% normalized, 8 s CD, 5.81% base mana | landed | strider_kick.go |
| HUN-DD-37 | DD | Lightning Reflexes Agility decreased | yes (Survival) | | landed | talents.go:165-169, 2%/rank |
| HUN-DD-38 | DD | Lacerating Strikes (31 pt): bleed over 21 s from Mongoose Bite | yes (Survival) | 1310533/1310536 | landed | mongoose_bite.go: 40% over 7 ticks of 3 s |

#### 2. Unannounced client changes (reverse check)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| HUN-U-01 | client 70425 (DB2 diff 70374-70425, Hunter section) | **Furious Howl is now raid wide** | minor (raid buff; no hunter preset) | 24597/24603/24604/24605 eff 0 ImplicitTarget_0 20 (party) → 56 (raid); tooltip "all party or raid members within $a1 yards". Same-build pattern as Battle Shout / Trueshot | missing (not simulated) | No Furious Howl anywhere (HUN-70009-01). Stacking with Battle Shout: Need to Verify already (notes.md:114). |
| HUN-U-02 | client (unchanged, but notes.md:114 misreads it) | Furious Howl lasts 60 sec with a 30 sec cooldown; the "15" is its radius | minor | 24597 DurationIndex 3 = SpellDuration 60000 ms (69893 and 70425); SpellCooldowns RecoveryTime/CategoryRecoveryTime 30000; EffectRadiusIndex_0 18 = 15 yd | n/a (doc fix) | notes.md:114 says "82 melee attack power for 15 sec" and "for a share of the fight". With 60 s duration and 30 s CD a wolf can keep it up 100% of the fight. Rank 4 at level 60 is 82 + 0.4 x 4 = 83.6 if the per-level growth applies. |
| HUN-U-03 | client 70425; FC "Multi-Shot: Is now instant (was a 0.5 sec cast). Not in Blizzard's notes, found in the game." | Multi-Shot instant | yes (all 9 ranged presets, tiny) | 2643 SpellMisc row 315832 CastingTimeIndex 3 (500 ms) → 1 (0) | **missing** (missed by worker) | sim/hunter/multi_shot.go:10-12 comment and :36-52 `CastTime: time.Millisecond * 500` / ModifyCast / CastTime func. |
| HUN-U-04 | client 70170 vs DN /1 | Furious Howl -40% shipped in the 70170 client, not 70009 | n/a | 70009 SpellEffect rows identical to 69893; change in 70009-70170 diff | n/a | DN post /1 was edited 10-16; either a 70009 server hotfix or a late-added line. No sim effect. |
| HUN-U-05 | client 70170 | Disengage threat doubled in the client a build before the notes (70291) | n/a | 781/14272/14273 bp doubled | n/a | |
| HUN-U-06 | client 70170 | Hunter-class Revelation buff 1323410 added: next Arcane Shot / Volley +100% crit, 1 charge (SpellClassSet 9, mask 10240) | minor | new spell, EffectMiscValue 7, bp 100, ProcCharges 1 | missing (no preset impact) | sim/common/enchant_effects.go:269ff models Revelation for casters only. No hunter preset carries Revelation (MM/BM: 2646 Agility, Survival: 1900 Crusader / 2564). |
| HUN-U-07 | client 70425 | Hunter Pet Scaling 415429 eff 8 and Natural Armor 24545-24630 aura 22 → 674; Dust Cloud 1265899-904 22 → 674 (70291); global armor-aura swap | n/a | | n/a | armor; worker noted Natural Armor. |
| HUN-U-08 | client 70425 | Attributes_6 0x200000 added to Concussive Shot, Counterattack, Freezing/Frost Trap, Intimidation, Scatter Shot, Wing Clip, Wyvern Strike and pet Mine!/Pinch/Savage Rend/Tendon Rip/Web; Aimed Shot (all ranks) and Sniper Shot Attributes_16 0 → 128 | n/a | flags only | n/a | The Attributes_16 bit on the two cast-time shots fits "Auto Shot still fires while casting Aimed/Sniper Shot", which the sim already does (hunter.go:166-175). |
| HUN-U-09 | client 70291 | Explosive Trap Effect 14315 target/radius fix; Frost Trap radius; new DR helper spells (Freezing Trap, Scare Beast, Intimidation, Scatter Shot, Imp. Concussive Shot) | n/a | | n/a | single target fight |
| HUN-U-10 | client 70009 | EffectBonusCoefficient 1 → 0 on Arcane Shot Mana Reduction 23157, Devilsaur Tooth 24353, Improved Feign Death 24432, Improved Mend Pet 23560; Lacerate/Mend Pet Attributes_8 512 | n/a | spell power coefficient on cost/modifier auras, no damage | n/a | |
| HUN-U-11 | client 70425 | Scorpid Sting 18545 DispelType 0 → 4 (Poison) | n/a | 18545 is the old Str/Agi NPC version (level 1, -1.25/lvl), not the hunter's 3043 | n/a | |
| HUN-U-12 | hunter talent tree TraitTree 1091 | every node, rank count and curve compared across all 8 builds | - | only change: Resourcefulness 440529 effect 1 CurveID 82780 → 85425 in 70425 (HUN-70425-14) | - | No other hunter talent value moved between 69893 and 70425. |

Client-only builds: 70291→70334 changes only the Summon Azurky item; 70334→70338 and 70338→70374 change lookup tables only (no hunter spell or trait row). Confirmed: nothing for the hunter. 69977/70058/70124 sit inside the 69893-70170 diffs and 70205/70235/70245 inside 70170-70291; no hunter change in those pairs is unaccounted for above.

#### 3. Proposed fixes

1. **Sniper Shot gets Mortal Shots and Efficiency (HUN-70425-11).** sim/hunter/sniper_shot.go `registerSniperShotSpell`: add `CritDamageBonus: hunter.mortalShots(),` to the config and change the header comment (line 13) to "Mortal Shots and Efficiency cover it since client 1.60.1.70425 (both masks gained its bit 0x800000 in word 3); Barrage doesn't". sim/hunter/talents.go:62-65: add `SpellCode_HunterSniperShot` to the Efficiency case list and update the comments at :58-59 and :319-320. Test: extend TestEfficiencyClassMask with Sniper Shot (MM talents with Sniper Shot, Efficiency 5 → 85), and a new check that SniperShot.CritDamageBonus == 0.06 x MortalShots. Impact: Sniper Shot is roughly 8% of MM damage at ~35% crit; +30% crit damage is about +7-8% on it, so **about +0.5-0.8% DPS** for MM and LW, a little more if the 12-15% cheaper Sniper Shot lets the mana-gated rotation (cast while mana >= 25 x seconds left) fire it more often. Presets: hunter__marksmanship{,-duty-mark,-duty-scorpid}, hunter__lone-wolf{,-duty-mark,-duty-scorpid}; goldens TestMarksmanship.results, TestLoneWolf.results.
2. **Multi-Shot instant (HUN-U-03).** sim/hunter/multi_shot.go: drop `CastTime: time.Millisecond * 500`, the ModifyCast and the CastTime func (keep GCD and the 6 s shared CD); comment "instant since client 1.60.1.70425 (CastingTimeIndex 3 → 1; foreverchanges found it, not in Blizzard's notes)". Test: assert `hunter.MultiShot.DefaultCast.CastTime == 0`. Impact ~0 to +0.1% (the 0.5 s cast is shorter than the 1.5 s GCD and Auto Shot fires through it; damage lands earlier). All 9 ranged presets; all ranged goldens move slightly.
3. **Resourcefulness procs on ability crits only (HUN-70425-14).** sim/hunter/talents.go:397: `ProcMask: core.ProcMaskMeleeSpecial | core.ProcMaskRangedSpecial` (ProcTypeMask 272). Fix the comment at :360-363. Test: a white melee crit does not activate 1242688, a Raptor Strike crit can. Impact: mana only, uptime stays near full from Raptor Strike / Mongoose Bite / Strider Kick crits within 30 s, so **~0 DPS**. Presets hunter__survival, hunter__survival-duty-mark. Best landed on the hunter worker branch next to its Strider Kick change.
4. **Fire traps scale with AP (HUN-70425-04).** sim/hunter/immolation_trap.go OnSnapshot (:56-58): add `+ coeff * spell.MeleeAttackPower(target)` (or ranged AP; Need to Verify which) over the 5 ticks; same for explosive_trap.go's hit and dot. Gate the coefficient as a named report assumption until a log gives it. Impact guess: Immolation Trap is ~50 DPS for Survival (690 x 1.3 Clever Traps every 18 s); at a placeholder 10% of ~1500 AP over the trap, **+1%** for Survival. Presets hunter__survival, hunter__survival-duty-mark.
5. **Pet Expertise (HUN-70425-03).** sim/hunter/pet.go:241-245 add `stats.Expertise: ownerStats[stats.Expertise]`. No preset impact today (no hunter Expertise source).
6. **Furious Howl (HUN-70009-01, HUN-U-01, HUN-U-02).** No hunter preset changes, so only if the raid-buff question resolves "stacks": implement `RaidBuffs.furious_howl` in sim/core/buffs.go as +84 melee AP (rank 4, raid wide, 60 s / 30 s CD, full uptime) and replace the dead Wrath code in pet_abilities.go:239-280 with the Forever ranks (5/34/60/82 + per level, 60 focus). Would move every melee preset by roughly +2-3% if it stacks with Battle Shout; zero if it does not.
7. **Doc fixes.** hunter.go:144-145 and notes.md:112 (Trueshot Aura is raid wide in client 70425, not a party aura); notes.md:114 (Furious Howl lasts 60 s with a 30 s CD, 15 yd radius; rank 4 = 83.6 at 60 with per-level growth); orchestration/pet_buffs.md:147 (Ferocity known issue is "visual only" per the 10-08 note).

#### 4. Need to Verify candidates

* **Fire traps' attack power share (server side).** Blizzard's 10-15 notes say "The Hunter's Fire traps now correctly scale with Attack Power", but client 1.60.1.70425 gives Immolation Trap's tick (14301, 138 a tick x 5 at rank 5) and Explosive Trap's hit and dot (14315: 229 ± 24%, 33 a tick) no attack power coefficient (BonusCoefficientFromAP 0) and no spell power coefficient, so the server adds it and the share is unknown, as is whether it reads melee or ranged attack power. The sim deals the flat client numbers. Survival casts Immolation Trap on its 18 sec cooldown, about 5% of its damage. Beta: an Immolation Trap's ticks at level 60 (or any level) with and without Aspect of the Beast / Aspect of the Hawk, attack power known from the sheet.
* **Savage Strikes: crit or damage?** Blizzard's hunter deep dive (2026-09-30) says Savage Strikes "now increases all melee damage by 2/4%", but the client (19159, still at 1.60.1.70425) says "Increases the critical strike chance of all your melee abilities by $s1%", and the sim follows the client (2% crit a rank on Raptor Strike, Mongoose Bite, Wing Clip, Strider Kick, Lacerate and Lacerating Strikes). For Survival's 2 ranks the two readings differ by roughly 1% DPS (4% melee damage incl. white swings vs 4% crit on abilities). Beta: the talent's tooltip in game and a Raptor Strike crit rate with 0 and 2 ranks.
* **Furious Howl (refines notes.md:114, the stacking question is not re-raised).** The client gives Furious Howl a 60 sec duration (DurationIndex 3) and a 30 sec cooldown, so a wolf keeps it up the whole fight, not "a share of the fight"; the 15 is its radius in yards. Rank 4 (24597) is 82 plus 0.4 a level from 56, 83.6 at 60 if the growth runs to level 60. The 40% cut announced for 70009 only reached the client in 70170. Beta: the buff's duration and value on a level 60 wolf's party member.
* **Pet Expertise coefficient.** "Pets now correctly benefit from Expertise" has no client data (Hunter Pet Scaling 415429 has no expertise effect), so whether it is the hunter's Expertise 1:1, and whether any hunter source of Expertise exists at 60, is open. No sim effect today.

#### 5. Count

Section 1 (70 announced rows): landed 44, missing 4, partial 1, wrong 1, not simulated on purpose 3, in flight (worker PR) 2, n/a 14, ambiguous (needs verify) 1. Section 2 (12 unannounced rows): missing 3, n/a 8, info 1. All 82 rows: landed 44, missing 7, partial 1, wrong 1, not simulated on purpose 3, in flight 2, n/a 22, ambiguous 1, info 1.

### Rogue: WoW Forever beta update audit (builds 70009 to 70425)

Scope: sim/rogue, the shared code the rogue presets use (racials, the armor debuffs), and the 12 presets in tools/dps_report/req/rogue__*.json. The sim of record is master at 6a87c9c. The rogue class branch (origin/claude/wow-forever-dps-presets-m9u15h-rogue, 234dafac1, 2026-10-13) differs from master in sim/rogue only by one line in talents.go (the Blade Flurry target count) and the golden test results. Master already holds all of that branch's rogue mechanics. Nobody is porting 70425 for the rogue.

All 12 presets are Undead. They carry Deadly Poison in the main hand and Instant Poison in the off hand (RogueOptions), a Dense Sharpening Stone on both weapons, and the raid's Windfury Totem and Strength of Earth Totem. The Assassination, Combat and Seal Fate presets take Cold Blood. The talent strings, decoded against proto/rogue.proto, contain:
- Mutilate presets: Malice 5, Ruthlessness 3, Improved Slice and Dice 3, Relentless Strikes, Lethality 3, Vile Poisons 3, Cold Blood, Improved Poisons 5, Mutilate, Seal Fate 3; Improved Eviscerate 3 and Puncturing Wounds 3 in Combat.
- Combat Swords: Improved Eviscerate 3, Improved Sinister Strike 2, Precision 3, Flawless Execution, Dual Wield Specialization 5, Blade Flurry, Hack and Slash 5, Weapon Expertise 2, Aggression 3, Adrenaline Rush.
- Subtlety: Opportunity 2, Initiative 3, Ghostly Strike, Premeditation, Serrated Blades 3, Preparation, Hemorrhage. It has no Quietus, Cutthroat or Thousand Cuts.
- No preset takes Venom, Vigor or Improved Expose Armor.

Sources:
- dev notes (DN) are https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696, posts /1 (70009), /4 (70170), /5 (70291) and /6 (70425).
- The deep dive (DD) is https://worldofwarcraft.blizzard.com/en-us/news/24310968, 2026-10-15, which describes the 70425 design.
- fc is foreverchanges.pro/patch-notes.

#### 1. Announced changes

| ID | Source (link, date, build) | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| ROG-70425-01 | DN /6, 2026-10-15, 70425 | Rogue poisons no longer activate against targets that are attacking the rogue. | n/a | server side (no poison row changed) | n/a | Poisons proc only from the rogue's own hits (OnSpellHitDealt, sim/rogue/poisons.go:80-88, 105-112, 129-137). The sim never had the bug. |
| ROG-70425-02 | DN /6, 70425 | Sap no longer activates weapon procs. | n/a | SpellMisc 316193/319142/324065 (Sap 2070/6770/11297): Attributes_3 1 → 196609 (adds 0x10000 can't-trigger-proc and 0x20000), Attributes_4 0 → 8388608, Attributes_6 4 → 2097156 | n/a | Sap is not simulated. |
| ROG-70425-03 | DN /6, 70425 | Swapping specializations no longer keeps Energy full. | n/a | server side | n/a | none |
| ROG-70425-04 | DN /6, 70425 (fc: "Cold Blood is spent when Mutilate crits") | Assassination: Cold Blood is now consumed when Mutilate lands a critical strike. | yes (the Mutilate presets take Cold Blood) | server side. 14177 is unchanged: aura 107, +100 crit, ProcCharges 1, ProcTypeMask 87376, 3 min cooldown. Its mask_1 bit 2 covers both Mutilate halves (1310705/1310706, SpellClassMask_1 = 2). | landed | sim/rogue/talents.go:122-126: the aura deactivates on the first Cold Blooded hit, which is Mutilate's main hand half (mutilate.go:76-78). The Cold Blood cooldown waits for a 5-point Eviscerate (talents.go:149-151), so in the presets Cold Blood rarely lands on a Mutilate. See DD-13 for the off-hand half. |
| ROG-70425-05 | DN /6, 70425 (Races, Undead) | Touch of the Dead: 20% chance for Paladins, Rogues and Warriors, a 25 sec internal cooldown, no scaling with Shadow modifiers, no reduction for low-rank spells. | yes (all 12 presets are Undead) | SpellAuraOptions 243729 (1260189): ProcChance 5 → 20, ProcCategoryRecovery 1000 → 25000 | landed | sim/core/racials_forever.go:319-385: procChance 0.20 (:321), 25 sec ICD (:350), and the roll needs a damage component (:374). The drain is a plain Shadow spell with DamageMultiplier 1 that ignores spell rank. PR #50. |
| ROG-70425-06 | DN /6, 70425 (Races) | The other racials in the rogue race table: Human Ruthless (2% crit, any weapon), Dwarf Earth Fury (+2% crit damage), Orc Savage (1% crit) and Blood Fury (3 min), Tauren Endurance (crit damage instead of hit), Troll Regeneration, Skyborne Wind Blessed (3%), Gnome Eureka! (8 sec, no charges), Night Elf Quickness. | minor (race table only, every preset is Undead) | Racial section of diffs/70374-70425: Ruthless 20597 loses SpellEquippedItems 255480; Savage 20574 loses 255488; Earth Fury 1259719 aura 290 → 163, bp 1 → 2; Endurance 20550 aura 54 → 163, bp 1 → 2; Eureka 1259812 DurationIndex 8 → 31 (8 sec) and its SpellAuraOptions row (3 charges) removed; Blood Fury 20572 cooldown 120000 → 180000 | landed | Eureka! and Blood Fury came in with PR #50 (racials_forever.go:209-270), Savage with #53. Human, Dwarf, Tauren and Skyborne came in with PR #56 (b94432154), merged to origin/master 4f5c608 after the 6a87c9c checkout. 6a87c9c still has the sword-only Human crit (:25), the Tauren hit (:64) and Skyborne at 1.01 (:78). |
| ROG-70425-07 | DN /6, 70425 (Shaman) | Grace of Air Totem is no longer exclusive with Windfury Totem. Grace of Air, Strength of Earth and Windfury are now raid-wide. | yes (every rogue preset runs on the raid's totems) | Shaman scope | ambiguous (needs verify) | The raid-wide totems came in with PR #50. The rogue requests (like every melee request) have totemWeaponBuff Windfury and strengthOfEarthTotem, but no graceOfAirTotem. Now that the two stack, whether the reference raid adds Grace of Air (+77 Agility at rank 2) is a raid composition question for standards 23. See §3, F5. |
| ROG-70425-08 | DN /6, 70425 | Wands need 0.5 sec standing still, scrolls trigger the GCD, Eureka! can't be used in Stealth, other General/UI/PvP lines. | n/a | none | n/a | none |
| ROG-70291-01 | DN /5, 2026-10-08, 70291 | No rogue class line. The Rogue quest map markers, critical strikes in PvP, weapon swaps within a GCD, combo point art and the Rage fixes are the only lines that touch a rogue. | n/a | Rogue section of diffs/70170-70291 = Expose Armor aura and PvP helper spells, see U-02 and U-08 | n/a | none |
| ROG-70170-01 | DN /4, 2026-10-01, 70170 | A missed Expose Armor no longer loses all combo points. | minor (no preset casts it; the raid's debuff covers it) | SpellMisc 320813/321312/321534/323543/322540 (8647/8649/8650/11197/11198): Attributes_1 138412544 → 134218240, Attributes_4 8 → 1048584 | landed | sim/rogue/expose_armor.go:188-198: SpendComboPoints only runs when result.Landed(), and a miss refunds. |
| ROG-70170-02 | DN /4, 70170 | Subtlety: Setup only adds combo points to the current target when that target is the one whose attack was dodged. | n/a | SpellEffect 692439 (13983) aura 42 → 4 (dummy, server-scripted) | n/a | Setup is not simulated (it needs incoming attacks). |
| ROG-70170-03 | DN /4, 70170 | The Cooldown Manager supports the rogue; new combo point art. | n/a | none | n/a | none |
| ROG-70170-04 | DN /4, 70170 (Races, Gnome) | Eureka! no longer benefits periodic effects at all. | minor (Gnome rogue in the race table only) | SpellEffect 1269066/1269067 (1259812): EffectSpellClassMask_0 68290334 → 100794910, mask_1 → 2097408. The mask covers Sinister Strike, Backstab, Gouge, Ambush, Eviscerate, Hemorrhage, Ghostly Strike and Mutilate. It leaves out Rupture (1048576), Garrote (256), Expose Armor, Slice and Dice and Venom. | wrong | sim/core/racials_forever.go:225-233. `affects` takes any APL spell with a cost and a non-zero DamageMultiplier, so Rupture (rupture.go:43) and Garrote (garrote.go:47) get +10% damage and the cost cut. That was G1 in orchestration/forever_class_changes.md and it is still open. |
| ROG-70009-01 | DN /1, 2026-09-24, 70009 | Sap now flags the rogue for PvP. | n/a | Sap SpellMisc Attributes_3 0 → 1 | n/a | none |
| ROG-70009-02 | DN /1, 70009 (Races, Gnome) | Eureka! is a 10% discount for every class (the rogue's was 20% Energy). | minor | SpellEffect 1269066 EffectBasePointsF -20 → -10 | landed | racials_forever.go:222 costReduction 10 |
| ROG-70009-03 | DN /1, 70009 (Races, Undead) | Touch of the Grave no longer breaks CC or activates from abilities that have no damage component. | yes (Undead presets) | server side | landed | racials_forever.go:374 (`spell.DamageMultiplier != 0`). An Expose Armor or a Slice and Dice doesn't roll. |
| DD-01 | DD, 2026-10-15 (70425 design) | Poisons have more charges. | yes | 11340 Instant Poison VI eff1 bp 175, 11356 Deadly Poison IV 160, 25347 Deadly Poison V 180, 13219 Wound Poison 90. No poison row changed during the beta. | not simulated on purpose | poisons.go:20 `TODO: Add charges`: poisons never run out, which matches a 180 sec fight with poisons reapplied before the pull. |
| DD-02 | DD | Poisons go on a weapon together with a temporary enchant (stones, oils). | yes (+9.5 to +13 DPS per orchestration R1) | Poisons use SpellEffect 360 (class imbue), and stones and oils keep 54 | landed | rogue.go:232-247 Poisons() reads RogueOptions main_hand_poison/off_hand_poison and leaves Consumes for the Dense Sharpening Stone. The presets carry both. |
| DD-03 | DD | Feint's threat reduction is doubled. | n/a | none | n/a | feint.go |
| DD-04 | DD | Sap no longer breaks Stealth. Distract lowers Stealth detection. | n/a | Sap/Distract Attributes_6 bit 0x4 | n/a | none |
| DD-05 | DD | Expose Armor matches 5 Sunder Armors at the same level, and Sunder Armor can be used on a target with Expose Armor without overwriting it unless it is weaker. | minor (the raid debuff; no preset casts Expose Armor) | 11597 Sunder bp -450 × 5 = 2250. 11198 Expose Armor bp 0 (the value is server side). Aura 22 → 674 on Expose Armor in 70291 and on Sunder in 70425 (see U-02). | landed | sim/core/debuffs.go:831-857 exposeArmorAura(2250), expose_armor.go:132-134 (450 per combo point at rank 5). Both share the MajorArmorReduction exclusive category with priority = armor, so a weaker Sunder doesn't replace it. |
| DD-06 | DD | Rogues can equip one-handed axes. | minor | SkillLineAbility (per forever_class_changes line 43) | landed | hack_and_slash.go:23-25 treats axes like swords. The swords preset stays on swords (standards 16). |
| DD-07 | DD | Improved Gouge moved to Assassination row 1. | n/a | the tree is the same in every build (see §2) | n/a | none |
| DD-08 | DD | Remorseless Attacks now also affects Mutilate. | n/a (it procs on kills; one target) | 14144 | n/a | not implemented, as there are no kills in a single-target sim |
| DD-09 | DD | Malice raises crit chance with all attacks and poisons. | yes (Mutilate presets 5/5) | 14138 aura 290 (all crit) bp 5 at 5/5 | landed | talents.go:25, 28: +1% melee and +1% spell crit per rank. Poisons roll spell crit. |
| DD-10 | DD | Murder also covers Giants, and no longer Beasts or Dragonkin. | yes (the -humanoid presets, 2/2) | 14158 aura 168 bp 4 at 2/2 | landed | talents.go:61-81: ×1.02/1.04 against Humanoid and Giant, with no crit damage part |
| DD-11 | DD | Improved Expose Armor redesigned: an Energy discount and a combo point refund at 5 points. | minor (no preset takes it) | 14168 eff0 aura 107 bp -10 (Energy), eff1 aura 231 triggers 1310697 (+1 combo point). 70009 added its proc row (ProcChance 100, ProcCategoryRecovery 500), which is the unannounced U-03. | landed | expose_armor.go:138 (25 - 5/rank), :193-195 (refunds ImprovedExposeArmor combo points at 5) |
| DD-12 | DD | Lethality now also applies to Mutilate, and its crit damage bonus is lower. | yes (Mutilate 3/5, Seal Fate 5/5) | 14128 aura 108 misc 15 covers Sinister Strike, Backstab, Hemorrhage, Ghostly Strike and both Mutilate halves. Its bp 6 is the base spell value; the ranks come from curves that were not downloaded. The Forever tooltips (hyjal 69876, forever-wiki) read 4/8/12/16/20%. | landed | talents.go:485-488 0.04/rank on Sinister Strike, Backstab, Hemorrhage, Ghostly Strike and both Mutilate halves (mutilate.go:27, 67) |
| DD-13 | DD | Cold Blood moved to the 16-point medal and now also applies to Mutilate. | yes (Mutilate presets) | 14177 mask_1 = 6 covers both Mutilate halves 1310705 and 1310706 (mask_1 bit 2) | partial | Only Mutilate's main hand half gets +100% crit. mutilateOH (mutilate.go:17-37) lacks SpellFlagColdBlooded, and Cold Blood is spent on the main hand hit before the off hand swings (mutilate.go:76-79). Close to no DPS impact today, because Cold Blood is held for Eviscerate (talents.go:149-151). |
| DD-14 | DD | Improved Poisons: also a large chance for a poison proc not to use a charge. | yes (Mutilate 5/5, Seal Fate 2/5) | 14113 eff0 aura 107 misc 18 bp 10 at 5/5 (proc chance); eff1 dummy bp 50 (charge save) | landed | poisons.go:40-42 +2%/rank proc chance. The charge part doesn't matter because charges are not simulated (DD-01). |
| DD-15 | DD | Vigor is a 2-point row 5 talent with the same total. | minor (no preset takes it) | 14983 aura 35 bp 10 at 2/2 | landed | rogue.go:171 maxEnergy 100 + 5/rank |
| DD-16 | DD | Mutilate: 60 Energy, strikes with both daggers, does more damage to poisoned targets, awards 2 combo points, and each hand procs on its own. | yes | 1310707 cost 60, eff0 2 combo points, eff3 bp 20 (+20% poisoned). Halves 1310705/1310706 (bp 23) and trainer ranks up to 1241586 (bp 67 at 60) do 75% weapon damage. | landed | mutilate.go:48-50 (60 Energy), :82 (2 combo points), :95-101 (50.25 + 0.75 × weapon, ×1.2 under Deadly or Wound Poison). The halves are separate MH and OH proc-mask spells. Only a main hand dagger is required (:58-63, notes.md NTV). |
| DD-17 | DD | Improved Kidney Shot's damage bonus only applies to the rogue's own attacks and poisons. | n/a | 14174 | n/a | Kidney Shot is not in any rotation |
| DD-18 | DD | Seal Fate no longer requires Cold Blood. "Mutilate will trigger this exactly once if either or both hands generates a critical strike." | yes (Mutilate 3/5) | 14186 ProcChance 100 at 5/5, ProcCategoryRecovery 500 (20%/rank from the curve) | partial | talents.go:167-201 rolls on each Mutilate half. Both halves carry SpellFlagBuilder (mutilate.go:22, 45), and the 500 ms ICD is only used on a successful roll (:195-198). When both hands crit and the first roll fails, the off hand rolls again: 1-(1-p)² instead of p. See F1. |
| DD-19 | DD | Venom: new 31-point 25 Energy finisher. +30% poison damage, +10% chance to apply poisons, duration set by combo points. | minor (no preset takes it) | 1310703 eff1/2 aura 108 bp 30 (direct and periodic poisons), eff3 aura 107 misc 18 bp 10, cost 25, duration 6 sec + 3 sec per combo point | landed | venom.go:17-24 (9-21 sec), :32-37 (+0.1 chance, ×1.3), :49 (25 Energy). It still uses Envenom's ID 32645 (:15) as a placeholder; the client now has 1310703. |
| DD-20 | DD | Improved Eviscerate moved to Combat row 1, with a bigger damage bonus. | yes (every non-Subtlety preset 3/3) | 14162 aura 108 bp 20 at 3/3 | landed | eviscerate.go:59 7/13/20% |
| DD-21 | DD | Puncturing Wounds (formerly Improved Backstab): also raises Mutilate's crit chance by less, and gives Backstab a chance at an extra combo point. | yes (Combat Daggers, Mutilate 3/3) | 1224716 bp 30 (Backstab crit), 45 (extra combo point via 1310710), 15 (Mutilate crit) at 3/3 | landed | backstab.go:115 (15%/rank), :143 (10%/rank); mutilate.go:25, 65 (5%/rank) |
| DD-22 | DD | Deflection gives twice the Parry. | n/a | none (fc's 70170 "Deflection 1-5%" line is the Hunter talent) | n/a | talents.go:24 |
| DD-23 | DD | Precision raises hit with all spells and attacks, poisons included. | yes (Combat and Seal Fate 3/3) | 13705 aura 54 + 55 bp 1 per rank | landed | talents.go:26, 29: +1% melee and +1% spell hit per rank |
| DD-24 | DD | Endurance cuts the Sprint and Evasion cooldowns by a percentage. | n/a | 13742 bp -60 at 2/2 | n/a | evasion.go:33 |
| DD-25 | DD | Flawless Execution: new 16-point medal, Eviscerate costs 10 less Energy. | yes (every Combat build) | 1310711 aura 107 bp -10, Eviscerate mask | landed | eviscerate.go:41 (35 - 10) |
| DD-26 | DD | Dual Wield Specialization raises off-hand damage by less. | yes (Combat 5/5, Seal Fate Swords 4/5) | 13715 aura 122 bp 50 (base spell; ranks are curve-driven). The Forever tooltips read 5/10/15/20/25%. | landed | talents.go:42-44 1 + 0.05/rank. Not ambiguous, but see NTV-5. |
| DD-27 | DD | Hack and Slash: the four weapon specializations in one talent. Swords and axes get extra attacks, maces ignore armor instead of stunning, daggers and fists get crit. | yes | 13960 bp 5 (crit) / 15 (mace armor) / 5 (extra attack) at 5/5, ProcCategoryRecovery 200 | landed | hack_and_slash.go:24 (1%/rank extra attack), :30-44 (1%/rank crit), :49 (3%/rank armor), :56 (200 ms ICD) |
| DD-28 | DD | Weapon Expertise redesigned: lowers the target's chance to dodge or parry. | yes (Combat 2/2) | 30919 aura 240 (expertise) bp 2 at 2/2 | landed | talents.go:233-239 (1% expertise per rank) |
| DD-29 | DD | Aggression now also raises Backstab damage. | yes (Combat 3/3) | 18427 aura 108 mask 131078 (Sinister Strike, Backstab, Eviscerate) bp 2 per rank | landed | sinister_strike.go:39, backstab.go:26, eviscerate.go:60 (2/4/6%, multiplied with the other modifiers per notes.md) |
| DD-30 | DD | Camouflage, Master of Deception, Dirty Tricks, Improved Distract, Heightened Senses, Setup (3 points = 100%). | n/a | none | n/a | none |
| DD-31 | DD | Opportunity now also affects Mutilate. | yes (Subtlety 2/2) | 14057 mask 516 (Backstab, Ambush) + mask_1 6 (Mutilate), eff1 Garrote periodic, bp 10 at 2/2 | landed | backstab.go:25, ambush.go:21, mutilate.go:29, 69, garrote.go:48 (5%/rank) |
| DD-32 | DD | Improved Ambush moved to row 2. | minor (Subtlety 3/3, but no Stealth opener per the user's ruling and no Cutthroat, so Ambush is never cast) | 14079 aura 107 misc 7 bp 45 at 3/3 | landed | ambush.go:49 (15%/rank crit) |
| DD-33 | DD | Initiative: 100% at 3 points. | minor (Ambush and Garrote not cast) | 13976 bp 100 at 3/3 | landed | talents.go:210 33/67/100% |
| DD-34 | DD | Ghostly Strike does more damage with a dagger. | yes (Subtlety) | 14278 eff0 E31 bp 125, dummy 180 (dagger), 40 Energy, 20 sec cooldown, 7 sec | landed | ghostly_strike.go:37 (40 Energy), :47 (20 sec), :54 (1.8 with a dagger, else 1.25) |
| DD-35 | DD | Premeditation moved to the 16-point medal, duration 20 sec. | yes (Subtlety) | 14183: 2 combo points, duration 20000, cooldown 120000 | landed | premeditation.go:30 (2 min), :34 (2 points, no Stealth needed). The 20 sec expiry of the points is not modelled, but the rotation spends them well inside 20 sec. |
| DD-36 | DD | Serrated Blades ignores a percentage of armor that rises with rank, instead of a flat value. | yes (Subtlety 3/3) | 14171 eff0 aura 280 bp 9 at 3/3; eff1 aura 108 Rupture bp 30 | landed | talents.go:243-249 (3%/rank armor), rupture.go:43 (Rupture ×1.1/1.2/1.3) |
| DD-37 | DD | Dirty Deeds lets Garrote be used from the front. | n/a | none | n/a | garrote.go:30, 44 (Garrote isn't cast) |
| DD-38 | DD | Hemorrhage does more damage with a dagger, and raises the rogue's Rupture damage on the target for 15 sec. | yes (Subtlety) | 16511 E31 bp 100, dummy 145 (dagger), aura 271 bp 15 on the Rupture mask, 15 sec, 35 Energy | landed | hemorrhage.go:12 (Rupture ×1.15), :27 (15 sec), :53 (1.45 with a dagger); rupture.go snapshot. Cold Blood is left off, matching the client's mask. |
| DD-39 | DD | Quietus (new, row 6): Sinister Strike, Ghostly Strike and Hemorrhage do more damage below 35% health. | minor (no preset takes it) | 1310728 bp 10 at 5/5, bp 35 (health threshold) | landed | talents.go:333-342 (2%/rank under 35%) |
| DD-40 | DD | Cutthroat (new, row 6): Backstab can let the next Ambush skip Stealth. | minor (no preset takes it) | 462708 bp 15 at 5/5 | landed | talents.go:252-284 (3%/rank, 10 sec proc). Placeholder aura ID 460534. |
| DD-41 | DD | Thousand Cuts (new, 31 points): Rupture ticks cut the next Hemorrhage or Backstab's cost by 3 Energy, stacking 5 times. | minor (no preset takes it) | 1310721 ProcCategoryRecovery 1900, triggers 1310723 (bp -3, 5 stacks, 10 sec, Backstab and Hemorrhage mask) | landed | talents.go:287-330. The 1.9 sec ICD isn't coded, but Rupture ticks every 2 sec so it can't matter. |
| DD-42 | DD | Removed: Dagger, Mace, Sword and Fist Weapon Specialization, Sleight of Hand, Deadliness. | yes | trait tree 1111 has 53 nodes and none of them | landed | proto/rogue.proto has no fields for them; TalentTreeSizes {17,17,19} (rogue.go:40) |

#### 2. Unannounced client changes (reverse check)

Rogue talent tree (TraitTree 1111, 53 nodes) checked by script across 69893, 70009, 70170, 70291, 70334, 70338, 70374 and 70425: no node, entry, rank count or position changed in any build. Every talent statement in the deep dive was already in the client the beta opened on.

The client-only builds 70291 → 70334 → 70338 → 70374 have no rogue spell and no rogue trait row in their diffs. That confirms nothing changed for the rogue. 70058, 70124, 70205, 70235 and 70245 have no CSVs; their windows fall inside 70009 → 70170 and 70170 → 70291, and every rogue row in those diffs is accounted for below or in §1.

| ID | Source (link, date, build) | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| U-01 | diffs/70374-70425 (Rogue), 70425 | New spell Death Mark, 4 ranks (1316709 L40, 1325756 L48, 1325757 L56, 1325758 L60). A finisher that stores a share of the rogue's damage per combo point and deals it after 6 sec, capped by a flat amount + AP + crit (rank 4: 156 + 245/CP + AP term). 25 Energy, 10 sec cooldown. In the same build, Relentless Strikes 14179 mask_2 changed 67108864 → 67108868 and Eureka! 1259812 mask_2 0 → 4, so both now cover it. | minor today (can't be learned) | No SkillLineAbility row, no TraitDefinition, no ItemEffect in 70425, so no rogue can learn it yet. Not in any note, deep dive or fc. | ambiguous (needs verify) | Nothing in sim/rogue. Watch for it to become trainable; see NTV-1. |
| U-02 | diffs/70170-70291, 70291 (and 70425 for Sunder) | Expose Armor (all 5 ranks) aura 22 (mod resistance) → 674. Sunder Armor 7386/11597, Stoneform and Sebacious Poison moved to 674 in 70425. | minor | SpellEffect 686639/687242/687526/688758/689964 EffectAura 22 → 674; 11597 bp -450 unchanged | landed | This is the client side of DD-05 (a new armor aura that lets Sunder coexist with Expose Armor). The sim already models it as one exclusive category with priority (debuffs.go:798-857). |
| U-03 | diffs/69893-70009, 70009 | Improved Expose Armor 14168 eff1 aura 4 → 231 with trigger 1310697, and a new SpellAuraOptions 248365 (ProcChance 100, ProcCategoryRecovery 500, ProcTypeMask 4112). This is the combo point refund. | minor | 1310697 E30 bp 1 (+1 combo point per proc) | landed | expose_armor.go:193-195 |
| U-04 | diffs/70009-70170 and 70374-70425 | Premeditation 14183: Attributes_3 0 → 131072 (0x20000) in 70170, Attributes_6 0 → 4 in 70425. Sap and Distract already carry bit 4, which goes with not breaking Stealth. | n/a | flags only; duration, cooldown and points are unchanged | n/a | No sim change. Premeditation is already usable outside Stealth (premeditation.go:12-14). |
| U-05 | diffs/70374-70425, 70425 | Attributes_6 0 → 2097152 (bit 21) on Rupture (all 6 ranks), Garrote (all 6), Kick, Gouge, Kidney Shot, Cheap Shot, Blind, Riposte, Crippling Poison, Silenced, Between the Eyes, Saber Slash, Quick Draw, Waylay, and Paladin Rebuke and the racials War Stomp, Stoneform, Will to Survive. | ambiguous (Rupture and Garrote are DPS spells) | SpellMisc rows 316047…321837 (Rupture), 314449…323407 (Garrote) | ambiguous (needs verify) | Mostly crowd control and interrupts, so probably a PvP or diminishing returns flag. It would only matter if it touches Rupture's damage; the sim keeps Rupture as is. See NTV-4. |
| U-06 | diffs/70374-70425 (Rogue), 70425 | The rogue's Eureka! 1259812 mask_2 0 → 4 (now covers Death Mark), and the 8 sec / no-charge rework (see ROG-70425-06). The 70170 mask change that took periodic spells out is ROG-70170-04. | minor | as ROG-70170-04 | wrong | Same root cause as ROG-70170-04: racials_forever.go:225-233 ignores the class mask, so Rupture and Garrote are boosted. One fix covers both rows. |
| U-07 | diffs/69893-70009, 70009 | Improved Kick 24434, Shadowflame Fury 461253 and Sprint Duration Increase 23049 EffectBonusCoefficient 1 → 0 | n/a | flags only | n/a | none |
| U-08 | diffs/70170-70291, 70374-70425 | New PvP helper auras (Blind 1324477, Gouge 1324481, Kidney Shot 1324474/5, Sap 1324482-4, Silenced 1324551, Silenced - Kick 1324485, Counterspell - Silenced 1324401), Stealth 1326175, Distract Attributes_6 4 → 2097156, Undo Crowd Control Rogue 22891 | n/a | server side | n/a | none |
| U-09 | diffs/70374-70425 | Season of Discovery leftovers: Crimson Tempest Attributes_6 8388608 → 10485760, Sebacious Poison aura 22 → 674, plus U-05's bit on Between the Eyes, Saber Slash, Quick Draw and Waylay | n/a | not learnable in Forever | n/a | none |
| U-10 | diffs/70009-70170 (No class) | "Precision" 1306583 bp 5 → 1 is the staff enchant Imbue Precision, not the rogue talent 13705 | n/a | 1306583 | n/a | none (here so nobody mistakes it for the rogue talent) |

Tooltip-only rogue rows were skipped: 70170's Expose Armor, Venom, Main Gauche, Restless Blades, Rupture 1271514, Stealth, Shadowmeld, Relentless Strikes Effect, Tier 1 3P and Improved Pick Pocket wording, and 70291's Vanish wording. The Racial rows other than Touch of the Grave and Eureka! are in ROG-70425-06.

#### 3. Proposed fixes

**F1. Seal Fate rolls once per Mutilate (DD-18, partial)**
- Change: in `applySealFate` (sim/rogue/talents.go:167-201), give the aura a `lastMutilateRoll time.Duration = -1`.
- In OnSpellHitDealt, when the spell is `rogue.Mutilate` or `rogue.mutilateOH` and `lastMutilateRoll == sim.CurrentTime`, return. Otherwise set `lastMutilateRoll = sim.CurrentTime` before the crit roll.
- Both halves land at the same timestamp (mutilate.go:78-79), so either hand's crit triggers at most one roll.
- An alternative: drop SpellFlagBuilder from mutilateOH and roll Seal Fate once inside Mutilate's ApplyEffects from `mhResult.DidCrit() || ohResult.DidCrit()`.
- Test: `TestSealFateOncePerMutilate` in sim/rogue. Use Seal Fate 1/5 (20%) and +100% crit on both halves (BonusCritRating), run N Mutilates, and assert the Seal Fate combo point rate is about 0.20, not 0.36.
- Impact: about -0.5 to -1% (3 to 7 DPS). With about 45% crit per half, both halves crit about 20% of the time, and the second roll adds 0.2 × 0.6 × 0.4 ≈ 0.05 extra combo points per Mutilate at Seal Fate 3/5.
- Presets: rogue__mutilate, rogue__mutilate-humanoid. The golden file is sim/rogue/dps_rogue/TestAssassinationMutilate.results.

**F2. Cold Blood on both Mutilate halves (DD-13, partial)**
- Change: add SpellFlagColdBlooded to mutilateOH (mutilate.go:22).
- In the Cold Blood aura (talents.go:122-126), don't deactivate on `rogue.Mutilate` (the main hand half). Let mutilateOH's hit spend it, so it is spent once, after both halves.
- That also matches the 70425 note that a Mutilate crit consumes Cold Blood.
- Test: with Cold Blood active, one Mutilate gives two crits and Cold Blood is gone afterwards.
- Impact: about 0 for the current presets, because Cold Blood is held for a 5-point Eviscerate (talents.go:149-151). It matters only if a rotation spends Cold Blood on Mutilate.
- Presets: rogue__mutilate*. A rotopt pass could test Cold Blood on Mutilate against Eviscerate after the fix.

**F3. Eureka! respects the class spell mask (ROG-70170-04, U-06, wrong)**
- Change: in sim/core/racials_forever.go:225-233, leave out spells whose damage is only periodic.
- For the rogue, either exclude SpellCode_RogueRupture and SpellCode_RogueGarrote, or add a generic `spell.Dot != nil && spell has no direct damage` test.
- The client's rogue mask covers only Sinister Strike, Backstab, Gouge, Ambush, Eviscerate, Hemorrhage, Ghostly Strike, Mutilate and (now) Death Mark. The cost cut should follow the same mask.
- This file is shared: route it through the orchestrator together with the warlock, priest and warrior parts of G1.
- Test: a Gnome rogue's Rupture tick is the same with and without Eureka! up.
- Impact: Gnome rogues in the race table only, a fraction of a DPS. No preset is affected.

**F4. Cosmetic and stale comments (no DPS change)**
- Venom's action ID 32645 → 1310703 (venom.go:15). Remove its "TODO: tooltip showed no Energy cost"; the client cost is 25.
- rogue.go:130-135 still says Eureka! "buffs the next 3 abilities". Since 70425 it is an 8 sec window with no charges. The hold for an Eviscerate within 8 sec is still sensible.

**F5. Grace of Air in the reference raid (ROG-70425-07, cross-scope)**
- In 70425, Grace of Air Totem stacks with Windfury Totem and both are raid-wide.
- The rogue requests (and every other melee request; only the hunter presets set graceOfAirTotem) have Windfury and Strength of Earth but no Grace of Air.
- Whether the shared raid should add it is a raid composition ruling (standards 23, "Shaman: no totem duty"), not a rogue fix.
- If it is added, expect roughly +2.5 to 3% for the rogues (77 Agility ≈ 77 AP + about 1.3% crit). spec_assumptions.md:109 put Grace of Air at about 21 DPS when it was the alternative to Windfury.
- Presets: all 12.

Nothing is missing outright. No 70425 rogue change needs a preset or gear update beyond F1, F2 and F5.

#### 4. Need to Verify candidates

**NTV-1. Death Mark (new rogue finisher in client 70425).**
- Client 70425 adds Death Mark in four ranks (1316709 at 40, 1325756 at 48, 1325757 at 56, 1325758 at 60). It is a 25 Energy finisher with a 10 sec cooldown that marks the target for 6 sec, stores a share of the rogue's damage per combo point, and then deals it, capped at (rank 4) 156 + 245 per point plus an AP and crit term.
- Relentless Strikes and Eureka! were extended to cover it in the same build.
- No trainer (SkillLineAbility), talent or item teaches it yet, and Blizzard's notes, the deep dive and foreverchanges don't mention it.
- Settles it: the trainer list or spellbook of a rogue of level 40 or more on the next build, or the next patch notes. If it ships, it is a new finisher that every rogue rotation must weigh against Eviscerate and Rupture. The store ratio and the AP term are in the tooltip variables ($<ratio>, $<maxAP>, $<attackP>), which need the SpellVariable table we didn't download.

**NTV-2. Seal Fate on Mutilate: one roll, or one per hand?**
- The deep dive says Mutilate "will trigger [Seal Fate] exactly once if either or both hands generates a critical strike". The client gives Seal Fate a 500 ms ProcCategoryRecovery, which only says it can't fire twice.
- The sim rolls on each half until one succeeds, so with both hands critting the chance is 1-(1-p)² (F1).
- Settles it: on the beta, with Seal Fate 1/5 (20%) and Cold Blood (+100% crit on both halves, if NTV-3 holds), count Seal Fate combo points over about 50 Cold Blooded Mutilates. About 20% means one roll and about 36% means one per hand. Mutilate is a 21-point talent, so this needs a level 40 or higher rogue.

**NTV-3. Cold Blood and Mutilate's two halves.**
- Cold Blood's client mask (14177, SpellClassMask_1 = 6) covers both Mutilate halves (1310705 and 1310706, mask_1 = 2), so both should get +100% crit. Notes 70425 says it is "consumed when Mutilate lands a critical strike".
- The sim gives the bonus to the main hand half only and spends it there.
- Open: (1) do both halves crit under Cold Blood? (2) Is Cold Blood kept if Mutilate is dodged or misses?
- Settles it: Cold Blood → Mutilate on a training dummy. Two crits in the log means both halves count. Then repeat against a mob that dodges, and check whether the buff stays.

**NTV-4. The Attributes_6 bit 21 added to Rupture and Garrote in 70425.**
- Client 70425 sets SpellMisc Attributes_6 bit 21 (2097152) on every rank of Rupture and Garrote, alongside the rogue's interrupts and crowd control (Kick, Gouge, Kidney Shot, Cheap Shot, Blind), Riposte, Crippling Poison, Paladin Rebuke and the racials War Stomp, Stoneform and Will to Survive.
- The set looks like a PvP or diminishing returns grouping, and no patch note names it.
- If it changes how the bleeds roll or scale, Rupture would be off for the Subtlety and Seal Fate presets.
- Settles it: Rupture's 5-point ticks at 60 on the next build compared with the same character's ticks before 70425. Also look for a new PvP note about rogue stuns and bleeds.

**NTV-5. Talent rank values that come only from curves (Dual Wield Specialization, Lethality).**
- The base spells still hold Classic numbers: Dual Wield Specialization 13715 bp 50 (aura 122) and Lethality 14128 bp 6.
- The ranks come from TraitDefinitionEffectPoints curves (not downloaded). The Forever tooltips (hyjal 69876, forever-wiki) read 5% per rank for Dual Wield Specialization (25% at 5/5) and 4% per rank for Lethality, and the sim uses those. The deep dive agrees on direction for both ("increase to off-hand damage reduced", "bonus to critical damage also lowered some").
- Not moved by any build since 69893. Dual Wield Specialization is worth about 3% to the Combat and Seal Fate Swords presets if the tooltip were wrong.
- Settles it: the rogue's off-hand white hit range on the character sheet (or log) at Dual Wield Specialization 0 and 5. ×1.25 means the tooltip is right; ×1.5 means it is Classic's.

#### 5. Count

68 rows (§1: 58, §2: 10): landed 38, partial 2, wrong 2, missing 0, not simulated on purpose 1, in flight (worker PR) 0, ambiguous (needs verify) 3, n/a 22.

### Priest: game-update audit ledger (builds 70009 to 70425)

Scope: sim/priest (+ the priest parts of shared racials that every priest preset uses), presets
tools/dps_report/req/priest__shadow.json, priest__shadow-holy.json, priest__smite.json (all Undead,
level 60). Master checked at 6a87c9c. Worker branch `origin/claude/wow-forever-dps-presets-m9u15h-priest`
(abbreviated **PB** below; commits 86b297cce, e63451452, 8e3267b2b, bd88453a9, 0d248b50c) holds the
70425 port. Client values read with python from `S/db2/*.csv` (S = the scratchpad).

Links: P1 = https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/1
(2026-09-24, 70009), P4 = .../2360696/4 (2026-10-01, 70170), P5 = .../2360696/5 (2026-10-08, 70291),
P6 = .../2360696/6 (2026-10-15, 70425), DD = https://worldofwarcraft.blizzard.com/en-us/news/24301514
(Priest+Warrior deep dive, design at about 70170).

#### 1. Announced changes

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| PRI-70425-01 | P6, 10-15, 70425 | Devouring Plague cooldown 24 s (was 1 min) | yes (all 3 presets) | SpellCooldowns 2944..19280 CategoryRecoveryTime 60000 -> 24000. Use 24 s | in flight (worker PR) | master devouring_plague.go:22 `time.Minute`; PB `time.Second * 24`, test TestCooldowns70425. PB meta.json rotation text still says "off its 1 min cooldown" (wording only) |
| PRI-70425-02 | P6 | Shadow Word: Death cooldown 18 s ("was 12") | yes (shadow, shadow-holy; smite only at fight end) | SpellCooldowns 1309595/633/635/636 CategoryRecoveryTime 15000 -> 18000. Use 18 s (old value was 15, not 12, see NtV-1) | in flight (worker PR) | master shadow_word_death.go:59 `time.Second * 15`; PB 18 s |
| PRI-70425-03 | P6 | Power in Light now also works if the target has Shadow Word: Pain | yes (smite) | 1309969 tooltip "...your Holy Fire or Shadow Word: Pain"; effect is a dummy (aura 4), server side | **missing (missed by worker)** | penance.go:103-112 `powerInLightMultiplier` loops only `priest.HolyFire`; PB unchanged. notes.md:126 also still says Holy Fire only |
| PRI-70425-04 | P6 | Twin Disciplines now benefits SW:Death, Mind Flay, Starshards, Penance | yes (shadow takes 0/5 on PB; smite 5/5) | 1225132 eff0 (direct) mask_1 16 -> 98322 (+SWD 0x2, +Penance bolt 0x8000), eff1 (DoT, misc 22) mask_0 32864 -> 10518624 (+MF 0x800000, +Starshards 0x200000); 5% at 5/5 | in flight (worker PR) | master talents.go:69-80 only SW:P and Chastise (also missed DP, which 70338/70374 already had via mask_3 0x10); PB adds SWD, Chastise, Penance, SWP, MF, Starshards, DP. Values checked against masks |
| PRI-70425-05 | P6 | Shadow Word: Death benefits from Shadowform's crit damage bonus | yes (shadow, shadow-holy) | 15473 eff4 (misc 15 crit dmg, +100) mask_1 0 -> 2 (SWD) | in flight (worker PR) | master talents.go:346-347 lacks SWD; PB adds it, TestShadowformCritBonus checks 2.0 |
| PRI-70425-06 | P6 | Mental Strength moved to row 2, +2/4/6/8/10% Intellect | yes (shadow takes it) | 18551 aura 137 bp 15 unchanged; rank values are a trait curve (not downloaded), tooltip read 2..10 | in flight (worker PR) | master talents.go:36 `0.03` a rank; PB `0.02`, tree/proto reordered |
| PRI-70425-07 | P6 | Improved Mind Flay: each MF tick deals 10/20% more than the previous tick | yes (shadow, shadow-holy 2/2) | 1225139 eff0 aura 108 -> 4 (dummy, bp 20), effects 1-3 (range, slow, +50) removed: server side | in flight (worker PR); ambiguous (needs verify) | master talents.go:169-178 flat +10%/rank; PB mind_flay.go OnTick multiplies by `1.2^(tick-1)` (compounding: 1, 1.2, 1.44). Additive (1, 1.2, 1.4) also fits the text, see NtV-2 |
| PRI-70425-08 | P6 | New Shadow talent Dark Inspiration (row 4): MB and full MF give party +3..15% Spirit for 10 s | minor (PB shadow takes 3/5; Spirit only feeds regen/Spiritual Guidance) | 1326159 dummy bp 15, buff 1326168 aura 80 (stat %) misc 4 (Spirit) bp 15, duration index 1; TraitDefinition 135353 15274 -> 1326159 | in flight (worker PR) | PB talents.go `registerDarkInspiration`, triggers in mind_blast.go and on MF last tick; party part not simulated (priest only) |
| PRI-70425-09 | P6 | Improved Mind Blast moved row 3 -> 2 | yes (tree layout, talent strings) | TraitNode positions; TraitDefinition 135361 15318 -> 15273 | in flight (worker PR) | PB proto/priest.proto reorder, new talent strings in all 3 presets |
| PRI-70425-10 | P6 | Improved Fade and "Shadow Infinity" (= Shadow Affinity) removed | n/a (threat) | TraitNodeEntry: Improved Fade 15274 and Shadow Affinity 15318 gone (fc notes agree it is Shadow Affinity) | in flight (worker PR) | master talents.go:141 `applyShadowAffinity`; PB deletes it |
| PRI-70425-11 | P6 | Vampiric Embrace cooldown 30 s (was 1 min) | minor (no preset casts VE) | SpellCooldowns 15286 RecoveryTime 60000 -> 30000 | in flight (worker PR) | master vampiric_embrace.go:57; PB 30 s |
| PRI-70425-12 | P6 | Vampiric Embrace can't be resisted | minor | SpellMisc 15286 Attributes_3 67108992 -> 67371136 (+0x40000, the same "can't miss" bit Shadow Weaving got in 70170) | **missing (missed by worker)** | vampiric_embrace.go:62 still `spell.OutcomeMagicHit` on master and PB; should be `OutcomeAlwaysHit` under Forever |
| PRI-70425-13 | P6 | Holy Fire rank 5 ticks no harder than rank 6 | n/a (rank 8 used) | SpellEffect 15265 eff1 bp 13 -> 9 | in flight (worker PR) | master holy_fire.go:19 `13`; PB `9` |
| PRI-70425-14 | P6 | Inner Focus used up by one spell even when a spell is queued into Penance | n/a (sim has no spell queue; no preset casts Penance) | server side | n/a | |
| PRI-70425-15 | P6 | Devouring Contagion spreads to 2 targets; the bounced DP costs no mana, no cast, no CD reset | n/a (one boss) | 1309950 eff2 added (dummy bp 2) | not simulated on purpose | talents.go `applyDevouringContagion` comment: jump does nothing vs one boss |
| PRI-70425-16 | P6 | Silent Resolve threat cut applies to all spells | n/a (threat) | 14523 eff0 misc 2 -> 34 (Holy -> Holy+Shadow school mask) | n/a | master talents.go:24-31 Holy only |
| PRI-70425-17 | P6 | Improved PW:Shield moved to row 3 | n/a | tree layout | in flight (worker PR) | PB proto reorder |
| PRI-70425-18 | P6 | Renew from several priests, PW:S messages, SHIELDED text, Mind Vision, Divine Aegis tooltips, Shadow Weaving tooltip | n/a | 431622 eff1 added bp 20 (shield cap); 15257 tooltip now "damage the target takes from you" | n/a | Shadow Weaving tooltip matches how the sim already models it (caster-only, talents.go:207-252) |
| PRI-70425-19 | P6 Legacy | The Quick and the Dead affects more Priest spells | n/a (out of combat) | 1262312 mask now hits Penance, Holy Nova | n/a | |
| PRI-70425-20 | P6 Races | Undead Touch of the Grave: 30% for casters, 25 s ICD, **no longer scales with Shadow increases on target or caster** | yes (all 3 presets are Undead) | SpellAuraOptions 1260201 ProcChance 10 -> 30, ProcCategoryRecovery 1000 -> 25000 | partial / **wrong** (shared racial, #50) | sim/core/racials_forever.go:319-390: 30% and 25 s landed. The drain (line 334) has no SpellFlagIgnoreModifiers, so Shadowform, Darkness and Shadow Weaving (priest PseudoStats) and target Shadow debuffs scale it |
| PRI-70425-21 | P6 Races | Gnome Eureka! 8 s buff, now also PW:S | n/a (presets Undead) | 1259823 DurationIndex 8 -> 31, charges removed | landed (shared #50) | racials_forever.go:209-245 (8 s); not priest-specific |
| PRI-70425-22 | P6 Classes | Wands need 0.5 s standing still | n/a (wand not simulated) | Shoot 5019 Attributes_4 0x1000000 removed | n/a | |
| PRI-70291-01 | P5, 10-08, 70291 | Smite ranks 1-4 base points retuned | minor (presets use ranks 7/8) | 585 bp 15 ppl 0.1; 591 bp 28->27 ppl 0.6->0.3; 598 bp 46->45 ppl 0.8->0.5; 984 bp 59->60 ppl 1->0.7 | landed | smite.go:16-17 match the client at each rank's cap (e.g. r4 58.96-68.04 = 63.5 mid). Ranks 5-8 unchanged in client (r8 170 bp, 1.3/lvl, 0.714). Rank 8 x1.09 is the user-ruled assumption priest_smite8_tuning (smite.go:24) |
| PRI-70291-02 | P5 | Penance damage reduced (2 s channel, not 3), coefficient | minor (smite takes Penance for PI, but never casts it) | bolts 402284 bp 81->42 +0.29/lvl, 1240727 113->50 +0.33, 1240730 180->68 +0.43, 1316993 131->92; coef 0.285 -> 0.19 all ranks | landed | penance.go:18-19 (44.61/52.97/71.87/92 at caps), `BonusCoefficient: 0.19`; 3 bolts (now + 2 ticks) |
| PRI-70291-03 | P5 | Penance mana costs (rank 1 +50%) | minor | SpellPower 402174 100->150, 1240720 185->220, 1316995 355->385 (1240721 270) | landed | penance.go:20 {150,220,270,385} |
| PRI-70291-04 | P5 | Lesser Heal / Heal low ranks, Penance healing r2/r4 | n/a | 2050/2052/2053/2054/2055, 402289/1240723/1240724/1316991 | n/a | |
| PRI-70170-01 | P4, 10-01, 70170 | Devouring Plague can crit | yes | DP 2944..19280 SpellMisc Attributes_8 0 -> 512 | landed | core/ruleset.go:15 every Forever DoT tick rolls crit; devouring_plague.go has no SpellFlagNoPeriodicCrit; Shadowform crit bonus applies (talents.go:347) |
| PRI-70170-02 | P4 | SW:Death no longer always gets Early Demise | yes (shadow presets take ED 2/2 on PB) | server side | landed | shadow_word_death.go:69 `sim.IsExecutePhase20()`, 15%/rank (client 1310076 bp 30/20) |
| PRI-70170-03 | P4 | Inner Focus crit no longer applies to periodic effects | yes (IF before MB / Holy Fire) | 14751 eff1 mask_0 -615203184 -> -638304624: removed SW:P 0x8000 and DP 0x2000000, **added** MF 0x800000 and Starshards 0x200000 | landed; channels ambiguous (needs verify) | talents.go:277-280 excludes PureDot and Channeled; holy_fire.go removes it from HF ticks. The client mask now holds the two channels, see NtV-3 |
| PRI-70170-04 | P4 | Shadow Weaving can no longer fail to apply | yes | 15258 Attributes_3 128 -> 262272 (+0x40000) | landed | talents.go:239-251 proc spell has no hit roll; proc chance 33/67/100% (15257 trait curve) stays |
| PRI-70170-05 | P4 General | Ranks "vastly below" your level get less spell power and proc chance | yes (smite rank 7 fallback) | server side | landed (assumption) | priest.go:150 `downrankCoefficient` (fc formula, 5%/level beyond 17); notes.md:132 already lists it |
| PRI-70170-06 | P4 | Cooldown Manager / UI items | n/a | | n/a | |
| PRI-70009-01 | P1, 09-24, 70009 | Renew can crit; PW:S can overwrite without Weakened Soul | n/a | Renew Attributes_8 +512; PW:S Attributes_4 +0x100000 | n/a | healing spells not registered (priest.go:110-116) |
| PRI-70009-02 | P1 Night Elf | Starshards gains Twilight Focus pushback resistance | minor (no NE preset) | 14913 mask_0 9846416 -> 11943568 (+0x200000 Starshards) | in flight (worker PR) | master has no Twilight Focus at all; PB talents.go `applyTwilightFocus` (23/47/70%, incl. Starshards), TestTwilightFocus |
| PRI-70009-03 | P1 Undead | Dark Sacrifice no longer breaks CC; mana scales with Spirit | yes (all presets cast it below 40% mana) | 1277324-28 tooltip `${$o2+$SPI}`, eff0 aura 3 -> 226 | landed (assumption) | dark_sacrifice.go:59-65 1600 + Spirit; existing notes.md:119 item |
| PRI-70009-04 | P1 Gnome | Contingency Plan can crit; Confounding Flash visual | n/a | | n/a | |
| PRI-70009-05 | P1 Races/Items | Eureka! 10% for every class (priests were 15%); Touch of the Grave needs a damage component; wands no longer gain spell damage | n/a / shared | 1259823 eff0 -15 -> -10 | landed (shared) / n/a | racials_forever.go (Eureka 10%, ToG `spell.DamageMultiplier != 0`); wands not simulated |
| PRI-DD-01 | DD | DP and Fear Ward for all races at 20; SW:Death baseline at 32 (15 s CD at that build); Divine Spirit baseline | yes | SkillLineAbility race mask -1 for DP | landed | priest.go:93-101 registers DP for every race; SWD; AddRaidBuffs DivineSpirit (priest.go:82) |
| PRI-DD-02 | DD | Starshards: more damage, 30 s cooldown | minor (NE) | 19305 bp 300/tick, coef 0.167, CategoryRecoveryTime 30000 | landed | starshards.go:20, 64 |
| PRI-DD-03 | DD | Chastise (Dwarf): damage + root, Humanoids, 2 min | minor | 1277331-35 TargetCreatureType 0 -> 64 (70009) | landed | chastise.go:62-64 Humanoid check, 2 min |
| PRI-DD-04 | DD | Shadowform: non-healing spells allowed, -50% Shadow cost, +100% crit bonus, lower damage bonus | yes | 15473: aura 79 +10% Shadow, aura 72 -50%, aura 108 misc 15 +100, -15% physical taken, 40% base mana | landed | talents.go:337-399 |
| PRI-DD-05 | DD | Shadow Weaving: caster-only, smaller per stack | yes | 15258 aura 270 bp 2, 5 stacks, 15 s | landed | talents.go:207-252 (2% a stack on the priest's Shadow school multiplier) |
| PRI-DD-06 | DD | Shadow Focus now +hit (1%/rank); Holy Precision up to 18%; Mental Agility 3/7/10 incl. Smite/HF; Meditation larger; Searing Light all Holy; Spiritual Guidance ~1/3 for damage; Twin Disciplines (instants); Power in Light (HF); Early Demise; Devouring Contagion (25/50% DP cost) | yes | 15260 bp 5; 1309957 bp 18; 14520 bp -10; 14521 bp 50; 14909 aura 79 Holy; 14901 bp 8; 1309950 bp -50 (DP via mask_3 0x10 since 70170) | landed | talents.go:91-207; values equal the tooltip-read curves in tools/forever_talents/data/priest.json |
| PRI-DD-07 | DD | Penance: 2 s channel, 12 s cooldown | minor | 1316995 CategoryRecoveryTime 12000 | landed | penance.go:85-98 |
| PRI-DD-08 | DD | Twilight Focus (ex Healing Focus): pushback protection on all spells, 3 points | yes once raid hits push back casts (presets have raidDamageHitsPerMinute 2; PB smite takes 3/3) | 14913 aura 108 misc 9 bp 70 | in flight (worker PR) | master: no pushback protection anywhere in sim/priest; PB applyTwilightFocus; core/cast.go:115 uses `PushbackReduction` |
| PRI-DD-09 | DD | Vampiric Embrace 30 s duration (1 min CD at that build), Improved VE removed | minor | | landed | vampiric_embrace.go:19 30 s, 20% heal |
| PRI-DD-10 | DD | Healing/utility talents and spells (Binding Heal, PoM, Divine Aegis, Renewed Hope, Soul Warding, Fade, PoH, Elune's Grace, Divine Grace, Contingency Plan, Confounding Flash...) | n/a | | n/a | |

#### 2. Unannounced client changes (reverse check)

Client-only builds: 70291->70334 changed only Summon Azurky (item), 70334->70338 and 70338->70374
changed no spells. Nothing for the priest: confirmed. 70009->70170 and 70170->70291 priest sections
hold only the items below plus tooltip text.

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| PRI-U-01 | DB2 70170 | Inner Focus crit mask gained Mind Flay and Starshards (and lost SW:P and DP) | minor (presets spend IF on MB or HF) | 14751 eff1 mask_0 bits +0xA00000, -0x2008000 | ambiguous (needs verify) | talents.go:277-280; see NtV-3 |
| PRI-U-02 | DB2 70170 | Devouring Contagion cost mod moved from mask_0 bit 25 to mask_3 0x10 | none (still hits DP) | 1309950 eff0 mask_0 33554432 -> 0 (mask_3 16) | landed | talents.go:184-195 |
| PRI-U-03 | DB2 70338/70374 state | Twin Disciplines DoT mask and Inner Focus free mask already held DP (mask_3 0x10) before 70425 | yes (shadow, smite) | 1225132 eff1 mask_3 16; 14751 eff0 mask_3 268436624 | wrong on master; in flight (worker PR) | master talents.go:76 (no DP) and :275 (DP excluded from free cast); PB e63451452 fixes both |
| PRI-U-04 | DB2 70170 | New priest Revelation aura 1323377 (+100% crit next non-periodic spell; mask: MB, MF, SWD, Starshards, Chastise, Penance, Smite, **not Holy Fire**) | n/a (no priest preset uses Enchant Weapon - Revelation) | 1323377 added, eff0 aura 107 bp 100 | n/a | sim/common/enchant_effects.go:282-322 gives it to every direct APL spell incl. Holy Fire; matters only if a priest set ever takes Revelation |
| PRI-U-05 | DB2 70425 | Attributes_6 0x200000 on CC spells, Mind Flay, PW:S, Shoot (TC: CLIENT_UI_TARGET_EFFECTS) | n/a (UI) | many SpellMisc rows | n/a | |
| PRI-U-06 | DB2 70425 | Inner Fire / Shadowfiend Scaling aura 22 -> 674 | n/a | | n/a | |
| PRI-U-07 | DB2 70009 | Improved PW:S coefficient 1 -> 0; Lightwell Renew; Chastise Humanoid restriction | n/a / landed | 24191, 468278; 1277331-35 | n/a / landed | chastise.go:62 |

Damage spells unchanged from 70009 to 70425 (checked against the sim): Mind Blast r9 (bp 485,
2.6/lvl, 0.429, 350 mana: mind_blast.go matches 476.93-503.47), Mind Flay r6 (130/tick, 0.167),
SW:P r8 (127/tick, 0.2), DP r6 (106/tick, 0.1, 985 mana), SWD r4 (448, 2.5/lvl, 0.429, 340 mana),
Holy Fire r8 (208 +2/lvl, 0.75 / 15 a tick 0.05), Smite r8, Starshards r7, Power Infusion (20%,
3 min, 20% base mana).

#### 3. Proposed fixes

1. **Power in Light on Shadow Word: Pain (PRI-70425-03).** sim/priest/penance.go `powerInLightMultiplier`:
   after the Holy Fire loop, also return `1 + 0.02*rank` when any `priest.ShadowWordPain[r]` has
   `Dot(target).IsActive()`. Update the comment, notes.md:126 and the smite meta "Holy Fire ... for
   Power in Light" note. Test: in sim/priest (a smite test file), put only SW:P on the target and
   check Smite's damage multiplier is 1.10 at 5/5, 1.0 with neither DoT. DPS: smite +0.3 to +1%
   (Smites that land in the Holy Fire gap; rotopt may then re-weigh Holy Fire upkeep). Presets:
   priest__smite.json only (shadow presets take no Power in Light).
2. **Vampiric Embrace can't be resisted (PRI-70425-12).** sim/priest/vampiric_embrace.go:62
   `spell.OutcomeMagicHit` -> `spell.OutcomeAlwaysHit` under `priest.Env.IsForever()` (Classic keeps
   the roll). Test: 1000 casts against a level 63 target never miss. DPS: 0 for the presets (none
   casts VE); fixes the party-heal metric only.
3. **Touch of the Grave ignores Shadow modifiers (PRI-70425-20, shared racial).**
   sim/core/racials_forever.go:334 add `SpellFlagIgnoreAttackerModifiers | SpellFlagIgnoreTargetModifiers`
   (keep DamageMultiplier 1) to the drain spell. Test in sim/core: an Undead priest with Shadowform
   and Darkness 5/5 drains exactly 5% of max health. DPS: about -2 to -3 for priest__shadow and
   priest__shadow-holy (drain ~190 every ~27 s, today scaled ~1.33x by Shadowform x Darkness x
   Weaving), ~0 for priest__smite. Route to the shared-racials owner; also lowers Undead warlocks.
4. **Merge the worker PR** for every "in flight" row; before merging, fix the stale PB meta text
   "off its 1 min cooldown" in priest__shadow.meta.json and priest__shadow-holy.meta.json
   (rotation[3]) and "Raiments 4/5-piece ... Devouring Plague without a cooldown" (now worth less with
   a 24 s base cooldown; regenerate the +40.2 figure).

#### 4. Need to Verify candidates

* **NtV-1 Shadow Word: Death's old cooldown.** The 15 October notes say "increased to 18 seconds (was
  12 seconds)", but client 70374 had 15000 ms (SpellCooldowns 1309595/633/635/636) and the deep dive
  said 15 s. The new 18 s agrees everywhere, so the sim's 18 s (PB) stands; only the "was" differs.
  Nothing to test unless a later build lands on 12 or 15; check the in-game tooltip at 32+.
* **NtV-2 Improved Mind Flay stacking.** "Each instance of damage of Mind Flay deals 10/20% more damage
  than its previous instance" (70425). The client turned the talent into a dummy (1225139 aura 4, bp
  20), so the server decides. PB reads it as compounding (ticks x1, x1.2, x1.44 at 2/2); additive
  (x1, x1.2, x1.4) is the other reading, ~1.5% less Mind Flay damage, maybe -0.5% shadow DPS. A 2/2
  priest's three Mind Flay ticks on a training dummy (no crits, same spell power) settle it: third
  tick / first tick = 1.44 or 1.40.
* **NtV-3 Inner Focus and channels.** 70170's note: Inner Focus "no longer applies its Critical Strike
  increase to Periodic Effects", and the tooltip now says "non-periodic spell", but the same build
  added Mind Flay and Starshards to Inner Focus's crit mask (14751 eff1) and took SW:P and DP out.
  The 70291 Eureka note says "Channeled spells do not count as periodics", which suggests Inner Focus
  then crits every Mind Flay tick. The sim keeps channels out (talents.go:277-280). No preset spends
  Inner Focus on Mind Flay, so the DPS effect is nil today; it decides whether "Inner Focus -> Mind
  Flay" is a valid rotation. Beta: Inner Focus then a full Mind Flay, count tick crits over many casts.
* **NtV-4 Power in Light's size.** The talent spell holds bp 15 (1309969, dummy) while the tooltip
  read gives 2/4/6/8/10% (curve 109048 not downloaded). The sim uses 2% a rank. A 5/5 Smite with and
  without Holy Fire (or now SW:P) on the target settles 10% vs 15%: smite preset +1.5% if 15%.
* **NtV-5 Touch of the Grave and Shadow modifiers** (shared): the note says it no longer scales with
  Shadow increases; the sim scales it today (fix 3). Beta: an Undead Shadowform priest's drains equal
  5% of max health with and without Shadowform.

#### 5. Count

54 rows (47 announced/design + 7 unannounced), first status of each row: landed 20, in flight
(worker PR) 14, missing 2 (Power in Light on SW:P, VE unresistable; both missed by worker), wrong 2
(Touch of the Grave Shadow scaling, shared; master's Twin Disciplines / Inner Focus DP masks, fixed
in PB), ambiguous (needs verify) 1 (Inner Focus crit mask channels; also flagged on PRI-70170-03 and
on PRI-70425-07 Improved Mind Flay), not simulated on purpose 1, n/a 14.

### Shaman: game-update ledger (builds 70009 to 70425)

Scope: sim/shaman, the shaman totem raid buffs in sim/core (buffs.go, totem_weapon_buffs.go), presets tools/dps_report/req/shaman__*.json.
Checkout: master 6a87c9c (origin/master has since gained #56, racials only). Worker branch: origin/claude/wow-forever-dps-presets-m9u15h-shaman (0edc3f808 "client 1.60.1.70425 data (work in progress)", merged with master at 28597f4ce). No other branch touches the totem buff code in sim/core.

Links: P1 = https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/1 (2026-09-24, 70009); P4 = .../2360696/4 (2026-10-01, 70170); P5 = .../2360696/5 (2026-10-08, 70291); P6 = .../2360696/6 (2026-10-15, 70425); DD = https://worldofwarcraft.blizzard.com/en-us/news/24302097 (Mage and Shaman deep dive, 2026-10-07, describes the design at 70170).

Presets (all Orc, level 60, vs level 63): elemental-level60, elemental-level60-mana-tide (LB 15208 / 10392, Lava Burst 1238300, Flame Shock 29228, Earth Shock, Mana Tide 17359); enhancement-level60, -imp-gw, -mana-tide, -water-shield (Stormstrike, Fire Nova 408345, Earth/Flame Shock, LB 15208, Windfury Weapon, own SoE + Windfury Totem, Magma 10587 in the last two). No preset casts Chain Lightning or Water Shield. Published Fire Nova share (ui/dps_report/data/abilities.json): enh-level60 111.0 DPS (13.3%), imp-gw 109.8, water-shield 64.1, mana-tide 58.0.

#### 1. Announced changes

##### Build 70425 (P6, 2026-10-15)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHA-70425-01 | P6, 10-15, 70425 | Totemic Recall costs 22 s at the trainer; Grounding Totem is destroyed after redirecting; Chain Heal visual | n/a | server side / visual | n/a | not simulated |
| SHA-70425-02 | P6 | Windfury Weapon's attacks show as white hits in combat text | n/a | 439440 SpellMisc Attributes_10 0 -> 0x2000000 (display flag) | n/a | cosmetic |
| SHA-70425-03 | P6 | Grace of Air Totem is no longer exclusive with Windfury Totem and Flametongue Totem | yes (raid buff composition; no preset carries both today) | GoA buffs 8836 / 10626 / 25360 SpellMisc Attributes_11 1024 -> 0 (the exclusivity bit added in 70009 is gone); WF 8515/10609/10612 and FT 8230/8250/10521/15036 keep 1024 | **wrong** | sim/core/buffs.go:2009 still puts the Forever Windfury Totem buff above Grace of Air in `graceOfAirWindfuryCategory` (Priority 2 vs GoA's Priority 1 at buffs.go:1604), so any WF Totem buff (raid `TotemWeaponBuff` or own totem) zeroes GoA's 77 Agility. Test sim/shaman/enhancement/grace_of_air_test.go:40 asserts the old rule. Not on the shaman worker branch or any other branch: **missed by worker** (no owner, shared core) |
| SHA-70425-04 | P6 | Windfury Totem and Flametongue Totem keep their exclusivity | yes | Attributes_11 1024 still on WF and FT buff spells | landed | one totem-weapon slot: core/totem_weapon_buffs.go:119-138 (`applyRaidTotemWeaponBuff`, first buff only) and sim/shaman/shaman.go:89-124 (`setTotemWeaponBuff` swaps own and raid buff) |
| SHA-70425-05 | P6 | Grace of Air, Strength of Earth, Windfury and Flametongue Totems are raid wide | minor (single-player sim) | SpellEffect Effect 35 -> 65 (party -> raid area aura) on 8836/10626/25360, 8076/8162/8163/10441/25362, 8515/10609/10612, 8230/8250/10521/15036; tooltips "party or raid" | landed | the raid-buff toggles (buffs.go:407-417) never were party-limited; wording in orchestration/standards.md:140 (#51). Fire Resistance Totem too (section 2) |
| SHA-70425-06 | P6 | Chain Lightning mana cost about -35% on all ranks | minor (no preset casts CL) | SpellPower 421 225->145, 930 305->200, 2860 390->250, 10605 485->315 | in flight (worker PR) | master chain_lightning.go:24 `{0, 225, 305, 390, 485}`; worker `{0, 145, 200, 250, 315}` |
| SHA-70425-07 | P6 | Fire Nova spell power coefficient 0.143 (was 0.214) | yes (4 Enhancement presets) | SpellEffect 1072539..1072545 (408423/408424/408426/408427/408428) EffectBonusCoefficient 0.214 -> 0.143 | **missing** (missed by worker) | sim/shaman/fire_totems.go:241 `FireNovaSpellCoeff = {0, .214 x5}`; worker branch unchanged |
| SHA-70425-08 | P6 | Fire Nova base damage significantly reduced | yes (4 Enh presets) | 408423 bp 50->33 rppl 1.1->0.4; 408424 102->55, 1.6->0.6; 408426 183->130, 2.2->0.9; 408427 284->179, 2.8->1.1; 408428 403->239, 3.4->1.3. At each rank's cap (learn+5), with the client Variance: r1 32.31-37.69, r2 54.28-61.72, r3 126.27-142.73, r4 173.39-195.61, r5 232.02-258.98 (was 396.95-443.05) | **missing** (missed by worker) | fire_totems.go:239-240 still the 70338 table and `{57, 3.4}` growth; the comment at :225-234 also |
| SHA-70425-09 | P6 | Fire Nova mana cost significantly reduced | yes | SpellPower 408341..408345 95/170/280/395/520 -> 65/120/210/290/375 | in flight (worker PR) | master fire_totems.go:242; worker has the new costs |
| SHA-70425-10 | P6 | Flametongue Totem and Windfury Totem only affect players | minor | SpellMisc Attributes_3 0 -> 0x100 (only players) on 8230/8250/10521/15036/8515/10609/10612 | landed | pets never get it: `applyRaidTotemWeaponBuff` returns when `!character.HasMHWeapon()` (totem_weapon_buffs.go:123), and pets have no equipped main hand. SoE/GoA still reach pets (correct, no flag) |
| SHA-70425-11 | P6 | Elemental Focus: its Clearcasting can no longer be spent on the Fire Nova that triggered it | yes (Enh presets take Elemental Focus and cast Fire Nova) | server side | landed | sim/shaman/talents.go:239-243: the Clearcasting aura skips consumption on the cast that just activated it |
| SHA-70425-12 | P6 | Elemental Mastery returns as the 31-point talent: LB and Lava Burst cast 0.5 s faster, Chain Lightning has no cooldown, 15 s, 3 min CD, needs Lava Burst | yes (Elemental presets could take it) | 16166 added: eff0 A107 misc 10 (cast time) -500 ms, mask LB 0x1 + LvB 0x1000 (word 1); eff1 A108 misc 11 (cooldown) -100% on CL (0x2); DurationIndex 8 = 15 s; RecoveryTime 180000; **ProcCharges 1**, ProcTypeMask 0x15550; TraitDefinition 134219 408490 -> 16166 | in flight (worker PR) | worker talents.go `registerElementalMasteryCD` (15 s aura, -500 ms DefaultCast, CL CD 0, 3 min, MCD). Worker's elemental preset strings `550433100000135-...` spend 50 of 51 points and do not take Elemental Mastery (WIP). Charges question: NTV-1 |
| SHA-70425-13 | P6 | Lava Burst moves to row 5 (21-point), no prerequisite, new rank 4, ranks reflowed | yes | new 1326047: bp 220 rppl 1.3 coef 0.714 Variance 0.2533, 265 mana, L60, cast 2500, CD 10000; 408490 -> L30 bp 164->123 rppl 0.9->0.8 mana 165->135; 1238299 -> L40 196->164, 1.1->0.9, 230->165; 1238300 -> L50 220->196, 1.3->1.1, 265->230 | in flight (worker PR) | master lava_burst.go:16-22 has 3 ranks (rank 3 = 1238300 at 60, 192.14-247.86, 265 mana: identical numbers to the new rank 4, so 0 DPS change). Worker: 4 ranks, presets/APLs switch to 1326047 |
| SHA-70425-14 | P6 | Tree moves: Elemental Fury needs Lightning Overload, LO needs Elemental Alacrity, Alacrity row 3->2, Warding row 2->3, Earthbound row 4 (16-point), column rearrangement, Tidal Focus row 1 / Totemic Focus row 2 | minor (talent strings) | TraitDefinition 134226 16578->28996, 134232 28996->16578, 134221 1222988->408490, 134223 16120->1222988, 134190 16173->16179, 134196 16179->16173 (effect points move with the spell, no value change); TraitNode positions | in flight (worker PR) | worker proto/shaman.proto field reorder, tools/forever_talents, ui/core/talents/trees/shaman.json, every shaman preset's talentsString |
| SHA-70425-15 | P6 | Call of Thunder removed | yes (both Elemental presets take it: +3% crit on LB/CL) | TraitDefinition 134223 16120 -> 1222988 (node now Earthbound); 16120 has no tree node | in flight (worker PR) | master electric_spell.go:55 `BonusCritRating: TernaryFloat64(CallOfThunder, 3, 0)`; worker deletes it and the proto field |
| SHA-70425-16 | P6 | Elemental Reach also cuts Totemic Projection's CD by 10/20 s | n/a | 28999 new eff2 A107 misc 11 bp -20000 mask word3 0x80 | n/a | no Totemic Projection in the sim; the Flame Shock range loss is in section 2 |
| SHA-70425-17 | P6 | Call of Flame also extends Flame Shock by 3/6/9 s | yes (all 6 presets take Call of Flame 3) | 16038 new eff2 A107 misc 1 (duration) bp 9000 mask 0x10000000 (Flame Shock), curve 126024 -> 3000 a point | in flight (worker PR) | master flame_shock.go:36 `numTicks := 4`; worker `4 + CallOfFlame` ticks of 3 s with the per-tick damage unchanged (base/4) |
| SHA-70425-18 | P6 | Clearcasting is no longer consumed by Lightning Overload | yes (Elemental takes LO and Elemental Focus) | server side | landed | lightning_overload.go:16 gives overload copies `SpellFlagNoOnCastComplete`, so they neither spend nor trigger Clearcasting (talents.go:234-254) |
| SHA-70425-19 | P6 | Fire Nova tooltip follows Trial by Fire (Improved Fire Nova renamed) | n/a | 16086 name "Improved Fire Nova" -> "Trial by Fire" | n/a | the Magma Totem part is unannounced, section 2 U-01 |
| SHA-70425-20 | P6 | Stormstrike's debuff is no longer consumed if the spell does not land | yes (4 Enh presets: ES/LB/CL after Stormstrike; 1.6 ES and 0.9 LB misses per 180 s fight) | server side | **missing** (missed by worker) | stormstrike.go:120-131 `spendStormstrike` deactivates the mark before the roll; called before `CalcDamage` in lightning_bolt.go:86, chain_lightning.go:84, earth_shock.go:84 |
| SHA-70425-21 | P6 | Stormstrike can't be consumed by other players or pets | yes | server side | landed | per-shaman mark (stormstrike.go:43-52 `stormstrikeMarks`, the shared Era debuff skipped under Forever) |
| SHA-70425-22 | P6 | Stormstrike costs 8% of base mana (was 125) | yes (4 Enh presets) | SpellPower 182998 ManaCost 125 -> 0, PowerCostPct 0 -> 8 -> 121.6 at 60 (base mana 1520) | in flight (worker PR) | master stormstrike.go:62 `FlatCost: 125`; worker `BaseCost: .08`, test updated |
| SHA-70425-23 | P6 | Mana Tide Totem gets a rank at 38 (ranks 25 / 38 / 48 / 58) | minor (top rank unchanged) | new 1325256/1325257 rank 4 L58 290 a tick, cost 60; 17354 -> L38 197->133, cost 30->20; 17359 -> L48 290->197, cost 60->30 | in flight (worker PR) | master water_totems.go:207-210 3 ranks (17359 = 290). Master presets cast 17359 at 290 = new rank 4's value; worker presets cast 1325256. 0 DPS change |
| SHA-70425-24 | P6 | Restorative Totems raises all of Healing Stream (was base) | n/a | 16187 eff1 misc 8 -> 22 | n/a | healing |
| SHA-70425-25 | P6 | Water Shield triggers from any of the shaman's crit spells (was heals) | minor (no preset casts Water Shield; enhancement-water-shield keeps Lightning Shield, "worth more at every raid hit rate") | SpellAuraOptions 223904 ProcTypeMask 0x262a8 -> 0x362a8 (adds 0x10000, harmful magic spell done) | **missing** (missed by worker) | water_shield.go:62-66 only `OnHealDealt` crits. Could flip the water-shield preset's shield choice |

##### Build 70291 (P5, 2026-10-08)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHA-70291-01 | P5, 10-08, 70291 | Flametongue Totem uses a shapeshifted target's attack speed, not its weapon's | minor (feral presets take Windfury, not FT Totem) | server side | **wrong** | core/totem_weapon_buffs.go:108 `weaponSpeed = character.MainHand().SwingSpeed` (the item's speed, e.g. a 3.x staff) where a Forever cat swings at 1.0 (sim/druid/forms.go:78) |
| SHA-70291-02 | P5 | Lightning Bolt ranks 1-5 rebased so every rank is an upgrade | minor (L60 casts rank 10, fallback rank 8) | 403 bp 14->15 rppl .4->.1; 529 29->27, .5->.1; 548 rppl .6->.3; 915 56->61, .6->.4; 943 72->82, .7->.5 (rank 5 at 31: 84.5) | landed | lightning_bolt.go:23-24: r1 15.5, r2 27.5, r3 46.5, r4 63.0, r5 84.5 averages at cap, all match |
| SHA-70291-03 | P5 | Elemental Focus is no longer activated and immediately consumed by Fire Nova | yes (Enh) | server side | landed | talents.go:239-243 (same guard as SHA-70425-11) |
| SHA-70291-04 | P5 | Water Shield shows the mana restored; it never had the 15 s cooldown | minor | SpellCooldowns 84832 CategoryRecoveryTime 15000 -> 0; new eff1 A4 bp 1 | landed | water_shield.go:13-15, no cooldown, recast when the globes are gone |
| SHA-70291-05 | P5 | Mana Tide Totem shows its pulse visual again | n/a | visual | n/a | |

##### Build 70170 (P4, 2026-10-01)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHA-70170-01 | P4, 10-01, 70170 | Ghost Wolf more visible; Disease Cleansing Totem lasts 5 min; Totemic Recall restores mana | n/a | 8170 DurationIndex 4 -> 5; Totemic Recall 1323420 / 39104 added | n/a | not simulated |

##### Build 70009 (P1, 2026-09-24)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHA-70009-01 | P1, 09-24, 70009 | Lightning Bolt ranks 3 and 4 hit harder | minor | 548 bp 35->45, 915 50->56 (915 later 61 in 70291) | landed | superseded by SHA-70291-02, which the sim has |
| SHA-70009-02 | P1 | Windfury Totem and Flametongue Totem buff names | n/a | SpellName 8515/10609/10612, 8230/8250/10521/15036 | n/a | |
| SHA-70009-03 | P1 | Flametongue Totem has no duration and is replaced at once when another totem goes down | minor (no preset drops FT Totem) | 8230/8250/10521/15036 DurationIndex 1 -> 0, Effect 6 -> 35 area aura (aura 4 -> 42); passive 8229/8251/10524/16388 removed | landed | FlametongueTotemBuffAura is NeverExpires and lives with the totem (fire_totems.go:335-372, placeTotem in the fire slot) |
| SHA-70009-04 | P1 | Flametongue Totem can't be stacked by recasting | minor | as above (one area aura) | landed | shaman.go:93-99 keeps one totem-weapon buff |
| SHA-70009-05 | P1 | Flametongue Totem no longer stacks with Flametongue Weapon | minor | 8230.. Attributes_11 0 -> 1024; FT Weapon tooltip "disables any benefit ... from Flametongue Totem" | landed | totem_weapon_buffs.go:105 `mainHandHasImbue(flametongueWeaponEnchantIds)`; shaman.go:90 |
| SHA-70009-06 | P1 | Flametongue Totem no longer stacks with Windfury Totem | minor | Attributes_11 1024 on both | landed | single totem-weapon slot (SHA-70425-04) |
| SHA-70009-07 | P1 | Tranquil Air, Windfury and Grace of Air Totems no longer stack, even from different shamans | yes | Attributes_11 0 -> 1024 on 8836/10626/25360, 8515/10609/10612, 25909 | n/a (superseded) | the GoA part was reverted in 70425 (SHA-70425-03, where the sim is now wrong); Tranquil Air is removed (DD) |
| SHA-70009-08 | P1 | Elemental Fury (row 3) and Elemental Alacrity (row 6) swap | n/a | tree | n/a | superseded by SHA-70425-14 |
| SHA-70009-09 | P1 | Lava Burst ranks 1 and 2 about 10% above the matching LB rank | minor | 408490 bp 113 -> 164, 1238299 179 -> 196 | landed | master lava_burst.go:16 (r1 avg 171.2 at 48, r2 204.8 at 58); then reflowed by SHA-70425-13 |
| SHA-70009-10 | P1 | Rage of the Farseer no longer raises spell casting speed | yes (4 Enh presets) | 425336 eff2 A65 bp 30 removed; remaining A342 bp 30 (melee and ranged speed), 25 s, 180000 ms | landed | talents.go:580-623: `MultiplyMeleeSpeed(1.30)` only, 25 s, 3 min |
| SHA-70009-11 | P1 | Riptide can crit | n/a | 408521.. Attributes_8 | n/a | healing |

##### Deep dive (DD, 2026-10-07; design as of 70170)

Talents checked against shaman_audit.md section 8 (all items ticked) and the code; values there come from the 70338 trait curves.

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHA-DD-01 | DD | All totems have a 1.0 s GCD | yes | 70338 audit (forever_class_changes.md:48) | landed | totems.go:11 `totemGCD = time.Second` |
| SHA-DD-02 | DD | Persistent-buff totems last 5 min | yes | | landed | air_totems.go:15, earth_totems.go:43, water_totems.go:69, fire_totems.go:336 |
| SHA-DD-03 | DD | Resistance totems are raid wide; totem tracking UI; Tranquil Air removed | n/a | | n/a | |
| SHA-DD-04 | DD | Imbues last 60 min and stack with stones and oils | yes | | landed | presets run Windfury Weapon + Shadow Oil; totem_weapon_buffs.go:10-24 |
| SHA-DD-05 | DD | Fire Nova is a spell that erupts at the current fire totem | yes | 408341.. E3 dummy, damage spells 408423.. | landed | fire_totems.go:225-300 (needs a fire totem, 10 s CD less 2 s a Trial by Fire point) |
| SHA-DD-06 | DD | Windfury Totem: party buff, no benefit with own Windfury Weapon, not with GoA or FT Totem | yes | 8515 A42 proc aura | landed (GoA part superseded) | Windfury Weapon block buffs.go:2001; GoA exclusivity now wrong, SHA-70425-03 |
| SHA-DD-07 | DD | Flametongue Totem: party buff, no benefit with own FT Weapon, not with WF Totem | minor | | landed | SHA-70009-05/06 |
| SHA-DD-08 | DD | Totemic Projection, Totemic Recall, Call of the Elements / Spirits / Ancestors | n/a | | n/a | presets place totems pre-pull one by one |
| SHA-DD-09 | DD | Lightning Bolt ranks 4+ cast 2.5 s, coefficient reduced to match | yes | 15208 cast 2500, coef 0.714 | landed | lightning_bolt.go:28-32 (rank 10 196 + 1.2/level at 56, x1.12 by user ruling `sham_lb_31plus`) |
| SHA-DD-10 | DD | Chain Lightning cast 2.0 s, coefficient reduced | minor | 421/930/2860/10605 cast 2000, coef .571/.571/.517/.571, CD 6000 | landed | chain_lightning.go:21, 53-54 |
| SHA-DD-11 | DD | Convection also discounts Lava Burst | yes | | landed | lava_burst.go:49 |
| SHA-DD-12 | DD | Concussion no longer applies to Frost Shock or Flame Shock | yes | | landed | talents.go:108-121 (LB, CL, ES only) |
| SHA-DD-13 | DD | Reverberation moves up | yes | | landed | shocks.go:11 (0.2 s a point) |
| SHA-DD-14 | DD | Call of Flame also boosts Flame Shock, Fire Nova and Lava Burst | yes | | landed | talents.go:136-153; Fire Nova multiplier fire_totems.go:288 |
| SHA-DD-15 | DD | Elemental Devastation moves to row 2 | yes | 70425 text "direct damage spells" | landed | talents.go:250-279 (3% melee crit a point, 10 s, Fire Nova excluded per beta) |
| SHA-DD-16 | DD | Elemental Focus triggers less on low ranks | minor | server side (16164 ProcChance 10 flat) | not simulated on purpose | flat 10%; notes.md:34, spec_assumptions.md:665 |
| SHA-DD-17 | DD | Elemental Alacrity: 3 points, 0.5 s off LB, CL and Lava Burst | yes | | landed | talents.go:183-185 (170/330/500 ms), electric_spell.go:49, lava_burst.go:53 |
| SHA-DD-18 | DD | Improved Fire Nova: more damage, lower cooldown | yes | 16086 | landed | talents.go:195-201 (10% and 2 s a point) |
| SHA-DD-19 | DD | Eye of the Storm: pushback protection on LB/CL/LvB | minor | | landed | talents.go:191 |
| SHA-DD-20 | DD | Call of Thunder: 16-point, +3% crit LB/CL | yes | | n/a (superseded) | removed in 70425, SHA-70425-15 |
| SHA-DD-21 | DD | Elemental Reach range; Earthbound; Earth's Grasp; Improved Ghost Wolf; Anticipation; Toughness; Spirit Weapons threat; Natural Grace; Improved Reincarnation; Ancestral Healing; Healing Focus; Healing Way; Riptide; Restorative Totems; Mana Tide row; Totemic Focus on Projection/Call | n/a | | n/a | range, CC, tanking, healing or tree only |
| SHA-DD-22 | DD | Lightning Overload: free half-damage copy, no threat | yes | overload copies e.g. 408477 = half of 15208 (bp 98, rppl 0.6, coef 0.357) | landed | lightning_overload.go:10-38 (3/7/10%) |
| SHA-DD-23 | DD | Elemental Fury moves | yes | | landed | talents.go:281-298 (20% crit damage a point) |
| SHA-DD-24 | DD | Lava Burst: 2.5 s, +20% with own Flame Shock, 10 s CD | yes | eff1 E3 bp 20 | landed | lava_burst.go:9, 52-59, 67-71 |
| SHA-DD-25 | DD | Elemental Mastery removed | yes | | n/a (superseded) | back in 70425, SHA-70425-12 |
| SHA-DD-26 | DD | Thundering Strikes: crit of all spells and attacks | yes | | landed | talents.go:40-41 |
| SHA-DD-27 | DD | Ancestral Knowledge raises Intellect | yes | 17485 bp 10 at 5/5 | landed | talents.go:28-30 (2% a point, multiplies Kings) |
| SHA-DD-28 | DD | Mental Dexterity: AP from Intellect | yes | 415140.. | landed | talents.go:32-34 |
| SHA-DD-29 | DD | Elemental Weapons moves up | yes | | landed | flametongue_weapon.go:59, windfury_weapon.go:32, rockbiter_weapon.go:85 |
| SHA-DD-30 | DD | Shamanistic Focus: big discount on Shocks and Lightning Shield | yes | 1223030 A108 bp -45 | landed | talents.go:166 (x0.55), shocks.go:23, lightning_shield.go:127 |
| SHA-DD-31 | DD | Flurry 5% a point up to 25% | yes | | landed | talents.go:433 |
| SHA-DD-32 | DD | Stormstrike: 16-point, 8 s CD, next LB/CL/ES +20% | yes | 17364 RecoveryTime 8000, dur 12000, A271 bp 20 mask 0x100003 | landed | stormstrike.go:15, 60-63, 96-104 |
| SHA-DD-33 | DD | Mental Quickness: spell power from Intellect | yes | 30812 bp 30 at max | landed | talents.go:36-38 (15% a point) |
| SHA-DD-34 | DD | Improved Stormstrike: casting regen and CD resets | yes | 1223031 / 1238931 | landed | talents.go:483-525 |
| SHA-DD-35 | DD | Maelstrom Weapon: stacks from melee, faster / instant LB | yes | 408498 / 408505; rate server side | landed | talents.go:527-575 (2 PPM a point assumed; notes.md:18) |
| SHA-DD-36 | DD | Rage of the Farseer: +30% attack speed 25 s, 3 min | yes | 425336 | landed | talents.go:580-623 |
| SHA-DD-37 | DD | Baseline: Enhancing Totems, 2H Axes and Maces, Weapon Mastery, Improved Weapon Totems | yes | Weapon Mastery 29082..29088 (A79 +2..10% physical) still in SkillLineAbility 373 | not simulated on purpose | Weapon Mastery off by user ruling after a beta test (beta_handoff.md:176, 254); Enhancing Totems folded into Forever totem values (buffs.go:398-404) |
| SHA-DD-38 | DD | Mindfulness: Spirit regen while casting | yes (mana) | 1223033 | landed | talents.go:65 (17/33/50%) |
| SHA-DD-39 | DD | Tidal Focus: +hit for all spells and attacks | yes (Enh water-shield / mana-tide, Elemental mana-tide) | | landed | talents.go:72-81 (1% melee and spell hit a point) |
| SHA-DD-40 | DD | Water Shield: 3 globes, mana when hit or on heal crits | minor | 408510 / 408511 bp 2 (% max mana), ICD 3500 | landed | water_shield.go (heal-crit trigger now outdated, SHA-70425-25) |
| SHA-DD-41 | DD | Tidal Mastery no longer affects Lightning spells | yes | | landed | talents.go:96-104 (heals only) |

#### 2. Unannounced client changes

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHA-U-01 | diff 70374-70425; fc lists only the rename | Trial by Fire (Improved Fire Nova) also adds 10/20 s to Magma Totem | minor (Magma presets don't take the talent) | 16086 new eff2 A107 misc 1 bp 20000 mask 0x1000, curve 126023 | in flight (worker PR) | worker fire_totems.go `20 s + improvedFireNovaMagmaDuration()` |
| SHA-U-02 | diff 70374-70425 (fc: "not in Blizzard's notes") | Lava Burst rank 1 learned at 30, cheaper and weaker | minor | see SHA-70425-13 | in flight (worker PR) | |
| SHA-U-03 | diff 70374-70425 | Lava Burst ranks 3 and 4: the Flame Shock bonus (eff1, E3 dummy, bp 20) carries EffectRealPointsPerLevel 4.2 | minor (leveling) | 1238300 / 1326047 eff1 rppl 4.2; at L60 rank 4 is at its base level, so 20%; rank 3 (now L50, cap 58) would read 20 + 4.2 x 8 = 53.6% if the server scales a dummy | ambiguous (needs verify) | sim: flat 20% (lava_burst.go:9). NTV-3 |
| SHA-U-04 | diff 70374-70425 | Elemental Reach no longer extends Flame Shock's range | n/a | 28999 eff1 A107 misc 5 mask 0x10000000 -> A4 dummy | n/a | no range model |
| SHA-U-05 | diff 70374-70425; fc "found in the game" | Fire Resistance Totem raid wide | n/a | tooltips 8184/10537/10538 | n/a | |
| SHA-U-06 | diff 70374-70425 | Earth Shock (all ranks), Frost Shock, Earthbind, Earthbound, Frostbrand Attack: SpellMisc Attributes_6 0 -> 0x200000 | n/a | attr6 bit 21 = show mechanic as combat text | n/a | cosmetic |
| SHA-U-07 | diff 70374-70425 | Mana Tide 16190 / 16191 level 25 -> 1 | n/a | talent-taught rank | in flight (worker PR) | worker `ManaTideTotemLevel[1] = 1` |
| SHA-U-08 | diff 70374-70425 | Healing Stream attr8; Shaman Pet Scaling 440595 and Shield Mastery 408525 aura 22 -> 674; Sentry Totem text | n/a | | n/a | healing / unused SoD leftovers |
| SHA-U-09 | diff 70374-70425 | Elemental Mastery ProcCharges 1 | yes | 16166 SpellAuraOptions 128637 ProcChance 100 ProcCharges 1 ProcTypeMask 0x15550 | ambiguous (needs verify) | NTV-1 |
| SHA-U-10 | diff 70170-70291 | Mana Tide 16190 / 16191 learned at 25 (was 40); aura options removed | n/a | SpellLevels 102249 / 102336 | landed | master water_totems.go:210 rank 1 at 25 |
| SHA-U-11 | diff 70170-70291 | Grounding Totem name; Ghost Wolf misc; Totemic Recall 39104 added; Clearcasting / Ancestral Fortitude descriptions | n/a | | n/a | |
| SHA-U-12 | diff 70009-70170 | Clearcasting (16246) mask: Fire Nova added (0x08000000), heal bits removed (word1 0x40000, word3 0x40000000); now LB, CL, ES, FS, Frost Shock, Fire Nova, Lava Burst | yes (Enh) | eff0 A108 misc 14 mask 0x90100003 -> 0x98100003, word1 0x41000 -> 0x1000 | landed | talents.go:256-263 (`isShamanDamagingSpell` covers exactly these; totems and heals excluded) |
| SHA-U-13 | diff 70009-70170 | Fire Nova 408341..408345 get SpellClassMask 0x08000000 (talents and Clearcasting can see them) | yes | SpellClassOptions 71108..71112 | landed | Fire Nova is `SpellFlagShaman`, ProcMaskSpellDamage |
| SHA-U-14 | diff 70009-70170 | Totem of Charged Flames (1270478, renamed Flame Charged) also discounts Flame Shock | minor | eff0 A107 bp -5 mask 0x10000001 | landed | items.go:119-130 |
| SHA-U-15 | diff 70009-70170 | Flametongue Attack 10444 / 29469 / 29470 bp 1 -> 0 (the FT Weapon tooltip's flat add-on) | minor | | landed | the sim adds no flat extra (flametongue_weapon.go:22-32) |
| SHA-U-16 | diff 70009-70170 | Revelation 1323418 added under SpellClassSet 11 | n/a here | +100% crit next non-periodic spell, 1 charge | n/a | weapon enchant proc, item scope (notes.md:128) |
| SHA-U-17 | diff 69893-70009 | Lightning Overload copies 408439..408477: Variance and RealPointsPerLevel cut | minor | 408477 = exactly half of 15208 at 70425 | landed | sim takes 0.5 x the main hit (lightning_overload.go:19); low-rank ratios in notes.md:215 |
| SHA-U-18 | diff 69893-70009 | EffectBonusCoefficient 1 -> 0 on aura effects (Charged Lightning, Totem Affinity, Enchanted Totem, Ghost Wolf Speed, Copy of Improved Mana/Health Totems) | n/a | | n/a | no effect on modifiers |
| SHA-U-19 | diff 69893-70009 | Windfury Totem 8515.. aura 4 -> 42 (dummy -> proc trigger aura) | yes | | landed | proc model buffs.go:1999-2023 |

Client-only builds: 70291 -> 70334 changes only Summon Azurky (item), 70334 -> 70338 and 70338 -> 70374 change no spell. Nothing for the shaman, confirmed. 70058, 70124, 70205, 70235 and 70245 have no DB2 snapshot of their own; the 70009 -> 70170 and 70170 -> 70291 diffs hold nothing for the shaman that the notes or this ledger do not cover.

#### 3. Proposed fixes

F1. **Fire Nova damage and coefficient (SHA-70425-07/08)**, sim/shaman/fire_totems.go:239-241 (and the comment at :225-234):
- `FireNovaBaseDamage` -> `{{0, 0}, {32.31, 37.69}, {54.28, 61.72}, {126.27, 142.73}, {173.39, 195.61}, {232.02, 258.98}}`
- `FireNovaScaling` -> `{{}, {17, 0.4}, {27, 0.6}, {37, 0.9}, {47, 1.1}, {57, 1.3}}`
- `FireNovaSpellCoeff` -> `{0, .143, .143, .143, .143, .143}`
- Mana costs as on the worker branch. Test: update TestOrcShamanFireNova (sim/shaman/enhancement/periodic_test.go:353) to rank 5 at 60 = 232.02-258.98 + 14.3% SP, and the goldens.
- Impact: the published hit (782 on enh-level60, Call of Flame 3 x Trial by Fire 2 = 1.38) implies about 690 fire SP; the new hit is about 0.6x, so Fire Nova falls from 111 to about 67 DPS: about -44 (-5%) on enhancement-level60 and -imp-gw, about -25 (-3%) on -water-shield and -mana-tide, before the cheaper cost and a rotation re-check (Fire Nova may drop down the priority). Presets: the four `shaman__enhancement-*`.

F2. **Grace of Air stacks with Windfury and Flametongue Totems (SHA-70425-03)**, sim/core/buffs.go:2006-2010: under Forever drop `aura.NewExclusiveEffect(graceOfAirWindfuryCategory, ...)` from `WindfuryTotemBuffAura` (GoA then always gives its Agility; buffs.go:1604 can keep the category, it has no other member) and rewrite the comment at :1616-1622. Test: turn TestOrcShamanGraceOfAirWithWindfury (grace_of_air_test.go:40) into "GoA's 77 Agility stays with another shaman's or our own Windfury Totem". Impact today: 0 (no preset has both raid buffs; a shaman's own WF is blocked by its Windfury Weapon). But the 26 melee presets that carry `totemWeaponBuff: Windfury` (rogue 12, paladin 9, warrior 7, druid 2, hunter survival 2; all already carry Strength of Earth) could now also carry `graceOfAirTotem` from a second raid shaman: +77 Agility, roughly +2 to +5%. That is a composition ruling for the user (NTV-2), shared with the other class owners.

F3. **Stormstrike mark kept on a miss (SHA-70425-20)**, sim/shaman/stormstrike.go:117-131: split `spendStormstrike` into a peek that returns 1.2 without deactivating, and consume the mark only when `result.Landed()` (lightning_bolt.go:86-89, chain_lightning.go:84-87, earth_shock.go:84-87). Make the overload copy not spend it (it runs the same ApplyEffects, and after the fix would find a mark left by a missed bolt; NTV-4). Test: in stormstrike_test.go, a forced full resist on Earth Shock leaves the mark for the next Lightning Bolt. Impact: about 2.5 marked spells a 180 s fight miss today, +2 DPS (+0.2%) on the four Enhancement presets.

F4. **Water Shield on spell crits (SHA-70425-25)**, sim/shaman/water_shield.go:62-66: add `OnSpellHitDealt` that calls `consumeGlobe` on `result.DidCrit()` for harmful non-periodic spells (`ProcMaskSpellDamage | ProcMaskSpellDamageProc`, so Flametongue procs count; client flag 0x10000 has no periodic bit). Test: TestOrcShamanWaterShield (mana_returns_test.go:50) case with Earth Shock crits and the 3.5 s cooldown. Impact on published presets: 0 (none casts Water Shield); re-run enhancement-water-shield with Water Shield up to see whether its "keeps Lightning Shield" choice still holds.

F5. **Flametongue Totem by the shapeshifted swing speed (SHA-70291-01)**, core/totem_weapon_buffs.go:107-111: use `character.AutoAttacks.MH().SwingSpeed` / `OH()` (the cat's 1.0) in place of the item's. Test: a Forever cat with FT Totem raid buff procs `21.92/4 x 1.0` per swing. Impact 0 (feral presets take Windfury).

F6. **Worker branch (in flight), to finish**: the Elemental presets on the worker branch spend 50 of 51 points (`550433100000135-052-05305`) and do not take Elemental Mastery; decide its charges (NTV-1) and the 51st point, and APL use of 16166. Call of Thunder removal (-3% LB/CL crit, about -1 to -2% on both Elemental presets) and Call of Flame's +3 Flame Shock ticks (Flame Shock is 13.6% of Elemental DPS and 7.4% of Enhancement) are the large in-flight moves.

#### 4. Need to Verify candidates

* **NTV-1. Elemental Mastery's charges (client 70425, 16166).** The notes and the tooltip say Lightning Bolt and Lava Burst cast 0.5 sec faster and Chain Lightning has no cooldown "for 15 seconds", 3 min cooldown. The client aura has a 15 sec duration but also ProcCharges 1 with proc flags 0x15550 (spell casts), Era's "next spell" setup, so it may be spent by the first Lightning Bolt or Lava Burst, which would make it worth one 0.5 sec cut instead of about five. The worker models 15 sec with no charges. A beta test: press it and cast two Lightning Bolts, and watch whether the buff stays and the second cast is 0.5 sec shorter.

* **NTV-2. Grace of Air with Windfury Totem from two shamans (client 70425).** Grace of Air lost its exclusivity bit (Attributes_11 1024 -> 0 on 8836 / 10626 / 25360) and all four buff totems became raid area auras, while Windfury and Flametongue Totems keep the bit. So a raid with two shamans gives the melee both GoA's 77 Agility and Windfury Totem. The sim still blocks GoA under any Windfury Totem buff (buffs.go:2009). Once fixed, should the melee presets (rogue, warrior, Ret, Feral, Survival) carry both, as the raid has several shamans on both factions? A user ruling, roughly +2 to +5% for those presets. Also open: does a second shaman's copy of the same totem (two SoE) do anything (notes.md:17).

* **NTV-3. Lava Burst's Flame Shock bonus by level (client 70425).** The bonus is a dummy effect, bp 20, and ranks 3 (1238300) and 4 (1326047) carry RealPointsPerLevel 4.2 on it. If the server scales it, rank 3 (learned at 50, capped at 58) gives 53.6% at 58 and rank 4 stays 20% at 60. Level 60 is unaffected; leveling between 50 and 59 would be. A Lava Burst with and without Flame Shock at level 51 to 58 settles it.

* **NTV-4. Stormstrike and Lightning Overload (70425 notes).** The mark is no longer spent by a spell that does not land. Does an Overload copy spend it, or get the 20%? The sim's overload runs the bolt's own effects; today the main bolt always spends the mark first, but after F3 a missed bolt would leave it for its overload. Also: does a partially resisted spell count as landed (the sim says yes).

* **NTV-5. Water Shield's new trigger (70425).** The proc flags gained 0x10000 (harmful magic spells), so any damage spell crit should spend a globe. Do Flametongue Weapon and Flametongue Totem procs and Lightning Overload copies count, and does the 3.5 sec globe cooldown still hold? Ten minutes of Enhancement play with Water Shield up, counting globes against spell crits, settles it.

* **NTV-6. Flametongue Totem on a cat (70291 notes).** The notes say it uses a shapeshifted target's attack speed; the sim uses the equipped weapon's. Only matters if a Feral preset ever takes Flametongue Totem.

#### 5. Count

Rows by status (sections 1 and 2, 102 rows): landed 57, in flight (worker PR) 12, missing 4, wrong 2, not simulated on purpose 2, ambiguous (needs verify) 2, n/a 23 (3 of them superseded by a later build).

### Mage: game-update audit ledger (builds 70009 to 70425)

Scope: sim/mage, presets tools/dps_report/req/mage__{arcane,fire,frost,frostfire}.json. Master = origin/master 6a87c9c. Worker branch = `origin/claude/wow-forever-dps-presets-m9u15h-mage` (B below), commits ce0438056 "port to client 1.60.1.70425 (sim, tree, templates; presets to follow)" and bef7804e4 (Burning Soul / Improved Channeling pushback). B touches arcane_blast.go, arcane_missiles.go, new arcane_wrath.go, fire_blast.go, ice_lance.go, mage.go, talents.go, forever_test.go, proto/mage.proto, the talent tree JSONs, rotopt templates and forever-wiki/classes/mage.md. **B does not touch any tools/dps_report/req/mage__* file, notes.md, or Heating Up.**

Sources: devnotes post 1 (https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/1), post 4 (.../4), post 5 (.../5), post 6 (.../6); deep dive https://worldofwarcraft.blizzard.com/en-us/news/24302097 (published 2026-10-07, so it describes ~70170 design); fc_patch-notes (no mage line marked "Found in the game / Not in Blizzard's notes"). Client values from S/db2 CSVs and S/diffs.

Variance convention (sim and worker): damage = base * (1 +/- Variance/2) at the level cap.

#### 1. Announced changes

##### Build 70425 (post 6, 2026-10-15)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| MAGE-70425-01 | post 6, 10-15, 70425 (two lines: Frost section and "New spell Frost Blast") | Frost Blast: new spell at 24, 1.5 s cast, 40% slow, "affected by all talents that affect Frostbolt", more mana and a little more damage than R1 Frostbolt | minor | Spell 1326053/54/55 ADDED (lvls 24/36/48). R3: base 77 +0.6/lvl to 60 -> 84.2, Variance 0.105 -> 79.8-88.6, coef 0.343, CastTimeIndex 16 (1.5 s), 90 mana, Duration 9 s, slow -40. Improved Frostbolt (11070) and Arctic Reach tooltips now name it | not simulated on purpose (B wiki) | No frost_blast.go on master or B. B's forever-wiki/classes/mage.md: "damage per second of cast is under half of Frostbolt's, so the sim leaves it out". Not counted: it is a Chill (Fingers of Frost rolls), a Winter's Chill applier and gets Improved Frostbolt's -0.5 s (1.0 s cast, GCD-capped 1.5 s). See fix F. |
| MAGE-70425-02 | post 6 | Presence of Mind is consumed by spell casts (bug fix) | yes | server side | landed | sim always consumed it: talents.go:343-348 (PoM OnCastComplete deactivates on any spell with a cast time). |
| MAGE-70425-03 | post 6 | Missile Barrage no longer reduces cost/duration of a second Arcane Missiles | yes | server side | landed | arcane_missiles.go:100-102 deactivates MissileBarrageAura when the channel starts. |
| MAGE-70425-04 | post 6 | Missile Barrage moves to row 1, 2 points; Arcane Missiles hits can also trigger it | yes | 400588 SpellAuraOptions ProcChance 50 -> 100; TraitNodeEntry 130531 MaxRanks 1 -> 2; new TraitDefinitionEffectPoints 25203 (curve 126757); effect0 base 40 (AB chance at max rank), tooltip: AM hit m1/4, Arcane Wrath/FB/Frostbolt/FFB m1/2. At 2/2: AB 40%, AM hit 10%, others 20% | in flight (worker PR) | B talents.go:271-335 (per point AB 0.20, AM tick 0.05 on landed hit, AW/FB/Frostbolt/FFB 0.10); B tree arcane idx 3 maxPoints 2; proto `int32 missile_barrage = 4`. Master still bool, 40/20%, no AM-hit roll (talents.go:264-306). |
| MAGE-70425-05 | post 6 | Arcane Blast redesign: 1.5 s cast; each cast +15% of Int as spell damage, stacks 4, 15 s from first application (not refreshed, not ended by other spells); cost growth kept; damage and coefficient -40% | yes | 400574..1239700: coef 0.714 -> 0.429; CastTimeIndex 19 -> 16 (1.5 s). R5 1239700 base 394 -> 236 (+1.6/lvl, base lvl 60), Variance 0.151 -> 218.2-253.8. 400573 effect0 aura 108 -> 174 (spell dmg % of stat), base 10 -> 15, MiscValue 126 (all magic schools) / stat 3 (Int); effect1 +175% AB cost kept; ProcCharges 1 -> 0, ProcTypeMask -> 0; DurationIndex 31 -> 8 (15 s); 2nd damage effect (row 1120352) REMOVED | in flight (worker PR) | Master wrong: arcane_blast.go:26 `ArcaneBlastStackBonus = 0.10`, :48 8 s, :98 2500 ms, :104 0.714. B arcane_blast.go: 218.18-253.82, 1.5 s, 0.429, `ArcaneBlastIntellectPerStack = 0.15` via dynamic Int->SpellPower stat deps, 15 s, Activate only if not active, AM bonus code removed. Verified values match client. |
| MAGE-70425-06 | post 6 | Arcane Wrath: new 16-point gold medal talent, Fireball-like base, 3.5 s cast, +spell power = 5% of max mana, 6 s CD | yes | 1326079-82 ADDED, ranks at 25/36/48/60. R4: base 451 (+3/lvl), Variance 0.201 -> 405.7-496.3, EffectBonusCoefficient 1, effect1 dummy 5 (%max mana), CastTimeIndex 22 (3.5 s), RecoveryTime 6000, 395 mana, Speed 24. TraitNodeEntry 141289 added (MaxRanks 1). Arcane Focus/Geometry/Impact/Mind/Instability/Power and Improved Channeling masks gain SpellClassMask_1 32768 (AW) | in flight (worker PR) | B sim/mage/arcane_wrath.go (405.67-496.33, 395 mana, 3.5 s, 6 s CD, coef 1.0, +0.05*MaxMana*coef added to base, MissileSpeed 24, SpellFlagMage so school/all-spell talents apply). B rotopt mage_arcane.go casts it (on CD, PoM'd). Coefficient 1.0 is the client's raw value: NtV 4. |
| MAGE-70425-07 | post 6 | Heating Up is no longer consumed if Presence of Mind was active for the Pyroblast | minor | 400625 SpellAuraOptions ProcTypeMask_0 65536 -> 87376, Attributes_0 0 -> 262144, Attributes_6/7 cleared (server decides the consumption) | missing (missed by worker) | Master talents.go:520-524 and B talents.go:601-605 deactivate Heating Up on every Pyroblast cast. No preset has PoM + Heating Up today (Arcane has PoM, no Fire tree; Fire/Frostfire have no PoM). Fix B. |
| MAGE-70425-08 | post 6 | Heating Up always gains an extra charge each time the mage casts Fire Blast (a Fire Blast crit gives 2) | yes | 400624 tooltip adds "Fire Blast grants an additional stack every time it deals damage" (server side; no new effect row) | missing (missed by worker) | Master talents.go:528-537, B talents.go:608-620: only crits of FB/FFB/Fire Blast/Scorch add a stack. B forever_test.go TestMageHeatingUp tests the old rule. Fire and Frostfire templates cast Fire Blast on CD (B rotopt mage_fire.go:60-65, mage_frostfire.go:80-81). Fix A. |
| MAGE-70425-09 | post 6 | Wake of Fire reverted to Improved Fire Blast: no kill bonus, +8/15 yd Fire Blast range (CD reduction kept) | yes (CD) | 11078 renamed; effect1 aura 4 base 50 -> aura 107 base 15, MiscValue 5 (range), mask Fire Blast; ProcTypeMask row removed. Effect0 -2000 ms (CD) unchanged | in flight (worker PR) | Master fire_blast.go:68 already takes 1 s/point (`Talents.WakeOfFire`), kill bonus never simulated. B renames to `ImprovedFireBlast` (fire_blast.go, proto field 20). Fire meta keyTalents still say "Wake of Fire". |
| MAGE-70425-10 | post 6 | Critical Mass: up to 5 points, 10% total | yes | TraitNodeEntry 130513 MaxRanks 3 -> 5; 11115 base 2 (aura 107 crit, Fire mask) | in flight (worker PR) | Code is per point (master talents.go:129-135, B :152-158, 2% * points); B tree maxPoints 5. Presets still 3 points (Fire, Frostfire): fix C. |
| MAGE-70425-11 | post 6 | Fire Power redesign: 3 points, +1/2/3% Fire damage per mage Fire effect on target, max 12% (4 effects); mixed-school effects don't count; effects = Ignite, Imp. Scorch, Fireball, Pyroblast, Impact; FFB benefits but doesn't count | yes | TraitNodeEntry 130511 MaxRanks 5 -> 3; 11124 effects 0/1 aura 108 (flat % on masked spells) -> aura 4 dummy, base 2 -> 3, masks cleared; TraitDefinitionEffectPoints 24560 removed (server-side mechanic) | in flight (worker PR) | Master wrong: talents.go:137-145 flat +2%/point additive on all Fire spells but Ignite. B talents.go:477-545: per point 1% * count of {Fire Vulnerability, active DoTs of the mage's pure-Fire spells: Fireball, Pyroblast, Flamestrike, Ignite} capped 4, multiplied on the attack table with Fire Vulnerability; Ignite gets none; FFB direct takes it, its DoT doesn't count. Impact (stun) not counted, boss is immune: fine. Flamestrike counts though the notes don't list it (no preset casts it): NtV 3. |
| MAGE-70425-12 | post 6 | Frost Channeling now reduces threat; Ice Block transparency | n/a | - | n/a | threat / visual. |
| MAGE-70425-13 | post 6 | Clearcasting: no roll per target hit; not from Dampen/Amplify Magic; not out of combat | minor | server side | landed | talents.go:234-256: one 1 s ICD on Arcane Concentration, so an AoE cast rolls once; Dampen/Amplify and out-of-combat casts are not in any rotation. |
| MAGE-70425-14 | post 6 | Frostbite and Fingers of Frost can trigger from Ice Armor / Frost Armor | n/a | 11071 Frostbite aura 109 -> 42 (proc trigger), new ProcTypeMask 65536; Frost/Ice Armor effects aura 22 -> 674 | not simulated on purpose | Needs melee attackers on the mage. Raid hits (core/raid_hits.go) are boss Fire spells, not melee; a boss can't be frozen. |
| MAGE-70425-15 | post 6 | Ice Lance vs Frozen: +200% (was +300%) | yes | 400640, 1240044-47, 1312002 effect1 base 300 -> 200 (x3 total) | in flight (worker PR) | Master wrong: ice_lance.go:29 `IceLanceFrozenMultiplier = 4`. B ice_lance.go:29 = 3. Frost and Frostfire presets Ice Lance on FoF. |
| MAGE-70425-16 | post 6 | Cone of Cold: first 4 targets slowed 10% more for 3 s longer; visual moved | n/a | 10161 effects 2/3/4 ADDED (4, -10, 3) | n/a | No Cone of Cold in the sim (single-target slow). |
| MAGE-70425-17 | post 6 (two lines) | Winter's Chill redesign: no crit; Frostbolt / Frost Blast / FoF Ice Lance 20-100% chance to apply, stacks 5, consumed by next Cone of Cold; line A: +40% CoC damage per stack; line B: +5% damage and +7% spell damage per stack | n/a (was yes) | 12579 aura 308 (crit) -> 271 (dmg taken from caster), base 2 -> 40, mask 131104 (Frostbolt+IL) -> 512 (Cone of Cold); CumulativeAura 5; 11180 unchanged (base 5 = stacks) | in flight (worker PR) | Master wrong: talents.go:664-705 + personalCritRating still give +2% crit per stack to Frostbolt/IL. B removes applyWintersChill and the crit (CoC not simulated). Frost preset spends 5 points in it (dead after the port): fix C. Value conflict 40% vs 5%+7%: NtV 1. |

##### Build 70291 (post 5, 2026-10-08)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| MAGE-70291-01 | post 5, 10-08, 70291 | Impact can't proc from Flamestrike's DoT | n/a | server side | not simulated on purpose | Impact is a stun, notSimulated in the tree (raid boss immune). |
| MAGE-70291-02 | post 5 | Frostbolt ranks 1-5 base/per-level reworked | n/a (leveling) | 116/837/205/7322/8406 e.g. 8406 base 95 -> 104, 1.5 -> 1.3/lvl | landed | frostbolt.go:16-17 (R5 104.24-114.16, {30, 1.3}); data from 70338. |
| MAGE-70291-03 | post 5 | Fireball ranks 1-5 reworked | n/a (leveling) | 8400 base 108 -> 120, 1.5 -> 1.4/lvl | landed | fireball.go:19-20 (R5 107.11-144.09, {28, 1.4}). |
| MAGE-70291-04 | post 5 | Arcane Missiles ranks 1-3 reworked | n/a (leveling) | 7270 base 44 -> 47, 0.5 -> 0.4/lvl | landed | arcane_missiles.go:20 ({28, 0.4}). |

##### Build 70170 (post 4, 2026-10-01)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| MAGE-70170-01 | post 4, 10-01, 70170 | Comprehension scrolls: balance pass, not castable while moving | minor | Comprehend Scroll spells attributes | landed | forever_staff_imbue.go uses 70338 values (after the pass); moving n/a. |
| MAGE-70170-02 | post 4 | Hot Streak renamed Heating Up | yes | 400624/400625 Name Hot Streak -> Heating Up | landed | talents.go:497-540. |
| MAGE-70170-03 | post 4 | Combustion back to 3 charges (was 4) | yes | 11129 ProcCharges 4 -> 3 | landed | talents.go:560 `const maxCrits = 3`. |
| MAGE-70170-04 | post 4 | Improved Scorch's Fire Vulnerability doesn't roll a second resist | yes | 22959 Attributes_3 change | landed | scorch.go:71-77 (applies on a landed Scorch, no own hit roll). |
| MAGE-70170-05 | post 4 | Winter's Chill application doesn't roll resist | n/a now | 12579 Attributes_3 128 -> 262272 | landed (superseded) | master applied on landed Frost hits with no own roll; redesigned in 70425 (MAGE-70425-17). |
| MAGE-70170-06 | post 4 UI note | Low-rank spells: less spell power and lower talent proc chance (R1 Frostbolt 0% Frostbite) | n/a | server side | n/a | presets cast max ranks only. |
| MAGE-70170-07 | deep dive 10-07 (~70170) | Baseline/talent design statements: Frostfire Bolt (both schools, higher SP, lower resist), Mage Armor regen up, Arcane Focus = hit, Improved Channeling AB protection, Arcane Subtlety, Arcane Concentration low-rank penalty, Arcane Impact all Arcane crit, Arcane Mind Int + Arcane crit dmg, Arcane Instability, Incineration (+IL, AB), Improved Fireball (+FFB), Ignite per caster + rolling, Burning Soul 3 pts, Improved Scorch own-mage only, Heating Up, Elemental Precision = hit, Ice Lance, Cold Snap 10 min, Fingers of Frost, Shatter | yes | (70338 data) | landed | talents.go:44-49 (Arcane Focus hit), :65-71 Arcane Impact, :78-86 Arcane Mind (2% Int, +20% crit bonus/pt), :88-99 Instability, :140-150 Incineration incl. IL/AB, :624-650 Cold Snap 10 min, ignite.go per-mage rolling, scorch/ImprovedScorchAuras per mage, frostfire_bolt.go:59 Improved Fireball, armors.go:78 Mage Armor 50%. Improved Channeling / Burning Soul pushback land on B only (bef7804e4); they matter since #55 raid hits push back casts: in flight. |
| MAGE-70170-08 | deep dive 10-07 | Arcane Blast (2.5 s, +dmg to other spells, 8 s), Missile Barrage 16-pt 1-point talent, Wake of Fire kill bonus, Winter's Chill Frostbolt/IL crit | - | - | n/a (superseded by 70425) | see MAGE-70425-04/05/09/17. |

##### Build 70009 (post 1, 2026-09-24)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| MAGE-70009-01 | post 1, 09-24, 70009 | Arcane Missiles checks LoS once | n/a | AM spells Attributes_2 +4 | n/a | movement/LoS. |
| MAGE-70009-02 | post 1 | Wake of Fire buff 30 s (was 20) | n/a | 1312934 DurationIndex 18 -> 9 | n/a | kill bonus never simulated; removed in 70425. |
| MAGE-70009-03 | post 1 | Hot Streak (Heating Up) buff 20 s (was 15) | yes | 400625 DurationIndex 8 -> 18 (20 s) | landed | talents.go:510 `Duration: time.Second * 20`. |
| MAGE-70009-04 | post 1 | Ignite no longer double dips on % damage modifiers | yes | 412545 Attributes_6/10 | landed | ignite.go:35 `SpellFlagIgnoreModifiers` (skips attacker and target multipliers, incl. the attack table hook). |
| MAGE-70009-05 | post 1 known issues | Arcane Blast / mana gem not tracked by Cooldown Manager | n/a | - | n/a | UI. |

#### 2. Unannounced client changes (mage-tagged, plus "No class" grep)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| MAGE-U-01 | diff 70374-70425 | Missile Barrage ProcChance 50 -> 100 (old notes.md item "ProcChance 50") | yes | 400588 SpellAuraOptions 222941 | in flight (worker PR) | B uses the tooltip numbers at 100%; notes.md item is now settled but B didn't edit notes.md (fix D). |
| MAGE-U-02 | diff 70374-70425 | New buff spell 1326083 "Missile Barrage" (dummy copies of 400589: -50% duration, -500 period, -100% cost, 15 s) | n/a | 1326083 ADDED | n/a | Visual/tracking copy; 400589 effects unchanged. |
| MAGE-U-03 | diff 70374-70425 | Arcane Blast 400573 180 s SpellCooldowns row removed | n/a | row 83253 removed | n/a | aura only. |
| MAGE-U-04 | diff 70374-70425 | Attributes_6 bit 2097152 added to Frostbolt, FFB, Frost Nova, Cone of Cold, Blast Wave, Chilled, Polymorph, Counterspell, Ice Block, Impact, Frostbite | n/a | SpellMisc Attributes_6 0 -> 2097152 | n/a | Unknown flag, likely CC/chill tagging; no damage field moved. |
| MAGE-U-05 | diff 70374-70425 | Improved Frostbolt tooltip now covers Frost Blast; Arctic Reach likewise | minor | 11070, 16757 Description only | not simulated on purpose | follows Frost Blast (70425-01). |
| MAGE-U-06 | diff 70374-70425 | Heating Up 400625 ProcTypeMask 65536 -> 87376 and attribute changes | minor | 223474 | missing (missed by worker) | the client side of 70425-07/08; consumption logic is server side. |
| MAGE-U-07 | diff 70374-70425 | Frost Armor 12556 DurationIndex 21 -> 0; armors' aura 22 -> 674 | n/a | SpellMisc / SpellEffect | n/a | defensive / proc on melee attackers. |
| MAGE-U-08 | diff 70374-70425, tagged Mage | Firebolt 18833 (Alcor's Sunrazor proc): base 90 Var 0.333 -> 105 Var 0.20 (75-105 -> 94.5-115.5); 8 s category CD row removed | minor (item) | SpellEffect 696289 | wrong (item scope) | sim/common/item_effects.go:189 still `..., 18833, Fire, 75, 30, ...` (75-105). Hand to the item auditor. |
| MAGE-U-09 | diff 70009-70170 | Revelation 1248808 (Enchant Weapon - Revelation buff): aura 290 (all crit) -> 107 (+100% crit on next non-periodic spell, 1 charge) | n/a for mage presets | SpellEffect 1251976 | n/a (item/shaman scope) | sim/common/enchant_effects.go:11-21; mage presets don't use it (notes.md). |
| MAGE-U-10 | diff 69893-70009 | Fire Blast 461896 (70 dmg, lvl 60, 115 mana; not the trainer spell) coef 0.15 -> 0 | n/a | SpellEffect 1170810 | n/a | Not a player trainer rank (10199 is R7, unchanged); no sim reference. |
| MAGE-U-11 | diff 69893-70009 / 70009-70170 | Eureka! (1259817) -50 -> -10 cost, masks; 70425: 8 s, no charges | yes (racial) | SpellEffect 1269077; SpellAuraOptions removed; DurationIndex 8 -> 31 | landed | shared port #50 (racials), not in mage scope. Touch of the Grave (Undead mage presets) likewise #50. |
| MAGE-U-12 | all diffs | Low-rank Frostbolt/Fireball/AM data, Polymorph MiscValue, Blink variants, DNT teleports, Chilled 1323731, NPC Fireball/Frostbolt variants, "No class" mage-named NPC spells | n/a | - | n/a | none at L60 ranks. |
| MAGE-U-13 | client-only builds | 70291->70334: only Summon Azurky (item); 70334->70338 and 70338->70374: no spell changes, no mage TraitNodeEntry rows | n/a | diffs | confirmed: nothing for mage | - |
| MAGE-U-14 | 70009/70170/70291 trait tables | TraitNodeEntry/TraitDefinitionEffectPoints changes in those builds map to druid/warrior/warlock/priest defs only | n/a | 134411/12, 145858/59, 135412, 135469, 135484, 145863/64, 142603/04 | n/a | no mage talent rank changes before 70425. |

#### 3. Proposed fixes

A. **Heating Up: Fire Blast always adds a stack (MAGE-70425-08), missed by worker.** On B, `sim/mage/talents.go` `applyHeatingUp`, "Heating Up Talent" OnSpellHitDealt (B :610-619): handle `SpellCode_MageFireBlast` before the crit check: if `result.Landed()`, Activate + AddStack once; then the existing crit path adds the second. Keep FB/FFB/Scorch crit-only. Test: extend `TestMageHeatingUp` (forever_test.go:46): a non-crit Fire Blast gives 1 stack, a crit Fire Blast 2. Presets: mage__fire, mage__frostfire (both cast Fire Blast on CD and Pyroblast on stacks). Impact guess: about +1 to +4% Fire, +1 to +2% Frostfire (one guaranteed stack every 6 s with Improved Fire Blast 2/2); retune the rotopt pyroStacks / fireBlastHeating knobs after.

B. **Heating Up kept when Presence of Mind makes the Pyroblast (MAGE-70425-07), missed by worker.** Same function, Heating Up aura OnCastComplete (B :601-605): don't deactivate when PoM was active for that cast. Aura callback order is not guaranteed, so have PoM's OnCastComplete (talents.go:343) record `mage.pomUsedAt = sim.CurrentTime` before deactivating, and skip HU's deactivation when `PoM aura active || mage.pomUsedAt == sim.CurrentTime`. Test: PoM + Pyroblast with 2 stacks leaves 2 stacks. No preset affected today (no build has both); 0 DPS on the report.

C. **Port the four presets (worker: "presets to follow").** On B the master talent strings decode wrongly or are invalid: Arcane tree gained Missile Barrage at idx 3 and Arcane Wrath at idx 14 (every Arcane string shifts: mage__arcane `051225003100311531`, mage__fire `05020500300001`); Fire Power now caps at 3 (mage__fire has 5, mage__frostfire 4); Critical Mass now 5 (Combustion still needs it maxed); Missile Barrage is 2 points (mage__fire took 1); mage__frost has 5 points in Winter's Chill, which now only buffs Cone of Cold (dead points). Also rename keyTalents "Wake of Fire" -> "Improved Fire Blast" in mage__fire.meta.json and drop "Winter's Chill" from mage__frost.meta.json; give Arcane a build with Arcane Wrath. Then rerun rotopt and the dps-report for all four. Impact: all four presets move (Arcane redesign, Ice Lance x4 -> x3 on Frost/Frostfire, Fire Power/Critical Mass on Fire/Frostfire, Frost loses Winter's Chill crit: probably several % down unless respecced).

D. **notes.md Mage section is stale after 70425** (B does not edit it): remove or rewrite "Arcane Blast's bonus on Arcane Missiles" (bonus gone), "Missile Barrage's 400588 has ProcChance 50" (now 100), "Ice Lance against a Frozen target (sim: 4 times)" (client 200 -> x3), "Arcane Blast's and Arcane Power's periodic damage masks" (AB has no damage aura now), "Ice Lance's two other dummies ... Winter's Chill" (Winter's Chill is a CoC debuff now). Add items 1-5 below.

E. **Item scope: Alcor's Sunrazor Firebolt 18833** (MAGE-U-08): sim/common/item_effects.go:189 `75, 30` -> `94.5, 21` (client 70425 base 105, Var 0.20). Belongs to the item auditor; not a mage preset item.

F. **Optional: Frost Blast check (MAGE-70425-01).** Not a DPS spell by itself (R3 ~84 + 34.3% SP per 1.5 s vs Frostbolt ~535 + 81.4% per 2.5 s), but it is a Chill for Fingers of Frost on a 1.5 s GCD and Ice Lance on FoF is x3 with Shatter. A quick rotopt pass with a Frost Blast filler on mage__frost would close it; implement as frost_blast.go with Frostbolt's talent hooks (Improved Frostbolt -0.1 s/pt, Piercing Ice, Ice Shards, Frost Channeling, FoF chill roll) only if it gains.

#### 4. Need to Verify candidates

1. **Winter's Chill per-stack value (sim: not modelled, Cone of Cold isn't simulated).** The 70425 notes say two things: "Cone of Cold deals 40% increased damage per stack" and "increasing the target's damage taken from your next Cone of Cold by 5% and it gains an additional 7% of your spell damage, stacking up to 5 times"; foreverchanges follows the second. Client 12579 is aura 271 (damage taken from the caster) at 40 with the Cone of Cold mask and 5 stacks, which supports the first. Only matters for AoE/Cone of Cold builds. Beta (needs 25+ Frost points): Cone of Cold on a dummy with 0 and 5 stacks.

2. **Missile Barrage per-rank split (sim on B: per point AB 20%, each landed missile 5%, Arcane Wrath/Fireball/Frostbolt/Frostfire Bolt 10%).** The 70425 client gives 400588 base 40 with a 2-rank trait curve (126757, not downloaded) and ProcChance 100 (was 50, which settles the old notes.md item). The tooltip formulas are m1, m1/4, m1/2. Whether rank 1 is half of rank 2, and whether the missile roll is per hit or per channel, are server side. A log of 100 Arcane Missiles channels at 1/2 and 2/2 settles it.

3. **Fire Power's effect list and stacking (sim on B: Fire Vulnerability counts once regardless of stacks, plus the mage's active Fireball, Pyroblast, Flamestrike and Ignite DoTs, cap 4; multiplied with Fire Vulnerability on the attack table).** Blizzard lists Ignite, Improved Scorch, Fireball, Pyroblast and Impact, not Flamestrike; the client turned the talent into a dummy (aura 4), so the server counts. Whether several Fireball DoTs from different ranks count separately, whether Fire Vulnerability counts per stack, and whether the bonus adds to or multiplies the other % modifiers are open. Beta: Scorch damage with 0/1/2/3 Fire effects up.

4. **Arcane Wrath's coefficient and its mana bonus (sim on B: 100% of spell power, and 5% of max mana added as spell damage through the same 100%).** The client effect carries EffectBonusCoefficient 1, the value earlier builds zeroed on several non-damage spells as a placeholder (Chaos Fire, Frosty Zap, Improved Counterspell in 70009), though 1.0 is also the classic 3.5 s cast value. Two Arcane Wrath hits at different spell power (and at different max mana, e.g. with and without Arcane Intellect) settle both. It decides Arcane Wrath's place in the Arcane rotation.

5. **Arcane Blast stack timing (sim on B: the stack lands after the hit, so a cast doesn't use its own new 15% of Intellect; the buff keeps the first application's 15 s and later casts don't refresh it).** Server side. A 5-cast AB chain timed past 15 s, with the spell damage readout, settles both; Heating Up's Fire Blast stack (fix A) similarly: does a missed Fire Blast give it ("every time it deals damage")?

#### 5. Count

Announced rows (34, some n/a lines merged): landed 14 (incl. 1 superseded), in flight (worker PR) 8, missing 2 (both Heating Up, missed by worker), not simulated on purpose 3, n/a 7, wrong 0 on B (master is wrong on the 8 in-flight rows until B merges). Unannounced rows: 14 (in flight 1, missing 1, wrong 1 item-scope, not simulated on purpose 1, landed 1, n/a 9). Presets: all 4 still on the 70338 tree on both master and B (fix C).

### Warlock and warlock pets: game-update ledger (builds 70009 to 70425)

Scope: sim/warlock (spells, talents, demons), tools/dps_report/req/warlock__*.json (8 presets: affliction, demonology, destruction, destruction_fire, each plus a duty-curse twin). Checkout origin/master 6a87c9c. Worker branch: origin/claude/wow-forever-dps-presets-m9u15h-warlock (commits fb88b9332, 56cff065f).

Links used below:
- P1 = devnotes post 1, 2026-09-24, build 70009, https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/1
- P4 = post 4, 2026-10-01, build 70170, .../2360696/4
- P5 = post 5, 2026-10-08, build 70291, .../2360696/5
- P6 = post 6, 2026-10-15, build 70425, .../2360696/6
- DD = Rogue and Warlock deep dive, https://worldofwarcraft.blizzard.com/en-us/news/24310968, published 2026-10-07 (so it describes the 70170 design; two of its statements, Malediction and Shadow Mastery, are superseded by 70425)
- FC = foreverchanges.pro/patch-notes

Preset talent use on master (proto order, sim/warlock talents): Malediction (field 4) is 3 in affliction, 1 in destruction, 5 in destruction_fire, 0 in demonology. Shadow Mastery (field 16) is 5 in affliction, 0 elsewhere. Races: Gnome on affliction, demonology, destruction; Undead on destruction_fire.

#### 1. Announced changes

##### Build 70425 (P6, 2026-10-15)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WL-70425-01 | P6 | Malediction replaced with Grim Reach (Affliction spell range +10/20%) | yes | TraitDefinition 135454 SpellID 1225177 -> 18218; TraitNodeEntry 130653 (node 105923) MaxRanks 5 -> 2; Grim Reach 18218 e0 aura 108 misc 5 (range) bp 20; TraitDefinitionEffectPoints 24355 (Malediction e1) removed. Sim should give 0 damage. | in flight (worker PR) | Master still has it: sim/warlock/talents.go:147-157 `applyMalediction`, `bonus := 0.01 * float64(warlock.Talents.Malediction)` on every DoT plus Immolate's ticks; proto/warlock.proto:11 `malediction = 4`. Branch deletes it, renames the field `grim_reach = 4` and re-searches the talents (Malediction points moved to Nightfall / Improved Bane of Agony / Malevolence / Soul Harvest). |
| WL-70425-02 | P6, FC | Shadow Mastery 2/4/6/8/10% Shadow damage (was 1/2/3/4/5%) | yes (affliction 5/5) | SpellEffect 18271 e0 (misc 0, direct) and e1 (misc 22, periodic) bp 2 unchanged; ranks come from curves 83215 / 88067 (CurvePoint not in this download; the worker read CurvePoint and found 2% a rank). Masks: direct = Shadow Bolt, Shadowburn, Death Coil; periodic = Corruption, Banes, Drain Life/Soul, Siphon Life, Wrack. Sim should use 0.02 a point. | in flight (worker PR) | Master sim/warlock/talents.go:258 `bonus := 0.01 * float64(warlock.Talents.ShadowMastery)`. Branch: `0.02` plus TestWarlockShadowMastery in sim/warlock/dps/mechanics_test.go. Spell lists at talents.go:21-27 match the client masks. |
| WL-70425-03 | P6, FC | Destructive Reach: range of Destruction spells only | n/a (range) | 17917 e0 mask 4769791,327747,2 -> 997,320,0; e1 (radius +10%, mask_1 65536) removed | n/a | Not modelled (range). |
| WL-70425-04 | P6, FC | Soul Link can be cast on the pet again | yes (demonology APL casts Soul Link) | 19028 SpellLevels BaseLevel 40 -> 0, SpellLevel 40 -> 1; 25228 unchanged: aura 79 +3% all schools, aura 81 30% | landed | Sim never had the bug: sim/warlock/talents.go:468-522 (`DamageDealtMultiplier *= 1.03`, damage taken /1.3), cast in ui/warlock/apls/level60_demonology*.apl.json. |
| WL-70425-05 | P6, FC | Summon Incubus no longer free with Reagent Economy perk; Curse of Recklessness can't be on a target twice | n/a | 713 Attributes_5 2 -> 0; CoR stacking is server side | n/a | Sim has one CoR aura per target (core.CurseOfRecklessnessAura). |
| WL-70425-06 | P6 (Classes) | Wands need 0.5 s standing still | n/a | Shoot 5019 Attributes_4 16777216 -> 0 | n/a | No warlock preset or APL uses a wand (no wand code in sim/warlock). |
| WL-70425-07 | P6 (Undead), FC | Touch of the Grave: 30% for casters, 25 s ICD, no Shadow scaling | yes (destruction_fire, Undead) | 1260201 SpellAuraOptions ProcChance 10 -> 30, ProcCategoryRecovery 1000 -> 25000 | landed (PR #50, racial scope) | sim/core/racials_forever.go:319-325 `procChance = 0.30` for ClassWarlock; ICD `time.Second * 25` (racials_forever.go:349-350). |
| WL-70425-08 | P6 (Gnome), FC | Eureka!: 8 s buff, no charges | yes (Gnome presets) | 1259821 DurationIndex 8 -> 31 (8 s), SpellAuraOptions 243705 (3 charges) removed | landed (PR #50) for duration; spell coverage wrong, see U-02 | sim/core/racials_forever.go:239 `Duration: time.Second * 8`. |

##### Build 70291 (P5, 2026-10-08)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WL-70291-01 | P5 | Current pet is stunned while a new pet is summoned | n/a | server side | n/a | Presets summon before the pull; Demonic Sacrifice before pull is an option, not a resummon. |
| WL-70291-02 | P5, FC | Shadow Bolt ranks 1-4 base points / points per level smoothed | minor (L60 casts rank 10/11) | SpellEffect: 686 ppl 0.3 -> 0.2 (bp 13); 695 bp 25 -> 23, ppl 0.6 -> 0.3; 705 bp 41 -> 40, ppl 0.7 -> 0.6; 1088 bp 56 -> 58, ppl 0.9 -> 0.8 | landed | sim/warlock/spell_data.go:13-16: `{686 Base: 13, PerLevel: 0.2}`, `{695 Base: 23, PerLevel: 0.3}`, `{705 Base: 40, PerLevel: 0.6}`, `{1088 Base: 58, PerLevel: 0.8}`. All 82 rank rows of spell_data.go (incl. Imp Firebolt, Lash of Pain) match 70425 base, per-level, variance, coefficient, cost and cast time. |
| WL-70291-03 | P5, FC | Demonic Brand: pet hits are high threat only when they deal Shadow | n/a (threat) | 1293695 description only | n/a | Threat. Damage part at talents.go:733-786 unchanged. |

##### Build 70170 (P4, 2026-10-01)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WL-70170-01 | P4, FC | Drain Soul stops when another spell is cast | minor (no L60 APL uses Drain Soul as filler) | SpellInterrupts ChannelInterruptFlags_0 97288 -> 31756 on 1120, 8288, 8289, 11675 | landed | sim/warlock/drain_soul.go:10-11: any cast ends a channel in the sim. |
| WL-70170-02 | P4 | Incubus model | n/a | - | n/a | - |
| WL-70170-03 | P4, FC | Hellfire can crit | n/a (not in any preset) | Hellfire Effect 5857/11681/11682 Attributes_2 541065216 -> 4194304 (drops 0x20000000 cannot-crit); Attributes_8 0 -> 512 | not simulated on purpose | No Hellfire in sim/warlock (self-damage AoE). |
| WL-70170-04 | P4, FC | Soul Harvesting renamed Soul Harvest; really gives its 50/100% regen | n/a (needs kills) | 437032 name; 1242853 e0 aura 379 -> 110, misc 4 -> 0 | n/a | Not modelled (boss fight has no kills). |
| WL-70170-05 | P4 | Pet Aggressive mode back | n/a | - | n/a | - |

##### Build 70009 (P1, 2026-09-24)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WL-70009-01 | P1 | Life Tap tooltip: health converted = mana gained, scales with Spirit only | yes (mana) | Spell 1454-11689 Description_lang only: `${($m1+$SPI*1)*(1+$18182m1/100)}` for both health and mana; 11689 bp 420, ppl 1 | landed | sim/warlock/lifetap.go:8-50: `health := (baseAmount + Spirit) * impLifeTap`, mana = health (x lifeTapBonus from items), no spell power. |
| WL-70009-02 | P1 | Voidwalker Sacrifice scales with 10% of spell healing | n/a | server side | n/a | Shield. |
| WL-70009-03 | P1 | Cooldown manager, pet flyouts, gamepad pet bar; known issue: banes/curses tracking on retarget | n/a | - | n/a | UI. |

##### Deep dive (DD, 2026-10-07, design at 70170)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WL-DD-01 | DD | Curse of Agony / Doom are Banes; one bane per warlock per target; curse effects also hit banes | yes | Bane tooltips "Only one Bane per Warlock"; Amplify Curse 18288 e0 mask 33792 covers Bane of Agony (+50%) | landed | sim/warlock/curses.go:10-30 activateBane / activateCurse; Amplify Curse x1.5 on Bane of Agony curses.go:72-75; TestWarlockBanesAndCurses. |
| WL-DD-02 | DD | Curse of the Elements covers all magic damage | yes (duty-curse presets, raid debuff) | 1311680 e0 aura 22 -75 misc 126, e1 aura 87 +10% misc 126 (all six magic schools) | landed | sim/core/debuffs.go:656-672, foreverCurseOfElementsResistance/Damage `{0,30,45,60,75}` / `{0,.04,.06,.08,.10}` (debuffs.go:285-286) match all 4 ranks. |
| WL-DD-03 | DD | Curse of Shadows removed | yes | no warlock Curse of Shadow in SkillLineAbility | landed | sim/warlock has no Curse of Shadow (curses.go:14). |
| WL-DD-04 | DD | Firestone / Spellstone are weapon imbues stacking with oils | yes | 23483 +21 Fire (aura 13) and 2% crit (aura 57); 1237165 +21 Shadow and 2% haste (aura 65); enchants 8059-8061 duration 7200 | landed | sim/warlock/talents.go:82-117 (`{56, 21, 2}`, `{60, 21, 2}`), added on top of consumes' oil. |
| WL-DD-05 | DD | Life Tap scales with Spirit | yes | see WL-70009-01 | landed | lifetap.go:44. |
| WL-DD-06 | DD | Drain Soul progressive Soul Shard chance | n/a | server side; 70425 adds a helper spell Drain Soul 1326042 (dummy aura) | not simulated on purpose | Shards not counted (notes.md:175). |
| WL-DD-07 | DD | Pets scale from the summoner's stats | yes | Warlock Pet Scaling 416189 all base points 0: server side | landed (amounts assumed) | sim/warlock/pet.go:216 makeStatInheritance (SoD coefficients), IgnoresPlayerBuffs; open item notes.md:166. |
| WL-DD-08 | DD | Improved Life Tap to row 1 | minor | 18182 | landed | lifetap.go:22 10% a point. |
| WL-DD-09 | DD | Suppression: up to 5% hit, -20% threat | yes | 18174 e0 aura 55 5, e1 aura 54 5, e2 aura 10 -20 | landed | talents.go:123-133 (1% spell and melee hit, 4% threat a point). |
| WL-DD-10 | DD | Improved Corruption also +10% Corruption damage | yes | 17810 e0 aura 107 -2000 ms cast, e1 aura 108 10 | landed | talents.go:135-145 `0.02` a point periodic; corruption.go:33 0.4 s a point. |
| WL-DD-11 | DD | Malediction: new, up to 5% periodic | yes | superseded by 70425 (WL-70425-01) | in flight (worker PR) | Removal pending, see WL-70425-01. |
| WL-DD-12 | DD | Improved Drain Soul renamed Soul Harvest | n/a | - | n/a | - |
| WL-DD-13 | DD | Improved Drains up to 20% (Drain Life, Drain Soul, Wrack) | yes | 403511 e0 aura 108 20, 3 ranks | landed | talents.go:159-169 `{0, .07, .13, .20}`. |
| WL-DD-14 | DD | Improved Bane of Agony: 2 points, more damage | yes | 18827 e0 aura 108 bp 2, 2 ranks (rank values from curve, not in this download) | landed | talents.go:171-181 5% a point (70338 curve audit). |
| WL-DD-15 | DD | Fel Concentration 3 points, covers Wrack | n/a (pushback) | 17783 e0 70 | n/a | - |
| WL-DD-16 | DD | Pandemic: up to +100% crit bonus on Affliction DoTs and drains | yes | 427712 e0 misc 15 bp 100, mask Corruption, Banes, Drain Life/Soul, Siphon Life, Wrack | landed | talents.go:183-194 `CritDamageBonus += points/3`. |
| WL-DD-17 | DD | Nightfall also from Drain Soul and Wrack | yes | 18094 e0 bp 4, 2 ranks | landed | talents.go:209-249: 2% a point on Corruption, Drain Life, Drain Soul, Wrack ticks. |
| WL-DD-18 | DD | Curse of Exhaustion to row 4, 30% snare | n/a | 18223 e0 aura 33 -30 | n/a | - |
| WL-DD-19 | DD | Malevolence up to 5% Shadow crit | yes | 1310949 e0 aura 107 misc 7 bp 5, Shadow spell mask (incl. Lash of Pain) | landed | talents.go:196-207 1% a point; Lash of Pain in succubus.go. |
| WL-DD-20 | DD | Soul Siphon up to 12% per other Affliction effect, max 36% | yes | 17804 e0 aura 4 bp 12 (dummy: count is server side) | landed (count assumed) | talents.go:281-311: 4% a point per effect, max 3; open item notes.md:163. |
| WL-DD-21 | DD | Shadow Mastery damage bonus reduced | yes | superseded by 70425 (WL-70425-02) | in flight (worker PR) | See WL-70425-02. |
| WL-DD-22 | DD | Wrack: Shadow damage a second, +10% from your other Shadow DoTs for 6 s | yes | 1316697: 6 s, 200 mana, e0 bp 36 per 1 s, coef 0.143; e1 aura 271 bp 10 mask 1026 (Corruption, Bane of Agony) | landed | sim/warlock/wrack.go:9-70 (36, 0.143, 6 ticks, 200 mana); talents.go:267-279 x1.1 on Corruption and Bane of Agony only, which is exactly the client mask. |
| WL-DD-23 | DD | Affliction talents removed (Imp. CoW, Imp. Drain Life/Mana, Imp. CoEx, Grim Reach, Dark Pact) | n/a | Grim Reach came back at 70425 | n/a | Not in proto. |
| WL-DD-24 | DD | Improved Imp no longer improves Blood Pact | minor | 18694 e0/e1 30 | landed | sim/warlock/warlock.go:165-167; imp.go:118 Firebolt +10% a point. |
| WL-DD-25 | DD | Demonic Embrace no longer lowers Spirit | minor | 70425 removed 18697 e1 (aura 137 bp -1 misc 4 = Spirit) | landed | talents.go:317-326 Stamina only; notes.md:172 can now be closed (see section 4). |
| WL-DD-26 | DD | Unholy Power to row 1, smaller pet bonus | yes (demonology) | 18769 e0 aura 107 bp 10, 5 ranks | landed | talents.go:328-339 2% a point on all demon damage. |
| WL-DD-27 | DD | Improved Health Funnel, Demonic Aegis, Improved Voidwalker, Improved Felhunter | n/a | - | n/a | Healing / armor / utility. |
| WL-DD-28 | DD | Fel Vitality up to 15% demon health/mana and warlock mana | minor | 18731 e0 15, e1 aura 178 15 | landed | talents.go:341-353 5% a point. |
| WL-DD-29 | DD | Demonic Energies: pet heal 15% of spell damage; pet gets up to 100% of Life Tap mana | minor | 1225214 e0 15, e1 100, 2 ranks | landed (mana part) | lifetap.go:23-54 `{0, 0.5, 1}`; the heal does nothing for DPS. |
| WL-DD-30 | DD | Fel Domination 16-point gold medal; Master Summoner row 3 | n/a | - | n/a | Tree layout. |
| WL-DD-31 | DD | Demonic Sacrifice redesign: 2 h, Imp Shadow, Succubus Fire, VW mana, FH health | yes (demonology) | 18789 aura 79 15% misc 32 (Shadow); 18791 15% misc 4 (Fire); 18792 aura 21 2% / 4 s; 18790 aura 20 3% / 4 s; all 7200000 ms | landed | talents.go:524-653. |
| WL-DD-32 | DD | Decimation: Soul Fire cd -90%, below 35% SB/Searing Pain: cast -40%, no shard, more damage | yes (demonology 2/2) | 440870 e0 40, e1 -90, e2 35, e3 0 (damage server side) | landed | talents.go:689-731 (20% cast a point, 3% damage a point), soul_fire.go:37 (45% cd a point); open item notes.md:164. |
| WL-DD-33 | DD | Demonic Brand | minor (no preset takes it) | 1293695 e1 bp 3 | landed | talents.go:733-796; open item notes.md:165. |
| WL-DD-34 | DD | Soul Link 21-point gold medal | yes (demonology) | 25228: 3% damage, 30% transfer | landed | talents.go:468-522. |
| WL-DD-35 | DD | Demonic Knowledge: spell damage up to 33% of level, warlock and demon | yes (demonology 3/3) | 412732 e0 100, 3 ranks | landed | talents.go:655-687 `{0, .33, .67, 1} * level`. |
| WL-DD-36 | DD | Master Demonologist redesign (Imp Fire, VW Physical taken, Succ Shadow, FH Magic taken) | yes (demonology 5/5) | 23785 e0-e3 bp 10, 5 ranks | landed | talents.go:385-466 2% a point. |
| WL-DD-37 | DD | Demonic Pact: sacrifice kept when summoning another demon | yes (demonology) | 425464 | landed | talents.go:603-612, 614-633; TestWarlockDemonicPact. |
| WL-DD-38 | DD | Demonology talents removed (Imp. Healthstone, Fel Intellect, Fel Stamina, Imp. Subjugate, Imp. Firestone/Spellstone) | n/a | - | n/a | Not in proto. |
| WL-DD-39 | DD | Destructive Reach row 1, all damaging spells | n/a (range) | superseded by 70425 | n/a | - |
| WL-DD-40 | DD | Improved Shadow Bolt: own Shadow damage only, no charges | yes | 17794 aura 270 bp 4 misc 32, 12 s | landed | talents.go:802-834 (4% a point, 12 s, all warlock Shadow hits and ticks). |
| WL-DD-41 | DD | Bane row 1, also Incinerate | yes | 17788 e0 -500 ms (Shadow Bolt, Immolate, Incinerate), e1 -2000 ms (Soul Fire, Hellfire) | landed | talents.go:836-850 (0.1 s / 0.4 s a point). |
| WL-DD-42 | DD | Cataclysm row 2, up to 10% Destruction cost | yes | 17778 e0 misc 14 bp -10 | landed | talents.go:852-864 `{0,3,6,10}`. |
| WL-DD-43 | DD | Molten Skin | n/a | - | n/a | - |
| WL-DD-44 | DD | Aftermath: up to +50% Immolate initial damage, Conflagrate daze | yes | 18119 e1 aura 108 50, e0 aura 42 100 | landed | talents.go:866-876 10% a point impact. |
| WL-DD-45 | DD | Ruin row 3, 5 points | yes | 17959 e0 misc 15 bp 100 | landed | talents.go:878-888 20% a point crit bonus. |
| WL-DD-46 | DD | Agonizing Flames: 3 points, Destruction damage plus Searing Pain crit | yes | 17927 e0 crit 10 (Searing Pain), e1 damage 10, e2 periodic 10 (Immolate, Rain of Fire) | landed | talents.go:890-907 `{0, .03, .07, .10}`. |
| WL-DD-47 | DD | Conflagrate 16-point; Intensity 3 points; Pyroclasm wording | n/a | - | n/a | Layout / pushback / stun. |
| WL-DD-48 | DD | Fire and Brimstone: Conflagrate crit | yes | 412751 e1 aura 107 25 | landed | talents.go:909-919 `{0, 8, 17, 25}`. |
| WL-DD-49 | DD | Bane of Havoc: 15% of damage to others copied to the target | n/a (single target presets) | 1225228 e0 15, 300 s | not simulated on purpose | No preset has a second target to feed it. |
| WL-DD-50 | DD | Shadow and Flame: Conflag -> Shadow +, Shadowburn -> Fire +, 20 s; keep Immolate; shard refund | yes (destruction_fire) | 426316 e0 20, e1 100, e2/e3 10, 5 ranks | landed | talents.go:921-960 (2% a point, 20 s), conflagrate.go:49 20% a point to keep Immolate. |
| WL-DD-51 | DD | Incinerate: 31-point, +25% with Immolate, 2.5 s cast | yes | 1293813 bp 217, ppl 1.4, var 0.15, coef 0.714, 325 mana, cast 2500; e1 bp 25 | landed | sim/warlock/incinerate.go:9-44 (x1.25 with Immolate), spell_data.go:136. |
| WL-DD-52 | DD | Destruction talents removed (Imp. Firebolt, Imp. Lash of Pain, Devastation, Imp. Immolate, Emberstorm) | n/a | - | n/a | Not in proto. |

#### 2. Unannounced client changes (DB2 reverse check)

Client-only builds: 70291 -> 70334 -> 70338 -> 70374 hold no warlock change (the 70291-70334 diff is one item spell, Summon Azurky; the other two are empty). 70205 / 70235 / 70245 / 70058 / 70124 were not downloaded, but the 70009 -> 70170 and 70170 -> 70291 diffs span them. The talent tree (TraitTree 1116) is identical in every build except: Soul Harvesting -> Soul Harvest (70170) and the 70425 node changes listed below. A spell-by-spell check of sim/warlock/spell_data.go against 70425 finds no damage, coefficient, cost or cast time drift.

| ID | Build | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| WL-U-01 | 70170 | New warlock-class Revelation aura (enchant proc) | minor (no warlock preset uses the Revelation enchant) | 1323392 added, SpellClassSet 5, +100% crit, 1 charge, mask covers Shadow Bolt, Searing Pain, Soul Fire, Shadowburn, Conflagrate, Incinerate, Death Coil, RoF/Hellfire and also the channels Drain Life, Drain Soul, Wrack | ambiguous (needs verify) | sim/common/enchant_effects.go:282-330 `isDirectCast` excludes channels; the client mask includes them. See section 4. |
| WL-U-02 | 70170, 70425 | Eureka! warlock mask: non-periodic only | yes (Gnome affliction, demonology, destruction and duty-curse twins) | 1259821 at 70425: e0 cost -10% and e1 damage +10% (misc 0) on Shadow Bolt, Immolate (direct), Searing Pain, Soul Fire, Shadowburn, Conflagrate, Death Coil, RoF, Hellfire; e2 periodic +10% (misc 22) on Drain Life, Drain Soul, Wrack only. Corruption, Banes, Siphon Life, Incinerate and Immolate's ticks are out. | wrong (cross-scope: racials) | sim/core/racials_forever.go:220-232 `affects` takes any APL spell with a cost and DamageMultiplier != 0, so Corruption, Banes, Siphon Life, Immolate ticks and Incinerate get +10% damage and -10% cost. Same item as G1 in orchestration/forever_class_changes.md:23, still open on master. |
| WL-U-03 | 70425 | Curse of Recklessness loses its attack power dummy; armor moves to aura 674 | yes (raid debuff) | 704/7658/7659/11717 e0 aura 4 (20/45/65/90) -> aura 92 bp 0; e2 removed; e1 aura 22 -> 674 at 70291, still -505 at rank 4 | landed | sim/core/debuffs.go:859-865 Forever CoR = 505 armor, no AP; TestWarlockCurseOfRecklessnessLands. |
| WL-U-04 | 70425 | Demonic Embrace's Spirit effect deleted | minor | 18697 SpellEffect 696833 (e1 aura 137 bp -1 misc 4) removed | landed | Sim already takes no Spirit (talents.go:317-326). The worker's branch did not note it; notes.md:172 stays open (missed by worker, docs only). |
| WL-U-05 | 70425 | Siphon Life ticks can crit (SPELL_ATTR8 periodic can crit) | yes (affliction if talented) | 18265/18879/18880/18881 Attributes_8 0 -> 512 | landed | Sim already crits it (forever-bugs #303), TestWarlockSiphonLifeCrits (mechanics_test.go:110). |
| WL-U-06 | 70425 | Amplify Curse: Curse of Exhaustion part removed / now 50% | n/a | 18288 e1 (aura 107 -20 misc 3, mask CoEx) removed; e0 +50% on Bane of Agony unchanged | n/a | Bane of Agony x1.5 unchanged (curses.go:72-75); no preset takes it. |
| WL-U-07 | 70425 | Spell family bit mask_1 256 moved from Unstable Affliction / Invocation to Hellfire Effect | n/a | 427717, 1219436, 1242971/2, 454197 mask_1 256 -> 0; 5857/11681/11682 0 -> 256; Malediction e1 2048 -> 256 | n/a | Puts Hellfire under the Destruction talent masks; Hellfire and UA are not simulated. |
| WL-U-08 | 70425 | Demon Skin / Demon Armor armor via aura 674 | n/a | 687/696/20798, 706/1086/11733-11735 aura 22 -> 674, same values | n/a | armors.go armor values unaffected. |
| WL-U-09 | 70425 | Portal of Summoning 100% base mana, 1 h cd, 5 reagents; new Ritual of Summoning 1326140, Healthstone 1326057 | n/a | 437169 RecoveryTime 120000 -> 3600000, PowerCostPct 12 -> 100 | n/a | - |
| WL-U-10 | 70425 | ATTR6 0x200000 on Fear, Howl, Banish, Death Coil, Seduction, Spell Lock, Subjugate, Devour, Lash of Pain (NPC), UA silence | n/a | SpellMisc Attributes_6 0 -> 2097152 | n/a | Client UI target effect flag; no damage field changed (Death Coil damage identical). |
| WL-U-11 | 70291 | PvP / helper copies (Bane of Havoc, Banish, Death Coil, Fear, Subjugate 1324xxx; Drain Mana 1324196/200/229; Fear 1323802) | n/a | added rows, no damage | n/a | - |
| WL-U-12 | 70291 | Destructive Reach gained a second effect point (radius), removed again at 70425 | n/a | TraitDefinitionEffectPoints 25188 added (e1 curve 125544) | n/a | - |
| WL-U-13 | 70009 | Eureka! cost -50% -> -10% (racial) | yes (Gnome) | 1259821 e0 bp -50 -> -10 | landed (racial) | racials_forever.go:222 `costReduction := int32(10)`. |
| WL-U-14 | 70009 | Coefficient bookkeeping (Improved Rain of Fire/Hellfire 24430, Searing Pain pushback 23046, Reduced Cleave Cooldown) EffectBonusCoefficient 1 -> 0; Health Funnel / Consume Shadows / Anguish attribute flags; Summon Voidwalker 1322007 added | n/a | as listed | n/a | No damage effect. |
| WL-U-15 | 70170 | Drain Soul channel interrupt flags; Hellfire crit flag; Soul Harvest aura | covered by WL-70170-01/03/04 | - | - | - |

#### 3. Proposed fixes

1. **Malediction removal and Shadow Mastery 2% (WL-70425-01, -02).** Already done on the worker branch: sim/warlock/talents.go drop `applyMalediction` (lines 147-157 and the call at line 49), `applyShadowMastery` 0.01 -> 0.02 (line 258), proto field 4 `malediction` -> `grim_reach`, talent strings re-searched, TestWarlockShadowMastery added. Impact from the branch's results files (FullBuffs, Consumes, long single target): affliction 1238.5 -> 1274.1 (+2.9%), destruction 1246.6 -> 1238.3 (-0.7%), destruction_fire 1192.7 -> 1182.4 (-0.9%), demonology unchanged. Presets: warlock__affliction(-duty-curse), warlock__destruction(-duty-curse), warlock__destruction_fire(-duty-curse); talentsString changes in all six req files, demonology untouched. Gap: the branch's *.meta.json `published`/`verify` dps are still the master numbers (982.3, 990.0, 973.6) and need the report regeneration after merge.
2. **Eureka! warlock coverage (WL-U-02), cross-scope, for the racials owner.** In sim/core/racials_forever.go `registerEureka`: on gain apply the 10% to `ImpactDamageMultiplierAdditive`-style direct damage only, add it to periodic damage only for channelled spells (SpellFlagChanneled), and skip the cost cut and damage for spells outside the class mask (for the warlock: Corruption, Bane of Agony, Bane of Doom, Siphon Life, Incinerate; Immolate keeps the direct part only). A per-class exclusion list of spell codes is the simplest form. Test: a Gnome warlock with Eureka active, Corruption tick and Incinerate unchanged, Shadow Bolt +10%, Drain Life tick +10%. Impact: small, -0.2 to -0.8% on the three Gnome builds (affliction most, through Corruption and Bane of Agony applied in the window); the prior 70338 estimate was -0.3 to -0.8 with 3 charges. Presets: warlock__affliction, warlock__demonology, warlock__destruction and their duty-curse twins.
3. **notes.md:172 (Demonic Embrace Spirit), docs only.** Close it: client 70425 deleted the Spirit effect, so the sim's "no Spirit" is now the client. No DPS change.

#### 4. Need to Verify candidates

* Shadow Mastery's 70425 value. The 2% a point rests on the dev notes and the CurvePoint rows the warlock worker read; the talent's own SpellEffect points (2 on both effects) never changed between 69893 and 70425, and the deep dive of 2026-10-07 said its "damage bonus reduced". The sim (on the worker branch) takes 2 / 4 / 6 / 8 / 10%. Worth 3% on Affliction. A 70425 client tooltip at 5/5, or Shadow Bolt damage with and without the talent at a known spell power, settles it.
* Eureka! for warlocks. Client 70425 limits Eureka's cost and damage to the direct spells plus the drains' and Wrack's ticks; Corruption, the banes, Siphon Life and Incinerate are outside its mask, and Immolate gets it on its hit only. The sim gives every damaging warlock spell both halves. The masks could be stale if the server applies its own list. A Gnome warlock's Corruption tick and Incinerate inside and outside the 8 sec window (combat log) settle it; worth up to about 0.8% on the Gnome builds.
* Revelation for warlocks. Since 70170 the client has a warlock-only Revelation aura (1323392) whose mask includes Drain Life, Drain Soul and Wrack, while the sim's generic Revelation leaves channels out because their ticks roll crit. Only matters if a warlock preset takes the Revelation weapon enchant (none does today). A warlock with the enchant channelling Drain Life after a proc settles whether a channel uses it.
* Malediction versus Grim Reach. The deep dive (2026-10-07) lists Malediction as a new talent and Grim Reach as removed; the 70425 notes and client reverse both. The sim follows 70425 (branch). Nothing to test unless Blizzard flips it back; flag for the next build diff.

#### 5. Count

Rows by status (sections 1 and 2, 86 rows): landed 47 (6 of them qualified: racial PRs, assumed server-side amounts, Eureka duration only), in flight (worker PR) 4, wrong 1 (cross-scope racial), ambiguous (needs verify) 1, not simulated on purpose 3, n/a 29, covered by other rows 1; missing 0, partial 0.

### Druid: game-update ledger (builds 70009 to 70425)

Scope: sim/druid, the druid parts of sim/core (buffs, consumes), presets tools/dps_report/req/druid__*.json (balance, balance-duty-ff, balance-wrath, balance-wrath-duty-ff, feral-cat, feral-cat-duty-ff; all Night Elf).
Checkout: origin/master 6a87c9c (druid 70425 port #52 merged; shared port #50 merged). Every "landed" below was read in code and compared with the 70425 client CSVs (S/db2/*.70425.csv), not taken from #52's notes.

Links: post 1 = https://us.forums.blizzard.com/en/wow/t/wow-forever-beta-development-notes-updated-october-15/2360696/1 (2026-09-24, build 70009), post 4 = .../2360696/4 (2026-10-01, 70170), post 5 = .../2360696/5 (2026-10-08, 70291), post 6 = .../2360696/6 (2026-10-15, 70425). Deep dive = https://worldofwarcraft.blizzard.com/en-us/news/24301515 (Hunter + Druid, 2026-09-30, describes 70170 with Shifting Power "new to next week's build"). fc = foreverchanges.pro/patch-notes.

Presets that use what: feral-cat and feral-cat-duty-ff (Shred 9830, Rip 9896, Ferocious Bite 22829, Shifting Power 1322605, Berserk 417141, Elune's Light; duty-ff adds Faerie Fire (Feral) 17392; Howling Idol 272427; talents 05002-05200332021031212051-05501, i.e. Shifting Power, Imp. Shifting Power 2, LotP, Natural Instinct 2, Blood Frenzy 1, Savage Fury 2, Sharpened Claws 2, Rend and Tear 5, Berserk, Furor 5, Naturalist 5, Natural Shapeshifter 1). Balance presets: Wrath 9912, Starfire 9876, Moonfire 9835, Insect Swarm 24977, Moonkin Form, Eclipse (balance only), Clearcasting; duty-ff adds caster Faerie Fire 9907.

#### 1. Announced changes

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| DRU-70425-01 | post 6, 2026-10-15, 70425 | Omen of Clarity can no longer trigger from Faerie Fire in Cat/Bear/Dire Bear (it costs nothing there) | yes (feral-cat-duty-ff) | server side (FF 770/778/9749/9907 only gain a threat effect, see -03) | landed | sim/druid/faerie_fire.go:49 FF is `ProcMaskSpellDamage` without `SpellFlagOmen`; talents.go:444 the spell branch needs `SpellFlagOmen`, so no FF ever procs OoC. Side note: caster/Moonkin FF (costs 115 mana) can proc OoC in game but never in the sim (balance-*-duty-ff), well under 1 DPS. |
| DRU-70425-02 | post 6, 70425 | Faerie Fire generates threat in Bear/Dire Bear | n/a | SpellEffect 1363296-1363299 ADDED (EffectIndex 3, Effect 63 threat) on 770/778/9749/9907 | n/a | tank only |
| DRU-70425-03 | post 6, 70425 | Only one Faerie Fire debuff on a target | n/a | server side | n/a | sim already uses one exclusive FF aura (core.FaerieFireAura / FaerieFireFeralAura) |
| DRU-70425-04 | post 6, 70425 | Shred rank 1 learned at 20 (was 22) | minor (leveling) | SpellLevels 98857 Base/SpellLevel 22 -> 20; SpellEffect 683154 (5221 eff 0, weapon damage flat) 24 -> 22 (the flat change is unannounced) | landed | sim/druid/shred.go:115-116 `ShredFlatDamage = {0, 22, ...}`, `ShredLevel = {0, 20, ...}` |
| DRU-70425-05 | post 6, 70425 | Rejuvenation high ranks no longer block other druids' Rejuvenation | n/a | SpellMisc Attributes_4 1048576 -> 0 on 11 Rejuvenation ranks (and on 9 Regrowth ranks, unannounced) | n/a | healing |
| DRU-70425-06 | post 6, 70425 | Heart of the Wild, Feral Swiftness, Thick Hide, Leader of the Pack apply again after login | n/a | Cat Form passive 5759 SpellMisc Attributes_0 168099856 -> 262160 (likely this fix) | n/a | sim always applies talents at build |
| DRU-70425-07 | post 6, 70425 | Several Insect Swarms no longer stack their hit reduction | n/a | SpellMisc Attributes_3 128 -> 0 on 5570/24974/24975/24976/24977; 24976 eff 1 EffectAttributes 0 -> 64. Damage unchanged (24977 tick 31, coef 0.158) | n/a | defensive part; sim's insect_swarm.go:16 ticks 31 still match |
| DRU-70425-08 | post 6, 70425 | Moonkin Aura only raises spell crit, raid wide | yes (all balance presets; every caster's raid buffs) | 24907 eff 701335: Effect 35 -> 65 (party -> raid), Aura 290 -> 57 (spell crit), bp 3 | landed | sim/core/buffs.go:291-296 `character.AddStat(stats.SpellCrit, 3*SpellCritRatingPerCritChance)`; druid.go:124-126 sets it in Moonkin Form (#50/#52) |
| DRU-70425-09 | post 6, 70425 | Leader of the Pack only raises melee and ranged crit, raid wide | yes (feral presets self-buff) | 24932 eff 702727: Effect 35 -> 65, Aura 290 -> 52, bp 3 | landed | buffs.go:298-302 `stats.MeleeCrit: 3 * CritRatingPerCritChance`; druid.go:128-130 |
| DRU-70425-10 | post 6, 70425 | Shred tooltip counts Savage Fury (visual only) | n/a | Spell 5221-9830 Description "plus 54..180" -> "$s1" | n/a | tooltip |
| DRU-70425-11 | post 6, 70425 | Furor keeps Cat Form Energy through Bear/Dire Bear Form; bears still don't regenerate Energy | n/a | 5487/9634 SpellEffect 1363096/1363092 ADDED (aura 110 power regen %, -100, misc 3 Energy); Furor 17056 text "outside Bear Form" | not simulated on purpose | no Bear Form in the sim (druid.go:186); cat-only Furor at forms.go:211-266 is unaffected |
| DRU-70425-12 | post 6, 70425 | Wands need 0.5 s standing still | n/a | server side | n/a | no druid preset uses a wand |
| DRU-70291-01 | post 5, 2026-10-08, 70291 | Rage from damage taken fixes; crit Rage for bears 100% (was 75%) | n/a | Bear/Dire Bear passive SpellAuraOptions 248463/248464 (added in 70170) REMOVED: server side | n/a | bear only |
| DRU-70291-02 | post 5, 70291 | Mana Potions usable in Bear/Dire Bear/Cat Form | minor | server side | landed | sim/core/character.go:750-751 `canUseItemsInForm` true under Forever; consumes.go:1332 mana potion gate |
| DRU-70291-03 | post 5, 70291 | Thorns back to Vanilla base, spell power ratio 6% (was 13.3%) | n/a (DPS players take no melee; raid hits are spell damage, core/raid_hits.go:38) | SpellEffect bp: 467 4->3, 782 9->6, 1075 11->9, 8914 13->12, 9756 16->15, 9910 22->18 | partial | sim/core/buffs.go:792-794 `foreverThornsDamage = {0, 3, 6, 9, 12, 15, 18}` matches; the 6% of the druid's spell power is left out (comment at :789) |
| DRU-70291-04 | post 5, 70291 | Cat Form no longer reduces max health | n/a | server side | n/a | |
| DRU-70291-05 | post 5, 70291 | Bear threat bonus 50% (was 30%); shapeshift-override effects; Demoralizing Roar threat | n/a | 21178 eff 700189 bp 30 -> 50; Demo Roar 99/1735/9490/9747/9898 Effect 63 added (8..42 + 0.8/lvl) | n/a | tank |
| DRU-70291-06 | post 5, 70291 | Wrath ranks 1-5 base and per-level retuned | minor (ranks 1-5 unused at 60) | 70425 values: r1 5176 15 +0.2 (14.58-17.02 at 5), r2 5177 21 +0.3, r3 5178 30 +0.5, r4 5179 37 +0.6, r5 5180 45 +0.7 (46.01-52.39 at 36); r6-8 untouched (54+0.9, 68+0.9, 91+1.0) | landed | sim/druid/wrath.go:17-18 every rank range matches the client at its cap |
| DRU-70291-07 | post 5, 70291 | Predatory Instincts renamed Natural Instinct; also spell healing +12/25% of Intellect | yes (feral, 2 points) | SpellName 1223242; new eff 1361117 aura 175 bp 25; crit dmg eff aura 108 misc 15 bp 20 | landed | talents.go:290-304 +10%/point crit damage bonus on Claw/Rake/Shred/Bite; healing part n/a |
| DRU-70291-08 | post 5, 70291 | Shifting Power can no longer be used at full Energy | minor (preset casts at <= 50 Energy) | server side | missing | sim/druid/shifting_power.go:22-45 no energy check (no ExtraCastCondition) |
| DRU-70291-09 | post 5, 70291 | Thick Hide implemented in code | n/a | 1306459 aura 665 -> 673, eff 1 bp -600 -> 180; 16929 eff 1 RealPointsPerLevel 0 -> 3 | n/a | talents.go:55-59 armor only, tank |
| DRU-70170-01 | post 4, 2026-10-01, 70170 | Bear/Dire Bear: +75% Rage on crits | n/a | 1178/9635 SpellAuraOptions proc added (later removed) | n/a | bear |
| DRU-70170-02 | post 4, 70170 | Swipe gains 3% of attack power | n/a | server side (no Swipe row changed; 9908 has no AP coefficient) | not simulated on purpose | sim/druid/swipe.go:16 registerSwipeBearSpell is not registered (druid.go:196); if bears return: add `0.03*spell.MeleeAttackPower(target)` per hit |
| DRU-70170-03 | post 4, 70170 | Tiger's Fury removed (also deep dive) | yes (feral) | TraitDefinition 134412 SpellID 417046 -> 1322605; T1 5P 1301247 and 1291059 retargeted from Tiger's Fury to Shifting Power | landed | sim/druid/tigers_fury.go:194 `if druid.Level < 24 || druid.Env.IsForever() { return }` |
| DRU-70170-04 | post 4, 70170 | All forms immune to Disarm | n/a | Aura 77 misc 3 (disarm immunity) added to 1178/3025/9635/5421/24905 passives | n/a | |
| DRU-70170-05 | post 4, 70170 | Faerie Fire no longer resets your swing timer | yes (feral-cat-duty-ff) | SpellInterrupts 123931/124015/125067/125881 (770/778/9749/9907, InterruptFlags 8 = auto attack) REMOVED; still absent in 70425 | wrong | sim/druid/faerie_fire.go:43 `flags \|= core.SpellFlagAPL \| core.SpellFlagResetAttackSwing` for every FF, the cat's 17392 included; core/cast.go:241-244 then calls `StopMeleeUntil(sim, now, false)`, restarting the 1.0 s cat swing on each FF |
| DRU-70170-06 | post 4, 70170 | Primal Bite threat doubled | n/a | text only | n/a | bear |
| DRU-70170-07 | post 4, 70170 | King of the Jungle removed | yes (talent string) | gone from tree 1089 (node list 70338 = 70425) | landed | proto/druid.proto:28-47 has no King of the Jungle; Cat Form cost forms.go:228-231 only Natural Shapeshifter (KotJ survives only in the uncompiled _enrage.go and commented bear code) |
| DRU-70170-08 | post 4, 70170 | Shifting Power (row 4 gold medal): Cat Form only, 16 s cooldown, costs 55% of base mana, gives 40 Energy, cost cut by shapeshift cost reductions | yes (feral) | 1322605: SpellPower PowerCostPct 55; SpellCooldowns Recovery 16000, StartRecovery 1000; Effect 30 bp 40 misc 3; CasterAuraSpell 768 | landed | sim/druid/shifting_power.go:27-44 (0.55 base mana, naturalShapeshifterFactor, GCD 1 s, 16 s, +40 Energy, Cat only) |
| DRU-70170-09 | post 4, 70170 | Improved Shifting Power: -4/-8 s | yes (feral, 2 points) | 1322670 aura 107 misc 11 (cooldown); rank values in curve 124948 (CurvePoint not downloaded; settled in 70338 audit) | landed | shifting_power.go:38 `16s - 4s*points` |
| DRU-70170-10 | fc 70170, unannounced | Shredding Attacks moved to row 3, Predatory Instincts to row 5 right (tree layout) | yes (talent strings) | tree 1089: node list identical in 70338, 70374 and 70425 | landed | proto/druid.proto:28-47 order matches tree 1089 rows; presets decode as intended (feral/talents_test.go) |
| DRU-70009-01 | post 1, 2026-09-24, 70009 | Wrath base damage ~+50% (deep dive: much cheaper, lower damage) | yes (balance, balance-wrath and duty) | 9912: bp 61 -> 91, per level 0.7 -> 1.0, var 0.112, coef 0.571, cost 120, cast 2000 ms (SpellCastTimes 5) | landed (rank 8 deliberately x1.085) | wrath.go:17-22 match the client (r8 91.57-102.43, coef 0.571, 120 mana, 2000 ms); wrath.go:24-27,45-48 `WrathRank8DamageScale = 1.085` (user ruling 2026-10-15, notes.md Druid) puts r8 at 99.35-111.14 |
| DRU-70009-02 | post 1, 70009 | Rejuvenation, Tranquility (and Wild Growth) can crit | n/a | SpellMisc Attributes_8 +512 on all ranks | n/a | healing |
| DRU-70009-03 | post 1, 70009 | Thorns scales with the caster's spell power when present | n/a | server side | partial | see DRU-70291-03 |
| DRU-70009-04 | post 1, 70009 | Mangle renamed Primal Bite, debuff removed | n/a | SpellName 407995/1238069/1238070/1238073 | n/a | bear |
| DRU-70009-05 | post 1, 70009 | Primal Fury renamed back to Blood Frenzy | yes (feral, 1 point) | SpellName 16958/16959/16953 Primal Fury -> Blood Frenzy; 16958 bp 100 (curve) | landed | talents.go:320-344 builder crit gives +1 CP at 50%/point |
| DRU-70009-06 | post 1, 70009 | Feral Charge (Cat) can't target allies; Wild Growth crit | n/a | 414924 eff 1 added | n/a | |
| DRU-DD-01 | deep dive, 2026-09-30, 70170 | Potions, Thistle Tea, Rage potions usable in forms | yes (feral) | server side | landed | core/character.go:750-751; sim/druid/thistle_tea.go:15-16 |
| DRU-DD-02 | deep dive, 70170 | Weapon enchants, stones, procs, Windfury Totem work in forms | yes (feral) | server side | landed | core/consumes.go:87-97 weaponEffectsWork; feral raid has Windfury Totem |
| DRU-DD-03 | deep dive, 70170 | Cat/Bear auto attack DPS = equipped weapon DPS at 1.0 / 2.5 s; weapon-damage abilities use it | yes (feral) | SpellShapeshiftForm 1 CombatRoundTime 1000, DamageVariance 0.4 (per #52's notes; table not in S/db2) | landed | forms.go:44-65 GetCatWeapon (weapon DPS x 1.0, +/-20%); notes.md Druid already lists the AP-per-swing question |
| DRU-DD-04 | deep dive, 70170 | Racials usable in forms (Shadowmeld, War Stomp, Walk on Air) | minor | server side | landed | racials are not DruidSpells (no FormMask); feral APL casts Elune's Light 1259799 in Cat Form |
| DRU-DD-05 | deep dive, 70170 | Claw damage increased | minor (no preset uses Claw) | 9850 weapon flat 115, 110% | landed | claw.go:11-15 `ClawFlatDamage ... 115`, `ClawWeaponMultiplier = 1.1` |
| DRU-DD-06 | deep dive, 70170 | Rake gains increased damage from attack power | minor (no preset uses Rake) | 9904 bp 61 and 34 a tick, no BonusCoefficientFromAP: server side | missing | rake.go:62-86 fixed 61 / 34 per tick, no AP term; coefficient unknown (NtV below) |
| DRU-DD-07 | deep dive, 70170 | Shred damage increased significantly | yes (feral) | 9830: eff 0 weapon flat 80, eff 1 155%, 60 Energy, level 54 | landed | shred.go:115-118, 140 (`60 - 6*ShreddingAttacks`; 16966 bp -6 a point) |
| DRU-DD-08 | deep dive, 70170 | Bear kit: Swipe AP, Lacerate (42), Demoralizing Roar, Enrage 10 Rage, Frenzied Regen 100%, Growl, Cower | n/a | | n/a | tank |
| DRU-DD-09 | deep dive, 70170 | Hurricane cooldown removed | minor | | landed | hurricane.go:10-48 no CD in the cast config |
| DRU-DD-10 | deep dive, 70170 | Faerie Fire castable in Bear/Cat/Dire Bear, free, 6 s cooldown | yes (duty-ff) | server side (no separate feral spell in the client) | landed | faerie_fire.go:28-41 (free, 1 s GCD, 6 s CD; action id 17392 is cosmetic) |
| DRU-DD-11 | deep dive, 70170 | Omen of Clarity baseline at 20; Clearcasting never used by Wrath | yes (all) | 16864 cooldown 10 s; chance server side | landed | talents.go:372-460 (Wrath excluded at :384 and :403; 2 PPM melee, spell guess `2*max(cast,1.5)/60`, NtV exists in notes.md) |
| DRU-DD-12 | deep dive, 70170 | Balance talents: Improved Wrath also cuts cost, Genesis, Moonglow (all damage spells), Improved Moonfire 2 pts, Nature's Majesty, Nature's Reach (hit), Nature's Splendor, Insect Swarm talent, Vengeance on all damage spells, Nature's Grace 3 s haste + GCD, Eclipse, Moonfury | yes (balance) | tree 1089 unchanged 70338 -> 70425; rank curves settled 2026-10-12 (notes.md) | landed | talents.go:26-42, 95-269; moonfire.go:45-50 and insect_swarm.go:31-37 for Nature's Splendor |
| DRU-DD-13 | deep dive, 70170 | Moonkin Form: OoC chance doubled; party crit to all, exclusive with LotP | yes (balance) | 24858 eff 5 OoC mods; crit part superseded by DRU-70425-08 | landed | talents.go:446-455 (x2 chance, half ICD); crit per 70425 |
| DRU-DD-14 | deep dive, 70170 | Feral talents: Ferocity (+Primal Bite), Heart of the Wild, Savage Fury (now Shred, lower), Sharpened Claws 2 pts, Shredding Attacks, Blood Frenzy, Predatory Strikes, Natural Reaction, Rend and Tear, Leader of the Pack | yes (feral) | 16998 (Savage Fury, two effects incl. periodic misc 22), 16942, 16966, 16972 (aura 99 AP, bp 50 = 50% of level/pt) | landed | talents.go:44-59, 271-318; forms.go:85-102. The deep dive calls the +20% crit damage talent "Predatory Strikes"; the client gives that to Natural Instinct (1223242) and makes Predatory Strikes attack power; sim follows the client |
| DRU-DD-15 | deep dive, 70170 | Berserk: 15 s, 3 min, +100% crit on Cat combo builders, Primal Bite x3 | yes (feral) | 417141 | landed | berserk.go:13-56 |
| DRU-DD-16 | deep dive, 70170 | Furor: Energy kept outside Cat Form (lost in Bear) | minor (presets never leave Cat Form) | 17056 m2 | landed | forms.go:211-266 (20%/pt of last Energy + 2/pt/s, cap 20/pt) |
| DRU-DD-17 | deep dive, 70170 | Naturalist: also +% all damage; Reflection; Living Spirit | yes (Naturalist 5 in all presets) | 17069 aura 79 | landed | talents.go:64-71 (`DamageDealtMultiplier *= 1 + 0.01*points`) |
| DRU-DD-18 | deep dive, 70170 | Resto and tank talents, Revive, Regrowth cost, Barkskin/Roots/Nature's Grasp in form, Overgrowth, Improved Entangling Roots | n/a | | n/a | |

#### 2. Unannounced client changes (reverse check)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| DRU-U-01 | DB2 70374->70425 (fc: not in Blizzard's notes) | Howling Idol's Shifting Power cooldown cut 1 s -> 0.5 s | yes (both feral presets wear 272427) | 1291059 eff 1310233 bp -1000 -> -500 (was -3000 before 70170) | landed | sim/druid/items.go:99-109 `spell.CD.Duration -= time.Millisecond * 500` (feral 717.4 -> 709.6) |
| DRU-U-02 | DB2 70170->70291 (fc lists it under Shifting Power) | Wolfshead Helm: +5 Energy (was 20) from activating Shifting Power (text since 70170; Tiger's Fury before). No Energy on Cat Form shift | minor (not in report presets; in ui/feral_druid/gear_sets p2-p6) | 17768 eff 694585 bp 20 -> 5; 1310990 Effect 30 energize bp 20 -> 5 | wrong | sim/druid/forms.go:219-222, 255 adds +20 Energy on every Cat Form shift under Forever too; shifting_power.go:42-44 adds nothing for the helm |
| DRU-U-03 | DB2 70009->70170 | "Tier 1 Feral 5P" Shifting Power cooldown -3 s -> -1 s | yes (feral synthetic set, 5 pieces) | 1301247 eff 1326284 bp -3000 -> -1000, retargeted from Tiger's Fury; unchanged in 70425 | landed | item_sets_pve.go:275-283 Grovekeeper Ferocity 5P `-= time.Second` (set to spell link not in the downloaded tables, see NtV) |
| DRU-U-04 | DB2 70009->70170 | New "Revelation" druid variant of Enchant Weapon - Revelation (+100% crit on next non-periodic spell) | n/a (standards.md 7: casters other than enhancement don't use Revelation) | 1323400 added (ProcCharges 1, class masks 4194341/2) | n/a | item scope; sim/common/enchant_effects.go:11-21 |
| DRU-U-05 | DB2 70374->70425 | Feral Swiftness movement speed 30% -> 15% | n/a | 24867 eff 701113 bp 30 -> 15 | n/a | movement |
| DRU-U-06 | DB2 70374->70425 | SpellMisc Attributes_6 +0x200000 on Rake, Rip, Pounce, Pounce Bleed, Lacerate, Bash, Feral Charge, Thrash, Hurricane (1289130), Entangling Roots, Hibernate, Skull Bash | n/a (no value change) | flag only, likely the disarm/form rules | n/a | |
| DRU-U-07 | DB2 70170->70291, 70374->70425 | Mark/Gift of the Wild, Barkskin, Tree Form: EffectAura 22 -> 674 (resistance variant); Tree of Life passive 439734 reworked; The Green Tower 17154 | n/a | as listed | n/a | resist/heal/tank |
| DRU-U-08 | DB2 70374->70425 | "Wrath" 18104 (item on-hit proc, ItemEffect 99146) bp 108 -> 126, range 2 -> 7 | n/a (no druid preset item) | | n/a | item scope |
| DRU-U-09 | DB2 70170->70291 | New Moonfire 1323651 (level 20, 95 mana, 40 direct + 10/3 s); no SkillLineAbility row | n/a | not player-learned | n/a | |
| DRU-U-10 | DB2 69893->70009, 70009->70291 | Starfall 20687 (SpellLevel 62 NPC area aura) bp 800 -> 1350; Starfall 1286114/1300361, Lunar/Umbral Fire and Venting (NPC, "players within yards") | n/a | | n/a | NPC spells |
| DRU-U-11 | DB2 69893->70009 | EffectBonusCoefficient 1 -> 0 on 6 aura spells (Improved Wrath 1248996, Improved Wrath/Starfire 24433/468260, Clever Shapeshifter, Enchanted Mushroom, Staff of the Glade) | n/a | coefficient on dummy/aura effects | n/a | |
| DRU-U-12 | DB2 70009->70170 | Cat Form passive 3025 adds aura 601 on class mask 1024 (Faerie Fire) and disarm immunity; Bear/Cat/Dire Bear descriptions | n/a | eff 1359783, 1359527 | n/a | FF in form already modelled (DRU-DD-10) |
| DRU-U-13 | DB2 70009->70170 | Reduced Maul and Swipe Cost 28855 class mask 64 -> 0 | n/a | Idol of Brutality, bear | n/a | |
| DRU-U-14 | fc 70170 | Trainers teach Primal Bite ranks 2-4 (36/48/60) | n/a | | n/a | bear |

Client-only builds: 70291->70334 changes one spell (Summon Azurky, a pet) and 70334->70338, 70338->70374 change 0 spells; nothing druid. 69977, 70058, 70124, 70205, 70235 and 70245 have no DB2 snapshot; the spanning diffs (69893->70009, 70009->70170, 70170->70291) hold only the rows above, so none is druid-relevant beyond what is listed. Caveat: CurvePoint is not downloaded, so a talent rank value changed only through its curve would not show; fc reports none for druid, and tree 1089 node/entry lists are identical in 70338, 70374 and 70425.

Spot checks against 70425 that came out equal to the sim (no row needed): Starfire r1-r7 (78.66-96.34 ... 349.96-412.04, coef 1.0, 3500 ms), Moonfire r10 (128.7-150.5, 60 a tick, 0.15/0.13), Insect Swarm r5 (31 a 2 s tick, 0.158), Claw r5, Rake r4 (61/34), Rip r6 (15 + 25.5/CP), Ferocious Bite r4 (70 + 128/CP), Shred r1-r5 (22/32/44/64/80, 155%), Moonkin/LotP bp 3.

#### 3. Proposed fixes

**F1 - Faerie Fire must not reset the swing (DRU-70170-05, wrong).**
- File: sim/druid/faerie_fire.go, registerFaerieFireSpell. Old: `flags |= core.SpellFlagAPL | core.SpellFlagResetAttackSwing` (line 43). New: `flags |= core.SpellFlagAPL` and add `core.SpellFlagResetAttackSwing` only when `!druid.Env.IsForever()` (Env is set by Initialize, where this runs). The explicit `StopMeleeUntil` at :71-73 only fires in Humanoid/Moonkin and can stay or get the same Forever guard (casters don't melee).
- Test: sim/druid/feral/forms_forever_test.go, `TestForeverFaerieFireKeepsSwing`: cat in Forever, advance 0.6 s into a swing, cast FF, assert the main-hand next swing time is unchanged.
- Impact: the duty preset refreshes FF about every 33 s (refresh at < 7 s of 40), about 6 casts in 180 s, each throwing away about 0.5 s of a 1.0 s swing: roughly 3 s of white damage, +3 to +4 DPS (~0.5%). The FF duty cost (709.6 - 704.6 = 5.0) would fall to about 1 to 2.
- Presets: druid__feral-cat-duty-ff only (balance duty casters don't swing; feral-cat has no FF).

**F2 - Wolfshead Helm under Forever (DRU-U-02, wrong, minor).**
- sim/druid/forms.go:255: give the +20 only when `!druid.Env.IsForever()`. sim/druid/shifting_power.go ApplyEffects: when the head is `WolfsheadHelm` and Forever, add 5 more Energy (spell 1310990, own metrics).
- Test: feral/forms_forever_test.go, Wolfshead on: Cat Form shift gives 0 extra; Shifting Power gives 45.
- Impact: none on the report presets (synthetic helm); fixes the UI's legacy gear sets.

**F3 - Shifting Power at full Energy (DRU-70291-08, missing, minor).**
- shifting_power.go:22: `ExtraCastCondition: func(...) bool { return druid.CurrentEnergy() < druid.MaxEnergy() }`.
- Test: feral talents_test, cast at 100 Energy fails, at 99 succeeds.
- Impact: 0 on presets (APL casts at <= 50 Energy).

**F4 - Rake attack power (DRU-DD-06, missing, minor).** No code until the coefficient is known (NtV N1). If added, Rake still won't enter the feral APL unless the knob search picks it.

No change proposed for Thorns' 6% spell power (no DPS effect) or Swipe's AP (no Bear Form in the sim).

#### 4. Need to Verify candidates

* N1 Rake's attack power (Forever). Blizzard's class deep dive says "Rake: Now gains increased damage from Attack Power", but client 70425 gives Rake 9904 no attack power coefficient (61 initial, 34 a tick over 9 sec, as in 70338), so the server adds it and the sim has none (sim/druid/rake.go). No preset uses Rake today; with the scaling it could join the feral rotation as a second builder or a Rend and Tear enabler. A Rake hit and its ticks at two known attack power values on the beta settle it.
* N2 Grovekeeper Ferocity 5 piece's spell. The sim takes 1 sec off Shifting Power, which is the "1.60.0 - Item - Tier 1 - Druid - Feral 5P Bonus - Shifting Power" spell 1301247 (-1000 ms since 70170, unchanged in 70425). Howling Idol's 1291059 dropped to -500 ms in 70425; if the set links to 1291059 instead, the 5 piece is now worth half (about 4 DPS on both feral presets). The client's ItemSetSpell row for set 2115 (not in our download) or the set's in-game tooltip settles it.
* N3 Omen of Clarity from a caster Faerie Fire. Build 70425 says Faerie Fire can't trigger Omen of Clarity in Cat or Bear Form because it costs nothing there, which implies the 115 mana caster Faerie Fire (and in Moonkin Form) still can. The sim never lets Faerie Fire proc it (faerie_fire.go has no SpellFlagOmen). Worth well under 1 DPS to the balance duty presets; a few dozen caster Faerie Fires watched for Clearcasting settle it.
* (Existing, unchanged) Wrath rank 8 x1.085 over the client's 91 + 1.0/level (WrathRank8DamageScale, user ruling 2026-10-15), Omen of Clarity's proc chance, Rip and Bite attack power, the cat weapon's AP-per-swing: already in notes.md, Druid.

#### 5. Count

Rows by status (69 rows: 55 announced or deep dive, 14 unannounced): landed 31, n/a 30, wrong 2 (Faerie Fire swing reset, Wolfshead Helm), missing 2 (Shifting Power full-Energy check, Rake attack power), partial 2 (Thorns spell power, twice), not simulated on purpose 2 (Swipe AP, Furor through Bear Form), in flight 0, ambiguous 0.

### Shared systems ledger: races, general mechanics, items, consumables, enchants

Scope: every race line 70009 to 70425, general mechanics (wands, downranking, raid-wide buffs, Salvation, scrolls,
Predatory Senses, Wizard Oil, Mystic Mushroom / Insight, PvP crit, generic Rage), the unannounced item changes
(the 70009 EffectBonusCoefficient wipe, ItemSparse for the preset gear) and SpellItemEnchantment.

Checked against **origin/master 4f5c60833** (PR #56 merged the full 70425 racial diff on top of #50, #53, #55), read with
`git show origin/master:<path>`; line numbers below are on that commit. No class worker branch (hunter, mage, paladin,
priest, shaman, warlock) touches `sim/core`, so nothing shared is in flight on them; the only in-flight row is the
damage-taken Rage formula (#55, being re-checked by a mechanics auditor), per the coordinator.

Scripts (all `python3 -I`, in `S/tools/`): `se.py <build> <spellIDs>` (SpellEffect / Misc / AuraOptions / Cooldowns /
EquippedItems / ClassOptions dump), `coef.py` (the 70009 coefficient wipe), `sie.py` (SpellItemEnchantment diff per
build), `items.py` and `itemchg.py` (ItemSparse diffs, preset gear and high-level items).

Links: post 1 = https://us.forums.blizzard.com/en/wow/t/2360696/1 (2026-09-24, 70009), post 4 = .../2360696/4
(2026-10-01, 70170), post 5 = .../2360696/5 (2026-10-08, 70291), post 6 = .../2360696/6 (2026-10-15, 70425).

#### 1. Announced changes

##### Build 70425 (post 6, 2026-10-15)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHR-70425-01 | post 6, 70425 | Tauren Endurance: 2% increased crit damage and healing (was 1% hit) | yes | SpellEffect 20550 idx1 aura 54→163, bp 1→2; idx2 aura 55→50, bp 1→2; idx0 aura 133 (health) 5 unchanged. Sim should give +5% health, crit damage ×1.02, no hit | landed | racials_forever.go:72-73 `MultiplyStat(Health,1.05)`, `critDamageBonus(1.02)` (applied at :96-104 to every attack table's CritMultiplier). What "2%" multiplies is already Need to Verify (notes.md:116) |
| SHR-70425-02 | post 6, 70425 | Plainsrunning +3% mounted speed and a minimum-distance rule; Cultivation 15 min; War Stomp 90 s | n/a | 1259918 new idx3 aura 130 bp 3; 1312643 CategoryRecoveryTime 3600000→900000; 20549 RecoveryTime 120000→90000 | n/a | movement / utility |
| SHR-70425-03 | post 6, 70425 | Dwarf Mace Specialization replaced with Earth Fury: 2% increased crit damage and healing | yes | 1259719 renamed; idx0 aura 290→163, bp 1→2; idx1 ADDED aura 50 bp 2. **SpellEquippedItems 255482 (class 2, subclass mask 48 = Mace, 2H Mace) is still on the spell** (Ruthless's and Savage's rows were removed) | landed | racials_forever.go:36 `critDamageBonus(1.02)` for every Dwarf, with any weapon. Values right; the kept mace row is a new Need to Verify (section 4) |
| SHR-70425-04 | post 6, 70425 | Find Treasure: extra chest items | n/a | server side | n/a | |
| SHR-70425-05 | post 6, 70425 | Skyborne Wind Blessed: haste 3% (was 1%) | yes | 1259710 idx0 aura 342 (melee+ranged haste) bp 1→3; idx1 aura 65 (cast speed) bp 1→3 | landed | racials_forever.go:85-87 Melee, Ranged and Cast speed multipliers `*= 1.03` |
| SHR-70425-06 | post 6, 70425 | Skysight / Read Leyline castable moving, Read Leyline 30 s min, Walk on Air 90 s, more Convergences | n/a | | n/a | regeneration / movement |
| SHR-70425-07 | post 6, 70425 | Human Sword Specialization replaced with Ruthless: 2% crit chance | yes | 20597 renamed; aura 290 bp 2 unchanged; SpellEquippedItems 255480 (swords, mask 384) REMOVED | landed | racials_forever.go:27-29 +2% melee and spell crit, no weapon test (hunter's ranged-crit removal deleted by #56). Ranged crit open in notes.md:115 |
| SHR-70425-08 | post 6, 70425 | Orc Axe Specialization replaced with Savage: 1% crit chance | yes | 20574 renamed; aura 290 bp 1; SpellEquippedItems 255488 (axes) REMOVED | landed | racials_forever.go:62-63 (#53) |
| SHR-70425-09 | post 6, 70425 | Blood Fury cooldown 3 min (was 2) | yes | SpellCooldowns 55612 RecoveryTime 120000→180000 | landed | racials_forever.go:187 `time.Minute * 3` |
| SHR-70425-10 | post 6, 70425 | Shatter Curse dispels instead of immunity, magic damage taken −5% (was −15%) | n/a | 1299026 idx0 bp −15→−5; idx2 aura 41→effect 38 | n/a | defensive |
| SHR-70425-11 | post 6, 70425 | Touch of the Dead (= Touch of the Grave): 30% casters, 20% Paladin/Rogue/Warrior, 25 s internal cooldown | yes | SpellAuraOptions 243733 (1260201) ProcChance 10→30, ProcCategoryRecovery 1000→25000; 243729 (1260189) 5→20, 1000→25000 | landed | racials_forever.go:342 `0.20`, :346 `0.30` (Mage, Priest, Warlock), :371 25 s ICD |
| SHR-70425-12 | post 6, 70425 | Touch of the Dead no longer scales with Shadow increases on the target or the caster | yes | drain 1260198 SpellMisc 813545: Attributes_3 0→0x20000000, Attributes_4 0→0x100, Attributes_6 0→0x28000000, Attributes_9 0→0x1000000 (the no-done/no-taken bonus flags). Sim should deal a flat 5% of max health | **missing** | racials_forever.go:351-358: the drain has `DamageMultiplier: 1` and flags `SpellFlagPassiveSpell \| SpellFlagNoOnCastComplete \| SpellFlagBinary`, no `SpellFlagIgnoreModifiers`, so Shadowform, Shadow Weaving, Curse of Shadow / Elements, Shadow Mastery and the like still raise it |
| SHR-70425-13 | post 6, 70425 | Touch of the Dead no longer weaker when a lower spell rank triggers it | yes | server side | landed | the sim never scaled it by rank (flat `MaxHealth() * 0.05`, :361) |
| SHR-70425-14 | post 6, 70425 | Will of the Forsaken not under Scatter Shot | n/a | 7744 SpellMisc attribute changes | n/a | |
| SHR-70425-15 | post 6, 70425 | Eureka!: 8 sec buff, no charges, same damage and cost effect | yes | all five Eureka spells (1259812 Rogue, 1259813 Warrior, 1259817 Mage, 1259821 Warlock, 1259823 Priest; SpellClassSet 8/4/3/5/6): DurationIndex 8 (15 s)→31 (8 s), SpellAuraOptions rows (ProcCharges 3) REMOVED; bp −10 cost / +10 damage unchanged | landed | racials_forever.go:260 8 s, :268-277 every affected spell ×1.1 and −10% cost while up, 2 min CD :288 |
| SHR-70425-16 | post 6, 70425 | Eureka! now also benefits Power Word: Shield; can't be used in stealth | n/a | 1259823 masks | n/a | healing / stealth |
| SHR-70425-17 | post 6, 70425 | Night Elf Quickness speed stacks; Troll Regeneration 20% health regen (was 10%) | n/a | 20582 idx1 aura 31→129; 20555 bp 10→20 (auras 88, 116) | n/a | movement / health regen |
| SHR-70425-18 | post 6, 70425 | Wands need you to stand still 0.5 sec before shooting | n/a | Shoot 5019 SpellMisc Attributes_4 0x1000000→0; Shoot 1316579 Attributes_6 0→0x200000; the delay itself is server side | not simulated on purpose | the sim has no wand Shoot spell (no `RangedWeaponTypeWand` handling outside tests); the presets' wands are stat sticks |
| SHR-70425-19 | post 6, 70425 | Moonkin Aura only spell crit, raid wide | yes | 24907 idx0 effect 35→65 (raid area), aura 290→57 | landed | buffs.go:294-296 +3% spell crit (#50). Caster reqs carry `moonkinAura` |
| SHR-70425-20 | post 6, 70425 | Leader of the Pack only melee and ranged crit, raid wide | yes | 24932 idx0 effect 35→65, aura 290→52 | landed | buffs.go:298-302 +3% MeleeCrit (also ranged in this sim) (#50). Melee and hunter reqs carry `leaderOfThePack` |
| SHR-70425-21 | post 6, 70425 | Trueshot Aura is raid wide | yes | 20906 idx0 effect 35→65 (also 19506, 20905, 1299346, 1299348); tooltip "party or raid members" | ambiguous (needs verify) | buffs.go:1699-1720 gives ranged AP only. Mechanic is fine; the raid composition is not: the BM, Lone Wolf and Survival reqs have `trueshotAura` false (tools/rotopt/mksettings/spec_hunter.go:54, :63) and notes.md:112 still says "a party aura under Forever". Needs a ruling on whether the level 60 raid has a Marksmanship hunter |
| SHR-70425-22 | post 6, 70425 | Battle Shout is raid wide | yes | 25289 (and ranks 5242…11551) SpellEffect ImplicitTarget 20→56 | landed | composition: every melee, hunter and caster req except warlocks and the warrior duty variants has `battleShout: Improved` (the duty variants cast their own); standards.md:135. Value buffs.go:1641 (139 AP at 60) |
| SHR-70425-23 | post 6, 70425 | Grace of Air, Strength of Earth, Windfury and Flametongue Totems are raid wide (Fire Resistance Totem too, unannounced) | yes | 25360 / 25362 effect 35→65; Windfury / Flametongue tooltips "party or raid" | landed (SoE, WF) | composition: rogue, warrior, feral, ret paladin and survival reqs carry `strengthOfEarthTotem` + `totemWeaponBuff: Windfury`; enhancement / elemental shamans drop their own (standards.md:140) |
| SHR-70425-24 | post 6, 70425 | Grace of Air Totem is no longer exclusive with Windfury and Flametongue Totem | yes | Grace of Air 25360 SpellMisc 332984 Attributes_11 0x400→0 (the exclusivity flag) | **missing** | buffs.go:2009: under Forever the Windfury Totem buff still registers `graceOfAirWindfuryCategory` at priority 2, which switches Grace of Air's Agility off (comment :1616-1621 "For now Forever doesn't let the two stack"). No melee req carries `graceOfAirTotem` either |
| SHR-70425-25 | post 6, 70425 | Flametongue and Windfury Totem only affect players | n/a | | landed | pets take no player buffs under Forever (buffs.go:510, standards 14) |
| SHR-70425-26 | post 6, 70425 | Blessing of Salvation → Salvation / Mass Salvation (not exclusive with blessings), Revoke | n/a | 1038, 1912, 25895 renamed / reworded | n/a | threat only; routed to the paladin worker (#54) |
| SHR-70425-27 | post 6, 70425 | Mystic Mushroom and the Insight enchant now increase Spirit | n/a | Insight 1299796 (from enchant 8216 → 1248758, 35% proc, 45 s ICD) aura 137 misc_1 0→16 (Spirit mask), bp 100 | n/a | neither is in a preset or the sim |
| SHR-70425-28 | post 6, 70425 | Scrolls incur a global cooldown | n/a | stat scrolls (12174, 12176, 12177, 12179) already had StartRecoveryTime 1500 in 70374: server side | n/a | presets use scrolls only before the pull (raid buffs; the mage staff imbue scroll is a class option) |
| SHR-70425-29 | post 6, 70425 | Draught of Predatory Senses is a potion, on the combat potion cooldown, shorter | n/a | 1250889 renamed Potion of Predatory Senses, DurationIndex 30→23, Attributes_3 changed | n/a | stealth detection |
| SHR-70425-30 | post 6, 70425 | Crafted item sell prices, Battleground Marks 30, Legacy perks, tutorial, UI | n/a | | n/a | |

##### Build 70291 (post 5, 2026-10-08)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHR-70291-01 | post 5, 70291 | Rage from taking damage: creature health instead of the player's, crit / crushing multipliers count, absorbs ignored, rebalanced for 20-40% armor | yes | server side | in flight (worker PR) | rage.go:81-110 `ForeverDamageTakenRage` (#55); being re-checked by the mechanics auditor (coordinator) |
| SHR-70291-02 | post 5, 70291 | Multiple weapon swaps within one GCD | n/a | | n/a | |
| SHR-70291-03 | post 5, 70291 | Critical strikes in PvP no longer have reduced efficacy | n/a | | n/a | PvP |
| SHR-70291-04 | post 5, 70291 | Attack Power armor kits no longer grant twice as much Ranged Attack Power | minor | SpellItemEnchantment 8483, 8486, 8488, 8491: Effect_2 (type 5, arg 39 = ranged AP) removed, the AP line stays | landed | db.json 1254965 Forceful Rugged Armor Kit = AP 10 + RAP 10 (once each, as AP from items gives both). No preset uses an AP kit |
| SHR-70291-05 | post 5, 70291 | Revelation enchant buff survives Ghost Wolf / Druid forms; Discolored Healing Potion no longer breaks CC; quest reward rarity and Desolace rewards | n/a | | n/a | no preset uses Revelation (sim-status priest note); quest gear is levelling |

##### Build 70170 (post 4, 2026-10-01; same build as the 10-02 warrior post)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHR-70170-01 | post 4, 70170 | Eureka! no longer benefits periodic effects at all; channeled spells do not count as periodic | yes | 70009→70170 rewrote every Eureka mask and added "non-periodic" to four tooltips (1259812/13/21/23). Decoded masks (70425): the cost (idx0, misc 14) and damage (idx1, misc 0) masks leave out the DoTs (Rogue lacks Rupture 0x100000 and Garrote 0x100; Warrior lacks Rend 0x20; Priest lacks SW:P 0x8000; Warlock lacks Corruption 0x2 and Curse of Agony 0x400), and idx2 (misc 22 = periodic damage) covers only channels (Warlock flags 0x8, 0x4000 = Drain Life, Drain Soul; Priest 0x800000 = Mind Flay). Mage idx2 mask is empty | **partial** | racials_forever.go:246-254 picks every APL spell with a cost and a damage multiplier, and :270 multiplies `spell.DamageMultiplier`, which `AttackerDamageMultiplier` (spell_result.go:635) applies to ticks too. So a Rupture, Rend, Corruption, SW:P, Immolate or Flame Shock cast in the 8 s costs 10% less and its ticks (snapshotted) deal 10% more. The race table text (build_races.py PART_INFO 'eureka') already says "non-periodic" |
| SHR-70170-02 | post 4, 70170 (UI) | Spells cast at ranks vastly below your level get less spell power and a lower chance to trigger class abilities and talents | yes | server side (no numbers; foreverchanges' 5% a level past learn level + 17) | partial | per class: priest.go:150 and shaman talents.go:176 `downrankCoefficient` (+ paladin exorcism.go); mage, warlock, druid presets cast top ranks. The lower proc chance (e.g. rank 1 Frostbolt → 0% Frostbite) is not modelled anywhere; no preset casts a rank that low (spec_assumptions.md:666) |
| SHR-70170-03 | post 4, 70170 | Players generate 75%, then 100%, more Rage from a critical basic attack | yes | server side | landed | warrior.go:179 `AddCritDealtRageMultiplier(2)`, applied in rage.go:173 (generic part); bear 100% is the druid ledger's |
| SHR-70170-04 | post 4, 70170 | Cannibalize not while immune to Physical; racials in the Cooldown Manager | n/a | | n/a | |

##### Build 70009 (post 1, 2026-09-24)

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHR-70009-01 | post 1, 70009 | Gnome Eureka!: 10% discount on Mana, Rage or Energy for every class | yes | 1259817 bp −50→−10, 1259823 −15→−10, 1259812 −20→−10, 1259821 −50→−10, 1259813 −40→−10 | landed | racials_forever.go:243 `costReduction := int32(10)` |
| SHR-70009-02 | post 1, 70009 | Touch of the Grave no longer breaks CC or triggers from spells without a damage component (SW:P on cast, not ticks) | yes | server side (1260198 tooltip only) | landed | racials_forever.go:395 rolls only when `spell.DamageMultiplier != 0`; ticks don't roll (CallbackOnCastComplete / NoOnCastComplete split) |
| SHR-70009-03 | post 1, 70009 | Contingency Plan can crit; Confounding Flash visual; Starshards pushback; Dark Sacrifice CC and tooltip | n/a here | priest racial spells | n/a | priest ledger (Starshards, Dark Sacrifice, Chastise are priest-only racials in the sim) |
| SHR-70009-04 | post 1, 70009 | Gnome / Human / Tauren cosmetic and Cultivation level rule | n/a | | n/a | |
| SHR-70009-05 | post 1, 70009 | Minor / Lesser / plain Wizard Oil back to 8 / 16 / 24 spell power; Brilliant unchanged | yes | 25110 bp 25→16 (auras 13 and 135), 25111 bp 30→24; 25113 Brilliant 36 + 1% crit unchanged | landed | consumes.go:125-145: 8 / 16 / 24 / 36 + 1% crit, plus the same healing under Forever. 21 presets use Wizard Oil |
| SHR-70009-06 | post 1, 70009 | Wands no longer gain spell damage | n/a | server side | not simulated on purpose | no wand attack in the sim |
| SHR-70009-07 | post 1, 70009 | Respec cost, guns and bows ammo, quest reward item levels | n/a | | n/a | |

The class deep dives (news 24301514, 24301515, 24302097, 24302099, 24310968) say nothing about racials, wands, raid
buffs or consumables beyond the class rows, apart from the pet-buff point already in standards 14.

#### 2. Unannounced client changes

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHR-U-01 | DB2 69893→70009 | **EffectBonusCoefficient 1 → 0 on item spells**: 1,043 SpellEffect rows in 875 spells (the diff's "835 spells whose only change"), plus 10 rows with a real coefficient set to 0 | minor | By effect type (coef.py): 1,029 of the 1,043 are effects where the coefficient does nothing: trigger spell (64), dummy (3), create item (24), stat / AP / resistance auras (aura 29, 99, 13, 22, 42…), summon, energize, teleport. So it is a data cleanup of a placeholder 1, not "item procs stop scaling". The 14 damage / heal / absorb rows that lost a 1: Shock 2606-2610, Drain Life 16375, 21170, 24585, 461683, Frost Nova 463448, Stratholme Holy Water 17291, Mark of the Dragon Lord 17252 (absorb), Herbal Soil Poultice 1246026, CHU's quest spell 4506; the 10 real coefficients zeroed: Shock 1213355 (0.15), Chain Lightning 463946 (0.10), Lightning Bolt 434842 (0.10), Fire Blast 461896 (0.15), Drain Life 446317 / 462221 / 468184, Infernal Lasso 447590, Mind Rot 606, Trail Snack 436432. Every proc the presets use is unchanged in all builds: Shadow Oil 350026 0.56, Frost Oil 350025 0.269, Dragonbreath Chili 15851 1 (65 base), Crusader 458112 → Holy Strength 20007 +100 Str; the stat enchants' aura spells (15396, 15401, 15403, 19984, 13823, 19988, 23793, 23794, 27836, 13365, 13372, 19992) are identical in 69893 and 70425 | wrong (minor, no preset) | sim/common/item_effects.go:2082 Shadowstrike proc 461683 `BonusCoefficient: 1.0` (client 0 since 70009); :1700 Masterwork Stormhammer 463946 has none (matches 0). consumes.go:258, :309, :564 match the client |
| SHR-U-02 | DB2 69893→70009 | The Human Spirit 20598 aura 137 gets EffectMiscValue_1 0→16 (the Spirit mask; the same fix as Insight in 70425) | yes | bp 5 | landed | racials_forever.go:27 Spirit ×1.05 |
| SHR-U-03 | DB2 69893→70009 | Frenzy potions: aura 13 (physical damage +40) → aura 99 + 124 melee and ranged AP, doubled (Major 40→80, Frenzy 14→28); Spellblasting potions raised (Major 40→47, Greater 20→23) | yes | 1251940, 1251936, 1251932, 1251924 | landed | consumes.go:1564 `meleeAndRangedAP(80)`, :1576 SpellDamage 47; 14 presets use Major Frenzy, 7 Major Spellblasting |
| SHR-U-04 | DB2 70009→70170 | Mageblood Potion 24363 (Mana Regeneration, 12 mp5) loses its potion category and 60 s CategoryRecoveryTime | minor | SpellCooldowns 55801 removed, SpellCategories Category 4→0 | landed | consumes.go:821 treats it as a mana-regen elixir slot, not a potion |
| SHR-U-05 | DB2 70374→70425 | Earth Fury keeps the Mace Specialization equipped-item row | yes | SpellEquippedItems 255482 (1259719, class 2, subclass mask 48) not removed, where 255480 (Ruthless) and 255488 (Savage) were | ambiguous (needs verify) | sim gives it with any weapon (racials_forever.go:36). Section 4 |
| SHR-U-06 | DB2 70374→70425 | Eureka! Rogue (1259812) cost mask gains EffectSpellClassMask_2 bit 2; Rogue Eureka gets Attributes_1 0x20, Attributes_2 0x4000 | minor | | n/a | the sim does not use the per-class masks (generic damaging-ability test) |
| SHR-U-07 | DB2 70374→70425 | Stoneform, Will to Survive, War Stomp Attributes_6 0x200000; Stoneform 7020 aura 22→674 | n/a | | n/a | defensive |
| SHR-U-08 | DB2 (all builds) | Troll Beast Slaying, Dwarf Big Game Hunter and Skyborne Elemental Insight are aura 168 (damage done, 5%), not crit damage | minor | aura 168 at 5% (70338, unchanged) | wrong (known) | racials_forever.go:114 `at.CritMultiplier *= multiplier` double-dips crit; notes.md:52 lists it; the fix lives only in the local dashboard patch 0003. Only the race table's Beast / Elemental target variants see it |
| SHR-U-09 | DB2 70374→70425 | 46 items of item level ≥55 change level: the Alterac Valley rewards 65→60 (Tomes 19308-19312, Don Julio's Band 19325, Don Rodrigo's Band 21563, Ironbark Staff, Sageclaw, Mindfang, The Lobotomizer…, stats re-budgeted e.g. 19325 StatPercentEditor 4074/5185/3704/5926 → 4000/5600/4000/5600), and 31 quest items at 56-60 | n/a | ItemSparse ItemLevel | n/a | the presets wear synthetic gear; Don Julio's Band 19325 is only named in comments (forever_synthetic_items_druid.go:317, _rogue.go:165, :319) |
| SHR-U-10 | ItemSparse, preset gear | The level 60 presets wear 245 distinct items; 241 are synthetic (990001…998115, not in the client). The 4 client items, Howling Idol 272427, Swarming Idol 272430, Burning Totem 272433, Libram of Law 272435, are identical in every build 69893…70425 | — | items.py | landed | nothing to port |
| SHR-U-11 | SpellItemEnchantment, all builds | The 23 enchants the presets use: every enchant row and its stat spell identical 69893…70425 (Sniper Scope 664, Lesser Arcanum of Voracity 1506/1508/1509, Bracer Superior Strength 1885, Boots Greater Agility 1887, Cloak Greater Resistance 1888, Chest Greater Stats 1891, Crusader 1900, Weapon / 2H Agility 2564 / 2646, the Forever gloves, necklace, bracer, cloak and Spellblasting enchants 7646, 7647, 7662, 7665, 7667, 8213, 8218, 904, 856, 2563). The Rugged armor kit enchants 8484-8492 only gain ItemLevelMin 15/25/35 (70009) | — | sie.py | landed | other enchant changes: Flametongue 3/4/523/1665/1666 lose the 415138 trigger (70170, shaman ledger), Comprehension imbues 8700/8706/8711/8714/8718 become stat enchants (70170, mage ledger), 847 renamed All Stats +1, 8734 Intellect added (70425) |
| SHR-U-12 | DB2 70374→70425 | Insight 18820 (an item proc): "next spell −500 mana" → all spells −200 mana for 12 s | n/a | aura 553→73, bp −500→−200, DurationIndex 1→29 | n/a | not in the sim or presets |
| SHR-U-13 | DB2 client-only builds | 70291→70334: only Summon Azurky (pet). 70334→70338 and 70338→70374: no spell changes. Nothing in this scope. 69977, 70058, 70124, 70205, 70235, 70245 have no DB2 snapshots of their own; the spanning diffs hold nothing that foreverchanges does not tie to an announced build | — | | n/a | confirms foreverchanges |

##### Race tables the DPS report uses

| ID | Source | Change | L60 DPS? | Client check | Sim status | Code location / note |
|---|---|---|---|---|---|---|
| SHR-RT-01 | PR #56 | Race tab text and parts (Ruthless, Earth Fury, Savage, Eureka 8 s, Blood Fury 3 min, ToG 20/30% 25 s, Endurance crit damage, Wind Blessed 3%) | yes | 70425 | landed | build_races.py PART_INFO :24-45, run_races.py PARTS :29-33 |
| SHR-RT-02 | tooling | The DASH_RACIALS_OFF switch the race breakdown needs is stale | yes | — | **partial** | tools/dps_report/patches/0002-…patch knows only `sword`, `mace`, `axe` (no `ruthless`, `earthfury`, `savage`), and `git apply --check` fails against master on racials_forever.go:14 and hunter.go:349 (the hunk patches `removeRacialRangedCrit`, which #56 deleted). run_races.py's `off=ruthless/earthfury/savage` would leave those racials on, so the breakdown would book them under "base stats" |

#### 3. Proposed fixes

1. **Grace of Air with Windfury (SHR-70425-24).** buffs.go:2009, `WindfuryTotemBuffAura`: drop the
   `aura.NewExclusiveEffect(graceOfAirWindfuryCategory, false, ExclusiveEffect{Priority: 2})` line under Forever
   (keep it for Era if Era needs it; it is Forever-only now), and rewrite the comment at :1616-1621. Test in
   `sim/core` (or `consumes_forever_test.go` style): a Forever rogue with raid `graceOfAirTotem` and
   `totemWeaponBuff: Windfury` keeps the 77 Agility while the Windfury buff is up. On its own this changes no preset
   (no melee req has Grace of Air). With a ruling that the level 60 raid has a second shaman (raid-wide totems make
   it the natural layout), add `graceOfAirTotem: Regular` to the rogue (12), warrior (6), feral (2), retribution
   paladin (10) and survival hunter (2) reqs: about +4-5% for rogues, +2-3% for warriors, feral and survival, +1-2%
   for ret (77 Agility).
2. **Touch of the Grave modifiers (SHR-70425-12).** racials_forever.go:355: flags
   `SpellFlagPassiveSpell | SpellFlagNoOnCastComplete | SpellFlagBinary | SpellFlagIgnoreModifiers`. Test: an Undead
   shadow priest in Shadowform with Shadow Weaving up drains exactly 5% of max health. Under 0.3% for the Undead
   presets (mage arcane/fire/frostfire, priest shadow/shadow-holy/smite, warlock destruction_fire, paladins, rogues,
   fury warrior); most for the shadow priest and the fire warlock.
3. **Eureka! and periodic damage (SHR-70170-01).** racials_forever.go:246-277: leave out spells whose damage is only
   periodic and not channeled (a `Dot`/`AOEDot` with no direct damage and no `SpellFlagChanneled`) from both the
   cost and the damage; for mixed spells (Immolate, Flame Shock, Holy Fire) raise only the direct part, e.g.
   `spell.ImpactDamageMultiplierAdditive += 0.1` in place of `DamageMultiplier *= 1.1` (spell_result.go:639 already
   splits impact and periodic); keep `DamageMultiplier` for channels. Test: a Gnome warlock's Corruption tick and
   cost are unchanged under Eureka!, Shadow Bolt is ×1.1 and −10% mana, Drain Life ticks ×1.1. Impact about 0.1-0.3%
   on the six Gnome warlock presets and the Gnome rows of the race table.
4. **Trueshot Aura composition (SHR-70425-21).** A ruling, then: `trueshotAura: true` for hunter__beast-mastery*,
   hunter__lone-wolf* and hunter__survival* (spec_hunter.go:54, :63) and fix notes.md:112. Rank 5 at 100 RAP is
   about +2-3% for a ranged hunter.
5. **Race-table patch (SHR-RT-02).** Re-export patch 0002 against master with the parts `ruthless` (Human crit),
   `earthfury` (Dwarf crit damage), `savage` (Orc crit), `endurance` (Tauren health + crit damage) and without the
   hunter.go hunk; check patch 0003 still applies (it does not touch the changed lines). No DPS change; it fixes the
   Races tab breakdown.
6. **Shadowstrike coefficient (SHR-U-01).** item_effects.go:2082: `BonusCoefficient: core.TernaryFloat64(isForever, 0, 1.0)`
   or 0 outright under Forever. No preset.
7. **Beast Slaying crit double dip (SHR-U-08).** Move patch 0003 to master: delete racials_forever.go:114 (and
   racials.go's Era line only if Era wants it). Beast / Elemental race-table variants only.

#### 4. Need to Verify candidates

* **Does Dwarf Earth Fury still need a mace?** Client 70425 replaced Mace Specialization (1259719) with Earth Fury,
  "Increases your critical strike bonus to damage and healing by 2%", with no word on weapons, and removed the
  equipped-item rows of Human Ruthless (255480) and Orc Savage (255488). But Earth Fury's own row, SpellEquippedItems
  255482 (weapon class, subclass mask 48: one- and two-handed maces), is still there, so the server may only apply it
  with a mace equipped. The sim gives it to every Dwarf whatever the weapon (racials_forever.go:36). Beta: a level 30
  Dwarf with a non-mace weapon lands crits on a training dummy, then again with a mace; if the crit sizes only move
  with the mace, the sim needs the old `hasWeaponOfType(Mace)` gate back (and the race table its weapon variant).
* **Grace of Air with Windfury Totem (2026-10-15 notes vs the sim).** Blizzard's notes and the client (Grace of Air
  25360 drops Attributes_11 0x400) say the two stack now; the sim still turns Grace of Air off while Windfury is up
  (buffs.go:2009). Beta: two shamans, one Grace of Air and one Windfury Totem, a melee in range; the character sheet's
  Agility (UnitStat) should keep the 77 while the Windfury buff is up.
* **Eureka! on damage over time and channels.** The 2026-10-01 notes say periodic effects get nothing and channels
  are not periodic. The 70425 masks agree (DoTs out of the cost and damage masks, a periodic-damage mod only for
  Drain Life, Drain Soul and Mind Flay), and the sim gives every DoT cast inside the window both the 10% discount and
  10% on all its ticks. Beta: a Gnome warlock casts Corruption with and without Eureka! up (mana cost and tick size),
  then Drain Life (ticks should be 10% bigger).
* **Trueshot Aura in the level 60 raid.** Client 70425 makes it a raid area aura (effect 65, "party or raid
  members"); notes.md:112 still treats it as party-wide, and only the Marksmanship presets have it. Not a beta test but
  a composition ruling: does the reference raid carry a Marksmanship hunter, so every hunter preset gets 100 ranged
  attack power?
* **The 70009 EffectBonusCoefficient wipe.** 1,029 of the 1,043 zeroed rows are effects where the coefficient has no
  meaning (stat auras, triggers, dummies, item creation), so it reads as a client data cleanup. If the server reads the
  client value for procs, about 24 damaging item procs (Shadowstrike's Drain Life 461683, Masterwork Stormhammer's
  Chain Lightning 463946, the Shock 2606-2610 procs…) no longer scale with spell power; none is in a preset. A beta log
  of one such proc at two spell power values would settle it; low priority.

#### 5. Count

61 rows (announced: 70425 30, 70291 5, 70170 4, 70009 7; unannounced 13; race tables 2): landed 25, missing 2,
partial 3, wrong 2, not simulated on purpose 2, in flight (worker PR) 1, ambiguous (needs verify) 2, n/a 24.
