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
import Level60DemonologyAPL from './apls/level60_demonology.apl.json';
import Level60DestructionAPL from './apls/level60_destruction.apl.json';
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
export const RotationBasic = PresetUtils.makePresetAPLRotation('Basic', BasicRotation);

export const APLPresets = [RotationLevel60Demonology, RotationLevel60Destruction, RotationLevel60Affliction, RotationBasic];

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

export const TalentPresets = [TalentsLevel60Demonology, TalentsLevel60Destruction, TalentsLevel60Affliction];

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
const PresetBuildOptionsSuccubus = {
	armor: Armor.DemonArmor,
	summon: Summon.Succubus,
	sacrificedPet: Summon.NoSummon,
	weaponImbue: WarlockWeaponImbue.Spellstone,
};

// The level 60 consumables. On 2026-10-12 we tried every consumable the sim applies in each slot
// on the Demonology build at 180 sec: Nightfin Soup +7.8 over Runn Tum Tuber Surprise, Goblin
// Sapper +4.1 (Engineering), Mageblood Potion +4.4, Cerebral Cortex Compound +10.7, Demonic Rune
// +15.8, Major Mana Potion +23.7, Elixir of Shadow Power +23.3 and Brilliant Wizard Oil +32.5
// (Wizard Oil +15.7). Under Forever the Spellstone is an imbue that stacks with an oil. Elixir of
// Greater Firepower is Forever's Elixir of Holy Power, so the plain Elixir of Firepower takes the
// slot, for Immolate and Conflagrate. Goblin Land Mine (-7) and Flask of Distilled Wisdom (-88)
// lost.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodNightfinSoup,
	mainHandImbue: WeaponImbue.BrilliantWizardOil,
	manaRegenElixir: ManaRegenElixir.MagebloodPotion,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The shared level 60 raid (tools/rotopt/mksettings): both paladin blessings, Judgement of Wisdom,
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
	blessingOfMight: TristateEffect.TristateEffectImproved,
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

// The three Level 60 builds, Human, at 120 / 180 / 300 sec (20000 iterations, 2026-10-12):
// Demonology 983.5 / 970.7 / 959.9, Destruction 903.9 / 897.4 / 891.8, Affliction 897.3 / 889.5
// / 884.3. Destruction and Affliction are 7.5% and 8.4% behind Demonology at 180 sec, inside the
// 10% that earns a build of its own, and each is deep in a tree of its own.
//
// Human is the best warlock race for all three, then Gnome, Undead, Orc and Troll. Its 2% crit
// comes from Sword Specialization with the synthetic Azuresong Mageblade, the Era Phase 1 sword.
// Without that crit the Human Demonology warlock is 934.7 at 180 sec, last of the five, so a
// staff or a dagger would put the Gnome on top. The best Horde race is the Undead, 6.0 behind the
// Human on Demonology at 180 sec. The full table is in forever-wiki/sim-status.md.
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

export const PresetBuilds = [PresetBuildLevel60Demonology, PresetBuildLevel60Destruction, PresetBuildLevel60Affliction];
