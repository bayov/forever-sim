// The game's sounds: OGG files from the client, mixed the way the client's sound engine does.
//
// The listener stands at your character and turns with the camera (the client's
// Sound_ListenerAtCharacter). Your own sounds (swings, casts, steps) play flat, with no 3D
// position. Sounds away from you pan left and right and fade with distance, with plain stereo
// panning: the browser's HRTF mode muffles these 22 kHz files and colors them.
//
// A sound kit lists several files, and the client plays one of them at random each time, with a
// little pitch variation, so a swing doesn't sound the same twice in a row.

import * as THREE from 'three';
import { file } from './files';

export type SoundRef = {
	fdid: number;
	volume?: number;
	kit?: number;
	weight?: number;
	loop?: boolean;
	minDistance?: number;
	cutoff?: number;
	pitchPlus?: number;
	pitchMinus?: number;
	volumeMinus?: number;
	inner2D?: number;
	cooldownMs?: [number, number];
	fadeInMs?: number;
	fadeOutMs?: number;
};

// A playing sound, which a loop needs to stop.
export type Playing = { stop(fadeMs?: number): void };

const buffers = new Map<number, Promise<AudioBuffer | undefined>>();

export class Sounds {
	readonly listener = new THREE.AudioListener();
	// The listener's spot: at your character, turned like the camera.
	readonly anchor = new THREE.Object3D();
	volume = 0.6;
	// Your character: sounds on it play flat.
	self?: THREE.Object3D;
	private last = new Map<string, number>();
	private ready = new Map<string, number>();
	private moving: { panner: PannerNode; at: THREE.Object3D; done: boolean }[] = [];
	private readonly tmp = new THREE.Vector3();

	constructor(scene: THREE.Object3D) {
		this.anchor.add(this.listener);
		scene.add(this.anchor);
		this.listener.setMasterVolume(this.volume);
		// A limiter after the master volume, so a swing, an impact, a cast and a grunt landing
		// together don't clip.
		const c = this.listener.context;
		const limiter = c.createDynamicsCompressor();
		limiter.threshold.value = -3;
		limiter.knee.value = 0;
		limiter.ratio.value = 20;
		limiter.attack.value = 0.002;
		limiter.release.value = 0.1;
		this.listener.gain.disconnect();
		this.listener.gain.connect(limiter).connect(c.destination);
		// Browsers start audio only after a key press or click.
		const resume = () => {
			if (c.state !== 'running') c.resume();
		};
		addEventListener('keydown', resume);
		addEventListener('pointerdown', resume);
	}

	// place moves the listener: to your character's head, turned like the camera.
	place(pos: THREE.Vector3, camera: THREE.Camera) {
		this.anchor.position.copy(pos);
		camera.getWorldQuaternion(this.anchor.quaternion);
		this.anchor.updateMatrixWorld(true);
		// Sounds on moving things (a bolt in flight) follow them.
		const t = this.listener.context.currentTime;
		this.moving = this.moving.filter(m => !m.done);
		for (const m of this.moving) {
			m.at.getWorldPosition(this.tmp);
			m.panner.positionX.setTargetAtTime(this.tmp.x, t, 0.02);
			m.panner.positionY.setTargetAtTime(this.tmp.y, t, 0.02);
			m.panner.positionZ.setTargetAtTime(this.tmp.z, t, 0.02);
		}
	}

	private buffer(fdid: number): Promise<AudioBuffer | undefined> {
		let p = buffers.get(fdid);
		if (!p) {
			p = file(fdid)
				.then(buf => this.listener.context.decodeAudioData(buf))
				.catch(e => {
					console.warn(`sound ${fdid}`, e);
					return undefined;
				});
			buffers.set(fdid, p);
		}
		return p;
	}

	// preload decodes a kit's files ahead of their first play.
	async preload(kit: SoundRef[] | undefined) {
		await Promise.all((kit ?? []).map(s => this.buffer(s.fdid)));
	}

