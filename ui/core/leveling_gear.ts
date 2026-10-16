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
// an item for a slot. It's of about our level, and its quality is mostly white early on, then
// mostly green, with a few blues that grow more common with level. We wear it only when it
// beats what we have in that slot by our stat weights at that level. So most finds go unused,
// and a slot can keep a lucky blue for ten levels.
//
// Most finds are quest rewards, and there we pick the best of a few for our spec, the way an
// enhancement shaman takes the mail with Strength over the cloth with Intellect. The rest are
// drops, which have whatever stats they have. On top of those we run the dungeons of our
// faction once or twice in their level range, and their bosses drop the blues that make a
// big step, like Ragefire Chasm's at 13 or the axe from Ruins of Lordaeron.
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

// Of the finds, this share are quest rewards, where we take the best of QUEST_CHOICES for us.
// The rest are drops.
export const QUEST_SHARE = 0.6;
export const QUEST_CHOICES = 3;

// The dungeons we run while leveling, if our faction goes there. Each one we run with this
// chance, twice: 1 and 3 levels after the lowest level its loot needs. Each boss drops one item
// from its own loot, or two half the time, and we win an item we'd wear with WIN_CHANCE, since
// others in the group roll on it too. Raids and battlegrounds aren't here.
export const DUNGEON_RUN_CHANCE = 0.8;
export const DUNGEON_RUN_LEVELS = [1, 3];
export const DUNGEON_WIN_CHANCE = 0.5;
export const LEVELING_DUNGEONS: Array<{ zoneId: number; faction?: Faction }> = [
	{ zoneId: 2437, faction: Faction.Horde }, // Ragefire Chasm, in Orgrimmar
	{ zoneId: 16611, faction: Faction.Horde }, // Ruins of Lordaeron, Forever's, under Undercity
	{ zoneId: 16919, faction: Faction.Alliance }, // The Hall of Thanes, Forever's, a guess that it's the dwarves'
	{ zoneId: 1581, faction: Faction.Alliance }, // The Deadmines
	{ zoneId: 718 }, // Wailing Caverns
	{ zoneId: 209 }, // Shadowfang Keep
	{ zoneId: 717, faction: Faction.Alliance }, // The Stockade
	{ zoneId: 719 }, // Blackfathom Deeps
	{ zoneId: 721 }, // Gnomeregan
	{ zoneId: 491 }, // Razorfen Kraul
	{ zoneId: 796 }, // Scarlet Monastery
	{ zoneId: 722 }, // Razorfen Downs
	{ zoneId: 1337 }, // Uldaman
	{ zoneId: 1176 }, // Zul'Farrak
	{ zoneId: 2100 }, // Maraudon
	{ zoneId: 1477 }, // The Temple of Atal'Hakkar
	{ zoneId: 1584 }, // Blackrock Depths
	{ zoneId: 1583 }, // Blackrock Spire
	{ zoneId: 2557 }, // Dire Maul
	{ zoneId: 2017 }, // Stratholme
	{ zoneId: 2057 }, // Scholomance
];

// The dungeons above and the raids and battlegrounds, whose loot drops only there.
const INSTANCE_ZONES = new Set([...LEVELING_DUNGEONS.map(d => d.zoneId), 15475, 1977, 2159, 2677, 2717, 2597, 3277, 3358]);

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

// Where an item came from in the plan: the white starting gear, a quest reward, a drop, a
// dungeon boss, or a pin.
export type LevelingGearSource = 'start' | 'quest' | 'drop' | 'dungeon' | 'pinned';

export interface LevelingGearPiece {
	item: Item;
	randomSuffix?: ItemRandomSuffix;
	source: LevelingGearSource;
	// The dungeon it dropped in.
	dungeon?: string;
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
	// A zone's name, for the dungeons.
	zoneName?: (zoneId: number) => string | undefined;
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
	const dungeonRuns = planDungeonRuns(inputs, pool);

	// We start in white chest, legs and boots, with a white weapon.
	for (const slot of [ItemSlot.ItemSlotChest, ItemSlot.ItemSlotLegs, ItemSlot.ItemSlotFeet, ItemSlot.ItemSlotMainHand]) {
		const item = pool.plainItem(SLOT_TYPE[slot], 1, 1, () => 0);
		if (item) gear[slot] = { item, source: 'start', since: 1 };
	}

