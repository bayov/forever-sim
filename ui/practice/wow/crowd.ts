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
import { drawOrder } from './particles';
import { poseClient } from './poseClient';

// The bone texture is this many texels wide: four texels to a bone, so 256 bones a row.
const WIDTH = 1024;
const BONES_PER_ROW = WIDTH / 4;

// A skeleton's bones in the texture: where they start, and where they come from.
type Skeleton = {
	offset: number;
	// How many of the crowd's meshes use it (a body and its cape).
	users: number;
	skeleton: THREE.Skeleton;
	// The model that poses the bones (model.ts), whose skinning matrices we copy.
	model?: Model;
	frame: number;
	// The model's pose we last copied (Model.poseVersion).
	version: number;
};

// The painted skins of one size, as layers of one texture array.
// uploaded says the GPU has the array, so a new layer goes up on its own.
// The gear's textures go in arrays too, by size and wrapping (wrapS, wrapT), for the merged
// picture (see Picture).
type SkinArray = { width: number; height: number; layers: number; capacity: number; data: Uint8Array; texture: THREE.DataArrayTexture; uniform: { value: THREE.DataArrayTexture }; uploaded: boolean; wrapS?: THREE.Wrapping; wrapT?: THREE.Wrapping };

// A copy's data in its batch's buffer: [attribute, offset, size], STRIDE floats in all.
const layout: [string, number, number][] = [
	['crowdBone', 0, 1],
	['crowdLayer', 1, 1],
	['crowdColor', 2, 4],
	['crowdPlace', 6, 16],
	['crowdMaskA', 22, 4],
	['crowdMaskB', 26, 4],
];
const STRIDE = 32;
// The map of a batch whose skins come from an array (see makeBatch).
const standIn = new THREE.DataTexture(new Uint8Array([128, 128, 128, 255]), 1, 1);
standIn.colorSpace = THREE.SRGBColorSpace;
standIn.needsUpdate = true;

// The merged shadow: the raid test's raiders cast their solid shadows (bodies and gear, by side) in
// one draw call. It's one mesh of all their shadow triangles, each vertex with its bones in the
// bone texture and its raider's place in the world in a places texture, rebuilt as raiders join.
// Parts cut by their texture's alpha and the effects' shadows still draw in their batches.

// The layer of the batches that draw only in the shadows. The room's camera sees it only while
// the shadows draw (scene.ts).
export const SHADOW_LAYER = 30;

// The place of a merged entry that's hidden: everything at one point.
const nowhere = new Array<number>(16).fill(0);

// One copy's data, written here first to see whether it changed.
const copy = new Float32Array(STRIDE);

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
	// The batches whose shadows draw together (solid, not cut by alpha), with a material of each,
	// and the shadow batches (one per side), which every raider of the body is in.
	solid: Map<string, THREE.Material>;
	shadows: Batch[];
};

