# Fairness pass 15 (2026-10-30, master d50909a)

Result: quiet (no new finding, no open finding). Full pass: gear check, full race sweep, new baseline.

## Why it ran

Since 334c7de, 64 commits, nearly all the raid sim (M0 to M2): sim/compsearch (new), and shared sim/core changes
(camp buffs, per-player raid buffs, cast weapon totems, party Blood Pact, Dark Inspiration, Mana Tide per member,
Battle Shout / Furious Howl labels by value, shared Judgements of the Crusader and Wisdom, attack tables by unit
index), plus class code for Seal of Wisdom (paladin), caster Faerie Fire in forms (druid, Feral preset), Hunter's
Mark handover and pet Furious Howl, shaman totem receivers, warrior shouts, warlock, priest. 19 new duty / pet
variants in tools/dps_report/req (0bb3ec9b6).

## What I checked

- Gear: `audit1b.py` passes 5a and 1b on all 19 lines.
- Race sweep (5k, seed 1, race overrides applied): every one of the 37 existing builds at every race is identical
  to the dd663e6 baseline (0.0 everywhere). The raid sim's sim/core changes leak nothing into single-player presets.
- Published against current: shared 20k run (`vs_published.txt`) has 76 of 77 within 0.5; Frostfire +0.9, the
  known 8624ba126 size, inside tolerance.
- New pet variants (BM / MM / Survival x Tallstrider / wolf): Orc best for each, sweep within 0.3 of published.
- Duty variants (JoW / JoC, Recklessness, Sunder, Faerie Fire) are raid-sim role builds; not in the race sweep.
- Best race against report race: Fire Gnome +1.1, Affliction Orc +0.1, the 20k-measured ties (passes 11, 12).

## Tool fix (third time)

The race-override fix to `tools/audit/fairness/sweep.py` (apply each race's req/races/ override through the tree's
`race_override.apply`, DASH_DIR pointed at AUDIT_TREE) was again missing from master; re-applied in the working
tree for the run's commit. Without it Dwarf, Gnome rogue, Smite and High Order rows read 2 to 15 DPS low.

## Ranking

Unchanged from pass 14 (top BM Orc 1131.8, bottom Subtlety 717.1; 1.58x). Full table in fairness_report.md §1 and
ref/fairness/table_baseline.md. The baseline is now d50909a.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, Flametongue Totem with the Windfury imbue, boss partial resists,
Rage k and pet scaling (see passes/fairness.md).
