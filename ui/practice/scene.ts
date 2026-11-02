// The 3D room: a flat arena, the dummy and you. Units are capsules with their race's height and
// combat reach until the model renderer draws real models (docs/practice_room.md, "Rendering").
//
// The sim's world has x east and y north, in yards. Three.js has y up, so a sim point (x, y) is
// (x, 0, -y) here, and a facing angle turns the same way.

import * as THREE from 'three';
import { newTextures } from './wow/files';
import type { Model } from './wow/model';

export type CameraState = { yaw: number; pitch: number; distance: number };

export const toThree = (x: number, y: number, z = 0) => new THREE.Vector3(x, z, -y);

function capsule(radius: number, height: number, color: number): THREE.Group {
	const group = new THREE.Group();
	const body = new THREE.Mesh(
		new THREE.CapsuleGeometry(radius, Math.max(0.01, height - 2 * radius), 6, 16),
		new THREE.MeshStandardMaterial({ color, roughness: 0.7 }),
	);
	body.position.y = height / 2;
	body.castShadow = true;
	group.add(body);
	// A nose that shows which way the unit faces.
	const nose = new THREE.Mesh(new THREE.ConeGeometry(radius * 0.35, radius * 0.9, 12), new THREE.MeshStandardMaterial({ color: 0xffe08a }));
	nose.rotation.z = -Math.PI / 2;
	nose.position.set(radius * 1.05, height * 0.78, 0);
	group.add(nose);
	return group;
}

const totemColors: Record<string, number> = { earth: 0x8b5a2b, fire: 0xd8442a, water: 0x2a7bd8, air: 0xc8d8f0 };

// The graphics settings, with presets like the game's.
export type Graphics = { preset: 'low' | 'medium' | 'high' | 'custom'; renderScale: number; shadows: 'off' | 'low' | 'medium' | 'high' };

export const graphicsPresets: Record<'low' | 'medium' | 'high', Graphics> = {
	low: { preset: 'low', renderScale: 0.75, shadows: 'off' },
	medium: { preset: 'medium', renderScale: 1, shadows: 'low' },
	high: { preset: 'high', renderScale: 1, shadows: 'high' },
};

export class Room {
	readonly renderer: THREE.WebGLRenderer;
	private readonly sun: THREE.DirectionalLight;
	readonly scene = new THREE.Scene();
	readonly camera: THREE.PerspectiveCamera;
	readonly player: THREE.Group;
	readonly dummy: THREE.Group;
	readonly dummyHit: THREE.Mesh;
	private ring: THREE.Mesh;
	private meleeRing: THREE.Mesh;
	private totems = new Map<string, THREE.Group>();
	readonly totemModels = new Map<string, Model>();
	private dying: { model: Model; group: THREE.Group; until: number }[] = [];
	private clock = 0;
	// totemModel loads a totem's real model by its element, when the client's files are there.
	totemModel?: (element: string) => Promise<Model | undefined>;
	private readonly raycaster = new THREE.Raycaster();
	private readonly pickable: THREE.Object3D[] = [];
	private playerHeight: number;

