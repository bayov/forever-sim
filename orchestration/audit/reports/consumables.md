# Consumables pass 69, master f331f0cff (since f23ea25d2)

Findings: consumables-20261104-1 (low): presets.ts still tick Salvation after 90e0bb01a dropped it from mksettings/requests.

Verified:
- forever_enchants.json: only the meta build moved (70755, same build as last pass); no enchant changed.
- assumptions_def.py: DPS-report text/ruling updates only (Trueshot +100, MW 20 ppm, Ice Lance, 2H spec), no consume values.
- mksettings main.go: Salvation removed (finding above); rest unchanged.
- presets.ts diffs (balance, enh, feral, ret, shadow): comments, talents (shadow Smite), Feral FF duty now takes the EZ-Thro Dark Bomb (same as DefaultConsumes; Dark Bomb is usable in form per standing note). No Consumes/raid changes otherwise.
- scan.sh: banned/Revelation/trinkets/crit-elixir sections empty. Counts: minor-armor effects 3 (ref says 2): the third is Dust Cloud (debuffs.go:1167, 30 sec, sharing the slot); no preset uses it, FF and CoR still share the slot. Scan expectation should be 3 ; totems as expected.
- wago latest: 1.60.1.70755, nothing newer. Class branches are stale (hundreds to thousands of commits behind), not audited.
No other open items.
