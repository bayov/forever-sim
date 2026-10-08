import { UnitMetadataList } from './player.js';
import { Encounter as EncounterProto, PresetEncounter, PresetTarget, Target as TargetProto } from './proto/common.js';
import { Sim } from './sim.js';
import { EventID, TypedEvent } from './typed_event.js';

const DEFAULT_DURATION = 120;
const DEFAULT_VARIATION = 15;
const DEFAULT_EXECUTE_20 = 0.2;
const DEFAULT_EXECUTE_25 = 0.25;
const DEFAULT_EXECUTE_35 = 0.35;
// The share of a PvP fight spent out of melee range, when PvP mode is first turned on.
const DEFAULT_PVP_MELEE_DOWNTIME = 0.7;

// Manages all the settings for an Encounter.
export class Encounter {
	readonly sim: Sim;

	private duration = DEFAULT_DURATION;
	private durationVariation = DEFAULT_VARIATION;
	private executeProportion20 = DEFAULT_EXECUTE_20;
	private executeProportion25 = DEFAULT_EXECUTE_25;
	private executeProportion35 = DEFAULT_EXECUTE_35;
	private useHealth = false;
	private pvp = false;
	private pvpMeleeDowntime = DEFAULT_PVP_MELEE_DOWNTIME;

	targets!: Array<TargetProto>;
	targetsMetadata: UnitMetadataList;
	presetTargets!: Array<PresetTarget>;

	readonly targetsChangeEmitter = new TypedEvent<void>();
	readonly durationChangeEmitter = new TypedEvent<void>();
	readonly executeProportionChangeEmitter = new TypedEvent<void>();

	// Emits when any of the above emitters emit.
	readonly changeEmitter = new TypedEvent<void>();

	constructor(sim: Sim) {
		this.sim = sim;
		this.targetsMetadata = new UnitMetadataList();

		sim.waitForInit().then(() => {
			const presetTarget = Encounter.getDefaultTarget(sim);

			this.targets = [TargetProto.clone(presetTarget.target!)];

			[this.targetsChangeEmitter, this.durationChangeEmitter, this.executeProportionChangeEmitter].forEach(emitter =>
				emitter.on(eventID => this.changeEmitter.emit(eventID)),
			);
		});
	}

	get primaryTarget(): TargetProto {
		return TargetProto.clone(this.targets[0]);
	}

	getDurationVariation(): number {
		return this.durationVariation;
	}
	setDurationVariation(eventID: EventID, newDuration: number) {
		if (newDuration == this.durationVariation) return;

		this.durationVariation = newDuration;
		this.durationChangeEmitter.emit(eventID);
	}

	getDuration(): number {
		return this.duration;
	}
	setDuration(eventID: EventID, newDuration: number) {
		if (newDuration == this.duration) return;

		this.duration = newDuration;
		this.durationChangeEmitter.emit(eventID);
	}

	getExecuteProportion20(): number {
		return this.executeProportion20;
	}
	setExecuteProportion20(eventID: EventID, newExecuteProportion20: number) {
		if (newExecuteProportion20 == this.executeProportion20) return;

		this.executeProportion20 = newExecuteProportion20;
		this.executeProportionChangeEmitter.emit(eventID);
	}
	getExecuteProportion25(): number {
		return this.executeProportion25;
	}
	setExecuteProportion25(eventID: EventID, newExecuteProportion25: number) {
		if (newExecuteProportion25 == this.executeProportion25) return;

		this.executeProportion25 = newExecuteProportion25;
		this.executeProportionChangeEmitter.emit(eventID);
	}
	getExecuteProportion35(): number {
		return this.executeProportion35;
	}
	setExecuteProportion35(eventID: EventID, newExecuteProportion35: number) {
		if (newExecuteProportion35 == this.executeProportion35) return;

		this.executeProportion35 = newExecuteProportion35;
		this.executeProportionChangeEmitter.emit(eventID);
	}

