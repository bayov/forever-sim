// Spell effects and sounds in the room, from the client's spell visual tables (the server reads
// them, cmd/practice/visuals.go): the glow in your hands while you cast, the bolt that flies to
// the dummy, the flash where it lands, Lightning Shield's orbs while the aura is up, the imbue's
// glow on your weapon, the totems' own models, and the swing and impact sounds of melee.

import * as THREE from 'three';
import { Model } from './model';
import { SoundRef, Sounds } from './sound';
import type { UnitView } from './unit';

type KitSound = SoundRef & { kit?: number; weight?: number; loop?: boolean };
// anim is the effect model's own animations: start (birth), loop (hold, 158 for most) and end.
type KitModel = {
	fdid: number;
	attachment: number;
	scale?: number;
	offset?: [number, number, number];
	yaw?: number;
	anim?: { start?: number | null; loop?: number | null; end?: number | null };
};

// playEffect plays an effect model's hold animation when it has one, else its first.
function playEffect(model: Model, km: KitModel) {
	const loop = km.anim?.loop;
	if (loop != null && model.has(loop)) model.play(loop, { loop: true, fade: 0 });
	else if (model.has(158)) model.play(158, { loop: true, fade: 0 });
	else model.play(model.m2.sequences[0]?.id ?? 0, { loop: true, fade: 0 });
}

export type VisualEvent = {
	event: string;
	until?: string;
	target: 'caster' | 'target' | 'area';
	delayMs?: number;
	sounds?: KitSound[];
	models?: KitModel[];
	selfOnly?: boolean;
};

export type SpellVisual = {
	spell: number;
	self?: boolean;
	anims: { precast?: number | null; cast?: number | null; channel?: number | null };
	events: VisualEvent[];
	missile?: { model?: KitModel; speed: number; attachment?: number; destinationAttachment?: number; sound?: KitSound[] } | null;
	totem?: { element: string; display: number; model: number } | null;
};

export type MeleeSounds = { swing?: SoundRef[]; swingCrit?: SoundRef[]; impact?: SoundRef[]; impactCrit?: SoundRef[] };

const visuals = new Map<number, Promise<SpellVisual | undefined>>();

export function spellVisual(spellId: number): Promise<SpellVisual | undefined> {
	let p = visuals.get(spellId);
	if (!p) {
		p = fetch(`/api/visuals/spell/${spellId}`)
			.then(r => (r.ok ? r.json() : undefined))
			.catch(() => undefined);
		visuals.set(spellId, p);
	}
	return p;
}

// kits groups a list of sounds by their kit: the client plays one file of each kit, picked at
// random by weight.
function kits(sounds: KitSound[] | undefined): KitSound[][] {
	const by = new Map<number, KitSound[]>();
	for (const s of sounds ?? []) {
		const k = s.kit ?? s.fdid;
		if (!by.has(k)) by.set(k, []);
		by.get(k)!.push(s);
	}
	return [...by.values()];
}

// An effect model on a unit, which ends when its time is up or when it's stopped.
type Live = { model: Model; until: number; stopped: boolean; fadeUntil?: number };

export class Effects {
	private live: Live[] = [];
	private missiles: { model: Model; from: THREE.Vector3; to: () => THREE.Vector3; speed: number; pos: THREE.Vector3 }[] = [];
	private precast: Live[] = [];
	private auras = new Map<number, Live[]>();
	private imbue?: { name: string; models: Model[] };
	private clock = 0;
	melee: MeleeSounds = {};

	constructor(
		readonly world: THREE.Object3D,
		readonly sounds: Sounds,
	) {}

