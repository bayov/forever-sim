import {
	ArmorType,
	Class,
	Faction,
	HandType,
	ItemQuality,
	ItemRandomSuffix,
	ItemSlot,
	ItemType,
	PseudoStat,
	RangedWeaponType,
	Stat,
	WeaponType,
} from './proto/common.js';
import { UIItem as Item, UIItem_FactionRestriction as FactionRestriction } from './proto/ui.js';
import { Stats } from './proto_utils/stats.js';

// The leveling gear plan: what a character wears at each level from 1 to 60, when it gears
// up the way we do while speed leveling (docs/leveling_gear.md).
//
// We don't look for the best item at each level. We level up and now and then come across
// an item for a slot, from a quest reward or a drop. It's of about our level, and its
// quality is mostly white early on, then mostly green, with a few blues that grow more
// common with level. It has whatever stats it has, often ones we don't want. We wear it only
// when it beats what we have in that slot by our stat weights at that level. So most finds
// go unused, and a slot can keep a lucky blue for ten levels.
//
// The items are real ones from the item database: one we could find at that level, of that
// quality and for that slot, picked at random. A random suffix item rolls one of its
// suffixes. The database has no white items, so a white find is a plain piece we make up
// here, with the armor or weapon damage of its item level and no stats.
//
// A plan is one random history, from its seed. The same seed and pins give the same plan,
// and a pin doesn't change any other find, because each level and slot rolls its own random
// numbers.

// One level's finds, per kind of slot: how many items we come across while leveling through
// it, on average. Two ring and two trinket slots share their finds.
//
// These are guesses (notes.md, Need to Verify). At 0.3 a slot sees a find every three or
// four levels, and most finds don't beat what we wear.
export const FINDS_PER_LEVEL: Partial<Record<ItemType, number>> = {
	[ItemType.ItemTypeHead]: 0.25,
	[ItemType.ItemTypeNeck]: 0.15,
	[ItemType.ItemTypeShoulder]: 0.25,
	[ItemType.ItemTypeBack]: 0.25,
	[ItemType.ItemTypeChest]: 0.3,
	[ItemType.ItemTypeWrist]: 0.3,
	[ItemType.ItemTypeHands]: 0.3,
	[ItemType.ItemTypeWaist]: 0.3,
	[ItemType.ItemTypeLegs]: 0.3,
	[ItemType.ItemTypeFeet]: 0.3,
	[ItemType.ItemTypeFinger]: 0.3,
	[ItemType.ItemTypeTrinket]: 0.06,
	[ItemType.ItemTypeWeapon]: 0.4,
	[ItemType.ItemTypeRanged]: 0.1,
};

// The quality of a find by our level, as the share of white, green and blue. In between two
// rows it goes in a straight line. Greens take over around level 12. Blues come from the
// first dungeons on and slowly grow more common.
export const QUALITY_MIX: Array<{ level: number; white: number; green: number; blue: number }> = [
	{ level: 1, white: 0.85, green: 0.15, blue: 0 },
	{ level: 8, white: 0.6, green: 0.4, blue: 0 },
	{ level: 12, white: 0.4, green: 0.58, blue: 0.02 },
	{ level: 18, white: 0.15, green: 0.8, blue: 0.05 },
	{ level: 25, white: 0.08, green: 0.82, blue: 0.1 },
	{ level: 35, white: 0.05, green: 0.81, blue: 0.14 },
	{ level: 45, white: 0.04, green: 0.78, blue: 0.18 },
	{ level: 60, white: 0.03, green: 0.72, blue: 0.25 },
];

// A find is of about our level: its item level is from 3 under ours to 5 over, and we must be
// able to wear it already.
const ILVL_BELOW = 3;
const ILVL_ABOVE = 5;

export const LEVELING_GEAR_MAX_LEVEL = 60;

// The slots the plan fills, in the order the UI lists them.
export const LEVELING_GEAR_SLOTS: Array<ItemSlot> = [
	ItemSlot.ItemSlotHead,
	ItemSlot.ItemSlotNeck,
	ItemSlot.ItemSlotShoulder,
	ItemSlot.ItemSlotBack,
	ItemSlot.ItemSlotChest,
	ItemSlot.ItemSlotWrist,
	ItemSlot.ItemSlotHands,
	ItemSlot.ItemSlotWaist,
	ItemSlot.ItemSlotLegs,
	ItemSlot.ItemSlotFeet,
	ItemSlot.ItemSlotFinger1,
	ItemSlot.ItemSlotFinger2,
	ItemSlot.ItemSlotTrinket1,
	ItemSlot.ItemSlotTrinket2,
	ItemSlot.ItemSlotMainHand,
	ItemSlot.ItemSlotOffHand,
	ItemSlot.ItemSlotRanged,
];

