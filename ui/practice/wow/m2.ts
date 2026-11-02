// Readers for WoW's model files: the M2 model (MD21 chunks), its .skin (which vertices and
// triangles to draw, in which batches), its .skel skeleton and its .anim animation files.
//
// We follow wow.export (MIT, Kruithne and Marlamin) and the format pages on wowdev.wiki. WoW's
// model space has x forward, y left and z up. Three.js has y up, so every point (x, y, z) becomes
// (x, z, -y) as we read it, and so does every rotation.

export class Reader {
	readonly view: DataView;
	pos = 0;

	constructor(
		readonly buf: ArrayBuffer,
		readonly base = 0,
		readonly length = buf.byteLength - base,
	) {
		this.view = new DataView(buf, base, length);
	}

	get remaining() {
		return this.length - this.pos;
	}
	u8() {
		return this.view.getUint8(this.pos++);
	}
	i8() {
		return this.view.getInt8(this.pos++);
	}
	u16() {
		const v = this.view.getUint16(this.pos, true);
		this.pos += 2;
		return v;
	}
	i16() {
		const v = this.view.getInt16(this.pos, true);
		this.pos += 2;
		return v;
	}
	u32() {
		const v = this.view.getUint32(this.pos, true);
		this.pos += 4;
		return v;
	}
	i32() {
		const v = this.view.getInt32(this.pos, true);
		this.pos += 4;
		return v;
	}
	f32() {
		const v = this.view.getFloat32(this.pos, true);
		this.pos += 4;
		return v;
	}
	// vec3 reads a point and turns it into Three.js axes.
	vec3(): [number, number, number] {
		const x = this.f32();
		const y = this.f32();
		const z = this.f32();
		return [x, z, -y];
	}
	fourcc() {
		return String.fromCharCode(this.u8(), this.u8(), this.u8(), this.u8());
	}
	// array reads an M2Array (count, offset) header.
	array(): { n: number; ofs: number } {
		const n = this.u32();
		const ofs = this.u32();
		return { n, ofs };
	}
	at<T>(ofs: number, fn: () => T): T {
		const back = this.pos;
		this.pos = ofs;
		const v = fn();
		this.pos = back;
		return v;
	}
}

// An event in the model's animations: $HIT (a swing connects), $CSS (the swing's whoosh), $SHL and
// $SHR (a weapon moves to or from its sheath), footsteps and such, at times in each sequence.
export type M2Event = { id: string; data: number; bone: number; times: { n: number; ofs: number }[] };

// Track values come in these shapes.
// 'gravity' is a particle emitter's packed gravity (emitter flag 0x800000): a direction in two
// int8s and a strength in an int16, which we unpack to an acceleration vector.
export type TrackKind = 'vec3' | 'rgb' | 'quat' | 'scale' | 'i16' | 'u8' | 'f32' | 'gravity';

const kindSize: Record<TrackKind, number> = { vec3: 12, rgb: 12, quat: 8, scale: 12, i16: 2, u8: 1, f32: 4, gravity: 4 };
const kindWidth: Record<TrackKind, number> = { vec3: 3, rgb: 3, quat: 4, scale: 3, i16: 1, u8: 1, f32: 1, gravity: 3 };

// A decoded track for one sequence: key times in ms and the values, `width` floats per key.
export type Keys = { times: Uint32Array; values: Float32Array };

// Track is an animated value (a bone's rotation, a color's alpha). It has one list of keys per
// sequence. The keys of an inline sequence sit in the model file, and the keys of the others in
// that sequence's .anim file, at the same offsets. So we keep the offsets and read the keys when a
// sequence first plays.
export class Track {
	readonly width: number;
	private keys = new Map<number, Keys | null>();

	constructor(
		readonly kind: TrackKind,
		readonly interpolation: number,
		readonly globalSeq: number,
		readonly times: { n: number; ofs: number }[],
		readonly values: { n: number; ofs: number }[],
	) {
		this.width = kindWidth[kind];
	}

	static read(r: Reader, kind: TrackKind): Track {
		const interpolation = r.u16();
		const globalSeq = r.u16();
		const t = r.array();
		const v = r.array();
		const times = r.at(t.ofs, () => Array.from({ length: t.n }, () => r.array()));
		const values = r.at(v.ofs, () => Array.from({ length: v.n }, () => r.array()));
		return new Track(kind, interpolation, globalSeq, times, values);
	}

