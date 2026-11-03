# Consumables pass 67 (master f5c99ad018)

Trigger: 20 watched files moved since 9da90fcde (forced, round 70753 phase 3). Client 1.60.1.70753, no newer 1.60 build on wago.

## Verified
- sim/core/buffs.go: Trueshot Aura ranks 20 / 35 / 50 / 75 / 100 (rank 5 = 100). Matches the Nov 2 notes line 172 (ref/client_update/sources/devnotes_post9_2026-11-02.txt) and hotfix 113534. Forever-only code path, cited in the comment, listed in round_70753/hunter.md. Not a finding (standards 37).
- sim/core/consumes.go: Dragonfire Potion and Caustic Smog Potion no longer crit (Attributes[2] 0x20000000, SpellFlagNoPeriodicCrit, OutcomeAlwaysHit). Comments cite the flag. No preset depends on them for a ranking.
- sim/core/debuffs.go: new ForeverDemoralizingRoarAura (bear helper, druid). No preset or mksettings raid change.
- assumptions_def.py: druid_rip_31plus row removed (not buff related).
- presets.ts (all DPS specs): diffs are talents, rotations, races, comments and the dropped rogue "(vs Humanoid)" builds. No Consumes.create, RaidBuffs, Debuffs or enchant line changed. mksettings did not move.
- New warlock default gear (Phase1 Synthetic) and Affliction Wrack preset: scan clean (trinkets synthetic, no banned enchants).
- Resto shaman gear set added: out of scope.
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty. Counts: minor-armor exclusive effects 3 (known, Dust Cloud; ref expects 2); mksettings default MobTypeUnknown; Salvation still on (planned drop); raid totems unchanged.
- Class branches show as "ahead" only through stale copies (gotcha 1).

## Open
- None.
