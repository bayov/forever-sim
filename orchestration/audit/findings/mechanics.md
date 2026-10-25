# Findings: Forever mechanics

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### mechanics-20261025-1: Spellstone gives Fire spell damage too, but a server hotfix made it Shadow only
- severity: low
- affects: warlock (Affliction, Demonology, Destruction and their duty-curse builds, all on Spellstone)
- found: 2026-10-25, master 54878fa3d
- evidence: `sim/warlock/talents.go` applyWeaponImbue gives Major Spellstone 21 FirePower and 21 ShadowPower, citing the client rows' Misc 36 (Fire | Shadow). Hotfix push 112965 on 1.60.1.70532 (wago.tools/hotfixes, first seen 2026-10-23) rewrites the damage effects of 1237162 / 1237164 / 1237165 ("Increase Shadow Dam", SpellEffect 1233765 / 1233768 / 1233770) to EffectMiscValue_0 32, Shadow only; base points 14 / 17 / 21 and the haste effect are unchanged. Tier 1 (server data on the current build). Reproduce: `python3 -I tools/audit/mechanics/hotfixes.py $AUDIT_DB 70532 | grep "Increase Shadow"`. Measured with the patch below (seed 1, 20k iterations, 180 s, paired with the shared master run): Affliction 1035.28 -> 1033.78 (-1.5), Demonology 1065.09 -> 1063.73 (-1.4), Destruction 1011.77 -> 1008.98 (-2.8), duty-curse builds the same. Fire's share is Immolate (and Conflagrate for Destruction). Firestone's rows were not hotfixed. Sure of the data; the sim value is just the client row the hotfix replaced.
- proposed fix: under Forever add only ShadowPower for Spellstone (patch `$AUDIT_WORK/mechanics/spellstone.patch`): `if !warlock.Env.IsForever() { warlock.AddStat(stats.FirePower, r.damage) }`, and fix the comment's "Fire and Shadow damage". Spellstone should still beat Firestone for the Shadow builds, since Fire is only 10-20% of their damage (an estimate, not measured).
- needs: worker fix

## Resolved
