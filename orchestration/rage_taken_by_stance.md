# Rage from hits taken, by stance

**Implemented 2026-11-25 in b4cab8498:** the sim's Rage from a hit taken is the formula on the unmitigated hit times the stance's damage taken modifier only (`ForeverDamageTakenRageModsStance`, report assumption `warr_dt_rage_mods`), on every hit that gives Rage, spells and ticks included. The thin points below are To Verify lines in notes.md.

Checked 2026-11-25. Question for whoever takes this: should the raid sim's Rage from damage taken
follow the stance? Today it doesn't (`warr_dt_rage_mods` = None, standards 21).

## Bottom line

Our combat logs say it should. In plain melee hits, the Rage moves with what the stance does to the
damage: 12% less in Defensive Stance and 11% more in Berserker Stance, against the formula the sim
uses today. Battle Stance fits that formula.

| Stance | Hits | Warriors | Logs | Rage / formula, median (95% of bootstraps) | Damage / raw, after armor | Rage / formula / damage share |
|---|---|---|---|---|---|---|
| Battle | 99 | 45 | 14 | 1.017 (0.995 to 1.053) | 1.015 | 1.006 |
| Berserker | 195 | 25 | 16 | 1.109 (1.099 to 1.120) | 1.105 | 1.005 |
| Defensive | 1,995 | 94 | 19 | 0.878 (0.873 to 0.883) | 0.901 | 0.989 |

- "Formula" is `core.ForeverDamageTakenRage` with no stance: 2.5 x (1 - A) x raw / c.
- "Damage / raw, after armor" is the share of the hit the stance lets through. It shows the log
  has the stances right: 0.9, 1.0 and 1.1.
- The last column is the test. With the stance's share taken out, all three stances come to about 1.

## How it was measured

`orchestration/tools/rage_taken.py`, on all 38 Forever logs we have (2026-10-21 to 2026-11-24):

```sh
cd ~/wow-combat-logs/forever && python3 -I <repo>/orchestration/tools/rage_taken.py --more *.txt.xz
```

The logs are in a private archive, so the tables here are what can be shared.

- A sample is a creature's melee hit on a warrior, within 3.5 sec of the one before. Nothing else
  that could move Rage happened in between: no cast, no swing of theirs, no aura or energize event.
- Its Rage is the change of the rage bar between the two hits' `SWING_DAMAGE_LANDED` snapshots.
- Stances come from the stance casts and auras (2457, 71, 2458) in the log. 16 hits with no known
  stance are left out of the table above.
- The mob's level comes from its own `SWING_DAMAGE` lines.
- Only plain hits with raw damage above 20: no crit, crushing blow, block or absorb. 2,306 hits of
  2,710 samples, from 157 warriors, with 1 that moved the bar by 0 or less left out.

## By mob level

| Stance | Mobs | Hits | Warriors | Logs | Rage / formula (95%) | Damage share | Rage / formula / share |
|---|---|---|---|---|---|---|---|
| Battle | 10-19 | 38 | 14 | 4 | 1.064 (1.033-1.114) | 1.030 | 1.008 |
| Battle | 20-29 | 25 | 12 | 5 | 0.995 (0.892-1.069) | 1.013 | 1.007 |
| Battle | 30-39 | 22 | 12 | 7 | 1.012 (0.918-1.103) | 1.010 | 1.005 |
| Battle | 40-49 | 8 | 4 | 3 | 0.950 (0.896-0.985) | 1.006 | 0.954 |
| Berserker | 40-49 | 49 | 7 | 4 | 1.055 (1.005-1.084) | 1.103 | 0.988 |
| Berserker | 50-59 | 130 | 18 | 12 | 1.127 (1.113-1.142) | 1.105 | 1.014 |
| Berserker | 60-63 | 10 | 4 | 4 | 0.751 (0.663-1.056) | 1.103 | 0.683 |
| Defensive | 20-29 | 467 | 21 | 6 | 0.896 (0.875-0.907) | 0.912 | 0.956 |
| Defensive | 30-39 | 337 | 33 | 7 | 0.849 (0.835-0.866) | 0.903 | 0.965 |
| Defensive | 40-49 | 151 | 25 | 8 | 0.868 (0.851-0.884) | 0.902 | 0.978 |
| Defensive | 50-59 | 996 | 27 | 11 | 0.885 (0.877-0.893) | 0.901 | 1.005 |
| Defensive | 60-63 | 40 | 8 | 6 | 0.921 (0.879-0.932) | 0.901 | 1.023 |

## Other checks

- **Blocks don't cut the Rage.** 373 blocked hits in Defensive Stance: 0.856 (0.841 to 0.873) of the
  formula on the whole hit, while only 0.479 of the hit got through. Plain ones are 0.878.
- **Armor doesn't either.** Defensive hits where our armor takes 20 to 30% / 30 to 40% / 40 to 50% /
  50 to 60%: 0.884 / 0.895 / 0.851 / 0.886 (45, 456, 505 and 982 hits).
- **Whose hit the change belongs to.** Giving each change to the hit before instead (`--shift`)
  blurs it, so the snapshot does hold the hit's own Rage: Battle 1.049 (0.975-1.125), Berserker
  1.090 (1.039-1.130), Defensive 0.853 (0.842-0.868). The order of the stances stays the same.
- So stances differ from block and armor. The unmitigated hit times the stance's modifier fits.

## Against it, and what's thin

1. One warrior with 8 hits in Battle and 300 in Defensive has it the other way: 0.763 and 0.925.
   Another has Berserker 1.140 (56 hits) and Defensive 0.919 (61 hits). Those are the only two with
   8 hits or more in two stances, so this is mostly a comparison between warriors.
2. Berserker against level 60 to 63 mobs is 0.751, on 10 hits from 4 warriors.
3. Battle has 99 hits, and 8 of them against mobs above 40. So the claim that Battle is 1 at 50 to
   63, where the raid sim needs it, rests on the formula, not on hits here.
4. Defensive is 0.85 to 0.92 by level group, so the formula itself is off by a few percent by level.
5. Other modifiers aren't measured: Shield Wall, Death Wish, Recklessness, talents, debuffs on us.
   `ForeverDamageTakenRageModsAll` scales by all of them, and only the stances are shown here.
6. Only melee hits. The raid damage hits (standards 24) are Fire spells, and nothing here says
   whether a stance changes the Rage from those.
7. The sim's 2.5 comes from wowsims/forever's 187 melee hits. This note doesn't know which stances
   those were in, and it wasn't checked against their data.

## What it goes against

- The user's ruling of 2026-10-16 ("being more defensive won't penalize rage generation") and
  standards 21's "in any stance".
- `warr_dt_rage_mods` in `orchestration/spec_assumptions.md`, resolved to None on 2026-10-23 because
  "nothing supports our modifiers counting".
- On the other side, the user ruled on 2026-11-22 that a rule measured from a Forever log wins.

## What it would change

- `ForeverDamageTakenRageMods` in `sim/core/rage.go` picks the rule: None (today), All, or Increases
  only. A stance-only rule would be new, and is the one the evidence covers.
- Defensive Stance tanks would make about 10% less Rage from hits. A DPS warrior in Berserker Stance
  would get about 10% more from the raid damage hits, which are 2 a minute, so little changes there.
- Leveling sessions: a workspace that isn't pushed yet already uses All there. Raid sims there still
  use None.

## What would settle it

- One warrior who takes plain hits from the same mob in Battle, Defensive and Berserker Stance, 30
  or more in each, without casting anything else. A level 50 to 63 mob covers the raid case.
- wowsims/forever's log store: the same test on their logs, by stance.
