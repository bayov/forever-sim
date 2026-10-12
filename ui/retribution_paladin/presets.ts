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
	StrengthBuff,
	TotemWeaponBuff,
	TristateEffect,
	UnitReference,
	UnitReference_Type,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { PaladinAura, PaladinOptions as RetributionPaladinOptions, PaladinSeal } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import APLBasicRetJson from './apls/basic_ret.apl.json';
import Level30p5APLJSON from './apls/level30p5.apl.json';
import Level60APLJSON from './apls/level60.apl.json';
import Level60ShockAPLJSON from './apls/level60_shock.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import Level30p5GearJSON from './gear_sets/level30p5.gear.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear, { tooltip: 'No gear equipped.' });
// The gear search at level 30 with 26 talent points and Seal of Command, over everything
// an Undead paladin can equip, started from the old level 20 set. The professions are
// Blacksmithing and Enchanting, and the Scourge Invasion drops and the items no source is
// known for are left out. Corpsemaker (Razorfen Kraul) takes the two hander for 8.8 DPS. Blacksmithing makes
// the helm, shirt, leggings and belt (the Crusader's chain pieces and Justicar's Belt,
// bind on pickup) and the Hard Gold Boots. Bloodmage Mantle comes from the Scarlet
// Monastery Library and Unearthed Bands of Power from Uldaman.
//
// The first search took Scorn's Icy Choker (+0.7). Scorn only spawns in the Scourge
// Invasion, so we left his drops out and Mark of the Pack Leader keeps the neck. Manual
// Crowd Pummeler is left out too, see the shaman's Level 30 + 5 set.
//
// On 2026-10-05 the two-hander went from Lesser Strength (+15 Strength) to Impact (+6 weapon
// damage), because the +15 Strength and +15 Agility enchants are not in the beta. That is
// 130.7 against 131.7 at 60 sec. Revelation adds nothing here (128.8, the same as no enchant),
// because the rotation casts no spell that can trigger it.
export const GearLevel30p5 = PresetUtils.makePresetGear('Level 30 + 5', Level30p5GearJSON, {
	tooltip: 'Level 30 gear with Corpsemaker and Blacksmithing crafts.',
	group: 'Level 30',
});

