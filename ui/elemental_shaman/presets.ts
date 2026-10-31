import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	FirePowerBuff,
	Flask,
	Food,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	MobType,
	NaturePowerBuff,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	SpellPowerBuff,
	SpiritElixir,
	Stat,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { ElementalShaman_Options as ElementalShamanOptions } from '../core/proto/shaman.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import DefaultAPLJson from './apls/default.apl.json';
import Level60APLJson from './apls/level60.apl.json';
import Level60ManaTideAPLJson from './apls/level60_mana_tide.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import Phase1GearJSON from './gear_sets/phase_1.gear.json';
import Phase2GearJSON from './gear_sets/phase_2.gear.json';
import Phase3GearJSON from './gear_sets/phase_3.gear.json';
import Phase4GearJSON from './gear_sets/phase_4.gear.json';
import Phase5GearJSON from './gear_sets/phase_5.gear.json';
import Phase6GearJSON from './gear_sets/phase_6.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearPhase1 = PresetUtils.makePresetGear('Phase 1', Phase1GearJSON, { group: 'Era' });
export const GearPhase2 = PresetUtils.makePresetGear('Phase 2', Phase2GearJSON, { group: 'Era' });
export const GearPhase3 = PresetUtils.makePresetGear('Phase 3', Phase3GearJSON, { group: 'Era' });
export const GearPhase4 = PresetUtils.makePresetGear('Phase 4', Phase4GearJSON, { group: 'Era' });
export const GearPhase5 = PresetUtils.makePresetGear('Phase 5', Phase5GearJSON, { group: 'Era' });
export const GearPhase6 = PresetUtils.makePresetGear('Phase 6', Phase6GearJSON, { group: 'Era' });

// A guess at Phase 1 BiS under Forever, made of the synthetic items in
// tools/database/forever_synthetic_items_shaman.go (ids from 990101, not in the game).
//
// Forever's level 60 gear in the client is not tuned yet, and Molten Core and Onyxia have no
// Forever data. The synthetic items take the place of the Era Phase 1 elemental pieces (the
// Earthfury set, Choker of the Fire Lord, Deep Earth Spaulders, Robe of Volatile Power, Mana
// Igniting Cord, Ring of Spell Power and Staff of Dominance), built on the cross-class standard
// (orchestration/standards.md, item 1b since the 2026-10-13 gear round): Forever's budget for a
// pure caster (each armor piece carries slot mod * (1.125 * item level + 2.0) * 1.053 / 1.026 /
// 0.995 in weighted stats, spell power at 0.86, 4 stat lines a piece by the 1.7-norm shape check),
// Forever's doubled caster weapon (the staff carries 193.4), the Elemental mix of gear_targets.md
// (15% Stamina, 18% Intellect, 10% Spirit, 35% spell power, 11% crit) and hit to the spell cap
// against a level 63 boss with Tidal Focus 5 and the 2 piece bonus. The armor is 709.3 and the staff
// 193.3 weighted points. The hands and legs carry The Spiritcaller's Storm, Forever's elemental Tier 1
// set, so only its 2 piece bonus (+1% hit) is live. The 5 piece bonus (Lava Burst 1 sec sooner) would
// add more. The file has the budget math, the shares and the shape check per piece.
//
// The enchants are the best a Phase 1 shaman can get outside the raids: Enchant 2H Weapon -
// Spellblasting (Forever, +65 spell power), Gloves - Natural Power, Bracer - Greater Spellpower,
// Mystic Rugged Armor Kits on the chest, legs and feet, Lesser Arcanum of Voracity (Intellect)
// on the head and Forever's Enchant Necklace - Spell Power. Revelation on the staff was 1.3 ahead
// of the old Enchant Weapon - Spell Power (Molten Core only) and Spellblasting about 18, so
// Revelation is about 17 behind. The relic is Burning Totem (Flame Shock 3 sec longer),
// 7.1 ahead of no relic and 4.8 ahead of Polished Driftwood Icon.
//
// Two synthetic item level 71 trinkets (standards 8a and 1b): Forever removed the old dungeon
// trinkets and the sim doesn't model its launch ones, so each spec gets a stat trinket (37 spell
// power and 14 crit rating) and an on-use one (140 spell power for 20 sec every 2 min, the weight of
// Talisman of Ephemeral Power). Both are synthetic, not in the game.
//
// The Level 60 build is 674.3 / 673.9 / 663.3 at 120 / 180 / 300 sec on it (663.7 at 300 before the
// downranking penalty). Before the gear round
// it was 648.7 / 646.1 / 638.1 (the 1a set, no trinkets) and 693.3 / 690.3 / 680.4 before the fair
// mix. The set stays fixed, it was not gear searched.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Elemental', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic, GearPhase1],
	[Phase.Phase2]: [GearPhase2],
	[Phase.Phase3]: [GearPhase3],
	[Phase.Phase4]: [GearPhase4],
	[Phase.Phase5]: [GearPhase5],
	[Phase.Phase6]: [GearPhase6],
};

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLDefault = PresetUtils.makePresetAPLRotation('Default', DefaultAPLJson, { group: 'Era' });

