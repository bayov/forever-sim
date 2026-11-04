# Round 70753 phase 5: paladin, talents re-searched with hit-trade gear steps

Worker: paladin class worker of phase 5 (brief `briefs/phase5_hit_trade_class.md`). Branch `phase5-hit-paladin`, from
`phase5-hit-trade` at aed188a4b with origin/phase4-followups merged in first. That brought the paladin hotfix 113542
commit f6a98ac (Vengeance 5 stacks, Two-Handed Weapon Specialization on Holy). It was not on master yet. The commit
moves no DPS: the old numbers below match the published ones to the decimal.

## Result

**No paladin talent string moves.** All four searched builds keep their talents, and their duty variants keep them too.
Only the twist builds (Precision 3/3) change, and only in gear: they wear step **-3** (30 hit rating
traded for 30 crit rating). That brings them from 11% to 8% melee hit and gains **+11 to +13 DPS** at every length. The
sub Holy, vs Undead and Ret Shock builds have no melee hit talent (Divine Precision is spell hit and Holy Strike hit
only, `hit_sweep.md`), so they sit on step +0 and their numbers do not move.

With step -3, **twist now beats the class default (sub Holy) at all three lengths**: 1107.7 / 1103.2 / 1093.3
against 1093.3 / 1091.2 / 1085.8 (+14.4 / +12.0 / +7.5). Before, they tied at 180 sec (1092.1 / 1091.2), which is
why sub Holy is the default. That default call belongs to the phase 5 session.

## Defect found: the ret steps were not in db.bin

On `phase5-hit-trade`, every `-hit-steps` run on a Retribution set panicked with `No item with id: 992300`. The
`hitStepsRet` block that `hit_steps.py` writes into `tools/database/forever_synthetic_items_paladin.go` has no
`Quality` (the base ret items get theirs, plus level, phase, expansion and class, from the file's `retribution()`
wrapper). gen_db's simmable filter drops items below Uncommon, so all 6 ret step items went to `leftover_db.bin`.
Fix committed on this branch (first commit): the step `init` sets the same five fields before appending. **db.bin is
not regenerated here** (the brief says not to touch it). I regenerated it locally only to build my rotopt binary:
`go run tools/database/gen_db/*.go -outDir=./assets -gen=db`, and only the 6 items moved from leftover_db to db.
The phase 5 session must regenerate db.bin, and `hit_steps.py` should get the same fix so a rerun does not lose it.

## Per build

Undead, seed 1, 20,000 iterations, DPS at 120 / 180 / 300 sec (error in brackets). Old = current talents on the
current gear set, no `-hit-steps`. New = picked talents on the step.

| Build | Talents (old = new) | Gear step (file) | Melee hit before / after (cap 8%) | Old | New |
|---|---|---|---|---|---|
| `paladin__level_60` (twist) | `050050002-503-05233101201301351` | -3, `level60_phase1_synthetic_hit-3` | 11% / 8% | 1095.5 (0.5) / 1092.1 (0.4) / 1080.3 (0.3) | 1107.7 (0.5) / 1103.2 (0.4) / 1093.3 (0.3) |
| `paladin__level_60-duty-crusader` | twist | -3, `level60_phase1_synthetic_crusader_duty_hit-3` | 11% / 8% | 1078.0 (0.5) / 1082.9 (0.4) / 1075.9 (0.3) | 1090.0 (0.5) / 1094.8 (0.4) / 1088.3 (0.3) |
| `paladin__level_60-duty-wisdom` | twist | -3, `level60_phase1_synthetic_hit-3` | 11% / 8% | 1069.6 (0.5) / 1075.3 (0.4) / 1069.1 (0.3) | 1081.8 (0.5) / 1086.0 (0.4) / 1080.5 (0.3) |
| `paladin__level_60_sub_holy` (default) | `0512500022-3-05233101201301351` | +0, `level60_phase1_synthetic` | 8% / 8% | 1093.3 (0.5) / 1091.2 (0.4) / 1085.8 (0.3) | same |
| `paladin__level_60_sub_holy-duty-crusader` | sub Holy | +0, `level60_phase1_synthetic_crusader_duty` | 8% / 8% | 1076.9 (0.5) / 1082.8 (0.4) / 1083.3 (0.3) | same |
| `paladin__level_60_sub_holy-duty-wisdom` | sub Holy | +0, `level60_phase1_synthetic` | 8% / 8% | 1069.2 (0.5) / 1075.0 (0.4) / 1075.6 (0.3) | same |
| `paladin__level_60_sub_holy_undead` | `0502520022-2-05233101201301351` | +0, `level60_phase1_synthetic_undead` | 8% / 8% | 1222.9 (0.5) / 1221.4 (0.4) / 1214.1 (0.3) | same |
| `paladin__level_60_shock_ret` | `05125210220021--0523310120130133` | +0, `level60_phase1_synthetic` | 8% / 8% | 1037.5 (0.4) / 1030.2 (0.3) / 1025.5 (0.3) | same |
| `paladin__level_60_shock_ret-duty-crusader` | Shock | +0, `level60_phase1_synthetic_crusader_duty` | 8% / 8% | 1024.5 (0.4) / 1029.3 (0.3) / 1027.9 (0.3) | same |
| `paladin__level_60_shock_ret-duty-wisdom` | Shock | +0, `level60_phase1_synthetic` | 8% / 8% | 1015.5 (0.4) / 1020.0 (0.3) / 1018.7 (0.3) | same |