// Synthetic Phase 1 retribution gear, from tools/database/forever_synthetic_items_paladin.go (ids
// from 992000, not in the game).
//
// Forever's level 60 client gear is not tuned yet and Molten Core and Onyxia have no Forever data,
// so a gear search on real items ends up in Classic gear. These items rebuild the Era Phase 1
// retribution set (Lionheart Helm, Onyxia Tooth Pendant, Truestrike Shoulders, Onslaught Girdle,
// Band of Accuria, Quick Strike Ring, Spinal Reaper and the rest) under the budget standard and
// item level table every class's synthetic set uses (66 for head, shoulders, chest, hands, legs
// and feet, 71 for neck, back, wrist, waist and rings, 76 for the two-hander), with a hybrid
// Strength, spell power and Intellect mix, about 10% Spirit, 9% hit and 5% crit. The head and
// chest carry Justice Battlegear, Forever's Tier 1 retribution set, so only its 2-piece bonus (1%
// attack and cast speed) is live, the same shape as every class's preset. The 4-piece (36 attack
// power against Undead) adds nothing on the raid boss, and the 5-piece (0.5 sec off Judgement)
// would add 5.3 at 180 sec on the Level 60 build, 5.4 on sub Holy and 9.1 on Shock. The file has
// the budget math. 9% hit is the melee cap from gear. The Level 60 build's Precision takes it to
// 12%, which only its spells use (they miss up to 16%): trading 1% of it for 10 Strength or 18
// spell power moved every build by 4 DPS or less.
//
// The two-hander is an axe with Spinal Reaper's 203 - 305 at 3.4, the Era Phase 1 retribution
// weapon, so no paladin race gets a weapon racial from it. The relic is Libram of Law (Pix
// Xizzix, 4% Judgement of Righteousness damage), 0.8 ahead of no libram on the twist build and
// 3.8 on Shock (measured on the earlier set). Libram of Fervor does nothing for these builds. The trinket slots are empty:
// Hand of Justice and Blackhand's Breadth are not in Forever's client, and the sim does not have
// Forever's launch trinkets yet (the orchestrator's ruling for every class). The enchants are the
// best Phase 1 ones, Forever's Enchanting 300 enchants included: Lesser Arcanum of Voracity (+8
// Strength) on head and legs, Necklace Spell Power, Cloak Agility, Greater Stats, Bracer Greater
// Spellpower, Gloves Holy Power, a Mystic Rugged Armor Kit on the boots and Spellblasting on the
// two-hander. Spellblasting is 10 DPS over Crusader on the twist build and 27 on Shock, Holy Power
// 2 to 7 over Gloves Superior Strength, and the Rugged kits lose on chest and legs and tie on the
// boots. The gear stays fixed for the talent, rotation and race searches.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {};

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// The paladin_ret template with the knobs the level 60 search settled on (see
// tools/rotopt/paladin_presets.sh), searched over 120, 180 and 300 sec together on the Level 60
// talents and the synthetic set. Judgement of the Crusader comes from the raid (our own Seal of
// the Crusader opener is 11 behind). Consecration goes first down to 20% mana. Seal of Command
// stays up and Twist of Light twists to Seal of Righteousness and back in the last 1.5 sec
// before each swing, so every swing gets both seals (126 DPS over no twisting). Judgement and
// Holy Strike on cooldown, Seal of Righteousness judged whenever Judgement is ready (no swing
// window, 3 DPS over waiting for one), Hammer of Wrath below 20% health. Seal of Righteousness as the seal
// that stays up is 1.5 behind.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Level 60', Level60APLJSON, {
	tooltip: 'Level 60 rotation: Consecration, Seal of Command twisted with Seal of Righteousness before every swing, Judgement and Holy Strike on cooldown.',
	group: 'Level 60',
});
// The rotation for the Level 60 Shock talents (paladin_presets.sh, searched the same way): no
// Twist of Light, so no twisting. Consecration first down to the last point of mana, Judgement on
// cooldown on Seal of Righteousness, Holy Strike on cooldown (90 DPS), Light's Vigil above 40%
// mana just before Holy Shock is ready (64), so the Holy Shock after it detonates it and starts no
// cooldown, then Holy Shock (148), and Hammer of Wrath below 20% health (23).
export const APLLevel60Shock = PresetUtils.makePresetAPLRotation('Level 60 Shock', Level60ShockAPLJSON, {
	tooltip: "Level 60 Holy Shock rotation: Consecration, Light's Vigil and Holy Shock, Judgement and Holy Strike on cooldown.",
	group: 'Level 60',
});
export const APLBasicRet = PresetUtils.makePresetAPLRotation('Basic Ret', APLBasicRetJson, {
	tooltip: 'Basic level 60 Retribution rotation.',
	group: 'Level 60',
});
// The paladin_ret template with the knobs the level 30 search settled on (see
// tools/rotopt/paladin_presets.sh), with the level 30 ranks: Seal of the Crusader judged
// once for the fight, then Consecration whenever it is off cooldown, Seal of Command judged
// on every cooldown and Holy Strike on cooldown. Consecration goes first and spends the
// mana to the last point.
export const APLLevel30p5 = PresetUtils.makePresetAPLRotation('Level 30 + 5', Level30p5APLJSON, {
	tooltip: 'Level 30 Seal of Command rotation with Consecration and Holy Strike on cooldown.',
	group: 'Level 30',
});

export const APLPresets = {
	[Phase.Phase1]: [],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [APLLevel60, APLLevel60Shock, APLBasicRet, APLLevel30p5],
	[Phase.Phase5]: [],
};

