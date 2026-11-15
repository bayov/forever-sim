# update_1115: PRs #129 and #130, the new Windfury Totem rule, re-optimization

Landed on master at 25ce81de7 (dev branch `update-1115`, same commit). DPS Report deployed:
https://forever-dps.pages.dev, deploy id `53d112b1` (44 files, each checked live).

## Changes and how each was confirmed

- **PR #129** (build 70938 items, Wild Growth 0.066 a tick, Imp Firebolt ranks, Seal/Rune of the
  Dawn). Merged at b02b98c:
  - `forever_seen_in_game.json` merged by item, taking the later `seen` date.
  - `forever_items.go` kept both sides; the db was regenerated.
- **PR #130** (Windfury Totem under Forever). Merged at d4325b7:
  - Every landed main-hand melee hit rolls the totem, white or special (`windfuryTotemRollsOnForever`, `sim/core/buffs.go`).
  - Extra attacks and the Seal of Command proc (`SpellFlagNoWindfuryTotem`) never roll it.
  - The internal cooldown is 100 ms.
  - Judgement of Command is a melee special again; Righteousness and Fury stay ranged.
  - The old client table and its generator are gone.
  - Confirmed by tests: TestOrcShamanWindfuryTotemOnSpecials, TestJudgementsAndWindfuryTotem (replaces TestJudgementsDontRollWindfuryTotem), TestForeverWindfuryTotemFurySpecials, `windfury_totem_rolls_forever_test.go`. Standards 17 rewritten.
  - Edge case kept: a Windfury Weapon attack set off by a totem extra swing doesn't roll the totem.
- **Master's Judgement of Command change (6ce77e5).** It is the same change as PR #130's, so one implementation was kept and the comments joined. The paladin tests pass.
- **Wild Growth.** The three ranks already shared one timer. TestDruidWildGrowthRanksShareCooldown pins it, with category 1237 cited.
- **Windfury Totem's attack power stays** (+95/+179/+246; user ruling). The TASKS item stays open for a log.
- **New assumptions, both tagged "Probably a bug" on the page.** The chip shows in the assumption list, the tooltip and More details; the Method tab has no assumption list.
  - Execute rolls the totem twice (knob `warr_exec_wf2`). The default is once. Applies to Arms, Arms vs Dragonkin, Fury and Fury 2H; Tanking has no assumptions.
  - Blood Fury's weapon damage (knob `orc_bf_weapon=0` gives AP only). The default keeps the weapon damage. Applies to the 4 Orc warriors, the 4 Enhancement presets, and BM, MM and Survival.
  - Test: Execute set off the totem on 0.170 of casts with one roll and 0.314 with two.
- **Mutilate** still rolls once (user ruling); nothing changed.
- **Full test suite** (`./sim/...` without sim/web): every package passes.
  - The only failure was TestRaidForeverReportHashesStable, on the 30 intended moves. The hashes were rewritten at publish and the test passes now.
  - No `*.results` golden moved.
  - Nothing failed before the round that still fails. The shaman golden failures noted on master fe8793d are gone.

## Moved builds, 180 s DPS (live before → published)

All builds were stale. The probe found 32 moved: the 30 below plus BM and MM, which only gain the new Blood Fury assumption rows. The three audit builds (priest Smite, Demonology and Destruction Fire curse duties) were re-simmed in full and came out unchanged.

| Build | Before | Now | Cause |
|---|---|---|---|
| Fury | 1051.1 | 1092.8 | Bloodthirst, Whirlwind's main-hand strike and the other specials now roll the totem |
| Fury 2H | 1032.9 | 1103.0 | Same; a two-hander's specials gain the most |
| Arms | 1043.3 | 1069.3 | New rule (Mortal Strike, Whirlwind, Overpower roll), plus the Windfury gate pick |
| Arms vs Dragonkin | 1101.4 | 1126.6 | New rule |
| Ret (twist, level_60) | 980.7 | 1023.0 | Holy Strike and Judgement of Command roll the totem; Command is melee again; master's Two-Handed Weapon Specialization Holy change; 1 Unsealed Wrath stack pick |
| Ret Shock | 951.9 | 1009.9 | Same, plus the switch to Seal of Command with no twist |
| Ret Sub Holy | 971.0 | 1011.6 | Same as twist, 1 stack |
| Ret Undead | 1102.2 | 1138.0 | Same, 2 stacks |
| Enhancement (4) | 1040–1054 | 1066–1080 (about +26) | Stormstrike rolls the totem, plus the hold for Windfury Weapon |
| Feral cat | 1155.5 | 1185.9 | Shred, Claw, Rake, Rip and Bite roll the totem, plus Bite at 65 and the gate |
| Survival (3) | 1014.5 / 981.9 / 991.5 | 1034.6 / 1001.8 / 1011.1 | Mongoose Bite, Wing Clip and the other main-hand instants now roll the totem (Raptor Strike already did) |
| Subtlety | 863.0 | 866.8 | Hemorrhage now rolls the totem; the other rogue builds' abilities already rolled it under the old table |

