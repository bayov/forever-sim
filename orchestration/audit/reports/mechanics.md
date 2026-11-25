# Mechanics pass (2026-11-25, phase0-step4-publish 01cc70b80)

Quiet. The only sim change in range 63129ce35..01cc70b80 is 1fed43cbf (sim/core/attack.go,
apl_values_auto_attacks.go): under Forever an extra attack keeps the main hand's schedule, and APL Auto Time To Next
now reads that kept schedule (`MainhandScheduledSwingAt`) for MainHand, Melee and Any instead of the pending extra
attack's swing at now.

- Read against Forever's rule (extra attacks don't reset the swing timer, forever_class_changes.md): correct.
- The commit message named a TestWarriorWfGate that doesn't exist. The branch adds
  TestForeverAutoTimeToNextKeepsSchedule (sim/core/auto_time_to_next_forever_test.go, 4c26b4b66): it queues a
  Windfury Totem extra attack and checks all three readings equal the kept schedule; on the old code Melee and Any
  read 0s.
- DPS effect in the shared run: Arms +1.2, Arms duties +0.6 to +0.9, Fury 2H +0.7 against the metas (with 78802e6e).

Also on the branch: the bulk sim progress reporter race (a1efdca09, `send on closed channel` in TestAbort) is fixed;
no sim result reads it.

## Open findings

Unchanged in range: mechanics-20261124-1, -20261123-1, -20261120-2, -20261118-2, -20261118-1, -20261113-2, -20261104-1.
