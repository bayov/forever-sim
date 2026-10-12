# Fairness report: Level 60 DPS presets (2026-10-12)

From the fairness auditor. The question this answers: is every DPS spec simmed under equally strong and equally representative conditions, so their numbers compare? Six passes over all nine class branches, 09:50 to 15:10 UTC. Final heads: warrior 82610f8, paladin e855d22, hunter d4554f1, rogue 6df047e, priest 3da596a, shaman 6bc4225, mage c2a9f7a, warlock 820c83b, druid 672e891.

Method: each class's shipped Level 60 build (gear, talents, APL), built with mksettings and run with wowsimcli. The auditor read code and logs and did not use rotopt. Measurements change one input at a time against the same seed. Two switches existed only in the auditor's local worktrees and were never pushed: one turns weapon racials off, one skips set-bonus tiers.

## 1. Cross-spec table

Settings: 180 sec, 5000 iterations, seed 1, standard error about 0.5 to 1. Each spec was swept over every race its class can play in Forever (Skyborne included; Windshaper is not a mage). The racials-off column is the best race and DPS with the weapon-type racials (Human sword, Orc axe, Dwarf mace) switched off.

| # | Spec | Best race | DPS | Best Horde / Alliance | Racials off |
|---|---|---|---|---|---|
| 1 | Demonology | Gnome | 1087.9 | Undead 1082.4 / Gnome | same |
| 2 | Destruction | Gnome | 983.9 | Undead 983.7 / Gnome | same |
| 3 | Affliction | Gnome | 977.0 | Orc 973.7 / Gnome | same |
| 4 | Destruction Fire | Undead | 976.6 | Undead / Gnome 976.2 | same |
| 5 | Retribution | Undead | 961.9 | Undead / Human 946.6 | same |
| 6 | Fury (dual wield) | Undead | 943.2 | Undead / Human | same |
| 7 | Marksmanship | Night Elf | 911.9 | Orc / Night Elf | same |
| 8 | Frostfire | Undead | 901.8 | Undead / High Order 901.4 | same |
| 9 | Beast Mastery | Night Elf | 881.1 | Orc / Night Elf | same |
| 10 | Fire | Undead | 869.1 | Undead / Gnome 865.4 | same |
| 11 | Shadow | Undead | 862.5 | Undead / Gnome 858.5 | same |
| 12 | Arms | Undead | 831.4 | Undead / Human 825.4 | same |
| 13 | Fury (two-hand) | Undead | 825.6 | Undead / Human 821.7 | same |
| 14 | Arcane | Undead | 823.0 | Undead / Gnome 813.1 | same |
| 15 | Balance | Night Elf | 807.6 | Tauren 800.1 / Night Elf | same |
| 16 | Enhancement | Orc | 791.9 | Orc / Dwarf 774.2 | Orc 784.0 |
| 17 | Survival | Human | 778.3 | Orc / Human | Night Elf 773.5 |
| 18 | Frost | Undead | 743.4 | Undead / High Order 742.8 | same |
| 19 | Elemental | Orc | 690.8 | Orc / Dwarf 684.2 | same |
| 20 | Rogue Combat Daggers | Undead | 690.2 | Undead / Gnome 677.0 | same |
| 21 | Smite | Undead | 684.9 | Undead / Gnome 679.9 | same |
| 22 | Rogue Mutilate | Undead | 658.2 | Undead / Gnome 641.7 | same |
| 23 | Rogue Seal Fate Daggers | Undead | 657.0 | Undead / Gnome 645.7 | same |
| 24 | Rogue Combat Swords | Undead | 653.1 | Undead / Human 648.3 (635.8 off) | same |
| 25 | Feral | Night Elf | 648.6 | Windshaper 646.8 / Night Elf | same |
| 26 | Rogue Subtlety | Undead | 627.1 | Undead / Orc 615.8 | same |
| 27 | Rogue Seal Fate Swords | Undead | 619.2 | Undead / Human 616.6 | same |

The Mutilate + Venom build was dropped after this sweep (7a6b0d4): Venom loses DPS at every combo point count as modelled.

Every preset's race matches the sweep. The one exception is Destruction Fire, where Gnome and Undead are 0.4 apart, within noise. Weapon racials now move only two specs:
- Enhancement: the Orc axe is worth 7.9 DPS.
- Survival: the Human sword is worth 14.4 DPS, so with racials off Night Elf is best.

Every caster uses a staff, so no caster gets a weapon racial. The results auditor's table matches this one (it reproduces every class's published numbers).

## 2. What is on equal footing (the standards in orchestration/standards.md)

