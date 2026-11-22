# Phase 0 gear step 2: mail and plate, 2026-11-22

Brief: `briefs/phase0_gear_step2_1122.md`. Slots: head, shoulders, chest, wrists, hands, waist, legs and feet in
mail and plate, for the 9 specs that wear them: ret, arms, fury, fury_dw, ele, enh, hunter_mm, hunter_bm, hunter_sv.

## How the candidates were found

The specs other than Enhancement have no list yet, so `RANK_ONLY=1` had no `FROM`. Each spec got a stub `FROM`
list: the report's synthetic gear as picks in every slot but the 8 armor slots, which start empty. Then
`RANK_ONLY=1 SLOTS=head,shoulder,chest,wrist,hands,waist,legs,feet` at low iterations (`-iters 300
-gear-rank-confirm-iters 1500`, `AOE_TARGETS=0`). That ranks each slot's items on their stats against a full set.
A spec took about 20 min on 4 cores. The stub and runner aren't in the repo (scratch only). Step 3's full searches
won't need them.

## Candidates (top 3 mail or plate per slot, Hyjal tier left out, * doubtful or infeasible)

| Spec | head | shoulder | chest | wrist | hands | waist | legs | feet |
|---|---|---|---|---|---|---|---|---|
| ret | Lionheart Helm*, Stalwart Helm, Helm of the Executioner | Justice Spaulders*, Defiler's Lamellar Spaulders*, Darkspear Epaulets | Onyxia Scale Breastplate*, Savage Gladiator Chain, Ogre Forged Hauberk | Battleborn Armbraces, Spike Crested Bracers, Slashclaw Bracers | Justice Gauntlets*, Heartflame Gauntlets, Raider Gauntlets* | Marksman's Girdle, Vindicator's Belt, Defiler's Lamellar Girdle* | Pristine Scorpid Leggings, Outrider's Plate Legguards*, Titanic Leggings* | Justice Greaves*, Bloodmail Boots, Depleted Thorium Sabatons* |
| arms | Lionheart Helm*, Exalted Helmet of the Tiger, Triumphant Skullcap of the Tiger | Darkspear Epaulets, Shoulders of Glory*, Defiler's Lamellar Spaulders* | Savage Gladiator Chain, Onyxia Scale Breastplate*, Ogre Forged Hauberk | Berserker Bracers*, Vambraces of the Sadist, Frozen Steel Vambraces* | Raider Gauntlets*, Gauntlets of Glory*, Vice Grips of the Tiger | Defiler's Chain Girdle*, Defiler's Lamellar Girdle*, Valiant's Waistguard | Cloudkeeper Legplates*, Outrider's Chain Leggings*, Titanic Leggings* | Greaves of Glory*, Scalegut Treaders, Exalted Sabatons of Strength |
| fury | Lionheart Helm*, Stalwart Helm, Helm of the Executioner | Shoulders of Glory*, Darkspear Epaulets, Pauldrons of Glory* | Savage Gladiator Chain, Onyxia Scale Breastplate*, Ogre Forged Hauberk | Battleborn Armbraces, Slashclaw Bracers, Berserker Bracers* | Gauntlets of Glory*, Raider Gauntlets*, Vice Grips of the Tiger | Brigam Girdle, Marksman's Girdle, Valiant's Waistguard | Titanic Leggings*, Outrider's Chain Leggings*, Outrider's Plate Legguards* | Bloodmail Boots, Windreaver Greaves, Greaves of Glory* |
| fury_dw | Lionheart Helm*, Stalwart Helm, Helm of the Executioner | Shoulders of Glory*, Hermit's Pauldron*, Lightforge Spaulders | Savage Gladiator Chain, Furbolg Rib Protector, Onyxia Scale Breastplate* | Battleborn Armbraces, Slashclaw Bracers, Spike Crested Bracers | Gauntlets of Glory*, Voone's Vice Grips, Raider Gauntlets* | Marksman's Girdle, Brigam Girdle, Valiant's Waistguard | Titanic Leggings*, Legplates of Glory*, Outrider's Plate Legguards* | Greaves of Glory*, Windreaver Greaves, Bloodmail Boots |
| ele | Five Thunders (Coif, Face, Crown) | Five Thunders (Pauldrons, Mantle, Spaulders) | Five Thunders (Vest, Chain, Tunic) | Five Thunders (Bindings, Bands, Bracers) | Gauntlets and Fists of The Five Thunders, Skyfury Gauntlets | Sash of the Windreaver, Masterwork and Mercurial Girdle of Nature's Wrath | Pristine Scorpid Leggings, Masterwork Legplates and Mercurial Legguards of Nature's Wrath | Five Thunders (Slippers, Treads, Greaves) |
| enh | Black Dragonscale Helm, Face of The Five Thunders, Blue Dragonscale Helm | Hermit's Pauldron*, Pauldrons of The Five Thunders, Darkspear Epaulets | Chain of The Five Thunders, Onyxia Scale Breastplate*, Furbolg Rib Protector | Slashclaw Bracers, Time-Forgotten Bindings, Bands of The Five Thunders | Gauntlets of The Five Thunders, Spiritcaller Gloves*, Skyfury Gauntlets | Girdle of The Five Thunders, Marksman's Girdle, Cord of The Five Thunders | Outrider's Chain Leggings*, Kilt of The Five Thunders, Spiritcaller Leggings* | Spiritcaller Treads*, Bloodmail Boots, Greaves of Withering Despair |
| hunter_mm | Black Dragonscale Helm, Backwood Helm, Triumphant Skullcap of the Falcon | Darkspear Epaulets, Defiler's Chain Pauldrons*, Bone Golem Shoulders | Furbolg Rib Protector, Atal'ai Breastplate of Agility, Ogre Forged Hauberk | Slashclaw Bracers, Marksman Bands, Satyrhide Armbands | Wildstalker's Gauntlets*, Mourning Handguards, Masterwork Gauntlets of Agility | Marksman's Girdle, Ranger's Belt, Warpwood Binding | Outrider's Chain Leggings*, Wildstalker's Legguards*, Leggings of the Plague Hunter* | Windreaver Greaves, Wildstalker's Greaves*, Bloodmail Boots |
| hunter_bm | as MM, plus Dinofrill Casque | as MM, Stratholme Militia Shoulderguard | as MM | Slashclaw Bracers, Satyrhide Armbands, Windtalker's Wristguards* | Voone's Vice Grips, Mourning Handguards, Gauntlets of Accuracy | as MM | as MM | as MM |
| hunter_sv | Black Dragonscale Helm, Backwood Helm, Crown of Tyranny | as MM | Onyxia Scale Breastplate*, Savage Gladiator Chain, Ogre Forged Hauberk | Slashclaw Bracers, Windtalker's Wristguards*, Satyrhide Armbands | Wildstalker's Gauntlets*, Mourning Handguards, Bloodmail Gauntlets | Marksman's Girdle, Defiler's Chain Girdle*, Ranger's Belt | as MM | as MM |

