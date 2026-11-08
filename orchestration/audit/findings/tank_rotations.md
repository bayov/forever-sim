# Findings: tank_rotations

## Open

### tank_rotations-20261104-11: Warrior Max survival's armor trinket held for the last 35 s wins only on Broodlord
- severity: medium
- affects: Protection warrior Max survival
- found: 2026-11-04 (pass 2), master b38ff16
- evidence: round 2 gave Max survival the armor on-use trinket 990617 (tools/tankbench/req/warrior_protection_max_survival.json:61) with both trinket lines held by `trinketAt=35` and a uses-lost guard (:146, :163); the lines do what their notes say. Against Balanced's choice, 991018 used on the pull (paired, 3k, B minus preset): -21.5 ± 10.8 on the scoring mean, but Broodlord -483 and every other boss but Onyxia favours 991018 (Chromaggus +94, Golemagg +49, Magmadar +78, Onyxia -38, Vaelastrasz +171; Ragnaros +182): without Broodlord, 991018 on the pull is about +71. The hold alone (990617 on the pull vs held 35 s): -104.7 ± 10.1, of which Broodlord -599. The same Broodlord-script dependency as finding 3 (tools/encounters/broodlord.json's first-ability timers, called "probably wrong" in that file). Sure: medium.
- proposed fix: decide with finding 3's ruling. If Broodlord's timers aren't trusted, Max survival goes back to 991018 on the pull (Balanced's trinket); if they are, keep it. Repro: `abtest -iters 3000 $P/A_warrior_protection_max_survival.json $P/ws_aptrink.json`.
- needs: user ruling
- round 3 (warrior, user call kept): 2026-11-04, a624264, new engine, 10k paired at w = 3, 990617 held 35 s minus 991018 on the pull: +12.8 ± 5.4 (seed 1), +15.9 ± 4.8 (seed 2); Broodlord +437 / +433, Onyxia +51 / +65, Magmadar -76 / -72, Golemagg -57 / -49, Chromaggus -101 / -106, Vaelastrasz -178 / -176; without Broodlord -72.0 ± 3.3 / -67.6 ± 3.7. Max survival keeps the armor trinket; listed as call 7 in docs/tank_warrior.md next to paladin Bulwark (-3). Left open: user call.
- pass 3: still open (user call 7, docs/tank_warrior.md). Hold 30 s against 35 s, 3k seed 11: -29.6 ± 5.7, so if the trinket stays it stays at 35 s.
- pass 4: still open (user call); a07941531 keeps 990617 held 35 s; not re-tested.
- pass 5: still open (user call); bba052f75 unchanged; not re-tested.
- pass 6: still open (user call); 9ba00c2dc unchanged; not re-tested.

### tank_rotations-20261104-3: Paladin's Templar's Bulwark timer was searched on a 10 s grid and misses an optimum at 13 s, which mostly fits Broodlord's script
- severity: high (by the numbers; see the caveat)
- affects: Protection paladin, all three presets
- found: 2026-11-04, master f62bfc74
- evidence: the Bulwark line's const `10s` (ui/protection_paladin/apls/level60_*.apl.json:8); `bulwarkAt` Step 10 (tools/rotopt/paladin_tank.go:48). Paired, 3k: 13 s gives Balanced +73.3 ± 5.0 (+2.9%), Max survival +219.9 ± 14.9 (+17%), Max threat +23.1 ± 1.5; 12 s gives +19.5 / +58.5 / +5.6, 14 s +47.7 (Balanced), 15 s -52.9. Almost all of the gain is Broodlord (+437 at Balanced): 13 s covers his first Mortal Strike, Knock Away and Blast Wave (tools/encounters/broodlord.json firstAt 25±5, 22.5±6.5, 20±7.5, timers the file calls "probably wrong" vmangos values). Vaelastrasz -62, Ragnaros -20 to -61.
- proposed fix: the knob gets Step 1–2 and a re-search. Adopt 13 s only if Broodlord's first-ability timers are trusted; otherwise keep 10 s and fix the timers first. Reproduce B: `tools/audit/tank_rotations/mk_paladin.py A.json B.json bulwark_13s`.
- needs: user ruling
- left open: 2026-11-04 (tank-paladin, e9751cc). bulwarkAt now steps by 1 sec, and rotopt holds it at 10 (`-freeze bulwarkAt`). Paired, 3,000 iterations, 13 s minus 10 s: on the five scoring bosses without Broodlord Max threat -1.1 ± 1.1, Balanced -6.2 ± 3.0, Max survival -18.5 ± 9.0 (Vaelastrasz -24 / -66 / -199, Onyxia +9 / +26 / +79); with Broodlord +16.5 ± 1.4 / +49.3 ± 3.9 / +147.9 ± 11.7 (Broodlord +105 / +327 / +980). 13 s wins only on Broodlord's suspect timers, so 10 s stays. The user's call (docs/tank_paladin.md, Calls): 13 s if Broodlord's timers are trusted, and Bulwark on a timer every pull at all.
- pass 2: still open (user ruling). Bulwark 13 s vs 10 s without Broodlord: -3.6 mean over 5 disjoint 3k seeds (-0.1 to -9.4), -4.4 / -4.2 at 10k, so 10 s holds there; the sign rests on Vaelastrasz (-55) against Onyxia (+23) and Magmadar (+11), i.e. on the encounter set, not on noise. Finding 11 below is the warrior's version of the same Broodlord question.
- pass 3: still open (user ruling, docs/tank_paladin.md call 7); not re-tested.
- pass 4: still open (user ruling); a07941531 keeps 10 s at all three presets; not re-tested.
- pass 5: still open (user ruling); bba052f75 unchanged; not re-tested.
- pass 6: still open (user ruling); 9ba00c2dc unchanged; not re-tested.

