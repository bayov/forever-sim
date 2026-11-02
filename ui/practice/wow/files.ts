// The client's files, through the practice server's asset route (cmd/practice/assets.go). Each
// file comes once per page and decoded textures are shared.

import * as THREE from 'three';
import { decodeBLP, Image } from './blp';

const files = new Map<number, Promise<ArrayBuffer>>();

export function file(fdid: number): Promise<ArrayBuffer> {
	let p = files.get(fdid);
	if (!p) {
		p = fetch(`/api/assets/${fdid}`).then(async r => {
			if (!r.ok) {
				const body = await r.json().catch(() => ({}));
				throw new Error(`file ${fdid}: ${body.error ?? r.status}`);
			}
			return r.arrayBuffer();
		});
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

export function image(fdid: number): Promise<Image> {
	let p = images.get(fdid);
	if (!p) {
		p = file(fdid).then(decodeBLP);
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

export function imageTexture(img: Image, wrapU = true, wrapV = true): THREE.Texture {
	const t = new THREE.DataTexture(img.rgba, img.width, img.height, THREE.RGBAFormat);
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
