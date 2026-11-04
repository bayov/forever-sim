# Protection paladin tank: handoff (2026-11-04, stopped mid step 6)

**Finished on branch `tank-paladin` (2026-11-04, its own cloud session):** step 6 (talent, knob and race searches on
the merged code), the three requests, APLs and UI presets, step 7, the doc's Presets, Numbers and Calls sections.
The report hashes move by the paladin's 12 zero-uptime auras only. The "Left" list below is the history.

Branch `worktree-agent-a7e349827790bf934`. The brief is `orchestration/brief_tanks.md`, "Restart after the 70753
update (T2, 2026-11-04)". The doc is `docs/tank_paladin.md`; its sections up to Rotation are current, and
`PRESETS_PLACEHOLDER` and `CALLS_PLACEHOLDER` are still to be written.

## Done

- **A, realignment** (commit bf0ea76b6): the downranking penalty on Seal of Fury's hit and Judgement of Fury
  (`DownrankLevel`); the UI rotations read Redoubt 20128 and cast Holy Shield rank 4 (TestProtection's settings runs
  had halved); the synthetic set's hit is sized to the cap (+29 Balanced); TestForeverProtTankScore counts tag 0
  only and allows Swift Judgement; a provisional `tools/tankbench/req/paladin_protection_balanced.json` (Human, old
  talents, the template's default rotation). The doc's realignment section has the audit table against 70753/70755
  and hotfixes 113483/113542, and what the update moved: Balanced, mean of six, **2085 before, 1133 after** (-8 from
  the paladin's rules, -944 from SafeDPS's 1-threat reference).
- **The orchestrator's questions, answered in the doc** (realignment section):
  - The ten report paladin hashes moved on 659cad198 only through 12 zero-uptime auras the tank code registers.
    Their DPS, TPS, casts, damage and threat are identical to master's.
  - The ten warrior hashes moved on 659cad198 on threat only, with DPS identical. That is the warrior work's
    (Blood Craze etc.).
  - TestRetribution's two tank-set lines were stale since merge bc2bfa009, after master's baseline move
    354.57 -> 330.13.
  - `TestForeverConsumeProcDamage` fails on 659cad198 because the warden's synthetic head 999001
    (`tools/database/forever_synthetic_items_shaman_tank.go:82`) collides with the test's
    `testProcWeaponID = 999_001`. That one is for the warden.
- **Merge of origin/master 8a8ba3bc1** (this commit), which brings the shared absorb pipeline and the Holy paladin
  healer:
  - Conflicts resolved: paladin.go keeps both SpellCode lists; the db is master's; report_hashes.txt is master's.
  - **The absorbs are moved onto `sim/core/absorb.go`**:
    - Seal of Fury's shield (its own helpful spell 20423 tag 1 books the healing).
    - Templar's Bulwark (the Bulwark spell books it).
    - `sim/paladin/absorb.go` is deleted.
    - Every paladin unit test keeps its numbers.
  - TestProtection's one new master line (AllItems-HolySynthetic, TPS 548.838 -> 548.902) is taken: master
    computed it with its rank 3 Holy Shield rotation and without the tank code.
  - The gear set `ui/protection_paladin/gear_sets/level60_phase1_synthetic.gear.json` is written, the bench request's
    equipment.

## Left

1. **report_hashes.txt after the merge**: the ten paladin lines (and the ten warrior lines) move again against
   master's file. Check that it is still aura-only before rewriting the paladin lines. The check is a temporary
   test that dumps each report request's DPS/TPS, per-action casts, damage and threat, and aura uptimes, run in a
   master extract and in the branch, then diffed. The scratch copy is `zz_dump_test.go` in the scratchpad's `palt/`
   folder and is not committed; it is about 40 lines on `raidReqNames`, `raidReq` and `runRaid` in sim/.
   Then `RAID_FOREVER_HASHES=... go test -tags=with_db ./sim/ -run 'TestRaidForeverReportHashes$'` and
   rewrite only the paladin lines.
2. Step 6, the rest (below).
3. The requests `tools/tankbench/req/paladin_protection_{max_threat,balanced,max_survival}.json`:
   - The synthetic set and its enchants (as in the provisional balanced one), Steadfast Libram 279247, Seal of
     Fury, Devotion Aura, Righteous Fury, Engineering/Blacksmithing.
   - Each preset's race, talents and APL (rotopt `-out`).
   - The balanced one replaces today's. Remove `paladin_protection_valor.json` if nothing reads it.
   - TestForeverProtTankScore reads the balanced one, so recheck its cast bounds.
4. The UI:
   - `ui/protection_paladin/apls/level60_{max_threat,balanced,max_survival}.apl.json`.
   - `presets.ts`: GearLevel60Synthetic, the three APLs, talents, and three `makePresetBuild` builds with `race`
     and `options: {primarySeal: Fury, righteousFury: true, aura: DevotionAura}`, modeled on
     `ui/tank_warrior/presets.ts`.
   - `sim.ts`: `builds: [...]`. The launch state stays as it is.
   - Then `npx tsc --noEmit -p .`.
5. Step 7: tankbench on all seven encounters at 1,000 iterations for the three presets, into the doc next to the
   other tanks' numbers. Fill PRESETS_PLACEHOLDER and CALLS_PLACEHOLDER. The calls to write up:
   - Seal of Fury effect 0's dummy.
   - The immunities left out of scoring.
   - Pursuit of Justice.
   - Hit to the cap out of Sta/Str.
   - Templar's Bulwark Magic (the client carries it, contrary to round_70753/paladin.md's "notes-only").
   - Undead over Human/Dwarf.
