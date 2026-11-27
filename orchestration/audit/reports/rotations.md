# Rotations pass 85 (2026-11-27, master 34994816e)

Result: quiet. Range f9fa79917..34994816e: Phase 0 lists drop the Armaments of Battle rewards (6bd2d60a2),
Chained Essence of Eranikus cloud ticks crit at a fixed 2.7% (5591bd019), Grayston's 2H Shockadin presets
added and dropped again (27a8031dc..d70972fa3), Enhancement trinket rerank, logcheck replay windows.

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py f9fa79917 34994816e7` | 10 warlock APLs, one line each: the Bane of Doom hold reads a new trinket id, 22678 (Talisman of Ascendance, now out) -> 13968 (Eye of the Beast) in the 6 Destruction / Destruction Fire files, -> 10455 (Chained Essence) in the 3 Demonology files and Affliction Shadow Bolt |
| APL trinket ids vs Phase 0 gear sets | match: Destruction wears 13968 + 272438, Demonology 13968 + 10455 (APL reads 10455, per the commit the larger use value), Affliction 272438 + 10455 |
| Req rotations changed | none; only `paladin__level_60_shock_ret.meta.json`. The warlock reqs still wear 22678 and embed the old 22678 hold, so they stay self-consistent until the report regenerates them from the new lists |
| Non-test sim `.go` changed | `sim/common/item_effects.go` only: Chained Essence tick outcome (crit 2.7% fixed, x1.5). Trinket damage, no rotation or resource mechanic. 40 reqs wear 10455 |
| Shared run vs published (`runs/vs_published.txt`) | 31 of 68 moved > 0.5 DPS, all within +-0.15%; the losses (-0.6 to -1.5) are 10455 wearers, as expected from the lower tick crit. No regress run (under its 1.0 bar for most, no rotation cause) |
| Open findings landed? | 20261126-3 (Might of Shahram stun), 20261122-1 (Shock Ret order), 20261112-1 (rogue poisons) still open, untouched in range |
| Baseline | not refreshed (moves ~1 DPS on 10455 wearers only): `baseline_ab3fff7_3000it_seed424242.tsv`. Refresh on the next pass that runs `baseline.py` |

## Phase 0 trinkets open item (warlock part)

Done for warlock: the trinket lines name items the builds wear. Note: the Bane of Doom hold (skip Doom while the
trinket is 59-60 s from ready) was built for a 2 min trinket. Eye of the Beast is a 5 min cooldown and Chained
Essence 10 min (sim/common/item_effects.go:4269, :3751), so in a fight of 300 s or less the clause never fires
(Eye comes back at ~300 s, when under 60 s remain). Harmless; a cleanup, not a finding. Other classes still to do.

## Optional

None new.
