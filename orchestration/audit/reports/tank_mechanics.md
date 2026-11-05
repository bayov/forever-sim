# tank_mechanics pass 4 (2026-11-05, master a07941531, client 1.60.1.70790+hf)

Due on watched files: tank rounds 4 and 5 (full raid consumables, talents and knobs re-searched), the -15/-16 fixes (22a7d79, 464deda), Feral's Ferocious Bite (ff56f8d). Result: quiet.

**Getting current:** wago's latest is still 1.60.1.70790 (no new build). Blizzard's development notes thread could not be read: the forum answered "performed this action too many times" (HTTP 422) on two tries, so `dev_notes_seen` stays at post 9 (2026-11-04T02:12:27Z) and the next pass should check it. No update_ledger changes since dc303b0.

**What changed (watched, non-test):** sim/druid/ferocious_bite.go (Cat only; no bear request casts it) and the 12 tankbench requests. No engine file changed.

## New consumables in the tank presets, against the client (tier 1)
| Item | Client row | Sim (consumes.go) | Verdict |
|---|---|---|---|
| Goblin Sapper Charge 13241 | eff 0: 600 Fire, Variance 0.5, radius around caster; eff 1: 500 to the caster, Variance 0.5 | 450-750 to targets in 10 yd, 375-625 to self (1394, 1386-1389) | right |
| Sapper self-hit | (no client row on how it's booked) | RemoveHealth → tank score damage taken and healing needed; Rage from damage taken for warrior and bear | right by Classic rules (tier 3) |
| Greater Stoneshield 17540 | +2000 armor, 120 s | BonusArmor 2000, 120 s (1583); Dire Bear Form scales item Armor only (bear.go:141) | right |
| Flask of Natural Precision 1293742 | +60 Stamina; dummy 5 and A_MOD_EXPERTISE for the zone bonus | +60 Stamina, 5% expertise only with NaturalFlaskZone (77-117) | right |
| Rumsey Rum Black Label 25804 | +15 Stamina | +15 Stamina (560) | right |
| Spirit of Zanza 24382 | +50 Stamina, +50 Spirit | same (1045) | right |
| Ground Scorpok Assay | +25 Agility (Strike of the Scorpok 10669) | +25 Agility (1054) | right |

## Talents and spells newly taken
| Spec | Change | Client | Sim | Verdict | Tier |
|---|---|---|---|---|---|
| Prot paladin | Improved Seals 5 (`05005-...`) | 20224 aura 108, 3%/rank (trait), mask [0x2000400, 0x20000800, 0x40] reaches the Seal of Fury hit 20418 (mask[1] 0x800) and Judgement of Fury 20414 (mask[2] 0x40) | 1.03/rank on both (talents.go:72, seal_of_fury.go:106/125); TestImprovedSealsOnSealOfFury | right | 1 |
| Prot paladin | boots 929 → 1887 (Greater Agility) | (enchant rows checked pass 2 for tank enchants) | | not re-checked | |
| Prot warrior | Deflection 4 → 5, Blood Craze 1 → 0 | ranks unchanged since pass 1 | | right | 1 |
| Warden | Searing Totem r6 10438 (L60, 170 mana), Fire Nova r5 408345 (L52, 375 mana, 10 s), Mana Spring r4 | Fire Nova hit 408428: 239 + 1.3/level, Variance 0.1098, 14.3% coefficient | 232.03-258.97 at the fifth level, 0.143, 375, 10 s (fire_totems.go:245-252); totem attacks threat 0, Fire Nova 1x (docs/tank_shaman.md:411) | right | 1 (rows), 3 (totem threat) |

## Engine
No engine file changed. The pass-3 -15 and -16 fixes are on master and already resolved. Tests: core, paladin, druid, warrior, shaman and tankscore packages, 14 ok, none failing (`go test -tags=with_db`).

## Left open (not findings)
- The sapper's self-hit rolls the caster's spell crit (OutcomeMagicHitAndCrit on self), so a self-crit can trigger Reckoning, Redoubt or Enrage. Whether the self part can crit in game is unknown; it fires once a fight, so the effect is far below noise.
- Development notes since post 9: unread (rate limit).

## New findings
None.
