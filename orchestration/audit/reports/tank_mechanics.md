# Tank mechanics pass 33 (2026-11-25, master 0404484fb)

No new finding. Client still 1.60.1.71091. Range 5536db4bc..0404484fb on the watched paths:

| Commit | Tank reach | Verdict |
|---|---|---|
| b4cab8498 Rage from hits taken: stance modifier, x2 for spells, ticks, specials | Prot warrior, bear | Implemented as the new standards 21 says (tier 1.5 logs: our 38 Forever logs for the stance; wowsims/forever EV-0052, EV-0118 for the x2) |
| 3275a60b1 / 9701cff08 world buffs off in raid instances (Encounter.OutsideRaid) | none | No tankbench request carries a world buff |
| 5dbc01412 Windfury / Flametongue Totem raid scope (comments only in air/fire_totems.go) | none | Comments |

## b4cab8498, read in full (sim/core/rage.go, spell_result.go, stances.go)

- White hit: attacker damage (crit/crush applied) x StanceDamageTakenMultiplier only. Defensive 0.9, Berserker 1.1
  (stances.go apply and remove symmetrically); Shield Wall, Recklessness, T-set 0.9, Barkskin don't count. Blocks and
  armor cut none. Bear has no stance multiplier, so x1.
- Non-Physical hit or any tick: same D x2 (ForeverSpellDamageTakenRageMultiplier).
- Physical special (raid damage model melee events, ProcMaskMeleeMHSpecial via CalcAndDealDamage): PreAbsorbDamage
  (after armor, block, flat reductions, before absorbs; set in ApplyPostOutcomeDamageModifiers) / TargetDamageMultiplier
  x stance, then x2. Warriors and druids have no DynamicDamageTakenModifiers, so dividing out TargetDamageMultiplier
  removes all non-stance modifiers for tanks.
- Hand-built results (no AttackerDamage) still count Damage / ResistanceMultiplier, which keeps Shield Wall etc. in;
  pre-existing, and no tank event path builds one (raid_damage.go uses CalcAndDealDamage). Not a finding.
- Unmeasured and labelled so in the code: the x2 for druids (logs were warriors), and the specials rule outside
  Defensive Stance (To Verify, notes.md).
- Tests: TestForeverWarriorDamageTakenRage (white, blocked, stances, Fire, ticks, special partly and fully blocked),
  TestForeverShieldWallKeepsRage; `go test -tags=with_db ./sim/warrior/tank_warrior/ ./sim/warrior/dps_warrior/
  -run 'Forever|Rage|TestP1TankWarrior'` passes. Commit reports Protection Rage from hits -10%, TPS -0.8%.

## Open findings

tank_mechanics-20261124-4, -20261124-3, -20261117-1, -20261108-1, -20261107-1 (unchanged).
