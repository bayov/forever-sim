# Phase 0 step 4, class workers (2026-11-24)

The level 60 DPS report moves from synthetic gear to the Phase 0 gear lists for every DPS build, with no
toggle. Tanks and healers stay on synthetic. You do one class group. The parent publishes, audits and
deploys at the end.

Start from master 69086e56e or later. It has:
- cdcb7e686: a SLOTS run now measures the final set, so the list's dps is fresh.
- feb2a7e63: bis_phase0.sh falls back to git when jj is absent.
- 963d1d01d: Raider gloves (272095 to 272101) and the Darkspear shoulders are open (Revered per the item
  data), all Darkspear items are tagged Horde, and the Superior Agility bracer enchant has its source.
- 69086e56e: every spec has `gearOut`. Full and RANK_ONLY runs write the spec's gear set
  `ui/<dir>/gear_sets/level60_phase0_<spec>.gear.json`. To write it from the current list alone:
  `GEAR_ONLY=1 SPEC=<spec> tools/rotopt/bis_phase0.sh`.

Read `briefs/phase0_step3_1122.md` (how the search runs), your group's `phase0_step3_<group>.md`, and
the notes in `orchestration/round_70753/hit_trade/` (the hit-trade results, `numbers.md`).

## 1. Finish the lists

Do what applies to your specs:
- **Stale SLOTS numbers:** feral, ret, the three rogue lists, shadow, demonology and destruction. Rerun
  RANK_ONLY to re-measure.
- **Feral:** its list predates d093d6e22 and be773f5eb (feral sim fixes). RANK_ONLY on current master.
- **Hunters:** Black Dragonscale Helm and Windreaver Greaves are now over the gold line, so the gear sets
  differ from the lists. RANK_ONLY all three.
- **Arms and Fury DW:** Raider Gauntlets now rank 2.1 and 2.3 above the hands pick. Re-rank hands.
- **Set trap:** the full search swaps one slot at a time, so it can stop short of a set whose bonus needs
  several pieces at once (Enhancement stuck at Five Thunders 4pc, the 6pc was 7.7 better). Check from a
  start set at the full set and the 3 and 5 piece points: feral (Feralheart, wears 1), shadow (Virtuous,
  wears 0), the warlocks (Deathmist, wear 0 or 1). Keep what wins.

## 2. Hit-dependent talents and rotations

For every report build of your class (`tools/dps_report/req/*.meta.json`, some specs have several builds),
re-check the talent and rotation choices that depend on hit at the Phase 0 set's hit, with the class's
`tools/rotopt/<class>_presets.sh`, `tools/rotopt/hitsteps.go` and `orchestration/tools/hit_steps.py`.
Update talents and APLs where the answer changed. If a build's talents change, RANK_ONLY its list again.

## 3. The report's requests

Regenerate your class's DPS requests and metas in `tools/dps_report/req/` with `bin/mksettings` (see
README.md and the class presets script), with `gearSet` set to the Phase 0 set and the `_hit±N` steps
gone. Builds that share a spec use that spec's set unless their talents need their own list.

Race rows: update `tools/dps_report/race_gear.py`, `race_gear.json` and `req/races/` for your class. A row
of the other faction can't wear the list's faction rewards: swap those slots for its own faction's best
(SLOTS with that race), and race-specific weapon skill picks where the race changes the answer.

Check each new request sims to its list's dps within noise. Do not run `pipeline.sh publish`, do not
commit anything under `ui/dps_report/data`, and do not deploy. Don't touch tank or healer files (a prot
warrior session works in parallel).

## When done

Write `orchestration/round_notes_2026-11-11/phase0_step4_<group>.md`: per build the Phase 0 dps, the
synthetic dps it replaces, talent or rotation changes, list changes and why. Merge to master (fetch and
merge first, never force-push), push your branch too. Send the parent one message: the master commit,
and in under 250 words, per build Phase 0 vs synthetic, what changed, and any real problem. Then stop.
