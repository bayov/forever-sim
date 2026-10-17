# Handoff: level 31+ rank tuning auditor

## The role

The level 60 DPS presets use spell ranks and deep talents that sit above Forever's playable cap (30).

The beta's balancing passes retuned only the ≤30 ranks. A 31+ rank is current for its spell's *design* (design changes hit every rank), but it never got the leveling-curve fine-tuning the ≤30 ranks get.

The auditor:
- keeps a verdict for each such spell: **tuned**, **stale-shaped**, **unchanged everywhere** or **unclear**;
- measures each one's DPS sensitivity by rescaling the 31+ data in a patched sim;
- watches for new client builds and master changes that could change either.

**Write only the report**, `orchestration/rank_tuning.md`, on the coordination branch `claude/wow-forever-dps-presets-m9u15h`. Make no other repo changes and post no PR comments.

**Workers don't change their sims on this audit's verdicts alone.** The orchestrator, or the user, rules.

The full method, the per-class verdict tables, the rank inversion scan and every pass's results are in that report. Read its "Method" and "Main finding" sections once before the first pass.

### Hard rules (from orchestration/auditors_common.md and the original brief)

- Don't use `tools/rotopt` (its searches, `-eval`, `-timeline`, `-breakdown`). `mksettings` and `wowsimcli` are fine.
- Never use vmangos as a source. Never decide from a tooltip alone; use client db2 data.
- wago.tools is the client data source. If it refuses a request, don't reach it any other way. A 503 is Cloudflare or the site being down; retry on a later pass.
- Put no model names in anything written.
- Commit trailer: the session's usual `Co-Authored-By:` / `Claude-Session:` lines.
- Push with `git push`. On a race, fetch and rebase. Never force-push.

## What to watch, and when a pass is quiet

1. **A new Forever client build on wago.tools:** `tools/audit/rank_tuning/wago_builds.sh` lists the 1.60.x builds. The baseline is **1.60.1.70450**.
2. **Master changes to the sensitivity spells:**

   ```
   git diff --stat <baseline sha> origin/master -- sim/warlock/curses.go sim/priest/smite.go \
     sim/druid/rip.go sim/druid/wrath.go sim/mage/frostbolt.go sim/paladin/holy_strike.go \
     sim/paladin/consecration.go sim/shaman/lightning_bolt.go sim/shaman/lava_burst.go sim/shaman/earth_shock.go
   ```

3. **Master changes to sim code or presets:**

   ```
   git diff --stat <baseline sha> origin/master -- sim tools/dps_report/req
   ```

   Ignore `*.results` and golden files.

Diff by sha, not by `git log --since`. A time window once missed a merge that landed a minute before it, and a by-sha diff can't do that.

**A pass is quiet** when there's no new build and no change under `sim/` (except results files) or `tools/dps_report/req`. A quiet pass writes nothing and messages no one.

## The pass, step by step

```bash
export AUDIT_WORK=<scratch dir>             # binary, 20k inputs and outputs, db2 tables
export AUDIT_TREE=<scratch checkout>        # throwaway: git worktree add --detach $AUDIT_TREE origin/master
A=tools/audit/rank_tuning                    # run from a checkout of the branch that holds these scripts
REF=orchestration/audit/ref/rank_tuning/baseline_d2633f4.csv

# 1. New client build?
$A/wago_builds.sh
AUDIT_WORK=$AUDIT_WORK $A/db2_fetch.sh 1.60.1.70450     # baseline tables, first time only
AUDIT_WORK=$AUDIT_WORK $A/db2_fetch.sh 1.60.1.<new>
python3 -I $A/db2_rowdiff.py $AUDIT_WORK/db 70450 <new>

# 2. Master moved? Run the two git diffs above.

# 3. If sim code or presets moved: build and re-sim.
$A/build.sh                                   # applies rt_sensitivity.patch, protoc, go build --tags=with_db
$A/run_all.sh                                 # all 66 presets + 16 sensitivities, ~5 min on 4 cores
#   or: $A/run_all.sh <preset> ...            # only the presets whose req changed, plus their jobs.txt lines
#   (summarize/--compare needs the other presets' bases too; for a partial run, copy the previous
#   run's m2_*.out.json into AUDIT_WORK first, then let run_all.sh overwrite only the named presets)
python3 $A/summarize.py $AUDIT_WORK --compare $REF
python3 $A/summarize.py $AUDIT_WORK --write <new baseline csv>   # once accepted
```

### What to compare, and the tolerances

**db2 diff:** any *changed* or *removed* row in SpellEffect, SpellLevels, SpellMisc, SpellPower, SpellCastTimes, SpellCooldowns, SpellDuration, CurvePoint or the seven Trait tables.
- If it touches a class spell, a rank, a proc or a talent, it's a finding. Name the spell and the field, and note whether it's ≤30, 31+ or every rank.
- Added rows that are mounts, pets, toys or transmog (skill lines 777 and 778, ensembles) are cosmetic. Say so in one line.
- Also look at added rows on class skill lines. SkillLineAbility's added list shows them.

