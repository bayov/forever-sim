# Results pass 108 (master ab3fff700)

Range bfbee78fc..ab3fff700. Watched changes: 92 files, mostly the warrior leveling sim (sim/warrior, sim/core hooks for leveling sessions,
tank_warrior tests), proto and db files, Phase 0 re-rank requests for warrior and hunter (Survival, MM, LW, BM). Shared run: 68 builds, 20k, seed 1.

## Run vs metas (published.dps)
29 of 68 off by more than 0.5 (38 over 0.15), largest warlock affliction_shadow_bolt +27.1; the same families as pass 107 (paladin level_60 +1.7 to +10.3,
shaman enhancement/elemental +2.6 to +4.7, warlock destruction_fire -1.0 to -3.6, Affliction +2.5 to +3.3) plus warrior arms +0.9, arms duty-sunder +1.9,
arms duty-battle-shout +1.0, fury +1.1 and balance-wrath-duty-ff +0.6. Hunters now match.

## Sim change check (paired, 20k seed 1)
Binaries 26ecec3f1 (warrior re-rank merge) and ab3fff700 on warrior__arms, warrior__fury, warrior__arms-duty-sunder: identical to the digit
(1013.01, 993.43, 965.79). So the warrior leveling work (rage from hits taken by stance in leveling sessions, ranks, Charge, big pulls) moves no
DPS request; the +1 warrior lag is meta vs request, not a sim effect.

## Drift
Same as pass 107: Survival (3) and Enhancement (4) Land Mine (results-20261126-1), feral sapper line (results-20261105-1), expected duty/encounter lines.
No new drift. Warrior and hunter requests match mksettings. Unread fields: only notes.
