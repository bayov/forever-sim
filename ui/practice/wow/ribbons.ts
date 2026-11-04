// M2 ribbon emitters: the trails behind a bolt in flight, the streaks of a hand glow, the swirl
// of a totem's flames.
//
// We follow the M2 format's ribbon settings (wowdev.wiki, "M2", ribbon emitters). A ribbon hangs
// off a bone. While it's visible, it leaves an edge behind it at its rate (edgesPerSecond): a
// short segment across the bone, from heightAbove over it to heightBelow under it. Each edge
// lives for edgeLifetime seconds, falls with the ribbon's gravity, and the edges joined up make
// a strip that fades out toward its tail. Its color, alpha and heights come from the emitter's
// tracks at the model's current animation.
//
// What we don't do yet: the texture slot track (we draw the first texture cell's column) and
// the priority plane (ribbons draw after the models, like particles).

import * as THREE from 'three';
import { texture } from './files';
import type { RibbonEmitter } from './m2';
import type { Model } from './model';
import { addBatch, drawOrder, particleOptions } from './particles';

// The most edges one ribbon keeps, whatever its rate and lifetime say.
const maxEdges = 256;

type Edge = { top: THREE.Vector3; bottom: THREE.Vector3; fall: number; age: number; r: number; g: number; b: number; a: number };
const freeEdges: Edge[] = [];

const vertexShader = /* glsl */ `
attribute vec4 color;
varying vec2 vUv;
varying vec4 vColor;
void main() {
	vUv = uv;
	vColor = color;
	gl_Position = projectionMatrix * viewMatrix * vec4(position, 1.0);
}`;

const fragmentShader = /* glsl */ `
uniform sampler2D map;
uniform float alphaTest;
varying vec2 vUv;
varying vec4 vColor;
void main() {
	vec4 c = texture2D(map, vUv) * vColor;
	if (c.a < alphaTest) discard;
	gl_FragColor = c;
	#include <colorspace_fragment>
}`;

// The strip data a point needs, one point after the other, STRIDE numbers each: by attribute,
// where it starts and how many numbers it takes. Interleaved, a batch's points go to the GPU in
// one upload.
const layout = [
	['position', 0, 3],
	['uv', 3, 2],
	['color', 5, 4],
] as const;
const STRIDE = 9;

// A batch draws every ribbon that looks the same (the same texture and blend) in one draw call,
// like the particles' batches (particles.ts). The ribbons fill their own points, and the batch
// puts them together, with the triangles between each ribbon's points, before the frame draws.
class Batch {
	readonly mesh: THREE.Mesh;
	private readonly geometry = new THREE.BufferGeometry();
	private buffer!: THREE.InterleavedBuffer;
	private index!: THREE.BufferAttribute;
	private capacity = 0;
	readonly members = new Set<Ribbon>();
	// How many points the members have in all.
	private total = 0;

	constructor(world: THREE.Object3D, material: THREE.ShaderMaterial) {
		this.grow(512);
		this.geometry.setDrawRange(0, 0);
		this.mesh = new THREE.Mesh(this.geometry, material);
		this.mesh.frustumCulled = false;
		this.mesh.renderOrder = drawOrder.ribbons;
		this.mesh.visible = false;
		addBatch(this);
		world.add(this.mesh);
	}

	private grow(n: number) {
		this.capacity = n;
		this.buffer = new THREE.InterleavedBuffer(new Float32Array(n * STRIDE), STRIDE);
		this.buffer.setUsage(THREE.DynamicDrawUsage);
		for (const [name, offset, size] of layout) this.geometry.setAttribute(name, new THREE.InterleavedBufferAttribute(this.buffer, size, offset));
		// Each point after a ribbon's first closes a quad of two triangles.
		this.index = new THREE.BufferAttribute(new Uint32Array(n * 3), 1);
		this.index.setUsage(THREE.DynamicDrawUsage);
		this.geometry.setIndex(this.index);
	}

	changed(was: number, now: number) {
		this.total += now - was;
		if (particleOptions.drawEmpty) this.mesh.visible = true;
	}

	leave(r: Ribbon) {
		if (!this.members.delete(r)) return;
		this.changed(r.count < 2 ? 0 : r.count * 2, 0);
	}

	fill() {
		if (this.total > this.capacity) {
			let n = this.capacity;
			while (n < this.total) n *= 2;
			this.grow(n);
		}
		const index = this.index.array as Uint32Array;
		let at = 0;
		let tris = 0;
		for (const r of this.members) {
			// n edges make 2n points.
			const points = r.count * 2;
			if (r.count < 2) continue;
			// A plain loop: a subarray is a new object each time.
			const from = r.data;
			const to = this.buffer.array as Float32Array;
			const o = at * STRIDE;
			for (let j = 0; j < points * STRIDE; j++) to[o + j] = from[j];
			for (let i = 0; i < r.count - 1; i++) {
				const a = at + i * 2;
				index[tris++] = a;
				index[tris++] = a + 1;
				index[tris++] = a + 2;
				index[tris++] = a + 1;
				index[tris++] = a + 3;
				index[tris++] = a + 2;
			}
			at += points;
		}
		this.geometry.setDrawRange(0, tris);
		this.mesh.visible = tris > 0 || particleOptions.drawEmpty;
		if (tris === 0) return;
		this.buffer.clearUpdateRanges();
		this.buffer.addUpdateRange(0, at * STRIDE);
		this.buffer.needsUpdate = true;
		this.index.clearUpdateRanges();
		this.index.addUpdateRange(0, tris);
		this.index.needsUpdate = true;
	}
}

