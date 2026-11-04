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
	Flask,
	Food,
	IndividualBuffs,
	MobType,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	Stat,
	StrengthBuff,
	SunderRamp,
	TotemWeaponBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { RogueOptions } from '../core/proto/rogue.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import Level60CombatDaggersAPL from './apls/level60_combat_daggers.apl.json';
import Level60CombatSwordsAPL from './apls/level60_combat_swords.apl.json';
import Level60MutilateAPL from './apls/level60_mutilate.apl.json';
import Level60SealFateDaggersAPL from './apls/level60_seal_fate_daggers.apl.json';
import Level60SealFateSwordsAPL from './apls/level60_seal_fate_swords.apl.json';
import Level60SubtletyAPL from './apls/level60_subtlety.apl.json';
import Level60SyntheticCombatDaggersGear from './gear_sets/level60_phase1_synthetic_combat_daggers.gear.json';
import Level60SyntheticDaggersGear from './gear_sets/level60_phase1_synthetic_daggers.gear.json';
import Level60SyntheticMutilateGear from './gear_sets/level60_phase1_synthetic_mutilate.gear.json';
import Level60SyntheticSealFateSwordsGear from './gear_sets/level60_phase1_synthetic_seal_fate_swords.gear.json';
import Level60SyntheticSubtletyGear from './gear_sets/level60_phase1_synthetic_subtlety.gear.json';
import Level60SyntheticSwordsGear from './gear_sets/level60_phase1_synthetic_swords.gear.json';
import Level60SyntheticSwordsHitPlus2Gear from './gear_sets/level60_phase1_synthetic_swords_hit+2.gear.json';
import Level60SyntheticCombatDaggersHitPlus2Gear from './gear_sets/level60_phase1_synthetic_combat_daggers_hit+2.gear.json';
import Level60SyntheticSealFateSwordsHitPlus3Gear from './gear_sets/level60_phase1_synthetic_seal_fate_swords_hit+3.gear.json';
import Level60SyntheticDaggersHitPlus3Gear from './gear_sets/level60_phase1_synthetic_daggers_hit+3.gear.json';
import Level60SyntheticSubtletyHitPlus2Gear from './gear_sets/level60_phase1_synthetic_subtlety_hit+2.gear.json';
import Level60SyntheticMutilateHitPlus3Gear from './gear_sets/level60_phase1_synthetic_mutilate_hit+3.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// A guess at Phase 1 BiS under Forever, made of the synthetic items in
// tools/database/forever_synthetic_items_rogue.go (ids from 994000, not in the game).
//
// Forever's level 60 client gear is not tuned yet, and Molten Core and Onyxia have no Forever
// data. So, as for the enhancement shaman, each piece takes the place of an Era Phase 1 BiS item
// (the Nightslayer and Devilsaur sets, Onyxia Tooth Pendant, Cape of the Black Baron, Quick Strike
// Ring, Don Julio's Band, Striker's Mark) on the shared item level table (66 for the tier slots, 71
// for the rest), priced by the cross-class Forever budget for physical melee (three stat lines a
// piece) in the fair stat mix of Season of Discovery rogue BiS (tools/database/forever_synthetic_items_rogue.go
// has the numbers). The swords set is for Sinister Strike, the daggers set for Backstab, Mutilate and
// Subtlety. The weapons have the type and speed each build sims best with among Era Phase 1 BiS speeds
// (standards 22), so Seal Fate Swords and Mutilate get their own copies of the sets with different
// weapons. Both carry 4% hit, the 8% cap with Precision 3/3 and the hands and legs, which carry
// the 2 piece bonus of Grimstitch Armor (Forever's crafted rogue set, 1% hit; Forever has no hit
// suppression, so the yellow cap against a level 63 boss is 8%). Subtlety takes no Precision and
// wears the daggers set with a hit wrist instead of the attack power one (5% on the gear, 6% in all,
// under the cap). The 4 and 5 piece
// bonuses (36 attack power against Humanoids, Eviscerate 5 Energy cheaper) are not live, see
// PresetBuildLevel60CombatSwords. The trinkets are synthetic too (31 Agility and 15 crit rating, and
// +335 attack power for 20 sec every 2 min): Forever removed Hand of Justice and Blackhand's Breadth,
// and the sim does not model its launch trinkets yet.
//
// The enchants are the best non-raid ones: Lesser Arcanum of Voracity (Agility) on head and legs,
// Enchant Necklace - Agility, Cloak - Agility (+5), Chest - Greater Stats, Superior Strength on
// bracers (+9) and gloves (+15, 1.3 DPS ahead of Greater Agility for Backstab), Boots - Greater
// Agility and Crusader on both weapons (3 ahead of Weapon - Agility).
export const GearLevel60SyntheticSwords = PresetUtils.makePresetGear('Combat Swords', Level60SyntheticSwordsGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Sinister Strike, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});
export const GearLevel60SyntheticDaggers = PresetUtils.makePresetGear('Seal Fate Daggers', Level60SyntheticDaggersGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Seal Fate Backstab, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});
// The daggers set with 10 hit rating on the wrist in place of 20 attack power: Subtlety takes no
// Precision, so its 6% hit is under the 8% cap.
export const GearLevel60SyntheticSubtlety = PresetUtils.makePresetGear('Subtlety', Level60SyntheticSubtletyGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Subtlety, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});
// The swords set with a 1.6 sec dagger off hand, which Seal Fate sims best with.
export const GearLevel60SyntheticSealFateSwords = PresetUtils.makePresetGear('Seal Fate Swords', Level60SyntheticSealFateSwordsGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Seal Fate Sinister Strike, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});
// The daggers set with a 2.5 sec mace off hand, which Combat Daggers sims best with since Windfury Totem
// procs from the main hand only (a mace takes Hack and Slash's armor ignore).
export const GearLevel60SyntheticCombatDaggers = PresetUtils.makePresetGear('Combat Daggers', Level60SyntheticCombatDaggersGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Combat Backstab, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});
// The daggers set with a 1.7 sec dagger main hand, which Mutilate sims best with, and the set's 1.6 sec
// dagger off hand: Mutilate needs a dagger in both hands (client 70338, 2026-10-23).
export const GearLevel60SyntheticMutilate = PresetUtils.makePresetGear('Mutilate', Level60SyntheticMutilateGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Mutilate, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

// The hit-trade steps (round 70753 phase 5, orchestration/round_70753/hit_trade/rogue.md): each build's set with 2 or 3% more
// gear hit in place of other stats (12 to 13% total hit against the 8% yellow cap; the poisons and white swings still gain).
// The six builds wear them; talents and rotations are unchanged.
export const GearLevel60SyntheticSwordsHitPlus2 = PresetUtils.makePresetGear('Combat Swords (hit+2)', Level60SyntheticSwordsHitPlus2Gear, {
	tooltip: 'The swords set with 2% more hit in place of other stats, for Sinister Strike.',
	group: 'Synthetic',
});
export const GearLevel60SyntheticCombatDaggersHitPlus2 = PresetUtils.makePresetGear('Combat Daggers (hit+2)', Level60SyntheticCombatDaggersHitPlus2Gear, {
	tooltip: 'The combat daggers set with 2% more hit in place of other stats, for Combat Backstab.',
	group: 'Synthetic',
});
export const GearLevel60SyntheticSealFateSwordsHitPlus3 = PresetUtils.makePresetGear('Seal Fate Swords (hit+3)', Level60SyntheticSealFateSwordsHitPlus3Gear, {
	tooltip: 'The seal fate swords set with 3% more hit in place of other stats, for Seal Fate Sinister Strike.',
	group: 'Synthetic',
});
export const GearLevel60SyntheticDaggersHitPlus3 = PresetUtils.makePresetGear('Seal Fate Daggers (hit+3)', Level60SyntheticDaggersHitPlus3Gear, {
	tooltip: 'The seal fate daggers set with 3% more hit in place of other stats, for Seal Fate Backstab.',
	group: 'Synthetic',
});
export const GearLevel60SyntheticSubtletyHitPlus2 = PresetUtils.makePresetGear('Subtlety (hit+2)', Level60SyntheticSubtletyHitPlus2Gear, {
	tooltip: 'The subtlety set with 2% more hit in place of other stats, for Subtlety.',
	group: 'Synthetic',
});
export const GearLevel60SyntheticMutilateHitPlus3 = PresetUtils.makePresetGear('Mutilate (hit+3)', Level60SyntheticMutilateHitPlus3Gear, {
	tooltip: 'The mutilate set with 3% more hit in place of other stats, for Mutilate.',
	group: 'Synthetic',
});

export const GearPresets = {
	[Phase.Phase1]: [
		GearLevel60SyntheticSwords,
		GearLevel60SyntheticDaggers,
		GearLevel60SyntheticSubtlety,
		GearLevel60SyntheticSealFateSwords,
		GearLevel60SyntheticCombatDaggers,
		GearLevel60SyntheticMutilate,
		GearLevel60SyntheticSwordsHitPlus2,
		GearLevel60SyntheticCombatDaggersHitPlus2,
		GearLevel60SyntheticSealFateSwordsHitPlus3,
		GearLevel60SyntheticDaggersHitPlus3,
		GearLevel60SyntheticSubtletyHitPlus2,
		GearLevel60SyntheticMutilateHitPlus3,
	],
	[Phase.Phase2]: [
		GearLevel60SyntheticSwords,
		GearLevel60SyntheticDaggers,
		GearLevel60SyntheticSubtlety,
		GearLevel60SyntheticSealFateSwords,
		GearLevel60SyntheticCombatDaggers,
		GearLevel60SyntheticMutilate,
		GearLevel60SyntheticSwordsHitPlus2,
		GearLevel60SyntheticCombatDaggersHitPlus2,
		GearLevel60SyntheticSealFateSwordsHitPlus3,
		GearLevel60SyntheticDaggersHitPlus3,
		GearLevel60SyntheticSubtletyHitPlus2,
		GearLevel60SyntheticMutilateHitPlus3,
	],
};

// Combat Swords is the class default since round 70753 (November 2 notes): with Aggression off Backstab and
// Flash of Steel on Shank it leads Combat Daggers by 16.8 at 180 sec (963.5 against 946.7). Combat Daggers
// was the default from 2026-10-23.
export const DefaultGear = GearLevel60SyntheticSwordsHitPlus2;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets[]
///////////////////////////////////////////////////////////////////////////

// The Level 60 rotations, from the tools/rotopt searches on the synthetic Phase 1 sets with the
// talents below (knob search, talent search, knob search again, 180 sec, see rogue_presets.sh).
// Seal Fate Daggers used the Combat Daggers rotation until 2026-11-02 (see below). Seal Fate Swords
// is the Combat Swords rotation with 4 point Eviscerates (2026-10-23, on the 8% sets: Seal Fate's extra
// points make the fifth come too late, +0.1 to +1.0 at 120 / 180 / 300 sec on both seed sets, +0.7 to
// +2.0 with its new talents). Mutilate Vanishes into an Ambush at 3 points or fewer since the same day
// (+0.7 to +1.8).
// Combat Daggers and Seal Fate Daggers are Gnome since 2026-10-27, so their rotation puts Eureka! on
// cooldown (eurekaCp 0) instead of holding it for a 5 point Eviscerate these 4 point Eviscerate builds
// seldom reach: +8.9 / +5.4 at 180 sec for the Gnome (20000 iterations, seed 1), 0.0 for other races.
// Both are Undead since the second re-pick that day; the rotation keeps Eureka! on cooldown (a no-op for
// the Undead), so it stays the Gnome's rotation too.
// Since the 2026-11-02 search (120, 180 and 300 sec together, each change kept only for 2 standard errors
// at 20000 iterations on seeds 1 and 1000003) every builder goes out whenever it is affordable: an extra
// attack keeps the swing timer under Forever since 2026-10-30, so a builder timed after a main hand swing
// only waits. Combat Daggers also drops its Vanish into Ambush and refreshes Slice and Dice with 1 sec left
// instead of 3, which costs Seal Fate Daggers 0.3, so Seal Fate Daggers has its own rotation again (the
// Combat Daggers one with the Vanish and the 3 sec refresh). The numbers are above PresetBuildLevel60CombatDaggers.
export const ROTATION_PRESET_LEVEL60_COMBAT_SWORDS = PresetUtils.makePresetAPLRotation('Combat Swords', Level60CombatSwordsAPL, {
	tooltip: 'Sinister Strike, Slice and Dice and 5 point Eviscerates, Blade Flurry before the pull.',
	group: 'Synthetic',
});
export const ROTATION_PRESET_LEVEL60_SEAL_FATE_SWORDS = PresetUtils.makePresetAPLRotation('Seal Fate Swords', Level60SealFateSwordsAPL, {
	tooltip: 'Sinister Strike, Slice and Dice and 4 point Eviscerates, Blade Flurry before the pull.',
	group: 'Synthetic',
});
export const ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS = PresetUtils.makePresetAPLRotation('Combat Daggers', Level60CombatDaggersAPL, {
	tooltip: 'Backstab, Slice and Dice and 4 point Eviscerates, Blade Flurry before the pull, Eureka! on cooldown.',
	group: 'Synthetic',
});
export const ROTATION_PRESET_LEVEL60_SEAL_FATE_DAGGERS = PresetUtils.makePresetAPLRotation('Seal Fate Daggers', Level60SealFateDaggersAPL, {
	tooltip: 'Backstab, Slice and Dice and 4 point Eviscerates, Vanish into Ambush at 3 points or fewer, Eureka! on cooldown.',
	group: 'Synthetic',
});
export const ROTATION_PRESET_LEVEL60_MUTILATE = PresetUtils.makePresetAPLRotation('Mutilate', Level60MutilateAPL, {
	tooltip: 'Mutilate, Slice and Dice and 4 point Eviscerates, Vanish into Ambush at 3 points or fewer.',
	group: 'Synthetic',
});
export const ROTATION_PRESET_LEVEL60_SUBTLETY = PresetUtils.makePresetAPLRotation('Subtlety', Level60SubtletyAPL, {
	tooltip: 'Hemorrhage, Ghostly Strike on cooldown, a 3 point Rupture, Premeditation and Vanish into Ambush at 3 points or fewer.',
	group: 'Synthetic',
});

export const APLPresets = {
	[Phase.Phase1]: [
		ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
		ROTATION_PRESET_LEVEL60_SEAL_FATE_SWORDS,
		ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
		ROTATION_PRESET_LEVEL60_SEAL_FATE_DAGGERS,
		ROTATION_PRESET_LEVEL60_MUTILATE,
		ROTATION_PRESET_LEVEL60_SUBTLETY,
	],
	[Phase.Phase2]: [
		ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
		ROTATION_PRESET_LEVEL60_SEAL_FATE_SWORDS,
		ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
		ROTATION_PRESET_LEVEL60_SEAL_FATE_DAGGERS,
		ROTATION_PRESET_LEVEL60_MUTILATE,
		ROTATION_PRESET_LEVEL60_SUBTLETY,
	],
};

//Need to add main hand equip logic or talent/rotation logic to map to Auto APL
export const DefaultAPLs: Record<number, PresetUtils.PresetRotation> = {
	[0]: ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
	[1]: ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
};

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/forever/talent-calc and copy the numbers in the url.

// Preset name must be unique. Ex: 'Backstab DPS' cannot be used as a name more than once

// Round 70753 (November 2 development notes, the reworked tree, since 2026-11-04 read from server hotfix push 113483 (1.60.1.70707+hf); strings remapped to its order):
// every build searched again on the new tree. The talent search moved points in batches of 1 to 5 at a
// time (150 valid neighbours a round at 3000 iterations, the best 5 checked at 10000, kept for 2 standard
// errors), alternated with the knob search over 120, 180 and 300 sec until neither moved. The numbers are
// above PresetBuildLevel60CombatDaggers. Before round 70753:
// The Level 60 talents, from the tools/rotopt talent searches on the synthetic Phase 1 sets (180 sec,
// 3000 iterations a candidate, the best moves checked again at 10000), seeded from each archetype
// and alternated with the knob search until neither moved. See PresetBuildLevel60CombatDaggers for
// the numbers. These are for the generic boss (no creature type), where Murder does nothing, so its
// two points go to Lethality (Vile Poisons for Seal Fate Daggers, whose Lethality is full). The
// "(vs Humanoid)" talents below keep Murder 2/2 for Humanoid and Giant bosses. On the 8% sets
// (2026-10-23) the searches kept Combat Swords, Combat Daggers, Mutilate and Subtlety. Seal Fate Daggers
// moved 3 points of Vile Poisons into Improved Poisons 5/5 (+0.3 to +0.7 at 120 / 180 / 300 sec on both
// seed sets; a tie against a Humanoid, so its vs Humanoid talents stay), and Seal Fate Swords 1 point
// (+0.6 to +1.0 with its 4 point Eviscerates, on both boss types).
export const TalentsLevel60CombatSwords = PresetUtils.makePresetTalents(
	'Combat Swords',
	SavedTalents.create({ talentsString: '00530320201-3200330112012321551' }),
	{ tooltip: 'Deep Combat for Sinister Strike: Adrenaline Rush, Flash of Steel, Blade Flurry, Hack and Slash, Aggression, Cold Blood.', group: 'Synthetic' },
);
export const TalentsLevel60SealFateSwords = PresetUtils.makePresetTalents(
	'Seal Fate Swords',
	SavedTalents.create({ talentsString: '0053032024105103-320033011201132' }),
	{
		tooltip: 'Seal Fate, Relentless Strikes, Venom and Cold Blood in Assassination, with Aggression and Dual Wield Specialization, for Sinister Strike.',
		group: 'Synthetic',
	},
);
export const TalentsLevel60CombatDaggers = PresetUtils.makePresetTalents(
	'Combat Daggers',
	SavedTalents.create({ talentsString: '005203002-3203230002012221551-005' }),
	{ tooltip: 'Deep Combat for Backstab: Adrenaline Rush, Flash of Steel, Blade Flurry, Puncturing Wounds, Hack and Slash, Opportunity.', group: 'Synthetic' },
);
export const TalentsLevel60SealFateDaggers = PresetUtils.makePresetTalents(
	'Seal Fate Daggers',
	SavedTalents.create({ talentsString: '0053032024105103-321303010201132' }),
	{
		tooltip: 'Seal Fate, Relentless Strikes, Venom and Cold Blood in Assassination, with Puncturing Wounds and Dual Wield Specialization, for Backstab.',
		group: 'Synthetic',
	},
);
export const TalentsLevel60Mutilate = PresetUtils.makePresetTalents('Mutilate', SavedTalents.create({ talentsString: '00530320251451031-320303-005' }), {
	tooltip: 'Assassination for Mutilate: Seal Fate, Relentless Strikes, Venom, Cold Blood, Vile Poisons, with Puncturing Wounds and Opportunity.',
	group: 'Synthetic',
});
export const TalentsLevel60Subtlety = PresetUtils.makePresetTalents('Subtlety', SavedTalents.create({ talentsString: '00530301-320003-115003231320021151' }), {
	tooltip: 'Subtlety for Death Mark, Boot Knife, Hemorrhage, Premeditation and Preparation, with Blood Splatter in Assassination.',
	group: 'Synthetic',
});

export const TalentPresets = {
	[Phase.Phase1]: [
		TalentsLevel60CombatSwords,
		TalentsLevel60SealFateSwords,
		TalentsLevel60CombatDaggers,
		TalentsLevel60SealFateDaggers,
		TalentsLevel60Mutilate,
		TalentsLevel60Subtlety,
	],
	[Phase.Phase2]: [
		TalentsLevel60CombatSwords,
		TalentsLevel60SealFateSwords,
		TalentsLevel60CombatDaggers,
		TalentsLevel60SealFateDaggers,
		TalentsLevel60Mutilate,
		TalentsLevel60Subtlety,
	],
};

export const DefaultTalentsAssassin = TalentsLevel60Mutilate;
export const DefaultTalentsCombat = TalentsLevel60CombatSwords;
export const DefaultTalentsSubtlety = TalentsLevel60Subtlety;

export const DefaultTalents = TalentsLevel60CombatSwords;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// Deadly Poison on the main hand and Instant Poison on the off hand, the best pair for every spec. They
// are rogue options because Forever's poisons are class imbues beside the sharpening stones.
export const DefaultOptions = RogueOptions.create({
	mainHandPoison: WeaponImbue.DeadlyPoison,
	offHandPoison: WeaponImbue.InstantPoison,
});

///////////////////////////////////////////////////////////////////////////
//                         Consumes/Buffs/Debuffs
///////////////////////////////////////////////////////////////////////////

// The level 60 consumes the Level 60 builds were searched with (tools/rotopt/mksettings/spec_rogue60.go
// has the measurements): Elixir of the Mongoose (the one +2% crit elixir) and Elixir of Ferocity, Juju
// Might, Ground Scorpok Assay, Flank au Poivre, Dragonbreath Chili, Major Frenzy Potion, Thistle Tea,
// the Goblin Sapper Charge and Goblin Land Mine (Engineering), Deadly Poison on the main hand and
// Instant Poison on the off hand (ahead of every other pair for every spec, by 3 to 44). The caster
// flask and elixir are left out, though Dragonbreath Chili makes them worth 8 DPS.
export const P1Consumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredRogueThistleTea,
	defaultPotion: Potions.MajorFrenzyPotion,
	dragonBreathChili: true,
	fillerExplosive: Explosive.ExplosiveGoblinLandMine,
	// The natural flask (standards 11, the user's decision 2026-10-23: the fight is in Hyjal Summit or the
	// Barrow Deeps): Precision, +28 to +44 over none at 180 sec. On the 8% sets Accuracy is 0.3 to 2.2
	// behind it (a tie within 0.4 on Subtlety; Combat Daggers takes it, P1CombatDaggersConsumes),
	// Swiftness 13 to 26 and Aggression 25 to 31 behind.
	flask: Flask.FlaskOfNaturalPrecision,
	food: Food.FoodFlankAuPoivre,
	mainHandImbue: WeaponImbue.DenseSharpeningStone,
	offHandImbue: WeaponImbue.DenseSharpeningStone,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
});