**Sims:** seeds are fixed, so an unchanged tree reproduces the baseline exactly.
- `summarize.py --compare` flags these:
  - a base that moved more than 0.4%;
  - a sensitivity whose percentage moved more than 0.2 points;
  - a sensitivity whose place shift changed, meaning which presets it would pass or fall behind among the 66 bases.
- A sensitivity percentage that moves means the spell's share or data changed. That's a finding.
- Base moves that only shift places are worth a short report note. Message the orchestrator only if a ranking effect named in the report changes, such as who is first, or a spec passing or falling behind another.
- Differences under about 1 DPS (stdev/√20000 ≈ 0.2–0.3) are noise. Say "level within noise".

### Where findings go

1. Add a dated `## Check <date> <time> UTC: ...` section to `orchestration/rank_tuning.md`, just above `## Cross-class summary`.
2. Commit and push to the coordination branch.
3. Then send a short summary to the orchestrator: the class, the spell, the % and the ranking effect.

If a build changes a class's 31+ or deep-tier data, also tell that class's worker the deltas, and ask it to port them in a PR against master with paired 20k before/after numbers (120/180/300 s, two seed sets), never merging. The orchestrator asked for this on 70425.

## Current baseline (2026-10-17)

- **Client:** 1.60.1.70450. It's cosmetic against 70425: all 19 tables row-diffed, no existing row changed.
- **Master:** last audited at **d2633f4**. b389688, master at handoff, changed nothing under `sim/` or the req files.
- **Reference numbers:** `orchestration/audit/ref/rank_tuning/baseline_d2633f4.csv`. Headline values (180 s, 20k iterations):

| Sensitivity (`jobs.txt` tag) | Base → rescaled | Change | Place of 66 |
|---|---|---|---|
| Demonology, Bane of Doom coefficient 4.0→2.0 (`c2`) | 1177.7 → 1130.7 | −4.0% | 1 → 2 (behind only its own duty-curse preset; Lone Wolf 1094.1) |
| Affliction / Destruction / Destruction Fire (`c2`) | 1066.6 / 1047.4 / 1019.2 | −4.2 / −4.0 / −3.6% | 7→12 / 9→15 / 13→18 |
| Ret twist / Shock / sub-Holy, Holy Strike 31+ ×0.52 and Consecration 31+ ×0.82 (`both`) | 889.2 / 901.7 / 883.1 | −1.4% each | 37→41 / 34→37 / 39→43 |
| Elemental, Lightning Bolt X ×1.12 / ×1.38 / ×1.38 with Lava Burst (`lb112` / `lb138` / `lb138lvb`) | 771.8 | +2.1 / +6.8 / +8.4% | 50 → 50 / 48 / 47 |
| Smite 31+ ×1.18 (`s118`) | 765.4 | +2.7% | 51 → 50 |
| Feral, Rip 31+ ×1.13 (`r113`) | 752.9 | +1.7% | 53 → 51 |
| Balance Wrath-only / Balance, Wrath 31+ ×1.17 (`w117`) | 825.9 / 872.8 | +1.8 / +0.4% | 47→46 / 43→42 |
| Frost, Frostbolt 31+ ×1.08 (`fb`) | 843.3 | +1.5% | 44 → 44 |
| Enhancement, Earth Shock 31+ ×1.16 (`es116`) | 892.2 | +1.0% | 36 → 35 (level with Ret Shock within noise) |

Where each multiplier comes from:
- ×1.12 and ×1.08 are 70291's ≤30 raise carried to the top rank.
- The others hold the spell at its last playable Forever/Era ratio.
- ×0.52 holds Holy Strike's flat part to its ≤30 slope.
- Coefficient 2.0 for Bane of Doom is a plausible-tuning probe, not a measured value.

## Open items

- **Bane of Doom's 4.0 coefficient:** Forever design, never tested. It's worth about 4% to every Shadow Bolt warlock.
  - It no longer decides first place since #102 (no level-based partial resists).
  - Its biggest effect is on Affliction and Destruction places.
  - It waits on in-game data: a Bane of Doom tick against known spell power in a log would settle it.
- **Stale-shaped verdicts with no Blizzard fix yet:** Lightning Bolt X (the largest swing), Smite VIII, Wrath VIII, Rip VI (31+ diverges from ≤30), Raptor Strike VIII's flat part, Earth Shock VII, Holy Strike VIII's flat part. Frostbolt X is a mild "unclear" lag.
  - A new build that retunes any 31+ rank of these settles it. Diff, re-run, and report.
