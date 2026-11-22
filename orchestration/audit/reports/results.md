# Results pass 97 (master 8b5137c6d6, client 1.60.1.71037)

Method: shared run (68 builds, 20k, seed 1, master binary) against builds.json, the metas (vs_published.txt) and the drift check.

| Check | Result |
|---|---|
| builds.json vs run | 68 of 68 within 1.5 (none off) |
| metas vs run | 49 of 68 off by more than 0.5 (known, results-20261104-2) |
| drift (regen.py vs reference) | only the known feral sapper line (results-20261105-1) |

Largest meta gaps: Affliction +38 to +43, Lone Wolf -22.4, rogue swords/daggers -6 to -15, Shock Ret +6 to +7, Destruction +6.
No new findings.