	getUseHealth(): boolean {
		return this.useHealth;
	}
	setUseHealth(eventID: EventID, newUseHealth: boolean) {
		if (newUseHealth == this.useHealth) return;

		this.useHealth = newUseHealth;
		this.durationChangeEmitter.emit(eventID);
		this.executeProportionChangeEmitter.emit(eventID);
	}

	getPvp(): boolean {
		return this.pvp;
	}
	setPvp(eventID: EventID, newPvp: boolean) {
		if (newPvp == this.pvp) return;

		this.pvp = newPvp;
		this.durationChangeEmitter.emit(eventID);
	}

	getPvpMeleeDowntime(): number {
		return this.pvpMeleeDowntime;
	}
	setPvpMeleeDowntime(eventID: EventID, newDowntime: number) {
		if (newDowntime == this.pvpMeleeDowntime) return;

		this.pvpMeleeDowntime = newDowntime;
		this.durationChangeEmitter.emit(eventID);
	}

	matchesPreset(preset: PresetEncounter): boolean {
		return preset.targets.length == this.targets.length && this.targets.every((t, i) => TargetProto.equals(t, preset.targets[i].target));
	}

	applyPreset(eventID: EventID, preset: PresetEncounter) {
		// We copy the preset's targets, so editing one later doesn't change the preset.
		this.targets = preset.targets.map(presetTarget => (presetTarget.target ? TargetProto.clone(presetTarget.target) : TargetProto.create()));
		this.targetsChangeEmitter.emit(eventID);
	}

	applyPresetTarget(eventID: EventID, preset: PresetTarget, index: number) {
		this.targets[index] = preset.target ? TargetProto.clone(preset.target) : TargetProto.create();
		this.targetsChangeEmitter.emit(eventID);
	}

	toProto(): EncounterProto {
		return EncounterProto.create({
			duration: this.duration,
			durationVariation: this.durationVariation,
			executeProportion20: this.executeProportion20,
			executeProportion25: this.executeProportion25,
			executeProportion35: this.executeProportion35,
			useHealth: this.useHealth,
			pvp: this.pvp,
			// Only a PvP fight carries the downtime, so a preset encounter without PvP still
			// matches the settings.
			pvpMeleeDowntime: this.pvp ? this.pvpMeleeDowntime : 0,
			targets: this.targets,
		});
	}

	fromProto(eventID: EventID, proto: EncounterProto) {
		TypedEvent.freezeAllAndDo(() => {
			this.setDuration(eventID, proto.duration);
			this.setDurationVariation(eventID, proto.durationVariation);
			this.setExecuteProportion20(eventID, proto.executeProportion20);
			this.setExecuteProportion25(eventID, proto.executeProportion25);
			this.setExecuteProportion35(eventID, proto.executeProportion35);
			this.setUseHealth(eventID, proto.useHealth);
			this.setPvp(eventID, proto.pvp);
			// Settings without PvP have no downtime saved, so turning PvP on starts from the default.
			this.setPvpMeleeDowntime(eventID, proto.pvp ? proto.pvpMeleeDowntime : DEFAULT_PVP_MELEE_DOWNTIME);
			// We copy the targets, so editing one later doesn't change the preset or the saved
			// settings they came from.
			this.targets = proto.targets.map(target => TargetProto.clone(target));
			this.targetsChangeEmitter.emit(eventID);
		});
	}

	applyDefaults(eventID: EventID) {
		const presetTarget = Encounter.getDefaultTarget(this.sim);
		this.fromProto(
			eventID,
			EncounterProto.create({
				duration: DEFAULT_DURATION,
				durationVariation: DEFAULT_VARIATION,
				executeProportion20: DEFAULT_EXECUTE_20,
				executeProportion25: DEFAULT_EXECUTE_25,
				executeProportion35: DEFAULT_EXECUTE_35,
				targets: [presetTarget.target!],
			}),
		);
	}

	static getDefaultTarget(sim: Sim): PresetTarget {
		const presetTargets = sim.db.getAllPresetTargets();
		return presetTargets[0];
	}
}
