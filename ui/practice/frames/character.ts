// The character frame (C): the Classic Era CharacterFrame. The Character tab is the paper doll:
// your worn items in their slots around your model, the resistances, and two stat boxes you
// switch between Base Stats, Melee, Ranged, Spell and Defenses. The Skills tab is in skills.ts.
// The Reputation tab has no factions, since the room has none.
//
// Sizes and anchors follow the Classic PaperDollFrame.xml. The numbers come from the sim's
// character sheet (link.sheet()).

import './character.css';
import type { Sheet } from '../protocol';
import type { RoomLink } from './link';
import { clickable, closeButton, div, itemButton, type ItemButton, Panel, selectTab, setTexture, tab, text, texture } from './panels';
import { glow, SND, sound, tex, texVars, TX } from './assets';
import { ModelView, portrait } from './model_view';
import { SkillsPage } from './skills';

const labelFrame = 131143; // glues/charactercreate/charactercreate-labelframe
const arrowUp = 130952; // chatframe/ui-chaticon-scrolldown-up
const arrowDown = 130951;

// The tabs left to right, by the game's tab number (0 Character, 1 Skills, 2 Reputation).
const tabOrder: [tab: number, label: string][] = [
	[0, 'Character'],
	[2, 'Reputation'],
	[1, 'Skills'],
];

type SlotDef = { name: string; item?: number; texture: number };

// The paper doll's slots: eight down each side and the weapons at the bottom. Item is the sheet's
// ItemSlot index. The sim has no shirt or tabard.
const leftSlots: SlotDef[] = [
	{ name: 'Head', item: 0, texture: TX.slotHead },
	{ name: 'Neck', item: 1, texture: TX.slotNeck },
	{ name: 'Shoulder', item: 2, texture: TX.slotShoulder },
	{ name: 'Back', item: 3, texture: TX.slotBack },
	{ name: 'Chest', item: 4, texture: TX.slotChest },
	{ name: 'Shirt', texture: TX.slotShirt },
	{ name: 'Tabard', texture: TX.slotTabard },
	{ name: 'Wrist', item: 5, texture: TX.slotWrists },
];
const rightSlots: SlotDef[] = [
	{ name: 'Hands', item: 6, texture: TX.slotHands },
	{ name: 'Waist', item: 7, texture: TX.slotWaist },
	{ name: 'Legs', item: 8, texture: TX.slotLegs },
	{ name: 'Feet', item: 9, texture: TX.slotFeet },
	{ name: 'Finger', item: 10, texture: TX.slotFinger },
	{ name: 'Finger', item: 11, texture: TX.slotFinger },
	{ name: 'Trinket', item: 12, texture: TX.slotTrinket },
	{ name: 'Trinket', item: 13, texture: TX.slotTrinket },
];

// The classes whose ranged slot holds a relic (a totem, libram or idol) in Classic Era.
const relicClasses = new Set(['ClassShaman', 'ClassPaladin', 'ClassDruid']);

// The class icons in targetingframe/ui-classes-circles (CLASS_ICON_TCOORDS), the portrait until
// the model loads.
const classCircle: Record<string, [number, number, number, number]> = {
	ClassWarrior: [0, 0.25, 0, 0.25],
	ClassMage: [0.25, 0.49609375, 0, 0.25],
	ClassRogue: [0.49609375, 0.7421875, 0, 0.25],
	ClassDruid: [0.7421875, 0.98828125, 0, 0.25],
	ClassHunter: [0, 0.25, 0.25, 0.5],
	ClassShaman: [0.25, 0.49609375, 0.25, 0.5],
	ClassPriest: [0.49609375, 0.7421875, 0.25, 0.5],
	ClassWarlock: [0.7421875, 0.98828125, 0.25, 0.5],
	ClassPaladin: [0, 0.25, 0.5, 0.75],
};

