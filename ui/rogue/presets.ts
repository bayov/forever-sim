import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	Flask,
	Food,
	IndividualBuffs,
	MobType,
	Profession,
	Race,
	RaidBuffs,
	SpellPowerBuff,
	StrengthBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
	SapperExplosive,
	Stat,
} from '../core/proto/common.js';
import { RogueOptions } from '../core/proto/rogue.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import BackstabAPL from './apls/combat_backstab.apl.json';
import BackstabRuptureAPL from './apls/combat_backstab_rupture.apl.json';
import BackstabOnCooldownAPL from './apls/combat_backstab_on_cooldown.apl.json';
import BackstabSndWindowAPL from './apls/combat_backstab_snd_window.apl.json';
import BackstabSweatyAPL from './apls/combat_backstab_sweaty.apl.json';
import SinisterStrikeAPL from './apls/combat_sinister_strike.apl.json';
import SinisterStrikeOnCooldownAPL from './apls/combat_sinister_strike_on_cooldown.apl.json';
import SinisterStrikeHoldForARAPL from './apls/combat_sinister_strike_hold_for_ar.apl.json';
import SinisterStrikeSndOpenerAPL from './apls/combat_sinister_strike_snd_opener.apl.json';
import SinisterStrikeSndWindowAPL from './apls/combat_sinister_strike_snd_window.apl.json';
import SinisterStrikeSweatyAPL from './apls/combat_sinister_strike_sweaty.apl.json';
import SinisterStrikeIEAAPL from './apls/combat_sinister_strike_iea.apl.json';
import MutilateAPL from './apls/forever_mutilate.apl.json';
import Level30p5BackstabAPL from './apls/level30p5_backstab.apl.json';
import Level30p5SinisterStrikeAPL from './apls/level30p5_sinister_strike.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import BackstabGearPreBiS from './gear_sets/combat_backstab_prebis.gear.json';
import SinisterStrikeGearPreBiS from './gear_sets/combat_sinister_strike_prebis.gear.json';
import BackstabGearP1BiS from './gear_sets/combat_backstab_p1_bis.gear.json';
import BackstabGearP2BiS from './gear_sets/combat_backstab_p2_bis.gear.json';
import SinisterStrikeGearP1BiS from './gear_sets/combat_sinister_strike_p1_bis.gear.json';
import SinisterStrikeGearP2BiS from './gear_sets/combat_sinister_strike_p2_bis.gear.json';
import Level30p5BackstabGear from './gear_sets/level30p5_backstab.gear.json';
import Level30p5SinisterStrikeGear from './gear_sets/level30p5_sinister_strike.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear, { tooltip: 'No gear equipped.' });
export const GearBackstabPreBiS = PresetUtils.makePresetGear('Backstab Pre-BiS', BackstabGearPreBiS, {
	tooltip: 'Level 60 pre-raid gear for Backstab, with daggers.',
	group: 'Level 60',
});
export const GearSinisterStrikePreBiS = PresetUtils.makePresetGear('Sinister Strike Pre-BiS', SinisterStrikeGearPreBiS, {
	tooltip: 'Level 60 pre-raid gear for Sinister Strike.',
	group: 'Level 60',
});
export const GearBackstabP1BiS = PresetUtils.makePresetGear('Backstab P1 BiS', BackstabGearP1BiS, {
	tooltip: 'Level 60 Phase 1 BiS for Backstab, with daggers.',
	group: 'Level 60',
});
export const GearBackstabP2BiS = PresetUtils.makePresetGear('Backstab P2 BiS', BackstabGearP2BiS, {
	tooltip: 'Level 60 Phase 2 BiS for Backstab, with daggers.',
	group: 'Level 60',
});
export const GearSinisterStrikeP1BiS = PresetUtils.makePresetGear('Sinister Strike P1 BiS', SinisterStrikeGearP1BiS, {
	tooltip: 'Level 60 Phase 1 BiS for Sinister Strike.',
	group: 'Level 60',
});
export const GearSinisterStrikeP2BiS = PresetUtils.makePresetGear('Sinister Strike P2 BiS', SinisterStrikeGearP2BiS, {
	tooltip: 'Level 60 Phase 2 BiS for Sinister Strike.',
	group: 'Level 60',
});
// The gear searches at level 30 with 26 talent points over everything an Undead rogue can
// equip, with Engineering and Leatherworking and the items no source is known for left
// out. Each ran from the old level 20 set (which keeps its four Defias pieces, because one
// slot at a time never breaks a set bonus) and from a seed with level 30 leather in those
// four slots. The second seed
// won both times, by 1.9 for Sinister Strike and 1.7 for Backstab.
//
// Both sets share Pathfinder Hat and Ghostwalker Boots, Ghostshard Talisman and Ironspine's
// Eye (Scarlet Monastery Graveyard), Unearthed Bands of Power (Uldaman), Infiltrator Armor,
// Tiger Hunter Gloves (Tiger Mastery), Triprunner Dungarees, Defiler's Chain Girdle (Arathi
// Basin, Horde) and Parachute Cloak (Engineering). Sinister Strike wields Ironspine's Fist
// and Silent Hunter (Horde quest Call to Arms). Backstab wields Silent Hunter in the main
// hand and Scout's Blade (Warsong Gulch, Horde) in the off hand. A quest reward can only
// be had once, so the search no longer puts Silent Hunter in both hands.
export const GearLevel30p5Backstab = PresetUtils.makePresetGear('Level 30 + 5 Backstab', Level30p5BackstabGear, {
	tooltip: "Level 30 Backstab gear with Silent Hunter and Scout's Blade.",
	group: 'Level 30',
});
export const GearLevel30p5SinisterStrike = PresetUtils.makePresetGear('Level 30 + 5 Sinister Strike', Level30p5SinisterStrikeGear, {
	tooltip: "Level 30 Sinister Strike gear with Ironspine's Fist and Silent Hunter.",
	group: 'Level 30',
});

