import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	FirePowerBuff,
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
	ShadowPowerBuff,
	SpellPowerBuff,
	SpiritElixir,
	Stat,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { SavedTalents } from '../core/proto/ui.js';
import {
	WarlockOptions as WarlockOptions,
	WarlockOptions_Armor as Armor,
	WarlockOptions_Summon as Summon,
	WarlockOptions_WeaponImbue as WarlockWeaponImbue,
} from '../core/proto/warlock.js';
import { Stats } from '../core/proto_utils/stats.js';
// apls
import Level60AfflictionAPL from './apls/level60_affliction.apl.json';
import Level60DemonologyAPL from './apls/level60_demonology.apl.json';
import Level60DestructionAPL from './apls/level60_destruction.apl.json';
import Level60DestructionFireAPL from './apls/level60_destruction_fire.apl.json';
import Level60DemonologyDutyCurseAPL from './apls/level60_demonology_duty_curse.apl.json';
import Level60DestructionDutyCurseAPL from './apls/level60_destruction_duty_curse.apl.json';
import Level60AfflictionDutyCurseAPL from './apls/level60_affliction_duty_curse.apl.json';
import Level60DestructionFireDutyCurseAPL from './apls/level60_destruction_fire_duty_curse.apl.json';
import Level60DemonologyDutyRecklessnessAPL from './apls/level60_demonology_duty_recklessness.apl.json';
import Level60DestructionDutyRecklessnessAPL from './apls/level60_destruction_duty_recklessness.apl.json';
import Level60AfflictionDutyRecklessnessAPL from './apls/level60_affliction_duty_recklessness.apl.json';
import Level60DestructionFireDutyRecklessnessAPL from './apls/level60_destruction_fire_duty_recklessness.apl.json';
import BasicRotation from './apls/rotation.apl.json';
// gear
import BlankGear from './gear_sets/blank.gear.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import MCGear from './gear_sets/mc.gear.json';
import PreBisGear from './gear_sets/prebis.gear.json';

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);
export const GearPreBis = PresetUtils.makePresetGear('Pre-BIS', PreBisGear);
export const GearMC = PresetUtils.makePresetGear('MC', MCGear);