	get animated() {
		return this.times.some(t => t.n > 0);
	}

	// get gives a sequence's keys, read from `source` (the model or the sequence's .anim data) the
	// first time. It gives null when the track has no keys there.
	get(seq: number, source: DataView | undefined): Keys | null {
		// A track on a global sequence loops on its own clock, with its keys in slot 0. A track with
		// one slot (a ghost wolf's tint) holds for every sequence.
		const slot = this.globalSeq !== 0xffff || this.times.length === 1 ? 0 : seq;
		const cached = this.keys.get(slot);
		if (cached !== undefined) return cached;
		const t = this.times[slot];
		const v = this.values[slot];
		if (!source || !t || !v || t.n === 0 || v.n === 0) {
			// Without the source the keys may come later (an .anim still loading), so we don't cache.
			if (source) this.keys.set(slot, null);
			return null;
		}
		if (t.ofs + t.n * 4 > source.byteLength) {
			this.keys.set(slot, null);
			return null;
		}
		const times = new Uint32Array(t.n);
		for (let i = 0; i < t.n; i++) times[i] = source.getUint32(t.ofs + i * 4, true);
		// Hermite and bezier keys carry in and out tangents after each value. We take the value only.
		const stride = kindSize[this.kind] * (this.interpolation >= 2 ? 3 : 1);
		const count = Math.min(v.n, t.n);
		if (v.ofs + count * stride > source.byteLength) {
			this.keys.set(slot, null);
			return null;
		}
		const values = new Float32Array(count * this.width);
		for (let i = 0; i < count; i++) {
			const o = v.ofs + i * stride;
			const w = i * this.width;
			switch (this.kind) {
				case 'vec3': {
					const x = source.getFloat32(o, true);
					const y = source.getFloat32(o + 4, true);
					const z = source.getFloat32(o + 8, true);
					values[w] = x;
					values[w + 1] = z;
					values[w + 2] = -y;
					break;
				}
				case 'rgb':
					values[w] = source.getFloat32(o, true);
					values[w + 1] = source.getFloat32(o + 4, true);
					values[w + 2] = source.getFloat32(o + 8, true);
					break;
				case 'scale': {
					const x = source.getFloat32(o, true);
					const y = source.getFloat32(o + 4, true);
					const z = source.getFloat32(o + 8, true);
					values[w] = x;
					values[w + 1] = z;
					values[w + 2] = y;
					break;
				}
				case 'quat': {
					// Compressed quaternions: int16 per part.
					const q = (k: number) => {
						const s = source.getInt16(o + k * 2, true);
						return (s < 0 ? s + 32768 : s - 32767) / 32767;
					};
					values[w] = q(0);
					values[w + 1] = q(2);
					values[w + 2] = -q(1);
					values[w + 3] = q(3);
					break;
				}
				case 'i16':
					values[w] = source.getInt16(o, true);
					break;
				case 'u8':
					values[w] = source.getUint8(o);
					break;
				case 'f32':
					values[w] = source.getFloat32(o, true);
					break;
				case 'gravity': {
					// wowdev.wiki M2, "Compressed particle gravity".
					const dx = source.getInt8(o) / 128;
					const dy = source.getInt8(o + 1) / 128;
					let dz = Math.sqrt(Math.max(0, 1 - dx * dx - dy * dy));
					let mag = source.getInt16(o + 2, true) * 0.04238648;
					if (mag < 0) {
						dz = -dz;
						mag = -mag;
					}
					// WoW's (x, y, z) is (x, z, -y) here.
					values[w] = dx * mag;
					values[w + 1] = dz * mag;
					values[w + 2] = -dy * mag;
					break;
				}
			}
		}
		const keys = { times: times.subarray(0, count), values };
		this.keys.set(slot, keys);
		return keys;
	}
}

