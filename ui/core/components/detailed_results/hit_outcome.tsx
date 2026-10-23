import { DamageDealtLog } from '../../proto_utils/logs_parser.js';

// How a hit landed, grouped by the color we show it in. Misses, dodges, parries and blocks
// share one color, because for us they all mean the hit lost damage.
export type HitOutcomeKind = 'crit' | 'hit' | 'glance' | 'avoid' | 'tick';

// The outcome words the logs use, see DamageDealtLog.result().
const outcomeKinds: Record<string, HitOutcomeKind> = {
	Crit: 'crit',
	Crush: 'crit',
	Hit: 'hit',
	Glance: 'glance',
	Miss: 'avoid',
	Dodge: 'avoid',
	Parry: 'avoid',
	Block: 'avoid',
	'Blocked Crit': 'avoid',
	Tick: 'tick',
};

export const hitOutcome = (ddl: DamageDealtLog): { label: string; kind: HitOutcomeKind; resisted: number } => {
	const label = ddl.miss
		? 'Miss'
		: ddl.dodge
		? 'Dodge'
		: ddl.parry
		? 'Parry'
		: ddl.glance
		? 'Glance'
		: ddl.blockedCrit
		? 'Blocked Crit'
		: ddl.block
		? 'Block'
		: ddl.crit
		? 'Crit'
		: ddl.crush
		? 'Crush'
		: ddl.tick
		? 'Tick'
		: 'Hit';
	const resisted = ddl.resistPercent;
	const prefix = ddl.isHealing() ? 'Heal ' : ddl.isShielding() ? 'Shield ' : '';
	return { label: prefix + label, kind: outcomeKinds[label], resisted };
};

export const outcomeBadge = (label: string, kind: HitOutcomeKind): HTMLElement =>
	(<span className={`outcome-badge outcome-${kind}`}>{label}</span>) as HTMLElement;

// Turns the outcome word in a rendered damage log line into a badge. We look for the text node
// that holds only the word, because the line also names the ability, which may contain "Hit".
export const badgeOutcomeInLogLine = (line: Element) => {
	if (line.querySelector('.outcome-badge')) return;
	const walker = document.createTreeWalker(line, NodeFilter.SHOW_TEXT);
	for (let node = walker.nextNode(); node; node = walker.nextNode()) {
		const word = node.textContent?.trim() ?? '';
		const kind = outcomeKinds[word];
		if (kind && node.textContent == word) {
			node.parentNode?.replaceChild(outcomeBadge(word, kind), node);
			return;
		}
	}
};
