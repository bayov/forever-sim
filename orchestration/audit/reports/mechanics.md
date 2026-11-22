# Mechanics pass 97 (2026-11-22, master 8b5137c6d)

Range b4b68c52e..8b5137c6d: 18 sim files moved (net diff; the older-dated commits in the range came in through
merges). No new client build (1.60.1.71037 still newest). One new Blizzard hotfix post (Nov 20). No new finding.

## Get current

| Source | Seen | Newer? |
|---|---|---|
| Client | 1.60.1.71037 (gate probe) | no |
| Hotfix thread t/2403040 | 14 posts, post 14 2026-11-21 00:18:29Z (post 1 edited 00:19:18Z) | yes, read below |
| Dev notes t/2360696 | last 2026-11-04 19:35Z | no |
| Blue tracker | only t/2403040 post 14 since 2026-11-20 18:00Z | yes, same post |
| Daily update | 601a91f6b, 7249392f2 (TASKS, ledger): PRs #133-#136 merged; PAL-HF-07, WAR-HF-02 landed | matches the code below |

Hotfix post 14 (November 20), mechanics lines:
- "Blade Flurry will now correctly cause Shank's damage to hit an additional target." The sim's Blade Flurry
  copies any ProcMaskMelee hit, and Shank is MeleeOHSpecial, so it already does. Single-target presets: no change.
- "Stormshroud Armor 3-piece bonus changed to a 3% chance (was 4%) on melee attack of restoring 18 energy (was 30)."
  c7bab3c7d matches. Client 71037 still has 23863 ProcChance 4 / 23864 30, so the change is server-side.
- Demonic Aegis snapshot fix: not modelled (no talent swaps in a sim).

## Commits triaged

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 1c80dafff | Paladin 2H Weapon Spec on the whole Holy school under Forever | Blizzard Nov 19 wording (tier 1); per-spell callers (holy_strike, seals, SoC, SoR) now get 1 under Forever, Consecration's own multiplier dropped, so nothing takes it twice; client 20111 #0 aura 79 bp 2 school mask 3 | ok; mechanics-20261120-1 already resolved |
| 30a645072 (WAR-HF-02) | Flurry not spent by Heroic Strike / Cleave under Forever | HS / Cleave carry MHSpecial so they return early; Windfury swings are white and still spend | ok; mechanics-20261118-1 stays open |
| 6de6861c5 | Shank procs poisons, enchants, chance-on-hit, Hack and Slash | client 71037 SpellMisc still has Attributes_3 0x01070000, Attributes_4 0x800000; the change rests on server push 114268 (see below) | ok on the hotfix; not baked yet |
| 1cd2c75ce | Vilerend Slicer proc 75 -> 225 | ItemEffect 99213 -> 1320808 Wound, bp 225 (Era 16405: 75) | ok, tier 1 |
| c7bab3c7d | Stormshroud 3P 3% / 18 energy under Forever | Nov 20 hotfix post | ok, tier 1 |
| abefd2f86 | Barov Peasant Caller: 3 servants, 15 s, 6 min | ItemEffect 99908 / 99909: 360000 ms (ok); category 94, 60 s (Horde) / 20 s (Alliance) shared with summon trinkets (sim has none; no preset pairs them); spell 18307 / 18308 SpellDuration 20 s vs log 15 s | ok on the logs; shared category noted in verdicts |
| buffs.go, stormstrike.go, shaman talents.go, item_sets_pve.go | comments only (70959 values for Windfury, Stormstrike 100% / 8 s, Maelstrom PPM) | code already used them | ok |

Shank: `hotfixes.py fetch 1.60.1.71037 SpellMisc` was still pulling when the pass ended, so the push 114268 rows
(SpellMisc 874521 / 874773) were not read back here; the commit quotes them and foreverchanges' "Found in the game"
says the same. Check on the next pass if the client bakes it.

## Open findings (unchanged)

mechanics-20261120-2 (Death Mark T1 5P), 20261118-2 (hasted GCD by school), 20261118-1 (Windfury spends Flurry),
20261113-2 (Weakening Disease), 20261104-1 (druid scales). None fixed on master.
