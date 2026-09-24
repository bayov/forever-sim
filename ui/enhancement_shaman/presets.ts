import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	FirePowerBuff,
	Flask,
	Food,
	IndividualBuffs,
	ManaRegenElixir,
	MobType,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	SpellPowerBuff,
	Stat,
	StrengthBuff,
	TristateEffect,
	UnitReference,
	UnitReference_Type,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { EnhancementShaman_Options as EnhancementShamanOptions, ShamanSyncType } from '../core/proto/shaman.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import DefaultAPLJSON from './apls/default.apl.json';
import GraceOfAirAPLJSON from './apls/grace_of_air.apl.json';
import Level20APLJSON from './apls/level20.apl.json';
import OptimizedAPLJSON from './apls/optimized.apl.json';
import WindfuryAPLJSON from './apls/windfury.apl.json';
import Level20GearJSON from './gear_sets/level20.gear.json';
import Phase1GearJSON from './gear_sets/phase_1.gear.json';
import Phase2GearJSON from './gear_sets/phase_2.gear.json';
import Phase3GearJSON from './gear_sets/phase_3.gear.json';
import Phase5GearJSON from './gear_sets/phase_5.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearPhase1 = PresetUtils.makePresetGear('Phase 1', Phase1GearJSON);
export const GearPhase2 = PresetUtils.makePresetGear('Phase 2', Phase2GearJSON);
export const GearPhase3 = PresetUtils.makePresetGear('Phase 3', Phase3GearJSON);
export const GearPhase5 = PresetUtils.makePresetGear('Phase 5', Phase5GearJSON);
// The level 20 set comes from the tools/rotopt gear search over every item an Orc shaman
// can equip, Forever quest rewards included, without the Scourge Invasion drops (The Axe
// of Severing, Abomination Skin Leggings) the first search picked. It assumes Engineering
// and Leatherworking: Spellpower Goggles Xtreme is the only head slot worth wearing at
// 20, and the crafted leather sets take most of the armor slots. With the 11 points a
// level 20 has, the search wants Strength and attack power.
//
// Two items from outside the dungeons are worth 2.5 DPS together. Philanthropist's Ring
// (6 Intellect, 10 spell power) is the Greater Friend of the Library reward for 20 books,
// four of them in level 30 zones, and every class can take it. Mark of the Pack Leader
// (5 Strength, 5 Intellect) drops off Humar the Pridelord, a rare in the Barrens, which
// ForeverChanges has from player reports it has not checked.
//
// Fletcher's Gloves are 14 crit rating, which is 4.3% crit at level 20 if rating scales
// with level the way we assume (see core.RatingPerPercent). Forever pays crit from gear
// on spells too, so they beat Jutebraid Gloves by 1 DPS. The search still wants the
// Totemic Leather Belt over the Screecher Belt for 0.1 DPS at 60 sec, but it loses 0.6
// at 300 because the shaman runs out of mana, so we keep the Screecher Belt.
//
// The enchants are the best a level 20 can get, Enchanting 225 at most, and are worth 7
// DPS. The shaman takes spell power wherever there is a choice (the Spell Power necklace
// enchant, Mystic armor kits, Lesser Healing Power on the bracers for its 6 spell damage)
// because the shocks, Searing Totem and Lightning Shield all scale with it. The two
// hander takes +15 Strength and the gloves +7 Strength.
//
// The two hander is the one slot where the race decides it. Forsaken Greataxe (The Wrath
// of Rath'mael) and Hammerbone (Leaders of the Fang) sit within a few tenths of a DPS of
// each other, and the Orc's Axe Specialization is what separates them. At 300000
// iterations the axe does 75.0 against 74.9 at 60 sec and 61.1 against 60.9 at 300, and
// the mace is 0.3 ahead at 30. A Troll or a Tauren has no racial for the axe and wants
// Hammerbone. The gear search cannot call this one. It only takes a swap that gains more
// than twice the noise of its runs, about 0.09 DPS at 20000 iterations, so whichever
// weapon it starts with is the one it keeps.
export const GearLevel20 = PresetUtils.makePresetGear('Level 20', Level20GearJSON);

export const GearPresets = {
	[Phase.Phase1]: [GearPhase1, GearLevel20],
	[Phase.Phase2]: [GearPhase2],
	[Phase.Phase3]: [GearPhase3],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [GearPhase5],
	[Phase.Phase6]: [],
};

