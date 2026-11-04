// The game's glue screens (the login and loading screens) and its interface sounds: the loading
// art and bar, the glue buttons and dialog frame for the setup page, and the clicks and panel
// sounds. The textures become CSS variables that styles.css uses under `.glue`. When one fails to
// load we leave `.glue` off, and the page keeps its plain look.

import { decodeBLP, Image } from './wow/blp';

// The client's files by FileDataID.
const tex = {
	// interface/glues/loadingscreens: the Barrens are in Kalimdor. The wide art is a 16:9 picture
	// squeezed into a square texture, and the other one is 4:3.
	artWide: 343002,
	art: 131848,
	// interface/glues/loadingbar
	barBorder: 131811,
	barFill: 131812,
	spark: 131815,
	// interface/glues/common
	bigUp: 131183,
	bigDown: 131179,
	bigHighlight: 131181,
	bigGlow: 131180,
	panelUp: 131173,
	panelDown: 131170,
	panelHighlight: 131172,
	panelDisabled: 131169,
	editBox: 131186,
	logo: 538639,
	// interface/dialogframe and interface/chatframe
	dialogBorder: 131072,
	dialogBackground: 131071,
	dialogHeader: 131080,
	arrowDown: 130952,
};

// Friz Quadrata (fonts/frizqt__.ttf), the font of the game's glue screens.
const glueFont = 615960;

// The interface sounds (sound/interface).
export const uiSounds = {
	// The glue screens' buttons.
	glueButton: 567437,
	// Buttons, dropdowns and checkboxes in the game's interface.
	click: 567481,
	enterWorld: 567419,
	menuOpen: 567490,
	menuClose: 567464,
	spellbookOpen: 567440,
	spellbookClose: 567496,
};

// The login screen music (sound/music/gluescreenmusic/wow_main_theme.mp3).
const loginMusic = 53223;

// Glue files come straight from the asset route, not through wow/files.ts, so they don't count
// in the room's loading bar or in its remembered file list.
const raw = new Map<number, Promise<ArrayBuffer>>();
function fetchFile(fdid: number): Promise<ArrayBuffer> {
	let p = raw.get(fdid);
	if (!p) {
		p = fetch(`/api/assets/${fdid}`).then(r => {
			if (!r.ok) throw new Error(`file ${fdid}: ${r.status}`);
			return r.arrayBuffer();
		});
		p.catch(() => raw.delete(fdid));
		raw.set(fdid, p);
	}
	return p;
}

const images = new Map<number, Promise<Image>>();
function image(fdid: number): Promise<Image> {
	let p = images.get(fdid);
	if (!p) {
		p = fetchFile(fdid).then(decodeBLP);
		images.set(fdid, p);
	}
	return p;
}

function canvasOf(img: Image, x = 0, y = 0, w = img.width, h = img.height): HTMLCanvasElement {
	const c = document.createElement('canvas');
	c.width = w;
	c.height = h;
	c.getContext('2d')!.putImageData(new ImageData(img.rgba, img.width, img.height), -x, -y);
	return c;
}

async function url(fdid: number): Promise<string> {
	return canvasOf(await image(fdid)).toDataURL();
}

// glowUrl is for the textures the game adds on top (button glows and highlights): black is
// nothing there. We turn their brightness into alpha, so they look the same drawn normally.
async function glowUrl(fdid: number): Promise<string> {
	const img = await image(fdid);
	const rgba = new Uint8ClampedArray(img.rgba);
	for (let i = 0; i < rgba.length; i += 4) {
		const a = (Math.max(rgba[i], rgba[i + 1], rgba[i + 2]) * rgba[i + 3]) / 255;
		const k = a > 0 ? 255 / Math.max(rgba[i], rgba[i + 1], rgba[i + 2]) : 0;
		rgba[i] *= k;
		rgba[i + 1] *= k;
		rgba[i + 2] *= k;
		rgba[i + 3] = a;
	}
	return canvasOf({ ...img, rgba }).toDataURL();
}

