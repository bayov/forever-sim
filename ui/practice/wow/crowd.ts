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
// transparency, for the parts that animate them.
//
// A character's skin is a picture we paint for its armor (compose.ts), so raiders in different
// gear have different skins. The skins of one size go into one texture array, and each copy of a
// body part reads its layer. So the bodies of one race and sex still draw in one call, whatever
// they wear. Copies off screen are left out, unless they
// cast a shadow onto what's on screen.

import * as THREE from 'three';
import { forget } from './compose';
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

// The painted skins of one size, as layers of one texture array.
// uploaded says the GPU has the array, so a new layer goes up on its own.
type SkinArray = { width: number; height: number; layers: number; capacity: number; data: Uint8Array; texture: THREE.DataArrayTexture; uniform: { value: THREE.DataArrayTexture }; uploaded: boolean };

// The masks have room for this many submeshes (a character model has up to about 150), 24 to a
// float.
const MASK_BITS = 192;

function bitSet(mask: Float32Array, bit: number): boolean {
	return Math.floor(mask[Math.floor(bit / 24)] / 2 ** (bit % 24)) % 2 === 1;
}

// copyLayer copies a painted skin into its layer, at half size when the array's layers are.
function copyLayer(t: THREE.Texture, a: SkinArray, layer: number) {
	const image = t.image as { width: number; height: number; data: Uint8ClampedArray };
	const out = a.data;
	const o = layer * a.width * a.height * 4;
	if (image.width === a.width) {
		out.set(image.data, o);
		return;
	}
	// Each pixel is the average of the 2 by 2 it stands for.
	const src = image.data;
	const w = image.width;
	for (let y = 0; y < a.height; y++) {
		for (let x = 0; x < a.width; x++) {
			const s = (y * 2 * w + x * 2) * 4;
			const d = o + (y * a.width + x) * 4;
			for (let c = 0; c < 4; c++) out[d + c] = (src[s + c] + src[s + 4 + c] + src[s + w * 4 + c] + src[s + w * 4 + 4 + c] + 2) >> 2;
		}
	}
}

// A body is a race and sex's character model, shared by the raiders of that race and sex in any
// gear. Its batches hold every part any of them shows (shown), merged by how they look, and each
// raider's mask hides the parts it doesn't show. So the bodies of one race and sex draw in a few
// calls, however their gear differs.
type Body = {
	full: THREE.BufferGeometry;
	// Each vertex's submesh.
	sub: THREE.BufferAttribute;
	shown: Set<number>;
	// The skin batches in each batch, by its key.
	parts: Map<string, Set<number>>;
	batches: Map<string, Batch>;
};

