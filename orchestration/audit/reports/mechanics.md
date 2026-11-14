# Mechanics pass 81 (2026-11-14, master 1bf5feb4b)

Quiet. No new findings. Open findings mechanics-20261113-2, 20261113-1, 20261112-1 and 20261104-1 stay open; nothing on
master touched them.

## Get current

- Client: newest 1.60 build is still 1.60.1.70938 (`wago.py latest --probe 300`). No new spell data.
- Blizzard: dev notes t/2360696 newest post 10 (2026-11-04, unedited since); hotfix thread t/2403040 newest post 9
  (2026-11-12 19:10 UTC). Blue posts by Kaivax, Aggrend, Gnedi, Linxy, Brenoch: nothing newer than 2026-11-12 19:10
  (a podcast link, 2026-11-12 17:46, no mechanics).
- Daily update 4ffc171e1 (TASKS.md, 2026-11-13): 70938 table pull (already read on pass 80), item reveals Nov 11-13
  (Quel'Serrar, Lion Horn, Alcor's Sunrazor, Felstriker, Runeblade, Windreaper, Gloves of Holy Might; tracked there as a
  needs-your-call list, not duplicated here), Bane of Havoc 1225228 now on the GCD (not modelled in the sim, so no
  effect), speed trinkets now share cooldown category 2609 (movement trinkets, no DPS effect). It also lists server
  hotfix SpellMisc rows after push 113814 as unread; this pass did not read them either (open item, below).

## Commits 3a86d4a6b..1bf5feb4b (sim code)

| Commit | What | Verdict |
|---|---|---|
| 2745da6ce | Stat weights: attribute step 10 (was 1) so Blessing of Kings' 10% survives the floor after multipliers | Measurement method only; no combat formula, buff value or proc changes. The floor it works around is existing engine behaviour. Golden results (`sim/testdata/report_hashes.txt`, *.results) move only in stat weights. Not a finding. |

## Open items for the next pass

- Read wago.tools hotfix SpellMisc rows for pushes after 113605 (our `hotfixes_seen.newest_push`) and diff against
  70938 SpellMisc for spells the sim models.
