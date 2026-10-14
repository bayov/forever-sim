# Spec assumptions: inventory for the dashboard's assumption scenarios

Step 1 ("Inventory") of [dashboard_next_round.md](dashboard_next_round.md), 2026-10-13. For each Level 60 preset, this lists the assumptions where Forever's real value is unknown or estimated and that move the preset's DPS by about 0.5% or more, or could change a ranking. Each one gets three settings (Low / Expected / High) for the dashboard's segmented control.

Sources, all read-only:
- The merged tree `/home/user/w-dash`, branch `dashdata` (head 7e06ebde6, all nine class branches). Every `file:line` below is in that tree.
- `notes.md` "Need to Verify", `forever-wiki/sim-status.md`, `shaman_audit.md`.
- `orchestration/rank_tuning.md`, `pet_buffs.md` and `standards.md` on this branch.
- The bodies of class PRs #1 to #9 (Priest, Warlock, Paladin, Warrior, Hunter, Mage, Shaman, Druid, Rogue).
- Code comments marked TODO, guess, assumed or placeholder.
- The web, for analogues: wowsims/sod's Maelstrom Weapon code and warcraft.wiki.gg's Season of Discovery rune page.

## Conventions

- **Page wording:** the report shows each assumption in plain words (a question-style label, one line on what's unknown, setting labels that say what each choice means). That wording lives in `tools/dps_report/assumptions_def.py` (rewritten 2026-10-14); the labels in this file are the inventory's working names, matched by key.
- **Sim now** marks the setting the shipped presets use. Use it as the dashboard's default, so the default ranking matches the presets. "Expected" is our best guess for release; where it differs from Sim now, the entry says why.
- **Sensitivity:**
  - **PR**: measured by the class session and quoted from its PR, wiki or notes. The gear set is named when it isn't the final 1b set.
  - **est.**: my back-of-envelope estimate from damage shares and talent multipliers. No sim was run for this inventory. Step 3 replaces every est. with a measured delta.
  - Percentages are of the build's 180 s DPS on the final presets (Undead / Gnome / Night Elf / Orc as shipped).
- **High = Expected** means no evidence supports a third value; the dashboard can show two segments or a disabled third. These entries are listed in the summary at the end.
- **Knob:** the step 2 patch. These are dashboard-only patches in the local merged tree.
  - **Convention:** one helper in a new `sim/core/assumptions.go`, `core.Assume(key string, def float64) float64`. It parses the `DASH_ASSUME` environment variable once (`key=value,key=value`) and returns `def` when the key is absent.
  - **Precedent:** this follows the existing `DASH_NAMES` hook in `sim/core/environment.go:197`, and the rank tuning auditor's env-var hooks.
  - **Patch contents:** each patch replaces the constant at the listed line with `core.Assume("<key>", <constant>)`. A few need a second line or a branch, as noted.

## Summary

| Class | Assumptions | Keys |
|---|---|---|
| All classes | 4 | `flat_bonus_inside_pct`, `wf_totem_offhand`, `downrank_penalty`, `oil_proc_damage` |
| Priest | 2 | `priest_sf_tick_crit`, `priest_smite8_tuning` |
| Warlock | 5 | `lock_pet_scaling`, `lock_demon_melee`, `lock_bod_coef`, `lock_incinerate_bonus`, `lock_decimation` |
| Paladin | 4 (+2 shared) | `pal_sor_sp_double`, `pal_consecration_coef`, `pal_2h_spec_holy`, `pal_31plus_ranks` (+ `flat_bonus_inside_pct`, `oil_proc_damage`) |
| Warrior | 3 (+1 shared) | `warr_rage_k`, `warr_rage_oh`, `warr_crit_rage` (+ `wf_totem_offhand`) |
| Hunter | 6 (+1 shared) | `hunter_pet_scaling`, `hunter_pet_speed`, `hunter_tsa_hm_rank`, `hunter_shot_ap_shares`, `hunter_crit_per_agi`, `hunter_raptor_31plus` (+ `wf_totem_offhand` for Survival) |
| Mage | 3 | `mage_ice_lance_coef`, `mage_ab_on_missiles`, `mage_frostbolt10` |
| Shaman | 3 (+1 shared) | `sham_mw_ppm`, `sham_lb_31plus`, `sham_es7_tuning` (+ `oil_proc_damage` for Enhancement) |
| Druid | 4 | `druid_ooc_rate`, `druid_shred_flat`, `druid_rip_31plus`, `druid_wrath8_tuning` |
| Rogue | 0 (+2 shared) | `flat_bonus_inside_pct` (Backstab, Mutilate), `wf_totem_offhand` |

That makes 34 assumptions. The rogue has no material assumption of its own: the rank tuning audit found its spells Era-identical or client-tuned, and its open items are under 0.5% or client-sourced (see "Considered and left out"). Its two scenarios are the shared ones.

### The ones that matter most

1. **`mage_ice_lance_coef`.** Ice Lance's spell power coefficient is 0 in the client; Season of Discovery's server gave the same spell 42.9%. At 42.9%, Frostfire and Frost gain about 165 DPS (+18% / +22%, PR, earlier gear) and Frost moves from last to mid-table among mages. This is the largest single swing in the inventory.
2. **`lock_pet_scaling` + `lock_demon_melee`.** Both are Season of Discovery placeholders or 1.12 carry-overs.
   - Without owner scaling, every warlock build loses about 4%; with the Succubus melee halved, another 4%; with both, 7% (PR).
   - Demonology leads the ranking (1109.8), and the other three builds sit just above Retribution (969 to 985 against 960 to 967). These decide that order.
3. **`lock_bod_coef`.** Bane of Doom's 4.0 coefficient is about 12% of every Shadow Bolt build's damage. At 2.0, every Shadow Bolt warlock loses 4.1% (PR), and Destruction and Affliction drop below the Retribution builds. I still expect 4.0 (it is the duration/15 rule), but it is untested in any beta build.
4. **`pal_consecration_coef` and `pal_sor_sp_double`.** Consecration's 9.5% a tick comes from an old tooltip, against Era's 4.2% (est. −4% at Era's value). Seal of Righteousness counting spell power twice comes from Season of Discovery tests (−5.1% twist, −4.4% Shock if counted once, PR). Together these could move every Retribution build about 9%.
5. **`hunter_pet_speed` and `hunter_pet_scaling`.**
   - Pet speed: Beast Mastery's 1.0 sec cat hitting as hard as a 2.0 sec pet comes from beta players, and nobody knows if it is intended (est. −5 to −8% Beast Mastery if Classic's speed scaling returns).
   - Pet scaling: Season of Discovery's 22% scaling instead of the measured 10% would add 41 / 23 / 18 to Beast Mastery / Marksmanship / Survival (PR).

Close behind: `mage_ab_on_missiles` (est. −8% / +4% Arcane), `flat_bonus_inside_pct` (est. −3.7% dagger rogues and +3% Retribution, in opposite directions) and `sham_lb_31plus` (+1.8% / +5.8% Elemental, PR).

---

## All classes

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `flat_bonus_inside_pct` | Flat bonus inside the weapon percent | Combat Daggers, Seal Fate Daggers, Subtlety (Ambush), Mutilate; Ret twist, sub Holy, Shock | inside (client effect order) | after the percent (tooltip read literally) | inside | = Expected | est. Backstab builds −3.7 to −4%, Mutilate +1.7%, Ret +3% (all at Low) |
| `wf_totem_offhand` | Windfury Totem procs from off-hand hits | Fury (dual wield), all six rogue builds, Survival | yes | main hand only | yes | = Expected | est. −2 to −3% DW Fury and rogues, about −1.5% Survival (at Low) |
| `downrank_penalty` | Downrank penalty size at 60 | Smite (rank 7 fallback; rank 4 before), Elemental (Lightning Bolt VIII fallback) | FC formula on both branches (a local helper on each) | stronger (10% a level past 17) | FC formula (5% a level past 17) | none | PR Smite −0.3% at 180 s, −6.3% at 300 s (FC vs none); est. Elemental about −1% at 300 s, 0 at 180 s |
| `oil_proc_damage` | Shadow Oil and Frost Oil proc damage | Ret twist, sub Holy, Shock (and their Undead encounters); all four Enhancement builds (all Shadow Oil) | Forever's tooltips (63–71 / 39–43) | Era's ranges (52–61 / 33–38) | tooltips | = Expected | −0.06 to −0.08% at Low (about 0.6 DPS; assumption sims, 180 s) |

### `flat_bonus_inside_pct`: flat bonus inside the weapon percent

- **Builds:**
  - Rogue: Combat Daggers, Seal Fate Daggers (Backstab), Subtlety (Ambush, about 2 a fight) and Mutilate.
  - Paladin: Level 60 twist, sub Holy and Shock (Holy Strike).
  - Shred follows the same rule but has its own tooltip question, so it is `druid_shred_flat`.
- **Sim now:** these strikes add a flat amount to the weapon damage before the weapon percent, so the flat part gets multiplied too:
  - Backstab rank 8 is 1.5 × (weapon + 140) = +210. `sim/rogue/backstab.go:14-16`, applied at `:68`.
  - Ambush is the same at 2.5×. `sim/rogue/ambush.go:13-14`, `:59`.
  - Mutilate is 0.75 × (weapon + 67) = +50.25 a hand. `sim/rogue/mutilate.go:15`, `:95-96`.
  - Holy Strike rank 8 is 0.5 × (normalized weapon + 93 + 42.9% of spell power). `sim/paladin/holy_strike.go:51`, `:99`, `:102-103`.
- **Source:** the Forever client (1.60.1.70338). Each spell is a normalized-weapon effect (121) carrying the flat part, followed by a weapon-percent effect (31). Backstab's effect data is identical to Era's, where the server adds the flat part before the percent (the rule that gives Era's Backstab 225 at rank 9).
- **Why uncertain:** Forever's tooltips print the raw flat value as if it came after the percent: Backstab "plus 150" at rank 9 where Era's says 225, and Holy Strike "plus 93". No beta log has shown a Backstab or Holy Strike hit with known weapon damage (notes.md, Rogue and Paladin). If the server really applies the tooltip, the flat part moves outside the percent. That lowers Backstab and Ambush and raises Mutilate and Holy Strike (Holy Strike's spell power part is then no longer halved).
- **Settings:**
  - **Low, "after the percent".** Read the tooltip literally: Backstab 1.5 × weapon + 140, Mutilate 0.75 × weapon + 67, Holy Strike 0.5 × weapon + 93 + 42.9% of spell power. Forever's tooltip generator would then be the authority.
  - **Expected, "inside" (Sim now).** The client's effect order, and Era's server rule on identical Backstab data.
  - **High:** = Expected. No third reading exists.
- **Sensitivity (est., at Low):**
  - Backstab loses 70 base a hit: about 148 a landed Backstab after Opportunity, Aggression, Murder and its crits, or −26 DPS (−3.7%) on Combat Daggers and about −4% on Seal Fate Daggers.
  - Subtlety loses under 0.5% (Ambush only).
  - Mutilate gains about 33 base a cast, +1.7%.
  - Holy Strike gains about 196 base a cast, about +30 DPS (+3%) on each Retribution build.
- **Knob:** one key read in three places:
  - `sim/rogue/backstab.go:68` and `sim/rogue/ambush.go:59`: when the key is 0, compute `weapon*mult + flat*mult/…`. Simplest: divide `flatDamageBonus` by the weapon multiplier (1.5, 2.5) up front.
  - `sim/rogue/mutilate.go:96`: use `flat/0.75`.
  - `sim/paladin/holy_strike.go:99,103`: take the flat part and its 0.429 coefficient out of the `rank.weaponPercent` product.

### `wf_totem_offhand`: Windfury Totem procs from off-hand hits

- **Builds:** Fury (dual wield), all six rogue builds, Survival. These are the dual wielders in a Windfury Totem group (`tools/rotopt/mksettings/spec_warrior.go:12`, `spec_rogue60.go:16`, `spec_hunter.go:38`). Two-hand Fury, Arms and Retribution have no off hand.
- **Sim now:** yes. Under Forever the totem's proc mask is `ProcMaskMelee`. `sim/core/buffs.go:1893-1896` (`WindfuryTotemBuffAura`).
- **Source:**
  - The Forever client (70291, 70338) makes the totem a party aura with ProcTypeMask 0x14 and nothing tying it to a weapon.
  - The user expects off-hand procs (notes.md, 2026-10-10).
  - Era put an enchant on the main hand, so only main-hand hits procced it.
- **Why uncertain:** the proc rules are server side and no log has checked them. The totem needs level 32, beyond the beta's cap.
- **Settings:**
  - **Low, "main hand only".** Era's behaviour, and the sim's own Classic ruleset.
  - **Expected, "any melee hit" (Sim now).** The client's party aura has no weapon condition, and the user expects it.
  - **High:** = Expected.
- **Sensitivity (est., at Low):** with the settled 1.5 sec ICD, off-hand hits add proc chances between main-hand swings.
  - About a third of the totem's procs on a dual wielder come from off-hand events.
  - The totem is worth about 45 DPS to a rogue (+24 over Grace of Air, PR), so main hand only costs rogues and dual-wield Fury about 2 to 3%, and Survival about 1.5%.
- **Knob:** `sim/core/buffs.go:1894-1896`, set `procMask = ProcMaskMelee` only when the key is 1.

### `downrank_penalty`: downrank penalty size at 60

- **Builds:** only the presets that cast a spell-power rank learned more than 17 levels below 60, i.e. before level 43. The Forever class changes auditor checked every Level 60 preset's casts (result JSON, 300 iterations) and every Level 60 APL's spell IDs against client 70338 SpellLevels (2026-10-13). Two builds qualify:
  - Priest Smite: a low-rank Smite fallback when mana runs short. It was rank 4 (level 22); since the penalty it is rank 7 (level 46).
  - Elemental shaman (Level 60 and + Mana Tide): Lightning Bolt VIII (10391, level 38) below 20% mana. At 180 s it is 0.2 casts and 1 DPS; at 300 s 15.6 casts and 58 DPS.
  - No other preset casts such a rank. The warrior, rogue, hunter and feral low-level IDs are single-rank or attack-power abilities, cooldowns and stances; the warlock, mage, Balance and Ret rotations use top ranks only.
- **Sim now:** the priest branch applies foreverchanges' formula to Smite, Holy Fire, Mind Blast, SW:Pain and Devouring Plague (`sim/priest/priest.go` `downrankCoefficient`, commits 6ee9d02eb / ea9eb5c22 / 333cbeb04). The shaman branch has no penalty (`sim/shaman/lightning_bolt.go:25` keeps rank 8's full 0.714).
- **Source:**
  - Blizzard's development notes, 2026-10-01 (us.forums.blizzard.com/en/wow/t/2360696), tier 3: ranks "vastly below your current level receive less benefit from your Spell Damage", and they trigger class abilities less often (at 60 a rank 1 Frostbolt never procs Frostbite).
  - foreverchanges.pro/downranking, tier 4: a rank keeps 1 − 0.05 × (L − 17 − T) of its coefficient (L your level, T the rank's learn level), between 0 and 1. Classic's level-20 cut is gone. It was measured on heals in the beta up to level 30; damage spells and level 60 are its author's assumption.
- **Why uncertain:** the formula was never measured above level 30 or on a damage spell. The proc part has no numbers at all.
- **Settings:**
  - **Low, "stronger".** 10% a level past 17, so a rank keeps nothing from 27 levels below. No source; it brackets a steeper curve at 60.
  - **Expected, "FC formula".** 5% a level past 17; nothing after 37 levels below.
  - **High, "none".** Classic Era's rule minus its level-20 cut, i.e. no penalty for any rank the presets cast.
- **Sensitivity:**
  - Smite: FC vs none, with the rotation re-searched under the penalty: 708.8 → 706.5 at 180 s, 687.0 → 643.8 at 300 s (PR, priest branch).
  - Elemental, est.: Lightning Bolt VIII keeps 75% under FC (5 levels past 17). That is about −5 to −7 DPS (−1%) at 300 s and 0 at 180 s, before re-tuning the fallback to rank IX (level 44).
- **Knob:** one shared helper (proposal): move the priest's `downrankCoefficient` to `sim/core` as `(*Character).ForeverDownrankCoefficient(coef, learnLevel)`, reading `core.Assume("downrank_penalty_pct", 5)` (0 = none). The priest session owns the move. The shaman applies it in `lightning_bolt.go` / `chain_lightning.go`; nobody else needs it.

### `oil_proc_damage`: Shadow Oil and Frost Oil proc damage

- **Builds:** every preset with an oil on a weapon: the three Retribution builds and their three Undead encounters, and all four Enhancement builds. All of them use Shadow Oil; no preset uses Frost Oil, which takes the same setting.
- **Sim now:** Forever's tooltips. Shadow Oil's proc (Shadow Bolt, 350026) deals 63 to 71, Frost Oil's (Frostbolt, 350025) 39 to 43. `sim/core/consumes.go:231-243` (`shadowOilDamageRange`, `frostOilDamageRange`); Classic Era keeps 52 to 61 and 33 to 38. The proc chances (15% and 10%) are Classic's.
- **Source:**
  - Wowhead's Forever tooltips (dataEnv 16, checked live): 350026 says 63 to 71, 350025 says 39 to 43. Wowhead Classic's tooltip for 350026 says 52 to 61.
  - The consumables auditor read the server formula from the TrinityCore port: the damage scales with `level -= SpellLevel`. Forever changed the spells' BaseLevel to 0 but kept their SpellLevel (12 for Shadow Oil, 8 for Frost Oil), so the server formula gives Era's ranges, while the tooltip shows the higher ones.
- **Why uncertain:** the tooltip and the server formula disagree, and no beta log shows an oil hit at level 60.
- **Settings:**
  - **Low, "Era's".** Shadow Oil 52 to 61, Frost Oil 33 to 38: the server formula with SpellLevel 12 (Shadow) and 8 (Frost).
  - **Expected, "tooltip" (Sim now).** Shadow Oil 63 to 71, Frost Oil 39 to 43, as Forever's tooltips print.
  - **High:** = Expected.
- **Sensitivity (assumption sims, 180 s, at Low):** −0.06 to −0.08% on every Retribution and Enhancement build, about 0.6 DPS (Ret twist 981.7 → 981.0, Enhancement 813.9 → 813.4). The same at 120 and 300 s. Below the 0.5% threshold, but listed by user ruling (2026-10-13) because the evidence conflicts.
- **Knob:** `sim/core/consumes.go`, both range helpers return Era's ranges when `core.Assume("oil_era", 0)` is 1.

---

## Priest

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `priest_sf_tick_crit` | Shadowform's 200% crit on ticks | Shadow, Shadow + Holy | 200% on SW:P, Mind Flay, DP ticks | 150% on ticks | 200% | = Expected | est. −3% (Low) |
| `priest_smite8_tuning` | Smite VIII 31+ tuning | Smite | client (×1.00) | ×1.00 | ×1.09 | ×1.18 | +1.3% at High (PR) |

### `priest_sf_tick_crit`: Shadowform's double crit bonus on ticks

- **Builds:** Shadow, Shadow + Holy.
- **Sim now:** Shadowform adds +100% crit bonus (crits deal 200%) to Shadow Word: Pain, Mind Blast, Mind Flay and Devouring Plague. Their ticks crit under Forever, so the ticks get 200% crits. `sim/priest/talents.go:346-348` (the mask) and `:361-362`. The periodic crit rule is `sim/core/ruleset.go:15-16`.
- **Source:** Shadowform's client class mask covers the three periodic spells. That only matters if ticks crit, which the beta settled for Flame Shock (shaman PR #7, "Periodic crits: confirmed in the beta logs"). Forever's Inner Focus tooltip ("a non-periodic spell") agrees.
- **Why uncertain:** no priest log at level 30 shows a Shadow Word: Pain or Mind Flay tick crit, or its size (notes.md, Priest).
- **Settings:**
  - **Low, 150% on ticks.** The double bonus reaches only Mind Blast, the server treating it as a direct-damage modifier, as most crit-damage auras were in 1.12.
  - **Expected, 200% (Sim now).** The client's mask names the periodic spells; WotLK's Shadowform also let these periodic spells crit for double.
  - **High:** = Expected.
  - "Ticks never crit" is not offered: the beta showed periodic crits.
- **Sensitivity (est., at Low):** ticks are about 69% of Shadow's damage (Mind Flay 44%, Shadow Word: Pain 20%, Devouring Plague 5%; rank tuning audit), with about 10% crit. Halving their crit bonus costs about 0.69 × 0.10 × 0.5, about −3.5%; call it −3%, since Shadow Weaving and the other multipliers stay.
- **Knob:** `sim/priest/talents.go:361`. Give the periodic spells `CritDamageBonus += Assume(...)` (0.5 or 1) and keep Mind Blast at 1. That needs a split of `critSpell` into direct and periodic.

### `priest_smite8_tuning`: Smite VIII looks partly untuned

- **Builds:** Smite.
- **Sim now:** client values, Smite rank 8 166.62 to 186.38 base. `sim/priest/smite.go:16` (rank 8 entry) and `:17` (growth).
- **Source:** client 1.60.1.70338.
- **Why uncertain:** Smite's ratio to Era is 0.51 at rank 4 (level 30) but 0.42 / 0.41 / 0.43 at ranks 6 to 8. Build 70291 retuned ranks 1 to 4 and no 31+ rank (rank tuning audit).
- **Settings:**
  - **Low, ×1.00 (Sim now).** The client as is: no 31+ rank has changed in 16 beta builds.
  - **Expected, ×1.09.** Halfway. Blizzard's level-curve passes have moved the top playable rank 0 to 12%, so a partial retune before release is likely but its size is not known.
  - **High, ×1.18.** Held at rank 4's 0.51 ratio to Era.
- **Sensitivity:** High +1.3% (PR, rank tuning audit); Expected about +0.65% (est.).
- **Knob:** multiply rank 8's `SmiteBaseDamage` (`sim/priest/smite.go:16`) by the key, applied where the range is built (`:37-38`, `if rank == 8`).

---

## Warlock

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `lock_pet_scaling` | Demon scaling from the warlock | all four | SoD placeholder | none | SoD | SoD + warlock spell crit | Low −4.5% Demo, −4.0% others (PR) |
| `lock_demon_melee` | Succubus melee damage | all four | 1.12 (95–131) | halved | 1.12 | = Expected | Low −3.9 to −4.0% (PR) |
| `lock_bod_coef` | Bane of Doom coefficient | Demo, Destro, Aff | 4.0 | 2.0 | 4.0 | = Expected | Low −4.1% (PR; 1.0 gives −6.1%) |
| `lock_incinerate_bonus` | Incinerate's 25% with Immolate | Destruction Fire | whole hit | base damage only | whole hit | = Expected | est. −4% Fire (Low) |
| `lock_decimation` | Decimation's damage part | Demonology | 3% a point | 0 | 3% a point | = Expected | −1.0% (Low, PR A/B −11.6) |

### `lock_pet_scaling`: Warlock Pet Scaling coefficients

- **Builds:** Demonology, Destruction, Affliction, Destruction Fire (the Succubus is out for every build).
- **Sim now:** the demon gets:
  - 0.5 attack power and 0.25 spell damage per point of the warlock's highest-school spell power;
  - 100% of the warlock's spell hit as spell hit and 80% as melee hit;
  - no crit, Stamina or Intellect.
  
  `sim/warlock/pet.go:216-227` (values at `:222`, `:223`, `:225`).
- **Source:**
  - Blizzard's warlock deep dive (2026-10-07: "pets now scale from their summoner's stats") and the client's Warlock Pet Scaling aura 416189, every amount 0 (server side).
  - The numbers are Season of Discovery's: the 2024-12-04 hotfix for hit, and a 2024-01-06 player test for AP and spell damage (the tester was unsure of the spell damage). Details in `orchestration/pet_buffs.md` section 5.
- **Why uncertain:** no Forever measurement exists; the beta's 09-17 known issues say the sheet doesn't show the coefficient. PR #2 calls the demon "the least sourced input of any spec".
- **Settings:**
  - **Low, none.** Forever demons inherit nothing measurable; a reading of the deep dive as health and armor only.
  - **Expected, SoD (Sim now).** Forever carried Season of Discovery's aura 416189 unchanged, so Season of Discovery's server values are the closest analogue.
  - **High, SoD plus all of the warlock's spell crit.** Forever's hunter pets take all of the hunter's crit (beta-measured, `pet_buffs.md` S7), and aura 416189 has a crit slot.
- **Sensitivity:**
  - Low (PR, first set, Gnome): Demonology 1087.2 → 1037.8 (−4.5%), Destruction 984.9 → 946.1 (−3.9%), Affliction 976.2 → 937.0 (−4.0%), Destruction Fire 976.2 → 937.3 (−4.0%).
  - Without the scaling, the Fire build would sacrifice the Succubus again (950.8).
  - High: est. +0.5 to +1%. The Succubus is about 10% of Demonology's damage, and its crit would rise by about 7 points.
- **Knob:** `sim/warlock/pet.go:222-225`. Scale each field by `Assume("lock_pet_scaling", 1)` (0 for Low); for High add `stats.SpellCrit` and `stats.MeleeCrit` from `ownerStats[stats.SpellCrit]`.

### `lock_demon_melee`: the Succubus's melee damage

- **Builds:** all four (the Succubus is out in each).
- **Sim now:** 1.12's level 60 base melee, 95 to 131 every 2.0 sec. `sim/warlock/succubus.go:93-95`.
- **Source:** 1.12 creature data. Forever halved the demons' spells in the client (Lash of Pain 50, 1.12 99; Firebolt 43 to 48, 1.12 85 to 96), but melee is server side.
- **Why uncertain:** nothing gives Forever's demon melee (client 70338 DB2, wow.gg's Demonology guide, ForeverChanges; notes.md). If Blizzard halved the melee along with the spells to make room for the new scaling, the Succubus is over-valued. The sim's Succubus also beats the Imp by 55 to 60 DPS, where 1.12 had them closer.
- **Settings:**
  - **Low, halved (47.5 to 65.5).** Matches the client's halving of every demon spell.
  - **Expected, 1.12 (Sim now).** No source shows a change, and server creature data usually carries over.
  - **High:** = Expected.
- **Sensitivity (PR, first set):** Low gives Demonology −43 (−4.0%) and the other builds −38 to −39 (−3.9 to −4.0%). Combined with no owner scaling: −76 / −62 to −63 (−7.0%, −6.4%).
- **Knob:** `sim/warlock/succubus.go:93-94`, `BaseDamageMin/Max *= Assume("lock_demon_melee", 1)`.

### `lock_bod_coef`: Bane of Doom's spell power coefficient

- **Builds:** Demonology, Destruction, Affliction (Destruction Fire doesn't cast it).
- **Sim now:** 1742 Shadow plus 4.0 × spell power after 60 sec. `sim/warlock/curses.go:290` (coefficient) and `:292` (base).
- **Source:** client spell 603, Forever's own value (Era's client is 1.0).
- **Why uncertain:** it hasn't changed in any of the 16 beta builds, and the spell is learned at 60, so no player has cast it. Forever's other DoTs got 1.2 to 1.6× their Era coefficients; this one got 4×. The rank tuning audit rates it "Unclear (Forever design, untested)".
- **Settings:**
  - **Low, 2.0.** The rank tuning auditor's test value, half the client's. 1.0 (Era's client) is the extreme: −6.1%.
  - **Expected, 4.0 (Sim now).** The client value, and exactly Blizzard's DoT rule of duration / 15 sec (60 / 15).
  - **High:** = Expected.
- **Sensitivity (PR, rank tuning audit):** Low −4.1% on every Shadow Bolt build (Demonology 1087 → 1043, Destruction 985 → 945, Affliction 976 → 935). That puts Destruction and Affliction below Retribution twist and sub Holy.
- **Knob:** `sim/warlock/curses.go:290`, `BonusCoefficient: core.Assume("lock_bod_coef", 4)`.

### `lock_incinerate_bonus`: what Incinerate's 25% with Immolate covers

- **Builds:** Destruction Fire (Incinerate is its filler, about 52 casts a fight).
- **Sim now:** with the warlock's Immolate up, the whole hit (base and spell power) does 25% more. `sim/warlock/incinerate.go:39-46`.
- **Source:** the tooltip; the client's second effect is a script holding 25.
- **Why uncertain:** the script's meaning is server side, and Incinerate is a level 40 talent, beyond the beta cap (notes.md).
- **Settings:**
  - **Low, 25% of the base damage only.** WotLK's Incinerate gave its Immolate bonus as a flat extra of about 25% of the base, which didn't scale with spell power.
  - **Expected, whole hit (Sim now).** The tooltip reads "25% increased damage", and the client stores a percent.
  - **High:** = Expected.
- **Sensitivity (est., at Low):** rank 3 is 217 base plus 0.714 × about 550 spell power, about 393. Losing a quarter of the spell power part is about −13% per Incinerate. Incinerate is about 31% of the Fire build's damage, so about −4%. The Fire build might then fall back to Shadow Bolt Destruction.
- **Knob:** `sim/warlock/incinerate.go:43-46`. For Low, compute a separate flat add of `0.25 * rolled base` instead of the multiplier.

### `lock_decimation`: Decimation's damage part

- **Builds:** Demonology (2 points).
- **Sim now:** Shadow Bolt and Searing Pain deal 3% more a point against a target under 35% health. `sim/warlock/talents.go:727`.
- **Source:** the tooltip only ("they deal 3% increased damage"); the client's effect has 0 points.
- **Why uncertain:** the server fills the value; a level 40 Demonology warlock and a mob under 35% would settle it (notes.md).
- **Settings:**
  - **Low, 0.** The client effect as stored.
  - **Expected, 3% a point (Sim now).** The tooltip and its trait curve (3 / 6).
  - **High:** = Expected.
- **Sensitivity (PR):** the talent verification A/B is −11.6 on Demonology (first set), −1.0%. The damage part is all of it, since the Soul Fire half isn't used.
- **Knob:** `sim/warlock/talents.go:727`, `0.03 → core.Assume("lock_decimation", 0.03)`.

---

## Paladin

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `pal_sor_sp_double` | Seal of Righteousness takes spell power twice | twist, sub Holy, Shock | twice (2 × 11% on a two-hander) | once (10%) | twice | = Expected | Low −5.1% twist, −4.4% Shock (PR) |
| `pal_consecration_coef` | Consecration's coefficient a tick | twist, sub Holy, Shock | 9.5% | 4.2% (Era) | 9.5% | 10.55% | est. Low −4%, High +0.7% |
| `pal_2h_spec_holy` | Two-Handed Weapon Spec on Holy seal damage | twist, sub Holy, Shock | yes | Physical only | yes | = Expected | Low −1.1% twist (PR) |
| `pal_31plus_ranks` | Holy Strike and Consecration 31+ ranks | twist, sub Holy, Shock | client | held to the ≤30 trend | client | = Expected | Low −1.2% (PR) |

Retribution is also affected by `flat_bonus_inside_pct` (Holy Strike; est. +3% at its Low) and `oil_proc_damage` (Shadow Oil; −0.07% at its Low).

### `pal_sor_sp_double`: Seal of Righteousness counts spell power twice

- **Builds:** Level 60 twist, sub Holy (both twist Seal of Righteousness in on every swing) and Shock (Seal of Righteousness stays up).
- **Sim now:**
  - The proc's coefficient is 10% under Forever, ×1.1 on a two-hander (`sim/paladin/sor.go:105-112`).
  - It is added once into `baseDamage` (`:131`) and once more through `BonusCoefficient` (`:127`), so the effective coefficient is 2 × 11%.
  - Judgement of Righteousness is 50% (`:94`).
- **Source:** Season of Discovery sim tests ("from testing in SoD", `:111`). The Forever client has the same data as Season of Discovery's: Judgement of Righteousness (20286) carries the seal's 1786 / 47 at a 20% coefficient, twice the proc spell's 10%.
- **Why uncertain:** the double count is a Season of Discovery server behaviour inferred from data, not a Forever measurement. A level 30 paladin with and without spell power would settle it.
- **Settings:**
  - **Low, once (10%, no two-hander factor).** The proc spell's own coefficient, read plainly.
  - **Expected, twice (Sim now).** Forever's client matches Season of Discovery's data, where the double count was measured.
  - **High:** = Expected.
- **Sensitivity (PR, results audit before the fairness round):** Low gives twist −49 (−5.1%) and Shock −42 (−4.4%); sub Holy is like twist. The build order doesn't change.
- **Knob:** `sim/paladin/sor.go:131`. When the key is 1, drop the explicit `spell.BonusCoefficient*(...)` term; `:111` drops the ×1.1.

### `pal_consecration_coef`: Consecration's spell power coefficient

- **Builds:** twist and sub Holy (Consecration down to 20% mana, worth 109 DPS) and Shock (to the last point of mana).
- **Sim now:** 9.5% of spell power a tick, 8 ticks, on the whole tick (the "everyone" quarter and the first-four extra together). `sim/paladin/consecration.go:47` (Forever) against `:44` (Era 0.042).
- **Source:** "an older Forever tooltip" (`:13-15`). The damage runs through a server script, and no page shows the coefficient (sim-status, 2026-09-24 notes).
- **Why uncertain:**
  - The script's coefficient is unpublished.
  - Rank 5's base is only 39 a tick, so at about 700 spell power the coefficient is about two thirds of each tick.
  - The rank tuning audit also flags Consecration's ranks.
- **Settings:**
  - **Low, 4.2% a tick.** Era's 33% over 8 ticks, if the old tooltip's 9.5% was a leftover.
  - **Expected, 9.5% a tick (Sim now).** The only Forever number anyone has seen.
  - **High, 10.55% a tick.** 9.5% on the first-four extra plus a quarter of Era's 4.2% on the "everyone" part, if the two parts scale separately.
- **Sensitivity (est.):**
  - At Low, each tick goes from about 105 to 68 (−35%). About −38 DPS on twist (−3.9%) and a similar −4% on Shock.
  - At High, about +0.7%.
- **Knob:** `sim/paladin/consecration.go:47`, `tickCoefficient = core.Assume("pal_consecration_coef", 0.095)`.

### `pal_2h_spec_holy`: Two-Handed Weapon Specialization on Holy seal damage

- **Builds:** all three (the synthetic axe is a two-hander, 3/3 in every build).
- **Sim now:** 2 / 4 / 6% on white hits, Holy Strike, the Seal of Righteousness and Seal of Command procs, and Judgement of Command. `sim/paladin/talents.go:228-233`, used as `DamageMultiplier` in `sor.go` and `soc.go`.
- **Source:** the Era sim's choice. Client 20111 is aura 79 misc 1 (Physical), the same as Era.
- **Why uncertain:** the client mask says Physical, which would leave out the Holy seal damage; the Era sim applies it because the seals roll as melee (PR #3, "open assumption").
- **Settings:**
  - **Low, Physical only.** The client mask taken at its word.
  - **Expected, yes (Sim now).** The Era servers' behaviour on the identical aura, as the Era sim models it.
  - **High:** = Expected.
- **Sensitivity (PR):** Low gives twist about −11 (−1.1%). Sub Holy is similar; Shock was not measured.
- **Knob:** `sim/paladin/talents.go:233`. Return 1 for Holy-school callers when the key is 0 (pass the school, or use a separate `getWeaponSpecializationModifierHoly()` in `sor.go` / `soc.go`).

### `pal_31plus_ranks`: Holy Strike's flat part and Consecration's 31+ ranks

- **Builds:** all three.
- **Sim now:** client values: Holy Strike rank 8 flat 93 (+3.2 a level), Consecration rank 5 312. `sim/paladin/holy_strike.go:48-51` and `sim/paladin/consecration.go:27-30`.
- **Source:** client 1.60.1.70338 and wowhead Forever.
- **Why uncertain** (rank tuning audit):
  - Holy Strike's flat part grows 0.1 to 0.3 a level up to 30, then 1.0 to 3.2 a level, a jump right past the beta cap.
  - Consecration goes from 0.67× Era at 30 to 0.81× at 60, with Era mana costs.
  - Both look stale-shaped, generous at 60.
- **Settings:**
  - **Low, held to the ≤30 trend.** Holy Strike flat ×0.52, Consecration 31+ ranks ×0.82.
  - **Expected, client (Sim now).** No 31+ rank changed in any beta build; Blizzard's later passes typically moved the top rank only 0 to 3%.
  - **High:** = Expected.
- **Sensitivity (PR):** Low gives −1.2% on every build (twist 962 → 951, sub Holy 957 → 946, Shock 936 → 925 on the audit's set), split −0.4% Holy Strike and −0.8% Consecration.
- **Knob:** scale `holy` and `scale` for ranks 5 to 8 (`holy_strike.go:48-51`) and `foreverAll` / `foreverExtra` for ranks 3 to 5 (`consecration.go:27-30`) by the key.

---

## Warrior

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `warr_rage_k` | How normalised Rage grows with level | Fury, Fury two-hand, Arms | midpoint, k ×1.5 at 60 (standards 21) | flat k (×1) | midpoint (×1.5 at 60) | Classic-like (×2 at 60) | flat: Fury 950.6 at 180 s; others to measure |
| `warr_rage_oh` | Off-hand factor of normalised Rage | Fury | ×1.0 | ×0.5 | ×1.0 | = Expected | Low −9.4% Fury (measured with flat k) |
| `warr_crit_rage` | Extra Rage from critical white hits | Fury, Fury two-hand, Arms | +100% | +75% | +100% | = Expected | est. Low −2 to −3% Fury, about −1.5% Arms |

Dual-wield Fury is also affected by `wf_totem_offhand` (est. −2 to −3% at its Low).

### `warr_rage_k`: how normalised Rage grows with level

- **Builds:** Fury, Fury two-hand, Arms.
- **Sim now (standards 21):** a landed, dodged or parried white hit gives k(L) × weapon swing speed in Rage. k30 is 3.46 for a one-hander and 4.5 for a two-hander, and k(L) = k30 × (1 + s × (L − 30) / 30) above level 30, with s = 0.5 (Expected). Dual Wield Specialization and the `warr_crit_rage` crit bonus multiply on top. The code is `sim/core/rage.go`, behind `Env.IsForever()`; only the warrior enables it.
- **Source:** Blizzard's Priest & Warrior deep dive ("fully-normalized ... based on the speed of the weapon that dealt the damage"), tier 3. The constants are beta testers' measurements at level 30 or below (forum thread 2355684; Icy Veins and wowforeverbuilds repeat them), tier 5.
- **Why uncertain:** the beta stopped at level 30, and Blizzard gives no formula. At 30 the measured k gives about Classic's Rage per swing; flat at 60 it gives about half (a 2.6 s main hand: 9.0 against Classic's 18.6 for an average swing). Nobody knows whether Blizzard scales it with level. The full test plan is under "Rage needs a beta test" in orchestration/followups.md.
- **Settings (user ruling 2026-10-13):**
  - **Low, flat (s = 0).** The level 30 constants unchanged at 60.
  - **Expected, midpoint (s = 0.5).** ×1.5 at 60.
  - **High, Classic-like (s = 1).** ×2 at 60, close to Classic's Rage per average swing.
- **Sensitivity:** with flat k, Fury is 950.6 at 180 s (it was 1055.0 under Classic's damage-based Rage), Fury two-hand 835.0 and Arms 861.4. The warrior re-measures all three settings.
- **Knob:** a package variable for the level slope s, next to `core.ForeverNormalizedRageK` (which multiplies k). The dashboard sets it with `DASH_ASSUME`.

### `warr_rage_oh`: off-hand factor of normalised Rage

- **Builds:** Fury (dual wield).
- **Sim now:** ×1.0 before Dual Wield Specialization (×1.5 at 5/5).
- **Source:** none for the off hand. Blizzard's post only says Rage depends on the weapon's speed.
- **Settings:**
  - **Low, ×0.5.** TBC's normalised formula, where the off hand gets half the main hand's factor (1.75 against 3.5).
  - **Expected, ×1.0.**
  - **High:** = Expected.
- **Sensitivity:** Low is −9.4% on Fury (861.7 against 950.6 at 180 s, measured with flat k).
- **Knob:** `core.ForeverNormalizedRageOffHand`.

### `warr_crit_rage`: extra Rage from critical white hits

- **Builds:** Fury, Fury two-hand, Arms.
- **Sim now:** a critical white hit (main or off hand) gives twice its normal Rage, on top of the doubled damage, so 4× a normal hit's Rage where Classic gives 2×. Abilities get nothing extra. `sim/warrior/warrior.go:171-175` (`AddCritDealtRageMultiplier(2)`), hook `sim/core/rage.go:160`.
- **Source:** Blizzard's 1 October warrior notes: "75% increased Rage when landing a critical strike with a basic attack", then 100% in the second warrior pass the same day.
- **Why uncertain:** no beta combat log has checked it. A few dozen white swings at level 30 would show it (notes.md, Warrior).
- **Settings:**
  - **Low, +75%.** The first 1 October number.
  - **Expected, +100% (Sim now).** The later pass of the same day.
  - **High:** = Expected. Heroic Strike and Cleave replace the swing and make no Rage, so a wider scope adds nothing.
- **Sensitivity (est.):**
  - At Low, white-hit Rage drops about 8% (crits at about 25% of landed swings, glancing 40%), about 6% of all Rage.
  - Fury is Rage-starved (Bloodthirst slips GCDs), so −2 to −3% on Fury and two-hand Fury. Arms has more non-white Rage (Bloodthrill, Overpower), so about −1.5%.
- **Knob:** `sim/warrior/warrior.go:174`, `AddCritDealtRageMultiplier(core.Assume("warr_crit_rage", 2))` (1.75 for Low).

---

## Hunter

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `hunter_pet_scaling` | Pet scaling from the hunter | BM, MM, SV | 10% RAP + 1:1 hit | 10% RAP, no hit | 10% RAP + hit | SoD: 22% RAP + hit | Low −17 / −10 / −10, High +41 / +23 / +18 (PR) |
| `hunter_pet_speed` | Pet hits normalized to a 2.0 sec pet | BM, MM | yes (1.0 sec cat) | Classic: hit scales with speed | yes | = Expected | est. Low −5 to −8% BM, −2 to −4% MM |
| `hunter_tsa_hm_rank` | Trueshot Aura and Hunter's Mark at 60 | LW, MM, BM | 75 / 98 (strongest learned rank) | = Expected | 75 / 98 | 100 / 140 | est. Expected +1.2%, High +3% (from PR's TSA +6.4 LW) |
| `hunter_shot_ap_shares` | Arcane Shot and Serpent Sting AP shares | LW, MM, BM | 0.11 / 0.035 a tick | 0.10 / 0.030 | 0.11 / 0.035 | 0.12 / 0.040 | est. ±1.3% |
| `hunter_crit_per_agi` | Crit per Agility at 60 | all four | Classic table (53 Agi a 1%) | Classic | ×1.22 | = Expected | est. +2% (Expected) |
| `hunter_raptor_31plus` | Raptor Strike 31+ flat bonuses | Survival | client (70 at rank 8) | 70 | 96 | 123 | High +1.2% (PR) |

Survival is also affected by `wf_totem_offhand` (est. about −1.5% at its Low).

### `hunter_pet_scaling`: Hunter Pet Scaling (attack power and hit)

- **Builds:** Beast Mastery (pet 26% of the damage), Marksmanship (14%), Survival. Lone Wolf has no pet.
- **Sim now:** the pet takes 10% of the hunter's ranged attack power, all of the hunter's melee crit and its melee hit 1:1. `sim/hunter/pet.go:226-237` (values at `:232-234`).
- **Source:**
  - Attack power and crit: beta measurements (ForeverChanges' hunter pets page, the forever-hunter wiki).
  - Hit: Season of Discovery's Hunter Pet Scaling 415429 as a placeholder. Blizzard's 09-17 known issues say pet hit takes "the master's scaling coefficient" without giving it.
- **Why uncertain:** the 10% was measured at levels 20 to 30, and the hit share is unmeasured.
- **Settings:**
  - **Low, 10% RAP, no hit inheritance.** Only what the beta measured.
  - **Expected, 10% RAP plus 1:1 hit (Sim now).** The measured attack power, plus the hit slot the known issue confirms exists.
  - **High, Season of Discovery's full scaling.** 22% of RAP plus hit, the same aura 415429 at Season of Discovery's server values.
- **Sensitivity (PR, 2026-10-12 gear):** Low −17 BM (−1.8%), −10 MM (−1.0%), −10 SV (−1.2%). High +41 BM (+4.2%), +23 MM (+2.3%), +18 SV (+2.1%).
- **Knob:** `sim/hunter/pet.go:232` coefficient `Assume("hunter_pet_ap", 0.1)` and `:234` factor `Assume("hunter_pet_hit", 1)`. The setting sets both.

### `hunter_pet_speed`: a fast pet hits as hard as a 2.0 sec pet

- **Builds:** Beast Mastery (a 1.0 sec cat) and Marksmanship (a cat; check that its speed option is the same).
- **Sim now:** every pet's auto attack hits the size of a 2.0 sec pet's (18.17 to 27.66 × 2) at any attack speed, so the fastest pet deals the most. `sim/hunter/pet.go:142-149` (`applyForeverRules`; the override at `:148-149`).
- **Source:** beta players, gathered on ForeverChanges' hunter pets page and the forever-hunter wiki. The code comment says "nobody knows yet whether that is intended".
- **Why uncertain:** it could be a beta bug that Blizzard fixes, which would bring back Classic's hit size scaling with the swing.
- **Settings:**
  - **Low, Classic.** The hit scales with attack speed (`sim/hunter/pet.go:64-65`), so every pet has the same white DPS.
  - **Expected, as on the beta (Sim now).** The only observed behaviour.
  - **High:** = Expected.
- **Sensitivity (est.):** at Low the 1.0 sec cat's white damage halves. White hits are about half of the pet's damage, so about −5 to −8% on Beast Mastery and −2 to −4% on Marksmanship. Not simmed; Frenzy's haste makes the share uncertain.
- **Knob:** `sim/hunter/pet.go:148-149`. Skip the override when the key is 0.

### `hunter_tsa_hm_rank`: Trueshot Aura and Hunter's Mark at level 60

- **Builds:** Lone Wolf, Marksmanship, Beast Mastery. Both are ranged attack power only, so Survival is unaffected.
- **Sim now:** the strongest learned rank (standards 8a, `bestLearnedRank` in sim/core on the hunter branch): Trueshot Aura rank 4 +75 RAP and Hunter's Mark rank 3 +98 RAP.
- **Source:**
  - Client 70338: Trueshot Aura ranks 30 / 40 / 50 / 75 / 50, Hunter's Mark 26 / 59 / 98 / 71. Both top ranks sit below the rank before them.
  - Standards 8a: use the strongest learned rank. The hunter branch implements it (PR #5), worth +10.7 to +14.3 DPS.
- **Why uncertain:** a top rank weaker than the one before is either a client data error Blizzard fixes, or a value players work around by casting the lower rank.
- **Settings:**
  - **Low:** = Expected. A player can always cast the stronger lower rank, so the weaker top ranks are never worse.
  - **Expected, 75 / 98 (Sim now).** The strongest learned rank.
  - **High, 100 / 140.** Blizzard fixes the top ranks onto the curve of the ranks before them: Trueshot Aura +25 a rank; Hunter's Mark about +40 a rank, Forever's ranks 1 to 3 running about 1.3× Classic's 20 / 45 / 75 / 110.
- **Sensitivity:**
  - PR: Trueshot Aura rank 4 alone is +6.4 LW, +5.6 MM, +5.0 BM (2026-10-12 gear).
  - est.: Hunter's Mark's +27 adds about as much, so Expected is about +1.2% and High about +3%.
- **Knob:** the level 60 entries of `foreverTrueshotAuraRAP` and `foreverHuntersMark` in sim/core (buffs.go / debuffs.go). For High, set them to `Assume("hunter_tsa60", 50)` / `Assume("hunter_hm60", 71)` with 100 / 140; `bestLearnedRank` then picks them.

### `hunter_shot_ap_shares`: Arcane Shot and Serpent Sting attack power shares

- **Builds:** Lone Wolf, Marksmanship, Beast Mastery. Arcane Shot is 13 to 18% and Serpent Sting 9 to 10% of their damage.
- **Sim now:** Arcane Shot 0.11 of RAP (`sim/hunter/arcane_shot.go:16`); Serpent Sting 0.035 of RAP a tick, read live (`sim/hunter/serpent_sting.go:17`).
- **Source:** fits to level 20 to 30 beta combat logs (ElliotWood/Forever PR #489). The client gives neither spell an attack power coefficient. The forever-hunter wiki says 10% and 3% a tick.
- **Why uncertain:** they were fitted below level 30 with small attack power; a level 60 log has not checked them.
- **Settings:**
  - **Low, 0.10 / 0.030.** The forever-hunter wiki's figures.
  - **Expected, 0.11 / 0.035 (Sim now).** The combat-log fit.
  - **High, 0.12 / 0.040.** The fit's distance from the wiki, mirrored. The data constrain the shares only to about ±10%.
- **Sensitivity (est.):** at about 2100 RAP the attack power part is about half of each spell. Low gives Arcane Shot −4.5% and Serpent Sting −6.7%, about −1.3% on each build; High gives about +1.3%.
- **Knob:** the two constants at `arcane_shot.go:16` and `serpent_sting.go:17`.

### `hunter_crit_per_agi`: the hunter's crit per point of Agility

- **Builds:** all four. Lone Wolf, Marksmanship and Beast Mastery stack Agility; the pets inherit the hunter's melee crit.
- **Sim now:** Classic's table, 0.0189% a point at 60 (53 Agility a 1%), on Classic's level curve. `sim/core/base_stats.go:176` (`CritPerAgiAtLevel`), curve at `:396-398`.
- **Source:** Classic base stats.
- **Why uncertain:** at level 1 the beta's hunter gets 0.2174 crit a point where the sim has 0.1783 (dodge 0.4348 against 0.3567), 22% more (notes.md, All classes, 2026-10-10). The paladin, rogue, priest, mage and druid match the sim at level 1, so Forever changed the hunter's table; how that carries to 60 is unknown.
- **Settings:**
  - **Low, Classic (Sim now).** The level 1 difference stays a level 1 quirk.
  - **Expected, ×1.22 at every level.** The beta's measured level 1 ratio, carried to 60.
  - **High:** = Expected; no data above level 1.
- **Sensitivity (est.):** about 500 Agility gives 9.4% crit at Classic's rate and 11.5% at ×1.22. +2.1% crit is about +2% on the ranged specs (Mortal Shots' 2.3× crits) and slightly less on Survival.
- **Knob:** `sim/core/base_stats.go:398`. Multiply by `Assume("hunter_crit_per_agi", 1)` when `c == proto.Class_ClassHunter`. The warrior's level 1 difference is −3% and its Agility is small, so it isn't worth a knob.

### `hunter_raptor_31plus`: Raptor Strike's level 31+ flat bonuses

- **Builds:** Survival (Raptor Strike 11% of its damage).
- **Sim now:** client flat bonuses 35 / 40 / 55 / 70 at ranks 5 to 8. `sim/hunter/raptor_strike.go:15`.
- **Source:** client 70338.
- **Why uncertain:** level 24's flat is 0.88 of Era, but the 32 to 56 ranks sit at 0.50 to 0.70, so the playable ratio isn't kept (rank tuning audit, "stale-shaped").
- **Settings:**
  - **Low, 70 (Sim now).** The client.
  - **Expected, 96.** Halfway: a partial retune of the kind Blizzard has made to the ≤30 ranks.
  - **High, 123.** Held at 0.88 of Era's 140, ×1.76.
- **Sensitivity:** High +1.2% (PR, 761 → 770); Expected about +0.6% (est.).
- **Knob:** `sim/hunter/raptor_strike.go:15`. Rank 8 entry `Assume("hunter_raptor8", 70)`; the preset casts rank 8 only.

---

## Mage

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `mage_ice_lance_coef` | Ice Lance spell power coefficient | Frostfire, Frost | 0 (not cast) | 0 | 0.143 | 0.429 | High +18% Frostfire, +22% Frost (PR, earlier gear); est. Expected +3 to +6% |
| `mage_ab_on_missiles` | Arcane Blast stacks on Arcane Missiles | Arcane | 10% a stack | 0 | 10% | 15% | est. Low −8%, High +4% |
| `mage_frostbolt10` | Frostbolt X gets 70291's raise | Frost | client (×1.00) | ×1.00 | ×1.04 | ×1.08 | High +2.8% (PR) |

### `mage_ice_lance_coef`: Ice Lance's spell power coefficient

- **Builds:** Frostfire and Frost. Fire and Arcane don't take Ice Lance.
- **Sim now:** 0 on all six ranks (`sim/mage/ice_lance.go:26`), with ×4 against Frozen targets (`:22-23`). Casting it on Fingers of Frost charges loses, so no rotation casts it.
- **Source:** client 70338. Every other new Forever damage spell carries a client coefficient (Frostfire Bolt 0.814, Arcane Blast 0.714).
- **Why uncertain:**
  - Season of Discovery's client also had 0, and a server script gave it 42.9%.
  - Forever's rank 2 is Season of Discovery's rune spell (400640).
  - Forever also lowered the Frozen bonus (client 300 where Season of Discovery's was 400), so Blizzard did touch the spell.
  - The talent is reachable on the beta (20 Frost points), but no test has been posted.
- **Settings:**
  - **Low, 0 (Sim now).** The client value.
  - **Expected, 0.143.** WotLK's Ice Lance coefficient (1/7). It is the middle ground between leaving the client's 0 and copying Season of Discovery's script onto a spell Forever already retuned. This is a judgement call.
  - **High, 0.429.** Season of Discovery's server value on the same spell, the standard instant-cast rule.
- **Sensitivity:**
  - High: about +165 DPS on Frostfire and Frost (PR, on the earlier 526 spell power set: +18% / +22%). Frost would move ahead of Fire.
  - Expected: est. +3 to +6%. Ice Lance on Fingers of Frost charges probably wins only narrowly at a ×4 Frozen multiple.
- **Knob:**
  - `sim/mage/ice_lance.go:26`, `IceLanceSpellCoeff` → `core.Assume(...)`.
  - The rotation must also cast it: add an Ice Lance-on-Fingers-of-Frost line behind a knob in `tools/rotopt` (`mage_frost` / `mage_frostfire` templates), and re-run the knob search for each setting.

### `mage_ab_on_missiles`: Arcane Blast's stack bonus on Arcane Missiles

- **Builds:** Arcane (Arcane Missiles 46% of its damage, cast after 2 or 4 Blast stacks).
- **Sim now:** 10% a stack, snapshotted at the start of the channel. `sim/mage/arcane_blast.go:26` (`ArcaneBlastStackBonus`), read by `arcaneBlastBonus()` at `:116-121` and `sim/mage/arcane_missiles.go:106`.
- **Source:** the tooltip ("all your other spells").
- **Why uncertain:** in the client, the stack buff 400573 leaves Arcane Missiles out of its damage part. A second buff, 400586 (15% a stack on Arcane Missiles only, Season of Discovery's), is still in the client. A missile after one Arcane Blast settles it on the beta (level 20+).
- **Settings:**
  - **Low, 0.** 400573's data, Missiles excluded.
  - **Expected, 10% a stack (Sim now).** The Forever tooltip.
  - **High, 15% a stack.** 400586, Season of Discovery's Missiles-specific buff, still in the client.
- **Sensitivity (est.):** with about 3 stacks a Missiles cast, the attacker multiplier moves from about 1.63 to 1.33 at Low (−18% Missiles, −8% Arcane) and to 1.78 at High (+9% Missiles, +4% Arcane). The best stack count may change, so re-run the Arcane knob search for each setting.
- **Knob:** split the constant. Add `arcaneMissilesStackBonus := core.Assume("mage_ab_on_missiles", 0.10)` and use it in `arcaneBlastBonus()` (`arcane_blast.go:120`), which only Missiles reads (`arcane_missiles.go:106`).

### `mage_frostbolt10`: does build 70291's Frostbolt raise reach rank 10?

- **Builds:** Frost (Frostbolt 98% of its damage).
- **Sim now:** client values, rank 10 382 to 413 at 60. `sim/mage/frostbolt.go:16` (rank 10 entry) and `:17`.
- **Source:** client 70338.
- **Why uncertain:** 70291 raised the level 20 and 26 Frostbolt ranks (+8% at 26's cap) and left 31+ alone. The 31+ ratios to Era already sit above the playable one, so it's unclear whether the raise should carry (rank tuning audit, "mild lag").
- **Settings:**
  - **Low, ×1.00 (Sim now).**
  - **Expected, ×1.04.** Halfway; the evidence points both ways.
  - **High, ×1.08.** 70291's raise applied to rank 10.
- **Sensitivity:** High +2.8% (PR, 744 → 764); Expected about +1.4% (est.).
- **Knob:** scale rank 10's `FrostboltBaseDamage` (`sim/mage/frostbolt.go:16`) where the range is built.

---

## Shaman

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `sham_mw_ppm` | Maelstrom Weapon procs a minute (5/5) | all four Enhancement builds | 10 PPM | 5 | 10 | 20 | est. Low −3%, High +5% |
| `sham_lb_31plus` | Lightning Bolt 31+ ranks | Elemental, Elemental Mana Tide | client (×1.00) | ×1.00 | ×1.12 | ×1.38 | Expected +1.8%, High +5.8% (PR) |
| `sham_es7_tuning` | Earth Shock VII tuning | all four Enhancement builds | client (×1.00) | ×1.00 | ×1.08 | ×1.16 | High +0.8% (PR) |

### `sham_mw_ppm`: Maelstrom Weapon's proc rate

- **Builds:** Enhancement Level 60, Imp GW, Water Shield, Mana Tide (all 5/5). Lightning Bolt goes out only at 5 stacks and is about 7% of the damage.
- **Sim now:** 2 procs a minute a point, 10 PPM at 5/5, on any landed melee hit. `sim/shaman/talents.go:497-499` (TODO: "The tooltip has no proc rate, beta testing will have to measure it").
- **Source:** our guess (shaman_audit.md 2.7: "2 a minute a point, a guess"). The client leaves the rate to the server.
- **Why uncertain:** the talent needs level 34+, and no beta character can test it.
- **Settings:**
  - **Low, 5 PPM.** Half the analogues, for a Forever that spreads a weaker proc over its five ranks (1 a point). This is the weakest-grounded value in the inventory.
  - **Expected, 10 PPM (Sim now).** Season of Discovery's rune base rate (wowsims/sod `sim/shaman/runes.go`: "Base PPM is 10") and the 2-a-rank rate of the talent this one copies.
  - **High, 20 PPM.** Season of Discovery's rate with Windfury Weapon on a two-hander (10 + 5 + 5; the 2024-02-20 hotfix, warcraft.wiki.gg), which is exactly what the preset wields.
- **Sensitivity (est.):** Lightning Bolt casts scale with the rate, and Enhancement has spare GCDs between Stormstrike and the shocks. Halving costs about −3% and doubling adds about +5%.
- **Knob:** `sim/shaman/talents.go:499`. Change `2*float64(...)` to `Assume("sham_mw_ppm", 10)/5*float64(...)`.

### `sham_lb_31plus`: Lightning Bolt's level 31+ ranks

- **Builds:** Elemental (Lightning Bolt about 65% of its damage with overloads), Elemental Mana Tide.
- **Sim now:** client values, rank 10 189.92 to 211.68 plus 1.2 a level to 60. `sim/shaman/lightning_bolt.go:23` (rank 10 entry) and `:24`.
- **Source:** client 70338 (the ForeverChanges downrank calculator reads the same).
- **Why uncertain** (rank tuning audit, "stale-shaped"):
  - Lightning Bolt is 0.61× Era at levels 26, 32 and 38, then 0.54 / 0.48 / 0.44 at 44 / 50 / 56.
  - At 60 it does 78 base a cast-second against 84 to 158 for the other classes' nukes.
  - Build 70291 raised the top playable rank 12% and no 31+ rank.
- **Settings:**
  - **Low, ×1.00 (Sim now).** The client.
  - **Expected, ×1.12.** 70291's raise on the same spell, applied to rank 10.
  - **High, ×1.38.** Held at the 0.61 ratio of levels 26 to 38.
- **Sensitivity (PR):** Expected +1.8%, High +5.8% (690 → 730 on an earlier set); with Lava Burst ×1.38 as well, +7.2%. Elemental stays in the bottom third either way.
- **Knob:** scale rank 10's `LightningBoltBaseDamage` (`lightning_bolt.go:23`) where the range is built (`:51-52`). Optionally do the same to Lava Burst rank 3 (`lava_burst.go:16`), since it inherits the flat curve.

### `sham_es7_tuning`: Earth Shock VII

- **Builds:** all four Enhancement builds (Earth Shock about 10% of the damage).
- **Sim now:** client values, rank 7 293.07 to 308.93. `sim/shaman/earth_shock.go:16` (rank 7) and `:17`.
- **Source:** client 70338.
- **Why uncertain:** 0.66× Era at level 24, then 0.57 / 0.55 / 0.57 from 36 to 60 (rank tuning audit, "stale-shaped, mild").
- **Settings:**
  - **Low, ×1.00 (Sim now).**
  - **Expected, ×1.08.** Halfway.
  - **High, ×1.16.** Held at 0.66.
- **Sensitivity:** High +0.8% (PR); Expected about +0.4% (est.).
- **Knob:** scale rank 7's range at `earth_shock.go:37-38`.

---

## Druid

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `druid_ooc_rate` | Omen of Clarity proc rate | Feral, Balance, Balance Wrath | 2 PPM; spells 2 × cast / 60 | spells 6% a cast | 2 PPM | 3.5 PPM | est. Feral High +1.5 to +2%; Balance under 1% |
| `druid_shred_flat` | Shred's flat bonus | Feral | 80 inside the 155% (+124) | +80 | +124 | +180 | est. Low −3.5%, High +4.5% |
| `druid_rip_31plus` | Rip's level 31+ ranks | Feral | client | client | Era | ≤30 trend ×1.13 | Expected +1.6%, High +2.1% (PR) |
| `druid_wrath8_tuning` | Wrath VIII tuning | Balance Wrath, Balance | client (×1.00) | ×1.00 | ×1.085 | ×1.17 | High +1.6% Wrath build, +0.4% Balance (PR) |

### `druid_ooc_rate`: Omen of Clarity's proc chance

- **Builds:** Feral (each Clearcasting is a free Shred), Balance and Balance Wrath (free Starfires; the Wrath build casts Starfire only on Clearcasting).
- **Sim now:**
  - Melee: 2 PPM (`sim/druid/talents.go:424`).
  - Spells: 2 × max(cast time, 1.5 sec) / 60 a cast (`:445`).
  - Moonkin Form doubles the chance and halves the 10 sec cooldown (`:446-448`, `:452-455`).
- **Source:**
  - Melee: 1.12's rate.
  - Spells: our own guess (`:369-371`).
  - The Moonkin modifiers and the cooldown are the client's (16864).
- **Why uncertain:** the client leaves the chance to the server (notes.md, Druid).
- **Settings:**
  - **Low, a flat 6% a spell cast** (12% in Moonkin Form), melee 2 PPM. WotLK's spell rule. It cuts Starfire's chance from 20% to 12% in Moonkin Form and leaves Wrath about the same.
  - **Expected, the sim's rate (Sim now).** 1.12's 2 PPM, the version Forever's baseline passive was ported from.
  - **High, 3.5 PPM melee, and spells at 3.5 × cast / 60.** The rate later expansions gave Omen of Clarity.
- **Sensitivity (est.):** Feral's 1.0 sec paw gets about 1.6 procs a minute now, 2.8 at High: about +12 DPS (+1.5 to +2%). Balance is mostly mana-neutral (under 1% either way). The Wrath build moves with its Clearcasting Starfires (under ±1.5%).
- **Knob:** `sim/druid/talents.go:424` (PPM) and `:445` (the spell formula). Low needs a branch: `chance = 0.06` when the key says so.

### `druid_shred_flat`: Shred's flat bonus

- **Builds:** Feral (Shred 36% of the damage).
- **Sim now:** (weapon damage + 80) × 1.55, so 124 extra at rank 5. `sim/druid/shred.go:14` (`ShredFlatDamage`), `:17` (1.55), applied at `:57` and `:71`.
- **Source:** the client's effects (80, and 155% weapon damage), read with the same add-before-percent rule as Backstab (`flat_bonus_inside_pct`).
- **Why uncertain:** the tooltip still says "plus 180", 1.12's 80 × 2.25, so either the tooltip is stale or the server keeps 180. A few Shreds against a target with known armor settle it on the beta (notes.md, Druid).
- **Settings:**
  - **Low, +80 after the 155%.** The client numbers read as tooltip-style, the Low reading of `flat_bonus_inside_pct`.
  - **Expected, +124 (Sim now).** The client's effect order.
  - **High, +180.** The Forever tooltip, if the server kept 1.12's flat amount.
- **Sensitivity (est.):** at about 2000 cat attack power a Shred's base is about 431. Low is −44 (−10% Shred, about −3.5% Feral); High is +56 (+13% Shred, about +4.5% Feral).
- **Knob:** `sim/druid/shred.go:57,71`. Add `Assume("druid_shred_extra", 124) - 80*1.55` after the multiplier (DamageMultiplier is 1.55, so divide that term by 1.55 when adding it inside).

### `druid_tf_scope`: removed

Tiger's Fury isn't in Forever (Blizzard's Druid deep dive; nothing in client 70338 grants it), and the druid branch removes it. So it's no longer an assumption.

### `druid_rip_31plus`: Rip's level 31+ ranks

- **Builds:** Feral (Rip 22 to 25% of the damage, at 4 combo points).
- **Sim now:** client per-combo-point values 8.5 / 12.7 / 18.2 / 25.5 at ranks 3 to 6. `sim/druid/rip.go:22-25`.
- **Source:** client 70338.
- **Why uncertain:** Forever's ranks up to level 30 sit above Era per combo point (4.4 / 7.2 against 4 / 7), but the 31+ ranks sit 6 to 9% below (rank tuning audit).
- **Settings:**
  - **Low, the client (Sim now).**
  - **Expected, Era's per-combo-point values** (9 / 14 / 20 / 28). Forever's own ≤30 ranks are at or above Era, so Era is the floor of that design.
  - **High, the ≤30 trend, ×1.13.**
- **Sensitivity (PR):** Expected +1.6%, High +2.1% (648 → 662 on an earlier set).
- **Knob:** scale `dmgTickPerCombo` (and `dmgTickBase`) of rank 6 at `rip.go:25` by the key.

### `druid_wrath8_tuning`: Wrath VIII

- **Builds:** Balance Wrath (Wrath 59% of the damage) and Balance (14%).
- **Sim now:** client values, rank 8 91.57 to 102.43. `sim/druid/wrath.go:17` (rank 8) and `:18`.
- **Source:** client 70338.
- **Why uncertain:** build 70009 multiplied every Wrath rank by 1.5, then 70291 retuned only ranks 1 to 5. Forever over Era is 0.42 at level 30 and then 0.36 / 0.34 / 0.36 (rank tuning audit).
- **Settings:**
  - **Low, ×1.00 (Sim now).**
  - **Expected, ×1.085.** Halfway.
  - **High, ×1.17.** Held at 0.42.
- **Sensitivity (PR):** High +1.6% on Balance Wrath (760 → 772 on an earlier set), +0.4% on Balance. Expected about half of that (est.).
- **Knob:** scale rank 8's range from `wrath.go:17`.

---

## Rogue

No assumptions of its own. The rogue's scenarios are the shared ones:
- **`flat_bonus_inside_pct`:** Backstab, Ambush and Mutilate. Est. −3.7% Combat Daggers, about −4% Seal Fate Daggers, +1.7% Mutilate and under 0.5% Subtlety at its Low.
- **`wf_totem_offhand`:** every build dual wields in the Windfury Totem group. Est. −2 to −3% at its Low.

---

## Settled, not a scenario

- **Windfury Totem's internal cooldown: 1.5 sec.**
  - User ruling 2026-10-13, standards 17; the sim already uses it (`sim/core/buffs.go:1827`).
  - The Forever client's 100 ms and Era's no cooldown are closed.
  - Its off-hand question stays open, as `wf_totem_offhand`.
- **Rage of the Farseer's cooldown: 3 min.** It was a guess in sim-status, but the client (1.60.1.70291) and Blizzard's 2026-10-07 deep dive both give 3 minutes (`sim/shaman/talents.go:553-555`).
- **Lava Burst's cast time and cooldown.** These are the client's 2.5 sec and 10 sec now, not WotLK's (shaman_audit.md 7.14). Its Flame Shock 20% is below the threshold; see the next section.
- **One +2% crit elixir a character** (standards 10). A second one would add 10 to 36 a build (Enhancement +36, warrior +11 to +15, Ret +12 to +15, warlock +20, the casters +10 to +21). This is a ruling, not a sim constant. If the orchestrator wants it as a scenario, it's a consumes toggle per build, not a code knob.
- **No Ahn'Qiraj book ranks** (standards 8a). With them: Frost +50, Arcane +30, Fire +17, warlocks +7 to +13, rogues +13 to +18, Fury +22.6, Balance +13. A scenario would be `env.IncludeAQRanks()` flipped, if the orchestrator reopens it.
- **Pets and demons take no player buffs** (standards 14; client flag SPELL_ATTR3 0x100, forever-bugs #27).
- **Windfury Weapon rank 4: 333 attack power at 60.** wowhead's 433 is its level 68 value (`sim/shaman/windfury_weapon.go:16-22`).

## Considered and left out

These are below the 0.5% threshold, client-sourced, or have no sensible range.
- **Lava Burst's Flame Shock 20%: add or multiply with Call of Flame** (1.35 or 1.38; `sim/shaman/lava_burst.go:9`, `:67-69`).
  - This is one of the round's examples, but it is about 0.3%: Lava Burst is about 14% of Elemental's damage, and the difference is 2.2% of that.
  - Season of Discovery's sure crit is not a candidate: Forever's tooltip says 20% more damage.
- **Rip and Ferocious Bite attack power** (server side; `sim/druid/rip.go:81`, `ferocious_bite.go:112`): no sensible spread. The preset rips at 4 combo points, where 1.12's 24% and WotLK's 1% a point a tick (24% over 6 ticks) agree. Bite is 7% of the damage.
- **Ignite: no cap (Season of Discovery) or a stop after 5 crits (Era).** Five crits inside one 4 sec window is rare at the presets' crit, under 0.5% for Fire and Frostfire.
- **Premeditation out of Stealth** (Subtlety, −7.7 if it needed Stealth). Client 14183 dropped the stealth-only attribute, and ForeverChanges reads it the same way. Preparation's reset list is Classic's.
- **Deadly Poison re-snapshot on each application.** It measured +6.3 on an old Mutilate set, but what still feeds it on the final builds is unclear. Likely under 0.5%.
- **Percent talent modifiers: add or multiply** (rogue Improved Eviscerate × Aggression, Opportunity × Aggression; the warlock adds). +0.7 on Sinister Strike; under 0.5% everywhere.
- **Whirlwind's off-hand strike.** Blizzard's 1 October notes made it baseline; only its exact form (normalized off hand, 50%, Dual Wield Specialization) is server side, about 1% of Fury in total.
- **Weaponmaster's 200 ms internal cooldown, Deep Wounds' tick restart, Heroic Strike counting for Unbridled Wrath.** All small or unknowable without a log.
- **Dark Sacrifice's mana** (Undead priests). Shadow never runs out of mana; it only moves the Smite race table, where Undead leads the Gnome by 5.5.
- **Shadow Word: Death, Demonic Pact, Wrack, Incinerate's base, Sniper Shot, Light's Vigil, the other 31-point capstones.** Forever's own values, untested, with no analogue to give a range.
- **Arcane Shot and Aimed Shot cooldown sharing, Auto Shot through casts, Expose Prey, Lacerate, Lacerating Strikes.** Beta-tested or client data, or under the threshold.
- **Elemental Focus on low ranks, Searing Totem's reach from range, Elemental Devastation's size** (the client curve). All under 0.5%.
- **The encounter itself** (level 63, 3731 armor, Humanoid). This moves every spec, but it is an encounter setting, not a spec assumption.

## Where no sensible range was found

These entries have only two sensible settings (High = Expected): either the evidence gives two readings and nothing beyond them, or the higher value is the sim's own.
- `flat_bonus_inside_pct`, `wf_totem_offhand`, `priest_sf_tick_crit`, `pal_sor_sp_double`, `pal_2h_spec_holy`, `pal_31plus_ranks`, `warr_crit_rage`, `hunter_pet_speed`, `hunter_crit_per_agi`, `lock_demon_melee`, `lock_bod_coef`, `lock_incinerate_bonus`, `lock_decimation`, `oil_proc_damage` and `druid_tf_scope`.

Weakly grounded values:
- `sham_mw_ppm` Low (5 PPM): no analogue below 10.
- `hunter_shot_ap_shares` High: mirrored, not sourced.
- The halfway "Expected" of the rank-tuning entries (`priest_smite8_tuning`, `hunter_raptor_31plus`, `mage_frostbolt10`, `sham_es7_tuning`, `druid_wrath8_tuning`): a judgement between "client as is" and "held at the ≤30 ratio".
- `mage_ice_lance_coef` Expected (0.143): a judgement between the client's 0 and Season of Discovery's 0.429. Its High is the best-sourced alternative.

No sensible range at all: Rip and Ferocious Bite attack power, and the untested 31-point capstones (listed above).