// Forever's level 60 gear is not tuned yet and Molten Core has no Forever data, so the Level 60
// presets wear synthetic items (tools/database/forever_synthetic_items_warlock.go): the Era Phase 1
// BiS (the Felheart set, Robe of Volatile Power, Mana Igniting Cord, Choker of the Fire Lord) rebuilt
// on Forever's budget for a caster (standards 1b: spell power at 0.86, each armor piece at slot mod x
// (1.125 item level + 2) x 1.053 / 1.026 / 0.995 in linear stats, four stat lines a piece within 4% of
// the 1.7-norm shape target) in the fair stat mix of real level 60 warlock gear (standards 5a, Season of
// Discovery BiS proportions): Stamina 14.4%, Intellect 16.0%, Spirit 10.1%, spell power 38.3%, crit
// 10.4% (412 spell power, 6.9% crit), hit to the 16% cap against a level 63 boss: 10% from gear with
// Suppression 5/5 and the 2 piece. Armor 709.6, the staff (item level 76) at Forever's caster rate of
// 193.4 with Enchant 2H Weapon - Spellblasting (+65 spell power), the wand 21.3. Two synthetic
// trinkets: +38 spell power and 13 crit rating, and an on-use 140 spell power for 20 sec every 2 min
// (no real Phase 1 trinket is obtainable or simulated). At 180 sec the builds went from the first,
// all-spell-power set (524 spell power, 5% crit, no trinkets) at 1088.5 / 984.9 / 976.2 / 976.2 to
// 1053.2 / 948.2 / 939.8 / 935.0 on the fair mix (1a) and 1115.5 / 990.0 / 982.3 / 973.6 now (the on-use trinket used before the pull, Death Coil and Searing Pain when the filler would not land before the end)
// (Demonology / Destruction / Affliction / Destruction fire). Only the head and chest carry
// Demonheart Raiment, Forever's warlock Tier 1 set, so only its 2 piece (1% hit) is live; on the first
// set the 5 piece's 20% more mana from Life Tap would have added 12.9 on Demonology and 8.4 to 8.8 on
// Destruction and Affliction at 180 sec (the 4 piece is 21 spell power against Demons only).
// Enchants, the best Phase 1 one per slot with no raid source: Lesser Arcanum of Voracity on the head
// (Arcanum of Focus is a Zandalar one), Enchant Necklace - Spell Power, the Mystic Rugged Armor Kit
// (+6 spell power) on the chest, hands, legs and feet, Bracer - Greater Spellpower and Spellblasting on
// the staff. No Phase 1 enchant adds damage on the shoulders or back.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Demonology', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game. Hit capped with Suppression 5/5.',
	group: 'Synthetic',
});
export const GearPresets = [GearLevel60Phase1Synthetic, GearBlank, GearPreBis, GearMC];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The Level 60 rotations come from tools/rotopt's warlock template (tools/rotopt/warlock.go),
// searched on each build at 180 sec with the Level 60 settings (see warlock_presets.sh). All
// three keep Bane of Doom up when it lands before the fight ends and Bane of Agony otherwise,
// then Corruption and Immolate when they are down, Shadow Bolt as the filler, and Life Tap only
// under 5% mana. A knob search kept every one of those. Against them, at 180 sec on Demonology:
// Bane of Agony alone is 41 behind, no bane 118, no Corruption 45, no Immolate 7, Soul Fire
// during Decimation 11, Death Coil on cooldown 9, Curse of Recklessness 6, Life Tap at 20% mana
// 10. Searing Pain as the filler is 229 behind. Wrack, Incinerate, Siphon Life, the drains and
// Shadowburn never made a build, Shadow Bolt does more with the same time. The opening Shadow Bolt
// starts 2.5 sec before the pull so it lands just after it (damage before the pull does not
// count), Life Tap also goes out when the filler costs more than the mana left, and Demonology
// recasts Corruption and Immolate 1.5 sec early and taps once before the pull to pay for Soul
// Link (the rotation audit's fixes, +10 to +14 at 180 sec). Since 2026-10-23 every build casts a new
// Immolate only with more than 15 sec of the fight left, and Demonology and the fire build a new
// Corruption too (the knob search's near misses, positive at 120 / 180 / 300 sec on both seed sets:
// Demonology +0.8 to +2.2, Destruction +0.3 to +0.6, Affliction +0.3 to +0.8, fire +0.5 to +1.4 with
// its Immolate recast 1.5 sec early).
export const RotationLevel60Demonology = PresetUtils.makePresetAPLRotation('Demonology', Level60DemonologyAPL, {
	group: 'Synthetic',
	tooltip: 'The Level 60 rotation with Soul Link before the pull, for the Demonology talents.',
});
// Destruction adds Conflagrate on cooldown while Immolate is up. Shadow and Flame 5/5 keeps the
// Immolate every time. Destruction and Affliction recast Immolate once it has less left than its
// own cast time, which clips no tick, and Corruption only when it is gone (an early instant
// Corruption clips its last tick).
export const RotationLevel60Destruction = PresetUtils.makePresetAPLRotation('Destruction', Level60DestructionAPL, {
	group: 'Synthetic',
	tooltip: 'The Level 60 rotation with Conflagrate, for the Destruction talents.',
});
export const RotationLevel60Affliction = PresetUtils.makePresetAPLRotation('Affliction', Level60AfflictionAPL, {
	group: 'Synthetic',
	tooltip: 'The Level 60 rotation for the Affliction talents.',
});
// Destruction fire: Incinerate while Immolate is up (25% more), Conflagrate on cooldown, Immolate
// opens the fight (started 1.5 sec before the pull, so it lands as the fight starts) and is
// recast 1.5 sec before it runs out, Corruption too.
export const RotationLevel60DestructionFire = PresetUtils.makePresetAPLRotation('Destruction Fire', Level60DestructionFireAPL, {
	group: 'Synthetic',
	tooltip: 'The Level 60 rotation with Incinerate as the filler, for the Destruction fire talents.',
});
// The raid curse duty rotations (standards 23): the base rotation plus Curse of the Elements or Curse of
// Recklessness, which this warlock keeps up for the raid (put up right after the opener and recast
// before it drops). Under Forever the curse has its own slot, so the bane stays.
export const RotationLevel60DemonologyDutyCurse = PresetUtils.makePresetAPLRotation('Demonology · Curse of the Elements duty', Level60DemonologyDutyCurseAPL, {
	group: 'Synthetic',
	tooltip: 'The Demonology rotation with Curse of the Elements kept up for the raid.',
});
export const RotationLevel60DestructionDutyCurse = PresetUtils.makePresetAPLRotation(
	'Destruction · Curse of the Elements duty',
	Level60DestructionDutyCurseAPL,
	{
		group: 'Synthetic',
		tooltip: 'The Destruction rotation with Curse of the Elements kept up for the raid.',
	},
);
export const RotationLevel60AfflictionDutyCurse = PresetUtils.makePresetAPLRotation('Affliction · Curse of the Elements duty', Level60AfflictionDutyCurseAPL, {
	group: 'Synthetic',
	tooltip: 'The Affliction rotation with Curse of the Elements kept up for the raid.',
});
export const RotationLevel60DestructionFireDutyCurse = PresetUtils.makePresetAPLRotation(
	'Destruction Fire · Curse of the Elements duty',
	Level60DestructionFireDutyCurseAPL,
	{
		group: 'Synthetic',
		tooltip: 'The Destruction Fire rotation with Curse of the Elements kept up for the raid.',
	},
);
export const RotationLevel60DemonologyDutyRecklessness = PresetUtils.makePresetAPLRotation(
	'Demonology · Curse of Recklessness duty',
	Level60DemonologyDutyRecklessnessAPL,
	{
		group: 'Synthetic',
		tooltip: 'The Demonology rotation with Curse of Recklessness kept up for the raid.',
	},
);
export const RotationLevel60DestructionDutyRecklessness = PresetUtils.makePresetAPLRotation(
	'Destruction · Curse of Recklessness duty',
	Level60DestructionDutyRecklessnessAPL,
	{
		group: 'Synthetic',
		tooltip: 'The Destruction rotation with Curse of Recklessness kept up for the raid.',
	},
);
export const RotationLevel60AfflictionDutyRecklessness = PresetUtils.makePresetAPLRotation(
	'Affliction · Curse of Recklessness duty',
	Level60AfflictionDutyRecklessnessAPL,
	{
		group: 'Synthetic',
		tooltip: 'The Affliction rotation with Curse of Recklessness kept up for the raid.',
	},
);
export const RotationLevel60DestructionFireDutyRecklessness = PresetUtils.makePresetAPLRotation(
	'Destruction Fire · Curse of Recklessness duty',
	Level60DestructionFireDutyRecklessnessAPL,
	{
		group: 'Synthetic',
		tooltip: 'The Destruction Fire rotation with Curse of Recklessness kept up for the raid.',
	},
);
export const RotationBasic = PresetUtils.makePresetAPLRotation('Basic', BasicRotation);

