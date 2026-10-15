# Leveling audit log

What the leveling auditor (`orchestration/auditor_leveling.md`) tried on each template, kept or
not, so later audits skip the dead ends. Newest first. Numbers are mobs killed in a 20 min
session (standard error), at the iterations given, against the rotation the round started
from.

## shaman_enh_leveling, Level 30 solo (`tools/rotopt/leveling/level30_enh_solo.json`)

### Round 1, 2026-10-15 (from the pass 1 rotation, 90.2 mobs)

Kept:

- **Engine bug, fixed:** when the last mob died, our target moved to the mob we'd found and
  not pulled, still at melee range for the sim, so a swing pulled it from up to 40 yards (and
  cut rests short). Now our target moves to it at its real distance, and nothing the rotation
  does goes while we sit, bandage or lie dead. Test: `TestLevelingRestWithMobWaiting`.
- **Look ahead:** look for the next mob while fighting at any health over 40 to 50%, pull it
  only when nearly full (over 90% health and 50% mana in combat, or out of combat once we
  don't need a rest), else rest with it waiting. Searching drops from about 235 to 130 to 170
  sec a session. 90.2 to 97.8 (4000 sessions, both seeds) in the template
  (`lookAheadHealth`, `chainHealth`, `chainMana`, `restHealth`).
- **Don't let a pull line stand us up:** the walk and pull lines wait for us to be fit to pull
  (part of the above); on its own +0.6 to +0.7.
- **Walk in instead of a Flame Shock pull** on top of look ahead: +0.1 to +0.6 (`pullSpell=0`).

Tried, no help (2000 sessions): Earth Shock only above 150, 250 or 400 target health (-1.3 to
-2.2); Stormstrike only above 200 (flat); both shocks gated at 250 (-2.9); Earth Shock rank 3
for rank 4 (+0.8 / -0.7, mixed); a combat heal at 40% with look ahead (-0.3 or worse, more
deaths) or at 20% (-0.1 to -0.5); Lightning Shield only in combat (within noise); pulling the
waiting mob out of combat at 40 to 60% health instead of resting (89.9 to 95.2); looking ahead
at any health (96.4, 0.25 deaths).

Wished for: Is Resting; the Pull Target's aggro radius (or whether it would notice us at
once); a plain "has a Pull Target"; rest until the Pull Target comes to us; the current
fight's expected damage.
