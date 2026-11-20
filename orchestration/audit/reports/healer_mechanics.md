# healer_mechanics pass 24 (2026-11-20, master e6eaa9470)

Range 813f2b1e9..e6eaa9470 (9 commits): no file under `sim/`, `docs/healing_*`, `tools/healbench` or the healer presets
changed (BiS lists, rotopt, TASKS, audit files only). Client 1.60.1.70959 -> 1.60.1.71037 (wago latest). Update ledger
unchanged; Blizzard's Nov 19 post (t/2403040 post 13, via TASKS.md) is Retribution only.

## Client 71037 diff (`python3 tools/sources/wago.py diff 1.60.1.70959 1.60.1.71037 --set mechanics`)

| Row | Change | Healer effect | Verdict |
|---|---|---|---|
| Vampiric Embrace 15286 SpellEffect 694156 (#0) | bp 20 -> 10; new #1 bp 100 ("You receive $s2% increased healing on yourself"); tooltip text | the Nov 16 hotfix, now in the client; sim `vampiricEmbraceHealing` gives 10% party, 20% caster | match, tier 1 |
| Aura of the Blue Dragon 23688 / 23684 ProcChance | 2 -> 3 / 100 -> 2 | already modeled from hotfix push 114044 (`sim/common/item_effects.go:3542-3546`) | match, tier 1 |
| Genesis 1223081 #0 (1210794) mask word 3 | +0x80 | only druid spell with that bit is Pounce Bleed: damage, no heal | no healer effect |
| Heroism 23682 (item heal proc), Lightning Strike, Windfury Totem categories, Reagent Economy, new NPC spells | various | no healer preset spell or engine path | none |
| SpellPower, SpellCastTimes, SpellDuration | no change | heal costs, cast times, HoT durations unchanged | - |

No heal rank, healing talent or set bonus row was added, removed or changed. Riptide (finding 11) still 150/180/230 in
the client.

## Tests

Not rerun: no sim code changed since pass 23, which ran the healer suites green.

## Findings

None new. Open: 10 (untested paths, low), 11 (Riptide cost 147/180/231 vs client 150/180/230, low), unchanged.
