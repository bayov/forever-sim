// A crowd of units drawn together: every copy of a part that looks the same (the same model, the
// same geosets, the same textures) draws in one draw call, however many units wear it.
//
// In a raid of 40 the renderer spent most of the frame on draw calls: each raider's body came in
// about 11 parts, plus its helm, shoulders, weapons and cape, and the shadows drew them all again.
// The raiders are only a few looks (four races, two sexes, the same gear), so a part like "the orc
// male's chest in this armor" draws once for every orc male in the raid.
//
// Each unit still animates its own skeleton. We put every bone's skinning matrix (in its model's
// space) into one big texture, and each copy of a part reads its unit's bones from there by an
// offset, with its model's place in the world on top. A copy also carries its colors and
// transparency, for the parts that animate them. Copies off screen are left out, unless they
// cast a shadow onto what's on screen.

import * as THREE from 'three';
import { skins, type Model } from './model';

// The bone texture is this many texels wide: four texels to a bone, so 256 bones a row.
const WIDTH = 1024;
const BONES_PER_ROW = WIDTH / 4;

// A skeleton's bones in the texture: where they start, and where they come from.
type Skeleton = {
	offset: number;
	// How many of the crowd's meshes use it (a body and its cape).
	users: number;
	skeleton: THREE.Skeleton;
	// The bones' skinning matrices, which their model poses (model.ts).
	skinning: Float32Array;
	frame: number;
};

// One kind of part, drawn for every copy.
type Batch = {
	mesh: THREE.Mesh;
	geometry: THREE.InstancedBufferGeometry;
	material: THREE.Material & { color?: THREE.Color; opacity: number };
	bones: THREE.InstancedBufferAttribute;
	colors: THREE.InstancedBufferAttribute;
	places: THREE.InstancedBufferAttribute;
	count: number;
	// How many of the crowd's meshes have a part in it.
	users: number;
	castShadow: boolean;
};

// One model's mesh in the crowd, with the batch each of its parts draws in.
type Entry = {
	model: Model;
	mesh: THREE.SkinnedMesh;
	skeleton: Skeleton;
	parts: { batch: Batch; source: THREE.Material & { color?: THREE.Color; opacity: number } }[];
};

const ids = new WeakMap<object, number>();
let nextId = 1;
function idOf(o: object): number {
	let id = ids.get(o);
	if (!id) ids.set(o, (id = nextId++));
	return id;
}

const sphere = new THREE.Sphere();
const frustum = new THREE.Frustum();
const projection = new THREE.Matrix4();

export class Crowd {
	readonly group = new THREE.Group();
	private entries: Entry[] = [];
	private byModel = new Map<Model, Entry>();
	// The free places in the bone texture, as [first bone, count], in order.
	private free: [number, number][] = [];
	private skeletons = new Map<THREE.Skeleton, Skeleton>();
	private batches = new Map<string, Batch>();
	private bones = 0;
	private data = new Float32Array(WIDTH * 8 * 4);
	private texture = this.makeTexture(8);
	private readonly uniform = { value: this.texture as THREE.Texture };
	private frame = 0;

	constructor(
		readonly scene: THREE.Scene,
		// The sun, whose shadow camera tells which units off screen cast a shadow onto the screen.
		readonly sun?: THREE.DirectionalLight,
		readonly shadows?: () => boolean,
	) {
		this.group.matrixAutoUpdate = false;
		scene.add(this.group);
	}

	// add puts models in the crowd: a unit's body, gear and weapons, or a spell's effect. Their own
	// meshes stop drawing, but they stay in the scene, so what hangs off their bones (weapons, the
	// hand glows of a cast) still follows them. A model without bounds (an effect, model.ts cull)
	// always draws. A disposed model leaves the crowd.
	add(models: Model[]) {
		for (const model of models) {
			const mesh = model.mesh;
			const geometry = mesh.geometry;
			const position = geometry.getAttribute('position');
			if (!geometry.index || !position || this.byModel.has(model)) continue;
			const skeleton = this.skeletonOf(mesh.skeleton);
			const mats = (Array.isArray(mesh.material) ? mesh.material : [mesh.material]) as Entry['parts'][number]['source'][];
			const parts: Entry['parts'] = [];
			for (const g of geometry.groups) {
				const mi = g.materialIndex ?? 0;
				const source = mats[mi];
				if (!source?.visible) continue;
				const map = (source as THREE.MeshLambertMaterial).map;
				const key = [idOf(position), idOf(geometry.index), g.start, g.count, mi, map ? idOf(map) : 0].join();
				let batch = this.batches.get(key);
				if (!batch) {
					const b = model.skin.batches[mi];
					const solid = (model.m2.materials[b?.material ?? -1]?.blend ?? 0) <= 1;
					batch = this.makeBatch(geometry, g.start, g.count, source, solid);
					this.batches.set(key, batch);
				}
				parts.push({ batch, source });
			}
				// The model's own mesh draws no more, but what hangs off its bones (its sockets) still
			// follows them. A hidden mesh would hide those too, so it goes on a layer no camera
			// looks at.
			mesh.layers.set(31);
			const entry = { model, mesh, skeleton, parts };
			this.entries.push(entry);
			this.byModel.set(model, entry);
			model.crowd = this;
			for (const p of parts) this.grow(p.batch);
		}
	}

