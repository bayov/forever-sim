# Results pass 93, master 5f20dfdf4 (2026-11-17)

Shared run (68 builds, 20k, seed 1) vs builds.json: 20 of 68 off by more than 1.5, all explained by sim commits since 1a85df1d7.

| Build family | builds.json | run | Commit (paired 20k seed 1) |
|---|---|---|---|
| Beast Mastery | 1182.2 | 1114.8 | b075ae64d pet white swing no hunter crit: -66.9 |
| Marksmanship | 1234.3 | 1201.0 | b075ae64d: -33.0 |
| Survival | 1034.6 | 1012.7 | b075ae64d -37.4, then +15 from 4a8541e46, eada6763e, 29481dd0e |
| Lone Wolf | 1291.3 | 1291.3 | none (no pet) |
| Arms, Fury families | 1069 to 1103 | +8 to +10 | 2ebdc14d6 +3, ad1d7f469 (hotfix 114097) +6 to +8 |
| Subtlety | 866.8 | 872.0 | cf95c3e32 Death Mark +5.6 |

Drift: only the known feral sapper line (results-20261105-1). Unread fields: none besides notes.
The hunter pet crit change cites wowsims/forever 7e0e0c2e46 (EV-0134, 154k landed swings). It is large (BM -5.7%) and awaits a re-sim. Hunter and warrior builds need republishing.
