# What the converter translates

`convert.py` (and `cmpsim`, `theirs/cmpsim/main.go.txt`) do all of these. Each one left out
gives a wrong result. A new class or rotation may need more of the same kind.

## Gear, enchants, stats

- **Synthetic gear.** Our items 990000 and up don't exist in theirs. `items.json` adds their
  stats (`cmpsim` and `cmpstats` load it); `cmpsim` registers what they do beyond stats: on-use
  trinkets 994202, 993016, 993117, 991018, 990015, 990115, 996032, 997022, 995034, 998015,
  998115, 992015, and the hunter, warrior, Enhancement, Feral and Retribution sets' 2 piece
  (1% attack speed; without it theirs ran 5 to 8 DPS low). A new synthetic effect goes in
  `cmpsim`.
- **Weapon DPS:** their Cat and Bear paw reads the item's `unroundedDps`, which `items.json` sets
  for our synthetic weapons (without it Feral is 164 DPS low).
- **Set bonuses theirs doesn't register** go in as bonus stats (`SET_BONUS`), 1% hit as 10 hit
  rating: Grimstitch Armor, the Elemental set's The Spiritcaller's Storm, Manaflare, Demonheart
  Raiment, Raiments of Conviction (Shadow), Grovekeeper's Eclipse (Balance). Vestments of
  Conviction (Smite) gives 10 Spirit. Our synthetic priest sets name two pieces, so only the
  2 piece is live.
- **Enchants:** ours keys Forever's spell ids, theirs enchant ids. Matched by effect id, else
  name and stats; `ENCH` holds hand picks. Theirs silently drops an unmapped enchant.
- **Stats:** their spell stats are `...Damage` where ours are `...Power`; their `SpellDamage` is
  our `SpellPower` plus our `SpellDamage`. `layers.py` and `align.py` compare them that way.
- **Races:** theirs spells Skyborne `RaceWindshaperSkyborne`.
- **Their sheet leaves out talent crit applied per attack** (Cruelty, Weaponmaster's axe crit),
  so `layers.py` shows ours 5 to 10% crit higher on warriors. Compare crits per hit in
  `breakdown.py`.

## Consumables and buffs

- **Consumables:** our enum names to their item ids (`C`). A potion is their `resourcePotId`
  (mana or rage) or `offensivePotId`, and `OtherActionPotion` becomes
  `OtherActionResourcePotion` or `OtherActionOffensivePotion` to match. Theirs applies one
  explosive of a kind, so ours drops its filler (`--keep-filler` keeps it). Our Elixir of
  Greater Firepower on priests is Forever's Elixir of Holy Power (21546).
- **Stones:** theirs refuses a sharpening stone on a blunt weapon; a weightstone (same +8) goes
  in.
- **Buffs:** group buffs are on/off under the single-target names (`arcaneIntellect`,
  `markOfTheWild`, `powerWordFortitude`, `shadowProtection`). Ranked buffs are ints, 1 for the
  trainer rank and 2 for the book's (`strengthOfEarthTotem`, `graceOfAirTotem`,
  `greaterBlessingOfMight`, `greaterBlessingOfWisdom`); `battleShout` 1 is rank 6, 2 rank 7.
  Our raid Flametongue Totem is `flametongueTotem`. Mana Spring Totem is a party buff there.
  Their Battle Shout lands just after the pull's sheet, so `align.py` adds its 115 attack power.
  Our Improved Battle Shout, Blessing of Might and Gift of the Wild match their plain ones.
- **Sunder Armor:** both ramp the raid's copy to 5 stacks at 6 s, so ours keeps its ramp.
- **Raid damage taken:** theirs has none. `--no-raid-damage` turns ours off. For shamans their
  `shieldProcrate` is set to the same hits a minute, spending Lightning and Water Shield charges
  as ours does.

## Rotation

- **Values theirs lacks:** `spellUsesRemaining` (also `<= k`) and `spellUsesLostByDelay` are
  rewritten from `remainingTime` with cooldowns in `CD`; `moveTime` and `inMeleeRange` from
  `unitDistance` and `unitMovementSpeed`. `autoTimeToNext` and `autoTimeSinceLast` take
  `MainHandAuto`, `OffHandAuto`, `RangedAuto` (`autoSwingTime` keeps `MainHand`).
- **A talent spell the build lacks** (`NOSPELLS=<id>,...`): ours reads a value naming it as
  false, theirs drops it from its `and`, making the `and` true.
- **Vanish and Sweeping Strikes** are DPS cooldowns in theirs, so `autocastOtherCooldowns` would
  Vanish at the pull and Sweeping Strikes every 30 s (32 DPS on Arms). A never-true line names
  each.
