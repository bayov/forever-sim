# Pass: results and spec strength

Follow `orchestration/audit/PASS.md` around this doc. It started as the long-running results auditor's handoff (2026-10-17); keep it current when the method changes.

For a new auditor who has only the repo. It covers the pass as it ran through pass 42 (2026-10-17). The role itself is in
`orchestration/auditor_results.md` and the shared rules in `orchestration/auditors_common.md`. In short: explain why each spec is as
strong or weak as it is, flag numbers that look like modelling errors rather than balance, and check that every build is simmed
the same way (same raid, budget, iterations and fight lengths).

Standing rules from the briefs:
- Read-only on code, presets and docs; don't commit (the audit run commits your files) or comment on PRs. Findings go to `orchestration/audit/findings/results.md`.
- Don't use `tools/rotopt` (its searches, `-eval`, `-timeline`, `-breakdown`). mksettings, which only writes the settings JSON, and wowsimcli are fine.
- Don't put model names in anything you write.
- Sources for Forever facts, best first: the Forever client data (wago.tools db2, newest build), then Season of Discovery, then Anniversary or 2019 Classic. Avoid vmangos. Tooltips can be wrong.

## What to watch

Only `origin/master`. Every class PR through #114 is merged, and since pass 34 the method is master plus the DPS report's own
requests. The class branches are history.

