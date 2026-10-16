# Assumptions audit (2026-10-16)

Every assumption the Level 60 DPS report and the sim maintain, checked against the newest evidence. Sources, strongest first:
1. Client data at 1.60.1.70425 (the newest build on wago.tools; compared with 70338): SpellEffect, SpellMisc, SpellAuraOptions, Spell (descriptions), SpellDescriptionVariables, TraitDefinitionEffectPoints and CurvePoint.
2. Blizzard's words. The "WoW Forever Beta Development Notes" thread (https://us.forums.blizzard.com/en/wow/t/2360696): post 1 (24 September), post 4 (1 October), post 5 (8 October), post 6 (15 October, mirrored on wowhead as "Massive Racial Updates"). Also Kaivax's 2 October warrior post (t/2369360), the class deep dives, and AidanZMoon's replies on the community bug tracker (ClassicWoWCommunity/forever-bugs).
3. Measurements:
   - the user's beta logs, through beta_results.txt, beta_log.txt and shaman_audit.md (no raw logs are on disk);
   - community logs and threads: magey/forever-warrior issue 3 and the advanced combat log in Arlewow's gist, forever-bugs #208 and #331, EU forum thread 632244.

Tier 3 precedents (SoD, TBC, Wrath) don't settle anything here.

Coordination: the update auditor (session_01DLHrNhMbEVhmSed2oDxNRS) is building `orchestration/update_ledger.md`, which checks every announced game change against the client and the sim. It is now in PR #79 (branch claude/wow-forever-dps-presets-m9u15h-updates-ledger). Its verdicts on the report rows agree with this audit's: warr_crit_rage, Seal of Righteousness once and Consecration 0.042 are settled; the top ranks, Ice Lance, Bane of Doom, the pet scaling auras, the proc rates, Rend and Windfury Totem are unchanged in 70425. One addition: the client's hidden "Rule of Rage (DND)" 1322574 went from 10 to 100 base points in 70291, which also backs crit Rage at +100%. Its PR #69 covers Touch of the Grave and Gnome Eureka!.

Verdicts:
- **S=** settled, and the sim already matches;
- **S!** settled, and the sim must change;
- **N** narrowed;
- **O** still open;
- **X** obsolete.