	// attach loads an effect model onto a unit's attachment point (or its feet).
	private async attach(view: UnitView | undefined, km: KitModel, seconds: number): Promise<Live | undefined> {
		const unit = view?.model;
		if (!unit) return undefined;
		const model = await Model.load(km.fdid).catch(e => {
			console.warn('effect model', km.fdid, e);
			return undefined;
		});
		if (!model) return undefined;
		const point = unit.attachPoint(km.attachment) ?? unit.root;
		if (km.offset) model.root.position.set(km.offset[0], km.offset[2], -km.offset[1]);
		if (km.scale) model.root.scale.setScalar(km.scale);
		if (km.yaw) model.root.rotation.y = km.yaw;
		point.add(model.root);
		model.enableParticles(this.world);
		playEffect(model, km);
		const live: Live = { model, until: this.clock + seconds, stopped: false };
		this.live.push(live);
		return live;
	}

	private playSounds(sounds: KitSound[] | undefined, at?: THREE.Object3D) {
		for (const kit of kits(sounds)) {
			// Looping sounds (a precast hum) are left out: we have no way to stop them yet.
			if (kit[0].loop) continue;
			this.sounds.play(kit, at);
		}
	}

	// event plays one visual event on a unit: its sounds and effect models.
	private async event(ev: VisualEvent, view: UnitView | undefined, seconds: number): Promise<Live[]> {
		const delay = ev.delayMs ?? 0;
		if (delay > 0) await new Promise(r => setTimeout(r, delay));
		this.playSounds(ev.sounds, view?.model?.root);
		const out: Live[] = [];
		for (const km of ev.models ?? []) {
			const l = await this.attach(view, km, seconds);
			if (l) out.push(l);
		}
		return out;
	}

	// castStart shows the precast glow and returns the precast animation to hold.
	async castStart(spellId: number, caster: UnitView | undefined): Promise<number | undefined> {
		const v = await spellVisual(spellId);
		if (!v) return undefined;
		this.castStop();
		for (const ev of v.events.filter(e => e.event === 'precast')) {
			this.event(ev, caster, 60).then(l => this.precast.push(...l));
		}
		return v.anims.precast ?? undefined;
	}

	castStop() {
		for (const l of this.precast) l.stopped = true;
		this.precast = [];
	}

	// cast plays a spell going off: the cast kit on the caster and the missile. It gives the
	// caster's cast animation.
	async cast(spellId: number, caster: UnitView | undefined, target: UnitView | undefined): Promise<SpellVisual | undefined> {
		const v = await spellVisual(spellId);
		this.castStop();
		if (!v) return undefined;
		for (const ev of v.events.filter(e => e.event === 'cast')) this.event(ev, caster, 1.2);
		const m = v.missile;
		if (m?.model && m.speed > 0 && caster?.model && target?.model) this.launch(m, caster, target);
		return v;
	}

	private async launch(m: NonNullable<SpellVisual['missile']>, caster: UnitView, target: UnitView) {
		const model = await Model.load(m.model!.fdid).catch(() => undefined);
		if (!model || !caster.model || !target.model) return;
		const from = new THREE.Vector3();
		(caster.model.attachPoint(m.attachment ?? 38) ?? caster.model.root).getWorldPosition(from);
		const dest = target.model.attachPoint(m.destinationAttachment ?? 34) ?? target.model.root;
		const to = () => dest.getWorldPosition(new THREE.Vector3());
		model.root.position.copy(from);
		this.world.add(model.root);
		model.enableParticles(this.world);
		playEffect(model, m.model!);
		for (const kit of kits(m.sound)) if (!kit[0].loop) this.sounds.play(kit, model.root);
		this.missiles.push({ model, from, to, speed: m.speed, pos: from.clone() });
	}

	// impact plays where a spell lands.
	async impact(spellId: number, target: UnitView | undefined, caster: UnitView | undefined) {
		const v = await spellVisual(spellId);
		if (!v) return;
		for (const ev of v.events.filter(e => e.event === 'impact' && !e.selfOnly)) {
			const on = ev.target === 'caster' || v.self ? caster : target;
			this.event(ev, on, 1.5);
		}
	}