const SLOT_TYPE: Record<number, ItemType> = {
	[ItemSlot.ItemSlotHead]: ItemType.ItemTypeHead,
	[ItemSlot.ItemSlotNeck]: ItemType.ItemTypeNeck,
	[ItemSlot.ItemSlotShoulder]: ItemType.ItemTypeShoulder,
	[ItemSlot.ItemSlotBack]: ItemType.ItemTypeBack,
	[ItemSlot.ItemSlotChest]: ItemType.ItemTypeChest,
	[ItemSlot.ItemSlotWrist]: ItemType.ItemTypeWrist,
	[ItemSlot.ItemSlotHands]: ItemType.ItemTypeHands,
	[ItemSlot.ItemSlotWaist]: ItemType.ItemTypeWaist,
	[ItemSlot.ItemSlotLegs]: ItemType.ItemTypeLegs,
	[ItemSlot.ItemSlotFeet]: ItemType.ItemTypeFeet,
	[ItemSlot.ItemSlotFinger1]: ItemType.ItemTypeFinger,
	[ItemSlot.ItemSlotFinger2]: ItemType.ItemTypeFinger,
	[ItemSlot.ItemSlotTrinket1]: ItemType.ItemTypeTrinket,
	[ItemSlot.ItemSlotTrinket2]: ItemType.ItemTypeTrinket,
	[ItemSlot.ItemSlotMainHand]: ItemType.ItemTypeWeapon,
	[ItemSlot.ItemSlotOffHand]: ItemType.ItemTypeWeapon,
	[ItemSlot.ItemSlotRanged]: ItemType.ItemTypeRanged,
};

// Where an item came from in the plan: the white starting gear, a find, or a pin.
export type LevelingGearSource = 'start' | 'found' | 'pinned';

export interface LevelingGearPiece {
	item: Item;
	randomSuffix?: ItemRandomSuffix;
	source: LevelingGearSource;
	// The level we put it on at.
	since: number;
}

export interface LevelingGearLevel {
	level: number;
	gear: Partial<Record<ItemSlot, LevelingGearPiece>>;
	// The slots that got a new item at this level.
	changed: Array<ItemSlot>;
	// How many items we came across at this level, worn or not.
	finds: number;
}

// An item we put in the plan by hand: we have it from this level on, if it beats what the
// plan has in that slot.
export interface LevelingGearPinInput {
	level: number;
	slot: ItemSlot;
	item: Item;
	randomSuffix?: ItemRandomSuffix;
}

// What our class can wear and wield (from proto_utils/utils.ts). The caller passes them in, so
// this file runs without the rest of the UI.
export interface LevelingGearClassRules {
	maxArmorTypeAt: (level: number) => ArmorType;
	weaponTypes: Array<{ weaponType: WeaponType; canUseTwoHand?: boolean }>;
	rangedWeaponTypes: Array<RangedWeaponType>;
	canDualWield: boolean;
}

export interface LevelingGearInputs {
	// Every item in the database, and the random suffixes by id.
	items: Array<Item>;
	randomSuffix: (id: number) => ItemRandomSuffix | undefined;
	playerClass: Class;
	classRules: LevelingGearClassRules;
	faction: Faction;
	seed: number;
	pins: Array<LevelingGearPinInput>;
	// The stat weights we judge items by at a level.
	weightsAt: (level: number) => Stats;
	maxLevel?: number;
}

export interface LevelingGearPlan {
	levels: Array<LevelingGearLevel>;
	// The plain white items the plan made up, so the UI and the sim can look them up.
	madeUpItems: Array<Item>;
}

