import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
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
	NaturePowerBuff,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SapperExplosive,
	SpellPowerBuff,
	SpiritElixir,
	Stat,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { ElementalShaman_Options as ElementalShamanOptions } from '../core/proto/shaman.js';
import { SavedTalents } from '../core/proto/ui.js';
import { Stats } from '../core/proto_utils/stats.js';
import DefaultAPLJson from './apls/default.apl.json';
import Level60APLJson from './apls/level60.apl.json';
import Level60ManaTideAPLJson from './apls/level60_mana_tide.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';
import Phase1GearJSON from './gear_sets/phase_1.gear.json';
import Phase2GearJSON from './gear_sets/phase_2.gear.json';
import Phase3GearJSON from './gear_sets/phase_3.gear.json';
import Phase4GearJSON from './gear_sets/phase_4.gear.json';
import Phase5GearJSON from './gear_sets/phase_5.gear.json';
import Phase6GearJSON from './gear_sets/phase_6.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearPhase1 = PresetUtils.makePresetGear('Phase 1', Phase1GearJSON, { group: 'Era' });
export const GearPhase2 = PresetUtils.makePresetGear('Phase 2', Phase2GearJSON, { group: 'Era' });
export const GearPhase3 = PresetUtils.makePresetGear('Phase 3', Phase3GearJSON, { group: 'Era' });
export const GearPhase4 = PresetUtils.makePresetGear('Phase 4', Phase4GearJSON, { group: 'Era' });
export const GearPhase5 = PresetUtils.makePresetGear('Phase 5', Phase5GearJSON, { group: 'Era' });
export const GearPhase6 = PresetUtils.makePresetGear('Phase 6', Phase6GearJSON, { group: 'Era' });

// A guess at Phase 1 BiS under Forever, made of the synthetic items in
// tools/database/forever_synthetic_items_shaman.go (ids from 990101, not in the game).
//
// Forever's level 60 gear in the client is not tuned yet, and Molten Core and Onyxia have no
// Forever data. The synthetic items take the place of the Era Phase 1 elemental pieces (the
// Earthfury set, Choker of the Fire Lord, Deep Earth Spaulders, Robe of Volatile Power, Mana
// Igniting Cord, Ring of Spell Power and Staff of Dominance), rebuilt on the cross-class
// standard (orchestration/standards.md): its item levels and budget, Forever's cheaper spell
// power, about 10% of the budget in Spirit, and hit to the spell cap against a level 63 boss
// with Tidal Focus 5 and the 2 piece bonus. The hands and legs carry The Spiritcaller's Storm,
// Forever's elemental Tier 1 set, so only its 2 piece bonus (+1% hit) is live. The 5 piece bonus
// (Lava Burst 1 sec sooner) would add more. The file has the budget math.
//
// The enchants are the best a Phase 1 shaman can get outside the raids: Enchant 2H Weapon -
// Spellblasting (Forever, +65 spell power), Gloves - Natural Power, Bracer - Greater Spellpower,
// Mystic Rugged Armor Kits on the chest, legs and feet, Lesser Arcanum of Voracity (Intellect)
// on the head and Forever's Enchant Necklace - Spell Power. Revelation on the staff was 1.3 ahead
// of the old Enchant Weapon - Spell Power (Molten Core only) and Spellblasting about 18, so
// Revelation is about 17 behind. The relic is Burning Totem (Flame Shock 3 sec longer),
// 7.1 ahead of no relic and 4.8 ahead of Polished Driftwood Icon.
//
// No trinkets: Forever removed the old dungeon trinkets (Briarwood Reed and Royal Seal of
// Eldre'Thalas aren't in the 70338 client), Talisman of Ephemeral Power is Molten Core only, and
// the sim doesn't model Forever's launch trinkets yet. Trinkets excluded until they are.
//
// On 2026-10-12 the Level 60 build is 693.3 / 690.3 / 680.4 at 120 / 180 / 300 sec on it. The
// set stays fixed, it was not gear searched.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

