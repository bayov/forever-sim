# Round 70753 phase 5, hit trade: shaman (BLOCKED, nothing searched)

No talent search and no confirm run ran for any of the six shaman builds, so there are no new talents or DPS numbers.

## Why

`tools/rotopt` doesn't build in this session's checkout. The generated protos (`sim/core/proto/*.pb.go`, gitignored) are
missing, and the machine has no `protoc` (and no `buf`). `go build -tags=with_db ./tools/rotopt` fails with "no required
module provides package github.com/wowsims/classic/sim/core/proto". The session's permission policy refused fetching
`protoc` (GitHub release zip) and `buf` (`go install`). Unblock: put `protoc` on PATH and run
`make sim/core/proto/api.pb.go` (it writes only the gitignored `.pb.go` files), then run the commands below.

## The setup each build would run with

From the reqs/metas, `tools/rotopt/shaman_presets.sh` and `hit_sweep.md`:

| build | current talents | APL | gear set | pins (`-require`) | hit now vs cap |
|---|---|---|---|---|---|
| shaman__enhancement-level60 | `353003102-055030110051012251-02` | optimized | `ui/enhancement_shaman/gear_sets/level60_phase1_synthetic` | spiritWeapons | melee 10/8 (Tidal Focus 2/5) |
| shaman__enhancement-imp-gw | `353003102-055030110051212151-01` | optimized | same | spiritWeapons, improvedGhostWolf | melee 9/8 (Tidal Focus 1/5) |
| shaman__enhancement-water-shield | `052002-054030210051012151-053020001` | magma | same | spiritWeapons, waterShield | melee 13/8 (Tidal Focus 5/5) |
| shaman__enhancement-mana-tide | `03-055030310051012051-053051001001` | mana_tide | `..._burning_totem` | spiritWeapons, waterShield, manaTideTotem | melee 13/8 (Tidal Focus 5/5) |
| shaman__elemental-level60 | `5501331003001351-05500101-053` | level60 | `ui/elemental_shaman/gear_sets/level60_phase1_synthetic` | (none) | spell 16/16 (gear 10, 2pc 1, Tidal Focus 5/5) |
| shaman__elemental-level60-mana-tide | `5501331003001351-04-053150001001` | level60_mana_tide | same | manaTideTotem | spell 16/16 |

Commands (per build `<id>`, run from the repo root):

- Search: `go run --tags=with_db ./tools/rotopt -template <shaman_enh|shaman_ele> -settings tools/dps_report/req/<id>.json
  -level 60 -knobs <the preset's knobs> -durs 120,180,300 -combined -talent-search -require <pins> -iters 3000
  -hit-steps <gear set> -hit-cap <melee:8 -hit-span 1 | spell:16>`
- Confirm old, per d in 120, 180, 300: `... -eval -confirm-iters 20000 -durs d` (current talents, no `-hit-steps`).
- Confirm new: the same plus `-talents <pick> -hit-steps <gear set> -hit-cap <cap> -hit-step <step>`.

## Notes for whoever runs it

- The Enhancement gear sets have steps hit-1 to hit-5 and hit+1 to hit+2. Water Shield and Mana Tide at Tidal Focus 5/5
  need step -5 (the last) to land on 8, so a `-hit-span 1` search that wants -6 can't get it.
- rotopt has no worker flag; the search scores 4 candidates at once (`scoreParallel` in `search.go`), so `GOMAXPROCS=2`
  holds a run to 2 cores. On a 4-core machine that means two runs side by side.
- The Elemental builds sit exactly on the cap, so the open question there is only whether moving Tidal Focus points
  into other talents and wearing a hit+N step beats keeping them.
