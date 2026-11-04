# Round 70753 phase 5, rogue: talents with hit-trade gear steps

Branch `phase5-hit-rogue`, started from `phase5-hit-trade` (aed188a4b). Six Level 60 builds, none with a `baseBuild`.

## Result

**No talent changes.** Every build keeps its talents, with Precision 3/3 and the talents that need it. Every build
wants **more gear hit than its set carries, not less**: the best step is +2 or +3 (12–13% total hit against the 8%
yellow cap), worth +4.3 to +5.8 DPS at each fight length.

| Build | Gear step (file) | Total hit before → after (yellow cap 8%) | Old DPS 120 / 180 / 300 | New DPS 120 / 180 / 300 |
|---|---|---|---|---|
| Combat Swords | +2 (`level60_phase1_synthetic_swords_hit+2`) | 10% → 12% | 990.2±0.3 / 958.6±0.3 / 932.6±0.2 | 995.8±0.3 / 964.3±0.3 / 937.7±0.2 |
| Combat Daggers | +2 (`level60_phase1_synthetic_combat_daggers_hit+2`) | 14% → 16% | 975.7±0.3 / 949.8±0.2 / 926.3±0.2 | 980.2±0.3 / 954.5±0.2 / 930.9±0.2 |
| Seal Fate Swords | +3 (`level60_phase1_synthetic_seal_fate_swords_hit+3`) | 10% → 13% | 959.3±0.4 / 940.4±0.3 / 924.6±0.2 | 963.9±0.4 / 945.0±0.3 / 929.0±0.2 |
| Seal Fate Daggers | +3 (`level60_phase1_synthetic_daggers_hit+3`) | 10% → 13% | 919.6±0.3 / 903.4±0.2 / 889.2±0.2 | 925.0±0.3 / 909.1±0.2 / 895.0±0.2 |
| Subtlety | +2 (`level60_phase1_synthetic_subtlety_hit+2`) | 11% → 13% | 895.9±0.3 / 881.3±0.2 / 867.3±0.2 | 900.5±0.3 / 885.6±0.2 / 871.7±0.2 |
| Mutilate | +3 (`level60_phase1_synthetic_mutilate_hit+3`) | 10% → 13% | 873.2±0.3 / 858.3±0.2 / 847.4±0.2 | 878.4±0.3 / 863.6±0.2 / 852.7±0.2 |

20,000 iterations, seed 1, the build's report request (180 ± 20 s style variation as the request has it), each
fight length on its own. "Old" is the current talents on the current set (no `-hit-steps`). "New" is the same talents
on the step. Total hit is the player's final melee hit on the set (gear, Precision 3/3, the 2-piece set bonus, the
request's buffs and consumables). It is also the spell hit the poisons roll on, since gear hit rating raises both
(see below).

Talent strings, old = new:

| Build | Talents | Pin kept |
|---|---|---|
| Combat Swords | `00530320201-3200330112012321551` | adrenalineRush |
| Combat Daggers | `005203002-3203230002012221551-005` | adrenalineRush |
| Seal Fate Swords | `0053032024105103-320033011201132` | sealFate |
| Seal Fate Daggers | `0053032024105103-321303010201132` | sealFate |
| Subtlety | `00530301-320003-115003231320021151` | hemorrhage |
| Mutilate | `00530320251451031-320303-005` | mutilate |

## Why the step goes up, not down

The cap step (total hit back to 8%) is -2 for four builds, -3 for Subtlety and -4 for Combat Daggers. On all six,
DPS is flat from the cap step to the set itself (step 0), then rises at +1 to +3. Each step scan used the final
talents, 120/180/300 s combined, 10,000 iterations, ±0.2:

| Step | -6 | -5 | -4 | -3 | -2 | -1 | 0 | +1 | +2 | +3 |
|---|---|---|---|---|---|---|---|---|---|---|
| Combat Swords | | 955.6 | 957.8 | 959.1 | *960.6* | 960.8 | 960.3 | 963.3 | **965.7** | 965.5 |
| Combat Daggers | | | *941.3* | 944.4 | 947.1 | 949.5 | 950.9 | 953.6 | **955.3** | 954.8 |
| Seal Fate Swords | | 931.2 | 937.0 | 942.6 | *943.3* | 941.9 | 941.6 | 943.9 | 946.0 | **946.1** |
| Seal Fate Daggers | | 891.9 | 897.5 | 903.6 | *904.7* | 904.3 | 903.9 | 906.4 | 908.9 | **909.7** |
| Subtlety | 872.2 | 877.0 | 879.6 | *881.7* | 881.5 | 881.5 | 881.3 | 883.7 | **885.9** | 885.4 |
| Mutilate | | 843.7 | 848.8 | 855.9 | *859.4* | 859.1 | 859.4 | 861.9 | 864.3 | **864.8** |