Paths whose changes can move a published number:

    sim/  proto/  tools/rotopt/mksettings/  tools/database/  assets/database/
    ui/*/presets.ts  ui/*/apls/  ui/*/gear_sets/  tools/dps_report/req/

`assets/database/` is new to the list: the `with_db` binaries embed it, and my pass prompt didn't name it. It hasn't changed
since the last full run, so nothing was missed.

Also read, without simming:
- `ui/dps_report/data/builds.json`: the dashboard's published DPS. A republish changes it.
- `orchestration/standards.md` and `orchestration/spec_assumptions.md`: rulings and assumptions. A new ruling usually comes with a sim change.
- Each meta's `published.source` (`tools/dps_report/req/<id>.meta.json`): which commit and re-sim a number came from.
- `.agents/skills/dps-report/state.json`: the report version, the commit it was published from, and the current client build.

External sources (new client builds on https://wago.tools/builds, Blizzard's Forever dev and beta notes) reach the sim through
the class workers' ports and the user's rulings. I didn't poll them myself; the pass catches them as code changes. When a port
cites the client, check the cited values against the db2 tables (see `orchestration/audit/ref/results/MANIFEST.md` for the URLs).

### Is the pass quiet?

    git fetch origin master
    git log --oneline <last audited sha>..origin/master
    git diff --stat <last audited sha> origin/master -- <the paths above>

- Nothing in the watched paths changed: quiet. Say so in one line, with the new master sha.
- Only `ui/dps_report/` changed (a report republish): check the dashboard against the reference, with no sims needed:
  `python3 tools/audit/results/compare.py builds ui/dps_report/data/builds.json orchestration/audit/ref/results/req_summary.txt`.
  Rows that differ from the previous builds.json should be moves you already verified. Pass 41 is an example: report v35 moved 7
  rows, all verified in pass 40.
- `tools/database/` changes that are only the Wowhead scrapers (`tools/database/forever_wowhead/*.py`) or `gen_db`, with no
  regenerated `assets/database/`, don't reach the binaries. Example: b389688f7.
- Anything else: run the full pass (about 5 minutes). When unsure, run it.

## Setup (fresh container)

    apt-get install -y protobuf-compiler || (apt-get update && apt-get install -y protobuf-compiler)
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    export PATH=$PATH:$HOME/go/bin
    source tools/audit/setup.sh                         # AUDIT_WORK and the toolchain
    export AUDIT_TREE=$(dirname $(tools/audit/build_cli.sh origin/master))/tree   # a detached checkout of origin/master
    export AUDIT_WORK=$AUDIT_WORK/results; mkdir -p $AUDIT_WORK   # binaries, runs, drift output
    # The loop's shared run ($AUDIT_WORK/../runs/master/summary.tsv: every request, 20k, seed 1, the same settings as runreq.sh)
    # can stand in for step 2; check its settings match before you reuse it.
    export AUDIT_REF=$AUDIT_TREE/orchestration/audit/ref/results

Re-run the proto generation after every master update (`build.sh` does it). Without it, a build after a merge fails on
undefined proto fields (for example ReplaySeeds or IterationLog). The generated `*.pb.go` files are gitignored, so this
leaves the tree clean.

## The pass

    bash tools/audit/results/pass.sh <tag> [iterations=20000]      # SKIP_BUILD=1 to reuse the binaries, JOBS=6 by default

What it does, in order (each script has a usage line at the top):
1. `build.sh`: protos, then `wowsimcli` and `mksettings` with `--tags=with_db` into `$AUDIT_WORK/bin`.
2. `runreq.sh` on every `tools/dps_report/req/<id>.json` (66): the committed request as is, iterations set to 20,000 and seed `"1"`,
   at the request's own 180 s. It prints `<id> mine <dps> pub <published.dps>`. Output goes to `$AUDIT_WORK/runs/<tag>/summary.txt`.
3. `compare.py summary`: flags any request more than 0.15 off published, then lists every build that moved since `req_summary.txt`.
4. `regen.py`: the drift check (below), written to `drift_report.txt` and diffed against the reference.
5. `unread_fields.py`: every field set in a request that no sim code reads. Expected output: only `UNREAD notes` (APL rotation notes, read only by the UI).

The pipeline that publishes the numbers (`tools/dps_report/README.md`) runs a merged tree with local patches 0001 to 0004. Those
patches do nothing unless their env vars are set, so a plain master binary reproduces the published numbers exactly.

### Tolerances and what counts as a finding

- **Published vs mine:** at 20k iterations and seed 1, a difference of 0.15 or less counts as exact; all 66 are within that today.
  Over 0.15, check the meta's `published.source` first: report re-sims lag master by hours, so a gap right after a sim change is
  "stale, waiting on the re-sim", not a bug. Say which builds and by how much. A gap that persists after a re-sim is a finding.
- **Never judge at 5,000 iterations.** Noise there reached 2.4 (Fury) and 0.5 (Elemental). If a quick run is more than 0.5 off, re-run at 20k before saying anything.
- **Moves against the reference:** every MOVED line needs a commit that explains it (a sim change, a ruling, a preset change).
  Verify the size where you can (paired runs: the same request on the binaries before and after). An unexplained move is a finding.
- **Drift:** `drift_report.txt` should equal the reference, except for changes the commits explain. A new difference means
  mksettings and the committed request disagree, so anyone who regenerates the request by the README gets a different number. Measure the DPS effect:
  sim `$AUDIT_WORK/drift/<id>.regen.json` at 20k, seed 1, and compare. Report it to the owner of whichever side is wrong (the class
  worker for `spec_*.go` or presets, the report owner for the request). Example: the Ret aura drift (mksettings wrote Retribution
  Aura, requests had Concentration; Shock −2.9). It was fixed in #106.
- **Unread fields:** anything besides `notes` is a finding (a request carrying a field the sim dropped, like the old `startingTotems`).
- **Modelling outliers:** a spec far above or below what its mechanics suggest, a big move without a reason, two specs with oddly
  identical numbers, a race result driven by an accident, a talent the search ignored that should matter. Check with code,
  client data, the result JSON (`summ.py <out.json>` gives the per-action breakdown) and a combat log (`sim.sh` with iterations
  of 2 or fewer turns on debug). Then do the arithmetic. Say how sure you are.

### Expected drift (standing; don't re-flag)

All of these are in the reference `drift_report.txt` and in the metas' notes:
- **Duty presets** (standards 23, "Raid duties"): the raid's copy of the duty is off in the request. Faerie Fire and Curse of
  Recklessness for the druid FF duties, Judgement of the Crusader for paladin `-duty-crusader`, Curse of the Elements for warlock
  `-duty-curse`, Battle Shout for warrior `-duty-battle-shout`.
- **Encounter presets:** `mobType` Undead (paladin), Humanoid (rogue), Dragonkin (`warrior__arms_dragonkin`). mksettings isn't given a mob type.
- **`paladin__level_60` and `level_60_sub_holy`** (and their Crusader duties): `primarySeal` Command, per the notes. mksettings defaults to Righteousness.
- **Priest consumes:** `priest__shadow-holy` adds Greater Mageblood. `priest__smite` takes the Smite consumes (Major Mana Potion, Greater Firepower, Greater Mageblood, Greater Spirit).
- **Warlock leftovers:** all 8 requests carry `sanctityAura: false` and `curseOfShadow: false`, which mksettings no longer writes. Both are false, so they have no effect. They can go at the next regeneration.

`regen.py` maps each id to a mksettings spec. All paladins are `ret`, all priests `shadow` (Smite too), and rogue Mutilate
and Subtlety are `rogue_assassination` and `rogue_subtlety`, with every other rogue `rogue_combat`. Hunter duties get
`_mark` or `_scorpid`. A new request id may need a new mapping; an unmapped id fails with MKSETTINGS FAIL or a KeyError.

## Findings: where they go

Every finding goes to `orchestration/audit/findings/results.md` (format in `PASS.md`), with its evidence (file and line, numbers, the paired sim or arithmetic) and a suggested change. A class-specific one names the class in `affects`; a cross-class one is `severity: high`. The fairness auditor reads your findings file, so nothing is sent to it.

Every pass writes `reports/results.md`: the base preset table when numbers moved, and your explanation of the ranking (below). When a new run becomes the reference, replace `ref/results/req_summary.txt` (and `drift_report.txt` when it changed) and record the sha in `state/results.json` as `baseline`.

Don't repeat a finding the owner already answered unless you have new evidence. Don't send nits.

## Baseline (2026-10-17)

- **Last full run:** master 1735382cd (pass 40, 2026-10-17 00:34 UTC). All 66 requests are exact against published; the drift report is the reference copy; unread fields: only `notes`.
- **Quiet since:** d2633f4a4 (pass 41: report v35 republish, builds.json matches the reference on all 66) and b389688f7 (Wowhead scraper only).
- **Next pass starts from 1e2f7ec5b (pass 44, quiet; b803e8568 audited, no move).**
- **Client build:** 1.60.1.70450 (spells, talents, racials and enchants identical to 70425). The report is v35, published from 437b92161.
- **Base presets** (180 s, 20k, seed 1; the full 66 are in `orchestration/audit/ref/results/req_summary.txt`):

| Spec | DPS | Spec | DPS |
|---|---|---|---|
| Demonology | 1177.7 | Ret Shock | 901.7 |
| Lone Wolf | 1094.1 | Enhancement (Imp GW 887.8) | 892.2 |
| Frostfire | 1082.5 | Ret | 889.2 |
| Fire | 1067.0 | Ret sub Holy | 883.1 |
| Affliction | 1066.6 | Balance | 872.8 |
| Destruction | 1047.4 | Frost | 843.3 |
| Marksmanship | 1045.3 | Balance Wrath | 825.9 |
| Destruction (Fire) | 1019.2 | Elemental | 771.8 |
| Fury | 989.8 | Smite | 765.4 |
| Arms | 977.2 | Feral | 752.9 |
| Beast Mastery | 970.7 | Rogue Swords | 726.4 |
| Fury 2H | 970.1 | Rogue Daggers | 725.9 |
| Arcane | 964.8 | Seal Fate Daggers | 687.1 |
| Shadow | 960.7 | Mutilate | 684.2 |
| Survival | 935.4 | Seal Fate Swords | 678.1 |
| Shadow+Holy | 917.1 | Subtlety | 656.3 |

## Open items

- **No open findings from the results audit.**
- **Pending assumptions with no knob yet**, owned by the classes and the assumptions audit, not by this one:
  `level_partial_resist`, `pal_soc_ppm`, `lock_pet_crit`, `lock_stone_with_oil`, `bleed_tick_crit`, `warr_dt_rage_mods`,
  `warr_dt_rage_k` (`orchestration/spec_assumptions.md`). When one gets a knob or a ruling, expect moves in the classes it names.

## Settled: don't re-flag

The rulings and their evidence are in `orchestration/standards.md` and the "Settled" section of `orchestration/spec_assumptions.md`.
- **Standards 21:** normalised Rage with level scaling. Off-hand Rage ×0.5 (user ruling 2026-10-16, from a beta combat log).
- **Standards 22:** weapon type and speed per preset, capped at Phase 1 BiS speeds. This made the fast-weapon Seal of Righteousness question moot.
- **Standards 23:** raid duties. Shamans drop their own totems; Sunder Armor opener; Scorpid Sting labelled defensive (−2% hit to the boss).
- **Standards 24 and 25:**
  - raid hits of 1500 Fire damage, 2 a minute;
  - Forever's Enrage: 30% chance on any damaging hit taken, +2% physical damage a point (checked against client SpellAuraOptions for 12317: ProcChance 30, ProcTypeMask 139944);
  - Grace of Air in melee raids;
  - Windfury Totem on the main hand only.
- **Client and sim changes:** the client 70425 ports; no level-based partial resist (#102); channel pushback (ab517bf); Salvation in every request (#108); Concentration Aura for Ret at 60 (#106).
- **Shaman:** Enhancement gets the raid's Flametongue Totem (`sham_ftt_stack`, one value).
- **Paladin:** Seal of Righteousness counts spell power once, flat per swing (#54).
- **Mage and hunter:** Ice Lance coefficient 0.429; hunter Trueshot Aura and Hunter's Mark at top ranks.
- **Warrior:** Rend's attack power share.
- **Rogues are low, and it isn't a modelling error.** Audited 2026-10-12 and rechecked since:
  - Forever's rogue nerfs cost about 59 DPS.
  - Warriors have Execute and Heroic Strike, and Rage that tops them up; rogues have nothing comparable.
  - Only Subtlety uses Rupture, set up by Hemorrhage (+15% to the rogue's own Rupture under Forever).
  - Combat without Rupture is right: Rupture at 5 combo points is 14.5 worse, and at 3 it is 23 worse. A 5-point Rupture is about 1,740 damage over 16 s, against about 1,970 at once from Eviscerate.
  - Rupture's base damage matches the client. Its attack power share (Classic's 24% at 5 points) is server-side and unverified, but no plausible value makes Rupture win for Combat.

## Explaining the ranking (the other half of the role)

When numbers move, explain the new order and not just the deltas. Account for each gap by naming its causes: talents and their size, spell
coefficients and base damage, gear scaling, the synthetic weapon, the raid buffs and debuffs a spec gets, resource limits, and Forever's changes against
Classic. Back each one with the breakdown (`summ.py`) and a paired sim. Uncertainties the classes already document (for example the warlock
demons' melee and stat scaling behind Demonology's lead) are context, not findings. Point to their assumption row.

## Gotchas

- `runreq.sh` reads the meta next to the request (`<id>.meta.json`). For a modified copy elsewhere, read the DPS from the out JSON (`raidMetrics.dps.avg`) yourself.
- `compare.py` and the reference use the request ids, so a renamed or added request shows as REMOVED or ADDED. Check the commit and update the reference.
- After a full pass that changes the baseline, replace `req_summary.txt` and `drift_report.txt` in `orchestration/audit/ref/results/` (or wherever the routine keeps them) and update this file's Baseline. Otherwise the next pass reports the same moves again.
- builds.json and the metas can disagree for a few hours around a republish: the metas' `published.dps` changed in pass 40, and builds.json caught up in v35. The metas are what `runreq.sh` checks.
- Check your own times and counts before sending. Two corrections I had to send: a re-check time, and "40 stale presets" that were really 35 (28 melee, 7 casters).