	// warm puts models in the crowd and shows every part, so the renderer can compile the crowd's
	// shaders for them while the room loads. It gives what takes them out again.
	warm(models: Model[]): () => void {
		const fresh = models.filter(m => !this.byModel.has(m));
		this.add(fresh);
		for (const b of this.batches.values()) b.mesh.visible = true;
		return () => {
			for (const m of fresh) this.remove(m);
		};
	}

	// remove takes a model out of the crowd, and it draws on its own again.
	remove(model: Model) {
		const entry = this.byModel.get(model);
		if (!entry) return;
		this.byModel.delete(model);
		model.crowd = undefined;
		const i = this.entries.indexOf(entry);
		this.entries[i] = this.entries[this.entries.length - 1];
		this.entries.pop();
		for (const p of entry.parts) p.batch.users--;
		entry.mesh.layers.set(0);
		const k = entry.skeleton;
		if (--k.users > 0) return;
		this.skeletons.delete(k.skeleton);
		this.release(k.offset, k.skeleton.bones.length);
	}

	// release gives a skeleton's bones back, merging the free place with its neighbors.
	private release(offset: number, count: number) {
		const free = this.free;
		let i = 0;
		while (i < free.length && free[i][0] < offset) i++;
		free.splice(i, 0, [offset, count]);
		if (i + 1 < free.length && free[i][0] + free[i][1] === free[i + 1][0]) {
			free[i][1] += free[i + 1][1];
			free.splice(i + 1, 1);
		}
		if (i > 0 && free[i - 1][0] + free[i - 1][1] === free[i][0]) {
			free[i - 1][1] += free[i][1];
			free.splice(i, 1);
		}
	}

	// allocate finds a place for count bones: the first free place they fit in, or the end.
	private allocate(count: number): number {
		const free = this.free;
		for (let i = 0; i < free.length; i++) {
			const f = free[i];
			if (f[1] < count) continue;
			const offset = f[0];
			f[0] += count;
			f[1] -= count;
			if (f[1] === 0) free.splice(i, 1);
			return offset;
		}
		const offset = this.bones;
		this.bones += count;
		return offset;
	}

	// skeletonOf gives a skeleton its place in the bone texture. A cape shares the body's.
	private skeletonOf(s: THREE.Skeleton): Skeleton {
		let k = this.skeletons.get(s);
		if (k) {
			k.users++;
			return k;
		}
		const n = s.bones.length;
		const skinning = skins.get(s) ?? new Float32Array(n * 16);
		k = { offset: this.allocate(n), users: 1, skeleton: s, skinning, frame: -1 };
		this.skeletons.set(s, k);
		const rows = Math.ceil(this.bones / BONES_PER_ROW);
		if (rows > this.texture.image.height) {
			// The texture grows by 8 rows (2048 bones) at a time, since all of it goes to the GPU
			// each frame.
			const height = Math.ceil(rows / 8) * 8;
			const data = new Float32Array(WIDTH * height * 4);
			data.set(this.data);
			this.data = data;
			this.texture.dispose();
			this.texture = this.makeTexture(height);
			this.uniform.value = this.texture;
		}
		return k;
	}

	private makeTexture(height: number): THREE.DataTexture {
		const t = new THREE.DataTexture(this.data, WIDTH, height, THREE.RGBAFormat, THREE.FloatType);
		t.minFilter = t.magFilter = THREE.NearestFilter;
		t.generateMipmaps = false;
		t.needsUpdate = true;
		return t;
	}

