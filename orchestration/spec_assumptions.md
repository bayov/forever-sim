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
| All classes | 1 | `wf_totem_offhand` (`flat_bonus_inside_pct` settled 2026-10-16) |
| Priest | 2 | `priest_sf_tick_crit`, `priest_smite8_tuning` |
| Warlock | 4 | `lock_pet_scaling`, `lock_demon_melee`, `lock_bod_coef`, `lock_incinerate_bonus` (one value, kept for reference) |
| Paladin | 4 | `pal_2h_spec_holy`, `pal_31plus_ranks`, `pal_soc_wf`, `pal_soc_ppm` (pending) |
| Warrior | 3 (+1 shared) | `warr_rage_k`, `warr_rage_oh` (one value, 0.5: user ruling 2026-10-16), `warr_rend_ap` (+ `wf_totem_offhand`); `warr_crit_rage` settled 2026-10-16 |
| Hunter | 5 (+1 shared) | `hunter_pet_scaling`, `hunter_pet_speed`, `hunter_tsa_hm_rank`, `hunter_crit_per_agi`, `hunter_raptor_31plus` (+ `wf_totem_offhand` for Survival) |
| Mage | 3 (+1 pending) | `mage_ice_lance_coef`, `mage_frostbolt10`, `mage_imbue_scrolls` (+ `mage_cone_of_cold`, pending since 2026-10-23) |
| Shaman | 3 | `sham_mw_ppm`, `sham_lb_31plus`, `sham_es7_tuning` |
| Druid | 3 | `druid_ooc_rate` (one value, kept for reference), `druid_rip_31plus`, `druid_wrath8_tuning` |
| Rogue | 0 (+1 shared) | `wf_totem_offhand` |

