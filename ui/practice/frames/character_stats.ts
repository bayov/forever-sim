// The stat lines of Forever's character frame (the Character Stats pane): the categories of
// Camelot/PaperDollFrameConstants.lua (PAPERDOLL_STATCATEGORIES) and the Resistances that
// CharacterFrame.lua adds, each line formatted like Camelot/PaperDollFrame.lua and
// PaperDollFrameStats.lua, with the client's GlobalStrings.
//
// The numbers come from the sim's character sheet (link.sheet()). We leave out what the sheet
// doesn't have: Movement Speed, and the tooltip lines for crit and dodge from Agility, spell crit
// from Intellect and regeneration from Spirit.

import type { Sheet } from '../protocol';

export type StatTip = { title: string; lines: string[] };
export type StatRow = { label: string; value: string; color?: 'up' | 'down'; icon?: string; tip: StatTip };
export type StatCategory = { name: string; rows: StatRow[] };

// What the stat lines need besides the sheet.
export type Facts = { cls: string; level: number };

const round = (v: number) => Math.round(v + 1e-6);
// big is BreakUpLargeNumbers: a comma every three digits.
const big = (v: number) => String(round(v)).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
const stat = (s: Sheet, i: number) => s.stats[i] ?? 0;

// lines splits a GlobalStrings text into tooltip lines. A blank line stays as a gap.
const lines = (...texts: string[]) => texts.flatMap(t => t.split('\n'));

// colorOf is green for a stat above its base (gear and buffs), red below it.
function colorOf(s: Sheet, i: number): 'up' | 'down' | undefined {
	const d = stat(s, i) - (s.base[i] ?? 0);
	return d > 0.5 ? 'up' : d < -0.5 ? 'down' : undefined;
}

// formatStat is PaperDollFormatStat's title: "Attack Power 1,234 (1000+234)".
function formatStat(name: string, total: number, base: number): string {
	const t = round(total);
	const b = round(base);
	if (t === b) return `${name} ${big(t)}`;
	return `${name} ${big(t)} (${b}${t > b ? '+' : ''}${t - b})`;
}

const statNames = ['Strength', 'Agility', 'Stamina', 'Intellect', 'Spirit'];
const apPerStrength: Record<string, number> = { ClassWarrior: 2, ClassPaladin: 2, ClassShaman: 2, ClassDruid: 2 };
const manaUsers = new Set(['ClassShaman', 'ClassPaladin', 'ClassHunter', 'ClassPriest', 'ClassMage', 'ClassWarlock', 'ClassDruid']);
// Only warriors, paladins, hunters and rogues learn Parry. The sim gives every unit the 5% base.
const parryClasses = new Set(['ClassWarrior', 'ClassPaladin', 'ClassHunter', 'ClassRogue']);
// The classes whose hit and crit tooltips list ranged too.
const rangedClasses = new Set(['ClassWarrior', 'ClassHunter', 'ClassRogue']);

// firstTwenty is the game's rule for Health and Mana: the first 20 points give 1 each.
const firstTwenty = (v: number, per: number) => Math.min(20, v) + Math.max(0, v - 20) * per;

function primary(s: Sheet, f: Facts, i: number): StatRow {
	const v = round(stat(s, i));
	const name = statNames[i];
	let tip: string[] = [];
	if (i === 0) {
		tip = [`Increases Attack Power by ${v * (apPerStrength[f.cls] ?? 1)}`];
		if (f.cls === 'ClassShaman') tip.push(`Increases Block Value by ${Math.floor(v / 20)}`);
	} else if (i === 1) tip = [`Increases Armor by ${v * 2}`];
	else if (i === 2) tip = [`Increases Health by ${firstTwenty(v, 10)}`];
	else if (i === 3 && manaUsers.has(f.cls)) tip = [`Increases Mana by ${firstTwenty(v, 15)}`, 'Increases Weapon Skill improvement rate'];
	else if (i === 4) tip = ['Mana Regeneration from Spirit is halted for 5 sec after spending Mana', '', 'Health Regeneration is increased by 33% while sitting or reclining'];
	return { label: name, value: big(v), color: colorOf(s, i), tip: { title: formatStat(name, stat(s, i), s.base[i] ?? 0), lines: tip } };
}

