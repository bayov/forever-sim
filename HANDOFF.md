# Session handoff: shaman mechanics audit

The user moved our local Claude Code session to Claude Code cloud for a few hours (2026-10-09). This file holds everything the cloud agent needs to keep going, and the cloud agent writes its progress back here so the local session can pick up again.

**Cloud agent: read this whole file first. Before the session ends, fill in the [Cloud session log](#cloud-session-log) at the bottom, commit and push (see [Commits](#commits)).**

## What the repo is

A fork of wowsims Classic for WoW Forever (Blizzard's Classic+ realm, now in beta up to level 30). The Forever ruleset is the default (`SimOptions.ruleset`). `forever-wiki/sim-status.md` lists what the sim models under Forever.

The user plays a level 30 Orc shaman on the beta and tests mechanics there. They send screenshots of the chat combat log.

## Standing rules from the user

These come from the user's private settings and earlier sessions, which the cloud agent can't see.

### Writing style (the user's global instructions)

- Be concise. No pleasantries or filler, only technical substance.
- Never use an em dash or a semicolon. Use a plain dash, a colon or parentheses.
- Write in the user's voice: plain sentences a person would say out loud. Explaining well beats compressing hard. One idea per sentence, simple connectives ("but", "or", "when", "because").
- Avoid LLM mannerisms: fragment-colon punch lines ("Generous on purpose: ..."), clause-packing, and compressed clever verbs ("caps", "owns", "absorbs").
- Code comments: a short summary line, a blank line, then a paragraph when the story needs it. Use "we" for what the code does ("we give a generous timeout"). Ground the why in concrete scenarios with real names from the code. No comment when the code is obvious.
- Use they/them for anyone whose pronouns aren't stated.
- Report sooner. When the harness asks for a status, give a short line and continue.
- Do research directly, not through background agents.

### Working rules

- **Tooltips (verbatim):** "Also don't take all tooltips at face value, tooltips in forever beta might have bugs. Ask me before changing anything based on tooltips".
- **Client data (2026-10-09):** use https://wago.tools/db2 as a source for the latest Forever client. The newest build is 1.60.1.70291 (8 October). See [Tools](#tools).
- **Unsettled mechanics** go in `notes.md` under `# Need to Verify`, grouped by class. "Beta" marks the ones the user can test at level 30 or below. Things only testable at 60 go in `## At 60`. When one gets settled, remove its line and fix the code.
- **Rotation ideas** go in `notes.md` (top of the file), not into presets. Rerun knob or talent searches only when the user asks.
- **Imbues in the proc audit (verbatim):** "Btw, don't forget that at some point we also need to iron out what procs each ability including imbues. Just make sure we don't skip weapon imbues because we did some testing now". That's audit item 4.13.
- **Commit without asking** after each finished and verified change (see [Commits](#commits)).
- **Beta tests:** `/run` lines must be 255 characters or less. A `/run` combat log event frame shows no events on the beta client, so we read the chat combat log screenshots. A hit's whole damage is the shown amount plus the overkill. Melee crits are 2x a normal hit, spell crits 1.5x.
- **Characters:** Horde only. Orc shaman with Engineering + Leatherworking. Shaman talent searches must pin Spirit Weapons (rotopt `-require spiritWeapons`). Never drop candidate gear silently, mark doubtful items instead.
- **Searches stay fast:** stat weight runs about 2 min in total, gear chains with `-iters 1200 -gear-screen-iters 200 -gear-screen-keep 12 -confirm-iters 10000` and enchants in a separate pass at the end.

## The audit

The user's goal: audit every shaman mechanic the sim models, one by one. The plan, the status of each item, the findings and every fix so far are in `shaman_audit.md`. Read its top part for the process.

For each item:

1. Summarize what the sim does, with file references.
2. Cross-check it with wago.tools (client 70291), ForeverChanges (foreverchanges.pro), wowhead Forever, vmangos or cmangos, and the SoD sim.
3. Ask the user whether it sounds right. Fix it first if they say so.
4. After their OK, test it: a Go test in `sim/shaman/enhancement/*_test.go` (the `TestOrcShaman*` tests are the pattern), and a beta test when one helps.

Status marks: `[ ]` not started, `[?]` waiting for the user's OK, `[t]` OK'd and waiting for its test, `[x]` done.

Done so far: sections 1 to 3, items 4.1 and 4.2, and 8.24 (Elemental Weapons). `grep -n "Fixed (" shaman_audit.md` lists every fix.

There are pre-checks (summaries not yet shown to the user) for items 4.3 to 4.7, sections 5 and 6, and parts of sections 7 and 8, in the "Pre-checks" sections of `shaman_audit.md`. Section 7 items 7.2 to 7.6 (the imbues) have beta damage tests already but still need their full summary, OK and test.

## Where we stopped: item 4.3, attack table

I sent the user the summary below and they haven't answered yet. `shaman_audit.md` still shows 4.3 as `[ ]`. Mark it `[?]` and move this summary into its section when you pick it up.

**The sim has three kinds of roll:**

- **White hits, one roll** (`sim/core/spell_outcome.go` outcomeMeleeWhite): auto attacks and Windfury Totem's extra attack. From behind (the default): miss, dodge, glancing, crit, hit. In front of the target (PvP, or the "in front of target" option): miss, dodge, parry, glancing, block, crit, hit.
- **Yellow attacks, two rolls** (outcomeMeleeWeaponSpecialHitAndCrit): Stormstrike and Windfury Weapon's two attacks. The first roll is miss, dodge, then parry and block from the front, then hit. A second roll decides the crit. Yellow attacks never glance. From the front, a blocked attack can't crit.
- **Spell table** (section 5): the Flametongue, Frostbrand and Flametongue Totem procs, the shocks, Lightning Bolt and the totems. Lightning Shield orbs always hit.

**Cross-check:**

- vmangos `src/game/Objects/SpellCaster.cpp` RollMeleeOutcomeAgainst has the same white hit order. Creatures dodge from any side, and parry and block need the front. Only white swings and extra attacks can glance.
- vmangos puts the crit of yellow attacks inside the same single roll. The sim uses a second roll, like the upstream wowsims Classic sim. Against a level 63 boss with 8% miss and 6.5% dodge, our two rolls give Stormstrike about 15% fewer crits than vmangos would. I proposed keeping two rolls, and we can't tell them apart at level 30.
- **Open: Windfury Weapon.** The client (wago 70291, same as Era) makes spell 8233 "2 extra attacks" plus 46 attack power for 1.5 sec. In 1.12 server cores those are white swings that can glance. The sim rolls them yellow, like the SoD sim (`sim/shaman/windfury_weapon.go` has a TODO about it). The beta log names them "Your Windfury Weapon hit Barrens Giraffe 72 Physical" where a white swing reads "Your Melee hit", which points to yellow but doesn't settle it. Glancing would cost Windfury about 14% of its damage against a level 63 boss (40% glancing at 65% damage). It also decides whether Windfury hits use Flurry charges.

**The proposed beta test (also covers 4.4 to 4.6):**

1. Check the axe skill is 150 of 150:
   ```
   /run for i=1,GetNumSkillLines() do local n,_,_,r,_,_,m=GetSkillLineInfo(i) if n and (n:find("Axe") or n:find("Mace")) then print(n,r,m) end end
   ```
2. Fight level 32 or 33 mobs (not elite) with Windfury Weapon on and Stormstrike on cooldown. Screenshot the log until there are about 30 Windfury Weapon hits.
3. Melee hits should glance about 30% of the time on a level 32 mob and 40% on a level 33. If Windfury hits were white, 30 of them with no glancing would happen less than once in 10,000 runs. The mobs face the player, so parries and blocks can show up too.

**Next steps by outcome:**

- No glancing Windfury hits: Windfury stays yellow. Mark 4.3 `[t]`.
- Any glancing Windfury hit: Windfury Weapon rolls on the white table. Tell the user and ask before changing `windfury_weapon.go`. Then check Flurry (item 8.x) and the proc matrix (4.13) for Windfury.
- Count the melee hits too: miss, dodge, parry, block, glancing and crit for "Melee", "Windfury Weapon" and "Stormstrike", with the mob levels. They feed 4.4 to 4.6. Record them in the Findings of `shaman_audit.md`.
- The Go test for 4.3 after the OK: something like `attack_table_test.go` TestOrcShamanAttackTable. A level 60 shaman against a level 63 boss from behind over a long fight: white hits miss about 8%, dodge 6.5%, glance 40%, never parry or block. Stormstrike and Windfury Weapon never glance, and their crits come from a second roll. In front, parries (14%) and blocks show up. It should fail when the order changes or a yellow attack glances.

## Other open items

- **Flametongue Weapon cut:** at level 30 each rank hits a flat amount below the client's numbers (0, about 2 and about 5 for ranks 1 to 3, the same at every weapon speed, before Elemental Weapons). The sim has no cut (the user's call on 2026-10-08). I offered to add the measured cuts for ranks 2 and 3 only, and the user hasn't answered. Details in `shaman_audit.md` section 7 part 2 and `notes.md` Need to Verify.
- **Weapon Mastery** (+10% weapon damage): Blizzard's Mage and Shaman deep dive (2026-10-07) says it moved to baseline, but the client has no baseline spell and the level 30 character sheet shows no bonus. It's in `notes.md` Need to Verify with a beta test.
- **Stormstrike speed:** the sim uses the normalized speed plus 0.3 (3.6 for two-handers). Need to Verify has the test that would rule out the other models.
- The badge tooltip UI fix (commit d3614cc3) was only typechecked. The user may report back.

## Results from this session (2026-10-09)

- Blizzard's Mage and Shaman deep dive matches the sim and client 70291 except Weapon Mastery (above). Enhancing Totems and Improved Weapon Totems are already in the client's Grace of Air (89) and Flametongue Totem (548) values. Rage of the Farseer's 3 min cooldown is confirmed (commit bbe66f33).
- ForeverChanges' level 30 Enhancement BiS sims at 189.8 against our Level 30 + 5 preset's 199.0 (120 sec). Their items with our talents 192.0, our items with their talents 196.8. No preset changed.

## Tools

- **Client spell data:** `python3 tools/wago/spell.py <spell ID or name regex>...` prints a spell's duration, cast time, cooldown and effects from wago.tools (build 1.60.1.70291 by default, `--build` to change it). It downloads the tables to `~/.cache/wago` the first time. For other tables use `https://wago.tools/db2/<Table>/csv?build=1.60.1.70291`. Era is `?product=wow_classic_era`, where `EffectBasePoints` is the value minus 1.
- **Server cores:** vmangos `https://raw.githubusercontent.com/vmangos/core/development/src/game/Objects/SpellCaster.cpp` (melee and spell rolls) and `Unit.cpp`. cmangos is github.com/cmangos/mangos-classic.
- **ForeverChanges:** foreverchanges.pro has talents, spellbook, changes and BiS lists for client 70245 and later.
- **Tests:** `go test --tags=with_db ./sim/...`. To refresh golden results, rename the `*.results.tmp` files to `*.results`. The shaman mechanics tests are in `sim/shaman/enhancement/`.
- **rotopt** (`tools/rotopt`) runs evaluations and searches. Example: `rotopt -settings s.json -apl Level_30_5.apl.json -eval -durs 120 -iters 20000 -confirm-iters 50000`. The settings file is an IndividualSimSettings export from the UI. `tools/rotopt/mksettings` is outdated: it writes Rockbiter Weapon in consumes and Alchemy, so fix the shaman imbue (`enhancementShaman.options.shamanImbue`) and professions (Engineering + Leatherworking) by hand.

## Commits

- Commit each finished change by topic, with a one-line imperative message in plain ASCII (no em dashes or semicolons), and push `master` to origin (`git@github.com:bayov/forever-sim.git`). `git log --oneline -15` shows the style.
- Locally we use jj (`.claude/skills/cm` has the steps). In the cloud, plain git is fine.
- Other agents push to `master` too. Fetch and rebase on `origin/master` before each push, and never force push. If you can't push to `master`, push a branch named `cloud-handoff` and say so in the log.
- `beta_log.txt` holds the raw hits from earlier beta tests. A local agent has it deleted in its uncommitted working copy, and I don't know why. Ask the user before adding to it. The results themselves go in `shaman_audit.md` anyway.

## Cloud session log

Cloud agent: add an entry here as you go, and finish it before the user leaves. The local session reads only this section to catch up, so put everything in:

- What you discussed with the user, and their decisions (quote the important ones).
- Every commit, with its hash and message.
- Audit status marks you changed, and items now waiting for the user.
- Beta test results the user sent (the numbers, not just the conclusion).
- New open questions and offers the user hasn't answered.
- Anything you started but didn't finish.

<!-- Entries go below this line. -->
