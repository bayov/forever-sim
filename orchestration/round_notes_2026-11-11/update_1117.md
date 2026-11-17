# update_1117: PRs #127, #131 and #132, the Nov 13-17 hotfixes, re-sim and publish

Landed on master at 8c60000cf (dev branch `update-1117`, same commit). DPS Report deployed:
https://forever-dps.pages.dev. The round's publish was deploy `f0c58c5b` at ab7915473. The final deploy, `822fe46f`, came
after the post-merge check below (44 files, each checked live). stale.mjs ends at 0 of 68 builds and 0 of 2 modes stale.

## Changes and their sources

- **PR #131** (build 70938 data) merged at 178e57772. Where `forever_seen_in_game.json` conflicted, master's in-game rows
  (Bayov) were kept and the db was regenerated.
- **PR #132** merged at 7ef6f3a56. db.bin was regenerated.
- **PR #127** adds nothing master lacks: Soulwoven Drape 276072 was already in db.json. The PR was closed.
- **The comparison session's landing** (6a816c857, merged here):
  - hunter pets' white swings no longer use the hunter's crit;
  - Predator's Edge counts twice on Raptor Strike;
  - a queued Raptor Strike keeps the off-hand miss penalty;
  - traps roll spell hit;
  - Death Mark changes;
  - Unbridled Wrath from melee abilities except Whirlwind (wago push 114097);
  - a Cleave spends one Reckless Swings charge;
  - a missed or dodged Execute keeps its extra rage;
  - `pal_soc_wf` is now one value.
- **Execute rolls Windfury Totem once.** `warr_exec_wf2` is now one value, with no bug flag (d5529a566). Source:
  Blizzard's Nov 16 post, "Fixed an issue where Mutilate and Execute would have an extra attempt". The knob
  `AssumeWindfuryPreAttack` is still in patch 0004, dead and harmless.
- **A Shieldsworn Holy Strike no longer rolls Windfury Totem** (`SpellFlagNoWindfuryTotem`, Blizzard Nov 16, 4e81669a9).
  Test: TestShieldswornHolyStrikeNoWindfuryTotem.
- **Vampiric Embrace** heals the party 10% and the caster 20% of the priest's own Shadow damage (Blizzard Nov 16,
  4e81669a9). Test: TestRaidForeverVampiricEmbrace.
- **Darkmoon Cards** (wago push 114044, 3ec610c40):
  - Blue Dragon procs on 3% of casts under Forever.
  - Heroism heals 158-238.
  - Maelstrom gets a comment only.
  - Test: `sim/common/forever_hotfix_114044_test.go`.
- **Items:**
  - Wago pushes 114062/114063 (Blizzard hotfixes Nov 13), checked against Wowhead Forever tooltips on 2026-11-17:
    - Darkspear Signet 272068-71;
    - Militant's Band 271913/271915;
    - Souvenir Sea Shell 274749;
    - Everstaring Eye 284097 (Spirit 6).
  - Militant's Band 271914/271916 take the announced +1 mp5 (16b197c5b).
  - Bad Hat 260212 loses +23 Shadow damage (push 114063, 52596be22).
  - Darkmoon Cards Heroism and Maelstrom are item level 65.
  - Onyxia Blood Talisman gets its Forever stats.
  - Escape Plan does 54-102 damage.
  - Test: TestForeverSeenInGameClearsStats.
- **Totem of Tides 22345:** Riptide costs 5% less. Test: TestOrcShamanTotemOfTides.
- **Idol of Rejuvenation 22398:** Regrowth's direct heal is 5% bigger on a target with our Rejuvenation (192cb2e06).
  Test: TestDruidIdolOfRejuvenation.
- **Blizzard's Nov 16 post, the rest:** an agent checked it line by line. These already matched or have no sim effect:
  Wild Growth, Genesis/Pounce, Dark Inspiration, poison ranks and Shank, Windfury's armor, rage and ICD, Mutilate,
  Shield Slam/Bash, Maelstrom Weapon, Thundering Strikes, world buffs on death (push 114137), Flametongue (113955), and
  the push 114044 item cooldowns.
- **Full test suite** (`./sim/...` without sim/web): every package passes, and nothing failed before the round.
  - The only golden that moved is TestBeastMastery, from the BM pick.
  - In report_hashes, the three BM rows moved.

## Re-optimization (rotopt, seeds 1 and 1000003, 120/180/300 s, 20k iterations)

- **Beast Mastery picks `weaveRaptor=0`:** the hunter walks in for every main-hand swing the weave fits, not only when
  Raptor Strike is ready (084813024). Gains per cell (120 / 180 / 300 s):

  | Build | Seed 1 | Seed 1000003 |
  |---|---|---|
  | Base | +0.2 / +0.4 / +2.3 | +0.8 / +1.0 / +2.2 |
  | -wolf | +1.0 / +1.0 / +2.3 | +1.1 / +1.0 / +2.3 |
  | -tallstrider | +1.0 / +0.9 / +2.1 | +0.9 / +1.0 / +2.1 |
- **Marksmanship and Survival:** nothing won.
  - MM tried weaveRaptor, rapidPrepull and weaveManaPerSec; the best was +0.4.
  - SV tried trapFirst, wingClipHold and the trap and swing-gate knobs; every one lost.