// planLevelingGear plays out one leveling journey and says what we wear at each level.
export function planLevelingGear(inputs: LevelingGearInputs): LevelingGearPlan {
	const maxLevel = inputs.maxLevel ?? LEVELING_GEAR_MAX_LEVEL;
	const pool = new ItemPool(inputs);
	const gear: Partial<Record<ItemSlot, LevelingGearPiece>> = {};
	const levels: Array<LevelingGearLevel> = [];

	// We start in white chest, legs and boots, with a white weapon.
	for (const slot of [ItemSlot.ItemSlotChest, ItemSlot.ItemSlotLegs, ItemSlot.ItemSlotFeet, ItemSlot.ItemSlotMainHand]) {
		const item = pool.plainItem(SLOT_TYPE[slot], 1, 1, () => 0);
		if (item) gear[slot] = { item, source: 'start', since: 1 };
	}

	for (let level = 1; level <= maxLevel; level++) {
		const weights = inputs.weightsAt(level);
		const changed: Array<ItemSlot> = [];
		const offer = (piece: LevelingGearPiece) => {
			const slot = bestSlotFor(piece, gear, weights, inputs.classRules);
			if (slot === undefined) return;
			gear[slot] = piece;
			if (slot == ItemSlot.ItemSlotMainHand && piece.item.handType == HandType.HandTypeTwoHand) delete gear[ItemSlot.ItemSlotOffHand];
			if (!changed.includes(slot)) changed.push(slot);
		};

		let finds = 0;
		for (const [typeKey, rate] of Object.entries(FINDS_PER_LEVEL)) {
			const type = Number(typeKey) as ItemType;
			const rand = mulberry32(hash(inputs.seed, level, type));
			const count = poisson(rate!, rand);
			for (let i = 0; i < count; i++) {
				const piece = pool.find(type, level, rand);
				if (!piece) continue;
				finds++;
				offer(piece);
			}
		}
		for (const pin of inputs.pins) {
			if (pin.level == level) offer({ item: pin.item, randomSuffix: pin.randomSuffix, source: 'pinned', since: level });
		}

		levels.push({ level, gear: { ...gear }, changed, finds });
	}
	return { levels, madeUpItems: pool.madeUpItems() };
}

// The slot a new piece goes in, or undefined when it beats nothing we wear. A ring or a
// trinket replaces the weaker of the two, and we don't wear two of a unique item.
function bestSlotFor(
	piece: LevelingGearPiece,
	gear: Partial<Record<ItemSlot, LevelingGearPiece>>,
	weights: Stats,
	rules: LevelingGearClassRules,
): ItemSlot | undefined {
	const item = piece.item;
	const slots = slotsFor(item, rules);
	let best: ItemSlot | undefined;
	let bestGain = 0;
	for (const slot of slots) {
		const other =
			slot == ItemSlot.ItemSlotFinger1
				? ItemSlot.ItemSlotFinger2
				: slot == ItemSlot.ItemSlotFinger2
				? ItemSlot.ItemSlotFinger1
				: slot == ItemSlot.ItemSlotTrinket1
				? ItemSlot.ItemSlotTrinket2
				: slot == ItemSlot.ItemSlotTrinket2
				? ItemSlot.ItemSlotTrinket1
				: undefined;
		if (other !== undefined && gear[other]?.item.id == item.id) continue;
		const gain = scorePiece(piece, slot, weights) - currentScore(slot, gear, weights, piece);
		if (gain > bestGain + 1e-9) {
			best = slot;
			bestGain = gain;
		}
	}
	return best;
}

// What we give up in a slot to wear the piece there. A two-hander also takes the off hand,
// and an off hand item can't go next to a two-hander.
function currentScore(slot: ItemSlot, gear: Partial<Record<ItemSlot, LevelingGearPiece>>, weights: Stats, piece: LevelingGearPiece): number {
	const worn = gear[slot];
	let score = worn ? scorePiece(worn, slot, weights) : 0;
	if (slot == ItemSlot.ItemSlotMainHand && piece.item.handType == HandType.HandTypeTwoHand) {
		const offHand = gear[ItemSlot.ItemSlotOffHand];
		if (offHand) score += scorePiece(offHand, ItemSlot.ItemSlotOffHand, weights);
	}
	if (slot == ItemSlot.ItemSlotOffHand && gear[ItemSlot.ItemSlotMainHand]?.item.handType == HandType.HandTypeTwoHand) {
		return Infinity;
	}
	return score;
}

