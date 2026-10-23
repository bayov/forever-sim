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
	IntellectElixir,
	ManaRegenElixir,
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
import BeastMasteryMarkAPL from './apls/bm_mark.apl.json';
import BeastMasteryScorpidAPL from './apls/bm_scorpid.apl.json';
import BeastMasteryTurretAPL from './apls/bm_turret.apl.json';
import LoneWolfAPL from './apls/lw.apl.json';
import LoneWolfMarkAPL from './apls/lw_mark.apl.json';
import LoneWolfScorpidAPL from './apls/lw_scorpid.apl.json';
import LoneWolfTurretAPL from './apls/lw_turret.apl.json';
import MarksmanshipAPL from './apls/mm.apl.json';
import MarksmanshipMarkAPL from './apls/mm_mark.apl.json';
import MarksmanshipScorpidAPL from './apls/mm_scorpid.apl.json';
import MarksmanshipTurretAPL from './apls/mm_turret.apl.json';
import SurvivalAPL from './apls/sv.apl.json';
import SurvivalMarkAPL from './apls/sv_mark.apl.json';
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
// Forever data. Each piece takes the place of an Era Phase 1 hunter BiS piece (the Giantstalker
// set, Onyxia Tooth Pendant, Wristguards of True Flight, the Devilsaur hands and legs, Quick Strike
// Ring, Barbarous Blade, Rhok'delar) on the cross-class item level table (66 for the six armor
// pieces, 71 for neck, back, wrist, waist and rings, 76 for the two-hander, 71 for the ranged weapon) and the
// cross-class budget 1b (orchestration/standards.md, 2026-10-13): each armor piece's weighted stats
// are slot mod x (1.125 ilvl + 2.0) x the hunter's 1.028 / 1.056 / 0.977 per tier, on 4 / 4 / 3
// stat lines. The file has the math. Ranged set totals: armor 704.0 (target 706.1), two-hander
// 85.6, ranged weapon 19.8. Beast Mastery, Marksmanship and Lone Wolf share the Ranged set: they shoot
// on the same stats, and the pet takes 10% of the hunter's ranged attack power and all of its crit.
// The ranged weapon keeps Rhok'delar's damage per second as a 2.5 sec bow, Striker's Mark's speed and
// spread, the Phase 1 BiS speed the three ranged specs sim best with (standards 22, Ranged: +3.5 /
// +2.7 / +3.3 over Rhok'delar's 2.9 sec bow at 180 sec), with the 15% quiver.
// Stamina 17.2% of the points and Spirit 10% (the 5a targets), hit 80 rating (8%, the ranged cap
// against a level 63 boss with no hit suppression), then Agility 35.2%, crit 18.4%, ranged attack power 4.7%, Intellect
// 3.5% and Strength 1.1%. The head and chest carry Wildstalker Armor, so only its 2 piece bonus (1%
// attack and cast speed) is live.
// Trinkets are synthetic too (standards 1b), since Forever's launch trinkets are not simulated: 32
// Agility and 14 crit rating (45.9 of budget), and an on-use 415 ranged attack power for 20 sec
// every 2 min (37 weighted points over a 3 minute fight, Earthstrike's size).
export const GearLevel60Ranged = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic Ranged', Level60RangedGear, {
	tooltip: 'A guess at Phase 1 BiS for a ranged hunter under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});
