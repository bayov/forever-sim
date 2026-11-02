// The room's ground: a piece of a real map from the client, with its ground textures and the
// doodads (trees, rocks, bushes) that stand on it.
//
// The fight is on flat ground, so we flatten the map under the play area and blend into the real
// hills further out. A spot that is already flat is best (Terrain.load's caller picks one).

import * as THREE from 'three';
import { CHUNK, parseDoodads, parseTerrain, parseWdt, TerrainChunk, tileAt, tileIndex, vertexAt } from './adt';
import { file, image } from './files';
import { Model } from './model';

export type TerrainSpot = {
	// The map's WDT file.
	wdt: number;
	// The world point the room's middle (the dummy) stands on.
	x: number;
	y: number;
	// How far out we draw the ground and the doodads.
	radius: number;
	// The ground is flat out to `flat`, then blends into the real heights by `blend`.
	flat: number;
	blend: number;
	// No doodads stand this close to the middle, so nothing blocks the fight.
	clear: number;
};

export class Terrain {
	readonly group = new THREE.Group();
	readonly doodads: Model[] = [];

	private constructor(readonly spot: TerrainSpot) {
		this.group.name = 'terrain';
	}

	// load reads the tiles around the spot and builds the ground. The doodads load after, through
	// loadDoodads.
	static async load(spot: TerrainSpot): Promise<Terrain> {
		const t = new Terrain(spot);
		const wdt = parseWdt(await file(spot.wdt));
		const bigAlpha = (wdt.flags & 0x84) !== 0;
		const [x0, y0] = tileAt(spot.x + spot.radius, spot.y + spot.radius);
		const [x1, y1] = tileAt(spot.x - spot.radius, spot.y - spot.radius);
		const chunks: TerrainChunk[] = [];
		const doodadLists = await Promise.all(
			range(x0, x1).flatMap(tx =>
				range(y0, y1).map(async ty => {
					const files = wdt.tiles.get(tileIndex(tx, ty));
					if (!files) return [];
					const [root, tex0, obj0] = await Promise.all([file(files.root), file(files.tex0), files.obj0 ? file(files.obj0) : undefined]);
					for (const c of parseTerrain(root, tex0, bigAlpha)) {
						// The chunk's middle is within the radius (with a chunk to spare).
						if (Math.hypot(c.x - CHUNK / 2 - spot.x, c.y - CHUNK / 2 - spot.y) < spot.radius + CHUNK) chunks.push(c);
					}
					return obj0 ? parseDoodads(obj0) : [];
				}),
			),
		);
		t.base = t.baseHeight(chunks);
		await t.buildGround(chunks);
		t.pending = doodadLists.flat().filter(d => {
			const r = Math.hypot(d.x - spot.x, d.y - spot.y);
			return r > spot.clear && r < spot.radius;
		});
		return t;
	}

	// The ground's height at the middle, which becomes the room's 0.
	private base = 0;
	private pending: { fdid: number; x: number; y: number; z: number; rot: [number, number, number]; scale: number }[] = [];

	private baseHeight(chunks: TerrainChunk[]): number {
		let sum = 0;
		let n = 0;
		for (const c of chunks) {
			for (let i = 0; i < 145; i++) {
				const [x, y, z] = vertexAt(c, i);
				if (Math.hypot(x - this.spot.x, y - this.spot.y) < this.spot.flat) {
					sum += z;
					n++;
				}
			}
		}
		return n ? sum / n : 0;
	}

	// height gives the ground's height in the room at a world point, flattened near the middle.
	private height(x: number, y: number, z: number): number {
		const d = Math.hypot(x - this.spot.x, y - this.spot.y);
		const k = smoothstep(this.spot.flat, this.spot.blend, d);
		return (z - this.base) * k;
	}

