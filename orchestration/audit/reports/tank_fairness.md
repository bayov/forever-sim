# tank_fairness pass 8: 2026-11-11, master 707ae92

Due: 10 watched files changed since 27d0d59 (standards 25, shaman Nov 10 tuning, druid and warrior items). The
range since the last re-sim (pass 6) also brings the Natural flask cut (032692409, client 70862, merged in ca07df649)
and Windfury Totem's white-swing-only rule (0ce9dbf40). All 12 presets re-simmed.

Run: `go build -tags with_db ./tools/tankbench` (without the tag the database is empty: "No item with id"), then
`tankbench -iters 3000 -seed 1 -encounter tools/encounters/<boss>.json <request>` on the six scoring bosses. Table in
`ref/tank_fairness/summary.tsv`, summary by `tools/audit/tank_fairness/summary.py` (new helper).

## What moved and why

Six-boss mean at the own weight, against pass 6 (9ba00c2):

| preset | pass 6 | now | change |
|---|---|---|---|
| warrior MT / Bal / MS | 4650 / 3867 / 2151 | 3970 / 3236 / 1564 | -680 / -631 / -587 |
| paladin MT / Bal / MS | 3249 / 2826 / 1922 | 3166 / 2751 / 1559 | -83 / -74 / -363 |
| bear MT / Bal / MS | 3491 / 2500 / 233 | 2663 / 1717 / -579 | -828 / -783 / -811 |
| Warden MT / Bal / MS | 3169 / 1794 / -1994 | 2498 / 1134 / -2741 | -670 / -660 / -747 |

Attribution (Magmadar, old binary at ca5e35c vs master, 3000 iterations):
- Warrior and bear: the Natural flask cut is all of it. With Flask of the Titans in place of the Natural flask the two
  binaries give identical scores. The paladin moved least because its Max threat and Balanced take Supreme Power.
- Warden: the flask cut plus Blizzard's Nov 10 shaman tuning (Stormstrike 150% -> 100% and 6 -> 8 sec, Mental
  Dexterity 50/100/150 -> 33/66/100%, notes-only under standards 37). With Titans on both binaries, Balanced on
  Magmadar went 1978 -> 1689, TPS 1593 -> 1446. That is a real game change, not a fairness gap. But the Warden's talents
  and APL were tuned under the old values, so they need a re-check (tank_rotations / Warden worker).
- Standards 25 (imbue and weapon totem together) doesn't reach a tank: the Warden imbues Rockbiter, which never turned
  Windfury Totem off. Sweeping Strikes, Idol of Ferocity and Wild Growth don't touch a tank preset.

## Ranking now (seed 1, 3000 iterations, six-boss mean)

| weight | order | scores |
|---|---|---|
| Max threat | Warrior, Paladin, Bear, Warden | 3970 / 3166 / 2663 / 2498 |
| Balanced | Warrior, Paladin, Bear, Warden | 3236 / 2751 / 1717 / 1134 |
| Max survival | Warrior = Paladin, Bear, Warden | 1564 / 1559 / -579 / -2741 |

Balanced TPS: warrior 1950, paladin 1500, bear 1425, Warden 1380. Balanced healing needed: paladin 245, warrior 337,
bear 474, Warden 713. The bear dropped from second to third at Max threat. The Warden is still last: -56% on score and about 2× healing
needed against the others' mean at Balanced (standards 40, class gap). The paladin's low damage taken is still
finding -20261104-2. Anniversary logs: unchanged (warrior main tanks on every boss, druids on four, no paladin).

## Presets reward their weight

Every spec's preset wins at its own weight (seed 1, six-boss mean). The warrior gets 3970 / 3236 / 1564 at its own
weights, against 3878 (Balanced at 1/3), 3097 (Max threat at 1) and 1312 (Balanced at 3). The paladin's Max threat and
Balanced tie at 1/3 (3166.0 / 3165.5, its call 14). Bear and Warden separate cleanly.

## Same terms

Requests unchanged in the range. The open asymmetries stand:
- -20261105-2 (high): Warden presets on Night Elf and Undead.
- -20261105-1 (high): the paladin has no potion action.
- -20261104-2 (medium): the paladin's damage taken; waits on the user.
- -20261104-5 (low): the legacy warrior request.

New: the flasks no longer match each preset's objective.

| preset | Natural -> Titans, seed 1 / seed 2 (paired, -labeled) |
|---|---|
| warrior Max survival | +46.7 / +44.0 |
| bear Max survival | +131.6 / +131.5 |
| paladin Max survival | +145.7 / +112.3 |
| Warden Max survival | +256.0 / +245.8 |
| Warden Balanced | +17.9 / +19.7 |
| Max threat (warrior, bear, Warden), seed 1 | -190 / -159 / -65 |
| Balanced (warrior, bear), seed 1 | -116 / -68 |

## Findings

- New: tank_fairness-20261111-1 (high). Flask of the Titans beats the halved Natural flasks in every Max survival
  preset and in the Warden's Balanced. With it, the paladin's Max survival (1724) passes the warrior's (1627).
- Open item, not a finding: re-tune the Warden after the Nov 10 shaman tuning (tank_rotations).

Housekeeping: an earlier mistyped run left `/tank_fairness/` (a few failed outputs) at the filesystem root. The
safety check stopped its removal, so it needs a manual `rm -r /tank_fairness`.