- **Untested 31-point spells, marked unclear:**
  - Incinerate, Wrack, Demonic Pact
  - Light's Vigil
  - Thousand Cuts, Venom, Adrenaline Rush
  - Moonkin Form, Wild Growth
  - Shadow Word: Death
  - Sniper Shot III
  - Consecration's server-scripted damage (client stores a dummy 30)
- **Not covered:**
  - SoD's server-side baseline-rank values (not in the client).
  - Pet ability ranks other than the hunter cat's.
  - The proc values of Forever's new rotational talents.
- **Pending user rulings:** none for this audit. Bane of Doom stays a sensitivity, not a change.

## Settled; don't re-flag

- **"Strongest learned rank" ruling (standards f8ae7cdc):**
  - The only real inversions in preset or raid use are hunter Aspect of the Hawk VI < V, Trueshot Aura V < IV and Hunter's Mark IV < III. Already handled by the hunter worker.
  - Data errors are excluded (89729b6). One such: Seal of Fury rank 4 (20419), coefficient 0.9 against 0 for rank 7. Ret doesn't use it.
- **Not inversions:**
  - Lightning Bolt, Chain Lightning and Lava Burst overload copies.
  - The Penance, Holy Shock, Flash of Light and Holy Light helper spells.
  - The SoD rune entries Lacerate 414647 and Tiger's Fury 417045.
  - The scripted dummies on Blizzard, Hurricane, Volley and Rain of Fire.
- **Tuned or "unchanged everywhere", don't revisit without a new build:**
  - Shadow Bolt, Fireball, Mind Blast and the other subtractive or multiplicative curves.
  - Seal and Judgement of Righteousness and Command, Execute, Heroic Strike, Whirlwind, Backstab, Sinister Strike, Eviscerate, Ferocious Bite, Windfury Weapon, pet Claw and Bite (all Era-identical).
  - Tiger's Fury, Multi-Shot and Stormstrike (redesigns).
- **70374 = 70338 and 70450 = 70425:** apart from cosmetic additions, the data is identical.
- **70425's deliberate deep-tier edits are ported on master:**
  - Lava Burst moved to a 20-point talent, ranks at 30/40/50.
  - Consecration ticks at 0.042.
  - Arcane Blast, Judgements, Devouring Plague and Chain Lightning costs; Holy Fire V; and the rest in the 70425 section of the report.
  - Ported by warlock #59, priest #58/#60, warrior #62, druid #61/#64, hunter #65, rogue #66, paladin #54/#81, mage #89/#90 and shaman #91.

## Gotchas

**Building the sim:**
- Build with `--tags=with_db`. Without it every preset fails with "No item with id: 9950xx" (the synthetic items live in the embedded db).
- Regenerate the Go protos after every checkout (`protoc ... --go_out=./sim/core`). A stale `api.pb.go` fails with `it.Kills undefined` and similar. `protoc-gen-go` lives in `$(go env GOPATH)/bin`.

**The sensitivity patch:**
- `rt_sensitivity.patch` hooks each spell's base damage after its `<Spell>BaseDamageAt(rank, level)` call. Smite and Rip were patched before those helpers existed.
- If a class refactor breaks the patch, re-add the same `if m, err := strconv.ParseFloat(os.Getenv("RT_..."), 64); err == nil && <level> > 30 { low *= m; high *= m }` hook by hand, and add `os` and `strconv` to the imports.
- Gate conditions:
  - Lightning Bolt (`RT_LB_MULT`) scales only the top rank (`rank == LightningBoltRanks`).
  - Lava Burst (`RT_LVB_MULT`) scales every rank, since all of them are deep-tier.
  - Bane of Doom (`RT_BANE_COEF`) replaces the coefficient (default 4).
  - The rest scale ranks whose level is above 30.
- Unset env vars leave the sim unchanged, so the `base` tag is the real master behaviour.

**Running sims:**
- `run2.sh` writes the 20k input once per preset (`m2_<preset>.in.json`). Delete it whenever the req file changes, or the run silently uses the old preset. `run_all.sh` does this for you.
- Parallel runs interleave their stdout lines. Read results from the `.out.json` files with `summarize.py`, not from the printed lines.
- The ranking is out of all 66 presets, duty variants included (curse duty, Battle Shout duty, Faerie Fire duty and so on). Name the neighbouring presets when you report a place change.

**Diffing builds:**
- Diff db2 tables by row ID. A class-spell-only diff missed 70425's proc and aura spells (Seal of Righteousness procs, Consecration ticks, Fire Nova damage, Mana Tide, raid auras). Diff everything, then sort.
- Forever SpellEffect uses `EffectBasePointsF` (average) and `Variance` (spread ÷ average). Era uses `EffectBasePoints + (DieSides+1)/2`.
- When comparing ranks, skip script effects (Effect 77, 3, 64). Prefer the effect that carries a coefficient.
- Run downloaded CSVs through `python3 -I`, and keep the scripts outside the download directory.
