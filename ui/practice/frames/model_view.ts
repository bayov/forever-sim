// Your character's model as a picture, for the character frame's model and the portraits.
//
// We draw the room's own model of you (link.three()) once into an offscreen picture with a
// camera in front of you, read it back and show it in a canvas. Only you are in it: the camera
// sees one layer, and we put that layer on your model and the scene's lights. Nothing runs per
// frame. The character frame draws again when it opens and while you drag to turn the model.

import * as THREE from 'three';
import type { RoomLink } from './link';
import { div } from './panels';

// The layers only our cameras see. The room's camera doesn't see them, so the model and the
// lights keep them for good. The body layer has only your body, without gear models or weapons,
// for measuring how tall you stand.
const MODEL_LAYER = 29;
const BODY_LAYER = 28;

// Where your body stands, measured from a picture of it: its feet and the top of its head, and
// its middle on the ground.
export type Stance = { bottom: number; top: number; x: number; z: number };

type Shot = {
	// The turn around you, in radians, from straight in front.
	yaw: number;
	// 'body' frames all of you, 'face' your head (the portraits).
	framing: 'body' | 'face';
};

const camera = new THREE.PerspectiveCamera(30, 1, 0.05, 100);
camera.layers.set(MODEL_LAYER);
const ortho = new THREE.OrthographicCamera();
ortho.layers.set(BODY_LAYER);
const quat = new THREE.Quaternion();
const forward = new THREE.Vector3();
const look = new THREE.Vector3();
const part = new THREE.Box3();
const savedColor = new THREE.Color();
const savedViewport = new THREE.Vector4();

type Target = { rt?: THREE.WebGLRenderTarget; pixels: Uint8Array };
const picture: Target = { pixels: new Uint8Array(0) };
const measuring: Target = { pixels: new Uint8Array(0) };

// renderTo draws the scene with a camera into a target of the given size and reads its pixels
// back, bottom row first. We change the renderer and the scene only for this one picture and
// put each thing back.
function renderTo(renderer: THREE.WebGLRenderer, scene: THREE.Scene, cam: THREE.Camera, t: Target, w: number, h: number, samples: number): Uint8Array {
	if (!t.rt) {
		t.rt = new THREE.WebGLRenderTarget(w, h, { samples });
		// We want the picture in sRGB like the screen, so the model's shaders are the ones the room
		// already compiled. Three gives a render target sRGB output only when it's the XR one, so
		// we mark ours that way and store the bytes as they come (RGBA8).
		t.rt.texture.colorSpace = THREE.SRGBColorSpace;
		t.rt.texture.internalFormat = 'RGBA8';
		(t.rt as THREE.WebGLRenderTarget & { isXRRenderTarget: boolean }).isXRRenderTarget = true;
	} else if (t.rt.width !== w || t.rt.height !== h) t.rt.setSize(w, h);
	if (t.pixels.length !== w * h * 4) t.pixels = new Uint8Array(w * h * 4);

	const saved = {
		target: renderer.getRenderTarget(),
		autoClear: renderer.autoClear,
		alpha: renderer.getClearAlpha(),
		background: scene.background,
		before: scene.onBeforeRender,
		shadowAuto: renderer.shadowMap.autoUpdate,
		shadowNeeds: renderer.shadowMap.needsUpdate,
	};
	renderer.getClearColor(savedColor);
	renderer.getViewport(savedViewport);
	try {
		// The room's hook prepares the crowd for its own camera, which this picture must not do.
		scene.onBeforeRender = () => {};
		scene.background = null;
		// The shadows keep the room's last shadow map instead of drawing a new one with only you.
		renderer.shadowMap.autoUpdate = false;
		renderer.shadowMap.needsUpdate = false;
		renderer.autoClear = true;
		renderer.setClearColor(0x000000, 0);
		renderer.setRenderTarget(t.rt);
		renderer.render(scene, cam);
		renderer.readRenderTargetPixels(t.rt, 0, 0, w, h, t.pixels);
	} finally {
		renderer.setRenderTarget(saved.target);
		renderer.setViewport(savedViewport);
		renderer.setClearColor(savedColor, saved.alpha);
		renderer.autoClear = saved.autoClear;
		renderer.shadowMap.autoUpdate = saved.shadowAuto;
		renderer.shadowMap.needsUpdate = saved.shadowNeeds;
		scene.background = saved.background;
		scene.onBeforeRender = saved.before;
	}
	return t.pixels;
}