	// setAuras keeps an aura's effect (Lightning Shield's orbs) on you while it's up.
	async setAuras(spellIds: number[], self: UnitView | undefined) {
		const want = new Set(spellIds);
		for (const [id, lives] of this.auras) {
			if (!want.has(id)) {
				for (const l of lives) l.stopped = true;
				this.auras.delete(id);
			}
		}
		for (const id of want) {
			if (this.auras.has(id)) continue;
			this.auras.set(id, []);
			const v = await spellVisual(id);
			const evs = v?.events.filter(e => e.event === 'aura') ?? [];
			for (const ev of evs) {
				const lives = await this.event({ ...ev, sounds: [] }, self, 3600);
				if (this.auras.has(id)) this.auras.get(id)!.push(...lives);
				else for (const l of lives) l.stopped = true;
			}
		}
	}

	// setImbue puts the imbue's glow on the main hand weapon, by its name ("WindfuryWeapon").
	async setImbue(name: string | undefined, weapon: Model | undefined) {
		const key = name ?? '';
		if (this.imbue?.name === key) return;
		for (const m of this.imbue?.models ?? []) m.dispose();
		this.imbue = { name: key, models: [] };
		if (!name || !weapon) return;
		const imbues: Record<string, number[]> = await fetch('/api/visuals/imbues')
			.then(r => r.json())
			.catch(() => ({}));
		const ids = Object.entries(imbues).find(([n]) => n.replace(/\s/g, '').toLowerCase() === name.replace(/\s/g, '').toLowerCase())?.[1];
		if (!ids?.length) return;
		const glow: { models?: KitModel[] } = await fetch(`/api/visuals/enchant/${ids[ids.length - 1]}`)
			.then(r => r.json())
			.catch(() => ({}));
		for (const km of glow.models ?? []) {
			const model = await Model.load(km.fdid).catch(() => undefined);
			if (!model || this.imbue.name !== key) {
				model?.dispose();
				continue;
			}
			// The glow's slot is the weapon model's attachment with the same number.
			(weapon.attachPoint(km.attachment) ?? weapon.root).add(model.root);
			model.enableParticles(this.world);
			playEffect(model, km);
			this.imbue.models.push(model);
		}
	}

	swing(at: THREE.Object3D | undefined, crit: boolean, landed: boolean) {
		this.sounds.play(crit ? (this.melee.swingCrit ?? this.melee.swing) : this.melee.swing, at, 'swing');
		if (landed) this.sounds.play(crit ? (this.melee.impactCrit ?? this.melee.impact) : this.melee.impact, at, 'impact');
	}

	update(dt: number) {
		this.clock += dt;
		for (let i = this.live.length - 1; i >= 0; i--) {
			const l = this.live[i];
			if (!l.stopped && this.clock >= l.until) l.stopped = true;
			if (l.stopped && l.fadeUntil === undefined) {
				// Stop emitting and hide the mesh, then let the particles die out.
				l.fadeUntil = this.clock + 2;
				if (l.model.particles) l.model.particles.emitting = false;
				l.model.mesh.visible = false;
			}
			if (l.fadeUntil !== undefined && this.clock >= l.fadeUntil) {
				l.model.dispose();
				this.live.splice(i, 1);
				continue;
			}
			l.model.update(dt);
		}
		for (let i = this.missiles.length - 1; i >= 0; i--) {
			const m = this.missiles[i];
			const to = m.to();
			const step = m.speed * dt;
			const d = to.clone().sub(m.pos);
			if (d.length() <= step) {
				m.model.dispose();
				this.missiles.splice(i, 1);
				continue;
			}
			d.normalize();
			m.pos.addScaledVector(d, step);
			m.model.root.position.copy(m.pos);
			// The missile flies nose first (+x).
			m.model.root.quaternion.setFromUnitVectors(new THREE.Vector3(1, 0, 0), d);
			m.model.update(dt);
		}
		for (const m of this.imbue?.models ?? []) m.update(dt);
	}
}
