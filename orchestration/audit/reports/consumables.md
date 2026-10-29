# Consumables pass 53 report

Master dcd927490 (last audited 52ec62e99). Client 70658 -> 70685.

## Findings
None.

## Verified
- Changed files: sim/core/buffs.go, sim/core/totem_weapon_buffs.go. The diff is a metrics/label refactor of the
  Windfury Totem aura: new WindfuryTotemSpellId {8512, 10613, 10614} (all "Windfury Totem" in the 70685 client),
  the proc buff is now labelled "Windfury Attack" and the aura reports under the totem spell id. Values are
  unchanged (AP 95/179/246, 1 sec buff under Forever, main hand only, blocked by Windfury Weapon on the MH).
- builddiff.py 70658 -> 70685: 0 consumable spell changes; 53 Forever enchants checked, 0 changed.
- scan.sh on master: banned, Revelation, trinkets and crit-elixir sections empty. Counts match the baseline
  (raid totems as ruled; Salvation on; Sunder on; MobTypeUnknown default).
- Class branches: the changes.sh "ahead" lists are stale copies of shared files (gotcha 1); not audited.

## Nit (no finding)
scan.sh's "minor-armor exclusive effects (expect 2)" now reads 3: the third is Dust Cloud (ForeverDustCloudAura,
hunter pet), already in debuffs.go, not in this window. The scan's expectation text is stale.

No open items.