export const APLPresets = [
	RotationLevel60Demonology,
	RotationLevel60DemonologyDutyCurse,
	RotationLevel60DemonologyDutyRecklessness,
	RotationLevel60Affliction,
	RotationLevel60AfflictionDutyCurse,
	RotationLevel60AfflictionDutyRecklessness,
	RotationLevel60Destruction,
	RotationLevel60DestructionDutyCurse,
	RotationLevel60DestructionDutyRecklessness,
	RotationLevel60DestructionFire,
	RotationLevel60DestructionFireDutyCurse,
	RotationLevel60DestructionFireDutyRecklessness,
	RotationBasic,
];

export const DefaultAPL = RotationLevel60Demonology;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Searched with tools/rotopt -talent-search on the synthetic set at 180 sec, from seven starting
// builds (deep Affliction, deep Destruction, Demonic Pact with Destruction or Affliction,
// Affliction and Demonology without Pact, Demonic Sacrifice and Ruin with no demon out, and a
// Wrack build), then again on the searched rotation. Every Demonic Pact start ended on the same
// build. It sacrifices the Imp before the pull (15% Shadow damage) and keeps the Succubus out
// (Master Demonologist's 10% Shadow damage, Soul Link, Demonic Knowledge's 60 spell damage).
// Searched again from every preset on 2026-10-23, after the demons took wowsims/forever's measured
// shares and melee (about 4% less): the three Shadow builds kept every point, and the fire build
// moved to its Demonic Sacrifice build. A Demonic Sacrifice of the Imp with no demon out (the search
// from 252-0050003201-055030500310035 and 2550102003520005-0050003201-055) is 31 behind on
// Destruction and 84 on Affliction.
// Since 2026-10-24 Demonology takes Suppression 5/5 (standards 26: its 20% less threat is a must, whatever
// the sim says). The talent search with Suppression pinned (tools/rotopt -require suppression, on the
// shared set, which it caps) took the three points from Agonizing Flames and one from Ruin, and put one
// in Improved Corruption; the same for the curse duty build. Dropping Demonic Pact to keep Agonizing
// Flames is 90 behind. Against the old one point build on its hit set (011-0305003131221001351-0550005003)
// Suppression costs 67.1 / 68.1 / 66.2 at 120 / 180 / 300 sec (66.8 on curse duty at 180 sec).
export const TalentsLevel60Demonology = PresetUtils.makePresetTalents('Demonology', SavedTalents.create({ talentsString: '051-0305003131221001351-0550004' }), {
	group: 'Synthetic',
	tooltip: 'Demonic Pact with Suppression 5/5, Improved Shadow Bolt and Ruin.',
});
// The best build with 30 points in Destruction: Conflagrate, Fire and Brimstone and Shadow and
// Flame, with Pandemic and Malevolence from Affliction. Incinerate and Bane of Havoc lost to two
// points of Malevolence and Shadowburn to more Pandemic. Under client 1.60.1.70425 Malediction's
// point went to Nightfall (searched again 2026-10-16).
export const TalentsLevel60Destruction = PresetUtils.makePresetTalents('Destruction', SavedTalents.create({ talentsString: '255000000351--055030500310035' }), {
	group: 'Synthetic',
	tooltip: 'Conflagrate and Shadow and Flame, with Pandemic and Malevolence.',
});
// The best build with 30 points in Affliction: Shadow Mastery, Pandemic, Malevolence and
// Nightfall, with Improved Shadow Bolt, Ruin and Agonizing Flames. Under client 1.60.1.70425
// Shadow Mastery is 10% and Malediction is gone: its 3 points went to Improved Bane of Agony and one
// Soul Harvest point that only fills the 25 point gate (searched again 2026-10-16). Wrack and Soul Siphon lost
// their points to Ruin, the drains being no match for Shadow Bolt. A damage over time variant with
// Siphon Life, Soul Siphon and Wrack (25530000035001351--0550005003, the best build with both, on
// the same Shadow Bolt rotation) is 55 behind, and casting its Wrack costs 34; its drain filler
// versions lose 124 to 349. It played the same as this build, so it is not shipped as a spec.
export const TalentsLevel60Affliction = PresetUtils.makePresetTalents('Affliction', SavedTalents.create({ talentsString: '2550102003520005--0550305003' }), {
	group: 'Synthetic',
	tooltip: 'Shadow Mastery, Pandemic and Nightfall, with Improved Shadow Bolt and Ruin.',
});

