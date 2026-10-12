import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	Flask,
	Food,
	IndividualBuffs,
	ManaRegenElixir,
	MobType,
	PartyBuffs,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	SpellPowerBuff,
	Stat,
	TristateEffect,
	UnitReference,
	WeaponImbue,
} from '../core/proto/common.js';
import { BalanceDruid_Options as BalanceDruidOptions } from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import Balance from './apls/balance.apl.json';
import Level60APLJSON from './apls/level60.apl.json';
import Level60WrathAPLJSON from './apls/level60_wrath.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import P0BISGear from './gear_sets/p0.bis.gear.json';
import P1BISGear from './gear_sets/p1.bis.gear.json';
import P2BISGear from './gear_sets/p2.bis.gear.json';
import P3BISGear from './gear_sets/p3.bis.gear.json';
import P4BISGear from './gear_sets/p4.bis.gear.json';
import P5BISGear from './gear_sets/p5.bis.gear.json';
// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearP0BIS = PresetUtils.makePresetGear('Pre-BiS', P0BISGear);
export const GearP1BIS = PresetUtils.makePresetGear('P1 BiS', P1BISGear);
export const GearP2BIS = PresetUtils.makePresetGear('P2 BiS', P2BISGear);
export const GearP3BIS = PresetUtils.makePresetGear('P3 BiS', P3BISGear);
export const GearP4BIS = PresetUtils.makePresetGear('P4 BiS', P4BISGear);
export const GearP5BIS = PresetUtils.makePresetGear('P5 BiS', P5BISGear);

// The level 60 Phase 1 set for Forever, made of synthetic items that are not in the game
// (tools/database/forever_synthetic_items_druid.go, which has the budget math). Each piece
// takes an Era Phase 1 BiS item's place at its item level: 474 spell power, 11% of the budget
// on Spirit, 3% hit and 4% crit. The head, shoulders, chest, hands, legs and feet carry the
// bonuses of Grovekeeper Eclipse, Forever's crafted level 60 Balance set (1% hit at 2 pieces
// and 3 sec more Insect Swarm at 5). The trinkets are real: Talisman of Ephemeral Power and
// Briarwood Reed. Enchants are the Phase 1 ones a caster takes: Lesser Arcanum of Voracity
// (Intellect) on the head and legs, Forever's +6 spell power on the neck, Greater Stats, +7
// Intellect, +5 Spirit on the boots and Spell Power on the staff.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic],
	[Phase.Phase5]: [GearP0BIS, GearP1BIS, GearP2BIS, GearP3BIS, GearP4BIS, GearP5BIS],
};

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The tools/rotopt druid_balance template's best (druid_presets.sh) on the synthetic set and
// the Level 60 talents: Insect Swarm and Moonfire whenever they are down, a free Starfire on
// Clearcasting (Wrath does not use it up), and Wrath until Eclipse has a charge, then Starfire
// (so Wrath and Starfire take turns, each Starfire 0.5 sec shorter). The knob search kept that
// at 120, 180 and 300 sec. Dropping Moonfire costs 53 DPS at 180 sec, Insect Swarm 58, and
// casting only Starfire 57. Innervate is left to its own trigger (at 500 mana), 3.9 ahead of
// casting it at 40% mana. Stopping the dots or switching to Wrath alone below a mana floor
// only lost, the set does not run dry in a raid fight.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Level 60', Level60APLJSON, {
	tooltip: "The optimizer's best level 60 rotation: both dots up, then Wrath and Starfire in turn for Eclipse.",
	group: 'Level 60',
});
// export const APLP1Balance = PresetUtils.makePresetAPLRotation('Balance', P1APL);
export const DefaultBalance = PresetUtils.makePresetAPLRotation('Default Balance', Balance);
// Wrath as the only filler (1.5 sec with Improved Wrath, half the mana), for a moonkin that
// would rather not weave: 661.4 / 656.6 / 661.0 at 120 / 180 / 300 sec, 8% behind the Level 60
// rotation. A talent search with this rotation found the same talents (Eclipse and Improved
// Starfire still pay for the free Clearcasting Starfires). Starfire alone is 659.5 at 180 sec.
export const APLLevel60Wrath = PresetUtils.makePresetAPLRotation('Level 60 Wrath', Level60WrathAPLJSON, {
	tooltip: 'Both dots up, a free Starfire on Clearcasting, and Wrath otherwise.',
	group: 'Level 60',
});
export const APLPresets = {
	[Phase.Phase1]: [APLLevel60, APLLevel60Wrath, DefaultBalance],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search's best (tools/rotopt -talent-search) on the synthetic set with the Level 60
// rotation at 180 sec: 41 Balance (everything that does damage, Genesis 5 and Moonglow 3) and
// Naturalist 5 (1% damage a point) from Restoration. Searches started from that build with
// Reflection and Subtlety and from 41 / 10 / 0 with Heart of the Wild, and the last moves traded
// Moonglow, Genesis and Heart of the Wild points within the noise. The other archetypes are well
// behind at 180 sec: 31 / 0 / 20 (Reflection and Tranquil Spirit, no Moonfury)
// 642.8, 41 / 10 / 0 685.0, against 715.2 (Tauren).
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '5532220115501351--505' }), {
	tooltip: 'The talent search best on the synthetic Phase 1 set: 41 Balance, Naturalist from Restoration.',
	group: 'Level 60',
});

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = BalanceDruidOptions.create({
	innervateTarget: UnitReference.create(),
});

