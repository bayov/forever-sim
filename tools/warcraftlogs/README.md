# Warcraft Logs miner

Pulls public Warcraft Logs reports through the v2 API and answers the `notes.md` questions
that need many players' logs. These are proc cooldowns (Windfury Totem and Windfury Weapon,
items 4.11 and 4.12), proc rates (Maelstrom Weapon) and glancing blows by target. We built it
before WoW Forever's release and tested it on Season of Discovery logs (2026-10-13). The SoD
numbers only show that the tool works. They say nothing about Forever, so they never go into
`notes.md` or the sim.

## Setup

Make an API client at https://www.warcraftlogs.com/api/clients/ (any name, redirect URL
`http://localhost`, not public). Put its ID and secret in `~/.config/forever-sim/wcl.json`:

```json
{"client_id": "...", "client_secret": "..."}
```

## Going easy on the API

Every client gets 3600 points an hour. A report's fights, actors and abilities cost about 2
points, and one page of events (up to 10000) about 1. The tool spends as little as it can:

- Every response is cached in `~/.cache/forever-sim/wcl/`. Report data stays there forever,
  because a report doesn't change once uploaded. Rankings, zone lists and report listings
  are kept for a day. So we can rerun an analysis with other flags for free.
- We wait 1 sec between requests.
- Each run stops when it has spent `--budget` points (300 by default). It also stops when
  the hour has spent 90% of its points, so other runs still have room. A stopped run keeps
  what it fetched, and the next run picks up from the cache.
- A 429 (rate limited) answer stops the run. We never retry it.

Every command ends with a line like `[wcl] 29 requests, 47 from cache, 28.0 points spent this run`.

## Commands

All of them take `--site` first (the Warcraft Logs subdomain: `sod`, `vanilla`, `fresh`,
`classic`). The default is `sod`.

```sh
W="python3 tools/warcraftlogs/wclcli.py"

# Zones, encounters and ranking partitions of a site.
$W --site sod zones

# Pick fights to read. From the rankings of a class and spec on some encounters:
$W sample -o sod_enh.json --class Shaman --spec Enhancement --encounter 201118 --per-encounter 15
# Or every player of a class in the newest public reports of a zone (finds trash and
# dungeon runs too, for when a game has no rankings yet):
$W sample -o forever_war.json --zone <ID> --reports 10 --class Warrior
# Or one player by hand (leave the fight empty for the whole report):
$W sample -o one.json --report DqwpC17Z328HjdQB:116:Kim
# --append adds to an existing sample file.

# Every ability a player shows up with, by event type, to find a game's spell IDs.
$W abilities DqwpC17Z328HjdQB:116:Kim

# Gaps between procs. Two procs closer than a cooldown rule that cooldown out.
$W gaps sod_enh.json --ability 439440

# White swing outcomes on bosses (--targets all for every target), with 95% intervals.
$W glance sod_enh.json

# Procs per landed hit, per minute and as PPM.
$W procrate sod_enh.json --ability 408505 --triggers 1,17364,408507
```

The analysis commands take `--limit N` to read only the first N entries of a sample. That's
useful for a first look that costs little.

### How procs are counted

`--ability` is the proc's spell IDs and `--type` the event types that count. The default for
`gaps` is `damage,extraattacks` and for `procrate` it's `applybuff,applybuffstack,refreshbuff`.
A buff counts for the player who has it (a Windfury Totem buff from the shaman on a warrior
counts for the warrior). Everything else counts for the player who did it.

- Damage and extra attack events closer than `--merge` ms (50) are one proc. SoD's Windfury
  Weapon logs two attacks per proc, usually on the same ms but up to 21 ms apart.
- Buff events count one by one, because two procs can land on the same ms. A stack added
  logs `applybuffstack` and then `refreshbuff` 0 or 1 ms later, so we drop a `refreshbuff`
  that comes within 5 ms after a stack event. A proc at full stacks logs only `refreshbuff`,
  and that one counts.

`gaps` prints how many gaps fall under each limit, a histogram and the shortest gaps. For each
of the shortest gaps it shows what came before each proc: the player's last physical cast
(white swing or weapon strike) up to 400 ms before it. In SoD a proc lands 20 to 70 ms after
the hit that caused it. A Windfury Weapon cooldown shows up as an empty histogram below it and
a wall right at it. SoD's Windfury Weapon has nothing under 1503 ms over 145 gaps, which fits
its 1.5 sec cooldown.

### Weapon speeds for PPM

The log doesn't say which hand a white swing came from. `procrate` reads the weapons from
each player's combatantinfo gear and looks up their speeds in `assets/database/db.json`.
Items the database lacks (every SoD item) need `--speed ITEM=SPEED`. We assume each hand
swings in proportion to 1 / its speed, which splits the white hits between the hands. Yellow
hits count as main hand. `--main-hand-only` is for procs that the off hand can't trigger,
like Windfury Weapon.

## What the log gives us

What we learned from SoD logs (they should hold for Forever, which reuses the client):

- Damage events have `hitType`: 0 miss, 1 hit, 2 crit, 4 block, 6 glance, 7 dodge, 8 parry,
  15 crush. `amount` is after armor. `unmitigatedAmount` should be the log's raw field, but
  on SoD crits it is smaller than `amount`, so check it before trusting it.
- White swings of both hands are ability 1 ("Melee") with no hand flag.
- Extra attacks (Thrash, and maybe Forever's Windfury) are `extraattacks` events.
- With the `Buffs` data type, the API reads `sourceID` as the unit that has the aura. So to
  get the buffs on a player we pass the player as `sourceID`.
- combatantinfo has item IDs and enchants but no weapon speeds.
- We don't get the target's level. Bosses are actors with subType `Boss`, and the raid
  bosses are 3 levels over the level cap.

## On WoW Forever's release

1. Find the site. Try `zones` on the candidate subdomains, or open a Forever report on the
   website and read the subdomain from its URL.
2. Take one report with an enhancement shaman, one with a warrior in a shaman's group, and
   run `abilities` on both. Note the IDs of the Windfury Weapon proc, the Windfury Totem
   proc and Maelstrom Weapon, and whether each one logs as damage, extra attacks or a buff.
3. Build samples. Use rankings when Forever has them, and `--zone` reports otherwise. Start
   with `--limit 5`, check the output, then go wider. A player's fight costs about 4 points
   the first time (its report's data, the player's events and the buffs on them). So about
   80 new fights fit in one run's default budget.
4. For the cooldowns, run `gaps` with the IDs from step 2. For the Windfury Totem use
   players who don't have Windfury Weapon (warriors and rogues).
