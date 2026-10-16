import { ActionId } from '../../proto_utils/action_id.js';
import { SimLog } from '../../proto_utils/logs_parser.js';
import { TypedEvent } from '../../typed_event.js';

export type EventCategory = 'casts' | 'damage' | 'auras' | 'resources' | 'session' | 'other' | 'debug';

export const eventCategories: Array<{ category: EventCategory; label: string }> = [
	{ category: 'casts', label: 'Casts' },
	{ category: 'damage', label: 'Damage' },
	{ category: 'auras', label: 'Auras' },
	{ category: 'resources', label: 'Resources' },
	{ category: 'session', label: 'Session' },
	{ category: 'other', label: 'Other' },
	{ category: 'debug', label: 'Debug' },
];

// The categories that only have events in a leveling session. Outside one their chip has
// nothing to show, so the Events view hides it.
export const levelingOnlyCategories: Array<EventCategory> = ['session'];

// The sim's own debug lines (the ones with [DEBUG]) are noise most of the time, so they start off.
const defaultCategories: Array<EventCategory> = ['casts', 'damage', 'auras', 'resources', 'session', 'other'];

// The leveling session's lines that only say where we are, how we move and which state we're
// in. The replay's map and the session strip read them, but in the log they are noise.
const levelingDebugLine = /Leveling: (we're at \(|now \w+$|Target \d+ at \(|\{[^}]*\} at \()/;

// What the Events tab shows of an iteration. The timeline, the chart and the log all read the
// same filter, so switching between them keeps what we narrowed down.
//
// Not every setting means something in every view. The chart only follows the time range, and
// the timeline has no rows for "Other" events (stat changes and the like).
export class EventFilter {
	readonly changeEmitter = new TypedEvent<void>('EventFilter');

	private searchQuery = '';
	private keywords: Array<string> = [];

	// In seconds from the pull. Null means the start or the end of the iteration.
	from: number | null = null;
	to: number | null = null;

	readonly categories = new Set<EventCategory>(defaultCategories);
	hidden: Array<ActionId> = [];
	// Hides the timeline's lanes for buffs and debuffs that are up for the whole fight, like
	// Blessing of Kings or Mark of the Wild. They only fill a lane with one long bar.
	hideAlwaysUp = false;

	get search(): string {
		return this.searchQuery;
	}

	setSearch(query: string) {
		this.searchQuery = query;
		this.keywords = parseKeywords(query);
		this.emit();
	}

	setRange(from: number | null, to: number | null) {
		this.from = from;
		this.to = to;
		this.emit();
	}

	toggleCategory(category: EventCategory) {
		if (this.categories.has(category)) {
			this.categories.delete(category);
		} else {
			this.categories.add(category);
		}
		this.emit();
	}

	hide(actionId: ActionId) {
		if (this.isHidden(actionId)) return;
		this.hidden.push(actionId);
		this.emit();
	}

	unhide(actionId: ActionId) {
		this.hidden = this.hidden.filter(hiddenId => !hiddenId.equals(actionId));
		this.emit();
	}

	toggleHideAlwaysUp() {
		this.hideAlwaysUp = !this.hideAlwaysUp;
		this.emit();
	}

	unhideAll() {
		this.hidden = [];
		this.emit();
	}

	reset() {
		this.searchQuery = '';
		this.keywords = [];
		this.from = null;
		this.to = null;
		this.categories.clear();
		defaultCategories.forEach(category => this.categories.add(category));
		this.hidden = [];
		this.hideAlwaysUp = false;
		this.emit();
	}

	isActive(): boolean {
		const defaultCategoriesOn = this.categories.size == defaultCategories.length && defaultCategories.every(category => this.categories.has(category));
		return this.keywords.length > 0 || this.from != null || this.to != null || !defaultCategoriesOn || this.hidden.length > 0 || this.hideAlwaysUp;
	}

	isHidden(actionId: ActionId): boolean {
		return this.hidden.some(hiddenId => hiddenId.equals(actionId));
	}

	inRange(time: number): boolean {
		return (this.from == null || time >= this.from) && (this.to == null || time <= this.to);
	}

	// Whether there is anything to search for. Without it every text matches, and the log can
	// skip making the rows it doesn't show.
	hasSearch(): boolean {
		return this.keywords.length > 0;
	}

	// Whether the text has every word of the search. Quoted phrases have to match whole.
	matchesText(lowerCaseText: string): boolean {
		return this.keywords.every(keyword => lowerCaseText.includes(keyword));
	}

	matchesLog(log: SimLog, lowerCaseText: string): boolean {
		return this.matchesLogExceptText(log, EventFilter.categoryOf(log)) && this.matchesText(lowerCaseText);
	}

	// Everything matchesLog asks but the search, for a log that already knows its category.
	matchesLogExceptText(log: SimLog, category: EventCategory): boolean {
		return this.categories.has(category) && this.inRange(log.timestamp) && !(log.actionId && this.isHidden(log.actionId));
	}

	static categoryOf(log: SimLog): EventCategory {
		if (log.raw.includes('[DEBUG]')) return 'debug';
		if (log.raw.includes('] Leveling: ')) return levelingDebugLine.test(log.raw) ? 'debug' : 'session';
		// We ask every question before the first if. The log types look alike to TypeScript, so a
		// false answer would otherwise narrow log down to nothing for the next question.
		const isDamage = log.isDamageDealt();
		const isCast = [log.isCastBegan(), log.isMajorCooldownUsed(), log.isExtraAttack(), log.isCooldown(), log.isTotem()].some(Boolean);
		const isAura = [log.isAuraEvent(), log.isAuraStacksChange()].some(Boolean);
		const isResource = [log.isResourceChanged(), log.isMaxResources()].some(Boolean);
		if (isDamage) return 'damage';
		if (isCast) return 'casts';
		if (isAura) return 'auras';
		if (isResource) return 'resources';
		return 'other';
	}

	private emit() {
		this.changeEmitter.emit(TypedEvent.nextEventID());
	}
}

const parseKeywords = (query: string): Array<string> => {
	const keywords: Array<string> = [];
	const matchQuotesRegex = /"([^"]+)"|\S+/g;
	let match;
	while ((match = matchQuotesRegex.exec(query))) {
		keywords.push((match[1] ?? match[0]).toLowerCase());
	}
	return keywords;
};
