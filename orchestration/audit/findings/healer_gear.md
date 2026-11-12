# Findings: healer_gear

## Open

### healer_gear-20261112-1: all five healer presets waste the spell-damage elixir and food; Greater Cleric's Elixir and Sage's Tea heal
- severity: high (+1.0 to +2.0% damage carried on 8 of the 15 priest, paladin, shaman and druid cells measured, 0 on the rest)
- affects: Restoration druid, Restoration shaman, Holy paladin, Holy priest, Discipline priest (consumes in ui/*/presets.ts DefaultConsumes and tools/healbench/req/*.json)
- found: 2026-11-12, master 55cd92b09
- evidence: the presets carry Greater Arcane Elixir (spellPowerBuff) and Lobster Stew; sim/core/consumes.go:917 and :648 add stats.SpellDamage, which Spell.HealingPower (sim/core/spell_result.go:303) does not read (SpellPower + HealingPower only). Removing the Arcane Elixir changes nothing (carry identical, all five, chromaggus). Greater Cleric's Elixir (healingPowerBuff, +40 healing, consumes.go:949) and Sage's Tea (+44 healing, consumes.go:670) are in the sim and the UI (consumables.ts:443, :1049) and unused. Healbench -role tank -carry, 2000 iterations, paired seeds, base vs swap (Cleric's in place of Arcane, Sage's Tea in place of Lobster Stew): chromaggus carry druid 1.051 to 1.047, shaman 0.965 to 0.977 (+1.2%), paladin 0.915 to 0.929 (+1.4%), Holy priest 0.953 to 0.970 (+1.7%), Discipline 0.873 to 0.893 (+2.0%); golemagg druid 1.953 to 1.988 (+1.8%), shaman 1.916 to 1.950 (+1.8%), paladin +0.6%, Holy +1.1%, Disc +0.1%; onyxia shaman +1.6%, the rest within noise (about 0.01 carry). Files: AUDIT_WORK/healer_gear/sweep2.txt, sweep.py. Recipes: Greater Cleric's needs Alchemy 195 and a faction vendor (forever_wowhead_recipes.json id 250369), Sage's Tea recipe (249884, skill 250) lists no source in the cached page, so its obtainability is unverified. Whether the Cleric's elixir shares a slot with Mageblood or Sages in game is not modelled (no slot limit in the sim); the swap uses the Arcane Elixir's slot only.
- proposed fix: healingPowerBuff GreaterClericsElixir and drop spellPowerBuff in the five presets and healbench requests; food FoodSagesTea if its recipe is obtainable at launch (otherwise the best obtainable healing food); then re-run the healer bench.
- needs: orchestrator fix

### healer_gear-20261112-2: Brilliant Wizard Oil heals more than Brilliant Mana Oil on the fights with spare mana
- severity: medium (+0.5 to +2.5% carry on golemagg and onyxia for shaman, Holy, Discipline; mana-starved chromaggus 0 to +1%; within noise on some cells)
- affects: all five healer presets (mainHandImbue BrilliantManaOil)
- found: 2026-11-12, master 55cd92b09
- evidence: Forever's wizard oils add healing equal to their spell power (consumes.go addForeverOilHealing; Brilliant 36 spell power + 36 healing + 1% crit = 72 healing) against Brilliant Mana Oil's 30 healing + 15 MP5. Healbench tank carry, 2000 iterations, swap vs swapBW: golemagg druid +0.4%, shaman 0, paladin +1.7%, Holy +0.9%, Disc +2.0%; onyxia druid +0.7%, shaman +0.8%, paladin 0, Holy +0.9%, Disc -0.3%; chromaggus druid +1.0%, shaman -0.3%, paladin +0.1%, Holy 0, Disc -0.9%. Mixed and near the noise (about 0.01 carry); not a clear win per cell, positive on average (+0.5%).
- proposed fix: re-measure with more seeds and the raid role before changing; if it holds, BrilliantWizardOil.
- needs: orchestrator fix

### healer_gear-20261112-3: Flask of Distilled Wisdom beats Supreme Power for the priests on mana-limited fights
- severity: medium (Holy +5.8%, Discipline +8.7% carry on Chromaggus; -0.5 to -2.4% on Onyxia and Golemagg)
- affects: Holy priest, Discipline priest (druid, shaman, paladin lose 0.3 to 3% with it everywhere except shaman Chromaggus +0.8%)
- found: 2026-11-12, master 55cd92b09
- evidence: Flask of Distilled Wisdom (+2000 mana, consumes.go:62) against Supreme Power (+150 spell power). Healbench tank carry, 2000 iterations, swapDW vs swap: Chromaggus (115 s, priests end with 830 and 406 mana, Holy out of mana in the 128 s) Holy 0.970 to 1.035, Discipline 0.893 to 0.971; Golemagg Holy 2.006 to 1.977, Disc 1.995 to 1.964; Onyxia Holy 1.378 to 1.382, Disc 1.363 to 1.342. The flask is one per preset while the best depends on the fight length, so a ruling is needed on how presets pick (per boss, or the mean over the seven bosses and both roles, not measured here: the raid role is slow on the shared machine).
- proposed fix: bench every boss and both roles with each flask and pick by the mean carry, as the objective for a healer preset; record the answer for the priests.
- needs: user ruling


## Resolved
