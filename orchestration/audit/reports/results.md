# Results pass 95, master a6677362f (2026-11-18)

Shared run (68 builds, 20k, seed 1, CLI a6677362f) vs builds.json: 68 of 68 within 1.5. builds.json was republished by 0db5b9397
(re-sim for haste shortening a spell's GCD, floored at 1 s, user ruling 2026-11-18) and 392503f55 (full regeneration; DPS builds unchanged).
The other watched changes (report switches folding the local patches, parallel/speedup, healer reaction time, labeled rolls) move no DPS build.

Metas vs run: 45 of 68 off by more than 0.5 (was 24). The extra ones are small moves from the GCD change that re-sim updated in builds.json only.

| Family | n | run minus meta |
|---|---|---|
| Warlock | 13 | Affliction +38 to +43, rest +0.8 to +6.6 |
| Rogue | 5 | -1.2 to -20.1 |
| Hunter | 4 | Lone Wolf -22.4, rest <= 0.7 |
| Paladin | 8 | -0.8 to -4.0 |
| Druid | 4 | +3.9 to +5.0 |
| Shaman / Mage / Warrior | 4 / 2 / 5 | <= 4.0 / -4.9 / <= 0.8 |

Drift (regen.py, fresh mksettings): only the known feral sapper line (results-20261105-1); the reference drift_report predates it.
Unread fields: none. No new findings; the metas lag is recorded under results-20261104-2.
