# Leveling audit log

What the leveling auditor (`orchestration/auditor_leveling.md`) tried on each template, kept or
not, so later audits skip the dead ends. Newest first. Numbers are mobs killed in a 20 min
session (standard error), at the iterations given, against the rotation the round started
from.

## shaman_enh_leveling, the enhancement leveling plan 10 to 60

Numbers here are XP an hour (the plan's score) at 4000 sessions on seeds 1 and 1000001, or
2000 on seed 1 for the screens, against the searched plan's rotation at the level.
`docs/leveling_plan.md` (The audit) has what we kept.

### Round 1, 2026-10-19 (the searched plan, 107h 5m, to 102h 8m)

Kept: Stoneskin Totem for the earth totem (every pull at 10 to 19 and 44 to 59, from two
pulled at 20 to 43), Earth Shock a rank down at 14 to 19, pull in combat only under 70% target
health at 10 to 19, Flame Shock on a fresh target (above 70%) at 20 to 31, Searing reach 20 at
20 to 31, no Fire Nova from 22, Frost Shock rank 1 at 22 to 24, Stormstrike above the totems at
25+, Maelstrom bolt from its 4th point, Searing with one mob only above 40% target health at 32
to 44, Searing only while we stand at 44+, Flame Shock DoT with Call of Flame. Talents:
Restoration at 38 (Tidal Focus 4 and 5, Mindfulness), Maelstrom Weapon 43 to 47, Rage of the
Farseer 48, then Convection and Call of Flame.

Tried, no help:

- 10 to 19: Lightning Shield off -8.9% or rank 1 -3.5%, Earth Shock off -24%, Earth Shock only
  above 15% -0.5%, Flame Shock DoT -2.8%, rank 1 Lightning Bolt pull -4%, Lightning Bolt
  between swings -9 to -26% (casting stops the swing and dodge), Flametongue Weapon -0.2 to
  -1.8%, walk to the found mob only within 25 to 40 yards -2 to -3%, no in-combat chain pulls
  -0.5%, chainMax 3 -1 to -4%, Earth Shock rank 1 at 10 to 13 -2 to -3.5%, Stoneskin rank 1 -1.2%,
  shock floor, rest and chain grids within ±0.6%.
- 20 to 31: Stoneskin from one pulled, Stoneskin ranks 1 and 2, Searing only at two pulled -1
  to -1.8% or rank 1 -1.7 to -3.7%, Searing reach 10 or 25, Earth Shock rank 3 for rank 4 -3%,
  holding Earth Shock for the mark, Frost Shock at 25+ -2.3%, Flametongue -4% and Windfury
  Weapon -3.6% at 30, Healing Stream at 30 with Stoneskin -0.6%, look-ahead and rest grids flat.
- 32 to 44: Maelstrom bolt at 3 or 4 stacks -0.8 to -3%, Flametongue -0.3 to -2.2%, Frostbrand
  -2.3 to -4.4%, Windfury Weapon -6 to -9%, Grace of Air -4%, Magma with two on us -1 to -2.4%,
  Searing rank 1 -1.3 to -2.5%, Healing Stream -0.1 to -3.3%, bandages -0.9 to -6.5%, rest to 80%
  health -1%, Earth Shock ranks 3 and 4 down to -2.5%, the same Searing gate on every totem
  -0.1 to -0.4%, Blood Fury on a fresh target 0 to +0.4% (left out), Rage of the Farseer on a
  fresh target +0.1 to +0.4% (lost in the route change).
- 45 to 59: Lightning Bolt as a filler -0.2 to -10%, Maelstrom bolt at 4 stacks -2%, top-rank
  Lightning Bolt pull -2.5%, Earth Shock ranks 5, 4 or 1 -0.6 to -2.6%, Flame Shock rank 1 -4.2%,
  Fire Nova on two -0.1 to -0.8%, running search -22%, Water Shield -8.1%, thresholds within
  ±0.4%. The rest mana trigger does nothing there: rests start on health.
- Talents: Mana Tide as a real route 43 min worse over 49 to 59, Elemental tail after the
  Restoration turn with the old rotations 26 to 40 min worse (5 min better with the new ones),
  the Restoration package at 26 to 41 or 44 to 48 9 to 30 min worse, Maelstrom before
  Mindfulness 22.7 min worse, Toughness or Ancestral Knowledge tails 5 to 14 min worse.

Wished for: the Pull Target's aggro radius (or whether it notices us at once), which earth
totem is up, a target's time to die, "has the talent" as a value, and a rank pick by level in
the template.

## shaman_enh_leveling, Level 30 solo (`tools/rotopt/leveling/level30_enh_solo.json`)

### Pass 4, 2026-10-15 (one imbue: Windfury Weapon alone, 91.9 mobs)

The settings had Rockbiter Weapon in the consumes and Windfury Weapon in the class options,
and the sim put both on the main hand. Now the class options' imbue wins and the settings
have Windfury alone. Every number above this entry is with both imbues.

- **Fire Nova with two mobs on us** (`fireNova=1`): 91.89 / 91.86 to 92.98 / 92.99 (4000
  sessions, seeds 1 and 1000001). Kept. Round 2 had it at -1.0 with both imbues.
- Nothing else moved.

### Pass 3, 2026-10-15 (optimizer from the round 2 knobs, 99.53 mobs)

- **Pull with rank 1 Earth Shock** (`pullSpell=1`) instead of walking in: 99.53 / 99.48 to
  100.02 / 99.89 (4000 sessions, seeds 1 and 1000001). With look ahead the mob is often found
  while we fight, and the shock brings it in without a walk. Kept.
- Nothing else moved: combat heal at 40% -0.6, 20% -1.3; other knobs flat or worse.

### Round 2, 2026-10-15 (from 97.8 mobs, after pass 2)

Rests wait on mana, not health: 1305 of 1478 rests (40 sessions) ended at about full health
with mana just reaching its target, since drinking gives 1.8% a second and eating 4%. Mana
spent on damage pays (no Stormstrike -20, no Earth Shock -8, no Lightning Shield -4.2), so the
gains are cheaper mana, not less casting.

Kept (4000 sessions, seeds 1 and 1000001; "ablation" is taking it out of the final):

- **Mana Spring Totem in combat, no Totemic Recall.** Recall's global cooldown before every
  search costs more than a quarter of a 40 mana totem gives back; the optimizer missed it
  because the two knobs only pay together. Ablation -0.93.
- **Searing Totem rank 1 (25 mana) with two mobs pulled.** The top rank with one mob on us is
  -2.2. Ablation -0.43 (`searingRank1`, `searingMinPulled`).
- **Rest mana to 80%, health to 90%.** Ablation -0.34 (`restUntilMana`).
- **Pull the next mob as the current one dies:** one mob pulled, it's under 20%, we're over
  70% health and 40% mana. Ablation -0.40 (`chainDyingBelow`, `chainDyingHealth`,
  `chainDyingMana`).
- Together: 97.82 / 97.69 to 99.53 / 99.48.
- **Engine bug, fixed:** our own lower-rank Strength of Earth and Stoneskin totems gave the
  top rank's buff (rank 1 Stoneskin scored 100.4 for that reason). Test: `TestOwnTotemRank`.

Tried, no help: rest grid (restHealth 70/80, health to 100, mana to 50/60/70/100, restMana 30)
0 to -3; a lower rest threshold with no mob waiting -0.1; bandage or heal between pulls when
health is low and mana high (flat, rarely fires); combat heal at 20% -0.85, 40% -1.0, off -4.9;
Lesser Healing Wave rank 1 -0.3; Stoneskin rank 3 -0.4 to -2.0; Strength of Earth rank 2 at two
pulled -0.85 to -1.2; Magma at two on us -7.6 to -8; Healing Stream -0.2 to -0.7; Fire Nova at
two on us -1.0; Mana Spring only before rests -0.35 or only at two pulled -0.4; Lightning Bolt
during the run-in -7.3 to -9.5; Lightning Shield rank 2 -0.7; Earth Shock rank 3 -0.15 (mixed)
or rank 2 -1.3; shock mana floor 30/50 flat, 70 -0.15; looking for a third mob -5 to -11
(deaths 0.25 to 0.7: a found mob inside its aggro radius joins at once); no chain pull in
combat -2.1; chainHealth 80, chainMana 40/60, lookAheadHealth 30/50 flat or worse; Windfury
Weapon -8.2, Flametongue Weapon flat (settings, not the APL).

Wished for: the session's state (ready, searching, pulling), so Totemic Recall could go out
while we run; the Pull Target's aggro radius; the current target's time to die or absolute
health; how long a rest will take on each bar.

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
