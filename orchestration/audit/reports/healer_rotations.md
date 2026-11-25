# Healer rotations pass 27 (findings: new -20261125-1, update on -20261104-2)

- Master: 0404484fb (`last.master` was 5536db4bc). Audited 2026-11-25.
- Bench: `tools/healbench` built on master (after `protoc -I=./proto --go_out=./sim/core ./proto/*.proto`;
  `setup.sh` doesn't generate the protos), default seed, `-carry`, 1,000 iterations, via
  `tools/audit/healer_rotations/baseline.sh`. Ref `ref/healer_bench/summary.tsv` refreshed.

## What changed since 5536db4bc

- Watched: `sim/core/spell_result.go` (b4cab8498, `PreAbsorbDamage` for warrior Rage from hits taken; no heal
  effect), `sim/shaman/air_totems.go` and `fire_totems.go` (5dbc01412, comments only).
- Not watched but healer-relevant: f1942793d (Arcane Intellect and Divine Spirit overwrite the elixirs of their
  stat, EV-0185). The healer requests carry Elixir of Greater Intellect and Elixir of Sages, so every healer
  lost their Intellect/Spirit. World buffs off in raid instances (3275a60b1, 9701cff08): the healer requests
  have none.
- No healer APL, healer request, encounter file, heal, talent or healbench main.go change.

## Baseline (1,000 it) vs ref

63 rows: 60 moved, 14 by more than 0.02 carry, mean carry -0.006, mean mana left -455.

| request | boss | carry ref -> now | mana left | OOM | other deaths |
|---|---|---|---|---|---|
| priest_holy | chromaggus | 0.970 -> 0.911 | 782 -> 488 | (126 s) -> 113 s | 4.6% -> 11.4% |
| priest_discipline | chromaggus | 0.871 -> 0.828 | 341 -> 327 | 101 s -> 95 s | 23.6% -> 44.0% |
| druid_restoration | golemagg | 1.952 -> 1.902 | 6,250 -> 5,820 | | |
| druid_restoration | chromaggus | 1.037 -> 1.011 | 2,711 -> 1,874 | (228) -> (185 s) | 0.4% -> 1.0% |
| druid_restoration_raid | magmadar | 1.828 -> 1.790 | | | |
| druid_restoration_raid | onyxia | 0.988 -> 0.961 | | | 6.5% -> 5.4% |
| paladin_holy | golemagg | 1.806 -> 1.778 | | | |
| paladin_holy_raid | broodlord | 0.904 -> 0.882 | | | |
| priest_discipline | broodlord | 0.958 -> 0.935 | | | |
| priest_holy_raid | golemagg | 2.512 -> 2.484 | | | |
| shaman_restoration | magmadar | 2.111 -> 2.085 | | | |
| shaman_restoration_raid | chromaggus | 1.062 -> 1.107 | | | |
| shaman_restoration_raid | onyxia | 0.889 -> 0.921 | | | 9.3% -> 9.6% |
| shaman_restoration_raid | broodlord | 0.884 -> 0.852 | | | 11.5% -> 13.1% |

Cause confirmed with healbench built at f1942793d^ and f1942793d (priest_holy tank): Chromaggus 0.970/782 ->
0.912/486, Golemagg 2.000/6,501 -> 1.990/6,019. Dropping both elixirs from the master request gives 0.904/463
(noise), so they are dead slots. The rule is right (in-game evidence); the finding is the new OOM at a judged
length for the Holy tank healer.

## Changes tried

None this pass (time went to confirming the cause). Next pass: one-line mana lines for priest_holy tank on
Chromaggus (lower bound Greater Heal rank below a mana %), and re-check -20261115-1's Wild Growth gate.

## Open findings

-20261125-1 (new), -20261124-1 (filler pin), -20261118-1, -20261115-1, -20261104-2 (Disc OOM, worse), and the rest
unchanged.
