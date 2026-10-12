import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	Alcohol,
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
import { ShadowPriest_Options as Options, ShadowPriest_Options_Armor as Armor } from '../core/proto/priest.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import Level60APL from './apls/level60.apl.json';
import Level60SmiteAPL from './apls/level60_smite.apl.json';
import P1APL from './apls/p1.apl.json';
import Level60Phase1SmiteGear from './gear_sets/level60_phase1_smite.gear.json';
import Level60Phase1SyntheticGear from './gear_sets/level60_phase1_synthetic.gear.json';
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
// tools/database/forever_synthetic_items_priest.go (ids from 995000, not in the game).
//
// Forever's level 60 gear in the client is not tuned yet, and Molten Core and Onyxia have no
// Forever data, so a gear search on real items ends up in Classic gear. The synthetic items
// take the place of the Era Phase 1 shadow BiS (the Molten Core tier, Choker of the Fire Lord,
// Mana Igniting Cord, Azuresong Mageblade, Fire Runed Grimoire, the Molten Core rings) at Phase 1
// epic budgets with Forever's cheaper spell power (the budget standard all classes use, see the
// file): 561 spell power, 75 Spirit (about 10% of the budget), 3% hit and 4% crit. The six armor pieces carry Raiments of Conviction, Forever's
// Tier 1 shadow set (1% hit, and Devouring Plague without a cooldown). The trinkets are
// Talisman of Ephemeral Power and Briarwood Reed. The file has the budget math.
//
// The enchants are Phase 1 caster ones: Lesser Arcanum of Voracity (Intellect) on head and
// legs, Forever's Enchant Necklace - Spell Power, Greater Stats, Greater Intellect on the
// wrists, Spirit on the boots and Spell Power on the weapon. The main hand is a sword like
// Azuresong Mageblade, which is what puts a Human level with an Undead (see the build). The
// Smite build wants the same stats and wears the same pieces with another set bonus (below).
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

// The same set for the Smite build, with its six armor pieces swapped for copies that carry
// Vestments of Conviction, Forever's Tier 1 Discipline and Holy set (10 Spirit, 9 spell damage,
// Penance 1 sec shorter), in place of Raiments of Conviction. For the Smite build it is 4.1 DPS
// ahead at 180 sec and 7.8 at 300 (the gear auditor's A/B).
export const GearLevel60Phase1Smite = PresetUtils.makePresetGear('Level 60 Phase 1 Smite', Level60Phase1SmiteGear, {
	tooltip: 'The synthetic Phase 1 set with the Discipline and Holy Tier 1 bonuses, for the Smite build.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic, GearLevel60Phase1Smite, GearP0BIS, GearP1BIS],
};

export const DefaultGear = GearPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLP1Shadow = PresetUtils.makePresetAPLRotation('Shadow', P1APL);

// The tools/rotopt priest_shadow template after a knob search on the Level 60 build (3000
// iterations a candidate, 20000 to confirm, at 120, 180 and 300 sec, 2026-10-12). Shadowform
// before the pull, Shadow Word: Pain and Devouring Plague whenever they are down (Raiments of
// Conviction take Devouring Plague's cooldown), Inner Focus on Mind Blast, Mind Blast and
// Shadow Word: Death on cooldown, Mind Flay to fill. All three lengths agree on it.
// - Mind Flay is not clipped for Mind Blast. Mind Blast's 5.5 sec cooldown against back to back
//   3 sec channels rarely comes up mid channel, so clipping gives up ticks for half a Mind Blast
//   more a fight: 10.6 behind after the first tick at 180 sec.
// - Mind Flay is clipped after its second tick when Shadow Word: Pain or Devouring Plague fell
//   off during it, so the damage over time spell goes back up sooner (+3.6 at 180 sec). Without
//   it a 3 sec channel left either one down for up to 3 sec.
// - Vampiric Embrace is left out. It heals the party and deals no damage, and keeping it up costs
//   17.6 DPS at 180 sec in GCDs. A raid that wants its healing can turn the knob on.
// - Mind Blast above Shadow Word: Death, 2.3 ahead, so Inner Focus lands on Mind Blast.
// - Starshards (Night Elf) and Chastise (Dwarf) are left out. In Shadowform a Starshards tick
//   hits no harder than a Mind Flay tick and the 6 sec channel holds up Mind Blast, 30 DPS
//   lost at any length. Chastise loses 4.7 to the Mind Flay GCD it takes.
// - Dark Sacrifice (Undead) under 40% mana. Mana never gets that low in a raid fight on this
//   set (about 25 min to empty), so it is there for longer fights.
//
// A hand audit of the timeline on the final build (2026-10-12): Shadowform holds all fight,
// Shadow Weaving reaches 5 stacks 6 sec in and stays there, Inner Focus goes on the first Mind
// Blast and again after its 3 min, Shadow Word: Pain and Devouring Plague go back up within a GCD
// of falling off, no GCD is idle, and mana stays above 6000.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Level 60', Level60APL, {
	tooltip: "The optimizer's best level 60 rotation: both damage over time spells up, Mind Blast and Shadow Word: Death on cooldown, Mind Flay to fill.",
	group: 'Level 60',
});

