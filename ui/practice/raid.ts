// A raid's worth of other players and a boss, to see how the room holds up with 40 people.
//
// Open the room with ?raid=40 (or another count). Raiders of the four Horde races stand where a
// raid would: two tanks in front of the boss, the melee around its back, and casters, hunters and
// healers in rows in front of it. Each wears its class's raid gear: a dungeon, tier 1 or tier 2 set,
// with weapons for its role and a cloak, so two mages rarely look the same. The melee swing, the casters cast with
// their hand glows and bolts, the healers heal the tanks, the boss hits the main tank, and the
// ranged move now and then. Nothing here talks to the sim: it's only for drawing.

import * as THREE from 'three';
import { toThree } from './scene';
import type { Crowd } from './wow/crowd';
import type { Effects } from './wow/effects';
import { drawOrder } from './wow/particles';
import { Anim } from './wow/model';
import { Appearance, UnitView } from './wow/unit';

// The casters' spells: Fireball, Frostbolt, Shadow Bolt, Lightning Bolt, and Greater Heal.
const spells = { mage: [25306, 25304], warlock: [25307], shaman: [15208], priest: [25314] };
export const raidSpells = [25306, 25304, 25307, 15208, 25314];

type Role = 'tank' | 'melee' | 'caster' | 'healer' | 'hunter';

type Raider = {
	view: UnitView;
	role: Role;
	spells: number[];
	home: [number, number];
	pos: [number, number];
	// Where it's running to, when it moves.
	to?: [number, number];
	next: number;
	moveAt: number;
	// Time the animation hasn't caught up on yet (see animEvery).
	behind: number;
};

// animEvery is how often a raider's animation updates. Like the game, we animate the ones off
// screen and far away less often: off screen 10 times a second, past 40 yards 30, past 20 yards 60,
// and near you every frame.
function animEvery(onScreen: boolean, distance: number): number {
	if (!onScreen) return 0.1;
	if (distance > 40) return 1 / 30;
	if (distance > 20) return 1 / 60;
	return 0;
}

const frustum = new THREE.Frustum();
const viewProjection = new THREE.Matrix4();
const sphere = new THREE.Sphere(new THREE.Vector3(), 2.5);

// The races (2 orc, 5 undead, 6 tauren, 8 troll) and the classes each can be, by role.
const looks: Record<Role, [number, number][]> = {
	tank: [[2, 1], [6, 1]],
	melee: [[2, 4], [5, 4], [8, 4], [2, 1], [6, 1], [8, 1], [2, 7], [6, 7], [8, 7], [5, 1]],
	caster: [[5, 8], [8, 8], [2, 9], [5, 9], [2, 7], [6, 7], [8, 7]],
	healer: [[5, 5], [8, 5], [6, 11], [6, 7]],
	hunter: [[2, 3], [6, 3], [8, 3]],
};
const classSpells: Record<number, number[]> = { 8: spells.mage, 9: spells.warlock, 7: spells.shaman, 5: spells.priest, 11: spells.priest };

