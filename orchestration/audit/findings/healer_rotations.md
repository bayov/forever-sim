# Findings: healer_rotations

## Open

### healer_rotations-20261111-1: Restoration druid raid healer loses its emergency Nature's Swiftness Regrowth to Wild Growth since NS also makes Wild Growth instant (-0.046 mean carry, Magmadar -0.184)
- severity: high (Magmadar 2.002 -> 1.818 carry, about 9%; every boss file down)
- affects: Restoration druid raid healer (`ui/restoration_druid/apls/raid_healer.apl.json`, `tools/healbench/req/druid_restoration_raid.json`). Not the tank healer (no Wild Growth line, rows unchanged).
- found: 2026-11-11, master 707ae92cc
- evidence: 9628e85b6 adds Wild Growth to Nature's Swiftness's affected spells (`sim/druid/healing.go` ~752). The APL leaves NS to `autocastOtherCooldowns` (line 1), which raises the buff at 0 s; before, the buff stayed up until line 4's emergency Regrowth (lowest under 60%) used it. Now the first Wild Growth (line 5, an ordinary cast) uses it. One logged Magmadar iteration: old rule, an instant Regrowth (9858, 0 s); master, an instant Wild Growth (1238215, 0 s) and Regrowth cast at 1.96 s. Raid role, own lengths, 1,000 it, default seed, ref -> master: Onyxia 0.997 -> 0.943, Broodlord 0.920 -> 0.906, Chromaggus 1.109 -> 1.092, Golemagg 2.660 -> 2.658, Magmadar 2.002 -> 1.818, Ragnaros 1.270 -> 1.240, Vael 1.102 -> 1.079. Bisected with benches built at master with one file reverted: reverting `sim/druid/healing.go` to 9628e85b6^ gives back the ref rows exactly; reverting `sim/core/dot.go` (f20d5938a) changes nothing.
- tried (paired, 1,000 it, master): (A) insert `castSpell 17116` with line 4's condition (lowest under 60%) just before line 4: 0.998 / 0.919 / 1.108 / 2.655 / 2.009 / 1.274 / 1.108, mean 1.439 against 1.391 (+0.048, same as before the hotfix). At 180 s: Magmadar 1.585 -> 1.660, Ragnaros 1.172 -> 1.191, Onyxia 0.823 -> 0.865. At 300 s: 1.467 -> 1.578, 1.059 -> 1.137, 0.762 -> 0.785, with no less mana left. It also runs unchanged on the old rule (same numbers to 0.003). (B) Wild Growth gated on NS not active: mean 1.406, Vael -0.095, worse than A. (C) NS line before Healing Touch 11 (line 9): mean 1.391, no gain (Wild Growth at line 5 still comes first).
- proposed fix: variant A. Add `{"action":{"castSpell":{"spellId":{"spellId":17116}},"condition":<line 4's condition>},"notes":"Nature's Swiftness for the emergency Regrowth below."}` as the new line 4 of the raid healer APL and its request, and correct line 1's note. Aside: the comment above `registerNaturesSwiftnessCD` (`sim/druid/healing.go` ~740) still says "Wild Growth isn't one".
- needs: worker fix

### healer_rotations-20261104-3: Restoration shaman tank healer's Lesser Healing Wave 5 line never fires (same mana gate as rank 6 just above it)
- severity: low (no carry to gain; a dead line in a shipped preset)
- affects: Restoration shaman tank healer (`ui/restoration_shaman/apls/tank_healer.apl.json` line 9, `tools/healbench/req/shaman_restoration.json`)
- found: 2026-11-04, master 0facd9428
- evidence: since the re-tune 324d64834, line 8 casts Lesser Healing Wave 6 (10468) on the tank whenever mana >= 50%, and line 9 casts rank 5 (10467) on exactly the same condition. Rank 6 is always affordable at 50% mana, so rank 5 can never be reached. Logged iterations (`healbench -role tank -encounter tools/encounters/{golemagg,chromaggus}.json -dur 300 -scale 1.5 -log`) show 10467 cast 0 times, against 10468 at 45 / 18 casts. Sure: the conditions are identical.
- proposed fix: drop line 9, or give rank 5 its own condition (for example a smaller tank deficit, or the 50-70% mana band) and measure it. No carry change is expected from dropping it.
- needs: worker fix

