# Rotations pass 77 (2026-11-25, phase0-step4-publish 226ada86b)

Findings: 1 found and fixed in the pass (rotations-20261125-1). Range 63129ce35..226ada86b: 78802e6e put real
on-use trinket lines in the APLs and requests; 1fed43cbf (Auto Time To Next); rotation_text.json wording (ff502e808).

## Checked

| Check | Result |
|---|---|
| Each new trinket line against the same request without it (autocast on cooldown), 20k, seed 1 | Two lines lose DPS: warlock Talisman of Ascendance pre-pull at -0.1s (-1.9 to -4.9 on 5 builds) and Fury DW Blackhand's Breadth held for the last 30 s (-2.0 to -3.4 on 3 builds). Arms and Enhancement lines gain or tie. |
| Fix 67fe28cf0 | Talisman line removed from ui/warlock APLs, warlock requests and race requests; Blackhand's line on cooldown in the 3 Fury DW requests and APLs. Measured: Affliction 960.87 -> 965.79, Destruction 877.59 -> 881.25, Demonology 892.02 -> 894.19, Fire 833.57 -> 835.28, Shadow Bolt 808.17 -> 810.11, Fury DW 1005.64 -> 1008.99, Sunder 971.23 -> 974.55, Battle Shout 1000.87 -> 1002.87. |
| Rotation card text | 4 Warrior and 1 Feral code-like conditions worded; build_data.log has no "read like code" warning. |
| Tank racial lines | Pinned by sim/warrior/tank_warrior/racials_apl_test.go (4c26b4b66). |

## Open findings

None new.
