# Consumables pass 96 (master e6eaa9470)

Trigger: client build 1.60.1.70959 -> 1.60.1.71037. No watched file moved on master (813f2b1e9..e6eaa9470, 9 commits, none in watched paths).

- builddiff 70959 -> 71037: 3 consumable spell changes, none used by a Level 60 preset or level60 gear set:
  Poison Mushroom 6727 (effect misc 3 -> 226), Camper's Lure 1331725 (new), Fishing Skill +60 1331727 (new).
  forever_enchants.go: 58 enchants checked, 0 changed.
- scan.sh on master: everything above "== counts" empty. Totem counts unchanged. Minor-armor exclusive effects count is 3
  (Faerie Fire, Curse of Recklessness, ForeverDustCloudAura at debuffs.go:1170); the "expect 2" note in the pass doc
  predates the Dust Cloud aura, debuffs.go did not move this pass. Not a finding.
- Class branches: stale copies (gotcha 1), not re-scanned.

Result: no new findings. Two low findings stay open (enchant 41 name, feral mksettings Sapper).