	for (let level = 1; level <= maxLevel; level++) {
		const weights = inputs.weightsAt(level);
		const changed: Array<ItemSlot> = [];
		const gainOf = (piece: LevelingGearPiece) => bestSlotFor(piece, gear, weights, inputs.classRules);
		const offer = (piece: LevelingGearPiece) => {
			const slot = gainOf(piece).slot;
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
				const quest = rand() < QUEST_SHARE;
				const choices: Array<LevelingGearPiece> = [];
				for (let c = 0; c < (quest ? QUEST_CHOICES : 1); c++) {
					const piece = pool.find(type, level, rand);
					if (piece) choices.push(piece);
				}
				if (choices.length == 0) continue;
				finds++;
				// A quest lets us choose, so we take the reward that helps us most.
				const best = choices.reduce((a, b) => (gainOf(b).gain > gainOf(a).gain ? b : a));
				offer({ ...best, source: quest ? 'quest' : 'drop' });
			}
		}
		for (const run of dungeonRuns.filter(run => run.level == level)) {
			const rand = mulberry32(hash(inputs.seed, level, 1000 + run.zoneId));
			for (const loot of run.bosses) {
				const drops = 1 + (rand() < 0.5 ? 1 : 0);
				for (let i = 0; i < drops; i++) {
					const usable = loot.filter(item => pool.canWearAt(item, level));
					if (usable.length == 0) continue;
					const piece = pool.withSuffix(usable[Math.floor(rand() * usable.length)], rand, 'dungeon', level);
					piece.dungeon = run.name;
					if (gainOf(piece).slot === undefined || rand() >= DUNGEON_WIN_CHANCE) continue;
					finds++;
					offer(piece);
				}
			}
		}
		for (const pin of inputs.pins) {
			if (pin.level == level) offer({ item: pin.item, randomSuffix: pin.randomSuffix, source: 'pinned', since: level });
		}

		levels.push({ level, gear: { ...gear }, changed, finds });
	}
	return { levels, madeUpItems: pool.madeUpItems() };
}

interface DungeonRun {
	level: number;
	zoneId: number;
	name: string;
	// Each boss's loot that we could use.
	bosses: Array<Array<Item>>;
}

// The dungeon runs of this journey: which dungeons of our faction we run, and at what levels.
function planDungeonRuns(inputs: LevelingGearInputs, pool: ItemPool): Array<DungeonRun> {
	const runs: Array<DungeonRun> = [];
	for (const dungeon of LEVELING_DUNGEONS) {
		if (dungeon.faction !== undefined && dungeon.faction != inputs.faction) continue;
		const bosses = pool.bossLoot(dungeon.zoneId);
		const levels = bosses
			.flat()
			.map(item => item.requiredLevel)
			.filter(level => level > 0);
		if (levels.length == 0) continue;
		const rand = mulberry32(hash(inputs.seed, 0, 1000 + dungeon.zoneId));
		if (rand() >= DUNGEON_RUN_CHANCE) continue;
		const start = Math.min(...levels);
		const name = inputs.zoneName?.(dungeon.zoneId) ?? `zone ${dungeon.zoneId}`;
		for (const offset of DUNGEON_RUN_LEVELS) {
			runs.push({ level: Math.min(start + offset, LEVELING_GEAR_MAX_LEVEL), zoneId: dungeon.zoneId, name, bosses });
		}
	}
	return runs;
}

