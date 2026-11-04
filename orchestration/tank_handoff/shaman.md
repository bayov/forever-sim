# Warden shaman handoff (2026-11-04)

Branch `worktree-agent-a2b9a52e5c99bccc6`, stopped on the orchestrator's word (the container's 4 cores were
overloaded, load 35). The full write-up is in `docs/tank_shaman.md` (Presets, Numbers, Calls for the user, Why
the Warden takes more damage).

## Done

- **Task A (realign to client 70753 and the tree rework):** the Enhancement string remap (Improved Ghost Wolf to
  row 4), the tests (`sim/shaman/warden/forever_test.go`: talent indices, Mental Dexterity 50/100/150%,
  Stormstrike 6 s), the Warden golden regenerated (Spirit Weapons' 1.302 restored), commit 0d9606987.
- **Id move** to 999201-999216 (collided with sim/core's test mace 999_001); `TestForeverConsumeProcDamage`
  passes on the regenerated db; the Elemental and Enhancement goldens carry the new AllItems rows (additions and
  renames only). Commit 617ea865d.
- **Diagnosis** (doc section "Why the Warden takes more damage"): a real class gap, not gear or rotation. DTPS
  about 30% higher (no plate, no active block, no damage-reduction stance: +1,200 armor -8%, +45% block -7%, a
  stance -10% give the warrior's DTPS); 15% crushing blows against the warrior's 4%. Broodlord is a health
  threshold: Tauren only; build B's Improved Reincarnation 2 (4% health) takes his death floor from 4.3% to 0.7%.
- **Task B, step 6 (presets):** talent and knob searches done, races checked, requests, APLs, gear sets,
  `presets.ts` and `sim.ts` written; `npx tsc --noEmit -p .` has no error in `ui/warden_shaman` (the 18 errors it
  prints are in `ui/feral_tank_druid` and `ui/tank_warrior`: placeholders already on the base branch).
- **Step 7 (numbers):** all three presets on all seven encounters at 1,000 iterations, in the doc, next to the
  other tanks' current requests on this branch (pre-search).

## The presets

| Preset | Talents | Trinket | Knobs | Race |
|---|---|---|---|---|
| Max threat | `05050001-050032303551012201-050002` | 990015 (AP) | defaults + `waterTotem=1` | Tauren |
| Balanced | `05050002-050032113551012201-050002` | 990015 (AP) | defaults, `armorTrinket=0` | Tauren |
| Max survival | `05050002-050032113551012201-050002` | 999216 (armor) | defaults, `armorTrinket=1` | Tauren |

Scores on the six (rotopt objective, 3,000 iterations, paired seeds):
- Balanced: B 463.4 (31.9); the Max threat build 466.2 (33.7), a tie with half the time to out of mana, so B.
- Max threat: the Max threat build 2,526.3 (12.4) against B 2,481.8 (11.1); round 4's
  `0505-051031303551012201-051002` 2,512.7 (14.1).
- Max survival: B -5,409.2 (98.1); the Max threat build -5,532.9 (107.4); the search's mana tie-break
  `05050002-055032113501012201-050002` -5,459.9 (85.8).
- The realigned old build `35050013-0502333135510102`: Balanced -471.5 (1,391.9) at 1,500 iterations.

Tankbench means of the six (1,000 iterations): Max threat 2,552 / 547 / -5,469 (TPS 1,619, SafeDPS 1,777,
healing needed 1,066); Balanced 2,493 / 501 / -5,475 (1,589, 1,745, 1,055); Max survival 2,410 / 476 / -5,326
(1,548, 1,688, 1,017). Balanced before the update -1,170 (f9cc1ecaf), after the remap -618, now 501.

Races (build B, 1,000 iterations, MT / Bal / MS): Orc 2,441 / 123 / -6,830, Troll 2,423 / 117 / -6,802, Dwarf
2,342 / 10 / -6,987, Skyborne Windshaper 2,325 / -76 / -7,279; Tauren 2,501 / 515 / -5,312 (600 iterations).

## Left

- **The full test run** (`go test -tags=with_db -timeout 40m` over `./sim/...` without sim/practice and
  sim/web) was stopped part way. What it had shown: `TestRaidForeverReportHashesStable` fails on paladin and
  warrior DPS builds. That is the base branch's: `sim/testdata/report_hashes.txt` was last written at 661b3befe,
  before b57f5cc0a (downranking penalty) and 659cad198 (Blood Craze), and this branch changes nothing under
  sim/paladin, sim/warrior, sim/core or sim/testdata (`git diff 659cad198 HEAD`). The shaman and core tests
  passed before the id-move commit. Rerun the whole suite on a regenerated db
  (`go run ./tools/database/gen_db -outDir=assets -gen=db`, then the same with `-tags=with_db`; don't commit
  `assets/database/db.*`).
- Optional: the Max threat knob search was run on build B; rerunning it on the Max threat talents could move a
  knob (the talent change is small: Elemental Weapons 3 for Shamanistic Focus and a point of Elemental Warding).
- The other tanks' numbers in the doc are their pre-search requests; refresh them when their presets land.

## Commands

Every search ran from the worktree root on the six scoring encounters:

```
E=tools/encounters
ENC=$E/magmadar.json,$E/golemagg.json,$E/onyxia.json,$E/broodlord.json,$E/vaelastrasz.json,$E/chromaggus.json
go build -tags=with_db -o rotopt ./tools/rotopt
# talent search (B = 05050002-050032113551012201-050002); Max survival on the armor trinket request
./rotopt -template shaman_tank -settings tools/tankbench/req/shaman_warden_balanced.json -tank balanced \
  -tank-encounter $ENC -talents B -talent-search -require spiritWeapons -iters 600 -confirm-iters 3000 \
  -knobs armorTrinket=0
./rotopt ... -settings tools/tankbench/req/shaman_warden_max_threat.json -tank max-threat ... -knobs armorTrinket=0,waterTotem=1
./rotopt ... -settings tools/tankbench/req/shaman_warden_max_survival.json -tank max-survival ... -knobs armorTrinket=1
# knob search, writing the APL
./rotopt -template shaman_tank -settings <request> -tank <preset> -tank-encounter $ENC -talents <talents> \
  -iters 600 -confirm-iters 3000 -knobs armorTrinket=<0|1> -out ui/warden_shaman/apls -name level60_<preset>
# one candidate's score: the same with -eval -iters 3000 -confirm-iters 3000
# numbers
go build -tags=with_db -o tankbench ./tools/tankbench
./tankbench -encounter tools/encounters/<boss>.json -iters 1000 tools/tankbench/req/shaman_warden_<preset>.json
```

Earth Shock's threat multiplier at 1 (call (a)) was measured by setting `spell.ThreatMultiplier = 1` in
`sim/shaman/earth_shock.go` for a scratch tankbench build and reverting it.
