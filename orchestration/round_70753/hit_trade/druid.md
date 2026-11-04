# Round 70753 phase 5: druid hit-trade talent searches

Talent searches from each build's current talents, with the hit-trade gear steps (`tools/rotopt` `-hit-steps`), over the
mean of 120, 180 and 300 s (`-combined`, 3,000 iterations, refined at 10,000). The searches kept the pins: Subtlety for
Balance and Wrath, Feral Swiftness for the cat. The knobs are those in `tools/rotopt/druid_presets.sh`. The Faerie Fire
duty builds (meta `baseBuild`) were never searched separately, so they take their base build's talents.
Confirmations: 20,000 iterations, seed 1 (the request's seed), one run per length. Values are DPS (standard error).
"Old" means current talents and current gear, with no `-hit-steps`. "New" means the picked talents with
`-hit-steps ... -hit-step <step>`.

## Summary

| Build | Old talents | New talents | Gear step | Old 120 / 180 / 300 | New 120 / 180 / 300 | Δ |
|---|---|---|---|---|---|---|
| druid__feral-cat | `05002-53200032023012125001-05502` | `050022-53200032023012125001-055` | `level60_phase1_synthetic_hit-4` | 1150.5 (0.3) / 1139.0 (0.3) / 1124.4 (0.2) | 1176.5 (0.3) / 1164.9 (0.3) / 1146.4 (0.2) | +26.0 / +25.9 / +22.0 |
| druid__feral-cat-duty-ff | same as base | same as base | `level60_phase1_synthetic_hit-4` | 1143.3 (0.3) / 1134.1 (0.3) / 1120.4 (0.2) | 1169.4 (0.3) / 1160.1 (0.3) / 1144.2 (0.2) | +26.1 / +26.0 / +23.8 |
| druid__balance | `5232220115501351--505003` | `5432200115501351--505003` (see note) | `level60_phase1_synthetic_hit+4` | 955.1 (0.3) / 952.1 (0.3) / 944.1 (0.2) | 956.9 (0.3) / 952.0 (0.3) / 943.9 (0.2) | +1.8 / −0.1 / −0.2 |
| druid__balance-duty-ff | same as base | same as base | `level60_phase1_synthetic_hit+4` | 916.0 (0.3) / 915.5 (0.3) / 910.0 (0.2) | 918.1 (0.3) / 915.4 (0.3) / 910.1 (0.2) | +2.1 / −0.1 / +0.1 |
| druid__balance-wrath | `5502220115501351--505003` | unchanged | `level60_phase1_synthetic_wrath` (+0) | 858.3 (0.2) / 861.5 (0.2) / 859.3 (0.2) | the same | 0 |
| druid__balance-wrath-duty-ff | same as base | unchanged | `level60_phase1_synthetic_wrath` (+0) | 823.9 (0.2) / 829.0 (0.2) / 829.4 (0.2) | the same | 0 |

## Feral cat (and its Faerie Fire duty)

- Talent change: Nature's Reach 0 → 2 (Balance; +4% melee and spell hit) for Natural Shapeshifter 2 → 0 (Restoration).
- Gear: `ui/feral_druid/gear_sets/level60_phase1_synthetic_hit-4.gear.json`, with 40 hit rating less on the gear
  re-spent on Agility, Strength and crit (`orchestration/tools/hit_steps.json`, "feral").
- Hit before and after: melee 8% = cap 8% (gear 8, talents 0) → 8% = cap (gear 4, Nature's Reach 4). The tool reports
  "total hit on the set 12.00% against 8%" for the step, i.e. what the unstepped set would carry with the talent. Spell hit for the
  duty's Faerie Fire stays 8% against 16%.
- Search: +24.8 combined (noise 0.6) at 3,000 iterations, 1162.9 refined at 10,000 against 1138.6. Nature's Reach 1
  with Natural Shapeshifter 1 was +13.3. Nothing else moved.
- Mana: without Natural Shapeshifter, mana at the end falls from about 4,000 to 2,459 / 1,597 / 482 at
  120 / 180 / 300 s, with a time-to-OOM of about 280 s. The 300 s gain is still +22.0. Fights longer than 300 s
  would run dry; 300 s is the top of the report's range, so nothing in the report runs past it.
- Odd: the search wanted the full −4 step, which is as far as the steps go ("down": 4); Nature's Reach 2/2 is the only
  hit talent, so no further step could be used anyway.

## Balance (and its Faerie Fire duty)

- Talent change the search kept: Nature's Reach 2 → 0 (−4% spell hit) for Genesis 0 → 2, worn with step +4
  (`ui/balance_druid/gear_sets/level60_phase1_synthetic_hit+4.gear.json`, 40 hit rating more on the gear paid from
  spell power and crit).
- Hit before and after: spell 16% = cap 16% (gear 11 + set 1 + Nature's Reach 4) → 16% = cap (gear + set 16, talents 0).
- Search: +1.1 combined (noise 0.5) at 3,000 iterations; refined at 10,000 it is 950.9 against a 950.3 start, within
  noise. Genesis 1 with Nature's Reach 1 was +0.6.
- At 20,000 iterations it is +1.8 at 120 s and −0.1 / −0.2 at 180 / 300 s. The duty build shows the same: +2.1 / −0.1 / +0.1. Mana at the end of 300 s
  drops from 1,207 to 767.
- **Recommendation: keep the current talents and gear.** The change is a tie at 180 and 300 s. Nature's Reach is also
  the range talent: standard 33 puts Balance at 36 yd only because Nature's Reach 2/2 makes its 30 yd spells reach
  that far. Without it the druid could not cast from 36 yd. The raid sim does not check spell range outside
  leveling (`sim/core/spell.go:570`), so the search does not see that cost. Standing at 30 yd would shorten the flight
  time slightly, which the 36 yd request does not credit.

## Balance Wrath (and its Faerie Fire duty)

- Unchanged. The search kept `5502220115501351--505003` at step +0, 859.8 combined. The best Nature's Reach trade
  (Nature's Reach −1 for a point elsewhere, step +2) was −2.8. The New column equals Old because step +0 is the
  current gear set.
- Hit: spell 16% = cap 16% before and after.

## Commands

Build: `go build --tags=with_db -o rotopt ./tools/rotopt`. In a fresh container this first needs
`protoc -I=./proto --go_out=./sim/core ./proto/*.proto`. Knobs:

```
K_BAL="innervateMana=0,filler=3,eclipseStacks=2,prepull=1,prepullTime=3,prepullTrinket=1,cdStack=1,endMoonfire=1,elune=1,wrathFallback=1,clearcastStarfire=0"
K_WR="innervateMana=0,filler=1,prepull=2,prepullTime=1.5,prepullTrinket=1,cdStack=1,endMoonfire=2,endMoonfireSlack=1.8,elune=1"
K_FER="rake=1,rakeMinTime=6,shadowmeld=1,shadowmeldEnd=10,meldBeforeRake=1,ripCP=4,shiftEnergy=60,shiftHoldForBite=0,endBite=2,ripMinTime=8,biteRipTime=4,biteEnergy=45,cdStack=1,tigersFury=0,elune=1,bomb=1,bombEnergy=60,rakeAt5=1,meldAt5=1"
```

The duty builds add `faerieFire=1` (Balance, Wrath), or `faerieFire=1,ffRefresh=7` (Feral).

Searches:

```
rotopt -template druid_balance -settings tools/dps_report/req/druid__balance.json -knobs "$K_BAL" -talent-search -require subtlety -combined -durs 120,180,300 -hit-steps ui/balance_druid/gear_sets/level60_phase1_synthetic -hit-cap spell:16
rotopt -template druid_balance -settings tools/dps_report/req/druid__balance-wrath.json -knobs "$K_WR" -talent-search -require subtlety -combined -durs 120,180,300 -hit-steps ui/balance_druid/gear_sets/level60_phase1_synthetic_wrath -hit-cap spell:16
rotopt -template druid_feral -settings tools/dps_report/req/druid__feral-cat.json -knobs "$K_FER" -talent-search -require feralSwiftness -combined -durs 120,180,300 -hit-steps ui/feral_druid/gear_sets/level60_phase1_synthetic -hit-cap melee:8
```

Confirmations, for d in 120, 180 and 300:

```
# old
rotopt -template <druid_balance|druid_feral> -settings tools/dps_report/req/<id>.json -knobs "<knobs>" -eval -confirm-iters 20000 -durs $d
# new
rotopt -template druid_feral -settings tools/dps_report/req/druid__feral-cat[-duty-ff].json -knobs "<knobs>" -talents 050022-53200032023012125001-055 -hit-steps ui/feral_druid/gear_sets/level60_phase1_synthetic -hit-cap melee:8 -hit-step -4 -eval -confirm-iters 20000 -durs $d
rotopt -template druid_balance -settings tools/dps_report/req/druid__balance[-duty-ff].json -knobs "<knobs>" -talents 5432200115501351--505003 -hit-steps ui/balance_druid/gear_sets/level60_phase1_synthetic -hit-cap spell:16 -hit-step +4 -eval -confirm-iters 20000 -durs $d
rotopt -template druid_balance -settings tools/dps_report/req/druid__balance-wrath[-duty-ff].json -knobs "<knobs>" -talents 5502220115501351--505003 -hit-steps ui/balance_druid/gear_sets/level60_phase1_synthetic_wrath -hit-cap spell:16 -hit-step +0 -eval -confirm-iters 20000 -durs $d
```

Use `-print-stats` with the same flags to see the step a build wears.
