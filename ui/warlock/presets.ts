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
	ManaRegenElixir,
	MobType,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	ShadowPowerBuff,
	SpellPowerBuff,
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
import Level60AfflictionDoTAPL from './apls/level60_affliction_dot.apl.json';
import Level60DemonologyAPL from './apls/level60_demonology.apl.json';
import Level60DestructionAPL from './apls/level60_destruction.apl.json';
import Level60DestructionFireAPL from './apls/level60_destruction_fire.apl.json';
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
// presets wear synthetic items (tools/database/forever_synthetic_items_warlock.go): the Era Phase
// 1 BiS (the Felheart set, Robe of Volatile Power, Mana Igniting Cord, Choker of the Fire Lord,
// Azuresong Mageblade) rebuilt at Phase 1 epic budgets with Forever's cheaper spell power, about
// 10% of the budget on Spirit, 3% hit and 5% crit. Six of them carry Demonheart Raiment, Forever's
// warlock Tier 1 set (1% hit, 20% more mana from Life Tap). The trinkets are the real Talisman of
// Ephemeral Power and Briarwood Reed. One set serves all three specs, the file says why. The
// enchants are the Phase 1 ones: Arcanum of Focus on the head and legs, Greater Stats, Greater
// Intellect, Spirit on the boots and Spell Power on the sword.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
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
// Bane of Agony alone is 34 behind, no bane 111, no Corruption 45, no Immolate 9, Soul Fire
// during Decimation 9, Death Coil on cooldown 6, Curse of Recklessness 8, Life Tap at 20% mana
// 9. Searing Pain as the filler is 185 behind. Wrack, Incinerate, Siphon Life, the drains and
// Shadowburn never made a build, Shadow Bolt does more with the same time.
export const RotationLevel60Demonology = PresetUtils.makePresetAPLRotation('Level 60 Demonology', Level60DemonologyAPL, {
	group: 'Level 60',
	tooltip: 'The Level 60 rotation with Soul Link before the pull, for the Demonology talents.',
});
// Destruction adds Conflagrate on cooldown while Immolate is up. Shadow and Flame 5/5 keeps the
// Immolate every time.
export const RotationLevel60Destruction = PresetUtils.makePresetAPLRotation('Level 60 Destruction', Level60DestructionAPL, {
	group: 'Level 60',
	tooltip: 'The Level 60 rotation with Conflagrate, for the Destruction talents.',
});
export const RotationLevel60Affliction = PresetUtils.makePresetAPLRotation('Level 60 Affliction', Level60AfflictionAPL, {
	group: 'Level 60',
	tooltip: 'The Level 60 rotation for the Affliction talents.',
});
// Destruction fire: Incinerate while Immolate is up (25% more), Conflagrate on cooldown, Immolate
// opens the fight and is recast 2 sec before it runs out.
export const RotationLevel60DestructionFire = PresetUtils.makePresetAPLRotation('Level 60 Destruction Fire', Level60DestructionFireAPL, {
	group: 'Level 60',
	tooltip: 'The Level 60 rotation with Incinerate as the filler, for the Destruction fire talents.',
});
// Affliction damage over time: Siphon Life on top, and Wrack while Corruption and Bane of Agony are
// both up (its 10% only counts on those two, so in practice at the end of the fight, when Bane of
// Doom would not land). Wrack whenever the damage over time spells are up was 17 behind, never
// casting it 0.5 ahead (inside the noise).
export const RotationLevel60AfflictionDoT = PresetUtils.makePresetAPLRotation('Level 60 Affliction DoT', Level60AfflictionDoTAPL, {
	group: 'Level 60',
	tooltip: 'The Level 60 rotation with Siphon Life and Wrack, for the Affliction damage over time talents.',
});
export const RotationBasic = PresetUtils.makePresetAPLRotation('Basic', BasicRotation);

