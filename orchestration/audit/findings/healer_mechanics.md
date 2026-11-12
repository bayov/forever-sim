# Findings: healer_mechanics

## Open

### healer_mechanics-20261104-11: Riptide's Forever mana is 0.6 x the old client cost; client 70790 now gives 150/180/230
- severity: low (rank 3, the one the level 60 presets cast, is 1 mana dear at 231 vs 230; rank 1 is 3 mana cheap)
- affects: Restoration shaman (all presets and leveling)
- found: 2026-11-04, master 0facd9428
- evidence: client 1.60.1.70790 (wago.tools, tier 1) puts the dev notes' Riptide into the tables: SpellPower 304804,
  310863, 310864 ManaCost 245/300/385 -> 150/180/230; SpellCooldowns 84834, 94614, 94615 CategoryRecoveryTime 6000 ->
  12000; SpellEffect 1072707, 1236793, 1236796 (effect 0) Effect 10 HEAL -> 3 DUMMY, base 0. HoT ticks (89/115/161,
  coef .1) and the Chain Heal bonus row are unchanged. The sim (`sim/shaman/healing.go:104-135`) keeps the 70755 row
  values and scales the cost by `RiptideNotesManaScale` 0.6 rounded: 147/180/231. Cooldown and no instant heal already
  match. `python3 tools/sources/wago.py diff 1.60.1.70755 1.60.1.70790 --set mechanics` reproduces it. Sure.
- proposed fix: set RiptideRanks' manaCost to 150/180/230, cooldown 12 s, instant heal base 0 from 70790 and drop the
  notes switch (`riptideNotes`, `RiptideNotesManaScale`, `RiptideNotesCooldown`) so Forever reads the client; update the
  Riptide table in `docs/healing_shaman.md` (:284), the tests that pin 231/147 (`sim/shaman/enhancement/healing_test.go:294`,
  `healing_talents_test.go:105`), and close shaman call 1 (the client now agrees with the notes). DPS/HPS
  effect is negligible (1 mana a cast at 60); not measured.
- needs: worker fix
- update 2026-11-07 (master 12aa2796f, client 70862): SpellPower 304804/310863/310864 still 150/180/230; the sim still
  computes 147/180/231 (`sim/shaman/healing.go:117,132`). Still open.

### healer_mechanics-20261104-10: many heal ranks, healing talents and engine paths have no test that fails when they break
- severity: low (none found wrong apart from the findings above)
- affects: all four healers and the engine
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: the lists are in `reports/healer_mechanics.md` ("Untested"). The largest gaps: the raid damage share below one
  target (finding 1), heal assignments, Lowest Health by percent, the 1.5x heal crit and HoT tick crits, cancelled heals'
  cost and five-second rule, two absorbs on one unit; Swiftmend's HoT surviving (finding 2); most middle heal ranks of every
  class; Holy Power on Holy Shock and Light's Vigil; Spiritual Healing / Holy Specialization beyond Greater Heal; Prayer of
  Mending's triggers; Riptide tick crits and Tidal Mastery on them; all healer set bonuses; and no spec suite at all in
  `sim/druid/restoration` (`sim/priest/healing/healing_priest_test.go` has only commented-out suites).
