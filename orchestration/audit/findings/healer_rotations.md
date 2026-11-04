# Findings: healer_rotations

## Open

### healer_rotations-20261104-1: Holy priest presets carry lines that never fire (Power Infusion without the talent; the raid healer's Prayer of Healing, Renew, Greater Heal and Inner Focus behind Flash Heal 4)
- severity: low (no carry to gain: every reorder tried loses)
- affects: Holy priest tank healer and raid healer (`ui/healing_priest/apls/holy_{tank,raid}_healer.apl.json`, `tools/healbench/req/priest_holy{,_raid}.json`); the rotation text in `docs/healing_priest.md` (Presets, "Every rotation casts ... Power Infusion on the priest when a big heal is due", and the Holy raid bullet)
- found: 2026-11-04, master a1b1163d2
- evidence:
  - Power Infusion (10060, line 2 of both Holy APLs) needs the Discipline tree's 31-point talent; the Holy build `00500503113-03505103030121531` has 18 Discipline points. In logged iterations (Golemagg ×1.6 180 s, Magmadar ×2.0) the Holy tank healer never casts 10060; the Discipline preset casts it 1-2 times. Dead line, harmless.
  - Holy raid healer: Flash Heal 7/4 (lines 5-6) take any hole that Prayer of Healing 5, Renew (×0.6), Greater Heal 5/2 and Inner Focus (lines 7-11) would take, so those never fire. Logged iterations at 300 s near the carry scale (`healbench -role raid -encounter <boss> -scale S -dur 300 -log`): Onyxia ×0.7, Magmadar ×1.7, Ragnaros ×1.25, Chromaggus ×0.9 and Golemagg ×2.0 (180 s): 0 casts of 25316, 25315, 25314, 10963 in all five; Inner Focus 1 cast in one.
  - Moving them up loses (1,000 it, paired seeds, each boss at its own length, -role raid, mean of 7 files): Prayer of Healing ahead of Flash Heal -0.094 (Golemagg 2.064 -> 1.762, Magmadar 1.693 -> 1.438); Renew ahead -0.182; Inner Focus ahead of Flash Heal (when Flash Heal 7 fits) -0.003 (noise). So the order is right and only the text is wrong.
- proposed fix: drop Power Infusion from the two Holy APLs (or keep it and say it is for a Discipline build), drop or annotate the raid healer's unreachable lines 7-11, and correct `docs/healing_priest.md`'s rotation text (via `orchestration/tools/healers/priest_tuning/rots.py`, which writes them). No carry change expected.
- needs: worker fix

## Resolved
