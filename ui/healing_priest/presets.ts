import * as PresetUtils from '../core/preset_utils.js';
import {
	Conjured,
	Consumes,
	Debuffs,
	Flask,
	Food,
	IndividualBuffs,
	IntellectElixir,
	ManaRegenElixir,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	SpellPowerBuff,
	SpiritElixir,
	SunderRamp,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { HealingPriest_Options as HealingPriestOptions } from '../core/proto/priest.js';
import { SavedTalents } from '../core/proto/ui.js';
import DisciplineTankHealerAPLJson from './apls/discipline_tank_healer.apl.json';
import HolyRaidHealerAPLJson from './apls/holy_raid_healer.apl.json';
import HolyTankHealerAPLJson from './apls/holy_tank_healer.apl.json';
import Level60Phase1SyntheticGearJSON from './gear_sets/level60_phase1_synthetic.gear.json';

// The Holy and Discipline priest of the healing sim (docs/healing_priest.md). Its numbers come from
// tools/healbench, which scores a healer by the damage it carries in a role, as for the Restoration shaman.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

// The healing priest synthetic Phase 1 set (tools/database/forever_synthetic_items_priest_healing.go, ids
// 990501 to 990514), on the healing priest row of orchestration/gear_targets.md (standards 35): Wowhead's
// Classic healing priest pre-raid list averaged with the Forever Restoration shaman list, since the Forever
// priest guides list no items yet. Spell power and healing on top of it, Intellect, Spirit, crit and MP5, no
// hit. The Restoration shaman's enchants and trinkets, Spellblasting on the staff, a Holy wand.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Healing', Level60Phase1SyntheticGearJSON, {
	tooltip: 'A guess at Phase 1 BiS for a healer under Forever, made of synthetic items that are not in the game.',
	group: 'Synthetic',
});

export const GearPresets = [GearLevel60Phase1Synthetic];

export const DefaultGear = GearLevel60Phase1Synthetic;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

// All three cast Flash Heal 7 at whoever in the assignment drops under the emergency line, Power Infusion on
// the priest when a big heal is due, Inner Focus right before the biggest heal, potions and runes through
// the other cooldowns, and bind at most 4 ranks of a heal (the user's ruling), each cast when what's missing
// covers its Expected Healing.
//
// Holy tank: Prayer of Mending and Renew on the tank whenever they're off, Flash Heal 7 and 4 first while mana
// is over 70%, then Greater Heal 5, 4 and 2 and Flash Heal 7 and 4. On tools/healbench (tank role, damage
// carried) it carries 1.23 of Onyxia's damage (the shaman tank healer 1.02) and 0.89 of Broodlord's (0.62);
// over 300 s it OOMs and carries 0.96 (0.96). No shield: on the tank it costs more mana than it saves.
export const APLHolyTankHealer = PresetUtils.makePresetAPLRotation('Holy tank healer', HolyTankHealerAPLJson, {
	tooltip: 'Prayer of Mending and Renew on the tank, Flash Heal while mana lasts, then bound ranks of Greater Heal.',
	group: 'Healing sim',
});
// HOLY_RAID_NOTE
export const APLHolyRaidHealer = PresetUtils.makePresetAPLRotation('Holy raid healer', HolyRaidHealerAPLJson, {
	tooltip: 'Flash Heal, Prayer of Healing and Renew at the most hurt, Prayer of Mending, a shield for whoever drops low.',
	group: 'Healing sim',
});
// Discipline tank: Power Word: Shield whenever Weakened Soul is off the tank, Penance once what's missing covers
// its bolts, Renew, then the Holy tank healer's Flash Heal and Greater Heal. It carries 1.22 of Onyxia's damage
// and 0.88 of Broodlord's, under the Holy build's on both, and 0.70 over 300 s, since the shield's mana runs
// it dry by 150 s. Shielding only when mana is full or the tank is hurt carries 0.82 over 300 s for 1.15 and 0.83.
export const APLDisciplineTankHealer = PresetUtils.makePresetAPLRotation('Discipline tank healer', DisciplineTankHealerAPLJson, {
	tooltip: 'Power Word: Shield, Penance and Renew on the tank, then bound ranks of Greater Heal and Flash Heal.',
	group: 'Healing sim',
});

