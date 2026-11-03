# Pass: tank mechanics

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03 for the tank sim's tanks
(`docs/tank_sim.md`, `orchestration/brief_tanks.md`), as `mechanics.md` is for the DPS specs.

You check that every tank ability, tank talent and tanking mechanic the tank presets use is WoW Forever's: the
Protection warrior, Protection paladin, Feral tank (bear) and Warden shaman, and the engine under them (the boss's
attack table, threat, Rage from damage taken, block, armor, the tank sheet, the tank score's replay).
`ref/tank_mechanics/verdicts.md` holds what's verified, settled and open: read it first, and keep it current.

## Rules

- Read-only on code, as `mechanics.md`: proposed changes as patches under `$AUDIT_WORK/tank_mechanics/`, quoted in
  the finding. Small fights to test a claim are fine (`tools/tankbench`, or a test-style request).
- Sources in tier order, as `mechanics.md` (tier 1 the Forever client on wago.tools and Blizzard's Forever notes;
  wowsims/forever's in-game measurements and combat logs next; tier 2 Season of Discovery; tier 3 Era, Anniversary,
  2019 Classic; never vmangos, never a tooltip alone). Threat values are rarely in the client: say which tier each
  rests on. Label every verdict with its tier.
- Findings in `findings/tank_mechanics.md`. Engine findings are `needs: orchestrator fix`, class findings
  `needs: worker fix` with the spec in `affects`.

## Setup

```sh
source tools/audit/setup.sh
export AUDIT_DB=$(python3 tools/sources/wago.py fetch $(python3 tools/sources/wago.py latest)+hf --set mechanics)
go build -tags=with_db -o $AUDIT_WORK/tankbench ./tools/tankbench
```

## The pass

1. **Get current**, as `mechanics.md` step 1: the newest client build, Blizzard's posts since `dev_notes_seen`,
   and `orchestration/update_ledger.md` since your last pass. Anything newer than a verdict reopens it.
2. **What changed:** `git log <last.master>..origin/master` on the watched paths. Read every changed tank
   ability, talent and engine file in full, not just the hunk.
3. **Each tank's abilities and talents** (on the first pass every one, then what changed or what a new build
   touched): ranks and levels, values, costs, cooldowns, durations, threat (multipliers and flat threat),
   mitigation and avoidance, procs and their rates, form, stance, aura and shield effects, against SpellEffect,
   SpellMisc, SpellLevels, SpellPower, SpellCooldowns, SpellDuration, SpellShapeshiftForm and the talent rows
   (EffectAura, misc values, class masks, trait curves). The classes' docs (`docs/tank_<class>.md`) carry their
   audit tables; check them, don't copy them.
4. **The engine**: the boss's attack table against a tank (miss, dodge, parry, block, crit under Forever's boss
   crit and `classic_crit`, crushing blows), Defense, armor and its cap, block value, Rage from damage taken
   (standards 21), threat (stances, forms, Righteous Fury, Salvation, the tank's own debuffs,
   `tankscore.OwnDebuffs`), boss abilities as raid damage events (melee ones rolling the tank's avoidance), the
   tank sheet (`sim/core/tank_sheet.go`) against the rolled outcomes, and the tank score's replay
   (`sim/core/tank_score.go`: nothing in a scoring run may read the tank's health).
5. **The tests**: for each tank ability and talent, is there a test that fails when the mechanic breaks? List the
   ones without (a finding, low unless the mechanic is untested and wrong).
6. Report: a table per spec (ability or talent, client, sim, verdict, tier), the engine checks, new findings.

## Settled (don't reopen without new tier 1 evidence)

- The user's tank rulings in `docs/tank_sim.md` (weights and the three presets, healing needed as the survival
  score, Forever's 0.6% boss crit until a dungeon log, the stand-in healer, a tank applies its own debuffs,
  Anniversary encounters score tanks).
- Rage from damage taken (standards 21) and from white hits (k flat, standards 21's level scaling ruling).
