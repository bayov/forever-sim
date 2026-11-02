// Keys, bindings, action bar slots and macros, as WoW Forever reads them (docs/practice_room.md,
// "Keybinds and action bars").

import { ToolkitSpell } from './protocol';

// keyName turns a key event into WoW's key name, like "SHIFT-R" or "NUMPAD0".
const codeNames: Record<string, string> = {
	Space: 'SPACE',
	Tab: 'TAB',
	Escape: 'ESCAPE',
	Enter: 'ENTER',
	Backspace: 'BACKSPACE',
	CapsLock: 'CAPSLOCK',
	Minus: '-',
	Equal: '=',
	Backquote: '`',
	BracketLeft: '[',
	BracketRight: ']',
	Semicolon: ';',
	Quote: "'",
	Comma: ',',
	Period: '.',
	Slash: '/',
	Backslash: '\\',
	ArrowUp: 'UP',
	ArrowDown: 'DOWN',
	ArrowLeft: 'LEFT',
	ArrowRight: 'RIGHT',
	Insert: 'INSERT',
	Delete: 'DELETE',
	Home: 'HOME',
	End: 'END',
	PageUp: 'PAGEUP',
	PageDown: 'PAGEDOWN',
	NumLock: 'NUMLOCK',
	NumpadDivide: 'NUMPADDIVIDE',
	NumpadMultiply: 'NUMPADMULTIPLY',
	NumpadSubtract: 'NUMPADMINUS',
	NumpadAdd: 'NUMPADPLUS',
	NumpadDecimal: 'NUMPADDECIMAL',
	NumpadEnter: 'ENTER',
};

// lockableCodes are the key codes the room knows, for Keyboard Lock.
export function lockableCodes(): string[] {
	const codes = Object.keys(codeNames);
	for (let c = 65; c <= 90; c++) codes.push('Key' + String.fromCharCode(c));
	for (let d = 0; d <= 9; d++) codes.push('Digit' + d, 'Numpad' + d);
	for (let f = 1; f <= 24; f++) codes.push('F' + f);
	return codes;
}

export function baseKeyName(code: string): string | undefined {
	if (code.startsWith('Key')) return code.slice(3);
	if (code.startsWith('Digit')) return code.slice(5);
	if (/^F\d+$/.test(code)) return code;
	if (code.startsWith('Numpad') && /\d$/.test(code)) return 'NUMPAD' + code.slice(6);
	return codeNames[code];
}

export type Modifiers = { alt: boolean; ctrl: boolean; shift: boolean };

export function withModifiers(base: string, mods: Modifiers): string {
	let name = base;
	if (mods.shift) name = 'SHIFT-' + name;
	if (mods.ctrl) name = 'CTRL-' + name;
	if (mods.alt) name = 'ALT-' + name;
	return name;
}

export function mouseButtonName(button: number): string | undefined {
	switch (button) {
		case 1:
			return 'BUTTON3';
		case 3:
			return 'BUTTON4';
		case 4:
			return 'BUTTON5';
	}
	return undefined;
}

// commandFor finds a key's command. WoW falls back to the key's unmodified binding when the
// modified key is unbound, and the user's [mod:alt] macros depend on that.
export function commandFor(bindings: Record<string, string>, base: string, mods: Modifiers): string | undefined {
	return bindings[withModifiers(base, mods)] ?? (mods.alt || mods.ctrl || mods.shift ? bindings[base] : undefined);
}

// slotForCommand maps a bar button command to its action slot (Blizzard's table, and
// EllesmereUI's extra bars).
export function slotForCommand(command: string): number | undefined {
	let m = /^ACTIONBUTTON(\d+)$/.exec(command);
	if (m) return +m[1];
	m = /^MULTIACTIONBAR(\d)BUTTON(\d+)$/.exec(command);
	if (m) {
		const bar = +m[1];
		const button = +m[2];
		const base: Record<number, number> = { 1: 60, 2: 48, 3: 24, 4: 36, 5: 144, 6: 156, 7: 168 };
		if (base[bar] !== undefined) return base[bar] + button;
	}
	m = /^EUI_BAR(\d+)_BUTTON(\d+)$/.exec(command);
	if (m) {
		const bar = +m[1];
		const button = +m[2];
		if (bar === 9) return 12 + button;
		if (bar === 10) return 108 + button;
		if (bar >= 11 && bar <= 13) return 72 + (bar - 11) * 12 + button;
	}
	return undefined;
}

// commandForSlot is the first bar command that presses a slot, to show its key.
export function keysForSlot(bindings: Record<string, string>, slot: number): string[] {
	const keys: string[] = [];
	for (const [key, command] of Object.entries(bindings)) {
		if (slotForCommand(command) === slot) keys.push(key);
	}
	return keys;
}

