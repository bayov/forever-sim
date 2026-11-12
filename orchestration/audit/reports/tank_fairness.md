# tank_fairness pass 10: 2026-11-12, master 67ad7c6a4

Due: forced re-audit after the 2026-11-12 tank fixes (`git log cb0d928fd..67ad7c6a4`): Warden re-tune for the Nov 10
shaman notes, flasks re-picked after the Natural flask cut, Wizard Oil on the paladin, ranged Judgement, the tank_gear
enchants, bear Frenzied Regeneration at 70 Rage and the Max survival potion knob, Warden Balanced Healing Stream,
warrior Max survival Shield Wall at 13 s, Windfury Totem no longer procced by Heroic Strike / Cleave.
Question this round: is every tank preset at its best under the same rules?

## Method

`tools/audit/tank_fairness/bench.sh <dir> 3000 <seed> <requests>` (new: the pass 9 run.sh was never committed) on the
six scoring bosses; `summary.py <dir>` averages at each request's own weight. Variants are copies of the request with
one field changed, seed 1 and seed 2 each against the base at the same seed (paired fights). New baseline in
`ref/tank_fairness/summary.tsv`.

## Ranking (seed 1, 3000 iterations, six-boss mean, own weight)

| weight | order | scores | pass 9 |
|---|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 3810 / 3187 / 2663 / 2535 | 3970 / 3166 / 2663 / 2498 |
| Balanced | Warrior, Paladin, Bear, Warden | 3106 / 2774 / 1717 / 1184 | 3236 / 2751 / 1717 / 1134 |
| Max survival | Paladin, Warrior, Bear, Warden | 1711 / 1573 / 14 / -2385 | 1559 / 1564 / -579 / -2741 |

The warrior lost 130 to 160 at Max threat and Balanced (TPS 2031 to 1954 and 1951 to 1893): Windfury Totem no longer
procs from Heroic Strike (user ruling 2026-11-12). The bear's Max survival gained 590 (Titans, FR at 70 Rage), the
Warden +40 to +360, the paladin +20 to +150 (Supreme Power, Wizard Oil, ranged Judgement). The paladin now leads Max
survival by 138.

TPS / DTPS / healing needed at Balanced: warrior 1893 / 380 / 344, paladin 1506 / 315 / 241, bear 1425 / 587 / 474,
Warden 1343 / 673 / 665. Spreads beyond the 15% / 25% marks are the ones explained in pass 4 to 9 (the Warden's mail
gap, standards 40; the paladin's block value, finding -20261104-2); nothing new. The Anniversary logs still show
warrior and druid main tanks only, matching warrior and bear as the real-raid picks; the sim's paladin lead in
survival stays the open question of -20261104-2.

## What each preset rewards (seed 1, every preset at every weight)

| preset | Max threat | Balanced | Max survival |
|---|---|---|---|
| bear MT / Bal / MS | **2663** / 2600 / 2175 | 1506 / **1717** / 1635 | -1968 / -931 / **14** |
| paladin MT / Bal / MS | **3187** / 3182 / 3070 | 2756 / **2774** / 2730 | 1463 / 1547 / **1711** |
| Warden MT / Bal / MS | **2535** / 2411 / 2341 | 968 / **1184** / 1160 | -3732 / -2498 / **-2385** |
| warrior MT / Bal / MS | **3810** / 3760 / 3190 | 2912 / **3106** / 2786 | 218 / 1145 / **1573** |

Each preset leads its own weight; three distinct builds per spec (the paladin's Max threat and Balanced share a build
and tie at Max threat, its call 14).

## Same terms (requests side by side)

Equal for all twelve: level 60, reaction 200 ms, Engineering + Blacksmithing, raid buffs and debuffs, Kings and
improved Might, Mongoose, Phalanx, Greater Fortitude, Juju Power and Might, Dragonbreath Chili, Goblin Sapper, Rumsey
Rum. Per-objective picks (flask, food, Zanza, potion, stone or oil, enchants) differ as their workers measured them.

| check | result |
|---|---|
| Paladin without Blessing of Wisdom (every other mana spec, the Warden included, has it) | +Wisdom: -2.2 / -1.3 / -7.5 (seed 1), -2.5 / -1.2 / -2.7 (seed 2): no gain, not a finding |
| Bear + Wisdom (Balanced) | -0.8 / -0.8: nothing |
| Paladin Greater Stoneshield + potion action (open -20261105-1) | +60.7 / +174.7 / +408.8 (seed 1), +57.5 / +168.3 / +394.7 (seed 2) |
| Warden legal races (open -20261105-2), against Night Elf / Night Elf / Undead | MT Troll -57 / -62, Orc -60 / -63; Bal Troll -84 / -71, Orc -87 / -91; MS Tauren -153 / -164 |
| Warrior Elemental Sharpening Stone (raid-only) against Dense Sharpening | +21.3 / +19.7 / +11.2 (seed 1), +20.2 / +20.3 / +14.3 (seed 2): **new finding -20261112-1** |
| Same, Dense Weightstone | a further -12 |
| Warden without Mageblood (raid-only), MT / Bal | -3.3 / -5.3 (seed 1), +0.2 / -3.7 (seed 2): noise |
| Race tables (ui/dps_report/data/tank_races.json, c46062a) | warrior, paladin, bear presets top their tables by 78 to 300; Warden presets illegal (open) |

Talents and rotations are the tank_rotations pass's (pass 10 today, its four findings fixed in 67ad7c6a4); not
re-searched here.

## Findings

- New: tank_fairness-20261112-1 (medium, user ruling): warrior keeps the raid-only Elemental Sharpening Stone (+20 at
  each weight) that the paladin and bear gave up; the Warden's Mageblood (also raid-only) is noise.
- Open, re-measured: -20261105-1 paladin potion (now +58 / +172 / +402, Max survival +24%), -20261105-2 Warden race
  (-60 / -80 / -158 to go legal). Open, unchanged: -20261104-2 (user calls), -20261104-5 (legacy request).

Left behind by this pass: an empty directory `/tank_fairness` at the filesystem root (a stray build output path,
moved to $AUDIT_WORK; the safety check would not let the pass remove the directory).
