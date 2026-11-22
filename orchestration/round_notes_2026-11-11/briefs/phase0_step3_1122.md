# Phase 0 gear, step 3: the full searches (2026-11-22)

The plan is in `briefs/phase0_gear_1122.md`. Steps 1 and 2 are done: the tooling and sets
(`phase0_gear_1122.md`), and the hand check of every slot's candidates, with effects, hidden stats
and availability tags (`phase0_step2_weapons.md`, `phase0_step2_jewelry.md`,
`phase0_step2_mail-plate.md`, `phase0_step2_cloth-leather.md`, all under
`orchestration/round_notes_2026-11-11/`). Read the notes that cover your specs' armor and weapons
before you start. DPS specs only, as the user asked.

Your session prompt names your group and its specs (keys of `tools/rotopt/phase0_specs.json`).

## Task

For each spec in your group:

1. Merge origin/master. Run `DRY_RUN=1 SPEC=<spec> tools/rotopt/bis_phase0.sh` and compare its
   pool file with the one in `phase0_pool/`. Look into any new unmodelled effect or unlisted item
   before the search.
2. Run the full search: `SPEC=<spec> tools/rotopt/bis_phase0.sh`, in the background with its
   output in a file. It writes `ui/<dir>/bis/level60_phase0_<spec>.bis.json` and its effects file.
   Check `nproc` and run as many specs at once as the cores allow without slowing each other. Wait
   for the task notification, don't poll.
3. If the run stops on unlisted picks, check each one (source, cost, and whether it's open before
   the raids, by the rules in `tools/rotopt/phase0.go` and the step 2 notes) and tag it in
   `phase0_content.json` or `phase0_enchants.json`. If a pick turns out "out", run that spec's
   search again. Never ship a list with `ALLOW_UNLISTED=1`.
4. Hand-check the list:
   - Check each pick's source and availability.
   - Flag every pick with an unmodelled effect, hidden stats or a doubtful tag.
   - Compare the list with a public pre-raid or Phase 0 BiS list for the spec, if the web-fetch
     agent can reach one. Never bypass Wowhead's bot check; other sites, or none, are fine.
   - For each big difference, give our reason or fix our mistake.
5. Calibrate: sim the spec's report build at 180s twice, once on its Phase 0 set and once on the
   report's synthetic gear, both at 10000 iterations or more. Give both DPS numbers and the gap.
   This answers whether the synthetic gear is set right, and step 4 needs it.

## Shared files

Other step 3 workers edit `tools/rotopt/phase0_content.json`, `tools/rotopt/phase0_enchants.json`
and `docs/bis_lists.md` too. Merge origin/master often, and keep both sides on a conflict. Each
spec's list files are yours alone.

## Coordination

- The wowsims/forever comparison session (`session_01TXnjyQoHMJg2CjHPYTGUMX`) is re-tuning Frost
  and Frostfire: it writes the mage rotopt templates, `mage_presets.sh`, `ui/mage/apls`, mage
  goldens and hashes, then re-sims and deploys the report's mage builds. Don't touch those files.
- The mage group runs arcane and fire now. It holds frost and frostfire until the parent session
  messages that the re-tune is on master, then merges and runs them.
- Don't touch the report (`tools/dps_report`, the deploy), presets, rotations, talents or bayovui.
  Step 4 builds the Phase 0 report variant.

## Rules

- Never force-push. End every commit message with your session prompt's trailer lines.
- Keep model names out of files. Keep long runs in subagents or in the background.
- Merge to master and push without a PR, then push to your own development branch too.
- User style in anything they read: concise, no em dash or semicolon, "180s" not "180 s".

## When done

Write `orchestration/round_notes_2026-11-11/phase0_step3_<group>.md` with, for each spec:
- the picks per slot, with source;
- the flags from the hand check;
- the BiS list comparison;
- the calibration numbers.

Push it. Then send one message to `@parent` with the master commit and a summary under 250 words:
each spec's Phase 0 DPS against its synthetic-gear DPS, notable picks, and the calls only the
user can make (doubtful items, costs, sources we can't confirm). Then stop.
