# Results audit, pass 76 (master fb584eafa, 2026-11-04)

Trigger: warrior Fury follow-up (7b4dc5d97), report hashes, presets.ts cleanups.

- Shared run (68 builds, 20k, seed 1) vs metas: 62 exact within 0.15; only the six shaman metas lag (+199 to +277, open finding results-20261102-1).
- Warrior: all six Fury metas match the run (fury 1152.97, fury-bs 1150.52, fury-sunder 1117.66, fury_2h 1173.28, 2h-bs 1168.75, 2h-sunder 1134.61). builds.json is still pre-follow-up for fury_2h and its two duties (-3.4, -3.4, -2.0), expected until the republish.
- Drift (regen.py): no warrior drift; no new differences vs the reference (the diffs are earlier fixes: feral duty-ff explosive, rogue subtlety race, affliction_shadow_bolt rename).
- presets.ts changes are one-line removals with no request change; no number moved.
- Result: quiet, no new findings.