// The resistances top to bottom: the stat index and the icon's rows in
// paperdollinfoframe/ui-character-resistanceicons.
const resistances: [name: string, stat: number, top: number, bottom: number][] = [
	['Arcane', 35, 0.2265625, 0.3359375],
	['Fire', 36, 0, 0.11328125],
	['Nature', 38, 0.11328125, 0.2265625],
	['Frost', 37, 0.3359375, 0.4453125],
	['Shadow', 39, 0.4453125, 0.5546875],
];

// art lays a panel's four art pieces at 2, -1 like the FrameXML. The pieces overlap by one unit,
// or the seams between them show as faint lines at a fractional UI scale.
function art(parent: HTMLElement, files: [number, number, number, number]) {
	const [tl, tr, bl, br] = files;
	texture(parent, tl, { x: 2, y: -1, w: 257, h: 257 });
	texture(parent, tr, { x: 258, y: -1, w: 128, h: 257 });
	texture(parent, bl, { x: 2, y: 255, w: 257, h: 256 });
	texture(parent, br, { x: 258, y: 255, w: 128, h: 256 });
}

// The sim's enum names in the game's words: "RaceNightElf" is "Night Elf".
function enumName(s: string, prefix: string): string {
	return s.replace(prefix, '').replace(/([a-z])([A-Z])/g, '$1 $2');
}

type Tip = { title: string; lines?: string[] };
type StatLine = { label: string; value: string; color?: 'up' | 'down'; tip?: Tip };
// Only warriors, paladins, hunters and rogues learn Parry in Classic, so the game shows 0% for
// the rest. The sim gives every unit the 5% base.
const parryClasses = new Set(['ClassWarrior', 'ClassPaladin', 'ClassHunter', 'ClassRogue']);
function parry(s: Sheet, f: Facts): number {
	return parryClasses.has(f.cls) ? (s.stats[32] ?? 0) : 0;
}

type Category = { key: string; name: string; lines: (s: Sheet, f: Facts) => StatLine[] };

// What the stat lines need besides the sheet.
type Facts = { cls: string; level: number };

const round = (v: number) => Math.round(v + 1e-6);
const pct = (v: number) => `${v.toFixed(2)}%`;

// colorOf is green for a stat above its base (gear and buffs), red below it.
function colorOf(s: Sheet, i: number): 'up' | 'down' | undefined {
	const d = (s.stats[i] ?? 0) - (s.base[i] ?? 0);
	return d > 0.5 ? 'up' : d < -0.5 ? 'down' : undefined;
}

// statTitle is the game's "Strength 120 (100+20)" title line.
function statTitle(s: Sheet, name: string, i: number): string {
	const v = round(s.stats[i] ?? 0);
	const b = round(s.base[i] ?? 0);
	if (v === b) return `${name} ${v}`;
	return `${name} ${v} (${b}${v > b ? '+' : '-'}${Math.abs(v - b)})`;
}

const apPerStrength: Record<string, number> = { ClassWarrior: 2, ClassPaladin: 2, ClassShaman: 2, ClassDruid: 2 };
const manaUsers = new Set(['ClassShaman', 'ClassPaladin', 'ClassHunter', 'ClassPriest', 'ClassMage', 'ClassWarlock', 'ClassDruid']);

// healthFrom is the game's first 20 points give 1 each, the rest give `per` each.
const firstTwenty = (v: number, per: number) => Math.min(20, v) + Math.max(0, v - 20) * per;

function baseStat(name: string, i: number, tip: (v: number, f: Facts) => string[]): (s: Sheet, f: Facts) => StatLine {
	return (s, f) => ({
		label: `${name}:`,
		value: String(round(s.stats[i] ?? 0)),
		color: colorOf(s, i),
		tip: { title: statTitle(s, name, i), lines: tip(round(s.stats[i] ?? 0), f) },
	});
}

function armorLine(s: Sheet, f: Facts): StatLine {
	const armor = round(s.stats[26] ?? 0);
	const cut = (armor / (armor + 400 + 85 * f.level)) * 100;
	return { label: 'Armor:', value: String(armor), color: colorOf(s, 26), tip: { title: statTitle(s, 'Armor', 26), lines: [`Reduces Physical Damage taken by ${cut.toFixed(2)}%`] } };
}

