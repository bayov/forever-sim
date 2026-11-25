# Phase 0 weapon, trinket and set re-rank on the new sim (2026-11-25)

The DPS lists were ranked before the last sim batches: items (db86a18dc), procs (aed8856d1), casters
(efeb06b95: resists, overloads, Windfury and Flurry) and the log conflicts (22d2cf952: throws keep the swing,
stun explosives do nothing to level 63 bosses, extra attacks roll Windfury Totem). These move proc weapons,
trinkets and set bonuses most. The prices round (master 6ba024d1b, notes `phase0_prices_1125.md`) re-ranked
only the slots whose gold line changed. You re-rank the rest on current master, then publish and deploy.

Read first: `briefs/phase0_step3_1122.md` (how the search runs), `briefs/phase0_step4_1124.md` (gear sets,
requests, race rows, the set trap), `phase0_prices_1125.md`, `tools/rotopt/bis_phase0.sh`,
`tools/rotopt/phase0_prices.py`, `orchestration/standards.md` clause 43.

## 1. Re-rank

For every DPS spec list in `tools/rotopt/phase0_specs.json`: re-rank the weapon slots (main hand, off hand,
two-hand, ranged, and the relic slot where the class has one) and both trinkets (`SLOTS=`), then the set
trap check (a set bonus that needs several pieces at once). Delegate per class group to cloud sessions
(create_session with this repo, at most one level below you). Each group session gets one short brief and
hands back its spec list commits and per build what changed.

## 2. Prices

`goldPerDps` is stored per spec and depends on the DPS an item adds, so re-run `phase0_prices.py --write`
after the re-ranks and re-rank any slot whose pick crosses the line. This session has bayov/wow attached for
the inputs. Build them as the prices notes say. Also:
- Price Crystal-Infused Cleaver, which has no price.
- Price the faction copies of each set (`<gearSet>_horde` / `_alliance`), which the script skips.

## 3. Open items from the prices round

- Feral APL lines still cast Dark Bomb, which now does nothing to level 63 bosses. A/B against Goblin Land
  Mine and none, keep the best.
- `enhancement-imp-gw` talents fail talent_check on master. Fix them.
- Windshaper rotations were not re-checked on the Horde gear. Re-check.
- Hunters sit at 6.6% hit with Accurascope out. Check the hunter lists aren't missing a hit item the line
  allows.
- Virtuous Cowl, Hyper Deluxe and Bloodseeker are doubtful picks. Check them against standards 42 (unsure
  availability stays in) and note the answer.

## 4. Publish, audit, deploy

Gear sets (`GEAR_ONLY=1 SPEC=<spec>`), requests and race rows (`tools/dps_report/race_gear.py`), check each
request sims to its list. `stale.mjs`, `pipeline.sh hydrate/run/publish`, `build_data.py`, `changes.py` and a
Changes tab entry. Run the gear, results and rotations auditors on the changed builds, fix round, merge to
master (fetch and merge first, never force-push), `make dps-report-deploy`, check the live page.

The tank and healer worker (branch `phase0-tankheal`) is still searching. Don't edit its files, and merge
master often.

## Hand back

One message to the parent, under 300 words: master commit, deploy check, per build what changed (item,
DPS before and after), the open items' answers, anything left open. Then stop.
