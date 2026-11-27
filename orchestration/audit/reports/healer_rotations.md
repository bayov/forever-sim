# healer_rotations: branch pass, phase0-cleanup 782ad6089 (2026-11-27)

Scoped by hand to the priest tank healers (ff87e66b: Flash Heal first lines also need mana % >= the fight's remaining %);
state not advanced. Full baseline re-run at HEAD (`baseline.sh`, 1,000 it) and written to ref/healer_bench/summary.tsv.
Druid, paladin and shaman rows moved too, from master's Phase 0 gear commits (eef393de, d74a5326), which aren't this
branch's. The master pass judges those.

## Priest rows vs the previous ref (carry)
| request | ony | brood | chrom | golem | magma | rag | vael |
|---|---|---|---|---|---|---|---|
| priest_holy | 1.316 = | 0.881 = | 0.878 -> 0.912 | 1.861 -> 1.867 | 2.011 = | 0.567 -> 0.569 | 0.841 = |
| priest_discipline | 1.354 = | 0.926 -> 0.941 | 0.866 -> 0.877 | 1.938 -> 1.950 | 2.015 -> 2.027 | 0.601 -> 0.595 | 0.893 = |
| priest_holy_raid | unchanged on all 7 | | | | | | |

## Mana side, master vs HEAD request (1,000 it, tank role)
| | Ony 180 | Ony 300 | Chrom 180 | Chrom 300 |
|---|---|---|---|---|
| Holy | 0.892 -> 0.888 | 0.711 -> 0.720 | 0.704 -> 0.706 | 0.544 -> 0.544 |
| Disc | 0.855 -> 0.855 | 0.696 -> 0.694 | 0.667 -> 0.664 | 0.541 -> 0.540 |

- The new condition paces mana to the fight's own clock, which standards 31 allows. No length loses by more than noise.
- healer_rotations-20261125-1 (Holy, out of mana on Chromaggus) improved but is still open: out of mana at 110 s of 115.
  An update line was added.
- No new findings.
