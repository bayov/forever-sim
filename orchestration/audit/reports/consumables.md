# Consumables report, pass 51 (master 1274e15b3)

Quiet: no findings.

- Master 224b8a4e0 -> 1274e15b3: moved buffs.go, assumptions_def.py, standards.md (git diff read). No change to presets.ts, level60 gear, mksettings, consumes.go, debuffs.go, sim/common, forever_enchants.go or the mage staff imbue.
- buffs.go: Trueshot Aura ranks now 15 / 22 / 30 / 40 / 50 (rank 5 = 50), replacing the 2026-10-15 ruling of 100; commit 73cedc9a9 cites user ruling 2026-10-28 and the launch-realm hotfix (foreverchanges.pro). Client rank 5 is 50 (70338 to 70658), so it agrees with the client and the new ruling. Presets and mksettings hard-code no Trueshot value (raid flag only), so nothing disagrees.
- assumptions_def.py: hunter_tsa_hm_rank is now one value (Rank 5 at 50); text matches buffs.go. The Low `hunter_tsa60` knob is gone and nothing else references it outside docs.
- standards.md: new rule 30 (Changes tab) is report-only; not a consumables matter.
- Client: wago latest is still 1.60.1.70658, already audited.
- scan.sh on master: banned items, Revelation outside shamans, non-synthetic trinkets, multi-crit-elixir all empty. Raid totems unchanged. MinorArmorReduction count 3 (third is the hunter pet Dust Cloud, as before).
- Class branches: not audited (unrelated history shows thousands of commits "ahead"), as in earlier passes.
- Not checked: the hunter Aspect of the Hawk hotfix value (outside this pass's files); `go test ./sim/hunter` failed to run here (no with_db build tag tried).

No open items.
