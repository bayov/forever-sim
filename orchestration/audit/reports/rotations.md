# Rotations pass 71 (2026-11-22, master 8b5137c6d)

Not quiet: one rotation changed and the sim changed for paladins, rogues, warriors and shamans.

## What changed (b4b68c52e..8b5137c6d)

- APL logic (`apl_logic_diff.py`): only `ui/retribution_paladin/apls/level60_undead.apl.json` (1b4a74f33, Holy Strike
  moved ahead of Judgement and the seal lines after the Nov 19 Two-Handed Weapon Specialization hotfix).
- Sim: 2H Weapon Specialization raises all Holy damage (1c80dafff), Shank procs poisons/enchants (6de6861c5), Flurry
  hotfix in warrior talents, Stormshroud 3P, shaman Stormstrike/talents and item sets, Barov Peasant Caller (abefd2f86).

## Baseline (3000 it, seed 424242), new file `ref/rotations/baseline_8b5137c_3000it_seed424242.tsv`

Rows that moved by 0.05% or more against baseline_a667736:

| preset | old | new | move |
|---|---|---|---|
| paladin level_60 / -crusader / -wisdom | 1021.9 / 1014.3 / 1005.6 | 1026.8 / 1019.0 / 1010.5 | +0.48% |
| paladin shock_ret (+ duties) | 1007.6 | 1016.9 | +0.92% |
| paladin sub_holy (+ duties) | 1007.9 | 1012.9 | +0.5% |
| paladin sub_holy_undead | 1137.8 | 1150.7 | +1.14% |
| rogue combat-daggers / combat-swords | 936.3 / 943.6 | 940.7 / 949.1 | +0.47 / +0.58% |
| rogue mutilate / sf-daggers / sf-swords / subtlety | 865.8 / 903.1 / 940.1 / 872.1 | 869.9 / 908.0 / 941.7 / 875.8 | +0.16 to +0.55% |
| warrior fury, fury-bs, fury-sunder, fury_2h, 2h-sunder | | | +0.05 to +0.32% |

No shaman, hunter, druid, mage, priest or warlock row moved. All moves are the sim changes above.

## Checks

- **Undead Ret reorder (1b4a74f33) verified.** Paired ab.py, B = the old order, B-A at 120/180/300 s:
  set 1e6 -0.26 / -0.47 / -0.72, set 2e8 -0.55 / -0.92 / -0.87. The new order wins on every cell (small, as the commit says).
- **Shock Ret Holy Strike placement** (the only Ret build still running Holy Strike late): filed
  rotations-20261122-1 (HoW then HS right after Consecration, pooled +1.46 / +2.25 / -0.23). Holy Strike alone moved up is a
  fight-length trade-off (+3.6 at 180, -2.3 at 300); mana gates on it make 300 s worse (-4.5 at 30%, -12.4 at 50%).
- Open finding rotations-20261112-1 (rogue poison swap) has not landed: every rogue req still has Deadly Poison on the main hand.
  Shank now procs poisons, so the gain may have shifted a little for builds that use Shank; not re-measured.

## Optional

- None new.

## Not done this pass

- Shock Ret duty variants (crusader, wisdom) not A/B'd for the finding's reorder; they share the order.
- Paladin hasted-GCD re-timing item: its trigger ("any paladin rotation changes") fired with 1b4a74f33; the re-timing
  itself is a worker search, not done here.
- Warrior HS/Cleave threshold re-search and Windfury ability-timing items: unchanged.
