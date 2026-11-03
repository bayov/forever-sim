// A bare model viewer for checking the model readers against wow.export: /viewer.html?model=917116
// with optional &anim=5, &geosets=0,101,201 and &tex<type>=<fdid>.

import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { texture } from './wow/files';
import { Model } from './wow/model';
import { fillBatches } from './wow/particles';
import { Appearance, UnitView } from './wow/unit';

const q = new URLSearchParams(location.search);
const canvas = document.getElementById('c') as HTMLCanvasElement;
const panel = document.getElementById('panel')!;
const renderer = new THREE.WebGLRenderer({ canvas, antialias: true });
renderer.setPixelRatio(devicePixelRatio);
const scene = new THREE.Scene();
scene.background = new THREE.Color(0x3a4450);
scene.add(new THREE.HemisphereLight(0xffffff, 0x445566, 1.5));
const sun = new THREE.DirectionalLight(0xffffff, 1.5);
sun.position.set(3, 5, 4);
scene.add(sun);
scene.add(new THREE.GridHelper(10, 10, 0x666666, 0x444444));
const camera = new THREE.PerspectiveCamera(45, 1, 0.05, 100);
camera.position.set(...((q.get('cam') ?? '3,2,3').split(',').map(Number) as [number, number, number]));
const controls = new OrbitControls(camera, canvas);
Model.camera = camera;
controls.target.set(...((q.get('look') ?? '0,1,0').split(',').map(Number) as [number, number, number]));

function resize() {
	renderer.setSize(innerWidth, innerHeight, false);
	camera.aspect = innerWidth / innerHeight;
	camera.updateProjectionMatrix();
}
addEventListener('resize', resize);
resize();

const status = (s: string) => (panel.textContent = s);
let model: Model | undefined;
(window as unknown as { viewer: unknown }).viewer = { get model() { return model; } };

// ?char=race=2&sex=0&class=7&items=16:17104 draws a dressed character through the room's own code.
async function dressed(query: string) {
	const a: Appearance = await (await fetch(`/api/appearance/character?${query}`)).json();
	const group = new THREE.Group();
	scene.add(group);
	const view = new UnitView(group, []);
	await view.load(a);
	view.setDrawn(q.get('drawn') === '1');
	model = view.model;
	(window as unknown as { view: unknown }).view = view;
	let last = performance.now();
	setInterval(() => {
		const now = performance.now();
		view.update((now - last) / 1000, { forward: 0, left: 0, airborne: false, walking: false }, q.get('drawn') === '1');
		last = now;
	}, 16);
	panel.textContent = JSON.stringify(a.attachments?.map(x => [x.slot, x.attachment, x.model, x.replaceable]));
	document.body.dataset.ready = '1';
}

async function main() {
	const char = q.get('char');
	if (char) return dressed(char);
	const fdid = Number(q.get('model') ?? 917116);
	status(`loading ${fdid}...`);
	const t0 = performance.now();
	model = await Model.load(fdid);
	if (q.get('rest')) model.restPose();
	scene.add(model.root);
	model.enableParticles(scene);
	const geosets = q.get('geosets');
	if (geosets) model.setGeosets(geosets.split(',').map(Number));
	for (const [k, v] of q) {
		if (k.startsWith('tex')) model.setTexture(Number(k.slice(3)), await texture(Number(v)));
	}
	const anim = Number(q.get('anim') ?? 0);
	model.play(anim, { loop: true });
	const m2 = model.m2;
	panel.innerHTML = '';
	const info = document.createElement('div');
	info.textContent = `${m2.name}: ${m2.particles.length} emitters, ${m2.ribbons.length} ribbons, ${m2.bones.length} bones, ${m2.sequences.length} sequences, ${model.skin.submeshes.length} submeshes, loaded in ${Math.round(performance.now() - t0)}ms`;
	panel.append(info);
	const ids = document.createElement('div');
	ids.textContent = `geosets: ${model.submeshIds().join(' ')}`;
	ids.style.maxWidth = '600px';
	panel.append(ids);
	const animInput = document.createElement('input');
	animInput.placeholder = 'animation id';
	animInput.value = String(anim);
	animInput.onchange = () => model!.play(Number(animInput.value), { loop: true });
	panel.append(animInput);
	const geoInput = document.createElement('input');
	geoInput.placeholder = 'geosets';
	geoInput.value = geosets ?? '';
	geoInput.onchange = () => model!.setGeosets(geoInput.value ? geoInput.value.split(/[ ,]+/).map(Number) : undefined);
	panel.append(geoInput);
	document.body.dataset.ready = '1';
}

let last = performance.now();
renderer.setAnimationLoop(now => {
	const dt = Math.min(0.1, (now - last) / 1000);
	last = now;
	if (!q.get('rest') && !q.get('char')) model?.update(dt);
	controls.update();
	fillBatches();
	renderer.render(scene, camera);
});

main().catch(e => {
	status(String(e?.stack ?? e));
	document.body.dataset.ready = 'error';
});
