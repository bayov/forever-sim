// A WoW model drawn with Three.js: its mesh, materials, skeleton and animations.
//
// WoW moves a bone around its pivot: T(pivot) T(move) R S T(-pivot), on top of its parent. We pose
// the bones ourselves, in flat arrays: each bone's place in the model is its parent's times
// T(pivot - parent's pivot + move) R S, and the skinning takes that times T(-pivot).
//
// The bones used to be Three.js bones in the scene. In a raid of 40 that was 9000 objects the
// renderer placed and walked through twice a frame (for the shadows and the picture), which was
// most of the time it spent before drawing. Now the scene holds only the sockets: the attachment
// points and emitters that something hangs off, which follow their bone.

import * as THREE from 'three';
import { file, texture } from './files';
import { animData, Batch, Keys, M2, M2Event, sample, sampleTime, Skin, Track, Vec } from './m2';
import { Particles } from './particles';
import { Ribbons } from './ribbons';

// AnimationData ids we play.
export const Anim = {
	Stand: 0,
	Death: 1,
	Walk: 4,
	Run: 5,
	Dead: 6,
	CombatWound: 9,
	ShuffleLeft: 11,
	ShuffleRight: 12,
	WalkBackwards: 13,
	Stun: 14,
	HandsClosed: 15,
	AttackUnarmed: 16,
	Attack1H: 17,
	Attack2H: 18,
	Attack2HL: 19,
	ParryUnarmed: 20,
	Parry1H: 21,
	ReadyUnarmed: 25,
	Ready1H: 26,
	Ready2H: 27,
	Ready2HL: 28,
	CombatCritical: 10,
	Dodge: 30,
	SpellPrecast: 31,
	SpellCast: 32,
	SpellCastArea: 33,
	JumpStart: 37,
	Jump: 38,
	JumpEnd: 39,
	Fall: 40,
	ReadySpellDirected: 51,
	ReadySpellOmni: 52,
	SpellCastDirected: 53,
	SpellCastOmni: 54,
	BattleRoar: 55,
	ReadyAbility: 56,
	Special1H: 57,
	Special2H: 58,
	AttackOff: 87,
	Sheath: 89,
	HipSheath: 90,
	RunRight: 92,
	RunLeft: 93,
	AttackUnarmedOff: 117,
	ChannelCastDirected: 124,
	ChannelCastOmni: 125,
	Birth: 127,
	Sprint: 143,
	JumpLandRun: 187,
} as const;

// When a model lacks an animation, the client plays a related one (AnimationData.Fallback).
const fallbacks: Record<number, number> = {
	[Anim.WalkBackwards]: Anim.Walk,
	[Anim.RunRight]: Anim.Run,
	[Anim.RunLeft]: Anim.Run,
	[Anim.ShuffleLeft]: Anim.Walk,
	[Anim.ShuffleRight]: Anim.Walk,
	[Anim.Sprint]: Anim.Run,
	[Anim.JumpLandRun]: Anim.Run,
	[Anim.Walk]: Anim.Run,
	[Anim.Run]: Anim.Stand,
	[Anim.Attack1H]: Anim.AttackUnarmed,
	[Anim.Attack2H]: Anim.Attack1H,
	[Anim.Attack2HL]: Anim.Attack2H,
	[Anim.AttackOff]: Anim.Attack1H,
	[Anim.AttackUnarmed]: Anim.Stand,
	[Anim.AttackUnarmedOff]: Anim.AttackUnarmed,
	[Anim.Ready1H]: Anim.ReadyUnarmed,
	[Anim.Ready2H]: Anim.Ready1H,
	[Anim.Ready2HL]: Anim.Ready2H,
	[Anim.ReadyUnarmed]: Anim.Stand,
	[Anim.ReadySpellDirected]: Anim.ReadySpellOmni,
	[Anim.ReadySpellOmni]: Anim.SpellPrecast,
	[Anim.SpellCastDirected]: Anim.SpellCastOmni,
	[Anim.SpellCastOmni]: Anim.SpellCast,
	[Anim.SpellCastArea]: Anim.SpellCast,
	[Anim.ChannelCastDirected]: Anim.ReadySpellOmni,
	[Anim.ChannelCastOmni]: Anim.SpellCastOmni,
	[Anim.SpellPrecast]: Anim.Stand,
	[Anim.SpellCast]: Anim.Stand,
	[Anim.Special1H]: Anim.Attack1H,
	[Anim.Special2H]: Anim.Attack2H,
	[Anim.JumpStart]: Anim.Jump,
	[Anim.JumpEnd]: Anim.Stand,
	[Anim.Fall]: Anim.Jump,
	[Anim.Jump]: Anim.Stand,
	[Anim.CombatWound]: Anim.Stand,
	[Anim.CombatCritical]: Anim.CombatWound,
	[Anim.Dodge]: Anim.Stand,
	[Anim.Parry1H]: Anim.ParryUnarmed,
	[Anim.ParryUnarmed]: Anim.Stand,
	[Anim.BattleRoar]: Anim.SpellCast,
	[Anim.Birth]: Anim.Stand,
};

export type PlayOptions = { loop?: boolean; speed?: number; fade?: number; restart?: boolean };

type Playing = {
	anim: number;
	// When it started, in the model's own clock (ms), for the upper body's fade in.
	started: number;
	seq: number;
	time: number;
	speed: number;
	loop: boolean;
	source?: DataView;
	done: boolean;
};

type BatchMaterial = {
	material: THREE.Material & { opacity: number; color?: THREE.Color };
	batch: Batch;
	submesh: number;
	textureType: number;
	baseOpacity: number;
	// Whether the material starts see-through, for a model used again (recycle).
	baseTransparent: boolean;
};

// Bones that move the fingers: right hand 8 to 12, left hand 13 to 17 (key bone ids).
const rightFingers = [8, 9, 10, 11, 12];
const leftFingers = [13, 14, 15, 16, 17];

const upAxis = new THREE.Vector3(0, 1, 0);
const twistQ = new THREE.Quaternion();
const parentQ = new THREE.Quaternion();
const boneQ = new THREE.Quaternion();
const twistAxis = new THREE.Vector3();
const parentWorld = new THREE.Matrix4();

// skins gives each skeleton's skinning matrices (in its model's space), for the crowd (crowd.ts).
// A cape draws with the body's skeleton, so it finds the body's here.
export const skins = new WeakMap<THREE.Skeleton, Float32Array>();

// A socket is an object that follows a bone, at an offset from its pivot.
type Socket = { object: THREE.Object3D; bone: number; x: number; y: number; z: number };

