// A WoW model drawn with Three.js: its mesh, materials, skeleton and animations.
//
// Each bone is a Three.js bone. WoW moves a bone around its pivot: T(pivot) T(move) R S T(-pivot),
// on top of its parent. A Three.js bone with the position (pivot - parent's pivot + move) and a
// bind inverse of T(-pivot) gives the same result, so the skinning is Three.js's own.

import * as THREE from 'three';
import { file, texture } from './files';
import { animData, Batch, Keys, M2, sample, Skin, Track } from './m2';
import { Particles } from './particles';

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
	[Anim.ReadySpellDirected]: Anim.SpellPrecast,
	[Anim.ReadySpellOmni]: Anim.SpellPrecast,
	[Anim.SpellCastDirected]: Anim.SpellCast,
	[Anim.SpellCastOmni]: Anim.SpellCast,
	[Anim.SpellCastArea]: Anim.SpellCast,
	[Anim.ChannelCastDirected]: Anim.SpellCastDirected,
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
	[Anim.Dodge]: Anim.Stand,
	[Anim.Parry1H]: Anim.ParryUnarmed,
	[Anim.ParryUnarmed]: Anim.Stand,
	[Anim.BattleRoar]: Anim.SpellCast,
	[Anim.Birth]: Anim.Stand,
};

export type PlayOptions = { loop?: boolean; speed?: number; fade?: number; restart?: boolean };

type Playing = {
	anim: number;
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
};

// Bones that move the fingers: right hand 8 to 12, left hand 13 to 17 (key bone ids).
const rightFingers = [8, 9, 10, 11, 12];
const leftFingers = [13, 14, 15, 16, 17];

export class Model {
	readonly root = new THREE.Group();
	readonly mesh: THREE.SkinnedMesh;
	readonly bones: THREE.Bone[];
	private readonly mats: BatchMaterial[] = [];
	private readonly replaceable = new Map<number, THREE.Texture>();
	private readonly anims = new Map<number, Promise<DataView | undefined>>();
	private readonly ready = new Map<number, DataView>();
	private base: Playing;
	private fadeFrom?: Playing;
	private fadeLeft = 0;
	private fadeTotal = 0;
	private overlay?: Playing;
	private overlayFade = 0;
	private upper: boolean[];
	private globalTime = 0;
	private closed = { right: false, left: false };
	private scratch = { t: [0, 0, 0], r: [0, 0, 0, 1], s: [1, 1, 1], t2: [0, 0, 0], r2: [0, 0, 0, 1], s2: [1, 1, 1] };
	geosets?: Set<number>;
	particles?: Particles;

	private constructor(
		readonly m2: M2,
		readonly skin: Skin,
	) {
		const geometry = buildGeometry(m2, skin);
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
		this.mesh.castShadow = true;

		this.bones = m2.bones.map(() => new THREE.Bone());
		m2.bones.forEach((b, i) => {
			const bone = this.bones[i];
			const parent = b.parent >= 0 ? m2.bones[b.parent] : undefined;
			bone.position.set(b.pivot[0] - (parent?.pivot[0] ?? 0), b.pivot[1] - (parent?.pivot[1] ?? 0), b.pivot[2] - (parent?.pivot[2] ?? 0));
			if (b.parent >= 0) this.bones[b.parent].add(bone);
			else this.mesh.add(bone);
		});
		const inverses = m2.bones.map(b => new THREE.Matrix4().makeTranslation(-b.pivot[0], -b.pivot[1], -b.pivot[2]));
		if (this.bones.length === 0) {
			// A model with no bones still needs one for the skinning.
			const bone = new THREE.Bone();
			this.mesh.add(bone);
			this.bones.push(bone);
			inverses.push(new THREE.Matrix4());
		}
		this.mesh.bind(new THREE.Skeleton(this.bones, inverses), new THREE.Matrix4());
		this.root.add(this.mesh);

		// The upper body is the lower spine (key bone 4) and everything on it, which an attack
		// moves while the legs keep running.
		this.upper = m2.bones.map(() => false);
		const spine = m2.keyBoneLookup[4];
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
		this.base = this.playing(Anim.Stand, { loop: true });
	}

