// What each member of the comp editor's raid can give, and the icons the page draws for it.
//
// The raid's effects are the resolver's gifts (sim/compsearch/effects.go), by the names its buff records
// use. Before a run we only know who could give each one: a build's `gives` and its duty variants, or a
// placeholder's preset. After a run the resolver's record says what was up.
import { Blessing, Duty, PaladinJudgement, RaidPackage, RaidPreset, RaidTotem } from '../core/proto/api';
import { Class } from '../core/proto/common';
import { CLASS_ICONS, RaidPresetInfo, ReportBuild } from './data';

export const wowIcon = (name: string) => `https://wow.zamimg.com/images/wow/icons/large/${name}.jpg`;

export interface RaidEffect {
	// The resolver's name for it, as its buff records and the builds' `gives` write it.
	name: string;
	// The name the page shows, when it differs.
	label?: string;
	icon: string;
	debuff?: boolean;
	// Who can give it, for the Who gives what table.
	from: string;
}

// The raid's effects in the order the page's strip shows them: the long buffs, the held auras and totems,
// then the boss's debuffs.
export const RAID_EFFECTS: RaidEffect[] = [
	{ name: 'Fortitude', label: 'Prayer of Fortitude', from: 'Priest: Discipline or Holy healer, Shadow or Smite', icon: wowIcon('spell_holy_prayeroffortitude') },
	{ name: 'Gift of the Wild', from: 'Druid: Feral tank, Restoration healer, Balance or Feral', icon: wowIcon('spell_nature_regeneration') },
	{ name: 'Arcane Brilliance', from: 'Mage', icon: wowIcon('spell_holy_arcaneintellect') },
	{ name: 'Prayer of Spirit', from: 'Priest: Discipline or Holy healer, Shadow or Smite', icon: wowIcon('spell_holy_prayerofspirit') },
	{ name: 'Blessings', label: 'Greater Blessings', from: 'Paladin: Holy or Protection, or Retribution. One blessing per class per paladin', icon: wowIcon('spell_magic_greaterblessingofkings') },
	// Forever's Salvation (spell 1038) is no longer a Blessing: any paladin gives it, and it lowers threat only.
	{ name: 'Salvation', from: 'Any paladin', icon: wowIcon('spell_holy_sealofsalvation') },
	{ name: 'Battle Shout', from: 'Protection warrior tank, or a warrior on its duty', icon: wowIcon('ability_warrior_battleshout') },
	{ name: 'Furious Howl', from: 'Hunter with a wolf', icon: wowIcon('ability_hunter_pet_wolf') },
	{ name: 'Trueshot Aura', from: 'Marksmanship hunter', icon: wowIcon('ability_trueshot') },
	{ name: 'Leader of the Pack', from: 'Feral druid: tank or cat', icon: wowIcon('spell_nature_unyeildingstamina') },
	{ name: 'Moonkin Aura', from: 'Balance druid', icon: wowIcon('spell_nature_moonglow') },
	{ name: 'Strength of Earth', label: 'Strength of Earth Totem', from: 'Shaman: Restoration or Warden, or Enhancement', icon: wowIcon('spell_nature_earthbindtotem') },
	{ name: 'Grace of Air', label: 'Grace of Air Totem', from: 'Restoration or Warden shaman', icon: wowIcon('spell_nature_invisibilitytotem') },
	{ name: 'Windfury Totem', from: 'Shaman: Restoration or Warden, or Enhancement', icon: wowIcon('spell_nature_windfury') },
	{ name: 'Flametongue Totem', from: 'Restoration or Warden shaman', icon: wowIcon('spell_nature_guardianward') },
	{ name: 'Sunder upkeep', label: 'Sunder Armor', from: 'Protection warrior tank, or a warrior on its duty', icon: wowIcon('ability_warrior_sunder'), debuff: true },
	{ name: 'Minor armor slot', label: 'Faerie Fire or Curse of Recklessness', from: 'Druid (Faerie Fire), a warlock on its duty, or a hunter with a Tallstrider', icon: wowIcon('spell_nature_faeriefire'), debuff: true },
	{ name: 'Curse of the Elements', from: 'Warlock on its duty', icon: wowIcon('spell_shadow_chilltouch'), debuff: true },
	{ name: 'Judgement of the Crusader', from: 'Paladin: Holy or Protection, or Retribution on its duty', icon: wowIcon('spell_holy_holysmite'), debuff: true },
	{ name: 'Judgement of Wisdom', from: 'Paladin: Holy or Protection, or Retribution on its duty', icon: wowIcon('spell_holy_righteousnessaura'), debuff: true },
];

