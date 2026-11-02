// Spell effects and sounds in the room, from the client's spell visual tables (the server reads
// them, cmd/practice/visuals.go): the glow in your hands while you cast, the bolt that flies to
// the dummy, the flash where it lands, Lightning Shield's orbs while the aura is up, the imbue's
// glow on your weapon, the totems' own models, and the swing and impact sounds of melee.

import * as THREE from 'three';
import { prefetch } from './files';
import { Model } from './model';
import { Playing, SoundRef, Sounds } from './sound';
import type { UnitView } from './unit';

type KitSound = SoundRef;
// anim is the effect model's own animations: start (birth), loop (hold, 158 for most) and end.
type KitModel = {
	fdid: number;
	attachment: number;
	scale?: number;
	offset?: [number, number, number];
	yaw?: number;
	anim?: { start?: number | null; loop?: number | null; end?: number | null };
};

// The effect model animations most kits use: 0 the birth, 158 the hold and 159 the decay.
const Birth = 0;
const Hold = 158;
const Decay = 159;

function holdAnim(model: Model, km: KitModel): number | undefined {
	const loop = km.anim?.loop ?? Hold;
	return model.has(loop) ? loop : undefined;
}

// playEffect starts an effect model. An effect that holds (a precast glow, an aura, an imbue)
// plays its birth once, then its hold until it stops. A one-shot (a cast flash, an impact) plays
// its birth once. It gives how long the first animation runs, in seconds.
function playEffect(model: Model, km: KitModel, hold: boolean): number {
	const start = km.anim?.start ?? Birth;
	const loop = holdAnim(model, km);
	if (hold && loop !== undefined && (!model.has(start) || start === loop)) {
		model.play(loop, { loop: true, fade: 0 });
		return Infinity;
	}
	const first = model.has(start) ? start : (model.m2.sequences[0]?.id ?? 0);
	const seq = model.play(first, { loop: false, fade: 0, restart: true });
	return (model.m2.sequences[seq]?.duration ?? 0) / 1000;
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

const imbueList = () =>
	(imbuesP ??= fetch('/api/visuals/imbues')
		.then(r => r.json() as Promise<Record<string, number[]>>)
		.catch(() => ({}) as Record<string, number[]>));
let imbuesP: Promise<Record<string, number[]>> | undefined;

const enchantVisuals = new Map<number, Promise<{ models?: KitModel[] }>>();
function enchantVisual(id: number): Promise<{ models?: KitModel[] }> {
	let p = enchantVisuals.get(id);
	if (!p) {
		p = fetch(`/api/visuals/enchant/${id}`)
			.then(r => r.json())
			.catch(() => ({}));
		enchantVisuals.set(id, p);
	}
	return p;
}

// imbueEnchant gives the enchant id of an imbue by its name ("WindfuryWeapon"): the top rank's.
async function imbueEnchant(name: string): Promise<number | undefined> {
	const imbues = await imbueList();
	const want = name.replace(/\s/g, '').toLowerCase();
	const ids = Object.entries(imbues).find(([n]) => n.replace(/\s/g, '').toLowerCase() === want)?.[1];
	return ids?.length ? ids[ids.length - 1] : undefined;
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

// An effect model on a unit. A one-shot ends when its animation does. One that holds runs its
// hold until it's stopped, then its decay.
type Live = { model: Model; km: KitModel; hold: boolean; until: number; stopped: boolean; ending?: boolean; fadeUntil?: number };

export class Effects {
	private live: Live[] = [];
	private lastImpact = new Map<string, number>();
	private totemCasting = new Map<Model, { lives: Live[]; sounds: Promise<Playing | undefined>[] }>();
	private missiles: { onLand?: () => void; model: Model; km: KitModel; from: THREE.Vector3; to: () => THREE.Vector3; speed: number; pos: THREE.Vector3; sounds: Promise<Playing | undefined>[] }[] = [];
	private precast: Live[] = [];
	// The precast hum, a loop that stops when the cast ends.
	private precastSounds: Promise<Playing | undefined>[] = [];
	private auras = new Map<number, { lives: Live[]; sounds: Promise<Playing | undefined>[] }>();
	private imbues = new Map<string, { name: string; models: { model: Model; km: KitModel }[] }>();
	private clock = 0;
	melee: MeleeSounds = {};
	steps?: SoundRef[];
	grunts?: SoundRef[];
	draw?: SoundRef[];
	sheathe?: SoundRef[];

	constructor(
		readonly world: THREE.Object3D,
		readonly sounds: Sounds,
	) {}

	// attach loads an effect model onto a unit's attachment point (or its feet). It holds until
	// stopped, or plays once.
	private async attach(unit: Model | undefined, km: KitModel, hold: boolean): Promise<Live | undefined> {
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
		const first = playEffect(model, km, hold);
		// A one-shot with no animation of its own shows for a second.
		const seconds = hold ? Infinity : first > 0 ? first : 1;
		const live: Live = { model, km, hold, until: this.clock + seconds, stopped: false };
		this.live.push(live);
		return live;
	}

	// playSounds plays one file of each kit. It gives the loops, which the caller stops.
	private playSounds(sounds: KitSound[] | undefined, at?: THREE.Object3D): Promise<Playing | undefined>[] {
		const loops: Promise<Playing | undefined>[] = [];
		for (const kit of kits(sounds)) {
			const p = this.sounds.play(kit, at);
			if (kit[0].loop) loops.push(p);
		}
		return loops;
	}

	// event plays one visual event on a unit's model: its sounds and effect models. A held event
	// keeps its models and loop sounds (in `loops`) until stopped.
	async event(ev: VisualEvent, unit: Model | undefined, hold: boolean, loops?: Promise<Playing | undefined>[]): Promise<Live[]> {
		const delay = ev.delayMs ?? 0;
		if (delay > 0) await new Promise(r => setTimeout(r, delay));
		const playing = this.playSounds(ev.sounds, unit?.root);
		if (loops) loops.push(...playing);
		else for (const p of playing) p.then(h => h?.stop());
		const out: Live[] = [];
		for (const km of ev.models ?? []) {
			const l = await this.attach(unit, km, hold);
			if (l) out.push(l);
		}
		return out;
	}

	// preload loads every file these spells' effects use (models, their textures and sounds) and
	// the imbues' glows, so nothing loads mid fight.
	async preload(spellIds: number[]) {
		const models = new Set<number>();
		const sounds: KitSound[] = [];
		const vs = await Promise.all(spellIds.map(id => spellVisual(id)));
		for (const v of vs) {
			if (!v) continue;
			for (const ev of v.events) {
				for (const km of ev.models ?? []) models.add(km.fdid);
				sounds.push(...(ev.sounds ?? []));
			}
			if (v.missile?.model) models.add(v.missile.model.fdid);
			sounds.push(...(v.missile?.sound ?? []));
		}
		const imbues = await imbueList();
		for (const name of Object.keys(imbues)) {
			const id = await imbueEnchant(name);
			const glow = id ? await enchantVisual(id) : {};
			for (const km of glow.models ?? []) models.add(km.fdid);
		}
		await prefetch([...models, ...sounds.map(s => s.fdid)]);
		await Promise.all([
			...[...models].map(fdid =>
				Model.load(fdid)
					.then(m => m.dispose())
					.catch(e => console.warn('effect model', fdid, e)),
			),
			this.sounds.preload(sounds),
		]);
	}

	// castStart shows the precast glow and returns the precast animation to hold.
	async castStart(spellId: number, caster: UnitView | undefined): Promise<number | undefined> {
		const v = await spellVisual(spellId);
		if (!v) return undefined;
		this.castStop();
		for (const ev of v.events.filter(e => e.event === 'precast')) {
			this.event(ev, caster?.model, true, this.precastSounds).then(l => this.precast.push(...l));
		}
		return v.anims.precast ?? undefined;
	}

	castStop() {
		for (const l of this.precast) l.stopped = true;
		this.precast = [];
		for (const p of this.precastSounds) p.then(h => h?.stop());
		this.precastSounds = [];
	}

	// cast plays a spell going off: the cast kit on the caster and the missile. It gives the
	// caster's cast animation.
	async cast(spellId: number, caster: UnitView | undefined, target: UnitView | undefined): Promise<SpellVisual | undefined> {
		const v = await spellVisual(spellId);
		this.castStop();
		if (!v) return undefined;
		for (const ev of v.events.filter(e => e.event === 'cast')) this.event(ev, caster?.model, false);
		const m = v.missile;
		if (m?.model && m.speed > 0 && caster?.model && target?.model) this.launch(m, caster.model, target.model);
		return v;
	}

	// totemPrecast shows a totem's glow while it casts (the Searing Totem's bolt), until its
	// next totemCast.
	async totemPrecast(spellId: number, totem: Model | undefined) {
		const v = await spellVisual(spellId);
		if (!v || !totem) return;
		this.totemPrecastStop(totem);
		const held: Live[] = [];
		const sounds: Promise<Playing | undefined>[] = [];
		this.totemCasting.set(totem, { lives: held, sounds });
		for (const ev of v.events.filter(e => e.event === 'precast')) {
			this.event(ev, totem, true, sounds).then(l => {
				held.push(...l);
				if (this.totemCasting.get(totem)?.lives !== held) for (const x of l) x.stopped = true;
			});
		}
	}

	private totemPrecastStop(totem: Model) {
		const c = this.totemCasting.get(totem);
		if (!c) return;
		this.totemCasting.delete(totem);
		for (const l of c.lives) l.stopped = true;
		for (const p of c.sounds) p.then(h => h?.stop());
	}

	// totemCast plays a totem's spell going off: its cast kit on the totem and its sounds there,
	// then the bolt to the target, if it has one. onLand runs when it lands (at once without a
	// bolt).
	async totemCast(spellId: number, totem: Model | undefined, target: Model | undefined, onLand: () => void) {
		const v = await spellVisual(spellId);
		if (!v || !totem) {
			onLand();
			return;
		}
		this.totemPrecastStop(totem);
		for (const ev of v.events.filter(e => e.event === 'cast')) this.event(ev, totem, false);
		const m = v.missile;
		if (m?.model && m.speed > 0 && target) this.launch(m, totem, target, onLand);
		else onLand();
	}

	// launch sends a spell's bolt from one model to another. A totem has no bolt attachment
	// (38), so it shoots from its spell hand (22).
	private async launch(m: NonNullable<SpellVisual['missile']>, from: Model, to: Model, onLand?: () => void) {
		const model = await Model.load(m.model!.fdid).catch(() => undefined);
		if (!model) {
			onLand?.();
			return;
		}
		const start = new THREE.Vector3();
		(from.attachPoint(m.attachment ?? 38) ?? from.attachPoint(22) ?? from.root).getWorldPosition(start);
		const dest = to.attachPoint(m.destinationAttachment ?? 34) ?? to.root;
		const target = () => dest.getWorldPosition(new THREE.Vector3());
		model.root.position.copy(start);
		this.world.add(model.root);
		model.enableParticles(this.world);
		playEffect(model, m.model!, true);
		const sounds = kits(m.sound).map(kit => this.sounds.play(kit, model.root));
		this.missiles.push({ model, km: m.model!, from: start, to: target, speed: m.speed, pos: start.clone(), sounds, onLand });
	}

	// buffBurst plays a buff's visual once on a unit (a totem's buff landing on you).
	async buffBurst(spellId: number, on: UnitView | undefined) {
		const v = await spellVisual(spellId);
		if (!v) return;
		for (const ev of v.events.filter(e => e.event !== 'precast' && e.event !== 'cast')) this.event(ev, on?.model, false);
	}

	// spellImpact plays where a spell lands.
	async spellImpact(spellId: number, target: UnitView | undefined, caster: UnitView | undefined) {
		// Two hits of one spell at once (a pulse that hits twice) show one impact.
		const now = performance.now();
		const key = `${spellId}`;
		if (now - (this.lastImpact.get(key) ?? -1e9) < 150) return;
		this.lastImpact.set(key, now);
		const v = await spellVisual(spellId);
		if (!v) return;
		for (const ev of v.events.filter(e => e.event === 'impact' && !e.selfOnly)) {
			const on = ev.target === 'caster' || v.self ? caster : target;
			this.event(ev, on?.model, false);
		}
	}

	// setAuras keeps an aura's effect (Lightning Shield's orbs) on you while it's up.
	async setAuras(spellIds: number[], self: UnitView | undefined) {
		const want = new Set(spellIds);
		for (const [id, a] of this.auras) {
			if (!want.has(id)) {
				for (const l of a.lives) l.stopped = true;
				for (const p of a.sounds) p.then(h => h?.stop());
				this.auras.delete(id);
			}
		}
		for (const id of want) {
			if (this.auras.has(id)) continue;
			const a = { lives: [] as Live[], sounds: [] as Promise<Playing | undefined>[] };
			this.auras.set(id, a);
			const v = await spellVisual(id);
			const evs = v?.events.filter(e => e.event === 'aura') ?? [];
			for (const ev of evs) {
				const lives = await this.event(ev, self?.model, true, a.sounds);
				a.lives.push(...lives);
				if (this.auras.get(id) !== a) {
					// The aura ended while its effect loaded.
					for (const l of lives) l.stopped = true;
					for (const p of a.sounds) p.then(h => h?.stop());
				}
			}
		}
	}

	// setImbue puts an imbue's glow on a weapon, by the imbue's name ("WindfuryWeapon"). hand
	// is 'main' or 'off'.
	async setImbue(hand: string, name: string | undefined, weapon: Model | undefined) {
		const key = `${name ?? ''}@${weapon ? weapon.m2.name : ''}`;
		const was = this.imbues.get(hand);
		if (was?.name === key) return;
		for (const m of was?.models ?? []) m.model.dispose();
		const now = { name: key, models: [] as { model: Model; km: KitModel }[] };
		this.imbues.set(hand, now);
		if (!name || !weapon) return;
		const id = await imbueEnchant(name);
		if (!id) return;
		const glow = await enchantVisual(id);
		for (const km of glow.models ?? []) {
			const model = await Model.load(km.fdid).catch(() => undefined);
			if (!model || this.imbues.get(hand) !== now) {
				model?.dispose();
				continue;
			}
			// The glow's slot is the weapon model's attachment with the same number.
			(weapon.attachPoint(km.attachment) ?? weapon.root).add(model.root);
			model.enableParticles(this.world);
			playEffect(model, km, true);
			now.models.push({ model, km });
		}
	}

	// swingSound is the weapon's whoosh, when the swing's animation gets to it.
	swingSound(at: THREE.Object3D | undefined) {
		this.sounds.play(this.melee.swing, at, 'swing');
	}

	// impact is a white swing landing: the weapon hitting the target.
	impact(at: THREE.Object3D | undefined, crit: boolean, landed: boolean) {
		if (landed) this.sounds.play(crit ? (this.melee.impactCrit ?? this.melee.impact) : this.melee.impact, at, 'impact');
	}

	// grunt is your character's exertion on a crit or a special attack. The kit's own cooldown
	// (1 to 2 seconds) keeps it from playing every swing.
	grunt(at: THREE.Object3D | undefined) {
		this.sounds.play(this.grunts, at, 'grunt');
	}

	// unitSound plays a sound of your character's animations.
	unitSound(kind: 'swing' | 'footstep' | 'draw' | 'sheathe', at: THREE.Object3D | undefined) {
		switch (kind) {
			case 'swing':
				return this.swingSound(at);
			case 'footstep':
				return this.footstep(at);
			case 'draw':
				return this.sounds.play(this.draw, at, 'draw');
			case 'sheathe':
				return this.sounds.play(this.sheathe, at, 'sheathe');
		}
	}

	// footstep plays a step on the ground.
	footstep(at: THREE.Object3D | undefined) {
		this.sounds.play(this.steps, at, 'step');
	}

	update(dt: number) {
		this.clock += dt;
		for (let i = this.live.length - 1; i >= 0; i--) {
			const l = this.live[i];
			const m = l.model;
			if (l.hold && !l.stopped && m.baseDone) {
				// The birth is done: on to the hold.
				const loop = holdAnim(m, l.km);
				if (loop !== undefined) m.play(loop, { loop: true, fade: 0 });
			}
			if (!l.stopped && this.clock >= l.until) l.stopped = true;
			if (l.stopped && l.hold && l.ending === undefined) {
				// A held effect plays its decay, when it has one, before it goes.
				const end = l.km.anim?.end ?? Decay;
				l.ending = m.has(end);
				if (l.ending) {
					const seq = m.play(end, { loop: false, restart: true, fade: 0 });
					l.until = this.clock + (m.m2.sequences[seq]?.duration ?? 0) / 1000;
					m.emitting = false;
				}
			}
			if (l.ending && this.clock >= l.until) l.ending = false;
			if (l.stopped && !l.ending && l.fadeUntil === undefined) {
				// Stop emitting and hide the mesh, then let the particles die out.
				l.fadeUntil = this.clock + 2;
				m.emitting = false;
				m.mesh.visible = false;
			}
			if (l.fadeUntil !== undefined && this.clock >= l.fadeUntil) {
				m.dispose();
				this.live.splice(i, 1);
				continue;
			}
			m.update(dt);
		}
		for (let i = this.missiles.length - 1; i >= 0; i--) {
			const m = this.missiles[i];
			const to = m.to();
			const step = m.speed * dt;
			const d = to.clone().sub(m.pos);
			if (d.length() <= step) {
				for (const p of m.sounds) p.then(h => h?.stop(100));
				m.model.dispose();
				m.onLand?.();
				this.missiles.splice(i, 1);
				continue;
			}
			d.normalize();
			m.pos.addScaledVector(d, step);
			m.model.root.position.copy(m.pos);
			// The missile flies nose first (+x).
			m.model.root.quaternion.setFromUnitVectors(new THREE.Vector3(1, 0, 0), d);
			if (m.model.baseDone) {
				const loop = holdAnim(m.model, m.km);
				if (loop !== undefined) m.model.play(loop, { loop: true, fade: 0 });
			}
			m.model.update(dt);
		}
		for (const imbue of this.imbues.values()) {
			for (const { model, km } of imbue.models) {
				if (model.baseDone) {
					const loop = holdAnim(model, km);
					if (loop !== undefined) model.play(loop, { loop: true, fade: 0 });
				}
				model.update(dt);
			}
		}
	}
}
