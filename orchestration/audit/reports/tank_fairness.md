# tank_fairness pass 33: 2026-11-25, master 85691655a

Due: 31 watched files changed since c3bbe33. The ones that reach a tank score are 2f42f4a44 ("Tank batch: adopt
wowsims/forever's measured threat and damage-taken rules": warrior rage-gain threat, Heroic Strike / Cleave / Slam /
Thunder Clap flat threat, Shield Slam 2 x block value and 255 threat, Shield Block's last charge; bear Lacerate 10/3
and Primal Bite 3x threat, Faerie Fire 2 x level in bear forms, Blood Frenzy rage threat, Enrage armor; flat physical
reduction capped at half the hit) and 22d2cf952 (Shield Slam a main-hand special that suppresses weapon procs and the
totem). No change to the tank requests, `tools/tankbench` or `tools/encounters`.

## Method

tankbench built in the shared master worktree (/tmp/audit-cache/cli/85691655ab99/tree, `--tags=with_db`), `bench.sh`
on the 12 preset requests (plus the base `warrior_protection`) x six scoring bosses, 3,000 iterations, seed 1;
`summary.py --tsv` diffed against the old baseline; baseline refreshed (ref/tank_fairness/summary.tsv, 12 presets).
Per-ability threat on Magmadar (`tankbench -actions`, 1,000 iterations, seed 1) at this master against a tankbench
built at 0404484fb. Seed 2 for the two bear presets that tie.

## What moved (six-boss mean at own weight, Max threat / Balanced / Max survival)

| tank | old | new | why |
|---|---|---|---|
| Warrior | 4229 / 3459 / 1814 | 4097 / 3251 / 1617 | TPS -50 to -77: Shield Slam -27 TPS (damage +24 DPS), Thunder Clap 1x -25, rage-gain threat -21 (Magmadar, Balanced) |
| Paladin | 3240 / 2797 / 1778 | 3282 / 2863 / 1875 | flat physical reduction / damage-taken rules: DTPS -4 to -6, healing needed -8 to -11 |
| Bear | 3129 / 2151 / 455 | 4537 / 3342 / 1301 | TPS +46%: Lacerate 248 -> 825 TPS (10/3), Primal Bite 436 -> 654 (3x), Magmadar Balanced |
| Warden | 2600 / 1234 / -2306 | 2606 / 1244 / -2277 | within noise |

## Ranking now

| weight | order | scores |
|---|---|---|
| Max threat | Bear, Warrior, Paladin, Warden | 4537 / 4097 / 3282 / 2606 |
| Balanced | Bear, Warrior, Paladin, Warden | 3342 / 3251 / 2863 / 1244 |
| Max survival | Paladin, Warrior, Bear, Warden | 1875 / 1617 / 1301 / -2277 |

TPS (Balanced): bear 2371, warrior 1987, paladin 1531, Warden 1367. Healing needed (Balanced): paladin 233, warrior 341,
bear 468, Warden 664.

The bear going from third to first at Max threat and Balanced is a sourced mechanic: wowsims/forever's EV-0034 /
EV-0037 (ThreatLog plus combat logs, levels 38 to 60) read Lacerate at 10/3 of its damage before the form's multiplier,
and the rank 5 ticks at 10/3 x 2.15 per stack, so stacking with our 2.15 Bear Form threat is right (their
docs/evidence/EV-0037 lines 52-63, checked this pass). The warrior's loss is the same batch's logged warrior threat
rules (EV-0004, EV-0043 to 0050). Both taken as real. Spread: the Warden is 43% under the bear at Max threat and
-2277 at Max survival, the known class gap (standards 40). The paladin's lead at Max survival grew to 258 (16%).

Against the Anniversary logs (tools/encounters/*.json healing.notes): warriors main-tank every boss, druids four;
no paladin or shaman. The sim now ranks the bear first on threat, which the logs neither confirm nor contradict (they
give damage taken, not threat); the paladin leading survival still disagrees with the logs (finding -20261104-2).

## What each preset rewards (six-boss mean; rows = preset, columns = weight)

| preset | w 1/3 | w 1 | w 3 |
|---|---|---|---|
| bear max_threat | 4537 | 3378 | -102 |
| bear balanced | 4193 | 3342 | 789 |
| bear max_survival | 3279 | 2784 | 1301 |
| paladin max_threat / balanced / max_survival | 3282 / 3249 / 3120 | 2849 / 2863 / 2809 | 1546 / 1703 / 1875 |
| warrior max_threat / balanced / max_survival | 4097 / 3896 / 3197 | 3177 / 3251 / 2802 | 417 / 1316 / 1617 |
| Warden max_threat / balanced / max_survival | 2606 / 2463 / 2382 | 1037 / 1244 / 1217 | -3672 / -2414 / -2277 |

Every preset wins its own weight except the bear's Balanced: its Max threat build scores 3377.7 at w 1 against the
Balanced build's 3341.8 (seed 1, about ±10 each), and 3327.7 against 3321.0 on seed 2. The two share talents and differ
in enchants and Zanza buff, tuned when Lacerate threatened a third of what it does now. New finding -20261125-1.
The three builds remain distinct for every spec.

## Same terms

Requests unchanged since pass 32; the open unequal terms stand:
- -20261105-1 paladin potion: all three requests still Major Mana Potion, no potion action (would add about +374 at
  Max survival; the paladin leads there now anyway).
- -20261105-2 Warden races: still Night Elf / Night Elf / Undead.
- -20261124-1 Stoneform: no paladin or Warden request casts 20594.

## Findings

- New: tank_fairness-20261125-1 (low): the bear's Balanced preset no longer beats its Max threat preset at the
  Balanced weight after 2f42f4a44; needs a bear re-tune (and likely a warrior one: Shield Slam and Thunder Clap now
  threaten less per rage).
- Open, unchanged: -20261124-1, -20261105-2, -20261105-1, -20261104-2.