type Weapon = NonNullable<Sheet['weapons']>[number];

// The weapon skills by the sim's weapon type, with their PseudoStat bonus index. Axes, maces and
// swords are two-handed when nothing is in the off hand.
const weaponSkills: Record<string, [one: string, two: string, onePseudo: number, twoPseudo: number]> = {
	WeaponTypeAxe: ['Axes', 'Two-Handed Axes', 10, 13],
	WeaponTypeMace: ['Maces', 'Two-Handed Maces', 9, 12],
	WeaponTypeSword: ['Swords', 'Two-Handed Swords', 8, 11],
	WeaponTypeDagger: ['Daggers', 'Daggers', 7, 7],
	WeaponTypeFist: ['Fist Weapons', 'Fist Weapons', 6, 6],
	WeaponTypePolearm: ['Polearms', 'Polearms', 14, 14],
	WeaponTypeStaff: ['Staves', 'Staves', 15, 15],
	RangedWeaponTypeBow: ['Bows', 'Bows', 16, 16],
	RangedWeaponTypeCrossbow: ['Crossbows', 'Crossbows', 17, 17],
	RangedWeaponTypeGun: ['Guns', 'Guns', 18, 18],
	RangedWeaponTypeThrown: ['Thrown', 'Thrown', 19, 19],
};

function skillLine(s: Sheet, f: Facts, w: Weapon): string | undefined {
	const k = w.type && weaponSkills[w.type];
	if (!k) return undefined;
	const offHandEmpty = !(s.items ?? []).some(it => it.slot === 15);
	const two = w.slot === 'mainHand' && offHandEmpty;
	const mod = round(s.pseudo?.[two ? k[3] : k[2]] ?? 0);
	const cap = f.level * 5;
	return `${two ? k[1] : k[0]}: ${cap}${mod ? ` (${mod})` : ''}/${cap}`;
}

function damage(s: Sheet, f: Facts, w: Weapon, label: string): StatRow {
	const mult = s.damageMultiplier || 1;
	const lo = Math.max(1, Math.floor(w.min * mult));
	const hi = Math.max(1, Math.ceil(w.max * mult));
	const sep = lo < 100 && hi < 100 ? ' - ' : '-';
	const bonus = Math.abs(mult - 1) > 1e-4;
	let dmg = `${Math.max(1, Math.floor(w.min))} - ${Math.max(1, Math.ceil(w.max))}`;
	if (bonus) dmg += ` x${Math.round(mult * 100)}%`;
	const tip = [`Attack Speed (seconds): ${w.speed.toFixed(2)}`, `Damage: ${dmg}`, `DPS: ${((lo + hi) / (2 * w.speed)).toFixed(1)}`];
	const skill = skillLine(s, f, w);
	if (skill) tip.push(skill);
	return { label, value: `${lo}${sep}${hi}`, color: mult > 1.0001 ? 'up' : mult < 0.9999 ? 'down' : undefined, tip: { title: label, lines: tip } };
}

function attackPower(s: Sheet, i: number, name: string, ranged: boolean): StatRow | undefined {
	const v = stat(s, i);
	if (round(v) === 0) return undefined;
	return {
		label: name,
		value: big(v),
		color: colorOf(s, i),
		tip: {
			title: formatStat(name, v, s.base[i] ?? 0),
			lines: lines(
				ranged
					? `Increases Ranged Weapon damage per second by ${(Math.max(0, v) / 14).toFixed(1)}\n\nIncreased by Agility\n\nEvery 14 Ranged Attack Power adds 1 damage per second`
					: `Increases Melee Weapon damage per second by ${(Math.max(0, v) / 14).toFixed(1)}\n\nIncreased by Strength\n\nEvery 14 Attack Power adds 1 damage per second`,
			),
		},
	};
}

const pct1 = (v: number) => `${v.toFixed(1)}%`;
const pct2 = (v: number) => v.toFixed(2);

// The spell schools with their power stat, in the game's order (holy to arcane).
const schools: [string, number][] = [
	['Holy', 9],
	['Fire', 7],
	['Nature', 10],
	['Frost', 8],
	['Shadow', 11],
	['Arcane', 6],
];