	// buildGround makes the ground one mesh, so it is one draw call. Every chunk's four textures
	// come from one texture array, and its alpha map from its cell in one alpha atlas.
	private async buildGround(chunks: TerrainChunk[]) {
		const spot = this.spot;
		const fdids = [...new Set(chunks.flatMap(c => c.layers.filter(id => id > 0)))];
		const layerOf = new Map(fdids.map((id, i) => [id, i]));
		const layers = await layerArray(fdids);

		// The alpha atlas: a grid of the chunks' 64 by 64 alpha maps.
		const across = Math.ceil(Math.sqrt(chunks.length));
		const size = across * 64;
		const atlas = new Uint8Array(size * size * 4);
		chunks.forEach((c, n) => {
			const ax = (n % across) * 64;
			const ay = Math.floor(n / across) * 64;
			for (let row = 0; row < 64; row++) atlas.set(c.alpha.subarray(row * 256, row * 256 + 256), ((ay + row) * size + ax) * 4);
		});
		const alpha = new THREE.DataTexture(atlas, size, size, THREE.RGBAFormat);
		alpha.magFilter = THREE.LinearFilter;
		alpha.minFilter = THREE.LinearFilter;
		alpha.needsUpdate = true;

		const count = chunks.length * 145;
		const pos = new Float32Array(count * 3);
		const color = new Float32Array(count * 3);
		const tileUv = new Float32Array(count * 2);
		const alphaUv = new Float32Array(count * 2);
		const layerIds = new Float32Array(count * 4);
		const index: number[] = [];
		const cells = cellIndex();
		chunks.forEach((c, n) => {
			const ax = (n % across) * 64;
			const ay = Math.floor(n / across) * 64;
			const ids = [0, 1, 2, 3].map(l => layerOf.get(c.layers[l]) ?? 0);
			for (let i = 0; i < 145; i++) {
				const v = n * 145 + i;
				const [x, y, z] = vertexAt(c, i);
				// World (x north, y west, z up) to Three (x, z, -y), from the room's middle.
				pos[v * 3] = x - spot.x;
				pos[v * 3 + 1] = this.height(x, y, z);
				pos[v * 3 + 2] = -(y - spot.y);
				const row = Math.floor(i / 17);
				const col = i % 17;
				const inner = col >= 9;
				const u = (inner ? col - 9 + 0.5 : col) / 8;
				const w = (inner ? row + 0.5 : row) / 8;
				// A texture repeats 8 times across a chunk.
				tileUv[v * 2] = u * 8;
				tileUv[v * 2 + 1] = w * 8;
				// Half a pixel in from the cell's edges, so a neighbor's alpha doesn't bleed in.
				alphaUv[v * 2] = (ax + 0.5 + u * 63) / size;
				alphaUv[v * 2 + 1] = (ay + 0.5 + w * 63) / size;
				for (let l = 0; l < 4; l++) layerIds[v * 4 + l] = ids[l];
				// Vertex colors are BGRA, with 127 as 1.
				for (let k = 0; k < 3; k++) color[v * 3 + k] = c.colors ? c.colors[i * 4 + 2 - k] / 127 : 1;
			}
			for (const k of cells) index.push(n * 145 + k);
		});
		const g = new THREE.BufferGeometry();
		g.setAttribute('position', new THREE.BufferAttribute(pos, 3));
		g.setAttribute('color', new THREE.BufferAttribute(color, 3));
		g.setAttribute('uv', new THREE.BufferAttribute(tileUv, 2));
		g.setAttribute('alphaUv', new THREE.BufferAttribute(alphaUv, 2));
		g.setAttribute('groundLayers', new THREE.BufferAttribute(layerIds, 4));
		g.setIndex(index);
		g.computeVertexNormals();
		smoothSeams(g);
		const mesh = new THREE.Mesh(g, groundMaterial(layers, alpha));
		mesh.receiveShadow = true;
		mesh.name = 'ground';
		this.group.add(mesh);
	}