function slotsFor(item: Item, rules: LevelingGearClassRules): Array<ItemSlot> {
	switch (item.type) {
		case ItemType.ItemTypeFinger:
			return [ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2];
		case ItemType.ItemTypeTrinket:
			return [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2];
		case ItemType.ItemTypeWeapon: {
			if (item.handType == HandType.HandTypeOffHand || [WeaponType.WeaponTypeShield, WeaponType.WeaponTypeOffHand].includes(item.weaponType)) {
				return [ItemSlot.ItemSlotOffHand];
			}
			if (item.handType == HandType.HandTypeTwoHand || item.handType == HandType.HandTypeMainHand) return [ItemSlot.ItemSlotMainHand];
			return rules.canDualWield ? [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand] : [ItemSlot.ItemSlotMainHand];
		}
		case ItemType.ItemTypeRanged:
			return [ItemSlot.ItemSlotRanged];
	}
	const slot = LEVELING_GEAR_SLOTS.find(s => SLOT_TYPE[s] == item.type);
	return slot === undefined ? [] : [slot];
}

// scorePiece is a piece's worth by the stat weights, like the gear picker's EP. Under Forever
// an item's hit and crit count for melee and spells both, so we price them once, at the
// larger weight. Bonus armor takes the armor price when it has none of its own.
export function scorePiece(piece: LevelingGearPiece, slot: ItemSlot, weights: Stats): number {
	const item = piece.item;
	const stats = new Stats(item.stats).add(new Stats(piece.randomSuffix?.stats ?? []));
	let score = 0;
	const skip = new Set([Stat.StatMeleeHit, Stat.StatSpellHit, Stat.StatMeleeCrit, Stat.StatSpellCrit, Stat.StatArmor, Stat.StatBonusArmor]);
	for (let stat = 0; stat < item.stats.length; stat++) {
		if (!skip.has(stat)) score += stats.getStat(stat) * weights.getStat(stat);
	}
	const hit = stats.getStat(Stat.StatMeleeHit) + stats.getStat(Stat.StatSpellHit) + item.hitRating / 10;
	const crit = stats.getStat(Stat.StatMeleeCrit) + stats.getStat(Stat.StatSpellCrit) + item.critRating / 14;
	score += hit * Math.max(weights.getStat(Stat.StatMeleeHit), weights.getStat(Stat.StatSpellHit));
	score += crit * Math.max(weights.getStat(Stat.StatMeleeCrit), weights.getStat(Stat.StatSpellCrit));
	const armorWeight = weights.getStat(Stat.StatArmor);
	const bonusArmorWeight = weights.getStat(Stat.StatBonusArmor) || armorWeight;
	score += stats.getStat(Stat.StatArmor) * armorWeight + stats.getStat(Stat.StatBonusArmor) * bonusArmorWeight;
	if (item.weaponSpeed > 0) {
		const dps = getWeaponDPS(item);
		if (slot == ItemSlot.ItemSlotMainHand) score += dps * weights.getPseudoStat(PseudoStat.PseudoStatMainHandDps);
		else if (slot == ItemSlot.ItemSlotOffHand) score += dps * weights.getPseudoStat(PseudoStat.PseudoStatOffHandDps);
		else if (slot == ItemSlot.ItemSlotRanged) score += dps * weights.getPseudoStat(PseudoStat.PseudoStatRangedDps);
	}
	// An item with nothing we price still beats an empty slot, a little.
	return score + 0.01;
}

// The items we can come across, by kind of slot and quality.
class ItemPool {
	private readonly inputs: LevelingGearInputs;
	private readonly byTypeQuality = new Map<string, Array<Item>>();
	private readonly plain: PlainItems;
	private readonly usedPlain = new Map<number, Item>();

	constructor(inputs: LevelingGearInputs) {
		this.inputs = inputs;
		this.plain = new PlainItems(inputs.items);
		for (const item of inputs.items) {
			if (this.canFind(item)) push(this.byTypeQuality, `${item.type}/${item.quality}`, item);
		}
	}

	// Whether we could come across the item while leveling: a green or a blue our faction and
	// class can use, from a drop or a quest, or a world drop with no source. Crafted, vendor
	// and reputation items we'd have to go out of our way for.
	private canFind(item: Item): boolean {
		if (item.quality != ItemQuality.ItemQualityUncommon && item.quality != ItemQuality.ItemQualityRare) return false;
		if (item.classAllowlist.length > 0 && !item.classAllowlist.includes(this.inputs.playerClass)) return false;
		const restriction = item.factionRestriction;
		if (restriction == FactionRestriction.ALLIANCE_ONLY && this.inputs.faction == Faction.Horde) return false;
		if (restriction == FactionRestriction.HORDE_ONLY && this.inputs.faction == Faction.Alliance) return false;
		if (item.sources.length > 0 && !item.sources.some(s => s.source.oneofKind == 'drop' || s.source.oneofKind == 'quest')) return false;
		if (item.type == ItemType.ItemTypeWeapon) return this.canWield(item.weaponType, item.handType == HandType.HandTypeTwoHand);
		if (item.type == ItemType.ItemTypeRanged) return this.inputs.classRules.rangedWeaponTypes.includes(item.rangedWeaponType);
		return true;
	}

