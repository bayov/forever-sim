# Rotations pass 55 (1 finding)

- Master: a07941531 (from 0facd9428, 86 commits), audited 2026-11-05. Client 1.60.1.70753 plus server hotfix
  push 113403 (Ferocious Bite).
- Not quiet: sim `.go` changed (`sim/druid/ferocious_bite.go`, `sim/core/consumes.go` explosives in form,
  `major_cooldown.go` BonusArmor trinkets as survival CDs, `spell_result.go` heal threat in replays,
  `character.go` comment). APL logic changed in 8 files, all tank rotations (feral tank, Protection paladin,
  tank warrior, Warden), out of scope. No Level 60 DPS APL changed in logic.
- Req changes: feral cat and its FF duty take the Goblin Sapper Charge (e77fb293b); Survival (3 presets)
  talents 502-005005-... -> 502-0050051-500230230250022151 (9660fffa8); meta/notes edits otherwise.
- Regression baseline refreshed: `ref/rotations/baseline_a079415_3000it_seed424242.tsv`.

## What moved (baseline, 3k it, seed 424242, vs 0facd94)

| Builds | Move | Cause |
|---|---|---|
| Feral cat FF duty / feral cat | +5.3 / +3.3 | ff56f8ddc hotfix 113403 (FB extra-Energy damage +20-24%), e77fb293b Goblin Sapper in Cat Form. |
| Survival wolf / tallstrider / plain | -1.6 / -1.4 / +0.2 | 9660fffa8 talents (Rapid Killing for Resourcefulness). Within 3k-it noise for a talent change (RNG stream shifts); not A/B'd, since ab.py varies only the APL. |

Every other row moved under 0.05 DPS. The `vs_published.txt` moves on rogue, Ret, shaman and arcane predate
this pass's range (pass 54 attributed them).

## Screens (`ab.py`, 20k it, 8 blocks, shared CLI a0794153163a)

Feral cat, 5-CP Ferocious Bite energy gate (now 45), B-A at 120/180/300:

| Preset | Gate | S0 | 120 | 180 | 300 |
|---|---|---|---|---|---|
| feral-cat | 35 | 1e6 | -2.61 | -2.92 | -3.41 |
| feral-cat | 55 | 1e6 | +1.25 | +1.18 | +2.10 |
| feral-cat | 55 | 2e8 | +1.40 | +1.22 | +1.92 |
| feral-cat | 65 | 1e6 | +1.63 | +1.86 | +3.28 |
| feral-cat | 65 | 2e8 | +1.80 | +1.96 | +3.12 |
| feral-cat | 75 | 1e6 | -0.42 | +0.19 | +1.85 |
| feral-cat-duty-ff | 65 | 1e6 | +0.70 | +0.89 | +2.06 |
| feral-cat-duty-ff | 65 | 2e8 | +1.18 | +1.33 | +2.06 |

Gate 65 -> finding rotations-20261105-1.

## Optional

- None new. Pass 26's optional list in `passes/rotations.md` still stands.

## Open items

- Survival pet presets' -1.5 after the talent change was not measured paired (needs a talent A/B, not an APL one).
