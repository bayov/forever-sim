// The room's ground: a piece of a real map from the client, with its ground textures and the
// doodads (trees, rocks, bushes) that stand on it.
//
// The fight is on flat ground, so we flatten the map under the play area and blend into the real
// hills further out. A spot that is already flat is best (Terrain.load's caller picks one).

import * as THREE from 'three';
import { CHUNK, parseDoodads, parseTerrain, parseWdt, TerrainChunk, tileAt, tileIndex, vertexAt } from './adt';
import { file, texture } from './files';
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

	private async buildGround(chunks: TerrainChunk[]) {
		const spot = this.spot;
		const textures = new Map<number, Promise<THREE.Texture | undefined>>();
		const tex = (fdid: number) => {
			let p = textures.get(fdid);
			if (!p) {
				p = fdid > 0 ? texture(fdid).catch(() => undefined) : Promise.resolve(undefined);
				textures.set(fdid, p);
			}
			return p;
		};
		const blank = new THREE.DataTexture(new Uint8Array([0, 0, 0, 255]), 1, 1);
		blank.needsUpdate = true;

		const index = cellIndex();
		const geometries: THREE.BufferGeometry[] = [];
		const meshes = await Promise.all(
			chunks.map(async c => {
				const pos = new Float32Array(145 * 3);
				const alphaUv = new Float32Array(145 * 2);
				const color = new Float32Array(145 * 3);
				for (let i = 0; i < 145; i++) {
					const [x, y, z] = vertexAt(c, i);
					const h = this.height(x, y, z);
					// World (x north, y west, z up) to Three (x, z, -y), from the room's middle.
					pos[i * 3] = x - spot.x;
					pos[i * 3 + 1] = h;
					pos[i * 3 + 2] = -(y - spot.y);
					const row = Math.floor(i / 17);
					const col = i % 17;
					const inner = col >= 9;
					alphaUv[i * 2] = (inner ? col - 9 + 0.5 : col) / 8;
					alphaUv[i * 2 + 1] = (inner ? row + 0.5 : row) / 8;
					// Vertex colors are BGRA, with 127 as 1.
					for (let k = 0; k < 3; k++) color[i * 3 + k] = c.colors ? c.colors[i * 4 + 2 - k] / 127 : 1;
				}
				const g = new THREE.BufferGeometry();
				g.setAttribute('position', new THREE.BufferAttribute(pos, 3));
				g.setAttribute('uv', new THREE.BufferAttribute(alphaUv, 2));
				g.setAttribute('color', new THREE.BufferAttribute(color, 3));
				g.setIndex(index);
				g.computeVertexNormals();
				geometries.push(g);

				const alpha = new THREE.DataTexture(c.alpha, 64, 64, THREE.RGBAFormat);
				alpha.magFilter = THREE.LinearFilter;
				alpha.minFilter = THREE.LinearFilter;
				alpha.needsUpdate = true;
				const layers = await Promise.all([0, 1, 2, 3].map(l => (c.layers[l] ? tex(c.layers[l]) : Promise.resolve(undefined))));
				const mat = groundMaterial(
					layers.map(t => t ?? blank),
					alpha,
				);
				const mesh = new THREE.Mesh(g, mat);
				mesh.receiveShadow = true;
				mesh.name = 'ground';
				return mesh;
			}),
		);
		smoothSeams(geometries);
		for (const m of meshes) this.group.add(m);
	}

	// loadDoodads loads the doodads' models and stands them on the ground. They don't animate.
	async loadDoodads() {
		const spot = this.spot;
		const byFile = new Map<number, Promise<boolean>>();
		await Promise.all(
			this.pending.map(async d => {
				// A model that failed once (a missing file) isn't tried for every copy.
				if ((await byFile.get(d.fdid)) === false) return;
				const load = Model.load(d.fdid);
				byFile.set(
					d.fdid,
					load.then(
						() => true,
						() => false,
					),
				);
				const m = await load.catch(() => undefined);
				if (!m) return;
				m.root.position.set(d.x - spot.x, this.height(d.x, d.y, d.z), -(d.y - spot.y));
				// Placement rotations are in degrees, around the placement axes (the client turns a
				// model by its y, then z, then x). Placement space is Three's turned 90 degrees.
				m.root.rotation.set(THREE.MathUtils.degToRad(d.rot[2]), THREE.MathUtils.degToRad(d.rot[1] - 180), THREE.MathUtils.degToRad(-d.rot[0]), 'YZX');
				m.root.scale.setScalar(d.scale);
				m.update(0);
				m.root.updateMatrixWorld(true);
				m.mesh.computeBoundingSphere();
				m.mesh.frustumCulled = true;
				this.group.add(m.root);
				this.doodads.push(m);
			}),
		);
		this.pending = [];
	}
}

