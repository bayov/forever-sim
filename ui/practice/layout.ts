// Where the HUD's frames sit and how they look, read from the user's EllesmereUI profile
// (cmd/practice/ui.go) so the room looks like their game. Without a profile, the room uses the
// same numbers as ours.
//
// Positions are in WoW's UI units, from the middle of the screen with y up, like the profile
// stores them. The screen is 768 / uiScale units tall (1200 at our 0.64), so a frame takes the
// same share of the window as in the game, whatever the window's size.

export type Point = 'CENTER' | 'LEFT' | 'RIGHT' | 'TOP' | 'BOTTOM' | 'TOPLEFT' | 'TOPRIGHT' | 'BOTTOMLEFT' | 'BOTTOMRIGHT';

// A rectangle on screen: its anchor point is at (x, y).
export type Rect = { x: number; y: number; w: number; h: number; point: Point };

export type Color = { r: number; g: number; b: number };

export type BarLayout = {
	key: string;
	first: number; // the action slot of its first button
	rect: Rect;
	cols: number;
	rows: number;
	size: number;
	padding: number;
	showEmpty: boolean;
	macroColor: Color;
	keyFont: number;
	macroFont: number;
};

export type HudLayout = {
	uiHeight: number;
	player: Rect;
	playerColor: Color;
	playerTextSize: number;
	target: Rect;
	targetColor: Color;
	targetTextSize: number;
	targetDebuffSize: number;
	targetOfTarget: Rect;
	castbar: Rect;
	castColor: Color;
	castTimerSize: number;
	castTextSize: number;
	castLatency: { r: number; g: number; b: number; a: number };
	manaColor: Color;
	meter: Rect;
	meterBarHeight: number;
	tooltip: { x: number; y: number };
	nameplate: { w: number; h: number };
	bars: BarLayout[];
};

// The action slots of each EllesmereUI bar (EUI_ActionBars_Secure.lua, BAR_SLOT_OFFSETS).
const barFirstSlot: Record<string, number> = {
	MainBar: 1,
	Bar2: 61,
	Bar3: 49,
	Bar4: 25,
	Bar5: 37,
	Bar6: 145,
	Bar7: 157,
	Bar8: 169,
	Bar9: 13,
	Bar10: 109,
	Bar11: 73,
	Bar12: 85,
	Bar13: 97,
};

type Any = Record<string, any>; // eslint-disable-line @typescript-eslint/no-explicit-any

function num(v: unknown, fallback: number): number {
	return typeof v === 'number' && isFinite(v) ? v : fallback;
}

function color(v: unknown, fallback: Color): Color {
	const c = v as Any | undefined;
	return c && typeof c.r === 'number' ? { r: c.r, g: c.g, b: c.b } : fallback;
}

export function css(c: Color, a = 1): string {
	return `rgba(${Math.round(c.r * 255)}, ${Math.round(c.g * 255)}, ${Math.round(c.b * 255)}, ${a})`;
}

// Our own profile's bars, for a character with no EllesmereUI: nine blocks of 4 x 3 buttons
// along the bottom and four smaller ones at the left.
function defaultBars(): Any {
	const bars: Any = {};
	const positions: Any = {};
	const bottom = ['MainBar', 'Bar2', 'Bar3', 'Bar4', 'Bar5', 'Bar6', 'Bar7', 'Bar8', 'Bar9'];
	bottom.forEach((key, i) => {
		positions[key] = { point: 'LEFT', x: -558.3 + i * 124.5, y: -552.5 };
		bars[key] = {};
	});
	['Bar10', 'Bar11', 'Bar12', 'Bar13'].forEach((key, i) => {
		positions[key] = { point: 'LEFT', x: -1063.8 + i * 117.8, y: -262.8 };
		bars[key] = { buttonWidth: 28, buttonHeight: 28, alwaysShowButtons: false };
	});
	return { bars, barPositions: positions };
}

const mainBarDefaults = { buttonWidth: 29.44, buttonHeight: 29.44, buttonPadding: 1.67, overrideNumIcons: 12, overrideNumRows: 3 };

