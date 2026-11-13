# Consumables pass 83, master 97da1cca5

No new findings.

- Moved since 71562a8b0 (watched files): item_effects.go (Ebon Hilt of Marduk Forever tick 28 per 3 sec, no preset use), buffs.go (Windfury Totem no longer rolls on extra attacks, standards 17), assumptions_def.py (paladin hotfix assumptions), standards.md (item 17), ui/retribution_paladin/presets.ts (talent strings and comments only; no consumes change), ui/rogue/presets.ts (comment only).
- New client build 1.60.1.70938 (was 70888). builddiff.py: one new drink, Morning Glory Dew (item 287964, no preset uses it); Bracer and Boots Lesser Spirit enchant 8203 now carries effect arg 6 where it carried 124, matching forever_enchants.go's +4 Spirit note (hotfix 113844); no preset uses either enchant. No other consumable or enchant changed.
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty. Totem counts match the rulings. Minor-armor count 3 against expected 2 is ForeverDustCloudAura (known, scan expectation stale).
- Open items: the two low findings in findings/consumables.md are unchanged.
