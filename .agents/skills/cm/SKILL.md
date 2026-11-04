---
name: cm
description: Commit all local changes with jj in topic commits and push master. Use when the user runs /cm or says "commit" or "push".
---

# Commit and push all local changes

Several agents often share one jj working copy, so it usually holds unrelated pieces of
work: a UI change, a regear with its new presets, a DB regen, notes. Split them into
commits by topic, then push `master` to origin.

Run every command from the repo root, `/home/bayov/projects/forever-sim`.

## 1. See what changed

```bash
cd /home/bayov/projects/forever-sim
jj diff --stat | tail -60
jj diff --summary
```

jj tracks every file that isn't ignored, so look for added files that don't belong: build
output, binaries, big logs, rotopt result dumps, scratch scripts. Ask the user before
committing one. They usually want it deleted or added to `.gitignore`.

## 2. Commit by topic

Give each piece of work its own commit with a real message, by listing its paths:

```bash
jj commit -m "Move custom presets into a Custom folder with a save modal" \
	ui/core/components/saved_data_manager.tsx ui/core/components/preset_tree.tsx \
	ui/scss/core/components/_saved_data_manager.scss
```

Some files go with the work that produced them:
- The generated item DB (`assets/database/*`) goes with the `assets/db_inputs/*` or
  `tools/database` change that regenerated it.
- A regear's `ui/*/gear_sets/*.json`, `ui/*/apls/*.apl.json` and `presets.ts` changes go
  together, with the `forever-wiki/` or `notes.md` lines that record the result.

Don't split a file that holds two topics. Put it in the commit it fits best.

When a commit fails with "Could not acquire lock for index file" (another agent's jj
command holds `.git/index.lock`), it usually went through anyway, but jj leaves a second,
divergent copy of the change. Check with `jj log -r 'master..@'`. When the extra copy has
no description and `jj diff --from <its commit> --to @` shows no changes, `jj abandon` it.

Then commit what's left. Skip this when `jj diff --stat` prints "0 files changed", or you
push an empty commit.

```bash
jj commit -m "Update notes and wiki"
```

Messages are one short line in the imperative, in plain ASCII (no em dashes or
semicolons). Say what changed in domain words, like "Give the Level 30 PvP preset the
High HP gear" or "Fix Healing Stream ticks under Forever". See `jj log --limit 15` for the
style. Never use a placeholder like `c` or `wip`.

## 3. Push

```bash
jj bookmark set master -r @-
jj git push --remote origin --bookmark master
jj bookmark list --all master
```

Origin is `git@github.com:bayov/forever-sim.git`, over SSH, since git has no credentials
for the https URL.

The push is done when `master` and `master@origin` show the same commit. Ignore
`master@upstream`: that's the wowsims repo we forked, and we never push there.

When jj refuses to move `master` backwards or sideways, stop and tell the user. Don't pass
`--allow-backwards`. When GitHub rejects the push (a file over 100 MB, say), tell the user
what it said.

## 4. Report

Tell the user each commit's ID and message, and that origin's `master` matches.
