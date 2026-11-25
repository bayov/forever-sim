# Pass: healer Phase 0 gear

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03 for the synthetic sets, rewritten 2026-11-25
when healers moved to Phase 0 lists (`orchestration/standards.md` 35 and 43). `gear.md` is the same pass for the DPS
specs: its rules, gotchas and settled points hold here unless this doc says otherwise.

You check the healers' Phase 0 lists: the rows of `tools/rotopt/phase0_specs.json` with `"bench": "heal"`, one list
per spec, their gear sets (`gearOut`), the requests that wear them (`tools/healbench/req/<spec>.json` and
`<spec>_raid.json`, `.raid.parties[0].players[0].equipment`), and the other-faction sets
(`<gearOut>_<horde|alliance>.gear.json`) the race sims put on a race of the other faction
(`tools/dps_report/bench_races.py`).

## What differs from the DPS lists

- **Score:** the healer report's damage carried (healbench carry), the mean over the seven bosses in each of the
  spec's roles (`healRoles`), the roles weighed alike. The raid role sims the raid request (`raidReqFile`) under the
  list. Not DPS.
- **Gold line:** `MAX_GOLD_PER_PCT`, 3000g for each 1% of the spec's damage carried (the row's `scoreRef`) an item
  adds. Only a price made for the spec's own list counts (`goldPerDps` by spec, `phase0_prices.py`).
  `phase0_check.py`'s `goldPerDps` line is the DPS one: read its gold flags on a healer row against this line instead.
- **Uses:** they count inside the boss fights (no 90s use fight).

## The pass

1. **What changed** since `last.master` on the watched paths.
2. **Mechanical checks:** `python3 tools/audit/gear/phase0_check.py`, its lines for the healer rows. A row with no
   list yet is expected until its worker lands it: note it, don't flag it.
3. **Each changed list, by hand:** source, faction, effects and stats, weapons, as `gear.md` step 4. The relic's
   effect is modelled (`sim/<class>/items.go`) or flagged. Spirit, MP5 and hit are what the search picked, so there's
   no stat share to check.
4. **The requests wear the list:** both of a spec's requests (tank healer and raid healer) wear its gear set. A
   request still on synthetic items (ids 99xxxx, names with "Synthetic") has no list yet.
5. **The other faction:** each `_horde` or `_alliance` set differs from the list only in slots the spec's faction
   rewards fill, and holds no pick of the spec's faction only.
6. Report: per spec the list's carry, faction, picks that changed, doubtful picks and unmodelled effects; new
   findings.

## Synthetic sets (history)

Until 2026-11-25 healers wore synthetic Phase 1 sets (`tools/database/forever_synthetic_items_<class>_<spec>.go`,
standards 1b, 5, 5a and 35, the healer rows of `orchestration/gear_targets.md`). A spec whose list hasn't landed still
wears one. This pass's synthetic checks (budget, shape and stat share per piece and set, `orchestration/tools/healers/`)
are in this file's git history.