export function readLayout(ui: Any | undefined): HudLayout {
	const uf: Any = ui?.unitFrames ?? {};
	const pos: Any = uf.positions ?? {};
	const player: Any = uf.player ?? {};
	const target: Any = uf.target ?? {};
	const tot: Any = uf.targettarget ?? {};
	const rb: Any = ui?.resourceBars ?? {};
	const cast: Any = rb.castBar ?? {};
	const dm: Any = ui?.damageMeter ?? {};
	const win: Any = dm.window ?? {};
	const np: Any = ui?.nameplates ?? {};
	const tip: Any = ui?.tooltipFixedPos ?? {};
	const ab: Any = ui?.actionBars ?? defaultBars();
	const green = { r: 0.31, g: 0.616, b: 0.239 };

	const bars: BarLayout[] = [];
	const mainSettings: Any = { ...mainBarDefaults, ...(ab.bars?.MainBar ?? {}) };
	for (const [key, first] of Object.entries(barFirstSlot)) {
		const p: Any | undefined = ab.barPositions?.[key];
		const own: Any | undefined = ab.bars?.[key];
		if (!p || !own || own.alwaysHidden || own.barVisibility === 'never') continue;
		// Each bar keeps the main bar's settings it didn't change.
		const s: Any = { ...mainSettings, ...own };
		const size = num(s.buttonWidth, 29.44);
		const padding = num(s.buttonPadding, 1.67);
		const count = num(s.overrideNumIcons, 12);
		const rows = Math.max(1, num(s.overrideNumRows, 1));
		const cols = Math.ceil(count / rows);
		bars.push({
			key,
			first,
			rect: { x: num(p.x, 0), y: num(p.y, 0), w: cols * size + (cols - 1) * padding, h: rows * size + (rows - 1) * padding, point: (p.point ?? 'CENTER') as Point },
			cols,
			rows,
			size,
			padding,
			showEmpty: s.alwaysShowButtons !== false,
			macroColor: color(s.macroFontColor, { r: 0.784, g: 0.597, b: 0.321 }),
			keyFont: num(s.keybindFontSize, 11),
			macroFont: num(s.macroFontSize, 8),
		});
	}

	const playerPos: Any = pos.player ?? { x: -313.9, y: -228.3 };
	const targetPos: Any = pos.target ?? { x: 313.9, y: -228.1 };
	const totPos: Any = pos.targettarget ?? { x: 485, y: -232.8 };
	const castPos: Any = cast.unlockPos ?? { x: 0, y: -314.4 };
	const winPos: Any = win.position ?? { x: 707.8, y: -490 };
	const scale = num(ui?.uiScale, 0.64);
	return {
		uiHeight: 768 / (scale > 0.3 ? scale : 0.64),
		player: { x: num(playerPos.x, -313.9), y: num(playerPos.y, -228.3), w: num(player.frameWidth, 237.8), h: num(player.healthHeight, 30), point: 'CENTER' },
		playerColor: color(player.customFillColor, green),
		playerTextSize: num(player.rightTextSize, 16),
		// EllesmereUI's target frame is its health bar and the power bar under it (6 high unless
		// changed). A target with no power, like the dummy, fills the whole frame with health.
		target: {
			x: num(targetPos.x, 313.9),
			y: num(targetPos.y, -228.1),
			w: num(target.frameWidth, 237.8),
			h: num(target.healthHeight, 23.9) + (['below', 'above', undefined].includes(target.powerPosition) ? num(target.powerHeight, 6) : 0),
			point: 'CENTER',
		},
		targetColor: color(target.customFillColor, green),
		targetTextSize: num(target.leftTextSize, 16),
		targetDebuffSize: num(target.debuffSize, 30),
		targetOfTarget: { x: num(totPos.x, 485), y: num(totPos.y, -232.8), w: num(tot.frameWidth, 100), h: num(tot.healthHeight, 20), point: 'CENTER' },
		castbar: { x: num(castPos.x, 0), y: num(castPos.y, -314.4), w: num(cast.width, 337.8), h: num(cast.height, 22.2), point: 'CENTER' },
		castColor: { r: num(cast.fillR, 0.447), g: num(cast.fillG, 0.537), b: num(cast.fillB, 0.551) },
		castTimerSize: num(cast.timerSize, 16),
		castTextSize: num(cast.spellTextSize, 12),
		castLatency: { r: num(cast.latencyR, 1), g: num(cast.latencyG, 0.537), b: num(cast.latencyB, 0.537), a: num(cast.latencyA, 0.3) },
		manaColor: color(ui?.customColors?.power?.MANA, { r: 0.296, g: 0.499, b: 0.759 }),
		meter: { x: num(winPos.x, 707.8), y: num(winPos.y, -490), w: num(win.width, 220), h: num(win.height, 220), point: 'CENTER' },
		meterBarHeight: num(dm.barHeight, 16),
		tooltip: { x: num(tip.centerX, 921.1), y: num(tip.centerY, -291.9) },
		nameplate: { w: 120 + num(np.healthBarWidth, 0), h: num(np.healthBarHeight, 13) },
		bars,
	};
}

// The Bayov class pack (bayovui/auras/common/anchors/group.py), in UI units: it is drawn in 4K
// pixels there, 1.8 to a unit at our scale. The resource bars' bottom lines up with the player
// frame's bottom, the maintain cells sit over them, and the main and utility bars under them.
export const pack = {
	iconW: 68 / 1.8,
	iconH: 48 / 1.8,
	gap: 4 / 1.8,
	powerH: 24 / 1.8,
	swingH: 6 / 1.8,
	pipH: 7 / 1.8,
	perRow: 9,
	get width() {
		return this.perRow * this.iconW + (this.perRow - 1) * this.gap;
	},
};

// place puts an element on its rect, with u pixels to a UI unit.
export function place(el: HTMLElement, r: Rect, u: number, W: number, H: number) {
	const ax = r.point.includes('LEFT') ? 0 : r.point.includes('RIGHT') ? 1 : 0.5;
	const ay = r.point.includes('TOP') ? 1 : r.point.includes('BOTTOM') ? 0 : 0.5;
	const left = W / 2 + (r.x - r.w * ax) * u;
	const top = H / 2 - (r.y + r.h * (1 - ay)) * u;
	el.style.left = `${Math.round(left)}px`;
	el.style.top = `${Math.round(top)}px`;
	el.style.width = `${Math.round(r.w * u)}px`;
	el.style.height = `${Math.round(r.h * u)}px`;
}
