import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	Explosive,
	FirePowerBuff,
	Flask,
	Food,
	IndividualBuffs,
	MobType,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SpellPowerBuff,
	Stat,
	StrengthBuff,
	TristateEffect,
	UnitReference,
	UnitReference_Type,
	ZanzaBuff,
} from '../core/proto/common.js';
import { PaladinAura, PaladinOptions as RetributionPaladinOptions, PaladinSeal } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import APLBasicRetJson from './apls/basic_ret.apl.json';
import Level20APLJSON from './apls/level20.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import Level20GearJSON from './gear_sets/level20.gear.json';
import Level20AoEGearJSON from './gear_sets/level20_aoe.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);
// The tools/rotopt gear search at level 20 over everything an Undead paladin can equip
// (Scourge Invasion drops left out), on the wowhead Forever item data.
//
// The weapon is Wolfsbane, the two hander at the end of the Undead paladin quest chain,
// and it is worth 4 DPS on its own: 69-105 over 3.4 sec where Hammerbone out of Wailing
// Caverns swings 55-84, and the sim does not even run its Holystorm proc. Most of the
// armor is the Totemic Leather set the Leatherworker makes for themselves, and the rest
// is Forever's own spell power gear: Spellpower Goggles Xtreme, Technician's Bracers from
// A Fine Mess in Gnomeregan, Fletcher's Gloves, and Chestnut Mantle from Allegiance to the
// Old Gods.
//
// The neck and rings are new Forever items. Mark of the Pack Leader drops off Humar the
// Pridelord, a rare in the Barrens (ForeverChanges has it from player reports it has not
// checked). Philanthropist's Ring is the Greater Friend of the Library reward for 20
// books, four of them in level 30 zones. First Mate Band drops off Mr. Smite in the
// Deadmines. Together they are 1.5 DPS over the Warsong Gulch neck and ring and the Coral
// Band.
//
// Mail is open to a paladin at 20, and the search does look at the mail ForeverChanges
// lists (Veteran's Silvered Chain Helm, Grip of Fear, Vilewalkers, Juggernaut
// Leggings). It keeps the leather, because the Totemic pieces carry spell power for
// Consecration. Four pieces of Defias Leather are 0.2 ahead on one target but 2 to 6
// behind once there is a second one, so we leave that set to the rogue.
//
// Fletcher's Gloves are 14 crit rating, 4.3% crit at level 20 if rating scales with level
// the way we assume (see core.RatingPerPercent), and Forever pays that crit on Seal of
// Command and Consecration too.
//
// The enchants are the best a level 20 can get, Enchanting 225 at most, and are worth 9
// DPS on one target and 15 on three: +15 Strength on Wolfsbane, +7 Strength on the
// gloves, and spell power everywhere else there is a choice (the Spell Power necklace
// enchant, Mystic armor kits, Lesser Healing Power on the bracers for its 6 spell damage).
//
// The set assumes Engineering and Leatherworking, which is 4.5 DPS. A level 20 head slot
// has nothing in it but the goggles, so that is most of the gap.
export const GearLevel20 = PresetUtils.makePresetGear('Level 20', Level20GearJSON);
// The same search against three enemies, where Consecration is most of the damage and a
// point of spell power is worth four of Strength. Wolfsbane stays, it is far enough ahead
// of every other weapon, and the slots the Totemic Leather set does not win take a spell
// power piece instead (Heavy Woolen Cloak, Stormrider's Leather Belt, Dark Ritual
// Leggings out of Blackfathom Deeps, Advisor's Ring), and the gloves take a Mystic Thick
// Armor Kit instead of Strength. Worth 3.1 DPS on four targets and 1.5 on three at 60 sec,
// even on two and 1.7 behind on one.
export const GearLevel20AoE = PresetUtils.makePresetGear('Level 20 AoE', Level20AoEGearJSON);

export const GearPresets = {};

export const DefaultGear = GearBlank;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLBasicRet = PresetUtils.makePresetAPLRotation('Basic Ret', APLBasicRetJson);
// The paladin_ret template with the knobs the level 20 search settled on (see
// tools/rotopt/paladin_presets.sh): Seal of the Crusader judged once for the fight, Seal
// of Command judged on every cooldown, Holy Strike on cooldown, Consecration down to 20%
// mana.
export const APLLevel20 = PresetUtils.makePresetAPLRotation('Level 20', Level20APLJSON);

export const APLPresets = {
	[Phase.Phase1]: [],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [APLBasicRet, APLLevel20],
	[Phase.Phase5]: [],
};

