# Mechanics pass 75 (2026-11-12, master cb0d928fd6)

Three new findings, all from Blizzard's November 11 hotfixes (t/2403040 post 8, tier 1), which the repo
does not track yet. mechanics-20261104-1 (druid scales) is still open: no commit touched `tuning_70753.go`.

## Range
33db32b368..cb0d928fd6 (origin/master 71994fca6 adds no sim Go): one sim commit.

| commit | what | verdict |
|---|---|---|
| cb0d928fd | Talent tooltips. `sim/druid/talents.go` comment only; `sim/shaman/talents.go` and `proto/shaman.proto` rename improved_fire_nova to trial_by_fire (field 9 kept); `sim/priest/priest.go` TalentTreeSizes Shadow 18 -> 17 | No mechanic change. Checked `FillTalentsProto` (`sim/core/character.go:864-881`): the offset is added after each tree, so the last tree's size is never read. No leftover users of the old field name. |

Daily update / rulings: e89962738 restates standards 17 (Windfury Totem white hits only, "for now") and
keeps four Nov 9-10 readings until logs; user rulings, not findings. No forever-update merges.

## Get current
- Client: newest 1.60 build is still 1.60.1.70888 (`wago.py latest --probe 300`).
- Dev notes t/2360696: unchanged, post 10 (2026-11-04) newest.
- Hotfix thread t/2403040: post 6 (Nov 10) edited 2026-11-12 00:46:56 (text still carries the imbue + totem line
  the sim follows; edit not diffed), post 7 (Auction House changes deferred to Nov 17, no sim effect), post 8
  (Nov 11 hotfixes, new). Blue tracker beyond these threads not read this pass.

## November 11 hotfixes, sim-relevant lines

| line (tier 1) | sim | outcome |
|---|---|---|
| "With the fix to Windfury Totem activating from Holy Strike, we have changed Unsealed Fury" | WF totem procs on white MH swings only under Forever (`buffs.go:2168`) | mechanics-20261112-1, medium, user ruling |
| "Unsealed Fury no longer generates echoes from Windfury Totem attacks" | Unsealed Wrath stores/spends on every landed white swing incl. extra attacks (`unsealed_wrath.go:61`) | mechanics-20261112-2, medium (estimate), worker fix |
| "Judgement changed to now be considered a Ranged Attack ... no longer counts as a Melee attack ... uses the Paladin's Ranged Hit chance" | JoC / JoR are MeleeMHSpecial with melee outcomes (`soc.go:112,121`, `sor.go:113,118`) | mechanics-20261112-3, low (estimate), worker fix |
| Rend and Tear now works on Pounce Bleed | Pounce is not in any feral rotation | no DPS effect; worth a note for the druid worker |
| Volley no longer activates procs | Volley not in hunter presets | none |
| Flametongue Totem Rank 2 fixed | presets drop the raid Flametongue Totem | none |
| Bloody Claws of Attack +1 Fist; Earthen Silk Slippers spell damage | item data | already in TASKS.md (Bloody Claws); slippers for the item/data owner |

DPS sizes are estimates: no sims run (the fixes need code changes; measure when a worker lands them).

## Open items for the next pass
- Check whether a 70888+ hotfix push or a newer build carries the Judgement and Unsealed Wrath changes in SpellMisc /
  proc rows (the hotfix row fetch is still slow; wago.tools hotfixes search newest first).
- Diff the edited post 6 against the copy the daily update quoted, if any.
