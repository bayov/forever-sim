// A unit's real model in the room: its body, gear and animations. Until the model loads (or when
// the client's files aren't there), the unit stays a capsule.
//
// The animations follow what the unit does, the way the client picks them: run, walk, backpedal
// and strafe by the way you move, jump start, loop and land, the ready stance while attacking,
// swings (on the upper body while you run), the precast loop during a cast and the cast at its end.

import * as THREE from 'three';
import { compose, Composite } from './compose';
import { prefetch, texture } from './files';
import { Anim, Model } from './model';

// sheathedAttachment is where a weapon hangs while sheathed, or -1 when it hides then.
export type AppearanceAttachment = {
	slot: number;
	attachment: number;
	model: number;
	textures?: number[];
	replaceable?: Record<string, number>;
	sheathe?: number;
	sheathedAttachment?: number;
};

export type Appearance = {
	model: number;
	creatureDisplayId?: number;
	scale?: number;
	geosets?: number[];
	textures?: Record<string, Composite>;
	replaceable?: Record<string, number>;
	attachments?: AppearanceAttachment[];
	collections?: { slot: number; model: number; textures?: number[]; replaceable?: Record<string, number>; geosets?: number[] }[];
};

function applyReplaceable(m: Model, replaceable: Record<string, number> | undefined, textures: number[] | undefined): Promise<unknown> {
	const byType = replaceable ?? (textures?.[0] ? { '2': textures[0] } : {});
	return Promise.all(
		Object.entries(byType).map(([type, fdid]) =>
			texture(fdid)
				.then(t => m.setTexture(Number(type), t))
				.catch(e => console.warn('texture', fdid, e)),
		),
	);
}

// Where a sheathed weapon hangs, by the item's sheathe type: 1 two-hand and 2 staff on the back,
// 3 one-hand at the hips, 4 a shield on the back (attachment ids 26 to 33).
function sheathPoint(sheathe: number | undefined, offHand: boolean): number | undefined {
	switch (sheathe) {
		case 1:
		case 2:
			return offHand ? 27 : 26;
		case 3:
			return offHand ? 33 : 32;
		case 4:
			return 28;
	}
	return undefined;
}

type Held = { model: Model; slot: number; sheathe?: number; hand?: THREE.Object3D; sheath?: THREE.Object3D; hiddenSheathed?: boolean };

export type Motion = {
	// Velocity relative to facing, in yards a second: forward and left.
	forward: number;
	left: number;
	// How fast the facing turns, in radians a second (left is positive).
	turn?: number;
	airborne: boolean;
	walking: boolean;
};

// angleTo gives the signed turn from one angle to another, the short way round.
function angleTo(from: number, to: number): number {
	let d = (to - from) % (2 * Math.PI);
	if (d > Math.PI) d -= 2 * Math.PI;
	if (d < -Math.PI) d += 2 * Math.PI;
	return d;
}

export class UnitView {
	model?: Model;
	private items: Model[] = [];
	private held: Held[] = [];
	private drawn = false;
	// Where the weapons are headed while the draw or sheathe animation plays.
	private drawing?: boolean;
	private wasAirborne = false;
	private oneShot = false;
	// A swing or a cast's release, which carries on in the upper body when you start to move.
	private oneShotUpper = false;
	private casting?: { precast: number; loop: boolean };
	private readyUntil = 0;
	private twoHanded = false;
	private twoHandedLoose = false;
	private hasOffHand = false;
	private mainHandEmpty = true;
	// The body turns toward where you move (strafing runs sideways), relative to the facing.
	private bodyYaw = 0;
	private twists = false;
	private timers: { at: number; fn: () => void }[] = [];
	// What happens when the swing in progress connects (its $HIT): the impact sound and the
	// target's flinch. A hit whose swing never gets there goes off at its deadline.
	private hits: { deadline: number; fn: () => void }[] = [];
	clock = 0;
	// onSound hears the animation's sounds: 'swing' at the whoosh, 'footstep' at each step.
	onSound?: (kind: 'swing' | 'footstep' | 'draw' | 'sheathe', at: THREE.Object3D) => void;

	constructor(
		readonly group: THREE.Group,
		readonly placeholder: THREE.Object3D[],
		// Where gear's particles live (the scene), or nowhere for a unit without them.
		readonly world?: THREE.Object3D,
	) {}