	private canWield(weaponType: WeaponType, twoHand: boolean): boolean {
		return this.inputs.classRules.weaponTypes.some(wt => wt.weaponType == weaponType && (!twoHand || wt.canUseTwoHand));
	}

	// An item we come across at this level, or undefined when there's none of the rolled kind.
	find(type: ItemType, level: number, rand: () => number): LevelingGearPiece | undefined {
		const quality = rollQuality(level, rand());
		if (quality == ItemQuality.ItemQualityCommon) {
			const ilvl = Math.max(1, level + Math.floor(rand() * (ILVL_ABOVE + 1)));
			const item = this.plainItem(type, ilvl, level, rand);
			return item && { item, source: 'found', since: level };
		}
		const maxArmor = this.inputs.classRules.maxArmorTypeAt(level);
		const candidates = (this.byTypeQuality.get(`${type}/${quality}`) ?? []).filter(
			item =>
				item.requiredLevel <= level &&
				item.ilvl >= level - ILVL_BELOW &&
				item.ilvl <= level + ILVL_ABOVE &&
				(item.armorType == ArmorType.ArmorTypeUnknown || item.armorType <= maxArmor),
		);
		if (candidates.length == 0) return undefined;
		const item = candidates[Math.floor(rand() * candidates.length)];
		let randomSuffix: ItemRandomSuffix | undefined;
		if (item.randomSuffixOptions.length > 0) {
			randomSuffix = this.inputs.randomSuffix(item.randomSuffixOptions[Math.floor(rand() * item.randomSuffixOptions.length)]);
		}
		return { item, randomSuffix, source: 'found', since: level };
	}

	// A plain white item of this kind and item level: armor of the best type we can wear at this
	// level, or a weapon of a kind we can use. Necks, rings and trinkets have no white ones
	// worth wearing.
	plainItem(type: ItemType, ilvl: number, level: number, rand: () => number = () => 0): Item | undefined {
		let item: Item | undefined;
		if (type == ItemType.ItemTypeWeapon) {
			const kinds = PLAIN_WEAPONS.filter(([weaponType, twoHand]) => this.canWield(weaponType, twoHand));
			if (kinds.length == 0) return undefined;
			const [weaponType, twoHand] = kinds[Math.floor(rand() * kinds.length)];
			item = this.plain.weapon(weaponType, twoHand, ilvl);
		} else {
			const armorType = type == ItemType.ItemTypeBack ? ArmorType.ArmorTypeCloth : this.inputs.classRules.maxArmorTypeAt(level);
			item = this.plain.armor(type, armorType, ilvl);
		}
		if (item) this.usedPlain.set(item.id, item);
		return item;
	}

	madeUpItems(): Array<Item> {
		return [...this.usedPlain.values()];
	}
}

// The plain white weapons: a one-hander or a two-hander of each kind that has one.
const PLAIN_WEAPONS: Array<[WeaponType, boolean]> = [
	[WeaponType.WeaponTypeAxe, false],
	[WeaponType.WeaponTypeAxe, true],
	[WeaponType.WeaponTypeMace, false],
	[WeaponType.WeaponTypeMace, true],
	[WeaponType.WeaponTypeSword, false],
	[WeaponType.WeaponTypeSword, true],
	[WeaponType.WeaponTypeDagger, false],
	[WeaponType.WeaponTypeFist, false],
	[WeaponType.WeaponTypeStaff, true],
	[WeaponType.WeaponTypePolearm, true],
];
const PLAIN_ARMOR_TYPES = [
	ItemType.ItemTypeHead,
	ItemType.ItemTypeShoulder,
	ItemType.ItemTypeBack,
	ItemType.ItemTypeChest,
	ItemType.ItemTypeWrist,
	ItemType.ItemTypeHands,
	ItemType.ItemTypeWaist,
	ItemType.ItemTypeLegs,
	ItemType.ItemTypeFeet,
];
const PLAIN_MAX_ILVL = 65;

