# Pass: consumables, buffs and debuffs

Follow `orchestration/audit/PASS.md` around this doc. It started as the long-running consumables auditor's handoff (2026-10-17); keep it current when the method changes.

This audit checks that every Level 60 DPS preset uses Forever Phase 1 consumables, raid, party and individual
buffs, debuffs and enchants that are:

- correct against the Forever client;
- obtainable at launch;
- consistent with the standing rulings.

It's the per-class correctness side. The fairness auditor owns cross-class equality.

Read the briefs first: `orchestration/auditors_common.md`, `orchestration/auditor_consumes.md` and
`orchestration/standards.md`, on the coordination branch `claude/wow-forever-dps-presets-m9u15h` (they're on master
too). Standards 1b, 8a, 18, 19, 20 and 23 are the ones this audit enforces most.

Hard rules:

- Never change code, presets or docs, and don't commit (the audit run commits your files). No PR comments. Findings go to `orchestration/audit/findings/consumables.md`.
- Don't use `tools/rotopt` searches or any optimiser (-eval, -timeline, -breakdown). Reason from the sim code, the
  client data and the spec's needs. A few paired sims with mksettings + wowsimcli are fine.
- Put no model names in anything written.

## Baseline (end of the session-based audit, 2026-10-17)

- Last audited master: `b389688f7` (pass 32; pass 31 was `d2633f4`). No class branch or the consumes branch was
  ahead of master in audited files.
- Client: Forever `1.60.1.70450` (wago `wowf`/`wow_cn_beta` products). Its consumables and enchants are identical
  to 70425's (see `orchestration/audit/ref/consumables/builddiff_70425_to_70450.txt`).
- Reference files: `orchestration/audit/ref/consumables/` (scan output, per-spec consumes baseline, build diffs,
  MANIFEST for the db2 tables).
- Open items: none. Pending user rulings in this area: none.

## What you watch

1. **origin/master** (all class branches have merged into it). These paths:
   - `ui/<spec>/presets.ts` for the eleven Level 60 DPS specs: warrior, retribution_paladin, hunter, rogue,
     shadow_priest, elemental_shaman, enhancement_shaman, mage, warlock, balance_druid, feral_druid;
   - `ui/<spec>/gear_sets/level60*`;
   - `tools/rotopt/mksettings/spec_*.go` and `main.go` (the shared raid);
   - `sim/core/consumes.go`, `buffs.go`, `debuffs.go`, `totem_weapon_buffs.go`;
   - `sim/common/` (enchant and item effects);
   - `tools/database/forever_enchants.go`;
   - `tools/dps_report/assumptions_def.py` (`oil_proc_damage` and any buff-related assumption);
   - `sim/mage/forever_staff_imbue.go`.

   Ignore tank_warrior, protection_paladin, healing_priest, holy_paladin, warden_shaman and the older non-level60
   gear sets (phase_N.gear.json, p0/p1.bis, mc.gear.json). Those still carry Era enchants like 2504/2544 and are
   out of scope.
2. **The class branches** `origin/claude/wow-forever-dps-presets-m9u15h-<class>`
   (warrior, paladin, hunter, rogue, priest, shaman, mage, warlock, druid) and the shared
   `...-m9u15h-consumes`. Only a branch **ahead of master** needs auditing, and only its own commits.
3. **New Forever client builds** on wago.tools:
   `curl -sS https://wago.tools/api/builds`, versions starting `1.60`.
   On a new build, fetch it and run `builddiff.py`.
4. Blizzard's Forever development notes, when a branch or user ruling cites them. Tooltips and wowhead Forever pages
   for item sources.

**A pass is quiet** when `changes.sh` lists nothing for master or branches and wago has no newer 1.60 build. Then
send a two-line "no findings, nothing moved" summary and stop.

## Setup

```sh
source tools/audit/setup.sh             # AUDIT_WORK and the toolchain
export AUDIT_TREE=$PWD                  # the repo checkout (master)
export AUDIT_WORK=$AUDIT_WORK/consumables; mkdir -p $AUDIT_WORK
tools/audit/consumables/fetch_db2.sh 1.60.1.70450   # and the previous build, to diff against
git -C $AUDIT_TREE show origin/master:tools/database/forever_enchants.go > $AUDIT_WORK/fe.go
```

Sims, only when a choice needs one:

```sh
PATH=$PATH:$HOME/go/bin make -B sim/core/proto/api.pb.go   # after any proto change
go run --tags=with_db ./tools/rotopt/mksettings -spec feral -gear <gear.json> -apl <apl.json> \
  -talents <string> -race <race> -duration 180 -out base.json
```

Then make variants with `python3 -I tools/audit/consumables/var.py base.json out 'name:{"consumes":{...}}'` and run
`tools/audit/consumables/run.sh out_*.json`. Use 20000 iterations, the same seed, and paired comparisons.

## The pass, step by step

