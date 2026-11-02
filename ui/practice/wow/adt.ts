// Readers for WoW's map files: the map's WDT (which files hold each 533 yard tile) and a tile's
// ADT files (the root file has the heights, tex0 the ground textures, obj0 the doodads).
//
// We follow the format pages on wowdev.wiki (WDT, ADT/v18). Map coordinates are the world's: x
// north, y west, z up. A map is 64 by 64 tiles, a tile is 16 by 16 chunks, and a chunk has a 9 by 9
// grid of heights with an 8 by 8 grid of heights in the middle of its cells (145 in all).

export const TILE = 1600 / 3;
export const CHUNK = TILE / 16;
const UNIT = CHUNK / 8;
// The map's middle is the world's (0, 0). Placement coordinates count from the map's corner.
const HALF = 32 * TILE;

export type TileFiles = { root: number; obj0: number; tex0: number };

export type Wdt = { flags: number; tiles: Map<number, TileFiles> };

export type TerrainChunk = {
	// The chunk's corner (its north west, the largest x and y) and its base height.
	x: number;
	y: number;
	z: number;
	// 145 heights over z: 9 outer then 8 inner, row by row.
	heights: Float32Array;
	// 145 vertex colors (BGRA, 127 is 1), when the map has them.
	colors?: Uint8Array;
	// The ground textures (file ids), base first. Each later one has an alpha map.
	layers: number[];
	// 64 by 64 alpha, RGB for layers 1 to 3.
	alpha: Uint8Array;
};

export type Doodad = { fdid: number; x: number; y: number; z: number; rot: [number, number, number]; scale: number };

function chunks(buf: ArrayBuffer, start = 0, end = buf.byteLength): { id: string; at: number; size: number }[] {
	const v = new DataView(buf);
	const out: { id: string; at: number; size: number }[] = [];
	for (let p = start; p + 8 <= end; ) {
		// Chunk ids are stored reversed ("REVM" for MVER).
		const id = String.fromCharCode(v.getUint8(p + 3), v.getUint8(p + 2), v.getUint8(p + 1), v.getUint8(p));
		const size = v.getUint32(p + 4, true);
		out.push({ id, at: p + 8, size });
		p += 8 + size;
	}
	return out;
}

export function tileIndex(x: number, y: number): number {
	return y * 64 + x;
}

// tileAt gives the tile a world point is on, as [x, y] in the WDT's grid.
export function tileAt(worldX: number, worldY: number): [number, number] {
	return [Math.floor(32 - worldY / TILE), Math.floor(32 - worldX / TILE)];
}

export function parseWdt(buf: ArrayBuffer): Wdt {
	const v = new DataView(buf);
	let flags = 0;
	const tiles = new Map<number, TileFiles>();
	for (const c of chunks(buf)) {
		if (c.id === 'MPHD') flags = v.getUint32(c.at, true);
		if (c.id === 'MAID') {
			for (let i = 0; i < 4096; i++) {
				const o = c.at + i * 32;
				const root = v.getUint32(o, true);
				if (root) tiles.set(i, { root, obj0: v.getUint32(o + 4, true), tex0: v.getUint32(o + 12, true) });
			}
		}
	}
	return { flags, tiles };
}

