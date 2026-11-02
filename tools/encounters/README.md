# Boss encounters

Boss fights for the tank bench (`tools/tankbench -encounter`) and the heal bench
(`tools/healbench`), one file a boss, shared so the two sims score against the same fights
(docs/tank_sim.md, docs/healing_sim.md).

```json
{
  "name": "Broodlord Lashlayer",
  "source": "where each number comes from, with its source tier",
  "reports": ["Warcraft Logs report code:fight, for the numbers taken from logs"],
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
  avoidance and block. A healing cut sets `healing_taken_multiplier` and `debuff_seconds`.
- **Sources, best first** (the user, 2026-11-02; docs/data_sources.md): Warcraft Logs from Season
  of Discovery, then the Anniversary realms (Vanilla, then TBC), then Classic Era, then the
  private server cores. `tools/warcraftlogs/wclcli.py tank` reads a boss's damage on its main
  tank and the tank's healing from logs.
- The tank side (the boss's melee, its abilities on the tank, the healing on the tank) comes
  from the tank sim, the raid damage from the healing sim.
