# Handoff: show which buffs, debuffs and consumes actually apply, on the dashboard's consumes and buffs card

## User's request (verbatim)
"Dashboard lists many buffs / consumes that don't stack (e.g. expose+sunder, ff + cof, etc). Instead show only enabled ones, with disabled in lower opacity and tooltip explaining why"

## What to build
1. Data. For every preset's setup entry (the `setup` dataset, built by build_data.py from each request's raidBuffs / partyBuffs / individualBuffs / debuffs / consumes), decide whether it actually applies in the sim, and if not, why.
   Sources of truth, in /home/user/w-dash (merged local tree, branch dashdata, local only):
   - **Exclusive categories.** sim/core/buffs.go and sim/core/debuffs.go, with their ExclusiveEffect categories and priorities: only the strongest in a category applies. Examples:
     - Sunder and Expose Armor share the major armor category;
     - Faerie Fire and Curse of Recklessness: check whether they share a category under Forever;
     - Forever's Grace of Air and Windfury (graceOfAirWindfuryCategory);
     - Battle Shout vs Blessing of Might, if they exclude;
     - Demoralizing Shout and Roar;
     - and any others.
     Compute the winner for each preset from the actual values the sim would use at level 60 with Forever ranks.
   - **Not in Forever.** Things the sim ignores under Forever: Curse of Shadow, Sanctity Aura, any "improved" flags (Forever removed those talents), and anything else gated off in buffs.go / debuffs.go under the Forever ruleset.
   - **Measured cross-check.** Run each preset once (1 iteration, debug log, or the aura output) and confirm which setup auras are live, e.g. an aura that never activates.
   Add fields to the setup rows: `active` (bool) and `why` (a short plain sentence), e.g.
   - "Sunder Armor already removes the same 2250 armor; the two don't stack."
   - "Not in WoW Forever."
   - "Faerie Fire and Curse of Recklessness share the armor slot; Curse of Recklessness is stronger."
   Rebuild the setup dataset, and the tooltips dataset if its text should show the reason.
2. Page. On the Specs tab's consumes and buffs card (drawSetup / setupPlace):
   - active entries look as they do now;
   - inactive entries render at low opacity, about 0.35, with a small "×" or slash badge and no active frame;
   - the tooltip leads with the reason, in a muted line at the top;
   - inactive entries sort after active ones within their subsection;
   - add a small legend line ("Faded: listed in the sim settings but has no effect, tap for why") and a toggle "Hide inactive" (off by default).
3. The rest of the page stays unchanged.

## Context and rules (read the previous handoffs: HANDOFF_specpage.md, HANDOFF_races.md, HANDOFF_tooltips.md)
All in SCR=$SCR.
- Dashboard artifact: https://claude.ai/artifact/<DASHBOARD_ARTIFACT_ID>.
- IMPORTANT: files/index.html is now version 20 or later. Another session made the latest edits:
  - v18: the Talent build card merged into Talents;
  - v19: consumes and buffs grouped like the sim UI (Consumables / Raid Buffs / Party Buffs / Offensive Debuffs / Defensive Debuffs / Professions, with subsections), and the Imp badges removed;
  - v20: subsections flow side by side.
  The local $SCR/dash/page/index.html is OLDER (v17). Before anything else, read the live files/index.html (ArtifactData get, collection files, doc index.html, out_dir to save it), make that your base, and save it as $SCR/dash/page/index.html, keeping a backup of the old local file. Work on top of it, and pin if_version to the version you read when you publish.
- Pipeline and harness are in $SCR/dash (build_data.py, build_races.py, tooltips.py, rotation.py, rd/ checks: checklist2.js, races_check.js, specswitch.js, tips2.js).
  - Run the regression checks in both data modes, including ASYNC=1, before publishing.
  - Snapshot rule: no session history or process text on the page, and no "shipped" wording.
  - Publish only to this artifact. Upload changed data files as assets and update their dataset docs with updated.by "Claude (forever-sim)".
- No model names in files. Don't message other sessions. Never push or touch /home/user/forever-sim; local commits on dashdata are fine.

## Deliverable
Reply briefly with:
- the list of exclusion rules found, with which presets they affect;
- how many setup entries per preset became inactive;
- the checks result;
- 2 phone screenshot paths (light, dark) of the card with a tooltip open;
- the new files/index.html version.