export const GearPresets = {
	[Phase.Phase1]: [
		GearBackstabPreBiS,
		GearSinisterStrikePreBiS,
		GearBackstabP1BiS,
		GearSinisterStrikeP1BiS,
		GearLevel30p5Backstab,
		GearLevel30p5SinisterStrike,
	],
	[Phase.Phase2]: [
		GearBackstabPreBiS,
		GearSinisterStrikePreBiS,
		GearBackstabP2BiS,
		GearSinisterStrikeP2BiS,
		GearLevel30p5Backstab,
		GearLevel30p5SinisterStrike,
	],
};

export const DefaultGear = GearSinisterStrikePreBiS;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets[]
///////////////////////////////////////////////////////////////////////////

export const ROTATION_PRESET_BACKSTAB = PresetUtils.makePresetAPLRotation('Backstab', BackstabAPL, {
	tooltip: "The optimizer's best level 60 Backstab rotation.",
	group: 'Level 60',
});
export const ROTATION_PRESET_SINISTER_STRIKE = PresetUtils.makePresetAPLRotation('Sinister Strike', SinisterStrikeAPL, {
	tooltip: "The optimizer's best level 60 Sinister Strike rotation.",
	group: 'Level 60',
});
// Variants of the Backstab and Sinister Strike rotations, generated by
// tools/rotopt/rogue_presets.sh. The plain ones are the optimizer's best, these are the
// alternatives to compare against.
export const ROTATION_PRESET_BACKSTAB_RUPTURE = PresetUtils.makePresetAPLRotation('Backstab (Rupture at 5 points)', BackstabRuptureAPL, {
	tooltip: 'Backstab variant that uses Rupture as a 5 point finisher.',
	group: 'Level 60',
});
export const ROTATION_PRESET_BACKSTAB_ON_COOLDOWN = PresetUtils.makePresetAPLRotation('Backstab (cooldowns on cooldown)', BackstabOnCooldownAPL, {
	tooltip: 'Backstab variant that uses every cooldown as soon as it is ready.',
	group: 'Level 60',
});
export const ROTATION_PRESET_BACKSTAB_SND_WINDOW = PresetUtils.makePresetAPLRotation(
	'Backstab (cooldowns in a 20s Slice and Dice window)',
	BackstabSndWindowAPL,
	{ tooltip: 'Backstab variant that saves the cooldowns for a window with 20 sec of Slice and Dice.', group: 'Level 60' },
);
export const ROTATION_PRESET_SINISTER_STRIKE_ON_COOLDOWN = PresetUtils.makePresetAPLRotation('SS (cooldowns on cooldown)', SinisterStrikeOnCooldownAPL, {
	tooltip: 'Sinister Strike variant that uses every cooldown as soon as it is ready.',
	group: 'Level 60',
});
export const ROTATION_PRESET_SINISTER_STRIKE_HOLD_FOR_AR = PresetUtils.makePresetAPLRotation(
	'SS (everything with Adrenaline Rush)',
	SinisterStrikeHoldForARAPL,
	{ tooltip: 'Sinister Strike variant that holds every cooldown for Adrenaline Rush.', group: 'Level 60' },
);
export const ROTATION_PRESET_SINISTER_STRIKE_SND_OPENER = PresetUtils.makePresetAPLRotation(
	'SS (5 point Slice and Dice before cooldowns)',
	SinisterStrikeSndOpenerAPL,
	{ tooltip: 'Sinister Strike variant that puts up a 5 point Slice and Dice before the cooldowns.', group: 'Level 60' },
);
export const ROTATION_PRESET_SINISTER_STRIKE_SND_WINDOW = PresetUtils.makePresetAPLRotation(
	'SS (cooldowns in a 20s Slice and Dice window)',
	SinisterStrikeSndWindowAPL,
	{ tooltip: 'Sinister Strike variant that saves the cooldowns for a window with 20 sec of Slice and Dice.', group: 'Level 60' },
);
export const ROTATION_PRESET_BACKSTAB_SWEATY = PresetUtils.makePresetAPLRotation('Backstab (Sweaty)', BackstabSweatyAPL, {
	tooltip: 'Sweaty variant of the Backstab rotation.',
	group: 'Level 60',
});
export const ROTATION_PRESET_SINISTER_STRIKE_SWEATY = PresetUtils.makePresetAPLRotation('Sinister Strike (Sweaty)', SinisterStrikeSweatyAPL, {
	tooltip: 'Sweaty variant of the Sinister Strike rotation.',
	group: 'Level 60',
});
export const ROTATION_PRESET_SINISTER_STRIKE_IEA = PresetUtils.makePresetAPLRotation('Improved Expose Armor (SS)', SinisterStrikeIEAAPL, {
	tooltip: 'Sinister Strike rotation that also keeps Improved Expose Armor up.',
	group: 'Level 60',
});
export const ROTATION_PRESET_MUTILATE = PresetUtils.makePresetAPLRotation('Mutilate', MutilateAPL, {
	tooltip: 'Level 60 Mutilate rotation.',
	group: 'Level 60',
});
// The level 30 rotations (tools/rotopt/rogue_presets.sh): Slice and Dice, Rupture, the
// builder whenever the energy is there and Thistle Tea when low. The lines name the level
// 60 spell ranks and the sim resolves them to the rank the level knows, so the icons show a
// higher rank than what is cast. Sinister Strike puts Slice and
// Dice up at 2 points (+0.2 at 60 sec, a 4 point one is 1.5 ahead at 30 sec but 0.4 behind
// at 60). Backstab Ruptures at 5 points (+0.5) and Vanishes into an Ambush at 3 points or
// fewer (+0.6).
export const ROTATION_PRESET_LEVEL30P5_BACKSTAB = PresetUtils.makePresetAPLRotation('Level 30 + 5 Backstab', Level30p5BackstabAPL, {
	tooltip: 'Level 30 Backstab rotation. It Ruptures at 5 points and Vanishes into Ambush.',
	group: 'Level 30',
});
export const ROTATION_PRESET_LEVEL30P5_SINISTER_STRIKE = PresetUtils.makePresetAPLRotation('Level 30 + 5 Sinister Strike', Level30p5SinisterStrikeAPL, {
	tooltip: 'Level 30 Sinister Strike rotation with a 2 point Slice and Dice.',
	group: 'Level 30',
});