	private makeBatch(source: THREE.BufferGeometry, start: number, count: number, sourceMaterial: Batch['material'], solid: boolean): Batch {
		const geometry = new THREE.InstancedBufferGeometry();
		for (const [name, a] of Object.entries(source.attributes)) geometry.setAttribute(name, a);
		geometry.setIndex(source.index);
		geometry.setDrawRange(start, count);
		const bones = new THREE.InstancedBufferAttribute(new Float32Array(16), 1);
		const colors = new THREE.InstancedBufferAttribute(new Float32Array(16 * 4), 4);
		const places = new THREE.InstancedBufferAttribute(new Float32Array(16 * 16), 16);
		for (const a of [bones, colors, places]) a.setUsage(THREE.DynamicDrawUsage);
		geometry.setAttribute('crowdBone', bones);
		geometry.setAttribute('crowdColor', colors);
		geometry.setAttribute('crowdPlace', places);
		geometry.instanceCount = 0;

		const material = sourceMaterial.clone() as Batch['material'];
		material.color?.setRGB(1, 1, 1);
		material.opacity = 1;
		const before = sourceMaterial.onBeforeCompile;
		const key = sourceMaterial.customProgramCacheKey !== THREE.Material.prototype.customProgramCacheKey ? sourceMaterial.customProgramCacheKey() : '';
		material.onBeforeCompile = (shader, renderer) => {
			before.call(sourceMaterial, shader, renderer);
			shader.uniforms.crowdBones = this.uniform;
			shader.vertexShader = crowdVertex(shader.vertexShader);
			shader.fragmentShader = shader.fragmentShader
				.replace('#include <common>', '#include <common>\nvarying vec4 vCrowdColor;')
				.replace('#include <color_fragment>', '#include <color_fragment>\ndiffuseColor *= vCrowdColor;');
		};
		material.customProgramCacheKey = () => `crowd-${key}`;

		const depth = new THREE.MeshDepthMaterial();
		depth.onBeforeCompile = shader => {
			shader.uniforms.crowdBones = this.uniform;
			shader.vertexShader = crowdVertex(shader.vertexShader);
		};
		depth.customProgramCacheKey = () => 'crowd-depth';

		const mesh = new THREE.Mesh(geometry, material);
		mesh.frustumCulled = false;
		mesh.matrixAutoUpdate = false;
		mesh.castShadow = solid;
		mesh.customDepthMaterial = depth;
		mesh.visible = false;
		this.group.add(mesh);
		return { mesh, geometry, material, bones, colors, places, count: 0, users: 0, castShadow: solid };
	}

	// grow makes room in a batch for one more copy.
	private grow(b: Batch) {
		b.users++;
		if (b.users <= b.bones.count) return;
		let n = b.bones.count;
		while (n < b.users) n *= 2;
		b.bones = new THREE.InstancedBufferAttribute(new Float32Array(n), 1).setUsage(THREE.DynamicDrawUsage);
		b.colors = new THREE.InstancedBufferAttribute(new Float32Array(n * 4), 4).setUsage(THREE.DynamicDrawUsage);
		b.places = new THREE.InstancedBufferAttribute(new Float32Array(n * 16), 16).setUsage(THREE.DynamicDrawUsage);
		b.geometry.setAttribute('crowdBone', b.bones);
		b.geometry.setAttribute('crowdColor', b.colors);
		b.geometry.setAttribute('crowdPlace', b.places);
		// Three keeps the most copies a geometry's attributes hold from its first draw, and draws no
		// more than that.
		(b.geometry as { _maxInstanceCount?: number })._maxInstanceCount = undefined;
	}

