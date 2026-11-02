# Rotations pass 45 (2026-11-02)

- **Master:** 5a6e9c6a6 (last audited c23ef0440). Client 1.60.1.70707 (+ the launch realms' 2026-10-30 hotfix rows).
- **Result:** quiet. No new finding. rotations-20261101-1 (Frost 4th Arcane Blast stack) moved to Resolved as obsolete.

## What changed

- `apl_logic_diff.py c23ef0440 origin/master`: 17 APL files, all from the re-searches on the October 30
  hotfix trees (rotopt templates + `*_presets.sh`, by workers):
  - Ret paladin (all 15 presets): Unsealed Wrath twisting replaces the Twist of Light Echo lines (seal swap
    on `auraNumStacks` of 1311703 / 1311704); Shock presets drop the 1311595 lines and cast Holy Shock unconditioned.
  - Feral (2): Tiger's Fury 417141 lines gone, Shifting Power at Energy <= 60, Rip at 4+ CP, Bite at 45 Energy.
  - Balance Wrath (2), Arcane (Arcane Wrath 1326082 with Presence of Mind), Frost (Ice Barrier 13033 on
    Frigid Power 1326891 < 6 s, pre-pull at -4 s; Ice Lance on the last Fingers of Frost charge).
- Sim: `sim/paladin/*` and `sim/druid/*` hotfix talent rules, `sim/hunter/hunter.go` (tree position swap only),
  `sim/compsearch/presets.go`. Hunter req changes are the remapped BM talent string only.

## Regression

`baseline.py` on the shared CLI (3,000 it, seed 424242) against `baseline_c23ef04`. Talents, APLs and
reqs moved together, so these are the re-searches plus the hotfix rules, as their commits report:

| preset | before | after | delta |
|---|---|---|---|
| paladin level_60 (twist), duties | 960-976 | 1026-1040 | +64 to +66 |
| paladin sub_holy, duties / undead | 964-978 / 1069 | 1023-1038 / 1152 | +59 to +61 / +83 |
| paladin shock_ret, duties | 949-960 | 965-975 | +15 |
| paladin shock (deep Holy), duties | 904-913 | 846-855 | -54 to -62 |
| paladin shock_undead | 994.6 | 972.6 | -22.0 |
| druid feral-cat, duty-ff | 882-886 | 937-942 | +55 / +56 |
| druid balance, wrath, duties | 776-847 | 824-892 | +43 to +49 |
| mage arcane / frost | 894.4 / 930.0 | 939.1 / 945.4 | +44.8 / +15.4 |

Every other preset (hunters, warriors, rogues, warlocks, priests, shamans, fire / frostfire) within 0.1.
The deep Holy Shock drop matches the commit (306f78678: 908.4 -> 855.6 at 180 s on the hotfix tree).
New baseline: `ref/rotations/baseline_5a6e9c6_3000it_seed424242.tsv`.

## Checks

- **Dead lines** (new helper `tools/audit/rotations/dead_lines.py`, 300 it) on every druid, paladin, mage
  and hunter req. Nothing unexpected:
  - Frost: Arcane Blast (1239700) and Missile Barrage lines inert, talents gone (hence the resolved finding).
  - Arcane: the four "Never" mana gem / potion lines (by design). Paladin: Exorcism (non-Undead), Judgement
    20271 reads 0 because its casts are booked under the seal's judgement id (it does fire).
  - Hunter BM / MM / LW: Mongoose Bite and Strider Kick weaves (dodge / talent dependent, by design).
- **Frost Evocation** (Frost wands about 10 shots a fight at the end, has no Evocation line): adding
  Frostfire's Evocation line before the end-of-fight instant, S0 = 1e6, 20k it:
  | gate | 120 s | 180 s | 300 s |
  |---|---|---|---|
  | mana < 10%, > 75 s left | +0.00 | +0.00 | +0.01 |
  | mana < 20%, > 60 s left | +0.00 | +0.00 | -1.32 ±0.06 |
  | mana < 5%, > 45 s left | +0.00 | +0.00 | -0.02 |
  Frost runs dry only near the end; no gain.
- **Deep Holy Shock priority** (paladin__level_60_shock, S0 = 1e6): Holy Shock above Consecration
  -3.82 / -2.31 / +0.26; above Holy Strike -0.06 / -0.30 / +0.29. The current last-place line holds.

## Optional

None new. The earlier list in `passes/rotations.md` stands (the Frost Arcane Blast item is moot).

## Not done

- No full A/B screen of the re-searched paladin, druid and mage APLs; rotopt searched their knobs at
  120 / 180 / 300 s with 20k iterations, and the spot checks above found nothing. A next pass could screen
  the new Ret Unsealed Wrath twist lines and the feral Shifting Power gate by hand.
- Stale text, not a finding: `forever-wiki/classes/mage.md:62` still says Frost keeps three Arcane Blast
  stacks and drops Evocation; the Frost APL's Arcane Blast note still says 15% / 15 sec (both inert).
