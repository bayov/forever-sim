# Findings: tank_gear

## Open

### tank_gear-20261112-1: Warrior (all presets) and Warden (Max threat, Balanced) still wear Greater Stamina boots; Greater Agility (1887) scores higher
- severity: low (+3 to +10 on scores of 1,100 to 3,200, under 0.4%)
- affects: Protection warrior Max threat / Balanced / Max survival; Warden shaman Max threat / Balanced (paladin and bear already wear 1887 and 929 loses there)
- found: 2026-11-12, master 55cd92b09
- evidence: paired `abtest` (tools/audit/tank_rotations/abtest, built `-tags=with_db`), 10,000 iterations, six-boss scoring mean at the preset's weight, boots 929 -> 1887 on the request (tools/tankbench/req/*.json, item 990610 / 999210), seed 1 / seed 2: warrior Max threat +5.6 +/- 1.0 / +3.0 +/- 1.2, Balanced +4.5 +/- 1.8 / +3.6 +/- 1.9, Max survival +4.1 +/- 3.7 / +6.8 +/- 3.9; Warden Max threat +5.4 +/- 1.1 / +5.0 +/- 1.2, Balanced +9.0 +/- 2.1 / +10.0 +/- 2.1. Warden Max survival keeps 929 (-8.8 +/- 10.3 at 3,000). 1887 is live for boots in db.json and is the paladin/bear pick. docs/tank_warrior.md has no boots A/B (only "the boots keep it", line 370); docs/tank_shaman.md's boots A/B (line 259) was against vendor kits only, in tank round 3, before the Nov 10 changes.
- proposed fix: boots 1887 on the warrior (all three requests and gear sets) and the Warden Max threat and Balanced sets; record in the class docs.
- needs: worker fix

### tank_gear-20261112-2: Warrior Max threat (and probably Balanced) wrists: Superior Strength (1885) beats Superior Deflection (1248665)
- severity: low
- affects: Protection warrior Max threat, Balanced
- found: 2026-11-12, master 55cd92b09
- evidence: same method, 1248665 -> 1885 on wrist 990606-line item: Max threat +8.4 +/- 1.6 / +11.8 +/- 1.7 (10,000, seeds 1/2); Balanced +8.2 +/- 2.6 / +2.2 +/- 2.9 (10,000), +13.0 +/- 4.5 at 3,000; Max survival -20.5 +/- 9.6 (3,000), keeps 1248665. docs/tank_warrior.md says Superior Deflection stays at every weight (round 3, 2026-11-04) and records no Strength A/B for the warrior; the bear wears 1885 at Max threat for the same reason.
- proposed fix: Max threat wrist 1885; Balanced marginal (mean +5.2, about 2 sigma), re-run at more seeds before switching.
- needs: worker fix

### tank_gear-20261112-3: Warden Max threat: cloak Agility (1219587) and wrist Superior Deflection (1248665) now beat Superior Defense / Superior Stamina; Balanced cloak too
- severity: low
- affects: Warden shaman Max threat, Balanced
- found: 2026-11-12, master 55cd92b09
- evidence: after the Nov 12 retune (new Max threat talents, Titans at Balanced), paired 10,000, seeds 1/2: Max threat cloak 1889 -> 1219587 +3.9 +/- 1.0 / +3.3 +/- 1.1; Balanced cloak +4.2 +/- 1.8 / +4.2 +/- 2.0 (3,000 screen -10.1 +/- 8.9 at Max survival, keeps 1889). Max threat wrist 1886 -> 1248665 +4.4 +/- 1.5 / +5.8 +/- 1.5 (docs/tank_shaman.md line ~200: 1886 kept at -2.1 +/- 6.7 on the old build). Max survival unchanged (wrist/cloak/shield all keep). Shield 1886 still beats 863 at all three (-4.9 / -10.4 / -32.8).
- proposed fix: Max threat: cloak 1219587, wrist 1248665; Balanced: cloak 1219587. Combined effect not simmed (gains probably add, about +8 to +10 at Max threat).
- needs: worker fix

## Resolved

