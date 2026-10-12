import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	Food,
	IndividualBuffs,
	ManaRegenElixir,
	MobType,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	Stat,
	StrengthBuff,
	TotemWeaponBuff,
	TristateEffect,
	ZanzaBuff,
} from '../core/proto/common.js';
import {
	Hunter_Options as HunterOptions,
	Hunter_Options_Ammo as Ammo,
	Hunter_Options_PetAttackSpeed as PetAttackSpeed,
	Hunter_Options_PetType as PetType,
	Hunter_Options_QuiverBonus,
} from '../core/proto/hunter.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import BeastMasteryAPL from './apls/bm.apl.json';
import MarksmanshipAPL from './apls/mm.apl.json';
import SurvivalAPL from './apls/sv.apl.json';
import Level60RangedGear from './gear_sets/level60_phase1_synthetic_ranged.gear.json';
import Level60SurvivalGear from './gear_sets/level60_phase1_synthetic_survival.gear.json';
import P0BISGear from './gear_sets/p0.bis.gear.json';
import P1BISGear from './gear_sets/p1.bis.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearP0BIS = PresetUtils.makePresetGear('Pre-BiS', P0BISGear);
export const GearP1BIS = PresetUtils.makePresetGear('P1 BiS', P1BISGear);

// A guess at Phase 1 BiS under Forever, made of the synthetic items in
// tools/database/forever_synthetic_items_hunter.go (ids from 993000, not in the game).
//
// Forever's level 60 gear in the client is not tuned yet, and Molten Core and Onyxia have no
// Forever data. Each piece takes the place of an Era Phase 1 hunter BiS piece (the
// Giantstalker set, Onyxia Tooth Pendant, Wristguards of True Flight, the Devilsaur hands and
// legs, Quick Strike Ring, Barbarous Blade, Rhok'delar) with the full epic budget of its slot
// at item level 66 or the Era item's level. The file has the budget math. Beast Mastery and
// Marksmanship share the Ranged set: they shoot on the same stats, and the pet takes 10% of
// the hunter's ranged attack power and all of its crit. About 10% of the points are Spirit,
// 9% hit (the ranged hit cap against a level 63 boss) and 5% crit.
export const GearLevel60Ranged = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic Ranged', Level60RangedGear, {
	tooltip: 'A guess at Phase 1 BiS for a ranged hunter under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
// The Survival set: melee attack power and Strength in place of ranged attack power, a pair of
// one-handed swords (Brutality Blade's damage and speed) for Predator's Edge's off hand bonus,
// and 8% hit, since Survival takes Surefooted's 3%.
export const GearLevel60Survival = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic Survival', Level60SurvivalGear, {
	tooltip: 'A guess at Phase 1 BiS for a melee Survival hunter under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Ranged, GearLevel60Survival, GearP0BIS, GearP1BIS],
};

