# Consumables pass 60 (master dd053d24c)

One finding (low): consumables-20261101-1, ranged hunters' Goblin Sapper Charge is worth +0.16 at 25 yd (see findings).

Last audited master d1935640f; client 1.60.1.70707 still newest. Moved watched files, all checked:
- sim/core/consumes.go: Goblin Sapper Charge and Goblin Land Mine need DistanceFromTarget <= 10 (explosiveBlastRadius). Matches the user's change; no consumable value changed.
- Balance, Elemental, Mage: Sapper dropped from presets.ts and mksettings, both agree (30 / 20 / 25 yd).
- Shadow (incl. Smite, Holy) and Warlock: Sapper (and Smite's Land Mine) dropped, distance 5 -> 25 in presets.ts OtherDefaults and mksettings withSpec; they agree. No explosive is right: casters out of reach.
- Hunter: ranged builds FillerExplosive Land Mine -> EZ-Thro Dark Bomb in presets.ts and spec_hunter.go (agree); Survival keeps the Sapper at 5 yd. Dark Bomb +3.1 over none (20000 it, BM). Sapper only +0.16 (finding).
- scan.sh on master: all ban sections empty; minor-armor count 3 (known). No new client build; no class branch ahead.

Open: consumables-20261101-1.
