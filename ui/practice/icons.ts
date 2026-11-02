// Spell and item icons, from Wowhead's tooltip service like the class sims use. The names are
// cached in localStorage, so a second visit needs no requests.

const cacheKey = 'practice.icons.v1';
let cache: Record<string, string> = {};
try {
	cache = JSON.parse(localStorage.getItem(cacheKey) ?? '{}');
} catch {
	cache = {};
}
const pending = new Map<string, Promise<string>>();

function save() {
	try {
		localStorage.setItem(cacheKey, JSON.stringify(cache));
	} catch {
		// The icons still work, they're only fetched again next time.
	}
}

const fallbackIcon = 'inv_misc_questionmark';

export function iconUrl(name: string): string {
	return `https://wow.zamimg.com/images/wow/icons/large/${name}.jpg`;
}

// iconName gives the icon name of a spell or item, fetching it once.
export function iconName(kind: 'spell' | 'item', id: number): Promise<string> {
	const key = `${kind}:${id}`;
	if (cache[key]) return Promise.resolve(cache[key]);
	let p = pending.get(key);
	if (!p) {
		p = fetchIcon(kind, id).then(name => {
			cache[key] = name;
			save();
			return name;
		});
		pending.set(key, p);
	}
	return p;
}

// The melee swing, the one action the room shows that isn't a spell or item.
export const meleeIcon = 'inv_sword_04';

// The practice server looks icons up on Wowhead once and caches them, and answers the question
// mark for an id Wowhead doesn't know.
async function fetchIcon(kind: 'spell' | 'item', id: number): Promise<string> {
	try {
		const r = await fetch(`/api/icon/${kind}/${id}`);
		const json = await r.json();
		if (json.name) names[`${kind}:${id}`] = json.name;
		if (json.icon) return json.icon as string;
	} catch {
		// Offline: the question mark.
	}
	return fallbackIcon;
}

// Names Wowhead gave with the icons, for items the toolkit only knows by id.
export const names: Record<string, string> = {};

// setIcon fills an element's background with an icon.
export function setIcon(el: HTMLElement, kind: 'spell' | 'item', id: number) {
	const key = `${kind}:${id}`;
	el.dataset.icon = key;
	iconName(kind, id).then(name => {
		if (el.dataset.icon === key) el.style.backgroundImage = `url(${iconUrl(name)})`;
	});
}