// shortKey writes a key like the user's bars do: SQ for SHIFT-Q, CF1, M3 for the middle mouse
// button, N- for the numpad's minus.
const numpadShort: Record<string, string> = { NUMPADMINUS: 'N-', NUMPADPLUS: 'N+', NUMPADDECIMAL: 'N.', NUMPADMULTIPLY: 'N*', NUMPADDIVIDE: 'N/' };

export function shortKey(key: string): string {
	return key
		.replace(/NUMPAD(MINUS|PLUS|DECIMAL|MULTIPLY|DIVIDE)/, m => numpadShort[m])
		.replace('SHIFT-', 'S')
		.replace('CTRL-', 'C')
		.replace('ALT-', 'A')
		.replace('MOUSEWHEELUP', 'MwU')
		.replace('MOUSEWHEELDOWN', 'MwD')
		.replace('BUTTON', 'M')
		.replace('NUMPAD', 'N');
}

// What sits in a slot.
export type SlotContent =
	| { kind: 'spell'; id: number; name?: string }
	| { kind: 'item'; id: number; name?: string }
	| { kind: 'macro'; name: string; body: string; icon?: number };

// Spell names and ranks, for macros and for spells the client knows by another rank.
export type SpellNames = Record<string, { name: string; rank?: string; icon?: number }>;

// The toolkit, looked up by name and rank.
export class Toolkit {
	readonly bySpell = new Map<number, ToolkitSpell>();
	readonly byItem = new Map<number, ToolkitSpell>();
	private byName = new Map<string, { id: number; rank: number }[]>();

	constructor(
		readonly spells: ToolkitSpell[],
		readonly names: SpellNames,
	) {
		for (const s of spells) {
			if (s.spellId) {
				this.bySpell.set(s.spellId, s);
				const n = names[s.spellId];
				if (n) {
					const key = n.name.toLowerCase();
					const rank = parseInt((n.rank ?? '').replace(/\D/g, '')) || 0;
					const list = this.byName.get(key) ?? [];
					list.push({ id: s.spellId, rank });
					this.byName.set(key, list);
				}
			}
			if (s.itemId) this.byItem.set(s.itemId, s);
		}
		for (const list of this.byName.values()) list.sort((a, b) => b.rank - a.rank);
	}

	// spellByName finds "Frost Shock" (the highest rank) or "Frost Shock(Rank 1)".
	spellByName(text: string): number | undefined {
		const m = /^(.*?)\s*\(\s*rank\s*(\d+)\s*\)\s*$/i.exec(text);
		const name = (m ? m[1] : text).trim().toLowerCase();
		const list = this.byName.get(name);
		if (!list?.length) return undefined;
		if (m) return list.find(e => e.rank === +m[2])?.id;
		return list[0].id;
	}

	nameOf(spellId: number): string {
		return this.names[spellId]?.name ?? `Spell ${spellId}`;
	}

	// spellForSlot is the toolkit spell for a spell slot: its own id, or the same spell by name
	// at the rank the sim knows.
	resolveSpell(id: number): number | undefined {
		if (this.bySpell.has(id)) return id;
		const n = this.names[id];
		if (!n) return undefined;
		return this.spellByName(n.rank ? `${n.name}(${n.rank})` : n.name) ?? this.spellByName(n.name);
	}
}

// The state a macro's conditions read, at the press.
export type MacroContext = {
	mods: Modifiers;
	hasTarget: boolean;
	targetHarm: boolean;
	targetDead: boolean;
	mouseover: 'enemy' | 'none';
};

// An action the page sends to the session.
export type MacroAction =
	| { kind: 'cast'; spellId: number; target: 'target' | 'player' | 'mouseover' }
	| { kind: 'useItem'; itemId: number }
	| { kind: 'startAttack' }
	| { kind: 'stopAttack' }
	| { kind: 'clearTarget' };

const supportedCommands = new Set(['/cast', '/use', '/startattack', '/stopattack', '/cleartarget', '/stopcasting']);

export type ParsedMacro = { lines: { command: string; args: string }[]; showtooltip?: string; unsupported: string[] };

export function parseMacro(body: string): ParsedMacro {
	const out: ParsedMacro = { lines: [], unsupported: [] };
	for (const raw of body.split('\n')) {
		const line = raw.trim();
		if (!line) continue;
		if (line.startsWith('#showtooltip')) {
			out.showtooltip = line.slice('#showtooltip'.length).trim();
			continue;
		}
		if (line.startsWith('#')) continue;
		const [command, ...rest] = line.split(/\s+/);
		const cmd = command.toLowerCase();
		if (!supportedCommands.has(cmd)) {
			out.unsupported.push(cmd);
			continue;
		}
		out.lines.push({ command: cmd, args: rest.join(' ') });
	}
	return out;
}