// PlainItems are the white items the plan makes up, since the item database has none: armor
// with no stats, and weapons. Their armor and weapon DPS go with item level, at 90% of a green
// piece's armor and 85% of a green weapon's DPS. The green lines are fit to the database.
//
// They only depend on the database, so the database makes the same ones at startup
// (Database.getItemById), and gear with them in it still loads after a reload.
export class PlainItems {
	private readonly armorFit = new Map<string, [number, number]>();
	private readonly weaponFit = new Map<string, [number, number]>();

	constructor(items: Array<Item>) {
		const weaponSamples = new Map<string, Array<[number, number]>>();
		const armorSamples = new Map<string, Array<[number, number]>>();
		for (const item of items) {
			if (item.quality != ItemQuality.ItemQualityUncommon) continue;
			if (item.weaponSpeed > 0) push(weaponSamples, String(item.handType == HandType.HandTypeTwoHand), [item.ilvl, getWeaponDPS(item)]);
			else if (item.stats[Stat.StatArmor] > 0) push(armorSamples, `${item.type}/${item.armorType}`, [item.ilvl, item.stats[Stat.StatArmor]]);
		}
		armorSamples.forEach((samples, key) => this.armorFit.set(key, lineFit(samples)));
		weaponSamples.forEach((samples, key) => this.weaponFit.set(key, lineFit(samples)));
	}

	armor(type: ItemType, armorType: ArmorType, ilvl: number): Item | undefined {
		if (!PLAIN_ARMOR_TYPES.includes(type)) return undefined;
		const [a, b] = this.armorFit.get(`${type}/${armorType}`) ?? [0, 0];
		const name = `Plain ${ARMOR_NAMES[armorType] ?? ''} ${SLOT_NAMES[type] ?? 'Item'}`.replace(/\s+/g, ' ');
		const item = madeUpItem(plainId(type, armorType, ilvl), name, type, ilvl);
		item.armorType = armorType;
		item.stats[Stat.StatArmor] = Math.max(1, Math.round(0.9 * (a + b * ilvl)));
		return item;
	}

	weapon(weaponType: WeaponType, twoHand: boolean, ilvl: number): Item {
		const speed = twoHand ? 3.3 : 2.4;
		const [a, b] = this.weaponFit.get(String(twoHand)) ?? [0, 0];
		const dps = Math.max(1, 0.85 * (a + b * ilvl));
		const item = madeUpItem(
			plainId(ItemType.ItemTypeWeapon, weaponType * 10 + (twoHand ? 1 : 0), ilvl),
			`Plain ${weaponName(weaponType, twoHand)}`,
			ItemType.ItemTypeWeapon,
			ilvl,
		);
		item.weaponType = weaponType;
		item.handType = twoHand ? HandType.HandTypeTwoHand : HandType.HandTypeMainHand;
		item.weaponSpeed = speed;
		item.weaponDamageMin = Math.round(dps * speed * 0.8);
		item.weaponDamageMax = Math.round(dps * speed * 1.2);
		return item;
	}

	// Every plain item, for the database.
	all(): Array<Item> {
		const out: Array<Item> = [];
		for (let ilvl = 1; ilvl <= PLAIN_MAX_ILVL; ilvl++) {
			for (const type of PLAIN_ARMOR_TYPES) {
				for (const armorType of [ArmorType.ArmorTypeCloth, ArmorType.ArmorTypeLeather, ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate]) {
					const item = this.armor(type, armorType, ilvl);
					if (item) out.push(item);
				}
			}
			for (const [weaponType, twoHand] of PLAIN_WEAPONS) out.push(this.weapon(weaponType, twoHand, ilvl));
		}
		return out;
	}
}

// The made up items have ids from 970,000,000 up, one per kind and item level.
const PLAIN_ID_BASE = 970_000_000;
function plainId(type: ItemType, kind: number, ilvl: number): number {
	return PLAIN_ID_BASE + type * 100_000 + kind * 100 + ilvl;
}

export function isPlainItemId(id: number): boolean {
	return id >= PLAIN_ID_BASE && id < PLAIN_ID_BASE + 100_000_000;
}

