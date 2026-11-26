# tank_gear pass 20

Master f9fa79917 (last pass: 3e9748641). Triggered by 17 watched files: the Warden, bear Balanced and warrior Max threat
gear sets and requests, phase0_specs.json (enh talents only). Helpers: phase0_check.py, tankbench (`-tags=with_db`), 3,000
iterations, seed 1, six bosses, own column (tools/audit/tank_fairness/bench.sh, summary.py).

## Mechanical checks
`phase0_check.py`: 31 flags, 27 on tank rows, all "open item above the pick" within the bench's measured noise and
re-tested in phase0_tankheal_*.md (finding 4 updated: the noise is now 0.5% of scoreRef, still tight for Max survival).
Druid restoration has no list yet (expected). No gear-set, request, faction or tag flag on a tank row.

## Changed lists
| preset | race | list | bench s1 | changed since last pass |
|---|---|---|---|---|
| warrior Balanced | Undead | 3360.5 | 3359.0 +/- 9.6 | list re-ranked (was 3299.4) |
| warrior Max threat | Undead | 4725.7 | 4721.7 +/- 5.6 | wrist Frozen Steel Vambraces -> Bracers of Heroism (226863) |
| warrior Max survival | Undead | 1218.7 | 1222.0 +/- 12.8 | none |
| bear Balanced | Night Elf | 3771.6 | 3783.2 +/- 9.9 | re-searched from the Max threat list: Defias Enforcer's Mask, Horn of Ve'han, Truestrike Shoulders, Dawn Armor, Blackmist Armguards, Hammer of Thundering Hooves, Idol of the Ursine Twins |
| Warden Balanced | Troll | 2022.6 | 2009.2 +/- 8.7 | request drops the dead trinket line |
| Warden Max threat | Troll | 3427.9 | 3415.2 +/- 5.7 | cloak Shifting Cloak (18511), legs Kilt of The Five Thunders, gloves enchant Natural Power |
| Warden Max survival | Tauren | -1213.1 | -1202.8 +/- 19.6 | enchants; Alliance row (Dwarf) Theramore Epaulets, Flame Walkers |

Doubtful picks (all tagged and noted, standards 42; none is a finding): bear Balanced Horn of Ve'han (quest Work Well Done),
Idol of the Ursine Twins (craft, formula source unknown), Enchant 2H Weapon - Strength (no known source), Defias
Enforcer's Mask (no source in our data, likely Alcaz Prison); Warden Max survival Defiler's Chain Greaves (Horde, AB
unconfirmed).

## Requests and other faction
Every changed request wears its gear set (check clean). Warrior Max threat Alliance differs from Horde only in wrist
(13400) and hands (226861); Warden Max survival Alliance only in shoulder and feet (Theramore Epaulets, Flame Walkers).
Bear Balanced Horde set equals the Alliance one (no faction-only pick: Horn of Ve'han is a both-side quest).

## Findings
New: none. Resolved: -1 (warrior bench gap gone after the re-rank), -2 (19112 out), -3 (annotations rewritten, no worn
infeasible pick). Still open: -4 (tool noise).