// Destruction fire, its own spec: Incinerate, Conflagrate, Fire and Brimstone, Shadow and Flame and
// Aftermath, with Demonic Sacrifice (the Succubus sacrificed before the pull for 15% Fire damage, no
// demon out), Suppression 5/5 and Firestone. Searched with Incinerate and Conflagrate kept
// (-require). Since the demons take wowsims/forever's measured shares and melee (2026-10-23) the
// sacrifice build is back ahead: 986.1 against 981.0 for the best build with the Succubus out
// (25201020034--0050355003101351, the old preset's search) at 180 sec; with the Succubus out on this
// build 950.3, the Imp sacrificed 917.5, the Felhunter 885.4, and Spellstone 9.7 behind Firestone. A
// talent search from it moved nothing. Bane of Havoc's point does nothing in the sim.
export const TalentsLevel60DestructionFire = PresetUtils.makePresetTalents(
	'Destruction Fire',
	SavedTalents.create({ talentsString: '252-0050003201-0050345003101351' }),
	{ group: 'Synthetic', tooltip: 'Incinerate, Conflagrate and Shadow and Flame, with the Succubus sacrificed for 15% Fire damage.' },
);

export const TalentPresets = [TalentsLevel60Demonology, TalentsLevel60Destruction, TalentsLevel60Affliction, TalentsLevel60DestructionFire];