export const DefaultAPL = APLPresets[Phase.Phase4][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/forever/talent-calc and copy the numbers in the url.

// The level 60 talent search on the synthetic Phase 1 set (2026-10-12, tools/rotopt
// -talent-search at 180 sec with the Level 60 rotation, then every finalist at 120, 180 and 300
// sec with 50000 iterations, Undead). We started it from three archetypes: Retribution to Twist
// of Light with the side points in Holy, the same with them in Protection, and a Holy Power
// build that leaves Twist of Light out (30 Holy, 21 Retribution).
//
// Level 60: the Retribution tree down to Twist of Light (Benediction, Improved Judgement, Holy
// Conduit, Conviction, Vindication, Sanctified Judgement, Seal of Command, Sacred Arbiter,
// Two-Handed Weapon Specialization, Vengeance, Champion of the Light, Instrument of Law 1),
// Toughness 5 to reach Precision 3, and Divine Strength 5, Divine Intellect 2 and Improved Seals
// 3 in Holy. 974.4 / 968.6 / 963.6 at 120 / 180 / 300 sec on the final set. The search's last move (a Vindication
// point into Divine Intellect for mana) was a tie. Twist of Light is the build: without it the
// seals cannot be twisted, and the Holy Power build was 24% behind before any search.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '52003-503-05225331001330311' }), {
	tooltip: 'Level 60 talents from the talent search: Retribution to Twist of Light, Precision, Divine Strength and Improved Seals.',
	group: 'Level 60',
});
// Level 60 (sub Holy): the side points in Holy instead of Protection (Divine Strength 5, Divine
// Intellect 5, Improved Seals 3, Reverence 3, Purifying Power 2) and Instrument of Law 2 for a
// Holy Conduit point. It was the search's best from the Holy seed. 969.6 / 963.7 / 956.2 at 120 /
// 180 / 300 sec, 0.5% behind Level 60 at 120 sec and 0.8% at 300 because it has no Precision.
export const TalentsLevel60SubHoly = PresetUtils.makePresetTalents(
	'Level 60 (sub Holy)',
	SavedTalents.create({ talentsString: '550030032--05215331001330321' }),
	{ tooltip: 'Level 60 talents with the side points in Holy instead of Protection.', group: 'Level 60' },
);
// Level 60 Shock: the Holy Shock playstyle, its own talent and rotation search from a 31 Holy
// seed (Holy Shock, Divine Precision, Consecrated Ground, Holy Power, Light's Vigil). Divine
// Strength 5, Divine Intellect 5, Improved Seals 3, Reverence 3, Purifying Power 2, Illumination 1
// (to reach row 5), Divine Favor, Holy Shock, Divine Precision 2, Consecrated Ground 2, Holy
// Power 5 and Light's Vigil, with Benediction, Improved Judgement, Holy Conduit, Conviction,
// Vindication 2, Sanctified Judgement 3 and Sacred Arbiter in Retribution. 971.1 / 955.1 / 919.7
// at 120 / 180 / 300 sec on the final set (Spellblasting and Holy Power favour it), 1.4% behind
// Level 60 at 180 sec and 4.6% at 300: without Twist of Light it has nothing for the free GCDs
// between its cooldowns, and it runs out of mana in long fights. Light's Vigil is worth 64 and Holy Shock 148.
export const TalentsLevel60Shock = PresetUtils.makePresetTalents('Level 60 Shock', SavedTalents.create({ talentsString: '55003003201121251--05225230001' }), {
	tooltip: "Level 60 Holy Shock talents: deep Holy to Light's Vigil with Holy Power, and Retribution to Sacred Arbiter.",
	group: 'Level 60',
});
// Level 30 with 5 extra talent points (26 in all), Pursuit of Justice 2 pinned. Holy
// Shock is out of reach with it: Holy Shock needs 20 Holy points and Pursuit of Justice
// 10 Retribution points. We scored every Seal of Command build over the Retribution
// damage talents, with up to 8 Holy points in Divine Strength, Divine Intellect and
// Improved Seals (56865 builds). The best is Deflection 5, Improved Judgement 2, Holy
// Conduit 2, Conviction 4, Sanctified Judgement 3, Seal of Command, Sacred Arbiter,
// Two-Handed Weapon Specialization 3, Vengeance 2 and Champion of the Light 1, and it
// stays on top on the Level 30 + 5 gear.
//
// The 2026-10-01 development notes cut Champion of the Light to 20% of Intellect per
// point (was 33%). We rescored the best 150 builds and every build that moves its point
// to another talent, and this one is still first. The best build without Champion of
// the Light is 1.4 DPS behind at 60 sec.
export const TalentsLevel30p5 = PresetUtils.makePresetTalents('Level 30 + 5', SavedTalents.create({ talentsString: '--502240312013201' }), {
	tooltip: 'Level 30 with 5 extra talent points, all in Retribution for Seal of Command.',
	group: 'Level 30',
});

