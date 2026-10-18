import tippy from 'tippy.js';

import { UIItem, UIItem_HiddenStats } from '../../proto/ui';

// What we say about an item whose stats Forever still hides, by where the sim's stats
// came from (gen_db's MarkHiddenStats). An item without a mark has stats the game has
// published.
const HIDDEN_STATS_WORDS: Partial<Record<UIItem_HiddenStats, string>> = {
	[UIItem_HiddenStats.CLASSIC]: "Forever still hides this item's stats. The sim uses Classic's, which may have changed.",
	[UIItem_HiddenStats.PARTIAL]: "Forever still hides this item's stats. The sim only knows part of them.",
	[UIItem_HiddenStats.WOWHEAD]: "Forever still hides this item's stats. The sim uses Wowhead's, which nothing in game has confirmed.",
};

// The unconfirmed mark for an item: an orange question mark whose tooltip says where
// the sim's stats came from. Null when the game has published the item's stats.
//
// The item's own tooltip is Wowhead's, which we can't add a line to, so the words are
// on the mark.
export const hiddenStatsBadge = (item: UIItem, className: string): HTMLElement | null => {
	const words = HIDDEN_STATS_WORDS[item.hiddenStats];
	if (!words) return null;
	const badge = <i className={`fas fa-circle-question hidden-stats-badge ${className}`} />;
	tippy(badge, { content: words });
	return badge as HTMLElement;
};
