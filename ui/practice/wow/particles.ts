// M2 particle emitters: the sparks, glows and smoke of spell effects, totems and weapon imbues.
//
// We follow the M2 format's emitter settings (wowdev.wiki, "M2", particle emitters): each emitter
// spawns particles at its rate from a plane or a sphere around its bone, sends them off at its
// speed within its spread, pulls them down by its gravity, and over each particle's life blends
// its color, alpha and size through three keys (birth, middle and death). A particle is a camera
// facing quad, or a flat one with flag 0x1000, cut from the texture's grid of cells.
//
// What we don't do yet: splines (type 3 emitters), the tail mode, wind and model particles.

import * as THREE from 'three';
import { texture } from './files';
import type { ParticleEmitter, Vec } from './m2';
import type { Model } from './model';

const maxParticles = 600;

// The emitters' scratch values while filling the GPU's data, shared so a frame makes no garbage.
const fill = { color: new Float64Array(3), alpha: new Float64Array(1), size: new Float64Array(2), cellOut: new Float64Array(1), tmp: new THREE.Vector3() };

// While the loading screen warms the shaders up, the emitters draw even with no particles, so the
// renderer compiles their shader. In the fight an empty emitter isn't drawn at all.
export const particleOptions = { drawEmpty: false };

// Every particle and ribbon batch, which fillBatches fills before each frame draws.
const live = new Set<{ fill(): void }>();

// fillBatches puts each batch's particles together for the frame. It runs before the renderer
// sends the frame's vertices to the GPU (the scene's onBeforeRender): filled later, while the
// batch draws, the GPU would get the new list of triangles over the last frame's points, and a
// ribbon's triangles would reach across to another ribbon's points.
export function fillBatches() {
	for (const b of live) b.fill();
}

export function addBatch(b: { fill(): void }) {
	live.add(b);
}

type Particle = {
	pos: THREE.Vector3;
	vel: THREE.Vector3;
	age: number;
	life: number;
	tile: number;
	// The angle at birth and the turn rate (radians a second).
	angle: number;
	spin: number;
};

// Particles that died, kept to be born again, so a burst of sparks makes no garbage.
const freeParticles: Particle[] = [];

const vertexShader = /* glsl */ `
attribute vec3 iPos;
attribute vec2 iSize;
attribute vec4 iColor;
attribute vec4 iUv;
attribute float iRot;
uniform float uFlat;
varying vec2 vUv;
varying vec2 vQuad;
varying vec4 vColor;
void main() {
	float c = cos(iRot);
	float s = sin(iRot);
	vec2 corner = vec2(position.x * c - position.y * s, position.x * s + position.y * c) * iSize;
	vec4 mv;
	if (uFlat > 0.5) {
		mv = viewMatrix * vec4(iPos + vec3(corner.x, 0.0, corner.y), 1.0);
	} else {
		mv = viewMatrix * vec4(iPos, 1.0);
		mv.xy += corner;
	}
	gl_Position = projectionMatrix * mv;
	vUv = vec2(mix(iUv.x, iUv.z, uv.x), mix(iUv.w, iUv.y, uv.y));
	vQuad = uv;
	vColor = iColor;
}`;

const fragmentShader = /* glsl */ `
uniform sampler2D map;
uniform float alphaTest;
uniform float uByColor;
uniform float uSoft;
varying vec2 vUv;
varying vec2 vQuad;
varying vec4 vColor;
void main() {
	vec4 c = texture2D(map, vUv) * vColor;
	// A texture with no see-through part (a cloud tile) would draw a square. It fades out toward
	// the quad's edge instead, into a round puff.
	if (uSoft > 0.5) c.a *= 1.0 - smoothstep(0.35, 1.0, length(vQuad * 2.0 - 1.0));
	if (c.a < alphaTest) discard;
	// Adding by color still keeps to the texture's shape: its see-through part adds nothing.
	if (uByColor > 0.5) c.rgb *= c.a;
	gl_FragColor = c;
	#include <colorspace_fragment>
}`;

// overLife blends keys over a particle's life by their times (0 to 1 of the life). Without times,
// the keys spread evenly (the usual three: birth, middle and death).
function lifeAt(times: number[], n: number, i: number): number {
	return times.length === n ? times[i] : n === 1 ? 0 : i / (n - 1);
}