export const APLPresets = [
	RotationLevel60Demonology,
	RotationLevel60Destruction,
	RotationLevel60Affliction,
	RotationLevel60DestructionFire,
	RotationLevel60AfflictionDoT,
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
export const TalentsLevel60Demonology = PresetUtils.makePresetTalents(
	'Level 60 Demonology',
	SavedTalents.create({ talentsString: '011-0305003131221001351-0550005003' }),
	{ group: 'Level 60', tooltip: 'Demonic Pact with Improved Shadow Bolt and Ruin.' },
);
// The best build with 30 points in Destruction: Conflagrate, Fire and Brimstone and Shadow and
// Flame, with Pandemic and Malevolence from Affliction. Incinerate and Bane of Havoc lost to two
// points of Malevolence and Shadowburn to more Pandemic.
export const TalentsLevel60Destruction = PresetUtils.makePresetTalents(
	'Level 60 Destruction',
	SavedTalents.create({ talentsString: '25510000035--055030500310035' }),
	{ group: 'Level 60', tooltip: 'Conflagrate and Shadow and Flame, with Pandemic and Malevolence.' },
);
// The best build with 30 points in Affliction: Shadow Mastery, Pandemic, Malevolence and
// Nightfall, with Improved Shadow Bolt, Ruin and Agonizing Flames. Wrack and Soul Siphon lost
// their points to Ruin, the drains being no match for Shadow Bolt.
export const TalentsLevel60Affliction = PresetUtils.makePresetTalents(
	'Level 60 Affliction',
	SavedTalents.create({ talentsString: '2553000003520005--0550305003' }),
	{ group: 'Level 60', tooltip: 'Shadow Mastery, Pandemic and Nightfall, with Improved Shadow Bolt and Ruin.' },
);

// Destruction fire, its own spec: Incinerate, Conflagrate, Fire and Brimstone and Shadow and Flame,
// with the Succubus sacrificed before the pull for 15% Fire damage and no demon out, and
// Firestone. Searched with Incinerate and Conflagrate kept (-require). The Succubus out instead was
// 17 behind, the Imp 67.
export const TalentsLevel60DestructionFire = PresetUtils.makePresetTalents(
	'Level 60 Destruction Fire',
	SavedTalents.create({ talentsString: '252-0050003201-0050345003101351' }),
	{ group: 'Level 60', tooltip: 'Incinerate, Conflagrate and Shadow and Flame, with the Succubus sacrificed for Fire damage.' },
);
// Affliction damage over time, its own spec: everything down to Wrack, with Siphon Life, Soul
// Siphon, Improved Drains and Malediction. Searched with Wrack and Siphon Life kept (-require).
// Drain Life as the filler was 55 behind Shadow Bolt and Drain Soul 162, Drain Soul only below 25%
// of the boss's health lost too. The client gives the drains 10% a tick (Drain Soul 10% every 3
// sec), so Shadow Bolt stays the filler.
export const TalentsLevel60AfflictionDoT = PresetUtils.makePresetTalents(
	'Level 60 Affliction DoT',
	SavedTalents.create({ talentsString: '05550320035201351-0005003-03' }),
	{ group: 'Level 60', tooltip: 'Siphon Life, Soul Siphon and Wrack.' },
);

export const TalentPresets = [
	TalentsLevel60Demonology,
	TalentsLevel60Destruction,
	TalentsLevel60Affliction,
	TalentsLevel60DestructionFire,
	TalentsLevel60AfflictionDoT,
];

export const DefaultTalents = TalentsLevel60Demonology;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// The Succubus is out for every Level 60 build. Forever halved Firebolt, but the demons' melee
// is still 1.12's in the sim, so the Succubus is about 55 DPS ahead of the Imp on the
// Destruction and Affliction builds (notes.md). Spellstone and Firestone are within 3.
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
// on the Demonology build at 180 sec: Nightfin Soup +7.8 over Runn Tum Tuber Surprise, Goblin
// Sapper +4.1 (Engineering), Mageblood Potion +4.4, Cerebral Cortex Compound +10.7 (a Blasted Lands
// quest), Demonic Rune +15.8, Major Mana Potion +23.7, Elixir of Shadow Power +23.3 and Wizard Oil
// +15.7. Under Forever the Spellstone is an imbue that stacks with an oil. The presets model
// Forever's launch Phase 1, when no vanilla raid is open, so Brilliant Wizard Oil (+32.5, a
// Zandalar recipe) is out. Elixir of Greater Firepower is Forever's Elixir of Holy Power, so the plain Elixir of
// Firepower takes the slot, for Immolate and Conflagrate. Goblin Land Mine (-7) and Flask of
// Distilled Wisdom (-88) lost.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodNightfinSoup,
	mainHandImbue: WeaponImbue.WizardOil,
	manaRegenElixir: ManaRegenElixir.MagebloodPotion,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The shared level 60 raid (tools/rotopt/mksettings): Kings and Wisdom (a caster takes no Might), Judgement of Wisdom,
// every debuff, and a caster group's shaman with Mana Spring. Another warlock keeps Curse of the
// Elements up, so ours has its curse free (Forever gives a warlock a bane and a curse). Under
// Forever Improved Shadow Bolt, Shadow Weaving, Improved Scorch and Winter's Chill only help the
// player who applies them, so they are not here.
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
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: true,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	retributionAura: TristateEffect.TristateEffectImproved,
	sanctityAura: true,
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
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
	curseOfShadow: true,
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

// Engineering for the Goblin Sapper in the consumes, Tailoring for the cloth.
export const OtherDefaults = {
	distanceFromTarget: 25,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
	channelClipDelay: 150,
	race: Race.RaceHuman,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor), a 3 minute fight, and
// the default buffs, debuffs and consumes, like the enhancement shaman's.
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

// The Level 60 builds, Human, at 120 / 180 / 300 sec (20000 iterations, 2026-10-12): Demonology
// 988.1 / 976.5 / 964.9, Destruction 903.9 / 897.4 / 891.8, Affliction 897.3 / 889.5 / 884.3,
// Destruction fire 891.8 / 881.6 / 873.9, Affliction damage over time 747.0 / 732.6 / 724.4.
// Destruction and Affliction are inside 10% of Demonology, Destruction fire and Affliction damage
// over time play differently and ship as specs of their own.
//
// Human is the best warlock race overall (first on Demonology, Destruction and Affliction), then
// Gnome, Undead, Orc and Troll. Gnome is 3.8 ahead on Affliction damage over time and Undead 0.7
// on Destruction fire. The Human's 2% crit comes from Sword Specialization with the synthetic
// Azuresong Mageblade, the Era Phase 1 sword. Without that crit the Human Demonology warlock is
// 934.7 at 180 sec, last of the five, so a staff or a dagger would put the Gnome on top. The best
// Horde race is the Undead, 7.7 behind the Human on Demonology at 180 sec. The full table is in
// forever-wiki/sim-status.md.
export const PresetBuildLevel60Demonology = PresetUtils.makePresetBuild('Level 60 Demonology', {
	group: 'Level 60',
	tooltip: 'Level 60 Human Demonology warlock on the synthetic Phase 1 gear, the Imp sacrificed and the Succubus out, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Demonology,
	rotation: RotationLevel60Demonology,
	encounter: EncounterLevel60,
	race: Race.RaceHuman,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsDemonology,
});
export const PresetBuildLevel60Destruction = PresetUtils.makePresetBuild('Level 60 Destruction', {
	group: 'Level 60',
	tooltip: 'Level 60 Human Destruction warlock on the synthetic Phase 1 gear with the Succubus out, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Destruction,
	rotation: RotationLevel60Destruction,
	encounter: EncounterLevel60,
	race: Race.RaceHuman,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsSuccubus,
});
export const PresetBuildLevel60Affliction = PresetUtils.makePresetBuild('Level 60 Affliction', {
	group: 'Level 60',
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

export const PresetBuildLevel60DestructionFire = PresetUtils.makePresetBuild('Level 60 Destruction Fire', {
	group: 'Level 60',
	tooltip: 'Level 60 Human Destruction fire warlock on the synthetic Phase 1 gear, the Succubus sacrificed and no demon out, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60DestructionFire,
	rotation: RotationLevel60DestructionFire,
	encounter: EncounterLevel60,
	race: Race.RaceHuman,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsFire,
});
export const PresetBuildLevel60AfflictionDoT = PresetUtils.makePresetBuild('Level 60 Affliction DoT', {
	group: 'Level 60',
	tooltip: 'Level 60 Human Affliction damage over time warlock on the synthetic Phase 1 gear with the Succubus out, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60AfflictionDoT,
	rotation: RotationLevel60AfflictionDoT,
	encounter: EncounterLevel60,
	race: Race.RaceHuman,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: PresetBuildOptionsSuccubus,
});

export const PresetBuilds = [
	PresetBuildLevel60Demonology,
	PresetBuildLevel60Destruction,
	PresetBuildLevel60Affliction,
	PresetBuildLevel60DestructionFire,
	PresetBuildLevel60AfflictionDoT,
];