export const DefaultTalents = TalentsLevel60Demonology;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// The Succubus is out for the Shadow builds, the fire build sacrifices it. Forever's demons take no
// player buffs and scale from the warlock's stats (wowsims/forever's measured shares, notes.md). At
// 180 sec (2026-10-23): on Destruction the Succubus is 17.7 ahead of the Felhunter and 24.9 of the
// Imp; on Affliction 16.2 and 24.6. Since client 1.60.1.70630 Spellstone gives Shadow damage only:
// on Destruction it and Firestone are even (1009.1 and 1009.5 at 180 sec, 20000 iterations,
// 2026-10-27), on Affliction Spellstone is 4.3 ahead and on Demonology 6.9 (5000 iterations).
export const DefaultOptions = WarlockOptions.create({
	armor: Armor.DemonArmor,
	summon: Summon.Succubus,
	sacrificedPet: Summon.Imp,
	weaponImbue: WarlockWeaponImbue.Spellstone,
});

const PresetBuildOptionsDemonology = {
	armor: Armor.DemonArmor,
	summon: Summon.Succubus,
	sacrificedPet: Summon.Imp,
	weaponImbue: WarlockWeaponImbue.Spellstone,
};
const PresetBuildOptionsFire = {
	armor: Armor.DemonArmor,
	summon: Summon.NoSummon,
	sacrificedPet: Summon.Succubus,
	weaponImbue: WarlockWeaponImbue.Firestone,
};
const PresetBuildOptionsSuccubus = {
	armor: Armor.DemonArmor,
	summon: Summon.Succubus,
	sacrificedPet: Summon.NoSummon,
	weaponImbue: WarlockWeaponImbue.Spellstone,
};

// The level 60 consumables. On 2026-10-12 we tried every consumable the sim applies in each slot
// on the Demonology build at 180 sec: Goblin Sapper +4.1 (Engineering), Cerebral Cortex Compound
// +10.7 (a Blasted Lands quest), Demonic Rune +15.8, Elixir of Shadow Power +23.3 and Wizard Oil
// +15.7. With Forever's elixirs and foods simulated: a character takes one of the five +2% crit
// elixirs (Mongoose, Grizzly, Owl, Sages, Whale are role variants of one elixir), and Elixir of
// Sages with Elixir of Greater Intellect is 0.1 to 0.6 ahead of Elixir of the Owl with Elixir of
// Greater Spirit on every build (Sages' Spirit feeds Life Tap). Greater Mageblood Elixir is +1.9
// over Mageblood Potion, Lobster Stew +3.9 over Nightfin Soup, and Major Spellblasting Potion 12.9
// behind Major Mana Potion. Under Forever the Spellstone is an imbue that stacks with an oil. The presets model
// Forever's launch Phase 1, when no vanilla raid is open, so Brilliant Wizard Oil (a Zandalar
// recipe) is out. Elixir of Greater Firepower is Forever's Elixir of Holy Power, so the plain
// Elixir of Firepower takes the slot, for Immolate and Conflagrate. Flask of Distilled Wisdom (-88)
// lost. Explosives are taken only when they gain (standards 11): on every preset (2026-10-24, 120 / 180
// / 300 sec, both seed sets) the Goblin Sapper alone is ahead of none by 2.7 to 7.6, and adding any other
// explosive (Dense Dynamite, Thorium Grenade, EZ-Thro Dark Bomb, Goblin Land Mine, their 1 sec cast in
// place of a Shadow Bolt) costs 1.2 to 6.7; without the Sapper they lose 3.4 to 10.8 to none.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	// The fight is in a natural flask zone (standards 11), but the best natural flask, Aggression, is 80
	// to 89 behind the Flask of Supreme Power at 180 sec (the hit is capped).
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.WizardOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The shared level 60 raid (tools/rotopt/mksettings): Kings and Wisdom (a caster takes no Might), Judgement of Wisdom,
// every debuff, and a caster group's shaman with Mana Spring. Another warlock keeps Curse of the
// Elements up, so ours has its curse free (Forever gives a warlock a bane and a curse). Under
// Forever Improved Shadow Bolt, Shadow Weaving, Improved Scorch and Winter's Chill only help the
// player who applies them, so they are not here.
// Under Forever the demon takes no player buffs (notes.md) and scales from the warlock's stats
// instead. The raid's party buffs are the warlock's caster group's: no Battle Shout, Trueshot Aura
// or Leader of the Pack.
export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	bloodPact: TristateEffect.TristateEffectImproved,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	frostResistanceAura: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	salvation: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectMissing,
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
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	demoralizingRoar: TristateEffect.TristateEffectImproved,
	demoralizingShout: TristateEffect.TristateEffectImproved,
	exposeArmor: TristateEffect.TristateEffectMissing,
	faerieFire: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	scorpidSting: true,
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
});

