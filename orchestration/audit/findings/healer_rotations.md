# Findings: healer_rotations

## Open

### healer_rotations-20261104-2: Discipline tank healer runs out of mana on Chromaggus at the fight's own length (107 of 115 s), and on Golemagg at 300 s
- severity: medium (a mana failure at a judged length; the preset already trails the Holy tank healer there, 0.818 vs 1.002)
- affects: Discipline priest tank healer (`ui/healing_priest/apls/discipline_tank_healer.apl.json`, `tools/healbench/req/priest_discipline.json`)
- found: 2026-11-04, master 1e9eb71f7
- evidence: `healbench -role tank -encounter tools/encounters/chromaggus.json -carry -iters 4000 tools/healbench/req/priest_discipline.json`: carry 0.825, 431 mana left, OOM at 107 s (1,000 it: 0.818, 435, 106-107 s, on a1b1163d2 and on 1e9eb71f7). The only preset of nine to run dry at a boss length; the next lowest are Holy priest and shaman tank on Chromaggus with about 3,000 left. At 300 s on Golemagg it runs dry at 227 s (carry 0.891 against Holy's 1.201). One logged iteration on Golemagg at 300 s: it spends 38.7k mana (Flash Heal 24.6k, Renew 6.3k, Power Word: Shield 5.0k), the Holy tank healer 34.7k on more healing.
- tried (paired seeds): Power Word: Shield mana gate 0.9 (was 0.5): Chromaggus +0.021 at 4,000 it (0.825 -> 0.846), still OOM at 107 s; mean of the 7 files +0.008 (Vael +0.022, Golemagg -0.009); 300 s Golemagg/Ragnaros +0.005/+0.019. Gate 0.7: Chromaggus +0.008 at 4,000 it, mean +0.005. Flash Heal first gate 0.9 or no Flash Heal first: mean -0.008 / -0.012, OOM unchanged. Renew gated on 60% missing: mean -0.065.
- proposed fix: none of these stops the OOM, so this needs a worker to find a cheaper tank pattern for Disc on long, heavy tank fights (for example a lower Flash Heal rank under a mana line, or fewer Renews below 50% mana). The shield gate 0.9 is the best one-line change measured (+0.021 Chromaggus, +0.008 mean) and could go in meanwhile.
- needs: worker fix

### healer_rotations-20261104-1: Holy priest presets carry lines that never fire (Power Infusion without the talent; the raid healer's Prayer of Healing, Renew, Greater Heal and Inner Focus behind Flash Heal 4)
- severity: low (no carry to gain: every reorder tried loses)
- affects: Holy priest tank healer and raid healer (`ui/healing_priest/apls/holy_{tank,raid}_healer.apl.json`, `tools/healbench/req/priest_holy{,_raid}.json`); the rotation text in `docs/healing_priest.md` (Presets, "Every rotation casts ... Power Infusion on the priest when a big heal is due", and the Holy raid bullet)
- found: 2026-11-04, master a1b1163d2
- evidence:
  - Power Infusion (10060, line 2 of both Holy APLs) needs the Discipline tree's 31-point talent; the Holy build `00500503113-03505103030121531` has 18 Discipline points. In logged iterations (Golemagg ×1.6 180 s, Magmadar ×2.0) the Holy tank healer never casts 10060; the Discipline preset casts it 1-2 times. Dead line, harmless.
  - Holy raid healer: Flash Heal 7/4 (lines 5-6) take any hole that Prayer of Healing 5, Renew (×0.6), Greater Heal 5/2 and Inner Focus (lines 7-11) would take, so those never fire. Logged iterations at 300 s near the carry scale (`healbench -role raid -encounter <boss> -scale S -dur 300 -log`): Onyxia ×0.7, Magmadar ×1.7, Ragnaros ×1.25, Chromaggus ×0.9 and Golemagg ×2.0 (180 s): 0 casts of 25316, 25315, 25314, 10963 in all five; Inner Focus 1 cast in one.
  - Moving them up loses (1,000 it, paired seeds, each boss at its own length, -role raid, mean of 7 files, master 1e9eb71f7 with the raid damage fix): Prayer of Healing ahead of Flash Heal -0.094 (Magmadar 2.007 -> 1.661, Golemagg 2.542 -> 2.320); Renew ahead -0.202; Inner Focus ahead of Flash Heal (when Flash Heal 7 fits) -0.008. On a1b1163d2 the same: -0.094, -0.182, -0.003. Logs on 1e9eb71f7 (Onyxia x0.95, Magmadar x2.0, Golemagg x2.5, Ragnaros x1.3, 300 s): still 0 casts of 25316, 25315, 25314, 10963. So the order is right and only the text is wrong.
- proposed fix: drop Power Infusion from the two Holy APLs (or keep it and say it is for a Discipline build), drop or annotate the raid healer's unreachable lines 7-11, and correct `docs/healing_priest.md`'s rotation text (via `orchestration/tools/healers/priest_tuning/rots.py`, which writes them). No carry change expected.
- needs: worker fix

## Resolved
