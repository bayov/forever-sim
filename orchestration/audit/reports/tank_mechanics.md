# tank_mechanics pass 3 (2026-11-04, master dc303b026, client 1.60.1.70790+hf)

Forced by hand after tank round 3 landed. Round 3 brought the engine fix fd9538c (mana-gain threat unmultiplied, disjoint tank-tool seeds via tankscore.Seed, engine tests) and the four class workers' re-optimized presets. Three read-only subagents ran: one checked the pass-2 fixes and ran the tests, one reviewed the round-3 code, and one diffed the client build and checked the notes.

**Getting current:** the client moved from 70755 to 70790. Its tank rows are Stormstrike's 150% and 6 s (SpellEffect 1368308, SpellCooldowns 55241) and Vengeance's 5 stacks (SpellAuraOptions 247421); hotfixes 113380 and 113542 already cover both. Slam's 15 s and Improved Slam are Fury. Booming Voice 1360666 gains mask bit [1] 32, which is Piercing Howl 12323 only: no tank preset casts it. Bear, Warden shields and the paladin tank spells are unchanged. Blizzard's thread: post 9 is still the newest (edited 2026-11-04T02:12:27Z), so `dev_notes_seen` stays.

## Pass-2 findings
| Finding | Verdict | Evidence |
|---|---|---|
| -12 mana gain threat × multiplier | holds, resolved | mana.go:131-133 books 0.5 a mana through BookSplitThreat (no multiplier); the Classic path keeps one but returns early under Forever; TestBearManaGainThreat passes |
| -13 stale labels | holds, resolved | stormstrike.go:17-19, tank_shaman.md:309/311, bear_attacks.go:9-11, tank_warrior.md:262, forever_enchants.go:93-96, round_70753/paladin.md:138 |
| -14 scoring-run holes, SoD parry haste | holds, resolved | tank_score.go:55/110-121/157-164 maxHealth cap; apl_values_resources.go:44/83/157; 7 SoD files parryHaste false |
| -11 tests | holds, resolved | every named test exists and passes. Full run of core/druid/paladin/shaman/warrior/tankscore: 16 packages ok |

Test strength notes: TestMakeTankParryHaste only checks "more swings and DTPS with it on" (catches haste that never fires, not a wrong amount). TestScoringRunReadsNoHealth passes trivially if the warrior's potion never fires; that it fires was not checked.

## Round-3 code
| Item | Verdict | Tier |
|---|---|---|
| Warden trinketAt (shaman_tank.go:67-69, 154-166) | correct at the preset's 25 (pull, then 155 s on 180 s); no health read. Values 5-15 are near no-ops (fixed 20 s minTimeLeft). No test. -16 | sim |
| Warden armorTrinket=0 | trinket 999216 grants BonusArmor; major_cooldown.go:398-403 checks Armor, so it is typed DPS and autocast. 0 = 1. -16 | sim |
| Bear Enrage knob (druid_tank.go:61, max 60 → 100) | correct; Enrage 5229 rows match (10 Rage, 2 Rage a s for 10 s, 60 s cd, mask 144). TestEnrage covers it. The 16% armor cut rests on text only (already open) | 1 |
| Explosives in form | **wrong** for Goblin Sapper Charge 13241 and Goblin Land Mine 4100: Attributes_0 0x0, and Bear/Dire Bear Form flags 0xd8 (CanUseItems). Sapper adds +158/+189 to bear Max threat. -15 | 1 |
| Potions, runes, healthstones, EZ-Thro in form | allowed, and the client agrees (no 0x10000) | 1 |
| Dense Dynamite, Thorium/Iron Grenade, Solid Dynamite, Holy Hand Grenade in form | blocked, and the client agrees (0x10000) | 1 |
| tankscore.Seed (1 + (s-1)·1e9) | disjoint; tankbench, abtest and rotopt all use it; tested | sim |
| health.go / tank_score.go maxHealth | correct | sim |
| stormstrike.go | matches hotfix 113380 rows (150%, 6 s, 8% mana) | 1 |
| spell_result.go:145 healingThreat | reads current health. No tank books heal threat in a scoring run, so latent. -16 | sim |

## New findings
- -15 (high): the sapper and the land mine are blocked in a form; the client allows them. Bear score +6-7% at Max threat.
- -16 (low): Warden armor-trinket knob quirks; latent health read in heal threat.