// The Survival set: melee attack power and Strength in place of ranged attack power, a pair of
// one-handers for Predator's Edge's off hand bonus at Brutality Blade's damage per second (a 2.9 sec
// axe and a 1.6 sec dagger, the Era Phase 1 BiS types and speeds Survival sims best with, standards
// 22), and
// 5% hit, since Survival takes Surefooted's 3% (8% yellow cap). Totals: armor 705.5 (target 706.1),
// swords 34.5 each (the standard has no hunter one-hander: the physical 0.39 x the hunter /
// physical two-hander ratio, 0.42 x L(71)), bow 19.5 (Rhok'delar's 2.9 sec: Survival never shoots,
// so every ranged candidate sims the same for it). Stamina 15.4%, Spirit 10.1%, Agility 32.1%,
// crit 15.0%, Strength 11.2%, attack power 9.9%, hit 6.3%. Crusader on the main hand and Agility on
// the off hand (+4 over Agility on both). Trinkets: 32 Agility and 14 crit rating, and an on-use
// 335 attack power and ranged attack power for 20 sec every 2 min.
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
// Summon Hawk), Sniper Shot on its own, Rapid Fire, Bestial Wrath and the racials at the pull and
// on cooldown. The knob searches (last 2026-10-23, at 120 / 180 / 300 sec on the gear for the 8% hit
// cap): all three put Arcane Shot above Aimed Shot with no mana floor on it or on Serpent Sting
// (Beast Mastery +4.4, Serpent Sting last). Marksmanship and Lone Wolf cast Serpent Sting after
// Aimed Shot, Aimed Shot only above 40% mana, and Sniper Shot while the mana covers 10 a second of the
// fight left.
// The ranged rotations melee weave from 12 yd: between Auto Shots the hunter walks in for a main hand
// swing with Raptor Strike queued, when that holds the next Auto Shot back 1.2 sec or less and the
// shot cooldowns stay down meanwhile (Beast Mastery +9.5 over standing at range). Marksmanship and
// Lone Wolf weave only while the mana covers 25 a second of the fight left: weaving wins a short
// fight and costs a long one its mana (2026-10-23; at range Marksmanship is 4.7 behind at 120 sec,
// 1.3 and 1.8 ahead at 180 and 300). The turret rotations stand at range.
const weaveTooltip = ' Melee weaves from 12 yd: walks in for a main hand swing with Raptor Strike queued when the next Auto Shot waits 1.2 sec or less.';
const manaWeaveTooltip =
	' Melee weaves from 12 yd while the mana covers 25 a second of the fight left: walks in for a main hand swing with Raptor Strike queued when the next Auto Shot waits 1.2 sec or less.';
export const APLBeastMastery = PresetUtils.makePresetAPLRotation('Beast Mastery', BeastMasteryAPL, {
	tooltip:
		'Arcane Shot, Aimed Shot and Sniper Shot on cooldown, then Serpent Sting when down, Bestial Wrath and Rapid Fire on cooldown, no mana floors.' +
		weaveTooltip,
	group: 'Level 60',
});
export const APLMarksmanship = PresetUtils.makePresetAPLRotation('Marksmanship', MarksmanshipAPL, {
	tooltip:
		'Arcane Shot, Aimed Shot above 40% mana, Serpent Sting when down, and Sniper Shot while the mana lasts the fight, Rapid Fire on cooldown.' +
		manaWeaveTooltip,
	group: 'Level 60',
});
export const APLLoneWolf = PresetUtils.makePresetAPLRotation('Lone Wolf', LoneWolfAPL, {
	tooltip: 'The Marksmanship rotation, for no pet.' + manaWeaveTooltip,
	group: 'Level 60',
});
export const APLBeastMasteryTurret = PresetUtils.makePresetAPLRotation('Beast Mastery (turret)', BeastMasteryTurretAPL, {
	tooltip: 'The Beast Mastery rotation, standing at 25 yd.',
	group: 'Level 60',
});
export const APLMarksmanshipTurret = PresetUtils.makePresetAPLRotation('Marksmanship (turret)', MarksmanshipTurretAPL, {
	tooltip: 'The Marksmanship rotation, standing at 25 yd.',
	group: 'Level 60',
});
export const APLLoneWolfTurret = PresetUtils.makePresetAPLRotation('Lone Wolf (turret)', LoneWolfTurretAPL, {
	tooltip: 'The Lone Wolf rotation, standing at 25 yd.',
	group: 'Level 60',
});
// Melee: Mongoose Bite whenever Expose Prey opens it, Strider Kick and Immolation Trap on
// cooldown (the trap ahead of Raptor Strike since it scales with ranged attack power,
// hunter_trap_ap), Raptor Strike, Lacerate kept up and Wing Clip in every free global cooldown,
// held only for an opened Mongoose Bite (rotation audit pass 23: +3.5 to +3.8 over the old order
// with a Strider Kick hold too). The Wing Clip filler was the knob search's one change (+31 at 180
// sec): Survival runs out of global cooldowns long before it runs out of mana. It walks in 6 sec before the
// pull, since the fight starts at 25 yards.
export const APLSurvival = PresetUtils.makePresetAPLRotation('Survival', SurvivalAPL, {
	tooltip: 'Melee: Mongoose Bite, Strider Kick, Immolation Trap, Raptor Strike, Lacerate, Wing Clip filler.',
	group: 'Level 60',
});