export class Model {
	readonly root = new THREE.Group();
	readonly mesh: THREE.SkinnedMesh;
	// Each bone's place in the model (a 4 by 4 matrix, by columns), and its skinning matrix: the
	// place times T(-pivot).
	private readonly places: Float64Array;
	readonly skinning: Float32Array;
	// The bones' rotations after the twist and the billboards, which the twist needs for the
	// bones above.
	private readonly quats: Float64Array;
	// The bones in an order where each parent comes before its children.
	private readonly order: Int32Array;
	// How much of the twist each bone takes (twistShares), and whether it faces the camera.
	private readonly twistShare: Float64Array;
	private readonly billboard: Uint8Array;
	private readonly sockets: Socket[] = [];
	private readonly mats: BatchMaterial[] = [];
	private readonly replaceable = new Map<number, THREE.Texture>();
	private readonly anims = new Map<number, Promise<DataView | undefined>>();
	private readonly ready = new Map<number, DataView>();
	private base: Playing;
	// A switch of animation fades from the pose the bones had (the snapshot), so a second switch
	// in the middle of a fade doesn't pop.
	private pose0: Float32Array;
	private posed = false;
	private fadePose: Float32Array;
	private fadeLeft = 0;
	private fadeTotal = 0;
	private overlay?: Playing;
	private upper: boolean[];
	// The finger bones, with every joint below them, for holding a weapon.
	private fingers: { right: boolean[]; left: boolean[] };
	private points = new Map<number, THREE.Object3D>();
	// Whether any bone always faces the camera (bone flag 0x8): a totem's rune, the glow cards of
	// effects.
	private billboards = false;
	// The camera that billboards face.
	static camera?: THREE.Camera;
	// twist turns the upper body around the model's up axis, in radians. A strafe turns the legs
	// where you run and twists the upper body back to where you face.
	//
	// The twist is spread up the spine: the lower spine takes most of it, the chest a bit more and
	// the head the rest. So the head faces where you face but the chest stays
	// turned a little toward where you run.
	twist = 0;
	private events = new Map<number, { id: string; data: number; time: number }[] | null>();
	// onEvent hears the animations' events as they pass ($HIT, $CSS, $FSD footsteps, $SHL...).
	// upper says it came from the upper body's animation (a swing while running).
	onEvent?: (id: string, data: number, upper: boolean) => void;
	private globalTime = 0;
	private closed = { right: false, left: false };
	private scratch = { t: new Float64Array(3), r: Float64Array.of(0, 0, 0, 1), s: Float64Array.of(1, 1, 1), t2: new Float64Array(3), r2: Float64Array.of(0, 0, 0, 1), s2: Float64Array.of(1, 1, 1) };
	geosets?: Set<number>;
	// The crowd the model draws in (crowd.ts), if any.
	crowd?: { remove(m: Model): void };
	particles?: Particles;
	ribbons?: Ribbons;
	// The model's own textures, which load() waits for so a model never shows untextured.
	private readonly loading: Promise<unknown>[] = [];

	readonly m2: M2;
	readonly skin: Skin;

	private constructor(private readonly shared: Parsed) {
		const { m2, skin } = shared;
		this.m2 = m2;
		this.skin = skin;
		shared.users++;
		// The vertices are the same for every copy of the model, so the copies share them (and
		// their memory on the GPU). Each copy gets its own list of parts, which compact changes.
		const geometry = new THREE.BufferGeometry();
		for (const [name, a] of Object.entries(shared.geometry.attributes)) geometry.setAttribute(name, a);
		geometry.setIndex(shared.geometry.index);
		for (const g of shared.geometry.groups) geometry.addGroup(g.start, g.count, g.materialIndex);
		const materials: THREE.Material[] = [];
		skin.batches.forEach((batch, i) => {
			const bm = this.makeMaterial(batch, i);
			this.mats.push(bm);
			materials.push(bm.material);
		});
		// Draw the batches in their layer order, so a glow layer draws over its base.
		geometry.groups.sort((a, b) => {
			const ba = skin.batches[a.materialIndex!];
			const bb = skin.batches[b.materialIndex!];
			return ba.priority - bb.priority || ba.materialLayer - bb.materialLayer;
		});
		this.mesh = new THREE.SkinnedMesh(geometry, materials);
		this.mesh.frustumCulled = false;
		// Only the solid parts cast shadows. A glow or a see-through layer (a totem's aura, an
		// effect's flash) would darken the ground in a big square. A model with no solid parts (most
		// spell effects) isn't drawn into the shadows at all.
		const solid = skin.batches.map(batch => (m2.materials[batch.material]?.blend ?? 0) <= 1);
		this.mesh.castShadow = solid.some(x => x);
		this.mesh.onBeforeShadow = (_r, _o, _c, _sc, _g, depthMaterial, group) => {
			depthMaterial.depthWrite = solid[(group as unknown as { materialIndex?: number } | null)?.materialIndex ?? 0] ?? true;
		};
		this.mesh.onAfterShadow = (_r, _o, _c, _sc, _g, depthMaterial) => {
			depthMaterial.depthWrite = true;
		};

		// Three's skinning needs a skeleton, but its bones stay out of the scene: the skeleton's
		// matrices come straight from ours. The skinning is in the model's space (the mesh's), so
		// the mesh's own place in the world goes on top (the detached bind mode).
		const n = Math.max(1, m2.bones.length);
		const bones = Array.from({ length: n }, () => new THREE.Bone());
		this.places = new Float64Array(n * 16);
		this.skinning = new Float32Array(n * 16);
		for (let i = 0; i < n; i++) {
			this.places[i * 16] = this.places[i * 16 + 5] = this.places[i * 16 + 10] = this.places[i * 16 + 15] = 1;
			this.skinning[i * 16] = this.skinning[i * 16 + 5] = this.skinning[i * 16 + 10] = this.skinning[i * 16 + 15] = 1;
		}
		this.quats = new Float64Array(n * 4);
		const skeleton = new THREE.Skeleton(bones);
		const skinning = this.skinning;
		skeleton.update = function () {
			this.boneMatrices!.set(skinning);
			if (this.boneTexture) this.boneTexture.needsUpdate = true;
		};
		skins.set(skeleton, skinning);
		this.mesh.bindMode = THREE.DetachedBindMode;
		this.mesh.bind(skeleton, new THREE.Matrix4());
		this.root.add(this.mesh);
		this.order = boneOrder(m2);

		// The upper body is the lower spine (key bone 4) and everything on it, which an attack
		// moves while the legs keep running.
		this.upper = m2.bones.map(() => false);
		const spine = m2.keyBoneLookup[4];
		this.twistShare = new Float64Array(m2.bones.length);
		for (const [bone, share] of twistShares(m2, spine ?? -1, m2.keyBoneLookup[6] ?? -1)) this.twistShare[bone] = share;
		if (spine !== undefined && spine >= 0) {
			for (let i = 0; i < m2.bones.length; i++) {
				for (let p = i; p >= 0; p = m2.bones[p].parent) {
					if (p === spine) {
						this.upper[i] = true;
						break;
					}
				}
			}
		}
		const below = (keys: number[]) => {
			const out = m2.bones.map(() => false);
			for (let i = 0; i < m2.bones.length; i++) {
				for (let p = i; p >= 0; p = m2.bones[p].parent) {
					if (keys.includes(m2.bones[p].keyBoneId)) {
						out[i] = true;
						break;
					}
				}
			}
			return out;
		};
		this.fingers = { right: below(rightFingers), left: below(leftFingers) };
		this.billboard = Uint8Array.from(m2.bones, b => (b.flags & 0x8 ? 1 : 0));
		this.billboards = this.billboard.includes(1);
		this.pose0 = new Float32Array(m2.bones.length * 10);
		this.fadePose = new Float32Array(m2.bones.length * 10);
		this.base = this.playing(Anim.Stand, { loop: true });
	}

	// onLoad hears how long each model took to build (parse and set up, not the file fetch), for
	// the performance recorder.
	static onLoad?: (fdid: number, ms: number) => void;

	static async load(fdid: number): Promise<Model> {
		// A spell's effect plays for a second or two, and the raid's casters play one each cast. So a
		// copy that's done (recycle) waits to be used again, instead of a new one with its own
		// materials each time.
		const pooled = pool.get(fdid)?.pop();
		if (pooled) {
			pooled.reset();
			return pooled;
		}
		let p = parsed.get(fdid);
		if (!p) {
			p = (async () => {
				const data = await file(fdid);
				const m2 = new M2(data);
				if (m2.skins.length === 0) throw new Error(`model ${fdid} has no skin`);
				const skin = new Skin(await file(m2.skins[0]));
				return { m2, skin, geometry: buildGeometry(m2, skin), users: 0, indices: new Map() };
			})();
			parsed.set(fdid, p);
			p.catch(() => parsed.delete(fdid));
		}
		const shared = await p;
		const t0 = performance.now();
		const model = new Model(shared);
		model.fdid = fdid;
		Model.onLoad?.(fdid, performance.now() - t0);
		await Promise.all([model.loadAnim(model.base.seq), ...model.loading]);
		model.base.source = model.ready.get(model.base.seq);
		model.update(0);
		return model;
	}

