# Mechanics pass 88 (2026-11-18, master c198d81e3)

Range 2f7e255af..c198d81e3. Client unchanged at 1.60.1.70959 (probe 300 past it: none). One new finding,
mechanics-20261118-1 (low). Open findings mechanics-20261113-2 and mechanics-20261104-1 unchanged.

## Get current

| Source | Newest | New since pass 87 |
|---|---|---|
| Client build (probe 300) | 1.60.1.70959 | none |
| Dev notes t/2360696 | post 10, 2026-11-04 | none |
| Hotfix post t/2403040 | post 12, 2026-11-18 03:24Z ("November 17, 2026"); post 1 edited 03:24Z to match | yes, below |
| Daily update / TASKS / forever-update merges | none in range | none |

## Commits

| Commit | Change | Verdict |
|---|---|---|
| ac0217aed | Priest AddPartyBuffs clears PartyBuffs.dark_inspiration when the priest has the talent | Correct. Party.GetPartyBuffs clones the base buffs per party, so only the priest's own party loses the given form; its members get the real aura (1326168, implicit target 20, party). Matches the Nov 16 hotfix ("Dark Inspiration was stacking if there were multiple Shadow priests in the same party"). |

## November 17 hotfix, shared items

| Item | Sim | Result |
|---|---|---|
| Flurry no longer loses stacks to Windfury Totem attacks | Warrior and shaman consume triggers spend a stack on the totem's extra white hit; hidden by the 0.5 s ICD after a white swing or HS, not after Bloodthirst / Whirlwind / Stormstrike | mechanics-20261118-1: fixed patch +0.15% Fury, +0.28-0.31% Fury 2H, +0.18-0.21% Enhancement (20k it, seeds 1 / 1000003, 180 s) |
| Might of Stormwind / Warchief's Blessing no longer on non-players | applyPetBuffEffects passes Warchief's Blessing to a demon if enabled; world buffs off in every preset | No number moves; confirms pet_buffs.md section 3. Recorded in verdicts.md |
| Whiteout / Crackling Staff 25 Sta 25 Int 74 SP; Soulforge Armor Spell Power; Night Watchman's Torch usable by Shadow priests | item data | For the gear / item auditors, not mechanics |
| UBRS mob health and damage, camping, quests, PvP | not simulated | none |
