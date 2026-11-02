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

// The most edges one ribbon keeps, whatever its rate and lifetime say.
const maxEdges = 256;

type Edge = { top: THREE.Vector3; bottom: THREE.Vector3; fall: number; age: number; r: number; g: number; b: number; a: number };

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

class Ribbon {
	readonly mesh: THREE.Mesh;
	private readonly anchor = new THREE.Object3D();
	private readonly edges: Edge[] = [];
	private carry = 0;
	private readonly geometry = new THREE.BufferGeometry();
	private readonly pos: THREE.BufferAttribute;
	private readonly uv: THREE.BufferAttribute;
	private readonly color: THREE.BufferAttribute;
	private readonly v = [0, 0, 0];
	private readonly tint = new THREE.Color();
	private readonly top = new THREE.Vector3();
	private readonly bottom = new THREE.Vector3();

	constructor(
		readonly model: Model,
		readonly e: RibbonEmitter,
		world: THREE.Object3D,
	) {
		const bone = model.bones[e.bone] ?? model.bones[0];
		const pivot = model.m2.bones[e.bone]?.pivot ?? [0, 0, 0];
		this.anchor.position.set(e.position[0] - pivot[0], e.position[1] - pivot[1], e.position[2] - pivot[2]);
		bone.add(this.anchor);

		// Two vertices per edge, plus the head (where the ribbon is right now).
		const verts = (maxEdges + 1) * 2;
		this.pos = new THREE.BufferAttribute(new Float32Array(verts * 3), 3);
		this.uv = new THREE.BufferAttribute(new Float32Array(verts * 2), 2);
		this.color = new THREE.BufferAttribute(new Float32Array(verts * 4), 4);
		for (const a of [this.pos, this.uv, this.color]) a.setUsage(THREE.DynamicDrawUsage);
		const index: number[] = [];
		for (let i = 0; i < maxEdges; i++) {
			const a = i * 2;
			index.push(a, a + 1, a + 2, a + 1, a + 3, a + 2);
		}
		this.geometry.setIndex(index);
		this.geometry.setAttribute('position', this.pos);
		this.geometry.setAttribute('uv', this.uv);
		this.geometry.setAttribute('color', this.color);
		this.geometry.setDrawRange(0, 0);

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
		this.mesh = new THREE.Mesh(this.geometry, material);
		this.mesh.frustumCulled = false;
		this.mesh.renderOrder = 10;
		world.add(this.mesh);

		const tex = model.m2.textures[e.textures[0] ?? -1];
		if (tex && tex.type === 0 && tex.fdid > 0) {
			texture(tex.fdid, (tex.flags & 1) !== 0, (tex.flags & 2) !== 0)
				.then(t => {
					material.uniforms.map.value = t;
				})
				.catch(err => console.warn('ribbon texture', err));
		} else {
			// A replaceable ribbon texture, which we don't have: hide the ribbon.
			this.mesh.visible = false;
		}
	}

	private value(track: RibbonEmitter['heightAbove'], fallback: number): number {
		return this.model.trackValue(track, this.v) ? this.v[0] : fallback;
	}

	update(dt: number, emitting: boolean) {
		const e = this.e;
		if (!this.mesh.visible) return;
		this.anchor.updateWorldMatrix(true, false);
		const world = this.anchor.matrixWorld;
		const life = Math.max(0.01, e.edgeLifetime);

		// Age the edges and let them fall.
		for (let i = this.edges.length - 1; i >= 0; i--) {
			const edge = this.edges[i];
			edge.age += dt;
			if (edge.age >= life) {
				this.edges.splice(i, 1);
				continue;
			}
			if (e.gravity !== 0) {
				edge.fall += e.gravity * dt;
				edge.top.y -= edge.fall * dt;
				edge.bottom.y -= edge.fall * dt;
			}
		}

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
				this.edges.unshift({
					top: this.top.clone(),
					bottom: this.bottom.clone(),
					fall: 0,
					age: 0,
					r: this.tint.r,
					g: this.tint.g,
					b: this.tint.b,
					a: alpha,
				});
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
		const put = (top: THREE.Vector3, bottom: THREE.Vector3, age: number, r: number, g: number, b: number, a: number) => {
			const u = (Math.min(age, life) / life) * cellU;
			const fade = a * (1 - age / life);
			this.pos.setXYZ(n * 2, top.x, top.y, top.z);
			this.pos.setXYZ(n * 2 + 1, bottom.x, bottom.y, bottom.z);
			this.uv.setXY(n * 2, u, 0);
			this.uv.setXY(n * 2 + 1, u, cellV);
			this.color.setXYZW(n * 2, r, g, b, fade);
			this.color.setXYZW(n * 2 + 1, r, g, b, fade);
			n++;
		};
		if (shown) put(this.top, this.bottom, 0, this.tint.r, this.tint.g, this.tint.b, alpha);
		for (const edge of this.edges) {
			if (n > maxEdges) break;
			put(edge.top, edge.bottom, edge.age, edge.r, edge.g, edge.b, edge.a);
		}
		// n points make n - 1 quads of two triangles.
		this.geometry.setDrawRange(0, Math.max(0, n - 1) * 6);
		for (const a of [this.pos, this.uv, this.color]) a.needsUpdate = true;
	}

	dispose() {
		this.mesh.removeFromParent();
		this.geometry.dispose();
		(this.mesh.material as THREE.Material).dispose();
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

	dispose() {
		for (const r of this.ribbons) r.dispose();
	}
}
