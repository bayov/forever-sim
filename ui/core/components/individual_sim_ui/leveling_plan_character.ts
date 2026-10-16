import { Player } from '../../player';
import { RaidSimRequest } from '../../proto/api';
import { APLRotation } from '../../proto/apl';
import { Spec } from '../../proto/common';
import { Stats } from '../../proto_utils/stats';
import { Sim } from '../../sim';
import { TypedEvent } from '../../typed_event';

// The plan's character at one level of the leveling plan, for the parts of the Leveling Plan
// tab that need a whole character to show or edit something.
//
// The rotation editor lists the spells the character has, and those come from its level and
// talents. The item list of the gear picker has the items it can wear at its level, sorted by
// the stat weights of that level. So we keep a second character next to the one of the
// settings. It's the request of that level as the plan sims it, so it has the plan's gear,
// talents and rotation there, and nothing we do with it changes the character of the settings.
//
// It's not in the raid. Its unit references, like the rotation's "Self", are those of the
// settings' character, the first one in the raid.
export class LevelingPlanCharacter {
	readonly player: Player<Spec>;
	private readonly sim: Sim;
	private readonly makeRequest: (level: number) => RaidSimRequest | null;
	private loaded = 0;
	// True while a load puts the level's settings in place, so the parts that write the
	// character's changes back into the plan know these aren't edits.
	applying = false;
	// The load that started last, so an older one that finishes after it doesn't win.
	private loadCount = 0;
	// Fires when a load has put a level in place, with its stats and spells.
	readonly loadEmitter = new TypedEvent<void>('LevelingPlanCharacterLoaded');

	constructor(sim: Sim, spec: Spec, makeRequest: (level: number) => RaidSimRequest | null) {
		this.sim = sim;
		this.makeRequest = makeRequest;
		this.player = new Player<Spec>(spec, sim);
		this.player.getRaidIndex = () => 0;
		this.player.getPartyIndex = () => 0;
	}

	// The level it is now, or 0 before the first load.
	get level(): number {
		return this.loaded;
	}

	// Makes it the plan's character at a level. weights are the stat weights it judges items
	// by, and rotation is one to play there in place of the plan's, like the one we edit.
	async load(level: number, options: { weights?: Stats; rotation?: APLRotation } = {}) {
		const request = this.makeRequest(level);
		const proto = request?.raid?.parties[0]?.players[0];
		if (!request || !proto) return;
		if (options.rotation) proto.rotation = options.rotation;
		const count = ++this.loadCount;
		const eventID = TypedEvent.nextEventID();
		this.applying = true;
		try {
			TypedEvent.freezeAllAndDo(() => {
				this.player.fromProto(eventID, proto);
				if (options.weights) this.player.setEpWeights(eventID, options.weights);
			});
		} finally {
			this.applying = false;
		}
		this.loaded = level;
		const result = await this.sim.computeStatsOfRequest(request);
		if (count != this.loadCount) return;
		const stats = result.raidStats?.parties[0]?.players[0];
		if (result.errorResult || !stats) {
			console.warn(`Failed to compute the leveling plan's character at level ${level}: ${result.errorResult}`);
			return;
		}
		this.player.setCurrentStats(eventID, stats);
		await this.player.updateMetadata();
		if (count != this.loadCount) return;
		// The rotation editor's spell lists listen to this, like for the settings' character.
		this.sim.unitMetadataEmitter.emit(eventID);
		this.loadEmitter.emit(eventID);
	}
}