	private makeMaterial(batch: Batch, index: number): BatchMaterial {
		const m2 = this.m2;
		const mat = m2.materials[batch.material] ?? { flags: 0, blend: 0 };
		const texIndex = m2.textureCombos[batch.textureCombo];
		const tex = m2.textures[texIndex];
		const unlit = (mat.flags & 0x1) !== 0;
		const params: THREE.MeshLambertMaterialParameters = {
			side: mat.flags & 0x4 ? THREE.DoubleSide : THREE.FrontSide,
			depthWrite: (mat.flags & 0x10) === 0 && mat.blend <= 1,
			depthTest: (mat.flags & 0x8) === 0,
			// Glows that add light get no fog: the client fogs them to black, which adds nothing.
			fog: (mat.flags & 0x2) === 0 && mat.blend !== 3 && mat.blend !== 4,
		};
		// The client's blend modes as GL blend factors (wowdev.wiki M2 materials).
		const custom = (src: THREE.BlendingSrcFactor, dst: THREE.BlendingDstFactor) => {
			params.transparent = true;
			params.blending = THREE.CustomBlending;
			params.blendSrc = src;
			params.blendDst = dst;
		};
		switch (mat.blend) {
			case 1: // alpha key
				params.alphaTest = 224 / 255;
				break;
			case 2: // alpha blend
				params.transparent = true;
				break;
			case 3: // add with no alpha
				custom(THREE.OneFactor, THREE.OneFactor);
				break;
			case 4: // add
				custom(THREE.SrcAlphaFactor, THREE.OneFactor);
				break;
			case 5: // modulate
				custom(THREE.DstColorFactor, THREE.ZeroFactor);
				break;
			case 6: // modulate 2x
				custom(THREE.DstColorFactor, THREE.SrcColorFactor);
				break;
			case 7: // blend add: the color counts by its alpha, and the alpha hides what's behind
				custom(THREE.OneFactor, THREE.OneMinusSrcAlphaFactor);
				params.premultipliedAlpha = true;
				break;
		}
		// See-through layers that don't write depth (glows, spell rings) often lie right on the ground,
		// like Fire Nova's ring. We draw them a little in front of what's at their depth, or the
		// ground hides them.
		if (params.depthWrite === false) {
			params.polygonOffset = true;
			params.polygonOffsetFactor = -4;
			params.polygonOffsetUnits = -16;
		}
		const material = unlit ? new THREE.MeshBasicMaterial(params) : new THREE.MeshLambertMaterial(params);
		const comb = combiner(batch.shader, batch.textureCount);
		const tex2 = batch.textureCount > 1 ? m2.textures[m2.textureCombos[batch.textureCombo + 1]] : undefined;
		if (comb.env0 || (comb.mode > 0 && tex2 && tex2.type === 0 && tex2.fdid > 0)) {
			const uniforms = {
				map2: { value: null as THREE.Texture | null },
				wowMode: { value: tex2 && tex2.type === 0 && tex2.fdid > 0 ? comb.mode : 0 },
				wowEnv0: { value: comb.env0 },
				wowEnv1: { value: comb.env1 },
				wowUv2: { value: comb.uv2 },
			};
			if (uniforms.wowMode.value > 0) {
				this.loading.push(
					texture(tex2!.fdid, (tex2!.flags & 1) !== 0, (tex2!.flags & 2) !== 0)
						.then(t => (uniforms.map2.value = t))
						.catch(e => console.warn(`texture for ${m2.name}`, e)),
				);
			}
			material.onBeforeCompile = shader => {
				Object.assign(shader.uniforms, uniforms);
				shader.vertexShader = shader.vertexShader
					.replace(
						'#include <common>',
						'#include <common>\nattribute vec2 uv1;\nvarying vec2 vWowUv2;\nvarying vec3 vWowNormal;\nvarying vec3 vWowView;',
					)
					.replace(
						'#include <project_vertex>',
						// An unlit material without bones has no objectNormal (static doodads, terrain.ts).
						'#include <project_vertex>\nvWowUv2 = uv1;\nvWowView = mvPosition.xyz;\n#if defined(LAMBERT) || defined(USE_SKINNING) || defined(USE_ENVMAP)\nvWowNormal = normalize(normalMatrix * objectNormal);\n#else\nvWowNormal = normalize(normalMatrix * normal);\n#endif',
					);
				shader.fragmentShader = shader.fragmentShader
					.replace(
						'#include <common>',
						'#include <common>\nuniform sampler2D map2;\nuniform int wowMode;\nuniform bool wowEnv0;\nuniform bool wowEnv1;\nuniform bool wowUv2;\nvarying vec2 vWowUv2;\nvarying vec3 vWowNormal;\nvarying vec3 vWowView;',
					)
					.replace('#include <map_fragment>', combinerFragment);
			};
			material.customProgramCacheKey = () => 'wow-combiner';
		}
		const bm: BatchMaterial = {
			material,
			batch,
			submesh: batch.submesh,
			textureType: tex?.type ?? -1,
			baseOpacity: 1,
			baseTransparent: material.transparent,
		};
		if (tex && tex.type === 0 && tex.fdid > 0) {
			this.loading.push(
				texture(tex.fdid, (tex.flags & 1) !== 0, (tex.flags & 2) !== 0)
					.then(t => {
						(material as THREE.MeshLambertMaterial).map = t;
						material.needsUpdate = true;
					})
					.catch(e => console.warn(`texture for ${m2.name}`, e)),
			);
		}
		void index;
		return bm;
	}

	// setTexture sets a replaceable texture: 1 the skin, 2 an item's own, 6 the hair, 11 to 13 a
	// creature's skins.
	setTexture(type: number, t: THREE.Texture) {
		// The model's slot says whether the texture repeats (hair does) or clamps (a cape). A
		// texture shared with a slot that wraps the other way gets its own copy.
		const slot = this.m2.textures.find(x => x.type === type);
		if (slot) {
			const wrapS = slot.flags & 1 ? THREE.RepeatWrapping : THREE.ClampToEdgeWrapping;
			const wrapT = slot.flags & 2 ? THREE.RepeatWrapping : THREE.ClampToEdgeWrapping;
			if (t.wrapS !== wrapS || t.wrapT !== wrapT) {
				// Every model that wraps it the same way shares the copy.
				const copies = rewrapped.get(t) ?? new Map<number, THREE.Texture>();
				rewrapped.set(t, copies);
				const key = wrapS * 65536 + wrapT;
				let copy = copies.get(key);
				if (!copy) {
					copy = t.clone();
					copy.wrapS = wrapS;
					copy.wrapT = wrapT;
					copy.needsUpdate = true;
					copies.set(key, copy);
				}
				t = copy;
			}
		}
		this.replaceable.set(type, t);
		for (const bm of this.mats) {
			if (bm.textureType === type) {
				(bm.material as THREE.MeshLambertMaterial).map = t;
				bm.material.needsUpdate = true;
			}
		}
	}

	// setGeosets shows only these geosets (by id). Without a list, every submesh shows.
	setGeosets(ids: number[] | undefined) {
		this.geosets = ids ? new Set(ids) : undefined;
		for (const bm of this.mats) {
			const sub = this.skin.submeshes[bm.submesh];
			bm.material.visible = !this.geosets || this.geosets.has(sub.id);
		}
	}

