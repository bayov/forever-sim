# tank_mechanics pass 14 (2026-11-13, master 97da1cca5c, client 1.60.1.70938)

Result: quiet. No new findings; the two open findings stay open.

**Getting current:** wago's latest moved 1.60.1.70888 -> 1.60.1.70938. Blizzard's forum (notes thread 2360696, hotfix
threads 2403040 and 2404092) answered "too many times, try again later" all pass (other auditors fetching at once), so
no notes newer than `dev_notes_seen` were read; the next pass should re-check them. update_ledger.md unchanged.

**Client 70888 -> 70938** (`tools/audit/client_update/table_summary.py`, then the new
`tools/audit/tank_mechanics/spell_changes.py` for the per-spell rows):

| Spell | Change | Tank effect |
|---|---|---|
| 1280353 Consecration (mark) | SpellAuraOptions ProcChance 100 -> 101, ProcTypeMask 8 -> 0 | none: the sim's mark (sim/paladin/consecration.go:28) carries only Consecrated Ground's Holy bonus, no proc |
| 8250 Flametongue Totem | EffectBasePoints 0 -> 8248 | none: no Warden preset places Flametongue |
| 1328825/1328829 Poised (rogue, druid) | aura 4 -> 42 triggering new 1330741/1330763 | none for bear (Energy only) |
| 408439.. Lightning Bolt, 408479.. Chain Lightning | base points / per level | not sim rows |
| Wrack, Volley, Arcane Shot, Frost Nova, Divine Spirit, pet Great Stamina, camp items | various | not tank |

**Changed since 71562a8b03 (watched paths):** sim/core/attack.go and sim/paladin/unsealed_wrath.go (33b13e310 extra
attack sources; a1a1a81bb no extra attack rolls Windfury Totem). Read in full.

## Engine: extra attacks and Windfury Totem

| Check | Sim | Verdict, tier |
|---|---|---|
| No extra attack rolls Windfury Totem (standards 17) | buffs.go `rollsOn` = `!IsExtraAttack() && windfuryTotemRollsOnForever` | match, tier 1 (user's Nov 6-11 log audit) |
| Source kept on a stored attack (Reckoning stores) | `extraSourcesStored`, falls back to 21919 | correct |
| Source kept on the delayed white landing (batch later) | DealDamage restores `mh.extraSource` | correct |
| Next-swing ability swung by an extra attack (Heroic Strike, Maul) | `swing()` sets extraSource before the replaced cast | does not roll, as standards 17 |
| Every extra attack caller has a non-zero source | Reckoning 20178, Sword Spec 12815/13964, HoJ, Thrash 21919, WF itself | correct |
| Reset clears sources | `AutoAttacks.reset` | correct |

Every tank preset (warrior, paladin, bear, Warden) takes raid Windfury Totem, so the prot paladin's Reckoning swings now
roll it no more. Unsealed Wrath: no tank preset takes it (pass 12).

## Tests

`go test -tags=with_db ./sim/core/ ./sim/paladin/... ./sim/warrior/... ./sim/druid/... ./sim/shaman/...`: all ok.
`TestWindfuryTotemExtraAttacksDontRoll` covers any non-zero source (Reckoning's path is the same; no Reckoning-named
case, not a finding). tankbench builds.

## Open findings

- 20261108-1 (Warden shock threat under the Stormstrike mark untested): still open, sim/shaman unchanged.
- 20261107-1 (stale "client disagrees" threat labels): still open, stances.go:110 and bear.go:57 unchanged.