	constructor(
		readonly canvas: HTMLCanvasElement,
		_playerReach: number,
		playerHeight: number,
		dummyReach: number,
		dummyHeight: number,
	) {
		this.playerHeight = playerHeight;
		this.renderer = new THREE.WebGLRenderer({ canvas, antialias: true, powerPreference: 'high-performance', desynchronized: true } as THREE.WebGLRendererParameters);
		this.renderer.setPixelRatio(window.devicePixelRatio);
		this.renderer.shadowMap.enabled = true;
		this.scene.background = new THREE.Color(0x87a7c4);
		this.scene.fog = new THREE.Fog(0x87a7c4, 60, 160);
		this.camera = new THREE.PerspectiveCamera(70, 1, 0.1, 500);

		// The client lights models with one ambient color plus the sun (about 0.45 and 0.75 outdoors).
		// Lambert divides by pi, so the intensities carry a pi.
		this.scene.add(new THREE.AmbientLight(0xffffff, 0.45 * Math.PI));
		const sun = new THREE.DirectionalLight(0xfff2e0, 0.75 * Math.PI);
		this.sun = sun;
		sun.castShadow = true;
		// The shadows cover 50 yards around you and follow you, so their pixels stay small. The
		// biases keep a surface from shadowing itself in stripes.
		sun.shadow.camera.left = sun.shadow.camera.bottom = -25;
		sun.shadow.camera.right = sun.shadow.camera.top = 25;
		sun.shadow.camera.near = 1;
		sun.shadow.camera.far = 160;
		sun.shadow.bias = -0.0005;
		sun.shadow.normalBias = 0.04;
		this.scene.add(sun, sun.target);

		const ground = new THREE.Mesh(new THREE.CircleGeometry(120, 64), new THREE.MeshStandardMaterial({ color: 0x5f7d3f, roughness: 1 }));
		ground.rotation.x = -Math.PI / 2;
		ground.receiveShadow = true;
		ground.name = 'ground';
		this.scene.add(ground);
		const grid = new THREE.GridHelper(120, 60, 0x3d5228, 0x4a6431);
		grid.position.y = 0.01;
		this.scene.add(grid);

		this.dummy = capsule(dummyReach * 0.5, dummyHeight, 0x9c7a4f);
		this.scene.add(this.dummy);
		// The click hull: a cylinder of the dummy's combat reach.
		this.dummyHit = new THREE.Mesh(new THREE.CylinderGeometry(dummyReach * 0.6, dummyReach * 0.6, dummyHeight, 12), new THREE.MeshBasicMaterial({ visible: false }));
		this.dummyHit.position.y = dummyHeight / 2;
		this.dummyHit.name = 'dummy';
		this.dummy.add(this.dummyHit);
		this.pickable.push(this.dummyHit, ground);

		this.ring = new THREE.Mesh(new THREE.RingGeometry(dummyReach * 0.6, dummyReach * 0.75, 40), new THREE.MeshBasicMaterial({ color: 0xff3030, transparent: true, opacity: 0.8 }));
		this.ring.rotation.x = -Math.PI / 2;
		this.ring.position.y = 0.03;
		this.dummy.add(this.ring);
		// The melee range edge around the dummy, faint.
		this.meleeRing = new THREE.Mesh(new THREE.RingGeometry(4.95, 5.05, 64), new THREE.MeshBasicMaterial({ color: 0xffffff, transparent: true, opacity: 0.25 }));
		this.meleeRing.rotation.x = -Math.PI / 2;
		this.meleeRing.position.y = 0.02;
		this.dummy.add(this.meleeRing);

		// A body about as wide as a character, not the combat reach, which is longer.
		this.player = capsule(0.32, playerHeight, 0x6b4f9c);
		this.scene.add(this.player);
	}

	// placeholders are a unit's capsule parts, which hide when its model loads. The dummy's click
	// hull and rings stay.
	placeholders(group: THREE.Group): THREE.Object3D[] {
		return group.children.slice(0, 2);
	}

	resize() {
		const w = this.canvas.clientWidth;
		const h = this.canvas.clientHeight;
		this.renderer.setSize(w, h, false);
		this.camera.aspect = w / Math.max(1, h);
		this.camera.updateProjectionMatrix();
	}

	setFov(horizontalDeg: number) {
		// WoW's cameraFov is horizontal (check). Three.js takes the vertical one.
		const h = (horizontalDeg * Math.PI) / 180;
		const v = 2 * Math.atan(Math.tan(h / 2) / Math.max(0.1, this.camera.aspect));
		this.camera.fov = (v * 180) / Math.PI;
		this.camera.updateProjectionMatrix();
	}

	setUnit(group: THREE.Group, pos: [number, number], facing: number, height = 0) {
		group.position.copy(toThree(pos[0], pos[1], height));
		group.rotation.y = facing;
	}

	setTargeted(on: boolean) {
		this.ring.visible = on;
	}

	setTotems(totems: { element: string; pos: [number, number] }[]) {
		const seen = new Set<string>();
		for (const t of totems) {
			seen.add(t.element);
			let g = this.totems.get(t.element);
			if (!g) {
				g = new THREE.Group();
				const pole = new THREE.Mesh(new THREE.CylinderGeometry(0.12, 0.18, 1.1, 8), new THREE.MeshStandardMaterial({ color: totemColors[t.element] ?? 0xffffff }));
				pole.position.y = 0.55;
				pole.castShadow = true;
				g.add(pole);
				this.scene.add(g);
				this.totems.set(t.element, g);
				const group = g;
				this.totemModel?.(t.element).then(m => {
					if (!m) return;
					if (this.totems.get(t.element) !== group) {
						m.dispose();
						return;
					}
					pole.visible = false;
					group.add(m.root);
					m.enableParticles(this.scene);
					// It rises out of the ground (Birth), then stands.
					if (m.has(127)) m.play(127, { loop: false, fade: 0, restart: true });
					this.totemModels.set(t.element, m);
				});
			}
			g.position.copy(toThree(t.pos[0], t.pos[1]));
		}
		for (const [element, g] of this.totems) {
			if (!seen.has(element)) {
				this.totems.delete(element);
				const m = this.totemModels.get(element);
				this.totemModels.delete(element);
				// It plays its death (it sinks away) before it goes.
				if (m && m.has(1)) {
					const seq = m.play(1, { loop: false, fade: 100, restart: true });
					m.emitting = false;
					this.dying.push({ model: m, group: g, until: this.clock + (m.m2.sequences[seq]?.duration ?? 0) / 1000 });
				} else {
					this.scene.remove(g);
					m?.dispose();
				}
			}
		}
	}