// sample finds the value at a time (ms) into `out`, interpolating between keys. Rotations use a
// normalized lerp, which is what the client does.
export function sample(keys: Keys, width: number, time: number, out: number[], interpolation: number) {
	const { times, values } = keys;
	const n = times.length;
	if (n === 1 || time <= times[0]) {
		for (let k = 0; k < width; k++) out[k] = values[k];
		return;
	}
	if (time >= times[n - 1]) {
		for (let k = 0; k < width; k++) out[k] = values[(n - 1) * width + k];
		return;
	}
	// Binary search for the key before the time.
	let lo = 0;
	let hi = n - 1;
	while (hi - lo > 1) {
		const mid = (lo + hi) >> 1;
		if (times[mid] <= time) lo = mid;
		else hi = mid;
	}
	const a = lo * width;
	const b = hi * width;
	if (interpolation === 0) {
		for (let k = 0; k < width; k++) out[k] = values[a + k];
		return;
	}
	const t = (time - times[lo]) / Math.max(1, times[hi] - times[lo]);
	if (width === 4) {
		// Take the short way around.
		let dot = 0;
		for (let k = 0; k < 4; k++) dot += values[a + k] * values[b + k];
		const sign = dot < 0 ? -1 : 1;
		let len = 0;
		for (let k = 0; k < 4; k++) {
			out[k] = values[a + k] * (1 - t) + sign * values[b + k] * t;
			len += out[k] * out[k];
		}
		len = Math.sqrt(len) || 1;
		for (let k = 0; k < 4; k++) out[k] /= len;
		return;
	}
	for (let k = 0; k < width; k++) out[k] = values[a + k] * (1 - t) + values[b + k] * t;
}

export type Sequence = {
	id: number; // AnimationData id: 0 stand, 5 run, 17 one hand attack, ...
	variation: number;
	duration: number; // ms
	moveSpeed: number;
	flags: number;
	frequency: number;
	blendIn: number;
	blendOut: number;
	variationNext: number;
	aliasNext: number;
};

export type Bone = {
	keyBoneId: number;
	flags: number;
	parent: number;
	pivot: [number, number, number];
	translation: Track;
	rotation: Track;
	scale: Track;
};

export type Attachment = { id: number; bone: number; position: [number, number, number] };

export type M2Texture = { type: number; flags: number; fdid: number };

export type Material = { flags: number; blend: number };

export type ColorAnim = { color: Track; alpha: Track };

export type AnimFile = { anim: number; variation: number; fdid: number };

export type ParticleEmitter = {
	bone: number;
	position: [number, number, number];
	flags: number;
	texture: number;
	blend: number;
	emitterType: number;
	tileRows: number;
	tileCols: number;
	// Animated settings, read like any track.
	speed: Track;
	speedVariation: Track;
	verticalRange: Track;
	horizontalRange: Track;
	gravity: Track;
	lifespan: Track;
	emissionRate: Track;
	areaLength: Track;
	areaWidth: Track;
	enabled: Track;
	// Over each particle's life: keys at their times (0 to 1 of the life).
	colors: [number, number, number][];
	alphas: number[];
	scales: [number, number][];
	colorTimes: number[];
	alphaTimes: number[];
	scaleTimes: number[];
	cellTimes: number[];
	baseSpinVary: number;
	spinVary: number;
	midPoint: number;
	headCells: number[];
	lifespanVary: number;
	emissionVary: number;
	drag: number;
	spin: number;
	baseSpin: number;
	burstMultiplier: number;
	tailLength: number;
	twinkleSpeed: number;
	twinklePercent: number;
	twinkleScale: [number, number];
	followSpeed: [number, number];
	wind: [number, number, number];
	windTime: number;
	multiTexture: number[];
};

export type RibbonEmitter = {
	bone: number;
	position: [number, number, number];
	textures: number[];
	materials: number[];
	color: Track;
	alpha: Track;
	heightAbove: Track;
	heightBelow: Track;
	edgesPerSecond: number;
	edgeLifetime: number;
	gravity: number;
	rows: number;
	cols: number;
	visibility: Track;
};

