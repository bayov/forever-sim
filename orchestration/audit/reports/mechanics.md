# Mechanics pass 77 (2026-11-12, master 8306f1d52)

Quiet: no new findings. Open findings mechanics-20261112-1/-2/-3 and 20261104-1 stay open (PR #128, the
Judgement / Unsealed Wrath / Holy Strike work, is not on master yet).

## Range
a0b4d3be98..8306f1d52: three sim commits, one TASKS.md note (8306f1d52). No forever-update merges.

| commit | what | verdict |
|---|---|---|
| 368b2e181 | Hunter pet swings at owner's melee and attack speed (not ranged); autocast passes an ability short of focus | User's go 2026-11-12, from wowsims/forever d729b13992 / 8223020463 log reads. Core change checked: `Pet.FollowsOwnerMeleeSpeed` multiplies the pet by the owner's MeleeSpeedMultiplier at Enable and by each later `MultiplyMeleeSpeed` / `MultiplyAttackSpeed`; Disable removes the pet from `DynamicMeleeSpeedPets` and divides back `inheritedMeleeSpeed`. `MultiplyRangedSpeed` does not reach it (Quick Shots excluded, as the log says). Only `sim/hunter/pet.go:225` sets the flag, so no other pet moves. Owner haste rating is not inherited (no Level 60 hunter gear has it). Correct. |
| c380bbaef | Rake's direct hit adds 0.065 AP under Forever, outside the Nov 2 scale | User ruling 2026-11-12. Client 70888 `sp.py 1822`: eff 2 bp 27 apcoef 0, eff 6 bp 20 per 3000 apcoef 0, so the AP share is server-side like the tick's 0.055; the value rests on the 70862 log fit (0.0648, 26/26 hits). Not a finding. |
| f878f6079 | mksettings: leveling target is a generic boss two levels up (1.12 warrior elite armor) except at 20 and 30 | Leveling only; `lowLevelTarget` is used only below 60. No Level 60 or shared raid effect. |

## Getting current
- Client: 1.60.1.70888 is still the newest (gate `client_build`).
- Blizzard: dev notes t/2360696 last post 10 (2026-11-04T19:35Z, already seen); hotfix thread t/2403040 last
  post 8 (2026-11-12T00:48Z, already seen). Blue tracker not re-scanned beyond these threads.
- TASKS.md 8306f1d52 adds Nov 11 follow-ups: Judgement as ranged (already mechanics-20261112-3), Hunter Volley
  no longer activating procs and Rend and Tear on Pounce (class work, owned by the class workers), and items
  revealed Nov 11-12 (data work). Nothing new on a shared mechanic.

## Verdicts
Added rows for the hunter pet melee speed, the Rake hit AP share and the leveling target to `ref/mechanics/verdicts.md`.
