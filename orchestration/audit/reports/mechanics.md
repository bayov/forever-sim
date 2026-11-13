# Mechanics pass 79 (2026-11-13, master 97da1cca5)

One new finding (low): mechanics-20261113-1, Imp Firebolt ranks 2-4 changed in client 70938 and the sim keeps
the 70888 values (leveling only). mechanics-20261112-2 (Unsealed Wrath echoes from Windfury Totem swings) is
resolved by 33b13e310 / a1a1a81bb. mechanics-20261112-1 (Holy Strike and the totem) and 20261104-1 stay open.

## Commits 71562a8b0..97da1cca5 (sim code, net)

| Commit | What | Verdict |
|---|---|---|
| 33b13e310 (PR #128 via e8089abc3) | Unsealed Wrath: a Windfury Totem swing stores no echo, still spends other seals' stacks (`unsealed_wrath.go:85`) | Matches t/2403040 post 8 ("no longer generates echoes from Windfury Totem attacks"). Spending is a report assumption. Resolves -2 |
| a1a1a81bb | Under Forever no extra attack rolls Windfury Totem (`buffs.go` rollsOn uses `AutoAttacks.IsExtraAttack`); stored extra attacks keep their source (`attack.go` extraSourcesStored) | Standards 17 (user log audit). Verdict row updated. Stored-source fallback 21919 only when stacks are set directly. OK |
| 2fcb6fe2a | Volley: SpellFlagNoProcs | Tier 1 twice: Nov 11 hotfix, and 70938 gives Volley and its ticks SpellMisc Attributes_4 0x800000 (suppress weapon procs). OK |
| 02cf6c40e | Item reveals Nov 11-12 (item_effects.go, db) | Item data, the gear auditor's. No shared mechanic |

## Getting current

- Client: 70938 is new (gate). `wago.py diff 1.60.1.70888 1.60.1.70938 --by-spell`: 76 spell rows. Shared rows:
  none changed (no racial, raid buff, debuff, totem, consumable or trinket). SpellItemEnchantment 8203 stat
  124 -> 6 (Spirit, already in the sim). TraitDefinitionEffectPoints no change, CurvePoint +4.
- Baked hotfixes that the sim already has: Wrack (25.2, 1.75/level, 0.2002), Imp Firebolt rank 1, Nature's Swiftness
  covers Wild Growth, Idol of Ferocity mask, Volley's no-procs flag.
- Lightning Overload / Chain Lightning overload rows now equal half of their parent ranks (176.5 + 1.05/level vs
  15208's 353 + 2.1; coefficient 0.357 vs 0.714). The sim's 0.5 times the parent gives the same result. Verified.
- Class deltas: Imp Firebolt ranks 2-4 (new finding); Wild Growth coefficient 0.033 -> 0.066 (healer auditors);
  Poised reworked into a proc trigger, same 10 Energy (no sim change); hunter pet Great Stamina becomes flat Health
  (not simmed); Frost Nova, Flare, Bane of Havoc GCD category, Reagent Economy and Quick and the Dead masks (not
  DPS mechanics the presets use).
- ItemSparse: about 1040 rows changed (DmgVariance 0.2 -> 0 on some items, Bonding fixes). These are item data
  for the gear and client update auditors.
- Hotfixes: `hotfixes.py fetch 1.60.1.70938 --set mechanics` found rows, but none apply to SpellEffect,
  SpellMisc or SpellAuraOptions (`check` reports nothing to compare).
- Dev notes t/2360696: still post 10 (2026-11-04), post 1 not edited since 2026-11-03.
- Hotfix thread t/2403040: still post 9 (2026-11-12 19:10 UTC); post 1 not edited since 2026-11-12 00:47.
- Blue tracker: nothing newer than post 9 apart from a podcast (2412856, 2026-11-12 17:46).
- Daily update: TASKS.md commits only (Judgement readings join the verify list, weapon specialization wording,
  spell icons). These have no new mechanic for this auditor.
