# Findings: Forever client updates

Format and rules: `orchestration/audit/PASS.md`. Newest first.

## Open

### client_update-20261021-1: Champion of the Light may be back to 20/40/60% of Intellect (unannounced 10-16 hotfix)
- severity: medium (high if confirmed: it is spell power for every paladin preset that takes the talent; not measured)
- affects: paladin sub Holy, undead and twist presets (any build with Champion of the Light)
- found: 2026-10-21, master 599f9de
- evidence: foreverchanges.pro/patch-notes "Hotfix-2026-10-16" (third party, tier: unconfirmed): "Champion of the Light gives spell damage equal to 20/40/60% of your Intellect again, as before October 15 (10/20/30%)". Not in Blizzard's notes. Client 70532 (identical to 70450) can't confirm: the ranks come from trait curve 110350, and the Curve tables aren't downloaded. Sim: `sim/paladin/talents.go:49` uses `{0, .1, .2, .3}` (PAL-70425-05). Ledger row PAL-U-HF1.
- proposed fix: verify in game (the character sheet's spell power with 3/3 vs 0/3) or with the client's Curve/CurvePoint rows for curve 110350 in a hotfixed export. If confirmed, change the multipliers to `.2,.4,.6` (Forever-gated) and rerun the paladin presets.
- needs: user ruling (in-game check) then orchestrator fix

### client_update-20261021-2: Oct 20 mage Frost hotfix (Piercing Ice removed, Elemental Precision 3x2%), not in the client
- severity: low (needs verify; if real it moves Frost DPS by several percent, since Piercing Ice is +6% Frost damage at 3/3)
- affects: mage Frost and Frostfire presets; Missile Barrage builds (Frost Blast trigger)
- found: 2026-10-21, master 599f9de
- evidence: foreverchanges.pro "Hotfix-2026-10-20" (Piercing Ice removed; Elemental Precision 5x1% -> 3x2%; Ice Barrier +10% absorb and, by FC's wording only, +3% Frost damage) and "Hotfix-2026-10-16" (Frost Blast can trigger Missile Barrage). Not in Blizzard's notes. Client 70532 still has Piercing Ice node 130502 (def 135303, 3 ranks) and Elemental Precision entry 130507 MaxRanks 5, so it is server side or not real. Sim: `sim/mage/talents.go:170-191`. Already a master task (TASKS.md, commit 599f9de). Ledger rows MAGE-U-HF1..HF5.
- proposed fix: as in TASKS.md. Confirm through the wago hotfix table or in game, re-import the mage tree, drop Piercing Ice, make Elemental Precision 2%/rank over 3 ranks, re-optimize the Frost talents, and size it with dps_beforeafter.sh on mage__frost and mage__frostfire.
- needs: orchestrator fix (after verify)

## Resolved
