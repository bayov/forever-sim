# Fairness pass 20 (2026-11-01, master c23ef04)

Result: findings (1 new, low). Full pass: gear check, full race sweep, new baseline.

## Why it ran

Since dd053d2, 22 watched files: hunter pets lose Pet Aggression and swing for the creature dps less their own
AP (560579d78, e9344df43, re-simmed in 11de3318a), Moonkin Aura / Leader of the Pack 3% crit with everything, not
stacking (f962f6c8f), the two-hand white crit doubling the whole swing's Rage (5b2eec6de), the 2026-10-31 update
(Arcane Blast 12% Int / 100% mana / 18 s, Righteous Fury threat), the October 30 hotfix (Frost mage values 98e819d39,
Improved Drains no longer raises Wrack aea78721b), Kang's Rend proc, standards 21 (crit Rage).

## What I checked

- Gear: `audit1b.py` crashed again on "Warlock (every spec)"; the last pass's fix (and its `distance.py`) never
  reached master: the run's commit doesn't take `tools/audit/fairness/`. Re-fixed (one line per warlock target).
  19 of 20 pass; Demonology fails 5a (crit 10.4 vs 14.1), the open fairness-20261031-1.
- Race sweep (5k, seed 1) against dd053d2. Every move has its commit:

| Builds | Move | Cause |
|---|---|---|
| BM, BM wolf / tallstrider | -80.2, -67.2, -67.2 | pet: no Pet Aggression, new pet swing |
| MM, Survival and their pet variants | -38.0 to -45.5 | same |
| Frostfire, Frost | +30.3, +21.0 | 98e819d39 Frost hotfix values |
| Arcane | -15.5 | f619fb648 Arcane Blast |
| Arms, Fury 2H | +11.4, +12.4 | 5b2eec6de two-hand crit Rage |
| Feral | +0.2 | f962f6c8f LotP gives spell crit too |
| everything else (Lone Wolf, warlocks, Fury DW, rogues, shamans, Ret, priests, Balance, Fire) | +0.0 | no leak |

- Moonkin / LotP: 59 requests carry both, the 12 warlock requests Moonkin only (demons take no player buffs,
  standards 14, and Moonkin already gave the warlock spell crit), the druids neither (their own aura). No spec gains
  or loses from the change in an uneven way.
- Hunter pet swing against warlock demons: the demons' weapons are fitted to logged pet sheets with their own AP
  (sim/warlock/pet.go:44-63), so they don't carry the double count the hunter fix removed. No asymmetry.
- Published against current: builds.json matches the shared 20k run except the mages (Frostfire +30.2, Frost +21.5,
  Arcane -17.3) and the two-hand warriors (+11 to +12): c23ef0440 held the report data while the October 30 hotfix's
  report pass re-sims them. The results auditor's pass 68 lists the same gaps (not a bug); I confirm them
  independently (sweep gaps within the 5k noise), not filed. Until that pass lands the published ranking is off:
  Frostfire would be 2nd (shown 4th behind BM and MM's old numbers in the metas; builds.json has the hunters right),
  Fury 2H passes Survival, Arms passes Destruction, Frost passes Arcane.
- The metas' published.dps are stale for the hunters by -38 to -87 now (results-20261101-1, open).
- Best race against report race: best everywhere except Affliction (Orc +1.3, fairness-20261101-2) and Arcane
  (Undead +2.8 at 20k, new). Fire Gnome +0.4 is a tie.

## Ranking now (report race, 5k)

Lone Wolf 1087.8 is 1st (BM fell to 6th, 1045.2); Frostfire 1075.6 2nd; MM 1063.7; Fire 1056.7; Demonology 1046.5;
BM; Fury DW 1041.6; Fury 2H 1039.6; Survival 1024.8; Affliction; Arms 1008.7; ... Frost 928.8 over Arcane 894.1;
bottom Subtlety 732.4. Top over bottom 1.49x (was 1.54x).

## Findings

- fairness-20261101-3 (low, orchestrator): Arcane's report race Gnome is 2.8 under Undead after the Arcane Blast
  hotfix; re-pick it on the hotfix's report pass.

## Open items carried

- fairness-20261101-1 (caster distance), -2 (Affliction race), 20261031-1 (Demonology set): no change on master.
- Ice Lance Low range, Flametongue Totem stacking, boss partial resist, Rage k / pet scaling: unchanged.
- For the run: commit `tools/audit/fairness/audit1b.py` with this pass, or the fix is lost a third time.

## Baseline

`ref/fairness/sweep_baseline.log` and `table_baseline.md` replaced with this sweep (c23ef04).