// What each placeholder gives, by the resolver's names. A shaman gives one totem of each element, so
// Grace of Air and Windfury are either one.
const PRESET_GIVES: Partial<Record<RaidPreset, string[]>> = {
	[RaidPreset.RaidPresetProtectionWarrior]: ['Sunder upkeep', 'Battle Shout'],
	[RaidPreset.RaidPresetFeralTank]: ['Leader of the Pack', 'Gift of the Wild', 'Minor armor slot'],
	[RaidPreset.RaidPresetProtectionPaladin]: ['Blessings', 'Salvation', 'Judgement of the Crusader', 'Judgement of Wisdom'],
	[RaidPreset.RaidPresetHolyPaladin]: ['Blessings', 'Salvation', 'Judgement of the Crusader', 'Judgement of Wisdom'],
	[RaidPreset.RaidPresetWardenShaman]: ['Strength of Earth', 'Grace of Air', 'Windfury Totem', 'Flametongue Totem'],
	[RaidPreset.RaidPresetRestorationShaman]: ['Strength of Earth', 'Grace of Air', 'Windfury Totem', 'Flametongue Totem'],
	[RaidPreset.RaidPresetDisciplinePriest]: ['Fortitude', 'Prayer of Spirit'],
	[RaidPreset.RaidPresetHolyPriest]: ['Fortitude', 'Prayer of Spirit'],
	[RaidPreset.RaidPresetRestorationDruid]: ['Gift of the Wild', 'Minor armor slot'],
};

// What a duty gives the raid (sim/compsearch, dutyGift).
export const DUTY_EFFECT: Partial<Record<Duty, string>> = {
	[Duty.DutyBattleShout]: 'Battle Shout',
	[Duty.DutySunderUpkeep]: 'Sunder upkeep',
	[Duty.DutyFaerieFire]: 'Minor armor slot',
	[Duty.DutyCurseOfRecklessness]: 'Minor armor slot',
	[Duty.DutyCurseOfTheElements]: 'Curse of the Elements',
	[Duty.DutyJudgementOfTheCrusader]: 'Judgement of the Crusader',
	[Duty.DutyJudgementOfWisdom]: 'Judgement of Wisdom',
};

// The effects a member could give: a placeholder's, or a build's own and those of its duty variants.
export const effectsOf = (preset: RaidPreset | undefined, build: ReportBuild | undefined): string[] => {
	if (preset) return PRESET_GIVES[preset] ?? [];
	if (!build) return [];
	const own = build.gives == 'nothing' ? [] : build.gives.split(',').map(s => s.trim());
	// A paladin's Salvation comes with its Blessings.
	if (own.includes('Blessings')) own.push('Salvation');
	const duties = build.duties.map(d => (d.duty !== undefined ? DUTY_EFFECT[d.duty] : undefined)).filter((e): e is string => !!e);
	return [...new Set([...own, ...duties])];
};

// A party's mana package (Mana Spring, and Mana Tide), as the resolver records it.
export const PACKAGE_ICON = wowIcon('spell_nature_manaregentotem');
export const MANA_TIDE_ICON = wowIcon('spell_frost_summonwaterelemental');

export const BLESSING_ICONS: Partial<Record<Blessing, string>> = {
	[Blessing.BlessingKings]: wowIcon('spell_magic_greaterblessingofkings'),
	[Blessing.BlessingMight]: wowIcon('spell_holy_greaterblessingofkings'),
	[Blessing.BlessingWisdom]: wowIcon('spell_holy_greaterblessingofwisdom'),
};