export const DefaultConsumes = {
	[Phase.Phase1]: P1Consumes,
};

// Explosives are taken only when they gain (standards 11, the user's ruling 2026-10-23). The Goblin
// Sapper Charge gains for every build (+3.7 to +4.1 at 180 sec with the Land Mine out). The Land Mine's
// 1 sec cast resets the swing, and it loses for Combat Swords (1.3 to 2.9 at 120 / 180 / 300 sec on both
// seed sets), Combat Daggers (0.3 to 1.8) and Seal Fate Swords (-0.1 to 1.3); for Mutilate it is a tie
// (-0.3 to +0.3), so it stays out. Seal Fate Daggers (+0.1 to +0.7 at 120 and 180 sec, 0.3 to 0.4 behind
// at 300) and Subtlety (+0.6 to +1.1) keep it, with P1Consumes.
export const P1SapperOnlyConsumes = Consumes.create({
	...P1Consumes,
	fillerExplosive: Explosive.ExplosiveUnknown,
});

// Combat Daggers drinks the Flask of Natural Accuracy: its 5% hit does more for the white hits (the
// dual wield cap is 27%) than Precision's expertise, on the 9% set too (+0.1 to +0.6 at 120 / 180 / 300
// sec on five of the six seed set and length pairs, 0.4 behind on the sixth; 770.3 against 769.7 at 180 sec).
// The Combat builds hold a mace in the off hand, which takes a Dense Weightstone: a sharpening stone
// goes on an edged weapon only (the stones' SpellEquippedItems), and the weightstone adds the same 8.
export const P1CombatDaggersConsumes = Consumes.create({
	...P1SapperOnlyConsumes,
	flask: Flask.FlaskOfNaturalAccuracy,
	offHandImbue: WeaponImbue.DenseWeightstone,
});