export const APLPresets = {
	[Phase.Phase1]: [
		ROTATION_PRESET_BACKSTAB,
		ROTATION_PRESET_BACKSTAB_RUPTURE,
		ROTATION_PRESET_BACKSTAB_ON_COOLDOWN,
		ROTATION_PRESET_BACKSTAB_SND_WINDOW,
		ROTATION_PRESET_SINISTER_STRIKE,
		ROTATION_PRESET_SINISTER_STRIKE_ON_COOLDOWN,
		ROTATION_PRESET_SINISTER_STRIKE_HOLD_FOR_AR,
		ROTATION_PRESET_SINISTER_STRIKE_SND_OPENER,
		ROTATION_PRESET_SINISTER_STRIKE_SND_WINDOW,
		ROTATION_PRESET_BACKSTAB_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_IEA,
		ROTATION_PRESET_MUTILATE,
		ROTATION_PRESET_LEVEL30P5_BACKSTAB,
		ROTATION_PRESET_LEVEL30P5_SINISTER_STRIKE,
	],
	[Phase.Phase2]: [
		ROTATION_PRESET_BACKSTAB,
		ROTATION_PRESET_BACKSTAB_RUPTURE,
		ROTATION_PRESET_BACKSTAB_ON_COOLDOWN,
		ROTATION_PRESET_BACKSTAB_SND_WINDOW,
		ROTATION_PRESET_SINISTER_STRIKE,
		ROTATION_PRESET_SINISTER_STRIKE_ON_COOLDOWN,
		ROTATION_PRESET_SINISTER_STRIKE_HOLD_FOR_AR,
		ROTATION_PRESET_SINISTER_STRIKE_SND_OPENER,
		ROTATION_PRESET_SINISTER_STRIKE_SND_WINDOW,
		ROTATION_PRESET_BACKSTAB_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_IEA,
		ROTATION_PRESET_MUTILATE,
		ROTATION_PRESET_LEVEL30P5_BACKSTAB,
		ROTATION_PRESET_LEVEL30P5_SINISTER_STRIKE,
	],
};

