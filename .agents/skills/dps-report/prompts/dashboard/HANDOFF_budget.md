# Handoff: Forever item budget analysis (ilvl <= 30 only)

## Context
WoW Forever DPS presets project (repo /home/user/forever-sim, coordination branch claude/wow-forever-dps-presets-m9u15h).
The cross-class gear rules are in orchestration/standards.md (items 1, 1a, 2-7, 5a). Per-spec stat targets: orchestration/gear_targets.md.
An earlier analysis (orchestration/forever_item_budget.md, scripts in orchestration/tools/) used ALL Forever items. The user has now ruled
that only Forever items with item level <= 30 count: Forever's beta caps at level 30, and items above 30 may still carry stale, untuned data.

## Data
- /home/user/w-dash/assets/db_inputs/forever_wowhead_items.json: Wowhead's Forever item listing (12,114 items). Each item has level (ilvl),
  quality (2 green, 3 blue, 4 epic), slot (Wowhead inventory slot: 1 head, 2 neck, 3 shoulder, 5/20 chest, 6 waist, 7 legs, 8 feet, 9 wrist,
  10 hands, 11 finger, 12 trinket, 16 back, 13/17/21/22 weapons, 15/25/26 ranged, 23 off-hand, 28 relic), and eq stats
  (str, agi, sta, int, spi, critstrkrtng, hitrtng, splpwr, splheal, spldmg, atkpwr, rgdatkpwr, manargn, school spell power, resistances, ...).
- /home/user/w-dash/assets/db_inputs/wago_db2_items.csv: the Forever client's ItemSparse (static stat values in StatModifier_bonusStat_N /
  StatModifier_bonusAmount_N, ItemLevel, InventoryType, OverallQualityID). Stat enum: 3 Agi, 4 Str, 5 Int, 6 Spi, 7 Sta, and the classic ids
  for ratings / power. Use it to cross-check the Wowhead listing for ilvl <= 30 items.
- Era/Classic baseline for comparison: the same item IDs in Classic Era. /home/user/w-dash/assets/db_inputs/wowhead_item_tooltips.csv (classic tooltips, HTML)
  or the SoD repo's db (/home/user/wowsims/sod/assets/database/db.json, classic items + SoD items; stats array indices: 0 Str 1 Agi 2 Sta 3 Int 4 Spi 5 SP ... 17 AP 18 MeleeHit% 19 MeleeCrit% 27 RAP).
- Existing scripts to adapt: orchestration/tools/{primary.py, fit.py, single.py, epic.py, linear_total.py}. numpy and pandas are available, no scipy.

## What to answer (ilvl <= 30 Forever items only)
1. The combining exponent p (stats combine as (sum (w*v)^p)^(1/p)): test on primary-only items grouped by quality/slot/ilvl, as primary.py does. Is linear (p=1) ruled out?
2. Stat weights relative to primaries: spell power (splpwr), crit rating, hit rating, AP, RAP, MP5, healing. The earlier all-items fit gave spell power ~0.86-0.97 (not the standards' 0.55). Confirm or refute with ilvl <= 30 only, and say how many items back each weight.
3. Slot modifiers (head/chest/legs vs shoulder/hands/waist/feet vs neck/back/wrist/finger, and trinket if any data).
4. Budget per ilvl for green and blue at ilvl <= 30: fit V(ilvl) and the linear stat total per slot modifier (linear_total.py style).
5. Has Forever changed the low-level budget compared with Classic Era for the same items (same item IDs)? E.g. Forever item stat totals / Era totals by ilvl and quality. This decides whether the level 60 budget can be extrapolated from Era's level 60 items.
6. From 1-5, propose the most defensible level 60 epic budget for our synthetic gear (ilvl 66 armor, 71 jewelry/waist/wrist/back/one-handers, 76 two-handers, plus trinkets at 71): a per-piece linear total formula and a spell power weight, with the uncertainty stated. Compare with what we use now: standards item 1a, linear total = 1.73 x (1.2 ilvl - 36.4) x slot mod, spell power 0.55; and with the all-items proposal: slot mod x (1.28 ilvl - 10.9), spell power 0.86.

## Deliverables
- Update orchestration/forever_item_budget.md with a new section "ilvl <= 30 only (user ruling 2026-10-13)" that answers 1-6 with tables. Put new scripts in orchestration/tools/.
- Commit and push to claude/wow-forever-dps-presets-m9u15h. Fetch and rebase if the push races; never force-push. End the commit message with:
  <the attribution trailer lines your session prompt specifies>
- Don't message other sessions and don't touch class branches. No model names in files.
- Reply with a short summary: the numbers for 1-6 and your recommended formula.

## User decisions already made (for context, not for you to act on)
- Next gear round bundles per-spec Stamina (SoD BiS averages), Forever budget calibration (pending this analysis) and synthetic trinkets (two per spec at ilvl 71: a full-budget stat trinket and an on-use one).
- Hybrids (Ret, Enhancement) keep hybrid physical + spell stats (Forever devs' stated plan). Hunters keep ~10% Spirit.
- The user said to wait for their next message before starting the round. They have since sent the two points above (hybrid stats; ilvl <= 30 only).