// parseTerrain reads a tile's chunks from its root and tex0 files. `bigAlpha` is the WDT's flag
// 0x4 or 0x80: alpha maps of 8 bits a pixel.
export function parseTerrain(root: ArrayBuffer, tex0: ArrayBuffer, bigAlpha: boolean): TerrainChunk[] {
	const v = new DataView(root);
	const out: TerrainChunk[] = [];
	for (const c of chunks(root)) {
		if (c.id !== 'MCNK') continue;
		const h = c.at;
		const chunk: TerrainChunk = {
			x: v.getFloat32(h + 0x68, true),
			y: v.getFloat32(h + 0x6c, true),
			z: v.getFloat32(h + 0x70, true),
			heights: new Float32Array(145),
			layers: [],
			alpha: new Uint8Array(64 * 64 * 4),
		};
		for (const s of chunks(root, h + 128, c.at + c.size)) {
			if (s.id === 'MCVT') for (let i = 0; i < 145; i++) chunk.heights[i] = v.getFloat32(s.at + i * 4, true);
			if (s.id === 'MCCV') chunk.colors = new Uint8Array(root.slice(s.at, s.at + 145 * 4));
		}
		out.push(chunk);
	}

	const t = new DataView(tex0);
	const textures: number[] = [];
	let n = 0;
	for (const c of chunks(tex0)) {
		if (c.id === 'MDID') for (let i = 0; i + 4 <= c.size; i += 4) textures.push(t.getUint32(c.at + i, true));
		if (c.id !== 'MCNK') continue;
		const chunk = out[n++];
		if (!chunk) break;
		const sub = chunks(tex0, c.at, c.at + c.size);
		const mcly = sub.find(s => s.id === 'MCLY');
		const mcal = sub.find(s => s.id === 'MCAL');
		if (!mcly) continue;
		for (let l = 0; l * 16 < mcly.size && l < 4; l++) {
			const o = mcly.at + l * 16;
			chunk.layers.push(textures[t.getUint32(o, true)] ?? 0);
			const flags = t.getUint32(o + 4, true);
			if (l === 0 || !mcal || !(flags & 0x100)) continue;
			const alpha = readAlpha(tex0, mcal.at + t.getUint32(o + 8, true), flags & 0x200 ? 'rle' : bigAlpha ? 'big' : 'small');
			for (let i = 0; i < 4096; i++) chunk.alpha[i * 4 + l - 1] = alpha[i];
		}
	}
	return out;
}

// readAlpha reads one layer's 64 by 64 alpha map.
function readAlpha(buf: ArrayBuffer, at: number, kind: 'rle' | 'big' | 'small'): Uint8Array {
	const src = new Uint8Array(buf);
	const out = new Uint8Array(4096);
	if (kind === 'big') out.set(src.subarray(at, at + 4096));
	else if (kind === 'small') {
		for (let i = 0; i < 2048; i++) {
			const b = src[at + i];
			out[i * 2] = (b & 0xf) * 17;
			out[i * 2 + 1] = (b >> 4) * 17;
		}
	} else {
		// Run lengths: the top bit says fill (one value repeated) or copy, the rest is the count.
		let p = at;
		for (let o = 0; o < 4096; ) {
			const head = src[p++];
			const count = head & 0x7f;
			if (head & 0x80) {
				const value = src[p++];
				for (let i = 0; i < count && o < 4096; i++) out[o++] = value;
			} else for (let i = 0; i < count && o < 4096; i++) out[o++] = src[p++];
		}
	}
	return out;
}

// parseDoodads reads a tile's doodads (M2 models) from its obj0 file. We keep the ones named by
// file id (flag 0x40), which is every one in the current clients.
export function parseDoodads(obj0: ArrayBuffer): Doodad[] {
	const v = new DataView(obj0);
	const out: Doodad[] = [];
	for (const c of chunks(obj0)) {
		if (c.id !== 'MDDF') continue;
		for (let o = c.at; o + 36 <= c.at + c.size; o += 36) {
			const flags = v.getUint16(o + 34, true);
			if (!(flags & 0x40)) continue;
			const f = (k: number) => v.getFloat32(o + k, true);
			out.push({
				fdid: v.getUint32(o, true),
				// Placement coordinates: x runs east from the map's corner, y is up, z runs south.
				x: HALF - f(16),
				y: HALF - f(8),
				z: f(12),
				rot: [f(20), f(24), f(28)],
				scale: v.getUint16(o + 32, true) / 1024,
			});
		}
	}
	return out;
}

// heightAt gives a chunk's vertex position in world coordinates.
export function vertexAt(c: TerrainChunk, i: number): [number, number, number] {
	const row = Math.floor(i / 17);
	const col = i % 17;
	const inner = col >= 9;
	const r = inner ? row + 0.5 : row;
	const k = inner ? col - 9 + 0.5 : col;
	return [c.x - r * UNIT, c.y - k * UNIT, c.z + c.heights[i]];
}