	// loadDoodads loads the doodads and stands them on the ground. They don't animate, so each model
	// becomes a plain mesh in its first pose, and all copies of it are drawn together.
	async loadDoodads() {
		const spot = this.spot;
		const byFile = new Map<number, typeof this.pending>();
		for (const d of this.pending) byFile.set(d.fdid, [...(byFile.get(d.fdid) ?? []), d]);
		const matrix = new THREE.Matrix4();
		const q = new THREE.Quaternion();
		const euler = new THREE.Euler();
		const at = new THREE.Vector3();
		const scale = new THREE.Vector3();
		await Promise.all(
			[...byFile].map(async ([fdid, placed]) => {
				const m = await Model.load(fdid).catch(() => undefined);
				if (!m) return;
				const mesh = new THREE.InstancedMesh(posedGeometry(m), m.mesh.material, placed.length);
				placed.forEach((d, i) => {
					at.set(d.x - spot.x, this.height(d.x, d.y, d.z), -(d.y - spot.y));
					// Placement rotations are in degrees, around the placement axes (the client turns a
					// model by its y, then z, then x). Placement space is Three's turned 90 degrees.
					euler.set(THREE.MathUtils.degToRad(d.rot[2]), THREE.MathUtils.degToRad(d.rot[1] - 180), THREE.MathUtils.degToRad(-d.rot[0]), 'YZX');
					mesh.setMatrixAt(i, matrix.compose(at, q.setFromEuler(euler), scale.setScalar(d.scale)));
				});
				mesh.castShadow = true;
				mesh.onBeforeShadow = m.mesh.onBeforeShadow;
				mesh.onAfterShadow = m.mesh.onAfterShadow;
				mesh.computeBoundingSphere();
				this.group.add(mesh);
				this.doodads.push(m);
			}),
		);
		this.pending = [];
	}
}

// posedGeometry is a model's geometry in its first pose, without bones: each vertex where the
// skeleton puts it.
function posedGeometry(m: Model): THREE.BufferGeometry {
	m.update(0);
	m.root.updateMatrixWorld(true);
	m.mesh.skeleton.update();
	const src = m.mesh.geometry;
	const g = new THREE.BufferGeometry();
	for (const name of Object.keys(src.attributes)) if (name !== 'skinIndex' && name !== 'skinWeight') g.setAttribute(name, src.attributes[name]);
	const n = src.attributes.position.count;
	const pos = new Float32Array(n * 3);
	const v = new THREE.Vector3();
	for (let i = 0; i < n; i++) {
		m.mesh.getVertexPosition(i, v);
		pos[i * 3] = v.x;
		pos[i * 3 + 1] = v.y;
		pos[i * 3 + 2] = v.z;
	}
	g.setAttribute('position', new THREE.BufferAttribute(pos, 3));
	g.setIndex(src.index);
	for (const group of src.groups) g.addGroup(group.start, group.count, group.materialIndex);
	return g;
}

// layerArray puts the ground textures in one texture array, each scaled to the largest one's
// size.
async function layerArray(fdids: number[]): Promise<THREE.DataArrayTexture> {
	const images = await Promise.all(fdids.map(id => image(id).catch(() => undefined)));
	const size = Math.max(1, ...images.map(img => img?.width ?? 1));
	const data = new Uint8Array(size * size * 4 * Math.max(1, fdids.length));
	images.forEach((img, layer) => {
		if (!img) return;
		const base = layer * size * size * 4;
		for (let y = 0; y < size; y++) {
			const sy = Math.floor((y * img.height) / size);
			for (let x = 0; x < size; x++) {
				const sx = Math.floor((x * img.width) / size);
				const s = (sy * img.width + sx) * 4;
				const d = base + (y * size + x) * 4;
				data[d] = img.rgba[s];
				data[d + 1] = img.rgba[s + 1];
				data[d + 2] = img.rgba[s + 2];
				data[d + 3] = 255;
			}
		}
	});
	const t = new THREE.DataArrayTexture(data, size, size, Math.max(1, fdids.length));
	t.colorSpace = THREE.SRGBColorSpace;
	t.wrapS = t.wrapT = THREE.RepeatWrapping;
	t.magFilter = THREE.LinearFilter;
	t.minFilter = THREE.LinearMipmapLinearFilter;
	t.generateMipmaps = true;
	t.anisotropy = 4;
	t.needsUpdate = true;
	return t;
}

