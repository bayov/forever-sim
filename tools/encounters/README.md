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
  }
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

## From the logs

The top-level files are Season of Discovery's Normal raids, the first source in the order;
`anniversary/` has the same bosses from the Anniversary Hardcore realms' logs, to compare. In both, the
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
python3 tools/encounters/from_logs.py --site sod tools/encounters/logs/sod_*.json
python3 tools/encounters/from_logs.py --site anniversary --out tools/encounters/anniversary \
    tools/encounters/logs/anniversary_*.json
```

`logs/` keeps each sample (the fights read) and what `tank` mined from it, so the files can be
written again without the API. How each number is picked is in `from_logs.py` and
docs/tank_sim.md, Encounters.