export const P1CombatSwordsConsumes = Consumes.create({
	...P1SapperOnlyConsumes,
	offHandImbue: WeaponImbue.DenseWeightstone,
});

// The shared level 60 raid (both blessings, every debuff but Curse of Shadow), with the rogue in a melee
// group: Battle Shout, Strength of Earth and Windfury Totem, and Grace of Air from a second shaman in the raid as well (the user's ruling 2026-10-16, "both totems": client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide).
export const P1RaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	graceOfAirTotem: TristateEffect.TristateEffectRegular,
	leaderOfThePack: true,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
	thorns: TristateEffect.TristateEffectImproved,
	totemWeaponBuff: TotemWeaponBuff.TotemWeaponBuffWindfury,
	trueshotAura: true,
});

export const DefaultRaidBuffs = {
	[Phase.Phase1]: P1RaidBuffs,
};

export const P1IndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
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
	// The raid's Sunder Armor builds up from the pull (standard 32): a warrior tank stacks it and a DPS warrior casts one on its first GCD.
	sunderRamp: SunderRamp.create({ tank: true, pullSunders: 1 }),
	thunderClap: TristateEffect.TristateEffectImproved,
});

export const DefaultDebuffs = {
	[Phase.Phase1]: P1DefaultDebuffs,
};

