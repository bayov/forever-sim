# tank_rotations pass 31

Master 85691655ab (2026-11-25), against pass 30's master c3bbe33eab and the baseline of 0404484fb6.
Two new findings: 20261125-1 (high) and 20261125-2 (low).

## What changed

The wowsims/forever evidence batches. 2f42f4a44 is the tank batch. Warrior: Rage-gain threat is unscaled and
there's none before the pull, Heroic Strike 2.5 x level flat, Thunder Clap threat 2.5 -> 1, Battle Shout has no
flat threat, Shield Slam threat is 2 x block value + 255, Shield Block's last charge lasts one more batch, and the
damage-taken Rage armor table changed. Bear: Lacerate threatens 10/3 of its damage, Primal Bite 3x (was 2x),
Faerie Fire 2 x level, Enrage's armor cut reaches Agility armor. Damage taken: the flat physical reduction is capped
at half the hit. The proc, caster and item batches are b663145ad, efeb06b95, 295206ddf and 22d2cf952 (paladin SoW
and talents, core spell resistances and debuffs). No tank request, UI apl, encounter, tankbench or tankscore file
changed. rotopt mksettings changed for enh/feral/hunter only.

## Baseline (1,000 iterations, default seed; `ref/tank_bench/summary.tsv` rewritten)

| request (scoring mean) | pass 30 | now | change | SafeDPS | TPS | cause |
|---|---|---|---|---|---|---|
| warrior_protection_balanced | 3450.5 | 3241.2 | -209.3 | -106 | -77 | 2f42f4a44 threat rules |
| warrior_protection_max_threat | 4233.0 | 4102.7 | -130.3 | -67 | -50 | same |
| warrior_protection_max_survival | 1820.0 | 1617.0 | -203.0 | -107 | -74 | same |
| warrior_protection (legacy) | -49.0 | -399.2 | -350.2 | -178 | -86 | same |
| druid_bear_balanced | 2152.8 | 3338.3 | +1185.5 | +595 | +745 | Lacerate 10/3, Primal Bite 3x |
| druid_bear_max_threat | 3133.7 | 4541.7 | +1408.0 | +705 | +819 | same |
| druid_bear_max_survival | 431.8 | 1260.7 | +828.9 | +421 | +479 | same |
| paladin_protection_balanced | 2809.0 | 2857.7 | +48.7 | +14 | +14 | damage taken (DTPS -3, healing needed -6.5) and threat |
| paladin_protection_max_threat | 3243.7 | 3280.3 | +36.6 | +15 | +15 | same |
| paladin_protection_max_survival | 1787.2 | 1858.0 | +70.8 | +12 | +11 | same (Onyxia +189 ± 45) |
| shaman_warden_balanced / max_threat / max_survival | 1238.0 / 2599.7 / -2283.2 | 1245.7 / 2608.2 / -2234.0 | +8 / +9 / +49 | ~0 | ~+1 | damage taken; within noise per boss |

Every boss moves the same way for each class, so these are rule changes, not rotation changes. All were intended,
evidence-backed changes (each commit cites its EV ids), so per PASS.md they are not findings here. The outcome to
note: the bear (4541.7 Max threat, 3338.3 Balanced) now ranks above the warrior (4102.7, 3241.2) at Max threat and
Balanced. Before this, the warrior led by about 1,100. The Lacerate 10/3 rule settles that ranking. If anyone
doubts EV-0034/0037, that belongs to tank_mechanics.

## Rotation lines (re-checked under the new threat rules)

Paired abtest, scoring mean, 3,000 iterations, seeds 1 and 2, B minus preset, at the preset's weight:

| change | seed 1 | seed 2 | verdict |
|---|---|---|---|
| bear Max threat: no Demoralizing Roar | +102.2 ± 3.7 | +103.3 ± 3.8 | **wins (finding 20261125-1)** |
| bear Balanced: no Demoralizing Roar | +59.6 ± 8.6 | +39.6 ± 8.9 | **wins (finding 20261125-1)** |
| bear Max survival: no Demoralizing Roar | -44.2 ± 22.7 | -88.3 ± 25.7 | loses, keep |
| warrior Balanced: no Thunder Clap | +11.4 ± 10.6 | +24.4 ± 9.9 | 10k seed 3: +16.1 ± 5.6 (finding 20261125-2; Chromaggus -176) |
| warrior Max survival: no Thunder Clap | -289.7 ± 16.7 | -294.7 ± 16.2 | loses, keep |
| warrior Balanced: Shield Slam above Revenge | +0.5 ± 3.3 | -12.8 ± 3.0 | loses, keep |
| warrior Max threat: Shield Slam above Revenge | -1.5 ± 1.8 | -6.9 ± 1.7 | loses, keep |
| warrior Max survival: Shield Slam above Revenge | -26.3 ± 10.4 | -10.5 ± 9.3 | loses, keep |
| warrior Balanced Heroic Strike 40 -> 30 | -2.0 ± 5.9 | +8.0 ± 6.0 | noise, keep |
| warrior Max threat Heroic Strike 30 -> 20 | -0.8 ± 1.9 | +1.6 ± 2.1 | noise, keep |
| bear Balanced Maul 20 -> 10 / 30 | -0.7 / +0.6 | -0.9 / -3.1 | noise, keep |
| bear Max threat Maul 20 -> 10 / 30 | -0.8 / -0.6 | -2.0 / -0.7 | noise or loses, keep |
| bear Max survival Maul 40 -> 30 / 50 | +5.8 / -88.8 | -1.3 / -139.8 | 30 noise, 50 loses |
| bear Max survival Swipe 80 -> 70 / 90 | -142.6 / -3.4 | -128.1 / -17.7 | keep 80 |

Variants are in `$AUDIT_WORK/tank_rotations/v/`, made with `tools/audit/tank_rotations/mk_warrior.py <A> <B> del|move|setcond ...`.
Abtest is built from `tools/audit/tank_rotations/abtest`.

## Open findings

- 20261125-1 (new, high): bear Max threat and Balanced should drop Roar. rotopt's `roarRefresh` 0 already means
  "never", so a re-search should find it.
- 20261125-2 (new, low): warrior Balanced's Thunder Clap is about even after its threat cut. The result depends on
  the encounter mix.
- 20261122-1 (paladin Balanced Consecration floor): still open, not re-tested.
- 20261104-11, 20261104-3 (Broodlord-dependent trinket hold and Bulwark timer): user calls, untouched.

## Left for a later pass

- The Windfury main-hand timing item in `passes/tank_rotations.md` (user request 2026-11-15) is still not done.
- Paladin and Warden lines weren't re-tried. Their rules barely moved (+1 to +2.5%, damage taken).
- The setup needs `protoc -I=./proto --go_out=./sim/core ./proto/*.proto` before the first `go build` in a fresh
  session, because the generated proto isn't checked in.