export const DefaultGear = GearLevel60Ranged;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The rotations come from the tools/rotopt templates (hunter_ranged, hunter_sv), see
// tools/rotopt/hunter_presets.sh. Auto Shot keeps firing through casts under Forever, so the
// ranged rotations only weigh the shots against each other: Serpent Sting when it is down,
// Aimed Shot on the cooldown it shares with Multi-Shot, Arcane Shot on its own (shared with
// Summon Hawk), Rapid Fire, Bestial Wrath and the racials at the pull and on cooldown. A knob
// search on each searched build (2026-10-12, 180 sec) kept the starting knobs.
export const APLBeastMastery = PresetUtils.makePresetAPLRotation('Beast Mastery', BeastMasteryAPL, {
	tooltip: 'Serpent Sting, Aimed Shot and Arcane Shot on cooldown, Bestial Wrath and Rapid Fire on cooldown.',
	group: 'Level 60',
});
export const APLMarksmanship = PresetUtils.makePresetAPLRotation('Marksmanship', MarksmanshipAPL, {
	tooltip: 'Serpent Sting, Aimed Shot and Arcane Shot on cooldown, Rapid Fire on cooldown. Also the Lone Wolf rotation.',
	group: 'Level 60',
});
// Melee: Mongoose Bite whenever Expose Prey opens it, Strider Kick and Raptor Strike on
// cooldown, Lacerate kept up, Immolation Trap on cooldown and Wing Clip in every free global
// cooldown. The Wing Clip filler was the knob search's one change (+31 at 180 sec): Survival
// runs out of global cooldowns long before it runs out of mana.
export const APLSurvival = PresetUtils.makePresetAPLRotation('Survival', SurvivalAPL, {
	tooltip: 'Melee: Mongoose Bite, Strider Kick, Raptor Strike, Lacerate, Immolation Trap, Wing Clip filler.',
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [APLBeastMastery, APLMarksmanship, APLSurvival],
};

export const DefaultAPL = APLBeastMastery;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent searches (tools/rotopt -talent-search, a hill climb) ran on the synthetic sets in
// the mksettings raid at 180 sec, from several starting builds per spec, on 2026-10-12.
//
// Beast Mastery 31/20: Bestial Wrath, Frenzy 4, Ferocity, Unleashed Fury, Focused Fire and
// Deadly Aspects, and Careful Aim, Lethal Attacks, Trueshot Aura and Mortal Shots 4 in
// Marksmanship. Both Beast Mastery searches dropped Summon Hawk: on the cooldown it shares with
// Arcane Shot, the hawk deals less than the shot. A 22/24/5 hybrid without Bestial Wrath and
// Frenzy (50201205050002-0053050301502-5) ended 25 behind with the same rotation.
export const TalentsBeastMastery = PresetUtils.makePresetTalents('Beast Mastery', SavedTalents.create({ talentsString: '5020121505021241-00520503014' }), {
	tooltip: 'Level 60 Beast Mastery 31/20 from the talent search.',
	group: 'Level 60',
});
// Marksmanship 7/39/5 with a pet: Efficiency, Careful Aim, Rapid Killing, Improved Arcane Shot,
// Trueshot Aura, Mortal Shots, Barrage, Ranged Weapon Specialization, and Improved Tracking 5.
// The search drops Sniper Shot: its 4 sec cast is worth about as much as the Aimed and Arcane
// Shots it pushes back. A search from a ranged Survival build ended on this one too.
export const TalentsMarksmanship = PresetUtils.makePresetTalents('Marksmanship', SavedTalents.create({ talentsString: '502-005355250150305-5' }), {
	tooltip: 'Level 60 Marksmanship 7/39/5 from the talent search, with a pet.',
	group: 'Level 60',
});
// Marksmanship with Lone Wolf (20% damage with no pet out) and the pet dismissed: the same
// shots, Rapid Recuperation 1 for the mana the Serpent Sting regen gives.
export const TalentsLoneWolf = PresetUtils.makePresetTalents('Lone Wolf', SavedTalents.create({ talentsString: '5-005355251151305-5' }), {
	tooltip: 'Level 60 Marksmanship 5/41/5 with Lone Wolf, played without a pet.',
	group: 'Level 60',
});
// Survival 7/10/34, the melee spec: Predator's Edge, Expose Prey, Strider Kick, Lightning
// Reflexes, Lacerating Strikes, Savage Strikes, Surefooted, Improved Tracking, Clever Traps and
// Survivalist's Discipline, Deadly Aspects (Quick Strikes with Aspect of the Beast), Focused
// Fire, Lethal Attacks and Careful Aim. Two searches from different starting builds ended
// here.
export const TalentsSurvival = PresetUtils.makePresetTalents('Survival', SavedTalents.create({ talentsString: '502-005005-500230230250122151' }), {
	tooltip: 'Level 60 melee Survival 7/10/34 from the talent search.',
	group: 'Level 60',
});

export const TalentPresets = {
	[Phase.Phase1]: [TalentsBeastMastery, TalentsMarksmanship, TalentsLoneWolf, TalentsSurvival],
};

export const DefaultTalents = TalentsBeastMastery;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// A cat at 1.0 sec: Broken Tooth (Badlands) is the one 1.0 sec beast under Forever, and a
// Forever pet hits as hard at any speed, so the fastest pet deals the most.
export const DefaultOptions = HunterOptions.create({
	ammo: Ammo.ThoriumHeadedArrow,
	quiverBonus: Hunter_Options_QuiverBonus.Speed15,
	petAttackSpeed: PetAttackSpeed.One,
	petType: PetType.Cat,
	petUptime: 1,
});

export const LoneWolfOptions = HunterOptions.create({
	...DefaultOptions,
	petType: PetType.PetNone,
});

export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	food: Food.FoodGrilledSquid,
	manaRegenElixir: ManaRegenElixir.MagebloodPotion,
	petAgilityConsumable: 1,
	petAttackPowerConsumable: 1,
	petStrengthConsumable: 1,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
});