export class M2 {
	version = 0;
	name = '';
	flags = 0;
	globalLoops: number[] = [];
	sequences: Sequence[] = [];
	sequenceLookup: number[] = [];
	bones: Bone[] = [];
	keyBoneLookup: number[] = [];
	vertexCount = 0;
	// Per vertex: position (3), normal (3), uv (2), uv2 (2), bone weights (4) and indices (4).
	positions!: Float32Array;
	normals!: Float32Array;
	uvs!: Float32Array;
	uvs2!: Float32Array;
	boneWeights!: Uint8Array;
	boneIndices!: Uint8Array;
	colors: ColorAnim[] = [];
	textures: M2Texture[] = [];
	textureWeights: Track[] = [];
	textureTransforms: { translation: Track; rotation: Track; scale: Track }[] = [];
	materials: Material[] = [];
	textureCombos: number[] = [];
	textureCoordCombos: number[] = [];
	transparencyLookup: number[] = [];
	textureTransformLookup: number[] = [];
	boundsMin: [number, number, number] = [0, 0, 0];
	boundsMax: [number, number, number] = [0, 0, 0];
	attachments: Attachment[] = [];
	attachmentLookup: number[] = [];
	particles: ParticleEmitter[] = [];
	ribbons: RibbonEmitter[] = [];
	events: M2Event[] = [];
	skins: number[] = [];
	animFiles: AnimFile[] = [];
	skeleton = 0;
	// The model's data, where inline tracks point.
	data!: DataView;

	constructor(buf: ArrayBuffer) {
		const r = new Reader(buf);
		while (r.remaining >= 8) {
			const id = r.fourcc();
			const size = r.u32();
			const start = r.pos;
			switch (id) {
				case 'MD21':
					this.readMD20(buf, start, size);
					break;
				case 'SFID':
					// The first numSkinProfiles ids are the skins, then the LOD skins.
					for (let i = 0; i < size / 4; i++) this.skins.push(r.u32());
					break;
				case 'TXID':
					for (let i = 0; i < size / 4 && i < this.textures.length; i++) this.textures[i].fdid = r.u32();
					break;
				case 'AFID':
					for (let i = 0; i < size / 8; i++) this.animFiles.push({ anim: r.u16(), variation: r.u16(), fdid: r.u32() });
					break;
				case 'SKID':
					this.skeleton = r.u32();
					break;
			}
			r.pos = start + size;
		}
		// An old model without chunks starts with MD20 itself.
		if (!this.data && new DataView(buf).getUint32(0, true) === 0x3032444d) this.readMD20(buf, 0, buf.byteLength);
	}

