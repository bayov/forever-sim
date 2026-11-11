# Notes round 2026-11-11: every class after shaman (sim changes only)

The shaman changes landed first (`shaman.md` here, deploy f48be298). This session does the rest of the ledger's open
rows, which are sim and data changes, and re-optimizes the builds they move. It does **not** publish the report: a
separate full-run session re-sims every build afterwards, so the report changes once.

Read first:
- `orchestration/round_notes_2026-11-11/ledger.md` (Druid through Items) and the sources it names;
- `orchestration/round_notes_2026-11-11/shaman.md`, for how the shaman worker cited notes-only values;
- `orchestration/round_70753/README.md`;
- `orchestration/standards.md` 37 (announced values are modelled at the announced number) and 38 (Forever only).

A fresh clone first needs `PATH=$PATH:$HOME/go/bin protoc -I=./proto --go_out=./sim/core ./proto/*.proto`.

1. **Warlock.**
   - WL1: Wrack takes hotfix 113766's numbers.
   - WL2: Wrack no longer multiplies periodic damage. While it channels, Corruption, Siphon Life and Bane of Agony tick
     100% faster for the same total duration. Check how that interacts with Wrack's own ticks and with a DoT refreshed
     mid-channel.
   - WL3: Soul Siphon counts only Corruption, Siphon Life and Bane of Agony.
   - WL4: Imp Firebolt's low ranks.

   Then re-search the Affliction builds' talents and rotation (rotopt, the way `orchestration/round_70753/warlock.md`
   did), including whether Wrack still earns its place. Check the other warlock builds that might now want Wrack or
   lose it.
2. **Hunter.**
   - H1 and H2: Arcane Shot and Summon Hawk cost 11% of base mana at every rank (hotfix 113892).
   - H3: Hawk melee damage is about 80% higher (Blizzard's Nov 10 post; no row). No combat logs are in the container,
     so take the announced +80% on the current log-fitted swing (standards 37), and list it as notes-only with what a
     log would need to show.

   Re-check the BM/MM rotations' Arcane Shot use and the Hawk-or-other pet picks.
3. **Mage.** M1: Arcane Concentration "now correctly splits across the cast and every bolt". Read the post's exact
   words and the client's proc data, pick the reading they support, and list the other reading as a ruling for the
   user, with the DPS each gives. M2: check that Flamestrike ticks don't consume Clearcasting.
4. **Druid.**
   - D1: Nature's Swiftness also works on Wild Growth (healing sims).
   - D4: read hotfix 113860's Claw and Rake cost row and apply it if it moves a cost.
5. **Warrior.** WR1: Sweeping Strikes keeps its charges when no second target is hit. AoE only. Small; do it if it's
   simple.
6. **Items.**
   - I1 to I9: run the item and data pulls (the `update-sources` skill) so hotfixes 113886, 113944 and 113844 and the
     item flag changes reach the db. Model I3 to I5 and I9 if they're in any build or the gear search, otherwise note
     them.
   - Keep the shaman worker's hand entry for Rage of the Storm unless the pull now carries it. If it does, drop the
     hand entry.
   - Leave TASKS.md's already-tracked item work (engineering trinkets, Flamebreak Talisman, partial items) to its
     owners unless it falls out of the pull.
7. **Standards.** Amend `orchestration/standards.md` 25: since Blizzard's Nov 10 hotfix, a weapon imbue and a weapon
   totem both work, so the rule that the imbue turns the totem off is gone. Keep the user's "both totems" ruling. Say
   in the amendment that the game changed, not the ruling.
8. **Land.**
   - Rewrite `sim/testdata/report_hashes.txt` for what moved, and run the full suite in a subagent:
     `go test --tags=with_db $(go list ./sim/... | grep -v /sim/web) -timeout 40m`.
   - Packages that already failed on origin/master before you started stay as they are, but say which.
   - Update the req metas, presets.ts and the wiki for every build whose talents or rotation changed. Add the
     Changes-tab lines (`changes.py check` must list nothing uncovered). Don't run `pipeline.sh` or deploy.
   - Merge origin/master and push to master.

Rules:
- Never force-push.
- End every commit message with the trailer lines your session prompt specifies.
- Keep model names out of files.
- Keep long runs in the background or in subagents, so your own context stays small (the repo's CLAUDE.md).

When done:
- Write `orchestration/round_notes_2026-11-11/classes.md` with:
  - each change and how it was confirmed;
  - each re-optimized build's 180 s number before and after, on your own sims;
  - anything that needs a user ruling.
- Push it.
- Send one message to `@parent` with `send_message` (claude-code-remote MCP), with the master commit, the numbers and
  the rulings needed. Then stop.
