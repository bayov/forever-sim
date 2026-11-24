# Phase 0 step 4: mage (2026-11-24)

The four mage report builds now wear their Phase 0 sets (`ui/mage/gear_sets/level60_phase0_mage_<spec>.gear.json`, the
step 3 lists). Requests come from mksettings with the same arguments and `-gear level60_phase0_mage_<spec>`. Rerun with
the old sets, mksettings gave the old requests exactly, so the only change is the equipment.

## Per build

verify.py, 180s ± 20s, 20000 iterations, seed 1.

| Build | Race | Phase 0 | Synthetic it replaces | List dps | Step 3 calibration |
|---|---|---|---|---|---|
| Arcane | Gnome | 843.1 ± 0.3 | 962.3 (`_arcane_hit-1`) | 855.2 | 843.5 |
| Fire | Human | 920.1 ± 0.4 | 1057.9 | 930.0 | 921.1 |
| Frost | High Order | 726.4 ± 0.3 | 944.7 | 740.4 | 726.8 |
| Frostfire | Human | 873.5 ± 0.5 | 1064.3 | 883.4 | 872.4 |

- The requests match the step 3 calibration within 1. The list dps reads 9 to 14 higher. A fixed 180s fight moves
  the request's number only 0 to 6, so most of the gap is the list's own scoring (uses scored on a 90s fight,
  `use_dur` 90). This is the same for every spec, and it doesn't move the picks.
- No talent, rotation or list changes. The lists are fresh, from the step 3 rerun, and none of the step 4 list items
  (stale SLOTS, set trap) applies to mage.

## Hit

Spell hit on the Phase 0 set at the preset race: Arcane 12.3%, Fire 13.3%, Frost 11.0%, Frostfire 13.3%, against
the 16% cap. All are 2.7 to 5 under it, so Precision 3/3 stands and no hit-dependent talent or knob changes. The hit
trade's Arcane change (Flame Throwing to Master of Elements) was a tie without the gear step (+0.3 to +0.8) and is kept.

## Neck: Nacreous Shell Necklace

All four lists rank it first (0.6 to 2.8 over the pick), tagged `infeasible` at 217g per DPS, over the 30g a DPS cap.
That is why it isn't picked. The ranking is from the step 3 rerun on current item data, so no re-rank.

## Races

- **Faction:** a dry run at Troll leaves out 83 Alliance-only items per spec. The only pick among them is Arcane's
  Songstone of Ironforge. RANK_ONLY SLOTS=finger at Troll picks Eye of Orgrimmar (Horde, BRD quest "The Princess
  Saved?"), level with Channeler's Ring there. The Horde list (847.1 at Troll) is
  `phase0_step4_mage/level60_phase0_mage_arcane_horde.bis.json`, its set
  `ui/mage/gear_sets/level60_phase0_mage_arcane_horde.gear.json` (the finger slot only differs).
- **race_gear.py:** a race of the other faction from the preset race's now wears `<gearSet>_horde` (or `_alliance`)
  where that file exists, kept by the rule. Arcane's Orc, Undead and Troll rows wear the Horde set. Every other mage
  row has no swap (no hit steps on Phase 0 sets, and no mage Dwarf). Mages have no weapon skill question.
- **Race rotations,** re-simmed on the Phase 0 sets against the preset's rotation (paired, 20000 iterations):
  - Arcane High Order: +8.1 ± 0.1, kept.
  - Fire High Order: +0.08 ± 0.04, kept by the 2 SE rule (regenerated with the current knobs, same rotation).
  - Frostfire High Order -26.2 ± 0.4 and Troll -25.5 ± 0.4: deleted, with their `race_text.json` entries. They
    carried `fireBlastMana=0` from when the preset's was 80. The preset is now 40, and with 40 plus only the race
    knob they still lose (-5.5 and -1.3). Both rows now use the preset's rotation.

## Publish checks (2026-11-24)

- **Frost's -23% is the old comparator, not the gear.** The 944.7 it replaces is the 2026-11-02 number, from before the
  Ice Lance 0.11 fix (fairness pass 49: Frost 945 to 834.6, then the re-tune to 838.6). The old synthetic request
  sims 838.7 today, so the Phase 0 set costs Frost 13.3% (838.7 to 726.8), with Arcane 12.4%, Fire 12.5% and
  Frostfire 13.4%. Frostfire's 1064.3 is stale the same way (1007.4 today). Nothing in the set or rotation is broken.
- **List dps against requests.** A list's `dps` is the set on 180s with every use off plus each use's value on 90s
  (`use_values.go`), so a use counts for more there than in a 180s fight. Measured on Frost: Chained Essence of
  Eranikus adds 16.4 on a fixed 90s fight and 8.4 on 180s, Eye of the Beast 14.0 and 9.6. That is 12.5 of the 14.0
  gap. Each request sims to its list's plain 180s measure (`rotopt -eval`, the `[180]s` line) within noise:
  Arcane 843.1 / 843.5, Fire 920.1 / 921.1, Frost 726.8 / 727.2, Frostfire 873.5 / 872.4 (verify.py 20,000
  iterations seed 1 / eval). Nothing to fix. Frost's request reads 726.8 now against its meta's 726.4 (the Skyborne
  base offsets, e19d2df); the publish re-sim picks it up.