- proposed fix: add the tests named in the report, one heal or talent at a time, with the class's next change.
- needs: worker fix (engine items: orchestrator fix)
- update 2026-11-04 (master 0facd9428): the class workers added `sim/shaman/enhancement/healing_untested_test.go`
  and `healing_set_test.go` (rank table, Riptide tick crits, Ancestral Fortitude with two shamans, the T1 set),
  `sim/paladin/holy/healing_more_test.go` (lower ranks, Holy Shock and Light's Vigil ranks, Benediction, Holy Power),
  `sim/druid/healing_ranks_test.go` and Swiftmend's kept HoT in `sim/raid_healing_druid*_test.go`, and priest suites in
  `sim/priest/healing/healing_priest_test.go`; all pass. Still no engine test for the three engine fixes: the raid
  damage share under one target (`TestHealingFightRaidDamageShare` has no share x N < 1 case), the health clamp when
  max health falls, and the school crit bonus on direct heals; nor for heal assignments, Lowest Health by percent,
  cancelled heals, or two absorbs on one unit.

## Resolved

### healer_mechanics-20261111-13: Holy and Discipline priest suites fail on master: the AllItems-IronweaveBattlesuit golden is stale
- severity: low (test hygiene; no healer mechanic shown wrong)
- affects: Holy and Discipline priest test suites (also Shadow priest and Mage, same row)
- found: 2026-11-11, master 707ae92cc5
- evidence: `go test -tags=with_db ./sim/priest/healing/` fails only `Phase1-AllItems-IronweaveBattlesuit`: Holy HPS 249.986 ->
  252.680, Disc HPS 109.002 -> 109.557 (TPS, DTPS too); `./sim/priest/shadow/` DPS 941.581 -> 983.555 and `./sim/mage/`
  945.088 -> 999.375 on the same row; every other row of those suites passes. The eight Ironweave items and
  `ItemSetIronweaveBattlesuit` are unchanged since 27d0d59d6, so the move comes from something else since then (item, enchant or
  consumable data: c4e7fbab3, ca07df649, b1291e61a, `sim/core/consumes.go`); not bisected (a worktree at 27d0d59d6 fails setup
  without generated protos). A red suite hides the next real healer regression.
- proposed fix: find the commit that moved the row (bisect `go test -tags=with_db ./sim/mage/ -run TestP1Mage`), confirm the
  move is intended, then `make update-tests` for the four suites.
- needs: orchestrator fix
- resolved: 2026-11-12, cb0d928fd (no code change; a later full run regenerated the goldens, `go test --tags=with_db ./sim/priest/healing/` passes on current master)

### healer_mechanics-20261110-12: Nature's Swiftness doesn't reach Wild Growth; the Nov 10 hotfix makes Wild Growth instant under it
- severity: medium (Restoration druid raid carry +1.0% with the fix; tank carry unchanged)
- affects: Restoration druid (presets take both talents: `ui/restoration_druid/presets.ts:87`, string `05302001--5053025153113051`)
- found: 2026-11-10, master da272241e5
- evidence: foreverchanges.pro "A hotfix to the November 10 client" (first seen 16:51 UTC, launch realms' hotfix cache):
  "Nature's Swiftness now also works on Wild Growth, which it makes an instant cast instead of 3 sec". Tier 1 data (client
  hotfix cache) at second hand; wago's 70888 tables don't carry it yet (17116 effect 695239 mask unchanged), Blizzard's
  Nov 9 post doesn't mention it. TASKS.md (da272241e) lists it too. Sim: `sim/druid/healing.go` `registerNaturesSwiftnessCD`,
  `affected` takes only Healing Touch, Regrowth, Wrath; the comment says "Wild Growth isn't one".
  Measured with the patch below, `tools/healbench` 1000 iterations paired seeds (seed 101): `-role raid -carry
  tools/healbench/req/druid_restoration_raid.json` carry 0.866 -> 0.875 (HPS 2518 -> 2527, mana left 948 -> 865);
  `-role tank -carry tools/healbench/req/druid_restoration.json` 1.354 -> 1.354.
- proposed fix: add `|| spell.SpellCode == SpellCode_DruidWildGrowth` to `affected` (patch kept at
  `/tmp/audit-work/healer_mechanics/ns_wild_growth.patch`, one line), fix the comment, add a test that Wild Growth
  under Nature's Swiftness casts in 0 s and spends the buff. Confirm against wago once a build carries the mask.
- needs: worker fix
- resolved: 2026-11-11, 9628e85b6 (Wild Growth in `affected`; `TestDruidNaturesSwiftnessOnRegrowth` now checks an instant Wild Growth takes the buff)

### healer_mechanics-20261104-1: a raid damage event's share is rounded up to one target, so small-share abilities hit far too often
- severity: high (moves the raid carry by 3 to 30% and flips the shaman / priest order on Chromaggus)
- affects: engine, every healer's raid and `both` roles on the boss files (Onyxia, Chromaggus, Magmadar, Broodlord most)
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: `sim/core/raid_damage.go:188-196` rolls `count` from `share × non-tanks` and then `count = max(count, 1)`, so
  a cast with share × N < 1 always hits one player instead of one player with probability share × N.
  `tools/encounters/raid_from_logs.py:77-79` fits `everySeconds` on `share × hits` per cast, so the floor multiplies the
  intended damage: with healbench's 9 non-tanks, 24 of the 30 non-tank events across the boss files are below one
  target, and raid damage runs 1.15x to 2.83x the logs' (Onyxia 2.52x; an Onyxia combat log shows 19 Flame Breath hits
  where about 6 were intended). This also breaks `docs/healing_sim.md:428-429` ("Every file's events land what the logs
  did, within 0.1 a second") for a 9-player raid. Measured with healbench raid carry, floor kept → dropped:
  Onyxia shaman 0.685 → 0.889, priest 0.714 → 0.945; Chromaggus shaman 0.900 → 1.041, priest 0.904 → 0.987 (order
  flips); Magmadar +11% / +15%; Broodlord +7% / +3%. Onyxia's other-death rate falls from 53% to under 12%. Runs and
  modified encounter files: `$AUDIT_WORK/healer_mechanics/share_runs2/`, `enc/`, `run_share2.sh` (scratch, gone with
  the session). Sure of the code and the numbers.
- proposed fix: drop the `max(count, 1)` floor (a cast then hits nobody with chance 1 − share × N, keeping the mean
  right), or keep one target and stretch the event's interval by 1 / (share × N). Add a share × N < 1 case to
  `TestHealingFightRaidDamageShare`, then re-run the raid carry tables in `docs/healing_sim.md` and the class docs.
- needs: orchestrator fix
- resolved: 2026-11-04, b2c16f671 (re-checked on master 0facd9428)

### healer_mechanics-20261104-2: Swiftmend still consumes the Rejuvenation or Regrowth it reads; Forever's does not
- severity: high (tank carry +1.5% on average, Vaelastrasz +7.7%, Ragnaros +3.1%, Magmadar +2.4%)
- affects: Restoration druid (tank and `both` presets)
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: `sim/druid/healing.go:695` calls `hot.Cancel(sim)` before the heal; `TestDruidSwiftmend`
  (`sim/raid_healing_druid_test.go:381-390`) and the doc's Swiftmend section (`docs/healing_druid.md:129`, :140, calls
  2 and 3) assert it is consumed. Tier 1, two sources: Blizzard's druid article news 24301515 (saved copy
  `orchestration/audit/ref/client_update/sources/news24301515.txt:177`): "Swiftmend: … Redesigned to not consume the
  Rejuvenation or Regrowth effect on the target."; dev notes post 9 (us.forums.blizzard.com/en/wow/t/2360696/9, edited
  2026-11-04T02:12Z): "does not consume". Client 70755's 18562 tooltip still says "consumed" (a tooltip alone
  doesn't outrank the notes). The mechanics auditor's note that "the sim has no Swiftmend"
  (`orchestration/audit/ref/mechanics/verdicts.md:90`) is wrong; the sim has one. Measured: healbench `-carry`,
  500 iterations, the shipped requests and rotations with only the cancel removed: tank carry Vaelastrasz 0.689 →
  0.742, Ragnaros +3.1%, Magmadar +2.4%, mean +1.5%; effective HPS up to +23%; raid role flat within noise; no
  druid-vs-shaman tank order flips on these files. Lower bounds, since the rotations were tuned for a consuming Swiftmend.
  Sure.