1. `tools/audit/consumables/changes.sh <last audited master sha>`. If it lists nothing and wago has nothing new, the
   pass is quiet: send the summary and you're done.
2. For each moved file, read the diff (`git diff <sha> origin/master -- <file>`). Comment-only and DPS-number
   changes need nothing. Look at:
   - **Consumes objects:** every `Consumes.create` in presets.ts and the `consumes:` func in spec_*.go. They must
     match each other.
   - **RaidBuffs / Debuffs / IndividualBuffs objects:** these and the mksettings `raid:` funcs must match each other.
   - **Gear sets:** enchants and trinkets.
   - **Sim code that changes a buff, debuff, consumable or enchant value:** check every number against the client
     with `q.py` / `ench.py`. Compute level scaling as `min(level, MaxLevel) - max(BaseLevel, SpellLevel)` from
     SpellLevels, but see gotcha 3.
3. `tools/audit/consumables/scan.sh` on master and on any ahead branch (`scan.sh origin/claude/...-<class>`).
   Everything above `== counts` must be empty. Compare the counts against the ref file.
4. On a new client build, run `builddiff.py <prev> <new> $AUDIT_WORK/fe.go`.
   - "armor re-encode" lines aren't findings.
   - A CHANGED consumable a preset uses, or a changed Forever enchant, is a finding. Also check whether any
     preset uses the item, with `git grep` on presets.ts and level60 gear.
5. New consumable, enchant or real item in a preset: check its source.
   - `srcid.py <item id>` or `src.py <db2> "<name>"` for the item and its recipe.
   - It's out if its only source is an unopened raid: MC 2717, ZG 1977, BWL 2677, AQ40 3428, AQ20 3429,
     Naxx 3456, Onyxia 2159.
6. New consumable choice: sanity-check it's the best for the spec at Phase 1. One stat-preset elixir per slot,
   one crit elixir, the food, the flask, the weapon imbue. If in doubt, sim it paired against the next best.

## Standing checks (what scan.sh and the review enforce)

- **No raid-only sources** (rule 8a, launch Phase 1):
  - Out: Brilliant Wizard Oil, Elemental Sharpening Stone (MC plans), Enchant Weapon - Spell Power 2504 (MC
    formula), Arcanum of Focus 2544 (ZG), Mageblood Potion (Zandalar vendor), Talisman of Ephemeral Power (MC),
    Fire-toasted Bun (holiday item).
  - Allowed: Cerebral Cortex Compound (Blasted Lands quest); Elixir of Frost Power (holiday recipe, but the elixir
    is tradeable); Mighty Intellect 2568 (Thorium Brotherhood Revered, borderline).
- **Trinkets:** synthetic 99xxxx ids or empty (standards 1b). No real non-relic items. Relics are fine (Undermine
  vendor Pix Xizzix).
- **At most one** of the five +2% crit elixirs (Mongoose, Grizzly, Owl, Sages, Whale) per Consumes preset.
- **Revelation (1248805)** only on the shamans.
- **Standards 18:** Sunder Armor on, Expose Armor off, in presets.ts and mksettings. No rogue APL casts Expose
  Armor (8647, 11198).
- **Standards 19:** the default Level 60 encounter is a level 63 MobTypeUnknown boss. MobTypeHumanoid, Undead and
  Dragonkin appear only in the extra labelled "vs X" encounters and the low-level encounters (e.g. Interrogator
  Vishas).
- **Standards 20:** Faerie Fire and Curse of Recklessness share the MinorArmorReduction exclusive effect in
  debuffs.go (count 2). Both may be on in a preset; only one counts.
- **Standards 23 (duty presets):** a duty preset switches off exactly its own raid effect:
  - Hunter's Mark / Scorpid Sting: hunter mksettings `_mark` / `_scorpid`.
  - Faerie Fire: Balance switches off FF **and** CoR, which share the slot.
  - Warlock: Curse of the Elements.
  - Warriors: one Sunder at the pull.
  - Shamans have **no** totem duty. Every shaman drops its own party's totems, and the raid gives the shaman no
    totem it drops itself.
- **Air and weapon totems (client 70425, user ruling 2026-10-16 "both totems"):**
  - Melee specs (warrior, Ret, rogue, Feral, Survival) get Grace of Air + Windfury Totem + Strength of Earth from
    the raid.
  - Ranged hunters get Grace of Air.
  - Enhancement drops its own SoE / WF / Mana Spring and gets the second shaman's Grace of Air + Flametongue Totem
    (rank 4, 1363/25 = 54.52 per 4 sec).
  - Windfury Totem procs from the main hand only (user ruling 2026-10-16).
- **Salvation** is on in every Level 60 DPS preset and mksettings. It's threat only.
- **No Sanctity Aura or Curse of Shadow** (Forever has neither; its Curse of the Elements covers all magic schools,
  mask 126).

## Settled; don't re-flag

