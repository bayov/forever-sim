import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Flask,
	Food,
	HealingPowerBuff,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	Potions,
	Profession,
	Race,
	RaidBuffs,
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
// The tank healer: Nature's Swiftness when the tank is under 50% health, Riptide on the tank whenever
// it's off, Lesser Healing Wave 6 at the tank while mana is at least 50%, then Healing Wave 10,
// 8 and 6 at the tank, and Chain Heal 3, 2 and 1 at the most hurt when the tank is topped. Water
// Shield goes back up as soon as its globes are spent: a tank healer is seldom hit, so its globes go
// on its heals' crits. Lesser Healing Wave heals more a second than Healing Wave for a little more a
// mana, and the short fights end with mana to spare: on round 70753 (the notes' Riptide), 1.072
// against 0.959 mean carry over the bench's nine rows (1,000 iterations, one healer on the tank),
// 1.798 against 1.503 on Magmadar, 1.141 against 1.045 on Onyxia, 0.969 against 0.974 at 300 sec. The
// mana floor at 50 to 70% and Nature's Swiftness at 35 to 50% are within 0.01 (docs/healing_shaman.md).
export const APLTankHealer = PresetUtils.makePresetAPLRotation('Tank healer', TankHealerAPLJson, {
	tooltip: 'Riptide, Lesser Healing Wave above half mana and bound ranks of Healing Wave at the tank, Nature\'s Swiftness at half health, Chain Heal at the most hurt, with Water Shield, Mana Spring and Mana Tide.',
	group: 'Healing sim',
});
// The raid healer: Riptide at the most hurt whenever they have none, then Chain Heal 3, 2 and 1 at the
// most hurt, then Healing Wave 10 and 8 for one big hole. Raid damage spends Water Shield's globes
// fast, and a recast every few hits costs the healing its global cooldowns, so it goes back up only
// below 60% mana. On round 70753 (the notes' Riptide: no instant heal) it carries 1.123 mean over the
// bench's eight rows, against 1.183 on master before (1.196 with the client's three Riptide ranks);
// Riptide kept on the tank (-0.007), Lesser Healing Wave at the most hurt above 70% mana (-0.013) and
// other Water Shield or Mana Tide floors (no change: the fights end first) don't do better. Nature's Swiftness
// waits for someone in the assignment under 65% health and goes on an instant Healing Wave 10 at them (it was
// autocast before 2026-11-12): 1.392 -> 1.412 mean carry over the seven boss files (docs/healing_shaman.md).
export const APLRaidHealer = PresetUtils.makePresetAPLRotation('Raid healer', RaidHealerAPLJson, {
	tooltip: 'Chain Heal and Riptide at the most hurt, Healing Wave for a big hole, Nature\'s Swiftness on an instant Healing Wave under 65% health, with Water Shield, Mana Spring and Mana Tide.',
	group: 'Healing sim',
});

export const APLPresets = [APLTankHealer, APLRaidHealer];

export const DefaultAPL = APLTankHealer;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// 31 Restoration points to Riptide, with Water Shield, Mana Tide Totem, Healing Way and Nature's
// Swiftness, and Ancestral Knowledge 5 and Thundering Strikes 3 (1% crit a point on the heals) in
// Enhancement: on round 70753's tree (Improved Ghost Wolf moved to row 4) nothing else there helps a
// healer. The raid healer's: raid damage hits it, so Healing Focus keeps its heals from being pushed
// back (Ancestral Healing in its place carries 0.03 less on average, 0.075 on Vaelastrasz;
// docs/healing_shaman.md). Not searched beyond that: tools/rotopt's talent search scores DPS.
export const TalentsRestoration = PresetUtils.makePresetTalents('Restoration', SavedTalents.create({ talentsString: '-035-5530500315513151' }), {
	tooltip: 'Riptide, Water Shield, Mana Tide Totem, Healing Way and Nature\'s Swiftness.',
	group: 'Healing sim',
});

// The tank healer's: nothing hits a healer in the tank role, so Healing Focus's 3 points go to
// Ancestral Healing, which gives the tank 25% more armor from items for 15 sec after a heal crit
// (docs/healing_shaman.md). It carries 0.04 more on average (0.918 against 0.959 over the bench's
// nine rows).
export const TalentsRestorationTank = PresetUtils.makePresetTalents(
	'Restoration (tank healer)',
	SavedTalents.create({ talentsString: '-035-5530503015513151' }),
	{
		tooltip: 'Riptide, Water Shield, Mana Tide Totem, Healing Way, Nature\'s Swiftness and Ancestral Healing.',
		group: 'Healing sim',
	},
);

export const TalentPresets = [TalentsRestorationTank, TalentsRestoration];

export const DefaultTalents = TalentsRestorationTank;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// No shaman imbue: the main hand carries a Brilliant Mana Oil.
export const DefaultOptions = RestorationShamanOptions.create({});

// Each healer preset takes the consumables that score best on its own bench objective (healbench damage carried,
// mean of the 7 boss files in its role; docs/healing_sim.md, Consumables): Greater Cleric's Elixir (healing; the
// Arcane Elixir's spell damage doesn't heal), Flask of Supreme Power, the phase 1 oils only (Brilliant Wizard and
// Brilliant Mana Oil are Zandalar vendor formulas).
// The shaman takes Runn Tum Tuber Surprise (15 Intellect) and the Wizard Oil in both roles.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodRunnTumTuberSurprise,
	healingPowerBuff: HealingPowerBuff.GreaterClericsElixir,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.WizardOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
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
	talents: TalentsRestorationTank,
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