// Engineering for the Goblin Sapper in the consumes, Tailoring for the cloth. The distance and channel
// clip delay are the DPS Report's (tools/rotopt/mksettings).
export const OtherDefaults = {
	distanceFromTarget: 5,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
	channelClipDelay: 50,
	race: Race.RaceOrc,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor), a 3 minute fight, and
// the default buffs, debuffs and consumes, like the enhancement shaman's.
const level60Encounter = (name: string, debuffs: Debuffs) =>
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
					mobType: MobType.MobTypeUnknown,
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
			debuffs,
			buffs: DefaultIndividualBuffs,
			consumes: DefaultConsumes,
		},
	);
export const EncounterLevel60 = level60Encounter('Generic boss', DefaultDebuffs);
// The curse duty presets' encounter: the raid's Curse of the Elements is off, the warlock on duty
// brings it; the raid's other curses and Faerie Fire stay on (other players' duties).
export const EncounterLevel60CurseDuty = level60Encounter('Curse of the Elements duty', Debuffs.create({ ...DefaultDebuffs, curseOfElements: false }));
// The Curse of Recklessness duty presets' encounter: the warlock keeps Recklessness up in place of a
// druid's Faerie Fire. Under Forever the two share the minor armor slot (standards 20), so the raid's
// Faerie Fire and Curse of Recklessness are both off, and the armor comes only from our curse.
export const EncounterLevel60RecklessnessDuty = level60Encounter(
	'Curse of Recklessness duty',
	Debuffs.create({ ...DefaultDebuffs, curseOfRecklessness: false, faerieFire: false }),
);

// The Level 60 builds at 120 / 180 / 300 sec (20000 iterations, seed 1, 2026-10-23, the natural flask
// zone, wowsims/forever's measured demon shares and melee): Demonology 1081.4 / 1065.1 / 1035.4 (with
// Suppression 5/5 since 2026-10-24, was 1148.4 / 1133.2 / 1101.6),
// Destruction 1018.8 / 1011.8 / 992.4, Affliction 1044.9 / 1035.3 / 1013.3, Destruction fire
// 998.7 / 987.4 / 964.5, all Human (seed 1000003 within 0.9). Forever teaches no Ahn'Qiraj book ranks,
// so Shadow Bolt is rank IX, Immolate VII and Corruption VI. The demons take no player buffs and scale
// from the warlock's stats (notes.md), and their melee and spell crit are the warlock's spell crit with
// none of their own (Blizzard's intent, forever-bugs #331). The demon shares are read at levels 20 and
// 30, not 60. Every build sits at the 16% spell hit cap.
//
// Races at 180 sec, both seed sets (standards 16: a switch needs a lead over 0.5 on both): the Human is
// first on every build. Human Ruthless is 2% crit with every spell, and the demons take it from the
// warlock. Demonology 13.3 / 13.5 ahead of the Gnome and 3.1 / 2.6 of the Orc (with Suppression 5/5), Destruction 3.2 / 3.1 of
// the Orc, Affliction 3.2 / 3.2 of the Orc. Destruction fire is a tie with the Orc (0.1 / 0.0), so it
// stays Human. The best Horde race is the Orc on every build. The full table is in
// forever-wiki/sim-status.md.
// Race re-pick (2026-10-27, standards 29 and 16, after dev notes post 8: Orc Blood Fury 2 min, Troll
// Berserking 15 sec, Undead Touch of the Grave 15 sec): Demonology and Destruction move to the Orc,
// Destruction fire to the Undead. At 180 sec (verify.py, 20000 iterations, seed 1): Demonology 1051.8
// against the Human's 1050.2, Destruction 1001.5 against 1000.1, Destruction fire 986.4 against 981.9.
// Affliction stays Human (the Orc's 0.4 lead in the race sweep is a tie). The curse duty builds follow
// their base. The UI's default race is the Orc, the race of the default build (Demonology) and of
// Destruction.
export const PresetBuildLevel60Demonology = PresetUtils.makePresetBuild('Demonology', {
	group: 'Synthetic',
	tooltip: 'Level 60 Orc Demonology warlock on the synthetic Phase 1 gear, the Imp sacrificed and the Succubus out, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Demonology,
	rotation: RotationLevel60Demonology,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsDemonology,
});
export const PresetBuildLevel60Destruction = PresetUtils.makePresetBuild('Destruction', {
	group: 'Synthetic',
	tooltip: 'Level 60 Orc Destruction warlock on the synthetic Phase 1 gear with the Succubus out, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Destruction,
	rotation: RotationLevel60Destruction,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsSuccubus,
});
export const PresetBuildLevel60Affliction = PresetUtils.makePresetBuild('Affliction', {
	group: 'Synthetic',
	tooltip: 'Level 60 Human Affliction warlock on the synthetic Phase 1 gear with the Succubus out, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Affliction,
	rotation: RotationLevel60Affliction,
	encounter: EncounterLevel60,
	race: Race.RaceHuman,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsSuccubus,
});

