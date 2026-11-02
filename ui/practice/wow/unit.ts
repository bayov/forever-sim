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
	scale?: number;
	geosets?: number[];
	textures?: Record<string, Composite>;
	replaceable?: Record<string, number>;
	attachments?: AppearanceAttachment[];
	collections?: { slot: number; model: number; textures?: number[]; replaceable?: Record<string, number>; geosets?: number[] }[];
};

function applyReplaceable(m: Model, replaceable: Record<string, number> | undefined, textures: number[] | undefined) {
	const byType = replaceable ?? (textures?.[0] ? { '2': textures[0] } : {});
	for (const [type, fdid] of Object.entries(byType)) {
		texture(fdid)
			.then(t => m.setTexture(Number(type), t))
			.catch(e => console.warn('texture', fdid, e));
	}
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

type Held = { model: Model; slot: number; hand?: THREE.Object3D; sheath?: THREE.Object3D; hiddenSheathed?: boolean };

export type Motion = {
	// Velocity relative to facing, in yards a second: forward and left.
	forward: number;
	left: number;
	airborne: boolean;
	walking: boolean;
};

export class UnitView {
	model?: Model;
	private items: Model[] = [];
	private held: Held[] = [];
	private drawn = false;
	private wasAirborne = false;
	private oneShot = false;
	private casting?: { precast: number; loop: boolean };
	private readyUntil = 0;
	private twoHanded = false;
	private hasOffHand = false;
	private mainHandEmpty = true;
	clock = 0;

	constructor(
		readonly group: THREE.Group,
		readonly placeholder: THREE.Object3D[],
	) {}

	async load(a: Appearance) {
		await prefetch([a.model, ...(a.attachments ?? []).map(x => x.model), ...(a.collections ?? []).map(x => x.model)]);
		const model = await Model.load(a.model);
		if (a.geosets) model.setGeosets(a.geosets);
		for (const [type, comp] of Object.entries(a.textures ?? {})) {
			compose(comp)
				.then(t => model.setTexture(Number(type), t))
				.catch(e => console.warn('texture', type, e));
		}
		for (const [type, fdid] of Object.entries(a.replaceable ?? {})) {
			texture(fdid)
				.then(t => model.setTexture(Number(type), t))
				.catch(e => console.warn('texture', type, e));
		}
		if (a.scale) model.root.scale.setScalar(a.scale);
		for (const c of a.collections ?? []) {
			// Capes, belts and such are drawn with the body's bones. We give them our skeleton.
			Model.load(c.model)
				.then(m => {
					if (c.geosets) m.setGeosets(c.geosets);
					applyReplaceable(m, c.replaceable, c.textures);
					m.mesh.bind(model.mesh.skeleton, new THREE.Matrix4());
					model.root.add(m.mesh);
					this.items.push(m);
				})
				.catch(e => console.warn('collection', c.model, e));
		}
		for (const att of a.attachments ?? []) {
			Model.load(att.model)
				.then(m => {
					applyReplaceable(m, att.replaceable, att.textures);
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
					this.held.push({ model: m, slot: att.slot, hand, sheath, hiddenSheathed: hidden });
					if (att.slot === 16) {
						this.mainHandEmpty = false;
						this.twoHanded = att.sheathe === 1 || att.sheathe === 2;
					}
					if (offHand) this.hasOffHand = true;
					this.placeWeapons();
				})
				.catch(e => console.warn('attachment', att.model, e));
		}
		this.model = model;
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

	setDrawn(on: boolean) {
		if (on === this.drawn) return;
		this.drawn = on;
		this.placeWeapons();
	}

	private readyAnim() {
		if (this.mainHandEmpty) return Anim.ReadyUnarmed;
		return this.twoHanded ? Anim.Ready2H : Anim.Ready1H;
	}

	// swing plays a melee swing: the hand's attack, on the upper body while moving.
	// mainHand is the main hand weapon's model, for the imbue's glow.
	mainHand(): Model | undefined {
		return this.held.find(h => h.slot === 16)?.model;
	}

	swing(offHand: boolean, moving: boolean, special = false, specialAnim?: number) {
		const m = this.model;
		if (!m) return;
		let anim: number = this.mainHandEmpty ? Anim.AttackUnarmed : this.twoHanded ? Anim.Attack2H : Anim.Attack1H;
		if (offHand) anim = this.mainHandEmpty ? Anim.AttackUnarmedOff : Anim.AttackOff;
		if (special) anim = specialAnim ?? (this.twoHanded ? Anim.Special2H : Anim.Special1H);
		this.setDrawn(true);
		this.readyUntil = this.clock + 6;
		if (moving) m.playUpper(anim);
		else {
			m.play(anim, { loop: false, restart: true, fade: 80 });
			this.oneShot = true;
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
		}
	}

	castStop() {
		this.casting = undefined;
	}

	// wound plays the flinch when the unit takes a hit.
	wound() {
		const m = this.model;
		if (!m || this.oneShot) return;
		m.play(Anim.CombatWound, { loop: false, restart: true, fade: 60 });
		this.oneShot = true;
	}

	update(dt: number, motion: Motion, combat: boolean) {
		this.clock += dt;
		const m = this.model;
		if (!m) return;
		const speed = Math.hypot(motion.forward, motion.left);
		const moving = speed > 0.1;
		if (combat) this.readyUntil = Math.max(this.readyUntil, this.clock + 0.5);

		if (motion.airborne) {
			if (!this.wasAirborne) {
				m.play(Anim.JumpStart, { loop: false, fade: 60 });
				this.oneShot = false;
			} else if (m.baseAnim === Anim.JumpStart && m.baseDone) m.play(Anim.Jump, { loop: true, fade: 60 });
		} else if (this.wasAirborne) {
			m.play(moving ? Anim.JumpLandRun : Anim.JumpEnd, { loop: false, fade: 60 });
			this.oneShot = true;
		} else if (moving) {
			this.oneShot = false;
			let anim: number = motion.walking ? Anim.Walk : Anim.Run;
			if (motion.forward < -0.1) anim = Anim.WalkBackwards;
			else if (Math.abs(motion.forward) <= 0.1) anim = motion.left > 0 ? Anim.RunLeft : Anim.RunRight;
			// The legs keep pace with the ground: the run plays faster when you run faster.
			const seq = m.m2.sequences[m.m2.resolve(Math.max(0, m.m2.sequenceFor(anim)))];
			const rate = seq && seq.moveSpeed > 0 ? speed / seq.moveSpeed : 1;
			m.play(anim, { loop: true, speed: Math.max(0.5, Math.min(2, rate)), fade: 150 });
		} else if (this.oneShot && !m.baseDone) {
			// A swing, a cast's release or a landing plays to its end.
		} else if (this.casting) {
			this.oneShot = false;
			m.play(this.casting.precast, { loop: true, fade: 120 });
		} else {
			this.oneShot = false;
			m.play(this.clock < this.readyUntil ? this.readyAnim() : Anim.Stand, { loop: true, fade: 200 });
		}
		this.wasAirborne = motion.airborne;
		if (this.clock > this.readyUntil + 4 && !combat) this.setDrawn(false);
		m.update(dt);
		for (const it of this.items) it.update(dt);
	}
}