// One kind of part, drawn for every copy.
type Batch = {
	mesh: THREE.Mesh;
	geometry: THREE.InstancedBufferGeometry;
	material: THREE.Material & { color?: THREE.Color; opacity: number };
	// The array the batch reads its skins from, if any.
	skins?: SkinArray;
	// The copies' data, STRIDE floats each (see layout), in one buffer: their bone offset, skin
	// layer, color, place and mask (for a body's batches, see Body). One buffer goes up in one call,
	// and only when a copy changed.
	data: THREE.InstancedInterleavedBuffer;
	masked: boolean;
	// Whether a copy with no alpha draws nothing (its blending goes by the alpha), so we leave it out.
	clearAtZero: boolean;
	count: number;
	// The count the GPU has, and whether the data changed since.
	sent: number;
	dirty: boolean;
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
	// merged says the part draws in the merged picture (see Picture), not its batch.
	parts: { batch: Batch; source: THREE.Material & { color?: THREE.Color; opacity: number }; layer: number; merged?: boolean }[];
	// The submeshes a body shows (see Body), as 24 bits a float.
	mask?: Float32Array;
	// The batches its shadow draws in: all its solid parts in one (by side), apart from the ones
	// cut by their texture's alpha, which draw their own.
	shadows: Batch[];
	// A raider's body or gear has a slot in the places texture, for the merged shadow (see
	// MergedShadow) and the merged picture (see Picture). shadowMerged says its solid shadow is
	// in the merged shadow.
	slot?: number;
	shadowMerged?: boolean;
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
	// The shadow batches of each geometry (see shadowsOf).
	private shadowSets = new Map<string, Batch[]>();
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
	private bonesChanged = false;
	// The merged shadow (see MergedShadow): its meshes (one per side), the places texture of its
	// entries, and whether it needs building again.
	// mergeShadows is off with ?mergeshadows=0.
	mergeShadows = new URLSearchParams(location.search).get('mergeshadows') !== '0';
	private mergedShadows: THREE.Mesh[] = [];
	private slots = 0;
	private freeSlots: number[] = [];
	private placesData = new Float32Array(WIDTH * 4 * 4);
	private placesTexture = this.makePlaces(4);
	private readonly placesUniform = { value: this.placesTexture as THREE.Texture };
	private placesChanged = false;
	private shadowDirty = false;
	private shadowBuilt = 0;
	private readonly mergedDepth = this.makeMergedDepth();
	// The merged picture (see Picture): its meshes, one per kind of part. Off with ?mergepicture=0.
	mergePicture = new URLSearchParams(location.search).get('mergepicture') !== '0';
	private pictures: THREE.Mesh[] = [];
	private pictureDirty = false;
	private pictureBuilt = 0;

	// useSkinArrays puts painted skins in texture arrays. Off, each skin draws on its own (to
	// measure what the arrays save, with ?skinarrays=0).
	useSkinArrays = new URLSearchParams(location.search).get('skinarrays') !== '0';

	// useWorker poses the crowd in the pose worker (poseClient.ts). Off with ?poseworker=0.
	useWorker = new URLSearchParams(location.search).get('poseworker') !== '0';

	// noShadowSlots are the gear slots whose models cast no shadow: in a big raid the helms,
	// shoulders and cloaks, whose shadows mostly fall in the body's.
	noShadowSlots = new Set<number>();

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
					const solid = (model.m2.materials[b?.material ?? -1]?.blend ?? 0) <= 1 && !this.noShadowSlots.has(model.slot ?? 0);
					// A solid part's shadow draws with the model's other solid parts (shadowsOf), unless
					// its texture's alpha cuts it.
					batch = this.makeBatch(geometry, g.start, g.count, source, solid && source.alphaTest > 0, skins);
					this.batches.set(key, batch);
				}
				parts.push({ batch, source, layer: skins ? this.layerOf(skins, map!, model) : 0 });
			}
			// The model's own mesh draws no more, but what hangs off its bones (its sockets) still
			// follows them. A hidden mesh would hide those too, so it goes on a layer no camera
			// looks at.
			mesh.layers.set(31);
			// The parts draw by kind among the see-through things (for the perf recorder): a raider's
			// body and gear, or a spell effect.
			const order = body || model.slot !== undefined ? drawOrder.gear : drawOrder.effects;
			for (const p of parts) if (p.batch.users === 0) p.batch.mesh.renderOrder = order;
			const shadows = body ? body.shadows : this.shadowsOf(model, mats);
			const entry: Entry = { model, mesh, skeleton, parts, mask: body?.mask, shadows: [...shadows] };
			for (const b of entry.shadows) this.grow(b);
			// A raider's solid shadows (its body's and gear's) go into the merged shadow; an
			// effect's stay in their batches, since effects come and go.
			if ((this.mergeShadows || this.mergePicture) && (body || model.slot !== undefined)) {
				entry.slot = this.placeSlot();
				entry.shadowMerged = this.mergeShadows && entry.shadows.length > 0;
				this.shadowDirty = true;
				this.pictureDirty = true;
			}
			this.entries.push(entry);
			this.byModel.set(model, entry);
			model.crowd = this;
			// The crowd's models pose in the pose worker, when the page can share memory with it.
			if (this.useWorker) poseClient.attach(model);
			for (const p of parts) this.grow(p.batch);
		}
	}

	// bodyParts puts a character's body in its race and sex's shared batches (see Body), with the
	// mask of its parts that show. It's undefined for a model that isn't a painted body.
	private bodyParts(model: Model, mats: Entry['parts'][number]['source'][]): { parts: Entry['parts']; mask: Float32Array; shadows: Batch[] } | undefined {
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
			body = { full, sub: new THREE.BufferAttribute(sub, 1), shown: new Set(), parts: new Map(), batches: new Map(), solid: new Map(), shadows: [] };
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
				// A part cut by its texture's alpha casts its own shadow, unless its texture is a painted
				// skin (the shadows can't read the arrays, so they take it whole).
				const own = solid && mats[i].alphaTest > 0 && !skinsOf.has(i);
				const batch = this.makeBatch(full, 0, 0, mats[i], own, skinsOf.get(i), body.sub);
				if (solid && !own) body.solid.set(k, mats[i]);
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
		return { parts, mask, shadows: body.shadows };
	}

	// shadowsOf gives the batches a model's shadow draws in: one per side (front, back, both) for
	// all its solid parts that its texture's alpha doesn't cut. Copies of the model share them.
	private shadowsOf(model: Model, mats: Entry['parts'][number]['source'][]): Batch[] {
		if (this.noShadowSlots.has(model.slot ?? 0)) return [];
		const geometry = model.mesh.geometry;
		const key = `shadow${idOf(geometry.getAttribute('position'))}:${idOf(geometry.index!)}`;
		const known = this.shadowSets.get(key);
		if (known) return known;
		const bySide = new Map<number, [number, number][]>();
		const first = new Map<number, Entry['parts'][number]['source']>();
		for (const g of geometry.groups) {
			const mi = g.materialIndex ?? 0;
			const m = mats[mi];
			const b = model.skin.batches[mi];
			if (!m?.visible || !b || (model.m2.materials[b.material]?.blend ?? 0) > 1 || m.alphaTest > 0) continue;
			const list = bySide.get(m.side) ?? [];
			bySide.set(m.side, list);
			list.push([g.start, g.count]);
			if (!first.has(m.side)) first.set(m.side, m);
		}
		const out: Batch[] = [];
		for (const [side, ranges] of bySide) out.push(this.shadowBatch(geometry, ranges, first.get(side)!));
		this.shadowSets.set(key, out);
		return out;
	}

	// shadowBatch makes a batch that draws only in the shadows (on SHADOW_LAYER), for these ranges
	// of a geometry's index.
	private shadowBatch(geometry: THREE.BufferGeometry, ranges: [number, number][], source: Batch['material'], sub?: THREE.BufferAttribute, index?: THREE.BufferAttribute, start = 0): Batch {
		let count = 0;
		for (const [, c] of ranges) count += c;
		if (!index) {
			const src = geometry.index!.array as Uint16Array | Uint32Array;
			const array = src instanceof Uint32Array ? new Uint32Array(count) : new Uint16Array(count);
			let at = 0;
			for (const [s, c] of ranges) {
				array.set(src.subarray(s, s + c), at);
				at += c;
			}
			index = new THREE.BufferAttribute(array, 1);
		}
		const batch = this.makeBatch(geometry, start, count, source, true, undefined, sub);
		// The shadows need no texture here (no part in it is cut by alpha), and a painted skin may
		// have let go of its picture.
		(batch.material as THREE.MeshLambertMaterial).map = standIn;
		batch.material.alphaTest = 0;
		batch.geometry.setIndex(index);
		batch.geometry.setDrawRange(start, count);
		batch.mesh.layers.set(SHADOW_LAYER);
		this.batches.set(`shadowbatch${idOf(batch)}`, batch);
		return batch;
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
		// The shadows' ranges come after the batches', by side.
		const sides = new Map<number, [number, number][]>();
		for (const [k, list] of byKey) {
			const m = body.solid.get(k);
			if (!m) continue;
			const ranges = sides.get(m.side) ?? [];
			sides.set(m.side, ranges);
			ranges.push(...list);
		}
		let total = 0;
		for (const list of [...byKey.values(), ...sides.values()]) for (const [, c] of list) total += c;
		const out = src instanceof Uint32Array ? new Uint32Array(total) : new Uint16Array(total);
		let at = 0;
		const index = new THREE.BufferAttribute(out, 1);
		const put = (list: [number, number][]) => {
			const start = at;
			for (const [s, c] of list) {
				out.set(src.subarray(s, s + c), at);
				at += c;
			}
			return start;
		};
		for (const [k, list] of byKey) {
			const start = put(list);
			const batch = body.batches.get(k)!;
			batch.geometry.setIndex(index);
			batch.geometry.setDrawRange(start, at - start);
		}
		const shadows: Batch[] = [];
		for (const [side, list] of sides) {
			const start = put(list);
			const old = body.shadows.find(b => b.material.side === side);
			const m = [...body.solid.values()].find(x => x.side === side) as Batch['material'];
			const batch = old ?? this.shadowBatch(full, [], m, body.sub, index, start);
			batch.geometry.setIndex(index);
			batch.geometry.setDrawRange(start, at - start);
			shadows.push(batch);
		}
		// The raiders already in the body take the new shadow batches too.
		for (const b of shadows) {
			if (body.shadows.includes(b)) continue;
			body.shadows.push(b);
			for (const e of this.entries) if (e.mask && e.model.fullGeometry === full) {
				e.shadows.push(b);
				this.grow(b);
			}
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
		poseClient.detach(model);
		if (entry.slot !== undefined) {
			this.freeSlots.push(entry.slot);
			this.shadowDirty = true;
			this.pictureDirty = true;
		}
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

	// makeMergedDepth makes the merged shadow's depth material: each vertex reads its bones from
	// the bone texture and its entry's place from the places texture.
	private makeMergedDepth(): THREE.MeshDepthMaterial {
		const depth = new THREE.MeshDepthMaterial();
		depth.onBeforeCompile = shader => {
			shader.uniforms.crowdBones = this.uniform;
			shader.uniforms.crowdPlaces = this.placesUniform;
			shader.vertexShader = shader.vertexShader
				.replace(
					'#include <skinning_pars_vertex>',
					/* glsl */ `attribute vec4 crowdBoneIds;
attribute vec4 crowdWeights;
attribute float crowdSlot;
uniform highp sampler2D crowdBones;
uniform highp sampler2D crowdPlaces;
mat4 crowdMatrix(highp sampler2D tex, float i) {
	int j = int(i) * 4;
	int x = j % ${WIDTH};
	int y = j / ${WIDTH};
	return mat4(texelFetch(tex, ivec2(x, y), 0), texelFetch(tex, ivec2(x + 1, y), 0), texelFetch(tex, ivec2(x + 2, y), 0), texelFetch(tex, ivec2(x + 3, y), 0));
}`,
				)
				.replace(
					'#include <skinning_vertex>',
					/* glsl */ `mat4 crowdSkin = crowdMatrix(crowdPlaces, crowdSlot) * (crowdWeights.x * crowdMatrix(crowdBones, crowdBoneIds.x) + crowdWeights.y * crowdMatrix(crowdBones, crowdBoneIds.y)
		+ crowdWeights.z * crowdMatrix(crowdBones, crowdBoneIds.z) + crowdWeights.w * crowdMatrix(crowdBones, crowdBoneIds.w));
	transformed = (crowdSkin * vec4(transformed, 1.0)).xyz;`,
				)
				.replace('#include <skinbase_vertex>', '')
				.replace('#include <skinnormal_vertex>', '');
		};
		depth.customProgramCacheKey = () => 'crowd-merged-depth';
		return depth;
	}

	// skeletonOf gives a skeleton its place in the bone texture. A cape shares the body's.
	private skeletonOf(s: THREE.Skeleton): Skeleton {
		let k = this.skeletons.get(s);
		if (k) {
			k.users++;
			return k;
		}
		const n = s.bones.length;
		k = { offset: this.allocate(n), users: 1, skeleton: s, model: skins.get(s), frame: -1, version: -1 };
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
		t.wrapS = a.wrapS ?? THREE.ClampToEdgeWrapping;
		t.wrapT = a.wrapT ?? THREE.ClampToEdgeWrapping;
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
		const data = this.copyData(geometry, 16, !!sub);
		if (sub) geometry.setAttribute('crowdSub', sub);
		geometry.instanceCount = 0;

		const material = sourceMaterial.clone() as Batch['material'];
		// A batch that reads its skins from an array keeps a map only so its shader samples one (and
		// gets the texture coordinates). The first raider's own skin there would go to the GPU
		// (after we let go of its picture, as an empty one), so it gets a one-pixel stand-in.
		if (skins) (material as THREE.MeshLambertMaterial).map = standIn;
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
		const m = material;
		const clearAtZero =
			m.alphaTest > 0 || m.blending === THREE.NormalBlending || (m.blending === THREE.CustomBlending && (m.blendSrc === THREE.SrcAlphaFactor || m.premultipliedAlpha));
		return { mesh, geometry, material, data, skins, masked: !!sub, clearAtZero, count: 0, sent: 0, dirty: true, users: 0, castShadow: solid };
	}

	// copyData gives a batch's geometry a buffer for n copies, with the attributes that read it.
	private copyData(geometry: THREE.InstancedBufferGeometry, n: number, masked: boolean): THREE.InstancedInterleavedBuffer {
		const data = new THREE.InstancedInterleavedBuffer(new Float32Array(n * STRIDE), STRIDE);
		data.setUsage(THREE.DynamicDrawUsage);
		for (const [name, offset, size] of layout) {
			if (!masked && name.startsWith('crowdMask')) continue;
			geometry.setAttribute(name, new THREE.InterleavedBufferAttribute(data, size, offset));
		}
		// Three keeps the most copies a geometry's attributes hold from its first draw, and draws no
		// more than that.
		(geometry as { _maxInstanceCount?: number })._maxInstanceCount = undefined;
		return data;
	}

	// grow makes room in a batch for one more copy.
	private grow(b: Batch) {
		b.users++;
		if (b.users <= b.data.count) return;
		let n = b.data.count;
		while (n < b.users) n *= 2;
		b.data = this.copyData(b.geometry, n, b.masked);
		b.dirty = true;
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
		const one = copy;
		// put adds the copy in `one` to a batch. Most raiders stand still, so most copies are as
		// they were: we write (and send) only what changed.
		const put = (b: Batch) => {
			const out = b.data.array as Float32Array;
			const o = b.count++ * STRIDE;
			for (let j = 0; j < 30; j++) {
				if (out[o + j] !== one[j]) {
					out[o + j] = one[j];
					b.dirty = true;
				}
			}
		};
		for (const e of this.entries) {
			// A merged entry's place is in the merged shadow every frame, hidden or not (a sheathed
			// weapon that hides casts nothing: its place shrinks to a point).
			if (e.slot !== undefined) this.writePlace(e.slot, shown(e.mesh) ? e.mesh.matrixWorld.elements : nowhere);
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
			// The copy's bone offset, place and mask are the same in all its batches.
			one[0] = k.offset;
			const w = e.mesh.matrixWorld.elements;
			for (let j = 0; j < 16; j++) one[6 + j] = w[j];
			const mask = e.mask;
			for (let j = 0; j < 8; j++) one[22 + j] = mask ? mask[j] : 0;
			if (!e.shadowMerged && inShadow) {
				one[1] = 0;
				one[2] = one[3] = one[4] = one[5] = 1;
				for (const b of e.shadows) put(b);
			}
			for (const { batch: b, source, layer, merged } of e.parts) {
				if (merged || (!inView && !b.castShadow)) continue;
				one[1] = layer;
				const color = source.color;
				one[2] = color ? color.r : 1;
				one[3] = color ? color.g : 1;
				one[4] = color ? color.b : 1;
				one[5] = source.opacity;
				// A see-through layer faded all the way out (a flash between its flashes) still cost
				// its pixels. In the raid test a fifth of the see-through copies were.
				if (b.clearAtZero && source.transparent && source.opacity < 1 / 255) continue;
				put(b);
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
			// The copies' data goes up whole when it changed. It's a few kilobytes, and on Windows each
			// upload is its own copy on the GPU, which for 400 batches of 5 attributes each was
			// most of the 3ms before the shadows.
			if (b.dirty || b.count !== b.sent) {
				b.data.needsUpdate = true;
				b.dirty = false;
				b.sent = b.count;
			}
		}
		if (this.bonesChanged) this.texture.needsUpdate = true;
		this.bonesChanged = false;
		if (this.placesChanged) this.placesTexture.needsUpdate = true;
		this.placesChanged = false;
		if (this.shadowDirty && performance.now() - this.shadowBuilt > 300) this.buildMergedShadow();
		if (this.pictureDirty && performance.now() - this.pictureBuilt > 1000) this.buildPicture();
	}

	// placeSlot gives a merged entry its slot in the places texture.
	private placeSlot(): number {
		const free = this.freeSlots.pop();
		if (free !== undefined) return free;
		const slot = this.slots++;
		const rows = Math.ceil((this.slots * 4) / WIDTH);
		if (rows > this.placesTexture.image.height) {
			const height = Math.ceil(rows / 4) * 4;
			const data = new Float32Array(WIDTH * height * 4);
			data.set(this.placesData);
			this.placesData = data;
			this.placesTexture.dispose();
			this.placesTexture = this.makePlaces(height);
			this.placesUniform.value = this.placesTexture;
		}
		return slot;
	}

	private makePlaces(height: number): THREE.DataTexture {
		const t = new THREE.DataTexture(this.placesData, WIDTH, height, THREE.RGBAFormat, THREE.FloatType);
		t.minFilter = t.magFilter = THREE.NearestFilter;
		t.generateMipmaps = false;
		t.needsUpdate = true;
		return t;
	}

	// writePlace puts a merged entry's place in the world in the places texture, when it moved.
	private writePlace(slot: number, w: ArrayLike<number>) {
		const d = this.placesData;
		const o = slot * 16;
		for (let j = 0; j < 16; j++) {
			if (d[o + j] !== Math.fround(w[j])) {
				d[o + j] = w[j];
				this.placesChanged = true;
			}
		}
	}

	// gearArray gives the texture array for a gear texture's size and wrapping, or undefined for a
	// texture that wouldn't sample the same from an array (not the plain kind every BLP decodes to).
	private gearArray(t: THREE.Texture): SkinArray | undefined {
		const image = t.image as { width: number; height: number; data?: ArrayLike<number> | null };
		if (!(t as THREE.DataTexture).isDataTexture || !image?.data || t.userData.composed) return undefined;
		if (t.format !== THREE.RGBAFormat || t.type !== THREE.UnsignedByteType || t.colorSpace !== THREE.SRGBColorSpace) return undefined;
		if (t.magFilter !== THREE.LinearFilter || t.minFilter !== THREE.LinearMipmapLinearFilter || !t.generateMipmaps || t.anisotropy !== 4 || t.flipY) return undefined;
		if (image.data.length !== image.width * image.height * 4) return undefined;
		const key = `gear${image.width}x${image.height}:${t.wrapS}:${t.wrapT}`;
		let a = this.skinArrays.get(key);
		if (!a) {
			a = { width: image.width, height: image.height, layers: 0, capacity: 0, data: new Uint8Array(0), texture: undefined!, uniform: { value: null! }, uploaded: false, wrapS: t.wrapS, wrapT: t.wrapT };
			this.growSkins(a, 4);
			this.skinArrays.set(key, a);
		}
		return a;
	}

	// pictureKind tells how a part draws, for the merged picture: the parts of one kind draw
	// together. It's undefined for a part that draws on its own: see-through, cut by alpha, with
	// animated colors, or with a texture that can't go in an array.
	private pictureKind(e: Entry, part: Entry['parts'][number]): { key: string; maps: [SkinArray, number]; maps2?: [SkinArray, number] } | undefined {
		const m = part.source as THREE.MeshLambertMaterial & { userData: { wow?: { map2: { value: THREE.Texture | null }; wowMode: { value: number }; wowEnv0: { value: boolean }; wowEnv1: { value: boolean }; wowUv2: { value: boolean } } } };
		if (m.transparent || m.alphaTest > 0 || !m.depthWrite || m.polygonOffset || !m.map || m.vertexColors) return undefined;
		if (m.color && (m.color.r !== 1 || m.color.g !== 1 || m.color.b !== 1)) return undefined;
		// Animated colors and transparency vary by copy.
		const mats = e.mesh.material as THREE.Material[];
		const mi = mats.indexOf(m);
		const b = e.model.skin.batches[mi];
		if (!b) return undefined;
		const m2 = e.model.m2;
		const c = b.color >= 0 ? m2.colors[b.color] : undefined;
		if (c && (!e.model.steady(c.color, 1) || !e.model.steady(c.alpha, 32767))) return undefined;
		const wi = m2.transparencyLookup[b.weightCombo];
		if (wi !== undefined && m2.textureWeights[wi] && !e.model.steady(m2.textureWeights[wi], 32767)) return undefined;
		if (m.opacity !== 1) return undefined;
		let maps: [SkinArray, number];
		if (part.batch.skins) maps = [part.batch.skins, part.layer];
		else {
			const a = this.gearArray(m.map);
			if (!a) return undefined;
			maps = [a, this.layerOf(a, m.map)];
		}
		const wow = m.userData.wow;
		let maps2: [SkinArray, number] | undefined;
		if (wow && wow.wowMode.value > 0) {
			const t2 = wow.map2.value;
			const a2 = t2 ? this.gearArray(t2) : undefined;
			if (!t2 || !a2) return undefined;
			maps2 = [a2, this.layerOf(a2, t2)];
		}
		const combiner = wow ? `${wow.wowMode.value}:${wow.wowEnv0.value}:${wow.wowEnv1.value}:${wow.wowUv2.value}` : 'plain';
		const key = [m.type, m.side, m.fog, m.depthTest, combiner, idOf(maps[0]), maps2 ? idOf(maps2[0]) : 0].join('|');
		return { key, maps, maps2 };
	}

	// buildPicture builds the merged picture (see Picture) from every merged entry's parts that
	// can draw in it.
	private buildPicture() {
		this.pictureDirty = false;
		this.pictureBuilt = performance.now();
		type Kind = { source: Entry['parts'][number]['source']; maps: SkinArray; maps2?: SkinArray; pos: number[]; normal: number[]; uv: number[]; uv1: number[]; bones: number[]; weights: number[]; slot: number[]; layers: number[]; index: number[] };
		const kinds = new Map<string, Kind>();
		for (const e of this.entries) {
			for (const part of e.parts) part.merged = false;
			if (!this.mergePicture || e.slot === undefined) continue;
			const offset = e.skeleton.offset;
			for (const part of e.parts) {
				const kind = this.pictureKind(e, part);
				if (!kind) continue;
				let out = kinds.get(kind.key);
				if (!out) kinds.set(kind.key, (out = { source: part.source, maps: kind.maps[0], maps2: kind.maps2?.[0], pos: [], normal: [], uv: [], uv1: [], bones: [], weights: [], slot: [], layers: [], index: [] }));
				part.merged = true;
				const g = part.batch.geometry;
				const index = g.index!.array;
				const pos = g.getAttribute('position').array;
				const nor = g.getAttribute('normal').array;
				const uv = g.getAttribute('uv').array;
				const uv1 = g.getAttribute('uv1').array;
				const si = g.getAttribute('skinIndex').array;
				const sw = g.getAttribute('skinWeight').array;
				const sub = part.batch.masked ? g.getAttribute('crowdSub').array : undefined;
				const remap = new Map<number, number>();
				const end = g.drawRange.start + g.drawRange.count;
				for (let t = g.drawRange.start; t + 2 < end; t += 3) {
					// A body's triangle shows when its submesh does (the raider's mask).
					if (sub && e.mask) {
						const sm = sub[index[t]];
						if (Math.floor(e.mask[Math.floor(sm / 24)] / 2 ** (sm % 24)) % 2 !== 1) continue;
					}
					for (let k = 0; k < 3; k++) {
						const v = index[t + k];
						let m = remap.get(v);
						if (m === undefined) {
							m = out.slot.length;
							remap.set(v, m);
							out.pos.push(pos[v * 3], pos[v * 3 + 1], pos[v * 3 + 2]);
							out.normal.push(nor[v * 3], nor[v * 3 + 1], nor[v * 3 + 2]);
							out.uv.push(uv[v * 2], uv[v * 2 + 1]);
							out.uv1.push(uv1[v * 2], uv1[v * 2 + 1]);
							for (let j = 0; j < 4; j++) {
								out.bones.push(offset + si[v * 4 + j]);
								out.weights.push(sw[v * 4 + j]);
							}
							out.slot.push(e.slot);
							out.layers.push(kind.maps[1], kind.maps2?.[1] ?? 0);
						}
						out.index.push(m);
					}
				}
			}
		}
		for (const m of this.pictures) {
			m.removeFromParent();
			m.geometry.dispose();
			(m.material as THREE.Material).dispose();
		}
		this.pictures = [];
		for (const [key, d] of kinds) {
			const g = new THREE.BufferGeometry();
			g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(d.pos), 3));
			g.setAttribute('normal', new THREE.BufferAttribute(new Float32Array(d.normal), 3));
			g.setAttribute('uv', new THREE.BufferAttribute(new Float32Array(d.uv), 2));
			g.setAttribute('uv1', new THREE.BufferAttribute(new Float32Array(d.uv1), 2));
			g.setAttribute('crowdBoneIds', new THREE.BufferAttribute(new Float32Array(d.bones), 4));
			g.setAttribute('crowdWeights', new THREE.BufferAttribute(new Float32Array(d.weights), 4));
			g.setAttribute('crowdSlot', new THREE.BufferAttribute(new Float32Array(d.slot), 1));
			g.setAttribute('crowdLayers', new THREE.BufferAttribute(new Float32Array(d.layers), 2));
			g.setIndex(d.slot.length > 65535 ? new THREE.BufferAttribute(new Uint32Array(d.index), 1) : new THREE.BufferAttribute(new Uint16Array(d.index), 1));
			const mesh = new THREE.Mesh(g, this.pictureMaterial(d.source, d.maps, d.maps2, key));
			mesh.frustumCulled = false;
			mesh.matrixAutoUpdate = false;
			mesh.renderOrder = drawOrder.gear;
			this.group.add(mesh);
			this.pictures.push(mesh);
		}
	}

	// pictureMaterial makes a merged picture kind's material: its part's own, reading its textures
	// from the arrays by each vertex's layers, and its bones and place from the textures.
	private pictureMaterial(source: Batch['material'], maps: SkinArray, maps2: SkinArray | undefined, key: string): THREE.Material {
		const material = source.clone() as Batch['material'];
		(material as THREE.MeshLambertMaterial).map = standIn;
		const before = source.onBeforeCompile;
		material.onBeforeCompile = (shader, renderer) => {
			before.call(source, shader, renderer);
			shader.uniforms.crowdBones = this.uniform;
			shader.uniforms.crowdPlaces = this.placesUniform;
			shader.uniforms.crowdMaps = maps.uniform;
			shader.uniforms.crowdMaps2 = (maps2 ?? maps).uniform;
			shader.vertexShader = pictureVertex(shader.vertexShader);
			shader.fragmentShader = shader.fragmentShader
				.replace(
					'#include <common>',
					`#include <common>
uniform highp sampler2DArray crowdMaps;
uniform highp sampler2DArray crowdMaps2;
varying vec2 vCrowdLayers;
vec4 crowdSample(vec2 uv) { return texture(crowdMaps, vec3(uv, vCrowdLayers.x)); }
vec4 crowdSample2(vec2 uv) { return texture(crowdMaps2, vec3(uv, vCrowdLayers.y)); }`,
				)
				.replace('#include <map_fragment>', '#ifdef USE_MAP\ndiffuseColor *= crowdSample(vMapUv);\n#endif')
				.replace('texture2D(map, ', 'crowdSample(')
				.replace('texture2D(map2, ', 'crowdSample2(');
		};
		material.customProgramCacheKey = () => `crowd-picture-${key}`;
		return material;
	}

	// buildMergedShadow builds the merged shadow (MergedShadow) from every merged entry's solid
	// shadow parts: their triangles, with each vertex's bones in the bone texture and its entry's
	// slot in the places texture.
	private buildMergedShadow() {
		this.shadowDirty = false;
		this.shadowBuilt = performance.now();
		const bySide = new Map<number, { pos: number[]; bones: number[]; weights: number[]; slot: number[]; index: number[] }>();
		for (const e of this.entries) {
			if (!e.shadowMerged || e.slot === undefined) continue;
			const offset = e.skeleton.offset;
			for (const b of e.shadows) {
				const g = b.geometry;
				const index = g.index!.array;
				const pos = g.getAttribute('position').array;
				const si = g.getAttribute('skinIndex').array;
				const sw = g.getAttribute('skinWeight').array;
				const sub = b.masked ? g.getAttribute('crowdSub').array : undefined;
				const side = b.material.side;
				let out = bySide.get(side);
				if (!out) bySide.set(side, (out = { pos: [], bones: [], weights: [], slot: [], index: [] }));
				const remap = new Map<number, number>();
				const end = g.drawRange.start + g.drawRange.count;
				for (let t = g.drawRange.start; t + 2 < end; t += 3) {
					// A body's triangle shows when its submesh does (the raider's mask).
					if (sub && e.mask) {
						const sm = sub[index[t]];
						if (Math.floor(e.mask[Math.floor(sm / 24)] / 2 ** (sm % 24)) % 2 !== 1) continue;
					}
					for (let k = 0; k < 3; k++) {
						const v = index[t + k];
						let m = remap.get(v);
						if (m === undefined) {
							m = out.slot.length;
							remap.set(v, m);
							out.pos.push(pos[v * 3], pos[v * 3 + 1], pos[v * 3 + 2]);
							for (let j = 0; j < 4; j++) {
								out.bones.push(offset + si[v * 4 + j]);
								out.weights.push(sw[v * 4 + j]);
							}
							out.slot.push(e.slot);
						}
						out.index.push(m);
					}
				}
			}
		}
		for (const m of this.mergedShadows) {
			m.removeFromParent();
			m.geometry.dispose();
		}
		this.mergedShadows = [];
		for (const [side, d] of bySide) {
			const g = new THREE.BufferGeometry();
			g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(d.pos), 3));
			g.setAttribute('crowdBoneIds', new THREE.BufferAttribute(new Float32Array(d.bones), 4));
			g.setAttribute('crowdWeights', new THREE.BufferAttribute(new Float32Array(d.weights), 4));
			g.setAttribute('crowdSlot', new THREE.BufferAttribute(new Float32Array(d.slot), 1));
			g.setIndex(d.slot.length > 65535 ? new THREE.BufferAttribute(new Uint32Array(d.index), 1) : new THREE.BufferAttribute(new Uint16Array(d.index), 1));
			const mesh = new THREE.Mesh(g, new THREE.MeshBasicMaterial({ side: side as THREE.Side }));
			mesh.customDepthMaterial = this.mergedDepth;
			mesh.castShadow = true;
			mesh.frustumCulled = false;
			mesh.matrixAutoUpdate = false;
			mesh.layers.set(SHADOW_LAYER);
			this.group.add(mesh);
			this.mergedShadows.push(mesh);
		}
	}

	// writeBones puts a skeleton's skinning matrices in the texture, when they changed since. A
	// raider far away poses only now and then, and gear that never moves only once.
	private writeBones(k: Skeleton) {
		const m = k.model;
		if (!m || m.poseVersion === k.version) return;
		k.version = m.poseVersion;
		this.data.set(m.skinning, k.offset * 16);
		this.bonesChanged = true;
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

// pictureVertex skins a merged picture's vertex by its bones in the bone texture and its raider's
// place in the places texture (see Picture), in place of Three's skinning, the way crowdVertex
// does for a batch's copies.
function pictureVertex(src: string): string {
	return src
		.replace(
			'#include <skinning_pars_vertex>',
			/* glsl */ `attribute vec4 crowdBoneIds;
attribute vec4 crowdWeights;
attribute float crowdSlot;
attribute vec2 crowdLayers;
uniform highp sampler2D crowdBones;
uniform highp sampler2D crowdPlaces;
varying vec2 vCrowdLayers;
mat4 crowdMatrix(highp sampler2D tex, float i) {
	int j = int(i) * 4;
	int x = j % ${WIDTH};
	int y = j / ${WIDTH};
	return mat4(texelFetch(tex, ivec2(x, y), 0), texelFetch(tex, ivec2(x + 1, y), 0), texelFetch(tex, ivec2(x + 2, y), 0), texelFetch(tex, ivec2(x + 3, y), 0));
}`,
		)
		.replace(
			'void main() {',
			/* glsl */ `void main() {
	vCrowdLayers = crowdLayers;
	mat4 crowdSkin = crowdMatrix(crowdPlaces, crowdSlot) * (crowdWeights.x * crowdMatrix(crowdBones, crowdBoneIds.x) + crowdWeights.y * crowdMatrix(crowdBones, crowdBoneIds.y)
		+ crowdWeights.z * crowdMatrix(crowdBones, crowdBoneIds.z) + crowdWeights.w * crowdMatrix(crowdBones, crowdBoneIds.w));
	vec3 crowdNormal = (crowdSkin * vec4(normal, 0.0)).xyz;`,
		)
		.replace('#include <skinbase_vertex>', '')
		.replace('#include <skinnormal_vertex>', 'objectNormal = (crowdSkin * vec4(objectNormal, 0.0)).xyz;')
		.replace('#include <skinning_vertex>', 'transformed = (crowdSkin * vec4(transformed, 1.0)).xyz;')
		.replace('vWowNormal = normalize(normalMatrix * normal);', 'vWowNormal = normalize(normalMatrix * crowdNormal);');
}
