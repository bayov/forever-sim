// Posing a model's bones from its animations: the bones' places in the model and their skinning
// matrices, from the animation state the model keeps (model.ts).
//
// A Poser runs on the page for your own character and in the pose worker (poseWorker.ts) for the
// raid test's crowd, from the same code, so both pose alike. It needs only the model's file and
// its animations, no Three scene.
//
// WoW moves a bone around its pivot: T(pivot) T(move) R S T(-pivot), on top of its parent. We pose
// the bones in flat arrays: each bone's place in the model is its parent's times
// T(pivot - parent's pivot + move) R S, and the skinning takes that times T(-pivot).

import * as THREE from 'three';
import { Keys, M2, sample, sampleTime, Track, Vec } from './m2';

// What a model asks its poser for, a frame: where its animations are.
export type PoseRequest = {
	// The model's own clock (ms), for tracks on a global sequence.
	globalTime: number;
	base: { seq: number; time: number; source?: DataView };
	// The old pose's share while a switch of animation fades in (0 when none). snapshot says the
	// switch just happened: the fade starts from the pose the bones have now.
	fade: number;
	snapshot: boolean;
	// The upper body's one-shot (a swing while running) and its weight (0 when none).
	overlay?: { seq: number; time: number; source?: DataView };
	overlayWeight: number;
	// The hands' closed pose, for holding weapons.
	hands?: { seq: number; source?: DataView; right: boolean; left: boolean };
	twist: number;
	// For a model with billboards: the camera's rotation and the model's place in the world.
	camera?: THREE.Quaternion;
	world?: THREE.Matrix4;
};

// The messages to the pose worker (poseWorker.ts). A frame's requests come as one Float64Array,
// RECORD numbers a model (see PoseClient.request for what's where). A model's two sets of bone
// matrices are shared memory, which the page reads and the worker writes.
export type ToWorker =
	| { type: 'file'; fdid: number; bytes: ArrayBuffer }
	| { type: 'anim'; fdid: number; seq: number; bytes: ArrayBuffer }
	| { type: 'model'; id: number; fdid: number; places: SharedArrayBuffer[]; skinning: SharedArrayBuffer[] }
	| { type: 'drop'; id: number }
	| { type: 'frame'; frame: number; data: Float64Array };
export const RECORD = 40;

// Bones that move the fingers: right hand 8 to 12, left hand 13 to 17 (key bone ids).
const rightFingers = [8, 9, 10, 11, 12];
const leftFingers = [13, 14, 15, 16, 17];

const upAxis = new THREE.Vector3(0, 1, 0);
const twistQ = new THREE.Quaternion();
const parentQ = new THREE.Quaternion();
const boneQ = new THREE.Quaternion();
const twistAxis = new THREE.Vector3();
const parentWorld = new THREE.Matrix4();
const billboardTmp = { q2: new THREE.Quaternion(), v: new THREE.Vector3(), s: new THREE.Vector3(), m: new THREE.Matrix4() };
// faceFront turns the model's front (+x) to the camera's back (+z), which faces the viewer.
const faceFront = new THREE.Quaternion().setFromAxisAngle(new THREE.Vector3(0, 1, 0), -Math.PI / 2);

export class Poser {
	// Each bone's place in the model (a 4 by 4 matrix, by columns), and its skinning matrix: the
	// place times T(-pivot). The pose worker points them at memory shared with the page.
	places: Float64Array;
	skinning: Float32Array;
	// The bones' rotations after the twist and the billboards, which the twist needs for the
	// bones above.
	private readonly quats: Float64Array;
	// The bones in an order where each parent comes before its children.
	private readonly order: Int32Array;
	// How much of the twist each bone takes, and whether it faces the camera.
	private readonly twistShare: Float64Array;
	readonly billboard: Uint8Array;
	readonly billboards: boolean;
	// The upper body (the lower spine and all on it), and the fingers with every joint below them.
	private readonly upper: boolean[];
	private readonly fingers: { right: boolean[]; left: boolean[] };
	// The pose the bones have (moves, rotations and scales, 10 numbers a bone), and the one a fade
	// starts from.
	private readonly pose0: Float32Array;
	private readonly fadePose: Float32Array;
	private readonly scratch = { t: new Float64Array(3), r: Float64Array.of(0, 0, 0, 1), s: Float64Array.of(1, 1, 1), t2: new Float64Array(3), r2: Float64Array.of(0, 0, 0, 1), s2: Float64Array.of(1, 1, 1) };
	// The time to sample at, in a Float64Array so it isn't boxed on each call.
	private readonly at = new Float64Array(1);
	private keyTime = -1;
	private globalTime = 0;

	constructor(readonly m2: M2) {
		const n = Math.max(1, m2.bones.length);
		this.places = new Float64Array(n * 16);
		this.skinning = new Float32Array(n * 16);
		for (let i = 0; i < n; i++) {
			this.places[i * 16] = this.places[i * 16 + 5] = this.places[i * 16 + 10] = this.places[i * 16 + 15] = 1;
			this.skinning[i * 16] = this.skinning[i * 16 + 5] = this.skinning[i * 16 + 10] = this.skinning[i * 16 + 15] = 1;
		}
		this.quats = new Float64Array(n * 4);
		this.order = boneOrder(m2);
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
	}