- **Warriors:** each moved less than 1% (+0.8 to +0.9%), so they weren't re-optimized.

## Moved builds, 180 s DPS (live before → published)

| Build | Before | Now | Cause |
|---|---|---|---|
| BM | 1182.2 | 1115.4 | Pet white swings lose the hunter's crit; the weave pick adds about +0.4 |
| BM Tallstrider | 1126.6 | 1057.1 | Same |
| BM Wolf | 1140.7 | 1065.6 | Same |
| MM | 1234.3 | 1201.3 | Pet crit |
| MM Tallstrider | 1203.8 | 1173.8 | Pet crit |
| MM Wolf | 1211.7 | 1179.4 | Pet crit |
| Survival | 1034.6 | 1012.1 | Pet crit; traps roll spell hit |
| SV Tallstrider | 1001.8 | 982.5 | Same |
| SV Wolf | 1011.1 | 989.1 | Same |
| Subtlety | 866.8 | 872.2 | Death Mark (comparison landing) |
| Arms | 1069.3 | 1078.7 | Unbridled Wrath from specials (push 114097); Execute keeps rage on a miss |
| Arms vs Dragonkin | 1126.6 | 1135.7 | Same |
| Fury | 1092.8 | 1102.1 | Same |
| Fury 2H | 1103.0 | 1112.4 | Same |

- **Warrior duty presets:** each rose 8.0 to 9.8.
- **Paladins:** the four paladin builds moved only in the assumptions step, since `pal_soc_wf` is now one value. Their
  DPS is unchanged, and so are the three audit builds.

## Tanking and Healing

Both were re-simmed whole.

| Tanking preset | Before | Now | Cause |
|---|---|---|---|
| Prot Paladin Balanced | 2836.5 | 2809.3 | Shieldsworn Holy Strike no longer rolls Windfury Totem |
| Prot Paladin Max survival | 1833.5 | 1780.2 | Same |
| Prot Paladin Max threat | 3272.8 | 3248.8 | Same |
| Prot Warrior Balanced | 3443.2 | 3470.2 | Push 114097 and Execute's rage |
| Prot Warrior Max survival | 1795.2 | 1847.5 | Same; the Warrior is 1st again |
| Prot Warrior Max threat | 4235.0 | 4221.3 | Same |

- **Bear and Warden:** unchanged.
- **Race tables:** they moved for the six paladin and warrior presets. Undead stays on top everywhere.
- **Healing:** all 14 main-run outputs are byte-identical, so no carry moved. No healer wears a changed item, and
  Vampiric Embrace only adds to the priest's own healing from their Shadow damage.

## After the merge

Merging master after the publish brought in other sessions' Manual Crowd Pummeler cooldown (6ea6e47b4) and Five Thunders
set (6875e3b1e). They marked the whole report stale. No number moved:
- The probe cleared all 68 builds.
- Four audit builds were re-simmed in full and came out unchanged: Balance Wrath and its -duty-ff, Arcane, and the Fury
  sunder duty.
- No DPS, tank or heal request wears item 9449 or a Five Thunders piece, so Tanking and Healing were recorded current
  without a re-sim.
- Two Changes-tab lines got corrected history: the old Max survival line now says the Warrior is back on top, and the
  Windfury line notes Paladin Balanced's return to 2,837 before Shieldsworn.
- Commits 2a6ee54ea and 8c60000cf; deploy `822fe46f`.

## Rulings needed

1. **The Beast Mastery weave pick is narrow:** base BM gains only +0.2 and +0.4 at 120 and 180 s on seed 1, though it is
   positive in every cell and +1 to +2.3 on the pet variants. Keep it?
2. **Shieldsworn:** every Holy Strike carries the Windfury flag whenever the talent is taken, and only Windfury Totem is
   excluded. Is that the intended reading of the Nov 16 note?
3. **Darkmoon Card: Heroism's trigger:** the client says ProcChance 100 with no PPM, but the sim uses PPM 2. It needs an
   in-game check.
4. **Vampiric Embrace edge case:** the `PartyBuffs.dark_inspiration` form would stack with a real Shadow priest in the
   raid.
5. **The dead knob `AssumeWindfuryPreAttack`** is left in patch 0004. Should it be removed?
6. **Warriors weren't re-optimized,** since each moved less than 1%. Should the Heroic Strike and Cleave thresholds be
   re-searched anyway, now that they proc Unbridled Wrath?

## Still open in TASKS

- Windfury Totem's attack power (push 114091) still needs a log.
- The hidden_items rerun needs a local client.
- The Elemental phase_3 gear search.
- The 22336/22315 gaps.
- Cannibal's Cleaver.

## Rulings (user, 2026-11-17)

1. The Beast Mastery weave pick (`weaveRaptor=0`) stays.
2. Shieldsworn: the user asked what it is. Explained; their answer on the reading is still open.
3. Darkmoon Card: Heroism stays at PPM 2 for now.
4. Fix the stacking of the given `PartyBuffs.dark_inspiration` with a real Shadow priest's party aura.
5. Remove the dead `AssumeWindfuryPreAttack` knob (`warr_exec_wf2`) from patch 0004.
6. No warrior re-search now. The next warrior rotation search re-searches Heroic Strike and Cleave
   (`orchestration/audit/passes/rotations.md`, open items).