// The armor sets by class (the dungeon set, tier 1 and tier 2), as item ids for head, shoulders,
// chest, waist, legs, feet, wrists and hands.
const armorSlots = [1, 3, 5, 6, 7, 8, 9, 10];
const sets: Record<number, number[][]> = {
	// Warrior: Valor, Might, Wrath.
	1: [
		[16731, 16733, 16730, 16736, 16732, 16734, 16735, 16737],
		[16866, 16868, 16865, 16864, 16867, 16862, 16861, 16863],
		[16963, 16961, 16966, 16960, 16962, 16965, 16959, 16964],
	],
	// Hunter: Beaststalker, Giantstalker, Dragonstalker.
	3: [
		[16677, 16679, 16674, 16680, 16678, 16675, 16681, 16676],
		[16846, 16848, 16845, 16851, 16847, 16849, 16850, 16852],
		[16939, 16937, 16942, 16936, 16938, 16941, 16935, 16940],
	],
	// Rogue: Shadowcraft, Nightslayer, Bloodfang.
	4: [
		[16707, 16708, 16721, 16713, 16709, 16711, 16710, 16712],
		[16821, 16823, 16820, 16827, 16822, 16824, 16825, 16826],
		[16908, 16832, 16905, 16910, 16909, 16906, 16911, 16907],
	],
	// Priest: Devout, Prophecy, Transcendence.
	5: [
		[16693, 16695, 16690, 16696, 16694, 16691, 16697, 16692],
		[16813, 16816, 16815, 16817, 16814, 16811, 16819, 16812],
		[16921, 16924, 16923, 16925, 16922, 16919, 16926, 16920],
	],
	// Shaman: The Elements, Earthfury, Ten Storms.
	7: [
		[16667, 16669, 16666, 16673, 16668, 16670, 16671, 16672],
		[16842, 16844, 16841, 16838, 16843, 16837, 16840, 16839],
		[16947, 16945, 16950, 16944, 16946, 16949, 16943, 16948],
	],
	// Mage: Magister, Arcanist, Netherwind.
	8: [
		[16686, 16689, 16688, 16685, 16687, 16682, 16683, 16684],
		[16795, 16797, 16798, 16802, 16796, 16800, 16799, 16801],
		[16914, 16917, 16916, 16818, 16915, 16912, 16918, 16913],
	],
	// Warlock: Dreadmist, Felheart, Nemesis.
	9: [
		[16698, 16701, 16700, 16702, 16699, 16704, 16703, 16705],
		[16808, 16807, 16809, 16806, 16810, 16803, 16804, 16805],
		[16929, 16932, 16931, 16933, 16930, 16927, 16934, 16928],
	],
	// Druid: Wildheart, Cenarion, Stormrage.
	11: [
		[16720, 16718, 16706, 16716, 16719, 16715, 16714, 16717],
		[16834, 16836, 16833, 16828, 16835, 16829, 16830, 16831],
		[16900, 16902, 16897, 16903, 16901, 16898, 16904, 16899],
	],
};

// The weapons by class and role, as [main hand, off hand, ranged] (0 for none). Three choices
// each, from Molten Core and Blackwing Lair.
const weapons: Record<string, number[][]> = {
	// Brutality Blade, Chromatically Tempered Sword or Maladath, with Drillborer Disk or the
	// Elementium Reinforced Bulwark.
	'1:tank': [
		[18832, 17066, 0],
		[19352, 19349, 0],
		[19351, 17066, 0],
	],
	// Ashkandi, Bonereaver's Edge, or Brutality Blade and Deathbringer.
	'1:melee': [
		[19364, 0, 0],
		[17076, 0, 0],
		[18832, 17068, 0],
	],
	// Core Hound Tooth and Perdition's Blade, Dragonfang Blade and Gutgore Ripper, or two swords.
	'4:melee': [
		[18805, 18816, 0],
		[19346, 17071, 0],
		[19352, 19351, 0],
	],
	// A polearm or a two-handed sword, with Striker's Mark, Rhok'delar or a gun.
	'3:hunter': [
		[19874, 0, 17069],
		[12583, 0, 18713],
		[18520, 0, 19368],
	],
	// Spinal Reaper, Herald of Woe or Draconic Avenger.
	'7:melee': [
		[17104, 0, 0],
		[19357, 0, 0],
		[19354, 0, 0],
	],
	// Lok'amir, Aurastone Hammer or Spineshatter, with a shield.
	'7:caster': [
		[19360, 17106, 0],
		[17105, 19349, 0],
		[19335, 17106, 0],
	],
	// Staff of Dominance, Azuresong Mageblade and the Fire Runed Grimoire, or the Staff of the
	// Shadow Flame.
	'8:caster': [
		[18842, 0, 0],
		[17103, 19142, 0],
		[19356, 0, 0],
	],
	// Shadow Wing Focus Staff, Sorcerous Dagger and the Tome of the Ice Lord, or the Staff of the
	// Shadow Flame.
	'9:caster': [
		[19355, 0, 0],
		[18878, 19310, 0],
		[19356, 0, 0],
	],
	// Benediction, Lok'amir and the Master Dragonslayer's Orb, or Aurastone Hammer and the Fire
	// Runed Grimoire.
	'5:healer': [
		[18608, 0, 0],
		[19360, 19366, 0],
		[17105, 19142, 0],
	],
	// Staff of Dominance, Hammer of Ji'zhi, or Lok'amir and the Master Dragonslayer's Orb.
	'11:healer': [
		[18842, 0, 0],
		[21703, 0, 0],
		[19360, 19366, 0],
	],
};
weapons['7:healer'] = weapons['7:caster'];

