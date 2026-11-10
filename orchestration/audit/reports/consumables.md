# Consumables pass 77, master ca5e35c3f

Moved: forever_enchants.go (enchants 41, 66 re-valued), assumptions_def.py (sham_mw_ppm now one value, every hit).
No newer wago build (70862). Class branches show no useful ahead state.

Findings: consumables-20261110-1 (low, enchant 41 name). Open from before: 20261107-1 (Natural flasks halved), 20261105-1.

Verified: ench.py 41 = Stamina +2, 66 = Stamina +3, matching forever_enchants.go. sham_mw_ppm change agrees with notes.md (100% chance, spell 408498). scan.sh clean above counts.
Count note: minor-armor exclusive effects is 3 (was 2): Forever Dust Cloud aura joins Faerie Fire and CoR; not checked whether a preset uses it.