	private readMD20(buf: ArrayBuffer, base: number, size: number) {
		const r = new Reader(buf, base, size);
		this.data = r.view;
		if (r.fourcc() !== 'MD20') throw new Error('not an M2 model');
		this.version = r.u32();
		const name = r.array();
		this.name = r.at(name.ofs, () => {
			let s = '';
			for (let i = 0; i < name.n - 1; i++) s += String.fromCharCode(r.u8());
			return s;
		});
		this.flags = r.u32();
		const loops = r.array();
		this.globalLoops = r.at(loops.ofs, () => Array.from({ length: loops.n }, () => r.u32()));
		const seqs = r.array();
		this.sequences = r.at(seqs.ofs, () => Array.from({ length: seqs.n }, () => readSequence(r)));
		const seqLookup = r.array();
		this.sequenceLookup = r.at(seqLookup.ofs, () => Array.from({ length: seqLookup.n }, () => r.i16()));
		const bones = r.array();
		this.bones = r.at(bones.ofs, () => Array.from({ length: bones.n }, () => readBone(r)));
		const keyBones = r.array();
		this.keyBoneLookup = r.at(keyBones.ofs, () => Array.from({ length: keyBones.n }, () => r.i16()));
		const verts = r.array();
		this.readVertices(r, verts.n, verts.ofs);
		r.u32(); // the skin count, which SFID gives too
		const colors = r.array();
		this.colors = r.at(colors.ofs, () => Array.from({ length: colors.n }, () => ({ color: Track.read(r, 'rgb'), alpha: Track.read(r, 'i16') })));
		const textures = r.array();
		this.textures = r.at(textures.ofs, () =>
			Array.from({ length: textures.n }, () => {
				const type = r.u32();
				const flags = r.u32();
				r.array(); // the file name, which old models had and TXID replaces
				return { type, flags, fdid: 0 };
			}),
		);
		const weights = r.array();
		this.textureWeights = r.at(weights.ofs, () => Array.from({ length: weights.n }, () => Track.read(r, 'i16')));
		const transforms = r.array();
		this.textureTransforms = r.at(transforms.ofs, () =>
			Array.from({ length: transforms.n }, () => ({ translation: Track.read(r, 'vec3'), rotation: Track.read(r, 'quat'), scale: Track.read(r, 'scale') })),
		);
		r.array(); // replaceable texture lookup
		const mats = r.array();
		this.materials = r.at(mats.ofs, () => Array.from({ length: mats.n }, () => ({ flags: r.u16(), blend: r.u16() })));
		r.array(); // bone combos: the vertices already name their bones
		const texCombos = r.array();
		this.textureCombos = r.at(texCombos.ofs, () => Array.from({ length: texCombos.n }, () => r.i16()));
		const coordCombos = r.array();
		this.textureCoordCombos = r.at(coordCombos.ofs, () => Array.from({ length: coordCombos.n }, () => r.i16()));
		const transp = r.array();
		this.transparencyLookup = r.at(transp.ofs, () => Array.from({ length: transp.n }, () => r.u16()));
		const ttLookup = r.array();
		this.textureTransformLookup = r.at(ttLookup.ofs, () => Array.from({ length: ttLookup.n }, () => r.i16()));
		this.boundsMin = r.vec3();
		this.boundsMax = r.vec3();
		// vec3 turned -y into z, so the box's z ends swap.
		[this.boundsMin[2], this.boundsMax[2]] = [Math.min(this.boundsMin[2], this.boundsMax[2]), Math.max(this.boundsMin[2], this.boundsMax[2])];
		r.f32(); // bounding sphere
		r.pos += 24 + 4; // collision box and sphere
		r.array(); // collision triangles
		r.array(); // collision vertices
		r.array(); // collision normals
		const atts = r.array();
		this.attachments = r.at(atts.ofs, () =>
			Array.from({ length: atts.n }, () => {
				const id = r.u32();
				const bone = r.u16();
				r.u16();
				const position = r.vec3();
				Track.read(r, 'u8'); // whether attached models animate
				return { id, bone, position };
			}),
		);
		const attLookup = r.array();
		this.attachmentLookup = r.at(attLookup.ofs, () => Array.from({ length: attLookup.n }, () => r.i16()));
		const events = r.array();
		this.events = r.at(events.ofs, () =>
			Array.from({ length: events.n }, () => {
				const id = r.fourcc();
				const data = r.u32();
				const bone = r.u32();
				r.pos += 12 + 4; // position, interpolation and global sequence
				const t = r.array();
				const times = r.at(t.ofs, () => Array.from({ length: t.n }, () => r.array()));
				return { id, data, bone, times };
			}),
		);
		r.array(); // lights
		r.array(); // cameras
		r.array(); // camera lookup
		const ribbons = r.array();
		const particles = r.array();
		// Effects are a bonus: a model whose emitters we misread still draws.
		try {
			this.ribbons = r.at(ribbons.ofs, () => Array.from({ length: ribbons.n }, () => readRibbon(r)));
		} catch (e) {
			console.warn(`model ${this.name}: ribbons unreadable`, e);
		}
		try {
			this.particles = r.at(particles.ofs, () => Array.from({ length: particles.n }, () => readParticle(r, this.flags)));
		} catch (e) {
			console.warn(`model ${this.name}: particles unreadable`, e);
		}
	}

	private readVertices(r: Reader, n: number, ofs: number) {
		this.vertexCount = n;
		this.positions = new Float32Array(n * 3);
		this.normals = new Float32Array(n * 3);
		this.uvs = new Float32Array(n * 2);
		this.uvs2 = new Float32Array(n * 2);
		this.boneWeights = new Uint8Array(n * 4);
		this.boneIndices = new Uint8Array(n * 4);
		const v = r.view;
		for (let i = 0; i < n; i++) {
			const o = ofs + i * 48;
			const x = v.getFloat32(o, true);
			const y = v.getFloat32(o + 4, true);
			const z = v.getFloat32(o + 8, true);
			this.positions[i * 3] = x;
			this.positions[i * 3 + 1] = z;
			this.positions[i * 3 + 2] = -y;
			for (let k = 0; k < 4; k++) {
				this.boneWeights[i * 4 + k] = v.getUint8(o + 12 + k);
				this.boneIndices[i * 4 + k] = v.getUint8(o + 16 + k);
			}
			const nx = v.getFloat32(o + 20, true);
			const ny = v.getFloat32(o + 24, true);
			const nz = v.getFloat32(o + 28, true);
			this.normals[i * 3] = nx;
			this.normals[i * 3 + 1] = nz;
			this.normals[i * 3 + 2] = -ny;
			this.uvs[i * 2] = v.getFloat32(o + 32, true);
			this.uvs[i * 2 + 1] = v.getFloat32(o + 36, true);
			this.uvs2[i * 2] = v.getFloat32(o + 40, true);
			this.uvs2[i * 2 + 1] = v.getFloat32(o + 44, true);
		}
	}