//Need to add main hand equip logic or talent/rotation logic to map to Auto APL
export const DefaultAPLs: Record<number, PresetUtils.PresetRotation> = {
	[0]: ROTATION_PRESET_SINISTER_STRIKE,
	[1]: ROTATION_PRESET_BACKSTAB,
};

export const DefaultAPLBackstab = ROTATION_PRESET_BACKSTAB;
export const DefaultAPLSinisterStrike = ROTATION_PRESET_SINISTER_STRIKE;
export const DefaultAPLIEA = ROTATION_PRESET_SINISTER_STRIKE_IEA;
export const DefaultAPLMutilate = ROTATION_PRESET_MUTILATE;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/forever/talent-calc and copy the numbers in the url.

// Preset name must be unique. Ex: 'Backstab DPS' cannot be used as a name more than once

export const CombatBackstabTalents = PresetUtils.makePresetTalents('Backstab', SavedTalents.create({ talentsString: '005323102-30230320201515231-002' }), {
	tooltip: 'Level 60 Combat talents for Backstab with daggers.',
	group: 'Level 60',
});
export const CombatSinisterStrikeTalents = PresetUtils.makePresetTalents(
	'Sinister Strike',
	SavedTalents.create({ talentsString: '00532310301-32003311201515231' }),
	{ tooltip: 'Level 60 Combat talents for Sinister Strike.', group: 'Level 60' },
);
export const CombatSinisterStrikeIEATalents = PresetUtils.makePresetTalents(
	'Improved Expose Armor (SS)',
	SavedTalents.create({ talentsString: '005323123-32003311201515131' }),
	{ tooltip: 'Level 60 Sinister Strike talents with Improved Expose Armor.', group: 'Level 60' },
);
export const AssassinationMutilateTalents = PresetUtils.makePresetTalents(
	'Mutilate',
	SavedTalents.create({ talentsString: '00530310551021051-302303202004' }),
	{ tooltip: 'Level 60 Assassination talents for Mutilate.', group: 'Level 60' },
);
// Level 30 with 5 extra talent points (26 in all). With 26 points Seal Fate, Mutilate,
// Blade Flurry, Hack and Slash, Aggression and Weapon Expertise are all in reach, so we
// scored every build over the damage talents (Malice 5 and Murder 2 kept), split into
// deep Assassination, deep Combat and the builds in between (23517 for Sinister Strike,
// 31065 for Backstab). Deep Combat and Seal Fate lose, and so does Mutilate: the best
// Mutilate build does 107.4 at 60 sec where Backstab does 119.6 on the same daggers.
//
// The top builds are within 0.3 DPS of each other. Sinister Strike: Malice 5,
// Ruthlessness 3, Murder 2, Imp SnD 3, Relentless Strikes, Lethality 1, Vile Poisons 3,
// Imp Eviscerate 3, Imp SS 2 and Precision 3. Taking Vile Poisons 5, Improved Poisons and
// Vigor in place of the Combat points ties it. Backstab: Malice 5, Ruthlessness 1, Murder
// 2, Imp SnD 2, Relentless Strikes, Lethality 2, Imp Eviscerate 3, Lightning Reflexes 2,
// Puncturing Wounds 3, Precision 3 and Opportunity 2.
export const Level30p5SinisterStrikeTalents = PresetUtils.makePresetTalents(
	'Level 30 + 5 Sinister Strike',
	SavedTalents.create({ talentsString: '0053231013-320003' }),
	{ tooltip: 'Level 30 with 5 extra talent points for Sinister Strike, mostly Assassination.', group: 'Level 30' },
);
export const Level30p5BackstabTalents = PresetUtils.makePresetTalents('Level 30 + 5 Backstab', SavedTalents.create({ talentsString: '005122102-302303-002' }), {
	tooltip: 'Level 30 with 5 extra talent points for Backstab, split between Assassination and Combat.',
	group: 'Level 30',
});