// One kind of part, drawn for every copy.
type Batch = {
	mesh: THREE.Mesh;
	geometry: THREE.InstancedBufferGeometry;
	material: THREE.Material & { color?: THREE.Color; opacity: number };
	bones: THREE.InstancedBufferAttribute;
	colors: THREE.InstancedBufferAttribute;
	places: THREE.InstancedBufferAttribute;
	layers: THREE.InstancedBufferAttribute;
	// The copies' masks, for a body's batches (see Body).
	masks?: THREE.InstancedBufferAttribute[];
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
	// layer is the part's skin in its batch's texture array, or 0.
	parts: { batch: Batch; source: THREE.Material & { color?: THREE.Color; opacity: number }; layer: number }[];
	// The submeshes a body shows (see Body), as 24 bits a float.
	mask?: Float32Array;
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
	// The texture arrays of painted skins, by size, and each skin's layer.
	private skinArrays = new Map<string, SkinArray>();
	private skinLayers = new Map<THREE.Texture, number>();
	private skinUsers = new Map<THREE.Texture, Set<Model>>();
	// The bodies shared by race and sex, by their vertices.
	private bodies = new Map<number, Body>();
	// The free places in the bone texture, as [first bone, count], in order.
	private free: [number, number][] = [];
	private skeletons = new Map<THREE.Skeleton, Skeleton>();
	private batches = new Map<string, Batch>();
	private bones = 0;
	private data = new Float32Array(WIDTH * 8 * 4);
	private texture = this.makeTexture(8);
	private readonly uniform = { value: this.texture as THREE.Texture };
	private frame = 0;

	// useSkinArrays puts painted skins in texture arrays. Off, each skin draws on its own (to
	// measure what the arrays save, with ?skinarrays=0).
	useSkinArrays = new URLSearchParams(location.search).get('skinarrays') !== '0';

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
			const body = this.bodyParts(model, mats);
			const parts: Entry['parts'] = body?.parts ?? [];
			for (const g of body ? [] : geometry.groups) {
				const mi = g.materialIndex ?? 0;
				const source = mats[mi];
				if (!source?.visible) continue;
				const map = (source as THREE.MeshLambertMaterial).map;
				const skins = map ? this.skinArray(map) : undefined;
				const key = [idOf(position), idOf(geometry.index), g.start, g.count, mi, skins ? `skins${skins.width}x${skins.height}` : map ? idOf(map) : 0].join();
				let batch = this.batches.get(key);
				if (!batch) {
					const b = model.skin.batches[mi];
					const solid = (model.m2.materials[b?.material ?? -1]?.blend ?? 0) <= 1;
					batch = this.makeBatch(geometry, g.start, g.count, source, solid, skins);
					this.batches.set(key, batch);
				}
				parts.push({ batch, source, layer: skins ? this.layerOf(skins, map!, model) : 0 });
			}
			// The model's own mesh draws no more, but what hangs off its bones (its sockets) still
			// follows them. A hidden mesh would hide those too, so it goes on a layer no camera
			// looks at.
			mesh.layers.set(31);
			const entry: Entry = { model, mesh, skeleton, parts, mask: body?.mask };
			this.entries.push(entry);
			this.byModel.set(model, entry);
			model.crowd = this;
			for (const p of parts) this.grow(p.batch);
		}
	}

	// bodyParts puts a character's body in its race and sex's shared batches (see Body), with the
	// mask of its parts that show. It's undefined for a model that isn't a painted body.
	private bodyParts(model: Model, mats: Entry['parts'][number]['source'][]): { parts: Entry['parts']; mask: Float32Array } | undefined {
		if (!this.useSkinArrays) return undefined;
		const subs = model.skin.submeshes;
		if (subs.length > MASK_BITS || !mats.some(m => (m as THREE.MeshLambertMaterial).map?.userData.composed)) return undefined;
		const full = model.fullGeometry;
		const key = idOf(full.getAttribute('position'));
		let body = this.bodies.get(key);
		if (!body) {
			// Each vertex's submesh, for the mask.
			const n = full.getAttribute('position').count;
			const sub = new Float32Array(n);
			subs.forEach((m, i) => sub.fill(i, m.vertexStart, Math.min(n, m.vertexStart + m.vertexCount)));
			body = { full, sub: new THREE.BufferAttribute(sub, 1), shown: new Set(), parts: new Map(), batches: new Map() };
			this.bodies.set(key, body);
		}
		// The skin batches that show, each with the texture it draws with.
		const shown: number[] = [];
		const skinsOf = new Map<number, SkinArray>();
		for (let i = 0; i < mats.length; i++) {
			const m = mats[i];
			if (!m?.visible || !model.skin.batches[i]) continue;
			const map = (m as THREE.MeshLambertMaterial).map ?? null;
			const skins = map ? this.skinArray(map) : undefined;
			if (map?.userData.composed && !skins) return undefined;
			if (skins) skinsOf.set(i, skins);
			shown.push(i);
		}
		let grew = false;
		for (const i of shown) {
			if (body.shown.has(i)) continue;
			body.shown.add(i);
			grew = true;
		}
		// Each part's batch key: solid parts that look alike share one, and each see-through part has
		// its own (they draw in their order). A part with another texture than the same part on
		// another raider (a hair color) draws in a batch of its own.
		const keyOf = (i: number) => {
			const b = model.skin.batches[i];
			const skins = skinsOf.get(i);
			const map = (mats[i] as THREE.MeshLambertMaterial).map;
			const look = skins ? `skins${skins.width}x${skins.height}` : map ? idOf(map) : 0;
			if (mats[i].transparent) return `t${i}:${look}`;
			return [b.material, b.textureCombo, b.textureCount, b.shader, b.color, b.weightCombo, b.transformCombo, b.coordCombo, b.priority, b.materialLayer, look].join();
		};
		const keys = new Map<number, string>();
		for (const i of shown) {
			const k = keyOf(i);
			keys.set(i, k);
			const list = body.parts.get(k) ?? new Set<number>();
			body.parts.set(k, list);
			if (!list.has(i)) {
				list.add(i);
				grew = true;
			}
			if (!body.batches.has(k)) {
				const b = model.skin.batches[i];
				const solid = (model.m2.materials[b.material]?.blend ?? 0) <= 1;
				const batch = this.makeBatch(full, 0, 0, mats[i], solid, skinsOf.get(i), body.sub);
				body.batches.set(k, batch);
				this.batches.set(`body${idOf(full)}:${k}`, batch);
				grew = true;
			}
		}
		if (grew) this.rebuildBody(body, model);
		const mask = new Float32Array(MASK_BITS / 24);
		const parts: Entry['parts'] = [];
		const done = new Set<string>();
		for (const i of shown) {
			const sub = model.skin.batches[i].submesh;
			mask[Math.floor(sub / 24)] += (2 ** (sub % 24)) * (bitSet(mask, sub) ? 0 : 1);
			const k = keys.get(i)!;
			if (done.has(k)) continue;
			done.add(k);
			const map = (mats[i] as THREE.MeshLambertMaterial).map;
			const skins = skinsOf.get(i);
			parts.push({ batch: body.batches.get(k)!, source: mats[i], layer: skins && map ? this.layerOf(skins, map, model) : 0 });
		}
		return { parts, mask };
	}

	// rebuildBody makes the index of every part a body's raiders show, with each batch's range in
	// it, in the order the parts draw (model.ts sorts them by layer).
	private rebuildBody(body: Body, model: Model) {
		const full = body.full;
		const src = full.index!.array as Uint16Array | Uint32Array;
		const byKey = new Map<string, [number, number][]>();
		const order = [...full.groups].sort((a, b) => {
			const ba = model.skin.batches[a.materialIndex!];
			const bb = model.skin.batches[b.materialIndex!];
			return ba.priority - bb.priority || ba.materialLayer - bb.materialLayer;
		});
		for (const g of order) {
			const i = g.materialIndex ?? 0;
			for (const [k, members] of body.parts) {
				if (!members.has(i)) continue;
				const list = byKey.get(k) ?? [];
				byKey.set(k, list);
				list.push([g.start, g.count]);
			}
		}
		let total = 0;
		for (const list of byKey.values()) for (const [, c] of list) total += c;
		const out = src instanceof Uint32Array ? new Uint32Array(total) : new Uint16Array(total);
		let at = 0;
		const index = new THREE.BufferAttribute(out, 1);
		for (const [k, list] of byKey) {
			const start = at;
			for (const [s, c] of list) {
				out.set(src.subarray(s, s + c), at);
				at += c;
			}
			const batch = body.batches.get(k)!;
			batch.geometry.setIndex(index);
			batch.geometry.setDrawRange(start, at - start);
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

	// skinArray gives the texture array for a painted skin's size, or undefined for a texture that
	// isn't a painted skin (or whose picture already went to the GPU, so we can't copy it).
	private skinArray(t: THREE.Texture): SkinArray | undefined {
		if (!t.userData.composed || !this.useSkinArrays) return undefined;
		const image = t.image as { width: number; height: number; data: Uint8ClampedArray | null };
		if (!image.data && !this.skinLayers.has(t)) return undefined;
		// Skins over 1024 wide go in at half size, so a raid's take a quarter of the memory. A
		// raider is at most a few hundred pixels tall on screen.
		const half = image.width > 1024 ? 2 : 1;
		const key = `${image.width / half}x${image.height / half}`;
		let a = this.skinArrays.get(key);
		if (!a) {
			a = { width: image.width / half, height: image.height / half, layers: 0, capacity: 0, data: new Uint8Array(0), texture: undefined!, uniform: { value: null! }, uploaded: false };
			this.growSkins(a, 8);
			this.skinArrays.set(key, a);
		}
		return a;
	}

	// layerOf gives a painted skin its layer in the array, copying its picture in the first time.
	// Once every unit given the skin is in the crowd, we let go of its own picture (the GPU never
	// gets it), since a raid's skins were over 300MB.
	private layerOf(a: SkinArray, t: THREE.Texture, model?: Model): number {
		if (model) {
			const users = this.skinUsers.get(t) ?? new Set<Model>();
			this.skinUsers.set(t, users);
			users.add(model);
			if (users.size >= (t.userData.refs ?? 1) && (t.image as { data: unknown }).data) {
				queueMicrotask(() => {
					if ((this.skinUsers.get(t)?.size ?? 0) < (t.userData.refs ?? 1)) return;
					(t.image as { data: unknown }).data = null;
					forget(t);
				});
			}
		}
		const known = this.skinLayers.get(t);
		if (known !== undefined) return known;
		if (a.layers === a.capacity) this.growSkins(a, a.capacity * 2);
		const layer = a.layers++;
		copyLayer(t, a, layer);
		this.skinLayers.set(t, layer);
		// Before the array's first trip to the GPU, all of it goes. After, only the new layer.
		if (a.uploaded) a.texture.addLayerUpdate(layer);
		a.texture.needsUpdate = true;
		return layer;
	}

	// growSkins makes room for more layers. The new array goes to the GPU whole.
	private growSkins(a: SkinArray, capacity: number) {
		const data = new Uint8Array(a.width * a.height * 4 * capacity);
		data.set(a.data);
		a.data = data;
		a.capacity = capacity;
		a.texture?.dispose();
		const t = new THREE.DataArrayTexture(data, a.width, a.height, capacity);
		t.colorSpace = THREE.SRGBColorSpace;
		t.wrapS = t.wrapT = THREE.ClampToEdgeWrapping;
		t.magFilter = THREE.LinearFilter;
		t.minFilter = THREE.LinearMipmapLinearFilter;
		t.generateMipmaps = true;
		t.anisotropy = 4;
		t.needsUpdate = true;
		a.uploaded = false;
		t.onUpdate = () => (a.uploaded = true);
		a.texture = t;
		a.uniform.value = t;
	}

	private makeTexture(height: number): THREE.DataTexture {
		const t = new THREE.DataTexture(this.data, WIDTH, height, THREE.RGBAFormat, THREE.FloatType);
		t.minFilter = t.magFilter = THREE.NearestFilter;
		t.generateMipmaps = false;
		t.needsUpdate = true;
		return t;
	}

	private makeBatch(source: THREE.BufferGeometry, start: number, count: number, sourceMaterial: Batch['material'], solid: boolean, skins?: SkinArray, sub?: THREE.BufferAttribute): Batch {
		const geometry = new THREE.InstancedBufferGeometry();
		for (const [name, a] of Object.entries(source.attributes)) geometry.setAttribute(name, a);
		geometry.setIndex(source.index);
		geometry.setDrawRange(start, count);
		const bones = new THREE.InstancedBufferAttribute(new Float32Array(16), 1);
		const colors = new THREE.InstancedBufferAttribute(new Float32Array(16 * 4), 4);
		const places = new THREE.InstancedBufferAttribute(new Float32Array(16 * 16), 16);
		const layers = new THREE.InstancedBufferAttribute(new Float32Array(16), 1);
		for (const a of [bones, colors, places, layers]) a.setUsage(THREE.DynamicDrawUsage);
		geometry.setAttribute('crowdBone', bones);
		geometry.setAttribute('crowdColor', colors);
		geometry.setAttribute('crowdPlace', places);
		geometry.setAttribute('crowdLayer', layers);
		const masks = [new THREE.InstancedBufferAttribute(new Float32Array(16 * 4), 4), new THREE.InstancedBufferAttribute(new Float32Array(16 * 4), 4)];
		if (sub) {
			geometry.setAttribute('crowdSub', sub);
			geometry.setAttribute('crowdMaskA', masks[0].setUsage(THREE.DynamicDrawUsage));
			geometry.setAttribute('crowdMaskB', masks[1].setUsage(THREE.DynamicDrawUsage));
		}
		geometry.instanceCount = 0;

		const material = sourceMaterial.clone() as Batch['material'];
		material.color?.setRGB(1, 1, 1);
		material.opacity = 1;
		const before = sourceMaterial.onBeforeCompile;
		const key = sourceMaterial.customProgramCacheKey !== THREE.Material.prototype.customProgramCacheKey ? sourceMaterial.customProgramCacheKey() : '';
		material.onBeforeCompile = (shader, renderer) => {
			before.call(sourceMaterial, shader, renderer);
			shader.uniforms.crowdBones = this.uniform;
			shader.vertexShader = crowdVertex(shader.vertexShader, !!sub);
			shader.fragmentShader = shader.fragmentShader
				.replace('#include <common>', '#include <common>\nvarying vec4 vCrowdColor;\nvarying float vCrowdLayer;')
				.replace('#include <color_fragment>', '#include <color_fragment>\ndiffuseColor *= vCrowdColor;');
			if (skins) {
				// The skin comes from the copy's layer of the array, in place of the material's map.
				shader.uniforms.crowdSkins = skins.uniform;
				shader.fragmentShader = shader.fragmentShader
					.replace('#include <common>', '#include <common>\nuniform highp sampler2DArray crowdSkins;')
					.replace('#include <map_fragment>', '#ifdef USE_MAP\ndiffuseColor *= texture(crowdSkins, vec3(vMapUv, vCrowdLayer));\n#endif')
					.replace('texture2D(map, wowEnv0 ? wowEnvUv : vMapUv)', 'texture(crowdSkins, vec3(wowEnv0 ? wowEnvUv : vMapUv, vCrowdLayer))');
			}
		};
		material.customProgramCacheKey = () => `crowd-${key}${skins ? '-skins' : ''}${sub ? '-mask' : ''}`;

		const depth = new THREE.MeshDepthMaterial();
		depth.onBeforeCompile = shader => {
			shader.uniforms.crowdBones = this.uniform;
			shader.vertexShader = crowdVertex(shader.vertexShader, !!sub);
		};
		depth.customProgramCacheKey = () => `crowd-depth${sub ? '-mask' : ''}`;

		const mesh = new THREE.Mesh(geometry, material);
		mesh.frustumCulled = false;
		mesh.matrixAutoUpdate = false;
		mesh.castShadow = solid;
		mesh.customDepthMaterial = depth;
		mesh.visible = false;
		this.group.add(mesh);
		return { mesh, geometry, material, bones, colors, places, layers, masks: sub ? masks : undefined, count: 0, users: 0, castShadow: solid };
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
		b.layers = new THREE.InstancedBufferAttribute(new Float32Array(n), 1).setUsage(THREE.DynamicDrawUsage);
		b.geometry.setAttribute('crowdLayer', b.layers);
		if (b.masks) {
			b.masks = b.masks.map(() => new THREE.InstancedBufferAttribute(new Float32Array(n * 4), 4).setUsage(THREE.DynamicDrawUsage));
			b.geometry.setAttribute('crowdMaskA', b.masks[0]);
			b.geometry.setAttribute('crowdMaskB', b.masks[1]);
		}
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
			for (const { batch: b, source, layer } of e.parts) {
				if (!inView && !b.castShadow) continue;
				const i = b.count++;
				b.bones.array[i] = k.offset;
				b.layers.array[i] = layer;
				if (b.masks && e.mask) {
					const ma = b.masks[0].array;
					const mb = b.masks[1].array;
					for (let j = 0; j < 4; j++) {
						ma[i * 4 + j] = e.mask[j];
						mb[i * 4 + j] = e.mask[4 + j];
					}
				}
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
			// The copies' data goes up whole: it's a few kilobytes a batch, and asking for only the
			// part in use made five little objects a batch, every frame.
			b.bones.needsUpdate = true;
			b.colors.needsUpdate = true;
			b.places.needsUpdate = true;
			b.layers.needsUpdate = true;
			if (b.masks) b.masks[0].needsUpdate = b.masks[1].needsUpdate = true;
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
//
// A body's copies (masked) also hide the submeshes their raider doesn't show: those vertices go
// in front of the camera's near plane, so their triangles are cut away, and skip the skinning.
function crowdVertex(src: string, masked = false): string {
	const mask = masked
		? /* glsl */ `
attribute float crowdSub;
attribute vec4 crowdMaskA;
attribute vec4 crowdMaskB;
bool crowdShows() {
	float word = floor(crowdSub / 24.0);
	float bit = crowdSub - word * 24.0;
	vec4 m = word < 4.0 ? crowdMaskA : crowdMaskB;
	float w = mod(word, 4.0);
	float bits = w < 1.0 ? m.x : w < 2.0 ? m.y : w < 3.0 ? m.z : m.w;
	return mod(floor(bits / exp2(bit)), 2.0) > 0.5;
}`
		: '';
	return src
		.replace(
			'#include <skinning_pars_vertex>',
			/* glsl */ `attribute vec4 skinIndex;
attribute vec4 skinWeight;
attribute float crowdBone;
attribute vec4 crowdColor;
attribute mat4 crowdPlace;
attribute float crowdLayer;
varying float vCrowdLayer;
uniform highp sampler2D crowdBones;${mask}
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
			/* glsl */ `void main() {${masked ? '\n\tif (!crowdShows()) {\n\t\tgl_Position = vec4(0.0, 0.0, -2.0, 1.0);\n\t\treturn;\n\t}' : ''}
	vCrowdColor = crowdColor;
	vCrowdLayer = crowdLayer;
	mat4 crowdSkin = crowdPlace * (skinWeight.x * crowdBoneMatrix(crowdBone + skinIndex.x) + skinWeight.y * crowdBoneMatrix(crowdBone + skinIndex.y)
		+ skinWeight.z * crowdBoneMatrix(crowdBone + skinIndex.z) + skinWeight.w * crowdBoneMatrix(crowdBone + skinIndex.w));
	vec3 crowdNormal = (crowdSkin * vec4(normal, 0.0)).xyz;`,
		)
		.replace('#include <skinbase_vertex>', '')
		.replace('#include <skinnormal_vertex>', 'objectNormal = (crowdSkin * vec4(objectNormal, 0.0)).xyz;')
		.replace('#include <skinning_vertex>', 'transformed = (crowdSkin * vec4(transformed, 1.0)).xyz;')
		.replace('vWowNormal = normalize(normalMatrix * normal);', 'vWowNormal = normalize(normalMatrix * crowdNormal);');
}
