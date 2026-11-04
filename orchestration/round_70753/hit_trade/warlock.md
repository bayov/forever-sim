# Round 70753 phase 5: warlock hit-trade talent search

Brief: `briefs/phase5_hit_trade_class.md`, class warlock, branch `phase5-hit-warlock` (from `phase5-hit-trade` aed188a).

## Result

| Build | Old talents | New talents | Gear step | Hit before / after (cap 16) | Old 120 / 180 / 300 | New 120 / 180 / 300 |
|---|---|---|---|---|---|---|
| warlock__demonology | 051-0305003131221001351-0550004 | **101-0305003131221001351-0550005003** | +5, `ui/warlock/gear_sets/level60_phase1_synthetic_demonology_hit+5.gear.json` | 16 / 16 | 1057.7(0.4) / 1053.9(0.3) / 1029.7(0.3) | **1121.4(0.5) / 1121.2(0.4) / 1098.4(0.3)** |
| warlock__demonology-duty-curse | same as base | same as base | +5 (same file) | 16 / 16 | 1041.6(0.4) / 1043.6(0.4) / 1022.0(0.3) | 1103.0(0.5) / 1109.3(0.4) / 1089.3(0.3) |
| warlock__demonology-duty-recklessness | same as base | same as base | +5 (same file) | 16 / 16 | 1040.2(0.4) / 1036.6(0.3) / 1014.9(0.3) | 1102.1(0.5) / 1102.1(0.4) / 1081.8(0.3) |
| warlock__destruction_fire | 252-0050003201-0050345003101351 | **242-0050003201-0050355003101351** | +1, `ui/warlock/gear_sets/level60_phase1_synthetic_hit+1.gear.json` | 16 / 16 | 1002.5(0.3) / 996.7(0.2) / 976.5(0.2) | 1002.8(0.3) / 997.8(0.2) / 977.8(0.2) |
| warlock__destruction_fire-duty-curse | same as base | same as base | +1 (same file) | 16 / 16 | 987.1(0.3) / 986.9(0.2) / 969.1(0.2) | 987.3(0.3) / 987.8(0.2) / 970.2(0.2) |
| warlock__destruction_fire-duty-recklessness | same as base | same as base | +1 (same file) | 16 / 16 | 984.4(0.3) / 980.7(0.2) / 962.5(0.2) | 984.8(0.3) / 981.5(0.2) / 964.1(0.2) |
| warlock__destruction (default) | 255000000351--055030500310035 | unchanged | +0 (the set) | 16 / 16 | 1063.6(0.5) / 1063.4(0.4) / 1045.9(0.3) | = old |
| warlock__destruction-duty-curse | unchanged | unchanged | +0 | 16 / 16 | 1045.8(0.5) / 1052.6(0.4) / 1037.2(0.3) | = old |
| warlock__destruction-duty-recklessness | unchanged | unchanged | +0 | 16 / 16 | 1043.7(0.5) / 1046.4(0.4) / 1030.6(0.3) | = old |
| warlock__affliction | 05500021035001351-0050003201-005004 | unchanged | +0 | 16 / 16 | 1024.8(0.3) / 1034.1(0.3) / 1030.8(0.2) | = old |
| warlock__affliction-duty-curse | unchanged | unchanged | +0 | 16 / 16 | 1010.4(0.3) / 1025.1(0.3) / 1024.6(0.2) | = old |
| warlock__affliction-duty-recklessness | unchanged | unchanged | +0 | 16 / 16 | 1009.1(0.3) / 1019.9(0.3) / 1019.2(0.2) | = old |
| warlock__affliction_shadow_bolt | 2550102003520005--0550305003 | unchanged | +0 | 16 / 16 | 1015.5(0.5) / 1015.6(0.4) / 998.0(0.3) | = old |

DPS: rotopt `-eval`, 20,000 iterations per length, standard error in brackets, each request's own seed (1). The 180 s
column is the request's 180 ± 20 s. "= old" means the search kept the build's talents on the +0 step, which is the
set itself, so the new run is the old run (not rerun).

Hit: every build was Suppression 5/5 + 2pc set bonus 1 + 10 gear = 16. Demonology now has Suppression 0 + set 1 + 10
gear + 5 from the step = 16. Fire Destruction has Suppression 4 + set 1 + 10 + 1 = 16.

## Talent changes in words

- **Demonology**: Suppression 5 → 0 (the 5% hit now comes from gear, step +5), and the 5 points go to Improved Life Tap
  +1, Ruin 4 → 5 and Agonizing Flames 0 → 3 (+10% to Destruction spells, Shadow Bolt included:
  `sim/warlock/talents.go:962`). Soul Link (the pin) kept. Search path: suppression-1 ruin+1 (1057.9), suppression-3
  agonizingFlames+3 (1115.1), then improvedLifeTap+1 suppression-1 on the mana tie-break (1114.3), means of 120/180/300
  at 2000 iterations.