	// sequenceFor finds the first sequence (variation 0) that plays an animation, or -1.
	sequenceFor(anim: number): number {
		if (this.sequenceLookup.length > 0) {
			// The lookup is a hash table on the id with linear probing.
			const n = this.sequenceLookup.length;
			for (let i = 0; i < n; i++) {
				const idx = this.sequenceLookup[(anim + i * i) % n];
				if (idx < 0) break;
				if (this.sequences[idx]?.id === anim) return idx;
			}
		}
		return this.sequences.findIndex(s => s.id === anim);
	}

	// resolve follows a sequence's alias (flag 0x40) to the one that holds its keys.
	resolve(seq: number): number {
		let s = seq;
		for (let guard = 0; guard < 16 && this.sequences[s] && this.sequences[s].flags & 0x40; guard++) s = this.sequences[s].aliasNext;
		return s;
	}

	// inline says whether a sequence's keys are in the model file, not an .anim file.
	inline(seq: number): boolean {
		return (this.sequences[seq]?.flags & 0x20) !== 0;
	}

	animFileFor(seq: number): number {
		const s = this.sequences[seq];
		return this.animFiles.find(a => a.anim === s.id && a.variation === s.variation)?.fdid ?? 0;
	}

	// eventTimes gives the times (ms) in a sequence when an event fires, read from `source` (the
	// model or the sequence's .anim data).
	eventTimes(ev: M2Event, seq: number, source: DataView | undefined): number[] {
		const t = ev.times[seq];
		if (!source || !t || t.n === 0 || t.ofs + t.n * 4 > source.byteLength) return [];
		return Array.from({ length: t.n }, (_, i) => source.getUint32(t.ofs + i * 4, true));
	}

	attachment(id: number): Attachment | undefined {
		const i = this.attachmentLookup[id];
		return i !== undefined && i >= 0 ? this.attachments[i] : undefined;
	}
}

function readSequence(r: Reader): Sequence {
	const id = r.u16();
	const variation = r.u16();
	const duration = r.u32();
	const moveSpeed = r.f32();
	const flags = r.u32();
	const frequency = r.i16();
	r.u16();
	r.u32(); // replay min
	r.u32(); // replay max
	const blendIn = r.u16();
	const blendOut = r.u16();
	r.pos += 12 + 12 + 4; // bounds
	const variationNext = r.i16();
	const aliasNext = r.u16();
	return { id, variation, duration, moveSpeed, flags, frequency, blendIn, blendOut, variationNext, aliasNext };
}

function readBone(r: Reader): Bone {
	const keyBoneId = r.i32();
	const flags = r.u32();
	const parent = r.i16();
	r.u16(); // submesh
	r.u32(); // name CRC
	const translation = Track.read(r, 'vec3');
	const rotation = Track.read(r, 'quat');
	const scale = Track.read(r, 'scale');
	const pivot = r.vec3();
	return { keyBoneId, flags, parent, pivot, translation, rotation, scale };
}

// FBlocks are the particle settings that change over each particle's life: keys at fixed times
// from 0 to 1, with no interpolation flag or global sequence.
// fblockTimes reads an FBlock's key times (fixed16, 0 to 1 of the life) without moving on.
function fblockTimes(r: Reader): number[] {
	const back = r.pos;
	const t = r.array();
	r.pos = back;
	return r.at(t.ofs, () => Array.from({ length: t.n }, () => r.i16() / 32767));
}

function readFBlock(r: Reader, width: number): number[][] {
	const t = r.array();
	const v = r.array();
	return r.at(v.ofs, () =>
		Array.from({ length: Math.min(v.n, Math.max(t.n, v.n)) }, () =>
			Array.from({ length: width }, () => (width === 1 ? r.i16() / 32767 : r.f32())),
		),
	);
}

