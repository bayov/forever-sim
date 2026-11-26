# Rotations pass 82 (2026-11-26, master ab3fff7007)

Result: findings (2 new, low). Range bfbee78fc..ab3fff7007 (29 commits): warrior leveling sim (ranks by level,
stances, leveling sessions, big pulls, area reach), and the Phase 0 re-rank for warriors and hunters (weapons,
trinkets, race rows). No Level 60 APL file changed.

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py bfbee78fc ab3fff7007` | 0 APL files touched |
| Req rotations changed | none (only equipment in warrior and hunter reqs) |
| Non-test sim `.go` changed | 40 files (core and warrior), almost all gated on `encounter.leveling`; `InExecuteRange` equals `IsExecutePhase*` outside leveling |
| `regress.py` old bfbee78fcca3 vs new ab3fff700700, new reqs, 3000 it, seed 424242, min 0.05 | only the 7 paladin builds wearing Blackblade of Shahram moved (+0.8 to +1.6); all warriors unchanged. Cause: stunned targets no longer dodge/parry/block, and Might of Shahram stuns the stun-immune boss: rotations-20261126-3 |
| Open findings landed? | none of 20261126-1, 20261122-1, 20261112-1 (no APL change) |
| Phase 0 trinket lines, warriors and hunters | Fury DW now wears 22321+19289 (procs) but its APLs still cast 13965: rotations-20261126-2. Arms, Fury 2H wear 13965 with no line (autocast). BM/MM/LW wear 10455 and name it. Survival wears 19289+22321, no line. |
| Baseline | moved (paladin Shahram, plus re-rank gear since 8569165); new `baseline_ab3fff7_3000it_seed424242.tsv` |

## Baseline moves vs baseline_8569165 (3000 it, seed 424242)

Most moves come from req gear changes in the Phase 0 re-rank between 8569165 and ab3fff7 (passes 81-82), not sim
changes: the sim-only signal is the regress row above. Largest: warlock affliction_shadow_bolt +25.8, Balance -12 to
-17, MM -11.6 to -14.5, Survival +14 to +15, Smite -12.9, rogues -13 to +8, warriors -0.2 to -8.4.

## Not done this pass

- Phase 0 trinket check still pending for mage, priest, rogue, shaman, warlock (open item in passes/rotations.md).
- Windfury timing item, Flametongue-imbue item: untouched.

## Optional

None new.