// Hunter's Mark duty (standards 23): the base rotation plus the hunter's own Hunter's Mark, cast 1.5
// sec before the pull (it doesn't pull; user ruling 2026-10-15) and again before its 2 min run out,
// never held back for the hunter's own damage. The duty presets
// switch the raid's Hunter's Mark off; the base presets take the mark from the raid and never cast
// it.
const markDutyTooltip = " Plus Hunter's Mark 1.5 sec before the pull and before it drops, for the hunter on Hunter's Mark duty.";
export const APLBeastMasteryMark = PresetUtils.makePresetAPLRotation("Beast Mastery (Hunter's Mark duty)", BeastMasteryMarkAPL, {
	tooltip: 'The Beast Mastery rotation.' + markDutyTooltip,
	group: 'Level 60',
});
export const APLMarksmanshipMark = PresetUtils.makePresetAPLRotation("Marksmanship (Hunter's Mark duty)", MarksmanshipMarkAPL, {
	tooltip: 'The Marksmanship rotation.' + markDutyTooltip,
	group: 'Level 60',
});
export const APLLoneWolfMark = PresetUtils.makePresetAPLRotation("Lone Wolf (Hunter's Mark duty)", LoneWolfMarkAPL, {
	tooltip: 'The Lone Wolf rotation.' + markDutyTooltip,
	group: 'Level 60',
});
export const APLSurvivalMark = PresetUtils.makePresetAPLRotation("Survival (Hunter's Mark duty)", SurvivalMarkAPL, {
	tooltip: 'The Survival rotation.' + markDutyTooltip,
	group: 'Level 60',
});
// Scorpid Sting duty (standards 23), for the ranged hunters only (user ruling 2026-10-15: not melee
// Survival): Scorpid Sting kept up in the sting slot in place of Serpent Sting (a hunter has one
// sting on a target), refreshed in its last 2 sec (Beast Mastery: 4 sec, its busier global cooldowns
// let a 2 sec refresh lapse). The duty presets switch the raid's Scorpid Sting
// off; the base presets take it from the raid and never cast it.
const scorpidDutyTooltip = ' With Scorpid Sting kept up in place of Serpent Sting, for the hunter on Scorpid Sting duty.';
export const APLBeastMasteryScorpid = PresetUtils.makePresetAPLRotation('Beast Mastery (Scorpid Sting duty)', BeastMasteryScorpidAPL, {
	tooltip: 'The Beast Mastery rotation.' + scorpidDutyTooltip,
	group: 'Level 60',
});
export const APLMarksmanshipScorpid = PresetUtils.makePresetAPLRotation('Marksmanship (Scorpid Sting duty)', MarksmanshipScorpidAPL, {
	tooltip: 'The Marksmanship rotation.' + scorpidDutyTooltip,
	group: 'Level 60',
});
export const APLLoneWolfScorpid = PresetUtils.makePresetAPLRotation('Lone Wolf (Scorpid Sting duty)', LoneWolfScorpidAPL, {
	tooltip: 'The Lone Wolf rotation.' + scorpidDutyTooltip,
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [
		APLBeastMastery,
		APLBeastMasteryMark,
		APLMarksmanship,
		APLMarksmanshipMark,
		APLLoneWolf,
		APLLoneWolfMark,
		APLSurvival,
		APLSurvivalMark,
		APLBeastMasteryScorpid,
		APLMarksmanshipScorpid,
		APLLoneWolfScorpid,
		APLBeastMasteryTurret,
		APLMarksmanshipTurret,
		APLLoneWolfTurret,
	],
};

export const DefaultAPL = APLBeastMastery;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent searches (tools/rotopt -talent-search, a hill climb) ran on the synthetic sets in
// the mksettings raid at 180 sec, from several starting builds per spec, on 2026-10-12, and last
// on the standards 1b gear with its trinkets (2026-10-13), where every build held, and again on
// 2026-10-23 on the gear for the 8% hit cap, where Lone Wolf and Survival held.
//
// Beast Mastery 31/20: Bestial Wrath, Frenzy 4, Ferocity, Unleashed Fury, Focused Fire and
// Deadly Aspects, and Careful Aim, Lethal Attacks, Improved Arcane Shot 3, Improved Stings 2 and
// Mortal Shots 5 in Marksmanship. The 2026-10-23 search (8% hit cap gear, 180 sec) moved the
// Trueshot Aura point to Mortal Shots (+12.2): the raid's Trueshot Aura already covers the hunter.
// The searches drop Summon Hawk: for a point of Spirit Bond it is 11.8 behind at 180 sec.
export const TalentsBeastMastery = PresetUtils.makePresetTalents('Beast Mastery', SavedTalents.create({ talentsString: '5020121505021241-00520503005' }), {
	tooltip: 'Level 60 Beast Mastery 31/20 from the talent search.',
	group: 'Level 60',
});
// Marksmanship 7/39/5 with a pet: Efficiency 5, Careful Aim, Rapid Killing 1, Improved Arcane Shot 5,
// Trueshot Aura, Mortal Shots, Barrage, Ranged Weapon Specialization, Sniper Shot, Improved
// Stings, Deadly Aspects, Focused Fire and Improved Tracking 5. A search from a ranged Survival
// build ended on Marksmanship too. The 2026-10-23 search moved a Rapid Killing point (it never
// triggers on a single boss) into Efficiency 5 (+4.3 at 180 sec).
export const TalentsMarksmanship = PresetUtils.makePresetTalents('Marksmanship', SavedTalents.create({ talentsString: '502-0053551501503051-5' }), {
	tooltip: 'Level 60 Marksmanship 7/39/5 from the talent search, with a pet.',
	group: 'Level 60',
});
// Marksmanship with Lone Wolf (20% damage with no pet out) and the pet dismissed: Lone Wolf and Rapid
// Killing 2 where the pet build has Focused Fire.
export const TalentsLoneWolf = PresetUtils.makePresetTalents('Lone Wolf', SavedTalents.create({ talentsString: '5-0053552511503051-5' }), {
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

// A cat at 1.0 sec (Broken Tooth, Badlands). With the family abilities from the client (2026-10-23,
// 180 sec, 20000 iterations, rechecked on the gear for the 8% hit cap) the raptor's Savage Rend puts
// it 2.5 Beast Mastery, 1.5 Marksmanship and 1.5 Survival over a cat of the same speed, but at 1.2 sec
// it is 2.0, 1.0 and 0.8 behind the 1.0 sec cat, and no tamable raptor swings faster than 2.0 sec in
// the creature data (the beasts' Faster Attack passives aren't in it) nor in any source we have, so
// the preset keeps the cat. Slower Attack III (2.7 sec) costs 16.4, 8.7 and 9.4. The
// numbers below are from before the family abilities. A pet's hit scales with its attack speed, as in
// Classic (report assumption hunter_pet_speed, user ruling 2026-10-15: the beta's "every pet hits
// like a 2.0 sec pet" is treated as a bug), so its white damage a second is the same at any speed.
// Every pet family at 1.0, 1.2, 1.5 and 2.0 sec was simmed on Beast Mastery, Marksmanship and
// Survival (180 sec, 10000 iterations, 2026-10-15): the cat is still best (the raptor ties it, with
// the same damage modifier, Bite and Claw), 5 to 9 ahead of the next family (the owl, then the
// bat), and the cat's speeds sim within 1 DPS of each other (20000 iterations), so the preset
// keeps the 1.0 sec cat. The pet knows Cobra Reflexes (30% faster, hits as much smaller; +6 Beast
// Mastery, +3 Marksmanship and Survival with Pet Aggression's 8 a hit, 2026-10-23), as wowsims/forever's
// presets have it; Beast Training's point budget isn't known. The ammo is Swiftfeather Arrow, Forever's
// Engineering 300 arrow at 24.5 damage a second (+15 to +22 over Thorium Headed Arrow at range; the user
// put it in the presets on 2026-10-23).
export const DefaultOptions = HunterOptions.create({
	ammo: Ammo.SwiftfeatherArrow,
	quiverBonus: Hunter_Options_QuiverBonus.Speed15,
	petAttackSpeed: PetAttackSpeed.One,
	petType: PetType.Cat,
	petUptime: 1,
	cobraReflexes: true,
});

export const LoneWolfOptions = HunterOptions.create({
	...DefaultOptions,
	petType: PetType.PetNone,
});

// The level 60 consumables, tried one at a time on the Beast Mastery, Lone Wolf and Survival builds
// at 180 sec, last after the shared consumables merge (the list and the numbers are in
// tools/rotopt/mksettings/spec_hunter.go). Every build keeps Elixir of the Mongoose as its one +2%
// crit elixir, Ground Scorpok Assay, Juju Might and the Engineering explosives, and takes Elixir
// of Ferocity (+11 to +13 for the ranged specs), Flank au Poivre and Elixir of Greater Intellect.
// The ranged specs take Major Mana Potion (Major Frenzy Potion falls behind once the mana runs
// out), Demonic Rune, Greater Mageblood Elixir (Mageblood Potion's recipe only comes from
// Zandalar) and Flask of Distilled Wisdom. No pet consumables: Forever pets take no player buffs,
// and Juju and the pet scrolls are player-cast buffs on the pet (they would add 17 to Beast
// Mastery if Forever allows them).
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	fillerExplosive: Explosive.ExplosiveGoblinLandMine,
	// The natural flask (standards 11, 2026-10-23, rechecked on the gear for the 8% hit cap):
	// Aggression, on Beast Mastery / Lone Wolf at 180 sec 14.4 / 6.6 ahead of Swiftness, 37.1 / 39.6 of
	// Accuracy (the hit is capped) and 36.0 / 20.9 of the Flask of Distilled Wisdom. Major Mana Potion is
	// 25.7 / 46.1 ahead of Major Frenzy Potion, and the Dragonfire and Caustic Smog Potions are as far
	// behind.
	flask: Flask.FlaskOfNaturalAggression,
	food: Food.FoodFlankAuPoivre,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
});

// Survival never runs short of mana, so the rune, Mageblood and the flask do nothing for it. It
// takes Major Frenzy Potion (+5.9), Dragonbreath Chili (+3.2) and a Dense Sharpening Stone on each
// weapon (+9.2, measured on the swords; Elemental Sharpening Stone's plans only drop in Molten Core). Fire-toasted Bun's 2%
// hit would add 9, but Forever's is a seasonal item that expires 14 days after looting.
export const SurvivalConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultPotion: Potions.MajorFrenzyPotion,
	dragonBreathChili: true,
	fillerExplosive: Explosive.ExplosiveGoblinLandMine,
	// The natural flask (standards 11, 2026-10-23): Precision, +50.7 over none at 180 sec; on the gear
	// for the 8% hit cap +30.0 over Accuracy, +31.3 over Swiftness and +32.8 over Aggression. Major
	// Frenzy Potion stays 1.0 ahead of the Dragonfire Potion and 1.4 of Caustic Smog.
	flask: Flask.FlaskOfNaturalPrecision,
	food: Food.FoodFlankAuPoivre,
	mainHandImbue: WeaponImbue.DenseSharpeningStone,
	offHandImbue: WeaponImbue.DenseSharpeningStone,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
});

