# Mechanics pass 37 (2026-10-25 13:08 UTC, master 54878fa3d)

Range audited: daba9d1a5..54878fa3d (37 non-merge sim / mksettings commits, mostly the 2026-10-23
alignment round with wowsims/forever, merged 10-24 / 10-25). Result: findings (1 new, low).

## Get current

| Source | State |
|---|---|
| Client build | 1.60.1.70532, no build answered up to +300 |
| Server hotfixes | New this pass: every spell-table hotfix on 70425 / 70450 / 70532 read from wago.tools/hotfixes (tools/audit/mechanics/hotfixes.py, up to push 113205, 2026-10-23). See below |
| Dev notes thread t/2360696 | newest post /6 (10-16 00:15, edited 10-16 20:50), unchanged |
| Blue tracker | nothing new since pass 36: beta ended 10-23 (Kaivax), RP-PvP ruleset, podcast; Linxy posts are retail Midnight |
| Daily update | no forever-update merges; TASKS.md / ledger changes are the alignment round's (mage Frost hotfixes, Hunter's Mark 143, Champion of the Light, Fatal Wound, Baron Charr's Sceptre) |

## Shared commits checked

| Area | Commit(s) | Verdict |
|---|---|---|
| Rage: no Rage on misses / dodges / parries, crit adds one-hand rate, damage taken 2.5 × (1 − A) × D / c(attacker), periodic ticks, Rage threat | alignment round b3f15c1d1, c54f103bf | User-approved 2026-10-23 (standards 21), beta logs via wowsims/forever. verdicts.md rows updated |
| 80% refund of a missed ability's paid Rage; extra attacks swing through a resetting cast | 90f6f9719 | Refund on CurCast.Cost under Forever, Classic's 80% share; consistent |
| Spell resists in 10% steps around R / (|R| + 9L), Spell Penetration adds damage, damage shields never partial | alignment round e04953fb9 | Log-fitted, user-approved; the presets' boss has 0 resistance so only Spell Penetration gear moves |
| No hit suppression | alignment round 338eaf8fc | Standards (2026-10-23), logs |
| Natural flasks, explosives, damage potions, one crit elixir | f8557de17, 1bd41fea2, b0f368e6e | All client values match 70532 (table in verdicts.md) |
| Item procs and imbues | da4f2018a, e5cbda3cc | 22 proc rows match 70532 SpellEffect (base points and variance); Nightfall 15% misc 126 matches |
| Sets by client set ID | d86436d1d | Not re-derived this pass (class / gear auditors) |
| Night Elf Shadowmeld | 84523db10 | 70532 20580 allows combat use; 2 min in combat from the tooltip only (conservative) |
| Troll Shadowguard | f8317e9ab | Rows match; 96 vs 116 open (priest docs), up to about 0.7 DPS |
| Pets: inherited stats from owner totals, no consumables | d9549de24 | Scroll rows carry Attributes_3 0x100 on 70532 |
| Furious Howl in Battle Shout's slot | 1ca4850e1 | No tier 1 source; Era stacks them. No preset uses a Wolf. Open item, not filed |
| Dust Cloud 505, Demoralizing Screech, Hunter's Mark 143 | 1ca4850e1, 22eb80d0a | Match 70532 (Hunter's Mark 143 is also in the hotfix table) |
| Wand Shoot on the ranged table with spell crit | 1cc6e540e, c43c22730 | Log-based (wowsims/forever), class-side |

## Server hotfixes (new check)

wago.tools records the hotfix rows the server sends on top of each build. 270 SpellEffect, 259
SpellMisc and ~150 other spell-table records on 70532 (most for new NPC / quest spells the sim
doesn't reference). Of the rows the sim or presets use:

| Row | Hotfix | Status |
|---|---|---|
| Spellstone 1237162/4/5 damage school | Misc 36 -> 32 (Shadow only), push 112965 | **mechanics-20261025-1** (-1.4 to -2.8 DPS on the Spellstone warlocks) |
| Hunter's Mark rank 4 (14325) | 71 -> 143 | in the sim since 2026-10-23 |
| Champion of the Light curve 331369-71 | 10/20/30 -> 20/40/60 | in the sim (PAL-U-HF1) |
| Elemental Precision, Ice Barrier, Frost Ward, Frost Warding, Winter's Chill | mage hotfixes | in the sim / ledger (MAGE-U-HF1-5) |
| Bloodthrill duration | index 32 -> 28 (5 sec) | in the sim (WAR-U-01) |
| Elemental Focus 16164 | ProcChance 10 -> 100, bp 0 -> 10, tooltip $h -> $s1 | same 10% chance, reworded; no change |
| Touch of the Grave drain 1260198 | drops Attributes_3 0x20000000 and bits in 9, 10, 13; keeps Attributes_6 0x20000000 | dev notes' "no longer scales" still stands; watch |
| Elemental Mastery 16166 | COOLDOWN_ON_EVENT dropped | shaman worker may want to check; no measurable effect expected |
| Twilight Focus / Revelation masks | add Confounding Flash / Penance | no preset casts them |
| Curse of Weakness, finishers, Fire Nova, Windfury Weapon, Flamestrike | attribute bits only | no known DPS effect |

## Conclusions

- The alignment round's shared changes are user-approved and documented in standards; the client
  values they cite check out on 70532.
- Hotfix tables are a source the passes had not read; hotfixes.py now does it. Run it every pass
  and after each forever-update.
- Patch for the finding: `$AUDIT_WORK/mechanics/spellstone.patch` (scratch).
