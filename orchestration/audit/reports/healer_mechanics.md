# Healer mechanics: pass 7 (2026-11-10 UTC, master ca5e35c3f, client 1.60.1.70862)

Quiet pass. Due because 2 watched files changed on master, 91ad19535..ca5e35c3f, both in one commit
(bcc5d6cd4, "Maelstrom Weapon stacks on every melee hit, Flametongue rank 5 cut").

## Get current

- Newest client build on wago.tools: 1.60.1.70862, same as pass 6. No table re-check needed.
- `orchestration/update_ledger.md`: no change since pass 6 (last touched 2026-11-07).
- Dev notes: last checked 2026-11-10T06:46Z (`dev_notes_seen` unchanged, 10 posts).

## What changed

| File | Change | Healer reach | Verdict |
|---|---|---|---|
| `sim/shaman/talents.go` (`applyMaelstromWeapon`) | Under Forever every landed melee hit gives a stack (was 4 PPM a point); bolt trigger now 50% under Forever | The aura's `OnStacksChange` only edits `shaman.LightningBolt` cast time and cost; no heal is in its reach. No healer preset takes Maelstrom Weapon (only `ui/enhancement_shaman/presets.ts`). | Not a healer mechanic; nothing to check |
| `sim/shaman/flametongue_weapon.go` | Rank 5 cut of 7 points (log tier 1-ish, Forever server observation) | Melee weapon imbue damage only | Not a healer mechanic |

## Open findings

- 11 (Riptide cost): still open. `sim/shaman/healing.go:117,132` at ca5e35c3f still scales by
  `RiptideNotesManaScale` 0.6, giving 147/180/231 against the client's 150/180/230.
- 10 (untested heal paths): unchanged.

## New findings

None.
