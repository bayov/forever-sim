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
// Mana Igniting Cord, the Molten Core rings) by the cross-class standard
// (orchestration/standards.md): one item level table (66 for the tier slots, 71 for neck, back,
// wrist, waist, rings and wand) on Forever's caster budget (standards 1b: spell power at 0.86,
// each armor piece's linear total set by its slot and item level, four stat lines a piece by the
// 1.7-norm shape check), in the fair stat mix from Season of Discovery's level 60 shadow priest
// BiS (standards 5a: 16% Stamina, 15% Intellect, 10% Spirit, 37% spell power, 11% hit, 10% crit,
// the hit going to the cap first; the file has the table and every piece's numbers). Totals:
// armor 710.9, the staff 194.2 (Forever's caster staff rate, item level 76), the wand 21.2. The
// staff with Forever's Enchant 2H Weapon - Spellblasting (+65 spell power) is 28.8 ahead of the
// sword and off hand (149.1 + 43.4) with Mighty Intellect at 180 sec. The set has 398 spell
// power with the staff, 7% crit from items and 10% hit, which with Shadow Focus and the 2 piece
// bonus is the 16% a level 63 boss needs. Only the head and shoulders carry Raiments of
// Conviction, Forever's Tier 1 shadow set, so only its 2 piece (1% hit) is live. Its 4 and 5
// piece (21 spell power against Undead, Devouring Plague without a cooldown) would add 40.2 DPS
// at 180 sec (4.7%).
//
// The trinkets are synthetic (standards 1b and 8a): Forever has none of the old dungeon
// trinkets, Talisman of Ephemeral Power is a Molten Core drop, and the sim does not have
// Forever's launch trinkets. A stat trinket (36 spell power, 15 crit rating, 45.9 linear) and an
// on-use trinket (140 spell power for 20 sec every 2 min on the shared trinket cooldown, about
// 27 weighted points over 180 sec), the same pair for both specs. The bundled gear round
// (2026-10-13: the 1b budget and the trinkets) moved the Level 60 build from 844.8 to 853.7 at
// 180 sec.
//
// The enchants are the best that Phase 1 has outside the raids (7.5 ahead of the earlier
// picks): Lesser Arcanum of Voracity (Intellect) on the head, Forever's Enchant Necklace - Spell
// Power, the Mystic Rugged Armor Kit (6 spell power) on chest, hands, legs and feet, Enchant
// Bracer - Greater Spellpower (16) and Spellblasting on the staff. Forever's Revelation is not
// used: its proc model is a guess made for shaman shocks. No Phase 1 shoulder or cloak enchant
// does anything for a caster outside the raids.
export const GearLevel60Phase1Synthetic = PresetUtils.makePresetGear('Level 60 Phase 1 Synthetic', Level60Phase1SyntheticGear, {
	tooltip: 'A guess at Phase 1 BiS under Forever, made of synthetic items that are not in the game.',
	group: 'Level 60',
});

