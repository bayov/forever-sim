# Protection paladin, round 3: raw search, A/B and bench output

Branch tank-paladin-3 on master a096b3e (unmultiplied mana-gain threat, tank_mechanics-20261104-12; disjoint
tank-tool seeds, tank_rotations-20261104-13). Score = mean of the six scoring bosses (Magmadar, Golemagg, Onyxia,
Broodlord, Vaelastrasz, Chromaggus) at the preset's weight (Max threat 1/3, Balanced 1, Max survival 3); Ragnaros
is benched, not scored. Binaries: `go build -tags=with_db` of tools/tankbench, tools/audit/tank_rotations/abtest
and tools/rotopt.

## Files

| File | What |
|---|---|
| `before_10k_seed1.txt` | the three requests as at a096b3e (shield 1886), 10,000 iterations, seed 1 |
| `after_10k_seed1.txt` | the three final requests, 10,000 iterations, seed 1 |
| `matrix_3k_seed1.txt` | the final requests, 3,000 iterations, seed 1: each preset at each weight (the 3x3 matrix) |
| `talents_<preset>.txt` | rotopt talent search, on the new enchants (wrist 1248665, shield 863; Max threat also boots 1887) |
| `knobs_<preset>.txt` | rotopt knob search (Bulwark frozen at 10 s), final talents and gear; wrote the same APLs as committed |
| `races_{mt,bal,ms}.txt` | rotopt `-eval` at 3,000 for each race, final build |
| `ab/s<seed>_<variant>_<preset>.txt` | paired abtest, 3,000 iterations, A = the request as at a096b3e |
| `ab/s<seed>_onN_<variant>_<preset>.txt` | paired abtest, 3,000, A = N (wrist 1248665 + shield 863, boots 929) |
| `ab/s<seed>_onF_<variant>_<preset>.txt` | paired abtest, 3,000, A = F (the final request of that preset) |
| `ab_summary.txt` | `summ.py ab/*.txt`: six-boss mean diff ± paired error (* = over 2x), the five without Broodlord, Vaelastrasz |
| `gear_check.txt` | `python3 tools/audit/tank_gear/paladin.py` on the final UI set |
| `bench.py`, `mkvar.py`, `summ.py` | the scripts |

Variants (`mkvar.py <base> <out> item=enchant ...`): `wristSD` 990706=1248665 (Superior Deflection);
`cloakDef` 990704=1889 (Superior Defense, against Agility 1219587); `bootsAgi` 990710=1887 (Greater Agility),
`bootsAgi5` 990710=904 (Agility +5), against Greater Stamina 929; shield against 1886: `shBlock` 990714=863
(Lesser Block +2%), `shArmor` 990714=848 (Lesser Protection +30 armor), `shSpirit` 990714=1890 (Superior
Spirit +9); `combo` = wristSD+shBlock+bootsAgi, `comboNoBoots` = wristSD+shBlock (= N); `Nboots*`/`Ncloak*` the same
swaps on N; `Ntal02`/`Ntal03` talents 02-25300102513501531-50232 / 03-25300102513501531-50231 (the talent
search's mana tie-breaks); `FnightElf` race=RaceNightElf; `Fbulwark13` the template's APL at bulwarkAt=13
(`rotopt -eval -knobs bulwarkAt=13 -out`); `FmtBuild` the final Max threat request scored at the other weight;
`FmtTal` Max threat's talents on the Balanced build. Not tried on the shield: Frost Resistance (926; standards 36:
no resistances), Off-Hand Wisdom (1219579: held-in-off-hand only), Law of Nature (7603: SoD), Stamina +5 (852:
below 1886).

## Commands (from the repo root)

```
python3 orchestration/tank_handoff/paladin/round3/bench.py <tankbench> 10000 1 tools/tankbench/req/paladin_protection_{max_threat,balanced,max_survival}.json
python3 orchestration/tank_handoff/paladin/round3/bench.py <tankbench> 3000 1 tools/tankbench/req/paladin_protection_{max_threat,balanced,max_survival}.json
abtest -iters 3000 -seed <1|2> A_<preset>.json <variant>_<preset>.json tools/encounters/{magmadar,golemagg,onyxia,broodlord,vaelastrasz,chromaggus,ragnaros}.json
# ro = rotopt -template paladin_tank -settings <request> -tank <max-threat|balanced|max-survival>
#      -tank-encounter tools/encounters/{magmadar,golemagg,onyxia,broodlord,vaelastrasz,chromaggus}.json (comma separated)
ro -iters 1000 -confirm-iters 3000 -freeze bulwarkAt -talent-search -talents <preset's string> -race Undead   # talents_*
ro -iters 1000 -confirm-iters 3000 -freeze bulwarkAt -out <dir> -name level60_<preset>                     # knobs_*
ro -iters 3000 -eval -race <race>                                                                          # races_*
```

Order of work: the enchant A/Bs on the round-2 build (s1, s2), the talent searches on the new enchants (an earlier
start on the old shield was stopped when Lesser Block won), the knob search, the races, then the A/Bs on the
final build (onF).
