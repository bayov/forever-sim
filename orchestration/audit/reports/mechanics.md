# Mechanics pass 70 (2026-11-08, master 9ba00c2dc)

No new finding. Both open findings (mechanics-20261107-1 natural flasks, mechanics-20261104-1 druid scales) still open: no commit touched them.

## Range

73d5bfa9e..9ba00c2dc: four sim commits.

| Commit | Change | Verdict |
|---|---|---|
| 5a870eccc | Stormstrike mark applied after the damage calc so raw matches the log | Damage and threat unchanged; no mechanic moved |
| 2a60326a1 | `sim.Roll` rounds damage / heal rolls to whole points; three probability callers (cast pushback x2, an item's spell pick) moved to RandomFloat | Tier 1 (live logs, Searing Totem 1/4 1/2 1/4). No `sim.Roll` probability use left (grep of 105 calls). Fractional-end callers (hunter pet variance, item_sets_pve 67.5-82.5) shift < 0.5 per hit. Verdict added |
| 0193721bc | Downranking penalty heals only | User ruling 2026-11-08 on live logs (tier 1). Verdict row replaced |
| 9ba00c2dc | Windfury and Flametongue Totems: the last placed holds the weapon slot for everyone; Enhancement presets / mksettings / 4 report requests drop the raid Flametongue Totem | User observation on the beta; matches the exclusive verdict. Imbue-blocked buff still taking the slot is an assumption (notes.md). Report not resimmed (results auditor's) |

## Getting current

| Source | Result |
|---|---|
| Client build | 1.60.1.70862, unchanged (gate). |
| Dev notes thread | Still 10 posts; post 10 2026-11-04T19:35:16Z, post 9 edited 19:35:27Z, post 1 03:20:46Z. Nothing new. |
| Blue tracker | Kaivax 2026-11-07T01:07:59Z, "WoW: Forever Hotfixes - November 6, 2026" (t/2403040, post 4). Missed last pass by 13 s. |
| Daily update | 9fcbbccd2 TASKS.md: item effects Nov 7-8, push 113769 trinket cooldowns, 70862 diff notes. Item work, not shared mechanics. |

## November 6 hotfixes, shared or sim-relevant lines

| Line | Sim |
|---|---|
| Orc: "some Warlock pet abilities were benefitting too much from the Orc Command racial ability" | Matches 70862's Command 20575 #1 5 -> 0, never modelled; the sim's 5% pet damage (#0) holds. Nothing to do |
| Searing Dagger "only activated by spells that have a damaging component" | Item not in the sim |
| Mage Arcane Blast duration 18 s; Warlock Soul Siphon no longer boosts Wrack; Voidwalker health +40 | Class items, left to the class auditors |
| Druid Bear Form / Ravage tooltip fixes ("no change made to damage done") | Nothing to do |
