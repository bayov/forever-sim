# healer_mechanics pass 17 (2026-11-17, master 5f20dfdf4c)

Range 1a85df1d75..5f20dfdf4c, 1 watched file (sim/shaman/item_sets_pve.go). Client 1.60.1.70938 -> 1.60.1.70959
(`wago.py diff --set mechanics`). Blizzard: hotfix thread t/2403040 posts 7-11 and post 1's Nov 16 edit.

## What changed

| Source | Change | Healer reach | Verdict | Tier |
|---|---|---|---|---|
| sim/shaman/item_sets_pve.go (6875e3b1e) | Forever Five Thunders set 1679: 2pc +8 res, 3pc 8 mp5, 4pc 5% Furious Storm (65 SP, 10 s) on white hits and on damage/heal casts, 6pc +23 SP | healer pieces 227009-227016 exist, but no healer preset or healbench gear uses any piece (only Enhancement BiS) | n/a for healers | - |
| client 70959 SpellCooldowns 84777/94542/94543, SpellCategories 150689/173189/173190 | Wild Growth 6 s RecoveryTime -> 6 s CategoryRecoveryTime, Category 1237 (all ranks share) | Restoration druid | sim already shares one timer across ranks (sim/druid/healing.go:632-639): match | 1 |
| client 70959 SpellCategories 76027/76028/76053/76065 | Judgement of Light DefenseType melee -> ranged | Holy paladin's JoL is a debuff toggle only, no hit roll | no reach | 1 |
| client 70959 Gloves of Holy Might proc mask | damage-only procs | not healer gear | no reach | 1 |
| Nov 16 hotfix | Wild Growth ranks go on cooldown | as above | match | 1 |
| Nov 16 hotfix | Vampiric Embrace heals 10% (was 20%), caster takes 100% more | Shadow priest party healing; no healer preset | sim 20% for all: finding 16 | 1 |
| Nov 12-14 posts | Quick and the Dead perk on Divine Spirit cost; Darkspear signet mp5, Channeler's Ring SP, Everstaring Abomination Eye Spirit | legacy perk not simmed; item stats are the gear data's | no heal mechanic | 1 |

No other heal, healing talent or engine row changed in 70959 (rest of the diff is Windfury Totem, Stormstrike,
Maelstrom Weapon, Arcane Blast, buff-scroll attributes).

## Tests

`go test -tags=with_db ./sim/druid/... ./sim/shaman/... ./sim/priest/... ./sim/paladin/holy/...`: all ok.

## Findings

- New: healer_mechanics-20261117-16 (low): Vampiric Embrace party heal 20% vs Forever's 10% / 20% caster.
- Still open: 10 (test coverage), 11 (Riptide mana 147/180/231 vs client 150/180/230).
