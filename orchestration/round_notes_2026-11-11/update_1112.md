# Daily update 2026-11-12 landed (brief `briefs/update_1112.md`)

Master 58110f478 (data e805f787), report deploy 90ffa8ac. PR #128 closed: it landed through master.

## Changes and how each was confirmed

1. **PR #128 merged, then master's ranged Judgement kept.** Master's f1c07e13a (`rangedJudgement`) and PR #128's
   `judgementAsRangedAttack` agree on every point:
   - DefenseTypeRanged, ProcMaskRangedSpecial and CastTypeRanged;
   - one roll with No Active Defense (OutcomeMeleeSpecialNoBlockDodgeParry and OutcomeRangedHitAndCritNoBlock have the
     same body);
   - crit, weapon specializations and threat are the same.

   PR #128's helper was dropped. Kept from PR #128: the extra-attack source, `IsWindfuryTotemAttack()` and the Unsealed
   Wrath Windfury Totem gate. The merge left one gap, now fixed: a stored extra attack
   (`StoreExtraMHAttack`) lost its source, so a stored totem swing wasn't recognized (`extraSourcesStored`). No current
   preset reached it. Tests:
   - `TestWindfuryTotemAttackRecognized`: totem swings from a white swing, from Heroic Strike and from stored attacks
     are recognized;
   - `TestUnsealedWrathWindfuryTotemSwingFromAnyTrigger`: a totem swing stores no echo, whatever triggered it;
   - `TestJudgementsDontRollWindfuryTotem`: no judgement rolls the totem.

   Each test fails when its gate is reverted.
2. **Extra attacks never roll Windfury Totem** (parent's addition; the log audit of standards 17 found none do).
   `AutoAttacks.IsExtraAttack()` gates the roll in `WindfuryTotemBuffAura`, and `TestWindfuryTotemExtraAttacksDontRoll`
   covers Windfury, Hand of Justice, Sword Specialization and Thrash. Standards 17 now says so.
3. **Hunter Volley activates no procs** (Nov 11 hotfix). The only proc that could see Volley was the Undead Touch of
   the Grave, on its cast. The fix is the new `SpellFlagNoProcs`, which `ApplyProcTriggerCallback` skips;
   `TestForeverVolleyActivatesNoProcs` covers it. No build moves, since no preset casts Volley. Moving already cancels
   the channel. Rend and Tear on Pounce, Frost Nova ranks 2-4, Confounding Flash and Twilight Focus in Shadowform have
   no sim effect (TASKS item, ticked).
4. **Items revealed Nov 11-12.** `hidden.py` was rerun. Tooltips were compared with db.json:
   - fixed through `forever_seen_in_game.json`: Skullsmoke Pants (Int 15, 14 crit rating), Death Knight Sabatons
     (11 spell power), Gravetoll Gavel (healing 84);
   - Ebon Hilt of Marduk's Forever Corruption is now in `sim/common/item_effects.go`;
   - the other items match.

   No gear search ran: every changed item is an ilvl 59-61 blue, far below the synthetic Phase 1 sets, so none can
   enter a build. The 5 reveals stay flagged hidden until the client cache (2026-10-30 snapshot) is refreshed.
5. **Judgement assumptions** (`assumptions_def.py`): `pal_judge_weapon_spec`, `pal_judge_debuff_hit`, `pal_judge_crit`
   and `pal_unsealed_wrath_wf`, with the branch's choices kept. A post-Nov-12 Ret log settles them (TASKS item left open).
6. **Ret re-tune** (rotopt, Undead, 20k iterations, seeds 1 and 1000003, 120/180/300 s; kept only what won every cell):
   - Judgement of Command lost its melee-table edge, so Seal of Righteousness stays up and the twist returns to it
     after 3 Command copies.
   - On the Wisdom duties, Holy Strike goes ahead of Judgement; on Ret Shock, Holy Shock goes ahead of Holy Strike.
   - The Undead talents change: Reverence 2→1, Divine Intellect 2→3.
   - Prot paladin is unchanged: Righteousness loses at all three weights, and Max survival's Consecration at 80% mana
     is +3 ± 5.5, within noise.
7. **Publish.**
   - Hashes and goldens were rewritten (Combat Sinister Strike, P1 warrior, Enhancement). The druid feral-cat hashes
     were stale from c380bbaef, not a real change.
   - The full suite passes except `sim/web`, which needs `binary_dist`.
   - `TestLevelingSession` only needs `-tags=with_db`.
   - `changes.py check` lists nothing uncovered, and `stale.mjs` reports 0/68 builds stale.

## Moved builds, 180 s, live → new

| Build | Live | New | Cause |
|---|---|---|---|
| paladin__level_60 | 1004.2 | 983.6 | Unsealed Wrath totem gate, partly recovered by the Righteousness rotation |
| paladin__level_60-duty-crusader | 998.1 | 978.3 | same |
| paladin__level_60-duty-wisdom | 990.8 | 971.9 | same, plus Holy Strike first |
| paladin__level_60_sub_holy | 994.0 | 973.9 | same as level_60 |
| paladin__level_60_sub_holy-duty-crusader | 988.2 | 968.8 | same |
| paladin__level_60_sub_holy-duty-wisdom | 980.9 | 962.4 | same, plus Holy Strike first |
| paladin__level_60_sub_holy_undead | 1125.2 | 1113.8 | same, plus the talent change |
| paladin__level_60_shock_ret | 964.3 | 969.3 | Righteousness rotation, Holy Shock ahead of Holy Strike |
| paladin__level_60_shock_ret-duty-wisdom | 957.0 | 961.2 | Righteousness rotation |
| rogue__combat-swords | 946.4 | 943.8 | Sword Specialization swings no longer roll the totem (still ahead of Daggers, 936.3) |
| rogue__subtlety | 863.1 | 863.0 | extra attacks no longer roll the totem |

Defaults are unchanged: Ret level_60 for the paladin and Combat Swords for Combat.

## Tanking and Healing

- **Tanking:** only Protection Paladin moved, read as Reckoning and Hand of Justice swings no longer rolling the
  totem (not isolated).

  | Preset | Max threat | Balanced | Max survival |
  |---|---|---|---|
  | Max threat | 3266.7 → 3248.8 | 2820.3 → 2790.7 | 1481.5 → 1415.7 |
  | Balanced | 3234.3 → 3217.5 | 2837.5 → 2809.3 | 1647.0 → 1584.7 |
  | Max survival | 3112.5 → 3094.2 | 2786.8 → 2765.7 | 1810.7 → 1780.2 |

  Warrior, Bear and Warden are unchanged.
- **Healing:** nothing moved. All 14 bench runs were byte-identical to the published pack, so the mode was marked
  current with `--record-bench heal`.

## Deploy

90ffa8ac (https://forever-dps.pages.dev). The state is in `.agents/skills/dps-report/state.json`, the only state.json
in the repo.

## For a user ruling

1. The open parts of the Judgement hotfix are kept as the four assumptions above, until a post-Nov-12 Ret log with the
   totem up settles them.
2. Holy Strike first on the two Wisdom duties gains +0.4 to +2.5 DPS. It is positive in every cell, but round 70753
   left out gains under 1.
3. The Undead talent change gains +1.1 to +1.6 DPS, about 3 standard errors.
4. The `primarySeal` option still says Seal of Command, although the rotations now keep Righteousness up. It has no
   effect in the sim.
5. Prot paladin's drop comes from the extra-swing rule, which is the log audit's finding; it was not isolated per
   source.
