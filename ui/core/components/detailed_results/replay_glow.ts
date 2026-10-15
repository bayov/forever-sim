// The Pixel glow of the user's BayovCore addon (Glow.lua), which draws it the way ForeverAuras
// draws it on an Aura (Modern) icon.
//
// Each line is a square turned 45 degrees that travels clockwise around the icon's edge. We
// see it only through four strips along the edges, each exactly as wide as the glow is thick
// and flush inside the icon's 1px border. So a line has slanted, soft ends where it meets a
// strip's edge, and it adds light to the icon under it (additive blending) instead of covering
// it. The left and right strips stop at the top and bottom ones, so a line turning a corner
// isn't drawn twice there, which the additive blending would show as a brighter corner.

export interface GlowOptions {
	color: string;
	lines?: number;
	// Seconds for a line to go once around, 1 / frequency in the aura's settings.
	period?: number;
	// How long a line is, in pixels.
	length?: number;
	thickness?: number;
}

const active = new Set<PixelGlow>();
const glows = new WeakMap<HTMLElement, PixelGlow>();
let frameRequest = 0;

// Where a point this far along the edge is, from the bottom left corner. It goes up the left
// edge, then right along the top, then down and back, which is clockwise.
const along = (distance: number, w: number, h: number): [number, number] => {
	distance = ((distance % (2 * (w + h))) + 2 * (w + h)) % (2 * (w + h));
	if (distance < h) return [0, h - distance];
	distance -= h;
	if (distance < w) return [distance, 0];
	distance -= w;
	if (distance < h) return [w, distance];
	return [w - (distance - h), h];
};

class PixelGlow {
	readonly elem: HTMLElement;
	private readonly options: Required<Omit<GlowOptions, 'length'>> & { length?: number };
	private readonly edges: Array<{ elem: HTMLElement; squares: Array<HTMLElement>; x: number; y: number }> = [];
	private size: { w: number; h: number; side: number } | null = null;

	constructor(host: HTMLElement, options: GlowOptions) {
		this.options = { lines: 8, period: 4, thickness: 2, ...options };
		this.elem = document.createElement('div');
		this.elem.className = 'replay-glow';
		for (let i = 0; i < 4; i++) {
			const edge = document.createElement('div');
			edge.className = 'replay-glow-edge';
			const squares: Array<HTMLElement> = [];
			for (let line = 0; line < this.options.lines; line++) {
				const square = document.createElement('div');
				square.className = 'replay-glow-line';
				square.style.background = this.options.color;
				edge.appendChild(square);
				squares.push(square);
			}
			this.elem.appendChild(edge);
			this.edges.push({ elem: edge, squares, x: 0, y: 0 });
		}
		host.appendChild(this.elem);
	}

	// Lays out the strips and the size of the lines, once we know the icon's size.
	private measure(): boolean {
		const w = this.elem.clientWidth;
		const h = this.elem.clientHeight;
		if (!(w > 0 && h > 0)) return false;
		if (this.size && this.size.w == w && this.size.h == h) return true;
		const t = this.options.thickness;
		const { lines } = this.options;
		const length = Math.min(this.options.length ?? (w + h) * (2 / lines - 0.1), w, h);
		this.size = { w, h, side: Math.max(t, length) / Math.SQRT2 };
		const places = [
			{ x: 0, y: 0, width: w, height: t },
			{ x: 0, y: h - t, width: w, height: t },
			{ x: 0, y: t, width: t, height: h - 2 * t },
			{ x: w - t, y: t, width: t, height: h - 2 * t },
		];
		this.edges.forEach((edge, i) => {
			const place = places[i];
			Object.assign(edge.elem.style, { left: `${place.x}px`, top: `${place.y}px`, width: `${place.width}px`, height: `${place.height}px` });
			edge.x = place.x;
			edge.y = place.y;
			edge.squares.forEach(square => {
				square.style.width = `${this.size!.side}px`;
				square.style.height = `${this.size!.side}px`;
			});
		});
		return true;
	}

	update(now: number) {
		if (!this.measure()) return;
		const { w, h, side } = this.size!;
		const { lines, period } = this.options;
		const progress = (now / 1000 / period) % 1;
		const perimeter = 2 * (w + h);
		for (let line = 0; line < lines; line++) {
			const [x, y] = along((line / lines + progress) * perimeter, w, h);
			this.edges.forEach(edge => {
				edge.squares[line].style.transform = `translate(${(x - edge.x - side / 2).toFixed(2)}px, ${(y - edge.y - side / 2).toFixed(
					2,
				)}px) rotate(45deg)`;
			});
		}
	}
}

const tick = (now: number) => {
	active.forEach(glow => {
		if (!glow.elem.isConnected) {
			active.delete(glow);
		} else {
			glow.update(now);
		}
	});
	frameRequest = active.size > 0 ? requestAnimationFrame(tick) : 0;
};

// Turns the glow on the icon on or off. A glow of another color takes the place of the one
// that's there.
export const setGlow = (host: HTMLElement, options: GlowOptions | null) => {
	let glow = glows.get(host);
	if (glow && (!options || glow.elem.dataset.color != options.color)) {
		glow.elem.remove();
		active.delete(glow);
		glows.delete(host);
		glow = undefined;
	}
	if (!options || glow) return;
	glow = new PixelGlow(host, options);
	glow.elem.dataset.color = options.color;
	glows.set(host, glow);
	active.add(glow);
	glow.update(performance.now());
	if (!frameRequest) frameRequest = requestAnimationFrame(tick);
};

// The pale glow of a buff that's up, and the red one of something we should have up but don't.
export const BUFF_GLOW: GlowOptions = { color: '#ffe6aa', lines: 8, period: 4, length: 12, thickness: 2 };
export const MISSING_GLOW: GlowOptions = { color: '#ff3f37', lines: 8, period: 2, length: 12, thickness: 2 };
