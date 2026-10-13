# Handoff: a much better race analysis on the Forever DPS dashboard

## User's request (verbatim)
"Race analysis should be better. For each spec seeing how each available race for the spec compares against others. Seeing overall race usefulness to dps etc. Comparing races with conditional dmg on/off (weapon spec with appropriate weapon vs without, racials such as +5% dmg vs beasts or elementals against appropriate target or something else, etc)."
The user reads the dashboard on a PHONE.

## What exists
- Dashboard artifact: https://claude.ai/artifact/<DASHBOARD_ARTIFACT_ID> ("Dashboard" type). Page: the files/index.html doc, currently version 15. Data: datasets/<id> docs pointing at uploaded JSON blobs.
- Read the type's rules first: Artifact tool, action read, url above, paths ["SKILL.md", "artifact-type/reference/design-defaults.md"].
- Local pipeline in SCR=$SCR, under $SCR/dash. Read these before starting:
  - run_build.py: main run, a race sweep at 10k iterations / 180 s / preset settings, fight lengths, sample fights;
  - build_data.py: writes data/*.json;
  - req/<id>.json + .meta.json: the 36 preset requests;
  - out/<id>.json: sim outputs;
  - page/index.html: the live page, version 15;
  - rd/: Playwright harness (build_harness.py, shoot.js, suite.sh, checklist.js, tips.js), described in HANDOFF_redesign.md and HANDOFF_tooltips.md;
  - binaries $SCR/wowsimcli etc., built with -tags=with_db from the merged local tree /home/user/w-dash (branch dashdata, local only, never push).
- Today's race data:
  - the `races` dataset: each preset on every race its class can be, 180 s, 10k fights;
  - the Races tab: a heat grid, a faction table and a race-spread chart;
  - the spec page's races chart.
- Forever racials: /home/user/w-dash/ui/core/forever/racials.json; implementations in sim/core (grep the racial names).
  Forever changed several racials:
  - weapon specializations give +1% (Orc axe, Dwarf mace) or +2% (Human sword) crit to ALL spells and attacks while that weapon type is equipped, so casters can benefit from a sword, mace or axe too;
  - Troll Beast Slaying and Dwarf Big Game Hunter: +5% damage vs Beasts;
  - Skyborne Elemental Insight: +5% vs Elementals;
  - Skyborne Wind Blessed: +1% haste;
  - Undead Touch of the Grave, Night Elf Elune's Light, Gnome Eureka!, Tauren Endurance (+1% hit), Orc Blood Fury, Troll Berserking, Skyborne Skysight / Read Ley Line, and others.
  Check how the sim models each, and which ones are conditional on weapon type or target type.
- The encounter target has a mob type (proto MobType: Beast, Demon, Dragonkin, Elemental, Giant, Humanoid, Mechanical, Undead...). Presets fight a level 63 Humanoid.

## What to build
1. Data: new sims. For every preset and every race available to its class:
   a. **Baseline:** the existing race sweep (reuse it).
   b. **Racials off:** the race's base stats only, with its racial abilities and passives disabled. If the sim has no switch for this, add a DASHBOARD-ONLY local patch in /home/user/w-dash (e.g. an env var read where racials register). This splits each race's edge into "base stats" and "racial abilities", and gives a per-racial breakdown where feasible (turn racials off one at a time).
   c. **Weapon conditionals:** for each race with a weapon racial, sim the preset with its weapon(s) swapped to that racial's weapon type at the same budget and stats, i.e. "with the racial's weapon". Casters use a sword / mace / axe main hand plus off-hand instead of a staff, if legal for the class. Also sim "racial weapon not equipped", which is the baseline when the preset's weapon doesn't match.
      Way to do it: a copy of the preset's synthetic item with a different weaponType, via a request-embedded database if the sim's RaidSimRequest supports one, or local-only db additions in the merged tree that are never pushed. Weapon damage and speed should follow a real Era Phase 1 weapon of that type for physical builds; caster weapons only need the type.
      Skip type/class combinations the class can't use.
   d. **Target conditionals:** sim every preset × race against a Beast target and an Elemental target (and Undead / Demon if any racial or talent cares). These show Beast Slaying, Big Game Hunter and Elemental Insight, plus class effects that change with target type, which you should note.
   Use 5k-10k iterations, 180 s, the same seed for every variant, and the preset's settings otherwise. Run up to 4 in parallel; this is a lot of sims, so budget it.
   Then add new datasets (keep `races` compatible). Suggested:
   - `race_variants`: build, race, variant (baseline / racialsOff / weaponMatch / beast / elemental / ...), dps, se, plus vsBest within variant;
   - `race_breakdown`: build, race, component (base stats / each racial), dps;
   - `race_summary`: race, role (melee / ranged / caster), specs available, times best, average % behind best, median rank, best specs for it, and the same per faction.
   Add the dataset build steps to build_data.py, or to a new script that build_data.py calls.
2. Page: a much better Races tab, phone first. Keep the existing tabs and every existing feature. Start from page/index.html (version 15) and make an incremental change, as the redesign did. Ideas:
   - a race overview: race usefulness cards or a ranked list (times best, average gap, best for which roles), with a faction filter;
   - a per-spec race comparison: pick a spec (or tap one in the heat grid), and see all its races as bars with ± error, the gap to best, and a stacked split into "base stats" vs "racial abilities" with each racial labelled (tooltips on tap);
   - conditional toggles at the top of the Races tab, as a segmented control:
     - target type: Humanoid (default) / Beast / Elemental;
     - weapon: as the preset / race's racial weapon / racials off.
     The heat grid, ranking and charts all update to the chosen scenario. Make clear which racial each scenario switches on (e.g. "Beast: Troll and Dwarf +5%").
   - a short plain-language finding per class, generated from the data, e.g. "Undead best for Fury; Orc within 0.4% with an axe".
   - Tooltips on every mark, consistent with the rest of the page.
   The spec detail page's races chart can link into this.
3. Snapshot rule (user ruling): the page shows only the final data, with no session history. No "round", "before/after", "previous", PR numbers, or process words like "ruling" or "orchestrator". Keep build_data.py's note-cleaning step.
4. Test locally (harness in both data modes, phone 390 px light and dark, plus desktop). Keep a feature checklist against version 15 and check for no regressions: icons, timelines, tooltips, talent trees, every tab.
5. Publish to the SAME artifact only; never create a new artifact.
   - Upload new or changed data files as assets (Artifact publish, url, asset: true, one JSON file per call).
   - Create or update their datasets/<id> docs. For new docs, write {title, description, source: {kind: "file", url, name, format: "json"}, updated: {at, by: "Claude (forever-sim)"}} following the type's rules.
   - Update files/index.html with ArtifactData set, pinning if_version (currently 15).
   - Read the docs first so you pin the right versions.

## Rules
- No model names in files. Don't message other sessions.
- Never push or touch remote branches, and never touch /home/user/forever-sim.
- Local patches in /home/user/w-dash are fine; commit them on the local dashdata branch only.

## Deliverable
Reply briefly with:
- the new datasets and their sizes;
- headline findings: best race per role and the biggest conditional swings;
- what changed on the Races tab;
- the checklist result;
- 3-4 screenshot paths for me to show the user (phone);
- the new files/index.html version.