function modifiers(s: Sheet, f: Facts): StatRow[] {
	const rows: StatRow[] = [];
	const melee = stat(s, 18);
	const spell = stat(s, 13);
	const hit = Math.max(melee, spell);
	if (hit > 0)
		rows.push({
			label: 'Hit Chance',
			value: pct1(hit),
			tip: {
				title: `Hit Chance ${pct1(hit)}`,
				lines: [
					...(melee > 0 ? [`Increases Melee hit chance by ${pct2(melee)}%`] : []),
					...(melee > 0 && rangedClasses.has(f.cls) ? [`Increases Ranged hit chance by ${pct2(melee)}%`] : []),
					...(spell > 0 ? [`Increases Spell hit chance by ${pct2(spell)}%`] : []),
					...lines(`\nTo never miss Raid Bosses:\n  8.00% Melee\n17.00% Spells\n\nTo never miss Lvl ${f.level} Targets/Players:\n  5.00% Melee\n  4.00% Spells`),
				],
			},
		});
	const mcrit = stat(s, 19);
	const scrit = stat(s, 14);
	const crit = Math.max(mcrit, scrit);
	rows.push({
		label: 'Critical Strike',
		value: pct1(crit),
		tip: {
			title: `Critical Strike ${pct1(crit)}`,
			lines: [
				`Increases Melee critical chance by ${pct2(mcrit)}%`,
				...(rangedClasses.has(f.cls) ? [`Increases Ranged critical chance by ${pct2(mcrit)}%`] : []),
				`Increases Spell critical chance by ${pct2(scrit)}%`,
				'',
				...lines(
					'Increased by Agility for Attacks\nIncreased by Intellect for Spells\n\nMelee critical strikes deal 100% increased damage\n\nSpell and Healing critical strikes are 50% more effective\n\nMost periodic effects can critically strike',
				),
			],
		},
	});
	const mhaste = ((s.pseudo?.[27] ?? 1) - 1) * 100;
	const shaste = ((s.pseudo?.[34] ?? 1) - 1) * 100;
	const haste = Math.max(mhaste, shaste);
	if (round(haste * 10) > 0)
		rows.push({
			label: 'Haste',
			value: pct1(haste),
			tip: {
				title: `Haste ${pct1(haste)}`,
				lines: [...(mhaste > 0 ? [`Increases Melee attack speed by ${pct2(mhaste)}%`] : []), ...(shaste > 0 ? [`Increases Spell casting speed by ${pct2(shaste)}%`] : [])],
			},
		});
	const exp = stat(s, 22);
	if (exp > 0)
		rows.push({
			label: 'Expertise',
			value: pct1(exp),
			tip: {
				title: `Expertise ${pct1(exp)}`,
				lines: lines(
					`Increases the chance Melee attacks are not Dodged or Parried by ${pct1(exp)}\n\nRaid Boss chance to:\n  6.50% Dodge\n14.00% Parry\n\nEqual-Level Creature chance to:\n  5.00% Dodge\n  5.00% Parry`,
				),
			},
		});
	const arp = round(stat(s, 21));
	if (arp)
		rows.push({
			label: 'Armor Piercing',
			value: String(arp),
			tip: { title: `Armor Piercing ${arp}`, lines: [`Your Attacks ignore ${arp} of your enemies' Armor when attacking. Reducing an enemy's armor below 0 will increase your damage against them.`] },
		});
	const generic = stat(s, 5) + stat(s, 42);
	const bySchool = schools.map(([n, i]) => [n, round(generic + stat(s, i))] as const);
	const minSp = Math.min(...bySchool.map(([, v]) => v));
	const maxSp = Math.max(...bySchool.map(([, v]) => v));
	if (minSp)
		rows.push({
			label: 'Spell Damage',
			value: big(minSp),
			tip: {
				title: `Spell Damage ${big(minSp)}`,
				lines: [
					...bySchool.filter(([, v]) => v !== minSp).map(([n, v]) => `${n} ${big(v)}`),
					...lines(
						`Increases the damage of your Spells by up to ${maxSp}\n\nIncreased by Spell Damage and Spell Power\n\nLonger cast times typically benefit more from Spell Damage\n\nSpell ranks far below your current level benefit less from Spell Damage and trigger class abilities less often`,
					),
				],
			},
		});
	const heal = round(stat(s, 5) + stat(s, 41));
	if (heal)
		rows.push({
			label: 'Spell Healing',
			value: big(heal),
			tip: {
				title: `Spell Healing ${big(heal)}`,
				lines: lines(
					`Increases the healing of your Spells by up to ${heal}\n\nIncreased by Spell Healing and Spell Power\n\nLonger cast times typically benefit more from Spell Healing\n\nSpell ranks far below your current level benefit less from Spell Healing and trigger class abilities less often`,
				),
			},
		});
	const pen = round(stat(s, 16));
	if (pen) {
		const per = (0.75 / (f.level * 5)) * pen;
		rows.push({
			label: 'Spell Piercing',
			value: String(pen),
			tip: {
				title: `Spell Piercing ${pen}`,
				lines: lines(
					`Increases the Resistances your Spells ignore by ${pen}\n\nResistances reduce spell damage taken and increase the chance to resist by an average of ${per.toFixed(1)}% per point\n\nSpell Piercing may bring a creature below 0 Resistances, becoming Spell Vulnerability, increasing their spell damage taken by ${(per / 2).toFixed(1)}% per point`,
				),
			},
		});
	}
	return rows;
}