// border9 turns a backdrop edge file (eight square cells in a row: left, right, top, bottom and
// the four corners) into one 3x3 image for a CSS border-image. The top and bottom cells are
// drawn on their side in the file, like the game draws them.
async function border9(fdid: number): Promise<string> {
	const img = await image(fdid);
	const e = img.height;
	const src = canvasOf(img);
	const c = document.createElement('canvas');
	c.width = c.height = 3 * e;
	const g = c.getContext('2d')!;
	const cell = (i: number, x: number, y: number) => g.drawImage(src, i * e, 0, e, e, x, y, e, e);
	cell(0, 0, e);
	cell(1, 2 * e, e);
	cell(4, 0, 0);
	cell(5, 2 * e, 0);
	cell(6, 0, 2 * e);
	cell(7, 2 * e, 2 * e);
	// Both are turned a quarter clockwise: the top cell's outer side is its left, and the bottom
	// cell's is its right.
	for (const [i, y] of [
		[2, 0],
		[3, 2 * e],
	] as const) {
		g.save();
		g.translate(e * 1.5, y + e / 2);
		g.rotate(Math.PI / 2);
		g.drawImage(src, i * e, 0, e, e, -e / 2, -e / 2, e, e);
		g.restore();
	}
	return c.toDataURL();
}

let fontReady: Promise<void> | undefined;
function loadFont(): Promise<void> {
	fontReady ??= fetchFile(glueFont)
		.then(buf => new FontFace('Friz Quadrata', buf).load())
		.then(f => void document.fonts.add(f))
		.catch(e => console.warn('glue font', e));
	return fontReady;
}

// wideArt says whether the wide loading art fits this window better than the 4:3 one.
function wideArt(): boolean {
	return innerWidth / innerHeight > (16 / 9 + 4 / 3) / 2;
}

// art draws the loading screen picture into a new canvas, which CSS stretches back to its aspect.
export async function art(): Promise<HTMLCanvasElement> {
	const wide = wideArt();
	const c = canvasOf(await image(wide ? tex.artWide : tex.art));
	c.className = 'glue-art';
	c.style.setProperty('--aspect', String(wide ? 16 / 9 : 4 / 3));
	return c;
}

const vars = new Map<string, Promise<void>>();
// cssVars sets the textures as CSS variables on the page, once.
function cssVars(list: Record<string, () => Promise<string>>): Promise<void> {
	return Promise.all(
		Object.entries(list).map(([name, make]) => {
			let p = vars.get(name);
			if (!p) {
				p = make().then(u => document.documentElement.style.setProperty(`--${name}`, `url(${u})`));
				p.catch(() => vars.delete(name));
				vars.set(name, p);
			}
			return p;
		}),
	).then(() => undefined);
}

const loadingVars = { 'load-border': () => url(tex.barBorder), 'load-fill': () => url(tex.barFill), 'load-spark': () => glowUrl(tex.spark) };

// dressLoading turns the room's loading screen into the game's: the art behind and the bar near
// the bottom. We wait for it a few seconds at most. The plain screen shows until it comes.
export async function dressLoading(screen: HTMLElement): Promise<void> {
	const ready = Promise.all([art(), cssVars(loadingVars), loadFont()])
		.then(([canvas]) => {
			if (!screen.isConnected) return;
			screen.prepend(canvas);
			const box = screen.querySelector('.loading-box')!;
			box.append(Object.assign(document.createElement('div'), { className: 'loading-border' }));
			box.querySelector('.loading-bar > div')!.append(Object.assign(document.createElement('div'), { className: 'loading-spark' }));
			screen.classList.add('glue');
		})
		.catch(e => console.warn('loading screen art', e));
	await Promise.race([ready, new Promise(r => setTimeout(r, 4000))]);
}

