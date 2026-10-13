# Handoff: how real level 60 epics are budgeted per archetype (to predict Forever Phase 1 BiS)

## Context
WoW Forever DPS presets project (repo /home/user/forever-sim, coordination branch claude/wow-forever-dps-presets-m9u15h).
Each DPS spec gets a synthetic level 60 gear set (ilvl 66 armor; ilvl 71 jewelry, waist, wrist, back and one-handers;
ilvl 76 two-handers; two trinkets at ilvl 71). The rules are in orchestration/standards.md (items 1, 1a, 5a, 5b), with
per-spec stat targets in orchestration/gear_targets.md and the budget analysis in orchestration/forever_item_budget.md.
Read the last section of that file first ("ilvl <= 30 only"). Its scripts are in orchestration/tools/b30_*.py.

Findings so far:
- Forever left the low-level budget unchanged from Era. Stats combine on a p-norm with p of about 1.5 (not linear).
- Spell power weighs 0.86 on armor (0.93 on Classic-id items, 0.80 on Forever-new items; 35 single swaps fit 0.86 for 24
  items and 1.0 for 29). Caster weapons carry about 2.7 times the armor rate of spell power.
- The level 60 budget is slot mod x (1.125 ilvl + 2.0) as a weighted linear total, from 331 Era epics at ilvl 55-80.
- Standards 1a currently gives every piece the same weighted linear total. That removes the p-norm's
  "more stats per piece = more linear points" effect, which the warrior found workers could exploit.
- Real caster epics carry 4.4 stats per piece and physical ones 3.4. At SP 0.86, caster pieces sit at 1.04 of the common
  linear trend and physical pieces at 0.93.

## The user's ruling (verbatim)
"I want to as accurately as we can predict what phase 1 BiS sets will look like on WoW Forever on release. This means using
as accurate as possible budgeting formula that is used in Forvevr (i.e., if condfident that price of SP across ilvls is 0.86,
let's use that). If Epics in WoW SoD had special rules for casters/physcia/hybrids/etc, let's adopt similar rules. Basically,
if needed kerp analyzing to come up with our best guess of what end game BiS will look like."
Budget formula: the user picked the Era epics fit, slot mod x (1.125 x ilvl + 2.0).
Earlier rulings still in force:
- Only Forever items at ilvl <= 30 count as Forever evidence; ilvl 31+ may be untuned.
- Hybrids (Ret Paladin, Enhancement Shaman) keep hybrid physical + spell stats (Forever devs' stated plan).
- Hunters keep about 10% Spirit.
- Each set has a fair mix of strong and lesser-but-useful stats.
- Per-spec Stamina follows SoD BiS averages.

## Data
- Forever: /home/user/w-dash/assets/db_inputs/forever_wowhead_items.json (see b30_data.py for parsing, including the splheal-on-top fix).
- Era level 60 epics: the Wowhead Classic gear planner dump used by b30_era.py, and /home/user/wowsims/sod/assets/database/db.json
  (Classic and SoD items; stats array: 0 Str 1 Agi 2 Sta 3 Int 4 Spi 5 SP ... 17 AP 18 MeleeHit% 19 MeleeCrit% 27 RAP; check the proto for the rest).
  SoD items carry SoD's phase/ilvl; SoD level 60 phases (P4+) are the closest analogue to a "modern" itemization of level 60 epics with hybrid stats.
- SoD BiS lists: the SoD repo's ui presets (/home/user/wowsims/sod/ui/*/gear_sets or presets) and orchestration/sod_stat_mix.md.
- numpy and pandas are available, no scipy.

## Questions (answer each with numbers and item counts)
1. Archetype pricing. Classify Era level 60 epics (and SoD level 60 epics separately) as caster, physical (melee/ranged), hybrid
   (both physical and spell stats) and tank. On the p=1.5 norm with the measured weights, does each archetype sit on the same
   budget line, or does one carry systematically more or less? Report the residual per archetype and slot, so we know whether
   Blizzard (and SoD) priced archetypes differently or only the stat count differs.
2. Stats per piece: the distribution of stat count per piece by archetype and slot tier, for Era Phase 1 content
   (Molten Core, Onyxia, world bosses, pre-raid dungeons at ilvl 60-66) and for SoD level 60 epics. Also each archetype's typical
   budget split: share to primary stats, Stamina, spell power, hit, crit, AP, MP5.
3. Special rules to adopt. Look specifically at:
   - caster weapons' spell power rate;
   - ranged slot for casters (wands) and relics/idols/totems/librams;
   - the healing-vs-damage split;
   - hit and crit on caster vs physical items (weights 1.0/1.09 at Era 60);
   - Stamina share by archetype;
   - two-hander vs one-hander pricing;
   - how SoD itemized hybrid epics (Ret/Enh/Feral-type pieces carrying Str/Agi + SP/Int).
   Also check Forever ilvl <= 30 items that carry both physical and spell stats (count, stats per piece, residual vs the
   budget line) and Forever's Spirit/Int-to-SP swaps (is the swap priced at 0.86 or 1.0?).
4. Trinkets: Era/SoD level 60 stat trinkets and on-use trinkets per archetype. Budget of the passive part on the 1.5-norm,
   and the on-use effect size versus the passive stat value (average uptime x amount) so a synthetic on-use trinket is priced like real ones.
5. Proposal: a concrete pricing procedure for our synthetic pieces, to replace standards 1a. It must reproduce how real epics
   of each archetype are priced (p-norm with measured weights, or linear totals with per-archetype factors, whichever matches
   real items better), with realistic stats-per-piece limits per archetype and slot tier so the p-norm can't be gamed. Include any
   caster-weapon, relic, wand and trinket rules. Give a table: for each of the 27 builds' archetypes (pure caster, physical melee,
   hunter, hybrid Ret/Enh, feral), the expected weighted linear total per piece type under the proposal versus today's 1a, and the
   change in set totals. State the uncertainty.

## Deliverables
- A new file orchestration/bis_budget_rules.md answering 1-5 with tables, and scripts in orchestration/tools/bis_*.py.
- Commit and push to claude/wow-forever-dps-presets-m9u15h. Fetch and rebase if the push races; never force-push.
  End the commit message with:
  <the attribution trailer lines your session prompt specifies>
- Don't edit standards.md or gear_targets.md (the orchestrator will), don't message other sessions, and don't touch class branches.
  No model names in files.
- Reply with a short summary: the answers to 1-4 and the proposed procedure (5) with its per-archetype table.
