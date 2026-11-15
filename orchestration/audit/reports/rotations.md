# Rotations pass 66 (2026-11-15, master 70ea588bb)

Last audited master: d630df2cb. Client 1.60.1.70888 data (unchanged). Result: findings (1 new, low).

## What changed (d630df2cb..70ea588bb)

| path | change | reaches presets? |
|---|---|---|
| `sim/druid/ferocious_bite.go` (0cc2c3c8b) | an avoided Ferocious Bite refunds 80% of its 35 Energy (Forever) | yes: both Feral presets |
| `sim/druid/forms.go` (0cc2c3c8b) | Sharpened Claws adds spell crit in feral forms | Feral only; no Feral spell crit use seen in the baseline |
| report data, known-differences notes | output only | no |

`apl_logic_diff.py d630df2cb 70ea588bb`: 0 APL files touched. Baseline: druid rows re-simmed (3000 it, seed 424242):
feral-cat 1152.74 -> 1155.09, feral-cat-duty-ff 1147.41 -> 1149.77, Balance unchanged. New file
`ref/rotations/baseline_70ea588_3000it_seed424242.tsv` (other rows carried; no other sim change since 3a86d4a).

## Feral Ferocious Bite gate under the refund (finding rotations-20261105-1)

Still 45 on master; 65 still gains on both builds and both seed sets (pooled +2.49 / +1.67 at 180 s, +4.35 / +3.64 at 300 s).
Gate 75 (set 1): -0.05 / +0.44 / +3.28, worse. Update added to the finding.

## Windfury Totem timing (user request 2026-11-15, open item in passes/rotations.md)

What can be timed: only abilities that roll the totem under Forever (`windfury_totem_procs_forever.go`, standards 17). Of the
listed ones, Sinister Strike, Backstab, Mortal Strike, Heroic Strike and Raptor Strike roll it; Stormstrike, Holy Strike,
Crusader Strike, Judgement of Command, Bloodthirst, Claw and Shred never do, so their timing can't matter to the totem.
Heroic Strike, Raptor Strike, Cleave and Maul replace a swing (nothing to time). Also gated: Slam, Rend, Execute, Hamstring,
Eviscerate, Hemorrhage, Mongoose Bite, Wing Clip, Spearing Strike, as they appear.

Tool: `tools/audit/rotations/wf_gate.py REQ OUT [--mode proc|swing] [--ms 100]` (new). `proc` holds those abilities while
the totem's ICD runs after a proc; `swing` holds them 100 ms after every main-hand swing ("spread between swings").
A first version gated on the buff alone and was nearly inert (the 1 s buff drops once its charges are used); the
results below use the fixed gate. Paired `ab.py`, 20k it, CLI 70ea588bb116, B-A at 180 s, set 1 (S0 1e6):

| preset | proc gate | swing gate | A at 180 |
|---|---|---|---|
| warrior__arms | **+2.42 ± 0.25** (both sets, pooled +2.50) | +2.27 ± 0.55 | 1042.9 |
| warrior__arms_dragonkin | +1.60 ± 0.14 | - | 1100.8 |
| warrior__fury_2h | +0.30 (pooled +0.36) | - | 1033.0 |
| warrior__fury (DW) | -0.01 ± 0.18 | -0.58 ± 0.47 | 1052.1 |
| rogue__combat-swords | +0.62 ± 0.16 | -0.25 ± 0.25 | 943.7 |
| rogue__seal-fate-daggers | +0.18 ± 0.14 | +0.27 ± 0.27 | 904.3 |
| hunter__survival | +0.08 ± 0.10 | +0.26 ± 0.31 | 1014.9 |

The other rogue builds (combat-daggers, seal-fate-swords, mutilate, subtlety) were run only with the inert first gate
(0.00); by the Combat Swords and Seal Fate Daggers results they'd sit under the bar. Ret paladin and Feral have nothing
that rolls the totem. Arms is the only build past the bar: finding rotations-20261115-1.

Not done: **Enhancement, hold Stormstrike until Windfury Weapon's cooldown is back.** The imbue's 1.5 s ICD is a bare
`core.Cooldown` in `sim/shaman/windfury_weapon.go:95`, not on an aura or spell, so no APL value can read it; testing it
needs a sim change (e.g. attach the ICD to the "Windfury Imbue" aura so `auraIcdIsReadyWithReactionTime` sees it).
Enhancement presets carry no Windfury Totem, so the totem gates don't apply to them. Apart from this, the open item in
`passes/rotations.md` is done and can be removed (this pass may not edit that doc).

## Optional (under the bar)

- Combat Swords, Windfury proc gate on Sinister Strike/Eviscerate: +0.62 at 180 s (set 1 only).
- Fury 2H, same gate: +0.36 pooled at 180 s.

## Open findings

- rotations-20261115-1 (new, Arms Windfury gate), rotations-20261112-1 (rogue poison swap), rotations-20261109-1
  (Fury 2H 90-rage Heroic Strike), rotations-20261105-1 (Feral Bite gate 65): all unapplied on 70ea588bb.