// The ground shader: Lambert lighting and shadows, with each chunk's four textures blended by its
// alpha map.
function groundMaterial(layers: THREE.DataArrayTexture, alpha: THREE.Texture): THREE.MeshLambertMaterial {
	const mat = new THREE.MeshLambertMaterial({ vertexColors: true });
	mat.onBeforeCompile = shader => {
		shader.uniforms.uLayers = { value: layers };
		shader.uniforms.uAlpha = { value: alpha };
		shader.vertexShader = shader.vertexShader
			.replace(
				'#include <common>',
				'#include <common>\nattribute vec2 alphaUv;\nattribute vec4 groundLayers;\nvarying vec2 vTileUv;\nvarying vec2 vAlphaUv;\nflat varying vec4 vLayers;',
			)
			.replace('#include <begin_vertex>', '#include <begin_vertex>\nvTileUv = uv;\nvAlphaUv = alphaUv;\nvLayers = groundLayers;');
		shader.fragmentShader = shader.fragmentShader
			.replace(
				'#include <common>',
				'#include <common>\nprecision highp sampler2DArray;\nuniform sampler2DArray uLayers;\nuniform sampler2D uAlpha;\nvarying vec2 vTileUv;\nvarying vec2 vAlphaUv;\nflat varying vec4 vLayers;',
			)
			.replace(
				'#include <map_fragment>',
				`vec4 groundAlpha = texture2D(uAlpha, vAlphaUv);
				vec3 ground = texture(uLayers, vec3(vTileUv, vLayers.x)).rgb;
				ground = mix(ground, texture(uLayers, vec3(vTileUv, vLayers.y)).rgb, groundAlpha.r);
				ground = mix(ground, texture(uLayers, vec3(vTileUv, vLayers.z)).rgb, groundAlpha.g);
				ground = mix(ground, texture(uLayers, vec3(vTileUv, vLayers.w)).rgb, groundAlpha.b);
				diffuseColor.rgb *= ground;`,
			);
	};
	return mat;
}

// cellIndex gives the triangles of a chunk: each of its 8 by 8 cells is four triangles around the
// height in its middle.
let cells: number[] | undefined;
function cellIndex(): number[] {
	if (cells) return cells;
	cells = [];
	for (let r = 0; r < 8; r++) {
		for (let c = 0; c < 8; c++) {
			const mid = r * 17 + 9 + c;
			const tl = r * 17 + c;
			const tr = tl + 1;
			const bl = tl + 17;
			const br = bl + 1;
			// Counterclockwise seen from above, in Three's space.
			cells.push(mid, tr, tl, mid, br, tr, mid, bl, br, mid, tl, bl);
		}
	}
	return cells;
}

// smoothSeams averages the normals where chunks meet (each chunk has its own copy of its edge
// heights), so the light has no seams along their edges.
function smoothSeams(g: THREE.BufferGeometry) {
	const p = g.getAttribute('position') as THREE.BufferAttribute;
	const n = g.getAttribute('normal') as THREE.BufferAttribute;
	const key = (i: number) => `${Math.round(p.getX(i) * 10)},${Math.round(p.getZ(i) * 10)}`;
	const sums = new Map<string, THREE.Vector3>();
	for (let i = 0; i < p.count; i++) {
		const k = key(i);
		const s = sums.get(k) ?? new THREE.Vector3();
		s.x += n.getX(i);
		s.y += n.getY(i);
		s.z += n.getZ(i);
		sums.set(k, s);
	}
	for (let i = 0; i < p.count; i++) {
		const s = sums.get(key(i))!.clone().normalize();
		n.setXYZ(i, s.x, s.y, s.z);
	}
	n.needsUpdate = true;
}

function smoothstep(a: number, b: number, x: number): number {
	const t = Math.min(1, Math.max(0, (x - a) / (b - a)));
	return t * t * (3 - 2 * t);
}

function range(a: number, b: number): number[] {
	const out: number[] = [];
	for (let i = Math.min(a, b); i <= Math.max(a, b); i++) out.push(i);
	return out;
}