// The tools/rotopt shaman_ele search on the Level 60 talents, over the mean of 120, 180 and
// 300 sec: Mana Spring and Searing Totem, Flame Shock whenever its DoT is down, Lava Burst on
// cooldown unless Flame Shock runs out before the cast ends (+1.3 to +1.8), Lightning Bolt, and
// Lightning Bolt rank 8 below 20% mana (+7.6 at 300 sec over the top rank; rank 7 until Forever's
// downranking penalty, which leaves it 75% of its coefficient, cost it 11.6 at 300 sec). The fight opens with Mana Spring
// at -3.9 sec, a Lightning Bolt at -2.9 (it lands 0.1 sec into the fight) and Searing Totem at
// -0.9 (its first attack comes after the pull): +10.6 / +7.0 / +4.8 at 120 / 180 / 300 sec over
// putting the two totems down with no bolt. Chain Lightning loses on one target even above 80%
// mana (490 damage a second for 485 mana, against Lightning Bolt's 479 for 198). Earth Shock and
// mana floors on the shocks changed nothing as rotation spells, but an Earth Shock at the end of the fight,
// once a Lightning Bolt would no longer land in time, is +5.8 / +3.4 / +1.7 (the rotation auditor,
// 2026-10-15; +5.7 / +3.4 / +1.8 on Mana Tide). A caster stands at range, so the rotation has no
// Magma Totem or Fire Nova. tools/rotopt/shaman_presets.sh writes it.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Elemental', Level60APLJson, {
	tooltip: 'The tools/rotopt search result for the Elemental talents.',
	group: 'Synthetic',
});
// The same with Mana Tide Totem at 20% mana. At 40, 60 and 80% it was 1.1, 2.7 and 5.7 behind.
export const APLLevel60ManaTide = PresetUtils.makePresetAPLRotation('Elemental Mana Tide', Level60ManaTideAPLJson, {
	tooltip: 'The Elemental rotation with Mana Tide Totem, for the Elemental Mana Tide talents.',
	group: 'Synthetic',
});

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60, APLLevel60ManaTide, APLDefault],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultAPL = APLLevel60;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search (tools/rotopt -talent-search) on the synthetic set, from four seeds (deep
// Elemental with Restoration, with Enhancement, and two mixes), then the knob search, then the
// talent search again from three seeds on the new knobs, as the mean of 120, 180 and 300 sec. It
// ran again on the standard gear and consumes from two seeds and kept the build. Thundering
// Strikes 5 and Tidal Focus 5 (5% hit, Forever took Nature's Guidance away) are the big ones. On
// the earlier gear the next best build moved 4 points into Ancestral Knowledge and Shamanistic
// Focus out of Mindfulness and Tidal Focus and was 15 behind.
//
// Client 1.60.1.70425 removed Call of Thunder, moved Lava Burst to the 20-point row and made
// Elemental Mastery the 30-point talent. The talent search (from the old build without Call of
// Thunder, on the 2 raid hits a minute that now push casts back) took Elemental Mastery first
// (+22 as the mean of 120, 180 and 300 sec), moved 3 Reverberation points into Eye of the Storm
// (+8: it keeps 70% of the casts from being pushed back), and the Restoration points into
// Ancestral Knowledge, Shamanistic Focus and one Improved Lightning Shield. 731.3 / 721.4 /
// 712.6 at 120 / 180 / 300 sec.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Elemental', SavedTalents.create({ talentsString: '5501331003001351-055000101-053' }), {
	tooltip: 'Level 60 talents from the talent search: Elemental Mastery, Lava Burst, Eye of the Storm, Thundering Strikes and Tidal Focus.',
	group: 'Synthetic',
});
// Mana Tide Totem for the group's mana: Totemic Focus 5, Mindfulness 3, Tidal Focus 5 and the
// Mana Tide row, with one point less in Thundering Strikes and none in Ancestral Knowledge. The
// talent search with Mana Tide Totem pinned found it, and found it again on the standard gear.
// Its Water Shield point does nothing in the sim: the rotation keeps Lightning Shield up instead
// (see the Options note below). Under client 1.60.1.70425 it takes Elemental Mastery for Call of
// Thunder and Eye of the Storm for 3 Reverberation points, as Level 60 does (the talent search
// with Mana Tide Totem pinned). 721.3 / 711.2 / 697.4 at 120 / 180 / 300 sec, 1.4% behind
// Level 60 at 180.
export const TalentsLevel60ManaTide = PresetUtils.makePresetTalents(
	'Elemental Mana Tide',
	SavedTalents.create({ talentsString: '5501331003001351-04-053150001001' }),
	{ tooltip: 'Level 60 talents with 16 Restoration points for Mana Tide Totem.', group: 'Synthetic' },
);

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60ManaTide],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultTalents = TalentsLevel60;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// Lightning Shield and the raid damage in the Level 60 encounter (PresetUtils.LEVEL_60_RAID_DAMAGE,
// 2 direct hits a minute since 2026-10-16, when the raid hits moved to the encounter for every
// spec; the numbers below were taken at Elemental's own rate of 1 a minute). Each hit fires a
// Lightning Shield orb, so both Level 60 rotations keep the shield up. It goes before the pull at
// -5.4 sec (casting it in the fight instead lost 6), and in the fight just ahead of the low rank
// Lightning Bolt. At 180 sec Level 60 went from 690.6 to 694.0 and Mana Tide from 683.3 to 686.5
// (about +13 at 10 a minute), and it costs 0.4 with no hits at all, for the mana. Water Shield's
// 2% mana a globe was 20 behind at 6 a minute on Mana Tide: the GCDs cost more than the mana is
// worth.
//
// Since the rotation auditor's pass 21 (2026-10-16) the fight recast waits for half the mana and a
// minute of the fight left: at 1 hit a minute 3 orbs don't repay the GCD and the mana otherwise.
// Dropping the recast was a little ahead at 1 a minute but 5.6 behind at 10. Mana Spring is not
// recast with less than 20 sec left either. Level 60 is 699.0 / 695.3 / 679.7 at 120 / 180 / 300
// sec (+1.4 / +1.3 / +3.0) and Mana Tide 692.3 / 688.1 / 673.7 (+1.4 / +1.7 / +1.7).
//
// Measured again at 2 hits a minute with pushback, on the client 70425 builds (20000 iterations,
// 120 / 180 / 300 sec), the gate still wins or ties: Level 60 731.3 / 721.4 / 712.6 with it,
// 729.0 / 719.6 / 711.5 recasting whenever the shield is down, 729.0 / 719.7 / 712.3 with only the
// mana half, 731.4 / 720.7 / 711.1 with no recast in the fight. Mana Tide 721.3 / 711.2 / 697.4,
// 718.8 / 709.4 / 695.6, 719.0 / 710.0 / 697.3 and 721.5 / 710.9 / 697.0.
//
// The rotation auditor's pass 23 (2026-10-16): no new Flame Shock with less than 9 sec left (a DoT
// put up at the very end ticks too few times for its mana and GCD), and on Mana Tide the totem at
// 10% mana instead of 20%. Paired at 20000 iterations on two seed sets: Level 60 781.6 / 771.8 /
// 761.7 at 120 / 180 / 300 sec (+2.7 / +0.7 / +0.8), Mana Tide 770.8 / 761.4 / 744.3 (+2.6 / +1.1
// / +0.3). 3, 4.5, 6 and 12 sec gates were +0.1 to +0.6 on Level 60.
export const DefaultOptions = ElementalShamanOptions.create({});

