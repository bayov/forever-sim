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

Due: orchestration/standards.md changed since cb0d928. Standard 17 now has Windfury Totem roll on white main hand
swings and on abilities whose client spell needs a main hand weapon or comes on the next swing (Heroic Strike, Cleave,
Maul, Sunder Armor, Revenge, Execute, ...; not Stormstrike, Holy Strike, Swipe, seal procs, Judgements). User's go
2026-11-12 after the log audit; code a0b4d3be9 (`sim/core/windfury_totem_procs_forever.go`, `sim/core/buffs.go`).
No tank request changed in the range.

### Same terms

All 12 presets set the same raid buffs on the totems: Strength of Earth, Grace of Air, `totemWeaponBuff:
TotemWeaponBuffWindfury`. So the new mask reaches every tank on equal terms; which tank gains depends only on which of
its own abilities the client marks main hand. That is the ruled mechanic, not an asymmetry. Other terms unchanged since
pass 9 (open findings below).

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
| Max threat | Warrior, Paladin, Bear, Warden | 4210 / 3166 / 2780 / 2498 |
| Balanced | Warrior, Paladin, Bear, Warden | 3439 / 2751 / 1865 / 1134 |
| Max survival | Warrior, Paladin, Bear, Warden | 1757 / 1559 / -278 / -2741 |

The Max survival tie of pass 8/9 is gone: the warrior is ahead by 198. TPS spread (Balanced) widens: warrior 2046,
paladin 1500 (warrior +36%, from +30%). Within the 25% guideline the warrior's lead was already explained (Forever's
Revenge with Heroic Strike and Shield Slam); Windfury on those abilities adds to it by ruling. The order is the one
pass 8 checked against the Anniversary logs; not re-checked this pass.

### What a preset rewards

Cross-weight means from the same runs (each file prints all three weights): every preset wins its own weight. Warrior
4210 / 3439 / 1757 against its siblings' best 4069 / 3332 / 1548; bear 2780 / 1865 / -278 against 2718 / 1771 / -695.
Paladin Max threat and Balanced still share one build (3166.0 / 3165.5).

### Flask (finding 20261111-1)

Warrior Max survival with Flask of the Titans in place of Natural Precision, seed 1, 3000: 1758.3 ±12.8 against
1757.3. Titans' gain there fell from +47 to +1; more threat from Windfury makes Precision's expertise pay at w=3 too.
The paladin with Titans (1724, pass 8 measurement; the paladin is unchanged) is now behind the warrior either way.
Noted under the finding; bear, paladin and Warden parts of it unchanged.

### Findings

No new findings. Open findings 20261111-1 (new evidence above), 20261105-2, 20261105-1, 20261104-5, 20261104-2 are
unchanged on master.

Left for the user: an empty directory `/tank_fairness` (a build.log, bench.log, summary.tsv stub) made by a build run
with `$AUDIT_WORK` unset; the safety check blocked removing it.
