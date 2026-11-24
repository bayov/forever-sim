# Healer mechanics: pass 28

2026-11-24, master eceb090196 (last e29cfe00f6), client 1.60.1.71091 (last 71037). Result: findings (1 new, low).

## Why due

- master: `sim/core/mana.go` (ed1df1d51, "Innervate and Energized scale the MP5, not periodic energize ticks", EV-0114).
- client 1.60.1.71037 -> 1.60.1.71091.

## ed1df1d51 (read in full: mana.go, buffs.go, racials_forever.go, stats.go)

Under Forever, aura 110 (Innervate +4, Read Leyline's Energized +0.5) now multiplies MP5 as well as Spirit regen, except
`PseudoStats.EnergizeMP5`, the MP5 the sim adds for periodic energize auras (Blessing of Wisdom, raid Mana Spring).
Classic keeps Spirit only. Source: wowsims/forever EV-0114 log (70938), log evidence. Healer reach: the Restoration druid's
self-Innervate (both presets, at 35% mana) now also gets 5x its gear and elixir MP5 for 20 s.

Client check of every MP5 source the sim adds as a stat (71091, tier 1):

| Source | Client spell | Aura | Sim | Under aura 110 | Verdict |
|---|---|---|---|---|---|
| Blessing of Wisdom (and Greater) | 19742-25290, 25894/25918 | 24 energize | MP5 + EnergizeMP5 | unscaled | OK |
| Raid Mana Spring | 5677, 10491 | 24 energize (area) | MP5 + EnergizeMP5 | unscaled | OK |
| Shaman's own Mana Spring / Mana Tide | 10497 etc. | 24 | AddMana ticks | unscaled | OK |
| Camp Mana Well | 1230587 | 24 energize, 29/5 s | MP5 29 only | **scaled** | finding -17 |
| Nightfin Soup | 18194 | 24 | Classic only (Forever: spell damage) | n/a | OK |
| Mageblood elixirs | 24363, 1250942/4/8 | 85 regen | MP5 | scaled | OK |
| Wicked Regeneration | 1250932 | 85 (+161) | MP5 | scaled | OK |
| Warchief's Blessing | 1304594 | 85 | MP5 | scaled | OK (EV-0114 measured it) |
| Epiphany (priest T3) | 28804 | 85 | MP5 | scaled | OK |
| Atiesh (healer) | 28145 | 85 | MP5 | scaled | OK |
| Item MP5 | 21359-21366, 21618-21632... | 85 | MP5 | scaled | OK |

Engine verdict row for Innervate updated in `ref/healer_mechanics/verdicts.md`. Test: TestForeverAura110ScalesMP5
(sim/mage/forever_test.go) covers Innervate and the BoW exclusion; nothing covers the Mana Well. Doc
`docs/healing_druid.md:153` still says Spirit regen only (in the finding).

## Client 71037 -> 71091

`table_summary.py`: Spell 8 added / 6 changed, SpellEffect +15 / -5 / 42 changed, SpellMisc 7 added / 34 changed; no
Trait*, SpellClassOptions, SpellDuration or SpellCastTimes change. Rows that name heals or healer spells:

| Row | Change | Player spell? | Healer impact |
|---|---|---|---|
| Healing Wave 11986 | base 1502.7 -> 75.7, Attributes_0 cleared | no (NPC; player ranks 331-25357 unchanged) | none |
| Rejuvenation 1280564, Tranquility 1280634 | coeff 0.25 -> 0, base 5 -> 6, text | no (new-ID NPC/companion rows, not in sim) | none |
| Judgement of Light/Wisdom/Justice 25752/25757/25945 | duration 10 s -> 40 s | no (player ranks 20185-20355 already 40 s) | none |
| Judgement 20271 | Attributes_4 +0x800000 | yes | DPS-side paladin, left to mechanics |
| Nature's Bounty 27784 (Wildheart 5P) | misc 0 -> 3 (mana -> energy) | yes | feral only, unmodelled |
| Revitalize 23863/23864, Bloom/Wither, druid companion spells | proc/energy/new | no healer use | none |
| Chastise 1277331-5 | SpellShapeshift rows removed | priest, not in healer presets | none |

## Not done

- Blizzard hotfix thread t/2403040 and dev notes 2360696: the forum answered "performed this action too many times"
  (rate limit) on every fetch this pass; `dev_notes_seen` left at 2026-11-21. Next pass re-reads.
- No sims (none needed for a low finding; the 464 mana is arithmetic).

## Open findings

-17 (new, low), -11 (low), -10 (low). Unchanged: -10, -11.
