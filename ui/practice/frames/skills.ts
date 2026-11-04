// The character frame's Skills tab: the Classic SkillFrame. Your skills under their headers (Class
// Skills, Professions, Weapon Skills, Armor Proficiencies, Languages) as bars, and the picked
// skill's bar and description under the list.
//
// The game reads these from the server. We build them from small tables by class and race: the
// skills a level 60 of the class has trained, at the level's cap. Only the weapon skill and
// Defense bonuses come from the sim's sheet.

import type { RoomLink } from './link';
import { button, clickable, div, scrollFrame, text, texture } from './panels';
import { SND, sound, tex, texVars, glow, TX } from './assets';

const plusUp = 130838;
const plusHighlight = 130837;
const minusUp = 130821;
const horizontalBar = 130968; // classtrainerframe/ui-classtrainer-horizontalbar

// The weapon skills the sheet has a bonus for, by the PseudoStat enum.
const weaponPseudo: Record<string, number> = {
	Unarmed: 6,
	Daggers: 7,
	Swords: 8,
	Maces: 9,
	Axes: 10,
	'Two-Handed Swords': 11,
	'Two-Handed Maces': 12,
	'Two-Handed Axes': 13,
	Polearms: 14,
	Staves: 15,
	Bows: 16,
	Crossbows: 17,
	Guns: 18,
	Thrown: 19,
};

type ClassSkills = { spec: string[]; weapons: string[]; armor: string[] };

// What each class has trained by level 60 in Classic. Shaman first, the one we play. We list
// two-handed axes and maces for shamans because Forever's enhancement builds use them.
const classSkills: Record<string, ClassSkills> = {
	ClassShaman: {
		spec: ['Elemental Combat', 'Enhancement', 'Restoration'],
		weapons: ['Axes', 'Daggers', 'Fist Weapons', 'Maces', 'Staves', 'Two-Handed Axes', 'Two-Handed Maces', 'Unarmed'],
		armor: ['Cloth', 'Leather', 'Mail', 'Shield'],
	},
	ClassWarrior: {
		spec: ['Arms', 'Fury', 'Protection'],
		weapons: ['Axes', 'Bows', 'Crossbows', 'Daggers', 'Fist Weapons', 'Guns', 'Maces', 'Polearms', 'Staves', 'Swords', 'Thrown', 'Two-Handed Axes', 'Two-Handed Maces', 'Two-Handed Swords', 'Unarmed'],
		armor: ['Cloth', 'Leather', 'Mail', 'Plate Mail', 'Shield'],
	},
	ClassPaladin: {
		spec: ['Holy', 'Protection', 'Retribution'],
		weapons: ['Axes', 'Maces', 'Polearms', 'Swords', 'Two-Handed Axes', 'Two-Handed Maces', 'Two-Handed Swords', 'Unarmed'],
		armor: ['Cloth', 'Leather', 'Mail', 'Plate Mail', 'Shield'],
	},
	ClassHunter: {
		spec: ['Beast Mastery', 'Marksmanship', 'Survival'],
		weapons: ['Axes', 'Bows', 'Crossbows', 'Daggers', 'Fist Weapons', 'Guns', 'Polearms', 'Staves', 'Swords', 'Thrown', 'Two-Handed Axes', 'Two-Handed Swords', 'Unarmed'],
		armor: ['Cloth', 'Leather', 'Mail'],
	},
	ClassRogue: {
		spec: ['Assassination', 'Combat', 'Subtlety'],
		weapons: ['Bows', 'Crossbows', 'Daggers', 'Fist Weapons', 'Guns', 'Maces', 'Swords', 'Thrown', 'Unarmed'],
		armor: ['Cloth', 'Leather'],
	},
	ClassDruid: {
		spec: ['Balance', 'Feral Combat', 'Restoration'],
		weapons: ['Daggers', 'Fist Weapons', 'Maces', 'Staves', 'Two-Handed Maces', 'Unarmed'],
		armor: ['Cloth', 'Leather'],
	},
	ClassMage: { spec: ['Arcane', 'Fire', 'Frost'], weapons: ['Daggers', 'Staves', 'Swords', 'Unarmed', 'Wands'], armor: ['Cloth'] },
	ClassPriest: { spec: ['Discipline', 'Holy', 'Shadow Magic'], weapons: ['Daggers', 'Maces', 'Staves', 'Unarmed', 'Wands'], armor: ['Cloth'] },
	ClassWarlock: { spec: ['Affliction', 'Demonology', 'Destruction'], weapons: ['Daggers', 'Staves', 'Swords', 'Unarmed', 'Wands'], armor: ['Cloth'] },
};

// The languages each race speaks.
const raceLanguages: Record<string, string[]> = {
	RaceOrc: ['Orcish'],
	RaceUndead: ['Gutterspeak', 'Orcish'],
	RaceTauren: ['Orcish', 'Taurahe'],
	RaceTroll: ['Orcish', 'Troll'],
	RaceHuman: ['Common'],
	RaceDwarf: ['Common', 'Dwarven'],
	RaceNightElf: ['Common', 'Darnassian'],
	RaceGnome: ['Common', 'Gnomish'],
};