// facing gives the way your model looks, flat on the ground. It looks along +X of its group.
function facing(player: THREE.Object3D, yaw: number): THREE.Vector3 {
	player.getWorldQuaternion(quat);
	return forward.set(1, 0, 0).applyQuaternion(quat).setY(0).normalize().applyAxisAngle(THREE.Object3D.DEFAULT_UP, yaw);
}

// measure finds where your body stands. The meshes' own boxes are in their resting pose, which
// stands taller than the room's pose (an orc's box reaches well over its head), so we draw the
// body alone from the front and look where its pixels are.
export function measure(link: RoomLink): Stance | undefined {
	const three = link.three();
	if (!three) return undefined;
	const { renderer, scene, player } = three;
	player.updateWorldMatrix(true, true);
	// The body is the visible skinned mesh that reaches lowest (gear and weapons hang higher).
	let body: THREE.Mesh | undefined;
	const box = new THREE.Box3();
	player.traverseVisible(o => {
		const m = o as THREE.Mesh;
		if (!(m as THREE.SkinnedMesh).isSkinnedMesh || !m.geometry) return;
		if (!m.geometry.boundingBox) m.geometry.computeBoundingBox();
		part.copy(m.geometry.boundingBox!).applyMatrix4(m.matrixWorld);
		if (part.isEmpty()) return;
		if (!body || part.min.y < box.min.y - 0.01 || (Math.abs(part.min.y - box.min.y) <= 0.01 && part.max.y > box.max.y)) {
			body = m;
			box.copy(part);
		}
	});
	if (!body) return undefined;
	body.layers.enable(BODY_LAYER);
	for (const o of scene.children) if ((o as THREE.Light).isLight) o.layers.enable(BODY_LAYER);

	// An orthographic camera in front, so each row of pixels is one height.
	const size = box.max.y - box.min.y;
	const half = Math.max(size, box.max.x - box.min.x, box.max.z - box.min.z) * 0.75;
	const cx = (box.min.x + box.max.x) / 2;
	const cz = (box.min.z + box.max.z) / 2;
	const cy = (box.min.y + box.max.y) / 2;
	const f = facing(player, 0);
	ortho.left = -half;
	ortho.right = half;
	ortho.top = half;
	ortho.bottom = -half;
	ortho.near = 0.01;
	ortho.far = half * 6;
	ortho.position.set(cx, cy, cz).addScaledVector(f, half * 3);
	ortho.updateProjectionMatrix();
	ortho.lookAt(cx, cy, cz);
	const n = 128;
	const px = renderTo(renderer, scene, ortho, measuring, n, n, 0);
	let lo = -1;
	let hi = -1;
	for (let y = 0; y < n; y++) {
		const row = y * n * 4;
		for (let x = 0; x < n; x++) {
			if (px[row + x * 4 + 3] > 16) {
				if (lo < 0) lo = y;
				hi = y;
				break;
			}
		}
	}
	if (lo < 0) return undefined;
	const yOf = (row: number) => cy - half + (row / n) * 2 * half;
	return { bottom: yOf(lo), top: yOf(hi + 1), x: cx, z: cz };
}

