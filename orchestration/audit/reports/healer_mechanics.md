# healer_mechanics pass 34

Master f9fa79917 (last: 3e9748641, client 1.60.1.71139). Quiet: no new findings.

## Get current

- Client: newest build still 1.60.1.71139 (`wago.py latest`); checked last pass, nothing healing-related.
- Blizzard hotfix thread t/2403040: 16 posts, newest post 16 (2026-11-25 23:25 UTC) = `dev_notes_seen`. Nothing new.
- update_ledger.md: the 71139 section (Nov 23-25 hotfixes) adds nothing for healers beyond the open
  -20261126-18 (Genesis/Tranquility) and -19 (Oath of Light 60 min).

## What changed on the watched paths

| Commit | File | Healing? | Verdict |
|---|---|---|---|
| 7de8438a6 | sim/shaman/water_shield.go, items.go | yes: Totem of Sustaining (23200) gives Water Shield a 4th globe | match, tier 1 |
| 1223e764e | sim/core/spell_result.go | no: leveling-only hit-from-behind hook, guarded by `sim.Encounter.leveling` | out of scope |
| 5f50e8841 | (watched path touched by the re-rank) | no | - |

### Totem of Sustaining against the client (71139, tier 1)

| Field | Client | Sim |
|---|---|---|
| Item 23200 effect | ItemEffect 104176, TriggerType 1 (on equip), spell 28856 | `shaman.Ranged().ID == TotemOfSustaining` |
| 28856 | "Increased Water Shield Charges": Effect 6, aura 107 (add flat modifier), misc 4 (charges), +1 | globes 3 -> 4 (MaxStacks and OnGain stacks) |
| Reach | EffectSpellClassMask[1] = 0x20 | Water Shield 408510's SpellClassMask[1] = 0x20: matches |
| Water Shield 408510 | ProcCharges 3, ProcCategoryRecovery 3500, ProcTypeMask 0x362a8 | 3 (+1), 3.5 s, unchanged |

The user's logs (tools/logcheck/findings.md, "Water Shield gives 4 globes", ~700 shields since Nov 14) agree.
The 2% max mana per globe and the 3.5 s ICD are unchanged, so the totem is worth one extra 2% return per shield.

## Tests

No test covers the 4th globe (shields_test.go and healing_untested_test.go assert 3). Added as an update under
the open -20261104-10 (untested mechanics), not a new finding.

## Open findings (unchanged)

-19 Oath of Light 60 min, -18 Genesis on Tranquility, -17 Innervate on Mana Well, -11 Riptide cost,
-10 untested mechanics (Totem of Sustaining added).

Note for healer_gear: 23200 is in `tools/database/overrides.go` at Phase 5 and not in the shaman_restoration list,
though it drops in Scholomance (phase0_content.json "include": true); its scope, already noted in reports/healer_gear.md.
