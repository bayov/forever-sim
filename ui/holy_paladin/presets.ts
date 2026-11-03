import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Flask,
	Food,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SpellPowerBuff,
	SpiritElixir,
	SunderRamp,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { PaladinAura, PaladinOptions as HolyPaladinOptions } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import RaidHealerAPLJson from './apls/raid_healer.apl.json';
import TankHealerAPLJson from './apls/tank_healer.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';

// The Holy paladin, a healer of the healing sim (docs/healing_sim.md, H3; docs/healing_paladin.md has the
// audit, the tuning and the numbers). Its numbers come from tools/healbench, which scores a healer by the
// damage it carries in a role: the tank healer on the boss's melee and abilities on the tanks, the raid
// healer on the boss's damage on the rest. The presets come out of
// orchestration/tools/healers/paladin_holy_requests.py, with tools/healbench/req/paladin_holy*.json.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// The Holy synthetic Phase 1 set (tools/database/forever_synthetic_items_paladin_holy.go, ids 990301 to
// 990314), on the healer row of orchestration/gear_targets.md (standards 35): plate, healing and spell
// power, Intellect, Stamina, Spirit, MP5 and a little crit, no hit, a caster main hand and held off-hand,
// the Retribution set's spell power trinkets, and Libram of Economy (5% off Holy Light). The head and chest
// carry Justice Armor's 2-piece (+10 Spirit). The Restoration shaman's enchants, with Mighty Intellect on
// the main hand and Off-Hand - Wisdom.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Holy', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS for a Holy paladin under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);

export const GearPresets = [GearLevel60Phase1Synthetic, GearBlank];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// Both rotations put the trinkets, Major Mana Potion and Demonic Rune out through Autocast Other
// Cooldowns, cast Holy Shock at whoever in the assignment drops under 50% health and Flash of Light 6
// under 35%, Lay on Hands under 10%, and bind at most 4 ranks of a heal (the user's ruling), each cast when
// what's missing covers its Expected Healing.
//
// The tank healer: Oath of Light on the tank (at -1.5 sec and whenever it's off: 400 more on its Holy
// Lights, 115 on its Flashes), Divine Favor before a Holy Light 9 the tank misses, Holy Light 9, 8 and 7
// at the tank, and Flash of Light 6 at the most hurt in between.
export const APLTankHealer = PresetUtils.makePresetAPLRotation('Tank healer', TankHealerAPLJson, {
	tooltip: 'Oath of Light on the tank, bound ranks of Holy Light at it with Divine Favor, Flash of Light and Holy Shock for the rest.',
	group: 'Healing sim',
});
// The raid healer: the same heals at the most hurt, with no Oath.
export const APLRaidHealer = PresetUtils.makePresetAPLRotation('Raid healer', RaidHealerAPLJson, {
	tooltip: 'Bound ranks of Holy Light at the most hurt with Divine Favor, Flash of Light and Holy Shock for the rest.',
	group: 'Healing sim',
});

export const APLPresets = [APLTankHealer, APLRaidHealer];

export const DefaultAPL = APLTankHealer;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// 31 Holy points to Light's Vigil: Spiritual Focus 5, Healing Light 3, Reverence 2, Divine Favor,
// Illumination 5, Infusion of Light 2, Oath of Light, Holy Shock, Boundless Light 2, Holy Power 5; Divine
// Intellect 5 in Protection; Benediction 5, Improved Judgement 2, Conviction 3, Sanctified Judgement 3 and
// Pursuit of Justice 2 in Retribution. Compared on tools/healbench (docs/healing_paladin.md, Talents).
export const TalentsHoly = PresetUtils.makePresetTalents('Holy', SavedTalents.create({ talentsString: '53320015002101251-5-052330002' }), {
	tooltip: 'Deep Holy to Light\'s Vigil, with Oath of Light, Holy Power, Illumination and Conviction.',
	group: 'Healing sim',
});

// The raid sim's placeholder Holy paladin (sim/compsearch/presets.go), which gives its buffs.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '50320215022101251-510221305',
	}),
};

export const TalentPresets = [TalentsHoly];

export const DefaultTalents = TalentsHoly;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = HolyPaladinOptions.create({
	aura: PaladinAura.NoPaladinAura,
});

// The raid sim's placeholder keeps its options and consumables as they were.
export const RaidSimOptions = HolyPaladinOptions.create({
	aura: PaladinAura.DevotionAura,
});
export const RaidSimConsumes = Consumes.create({
	flask: Flask.FlaskUnknown,
	food: Food.FoodUnknown,
});

// The Restoration shaman's consumables: a Brilliant Mana Oil and the Major Mana Potion.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.BrilliantManaOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The shared level 60 raid buffs and debuffs, as the Restoration shaman's.
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
	sunderRamp: SunderRamp.create({ tank: true, pullSunders: 1 }),
	thunderClap: TristateEffect.TristateEffectImproved,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceHuman,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

export const PresetBuildTankHealer = PresetUtils.makePresetBuild('Tank healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Human on the Holy synthetic set with the tank healer rotation.',
	race: Race.RaceHuman,
	level: 60,
	gear: DefaultGear,
	talents: TalentsHoly,
	rotation: APLTankHealer,
});

export const PresetBuildRaidHealer = PresetUtils.makePresetBuild('Raid healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Human on the Holy synthetic set with the raid healer rotation.',
	race: Race.RaceHuman,
	level: 60,
	gear: DefaultGear,
	talents: TalentsHoly,
	rotation: APLRaidHealer,
});
