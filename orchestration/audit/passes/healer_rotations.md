# Pass: healer rotations and mana

Follow `orchestration/audit/PASS.md` around this doc. Written 2026-11-03, as `rotations.md` is for the DPS specs.

You audit every healer preset's rotations (`ui/restoration_shaman`, `ui/holy_paladin`, `ui/restoration_druid`,
`ui/healing_priest` apls, and `tools/healbench/req/*.json`, which embed them) for healing they leave on the table
and mana they waste, on `tools/healbench`, the healing sim's bench (`docs/healing_sim.md`, The score in the bench).

## Rules

- Don't change rotations; file findings with the A/B numbers. Paired runs: same seed (healbench's default), the
  same requests apart from the one line, 1,000 iterations a row, 4,000 to settle anything under 0.02 carry.
- The user's rulings bind the rotations: **bound ranks** (at most 4 ranks of a heal, each on its own condition;
  no action that picks the exact rank), assignments plus smart targets, and the score (damage carried, tank
  deaths under 2%, others under 5%, at the encounter's own length). A gain that breaks one isn't a finding.
- Judge every change at the boss files' own lengths (`-encounter tools/encounters/<boss>.json`, Anniversary set)
  in the role the preset is for (`-role tank` for tank healers, `-role raid` for raid healers), and at 180 and
  300 s (`-dur`) for the mana side. A change that wins one and loses the other is a finding only with both
  numbers and a proposal for which preset should take it.

## Setup

```sh
source tools/audit/setup.sh
go build -tags=with_db -o $AUDIT_WORK/healbench ./tools/healbench
H=$AUDIT_WORK/healbench
$H -role tank -encounter tools/encounters/onyxia.json -carry -iters 1000 tools/healbench/req/<request>.json
```

## The pass

1. **What changed** since `last.master`: the healer APLs and requests, `sim/**` heals, talents and the engine
   (heal targets, heal comm, mana), `tools/encounters/*.json`, `tools/healbench/main.go`.
2. **The baseline**: every healer request on every boss file in its role, 1,000 iterations, `-carry`. Compare it
   with `ref/healer_bench/summary.tsv` (carry, EHPS, overheal, mana left a row). A row that moved more than 0.02
   carry without a rotation change is a finding for the change that moved it (or for `healer_mechanics`).
   Then write the new `ref/healer_bench/summary.tsv`.
3. **Each rotation line**: does it do what its note says, in game terms (standards 31: no timing from a
   coincidence of fight length)? Is a cooldown wasted (Nature's Swiftness, Divine Favor, Inner Focus, Innervate,
   Mana Tide, potions and runes through `autocastOtherCooldowns`)? Does the healer go out of mana before the end
   at 180 and 300 s, or end with most of it unspent where a bigger rank would carry more? Is overheal high where a
   smaller bound rank or a cancel would save mana? Try the one-line changes that test these and report the ones
   that win.
4. Report: the baseline table, the changes tried with their numbers, new findings.