## Resolved

### tank_rotations-20261104-14: Warden Max threat and Balanced leave Searing Totem and Fire Nova unused; both in the last 60 s gain about 2% and 8%
- severity: high (Max threat +1.9%, Balanced +8%)
- affects: Warden shaman Max threat, Balanced (Max survival unconfirmed)
- found: 2026-11-04 (pass 3), master dc303b0
- evidence: the Warden priority lists end at Stormstrike, with no fire totem and no Fire Nova (tools/tankbench/req/shaman_warden_max_threat.json:221, shaman_warden_balanced.json:271; ui/warden_shaman/apls/level60_{max_threat,balanced}.apl.json). B adds Searing Totem (10438) before the pull, ahead of the water totem; in the fight, Searing Totem recast when down with at least 10 s left, and Fire Nova (408345) after Stormstrike at 40% mana or more, both only while remainingTime <= 60 s. Balanced's B also has finding 15's Mana Spring. Paired, 3k, B minus preset: Max threat +43.2 ± 9.7 (seed 21), +46.4 ± 8.4 (22), +48.4 ± 9.2 (23); Vaelastrasz +144 / +121 / +133, Golemagg, Magmadar and Onyxia +36 to +53, Chromaggus -37 / -34 / -39, Ragnaros -45 / +17 / -45. Balanced +70.5 ± 15.0 / +48.8 ± 13.1 / +76.0 ± 12.8; Vaelastrasz +171 / +143 / +156, Chromaggus +18 / +10 / -2. Max survival +22.6 ± 32.2 / +71.8 ± 36.3, unconfirmed. Without the 60 s gate Chromaggus loses about 100 (Earth Shock short of mana, 16.7 to 13.5 casts) and the gain doesn't confirm at Balanced (+52.4 ± 13.1 / +19.3 ± 13.6); a 40 s gate ties 60 s, 90 s -3.3 ± 3.2. Round 3's rotopt search moved one knob at a time from fireNova=0: fireTotem=1 alone was -18 to -38 inside its noise, and fireNova=1 does nothing without a fire totem, so the pair was never tried. Part of the Vaelastrasz gain is mana spent under his mana regen and given back as threat (mana-gain TPS 29.7 to 57.4), so it rests on fd9538c's mana-gain threat, which docs/tank_druid.md call 6 still lists as unverified; tank_mechanics territory. Sure: high that it scores, medium on the size.
- proposed fix: a rotopt knob that searches fireTotem and fireNova together, with a "last N s" gate (40-90), then a re-search of all three Warden presets. Repro: `abtest -iters 3000 -seed 22 tools/tankbench/req/shaman_warden_max_threat.json $P/w_f1n1g60b_max_threat.json` and `... shaman_warden_balanced.json $P/w_f1n1w1g60b_balanced.json` (P=$AUDIT_WORK/tank_rotations/probe/v; the variant files are scratch, so rebuild B from the lines above).
- needs: worker fix
- fixed: 2026-11-04, 4745ce4 (branch warden-trinket). rotopt's shaman_tank has a `fireLast` knob (fire totem and Fire Nova only in the last N s; the prepull totem still goes down); searched with fireTotem/fireNova together at all three presets (3k s3 screens of gates 0-90 s, Magma and Flametongue, Fire Nova's mana floor), confirmed at 10k on seeds 1 / 2: Searing + Fire Nova at 60 s, Max threat +53.9 ± 4.7 / +47.7 ± 5.0, Balanced (with Mana Spring) +67.7 ± 7.1 / +59.2 ± 6.4, Max survival +59.3 ± 17.7 / +55.7 ± 18.2. All three presets take it (docs/tank_shaman.md, Round 4 A/Bs).

