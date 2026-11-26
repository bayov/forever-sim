# Rotations pass 83 (2026-11-26, master 2af970d74)

Result: quiet. Range 6a447e231..2af970d74 (~40 commits): Phase 0 prices round 2 re-rank (equipment in
warrior, rogue, hunter, druid, shaman/ret requests), Balance trinket-line cleanup, warrior Prot leveling plan,
mksettings and phase0 price tools.

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py 6a447e231 origin/master` | 4 Balance APLs: 10455 pre-pull removed, Elune's Light (1259799) now unconditional. Exactly rotations-20261126-1's fix |
| Req rotations changed | only the four Balance reqs, same change |
| Non-test sim `.go` changed | `sim/common/guardians/barov_servant.go`: negative-wait guard (crash fix). No request wears Barov Peasant Caller (14022/14023), so no preset is reached; no regress run needed |
| Open findings landed? | -1 resolved (c1ebe2d79). -2 moot: round 2 put Fury DW back on 19289+13965 (832070749), so its 13965 line names a worn trinket again. -3, 20261122-1, 20261112-1 still open (no change) |
| Rotation item lines vs worn gear, all reqs | only consumables (sappers, potions, gems, oils, stones) not worn; no stale trinket line. Seal Fate Swords (now 19289+13965) has no 10455 line |
| Baseline | unchanged: `baseline_ab3fff7_3000it_seed424242.tsv` (no reached sim change) |

## Optional

None new.

## Left open

- Phase 0 trinkets open item: Balance now wears Talisman of Ascendance (22678, on-use) with no line (autocast);
  whether holding it for Elune's Light pays was not measured this pass.