// evalConditions checks one bracket group like "@mouseover,harm,nodead". It returns the unit
// the group picks, or undefined when it fails.
function evalGroup(group: string, ctx: MacroContext): 'target' | 'player' | 'mouseover' | undefined {
	let unit: 'target' | 'player' | 'mouseover' = 'target';
	for (const rawCond of group.split(',')) {
		const cond = rawCond.trim().toLowerCase();
		if (!cond) continue;
		if (cond.startsWith('@') || cond.startsWith('target=')) {
			const u = cond.replace(/^@|^target=/, '');
			if (u === 'mouseover') unit = 'mouseover';
			else if (u === 'player') unit = 'player';
			else unit = 'target';
			continue;
		}
		const exists = unit === 'player' || (unit === 'mouseover' ? ctx.mouseover !== 'none' : ctx.hasTarget);
		const harm = unit === 'mouseover' ? ctx.mouseover === 'enemy' : unit === 'target' && ctx.targetHarm;
		const dead = unit === 'target' && ctx.targetDead;
		let ok = true;
		if (cond === 'harm') ok = exists && harm;
		else if (cond === 'help') ok = exists && !harm;
		else if (cond === 'dead') ok = exists && dead;
		else if (cond === 'nodead') ok = exists && !dead;
		else if (cond === 'exists') ok = exists;
		else if (cond === 'noexists') ok = !exists;
		else if (cond.startsWith('mod') || cond.startsWith('nomod')) {
			const negate = cond.startsWith('no');
			const which = cond.split(':')[1];
			let held = ctx.mods.alt || ctx.mods.ctrl || ctx.mods.shift;
			if (which) {
				held = which.split('/').some(w => (w === 'alt' ? ctx.mods.alt : w === 'ctrl' ? ctx.mods.ctrl : w === 'shift' ? ctx.mods.shift : false));
			}
			ok = negate ? !held : held;
		}
		if (!ok) return undefined;
	}
	return unit;
}

// splitConditions reads "[a][b] Spell" into the groups and the argument. No groups means one
// empty group, which always passes.
function splitConditions(args: string): { groups: string[]; rest: string }[] {
	// A /cast line can hold several "conditions spell" clauses split by ';'.
	return args.split(';').map(clause => {
		const groups: string[] = [];
		let rest = clause.trim();
		while (rest.startsWith('[')) {
			const end = rest.indexOf(']');
			if (end < 0) break;
			groups.push(rest.slice(1, end));
			rest = rest.slice(end + 1).trim();
		}
		if (!groups.length) groups.push('');
		return { groups, rest };
	});
}

function firstMatch(args: string, ctx: MacroContext): { unit: 'target' | 'player' | 'mouseover'; arg: string } | undefined {
	for (const clause of splitConditions(args)) {
		for (const group of clause.groups) {
			const unit = evalGroup(group, ctx);
			if (unit) return { unit, arg: clause.rest };
		}
	}
	return undefined;
}

// runMacro turns a macro into the actions it does at this press. The first /cast or /use that
// resolves fires, as only one spell can go off per press.
export function runMacro(macro: ParsedMacro, ctx: MacroContext, toolkit: Toolkit, itemsByName: Map<string, number>): MacroAction[] {
	const actions: MacroAction[] = [];
	let cast = false;
	for (const line of macro.lines) {
		switch (line.command) {
			case '/startattack':
				if (firstMatch(line.args, ctx)) actions.push({ kind: 'startAttack' });
				break;
			case '/stopattack':
				if (firstMatch(line.args, ctx)) actions.push({ kind: 'stopAttack' });
				break;
			case '/cleartarget':
				if (firstMatch(line.args, ctx)) actions.push({ kind: 'clearTarget' });
				break;
			case '/cast': {
				if (cast) break;
				const m = firstMatch(line.args, ctx);
				if (!m) break;
				const id = toolkit.spellByName(m.arg);
				if (id !== undefined) {
					actions.push({ kind: 'cast', spellId: id, target: m.unit });
					cast = true;
				}
				break;
			}
			case '/use': {
				if (cast) break;
				const m = firstMatch(line.args, ctx);
				if (!m) break;
				const itemId = itemsByName.get(m.arg.toLowerCase()) ?? (/^\d+$/.test(m.arg) && +m.arg > 30 ? +m.arg : undefined);
				if (itemId !== undefined && toolkit.byItem.has(itemId)) {
					actions.push({ kind: 'useItem', itemId });
					cast = true;
				}
				break;
			}
		}
	}
	return actions;
}

