import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Encounter,
	Flask,
	Food,
	IndividualBuffs,
	IntellectElixir,
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
} from '../core/proto/common.js';
import { ShadowPriest_Options as Options, ShadowPriest_Options_Armor as Armor } from '../core/proto/priest.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import Level60APL from './apls/level60.apl.json';
import P1APL from './apls/p1.apl.json';
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
// epic budgets with Forever's cheaper spell power: 635 spell power, 85 Spirit (about 10% of the
// budget), 3% hit and 4% crit. The six armor pieces carry Raiments of Conviction, Forever's
// Tier 1 shadow set (1% hit, and Devouring Plague without a cooldown). The trinkets are
// Talisman of Ephemeral Power and Briarwood Reed. The file has the budget math.
//
// The enchants are Phase 1 caster ones: Lesser Arcanum of Voracity (Intellect) on head and
// legs, Forever's Enchant Necklace - Spell Power, Greater Stats, Greater Intellect on the
// wrists, Spirit on the boots and Spell Power on the weapon. The main hand is a sword like
// Azuresong Mageblade, which is what puts a Human on top of the race table (see the build).
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic, GearP0BIS, GearP1BIS],
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
// - Mind Flay is never clipped. Mind Blast's 5.5 sec cooldown against back to back 3 sec
//   channels rarely comes up mid channel, so clipping after the first or second tick gives up
//   ticks for half a Mind Blast more a fight: 3.2 and 1.8 behind at 180 sec.
// - Mind Blast above Shadow Word: Death, 2.3 ahead, so Inner Focus lands on Mind Blast.
// - Starshards (Night Elf) and Chastise (Dwarf) are left out. In Shadowform a Starshards tick
//   hits no harder than a Mind Flay tick and the 6 sec channel holds up Mind Blast, 30 DPS
//   lost at any length. Chastise loses 4.7 to the Mind Flay GCD it takes.
// - Dark Sacrifice (Undead) under 40% mana. Mana never gets that low in a raid fight on this
//   set (about 25 min to empty), so it is there for longer fights.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Level 60', Level60APL, {
	tooltip: "The optimizer's best level 60 rotation: both damage over time spells up, Mind Blast and Shadow Word: Death on cooldown, Mind Flay to fill.",
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60, APLP1Shadow],
};

export const DefaultAPL = APLPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Talent searches (tools/rotopt -talent-search, 3000 iterations a move, 20000 to confirm) on
// the synthetic Phase 1 set at 180 sec, then again at 300 sec and after every knob search,
// from deep Shadow with Discipline, deep Shadow with Holy, and 31 Discipline (Power Infusion)
// with 20 Shadow. On 2026-10-12 the build is 873.7 / 871.6 / 874.7 at 120 / 180 / 300 sec as a
// Human (873.5 / 871.2 / 874.2 Undead).
// - Shadow 33: Shadow Focus 5, Improved Shadow Word: Pain, Improved Mind Blast 5, Mind Flay
//   and Improved Mind Flay 2, Shadow Weaving 3, Devouring Contagion 2, Darkness 5, Shadowform.
//   Shadow Affinity and a point of Shadow Reach only fill the rows (threat and range).
// - Discipline 18: Twin Disciplines 5 (13 DPS), Mental Strength 5, Inner Focus (0.8 to 1.8).
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
// 849.6 / 848.1 / 835.9 at 120 / 180 / 300 sec as a Human, 2.7% behind at 180 sec. It also
// runs out of mana about 280 sec in (an Undead's Dark Sacrifice carries it to 450 sec).
export const TalentsLevel60Holy = PresetUtils.makePresetTalents('Level 60 + Holy', SavedTalents.create({ talentsString: '-3025011323-500321501201302051' }), {
	tooltip: 'The Level 60 Shadow points with 20 Holy points for healing utility, about 24 DPS behind.',
	group: 'Level 60',
});

export const TalentsP1Shadow = PresetUtils.makePresetTalents('Shadow', SavedTalents.create({ talentsString: '005030231301--500321501201302251' }));

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60Holy, TalentsP1Shadow],
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
// - Nightfin Soup (22 spell damage under Forever) is 7.9 ahead of Runn Tum Tuber Surprise and
//   3.1 ahead of Grilled Squid.
// - Goblin Sapper Charge +3.7. A filler explosive takes a GCD and costs 4.
// - Elixir of Wisdom +0.9. Greater Arcane Elixir is worth 16.0, Elixir of Shadow Power 18.3.
// - Wizard Oil +11. Brilliant Wizard Oil would be 12.5 ahead, but it is a Zul'Gurub recipe.
//   Shadow Oil and Dragonbreath Chili add nothing.
// - Flask of Supreme Power is 68 ahead of the Flask of Distilled Wisdom.
// - Mana never runs out (about 25 min to empty), so Mageblood adds nothing and is left out.
//   The potion and the rune are kept for longer fights.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodNightfinSoup,
	intellectElixir: IntellectElixir.ElixirOfWisdom,
	mainHandImbue: WeaponImbue.WizardOil,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	shadowPowerBuff: ShadowPowerBuff.ElixirOfShadowPower,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
});

// The shared level 60 raid of tools/rotopt/mksettings for a caster: both paladin blessings but
// Wisdom in place of Might, Arcane Brilliance, Mana Spring from a shaman in the group, and every
// debuff (Curse of Shadow is the one a shadow priest gets, Shadow Weaving is the priest's own
// under Forever).
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

// Every race a priest can be under Forever, on the Level 60 build with the same seed (40000
// iterations, standard error 0.2 or less), at 120 / 180 / 300 sec (2026-10-12):
//   Human     873.7 / 871.6 / 874.7  (best overall and Alliance)
//   Undead    873.5 / 871.2 / 874.2  (best Horde)
//   Gnome     868.5 / 866.2 / 867.6
//   Night Elf 867.8 / 864.5 / 867.5
//   Dwarf     859.3 / 857.1 / 860.1
//   Troll     858.3 / 856.1 / 859.2
// Human's 2% crit with swords comes from the main hand, a sword like Azuresong Mageblade, the
// Era Phase 1 BiS. With a staff (Staff of Dominance) a Human ties a Dwarf at 836 and an Undead
// is 14 ahead of both. Undead's Touch of the Grave is 14, Gnome's Eureka! 9, Night Elf's Elune's
// Light 7. Troll's Berserking only speeds up Mind Blast's cast, channels and the GCD do not
// speed up, so it is worth nothing. Starshards and Chastise lose DPS (see the rotation).
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
export const PresetBuildLevel60Holy = PresetUtils.makePresetBuild('Level 60 + Holy', {
	group: 'Level 60',
	tooltip: 'Level 60 with the Holy sub tree talents, 24 DPS behind Level 60.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60Holy,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceHuman,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
