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
import { RestorationShaman_Options as RestorationShamanOptions } from '../core/proto/shaman.js';
import { SavedTalents } from '../core/proto/ui.js';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import RaidHealerAPLJson from './apls/raid_healer.apl.json';
import TankHealerAPLJson from './apls/tank_healer.apl.json';

// The Restoration shaman, the healing sim's first healer (docs/healing_sim.md, H2). Its numbers come
// from tools/healbench, which scores a healer by the damage it carries in a role: the tank healer on
// the boss's melee and abilities on the tanks, the raid healer on the boss's damage on the rest.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// The Restoration synthetic Phase 1 set (tools/database/forever_synthetic_items_shaman_restoration.go, ids
// 990201 to 990213), on the healer row of orchestration/gear_targets.md (standards 35): spell power and
// healing on top of it, Intellect, Spirit, crit and MP5, no hit. The Elemental set's trinkets, Enchant Bracer
// and Necklace - Healing Power, Spellblasting on the staff, and the Polished Driftwood Icon (8% of mana regen
// while casting). Against the Elemental set it heals 13 to 18% more per mana. Its carry is ahead or level on
// 8 of the 14 boss and role rows (Golemagg's raid side 2.083 against 1.990, Vaelastrasz's 1.199 against
// 1.140) and up to 0.025 behind on the short spiky tank fights (Broodlord 0.618 against 0.643), with 63 crit
// rating against 100; Onyxia's tank side at 300 sec is 0.954 against 0.830.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Restoration', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS for a healer under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

export const GearPresets = [GearLevel60Phase1Synthetic];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// Both rotations put Water Shield up at -3 sec and Mana Spring Totem at -1.5 sec, cast Lesser Healing
// Wave at whoever in the assignment drops under 35% health, put Mana Tide Totem down at 35% mana and
// Mana Spring Totem back after it, and bind at most 4 ranks of a heal (the user's ruling), each cast
// when what's missing covers its Expected Healing.
//
// The tank healer: Riptide on the tank whenever it's off, Healing Wave 10, 8 and 6 at the tank, and
// Chain Heal 3, 2 and 1 at the most hurt when the tank is topped. Water Shield goes back up as soon as
// its globes are spent: a tank healer is seldom hit, so its globes go on its heals' crits. Against the
// bench's earlier rotation (the same heals, no mana tools; 1,000 iterations, one healer on the tank)
// it carries the same in Onyxia's 35 sec (1.031), and 0.959 against 0.836 at 180 sec and 0.830
// against 0.677 at 300, where the old one runs dry (178 and 185 sec).
export const APLTankHealer = PresetUtils.makePresetAPLRotation('Tank healer', TankHealerAPLJson, {
	tooltip: 'Riptide and bound ranks of Healing Wave at the tank, Chain Heal at the most hurt, with Water Shield, Mana Spring and Mana Tide.',
	group: 'Healing sim',
});
// The raid healer: Riptide at the most hurt when it fits, then Chain Heal 3, 2 and 1 at the most
// hurt, then Healing Wave 10 and 8 for one big hole. Raid damage spends Water Shield's globes fast,
// and a recast every few hits costs the healing its global cooldowns, so it goes back up only below
// 60% mana: 0.899 against 0.874 recasting at once on the bench's raid profile, 1.140 against 1.056 on
// Vaelastrasz, 0.844 against 0.796 on Broodlord, and 0.865 against 0.843 at 300 sec. Against the old
// rotation (Chain Heal only): 0.899 against 0.850 on the raid profile, 0.695 against 0.672 on Onyxia,
// 1.140 against 1.066 on Vaelastrasz, 0.990 against 0.928 on Chromaggus (2 healers, 7 Frost mages).
export const APLRaidHealer = PresetUtils.makePresetAPLRotation('Raid healer', RaidHealerAPLJson, {
	tooltip: 'Chain Heal and Riptide at the most hurt, Healing Wave for a big hole, with Water Shield, Mana Spring and Mana Tide.',
	group: 'Healing sim',
});

export const APLPresets = [APLTankHealer, APLRaidHealer];

export const DefaultAPL = APLTankHealer;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// 31 Restoration points to Riptide, with Water Shield, Mana Tide Totem, Healing Way and Nature's
// Swiftness, and Ancestral Knowledge, Mental Dexterity and Shamanistic Focus in Enhancement. Not
// searched yet: tools/rotopt's talent search scores DPS.
export const TalentsRestoration = PresetUtils.makePresetTalents('Restoration', SavedTalents.create({ talentsString: '-005102-5530500315513151' }), {
	tooltip: 'Riptide, Water Shield, Mana Tide Totem, Healing Way and Nature\'s Swiftness.',
	group: 'Healing sim',
});

export const TalentPresets = [TalentsRestoration];

export const DefaultTalents = TalentsRestoration;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// No shaman imbue: the main hand carries a Brilliant Mana Oil.
export const DefaultOptions = RestorationShamanOptions.create({});

// The Elemental shaman's level 60 consumables, with a Brilliant Mana Oil and the Major Mana Potion.
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

// The shared level 60 raid buffs and debuffs, as the Elemental shaman's.
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
	channelClipDelay: 50,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceSkyborneWindshaper,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

export const PresetBuildTankHealer = PresetUtils.makePresetBuild('Tank healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Windshaper Skyborne on the Restoration synthetic set with the tank healer rotation.',
	race: Race.RaceSkyborneWindshaper,
	level: 60,
	gear: DefaultGear,
	talents: TalentsRestoration,
	rotation: APLTankHealer,
});

export const PresetBuildRaidHealer = PresetUtils.makePresetBuild('Raid healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Windshaper Skyborne on the Restoration synthetic set with the raid healer rotation.',
	race: Race.RaceSkyborneWindshaper,
	level: 60,
	gear: DefaultGear,
	talents: TalentsRestoration,
	rotation: APLRaidHealer,
});
