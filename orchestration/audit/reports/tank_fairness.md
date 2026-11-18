# tank_fairness pass 21: 2026-11-18, master a6677362f

Due: 12 watched files changed since 8219f1083. What matters for tanks: standards 27 now lets spell haste shorten a
spell's GCD to 1s (ee35da473, c2efe4ae3), and healers see a hit a reaction time late with the idle skip waking for it
(7199c8f9c, `sim/core/target_death.go`, `health.go`). The rest is the DPS Report's assumption knobs (`core.Assume`,
default values unchanged: Rend AP, rage scaling, Seal of Command PPM, Consecration / Holy Strike 31+ ranks, Hunter's Mark,
Wrath rank 8) and the `SpellMetrics()` accessor rename, all identity at defaults. No change to `tools/tankbench`,
`tools/encounters` or the tank requests. Quiet: no new finding.

## Method

Built tankbench at a6677362f (`--tags=with_db`), ran `tools/audit/tank_fairness/bench.sh` on all 13 requests x six
scoring bosses, 3,000 iterations, seed 1, and compared with `ref/tank_fairness/summary.tsv` (same seed both builds).
Warrior, paladin and bear rows identical to the byte. Only the Warden moved; baseline refreshed (its 21 rows).

## What moved (Warden, six-boss mean at its own weight, seed 1)

| preset | score | ± | TPS | DTPS | healing needed |
|---|---|---|---|---|---|
| Max threat | 2644 -> 2628 (-16) | 4.5 | -4.1 | -0.5 | +5.3 |
| Balanced | 1280 -> 1281 (+1) | 6.8 | -2.0 | +0.4 | -1.4 |
| Max survival | -2242 -> -2211 (+30) | 17.8 | -2.9 | -0.9 | -4.5 |

Per boss -38 to +60. The Warden is the only tank that casts its own heals and shocks, so it is the one the healer-wake
fix and the spell GCD rule can reach; the paladin's spells are unchanged because none of its requests carries spell
haste. Not split between the two changes (both small and fair: the GCD rule is Forever's, standards 27; the wake fix
removes a lag no player had). Each Warden preset still wins its own weight: Max threat 2628 vs 2502 / 2412, Balanced
1281 vs 1256 / 1051, Max survival -2211 vs -2383 / -3681.

## Ranking (six-boss mean, own weight, seed 1)

| weight | order | pass 21 |
|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 4235 / 3246 / 3123 / 2628 |
| Balanced | Warrior, Paladin, Bear, Warden | 3475 / 2805 / 2131 / 1281 |
| Max survival | Warrior, Paladin, Bear, Warden | 1837 / 1776 / 364 / -2211 |

Unchanged in order. Open findings re-checked against the requests (byte-identical to pass 20): paladin Major Mana Potion
with no potion action (-20261105-1), Warden Night Elf / Night Elf / Undead (-20261105-2), paladin damage taken waiting on
the user's calls (-20261104-2), legacy `warrior_protection.json` still read (-20261104-5). None fixed.

Note: this pass's scratch landed in `/tank_fairness/` (AUDIT_WORK was unset in the shell that ran it); safe to delete.

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