	static async load(fdid: number): Promise<Model> {
		const m2 = new M2(await file(fdid));
		if (m2.skins.length === 0) throw new Error(`model ${fdid} has no skin`);
		const skin = new Skin(await file(m2.skins[0]));
		const model = new Model(m2, skin);
		await model.loadAnim(model.base.seq);
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
			fog: (mat.flags & 0x2) === 0,
		};
		switch (mat.blend) {
			case 1: // alpha key
				params.alphaTest = 224 / 255;
				break;
			case 2: // alpha blend
				params.transparent = true;
				break;
			case 3: // add with no alpha
			case 4: // add
			case 7: // blend add
				params.transparent = true;
				params.blending = THREE.AdditiveBlending;
				break;
			case 5: // modulate
			case 6: // modulate 2x
				params.transparent = true;
				params.blending = THREE.MultiplyBlending;
				params.premultipliedAlpha = true;
				break;
		}
		const material = unlit ? new THREE.MeshBasicMaterial(params) : new THREE.MeshLambertMaterial(params);
		const bm: BatchMaterial = {
			material,
			batch,
			submesh: batch.submesh,
			textureType: tex?.type ?? -1,
			baseOpacity: 1,
		};
		if (tex && tex.type === 0 && tex.fdid > 0) {
			texture(tex.fdid, (tex.flags & 1) !== 0, (tex.flags & 2) !== 0)
				.then(t => {
					(material as THREE.MeshLambertMaterial).map = t;
					material.needsUpdate = true;
				})
				.catch(e => console.warn(`texture for ${m2.name}`, e));
		}
		void index;
		return bm;
	}

	// setTexture sets a replaceable texture: 1 the skin, 2 an item's own, 6 the hair, 11 to 13 a
	// creature's skins.
	setTexture(type: number, t: THREE.Texture) {
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

	submeshIds(): number[] {
		return [...new Set(this.skin.submeshes.map(s => s.id))].sort((a, b) => a - b);
	}

	// attachPoint gives an object that follows an attachment point (1 right hand, 2 left hand,
	// 11 the helm, 18 the chest, ...), for weapons and effects. It's undefined when the model
	// doesn't have the point.
	attachPoint(id: number): THREE.Object3D | undefined {
		const a = this.m2.attachment(id);
		if (!a) return undefined;
		const bone = this.bones[a.bone] ?? this.bones[0];
		const pivot = this.m2.bones[a.bone]?.pivot ?? [0, 0, 0];
		const point = new THREE.Object3D();
		point.position.set(a.position[0] - pivot[0], a.position[1] - pivot[1], a.position[2] - pivot[2]);
		bone.add(point);
		return point;
	}

	has(anim: number): boolean {
		return this.m2.sequenceFor(anim) >= 0;
	}

	// play switches the whole body to an animation, fading from the one before.
	play(anim: number, opts: PlayOptions = {}) {
		if (!opts.restart && this.base.anim === anim && !this.base.done) {
			this.base.speed = opts.speed ?? 1;
			return;
		}
		const next = this.playing(anim, opts);
		if (next.seq === this.base.seq && !opts.restart && !this.base.done) {
			this.base.anim = anim;
			this.base.speed = next.speed;
			return;
		}
		this.fadeFrom = this.base;
		this.fadeTotal = this.fadeLeft = opts.fade ?? this.m2.sequences[next.seq]?.blendIn ?? 150;
		this.base = next;
		this.ensure(next);
	}

	// playUpper plays a one-shot animation on the upper body only (an attack while running).
	// Without a split skeleton it plays on the whole body.
	playUpper(anim: number, opts: PlayOptions = {}) {
		this.overlay = this.playing(anim, { ...opts, loop: false });
		this.overlayFade = 0;
		this.ensure(this.overlay);
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
		seq = this.m2.resolve(seq);
		return { anim, seq, time: 0, speed: opts.speed ?? 1, loop: opts.loop ?? true, source: this.ready.get(seq), done: false };
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

	private advance(p: Playing, dt: number) {
		// Until its keys arrive, an animation holds at its start.
		if (!p.source || p.done) return;
		const duration = this.m2.sequences[p.seq]?.duration ?? 0;
		p.time += dt * 1000 * p.speed;
		if (duration <= 0) {
			p.time = 0;
			if (!p.loop) p.done = true;
		} else if (p.time >= duration) {
			if (p.loop) p.time %= duration;
			else {
				p.time = duration - 1;
				p.done = true;
			}
		}
	}

	// update moves the animations on by dt seconds and poses the bones.
	update(dt: number) {
		const dtMs = dt * 1000;
		this.globalTime += dtMs;
		this.advance(this.base, dt);
		if (this.fadeFrom) {
			this.advance(this.fadeFrom, dt);
			this.fadeLeft -= dtMs;
			if (this.fadeLeft <= 0 || !this.base.source) {
				if (this.base.source) this.fadeFrom = undefined;
			}
		}
		if (this.overlay) {
			this.advance(this.overlay, dt);
			this.overlayFade = Math.min(1, this.overlayFade + dtMs / 120);
			if (this.overlay.done) this.overlay = undefined;
		}
		const handsSeq = this.closed.right || this.closed.left ? this.m2.resolve(this.m2.sequenceFor(Anim.HandsClosed)) : -1;
		const s = this.scratch;
		const m2 = this.m2;
		for (let i = 0; i < m2.bones.length; i++) {
			const b = m2.bones[i];
			let p: Playing = this.base;
			let time = p.time;
			let seq = p.seq;
			let source = p.source;
			if (handsSeq >= 0 && ((this.closed.right && rightFingers.includes(b.keyBoneId)) || (this.closed.left && leftFingers.includes(b.keyBoneId)))) {
				seq = handsSeq;
				time = 0;
				source = this.ready.get(handsSeq);
			} else if (this.overlay && this.overlay.source && this.upper[i]) {
				p = this.overlay;
				time = p.time;
				seq = p.seq;
				source = p.source;
			}
			if (!source && this.fadeFrom?.source) {
				// The new animation's keys are still loading: keep the old pose.
				p = this.fadeFrom;
				time = p.time;
				seq = p.seq;
				source = p.source;
			}
			this.pose(b.translation, b.rotation, b.scale, seq, time, source, s.t, s.r, s.s);
			if (this.fadeFrom?.source && this.base.source && this.fadeTotal > 0 && p === this.base) {
				const f = this.fadeFrom;
				this.pose(b.translation, b.rotation, b.scale, f.seq, f.time, f.source, s.t2, s.r2, s.s2);
				const w = Math.max(0, Math.min(1, this.fadeLeft / this.fadeTotal));
				blend(s, w);
			}
			const parent = b.parent >= 0 ? m2.bones[b.parent].pivot : undefined;
			const bone = this.bones[i];
			bone.position.set(b.pivot[0] - (parent?.[0] ?? 0) + s.t[0], b.pivot[1] - (parent?.[1] ?? 0) + s.t[1], b.pivot[2] - (parent?.[2] ?? 0) + s.t[2]);
			bone.quaternion.set(s.r[0], s.r[1], s.r[2], s.r[3]);
			bone.scale.set(s.s[0], s.s[1], s.s[2]);
		}
		this.updateColors();
		this.particles?.update(dt);
	}

	private trackKeys(track: Track, seq: number, source: DataView | undefined): { keys: Keys | null; time?: number } {
		if (track.globalSeq !== 0xffff) {
			const loop = this.m2.globalLoops[track.globalSeq] || 1;
			return { keys: track.get(0, this.m2.data), time: this.globalTime % loop };
		}
		return { keys: track.get(seq, source) };
	}

	private pose(tr: Track, rot: Track, sc: Track, seq: number, time: number, source: DataView | undefined, t: number[], r: number[], s: number[]) {
		t[0] = t[1] = t[2] = 0;
		r[0] = r[1] = r[2] = 0;
		r[3] = 1;
		s[0] = s[1] = s[2] = 1;
		const a = this.trackKeys(tr, seq, source);
		if (a.keys) sample(a.keys, 3, a.time ?? time, t, tr.interpolation);
		const b = this.trackKeys(rot, seq, source);
		if (b.keys) sample(b.keys, 4, b.time ?? time, r, rot.interpolation);
		const c = this.trackKeys(sc, seq, source);
		if (c.keys) sample(c.keys, 3, c.time ?? time, s, sc.interpolation);
	}

	// trackValue samples a track (an emitter's rate, say) at the current animation's time. It's
	// false when the track has no keys there.
	trackValue(track: Track, out: number[]): boolean {
		const k = this.trackKeys(track, this.base.seq, this.base.source);
		if (!k.keys) return false;
		sample(k.keys, track.width, k.time ?? this.base.time, out, track.interpolation);
		return true;
	}

	// enableParticles starts the model's particle emitters, with their particles in `world`.
	enableParticles(world: THREE.Object3D) {
		if (!this.particles && this.m2.particles.length > 0) this.particles = new Particles(this, world);
	}

	// updateColors applies animated colors and transparency to the batches that have them.
	private updateColors() {
		const out = [0, 0, 0];
		const m2 = this.m2;
		for (const bm of this.mats) {
			const { batch } = bm;
			let alpha = 1;
			if (batch.color >= 0 && m2.colors[batch.color]) {
				const c = m2.colors[batch.color];
				const k = this.trackKeys(c.alpha, this.base.seq, this.base.source);
				if (k.keys) {
					sample(k.keys, 1, k.time ?? this.base.time, out, c.alpha.interpolation);
					alpha *= out[0] / 32767;
				}
				const kc = this.trackKeys(c.color, this.base.seq, this.base.source);
				if (kc.keys && bm.material.color) {
					sample(kc.keys, 3, kc.time ?? this.base.time, out, c.color.interpolation);
					bm.material.color.setRGB(out[0], out[1], out[2]);
				}
			}
			const wi = m2.transparencyLookup[batch.weightCombo];
			if (wi !== undefined && m2.textureWeights[wi]) {
				const w = m2.textureWeights[wi];
				const k = this.trackKeys(w, this.base.seq, this.base.source);
				if (k.keys) {
					sample(k.keys, 1, k.time ?? this.base.time, out, w.interpolation);
					alpha *= out[0] / 32767;
				}
			}
			if (alpha < 0.999 || bm.material.transparent) {
				bm.material.opacity = Math.max(0, Math.min(1, alpha));
				if (alpha < 0.999) bm.material.transparent = true;
			}
		}
	}

	dispose() {
		this.particles?.dispose();
		this.root.removeFromParent();
		this.mesh.geometry.dispose();
		for (const bm of this.mats) bm.material.dispose();
	}
}

function blend(s: { t: number[]; r: number[]; s: number[]; t2: number[]; r2: number[]; s2: number[] }, w: number) {
	// w is the old pose's share.
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