type Weapon = NonNullable<Sheet['weapons']>[number];

function damageRange(w: Weapon, mult: number): [number, number] {
	return [Math.max(1, Math.floor(w.min * mult)), Math.max(1, Math.ceil(w.max * mult))];
}

function weaponTip(name: string, w: Weapon, mult: number): string[] {
	const [lo, hi] = damageRange(w, mult);
	return [name, `Attack Speed (seconds): ${w.speed.toFixed(2)}`, `Damage: ${lo} - ${hi}`, `Damage per second: ${((lo + hi) / 2 / w.speed).toFixed(1)}`];
}

function damageLines(s: Sheet, ranged: boolean, power: number, powerName: string): StatLine[] {
	const ws = s.weapons ?? [];
	const mh = ws.find(w => w.slot === (ranged ? 'ranged' : 'mainHand'));
	const oh = ranged ? undefined : ws.find(w => w.slot === 'offHand');
	const mult = s.damageMultiplier || 1;
	const dmgColor = mult > 1.0001 ? 'up' : mult < 0.9999 ? 'down' : undefined;
	const lines: StatLine[] = [];
	if (mh) {
		const [lo, hi] = damageRange(mh, mult);
		const tip = weaponTip(ranged ? 'Ranged' : 'Main Hand', mh, mult);
		if (oh) tip.push(...weaponTip('Off Hand', oh, mult));
		lines.push({ label: 'Damage:', value: `${lo} - ${hi}`, color: dmgColor, tip: { title: tip[0], lines: tip.slice(1) } });
		const speed = oh ? `${mh.speed.toFixed(2)} / ${oh.speed.toFixed(2)}` : mh.speed.toFixed(2);
		lines.push({ label: 'Speed:', value: speed, tip: { title: `Attack Speed ${speed}` } });
	} else {
		lines.push({ label: 'Damage:', value: 'N/A' }, { label: 'Speed:', value: 'N/A' });
	}
	const ap = round(s.stats[power] ?? 0);
	lines.push({
		label: 'Power:',
		value: String(ap),
		color: colorOf(s, power),
		tip: { title: statTitle(s, powerName, power), lines: [`Increases damage with ${ranged ? 'ranged' : 'melee'} weapons by ${(ap / 14).toFixed(1)} damage per second.`] },
	});
	const hit = s.stats[18] ?? 0;
	lines.push({ label: 'Hit:', value: `+${round(hit * 100) / 100}%`, tip: { title: `Hit Chance +${hit.toFixed(2)}%`, lines: [`Increases your chance to hit with ${ranged ? 'ranged' : 'melee'} attacks.`] } });
	const crit = s.stats[19] ?? 0;
	lines.push({ label: 'Crit Chance:', value: pct(crit), tip: { title: `${ranged ? 'Ranged' : 'Melee'} Critical Hit Chance ${pct(crit)}` } });
	return lines;
}

const schools: [string, number][] = [
	['Holy', 9],
	['Fire', 7],
	['Nature', 10],
	['Frost', 8],
	['Shadow', 11],
	['Arcane', 6],
];

