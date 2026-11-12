# Findings: healer_rotations

## Open

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
- update 2026-11-12, master 3065ba744: still dry on Chromaggus (carry 0.871, 341 mana left, OOM at 101 s, 1,000 it) with the new consumables.
- update 2026-11-12 (worker, on tankheal): re-screened on the new consumables, tank role, own lengths, 1,000 it,
  mean carry, seeds 101 / 202 (master 1.272 / 1.267): shield gate 0.9 1.278 / 1.268, 0.8 1.270 / 1.266, 1.0
  1.249 / 1.250; Renew 9 1.270, Renew 8 1.258; Greater Heal at 70% 1.265; emergency 40% 1.264; Flash Heal first
  over 95% 1.249; Flash Heal 5 below 30% / 50% mana 1.273 / 1.274; gate 0.9 + Renew 9 1.273 / 1.272; gate 0.9 +
  Flash Heal 5 below 50% mana 1.277 / 1.276 (best, +0.005 / +0.009, Chromaggus 0.880 / 0.882, out at 106 s); all
  three 1.268 / 1.265. The best loses at 180 / 300 s (Golemagg, Chromaggus, Ragnaros, seed 101): 0.738 / 0.591
  against 0.750 / 0.612 (it lasts longer, Golemagg 237 s against 210 s, but carries less), so none is adopted.
  None lasts Chromaggus at carried damage. Not tried: Inner Focus or potion timing (both through
  `autocastOtherCooldowns`/the Greater Heal line). `rots.py` has the `fhlow` / `fhlowgate` knobs for the Flash
  Heal variant. The OOM looks like the Discipline kit's cost on a mana-starved fight, not a rotation line:
  Distilled Wisdom wins Chromaggus by 0.075 (`docs/healing_sim.md`, Consumables) but loses the mean.
- needs: worker fix

## Resolved

### healer_rotations-20261112-2: Restoration shaman raid healer spends Nature's Swiftness on cooldown; held for an emergency Healing Wave it carries +0.010 more
- severity: medium (+0.7% mean carry on two seeds; under the 1% line)
- affects: Restoration shaman raid healer (`ui/restoration_shaman/apls/raid_healer.apl.json`, `tools/healbench/req/shaman_restoration_raid.json`). Not the tank healer (it already holds NS for the tank under 50%; see below).
- found: 2026-11-12, master 3065ba744 (tankheal)
- evidence: the raid APL has no Nature's Swiftness (16188) line, so `autocastOtherCooldowns` (line 0) raises the buff at the pull and on every cooldown, and the next ordinary Chain Heal / Healing Wave takes it. Variant: insert as lines 1-2 `castSpell 16188` if Current Health % of Lowest Health < T, then `castSpell 25357` (Healing Wave 10) on Lowest Health if aura 16188 is active (the druid's pattern since e05135ad3). Paired, 1,000 it, own lengths, `-role raid`, mean carry of the 7 files (Onyxia / Broodlord / Chromaggus / Golemagg / Magmadar / Ragnaros / Vaelastrasz):
  - master seed 101: 0.927 / 0.912 / 1.084 / 2.630 / 1.827 / 1.239 / 1.125, mean 1.392
  - T=0.55 seed 101: 0.957 / 0.910 / 1.098 / 2.652 / 1.823 / 1.234 / 1.148, mean 1.403 (+0.011); seed 202: 1.3995 against 1.3894 (+0.010)
  - T=0.45: +0.010 (seed 101), +0.006 (seed 202); T=0.35: -0.000; T=0.45 with Lesser Healing Wave 6 in place of Healing Wave 10: +0.000
  Reproduce: `tools/audit/healer_rotations/baseline.sh`-style loop, or `healbench -role raid -encounter tools/encounters/<boss>.json -carry -iters 1000 [-seed 202] <request>`. Fairly sure of the sign (two seeds, 5 of 7 files up on each); the size is near the bench's resolution. Mana side (180 / 300 s) not measured: NS costs no mana and the swap changes one cast in 180 s.