// The same set for the Smite build without hit rating: Holy Precision gives Smite, Holy Fire and
// Penance 18%, so it has its own pieces in the Smite stat mix (SoD's shadow proportions without
// hit: 16.6% Stamina, 17.8% Intellect, 10% Spirit, 43.5% spell power, 12.2% crit; armor 708.2,
// staff 194.0, wand 21.5) and the same two trinkets. The staff is 29.8 ahead of the sword and
// off hand, and the bundled gear round moved Smite from 711.1 to 708.1 at 180 sec. Its hands
// take Forever's Enchant Gloves - Holy Power (+20 Holy damage, 5.1 ahead of the armor kit). Its
// head and shoulders carry Vestments of Conviction, Forever's Tier 1 Discipline and Holy set, so
// only its 2 piece (10 Spirit) is live; the 4 and 5 piece (9 spell damage, Penance 1 sec
// shorter) would add 4.7.
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
// 3 sec before the pull and a Mind Blast 1.5 sec before it that lands at 0 (+7.4), Shadow Word:
// Pain and Devouring Plague whenever they are down and off cooldown, Inner Focus on Mind Blast, Mind Blast and
// Shadow Word: Death on cooldown, Mind Flay to fill. All three lengths agree on it.
// - Mind Flay is not clipped for Mind Blast. Mind Blast's 5.5 sec cooldown against back to back
//   3 sec channels rarely comes up mid channel, so clipping gives up ticks for half a Mind Blast
//   more a fight: 10.6 behind after the first tick at 180 sec.
//   At 120 sec clipping it after its second tick is 2.5 ahead; the rotation keeps the 180 sec
//   answer, which 300 sec agrees with.
// - Mind Flay is clipped after its second tick when Shadow Word: Pain or Devouring Plague fell
//   off during it, so the damage over time spell goes back up sooner (+3.6 at 180 sec). Without
//   it a 3 sec channel left either one down for up to 3 sec.
// - Vampiric Embrace is left out. It heals the party and deals no damage, and keeping it up costs
//   17.6 DPS at 180 sec in GCDs. A raid that wants its healing can turn the knob on.
// - Shadow Word: Death above Mind Blast since client 70425 (18 sec cooldown, 200% crits in
//   Shadowform): +1.5 / +1.8 / +1.7 at 120 / 180 / 300 sec. Inner Focus still waits for Mind
//   Blast.
// - Starshards (Night Elf) and Chastise (Dwarf) are left out. In Shadowform a Starshards tick
//   hits no harder than a Mind Flay tick and the 6 sec channel holds up Mind Blast, 30 DPS
//   lost at any length. Chastise loses 4.7 to the Mind Flay GCD it takes.
// - Dark Sacrifice (Undead) under 10% mana. Level 60 never gets that low (about 13 min to
//   empty); Shadow + Holy does since 70425, and casts it only when it must (+2.9 / +5.8 / +2.1
//   against 40%).
// - Troll: Shadowguard 3 sec before the pull, with Shadowform at -4.5 sec for it, and again
//   before Mind Flay once its 3 charges are spent with 60 sec or more left and above 40% mana.
//   The raid's 2 hits a minute fire its strikes at the boss: +10.6 / +8.9 / +7.2 at 120 / 180
//   / 300 sec for a Troll (2026-10-23). The other races only warn on its lines.
// - The on-use trinket and the Major Spellblasting Potion go at -0.1 sec so the pre-cast Mind
//   Blast landing at 0 gets them (rotation auditor, 2026-10-13): +1.0 / +0.7 / +0.4 at 120 / 180
//   / 300 sec, the same for Shadow + Holy.
//
// Client 70425 (2026-10-16): Devouring Plague's cooldown is 24 sec, as long as its damage over
// time, so it stays up all fight. The knob search on the new talents, with the raid hits,
// keeps the rest of the rotation: 921.4 / 924.9 / 921.9 at 120 / 180 / 300 sec as an Undead.
// Inner Focus right before Devouring Plague (now free under it) is no better.
//
// A hand audit of the timeline on the final build (2026-10-12): Shadowform holds all fight,
// Shadow Weaving reaches 5 stacks 6 sec in and stays there, Inner Focus goes on the first Mind
// Blast and again after its 3 min, Shadow Word: Pain and Devouring Plague go back up within a GCD
// of falling off, no GCD is idle, and mana stays above 6000.
//
// Wand Shoot (2026-10-23, forever-wiki/sim-status.md): from 10% mana to 20% with a minute or
// more to go, and when nothing else can be cast. The Shadow build never runs that low; Shadow +
// Holy goes from 916.0 / 918.9 / 829.7 / 711.6 to 916.0 / 918.9 / 859.3 / 786.4 at 120 / 180 /
// 300 / 480 sec.
export const APLLevel60 = PresetUtils.makePresetAPLRotation('Level 60', Level60APL, {
	tooltip: "The optimizer's best level 60 rotation: both damage over time spells up, Mind Blast and Shadow Word: Death on cooldown, Mind Flay to fill.",
	group: 'Level 60',
});

