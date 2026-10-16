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
	TotemWeaponBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { RogueOptions } from '../core/proto/rogue.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import BackstabAPL from './apls/combat_backstab.apl.json';
import BackstabOnCooldownAPL from './apls/combat_backstab_on_cooldown.apl.json';
import BackstabRuptureAPL from './apls/combat_backstab_rupture.apl.json';
import BackstabSndWindowAPL from './apls/combat_backstab_snd_window.apl.json';
import BackstabSweatyAPL from './apls/combat_backstab_sweaty.apl.json';
import SinisterStrikeAPL from './apls/combat_sinister_strike.apl.json';
import SinisterStrikeHoldForARAPL from './apls/combat_sinister_strike_hold_for_ar.apl.json';
import SinisterStrikeOnCooldownAPL from './apls/combat_sinister_strike_on_cooldown.apl.json';
import SinisterStrikeSndOpenerAPL from './apls/combat_sinister_strike_snd_opener.apl.json';
import SinisterStrikeSndWindowAPL from './apls/combat_sinister_strike_snd_window.apl.json';
import SinisterStrikeSweatyAPL from './apls/combat_sinister_strike_sweaty.apl.json';
import MutilateAPL from './apls/forever_mutilate.apl.json';
import Level30p5BackstabAPL from './apls/level30p5_backstab.apl.json';
import Level30p5SinisterStrikeAPL from './apls/level30p5_sinister_strike.apl.json';
import Level60CombatDaggersAPL from './apls/level60_combat_daggers.apl.json';
import Level60CombatSwordsAPL from './apls/level60_combat_swords.apl.json';
import Level60MutilateAPL from './apls/level60_mutilate.apl.json';
import Level60SubtletyAPL from './apls/level60_subtlety.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import BackstabGearP1BiS from './gear_sets/combat_backstab_p1_bis.gear.json';
import BackstabGearP2BiS from './gear_sets/combat_backstab_p2_bis.gear.json';
import BackstabGearPreBiS from './gear_sets/combat_backstab_prebis.gear.json';
import SinisterStrikeGearP1BiS from './gear_sets/combat_sinister_strike_p1_bis.gear.json';
import SinisterStrikeGearP2BiS from './gear_sets/combat_sinister_strike_p2_bis.gear.json';
import SinisterStrikeGearPreBiS from './gear_sets/combat_sinister_strike_prebis.gear.json';
import Level30p5BackstabGear from './gear_sets/level30p5_backstab.gear.json';
import Level30p5SinisterStrikeGear from './gear_sets/level30p5_sinister_strike.gear.json';
import Level60SyntheticDaggersGear from './gear_sets/level60_phase1_synthetic_daggers.gear.json';
import Level60SyntheticCombatDaggersGear from './gear_sets/level60_phase1_synthetic_combat_daggers.gear.json';
import Level60SyntheticMutilateGear from './gear_sets/level60_phase1_synthetic_mutilate.gear.json';
import Level60SyntheticSealFateSwordsGear from './gear_sets/level60_phase1_synthetic_seal_fate_swords.gear.json';
import Level60SyntheticSwordsGear from './gear_sets/level60_phase1_synthetic_swords.gear.json';

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
// weapons. Both carry 5% hit, the 9% cap with Precision 3/3 and the hands and legs, which carry
// the 2 piece bonus of Grimstitch Armor (Forever's crafted rogue set, 1% hit). The 4 and 5 piece
// bonuses (36 attack power against Humanoids, Eviscerate 5 Energy cheaper) are not live, see
// PresetBuildLevel60CombatSwords. The trinkets are synthetic too (31 Agility and 15 crit rating, and
// +335 attack power for 20 sec every 2 min): Forever removed Hand of Justice and Blackhand's Breadth,
// and the sim does not model its launch trinkets yet.
//
// The enchants are the best non-raid ones: Lesser Arcanum of Voracity (Agility) on head and legs,
// Enchant Necklace - Agility, Cloak - Agility (+5), Chest - Greater Stats, Superior Strength on
// bracers (+9) and gloves (+15, 1.3 DPS ahead of Greater Agility for Backstab), Boots - Greater
// Agility and Crusader on both weapons (3 ahead of Weapon - Agility).
export const GearLevel60SyntheticSwords = PresetUtils.makePresetGear('Level 60 Synthetic Swords', Level60SyntheticSwordsGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Sinister Strike, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
export const GearLevel60SyntheticDaggers = PresetUtils.makePresetGear('Level 60 Synthetic Daggers', Level60SyntheticDaggersGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Seal Fate Backstab and Subtlety, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
// The swords set with a 1.6 sec dagger off hand, which Seal Fate sims best with.
export const GearLevel60SyntheticSealFateSwords = PresetUtils.makePresetGear('Level 60 Synthetic Seal Fate Swords', Level60SyntheticSealFateSwordsGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Seal Fate Sinister Strike, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
// The daggers set with a 2.5 sec mace off hand, which Combat Daggers sims best with since Windfury Totem
// procs from the main hand only (a mace takes Hack and Slash's armor ignore).
export const GearLevel60SyntheticCombatDaggers = PresetUtils.makePresetGear('Level 60 Synthetic Combat Daggers', Level60SyntheticCombatDaggersGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Combat Backstab, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
// The daggers set with a 1.7 sec dagger main hand and a 1.8 sec sword off hand, which Mutilate sims best
// with: its off hand strike is normalised to 2.4 sec for any weapon but a dagger.
export const GearLevel60SyntheticMutilate = PresetUtils.makePresetGear('Level 60 Synthetic Mutilate', Level60SyntheticMutilateGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever for Mutilate, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [
		GearLevel60SyntheticSwords,
		GearLevel60SyntheticDaggers,
		GearLevel60SyntheticSealFateSwords,
		GearLevel60SyntheticCombatDaggers,
		GearLevel60SyntheticMutilate,
		GearBackstabPreBiS,
		GearSinisterStrikePreBiS,
		GearBackstabP1BiS,
		GearSinisterStrikeP1BiS,
		GearLevel30p5Backstab,
		GearLevel30p5SinisterStrike,
	],
	[Phase.Phase2]: [
		GearLevel60SyntheticSwords,
		GearLevel60SyntheticDaggers,
		GearLevel60SyntheticSealFateSwords,
		GearLevel60SyntheticCombatDaggers,
		GearLevel60SyntheticMutilate,
		GearBackstabPreBiS,
		GearSinisterStrikePreBiS,
		GearBackstabP2BiS,
		GearSinisterStrikeP2BiS,
		GearLevel30p5Backstab,
		GearLevel30p5SinisterStrike,
	],
};

export const DefaultGear = GearLevel60SyntheticSwords;

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

// The Level 60 rotations, from the tools/rotopt searches on the synthetic Phase 1 sets with the
// talents below (knob search, talent search, knob search again, 180 sec, see rogue_presets.sh).
// Seal Fate Swords and Seal Fate Daggers use the Combat Swords and Combat Daggers rotations, the
// knob search kept every knob.
export const ROTATION_PRESET_LEVEL60_COMBAT_SWORDS = PresetUtils.makePresetAPLRotation('Level 60 Combat Swords', Level60CombatSwordsAPL, {
	tooltip: 'Sinister Strike, Slice and Dice and 5 point Eviscerates, Blade Flurry before the pull. Also the Seal Fate Swords rotation.',
	group: 'Level 60',
});
export const ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS = PresetUtils.makePresetAPLRotation('Level 60 Combat Daggers', Level60CombatDaggersAPL, {
	tooltip: 'Backstab, Slice and Dice and 5 point Eviscerates, Blade Flurry before the pull. Also the Seal Fate Daggers rotation.',
	group: 'Level 60',
});
export const ROTATION_PRESET_LEVEL60_MUTILATE = PresetUtils.makePresetAPLRotation('Level 60 Mutilate', Level60MutilateAPL, {
	tooltip: 'Mutilate, Slice and Dice and 5 point Eviscerates.',
	group: 'Level 60',
});
export const ROTATION_PRESET_LEVEL60_SUBTLETY = PresetUtils.makePresetAPLRotation('Level 60 Subtlety', Level60SubtletyAPL, {
	tooltip: 'Hemorrhage, Ghostly Strike on cooldown, a 3 point Rupture, Premeditation and Vanish into Ambush at 3 points or fewer.',
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [
		ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
		ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
		ROTATION_PRESET_LEVEL60_MUTILATE,
		ROTATION_PRESET_LEVEL60_SUBTLETY,
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
		ROTATION_PRESET_MUTILATE,
		ROTATION_PRESET_LEVEL30P5_BACKSTAB,
		ROTATION_PRESET_LEVEL30P5_SINISTER_STRIKE,
	],
	[Phase.Phase2]: [
		ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
		ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
		ROTATION_PRESET_LEVEL60_MUTILATE,
		ROTATION_PRESET_LEVEL60_SUBTLETY,
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
		ROTATION_PRESET_MUTILATE,
		ROTATION_PRESET_LEVEL30P5_BACKSTAB,
		ROTATION_PRESET_LEVEL30P5_SINISTER_STRIKE,
	],
};

//Need to add main hand equip logic or talent/rotation logic to map to Auto APL
export const DefaultAPLs: Record<number, PresetUtils.PresetRotation> = {
	[0]: ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
	[1]: ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
};

export const DefaultAPLBackstab = ROTATION_PRESET_BACKSTAB;
export const DefaultAPLSinisterStrike = ROTATION_PRESET_SINISTER_STRIKE;
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
	SavedTalents.create({ talentsString: '00530310501-32003311201515231' }),
	{ tooltip: 'Level 60 Combat talents for Sinister Strike.', group: 'Level 60' },
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

// The Level 60 talents, from the tools/rotopt talent searches on the synthetic Phase 1 sets (180 sec,
// 3000 iterations a candidate, the best moves checked again at 10000), seeded from each archetype
// and alternated with the knob search until neither moved. See PresetBuildLevel60CombatDaggers for
// the numbers. These are for the generic boss (no creature type), where Murder does nothing, so its
// two points go to Lethality (Vile Poisons for Seal Fate Daggers, whose Lethality is full). The
// "(vs Humanoid)" talents below keep Murder 2/2 for Humanoid and Giant bosses.
export const TalentsLevel60CombatSwords = PresetUtils.makePresetTalents(
	'Level 60 Combat Swords',
	SavedTalents.create({ talentsString: '00530310501-32003311201515231' }),
	{ tooltip: 'Deep Combat for Sinister Strike: Adrenaline Rush, Blade Flurry, Hack and Slash, Cold Blood.', group: 'Level 60' },
);
export const TalentsLevel60SealFateSwords = PresetUtils.makePresetTalents(
	'Level 60 Seal Fate Swords',
	SavedTalents.create({ talentsString: '0053031035140005-3220132020141' }),
	{ tooltip: 'Seal Fate and Cold Blood in Assassination with Blade Flurry and Hack and Slash, for Sinister Strike.', group: 'Level 60' },
);
export const TalentsLevel60CombatDaggers = PresetUtils.makePresetTalents(
	'Level 60 Combat Daggers',
	SavedTalents.create({ talentsString: '005303104-30230320201515231-002' }),
	{ tooltip: 'Deep Combat for Backstab: Adrenaline Rush, Blade Flurry, Puncturing Wounds, Hack and Slash, Opportunity.', group: 'Level 60' },
);
export const TalentsLevel60SealFateDaggers = PresetUtils.makePresetTalents(
	'Level 60 Seal Fate Daggers',
	SavedTalents.create({ talentsString: '0053031055120005-322303200013-002' }),
	{ tooltip: 'Seal Fate and Cold Blood in Assassination with Puncturing Wounds and Blade Flurry, for Backstab.', group: 'Level 60' },
);
export const TalentsLevel60Mutilate = PresetUtils.makePresetTalents(
	'Level 60 Mutilate',
	SavedTalents.create({ talentsString: '0053031033150103-302303202015-002' }),
	{
		tooltip: 'Assassination for Mutilate: Seal Fate, Cold Blood, Vile Poisons, with Precision and Puncturing Wounds in Combat.',
		group: 'Level 60',
	},
);
export const TalentsLevel60Subtlety = PresetUtils.makePresetTalents(
	'Level 60 Subtlety',
	SavedTalents.create({ talentsString: '0053031035140004--3021203310113011' }),
	{ tooltip: 'Subtlety for Hemorrhage, Ghostly Strike, Premeditation and Preparation, with Seal Fate in Assassination.', group: 'Level 60' },
);

// The same builds for a Humanoid or Giant boss, where Murder 2/2 (+4% damage) is worth more than the
// points it takes from Lethality or Vile Poisons. The talent searches on a Humanoid boss chose these.
const humanoidTalents = (name: string, talentsString: string, tooltip: string) =>
	PresetUtils.makePresetTalents(`${name} (vs Humanoid)`, SavedTalents.create({ talentsString }), {
		tooltip: `${tooltip} With Murder 2/2, for a Humanoid or Giant boss.`,
		group: 'Level 60',
	});
export const TalentsLevel60CombatSwordsHumanoid = humanoidTalents(
	'Level 60 Combat Swords',
	'00532310301-32003311201515231',
	'Deep Combat for Sinister Strike.',
);
export const TalentsLevel60SealFateSwordsHumanoid = humanoidTalents(
	'Level 60 Seal Fate Swords',
	'0053231015140005-3220132020141',
	'Seal Fate with Blade Flurry for Sinister Strike.',
);
export const TalentsLevel60CombatDaggersHumanoid = humanoidTalents('Level 60 Combat Daggers', '005323102-30230320201515231-002', 'Deep Combat for Backstab.');
export const TalentsLevel60SealFateDaggersHumanoid = humanoidTalents(
	'Level 60 Seal Fate Daggers',
	'0053231053120005-322303200013-002',
	'Seal Fate with Puncturing Wounds for Backstab.',
);
export const TalentsLevel60MutilateHumanoid = humanoidTalents('Level 60 Mutilate', '0053231013150103-302303202015-002', 'Assassination for Mutilate.');
export const TalentsLevel60SubtletyHumanoid = humanoidTalents('Level 60 Subtlety', '0053231015140004--3021203310113011', 'Subtlety for Hemorrhage.');

export const TalentPresets = {
	[Phase.Phase1]: [
		TalentsLevel60CombatSwords,
		TalentsLevel60SealFateSwords,
		TalentsLevel60CombatDaggers,
		TalentsLevel60SealFateDaggers,
		TalentsLevel60Mutilate,
		TalentsLevel60Subtlety,
		TalentsLevel60CombatDaggersHumanoid,
		TalentsLevel60SealFateDaggersHumanoid,
		TalentsLevel60CombatSwordsHumanoid,
		TalentsLevel60SealFateSwordsHumanoid,
		TalentsLevel60MutilateHumanoid,
		TalentsLevel60SubtletyHumanoid,
		CombatBackstabTalents,
		CombatSinisterStrikeTalents,
		AssassinationMutilateTalents,
		Level30p5BackstabTalents,
		Level30p5SinisterStrikeTalents,
	],
	[Phase.Phase2]: [
		TalentsLevel60CombatSwords,
		TalentsLevel60SealFateSwords,
		TalentsLevel60CombatDaggers,
		TalentsLevel60SealFateDaggers,
		TalentsLevel60Mutilate,
		TalentsLevel60Subtlety,
		TalentsLevel60CombatDaggersHumanoid,
		TalentsLevel60SealFateDaggersHumanoid,
		TalentsLevel60CombatSwordsHumanoid,
		TalentsLevel60SealFateSwordsHumanoid,
		TalentsLevel60MutilateHumanoid,
		TalentsLevel60SubtletyHumanoid,
		CombatBackstabTalents,
		CombatSinisterStrikeTalents,
		AssassinationMutilateTalents,
		Level30p5BackstabTalents,
		Level30p5SinisterStrikeTalents,
	],
};

export const DefaultTalentsAssassin = TalentsLevel60Mutilate;
export const DefaultTalentsCombat = TalentsLevel60CombatSwords;
export const DefaultTalentsSubtlety = TalentsLevel60Subtlety;

export const DefaultTalentsBackstab = CombatBackstabTalents;
export const DefaultTalentsSinisterStrike = CombatSinisterStrikeTalents;
export const DefaultTalentsMutilate = AssassinationMutilateTalents;

export const DefaultTalents = TalentsLevel60CombatSwords;

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
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

export const DefaultDebuffs = {
	[Phase.Phase1]: P1DefaultDebuffs,
};

export const P1OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Leatherworking,
	race: Race.RaceOrc,
};

export const OtherDefaults = {
	[Phase.Phase1]: P1OtherDefaults,
};

///////////////////////////////////////////////////////////////////////////
//                                 Level 60
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the generic boss (level 63, 3731 armor, no creature type), a 3 minute
// fight, and the default buffs, debuffs and consumes above. It is here so switching back from the
// Level 30 builds restores everything they changed.
const level60Encounter = (name: string, mobType: MobType) =>
	PresetUtils.makePresetEncounter(
		name,
		Encounter.create({
			duration: 180,
			durationVariation: 20,
			executeProportion20: 0.2,
			executeProportion25: 0.25,
			executeProportion35: 0.35,
			...PresetUtils.LEVEL_60_RAID_DAMAGE,
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
			consumes: P1Consumes,
		},
	);
export const EncounterLevel60 = level60Encounter('Level 60', MobType.MobTypeUnknown);
export const EncounterLevel60Humanoid = level60Encounter('Level 60 (vs Humanoid)', MobType.MobTypeHumanoid);

// The Level 60 builds: the synthetic Phase 1 gear, the searched talents and rotation, the Level 60
// encounter and the best race at 180 sec. A build changes race only when the new one leads by more
// than 0.5 DPS on two seed sets (standards 16). On 2026-10-16 (client 70425 racials, raid hits),
// 20000 iterations, seed 1, at 120 / 180 / 300 sec, against the generic boss (no creature type), on
// the synthetic sets built to the Forever budget (standards 1b: armor 650.5, weapons 63.9, bow 19.5
// weighted linear points a set, plus two synthetic ilvl 71 trinkets, 31 Agility and 15 crit rating,
// and +335 attack power for 20 sec every 2 min), with a Dense Sharpening Stone beside the poison on
// each blade:
//   Combat Swords       730.3 / 704.7 / 685.5 Orc (the default)
//   Combat Daggers      720.8 / 698.7 / 680.2 Undead
//   Seal Fate Daggers   676.8 / 665.5 / 657.0 Undead, 4.8% behind Combat Daggers
//   Mutilate            680.6 / 669.9 / 661.7 Undead
//   Subtlety            648.1 / 634.3 / 624.5 Undead
//   Seal Fate Swords    676.6 / 666.5 / 657.0 Skyborne (Windshaper), 5.4% behind Combat Swords
// The Orc leads the Undead by 3.0 to 3.4 on Combat Swords (Savage), the Windshaper by 1.0 to 1.4 on
// Seal Fate Swords (Wind Blessed). On the dagger builds the Windshaper, High Order, Undead and Orc are
// within 0.7 and the Undead stays. The best Alliance race is the Human on Combat Swords and the High
// Order everywhere else. No racial depends on the weapon since client 70425.
// The weapons have the type and speed each build sims best with among Era Phase 1 BiS speeds
// (standards 22), at the Era BiS weapons' damage per second. A slow main hand and a fast off hand win:
// Combat Swords holds a 2.9 sec axe and a 1.7 sec axe, Seal Fate Swords the 2.9 sec axe and a 1.6 sec
// dagger, and Mutilate a 1.7 sec dagger and a 1.8 sec sword. The other dagger builds keep Perdition's
// Blade and Core Hound Tooth.
// The "(vs Humanoid)" builds (Murder 2/2, Humanoid boss, the same race as their build) at 180 sec:
// Combat Swords 729.2, Combat Daggers 720.4, Seal Fate Daggers 688.6, Mutilate 692.8, Subtlety 656.9,
// Seal Fate Swords 690.0.
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
		group: 'Level 60',
		tooltip,
		gear,
		talents,
		rotation,
		encounter,
		race,
		level: 60,
		bonusTalentPoints: 0,
		professions: [P1OtherDefaults.profession1, P1OtherDefaults.profession2],
	});
export const PresetBuildLevel60CombatDaggers = level60Build(
	'Level 60 Combat Daggers',
	'Level 60 Undead Combat Backstab on the synthetic Phase 1 dagger and a slow mace off hand.',
	GearLevel60SyntheticCombatDaggers,
	TalentsLevel60CombatDaggers,
	ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
);
export const PresetBuildLevel60SealFateDaggers = level60Build(
	'Level 60 Seal Fate Daggers',
	'Level 60 Undead Seal Fate Backstab on the synthetic Phase 1 daggers.',
	GearLevel60SyntheticDaggers,
	TalentsLevel60SealFateDaggers,
	ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
);
export const PresetBuildLevel60CombatSwords = level60Build(
	'Level 60 Combat Swords',
	'Level 60 Orc Combat Sinister Strike on a slow synthetic Phase 1 axe and a mace off hand, in a 3 minute raid fight.',
	GearLevel60SyntheticSwords,
	TalentsLevel60CombatSwords,
	ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
	EncounterLevel60,
	Race.RaceOrc,
);
export const PresetBuildLevel60SealFateSwords = level60Build(
	'Level 60 Seal Fate Swords',
	'Level 60 Windshaper Seal Fate Sinister Strike on the synthetic Phase 1 axe and a fast dagger off hand.',
	GearLevel60SyntheticSealFateSwords,
	TalentsLevel60SealFateSwords,
	ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
	EncounterLevel60,
	Race.RaceSkyborneWindshaper,
);
export const PresetBuildLevel60Mutilate = level60Build(
	'Level 60 Mutilate',
	'Level 60 Undead Assassination Mutilate on the synthetic Phase 1 dagger and a fast sword off hand.',
	GearLevel60SyntheticMutilate,
	TalentsLevel60Mutilate,
	ROTATION_PRESET_LEVEL60_MUTILATE,
);
export const PresetBuildLevel60Subtlety = level60Build(
	'Level 60 Subtlety',
	'Level 60 Undead Subtlety (Hemorrhage, Ghostly Strike, Premeditation) on the synthetic Phase 1 daggers.',
	GearLevel60SyntheticDaggers,
	TalentsLevel60Subtlety,
	ROTATION_PRESET_LEVEL60_SUBTLETY,
);

// The vs Humanoid builds: the same specs against a Humanoid boss (Giants too), with Murder 2/2.
// Extra presets, the generic boss builds above are the defaults.
const humanoidBuild = (
	name: string,
	gear: PresetUtils.PresetGear,
	talents: PresetUtils.PresetTalents,
	rotation: PresetUtils.PresetRotation,
	race = Race.RaceUndead,
) =>
	level60Build(
		`${name} (vs Humanoid)`,
		`${name} against a Humanoid or Giant boss, with Murder 2/2.`,
		gear,
		talents,
		rotation,
		EncounterLevel60Humanoid,
		race,
	);
export const PresetBuildLevel60CombatDaggersHumanoid = humanoidBuild(
	'Level 60 Combat Daggers',
	GearLevel60SyntheticCombatDaggers,
	TalentsLevel60CombatDaggersHumanoid,
	ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
);
export const PresetBuildLevel60SealFateDaggersHumanoid = humanoidBuild(
	'Level 60 Seal Fate Daggers',
	GearLevel60SyntheticDaggers,
	TalentsLevel60SealFateDaggersHumanoid,
	ROTATION_PRESET_LEVEL60_COMBAT_DAGGERS,
);
export const PresetBuildLevel60CombatSwordsHumanoid = humanoidBuild(
	'Level 60 Combat Swords',
	GearLevel60SyntheticSwords,
	TalentsLevel60CombatSwordsHumanoid,
	ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
	Race.RaceOrc,
);
export const PresetBuildLevel60SealFateSwordsHumanoid = humanoidBuild(
	'Level 60 Seal Fate Swords',
	GearLevel60SyntheticSealFateSwords,
	TalentsLevel60SealFateSwordsHumanoid,
	ROTATION_PRESET_LEVEL60_COMBAT_SWORDS,
	Race.RaceSkyborneWindshaper,
);
export const PresetBuildLevel60MutilateHumanoid = humanoidBuild(
	'Level 60 Mutilate',
	GearLevel60SyntheticMutilate,
	TalentsLevel60MutilateHumanoid,
	ROTATION_PRESET_LEVEL60_MUTILATE,
);
export const PresetBuildLevel60SubtletyHumanoid = humanoidBuild(
	'Level 60 Subtlety',
	GearLevel60SyntheticDaggers,
	TalentsLevel60SubtletyHumanoid,
	ROTATION_PRESET_LEVEL60_SUBTLETY,
);