export const P1OtherDefaults = {
	distanceFromTarget: 5,
	channelClipDelay: 50,
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
	race: Race.RaceUndead,
};

export const OtherDefaults = {
	[Phase.Phase1]: P1OtherDefaults,
};

///////////////////////////////////////////////////////////////////////////
//                                 Level 60
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the generic boss (level 63, 3731 armor, no creature type), a 3 minute
// fight, and the default buffs, debuffs and consumes above.
const level60Encounter = (name: string, mobType: MobType, consumes = P1Consumes) =>
	PresetUtils.makePresetEncounter(
		name,
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
					mobType,
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
			raidBuffs: P1RaidBuffs,
			debuffs: P1DefaultDebuffs,
			buffs: P1IndividualBuffs,
			consumes,
		},
	);
export const EncounterLevel60 = level60Encounter('Generic boss', MobType.MobTypeUnknown);
export const EncounterLevel60SapperOnly = level60Encounter('Mutilate', MobType.MobTypeUnknown, P1SapperOnlyConsumes);
export const EncounterLevel60CombatSwords = level60Encounter('Combat Swords', MobType.MobTypeUnknown, P1CombatSwordsConsumes);
export const EncounterLevel60CombatDaggers = level60Encounter('Combat Daggers', MobType.MobTypeUnknown, P1CombatDaggersConsumes);

