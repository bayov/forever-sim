# Pass: tank Phase 0 gear

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03 for the synthetic sets, rewritten 2026-11-25
when tanks moved to Phase 0 lists (`orchestration/standards.md` 36 and 43). `gear.md` is the same pass for the DPS
specs: its rules, gotchas and settled points hold here unless this doc says otherwise.

You check the tank presets' Phase 0 lists: the rows of `tools/rotopt/phase0_specs.json` with `"bench": "tank"`, one
list per preset (`tankPreset`: balanced, max-threat, max-survival), their gear sets (`gearOut`), the requests that
wear them (`tools/tankbench/req/<preset>.json`, `.raid.parties[0].players[0].equipment`), and the other-faction sets
(`<gearOut>_<horde|alliance>.gear.json`) the race sims put on a race of the other faction
(`tools/dps_report/bench_races.py`).

## What differs from the DPS lists

- **Score:** the tank report's own score (tankbench, `tools/tankscore`, at the preset's weight: max threat 1/3,
  balanced 1, max survival 3), the mean over the six scored bosses, each at its own length and healing. Not DPS.
- **Gold line:** `MAX_GOLD_PER_PCT`, 3000g for each 1% of score an item adds. 1% is of the row's `scoreRef`, the
  class's Balanced score on the synthetic set, for all three presets. Only a price made for the spec's own list
  counts (`goldPerDps` by spec, `phase0_prices.py`). `phase0_check.py`'s `goldPerDps` line is the DPS one: read its
  gold flags on a tank row against this line instead.
- **Uses:** they count inside the boss fights (no 90s use fight).

## The pass

1. **What changed** since `last.master` on the watched paths.
2. **Mechanical checks:** `python3 tools/audit/gear/phase0_check.py`, its lines for the tank rows. A row with no list
   yet is expected until its worker lands it: note it, don't flag it.
3. **Each changed list, by hand:** source, faction, effects and stats, weapons and shield, as `gear.md` step 4. Tank
   stats (Defense, dodge, parry, block, block value, armor) come from real items, so there's no stat share to check.
4. **The request wears the list:** each preset's request equipment is its gear set, and the three presets of a class
   wear their own lists. A request still on synthetic items (ids 99xxxx, names with "Synthetic") has no list yet.
5. **The other faction:** each `_horde` or `_alliance` set differs from the list only in slots the preset's faction
   rewards fill, and holds no pick of the preset's faction only.
6. Report: per preset the list's score, faction, picks that changed, doubtful picks and unmodelled effects; new
   findings.

## Synthetic sets (history)

Until 2026-11-25 tanks wore synthetic Phase 1 sets (`tools/database/forever_synthetic_items_<class>_tank.go`, standards
1b, 2, 4, 5a, 6, 7, 8a and 36). A preset whose list hasn't landed still wears one. This pass's synthetic checks (budget,
shape and stat share per piece and set) are in this file's git history.
