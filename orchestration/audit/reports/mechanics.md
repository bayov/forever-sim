# Mechanics pass 87 (2026-11-18, master 2f7e255af)

Range 5f20dfdf4..2f7e255af. Client unchanged at 1.60.1.70959. No new findings; both open findings unchanged
(mechanics-20261113-2 Weakening Disease, mechanics-20261104-1 druid scales).

## Get current

| Source | Newest | New since pass 86 |
|---|---|---|
| Client build (gate) | 1.60.1.70959 | none |
| Dev notes t/2360696 | post 10, 2026-11-04 | none |
| Hotfix post t/2403040 | post 11, 2026-11-17 02:17Z; post 1 edited 02:18Z (consolidates post 11) | none |
| Blue tracker | 2026-11-17 18:38Z (Blackmaw Hold abuse, not a mechanic) | nothing mechanical |
| Daily update / TASKS | Nov 15-17 items (WF Totem rulings, item reveals) | settled by user rulings already recorded |

## Commits triaged

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 3ec610c40 | Darkmoon Blue Dragon 3%, Heroism 158-238 (Forever only) | wago hotfix rows, push 114044: 23688 SpellAuraOptions ProcChance 3; 23682 SpellEffect bp 198, variance 0.4 | matches (tier 1) |
| 4e81669a9 | Shieldsworn Holy Strike: SpellFlagNoWindfuryTotem; Vampiric Embrace 10% of the casting priest's Shadow damage, x2 for the priest | post 11 wording quoted in code; Shield Slam/Bash already OH specials (sim/warrior/shield_slam.go:25) | matches (tier 1); flags.go change is a comment only |
| 2dbcd19a0 | Revealed items: Cannibal's Cleaver +25 AP vs Humanoid, Draconian Aegis +9 SP vs Dragonkin | tooltip-sourced item effects, no shared formula moved | no finding |
| 192cb2e06, 6a3cadbfc | Idol of Rejuvenation, Totem of Tides (healer items) | tooltip-only, healer scope | left to the healer auditors |

Knock-on: windfuryTotemRollsOnForever (sim/core/buffs.go:2216) is the only reader of SpellFlagNoWindfuryTotem; Holy Strike
is Retribution's too, but the flag is set only with the tank talent Shieldsworn, so DPS presets are unaffected.
