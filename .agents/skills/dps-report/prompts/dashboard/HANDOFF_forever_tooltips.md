# Handoff: Forever tooltips and Wowhead links on the dashboard

## User's request (verbatim)
"make sure all tooltips and item wowhead links in dashboard use new Forever tooltips (e.g. forever wowhead or forever wago) similar to our sim UI"

## How the sim UI does it (copy this behaviour)
- /home/user/w-dash/ui/core/constants/wowhead.ts: WOWHEAD_DOMAIN = 'forever', WOWHEAD_EXPANSION_ENV = 17, WOWHEAD_FALLBACK_DOMAIN = 'classic'.
- /home/user/w-dash/ui/core/wowhead.ts: how links and the nether tooltip endpoint are built (dataEnv = 16, Forever's database).
  - Page links: https://www.wowhead.com/forever/item=ID and /spell=ID.
  - Tooltip JSON: https://nether.wowhead.com/forever/tooltip/item/ID?dataEnv=16&locale=0, or whatever wowhead.ts uses.
- /home/user/w-dash/ui/core/item_tooltip_fallback.ts: synthetic items (990001+) and ids Forever's wowhead doesn't know get a tooltip from our own item database.
- Forever's wowhead DB lacks a few reworked spells (Flurry, Blade Flurry, Elemental Mastery...). For those the UI falls back to the Classic tooltip, and the dashboard should mark them "Classic text; Forever may differ".

## What to do
1. Data. The dashboard's tooltip text is built by $SCR/dash/tooltips.py (called from build_data.py) from Classic sources: wowhead_spell_tooltips.csv and wowhead_item_tooltips.csv in /home/user/w-dash/assets/db_inputs, plus the Forever wiki text. Replace that with Forever data:
   - For every spell id and item id the dashboard shows (names.json keys, gear, setup chips, consumes, enchants, racials, talents where an id exists), fetch the Forever tooltip from wowhead's nether endpoint with dataEnv 16. The outbound proxy is configured; respect rate limits and cache every response under $SCR/dash/forever_tt_cache/.
   - Parse the HTML into the existing tooltip line format. Use Classic only as the fallback, with the note.
   - Cross-check numbers against the Forever client (wago tables in db_inputs) where cheap.
   - Fix known contradictions: e.g. Curse of Recklessness should read Forever's 505 armor, not Classic's 640, and Expose Armor likewise. Where the sim's value differs from Forever wowhead's text, show wowhead's text plus a muted "Sim uses: …" line.
   - Keep the Forever wiki text for abilities only where wowhead Forever has nothing.
   - Talent tooltips can keep the talenttrees text, which comes from the Forever client.
   - Synthetic items keep their db-based tooltip.
   Report coverage: how many ids got Forever text, how many fell back to Classic, and how many have none.
2. Links. Every tooltip for a real spell or item gets a small "Wowhead (Forever) ↗" link to https://www.wowhead.com/forever/<item|spell>=ID, opening in a new tab.
   - Synthetic items, and ids Forever's wowhead lacks, get no link, or the Classic link labelled "Classic", following the sim UI's fallback.
   - Check the dashboard type's rules (Artifact read, paths ["SKILL.md"]) for external links. If they aren't allowed, show the URL as copyable text instead and tell me.
3. Page: start from the LIVE files/index.html, version 21 or later. Read it first (ArtifactData get with out_dir), back up the local $SCR/dash/page/index.html, and work on the live copy. Other sessions edit it too, so re-read and pin if_version right before you publish.

## Rules and context
- Pipeline, harness and publishing are in the earlier handoffs in SCR=$SCR: HANDOFF_specpage.md, HANDOFF_setup_active.md, HANDOFF_tooltips.md.
- Run the rd/ checks (checklist2.js, races_check.js, specswitch.js, tips2.js, setup_check.js) in plain, text and ASYNC=1 modes before publishing.
- Upload changed data files as assets and update their dataset docs (updated.by "Claude (forever-sim)").
- Snapshot rule: no session history or process text, no "shipped".
- No new artifacts. No model names in files. Don't message other sessions. Never push or touch /home/user/forever-sim; local commits on dashdata are fine.

## Deliverable
Reply briefly with:
- the coverage numbers;
- any notable Forever vs Classic text differences found;
- whether links are clickable;
- the checks result;
- 2 phone screenshots (an ability tooltip and an item tooltip, each with the link);
- the new files/index.html version.