- proposed fix: remove `hot.Cancel(sim)` (the HoT ticks on and still Blooms when it ends), fix the comment, the doc's
  Swiftmend section and calls 2-3, make `TestDruidSwiftmend` require the HoT to survive, and re-tune the Swiftmend and
  Rejuvenation lines of both druid presets.
- needs: worker fix
- resolved: 2026-11-04, bc8e5473b (re-checked on master 0facd9428)

### healer_mechanics-20261104-3: Ancestral Fortitude raises only item armor; its client aura is the all-armor one
- severity: medium (estimate, not measured: about +200 to 300 tank armor while it is up, under 1% of the tank score)
- affects: Restoration shaman (tank healer)
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: 16177 SpellEffect #0 EffectAura 101 misc 1 (MOD_RESISTANCE_PCT, armor), the same aura as the priest's
  Inspiration 14893, which the sim applies to all armor (`sim/priest/healing_talents.go:92-108`, `ArmorMultiplier`).
  The "from items" talents use aura 142 (Toughness 12299 / 20143). Only 16177's tooltip says "from items"; the talent
  16176's tooltip says "Increases your target's armor value". `sim/shaman/talents.go:772-814` follows the 16177 tooltip
  (`ApplyDynamicEquipScaling`), as does `docs/healing_shaman.md` item 4. Tier 1 aura types; how the server applies 101
  vs 142 is server side. About 70% sure.
- proposed fix: apply it as Inspiration (all armor), or document why aura 101 here means item armor; update
  `TestOrcShamanAncestralHealing`, `TestOrcShamanHealTalentPoints`, `TestHealingFightShamanAncestralFortitude` and the doc.
- needs: worker fix
- resolved: 2026-11-04, 4648ff419 (re-checked on master 0facd9428)

