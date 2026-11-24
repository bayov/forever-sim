# tank_fairness pass 30: 2026-11-24, master eceb09019

Due: 7 watched files changed since e29cfe00f, all the Protection warrior's racials (0f567e196, d757d5dd6, 8d5fdd664,
ae99ebd00, dfd015925): the three warrior requests and level 60 APLs gain Blood Fury, Berserking, Elune's Light and
Eureka! on cooldown, and Stoneform from 20 s at Balanced and Max survival (none at Max threat); rotopt's warrior_tank
template gains racialAt / stoneformAt knobs. No change to the paladin, bear or Warden requests, tools/tankbench or the
encounters. No sims this pass (gate: sims false, no shared run).

## Same terms

| tank | preset races | racial casts in APL | how racials fire in a score |
|---|---|---|---|
| Warrior | Undead x3 | all five named (Stoneform not at Max threat) | explicit lines; Undead's are no-ops |
| Paladin | Undead x3 | none | autocastOtherCooldowns |
| Bear | Night Elf / Tauren / Night Elf | none | autocastOtherCooldowns |
| Warden | Night Elf / Undead / Night Elf | none | autocastOtherCooldowns |

Autocast fires DPS-type racials (Blood Fury, Berserking, Elune's Light, Eureka!) but never a survival cooldown in a
scoring run (`survivalCooldownDue`, major_cooldown.go:152). Forever Stoneform is registered as survival type, so the
paladin's and Warden's Dwarf race rows never cast it while the warrior's does: new finding -20261124-1 (low; about
+10 Balanced / +40 Max survival per row, from the warrior's 10k sweep in racials_1124/stoneform_i10000_s1.tsv; no preset
race moves). Everything else on the requests is unchanged since pass 29.

## Numbers

The warrior presets are Undead, so every new racial line is a no-op for them (tank_rotations pass 27 found the warrior
baseline bit-identical after the first racial lines). The Troll regen and Stoneform code in racials_forever.go touches
only Troll and Dwarf, which no preset plays. Ranking and spreads stand as pass 29 measured them (baseline
ref/tank_fairness/summary.tsv kept, not re-simmed):

## Ranking (six-boss mean, own weight, seed 1)

| weight | order | scores |
|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4232 / 3240 / 3128 / 2628 |
| Balanced | Warrior, Paladin, Bear, Warden | 3469 / 2797 / 2130 / 1272 |
| Max survival | Warrior, Paladin, Bear, Warden | 1815 / 1778 / 347 / -2252 |

The warrior's Max survival lead over the paladin is now 37 (2%), about two standard errors; finding 20261105-1's
potion (+374) would still reverse it. Warrior builds at w 1/3, 1, 3: Max threat **4232** / 3307 / 533, Balanced
4115 / **3469** / 1532, Max survival 3421 / 3019 / **1815**: each leads its own weight. Other specs as in pass 27 below.

Quiet: no new finding. Open findings unchanged (20261105-2, 20261105-1, 20261104-2, 20261104-5).

## Earlier pass 27 (method and tables still current)

### Pass 27: 2026-11-23, master 5518f4873

Due: 1 watched file changed since 7a6d1beb9: `sim/core/attack.go`, a guard so a weapon attack with no weapon in the slot
(SwingSpeed 0, e.g. a hunter with an empty ranged slot) never swings instead of coming due forever. The rest of the range
is mage re-tunes, a Phase 0 item ban (Bracers of the Fallen Son) and a warrior DPS item test. No change to
`tools/tankbench`, `tools/encounters`, the tank requests, the tank UIs, `docs/tank_*.md` or any tank class package.
Quiet: no new finding.

## Method

Built tankbench at 5518f4873 (`make proto items`, `--tags=with_db`), ran `tools/audit/tank_fairness/bench.sh` on all
requests x six scoring bosses, 3,000 iterations, seed 1, and diffed `summary.py <dir> --tsv` against the baseline: all
84 preset rows identical to the byte (the legacy `warrior_protection.json` rows, finding 20261104-5, are left out of the
baseline as before). Every tank preset wields a main hand and none has a ranged attack loop that reaches the guard, so the
change can't touch them. Baseline unchanged. Ranking and the tables below are pass 26's, still current.

## Ranking (six-boss mean, own weight, seed 1)

| weight | order | pass 26 | pass 25 |
|---|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4235 / 3240 / 3128 / 2628 | 4235 / 3246 / 3123 / 2628 |
| Balanced | Warrior, Paladin, Bear, Warden | 3475 / 2797 / 2130 / 1272 | 3475 / 2805 / 2131 / 1281 |
| Max survival | Warrior, Paladin, Bear, Warden | 1837 / 1778 / 347 / -2252 | 1837 / 1776 / 364 / -2211 |

TPS (Balanced): warrior 2075, bear 1628, paladin 1515, Warden 1382. Healing needed (Balanced): paladin 243, warrior 340,
bear 474, Warden 666.

## What a preset rewards (six-boss mean at w 1/3, 1, 3; seed 1)