export const TalentPresets = {
	[Phase.Phase1]: [],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [TalentsLevel60, TalentsLevel60SubHoly, TalentsLevel60Shock, TalentsLevel30p5],
};

export const DefaultTalents = TalentPresets[Phase.Phase4][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = RetributionPaladinOptions.create({
	aura: PaladinAura.RetributionAura,
	primarySeal: PaladinSeal.Command,
});

// The level 60 consumables. We tried every consumable the sim applies in each slot on the
// synthetic set at 180 sec (and 300 for the mana ones), each against the full set. Flask of
// Supreme Power is worth 65, Dragonbreath Chili 25, Shadow Oil 30 over no oil (18 over an
// Elemental Sharpening Stone), Elixir of Greater Firepower 16 (Forever made it the Elixir of
// Holy Power, +40 Holy, so it is for the seals, judgements and Consecration), Elixir of the
// Mongoose 21 (before the one crit elixir rule), R.O.I.D.S. 9, Greater Arcane Elixir 15, Goblin Land Mine 4, Elixir of Shadow Power
// 1.5, Goblin Sapper 1.6. The Major Mana Potion and Demonic Rune are 39 and 28 at 300 sec, and
// the Major Spellblasting or Frenzy Potion in the potion's place lose 33 to 36 there. Flask of the
// Titans or Distilled Wisdom give nothing here.
//
// With Forever's Phase 1 consumables in the sim (the shared consumables commit): a character
// takes at most one of the five 2% crit elixirs (Mongoose, Grizzly, Owl, Sages, Whale, the
// orchestrator's ruling, see notes.md). With Greater Agility, Juju Power, Greater Intellect and
// Greater Spirit in the slots, the Level 60 build does 965.1 at 180 sec, and one crit elixir
// adds 13 (Mongoose, Grizzly) to 15 (Owl, Sages, Whale, all the same since it is the crit that
// counts). We take the Elixir of the Owl in the Intellect slot (980.5), on the Shock build too
// (937.1, Sages is 0.5 ahead there, inside the noise). These were on the earlier set with
// trinkets. On the final set, a second crit elixir (Mongoose in place of Greater Agility) would add
// 12.4 on the Level 60 build (968.6 to 981.0 at 180 sec) and 7.6 on Shock (955.1 to 962.7). Elixir of
// Giants is 1 to 2 behind Juju Power, and Elixir of Ferocity (18 Strength and Agility) beats Juju
// Power on the final set by 1.6 / 1.5 / 0.8 at 180 sec (twist / sub Holy / Shock). Greater Spirit is worth 1.3 on Shock and nothing on the
// twist build, Greater Agility 9 and 5. Lobster Stew beats Nightfin Soup by 3 (Mightfish Steak
// and Flank au Poivre are 3 to 4 behind), and Greater Mageblood Elixir takes the Mageblood
// Potion's place (the same at 180 sec, 3.3 at 300).
export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfGreaterAgility,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	dragonBreathChili: true,
	fillerExplosive: Explosive.ExplosiveGoblinLandMine,
	firePowerBuff: FirePowerBuff.ElixirOfGreaterFirepower,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfTheOwl,
	mainHandImbue: WeaponImbue.ShadowOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfGreaterSpirit,
	strengthBuff: StrengthBuff.ElixirOfFerocity,
	zanzaBuff: ZanzaBuff.ROIDS,
});

// The level 60 raid (tools/rotopt/mksettings): both paladin blessings, Kings, Judgement of the
// Crusader and of Wisdom from another paladin, every debuff, Curse of the Elements (10% Holy
// damage under Forever, worth 62 here). The paladin is in a melee group with a shaman, who puts
// down Strength of Earth, Windfury Totem (one air totem per party since the 2026-09-24 build) and
// Mana Spring.
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
	strengthOfEarthTotem: TristateEffect.TristateEffectRegular,
	thorns: TristateEffect.TristateEffectImproved,
	totemWeaponBuff: TotemWeaponBuff.TotemWeaponBuffWindfury,
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

