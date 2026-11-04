// The character frame's Skills tab, as Forever's Camelot/SkillsFrame.xml: your skills under
// their headers (Professions, Weapon Skills, Armor Proficiencies, Languages) in a scrolling list,
// each with its rank bar, and the picked skill's details in the pane on the right. Forever hides
// the Class Skills header.
//
// The game reads these from the server. We build them from small tables by class and race: the
// skills a level 60 of the class has trained, at the level's cap. Only the weapon skill and
// Defense bonuses come from the sim's sheet.

import { atlasTexture, setAtlas } from './atlas';
import { SidePane, progressBar, scrollBox, sliced, type ProgressBar, type ScrollBox } from './character_widgets';
import type { RoomLink } from './link';
import { clickable, div, text } from './panels';

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

type ClassSkills = { weapons: string[]; armor: string[] };

// What each class has trained by level 60 in Classic. Shaman first, the one we play. We list
// two-handed axes and maces for shamans because Forever's enhancement builds use them.
const classSkills: Record<string, ClassSkills> = {
	ClassShaman: {
		weapons: ['Axes', 'Daggers', 'Fist Weapons', 'Maces', 'Staves', 'Two-Handed Axes', 'Two-Handed Maces', 'Unarmed'],
		armor: ['Cloth', 'Leather', 'Mail', 'Shield'],
	},
	ClassWarrior: {
		weapons: ['Axes', 'Bows', 'Crossbows', 'Daggers', 'Fist Weapons', 'Guns', 'Maces', 'Polearms', 'Staves', 'Swords', 'Thrown', 'Two-Handed Axes', 'Two-Handed Maces', 'Two-Handed Swords', 'Unarmed'],
		armor: ['Cloth', 'Leather', 'Mail', 'Plate Mail', 'Shield'],
	},
	ClassPaladin: {
		weapons: ['Axes', 'Maces', 'Polearms', 'Swords', 'Two-Handed Axes', 'Two-Handed Maces', 'Two-Handed Swords', 'Unarmed'],
		armor: ['Cloth', 'Leather', 'Mail', 'Plate Mail', 'Shield'],
	},
	ClassHunter: {
		weapons: ['Axes', 'Bows', 'Crossbows', 'Daggers', 'Fist Weapons', 'Guns', 'Polearms', 'Staves', 'Swords', 'Thrown', 'Two-Handed Axes', 'Two-Handed Swords', 'Unarmed'],
		armor: ['Cloth', 'Leather', 'Mail'],
	},
	ClassRogue: {
		weapons: ['Bows', 'Crossbows', 'Daggers', 'Fist Weapons', 'Guns', 'Maces', 'Swords', 'Thrown', 'Unarmed'],
		armor: ['Cloth', 'Leather'],
	},
	ClassDruid: {
		weapons: ['Daggers', 'Fist Weapons', 'Maces', 'Staves', 'Two-Handed Maces', 'Unarmed'],
		armor: ['Cloth', 'Leather'],
	},
	ClassMage: { weapons: ['Daggers', 'Staves', 'Swords', 'Unarmed', 'Wands'], armor: ['Cloth'] },
	ClassPriest: { weapons: ['Daggers', 'Maces', 'Staves', 'Unarmed', 'Wands'], armor: ['Cloth'] },
	ClassWarlock: { weapons: ['Daggers', 'Staves', 'Swords', 'Unarmed', 'Wands'], armor: ['Cloth'] },
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


// The weapon skills the detail pane gives hit and crit numbers for (WEAPON_SKILL_LINES), and
// which of them are ranged.
const weaponSkillLines = new Set(['Swords', 'Axes', 'Bows', 'Guns', 'Maces', 'Two-Handed Swords', 'Staves', 'Two-Handed Maces', 'Two-Handed Axes', 'Daggers', 'Thrown', 'Crossbows', 'Wands', 'Polearms', 'Unarmed']);
const rangedSkills = new Set(['Bows', 'Guns', 'Thrown', 'Crossbows', 'Wands']);
const raidBossLevel = 63;

// signed is FormatSignedPercent: "+0.20%", "0.00%" or "-0.60%".
function signed(v: number): string {
	const t = `${(Math.abs(v) < 0.005 ? 0 : v).toFixed(2)}%`;
	return v >= 0.005 ? `+${t}` : t;
}

// glancing is the Classic glancing blow chance and damage cut against a target's defense, as the
// sim has them (core/target.go).
function glancing(skill: number, level: number, defense: number): [chance: number, penalty: number] {
	const capped = Math.min(skill, level * 5);
	const chance = Math.max(0, Math.min(0.4, 0.1 + (defense - capped) * 0.02));
	const lo = Math.max(0.01, Math.min(0.91, 1.3 - 0.05 * (defense - skill)));
	const hi = Math.max(0.2, Math.min(0.99, 1.2 - 0.03 * (defense - skill)));
	return [chance, 1 - (lo + hi) / 2];
}

// rankText is the bar's text: "300 / 300", or "300 (+5) / 300" with the bonus in green.
function rankText(s: Skill): Node[] {
	if (!s.modifier) return [document.createTextNode(`${s.rank} / ${s.max}`)];
	const m = document.createElement('span');
	m.className = s.modifier > 0 ? 'bz-green' : 'bz-red';
	m.textContent = `(${s.modifier > 0 ? '+' : ''}${s.modifier})`;
	return [document.createTextNode(`${s.rank} `), m, document.createTextNode(` / ${s.max}`)];
}

// SkillsPage is the Skills tab's content: the list on the left, the details on the right.
export class SkillsPage {
	private headers: Header[] = [];
	private list: ScrollBox;
	private selected?: Skill;
	private readonly pane: SidePane;
	private readonly rankBar: ProgressBar;

	constructor(
		page: HTMLElement,
		private readonly link: RoomLink,
		// expand opens the right pane when you pick a skill while it's collapsed.
		private readonly expand: () => void,
	) {
		// The ScrollBox sits 10, 40 into the left pane (which starts 20 down the frame) and 25, 15
		// from its other corner, with the long lines at its top and bottom.
		this.list = scrollBox(page, { x: 10, y: 60, w: 363, h: 409 }, { x: 378, y: 62, h: 403 });
		this.list.content.classList.add('bz-skill-list');
		atlasTexture(page, 'UI-Character-Info-ScrollLine-Long', { x: 10 + 363 / 2 - 192, y: 56, w: 384, h: 8 }, { cls: 'bz-cs-line' });
		atlasTexture(page, 'UI-Character-Info-ScrollLine-Long', { x: 10 + 363 / 2 - 192, y: 465, w: 384, h: 8 }, { cls: 'bz-cs-line' });
		this.pane = new SidePane(page, { x: 414, y: 34, w: 205, h: 436 });
		this.pane.el.classList.add('bz-right-pane');
		this.rankBar = progressBar(this.pane.extra, { w: 180 });
		this.rankBar.el.style.position = 'relative';
	}

	// refresh rebuilds the list from the sheet, keeping which headers are open.
	refresh() {
		const open = new Map(this.headers.map(h => [h.name, h.open]));
		this.headers = skillList(this.link);
		for (const h of this.headers) h.open = open.get(h.name) ?? true;
		const all = this.headers.flatMap(h => h.skills);
		this.selected = all.find(s => s.name === this.selected?.name) ?? all[0];
		this.render();
	}

	// update fits the scroll bar to the list, once the tab shows (a hidden list has no size).
	update() {
		this.list.update();
	}

	private render() {
		const c = this.list.content;
		c.textContent = '';
		for (const h of this.headers) {
			// SkillsHeaderTemplate: the collapse bar art, the name and a plus or minus.
			const hr = div(c, 'bz-skill-header');
			const bg = div(hr, 'bz-skill-header-bg');
			void sliced(bg, 'common-button-list-collapseExpand', [18, 0, 18, 0], true);
			const hi = div(hr, 'bz-skill-header-hi');
			void sliced(hi, 'common-button-list-collapseExpand', [18, 0, 18, 0], true);
			const name = text(hr, h.name, 'GameFontNormal');
			name.classList.add('bz-skill-header-name');
			const state = div(hr, 'bz-skill-header-state');
			void setAtlas(state, h.open ? 'common-button-list-minus' : 'common-button-list-plus', { useAtlasSize: true });
			clickable(hr, () => {
				h.open = !h.open;
				this.render();
			});
			if (!h.open) continue;
			for (const s of h.skills) {
				// SkillsEntryTemplate: the name on the left, the rank bar on the right, and a soft
				// white band when picked or hovered.
				const r = div(c, 'bz-skill-entry');
				r.classList.toggle('selected', s === this.selected);
				const band = div(r, 'bz-skill-band');
				const l = atlasTexture(band, 'charactercreate-customize-dropdown-linemouseover-side', {}, { cls: 'bz-band-l' });
				const m = atlasTexture(band, 'charactercreate-customize-dropdown-linemouseover-middle', {}, { cls: 'bz-band-m' });
				const rr = atlasTexture(band, 'charactercreate-customize-dropdown-linemouseover-side', {}, { cls: 'bz-band-r', flipX: true });
				void [l, m, rr];
				const name = text(r, s.name, 'GameFontHighlight');
				name.classList.add('bz-skill-name');
				const bar = progressBar(r, { w: 160 });
				bar.el.classList.add('bz-skill-bar');
				bar.set(s.rank / (s.max || 1), rankText(s));
				clickable(r, () => {
					this.selected = s;
					this.expand();
					this.render();
				});
			}
		}
		this.list.update();
		this.renderDetail();
	}

	// renderDetail fills the right pane with the picked skill (SkillDetailFrameMixin:Refresh).
	private renderDetail() {
		const s = this.selected;
		const p = this.pane;
		if (!s) {
			p.setEmpty('Select a skill to view its details.');
			this.rankBar.el.style.display = 'none';
			return;
		}
		p.clearEmpty();
		this.rankBar.el.style.display = '';
		const weapon = weaponSkillLines.has(s.name);
		p.setTitle(s.name);
		this.rankBar.set(s.rank / (s.max || 1), rankText(s));
		p.setDescription(s.description, weapon ? 40 : 120);
		p.reset();
		if (!weapon) return;
		const level = this.link.sheet()?.level ?? this.link.loaded.level;
		const skill = s.rank + s.modifier;
		const chance = (offset: number) => signed((skill - (level + offset) * 5) * 0.04);
		const ranged = rangedSkills.has(s.name);
		p.addSpacer(6);
		p.addCategory('Equal-Level Enemy');
		p.addWrapped(ranged ? hitLines(chance(0), chance(0)) : meleeLines(chance(0), chance(0)));
		p.addSpacer(8);
		p.addCategory('Against Raid Bosses');
		const boss = chance(raidBossLevel - level);
		if (ranged) p.addWrapped(hitLines(boss, boss));
		else {
			const [g, cut] = glancing(skill, level, raidBossLevel * 5);
			p.addWrapped([...meleeLines(boss, boss), '\n\n', { white: 'Glancing Blows' }, ` occur ${Math.round(g * 100)}% of the time and deal ${Math.round(cut * 100)}% less damage`]);
		}
	}
}

// The detail rows' text (WEAPON_SKILL_DETAIL_*): gold with the key words in white.
const meleeLines = (hit: string, crit: string) => [
	'Chance to ',
	{ white: 'Hit' },
	', and to avoid being ',
	{ white: 'Dodged' },
	' or ',
	{ white: 'Parried' },
	`: ${hit}\n\nChance to `,
	{ white: 'Critically Hit' },
	`: ${crit}`,
];
const hitLines = (hit: string, crit: string) => ['Chance to ', { white: 'Hit' }, `: ${hit}\n\nChance to `, { white: 'Critically Hit' }, `: ${crit}`];