	private isSelf(at: THREE.Object3D): boolean {
		for (let o: THREE.Object3D | null = at; o; o = o.parent) if (o === this.self) return true;
		return false;
	}

	// play plays one sound of a kit, at an object in the world or flat when there's none. The same
	// kit doesn't stack within 40ms (two hits in one batch), or within its own cooldown. It gives
	// a handle to stop it (for a loop).
	async play(kit: SoundRef[] | undefined, at?: THREE.Object3D, key?: string): Promise<Playing | undefined> {
		const c = this.listener.context;
		if (!kit || kit.length === 0 || c.state !== 'running') return undefined;
		const now = performance.now();
		const k = key ?? String(kit[0].kit ?? kit[0].fdid);
		if (now - (this.last.get(k) ?? -1e9) < 40 || now < (this.ready.get(k) ?? 0)) return undefined;
		this.last.set(k, now);
		const cd = kit[0].cooldownMs;
		if (cd && cd[1] > 0) this.ready.set(k, now + cd[0] + Math.random() * (cd[1] - cd[0]));
		const pick = pickWeighted(kit);
		const buffer = await this.buffer(pick.fdid);
		if (!buffer) return undefined;

		const src = c.createBufferSource();
		src.buffer = buffer;
		src.loop = !!pick.loop;
		src.playbackRate.value = 1 + (Math.random() * ((pick.pitchPlus ?? 0) + (pick.pitchMinus ?? 0)) - (pick.pitchMinus ?? 0));
		const gain = c.createGain();
		const volume = (pick.volume ?? 1) * (1 - Math.random() * (pick.volumeMinus ?? 0));
		const t = c.currentTime;
		if (pick.fadeInMs) {
			gain.gain.setValueAtTime(0, t);
			gain.gain.linearRampToValueAtTime(volume, t + pick.fadeInMs / 1000);
		} else gain.gain.value = volume;
		src.connect(gain);
		let tail: AudioNode = gain;
		let moving: { panner: PannerNode; at: THREE.Object3D; done: boolean } | undefined;
		if (at && !this.isSelf(at)) {
			at.getWorldPosition(this.tmp);
			const near = this.tmp.distanceTo(this.anchor.position) <= (pick.inner2D ?? 0);
			if (!near) {
				const p = c.createPanner();
				p.panningModel = 'equalpower';
				p.distanceModel = 'inverse';
				p.refDistance = pick.minDistance || 8;
				p.maxDistance = pick.cutoff || 45;
				p.rolloffFactor = 1;
				// Set where it is before it starts, so its first moment doesn't come from the middle.
				p.positionX.setValueAtTime(this.tmp.x, t);
				p.positionY.setValueAtTime(this.tmp.y, t);
				p.positionZ.setValueAtTime(this.tmp.z, t);
				gain.connect(p);
				tail = p;
				moving = { panner: p, at, done: false };
				this.moving.push(moving);
			}
		}
		tail.connect(this.listener.getInput());
		src.start(t);
		src.onended = () => {
			if (moving) moving.done = true;
			src.disconnect();
			gain.disconnect();
			tail.disconnect();
		};
		return {
			stop: (fadeMs = pick.fadeOutMs || 200) => {
				const now = c.currentTime;
				gain.gain.cancelScheduledValues(now);
				gain.gain.setValueAtTime(gain.gain.value, now);
				gain.gain.linearRampToValueAtTime(0, now + fadeMs / 1000);
				try {
					src.stop(now + fadeMs / 1000 + 0.01);
				} catch {
					// Already stopped.
				}
			},
		};
	}
}

// pickWeighted picks a kit's file by the entries' weights.
function pickWeighted(kit: SoundRef[]): SoundRef {
	const total = kit.reduce((a, s) => a + Math.max(0, s.weight ?? 1), 0);
	if (total <= 0) return kit[Math.floor(Math.random() * kit.length)];
	let roll = Math.random() * total;
	for (const s of kit) {
		roll -= Math.max(0, s.weight ?? 1);
		if (roll < 0) return s;
	}
	return kit[kit.length - 1];
}