export const DefaultAPL = APLPresets[Phase.Phase4][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/forever/talent-calc and copy the numbers in the url.

// Forever trees (beta client build 1.60.1.69876). Level 60: Improved Holy Strike, Divine
// Strength and Improved Seals in Holy, Toughness and Precision in Protection, the
// Retribution tree down to Twist of Light. Not searched yet, the level 60 sets are blank.
export const P4RetTalents = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '250003-51300-052053310012330301' }));
// Level 20 (11 points), the best of every 11 point build over the damage talents: Seal
// of Command with Deflection 5, Improved Judgement 1, Holy Conduit 2 and Conviction 2 in
// front of it. Deflection counts because the encounter has the enemies hitting the
// paladin, and every parry hastes the next swing. The builds behind it are within 0.5
// DPS (Benediction 5 in place of Deflection is 0.9 behind), Seal of Command itself is
// worth about 5. The same build wins against two, three and four enemies at 60 sec. Over
// two minutes of a pack the mana runs dry and Benediction 2 with Holy Conduit 3 gets 1.2
// ahead.
export const TalentsLevel20 = PresetUtils.makePresetTalents('Level 20', SavedTalents.create({ talentsString: '--50122001' }));

export const TalentPresets = {
	[Phase.Phase1]: [],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [P4RetTalents, TalentsLevel20],
};

export const DefaultTalents = TalentPresets[Phase.Phase4][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = RetributionPaladinOptions.create({
	aura: PaladinAura.RetributionAura,
	primarySeal: PaladinSeal.Righteousness,
});

export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	boglingRoot: false,
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	dragonBreathChili: true,
	fillerExplosive: Explosive.ExplosiveUnknown,
	firePowerBuff: FirePowerBuff.ElixirOfGreaterFirepower,
	food: Food.FoodBlessSunfruit,
	flask: Flask.FlaskOfSupremePower,
	//mainHandImbue: WeaponImbue.WildStrikes,
	//offHandImbue: WeaponImbue.MagnificentTrollshine,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.ROIDS,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfMight: TristateEffect.TristateEffectRegular,
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectRegular,
	fengusFerocity: false,
	moldarsMoxie: false,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	fireResistanceTotem: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	leaderOfThePack: true,
	moonkinAura: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfRecklessness: true,
	faerieFire: true,
	giftOfArthas: true,
	sunderArmor: true,
	judgementOfWisdom: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	improvedScorch: true,
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// A level 20 paladin soloing an instance boss: no raid buffs or consumes, its own
// Blessing of Kings (baseline at 20 under Forever, and 1 DPS ahead of Blessing of Might
// at 60 sec because the Intellect pays for more Consecrations), and the boss hitting the
// paladin so Retribution Aura gets to fire.
export const EncounterLevel20 = PresetUtils.makePresetEncounter(
	'Level 20',
	Encounter.create({
		duration: 60,
		durationVariation: 5,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		targets: [
			{
				id: 3654,
				name: 'Mutanus the Devourer',
				level: 22,
				mobType: MobType.MobTypeHumanoid,
				stats: new Stats().withStat(Stat.StatArmor, 922).withStat(Stat.StatHealth, 20000).asArray(),
				minBaseDamage: 40,
				damageSpread: 0.3333,
				swingSpeed: 2,
				parryHaste: true,
				tankIndex: 0,
			},
		],
	}),
	{
		tanks: [UnitReference.create({ type: UnitReference_Type.Player, index: 0 })],
		raidBuffs: RaidBuffs.create({}),
		debuffs: Debuffs.create({}),
		buffs: IndividualBuffs.create({ blessingOfKings: true }),
		consumes: Consumes.create({}),
	},
);
export const PresetBuildLevel20 = PresetUtils.makePresetBuild('Level 20', {
	gear: GearLevel20,
	talents: TalentsLevel20,
	rotation: APLLevel20,
	encounter: EncounterLevel20,
	race: Race.RaceUndead,
	level: 20,
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.RetributionAura,
		primarySeal: PaladinSeal.Command,
	}),
});

// A pack of two, three or four of the same mob, all hitting the paladin. The rotation
// and talents are the single target ones (Consecration is already on cooldown there, and
// putting it ahead of the seal and judgement lines loses a little), the gear is the AoE
// set.
const level20PackEncounter = (targets: number) =>
	PresetUtils.makePresetEncounter(
		`Level 20, ${targets} targets`,
		Encounter.create({
			...EncounterLevel20.encounter!,
			targets: Array.from({ length: targets }, () => EncounterLevel20.encounter!.targets[0]),
		}),
		{
			tanks: EncounterLevel20.tanks,
			raidBuffs: EncounterLevel20.raidBuffs,
			debuffs: EncounterLevel20.debuffs,
			buffs: EncounterLevel20.buffs,
			consumes: EncounterLevel20.consumes,
		},
	);
const level20PackBuild = (targets: number) =>
	PresetUtils.makePresetBuild(`Level 20, ${targets} targets`, {
		gear: GearLevel20AoE,
		talents: TalentsLevel20,
		rotation: APLLevel20,
		encounter: level20PackEncounter(targets),
		race: Race.RaceUndead,
		level: 20,
		options: RetributionPaladinOptions.create({
			aura: PaladinAura.RetributionAura,
			primarySeal: PaladinSeal.Command,
		}),
	});
export const PresetBuildLevel20x2 = level20PackBuild(2);
export const PresetBuildLevel20x3 = level20PackBuild(3);
export const PresetBuildLevel20x4 = level20PackBuild(4);
