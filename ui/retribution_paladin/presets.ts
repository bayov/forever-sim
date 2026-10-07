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
import Level30p5APLJSON from './apls/level30p5.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import Level30p5GearJSON from './gear_sets/level30p5.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear, { tooltip: 'No gear equipped.' });
// The gear search at level 30 with 26 talent points and Seal of Command, over everything
// an Undead paladin can equip, started from the old level 20 set. The professions are
// Blacksmithing and Enchanting, and the Scourge Invasion drops and the items no source is
// known for are left out. Corpsemaker (Razorfen Kraul) takes the two hander for 8.8 DPS. Blacksmithing makes
// the helm, shirt, leggings and belt (the Crusader's chain pieces and Justicar's Belt,
// bind on pickup) and the Hard Gold Boots. Bloodmage Mantle comes from the Scarlet
// Monastery Library and Unearthed Bands of Power from Uldaman.
//
// The first search took Scorn's Icy Choker (+0.7). Scorn only spawns in the Scourge
// Invasion, so we left his drops out and Mark of the Pack Leader keeps the neck. Manual
// Crowd Pummeler is left out too, see the shaman's Level 30 + 5 set.
//
// On 2026-10-05 the two-hander went from Lesser Strength (+15 Strength) to Impact (+6 weapon
// damage), because the +15 Strength and +15 Agility enchants are not in the beta. That is
// 130.7 against 131.7 at 60 sec. Revelation adds nothing here (128.8, the same as no enchant),
// because the rotation casts no spell that can trigger it.
export const GearLevel30p5 = PresetUtils.makePresetGear('Level 30 + 5', Level30p5GearJSON, {
	tooltip: 'Level 30 gear with Corpsemaker and Blacksmithing crafts.',
	group: 'Level 30',
});

export const GearPresets = {};

export const DefaultGear = GearBlank;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLBasicRet = PresetUtils.makePresetAPLRotation('Basic Ret', APLBasicRetJson, {
	tooltip: 'Basic level 60 Retribution rotation.',
	group: 'Level 60',
});
// The paladin_ret template with the knobs the level 30 search settled on (see
// tools/rotopt/paladin_presets.sh), with the level 30 ranks: Seal of the Crusader judged
// once for the fight, then Consecration whenever it is off cooldown, Seal of Command judged
// on every cooldown and Holy Strike on cooldown. Consecration goes first and spends the
// mana to the last point.
export const APLLevel30p5 = PresetUtils.makePresetAPLRotation('Level 30 + 5', Level30p5APLJSON, {
	tooltip: 'Level 30 Seal of Command rotation with Consecration and Holy Strike on cooldown.',
	group: 'Level 30',
});

export const APLPresets = {
	[Phase.Phase1]: [],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [APLBasicRet, APLLevel30p5],
	[Phase.Phase5]: [],
};

export const DefaultAPL = APLPresets[Phase.Phase4][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/forever/talent-calc and copy the numbers in the url.

// Forever trees (beta client build 1.60.1.70009). Level 60: Divine Strength and Improved
// Seals in Holy, Toughness and Precision in Protection, the Retribution tree down to Twist
// of Light. Not searched yet, the level 60 sets are blank. The 2026-09-24 build took
// Improved Holy Strike and Crusade away, and their 4 points went to Deflection 2 and Holy
// Conduit 2 so Twist of Light still has the 30 Retribution points it needs.
export const P4RetTalents = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '50003-513-25225331001330301' }), {
	tooltip: 'Level 60 Retribution talents down to Twist of Light. Not searched yet.',
	group: 'Level 60',
});
// Level 30 with 5 extra talent points (26 in all), Pursuit of Justice 2 pinned. Holy
// Shock is out of reach with it: Holy Shock needs 20 Holy points and Pursuit of Justice
// 10 Retribution points. We scored every Seal of Command build over the Retribution
// damage talents, with up to 8 Holy points in Divine Strength, Divine Intellect and
// Improved Seals (56865 builds). The best is Deflection 5, Improved Judgement 2, Holy
// Conduit 2, Conviction 4, Sanctified Judgement 3, Seal of Command, Sacred Arbiter,
// Two-Handed Weapon Specialization 3, Vengeance 2 and Champion of the Light 1, and it
// stays on top on the Level 30 + 5 gear.
//
// The 2026-10-01 development notes cut Champion of the Light to 20% of Intellect per
// point (was 33%). We rescored the best 150 builds and every build that moves its point
// to another talent, and this one is still first. The best build without Champion of
// the Light is 1.4 DPS behind at 60 sec.
export const TalentsLevel30p5 = PresetUtils.makePresetTalents('Level 30 + 5', SavedTalents.create({ talentsString: '--502240312013201' }), {
	tooltip: 'Level 30 with 5 extra talent points, all in Retribution for Seal of Command.',
	group: 'Level 30',
});

export const TalentPresets = {
	[Phase.Phase1]: [],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [P4RetTalents, TalentsLevel30p5],
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

// The paladin is a blacksmith and an enchanter. Gear another profession makes is only in
// the presets when it binds on equip, so it can be bought.
export const OtherDefaults = {
	profession1: Profession.Blacksmithing,
	profession2: Profession.Enchanting,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// A level 30 paladin soloing Interrogator Vishas (Scarlet Monastery Graveyard, level 32):
// no raid buffs or consumes, its own Blessing of Kings (2.4 DPS ahead of Blessing of Might
// at 60 sec), and the boss hitting the paladin so Retribution Aura gets to fire.
export const EncounterLevel30 = PresetUtils.makePresetEncounter(
	'Level 30',
	Encounter.create({
		duration: 60,
		durationVariation: 5,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		targets: [
			{
				id: 3983,
				name: 'Interrogator Vishas',
				level: 32,
				mobType: MobType.MobTypeHumanoid,
				stats: new Stats().withStat(Stat.StatArmor, 1063).withStat(Stat.StatHealth, 30000).asArray(),
				minBaseDamage: 52,
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
// Level 30 with 5 extra talent points. The optimizer puts Seal of Righteousness 2.6 DPS
// behind Seal of Command on Corpsemaker, even with its spell power counted twice. At
// 30 / 60 / 300 sec that is 130.5 / 131.8 / 105.3, and 187.9 on three targets.
export const PresetBuildLevel30p5 = PresetUtils.makePresetBuild('Level 30 + 5', {
	group: 'Level 30',
	tooltip: 'Level 30 Undead with 5 extra talent points and Seal of Command, soloing Interrogator Vishas.',
	gear: GearLevel30p5,
	talents: TalentsLevel30p5,
	rotation: APLLevel30p5,
	encounter: EncounterLevel30,
	race: Race.RaceUndead,
	level: 30,
	bonusTalentPoints: 5,
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.RetributionAura,
		primarySeal: PaladinSeal.Command,
	}),
});
