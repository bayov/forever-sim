# Results pass 105 (master 0404484fb)

Range 5536db4bc..0404484fb. Watched paths changed: sim/core (consumes, buffs, rage, target, stats), sim/common/item_effects.go,
sim/compsearch, sim/shaman (comments only), sim/warrior/stances.go, class goldens. No request, mksettings, proto-request or database change.

Shared run (68 builds, 20k, seed 1) vs metas: 60 of 68 moved more than 0.5 (largest Enhancement -25.5, paladins -9.5 to -17.7,
Arcane -14.6, Frostfire -13.5, warlocks -8 to -9.5, hunters -7 to -9, warriors +1.7 to +6.0).

Paired runs (same request, binaries 5536db4bc / f1942793d / 3275a60b1 / 0404484fb):

| Build | before | after f1942793d | after rage-taken |
|---|---|---|---|
| shaman enhancement-level60 | 990.6 | 965.2 | 965.2 |
| paladin level_60 | 956.8 | 941.0 | 941.0 |
| warlock demonology | 894.2 | 884.8 | 884.8 |
| hunter marksmanship | 1090.4 | 1082.6 | 1082.6 |
| mage arcane | 844.1 | 829.4 | 829.4 |
| priest shadow | 857.5 | 854.2 | 854.2 |
| warrior fury | 1009.0 | 1009.0 | 1014.9 |

Causes: f1942793d, Arcane Intellect and Divine Spirit overwrite the elixirs of their stat (EV-0185), so mana-limited builds lose 25
Int / Spirit (Enhancement OOM 0.7 -> 2.5 s); b4cab8498, Rage from hits taken follows the stance (warriors up). World buffs
(3275a60b1, 9701cff08), Armor Shatter slot (75dee7458), raid-scope check (5dbc01412) and compsearch changes move no request.
Mage Frost and Ret and others fall in the same elixir family. Not run: drift (mksettings and requests unchanged; only known
feral sapper line expected) and unread fields (no request change). The metas and builds.json need the full re-sim; see
results-20261104-2. No new finding.