Ret, the warriors and the hunters also rank leather (Timbermaw, Dawn Armor, Belt of the Archmage) over some mail and
plate. Those are the cloth-leather part's.

## Effects

None of the mail or plate candidates has an unmodelled "Use:" or "Chance on hit:" effect. The only one in the pools,
8367 Dragonscale Breastplate (absorbs 600 magical damage, 1 hour cooldown), is defensive and only +10 Stamina. It got a
comment in `sim/common/item_effects.go` next to the other uses we leave out. No effect was modelled, so `EFFECTS_ONLY=1`
has nothing new to show.

## Stats

- The dry run's hidden or partial stat items in these slots are the Hyjal Summit tier sets (Glory, Justice, Spiritcaller,
  Wildstalker at item level 65, sets 2101 to 2111, 280893 to 280963). They are raid loot, so "out", and their stats
  don't matter before Dec 9.
- 14867 Warleader's Leggings, 19111 Winteraxe Epaulets, 19112 Frozen Steel Vambraces and 22764 Ironvine Belt aren't in
  Forever's client ItemSparse, and the server calls them invalid (`forever_client_cache.json`). No Forever stats, so
  doubtful. 23087 and 23090 (Undead Slaying) are out like their siblings.
- Most candidates are seen in game (`forever_seen_in_game.json`) or ship in the client, so their stats are Forever's.
  These are only revealed or in the client cache, so the sim's stats may be Classic's: the random-suffix world drops
  (Hyperion, Exalted, Masterwork, Mercurial, Magnificent, Triumphant), 9640, 11627, 11767, 13142, 13387, 14554, 16671,
  16675, 16677, 16681, 16723, 16729, 18722, 272108, 277117. None is a likely pick but Lightforge Spaulders (fury_dw #3).
  They aren't marked doubtful, since no source has other stats for them.

## Availability

`phase0_content.json` gained about 180 mail and plate entries (every top 10 mail or plate item of the 9 specs, the Hyjal
sets and the essence crafts). `phase0_enchants.json` gained 4: Gloves - Holy Power (Beneris, Honored), Boots - Minor
Speed, Bracer - Minor Agility, and Bracer - Minor Healing Power (doubtful, one drop sample).

Check: re-reading each spec's rank output against the files, no mail or plate item and no enchant in my slots is unlisted
among the picks and top 5s. The cloth and leather items there are the other part's.

## Needs a user call

- **Malleable Essence of Nature.** 72 Forever crafts (Glory, Justice, Spiritcaller, Wildstalker, Grovekeeper,
  Conviction, Manaflare, Demonheart, Grimstitch, item level 60) have plans sold by Aisarra Nightmeadow (Mount Hyjal) for
  6 or 8 essences (274030). We know of no source for the essence. Tranquil Wristguards were already "out" as "needs raid
  essence". I marked the 36 mail and plate ones doubtful, which includes the Spiritcaller pieces the Enhancement check
  had left open. Enhancement's legs pick (Spiritcaller Leggings) is now doubtful. If the essence comes only from the
  raids, all 72 should be "out". The cloth and leather ones are the other part's.
- Lionheart Helm, the head pick for ret and every warrior, is doubtful: plans drop at 0.2% or less (UBRS, Strat).
  Open if the plans are on the auction house.
- Titanic Leggings (warrior legs) and Depleted Thorium Sabatons have plans with no known source. Doubtful.
- Raider Gauntlets (272095, 272096) are infeasible on a guess (Exalted like the other Raider pieces). Frostwolf Plate
  Belt is open at Honored, though the Alterac Valley queue is unconfirmed.