function overLife(keys: number[][], times: number[], t: number, out: Vec): Vec {
	const n = keys.length;
	if (n === 0) return out;
	let i = 0;
	while (i < n - 1 && t > lifeAt(times, n, i + 1)) i++;
	if (i >= n - 1 || t <= lifeAt(times, n, i)) {
		const k = t <= lifeAt(times, n, 0) ? keys[0] : keys[n - 1];
		for (let c = 0; c < k.length; c++) out[c] = k[c];
		return out;
	}
	const span = lifeAt(times, n, i + 1) - lifeAt(times, n, i);
	const f = span > 0 ? (t - lifeAt(times, n, i)) / span : 0;
	const a = keys[i];
	const b = keys[i + 1];
	for (let c = 0; c < a.length; c++) out[c] = a[c] * (1 - f) + b[c] * f;
	return out;
}

// srgbToLinear turns a 0-255 gamma color key into a linear 0-1 value, as the shader expects.
function srgbToLinear(c: number): number {
	const v = c / 255;
	return v <= 0.04045 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
}

// The particle data a quad needs, by attribute, with how many numbers each takes.
const layout = [
	['iPos', 3],
	['iSize', 2],
	['iColor', 4],
	['iUv', 4],
	['iRot', 1],
] as const;

// A batch draws the particles of every emitter that looks the same (the same texture, blend and
// facing) in one draw call. In the raid test 20 casters each had a few emitters on their hands,
// and they were over 200 draw calls on their own.
//
// The emitters fill their own data, and the batch puts it together before the frame draws.
class Batch {
	readonly mesh: THREE.Mesh;
	private readonly geometry = new THREE.InstancedBufferGeometry();
	private attrs: THREE.InstancedBufferAttribute[] = [];
	private capacity = 0;
	readonly members = new Set<Emitter>();
	// How many particles the members have in all.
	private total = 0;

	constructor(world: THREE.Object3D, material: THREE.ShaderMaterial) {
		const quad = new THREE.PlaneGeometry(1, 1);
		const g = this.geometry;
		g.index = quad.index;
		g.setAttribute('position', quad.getAttribute('position'));
		g.setAttribute('uv', quad.getAttribute('uv'));
		this.grow(64);
		g.instanceCount = 0;
		this.mesh = new THREE.Mesh(g, material);
		this.mesh.frustumCulled = false;
		this.mesh.renderOrder = 10;
		this.mesh.visible = false;
		addBatch(this);
		world.add(this.mesh);
	}

	private grow(n: number) {
		this.capacity = n;
		this.attrs = layout.map(([name, size]) => {
			const a = new THREE.InstancedBufferAttribute(new Float32Array(n * size), size);
			a.setUsage(THREE.DynamicDrawUsage);
			this.geometry.setAttribute(name, a);
			return a;
		});
		// Three keeps the most copies a geometry's attributes hold from its first draw, and draws no
		// more than that.
		(this.geometry as { _maxInstanceCount?: number })._maxInstanceCount = undefined;
	}

	// changed hears that a member's count of particles went from `was` to `now`.
	changed(was: number, now: number) {
		this.total += now - was;
		if (particleOptions.drawEmpty) this.mesh.visible = true;
	}

	leave(e: Emitter) {
		if (!this.members.delete(e)) return;
		this.changed(e.count, 0);
	}

	// fill puts the members' particles together, before the frame draws.
	fill() {
		if (this.total > this.capacity) {
			let n = this.capacity;
			while (n < this.total) n *= 2;
			this.grow(n);
		}
		let at = 0;
		for (const m of this.members) {
			const n = m.count;
			if (n === 0) continue;
			for (let k = 0; k < layout.length; k++) {
				const size = layout[k][1];
				(this.attrs[k].array as Float32Array).set(m.data[k].subarray(0, n * size), at * size);
			}
			at += n;
		}
		this.geometry.instanceCount = at;
		this.mesh.visible = at > 0 || particleOptions.drawEmpty;
		if (at === 0) return;
		for (const a of this.attrs) {
			a.clearUpdateRanges();
			a.addUpdateRange(0, at * a.itemSize);
			a.needsUpdate = true;
		}
	}
}