// The raid the level 60 numbers were measured under (tools/rotopt/mksettings): both paladin
// blessings, Arcane Brilliance and every debuff, the hunter's own Hunter's Mark, and a shaman
// in the hunter's group for Grace of Air and Mana Spring. Trueshot Aura comes from the
// hunter's own talent, not from the raid.
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	leaderOfThePack: true,
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	sanctityAura: true,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
});

// Survival melees, so its shaman drops Windfury Totem and Strength of Earth for the melee group
// in place of Grace of Air.
export const SurvivalRaidBuffs = RaidBuffs.create({
	...DefaultRaidBuffs,
	graceOfAirTotem: TristateEffect.TristateEffectMissing,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	totemWeaponBuff: TotemWeaponBuff.TotemWeaponBuffWindfury,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
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
	huntersMark: TristateEffect.TristateEffectImproved,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	scorpidSting: true,
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

export const OtherDefaults = {
	distanceFromTarget: 25,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceOrc,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

const level60Boss = Encounter.create({
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
});

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor, Humanoid) in a 3
// minute fight, and the raid and consumes above. The builds were searched at 180 sec.
export const EncounterLevel60 = PresetUtils.makePresetEncounter('Level 60', level60Boss, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});
export const EncounterLevel60Melee = PresetUtils.makePresetEncounter('Level 60 Melee', level60Boss, {
	tanks: [],
	raidBuffs: SurvivalRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});

const professions: [Profession, Profession] = [OtherDefaults.profession1, OtherDefaults.profession2];

export const PresetBuildBeastMastery = PresetUtils.makePresetBuild('Level 60 Beast Mastery', {
	group: 'Level 60',
	tooltip: 'Level 60 Beast Mastery on the synthetic Phase 1 ranged gear with a 1.0 sec cat, in a 3 minute raid fight.',
	gear: GearLevel60Ranged,
	talents: TalentsBeastMastery,
	rotation: APLBeastMastery,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	professions,
	options: DefaultOptions,
});
export const PresetBuildMarksmanship = PresetUtils.makePresetBuild('Level 60 Marksmanship', {
	group: 'Level 60',
	tooltip: 'Level 60 Marksmanship on the synthetic Phase 1 ranged gear with a 1.0 sec cat, in a 3 minute raid fight.',
	gear: GearLevel60Ranged,
	talents: TalentsMarksmanship,
	rotation: APLMarksmanship,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	professions,
	options: DefaultOptions,
});
export const PresetBuildLoneWolf = PresetUtils.makePresetBuild('Level 60 Lone Wolf', {
	group: 'Level 60',
	tooltip: 'Level 60 Marksmanship with Lone Wolf and no pet, on the synthetic Phase 1 ranged gear, in a 3 minute raid fight.',
	gear: GearLevel60Ranged,
	talents: TalentsLoneWolf,
	rotation: APLMarksmanship,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	professions,
	options: LoneWolfOptions,
});
export const PresetBuildSurvival = PresetUtils.makePresetBuild('Level 60 Survival', {
	group: 'Level 60',
	tooltip: 'Level 60 melee Survival on the synthetic Phase 1 Survival gear, in a 3 minute raid fight with a melee group.',
	gear: GearLevel60Survival,
	talents: TalentsSurvival,
	rotation: APLSurvival,
	encounter: EncounterLevel60Melee,
	race: Race.RaceOrc,
	professions,
	options: DefaultOptions,
});