// The level 60 consumables, Forever's Phase 1 ones (the shared consumables branch). On 2026-10-12
// we tried each candidate in its slot on the Level 60 build, as the mean DPS at 120, 180 and 300
// sec. On the earlier gear: Elixir of Nature Power +12.9, Elixir of the Sages +10.4 (2% crit),
// Lobster Stew +2.9 over Nightfin Soup, Greater Mageblood Elixir +0.9, Goblin Sapper Charge +4.3,
// Goblin Land Mine +1.2, Demonic Rune 9.6, Flask of Supreme Power 71 over Flask of Distilled
// Wisdom, Cerebral Cortex Compound 3.4, Elixir of Firepower 1.6 and Wizard Oil 11.9. Major
// Spellblasting Potion loses to the Major Mana Potion (-6). A character takes one of the five 2%
// crit elixirs (the standard): the Sages with Elixir of Greater Intellect, a tie with the Owl
// and Elixir of Greater Spirit (0.4 behind). If Forever lets them stack, the Owl on top of the
// Sages would add about 10. Brilliant Wizard Oil would add 10.8, but its recipe is from Zul'Gurub.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	// The Caustic Smog Potion (range 20) is +4.4 over the Major Mana Potion at 180 sec. Mana Tide
	// takes the Major Mana Potion (ManaTideConsumes).
	defaultPotion: Potions.CausticSmogPotion,
	// No filler explosive (standards 11, 2026-10-23: explosives only when they gain). The Goblin Land
	// Mine's 1 sec cast takes a second of casting: dropping it is +1.0 / +0.7 at 180 sec on seeds 1 and
	// 1000003 (+0.6 at 120 sec, -3.0 at 300 sec; Mana Tide +0.8 / +0.5, ahead at every length). The
	// EZ-Thro Dark Bomb ties the Land Mine (the same 450 average) and the EZ-Thro Thorium Grenade is
	// 0.8 behind it. The Goblin Sapper Charge, which needs no cast, stays (+4.2 at 180 sec).
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	// The fight is in a natural flask zone (standards 11), but the best natural flask, Aggression, is 74
	// behind the Flask of Supreme Power at 180 sec (Swiftness 78, Accuracy and Precision 95, the hit is
	// capped).
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.WizardOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	naturePowerBuff: NaturePowerBuff.ElixirOfNaturePower,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// Mana Tide runs shorter on mana, so it drinks the Major Mana Potion: -4.3 / +1.8 / +27.2 at 120 /
// 180 / 300 sec over the Caustic Smog Potion (20000 iterations, two seed sets, 2026-10-23).
export const ManaTideConsumes = Consumes.create({
	...DefaultConsumes,
	defaultPotion: Potions.MajorManaPotion,
});

