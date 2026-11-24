# tank_rotations pass 27

Master c7e3b3e1ef (2026-11-24, prot-racials-1124), against pass 26's baseline (566067a394). Focus: the Protection
Warrior's racial actives (Blood Fury, Berserking, Elune's Light, Eureka!) under the new `racial` / `racialAt` knobs.
Quiet: no new finding. Master moved to dfd0159257 during the pass (Stoneform knob, see APL check); the grid's
races don't cast Stoneform, so its numbers carry.

## What changed

Gate (e29cfe00 .. c7e3b3e): `tools/rotopt/warrior_tank.go` (racial knob, `tankRacialLines`), the three warrior
requests and `ui/tank_warrior/apls/level60_*.apl.json` (four racial lines each, on cooldown). No sim, encounter,
tankbench or tankscore change; no paladin, bear or Warden file.

## Baseline

The presets are Undead, so the racial lines are no-ops: the 24 warrior rows of `ref/tank_bench/summary.tsv`
(1,000 iterations, default seed) are byte-identical; the other tank requests are untouched. `summary.tsv` not
rewritten.

## APLs match rotopt

`rotopt -template warrior_tank -eval -iters 10 -out` with each preset's knob string, compared with the UI APL and
the request's rotation:

| master | knob strings | Balanced | Max survival | Max threat |
|---|---|---|---|---|
| c7e3b3e | `orchestration/tank_handoff/windfury_retune/knobs/warrior_*` | identical | identical | identical |
| dfd0159 | `orchestration/tank_handoff/warrior/racials_1124/knobs/warrior_*` (adds `stoneformAt=20`) | identical | identical | identical |

On dfd0159 the windfury_retune knob files no longer rebuild Balanced and Max survival (no `stoneformAt`); the
racials_1124 files are the current source. Not a finding (handoff history), but the next re-tune should start
from racials_1124.

## Racial timing, seed 2

Paired `abtest -iters 10000 -seed 2` (disjoint from the seed 1 grid), A = the preset as that race with racials on
cooldown from the pull (`racial=0`), B = the variant's rotopt APL, same race. Scoring mean of the six bosses at the
preset's weight, B minus A (negative = the pull wins):

| Variant | Orc Bal | Orc MS | Orc MT | Troll Bal | Troll MS | Troll MT |
|---|---|---|---|---|---|---|
| r1 (with AP trinket) | 0 | 0 | 0 | 0 | 0 | 0 |
| r2 last 13 s | -85.3 ± 0.2 | -80.5 ± 0.2 | -94.2 ± 0.2 | -45.9 ± 4.8 | -48.6 ± 8.5 | -48.9 ± 3.3 |
| r2 last 15 s | -76.7 | -75.5 | -85.4 | -43.3 ± 4.8 | -44.4 ± 8.5 | -42.5 ± 3.3 |
| r2 last 30 s | -38.4 | -46.1 | -42.6 | -25.0 ± 4.5 | -27.1 ± 7.9 | -22.1 ± 2.8 |
| r2 last 60 s | -7.8 ± 0.1 | -10.3 ± 0.1 | -8.6 ± 0.1 | -7.2 ± 2.0 | -8.6 ± 3.9 | -7.1 ± 1.3 |
| r3 from 5 s | -30.4 | -32.3 | -34.0 | -9.2 ± 4.8 | -19.7 ± 8.4 | -8.2 ± 3.5 |
| r3 from 10 s | -61.9 | -59.8 | -69.5 | -25.7 ± 5.0 | -26.1 ± 8.4 | -27.7 ± 3.4 |

| Variant | NE Bal | NE MS | NE MT | Gnome Bal | Gnome MS | Gnome MT |
|---|---|---|---|---|---|---|
| r1 | 0 | 0 | 0 | 0 | 0 | 0 |
| r2 last 13 s | -110.3 ± 1.2 | -118.2 ± 3.4 | -117.0 ± 0.8 | -73.2 ± 2.3 | -76.5 ± 5.2 | -75.8 ± 0.9 |
| r2 last 15 s | -99.2 | -111.1 | -106.2 | -64.7 | -70.6 | -67.4 |
| r2 last 30 s | -51.1 | -69.2 | -54.5 | -35.8 | -43.5 | -35.8 |
| r2 last 60 s | -13.2 ± 0.5 | -17.3 ± 1.8 | -14.3 ± 0.3 | -6.3 ± 0.8 | -9.1 ± 2.1 | -6.8 ± 0.3 |
| r3 from 5 s | -42.2 | -52.0 | -44.7 | -35.1 | -30.5 | -30.8 |
| r3 from 10 s | -78.1 | -90.2 | -84.4 | -66.0 | -68.6 | -64.2 |

Errors not shown repeat the row's order of magnitude (Orc ±0.2, Troll ±3–9, NE ±0.5–3.5, Gnome ±0.3–5).

- **The pull wins every cell on seed 2**, as on seed 1. Every hold or delay loses on every scoring boss
  (Vaelastrasz is 0 for holds of 30 s or more only because its 17 s fight makes them the pull). The losses grow
  with the hold, since the fights are 17–115 s and the threat lead is built in the first seconds.
- **r1 is the pull** at every preset: Balanced and Max threat use 991018 on the pull (`trinketAt=0`), and Max
  survival wears no attack power trinket, so `spellTimeToReady` of the unworn trinket reads 0. The knob value is
  live only if a preset holds 991018.
- **Synced with Shield Wall** (Max survival, `wallAt=13`, r2 at 13 s): -80.5 (Orc), -48.6 (Troll), -118.2 (NE),
  -76.5 (Gnome). The racials buff threat, not mitigation; holding them for Shield Wall's window only loses.
- **Berserking for Rage starvation** (Troll, Berserking line hand-edited to `currentRage < N`, the rest pulled):
  Rage < 20: -42.8 ± 4.9 (Bal), -5.6 ± 3.1 (MS), -56.8 ± 3.5 (MT); Rage < 40: -29.8 ± 5.1 (Bal), 0.0 (MS, fires at
  the pull), -42.0 ± 3.4 (MT). Waiting for low Rage only delays it.
- Ragnaros (not scored) agrees except Troll Balanced, where holds gain about +110 (one noisy boss at w = 1; the
  six scoring bosses all lose).

The shipped `racial=0` stands for all four races at all three presets.

## Open findings

20261122-1 (paladin Consecration mana floor), 20261104-11 and 20261104-3 (Broodlord user calls) stand: no paladin
change, warrior trinket and Shield Wall lines untouched.

## Open items carried

- Windfury main-hand ability timing (user request 2026-11-15, pass doc Open items): still not done; out of scope for
  this racial-focused pass.

## Repro

```sh
W=$AUDIT_WORK/tank_rotations
rotopt -template warrior_tank -settings tools/tankbench/req/warrior_protection_<p>.json -tank <p-dashed> \
  -tank-encounter tools/encounters/vaelastrasz.json -knobs "$(cat orchestration/tank_handoff/warrior/racials_1124/knobs/warrior_<p>),racial=2,racialAt=30" \
  -eval -iters 10 -confirm-iters 10 -out $W/apl -name r2_30_<p>
python3 tools/audit/tank_rotations/mk_race_req.py tools/tankbench/req/warrior_protection_<p>.json Orc $W/A.json
python3 tools/audit/tank_rotations/mk_race_req.py tools/tankbench/req/warrior_protection_<p>.json Orc $W/B.json $W/apl/r2_30_<p>_*s.apl.json
abtest -preset <p> -iters 10000 -seed 2 $W/A.json $W/B.json
```
