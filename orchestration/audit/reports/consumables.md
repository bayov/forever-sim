# Consumables pass 100 (master 2de7021bf7, client 1.60.1.71037)

Result: quiet, no new findings.

Checked (diff 5518f4873..2de7021bf, one watched file):
- ui/enhancement_shaman/gear_sets/level60_phase0.gear.json (Phase 0 reshuffle): head 252605 + 2543, feet 1887, legs 1506, plus real-item swaps (227006, 18405, 227005). Item choice belongs to the gear auditor (state/gear.json pass 72 already notes this set).
- Enchants vs client 71037 (ench.py): 2543 Attack Speed +1% (Arcanum of Rapidity, item 18329); 1506 Strength +8 (Lesser Arcanum of Voracity 11645); 1887 Agility +7 (Boots Greater Agility, spell 20023, formula 16245). All exist, none raid-only by name. 2543 was already in the set.
- Sources not confirmed: wowhead returned empty pages (srcid.py NO SOURCE for 252605, 18405, 16245, 11645); recorded as "could not check", not guessed. 18405 Belt of the Archmage is Phase 2 in tools/database/overrides.go:47 (a gear question, not a consumable).
- scan.sh origin/master: banned, Revelation, crit-elixir lists empty. NONSYN trinkets 13965 and 10455 (enhancement Phase 0), accepted as in pass 98/99. Minor-armor exclusive count shows 3 vs "expect 2", unchanged since pass 99 (debuffs.go not in the diff).
- Wago: not re-probed beyond the 71037 fetch used here.

No open items besides the two low findings in findings/consumables.md.
