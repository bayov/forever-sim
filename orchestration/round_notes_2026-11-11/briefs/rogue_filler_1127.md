# Rogue filler talents (2026-11-27)

Combat Swords and Seal Fate Swords take Deflection 3/3 and Riposte 1/1. Riposte only arms after the rogue
parries, and the boss never swings at the rogue (the report's raid hits are Fire spells), so it never fires,
and Deflection only adds parry. Talent strings: `tools/dps_report/req/rogue__combat-swords.json`
(`00530320201-3200330112012321551`) and `rogue__seal-fate-swords.json` (`0053032024105103-320033011201132`).
Tree: `tools/forever_talents/data/rogue.json`.

User ruling (2026-11-27): rogue points that add no DPS go to utility: Endurance (Sprint and Evasion
cooldown), Improved Sprint, Lightning Reflexes (dodge) and the like, never Deflection or Riposte.

## Do

1. Record the ruling in `orchestration/standards.md`, and make the rotopt talent search follow it, so a tie at
   zero DPS goes to those utility talents and a later search won't put Riposte back.
2. First check whether the spare points have a DPS use (paired seeds, keep a swap only above noise):
   - Combat Swords: Adrenaline Rush needs only 1 filler point in rows 0-5. Try the other 4 in Vile Poisons
     (0/5, reachable).
   - Seal Fate Swords: Dual Wield Specialization needs 5 filler points in rows 0-3. Try Weapon Expertise 2/2
     and Vile Poisons 5/5.
3. Whatever filler is left goes to Endurance and Improved Sprint first, then Lightning Reflexes, keeping the
   row gating legal. Check all six rogue builds, their duty variants and race rows the same way.
   `talent_check` must pass.

## Publish

Requests, race rows and metas, `stale.mjs`, `pipeline.sh hydrate/run/publish` for the changed rows,
`build_data.py`, `changes.py`, one Changes tab entry (per build: talents changed, DPS before and after).
Run the rotations and results auditors on the changed builds. Merge to master (fetch and merge first, never
force-push), `make dps-report-deploy`, check the live page.

## Hand back

One message to the parent, under 200 words: master commit, deploy check, per build the talent change and DPS
before and after, the DPS swaps tried and their results. Then stop.