// The raid the level 60 numbers were measured under (tools/rotopt/mksettings): both paladin
// blessings, Arcane Brilliance and every debuff, the hunter's own Hunter's Mark, and a shaman
// in the hunter's group for Grace of Air and Mana Spring, and the raid's Trueshot Aura (a raid aura
// since client 1.60.1.70425, so a Marksmanship hunter in the raid covers every hunter).
export const DefaultRaidBuffs = RaidBuffs.create({
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
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
	salvation: true,
});

// Survival melees, so its shaman drops Windfury Totem and Strength of Earth for the melee group,
// on top of the Grace of Air every hunter has (the user's ruling 2026-10-16, "both totems":
// client 1.60.1.70425 lets Grace of Air stack with Windfury Totem and makes both raid-wide).
export const SurvivalRaidBuffs = RaidBuffs.create({
	...DefaultRaidBuffs,
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
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
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	demoralizingRoar: TristateEffect.TristateEffectImproved,
	demoralizingShout: TristateEffect.TristateEffectImproved,
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

// The hunter is an engineer for the Goblin Sapper and Land Mine (+12 together) and an alchemist.
// The synthetic gear is fixed, so no gear needs a profession.
export const OtherDefaults = {
	distanceFromTarget: 25,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceHuman,
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
	...PresetUtils.LEVEL_60_RAID_DAMAGE,
	...PresetUtils.LEVEL_60_ZONE,
	targets: [
		{
			id: 213336,
			name: 'Level 60',
			level: 63,
			mobType: MobType.MobTypeUnknown,
			stats: new Stats().withStat(Stat.StatAttackPower, 805).withStat(Stat.StatArmor, 3731).withStat(Stat.StatHealth, 127393).asArray(),
			minBaseDamage: 3000,
			damageSpread: 0.3333,
			swingSpeed: 2,
			parryHaste: true,
		},
	],
});

// The level 60 raid setup: the generic level 60 boss (level 63, 3731 armor, type Unknown) in a 3
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
	consumes: SurvivalConsumes,
});
// The same raids with the raid's Hunter's Mark off, for the hunter on Hunter's Mark duty.
export const MarkDutyDebuffs = Debuffs.create({ ...DefaultDebuffs, huntersMark: TristateEffect.TristateEffectMissing });
export const EncounterLevel60MarkDuty = PresetUtils.makePresetEncounter("Level 60 (Hunter's Mark duty)", level60Boss, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: MarkDutyDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});
export const EncounterLevel60MeleeMarkDuty = PresetUtils.makePresetEncounter("Level 60 Melee (Hunter's Mark duty)", level60Boss, {
	tanks: [],
	raidBuffs: SurvivalRaidBuffs,
	debuffs: MarkDutyDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: SurvivalConsumes,
});
// And with the raid's Scorpid Sting off, for the hunter on Scorpid Sting duty.
export const ScorpidDutyDebuffs = Debuffs.create({ ...DefaultDebuffs, scorpidSting: false });
export const EncounterLevel60ScorpidDuty = PresetUtils.makePresetEncounter('Level 60 (Scorpid Sting duty)', level60Boss, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: ScorpidDutyDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: DefaultConsumes,
});