// The level 60 raid (tools/rotopt/mksettings -spec ele): the shared level 60 raid buffs, Kings
// and Wisdom from the paladins, Judgement of Wisdom, and Curse of the Elements as the one debuff
// on the shaman's schools. Under Forever Improved Scorch and Winter's Chill only help the mage's
// own spells, and Curse of Shadow is gone. The shaman is the only one in its group, so its
// totems are the ones the rotation puts down (standards 23: every shaman drops its own totems; a
// caster party's only DPS totem is Mana Spring).
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	leaderOfThePack: true,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
	salvation: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	moldarsMoxie: false,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	curseOfRecklessness: true,
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	demoralizingRoar: TristateEffect.TristateEffectImproved,
	demoralizingShout: TristateEffect.TristateEffectImproved,
	faerieFire: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	scorpidSting: true,
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

// The shaman is an engineer for the Goblin Sapper Charge (+4.2 at 180 sec) and an alchemist. The
// race is Windshaper Skyborne, the best race since client 1.60.1.70425 gave Wind Blessed 3% haste
// (721.4 for the Orc at 180 sec before), see the races under the builds.
//
// The distance and the channel clip delay are the ones the DPS Report sims with.
export const OtherDefaults = {
	distanceFromTarget: 20,
	channelClipDelay: 50,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceSkyborneWindshaper,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor), a 3 minute fight,
