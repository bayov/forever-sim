// A raid's worth of other players and a boss, to see how the room holds up with 40 people.
//
// Open the room with ?raid=40 (or another count). Raiders of the four Horde races stand where a
// raid would: two tanks in front of the boss, the melee around its back, and casters, hunters and
// healers in rows in front of it. They all wear your gear. The melee swing, the casters cast with
// their hand glows and bolts, the healers heal the tanks, the boss hits the main tank, and the
// ranged move now and then. Nothing here talks to the sim: it's only for drawing.

import * as THREE from 'three';
import { toThree } from './scene';
import type { Effects } from './wow/effects';
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
};

// The races (2 orc, 5 undead, 6 tauren, 8 troll) and the classes each can be, by role.
const looks: Record<Role, [number, number][]> = {
	tank: [[2, 1], [6, 1]],
	melee: [[2, 4], [5, 4], [8, 4], [2, 1], [6, 1], [8, 1], [2, 7], [6, 7], [8, 7], [5, 1]],
	caster: [[5, 8], [8, 8], [2, 9], [5, 9], [2, 7], [6, 7], [8, 7]],
	healer: [[5, 5], [8, 5], [6, 11], [6, 7]],
	hunter: [[2, 3], [6, 3], [8, 3]],
};
const classSpells: Record<number, number[]> = { 8: spells.mage, 9: spells.warlock, 7: spells.shaman, 5: spells.priest, 11: spells.priest };

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
		readonly effects: Effects,
	) {}

	// load builds the raid: count people in all, so count - 1 others besides you.
	async load(count: number, items: string, getJSON: <T>(url: string) => Promise<T>) {
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
			this.boss = view;
		});

		// Eight at a time, so the practice server isn't asked for 39 character textures at once.
		const queue = roles.map((role, i) => ({ role, i }));
		const work = async () => {
			for (let job = queue.shift(); job; job = queue.shift()) {
				const options = looks[job.role];
				const [race, cls] = options[job.i % options.length];
				const sex = job.i % 3 === 0 ? 1 : 0;
				const a = await getJSON<Appearance>(`/api/appearance/character?race=${race}&sex=${sex}&class=${cls}&items=${items}`).catch(() => undefined);
				if (!a) continue;
				const view = new UnitView(new THREE.Group(), [], this.scene);
				await view.load(a).catch(() => undefined);
				if (!view.model) continue;
				this.scene.add(view.group);
				const home = homes[job.i];
				const role = job.role === 'caster' && !classSpells[cls] ? 'hunter' : job.role;
				this.raiders.push({ view, role, spells: classSpells[cls] ?? [], home, pos: [...home], next: Math.random() * 3, moveAt: 5 + Math.random() * 15 });
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
			r.view.update(dt, { forward, left: 0, airborne: false, walking: false }, true);
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