### tank_rotations-20261104-15: Warden Balanced doesn't put Mana Spring down before the pull, as Max threat does
- severity: medium (+1.4%)
- affects: Warden shaman Balanced (Max survival unconfirmed)
- found: 2026-11-04 (pass 3), master dc303b0
- evidence: Max threat's prepull is Lightning Shield -4.5 s, Stoneskin -3 s, Mana Spring -1.5 s, with an in-fight Mana Spring line; Balanced has no Mana Spring (tools/tankbench/req/shaman_warden_balanced.json prepullActions). B = Max threat's prepull and its Mana Spring line. Paired, 3k: +12.7 ± 3.9 (seed 21), +10.6 ± 4.4 (22), +10.8 ± 4.6 (23); Chromaggus +37 / +32 / +43, the rest within ±10. Max survival +8.6 ± 9.8 / +11.0 ± 9.8, unconfirmed. Put down only in the fight: +0.1 ± 16.8 (1k). Round 3's search saw waterTotem=1 at +8.2 / +8.6, inside its noise. Healing Stream instead: +24.2 ± 14.4 / +0.5 ± 14.4, swinging on Broodlord, unsettled. Sure: medium-high.
- proposed fix: waterTotem=1 (Mana Spring before the pull) at Balanced, searched together with finding 14. Repro: `abtest -iters 3000 -seed 22 tools/tankbench/req/shaman_warden_balanced.json $P/w_manaspringpp_balanced.json`.
- needs: worker fix
- fixed: 2026-11-04, 4745ce4 (branch warden-trinket). waterTotem=1 at Balanced, 10k: +9.3 ± 2.3 (seed 1), +10.8 ± 1.8 (seed 2); taken. Max survival +5.3 ± 5.8 / +4.0 ± 5.1, inside its error: not taken.