// The tools/rotopt priest_smite template after a knob search on the Smite talents (3000
// iterations a candidate, 20000 to confirm, at 120, 180 and 300 sec, 2026-10-12): Power
// Infusion on cooldown after a Smite 2 sec before the pull that lands at 0 (+6.0; the on-use
// trinket at -0.1 sec so it lands with the trinket up, +1.0 / +0.7 / +0.4), Holy Fire
// whenever its damage over time is down (Power in Light gives
// Smite 10% while it is up), Shadow Word: Pain, Devouring Plague above 60% mana, Chastise for a
// Dwarf, Inner Focus on cooldown for the next Smite, then rank 8 Smite while the mana lasts the fight
// at 35 mana a second and rank 7 Smite after. Forever's downranking penalty (Blizzard's
// development notes, 2026-10-01; foreverchanges.pro/downranking's measured size, see notes.md)
// leaves rank 7 (level 46) its full coefficient at 60 and the earlier rank 4 fallback none: the
// old rotation would fall to 701.4 / 545.0 at 180 / 300 sec, this one is 706.5 / 643.8.
// The search drops the spells that give the least damage for their mana: Mind Blast, Shadow
// Word: Death and even Penance (385 mana for three 92 damage bolts). Shadow Word: Death comes
// back once, when a Smite would no longer land before the fight ends: 725.5 / 716.5 / 645.7
// against 719.7 / 713.5 / 644.5 at 120 / 180 / 300 sec (rotation audit 19b, 2026-10-15). A timeline shows Smite
// every 2 sec (Divine Fury) and Holy Fire back up within a cast of falling off.
//
// Client 70425 (2026-10-16) cut Devouring Plague's cooldown to 24 sec, which would spend
// Smite's mana: the 40% floor fell to 733.2 / 709.5 / 617.9 at 120 / 180 / 300 sec. Above 60%
// with Inner Focus on cooldown it is 727.8 / 714.6 / 642.2 on the old talents; at 300 sec
// leaving it out is better still (+16), at 120 sec the 40% floor (+5). Devouring Plague is
// free under Inner Focus (its fourth class mask word), but Inner Focus on it is behind.
//
// Power in Light also counts Shadow Word: Pain since 70425 (update ledger PRI-70425-03), so
// Holy Fire's damage over time is no longer needed for it, and its mana goes to Smite:
// 744.0 / 724.6 / 639.7 at 120 / 180 / 300 sec against 729.3 / 716.8 / 646.6 with it, with
// Devouring Plague now above 80% mana (2026-10-16). The notes above are from before.
//
// Wand Shoot (2026-10-23, forever-wiki/sim-status.md): from 10% mana to 20% with a minute or
// more to go, and when nothing else can be cast: 785.6 / 765.4 / 675.3 / 580.4 to 785.7 /
// 765.7 / 689.6 / 619.4 at 120 / 180 / 300 / 480 sec.
export const APLLevel60Smite = PresetUtils.makePresetAPLRotation('Level 60 Smite', Level60SmiteAPL, {
	tooltip: "The optimizer's best rotation for the Smite talents: Shadow Word: Pain up for Power in Light, Smite to fill, a lower rank once mana runs short.",
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
// with 20 Shadow. On 2026-10-13 (the bundled gear round, searches rerun) the build is 853.5 / 854.4 / 851.7 at
// 120 / 180 / 300 sec as an Undead.
//
// Client 70425 (2026-10-16) removed Shadow Affinity, added Dark Inspiration, moved Mental
// Strength up and cut it to 2% a rank, made Improved Mind Flay grow tick by tick, widened Twin
// Disciplines to Mind Flay and Shadow Word: Death, and the raid hits made pushback count. The
// search from the carried over build (Shadow Affinity's points in Dark Inspiration) trades
// Holy Precision (hit for Holy spells only) and Martyrdom for Twilight Focus 3 (70% chance to
// ignore pushback on Mind Blast and Mind Flay) and Early Demise 2: 914.1 / 917.5 / 914.6 to
// 919.9 / 923.1 / 920.2 as an Undead, 924.6 / 928.6 / 926.0 as a Human. The notes below are
// from before 70425.
// - Shadow 31: Shadow Focus 5, Improved Shadow Word: Pain, Improved Mind Blast 5, Mind Flay
//   and Improved Mind Flay 2, Shadow Weaving 3, Devouring Contagion 2, Darkness 5, Shadowform.
//   Shadow Affinity and a point of Shadow Reach only fill the rows (threat and range).
// - Discipline 20: Twin Disciplines 5 (13 DPS), Mental Strength 5, Inner Focus (about 1).
//   These were measured on the first synthetic set (more spell power), the searches on the
//   final set end on the same build.
//   Mental Agility, Meditation and the row 2 points fill the rows: mana never runs out.
// - Early Demise 2 in place of 2 Mental Strength is 0.7 behind.
// - The Power Infusion build loses Shadowform and is out of mana 107 sec in: 601 at 180 sec.
export const TalentsLevel60 = PresetUtils.makePresetTalents('Level 60', SavedTalents.create({ talentsString: '00500501103-3-50052101231302251' }), {
	tooltip: 'Level 60 talents from the talent search: deep Shadow with Twin Disciplines, Mental Strength, Inner Focus and Twilight Focus.',
	group: 'Level 60',
});

// Client 70425 (2026-10-16): the Shadow points follow Level 60's (Early Demise 2 from Holy Reach,
// which only adds range), so it keeps 18 Holy points: 874.8 / 878.7 / 863.5 to 878.1 / 881.7 /
// 866.5 as an Undead, 882.0 / 886.4 / 802.9 as a Human, 4.5% behind Level 60 at 180 sec. An
// unconstrained search leaves Holy for the Level 60 build (923.4). The notes below are from
// before 70425.
// The same Shadow points with 20 in Holy instead of Discipline (Twilight Focus, Spell Warding,
// Inspiration, Improved Healing and the rest of the healing utility a raid may ask for). None of
// the Holy talents adds Shadow damage, so it gives up Twin Disciplines and Mental Strength:
// 834.8 / 832.7 / 830.9 at 120 / 180 / 300 sec as an Undead on its own consumes (see the build),
// 2.5% behind at 180 sec. A talent search from it trades Holy points for Twin Disciplines and
// Shadow (+9.8 with 13 Holy), which defeats the build's purpose, so it keeps its 20 Holy. With
// Dark Sacrifice left out of the rotation it would gain 5.1 (837.1).
export const TalentsLevel60Holy = PresetUtils.makePresetTalents('Level 60 + Holy', SavedTalents.create({ talentsString: '-3025011303-50052101231302251' }), {
	tooltip: 'The Level 60 Shadow points with 18 Holy points for healing utility, about 42 DPS behind.',
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
// On its own set with its own consumes (see the build) it is 712.4 / 706.5 / 643.8 at 120 / 180
// / 300 sec as an Undead, whose Dark Sacrifice comes when the mana runs low. Client 70425 (2026-10-16,
// with the raid hits): Twilight Focus 3 in place of Inspiration (70% chance to ignore pushback
// on Smite and Holy Fire), 712.7 / 700.9 / 637.2 to 723.2 / 710.6 / 641.0, 23% behind Shadow
// at 180 sec. That is 17% behind
// Shadow at 180 sec and 24% at 300: Smite is a mana hungry filler and nothing in the tree halves its cost the way
// Shadowform does for Shadow spells.
export const TalentsLevel60Smite = PresetUtils.makePresetTalents('Level 60 Smite', SavedTalents.create({ talentsString: '525035231030001001-30505110002-3' }), {
	tooltip: 'Holy damage talents: 31 Discipline (Power in Light, Penance, Power Infusion) and 17 Holy (Divine Fury, Searing Light, Twilight Focus).',
	group: 'Level 60',
});

export const TalentsP1Shadow = PresetUtils.makePresetTalents('Shadow', SavedTalents.create({ talentsString: '00503123103--50052101231302251' }));

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
//   Greater Intellect only +2.8 (Arcane Brilliance is on). A character takes one of the five 2%
//   crit elixirs (Mongoose, Grizzly, Owl, Sages, Whale); the Elixir of Sages on top of the Owl
//   would add 13.9. Spirit adds nothing, so the Spirit elixir slot stays empty. Greater Arcane Elixir is worth 16.0, Elixir of Shadow Power 18.3.
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
	// The fight is in a natural flask zone (standards 11), but the best natural flask, Aggression, is 54
	// / 52 / 69 behind the Flask of Supreme Power on Shadow / Shadow + Holy / Smite at 180 sec.
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
	shadowProtection: true,
	thorns: TristateEffect.TristateEffectImproved,
	trueshotAura: true,
	salvation: true,
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
	race: Race.RaceUndead,
};

///////////////////////////////////////////////////////////////////////////
//                                 Builds
///////////////////////////////////////////////////////////////////////////

// The level 60 raid setup: the sim's default boss (level 63, 3731 armor, type Unknown: a generic boss), a 3 minute
// fight, and the default buffs, debuffs and consumes, like the enhancement shaman's.
export const EncounterLevel60 = PresetUtils.makePresetEncounter(
	'Level 60',
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
		debuffs: DefaultDebuffs,
		buffs: DefaultIndividualBuffs,
		consumes: DefaultConsumes,
	},
);

// Every race a priest can be under Forever, on the Level 60 build with the same seed (20000
// iterations, standard error 0.2 or less), at 120 / 180 / 300 sec (2026-10-16, client 70425,
// with the raid hits):
//   Human     924.6 / 928.6 / 926.0  (best overall and Alliance)
//   Gnome     921.1 / 927.2 / 923.5
//   Undead    921.4 / 924.9 / 921.9  (best Horde)
//   Dwarf     919.6 / 923.5 / 920.8
//   Night Elf 918.3 / 920.8 / 917.8
//   Troll     909.5 / 913.4 / 910.7
// Client 70425 made Human's Ruthless 2% crit with any weapon (and spell), which puts Human ahead
// of Undead's Touch of the Grave. Troll's Berserking only speeds up Mind Blast's cast, channels
// and the GCD do not speed up, so it is worth nothing. Starshards and Chastise lose DPS (see the
// rotation).
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
// Without Twin Disciplines and Mental Strength the Holy build runs low on a long fight, so it
// adds the Greater Mageblood Elixir. With it the Major Spellblasting Potion beat the Major Mana
// Potion at every length on the earlier gear (rotation audit, 2026-10-12). Every race on it
// (2026-10-16, client 70425, with the raid hits, 120 / 180 / 300 sec):
//   Human     882.0 / 886.4 / 802.9  (best at 180 sec, and best Alliance)
//   Gnome     878.6 / 884.9 / 822.2
//   Undead    879.3 / 882.9 / 865.1  (best Horde, and best at 300 sec)
//   Dwarf     876.7 / 881.0 / 797.4
//   Night Elf 875.8 / 878.9 / 796.3
//   Troll     867.5 / 871.8 / 788.1
// Devouring Plague every 24 sec (client 70425) empties the mana before 300 sec. Undead's Dark
// Sacrifice keeps it going (865.1 at 300 sec); every other race runs dry. The preset follows
// the 180 sec fight like the others and takes the Human.
export const ConsumesHoly = Consumes.create({
	...DefaultConsumes,
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
	tooltip:
		'Level 60 with the Holy sub tree talents and Greater Mageblood, 42 DPS behind Level 60. As a Human it runs out of mana by 300 sec (802.9; an Undead holds 865.1 with Dark Sacrifice).',
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
// Spellblasting Potion is 8 behind at 180 sec) and add the Greater Mageblood Elixir. Forever
// turns the Elixir of Greater Firepower into the Elixir of Holy Power (40 Holy damage) for Smite
// and Holy Fire: +15.8 at 180 sec. The Elixir of Greater Spirit adds 0.2 (1.1 at 300 sec), the
// Elixir of Sages would add 6.4 but is a second crit elixir. Kreeg's Stout Beatdown is a Dire
// Maul item, which the presets leave out like the full Arcanums; it would add -0.2 / +0.2 /
// +0.9 at 120 / 180 / 300 sec. The Flask of Distilled Wisdom is 71 behind the Flask of Supreme
// Power even here.
export const ConsumesSmite = Consumes.create({
	...DefaultConsumes,
	defaultPotion: Potions.MajorManaPotion,
	firePowerBuff: FirePowerBuff.ElixirOfGreaterFirepower,
	manaRegenElixir: ManaRegenElixir.GreaterMagebloodElixir,
	spiritElixir: SpiritElixir.ElixirOfGreaterSpirit,
});

export const EncounterLevel60Smite = PresetUtils.makePresetEncounter('Level 60 Smite', EncounterLevel60.encounter, {
	tanks: [],
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	buffs: DefaultIndividualBuffs,
	consumes: ConsumesSmite,
});

// Smite as every race on its set and consumes (20000 iterations, 120 / 180 / 300 sec,
// 2026-10-16, client 70425, with the raid hits, without Holy Fire):
//   Undead    744.0 / 724.6 / 639.7  (best overall and Horde)
//   Gnome     742.0 / 706.8 / 609.8  (best Alliance)
//   Human     732.9 / 697.1 / 602.5
//   Dwarf     728.9 / 691.2 / 595.6  (the generic boss is no Humanoid, so no Chastise)
//   Night Elf 728.7 / 690.2 / 595.2
//   Troll     728.5 / 689.4 / 593.5
// Undead's Dark Sacrifice and Gnome's larger mana pool and cheaper Eureka! casts are worth the
// most once the mana runs out.
export const PresetBuildLevel60Smite = PresetUtils.makePresetBuild('Level 60 Smite', {
	group: 'Level 60',
	tooltip: 'Level 60 Undead Holy damage priest (Smite, Power Infusion) on the synthetic Phase 1 gear, 22% behind Shadow.',
	gear: GearLevel60Phase1Smite,
	talents: TalentsLevel60Smite,
	rotation: APLLevel60Smite,
	encounter: EncounterLevel60Smite,
	race: Race.RaceUndead,
	level: 60,
	bonusTalentPoints: 0,
	professions: [OtherDefaults.profession1, OtherDefaults.profession2],
});
