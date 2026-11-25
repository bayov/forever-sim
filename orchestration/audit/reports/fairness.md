# Fairness pass 59 (2026-11-25, master 0404484fb)

Full pass: sim/core, warrior, shaman totems and standards.md changed (range 5536db4bc..0404484fb). 5k race sweep
(seed 1, every race, the report's own requests with only the race changed) on the shared 0404484fb CLI, table.py
against the 63129ce35 baseline, phase0_check.py, and the loop's 20k run against published.

## What moved, and why

| Commit | Rule | Effect on report builds |
|---|---|---|
| f1942793d, b3d294a00 | Arcane Intellect / Divine Spirit overwrite Int and Spirit elixirs (Cunning's Int too) | casters, hunters, paladins, shamans down: Enh -25.9, Ret -16.0, Arcane -13.4, Frostfire -13.6, warlocks -6 to -9, hunters -7 to -9 |
| b4cab8498 | Rage from hits taken by stance; spell hits and ticks twice | warriors up: Fury DW +5.5, Fury 2H +4.3, Arms +2.8 |
| 3275a60b1, 9701cff08 | world buffs (and Zandalar) off in raids | none (no report build ticks one; no report hash moved) |
| 75dee7458 | Armor Shatter / Puncture Armor share the minor armor slot | none (no report hash moved) |
| 78802e6e5, 02cf1139b (earlier, now re-simmed) | Feral trinket APLs | Feral +2.5, already published |

Rogues and Feral match the baseline within noise (largest +0.8). No class whose code and consumables were
untouched moved: no shared change leaked. Each rule applies to every spec that has the stat or resource; only the
report has not been re-simmed. Enhancement and Ret lose most because Elixir of Cunning's Intellect now counts as
an Intellect elixir; Mongoose would give the same (their Sages / Owl already carry the one elixir crit; paired 20k,
Cunning = Mongoose to 0.1).

## Published against current (confirms results-20261104-2, updated today by the results pass; not filed again)

Published numbers (re-simmed at 226ada86b) lag 0404484fb by -26.7 (Enh) to +6.0 (Fury 2H). The 20k run lists 61
builds outside noise. At current numbers Frostfire drops below Seal Fate Swords, Arcane below Seal Fate Daggers and
Balance below Mutilate. Fix: regenerate the report for the moved hashes.

## Race

- Marksmanship High Order +6.6 over Orc (fairness-20261125-1, still open); MM variants +7.1 / +7.7.
- Fire Undead +0.8 (was +2.9; near noise now), Shadow Troll +1.7 (fairness-20261124-2, still open).
- Balance: Night Elf best on the base rotation; Wrath High Order +4.7 (fairness-20261118-1, still open).
- Fury DW Undead +1.1 over Orc, High Order +0.7: a spread of 3.6 over eight races at 5k, within noise; not filed.
- Demonology Human +0.9, Frostfire Gnome +0.6, Lone Wolf High Order tie: noise. sub_holy Dwarf +1.7 and
  affliction_shadow_bolt Human +2.4 as before (variants).

## Gear

phase0_check.py: 3 '!' lines (Balance finger, Shadow chest, Shadow finger), already in the gear pass's report.
Every DPS request wears its spec's Phase 0 set. The lists' `dps` predate the elixir rule, so they now sit 5 to 27
above the current sims for the Int users (part of the same re-sim); rogues, Feral and warriors match within 2-6.

## Ranking (report race, 5k)

Top over bottom 1.69x (Lone Wolf 1164.8 / Smite 690.6). Full table: orchestration/audit/ref/fairness/table_baseline.md
(new baseline, 0404484fb).