### tank_rotations-20261104-13: rotopt's -seed overlaps the search's iterations, and the six-boss score's error is understated by about a quarter
- severity: medium (tooling; no preset choice found wrong)
- affects: every tank preset decision made with rotopt, tankbench or abtest
- found: 2026-11-04 (pass 2), master b38ff16
- evidence: the druid worker's report that "-seed has no effect" and that 3k and 10k means differ by 20-40 at the same build. -seed does reach the sim (tools/rotopt/main.go:90,99; tools/rotopt/run.go:210), but iteration i runs on seed RandomSeed+i (sim/core/sim.go:298-299, :428; split kept in sim/core/sim_concurrent.go:42-50), so -seed 2 at 3k replays 2,999 of seed 1's 3,000 iterations (Onyxia, bear: seed 1 1163.3, seed 2 1163.2, seed 1000001 1111.7), and a 10k confirm at one seed re-uses the 3k search's iterations. The 3k-vs-10k gap is sampling luck, not a bias in N: healing needed is the empirical 95% quantile (sim/core/tank_score.go:203-229), SafeDPS and the death floor don't depend on N, and bear Balanced at 30k is 1843.0 against 1844.5 ± 3.9 over 25 independent 1k seeds; seed 1's first 3,000 iterations are +30 (2.7 SD) lucky. Bosses run at one seed share iteration seeds, so their scores correlate (0.11-0.14) and the six-boss mean's SD is about 1.25× what rotopt (tools/rotopt/run.go:119-121, gainNoise) and abtest's scoring_mean assume (bear 19.3 vs 14.9 at 1k). Paired differences are fine: four of round 2's small calls (Shield Wall 17 vs 16 s, Sunder filler 15 vs 30, Bulwark 13 vs 10 without Broodlord, bear in-fight Enrage) keep their sign on 5 disjoint 3k seeds and two 10k seeds, and the reported paired ± covers the seed-to-seed spread. Sure: high. Scratch: $AUDIT_WORK/tank_rotations/noise/.
- proposed fix: (a) derive each iteration's seed from hash(seed, i) instead of seed+i (sim/core/sim.go:428), or step -seed by 1e6, and add -seed to tankbench and abtest; (b) rotopt's confirm and every "re-check at 10k" on a seed disjoint from the search's; (c) give each encounter its own seed offset, or multiply the scoring-mean error by 1.25 for absolute scores; (d) report absolute preset scores at one fixed count (10k or more) and never compare them across seeds or counts without their ±; (e) adopt a change only at |diff| > 2× paired error, confirmed on a second disjoint seed; gains under about 3 at 3k stay ties.
- needs: orchestrator fix
- fixed (a, c): 2026-11-04, 9e01b3e. tankscore.Seed maps -seed s to 1 + (s-1)*1e9 for tankbench, abtest and rotopt's tank runs (seed 1 unchanged, so existing numbers stand; other seeds disjoint below 1e9 iterations; sim.go's global seed+i kept so DPS results don't move); unpaired six-boss mean errors x1.25 in rotopt and abtest; docs/tank_sim.md "Seeds". (b), (d), (e) are practice for the workers: confirm on a different -seed.
- pass 3: (a) and (c) checked: tankscore.Seed gives disjoint seeds (11-14 and 21-23 used this pass), and the round-3 workers confirmed their calls on a second seed (warrior 2, bear 2, Warden 4 and 5), so (b) and (e) were followed. Every round-3 call re-checked on three more disjoint seeds keeps its sign (findings 9, 10, 12, the warrior's Shield Wall). (d) is practice.
- resolved: 2026-11-04, 9e01b3e

### tank_rotations-20261104-12: Bear Max threat's in-fight Enrage under 60 Rage sits at the knob's edge; higher wins a little
- severity: low
- affects: Feral bear Max threat
- found: 2026-11-04 (pass 2), master b38ff16
- evidence: ui/feral_tank_druid/apls/max_threat.apl.json:8, tools/tankbench/req/druid_bear_max_threat.json:128 (`currentRage < 60`); enrageRage's Max is 60 (tools/rotopt/druid_tank.go:61). Under 80 Rage, paired, 3k: +3.0 ± 0.5 (every boss ≥ 0 but Chromaggus -2.9 ± 2.1; Ragnaros +9.0 ± 4.7); 1k: under 70 +4.0 ± 0.7, under 90 +3.8, no condition +3.7, under 40 -21.8. Balanced's 40 holds (60: -8.1 ± 12.0; 20: -23.9 ± 10.3). Sure: high, but small (0.1%).
- proposed fix: raise the knob's Max to 100 (or allow "on cooldown") and re-search Max threat. Repro: `abtest -iters 3000 $P/A_druid_bear_max_threat.json $P/bt_enr80.json` (P=$AUDIT_WORK/tank_rotations/probe).
- needs: worker fix
- fixed: 2026-11-04, ad93b95. enrageRage's Max is 100 (tools/rotopt/druid_tank.go); the knob search on the new engine picks 70, +4.0 ± 0.5 over 60 at 3,000 on the disjoint seed 2 (80 +4.0, 100 +3.9: the same). Max threat's request and UI APL enrage under 70. docs/tank_druid.md, Presets (round 3).
- pass 3: holds. A/Bs against under 70, paired, 3k on seeds 11, 12 and 13: under 60 -3.4 ± 0.4 / -5.3 / -3.9; under 80 +0.5 ± 0.3 / +0.3 / +0.5, and +0.4 ± 0.1 at 10k on seed 14; no condition +0.3 / -0.1 / +0.2. 70 to 100 is a plateau: 80's +0.4 is 0.01% of 2,842, a tie, not worth a re-search.
- resolved: 2026-11-04, ad93b95

### tank_rotations-20261104-10: Warden Balanced loses about 96 by wearing 990015 on cooldown instead of the armor trinket held for the last 30 s
- severity: high (+17% of a 548 score)
- affects: Warden shaman Balanced
- found: 2026-11-04 (pass 2), master b38ff16
- evidence: tools/tankbench/req/shaman_warden_balanced.json trinket slot (990015) and the autocast line (:189; ui/warden_shaman/apls/level60_balanced.apl.json:10). B = 999216 in slot 14, held as in finding 9. Paired, 3k: +95.6 ± 7.3 (Broodlord +372, Onyxia +136, Golemagg +22, Vaelastrasz +18, Chromaggus +14, Magmadar +11; Ragnaros -8 ± 24); without Broodlord about +36, still a win. 1k: hold 25 s +95, 35 s +72. At Max threat the same swap loses -26.3 ± 3.8, so leave Max threat. Sure: medium-high. This is a rotation and gear choice, not the class gap ruled on in standards 40.
- proposed fix: with finding 9's knob, let rotopt choose the armor trinket at Balanced and re-search. Repro: `abtest -iters 3000 $P/A_shaman_warden_balanced.json $P/sb_armt30.json`.
- needs: worker fix
- fixed: 2026-11-04, 6a9dea8. Balanced wears 999216 held for the last 25 s: +98.5 ± 7.3 against 990015 on cooldown (round 2 build); on the final build 990015 loses -112.3 ± 8.1; 25 s over 30 s at 10k seeds 4 and 5 (+16.6 ± 4.2, +8.9 ± 3.2). Max threat keeps 990015 (-12.2 ± 4.2 held). Docs 6c73945.
- pass 3: holds. Balanced, 25 s against: 30 s +2.2 ± 9.2 / -5.0 ± 7.9 / -12.6 ± 7.7 (3k, seeds 11-13), -17.5 ± 4.1 (10k, seed 14; Broodlord -117); 20 s -20.9 / -14.5 / -20.1 (± 6 to 7.5); no hold (autocast) -126.2 / -148.7 / -143.8.
- resolved: 2026-11-04, 6a9dea8

### tank_rotations-20261104-9: Warden Max survival fires its armor trinket on the pull; held for the last 25-30 s it gains about 330
- severity: high (+7.5% of a -4357 score)
- affects: Warden shaman Max survival
- found: 2026-11-04 (pass 2), master b38ff16
- evidence: armor trinket 999216 is used by autocastOtherCooldowns, "trinkets on cooldown" (tools/tankbench/req/shaman_warden_max_survival.json:189, ui/warden_shaman/apls/level60_max_survival.apl.json:10); the template has no hold knob (tools/rotopt/shaman_tank.go:64, :149, :154), though the warrior's got `trinketAt` in round 2. B = an explicit top line for 999216 with the warrior's hold condition, (remainingTime <= 30s) or spellUsesLostByDelay(remainingTime - 30s, minTimeLeft 20s) > 0. Paired, 3k: +328.5 ± 17.0 (Broodlord +1459, Onyxia +284, Chromaggus +125, Golemagg +53, Magmadar +51, Vaelastrasz 0; Ragnaros +78 ± 63); without Broodlord about +92, still a win. Hold scan at 1k: 25 s +350, 30 s +334, 35 s +273, 40 s +186, 45 s +71. Sure: high.
- proposed fix: a `trinketAt` knob in shaman_tank.go like warrior_tank.go's, and a re-search of all three Warden presets (30 s for Max survival). Repro: `abtest -iters 3000 $P/A_shaman_warden_max_survival.json $P/ss_trg30.json`.
- needs: worker fix
- fixed: 2026-11-04, df481a4 (template knob `trinketAt`) and 6a9dea8 (searched: Max survival holds 999216 for the last 25 s, +344.3 ± 15.3 against autocast; 25 s over 30 s at 10k seeds 4 and 5, +29.5 ± 10.2 and +33.2 ± 11.0; docs 6c73945).
- pass 3: holds. Max survival, 25 s against: 30 s -29.3 ± 22.2 / -41.2 ± 19.0 / -15.0 ± 16.5 (3k, seeds 11-13), -34.4 ± 10.0 (10k, seed 14; Broodlord -260); 20 s -73.5 / -73.8 / -48.2 (± 18 to 25); autocast on the pull -359.8 / -398.3 / -353.5. 25 s, which round 3 confirmed only at 10k, now holds on four more disjoint seeds.
- resolved: 2026-11-04, df481a4 and 6a9dea8

### tank_rotations-20261104-1: Bear's Demoralizing Roar takes the pull's first GCD and costs about a tenth of the score
- severity: high
- affects: Feral bear, all three presets
- found: 2026-11-04, master f62bfc74
- evidence: Roar is the first priority line, so it takes the first GCD and 10 Rage ahead of Primal Bite and Lacerate, while SafeDPS is set in the first seconds (tools/rotopt/druid_tank.go:103 appends it above them; ui/feral_tank_druid/apls/*.apl.json:9, max_survival :10; tools/tankbench/req/druid_bear_{balanced,max_threat}.json:141, druid_bear_max_survival.json:164). Moving the line to just below "Lacerate to 5 stacks" (paired, `tools/audit/tank_rotations/abtest`, scoring mean): Balanced +197.5 ± 16.7 (1k; 1855 → ~2053, +10.6%), Max threat +207.8 ± 11.7 (1k), Max survival +176.9 ± 30.0 (3k). Every scoring boss gains (+160 to +244 at Balanced); Ragnaros +128 / +150. Healing needed moves -5 to +29 per boss. Removing Roar wins more (+224) but breaks the 11-02 ruling that the tank applies its own Roar, so it isn't proposed. Roar right after Primal Bite: only +40. Sure: high.
- proposed fix: Roar's line just below the Lacerate keep-up line in all three APLs and requests. Give the template a knob (e.g. `roarAfterLacerate`), since the hard-coded order kept rotopt from finding it. Reproduce B: `tools/audit/tank_rotations/mk_bear.py balanced B.json 'r_=pl.pop(1); pl.insert(3,r_)'` (`pop(2)/insert(4)` for max_survival).
- needs: worker fix
- fixed: 2026-11-04, bde6912 (rotopt knob `roarAfterLacerate`), f0ff977 (all three presets; +218 / +191 / +270 at own weight from the move alone, 1k)
- pass 2: holds. Roar below the Lacerate keep-up line in all three (ui/feral_tank_druid/apls/*.apl.json:11-12); moving it back to the top costs -234.9 ± 8.3 / -212.6 ± 11.6 / -120.3 ± 30.0 (Max threat / Balanced / Max survival, paired, 3k).
- resolved: 2026-11-04, f0ff977

### tank_rotations-20261104-2: Warrior's Shield Wall in the last 12 s is too late; the last 16–18 s wins at Balanced and Max survival
- severity: high
- affects: Protection warrior Balanced, Max survival
- found: 2026-11-04, master f62bfc74
- evidence: line `remainingTime <= 12s` (ui/tank_warrior/apls/level60_{balanced,max_survival}.apl.json:11, tools/tankbench/req/warrior_protection_{balanced,max_survival}.json:162). rotopt's `wallAt` grid is Min 12, Step 12 (tools/rotopt/warrior_tank.go:67), so 13–23 s was never searched (contradicts docs/tank_warrior.md:283-285, "no template knob moves Max survival"). Paired, 3k, scoring mean: 18 s gives Balanced +46.0 ± 7.1 (+1.2%) and Max survival +175.3 ± 19.7 (+11.8%); 16 s gives +37.1 ± 5.8 and +137 ± 16; 20 s +20 / +110; 22 s loses. Vaelastrasz (17 ± 8 s fight) goes the other way: -107 / -283 at 18 s, -56 / -132 at 16 s. Ragnaros +13 / +52 (noise). Max threat at 18 s is -2.4 ± 3.9, so leave it at 12 s.
- proposed fix: 18 s (or 16 s, which costs Vaelastrasz half as much) for Balanced and Max survival, and a finer grid in the template (Min 6, Step 2) with a re-search. Reproduce B: `tools/audit/tank_rotations/mk_warrior.py tools/tankbench/req/warrior_protection_balanced.json B.json setcond 4 '{"cmp":{"lhs":{"remainingTime":{}},"op":"OpLe","rhs":{"const":{"val":"18s"}}}}'`
- needs: worker fix
- fixed: 2026-11-04, 5d9a3aa18 (warrior): Shield Wall in the last 17 s for Balanced and Max survival (+20.7 ± 5.4, +113.2 ± 13.5 paired 3k vs 12 s; Vaelastrasz -112 / -163), Max threat stays 12 s; template grid 1 s (d9e001609).
- pass 2: holds. remainingTime <= 17s (Balanced, Max survival :11), 12 s at Max threat. Paired, 3k, variant minus preset: Balanced 12 s -27.4 ± 6.1, 16 s -8.8 ± 4.2, 18 s -5.6 ± 5.0; Max survival 12 s -89.9 ± 12.8, 16 s -16.2 ± 10.4, 18 s -20.6 ± 11.0. 16 vs 17 s holds on 5 disjoint seeds and at 10k (-5.4, -7.0).
- resolved: 2026-11-04, 5d9a3aa18

### tank_rotations-20261104-4: Warden's Stormstrike sits above Earth Shock, which makes more threat per GCD
- severity: high (Max threat); medium elsewhere
- affects: Warden shaman, all three presets
- found: 2026-11-04, master f62bfc74
- evidence: Earth Shock does about 2,225 threat a cast against Stormstrike's 1,500 (Golemagg, -actions), and with Stormstrike first it waits whenever both are ready. Order hard-coded at tools/rotopt/shaman_tank.go:150; Stormstrike at tools/tankbench/req/shaman_warden_*.json:194/204/208. Swapping them, paired, 3k, scoring mean: Max threat +62.0 ± 16.2 (+2.4%); Balanced +51.2 ± 48.6; Max survival -18.7 ± 148. SafeDPS rises on every scoring boss at all three weights (Vaelastrasz +75, Onyxia +43); the Balanced and Max survival error comes from Broodlord's death floor near 5% (without Broodlord: +67 / +48).
- proposed fix: emit the shock before Stormstrike in the template (swaps lines 12↔13 in level60_max_threat/max_survival, 11↔12 in level60_balanced, and the requests). Confirm Balanced and Max survival with a rotopt -eval first. Reproduce: `tools/audit/tank_rotations/mk_warden.py <preset> swap 'i=[j for j,a in enumerate(pl) if a["action"].get("castSpell",{}).get("spellId",{}).get("spellId")==17364][0]; pl[i],pl[i+1]=pl[i+1],pl[i]'`.
- needs: worker fix
- fixed: 2026-11-04, 952a7a4 (template knob `shockFirst`, default 1) and dd4b82a (the three presets re-searched with it; paired: Max threat +91.6 ± 10.6, Balanced +105.0 ± 12.5, Max survival +129.3 ± 33.1).
- pass 2: holds. Earth Shock above Stormstrike in all three; swapping back costs -99.0 ± 6.1 / -88.1 ± 11.6 / -95.9 ± 34.8 (3k).
- resolved: 2026-11-04, dd4b82a

### tank_rotations-20261104-5: Bear Max survival's Enrage at -3 s loses to Balanced's -1 s
- severity: medium (+46, about 5% of a negative score of -844)
- affects: Feral bear Max survival
- found: 2026-11-04, master f62bfc74
- evidence: prepull shift -4.5 s and Enrage -3 s (tools/tankbench/req/druid_bear_max_survival.json:111, :125), reasoned in docs/tank_druid.md:224 as "so its armor cut ends sooner". Shift -2.5 s and Enrage -1 s, paired, 3k: +45.6 ± 16.6 on top of finding 1, +60.1 ± 17.6 on today's preset. Chromaggus -6 and Magmadar -13 (noise); Ragnaros -50 ± 200.
- proposed fix: shift at -2.5 s and Enrage at -1 s, as Balanced does, and fix the doc's reason. Reproduce: `tools/audit/tank_rotations/mk_bear.py max_survival B.json 'r_=pl.pop(2); pl.insert(4,r_); pre[0]["doAtValue"]["const"]["val"]="-2.5s"; pre[1]["doAtValue"]["const"]["val"]="-1s"'`.
- needs: worker fix
- fixed: 2026-11-04, f0ff977. Confirmed (-1 s beats -3 s by +50.7 ± 16.0, 3k), but the re-search puts Enrage in the fight at every weight, which beats both; the doc's reason is gone with it
- pass 2: holds. Enrage in the fight under 40 Rage (Balanced, Max survival :9); pre-pull -1 s loses by 60 (5 disjoint seeds at 3k: -46 to -69; 10k: -60, -70).
- resolved: 2026-11-04, f0ff977

### tank_rotations-20261104-6: Paladin's on-use trinket waits behind every GCD line
- severity: medium
- affects: Protection paladin, all three presets
- found: 2026-11-04, master f62bfc74
- evidence: `autocastOtherCooldowns` is the last line (ui/protection_paladin/apls/level60_*.apl.json:15, tools/rotopt/paladin_tank.go:117), noted "on a free GCD", but trinket 990716 (+165 Str, 20 s) is off the GCD, so the line only delays it. Moved to the top, paired, 3k: Balanced +13.9 ± 0.6 (+0.5%), Max threat +14.1 ± 0.2, Max survival +13.4 ± 1.8 (+1.1%); mostly Vaelastrasz (+44). Ragnaros -2 to -21.
- proposed fix: move the line to the top, with the note "Trinkets on cooldown from the pull, off the global cooldown." Reproduce: `tools/audit/tank_rotations/mk_paladin.py A.json B.json autocast_top`.
- needs: worker fix
- fixed: 2026-11-04, 86b9631 (the presets; rotopt's template in e9751cc).
- pass 2: holds. autocastOtherCooldowns is line 8 in all three; moving it last costs -12.9 ± 0.2 / -12.3 ± 0.4 / -11.0 ± 1.3 (3k).
- resolved: 2026-11-04, 86b9631

### tank_rotations-20261104-7: Warrior Max threat's Sunder filler at 30 Rage is a little high
- severity: low
- affects: Protection warrior Max threat
- found: 2026-11-04, master f62bfc74
- evidence: priorityList[10] `currentRage >= 30`; `>= 15` gives +3.9 ± 1.1 (3k, paired, every boss ≥ 0, Ragnaros +1.9 ± 10.6). Bloodrage on cooldown gives +1.4 ± 0.5 too (noted only).
- proposed fix: Sunder filler at 15 Rage, if the template is re-searched anyway.
- needs: worker fix
- fixed: 2026-11-04, 5d9a3aa18: Sunder filler at 15 Rage on the re-searched Max threat talents (+6.9 ± 1.6; -0.3 ± 0.8 on the old ones).
- pass 2: holds. currentRage >= 15 (level60_max_threat.apl.json:17); 30 loses -6.9 ± 1.6 (3k), -5.4 to -8.1 on 5 disjoint seeds, -5.7 at 10k.
- resolved: 2026-11-04, 5d9a3aa18

### tank_rotations-20261104-8: Rotation notes that don't say what the line does (standards 31)
- severity: low
- affects: Warden shaman, Protection warrior, Feral bear
- found: 2026-11-04, master f62bfc74
- evidence: (a) Warden's shield line, a Lightning Shield cast, is noted "Water Shield needs the Restoration talent" (tools/rotopt/shaman_tank.go:118, every level60_*.apl.json). (b) Warrior's Defensive Stance line (spell 71) has no note in all three APLs. (c) docs/tank_druid.md:217 says "Enrage at the pull and 1 sec before it", but the APL casts it only before the pull (an in-fight Enrage loses -5.3 ± 1.3). Bear Max survival's Frenzied Regeneration at 100 Rage fires about once, at the pull's Rage peak, not "on cooldown" as its note says (the line itself is right: removing it costs 432).
- proposed fix: (a) "Lightning Shield whenever its orbs are gone: their threat beats the 370 mana." (b) "Defensive Stance kept up: the tank's stance." (c) the doc: "Enrage 1 sec before the pull"; the Frenzied Regeneration note: "at the pull's Rage peak".
- needs: worker fix
- fixed (shaman part, (a)): 2026-11-04, 952a7a4 (the template's shield note; the APLs carry it since af78f8a).
- fixed (warrior part b): 2026-11-04, c588675a6 and 61db5cccb. (a) and (c) are the shaman and druid workers'.
- fixed (bear part): 2026-11-04, bde6912 (Frenzied Regeneration's note: "at the pull's Rage peak"), f0ff977 (the doc's Enrage wording, rewritten with the new presets)
- pass 2: all three parts hold (Warden shield note level60_*.apl.json:8-9; warrior Defensive Stance note balanced/max_survival:19, max_threat:18; bear FR note :8; docs/tank_druid.md:164 labels the pre-pull Enrage as round 1).
- resolved: 2026-11-04, f0ff977