function readFBlockU16(r: Reader): number[] {
	r.array();
	const v = r.array();
	return r.at(v.ofs, () => Array.from({ length: v.n }, () => r.u16()));
}

function readFBlockScales(r: Reader): [number, number][] {
	const t = r.array();
	const v = r.array();
	void t;
	return r.at(v.ofs, () => Array.from({ length: v.n }, () => [r.f32(), r.f32()] as [number, number]));
}

function readRibbon(r: Reader): RibbonEmitter {
	r.i32(); // id
	const bone = r.u32();
	const position = r.vec3();
	const tex = r.array();
	const textures = r.at(tex.ofs, () => Array.from({ length: tex.n }, () => r.u16()));
	const mat = r.array();
	const materials = r.at(mat.ofs, () => Array.from({ length: mat.n }, () => r.u16()));
	const color = Track.read(r, 'rgb');
	const alpha = Track.read(r, 'i16');
	const heightAbove = Track.read(r, 'f32');
	const heightBelow = Track.read(r, 'f32');
	const edgesPerSecond = r.f32();
	const edgeLifetime = r.f32();
	const gravity = r.f32();
	const rows = r.u16();
	const cols = r.u16();
	Track.read(r, 'i16'); // texture slot, unused
	const visibility = Track.read(r, 'u8');
	r.i16(); // priority plane
	r.u8(); // color index
	r.u8(); // texture transform
	return { bone, position, textures, materials, color, alpha, heightAbove, heightBelow, edgesPerSecond, edgeLifetime, gravity, rows, cols, visibility };
}

function readParticle(r: Reader, modelFlags: number): ParticleEmitter {
	r.i32(); // id
	const flags = r.u32();
	const position = r.vec3();
	const bone = r.u16();
	const texture = r.u16();
	r.array(); // geometry model file name
	r.array(); // recursion model file name
	const blend = r.u8();
	const emitterType = r.u8();
	r.u16(); // particle color index
	r.u8(); // particle type
	r.u8(); // head or tail
	r.u16(); // texture tile rotation
	const tileRows = r.u16();
	const tileCols = r.u16();
	const speed = Track.read(r, 'f32');
	const speedVariation = Track.read(r, 'f32');
	const verticalRange = Track.read(r, 'f32');
	const horizontalRange = Track.read(r, 'f32');
	const gravity = Track.read(r, flags & 0x800000 ? 'gravity' : 'f32');
	const lifespan = Track.read(r, 'f32');
	const lifespanVary = r.f32();
	const emissionRate = Track.read(r, 'f32');
	const emissionVary = r.f32();
	const areaLength = Track.read(r, 'f32');
	const areaWidth = Track.read(r, 'f32');
	Track.read(r, 'f32'); // z source
	const colorTimes = fblockTimes(r);
	const colors = readFBlock(r, 3) as [number, number, number][];
	const alphaTimes = fblockTimes(r);
	const alphas = readFBlock(r, 1).map(a => a[0]);
	const scaleTimes = fblockTimes(r);
	const scales = readFBlockScales(r);
	const scaleVary: [number, number] = [r.f32(), r.f32()];
	void scaleVary;
	const cellTimes = fblockTimes(r);
	const headCells = readFBlockU16(r);
	readFBlockU16(r); // tail cells
	const tailLength = r.f32();
	const twinkleSpeed = r.f32();
	const twinklePercent = r.f32();
	const twinkleScale: [number, number] = [r.f32(), r.f32()];
	r.f32(); // burst multiplier
	const drag = r.f32();
	const baseSpin = r.f32();
	const baseSpinVary = r.f32();
	const spin = r.f32();
	const spinVary = r.f32();
	r.pos += 24; // tumble
	const wind: [number, number, number] = r.vec3();
	const windTime = r.f32();
	const followSpeed1 = r.f32();
	r.f32(); // follow scale 1
	const followSpeed2 = r.f32();
	r.f32(); // follow scale 2
	const followSpeed: [number, number] = [followSpeed1, followSpeed2];
	r.array(); // spline points
	const enabled = Track.read(r, 'u8');
	// Then the multi texture settings (two pairs of 2 x int16 each), which we don't use.
	r.pos += 16;
	void modelFlags;
	return {
		bone,
		position,
		flags,
		texture,
		blend,
		emitterType,
		tileRows: Math.max(1, tileRows),
		tileCols: Math.max(1, tileCols),
		speed,
		speedVariation,
		verticalRange,
		horizontalRange,
		gravity,
		lifespan,
		emissionRate,
		areaLength,
		areaWidth,
		enabled,
		colors,
		alphas,
		scales,
		colorTimes,
		alphaTimes,
		scaleTimes,
		cellTimes,
		baseSpinVary,
		spinVary,
		midPoint: 0.5,
		headCells,
		lifespanVary,
		emissionVary,
		drag,
		spin,
		baseSpin,
		burstMultiplier: 1,
		tailLength,
		twinkleSpeed,
		twinklePercent,
		twinkleScale,
		followSpeed,
		wind,
		windTime,
		multiTexture: [],
	};
}

