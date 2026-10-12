import { Phase } from '../core/constants/other.js';
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
	SpellPowerBuff,
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
// Igniting Cord, Ring of Spell Power and Staff of Dominance), rebuilt at Phase 1 epic budgets
// with Forever's cheaper spell power and about 10% of the budget in Spirit. Six of them carry
// The Spiritcaller's Storm, Forever's elemental Tier 1 set (+1% hit, Lava Burst 1 sec sooner).
// The file has the budget math.
//
// The trinkets are real Phase 1 items: Talisman of Ephemeral Power and Briarwood Reed. The
// relic is Burning Totem (Flame Shock 3 sec longer). The enchants are the Phase 1 caster ones
// (Arcanum of Focus, Greater Stats, Greater Intellect, Enchant Weapon - Spell Power) and
// Forever's Enchant Necklace - Spell Power.
//
// GEAR_NUMBERS
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

// ROTATION_NUMBERS
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Level 60', Level60APLJson, {
	tooltip: 'The tools/rotopt search result for the Level 60 talents.',
	group: 'Level 60',
});
// MANA_TIDE_ROTATION_NUMBERS
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

// TALENT_NUMBERS
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '5504301300103051-052-05305' }), {
	tooltip: 'Level 60 talents from the talent search: Lava Burst, Thundering Strikes, Mindfulness and Tidal Focus.',
	group: 'Level 60',
});
// MANA_TIDE_TALENT_NUMBERS
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

// CONSUMES_NUMBERS
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	firePowerBuff: FirePowerBuff.ElixirOfFirepower,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodNightfinSoup,
	mainHandImbue: WeaponImbue.WizardOil,
	manaRegenElixir: ManaRegenElixir.MagebloodPotion,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
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

// PROFESSIONS_NUMBERS
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

// BUILD_NUMBERS
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
