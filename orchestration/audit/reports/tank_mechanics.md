# tank_mechanics pass 8 (2026-11-10, master ca5e35c3f, client 1.60.1.70862+hf)

Due on 2 watched shaman files changed (91ad19535..ca5e35c3f). Result: quiet.

**Getting current:** wago's latest is still 1.60.1.70862 (pass 6 covered it); dev notes checked by the loop today (state `checked_at`). No ledger entries in range.

**What changed:** one commit on the watched paths, bcc5d6cd4.

| Change | Tank effect | Verdict |
|---|---|---|
| Maelstrom Weapon (talents.go `applyMaelstromWeapon`): under Forever every landed melee hit gives a stack; Lightning Bolt 50% with Totem of the Storm | None: Enhancement talent 17 is 0 in all four Warden presets (ui/warden_shaman/presets.ts:99-111) | n/a for tanks (Enhancement's mechanics auditor owns it; logs, tier 1) |
| Flametongue Weapon rank 5 cut 7 (flametongue_weapon.go) | None: Warden presets and test imbue Rockbiter (presets.ts:130, warden_test.go:57) | n/a for tanks |

Other commits in range: 04bdcdb9e (enchants 41 Minor Health, 66 Minor Stamina to client values; no level 60 tank preset uses them), item folds, logcheck tooling. No change in sim/core, sim/warrior, sim/paladin, sim/druid, sim/shaman/warden or tools/tankbench; TestWardenShaman.results unchanged.

**Engine and tests:** untouched this range; pass 7's checks stand.

**Open findings:** 20261107-1 (threat labels) and 20261108-1 (Stormstrike-mark shock threat test) remain open; neither file touched.