- **Charge:** ours opens with it for 15 rage at any range; theirs moves in and needs 8 yd, so a
  rotation opening with it starts theirs at 20 yd. Their Charge also holds off-GCD items after
  the first GCD to 3 s (costs nothing measurable on Fury).
- **Totems:** ours asks by the buff on the shaman (`auraIsActive`) or, for a fire totem, as a DoT
  (`dotIsActive`). Theirs keys neither by the totem's spell, so the converter asks the
  element's slot (`totemRemainingTime`). Untranslated, theirs recasts Strength of Earth 50 times
  and goes OOM.
- **Off-GCD items:** theirs fires one only at the next GCD check, so a rotation using one in
  between (Thistle Tea at 10 energy) loses it there. Compare with care.
- **The next spell is picked before a hardcast completes** in theirs, so a line whose condition
  the ending cast changes reads the old state (untranslated: Fire 180 DPS from double
  Pyroblasts; extra Scorches, Arcane Blasts, Starfires on one Clearcasting). The converter adds
  `not spellIsCasting(X)` to such a line (`SELF_SPENT`) or counts the ending cast's stack
  (`STACKS`: one stack fewer needed, no refresh while casting). Ours' `spellInFlight` becomes
  `spellInFlight or spellIsCasting` for every class. A new class with a stacking or consumed buff
  in a condition needs the same.
- **A channel and the wand:** their pass continues after a channel starts, so a wand line below it
  (Shoot has no GCD) fires and cancels the channel (Arcane 300, Shadow 330 DPS). Their wand line
  waits while Arcane Missiles, Evocation or any channel the rotation names is channeling.

## Spell ids and ranks

Theirs registers only some ranks; `remap()` moves ours.

- Rogue: Backstab VIII, Eviscerate VIII, Deadly Poison IV, Vanish rank 2, Mutilate's last rank,
  Berserking 20554.
- Hunter: top ranks only, the AQ books' (Aspect of the Hawk VII, Serpent Sting IX), so hunter
  runs use `ours_aq_cli` with our ids moved up (`OURS_AQ=0` runs ours on its own ranks; the
  difference is the AQ ranks' worth). Raptor Strike's queue is tag 3 in ours, tag 1 in theirs.
- Warrior: Whirlwind is cast as its main hand copy, tag 14 (untagged it never casts, about 100
  DPS on Fury). Charge has no rank. Berserking 20554.
- Shaman: Berserking 20554; every rank our rotations cast.
- Mage: Berserking 20554; only Frostfire Bolt's top rank, so a downranking line (about 1 DPS)
  never casts there.
- Warlock, priest, druid: every rank our rotations cast.
- Paladin: Holy Shock is keyed by its damage row (rank 4, 20930, is 25902).

## Class options

- **Hunter:** the quiver and pet speed enums, the Crocolisk spelling, and their Beast Mastery
  tree with Pathfinding and Improved Revive Pet swapped (5th and 6th).
- **Mage:** our armor enum to `defaultMageArmor`, staff imbue scrolls to `mhWeaponImbueId`
  277500 to 277502.
- **Warlock:** Spellstone and Firestone to `mhWeaponImbueId` 13603 and 13701, `sacrificedPet` to
  `sacrificedDemon`, 32 Soul Shards. Curse duty: their `curseOptions` names the warlock's own
  curse and turns its raid copy off. The converter sets `Elements` or `Recklessness` when our
  rotation casts that curse with the raid's copy off, else `NoCurse` (every raid curse on).
- **Priest:** our `shadowPriest` armor to their `dpsPriest` options. Ours doesn't count a DoT's
  hit roll in the hit columns, so `breakdown.py` shows misses on Shadow Word: Pain and Devouring
  Plague only in theirs. Compare ticks per cast instead.
- **Druid:** `balanceDruid` and `feralDruid` to their `balanceDruid` and `feralCatDruid`
  (Innervate's target, the cat's `prepullClearcasting`). Theirs never autocasts Innervate; ours
  does at low mana.
- **Paladin:** theirs plays the rotation only (seal and aura from it), so `retributionPaladin`
  carries empty options.

## Reading the results

- Row labels that differ (ours, theirs): Seal of Command hit 20947, 20424; its Judgement 20966,
  20968; theirs splits Seal of Righteousness and Seal of Command into the swing's and the stored
  copy (tags 1311704, 1311703); Shadow Oil 1382, 350026; Touch of the Grave 1260189, 1260198.
- Their idle pets (the warlock's four demons) have no average in their results: `run.py` and
  `breakdown.py` read it as 0. Theirs files a cast on the caster (a potion) under the caster's
  own target, so `breakdown.py` counts casts on every target.
