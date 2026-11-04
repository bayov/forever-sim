# Protection paladin, round 5: talents, enchants, race and rotation on the round 4 consumables

Branch tank-fixes6 on master 9495b20. Score = mean of the six scoring bosses (Magmadar, Golemagg, Onyxia, Broodlord,
Vaelastrasz, Chromaggus) at the preset's own weight (Max threat 1/3, Balanced 1, Max survival 3); Ragnaros is benched,
not scored. Binaries: `go build -tags=with_db` of tools/tankbench, tools/rotopt and tools/audit/tank_rotations/abtest
into /tmp/claude-0/paladin_r5/. Summary in docs/tank_paladin.md ("Round 5"; call 14 resolved).

Result: Balanced takes Max threat's talents (`05005-25300102413001531-5023`) and Greater Agility boots (1887), keeping
Undead and Spirit of Zanza; Max survival takes the boots; Max threat takes `consecrationMana=0` (+1.0 ± 0.2 / +0.7 ±
0.2). Undead stays at every preset: of the paladin's races (Dwarf, Human, Undead; `paladinRaces` in
`ui/core/proto_utils/utils.ts`) it leads at every weight. The race screen also ran races a paladin can't be (Night Elf
and the rest); those runs are kept only as raw output and nothing uses them: `talents_mt_ne.txt`,
`knobs_max_threat_nightelf_illegal.txt`, `aplout/max_threat_nightelf_illegal/`, `ab/nightelf_illegal/` (the first
finals with a Night Elf Max threat), `var/nnightElf_*`, `var/ntauren_*`, `var/nr{Gnome,Orc,Troll,SkyborneHighOrder}_*`
and their `ab/` files.

## Files

| File | What |
|---|---|
| `req_before/` | the three requests as at 9495b20 (round 4) |
| `var/` | variants (`mkvar.py`); `base_*` = req_before, `n_*` = the new base (Balanced on Max threat's build, boots 1887 everywhere), `f_*` = final (n plus `consecrationMana=0` at Max threat, = `ncons0_max_threat`) = the committed requests |
| `ab/s<seed>_<A>_vs_<B>_<preset>_<iters>.txt` | paired abtest at the preset's own weight |
| `ab_summary.txt` | `summ.py ab/*.txt`: six-boss mean diff ± paired error (* = over 2x), the five without Broodlord, Vaelastrasz |
| `talents_{mt,bal,bal_from_mt,ms}.txt` | rotopt talent searches (Max threat; Balanced from its string; Balanced from Max threat's string on n; Max survival) |
| `knobs_<preset>.txt`, `aplout/` | rotopt knob searches on the final builds (APLs written equal the committed ones); `aplout/k_*` the knob variants A/B'd |
| `before_10k_seed{1,2}.txt`, `after_10k_seed{1,2}.txt` | `bench.py` (tankbench) on the before / after requests, 10,000 iterations |
| `rows_raw/`, `summary_paladin_rows.tsv` | the 24 summary.tsv rows (1,000 iterations, default seed), replaced in `orchestration/audit/ref/tank_bench/summary.tsv` |
| `gear_check.txt` | `python3 tools/audit/tank_gear/paladin.py` on the UI set |
| `tests.txt` | `go test -tags=with_db ./sim/paladin/protection/... ./tools/tankbench/ ./tools/tankscore/ ./tools/rotopt/` and `npx tsc --noEmit` |

Variant names: `mtbuild` Max threat's talents + boots 1887, `mttal` talents only, `boots1887`, `mtz` Max threat's whole
request (Scorpok); on n: `ncloakDef` 990704=1889, `nshArmor` 990714=848, `nshSpirit` 990714=1890, `nshSta`
990714=1886, `nwristSta` 990706=1886, `nboots904`, `nboots929`, `nnightElf`, `ntauren`, `nr<Race>`, `ntb` the talent
searches' mana tie-break strings, `nreseal30` / `ncons80` / `ncons0` knob APLs; `f<swap>_max_threat` enchant swaps
on the (illegal) Night Elf build, not used.

## Commands (from the repo root)

```
run.sh <seed> <iters> <A> <variant>[:<preset>]...    # paired abtest, var/<A>_<p>.json vs var/<variant>_<p>.json
search.sh talents|knobs <label> <request> <tank> <talents> [race]   # rotopt as round 3 (1,000 / 3,000, -freeze bulwarkAt)
phaseA.sh       # the four talent searches
enchants.sh     # the enchant and race screen on n, 3,000, seed 1
final.sh        # before vs f paired, 10,000, seeds 1 and 2; bench.py before/after
rows.sh         # the summary.tsv rows (then bench_summary.py)
```

Order of work: Balanced against Max threat's build (10,000, both seeds), the same at Max survival; the talent
searches; the enchant and race screen on n; the knob searches and their closest knobs at 10,000; the finals. (A first
pass took Night Elf for Max threat; it is not a paladin race, so Max threat was set back to Undead, its knob search
re-run on Undead, and its finals and bench rows redone.)
