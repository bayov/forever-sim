# Rotations pass 81 (2026-11-26, master bfbee78fc)

Result: findings (1 new, low). Range 85691655a..bfbee78fc: the Phase 0 pricing and re-rank rounds (gear, weapons, trinkets,
race rows, rotopt Phase 0 tools) and 35a2e9dca (Feral Land Mine line). No sim change.

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py 85691655a bfbee78fc` | 2 files: feral `level60`, `level60_ff`, line 4 itemId 260817 -> 4395 |
| Non-test sim `.go` changed | none; regression baseline kept (`baseline_8569165_3000it_seed424242.tsv`) |
| Req rotations changed | feral-cat, feral-cat-duty-ff only (same Land Mine line) |
| Open findings landed? | rotations-20261125-1 (Feral Dark Bomb) resolved as a variant (keeps the swing term, measured +0.52 vs +0.58); 20261122-1 and 20261112-1 stay open |
| Phase 0 trinket lines, re-ranked builds | trinket changes in the reqs: Balance 10455 -> 22678 (on-use), Feral 19289 -> 22321 and Ret/Shock Ret/Sub Holy 13965 -> 22321 (both procs, no APL line). Feral and Ret still wear 10455, which their APLs name. Balance APLs still name 10455: rotations-20261126-1 |

## Balance trinket lines (druid__balance, 20k it, 8 blocks, CLI bfbee78fcca3, A 802.6 at 180 s)

| variant | set | 120 | 180 | 300 |
|---|---|---|---|---|
| B3 drop 10455 pre-pull, Elune's Light on cooldown | 1e6 | +0.00 | +0.00 | +0.00 |
| B3 on balance-wrath (A 714.3) | 1e6 | +0.00 | +0.00 | +0.00 |
| B2 drop pre-pull, gate Elune's Light on 22678 | 1e6 | +0.00 | -0.03 | +0.07 |
| B1 pre-pull 22678 and gate on it | 1e6 | -1.76 | -1.21 | -0.63 |
| B1 | 2e8 | -1.73 | -1.17 | -0.73 |

The dead lines cost nothing (B3 is bit-identical); pre-pulling the Talisman loses, as it did for warlocks.

## Optional

None new.

## Still open (pass doc items)

Phase 0 trinkets: done for the builds re-ranked in this range (Balance, Feral, Ret family); other classes were checked by
earlier passes only where their trinkets moved. Warrior HS/Cleave re-search, paladin hasted-GCD re-timing, Windfury
main-hand timing, Night Elf vs Tauren Balance, Elemental Mastery hold, Enhancement imbue: unchanged, not run this pass.

Note: a scratch slip left copies of the A and B1-B3 Balance APL JSONs in `/rotations` on the audit machine (outside the repo); safe to delete.