### tank_gear-20261104-8: DPS Report names enchants by effect id alone, so shared ids show another slot's enchant
- severity: low
- affects: DPS Report gear tables: Tanking (all wrists with 1886, paladin and Warden shields with 929 in the current data) and the DPS mode (4 Hands rows)
- found: 2026-11-04, master dc303b0
- evidence: tools/dps_report/build_data.py:62 builds `enchant_names = {e['effectId']: e['name'] ...}`, so the last db.json entry for an effect id wins; tools/dps_report/build_data.py:516 and tools/dps_report/tankheal.py:351 look names up by id only. db.json has 1886 as Bracer Superior Stamina and Shield Greater Stamina, 929 as Bracer and Boots Greater Stamina, 1887 as Gloves and Boots Greater Agility. ui/dps_report/data/tank_gear.json: 9 Wrist rows read "Enchant Shield - Greater Stamina", 6 Off hand rows "Enchant Boots - Greater Stamina"; data/gear.json: 4 Hands rows "Enchant Boots - Greater Agility". Labels only; the sims use the right stats. After the round-3 regen the paladin Max threat boots (1887) and the Warden Max threat wrist and every shield 1886 still hit it.
- proposed fix: pick the name by (effectId, slot): among the db.json enchants with that effect id, take the one whose type (or enchantType for shields and 2H) matches the item's slot, as the UI's lookupItemSpec does (ui/core/proto_utils/database.ts).
- needs: orchestrator fix
- fixed: 2026-11-04, 9883bd54f. build_data.py's enchant_for / enchant_name pick the db.json enchant by effect id and the slot's item type (a shield enchant on a weapon slot only for a shield); the gear rows (build_data.py, tankheal.py) and the enchant tooltips (tooltips.py, now one per enchant record) use it. After the regen no Wrist, Hands or Feet row names another slot's enchant.

### tank_gear-20261104-7: Cloak enchant differs across tanks with no A/B: three wear Agility +5 where the Warden wears Superior Defense
- severity: low
- affects: Protection warrior, Protection paladin, Feral bear (Warden for comparison)
- found: 2026-11-04, master b38ff16
- evidence: the warrior, paladin and bear backs (990604, 990704, 990804) take 1219587 Enchant Cloak - Agility (+5 Agility, 5 weighted at 1b/36 prices); the Warden's (999204) takes 1889 Superior Defense (+70 armor, 7 weighted at 36's 0.1 a point of bonus armor). Both are outside the raids. None of docs/tank_{warrior,paladin,druid,shaman}.md records the comparison. For the bear, armor from enchants may also take Dire Bear Form's multiplier, which would favour 1889 further; Agility's dodge, crit and threat may favour 1219587 at Max threat. Not simmed; the difference is about 2 weighted points, under 0.3% of a set (estimate). Confidence about 60% that one of the three would switch at Max survival.
- proposed fix: A/B 1889 against 1219587 on the back for each of the three at each preset's weight; keep the winner and record it in the class doc.
- needs: worker fix
- fixed (Warden part): 2026-11-04, 6a9dea8 (docs 6c73945). Agility +5 (1219587) against Superior Defense (1889) on the Warden's back: +1.8 ± 3.0 / -1.8 ± 4.5 / -7.2 ± 15.3 on the final builds (seeds 1 and 2 alike); 1889 stays at all three.
- bear part: fixed: 2026-11-04, ad93b95. Superior Defense 1889 against Agility 1219587 on the back: -4.6 ± 1.1 / -9.3 ± 3.0 / -1.8 ± 8.8 at Max threat / Balanced / Max survival (seed 2: -6.9 ± 1.2 / -10.8 ± 3.5 / +5.2 ± 9.6), so Agility stays at every weight; recorded in docs/tank_druid.md, Gear.
- fixed (paladin): 2026-11-04, b07bdf476: Superior Defense 1889 against Agility 1219587 at 1/3, 1, 3 (seed 1 / 2): -3.1 / -4.7, -2.1 / +0.2, -0.2 / +8.6; on the final wrist and shield -4.3 / -6.2 / -14.8. Agility stays; recorded in docs/tank_paladin.md (85db121aa).
- fixed (warrior part): 2026-11-04, a624264: paired 10k at each weight, 1889 minus 1219587: Max threat -6.3 / -6.0 ± 1.0 (seeds 1/2), Balanced -5.6 / +0.5 / -3.4 / -5.1 ± 1.7 (seeds 1-4), Max survival -10.1 / +1.2 / +0.1 / -1.3 ± 4.4 (mean -2.5 ± 2.2). Agility stays; recorded in docs/tank_warrior.md (Round 3 gear A/Bs). Paladin and bear parts are their workers'.
- resolved: 2026-11-04 (pass 3), dc303b0. All four backs A/B'd and recorded (warrior a624264, paladin b07bdf476, bear ad93b95, Warden 6a9dea8): Agility 1219587 stays on the warrior, paladin and bear, Superior Defense 1889 on the Warden, as every gear set and tankbench request shows.