	// compact draws the parts that look alike in one draw call, once the model's look is set.
	//
	// A character's body comes in about 16 parts (skin, hair, face, each piece of armor's geoset),
	// and the renderer draws each part on its own. In a raid of 40 that was most of the frame. Parts
	// with the same material, textures and animated colors look the same, so we put their triangles
	// side by side and draw them as one. The hidden geosets drop out, so setGeosets can't show them
	// again afterwards.
	compact() {
		const g = this.mesh.geometry;
		const index = g.index;
		const mats = this.mesh.material as THREE.Material[];
		if (!index || !Array.isArray(mats)) return;
		const key = (i: number) => {
			const b = this.skin.batches[i];
			return [b.material, b.textureCombo, b.textureCount, b.shader, b.color, b.weightCombo, b.transformCombo, b.coordCombo, b.priority, b.materialLayer].join();
		};
		type Merged = { materialIndex: number; ranges: [number, number][] };
		const merged: Merged[] = [];
		const opaque = new Map<string, Merged>();
		let last: { key: string; m: Merged } | undefined;
		for (const group of g.groups) {
			const i = group.materialIndex ?? 0;
			if (!mats[i]?.visible) continue;
			const k = key(i);
			// The renderer sorts the solid parts itself, so any two alike can merge. See-through parts
			// draw in their order, so only neighbors merge.
			const solid = !mats[i].transparent;
			let m = solid ? opaque.get(k) : last?.key === k ? last.m : undefined;
			if (!m) {
				m = { materialIndex: i, ranges: [] };
				merged.push(m);
				if (solid) opaque.set(k, m);
			}
			m.ranges.push([group.start, group.count]);
			last = { key: k, m };
		}
		const src = index.array as Uint16Array | Uint32Array;
		let total = 0;
		for (const m of merged) for (const r of m.ranges) total += r[1];
		g.clearGroups();
		let at = 0;
		for (const m of merged) {
			const start = at;
			for (const [, c] of m.ranges) at += c;
			g.addGroup(start, at - start, m.materialIndex);
		}
		// Copies of the model with the same parts showing share the new index, so the crowd
		// (crowd.ts) can draw them together.
		const parts = merged.map(m => `${m.materialIndex}:${m.ranges.map(r => r.join('+')).join(',')}`).join('|');
		const indices = this.shared.indices;
		let shared = indices.get(parts);
		if (!shared) {
			const out = src instanceof Uint32Array ? new Uint32Array(total) : new Uint16Array(total);
			at = 0;
			for (const m of merged) {
				for (const [s, c] of m.ranges) {
					out.set(src.subarray(s, s + c), at);
					at += c;
				}
			}
			shared = new THREE.BufferAttribute(out, 1);
			indices.set(parts, shared);
		}
		g.setIndex(shared);
	}

	// cull lets the renderer skip the model when it's off screen, and skip its shadow when it's
	// outside the shadow's area. The bounds come from the model's resting pose, with room to spare
	// for its animations. Spell effects stay always drawn, since their bones fly far from the rest
	// pose.
	cull(margin = 1.5) {
		const g = this.mesh.geometry;
		g.computeBoundingSphere();
		const sphere = g.boundingSphere?.clone();
		if (!sphere) return;
		sphere.radius = sphere.radius * margin + 0.5;
		this.mesh.boundingSphere = sphere;
		this.mesh.frustumCulled = true;
	}

	// fullGeometry is the model's vertices and every part (all its geosets), shared by its copies.
	get fullGeometry(): THREE.BufferGeometry {
		return this.shared.geometry;
	}

	submeshIds(): number[] {
		return [...new Set(this.skin.submeshes.map(s => s.id))].sort((a, b) => a - b);
	}

	// attachPoint gives an object that follows an attachment point (1 right hand, 2 left hand,
	// 11 the helm, 18 the chest, ...), for weapons and effects. It's undefined when the model
	// doesn't have the point.
	attachPoint(id: number): THREE.Object3D | undefined {
		const cached = this.points.get(id);
		if (cached) return cached;
		const a = this.m2.attachment(id);
		if (!a) return undefined;
		const point = this.socket(a.bone, a.position);
		this.points.set(id, point);
		return point;
	}

	// restPose puts every bone back where it rests, for the model viewer.
	restPose() {
		for (let i = 0; i < this.m2.bones.length; i++) {
			const o = i * 16;
			const pivot = this.m2.bones[i].pivot;
			this.places.fill(0, o, o + 16);
			this.skinning.fill(0, o, o + 16);
			for (const d of [0, 5, 10, 15]) this.places[o + d] = this.skinning[o + d] = 1;
			this.places[o + 12] = pivot[0];
			this.places[o + 13] = pivot[1];
			this.places[o + 14] = pivot[2];
		}
		for (const socket of this.sockets) this.placeSocket(socket);
	}

	// socket gives an object that follows a bone, at a point in the model (where the bone's pivot
	// is at rest), for an attachment point or an emitter.
	socket(bone: number, at: ArrayLike<number>): THREE.Object3D {
		const b = this.m2.bones[bone] ? bone : 0;
		const pivot = this.m2.bones[b]?.pivot ?? [0, 0, 0];
		const object = new THREE.Object3D();
		object.matrixAutoUpdate = false;
		const s: Socket = { object, bone: b, x: at[0] - pivot[0], y: at[1] - pivot[1], z: at[2] - pivot[2] };
		this.sockets.push(s);
		this.placeSocket(s);
		this.mesh.add(object);
		return object;
	}

	// placeSocket puts a socket where its bone is now: the bone's place times T(offset).
	private placeSocket(s: Socket) {
		const m = this.places;
		const o = s.bone * 16;
		const e = s.object.matrix.elements;
		for (let j = 0; j < 12; j++) e[j] = m[o + j];
		e[12] = m[o] * s.x + m[o + 4] * s.y + m[o + 8] * s.z + m[o + 12];
		e[13] = m[o + 1] * s.x + m[o + 5] * s.y + m[o + 9] * s.z + m[o + 13];
		e[14] = m[o + 2] * s.x + m[o + 6] * s.y + m[o + 10] * s.z + m[o + 14];
		e[15] = 1;
		s.object.matrixWorldNeedsUpdate = true;
	}

	// preloadAnims loads the keys of these animations (or the ones the model plays in their
	// place), so the first time one plays it doesn't hold a frozen pose while its file loads.
	async preloadAnims(anims: number[]) {
		const seqs = new Set<number>();
		for (const a of anims) {
			const p = this.playing(a, {});
			for (const v of this.variations(this.m2.sequenceFor(p.anim) >= 0 ? this.m2.sequenceFor(p.anim) : p.seq)) seqs.add(this.m2.resolve(v));
			seqs.add(p.seq);
		}
		if (this.closed.right || this.closed.left || anims.includes(Anim.HandsClosed)) {
			const hands = this.m2.sequenceFor(Anim.HandsClosed);
			if (hands >= 0) seqs.add(this.m2.resolve(hands));
		}
		await Promise.all([...seqs].map(seq => this.loadAnim(seq)));
	}

	has(anim: number): boolean {
		return this.m2.sequenceFor(anim) >= 0;
	}

