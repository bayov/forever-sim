# Fairness pass 67 (2026-11-27, branch phase0-regear 313069ce)

Scoped pass on the branch head, not master (master was 046d0e4b8). Range 046d0e4b8..313069ce: the 29 report
requests re-geared to their current Phase 0 gear sets (94d6cc79, race rows and verify 48cc3aed) and re-published
(313069ce); the Ret trinket re-rank (ee28c064, RANK_ONLY SLOTS=trinket, picks kept); a comment-only change to
Talisman of Ascendance (8d45b49e). No race sweep: no sim, class or database change; the 2af970d74 sweep stays the
baseline. Two findings resolved, none new.

## The 29 re-geared requests (fairness-20261127-1, resolved)

phase0_check.py at 313069ce flags no DPS request ("equipment is not <set>" and the out items are gone). Published
moves (builds.json 046d0e4b8 -> 313069ce) against pass 66's paired gear-set measurement:

| Build | Before | After | Move | Measured |
|---|---|---|---|---|
| Affliction | 962.5 | 945.1 | -17.4 | -17.2 |
| Affliction duty-curse | 948.1 | 935.3 | -12.8 | -12.4 |
| Affliction Shadow Bolt | 818.4 | 811.2 | -7.2 | -8.0 |
| Demonology | 873.9 | 867.1 | -6.8 | -6.6 |
| Destruction | 866.3 | 866.1 | -0.2 | +0.2 |
| Balance | 801.7 | 795.5 | -6.2 | -5.6 |
| Feral | 1057.5 | 1054.0 | -3.5 | -3.3 |
| Fury 2H | 994.2 | 991.3 | -2.9 | -2.8 |
| Elemental | 855.7 | 857.6 | +1.9 | +1.7 |
| Enhancement | 966.0 | 972.3 | +6.3 | +6.1 |
| Smite | 678.7 | 678.6 | -0.1 | -0.0 |

Duties and variants move with their base (Demonology duties -5.4 / -5.1, Destruction duties +0.8 / +0.9, Destruction
Fire +0.5 to +1.4, Balance FF -5.1, Wrath -2.5 / -1.4, Feral FF -3.4, Fury 2H duties -2.7, Enhancement imp-gw +6.3,
mana-tide +5.9, water-shield +6.2). Exactly the 29 builds moved. All 29 requests simmed at 313069ce (10k, seed 1) are
within 0.9 of builds.json, and every meta.published.dps equals builds.json. Ranking: Affliction 9th -> 11th,
Enhancement imp-gw 10th -> 9th; Demonology duties drop one or two places, Destruction Fire duties rise one.

## Ret trinkets and Maelstrom (fairness-20261126-4, resolved)

Paired 10k, seeds 1 / 1000003, paladin__level_60 (the Ret list's build), 313069ce CLI:

| Run | Set (Chained Essence + Breadth) | Maelstrom for Breadth | Maelstrom for Chained Essence |
|---|---|---|---|
| Report fight, 180 +/- 20 s | 931.75 / 932.16 | 934.23 / 934.74 (+2.5) | 935.81 / 936.37 (+4.1) |
| Two-part score (standards 43) | 943.55 / 943.66 | 942.25 / 942.77 (-1.1) | 938.87 / 939.61 (-4.4) |

Use values: Chained Essence 16.2 on 90 s vs 8.0 on 180 s; Blackhand's Breadth 6.6 vs 3.2. The two-part score for
Maelstrom-for-Breadth reproduces rotopt's 20k confirm (-0.9). So the list follows the rule, and the report fight's
preference for Maelstrom is the rule's 90 s use scoring, not a Ret defect. (The Shock Ret +2.4 is the same effect.)

### The 90 s use scoring across the DPS lists (rough, not simmed)

All 26 DPS lists score at 180 s with use_dur 90, so every spec gets the same rule. From each list's own numbers
(pick's `cooldown.loss - loss` is its 90 s use value; a passive non-pick wins the 180 s fight roughly when its
loss is under half that), these picks would likely lose to a passive trinket in the report fight:
Ret (Maelstrom, ~-10 / -5), Feral (Behemoth Bones, ~-9), Affliction (Bolt's Charged Eyeball, ~-8), Lone Wolf
(Heart of Wyrmthalak, ~-7), Smite, Destruction, MM, Demonology, Elemental, rogues, mages, Balance, Fury DW
(~-1 to -5 each); Arcane's Royal Seal estimate (-12) looks too large to trust without a sim. The effect is broad and
by ruling (the user's 2026-11-20 short-fight call, standards 43), so it is not filed. If the user ever wants the
report to show each spec's 180 s-best trinkets, this is the list to start from:
/tmp/claude-0/aud-fair/usegap.txt (scratch; script /tmp/claude-0/aud-fair/bin/usegap.py).

phase0_check.py still prints "ret trinket: open Darkmoon Card: Maelstrom sits 1.85 above the pick": that is the
rank pass's number, which the 20k confirm overrides.

## Other checks

- phase0_check.py: no DPS-request flag; remaining "open above pick" lines are tank lists, Destruction's main hand
  (1.75) and Enhancement's main hand (1.77), both within the rank's noise and unchanged by this range.
- List dps against the report: Ret's list 941.51 vs 931.5 published; the 10 DPS gap is the 90 s use credit above.

## Open items

- None new. The open race findings (20261118-1, 20261124-2, 20261125-1, 20261125-3, 20261126-2) and the
  Blackblade stun (20261126-3) were not re-measured: this range doesn't touch them.
