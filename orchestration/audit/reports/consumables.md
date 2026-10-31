# Consumables pass 59 (master d96adcce1)

No findings. Last audited master e05fd732e; client 1.60.1.70707 is still the newest build. Class branches are not ahead in any audited way (stale copies only).

Moved watched files, all checked:
- ui/*/presets.ts (eleven specs): the only change is DefaultDebuffs `sunderArmor: true` -> `sunderRamp: SunderRamp.create(...)` (standard 32, user ruling 2026-10-31). Ten specs use `{ tank: true, pullSunders: 1 }`, warrior uses `{ tank: true }` (its own pull Sunder is the DPS warrior's). Matches mksettings setSunder (ramp default, pull 1, 0 for a warrior). No Consumes, enchant, Expose Armor or duty value changed.
- tools/rotopt/mksettings/main.go: new -sunder / -sunder-pull flags, applied at level 60 only. Ramp code in sim/core/debuffs.go (450 per stack, 5 stacks = 2250, same as Expose) reaches 5 stacks at 4.5 s (tank plus one pull Sunder) or 6 s (tank only).
- orchestration/standards.md: standard 18 now notes the ramp, new 32; consistent with the presets.
- scan.sh on master: all ban sections empty; minor-armor exclusive-effect count 3 (known: Faerie Fire, Curse of Recklessness, Tallstrider Dust Cloud).

No open items.