	// play switches the whole body to an animation, fading from the pose before. It gives the
	// sequence it picked (a one-shot picks one of its variations at random).
	play(anim: number, opts: PlayOptions = {}): number {
		if (!opts.restart && this.base.anim === anim && !this.base.done) {
			this.base.speed = opts.speed ?? 1;
			return this.base.seq;
		}
		const next = this.playing(anim, opts);
		if (
			this.m2.sequences[next.seq]?.variation === this.m2.sequences[this.base.seq]?.variation &&
			next.seq === this.base.seq &&
			!opts.restart &&
			!this.base.done
		) {
			this.base.anim = anim;
			this.base.speed = next.speed;
			return this.base.seq;
		}
		if (this.posed) {
			this.fadePose.set(this.pose0);
			this.fadeTotal = this.fadeLeft = opts.fade ?? this.m2.sequences[next.seq]?.blendIn ?? 150;
		}
		this.base = next;
		this.ensure(next);
		return next.seq;
	}

	// playUpper plays a one-shot animation on the upper body only (an attack while running).
	playUpper(anim: number, opts: PlayOptions = {}): number {
		this.overlay = this.playing(anim, { ...opts, loop: false });
		this.ensure(this.overlay);
		return this.overlay.seq;
	}

	// baseToUpper moves a one-shot playing on the whole body (a swing) to the upper body, where it
	// plays on while the legs start to run.
	baseToUpper() {
		if (this.base.loop || this.base.done) return;
		this.overlay = { ...this.base, started: this.globalTime - 1000 };
	}

	// eventTimes gives the times (ms) of an event ($HIT...) in a sequence.
	eventTimes(id: string, seq: number): number[] {
		return (this.eventsOf(seq) ?? []).filter(e => e.id === id).map(e => e.time);
	}

	private eventsOf(seq: number): { id: string; data: number; time: number }[] | null {
		let list = this.events.get(seq);
		if (list !== undefined) return list;
		const source = this.m2.inline(seq) ? this.m2.data : this.ready.get(seq);
		if (!source) return null;
		list = [];
		for (const ev of this.m2.events as M2Event[]) {
			for (const time of this.m2.eventTimes(ev, seq, source)) list.push({ id: ev.id, data: ev.data, time });
		}
		this.events.set(seq, list);
		return list;
	}

	get baseAnim() {
		return this.base.anim;
	}

	get baseDone() {
		return this.base.done;
	}

	get overlayActive() {
		return !!this.overlay && !this.overlay.done;
	}

	// closeHands curls a hand's fingers, for holding a weapon.
	closeHands(right: boolean, left: boolean) {
		this.closed = { right, left };
		const seq = this.m2.sequenceFor(Anim.HandsClosed);
		if ((right || left) && seq >= 0) this.loadAnim(this.m2.resolve(seq));
	}

	private playing(anim: number, opts: PlayOptions): Playing {
		let a = anim;
		let seq = this.m2.sequenceFor(a);
		for (let guard = 0; seq < 0 && guard < 8 && fallbacks[a] !== undefined; guard++) {
			a = fallbacks[a];
			seq = this.m2.sequenceFor(a);
		}
		if (seq < 0) seq = 0;
		if (opts.loop === false) seq = this.variation(seq);
		seq = this.m2.resolve(seq);
		return { anim, started: this.globalTime, seq, time: 0, speed: opts.speed ?? 1, loop: opts.loop ?? true, source: this.ready.get(seq), done: false };
	}

	// variations lists a sequence and its other variations (two different swings, say).
	private variations(seq: number): number[] {
		const out: number[] = [];
		for (let s = seq; s >= 0 && s < this.m2.sequences.length && out.length < 16; s = this.m2.sequences[s].variationNext) out.push(s);
		return out;
	}

	// variation picks one of a sequence's variations at random, by their frequencies.
	private variation(seq: number): number {
		const list = this.variations(seq);
		if (list.length < 2) return seq;
		const weights = list.map(s => Math.max(0, this.m2.sequences[s].frequency));
		const total = weights.reduce((a, b) => a + b, 0);
		if (total <= 0) return list[Math.floor(Math.random() * list.length)];
		let roll = Math.random() * total;
		for (let i = 0; i < list.length; i++) {
			roll -= weights[i];
			if (roll < 0) return list[i];
		}
		return seq;
	}

	private ensure(p: Playing) {
		if (p.source) return;
		this.loadAnim(p.seq).then(source => {
			p.source = source;
		});
	}

	private loadAnim(seq: number): Promise<DataView | undefined> {
		let p = this.anims.get(seq);
		if (!p) {
			if (this.m2.inline(seq)) p = Promise.resolve(this.m2.data);
			else {
				const fdid = this.m2.animFileFor(seq);
				const chunked = (this.m2.flags & 0x200000) !== 0 || this.m2.skeleton > 0;
				p = fdid
					? file(fdid)
							.then(buf => animData(buf, chunked))
							.catch(e => {
								console.warn(`animation ${this.m2.sequences[seq]?.id} of ${this.m2.name}`, e);
								return undefined;
							})
					: Promise.resolve(undefined);
			}
			p.then(v => {
				if (v) this.ready.set(seq, v);
			});
			this.anims.set(seq, p);
		}
		return p;
	}

	private advance(p: Playing, dt: number, events: false | 'base' | 'upper') {
		// Until its keys arrive, an animation holds at its start.
		if (!p.source || p.done) return;
		const duration = this.m2.sequences[p.seq]?.duration ?? 0;
		const before = p.time;
		p.time += dt * 1000 * p.speed;
		let wrapped = false;
		if (duration <= 0) {
			p.time = 0;
			if (!p.loop) p.done = true;
		} else if (p.time >= duration) {
			if (p.loop) {
				p.time %= duration;
				wrapped = true;
			} else {
				p.time = duration - 1;
				p.done = true;
			}
		}
		if (!events || !this.onEvent) return;
		const list = this.eventsOf(p.seq);
		if (!list?.length) return;
		// The events between the last frame and this one. An event at 0 fires when the animation
		// starts.
		const from = before === 0 && dt > 0 ? -1 : before;
		for (const e of list) {
			const passed = wrapped ? e.time > from || e.time <= p.time : e.time > from && e.time <= p.time;
			if (passed) this.onEvent(e.id, e.data, events === 'upper');
		}
	}

