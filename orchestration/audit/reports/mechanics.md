# Mechanics pass 48 (2026-10-30, master d50909a37)

Range audited: 334c7decd..d50909a37. 25 code commits, all the raid sim (milestones B, M1, M1b, M2): `sim/compsearch` (new), shared engine (`buffs.go`, `debuffs.go`, `camp_buffs.go`, `totem_weapon_buffs.go`, `raid.go`, `environment.go`, `presim.go`, `aura*.go`), and druid, hunter, paladin, priest, shaman, warlock, warrior class files. TASKS.md changed only for the raid sim; no `forever-update/*` merge, no update-ledger change.

## Get current

- Client: 1.60.1.70685 is still the newest (gate). Every verdict holds.
- Dev notes thread 2360696: still 8 posts; post 7 (edited 10-27 16:47:14) and post 8 (edited 10-27 16:46:55) unchanged.
- Blue tracker after 10-29 00:51: Forever posts are the November 4 launch announcement, a podcast and a starter guide. No mechanics.
- Server hotfixes: `tools/audit/mechanics/hotfixes.py` is still missing from the tree; not re-read.

## Single-player effect

The shared run against the published report: 1 of 77 builds moved by more than 0.5 DPS (mage frostfire +0.9, +0.08%). No mage file in range, so this is no mechanic change. The raid sim's engine changes keep the single-player paths (Blood Pact moved to party buffs, read as max(raid, party); a lone warrior's shout still counts 4 extra threat hits).

## Client values the new code claims (70685)

| Code | Claim | Client | Verdict |
|---|---|---|---|
| `core/camp_buffs.go` | 9 camp buffs, values and spell IDs | 1230098 aura 137 8% all; 1230164 90 AP; 1230172 34 Str (effect 35); 1229519 2% (52 + 57); 1229513 25 Int; 1229718 32 Spi; 1230587 29 MP5; 1230653 308 armor / 13 stats / 22 resists; 1230124 56 Sta | Match. Exclusivity with class buffs comes from the tooltip only and the code says it is unverified. |
| `core/debuffs.go` Dust Cloud | 65 / 175 / 285 / 395 / 505, 1265899-1265904 | aura 674, those base points, 30 sec | Match |
| `core/buffs.go` Furious Howl 83 | 24597 at 60 | 82 + 0.4 x (60 - 56) = 83.6 | Match (floor) |
| `druid/faerie_fire.go` | Feral FF gone; caster FF in Cat / Bear, free there | 16857, 17390-17392 have no rows; 9907 ShapeshiftMask 0x40000091 (Cat, Bear, Dire Bear, Moonkin), 115 mana, 1500 GCD; dev notes "costs no resources in those forms" | Match; no SpellFlagOmen, so no Omen of Clarity from it (dev notes) |
| `paladin/sow.go` | SoW 38 / 48 / 58, 135 / 170 / 200 mana, 50 / 71 / 90 mana procs; JoW 40 sec, 33 / 46 / 59 | 20166 / 20356 / 20357, 20168 / 20350 / 20351, 20186 / 20354 / 20355 (dur 40000), 20268 / 20352 / 20353 | Match. 50% seal proc chance is an assumption (client 100, server-side); no preset casts it. |
| `warrior/shouts.go`, `hunter/pet_forever.go` | Battle Shout and Furious Howl reach the raid | 25289, 24597 implicit target 56 | Match |
| `warlock/warlock.go` | Blood Pact is a party aura | 11767 effect 35 | Match |
| `shaman/water_totems.go` | Mana Spring reaches the party | 5677 effect 35 | Match |
| `priest/talents.go` | Dark Inspiration reaches the party | 1326168 implicit target 20 | Match; given form at 0.999 uptime from the report's Shadow builds |

No DPS sims run: no single-player number moved and every claimed value matches the client.

## Findings

- No new findings. mechanics-20261027-2 (Amplify Curse vs Bane of Agony) stays open (parked for an in-game test); no warlock curse change in range.
- verdicts.md: added rows for camp buffs, area reach, Furious Howl, Dust Cloud, Forever Faerie Fire and Seal / Judgement of Wisdom.