// The batches of each world (the scene), by how their particles look.
const batches = new WeakMap<THREE.Object3D, Map<string, Batch>>();

class Emitter {
	private readonly particles: Particle[] = [];
	// Where the emitter is: it follows its bone.
	private readonly anchor: THREE.Object3D;
	private carry = 0;
	// The particles' data by attribute (layout), which the batch draws.
	readonly data = layout.map(([, size]) => new Float32Array(maxParticles * size));
	count = 0;
	private batch?: Batch;
	// Particles that keep to the emitter (flag 0x10) live in its space, the others in the world's.
	private readonly local: boolean;
	private readonly v = { a: new Float64Array(1), g: new Float64Array(3) };
	private readonly alphaKeys: number[][];
	private readonly cellKeys: number[][];
	private readonly worldScale = new THREE.Vector3();
	private readonly origin = new THREE.Vector3();
	// Whether the emitter's texture loaded. Until then it draws nothing.
	private ready = false;

	constructor(
		readonly model: Model,
		readonly e: ParticleEmitter,
		world: THREE.Object3D,
	) {
		this.anchor = model.socket(e.bone, e.position);
		// Flag 0x10 keeps the particles with the emitter. Without it they stay where they were born.
		this.local = (e.flags & 0x10) !== 0;
		this.alphaKeys = e.alphas.map(a => [a]);
		this.cellKeys = e.headCells.map(c => [c]);

		const material = new THREE.ShaderMaterial({
			vertexShader,
			fragmentShader,
			uniforms: { map: { value: null }, alphaTest: { value: 0.01 }, uFlat: { value: e.flags & 0x1000 ? 1 : 0 }, uByColor: { value: e.blend === 1 ? 1 : 0 }, uSoft: { value: 0 } },
			transparent: true,
			depthWrite: false,
			side: THREE.DoubleSide,
		});
		// Particle blend modes (wowdev.wiki M2): 0 opaque, 1 add by color, 2 alpha, 3 alpha key,
		// 4 add by alpha.
		switch (e.blend) {
			case 0:
				material.transparent = false;
				material.depthWrite = true;
				break;
			case 1:
				material.blending = THREE.CustomBlending;
				material.blendSrc = THREE.SrcColorFactor;
				material.blendDst = THREE.OneFactor;
				break;
			case 2:
				break;
			case 3:
				material.uniforms.alphaTest.value = 0.7;
				material.depthWrite = true;
				break;
			default:
				material.blending = THREE.AdditiveBlending;
		}
		// Flat particles (flag 0x1000) lie on the ground, like Fire Nova's ring. The ground is at the
		// same height, so we draw them a little in front of it, or the ground hides them.
		if (e.flags & 0x1000) {
			material.polygonOffset = true;
			material.polygonOffsetFactor = -4;
			material.polygonOffsetUnits = -16;
		}

		// With flag 0x10000000 the texture field packs three texture indexes, 5 bits each, which
		// the client blends. We draw the first.
		const texIndex = e.flags & 0x10000000 ? e.texture & 0x1f : e.texture;
		const tex = model.m2.textures[texIndex];
		// Until its texture loads the emitter has no batch and draws nothing: an empty texture reads
		// as solid black. A replaceable particle texture, which we don't have, never draws.
		if (tex && tex.type === 0 && tex.fdid > 0) {
			texture(tex.fdid, false, false)
				.then(t => {
					if (this.disposed) return;
					const soft = e.blend !== 0 && e.blend !== 3 && opaque(t);
					material.uniforms.map.value = t;
					material.uniforms.uSoft.value = soft ? 1 : 0;
					const key = `${t.id}:${e.blend}:${e.flags & 0x1000}:${soft}`;
					let byKey = batches.get(world);
					if (!byKey) batches.set(world, (byKey = new Map()));
					let batch = byKey.get(key);
					if (!batch) byKey.set(key, (batch = new Batch(world, material)));
					else material.dispose();
					batch.members.add(this);
					this.batch = batch;
					this.ready = true;
				})
				.catch(err => console.warn('particle texture', err));
		}
	}

	private disposed = false;