	// update moves the animations on by dt seconds and poses the bones.
	update(dt: number) {
		const dtMs = dt * 1000;
		this.globalTime += dtMs;
		this.advance(this.base, dt, 'base');
		this.fadeLeft = Math.max(0, this.fadeLeft - dtMs);
		const overlay = this.overlay;
		let ow = 0;
		if (overlay) {
			this.advance(overlay, dt, 'upper');
			if (overlay.done) this.overlay = undefined;
			else if (overlay.source) {
				// The upper body fades into the attack and back out of it.
				const duration = this.m2.sequences[overlay.seq]?.duration ?? 0;
				ow = Math.min(1, (this.globalTime - overlay.started) / 120, (duration - overlay.time) / 150);
				ow = Math.max(0, ow);
			}
		}
		const handsSeq = this.closed.right || this.closed.left ? this.m2.resolve(this.m2.sequenceFor(Anim.HandsClosed)) : -1;
		const handsSource = handsSeq >= 0 ? this.ready.get(handsSeq) : undefined;
		const s = this.scratch;
		const m2 = this.m2;
		const base = this.base;
		// Until the new animation's keys arrive, the bones keep the pose they had.
		const fadeW = !base.source && this.posed ? 1 : this.fadeTotal > 0 && this.posed ? this.fadeLeft / this.fadeTotal : 0;
		const camera = this.billboards ? Model.camera : undefined;
		if (camera) {
			this.mesh.updateWorldMatrix(true, false);
			camera.getWorldQuaternion(billboardTmp.q1);
		}
		const places = this.places;
		const quats = this.quats;
		const skin = this.skinning;
		for (let k = 0; k < this.order.length; k++) {
			const i = this.order[k];
			const b = m2.bones[i];
			const finger = handsSource && ((this.closed.right && this.fingers.right[i]) || (this.closed.left && this.fingers.left[i]));
			if (finger) {
				this.at[0] = 0;
				this.pose(b.translation, b.rotation, b.scale, handsSeq, handsSource, s.t, s.r, s.s);
			} else {
				this.at[0] = base.time;
				this.pose(b.translation, b.rotation, b.scale, base.seq, base.source, s.t, s.r, s.s);
				if (fadeW > 0) {
					const o = i * 10;
					const f = this.fadePose;
					s.t2[0] = f[o];
					s.t2[1] = f[o + 1];
					s.t2[2] = f[o + 2];
					s.r2[0] = f[o + 3];
					s.r2[1] = f[o + 4];
					s.r2[2] = f[o + 5];
					s.r2[3] = f[o + 6];
					s.s2[0] = f[o + 7];
					s.s2[1] = f[o + 8];
					s.s2[2] = f[o + 9];
					blendWeight[0] = fadeW;
					blend(s);
				}
				if (ow > 0 && overlay && this.upper[i]) {
					this.at[0] = overlay.time;
					this.pose(b.translation, b.rotation, b.scale, overlay.seq, overlay.source, s.t2, s.r2, s.s2);
					blendWeight[0] = ow;
					blend(s);
				}
			}
			const o = i * 10;
			const p0 = this.pose0;
			p0[o] = s.t[0];
			p0[o + 1] = s.t[1];
			p0[o + 2] = s.t[2];
			p0[o + 3] = s.r[0];
			p0[o + 4] = s.r[1];
			p0[o + 5] = s.r[2];
			p0[o + 6] = s.r[3];
			p0[o + 7] = s.s[0];
			p0[o + 8] = s.s[1];
			p0[o + 9] = s.s[2];
			const parent = b.parent >= 0 ? m2.bones[b.parent].pivot : undefined;
			// The rotation goes straight into quats. Through a Three quaternion's set, each of its four
			// numbers was boxed, every bone. Only the twist and the billboards need one.
			const q = i * 4;
			quats[q] = s.r[0];
			quats[q + 1] = s.r[1];
			quats[q + 2] = s.r[2];
			quats[q + 3] = s.r[3];
			const share = this.twist !== 0 ? this.twistShare[i] : 0;
			const turned = share !== 0 || (camera && this.billboard[i]);
			if (turned) boneQ.fromArray(quats, q);
			if (share !== 0) {
				// The twist is around the model's up (y, since the bones are in Three's frame). The
				// parent bones lean in the run, so we turn that axis into the parent's frame.
				parentQ.identity();
				for (let p = b.parent; p >= 0; p = m2.bones[p].parent) parentQ.premultiply(twistQ.fromArray(quats, p * 4));
				twistAxis.copy(upAxis).applyQuaternion(parentQ.invert());
				boneQ.premultiply(twistQ.setFromAxisAngle(twistAxis, this.twist * share));
			}
			if (camera && this.billboard[i]) {
				// A billboard turns its +x (the model's front) to the camera, with its up staying the
				// camera's up: its rotation is the parent's in the world, undone, then the camera's.
				parentWorld.copy(this.mesh.matrixWorld);
				if (b.parent >= 0) parentWorld.multiply(billboardTmp.m.fromArray(places, b.parent * 16));
				parentWorld.decompose(billboardTmp.v, billboardTmp.q2, billboardTmp.s);
				boneQ.copy(billboardTmp.q2.invert()).multiply(billboardTmp.q1).multiply(faceFront);
			}
			if (turned) boneQ.toArray(quats, q);
			const at = placeAt;
			at[0] = b.pivot[0] - (parent?.[0] ?? 0) + s.t[0];
			at[1] = b.pivot[1] - (parent?.[1] ?? 0) + s.t[1];
			at[2] = b.pivot[2] - (parent?.[2] ?? 0) + s.t[2];
			at[3] = quats[q];
			at[4] = quats[q + 1];
			at[5] = quats[q + 2];
			at[6] = quats[q + 3];
			at[7] = s.s[0];
			at[8] = s.s[1];
			at[9] = s.s[2];
			place(places, i * 16, b.parent >= 0 ? b.parent * 16 : -1);
			// The skinning matrix is the place times T(-pivot).
			const m = i * 16;
			for (let j = 0; j < 12; j++) skin[m + j] = places[m + j];
			skin[m + 12] = places[m + 12] - places[m] * b.pivot[0] - places[m + 4] * b.pivot[1] - places[m + 8] * b.pivot[2];
			skin[m + 13] = places[m + 13] - places[m + 1] * b.pivot[0] - places[m + 5] * b.pivot[1] - places[m + 9] * b.pivot[2];
			skin[m + 14] = places[m + 14] - places[m + 2] * b.pivot[0] - places[m + 6] * b.pivot[1] - places[m + 10] * b.pivot[2];
			skin[m + 15] = 1;
		}
		for (const socket of this.sockets) this.placeSocket(socket);
		if (base.source) this.posed = true;
		this.updateColors();
		this.particles?.update(dt);
		this.ribbons?.update(dt);
	}

	// trackKeys gives a track's keys for a sequence. A track on a global sequence runs on its own
	// clock: then keyTime is the time to sample at, else it's -1 (sample at the animation's time).
	private keyTime = -1;
	private trackKeys(track: Track, seq: number, source: DataView | undefined): Keys | null {
		if (track.globalSeq !== 0xffff) {
			const loop = this.m2.globalLoops[track.globalSeq] || 1;
			this.keyTime = this.globalTime % loop;
			return track.get(0, this.m2.data);
		}
		this.keyTime = -1;
		// A one-slot track keeps its keys in the model file.
		return track.get(seq, track.times.length === 1 ? this.m2.data : source);
	}

	// sampleTrack samples a track at the time in `at` (a Float64Array, so the time isn't boxed on
	// each call).
	private readonly at = new Float64Array(1);
	private sampleTrack(track: Track, seq: number, source: DataView | undefined, out: Vec): boolean {
		const keys = this.trackKeys(track, seq, source);
		if (!keys) return false;
		sampleTime[0] = this.keyTime >= 0 ? this.keyTime : this.at[0];
		sample(keys, track.width, out, track.interpolation);
		return true;
	}

	private pose(tr: Track, rot: Track, sc: Track, seq: number, source: DataView | undefined, t: Vec, r: Vec, s: Vec) {
		t[0] = t[1] = t[2] = 0;
		r[0] = r[1] = r[2] = 0;
		r[3] = 1;
		s[0] = s[1] = s[2] = 1;
		this.sampleTrack(tr, seq, source, t);
		this.sampleTrack(rot, seq, source, r);
		this.sampleTrack(sc, seq, source, s);
	}

	// trackValue samples a track (an emitter's rate, say) at the current animation's time. It's
	// false when the track has no keys there.
	trackValue(track: Track, out: Vec): boolean {
		this.at[0] = this.base.time;
		return this.sampleTrack(track, this.base.seq, this.base.source, out);
	}

	// emit turns the particles and ribbons on or off (an effect that fades out stops emitting).
	set emitting(on: boolean) {
		if (this.particles) this.particles.emitting = on;
		if (this.ribbons) this.ribbons.emitting = on;
	}