// dressSetup gives the setup page the glue screen's look: the art dimmed behind, the logo, the
// dialog frame and the glue buttons.
export async function dressSetup(screen: HTMLElement): Promise<void> {
	try {
		const [canvas] = await Promise.all([
			art(),
			cssVars({
				'glue-big-up': () => url(tex.bigUp),
				'glue-big-down': () => url(tex.bigDown),
				'glue-big-highlight': () => glowUrl(tex.bigHighlight),
				'glue-big-glow': () => glowUrl(tex.bigGlow),
				'glue-panel-up': () => url(tex.panelUp),
				'glue-panel-down': () => url(tex.panelDown),
				'glue-panel-highlight': () => glowUrl(tex.panelHighlight),
				'glue-panel-disabled': () => url(tex.panelDisabled),
				'glue-edit': async () => canvasOf(await image(tex.editBox), 50, 14, 156, 36).toDataURL(),
				'glue-logo': () => url(tex.logo),
				'glue-dialog-border': () => border9(tex.dialogBorder),
				'glue-dialog-bg': () => url(tex.dialogBackground),
				'glue-header': async () => canvasOf(await image(tex.dialogHeader), 118, 0, 276, 82).toDataURL(),
				'glue-arrow': () => url(tex.arrowDown),
			}),
			// The loading screen's too, so it shows at once after Enter.
			cssVars(loadingVars),
			loadFont(),
		]);
		screen.prepend(canvas);
		screen.classList.add('glue');
	} catch (e) {
		console.warn('glue screen', e);
	}
}

// ---- Sounds ----

// The room's master volume (Sounds.volume in wow/sound.ts), so a click is as loud there as here.
const volume = 0.6;
let audio: { ctx: AudioContext; out: GainNode } | undefined;
const buffers = new Map<number, Promise<AudioBuffer | undefined>>();

function context() {
	if (!audio) {
		const ctx = new AudioContext();
		const out = ctx.createGain();
		out.gain.value = volume;
		out.connect(ctx.destination);
		audio = { ctx, out };
	}
	return audio;
}

function buffer(fdid: number): Promise<AudioBuffer | undefined> {
	let p = buffers.get(fdid);
	if (!p) {
		p = fetchFile(fdid)
			.then(buf => context().ctx.decodeAudioData(buf))
			.catch(e => {
				console.warn(`ui sound ${fdid}`, e);
				return undefined;
			});
		buffers.set(fdid, p);
	}
	return p;
}

// preloadSounds decodes the interface sounds ahead of the first click.
export function preloadSounds() {
	for (const fdid of Object.values(uiSounds)) void buffer(fdid);
}

// playSound plays an interface sound, flat. Browsers start audio only after a click or key
// press, and every caller is one.
export function playSound(fdid: number) {
	const { ctx, out } = context();
	if (ctx.state !== 'running') void ctx.resume();
	void buffer(fdid).then(b => {
		if (!b) return;
		const s = ctx.createBufferSource();
		s.buffer = b;
		s.connect(out);
		s.start();
	});
}

// ---- Login music ----

const musicKey = 'practice.music';
let music: HTMLAudioElement | undefined;

export function musicOn(): boolean {
	try {
		return localStorage.getItem(musicKey) === '1';
	} catch {
		return false;
	}
}

export function setMusic(on: boolean) {
	try {
		localStorage.setItem(musicKey, on ? '1' : '0');
	} catch {
		// Only this visit keeps it.
	}
	if (on) startMusic();
	else stopMusic();
}

// startMusic plays the login music in a loop. When the browser doesn't let us yet, we try again
// on the first click or key press.
export function startMusic() {
	if (!musicOn()) return;
	if (!music) {
		music = new Audio(`/api/assets/${loginMusic}`);
		music.loop = true;
	}
	music.volume = volume * 0.5;
	music.play().catch(() => {
		const retry = () => {
			removeEventListener('pointerdown', retry);
			removeEventListener('keydown', retry);
			startMusic();
		};
		addEventListener('pointerdown', retry);
		addEventListener('keydown', retry);
	});
}

// stopMusic fades the music out, like the game does when you enter the world.
export function stopMusic(fadeMs = 1500) {
	const m = music;
	if (!m || m.paused) return;
	const from = m.volume;
	const start = performance.now();
	const step = () => {
		const t = (performance.now() - start) / fadeMs;
		if (t >= 1 || m !== music) {
			m.pause();
			m.volume = from;
			return;
		}
		m.volume = from * (1 - t);
		requestAnimationFrame(step);
	};
	step();
}
