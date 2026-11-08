# Rotations pass 56 (1 finding)

- Master: 9ba00c2dc (from 73d5bfa9e, 4 commits in the watched paths), audited 2026-11-08. Client 1.60.1.70862
  per the gate (state had 70753; no req or APL change came with it in this range).
- Not quiet: non-test sim `.go` changed that the presets reach:
  - 2a60326a1: spell damage and healing roll in whole points (`sim/core/sim.go`, `cast.go`, `item_effects.go`).
  - 0193721bc: Forever's downranking penalty hits heals only (user ruling 2026-11-08); damage ranks keep the full
    coefficient (`sim/core/downrank.go`, shaman shocks and Lightning Bolt).
  - 5a870eccc: Stormstrike mark applied after the damage calc (log raw only; damage unchanged).
  - 9ba00c2dc: Windfury and Flametongue Totem replace each other (`sim/core/totem_weapon_buffs.go`); the four
    Enhancement reqs and presets drop the raid's Flametongue Totem.
- `apl_logic_diff.py 73d5bfa9e 9ba00c2dc`: 0 APL files touched.
- Regression baseline refreshed: `ref/rotations/baseline_9ba00c2_3000it_seed424242.tsv`.

## What moved (baseline, 3k it, seed 424242, vs a079415)

| Builds | Move | Cause |
|---|---|---|
| Enhancement mana-tide / imp-gw / level60 / water-shield | -22.1 / -20.2 / -19.8 / -19.6 | 9ba00c2dc: no raid Flametongue Totem (req change). |

Every other row moved under 0.5 DPS: the whole-point rolls and heals-only downranking don't move any DPS
preset measurably (no Level 60 DPS rotation casts a damage rank learned below 43).

## Screens (`ab_req.py`, 20k it, 8 blocks, shared CLI 9ba00c2dc473)

Open item from the pass doc: Flametongue Weapon (imbue) with our own Windfury Totem, in place of Windfury
Weapon, on every Enhancement build. Only `shamanImbue` changes; rotations as on master. Pooled B-A over
S0 = 1e6 and 2e8 at 120 / 180 / 300 s:

| Preset | 120 | 180 | 300 | Verdict |
|---|---|---|---|---|
| mana-tide | +5.87 | +5.61 | +3.11 | finding rotations-20261108-1 |
| water-shield | +3.79 | +3.43 | +1.50 | finding rotations-20261108-1 |
| level60 | +0.68 | -1.06 | -3.93 | keep Windfury Weapon |
| imp-gw | +2.19 | +0.59 | -2.92 | fight-length trade-off; keep |

Per-set numbers are in the finding. The two builds that gain are the ones with less melee weight (Mana Tide
gives up Windfury-boosting talents' share, Water Shield runs on mana); plausibly Flametongue Weapon's flat
spell hits plus Windfury Totem's procs beat Windfury Weapon there. Not traced further.

New helper: `tools/audit/rotations/ab_req.py`, a paired A/B of two whole request files (for imbue, consume or
option screens, which `ab.py` can't vary).

## Open findings

- rotations-20261105-1 (Feral Ferocious Bite gate 45 -> 65): still open; master's level60.apl.json:10 has 45.
- rotations-20261108-1: new, above.

## Optional

None this pass.

## Left for later

- Re-search the Enhancement rotations under the new totem rules (and under Flametongue Weapon for Mana Tide and
  Water Shield); this pass doesn't use rotopt.
- Elemental Mastery hold, Balance Night Elf, warlock duty-curse variants: unchanged from the pass doc.