export const TalentPresets = {
	[Phase.Phase1]: [
		CombatBackstabTalents,
		CombatSinisterStrikeTalents,
		CombatSinisterStrikeIEATalents,
		AssassinationMutilateTalents,
		Level30p5BackstabTalents,
		Level30p5SinisterStrikeTalents,
	],
	[Phase.Phase2]: [
		CombatBackstabTalents,
		CombatSinisterStrikeTalents,
		CombatSinisterStrikeIEATalents,
		AssassinationMutilateTalents,
		Level30p5BackstabTalents,
		Level30p5SinisterStrikeTalents,
	],
};

export const DefaultTalentsAssassin = TalentPresets[Phase.Phase2][0];
export const DefaultTalentsCombat = TalentPresets[Phase.Phase2][0];
export const DefaultTalentsSubtlety = TalentPresets[Phase.Phase2][0];

export const DefaultTalentsBackstab = TalentPresets[Phase.Phase2][0];
export const DefaultTalentsSinisterStrike = TalentPresets[Phase.Phase2][1];
export const DefaultTalentsIEA = TalentPresets[Phase.Phase2][2];
export const DefaultTalentsMutilate = TalentPresets[Phase.Phase2][3];

export const DefaultTalents = DefaultTalentsSinisterStrike;

///////////////////////////////////////////////////////////////////////////
//                                Build Presets
///////////////////////////////////////////////////////////////////////////
export const PresetBuildBackstab = PresetUtils.makePresetBuild('Backstab', {
	group: 'Level 60',
	tooltip: 'Level 60 Combat Backstab on Phase 2 BiS daggers.',
	gear: GearBackstabP2BiS,
	talents: DefaultTalentsBackstab,
	rotation: DefaultAPLBackstab,
});
export const PresetBuildSinisterStrike = PresetUtils.makePresetBuild('Sinister Strike', {
	group: 'Level 60',
	tooltip: 'Level 60 Combat Sinister Strike on Phase 2 BiS gear.',
	gear: GearSinisterStrikeP2BiS,
	talents: DefaultTalentsSinisterStrike,
	rotation: DefaultAPLSinisterStrike,
});
export const PresetBuildIEA = PresetUtils.makePresetBuild('IEA', {
	group: 'Level 60',
	tooltip: 'Level 60 Sinister Strike with Improved Expose Armor, on Phase 2 BiS gear.',
	gear: GearSinisterStrikeP2BiS,
	talents: DefaultTalentsIEA,
	rotation: DefaultAPLIEA,
});
export const PresetBuildMutilate = PresetUtils.makePresetBuild('Mutilate', {
	group: 'Level 60',
	tooltip: 'Level 60 Assassination Mutilate on Phase 2 BiS daggers.',
	gear: GearBackstabP2BiS,
	talents: DefaultTalentsMutilate,
	rotation: DefaultAPLMutilate,
});