const categories: Category[] = [
	{
		key: 'base',
		name: 'Base Stats',
		lines: (s, f) => [
			baseStat('Strength', 0, (v, f) => [`Increases Attack Power by ${v * (apPerStrength[f.cls] ?? 1)}`])(s, f),
			baseStat('Agility', 1, v => [`Increases Armor by ${v * 2}`])(s, f),
			baseStat('Stamina', 2, v => [`Increases Health by ${firstTwenty(v, 10)}`])(s, f),
			baseStat('Intellect', 3, (v, f) => (manaUsers.has(f.cls) ? [`Increases Mana by ${firstTwenty(v, 15)}`] : []))(s, f),
			baseStat('Spirit', 4, () => ['Increases Health and Mana regeneration'])(s, f),
			armorLine(s, f),
		],
	},
	{ key: 'melee', name: 'Melee', lines: s => damageLines(s, false, 17, 'Melee Attack Power') },
	{ key: 'ranged', name: 'Ranged', lines: s => damageLines(s, true, 27, 'Ranged Attack Power') },
	{
		key: 'spell',
		name: 'Spell',
		lines: s => {
			const generic = (s.stats[5] ?? 0) + (s.stats[42] ?? 0);
			const heal = (s.stats[5] ?? 0) + (s.stats[41] ?? 0);
			const mp5 = round(s.stats[12] ?? 0);
			return [
				{
					label: 'Bonus Damage:',
					value: String(round(generic)),
					tip: { title: `Bonus Damage ${round(generic)}`, lines: schools.map(([n, i]) => `${n}: ${round(generic + (s.stats[i] ?? 0))}`) },
				},
				{ label: 'Bonus Healing:', value: String(round(heal)), tip: { title: `Bonus Healing ${round(heal)}`, lines: ['Increases your healing by up to this much.'] } },
				{ label: 'Hit:', value: `+${round((s.stats[13] ?? 0) * 100) / 100}%`, tip: { title: `Spell Hit Chance +${(s.stats[13] ?? 0).toFixed(2)}%` } },
				{ label: 'Crit Chance:', value: pct(s.stats[14] ?? 0), tip: { title: `Spell Critical Hit Chance ${pct(s.stats[14] ?? 0)}` } },
				{ label: 'Mana Regen:', value: String(mp5), tip: { title: `Mana Regen ${mp5}`, lines: [`${mp5} mana per 5s from gear and buffs, also while casting.`] } },
			];
		},
	},
	{
		key: 'defense',
		name: 'Defenses',
		lines: (s, f) => {
			const cap = f.level * 5;
			const bonus = round(s.stats[28] ?? 0);
			const def: StatLine = {
				label: 'Defense:',
				value: String(cap + bonus),
				color: bonus > 0 ? 'up' : bonus < 0 ? 'down' : undefined,
				tip: {
					title: bonus ? `Defense ${cap + bonus} (${cap}${bonus > 0 ? '+' : ''}${bonus})` : `Defense ${cap}`,
					lines: bonus ? [`Increases chance to Dodge, Block and Parry by ${(bonus * 0.04).toFixed(2)}%`, `Decreases chance to be hit and critically hit by ${(bonus * 0.04).toFixed(2)}%`] : [],
				},
			};
			return [
				armorLine(s, f),
				def,
				{ label: 'Dodge:', value: pct(s.stats[31] ?? 0), tip: { title: `Dodge Chance ${pct(s.stats[31] ?? 0)}` } },
				{ label: 'Parry:', value: pct(parry(s, f)), tip: { title: `Parry Chance ${pct(parry(s, f))}` } },
				{ label: 'Block:', value: pct(s.stats[29] ?? 0), tip: { title: `Block Chance ${pct(s.stats[29] ?? 0)}` } },
			];
		},
	},
];

function stored(key: string, fallback: string): string {
	try {
		return localStorage.getItem(key) ?? fallback;
	} catch {
		return fallback;
	}
}

function store(key: string, value: string) {
	try {
		localStorage.setItem(key, value);
	} catch {
		// Only this visit keeps it.
	}
}