Gear set files are under `ui/retribution_paladin/gear_sets/` (`.gear.json`). Step -3 re-spends the hit into crit
(`hit_steps.json` "ret": crit first, then Agility, Spell Power). Spell hit stays below 16% on every build (twist 8%
after the step, the others 8%). The `-hit-span 1` searches also tried one step more hit, and none kept it.

## The searches (mean of 120, 180 and 300 sec, 3,000 iterations)

The searched builds are the four that were searched on their own in round 70753. The duties keep their base build's
talents (meta `baseBuild`), as the brief says.

- **Twist**: start 1101.4, best = start. The closest change that touches hit, Precision -1 for Reverence +1,
  loses 6.0 even with the extra step's crit (Precision -2: -11.4). The top swaps that score +0.0 move points between
  talents that do nothing for DPS (Vindication, Deflection, Eye for an Eye, ...), so I kept the string.
- **Sub Holy**: start 1090.4, best = start. Divine Precision -2 for Divine Intellect +2: -4.4.
- **Sub Holy vs Undead**: start 1221.8, best = start.
- **Ret Shock**: start 1030.8. The top swap, Deflection +1 for Benediction -1, scores +0.7 against noise 0.7, and
  the final line keeps the start string.

## Commands

Run from the repo root. `$K_*` are the knobs of `tools/rotopt/paladin_presets.sh` for each build's rotation (twist:
`level60`; sub Holy: `level60_sub_holy`; duties: `level60_crusader_duty` / `level60_wisdom_duty` /
`level60_shock_ret_*_duty`; vs Undead: `level60_undead`; Shock: `level60_shock_ret`). `G=ui/retribution_paladin/gear_sets`.
I ran `rotopt` as a binary built with `go build -tags=with_db ./tools/rotopt` against the locally fixed db.bin (above).

Search (per searched build; Ret Shock: `-require pursuitOfJustice,holyShock`):

    rotopt -template paladin_ret -settings tools/dps_report/req/paladin__<id>.json -level 60 -knobs "$K" \
      -talent-search -require pursuitOfJustice -combined -durs 120,180,300 \
      -hit-steps $G/<gear set> -hit-cap melee:8 -hit-span 1

Confirm, for d in 120, 180, 300:

    # old
    rotopt -template paladin_ret -settings tools/dps_report/req/paladin__<id>.json -level 60 -knobs "$K" \
      -eval -confirm-iters 20000 -durs $d
    # new
    rotopt ... -eval -confirm-iters 20000 -durs $d -hit-steps $G/<gear set> -hit-cap melee:8 -hit-step=<-3 or +0>

## Notes

- Every step the builds need exists (twist needs -3, and the sets have -4 to +2). Nothing hit the end of the range.
- The twist duties' gain (+11 to +12) matches the twist build's, as expected: they wear the same Precision 3.
- Write `-hit-step` with `=` (`-hit-step=-3`): a bare `-3` after a space can be read as a flag.