// and the default buffs, debuffs and consumes, as for the enhancement shaman. The synthetic
// Phase 1 set, the talents and the rotation were searched at 120, 180 and 300 sec.
export const EncounterLevel60 = PresetUtils.makePresetEncounter(
	'Generic boss',
	Encounter.create({
		duration: 180,
		durationVariation: 20,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		...PresetUtils.LEVEL_60_RAID_DAMAGE,
		...PresetUtils.LEVEL_60_ZONE,
		targets: [
			{
				id: 213336,
				name: 'Level 60',
				level: 63,
				mobType: MobType.MobTypeUnknown,
				stats: new Stats().withStat(Stat.StatAttackPower, 805).withStat(Stat.StatArmor, 3731).withStat(Stat.StatHealth, 127393).asArray(),
				minBaseDamage: 3000,
				damageSpread: 0.3333,
				swingSpeed: 2,
				parryHaste: true,
			},
		],
	}),
	{
		tanks: [],
		raidBuffs: DefaultRaidBuffs,
		debuffs: DefaultDebuffs,
		buffs: DefaultIndividualBuffs,
		consumes: DefaultConsumes,
	},
);
export const EncounterLevel60ManaTide = PresetUtils.makePresetEncounter('Elemental Mana Tide', EncounterLevel60.encounter, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: ManaTideConsumes,
});

// On 2026-10-13, on the gear round's set (standards 1b, with the two synthetic trinkets) and the
// standard consumes (20000 iterations), Level 60 is 674.3 / 673.9 / 663.3 at 120 / 180 / 300 sec and
// Level 60 + Mana Tide 667.7 / 666.9 / 656.7 (with the downranking penalty and the rank 8 fallback). The talent and knob searches ran again on the new set and
// kept both builds. On the 1a set they were 648.7 / 646.1 / 638.1 and 642.4 / 639.3 / 631.3.
//
// The races (same settings and seed): Orc 674.3 / 673.9 / 663.3, Windshaper Skyborne 673.2 / 672.4 /
// 661.8, Troll 671.2 / 669.1 / 659.2, Dwarf (the only Alliance shaman) 667.8 / 666.8 / 656.8, Tauren
// 667.8 / 666.7 / 656.7. Best overall and best Horde: Orc, 1.5 ahead of Windshaper Skyborne at 180 sec.
// Best Alliance: Dwarf (the only Alliance shaman). On Mana Tide: Orc 666.9, Windshaper Skyborne 665.2,
// Troll 662.0, Dwarf 659.8, Tauren 659.8 at 180 sec. The staff is the Era Phase 1 weapon type and no
// race has a staff racial; Tauren's 1% hit is wasted at the hit cap.
//
// Client 1.60.1.70425 (2026-10-16, the 70425 talents, 2 raid hits a minute, 20000 iterations, 180
// sec): Windshaper Skyborne 724.9 (Wind Blessed is 3% haste now), Orc 721.4 (Savage's 1% crit
// counts with the staff now), Dwarf 719.3 and Tauren 719.1 (2% crit damage each), Troll 710.1.
// Best overall and best Horde: Windshaper Skyborne. Best Alliance: Dwarf. On Mana Tide:
// Windshaper Skyborne 714.6, Orc 711.2, Dwarf 708.7, Tauren 708.6, Troll 700.1.
//
// The regear round (2026-10-23, 20000 iterations, seeds 1 and 1000003, the natural flask zone on): the
// set stays at the spell cap, and the knob and talent searches kept both builds. Level 60 is 786.1 / 776.4
// / 740.5 at 120 / 180 / 300 sec; Mana Tide 770.9 / 761.4 / 744.3 with its Major Mana Potion (775.2 /
// 759.6 / 717.0 with the Caustic Smog Potion). Races at 180 sec (seed 1; seed 1000003 agrees within 0.8):
// Windshaper Skyborne 776.4, Orc 773.2, Dwarf 771.0, Tauren 770.8, Troll 761.4; on Mana Tide 761.4,
// 757.4, 754.6, 754.5, 745.8. Best overall and best Horde: Windshaper Skyborne. Best Alliance: Dwarf.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Elemental', {
	group: 'Synthetic',
	tooltip: 'Level 60 Windshaper Skyborne on the synthetic Phase 1 gear with the Elemental rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceSkyborneWindshaper,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
export const PresetBuildLevel60ManaTide = PresetUtils.makePresetBuild('Elemental Mana Tide', {
	group: 'Synthetic',
	tooltip: 'Level 60 with the Mana Tide talents and rotation.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60ManaTide,
	rotation: APLLevel60ManaTide,
	encounter: EncounterLevel60ManaTide,
	race: Race.RaceSkyborneWindshaper,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