- proposed fix: add the two lines above with T=0.55 (Healing Wave 10, the big rank, as the druid's Healing Touch), and say in the note that NS waits for an emergency.
- needs: worker fix
- resolved: 2026-11-12, Nature's Swiftness under 65% and an instant Healing Wave 10 at the lowest as lines 1-2 of the
  raid APL and request (line 0's note no longer names NS; `docs/healing_shaman.md`). Raid role, own lengths, 1,000
  it, mean carry, seed 101: autocast 1.392, 45% 1.402, 50% 1.411, 55% 1.403, 60% 1.395, 65% 1.412, 70% 1.406, 75%
  1.390; seed 202: autocast 1.389, 50% 1.404, 55% 1.400, 65% 1.411 (+0.020 / +0.022 for 65%).

### healer_rotations-20261112-1: Restoration druid tank healer's Nature's Swiftness threshold (60%) is too high since health values read seen health; 45% carries +0.027
- severity: high (+2.2% mean carry; every boss file up)
- affects: Restoration druid tank healer (`ui/restoration_druid/apls/tank_healer.apl.json` line 2, `tools/healbench/req/druid_restoration.json`, `orchestration/tools/healers/druid_restoration_apl.py`). The raid healer's 45% is still best.
- found: 2026-11-12, master 3065ba744 (tankheal)
- evidence: e05135ad3 picked 60% (Lowest Health) on true health; 65b5d49cf then made Current Health % of Lowest Health read the health the healer has seen. Bisect on the pre-consumables request: Magmadar 2.102 (3fc5898c9) -> 2.003 (65b5d49cf), unchanged since. Paired, 1,000 it, seed 101, own lengths, `-role tank`, carry (Onyxia / Broodlord / Chromaggus / Golemagg / Magmadar / Ragnaros / Vaelastrasz, mean):
  - master 60%: 1.386 / 0.873 / 1.010 / 1.864 / 2.004 / 0.629 / 0.927, 1.242
  - 45%: 1.403 / 0.893 / 1.023 / 1.951 / 2.045 / 0.634 / 0.932, 1.269 (+0.027, 7 of 7 up)
  - 35%: 1.268; 40%: 1.267; 50%: 1.258; 70%: 1.224
  Raid healer (45%): 35% 1.400, 55% 1.406 against 1.414, so it stays. Sure: a plateau from 35 to 50%, all above 60%.
- proposed fix: Nature's Swiftness on the tank healer at Lowest Health under 45% (was 60%), in the APL, the request, the generator and `docs/healing_druid.md`.
- needs: worker fix
- resolved: 2026-11-12, the tank healer's Nature's Swiftness at 45% (APL, request, `docs/healing_druid.md`; the
  generator's `ns_at=0.45`). Tank role, own lengths, 1,000 it, mean carry, seeds 101 / 202: 40% 1.267 / 1.268, 45%
  1.269 / 1.271, 50% 1.258 / 1.268, 60% 1.242 / 1.244 (+0.027 / +0.027, every file up on both seeds); 180 / 300 s
  (Onyxia, Golemagg, Chromaggus, seed 101) 1.150 / 0.946 against 1.145 / 0.940.

### healer_rotations-20261104-3: Restoration shaman tank healer's Lesser Healing Wave 5 line never fires (same mana gate as rank 6 just above it)
- severity: low (no carry to gain; a dead line in a shipped preset)
- affects: Restoration shaman tank healer (`ui/restoration_shaman/apls/tank_healer.apl.json` line 9, `tools/healbench/req/shaman_restoration.json`)
- found: 2026-11-04, master 0facd9428
- evidence: since the re-tune 324d64834, line 8 casts Lesser Healing Wave 6 (10468) on the tank whenever mana >= 50%, and line 9 casts rank 5 (10467) on exactly the same condition. Rank 6 is always affordable at 50% mana, so rank 5 can never be reached. Logged iterations (`healbench -role tank -encounter tools/encounters/{golemagg,chromaggus}.json -dur 300 -scale 1.5 -log`) show 10467 cast 0 times, against 10468 at 45 / 18 casts. Sure: the conditions are identical.
- proposed fix: drop line 9, or give rank 5 its own condition (for example a smaller tank deficit, or the 50-70% mana band) and measure it. No carry change is expected from dropping it.
- needs: worker fix
- resolved: 2026-11-12, line 9 (rank 5) dropped from the APL and the request, line 8's note says rank 6 only.
  Tank role, own lengths, 1,000 it, seed 101: 1.2571 with or without it (byte-identical rows); rank 5 on its own
  35-50% or 40-50% mana band 1.254 / 1.253, so it isn't kept.

### healer_rotations-20261111-1: Restoration druid raid healer loses its emergency Nature's Swiftness Regrowth to Wild Growth since NS also makes Wild Growth instant (-0.046 mean carry, Magmadar -0.184)
- severity: high (Magmadar 2.002 -> 1.818 carry, about 9%; every boss file down)
- affects: Restoration druid raid healer (`ui/restoration_druid/apls/raid_healer.apl.json`, `tools/healbench/req/druid_restoration_raid.json`). Not the tank healer (no Wild Growth line, rows unchanged).
- found: 2026-11-11, master 707ae92cc
- evidence: 9628e85b6 adds Wild Growth to Nature's Swiftness's affected spells (`sim/druid/healing.go` ~752). The APL leaves NS to `autocastOtherCooldowns` (line 1), which raises the buff at 0 s; before, the buff stayed up until line 4's emergency Regrowth (lowest under 60%) used it. Now the first Wild Growth (line 5, an ordinary cast) uses it. One logged Magmadar iteration: old rule, an instant Regrowth (9858, 0 s); master, an instant Wild Growth (1238215, 0 s) and Regrowth cast at 1.96 s. Raid role, own lengths, 1,000 it, default seed, ref -> master: Onyxia 0.997 -> 0.943, Broodlord 0.920 -> 0.906, Chromaggus 1.109 -> 1.092, Golemagg 2.660 -> 2.658, Magmadar 2.002 -> 1.818, Ragnaros 1.270 -> 1.240, Vael 1.102 -> 1.079. Bisected with benches built at master with one file reverted: reverting `sim/druid/healing.go` to 9628e85b6^ gives back the ref rows exactly; reverting `sim/core/dot.go` (f20d5938a) changes nothing.
- tried (paired, 1,000 it, master): (A) insert `castSpell 17116` with line 4's condition (lowest under 60%) just before line 4: 0.998 / 0.919 / 1.108 / 2.655 / 2.009 / 1.274 / 1.108, mean 1.439 against 1.391 (+0.048, same as before the hotfix). At 180 s: Magmadar 1.585 -> 1.660, Ragnaros 1.172 -> 1.191, Onyxia 0.823 -> 0.865. At 300 s: 1.467 -> 1.578, 1.059 -> 1.137, 0.762 -> 0.785, with no less mana left. It also runs unchanged on the old rule (same numbers to 0.003). (B) Wild Growth gated on NS not active: mean 1.406, Vael -0.095, worse than A. (C) NS line before Healing Touch 11 (line 9): mean 1.391, no gain (Wild Growth at line 5 still comes first).
- proposed fix: variant A. Add `{"action":{"castSpell":{"spellId":{"spellId":17116}},"condition":<line 4's condition>},"notes":"Nature's Swiftness for the emergency Regrowth below."}` as the new line 4 of the raid healer APL and its request, and correct line 1's note. Aside: the comment above `registerNaturesSwiftnessCD` (`sim/druid/healing.go` ~740) still says "Wild Growth isn't one".
- needs: worker fix
- resolved: 2026-11-12, e05135ad3 (both druid presets cast Nature's Swiftness on their own line; Subtlety 2 -> Heart of the
  Wild 2). Raid: NS on an instant Regrowth when the lowest is under 45% (the audit's variant A, 60%, was level at own
  lengths, 1.439 against 1.438, and lower at 180 / 300 s: Onyxia, Magmadar, Ragnaros 1.215 / 1.148 against 1.241 /
  1.158). Healbench 1,000 it, own lengths: Onyxia 0.963, Broodlord 0.919, Chromaggus 1.108, Golemagg 2.653, Magmadar
  2.053, Ragnaros 1.271, Vaelastrasz 1.097, mean 1.391 -> 1.438. Tank (re-tuned the same way): NS on an instant
  Healing Touch 11 under 60%, 1.235 -> 1.273, and 180 / 300 s 1.252 / 1.008 -> 1.264 / 1.016. NS on Wild Growth (45,
  60, 80%) loses everywhere. Screens and talents in `docs/healing_druid.md` (Nature's Swiftness on Wild Growth).

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

