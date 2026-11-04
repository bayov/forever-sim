# Pass: healer mechanics

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03 for the healing sim's healers
(`docs/healing_sim.md`, `orchestration/brief_healers.md`), as `mechanics.md` is for the DPS specs.

You check that every heal, healing talent and healing-fight mechanic the healer presets use is WoW Forever's:
the Restoration shaman, Holy paladin, Restoration druid and Holy and Discipline priest, and the engine under them.
`ref/healer_mechanics/verdicts.md` holds what's verified, settled and open: read it first, and keep it current.

## Rules

- Read-only on code, as `mechanics.md`: proposed changes as patches under `$AUDIT_WORK/healer_mechanics/`, quoted
  in the finding. Small healing fights to test a claim are fine (`tools/healbench`, or a test-style request).
- Sources in tier order, as `mechanics.md` (tier 1 the Forever client on wago.tools and Blizzard's Forever notes;
  tier 2 Season of Discovery; tier 3 Era, Anniversary, 2019 Classic; never vmangos, never a tooltip alone).
  Label every verdict with its tier.
- Findings in `findings/healer_mechanics.md`. Engine findings are `needs: orchestrator fix`, class findings
  `needs: worker fix` with the spec in `affects`.

## Setup

```sh
source tools/audit/setup.sh
export AUDIT_DB=$(python3 tools/sources/wago.py fetch $(python3 tools/sources/wago.py latest) --set mechanics)
go build -tags=with_db -o $AUDIT_WORK/healbench ./tools/healbench
```

## The pass

1. **Get current**, as `mechanics.md` step 1: the newest client build, Blizzard's posts since `dev_notes_seen`,
   and `orchestration/update_ledger.md` since your last pass. Anything newer than a verdict reopens it.
2. **What changed:** `git log <last.master>..origin/master` on the watched paths. Read every changed heal, talent
   and engine file in full, not just the hunk.
3. **Each healer's spells** (on the first pass every one, then what changed or what a new build touched): for
   each heal rank, the base heal and its growth, variance, cost, cast time, coefficient (and the class's
   downranking rule), cooldown, HoT ticks and period, jumps and falloff, and what the talents and set bonuses
   change, against SpellEffect, SpellMisc, SpellLevels, SpellPower, SpellCastTimes, SpellCooldowns,
   SpellDuration and the talent rows' SpellEffect (EffectAura, misc values and class masks: which spells a
   modifier reaches). The classes' docs (`docs/healing_<class>.md`) carry their audit tables; check them, don't
   copy them.
4. **The engine**: health and death, heal targets (Tank, Most Hurt, Lowest Health, assignments), heals in flight
   and Expected Healing, cancelling, overheal booking, absorbs (Power Word: Shield, Divine Aegis), HoT
   snapshotting and refresh, crit heals (1.5x unless a talent says otherwise), mana regen while casting (the
   five second rule, Spirit regen under Forever), Innervate, Mana Tide and Mana Spring reach (party), raid
   damage events (`sim/core/raid_damage.go`) against Forever's spell school and resistance rules.
5. **The tests**: for each heal and healing talent, is there a test that fails when the mechanic breaks? List
   the ones without (a finding, low unless the mechanic is untested and wrong).
6. Report: a table per spec (spell or talent, client, sim, verdict, tier), the engine checks, new findings.

## Settled (don't reopen without new tier 1 evidence)

- The user's healing rulings in `docs/healing_sim.md` (the score, assignments, bound ranks, scares reported only).
- Healing on top of spell power is a separate stat, priced at 0.45 in gear (standards 1b, 35).