| spec | Max threat build | Balanced build | Max survival build |
|---|---|---|---|
| Warrior | **4235** / 3319 / 569 | 4120 / **3475** / 1537 | 3433 / 3034 / **1837** |
| Paladin | **3240** / 2787 / 1428 | 3205 / **2797** / 1575 | 3088 / 2760 / **1778** |
| Bear | **3128** / 1972 / -1493 | 3000 / **2130** / -480 | 2390 / 1879 / **347** |
| Warden | **2628** / 1053 / -3672 | 2493 / **1272** / -2392 | 2401 / 1238 / **-2252** |

Each preset leads its own weight; the paladin's Balanced build leads its Max threat build at w 1 by 10 only (16 in pass 19).

Open findings re-checked (unchanged): paladin Major Mana Potion with no potion action (-20261105-1), Warden Night Elf /
Night Elf / Undead (-20261105-2), paladin damage taken waiting on the user's calls (-20261104-2), legacy
`warrior_protection.json` still read (-20261104-5; now also `tools/dps_report/tankheal.py`, `tools/healbench/main_test.go`).

---

Pass 19 (carried, still current):

### tank_fairness pass 19: 2026-11-18, master 2f7e255af

Due: 4 watched files changed since 5f20dfdf4: `sim/paladin/holy_strike.go` (Nov 16 hotfix: a Shieldsworn Holy Strike
strikes with the shield and no longer rolls Windfury Totem, 4e81669a9), `sim/druid/{healing,items_restoration}.go`
(Idol of Rejuvenation, Restoration only) and `sim/shaman/items.go` (Totem of Tides, Riptide only). Also in range:
Darkmoon Card Heroism / Blue Dragon hotfix values (no tank request wears either). Quiet: no new finding.

## Method

No change to `tools/tankbench`, `tools/encounters` or the tank requests. Built tankbench at 2f7e255af (`--tags=with_db`;
without the tag the synthetic items, e.g. 990701, are missing and every run panics), ran
`tools/audit/tank_fairness/bench.sh` on all 13 requests x six scoring bosses, 3,000 iterations, seed 1, and compared with
`ref/tank_fairness/summary.tsv`. Bear, Warden and warrior rows identical; only the paladin moved. Paladin Max survival
re-run on seed 2. Baseline refreshed (the 21 paladin rows).

Fairness of the hotfix: the same post names Shield Slam and Shield Bash; the warrior already treats both as off-hand
shield strikes with no Windfury roll (`sim/warrior/shield_slam.go:25`, `sim/warrior/talents.go:125`), so the two shield
tanks are on the same rule now.

## Same terms

Requests byte-identical to pass 18. The open unequal terms stand: paladin Major Mana Potion with no potion action
(-20261105-1), Warden Night Elf / Night Elf / Undead (-20261105-2), legacy `warrior_protection.json` (-20261104-5).
The Nov 16 Holy Strike fix is Blizzard's own, and the warrior already had the same rule for its shield strikes, so the paladin's loss is fair.

## Ranking (six-boss mean, own weight, seed 1)

| weight | order | pass 19 | pass 18 |
|---|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4235 / 3246 / 3123 / 2644 | 4235 / 3275 / 3123 / 2644 |
| Balanced | Warrior, Paladin, Bear, Warden | 3475 / 2805 / 2131 / 1280 | 3475 / 2843 / 2131 / 1280 |
| Max survival | **Warrior, Paladin**, Bear, Warden | 1837 / 1776 / 364 / -2242 | 1837 / 1798 / 364 / -2242 |

Paladin deltas (seed 1, same seed both builds): -29 / -38 / -22 score, TPS -13 / -14 / -15 (about 0.9%), DTPS +1,
healing needed 0 / +4 / -0.5. Per boss the score moves -22 to -76, Onyxia Max survival +60 (noise in healing needed).

Max survival, seed 2: paladin 1785.7 ± 14.0. Two-seed means: warrior 1830 (1836.7 / 1822.3, pass 18), paladin 1781.
The pass 17-18 tie is now a 49-point (2.7%) warrior lead, outside the error. It comes from the game's own fix, so it is
fair. The paladin's open potion action (-20261105-1, +374 at Max survival in pass 13) would still reverse it.

Spread unchanged in kind: Warden 38% behind the warrior at Max threat (class gap, standards 40), bear 4% behind the
paladin at Max threat and 24% at Balanced. The logs disagree as before (no paladin main tank on any boss; warriors on
all six, druids on four).

## What a preset rewards

Paladin builds at each weight (seed 1, six-boss mean): Max threat 3246 / 2788 / 1416, Balanced 3214 / 2804 / 1576,
Max survival 3092 / 2763 / 1776 (columns w 1/3, 1, 3). Each still leads its own weight; Balanced leads Max threat at
w 1 by 16 only. Other specs unchanged from pass 18.

## Open item

The warrior's talents were tuned before Unbridled Wrath procced on Heroic Strike; a re-tune may find a few more points
(tank_rotations' or the warrior worker's job, not a fairness finding).
