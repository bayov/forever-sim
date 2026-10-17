# Pass: Forever mechanics

Follow `orchestration/audit/PASS.md` around this doc. It started as the long-running mechanics auditor's handoff
(2026-10-17); keep it current when the method changes.

You check that every mechanic the presets use is WoW Forever's, not Classic Era's, Season of Discovery's or a wowsims
default. That covers the shared engine (`sim/core`), the shared raid (buffs, debuffs, consumes), racials, and class
assumptions inherited from Era or SoD. The background briefs are `orchestration/auditors_common.md` and
`orchestration/auditor_forever.md`; the rulings are in `orchestration/standards.md`.
`orchestration/audit/ref/mechanics/verdicts.md` holds what is verified, settled and open: read it first, and keep it
current when a pass verifies or settles something.

## Rules

- **Read-only on code.** Building settings with mksettings and running the CLI locally is allowed. Don't use
  tools/rotopt (searches, -eval, -timeline, -breakdown). Write a proposed code change as a patch file under
  `$AUDIT_WORK/mechanics/` and put it, or its key hunk, in the finding's proposed fix.
- **Sources, in tier order:**
  - Tier 1: the Forever beta client on wago.tools, beta development notes (us.forums.blizzard.com/en/wow/t/2360696 and
    later posts), blue posts, Blizzard Forever articles (news.blizzard.com 24301514 for warrior and priest, 24301515
    for druid), beta logs.
  - Tier 2: Season of Discovery.
  - Tier 3: Era, Anniversary and 2019 Classic.
  - Label every verdict with its tier. Never cite vmangos. Never decide from a tooltip alone.
- **Where findings go:** `orchestration/audit/findings/mechanics.md` (format in `PASS.md`). Shared-engine, shared-raid
  and racial findings are `needs: orchestrator fix`; class findings are `needs: worker fix` with the class in
  `affects`. Pet buffs: see `orchestration/pet_buffs.md` before flagging anything there.
- **No model names** in anything written.

## Setup

```sh
source tools/audit/setup.sh                            # AUDIT_WORK, protoc, protoc-gen-go
export AUDIT_TREE=$PWD                                  # the repo checkout
# Client tables (see orchestration/audit/ref/mechanics/MANIFEST.md):
python3 tools/sources/wago.py fetch 1.60.1.70425 --set mechanics
export AUDIT_DB=$(python3 tools/sources/wago.py fetch 1.60.1.70450 --set mechanics)   # prints the build's folder
# Sim, only when a pass needs numbers (the loop may already have built one: use its CLI path):
CLI=$(tools/audit/build_cli.sh origin/master)
go run --tags=with_db ./tools/rotopt/mksettings -spec <name> -gear <set> -apl <rotation> -talents <str> -race <Race> -duration 180 -out $AUDIT_WORK/mechanics/x.json
$CLI sim --infile $AUDIT_WORK/mechanics/x.json --outfile $AUDIT_WORK/mechanics/out.json && python3 tools/audit/mechanics/brk.py $AUDIT_WORK/mechanics/out.json
```

## The pass

**Get current first, on every pass.** The sim must never keep a mechanic Forever has since changed or
removed. Before looking at commits, establish the newest state of the game, and treat anything newer
than a verdict as a reason to re-check it:

- **Client build:** the newest 1.60 build (the gate's `client_build`, or `python3 tools/sources/wago.py latest --probe 300`).
  If it's newer than the build a verdict in `ref/mechanics/verdicts.md` rests on, re-check that verdict's spell
  data on the new build (step 5).
- **Blizzard:** every development-notes post, hotfix post and blue post on the WoW Forever forums newer than
  `dev_notes_seen` in your state (the thread us.forums.blizzard.com/en/wow/t/2360696 and the category's blue
  tracker; posters Kaivax, Aggrend, Gnedi, Linxy and Brenoch). An edited post keeps its date, so compare edit times.
- **What the daily update found:** the WoW Forever Daily Update routine reads the news every day and lands what it
  finds in the repo. Read what changed since your last pass:
  `git log --format='%h %ad %s' --date=short <last.master>..origin/master -- TASKS.md orchestration/update_ledger.md`,
  the merged `forever-update/*` PRs (`git log --merges --oneline <last.master>..origin/master | grep forever-update`),
  and the TASKS.md items they added.
- **Re-verify:** for each new change to a shared mechanic, racial, raid buff or debuff, consumable or enchant,
  check the matching row of `verdicts.md` and the sim. A sim value that a newer tier 1 source contradicts is a
  finding (`severity: high` when it moves DPS by 1% or more), even when no commit touched it. Update the row and the
  file's "as of" line when you confirm a value still holds on the newer build.

Record what you read as `dev_notes_seen` and `client_build` in your state, so the next pass starts from there.

1. **Fetch:** `git fetch -q origin --prune`.
2. **List what moved** since the master commit you last audited (`last.master` in `state/mechanics.json`). Diff by
   sha, not by time: a time window once missed a merge that landed a minute before it.
   `git log --no-merges --format='%h %cd %s' --date=format:'%m-%d %H:%M' <last.master>..origin/master -- 'sim/**/*.go' 'tools/rotopt/mksettings/**' ':!*_test.go'`
3. **Triage.** Read the diffs of commits touching:
   - `sim/core/` (rage.go, buffs.go, debuffs.go, racials_forever*.go, spell_result.go, spell_resistances.go,
     target.go, totem_weapon_buffs.go, mob_stats*.go, consumes)
   - `tools/rotopt/mksettings/` (shared raid and consumes)
   - any class file that claims a client value.
   For each claimed client value, look it up:
   - `python3 tools/audit/mechanics/sp.py <spellID or name regex>` for effects, auras, base points, procs, cooldowns,
     durations, class masks.
   - SpellMisc attributes for flags.
   - Era via `cmp.py`.
4. **Check knock-on effects.** A table change can move a formula elsewhere. Example: mob health feeds damage-taken
   Rage. `git grep` for the changed symbol's users.
5. **New client build.** The gate already probes wago.tools (300 builds past the newest listed one) and tells you
   when there's a newer build than `client_build` in your state. Then:
   - run `python3 tools/sources/wago.py fetch <new> --set mechanics`, then
     `python3 -I tools/sources/wago.py diff <old> <new> --by-spell > $AUDIT_WORK/mechanics/diff.txt`;
   - group the diff by spell name, grep it for shared spells (racials, raid buffs and debuffs, totems, consumes,
     trinkets, enchants), and diff SpellItemEnchantment, TraitDefinitionEffectPoints and CurvePoint separately;
   - class deltas: the rank tuning auditor tracks 31+ ranks in `orchestration/rank_tuning.md`, and the client update
     auditor writes the ledger; file class deltas the sim must port as findings.
   - Set `client_build` in your state.
6. **Blizzard notes:** covered by "Get current first" above. The gate starts you at least every 72 hours even when
   nothing changed in the repo, so this check never goes stale for long.
   Quote the exact wording for anything you rely on. Record the newest post you read as `dev_notes_seen` in your state
   (its URL and date), so the next pass reads only what's newer.
7. **Report.** New findings go to the findings file with evidence (spell IDs, fields, values, build) and a proposed
   fix; update `ref/mechanics/verdicts.md` with anything newly verified or settled; write `reports/mechanics.md`. A
   quiet pass (no shared-mechanic commits, no new build, no new Blizzard or daily-update information) only updates
   the state, including `dev_notes_seen`.

## What counts as a finding

The sim disagrees with tier 1 data, or a change silently moves a shared formula. File it whatever the DPS size, but
say the size. A user ruling or report assumption that departs from the client on purpose is not a finding. Stale docs
or comments that state wrong numbers are low-severity findings.

## Baseline (2026-10-17 13:00 UTC)

- **Master:** b389688f7, last audited (pass 28).
- **Client:** 70450 (spell data = 70425). No build answered between 70451 and 70750.
- **Damage-taken Rage at 60:** 3.4 Rage per 1500 hit; H(60) = 5356.
- **Raid hits:** 2 a minute at 1500 on the Level 60 warriors and Enhancement, 1 on Elemental.
- **Effects of the 70425 racial fixes:** Undead −2 to −17 DPS, Orc −3 to −7, Gnome warlocks +1 to +2.
- **Effects of #57** (Fury / Fury 2H / Arms at 180 s, seeds 1 / 1000003): −0.42 / −0.17, −0.57 / −0.23, −0.07 / −0.05.

## Gotchas

- The wago.tools builds API lags (it never listed 70425 or 70450). Find new Forever builds by probing db2 URLs.
- Forever's `EffectBasePointsF` is the real value. SpellLevels' columns differ from Era's; read them by name.
- Attribute bits: SPELL_ATTR0_COOLDOWN_ON_EVENT is Attributes_0 0x02000000. 0x10000000 is NOT_IN_COMBAT_ONLY_PEACEFUL
  (Stealth has it). One wrong bit here once flipped Nature's Swiftness.
- Effect 35 is a party area aura and effect 65 a raid area aura. Aura 290 is all crit, 57 spell crit, 52 weapon crit,
  163 crit damage bonus (it multiplies the whole crit multiplier), 674 is Forever's armor aura (same values as 22).
- `SpellResult.PreOutcomeDamage` is set after armor, so `RawDamage()` divides it back. `AttackerDamage` is the damage
  before anything on the target's side.
- SpellLevels BaseLevel 0 on Forever consumables is normal; don't read it as a wrong level.
- Don't trust a single table snapshot after copying files around. Re-fetch if a diff suddenly comes out empty.
