# tank_fairness pass 11: 2026-11-12, master 4fb90c298
# tank_fairness pass 11: 2026-11-12, master 8306f1d

Due: `sim/druid/rake.go` changed since a0b4d3b (c380bbaef: Forever Rake's direct hit adds 6.5% of attack power,
outside the Nov 2 scale; user's ruling 2026-11-12). Quiet.

## Why no re-sim

- Rake is registered as a Cat form spell (`sim/druid/rake.go:45`, `RegisterSpell(Cat, ...)`); a bear can't cast it.
- None of the three bear requests (`tools/tankbench/req/druid_bear_*.json`) casts Rake (no spell 1822/1823/1824/9904).
- Other Rake uses in `sim/druid/talents.go` (Feral Aggression, Open Wounds, Blood Frenzy-style checks at 125, 331-337,
  353, 375) read the Rake spell or its dot only when it exists or is active, never in a bear rotation.
- No tank request, other watched file or standard changed in a0b4d3b..8306f1d.

So the tank numbers, ranking and same-terms checks of pass 10 stand unchanged; `ref/tank_fairness/summary.tsv` kept.
Open findings unchanged (20261111-1, 20261105-2, 20261105-1, 20261104-2, 20261104-5).


### Pass 10 detail (master a0b4d3b)

Due: by-hand pass after the tank changes since pass 10 (67ad7c6a4): the main-hand Windfury Totem rule (standards 17:
white main hand swings plus main-hand-flagged and on-next-swing abilities roll it, 2a0208136), the presets re-tuned on
it (ca7c42400: bear Max threat Barkskin from 20 s, bear Max survival Frenzied Regeneration at 80 Rage, Swipe filler, new
talents), warrior Dense Sharpening Stone and Warden Greater Mageblood Elixir (09a2f47ef, closing -20261112-1).
Question: is every tank preset at its best under the same rules (talents, rotation, gear, flask, elixirs, food and other
consumables, race, role picks)? Flask per preset re-checked after the Windfury rule.

## Method

New helpers: `tools/audit/tank_fairness/variants.py` writes one-field copies of each request (every other flask, food,
Zanza, potion, weapon stone or oil, Mageblood, head / legs / neck / back / chest / wrist / hands / feet enchants, the
2H and shield enchant, the class's legal races), `bench.sh` sims them on the six scoring bosses, `paired.py` compares each
with its base at the same seed (paired fights), six-boss mean at the preset's own weight. 339 requests at 3,000
iterations seed 1; the 32 with a gain of +3 or more, the races and the combinations at seed 2; the winners at 10,000
iterations on seeds 1 and 2. New baseline: `ref/tank_fairness/summary.tsv` (seed 1, 3,000).

## Ranking (seed 1, 3,000 iterations, six-boss mean, own weight)

| weight | order | scores | pass 10 |
|---|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4202 / 3187 / 2789 / 2536 | 3810 / 3187 / 2663 / 2535 |
| Balanced | Warrior, Paladin, Bear, Warden | 3426 / 2774 / 1865 / 1189 | 3106 / 2774 / 1717 / 1184 |
| Max survival | **Warrior**, Paladin, Bear, Warden | 1762 / 1711 / 191 / -2385 | 1573 / 1711 / 14 / -2385 |

The Windfury rule gave the warrior +392 / +320 / +189 (Sunder Armor, Revenge, Heroic Strike roll it again; TPS 1954 ->
2154 at Max threat) and the bear +126 / +148 / +177 (Maul; plus the re-tune). The paladin (Holy Strike, never) and
Warden (own imbue) are unchanged. The warrior now leads Max survival by 51; the paladin's open potion (+409) and the
new enchant finding (+78) would each put the paladin back ahead.

TPS / DTPS / healing needed at Balanced: warrior 2052 / 379 / 340, paladin 1506 / 315 / 241, bear 1483 / 587 / 464,
Warden 1345 / 673 / 665. The warrior's TPS is now 36% over the paladin's; that gap is the Windfury rule as logged
(standards 17: Sunder 57/428, Revenge 33/257 procs). Other spreads are the ones explained in pass 4 to 10.

## What each preset rewards (seed 1, every preset at every weight)

| preset | Max threat | Balanced | Max survival |
|---|---|---|---|
| bear MT / Bal / MS | **2789** / 2718 / 2305 | 1633 / **1865** / 1776 | -1833 / -695 / **191** |
| paladin MT / Bal / MS | **3187** / 3182 / 3070 | 2756 / **2774** / 2730 | 1463 / 1547 / **1711** |
| Warden MT / Bal / MS | **2536** / 2415 / 2341 | 973 / **1189** / 1160 | -3715 / -2489 / **-2385** |
| warrior MT / Bal / MS | **4202** / 4073 / 3382 | 3299 / **3426** / 2977 | 590 / 1485 / **1762** |

Each preset leads its own weight; three distinct builds per spec (paladin Max threat and Balanced tie at Max threat, its
call 14).

### Same terms

Raid buffs, debuffs (party Windfury Totem for all), sim options, level, professions (Engineering + Blacksmithing),
Kings + improved Might, Mongoose, Phalanx, Greater Fortitude, Juju Power / Might, chili, Sapper, Rumsey Rum: identical
on all twelve requests (checked by diff). Weapon slot now under one rule (no raid-only stones or oils).

## Flasks after the Windfury rule (seed 1, the best other flask against the preset's)

| preset | flask | best other |
### Re-sim

`go build -tags with_db -o $AUDIT_WORK/tank_fairness/tankbench ./tools/tankbench`, then
`tools/audit/tank_fairness/bench.sh <dir> 3000 1 tools/tankbench/req/<12 presets>` and `summary.py <dir> --tsv`.
bench.sh and summary.py are now committed on master (pass 8/9 named them; they existed only on branch tankheal).

Six-boss mean, seed 1, 3000 iterations, against pass 9:

| preset | score | TPS | DTPS | healing needed |
|---|---|---|---|---|
| warrior Max threat | 3970 -> 4210 (+240 ±5.5) | +5.9% | -0.3 | +2.6 |
| warrior Balanced | 3236 -> 3439 (+203 ±7.2) | +4.9% | -0.6 | -5.1 |
| warrior Max survival | 1564 -> 1757 (+194 ±12.4) | +4.9% | -0.4 | -4.8 |
| bear Max threat | 2663 -> 2780 (+117 ±4.6) | +4.2% | 0 | -2.8 |
| bear Balanced | 1717 -> 1865 (+147 ±7.7) | +4.1% | 0 | -9.8 |
| bear Max survival | -579 -> -278 (+300 ±13.0) | +4.3% | +1.1 | -20.3 |
| paladin, Warden (all 6) | identical | identical | identical | identical |

As expected from the mask: the warrior gains from Revenge, Sunder Armor and Heroic Strike, the bear from Maul. The
paladin (Holy Strike, seal procs, Judgements) and the Warden (Stormstrike) roll it on white swings only, as before.
The healing needed falls a little for the warrior and bear: with more threat the scorer lets them play more defensively.
Largest single boss-row score change: 502.

### Ranking (new baseline, ref/tank_fairness/summary.tsv)

| weight | order | scores |
|---|---|---|
| warrior MT / Bal / MS | Natural Precision / Natural Precision / Titans | Titans -212 / Titans -125 / Natural Precision -34 |
| bear MT / Bal / MS | Natural Precision / Natural Precision / Titans | Natural Accuracy -99 / Natural Accuracy -76 / Natural Precision -129 |
| paladin MT / Bal / MS | Supreme Power (all) | Natural Precision -226 / Titans -188 / Titans +6.0, -1.5 seed 2 (tie) |
| Warden MT / Bal / MS | Natural Precision / Titans / Titans | Supreme Power -8.7 / Natural Precision -34 / Natural Precision -245 |

Every flask holds. More threat did not move the warrior or bear to Titans at Balanced (-125, -82).

## Other consumables, enchants, races (3,000, seed 1, then seed 2 / 10,000 for the winners)
### What a preset rewards

| check | result |
|---|---|
| Food, Zanza, potion, Mageblood, stone / oil (every preset) | nothing over +5 at both seeds (Warden MT Tender Wolf +4.3 / -4.0, bear Bal foods +1) |
| Paladin Dense Sharpening Stone for the Wizard Oil | below the oil |
| **Enchants: head / legs Voracity Agility or Strength, neck Agility, gloves Greater Agility** | 10,000 seed 1 / 2: paladin MS +78 / +62, Bal +26 / +25; Warden MS +46 / +52, Bal +32 / +24, MT +19 / +16; warrior MT +21 / +21, Bal +6 / +6, MS (neck) +14 / +13; bear MS (gloves) +14 / +24: **new finding -20261112-2 (high)** |
| Shield Lesser Block on the warrior, Greater Stamina on the paladin; wrists, cloak, boots, chest | nothing (the pass 7 tank_gear fixes hold) |
| **Bear Max survival Tauren** | +21.7 / +41.2 (3,000), +32.2 / +12.3 (10,000): **new finding -20261112-3 (medium)**; Night Elf still best at Max threat (-150) and Balanced (-114) |
| Warrior races | Undead best everywhere by 99 (Skyborne, Max threat) to 259 (Tauren, Max survival) |
| Paladin races | Undead best (Human -73 / -134 / -272) |
| Warden legal races (open -20261105-2) | best legal Troll -59 / -67, Troll -84 / -73, Tauren -153 / -164 |
| Paladin Greater Stoneshield + potion action (open -20261105-1) | +60.7 / +57.5, +174.7 / +168.3, +408.8 / +394.7 |

Talents and rotations are the tank_rotations pass's; the cross-weight table above shows no preset beaten at its own
weight by another preset of its spec.
### Flask (finding 20261111-1)

## Paladin potion question

TASKS.md and the round ledger (orchestration/round_notes_2026-11-11/ledger.md) have no item for -20261105-1, and
docs/tank_paladin.md:312 still records Greater Stoneshield at 0.0 (measured without a potion action). It needs a worker.

### Findings

- New: -20261112-2 (high) agility enchants on head / legs / neck / gloves, paladin MS +78 (flips Max survival back),
  Warden +16 to +52, warrior +6 to +21, bear MS +14 to +24. -20261112-3 (medium) bear Max survival Tauren +32 / +12.
- Open, re-measured: -20261105-1 paladin potion (+58 / +172 / +402), -20261105-2 Warden race (-63 / -79 / -158 to go
  legal). Open, unchanged: -20261104-2 (user calls (a)-(c)), -20261104-5 (legacy request).

Note: an unset `$AUDIT_WORK` in one shell sent the first outputs to `/tank_fairness/` (the stray root directory of pass
10); they were moved to `$AUDIT_WORK/tank_fairness/p11`, and `/tank_fairness/` is empty again (the safety check keeps
its removal for a person).
