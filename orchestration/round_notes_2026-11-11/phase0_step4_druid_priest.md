# Phase 0 step 4: druid and priest DPS builds (2026-11-24)

Balance, Feral, Shadow and Smite, and every DPS druid and priest build of the report, moved from the synthetic sets to
the Phase 0 lists. verify.py, 180s ± 20s, 20,000 iterations, seed 1. Builds that share a spec use that spec's set.

| Build | Phase 0 | Synthetic | Set |
|---|---|---|---|
| druid__balance | 822.15 | 952.2 | level60_phase0_balance |
| druid__balance-duty-ff | 788.30 | 915.5 | level60_phase0_balance |
| druid__balance-wrath | 729.86 | 861.5 | level60_phase0_balance |
| druid__balance-wrath-duty-ff | 702.29 | 829.0 | level60_phase0_balance |
| druid__feral-cat | 1054.75 | 1185.9 | level60_phase0_feral |
| druid__feral-cat-duty-ff | 1049.76 | 1179.6 | level60_phase0_feral |
| priest__shadow | 857.46 | 948.7 | level60_phase0_shadow |
| priest__shadow-holy | 818.47 | 906.6 | level60_phase0_shadow |
| priest__smite | 695.29 | 781.0 | level60_phase0_smite |

A list's `dps` is not a plain sim: it is the set on a 180s fight with every use off, plus each use's value on a 90s
fight. Each request matches its list's plain sim of the final set (the `[180]s` line at the end of the rank log):
Balance 822.7, Shadow 857.7, Feral 1054.8, Smite 696.7. Smite is 1.4 under, from the ± 20s fight length on a
mana-bound build.

## Lists

- **Feral:** RANK_ONLY on the full list. No confirmed swap, the set is unchanged, 1067.3.
- **Shadow:** (the list was re-ranked once more on the 2026-11-24 talents, see Hit, talents and rotations: 872.38, same
  picks.) RANK_ONLY, then SLOTS=neck once the Nacreous price was back: Orb of the Darkmoon for Dark Advisor's
  Pendant (0.25, noise). 870.66, was 875.21.
- **Balance:** SLOTS=neck for the Nacreous check. Unchanged, 839.67.
- **Smite:** RANK_ONLY on its new rotation. The set is unchanged, 705.57 (697.64 on the old rotation).
- **Nacreous Shell Necklace** (22403) ranks over the neck pick on Balance (+1.57) and Shadow, but its 217.4g a DPS is
  over the 30g line. The Enhancement price pass 5e556ff had dropped its price, so Shadow's first RANK_ONLY picked it.
  Restored (master da6443e), and the Shadow neck re-ranked.

## Set trap

- **Feral** (Feralheart, wears 1, the head): evals at 180s give 3 pieces 1045.2, 6 pieces 1031.2, 8 pieces 1003.8
  against 1054.8. A full search from the 6-piece start converges to 1045.9 against the list's 1067.3 (the same scoring).
  The list stays.
- **Shadow** (Virtuous, wears 0): the search from the 6-piece start made no swap, 860.5 against 870.7. The 3-piece start
  broke the set and came back to the list (870.5). The list stays.

## Hit, talents and rotations

Phase 0 hit (gear and talents, dashstats): Balance 7% spell, Feral 5% melee (1% from gear), Shadow 9% (4% from gear), Smite 21% with Holy Precision (3% from gear).

- **Balance and Wrath:** talents and knobs unchanged. The Wrath knob search moved the filler to a Starfire weave, which
  isn't the build, so it ran again with the filler frozen: no change.
- **Feral:** 3% under the 8% cap, so Nature's Reach 2 is worth its points. The talent search at that hit keeps the
  talents, and the knobs stay.
- **Shadow:** knobs unchanged. The first talent search was cut at the 2h limit; it ran to the end on 2026-11-24
  (`-talent-search -require silentResolve`, 120/180/300s, 3000 iterations, 20000 to confirm, from the old talents on the
  Phase 0 set; 2,688 sim runs). Three rounds: Wand Specialization 2 for Mental Strength 2 (+3.1), Mental Strength 1
  for Meditation 1 (the mana rate: time to out of mana 376s to 486s), Improved Mind Blast 4 for Shadow Reach 2.
  Talents `02530100103-3-50042201231302251`, was `00530400102-3-50052101231302251`. 857.2 against 853.9 at
  120/180/300s (20000 iterations, SE 0.1 each): +3.3. At the request's 180 ± 20s, 857.46 against 855.19 (+2.3, SE
  0.19). Presets, the request and its meta carry them; the list was re-ranked on them (RANK_ONLY, 872.38, was 870.66,
  the same picks; Virtuous Cowl reads +0.66 over Spellweaver's Turban but stays "doubtful", unavailable-grade).
  The plain sim of the final set is 857.7 and the verified request 857.46.
- **Shadow/Holy:** unchanged. Its 20 Holy points leave no Discipline for Wand Specialization, Mental Strength or
  Meditation, and the Shadow swap (Improved Mind Blast for Shadow Reach) reads 807.2 against 807.3 on it. An
  unconstrained search from it moves Holy points into Discipline (+36), which defeats the build.
- **Smite:** the knob search gains 14.7 (combined 120, 180 and 300s). New defaults in `tools/rotopt/priest_smite.go`:
  Devouring Plague above 30% mana (was 80), Dark Sacrifice below 60% (was 40), rank 3 as the cheaper Smite (was 7),
  max rank while mana lasts at 20 a second (was 35). 713.6 / 695.5 / 666.7 against 702.2 / 687.9 / 641.5 at 120 / 180
  / 300s. `ui/shadow_priest/apls/level60_smite.apl.json` regenerated.

## Races

- **Smite race rotations:** the old Holy Fire and no Devouring Plague overrides lose 11 to 17 to the new defaults. A
  search per race moves only the Smite mana rate (30, Night Elf 25): +1.1 to +1.4, +0.3 at Undead. The five files are
  rewritten.
- **Balance Skyborne rows:** re-checked, +7.7 at 180s.
- **race_gear.py:** no Phase 0 set has hit steps. The priest Dwarf hit-1 files are gone for Shadow and Shadow/Holy
  (10% hit, under the 16% cap). Smite carries 21% with Holy Precision, so the Dwarf is 1% over and keeps a bonus stats
  swap (+3.2), merged into its rotation file.
- **Faction:** only Feral's Horde rows lose picks: Protector's Band (finger) and Theramore Pauldrons (shoulder).
  RANK_ONLY SLOTS=shoulder,finger at Tauren gives `level60_phase0_feral_horde.gear.json`: Darkspear Pauldrons
  (272105) and Legionnaire's Band (19510). The list reads 1046.28 at Tauren (plain 1032.7) against 1067.30 at Night Elf.
  It is `phase0_step4_druid_priest/level60_phase0_feral_horde.bis.json`. race_gear.py puts Tauren and Skyborne
  Windshaper on it, kept by the rule. Duty variants take their base build's files.
- Balance, Shadow and Smite have no faction picks.

## Caveats

- The Balance and Smite APLs' prepull trinket casts the synthetic item 995034. On Phase 0 gear that does nothing.
- Nothing published or deployed: `pipeline.sh publish` is the parent's step.