// armorReduction is the Classic armor formula against an enemy of your level.
function armorReduction(armor: number, level: number): number {
	const l = level < 60 ? level : level + 4.5 * (level - 59);
	return Math.min(75, (armor / (armor + 400 + 85 * l)) * 100);
}

function defense(s: Sheet, f: Facts): StatRow[] {
	const cap = f.level * 5;
	const bonus = round(stat(s, 28));
	const def = cap + bonus;
	const per = (Math.max(0, bonus) * 0.04).toFixed(2);
	const rows: StatRow[] = [
		{
			label: 'Defense',
			value: `${def} / ${cap}`,
			color: bonus > 0 ? 'up' : bonus < 0 ? 'down' : undefined,
			tip: {
				title: formatStat('Defense', def, cap),
				lines: lines(
					`Increases Dodge, Block, and Parry chance by ${per}%, and decreases chance to be Hit or Critically Hit by ${per}%\n\nCritical Strikes from creatures deal 200% damage. Creatures 3+ levels above you can land Crushing Blows for 150% damage\n\nAt 440 Defense, cannot be Critically Hit by Raid Bosses (-5.60% Critical Strike chance)`,
				),
			},
		},
	];
	const dodge = stat(s, 31);
	if (dodge > 0)
		rows.push({
			label: 'Dodge',
			value: pct1(dodge),
			tip: {
				title: `Dodge Chance ${pct1(dodge)}`,
				lines: lines(
					`Increases chance to Dodge by ${pct2(dodge)}%\n\nIncreased by Agility and Defense\n\nDodging nullifies the attack\n\nFor Players, only Melee attacks from the front may be Dodged\n\nFor Creatures, Melee attacks may be Dodged from any direction`,
				),
			},
		});
	// Block shows only with a shield: an off hand item that isn't a weapon.
	const shield = (s.items ?? []).some(it => it.slot === 15) && !(s.weapons ?? []).some(w => w.slot === 'offHand');
	const block = stat(s, 29);
	if (shield && block > 0)
		rows.push({
			label: 'Block',
			value: `${Math.floor(block + 0.5)}%`,
			tip: {
				title: `Block Chance ${pct1(block)}`,
				lines: lines(
					`Increases chance to Block by ${pct2(block)}%\n\nIncreased by Defense\n\nBlock Value ${round(stat(s, 30))}\nIncreased by Strength\n\nBlocking reduces the attack's damage by your Block Value\n\nOnly Melee and Ranged attacks from the front may be Blocked`,
				),
			},
		});
	const parry = parryClasses.has(f.cls) ? stat(s, 32) : 0;
	if (parry > 0)
		rows.push({
			label: 'Parry',
			value: pct1(parry),
			tip: {
				title: `Parry Chance ${pct1(parry)}`,
				lines: lines(
					`Increases chance to Parry by ${pct2(parry)}%\n\nIncreased by Defense\n\nParrying nullifies the attack and reduces the time until the defender's next Melee attack by 40%\n\nOnly Melee attacks from the front may be Parried`,
				),
			},
		});
	const armor = stat(s, 26);
	rows.push({
		label: 'Armor',
		value: big(armor),
		color: colorOf(s, 26) === 'up' ? 'up' : undefined,
		tip: {
			title: `Armor ${big(armor)}`,
			lines: [`Reduces Physical Damage taken by ${armorReduction(armor, f.level).toFixed(2)}%`, '', 'Increased by Agility', '', 'Damage reduction from Armor cannot exceed 75.00%'],
		},
	});
	return rows;
}