	// placeCamera puts the camera behind and above you, pulled in when the ground is in the way.
	placeCamera(pos: [number, number], height: number, cam: CameraState) {
		// The camera pivots a little above your head, so you show in the lower middle.
		const head = toThree(pos[0], pos[1], height + this.playerHeight + 0.35);
		// The way the camera looks: along its yaw, and down by its pitch.
		const look = new THREE.Vector3(Math.cos(cam.yaw) * Math.cos(cam.pitch), -Math.sin(cam.pitch), -Math.sin(cam.yaw) * Math.cos(cam.pitch));
		let distance = cam.distance;
		// A camera looking up sits low: it pulls in so it stays above the ground.
		if (look.y > 0) distance = Math.min(distance, (head.y - 0.2) / look.y);
		this.camera.position.copy(head).addScaledVector(look, -distance);
		this.camera.lookAt(head);
	}

	// warmUp gets the fight's shaders and textures ready while the loading screen is up.
	//
	// The first time the renderer draws a kind of material it compiles a shader for it, and the
	// first time it draws a texture it sends it to the GPU. Both happened mid fight and froze a
	// frame for up to 80ms, the first time a spell's effect played. So we put every effect and totem
	// model in the scene once (with their particles), compile what they need, and send every texture
	// we decoded. We keep the models (out of the scene), because disposing them would let the
	// renderer throw away the shaders we just made.
	async warmUp(models: Model[]) {
		const group = new THREE.Group();
		group.position.copy(this.camera.position);
		for (const m of models) {
			group.add(m.root);
			m.enableParticles(group);
			m.update(0);
		}
		this.scene.add(group);
		try {
			await this.renderer.compileAsync(this.scene, this.camera);
			for (const t of newTextures.list) this.renderer.initTexture(t);
		} finally {
			this.scene.remove(group);
			newTextures.keep = false;
			newTextures.list = [];
			this.warmModels.push(...models);
		}
	}
	private readonly warmModels: Model[] = [];

	// pick finds the unit under a screen point: 'enemy' for the dummy, 'ground' or undefined.
	pick(ndcX: number, ndcY: number): 'enemy' | 'ground' | undefined {
		this.raycaster.setFromCamera(new THREE.Vector2(ndcX, ndcY), this.camera);
		// Only the dummy's hull and the ground can be picked. Testing the whole scene would test
		// every triangle of the skinned character models, which took 25ms per mouse move.
		const hits = this.raycaster.intersectObjects(this.pickable, false);
		for (const h of hits) {
			if (h.object.name === 'dummy') return 'enemy';
			if (h.object.name === 'ground') return 'ground';
		}
		return undefined;
	}

	// screenPoint is where a world point shows on screen, in CSS pixels, for combat text.
	screenPoint(pos: [number, number], height: number): { x: number; y: number; visible: boolean } {
		const v = toThree(pos[0], pos[1], height).project(this.camera);
		return {
			x: ((v.x + 1) / 2) * this.canvas.clientWidth,
			y: ((1 - v.y) / 2) * this.canvas.clientHeight,
			visible: v.z < 1,
		};
	}

	// animate moves the totems' own animations and particles on.
	animate(dt: number) {
		this.clock += dt;
		for (const m of this.totemModels.values()) {
			if (m.baseAnim === 127 && m.baseDone) m.play(0, { loop: true, fade: 150 });
			m.update(dt);
		}
		this.dying = this.dying.filter(d => {
			if (this.clock < d.until) {
				d.model.update(dt);
				return true;
			}
			this.scene.remove(d.group);
			d.model.dispose();
			return false;
		});
	}

	render() {
		const p = this.player.position;
		this.sun.target.position.copy(p);
		this.sun.position.set(p.x + 30, p.y + 60, p.z + 20);
		this.renderer.render(this.scene, this.camera);
	}

	// setGraphics applies the graphics settings: the render scale and the shadows.
	setGraphics(g: Graphics) {
		this.renderer.setPixelRatio(window.devicePixelRatio * g.renderScale);
		this.resize();
		const shadows = g.shadows !== 'off';
		const size = g.shadows === 'high' ? 4096 : g.shadows === 'medium' ? 2048 : 1024;
		// Three's PCF shadows are soft already (PCFSoftShadowMap is deprecated and falls back to it).
		const type = THREE.PCFShadowMap;
		if (shadows !== this.renderer.shadowMap.enabled || type !== this.renderer.shadowMap.type) {
			this.renderer.shadowMap.enabled = shadows;
			this.renderer.shadowMap.type = type;
			// The materials compile again with or without shadows.
			this.scene.traverse(o => {
				const m = (o as THREE.Mesh).material;
				for (const x of Array.isArray(m) ? m : m ? [m] : []) x.needsUpdate = true;
			});
		}
		if (this.sun.shadow.mapSize.width !== size) {
			this.sun.shadow.mapSize.set(size, size);
			this.sun.shadow.map?.dispose();
			this.sun.shadow.map = null;
		}
	}
}
