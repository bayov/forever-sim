# Handoff: refresh the Forever DPS dashboard with the bundled-gear-round results

## Context
- The dashboard artifact is https://claude.ai/artifact/<DASHBOARD_ARTIFACT_ID> (a "Dashboard" type artifact; the user reads it on a phone).
- It shows numbers from before today's gear changes. All nine class branches have since finished the "bundled gear round":
  - new synthetic gear under standards 1b (orchestration/standards.md on branch claude/wow-forever-dps-presets-m9u15h);
  - two synthetic trinkets per spec;
  - per-spec Stamina.
- Talents, rotations and races did not change in that round.
- Your job: re-run the dashboard pipeline on the new class heads and republish every dataset.

## Expected class heads (fetch the latest of each; use newer if present)
Branches are origin/claude/wow-forever-dps-presets-m9u15h-<class>; the shared consumes branch is ...-consumes.

| Class | Head |
|---|---|
| warrior | 7348a81 |
| paladin | 6bf522f |
| hunter | 23a266f |
| rogue | d880922 |
| priest | 10bae4e |
| shaman | 148542d |
| mage | 4230073 or later (a staff ilvl fix was pending) |
| warlock | 1750d98 |
| druid | 2444b9f |

Fairness auditor's new 180 s, best-race numbers, for verification (expect about ±2):

| Build | DPS | Build | DPS | Build | DPS |
|---|---|---|---|---|---|
| Demo | 1110.3 | Fury | 1020.2 | MM | 1006.9 |
| Destro | 984.1 | Affli | 978.4 | BM | 969.5 |
| DestroFire | 968.6 | Ret | 967.3 | Frostfire | 900.3 |
| Arms | 890.1 | Fury2H | 886.8 | Fire | 872.1 |
| SV | 857.9 | Shadow | 853.9 | Arcane | 830.1 |
| Balance | 824.8 | Enh | 796.5 | Frost | 737.5 |
| CDag | 714.3 | CSwd | 691.5 | Feral | 688.5 |
| SFDag | 683.3 | Mut | 680.6 | Smite | 676.9 |
| Ele | 674.6 | SFSwd | 656.8 | Sub | 652.8 |

Other numbers, from the class sessions:
- Lone Wolf 1015.6
- Shadow + Holy 832.0
- Balance Wrath 775.0
- Ret sub Holy 962.1, Ret Shock 959.8
- The other Enhancement / Elemental variants: take them from the shaman PR #7 body.

Every class's PR body (bayov/forever-sim #1-#9) has the full new table. Use the GitHub MCP tool mcp__github__pull_request_read (load via ToolSearch), read only. PR numbers:

| Class | PR |
|---|---|
| Priest | #1 |
| Warlock | #2 |
| Paladin | #3 |
| Warrior | #4 |
| Hunter | #5 |
| Mage | #6 |
| Shaman | #7 |
| Druid | #8 |
| Rogue | #9 |

## Environment
- Merged local tree: /home/user/w-dash, branch `dashdata`. It is local only: never push it or touch remote branches.
  - It has uncommitted local edits:
    - a DASH_NAMES aura-label dump in sim/core/environment.go (finalize), which the pipeline needs;
    - the duplicate `manaAbove` removed from tools/rotopt/priest_shadow.go (PR #1 and PR #6 both define it);
    - a regenerated assets/database/db.{json,bin}.
  - An untracked `mksettings` binary.
  - Stash or commit those locally first, then merge each class head (git fetch origin <branch>; git merge).
- Resolving conflicts:
  - Code: take the class side for that class's files.
  - tools/rotopt/run.go: the hunter side won last time.
  - Re-apply the manaAbove removal if it comes back.
  - assets/database/db.json and db.bin: REGENERATE them after merging, with the repo's own db tooling (tools/database/gen_db; see the Makefile or the class branches' commit history for the exact command). That way every class's synthetic items and trinkets (tools/database/forever_synthetic_items_<class>.go), and the Shaman/Warlock spell renames in tools/database/overrides.go, are all in. Check that no class's synthetic item ids are missing.
- Re-apply the DASH_NAMES dump in sim/core/environment.go if the merge loses it. Look at `git stash show -p` or the current diff before merging.
- Rebuild into the scratchpad with -tags=with_db: wowsimcli, mksettings, rotopt, dashstats (tools/dashstats). Generated protos need protoc with ~/go/bin on PATH.
  Scratchpad: SCR=$SCR. Binaries go in $SCR/.
- 4 CPUs.

## Pipeline (scripts in $SCR/dash; read them first)
1. Requests. $SCR/dash/req/<id>.json plus .meta.json exist for all 36 builds; AGENT_BRIEF.md explains how they were made.
   Only gear changed, so update each request's player equipment (and anything else the class changed, e.g. enchants, trinkets, or a presets consume) from the build's gear set in ui/<spec>/gear_sets (meta.gearSet).
   Easiest: re-run $SCR/mksettings for the build and copy the equipment block, or rebuild the request the same way it was built.
   Then verify with `python3 $SCR/dash/verify.py REQ.json` against the new published 180 s number at the preset race (±2 DPS).
   Update meta.published (dps, overall / horde / alliance) and meta.verify from the PR bodies. Update the summary, caveats or consumes text where the PR changed them (e.g. "no trinkets" caveats are now wrong: each build has two synthetic trinkets).
2. Move the old $SCR/dash/out and $SCR/dash/weights to *_pre1b, then:
   - run run_build.py for every id (runner.sh pattern);
   - run weights.py for every id (40k iterations each: the long step, so run them in the background in parallel up to 4 at a time);
   - run build_data.py.
   build_data.py now also renames three spells (Rage of the Farseer, Bane of Agony, Bane of Doom) and builds the `tooltips` dataset (tooltips.py). Gear tooltips should pick up the new synthetic trinkets from db.json; check they show stats and the on-use effect, adding an "Use:" line from the class's sim/<class>/forever_synthetic_trinkets.go if needed.
3. Page text. The Method tab and any static notes in $SCR/dash/page/index.html, plus the `notes` dataset, describe the gear rules. Update them to the bundled round:
   - Forever budget per archetype (standards 1b) with spell power at 0.86;
   - per-spec Stamina from SoD BiS;
   - two synthetic ilvl 71 trinkets per spec (a stat trinket and a 20 s / 2 min on-use);
   - hybrids keep hybrid stats.
   Optional: a small "before → after this round" note per build, from the out_pre1b and out files.
   Keep the page under 256 KB.
4. Test locally with the headless harness (HANDOFF_tooltips.md in $SCR describes it; harness builders are in $SCR/dash/page), in both plain and stringified-data modes, at 375 px.
5. Publish:
   - Each changed data file as an asset: Artifact tool, action publish, url = the artifact, asset: true, file_path, one JSON file per call.
   - Then update the datasets/<id> docs (source.url and updated.at) and files/index.html with ArtifactData. Load it via ToolSearch "select:ArtifactData", read each doc's version first, and pin if_version. Use one batch where possible.
   - The type's rules: Artifact action read with paths ["SKILL.md"]. Don't render or screenshot the published dashboard.

## Rules
- No model names in any file. Don't message other sessions. Don't push anything.
- Never touch /home/user/forever-sim's working tree or any remote branch.

## Deliverable
Reply briefly with:
- every build's new 180 s DPS as verified against published (flag gaps over 2);
- what changed on the page;
- the new files/index.html version;
- anything that broke.
