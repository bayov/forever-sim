import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	Consumes,
	Flask,
	Food,
	UnitReference
} from '../core/proto/common.js';
import {
	FeralTankDruid_Options as DruidOptions,
	FeralTankDruid_Rotation as DruidRotation,
} from '../core/proto/druid.js';
import { SavedTalents } from '../core/proto/ui.js';
import DefaultApl from './apls/default.apl.json';
import BlankGear from './gear_sets/blank.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);

export const GearPresets = {
  [Phase.Phase1]: [
    GearBlank,
  ],
  [Phase.Phase2]: [
  ]
};

// TODO: Add Phase 2 preset and pull from map
export const DefaultGear = GearPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const DefaultRotation = DruidRotation.create({
	maulRageThreshold: 25,
	maintainDemoralizingRoar: true,
	lacerateTime: 8.0,
});

export const DefaultAPL = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const APLPresets = {
  [Phase.Phase1]: [
    DefaultAPL,
  ],
  [Phase.Phase2]: [
  ]
};

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Default talents, in the sim's per-tree digit format, on the trees of the launch realms' 2026-10-30
// hotfix (tools/forever_talents/client/druid.json): 0/41/10, the raid sim's Feral tank preset
// (docs/raid_sim.md). The bear talents: Ferocity, Furor, Feral Swiftness, Feral Instinct 2, Thick
// Hide, Savage Fury, Sharpened Claws, Predatory Strikes, Lacerate (a talent since the hotfix), Blood
// Frenzy, Leader of the Pack, Natural Instinct, Natural Reaction, Rend and Tear and Primal Bite
// (row 7 since the hotfix), and Heart of the Wild (Restoration since the hotfix) and Naturalist.
// Berserk left the tree, so its point went to Lacerate.
export const StandardTalents = {
	name: 'Standard',
	data: SavedTalents.create({
		talentsString: '-5522030202310212551-055',
	}),
};

export const TalentPresets = {
  [Phase.Phase1]: [
    StandardTalents,
  ],
  [Phase.Phase2]: [
  ]
};

// TODO: Add Phase 2 preset and pull from map
export const DefaultTalents = TalentPresets[Phase.Phase1][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = DruidOptions.create({
	innervateTarget: UnitReference.create(),
	startingRage: 20,
});

export const DefaultConsumes = Consumes.create({
	flask: Flask.FlaskUnknown,
	food: Food.FoodUnknown,
});
