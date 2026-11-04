# Protection warrior tank: handoff

**Status 2026-11-04 (cloud session): done.** The searches converged, races ran at all three weights (Undead at each),
the three presets, APLs and requests are written, step 7 ran, and the doc's Presets, Numbers and Calls are written
(`docs/tank_warrior.md`). The list under "Left" below is finished; the rest of this file is the earlier state.

Brief: `orchestration/brief_tanks.md` (its last section, "Restart after the 70753 update (T2)"). Doc: `docs/tank_warrior.md`.
Raw search output, the scripts and the Balanced APL found are in `orchestration/tank_handoff/warrior/`.

## Done (A: the realignment)

- Merged `claude/clever-galileo-qim1ee` (659cad198). Tank tests realigned: Defensive Stance 80% x Defiance 7/13/20%
  (x2.16), Enrage at Fury's 8th place (`-00000005`); `TestP1TankWarrior` golden moved TPS only (plus the synthetic
  set's AllItems rows). `TestP1DPSWarrior`'s golden left to the orchestrator, as asked.
- Talent strings: the tank strings spend no Fury point past the 7th, so they read the same in the reworked tree
  (`35--251533120331221351`, `35-05-251530120331201351`). Nothing to remap but the tests' Enrage string.
- Reckless Swings (Fury row 4, reachable with 16 Fury points): its 25% of base health is now taken (damage taken for the
  tank's score, `sim/warrior/reckless_swings.go`), test `TestForeverTankRecklessSwings`; template knob `recklessSwingsRage`.
- Slam (15 s under Forever): template knob `slamRage`. Every search drops it (-150 to -380 at every weight).
- Gore Drinker: unreachable with Shield Slam (53 of 51 points), not simmed. Blood Craze no threat: the orchestrator's.
- Hit to the 8% cap (standards 4): Finger II (990612) hit 16 -> 24, crit 13 -> 5 (`tools/database/forever_synthetic_items_warrior_tank.go`).
  Regenerate the DB locally to test it (`go run ./tools/database/gen_db -outDir=assets -gen=db`, then with `-tags=with_db`); not committed.
- `docs/tank_warrior.md`: audit tables redone against 70753 and the notes, "What the update changed", tests, template.
- **What the update moved** (same build `35-05-251530120331201351`, WIP Balanced APL, Orc, 3,000 iterations, six
  scoring encounters' mean): Balanced 3,360.8 -> 3,447.0 (+86); TPS 1,501 -> 2,182 (+45%), SafeDPS 2,298 -> 2,343 (+2%:
  the stance's +44.5% against the reference member's 0.7 -> 1 threat a point), healing needed 418 -> 420 (`before.md`,
  `after_first.md`; "before" is the old branch worktree-agent-ac0218fc619732709 built in scratch).

## Pre-existing failure, not ours

`sim/core` `TestForeverConsumeProcDamage` ("Shadow Oil: no proc spell 1382") fails on the branch's committed db.bin too.

## B: searches so far (rotopt template `warrior_tank`, the six scoring encounters, synthetic set, 600 iterations a
candidate and 3,000 to confirm unless said)

Commands: `orchestration/tank_handoff/warrior/opt.sh <preset> <out> <settings> [flags]` (it runs rotopt from the
worktree with `-template warrior_tank -tank <preset> -tank-encounter <the six> -iters 600 -confirm-iters 3000`).
Settings: `mkreq.py tools/tankbench/req/warrior_protection.json ui/tank_warrior/gear_sets/level60_phase1_synthetic.gear.json <talents> RaceOrc <out>`.
Fix opt.sh's `W=` path (scratch) for the new session.

Knobs everywhere (unless said): `shieldBlock=1,sbRage=0,demo=0,tc=1,tcRefresh=2,debuffsFirst=0,revengeFirst=0,sunderUp=1,sunderRage=30,hsRage=50,bloodrageRage=30,opener=0,chargeFrom=0,potion=1,shieldWall=2,wallAt=12,trinketAt=0,recklessSwingsRage=0,slamRage=0`
(opener=0 scores the same as the WIP's opener=2 once Vanguard is gone: Charge from Defensive Stance needs it; recklessSwingsRage=0 drops the line the talents don't use).

| Preset | Talents (best so far) | Score (Orc) | State |
|---|---|---|---|
| Balanced | `05-05050005-250530020310001351` (Deflection 5; Cruelty 5, Unbridled Wrath 5, Enrage 5; Protection 31 with Shield Slam, no Last Stand or Vanguard, Improved Sunder Armor 1) | 3,617.2 ± 8.2 (was 3,444.0 for `35-05-...`) | converged: knob search on these talents kept every knob (`knobs_bal2.txt`); talent search from the Reckless Swings start ended here (`ts_bal_rs.txt`); the plain start stuck at `35-0502-251530120331001351` 3,469 (`ts_bal.txt`) |
| Max threat | `3-05050015-250430020330001351` at round 5 (search killed mid-run) | 4,594.6 ± 12.1 | knobs (on the old talents, `knobs_mt1.txt`): tc=0, hsRage=30, revengeFirst=1, rest as above. Left: finish the talent search from this string, then knobs again |
| Max survival | start `05-05050005-250530020310001351` | 1,195.9 ± 24.2 (2,000 iterations a candidate, 6,000 to confirm) | knobs kept (`knobs_ms2.txt`, 2,000 iterations: demo=1 -62 ± 34, hsRage=30 -67 ± 26); talent search killed at its start. Left: the talent search (2,000 iterations: Max survival is noisy) |

Races at Balanced (3,000 iterations, `races_bal.txt`): **Undead 3,795.3**, Tauren 3,680.6, Human 3,637.4, the Skyborne
3,626.1, Night Elf 3,625.4, Troll 3,622.9, Orc 3,617.2, Dwarf 3,613.3, Gnome 3,610.8. Undead's lead is Touch of the
Grave's drain (heals the tank), modelled at its "up to 5% of max health" cap (`sim/core/racials_forever.go`): a call
for the user (at a lower drain, Tauren). Races at Max threat and Max survival not run yet.

Reckless Swings: no searched build keeps it; from a 16 Fury start the search traded it for Deflection (`ts_bal_rs.txt`
round 2, the knob on cooldown). Demoralizing Shout: off at every weight so far (the raid's is taken away,
`tankscore.OwnDebuffs`); at Max survival demo=1 is -62 ± 34.

## Left

1. Finish Max threat's and Max survival's talent searches, alternate with knob searches until neither moves; races at
   those two weights.
2. Write the presets: `tools/tankbench/req/warrior_protection_{max_threat,balanced,max_survival}.json` (balanced replaces
   `warrior_protection.json`; `mkreq.py ... <apl>` makes them), the three APLs (`-eval -out <dir> -name level60_<preset>`;
   `warrior/level60_balanced.apl.json` here is the Balanced one with recklessSwingsRage=100: regenerate with 0), and
   `ui/tank_warrior/presets.ts` (its talents and races are placeholders: `SURVIVAL_TALENTS`, `Race.RaceRACE_*`; the
   Max survival APL file it imports doesn't exist yet), then `npx tsc --noEmit -p .` (don't symlink node_modules into a
   worktree: `make` runs `npm ci` through the link and empties the main checkout's; it happened here and was restored).
3. Step 7: tankbench on all seven encounters, 1,000 iterations, the three presets (`bench.py` prints the table and the
   six's mean); write the numbers and the Calls for the user into `docs/tank_warrior.md` (its Presets, Numbers and
   Calls sections aren't written yet).
4. Full `go test -tags=with_db -timeout 40m` over `./sim/...` without practice and web.