6. The full `go test -tags=with_db -timeout 40m` over ./sim/... without sim/practice and sim/web.

## Step 6 so far (rotopt, the six scoring encounters, 600 iterations a candidate, 3,000 to confirm)

All of it was run on the pre-merge build. The healer merge may move it, so re-eval the bests with `-eval` at
3,000 first.

- **Balanced, Human**: talent search from `-55302102513521531-052` (2114.2) finished at
  **`04-25300102513501531-5023`: 2432.0 (3,000 iterations)**.
  - The changes: Benediction 5 -> Deflection 5, Divine Intellect 5 -> 2, Conviction +3, Divine Strength +4,
    Stoicism -2, Sacred Duty -2.
  - The knob search at that string kept every template default: seal=1, resealWindow=0, holyShieldMana=0,
    holyShieldFirst=1, prepullHolyShield=1, holyStrikeMana=0, judgementMana=0, swiftJudgement=1,
    consecrationMana=60, consecrationFirst=0, bulwarkAt=10.
  - The closest alternatives: bulwarkAt=20 -38, seal=2 (Righteousness) -73, swiftJudgement=0 -105,
    consecrationFirst=1 -107.
- **Race at Balanced** (that string, defaults, `-eval` 3,000): Human 2432.0, Dwarf 2413.3, **Undead 2544.2**. The
  likely reason is Will of the Forsaken against Magmadar's and Onyxia's fears; check it per encounter.
- **Balanced, Undead**, talent search from that string:
  - Round 1: a mana tie-break, Anticipation -1 / Sanctified Judgement +1.
  - Round 2: **`03-25300102513501531-50231`: 2569.4** (600 iterations).
  - Stopped in round 3; not confirmed at 3,000.
- **Max survival, Undead**, from the Balanced string (1242.4):
  - Round 1: tie-break `04-25300102513501531-50131`.
  - Round 2: **`03-25300102513501531-50231`: 1316.4** (600 iterations). This is the same string as the Undead
    Balanced one.
  - Stopped in round 3.
- **Max threat, Undead**, from the Balanced string (2986.4):
  - Round 1: `05-25300102413501531-5023` 2989.0.
  - Round 2: **`05005-25300102413001531-5023`: 3036.7** (600 iterations; Improved Seals +5, Anticipation -5,
    Toughness -1).
  - Stopped in round 3.
- No knob search yet at max-threat or max-survival. No flask or enchant check yet.

The commands. `ev.sh` is `tools/rotopt -template paladin_tank -settings <settings> -tank <weight>
-tank-encounter tools/encounters/{magmadar,golemagg,onyxia,broodlord,vaelastrasz,chromaggus}.json -iters <n>
-confirm-iters <n> <args>`, and the settings are the provisional balanced request:

```
rotopt ... -tank balanced -iters 600 -talent-search -confirm-iters 3000 -talents 03-25300102513501531-50231 -race Undead
rotopt ... -tank max-survival -iters 600 -talent-search -confirm-iters 3000 -talents 03-25300102513501531-50231 -race Undead
rotopt ... -tank max-threat -iters 600 -talent-search -confirm-iters 3000 -talents 05005-25300102413001531-5023 -race Undead
rotopt ... -tank <w> -iters 600 -confirm-iters 3000 -talents <best> -race Undead -out <dir> -name level60_<preset>   # knob search, writes the APL
rotopt ... -tank <w> -iters 3000 -eval -talents <t> -race <Human|Dwarf|Undead>                                       # race check
```

A talent search takes about 7 minutes a round on a loaded 4-core machine, so give background runs the 2-hour
limit.