// Cloaks from Molten Core, Onyxia, Blackwing Lair and Stratholme. (The Shroud of Pure Thought's
// texture isn't in the client.)
const cloaks = [17107, 19436, 17102, 19398, 17078, 18541, 19378, 13340];

// gear gives a raider's items for the appearance request: its class's set, its role's weapons and
// a cloak, picked by `pick` so raiders of one class wear different things.
function gear(cls: number, role: Role, pick: number): string {
	const armor = sets[cls]?.[pick % 3] ?? [];
	const kind = role === 'tank' ? 'tank' : role;
	const arms = weapons[`${cls}:${kind}`]?.[pick % 3] ?? [0, 0, 0];
	const out = armor.map((id, i) => `${armorSlots[i]}:${id}`);
	out.push(`15:${cloaks[(pick * 3 + cls) % cloaks.length]}`);
	arms.forEach((id, i) => id && out.push(`${16 + i}:${id}`));
	return out.join(',');
}

const boss: [number, number] = [0, 20];
// The boss faces the raid, down the room toward where you start.
const bossFacing = (3 * Math.PI) / 2;
const runSpeed = 7;

function facingTo(from: [number, number], to: [number, number]) {
	return Math.atan2(to[1] - from[1], to[0] - from[0]);
}

export class RaidTest {
	private raiders: Raider[] = [];
	private boss?: UnitView;
	private bossNext = 2;
	private clock = 0;

	constructor(
		readonly scene: THREE.Scene,
		readonly camera: THREE.Camera,
		readonly effects: Effects,
		// Where the raiders draw together. Without it they draw one by one, like you.
		readonly crowd?: Crowd,
	) {}

	// load builds the raid: count people in all, so count - 1 others besides you.
	async load(count: number, getJSON: <T>(url: string) => Promise<T>) {
		// With 20 or more, helms, shoulders and cloaks cast no shadow (their shadows mostly fall in
		// the body's), which saves a shadow draw call each.
		if (count >= 20) for (const slot of [1, 3, 15]) this.crowd?.noShadowSlots.add(slot);
		const roles: Role[] = [];
		const others = Math.max(0, count - 1);
		const plan: [Role, number][] = [
			['tank', 2],
			['melee', 10],
			['caster', 15],
			['healer', 7],
			['hunter', 5],
		];
		for (const [role, n] of plan) for (let i = 0; i < n; i++) roles.push(role);
		while (roles.length < others) roles.push('caster');
		roles.length = others;

		const ranged = roles.filter(r => r !== 'tank' && r !== 'melee').length;
		let melee = 0;
		let tanks = 0;
		let row = 0;
		const homes = roles.map(role => {
			if (role === 'tank') return [boss[0] + (tanks++ === 0 ? -1.5 : 1.5), boss[1] - 7] as [number, number];
			if (role === 'melee') {
				// Around the boss's back half.
				const a = ((10 + (160 * melee++) / 9) * Math.PI) / 180;
				return [boss[0] + 6.5 * Math.cos(a), boss[1] + 6.5 * Math.sin(a)] as [number, number];
			}
			// Three rows in front of the boss.
			const perRow = Math.ceil(ranged / 3);
			const i = row++;
			const r = 22 + 5 * Math.floor(i / perRow);
			const k = i % perRow;
			const a = ((200 + (140 * (k + 0.5)) / perRow) * Math.PI) / 180;
			return [boss[0] + r * Math.cos(a), boss[1] + r * Math.sin(a)] as [number, number];
		});

		const bossLook = getJSON<Appearance>('/api/appearance/creature/11121').then(async a => {
			const view = new UnitView(new THREE.Group(), [], this.scene);
			await view.load(a);
			this.scene.add(view.group);
			view.group.position.copy(toThree(boss[0], boss[1]));
			view.group.rotation.y = bossFacing;
			for (const m of view.models) m.mesh.renderOrder = drawOrder.boss;
			this.boss = view;
		});

		// Eight at a time, so the practice server isn't asked for 39 character textures at once.
		// k is the raider's place among those of its role.
		const seen = new Map<Role, number>();
		const queue = roles.map((role, i) => {
			const k = seen.get(role) ?? 0;
			seen.set(role, k + 1);
			return { role, i, k };
		});
		const work = async () => {
			for (let job = queue.shift(); job; job = queue.shift()) {
				const options = looks[job.role];
				const [race, cls] = options[job.k % options.length];
				const sex = job.i % 3 === 0 ? 1 : 0;
				// The raiders of one look (race and class) go through the sets in turn. The tanks wear
				// tier 1 and tier 2.
				const pick = job.role === 'tank' ? 1 + job.k : Math.floor(job.k / options.length) + job.k;
				const items = gear(cls, job.role, pick);
				const a = await getJSON<Appearance>(`/api/appearance/character?race=${race}&sex=${sex}&class=${cls}&items=${items}`).catch(() => undefined);
				if (!a) continue;
				const view = new UnitView(new THREE.Group(), [], this.scene);
				await view.load(a).catch(() => undefined);
				if (!view.model) continue;
				this.scene.add(view.group);
				this.crowd?.add(view.models);
				const home = homes[job.i];
				const role = job.role === 'caster' && !classSpells[cls] ? 'hunter' : job.role;
				this.raiders.push({ view, role, spells: classSpells[cls] ?? [], home, pos: [...home], next: Math.random() * 3, moveAt: 5 + Math.random() * 15, behind: 0 });
			}
		};
		await Promise.all([bossLook.catch(e => console.warn('raid boss', e)), ...Array.from({ length: 8 }, work)]);
	}

