# tank_mechanics pass 19

- master: 2f7e255af8 (from 5f20dfdf4c)
- client: 1.60.1.70959 (unchanged; wago.tools latest)
- result: quiet (no new findings)

## What changed on master (watched files)

| Commit | File | Change | Tank impact |
|---|---|---|---|
| 4e81669a9 | sim/paladin/holy_strike.go | With Shieldsworn, Holy Strike carries SpellFlagNoWindfuryTotem (Blizzard hotfixes 2026-11-16, t/2403040/11: Shield Slam, Shield Bash and Shieldsworn Holy Strikes no longer activate Windfury Totem) | Prot paladin: every Level 60 preset takes Shieldsworn (`-55302102513521531`, 17th point) and raid Windfury Totem, so Holy Strike stops rolling the totem. Other main hand procs unchanged (post names only the totem). Tested: TestShieldswornHolyStrikeNoWindfuryTotem (sim/paladin/retribution/windfury_totem_test.go:169), with/without talent |
| 192cb2e06 | sim/druid/healing.go, items_restoration.go | Idol of Rejuvenation: Regrowth +5% on a Rejuvenated target | none: healing only, no Feral tank effect |
| 6a3cadbfc | sim/shaman/items.go | Totem of Tides: Riptide cost -5% | none: Warden casts no Riptide |

## Checks

| Item | Sim | Verdict | Tier |
|---|---|---|---|
| Shieldsworn Holy Strike vs Windfury Totem | flag set on talent; `windfuryTotemRollsOnForever` (sim/core/buffs.go:2217) honours it | matches | 1 (Blizzard notes) |
| Shieldsworn's Holy Shield damage (shieldswornStrike) | ProcMaskSpellDamage (holy_shield.go:70), never a melee roll | matches | 1 notes / reading |
| Warrior Shield Slam vs Windfury Totem | ProcMaskMeleeOHSpecial under Forever (shield_slam.go:29-33), no totem roll | matches (same hotfix) | 1 notes |
| Warrior Shield Bash | not simulated | n/a | |

## Open findings re-checked

- 20261117-1 (Unbridled Wrath skips Sunder): still `result.Damage > 0` (sim/warrior/talents.go:238). Open.
- 20261108-1 (Warden shock threat under Stormstrike mark untested): unchanged. Open.
- 20261107-1 (stale 'client says 30% / 50' comments): sim/warrior/stances.go:109-114 unchanged. Open.

## Not done

- Blizzard forum notes not re-read beyond the Nov 16 post the commit cites; no new client build since 70959.
- Windfury Totem 70959 reshape: still awaiting the user ruling (verdicts Open).
