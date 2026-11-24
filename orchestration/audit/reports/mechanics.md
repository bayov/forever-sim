# Mechanics pass 102 (2026-11-24, master eceb09019, client 1.60.1.71091)

Range: e29cfe00f..eceb09019. Why it ran: 9 watched files changed, and client build 71037 -> 71091.

## Sim commits (shared engine, racials)

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 8d5fdd664, 7a2311b36 | Troll Regeneration 20%/20%, ticks every 2 s from the first damage taken; Forever Stoneform | 20555 aura 88/116 bp 20; 20594 aura 87 bp -10 Physical, 8 s, 3 min, GCD | matches 71091 (tier 1) |
| ed1df1d51 | Innervate and Energized scale MP5 too; BoW and Mana Spring energize left unscaled | 29166 aura 110 bp 400; wowsims/forever EV-0114 ledger (70938 logs) | accepted (log); Energized's MP5 share inferred |
| 9c08e16a2 | Blood Fury multiplies SpellDamage | 20572 aura 317 bp 10 | right |
| 0352d95dd, e19d2df85 | Shaman L60 class row, Skyborne race offsets | wowsims/forever StatLedger sheets (EV-0181, EV-0093) | accepted (log data) |
| 18ecefc36 | comments only | - | - |

## Client 71091 vs 71037

69 spells changed (`wago.py diff --by-spell`). Shared ones: Execute, Mutilate, Judgement 20271 and Mind Vision gain
Attributes_4 0x800000 (suppress weapon procs), Shank loses it; JoL/JoW/JoJ DurationIndex 64; Revitalize 23864 3% / 18
energy (already in crafted.go); Nature's Bounty 27784 power type 3 (its melee part is not modelled). The rest are
class deltas (druid 128xxxx/1331xxx spells, rogue Shank/Boot Knife, hunter Serpent Sting/Multi-Shot), all listed in
TASKS.md's open "build 71091" item for the client update and class workers; none is filed here.

## Blizzard

- Hotfix post t/2403040/15 (created 2026-11-24, notes of Nov 23), read: Darkmoon Faire buffs halved (new finding
  mechanics-20261124-1, server side: 71091 still has Sayge's 10%), Fengus' Ferocity halved (PR #139), extra weapon procs
  fixed (the 0x800000 rows above; Execute's Windfury roll stays a TASKS item), Intensity on Incinerate, Mind Vision procs.
- Dev notes t/2360696: newest still post 10 (2026-11-04). Forum JSON was rate limited this pass; the blue tracker was
  not re-read.

## Findings

- New: mechanics-20261124-1 (low): Darkmoon Faire fortunes still 10%. No preset uses them.
- Open findings unchanged.
