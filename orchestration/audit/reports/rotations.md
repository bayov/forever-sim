# Rotations and resources: pass 30 (2026-10-25T13:07:37Z)

Master 54878fa3d (last audited daba9d1a5). Result: quiet, 0 new findings.

## What changed

104 commits, mostly the regear and rotation round (class workers, 2026-10-23/24) and its follow-ups:
- 52 Level 60 APL files changed logic, covering every class except Balance and Elemental.
- 163 non-test sim `.go` files changed. These include the Forever rage rework (misses give no Rage, an 80% refund on a missed ability, crit Rage cut, damage-taken Rage), pet stat inheritance, wands (`sim/core/wand.go`), item and set rows, and class talents.
- New rulings in `standards.md`: the 8% physical hit cap, one crit elixir, explosives only where they gain (11), and the must-take talents (26).

This is a large-change pass, so I did the most important checks myself. There was no subagent fan-out.

## Checks

| Check | Result |
|---|---|
| APL validation warnings on all 64 req presets (`core.ComputeStats` rotation warnings, scratch helper) | 309 warnings. Every one is a settled inert line: a race-specific line (Berserking, Blood Fury, Shadowmeld, Shadowguard 19312, Dark Sacrifice, Chastise), a talent-less line (Adrenaline Rush, Cold Blood, Blade Flurry, Preparation, Cutthroat 462707, Missile Barrage 400589, Sweeping Strikes 12328, hunter talent shots), an explosive no longer in the consumables, or the duty-FF CoR check. No wrong spell id or rank. |
| Old (daba9d1a5) rotation vs new rotation, both on the new sim and the new req, 180 s, 16k iterations, s0=1e6 | 54 presets with a changed rotation. New wins or ties on all but 3 (table below). Large new-rotation gains: Feral +57, Subtlety +59, Ret Shock +38 to +43, Frostfire +44, Frost +18, BM +15, LW/MM +9 to +10. |
| Shared run vs published (`vs_published.txt`) | 0 of 64 moved more than 0.5 DPS. |
| Open findings to verify | None open. |

### The presets where the old rotation read positive, confirmed on both seed sets (B = old rotation, 20k iterations)

| Preset | s0 | 120 s | 180 s | 300 s |
|---|---|---|---|---|
| warrior__arms | 1e6 | −3.02 ±0.69 | +0.36 ±0.54 | +0.08 ±0.34 |
| warrior__arms | 2e8 | −3.29 ±0.34 | −0.24 ±0.46 | +0.23 ±0.27 |
| rogue__mutilate | 1e6 | +0.12 ±0.43 | +0.96 ±0.32 | +0.24 ±0.25 |
| rogue__mutilate | 2e8 | −0.40 ±0.20 | +0.56 ±0.29 | +0.19 ±0.19 |

- Arms: the new rotation (pre-pull Charge, no explosives) wins by about 3 at 120 s and ties at 180 and 300 s. Nothing to file.
- Mutilate: the old rotation pools to +0.76 at 180 s and −0.14 at 120 s. That's under the bar, so it's listed as optional. Mutilate-humanoid read −1.03 on the screen.

## Optional (under the bar)

- Rogue Mutilate: the pre-round rotation reads +0.76 at 180 s, pooled. The likely source is the new Vanish/Shadowmeld→Ambush lines. Not isolated.
- Older optional items carry over unchanged (see `passes/rotations.md`).

## Baseline

The sim and the req files both changed, so the baseline moved. New file: `ref/rotations/baseline_54878fa_3000it_seed424242.tsv`; the old d2633f4 file is deleted. It moves −114 (Demonology) to +202 (BM) against d2633f4. Those moves come from regear, talent and sim changes; they are not a regression in the rotations. `paladin__level_60-duty-crusader` and `paladin__level_60_undead` were removed by ruling (e949ac1d6, 1f9229344).

## Not done this pass (open items)

- Hand-edited A/B probes of the new rotations' individual lines: thresholds re-tuned under the rage rework (Fury HS 55, Arms HS 90 / Slam 20), the hunter melee-weave gates, the Enhancement Earth Shock without the Stormstrike hold, and the Shadow priest wand fallback. The workers measured these with rotopt in the round. The old-vs-new screen only shows that no reverted change gains.
- The open items in `passes/rotations.md` are still open: Balance Night Elf, the Elemental Mastery hold, and the warlock duty-curse screens.