// statDropdown is a UIDropDownMenuTemplate with a list of choices under it.
function statDropdown(parent: HTMLElement, box: { x: number; y: number }, middle: number, options: [string, string][], value: string, onChange: (v: string) => void) {
	const width = 25 + middle + 25;
	const e = div(parent, 'bz-dd', { ...box, w: width });
	texture(e, labelFrame, { x: 0, y: -17, w: 25, h: 64 }, { coords: [0, 0.1953125, 0, 1], size: [128, 64] });
	texture(e, labelFrame, { x: 25, y: -17, w: middle, h: 64 }, { coords: [0.1953125, 0.8046875, 0, 1], size: [128, 64] });
	texture(e, labelFrame, { x: 25 + middle, y: -17, w: 25, h: 64 }, { coords: [0.8046875, 1, 0, 1], size: [128, 64] });
	const label = text(e, '', 'GameFontHighlightSmall', { x: 0, w: width - 43 });
	label.classList.add('bz-dd-text');
	const arrow = div(e, 'bz-dd-button', { x: width - 16 - 24 });
	arrow.style.backgroundImage = 'var(--bz-dd-arrow)';
	let list: HTMLDivElement | undefined;
	const closeList = () => {
		list?.remove();
		list = undefined;
		removeEventListener('mousedown', outside);
	};
	const outside = (ev: MouseEvent) => {
		if (list && !list.contains(ev.target as Node)) closeList();
	};
	const show = (v: string) => (label.textContent = options.find(o => o[0] === v)?.[1] ?? '');
	show(value);
	clickable(e, () => {
		if (list) {
			closeList();
			return;
		}
		sound(SND.menuOption);
		list = div(parent, 'bz-dd-list', { x: box.x + 16, y: box.y + 26 });
		for (const [v, l] of options) {
			const item = div(list, 'bz-dd-item');
			item.classList.toggle('checked', v === value);
			text(item, l, 'GameFontHighlightSmall');
			clickable(item, () => {
				value = v;
				show(v);
				closeList();
				sound(SND.menuOption);
				onChange(v);
			});
		}
		addEventListener('mousedown', outside);
	});
	return { el: e, close: closeList };
}

type StatBox = { rows: { el: HTMLDivElement; label: HTMLDivElement; value: HTMLDivElement; tip?: Tip }[]; key: string };

export class CharacterFrame extends Panel {
	readonly width = 384;
	readonly height = 512;
	stackWidth = 353;
	openSound = SND.characterOpen;
	closeSound = SND.characterClose;
	// The open tab: 0 Character, 1 Skills, 2 Reputation (TOGGLECHARACTER0, 1 and 2).
	tab = 0;
	private pages: HTMLElement[] = [];
	private tabs: HTMLElement[] = [];
	private slots: { button: ItemButton; def: SlotDef; id?: number }[] = [];
	private boxes: StatBox[] = [];
	private dropdowns: { close(): void }[] = [];
	private resValues: HTMLDivElement[] = [];
	private levelText?: HTMLDivElement;
	private nameText?: HTMLDivElement;
	private portraitEl?: HTMLDivElement;
	private model?: ModelView;
	private skills?: SkillsPage;
	private filled = false;

	constructor(readonly link: RoomLink) {
		super();
	}

	protected async build(root: HTMLElement) {
		await texVars({
			'bz-dd-arrow': () => tex(arrowUp),
			'bz-dd-arrow-down': () => tex(arrowDown),
			'bz-dd-arrow-highlight': () => glow(TX.mouseHilight),
		});
		const { loaded } = this.link;
		// The portrait sits under the art, which has a round hole for it.
		this.portraitEl = div(root, 'bz-char-portrait', { x: 7, y: 6, w: 60, h: 60 });
		const circle = classCircle[loaded.class];
		if (circle) setTexture(this.portraitEl, TX.classCircles, { coords: circle, size: [256, 256] });

		const paper = div(root, 'bz-char-page');
		const skills = div(root, 'bz-char-page');
		const rep = div(root, 'bz-char-page');
		this.pages = [paper, skills, rep];
		this.buildPaperDoll(paper);
		art(skills, [TX.skillsTopLeft, TX.skillsTopRight, TX.skillFrameBotLeft, TX.skillFrameBotRight]);
		this.skills = new SkillsPage(skills, this.link, () => this.close());
		art(rep, [TX.generalTopLeft, TX.generalTopRight, TX.generalBottomLeft, TX.generalBottomRight]);
		text(rep, 'Faction', 'GameFontHighlightSmall', { x: 70, y: 58 });
		text(rep, 'Standing', 'GameFontHighlightSmall', { x: 215, y: 58 });
		text(rep, 'No factions yet', 'GameFontDisable', { x: 24, y: 200, w: 320 }, 'center').classList.add('bz-rep-empty');

		// The name over all tabs, the close button and the tabs under the frame.
		this.nameText = text(root, this.link.playerName(), 'GameFontHighlight', { x: 74, y: 17, w: 250 }, 'center');
		closeButton(root, { x: 322, y: 8 }, () => this.close());
		this.buildTabs(root);
		this.showPage();

		void this.link.whenReady().then(() => {
			this.fill();
			if (this.isOpen) this.shown();
		});
	}