	// loadEffectTextures loads the textures of the model's particles and ribbons, which otherwise
	// load the first time the model shows them. A totem's fire is a 2048 by 2048 texture that took
	// over 100ms to decode at the first Searing Totem.
	loadEffectTextures(): Promise<unknown> {
		const m2 = this.m2;
		const loads: Promise<unknown>[] = [];
		for (const e of m2.particles) {
			const tex = m2.textures[e.flags & 0x10000000 ? e.texture & 0x1f : e.texture];
			if (tex && tex.type === 0 && tex.fdid > 0) loads.push(texture(tex.fdid, false, false).catch(() => undefined));
		}
		for (const r of m2.ribbons) {
			const tex = m2.textures[r.textures[0] ?? -1];
			if (tex && tex.type === 0 && tex.fdid > 0) loads.push(texture(tex.fdid, (tex.flags & 1) !== 0, (tex.flags & 2) !== 0).catch(() => undefined));
		}
		return Promise.all(loads);
	}

	// enableParticles starts the model's particle emitters, with their particles in `world`.
	enableParticles(world: THREE.Object3D) {
		if (!this.particles && this.m2.particles.length > 0) this.particles = new Particles(this, world);
		if (!this.ribbons && this.m2.ribbons.length > 0) this.ribbons = new Ribbons(this, world);
	}

	// updateColors applies animated colors and transparency to the batches that have them.
	private readonly colorOut = new Float64Array(3);
	private updateColors() {
		const out = this.colorOut;
		const m2 = this.m2;
		const base = this.base;
		for (const bm of this.mats) {
			const { batch } = bm;
			let alpha = 1;
			if (batch.color >= 0 && m2.colors[batch.color]) {
				const c = m2.colors[batch.color];
				this.at[0] = base.time;
				if (this.sampleTrack(c.alpha, base.seq, base.source, out)) alpha *= out[0] / 32767;
				// M2 colors are gamma values, like the textures.
				if (bm.material.color && this.sampleTrack(c.color, base.seq, base.source, out))
					bm.material.color.setRGB(out[0], out[1], out[2], THREE.SRGBColorSpace);
			}
			const wi = m2.transparencyLookup[batch.weightCombo];
			if (wi !== undefined && m2.textureWeights[wi]) {
				this.at[0] = base.time;
				if (this.sampleTrack(m2.textureWeights[wi], base.seq, base.source, out)) alpha *= out[0] / 32767;
			}
			if (alpha < 0.999 || bm.material.transparent) {
				bm.material.opacity = Math.max(0, Math.min(1, alpha));
				if (alpha < 0.999) bm.material.transparent = true;
			}
		}
	}

	// The model's file, for the pool of copies to use again.
	private fdid = 0;

	// recycle takes a spell effect's model out of the scene and keeps it to be used again by the
	// next Model.load of the same file. Nothing may hold on to it after.
	recycle() {
		const list = pool.get(this.fdid) ?? [];
		pool.set(this.fdid, list);
		if (list.length >= 48) {
			this.dispose();
			return;
		}
		this.crowd?.remove(this);
		this.root.removeFromParent();
		this.particles?.reset();
		this.ribbons?.reset();
		list.push(this);
	}

	// reset puts a model used again back the way Model.load gives a new one.
	private reset() {
		const r = this.root;
		r.position.set(0, 0, 0);
		r.quaternion.identity();
		r.scale.set(1, 1, 1);
		r.visible = true;
		this.mesh.visible = true;
		for (const bm of this.mats) {
			bm.material.opacity = 1;
			bm.material.transparent = bm.baseTransparent;
			bm.material.color?.setRGB(1, 1, 1);
		}
		this.onEvent = undefined;
		this.globalTime = 0;
		this.twist = 0;
		this.closed = { right: false, left: false };
		this.overlay = undefined;
		this.fadeLeft = this.fadeTotal = 0;
		this.posed = false;
		this.base = this.playing(Anim.Stand, { loop: true });
		this.particles?.reset();
		this.ribbons?.reset();
		this.update(0);
	}

	dispose() {
		this.crowd?.remove(this);
		this.particles?.dispose();
		this.ribbons?.dispose();
		this.root.removeFromParent();
		const g = this.mesh.geometry;
		// While other copies still draw the shared vertices, we free only this copy's own parts, so
		// the renderer keeps the shared ones on the GPU.
		if (--this.shared.users > 0) {
			for (const name of Object.keys(g.attributes)) g.deleteAttribute(name);
			if (g.index === this.shared.geometry.index || [...this.shared.indices.values()].includes(g.index!)) g.setIndex(null);
		}
		g.dispose();
		for (const bm of this.mats) bm.material.dispose();
	}
}

// A model file read once: its parsed data and vertices, shared by every copy of the model (a
// raid of the same race, every Searing bolt), with how many copies are alive.
// indices holds the index of each set of parts that compact made, by the parts.
type Parsed = { m2: M2; skin: Skin; geometry: THREE.BufferGeometry; users: number; indices: Map<string, THREE.BufferAttribute> };
const parsed = new Map<number, Promise<Parsed>>();

// pool holds the effect models done playing, by file, to be used again (recycle).
const pool = new Map<number, Model[]>();

// rewrapped holds the copies of textures that a model wraps another way than the texture does.
const rewrapped = new WeakMap<THREE.Texture, Map<number, THREE.Texture>>();

// twistShares spreads a strafe's twist over the bones from the lower spine up to the head. The
// lower spine takes 55%, the bones between it and the head (chest and neck) 25% together, and the
// head the last 20%. On a strafe that leaves the chest about 30 degrees toward where you run. A model without a head on its spine twists the lower spine alone.
function twistShares(m2: M2, spine: number, head: number): [number, number][] {
	if (spine < 0) return [];
	const between: number[] = [];
	let found = false;
	for (let p = head >= 0 ? m2.bones[head].parent : -1; p >= 0; p = m2.bones[p].parent) {
		if (p === spine) {
			found = true;
			break;
		}
		between.push(p);
	}
	if (!found) return [[spine, 1]];
	if (between.length === 0) return [
		[spine, 0.6],
		[head, 0.4],
	];
	return [[spine, 0.55], ...between.reverse().map(b => [b, 0.25 / between.length] as [number, number]), [head, 0.2]];
}

// combiner works out how a batch mixes its textures, from its shader id and texture count
// (wowdev.wiki M2, "Shaders"): mode 0 one texture, 1 modulate, 2 modulate 2x, 3 add by the
// second's alpha, 4 modulate 2x by the first's alpha, 5 add by both alphas. env0 and env1 map a
// texture by the view (a sphere map, the shine on armor), and uv2 puts the second on the second
// UV set.
function combiner(shader: number, count: number): { mode: number; env0: boolean; env1: boolean; uv2: boolean } {
	if (shader & 0x8000) {
		// The client's shader table. The ones items use put the second texture on an env map.
		switch (shader & 0x7fff) {
			case 0:
				return { mode: 4, env0: false, env1: true, uv2: false };
			case 1:
				return { mode: 3, env0: false, env1: true, uv2: false };
			case 2:
				return { mode: 5, env0: false, env1: true, uv2: false };
		}
		return { mode: 0, env0: false, env1: false, uv2: false };
	}
	const env0 = (shader & 0x80) !== 0;
	if (count < 2) return { mode: 0, env0, env1: false, uv2: false };
	const modes: Record<number, number> = { 0: 0, 1: 1, 2: 1, 3: 3, 4: 2, 5: 1, 6: 2, 7: 3 };
	return { mode: modes[shader & 7] ?? 1, env0, env1: (shader & 0x8) !== 0, uv2: (shader & 0x4000) !== 0 };
}

