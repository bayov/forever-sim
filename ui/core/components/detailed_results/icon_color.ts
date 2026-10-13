// The main color of an ability's icon, so a bar that belongs to the ability can take its color:
// Flame Shock's bars come out red and Earth Shock's come out yellow.
//
// We draw the icon small on a canvas and look at its pixels. We count each pixel's hue, weighted
// by how vivid it is, and average the pixels of the strongest hue. Dark and grey pixels count
// little, so an icon's black frame and background don't wash the color out. Then we set the
// lightness and saturation to a middle range, so every bar reads on the dark background.
//
// The icons come from wow.zamimg.com, which allows reading them from other sites (CORS). When an
// icon can't load or be read, we return null and the bar keeps its default color.

const SAMPLE_SIZE = 16;
const HUE_BINS = 12;

const cache = new Map<string, Promise<string | null>>();

// Returns the color as "r, g, b", for rgba(var(--bar-rgb), alpha) in the styles.
export const iconColor = (iconUrl: string): Promise<string | null> => {
	let color = cache.get(iconUrl);
	if (!color) {
		color = loadIconColor(iconUrl);
		cache.set(iconUrl, color);
	}
	return color;
};

const loadIconColor = (iconUrl: string): Promise<string | null> =>
	new Promise(resolve => {
		const image = new Image();
		image.crossOrigin = 'anonymous';
		image.onload = () => {
			try {
				resolve(dominantColor(image));
			} catch {
				resolve(null);
			}
		};
		image.onerror = () => resolve(null);
		image.src = iconUrl;
	});

const dominantColor = (image: HTMLImageElement): string | null => {
	const canvas = document.createElement('canvas');
	canvas.width = SAMPLE_SIZE;
	canvas.height = SAMPLE_SIZE;
	const ctx = canvas.getContext('2d', { willReadFrequently: true })!;
	ctx.drawImage(image, 0, 0, SAMPLE_SIZE, SAMPLE_SIZE);
	const pixels = ctx.getImageData(0, 0, SAMPLE_SIZE, SAMPLE_SIZE).data;

	const bins = Array.from({ length: HUE_BINS }, () => ({ weight: 0, r: 0, g: 0, b: 0 }));
	for (let i = 0; i < pixels.length; i += 4) {
		const [r, g, b] = [pixels[i], pixels[i + 1], pixels[i + 2]];
		const [h, s, v] = rgbToHsv(r, g, b);
		const weight = s * v * v;
		if (weight <= 0) continue;
		const bin = bins[Math.floor(h * HUE_BINS) % HUE_BINS];
		bin.weight += weight;
		bin.r += r * weight;
		bin.g += g * weight;
		bin.b += b * weight;
	}

	const best = bins.reduce((a, b) => (b.weight > a.weight ? b : a));
	if (best.weight < 0.5) return null;
	const [h, s] = rgbToHsl(best.r / best.weight, best.g / best.weight, best.b / best.weight);
	const [r, g, b] = hslToRgb(h, Math.min(Math.max(s, 0.45), 0.8), 0.55);
	return `${Math.round(r)}, ${Math.round(g)}, ${Math.round(b)}`;
};

const rgbToHsv = (r: number, g: number, b: number): [number, number, number] => {
	const max = Math.max(r, g, b) / 255;
	const min = Math.min(r, g, b) / 255;
	const [h] = rgbToHsl(r, g, b);
	return [h, max == 0 ? 0 : (max - min) / max, max];
};

const rgbToHsl = (r: number, g: number, b: number): [number, number, number] => {
	r /= 255;
	g /= 255;
	b /= 255;
	const max = Math.max(r, g, b);
	const min = Math.min(r, g, b);
	const l = (max + min) / 2;
	if (max == min) return [0, 0, l];
	const d = max - min;
	const s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
	const h = max == r ? (g - b) / d + (g < b ? 6 : 0) : max == g ? (b - r) / d + 2 : (r - g) / d + 4;
	return [h / 6, s, l];
};

const hslToRgb = (h: number, s: number, l: number): [number, number, number] => {
	const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
	const p = 2 * l - q;
	const channel = (t: number) => {
		if (t < 0) t += 1;
		if (t > 1) t -= 1;
		if (t < 1 / 6) return p + (q - p) * 6 * t;
		if (t < 1 / 2) return q;
		if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
		return p;
	};
	return [channel(h + 1 / 3) * 255, channel(h) * 255, channel(h - 1 / 3) * 255];
};
