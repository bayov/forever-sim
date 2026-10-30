# Mechanics pass 50 (2026-10-30, master bbbaf2e3c)

Range audited: 36a772fd1..bbbaf2e3c. 8 code commits. TASKS.md / update ledger: build 70707 logged for the full pass, the Wowhead hidden-items decision, the raid sim M1 write-up; no `forever-update/*` merge.

## Get current

- Client: 1.60.1.70707, unchanged (the gate's build).
- Dev notes thread 2360696: still 8 posts; post 7 (edited 10-27 16:47:14) and post 8 (edited 10-27 16:46:55) unchanged. The topic JSON was rate-limited; read through `/t/2360696/posts.json` instead.
- Blue tracker: nothing after 10-29 21:00:50 (the launch-date post, no mechanics).

## Code range

| Commit | Change | Verdict |
|---|---|---|
| 74554e2e8 | Elixir of Wisdom 6 -> 3 Intellect | matches 70707: 3166 aura 29 bp 3; Lesser Intellect's 1250988 bp 6 kept. No preset uses it |
| 26262826e, a19f2e87e | Innervate waits for a mage's own Evocation (10% mana while Evocation is ready and > 90 s left); either aura ending keeps the other's full regen | scheduling, not a game value; the full-regen hand-off is correct for overlapping effects |
| 161f80b4e | `Player.untrained_spells`, `KnowsRank` / `HighestKnownRank`; shaman spells use it; Rockbiter IDs 8017-16316 | neutral when the list is empty; IDs match the client |
| 4ae4dc3c1 | Labeled random streams and reseeds for leveling sessions | leveling only; no mechanic |
| 26a471d74, 49edbab51, 7a7a0f88e | Raid sim required buffs, Auto PI / Innervate targets, Salvation as a raid effect | raid sim only; Salvation stays threat-only (verdicts row) |

## Conclusions

No new findings. Open: mechanics-20261030-1 (Tauren Endurance crit damage, user ruling), mechanics-20261027-2 (Amplify Curse / Bane of Agony); neither touched in this range.