// At level 60 the paladin is an engineer (the Goblin Sapper and the Land Mine) and a
// blacksmith. The level 30 build keeps Blacksmithing and Enchanting. Gear another profession
// makes is only in the presets when it binds on equip, so it can be bought.
export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceUndead,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor, Humanoid), a 3 minute
// fight, and the default buffs, debuffs and consumes, so switching back from the Level 30 build
// restores everything that build changed.
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
// Level 60: the Undead on the synthetic set with the Level 60 talents and rotation. Every race a
// paladin can be, on that build with the same seed at 50000 iterations (120 / 180 / 300 sec):
// Undead 974.4 / 968.6 / 963.6, Human 958.4 / 952.7 / 947.9, Dwarf 957.3 / 951.8 / 946.9. Undead
// is the best race and the only Horde paladin, Human the best Alliance one (5% Spirit), 1 ahead of
// Dwarf. These are also the numbers with weapon racials off, since the axe gets none. Touch of the Grave rolls on every Judgement, Holy Strike and Consecration. The synthetic
// two-hander is an axe like the Era Phase 1 Spinal Reaper, so neither Human's sword crit nor
// Dwarf's mace crit counts. With a sword of the same stats Human would be about 983 at 180 sec,
// ahead of Undead, and Dwarf with a mace about 959.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Level 60', {
	group: 'Level 60',
	tooltip: 'Level 60 Undead on the synthetic Phase 1 gear with the Level 60 talents and rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});
// Level 60 (sub Holy): the same with the sub Holy talents. Undead 969.6 / 963.7 / 956.2 at 120 /
// 180 / 300 sec, Human 954.8 / 948.8 / 942.3, Dwarf 953.7 / 947.8 / 940.6.
export const PresetBuildLevel60SubHoly = PresetUtils.makePresetBuild('Level 60 (sub Holy)', {
	group: 'Level 60',
	tooltip: 'Level 60 Undead with the side talent points in Holy instead of Protection.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60SubHoly,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: DefaultOptions,
});

// Level 60 Shock: the Holy Shock talents and rotation on Seal of Righteousness. Races at 120 /
// 180 / 300 sec (50000 iterations): Undead 971.1 / 955.1 / 919.7, Human 955.1 / 939.7 / 905.0,
// Dwarf 954.0 / 938.2 / 903.3.
export const PresetBuildLevel60Shock = PresetUtils.makePresetBuild('Level 60 Shock', {
	group: 'Level 60',
	tooltip: "Level 60 Undead Holy Shock build with Light's Vigil, on the synthetic Phase 1 gear.",
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Shock,
	rotation: APLLevel60Shock,
	encounter: EncounterLevel60,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.RetributionAura,
		primarySeal: PaladinSeal.Righteousness,
	}),
});

// A level 30 paladin soloing Interrogator Vishas (Scarlet Monastery Graveyard, level 32):
// no raid buffs or consumes, its own Blessing of Kings (2.4 DPS ahead of Blessing of Might
// at 60 sec), and the boss hitting the paladin so Retribution Aura gets to fire.
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
				tankIndex: 0,
			},
		],
	}),
	{
		tanks: [UnitReference.create({ type: UnitReference_Type.Player, index: 0 })],
		raidBuffs: RaidBuffs.create({}),
		debuffs: Debuffs.create({}),
		buffs: IndividualBuffs.create({ blessingOfKings: true }),
		consumes: Consumes.create({}),
	},
);
// Level 30 with 5 extra talent points. The optimizer puts Seal of Righteousness 2.6 DPS
// behind Seal of Command on Corpsemaker, even with its spell power counted twice. At
// 30 / 60 / 300 sec that is 130.5 / 131.8 / 105.3, and 187.9 on three targets.
export const PresetBuildLevel30p5 = PresetUtils.makePresetBuild('Level 30 + 5', {
	group: 'Level 30',
	tooltip: 'Level 30 Undead with 5 extra talent points and Seal of Command, soloing Interrogator Vishas.',
	gear: GearLevel30p5,
	talents: TalentsLevel30p5,
	rotation: APLLevel30p5,
	encounter: EncounterLevel30,
	race: Race.RaceUndead,
	level: 30,
	bonusTalentPoints: 5,
	professions: [Profession.Blacksmithing, Profession.Enchanting],
	options: RetributionPaladinOptions.create({
		aura: PaladinAura.RetributionAura,
		primarySeal: PaladinSeal.Command,
	}),
});
