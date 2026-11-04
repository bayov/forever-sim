# Findings: tank_gear

## Open

### tank_gear-20261104-1: Protection paladin set's Stamina share 30.2% is outside the row's 32.4 ±1.5
- severity: medium
- affects: Protection paladin (level60_phase1_synthetic)
- found: 2026-11-04, master f62bfc74b
- evidence: `python3 tools/audit/tank_gear/paladin.py`. Linear shares without trinkets: Sta 30.2 against 32.4 (tolerance 1.5), so it is 0.7 points out. The hit raise from 3.2 to 5.9 (50 rating plus Precision 3/3 = 8%, standards 4) was paid partly from ring and main-hand Stamina (tools/database/forever_synthetic_items_paladin_tank.go:22-38). gear_targets.md's hit-adjusted rule and standards 29 take extra hit from the other non-Stamina, non-Spirit stats, not from Stamina. The file comment prints "30.2 (32.4)" without flagging it. Every piece is within 1% linear and 6% norm. Confidence about 85%.
- proposed fix: move about 19 weighted points (6 at minimum) into Stamina from Defense (+0.9 over its hit-adjusted target) and block value (+0.4 over), for example on shoulder (BV 37) and wrist (BV 25). Keep each piece in budget. The TPS or survival effect is probably under 1% (estimated, not simmed).
- needs: worker fix
- fixed: 2026-11-04, e9751cc. Stamina 30.2% -> 32.6% (32.4 ±1.5): head, chest, legs Defense -2, shoulder block value -10, ring and hands Strength -2 / -3, back Agility -2, all into Stamina; every piece within 1% linear and 5% norm (`python3 tools/audit/tank_gear/paladin.py`).

### tank_gear-20261104-2: Protection warrior wears the DPS on-use AP trinket with no recorded A/B against the armor trinket
- severity: medium
- affects: Protection warrior
- found: 2026-11-04, master f62bfc74b
- evidence: ui/tank_warrior/gear_sets/level60_phase1_synthetic.gear.json:14 wears 991018 (+335 AP on use), not the defensive 990617 (1,665 bonus armor). Standards 36 says to take whichever scores higher on the tank objective. The Go comment (forever_synthetic_items_warrior_tank.go:48-51) points to "docs/tank_warrior.md, Gear", but that doc has no such comparison, and orchestration/tank_handoff/warrior has only a trinketAt search. By contrast, the paladin's defensive alternative was simmed lower, and the bear's choice is explained in presets.ts:49-52. Confidence about 75% that no A/B exists anywhere.
- proposed fix: run rotopt -eval on the six tank encounters with 991018 swapped for 990617, keep the winner, and record the result in docs/tank_warrior.md.
- needs: worker fix
- fixed: 2026-11-04, 5d9a3aa18: A/B at each weight: 990617 loses at Max threat (-209) and Balanced (-176), wins at Max survival held for the last 35 s (+19.0 ± 4.8, 10k, all from Broodlord); recorded in docs/tank_warrior.md.

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

### tank_gear-20261104-5: Stale comments in the tank gear files
- severity: low
- affects: Feral bear, Protection warrior (documentation only)
- found: 2026-11-04, master f62bfc74b
- evidence:
  - Bear: the header of forever_synthetic_items_druid_tank.go (lines 35-44) and sim/druid/forever_synthetic_trinkets.go say the bear wears the armor on-use trinket 990815. The preset (gear JSON, presets.ts:49) wears the Strength on-use 998115. 990815 is used only in bear_rules_test.go:893.
  - Warrior: the Go header's numbers are stale. It gives armor 650.1 (should be 649.6), set 760.5 (760.0), Defense 106 (107), crit 6.6% (6.4%) and trinket norm 38.4 (38.3).
- proposed fix: correct the comments to match the code.
- needs: worker fix
- fixed (warrior part): 2026-11-04, 1593a8278. The bear part is the druid worker's.

### tank_gear-20261104-6: Warden boots take Greater Stamina where Mystic Rugged Armor Kit is worth more
- severity: low
- affects: Warden shaman (both sets)
- found: 2026-11-04, master f62bfc74b
- evidence: phase1_synthetic_tank*.gear.json: the feet (999210) take enchant 929, Stamina +7 (7 weighted). Mystic Rugged Armor Kit (1254966) gives 6 SP and 40 armor, about 9.2 weighted, and Forceful Rugged (1254965) about 9.0. Both are vendor patterns (252941/252942, skill 235). That is about +0.2% of the weighted budget. This is not simmed, and the tank objective may still prefer Stamina.
- proposed fix: sim Mystic Rugged against Greater Stamina on the feet, and keep the winner.
- needs: worker fix

## Resolved