*Italics*: the cap step. **Bold**: the best step, which the 20,000-iteration "new" rows wear.

- Hit past 8% still pays a rogue. White swings dual wield against a 27% cap, and the poisons roll on spell hit
  (16% cap). Gear hit rating raises both: `-print-stats` on Combat Swords shows MeleeHit and SpellHit together at
  8 / 10 / 12 on steps -2 / 0 / +2.
- Below the set (-steps), the hit is paid back as the spec's best stat (Agility or Attack Power for swords: step -2 has
  +11 Agility and +31 AP against step 0). That trade is about even, so the curve is flat between the cap step and 0.
- Above the set (+steps), the hit is paid for with the least valuable stat first (`give` in `hit_steps.json`: Strength
  first). Step +2 on Combat Swords is -22 Strength (-22 AP) for +2% hit, a cheap trade. That is the jump at +1.
  It says Strength is the stat to drop for hit on these sets. It is not a talent effect.
- Seal Fate Swords and Seal Fate Daggers dip slightly from -2 to 0 (943.3 → 941.6, 904.7 → 903.9) before rising.
  That is within a few standard errors, but on both. Seal Fate Daggers' -steps buy crit first and Seal Fate Swords'
  buy Agility (`hit_steps.json`), so Seal Fate seems to rate those a bit above hit near the yellow cap.

## Precision with every step

The talent climb (below) scored each build on the steps within 2 of its own cap step, which (before a6c243fd5)
never reached the + steps for the cap -2 to -4 builds. So each final build was also scored against itself with 1 to 3
Precision points moved anywhere valid. Weapon Expertise (2/2) needs Precision 3/3, so it goes too, and all the freed
points are placed (every single-talent placement plus random spreads, valid only). The check used `-hit-span 6`, so
every step file was tried, at 3000 iterations, with the best 4 plus the build at 10,000 (combined 120/180/300):

| Build | Candidates | Current (best step) | Best without full Precision | Loss |
|---|---|---|---|---|
| Combat Swords | 30 | 965.7 | 951.3 (`005303202011-3211320112010321551`) | -14.4 |
| Combat Daggers | 16 | 955.3 | 940.3 (`005203002-3213220102010321551-005`) | -15.0 |
| Seal Fate Swords | 12 | 946.1 | 934.7 (`0053032024105103-320132021201032`) | -11.4 |
| Seal Fate Daggers | 10 | 909.7 | 898.0 (`0053032024105103-321312020201032`) | -11.7 |
| Subtlety | | | pending (still running at hand-back) | |
| Mutilate | | | pending (still running at hand-back) | |

Precision 3/3 + Weapon Expertise is worth 11 to 15 DPS over anything the freed 5 points buy, even when the gear may
put the hit back. In Forever's tree Weapon Expertise has Precision as its prerequisite. Several candidates tie to
the decimal: points that land in talents the rotation never uses.

## The talent climb

As in round 70753 (`orchestration/round_70753/rogue.md`, "Searches"): rotopt's one-point talent search stalls on
the row gates and on Weapon Expertise's prerequisite, so this used a stochastic climb.

- Each round: 60 random valid neighbours of the current build, each moving 1 to 5 points (30% of them always touch
  Precision), never taking a point from the pinned talent. Validity is rotopt's rule: 5 points per row above,
  prerequisites maxed, 51 points.
- Each neighbour was scored by `rotopt -talents-file` at 3000 iterations over 120/180/300 s combined, with
  `-hit-steps <set> -hit-cap melee:8 -hit-span 2`. That is the brief's flag set, so each build wears the best of its
  cap step ±2.
- The best 5 plus the current build were scored again at 10,000. A move was kept only if it beat the current build
  by 2 standard errors of the difference. The climb stopped after 2 rounds in a row kept nothing.
- Knobs were held at each rotation's preset knobs (`tools/rotopt/rogue_presets.sh`), with no knob search, since this
  phase does not touch the rotations.

Every build stopped after 2 rounds with nothing kept. Best neighbour at 10,000 (combined) against the current build:

| Build | Round 1 | Round 2 |
|---|---|---|
| Combat Swords | 960.8 vs 960.8 | 960.8 vs 960.8 |
| Combat Daggers | 947.5 vs 947.1 | 947.5 vs 947.1 |
| Seal Fate Swords | 943.3 vs 943.3 | 943.3 vs 943.3 |
| Seal Fate Daggers | 904.7 vs 904.7 | 904.7 vs 904.7 |
| Subtlety | 881.7 vs 881.7 | 881.7 vs 881.7 |
| Mutilate | 859.8 vs 859.4 | 857.6 vs 859.4 |

The exact ties are filler points moved between talents the build does not use.

## Odd things

- **The steps that win are + steps, and three builds pick +3, the highest file the sets have.** On Seal Fate Swords
  (+0.1), Seal Fate Daggers (+0.8) and Mutilate (+0.5), +3 is barely above +2, so the curve is flattening and +4
  would add little. But the search did want the last step.
- **Combat Daggers cannot reach its cap step.** Its total hit is 14% (the set carries Flask of Natural Accuracy), so
  the cap step is -6, but the set's steps stop at -4. The tool clamps the cap step to -4. Moot, since it wants +2.
- **The brief's `-hit-span 2` around the cap step misses the best step on every rogue build**, because rogues gain from
  hit up to well past the yellow cap. The coordinator says a6c243fd5 now also tries steps past the set's own step.
  The step scans above cover the whole range either way.
- The + steps' gain is a pricing question (they drop Strength for hit) as much as a hit question. If the sets were
  rebuilt with that Strength already traded for hit, step 0 would be the peak.

## Commands

Tools: `go build --tags=with_db -o $S/rotopt ./tools/rotopt` (it needs the generated `sim/core/proto/*.pb.go`:
`make sim/core/proto/api.pb.go` with protoc and protoc-gen-go). The scripts are in `rogue_scripts/` next to this file
(`climb.py`, `precheck.py`, `finish.sh`, `confirm.sh`). They expect `rotopt` in their folder, `S` set to that folder
and an `out/` folder in it, and they run from the repo root. `$K` is the build's knob string from its `gen60` line
in `tools/rotopt/rogue_presets.sh`. `<gs>` is the gear set suffix (swords, combat_daggers, seal_fate_swords,
daggers, subtlety, mutilate). `<tpl>` is rogue_ss, rogue_bs, rogue_bs, rogue_bs, rogue_sub, rogue_mut in the same order.

Climb (per build):

    N=60 python3 $S/climb.py rogue__<id> <tpl> "$K" ui/rogue/gear_sets/level60_phase1_synthetic_<gs> <pin> <current talents> $S/out

which runs, per round:

    $S/rotopt -template <tpl> -settings tools/dps_report/req/rogue__<id>.json -knobs "$K" -combined -durs 120,180,300 \
      -iters 3000 -talents-file <candidates> -hit-steps ui/rogue/gear_sets/level60_phase1_synthetic_<gs> \
      -hit-cap melee:8 -hit-span 2          # and again with -iters 10000 on the best 5 plus the current build

Step scan (each step N from -7 to +3 whose file exists):

    $S/rotopt -template <tpl> -settings tools/dps_report/req/rogue__<id>.json -knobs "$K" -talents <talents> \
      -combined -durs 120,180,300 -eval -confirm-iters 10000 \
      -hit-steps ui/rogue/gear_sets/level60_phase1_synthetic_<gs> -hit-cap melee:8 -hit-step <N>

Confirm, for d in 120, 180, 300:

    # old
    $S/rotopt -template <tpl> -settings tools/dps_report/req/rogue__<id>.json -knobs "$K" -durs <d> -eval -confirm-iters 20000
    # new
    $S/rotopt -template <tpl> -settings tools/dps_report/req/rogue__<id>.json -knobs "$K" -talents <talents> -durs <d> \
      -eval -confirm-iters 20000 -hit-steps ui/rogue/gear_sets/level60_phase1_synthetic_<gs> -hit-cap melee:8 -hit-step <best>

For example, Combat Swords new at 180 s: `-template rogue_ss -settings tools/dps_report/req/rogue__combat-swords.json
-knobs "$K" -talents 00530320201-3200330112012321551 -durs 180 -eval -confirm-iters 20000 -hit-steps
ui/rogue/gear_sets/level60_phase1_synthetic_swords -hit-cap melee:8 -hit-step +2` → 964.3 ± 0.3.

Precision check: `python3 $S/precheck.py rogue__<id> <tpl> "$K" ui/rogue/gear_sets/level60_phase1_synthetic_<gs>
<pin> <talents> $S/out` (`-hit-span 6`, 3000 then 10,000 iterations, combined).