The duty presets moved with their bases:
- Arms: Battle Shout 1065.0, Sunder 1023.5.
- Fury: Battle Shout 1088.7, Sunder 1055.0.
- Fury 2H: Battle Shout 1099.3, Sunder 1066.4.
- Feral FF duty: 1179.6.
- Ret duties: about +8 to +10 each over the old picks.

Master had republished the paladin numbers before this round's totem rule (8b12a8a); this publish replaces them.

## Picks (each wins on seeds 1 and 1000003 at 120, 180 and 300 s, 20,000 iterations)

- **Arms and its two duty presets: wfGate=1.** Main-hand abilities wait out the 100 ms after a white swing.
  - Arms: +0.4 to +2.0.
  - Sunder duty: +0.5 to +2.0.
  - Battle Shout duty: +0.2 to +1.3. That margin is narrow, so it needs a ruling.
- **Prot warrior Balanced and Max survival:**
  - Heroic Strike from 40 rage, Revenge ahead of Shield Slam, Thunder Clap refreshed under 8 s.
  - Balanced 3434.8 → 3450.9; Max survival 1773.1 → 1797.7.
- **Enhancement, all 4: Stormstrike waits for Windfury Weapon's 1.5 s cooldown.** +0.9 to +2.4 in every cell. It beat the idle hold and the totem gates.
- **Feral cat: Bite at 65 Energy (was 45) plus wfGate=2.** +2.7 to +4.1. The gate's own share is +0.6 to +1.1, which is narrow, so it needs a ruling.
- **Feral FF duty: Bite at 65.** +1.0 to +2.4. The gate wasn't kept there.
- **Ret paladin**, re-checked after the Two-Handed Weapon Specialization Holy change; nothing flipped:
  - level_60: 1 stack, twist window 25, mana-rate limit off. +4.8 to +6.7.
  - Sub Holy: 1 stack. +3.6 to +4.6.
  - The four Crusader and Wisdom duties: 1 stack. +8.2 to +10.4.
  - Undead: 2 stacks. +1.9 to +2.6.
  - Shock and Shock Wisdom: Seal of Command up, no twist. +33 to +38.

## Builds where nothing won

- **Fury and Fury 2H** (and their duties), **Arms vs Dragonkin**:
  - wfGate=1 lost or was noise: Fury −1.5 to +0.1; Dragonkin −1.5 to +0.4; Fury 2H mixed, −0.8 to +1.7.
  - wfGate=2 (spreading abilities between swings) lost: Fury −1.4 to −4.2.
  - The knob search found nothing.
- **Prot warrior Max threat**: Shield Block at 30 rage was noise; the gates lost.
  - Its Tanking numbers came out bit-identical. That is expected: Revenge, Sunder and Heroic Strike already rolled the totem under the old table, and Shield Slam is an off-hand special that rolls under neither rule.
- **Rogues, all 6** (Combat Swords and Daggers, Seal Fate Swords and Daggers, Mutilate, Subtlety). 180 s seed 1: 943.4 / 936.1 / 939.1 / 904.0 / 864.9 / 866.4.
  - The search found no gain over 1 DPS anywhere.
  - mhGap=0.1 was noise; Seal Fate Daggers gained +1.1 on seed 1 but not on seed 1000003.
  - Waiting for just before a swing lost up to 10.2; waiting to Eviscerate just before a swing lost up to 12.5.
  - Combat Swords keeps 5/5 Sword Specialization (each point moved out costs about 6). Seal Fate Swords stays without it (−22 to −70).
- **Survival (3)**:
  - Wing Clip or all instants held 0.1 s after a swing: −1.0 to +0.1.
  - Waiting for the next swing lost 19 to 61.
  - The knob search (126 runs) moved nothing.
- **Bear (3 presets)**: the gates lost (−7 to −177).
  - Max survival's Maul reserve was noisy (+20.8±9.6 on seed 1, +3.6±10.1 on seed 1000003), so it wasn't taken.
- **Prot paladin (3 presets)**:
  - The Holy-damage gate (wfIcdGate) and swingGap lost or were noise (−0.7 to 0).
  - Seal of Righteousness lost 43 to 204.
- **Ret gates**:
  - wfIcdGate: level_60 +0.1, Shock +0.0 to +0.3 (noise).
  - swingGap: −0.1 to −3.7.
  - Command with no twist on the twist builds: −12 to −16 (Undead −2 to −4).

## Timing-gate check finished (the rotation-audit pass can skip these)