const professionNames: Record<string, string> = {
	Alchemy: 'Alchemy',
	Blacksmithing: 'Blacksmithing',
	Enchanting: 'Enchanting',
	Engineering: 'Engineering',
	Herbalism: 'Herbalism',
	Leatherworking: 'Leatherworking',
	Mining: 'Mining',
	Skinning: 'Skinning',
	Tailoring: 'Tailoring',
};

// The client's skill descriptions (SkillLine), for the ones that have one.
const descriptions: Record<string, string> = {
	Defense: 'Higher defense makes you harder to hit and makes monsters less likely to land a crushing blow.',
	Unarmed: 'Higher skill increases your chance to hit.',
	'Fist Weapons': 'Allows for the use of fist weapons.  Chance to hit is determined by the Unarmed skill.',
	Cloth: 'Allows the wearing of cloth armor.',
	Leather: 'Allows the wearing of leather armor.',
	Mail: 'Allows the wearing of mail armor.',
	'Plate Mail': 'Allows the wearing of plate armor.',
	Shield: 'Allows the use of shields.',
};
const weaponDescription = 'Higher weapon skill increases your chance to hit.';

type Skill = { name: string; rank: number; max: number; modifier: number; description?: string };
type Header = { name: string; skills: Skill[]; open: boolean };

// skillList builds your skills, by header, from the class and race tables and the sheet.
function skillList(link: RoomLink): Header[] {
	const { loaded } = link;
	const sheet = link.sheet();
	const level = sheet?.level ?? loaded.level;
	const cap = level * 5;
	const cls = classSkills[loaded.class] ?? classSkills.ClassShaman;
	const headers: Header[] = [];
	const add = (name: string, skills: Skill[]) => skills.length && headers.push({ name, skills, open: true });

	add(
		'Class Skills',
		cls.spec.map(name => ({ name, rank: 1, max: 1, modifier: 0 })),
	);
	const player = (loaded.request as { raid?: { parties?: { players?: { profession1?: string; profession2?: string }[] }[] } })?.raid?.parties?.[0]?.players?.[0];
	const professions = [player?.profession1, player?.profession2].filter((p): p is string => !!p && !!professionNames[p]);
	// The sim knows which professions you have but not their skill. We show them at the cap.
	const professionCap = Math.min(300, cap);
	add(
		'Professions',
		professions.map(p => ({ name: professionNames[p], rank: professionCap, max: professionCap, modifier: 0 })),
	);
	const weapons: Skill[] = cls.weapons.map(name => {
		const p = weaponPseudo[name];
		const modifier = p !== undefined ? Math.round(sheet?.pseudo?.[p] ?? 0) : 0;
		return { name, rank: cap, max: cap, modifier, description: descriptions[name] ?? weaponDescription };
	});
	weapons.push({ name: 'Defense', rank: cap, max: cap, modifier: Math.round(sheet?.stats?.[28] ?? 0), description: descriptions.Defense });
	if (sheet?.weapons?.some(w => w.slot === 'offHand')) weapons.push({ name: 'Dual Wield', rank: 1, max: 1, modifier: 0 });
	weapons.sort((a, b) => a.name.localeCompare(b.name));
	add('Weapon Skills', weapons);
	add(
		'Armor Proficiencies',
		cls.armor.map(name => ({ name, rank: 1, max: 1, modifier: 0, description: descriptions[name] })),
	);
	add(
		'Languages',
		(raceLanguages[loaded.race] ?? []).map(l => ({ name: `Language: ${l}`, rank: 300, max: 300, modifier: 0 })),
	);
	return headers;
}

// The bar colors from SkillFrame.lua: blue for a skill with ranks, gray for one you just have.
const rankColors = { fill: 'rgb(0,0,255)', back: 'rgb(0,0,191)' };
const plainColors = { fill: 'rgb(128,128,128)', back: 'rgb(64,64,64)' };

type Bar = { el: HTMLDivElement; set(s: Skill): void };

// skillBar is a SkillRankFrameTemplate: the bar with the skill's name on the left and its rank on
// the right, in a thin border.
function skillBar(parent: HTMLElement, box: { x: number; y: number; w: number }): Bar {
	const el = div(parent, 'bz-skill-bar', { ...box, h: 12 });
	const back = div(el, 'bz-skill-bar-bg');
	const fill = div(el, 'bz-skill-bar-fill');
	div(el, 'bz-skill-bar-border');
	const name = text(el, '', 'GameFontHighlightSmall', { x: 6 });
	const rank = text(el, '', 'GameFontHighlightSmall');
	rank.classList.add('bz-skill-rank');
	return {
		el,
		set(s) {
			const c = s.max > 1 ? rankColors : plainColors;
			back.style.backgroundColor = c.back;
			fill.style.backgroundColor = c.fill;
			fill.style.width = `${Math.min(1, s.rank / (s.max || 1)) * 100}%`;
			name.textContent = s.name;
			rank.textContent = '';
			if (s.max <= 1) return;
			// The game shows a bonus as "300 (+5)/300", the bonus in green.
			if (!s.modifier) rank.textContent = `${s.rank}/${s.max}`;
			else {
				rank.append(`${s.rank} (`);
				const m = document.createElement('span');
				m.className = s.modifier > 0 ? 'up' : 'down';
				m.textContent = s.modifier > 0 ? `+${s.modifier}` : String(s.modifier);
				rank.append(m, `)/${s.max}`);
			}
		},
	};
}

