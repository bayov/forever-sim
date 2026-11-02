// BLP2, WoW's texture format: a palette, DXT1/3/5 blocks or plain BGRA. We decode the full size
// image to RGBA bytes, since we paint some textures together on a canvas before drawing them.

export type Image = { width: number; height: number; rgba: Uint8ClampedArray };

export function decodeBLP(buf: ArrayBuffer): Image {
	const v = new DataView(buf);
	const magic = String.fromCharCode(v.getUint8(0), v.getUint8(1), v.getUint8(2), v.getUint8(3));
	if (magic !== 'BLP2') throw new Error('not a BLP2 texture');
	const encoding = v.getUint8(8);
	const alphaDepth = v.getUint8(9);
	const alphaType = v.getUint8(10);
	const width = v.getUint32(12, true);
	const height = v.getUint32(16, true);
	const offset = v.getUint32(20, true);
	const size = v.getUint32(20 + 64, true);
	const data = new Uint8Array(buf, offset, Math.min(size, buf.byteLength - offset));
	const rgba = new Uint8ClampedArray(width * height * 4);
	switch (encoding) {
		case 1:
			decodePalette(v, data, width, height, alphaDepth, rgba);
			break;
		case 2:
			if (alphaType === 7) decodeDXT(data, width, height, 5, rgba);
			else if (alphaType === 1 && alphaDepth > 1) decodeDXT(data, width, height, 3, rgba);
			else decodeDXT(data, width, height, alphaDepth > 0 ? 1 : 0, rgba);
			break;
		case 3:
			for (let i = 0; i < width * height; i++) {
				rgba[i * 4] = data[i * 4 + 2];
				rgba[i * 4 + 1] = data[i * 4 + 1];
				rgba[i * 4 + 2] = data[i * 4];
				rgba[i * 4 + 3] = data[i * 4 + 3];
			}
			break;
		default:
			throw new Error(`BLP encoding ${encoding}`);
	}
	return { width, height, rgba };
}

function decodePalette(v: DataView, data: Uint8Array, width: number, height: number, alphaDepth: number, out: Uint8ClampedArray) {
	const n = width * height;
	// The palette is 256 BGRA colors after the header.
	const palette = 148;
	for (let i = 0; i < n; i++) {
		const c = palette + data[i] * 4;
		out[i * 4] = v.getUint8(c + 2);
		out[i * 4 + 1] = v.getUint8(c + 1);
		out[i * 4 + 2] = v.getUint8(c);
		let a = 255;
		if (alphaDepth === 8) a = data[n + i];
		else if (alphaDepth === 4) a = ((data[n + (i >> 1)] >> ((i & 1) * 4)) & 0xf) * 17;
		else if (alphaDepth === 1) a = (data[n + (i >> 3)] >> (i & 7)) & 1 ? 255 : 0;
		out[i * 4 + 3] = a;
	}
}

function rgb565(c: number): [number, number, number] {
	const r = (c >> 11) & 0x1f;
	const g = (c >> 5) & 0x3f;
	const b = c & 0x1f;
	return [(r << 3) | (r >> 2), (g << 2) | (g >> 4), (b << 3) | (b >> 2)];
}

// decodeDXT decodes DXT1 (kind 1 with one-bit alpha, 0 without), DXT3 or DXT5 blocks.
function decodeDXT(data: Uint8Array, width: number, height: number, kind: number, out: Uint8ClampedArray) {
	const blockSize = kind === 3 || kind === 5 ? 16 : 8;
	const bw = Math.max(1, (width + 3) >> 2);
	const bh = Math.max(1, (height + 3) >> 2);
	const colors = new Uint8Array(16);
	const alphas = new Uint8Array(16);
	for (let by = 0; by < bh; by++) {
		for (let bx = 0; bx < bw; bx++) {
			const o = (by * bw + bx) * blockSize;
			if (o + blockSize > data.length) return;
			let co = o;
			alphas.fill(255);
			if (kind === 3) {
				for (let i = 0; i < 16; i++) alphas[i] = ((data[o + (i >> 1)] >> ((i & 1) * 4)) & 0xf) * 17;
				co = o + 8;
			} else if (kind === 5) {
				const a0 = data[o];
				const a1 = data[o + 1];
				const table = [a0, a1, 0, 0, 0, 0, 0, 0];
				if (a0 > a1) for (let i = 1; i < 7; i++) table[i + 1] = ((7 - i) * a0 + i * a1) / 7;
				else {
					for (let i = 1; i < 5; i++) table[i + 1] = ((5 - i) * a0 + i * a1) / 5;
					table[6] = 0;
					table[7] = 255;
				}
				// 48 bits of 3-bit indices.
				let bits = 0;
				for (let i = 0; i < 6; i++) bits += data[o + 2 + i] * 2 ** (8 * i);
				for (let i = 0; i < 16; i++) {
					alphas[i] = table[Math.floor(bits / 2 ** (3 * i)) & 7];
				}
				co = o + 8;
			}
			const c0 = data[co] | (data[co + 1] << 8);
			const c1 = data[co + 2] | (data[co + 3] << 8);
			const p0 = rgb565(c0);
			const p1 = rgb565(c1);
			const palette: [number, number, number, number][] = [
				[...p0, 255],
				[...p1, 255],
				[0, 0, 0, 255],
				[0, 0, 0, 255],
			];
			// DXT1 has a three color mode (c0 <= c1) with or without alpha. DXT3 and 5 never do.
			if (c0 > c1 || kind === 3 || kind === 5) {
				for (let k = 0; k < 3; k++) {
					palette[2][k] = (2 * p0[k] + p1[k]) / 3;
					palette[3][k] = (p0[k] + 2 * p1[k]) / 3;
				}
			} else {
				for (let k = 0; k < 3; k++) palette[2][k] = (p0[k] + p1[k]) / 2;
				palette[3] = [0, 0, 0, kind === 0 ? 255 : 0];
			}
			const idx = data[co + 4] | (data[co + 5] << 8) | (data[co + 6] << 16) | (data[co + 7] << 24);
			for (let i = 0; i < 16; i++) colors[i] = (idx >>> (2 * i)) & 3;
			for (let py = 0; py < 4; py++) {
				const y = by * 4 + py;
				if (y >= height) break;
				for (let px = 0; px < 4; px++) {
					const x = bx * 4 + px;
					if (x >= width) break;
					const i = py * 4 + px;
					const c = palette[colors[i]];
					const p = (y * width + x) * 4;
					out[p] = c[0];
					out[p + 1] = c[1];
					out[p + 2] = c[2];
					out[p + 3] = kind === 1 ? c[3] : kind === 0 ? 255 : alphas[i];
				}
			}
		}
	}
}
