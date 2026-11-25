# Findings: Consumables, buffs and debuffs

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### consumables-20261125-3: DPS Report requests still throw the EZ-Thro Dark Bomb that does nothing to a raid boss
- severity: medium (the report's hunter and feral cat numbers; presets.ts say +7.9 to +11.9 hunter, +8.7 feral at 180 sec, taken from the preset comments, not re-measured here)
- affects: tools/dps_report/req/hunter__{beast-mastery,beast-mastery-tallstrider,beast-mastery-wolf,marksmanship,marksmanship-tallstrider,marksmanship-wolf,lone-wolf}.json, druid__feral-cat.json, druid__feral-cat-duty-ff.json
- found: 2026-11-25, master 85691655a
- evidence: sim/core/consumes.go stunExplosives (new, EV-0088, user ruling 2026-11-25): Dark Bomb and Thorium Grenades are fully immune on a stun-immune boss. ui/hunter/presets.ts and spec_hunter.go moved to ExplosiveDenseDynamite and ui/feral_druid/presets.ts and spec_feral.go to ExplosiveGoblinLandMine, but `git grep fillerExplosive origin/master -- tools/dps_report/req` still shows ExplosiveEZThroDarkBomb in all nine files (line 77 or 79). Also spec_feral.go still lacks the Goblin Sapper Charge (see 20261105-1).
- proposed fix: set the hunter requests to ExplosiveDenseDynamite and the feral ones to ExplosiveGoblinLandMine, then regenerate the report.
- needs: orchestrator fix

### consumables-20261125-2: Intellect and Spirit elixirs in every preset now add nothing under the raid's Arcane Brilliance and Divine Spirit
- severity: medium (estimate, not sim-measured; the elixir slot is wasted for mana users, up to about 17 DPS on Marksmanship per the sim comment)
- affects: ui/{balance_druid,elemental_shaman,enhancement_shaman,hunter,mage,warlock,shadow_priest,retribution_paladin}/presets.ts and the matching tools/rotopt/mksettings spec files (Greater Intellect, Owl, Sages, Greater Spirit); rogue, warrior, feral carry AB and DS too
- found: 2026-11-25, master 0404484fb
- evidence: sim/core/consumes.go elixirStatsOverwritten (new, from wowsims/forever EV-0185): under Forever Arcane Intellect (Brilliance rank 5 = 31) overwrites the Intellect elixirs (Greater Intellect and Owl 25, consumes.go:1054-1055) and Divine Spirit (40) overwrites the Spirit elixirs (Sages 25, Greater Spirit 18, :1062-1063). Every Level 60 raid has arcaneBrilliance and divineSpirit true (mksettings main.go:56,60; each presets.ts). So those elixir picks give 0 Intellect/Spirit; only their crit or other effects, if any, remain. Mark of the Wild still stacks.
- proposed fix: re-pick the elixir in those slots (a stat or damage elixir that is not overwritten, e.g. Fire/Shadow/Frost Power for casters, Mongoose or Giants for melee), or rule that the raid drops AB/DS; sim paired against the current pick. Also scan.sh's "minor-armor exclusive effects (expect 2)" now reads 3 by design (Puncture Armor and Armor Shatter joined the slot); update the expected count.
- needs: orchestrator fix

### consumables-20261125-1: DPS Report assumption natural_flask_zone still describes the pre-70862 flasks (60 Stamina, 4-5%)
- severity: low
- affects: tools/dps_report/assumptions_def.py (natural_flask_zone desc, line ~108)
- found: 2026-11-25, master 5536db4bc
- evidence: the desc says "give 60 Stamina anywhere ... 5% hit, 4% crit, 5% less chance to be dodged or parried, or 5% haste". sim/core/consumes.go applyNaturalFlask (since 032692409, client 70862) gives 30 Stamina and 2% for all four. The "1.2% to 3%" cost quoted in the same desc was simmed with the old values and may be stale too (not re-measured).
- proposed fix: change the desc to 30 Stamina and 2% hit, crit, expertise or haste; re-sim the cost figure.
- needs: orchestrator fix

### consumables-20261110-1: Chest/bracer enchant 41 is named "Minor Health" but the client calls spells 7420/7418 "Inferior Stamina"
- severity: low
- affects: tools/database/forever_enchants.go (lines 77, 81); tools/database/enchant_overrides.go:97,120 still carry the old +5 Health stats for effect 41
- found: 2026-11-10, master ca5e35c3f
- evidence: ench.py on client 70862: ENCH 41 = Stamina +2 (spells 7420 chest, 7418 bracer "Inferior Stamina", items 273459/273458); ENCH 66 = Stamina +3 (7863, 7457 Minor Stamina). The new foreverKeepEffect entries match the values; only the display name differs, and the old Health 5 overrides may shadow them depending on merge order (not checked in the built db). No preset uses them.
- proposed fix: rename to "Inferior Stamina"; confirm the generated db shows +2 Stamina, not +5 Health.
- needs: orchestrator fix

### consumables-20261105-1: Feral mksettings consumes lack the Goblin Sapper Charge the preset and requests now carry
- severity: low
- affects: tools/rotopt/mksettings/spec_feral.go (consumes func, lines 20-33); ui/feral_druid/presets.ts (DefaultConsumes)
- found: 2026-11-05, master a07941531
- evidence: a07941531 adds `sapperExplosive: SapperGoblinSapper` to the feral DefaultConsumes (works in form, sim/core/consumes.go explosivesUsableInForm; +3.1 at 180 sec per the preset comment, not re-measured here) and tools/dps_report/req/druid__feral-cat*.json carry it, but spec_feral.go has no SapperExplosive and its comment still says the cat "can use no explosives". A rotopt/mksettings regeneration would drop the Sapper.
- proposed fix: add `SapperExplosive: proto.SapperExplosive_SapperGoblinSapper` and fix the comment.
- needs: orchestrator fix


## Resolved

### consumables-20261107-1: Client 70862 halves the four Natural flasks (Stamina 60 to 30, zone bonus 4-5 to 2); sim still models 60 and 4-5
- severity: high (estimate, not measured: every warrior, rogue, hunter and cat preset uses a Natural flask; the zone bonus is 5% hit/expertise/haste or 4% crit, now 2 in the client data)
- affects: sim/core/consumes.go applyNaturalFlask (lines 82-115) and sim/core/consumes_forever_test.go:266-269; presets using FlaskOfNatural*: warrior (Precision, Aggression), rogue (Precision, Accuracy), hunter (Aggression, Precision), feral_druid (Aggression); mksettings spec_warrior/rogue60/rogue_combat/hunter/feral; dps_report requests
- found: 2026-11-07, master 12aa2796f
- evidence: `builddiff.py 1.60.1.70790 1.60.1.70862`: spells 1293740-1293743 (items 274273-274276) effect 0 (A_MOD_STAT Stamina) 60 -> 30, effect 1 (the A_DUMMY zone bonus) 5/4/5/5 -> 2/2/2/2 (Accuracy/Aggression/Precision/Swiftness). Auras after it unchanged. Whether the 2 reads as 2% (hit, crit, expertise, haste) is inferred from the old 5 = 5% mapping; no tooltip checked (wowhead not queried). Not sim-measured.
- proposed fix: user/orchestrator to rule on the new values, then set Stamina 30 and the bonuses to 2 (Aggression 2% crit, Accuracy 2% hit, Precision 2 expertise, Swiftness 2% haste), update the test and flask comments in presets; re-sim and re-pick flasks (Distilled Wisdom, Titans, Supreme Power may now beat a Natural flask for some specs); the hunter/cat NO_ZONE_FLASK logic stays.
- needs: orchestrator fix
- resolved: 2026-11-11, 032692409 (applyNaturalFlask 30 Stamina, 2% bonuses, test updated)

### consumables-20261104-1: ui presets.ts still tick Salvation after mksettings and the requests dropped it
- severity: low
- affects: ui/{balance_druid,elemental_shaman,enhancement_shaman,feral_druid,hunter,mage,restoration_shaman,retribution_paladin,rogue,shadow_priest,warlock,warrior}/presets.ts (`salvation: true`); this pass doc's "Salvation on" standing check
- found: 2026-11-04, master f331f0cff
- evidence: 90e0bb01a removed Salvation from mksettings main.go and all req/*.json (Forever removed it in the 2026-10-30 hotfix; threat only, DPS identical per the commit). `git grep -n "salvation: true" origin/master -- ui/*/presets.ts` still hits 12 files, and sim/core/buffs.go:374 still models it. presets.ts and mksettings now disagree (no DPS effect; UI shows a removed buff ticked).
- proposed fix: set salvation to false/remove it in the Level 60 RaidBuffs objects in presets.ts. Optionally remove the sim aura. Pass doc standing check updated to "Salvation off".
- needs: orchestrator fix
- resolved: 2026-11-04, fb584eafa. The `salvation: true` line is gone from all 12 Level 60 presets.ts RaidBuffs objects. The sim aura (buffs.go:374) stays, unused and harmless.

### consumables-20261101-1: Ranged hunters keep the Goblin Sapper Charge, worth +0.2 now, comments say +4.8
- severity: low
- affects: ui/hunter/presets.ts (DefaultConsumes, comments), tools/rotopt/mksettings/spec_hunter.go (hunter_bm, hunter_mm, hunter_lw)
- found: 2026-11-01, master dd053d24c
- evidence: the new 10 yd reach (sim/core/consumes.go explosiveBlastRadius, commit bb21e3b32) gates the Sapper on the static DistanceFromTarget; ranged hunters stand at 25 yd and only get inside 10 yd during a melee weave. tools/dps_report/req/hunter__beast-mastery.json at 180 s: 300 iterations cast the Sapper 98 times (about a third of fights); 20000 iterations with vs without it: 1125.55 vs 1125.40 (+0.16, same seed). The comments (spec_hunter.go, presets.ts) still say Sapper +4.8 / "goes off in a weave". Dark Bomb checked: 1125.55 vs 1122.42 with none (+3.1), matches the claimed +2.8 to +3.9.
- proposed fix: drop SapperExplosive from the ranged hunter consumes (or let the weave be planned so it fires), fix the +4.8 comment. Survival (5 yd) is unaffected.
- needs: orchestrator fix
- resolved: 2026-11-02, b33ad307d. The ranged hunters (Beast Mastery, Marksmanship, Lone Wolf and their pet variants) take no Sapper and keep the EZ-Thro Dark Bomb; Survival keeps it. Beast Mastery 1046.9 -> 1046.5, Marksmanship 1066.4 -> 1066.3, Lone Wolf 1090.0 -> 1089.9 at 180 sec; the +4.8 / "goes off in a weave" comments are fixed.

### consumables-20261030-1: Elixir of Wisdom gives +3 Intellect in client 70707, sim has 6
- severity: low
- affects: sim/core/consumes.go (intellectElixirs, ElixirOfWisdom); no Level 60 preset uses it
- found: 2026-10-30, master 36a772fd1
- evidence: builddiff.py 70685 -> 70707: spell 3166 (item 3383 Elixir of Wisdom) Aura 29 base points 6 -> 3 (misc 3 = Intellect). consumes.go:956 has 6. Same diff: Well Fed 1249521 changed to a different aura (30/356) and Fengus' Ferocity 22817 and Missile Barrage 400589 changed; none used by a preset or the sim (git grep). Unconfirmed whether ElixirOfLesserIntellect (also 6) maps to the same spell.
- proposed fix: set ElixirOfWisdom to 3 (check Lesser Intellect against its own spell with q.py)
- needs: orchestrator fix
- resolved: 2026-10-30, sim/core/consumes.go ElixirOfWisdom 6 -> 3 (Lesser Intellect is its own spell 1250988, still 6 in 70707)
