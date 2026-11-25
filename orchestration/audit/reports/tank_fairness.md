# tank_fairness pass 32: 2026-11-25, master 0404484fb

Due: 6 watched files changed since 5536db4: `sim/core/rage.go` and `sim/warrior/stances.go` (b4cab8498: Rage from hits
taken carries the warrior stance's damage taken modifier, Defensive 0.9 / Berserker 1.1, and twice the Rage from spell
hits, periodic ticks and Physical specials), `sim/core/target.go` (Encounter `outside_raid` for the world-buff rule),
`sim/shaman/{air,fire}_totems.go` (Windfury / Flametongue Totem raid scope, comments), `orchestration/standards.md` (43,
Phase 0 gold line). Also in range since the last sims (pass 29): Arcane Intellect / Divine Spirit overwrite their stat's
elixirs (f1942793d), world buffs and Spirit of Zandalar do nothing in a raid instance, Armor Shatter / Puncture Armor share
a slot. No change to the tank requests, `tools/tankbench` or `tools/encounters`.

## Method

tankbench built in the shared master worktree (/tmp/audit-cache/cli/0404484fb68f/tree, `--tags=with_db`), `bench.sh` on
the 12 preset requests x six scoring bosses, 3,000 iterations, seed 1; `summary.py --tsv` diffed against the old
baseline. Baseline refreshed (ref/tank_fairness/summary.tsv).

## What moved (six-boss mean at own weight, Max threat / Balanced / Max survival)

| tank | old | new | why |
|---|---|---|---|
| Warrior | 4232 / 3469 / 1815 | 4229 / 3459 / 1814 | Defensive Stance's 0.9 on melee Rage, about offset by 2x on spells/specials; TPS -3 to -5 |
| Paladin | 3240 / 2797 / 1778 | identical | no Rage, no affected consumable |
| Bear | 3128 / 2130 / 347 | 3129 / 2151 / 455 | 2x Rage from spell hits and specials, no stance modifier; Onyxia +272, Magmadar +119, Broodlord +136, Vael +102 at w 3 |
| Warden | 2628 / 1272 / -2252 | 2600 / 1234 / -2306 | Arcane Brilliance overwrites its Elixir of Greater Intellect; TPS -13 to -17 |

Warden check: the Balanced request without `intellectElixir` scores 1234.2, the same as with it (seed 1): the elixir is
dead. That is the ruled Forever mechanic, already open as consumables-20261125-2 for every preset; not a tank finding.
The bear's Max survival gain (+31%) is a Rage rule from wowsims/forever's logs that the warrior shares; Bear Form has
no damage taken modifier to divide out, so no asymmetry in the rule itself (tank_mechanics owns its correctness).

## Ranking (six-boss mean, own weight, seed 1)

| weight | order | scores |
|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4229 / 3240 / 3129 / 2600 |
| Balanced | Warrior, Paladin, Bear, Warden | 3459 / 2797 / 2151 / 1234 |
| Max survival | Warrior, Paladin, Bear, Warden | 1814 / 1778 / 455 / -2306 |

Spreads: Max threat warrior +31% over the paladin; Max survival warrior +36 (2%, about two standard errors) over the
paladin, which finding 20261105-1's potion (+374) would still reverse.

## What a preset rewards (six-boss mean at w 1/3, 1, 3; seed 1)

| spec | Max threat build | Balanced build | Max survival build |
|---|---|---|---|
| Warrior | **4229** / 3308 / 542 | 4106 / **3458** / 1514 | 3407 / 3009 / **1814** |
| Paladin | **3240** / 2787 / 1428 | 3205 / **2797** / 1575 | 3088 / 2760 / **1778** |
| Bear | **3128** / 1973 / -1492 | 3001 / **2151** / -400 | 2434 / 1939 / **455** |
| Warden | **2600** / 1027 / -3693 | 2458 / **1234** / -2436 | 2375 / 1205 / **-2306** |

Each preset leads its own weight; the three builds stay distinct.

Quiet: no new finding. Open findings re-checked, not fixed: 20261124-1 (no Stoneform outside the warrior),
20261105-2 (Warden Night Elf / Night Elf / Undead), 20261105-1 (paladin potion action), 20261104-2 (user calls),
20261104-5 (legacy warrior request).

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