const combinerFragment = /* glsl */ `
#ifdef USE_MAP
	vec3 wowR = reflect(normalize(vWowView), normalize(vWowNormal));
	float wowM = 2.0 * sqrt(wowR.x * wowR.x + wowR.y * wowR.y + (wowR.z + 1.0) * (wowR.z + 1.0));
	vec2 wowEnvUv = vec2(wowR.x / wowM + 0.5, 0.5 - wowR.y / wowM);
	vec4 wowT0 = texture2D(map, wowEnv0 ? wowEnvUv : vMapUv);
	vec4 wowC = wowT0;
	if (wowMode > 0) {
		vec4 wowT1 = texture2D(map2, wowEnv1 ? wowEnvUv : (wowUv2 ? vWowUv2 : vMapUv));
		if (wowMode == 1) wowC.rgb = wowT0.rgb * wowT1.rgb;
		else if (wowMode == 2) wowC.rgb = wowT0.rgb * wowT1.rgb * 2.0;
		else if (wowMode == 3) wowC.rgb = wowT0.rgb + wowT1.rgb * wowT1.a;
		else if (wowMode == 4) wowC.rgb = mix(wowT0.rgb, wowT0.rgb * wowT1.rgb * 2.0, wowT0.a);
		else if (wowMode == 5) wowC.rgb = wowT0.rgb + wowT1.rgb * wowT1.a * wowT0.a;
	}
	diffuseColor *= wowC;
#endif
`;

const billboardTmp = { q1: new THREE.Quaternion(), q2: new THREE.Quaternion(), v: new THREE.Vector3(), s: new THREE.Vector3(), m: new THREE.Matrix4() };

// placeAt holds the bone's move, rotation and scale for place. Passed as numbers, Chrome boxed
// each one, which was 70 bytes of garbage a bone.
const placeAt = new Float64Array(10);

// place sets a bone's place (at o) to its parent's (at parent, or the model when -1) times
// T(x, y, z) R(q) S(sx, sy, sz) from placeAt, the way Three's Matrix4.compose builds a matrix.
function place(m: Float64Array, o: number, parent: number) {
	const a = placeAt;
	const x = a[0];
	const y = a[1];
	const z = a[2];
	const qx = a[3];
	const qy = a[4];
	const qz = a[5];
	const qw = a[6];
	const sx = a[7];
	const sy = a[8];
	const sz = a[9];
	const x2 = qx + qx;
	const y2 = qy + qy;
	const z2 = qz + qz;
	const xx = qx * x2;
	const xy = qx * y2;
	const xz = qx * z2;
	const yy = qy * y2;
	const yz = qy * z2;
	const zz = qz * z2;
	const wx = qw * x2;
	const wy = qw * y2;
	const wz = qw * z2;
	const l0 = (1 - (yy + zz)) * sx;
	const l1 = (xy + wz) * sx;
	const l2 = (xz - wy) * sx;
	const l4 = (xy - wz) * sy;
	const l5 = (1 - (xx + zz)) * sy;
	const l6 = (yz + wx) * sy;
	const l8 = (xz + wy) * sz;
	const l9 = (yz - wx) * sz;
	const l10 = (1 - (xx + yy)) * sz;
	if (parent < 0) {
		m[o] = l0;
		m[o + 1] = l1;
		m[o + 2] = l2;
		m[o + 4] = l4;
		m[o + 5] = l5;
		m[o + 6] = l6;
		m[o + 8] = l8;
		m[o + 9] = l9;
		m[o + 10] = l10;
		m[o + 12] = x;
		m[o + 13] = y;
		m[o + 14] = z;
	} else {
		const p = parent;
		for (let r = 0; r < 3; r++) {
			const a = m[p + r];
			const b = m[p + 4 + r];
			const c = m[p + 8 + r];
			m[o + r] = a * l0 + b * l1 + c * l2;
			m[o + 4 + r] = a * l4 + b * l5 + c * l6;
			m[o + 8 + r] = a * l8 + b * l9 + c * l10;
			m[o + 12 + r] = a * x + b * y + c * z + m[p + 12 + r];
		}
	}
	m[o + 3] = m[o + 7] = m[o + 11] = 0;
	m[o + 15] = 1;
}

// boneOrder lists the bones so that each parent comes before its children.
function boneOrder(m2: M2): Int32Array {
	const out: number[] = [];
	const done = new Uint8Array(m2.bones.length);
	const visit = (i: number) => {
		if (done[i]) return;
		done[i] = 1;
		const p = m2.bones[i].parent;
		if (p >= 0 && p < m2.bones.length) visit(p);
		out.push(i);
	};
	for (let i = 0; i < m2.bones.length; i++) visit(i);
	return Int32Array.from(out);
}
// faceFront turns the model's front (+x) to the camera's back (+z), which faces the viewer.
const faceFront = new THREE.Quaternion().setFromAxisAngle(new THREE.Vector3(0, 1, 0), -Math.PI / 2);

// blendWeight is blend's weight, in a Float64Array so it isn't boxed on each call.
const blendWeight = new Float64Array(1);

function blend(s: { t: Vec; r: Vec; s: Vec; t2: Vec; r2: Vec; s2: Vec }) {
	// w is the old pose's share.
	const w = blendWeight[0];
	for (let k = 0; k < 3; k++) {
		s.t[k] = s.t[k] * (1 - w) + s.t2[k] * w;
		s.s[k] = s.s[k] * (1 - w) + s.s2[k] * w;
	}
	let dot = 0;
	for (let k = 0; k < 4; k++) dot += s.r[k] * s.r2[k];
	const sign = dot < 0 ? -1 : 1;
	let len = 0;
	for (let k = 0; k < 4; k++) {
		s.r[k] = s.r[k] * (1 - w) + sign * s.r2[k] * w;
		len += s.r[k] * s.r[k];
	}
	len = Math.sqrt(len) || 1;
	for (let k = 0; k < 4; k++) s.r[k] /= len;
}

// buildGeometry makes one geometry from the skin's vertices, with a group per batch.
function buildGeometry(m2: M2, skin: Skin): THREE.BufferGeometry {
	const n = skin.vertices.length;
	const pos = new Float32Array(n * 3);
	const nor = new Float32Array(n * 3);
	const uv = new Float32Array(n * 2);
	const uv2 = new Float32Array(n * 2);
	const si = new Uint16Array(n * 4);
	const sw = new Float32Array(n * 4);
	for (let i = 0; i < n; i++) {
		const v = skin.vertices[i];
		pos.set(m2.positions.subarray(v * 3, v * 3 + 3), i * 3);
		nor.set(m2.normals.subarray(v * 3, v * 3 + 3), i * 3);
		uv.set(m2.uvs.subarray(v * 2, v * 2 + 2), i * 2);
		uv2.set(m2.uvs2.subarray(v * 2, v * 2 + 2), i * 2);
		let total = 0;
		for (let k = 0; k < 4; k++) total += m2.boneWeights[v * 4 + k];
		for (let k = 0; k < 4; k++) {
			si[i * 4 + k] = m2.bones.length > 0 ? m2.boneIndices[v * 4 + k] : 0;
			sw[i * 4 + k] = total > 0 ? m2.boneWeights[v * 4 + k] / total : k === 0 ? 1 : 0;
		}
	}
	const g = new THREE.BufferGeometry();
	g.setAttribute('position', new THREE.BufferAttribute(pos, 3));
	g.setAttribute('normal', new THREE.BufferAttribute(nor, 3));
	g.setAttribute('uv', new THREE.BufferAttribute(uv, 2));
	g.setAttribute('uv1', new THREE.BufferAttribute(uv2, 2));
	g.setAttribute('skinIndex', new THREE.BufferAttribute(si, 4));
	g.setAttribute('skinWeight', new THREE.BufferAttribute(sw, 4));
	g.setIndex(new THREE.BufferAttribute(n > 65535 ? new Uint32Array(skin.indices) : skin.indices, 1));
	skin.batches.forEach((batch, i) => {
		const sub = skin.submeshes[batch.submesh];
		if (sub) g.addGroup(sub.indexStart, sub.indexCount, i);
	});
	return g;
}