	// load loads the model with everything on it (textures, gear, weapons) and swaps out the
	// capsule only once it's all there.
	async load(a: Appearance) {
		// A character in gear twists its chest on a strafe. A creature (the Ghost Wolf) turns whole.
		this.twists = !!a.attachments?.length;
		await prefetch([a.model, ...(a.attachments ?? []).map(x => x.model), ...(a.collections ?? []).map(x => x.model)]);
		const model = await Model.load(a.model);
		const waits: Promise<unknown>[] = [];
		if (a.geosets) model.setGeosets(a.geosets);
		for (const [type, comp] of Object.entries(a.textures ?? {})) {
			waits.push(
				compose(comp)
					.then(t => model.setTexture(Number(type), t))
					.catch(e => console.warn('texture', type, e)),
			);
		}
		waits.push(applyReplaceable(model, a.replaceable, undefined));
		if (a.scale) model.root.scale.setScalar(a.scale);
		for (const c of a.collections ?? []) {
			// Capes, belts and such are drawn with the body's bones. We give them our skeleton.
			waits.push(
				Model.load(c.model)
					.then(async m => {
						if (c.geosets) m.setGeosets(c.geosets);
						await applyReplaceable(m, c.replaceable, c.textures);
						m.mesh.bind(model.mesh.skeleton, new THREE.Matrix4());
						model.root.add(m.mesh);
						this.items.push(m);
					})
					.catch(e => console.warn('collection', c.model, e)),
			);
		}
		for (const att of a.attachments ?? []) {
			waits.push(
				Model.load(att.model)
					.then(async m => {
						await applyReplaceable(m, att.replaceable, att.textures);
						// Gear can have its own glow (the Crown of Destruction's fire).
						if (this.world) m.enableParticles(this.world);
						this.items.push(m);
						const isWeapon = att.slot === 16 || att.slot === 17 || att.slot === 18;
						const hand = model.attachPoint(att.attachment);
						if (!isWeapon) {
							(hand ?? model.root).add(m.root);
							return;
						}
						const offHand = att.slot === 17;
						const point = att.sheathedAttachment ?? sheathPoint(att.sheathe, offHand);
						const sheath = point !== undefined && point >= 0 ? model.attachPoint(point) : undefined;
						const hidden = point === -1;
						this.held.push({ model: m, slot: att.slot, sheathe: att.sheathe, hand, sheath, hiddenSheathed: hidden });
						if (att.slot === 16) {
							this.mainHandEmpty = false;
							this.twoHanded = att.sheathe === 1 || att.sheathe === 2;
							// Staves and polearms (sheathe type 2) swing loose, with both hands low.
							this.twoHandedLoose = att.sheathe === 2;
						}
						if (offHand) this.hasOffHand = true;
					})
					.catch(e => console.warn('attachment', att.model, e)),
			);
		}
		// Every animation the unit can play, so none waits on its file mid fight.
		waits.push(model.preloadAnims(Object.values(Anim)));
		await Promise.all(waits);
		for (const m of [model, ...this.items]) {
			m.compact();
			m.cull();
		}
		this.model = model;
		model.onEvent = (id, data, upper) => this.event(id, data, upper);
		this.placeWeapons();
		this.group.add(model.root);
		for (const p of this.placeholder) p.visible = false;
	}

	// placeWeapons puts the weapons in the hands when drawn, in their sheaths when not.
	private placeWeapons() {
		for (const h of this.held) {
			const where = this.drawn ? h.hand : (h.sheath ?? h.hand);
			where?.add(h.model.root);
			h.model.root.visible = this.drawn || !h.hiddenSheathed;
		}
		const right = this.drawn && this.held.some(h => h.slot === 16);
		const left = this.drawn && this.held.some(h => h.slot === 17);
		this.model?.closeHands(right, left);
	}

