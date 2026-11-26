# Phase 0 re-rank, class group workers (2026-11-25)

Parent brief: `briefs/phase0_rerank_1125.md`. The DPS lists were ranked before the sim batches db86a18dc (items),
aed8856d1 (procs), efeb06b95 (casters: resists, overloads, Windfury, Flurry) and 22d2cf952 (log conflicts: throws
keep the swing, stun explosives do nothing to level 63 bosses, extra attacks roll Windfury Totem). These move proc
weapons, trinkets and set bonuses most. Your session prompt names your group and its specs (keys of
`tools/rotopt/phase0_specs.json`).

Read first: `briefs/phase0_step3_1122.md`, `briefs/phase0_step4_1124.md` (gear sets, requests, race rows, the set
trap), `phase0_prices_1125.md`, `phase0_step4_<group>.md`, `tools/rotopt/bis_phase0.sh` (header), standards 42 and 43.

## Task, per spec

1. Merge origin/master. `RANK_ONLY=1 SLOTS=mainhand,offhand,ranged,trinket SPEC=<spec> tools/rotopt/bis_phase0.sh`
   (ranged covers relics), in the background with output in a file. Run as many specs at once as `nproc` allows.
   Wait for the notification, don't poll. Never ship `ALLOW_UNLISTED=1`, tag unlisted picks as step 3 says.
2. Set trap: a set bonus needing several pieces at once. Check from a start set at the full set and the bonus
   points for any set the class can wear in Phase 0 (step 4 section 1). Keep what wins.
3. Where talents or the pick change the answer, re-rank what it touches. If any pick changed:
   `GEAR_ONLY=1 SPEC=<spec>`, the faction copies (RACE runs, step 4 section 3), the requests and metas
   (`bin/mksettings`) and `tools/dps_report/race_gear.py` rows. Check each request sims to its list within noise.
4. Open items for your group (below).

Do not run `pipeline.sh publish`, commit under `ui/dps_report/data`, deploy, or run `phase0_prices.py`. Don't touch
tank or healer files (branch `phase0-tankheal` works on them). Shared files (`phase0_content.json`,
`phase0_enchants.json`, `race_gear.json`, `docs/bis_lists.md`): merge origin/master often, keep both sides.

## Open items

- druid_priest: feral APL lines still cast Dark Bomb, which does nothing to level 63 bosses. A/B Goblin Land Mine
  against none, keep the best. Virtuous Cowl (shadow): check against standard 42 and note the answer.
- shaman_ret: `enhancement-imp-gw` talents fail talent_check on master, fix them. Re-check the Windshaper rotations
  on the Horde gear.
- hunter: hunters sit at 6.6% hit with Accurascope out. Check the lists aren't missing a hit item the 30g per DPS
  line allows. Hyper Deluxe Sniper Rifle: check against standard 42 and note the answer.
- rogue: Bloodseeker: check against standard 42 and note the answer.

## When done

Write `orchestration/round_notes_2026-11-11/phase0_rerank_<group>.md`: per build the slot changes (item old->new),
list dps and request dps before and after, set trap result, open items' answers. Commit on your branch, then fetch
and merge to master and push both (never force-push). User style: concise, no em dash or semicolon, "180s".
Send the parent one message under 250 words: master commit, per build what changed with DPS before and after, open
items' answers, real problems. Then stop.