// Round 70753 (November 2 development notes; the rework isn't in client 70753, so the notes' values,
// standards 37): Shank, Poised, the reworked tree and Death Mark, talents and rotation searched again,
// races swept with each race's own rotation (A2's base stats in). 20000 iterations, seed 1 (seed 1000003
// within 0.6), 120 / 180 / 300 sec, before (master 5eb381750) and after:
//   Combat Swords       808.7 / 780.4 / 759.5  ->  995.0 / 963.5 / 937.0 Undead (the default)
//   Seal Fate Swords    757.4 / 745.5 / 735.0  ->  975.9 / 957.0 / 941.0 Undead
//   Combat Daggers      814.1 / 789.6 / 769.6  ->  972.8 / 946.7 / 923.3 Undead
//   Seal Fate Daggers   765.1 / 751.9 / 742.6  ->  936.9 / 920.7 / 906.4 Undead
//   Subtlety            746.8 / 731.8 / 722.3  ->  896.5 / 881.7 / 868.3 Night Elf (Undead 893.4 / 880.2 / 869.0)
//   Mutilate            754.6 / 742.7 / 734.4  ->  881.0 / 866.1 / 855.3 Undead
// Races at 180 sec: the Undead leads five builds (0.0 to 4.2 ahead of the Orc or the Gnome; Combat Swords
// ties the Orc and stays), the Night Elf Subtlety (+1.4 and +1.9 over the Undead on both seed sets, its
// Shadowmeld into Ambush). Best Horde race: the Undead everywhere. Best Alliance race: the Gnome (Eureka! on
// cooldown for Seal Fate Swords and Mutilate, tools/dps_report/req/races/) and the Night Elf on Subtlety.
// The six "(vs Humanoid)" builds (Murder 2/2 against a Humanoid boss) were dropped in round 70753 (user
// ruling 2026-11-03): with Murder removed they had their base builds' talents and simmed the same. orchestration/round_70753/rogue.md has the ledger, the searches and the audit.
// The Level 60 builds: the synthetic Phase 1 gear, the searched talents and rotation, the Level 60
// encounter and the best race at 180 sec. A build changes race only when the new one leads by more
// than 0.5 DPS on two seed sets (standards 16). On 2026-10-24 (the hit each set sims best with, 8% to
// 10%, and explosives only where they gain, standards 4 and 11; the natural flasks, the wowsims/forever
// alignment round and Shadowmeld since 2026-10-23), 20000 iterations, seed 1
// (seed 1000003 within 0.7), at 120 / 180 / 300 sec, against the generic boss (no creature type), on the
// synthetic sets built to the Forever budget (standards 1b: armor 650.5, weapons 63.9, bow 19.5
// weighted linear points a set, plus two synthetic ilvl 71 trinkets, 31 Agility and 15 crit rating, and
// +335 attack power for 20 sec every 2 min), with a Dense Sharpening Stone beside the poison on each blade
// (a Dense Weightstone on the Combat builds' off-hand mace since 2026-10-30, the same +8):
//   Combat Daggers      793.8 / 770.3 / 750.5 Skyborne (Windshaper) (the default), 9% hit
//   Combat Swords       789.9 / 764.0 / 743.8 Skyborne (Windshaper), 10% hit
//   Seal Fate Daggers   746.3 / 731.5 / 720.9 Orc, 5.0% behind Combat Daggers, 10% hit
//   Seal Fate Swords    742.1 / 728.9 / 717.0 Orc, 4.6% behind Combat Swords, 10% hit
//   Mutilate            730.4 / 718.9 / 710.4 Skyborne (Windshaper), 12.6 behind Seal Fate Daggers, 10% hit
//   Subtlety            735.8 / 715.9 / 702.6 Night Elf, 8% hit
// Before (2026-10-23, every set at 8% and every build with the Land Mine), at 180 sec: Combat Daggers
// 767.8, Combat Swords 761.3, Seal Fate Daggers 730.6, Seal Fate Swords 726.9, Mutilate 717.3, Subtlety
// 714.5. The races were swept on the 8% sets.
// The Windshaper's Wind Blessed haste leads the Combat builds (0.5 ahead of the High Order on Combat
// Daggers, 0.1 ahead of the Orc on Combat Swords) and stays on Mutilate (the Orc is 0.3 ahead on one seed
// set, 0.2 behind on the other). The Orc's Savage takes both Seal Fate builds (0.6 to 1.4 ahead of the
// Windshaper), and the Night Elf's Shadowmeld into Ambush takes Subtlety (0.9 to 1.1 ahead of the Orc).
// The best Alliance race is the High Order on the Combat builds (0.2 to 0.5 behind the Windshaper), the
// Human on Mutilate and both Seal Fate builds (0.1 to 0.4 behind) and the Night Elf on Subtlety. No
// racial depends on the weapon since client 70425.
// Race re-pick (2026-10-27, standards 29 and 16, on the 10-26 racial pass with each race's own rotation
// and gear): Combat Daggers and Seal Fate Daggers move to the Gnome with Eureka! on cooldown, Mutilate and
// Seal Fate Swords to the Undead. At 180 sec (verify.py, 20000 iterations, seed 1): Combat Daggers 770.5
// against the Windshaper's 765.6, Seal Fate Daggers 730.6 against the Orc's 725.3, Mutilate 717.3 against
// the Windshaper's 714.9, Seal Fate Swords 726.1 against the Orc's 721.9. Their vs Humanoid builds follow.
// Combat Swords stays Windshaper and Subtlety Night Elf.
// Second race re-pick (2026-10-27, standards 29 and 16, after dev notes post 8: Orc Blood Fury 2 min, Troll
// Berserking 15 sec, Undead Touch of the Grave 15 sec): Combat Daggers, Seal Fate Daggers, Combat Swords and
// Subtlety move to the Undead, so every Level 60 rogue is Undead. At 180 sec (verify.py, 20000 iterations,
// seed 1): Combat Daggers 773.3 against the Gnome's 770.5, Seal Fate Daggers 734.9 against the Gnome's
// 730.6, Combat Swords 765.4 against the Windshaper's 759.1, Subtlety 717.2 against the Night Elf's 714.9.
// Their vs Humanoid builds follow.
// Rotation re-search (2026-11-02; an extra attack keeps the swing timer since 2026-10-30, so the builders
// no longer wait for a main hand swing): rotopt at 120 / 180 / 300 sec, 20000 iterations, seed 1 (seed
// 1000003 within 0.8 of the same gain), before and after:
//   Combat Daggers      813.2 / 788.9 / 768.8  ->  815.3 / 790.8 / 770.7 (also no Vanish, 1 sec Slice and Dice refresh)
//   Combat Swords       804.3 / 777.0 / 755.9  ->  810.5 / 782.1 / 761.2
//   Seal Fate Daggers   764.4 / 751.4 / 742.0  ->  766.1 / 752.9 / 743.7
//   Seal Fate Swords    755.0 / 743.2 / 732.6  ->  759.0 / 747.1 / 736.6
//   Mutilate            754.7 / 742.7 / 734.4  ->  755.7 / 743.8 / 735.5
//   Subtlety            747.5 / 732.1 / 722.3  ->  747.9 / 732.9 / 723.4
// Every build stays Undead (180 sec, 20000 iterations, seed 1, each race with its own rotation): 2.7 to
// 6.2 ahead of the next race (the Orc, the Gnome or the Night Elf); the best Alliance race is the Gnome (the Night Elf on
// Subtlety).
// The weapons have the type and speed each build sims best with among Era Phase 1 BiS speeds
// (standards 22), at the Era BiS weapons' damage per second. A slow main hand and a fast off hand win:
// Combat Swords holds a 2.9 sec axe and a 2.5 sec mace, Combat Daggers a 1.8 sec dagger and the same
// mace (since Windfury Totem procs from the main hand only, a slow off hand wins for the Combat builds,
// and a mace takes Hack and Slash's armor ignore), Seal Fate Swords the 2.9 sec axe and a 1.6 sec
// dagger, and Mutilate a 1.7 sec dagger and a 1.6 sec dagger. Seal Fate Daggers and Subtlety keep
// Perdition's Blade and Core Hound Tooth.
// The "(vs Humanoid)" builds (Murder 2/2, Humanoid boss, the same race as their build) at 180 sec:
// Combat Daggers 794.0, Combat Swords 790.4, Seal Fate Daggers 755.8, Seal Fate Swords 754.6, Mutilate
// 743.8, Subtlety 741.4.
// Hit over the 8% yellow cap still helps the white hits (the dual wield cap is 27%), and dual wielders
// take the hit that sims best (standards 4, 2026-10-23): tools/database/forever_synthetic_items_rogue.go
// has the 8% / 9% / 10% sims. The Land Mine is out of Combat Swords, Combat Daggers, Seal Fate Swords and
// Mutilate, where it doesn't gain (P1SapperOnlyConsumes).
// Venom (25 Energy) loses DPS at every point count as modelled, about 30 at every length, so there is
// no Venom build. Grimstitch Armor's 4 and 5 piece bonuses would add about 8 and 6 to Combat Daggers.
// The trinkets are synthetic (Forever's real Phase 1 trinkets are removed or not simulated), and the
// fight does not open from Stealth (a raid rogue rarely does), so Subtlety gets Stealth from Vanish.
// forever-wiki/sim-status.md has the races, the searches, the audit and the talent by talent
// verification.
const level60Build = (
	name: string,
	tooltip: string,
	gear: PresetUtils.PresetGear,
	talents: PresetUtils.PresetTalents,
	rotation: PresetUtils.PresetRotation,
	encounter = EncounterLevel60,
	race = Race.RaceUndead,
) =>
	PresetUtils.makePresetBuild(name, {
		group: 'Synthetic',
		tooltip,
		gear,
		talents,
		rotation,
		encounter,
		race,
		level: 60,
		bonusTalentPoints: 0,
		professions: [P1OtherDefaults.profession1, P1OtherDefaults.profession2],
		options: DefaultOptions,
	});