	// setDrawn draws or sheathes the weapons. With `animate` it plays the draw (the weapon moves
	// when the hand reaches it), else they move at once.
	setDrawn(on: boolean, animate = false) {
		const m = this.model;
		if (on === this.drawn && this.drawing === undefined) return;
		if (this.drawing === on) return;
		if (!animate || !m || this.held.length === 0) {
			this.drawing = undefined;
			this.drawn = on;
			this.placeWeapons();
			return;
		}
		// Back weapons come off the back (Sheath), the others off the hips (HipSheath).
		const back = this.held.some(h => h.slot === 16 && h.sheathe !== 3);
		const anim = back ? Anim.Sheath : Anim.HipSheath;
		this.drawing = on;
		const seq = m.playUpper(anim);
		const at = m.eventTimes('$SHL', seq)[0] ?? m.eventTimes('$SHR', seq)[0] ?? 450;
		this.later(at / 1000, () => {
			if (this.drawing !== on) return;
			this.drawing = undefined;
			this.drawn = on;
			this.placeWeapons();
			if (this.model) this.onSound?.(on ? 'draw' : 'sheathe', this.model.root);
		});
	}

	private later(seconds: number, fn: () => void) {
		this.timers.push({ at: this.clock + seconds, fn });
	}

	// event handles the model's animation events.
	private event(id: string, _data: number, upper: boolean) {
		const m = this.model;
		if (!m) return;
		switch (id) {
			case '$HIT':
			case '$CAH':
				this.flushHits();
				break;
			case '$CSS':
				this.onSound?.('swing', m.root);
				break;
			case '$FSD':
				// The legs step. A swing on the upper body while running has its own step, which
				// the run's steps already cover.
				if (!upper) this.onSound?.('footstep', m.root);
				break;
		}
	}

	private flushHits() {
		const hits = this.hits;
		this.hits = [];
		for (const h of hits) h.fn();
	}

	// onHit runs fn when the swing in progress (or the one about to start) connects.
	onHit(fn: () => void) {
		this.hits.push({ deadline: this.clock + 0.8, fn });
	}

	private readyAnim() {
		if (this.mainHandEmpty) return Anim.ReadyUnarmed;
		if (this.twoHandedLoose) return Anim.Ready2HL;
		return this.twoHanded ? Anim.Ready2H : Anim.Ready1H;
	}

	// mainHand is the main hand weapon's model, for the imbue's glow.
	mainHand(): Model | undefined {
		return this.held.find(h => h.slot === 16)?.model;
	}

	offHand(): Model | undefined {
		return this.held.find(h => h.slot === 17)?.model;
	}

	// swing plays a melee swing: the hand's attack, on the upper body while moving.
	swing(offHand: boolean, moving: boolean, special = false, specialAnim?: number) {
		const m = this.model;
		if (!m) return;
		let anim: number = this.mainHandEmpty ? Anim.AttackUnarmed : this.twoHandedLoose ? Anim.Attack2HL : this.twoHanded ? Anim.Attack2H : Anim.Attack1H;
		if (offHand) anim = this.mainHandEmpty ? Anim.AttackUnarmedOff : Anim.AttackOff;
		if (special) anim = specialAnim ?? (this.twoHanded ? Anim.Special2H : Anim.Special1H);
		// A swing with the weapons away draws them at once.
		this.setDrawn(true);
		this.readyUntil = this.clock + 6;
		if (moving) m.playUpper(anim);
		else {
			m.play(anim, { loop: false, restart: true, fade: 80 });
			this.oneShot = true;
			this.oneShotUpper = true;
		}
	}

	// castStart starts the precast loop for a cast with a cast time.
	castStart(precast: number = Anim.ReadySpellDirected) {
		this.casting = { precast, loop: true };
	}

	// castEnd plays the cast's release (SpellCastDirected unless the spell names another).
	castEnd(anim: number = Anim.SpellCastDirected, moving = false) {
		this.casting = undefined;
		const m = this.model;
		if (!m) return;
		if (moving) m.playUpper(anim);
		else {
			m.play(anim, { loop: false, restart: true, fade: 80 });
			this.oneShot = true;
			this.oneShotUpper = true;
		}
	}

	castStop() {
		this.casting = undefined;
	}

	// wound plays the flinch when the unit takes a hit, the bigger one for a crit.
	wound(crit = false) {
		const m = this.model;
		if (!m) return;
		const playing = m.baseAnim === Anim.CombatWound || m.baseAnim === Anim.CombatCritical;
		if (this.oneShot && !m.baseDone && !(playing && crit)) return;
		m.play(crit ? Anim.CombatCritical : Anim.CombatWound, { loop: false, restart: true, fade: 60 });
		this.oneShot = true;
		this.oneShotUpper = false;
	}