// The tools/rotopt priest_smite template after a knob search on the Smite talents (3000
// iterations a candidate, 20000 to confirm, at 120, 180 and 300 sec, 2026-10-12): Power
// Infusion on cooldown, Holy Fire whenever its damage over time is down (Power in Light gives
// Smite 10% while it is up), Shadow Word: Pain, Devouring Plague above 20% mana, Chastise for a
// Dwarf, Inner Focus on cooldown for a Smite, then rank 8 Smite while the mana lasts the fight
// at 20 mana a second and rank 4 Smite after. Forever gives Smite ranks 3 and up the full
// coefficient, so the low rank keeps most of the damage for a third of the mana (from the
// rotation audit): +1.8 at 180 sec and +10 at 300 over rank 8 alone. Rank 3 is 0.9 ahead at 300
// sec, but it is learned at level 14 and a Classic style penalty below level 20 would cut it.
// The search drops the spells that give the least damage for their mana: Mind Blast, Shadow
// Word: Death and even Penance (385 mana for three 92 damage bolts). A timeline shows Smite
// every 2 sec (Divine Fury) and Holy Fire back up within a cast of falling off.
export const APLLevel60Smite = PresetUtils.makePresetAPLRotation('Level 60 Smite', Level60SmiteAPL, {
	tooltip:
		"The optimizer's best rotation for the Smite talents: Holy Fire up for Power in Light, Shadow Word: Pain, Smite to fill, a lower rank once mana runs short.",
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60, APLLevel60Smite, APLP1Shadow],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Talent searches (tools/rotopt -talent-search, 3000 iterations a move, 20000 to confirm) on
// the synthetic Phase 1 set at 180 sec, then again at 300 sec and after every knob search,
// from deep Shadow with Discipline, deep Shadow with Holy, and 31 Discipline (Power Infusion)
// with 20 Shadow. On 2026-10-12 the build is 882.7 / 882.1 / 883.5 at 120 / 180 / 300 sec as a
// Human (882.2 / 881.9 / 882.5 Undead).
// - Shadow 33: Shadow Focus 5, Improved Shadow Word: Pain, Improved Mind Blast 5, Mind Flay
//   and Improved Mind Flay 2, Shadow Weaving 3, Devouring Contagion 2, Darkness 5, Shadowform.
//   Shadow Affinity and a point of Shadow Reach only fill the rows (threat and range).
// - Discipline 18: Twin Disciplines 5 (13 DPS), Mental Strength 5, Inner Focus (about 1).
//   These were measured on the first synthetic set (more spell power), the searches on the
//   final set end on the same build.
//   Mental Agility, Meditation and the row 2 points fill the rows: mana never runs out.
// - Early Demise 2 in place of 2 Mental Strength is 0.7 behind.
// - The Power Infusion build loses Shadowform and is out of mana 107 sec in: 601 at 180 sec.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '005030211305--500321501201302051' }), {
	tooltip: 'Level 60 talents from the talent search: deep Shadow with Twin Disciplines, Mental Strength and Inner Focus.',
	group: 'Level 60',
});

// The same Shadow points with 20 in Holy instead of Discipline (Twilight Focus, Spell Warding,
// Inspiration, Improved Healing and the rest of the healing utility a raid may ask for). None of
// the Holy talents adds Shadow damage, so it gives up Twin Disciplines and Mental Strength:
// 852.3 / 850.4 / 852.0 at 120 / 180 / 300 sec as a Human on its own consumes (see the build),
// 3.6% behind at 180 sec.
export const TalentsLevel60Holy = PresetUtils.makePresetTalents('Level 60 + Holy', SavedTalents.create({ talentsString: '-3025011323-500321501201302051' }), {
	tooltip: 'The Level 60 Shadow points with 20 Holy points for healing utility, about 32 DPS behind.',
	group: 'Level 60',
});

