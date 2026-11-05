# Findings: Consumables, buffs and debuffs

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### consumables-20261105-1: Feral mksettings consumes lack the Goblin Sapper Charge the preset and requests now carry
- severity: low
- affects: tools/rotopt/mksettings/spec_feral.go (consumes func, lines 20-33); ui/feral_druid/presets.ts (DefaultConsumes)
- found: 2026-11-05, master a07941531
- evidence: a07941531 adds `sapperExplosive: SapperGoblinSapper` to the feral DefaultConsumes (works in form, sim/core/consumes.go explosivesUsableInForm; +3.1 at 180 sec per the preset comment, not re-measured here) and tools/dps_report/req/druid__feral-cat*.json carry it, but spec_feral.go has no SapperExplosive and its comment still says the cat "can use no explosives". A rotopt/mksettings regeneration would drop the Sapper.
- proposed fix: add `SapperExplosive: proto.SapperExplosive_SapperGoblinSapper` and fix the comment.
- needs: orchestrator fix


## Resolved

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