// SkillsPage is the Skills tab's content.
export class SkillsPage {
	private headers: Header[] = [];
	private list!: ReturnType<typeof scrollFrame>;
	private selected?: Skill;
	private detailBar!: Bar;
	private detailText!: HTMLDivElement;
	private allToggle!: HTMLDivElement;

	constructor(
		private readonly page: HTMLElement,
		private readonly link: RoomLink,
		close: () => void,
	) {
		void texVars({
			'bz-skill-bar': () => tex(TX.skillsBar),
			'bz-skill-border': () => tex(TX.skillsBarBorder),
			'bz-skill-border-highlight': () => tex(TX.skillsBarBorderHighlight),
			'bz-skill-plus': () => tex(plusUp),
			'bz-skill-minus': () => tex(minusUp),
			'bz-skill-plus-highlight': () => glow(plusHighlight),
		});
		// The "All" button opens or closes every header.
		const all = div(page, 'bz-skill-row bz-skill-header', { x: 22, y: 76, w: 60 });
		this.allToggle = div(all, 'bz-skill-toggle');
		text(all, 'All', 'GameFontNormal');
		clickable(all, () => {
			const open = !this.headers.every(h => h.open);
			for (const h of this.headers) h.open = open;
			this.render();
		});
		this.list = scrollFrame(page, { x: 20, y: 96, w: 299, h: 144 });
		// The bar between the list and the detail.
		texture(page, horizontalBar, { x: 15, y: 250, w: 256, h: 16 }, { coords: [0, 1, 0, 0.25], size: [256, 64] });
		texture(page, horizontalBar, { x: 271, y: 250, w: 75, h: 16 }, { coords: [0, 0.29296875, 0.25, 0.5], size: [256, 64] });
		// The detail under the list: the picked skill's bar and what it does.
		this.detailBar = skillBar(page, { x: 30, y: 272, w: 270 });
		this.detailText = text(page, '', 'GameFontHighlightSmall', { x: 30, y: 294, w: 280 });
		this.detailText.classList.add('bz-skill-desc');
		button(page, 'Close', { x: 264, y: 409, w: 80, h: 22 }, close);
	}

	// refresh rebuilds the list from the sheet, keeping which headers are open.
	refresh() {
		const open = new Map(this.headers.map(h => [h.name, h.open]));
		this.headers = skillList(this.link);
		for (const h of this.headers) h.open = open.get(h.name) ?? true;
		const names = new Set(this.headers.flatMap(h => h.skills.map(s => s.name)));
		if (!this.selected || !names.has(this.selected.name)) this.selected = this.headers.find(h => h.name === 'Weapon Skills')?.skills[0] ?? this.headers[0]?.skills[0];
		else this.selected = this.headers.flatMap(h => h.skills).find(s => s.name === this.selected!.name);
		this.render();
	}

	// update fits the scroll bar to the list, once the tab shows (a hidden list has no size).
	update() {
		this.list.update();
	}

	private render() {
		const c = this.list.content;
		c.textContent = '';
		const row = 15;
		let y = 0;
		this.allToggle.style.backgroundImage = this.headers.every(h => h.open) ? 'var(--bz-skill-minus)' : 'var(--bz-skill-plus)';
		for (const h of this.headers) {
			const hr = div(c, 'bz-skill-row bz-skill-header', { y, w: 290 });
			const t = div(hr, 'bz-skill-toggle');
			t.style.backgroundImage = h.open ? 'var(--bz-skill-minus)' : 'var(--bz-skill-plus)';
			text(hr, h.name, 'GameFontNormal');
			clickable(hr, () => {
				h.open = !h.open;
				this.render();
			});
			y += row;
			if (!h.open) continue;
			for (const s of h.skills) {
				const r = div(c, 'bz-skill-row', { y, w: 290 });
				r.classList.toggle('selected', s === this.selected);
				skillBar(r, { x: 20, y: 2, w: 268 }).set(s);
				clickable(r, () => {
					this.selected = s;
					sound(SND.menuOption);
					this.render();
				});
				y += row;
			}
		}
		c.style.height = `${y}px`;
		this.list.update();
		const s = this.selected;
		this.detailBar.el.style.display = s ? '' : 'none';
		if (s) this.detailBar.set(s);
		this.detailText.textContent = s?.description ?? '';
	}
}
