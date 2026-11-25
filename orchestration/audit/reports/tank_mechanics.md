# Tank mechanics pass (2026-11-25, phase0-step4-publish 01cc70b80)

No new finding. Range 63129ce35..01cc70b80: 1fed43cbf (Auto Time To Next with a pending extra attack, see the
mechanics report) and standards 43 (tanks keep their synthetic sets).

- 1fed43cbf: the Protection warrior APLs' swing-timer conditions read the kept schedule; the tank bench was re-simmed
  in 01cc70b80.
- tank_mechanics-20261124-3: the racial lines part is now pinned by TestForeverTankRacialLines
  (sim/warrior/tank_warrior/racials_apl_test.go, 4c26b4b66), all three level 60 requests, six races. The finding
  stays open for Eureka!'s damage and cost scope, Stoneform's damage taken and the tank-side Troll Regeneration.

## Open findings

tank_mechanics-20261124-4, -20261124-3 (updated), -20261117-1, -20261108-1, -20261107-1.