"Routed" means the change belongs to a class worker with an in-flight 70425 port (mage, shaman, hunter, druid, paladin #54). The orchestrator routes it, and this audit leaves those files alone.

## Summary

| Verdict | Report assumptions (36 at the start) | Other items (notes.md, code) |
|---|---|---|
| S= settled, sim matches | 4: `flat_bonus_inside_pct`, `druid_shred_flat`, `warr_crit_rage`, `pal_sor_sp_speed` | 12 |
| S! settled, sim must change | 3: `warr_rage_oh` (0.5), `pal_sor_sp_double` (once: 0.12 / 0.0765), `pal_consecration_coef` (0.042) | 8 |
| N narrowed | 3: `wf_totem_offhand`, `hunter_pet_speed`, `lock_pet_scaling` (new evidence, all pointing away from the current default; user rulings needed) | 3 |
| O still open | 26 | the rest of notes.md Need to Verify |
| X obsolete | 0 | 1 |

PRs:
- **#67** (warrior): `warr_crit_rage` settled, no DPS change.
- **#75**: `warr_rage_oh` settled at 0.5, so the sim changes: dual-wield Fury goes from 1055.0 to 961.1 at 180 s. Split from #67. The user ruled on Q6: "Use 0.5, and remove other assumption values (keep assumption but locked to one value, 0.5)".
- **#70** (priest and warlock): Power in Light counts Shadow Word: Pain (Smite +6.2), plus four notes items.
- **#72** (flat bonus): `flat_bonus_inside_pct` and `druid_shred_flat` settled, no DPS change.
- **#54** (paladin worker): takes `pal_sor_sp_double`, `pal_sor_sp_speed` and `pal_consecration_coef`, routed 2026-10-16.
- **This PR** adds the ledger and one new pending assumption (`level_partial_resist`).

## 1. Report assumptions (tools/dps_report/assumptions_def.py)

| Key | What is assumed (default) | Evidence found | Verdict | Action |
|---|---|---|---|---|
| `flat_bonus_inside_pct` | The flat part of Backstab, Ambush, Mutilate and Holy Strike sits inside the weapon percent (inside) | (a) forever-bugs #208: a level 22 rank 1 Shred hit 56–65 next to 19–24 white hits. Only "inside" fits: (w + 24 × 1.55) × armor ≈ 56–64, while "after" tops out at 54. AidanZMoon: "it's a tooltip bug for the flat damage". (b) Forever's Mutilate tooltip computes `${$1310706m1*$1310706m2/100}` (flat × 75%). (c) 70425 replaced Shred's hard-coded "plus 180" with `$s1`; the 15 October note says "just a visual bug. The damage is unchanged". (d) Effects 58 and 121, each followed by 31, take the same server path. | S= | PR #72: row removed |
| `wf_totem_offhand` | Off-hand hits proc Windfury Totem (yes) | 70338 and 70425 description of 8512 / 10614: "Each main hand hit has a 20% chance". 70425 edited that very sentence's neighbour ("party or raid members") and kept "main hand". The proc mask (0x14) names no hand. No log (needs level 32). | N, toward main hand only | **User question Q2** |
| `priest_sf_tick_crit` | Shadowform's +100% crit bonus reaches tick crits (200%) | Ticks crit (settled, see PR #70). 70425 adds Shadow Word: Death to Shadowform's mask; the 15 October note calls that a fix. No tick-crit size has been measured. | O | none |
| `priest_smite8_tuning` | Smite rank 8 ×1.09 | 10934 is unchanged in 70425 (17 builds unchanged). The 8 October note retuned ranks 1–4 only. | O | none |
| `lock_pet_scaling` | SoD shares, no crit | 416189's amounts are still 0 in 70425 (only effect 8 changed, aura 22→674). AidanZMoon on forever-bugs #331 (13–14 October): demons "should be getting their master's spell crit as their own melee AND spell crit chance… we aren't enabling the scaling until the baseline critical strike chance is removed". | N: launch intent includes crit | **User question Q4**: bring back "SoD + spell crit" (last measured +2.1 to +2.4%) |
| `lock_demon_melee` | 1.12 melee, 95–131 | No source. A wowhead comment on the 15 October notes says demon damage was "gutted"; nothing confirms it (no note, no bug, no measurement). | O | none |
| `lock_bod_coef` | 4.0 | 603 is unchanged in 70425: 1742 + 4.0 × spell power. | O | none |
| `lock_incinerate_bonus` | +25% on the whole hit (reference row) | The deep dive says "an additional 25% damage if the target is afflicted by Immolate". It's consistent, but doesn't say whether spell power is covered. | O | none |
| `pal_sor_sp_double` | Spell power counted twice (2 × 11% on a two-hander) | The 15 October note: "Fixed an issue causing Seal of Righteousness to gain almost twice as much benefit from Spell Damage than intended." 70425: 20293's description uses `$<1h>` / `$<2h>`, which SpellDescriptionVariables 1023 defines as `$2h=${0.12*$SPH}`, `$1h=${0.0765*$SPH}`. The proc spells' own 0.1 went to 0. | S!: once, 0.12 two-hander / 0.0765 one-hander | Routed to #54 |
| `pal_sor_sp_speed` | Spell power share flat per swing | Same variables: the share depends on one hand vs two hands, not on speed. | S= | Routed to #54 (row removal) |
| `pal_consecration_coef` | 9.5% a tick | 70425: 1280345–1280353 are at 0.042. The 15 October note: "changed to 0.042 (was 0.095)". | S! | #54 already ports 4.2% |
| `pal_2h_spec_holy` | Two-Handed Spec boosts the Holy seal damage | 20111 is unchanged in 70425 (Physical mask). | O | none |
| `pal_31plus_ranks` | Client values for Holy Strike's flat part and Consecration's 31+ ranks | Holy Strike is unchanged; Consecration's coefficient changed (above), its base didn't. | O | none |
| `pal_soc_wf` | Seal of Command hits proc Windfury Totem | No source | O | none |
| `warr_rage_k` | k ×1.5 at 60 | Every measurement is from level 8–30. The advanced log below confirms 3.46 × speed at level 13. | O | none |
| `warr_rage_oh` | Off hand ×1.0 | Arlewow's advanced combat log (https://gist.github.com/Arlewow/a622100ad149f4fc31ddf95f9b3aab2c, linked from magey/forever-warrior #3, 20 September, level 13): main hand 1.8 s gives 6.2–6.3 Rage a swing; off hand 1.6 s gives **2.7–2.8** (3.46 × 1.6 × 0.5). That's about 53 and 61 clean swings, glances included. Standards 21 kept 1.0 "until a test shows otherwise". | S!: 0.5 | PR #67 (Fury −93.4 / −94.4 at 180 s) |
| `warr_crit_rage` | +100% | The 1 October notes (post 4): "Additional Rage generated from landing Critical Strikes increased to 100% increased Rage (was 75%)". Kaivax repeats it on 2 October; the 8 October notes give Bear Form the same. | S= | PR #67: row removed |
| `warr_rend_ap` | 2% of attack power a tick | No number anywhere (deep dive: "significant"). 11574 is unchanged. | O | none |
| `warr_dt_rage_mods` (pending) | None of our modifiers count | No new source | O | none |
| `warr_dt_rage_k` (pending) | K = 20 | Only the pre-8-October fits exist (magey #3; forever-bugs #72's damage × 10 / max health). | O | none |
| `hunter_pet_scaling` | 10% of ranged AP plus 1:1 hit | No AP or hit number. The 15 October notes say pets now inherit the higher of melee and ranged crit, which belongs to the hunter port, not this row. | O | Crit change routed (hunter) |
| `hunter_pet_speed` | Classic: hit size scales with speed (the beta's same-size hit treated as a bug) | EU forum thread 632244 (28 September), level 16 Tallstriders: a 2.0 s pet hits 31–42 and a 1.3 s pet 27–38, so the per-hit damage isn't speed-scaled. forever-bugs #31 is still open, and no note through 15 October fixes it. | N: the beta behaviour persists | **User question Q3** |
| `hunter_tsa_hm_rank` | Top ranks fixed at +100 / +140 | 70425 is unchanged: Trueshot Aura 30 / 40 / 50 / 75 / **50**, Hunter's Mark 26 / 59 / 98 / **71**. Trueshot Aura is now raid-wide (effect 35→65; note "Trueshot Aura is now Raid wide"). The deep dive's "new lower-level ranks added, but otherwise unchanged" fits a Classic top rank of 100. | O | none |
| `hunter_crit_per_agi` | 1.22 × Classic | No data above level 1 | O | none |
| `hunter_raptor_31plus` | Rank 8 at 96 | 14266 is unchanged | O | none |
| `mage_ice_lance_coef` | 42.9% | Still 0 on all six ranks in 70425. | O | Routed: the frozen bonus changed 300→200 in 70425 (note: "changed to a 200% increase"). The sim's ×4 must become ×3, and Fingers of Frost charges count as frozen. |
| `mage_frostbolt10` | ×1.04 | 25304 / 10181 are unchanged | O | none |
| `mage_imbue_scrolls` | Available | method.gg (12 October, datamined): Comprehension 175 and player level 46. No launch statement. The 15 October "Scrolls now incur a global cooldown" doesn't touch a pre-pull imbue. | O | none |
| `sham_mw_ppm` | 10 PPM | 408498 / 408505 still read 100 (server). The deep dive says "a chance on every melee attack". | O | none |
| `sham_lb_31plus` | ×1.12 | 15208 is unchanged | O | none |
| `sham_es7_tuning` | ×1.08 | 10414 is unchanged | O | none |
| `raid_hits` | 2 a minute at 1500 | Nothing Forever-specific | O | none |
| `druid_ooc_rate` | 2 PPM | 16864 is unchanged (server). The 15 October note only stops Faerie Fire (Feral) from triggering it. | O | none |
| `druid_shred_flat` | +124 (80 inside the 155%) | As for `flat_bonus_inside_pct`; this is the measured spell. 70425's tooltip now reads "plus 80", so +180 has no source. | S= | PR #72: row removed |
| `druid_rip_31plus` | Client | 9896 is unchanged | O | none |
| `druid_wrath8_tuning` | ×1.085 | 9912 is unchanged | O | none |

The DASH_ASSUME knobs that die (the orchestrator rebuilds patch 0004):
- `warr_crit_rage`;
- `warr_rage_oh`: its row stays with the one setting 0.5, so the knob's default must be 0.5 and no setting overrides it;
- `flat_bonus_inside_pct`;
- `druid_shred_extra`;
- the paladin ones listed by #54: `pal_sor_sp_double`, `pal_sor_sp_speed`, `pal_consecration_coef`.

## 2. New or changed assumptions (proposals)

### 2a. `level_partial_resist`: added in this PR with status 'pending'
The sim adds Classic's 2% average partial resist per level the target is above the caster to every non-binary spell, Forever included: 6% against the level 63 boss (`sim/core/spell_resistances.go:166`, `spell.go:627`, `AverageMagicPartialResistPerLevelMultiplier`). notes.md (Shaman, "partial resists on dungeon bosses") says the sim drops it under Forever, but the code doesn't. shaman_audit 5.2 proposed dropping it, and that proposal still waits on the user.

Evidence: 0 partial resists in 313 landed non-binary hits on normal mobs 3 to 5 levels up (shaman_audit 5.2; Classic would give about 57). Nothing from bosses or elites.

Measured (10k iterations, seed 1, 180 s, today's master), DPS with 6% → with none:

| Build | 6% (now) | None | Δ |
|---|---|---|---|
| mage__fire | 868.8 | 930.4 | +7.1% |
| mage__arcane | 819.9 | 871.6 | +6.3% |
| mage__frost | 942.3 | 942.5 | 0 (Frostbolt is binary) |
| warlock__destruction | 981.3 | 1032.6 | +5.2% |
| warlock__affliction | 998.0 | 1050.8 | +5.3% |
| shaman__elemental-level60 | 684.5 | 728.4 | +6.4% |
| priest__shadow | 928.5 | 963.6 | +3.8% |
| druid__balance | 828.2 | 872.8 | +5.4% |
| paladin__level_60 (before #54) | 975.7 | 1016.8 | +4.2% |
| shaman__enhancement-level60 | 831.1 | 856.1 | +3.0% |
| warrior__fury | 1054.2 | 1055.0 | +0.1% |

```python
 dict(key='level_partial_resist', cls='All', label='Do bosses partly resist spells by level?', status='pending',
      short="Classic bosses shrugged off about 6% of non-binary spell damage. Forever's beta mobs above our level resisted none.",
      desc="In Classic every target above your level partly resists your non-binary spells: 2% of their damage on average per level, 6% against a level 63 boss. On the Forever beta, 313 hits on normal mobs 3 to 5 levels up showed no partial resist at all. Nobody has tested a boss. Binary spells such as Frostbolt are unaffected either way.",
      src="Beta logs (shaman_audit.md 5.2: 0 partial resists in 313 hits at +3 to +5); Classic's level rule (royalgiraffe's resist guide). No boss data.",
      builds=[b for b in ALL_BASE if not b.startswith(('rogue__', 'warrior__', 'hunter__'))],
      settings=[S('low', 'Classic (6% at +3)', '2% average partial resist per level above you', "Classic's rule, which the sim uses now.", {'level_partial_resist': 0.02}),
                S('expected', 'None (beta)', 'no level-based partial resist', 'What every beta mob above our level showed.', {'level_partial_resist': 0})],
      default='low', expected='expected'),
```

The knob is `core.Assume("level_partial_resist", AverageMagicPartialResistPerLevelMultiplier)` at its two uses. The default is the sim's current behaviour until the user rules (**Q1**).

### 2b. More proposals
Added as pending rows on 2026-10-16 (the orchestrator's request): `lock_pet_crit` (as its own row rather than a setting on `lock_pet_scaling`, so that row keeps simulating), `lock_stone_with_oil` and `bleed_tick_crit`. Measured at 150%: Feral −3.2%, Subtlety −1.5%, Arms −0.4%. spec_assumptions.md has the knob locations. Deep Wounds' double dip went to the warrior worker.

- **`lock_pet_scaling` High: "SoD + the warlock's spell crit (Blizzard's intent)".** AidanZMoon (forever-bugs #331) says demons should inherit the warlock's spell crit as melee and spell crit once the baseline crit is removed. Last measured at +2.1 to +2.4%. The setting was dropped on 2026-10-15; this is new evidence (**Q4**).
- **Bleed tick crit size (`bleed_tick_crit`).** The sim crits bleeds (Rip, Rake, Rend, Rupture, Garrote) for 200% as melee. 70425 confirms they can crit (SpellMisc Attributes_8 0x200 on all of them, plus 0x1000 on the physical ones), but not the size. The only measured periodic crit is Flame Shock at 1.5×.
  - Settings: Low "150% (spell-like)" and Expected "200% (melee)".
  - Feral (Rip about 23% of its damage) would move about −2.5% at Low; Arms (Rend) under 1%.
  - Worth a row if the Feral ranking is close.
- **Spellstone / Firestone with Wizard Oil (`lock_stone_with_oil`).** The sim stacks them (`sim/warlock/talents.go:92`, reasoning from the imbue slot type). That's worth about 15.7 DPS (about 1.5%) on all four warlocks. No client or beta source either way. Settings: Low "one or the other" and Expected "both".
- **Deep Wounds double-dipping the damage modifiers** (`sim/warrior/deep_wounds.go:44, 87`): Classic behaviour kept under Forever, 0.3–1% on the warriors. A candidate row for the warrior worker. Below the 0.5% bar for most builds.
- **Seal of Command at 7 PPM with a 1 sec ICD** (Classic's, server-side; `sim/paladin/soc.go`). Medium for the twist builds. A paladin worker call.

## 3. Things the evidence settled that the sim gets wrong, outside the report rows (routed)
- **Grace of Air now stacks with Windfury Totem.** The 15 October notes: "Grace of Air Totem is no longer exclusive with Windfury Totem and Flametongue Totem". The sim still switches Grace of Air off while a Windfury Totem buff is up (`sim/core/buffs.go`, `graceOfAirWindfuryCategory`, "For now Forever doesn't let the two stack").
  - With raid-wide totems, a raid with two shamans gives every melee both.
  - The raid composition is a user ruling (**Q5**). Shaman port and core.
- **Touch of the Grave** (covered by the update auditor's PR #69). The 15 October note says "Touch of the Dead no longer scales with Shadow increases on the target or the caster". The sim's drain still takes every Shadow modifier (`sim/core/racials_forever.go`, DamageMultiplier 1 and nothing ignoring modifiers). The chance and 25 s ICD were already ported. Small with the 25 s ICD. Racials / core.
- **Ice Lance against frozen targets:** ×3 (+200%), not ×4. Mage port.
- **Elemental Devastation** now procs on "Critical Strikes with direct damage spells" (70425 tooltip). Shaman port.
- **Pet crit:** the hunter's higher of melee and ranged crit (15 October note). The update ledger notes the sim has one crit stat, so it already behaves this way; nothing to port.
- **Smite rotation:** with Power in Light on Shadow Word: Pain (PR #70), dropping the Holy Fire line gains 4.7. Priest worker.
- **Doc fixes for routing:**
  - notes.md Druid "Shred's flat bonus" and Paladin "Holy Strike's damage" are settled by PR #72's evidence.
  - notes.md "At 60" says no preset takes Lightning Overload, but both Elemental presets do.
  - notes.md Hunter "Battle Shout 139" vs the sim's 115 (AQ rank excluded).

## 4. notes.md "Need to Verify": items with new evidence
Items not listed are still open with no new evidence. Verdict, then the evidence.

**All classes**
- Periodic crits (Priest section): **S=**. 70425 Attributes_8 0x200 on player DoTs; 1 October notes on Devouring Plague and Inner Focus; shaman tick crits at 1.5×. PR #70.
- Partial resist at +3: see 2a. **O**, measured cost above, Q1.
- Glancing: the user's logs fit the sim at +2 / +3, and +4 breaks the 40% cap (beta_results.txt). Narrowed, and the user already ruled "good enough". **O**.
- Spirit regen past 50 Spirit for non-shamans: no new data. **O**.
- Troll Regeneration is 20% (15 October). Leveling sim only; noted for the leveling owner.

**Warrior**
- Crit Rage: **S=** (PR #67).
- Off-hand Rage: **S!** (PR #67).
- Glancing blows give full normalised Rage: **S=**. Arlewow's log shows main-hand glances at 6.2–6.3 and off-hand glances at 2.7–2.8, the same as hits (followups step 4, glancing part).
- Weaponmaster's 200 ms ICD: **S=**. 70425 SpellAuraOptions for 12281 has ProcCategoryRecovery 200, chance 5.
- Dual Wield Specialization off-hand Rage is 10% a point: **S=**. 70425 trait curve 111755 (10 to 50); the 1 October note removed the old 20 to 100%. The sim keeps it on top of the halved off hand.
- Rend AP, Deep Wounds rolling, Whirlwind off hand: **O**.

**Priest** (PR #70)
- Shadow Weaving is caster-only: **S=**.
- Power in Light: **S!**, fixed.
- Devouring Plague's 24 s and Shadow Word: Death's 18 s cooldowns were already ported (#58).

**Warlock**
- Demonic Embrace's Spirit: **S=**. 70425 removed the effect (PR #70).
- Demon scaling: see `lock_pet_scaling`, **N**.
- Shadow Mastery 2% a point was already ported (#59).

**Paladin** (routed to #54)
- Seal of Righteousness spell power: **S!**.
- Consecration: **S!**.
- Holy Strike's flat part inside: **S=** (PR #72 evidence).
- Seal of Righteousness proc base points: **X**. The proc's coefficient moved into the aura's tooltip variable, so the question no longer bears on spell power.
- Improved Seals: already settled.

**Hunter**
- Trueshot Aura party vs raid: **S=**. Raid-wide in 70425 and the 15 October note; the sim's raid switch now matches.
- Auto Shot fires while casting: **S=**. The 15 October developers' note: "Your character will still fire Auto Shot while casting spells like Aimed Shot or Sniper Shot".
- Pet crit source: **S!** (routed, the higher of melee and ranged).
- Furious Howl stacking with Battle Shout: **O**. Both are raid area auras (target 56) on aura 99 in 70425.

**Mage** (routed)
- Ice Lance frozen bonus: **S!**, ×3.
- Arcane Blast, Missile Barrage (chance 50 → 100 in SpellAuraOptions), Fire Power, Winter's Chill: redesigned in 70425, so the old notes items are superseded. Port.

**Shaman** (routed)
- Grace of Air / Windfury stacking: **S!** (Q5).
- Fire Nova coefficient 0.143: port.
- Elemental Devastation direct spells only: **S!**.
- Frostbrand: one other shaman's two hits fit no cut. **O**.

**Druid**
- Shred's flat part: **S=** (PR #72).
- Tiger's Fury removed: **S=** (Blizzard's 1 October notes, "Tiger's Fury has been removed").
- Omen of Clarity: **O**.

**Rogue**
- Backstab's flat part: **S=** (PR #72).
- Mutilate off hand, Preparation, Premeditation: no new evidence. **O**.

## 5. Questions for the user
- **Q1. Level-based partial resist (6% at +3).**
  - Keep Classic's rule as the default, or drop it under Forever as shaman_audit 5.2 proposed, given 0 partials in 313 beta hits at +3 to +5 on normal mobs?
  - Dropping it raises every caster 4 to 7% (Fire +61.6, Arcane +51.7, Destruction +51.3, Elemental +43.9) and leaves Frost and the physical classes where they are.
  - Either way the report gets the other as a setting.
- **Q2. Windfury Totem off hand.** The 70425 description, edited in that build, still says "Each main hand hit has a 20% chance". Keep "any melee hit" as the default, or switch to main hand only? Main hand only costs dual-wield Fury and the rogues about 2 to 3%, and Survival about 1.5% (estimate).
- **Q3. Hunter pet hit size.** A level 16 measurement (1.3 s pet 27–38 against 2.0 s 31–42), the open bug #31 and no fix in any note through 15 October say the beta's same-size hit persists. Keep treating it as a bug (Classic speed scaling), or make the beta behaviour the default? The beta rule is +6.1% Beast Mastery and +3.2% Marksmanship.
- **Q4. Demon crit.** Blizzard says demons will inherit the warlock's spell crit once a baseline bug is fixed. Bring back "SoD + spell crit" as a setting (about +2.1 to +2.4%), or as the default?
- **Q5. Grace of Air with Windfury Totem.** These now stack, and both are raid-wide. Should melee raids assume both totems (two shamans), or keep one Air totem?
- **Q6. Off-hand Rage at 0.5 (PR #67).** Standards 21 called for this switch once a test showed it, and the test exists (an advanced combat log, about 60 off-hand swings, level 13). It drops dual-wield Fury about 9%. Merge as is, or wait for a level 30+ log?

## 6. User rulings (2026-10-16, relayed by the orchestrator)
- **Q1, level partial resist:** "Can you research more into this? I believe some kind of partial resists exists, lowering damage of spells. This is something we need To Verify." The sim keeps Classic's 6% default. `level_partial_resist` stays pending; see section 7 for the research.
- **Q2, Windfury Totem off hand:** "Main hand only, and keep as assumption with only 1 value." Settled by ruling, done in PR #86 (the orchestrator's): main-hand procs only under Forever, `wf_totem_offhand` is a one-value row, and the rogue totem test is flipped.
- **Q3, hunter pet hit size:** "Keep treating it as a bug." `hunter_pet_speed` is unchanged: Classic's attack-speed scaling stays the default, and the beta's same-size hit stays the High setting.
- **Q4, demon crit:** "Yes, as Blizzard intends", locked to one value. Settled by ruling, merged in PR #85 (the warlock worker's): demons inherit the warlock's spell crit, and `lock_pet_crit` is a one-value row.
- **Q5, Grace of Air with Windfury Totem:** "both totems". The update auditor has it.
- **Q6, off-hand Rage:** "Use 0.5, and remove other assumption values (keep assumption but locked to one value, 0.5)". Settled by ruling, merged in PR #75: `warr_rage_oh` is a one-value row at 0.5.

## 7. Q1 research: partial resists under Forever (2026-10-16)
The user believes "some kind of partial resists exists, lowering damage of spells". What the evidence shows:
- **School-resistance partials exist, in 10% steps (narrowed).**
  - Level 29 Elder Cloud Serpents (Nature 58 in the 1.12 / AzerothCore data) took 20% or 30% off all 15 Nature hits that landed: 9 orbs (4 at 20%, 5 at 30%) and 6 Earth Shocks (5 at 30%, 1 at 20%) (beta_results.txt:724-730, 961-962; shaman_audit.md:590).
  - Classic's 0/25/50/75% table can't produce a 20% or 30% hit. WotLK's 11-bucket table (TrinityCore 3.3.5 `CalcSpellResistedDamage`) fits the steps, but predicts about 21% of hits at 40%, and none showed (p about 0.013). A model that picks the two steps around the average (28 to 29%) fits best.
  - Both Classic's and WotLK's average agree here (29% and 28.6%), so 15 hits can't tell the formulas apart.
- **Level-based partials on normal mobs: settled, none.**
  - About 520 landed hits 3 to 5 levels up, all with resisted = 0: 257 + 56 + 38 in shaman_audit.md 5.2 and beta_results.txt:300, plus the 19:06-19:32 runs (crit tables at beta_results.txt:687-692, orbs at 704 and 827).
  - Classic predicts 18 to 30% of them partly resisted, WotLK about 75%, so both are rejected at p < 1e-25.
  - The method reads `resisted` (row[-7]) on SPELL_DAMAGE, and the hits came in at full base damage.
- **Bosses and elites: open.** The repo has no dungeon, elite or boss logs.
- **Binary spells (narrowed).** On the serpents, Earth Shock was fully resisted 3 of 9 times (RESIST, not MISS) and partly resisted when it landed. That looks like a full-resist roll followed by partials, while the orbs had 0 full resists in 9. Frost Shock and Frostbrand are untested.
- **Blizzard and community: nothing on level or partial resists.**
  - The dev notes only touch single spells: Improved Scorch, Winter's Chill, Setup, and Vampiric Embrace "can't be resisted".
  - The deep dives turned resist-chance talents into hit talents (Elemental Precision, Shadow Focus).
  - forever-bugs has no relevant issue.
- **Client data:** no binary or no-partial attribute differs from Era (wago SpellMisc 70425 against 1.15.9 70003).
- **Forever's own resistance formula (found 2026-10-16).** Forever's character sheet gives each resistance a percentage, from a helper no other Blizzard client has (Gethe/wow-ui-source, branch `forever`, `Interface/AddOns/Blizzard_UIPanels_Game/Camelot/PaperDollFrameStats.lua:234`, present since the first mirrored beta build 69876 and unchanged through 70425):
  `ExpectedSpellResistance(casterLevel, resistance) = 0.75 * resistance / (casterLevel * 5)`, shown as "average resist against a caster of the unit's level".
  - It is Classic's average from school resistance alone. It has no term for the attacker being higher level, and Classic's level-20 floor is gone. Classic Era's sheet (`Vanilla/PaperDollFrameStats.lua`) shows only Poor / Fair / Good.
  - It fits both beta findings: partials only where the target has school resistance (the serpents), and none from level (0 in about 520 hits at +3 to +5).
  - Blizzard hasn't said why. The likely reading is a deliberate simplification in line with the deep dives, which turned resist-chance talents into hit talents. The level gap now costs spell hit (misses), and only real resistance causes partial resists. The 10% steps match Blizzard's later resist table rather than Classic's 0/25/50/75.
  - For a level 63 boss with no school resistance this predicts no partial resists, which is the `level_partial_resist` Expected setting. It's a UI helper, not server code, so a boss log still settles it.
- **Report row:** `level_partial_resist` stays pending. Low is relabelled "bosses keep it (6% at +3)", Expected is none, and the default stays Classic. No 10%-step setting: the steps change variance, not mean DPS.
- **Settling tests:** in notes.md (All classes / Shaman "Beta: partial resists").
  - Vishas and Thalnos in SM Graveyard: 50 or more landed hits each.
  - 200 Nature hits and 60 Earth Shocks against 60 Lightning Bolts on Elder Cloud Serpents.
  - 40 Frost Shocks on Alterac Yetis.
- Lead for the raw log: AzerothCore gives Swirling Vortex (L33-34) 34 resistance in four schools, yet the 20:18-20:25 run reports no partials. If our spells hit it, Forever changed its resistance.
