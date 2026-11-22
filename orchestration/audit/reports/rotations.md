# Rotations pass 72 (2026-11-22, master 7a6d1beb9)

Not quiet (sim changed for presets), but no new findings.

## What changed (8b5137c6d..7a6d1beb9)

- APL logic (`apl_logic_diff.py`): no APL file touched. No req file changed.
- Sim: Ice Lance 0.11 of spell power under Forever (4192d6e49); Wrack ticks the half points of a dot's grid (3960dd1f6);
  warrior Deep Wounds takes the critting hand's weapon (990cade4e), avoided Execute refunds nothing (aa394d3be), Windfury
  Totem swing gives no Bloodthrill or Flurry charge and Execute rolls it twice (b7edf19bb); Maelstrom gets no stack from the
  totem (840029bd8); haste shortens Flame Shock's GCD (56e5f070d); JoW pays only on landed swings (bd7029410); hunter pet
  keeps its summon-time AP (051a21cd3); item effects (Linken's Boomerang, Cloak of Fire, Dungeon Set 1 reveal sets).
  tools/rotopt Phase 0 files are outside the presets.

## Baseline (3000 it, seed 424242), new file `ref/rotations/baseline_7a6d1be_3000it_seed424242.tsv`

| preset | old | new | move |
|---|---|---|---|
| mage frost | 946.2 | 835.1 | -11.74% |
| mage frostfire | 1065.2 | 1006.9 | -5.47% |
| shaman enhancement (4 builds) | 976.8-1079.4 | 966.9-1067.3 | -0.96 to -1.12% |
| hunter SV / BM / MM (9 builds) | | | -0.14 to -0.43% |
| warlock affliction (+ duties) | 1076.8 / 1058.2 / 1066.2 | 1089.2 / 1070.8 / 1079.2 | +1.16 to +1.23% |
| warrior fury (+ duties) | 1102.7 | 1113.1 | +0.85 to +0.95% |
| warrior fury_2h (+ duties) | 1113.8 | 1129.4 | +1.41 to +1.81% |
| warrior arms (+ duties) | 1078.6 | 1081.7 | +0.17 to +0.28% |
| paladins, other warlocks, elemental | | | within ±0.24% |

Every move matches a sim commit above; none comes from a rotation.

## Checks

- **Ice Lance after the 0.11 coefficient** (the big mover). Paired ab.py, set 1e6, B-A at 120/180/300 s:
  - Frost, B = no Fingers of Frost Ice Lance line: -42.6 / -43.9 / -44.6 (±0.5). The line still pays by far.
  - Frost, B = Ice Lance on any FoF charge (not just the last, with the bolt in flight): -19.7 / -19.7 / -19.3.
  - Frostfire, B = no "Ice Lance on the last FoF charge" line: +1.0 / -11.9 / -9.8. Keep the line.
  - All losses are far past the noise, so the second seed set (only needed to confirm a gain) was skipped.
- Affliction: Wrack is already the filler ahead of Shadow Bolt ("whenever nothing above it is due"); the half-point
  ticks only make it better, and there's no Wrack gate to re-tune.
- Open findings rotations-20261122-1 (Shock Ret HoW/HS order) and rotations-20261112-1 (rogue poison swap) have not landed:
  no APL or req changed.

## Optional

- None new.

## Not done this pass

- Warrior Execute (avoided Execute refunds nothing) and Deep Wounds changes: the Execute and HS/Cleave thresholds weren't
  re-searched; this belongs with the open warrior threshold re-search item.
- Paladin hasted-GCD re-timing and Windfury ability-timing items: unchanged.
- The machine was heavily loaded (load ~38 on 4 cores), so the pass was kept to the mage A/Bs.