### tank_gear-20261104-3: Tank enchants use stale Classic values; shield Greater Stamina and the Deflection bracer enchant are underrated
- severity: low
- affects: all four tank sets (shield: warrior, paladin, Warden; wrist: all four)
- found: 2026-11-04, master f62bfc74b
- evidence:
  - Shield: the sets use enchant 929 (Greater Stamina +7). The vendor formula 16217 (skill 265) gives +9 per forever_wowhead_recipes.json and client spell 20017. That is enchant 1886, the id the wrists already use. tools/database/enchant_overrides.go:221 still reads +7.
  - Wrist: Bracer Deflection (spell 13931) is Defense +7 in the client, but enchant_overrides.go:134 has Defense 3, so the sim can't value it. Superior Deflection (Defense +9, spell 1248665, a vendor formula) would be 13.5 weighted against Superior Stamina's 9, and nobody has evaluated it. Its formula-to-spell mapping is unverified, and Forever's currency-3402 Deflection formulas (249539, 249505) may not be Phase 1.
  - Chest Major Health is Major Stamina +10 in the client, which doesn't change the Greater Stats pick.
- proposed fix: apply the client stats to the stale enchant_overrides.go entries. Move the shields to 1886 (+2 Stamina). Sim Superior Deflection against Superior Stamina on the wrist for each tank, and keep the winner.
- needs: worker fix
- fixed: 2026-11-04, d10eaea1f (shared values: shield 1886 +9, Bracer Deflection +7, Superior Deflection in the DB) and 5d9a3aa18 (warrior wears Superior Deflection: +4.3 / +6.6 / +25.8 at 1/3, 1, 3). Paladin and Warden sets still put 929 (+7, still right for 929's other spells) on the shield: their workers move them to 1886. No set wore Bracer Deflection.
- reopened: 2026-11-04 (pass 2), master b38ff16. Only the warrior moved. (a) The paladin and Warden shields (990714, 999214) still carry 929 in both the UI gear sets (ui/protection_paladin/gear_sets/level60_phase1_synthetic.gear.json, ui/warden_shaman/gear_sets/phase1_synthetic_tank*.gear.json) and the tankbench requests (paladin_protection_*.json, shaman_warden_*.json). 929 is no longer a shield enchant in the DB (assets/database/db.json has 929 only as Bracer and Boots Greater Stamina; the shield entry is 1886). So the CLI and tankbench give the shield +7 Stamina through EnchantsByEffectID (sim/core/database.go:269, no slot check), the warrior gets +9, and the web UI's lookupItemSpec (ui/core/proto_utils/database.ts:184-190) searches only the off-hand slot's enchants, finds no 929 and loads the paladin and Warden shields with no enchant at all. (b) Superior Deflection (1248665, Defense +9, 13.5 weighted against Superior Stamina's 9) was simmed only for the warrior (+4.3 / +6.6 / +25.8 at 1/3, 1, 3). The paladin, bear and Warden wrists still wear 1886 Superior Stamina with no A/B recorded (no mention in docs/tank_{paladin,druid,shaman}.md). Not simmed here; the warrior's gain suggests a few points at Balanced and more at Max survival.
- proposed fix (pass 2): shields 929 -> 1886 in the paladin and Warden gear sets and requests; A/B 1248665 against 1886 on the wrist for the paladin, bear and Warden at each weight and keep the winner.
- fixed (Warden part): 2026-11-04, 71a107e (shield 999214: 929 -> 1886 in the three requests and both UI sets) and 6a9dea8 (A/Bs, docs/tank_shaman.md 6c73945: shield Lesser Block 863 against 1886 loses -8.6 ± 3.4 / -12.6 ± 4.9 / -40.6 ± 12.7; wrist Superior Deflection 1248665 taken at Balanced and Max survival on pooled gains, e.g. +15.8 ± 4.3 and +26.1 ± 12.2 at 10k, Max threat keeps 1886 at -2.1 ± 6.7; call (h)).
- bear part (a): 2026-11-04, 0b901ff: no shield, so no 929; every bear enchant id is live for its slot (docs/tank_druid.md, Gear). Part (b), the wrist 1248665 against 1886 A/B, waits for the round-3 score phase.
- fixed (warrior part, round 3): 2026-11-04, a624264: the warrior shield is 1886 in all three requests and both gear sets; wrist re-A/B'd on the new engine, 1886 minus 1248665 at 10k: -4.8 / -3.9 (Max threat), -10.4 / -0.8 / +0.3 / -7.3 (Balanced), -31.6 / -13.0 / -11.5 / -32.7 (Max survival); Superior Deflection stays. Paladin and Warden parts are their workers'.
- fixed (paladin): 2026-11-04, e37867b (shield 929 -> 1886 in the UI set and requests) and b07bdf476 (A/B at 1/3, 1, 3, paired, seeds 1 and 2: wrist Superior Deflection 1248665 over 1886 +10.5 / +20.8 / +64.2, taken; shield Lesser Block 863 over 1886 +24.6 / +35.3 / +68.8, taken); docs/tank_paladin.md round 3 (85db121aa).
- bear part (b): fixed: 2026-11-04, ad93b95. Wrist A/B at each weight (3,000, seeds 1 and 2; 10,000 on seeds 3 and 4 for the close ones): Superior Deflection 1248665 wins at Balanced (+6.1 ± 2.7 at 10k) and Max survival (+23.9 ± 7.3 at 10k) and is worn there; at Max threat Superior Strength 1885 wins (+15.1 ± 0.3, seed 2), Deflection +5.2 ± 1.9. docs/tank_druid.md, Gear.
- resolved: 2026-11-04 (pass 3), dc303b0. No shield carries 929 in any gear set or request: warrior and Warden 1886, paladin 863 Lesser Block (live shield enchant, +2% block, world-drop formula 11168; beats 1886 for the paladin, loses for the warrior and Warden, all recorded). Wrists A/B'd for all four and worn as the docs say: warrior and paladin 1248665 at every preset; bear 1248665 at Balanced and Max survival, 1885 Superior Strength at Max threat; Warden 1248665 at Balanced and Max survival, 1886 at Max threat. Every enchant id is live for its slot in db.json, so the web UI loads them.

### tank_gear-20261104-1: Protection paladin set's Stamina share 30.2% is outside the row's 32.4 ±1.5
- severity: medium
- affects: Protection paladin (level60_phase1_synthetic)
- found: 2026-11-04, master f62bfc74b
- evidence: `python3 tools/audit/tank_gear/paladin.py`. Linear shares without trinkets: Sta 30.2 against 32.4 (tolerance 1.5), so it is 0.7 points out. The hit raise from 3.2 to 5.9 (50 rating plus Precision 3/3 = 8%, standards 4) was paid partly from ring and main-hand Stamina (tools/database/forever_synthetic_items_paladin_tank.go:22-38). gear_targets.md's hit-adjusted rule and standards 29 take extra hit from the other non-Stamina, non-Spirit stats, not from Stamina. The file comment prints "30.2 (32.4)" without flagging it. Every piece is within 1% linear and 6% norm. Confidence about 85%.
- proposed fix: move about 19 weighted points (6 at minimum) into Stamina from Defense (+0.9 over its hit-adjusted target) and block value (+0.4 over), for example on shoulder (BV 37) and wrist (BV 25). Keep each piece in budget. The TPS or survival effect is probably under 1% (estimated, not simmed).
- needs: worker fix
- fixed: 2026-11-04, e9751cc. Stamina 30.2% -> 32.6% (32.4 ±1.5): head, chest, legs Defense -2, shoulder block value -10, ring and hands Strength -2 / -3, back Agility -2, all into Stamina; every piece within 1% linear and 5% norm (`python3 tools/audit/tank_gear/paladin.py`).
- resolved: 2026-11-04 (pass 2), e9751cc. Recomputed: Stamina 32.6 against 32.4 ±1.5; every piece within 1% linear and 5% norm.

### tank_gear-20261104-2: Protection warrior wears the DPS on-use AP trinket with no recorded A/B against the armor trinket
- severity: medium
- affects: Protection warrior
- found: 2026-11-04, master f62bfc74b
- evidence: ui/tank_warrior/gear_sets/level60_phase1_synthetic.gear.json:14 wears 991018 (+335 AP on use), not the defensive 990617 (1,665 bonus armor). Standards 36 says to take whichever scores higher on the tank objective. The Go comment (forever_synthetic_items_warrior_tank.go:48-51) points to "docs/tank_warrior.md, Gear", but that doc has no such comparison, and orchestration/tank_handoff/warrior has only a trinketAt search. By contrast, the paladin's defensive alternative was simmed lower, and the bear's choice is explained in presets.ts:49-52. Confidence about 75% that no A/B exists anywhere.
- proposed fix: run rotopt -eval on the six tank encounters with 991018 swapped for 990617, keep the winner, and record the result in docs/tank_warrior.md.
- needs: worker fix
- fixed: 2026-11-04, 5d9a3aa18: A/B at each weight: 990617 loses at Max threat (-209) and Balanced (-176), wins at Max survival held for the last 35 s (+19.0 ± 4.8, 10k, all from Broodlord); recorded in docs/tank_warrior.md.
- resolved: 2026-11-04 (pass 2), 5d9a3aa18. The A/B is in docs/tank_warrior.md; Max survival wears 990617 (level60_phase1_synthetic_armor_trinket, tankbench req max_survival), Max threat and Balanced 991018, matching it.

### tank_gear-20261104-4: Feral bear chest is 1.3% under its linear budget and the set's Stamina share 0.2 over tolerance
- severity: low
- affects: Feral bear tank
- found: 2026-11-04, master f62bfc74b
- evidence: `python3 tools/audit/tank_gear/druid.py`.
  - Chest 990805 (forever_synthetic_items_druid_tank.go:~108-111) is linear 71.5 against 72.44 (-1.3%). The header (lines 15-20) compares against a rounded 72 and prints the norm's -1.6% as if it were linear.
  - Set shares: Sta 35.1 against 33.4 ±1.5, Agi 22.5 against 19.7 (+2.8, inside 3). The dropped Int, crit, AP and SP shares (4.8 points) went partly to Stamina.
  - The stat trinket 990814 (23 Def + 1% dodge) is +1.4% linear. No whole Defense and dodge pair fits the band, so I take that as rounding and not a defect.
- proposed fix: chest Stamina 24 to 25, then re-check the norm. Move about 2 to 3 points of Stamina into Defense or Strength.
- needs: worker fix
- fixed: 2026-11-04, bde6912 (chest Stamina 25: linear 72.5 / 72.44; 3 Stamina to Strength: Stamina share 34.8 against 33.4 ±1.5)
- resolved: 2026-11-04 (pass 2), bde6912. Recomputed: chest 72.5 / 72.44 (+0.1%), Stamina 34.8 against 33.4 ±1.5.

### tank_gear-20261104-5: Stale comments in the tank gear files
- severity: low
- affects: Feral bear, Protection warrior (documentation only)
- found: 2026-11-04, master f62bfc74b
- evidence:
  - Bear: the header of forever_synthetic_items_druid_tank.go (lines 35-44) and sim/druid/forever_synthetic_trinkets.go say the bear wears the armor on-use trinket 990815. The preset (gear JSON, presets.ts:49) wears the Strength on-use 998115. 990815 is used only in bear_rules_test.go:893.
  - Warrior: the Go header's numbers are stale. It gives armor 650.1 (should be 649.6), set 760.5 (760.0), Defense 106 (107), crit 6.6% (6.4%) and trinket norm 38.4 (38.3).
- proposed fix: correct the comments to match the code.
- needs: worker fix
- fixed (bear part): 2026-11-04, bde6912 (both comments name 998115 as worn, 990815 as the search's alternative)
- fixed (warrior part): 2026-11-04, 1593a8278. The bear part is the druid worker's.
- resolved: 2026-11-04 (pass 2), bde6912 and 1593a8278. Warrior header now reads armor 649.6, set 760.0, matching the recompute.

### tank_gear-20261104-6: Warden boots take Greater Stamina where Mystic Rugged Armor Kit is worth more
- severity: low
- affects: Warden shaman (both sets)
- found: 2026-11-04, master f62bfc74b
- evidence: phase1_synthetic_tank*.gear.json: the feet (999210) take enchant 929, Stamina +7 (7 weighted). Mystic Rugged Armor Kit (1254966) gives 6 SP and 40 armor, about 9.2 weighted, and Forceful Rugged (1254965) about 9.0. Both are vendor patterns (252941/252942, skill 235). That is about +0.2% of the weighted budget. This is not simmed, and the tank objective may still prefer Stamina.
- proposed fix: sim Mystic Rugged against Greater Stamina on the feet, and keep the winner.
- needs: worker fix
- fixed: 2026-11-04, dd4b82a. Simmed (paired, 3,000): Mystic Rugged +2.7 ± 0.5 / -6.6 ± 2.7 / -26.1 ± 6.7 (Max threat / Balanced / Max survival), Forceful Rugged +1.6 / -7.7 / -27.2; Greater Stamina stays (Max threat shares Balanced's set; call (e) in docs/tank_shaman.md).
- resolved: 2026-11-04 (pass 2), dd4b82a (simmed; Greater Stamina kept, call (e) in docs/tank_shaman.md).