export const PresetBuildLevel60CombatDaggers = level60Build(
	'Combat Daggers',
	'Level 60 Undead Combat Backstab on the synthetic Phase 1 dagger and a slow mace off hand.',
	GearLevel60SyntheticCombatDaggersHitPlus2,
	TalentsLevel60CombatDaggers,
	ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
	EncounterLevel60CombatDaggers,
	Race.RaceUndead,
);
export const PresetBuildLevel60SealFateDaggers = level60Build(
	'Seal Fate Daggers',
	'Level 60 Undead Seal Fate Backstab on the synthetic Phase 1 daggers.',
	GearLevel60SyntheticDaggersHitPlus3,
	TalentsLevel60SealFateDaggers,
	ROTATION_PRESET_LEVEL60_SEAL_FATE_DAGGERS,
	EncounterLevel60,
	Race.RaceUndead,
);
export const PresetBuildLevel60CombatSwords = level60Build(
	'Combat Swords',
	'Level 60 Undead Combat Sinister Strike on a slow synthetic Phase 1 axe and a mace off hand, in a 3 minute raid fight.',
	GearLevel60SyntheticSwordsHitPlus2,
	TalentsLevel60CombatSwords,
	ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
	EncounterLevel60CombatSwords,
	Race.RaceUndead,
);
export const PresetBuildLevel60SealFateSwords = level60Build(
	'Seal Fate Swords',
	'Level 60 Undead Seal Fate Sinister Strike on the synthetic Phase 1 axe and a fast dagger off hand.',
	GearLevel60SyntheticSealFateSwordsHitPlus3,
	TalentsLevel60SealFateSwords,
	ROTATION_PRESET_LEVEL60_SEAL_FATE_SWORDS,
	EncounterLevel60SapperOnly,
	Race.RaceUndead,
);
export const PresetBuildLevel60Mutilate = level60Build(
	'Mutilate',
	'Level 60 Undead Assassination Mutilate on two synthetic Phase 1 daggers.',
	GearLevel60SyntheticMutilateHitPlus3,
	TalentsLevel60Mutilate,
	ROTATION_PRESET_LEVEL60_MUTILATE,
	EncounterLevel60SapperOnly,
	Race.RaceUndead,
);
export const PresetBuildLevel60Subtlety = level60Build(
	'Subtlety',
	'Level 60 Night Elf Subtlety (Death Mark, Boot Knife, Hemorrhage, Premeditation, Vanish and Shadowmeld into Ambush) on the synthetic Phase 1 daggers.',
	GearLevel60SyntheticSubtletyHitPlus2,
	TalentsLevel60Subtlety,
	ROTATION_PRESET_LEVEL60_SUBTLETY,
	EncounterLevel60,
	Race.RaceNightElf,
);