// A Holy damage priest: Smite, Holy Fire and Penance instead of Shadowform. Forever's
// Discipline tree is built for it (Power in Light, Twin Disciplines, Penance, Holy Precision),
// so it is its own spec here. Talent searches from 31 Discipline with 20 Holy, 31 Holy with 20
// Discipline and 31 Discipline with 20 Shadow (3000 iterations a move, 20000 to confirm, 180
// sec) end on 31 Discipline (Power in Light 5, Twin Disciplines 5, Holy Precision 3, Mental
// Agility 3, Inner Focus, Meditation 3, Mental Strength 5, Penance, Power Infusion) and 17 Holy
// (Holy Specialization 5, Divine Fury 5, Searing Light 2) with 3 points of Shadow Focus for
// Shadow Word: Pain's hit. The deep Holy start ends at 608.9.
//
// On its own set with its own consumes (see the build) it is 707.3 / 702.8 / 689.0 at 120 / 180
// / 300 sec as an Undead, whose Dark Sacrifice comes when the mana runs low. That is 20% behind
// Shadow at 180 sec and 22% at 300: Smite is a mana hungry filler and nothing in the tree halves its cost the way
// Shadowform does for Shadow spells.
export const TalentsLevel60Smite = PresetUtils.makePresetTalents('Level 60 Smite', SavedTalents.create({ talentsString: '525030231305001001-00505113002-3' }), {
	tooltip: 'Holy damage talents: 31 Discipline (Power in Light, Penance, Power Infusion) and 17 Holy (Divine Fury, Searing Light).',
	group: 'Level 60',
});

export const TalentsP1Shadow = PresetUtils.makePresetTalents('Shadow', SavedTalents.create({ talentsString: '005030231301--500321501201302251' }));

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60Holy, TalentsLevel60Smite, TalentsP1Shadow],
};

export const DefaultTalents = TalentPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = Options.create({
	armor: Armor.InnerFire,
});

// The level 60 consumes, tried in and out of the set on the Level 60 talents (180 sec, 20000
// iterations, 2026-10-12, before Inner Focus joined the build):
// - Lobster Stew (28 spell damage under Forever) is 2.8 ahead of Nightfin Soup (22), which is
//   7.9 ahead of Runn Tum Tuber Surprise and 3.1 ahead of Grilled Squid.
// - Goblin Sapper Charge +3.7. A filler explosive takes a GCD and costs 4.
// - Elixir of the Owl (25 Intellect, 2% crit) +16.4 over the Elixir of Wisdom, the Elixir of
//   Greater Intellect only +2.8 (Arcane Brilliance is on). Elixir of Sages (25 Spirit, 2% crit)
//   +13.6 on top. Greater Arcane Elixir is worth 16.0, Elixir of Shadow Power 18.3.
// - Major Spellblasting Potion (47 spell damage for 30 sec) +8.2 over the Major Mana Potion.
// - Wizard Oil +11. Brilliant Wizard Oil would be 12.5 ahead, but it is a Zul'Gurub recipe.
//   Shadow Oil and Dragonbreath Chili add nothing.
// - Flask of Supreme Power is 68 ahead of the Flask of Distilled Wisdom.
// - Cerebral Cortex Compound (25 Intellect, a Blasted Lands quest) +3.2 at 120 / 180 / 300 sec.
// - Mana never runs out (about 25 min to empty), so Mageblood adds nothing and is left out.
//   The rune is kept for longer fights.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorSpellblastingPotion,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfTheOwl,
	mainHandImbue: WeaponImbue.WizardOil,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The shared level 60 raid of tools/rotopt/mksettings for a caster: both paladin blessings but
// Wisdom in place of Might, Arcane Brilliance, Mana Spring from a shaman in the group, and every
// debuff. Forever has no Curse of Shadow, its Curse of the Elements raises Shadow damage
// (10%), and Shadow Weaving is the priest's own talent under Forever.
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

// Engineering is for the Goblin Sapper Charge in the consumes. Tailoring is the priest's
// crafting, nothing in the presets needs it.
export const OtherDefaults = {
	channelClipDelay: 50,
	distanceFromTarget: 30,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
	race: Race.RaceHuman,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor, Humanoid), a 3 minute
// fight, and the default buffs, debuffs and consumes, like the enhancement shaman's.
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