// The numbers are without the Ahn'Qiraj book ranks (Aspect of the Hawk 5, Serpent Sting 8) and
// with the shared raid buff caps (Grace of Air 77, Strength of Earth 42, Battle Shout 115).
// Races (2026-10-23, the gear for the 8% hit cap and the talents and rotations above, 20000
// iterations at 120 / 180 / 300 sec on seeds 1 and 1000003, forever-wiki/sim-status.md has the
// table): no racial depends on the weapon. Beast Mastery is best as a Human (Ruthless's 2% crit
// reaches its weave swings and, through the hunter's melee crit, the pet), 3.7 ahead of the Orc and 5.0
// of the Windshaper at 180 sec. The Windshaper is best for Marksmanship and Lone Wolf (Wind Blessed's
// 3% haste), 1.6 and 1.7 ahead of the High Order. For Survival the Human is best at 180 and 300 sec,
// 0.8 and 2.2 ahead of the Orc, which leads by 1.9 at 120 sec. With the Wildstalker Armor 4
// and 5 piece bonuses on top of the 2 piece (the presets carry the 2 piece only, as every class
// does), Lone Wolf would gain 17, Marksmanship 13, Beast Mastery 1 and Survival nothing, with the
// rotations re-tuned for the shorter Aimed Shot cooldown (measured on the 2026-10-12 gear, before
// standards 1b and its trinkets).
const professions: [Profession, Profession] = [OtherDefaults.profession1, OtherDefaults.profession2];