That makes 35 assumptions in `assumptions_def.py`, 7 of them pending (no knob yet: `level_partial_resist`, `pal_soc_ppm`, `lock_pet_crit`, `lock_stone_with_oil`, `bleed_tick_crit`, `warr_dt_rage_mods`, `warr_dt_rage_k`) (PR #54 removed `pal_sor_sp_double`, `pal_sor_sp_speed` and `pal_consecration_coef`, and the assumptions audit, orchestration/assumptions_audit.md, removed `flat_bonus_inside_pct`, `druid_shred_flat` and `warr_crit_rage`, all settled 2026-10-16). The 2026-10-15 review removed `mage_ab_on_missiles`, `lock_decimation`, `oil_proc_damage` and `hunter_shot_ap_shares`, and added `pal_sor_sp_speed` (user rulings 2026-10-15). An assumption left with one value stays in the report as a reference row (user ruling 2026-10-15). The rogue has no material assumption of its own: the rank tuning audit found its spells Era-identical or client-tuned, and its open items are under 0.5% or client-sourced (see "Considered and left out"). Its two scenarios are the shared ones.

### The ones that matter most

1. **`mage_ice_lance_coef`.** Ice Lance's spell power coefficient is 0 in the client; Season of Discovery's client had 0 too, and its server gave the same spell 42.9%. The sim uses 42.9% and Frost and Frostfire cast Ice Lance on every Fingers of Frost charge (user ruling 2026-10-15, PR #21): Frost went from last among the mages to second. At half that, Frost loses 10.9% and Frostfire 7.3%; at the client's 0, where the better rotation never casts Ice Lance, Frost loses 19.1% and Frostfire 9.4% (assumption sims, 180 s).
2. **`lock_pet_scaling` + `lock_demon_melee`.** Both are Season of Discovery placeholders or 1.12 carry-overs. (Superseded 2026-10-23 by wowsims/forever's measured shares and melee, see their sections; the numbers below are the old model's.)
   - Without owner scaling, every warlock build loses about 4%; with the Succubus melee halved, another 4%; with both, 7% (PR).
   - Demonology leads the ranking (1109.8), and the other three builds sit just above Retribution (969 to 985 against 960 to 967). These decide that order.
3. **`lock_bod_coef`.** Bane of Doom's 4.0 coefficient is about 12% of every Shadow Bolt build's damage. At 2.0, every Shadow Bolt warlock loses 4.1% (PR), and Destruction and Affliction drop below the Retribution builds. I still expect 4.0 (it is the duration/15 rule), but it is untested in any beta build.
4. **Seal of Righteousness and Consecration (settled 2026-10-16, PR #54).** Blizzard's 15 October beta notes fixed Seal of Righteousness "to gain almost twice as much benefit from Spell Damage than intended" and give Consecration 0.042. The sim counts the seal's spell power once (12% a hit with a two-hander, 7.65% with a one-hander) and Consecration at 4.2%, so `pal_sor_sp_double`, `pal_sor_sp_speed` and `pal_consecration_coef` left the report (see "Settled, not a scenario").
5. **`hunter_pet_speed` and `hunter_pet_scaling`.**
   - Pet speed: the beta's fast pets hit as hard as a 2.0 sec pet. The presets treat that as a bug and use Classic's speed-scaled hit (user ruling 2026-10-15, PR #22); the beta's rule would add 6.1% to Beast Mastery and 3.2% to Marksmanship (assumption sims, 180 s).
   - Pet scaling: Season of Discovery's 22% scaling instead of the measured 10% would add 41 / 23 / 18 to Beast Mastery / Marksmanship / Survival (PR).

Close behind: `sham_lb_31plus` (−1.9% / +4.1% Elemental around the presets' ×1.12, assumption sims).

---

## All classes

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `wf_totem_offhand` | Windfury Totem procs from off-hand hits | Fury (dual wield), all six rogue builds, Survival | no, main hand only (one value, user ruling 2026-10-16) | = Expected | main hand only | = Expected | none left (one value) |

### `wf_totem_offhand`: Windfury Totem procs from off-hand hits

- **Ruling (2026-10-16):** main hand only, kept as a one-value row. Client 70425 edited the totem's description to "Each main hand hit has a 20% chance", and the user ruled: "Main hand only, and keep as assumption with only 1 value". `sim/core/buffs.go` (`WindfuryTotemBuffAura`) now uses `ProcMaskMeleeMH` under Forever too. The history below is kept for reference.

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

---

## Priest

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `priest_sf_tick_crit` | Shadowform's 200% crit on ticks | Shadow, Shadow + Holy | 200% on SW:P, Mind Flay, DP ticks | 150% on ticks | 200% | = Expected | est. −3% (Low) |
| `priest_smite8_tuning` | Smite VIII 31+ tuning | Smite | ×1.09 | ×1.00 | ×1.09 | ×1.18 | −0.9% Low, +0.9% High (assumption sims, 180 s) |

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
- **Sim now:** rank 8's base damage ×1.09 (`SmiteRank8DamageScale`, applied in `SmiteBaseDamageAt`, `sim/priest/smite.go`; user ruling 2026-10-15, PR #21). The client's rank 8 is 166.62 to 186.38 base.
- **Source:** client 1.60.1.70338.
- **Why uncertain:** Smite's ratio to Era is 0.51 at rank 4 (level 30) but 0.42 / 0.41 / 0.43 at ranks 6 to 8. Build 70291 retuned ranks 1 to 4 and no 31+ rank (rank tuning audit).
- **Settings:**
  - **Low, ×1.00.** The client as is: no 31+ rank has changed in 16 beta builds.
  - **Expected, ×1.09 (Sim now, user ruling 2026-10-15).** Halfway. Blizzard's level-curve passes have moved the top playable rank 0 to 12%, so a partial retune before release is likely but its size is not known.
  - **High, ×1.18.** Held at rank 4's 0.51 ratio to Era.
- **Sensitivity (assumption sims, 180 s, against the ×1.09 default):** Low −0.9%, High +0.9%.
- **Knob:** `core.Assume("priest_smite8_tuning", SmiteRank8DamageScale)` in `SmiteBaseDamageAt` (patch 0004); the setting is the whole multiplier on the client value.

---

## Warlock

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `lock_pet_scaling` | Demon scaling from the warlock | all four | SoD | half of SoD | SoD | = Expected | −1.8 to −2.2% at Low (assumption sims, 180 s) |
| `lock_demon_melee` | Succubus melee damage | all four | 1.12 (95–131) | halved | 1.12 | = Expected | Low −3.9 to −4.0% (PR) |
| `lock_bod_coef` | Bane of Doom coefficient | Demo, Destro, Aff | 4.0 | 2.0 | 4.0 | = Expected | Low −4.1% (PR; 1.0 gives −6.1%) |
| `lock_incinerate_bonus` | Incinerate's 25% with Immolate | Destruction Fire | whole hit | = Expected | whole hit | = Expected | one value, kept for reference |

### `lock_pet_scaling`: Warlock Pet Scaling coefficients

- **Superseded 2026-10-23:** the sim no longer uses the Season of Discovery shares below. It takes wowsims/forever's measured ones (0.17 attack power and 0.10 spell damage a point of the warlock's spell damage, 8/17 of its spell hit as melee hit, 7.5 health a point of Stamina, 35% armor, MP5 from Intellect; c3d9ea9516, 9ce48b6aba, 19922b122e, 7097a63bc9, f5b2f7ef33; notes.md). The settings and the knob in `tools/dps_report/patches/0004` still describe the old model and need redoing before the next assumption run.

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
  - **Low, half of SoD.** Half of each Season of Discovery share (0.25 attack power and 0.125 spell damage a point of spell power, half the hit), if Forever toned the inherited shares down.
  - **Expected, SoD (Sim now).** Forever carried Season of Discovery's aura 416189 unchanged, so Season of Discovery's server values are the closest analogue.
  - **High:** = Expected. "Nothing" and "SoD plus the warlock's spell crit" were dropped (user ruling 2026-10-15).
- **Sensitivity (assumption sims, 180 s):** half of SoD gives Demonology −2.16%, Destruction −1.87%, Affliction −1.84%, Destruction Fire −1.87%. The dropped settings last measured −3.7 to −4.3% (nothing) and +2.1 to +2.4% (SoD plus crit) at 180 s.
- **Knob:** `sim/warlock/pet.go`, each inherited field scaled by `Assume("lock_pet_scaling", 1)` (0.5 for Low).

### `lock_demon_melee`: the Succubus's melee damage

- **Superseded 2026-10-23:** the Succubus swings for 66 to 103 at 60 on a fixed 186 attack power (0.698 of the creature auto attack dps, ±21.5%, wowsims/forever b207517048, from level 20 and 30 sheets). The 1.12 values below are no longer the sim's; the settings need redoing.

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
  - **Expected, whole hit (Sim now).** The tooltip reads "25% increased damage", and the client stores a percent. The only value, kept in the report as a reference row (user ruling 2026-10-15).
  - "25% of the base damage only" (WotLK's reading) was dropped; its last assumption sim gave −6.19% on the Fire build at 180 s.
- **Knob:** `sim/warlock/incinerate.go:43-46`. For Low, compute a separate flat add of `0.25 * rolled base` instead of the multiplier.

---

## Paladin

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `pal_2h_spec_holy` | Two-Handed Weapon Spec on Holy seal damage | twist, sub Holy, Shock | yes | Physical only | yes | = Expected | Low −1.1% twist (PR) |
| `pal_31plus_ranks` | Holy Strike and Consecration 31+ ranks | twist, sub Holy, Shock | client | held to the ≤30 trend | client | = Expected | Low −1.2% (PR) |

Retribution was also affected by `flat_bonus_inside_pct`, settled 2026-10-16 (Holy Strike; +3.6 to +4.2% at its Low, assumption sims).

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
| `warr_rend_ap` | Rend's attack power share | Arms, Arms vs Dragonkin | 2% of attack power a tick | none (client) | 2% a tick | 5% a tick | Low −2.0%, High +3.0% (assumption sims, 180 s) |

Dual-wield Fury is also affected by `wf_totem_offhand` (est. −2 to −3% at its Low).

### `warr_rage_k`: how normalised Rage grows with level

- **Builds:** Fury, Fury two-hand, Arms.
- **Sim now (standards 21):** a landed white hit gives k(L) × weapon swing speed in Rage; misses, dodges and parries give none (since 2026-10-23, wowsims/forever f9f9f21883). k30 is 3.46 for a one-hander and 4.5 for a two-hander, and k(L) = k30 × (1 + s × (L − 30) / 30) above level 30, with s = 0.5 (Expected). Dual Wield Specialization multiplies on top; a crit adds 100% of the swing's one-hand-rate Rage (settled; since 2026-10-23 a two-hander crit gives (4.5 + 3.46) × speed × the level factor, not 9 ×, wowsims/forever 5ac4ffbcb6). The code is `sim/core/rage.go`, behind `Env.IsForever()`; only the warrior enables it.
- **Source:** Blizzard's Priest & Warrior deep dive ("fully-normalized ... based on the speed of the weapon that dealt the damage"), tier 3. The constants are beta testers' measurements at level 30 or below (forum thread 2355684; Icy Veins and wowforeverbuilds repeat them), tier 5.
- **Why uncertain:** the beta stopped at level 30, and Blizzard gives no formula. At 30 the measured k gives about Classic's Rage per swing; flat at 60 it gives about half (a 2.6 s main hand: 9.0 against Classic's 18.6 for an average swing). Nobody knows whether Blizzard scales it with level. The full test plan is under "Rage needs a beta test" in orchestration/followups.md.
- **Settings (user ruling 2026-10-13):**
  - **Low, flat (s = 0).** The level 30 constants unchanged at 60.
  - **Expected, midpoint (s = 0.5).** ×1.5 at 60.
  - **High, Classic-like (s = 1).** ×2 at 60, close to Classic's Rage per average swing.
- **Sensitivity:** with flat k, Fury is 950.6 at 180 s (it was 1055.0 under Classic's damage-based Rage), Fury two-hand 835.0 and Arms 861.4. The warrior re-measures all three settings.
- **Knob:** a package variable for the level slope s, next to `core.ForeverNormalizedRageK` (which multiplies k). The dashboard sets it with `DASH_ASSUME`.

### `warr_dt_rage_k` and `warr_dt_rage_mods`: Rage from damage taken (pending, no knob in the report yet)

- **Sim now (standards 21, since 2026-10-23):** a hit or periodic tick taken gives 2.5 × (1 − A) × D / c(attacker level) Rage (`core.ForeverDamageTakenRage`). c is the Rage conversion at the attacker's level (230.6 at 60, 243.6 at 63) and A the armor assumed by the attacker's level: 20% to 42, 30% at 44, 36% at 46, 38% at 50, 40% at 63, interpolated. Our level, health and armor don't count. Berserker Rage doubles it. A 1500 raid hit from the level 63 boss gives 9.2 Rage (it was 3.4).
- **What changed on 2026-10-23 (user-approved alignment with wowsims/forever):** `warr_dt_rage_k` was K = 20 × (1 − a(our level)) × D / a normal mob's health at our level, with K fitted on pre-fix level 8 to 9 logs (magey/forever-warrior issue 3), and a High of K = 48. Both are superseded: wowsims/forever ee2d69b16a, 73b317c7cf and de550b7df9 fit the rule above on post-8 October logs (warriors at 18 and 20 and bears at 30, attackers 16 to 60, 187 melee hits at 16 to 42, geared and naked: health and armor don't move the rate). The key stays, now for the factor 2.5 (`core.ForeverDamageTakenRageFactor`): it is measured on melee hits, and their logs show spells landing at about 1 to 2 times it (Fireball about 1.2), which matters because the raid damage hits (standards 24) are Fire spells. `warr_dt_rage_mods` keeps its question and its None default: D is still the attacker's hit before any of our modifiers, times the crit or crushing blow multiplier, which is also what wowsims/forever counts (their pre-mitigation damage leaves out the defender's damage taken modifiers). Only the formula it feeds changed, and Berserker Rage now doubles the Rage instead of adding a flat 2 a melee hit.

### `warr_rend_ap`: Rend's attack power share

- **Builds:** Arms and Arms vs Dragonkin (the two-hand builds that keep Rend up).
- **Sim now:** each Rend tick adds 2% of attack power (`ForeverRendAPPerTick`, `sim/warrior/rend.go`; user ruling 2026-10-15, PR #22), about 14% of attack power over a full Rend. Since 2026-10-23 (user-approved) the tick reads attack power and the damage modifiers when it lands, not at the application, its crits take Impale, and it crits as before (wowsims/forever 1172f1c745, 36cfc58328, 872e6def91).
- **Source:** Blizzard's warrior development notes say Rend now "scales significantly" with attack power; client 70338 holds no coefficient. wowsims/forever 1172f1c745 measured 2% a tick on rank 2 (level 13 and 14 combat logs, 99 to 154 attack power: 7 + 0.02 × attack power, read on every tick), which backs Expected.
- **Why uncertain:** the share is server side, measured on rank 2 only, and Rend at 60 is untested.
- **Settings:**
  - **Low, none.** The client as it is.
  - **Expected, 2% a tick (Sim now).** A modest share.
  - **High, 5% a tick.** About 35% of attack power over a full Rend.
- **Sensitivity (assumption sims, 180 s, against 2%):** Low −2.0%, High +3.0% on both builds.
- **Knob:** `core.Assume("warr_rend_ap", ForeverRendAPPerTick)` in Rend's tick (patch 0004; in its snapshot until 2026-10-23).

---

## Hunter

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `hunter_pet_scaling` | Pet scaling from the hunter | BM, MM, SV | 10% RAP + 1:1 hit (17/8 as spell hit) | 10% RAP, no hit | 10% RAP + hit | SoD: 22% RAP + hit | Low −17 / −10 / −10, High +41 / +23 / +18 (PR) |
| `hunter_pet_speed` | Pet hits normalized to a 2.0 sec pet | BM, MM | no: Classic, the hit scales with speed | = Expected | Classic | the beta's same hit | High +6.1% BM, +3.2% MM (assumption sims, 180 s) |
| `hunter_tsa_hm_rank` | Trueshot Aura and Hunter's Mark at 60 | LW, MM, BM | 100 / 143 (Trueshot Aura fixed, Hunter's Mark hotfixed) | 50 / 71 (top ranks as written); Mid 75 / 98 (best older rank) | 100 / 143 | = Expected | Low −2.6 to −3.1%, Mid −1.4 to −1.7% (assumption sims, 180 s) |
| `hunter_crit_per_agi` | Crit per Agility at 60 | all four | ×1.22 | Classic (53 Agi a 1%) | ×1.22 | = Expected | Low −2.6 to −2.8% (assumption sims, 180 s) |
| `hunter_raptor_31plus` | Raptor Strike 31+ flat bonuses | Survival | 96 at rank 8 | 70 | 96 | 123 | Low −0.5%, High +0.5% (assumption sims, 180 s) |

Survival is also affected by `wf_totem_offhand` (est. about −1.5% at its Low).

### `hunter_pet_scaling`: Hunter Pet Scaling (attack power and hit)

- **Builds:** Beast Mastery (pet 26% of the damage), Marksmanship (14%), Survival. Lone Wolf has no pet.
- **Sim now:** the pet takes 10% of the hunter's ranged attack power, all of the hunter's melee crit (the higher of its melee and ranged crit, which is the melee crit in the sim) and its melee hit 1:1, plus 17/8 of that hit as spell hit since 2026-10-23 (forever-bugs #394, wowsims/forever 9ce48b6aba). `makeStatInheritance` in `sim/hunter/pet.go`. Since 2026-10-23 the pet's own attack power is a fixed 254 (30% of the level 60 creature dps as attack power) with none from Strength, its swing is the creature dps over the swing ±20%, and it has no crit from Agility (wowsims/forever 643d636345, e56aaeb1a9, 684209eb0a): the earlier sensitivities below predate that.
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
- **Sim now:** Classic's rule: a pet's hit scales with its attack speed, so every pet has the same white damage a second (`ForeverPetHitScalesWithSpeed`, `sim/hunter/pet.go`; user ruling 2026-10-15, PR #22). On the beta every pet's auto attack hits the size of a 2.0 sec pet's (97 to 145 at 60 since 2026-10-23, the creature dps over 2 sec ±20%; Classic's 18.17 to 27.66 × 2 before) at any attack speed, so the fastest pet deals the most.
- **Source:** beta players, gathered on ForeverChanges' hunter pets page and the forever-hunter wiki. The code comment says "nobody knows yet whether that is intended".
- **Why uncertain:** it could be a beta bug that Blizzard fixes, which would bring back Classic's hit size scaling with the swing.
- **Reconfirmed 2026-10-16:** the assumptions audit found the beta behaviour still standing: EU forum thread 632244 (level 16 Tallstriders, a 1.3 sec pet hitting 27 to 38 against 31 to 42 for a 2.0 sec pet), forever-bugs #31 still open, and no fix in any note through 15 October. Asked again, the user ruled "Keep treating it as a bug", so Classic's speed scaling stays the default.
- **Extended 2026-10-23 (user ruling):** Cobra Reflexes (25076, 30% attack speed "but reduces damage", no number; the presets have it on since 2026-10-23) takes the same reading: the pet's hits shrink with the speed, so it adds only Pet Aggression's 8 a hit on the extra swings and the procs (+6 Beast Mastery). The swing is the passive's client row since then (Faster Attack I 17%, 1.71 sec; Slower Attack III -26%, 2.7 sec). Under High a Cobra Reflexes pet's hits stay a 2.0 sec pet's too (wowsims/forever keeps TBC's 15% cut instead).
- **Settings:**
  - **Expected, Classic (Sim now).** The hit scales with attack speed, treating the beta's behaviour as a bug (user ruling 2026-10-15).
  - **High, as on the beta.** The only observed behaviour: every pet hits like a 2.0 sec pet.
  - **Low:** = Expected.
- **Sensitivity (assumption sims, 180 s, at High):** +6.1% Beast Mastery, +3.2% Marksmanship.
- **Knob:** `core.AssumeOk("hunter_pet_same_hit")` = 1 restores the beta's 2.0 sec hit in `applyForeverRules` (patch 0004).

### `hunter_tsa_hm_rank`: Trueshot Aura and Hunter's Mark at level 60

- **Builds:** Lone Wolf, Marksmanship, Beast Mastery. Both are ranged attack power only, so Survival is unaffected.
- **Sim now:** the top ranks cast directly, Trueshot Aura rank 5 +100 RAP (fixed onto the curve, user ruling 2026-10-15, PR #22) and Hunter's Mark rank 4 +143 RAP (`ForeverTrueshotAuraRank5RAP`, `ForeverHuntersMarkRank4` in sim/core; standards 8a). Hunter's Mark was +140 by the 2026-10-15 ruling; on 2026-10-23 the user took the +143 of the server hotfix foreverchanges.pro saw on 2026-10-21 (client 70532 still reads 71), so only Trueshot Aura's rank is still assumed.
- **Source:**
  - Client 70338: Trueshot Aura ranks 30 / 40 / 50 / 75 / 50, Hunter's Mark 26 / 59 / 98 / 71. Both top ranks sit below the rank before them.
  - Standards 8a: use the strongest learned rank. The hunter branch implements it (PR #5), worth +10.7 to +14.3 DPS.
- **Why uncertain:** a top rank weaker than the one before is either a client data error Blizzard fixes, or a value players work around by casting the lower rank.
- **Settings:**
  - **Low, 50 / 71, top ranks as written.** If only the weaker top ranks count.
  - **Mid, 75 / 98, best older rank.** A player casts the stronger lower rank (the presets' setting before PR #22).
  - **Expected, 100 / 143 (Sim now).** Blizzard fixes the top ranks onto the curve of the ranks before them: Trueshot Aura +25 a rank (user ruling 2026-10-15). Hunter's Mark was taken at about +40 a rank (140, Forever's ranks 1 to 3 running about 1.3× Classic's 20 / 45 / 75 / 110); the 2026-10-21 server hotfix put it at 143 (user, 2026-10-23).
  - **High:** = Expected.
- **Sensitivity (assumption sims, 180 s):** Low −2.6% BM, −3.1% LW, −2.8% MM; Mid −1.4% BM, −1.7% LW, −1.6% MM. Trueshot Aura is a raid buff in 38 of the 46 report requests (and in every hunter request since client 70425 made it a raid aura, 2026-10-16: the sensitivity above predates that and now covers Survival too), but it gives ranged attack power only, so only the hunters move.
- **Knob:** `Assume("hunter_tsa60", …)` and `Assume("hunter_hm60", …)` override the level 60 values `bestLearnedRank` returns (patch 0004); the settings set 75 / 98 and 50 / 71.

### `hunter_crit_per_agi`: the hunter's crit per point of Agility

- **Builds:** all four. Lone Wolf, Marksmanship and Beast Mastery stack Agility; the pets inherit the hunter's melee crit.
- **Sim now:** Classic's table ×1.22 at every level (`ForeverCritPerAgiScale`, a stat dependency in `sim/hunter/hunter.go` on top of Classic's 0.0189% a point at 60; user ruling 2026-10-15, PR #22), about 43 Agility a 1%.
- **Source:** Classic base stats.
- **Why uncertain:** at level 1 the beta's hunter gets 0.2174 crit a point where the sim has 0.1783 (dodge 0.4348 against 0.3567), 22% more (notes.md, All classes, 2026-10-10). The paladin, rogue, priest, mage and druid match the sim at level 1, so Forever changed the hunter's table; how that carries to 60 is unknown.
- **Settings:**
  - **Low, Classic.** The level 1 difference stays a level 1 quirk.
  - **Expected, ×1.22 at every level (Sim now).** The beta's measured level 1 ratio, carried to 60.
  - **High:** = Expected; no data above level 1.
- **Sensitivity (assumption sims, 180 s, at Low):** −2.6 to −2.8% on all four builds.
- **Knob:** `core.Assume("hunter_crit_per_agi", ForeverCritPerAgiScale)` in the stat dependency (patch 0004). The warrior's level 1 difference is −3% and its Agility is small, so it isn't worth a knob.

### `hunter_raptor_31plus`: Raptor Strike's level 31+ flat bonuses

- **Builds:** Survival (Raptor Strike 11% of its damage).
- **Sim now:** rank 8 adds 96 (`RaptorStrikeRank8BaseDamage`, `sim/hunter/raptor_strike.go`; user ruling 2026-10-15, PR #22); ranks 5 to 7 keep the client's 35 / 40 / 55.
- **Source:** client 70338.
- **Why uncertain:** level 24's flat is 0.88 of Era, but the 32 to 56 ranks sit at 0.50 to 0.70, so the playable ratio isn't kept (rank tuning audit, "stale-shaped").
- **Settings:**
  - **Low, 70.** The client.
  - **Expected, 96 (Sim now).** Halfway: a partial retune of the kind Blizzard has made to the ≤30 ranks.
  - **High, 123.** Held at 0.88 of Era's 140, ×1.76.
- **Sensitivity (assumption sims, 180 s, against 96):** Low −0.5%, High +0.5%.
- **Knob:** `core.Assume("hunter_raptor8", …)` on rank 8's base damage (patch 0004); the preset casts rank 8 only.

---

## Mage

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `mage_ice_lance_coef` | Ice Lance spell power coefficient | Frostfire, Frost | 0.429, cast on every Fingers of Frost charge | 0 (client, never casting Ice Lance); Mid 0.2145 (half of SoD) | 0.429 | = Expected | Low Frost −19.1%, Frostfire −9.4%; Mid −10.9% / −7.3% (assumption sims, 180 s) |
| `mage_frostbolt10` | Frostbolt X gets 70291's raise | Frost | ×1.04 | ×1.00 | ×1.04 | ×1.08 | −1.0% Low, +1.0% High (assumption sims, 180 s) |
| `mage_imbue_scrolls` | Mage staff imbue scrolls | all four mages | used | not available | used | = Expected | −0.8 to −1.5% at Low (assumption sims, 180 s) |
| `mage_cone_of_cold` (pending, no report row yet) | Frost at melee range with Cone of Cold | Frost | 25 yards, no Cone of Cold | = Sim now | = Sim now | 10 yards, Cone of Cold on cooldown with Winter's Chill 5 and Improved Cone of Cold 3 | +10.9% at High (910.5 against 820.7, 180 s, 20k iterations) |

### `mage_ice_lance_coef`: Ice Lance's spell power coefficient

- **Builds:** Frostfire and Frost. Fire and Arcane don't take Ice Lance.
- **Sim now:** 0.429 on all six ranks (`IceLanceSpellCoeff`, `sim/mage/ice_lance.go`), with ×4 against Frozen targets, and the Frost and Frostfire Level 60 rotations cast Ice Lance on every Fingers of Frost charge (user ruling 2026-10-15, PR #21).
- **Source:** client 70338 has 0. Season of Discovery's client had 0 on the same spell too (1.15.9 spell 400640 on wago.tools), and its server gave it 42.9%. Every other new Forever damage spell carries a client coefficient (Frostfire Bolt 0.814, Arcane Blast 0.714).
- **Why uncertain:**
  - Forever's server could add a different value, or none.
  - Forever also lowered the Frozen bonus (client 300 where Season of Discovery's was 400), so Blizzard did touch the spell.
  - The talent is reachable on the beta (20 Frost points), but no test has been posted.
- **Settings:**
  - **Low, 0 (the client).** The client's value, which means never casting Ice Lance: each build takes the better of its Level 60 rotation and the one before Ice Lance was added (`tools/dps_report/asm_apls/`, the APLs at 27ffd2063^1), and both builds keep the no-Ice-Lance rotation (user ruling 2026-10-15).
  - **Mid, 0.2145 (half of SoD).** If the server adds less than Season of Discovery's. It uses the Ice Lance rotation.
  - **Expected, 0.429 (Sim now).** Season of Discovery's server value for the same client data.
  - **High:** = Expected. WotLK's 0.143 was dropped (user ruling 2026-10-15).
- **Sensitivity (assumption sims, 180 s):** Low Frost −19.12% (−181.3 DPS), Frostfire −9.42% (−95.3); Mid Frost −10.86%, Frostfire −7.34%.
- **Knob:** `core.Assume("mage_ice_lance_coef", IceLanceSpellCoeff)` (patch 0004).

### `mage_frostbolt10`: does build 70291's Frostbolt raise reach rank 10?

- **Builds:** Frost (Frostbolt 98% of its damage).
- **Sim now:** rank 10's base damage ×1.04 (`FrostboltRank10DamageScale` in `FrostboltBaseDamageAt`, `sim/mage/frostbolt.go`; user ruling 2026-10-15, PR #21). The client's rank 10 is 382 to 413 at 60.
- **Source:** client 70338.
- **Why uncertain:** 70291 raised the level 20 and 26 Frostbolt ranks (+8% at 26's cap) and left 31+ alone. The 31+ ratios to Era already sit above the playable one, so it's unclear whether the raise should carry (rank tuning audit, "mild lag").
- **Settings:**
  - **Low, ×1.00.** The client.
  - **Expected, ×1.04 (Sim now, user ruling 2026-10-15).** Halfway; the evidence points both ways.
  - **High, ×1.08.** 70291's raise applied to rank 10.
- **Sensitivity (assumption sims, 180 s, against ×1.04):** Low −1.0%, High +1.0%.
- **Knob:** `core.Assume("mage_frostbolt10", FrostboltRank10DamageScale)` in `FrostboltBaseDamageAt` (patch 0004); the setting is the whole multiplier on the client value.

### `mage_imbue_scrolls`: the mage staff imbue scrolls

- **Builds:** all four mages.
- **Sim now:** each preset takes its best scroll on top of Wizard Oil, as the class option `staffImbue`: Greater Flame for Frostfire and Fire, Greater Frost for Frost, Precision for Arcane (user ruling 2026-10-15, PR #21).
- **Source:** the scrolls are mage-only items (277500 to 277502); foreverchanges lists them as hidden.
- **Why uncertain:** hidden items may not be in the game at launch.
- **Settings:**
  - **Low, not available.** No scroll.
  - **Expected, available (Sim now).**
  - **High:** = Expected.
- **Sensitivity (assumption sims, 180 s):** Low Frost −1.48%, Fire −1.23%, Frostfire −1.03%, Arcane −0.75%.
- **Knob:** a request edit (`no_staff_imbue` in `run_assumptions.py`) that takes the option off.

### `mage_cone_of_cold`: a Frost mage standing in Cone of Cold range (pending, added 2026-10-23)

- **Builds:** Frost.
- **Sim now:** the preset stands at 25 yards like every caster and never casts Cone of Cold (user decision 2026-10-23: keep the Frost preset at range).
- **Source:** Cone of Cold and Winter's Chill are simulated since 2026-10-23 (`sim/mage/cone_of_cold.go`, the 2026-10-20 hotfix push 112694: 5% damage taken and 7% of spell power on the cone per Winter's Chill stack, up to 5). Cone of Cold only reaches 10 yards, so the mage has to stand that close. wowsims/forever's Frost preset does (41214cc8d0: 10 yards, Cone of Cold on cooldown, Cold Snap when it has 5 sec or more left, talents `05022050011003--053500030100331251`).
- **Why uncertain:** whether a raid Frost mage can stand within 10 yards of the boss: melee cleaves, frontal and point-blank boss abilities, and the cone has to face the boss.
- **Settings:**
  - **Expected (Sim now):** at range, no Cone of Cold.
  - **High, at 10 yards:** three request edits. Rotation `tools/dps_report/asm_apls/mage__frost.cone_of_cold.apl.json` (the mage_frost template with the preset's knobs plus `coneOfCold=1,coldSnapCone=5`: Cold Snap, then Cone of Cold, right after the Ice Lance line), talents `05022050011003--053500030100331251` (wowsims/forever's, which parse on our tree as they are: Winter's Chill 5 and Improved Cone of Cold 3 in place of Frost Warding 2, Improved Blizzard 3, Permafrost and Ice Block, and Arcane Subtlety 2 and two more Arcane Meditation points in place of two Arcane Impact points and Incineration 3), and `distanceFromTarget` 10.
- **Sensitivity (180 s, 20k iterations, seed 1, on the Frost request):** 910.5 against the preset's 820.7 (+89.8, +10.9%). The preset at 10 yards without Cone of Cold is 824.8 (the bolts land sooner). Our preset's points with the cone (`05020050031001-02-053500030100331251`, Incineration 2) are 896.2, and the High talents at 25 yards, where the cone can't be cast, are 832.3. About 17 Cone of Cold casts a fight.
- **Knob:** none in `run_assumptions.py` yet; the pipeline needs an edit that sets the rotation, talents and distance together.

---

## Shaman

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `sham_mw_ppm` | Maelstrom Weapon procs a minute (5/5) | all four Enhancement builds | 10 PPM | 5 | 10 | 20 | est. Low −3%, High +5% |
| `sham_lb_31plus` | Lightning Bolt 31+ ranks | Elemental, Elemental Mana Tide | ×1.12 | ×1.00 | ×1.12 | ×1.38 | −1.9% Low, +4.1% High (assumption sims, 180 s) |
| `sham_es7_tuning` | Earth Shock VII tuning | all four Enhancement builds, both Elemental builds | ×1.08 | ×1.00 | ×1.08 | ×1.16 | −0.3 to −0.4% Low, +0.3 to +0.4% High on Enhancement; 0 on Elemental (assumption sims, 180 s) |

### `sham_mw_ppm`: Maelstrom Weapon's proc rate

- **Builds:** Enhancement Level 60, Imp GW, Water Shield, Mana Tide (all 5/5). Lightning Bolt goes out only at 5 stacks and is about 7% of the damage.
- **Sim now:** 2 procs a minute a point, 10 PPM at 5/5, on any landed melee hit. `sim/shaman/talents.go:497-499` (TODO: "The tooltip has no proc rate, beta testing will have to measure it").
- **Source:** our guess (shaman_audit.md 2.7: "2 a minute a point, a guess"). The client leaves the rate to the server.
- **Why uncertain:** the talent needs level 34+, and no beta character can test it.
- **Settings:**
  - **Low, 5 PPM.** Half the analogues, for a Forever that spreads a weaker proc over its five ranks (1 a point). This is the weakest-grounded value in the inventory.
  - **Expected, 10 PPM (Sim now).** Season of Discovery's rune base rate (wowsims/sod `sim/shaman/runes.go`: "Base PPM is 10") and the 2-a-rank rate of the talent this one copies.
  - **High, 15 PPM.** As far above the base rate as Low is below it. Season of Discovery's 20 PPM (its 2024-02-20 hotfix for Windfury Weapon on a two-hander) was dropped: it was a rune hotfix that doesn't carry over (user ruling 2026-10-15).
- **Sensitivity (assumption sims, 180 s):** Low −2.9 to −3.2%, High +2.5 to +3.1%.
- **Knob:** `sim/shaman/talents.go:499`. Change `2*float64(...)` to `Assume("sham_mw_ppm", 10)/5*float64(...)`.

### `sham_lb_31plus`: Lightning Bolt's level 31+ ranks

- **Builds:** Elemental (Lightning Bolt about 65% of its damage with overloads), Elemental Mana Tide, and all four Enhancement builds (their Maelstrom Weapon Lightning Bolts are rank 10).
- **Sim now:** rank 10's base damage ×1.12 (`LightningBoltRank10DamageScale` in `LightningBoltBaseDamageAt`, `sim/shaman/lightning_bolt.go`; user ruling 2026-10-15, PR #21). The client's rank 10 is 189.92 to 211.68 plus 1.2 a level to 60.
- **Source:** client 70338 (the ForeverChanges downrank calculator reads the same).
- **Why uncertain** (rank tuning audit, "stale-shaped"):
  - Lightning Bolt is 0.61× Era at levels 26, 32 and 38, then 0.54 / 0.48 / 0.44 at 44 / 50 / 56.
  - At 60 it does 78 base a cast-second against 84 to 158 for the other classes' nukes.
  - Build 70291 raised the top playable rank 12% and no 31+ rank.
- **Settings:**
  - **Low, ×1.00.** The client.
  - **Expected, ×1.12 (Sim now, user ruling 2026-10-15).** 70291's raise on the same spell, applied to rank 10.
  - **High, ×1.38.** Held at the 0.61 ratio of levels 26 to 38.
- **Sensitivity (assumption sims, 180 s, against ×1.12):** Elemental Low −1.9%, High +4.1%; Enhancement Low −0.2 to −0.3%, High +0.5 to +0.6%. Elemental stays in the bottom third either way.
- **Knob:** `core.Assume("sham_lb_31plus", LightningBoltRank10DamageScale)` in `LightningBoltBaseDamageAt` (patch 0004).

### `sham_es7_tuning`: Earth Shock VII

- **Builds:** all four Enhancement builds (Earth Shock about 10% of the damage), and both Elemental builds, whose rotation casts no Earth Shock (0 change).
- **Sim now:** rank 7's base damage ×1.08 (`EarthShockRank7DamageScale` in `EarthShockBaseDamageAt`, `sim/shaman/earth_shock.go`; user ruling 2026-10-15, PR #21). The client's rank 7 is 293.07 to 308.93.
- **Source:** client 70338.
- **Why uncertain:** 0.66× Era at level 24, then 0.57 / 0.55 / 0.57 from 36 to 60 (rank tuning audit, "stale-shaped, mild").
- **Settings:**
  - **Low, ×1.00.** The client.
  - **Expected, ×1.08 (Sim now, user ruling 2026-10-15).** Halfway.
  - **High, ×1.16.** Held at 0.66.
- **Sensitivity (assumption sims, 180 s, against ×1.08):** Enhancement Low −0.3 to −0.4%, High +0.3 to +0.4%.
- **Knob:** `core.Assume("sham_es7_tuning", EarthShockRank7DamageScale)` in `EarthShockBaseDamageAt` (patch 0004).

---

## Druid

| Key | Label | Builds | Sim now | Low | Expected | High | Sensitivity |
|---|---|---|---|---|---|---|---|
| `druid_ooc_rate` | Omen of Clarity proc rate | Feral, Balance, Balance Wrath | 2 PPM; spells 2 × cast / 60 | = Expected | 2 PPM | = Expected | one value, kept for reference |
| `druid_rip_31plus` | Rip's level 31+ ranks | Feral | client | = Expected | client | Era; ≤30 trend ×1.13 (Higher) | +1.2% Era, +1.6% ≤30 trend (assumption sims, 180 s) |
| `druid_wrath8_tuning` | Wrath VIII tuning | Balance Wrath, Balance | ×1.085 | ×1.00 | ×1.085 | ×1.17 | −0.8% Balance Wrath, −0.2% Balance Low, +0.8% / +0.2% High (assumption sims, 180 s) |

### `druid_ooc_rate`: Omen of Clarity's proc chance

- **Builds:** Feral (each Clearcasting is a free Shred), Balance and Balance Wrath (free Starfires; the Wrath build casts Starfire only on Clearcasting).
- **Sim now:**
  - Melee: 2 PPM (`sim/druid/talents.go:449`). Under Forever specials roll at the swing speed like white hits, 1.0 sec in Cat Form (`:468-469`); until 2026-10-23 they rolled at the equipped weapon's speed, 11.7% a special on a 3.5 sec weapon. Changed on wowsims/forever ca7b3ab462 and af6df6c0e4 (their issue #225: the weapon speed "is a game bug fixed in the next beta build"): Feral 8.49 to 6.16 Clearcasts a 180 sec fight, -10.4 DPS.
  - Spells: 2 × max(cast time, 1.5 sec) / 60 a cast (`:474`), an instant at 1.5 sec, as in wowsims/forever. Under Forever caster Faerie Fire rolls it too, though it deals no damage, and Faerie Fire (Feral) does not (`:496`, since 2026-10-23 on wowsims/forever 1a2f22fe3f and a51a807bf4: their logs list Faerie Fire in 23 of 365 procs).
  - Moonkin Form doubles the chance and halves the 10 sec cooldown (`:475-477`, `:482-484`).
- **Source:**
  - Melee: 1.12's rate, with Forever's specials at the swing speed (above).
  - Spells: our own guess (`:390-392`); wowsims/forever uses the same rate.
  - The Moonkin modifiers and the cooldown are the client's (16864).
- **Why uncertain:** the client leaves the chance to the server (notes.md, Druid).
- **Settings:**
  - **Expected, the sim's rate (Sim now).** 1.12's 2 PPM, the version Forever's baseline passive was ported from. The only value, kept in the report as a reference row (user ruling 2026-10-15).
  - WotLK's flat 6% a spell and the later expansions' 3.5 PPM were dropped. In the last assumption sims they moved the three druid builds −0.06 to +0.17% (6% a spell) and +0.24 to +1.44% (3.5 PPM), at 180 s.
- **Knob:** `sim/druid/talents.go:424` (PPM) and `:445` (the spell formula). Low needs a branch: `chance = 0.06` when the key says so.

### `druid_rip_31plus`: Rip's level 31+ ranks

- **Builds:** Feral (Rip 22 to 25% of the damage, at 4 combo points).
- **Sim now:** client per-combo-point values 8.5 / 12.7 / 18.2 / 25.5 at ranks 3 to 6. `sim/druid/rip.go:22-25`.
- **Source:** client 70338.
- **Why uncertain:** Forever's ranks up to level 30 sit above Era per combo point (4.4 / 7.2 against 4 / 7), but the 31+ ranks sit 6 to 9% below (rank tuning audit).
- **Settings:**
  - **Expected, the client (Sim now).** Expected moved to the client value (user ruling 2026-10-15).
  - **High, Era's per-combo-point values** (9 / 14 / 20 / 28). Forever's own ≤30 ranks are at or above Era, so Era is the floor of that design.
  - **Higher, the ≤30 trend, ×1.13.**
- **Sensitivity (assumption sims, 180 s):** Era +1.2%, the ≤30 trend +1.6%.
- **Knob:** scale `dmgTickPerCombo` (and `dmgTickBase`) of rank 6 at `rip.go:25` by the key.

### `druid_wrath8_tuning`: Wrath VIII

- **Builds:** Balance Wrath (Wrath 59% of the damage) and Balance (14%).
- **Sim now:** rank 8's base damage ×1.085 (`WrathRank8DamageScale` in `WrathBaseDamageAt`, `sim/druid/wrath.go`; user ruling 2026-10-15, PR #21). The client's rank 8 is 91.57 to 102.43.
- **Source:** client 70338.
- **Why uncertain:** build 70009 multiplied every Wrath rank by 1.5, then 70291 retuned only ranks 1 to 5. Forever over Era is 0.42 at level 30 and then 0.36 / 0.34 / 0.36 (rank tuning audit).
- **Settings:**
  - **Low, ×1.00.** The client.
  - **Expected, ×1.085 (Sim now, user ruling 2026-10-15).** Halfway.
  - **High, ×1.17.** Held at 0.42.
- **Sensitivity (assumption sims, 180 s, against ×1.085):** Low −0.8% Balance Wrath, −0.2% Balance, High +0.8% / +0.2%.
- **Knob:** `core.Assume("druid_wrath8_tuning", WrathRank8DamageScale)` in `WrathBaseDamageAt` (patch 0004).

---

## Rogue

No assumptions of its own. The rogue's scenarios are the shared ones:
- `flat_bonus_inside_pct` (Backstab, Ambush, Mutilate) was settled 2026-10-16 (see "Settled, not a scenario").
- **`wf_totem_offhand`:** every build dual wields in the Windfury Totem group. Est. −2 to −3% at its Low.

---

### `level_partial_resist` (pending, added 2026-10-16 by the assumptions audit)

- **Builds:** every caster and hybrid (mage, warlock, priest, Elemental, Enhancement, Balance, Retribution); not the rogues, warriors or hunters.
- **Sim now:** Classic's 2% average partial resist per level above the caster on every non-binary spell, 6% against the level 63 boss (`AverageMagicPartialResistPerLevelMultiplier`, `sim/core/spell_resistances.go`, `spell.go`). notes.md said the sim drops it under Forever, but the code keeps it.
- **Evidence:** 0 partial resists in about 520 landed non-binary hits on normal beta mobs 3 to 5 levels up (shaman_audit.md 5.2 and beta_results.txt; the earlier "313" counted only part of the log). Classic predicts 18 to 30% of such hits partly resisted, WotLK's model about 75%. Mobs with a school resistance do partly resist, in 10% steps: Elder Cloud Serpents (Nature 58) took 20% or 30% off all 15 Nature hits. No boss or elite data, and no Blizzard note or bug report on level resists (research 2026-10-16, orchestration/assumptions_audit.md section 7).
- **Settings:** Low (the sim now) is Classic's 6% at +3, relabelled "bosses keep it", the only reading the beta leaves room for. Expected is none (the beta). A 10%-step setting isn't needed: step size changes the variance, not the mean, and WotLK's 5 resistance a level would give 4.5% at +3, close to Low. The default stays on the sim's value: the user asked for more verification (Q1, 2026-10-16), with the settling test in notes.md.
- **Sensitivity (10k iterations, seed 1, 180 s, None against 6%):** Fire +7.1%, Arcane +6.3%, Elemental +6.4%, Balance +5.4%, Destruction +5.2%, Affliction +5.3%, Ret +4.2%, Shadow +3.8%, Enhancement +3.0%, Frost 0 (Frostbolt is binary).
- **Knob (to add):** `core.Assume("level_partial_resist", AverageMagicPartialResistPerLevelMultiplier)` at both uses.
- **Update 2026-10-23:** settled as none (user ruling 2026-10-16), and wowsims/forever's logs agree (ed4b4e93e5: no level-based partial resist in thousands of hits on mobs 1 to 9 levels up). The resist from school resistance now follows wowsims/forever's model (R / (|R| + 9L), 10% steps, Spell Penetration adding damage past the resistance; see notes.md). Against the 0-resistance boss it moves only builds with Spell Penetration.

### Pending rows added 2026-10-16 (assumptions audit, orchestrator's request)

None of these has a knob yet. The report-tree rebuild adds them; the defaults keep the sim's current behaviour.

- **`lock_pet_crit`: do demons inherit the warlock's spell crit?** (all four warlocks)
  - Source: AidanZMoon on forever-bugs #331 (13 to 14 October). Demons "should be getting their master's spell crit as their own melee AND spell crit chance", and the scaling stays off until the demons' wrong base crit is removed.
  - The sim gives the demon no crit from the warlock. Low (default): no crit, the beta today. Expected: the warlock's spell crit, Blizzard's intent.
  - Swing: +2.1 to +2.4% at 180 s ("SoD plus crit", the last assumption sims before the 2026-10-15 review dropped that setting).
  - Knob: `sim/warlock/pet.go`, `makeStatInheritance`. When `core.AssumeOk("lock_pet_crit")`, add the owner's SpellCrit to the demon's MeleeCrit and SpellCrit.
- **`lock_stone_with_oil`: do Firestone or Spellstone and Wizard Oil stack?** (all four warlocks)
  - The sim stacks them (`sim/warlock/talents.go`: the stones use the shaman imbue and rogue poison enchant effect in client 70338). No beta test.
  - Low: one or the other. Expected (default): both. Swing: about −16 DPS a build (−1.5%) at Low (sim-status.md).
  - Knob: a request edit `no_weapon_oil` in `run_assumptions.py` that takes the Wizard Oil off the main hand consume.
- **`bleed_tick_crit`: how hard bleed ticks crit** (Feral, Arms, Arms vs Dragonkin, Subtlety and Subtlety vs Humanoid)
  - Client 70425 gives Rip, Rake, Rend, Rupture and Garrote the periodic-crit attribute (SpellMisc Attributes_8 0x200). The sim crits them as melee, at 200%.
  - The only tick crits measured are spell ticks, at 1.5x (the shaman's Flame Shock).
  - Low: 150%. Expected (default): 200%.
  - Measured at Low (10k iterations, seed 1, 180 s): Feral 713.8 → 690.9 (−3.2%), Subtlety 634.6 → 624.9 (−1.5%), Arms 944.4 → 940.7 (−0.4%), Arms vs Dragonkin 982.4 → 978.7 (−0.4%).
  - Knob: `sim/core/spell_outcome.go`, `Dot.OutcomeTick`. For a Physical dot, multiply by `1 + (CritMultiplier − 1) × core.Assume("bleed_tick_crit", 1)` (0.5 for Low).

### `pal_soc_ppm` (pending, added 2026-10-16 by the assumptions audit)

- **Builds:** the twisting presets (Level 60, sub Holy, their Crusader duty copies, vs Undead and sub Holy vs Undead). The Shock builds never use Seal of Command.
- **Sim now:** 7 procs a minute with a 1 sec cooldown (`SealOfCommandPPM`, `sim/paladin/soc.go`).
- **Source:** the client (1.60.1.70425, the aura options of 20375 and 20915 to 20920) gives the aura a 100% chance on melee hits, a 1 sec recovery and no procs-per-minute entry, so the server's script sets the rate, as in Classic. Classic and Season of Discovery use 7 a minute. No Forever beta log has counted the procs.
- **Settings:** Low 5 a minute, Expected 7 (Classic), High 9. Low and High are a symmetric bracket two procs either side, not measurements.
- **Sensitivity (paired 20k, 180 s, seeds 1 and 1000003):** Low −36 (−3.9 to −4.4%), High +35 (+3.8 to +4.3%) on every twist build. Shock 0. At Low, Shock's lead over twist at 180 s grows from 2.9% to 7.6%; at High twist is 1.2% ahead.
- **Knob (to add):** `ppmm := paladin.AutoAttacks.NewPPMManager(core.Assume("pal_soc_ppm", SealOfCommandPPM), core.ProcMaskMelee)` in `registerSealOfCommand`.

## Settled, not a scenario

- **Off-hand factor of normalised Rage: ×0.5** (`warr_rage_oh`, Expected ×1.0 until 2026-10-16; the user ruled "Use 0.5, and remove other assumption values (keep assumption but locked to one value, 0.5)", so it stays in the report as a one-value row). Standards 21 kept 1.0 "until a test shows otherwise": a beta combat log with advanced logging (Arlewow's gist https://gist.github.com/Arlewow/a622100ad149f4fc31ddf95f9b3aab2c, linked from magey/forever-warrior issue 3, 20 September, level 13 warrior with a 1.8 sec main hand and 1.6 sec off hand dagger) shows 6.2 to 6.3 Rage a main hand swing (3.46 × 1.8) and 2.7 to 2.8 an off hand swing (3.46 × 1.6 × 0.5), glancing blows included, on about 50 and 60 clean swings (assumptions audit, 2026-10-16). The sim's `core.ForeverNormalizedRageOffHand` is 0.5 now; Dual Wield Specialization's 10% a point stays on top. Dual wield Fury 1055.0 → 961.1 at 180 sec (20k, seed 1; −94.4 on seed 1000003), the other warrior builds unchanged.
- **Extra Rage from critical white hits: +100%** (was `warr_crit_rage`, removed 2026-10-16 by the assumptions audit, orchestration/assumptions_audit.md).
  - Blizzard's 1 October beta development notes (us.forums.blizzard.com/en/wow/t/2360696, post 4, Warrior): "Players now generate 75% increased Rage when landing a critical strike with a basic attack", then in the same post "Additional Rage generated from landing Critical Strikes increased to 100% increased Rage (was 75%)". Kaivax's 2 October post (t/2369360) says the same.
  - The 8 October notes (post 5, Druid) raised Bear Form's crit Rage to the same 100% ("Was 75%").
  - The 75% Low was Blizzard's own superseded first number, so nothing is left to support it. The sim already gives +100% (`sim/warrior/warrior.go`, `AddCritDealtRageMultiplier(2)`), so no DPS moves.
- **Flat strike bonuses go inside the weapon percent** (was `flat_bonus_inside_pct` and `druid_shred_flat`, removed 2026-10-16 by the assumptions audit). Backstab, Ambush, Mutilate and Holy Strike carry their flat part on a normalized weapon damage effect (121), Shred and Claw on a weapon damage effect (58), each followed by a weapon percent effect (31); the server sums the flat part into the weapon damage and then takes the percent, which is what the sim does.
  - Measured: forever-bugs #208 (https://github.com/ClassicWoWCommunity/forever-bugs/issues/208, 2 to 7 October). A level 22 druid's rank 1 Shred hit 56 to 65 next to white hits of 19 to 24 on the same level 22 mob. Shred rank 1 is 155% weapon damage plus 24 in client 70338 (22 in 70425): (white + 24 × armor) × 1.55 gives about 56 to 64 at the mob's ~28% armor reduction, while 1.55 × white + 24 × armor tops out at 54. AidanZMoon (Blizzard) answered "it's a tooltip bug for the flat damage".
  - Blizzard's own text: Forever's Mutilate tooltip (1310707 and its ranks, unchanged in 70425) computes the flat part as `${$1310706m1*$1310706m2/100}`, the flat times the 75%. Client 70425 replaced Shred's stale "plus 180" (1.12's 80 × 2.25) with the raw `$s1`, and the 15 October notes call the Shred tooltip fix "just a visual bug. The damage is unchanged." So Forever's "plus N" tooltips print the flat part before the percent and aren't the damage rule.
  - Removed settings: Backstab / Ambush / Mutilate / Holy Strike "added after (tooltip)", and Shred +80 after and +180. The +180 has no source left now that the tooltip reads 80.
- **Seal of Righteousness's spell power: once, 12% a hit with a two-hander and 7.65% with a one-hander, flat per swing** (was `pal_sor_sp_double` and `pal_sor_sp_speed`, removed 2026-10-16, PR #54). Changed 2026-10-23 for the one-hander: 10%, and a weapon factor of 0.85 - 0.2 x max(0, 2.0 - speed), as wowsims/forever measured (49f94c9326: one-handers at 1.8, 2.2 and 2.8 fit 0.096 to 0.101 at 0 and 50 spell power, so the tooltip's 7.65% is wrong; 18 weapons from 1.7 to 3.7 for the factor). The presets hold two-handers, so no report number moves.
  - Blizzard's 15 October beta notes: "Fixed an issue causing Seal of Righteousness to gain almost twice as much benefit from Spell Damage than intended."
  - The client 70425 seal tooltip agrees: description variables 1023, `$2h = 0.12*$SPH` and `$1h = 0.0765*$SPH`, added once to the s3/87 to s3/25 base. Wowhead's Forever tooltip reads "(2068/87 + 0.0765 × SP) to (2068/25 + 0.12 × SP)".
  - It depends on one hand or two, not on the weapon's speed. Improved Seals multiplies all of it, and the target's damage taken bonus counts once.
  - The sim used to count 2 × 11% (Season of Discovery's tests). That was 46 to 51 DPS more at 180 s, and the proc spell's 0 would be 53 to 59 less.
- **Consecration's coefficient: 4.2% a tick** (was `pal_consecration_coef`, removed 2026-10-16, PR #54). Client 70425's damage spells (1280345 to 1280353) and Blizzard's 15 October notes both give 0.042. The 9.5% from 70338 and the old tooltip was a bug.
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
- **Rip and Ferocious Bite attack power** (server side; `sim/druid/rip.go:53`, `ferocious_bite.go:118`): no sensible spread. Since 2026-10-23 Rip's ticks read attack power live (wowsims/forever's rip.go). The preset rips at 4 combo points, where 1.12's 24% and WotLK's 1% a point a tick (24% over 6 ticks) agree. Bite is 7% of the damage.
- **Ignite: no cap (Season of Discovery) or a stop after 5 crits (Era).** Five crits inside one 4 sec window is rare at the presets' crit, under 0.5% for Fire and Frostfire.
- **Premeditation out of Stealth** (Subtlety, −7.7 if it needed Stealth). Client 14183 dropped the stealth-only attribute, and ForeverChanges reads it the same way. Preparation's reset list is Classic's.
- **Deadly Poison re-snapshot on each application.** It measured +6.3 on an old Mutilate set, but what still feeds it on the final builds is unclear. Likely under 0.5%.
- **Percent talent modifiers: add or multiply** (rogue Improved Eviscerate × Aggression, Opportunity × Aggression; the warlock adds). +0.7 on Sinister Strike; under 0.5% everywhere.
- **Whirlwind's off-hand strike.** Blizzard's 1 October notes made it baseline; only its exact form (normalized off hand, 50%, Dual Wield Specialization) is server side, about 1% of Fury in total.
- **Weaponmaster's 200 ms internal cooldown, Heroic Strike counting for Unbridled Wrath.** All small or unknowable without a log. (Deep Wounds' tick timer was on this list until 2026-10-23: wowsims/forever's log, c4a5402aa9, shows a refresh keeps it.)
- **Dark Sacrifice's mana** (Undead priests). Shadow never runs out of mana; it only moves the Smite race table, where Undead leads the Gnome by 5.5.
- **Shadow Word: Death, Demonic Pact, Wrack, Incinerate's base, Sniper Shot, Light's Vigil, the other 31-point capstones.** Forever's own values, untested, with no analogue to give a range.
- **Arcane Shot and Aimed Shot cooldown sharing, Auto Shot through casts, Expose Prey, Lacerate, Lacerating Strikes.** Beta-tested or client data, or under the threshold.
- **Elemental Focus on low ranks, Searing Totem's reach from range, Elemental Devastation's size** (the client curve). All under 0.5%.
- **Downrank penalty size at 60** (was `downrank_penalty`, removed 2026-10-15). Forever cuts the spell power of a rank learned more than 17 levels below you (foreverchanges' 5% a level, measured on heals to level 30). Only Smite's and Elemental's low-rank fallbacks could feel it, and all three settings (10% a level, 5%, none) gave the same DPS on all three builds at 120, 180 and 300 s. Smite's fallback is rank 7, learned at 46, inside the 17 levels. The priest's spells keep the 5% rule (`downrankCoefficient`, `sim/priest/priest.go`). Reopen it if a preset starts casting a rank learned before level 43.
- **Tiger's Fury's scope** (was `druid_tf_scope`, removed 2026-10-15). Forever removed Tiger's Fury (Blizzard's Druid deep dive; nothing in client 70338 grants it).
- **Arcane Blast's stack bonus on Arcane Missiles** (was `mage_ab_on_missiles`, removed, user ruling 2026-10-15). The sim keeps the tooltip's 10% a stack on Missiles (`ArcaneBlastStackBonus`, `sim/mage/arcane_blast.go`). The client's stack buff 400573 leaves Missiles out, and Season of Discovery's 400586 (15% a stack) is still in the client; the scenario's settings were 0 / 10% / 15% (−9.48% / 0 / +4.74% on Arcane in its last assumption sims).
- **Decimation's damage part** (was `lock_decimation`, removed, user ruling 2026-10-15). The sim keeps the tooltip's 3% a point under 35% health (`sim/warlock/talents.go`); the client's effect holds 0, which measured −1.08% on Demonology in the last assumption sims. Since 2026-10-23 it is a factor on the two spells' hits from the moment the fight reaches 35% (the SPELLMOD_DAMAGE row wowsims/forever reads, 3 / 6% in the 70338 trait curve), the same numbers as the target-side multiplier it was.
- **Settings dropped in the same review** (user ruling 2026-10-15): Ice Lance's 0.143 (WotLK; the client's 0 came back as Low in a later ruling the same day), the demon's "Nothing" and "SoD plus crit" scaling, Incinerate's "base damage only", and Omen of Clarity's 6% a spell and 3.5 PPM. `lock_incinerate_bonus` and `druid_ooc_rate` stay as one-value reference rows.
- **Shadow Oil and Frost Oil proc damage** (was `oil_proc_damage`, removed, user ruling 2026-10-15). The sim keeps Forever's tooltip damage (Shadow Oil 63 to 71, Frost Oil 39 to 43; `sim/core/consumes.go`). Era's ranges, which the server formula with the spells' unchanged SpellLevel would give, measured −0.06 to −0.08% on the Retribution and Enhancement builds.
- **Arcane Shot and Serpent Sting attack power shares** (was `hunter_shot_ap_shares`, removed, user ruling 2026-10-15). The sim uses the forever-hunter wiki's 10% and 3% a tick (`sim/hunter/arcane_shot.go`, `serpent_sting.go`; PR #22). The combat-log fit's 11% / 3.5% is within the fit's ±10%.
- **Settings dropped in the same pass** (user ruling 2026-10-15): Consecration's "split parts" 10.55% and Maelstrom Weapon's 20 PPM (replaced by 15 PPM).
- **The encounter itself** (level 63, 3731 armor, Humanoid). This moves every spec, but it is an encounter setting, not a spec assumption.

## Where no sensible range was found

These entries have only two sensible settings (High = Expected): either the evidence gives two readings and nothing beyond them, or the higher value is the sim's own.
- `wf_totem_offhand`, `priest_sf_tick_crit`, `pal_2h_spec_holy`, `pal_31plus_ranks`, `pal_soc_wf`, `hunter_pet_speed` (no Low: Classic is Expected), `hunter_crit_per_agi`, `lock_pet_scaling`, `lock_demon_melee`, `lock_bod_coef`, `mage_imbue_scrolls` and `druid_rip_31plus` (no Low: the client is Expected). `hunter_tsa_hm_rank` and `mage_ice_lance_coef` have two settings below Expected and none above. `lock_incinerate_bonus` and `druid_ooc_rate` have one value.

Weakly grounded values:
- `sham_mw_ppm` Low (5 PPM): no analogue below 10.
- The halfway "Expected" of the rank-tuning entries (`priest_smite8_tuning`, `hunter_raptor_31plus`, `mage_frostbolt10`, `sham_es7_tuning`, `druid_wrath8_tuning`): a judgement between "client as is" and "held at the ≤30 ratio".
- `mage_ice_lance_coef` Mid (half of SoD): a judgement, with no source of its own.
- `lock_pet_scaling` Low (half of SoD): the same.

No sensible range at all: Rip and Ferocious Bite attack power, and the untested 31-point capstones (listed above).
