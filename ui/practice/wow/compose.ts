// Character textures: the client paints the skin, face, underwear and every armor piece's texture
// into one body texture, section by section (CharComponentTextureSections). The server gives the
// layers in draw order (cmd/practice/appearance.go), and we paint them here.

import * as THREE from 'three';
import { image } from './files';
import type { Image } from './blp';

export type Layer = { fdid: number; x: number; y: number; w: number; h: number; blendMode: number; order?: number; alpha?: boolean };
export type Composite = { width: number; height: number; layers: Layer[] };

// scaled gives an image's pixels at the section's size, since HD textures can be larger than the
// section they fill.
function scaled(img: Image, w: number, h: number): Uint8ClampedArray {
	if (img.width === w && img.height === h) return img.rgba;
	const src = document.createElement('canvas');
	src.width = img.width;
	src.height = img.height;
	src.getContext('2d')!.putImageData(new ImageData(img.rgba, img.width, img.height), 0, 0);
	const dst = document.createElement('canvas');
	dst.width = w;
	dst.height = h;
	const ctx = dst.getContext('2d')!;
	ctx.imageSmoothingQuality = 'high';
	ctx.drawImage(src, 0, 0, w, h);
	return ctx.getImageData(0, 0, w, h).data;
}

// compose paints the layers. The blend modes are ChrModelTextureLayer's: 0 and 1 copy the
// layer, 4 multiplies, 6 overlays, 7 screens, and the rest blend by the layer's alpha.
//
// Units that look alike (the raid in your gear) ask for the same picture, so they share one
// texture. Then the crowd (crowd.ts) can draw their parts together.
const composed = new Map<string, Promise<THREE.Texture>>();

export function compose(c: Composite): Promise<THREE.Texture> {
	const key = JSON.stringify(c);
	let p = composed.get(key);
	if (!p) {
		p = paint(c).then(t => {
			t.userData.key = key;
			return t;
		});
		composed.set(key, p);
		p.catch(() => composed.delete(key));
	}
	// refs counts the units given the texture, so the crowd knows when it holds every one of them.
	return p.then(t => {
		t.userData.refs = (t.userData.refs ?? 0) + 1;
		return t;
	});
}

// forget drops a painted skin from the shared ones, once the crowd took its picture (and let go
// of it), so a unit asking for it later gets a new one.
export function forget(t: THREE.Texture) {
	composed.delete(t.userData.key);
}

async function paint(c: Composite): Promise<THREE.Texture> {
	const out = new Uint8ClampedArray(c.width * c.height * 4);
	// The client starts from mid grey where nothing paints.
	for (let i = 0; i < out.length; i += 4) {
		out[i] = out[i + 1] = out[i + 2] = 128;
		out[i + 3] = 255;
	}
	const layers = [...c.layers].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
	const images = await Promise.all(layers.map(l => image(l.fdid).catch(() => undefined)));
	layers.forEach((l, i) => {
		const img = images[i];
		if (!img || l.w <= 0 || l.h <= 0) return;
		const px = scaled(img, l.w, l.h);
		for (let y = 0; y < l.h; y++) {
			const ty = l.y + y;
			if (ty < 0 || ty >= c.height) continue;
			for (let x = 0; x < l.w; x++) {
				const tx = l.x + x;
				if (tx < 0 || tx >= c.width) continue;
				const s = (y * l.w + x) * 4;
				const d = (ty * c.width + tx) * 4;
				const a = l.alpha === false ? 1 : px[s + 3] / 255;
				switch (l.blendMode) {
					case 0:
					case 1:
						out[d] = px[s];
						out[d + 1] = px[s + 1];
						out[d + 2] = px[s + 2];
						out[d + 3] = px[s + 3];
						break;
					case 4:
						for (let k = 0; k < 3; k++) out[d + k] = out[d + k] * (1 - a) + ((out[d + k] * px[s + k]) / 255) * a;
						break;
					case 6:
						for (let k = 0; k < 3; k++) {
							const base = out[d + k] / 255;
							const top = px[s + k] / 255;
							const v = base < 0.5 ? 2 * base * top : 1 - 2 * (1 - base) * (1 - top);
							out[d + k] = out[d + k] * (1 - a) + v * 255 * a;
						}
						break;
					case 7:
						for (let k = 0; k < 3; k++) {
							const v = 255 - ((255 - out[d + k]) * (255 - px[s + k])) / 255;
							out[d + k] = out[d + k] * (1 - a) + v * a;
						}
						break;
					default:
						for (let k = 0; k < 3; k++) out[d + k] = out[d + k] * (1 - a) + px[s + k] * a;
						out[d + 3] = Math.max(out[d + 3], px[s + 3]);
				}
			}
		}
	});
	const t = new THREE.DataTexture(out, c.width, c.height, THREE.RGBAFormat);
	t.colorSpace = THREE.SRGBColorSpace;
	// The crowd (crowd.ts) puts painted skins in a texture array.
	t.userData.composed = true;
	t.wrapS = t.wrapT = THREE.ClampToEdgeWrapping;
	t.magFilter = THREE.LinearFilter;
	t.minFilter = THREE.LinearMipmapLinearFilter;
	t.generateMipmaps = true;
	t.anisotropy = 4;
	t.needsUpdate = true;
	// Once the texture is on the GPU we let go of our copy. Each unit's skin and armor makes its
	// own, so in a raid of 40 they were over half a gigabyte of the page's memory.
	t.onUpdate = () => {
		(t.image as { data: Uint8ClampedArray | null }).data = null;
	};
	return t;
}
