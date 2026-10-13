# Follow-ups after the nine class PRs

Items found late that don't affect any shipped Level 60 preset, left out of the class PRs to avoid another merge round across all nine branches.

- **Creature-type damage auras also multiply crit damage** (found by the rogue's talent verification, 2026-10-12). The client's creature-type damage auras (aura 168) raise damage only, but the shared code multiplies both DamageDealtMultiplier and CritMultiplier, so a crit comes out about 2.08x instead of 2x:
  - sim/core/racials.go (Troll Beast Slaying, 20557);
  - sim/core/racials_forever.go mobTypeDamageAura (Skyborne Elemental Insight);
  - sim/common/item_sets/item_sets_pve.go (the anti-Undead set bonuses).
  None fires against the Humanoid Level 60 boss, so no preset moves. The rogue fixed its own Murder talent (same pattern) in its PR. Hunter Humanoid/Monster Slaying (19151, 24293) do carry a second crit-damage aura, so the hunter code is correct.

- **Rage needs a beta test** (user, 2026-10-13: "rage needs to be well-tested eventually, for now we're making assumptions"). Standards 21 and `warr_rage_k`, `warr_rage_oh` and `warr_crit_rage` model Forever's Rage from beta testers' level 30 numbers and Blizzard's notes. None of it has been measured at 60. Until it is, every item below stays an assumption in spec_assumptions.md, and the dashboard shows its range. The test plan, for the highest level the beta allows, logging every Rage gain against the swing that caused it:
  1. **Main-hand white hits, normalised k.** Use a slow and a fast one-hander, and a slow and a fast two-hander. Swing a training dummy (or a mob at +3 levels) with no talents that add Rage. Check that Rage per hit = k × swing speed, independent of damage, AP and gear. The level 30 numbers are 3.46 (1H) and 4.5 (2H). Repeat at two levels to see whether k scales with level.
  2. **Off-hand factor.** Use the same weapons in the off hand, first without Dual Wield Specialization, then with each rank. The sim assumes ×1.0 before the talent; TBC's precedent is ×0.5. Find out whether the talent's off-hand Rage bonus still applies on top.
  3. **Crits.** Compare the Rage of crit and non-crit white hits from the same weapon. The sim assumes +100% (Blizzard's 2 October warrior post; the deep dive said +75%).
  4. **Glancing blows, dodges and parries.** Check whether a glancing blow gives full k × speed, and whether a dodged or parried swing refunds Rage, as the sim assumes, or gives nothing.
  5. **Non-white damage.** Check that Heroic Strike, Cleave, Execute, Deep Wounds and other ability damage give no Rage from damage dealt.
  6. **Rage from damage taken.** Blizzard: the original formula, "as if you always had 50% damage reduction from Armor", absorbs ignored. The 8 October notes add that it no longer uses the player's health, and that crits and crushing blows count. Measure Rage per hit taken at a known damage and armor. This only matters for tanks: the Level 60 DPS presets take no damage, so their numbers don't depend on it.
  7. **Bear Form.** Check whether the same rules apply to a druid in Bear Form. The druid has no DPS preset there.
  Owner: the warrior worker turns the results into sim constants and tests, and moves each measured item out of spec_assumptions.md.