| Area | Standard | Status |
|---|---|---|
| Armor budget | One item-level table: tier slots 66; neck, back, wrist, waist and rings 71 (402.5) | Every set between 397.2 (Feral) and 403.6 (rogue daggers); gear auditor pass 4 |
| Weapon budget | Two-handers 54.8 (item level 76). Physical one-handers about 20.5. Caster main hand plus off hand, or staff, about 73 | All sets |
| Ranged slot | A stat wand or bow (about 15.6), or a coded Forever relic | Libram of Law, Burning Totem, Swarming and Howling Idols |
| Hit | Each set optimised at budget prices, in practice at its spec's cap | Casters 10% gear plus 2-piece (Demo 14%, Balance 11%, Smite 0% via Holy talents); melee 9%; SV 6% + Surefooted |
| Set bonuses | 2-piece only. 4- and 5-piece values reported separately | All |
| Trinkets | None (Forever removed the dungeon ones; Talisman of Ephemeral Power is Molten Core) | All |
| Enchants | Best Forever Phase 1 enchant per slot, Enchanting 300 included | All. Head is Lesser Arcanum of Voracity for everyone, and nobody uses the full Dire Maul Arcanums |
| Crit elixirs | At most one of Mongoose, Grizzly, Owl, Sages, Whale | All |
| Consumables | Forever Phase 1 sources only, best per slot | All. The hunter merged the consumes branch in pass 6 |
| Raid | Shared main.go raid. Melee groups get Windfury, Strength of Earth and Battle Shout. Casters lose Might and gain Mana Spring | Consistent. Under Forever, Moonkin and Leader of the Pack are a shared 3% crit; Trueshot is ranged only; improved totems are neutralised |
| Spell and buff ranks | No rank taught only by an Ahn'Qiraj book | All classes use env.IncludeAQRanks(); raid buffs capped at Battle Shout VI, Strength of Earth IV, Grace of Air II |
| Pre-pull | Any class may pre-buff or pre-cast; damage counts from 0 | Every pre-cast lands at 0 or later (Frostfire Bolt +0.5, Fireball +1.0, Shadow Bolt +0.2, Lightning Bolt +0.5, Starfire 0, Arcane Blast 0, Mind Blast and Smite at 0 or later) |
| Stealth openers | No start in Stealth (user ruling) and no Prowl opener | Rogue opens without Stealth; Feral's Prowl opener was reverted (8bc55c6) |
| Pets | No player buffs and no pet consumables, for hunter pets and demons alike. Both inherit their owner's hit | Hunter 0d2cf5e; warlock 4742e1a |
| Encounter | Level 63, 3731 armor, Humanoid; 120/180/300 sec; same execute phases; reaction time 200 ms | Shared main.go, unchanged on every branch |
| Professions | Engineering for explosives; the second profession has no DPS effect except Tailoring for crafted sets | Feral has no explosives (cat form) |

## 3. Sensitivities: placeholders that move the ranking

| Item | Effect | Why it stays |
|---|---|---|
| Warlock demon scaling | Demonology 1022 to 1098 across the documented range (no owner scaling up to the Season of Discovery placeholder). The placeholder adds about 50 to 60 on demon-out builds | Forever's figures are server side and unmeasured. Sourced Season of Discovery figures: 0.5 x spell power as attack power, 25% as spell damage, hit 100% spell / 80% melee, no crit. **The warlock lead (Demo about 13% over Ret) rests partly on this.** |
| Bane of Doom coefficient | 4.0 in client 70338 (Era 1.0), never playable in the beta. At 2.0 every Shadow Bolt warlock build loses about 4% (rank tuning audit, orchestration/rank_tuning.md) | It is the current client value |
| Hunter pet scaling | On Season of Discovery's 22% of ranged attack power instead of Forever's measured 10%: BM +58, MM +32, SV +28 | The 10% is a Forever beta measurement and ranks above Season of Discovery in the source order |
| Pet consumables | +17 BM if Forever lets owners use Juju or scrolls on pets | Ruled off (item 14): consumables on a pet count as a player buff |
| 4- and 5-piece set bonuses | Off in presets. If on: Frostfire 5-piece +6.0%, Shadow 5-piece +5.0%, BM +2.0%, MM +1.7%, Feral +1.7%, others 0 to 1.3%. Warrior and rogue 4-piece (+36 AP vs Humanoids) would be live on the default boss (about +1.1%) | Forever Tier 1 pieces have no stats and no known source yet |
| Demon in melee (warlock) | Reported by the warlock with halved melee as its own range | Placeholder |

## 4. Inherent Forever asymmetries (not fixed, by design)

- **Gloves school power:** Forever Phase 1 has +20 gloves only for Holy, Arcane and Nature. Shadow, Fire and Frost casters (Shadow, every warlock, Frostfire, Fire, Frost) get only the +6 spell power kit, worth about 0.3 to 0.5%.
- **Feral in cat form** can't use weapon oils, stones, explosives or Windfury. This is structural.
- **Rogue deficit** (results auditor, client-confirmed): Forever's rogue nerfs (-59), the warrior's off-hand rage (+44 on the comparison side) and no Execute or Heroic Strike equivalent.
- **Elemental and Smite are low** from client coefficients: Lightning Bolt 196 base at 0.714, Smite 126 at 0.714, against Fireball 483 at 1.0.

## 5. What is not on equal footing

Nothing that the standards cover. Every class applies every standard on its final head. The only remaining differences are the sensitivities in section 3 and the inherent asymmetries in section 4.

## 6. History (what the passes changed, largest first)

- Crit elixirs: casters and hybrids drank Owl and Sages (+4% crit) while pure melee did not. Settled by the one-crit-elixir rule.
- Caster hit: casters were 4 to 6% under the spell hit cap. Trading spell power for hit was worth +0.4 to +5.2%. All now at cap.
- Caster weapon slots: a staff was 25 points behind main hand plus off hand, about 2.8% for Ele and Balance. Equalised at about 73.
- Armor item levels: the spread was 382 to 413 from each class's Era best-in-slot item levels. One table now.
- Set bonuses: 0 to 7.3% spread, with the warrior 4-piece live on the Humanoid boss. Now 2-piece only.
- Weapon racials: the Human sword gave 1.4 to 2.3% on casters with Azuresong. Casters on staves now; a racials-off column was added.
- Missing items:
  - Ret's Libram of Fervor did nothing; Libram of Law is now coded.
  - Druids had empty relic slots; Swarming and Howling Idols are now coded.
  - Neck enchants were missing on five classes.
  - Arcanum of Focus (Dire Maul, phase 2 in the database) was on Ele and Warlock.
  - Several pre-casts were missing: Arcane, Balance, Ele, Shadow, Smite.
- Hunter: hadn't merged the shared consumables until pass 6. Pet hit inheritance is now added.
- Race columns: corrected from preset defaults to real sweeps. Warlock moved to Gnome; mage moved to Undead.