// The batches of each world (the scene), by how their ribbons look.
const batches = new WeakMap<THREE.Object3D, Map<string, Batch>>();

class Ribbon {
	// Where the emitter is: it follows its bone.
	private readonly anchor: THREE.Object3D;
	private readonly edges: Edge[] = [];
	private carry = 0;
	// The strip's points (layout), which the batch draws.
	readonly data = new Float32Array((maxEdges + 1) * 2 * STRIDE);
	// The head (where the ribbon is now), as an edge for put.
	private readonly head: Edge = { top: new THREE.Vector3(), bottom: new THREE.Vector3(), fall: 0, age: 0, r: 1, g: 1, b: 1, a: 1 };
	// put's life and cell size, in a Float64Array so they aren't boxed on each call.
	private readonly strip = new Float64Array(3);
	// How many edges the strip has now.
	count = 0;
	private batch?: Batch;
	private disposed = false;
	private readonly v = new Float64Array(3);
	private readonly tint = new THREE.Color();
	private readonly top = new THREE.Vector3();
	private readonly bottom = new THREE.Vector3();
	// A ribbon with a replaceable texture, which we don't have, never shows.
	private hidden = false;

	constructor(
		readonly model: Model,
		readonly e: RibbonEmitter,
		world: THREE.Object3D,
	) {
		this.anchor = model.socket(e.bone, e.position);

		// Two vertices per edge, plus the head (where the ribbon is right now).
		const material = new THREE.ShaderMaterial({
			vertexShader,
			fragmentShader,
			uniforms: { map: { value: null }, alphaTest: { value: 0.004 } },
			transparent: true,
			depthWrite: false,
			side: THREE.DoubleSide,
		});
		const custom = (src: THREE.BlendingSrcFactor, dst: THREE.BlendingDstFactor) => {
			material.blending = THREE.CustomBlending;
			material.blendSrc = src;
			material.blendDst = dst;
		};
		// The model's blend modes, as for its batches (wowdev.wiki M2 materials).
		const mat = model.m2.materials[e.materials[0] ?? 0] ?? { flags: 0, blend: 0 };
		switch (mat.blend) {
			case 0:
				material.transparent = false;
				break;
			case 1:
				material.uniforms.alphaTest.value = 224 / 255;
				break;
			case 2:
				break;
			case 3:
				custom(THREE.OneFactor, THREE.OneFactor);
				break;
			case 4:
				custom(THREE.SrcAlphaFactor, THREE.OneFactor);
				break;
			case 5:
				custom(THREE.DstColorFactor, THREE.ZeroFactor);
				break;
			case 6:
				custom(THREE.DstColorFactor, THREE.SrcColorFactor);
				break;
			case 7:
				custom(THREE.OneFactor, THREE.OneMinusSrcAlphaFactor);
				break;
		}
		// The ribbon draws once its texture is there. A replaceable ribbon texture, which we don't
		// have, never draws.
		const tex = model.m2.textures[e.textures[0] ?? -1];
		if (tex && tex.type === 0 && tex.fdid > 0) {
			texture(tex.fdid, (tex.flags & 1) !== 0, (tex.flags & 2) !== 0)
				.then(t => {
					if (this.disposed) return;
					material.uniforms.map.value = t;
					const key = `${t.id}:${mat.blend}`;
					let byKey = batches.get(world);
					if (!byKey) batches.set(world, (byKey = new Map()));
					let batch = byKey.get(key);
					if (!batch) byKey.set(key, (batch = new Batch(world, material)));
					else material.dispose();
					batch.members.add(this);
					this.batch = batch;
				})
				.catch(err => console.warn('ribbon texture', err));
		} else this.hidden = true;
	}

	private value(track: RibbonEmitter['heightAbove'], fallback: number): number {
		return this.model.trackValue(track, this.v) ? this.v[0] : fallback;
	}