export const DefaultGear = GearPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// Three ways to run the air totem. Twisting keeps Windfury Totem up for the melee
// group and refreshes Grace of Air in between, the way a raid shaman plays, but the
// sim gives the shaman nothing for its own Windfury Totem (a shaman's Windfury Weapon
// does not stack with it), so twisting is pure cost here. The other two hold one totem.
export const APLGraceOfAir = PresetUtils.makePresetAPLRotation('Grace of Air', GraceOfAirAPLJSON);
export const APLWindfury = PresetUtils.makePresetAPLRotation('Windfury Totem', WindfuryAPLJSON);
export const APLDefault = PresetUtils.makePresetAPLRotation('Twist WF + GoA', DefaultAPLJSON);
// Generated by tools/rotopt/shaman_presets.sh (the shaman_enh template). Optimized is the
// knob search's best at 60: Stormstrike on its 8 sec cooldown, Flame Shock when its DoT
// is down and Earth Shock otherwise (Earth Shock spends the Stormstrike mark for +20%),
// Magma Totem (+6 over Searing), Lightning Bolt only at 5 Maelstrom Weapon stacks, Rage
// of the Farseer on cooldown with Blood Fury and Berserking waiting for it when that
// costs them no use.
export const APLOptimized = PresetUtils.makePresetAPLRotation('Optimized', OptimizedAPLJSON);
export const APLLevel20 = PresetUtils.makePresetAPLRotation('Level 20', Level20APLJSON);

export const APLPresets = {
	[Phase.Phase1]: [APLOptimized, APLGraceOfAir, APLWindfury, APLDefault, APLLevel20],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search's best on the Phase 2 gear (tools/rotopt -talent-search) with Spirit
// Weapons pinned (-require spiritWeapons): the sim scores no threat, but the 30% cut is a
// must in a raid, and it costs under 1 DPS here because Reverberation 3 and Call of
// Flame 3 (Magma Totem is the fire totem now) tie Reverberation 5 with Call of Flame 2
// once Stormstrike is on an 8 sec cooldown. Elemental Fury is worth 25 DPS over
// Convection, Shamanistic Focus 43, Rage of the Farseer 24, Mental Quickness 11.
// Improved Stormstrike is the mana talent: with no Mana Spring the shaman is dry at five
// minutes, and its two points (taken from Ancestral Knowledge) cost 2 DPS in a 4 minute
// fight and gain 6 at 8 minutes and 12 at 10. The dodge/parry reset half never fires on
// a boss the shaman is behind.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '05033305-053030031005112251' }));
// The same search with Improved Ghost Wolf 2 pinned as well (-require
// spiritWeapons,improvedGhostWolf). The two points come out of Ancestral Knowledge (5 to
// 3) and cost 3 DPS at any fight length. Every other place to take them from costs more:
// Improved Stormstrike is worth 40 once the pool runs low.
export const TalentsLevel60ImpGW = PresetUtils.makePresetTalents('Level 60 + Imp GW', SavedTalents.create({ talentsString: '05033305-051032031005112251' }));
export const TalentsLevel20 = PresetUtils.makePresetTalents('Level 20', SavedTalents.create({ talentsString: '-050030201' }));

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60ImpGW, TalentsLevel20],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = EnhancementShamanOptions.create({
	syncType: ShamanSyncType.Auto,
});

export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MajorManaPotion,
	defaultConjured: Conjured.ConjuredDemonicRune,
	dragonBreathChili: true,
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodBlessSunfruit,
	mainHandImbue: WeaponImbue.WindfuryWeapon,
	manaRegenElixir: ManaRegenElixir.MagebloodPotion,
	offHandImbue: WeaponImbue.WindfuryWeapon,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// The raid the level 60 numbers were measured under (the rogue search's export minus
// the totems): both paladin blessings, Judgement of Wisdom and every debuff, but no
// second shaman in the group, so the player's own Strength of Earth and Grace of Air
// are what the rotation casts and there is no Mana Spring. Judgement of Wisdom alone is
// 60 DPS here because without it the shaman is out of mana two minutes in.
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
	sanctityAura: true,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
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
	curseOfShadow: true,
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	demoralizingRoar: TristateEffect.TristateEffectImproved,
	demoralizingShout: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectImproved,
	faerieFire: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	scorpidSting: true,
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

// Engineering for the Goblin Sapper in the consumes (worth about 1 DPS).
export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
	race: Race.RaceOrc,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor), a 4 to 5 minute
// fight, and the default buffs, debuffs and consumes. It is here so switching back from
// the Level 20 build restores everything that build changed.
export const EncounterLevel60 = PresetUtils.makePresetEncounter(
	'Level 60',
	Encounter.create({
		duration: 280,
		durationVariation: 20,
		executeProportion20: 0.2,
		executeProportion25: 0.25,
		executeProportion35: 0.35,
		targets: [
			{
				id: 213336,
				name: 'Level 60',
				level: 63,
				mobType: MobType.MobTypeHumanoid,
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
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Level 60', {
	gear: GearPhase2,
	talents: TalentsLevel60,
	rotation: APLOptimized,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
});

// A level 20 shaman soloing an instance boss: no raid buffs or consumes, Rockbiter on
// the weapon, and the boss hitting the shaman so Lightning Shield gets to fire. Blessing
// of Kings is on because a group at that level almost always has a paladin in it, and the
// gear search ran with it.
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
		consumes: Consumes.create({
			mainHandImbue: WeaponImbue.RockbiterWeapon,
		}),
	},
);
export const PresetBuildLevel20 = PresetUtils.makePresetBuild('Level 20', {
	gear: GearLevel20,
	talents: TalentsLevel20,
	rotation: APLLevel20,
	encounter: EncounterLevel20,
	race: Race.RaceOrc,
	level: 20,
});
