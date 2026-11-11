# tank_mechanics pass 10 (2026-11-11, master 707ae92cc, client 1.60.1.70888+hf)

Due on 9 watched files changed (27d0d59d6..707ae92cc). Result: findings (1 new, low).

**Getting current:** wago's latest is still 1.60.1.70888 (no client diff to read). New since pass 9: Blizzard's hotfixes of November 10 (news article 24315281) and Kaivax's "Class Tuning Incoming - November 10" (forum thread 2404092), as the repo recorded them (forever-wiki/classes/shaman.md, warrior.md, druid.md). Tank-relevant items: shaman Stormstrike, Mental Dexterity, Maelstrom Weapon, imbue + weapon totem. No Protection warrior, Protection paladin or bear item. No update_ledger entry in range.

**Changed code, read in full where a tank reaches it:**

| Change | Tank reach | Client / source | Sim | Verdict |
|---|---|---|---|---|
| Stormstrike 100% weapon damage, 8 s (stormstrike.go:29, :78) | Warden (every preset) | notes 2026-11-10; client 70888 still 150% / 6 s (SpellEffect 1368308, SpellCooldowns 55241) | 1.0, 8 s under Forever; mark x1.2 and Rockbiter threat 72 x speed unchanged | right (tier 1 notes, notes-only) |
| Mental Dexterity 33 / 66 / 100% Int to AP (talents.go:39) | Warden presets with Mental Dexterity | notes 2026-11-10 (beta client 33 / 67 / 100) | .33 / .66 / 1, now for every env | right (tier 1 notes) |
| Maelstrom Weapon 5 + rank PPM, 20% a stack | none: no Warden preset takes it | notes 2026-11-10 | as notes | n/a |
| Imbue + weapon totem both work | Warden imbues Rockbiter, which already shared with a totem | notes 2026-11-10 | comment-only change in air/fire totems, shaman.go, totem_receivers.go | n/a |
| Sweeping Strikes keeps charges with no second target | Protection warrior takes no Sweeping Strikes | notes 2026-11-10 | sweeping_strikes.go:88 | n/a |
| Nature's Swiftness + Wild Growth; Idol of Ferocity -2 (Claw, Rake) | cat / healer only, no bear spell | hotfix 113885 / client 70888 | as source | n/a |

Warden golden (TestWardenShaman.results): dps about -7%, tps about -3% (e.g. 410.6 / 699.9 -> 382.3 / 677.2) from Stormstrike and Mental Dexterity. Ranking effect is tank_fairness's.

**Tests:** `go test -tags=with_db` on sim/shaman/warden, sim/warrior/tank_warrior, sim/druid/tank, sim/paladin/protection: all pass. TestForeverWardenTalents reads Mental Dexterity at 33 / 66 / 100% and TestForeverWardenAbilities Stormstrike's 8 s. (Outside this pass: sim/warrior/dps_warrior TestP1DPSWarrior fails on stale stats / DPS goldens, e.g. 65.025 -> 63.242.)

**Engine:** no sim/core tank file, tools/tankbench or tankscore change in range; pass 7's checks stand.

**Findings:** new 20261111-1 (docs/tank_shaman.md:369-370, :382-383, :404, :425 still give 150% / 6 s and 50 / 100 / 150%). Still open: 20261107-1 (stances.go:110, docs/tank_warrior.md:31, bear.go:57, docs/tank_druid.md:22, :418 labels), 20261108-1 (no shock-under-mark threat test).