// The level 60 consumables. On 2026-10-12 we tried every consumable the sim applies in each
// slot on the Level 60 build at 180 sec (before the talent search, 748 DPS). Taking each one
// out costs: Flask of Supreme Power 93 (Distilled Wisdom gives nothing, the moonkin does not
// run dry in a raid fight), Brilliant Wizard Oil 30 over Brilliant Mana Oil, Greater Arcane
// Elixir 22, Nightfin Soup 14 (Runn Tum Tuber Surprise and Sagefish Delight are 9 to 12
// behind), Goblin Sapper Charge 3.9, Major Mana Potion 2.2 and Demonic Rune 1.4. Mageblood
// Potion and Elixir of Wisdom are within the noise and stay for longer fights. Goblin Land
// Mine costs 4.9 (a global cooldown), the Elixir of Nature Power and the hit foods do nothing
// in the sim.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodNightfinSoup,
	mainHandImbue: WeaponImbue.BrilliantWizardOil,
	manaRegenElixir: ManaRegenElixir.MagebloodPotion,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
});

// The raid the Level 60 numbers were measured under (tools/rotopt/mksettings -spec balance):
// the shared level 60 raid with the moonkin in a caster group, so a shaman's Mana Spring
// Totem, Blessing of Wisdom and Kings but not Might, and its own Moonkin aura. Under Forever
// Curse of the Elements covers every school and Curse of Shadow is gone, so only the first is
// on. The melee debuffs do nothing for it.
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	sanctityAura: true,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
});

export const DefaultPartyBuffs = PartyBuffs.create({});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	curseOfRecklessness: true,
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

// An engineer for the Goblin Sapper Charge (+3.9 at 180 sec) and an alchemist, like the other
// level 60 presets. The moonkin casts from range, so it does not melee between casts.
export const OtherDefaults = {
	distanceFromTarget: 30,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceNightElf,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the shared Level 60 boss (level 63, 3731 armor, Humanoid), a 3 minute
// fight, and the buffs, debuffs and consumes above.
export const EncounterLevel60 = PresetUtils.makePresetEncounter(
	'Level 60',
	Encounter.create({
		duration: 180,
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
// The Level 60 build, 721.8 / 715.8 / 717.5 at 120 / 180 / 300 sec. Every race a druid can be
// under Forever on it with the same seed, 40,000 iterations (standard error 0.1 to 0.2):
//
//	                     120 sec  180 sec  300 sec
//	Night Elf             721.8    715.8    717.5   Elune's Light, best overall and Alliance
//	Tauren                719.3    714.8    717.2   1% hit, best Horde
//	Windshaper Skyborne   717.0    712.5    714.7   1% cast speed
//	High Order Skyborne   716.9    712.3    714.5
//
// No racial goes by weapon type, so the staff changes nothing here.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Level 60', {
	group: 'Level 60',
	tooltip: 'Level 60 Night Elf on the synthetic Phase 1 gear with the Level 60 rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceNightElf,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
export const PresetBuildLevel60Wrath = PresetUtils.makePresetBuild('Level 60 Wrath', {
	group: 'Level 60',
	tooltip: 'The Level 60 build with Wrath as the only filler.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60Wrath,
	encounter: EncounterLevel60,
	race: Race.RaceNightElf,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
