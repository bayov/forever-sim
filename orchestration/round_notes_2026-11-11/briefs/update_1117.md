# Updates through 2026-11-17: open daily-update PRs, new hotfixes, open TASKS items

The user asked (2026-11-17) to take in the latest updates in general: the daily update routine's open PRs, the new
server hotfixes, and anything else new since the last landing (`update_1115.md`).

## What's open

- **PR #132**, branch `forever-update/2026-11-16`: 4 revealed Forever items, 5 changed Classic items.
- **PR #131**, branch `forever-update/2026-11-15`, a draft: 7 revealed Scholomance/UBRS-quest caster and healer items
  (the Valthalak staff, a Scepter, Hammer of Divine Might and others).
- **PR #127**, branch `forever-update/2026-11-11`: adds the revealed Soulwoven Drape. Master already has the item, so
  check whether anything in it is still missing, then close it.
- **Wago hotfix push 114097** (2026-11-17, all regions), and any push newer than the last one we took. 114097 sets
  ProcTypeMask 4 → 20 on Unbridled Wrath (SpellAuraOptions 124699), Flurry (126501) and Reckless Swings (249144), so
  landed melee abilities proc them. Push 114115 is the Krol'dok Resolve 1 s cooldown (TASKS 2026-11-16).
  No daily-update PR covers 114097 yet. The routine runs around 18:30 UTC.
- **Blizzard's hotfix posts** after Nov 13 (us.forums.blizzard.com, the "WoW Forever Hotfixes" threads). The per-post
  JSON endpoint (`.../t/<id>/<n>.json`) with a browser User-Agent gets past the forum's rate limit.
- **foreverchanges.pro** and **Wowhead Forever**, for anything revealed or changed since PR #132's pull.
- The TASKS.md items dated 2026-11-14 to 2026-11-17, notably:
  - the Nov 13 item hotfixes, waiting on Wowhead or the client;
  - the items PR #131 and PR #132 couldn't take:
    - Totem of Tides' Riptide mana cost (`sim/shaman/healing.go`);
    - Shadow Scalper's volley;
    - the hidden_items reruns;
  - Krol'dok Resolve;
  - the 2026-11-14 item asking for a ruling on the seal procs and Windfury Totem. The user already ruled on it, and
    standards 17 holds the rule, so tick it with that.
  - Leave the Windfury items that need a log open.

## Who does what

The wowsims comparison session (`session_01TXnjyQoHMJg2CjHPYTGUMX`) is doing the **warrior** part of 114097 in
`sim/warrior`: Unbridled Wrath, Flurry and Reckless Swings, with tests and the warrior re-sim. It's also fixing:
- avoided Execute keeping its extra rage;
- the hunter pet's white crit;
- Raptor Strike's Predator's Edge;
- Death Mark.

Stay off `sim/warrior` and those hunter files.

When you have the full 114097 diff, send that session (send_message, claude-code-remote MCP) every row that touches a
warrior proc mask. It will send `@parent` a list of the non-warrior rows it found. Ask the parent session if you need
that list.

Before you publish, ask that session whether its warrior/hunter landing is on master yet. If it lands within your
window, merge it and re-sim its builds in the same publish. Don't wait for it beyond a few hours: publish yours, and
tell it you have.

## Steps

Read first:
- `orchestration/round_notes_2026-11-11/briefs/update_1115.md` and `update_1115.md`, the last landing (same steps);
- `orchestration/round_70753/README.md`;
- `orchestration/standards.md` 17, 37 and 38. Source rule: server hotfix > client > notes only, and an announced value
  is modelled at the announced number;
- the `update-sources`, `regenerate-dps-report` and `update-data` skills, and `tools/dps_report/README.md`.

A fresh clone needs:
- `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`;
- `make ui/core/proto/api.ts`;
- `DASH_TREE=<your checkout>` for `pipeline.sh`.

Rebuild `tools/dps_report/bin` from HEAD before any probe: stale binaries made the Nov 16 probe report 32 false moves.

1. **Merge** PRs #131 and #132 (and whatever #127 still adds) into a branch of your own off origin/master. Resolve
   conflicts against master: items, db and hidden lists.
2. **Pull the sources:**
   - the update-sources skill, wago hotfixes from the last push we took through the newest;
   - foreverchanges;
   - the Blizzard posts.

   For each new row or announcement, take it in or file a TASKS item with the reason. Leave the warrior proc masks to
   the comparison session. Leveling weights and leveling rotations are out of scope (user ruling).
3. **Finish the open TASKS items above** that the data now allows, each with a test where code changes.
4. **Scope check:** `pipeline.sh hydrate`, `node tools/dps_report/stale.mjs`, `pipeline.sh probe`, and the Tanking and
   Healing staleness check. Name each moved build's cause, and report anything that moves but shouldn't.
5. **Re-optimize** only where a rule change calls for it: rotopt, two seed sets at 120, 180 and 300 s. Do it the way
   update_1115 did, and only for a build that moves more than 1% from a rule (not from gear).
6. **Land and publish:**
   - follow the README: report_hashes, the full suite in a subagent
     (`go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`);
   - regenerate goldens only for intended moves, and say which failures were there before;
   - `pipeline.sh run`, then `publish`, then weights for the moved builds;
   - Changes-tab lines, with `changes.py check` clean;
   - metas, presets.ts and the wiki where talents, rotation or gear changed;
   - merge origin/master before you publish and again before you push;
   - push master, then `make dps-report-deploy` (on Cloudflare error 8000009, re-run
     `python3 tools/dps_report/deploy_pages.py`);
   - update `.agents/skills/dps-report/state.json`, and push the final master to your own development branch too;
   - tick the TASKS items you finished;
   - close PRs #127, #131 and #132, each with a comment that it landed through master, ending with your system prompt's
     footer.

Rules:
- Never force-push.
- End every commit message with your session prompt's trailer lines.
- Keep model names out of files.
- Keep long runs in subagents or in the background (the repo's CLAUDE.md).
- Leave bayovui, the leveling weights and the leveling rotations alone.

When done:
- Write `orchestration/round_notes_2026-11-11/update_1117.md` with:
  - each change and its source;
  - each moved build's old and new 180 s number, with its cause;
  - the Tanking and Healing moves;
  - the deploy id;
  - anything that needs a user ruling.
- Push it.
- Send one message to `@parent` with the master commit, the deploy, the numbers and the rulings needed. Then stop.