	get count() {
		return this.raiders.length + 1;
	}

	update(dt: number) {
		this.clock += dt;
		const tank = this.raiders.find(r => r.role === 'tank');
		if (this.boss) {
			if (this.clock >= this.bossNext && tank) {
				this.bossNext = this.clock + 2;
				this.boss.swing(false, false);
				tank.view.wound(Math.random() < 0.2);
			}
			this.boss.update(dt, { forward: 0, left: 0, airborne: false, walking: false }, true);
		}
		const cam = this.camera;
		cam.updateMatrixWorld();
		frustum.setFromProjectionMatrix(viewProjection.multiplyMatrices(cam.projectionMatrix, cam.matrixWorldInverse));
		for (const r of this.raiders) {
			let forward = 0;
			if (r.to) {
				const dx = r.to[0] - r.pos[0];
				const dy = r.to[1] - r.pos[1];
				const d = Math.hypot(dx, dy);
				const step = runSpeed * dt;
				if (d <= step) {
					r.pos = [...r.to];
					r.to = undefined;
				} else {
					r.pos = [r.pos[0] + (dx / d) * step, r.pos[1] + (dy / d) * step];
					forward = runSpeed;
				}
			} else if (r.role !== 'tank' && r.role !== 'melee' && this.clock >= r.moveAt) {
				// The ranged move a few yards now and then, like stepping out of something.
				r.moveAt = this.clock + 8 + Math.random() * 12;
				const a = Math.random() * 2 * Math.PI;
				r.to = [r.home[0] + 4 * Math.cos(a), r.home[1] + 4 * Math.sin(a)];
			}
			const facing = r.to ? facingTo(r.pos, r.to) : r.role === 'healer' && tank ? facingTo(r.pos, tank.pos) : facingTo(r.pos, boss);
			r.view.group.position.copy(toThree(r.pos[0], r.pos[1]));
			r.view.group.rotation.y = facing;
			if (!r.to && this.clock >= r.next) this.act(r, tank);
			sphere.center.copy(r.view.group.position).y += 1;
			r.behind += dt;
			if (r.behind >= animEvery(frustum.intersectsSphere(sphere), sphere.center.distanceTo(cam.position))) {
				r.view.update(r.behind, { forward, left: 0, airborne: false, walking: false }, true);
				r.behind = 0;
			}
		}
	}

	private act(r: Raider, tank: Raider | undefined) {
		if (r.role === 'tank' || r.role === 'melee') {
			r.next = this.clock + 1.2 + Math.random() * 1.4;
			r.view.swing(Math.random() < 0.3, false);
			return;
		}
		if (r.role === 'hunter') {
			r.next = this.clock + 2 + Math.random();
			r.view.castEnd(Anim.SpellCastDirected);
			return;
		}
		const spellId = r.spells[Math.floor(Math.random() * r.spells.length)];
		const castMs = 1800 + Math.random() * 1200;
		r.next = this.clock + castMs / 1000 + 0.2 + Math.random() * 0.5;
		r.view.castStart();
		const target = r.role === 'healer' ? tank?.view : this.boss;
		this.effects.otherCast(spellId, r.view, target, castMs).then(() => {
			if (!r.to) r.view.castEnd();
			else r.view.castStop();
		});
	}
}
