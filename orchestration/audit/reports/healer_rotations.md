# healer_rotations: pass 32, master 5ab589d85 (2026-11-27)

Since 34994816e: ff87e66b3 (priest tank healers pace their Flash Heal first lines), eef393de0 (Restoration druid
Phase 0 gear in both healbench requests), d74a53261 (Holy paladin and Restoration shaman requests drop the Armaments
of Battle rewards), 413f06cb0 (Talisman of Ascendance charges; no healer request wears it). The request diffs for
druid, paladin and shaman are gear only, with no rotation lines. Encounters, healbench main.go and the Frost mage filler
are unchanged.

## Baseline
`tools/audit/healer_rotations/baseline.sh`, 1,000 it, default seed: all 63 rows are byte-identical to
ref/healer_bench/summary.tsv. The phase0-cleanup branch pass (5ab589d85) had already refreshed that file at this
code, so the ref is unchanged. That branch pass left the non-priest moves for this pass to judge. Against the ref
before it (849024cbb), the moves over 0.02 carry are:

| request | boss | carry | mana left | tank / other deaths |
|---|---|---|---|---|
| druid_restoration | onyxia | 1.387 -> 1.337 | 5749 -> 5111 | 0 / 0 |
| druid_restoration | chromaggus | 1.011 -> 0.904 | 1882 -> 827 | 1.0% -> 18.4% / 1.0% -> 14.1% |
| druid_restoration | golemagg | 1.902 -> 1.789 | 5824 -> 5137 | 0 / 0 |
| druid_restoration | magmadar | 2.049 -> 1.940 | 5982 -> 5237 | 0 / 0 |
| druid_restoration | ragnaros | 0.623 -> 0.602 | 4740 -> 4616 | 84% -> 95% |
| druid_restoration | vaelastrasz | 0.931 -> 0.879 | 5438 -> 5117 | 6.2% -> 15.0% |
| druid_restoration_raid | golemagg | 2.494 -> 2.434 | | |
| druid_restoration_raid | magmadar | 1.824 -> 1.734 | | |
| druid_restoration_raid | vaelastrasz | 0.963 -> 1.029 | | other 5.3% -> 2.8% |
| paladin_holy | onyxia / golemagg | +0.026 / +0.028 | | |

All of these come from the deliberate Phase 0 gear change. The druid's tank carry is now in line with the other tank
healers on the Phase 0 set (Chromaggus: druid 0.904, shaman 0.927, Holy priest 0.912, paladin 0.827). The druid
ends Chromaggus with 827 mana left, and its projected run-out time is 126 s against a 115 s fight. The shaman moved
within noise. No finding: this is the intended gear scale, and the gear list belongs to healer_gear.

The priest rows (ff87e66b3) were judged by the branch pass at 180 and 300 s; none lost by more than noise.

## Open findings
- -20261125-1 (Holy tank healer out of mana on Chromaggus): unchanged; carry 0.912, 442 mana left, runs out at 110 s, other deaths 14.3%.
- -20261104-2 (Discipline tank healer, Chromaggus): carry 0.877, 376 mana left, runs out at 104 s; still open.
- -20261124-1 (filler not pinned), -20261118-1, -20261115-1: no new evidence (the raid rows are unchanged since the ref).

No new findings. No one-line rotation trials this pass, because no rotation line changed outside the priest lines the branch pass already measured.