export const GearPresets = {
	[Phase.Phase1]: [GearLevel60Phase1Synthetic, GearPhase1],
	[Phase.Phase2]: [GearPhase2],
	[Phase.Phase3]: [GearPhase3],
	[Phase.Phase4]: [GearPhase4],
	[Phase.Phase5]: [GearPhase5],
	[Phase.Phase6]: [GearPhase6],
};

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLDefault = PresetUtils.makePresetAPLRotation('Default', DefaultAPLJson, { group: 'Era' });

// The tools/rotopt shaman_ele search on the Level 60 talents, over the mean of 120, 180 and
// 300 sec: Mana Spring and Searing Totem, Flame Shock whenever its DoT is down, Lava Burst on
// cooldown unless Flame Shock runs out before the cast ends (+1.3 to +1.8), Lightning Bolt, and
// Lightning Bolt rank 7 below 20% mana (+2 over the top rank). The fight opens with Mana Spring
// at -3.9 sec, a Lightning Bolt at -2.9 (it lands 0.1 sec into the fight) and Searing Totem at
// -0.9 (its first attack comes after the pull): +10.6 / +7.0 / +4.8 at 120 / 180 / 300 sec over
// putting the two totems down with no bolt. Chain Lightning loses on one target even above 80%
// mana (490 damage a second for 485 mana, against Lightning Bolt's 479 for 198). Earth Shock and
// mana floors on the shocks changed nothing. A caster stands at range, so the rotation has no
// Magma Totem or Fire Nova. tools/rotopt/shaman_presets.sh writes it.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Level 60', Level60APLJson, {
	tooltip: 'The tools/rotopt search result for the Level 60 talents.',
	group: 'Level 60',
});
// The same with Mana Tide Totem at 20% mana. At 40, 60 and 80% it was 1.1, 2.7 and 5.7 behind.
export const APLLevel60ManaTide = PresetUtils.makePresetAPLRotation('Level 60 + Mana Tide', Level60ManaTideAPLJson, {
	tooltip: 'The Level 60 rotation with Mana Tide Totem, for the Level 60 + Mana Tide talents.',
	group: 'Level 60',
});

export const APLPresets = {
	[Phase.Phase1]: [APLLevel60, APLLevel60ManaTide, APLDefault],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultAPL = APLLevel60;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// The talent search (tools/rotopt -talent-search) on the synthetic set, from four seeds (deep
// Elemental with Restoration, with Enhancement, and two mixes), then the knob search, then the
// talent search again from three seeds on the new knobs, as the mean of 120, 180 and 300 sec. It
// ran again on the standard gear and consumes from two seeds and kept the build. Thundering
// Strikes 5 and Tidal Focus 5 (5% hit, Forever took Nature's Guidance away) are the big ones. On
// the earlier gear the next best build moved 4 points into Ancestral Knowledge and Shamanistic
// Focus out of Mindfulness and Tidal Focus and was 15 behind.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '5504301300103051-052-05305' }), {
	tooltip: 'Level 60 talents from the talent search: Lava Burst, Thundering Strikes, Mindfulness and Tidal Focus.',
	group: 'Level 60',
});
// Mana Tide Totem for the group's mana: Totemic Focus 5, Mindfulness 3, Tidal Focus 5 and the
// Mana Tide row, with one point less in Thundering Strikes and none in Ancestral Knowledge. The
// talent search with Mana Tide Totem pinned found it, and found it again on the standard gear.
// Its Water Shield point does nothing in the sim: the elemental options have no raid damage
// hits, and the boss never hits a caster. It is 686.4 / 682.0 / 673.7 at 120 / 180 / 300 sec,
// 1.2% behind Level 60 at 180.
export const TalentsLevel60ManaTide = PresetUtils.makePresetTalents(
	'Level 60 + Mana Tide',
	SavedTalents.create({ talentsString: '5504301300103051-04-053150001001' }),
	{ tooltip: 'Level 60 talents with 16 Restoration points for Mana Tide Totem.', group: 'Level 60' },
);

