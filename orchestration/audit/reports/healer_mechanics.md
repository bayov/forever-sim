# Healer mechanics, pass 15 (2026-11-15)

Master 70ea588bb1 against last pass d630df2cb0. Client 1.60.1.70938 (unchanged). Result: quiet. Findings 10, 11, 14, 15 still open.

## Get current

| Source | Seen | Result |
|---|---|---|
| Newest client (`wago.py latest`) | 1.60.1.70938, same as pass 14 | no new rows to check |
| Blizzard notes | thread t/2403040 | not re-read (no new build) |

## What changed (d630df2cb0..70ea588bb1, watched paths)

| File | Change | Healer reach | Verdict |
|---|---|---|---|
| `sim/druid/ferocious_bite.go` | avoided Bite refunds 80% of its 35 Energy under Forever | none: cat-form Energy finisher | no healer effect |
| `sim/druid/forms.go` | Sharpened Claws now adds 3%/pt spell crit too, via `GetFormShiftStats` | only in cat/bear form (`forms.go:144`, `bear.go:98`); no heals are cast in form, Resto presets don't shift | no healer effect |
| `sim/druid/feral/*` | Feral tests and results | none | n/a |

No heal, healing talent or engine file changed. sim/core, shaman, paladin, priest and tools/healbench untouched.

## Open findings re-checked

- 15 (Wild Growth coefficient): `sim/druid/healing.go:174` still 0.033. Open.
- 11 (Riptide mana): `sim/shaman/healing.go:117` still scales by `RiptideNotesManaScale` 0.6. Open.
- 10, 14: no change on the watched paths. Open.
