# Boss encounters

Boss fights for the tank bench (`tools/tankbench -encounter`) and, once it reads them, the heal
bench (`tools/healbench`), one file a boss, shared so the two sims score against the same fights
(docs/tank_sim.md, docs/healing_sim.md).

```json
{
  "name": "Broodlord Lashlayer",
  "source": "where each number comes from, with its source tier",
  "abilitiesSource": "where the abilities' numbers come from (from_logs.py keeps it and writes the rest of source)",
  "reports": ["Warcraft Logs report code:fight, for the numbers taken from logs"],
  "warcraftLogs": {"site": "sod", "encounter": 100612, "zone": "...", "difficulty": "Normal", "mined": "logs/sod_bwl.json"},
  "notes": "what the file leaves out",
  "encounter": { "a proto Encounter in protojson": "the boss's melee (targets), tank_swap_seconds, raid_damage" },
  "healing": {
    "healers": 3, "hpsEach": 300, "castSeconds": 2.5,
    "notes": "the raid's healing on its main tank: the tank bench's stand-in stream until the healing sim's healers heal"
  },
  "raidSource": "where the raid damage (who NotTanks) comes from, raid_from_logs.py",
  "raidReports": ["the kills it was mined from"],
  "raidHealing": {"healers": 4, "raidHps": 1100, "deathsOthersPerKill": 0.2, "notes": "what real raids had"}
}
```

- A boss's abilities on the tank are `raid_damage` events with `who` BossTarget or Tanks. A
  melee one (Mortal Strike, Cleave, Hateful Strike) sets `melee`, so it rolls the tank's
  avoidance and block. A healing cut sets `healing_taken_multiplier` and `debuff_seconds`. A DoT
  that stacks is one event a stack, since a sim DoT restarts instead of stacking.
- **Sources, best first** (the user, 2026-11-02, docs/data_sources.md): Warcraft Logs from Season
  of Discovery, then the Anniversary realms (Vanilla, then TBC), then Classic Era, then the
  private server cores. `tools/warcraftlogs/wclcli.py tank` reads a boss's damage on its main
  tank and the tank's healing from logs.
- The tank side (the boss's melee, its abilities on the tank, the healing on the tank) comes
  from the tank sim, the raid damage from the healing sim.

## The raid side, from the logs

The healing sim's part of each file: the damage on the players who aren't tanks, as `raidDamage`
events with `who` NotTanks, and what real raids had for healing it. `raid_from_logs.py` writes them
from what `wclcli.py raid` mined, after `from_logs.py` (which leaves NotTanks events in place but
copies the top-level files' into `sod/`). Each script picks a site's folder itself:

```sh
W="python3 tools/warcraftlogs/wclcli.py"
$W raid tools/encounters/logs/sod_mc_sample.json --difficulty 3 \
    --encounter 100664 --encounter 100670 --encounter 100672 -o tools/encounters/logs/sod_mc_raid.json
python3 tools/encounters/raid_from_logs.py --site anniversary tools/encounters/logs/anniversary_*_raid.json
python3 tools/encounters/raid_from_logs.py --site sod tools/encounters/logs/sod_*_raid.json
```

- Each ability that dealt 2% or more of the non-tanks' damage in 10% or more of the kills is an
  event: its unmitigated 5th to 95th percentile stretched to an even roll, the `share` of the
  non-tanks a cast hits (so a file fits a raid of any size), and a DoT's ticks an application and
  their spacing. Its rate, at random times from its median first cast, is the one that lands the
  logs' damage a second on each non-tank, at the logs' mitigation, over the rest of the kill; the
  abilities under the line are spread over the rest. The logs' cast counts are rough for a ground
  effect, their damage isn't.
- `healbench -role raid` and `both` run the logs' median kill (`raidHealing.fightSeconds`), since
  the file's `duration` is the main tank's time on the boss.
- `raidHealing` is what the raids had, to check the healing sim's healer counts against (the
  user's 20% margin): raid size, tanks, healers by Warcraft Logs' healer role and their classes,
  HPS each, the raid's healing on the tanks and the rest, the non-tanks' damage taken, deaths a
  kill, and the fight's 10th, 50th and 90th percentile length.
- `raidSource` says where each event's numbers come from, `raidReports` lists the kills.

## From the logs

The top-level files are the Anniversary Hardcore realms' logs, and they score tanks (the user,
2026-11-02). `sod/` has the same bosses from Season of Discovery's Normal raids, the first source in
the order, but tuned far above a pre-raid tank, to compare. In both, the
boss's melee, his weapon damage abilities, the fight's length and the healing on the main tank are
Warcraft Logs numbers. The other abilities keep the client's spell data and the scripts' timers.

```sh
W="python3 tools/warcraftlogs/wclcli.py"
# Fights: the newest public reports of a zone (rankings favour the fastest kills, and old
# ranked reports are archived).
$W --site sod sample -o tools/encounters/logs/sod_mc_sample.json --zone 2012 --reports 60 --class Priest
# What the bosses did to their main tanks, Normal difficulty only.
$W --site sod tank tools/encounters/logs/sod_mc_sample.json --difficulty 3 \
    --encounter 100664 --encounter 100670 --encounter 100672 -o tools/encounters/logs/sod_mc.json
# Into the files (it runs tankbench -boss-stats for the debuffs).
python3 tools/encounters/from_logs.py --site anniversary tools/encounters/logs/anniversary_*.json
python3 tools/encounters/from_logs.py --site sod tools/encounters/logs/sod_*.json
```

`logs/` keeps each sample (the fights read) and what `tank` mined from it, so the files can be
written again without the API. How each number is picked is in `from_logs.py` and
docs/tank_sim.md, Encounters.