	private buildTabs(root: HTMLElement) {
		// CharacterFrameTab1 is centered at 60 from the left and 62 up from the bottom. Each next
		// tab overlaps the one before by 16.
		const ctx = document.createElement('canvas').getContext('2d')!;
		ctx.font = "10px 'Friz Quadrata'";
		let x = 0;
		this.tabs = tabOrder.map(([n, label], i) => {
			const w = Math.ceil(ctx.measureText(label).width) + 40;
			if (i === 0) x = 60 - w / 2;
			const t = tab(root, label, { x: Math.round(x), y: 512 - 62 - 16, w }, () => {
				sound(SND.characterTab);
				this.tab = n;
				this.showPage();
			});
			x += w - 16;
			return t;
		});
	}

	private buildPaperDoll(page: HTMLElement) {
		art(page, [TX.characterL1, TX.characterR1, TX.characterL2, TX.characterR2]);
		this.levelText = text(page, '', 'GameFontHighlightSmall', { x: 74, y: 44, w: 250 }, 'center');

		// The model first, so the stats and resistances sit on it.
		this.model = new ModelView(page, { x: 65, y: 76, w: 233, h: 300 }, this.link);

		const ranged: SlotDef = relicClasses.has(this.link.loaded.class)
			? { name: 'Relic', item: 16, texture: TX.slotRelic }
			: { name: 'Ranged', item: 16, texture: TX.slotRanged };
		const bottom: SlotDef[] = [{ name: 'Main Hand', item: 14, texture: TX.slotMainHand }, { name: 'Off Hand', item: 15, texture: TX.slotSecondaryHand }, ranged];
		const place = (defs: SlotDef[], x0: number, y0: number, dx: number, dy: number) =>
			defs.forEach((def, i) => {
				const button = itemButton(page, { x: x0 + dx * i, y: y0 + dy * i }, def.texture, false);
				button.el.classList.add('bz-char-slot');
				const slot = { button, def, id: undefined as number | undefined };
				this.slots.push(slot);
				this.link.tooltip.attach(button.el, () => (slot.id ? { kind: 'item', id: slot.id } : { kind: 'lines', title: def.name }));
				// The doll is read only here, but a press must not reach the world behind.
				button.el.addEventListener('mousedown', e => e.stopPropagation());
			});
		place(leftSlots, 21, 74, 0, 41);
		place(rightSlots, 305, 74, 0, 41);
		place(bottom, 122, 386, 41, 0);

		resistances.forEach(([name, stat, top, bottom], i) => {
			const r = div(page, 'bz-res', { x: 297 - 32, y: 77 + 29 * i });
			texture(r, TX.resistanceIcons, { x: 0, y: 0, w: 32, h: 29 }, { coords: [0, 1, top, bottom], size: [32, 256] });
			const v = text(r, '0', 'GameFontHighlightSmall');
			this.resValues.push(v);
			this.link.tooltip.attach(r, () => {
				const s = this.link.sheet();
				return {
					kind: 'lines',
					title: s ? statTitle(s, `${name} Resistance`, stat) : `${name} Resistance`,
					lines: [`Increases the ability to resist ${name}-based attacks, spells and abilities.`],
				};
			});
		});

		// The two stat boxes under the model, each with its dropdown.
		const defaults = ['base', 'melee'];
		[67, 182].forEach((x, i) => {
			const key = `bz-char-stats-${i}`;
			const box: StatBox = { rows: [], key: stored(key, defaults[i]) };
			const el = div(page, 'bz-stat-box', { x, y: 291 });
			texture(el, TX.statBackground, { x: 0, y: 0, w: 115, h: 79 }, { coords: [0, 0.8984375, 0, 0.6171875], size: [128, 128] });
			for (let r = 0; r < 6; r++) {
				const row = div(el, 'bz-stat-row', { y: 3 + 13 * r });
				const label = text(row, '', 'GameFontNormalSmall');
				const value = text(row, '', 'GameFontHighlightSmall');
				value.classList.add('bz-stat-value');
				const entry = { el: row, label, value, tip: undefined as Tip | undefined };
				box.rows.push(entry);
				this.link.tooltip.attach(row, () => (entry.tip ? { kind: 'lines', title: entry.tip.title, lines: entry.tip.lines } : undefined));
			}
			this.boxes.push(box);
			const dd = statDropdown(
				page,
				{ x: x - 17, y: 266 },
				72,
				categories.map(c => [c.key, c.name]),
				box.key,
				v => {
					box.key = v;
					store(key, v);
					this.fillStats();
				},
			);
			this.dropdowns.push(dd);
		});
	}

