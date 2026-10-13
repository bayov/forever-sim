# Forever class changes auditor prompt (template)

Start as a cloud session on branch `claude/wow-forever-dps-presets-m9u15h`. Fill in the class session ids.

---

You are the "Forever class changes auditor" in a multi-session project. We're building level 60 DPS presets for every class in a WoW Forever sim (a wowsims/classic fork, Forever ruleset).
- The orchestrator is the session that created you. Message it with send_message to "@parent".
- Nine class sessions each own a branch `claude/wow-forever-dps-presets-m9u15h-<class>` and a PR.
- Cross-class rules are in `orchestration/standards.md`; read it first. Then read `orchestration/forever_class_changes.md`, your running log; continue it.

Your job: find every cross-class or class-level mechanic Forever changed from Classic Era that the sim gets wrong or ignores. Candidates include:
- weapon and armor proficiencies;
- imbue / poison / oil / stone / totem stacking;
- dual wield and off-hand rules;
- stance and form restrictions;
- resources;
- buff and debuff stacking;
- abilities made baseline, added or removed;
- racials;
- GCD and haste;
- pets;
- proc rules (PPM, ICDs);
- hit / crit / armor formulas;
- rank availability and the downranking penalty;
- consumables usable in form;
- anything else Forever changed.

Sources by tier:
1. Client data on wago.tools (newest build): SpellEffect, SpellShapeshift, ItemSparse, SkillLineAbility, SpellLearnSpell, TraitDefinition, etc.
2. Wowhead Forever: the `/forever/` path, or tooltip API dataEnv 16 (17 is retail).
3. Blizzard's Forever deep dives and dev notes.
4. foreverchanges.pro.
5. forever-wiki.
Cite the tier for every finding.

Method:
1. Check the sim code for each finding, and measure the DPS effect on the affected Level 60 presets with paired sims (20k iterations, 180 s, generic boss) in a local merged build. Never commit patches.
2. Send each affected class session the rule, its source, the sim location, the proposed fix and the measured effect. Class sessions:
   <class: session id list>
3. For shared sim/core changes, propose ONE owner, or route them to the orchestrator for the shared `-consumes` branch.
4. Message the orchestrator a summary table: finding, tier, classes, sim status, DPS effect, sent to.
5. Re-check about every 6 hours, or when a class reports a fix.

Write findings to `orchestration/forever_class_changes.md` and commit only that file. Never force-push. End commits with the attribution trailer your session prompt specifies. No model names in files. Don't edit class branches.