const ARMOR_NAMES: Partial<Record<ArmorType, string>> = {
	[ArmorType.ArmorTypeCloth]: 'Cloth',
	[ArmorType.ArmorTypeLeather]: 'Leather',
	[ArmorType.ArmorTypeMail]: 'Mail',
	[ArmorType.ArmorTypePlate]: 'Plate',
};

const SLOT_NAMES: Partial<Record<ItemType, string>> = {
	[ItemType.ItemTypeHead]: 'Cap',
	[ItemType.ItemTypeShoulder]: 'Shoulders',
	[ItemType.ItemTypeBack]: 'Cloak',
	[ItemType.ItemTypeChest]: 'Vest',
	[ItemType.ItemTypeWrist]: 'Bracers',
	[ItemType.ItemTypeHands]: 'Gloves',
	[ItemType.ItemTypeWaist]: 'Belt',
	[ItemType.ItemTypeLegs]: 'Pants',
	[ItemType.ItemTypeFeet]: 'Boots',
};

function weaponName(weaponType: WeaponType, twoHand: boolean): string {
	const name = WeaponType[weaponType].replace('WeaponType', '');
	return twoHand ? `Two-Handed ${name}` : name;
}

function madeUpItem(id: number, name: string, type: ItemType, ilvl: number): Item {
	return Item.create({
		id,
		name,
		icon: 'inv_misc_questionmark',
		type,
		stats: new Array(Stat.StatFeralAttackPower + 1).fill(0),
		ilvl,
		requiredLevel: Math.max(1, ilvl - 5),
		quality: ItemQuality.ItemQualityCommon,
		rangedWeaponType: RangedWeaponType.RangedWeaponTypeUnknown,
	});
}

function rollQuality(level: number, roll: number): ItemQuality {
	const mix = qualityMixAt(level);
	if (roll < mix.white) return ItemQuality.ItemQualityCommon;
	if (roll < mix.white + mix.green) return ItemQuality.ItemQualityUncommon;
	return ItemQuality.ItemQualityRare;
}

export function qualityMixAt(level: number): { white: number; green: number; blue: number } {
	const rows = QUALITY_MIX;
	if (level <= rows[0].level) return rows[0];
	for (let i = 1; i < rows.length; i++) {
		if (level <= rows[i].level) {
			const a = rows[i - 1];
			const b = rows[i];
			const t = (level - a.level) / (b.level - a.level);
			return { white: a.white + t * (b.white - a.white), green: a.green + t * (b.green - a.green), blue: a.blue + t * (b.blue - a.blue) };
		}
	}
	return rows[rows.length - 1];
}

// A weapon's DPS, like getWeaponDPS in proto_utils/equipped_item.ts, which we don't import
// because it brings the rest of the UI along.
function getWeaponDPS(item: Item): number {
	return (item.weaponDamageMin + item.weaponDamageMax) / 2 / (item.weaponSpeed || 1);
}

function push<K, V>(map: Map<K, Array<V>>, key: K, value: V) {
	const list = map.get(key);
	if (list) list.push(value);
	else map.set(key, [value]);
}

// The least squares line through the points, as [intercept, slope].
function lineFit(points: Array<[number, number]>): [number, number] {
	const n = points.length;
	if (n < 2) return [points[0]?.[1] ?? 0, 0];
	const mx = points.reduce((s, p) => s + p[0], 0) / n;
	const my = points.reduce((s, p) => s + p[1], 0) / n;
	const sxx = points.reduce((s, p) => s + (p[0] - mx) ** 2, 0);
	const sxy = points.reduce((s, p) => s + (p[0] - mx) * (p[1] - my), 0);
	const slope = sxx > 0 ? sxy / sxx : 0;
	return [my - slope * mx, slope];
}

function poisson(mean: number, rand: () => number): number {
	const limit = Math.exp(-mean);
	let k = 0;
	let p = rand();
	while (p > limit) {
		k++;
		p *= rand();
	}
	return k;
}

function hash(...values: Array<number>): number {
	let h = 2166136261;
	for (const v of values) {
		h ^= v | 0;
		h = Math.imul(h, 16777619);
		h ^= h >>> 13;
	}
	return h >>> 0;
}

function mulberry32(seed: number): () => number {
	let a = seed;
	return () => {
		a = (a + 0x6d2b79f5) | 0;
		let t = Math.imul(a ^ (a >>> 15), 1 | a);
		t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}
