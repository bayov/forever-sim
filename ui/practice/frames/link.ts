// What the Blizzard frames read from the room. main.ts makes one RoomLink and hands it to every
// frame, so the frames don't reach into the room's insides.

import type * as THREE from 'three';
import type { SlotContent, Toolkit } from '../actions';
import type { Sheet, ToolkitSpell } from '../protocol';
import type { Tooltip } from '../tooltip';

export type LoadedInfo = {
	// The sim's enum names: "ClassShaman", "RaceOrc".
	class: string;
	race: string;
	level: number;
	name: string;
	// The RaidSimRequest as JSON (protojson): raid.parties[0].players[0] has equipment.items
	// (by ItemSlot order), talentsString, consumes and the rest.
	request: unknown;
};

export interface RoomLink {
	readonly loaded: LoadedInfo;
	readonly tooltip: Tooltip;
	// The character's name as the room shows it.
	playerName(): string;
	// The toolkit (castable spells and usable items, with spell names) once the session started.
	toolkit(): Toolkit | undefined;
	// The character sheet from the sim, once the session started.
	sheet(): Sheet | undefined;
	// whenReady resolves when the toolkit and sheet are there.
	whenReady(): Promise<void>;
	itemName(id: number): string | undefined;
	// enemyName is the name of the enemy a hit landed on (HitEvent.target), for the combat log.
	enemyName(index: number): string;
	// labelOf is how the room names a toolkit entry: "Lightning Bolt (Rank 4)" or the item's name.
	labelOf(s: ToolkitSpell): string;
	// pickUp puts a spell or item on the cursor, to drop on an action button (the spellbook).
	pickUp(content: SlotContent): void;
	// three gives the room's renderer and scene and your character's model (its root object),
	// for a frame that draws the model. It's undefined until the models load.
	three(): { renderer: THREE.WebGLRenderer; scene: THREE.Scene; player: THREE.Object3D } | undefined;
}