export const JUDGEMENT_ICONS: Partial<Record<PaladinJudgement, string>> = {
	[PaladinJudgement.PaladinJudgementCrusader]: wowIcon('spell_holy_holysmite'),
	[PaladinJudgement.PaladinJudgementWisdom]: wowIcon('spell_holy_righteousnessaura'),
};

export const TOTEM_ICONS: Partial<Record<RaidTotem, string>> = {
	[RaidTotem.RaidTotemStrengthOfEarth]: wowIcon('spell_nature_earthbindtotem'),
	[RaidTotem.RaidTotemGraceOfAir]: wowIcon('spell_nature_invisibilitytotem'),
	[RaidTotem.RaidTotemWindfury]: wowIcon('spell_nature_windfury'),
	[RaidTotem.RaidTotemManaSpring]: wowIcon('spell_nature_manaregentotem'),
	[RaidTotem.RaidTotemFlametongue]: wowIcon('spell_nature_guardianward'),
};

export const DUTY_ICONS: Partial<Record<Duty, string>> = {
	[Duty.DutyBattleShout]: wowIcon('ability_warrior_battleshout'),
	[Duty.DutySunderUpkeep]: wowIcon('ability_warrior_sunder'),
	[Duty.DutyFaerieFire]: wowIcon('spell_nature_faeriefire'),
	[Duty.DutyCurseOfRecklessness]: wowIcon('spell_shadow_unholystrength'),
	[Duty.DutyCurseOfTheElements]: wowIcon('spell_shadow_chilltouch'),
	[Duty.DutyJudgementOfTheCrusader]: wowIcon('spell_holy_holysmite'),
	[Duty.DutyJudgementOfWisdom]: wowIcon('spell_holy_righteousnessaura'),
};

export const POWER_INFUSION_ICON = wowIcon('spell_holy_powerinfusion');
export const INNERVATE_ICON = wowIcon('spell_nature_lightning');

// Each build's spec icon, by its class and the report's spec name. A spec the table lacks takes its
// class's icon.
const SPEC_ICONS: Record<string, string> = {
	'Druid:Balance': 'spell_nature_starfall',
	'Druid:Feral': 'ability_druid_catform',
	'Hunter:Beast Mastery': 'ability_hunter_beasttaming',
	'Hunter:Marksmanship': 'ability_marksmanship',
	'Hunter:Survival': 'ability_hunter_swiftstrike',
	'Mage:Arcane': 'spell_holy_magicalsentry',
	'Mage:Fire': 'spell_fire_firebolt02',
	'Mage:Frost': 'spell_frost_frostbolt02',
	'Mage:Frostfire': 'ability_mage_frostfirebolt',
	'Paladin:Retribution': 'spell_holy_auraoflight',
	'Priest:Shadow': 'spell_shadow_shadowwordpain',
	'Priest:Smite': 'spell_holy_holysmite',
	'Rogue:Assassination': 'ability_rogue_eviscerate',
	'Rogue:Combat': 'ability_backstab',
	'Rogue:Subtlety': 'ability_stealth',
	'Shaman:Elemental': 'spell_nature_lightning',
	'Shaman:Enhancement': 'ability_shaman_stormstrike',
	'Warlock:Affliction': 'spell_shadow_deathcoil',
	'Warlock:Demonology': 'spell_shadow_metamorphosis',
	'Warlock:Destruction': 'spell_shadow_rainoffire',
	'Warrior:Arms': 'ability_warrior_savageblow',
	'Warrior:Fury': 'ability_warrior_innerrage',
};

export const buildIcon = (build: ReportBuild): string => {
	const name = SPEC_ICONS[`${build.className}:${build.spec}`];
	return name ? wowIcon(name) : CLASS_ICONS[build.class] ?? '';
};

// The icon of a member: its preset's, or its build's spec.
export const memberIcon = (info: RaidPresetInfo | undefined, build: ReportBuild | undefined, cls: Class): string =>
	info?.icon ?? (build ? buildIcon(build) : CLASS_ICONS[cls] ?? '');

// The party package's words and whether it has Mana Tide.
export const packageHasTide = (p: RaidPackage) => p == RaidPackage.RaidPackageManaSpringTide || p == RaidPackage.RaidPackageRestoration;
