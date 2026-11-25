# Mechanics pass 105 (2026-11-25, master 0404484fb)

Range: 5536db4bc..0404484fb. Client 1.60.1.71091 still newest (probe 300). No new Blizzard post:
dev notes t/2360696 8 posts (newest 2026-11-04), hotfix thread t/2403040 15 posts (newest #15,
2026-11-24 02:06Z, already read). No daily-update changes to TASKS.md or update_ledger.md in range.

## Commits checked

| Commit | Change | Check | Verdict |
|---|---|---|---|
| f1942793d | Arcane Intellect / Divine Spirit overwrite Int / Spirit elixirs (incl. Cunning's Int) | 71091: Cunning 1250918 aura 29 25/25, Sages 1250941 25 Spi + aura 290 2%; rule from EV-0185 in-game sheet | OK (log/in-game) |
| 75dee7458 | Armor Shatter + Puncture Armor in minor armor slot; Puncture 100/stack | 17315 aura 674 bp -100 | OK (tier 1 + EV-0147) |
| 3275a60b1 | World buffs off in raid instance | AuraDescription "No effect while in a raid instance" on 22888, 16609, 15366, 22817/8/20, all Sayge's | OK (tier 1) |
| 9701cff08 | Spirit of Zandalar off in raid too | 24425 lacks the text; user ruling | settled (ruling) |
| 5dbc01412 | Raid/party buff scope comments | comment only | OK |
| b4cab8498 | Damage-taken Rage: stance modifier; 2x for spells, ticks, Physical specials | stances carry no client damage-taken effect; log-sourced (rage_taken_by_stance.md, EV-0052, EV-0118); StanceDamageTakenMultiplier defaults to 1 | OK (logs) |
| 6992f4e80 | compsearch duty-variant fillers | not a mechanic | n/a |

## Shared run vs published

60 of 68 builds moved. Largest: Enhancement -2.6% (preset runs Greater Intellect, Cunning and
Sages under Arcane Brilliance and Divine Spirit, so all three elixirs' Int/Spirit now go unused;
Sages' 2% crit still counts), Paladin -1.6%, Arcane/Frostfire mage -1.5 to -1.7%, warlocks
about -1%, hunters -0.7%: consistent with f1942793d's mana loss. Warriors +0.2% (Rage change).
Preset elixir choice under AI/DS is a preset question, not a mechanics finding.

## Findings

None new. mechanics-20261124-1 (Sayge's fortunes 10%) gets a note: it now matters only outside
raid instances.
