# Consumables pass 63 (master fa865a3b8)

Findings: none new. Open: consumables-20261101-1 (low, hunter Sapper) unchanged.

Verified
- Moved since 5a6e9c6a6: assumptions_def.py (PAL_ALL drops the removed shock presets), spec_ret.go (ret_shock spec removed, same setup as ret), mage presets.ts (Frostfire default APL/talents, no consumes change), retribution_paladin presets.ts (ShockConsumes and the shock presets removed; ShockConsumes equalled DefaultConsumes, no leftover references).
- scan.sh: all banned/Revelation/trinket/crit-elixir sections empty; counts as before (minor-armor count prints 3, ref expects 2, known, debuffs.go unchanged).
- wago latest: 1.60.1.70707, no newer build. Class branches are stale copies, not audited.

No open items beyond the one finding above.
