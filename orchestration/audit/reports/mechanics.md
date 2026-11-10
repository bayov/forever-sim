# Mechanics pass 72 (2026-11-10, master ca5e35c3f)

No new finding. Both open findings (mechanics-20261107-1 natural flasks, mechanics-20261104-1 druid scales) are still open: no commit touched them.

## Range
91ad19535..ca5e35c3f, one sim commit, no forever-update merges, no TASKS.md or update_ledger.md changes.

| Commit | Change | Verdict |
|---|---|---|
| bcc5d6cd4 | Maelstrom Weapon: a stack on every landed melee hit under Forever (was 4 PPM a point); a Lightning Bolt with Totem of the Storm 50%. Flametongue Weapon rank 5 cut 7 (sim/shaman/talents.go, flametongue_weapon.go) | Class-side. Beta/live logs (tier 1): 2756 of 2760 white swings, 1553 of 1555 Stormstrikes, 428 of 430 Windfury swings, at every rank. Client agrees: 408498 proc=100, ppmid=0 in 70862 (sp.py). Classic path unchanged (2 PPM a point, standards 37). FT rank 5 cut rests on one log at one weapon speed; the comment says so. No shared code touched. Not a finding; MW proc-rate open item moved to settled in verdicts.md. |

## Get current
- Client: newest 1.60 build still 1.60.1.70862 (probe 300).
- Dev notes thread t/2360696: 8 posts, newest post 10 (2026-11-04T19:35:16Z), post 1 edit 2026-11-03T03:20:46Z; unchanged.
- Blue tracker since 2026-11-08T19:47:33Z: "WoW: Forever Hotfixes - November 9, 2026" (t/2403040 post 5, Kaivax 2026-11-10T04:23:46Z). No shared mechanic, raid buff, racial or consumable change. Class items for the class workers / daily update: Mage "Arcane Concentration's activation chance now correctly splits across the cast and every bolt (was granting 10% on cast and per bolt)"; Shaman "Lightning Overload now deals the correct amount of damage for Lightning Bolt"; Feral "Natural Instinct was not applying correctly to Lacerate or Pounce". Dungeon/creature tuning (Drowned City creatures -30% damage, Deathless Marrow) is not combat-model. Other posts: PvP faction balance (t/2408781), WoW Weekly; not mechanics.