	private value(track: ParticleEmitter['speed'], fallback = 0): number {
		return this.model.trackValue(track, this.v.a) ? this.v.a[0] : fallback;
	}

	update(dt: number, emitting: boolean) {
		const e = this.e;
		if (!this.ready) return;
		this.anchor.updateWorldMatrix(true, false);
		const world = this.anchor.matrixWorld;
		this.origin.setFromMatrixPosition(world);
		const enabled = this.model.trackValue(e.enabled, this.v.a) ? this.v.a[0] !== 0 : true;
		// Sizes, speeds and gravity grow with the model's scale (a kit's scale, a creature's), but
		// not with its bones' animated scale: a Fire Nova ring grows by moving its emitter out.
		this.worldScale.setFromMatrixScale(this.model.root.matrixWorld);
		const scale = this.worldScale.x || 1;
		// Gravity is an acceleration: down by a number, or a packed direction (flag 0x800000).
		const g = this.v.g;
		if (e.gravity.kind === 'gravity') {
			if (!this.model.trackValue(e.gravity, g)) g[0] = g[1] = g[2] = 0;
		} else {
			g[0] = g[2] = 0;
			g[1] = -this.value(e.gravity);
		}
		const drag = e.drag;

		// Age and move the particles.
		for (let i = this.particles.length - 1; i >= 0; i--) {
			const p = this.particles[i];
			p.age += dt;
			if (p.age >= p.life) {
				// The last one takes its place (the order doesn't matter), and the dead one waits to be
				// born again.
				const list = this.particles;
				list[i] = list[list.length - 1];
				list.pop();
				freeParticles.push(p);
				continue;
			}
			p.vel.x += g[0] * scale * dt;
			p.vel.y += g[1] * scale * dt;
			p.vel.z += g[2] * scale * dt;
			if (drag > 0) p.vel.multiplyScalar(Math.max(0, 1 - drag * dt));
			p.pos.addScaledVector(p.vel, dt);
		}

		// Spawn new ones.
		if (emitting && enabled) {
			const rate = this.value(e.emissionRate) * (1 + e.emissionVary * (Math.random() * 2 - 1));
			this.carry += Math.max(0, rate) * dt;
			const lifespan = this.value(e.lifespan, 1);
			const speed = this.value(e.speed);
			const speedVar = this.value(e.speedVariation);
			const vRange = this.value(e.verticalRange);
			const hRange = this.value(e.horizontalRange);
			const length = this.value(e.areaLength);
			const width = this.value(e.areaWidth);
			while (this.carry >= 1 && this.particles.length < maxParticles) {
				this.carry -= 1;
				const polar = vRange * (Math.random() * 2 - 1);
				const azimuth = hRange > 0 ? hRange * (Math.random() * 2 - 1) : Math.random() * Math.PI * 2;
				// The emitter's up is WoW's z, which is y here. A plane sends its particles up in a cone
				// (the vertical range is the angle from up). A sphere sends them out around it (the
				// vertical range is the angle above and below level): a Fire Nova ring with none
				// spreads flat.
				const p = freeParticles.pop() ?? { pos: new THREE.Vector3(), vel: new THREE.Vector3(), age: 0, life: 1, tile: -1, angle: 0, spin: 0 };
				const dir = p.vel;
				if (e.emitterType === 2) dir.set(Math.cos(polar) * Math.cos(azimuth), Math.sin(polar), Math.cos(polar) * Math.sin(azimuth));
				else dir.set(Math.sin(polar) * Math.cos(azimuth), Math.cos(polar), Math.sin(polar) * Math.sin(azimuth));
				const pos = p.pos;
				if (e.emitterType === 2) {
					// A sphere: start between the two radii (length the inner, width the outer) and
					// fly outward.
					const r = length + Math.random() * Math.max(0, width - length);
					pos.copy(dir).multiplyScalar(r);
				} else {
					pos.set((Math.random() * 2 - 1) * length * 0.5, 0, (Math.random() * 2 - 1) * width * 0.5);
				}
				const v = dir;
				pos.applyMatrix4(world);
				v.transformDirection(world);
				// A particle that keeps to its emitter holds its offset from it: it moves with your
				// hand, but the emitter's turning and animated scale don't fling it about.
				if (this.local) pos.sub(this.origin);
				v.multiplyScalar(speed * (1 + speedVar * (Math.random() * 2 - 1)) * scale);
				const cells = e.tileRows * e.tileCols;
				p.age = 0;
				p.life = Math.max(0.05, lifespan * (1 + e.lifespanVary * (Math.random() * 2 - 1)));
				p.tile = e.flags & 0x10000 ? Math.floor(Math.random() * cells) : -1;
				p.angle = e.baseSpin + Math.random() * e.baseSpinVary;
				// Flag 0x200 spins half the particles the other way.
				p.spin = (e.spin + Math.random() * e.spinVary) * (e.flags & 0x200 && Math.random() < 0.5 ? -1 : 1);
				this.particles.push(p);
			}
			if (this.carry > 1) this.carry = 1;
		} else this.carry = 0;

		// Fill the instance data.
		const n = this.particles.length;
		const { color, alpha, size, cellOut, tmp } = fill;
		const [pos, sizes, colors, uvs, rots] = this.data;
		color.fill(255);
		alpha[0] = 1;
		size[0] = size[1] = 1;
		cellOut[0] = 0;
		for (let i = 0; i < n; i++) {
			const p = this.particles[i];
			const t = p.age / p.life;
			overLife(this.e.colors, this.e.colorTimes, t, color);
			overLife(this.alphaKeys, this.e.alphaTimes, t, alpha);
			overLife(this.e.scales, this.e.scaleTimes, t, size);
			tmp.copy(p.pos);
			if (this.local) tmp.add(this.origin);
			pos[i * 3] = tmp.x;
			pos[i * 3 + 1] = tmp.y;
			pos[i * 3 + 2] = tmp.z;
			sizes[i * 2] = size[0] * 2 * scale;
			sizes[i * 2 + 1] = size[1] * 2 * scale;
			colors[i * 4] = srgbToLinear(color[0]);
			colors[i * 4 + 1] = srgbToLinear(color[1]);
			colors[i * 4 + 2] = srgbToLinear(color[2]);
			colors[i * 4 + 3] = Math.max(0, Math.min(1, alpha[0]));
			let cell = p.tile;
			if (cell < 0) {
				cell = this.cellKeys.length ? Math.round(overLife(this.cellKeys, this.e.cellTimes, t, cellOut)[0]) : 0;
			}
			const cols = this.e.tileCols;
			const rows = this.e.tileRows;
			const col = cell % cols;
			const row = Math.floor(cell / cols) % rows;
			uvs[i * 4] = col / cols;
			uvs[i * 4 + 1] = row / rows;
			uvs[i * 4 + 2] = (col + 1) / cols;
			uvs[i * 4 + 3] = (row + 1) / rows;
			rots[i] = p.angle + p.spin * p.age;
		}
		const was = this.count;
		this.count = n;
		this.batch?.changed(was, n);
	}