	// fill puts the sheet's items, stats and resistances in, once the session has started.
	private fill() {
		const s = this.link.sheet();
		if (!s) return;
		this.filled = true;
		const { loaded } = this.link;
		this.levelText!.textContent = `Level ${s.level} ${enumName(loaded.race, 'Race')} ${enumName(loaded.class, 'Class')}`;
		this.nameText!.textContent = this.link.playerName();
		const worn = new Map((s.items ?? []).map(it => [it.slot, it]));
		for (const slot of this.slots) {
			const it = slot.def.item !== undefined ? worn.get(slot.def.item) : undefined;
			slot.id = it?.id;
			slot.button.set(it ? { kind: 'item', id: it.id, quality: it.quality } : undefined);
		}
		resistances.forEach(([, stat], i) => {
			const v = this.resValues[i];
			v.textContent = String(round(s.stats[stat] ?? 0));
			v.classList.toggle('GameFontGreen', colorOf(s, stat) === 'up');
		});
		this.fillStats();
		this.skills?.refresh();
	}

	private fillStats() {
		const s = this.link.sheet();
		if (!s) return;
		const facts: Facts = { cls: this.link.loaded.class, level: s.level };
		for (const box of this.boxes) {
			const cat = categories.find(c => c.key === box.key) ?? categories[0];
			const lines = cat.lines(s, facts);
			box.rows.forEach((row, i) => {
				const l = lines[i];
				row.el.style.display = l ? '' : 'none';
				row.label.textContent = l?.label ?? '';
				row.value.textContent = l?.value ?? '';
				row.value.classList.toggle('up', l?.color === 'up');
				row.value.classList.toggle('down', l?.color === 'down');
				row.tip = l?.tip;
			});
		}
	}

	private showPage() {
		this.pages.forEach((p, i) => (p.style.display = i === this.tab ? '' : 'none'));
		selectTab(
			this.tabs,
			tabOrder.findIndex(([n]) => n === this.tab),
		);
		for (const d of this.dropdowns) d.close();
		if (this.isOpen && this.tab === 0) this.model?.show();
		if (this.tab === 1) this.skills?.update();
	}

	protected shown() {
		if (!this.filled) this.fill();
		if (this.portraitEl) portrait(this.portraitEl, this.link);
		if (this.tab === 0) this.model?.show();
		if (this.tab === 1) this.skills?.update();
	}

	protected hidden() {
		for (const d of this.dropdowns) d.close();
	}

	// showTab opens a tab (the frame is open already or opens next).
	showTab(tab: number) {
		if (tab === this.tab) return;
		this.tab = tab;
		if (!this.pages.length) return;
		if (this.isOpen) sound(SND.characterTab);
		this.showPage();
	}
}