### healer_rotations-20261104-2: Discipline tank healer runs out of mana on Chromaggus at the fight's own length (107 of 115 s), and on Golemagg at 300 s
- severity: medium (a mana failure at a judged length; the preset already trails the Holy tank healer there, 0.818 vs 1.002)
- affects: Discipline priest tank healer (`ui/healing_priest/apls/discipline_tank_healer.apl.json`, `tools/healbench/req/priest_discipline.json`)
- found: 2026-11-04, master 1e9eb71f7
- evidence: `healbench -role tank -encounter tools/encounters/chromaggus.json -carry -iters 4000 tools/healbench/req/priest_discipline.json`: carry 0.825, 431 mana left, OOM at 107 s (1,000 it: 0.818, 435, 106-107 s, on a1b1163d2 and on 1e9eb71f7). The only preset of nine to run dry at a boss length; the next lowest are Holy priest and shaman tank on Chromaggus with about 3,000 left. At 300 s on Golemagg it runs dry at 227 s (carry 0.891 against Holy's 1.201). One logged iteration on Golemagg at 300 s: it spends 38.7k mana (Flash Heal 24.6k, Renew 6.3k, Power Word: Shield 5.0k), the Holy tank healer 34.7k on more healing.
- tried (paired seeds): Power Word: Shield mana gate 0.9 (was 0.5): Chromaggus +0.021 at 4,000 it (0.825 -> 0.846), still OOM at 107 s; mean of the 7 files +0.008 (Vael +0.022, Golemagg -0.009); 300 s Golemagg/Ragnaros +0.005/+0.019. Gate 0.7: Chromaggus +0.008 at 4,000 it, mean +0.005. Flash Heal first gate 0.9 or no Flash Heal first: mean -0.008 / -0.012, OOM unchanged. Renew gated on 60% missing: mean -0.065.
- proposed fix: none of these stops the OOM, so this needs a worker to find a cheaper tank pattern for Disc on long, heavy tank fights (for example a lower Flash Heal rank under a mana line, or fewer Renews below 50% mana). The shield gate 0.9 is the best one-line change measured (+0.021 Chromaggus, +0.008 mean) and could go in meanwhile.
- update 2026-11-04, master 0facd9428: a7d9b57ef (Flash Heal 7 only, mana gate 0.85, shield gate 0.7) meant to fix this, but on master it still runs dry. Chromaggus at its own length: carry 0.866, 365 mana left, OOM at 101 s (1,000 it). Golemagg at 300 s: OOM at 210 s (500 it). The old APL from 1e9eb71f7 on today's bench gives 0.854, also OOM at 99 s, so the re-tune didn't move it. The bench's tank changes since then (the tank faces the boss; idle fillers in the tank role) raised the carry scale, so more damage has to be healed. The Holy tank healer now ends Chromaggus with only 677 mana left (it had 3,058) and Golemagg 300 s with 1,043. The variants (shield gate 0.9, emergency Flash Heal at 35%) were cut off by the wrap-up and are not measured.
- update 2026-11-05, master a07941531: unchanged on master (Chromaggus 0.866, 365 mana, OOM 101 s, 1,000 it).
  Paired 1,000 it: shield gate 0.9 (line 3, was 0.7) Chromaggus 0.889 / 428 mana / OOM 107 s, Golemagg 1.984
  (-0.007); Renew only at 40% mana or more: Chromaggus 0.865, OOM 103 s, Golemagg unchanged. Neither stops the OOM.
- update 2026-11-08, master 9ba00c2dc: unchanged (Chromaggus 0.866, 364 mana, OOM 101 s, 1,000 it).
- update 2026-11-10, master ca5e35c3f: unchanged (same row, byte-identical).
- update 2026-11-11, master 707ae92cc: unchanged (same row, byte-identical).
- update 2026-11-12, master cb0d928fd: unchanged (same row, byte-identical).
- needs: worker fix

## Resolved

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
- resolved: 2026-11-04, dda14db65 (Power Infusion dropped from both Holy APLs; the raid healer is down to 6 lines; doc text updated)

