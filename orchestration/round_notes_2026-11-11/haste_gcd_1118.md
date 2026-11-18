# Haste shortens a spell's GCD (2026-11-18)

Brief: `briefs/haste_gcd_1118.md`. Master e14da3fae (data commit b69e037f5), deploy dacc8ad1.

## The rule as coded

- Warchief's Blessing (16609) adds 15% cast speed under Forever (`sim/core/buffs.go` ApplyWarchiefsBlessing). Evidence: Lightning Bolt casts in 1.30s with it and 1.505s without.
- Under Forever a spell's GCD is 1.5 / (1 + spell haste), floored at 1s (`sim/core/cast.go` hastedGCD). Evidence: instants sit 1.27 to 1.33s apart with Warchief's Blessing up, and 1.52s without it.
- A "spell" has no melee or ranged defense type and a magic school. So Seals are hasted. Shouts and aspects (no school), melee and ranged abilities, and anything physical keep their GCD: the log has Stormstrike's at 1.40s or more.
- The floor follows wowsims/forever: every GCD above 0 ends at 1s or more after haste, a base GCD under 1s included (user ruling 2026-11-18, c2efe4ae3). No spell of ours has a base GCD under 1s, and `Cast.EffectiveTime` already floored the wait at 1s, so the floor moves nothing.
- Nature's Grace's 10% GCD cut composes with its haste: 1.35 / 1.1 = 1.227s alone, and 1.35 / (1.15 x 1.1) = 1.067s with Warchief's Blessing.
- The `haste_gcd` knob is gone from patch 0004. The assumption keeps only the open part.

## Tests

- `sim/core/hasted_gcd_test.go`, TestHastedGCD covers:
  - no haste, Warchief's Blessing (1.304s), a Seal;
  - the floor at haste 2, and a 500ms base GCD floored to 1s;
  - a totem at 1s, Nature's Grace with Warchief's Blessing (1.067s at 1.15 x 1.1);
  - melee, physical, ranged arcane, a rogue's 1s GCD, a shout, and GCD 0.
- TestWarchiefsBlessingCastSpeed: Forever only.
- `sim/shaman/enhancement/gcd_test.go`: Lightning Bolt at 100% haste gives 1.25s, Earth Shock with Warchief's Blessing 1.304s and at 100% haste 1s, Searing Totem 1s, Stormstrike 1.5s.
- `sim/raid_healing_druid_test.go`: Nature's Grace on heals gives 1.35/1.1.
- Goldens regenerated for Balance, Mage, the paladins, Shadow, the shamans and the warlocks, plus 31 report hashes.

## Open edge cases (notes.md, To Verify in-game)

- The 1s floor: a user ruling, not yet checked against a log.
- Haste other than Warchief's Blessing: Berserking, Spellstone, Nature's Grace.
- Which casts count as spells:
  - hunter magic shots are ranged, so they keep 1.5s;
  - Judgement is off the GCD in the sim;
  - Seals are hasted.

## DPS at 180s (before -> after)

No build moved by more than 1%, so nothing was re-optimized.

| Build | Before | After |
|---|---|---|
| Balance | 951.7 | 955.6 |
| Balance FF duty | 915.5 | 920.0 |
| Balance Wrath | 861.5 | 865.8 |
| Balance Wrath FF duty | 828.8 | 833.9 |
| Frost mage | 943.9 | 945.4 |
| Paladin Level 60 | 1023.0 | 1021.5 |
| Paladin crusader duty | 1016.3 | 1015.2 |
| Paladin wisdom duty | 1009.9 | 1006.2 |
| Shock Ret | 1009.9 | 1007.8 |
| Shock Ret crusader duty | 1008.8 | 1005.5 |
| Shock Ret wisdom duty | 999.6 | 996.3 |
| Sub Holy | 1011.6 | 1009.1 |
| Sub Holy crusader duty | 1006.0 | 1004.9 |
| Sub Holy wisdom duty | 999.7 | 995.9 |
| Sub Holy Undead | 1138.0 | 1137.4 |
| Elemental | 976.9 | 980.6 |
| Elemental Mana Tide | 959.6 | 963.4 |
| Enhancement | 1080.0 | 1079.9 |
| Enhancement Mana Tide | 977.2 | 976.1 |
| Enhancement Water Shield | 1025.7 | 1024.7 |
| Affliction | 1070.8 | 1076.4 |
| Affliction curse duty | 1060.3 | 1066.2 |
| Affliction recklessness duty | 1051.8 | 1058.1 |
| Affliction Shadow Bolt | 1015.1 | 1019.7 |
| Demonology | 1054.0 | 1056.1 |
| Demonology curse duty | 1043.9 | 1047.1 |
| Demonology recklessness duty | 1036.8 | 1040.4 |
| Destruction | 1062.8 | 1068.7 |
| Destruction curse duty | 1052.7 | 1058.9 |
| Destruction recklessness duty | 1046.2 | 1052.6 |

Some builds held their 180s number and moved only in their race, length or weight sims: Arcane, Fire, Frostfire, the three priests, Enhancement Imp GW and the three Destruction Fire builds. That is mostly Troll Berserking. No warrior, rogue, hunter or feral build moved, and no report preset turns on Warchief's Blessing.

The paladin drops are small (0.1 to 0.4%). Their rotations were timed around a 1.5s GCD, and seal and judgement timing shifts with the hasted Seal GCD.

## Tanking and Healing

- **Tanking:** only the Warden Shaman moved, and it stays 4th at every goal:
  - Balanced: 1,282.8 -> 1,288.0
  - Max survival: -2,221.2 -> -2,192.8
  - Max threat: 2,645.7 -> 2,621.0
- **Tank healers (carry):**
  - Restoration Shaman: 1.2571 -> 1.2796, 4th -> 1st.
  - Holy Priest: 1.2776, 1st -> 2nd.
  - Discipline Priest: 1.2717, 2nd -> 3rd.
  - Restoration Druid: 1.2687 -> 1.2674, 3rd -> 4th.
  - The top three are within 0.008.
- **Raid healers (order unchanged):**
  - Holy Priest: 1.4614 -> 1.4776
  - Holy Paladin: 1.4300 -> 1.4179
  - Restoration Shaman: 1.4116 -> 1.4006
  - Restoration Druid: 1.3924 -> 1.3864

## Merged before publish

- wowsims tick haste (8219f1083: Hurricane and Volley): the probe of all 68 builds was unchanged, and so were the Tanking and Healing benches.

## Rulings needed

1. **The `haste_gcd` assumption text.** It still says the log tested neither the floor nor other haste, and it cites wowsims for the floor. Should it now cite the user ruling? Rewording it makes all 68 builds stale, so it was left alone.
2. **Warchief's Blessing.** wowsims/forever gives it melee haste only. Ours adds 15% cast speed, from the log. That's in known_differences, and they may want to adopt it.
3. **Paladin small drops.** Should the paladin rotations be re-timed for the hasted Seal GCD? Every move is under 0.4%, below the 1% re-optimize bar.

## Rulings (user, 2026-11-18)

1. The `haste_gcd` assumption text stays as it is.
2. The paladin rotations stay under the 1% bar for now. Re-time them the next time a paladin rotation changes
   (`orchestration/audit/passes/rotations.md`, open items).
3. Warchief's Blessing stays off in every preset.