// The slot a new piece goes in and what it gains us there, or no slot when it beats nothing we
// wear. A ring or a trinket replaces the weaker of the two, and we don't wear two of a unique
// item.
function bestSlotFor(
	piece: LevelingGearPiece,
	gear: Partial<Record<ItemSlot, LevelingGearPiece>>,
	weights: Stats,
	rules: LevelingGearClassRules,
): { slot: ItemSlot | undefined; gain: number } {
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
	return { slot: best, gain: bestGain };
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
	// Each dungeon's boss loot we could use, by zone and boss.
	private readonly bosses = new Map<number, Map<number, Array<Item>>>();
	private readonly plain: PlainItems;
	private readonly usedPlain = new Map<number, Item>();

	constructor(inputs: LevelingGearInputs) {
		this.inputs = inputs;
		this.plain = new PlainItems(inputs.items);
		for (const item of inputs.items) {
			if (!this.canFind(item)) continue;
			// A dungeon boss's or a raid's loot only drops there, so it's not a find out in the
			// world.
			const instanceOnly = item.sources.length > 0 && item.sources.every(s => s.source.oneofKind == 'drop' && INSTANCE_ZONES.has(s.source.drop.zoneId));
			if (!instanceOnly) push(this.byTypeQuality, `${item.type}/${item.quality}`, item);
			for (const source of item.sources) {
				if (source.source.oneofKind != 'drop' || !source.source.drop.npcId) continue;
				const { zoneId, npcId } = source.source.drop;
				if (!this.bosses.has(zoneId)) this.bosses.set(zoneId, new Map());
				const loot = this.bosses.get(zoneId)!;
				if (!loot.get(npcId)?.includes(item)) push(loot, npcId, item);
			}
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

	bossLoot(zoneId: number): Array<Array<Item>> {
		return [...(this.bosses.get(zoneId)?.values() ?? [])];
	}

	// Whether we can wear the item at this level: its required level and our armor type then.
	canWearAt(item: Item, level: number): boolean {
		return item.requiredLevel <= level && (item.armorType == ArmorType.ArmorTypeUnknown || item.armorType <= this.inputs.classRules.maxArmorTypeAt(level));
	}

	// The item as a piece we have, with one of its random suffixes rolled.
	withSuffix(item: Item, rand: () => number, source: LevelingGearSource, level: number): LevelingGearPiece {
		let randomSuffix: ItemRandomSuffix | undefined;
		if (item.randomSuffixOptions.length > 0) {
			randomSuffix = this.inputs.randomSuffix(item.randomSuffixOptions[Math.floor(rand() * item.randomSuffixOptions.length)]);
		}
		return { item, randomSuffix, source, since: level };
	}

	// An item we come across at this level, or undefined when there's none of the rolled kind.
	find(type: ItemType, level: number, rand: () => number): LevelingGearPiece | undefined {
		const quality = rollQuality(level, rand());
		if (quality == ItemQuality.ItemQualityCommon) {
			const ilvl = Math.max(1, level + Math.floor(rand() * (ILVL_ABOVE + 1)));
			const item = this.plainItem(type, ilvl, level, rand);
			return item && { item, source: 'drop', since: level };
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
		return this.withSuffix(candidates[Math.floor(rand() * candidates.length)], rand, 'drop', level);
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
// piece's armor and 85% of a green weapon's DPS. The green lines are fit to the database. Each
// one wears the icon of the lowest level green of its kind, like a leather vest's or a
// two-handed mace's.
//
// They only depend on the database, so the database makes the same ones at startup
// (Database.getItemById), and gear with them in it still loads after a reload.
export class PlainItems {
	private readonly armorFit = new Map<string, [number, number]>();
	private readonly weaponFit = new Map<string, [number, number]>();
	private readonly icons = new Map<string, Item>();

	constructor(items: Array<Item>) {
		const weaponSamples = new Map<string, Array<[number, number]>>();
		const armorSamples = new Map<string, Array<[number, number]>>();
		for (const item of items) {
			if (item.quality != ItemQuality.ItemQualityUncommon) continue;
			const twoHand = item.handType == HandType.HandTypeTwoHand;
			if (item.weaponSpeed > 0) push(weaponSamples, String(twoHand), [item.ilvl, getWeaponDPS(item)]);
			else if (item.stats[Stat.StatArmor] > 0) push(armorSamples, `${item.type}/${item.armorType}`, [item.ilvl, item.stats[Stat.StatArmor]]);
			const iconKey = item.type == ItemType.ItemTypeWeapon ? `w/${item.weaponType}/${twoHand}` : `a/${item.type}/${item.armorType}`;
			const best = this.icons.get(iconKey);
			if (item.icon && (!best || item.ilvl < best.ilvl || (item.ilvl == best.ilvl && item.id < best.id))) this.icons.set(iconKey, item);
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
		item.icon = this.icons.get(`a/${type}/${armorType}`)?.icon ?? this.icons.get(`a/${type}/${ArmorType.ArmorTypeCloth}`)?.icon ?? item.icon;
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
		item.icon = this.icons.get(`w/${weaponType}/${twoHand}`)?.icon ?? item.icon;
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
