import { makeShamanImbueInput } from '../core/components/inputs/shaman_imbues';
import { makeRaidDamageHitsInput, makeRaidDamageHitsVariationInput } from '../core/components/inputs/shaman_raid_hits';
import { Spec } from '../core/proto/common.js';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

export const RaidDamageHitsInput = makeRaidDamageHitsInput<Spec.SpecEnhancementShaman>();

export const RaidDamageHitsVariationInput = makeRaidDamageHitsVariationInput<Spec.SpecEnhancementShaman>();

export const ShamanImbueInput = makeShamanImbueInput<Spec.SpecEnhancementShaman>();