	// prepare writes the bones and the copies to draw, once the scene's objects have their places
	// for the frame (the scene's onBeforeRender), before the shadows draw.
	prepare(camera: THREE.Camera) {
		if (this.entries.length === 0) return;
		this.frame++;
		frustum.setFromProjectionMatrix(projection.multiplyMatrices(camera.projectionMatrix, camera.matrixWorldInverse));
		const sun = this.sun;
		const shadows = !!sun && sun.castShadow && (this.shadows?.() ?? true);
		let shadowFrustum: THREE.Frustum | undefined;
		if (shadows && sun) {
			sun.shadow.updateMatrices(sun);
			shadowFrustum = sun.shadow.getFrustum();
		}
		for (const b of this.batches.values()) b.count = 0;
		for (const e of this.entries) {
			if (!shown(e.mesh)) continue;
			let inView = true;
			let inShadow = !!shadowFrustum;
			if (e.mesh.boundingSphere) {
				sphere.copy(e.mesh.boundingSphere).applyMatrix4(e.mesh.matrixWorld);
				inView = frustum.intersectsSphere(sphere);
				inShadow = !!shadowFrustum && shadowFrustum.intersectsSphere(sphere);
			}
			if (!inView && !inShadow) continue;
			const k = e.skeleton;
			if (k.frame !== this.frame) {
				k.frame = this.frame;
				this.writeBones(k);
			}
			for (const { batch: b, source } of e.parts) {
				if (!inView && !b.castShadow) continue;
				const i = b.count++;
				b.bones.array[i] = k.offset;
				(b.places.array as Float32Array).set(e.mesh.matrixWorld.elements, i * 16);
				const c = b.colors.array;
				const color = source.color;
				c[i * 4] = color ? color.r : 1;
				c[i * 4 + 1] = color ? color.g : 1;
				c[i * 4 + 2] = color ? color.b : 1;
				c[i * 4 + 3] = source.opacity;
				// A part that fades turns see-through, and stays so (model.ts updateColors).
				if (source.transparent && !b.material.transparent) {
					b.material.transparent = true;
					b.material.needsUpdate = true;
				}
			}
		}
		for (const b of this.batches.values()) {
			b.geometry.instanceCount = b.count;
			b.mesh.visible = b.count > 0;
			if (b.count === 0) continue;
			b.bones.clearUpdateRanges();
			b.bones.addUpdateRange(0, b.count);
			b.bones.needsUpdate = true;
			b.colors.clearUpdateRanges();
			b.colors.addUpdateRange(0, b.count * 4);
			b.colors.needsUpdate = true;
			b.places.clearUpdateRanges();
			b.places.addUpdateRange(0, b.count * 16);
			b.places.needsUpdate = true;
		}
		this.texture.needsUpdate = true;
	}

	// writeBones puts a skeleton's skinning matrices in the texture.
	private writeBones(k: Skeleton) {
		this.data.set(k.skinning, k.offset * 16);
	}

	dispose() {
		this.group.removeFromParent();
		for (const b of this.batches.values()) {
			b.geometry.dispose();
			b.material.dispose();
			(b.mesh.customDepthMaterial as THREE.Material).dispose();
		}
		this.texture.dispose();
		this.entries = [];
		this.batches.clear();
	}
}

// shown is whether a mesh would draw: it and everything it hangs from are visible, up to a scene.
function shown(o: THREE.Object3D): boolean {
	for (let x: THREE.Object3D | null = o; x; x = x.parent) {
		if (!x.visible) return false;
		if ((x as THREE.Scene).isScene) return true;
	}
	return false;
}

// crowdVertex skins a copy by its unit's bones in the bone texture, in place of Three's skinning
// (the crowd's meshes aren't SkinnedMeshes, so Three's skinning code is left out of their shader).
function crowdVertex(src: string): string {
	return src
		.replace(
			'#include <skinning_pars_vertex>',
			/* glsl */ `attribute vec4 skinIndex;
attribute vec4 skinWeight;
attribute float crowdBone;
attribute vec4 crowdColor;
attribute mat4 crowdPlace;
uniform highp sampler2D crowdBones;
varying vec4 vCrowdColor;
mat4 crowdBoneMatrix(float i) {
	int j = int(i) * 4;
	int x = j % ${WIDTH};
	int y = j / ${WIDTH};
	return mat4(texelFetch(crowdBones, ivec2(x, y), 0), texelFetch(crowdBones, ivec2(x + 1, y), 0), texelFetch(crowdBones, ivec2(x + 2, y), 0), texelFetch(crowdBones, ivec2(x + 3, y), 0));
}`,
		)
		.replace(
			'void main() {',
			/* glsl */ `void main() {
	vCrowdColor = crowdColor;
	mat4 crowdSkin = crowdPlace * (skinWeight.x * crowdBoneMatrix(crowdBone + skinIndex.x) + skinWeight.y * crowdBoneMatrix(crowdBone + skinIndex.y)
		+ skinWeight.z * crowdBoneMatrix(crowdBone + skinIndex.z) + skinWeight.w * crowdBoneMatrix(crowdBone + skinIndex.w));
	vec3 crowdNormal = (crowdSkin * vec4(normal, 0.0)).xyz;`,
		)
		.replace('#include <skinbase_vertex>', '')
		.replace('#include <skinnormal_vertex>', 'objectNormal = (crowdSkin * vec4(objectNormal, 0.0)).xyz;')
		.replace('#include <skinning_vertex>', 'transformed = (crowdSkin * vec4(transformed, 1.0)).xyz;')
		.replace('vWowNormal = normalize(normalMatrix * normal);', 'vWowNormal = normalize(normalMatrix * crowdNormal);');
}
