# Consumables pass 55 (master 36a772fd1, client 1.60.1.70707)

Result: findings (1 low).

## Moved files (d50909a37..36a772fd1)
- sim/core/buffs.go: Furious Howl now level-ranked via ForeverFuriousHowlAt; level 60 gives rank 24597, floor(82+0.4*4)=83, unchanged. A hunter test checks it against the wolf's ranks.
- sim/core/debuffs.go: Expose Armor and Curse of Recklessness below the top rank get their own labels (top rank 2250 / 505 keeps the old label); new ForeverCurseOfRecklessnessAura picks the rank by level. Level 60 values unchanged.
- ui/holy_paladin and ui/restoration_druid presets.ts: talent strings only; out of scope (non-DPS specs).

## Checks
- scan.sh on master: every ban section empty; counts as before. Minor-armor exclusive effects count is 3 (Dust Cloud joined the slot earlier), a stale scan expectation, not a finding.
- Client 70685 -> 70707 builddiff: 4 consumable spell changes (Elixir of Wisdom Int 6 -> 3, Well Fed 1249521, Fengus' Ferocity, Missile Barrage), 0 Forever enchant changes (53 checked). No preset uses any. Finding consumables-20261030-1 (low).

## Open items
- consumables-20261030-1 (low).