	update(dt: number, emitting: boolean) {
		const e = this.e;
		if (this.hidden) return;
		this.anchor.updateWorldMatrix(true, false);
		const world = this.anchor.matrixWorld;
		const life = Math.max(0.01, e.edgeLifetime);

		// Age the edges and let them fall.
		for (const edge of this.edges) {
			edge.age += dt;
			if (e.gravity !== 0) {
				edge.fall += e.gravity * dt;
				edge.top.y -= edge.fall * dt;
				edge.bottom.y -= edge.fall * dt;
			}
		}
		// The newest edges come first, so the ones whose time is up are at the end. They wait to
		// be used again, so a ribbon makes no garbage.
		const edges = this.edges;
		while (edges.length > 0 && edges[edges.length - 1].age >= life) freeEdges.push(edges.pop()!);

		// Where the ribbon is now, and how it looks.
		const shown = emitting && this.value(e.visibility, 1) !== 0;
		const above = this.value(e.heightAbove, 0);
		const below = this.value(e.heightBelow, 0);
		this.top.set(0, above, 0).applyMatrix4(world);
		this.bottom.set(0, -below, 0).applyMatrix4(world);
		if (this.model.trackValue(e.color, this.v)) this.tint.setRGB(this.v[0], this.v[1], this.v[2], THREE.SRGBColorSpace);
		else this.tint.setRGB(1, 1, 1);
		const alpha = Math.max(0, Math.min(1, this.value(e.alpha, 32767) / 32767));

		if (shown) {
			this.carry += e.edgesPerSecond * dt;
			if (this.carry >= 1) {
				this.carry -= Math.floor(this.carry);
				const edge = freeEdges.pop() ?? { top: new THREE.Vector3(), bottom: new THREE.Vector3(), fall: 0, age: 0, r: 1, g: 1, b: 1, a: 1 };
				edge.top.copy(this.top);
				edge.bottom.copy(this.bottom);
				edge.fall = 0;
				edge.age = 0;
				edge.r = this.tint.r;
				edge.g = this.tint.g;
				edge.b = this.tint.b;
				edge.a = alpha;
				this.edges.unshift(edge);
				if (this.edges.length > maxEdges) this.edges.length = maxEdges;
			}
		} else this.carry = 0;

		// The strip: the head (when shown) and then the edges from newest to oldest. The texture
		// runs along the ribbon from the head (u 0) to the end of its life (u 1), in the first
		// cell's column.
		const cols = Math.max(1, e.cols);
		const rows = Math.max(1, e.rows);
		const cellU = 1 / cols;
		const cellV = 1 / rows;
		let n = 0;
		this.strip[0] = life;
		this.strip[1] = cellU;
		this.strip[2] = cellV;
		if (shown) {
			const h = this.head;
			h.top.copy(this.top);
			h.bottom.copy(this.bottom);
			h.age = 0;
			h.r = this.tint.r;
			h.g = this.tint.g;
			h.b = this.tint.b;
			h.a = alpha;
			n = this.put(n, h);
		}
		for (const edge of this.edges) {
			if (n > maxEdges) break;
			n = this.put(n, edge);
		}
		// n edges make n - 1 quads. A ribbon with one edge has nothing to draw.
		const was = this.count;
		this.count = n;
		this.batch?.changed(was < 2 ? 0 : was * 2, n < 2 ? 0 : n * 2);
	}

	// put writes edge n's two points, and gives the next edge's number.
	private put(n: number, edge: Edge): number {
		const d = this.data;
		const life = this.strip[0];
		const u = (Math.min(edge.age, life) / life) * this.strip[1];
		const fade = edge.a * (1 - edge.age / life);
		const o = n * 2 * STRIDE;
		const { top, bottom } = edge;
		d[o] = top.x;
		d[o + 1] = top.y;
		d[o + 2] = top.z;
		d[o + STRIDE] = bottom.x;
		d[o + STRIDE + 1] = bottom.y;
		d[o + STRIDE + 2] = bottom.z;
		d[o + 3] = u;
		d[o + 4] = 0;
		d[o + STRIDE + 3] = u;
		d[o + STRIDE + 4] = this.strip[2];
		d[o + 5] = d[o + STRIDE + 5] = edge.r;
		d[o + 6] = d[o + STRIDE + 6] = edge.g;
		d[o + 7] = d[o + STRIDE + 7] = edge.b;
		d[o + 8] = d[o + STRIDE + 8] = fade;
		return n + 1;
	}

	// reset lets the strip go and starts over, for a model used again (model.ts recycle).
	reset() {
		for (const edge of this.edges) freeEdges.push(edge);
		this.edges.length = 0;
		this.carry = 0;
		const was = this.count;
		this.count = 0;
		this.batch?.changed(was < 2 ? 0 : was * 2, 0);
	}

	dispose() {
		this.disposed = true;
		this.batch?.leave(this);
		this.batch = undefined;
		this.anchor.removeFromParent();
	}
}

// Ribbons runs a model's ribbon emitters. The strips live in `world` (the scene), so a trail
// stays where it was drawn while the model moves on.
export class Ribbons {
	private readonly ribbons: Ribbon[];
	emitting = true;

	constructor(model: Model, world: THREE.Object3D) {
		this.ribbons = model.m2.ribbons.map(e => new Ribbon(model, e, world));
	}

	update(dt: number) {
		for (const r of this.ribbons) r.update(dt, this.emitting);
	}

	reset() {
		this.emitting = true;
		for (const r of this.ribbons) r.reset();
	}

	dispose() {
		for (const r of this.ribbons) r.dispose();
	}
}