- Warrior: Arms and its 2 duties, Arms vs Dragonkin, Fury and its 2 duties, Fury 2H and its 2 duties.
  - Fury's duty presets and Dragonkin only had the gate tried, with no knob search of their own.
- Prot warrior: Max threat, Balanced, Max survival.
- Enhancement: all 4.
- Feral cat and its FF duty; Bear: Max threat, Balanced, Max survival.
- Rogue: all 6.
- Survival: all 3.
- Ret paladin: all 10.
- Prot paladin: all 3.

Not checked: BM and MM hunters. They weren't in scope, since their shots don't use the totem.

## Tanking and Healing

- **Tanking**, each preset's score on its own goal:
  - Warrior: Balanced +22, Max survival +19, Max threat unchanged.
  - Paladin: +24 / +27 / +53.
  - Bear: +339 / +289 / +89. The new rule moves the bear most: its specials outside the old client table now roll the totem.
  - Warden: +84 / +59 / +116.
  - No rank changed.
- **Healing**: Wild Growth's coefficient doubled, yet the raid Resto Druid's carry fell from 1.414 to 1.392, moving it from 3rd to 4th. It spends less mana (healing per mana 2.19 → 2.44). Worth a look, but no ruling is needed.

## Assumption rows (180 s, alternate vs default)

- **Execute twice:** Fury 1092.8 → 1106.1, Fury 2H 1103.0 → 1125.6, Arms 1069.3 → 1092.1, Arms vs Dragonkin 1126.6 → 1146.4. That is +13.4 to +22.9.
- **Blood Fury as attack power only:** −5.3 to −6.5 for the warriors, −3.6/−3.7 for Enhancement, −1.8 to −3.5 for the hunters.

## Needs a user ruling (all four answered 2026-11-15, see Rulings below)

1. **Fury 2H (1103.0) now beats the default dual-wield Fury (1092.8) at 180 s.** The 2026-11-12 "switch to best" ruling would make it the Warrior default, as 7462c17 did. It hasn't been switched; it's flagged in the wiki.
2. **Narrow picks:** Arms Battle Shout duty's gate (+0.2 to +1.3), and the Feral cat gate's own share (+0.6 to +1.1).
3. **The Enhancement Stormstrike hold** depends on Windfury Weapon's 1.5 s cooldown being separate from the totem's (assumption `wf_weapon_totem_icd`).
4. **Ret twist builds:** `primarySeal` says Command where the main seal is Righteousness. This predates the round.

## Rulings (user, 2026-11-15)

1. **Two-hand Fury is the Warrior class default** ("switch to best": 1103.0 against dual wield Fury's 1092.8 at 180 s, deploy `53d112b1`). Done like 7462c17c8: both metas' `isDefault`, the UI's default gear, talents and rotation, the build order and the wiki. Only the flag moves; no re-sim, and the race sweep needs no run, since the report's race rows are per build (fury_2h has its own).
2. **Windfury Weapon and Windfury Totem have separate cooldowns** (weapon 1.5 s, totem 100 ms). Assumption `wf_weapon_totem_icd` stays as set and no longer says unknown; standards 17 states it. The Enhancement Stormstrike hold (Stormstrike waits for Windfury Weapon's 1.5 s cooldown) depends on it and stands.
3. **Keep both narrow picks:** the Arms Battle Shout duty's swing gate (+0.2 to +1.3) and the Feral cat gate during the totem's cooldown (+0.6 to +1.1).
4. **Ret twist builds' `primarySeal` label: left as is, user OK 2026-11-15.** The rotation casts its seals directly, so the label has no effect.

## Also

- **Changes tab:** public lines for:
  - the Windfury Totem rule;
  - the rotation picks;
  - the two probable-bug assumptions;
  - the Prot warrior re-tune;
  - Wild Growth;
  - Imp Firebolt ranks;
  - items;
  - Wildstalker 5pc.

  The Judgement of Command and Two-Handed Weapon Specialization lines were already there. `changes.py check`: 57 commits change build inputs, 0 uncovered.
- **Other updates:** the metas, presets.ts comments (Arms, Feral, Ret), the wiki (sim-status and six class pages), state.json and raid_sim_data.
- **compsearch `-stale`:** lists only four searches not stamped (m1_10_class_buffs, m1_20, m1_20_class_buffs, m1_40_class_buffs). Not re-run.
- **TASKS:** Wild Growth and the shaman-golden items are ticked. Left open for a log:
  - Windfury Totem's attack power (push 114091);
  - how its extra attack strikes;
  - the seal procs' non-roll (PR #130 reading);
  - the Execute and Mutilate rolls.
- **PRs #129 and #130** are closed with a comment saying they landed through master.
- **Merge note:** one merge commit on update-1115 (25ce81de7's parent chain) has a leftover "# Conflicts" block in its message. It can't be fixed without a force-push, so it stays.
