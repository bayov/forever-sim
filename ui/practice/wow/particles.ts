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
import type { ParticleEmitter } from './m2';
import type { Model } from './model';

const maxParticles = 600;

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
function overLife(keys: number[][], times: number[], t: number, out: number[]): number[] {
	const n = keys.length;
	if (n === 0) return out;
	const at = (i: number) => (times.length === n ? times[i] : n === 1 ? 0 : i / (n - 1));
	let i = 0;
	while (i < n - 1 && t > at(i + 1)) i++;
	if (i >= n - 1 || t <= at(i)) {
		const k = t <= at(0) ? keys[0] : keys[n - 1];
		for (let c = 0; c < k.length; c++) out[c] = k[c];
		return out;
	}
	const span = at(i + 1) - at(i);
	const f = span > 0 ? (t - at(i)) / span : 0;
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

class Emitter {
	readonly mesh: THREE.Mesh;
	private readonly particles: Particle[] = [];
	private readonly anchor = new THREE.Object3D();
	private carry = 0;
	private readonly pos: THREE.InstancedBufferAttribute;
	private readonly size: THREE.InstancedBufferAttribute;
	private readonly color: THREE.InstancedBufferAttribute;
	private readonly uvs: THREE.InstancedBufferAttribute;
	private readonly rot: THREE.InstancedBufferAttribute;
	private readonly geometry: THREE.InstancedBufferGeometry;
	// Particles that keep to the emitter (flag 0x10) live in its space, the others in the world's.
	private readonly local: boolean;
	private readonly v = { a: [0], g: [0, 0, 0] };
	private readonly alphaKeys: number[][];
	private readonly cellKeys: number[][];
	private readonly worldScale = new THREE.Vector3();
	private readonly origin = new THREE.Vector3();

	constructor(
		readonly model: Model,
		readonly e: ParticleEmitter,
		world: THREE.Object3D,
	) {
		const bone = model.bones[e.bone] ?? model.bones[0];
		const pivot = model.m2.bones[e.bone]?.pivot ?? [0, 0, 0];
		this.anchor.position.set(e.position[0] - pivot[0], e.position[1] - pivot[1], e.position[2] - pivot[2]);
		bone.add(this.anchor);
		// Flag 0x10 keeps the particles with the emitter. Without it they stay where they were born.
		this.local = (e.flags & 0x10) !== 0;
		this.alphaKeys = e.alphas.map(a => [a]);
		this.cellKeys = e.headCells.map(c => [c]);

		const quad = new THREE.PlaneGeometry(1, 1);
		const g = new THREE.InstancedBufferGeometry();
		g.index = quad.index;
		g.setAttribute('position', quad.getAttribute('position'));
		g.setAttribute('uv', quad.getAttribute('uv'));
		this.pos = new THREE.InstancedBufferAttribute(new Float32Array(maxParticles * 3), 3);
		this.size = new THREE.InstancedBufferAttribute(new Float32Array(maxParticles * 2), 2);
		this.color = new THREE.InstancedBufferAttribute(new Float32Array(maxParticles * 4), 4);
		this.uvs = new THREE.InstancedBufferAttribute(new Float32Array(maxParticles * 4), 4);
		this.rot = new THREE.InstancedBufferAttribute(new Float32Array(maxParticles), 1);
		for (const a of [this.pos, this.size, this.color, this.uvs, this.rot]) a.setUsage(THREE.DynamicDrawUsage);
		g.setAttribute('iPos', this.pos);
		g.setAttribute('iSize', this.size);
		g.setAttribute('iColor', this.color);
		g.setAttribute('iUv', this.uvs);
		g.setAttribute('iRot', this.rot);
		g.instanceCount = 0;
		this.geometry = g;

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
		this.mesh = new THREE.Mesh(g, material);
		this.mesh.frustumCulled = false;
		this.mesh.renderOrder = 10;
		world.add(this.mesh);

		// With flag 0x10000000 the texture field packs three texture indexes, 5 bits each, which
		// the client blends. We draw the first.
		const texIndex = e.flags & 0x10000000 ? e.texture & 0x1f : e.texture;
		const tex = model.m2.textures[texIndex];
		// Until its texture loads the emitter stays hidden: an empty texture reads as solid black.
		this.mesh.visible = false;
		if (tex && tex.type === 0 && tex.fdid > 0) {
			texture(tex.fdid, false, false)
				.then(t => {
					material.uniforms.map.value = t;
					material.uniforms.uSoft.value = e.blend !== 0 && e.blend !== 3 && opaque(t) ? 1 : 0;
					this.mesh.visible = true;
				})
				.catch(err => console.warn('particle texture', err));
		} else {
			// A replaceable particle texture, which we don't have: hide the emitter.
			this.mesh.visible = false;
		}
	}

	private value(track: ParticleEmitter['speed'], fallback = 0): number {
		return this.model.trackValue(track, this.v.a) ? this.v.a[0] : fallback;
	}

	update(dt: number, emitting: boolean) {
		const e = this.e;
		if (!this.mesh.visible) return;
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
				this.particles.splice(i, 1);
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
				const dir =
					e.emitterType === 2
						? new THREE.Vector3(Math.cos(polar) * Math.cos(azimuth), Math.sin(polar), Math.cos(polar) * Math.sin(azimuth))
						: new THREE.Vector3(Math.sin(polar) * Math.cos(azimuth), Math.cos(polar), Math.sin(polar) * Math.sin(azimuth));
				let pos: THREE.Vector3;
				if (e.emitterType === 2) {
					// A sphere: start between the two radii (length the inner, width the outer) and
					// fly outward.
					const r = length + Math.random() * Math.max(0, width - length);
					pos = dir.clone().multiplyScalar(r);
				} else {
					pos = new THREE.Vector3((Math.random() * 2 - 1) * length * 0.5, 0, (Math.random() * 2 - 1) * width * 0.5);
				}
				const v = dir.clone();
				pos.applyMatrix4(world);
				v.transformDirection(world);
				// A particle that keeps to its emitter holds its offset from it: it moves with your
				// hand, but the emitter's turning and animated scale don't fling it about.
				if (this.local) pos.sub(this.origin);
				v.multiplyScalar(speed * (1 + speedVar * (Math.random() * 2 - 1)) * scale);
				const cells = e.tileRows * e.tileCols;
				this.particles.push({
					pos,
					vel: v,
					age: 0,
					life: Math.max(0.05, lifespan * (1 + e.lifespanVary * (Math.random() * 2 - 1))),
					tile: e.flags & 0x10000 ? Math.floor(Math.random() * cells) : -1,
					angle: e.baseSpin + Math.random() * e.baseSpinVary,
					// Flag 0x200 spins half the particles the other way.
					spin: (e.spin + Math.random() * e.spinVary) * (e.flags & 0x200 && Math.random() < 0.5 ? -1 : 1),
				});
			}
			if (this.carry > 1) this.carry = 1;
		} else this.carry = 0;

		// Fill the instance data.
		const n = this.particles.length;
		const color = [255, 255, 255];
		const alpha = [1];
		const size = [1, 1];
		const cellOut = [0];
		const tmp = new THREE.Vector3();
		for (let i = 0; i < n; i++) {
			const p = this.particles[i];
			const t = p.age / p.life;
			overLife(this.e.colors, this.e.colorTimes, t, color);
			overLife(this.alphaKeys, this.e.alphaTimes, t, alpha);
			overLife(this.e.scales, this.e.scaleTimes, t, size);
			tmp.copy(p.pos);
			if (this.local) tmp.add(this.origin);
			this.pos.setXYZ(i, tmp.x, tmp.y, tmp.z);
			this.size.setXY(i, size[0] * 2 * scale, size[1] * 2 * scale);
			this.color.setXYZW(i, srgbToLinear(color[0]), srgbToLinear(color[1]), srgbToLinear(color[2]), Math.max(0, Math.min(1, alpha[0])));
			let cell = p.tile;
			if (cell < 0) {
				cell = this.cellKeys.length ? Math.round(overLife(this.cellKeys, this.e.cellTimes, t, cellOut)[0]) : 0;
			}
			const cols = this.e.tileCols;
			const rows = this.e.tileRows;
			const col = cell % cols;
			const row = Math.floor(cell / cols) % rows;
			this.uvs.setXYZW(i, col / cols, row / rows, (col + 1) / cols, (row + 1) / rows);
			this.rot.setX(i, p.angle + p.spin * p.age);
		}
		this.geometry.instanceCount = n;
		for (const a of [this.pos, this.size, this.color, this.uvs, this.rot]) a.needsUpdate = true;
	}

	dispose() {
		this.mesh.removeFromParent();
		this.geometry.dispose();
		(this.mesh.material as THREE.Material).dispose();
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