	// pose poses every bone for a request.
	pose(req: PoseRequest) {
		this.globalTime = req.globalTime;
		if (req.snapshot) this.fadePose.set(this.pose0);
		const s = this.scratch;
		const m2 = this.m2;
		const base = req.base;
		const fadeW = req.fade;
		const overlay = req.overlay;
		const ow = req.overlayWeight;
		const hands = req.hands;
		const handsSource = hands?.source;
		const camera = this.billboards ? req.camera : undefined;
		const places = this.places;
		const quats = this.quats;
		const skin = this.skinning;
		for (let k = 0; k < this.order.length; k++) {
			const i = this.order[k];
			const b = m2.bones[i];
			const finger = hands && handsSource && ((hands.right && this.fingers.right[i]) || (hands.left && this.fingers.left[i]));
			if (finger) {
				this.at[0] = 0;
				this.poseBone(b.translation, b.rotation, b.scale, hands.seq, handsSource, s.t, s.r, s.s);
			} else {
				this.at[0] = base.time;
				this.poseBone(b.translation, b.rotation, b.scale, base.seq, base.source, s.t, s.r, s.s);
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
					this.poseBone(b.translation, b.rotation, b.scale, overlay.seq, overlay.source, s.t2, s.r2, s.s2);
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
			// The numbers here never mix with a missing value (a root bone's parent pivot is zero, a
			// bone without a twist has a share of zero): Chrome boxes a decimal that meets another
			// kind of value, and that was 30 bytes a bone, every frame.
			const parent = b.parent >= 0 ? m2.bones[b.parent].pivot : zeroPivot;
			// The rotation goes straight into quats. Through a Three quaternion's set, each of its four
			// numbers was boxed, every bone. Only the twist and the billboards need one.
			const q = i * 4;
			quats[q] = s.r[0];
			quats[q + 1] = s.r[1];
			quats[q + 2] = s.r[2];
			quats[q + 3] = s.r[3];
			const share = this.twistShare[i] * req.twist;
			const turned = share !== 0 || (camera !== undefined && this.billboard[i] === 1);
			if (turned) boneQ.fromArray(quats, q);
			if (share !== 0) {
				// The twist is around the model's up (y, since the bones are in Three's frame). The
				// parent bones lean in the run, so we turn that axis into the parent's frame.
				parentQ.identity();
				for (let p = b.parent; p >= 0; p = m2.bones[p].parent) parentQ.premultiply(twistQ.fromArray(quats, p * 4));
				twistAxis.copy(upAxis).applyQuaternion(parentQ.invert());
				boneQ.premultiply(twistQ.setFromAxisAngle(twistAxis, share));
			}
			if (camera && req.world && this.billboard[i]) {
				// A billboard turns its +x (the model's front) to the camera, with its up staying the
				// camera's up: its rotation is the parent's in the world, undone, then the camera's.
				parentWorld.copy(req.world);
				if (b.parent >= 0) parentWorld.multiply(billboardTmp.m.fromArray(places, b.parent * 16));
				parentWorld.decompose(billboardTmp.v, billboardTmp.q2, billboardTmp.s);
				boneQ.copy(billboardTmp.q2.invert()).multiply(camera).multiply(faceFront);
			}
			if (turned) boneQ.toArray(quats, q);
			const at = placeAt;
			at[0] = b.pivot[0] - parent[0] + s.t[0];
			at[1] = b.pivot[1] - parent[1] + s.t[1];
			at[2] = b.pivot[2] - parent[2] + s.t[2];
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
	}

	// rest puts every bone back where it rests, for the model viewer.
	rest() {
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
	}

	// trackKeys gives a track's keys for a sequence. A track on a global sequence runs on its own
	// clock: then keyTime is the time to sample at, else it's -1 (sample at the animation's time).
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

	// sampleTrack samples a track at the time in `at`.
	private sampleTrack(track: Track, seq: number, source: DataView | undefined, out: Vec): boolean {
		const keys = this.trackKeys(track, seq, source);
		if (!keys) return false;
		sampleTime[0] = this.keyTime >= 0 ? this.keyTime : this.at[0];
		sample(keys, track.width, out, track.interpolation);
		return true;
	}

	// sampleAt samples a track at a time on the model's clock, for its colors and emitters.
	sampleAt(track: Track, seq: number, time: number, source: DataView | undefined, out: Vec, globalTime: number): boolean {
		this.globalTime = globalTime;
		this.at[0] = time;
		return this.sampleTrack(track, seq, source, out);
	}

	private poseBone(tr: Track, rot: Track, sc: Track, seq: number, source: DataView | undefined, t: Vec, r: Vec, s: Vec) {
		t[0] = t[1] = t[2] = 0;
		r[0] = r[1] = r[2] = 0;
		r[3] = 1;
		s[0] = s[1] = s[2] = 1;
		this.sampleTrack(tr, seq, source, t);
		this.sampleTrack(rot, seq, source, r);
		this.sampleTrack(sc, seq, source, s);
	}
}

// twistShares spreads a strafe's twist over the bones from the lower spine up to the head. The
// lower spine takes 55%, the bones between it and the head (chest and neck) 25% together, and the
// head the last 20%. On a strafe that leaves the chest about 30 degrees toward where you run. A
// model without a head on its spine twists the lower spine alone.
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
	if (between.length === 0)
		return [
			[spine, 0.6],
			[head, 0.4],
		];
	return [[spine, 0.55], ...between.reverse().map(b => [b, 0.25 / between.length] as [number, number]), [head, 0.2]];
}

// A root bone's parent pivot. It's an array of decimals (made with one, then zeroed), like the
// bones' pivots.
const zeroPivot = [0.5, 0.5, 0.5];
zeroPivot.fill(0);

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