- **Shadow / Frost Oil proc damage:**
  - Presets use Forever's tooltips: 63 to 71 / 39 to 43.
  - Era's 52 to 61 / 33 to 38 is the `oil_proc_damage` Low in the DPS report.
  - User ruling 2026-10-13. A beta log hit above 61 settles it.
  - The server-formula reading (scaling from SpellLevel) is the reason for the Low, not a finding.
- **Trueshot Aura / Hunter's Mark:**
  - Presets use the top ranks "fixed onto the curve", +100 / +140.
  - `hunter_tsa_hm_rank` keeps 75 / 98 (mid) and 50 / 71 (low).
  - User ruling 2026-10-15. The client still says 50 / 71.
- **Weaker top ranks elsewhere:** presets cast the strongest learned rank (16:10 ruling). Ahn'Qiraj book ranks are
  out (Battle Shout VII, SoE V, GoA III).
- **Moonkin Aura / Leader of the Pack:** client 70425 split them back to the Era form. Moonkin Aura gives 3% spell
  crit and LotP 3% melee crit, both raid-wide.
- **Feral:** potions and imbues work in Cat Form under Forever (druid's sim change). Dense Weightstone is on the
  mace; Elemental Sharpening Stone is MC-only.
- **Mage staff imbue scrolls** (items 277500 to 277502, level 46, Comprehension):
  - Fire and Frostfire take Greater Flame +20, Frost takes Greater Frost +20, Arcane takes Precision +1% crit.
  - They take the class imbue slot (SpellEffect 360), so they stack with Wizard Oil.
- **Dragonbreath Chili** is 61 to 69 under Forever.
- **Bracer Superior Agility (1248599) and 2H Strength (1248668):** in the database for comparison, unused because
  launch availability is unconfirmed.
- **Bracer Lesser Intellect:** +3 since 70425 (fixed).
- **Shaman Lightning Shield / raid damage hits and Ret's Concentration Aura:** rotation or class matters, not this
  audit's.

## Findings: what counts and where it goes

- **A finding** is any of:
  - a preset or mksettings using a raid-only or nonexistent item;
  - a buff, debuff, consumable or enchant value that disagrees with the client (or the ruling that overrides it);
  - presets.ts disagreeing with mksettings;
  - a duty preset switching the wrong raid effect;
  - a clearly worse consumable choice, by about 1 DPS or more in a paired sim.
- **Cosmetic only** (no DPS effect, e.g. a removed buff still ticked in the UI, or a stale docs number): report as a
  low-priority nit.
- **Where it goes:** `orchestration/audit/findings/consumables.md` (format in `PASS.md`). Class-owned files and shared files
  (sim/core, mksettings main.go, forever_enchants.go, assumptions, standards) alike are `needs: orchestrator fix`;
  name the class or shared file in `affects`. Cosmetic-only nits are `severity: low`.
- **Every pass** writes `reports/consumables.md`: findings first, then what was verified, then "no open items" or the list.

## Gotchas

1. **Each class branch carries stale copies of other classes' files.** Only audit a branch's own spec files (and
   only commits ahead of master). Grepping the whole branch flags other classes' old presets.
2. **db2 joins:**
   - ItemEffect has no item column. Map items through ItemXItemEffect.
   - Enchant spells are SpellEffect 53 (permanent), 54 (temporary) or 92. EffectMiscValue_0 is the
     SpellItemEnchantment id, whose EffectArg is a stat index (Effect 5) or a proc/aura spell id.
   - Class imbues (mage scrolls, poisons, warlock stones) are SpellEffect 360.
3. **Level scaling:** EffectRealPointsPerLevel times (min(level, MaxLevel) − max(BaseLevel, SpellLevel)) from
   SpellLevels, the server formula. Wowhead's Forever tooltips sometimes show a bigger number. The user went with
   the tooltip for the oils, so on a disagreement report both and let the user rule.
4. **Variance** is a fraction of the base: range = base ± base*variance/2.
5. **Nonexistent Forever items:** wowhead returns a generic "Forever Items" listing for them. That's how Hand of
   Justice, Blackhand's Breadth, Briarwood Reed and others were found not to exist. srcid.py shows "NO SOURCE".
   Cross-check with ItemSparse.
6. **Wowhead may 403 the container.** Don't guess a source. Say it couldn't be checked, or use the db2 plus the
   recipe item.
7. **Shapeshifted potions:** stat and mana potions don't fire while shapeshifted (`ShouldActivate` checks
   `!IsShapeshifted`). The druid changed this for Forever; recheck if consumes.go moves.
8. **Missing enum values:** a consumable enum that doesn't exist on a branch silently gives 0 (the hunter once used
   GreaterMagebloodElixir before merging the consumes branch). If a value looks unmodelled, check that the
   proto/enum exists on that branch.
9. **Building mksettings after a proto change** needs
   `PATH=$PATH:$HOME/go/bin make -B sim/core/proto/api.pb.go`.
10. **Item.csv** wasn't refetched for every build. builddiff.py falls back to the older build's (ClassID is stable).