export const TalentPresets = {
	[Phase.Phase1]: [TalentsLevel60, TalentsLevel60ManaTide],
	[Phase.Phase2]: [],
	[Phase.Phase3]: [],
	[Phase.Phase4]: [],
	[Phase.Phase5]: [],
	[Phase.Phase6]: [],
};

export const DefaultTalents = TalentsLevel60;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = ElementalShamanOptions.create({});

// The level 60 consumables, Forever's Phase 1 ones (the shared consumables branch). On 2026-10-12
// we tried each candidate in its slot on the Level 60 build, as the mean DPS at 120, 180 and 300
// sec. On the earlier gear: Elixir of Nature Power +12.9, Elixir of the Sages +10.4 (2% crit),
// Lobster Stew +2.9 over Nightfin Soup, Greater Mageblood Elixir +0.9, Goblin Sapper Charge +4.3,
// Goblin Land Mine +1.2, Demonic Rune 9.6, Flask of Supreme Power 71 over Flask of Distilled
// Wisdom, Cerebral Cortex Compound 3.4, Elixir of Firepower 1.6 and Wizard Oil 11.9. Major
// Spellblasting Potion loses to the Major Mana Potion (-6). A character takes one of the five 2%
// crit elixirs (the standard): the Sages with Elixir of Greater Intellect, a tie with the Owl
// and Elixir of Greater Spirit (0.4 behind). If Forever lets them stack, the Owl on top of the
// Sages would add about 10. Brilliant Wizard Oil would add 10.8, but its recipe is from Zul'Gurub.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	fillerExplosive: Explosive.ExplosiveGoblinLandMine,
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.WizardOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	naturePowerBuff: NaturePowerBuff.ElixirOfNaturePower,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The level 60 raid (tools/rotopt/mksettings -spec ele): the shared level 60 raid buffs, Kings
// and Wisdom from the paladins, Judgement of Wisdom, and Curse of the Elements as the one debuff
// on the shaman's schools. Under Forever Improved Scorch and Winter's Chill only help the mage's
// own spells, and Curse of Shadow is gone. The shaman is the only one in its group, so its
// totems are the ones the rotation puts down.
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

// The shaman is an engineer for the Goblin Sapper Charge and Goblin Land Mine (4.3 and 1.2)
// and an alchemist. The race is Orc, the best race (690.3 at 180 sec, Windshaper Skyborne 0.8
// behind), see the races under the builds.
export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: Race.RaceOrc,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor), a 3 minute fight,
// and the default buffs, debuffs and consumes, as for the enhancement shaman. The synthetic
// Phase 1 set, the talents and the rotation were searched at 120, 180 and 300 sec.
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

// On 2026-10-12, on the standard gear and consumes (20000 iterations), Level 60 is 693.3 /
// 690.3 / 680.4 at 120 / 180 / 300 sec and Level 60 + Mana Tide 686.4 / 682.0 / 673.7.
//
// The races (same settings and seed): Orc 693.3 / 690.3 / 680.4, Windshaper Skyborne 693.1 /
// 689.5 / 679.8, Troll 691.7 / 687.2 / 677.1, Dwarf (the only Alliance shaman) 687.3 / 683.7 /
// 674.4, Tauren 687.2 / 683.6 / 674.3. Orc is the best race and the best Horde race. The staff
// is the Era Phase 1 weapon type and no race has a staff racial; Tauren's 1% hit is wasted at
// the hit cap.
export const PresetBuildLevel60 = PresetUtils.makePresetBuild('Level 60', {
	group: 'Level 60',
	tooltip: 'Level 60 Orc on the synthetic Phase 1 gear with the Level 60 rotation, in a 3 minute raid fight.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60,
	rotation: APLLevel60,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
export const PresetBuildLevel60ManaTide = PresetUtils.makePresetBuild('Level 60 + Mana Tide', {
	group: 'Level 60',
	tooltip: 'Level 60 with the Mana Tide talents and the Mana Tide rotation.',
	gear: GearLevel60Phase1Synthetic,
	talents: TalentsLevel60ManaTide,
	rotation: APLLevel60ManaTide,
	encounter: EncounterLevel60,
	race: Race.RaceOrc,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