// Every race a priest can be under Forever, on the Level 60 build with the same seed (20000
// iterations, standard error 0.2 or less), at 120 / 180 / 300 sec (2026-10-12):
//   Human     882.7 / 882.1 / 883.5  (best overall and Alliance)
//   Undead    882.2 / 881.9 / 882.5  (best Horde)
//   Gnome     878.1 / 877.7 / 877.6
//   Night Elf 877.1 / 875.3 / 876.7
//   Dwarf     868.7 / 868.1 / 869.5
//   Troll     868.2 / 867.8 / 869.3
// Human and Undead are level within the noise. Human's 2% crit with swords comes from the main
// hand, a sword like Azuresong Mageblade, the Era Phase 1 BiS. With a staff (Staff of Dominance)
// a Human ties a Dwarf at 805 and an Undead is 14 ahead of both (on the earlier consumes). Undead's Touch of the Grave is
// 14, Gnome's Eureka! 9, Night Elf's Elune's Light 7. Troll's Berserking only speeds up Mind
// Blast's cast, channels and the GCD do not speed up, so it is worth nothing. Starshards and
// Chastise lose DPS (see the rotation).
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Level 60', {
	group: 'Level 60',
	tooltip: 'Level 60 Human on the synthetic Phase 1 gear with the Level 60 rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceHuman,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
// Without Twin Disciplines and Mental Strength the Holy build runs dry on a long fight, so it
// keeps the Major Mana Potion and adds the Greater Mageblood Elixir: on the Spellblasting
// potion it is 859.4 / 854.7 / 728.2 at 120 / 180 / 300 sec, with these 852.3 / 850.4 / 852.0.
export const ConsumesHoly = Consumes.create({
	...DefaultConsumes,
	defaultPotion: Potions.MajorManaPotion,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
});

export const EncounterLevel60Holy = PresetUtils.makePresetEncounter('Level 60 + Holy', EncounterLevel60.encounter, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: ConsumesHoly,
});

export const PresetBuildLevel60Holy = PresetUtils.makePresetBuild('Level 60 + Holy', {
	group: 'Level 60',
	tooltip: 'Level 60 with the Holy sub tree talents and mana consumes, 32 DPS behind Level 60.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Holy,
	rotation: APLLevel60,
	encounter: EncounterLevel60Holy,
	race: Race.RaceHuman,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});

// The Smite build is short of mana, so its consumes keep the Major Mana Potion (the Major
// Spellblasting Potion is 8 behind at 180 sec) and add the Greater Mageblood Elixir and Kreeg's
// Stout Beatdown (25 Spirit, 5 Intellect). Forever turns the Elixir of Greater Firepower into
// the Elixir of Holy Power (40 Holy damage) for Smite and Holy Fire: +15.8 at 180 sec. The
// Elixir of Sages adds 7.0 on top of the Owl, Kreeg's 0.2 (1.0 at 300 sec). The Flask of
// Distilled Wisdom is 71 behind the Flask of Supreme Power even here.
export const ConsumesSmite = Consumes.create({
	...DefaultConsumes,
	alcohol: Alcohol.AlcoholKreegsStoutBeatdown,
	defaultPotion: Potions.MajorManaPotion,
	firePowerBuff: FirePowerBuff.ElixirOfGreaterFirepower,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
});

export const EncounterLevel60Smite = PresetUtils.makePresetEncounter('Level 60 Smite', EncounterLevel60.encounter, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: ConsumesSmite,
});

// Smite as every race on its set and consumes (20000 iterations, 120 / 180 / 300 sec):
//   Undead    707.3 / 702.8 / 689.0  (best overall and Horde)
//   Gnome     711.9 / 700.1 / 682.4  (best Alliance)
//   Human     711.2 / 697.7 / 682.7
//   Night Elf 709.2 / 694.3 / 679.0
//   Troll     707.5 / 693.0 / 677.2
//   Dwarf     696.0 / 682.3 / 668.8
// Undead's Dark Sacrifice and Gnome's larger mana pool and cheaper Eureka! casts are worth the
// most once the mana runs out.
export const PresetBuildLevel60Smite = PresetUtils.makePresetBuild('Level 60 Smite', {
	group: 'Level 60',
	tooltip: 'Level 60 Undead Holy damage priest (Smite, Holy Fire, Power Infusion) on the synthetic Phase 1 gear, 20% behind Shadow.',
	gear: GearLevel60Phase1Smite,
	talents: TalentsLevel60Smite,
	rotation: APLLevel60Smite,
	encounter: EncounterLevel60Smite,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
