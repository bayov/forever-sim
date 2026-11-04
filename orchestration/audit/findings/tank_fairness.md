# Findings: tank_fairness

## Open

### tank_fairness-20261104-9: Stale tank notes and docs after round 2
- severity: low
- affects: Protection paladin and Warden UI preset tooltips; docs/tank_sim.md; ref/tank_bench/summary.tsv header
- found: 2026-11-04, master b38ff16
- evidence: ui/protection_paladin/presets.ts:140-147: the Max survival tooltip doesn't say it is Balanced's build, though the requests are byte-identical. ui/warden_shaman/presets.ts:120-130: nothing says Balanced and Max survival are within paired error (docs/tank_shaman.md says so). docs/tank_sim.md:8-14 'Where it stands (2026-11-03)' still has warrior and paladin on pre-raid plate and the Warden 'next'. At about line 450 it says 'class workers re-search them next', with no round-2 table. At about 478 'Bear Max survival (1185) now beats Balanced (1114)' was fixed in f0ff977. At about 561-565 the bear has 'Maul at 30 Rage' and 'not in its rotation yet: Enrage, Frenzied Regeneration'. orchestration/audit/ref/tank_bench/summary.tsv's header says master 23b2ec2, but its rows are round 2's (they match this pass's 3000-iteration run within noise on all 104 rows). docs/tank_paladin.md's Max survival 1363.3 is 1336.7 ± 13.7 at 3000 iterations (Balanced's file at w=3).
- proposed fix: one line in each tooltip; refresh docs/tank_sim.md's status and bear sections; update the summary.tsv header and the paladin Max survival number.
- needs: worker fix
- fixed (paladin): 2026-11-04, e37867b (Max survival tooltip says it is Balanced's build) and 73d70c283 (summary.tsv paladin rows re-benched for round 3; docs/tank_paladin.md's numbers from the 3,000-iteration matrix, round 2's 1,363 noted as 1,337 ± 14 at 3,000).

### tank_fairness-20261104-8: Round 2 searched the four tanks unevenly
- severity: low (no measured loss yet, but the search coverage differs)
- affects: Warden shaman, Protection paladin, Feral bear
- found: 2026-11-04, master b38ff16
- evidence: from the round-2 sections of docs/tank_*.md. The warrior re-searched all 9 races at 3 weights, A/B'd armor against AP trinket at 3 weights, and A/B'd wrist Superior Deflection (Defense +9) against Stamina, taking Deflection. The Warden's races and trinket A/B were not re-run on the new engine (docs/tank_shaman.md:65, 'Not re-run this round'). The engine change moved the autocast trinket, so the round-1 picks are unverified. The paladin ran no gear, trinket, enchant or libram A/B in round 2, and its wrist, boots and shield enchants have no Deflection or Defense check like the warrior's. The bear's doc says 'enchants and consumables were not re-checked'. Only the warrior committed raw search output (orchestration/tank_handoff/warrior/round2/); the other three can't be re-audited from files. Tauren led the Warden's round-1 race search by a wide margin, so a race change is unlikely.
- proposed fix: Warden race at 3000 iterations and its trinket A/B on the current engine. Paladin and Warden wrist Superior Deflection against Stamina, and shield and boots enchant A/B, at the three weights. Bear enchant re-check. Commit each class's search summary next to the warrior's.
- needs: worker fix
- fixed (paladin): 2026-11-04, b07bdf476: wrist, cloak, shield and boots enchant A/Bs, talents, knobs and races at the three weights on the current engine, each choice confirmed on a disjoint seed; gear rows re-checked (tools/audit/tank_gear/paladin.py passes). Raw output, scripts and commands in orchestration/tank_handoff/paladin/round3/.

### tank_fairness-20261104-7: Warden Max threat takes Titans where Natural Accuracy wins at its weight
- severity: medium (+30 at w=1/3, 1.3% of the preset's score)
- affects: Warden shaman Max threat (request and UI preset)
- found: 2026-11-04, master b38ff16
- evidence: pass 1's fix -1 asked for the flask to be re-picked at the Max threat weight; every class still uses one flask on all three presets. tankbench, 3000 iterations, seed 1, six-boss mean at w=1/3, *_max_threat.json with flask FlaskOfNaturalAccuracy (each mean ± about 4 to 6): Warden 2319.2 against 2288.8 (+30.4); warrior 3673.8 against 3696.8 (-23.0, Titans right); paladin 2781.8 against 2781.8 (0). Variant requests were in $AUDIT_WORK/tank_fairness/var/ (scratch).
- proposed fix: FlaskOfNaturalAccuracy on shaman_warden_max_threat.json and its UI preset, then re-bench.
- needs: worker fix

### tank_fairness-20261104-6: The paladin and Warden UI presets load without Kings and Might; the warrior and Warden UIs use other professions than their benches
- severity: medium (a user loading these presets in the sim gets a weaker tank than the bench ranks; it doesn't move tankbench)
- affects: Protection paladin, Warden shaman, Protection warrior UI defaults
- found: 2026-11-04, master b38ff16
- evidence: every tankbench request has individual buffs blessingOfKings plus improved blessingOfMight, and the Warden's also has improved Wisdom; all have Engineering + Blacksmithing. ui/protection_paladin/presets.ts:177-178: DefaultIndividualBuffs has only improved Blessing of Wisdom (no Kings, no Might). ui/warden_shaman/presets.ts:170-178: no blessings at all. The warrior and bear UIs have Kings + improved Might. Professions: ui/tank_warrior/presets.ts:198-199 Blacksmithing + Enchanting and ui/warden_shaman/presets.ts:192-193 Alchemy + Enchanting, while the paladin and bear match the requests. Improved Wisdom on the paladin is worth nothing (Balanced six-boss mean 2385.8 with it against 2387.3 without, 3000 iterations), so the requests are right and the UI is off.
- proposed fix: set each UI's DefaultIndividualBuffs and professions to its request's (Kings + improved Might; plus improved Wisdom for the Warden; Engineering + Blacksmithing).
- needs: worker fix
- fixed (paladin): 2026-11-04, e37867b: ui/protection_paladin/presets.ts individual buffs (Kings, improved Might), raid buffs, debuffs, options (Seal of Fury, Devotion Aura) and professions equal tools/tankbench/req/paladin_protection_*.json; consumables already did. Max threat gear set split in b07bdf476 so each preset's gear equals its request.

### tank_fairness-20261104-2: Paladin takes about half the warrior's damage at the same armor, avoidance and block share
- severity: medium (sure of the numbers, not of the cause; the paladin leads Max survival on Onyxia and Ragnaros)
- affects: Protection paladin, all presets
- found: 2026-11-04, master f62bfc7
- evidence: tankbench on every boss file, Balanced requests, 3000 iterations: paladin DTPS 147 / 182 on Magmadar / Golemagg against the warrior's 252 / 280, healing needed 31 / 83 against 185 / 232 (-43% to -90% against the mean of the others on 5 of 7 runs). Paladin and warrior sheets are nearly equal: armor 61.7% vs 60.4%, Defense 460 vs 458, 47% vs 48% of swings avoided, 45% vs 45% blocked. So the gap is in what a block or a hit leaves: Shield Specialization +15% block value a rank (sim/paladin/talents.go:47), Steadfast's block value while Holy Shield is up (sim/paladin/holy_shield.go:95), Improved Righteous Fury 3% a rank off all damage (sim/paladin/righteous_fury.go:38), Iron Creed 2% a rank for 6 sec after Holy Strike (sim/paladin/holy_strike.go:127); the warrior has no block value multiplier. Not isolated by a block-value test. The Anniversary logs in tools/encounters/*.json (healing.notes) show no paladin main tank on any boss (warriors on all, druids on four).
- proposed fix: tank_mechanics to check the size and stacking of those four effects against client 70753 and the November notes, and measure each one's share (switch each off in a scratch request).
- needs: orchestrator fix (send to tank_mechanics)
- checked: 2026-11-04 (tank-fixes-core, after the tank stands in front), not a bug. Every effect's size and stacking matches Forever's sources. On Magmadar the paladin takes 140 DTPS against the warrior's 244. Boss melee makes 86% of the gap. Swings, avoidance and block counts are the same, but a blocked hit leaves 217 against 470: block value 353 against 157. The paladin's 353 is 207 from gear × 1.6 + about 22 from Strength. Each effect switched off alone (DTPS gained, Magmadar / Golemagg), with its source:
  - Templar's Bulwark +38 / +31 (news24302099.txt:63, devnotes_post9 311 and 317).
  - Improved Righteous Fury 9% +25 / +28 (devnotes_post9:319).
  - Shield Specialization +30% block value +18 / +17 (devnotes_post9:321).
  - Steadfast +30% while Holy Shield is up +17 / +16 (client spell 1306749; client_update/diffs/70009-70170.md:1986).
  - Iron Creed 6% +14 / +16 (devnotes_post9:335).
  - Steadfast Libram's flat +30 block value +14 / +13 (forever_seen_in_game.json:79, BayovCore capture 70450).
  - Seal of Fury absorb +11 / +13 (its tooltip, 50% of the Holy damage dealt).

  The two % block-value effects add (×1.6, not ×1.69). With all of these off, the paladin takes 314 / 345, more than the warrior, whose 0.79 damage multiplier is bigger than the paladin's 0.86. Ruled out: no hit fully blocked, no reduction applied twice, the Seal of Fury shield doesn't stack, Bulwark expires.

  What's left is not a mechanic:
  - (a) The synthetic set's block value: 72 points over the warrior's, priced at the warrior's 0.4 a point while the paladin gets ×1.6 from them (+32 / +30 DTPS). This goes to tank_gear.
  - (b) Bulwark (100% of max health, 8 sec) is on a timer every pull of a 53 to 56 sec encounter, where it absorbs about 2070. A user ruling on whether a timed emergency button belongs in a scored preset.
  - (c) Confirm the Libram's +30 in game.
- needs: tank_gear (a), user ruling (b)
- paladin worker, 2026-11-04 (tank-paladin): (a) the set follows standards 36 at its own Tanks row (block value 6.7% against 6.8 hit-adjusted, 142 points with the shield's own after tank_gear-20261104-1); 36 prices block value at 0.4 for every tank and says nothing per class, so pricing it at the paladin's ×1.6 is a user ruling (docs/tank_paladin.md, Calls). (b) is in the same Calls. (c) Steadfast Libram's +30 block value rests on the in-game item capture (BayovCore ItemsSeen, client 70450, assets/db_inputs/forever_seen_in_game.json:79), the top tier for an item's stats; Wowhead lists no stats for it and no capture after 70753 exists, so a recheck in game on the current build is a call.
- pass 2 (2026-11-04, master b38ff16): still true after round 2 (tankbench 3000 iterations, Balanced): paladin healing needed 15 on Onyxia and 16 on Magmadar against about 280 to 350 for the others, -43% to -96% on all six bosses; score +15% to +73% over the others' mean. The mechanics were checked as correct; (a) block value priced at 0.4 for a ×1.6 class, (b) Bulwark on a timer in a scored preset and (c) the Libram's +30 in game are calls for the user (docs/tank_paladin.md, Calls). Stays open until the user rules.
- needs: user ruling (a), (b), (c)

### tank_fairness-20261104-5: Legacy warrior_protection.json (Orc, Warbringer gear) is still read by tests and tools
- severity: low
- affects: Protection warrior bench; healbench; encounters from_logs
- found: 2026-11-04, master f62bfc7
- evidence: tools/tankbench/req/warrior_protection.json uses real Warbringer item ids, no enchants, Orc, talents 35--251533120331221351, 1590 to 1740 TPS against 2150 to 2470 for the suffixed presets. orchestration/tank_handoff/warrior.md:63 says the Balanced request replaces it, but sim/raid_healing_test.go:35, sim/warrior/tank_warrior/forever_test.go:315, tools/encounters/from_logs.py:23 and tools/healbench/main.go:64 still read it (druid_bear.json was removed the same way). The glob tools/tankbench/req/*.json also runs it alongside the presets.
- proposed fix: point those four readers at warrior_protection_balanced.json and delete the file.
- needs: worker fix
- left open: the healing session decides warrior_protection.json's fate (brief); docs/tank_warrior.md calls it legacy (318fc7efb).
- pass 2 (2026-11-04, master b38ff16): unchanged; tools/dps_report/tankheal.py excludes it (NOT_PRESETS). Still the healing session's call.

## Resolved

### tank_fairness-20261104-4: Balanced and Max survival are the same build for the paladin and warrior; the bear's Balanced is its Max threat
- severity: medium
- affects: Protection paladin, Protection warrior, Feral bear presets
- found: 2026-11-04, master f62bfc7
- evidence: jq on tools/tankbench/req: paladin_protection_balanced.json and _max_survival.json have the same gear, talents and rotation, as do warrior_protection_balanced.json and _max_survival.json; their results match exactly on every boss. The bear's three presets differ only in rotation, and Balanced and Max threat are within about 1% on every boss. Max survival wins the Balanced weight for the bear on Golemagg and Onyxia (Onyxia 1663 vs 1476, 12%). Max threat has the highest TPS in 26 of 28 spec-boss cells (exceptions: bear and Warden with no encounter, Warden on Chromaggus).
- proposed fix: search the paladin and warrior at the Max survival weight (w = 3) for talents/rotation that differ, or rule that the weights don't separate them and ship two presets; re-search the bear's Balanced preset.
- needs: worker fix (or user ruling on dropping a preset)
- fixed (bear part): 2026-11-04, f0ff977 (Balanced re-searched; each bear preset now wins its own weight: Balanced beats Max survival by 89 ± 9 at w=1 and Max threat by 110 ± 9; no two are equal)
- warrior: 2026-11-04, 9f9c96e80: re-searched at w = 3, Balanced and Max survival share talents, race and rotation and differ only in the trinket (armor 990617 at Max survival); said in the doc and presets.ts, not forced apart.
- paladin, 2026-11-04 (tank-paladin): re-searched at w = 3 on the new engine (talents, knobs, race); Max survival's best is still Balanced's build, and its note says so. Not forced apart.
- fixed (Warden part): 2026-11-04, dd4b82a. The three Warden builds now differ, but Balanced's and Max survival's are within the paired error of each other at both weights (16.3 ± 11.1 at Balanced, 44.2 ± 32.5 at Max survival); the doc says so.
- pass 2 (2026-11-04, master b38ff16): 3×3 six-boss matrix (3000 iterations) shows every preset ahead at its own weight. The ties are the paladin's Balanced and Max survival (byte-identical files, md5 341ed8a0), warrior Max survival over Balanced at w=3 (+21.8, inside 2× err 48.6) and paladin Balanced over Max threat at w=1 (+12.5, inside 15.0). These are documented as 'the weights don't separate them', which the brief allows. The tooltips not saying so are in -9.
- resolved: 2026-11-04, b38ff16

### tank_fairness-20261104-3: Warden shaman needs 2 to 3 times the others' healing: Defense 346, 6% blocked, 15% crushed
- severity: medium (the Warden is last at every weight on every boss; whether that is intended is the question)
- affects: Warden shaman, all presets
- found: 2026-11-04, master f62bfc7
- evidence: tankbench, Balanced requests: healing needed +89% to +259% against the others on 6 of 7 runs (onyxia 646 vs 49 to 388; broodlord 1818 with 2.8% death at the encounter's own healing); no-encounter run 66% death at 600 HPS. Sheet: Defense 346 against 428 to 460, armor 54.2% against 60 to 64%, parry 6.2% and block 6.2% (15.7% each for the warrior), 15.1% of swings crush against 3 to 6% for the paladin and warrior; effective health 19.4k against 23.8k to 28.7k. Still 3x the warrior with both elixirs of finding -1 (584 vs 191 on Onyxia). docs/tank_sim.md still says the paladin and Warden have only blank sets, which no longer holds (they wear 9907xx / 9992xx).
- proposed fix: tank_gear to confirm the Warden set's Defense budget at its Tanks row is intended (standard 36), and tank_mechanics that the Warden's active mitigation is all modelled; if both hold, the user rules that the Warden ranks last. Update docs/tank_sim.md.
- needs: user ruling (after the two checks)
- ruling: 2026-11-04, user: a real class gap. The Warden is a mail class with much less defense built in than the plate tanks, and it won't necessarily be a viable raid tank. Don't compensate for it; keep its gear and mechanics checks to correctness. (standards.md, Raid duties)
- fixed: 2026-11-04, dd4b82a. Gear and mechanics checked for correctness only (docs/tank_shaman.md, Presets); hit over the cap: moving the 2.5% over 8% into crit ties at Balanced (415 against 404); docs/tank_sim.md's "no gear" line updated.
- pass 2 (2026-11-04, master b38ff16): ruled (standards 40) and the correctness checks are done; Warden healing needed +82% to +263% over the others on all six bosses, as ruled. Not re-raised.
- resolved: 2026-11-04, dd4b82a

### tank_fairness-20261104-1: Bear alone gets Phalanx and Greater Fortitude; the other tanks have no armor or health elixir
- severity: high (Balanced score +164 to +1672, healing needed -10% to -31% for the specs that lack them)
- affects: Protection warrior, Protection paladin, Warden shaman (all presets); Feral bear as the reference
- found: 2026-11-04, master f62bfc7
- evidence: `jq '.raid.parties[0].players[0].consumes' tools/tankbench/req/*_balanced.json`: druid_bear_* carry armorElixir ElixirOfThePhalanx, healthElixir ElixirOfGreaterFortitude and FlaskOfNaturalAccuracy (docs/tank_druid.md:195, the bear's own consumables audit); warrior, paladin and warden have FlaskOfTheTitans and neither elixir. Standard 11 (orchestration/standards.md:70) gives every spec the best elixir per slot. tankbench, 3000 iterations, fixed seed, Balanced requests with both elixirs added (TPS / healing needed / Balanced score vs baseline): warrior onyxia +5 / -46±4 / +247±25, broodlord +4 / -146±14 / +447±44; paladin onyxia +3 / -37±4 / +199±24, broodlord +4 / -106±12 / +327±38; warden onyxia 0 / -62±5 / +319±30, broodlord 0 / -557±65 / +1672±190. Swapping Titans for Natural Accuracy on top lowers their Balanced score (warrior -83/-155, paladin -48/-100, warden -20/-258), and the bear loses 220 with Titans, so each spec's current flask is right at the Balanced weight; only the elixirs are unequal. Side checks, both noise-level: the paladin has no weapon stone (Dense Sharpening Stone on its synthetic sword: +11 TPS), and the paladin alone has Greater Arcane Elixir (on the warden: nothing). Variant requests were in $AUDIT_WORK/tank_fairness/var/ (scratch).
- proposed fix: add ElixirOfThePhalanx and ElixirOfGreaterFortitude to the warrior, paladin and warden tank requests and their UI presets; re-pick the flask at the Max threat weight (Natural Accuracy gives +10 to +27 TPS); give the paladin a Dense Sharpening Stone for consistency.
- needs: worker fix
- fixed: 2026-11-04, 87f2880
- pass 2 (2026-11-04, master b38ff16): holds. All 12 preset requests carry ElixirOfThePhalanx and ElixirOfGreaterFortitude; the paladin has a Dense Sharpening Stone; the UI DefaultConsumes match. The Max threat flask was not re-picked: that part is finding -7.
- resolved: 2026-11-04, 87f2880
