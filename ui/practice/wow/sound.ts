// The game's sounds: OGG files from the client, played through the camera's listener so a sound
// at the dummy comes from the dummy.
//
// A sound kit lists several files, and the client plays one of them at random each time, so a
// swing doesn't sound the same twice in a row.

import * as THREE from 'three';
import { file } from './files';

export type SoundRef = { fdid: number; volume?: number };

const buffers = new Map<number, Promise<AudioBuffer | undefined>>();

export class Sounds {
	readonly listener = new THREE.AudioListener();
	volume = 0.6;
	private last = new Map<string, number>();

	constructor(camera: THREE.Camera) {
		camera.add(this.listener);
		this.listener.setMasterVolume(this.volume);
		// Browsers start audio only after a key press or click.
		const resume = () => {
			if (this.listener.context.state !== 'running') this.listener.context.resume();
		};
		addEventListener('keydown', resume);
		addEventListener('pointerdown', resume);
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
	preload(kit: SoundRef[] | undefined) {
		for (const s of kit ?? []) this.buffer(s.fdid);
	}

	// play plays one sound of a kit, at an object in the world or flat when there's none. The same
	// kit doesn't stack within 40ms (two hits in one batch).
	async play(kit: SoundRef[] | undefined, at?: THREE.Object3D, key?: string) {
		if (!kit || kit.length === 0 || this.listener.context.state !== 'running') return;
		const now = performance.now();
		const k = key ?? String(kit[0].fdid);
		if (now - (this.last.get(k) ?? -1e9) < 40) return;
		this.last.set(k, now);
		const pick = kit[Math.floor(Math.random() * kit.length)];
		const buffer = await this.buffer(pick.fdid);
		if (!buffer) return;
		const audio = at ? new THREE.PositionalAudio(this.listener) : new THREE.Audio(this.listener);
		audio.setBuffer(buffer);
		audio.setVolume(pick.volume ?? 1);
		if (audio instanceof THREE.PositionalAudio) {
			audio.setRefDistance(8);
			audio.setRolloffFactor(1);
			at!.add(audio);
		}
		audio.onEnded = () => {
			audio.removeFromParent();
			audio.disconnect();
		};
		audio.play();
	}
}