export type Submesh = {
	id: number; // the geoset id: 0 the body, 101 a hair style, 501 bare hands, ...
	vertexStart: number;
	vertexCount: number;
	indexStart: number;
	indexCount: number;
	centerBone: number;
	center: [number, number, number];
};

export type Batch = {
	flags: number;
	priority: number;
	shader: number;
	submesh: number;
	color: number;
	material: number;
	materialLayer: number;
	textureCount: number;
	textureCombo: number;
	coordCombo: number;
	weightCombo: number;
	transformCombo: number;
};

// Skin is a model's draw list: the vertices it uses (by model vertex index), the triangles over
// them, the submeshes (geosets) and the batches (a submesh with a material and its textures).
export class Skin {
	vertices: Uint32Array;
	indices: Uint16Array;
	submeshes: Submesh[];
	batches: Batch[];

	constructor(buf: ArrayBuffer) {
		const r = new Reader(buf);
		if (r.fourcc() !== 'SKIN') throw new Error('not a skin file');
		const verts = r.array();
		const tris = r.array();
		r.array(); // bone properties
		const subs = r.array();
		const batches = r.array();
		this.vertices = r.at(verts.ofs, () => {
			const a = new Uint32Array(verts.n);
			for (let i = 0; i < verts.n; i++) a[i] = r.u16();
			return a;
		});
		this.indices = r.at(tris.ofs, () => {
			const a = new Uint16Array(tris.n);
			for (let i = 0; i < tris.n; i++) a[i] = r.u16();
			return a;
		});
		this.submeshes = r.at(subs.ofs, () =>
			Array.from({ length: subs.n }, () => {
				const id = r.u16();
				const level = r.u16();
				const vertexStart = r.u16();
				const vertexCount = r.u16();
				const indexStart = r.u16() + (level << 16);
				const indexCount = r.u16();
				r.u16(); // bone count
				r.u16(); // bone combo start
				r.u16(); // bone influences
				const centerBone = r.u16();
				const center = r.vec3();
				r.pos += 12 + 4; // sort center and radius
				return { id, vertexStart, vertexCount, indexStart, indexCount, centerBone, center };
			}),
		);
		this.batches = r.at(batches.ofs, () =>
			Array.from({ length: batches.n }, () => {
				const flags = r.u8();
				const priority = r.i8();
				const shader = r.u16();
				const submesh = r.u16();
				r.u16(); // geoset index, the same as submesh
				const color = r.i16();
				const material = r.u16();
				const materialLayer = r.u16();
				const textureCount = r.u16();
				const textureCombo = r.u16();
				const coordCombo = r.u16();
				const weightCombo = r.u16();
				const transformCombo = r.u16();
				return { flags, priority, shader, submesh, color, material, materialLayer, textureCount, textureCombo, coordCombo, weightCombo, transformCombo };
			}),
		);
	}
}

// animData pulls the keys out of an .anim file. Old style files are the keys themselves. Chunked
// ones (models with flag 0x200000 or a .skel) keep them in AFM2, or AFSB for skeleton bones.
export function animData(buf: ArrayBuffer, chunked: boolean): DataView {
	if (!chunked) return new DataView(buf);
	const r = new Reader(buf);
	let found: DataView | undefined;
	while (r.remaining >= 8) {
		const id = r.fourcc();
		const size = r.u32();
		if (id === 'AFSB' || (id === 'AFM2' && !found)) found = new DataView(buf, r.pos, size);
		r.pos += size;
	}
	return found ?? new DataView(buf);
}