// The spell a macro shows on its button: #showtooltip's spell, or its first /cast.
export function macroIconSpell(macro: ParsedMacro, toolkit: Toolkit, ctx: MacroContext): number | undefined {
	if (macro.showtooltip) {
		const id = toolkit.spellByName(macro.showtooltip);
		if (id) return id;
	}
	for (const line of macro.lines) {
		if (line.command !== '/cast') continue;
		const m = firstMatch(line.args, ctx) ?? splitConditions(line.args).map(c => ({ unit: 'target' as const, arg: c.rest }))[0];
		if (m) {
			const id = toolkit.spellByName(m.arg);
			if (id) return id;
		}
	}
	return undefined;
}

// The bar rows the room draws, by slot ranges, in the order Blizzard's bars stack.
export const barRows: { name: string; first: number }[] = [
	{ name: 'Bar 1', first: 1 },
	{ name: 'Bar 2', first: 61 },
	{ name: 'Bar 3', first: 49 },
	{ name: 'Bar 4', first: 25 },
	{ name: 'Bar 5', first: 37 },
	{ name: 'Bar 6', first: 145 },
	{ name: 'Bar 7', first: 157 },
	{ name: 'Bar 8', first: 169 },
	{ name: 'Bar 9', first: 13 },
	{ name: 'Bar 10', first: 109 },
	{ name: 'Bar 11', first: 73 },
	{ name: 'Bar 12', first: 85 },
	{ name: 'Bar 13', first: 97 },
];

// commandForSlot is the bar button command that presses a slot, for binding a key to it.
export function commandForSlot(slot: number): string | undefined {
	if (slot >= 1 && slot <= 12) return `ACTIONBUTTON${slot}`;
	const bars: [number, string][] = [
		[60, 'MULTIACTIONBAR1BUTTON'],
		[48, 'MULTIACTIONBAR2BUTTON'],
		[24, 'MULTIACTIONBAR3BUTTON'],
		[36, 'MULTIACTIONBAR4BUTTON'],
		[144, 'MULTIACTIONBAR5BUTTON'],
		[156, 'MULTIACTIONBAR6BUTTON'],
		[168, 'MULTIACTIONBAR7BUTTON'],
		[12, 'EUI_BAR9_BUTTON'],
		[108, 'EUI_BAR10_BUTTON'],
		[72, 'EUI_BAR11_BUTTON'],
		[84, 'EUI_BAR12_BUTTON'],
		[96, 'EUI_BAR13_BUTTON'],
	];
	for (const [base, name] of bars) {
		if (slot > base && slot <= base + 12) return name + (slot - base);
	}
	return undefined;
}

// Racials by the keybind notes' "Racial 1" and "Racial 2" (the Bayov class pack's lists).
const racials1 = ['Blood Fury', 'Berserking', 'Will of the Forsaken', 'Stoneform', 'Escape Artist', 'Shadowmeld', 'War Stomp', 'Perception', 'Cannibalize'];
const racials2 = ['Shatter Curse'];

// matchNoteLabel finds what a keybind notes label like "Earth Shock (r1)", "Mana Potion" or
// "Dark/Demonic Rune" casts: a toolkit spell by name and rank, or a toolkit item whose name ends
// with the label ("Mana Potion" is the Major Mana Potion we carry).
export function matchNoteLabel(label: string, toolkit: Toolkit, itemNames: Map<number, string>, trinkets: number[]): SlotContent | undefined {
	const trinket = /^trinket ([12])$/i.exec(label.trim());
	if (trinket) {
		const id = trinkets[+trinket[1] - 1];
		return id && toolkit.byItem.has(id) ? { kind: 'item', id } : undefined;
	}
	const racial = /^racial ([12])$/i.exec(label.trim());
	const candidates = racial ? (racial[1] === '1' ? racials1 : racials2) : splitAlternatives(label);
	for (const text of candidates) {
		const m = /^(.*?)\s*\((?:r|rank\s*)(\d+)\)\s*$/i.exec(text);
		const spell = m ? toolkit.spellByName(`${m[1]}(Rank ${m[2]})`) : toolkit.spellByName(text);
		if (spell !== undefined) return { kind: 'spell', id: spell };
		const wanted = text.trim().toLowerCase();
		for (const [id, name] of itemNames) {
			const n = name.toLowerCase();
			if (toolkit.byItem.has(id) && (n === wanted || n.endsWith(' ' + wanted))) return { kind: 'item', id };
		}
	}
	return undefined;
}

// "Dark/Demonic Rune" means Dark Rune or Demonic Rune.
function splitAlternatives(label: string): string[] {
	const parts = label.split('/').map(p => p.trim());
	if (parts.length < 2) return parts;
	const last = parts[parts.length - 1].split(/\s+/);
	return parts.map((p, i) => (i < parts.length - 1 && !p.includes(' ') ? [p, ...last.slice(1)].join(' ') : p));
}