// The ground shader: Lambert lighting and shadows, with the chunk's four textures blended by its
// alpha map. A texture repeats 8 times across a chunk.
function groundMaterial(layers: THREE.Texture[], alpha: THREE.Texture): THREE.MeshLambertMaterial {
	const mat = new THREE.MeshLambertMaterial({ vertexColors: true });
	mat.onBeforeCompile = shader => {
		shader.uniforms.uLayer0 = { value: layers[0] };
		shader.uniforms.uLayer1 = { value: layers[1] };
		shader.uniforms.uLayer2 = { value: layers[2] };
		shader.uniforms.uLayer3 = { value: layers[3] };
		shader.uniforms.uAlpha = { value: alpha };
		shader.vertexShader = shader.vertexShader
			.replace('#include <common>', '#include <common>\nvarying vec2 vGroundUv;')
			.replace('#include <begin_vertex>', '#include <begin_vertex>\nvGroundUv = uv;');
		shader.fragmentShader = shader.fragmentShader
			.replace('#include <common>', '#include <common>\nuniform sampler2D uLayer0, uLayer1, uLayer2, uLayer3, uAlpha;\nvarying vec2 vGroundUv;')
			.replace(
				'#include <map_fragment>',
				`vec4 groundAlpha = texture2D(uAlpha, vGroundUv);
				vec2 tileUv = vGroundUv * 8.0;
				vec3 ground = texture2D(uLayer0, tileUv).rgb;
				ground = mix(ground, texture2D(uLayer1, tileUv).rgb, groundAlpha.r);
				ground = mix(ground, texture2D(uLayer2, tileUv).rgb, groundAlpha.g);
				ground = mix(ground, texture2D(uLayer3, tileUv).rgb, groundAlpha.b);
				diffuseColor.rgb *= ground;`,
			);
	};
	// Every chunk's material compiles to the same program.
	mat.customProgramCacheKey = () => 'wow-ground';
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

// smoothSeams averages the normals where chunks meet, so the light has no seams along their edges.
function smoothSeams(geometries: THREE.BufferGeometry[]) {
	const sums = new Map<string, THREE.Vector3>();
	const key = (p: THREE.BufferAttribute, i: number) => `${Math.round(p.getX(i) * 10)},${Math.round(p.getZ(i) * 10)}`;
	for (const g of geometries) {
		const p = g.getAttribute('position') as THREE.BufferAttribute;
		const n = g.getAttribute('normal') as THREE.BufferAttribute;
		for (let i = 0; i < p.count; i++) {
			const k = key(p, i);
			const s = sums.get(k) ?? new THREE.Vector3();
			s.x += n.getX(i);
			s.y += n.getY(i);
			s.z += n.getZ(i);
			sums.set(k, s);
		}
	}
	for (const g of geometries) {
		const p = g.getAttribute('position') as THREE.BufferAttribute;
		const n = g.getAttribute('normal') as THREE.BufferAttribute;
		for (let i = 0; i < p.count; i++) {
			const s = sums.get(key(p, i))!.clone().normalize();
			n.setXYZ(i, s.x, s.y, s.z);
		}
		n.needsUpdate = true;
	}
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
