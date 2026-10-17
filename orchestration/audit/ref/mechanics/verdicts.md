# Mechanics audit: verdicts and settled items (as of 2026-10-17 13:00 UTC, master b389688f7)

Tier: 1 = Forever client or Blizzard Forever notes, 2 = Season of Discovery, 3 = Era / Classic. Never cite vmangos.

## Baseline

- Client: 1.60.1.70450, current. Its spell data equals 70425's, apart from 22 cosmetic rows (pets, mounts, Adventurer's Tent, ensembles). TraitDefinitionEffectPoints, TraitNodeEntry and SpellItemEnchantment are identical to 70425's. 17 CurvePoint rows changed, none on a talent curve.
- The last balance build is 70425: 575 SpellEffect rows changed against 70338. The orchestration/rank_tuning.md table (8e30e52) lists the class deltas.
- The wago.tools builds API lists only 70374 / 70338 / 70334 under wow_cn_beta. 70425 and 70450 were found by probing the db2 CSV URLs.

## Shared engine, raid and racials: verified against the client (don't re-check unless the client changes)

| Item | Sim (master) | Source |
|---|---|---|
| Touch of the Grave | 20% chance (30% for Mage, Priest and Warlock), 25 sec ICD; the drain is 5% of max health, no damage mods | 70425: 1260189 / 1260201 ProcChance 20 / 30, ProcCategoryRecovery 25000; 1260198; 16 Oct notes |
| Eureka! | 8 sec, no charge limit, 10% cost / 10% damage, per-class masks (racials_forever_eureka.go), 2 min CD | 70425 duration index 31, no SpellAuraOptions row; CD 120000 on all five variants |
| Blood Fury | +10% AP and SP, 15 sec, 3 min CD | 70425 20572 RecoveryTime 180000 |
| Berserking | Flat 10%, no cost | pass 1 |
| Wind Blessed (Skyborne) | 3% melee, ranged and cast haste | 70425 1259710, both effects 3 |
| Earth Fury (Dwarf) | Crit multiplier ×1.02 on the whole multiplier (melee 2.04, spell 1.53) | 70425 1259719, aura 163 bp 2; Relentless Earthstorm precedent (tier 3) |
| Axe Specialization (Orc) and Ruthless (Human) | Crit with any weapon since 70425 | #53 / #56 |
| Troll Regeneration | 70425: 20% / 20% in combat (leveling only) | 20555 |
| Beast Slaying, Big Game Hunter, Elemental Insight | Damage only, no crit damage | aura 168; dbe2cf45a |
| Moonkin Aura / Leader of the Pack | Spell crit / weapon crit, 3% each, raid-wide, they stack | 70425 24907 effect 65 aura 57; 24932 aura 52 |
| Battle Shout, Trueshot Aura, Strength of Earth, Grace of Air, Windfury and Flametongue Totems | Raid-wide (effect 65) since 70425 | 70425 |
| Grace of Air | Stacks with Windfury Totem; Windfury and Flametongue Totems stay exclusive | 70425 Attributes_11 0x400 dropped; 16 Oct notes |
| Flametongue Totem | 548 / 781 / 1061 / 1363 per the N / 25 per 4 sec rule, white swings only (ProcTypeMask 4), no spell power | 70450 8253 / 8248 / 10523 / 16389 |
| Raid buff ranks (Arcane Intellect, Devotion Aura, Blood Pact, resistance auras, Scroll of Stamina, Battle Shout growth) | Client values by level | 70338 (unchanged in 70425); shaman audit 10.2 |
| Debuff ranks (Curse of Weakness, Demoralizing Shout / Roar, Thunder Clap −20%, Scorpid Sting −2% hit, Curse of the Elements 30–75 / 4–10% on all magic schools) | Client values | 70338; shaman audit 10.4 |
| Curse of Recklessness | Armor only; 70425 dropped the target AP part | 70425 |
| Salvation | −30% threat, raid buff, no DPS effect | 70425 1038 aura 10 bp −30 misc 127 |
| Normalised Rage | k 3.46 one-hand / 4.5 two-hand × swing speed, level factor s = 0.5; off hand 0.5 | user rulings, standards 21 |
| Rage from damage taken | 20 × (1 − a(L)) × D / H(L); a = 20% → 40%; H = Forever normal-mob health (5356 at 60); D = the attacker's hit × crit/crush multiplier, before any of our mitigation | 8 Oct notes post 5; warrior article 24301514; magey issue 3; #57; standards 21 |
| Nature's Swiftness | CD starts when the buff is used (COOLDOWN_ON_EVENT 0x02000000) | c44c2502a |
| Pushback | Can't push a cast back past its start | 9f8d7ece5 |
| Partial resists | No level-based partial resist under Forever | user ruling, 172ad4632 |
| Mob crit vs us (leveling) | No base 5%, 0.2% per level over | logs, a950d79d7 |
| Continuous energy | Confirmed by the user | |
| Periodic crits | Attributes_8 0x200 marks the DoTs that can crit; SpellFlagNoPeriodicCrit opts out | pass 1 |
| Druid forms | Weapon DPS at speed 1.0 / 2.5; enchants and procs work in form; potions and runes too | article 24301515, client flags |
| Ahn'Qiraj book ranks | Excluded (standards 8a, IncludeAQRanks) | |
| Strongest learned rank | bestLearnedRank | 55c9744 |

## Settled (don't re-flag)

- Windfury Totem cooldown 1.5 sec (standards 17, user ruling). Windfury Totem procs from the main hand only (user ruling).
- Bane of Doom: with the user.
- Report assumptions decided by user rulings 2026-10-15 / 16: Trueshot Aura rank 5 100; Hunter's Mark rank 4 140; Arcane Shot 0.10 / Serpent Sting 0.03; hunter crit per Agility ×1.22; pet hits scale with speed; Raptor Strike rank 8 96; Rend 2% AP a tick; Frostbolt rank 10 ×1.04; Ice Lance 0.429; Lightning Bolt rank 10 ×1.12; warr_crit_rage +100%; warr_rage_oh 0.5; level_partial_resist none; sham_ftt_stack (raid Flametongue Totem with Windfury Weapon).

## Open items (not blocking)

- warr_dt_rage_mods / warr_dt_rage_k report assumptions: pending, waiting on the report re-sim. warr_dt_rage_k High must be about 48, not 36, now that H(60) = 5356.
- Berserker Rage's extra Rage from spell hits (sim/warrior/berserker_rage.go) still uses Classic's attacker-level conversion under Forever. No Level 60 preset casts it.
- Normal mob health above level 35 is extrapolated (+30% from level 20 on), and it feeds damage-taken Rage at 60.
- Class-side open assumptions (owned by the class workers): Paladin 2H Weapon Specialization on seal procs, Seal of Righteousness stored base points, Maelstrom Weapon proc rate, Omen of Clarity chance, Arcane Blast's bonus on Arcane Missiles.