// A level 30 rogue soloing Interrogator Vishas (Scarlet Monastery Graveyard, level 32) for
// 60 sec. There are no raid buffs or debuffs, since the buff table holds level 60 values.
// Blessing of Kings is the exception: it is 10% of every attribute rather than a number off
// the level 60 table, so it is right at any level, and a group at 30 almost always has a
// paladin in it. Deadly Poison (level 30) goes in the main hand and Instant Poison in the
// off hand. That is 0.3 ahead of Deadly Poison on both weapons at 60 sec, and 0.9 to 4.2
// ahead of the other pairs. Thistle Tea is the conjured item.
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
			},
		],
	}),
	{
		raidBuffs: RaidBuffs.create({}),
		debuffs: Debuffs.create({}),
		buffs: IndividualBuffs.create({ blessingOfKings: true }),
		consumes: Consumes.create({
			defaultConjured: Conjured.ConjuredRogueThistleTea,
			mainHandImbue: WeaponImbue.DeadlyPoison,
			offHandImbue: WeaponImbue.InstantPoison,
		}),
	},
);
// Level 30 with 5 extra talent points. At 30 / 60 / 300 sec that is 137.9 / 130.4 / 122.8
// for Sinister Strike and 155.1 / 141.8 / 132.0 for Backstab. Backstab pulls ahead at 30
// because the daggers it gets there (Silent Hunter, Scout's Blade) are much better than
// the level 20 ones were.
export const PresetBuildLevel30p5Backstab = PresetUtils.makePresetBuild('Level 30 + 5 Backstab', {
	group: 'Level 30',
	tooltip: 'Level 30 Undead Backstab with 5 extra talent points, soloing Interrogator Vishas.',
	gear: GearLevel30p5Backstab,
	talents: Level30p5BackstabTalents,
	rotation: ROTATION_PRESET_LEVEL30P5_BACKSTAB,
	encounter: EncounterLevel30,
	race: Race.RaceUndead,
	level: 30,
	bonusTalentPoints: 5,
});
export const PresetBuildLevel30p5SinisterStrike = PresetUtils.makePresetBuild('Level 30 + 5 Sinister Strike', {
	group: 'Level 30',
	tooltip: 'Level 30 Undead Sinister Strike with 5 extra talent points, soloing Interrogator Vishas.',
	gear: GearLevel30p5SinisterStrike,
	talents: Level30p5SinisterStrikeTalents,
	rotation: ROTATION_PRESET_LEVEL30P5_SINISTER_STRIKE,
	encounter: EncounterLevel30,
	race: Race.RaceUndead,
	level: 30,
	bonusTalentPoints: 5,
});

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = RogueOptions.create({});

///////////////////////////////////////////////////////////////////////////
//                         Consumes/Buffs/Debuffs
///////////////////////////////////////////////////////////////////////////

export const P1Consumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredRogueThistleTea,
	dragonBreathChili: true,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodGrilledSquid,
	mainHandImbue: WeaponImbue.InstantPoison,
	offHandImbue: WeaponImbue.DeadlyPoison,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
});

export const DefaultConsumes = {
	[Phase.Phase1]: P1Consumes,
};

export const P1RaidBuffs = RaidBuffs.create({
	battleShout: TristateEffect.TristateEffectImproved,
	fireResistanceAura: true,
	fireResistanceTotem: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
	graceOfAirTotem: TristateEffect.TristateEffectRegular,
	leaderOfThePack: true,
	trueshotAura: true,
});

export const DefaultRaidBuffs = {
	[Phase.Phase1]: P1RaidBuffs,
};

export const P1IndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	fengusFerocity: false,
	rallyingCryOfTheDragonslayer: false,
	slipkiksSavvy: false,
	songflowerSerenade: false,
	spiritOfZandalar: false,
	warchiefsBlessing: false,
});

export const DefaultIndividualBuffs = {
	[Phase.Phase1]: P1IndividualBuffs,
};

export const P1DefaultDebuffs = Debuffs.create({
	curseOfRecklessness: true,
	faerieFire: true,
	sunderArmor: true,
});

export const DefaultDebuffs = {
	[Phase.Phase1]: P1DefaultDebuffs,
};

export const P1OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
};

export const OtherDefaults = {
	[Phase.Phase1]: P1OtherDefaults,
};
