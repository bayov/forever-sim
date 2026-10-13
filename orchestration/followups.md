# Follow-ups after the nine class PRs

Items found late that don't affect any shipped Level 60 preset, left out of the class PRs to avoid another merge round across all nine branches.

- **Creature-type damage auras also multiply crit damage** (found by the rogue's talent verification, 2026-10-12). The client's creature-type damage auras (aura 168) raise damage only, but the shared code multiplies both DamageDealtMultiplier and CritMultiplier, so a crit comes out about 2.08x instead of 2x:
  - sim/core/racials.go (Troll Beast Slaying, 20557);
  - sim/core/racials_forever.go mobTypeDamageAura (Skyborne Elemental Insight);
  - sim/common/item_sets/item_sets_pve.go (the anti-Undead set bonuses).
  None fires against the Humanoid Level 60 boss, so no preset moves. The rogue fixed its own Murder talent (same pattern) in its PR. Hunter Humanoid/Monster Slaying (19151, 24293) do carry a second crit-damage aura, so the hunter code is correct.