export const PresetBuildLevel60DestructionFire = PresetUtils.makePresetBuild('Destruction Fire', {
	group: 'Synthetic',
	tooltip: 'Level 60 Undead Destruction fire warlock on the synthetic Phase 1 gear, the Succubus sacrificed, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60DestructionFire,
	rotation: RotationLevel60DestructionFire,
	encounter: EncounterLevel60,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsFire,
});

// The raid curse duty presets (standards 23): each build with the warlock keeping a curse up for the
// raid itself, with the raid's own switched off. Curse of the Elements is the usual duty. Curse of
// Recklessness takes Faerie Fire's place, because under Forever the two share the minor armor slot.
const curseDutyBuild = (
	base: PresetUtils.PresetBuild,
	rotation: PresetUtils.PresetRotation,
	curse = 'Curse of the Elements',
	encounter = EncounterLevel60CurseDuty,
) => {
	const { name, ...options } = base;
	return PresetUtils.makePresetBuild(`${name} · ${curse} duty`, {
		...options,
		tooltip: `${options.tooltip} On curse duty for the raid, keeping ${curse} up itself.`,
		rotation,
		encounter,
	});
};
const recklessnessDutyBuild = (base: PresetUtils.PresetBuild, rotation: PresetUtils.PresetRotation) =>
	curseDutyBuild(base, rotation, 'Curse of Recklessness', EncounterLevel60RecklessnessDuty);
export const PresetBuildLevel60DemonologyDutyCurse = curseDutyBuild(PresetBuildLevel60Demonology, RotationLevel60DemonologyDutyCurse);
export const PresetBuildLevel60DestructionDutyCurse = curseDutyBuild(PresetBuildLevel60Destruction, RotationLevel60DestructionDutyCurse);
export const PresetBuildLevel60AfflictionDutyCurse = curseDutyBuild(PresetBuildLevel60Affliction, RotationLevel60AfflictionDutyCurse);
export const PresetBuildLevel60DestructionFireDutyCurse = curseDutyBuild(PresetBuildLevel60DestructionFire, RotationLevel60DestructionFireDutyCurse);
export const PresetBuildLevel60DemonologyDutyRecklessness = recklessnessDutyBuild(PresetBuildLevel60Demonology, RotationLevel60DemonologyDutyRecklessness);
export const PresetBuildLevel60DestructionDutyRecklessness = recklessnessDutyBuild(PresetBuildLevel60Destruction, RotationLevel60DestructionDutyRecklessness);
export const PresetBuildLevel60AfflictionDutyRecklessness = recklessnessDutyBuild(PresetBuildLevel60Affliction, RotationLevel60AfflictionDutyRecklessness);
export const PresetBuildLevel60DestructionFireDutyRecklessness = recklessnessDutyBuild(
	PresetBuildLevel60DestructionFire,
	RotationLevel60DestructionFireDutyRecklessness,
);

export const PresetBuilds = [
	PresetBuildLevel60Demonology,
	PresetBuildLevel60DemonologyDutyCurse,
	PresetBuildLevel60DemonologyDutyRecklessness,
	PresetBuildLevel60Affliction,
	PresetBuildLevel60AfflictionDutyCurse,
	PresetBuildLevel60AfflictionDutyRecklessness,
	PresetBuildLevel60Destruction,
	PresetBuildLevel60DestructionDutyCurse,
	PresetBuildLevel60DestructionDutyRecklessness,
	PresetBuildLevel60DestructionFire,
	PresetBuildLevel60DestructionFireDutyCurse,
	PresetBuildLevel60DestructionFireDutyRecklessness,
];
