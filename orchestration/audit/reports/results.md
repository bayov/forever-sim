# Results pass 94, master 2f7e255af (2026-11-18)

Shared run (68 builds, 20k, seed 1) vs builds.json: 68 of 68 within 1.5, so the republish caught up with the pass 93 moves
(Beast Mastery 1115.4, Marksmanship 1201.3, Survival 1012.1, Lone Wolf 1291.2, Subtlety 872.2, Affliction 1070.8).

Changes since 5f20dfdf4 (Nov 16 hotfixes, Forever update 2026-11-17: Shieldsworn Holy Strike no longer rolls Windfury Totem,
Vampiric Embrace heal, BM weaves for every main hand swing, druid Regrowth/Idol, item and db changes): none moves a published build
beyond the metas' known lag; hunter, warrior and Subtlety metas now match the run.

Metas vs run: 24 of 68 off by more than 0.5 (was 41 earlier), all in known families.

| Family | meta | run | note |
|---|---|---|---|
| Affliction (3) | 1034.1 / 1025.1 / 1019.9 | +36.7 / +35.2 / +31.9 | hotfix 113766, meta not re-simmed |
| Lone Wolf | 1313.7 | -22.4 | meta from round 70753; builds.json is 1291.2 |
| Rogue combat/seal fate (4) | | -10.5 to -20.1 | metas pre-date the regears; builds.json matches the run |
| Mage Fire | 1057.9 | -4.9 | noise-level |
| The rest (<= 1.2) | | | noise |

Drift: only the known feral sapper line (results-20261105-1; the reference drift_report predates it, replace it when that is fixed).
Unread fields: none (not even notes). No new findings.
