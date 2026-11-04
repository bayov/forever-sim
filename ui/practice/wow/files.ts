// The client's files, through the practice server's asset route (cmd/practice/assets.go). Each
// file comes once per page and decoded textures are shared.

import * as THREE from 'three';
import { decodeBLP, Image } from './blp';

const files = new Map<number, Promise<ArrayBuffer>>();

// fileProgress counts the files asked for and the ones that came back, for the loading bar. ids
// lists the files in the order we first asked for them.
export const fileProgress = { asked: 0, done: 0, ids: [] as number[] };

export function hasFile(fdid: number): boolean {
	return files.has(fdid);
}

export function file(fdid: number): Promise<ArrayBuffer> {
	let p = files.get(fdid);
	if (!p) {
		fileProgress.asked++;
		fileProgress.ids.push(fdid);
		p = fetch(`/api/assets/${fdid}`)
			.then(async r => {
				if (!r.ok) {
					const body = await r.json().catch(() => ({}));
					throw new Error(`file ${fdid}: ${body.error ?? r.status}`);
				}
				return r.arrayBuffer();
			})
			.finally(() => fileProgress.done++);
		files.set(fdid, p);
	}
	return p;
}

// prefetch asks the server to extract files in one go, before we read them one by one.
export async function prefetch(ids: number[]): Promise<void> {
	const want = [...new Set(ids.filter(id => id > 0 && !files.has(id)))];
	if (want.length === 0) return;
	await fetch('/api/assets/batch', { method: 'POST', body: JSON.stringify(want) }).catch(() => undefined);
}

const images = new Map<number, Promise<Image>>();

// onDecode hears how long each texture took to decode, for the performance recorder.
export const decodeHooks: { onDecode?: (fdid: number, ms: number) => void } = {};

export function image(fdid: number): Promise<Image> {
	let p = images.get(fdid);
	if (!p) {
		p = file(fdid).then(buf => {
			const t = performance.now();
			const img = decodeBLP(buf);
			decodeHooks.onDecode?.(fdid, performance.now() - t);
			return img;
		});
		images.set(fdid, p);
	}
	return p;
}

const textures = new Map<number, Promise<THREE.Texture>>();

// texture gives a texture for a BLP file. WoW's texture coordinates start at the image's top row,
// which is row 0 of the data, so the texture isn't flipped.
export function texture(fdid: number, wrapU = true, wrapV = true): Promise<THREE.Texture> {
	const key = fdid * 4 + (wrapU ? 2 : 0) + (wrapV ? 1 : 0);
	let p = textures.get(key);
	if (!p) {
		p = image(fdid).then(img => imageTexture(img, wrapU, wrapV));
		textures.set(key, p);
	}
	return p;
}

// newTextures holds the textures made while the room loads, for it to send them to the GPU before
// the fight. Once the room is in, `keep` is off and we stop holding them.
export const newTextures = { keep: true, list: [] as THREE.Texture[] };

export function imageTexture(img: Image, wrapU = true, wrapV = true): THREE.Texture {
	const t = new THREE.DataTexture(img.rgba, img.width, img.height, THREE.RGBAFormat);
	if (newTextures.keep) newTextures.list.push(t);
	t.colorSpace = THREE.SRGBColorSpace;
	t.wrapS = wrapU ? THREE.RepeatWrapping : THREE.ClampToEdgeWrapping;
	t.wrapT = wrapV ? THREE.RepeatWrapping : THREE.ClampToEdgeWrapping;
	t.magFilter = THREE.LinearFilter;
	t.minFilter = THREE.LinearMipmapLinearFilter;
	t.generateMipmaps = true;
	t.anisotropy = 4;
	t.needsUpdate = true;
	return t;
}