export const APLPresets = [APLHolyTankHealer, APLHolyRaidHealer, APLDisciplineTankHealer];

export const DefaultAPL = APLHolyTankHealer;

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Holy (18 Discipline, 33 Holy): the tank healer carries the most with it. Discipline (Power
// Infusion, Divine Aegis, Renewed Hope, Penance) carries less on every fight we bench, and a Penance build with
// 30 points in Holy sits between the two (docs/healing_priest.md).
export const TalentsHoly = PresetUtils.makePresetTalents('Holy', SavedTalents.create({ talentsString: '00500503113-03505103030121531' }), {
	tooltip: 'Prayer of Mending, Spiritual Guidance and Healing, Holy Binding, Litany of Light, with Inner Focus and Meditation.',
	group: 'Healing sim',
});
export const TalentsDiscipline = PresetUtils.makePresetTalents('Discipline', SavedTalents.create({ talentsString: '005005031330011531-0350510303' }), {
	tooltip: 'Power Infusion, Divine Aegis, Renewed Hope, Penance and Soul Warding, with Inspiration and Improved Healing.',
	group: 'Healing sim',
});

export const TalentPresets = [TalentsHoly, TalentsDiscipline];

export const DefaultTalents = TalentsHoly;

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

// The rotation casts Power Infusion (on the priest) and Inner Focus; Inner Fire isn't simulated.
export const DefaultOptions = HealingPriestOptions.create({});

// The Restoration shaman's level 60 consumables: a Brilliant Mana Oil and the Major Mana Potion.
export const DefaultConsumes = Consumes.create({
	defaultConjured: Conjured.ConjuredDemonicRune,
	defaultPotion: Potions.MajorManaPotion,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodLobsterStew,
	intellectElixir: IntellectElixir.ElixirOfGreaterIntellect,
	mainHandImbue: WeaponImbue.BrilliantManaOil,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	spiritElixir: SpiritElixir.ElixirOfSages,
	zanzaBuff: ZanzaBuff.CerebralCortexCompound,
});

// The shared level 60 raid buffs and debuffs, as the Restoration shaman's.
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
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
});

export const DefaultDebuffs = Debuffs.create({
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
	sunderRamp: SunderRamp.create({ tank: true, pullSunders: 1 }),
	thunderClap: TristateEffect.TristateEffectImproved,
});

// The races are within 0.02 of each other on the Holy tank healer (Onyxia and Broodlord, 300 iterations), so
// the shaman's Alliance pick and the Horde caster's: Human (10% Spirit) and Undead.
export const AllianceRace = Race.RaceHuman;
export const HordeRace = Race.RaceUndead;

export const OtherDefaults = {
	distanceFromTarget: 20,
	channelClipDelay: 50,
	profession1: Profession.Engineering,
	profession2: Profession.Alchemy,
	race: AllianceRace,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

export const PresetBuildHolyTankHealer = PresetUtils.makePresetBuild('Holy tank healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Human on the healing priest synthetic set, Holy talents, the Holy tank healer rotation.',
	race: AllianceRace,
	level: 60,
	gear: DefaultGear,
	talents: TalentsHoly,
	rotation: APLHolyTankHealer,
});

export const PresetBuildHolyRaidHealer = PresetUtils.makePresetBuild('Holy raid healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Human on the healing priest synthetic set, Holy talents, the Holy raid healer rotation.',
	race: AllianceRace,
	level: 60,
	gear: DefaultGear,
	talents: TalentsHoly,
	rotation: APLHolyRaidHealer,
});

export const PresetBuildDisciplineTankHealer = PresetUtils.makePresetBuild('Discipline tank healer', {
	group: 'Healing sim',
	tooltip: 'Level 60 Human on the healing priest synthetic set, Discipline talents, the Discipline tank healer rotation.',
	race: AllianceRace,
	level: 60,
	gear: DefaultGear,
	talents: TalentsDiscipline,
	rotation: APLDisciplineTankHealer,
});