// drawModel renders your model into a canvas of the given pixel size, framed by a stance from
// measure. It says whether it could: the models may not be loaded yet.
export function drawModel(link: RoomLink, canvas: HTMLCanvasElement, width: number, height: number, shot: Shot, stance: Stance): boolean {
	const three = link.three();
	if (!three) return false;
	const { renderer, scene, player } = three;
	const w = Math.max(1, Math.min(2048, Math.round(width)));
	const h = Math.max(1, Math.min(2048, Math.round(height)));
	player.traverse(o => o.layers.enable(MODEL_LAYER));
	for (const o of scene.children) if ((o as THREE.Light).isLight) o.layers.enable(MODEL_LAYER);

	const f = facing(player, shot.yaw);
	const size = stance.top - stance.bottom;
	const tan = Math.tan(THREE.MathUtils.degToRad(camera.fov / 2));
	camera.aspect = w / h;
	if (shot.framing === 'body') {
		// Head to feet from a tenth of the way down to four fifths, like the game's model frame:
		// the stats under the model cover your feet.
		const view = size / 0.7;
		look.set(stance.x, stance.bottom + size / 2 - view * 0.05, stance.z);
		camera.position.copy(look).addScaledVector(f, view / 2 / tan);
	} else {
		// The head and a bit of the shoulders: the eyes sit about an eighth of your height under
		// the top.
		const half = size * 0.13;
		look.set(stance.x, stance.top - size * 0.12, stance.z);
		camera.position.copy(look).addScaledVector(f, half / tan);
	}
	camera.updateProjectionMatrix();
	camera.lookAt(look);
	const px = renderTo(renderer, scene, camera, picture, w, h, 4);

	// The rows come bottom up, with the color times its alpha where the model is see-through.
	if (canvas.width !== w) canvas.width = w;
	if (canvas.height !== h) canvas.height = h;
	const ctx = canvas.getContext('2d');
	if (!ctx) return false;
	const img = ctx.createImageData(w, h);
	const out = img.data;
	const row = w * 4;
	for (let y = 0; y < h; y++) {
		const src = (h - 1 - y) * row;
		const dst = y * row;
		for (let i = 0; i < row; i += 4) {
			const a = px[src + i + 3];
			if (a === 0) continue;
			const k = a === 255 ? 1 : 255 / a;
			out[dst + i] = px[src + i] * k;
			out[dst + i + 1] = px[src + i + 1] * k;
			out[dst + i + 2] = px[src + i + 2] * k;
			out[dst + i + 3] = a;
		}
	}
	ctx.putImageData(img, 0, 0);
	return true;
}

// pixelSize is how many screen pixels an element covers (the layer's zoom and the device's).
function pixelSize(el: HTMLElement): [number, number] {
	const r = el.getBoundingClientRect();
	return [r.width * devicePixelRatio, r.height * devicePixelRatio];
}

// ModelView is the character frame's model: a picture you turn by dragging.
export class ModelView {
	readonly el: HTMLDivElement;
	private readonly canvas = document.createElement('canvas');
	private yaw = 0;
	private pending = 0;
	private stance?: Stance;

	constructor(
		parent: HTMLElement,
		box: { x: number; y: number; w: number; h: number },
		private readonly link: RoomLink,
	) {
		this.el = div(parent, 'bz-model', box);
		this.canvas.className = 'bz-model-canvas';
		this.el.append(this.canvas);
		this.el.addEventListener('mousedown', e => {
			if (e.button !== 0) return;
			e.stopPropagation();
			e.preventDefault();
			let last = e.clientX;
			const move = (m: MouseEvent) => {
				const dx = m.clientX - last;
				last = m.clientX;
				if (!dx) return;
				// Half a turn across the model's width.
				this.yaw += (dx / (this.el.getBoundingClientRect().width || 1)) * Math.PI;
				this.request();
			};
			const up = () => {
				removeEventListener('mousemove', move);
				removeEventListener('mouseup', up);
			};
			addEventListener('mousemove', move);
			addEventListener('mouseup', up);
		});
	}

	// show draws the model facing you, as the game does each time the frame opens.
	show() {
		this.yaw = 0;
		this.stance = measure(this.link);
		this.draw();
	}

	private request() {
		if (this.pending) return;
		this.pending = requestAnimationFrame(() => {
			this.pending = 0;
			this.draw();
		});
	}

	private draw() {
		const [w, h] = pixelSize(this.el);
		const ok = !!this.stance && w >= 2 && h >= 2 && drawModel(this.link, this.canvas, w, h, { yaw: this.yaw, framing: 'body' }, this.stance);
		this.canvas.style.visibility = ok ? '' : 'hidden';
	}
}

// portrait fills a round element with your character's face. It says whether it could (the
// models may not be loaded yet), so the caller can show something else and try again later.
export function portrait(el: HTMLElement, link: RoomLink): boolean {
	let canvas = el.querySelector<HTMLCanvasElement>('canvas.bz-portrait-canvas');
	if (!canvas) {
		canvas = document.createElement('canvas');
		canvas.className = 'bz-portrait-canvas';
		canvas.style.cssText = 'position:absolute;inset:0;width:100%;height:100%;border-radius:50%;background:#000';
		el.append(canvas);
	}
	const [w, h] = pixelSize(el);
	const stance = measure(link);
	const ok = !!stance && w >= 2 && h >= 2 && drawModel(link, canvas, w, h, { yaw: 0, framing: 'face' }, stance);
	canvas.style.visibility = ok ? '' : 'hidden';
	return ok;
}