### healer_mechanics-20261104-4: Benediction cuts Divine Favor's and Oath of Light's cost; the client mask can't reach them
- severity: low (about 5 mana a minute, under 0.05% of the preset's spend; estimate)
- affects: Holy paladin
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: Benediction 20101 e0 aura 108 misc 14, mask 0x1be11fbf / 0x26018242 / 0x20212100 / 0x8800; SpellClassOptions
  of 20216 (Divine Favor) and 1328548 (Oath of Light) are 0 / 0 / 0 / 0, so no class-mask modifier reaches them (Holy
  Shock, 0x200000, is in the mask). The sim applies `paladin.benediction()` at `sim/paladin/divine_favor.go:68` and
  `sim/paladin/holy_talents.go:128`; `docs/healing_paladin.md` repeats it in three places. The repo already follows the
  mask over the "all instant spells" tooltip for Hammer of Wrath (`sim/paladin/hammer_of_wrath.go:53`). Tier 1, sure.
- proposed fix: drop the multiplier on both spells and the three doc lines.
- needs: worker fix
- resolved: 2026-11-04, d64804b9f (re-checked on master 0facd9428)

### healer_mechanics-20261104-5: the Vestments of Conviction 5-piece misses Prayer of Mending
- severity: low (no preset wears the 5-piece: 0 today)
- affects: Discipline / Holy priest; the Smite synthetic set
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: 1301703 aura 107 misc 11 bp -1000, mask word 1 0x800020 = Penance and Prayer of Mending (70755, tier 1).
  `sim/priest/item_sets_pve.go:222-229` takes 1 s off Penance only. Sure.
- proposed fix: also take 1 s off Prayer of Mending's cooldown.
- needs: worker fix
- resolved: 2026-11-04, dda14db65 (re-checked on master 0facd9428)

### healer_mechanics-20261104-6: the Restoration shaman's synthetic gear carries no Tier 1 set (standards 6)
- severity: low (+10 Spirit live; the 4-piece +26 healing and 5-piece Riptide -1 s unreported)
- affects: Restoration shaman
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: client bonus spells 1300960 (+10 Spirit), 1301059 (totem radius), 1301093 (+26 healing, 9 of it damage),
  1301713 (Riptide -1 s, aura 107 mask word 2 0x10). `orchestration/standards.md:49` (rule 6): each class keeps its real
  Tier 1 set on two synthetic pieces with the 2-piece live, and reports the 4- and 5-piece gain. The druid, paladin and
  Enhancement / Elemental sets do; `tools/database/forever_synthetic_items_shaman_restoration.go` has none, and
  `docs/healing_shaman.md` says "the synthetic set doesn't use it". Sure of the data; whether it was deferred on purpose
  is for the orchestrator.
- proposed fix: register the set on two Restoration synthetic pieces, and report the 4- and 5-piece gain separately.
- needs: worker fix
- resolved: 2026-11-04, c938028ee (re-checked on master 0facd9428)

### healer_mechanics-20261104-7: direct heals don't take school crit bonuses that HoT ticks do
- severity: low (no preset wears a school-crit set today)
- affects: engine; the Vestments of Prophecy 5-piece (+2% Holy), The Ten Storms 5-piece (+3% Nature)
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: `sim/core/spell_result.go:306` `HealingCritChance` leaves out `GetSchoolBonusCritChance`; HoT ticks, which go
  through `SpellCritChance`, get it. Sure.
- proposed fix: add the school bonus term to `HealingCritChance`, with a test.
- needs: orchestrator fix
- resolved: 2026-11-04, 0aed93989 (re-checked on master 0facd9428)

### healer_mechanics-20261104-8: current health isn't clamped when max health falls
- severity: low
- affects: engine (tank under Last Stand and other max-health buffs)
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: when Last Stand ends (`sim/warrior/talents.go:613`) the tank keeps health above max; the next heal
  (`sim/core/health.go:113`) cuts it to max and books a negative gain into healing done. Sure of the code; size not measured.
- proposed fix: clamp current health to max when max health changes.
- needs: orchestrator fix
- resolved: 2026-11-04, 0aed93989 (re-checked on master 0facd9428)

### healer_mechanics-20261104-9: doc numbers wrong for some druid ranks and the paladin's Libram of Holy Alacrity
- severity: low (docs only; the sim's numbers are right)
- affects: Restoration druid, Holy paladin docs
- found: 2026-11-04, master a1b1163d2 (lines re-checked on f62bfc7)
- evidence: `docs/healing_druid.md` gives Healing Touch 8 as 1100-1311 (client 1095-1303), Regrowth 8 815-911 (800-894),
  Regrowth 7 659-738 (633-710), Tranquility 3 tick 205 (201.4), 2 tick 136 (133.2): growth past each rank's max level.
  `sim/druid/talents.go:113-116` comment says Genesis's HoT healing is unused; it is used. `docs/healing_paladin.md`
  (Items, call 9) says Libram of Holy Alacrity's buff 449982 has no effect rows in 70753; 70755 (byte-identical tables to
  70753 per the mechanics verdicts) has e0 aura 107 misc 10 bp -200 on Holy Light, 10 s, 1 charge, triggered from
  449980 (item 228175). The trigger condition is server side. Tier 1, sure.
- proposed fix: correct the numbers and the comment; answer paladin call 9 from the client rows.
- needs: worker fix
- resolved: 2026-11-04, d64804b9f (paladin call 9), bc8e5473b (druid tables, Genesis comment) (re-checked on master 0facd9428)