	dispose() {
		this.disposed = true;
		for (const p of this.particles) freeParticles.push(p);
		this.particles.length = 0;
		this.batch?.leave(this);
		this.batch = undefined;
		this.anchor.removeFromParent();
	}
}

// Particles runs a model's emitters. Their particles live in `world` (the scene), so they stay
// where they were born when the model moves on.
export class Particles {
	private readonly emitters: Emitter[];
	emitting = true;

	constructor(model: Model, world: THREE.Object3D) {
		this.emitters = model.m2.particles.map(e => new Emitter(model, e, world));
	}

	update(dt: number) {
		for (const e of this.emitters) e.update(dt, this.emitting);
	}

	dispose() {
		for (const e of this.emitters) e.dispose();
	}
}

// opaque is whether a texture has no see-through pixel at all.
const opaqueCache = new WeakMap<THREE.Texture, boolean>();
function opaque(t: THREE.Texture): boolean {
	let v = opaqueCache.get(t);
	if (v === undefined) {
		const data = (t.image as { data?: ArrayLike<number> } | undefined)?.data;
		v = !!data && data.length > 0;
		if (data) {
			for (let i = 3; i < data.length; i += 4) {
				if (data[i] < 250) {
					v = false;
					break;
				}
			}
		}
		opaqueCache.set(t, v);
	}
	return v;
}
