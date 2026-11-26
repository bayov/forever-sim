# Fairness pass 60 (2026-11-25, master 85691655a)

Full pass: sim/core, every class, the item database and standards.md changed (range c3bbe33e..85691655a, the
wowsims/forever evidence batches). 5k race sweep (seed 1, every race, the report's own requests with only the race
changed) on the 85691655a CLI, table.py against the 0404484fb baseline, phase0_check.py, the loop's 20k run against
published, paired 20k race checks and a paired explosive check (new helper tools/audit/fairness/explosives.py).

## What moved, and why

| Commit | Rule | Effect on report builds (5k sweep vs 0404484fb) |
|---|---|---|
| 60fde111f | pet white swing takes the hunter's crit (EV-0134) | Beast Mastery +41.7, Survival +29.7, MM +12.1 (Lone Wolf has no pet) |
| 295206ddf, b663145ad | Forever weapon PPMs, dungeon set procs, proc chains, JoC double roll, judgement refresh | rogues -11.6 to -26.4 (set procs and weapon PPMs), Ret / Enh small |
| efeb06b95 | resist constant 8L+19, binary spells, overloads, Flurry | Shadow +13.3, Elemental +14.2, Fire +5.9, Arcane +2.4 |
| 22d2cf952 | Dark Bomb misses stun-immune bosses, demon crit, extra attacks roll WF Totem, Shield Slam / Sweeping Strikes | Lone Wolf -12.5 (Dark Bomb now wasted), Fury DW -13.0 net with the procs batch |
| 2f42f4a44 | tank threat and damage-taken rules | DPS builds within noise |

Every move matches a commit that touches the class's mechanics, items or a shared rule it uses (the results pass's
bisect agrees). No class moved without a cause: no leak.

## Published against current (results-20261104-2 and results-20261125-1; not filed again)

The report lags by -27.8 (Enhancement) to +34.9 (Beast Mastery). Confirmed independently by this sweep. At current
numbers BM goes from 6th to 4th, Survival passes Ret, Shadow and Elemental pass Frostfire and the rogues, and
Mutilate drops to 26th. Fix: the full re-sim.

## Explosives (new, fairness-20261125-2, high)

Standards 11 now says the explosive picks are due a re-check (throws keep the swing timer, stun bombs miss bosses).
Only hunters and Feral were re-picked. Paired 10k, seeds 1 / 1000003, Goblin Land Mine over none:

| Build | Mine | Dynamite | Build | Mine | Dynamite |
|---|---|---|---|---|---|
| Fury DW | +3.3 | +2.5 | Ret | +7.3 | +6.5 |
| Fury 2H | +2.7 | +1.9 | Shock Ret | +6.8 | +6.0 |
| Arms | -1.4 | -2.1 | Sub Holy Ret | +7.1 | +6.3 |
| Arms dragonkin | +0.3 | -0.5 | Enhancement | +6.3 | +5.5 |
| Mutilate | +6.8 | +5.8 | Enh imp-gw | +5.0 | +4.2 |
| Combat Swords | +6.5 | +5.3 | Survival | +4.6 | +3.7 |
| Combat Daggers | +6.2 | +5.6 | Survival wolf | +5.2 | +4.3 |
| Seal Fate Swords | +6.9 | +6.1 | Fire / Affliction / Elemental | +0.0 | -5 to -6 |

Subtlety and Seal Fate Daggers already carry the mine (worth +7.0 / +6.6). With it, Mutilate passes Subtlety.

## Race

- Demonology: Undead +5.7 over Orc, paired 20k (fairness-20261125-3, new); Destruction Undead +2.2 (same finding).
- Marksmanship High Order +5.0 over Orc; wolf +6.7, tallstrider +6.4 (fairness-20261125-1, still open).
- Shadow: Undead (the report race) best again by 3.7 paired; Fire Undead +1.05 / Gnome +0.45 over Human paired, at the
  edge of noise (fairness-20261124-2, now Fire only).
- Balance: Night Elf best on the base rotation; Wrath High Order +3.8 (fairness-20261118-1, still open).
- Frostfire Gnome +1.6 over Human (5k), Fury DW Undead +1.4 over Orc (5k spread over 10 races 13), Ret variants
  Dwarf +2.8 / +3.0 over Undead (variants): noise or variant-only; not filed.

## Gear

phase0_check.py: the same 3 '!' lines (Balance finger, Shadow chest, Shadow finger), the gear pass's. Every DPS request
wears its spec's Phase 0 set. The lists' `dps` predate this range's rules and the elixir rule; part of the re-sim.

## Ranking (report race, 5k)

Top over bottom 1.66x (Lone Wolf 1152.3 / Smite 692.8). Full table: orchestration/audit/ref/fairness/table_baseline.md
(new baseline, 85691655a).