- **Fire Destruction**: Suppression 5 → 4 (1% from gear, step +1), Aftermath 4 → 5. Incinerate and Conflagrate (pins) kept.
  It's a small gain: +0.3 / +1.1 / +1.3.
- Destruction, Affliction and Shadow Bolt Affliction: the search tried and kept Suppression 5/5 on the set (no round
  improved on the start).

## Odd things to note

- **Most of Demonology's gain isn't from the gear trade.** The new talents on the plain set (step 0, 11% hit) already
  get 1088.0(0.4) at 180 s against the old 1053.9, which is +34. The +5 step adds another +33 (1121.2). Round 70753's
  searches ran with `-require suppression` (`tools/rotopt/warlock_presets.sh` header), and that kept Agonizing Flames
  out of reach. Dropping the pin was enough to find it. At 180 s, the stepwise numbers were 041-…-0550005 on +1 at
  1064.0 and 101-…-0550005003 on 0 at 1088.0, against 1121.2 on +5.
- **Demonology (1121.2) is now ahead of Destruction (1063.4)**, which is the class default (`isDefault`). It's the
  phase 5 session's call whether to move the default.
- Demonology's new build goes out of mana sooner: rotopt's tto field drops from 193 s to 156 s at 180 s, at 636 mana
  left. It's still the best DPS over all three lengths.
- Warlock steps only go up (+1 to +5): the sets are at the cap with Suppression 5/5, and no warlock race has a hit
  racial. Demonology used the last step (+5 = all of Suppression), so the tool covered the whole range. Nothing wanted
  a step that doesn't exist.
- Duty variants kept their base build's talents, as the brief says (none were searched on their own in round 70753). Their
  knobs are the base build's plus `curse=1` (Curse of the Elements) or `curse=2` (Recklessness), as in
  `warlock_presets.sh`.
- Pins (`-require`): demonology `soulLink`; destruction `conflagrate,improvedShadowBolt`; affliction
  `wrack,demonicSacrifice`; affliction_shadow_bolt `improvedShadowBolt`; destruction_fire `incinerate,conflagrate`.
  Suppression was left free.
- The rotations (knobs) weren't searched again on the new talents. Demonology's knobs were tuned for the old build, so
  a knob search on 101-…-0550005003 might add a little.

## Commands

Run from the repo root. `K_<build>` is the build's knob string from `tools/rotopt/warlock_presets.sh`
(`$base,...` for that build), with `,curse=1` or `,curse=2` added for the duty variants.

Talent search (one per base build):

```
go run --tags=with_db ./tools/rotopt -template warlock -settings tools/dps_report/req/warlock__<build>.json \
    -combined -durs 120,180,300 -talent-search -iters 2000 -knobs "$K_<build>" -require <pins> \
    -hit-steps ui/warlock/gear_sets/<set> -hit-cap spell:16
```

`<set>` is `level60_phase1_synthetic_demonology` for Demonology, `level60_phase1_synthetic` for the rest. For example,
Demonology:

```
go run --tags=with_db ./tools/rotopt -template warlock -settings tools/dps_report/req/warlock__demonology.json -combined -durs 120,180,300 -talent-search -iters 2000 -knobs "bane=2,corruption=1,immolate=1,lifeTapMana=5,shadowburn=0,soulFire=0,endInstants=1,berserking=1,doomAlign=1,filler=1,conflagrate=0,siphonLife=0,soulLink=1,prepullLifeTap=1,dotLead=1.5,prepullAt=2.5,corruptionMinTime=15,immolateMinTime=15" -require soulLink -hit-steps ui/warlock/gear_sets/level60_phase1_synthetic_demonology -hit-cap spell:16
```

Confirms (per request id, old then new):

```
# old
go run --tags=with_db ./tools/rotopt -template warlock -settings tools/dps_report/req/<id>.json -durs 120,180,300 -eval -confirm-iters 20000 -knobs "$K"
# new (demonology*: talents 101-0305003131221001351-0550005003, step +5, set level60_phase1_synthetic_demonology;
#      destruction_fire*: talents 242-0050003201-0050355003101351, step +1, set level60_phase1_synthetic)
go run --tags=with_db ./tools/rotopt -template warlock -settings tools/dps_report/req/<id>.json -durs 120,180,300 -eval -confirm-iters 20000 -knobs "$K" \
    -talents <new> -hit-steps ui/warlock/gear_sets/<set> -hit-cap spell:16 -hit-step <step>
```

The Demonology breakdown (180 s, 20,000): the same new-run command with `-durs 180`, using `-talents
101-0305003131221001351-0550005003 -hit-step 0` and `-talents 041-0305003131221001351-0550005 -hit-step +1`.

(Building rotopt in a fresh container first needs the Go protos: `make sim/core/proto/api.pb.go` with protoc and
protoc-gen-go v1.36.6 on PATH.)
