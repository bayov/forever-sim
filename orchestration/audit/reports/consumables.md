# Consumables report, pass 45 (master 99bd53ada)

Result: quiet, no findings.

## Moved since f62cb797
- ui/retribution_paladin/presets.ts: new "Level 60 Shock (Ret side)" build (talents, APL, build). It uses EncounterLevel60 with DefaultConsumes (same as the twist builds; the comment cites a paired Dragonfire Potion vs Major Mana Potion result), Concentration Aura, Seal of Command. No new consumable, enchant or raid buff; gear is the shared synthetic Phase 1 set.
- tools/dps_report/assumptions_def.py: adds paladin__level_60_shock_ret to the paladin build lists (PAL_ALL, new PAL_SOC, haste GCD list). No buff-related value changed.

## Verified
- scan.sh on master: banned items, Revelation, non-synthetic trinkets, double crit elixirs all empty.
- Counts: raid totems unchanged (Grace of Air + Windfury for melee, Flametongue for enhancement). Minor-armor count 3 is the known unused ForeverDustCloudAura (scan's expected 2 is stale).
- wago latest probe: 1.60.1.70532 (already audited), no newer build.
- Class branches carry only stale copies of other classes' files; nothing new.

No open items.