// 1176.9 / 1172.0 / 1159.3 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23, on seeds 1 and
// 1000003 within 0.5), the class's best at 180 sec: a Human, Mortal Shots 5 for Trueshot Aura, and
// Arcane Shot first with no mana floors. It was 1151.4 / 1145.6 / 1124.3 on the Windshaper before.
export const PresetBuildBeastMastery = PresetUtils.makePresetBuild('Level 60 Beast Mastery', {
	group: 'Level 60',
	tooltip: 'Level 60 Beast Mastery on the synthetic Phase 1 ranged gear with a 1.0 sec cat, in a 3 minute raid fight.',
	gear: GearLevel60Ranged,
	talents: TalentsBeastMastery,
	rotation: APLBeastMastery,
	encounter: EncounterLevel60,
	race: Race.RaceHuman,
	professions,
	options: DefaultOptions,
});
// 1176.5 / 1171.5 / 1157.9 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23), 0.4 / 0.5 / 1.4
// behind the base preset: the raid's Hunter's Mark is off and the hunter casts its own 1.5 sec before
// the pull and before it drops (standards 23).
export const PresetBuildBeastMasteryMark = PresetUtils.makePresetBuild("Level 60 Beast Mastery (Hunter's Mark duty)", {
	group: 'Level 60',
	tooltip:
		"Level 60 Beast Mastery on the synthetic Phase 1 ranged gear with a 1.0 sec cat, in a 3 minute raid fight, keeping Hunter's Mark up itself (the raid's mark off).",
	gear: GearLevel60Ranged,
	talents: TalentsBeastMastery,
	rotation: APLBeastMasteryMark,
	encounter: EncounterLevel60MarkDuty,
	race: Race.RaceHuman,
	professions,
	options: DefaultOptions,
});
// 1081.0 / 1076.5 / 1067.2 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23), 95.9 / 95.5 / 92.1
// behind the base preset: the raid's Scorpid Sting is off and the hunter keeps its own up in place of
// Serpent Sting, refreshed 4 sec before it drops (standards 23), on Beast Mastery's earlier knobs.
export const PresetBuildBeastMasteryScorpid = PresetUtils.makePresetBuild('Level 60 Beast Mastery (Scorpid Sting duty)', {
	group: 'Level 60',
	tooltip:
		"Level 60 Beast Mastery on the synthetic Phase 1 ranged gear with a 1.0 sec cat, in a 3 minute raid fight, keeping Scorpid Sting up itself in place of Serpent Sting (the raid's Scorpid Sting off).",
	gear: GearLevel60Ranged,
	talents: TalentsBeastMastery,
	rotation: APLBeastMasteryScorpid,
	encounter: EncounterLevel60ScorpidDuty,
	race: Race.RaceHuman,
	professions,
	options: DefaultOptions,
});
// 1184.8 / 1167.6 / 1140.2 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23), 4.4 behind
// Beast Mastery at 180 sec (1176.2 / 1145.7 / 1108.4 before the 8% hit cap gear, Efficiency 5 and the
// mana-gated weave).
export const PresetBuildMarksmanship = PresetUtils.makePresetBuild('Level 60 Marksmanship', {
	group: 'Level 60',
	tooltip: 'Level 60 Marksmanship on the synthetic Phase 1 ranged gear with a 1.0 sec cat, in a 3 minute raid fight.',
	gear: GearLevel60Ranged,
	talents: TalentsMarksmanship,
	rotation: APLMarksmanship,
	encounter: EncounterLevel60,
	race: Race.RaceSkyborneWindshaper,
	professions,
	options: DefaultOptions,
});
// 1183.5 / 1166.1 / 1137.3 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23), 1.3 / 1.5 / 2.9
// behind the base preset: the raid's Hunter's Mark is off and the hunter casts its own 1.5 sec before
// the pull and before it drops (standards 23).
export const PresetBuildMarksmanshipMark = PresetUtils.makePresetBuild("Level 60 Marksmanship (Hunter's Mark duty)", {
	group: 'Level 60',
	tooltip:
		"Level 60 Marksmanship on the synthetic Phase 1 ranged gear with a 1.0 sec cat, in a 3 minute raid fight, keeping Hunter's Mark up itself (the raid's mark off).",
	gear: GearLevel60Ranged,
	talents: TalentsMarksmanship,
	rotation: APLMarksmanshipMark,
	encounter: EncounterLevel60MarkDuty,
	race: Race.RaceSkyborneWindshaper,
	professions,
	options: DefaultOptions,
});
// 1091.3 / 1083.0 / 1079.4 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23), 93.5 / 84.6 / 60.8
// behind the base preset: the raid's Scorpid Sting is off and the hunter keeps its own up in place of
// Serpent Sting (standards 23).
export const PresetBuildMarksmanshipScorpid = PresetUtils.makePresetBuild('Level 60 Marksmanship (Scorpid Sting duty)', {
	group: 'Level 60',
	tooltip:
		"Level 60 Marksmanship on the synthetic Phase 1 ranged gear with a 1.0 sec cat, in a 3 minute raid fight, keeping Scorpid Sting up itself in place of Serpent Sting (the raid's Scorpid Sting off).",
	gear: GearLevel60Ranged,
	talents: TalentsMarksmanship,
	rotation: APLMarksmanshipScorpid,
	encounter: EncounterLevel60ScorpidDuty,
	race: Race.RaceSkyborneWindshaper,
	professions,
	options: DefaultOptions,
});
// 1174.2 / 1159.3 / 1122.8 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23): Lone Wolf's 20% is
// now 8.3 behind a cat at 180 sec (it was 1166.0 / 1139.5 / 1101.8). Holds up in a 5 minute fight with
// Sniper Shot and the weave held back for mana.
export const PresetBuildLoneWolf = PresetUtils.makePresetBuild('Level 60 Lone Wolf', {
	group: 'Level 60',
	tooltip: 'Level 60 Marksmanship with Lone Wolf and no pet, on the synthetic Phase 1 ranged gear, in a 3 minute raid fight.',
	gear: GearLevel60Ranged,
	talents: TalentsLoneWolf,
	rotation: APLLoneWolf,
	encounter: EncounterLevel60,
	race: Race.RaceSkyborneWindshaper,
	professions,
	options: LoneWolfOptions,
});
// 1172.9 / 1156.8 / 1119.3 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23), 1.3 / 2.5 / 3.5
// behind the base preset: the raid's Hunter's Mark is off and the hunter casts its own 1.5 sec before
// the pull and before it drops (standards 23).
export const PresetBuildLoneWolfMark = PresetUtils.makePresetBuild("Level 60 Lone Wolf (Hunter's Mark duty)", {
	group: 'Level 60',
	tooltip:
		"Level 60 Marksmanship with Lone Wolf and no pet, on the synthetic Phase 1 ranged gear, in a 3 minute raid fight, keeping Hunter's Mark up itself (the raid's mark off).",
	gear: GearLevel60Ranged,
	talents: TalentsLoneWolf,
	rotation: APLLoneWolfMark,
	encounter: EncounterLevel60MarkDuty,
	race: Race.RaceSkyborneWindshaper,
	professions,
	options: LoneWolfOptions,
});
// 1064.7 / 1059.3 / 1050.3 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23), 109.5 / 100.0 /
// 72.5 behind the base preset: the raid's Scorpid Sting is off and the hunter keeps its own up in
// place of Serpent Sting (standards 23).
export const PresetBuildLoneWolfScorpid = PresetUtils.makePresetBuild('Level 60 Lone Wolf (Scorpid Sting duty)', {
	group: 'Level 60',
	tooltip:
		"Level 60 Marksmanship with Lone Wolf and no pet, on the synthetic Phase 1 ranged gear, in a 3 minute raid fight, keeping Scorpid Sting up itself in place of Serpent Sting (the raid's Scorpid Sting off).",
	gear: GearLevel60Ranged,
	talents: TalentsLoneWolf,
	rotation: APLLoneWolfScorpid,
	encounter: EncounterLevel60ScorpidDuty,
	race: Race.RaceSkyborneWindshaper,
	professions,
	options: LoneWolfOptions,
});
// 1058.0 / 1049.3 / 1041.6 at 120 / 180 / 300 sec in the mksettings raid (Human, 20000 iterations,
// 2026-10-23; 1056.7 / 1047.8 / 1040.0 at 9% hit). The rotation walks in before the pull, so starting
// at 25 yards in the UI costs nothing.
export const PresetBuildSurvival = PresetUtils.makePresetBuild('Level 60 Survival', {
	group: 'Level 60',
	tooltip: 'Level 60 melee Survival on the synthetic Phase 1 Survival gear, in a 3 minute raid fight with a melee group.',
	gear: GearLevel60Survival,
	talents: TalentsSurvival,
	rotation: APLSurvival,
	encounter: EncounterLevel60Melee,
	race: Race.RaceHuman,
	professions,
	options: DefaultOptions,
});
// 1056.9 / 1047.9 / 1039.9 at 120 / 180 / 300 sec (20000 iterations, 2026-10-23), 1.1 / 1.4 / 1.7
// behind the base preset: the raid's Hunter's Mark is off and the hunter casts its own 1.5 sec before
// the pull and before it drops (standards 23).
export const PresetBuildSurvivalMark = PresetUtils.makePresetBuild("Level 60 Survival (Hunter's Mark duty)", {
	group: 'Level 60',
	tooltip:
		"Level 60 melee Survival on the synthetic Phase 1 Survival gear, in a 3 minute raid fight with a melee group, keeping Hunter's Mark up itself (the raid's mark off).",
	gear: GearLevel60Survival,
	talents: TalentsSurvival,
	rotation: APLSurvivalMark,
	encounter: EncounterLevel60MeleeMarkDuty,
	race: Race.RaceHuman,
	professions,
	options: DefaultOptions,
});