	update(dt: number, motion: Motion, combat: boolean) {
		this.clock += dt;
		const m = this.model;
		if (!m) return;
		for (let i = this.timers.length - 1; i >= 0; i--) {
			if (this.timers[i].at <= this.clock) {
				const t = this.timers[i];
				this.timers.splice(i, 1);
				t.fn();
			}
		}
		if (this.hits.some(h => h.deadline <= this.clock)) this.flushHits();
		const speed = Math.hypot(motion.forward, motion.left);
		const moving = speed > 0.1;
		const back = motion.forward < -0.1;
		if (combat) this.readyUntil = Math.max(this.readyUntil, this.clock + 0.5);
		// Auto attack on draws the weapons, the way the client does when you start attacking.
		if (combat && !this.drawn && this.drawing === undefined) this.setDrawn(true, true);

		// The model has no sideways run, so the client turns the legs toward where you move and
		// twists the upper body back to where you face: 45 degrees on a diagonal, 90 on a strafe.
		// Backing up diagonally, the legs face away from where they go.
		if (!motion.airborne) {
			const target = !moving ? 0 : back ? Math.atan2(-motion.left, -motion.forward) : Math.atan2(motion.left, motion.forward);
			this.bodyYaw += angleTo(this.bodyYaw, target) * Math.min(1, dt * 12);
		}
		m.root.rotation.y = this.bodyYaw;
		m.twist = this.twists ? -this.bodyYaw : 0;

		if (motion.airborne) {
			if (!this.wasAirborne) {
				m.play(Anim.JumpStart, { loop: false, fade: 60 });
				this.oneShot = false;
			} else if (m.baseAnim === Anim.JumpStart && m.baseDone) m.play(Anim.Jump, { loop: true, fade: 60 });
		} else if (this.wasAirborne) {
			m.play(moving ? Anim.JumpLandRun : Anim.JumpEnd, { loop: false, fade: 60 });
			this.oneShot = true;
			this.oneShotUpper = false;
		} else if (moving) {
			// A swing or a release in progress carries on in the upper body.
			if (this.oneShot && this.oneShotUpper && !m.baseDone) m.baseToUpper();
			this.oneShot = false;
			const landing = m.baseAnim === Anim.JumpLandRun && !m.baseDone;
			const anim: number = landing ? Anim.JumpLandRun : back ? Anim.WalkBackwards : motion.walking ? Anim.Walk : Anim.Run;
			// The legs keep pace with the ground: the run plays faster when you run faster.
			m.play(anim, { loop: !landing, speed: this.paceRate(anim, speed), fade: 150 });
		} else if (this.oneShot && !m.baseDone) {
			// A swing, a cast's release or a landing plays to its end.
		} else if (this.casting) {
			this.oneShot = false;
			m.play(this.casting.precast, { loop: true, fade: 120 });
		} else if (Math.abs(motion.turn ?? 0) > 0.5) {
			// Turning in place shuffles the feet.
			this.oneShot = false;
			m.play((motion.turn ?? 0) > 0 ? Anim.ShuffleLeft : Anim.ShuffleRight, { loop: true, fade: 150 });
		} else {
			this.oneShot = false;
			m.play(this.clock < this.readyUntil ? this.readyAnim() : Anim.Stand, { loop: true, fade: 200 });
		}
		this.wasAirborne = motion.airborne;
		if (this.clock > this.readyUntil + 4 && !combat && this.drawn && this.drawing === undefined) this.setDrawn(false, true);
		m.update(dt);
		for (const it of this.items) it.update(dt);
	}

	// paceRate is how fast a movement animation plays so the feet match the ground speed. The
	// sequence's own speed is its moveSpeed, or its fallback's (the backpedal has none: it's the
	// walk played backwards).
	private paceRate(anim: number, speed: number): number {
		const m = this.model!;
		const seqSpeed = (a: number) => {
			const i = m.m2.sequenceFor(a);
			return i >= 0 ? (m.m2.sequences[m.m2.resolve(i)]?.moveSpeed ?? 0) : 0;
		};
		let moveSpeed = seqSpeed(anim);
		if (moveSpeed <= 0 && anim === Anim.WalkBackwards) moveSpeed = seqSpeed(Anim.Walk);
		if (moveSpeed <= 0) return 1;
		const scale = m.root.scale.x || 1;
		return Math.max(0.5, Math.min(2.5, speed / (moveSpeed * scale)));
	}
}