// The resistances in the frame's order, with their stat and icon.
const resistances: [name: string, stat: number, atlas: string][] = [
	['Arcane', 35, 'UI-Character-Info-Resistance-Arcane'],
	['Fire', 36, 'UI-Character-Info-Resistance-Fire'],
	['Frost', 37, 'UI-Character-Info-Resistance-Frost'],
	['Nature', 38, 'UI-Character-Info-Resistance-Nature'],
	['Shadow', 39, 'UI-Character-Info-Resistance-Shadow'],
];

export const statAtlases = resistances.map(r => r[2]);

// statCategories builds the pane's categories from the sheet.
export function statCategories(s: Sheet, f: Facts): StatCategory[] {
	const general: StatRow[] = [
		{
			label: 'Health',
			value: big(stat(s, 34)),
			tip: {
				title: `Health ${big(stat(s, 34))}`,
				lines: lines(
					'Used for Taking Damage\nResults in death when reduced to 0\n\nMaximum increased by Stamina\nRegeneration increased by Spirit\n\nReplenishes automatically while out of combat\n\nEvery 1 Stamina adds 10 Health',
				),
			},
		},
	];
	if (manaUsers.has(f.cls))
		general.push({
			label: 'Mana',
			value: big(stat(s, 23)),
			tip: {
				title: `Mana ${big(stat(s, 23))}`,
				lines: lines(
					'Resource used to cast Spells and Abilities\n\nMaximum increased by Intellect\nRegeneration increased by Spirit\n\nReplenishes automatically after not spending Mana for 5 sec\n\nEvery 1 Intellect adds 15 Mana',
				),
			},
		});
	else if (f.cls === 'ClassRogue') general.push({ label: 'Energy', value: big(stat(s, 24) || 100), tip: { title: `Energy ${big(stat(s, 24) || 100)}`, lines: [] } });
	else if (f.cls === 'ClassWarrior') general.push({ label: 'Rage', value: big(stat(s, 25) || 100), tip: { title: `Rage ${big(stat(s, 25) || 100)}`, lines: [] } });

	const ws = s.weapons ?? [];
	const weapons: StatRow[] = [];
	const mh = ws.find(w => w.slot === 'mainHand');
	const oh = ws.find(w => w.slot === 'offHand');
	const rw = ws.find(w => w.slot === 'ranged');
	if (mh) weapons.push(damage(s, f, mh, 'Main Hand'));
	if (oh) weapons.push(damage(s, f, oh, 'Off Hand'));
	if (rw) weapons.push(damage(s, f, rw, 'Ranged'));
	const ap = attackPower(s, 17, 'Attack Power', false);
	if (ap) weapons.push(ap);
	const rap = attackPower(s, 27, 'Ranged Attack Power', true);
	if (rap) weapons.push(rap);

	const res: StatRow[] = resistances.map(([name, i, atlas]) => {
		const v = round(stat(s, i));
		const avg = Math.floor(((0.75 * v) / (f.level * 5)) * 100);
		return {
			label: name,
			value: big(v),
			color: colorOf(s, i) === 'up' ? 'up' : undefined,
			icon: atlas,
			tip: {
				title: `${name} ${v}`,
				lines: [`Improves resistance to ${name.toLowerCase()}-based attacks, spells, and abilities`, '', `Average resistance versus a level ${f.level} enemy: ${avg}%`],
			},
		};
	});

	return [
		{ name: 'General', rows: general },
		{ name: 'Primary Attributes', rows: [0, 1, 2, 3, 4].map(i => primary(s, f, i)) },
		{ name: 'Weapons', rows: weapons },
		{ name: 'Modifiers', rows: modifiers(s, f) },
		{ name: 'Defense', rows: defense(s, f) },
		{ name: 'Resistances', rows: res },
	].filter(c => c.rows.length);
}
